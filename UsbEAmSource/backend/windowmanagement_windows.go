package main

import (
	"errors"
	"fmt"
	"image"
	"strings"
	"syscall"
	"unsafe"
)

// windowManagement 所需的 user32 LazyProc（依赖 appicon_windows.go 中的 user32DLL）。
var (
	procGetWindowLongPtrW = user32DLL.NewProc("GetWindowLongPtrW")
	procGetClientRect     = user32DLL.NewProc("GetClientRect")
	procGetClassNameW     = user32DLL.NewProc("GetClassNameW")
	procGetWindowTextW    = user32DLL.NewProc("GetWindowTextW")
)

// windowManagementVirtualScreenBounds 返回虚拟屏幕矩形。
// [S] ASM 0x1409edbc0：四次 GetSystemMetrics（0x4c/0x4d/0x4e/0x4f =
// SM_X/Y/CX/CYVIRTUALSCREEN），尾段 edx=x+cx / edi=y+cy 合成
// image.Rect(x, y, x+cx, y+cy)。无回退分支（与 qrCodeVirtualScreenBounds 不同）。
func windowManagementVirtualScreenBounds() image.Rectangle {
	x := getSystemMetrics(0x4c) // SM_XVIRTUALSCREEN
	y := getSystemMetrics(0x4d) // SM_YVIRTUALSCREEN
	w := getSystemMetrics(0x4e) // SM_CXVIRTUALSCREEN
	h := getSystemMetrics(0x4f) // SM_CYVIRTUALSCREEN
	return image.Rect(x, y, x+w, y+h)
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
	return int(rect.Left), int(rect.Top), int(rect.Right-rect.Left), int(rect.Bottom-rect.Top)
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
// [S] ASM 0x1409edfa0：Call(&pt)；成功→(pt.X,pt.Y,nil)；失败→lastErr==nil 时用
// errors.New("未知错误")，再 fmt.Errorf("读取鼠标位置失败: %w",lastErr)，返回 (0,0,err)。
func windowManagementGetCursorPoint() (int32, int32, error) {
	var pt windowManagementPoint
	r1, _, lastErr := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	if r1 != 0 {
		return pt.X, pt.Y, nil
	}
	if lastErr == nil {
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
