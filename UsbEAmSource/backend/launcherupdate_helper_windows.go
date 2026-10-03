package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// openLauncherUpdateReadOnlyHandle 以只读 + 顺序扫描打开启动器更新文件句柄。
// [S 汇编 0x1408c4d20, 128B]：UTF16PtrFromString 失败 → (0,err)；CreateFile(GENERIC_READ,
// SHARE_READ, nil, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL|FILE_FLAG_SEQUENTIAL_SCAN, 0)。
func openLauncherUpdateReadOnlyHandle(path string) (windows.Handle, error) {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	return windows.CreateFile(
		ptr,
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_SEQUENTIAL_SCAN,
		0,
	)
}

// openLauncherUpdateReadOnlyInheritedHandle 在只读句柄上置 HANDLE_FLAG_INHERIT 使子进程可继承。
// [S 汇编 0x1408c4da0, 160B]：openLauncherUpdateReadOnlyHandle 失败 → (0,err)；
// SetHandleInformation(handle, HANDLE_FLAG_INHERIT, HANDLE_FLAG_INHERIT) 失败 →
// CloseHandle(handle) 后 (0,err)；成功 → (handle,nil)。
func openLauncherUpdateReadOnlyInheritedHandle(path string) (windows.Handle, error) {
	handle, err := openLauncherUpdateReadOnlyHandle(path)
	if err != nil {
		return 0, err
	}
	if err := windows.SetHandleInformation(handle, windows.HANDLE_FLAG_INHERIT, windows.HANDLE_FLAG_INHERIT); err != nil {
		windows.CloseHandle(handle)
		return 0, err
	}
	return handle, nil
}

// launcherUpdateProcessImagePath 查询进程完整映像路径并清理（filepath.Clean）。
// [S 汇编 0x1408c4960, 224B]：栈上 [0x8000]uint16 清零，size=0x8000；
// QueryFullProcessImageName(handle, 0, &buf[0], &size) 失败 → ("",err)；
// 否则 filepath.Clean(UTF16ToString(buf[:size]))。
func launcherUpdateProcessImagePath(handle windows.Handle) (string, error) {
	var buf [0x8000]uint16
	size := uint32(len(buf))
	if err := windows.QueryFullProcessImageName(handle, 0, &buf[0], &size); err != nil {
		return "", err
	}
	return filepath.Clean(windows.UTF16ToString(buf[:size])), nil
}

// terminateLauncherUpdateProcess 以 TERMINATE|SYNCHRONIZE 打开并终止更新进程，等待其退出。
// [S 汇编 0x1408c4e40, 192B]：OpenProcess(PROCESS_TERMINATE|SYNCHRONIZE,false,pid)
// 失败 → (handle,err)；成功 defer CloseHandle → TerminateProcess(handle,1) →
// WaitForSingleObject(handle,10000) → (handle,nil)。
func terminateLauncherUpdateProcess(pid uint32) (windows.Handle, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_TERMINATE|windows.SYNCHRONIZE, false, pid)
	if err != nil {
		return handle, err
	}
	defer windows.CloseHandle(handle)
	_ = windows.TerminateProcess(handle, 1)
	_, _ = windows.WaitForSingleObject(handle, 10000)
	return handle, nil
}

// restartRolledBackLauncher 用回滚目录与启动器名重新拉起进程。
// [S 汇编 0x1408c4f60, 192B]：filepath.Join(dir,name) → exec.Command(path)
// （无 args）→ cmd.Dir=dir → cmd.Start()。asm 额外携带 dead 中间参数。
func restartRolledBackLauncher(dir, name string) error {
	cmd := exec.Command(filepath.Join(dir, name))
	cmd.Dir = dir
	return cmd.Start()
}

// launcherUpdateSameIdentity 判定两个进程身份是否同一（短路径字段严格相等，
// UserSID/ImageSHA256 大小写不敏感，ImagePath 按 samePathFold）。
// [S 汇编 0x1408c4be0, 320B]：逐字段比对，UserSID/ImageSHA256 用 strings.EqualFold，
// ImagePath 用 samePathFold，其余 uint32/int64 严格相等。
func launcherUpdateSameIdentity(a, b launcherUpdateProcessIdentity) bool {
	if a.PID != b.PID || a.ParentPID != b.ParentPID || a.CreatedAt != b.CreatedAt || a.SessionID != b.SessionID {
		return false
	}
	if !strings.EqualFold(a.UserSID, b.UserSID) {
		return false
	}
	if a.IntegrityRID != b.IntegrityRID {
		return false
	}
	if !samePathFold(a.ImagePath, b.ImagePath) {
		return false
	}
	if !strings.EqualFold(a.ImageSHA256, b.ImageSHA256) {
		return false
	}
	if a.VolumeSerial != b.VolumeSerial || a.FileIndexHigh != b.FileIndexHigh {
		return false
	}
	return a.FileIndexLow == b.FileIndexLow
}

