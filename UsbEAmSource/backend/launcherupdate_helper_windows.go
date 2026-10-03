package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
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

// writeLauncherUpdateFrame 写一帧：先写 little-endian uint32 长度头，再写 payload。
// [S 汇编 0x1408c31e0, 288B]：len==0 或 len>0x810000 → "更新认证消息大小无效"；
// binary.Write(conn,LittleEndian,uint32(len)) 失败透传；否则 conn.Write(payload) 透传 err。
func writeLauncherUpdateFrame(conn io.Writer, payload []byte) error {
	if len(payload) == 0 || len(payload) > 0x810000 {
		return errors.New("更新认证消息大小无效")
	}
	if err := binary.Write(conn, binary.LittleEndian, uint32(len(payload))); err != nil {
		return err
	}
	_, err := conn.Write(payload)
	return err
}

// readLauncherUpdateFrame 读一帧：先读 little-endian uint32 长度头，再读满 payload。
// [S 汇编 0x1408c3300, 320B]：binary.Read 失败透传；size==0 或 int(size)>maxSize →
// "更新认证消息大小无效"；io.ReadAtLeast(conn,buf,int(size)) 失败透传；成功返回 (buf,nil)。
func readLauncherUpdateFrame(conn io.Reader, maxSize int) ([]byte, error) {
	var size uint32
	if err := binary.Read(conn, binary.LittleEndian, &size); err != nil {
		return nil, err
	}
	if size == 0 || int(size) > maxSize {
		return nil, errors.New("更新认证消息大小无效")
	}
	buf := make([]byte, int(size))
	if _, err := io.ReadAtLeast(conn, buf, int(size)); err != nil {
		return nil, err
	}
	return buf, nil
}

// createLauncherUpdateNamedPipe 创建认证命名管道：SDDL 授权 SYSTEM/Administrators/当前用户。
// [S 汇编 0x1408c2940, 352B]：GetCurrentProcessToken().GetTokenUser() 失败或 User.Sid 空 →
// "无法读取当前用户 SID"；SDDL "D:P(A;;GA;;;SY)(A;;GA;;;BA)(A;;GA;;;"+sid+")"；
// SecurityDescriptorFromString 失败透传；UTF16PtrFromString 失败透传；CreateNamedPipe
// (FILE_FLAG_OVERLAPPED|FIRST_PIPE_INSTANCE|PIPE_ACCESS_DUPLEX, PIPE_REJECT_REMOTE_CLIENTS,
// 1, 65536, 65536, 30000, sa)。
func createLauncherUpdateNamedPipe(pipeName string) (windows.Handle, error) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil || user == nil || user.User.Sid == nil {
		return 0, errors.New("无法读取当前用户 SID")
	}
	sddl := "D:P(A;;GA;;;SY)(A;;GA;;;BA)(A;;GA;;;" + user.User.Sid.String() + ")"
	sd, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		return 0, err
	}
	sa := &windows.SecurityAttributes{
		Length:             24,
		SecurityDescriptor: sd,
	}
	name, err := windows.UTF16PtrFromString(pipeName)
	if err != nil {
		return 0, err
	}
	return windows.CreateNamedPipe(name,
		windows.FILE_FLAG_OVERLAPPED|windows.FILE_FLAG_FIRST_PIPE_INSTANCE|windows.PIPE_ACCESS_DUPLEX,
		windows.PIPE_REJECT_REMOTE_CLIENTS,
		1, 65536, 65536, 30000, sa)
}

