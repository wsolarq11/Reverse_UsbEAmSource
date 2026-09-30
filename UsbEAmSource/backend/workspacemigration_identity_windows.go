package main

import (
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"golang.org/x/sys/windows"
)

// workspaceMigrationIdentityForExisting 为已存在的迁移路径构建 Windows 文件身份。
// [S] ASM 0x1409f46e0: UTF16PtrFromString；CreateFile(FILE_READ_ATTRIBUTES,
// SHARE_READ|WRITE|DELETE, OPEN_EXISTING, BACKUP_SEMANTICS|OPEN_REPARSE_POINT)；
// defer CloseHandle；GetFileInformationByHandle；REPARSE_POINT→
// newExtractError("迁移路径最终句柄指向 reparse point")；GetFinalPathNameByHandle
// 循环（buf 不足则 n+1 重试）；normalizeWorkspaceWindowsFinalPath(string(utf16.Decode))；
// identity{canonical, VolumeSerialNumber, FileIndexHigh<<32|FileIndexLow, true}。
// info 为死参数（源码保留但未使用）。
func workspaceMigrationIdentityForExisting(path string, _ os.FileInfo) (workspacePathIdentity, error) {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return workspacePathIdentity{}, err
	}

	handle, err := windows.CreateFile(
		ptr,
		windows.FILE_READ_ATTRIBUTES,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		return workspacePathIdentity{}, err
	}
	defer windows.CloseHandle(handle)

	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		return workspacePathIdentity{}, err
	}
	if info.FileAttributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return workspacePathIdentity{}, newExtractError("迁移路径最终句柄指向 reparse point")
	}

	buf := make([]uint16, 1024)
	for {
		n, err := windows.GetFinalPathNameByHandle(handle, &buf[0], uint32(len(buf)), 0)
		if err != nil {
			return workspacePathIdentity{}, err
		}
		if uint32(len(buf)) <= n {
			buf = make([]uint16, int(n)+1)
			continue
		}
		finalPath := normalizeWorkspaceWindowsFinalPath(string(utf16.Decode(buf[:n])))
		return workspacePathIdentity{
			canonical: finalPath,
			volume:    uint64(info.VolumeSerialNumber),
			file:      uint64(info.FileIndexHigh)<<32 | uint64(info.FileIndexLow),
			valid:     true,
		}, nil
	}
}

// workspaceMigrationPathIsReparse 判断路径是否为 reparse point（符号链接/挂载点）。
// [S] ASM 0x1409f4660: UTF16PtrFromString err→(false,err)；GetFileAttributes err→(false,err)；
// 否则 (attrs&FILE_ATTRIBUTE_REPARSE_POINT(0x400，bit10)!=0, nil)。返回 (bool, error)。
// 签名订正（batch 268）：batch 267 误落单 bool，validateWorkspaceMigrationExistingChain
// 调用点实证检查 err（asm 0x1409f42ea test rbx,rbx）→ 真实为双返回。
func workspaceMigrationPathIsReparse(path string) (bool, error) {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false, err
	}
	attrs, err := windows.GetFileAttributes(ptr)
	if err != nil {
		return false, err
	}
	return attrs&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0, nil
}

// normalizeWorkspaceWindowsFinalPath 规范化 Windows 长路径前缀。
// [S] ASM 0x1409f4d00：TrimSpace → ToLower → 前缀 `\\?\unc\`(8B，小写比较) →
// `\\`+trimmed[8:]；前缀 `\\?\`(4B) → trimmed[4:]；否则原样；最终 filepath.Clean。
// 语义：`\\?\C:\x`→`C:\x`，`\\?\UNC\srv\share`→`\\srv\share`（保留 `\\` 头）。
func normalizeWorkspaceWindowsFinalPath(path string) string {
	trimmed := strings.TrimSpace(path)
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, `\\?\unc\`) {
		return filepath.Clean(`\\` + trimmed[8:])
	}
	if strings.HasPrefix(lower, `\\?\`) {
		return filepath.Clean(trimmed[4:])
	}
	return filepath.Clean(trimmed)
}

// openWorkspaceMigrationSourceFileNoFollow 以 GENERIC_READ+OPEN_EXISTING 打开源文件，不跟随符号链接。
// [S] ASM 0x1409f4e20：openWorkspaceMigrationRegularFileWindows(path, 0x80000000, 3)。
func openWorkspaceMigrationSourceFileNoFollow(path string) (*os.File, error) {
	return openWorkspaceMigrationRegularFileWindows(path, windows.GENERIC_READ, windows.OPEN_EXISTING)
}

// openWorkspaceMigrationTargetFileNoFollow 以 GENERIC_WRITE+CREATE_NEW 打开目标文件，不跟随符号链接。
// [S] ASM 0x1409f4e80：openWorkspaceMigrationRegularFileWindows(path, 0x40000000, 1)。
func openWorkspaceMigrationTargetFileNoFollow(path string) (*os.File, error) {
	return openWorkspaceMigrationRegularFileWindows(path, windows.GENERIC_WRITE, windows.CREATE_NEW)
}

// openWorkspaceMigrationRegularFileWindows 以 NOFOLLOW 语义打开工作区迁移文件并验证是普通文件。
// [S] ASM 0x1409f4ee0：UTF16PtrFromString → CreateFile(FILE_FLAG_OPEN_REPARSE_POINT，
// sharemode = OPEN_EXISTING?READ|WRITE|DELETE:0) → GetFileInformationByHandle →
// (REPARSE_POINT|DIRECTORY)!=0 → "迁移文件最终句柄不是普通文件"；
// os.NewFile 返 nil → "无法包装迁移文件句柄"。
func openWorkspaceMigrationRegularFileWindows(path string, access uint32, disposition uint32) (*os.File, error) {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}

	shareMode := uint32(0)
	if disposition == windows.OPEN_EXISTING {
		shareMode = windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE | windows.FILE_SHARE_DELETE
	}

	handle, err := windows.CreateFile(
		ptr,
		access,
		shareMode,
		nil,
		disposition,
		windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		return nil, err
	}

	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		windows.CloseHandle(handle)
		return nil, err
	}

	if info.FileAttributes&(windows.FILE_ATTRIBUTE_REPARSE_POINT|windows.FILE_ATTRIBUTE_DIRECTORY) != 0 {
		windows.CloseHandle(handle)
		return nil, newExtractError("迁移文件最终句柄不是普通文件")
	}

	file := os.NewFile(uintptr(handle), path)
	if file == nil {
		windows.CloseHandle(handle)
		return nil, newExtractError("无法包装迁移文件句柄")
	}

	return file, nil
}
