// AUTO-RECONSTRUCTED — DOMAIN: plugin update 版本比较
// 研究用途。反汇编实证：
//
//	isPluginVersionUpdateAvailable 0x14092c600, 192B
//	comparePluginVersionText 0x14092c6c0, 263B
//
// 档位：[S] 体逐条对位；[S-sig] 签名实证、体骨架。
package main

import "strings"

// isPluginVersionUpdateAvailable 判断插件是否有版本更新可用。
// [S 汇编 0x14092c600, 192B]：current/latest 各 TrimSpace；latest 空 → false；
// current 空 → true；否则 comparePluginVersionText(current,latest) > 0。
func isPluginVersionUpdateAvailable(current, latest string) bool {
	current = strings.TrimSpace(current)
	latest = strings.TrimSpace(latest)
	if latest == "" {
		return false
	}
	if current == "" {
		return true
	}
	return comparePluginVersionText(current, latest) > 0
}

// comparePluginVersionText 比较两段插件版本文本，返回 >0 / 0 / <0。
// [S-sig 0x14092c6c0, 263B]：签名实证——(current, latest string) int；体待版本分片解析专项还原。
func comparePluginVersionText(current, latest string) int {
	_, _ = current, latest
	return 0
}
