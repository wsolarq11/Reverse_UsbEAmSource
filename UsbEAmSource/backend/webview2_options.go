package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// resolveLauncherWebViewUserDataPath 解析 WebView2 用户数据目录根路径。
// [S 汇编 0x1409dd840]：strings.TrimSpace → 空则 errors.New("WebView2 用户数据目录根路径不能为空")
// （@0x140c86df3,48B）→ filepath.Abs → os.MkdirAll(abs, 0o755=0x1ed)；任一步失败返回 err，
// 成功返回 (abs, nil)。
func resolveLauncherWebViewUserDataPath(path string) (string, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", errors.New("WebView2 用户数据目录根路径不能为空")
	}
	abs, err := filepath.Abs(trimmed)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return "", err
	}
	return abs, nil
}

// buildLauncherBrowserArgs 构建浏览器启动参数。
// [S-sig 汇编 0x1409dd920] 返回 []string（三字 slice）。参数为 0x448 字节结构（栈传递，
// 序言从 rsp+0x1148/0x1150 读，体首部 normalizePreferencesWithOptions(空,nil)），参数类型未落地，
// 暂以 LauncherConfig 占位，体待还原。
func buildLauncherBrowserArgs(opts LauncherConfig) []string { return nil }

// loadLauncherBrowserArgs 从配置文件加载浏览器启动参数。
// [S-sig 汇编 0x1409ddb40] (configPath string) []string：loadLauncherConfigIfExists → 失败或 !ok
// 返回 nil（test rbx / test al），否则把 cfg 相关 0x448 字节转交 buildLauncherBrowserArgs。体待还原。
func loadLauncherBrowserArgs(configPath string) []string { return nil }