// openLauncherUpdateNamedPipeClient 以 20ms 间隔重试连接认证命名管道直到超时。
// [S 汇编 0x1408c3040, 416B]：UTF16PtrFromString 失败透传；循环 CreateFile
// (GENERIC_READ|GENERIC_WRITE, 0, OPEN_EXISTING, FILE_FLAG_OVERLAPPED|FILE_ATTRIBUTE_NORMAL)；
// 成功返回 handle；ERROR_PIPE_BUSY/ERROR_FILE_NOT_FOUND 则 time.Sleep(20ms) 重试，
// time.Now().After(deadline) 超时 → "连接更新认证 named pipe 超时"；其他错误透传。
func openLauncherUpdateNamedPipeClient(pipeName string, timeout time.Duration) (windows.Handle, error) {
	deadline := time.Now().Add(timeout)
	name, err := windows.UTF16PtrFromString(pipeName)
	if err != nil {
		return 0, err
	}
	for {
		handle, err := windows.CreateFile(name,
			windows.GENERIC_READ|windows.GENERIC_WRITE,
			0, nil, windows.OPEN_EXISTING,
			windows.FILE_FLAG_OVERLAPPED|windows.FILE_ATTRIBUTE_NORMAL, 0)
		if err == nil {
			return handle, nil
		}
		if !errors.Is(err, windows.ERROR_PIPE_BUSY) && !errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
			return 0, err
		}
		if time.Now().After(deadline) {
			return 0, errors.New("连接更新认证 named pipe 超时")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// writeLauncherUpdateFrameWithTimeout 带超时的帧写入：后台 goroutine 写帧，超时关连接。
// [S 汇编 0x1408c3820, 640B]：make(chan error,1) + go func1 写帧；NewTimer 后 defer Stop；
// select timer.C → conn.Close() + "写入更新认证 named pipe 超时"；select ch → 返回写帧 err。
func writeLauncherUpdateFrameWithTimeout(conn *os.File, payload []byte, timeout time.Duration) error {
	ch := make(chan error, 1)
	go func() {
		ch <- writeLauncherUpdateFrame(conn, payload)
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-timer.C:
		if conn != nil {
			_ = conn.Close()
		}
		return errors.New("写入更新认证 named pipe 超时")
	case err := <-ch:
		return err
	}
}

// readLauncherUpdateFrameWithTimeout 带超时的帧读取：后台 goroutine 读帧，超时关连接。
// [S 汇编 0x1408c3440, 768B]：make(chan {data,err},1) + go func1 读帧；NewTimer 后 defer
// Stop；select timer.C → conn.Close() + "读取更新认证 named pipe 超时"；select ch → 返回结果。
func readLauncherUpdateFrameWithTimeout(conn *os.File, maxSize int, timeout time.Duration) ([]byte, error) {
	ch := make(chan struct {
		data []byte
		err  error
	}, 1)
	go func() {
		data, err := readLauncherUpdateFrame(conn, maxSize)
		ch <- struct {
			data []byte
			err  error
		}{data, err}
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-timer.C:
		if conn != nil {
			_ = conn.Close()
		}
		return nil, errors.New("读取更新认证 named pipe 超时")
	case r := <-ch:
		return r.data, r.err
	}
}

// connectLauncherUpdateNamedPipe 重叠模式等待命名管道客户端连接。
// [S 汇编 0x1408c2aa0, 1344B]：CreateEvent(nil,true,false,nil) 失败透传；defer CloseHandle；
// ConnectNamedPipe(pipe,&overlapped{HEvent})；err==nil 或 ERROR_IO_PENDING → 返回 nil；
// ERROR_PIPE_CONNECTED → WaitForSingleObject(event,毫秒超时)；WAIT_OBJECT_0 →
// GetOverlappedResult(wait=false) 失败透传后返回 nil；WAIT_TIMEOUT → CancelIoEx，非
// ERROR_NOT_FOUND 失败 → fmt.Errorf "取消更新认证 named pipe 连接失败: %w"，否则
// GetOverlappedResult(wait=true) + "等待更新认证 named pipe 连接超时"；其他 →
// fmt.Errorf "等待更新认证 named pipe 连接返回未知状态: %d"。
func connectLauncherUpdateNamedPipe(pipe windows.Handle, timeout time.Duration) error {
	event, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(event)

	overlapped := windows.Overlapped{
		HEvent: event,
	}

	err = windows.ConnectNamedPipe(pipe, &overlapped)
	if err == nil || errors.Is(err, windows.ERROR_IO_PENDING) {
		return nil
	}
	if !errors.Is(err, windows.ERROR_PIPE_CONNECTED) {
		return err
	}

	var waitMs uint32
	if timeout > 0 {
		ms := (timeout + 999999) / 1000000
		if ms >= 0xffffffff {
			waitMs = 0xfffffffe
		} else {
			waitMs = uint32(ms)
		}
	}

	result, err := windows.WaitForSingleObject(event, waitMs)
	if err != nil {
		_ = windows.CancelIoEx(pipe, &overlapped)
		return err
	}
	switch result {
	case windows.WAIT_OBJECT_0:
		var bytes uint32
		if err := windows.GetOverlappedResult(pipe, &overlapped, &bytes, false); err != nil {
			return err
		}
		return nil
	case uint32(windows.WAIT_TIMEOUT):
		if err := windows.CancelIoEx(pipe, &overlapped); err != nil && !errors.Is(err, windows.ERROR_NOT_FOUND) {
			return fmt.Errorf("取消更新认证 named pipe 连接失败: %w", err)
		}
		var bytes uint32
		_ = windows.GetOverlappedResult(pipe, &overlapped, &bytes, true)
		return errors.New("等待更新认证 named pipe 连接超时")
	default:
		_ = windows.CancelIoEx(pipe, &overlapped)
		return fmt.Errorf("等待更新认证 named pipe 连接返回未知状态: %d", result)
	}
}

// queryLauncherUpdateProcessIdentity 采集进程身份：PID/父PID/创建时间/会话/SID/完整性/镜像/哈希。
// [S 汇编 0x1408c3d60, 2112B]：OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION|SYNCHRONIZE)
// 失败透传，defer CloseHandle；GetProcessTimes 取创建时间；queryLauncherUpdateParentPID；
// ProcessIdToSessionId；OpenProcessToken(TOKEN_QUERY) 后 defer CloseHandle；GetTokenUser 取
// SID；launcherUpdateTokenIntegrityRID；launcherUpdateProcessImagePath 取镜像路径；
// openLauncherUpdateReadOnlyHandle + GetFileInformationByHandle 取卷序列号/文件索引；
// hashLauncherUpdateFile 取 SHA256。任一失败返回零值身份 + err。
func queryLauncherUpdateProcessIdentity(pid uint32) (launcherUpdateProcessIdentity, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.SYNCHRONIZE, false, pid)
	if err != nil {
		return launcherUpdateProcessIdentity{}, err
	}
	defer windows.CloseHandle(handle)

	var creation, exitTime, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(handle, &creation, &exitTime, &kernel, &user); err != nil {
		return launcherUpdateProcessIdentity{}, err
	}

	parentPID, err := queryLauncherUpdateParentPID(pid)
	if err != nil {
		return launcherUpdateProcessIdentity{}, err
	}

	var sessionID uint32
	if err := windows.ProcessIdToSessionId(pid, &sessionID); err != nil {
		return launcherUpdateProcessIdentity{}, err
	}

	var token windows.Token
	if err := windows.OpenProcessToken(handle, windows.TOKEN_QUERY, &token); err != nil {
		return launcherUpdateProcessIdentity{}, err
	}
	defer windows.CloseHandle(windows.Handle(token))

	tokenUser, err := token.GetTokenUser()
	if err != nil {
		return launcherUpdateProcessIdentity{}, err
	}

	integrityRID, err := launcherUpdateTokenIntegrityRID(token)
	if err != nil {
		return launcherUpdateProcessIdentity{}, err
	}

	imagePath, err := launcherUpdateProcessImagePath(handle)
	if err != nil {
		return launcherUpdateProcessIdentity{}, err
	}

	imageFile, err := openLauncherUpdateReadOnlyHandle(imagePath)
	if err != nil {
		return launcherUpdateProcessIdentity{}, err
	}
	defer windows.CloseHandle(imageFile)

	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(imageFile, &info); err != nil {
		return launcherUpdateProcessIdentity{}, err
	}

	imageHash, err := hashLauncherUpdateFile(imagePath)
	if err != nil {
		return launcherUpdateProcessIdentity{}, err
	}

	return launcherUpdateProcessIdentity{
		PID:           pid,
		ParentPID:     parentPID,
		CreatedAt:     creation.Nanoseconds(),
		SessionID:     sessionID,
		UserSID:       tokenUser.User.Sid.String(),
		IntegrityRID:  integrityRID,
		ImagePath:     imagePath,
		ImageSHA256:   imageHash,
		VolumeSerial:  info.VolumeSerialNumber,
		FileIndexHigh: info.FileIndexHigh,
		FileIndexLow:  info.FileIndexLow,
	}, nil
}
