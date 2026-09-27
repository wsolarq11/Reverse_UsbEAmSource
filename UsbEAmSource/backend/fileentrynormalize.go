// AUTO-RECONSTRUCTED — DOMAIN: file entry normalize
// 研究用途. 实证依据 dump 反汇编（1088B，0x14088b880）。
// 汇编确定性语义（逐条对齐）：
//   - Path 空 → 跳过条目（0x14088ba05 test + 恢复循环头）
//   - Name 空 → 从 Path 经 inferName 推导（0x14088ba4c）
//   - ID = idAllocator.Next，3 候选 [原ID, Name, Path]（0x14088baf8 构造，ecx=3）
//   - Tags 经 cleanStringList 清理（0x14088bb22）
package main

import "strings"

// normalizeFileEntries 规整文件条目列表，返回新切片。
// [S 汇编 0x14088b880, 1088B(0x440)]：Path 空跳过；Name 空 inferName 推导；ID=idAllocator.Next；Tags cleanStringList。
func normalizeFileEntries(entries []FileEntry) []FileEntry {
	if len(entries) == 0 {
		return nil
	}
	out := make([]FileEntry, 0, len(entries))

	// [S 汇编] idAllocator 在函数内构造（入口 map 初始化 + runtime.rand）
	var alloc idAllocator

	for _, e := range entries {
		// [S] Path 空则跳过条目
		path := strings.TrimSpace(e.Path)
		if path == "" {
			continue
		}

		// [S] Name 空则从 Path 推导
		name := strings.TrimSpace(e.Name)
		if name == "" {
			name = inferName(path)
		}

		// [S] ID = idAllocator.Next，候选 [原ID, Name, Path]，取最后非空 → 以 Path 为基
		id := alloc.Next([]string{e.ID, name, path})

		// [S] Tags 清理
		tags := cleanStringList(e.Tags)

		out = append(out, FileEntry{
			ID:   id,
			Name: name,
			Path: path,
			Tags: tags,
		})
	}
	return out
}
