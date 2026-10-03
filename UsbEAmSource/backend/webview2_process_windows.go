package main

import (
	"os"
	"path/filepath"
	"strings"
)

// resolveWebView2ProcessInspectUserDataDir 把 WebView2 用户数据目录规范化为绝对路径。
// [S] ASM 0x1409df7e0：TrimSpace → 空串直接返回 ("", nil)；否则交 filepath.Abs
// （Abs 自身返回 (string, error)，err 非 nil 时返回 ("", err)）。
func resolveWebView2ProcessInspectUserDataDir(dir string) (string, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return "", nil
	}
	return filepath.Abs(dir)
}

// resolveWebView2ProcessInspectHostExeName 返回宿主程序文件名；取不到或为空时回落默认名。
// [S] ASM 0x1409df840：os.Executable 出错 → 默认名；否则 filepath.Base + TrimSpace，
// 空串亦回落默认名。返回纯 string（错误被吞并回落）。
// 默认名实测为 "UsbEAm_Launcher.exe"（19B @0x140C5BD00；两条分支 strA/strB 同址，交叉验证一致）。
func resolveWebView2ProcessInspectHostExeName() string {
	const defaultName = "UsbEAm_Launcher.exe"
	exe, err := os.Executable()
	if err != nil {
		return defaultName
	}
	name := strings.TrimSpace(filepath.Base(exe))
	if name == "" {
		return defaultName
	}
	return name
}

// inspectWebView2Processes 枚举 WebView2 进程并构造归一化快照。
// [P] 存根：平台层进程枚举（0x1409debc0, 3104B，依赖 CreateToolhelp32Snapshot/
// OpenProcess/读命令行等 Windows API）待专项批次还原。签名已实证：
// func inspectWebView2Processes(userDataDir string) WebView2ProcessSnapshot。
func inspectWebView2Processes(userDataDir string) WebView2ProcessSnapshot {
	return WebView2ProcessSnapshot{}
}
