// AUTO-RECONSTRUCTED — normalizeShortcutPath ([S] 0x14088a8a0)
package main

import "strings"

// normalizeShortcutPath 规整快捷方式路径。
// [S 汇编 0x14088a8a0, 96B(0x60)]：TrimSpace(path) 非空则返回；否则 fallback 为快捷方式路径时返回其 Trim 值；否则空。
func normalizeShortcutPath(path string, fallback string) string {
	p := strings.TrimSpace(path)
	if p != "" {
		return p
	}
	if isShortcutFilePath(fallback) {
		return strings.TrimSpace(fallback)
	}
	return ""
}
