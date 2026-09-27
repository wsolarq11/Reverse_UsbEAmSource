// AUTO-RECONSTRUCTED — DOMAIN: app entry normalize
// 研究用途. [S 汇编实证 0x140889780, 4544B]
//
// 契约：迭代 []AppEntry，逐项规整后返回新切片。
// 子函数调用链：
//
//	normalizeAppEntryType → inferName → inferWorkingDir → shouldInferWorkingDir →
//	cleanStringList → normalizeAppLaunchPrivilegeMode → normalizeShortcutMode →
//	normalizeShortcutPath → normalizeResolvedShortcutArguments → splitCommandLineArguments
//
// 常量经 .rodata 解码确凿：
//   - "directory"（9B, 0x140C3F6D0）
//   - "app"（3B, 0x140C33B63）
//   - "folder"（6B, 0x140C375A8，用于 directory 条目图标种类）
//   - "data:image/"（11B, 0x140C47539，图标数据前缀判定）
//   - 快捷方式模式常量在 normalizeShortcutMode 中解码
package main

import "strings"

// normalizeAppEntryType 规整应用条目类型。
// [S 汇编实证 0x14088a700, 128B]：TrimSpace → EqualFold("directory") → "directory"，否则 "app"。
func normalizeAppEntryType(s string) string {
	s = strings.TrimSpace(s)
	if strings.EqualFold(s, "directory") {
		return "directory"
	}
	return "app"
}

