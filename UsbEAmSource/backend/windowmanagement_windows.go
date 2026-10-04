package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// windowManagement 所需的 user32 LazyProc（依赖 appicon_windows.go 中的 user32DLL）。
var (
	procGetWindowLongPtrW        = user32DLL.NewProc("GetWindowLongPtrW")
	procGetClientRect            = user32DLL.NewProc("GetClientRect")
	procGetClassNameW            = user32DLL.NewProc("GetClassNameW")
	procGetWindowTextW           = user32DLL.NewProc("GetWindowTextW")
	procIsWindow                 = user32DLL.NewProc("IsWindow")
	procGetWindowRect            = user32DLL.NewProc("GetWindowRect")
	procGetWindowThreadProcessId = user32DLL.NewProc("GetWindowThreadProcessId")
	procSetWindowLongPtrW        = user32DLL.NewProc("SetWindowLongPtrW")
	procSetWindowLongW           = user32DLL.NewProc("SetWindowLongW")
	procMonitorFromWindow        = user32DLL.NewProc("MonitorFromWindow")
	procGetMonitorInfoW          = user32DLL.NewProc("GetMonitorInfoW")
	procEnumDisplayMonitors      = user32DLL.NewProc("EnumDisplayMonitors")
	procSetCursorPos             = user32DLL.NewProc("SetCursorPos")
)

// EnumDisplayMonitors 的回调入口。asm 0x1409edd94 以 `mov rcx,[rip+disp]` 直接读该全局槽
// （槽位 0x141C5AB60 落在 .data 的 BSS 区，即 init 期由 syscall.NewCallback 写入）。
var windowManagementEnumDisplayMonitorCallback = syscall.NewCallback(windowManagementEnumDisplayMonitorProc)

// SetWindowLongPtr 前置清错误码所需的 kernel32.SetLastError
// （x/sys/windows 只导出 GetLastError，无 SetLastError，故自建）。
var (
	kernel32DLL      = windows.NewLazySystemDLL("kernel32.dll")
	procSetLastError = kernel32DLL.NewProc("SetLastError")
)

// windowManagementVirtualScreenBounds 返回虚拟屏幕矩形（4×int32）。
// [S] ASM 0x1409edbc0：四次 GetSystemMetrics（0x4c/0x4d/0x4e/0x4f =
// SM_X/Y/CX/CYVIRTUALSCREEN），尾段 `add edx,ecx` / `lea edi,[rax+rbx]` 为 **32 位**运算，
// 返回 rax/rbx/rcx/rdi = Left/Top/Right/Bottom（int32）。调用方 getLauncherBackgroundMetrics
// 用 movsxd 符号扩展，确证宽度为 int32 而非 image.Rectangle 的 int64。
// 无回退分支（与 qrCodeVirtualScreenBounds 不同）。
func windowManagementVirtualScreenBounds() windowManagementRECT {
	x := getSystemMetrics(0x4c) // SM_XVIRTUALSCREEN
	y := getSystemMetrics(0x4d) // SM_YVIRTUALSCREEN
	w := getSystemMetrics(0x4e) // SM_CXVIRTUALSCREEN
	h := getSystemMetrics(0x4f) // SM_CYVIRTUALSCREEN
	return windowManagementRECT{
		Left:   int32(x),
		Top:    int32(y),
		Right:  int32(x + w),
		Bottom: int32(y + h),
	}
}

// windowManagementPointInCornerGuard 判断坐标 (x,y) 是否落在矩形角落守卫带内。
// [S] ASM 0x1409ed9a0：radius<=0 || right<=left || top>=bottom → false；
// margin=clamp(radius+4,0,96) 再截到 width/height；margin<=0 → false；
// 返回 (左带||右带) && (上带||下带)。
func windowManagementPointInCornerGuard(x, y, left, top, right, bottom, radius int) bool {
	if radius <= 0 || right <= left || top >= bottom {
		return false
	}
	margin := radius + 4
	if margin < 0 {
		margin = 0
	} else if margin > 96 {
		margin = 96
	}
	if margin <= 0 {
		return false
	}
	width := right - left
	height := bottom - top
	if margin > width {
		margin = width
	}
	if margin > height {
		margin = height
	}
	leftBand := x >= left && x < left+margin
	rightBand := x >= right-margin && x < right
	topBand := y >= top && y < top+margin
	bottomBand := y >= bottom-margin && y < bottom
	return (leftBand || rightBand) && (topBand || bottomBand)
}

