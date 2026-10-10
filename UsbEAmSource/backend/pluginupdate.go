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

// findRemoteCatalogEntry 按 ID（TrimSpace + EqualFold 不区分大小写）在远程目录切片中查找条目。
// [S 汇编 0x14092b900, 544B]：遍历 entries（步长 0x138），EqualFold(TrimSpace(e.ID), TrimSpace(name))
// 命中返回 (e, true)，否则 (zero, false)。
func findRemoteCatalogEntry(name string, entries []PluginRemoteCatalogEntry) (PluginRemoteCatalogEntry, bool) {
	for _, e := range entries {
		if strings.EqualFold(strings.TrimSpace(e.ID), strings.TrimSpace(name)) {
			return e, true
		}
	}
	return PluginRemoteCatalogEntry{}, false
}

// findPluginInUpdateState 按 ID（TrimSpace + EqualFold）在更新状态插件清单中查找。
// [S 汇编 0x14092b6e0, 544B]：遍历 manifests（步长 0x178），EqualFold(TrimSpace(e.ID), TrimSpace(name))
// 命中返回 (e, true)，否则 (zero, false)。
func findPluginInUpdateState(name string, entries []PluginManifest) (PluginManifest, bool) {
	for _, e := range entries {
		if strings.EqualFold(strings.TrimSpace(e.ID), strings.TrimSpace(name)) {
			return e, true
		}
	}
	return PluginManifest{}, false
}
