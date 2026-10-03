// desktopwidgets_notification.go — 桌面小部件通知器平台分派（逆向还原）
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)

package main

// platformDesktopWidgetNotifierNotify 桌面小部件通知器的平台分派。
// [S] 反汇编实证：Windows 构建中 VA=0（编译期死代码消除，因 Notify 方法直连
// showDesktopWidgetNotification 不经由本函数），语义为无副作用空实现。
func platformDesktopWidgetNotifierNotify(title, body string, silent bool) {
}
