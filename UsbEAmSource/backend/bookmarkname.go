// AUTO-RECONSTRUCTED — DOMAIN: bookmark source name inference
// 研究用途. 实证依据 dump 反汇编（448B，0x140766dc0，source_funcs 行号）。
// 汇编确定性语义（0x140766df7 TrimSpace×2 / 0x140766e99 fmt.Sprintf）：
//   - TrimSpace(A)、TrimSpace(B)；A 非空且 B 空 → 返回 A
//   - A 非空且 B 非空 → 返回 fmt.Sprintf("%s / %s", A, B)（格式串 .rodata 0x140c392f9 解码确凿）
//   - A 空且 B 非空 → 返回 B
//   - A 空且 B 空 → 对 C 取 Dir→Base；结果为空或为 "." / "\" 则回退 inferName(C)
package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// inferBookmarkSourceName 依据两个候选名与来源路径推导显示的来源名称。
// [S 汇编 0x140766dc0, 448B(0x1c0)]：TrimSpace 两候选；A非空B空→A；A/B非空→fmt.Sprintf("%s / %s")；
// A空B非空→B；皆空→Dir→Base，非法回退 inferName。
func inferBookmarkSourceName(a, b, c string) string {
	ta := strings.TrimSpace(a)
	if ta != "" {
		tb := strings.TrimSpace(b)
		if tb == "" {
			return ta
		}
		return fmt.Sprintf("%s / %s", ta, tb)
	}
	if strings.TrimSpace(b) != "" {
		return strings.TrimSpace(b)
	}
	// 两候选皆空：从路径目录取基名；拼接空/非法则回退 inferName
	p := strings.TrimSpace(c)
	base := filepath.Base(filepath.Dir(p))
	if base == "" || base == "." || base == `\` {
		return inferName(p)
	}
	return base
}
