package main

import "github.com/wailsapp/wails/v3/pkg/application"

// screenshotSelectionToolbarShowWindowCommand 返回选区工具栏 ShowWindow 命令。
// [S] ASM 0x1409ad200: mov eax, 8; ret（SW_SHOWNA）。
func screenshotSelectionToolbarShowWindowCommand() int {
	return 8
}

// screenshotSelectionToolbarShowWindowFlags 返回选区工具栏显示 flags。
// [S] ASM 0x1409ad220: mov eax, 0x53; ret（返回常量 0x53=83）。
func screenshotSelectionToolbarShowWindowFlags() int {
	return 0x53
}

// screenshotSelectionToolbarRevealWindow 揭示选区工具栏窗口。
// [S-sig 0x1409ad240, 224B]：签名实证——(window application.Window, activate bool)。
// activate=true 经 itab fun[0x248] 直接激活；false 走 application.InvokeSync(func1
// ShowWindowNoActivate：取句柄→IsWindowVisible→ShowWindow→SetWindowPos)。体待
// wails 窗口句柄接口方法名专项还原。
func screenshotSelectionToolbarRevealWindow(window application.Window, activate bool) {
	_, _ = window, activate
}

// screenshotSelectionToolbarBringToTop 置顶选区工具栏窗口。
// [S-sig 0x1409ad3a0, 160B]：签名实证——(window application.Window)。
// nil 接口直接返回；否则 application.InvokeSync(func1 0x1409ad440 ShowWindowNoActivate：
// 取句柄→IsWindowVisible 不可见则 ShowWindow(8)→SetWindowPos(HWND_TOP,0x53))。
// 体待 wails 窗口句柄接口方法名专项还原。
func screenshotSelectionToolbarBringToTop(window application.Window) {
	_ = window
}

// SetMessages 设置选择工具栏消息（持锁，normalize 后写入 field(+0x68)）。
// [S-sig 0x1409a6780, 352B]：lock(+0x8) → normalizeScreenshotSelectionToolbarMessages →
// 写入 field(+0x68)。体待消息结构专项还原。
func (s *screenshotSelectionToolbarWindowService) SetMessages(a interface{}) {
	_, _ = s, a
}
