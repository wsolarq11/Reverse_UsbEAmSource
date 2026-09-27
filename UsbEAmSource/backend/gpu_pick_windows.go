// gpu_pick_windows.go — GPU 偏好拖放目标窗口拾取（逆向还原）
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)
// 基线：docs/goresym/source_funcs.txt + docs/goresym/pipeline/tmp/*.asm.txt

package main

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/w32"
	"golang.org/x/sys/windows"
)

// 拖放窗口拾取所需的 user32 LazyProc（依赖 appicon_windows.go 中的 user32DLL）。
var (
	procGetCursorPos        = user32DLL.NewProc("GetCursorPos")
	procWindowFromPoint     = user32DLL.NewProc("WindowFromPoint")
	procGetForegroundWindow = user32DLL.NewProc("GetForegroundWindow")
	procGetAncestor         = user32DLL.NewProc("GetAncestor")
)

// getCursorScreenPoint 读取当前鼠标屏幕坐标。
// [S] ASM 0x14085cf00: GetCursorPos LazyProc(1 arg)；r1!=0 → (pt.X,pt.Y,nil)；
// lastErr==nil || errors.Is(lastErr, syscall.Errno(0)) → errors.New("无法读取当前鼠标位置")；
// 否则 (0,0,lastErr)。
func getCursorScreenPoint() (int32, int32, error) {
	var pt w32.POINT
	r1, _, lastErr := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	if r1 != 0 {
		return pt.X, pt.Y, nil
	}
	if lastErr == nil || errors.Is(lastErr, syscall.Errno(0)) {
		return 0, 0, errors.New("无法读取当前鼠标位置")
	}
	return 0, 0, lastErr
}

// windowFromScreenPoint 返回指定屏幕坐标所在的窗口句柄。
// [S] ASM 0x14085d000: 打包 w32.POINT{X,Y}（x|y<<32），WindowFromPoint LazyProc(1 arg)；
// hwnd!=0 → (hwnd,nil)；lastErr==nil || Errno(0) → errors.New("无法识别鼠标所在窗口")；
// 否则 (0,lastErr)。
func windowFromScreenPoint(x, y int32) (uintptr, error) {
	pt := w32.POINT{X: x, Y: y}
	hwnd, _, lastErr := procWindowFromPoint.Call(uintptr(unsafe.Pointer(&pt)))
	if hwnd != 0 {
		return hwnd, nil
	}
	if lastErr == nil || errors.Is(lastErr, syscall.Errno(0)) {
		return 0, errors.New("无法识别鼠标所在窗口")
	}
	return 0, lastErr
}

// getForegroundWindow 返回前台窗口句柄。
// [S] ASM 0x14085d100: GetForegroundWindow LazyProc(0 arg)；hwnd!=0 → (hwnd,nil)；
// lastErr==nil || Errno(0) → (0,nil)；否则 (0,lastErr)。
func getForegroundWindow() (uintptr, error) {
	hwnd, _, lastErr := procGetForegroundWindow.Call()
	if hwnd != 0 {
		return hwnd, nil
	}
	if lastErr == nil || errors.Is(lastErr, syscall.Errno(0)) {
		return 0, nil
	}
	return 0, lastErr
}

// getWindowAncestor 返回窗口的祖先窗口。
// [S] ASM 0x14085d1a0: hwnd==0 → 0；否则 GetAncestor LazyProc(2 args)，直接返回 result。
func getWindowAncestor(hwnd uintptr, flags uintptr) uintptr {
	if hwnd == 0 {
		return 0
	}
	result, _, _ := procGetAncestor.Call(hwnd, flags)
	return result
}

// resolveProcessPathByPID 通过进程 ID 解析可执行文件完整路径。
// [S] ASM 0x14085c740: OpenProcess(0x1000=PROCESS_QUERY_LIMITED_INFORMATION,false,pid) →
// defer CloseHandle；QueryFullProcessImageName(handle,0,&buf,&size)；err→return；
// size==0 → errors.New("无法读取目标进程路径")；UTF16ToString 空 → errors.New("目标进程路径为空")；
// 否则返回 path。
func resolveProcessPathByPID(pid uint32) (string, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(handle)

	var buf [0x8000]uint16
	size := uint32(len(buf))
	if err := windows.QueryFullProcessImageName(handle, 0, &buf[0], &size); err != nil {
		return "", err
	}
	if size == 0 {
		return "", errors.New("无法读取目标进程路径")
	}
	path := windows.UTF16ToString(buf[:size])
	if path == "" {
		return "", errors.New("目标进程路径为空")
	}
	return path, nil
}

// resolveProcessNameByPID 通过进程 ID 解析进程名。
// [S] ASM 0x14085cae0: pid==0 → errors.New("进程 ID 无效")；
// CreateToolhelp32Snapshot(TH32CS_SNAPPROCESS,0) → defer CloseHandle；
// Process32First 遍历：ProcessID==pid → TrimSpace(UTF16ToString(ExeFile))，空 →
// errors.New("目标进程名为空")；Process32Next 耗尽 → errors.New("未找到目标进程")。
func resolveProcessNameByPID(pid uint32) (string, error) {
	if pid == 0 {
		return "", errors.New("进程 ID 无效")
	}
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	if err := windows.Process32First(snapshot, &entry); err != nil {
		return "", err
	}
	for {
		if entry.ProcessID == pid {
			name := strings.TrimSpace(windows.UTF16ToString(entry.ExeFile[:]))
			if name == "" {
				return "", errors.New("目标进程名为空")
			}
			return name, nil
		}
		if err := windows.Process32Next(snapshot, &entry); err != nil {
			return "", errors.New("未找到目标进程")
		}
	}
}

// resolveWindowProcessPath 解析窗口对应的进程路径、进程 ID。
// [S] ASM 0x14085c540: hwnd==0 → errors.New("窗口句柄无效")；
// GetWindowThreadProcessId(hwnd,&pid)；err→return；pid==0 →
// errors.New("无法识别目标窗口对应的进程")；resolveProcessPathByPID；err→return；
// 成功日志 gpuPickDebugLog("窗口已解析到进程: hwnd=%s pid=%d path=%q",...)，返回 (path,pid,nil)。
func resolveWindowProcessPath(hwnd uintptr) (string, uint32, error) {
	if hwnd == 0 {
		return "", 0, errors.New("窗口句柄无效")
	}
	var pid uint32
	if _, err := windows.GetWindowThreadProcessId(windows.HWND(hwnd), &pid); err != nil {
		return "", 0, err
	}
	if pid == 0 {
		return "", 0, errors.New("无法识别目标窗口对应的进程")
	}
	path, err := resolveProcessPathByPID(pid)
	if err != nil {
		return "", pid, err
	}
	gpuPickDebugLog("窗口已解析到进程: hwnd=%s pid=%d path=%q", fmt.Sprintf("0x%X", hwnd), pid, path)
	return path, pid, nil
}
