// AUTO-RECONSTRUCTED — DOMAIN: internet shortcut icon location
// 研究用途. 实证依据 dump 反汇编（416B，0x14086a9c0）。
// 汇编确定性语义：
//   - TrimSpace(iconFile) 空 → 空（0x14086a9f4/0x14086a9f7）
//   - TrimSpace(iconIndex) → strconv.Atoi，解析失败 idx=0（0x14086aa20）
//   - 含逗号分支：strings.Replace(iconFile, "\"", "", -1) 去引号 + concatstring3 拼接（0x14086aa78-0xb6）
//   - 结果 = fmt.Sprintf("%s,%d", icon, index)，格式串 .rodata 0x14086aaf8 解码 "%s,%d" 确凿
//     [P] 含逗号分支的 concat 中段拼装细节（引号成对）未逐字 trace，主路径覆盖 "%s,%d" 语义。
package main

import (
	"fmt"
	"strconv"
	"strings"
)

// buildInternetShortcutIconLocation 把图标文件与索引组装为 "file,index" 位置串。
// [S 汇编 0x14086a9c0, 416B]：TrimSpace 空→""；strconv.Atoi 失败 idx=0；
// 含逗号去引号；fmt.Sprintf("%s,%d")。 [P] 含逗号 concat 中段引号成对细节待精。
func buildInternetShortcutIconLocation(iconFile, iconIndex string) string {
	f := strings.TrimSpace(iconFile)
	if f == "" {
		return ""
	}
	idx := 0
	if t := strings.TrimSpace(iconIndex); t != "" {
		if n, err := strconv.Atoi(t); err == nil {
			idx = n
		}
	}
	// [S] 含逗号时去引号（含逗号做实序）：避免格式歧义，统一主路径 "%s,%d"
	if strings.Contains(f, ",") {
		f = strings.ReplaceAll(f, `"`, "")
	}
	return fmt.Sprintf("%s,%d", f, idx)
}
