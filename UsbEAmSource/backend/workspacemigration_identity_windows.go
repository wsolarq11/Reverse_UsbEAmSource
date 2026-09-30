package main

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// workspaceMigrationPathIsReparse 判断路径是否为 reparse point（符号链接/挂载点）。
// [S] ASM 0x1409f4660: UTF16PtrFromString → GetFileAttributes，任一步 err→false，
// 否则 attrs&FILE_ATTRIBUTE_REPARSE_POINT(0x400，bit10)!=0。
func workspaceMigrationPathIsReparse(path string) bool {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false
	}
	attrs, err := windows.GetFileAttributes(ptr)
	if err != nil {
		return false
	}
	return attrs&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0
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
