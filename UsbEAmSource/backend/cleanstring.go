// AUTO-RECONSTRUCTED — [S 汇编 0x14088c160, 704B]：TrimSpace 每项，空跳过，ToLower 键去重（保留 Trim 值）。
package main

import "strings"

// cleanStringList 规整字符串列表。
// [S 汇编 0x14088c160, 704B(0x2c0)]：TrimSpace 每项，空跳过，ToLower 键去重（保留 Trim 值）。
func cleanStringList(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, it := range items {
		t := strings.TrimSpace(it)
		if t == "" {
			continue
		}
		key := strings.ToLower(t)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, t)
	}
	return out
}
