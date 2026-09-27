// AUTO-RECONSTRUCTED — DOMAIN: config drag appid filter
// 研究用途. [S 汇编实证 0x140881f60, 800B]：遍历 entries，normalizeAppEntryType=="app" 且 TrimSpace 非空
// 保留（按 apps 过滤 dragLaunchAppIDs）。
package main

import "strings"

// filterDragLaunchAppIDsByAppEntries 按 app 条目过滤拖拽启动 appid。
// [S 汇编 0x140881f60, 800B(0x320)]：normalizeAppEntryType=="app" 建集合，按集合+TrimSpace 非空过滤。
func filterDragLaunchAppIDsByAppEntries(dragIDs []string, apps []AppEntry) []string {
	app := make(map[string]bool)
	for _, a := range apps {
		if normalizeAppEntryType(a.EntryType) == "app" {
			app[a.ID] = true
		}
	}
	out := make([]string, 0, len(dragIDs))
	for _, id := range dragIDs {
		if app[id] && strings.TrimSpace(id) != "" {
			out = append(out, id)
		}
	}
	return out
}
