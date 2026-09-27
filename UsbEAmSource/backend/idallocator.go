// AUTO-RECONSTRUCTED — DOMAIN: id allocator
// 研究用途. 实证依据见 asset_semantics 11.66 + 本轮 dump 复核（480B，0x14088cee0）：
//
//	非简单 uint64 自增；内部含 map[string]int64 计数。
//
// 汇编确定性语义：
//   - 入参为候选字符串切片（normalizeAppEntries 调用点 ecx/rdi=3 三元素，0x14088a0d0-0x14088a0e8）
//   - 遍历候选逐个 slugify，取最后一个非空结果（0x14088cf0f 循环 + test/je）
//   - 首次出现：map 置 1，返回裸 slug；已存在：递增并返回 fmt.Sprintf("%s-%d", slug, count)
//   - 格式串 .rodata 0x140c35cf9 解码 "%s-%d"；空 slug 兜底 4 字符 key（0x140c3491a）
package main

import "fmt"

type idAllocator struct {
	counts map[string]int64
}

// Next 从候选名中取最后一个非空 slug 作为该 ID 的基，生成唯一 ID。
// 首次返回裸 slug；同名后续返回 "<slug>-<seq>"（seq 自 2 起）。
// [S 汇编 0x14088cee0, 480B(0x1e0)]：遍历候选 slugify 取末非空；map 计数；"%s-%d"；空 slug 兜底 4 字符。
func (a *idAllocator) Next(names []string) string {
	if a.counts == nil {
		a.counts = make(map[string]int64)
	}
	s := ""
	for _, n := range names {
		if k := slugify(n); k != "" {
			s = k
		}
	}
	if s == "" {
		s = "item" // [S] 汇编空 slug 兜底（.rodata 4 字符，0x140c3491a）
	}
	c, ok := a.counts[s]
	if !ok {
		a.counts[s] = 1
		return s
	}
	a.counts[s] = c + 1
	return fmt.Sprintf("%s-%d", s, c+1)
}
