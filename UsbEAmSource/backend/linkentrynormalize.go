// AUTO-RECONSTRUCTED — DOMAIN: link entry normalize
// 研究用途. [S 汇编实证 0x14088a940, 2016B]
//
// 契约：迭代 []LinkEntry，逐项规整后返回新切片。
// 依赖 normalizeLinkIconMode（判定 upload/favicon）→ cleanStringList（清理 Args/Tags）。
//
// 调用上下文（normalizeLauncherConfigWithOptions）：
//
//	两次调用，分别传入不同 defaultIcon 字面量（5B 与 8B）。
//
// 常量经 .rodata 解码确凿：
//   - "upload"（6B, 0x140C376B6）
//   - "favicon"（7B, 0x140C39449）
//   - "data:image/"（11B, 0x140C47539）
package main

import "strings"

// normalizeLinkEntries 规整链接条目列表，返回新切片。
// [S 汇编 0x14088a940, 2016B(0x7e0)]：逐项规整，依赖 normalizeLinkIconMode + cleanStringList。
func normalizeLinkEntries(entries []LinkEntry, defaultIcon string) []LinkEntry {
	if len(entries) == 0 {
		return nil
	}
	out := make([]LinkEntry, 0, len(entries))

	// [S 汇编] idAllocator 在函数内构造，逐项 Next(title) 生成字符串 ID
	var alloc idAllocator

	for _, e := range entries {
		// 1. 规整 URL，空则跳过条目
		url := strings.TrimSpace(e.URL)
		if url == "" {
			continue
		}

		// 2. 规整 Name
		name := strings.TrimSpace(e.Name)

		// 3. 规整图标模式（upload / favicon）
		iconMode := normalizeLinkIconMode(e.IconMode, e.IconData, e.IconRef)

		// 4. 规整 IconData / IconRef：仅 upload 模式保留
		iconData := strings.TrimSpace(e.IconData)
		iconRef := strings.TrimSpace(e.IconRef)
		if iconMode != "upload" {
			iconData = ""
			iconRef = ""
		}

		// 5. 规整 Icon（空则用默认图标）
		icon := strings.TrimSpace(e.Icon)
		if icon == "" {
			icon = defaultIcon
		}

		// 6. 规整 IconURL
		iconURL := strings.TrimSpace(e.IconURL)

		// 7. 规整 Browser
		browser := strings.TrimSpace(e.Browser)

		// 8. 清理 Args / Tags
		args := cleanStringList(e.Args)
		tags := cleanStringList(e.Tags)

		// 9. 规整 LastLaunchedAt
		lastLaunchedAt := strings.TrimSpace(e.LastLaunchedAt)

		// 10. 修正 LaunchCount（负值置零）
		launchCount := e.LaunchCount
		if launchCount < 0 {
			launchCount = 0
		}

		// 11. 标题回退：Name 为空时使用 URL
		title := name
		if title == "" {
			title = url
		}

		// 12. 构建输出条目
		normalized := LinkEntry{
			ID:             alloc.Next([]string{title}),
			Name:           title,
			TitleLocked:    e.TitleLocked,
			Icon:           icon,
			Favorite:       e.Favorite,
			IconMode:       iconMode,
			IconData:       iconData,
			IconRef:        iconRef,
			IconURL:        iconURL,
			URL:            url,
			Browser:        browser,
			Args:           args,
			Tags:           tags,
			LaunchCount:    launchCount,
			LastLaunchedAt: lastLaunchedAt,
		}

		out = append(out, normalized)
	}
	return out
}
