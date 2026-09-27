// AUTO-RECONSTRUCTED -- DOMAIN: start menu icon resolution (Windows)
// 研究用途
//
// 契约来源：
//   - 符号地址：symbols.main.bak（0x1409c56c0/0x1409c5900/0x1409c5ac0/0x1409c5de0）
//   - 行号蓝图：source_funcs.txt File: startmenu_icon_windows.go L16-208
//   - 反汇编：tmp_disasm_shortcut/resolve{StartMenuIconData,DynamicStartMenuIconData,
//     ShortcutDisplayIconData,DynamicShortcutDisplayIconData}.asm.txt
//
// 依赖链：
//   - shortcut.go: isShortcutCustomIcon, parseShortcutIconLocation
//   - appicon_windows.go: resolveAppIconDataWithOptions,
//     resolveDynamicStartMenuAppIconData, resolveDynamicStartMenuAppIconDataWithIndex
//   - automaticpath_windows.go: classifyAutomaticWindowsPath
//
// 档位：[S] 反汇编逐条追译。
package main

import "strings"

// ========================================================================
// 开始菜单图标解析（静态扫描路口）
// ========================================================================

// resolveStartMenuIconData 解析开始菜单项的图标数据。
// [S 汇编实证 0x1409c56c0, 0x227B]：
//
//	第一步：调 resolveShortcutDisplayIconData(targetPath, iconLocation)；
//	若返回非空 -> 直接使用。
//	失败后：对 targetPath / path / iconLocation 三个候选路径，
//	逐个 classifyAutomaticWindowsPath=="local" -> resolveAppIconDataWithOptions，
//	取首个成功结果。
func resolveStartMenuIconData(path, targetPath, iconLocation string) (string, error) {
	iconData, customIconData := resolveShortcutDisplayIconData(targetPath, iconLocation)
	if iconData != "" {
		if customIconData != "" {
			return customIconData, nil
		}
		return iconData, nil
	}

	// [S] 三个候选的下标指向：先 targetPath, 再 path, 再 iconLocation
	candidates := make([]string, 0, 3)
	if tp := strings.TrimSpace(targetPath); tp != "" {
		candidates = append(candidates, tp)
	}
	if sp := strings.TrimSpace(path); sp != "" {
		candidates = append(candidates, sp)
	}
	if il := strings.TrimSpace(iconLocation); il != "" {
		candidates = append(candidates, il)
	}

	for _, c := range candidates {
		class, _ := classifyAutomaticWindowsPath(c, nil)
		if class != classLocal {
			continue
		}
		if data := resolveAppIconDataWithOptions(c, AppIconOptions{}); data != "" {
			return data, nil
		}
	}
	return "", nil
}

// resolveDynamicStartMenuIconData 动态解析开始菜单项图标。
// [S 汇编实证 0x1409c5900, 0x1A1B]：
//
//	调用 resolveDynamicShortcutDisplayIconData 替代静态版。
//	候选路径兜底调 resolveDynamicStartMenuAppIconData。
func resolveDynamicStartMenuIconData(path, targetPath, iconLocation string) (string, error) {
	iconData, customIconData := resolveDynamicShortcutDisplayIconData(targetPath, iconLocation)
	if iconData != "" {
		if customIconData != "" {
			return customIconData, nil
		}
		return iconData, nil
	}

	candidates := make([]string, 0, 3)
	if tp := strings.TrimSpace(targetPath); tp != "" {
		candidates = append(candidates, tp)
	}
	if sp := strings.TrimSpace(path); sp != "" {
		candidates = append(candidates, sp)
	}
	if il := strings.TrimSpace(iconLocation); il != "" {
		candidates = append(candidates, il)
	}

	for _, c := range candidates {
		class, _ := classifyAutomaticWindowsPath(c, nil)
		if class != classLocal {
			continue
		}
		if data := resolveDynamicStartMenuAppIconData(c, ""); data != "" {
			return data, nil
		}
	}
	return "", nil
}

