// AUTO-RECONSTRUCTED — DOMAIN: bookmark source normalize
// 研究用途. 实证依据 dump 反汇编（1568B，0x14088b260）。
// 摘要确定性语义（逐条对齐汇编）：
//   - 空条目跳过（Name/Browser/Path 均空，0x14088b458 continue）
//   - Name/Browser/Path 逐项 TrimSpace（0x14088b3e0/0x402/0x424）
//   - ID 由 idAllocator.Next 生成，候选为 [原ID, Name, Browser, Path] 四值（0x14088b68c-0x8a4）
//   - Enabled 规整：原值非 true 且 ID 空 → 置 true 并进入 Name 推导（0x14088b582-0x5f5）
//   - Name 空时经 describeBookmarkSourceDescriptor 后由 inferBookmarkSourceName 推导，
//     [P] describe 家族（resolveBookmarkSourceKind/describeFirefox|ChromiumBookmarkPath
//     /resolveBookmarkBrowserKind）深层未完整还原，本处以 infer 通用 path 回退承载。
package main

import "strings"

// normalizeBookmarkSources 规整书签来源列表，返回新切片。
// [S 汇编 0x14088b260, 1568B(0x620)]：空条目跳过；Name/Browser/Path TrimSpace；ID=idAllocator.Next；
// Enabled 规整；[P] describe 家族深层未还原（此处 infer 通用 path 回退承载）。
func normalizeBookmarkSources(entries []BookmarkSource) []BookmarkSource {
	if len(entries) == 0 {
		return nil
	}
	out := make([]BookmarkSource, 0, len(entries))

	// [S 汇编] idAllocator 在函数内构造（入口 map 初始化 + runtime.rand）
	var alloc idAllocator

	for _, e := range entries {
		// 逐项 TrimSpace（汇编三次：Name/Browser/Path）
		name := strings.TrimSpace(e.Name)
		browser := strings.TrimSpace(e.Browser)
		path := strings.TrimSpace(e.Path)

		// [S] 全空条目跳过
		if name == "" && browser == "" && path == "" {
			continue
		}

		// [S] 未启用且原 ID 为空 → 置启用，并进入 Name 推导
		enabled := e.Enabled
		if !enabled && strings.TrimSpace(e.ID) == "" {
			enabled = true
			// [S] Name 空时经描述/推断产出（[P] describe 深层未实现，用通用 path 回退）
			if name == "" {
				name = inferBookmarkSourceName(browser, "", path)
			}
		}

		// [S] ID = idAllocator.Next，候选 [原ID, Name, Browser, Path] 依序（0x14088b688-0x8a4 构造），
		// 取最后一个非空 → 通常以 Path 为 ID 基数
		id := alloc.Next([]string{e.ID, name, browser, path})

		out = append(out, BookmarkSource{
			ID:      id,
			Name:    name,
			Browser: browser,
			Path:    path,
			Enabled: enabled,
		})
	}
	return out
}