// windowManagementWrapTargetX 给定垂直坐标 y，找 y 所在显示器，返回 wrap 后的水平目标 x。
// [S] ASM 0x1409eda40：y∈[Top,Bottom) 且 Right>Left；direction=true 时
// best=max(best,Right-1)，否则 best=min(best,Left)；found 时返回 best∓3。
func windowManagementWrapTargetX(monitors []windowManagementRECT, y int32, direction bool) (int32, bool) {
	best := int32(0)
	found := false
	for i := range monitors {
		mon := &monitors[i]
		if mon.Top > y || y >= mon.Bottom {
			continue
		}
		if mon.Right <= mon.Left {
			continue
		}
		if found {
			if direction {
				if best < mon.Right-1 {
					best = mon.Right - 1
				}
			} else {
				if best > mon.Left {
					best = mon.Left
				}
			}
		} else {
			if direction {
				best = mon.Right - 1
			} else {
				best = mon.Left
			}
			found = true
		}
	}
	if found {
		if direction {
			return best - 3, true
		}
		return best + 3, true
	}
	return 0, false
}

// windowManagementWrapTargetY 给定水平坐标 x，找 x 所在显示器，返回 wrap 后的垂直目标 y。
// [S] ASM 0x1409edb00：x∈[Left,Right) 且 Bottom>Top；direction=true 时
// best=max(best,Bottom-1)，否则 best=min(best,Top)；found 时返回 best∓3。
func windowManagementWrapTargetY(monitors []windowManagementRECT, x int32, direction bool) (int32, bool) {
	best := int32(0)
	found := false
	for i := range monitors {
		mon := &monitors[i]
		if mon.Left > x || x >= mon.Right {
			continue
		}
		if mon.Bottom <= mon.Top {
			continue
		}
		if found {
			if direction {
				if best < mon.Bottom-1 {
					best = mon.Bottom - 1
				}
			} else {
				if best > mon.Top {
					best = mon.Top
				}
			}
		} else {
			if direction {
				best = mon.Bottom - 1
			} else {
				best = mon.Top
			}
			found = true
		}
	}
	if found {
		if direction {
			return best - 3, true
		}
		return best + 3, true
	}
	return 0, false
}

// windowManagementGetWindowLongPtr 获取窗口 LongPtr 值。
// [S] ASM 0x1409ee320：显式 Find 预加载 GetWindowLongPtrW（失败→nil→Call panic），
// 打包 [2]uintptr{hwnd,index}，LazyProc.Call(2 args)，直接返回 r1。
func windowManagementGetWindowLongPtr(hwnd uintptr, index int32) uintptr {
	proc := procGetWindowLongPtrW
	if proc.Find() != nil {
		proc = nil
	}
	r, _, _ := proc.Call(hwnd, uintptr(index))
	return r
}

// windowManagementGetClientRectValue 获取窗口客户区矩形，返回 (left,top,width,height)。
// [S] ASM 0x1409ee780：GetClientRect(hwnd,&rect)；r1==0 → (0,0,0,0)；
// 否则 (Left, Top, Right-Left, Bottom-Top)（int32 符号扩展为 int）。
func windowManagementGetClientRectValue(hwnd uintptr) (int, int, int, int) {
	var rect windowManagementRECT
	r1, _, _ := procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
	if r1 == 0 {
		return 0, 0, 0, 0
	}
	return int(rect.Left), int(rect.Top), int(rect.Right - rect.Left), int(rect.Bottom - rect.Top)
}

