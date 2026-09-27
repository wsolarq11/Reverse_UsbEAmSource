// AUTO-RECONSTRUCTED — shortcut path 辅助（internal，由 shortcut.go 提供基础判定）
// [S] 路径扩展名提取辅助函数
package main

import "strings"

// shortcutPathExt 剥离路径返回小写扩展名（内部辅助）。
// [S 汇编实证，内联]：TrimSpace → LastIndexAny("/\\") → LastIndexByte('.') → ToLower。
func shortcutPathExt(p string) string {
	p = strings.TrimSpace(p)
	if i := strings.LastIndexAny(p, "/\\"); i >= 0 {
		p = p[i+1:]
	}
	if i := strings.LastIndexByte(p, '.'); i >= 0 {
		return strings.ToLower(p[i+1:])
	}
	return ""
}
