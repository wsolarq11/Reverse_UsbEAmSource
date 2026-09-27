// AUTO-RECONSTRUCTED FUNCTIONS — DOMAIN: launcher update notification (toast)
// 研究用途
//
// 契约来源：
//   - 符号地址：symbols.main.bak
//   - 行号蓝图：source_funcs.txt launcherupdatenotification.go L12-41
//
// 档位：[S] 反汇编实证
package main

import "strings"

// normalizeLauncherUpdateNotificationTitle 规范化更新通知标题。
// [S 汇编 0x1408cfa60]
func normalizeLauncherUpdateNotificationTitle(title string) string {
	return strings.TrimSpace(title)
}

// normalizeLauncherUpdateNotificationMessage 规范化更新通知消息。
// [S 汇编 0x1408cfac0]
func normalizeLauncherUpdateNotificationMessage(msg string) string {
	return strings.TrimSpace(msg)
}

// truncateUTF16String 截断 UTF-16 字符串到最大长度（字节）。
// [S 汇编 0x1408cfb20]
func truncateUTF16String(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	return s[:maxBytes]
}