// windowManagementGetClassName 获取窗口类名（GetClassNameW 单阶段，256 缓冲）。
// [S] ASM 0x1409ee220：hwnd==0→空串；make([]uint16,256)→Call(hwnd,&buf[0],256)；
// n==0→空串；否则 TrimSpace(UTF16ToString(buf[:n]))。
func windowManagementGetClassName(hwnd uintptr) string {
	if hwnd == 0 {
		return ""
	}
	buf := make([]uint16, 256)
	n, _, _ := procGetClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 256)
	if n == 0 {
		return ""
	}
	return strings.TrimSpace(syscall.UTF16ToString(buf[:int(n)]))
}

// windowManagementGetWindowText 获取窗口标题（GetWindowTextW 两阶段：先取长度再取内容）。
// [S] ASM 0x1409ee0e0：hwnd==0→空串；Call(hwnd,0,0) 取长度 n；n==0→空串；
// make([]uint16,n+1)→Call(hwnd,&buf[0],n+1)；n2==0→空串；否则 TrimSpace(UTF16ToString(buf[:n2]))。
func windowManagementGetWindowText(hwnd uintptr) string {
	if hwnd == 0 {
		return ""
	}
	n, _, _ := procGetWindowTextW.Call(hwnd, 0, 0)
	if n == 0 {
		return ""
	}
	buf := make([]uint16, int(n+1))
	n2, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), n+1)
	if n2 == 0 {
		return ""
	}
	return strings.TrimSpace(syscall.UTF16ToString(buf[:int(n2)]))
}

// windowManagementGetCursorPoint 获取鼠标屏幕坐标（GetCursorPos）。
// [S] ASM 0x1409edfa0：Call(&pt)；成功→(pt.X,pt.Y,nil)；失败→lastErr==nil 或
// lastErr==syscall.Errno(0)（asm 里 cmp ErrnoItab + data==0）时用 errors.New("未知错误")，
// 再 fmt.Errorf("读取鼠标位置失败: %w",lastErr)，返回 (0,0,err)。
func windowManagementGetCursorPoint() (int32, int32, error) {
	var pt windowManagementPoint
	r1, _, lastErr := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	if r1 != 0 {
		return pt.X, pt.Y, nil
	}
	if lastErr == nil || errors.Is(lastErr, syscall.Errno(0)) {
		lastErr = errors.New("未知错误")
	}
	return 0, 0, fmt.Errorf("读取鼠标位置失败: %w", lastErr)
}

// windowManagementEnumDisplayMonitorProc 是 EnumDisplayMonitors 的回调：
// 收集有效显示器矩形到 dwData 指向的 []windowManagementRECT。
// [S] ASM 0x1409ede60：dwData==0 || lprcMonitor==nil → 返回 1；
// 矩形无效（Right<=Left 或 Bottom<=Top）→ 返回 1；否则 append(*slice, rect) 后返回 1。
func windowManagementEnumDisplayMonitorProc(hMonitor, hdcMonitor uintptr, lprcMonitor *windowManagementRECT, dwData unsafe.Pointer) uintptr {
	if dwData == nil || lprcMonitor == nil {
		return 1
	}
	r := *lprcMonitor
	if r.Right <= r.Left || r.Bottom <= r.Top {
		return 1
	}
	rects := (*[]windowManagementRECT)(dwData)
	*rects = append(*rects, r)
	return 1
}

// windowManagementGetWindowRect 获取窗口屏幕矩形（GetWindowRect）。
// [S] ASM 0x1409ee600：Call(hwnd,&rect)；成功→(Left,Top,Right,Bottom,nil)；
// 失败→lastErr==nil 或 Errno(0) 时用 errors.New("未知错误")，
// 再 fmt.Errorf("读取窗口位置失败: %w",lastErr)，返回 (0,0,0,0,err)。
func windowManagementGetWindowRect(hwnd uintptr) (int32, int32, int32, int32, error) {
	var rect windowManagementRECT
	r1, _, lastErr := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
	if r1 != 0 {
		return rect.Left, rect.Top, rect.Right, rect.Bottom, nil
	}
	if lastErr == nil || errors.Is(lastErr, syscall.Errno(0)) {
		lastErr = errors.New("未知错误")
	}
	return 0, 0, 0, 0, fmt.Errorf("读取窗口位置失败: %w", lastErr)
}

