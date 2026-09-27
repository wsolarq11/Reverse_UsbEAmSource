package main

import "github.com/wailsapp/wails/v3/pkg/application"

// setupLauncherTray 构建系统托盘（图标 + 提示 + 菜单：显示主窗、退出等）。
// [S-sig 汇编 0x1408b2fa0, 832B]：体依赖 Wails SystemTray 完整菜单构建（3 个闭包 + InvokeSync）。
// 序言保存 2 字参数：rax=含 app 字段的结构体指针（[rax+0x320] 传给 SystemTrayManager.New）、
// rbx=第二指针（闭包上下文 [8]，showLauncherFromTray 的 receiver）；体待还原。
func setupLauncherTray(app *application.App, bs *BootstrapService) error { return nil }
