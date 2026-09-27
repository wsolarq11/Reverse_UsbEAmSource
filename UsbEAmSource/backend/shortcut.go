// AUTO-RECONSTRUCTED — DOMAIN: shortcut path detection & icon resource utilities
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)
//
// 函数来源：
//
//	isShellLinkShortcutPath           0x1409c1d80  (96B)
//	isInternetShortcutPath            0x1409c1e20  (96B)
//	isShortcutFilePath                0x1409c1ec0  (96B)
//	normalizeShortcutIconResourcePath 0x1409c6160  (288B)
//	isShortcutCustomIcon              0x1409c6080  (224B)
//	parseShortcutIconIndex            0x1409c6500  (224B)
//	expandShortcutIconWindowsEnvironment 0x1409c63e0  (288B)
//	parseShortcutIconLocation         0x1409c61c0  (320B)
package main

import (
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

// ---- 快捷方式路径检测 ----

// isShellLinkShortcutPath 检测路径是否为 .lnk 快捷方式。
// [S 0x1409c1d80]
func isShellLinkShortcutPath(path string) bool {
	p := strings.TrimSpace(path)
	for i := len(p) - 1; i >= 0; i-- {
		c := p[i]
		if c == '\\' || c == '/' {
			break
		}
		if c == '.' {
			return strings.EqualFold(p[i:], ".lnk")
		}
	}
	return false
}

// isInternetShortcutPath 检测路径是否为 .url 快捷方式。
// [S 0x1409c1e20]
func isInternetShortcutPath(path string) bool {
	p := strings.TrimSpace(path)
	for i := len(p) - 1; i >= 0; i-- {
		c := p[i]
		if c == '\\' || c == '/' {
			break
		}
		if c == '.' {
			return strings.EqualFold(p[i:], ".url")
		}
	}
	return false
}

// isShortcutFilePath 检测路径是否为快捷方式（.lnk 或 .url）。
// [S 0x1409c1ec0]
func isShortcutFilePath(path string) bool {
	return isShellLinkShortcutPath(path) || isInternetShortcutPath(path)
}

// ---- 图标资源路径归一化 ----

// normalizeShortcutIconResourcePath 归一化快捷方式图标资源路径。
// [S 0x1409c6160]
func normalizeShortcutIconResourcePath(path string) string {
	p := strings.TrimSpace(path)
	p = strings.Trim(p, `"`)
	if p == "" {
		return ""
	}
	return filepath.Clean(p)
}

// isShortcutCustomIcon 检查快捷方式是否设了自定义图标。
// [S 0x1409c6080]
func isShortcutCustomIcon(path string, iconLocation string) bool {
	norm := normalizeShortcutIconResourcePath(path)
	if norm == "" {
		return false
	}
	if iconLocation == "" {
		return true
	}
	return strings.EqualFold(normalizeShortcutIconResourcePath(path),
		normalizeShortcutIconResourcePath(iconLocation))
}

// ---- 图标索引解析 ----

// parseShortcutIconIndex 解析图标索引字符串（跳过前导逗号）。
// [S 0x1409c6500]
func parseShortcutIconIndex(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	if s[0] == ',' {
		s = strings.TrimSpace(s[1:])
	} else {
		return 0, false
	}
	if s == "" {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return n, true
}

// ---- 环境变量展开 ----

// expandShortcutIconWindowsEnvironment 展开图标路径中的 Windows 环境变量 (%...%)。
// [S 0x1409c63e0]
func expandShortcutIconWindowsEnvironment(path string) (string, bool) {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return "", false
	}
	n, err := windows.ExpandEnvironmentStrings(ptr, nil, 0)
	if err != nil || n == 0 || n > 0x8000 {
		return "", false
	}
	buf := make([]uint16, n)
	_, err = windows.ExpandEnvironmentStrings(ptr, &buf[0], n)
	if err != nil {
		return "", false
	}
	return syscall.UTF16ToString(buf), true
}

// ---- 图标位置串解析 ----

// parseShortcutIconLocation 解析 "file,index" 格式的图标位置串。
// [S 0x1409c61c0]
func parseShortcutIconLocation(location string) (path string, resolvedPath string, index int, ok bool) {
	loc := strings.TrimSpace(location)
	if loc == "" {
		return "", "", 0, false
	}

	if loc[0] == '"' {
		// 引号包裹：在引号内找逗号
		idx := strings.IndexByte(loc[1:], '"')
		if idx < 0 {
			// 没有闭合引号，整体视为路径
			path = loc[1:]
			return path, "", 0, false
		}
		path = loc[1 : 1+idx]
		rest := loc[1+idx+1:]
		// 逗号可能在引号外
		if strings.HasPrefix(rest, ",") {
			idx2, ok2 := parseShortcutIconIndex(rest[1:])
			if ok2 {
				return path, "", idx2, true
			}
		}
		return path, "", 0, true
	}

	// 无引号：找最后一个逗号
	if commaIdx := strings.LastIndex(loc, ","); commaIdx >= 0 {
		path = loc[:commaIdx]
		idx, ok := parseShortcutIconIndex(loc[commaIdx+1:])
		if ok {
			return path, "", idx, true
		}
	}
	// 无逗号或解析失败
	return loc, "", 0, false
}