// windowManagementValidateSnapshotOwner 校验快照记录的窗口是否仍属于原进程。
// [S] ASM 0x1409ea780：hwnd==0 或 IsWindow(hwnd)==0 → errors.New("目标窗口已失效")；
// pid==0 → nil；GetWindowThreadProcessId(hwnd,&pid2)；pid2!=pid →
// errors.New("目标窗口句柄已被其他进程复用")；否则 nil。
func windowManagementValidateSnapshotOwner(hwnd uintptr, pid uint32) error {
	if hwnd == 0 {
		return errors.New("目标窗口已失效")
	}
	if r, _, _ := procIsWindow.Call(hwnd); r == 0 {
		return errors.New("目标窗口已失效")
	}
	if pid == 0 {
		return nil
	}
	var pid2 uint32
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid2)))
	if pid2 != pid {
		return errors.New("目标窗口句柄已被其他进程复用")
	}
	return nil
}

// windowManagementFlagName 是样式名查表的条目：Flag 为位掩码，Name 为对应常量名。
type windowManagementFlagName struct {
	Flag uint32
	Name string
}

// windowManagementFlagNames 把窗口样式位掩码翻译为可读名字串（以 " | " 连接）。
// [S] ASM 0x1409eede0：value==0 → ""；遍历表，Flag==0 或 value&Flag!=Flag 跳过，
// 否则收集 Name；末尾 strings.Join(names, " | ")。
func windowManagementFlagNames(value uint32, names []windowManagementFlagName) string {
	if value == 0 {
		return ""
	}
	flagNames := make([]string, 0, len(names))
	for _, e := range names {
		if e.Flag == 0 || value&e.Flag != e.Flag {
			continue
		}
		flagNames = append(flagNames, e.Name)
	}
	return strings.Join(flagNames, " | ")
}

// windowManagementStyleNames 返回窗口样式（WS_*）的可读名字串。
// [S] ASM 0x1409eebe0：duffcopy 把 17 条静态表（408B）复制到栈，调用 windowManagementFlagNames。
func windowManagementStyleNames(value uint32) string {
	return windowManagementFlagNames(value, []windowManagementFlagName{
		{0x08000000, "WS_POPUP"},
		{0x04000000, "WS_CHILD"},
		{0x02000000, "WS_MINIMIZE"},
		{0x01000000, "WS_VISIBLE"},
		{0x00800000, "WS_DISABLED"},
		{0x00400000, "WS_CLIPSIBLINGS"},
		{0x00200000, "WS_CLIPCHILDREN"},
		{0x00100000, "WS_MAXIMIZE"},
		{0x000C0000, "WS_CAPTION"},
		{0x00080000, "WS_BORDER"},
		{0x00040000, "WS_DLGFRAME"},
		{0x00020000, "WS_VSCROLL"},
		{0x00010000, "WS_HSCROLL"},
		{0x00008000, "WS_SYSMENU"},
		{0x00004000, "WS_THICKFRAME"},
		{0x00002000, "WS_MINIMIZEBOX"},
		{0x00001000, "WS_MAXIMIZEBOX"},
	})
}

// windowManagementExStyleNames 返回扩展窗口样式（WS_EX_*）的可读名字串。
// [S] ASM 0x1409eec60：duffzero 清栈后逐条填入 8 条静态表（192B），调用 windowManagementFlagNames。
func windowManagementExStyleNames(value uint32) string {
	return windowManagementFlagNames(value, []windowManagementFlagName{
		{0x00000008, "WS_EX_TOPMOST"},
		{0x00000020, "WS_EX_TRANSPARENT"},
		{0x00000080, "WS_EX_TOOLWINDOW"},
		{0x00000100, "WS_EX_WINDOWEDGE"},
		{0x00000200, "WS_EX_CLIENTEDGE"},
		{0x00040000, "WS_EX_APPWINDOW"},
		{0x00080000, "WS_EX_LAYERED"},
		{0x08000000, "WS_EX_NOACTIVATE"},
	})
}