// ========================================================================
// 快捷方式显示图标解析
// ========================================================================

// resolveShortcutDisplayIconData 解析快捷方式的显示图标数据（静态）。
// [S 汇编实证 0x1409c5ac0, 0x309B]：
//
//	参数 (targetPath, iconLocation) 均为 string。
//	1) parseShortcutIconLocation(iconLocation) 拆出 iconPath + iconIndex。
//	2) 解析成功且 classify(iconPath) == "local" ->
//	     resolveAppIconDataWithOptions(iconPath, {IconIndex: iconIndex})
//	     - 若 isShortcutCustomIcon(targetPath, iconPath) => 返回 (data, data)
//	     - 否则 => 返回 (data, "")
//	3) fallback：按 [trimmed iconLocation, trimmed targetPath] 顺序
//	   逐个 classify -> resolveAppIconDataWithOptions
//
// 返回 (iconData, customIconData) 两串。customIconData 非空表示显式自定义图标。
func resolveShortcutDisplayIconData(targetPath, iconLocation string) (string, string) {
	// ---- 精确图标位置解析 ----
	iconPath, _, iconIndex, ok := parseShortcutIconLocation(iconLocation)
	if ok && iconPath != "" {
		class, _ := classifyAutomaticWindowsPath(iconPath, nil)
		if class == classLocal {
			data := resolveAppIconDataWithOptions(iconPath, AppIconOptions{IconIndex: int32(iconIndex)})
			if data != "" {
				if isShortcutCustomIcon(targetPath, iconPath) {
					return data, data
				}
				return data, ""
			}
		}
	}

	// ---- fallback：2 候选 ----
	candidates := make([]string, 0, 2)
	// [S] 汇编先取 iconLocation（trimmed），后取 targetPath（trimmed）
	if il := strings.TrimSpace(iconLocation); il != "" {
		candidates = append(candidates, il)
	}
	if tp := strings.TrimSpace(targetPath); tp != "" {
		candidates = append(candidates, tp)
	}

	for _, c := range candidates {
		class, _ := classifyAutomaticWindowsPath(c, nil)
		if class != classLocal {
			continue
		}
		if data := resolveAppIconDataWithOptions(c, AppIconOptions{}); data != "" {
			return data, ""
		}
	}
	return "", ""
}

// resolveDynamicShortcutDisplayIconData 动态解析快捷方式显示图标。
// [S 汇编实证 0x1409c5de0, 0x292B]：
//
//	用 resolveDynamicStartMenuAppIconDataWithIndex / resolveDynamicStartMenuAppIconData
//	代替静态版的 resolveAppIconDataWithOptions。
func resolveDynamicShortcutDisplayIconData(targetPath, iconLocation string) (string, string) {
	iconPath, _, iconIndex, ok := parseShortcutIconLocation(iconLocation)
	if ok && iconPath != "" {
		class, _ := classifyAutomaticWindowsPath(iconPath, nil)
		if class == classLocal {
			data := resolveDynamicStartMenuAppIconDataWithIndex(iconPath, uint32(iconIndex))
			if data != "" {
				if isShortcutCustomIcon(targetPath, iconPath) {
					return data, data
				}
				return data, ""
			}
		}
	}

	// ---- fallback：2 候选 ----
	candidates := make([]string, 0, 2)
	if il := strings.TrimSpace(iconLocation); il != "" {
		candidates = append(candidates, il)
	}
	if tp := strings.TrimSpace(targetPath); tp != "" {
		candidates = append(candidates, tp)
	}

	for _, c := range candidates {
		class, _ := classifyAutomaticWindowsPath(c, nil)
		if class != classLocal {
			continue
		}
		if data := resolveDynamicStartMenuAppIconData(c, ""); data != "" {
			return data, ""
		}
	}
	return "", ""
}