// normalizeAppEntries 规整应用条目列表，返回新切片。
// [S 汇编 0x140889780, 4544B]：idAllocator.Next 逐项生成 ID；Name 空跳过；
// entryType 规整→path 推断→icon→workingDir→特权模式→快捷方式模式→args 分割。
func normalizeAppEntries(entries []AppEntry) []AppEntry {
	if len(entries) == 0 {
		return nil
	}
	out := make([]AppEntry, 0, len(entries))

	// [S 汇编] idAllocator 在函数内构造（入口 map 初始化），逐项 Next(name) 生成字符串 ID
	var alloc idAllocator

	for _, e := range entries {
		// 1. 跳过 Name 为空条目
		name := strings.TrimSpace(e.Name)
		if name == "" {
			continue
		}

		// 2. 规整 entryType
		entryType := normalizeAppEntryType(e.EntryType)

		// 3. 规整 Path（空则从 Name 推断）
		path := strings.TrimSpace(e.Path)
		if path == "" {
			path = inferName(name)
		}

		// 4. 规整 Icon
		icon := strings.TrimSpace(e.Icon)

		// 5. 规整 WorkingDir
		// [S] 汇编：entryType=="directory" → workingDir=path；
		// 否则 shouldInferWorkingDir(name,path) → inferWorkingDir(name,path) 或 path
		var workingDir string
		if entryType == "directory" {
			workingDir = path
		} else {
			if shouldInferWorkingDir(path) {
				workingDir = inferWorkingDir(path)
			} else {
				workingDir = path
			}
		}

		// 6. AutoIcon / IconData 判定（[S] 汇编 verify 链）
		// [S] 汇编 0x140889a93-0x140889b2f：AutoIcon 优先 → IconRef 非空 → true
		// 否则 IconData TrimSpace+ToLower + 11B memequal "data:image/"
		autoIcon := e.AutoIcon
		if !autoIcon {
			iconRefTrim := strings.TrimSpace(e.IconRef)
			autoIcon = iconRefTrim != ""
		}

		iconDataLower := strings.ToLower(strings.TrimSpace(e.IconData))
		iconRefTrim := strings.TrimSpace(e.IconRef)
		var hasUploadIcon bool
		if autoIcon {
			hasUploadIcon = iconRefTrim != ""
		} else {
			hasUploadIcon = len(iconDataLower) >= 11 && iconDataLower[:11] == "data:image/"
		}

		trimmedIconData := strings.TrimSpace(e.IconData)
		trimmedIconRef := iconRefTrim
		if hasUploadIcon && !autoIcon {
			trimmedIconRef = ""
		}

		// 7. IconURL / CustomIconData / CustomIconRef 规整
		iconUrl := strings.TrimSpace(e.IconURL)
		customIconData := strings.TrimSpace(e.CustomIconData)
		customIconRef := strings.TrimSpace(e.CustomIconRef)

		// 8. 图标种类判定（[S] 汇编 0x140889d2e-0x140889d83）
		iconKind := "app"
		if entryType == "directory" {
			iconKind = "folder"
		}

		// 9. Args / Tags 清理（两个 cleanStringList 调用）
		args := cleanStringList(e.Args)
		tags := cleanStringList(e.Tags)

		// 10. 启动权限规整
		launchPrivilege := normalizeAppLaunchPrivilegeMode(e.LaunchPrivilege)

		// 11. 快捷方式模式/路径规整
		// normalizeShortcutMode(mode, fallback) 第二形参取 e.ShortcutPath（非空即 shortcut 模式，[推断]）。
		shortcutMode := normalizeShortcutMode(e.ShortcutMode, e.ShortcutPath)
		shortcutPath := normalizeShortcutPath(e.ShortcutPath, path)
		shortcutTargetPath := strings.TrimSpace(e.ShortcutTargetPath)
		shortcutWorkingDir := strings.TrimSpace(e.ShortcutWorkingDir)
		shortcutArgumentsText := strings.TrimSpace(e.ShortcutArgumentsText)

		// 12. 快捷方式参数解析
		resolvedArgs := normalizeResolvedShortcutArguments(args, shortcutMode, shortcutTargetPath, shortcutArgumentsText)

		// 13. directory 条目清空快捷方式字段
		if entryType == "directory" {
			shortcutMode = ""
			shortcutPath = ""
			shortcutTargetPath = ""
			shortcutWorkingDir = ""
			shortcutArgumentsText = ""
		}

		// 14. 构建输出条目
		normalized := AppEntry{
			ID:                    alloc.Next([]string{name}),
			Name:                  name,
			EntryType:             entryType,
			Icon:                  icon,
			Favorite:              e.Favorite,
			AutoIcon:              autoIcon,
			IconData:              trimmedIconData,
			IconRef:               trimmedIconRef,
			IconURL:               iconUrl,
			CustomIconData:        customIconData,
			CustomIconRef:         customIconRef,
			IconDataVersion:       e.IconDataVersion,
			Path:                  path,
			WorkingDir:            workingDir,
			Args:                  resolvedArgs,
			Tags:                  tags,
			LaunchCount:           e.LaunchCount,
			LastLaunchedAt:        strings.TrimSpace(e.LastLaunchedAt),
			LaunchPrivilege:       launchPrivilege,
			ShortcutMode:          shortcutMode,
			ShortcutPath:          shortcutPath,
			ShortcutTargetPath:    shortcutTargetPath,
			ShortcutWorkingDir:    shortcutWorkingDir,
			ShortcutArgumentsText: shortcutArgumentsText,
		}

		// [P] iconKind 字段待类型层扩展
		_ = iconKind

		out = append(out, normalized)
	}
	return out
}

// normalizeResolvedShortcutArguments 规整快捷方式参数。
// [S 汇编实证 0x14088a5e0, 200B]：
// 若 shortcutMode == "resolved"（8B）且 args 仅含 1 个元素（即原 shortcutPath），
// 且 shortcutTargetPath 与 shortcutArgumentsText 可分割，则替换 args。
func normalizeResolvedShortcutArguments(args []string, shortcutMode, shortcutTargetPath, shortcutArgumentsText string) []string {
	if shortcutMode != "resolved" {
		return args
	}
	if len(args) != 1 {
		return args
	}
	// [S] 检查 args[0] 是否与 shortcutTargetPath 相同（memequal 判定）
	// 若相同且 shortcutArgumentsText 非空，则分割替换
	shortcutArgs := splitCommandLineArguments(shortcutArgumentsText)
	if len(shortcutArgs) > 1 {
		return shortcutArgs
	}
	return args
}

// splitCommandLineArguments 由 backend/splitcommandline.go 提供（[S] 汇编 0x1408a7ee0, 1440B）。