// windowManagementSetWindowLongPtr 设置窗口样式/扩展样式等 LongPtr 值。
// [S] ASM 0x1409ee3e0：runtime.LockOSThread + defer runtime.UnlockOSThread（funcval[0]=0x1400525e0）；
// SetLastError.Call(0) 清错误码；默认用 SetWindowLongPtrW，其 Find() 失败则退回 SetWindowLongW；
// Call(hwnd, uintptr(index), newValue)：r1!=0 → nil；lastErr 为 syscall.Errno(0) → nil；
// lastErr==nil → errors.New("未知错误")；否则 fmt.Errorf("更新窗口样式失败: %w", lastErr)。
func windowManagementSetWindowLongPtr(hwnd uintptr, index int32, newValue uintptr) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	procSetLastError.Call(0)
	proc := procSetWindowLongPtrW
	if procSetWindowLongPtrW.Find() != nil {
		proc = procSetWindowLongW
	}
	r1, _, lastErr := proc.Call(hwnd, uintptr(index), newValue)
	if r1 != 0 {
		return nil
	}
	if errno, ok := lastErr.(syscall.Errno); ok && errno == 0 {
		return nil
	}
	if lastErr == nil {
		lastErr = errors.New("未知错误")
	}
	return fmt.Errorf("更新窗口样式失败: %w", lastErr)
}

// windowManagementMonitorRectForWindow 返回窗口所在显示器的屏幕矩形。
// [S] ASM 0x1409ee840：MonitorFromWindow(hwnd, 2=MONITOR_DEFAULTTONEAREST)；hMonitor==0 →
// errors.New("无法定位目标窗口所在显示器")；GetMonitorInfoW(hMonitor,&mi)（Size=40）；
// 失败→lastErr==nil 或 Errno(0) 时 errors.New("未知错误")，再
// fmt.Errorf("读取显示器边界失败: %w",lastErr)；成功→Monitor 矩形的 (Left,Top,Right,Bottom,nil)。
func windowManagementMonitorRectForWindow(hwnd uintptr) (int32, int32, int32, int32, error) {
	hMonitor, _, _ := procMonitorFromWindow.Call(hwnd, 2)
	if hMonitor == 0 {
		return 0, 0, 0, 0, errors.New("无法定位目标窗口所在显示器")
	}
	mi := &windowManagementMonitorInfo{Size: 40}
	r1, _, lastErr := procGetMonitorInfoW.Call(hMonitor, uintptr(unsafe.Pointer(mi)))
	if r1 == 0 {
		if lastErr == nil || errors.Is(lastErr, syscall.Errno(0)) {
			lastErr = errors.New("未知错误")
		}
		return 0, 0, 0, 0, fmt.Errorf("读取显示器边界失败: %w", lastErr)
	}
	return mi.Monitor.Left, mi.Monitor.Top, mi.Monitor.Right, mi.Monitor.Bottom, nil
}

// windowManagementDisplayRects 枚举所有显示器矩形；失败或为空时回退为虚拟屏幕单矩形。
// [S] ASM 0x1409edce0：
//  1. GetSystemMetrics(0x50=SM_CMONITORS)，`test rax,rax; mov ecx,1; cmovle rax,rcx`
//     → count = max(r1, 1)；
//  2. makeslice(len=0, cap=count) 得 rects（持有于栈上 slice 头）；
//  3. EnumDisplayMonitors(0, 0, callback, &rects)（4 实参，回调经 BSS 全局槽传入）；
//  4. r1!=0 && len(rects)!=0 → 原样返回 rects；
//  5. 否则取 VirtualScreenBounds：Right<=Left 或 Bottom<=Top → 返回 nil；
//  6. 有效 → newobject 写入 4×int32 后返回单元素切片（rax/rbx=1/rcx=1）。
func windowManagementDisplayRects() []windowManagementRECT {
	count := getSystemMetrics(0x50) // SM_CMONITORS
	if count <= 0 {
		count = 1
	}
	rects := make([]windowManagementRECT, 0, count)
	r1, _, _ := procEnumDisplayMonitors.Call(
		0, 0,
		windowManagementEnumDisplayMonitorCallback,
		uintptr(unsafe.Pointer(&rects)),
	)
	if r1 != 0 && len(rects) != 0 {
		return rects
	}
	bounds := windowManagementVirtualScreenBounds()
	if bounds.Right <= bounds.Left || bounds.Bottom <= bounds.Top {
		return nil
	}
	return []windowManagementRECT{bounds}
}

