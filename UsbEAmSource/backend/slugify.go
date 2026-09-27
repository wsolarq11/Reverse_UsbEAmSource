// AUTO-RECONSTRUCTED — DOMAIN: slug generator for id allocator
// 研究用途. 实证依据见 asset_semantics 11.66 + 本轮 dump 刻度（576B，0x14088d0c0，source_funcs 47 行）。
// 汇编确凿顺序：TrimSpace → ToLower → Replacer.Replace(28 项, 0x14088d1c0) → 逐符文筛选
// (0-9/a-z 保留，其余 '-' 压缩, 0x14088d1f2-0x254) → Trim('-', 单字节 cutset, 0x14088d2b2)。
// Replacer 28 项 .rodata 解码（0x1411e6da0）：
//
//	\ / : . , _ -> '-'；()[]{}'   " -> 删除（空替换）
package main

import "strings"

// slugifyReplacer 将常见分隔符映射为 '-'、将括号/引号删除（汇编 28 项实证）。
var slugifyReplacer = strings.NewReplacer(
	`\`, "-",
	`/`, "-",
	`:`, "-",
	`.`, "-",
	`,`, "-",
	`_`, "-",
	`(`, "",
	`)`, "",
	`[`, "",
	`]`, "",
	`{`, "",
	`}`, "",
	`'`, "",
	`"`, "",
)

// slugify 将输入规整为 ASCII-only slug 标识片段。
// 仅保留数字与小写字母，其余符文压缩为单个 '-'，首尾 '-' 剔除。
// [S 汇编 0x14088d0c0, 576B(0x240)]：TrimSpace→ToLower→Replacer(28项)→逐符文筛选→Trim('-')。
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugifyReplacer.Replace(s)
	var b strings.Builder
	b.Grow(len(s))
	prevDash := false
	for _, r := range s {
		keep := (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z')
		if keep {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