// launcherUpdateTokenIntegrityRID 读取令牌完整性级别（TokenIntegrityLevel）的 RID。
// [S 汇编 0x1408c4a40, 416B]：GetTokenInformation 探测大小，size==0 →
// "无法读取令牌完整性级别"；分配缓冲再取，err 透传；buffer 首 8 字节为 SID 指针，
// SID 空或 SubAuthorityCount==0 → "令牌完整性 SID 无效"；否则 SubAuthority(count-1)。
func launcherUpdateTokenIntegrityRID(token windows.Token) (uint32, error) {
	var size uint32
	_ = windows.GetTokenInformation(token, windows.TokenIntegrityLevel, nil, 0, &size)
	if size == 0 {
		return 0, errors.New("无法读取令牌完整性级别")
	}
	buf := make([]byte, size)
	if err := windows.GetTokenInformation(token, windows.TokenIntegrityLevel, &buf[0], size, &size); err != nil {
		return 0, err
	}
	// TOKEN_MANDATORY_LABEL 布局：Label.Sid（*SID）后跟 Attributes（uint32）。
	label := (*struct {
		Sid        *windows.SID
		Attributes uint32
	})(unsafe.Pointer(&buf[0]))
	sid := label.Sid
	if sid == nil || sid.SubAuthorityCount() == 0 {
		return 0, errors.New("令牌完整性 SID 无效")
	}
	return sid.SubAuthority(uint32(sid.SubAuthorityCount()) - 1), nil
}

// queryLauncherUpdateParentPID 通过进程快照枚举查找指定 PID 的父进程 PID。
// [S 汇编 0x1408c46c0, 576B]：CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS,0)
// 失败 → (0,err)；defer CloseHandle；Process32First 失败 → (0,err)；循环比对
// ProcessID，命中 → (ParentProcessID,nil)；Process32Next 失败 → "未找到进程父 PID"。
func queryLauncherUpdateParentPID(pid uint32) (uint32, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(snapshot)
	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	if err := windows.Process32First(snapshot, &pe); err != nil {
		return 0, err
	}
	for {
		if pe.ProcessID == pid {
			return pe.ParentProcessID, nil
		}
		if err := windows.Process32Next(snapshot, &pe); err != nil {
			return 0, errors.New("未找到进程父 PID")
		}
	}
}

// launcherUpdateHandshakeProof 计算握手证明：固定前缀（含 \x00 终止）拼接 nonce 后
// SHA-256，结果十六进制小写编码。
// [S 汇编 0x1408c3b60, 512B]：sha256.Sum256("UsbEAm launcher update handshake v2\x00"+nonce)
// → hex.EncodeToString。
func launcherUpdateHandshakeProof(nonce string) string {
	sum := sha256.Sum256([]byte("UsbEAm launcher update handshake v2\x00" + nonce))
	return hex.EncodeToString(sum[:])
}

// validateLauncherUpdateHelperExecutablePath 校验 helper 可执行文件位于认证临时目录。
// [S 汇编 0x1408c5020, 384B]：TrimSpace+Clean 后过 validateLauncherUpdateAbsolutePath，
// 失败 → "更新 helper 自身路径无效"；Base 须 EqualFold "UsbEAm_Launcher_Updater.exe"；
// ToLower(Base(Dir)) 须等于 "usbeam-launcher-updater-"；samePathFold(Dir(Dir),os.TempDir())。
// 任一步不满足 → "更新 helper 不在认证临时目录"。
func validateLauncherUpdateHelperExecutablePath(path string) error {
	clean, err := validateLauncherUpdateAbsolutePath(filepath.Clean(strings.TrimSpace(path)))
	if err != nil {
		return errors.New("更新 helper 自身路径无效")
	}
	dir := filepath.Dir(clean)
	if !strings.EqualFold(filepath.Base(clean), "UsbEAm_Launcher_Updater.exe") {
		return errors.New("更新 helper 不在认证临时目录")
	}
	if strings.ToLower(filepath.Base(dir)) != "usbeam-launcher-updater-" {
		return errors.New("更新 helper 不在认证临时目录")
	}
	if !samePathFold(filepath.Dir(dir), os.TempDir()) {
		return errors.New("更新 helper 不在认证临时目录")
	}
	return nil
}

// cleanupOldLauncherUpdateHelpers 清理认证临时目录下过期（24h 前）的 helper 目录。
// [S 汇编 0x1408c51a0, 480B]：os.ReadDir(os.TempDir()) 失败即返回；遍历目录项，
// 仅当 IsDir 且 ToLower(Name)=="usbeam-launcher-updater-" 且 Info().ModTime() 早于
// now-24h 时 os.RemoveAll(filepath.Join(tempDir,name))。
func cleanupOldLauncherUpdateHelpers() {
	dir := os.TempDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-24 * time.Hour)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.ToLower(entry.Name()) != "usbeam-launcher-updater-" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if !info.ModTime().Before(cutoff) {
			continue
		}
		_ = os.RemoveAll(filepath.Join(dir, entry.Name()))
	}
}

// scheduleLauncherUpdateHelperCleanup 把 helper 可执行文件与其目录登记为重启后删除。
// [S 汇编 0x1408c5380, 608B]：os.Executable 失败即返回；对 exe 与 filepath.Dir(exe)
// 分别 UTF16PtrFromString 成功后 MoveFileEx(p,nil,MOVEFILE_DELAY_UNTIL_REBOOT)。
func scheduleLauncherUpdateHelperCleanup() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	if p, err := windows.UTF16PtrFromString(exe); err == nil {
		_ = windows.MoveFileEx(p, nil, windows.MOVEFILE_DELAY_UNTIL_REBOOT)
	}
	if d, err := windows.UTF16PtrFromString(filepath.Dir(exe)); err == nil {
		_ = windows.MoveFileEx(d, nil, windows.MOVEFILE_DELAY_UNTIL_REBOOT)
	}
}