// windowManagementWrappedCursorPoint 计算鼠标环绕后的新坐标。
// [S] ASM 0x1409ed5c0（0x3e0=992B）：
//
//  1. 线性扫描 monitors（stride 0x10）找首个满足 Left<=x<Right && Top<=y<Bottom 的矩形，
//     命中存 left/top/right/bottom；未命中（或 Right<=Left || Top>=Bottom）→ 返回 (x,y,false)。
//  2. windowManagementPointInCornerGuard(x,y,left,top,right,bottom,guardPx) 命中 → (x,y,false)。
//  3. 水平：wrapX && x<=left 时探测 (x-1,y)；该点不属于任何 monitor →
//     WrapTargetX(monitors,y,true)（最右）。wrapX && x>=right-1 时探测 (x+1,y)；
//     不属于任何 monitor → WrapTargetX(monitors,y,false)（最左）。
//  4. 垂直：wrapY && y<=top 时探测 (x,y-1)；不属于任何 monitor →
//     WrapTargetY(monitors,x,true)（最下）。wrapY && y>=bottom-1 时探测 (x,y+1)；
//     不属于任何 monitor → WrapTargetY(monitors,x,false)（最上）。
//  5. 水平与垂直均基于**原始 (x,y)** 判定（0x1409ed736 / 0x1409ed76a 读 [rsp+0x38] 原始 x，
//     非已 wrap 的 newX）；返回 ok = hOK | vOK（0x1409ed7c1 `or ecx,ebx`）。
func windowManagementWrappedCursorPoint(x, y int32, monitors []windowManagementRECT, wrapX, wrapY bool, guardPx int) (int32, int32, bool) {
	var left, top, right, bottom int32
	found := false
	for i := range monitors {
		mon := &monitors[i]
		if x < mon.Left || x >= mon.Right || y < mon.Top || y >= mon.Bottom {
			continue
		}
		left, top, right, bottom = mon.Left, mon.Top, mon.Right, mon.Bottom
		found = true
		break
	}
	if !found || right <= left || top >= bottom {
		return x, y, false
	}
	if windowManagementPointInCornerGuard(int(x), int(y), int(left), int(top), int(right), int(bottom), guardPx) {
		return x, y, false
	}
	newX, ok := x, false
	if wrapX {
		skipRight := false
		if x <= left {
			onAdjacent := false
			for i := range monitors {
				mon := &monitors[i]
				if x-1 < mon.Left || x-1 >= mon.Right || y < mon.Top || y >= mon.Bottom {
					continue
				}
				onAdjacent = true
				break
			}
			if !onAdjacent {
				if nx, hit := windowManagementWrapTargetX(monitors, y, true); hit {
					newX = nx
					ok = true
				}
				skipRight = true
			}
		}
		if !skipRight && x >= right-1 {
			onAdjacent := false
			for i := range monitors {
				mon := &monitors[i]
				if x+1 < mon.Left || x+1 >= mon.Right || y < mon.Top || y >= mon.Bottom {
					continue
				}
				onAdjacent = true
				break
			}
			if !onAdjacent {
				if nx, hit := windowManagementWrapTargetX(monitors, y, false); hit {
					newX = nx
					ok = true
				}
			}
		}
	}
	newY := y
	if wrapY {
		skipBottom := false
		if y <= top {
			onAdjacent := false
			for i := range monitors {
				mon := &monitors[i]
				if x < mon.Left || x >= mon.Right || y-1 < mon.Top || y-1 >= mon.Bottom {
					continue
				}
				onAdjacent = true
				break
			}
			if !onAdjacent {
				if ny, hit := windowManagementWrapTargetY(monitors, x, true); hit {
					newY = ny
					ok = true
				}
				skipBottom = true
			}
		}
		if !skipBottom && y >= bottom-1 {
			onAdjacent := false
			for i := range monitors {
				mon := &monitors[i]
				if x < mon.Left || x >= mon.Right || y+1 < mon.Top || y+1 >= mon.Bottom {
					continue
				}
				onAdjacent = true
				break
			}
			if !onAdjacent {
				if ny, hit := windowManagementWrapTargetY(monitors, x, false); hit {
					newY = ny
					ok = true
				}
			}
		}
	}
	return newX, newY, ok
}

