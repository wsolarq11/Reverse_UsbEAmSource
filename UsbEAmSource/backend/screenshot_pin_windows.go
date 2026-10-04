// screenshot_pin_windows.go — 截图钉窗口原生样式/透明度层（逆向还原）
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)

package main

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// ========================================================================
// Windows API 代理（DWMAPI + USER32 延迟加载）
// ========================================================================

var (
	dwmapiDLL = windows.NewLazySystemDLL("dwmapi.dll")

	procDwmSetWindowAttribute      = dwmapiDLL.NewProc("DwmSetWindowAttribute")
	procSetLayeredWindowAttributes = user32DLL.NewProc("SetLayeredWindowAttributes")
)

// DWMWA_BORDER_COLOR：DwmSetWindowAttribute 的边框颜色属性。
// DWMWA_COLOR_NONE：边框颜色属性取该值时禁用窗口边框（"抑制边框"）。
const (
	dwmwaBorderColor uint32  = 0x22       // DWMWA_BORDER_COLOR
	dwmwaColorNone   uint32  = 0xFFFFFFFE // DWMWA_COLOR_NONE
	lwaAlpha         uintptr = 0x2        // LWA_ALPHA：使用 bAlpha 参数
)

// screenshotPinSetLayeredWindowOpacity 设置钉窗口分层不透明度。
// [S] 反汇编实证 0x1409944a0, 256B：
//
//	hwnd==0 直接返回；opacity 夹到 [0.2, 1.0]（0.2/1.0 常量 @0x1411CD658/0x1411CD6D0）；
//	opacity<1.0 时 alpha=max(1, int(opacity*255))，否则 alpha=255（0xffffffff 低字节）；
//	随后 SetLayeredWindowAttributes(hwnd, 0, alpha, LWA_ALPHA=2)（proc @0x141BC1CB0）。
func screenshotPinSetLayeredWindowOpacity(hwnd uintptr, opacity float64) {
	if hwnd == 0 {
		return
	}
	if opacity < 0.2 {
		opacity = 0.2
	} else if opacity > 1.0 {
		opacity = 1.0
	}
	var alpha byte
	if opacity < 1.0 {
		a := int(opacity * 255)
		if a <= 0 {
			a = 1
		}
		alpha = byte(a)
	} else {
		alpha = 255
	}
	procSetLayeredWindowAttributes.Call(hwnd, 0, uintptr(alpha), lwaAlpha)
}

// screenshotPinSuppressWindowBorder 抑制钉窗口边框。
// [S] 反汇编实证 0x1409945a0, 160B：
//
//	hwnd==0 直接返回；否则 DwmSetWindowAttribute(hwnd, DWMWA_BORDER_COLOR=0x22,
//	&DWMWA_COLOR_NONE(-2), 4)（proc @0x141BC2658，颜色值 0xfffffffe 实证）。
func screenshotPinSuppressWindowBorder(hwnd uintptr) {
	if hwnd == 0 {
		return
	}
	attr := dwmwaColorNone
	procDwmSetWindowAttribute.Call(hwnd, uintptr(dwmwaBorderColor), uintptr(unsafe.Pointer(&attr)), 4)
}
