// AUTO-RECONSTRUCTED — DOMAIN: app display name inference
// 研究用途. [S 汇编 0x14088ca40, 480B]：从路径取 basename（去目录与扩展名），空/无名称回退空。
package main

import "strings"

// inferName 从路径取 basename（去目录与扩展名）。
// [S 汇编 0x14088ca40, 480B(0x1e0)]：TrimSpace→LastIndexAny(`\/`)→LastIndexByte('.')→空/非法回退空。
// [P] 多扩展名/.lnk 特殊策略待精确。
func inferName(path string) string {
	p := strings.TrimSpace(path)
	base := p
	if i := strings.LastIndexAny(base, `\/`); i >= 0 {
		base = base[i+1:]
	}
	if i := strings.LastIndexByte(base, '.'); i > 0 {
		base = base[:i] // [P] 多扩展名/.lnk 特殊策略待精确
	}
	if base == "" || base == "." || base == `\` || base == "/" {
		return ""
	}
	return base
}