// windowManagementMaybeWrapCursor 在启用环绕时把鼠标移到对侧屏幕边缘。
// [S] ASM 0x1409ed4a0（0x120=288B）：GetCursorPoint 失败 → false；
// WrappedCursorPoint(x,y,monitors,wrapX,wrapY,guardPx) 返回 ok=false → false；
// 否则 newobject([2]uintptr) 打包 movsxd 符号扩展后的 (newX,newY)，
// procSetCursorPos.Call(newX,newY)（LazyProc 槽 0x141BC1DC0，内存实证 Name='SetCursorPos'），
// 无条件返回 true（SetCursorPos 的返回值被丢弃）。
func windowManagementMaybeWrapCursor(wrapX, wrapY bool, monitors []windowManagementRECT, guardPx int) bool {
	x, y, err := windowManagementGetCursorPoint()
	if err != nil {
		return false
	}
	newX, newY, ok := windowManagementWrappedCursorPoint(x, y, monitors, wrapX, wrapY, guardPx)
	if !ok {
		return false
	}
	procSetCursorPos.Call(uintptr(newX), uintptr(newY))
	return true
}

// targetFromWindowProcessPick 由窗口拾取结果组装 WindowManagementTarget。
// [S] ASM 0x1409eea40（0x1a0=416B）：12 个参数字（9 寄存器 + 3 栈，实参槽
// [rsp+0x40]/[rsp+0x50]/[rsp+0x58..0x80]/[rsp+0x88]/[rsp+0x90..0x98]）；
// duffzero+0x142 清零 128 字节 = WindowManagementTarget（types_windowmgt.go L68-78）；
// Path=TrimSpace(processPath)，ProcessName=TrimSpace(processName) 或 filepath.Base(Path)
// （0x1409eeab8，编译符号 internal/filepathlite.Base），DisplayName=TrimSpace(displayName)
// 或 TrimSpace(title) 或 ProcessName（0x1409eeb7f `cmove` 两级回退），
// Title=**原始** title（0x1409eeb62 直读 [rsp+0x78]/[rsp+0x80]，不 Trim）；
// IconRef/IconURL 保持清零空串。
func targetFromWindowProcessPick(
	processPath string,
	processID uint32,
	processName string,
	displayName string,
	title string,
	hwnd uintptr,
	iconData string,
) WindowManagementTarget {
	trimmedPath := strings.TrimSpace(processPath)
	trimmedName := strings.TrimSpace(processName)
	if trimmedName == "" && trimmedPath != "" {
		trimmedName = filepath.Base(trimmedPath)
	}
	trimmedDisplay := strings.TrimSpace(displayName)
	if trimmedDisplay == "" {
		trimmedDisplay = strings.TrimSpace(title)
	}
	if trimmedDisplay == "" {
		trimmedDisplay = trimmedName
	}
	return WindowManagementTarget{
		HWND:        hwnd,
		ProcessID:   processID,
		ProcessName: trimmedName,
		Path:        trimmedPath,
		Title:       title,
		DisplayName: trimmedDisplay,
		IconData:    iconData,
	}
}
