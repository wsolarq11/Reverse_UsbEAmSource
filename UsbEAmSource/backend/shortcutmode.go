// AUTO-RECONSTRUCTED — DOMAIN: shortcut mode canonical
// 研究用途. [S 汇编 0x14088a780, 288B]：TrimSpaceToLower 后归一到 "resolved"/"shortcut"；否则回退原小写
// （原汇编另含 isShortcutFilePath 回退分支 [P]）。
package main

import "strings"

// normalizeShortcutMode 归一快捷方式模式。
// [S 汇编 0x14088a780, 288B(0x120)]：ToLower(TrimSpace(mode)) 为 resolved/shortcut 则返回；
// 否则 fallback 非空 → "shortcut"；mode 为快捷方式路径 → "shortcut"；否则 ""。
func normalizeShortcutMode(mode string, fallback string) string {
	m := strings.ToLower(strings.TrimSpace(mode))
	if m == "resolved" {
		return "resolved"
	}
	if m == "shortcut" {
		return "shortcut"
	}
	if strings.TrimSpace(fallback) != "" {
		return "shortcut"
	}
	if isShortcutFilePath(mode) {
		return "shortcut"
	}
	return ""
}
