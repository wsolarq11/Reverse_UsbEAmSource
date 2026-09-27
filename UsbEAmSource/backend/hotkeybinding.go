// AUTO-RECONSTRUCTED — DOMAIN: hotkey binding normalizer
// 研究用途. [S 汇编实证 0x1408887c0, 336 行 + 4 辅助函数 410 行]
package main

import (
	"sort"
	"strings"
)

// hotkeyBindingPredicate 提供快捷键键名验证函数。
// 汇编实证 normalizeShortcutBindingWithOptions @0x140888847：
// `mov rdx,[rsp+0x1e8]; mov rcx,[rdx]` 取 pred 偏移 0 的函数值，
// `call rcx` 以 key string 调用，返回 cl(bool)。
type hotkeyBindingPredicate struct {
	isValidKey func(key string) bool
}

// launcherHotkeyBindingPredicate 是启动器热键的包级静态谓词（汇编 0x1410969a0）。
// 原始 .data 布局为 16 字节双函数值：{canonicalHotkeyKey(0x140888f20),
// canonicalSearchCategoryShortcutKey(0x140889480)}；normalizeShortcutBindingWithOptions
// 仅取偏移 0（canonicalHotkeyKey 的 bool 语义），故以闭包固定首字段语义。
var launcherHotkeyBindingPredicate = &hotkeyBindingPredicate{
	isValidKey: func(key string) bool {
		_, ok := canonicalHotkeyKey(key)
		return ok
	},
}

// standaloneModifierNames 是 flag2 路径下允许作为独立快捷键的规范名。
var standaloneModifierNames = map[string]bool{
	"Alt": true, "Cmd": true, "Win": true, "Ctrl": true,
	"Shift": true, "Super": true, "Option": true,
}

// shortcutModifierSortWeight 返回修饰键排序权重。
// [S 汇编 0x140888e60, 74 行] 纯表驱动：
// "Alt"=0, "Option"=0, "Ctrl"=1, "Shift"=2, "Cmd/Win/Super"=3, 其余=99
func shortcutModifierSortWeight(s string) int {
	switch s {
	case "Alt", "Option":
		return 0
	case "Ctrl":
		return 1
	case "Shift":
		return 2
	case "Cmd", "Win", "Super":
		return 3
	default:
		return 99
	}
}

// canonicalSearchCategoryShortcutKey 规范快捷键的键部分（搜索分类路径）。
// [S 汇编 0x140889480, 80 行]：
// TrimSpace → ToLower → 空→false → “ ` “/`~`/`backquote` → "~",true
// → 否则 fallthrough canonicalHotkeyKey(原始key, 非 lowered)
func canonicalSearchCategoryShortcutKey(key string) (string, bool) {
	lower := strings.ToLower(strings.TrimSpace(key))
	if lower == "" {
		return "", false
	}
	if lower == "`" || lower == "~" || lower == "backquote" {
		return "~", true
	}
	return canonicalHotkeyKey(strings.TrimSpace(key))
}

// canonicalHotkeyKey 规范键盘键名。
// [S 汇编 0x140888f20, 538 行] 精简实现（覆盖 asm 跳转表全部键）。
func canonicalHotkeyKey(key string) (string, bool) {
	trimmed := strings.TrimSpace(key)
	lower := strings.ToLower(trimmed)
	if lower == "" {
		return "", false
	}

	// 单字符 a-z/0-9 → 大写
	if len(lower) == 1 {
		c := lower[0]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			return strings.ToUpper(trimmed), true
		}
	}

	switch lower {
	case "esc", "escape":
		return "Esc", true
	case "tab":
		return "Tab", true
	case "space":
		return "Space", true
	case "enter", "return":
		return "Enter", true
	case "backspace", "bs":
		return "Backspace", true
	case "insert", "ins":
		return "Insert", true
	case "delete", "del":
		return "Delete", true
	case "home":
		return "Home", true
	case "end":
		return "End", true
	case "pageup", "pgup":
		return "PageUp", true
	case "pagedown", "pgdn", "pgdown":
		return "PageDown", true
	case "up":
		return "Up", true
	case "down":
		return "Down", true
	case "left":
		return "Left", true
	case "right":
		return "Right", true
	case "printscreen", "prtsc", "prtscr":
		return "PrintScreen", true
	case "scrolllock":
		return "ScrollLock", true
	case "pause", "break":
		return "Pause", true
	case "numlock":
		return "NumLock", true
	case "capslock", "caps":
		return "CapsLock", true
	case "shift":
		return "Shift", true
	case "ctrl", "control":
		return "Ctrl", true
	case "alt":
		return "Alt", true
	case "meta":
		return "Meta", true
	case "super":
		return "Super", true
	case "os":
		return "OS", true
	case "win", "windows":
		return "Win", true
	case "option":
		return "Option", true
	case "command", "cmd":
		return "Cmd", true
	case "menu", "apps":
		return "Menu", true
	// F<num> (1-24)
	default:
		if len(lower) > 1 && lower[0] == 'f' {
			n := 0
			ok := true
			for i := 1; i < len(lower); i++ {
				c := lower[i]
				if c < '0' || c > '9' {
					ok = false
					break
				}
				n = n*10 + int(c-'0')
			}
			if ok && n >= 1 && n <= 24 {
				if n < 10 {
					return "F" + string(rune('0'+n)), true
				}
				return "F" + string(rune('0'+n/10)) + string(rune('0'+n%10)), true
			}
		}
		return "", false
	}
}

// normalizeShortcutBindingWithOptions 规范化快捷键绑定。
// [S 汇编 0x1408887c0, 336 行] 实证流程：
//
//	TrimSpace → strings.Split("+") →
//	取最后一段 → pred.isValidKey(key) —— 失败则 flag2 && len==1 时 canonicalHotkeyModifier 兜底
//	前 len-1 段逐段 TrimSpace + canonicalHotkeyModifier，任一失败 → ""
//	map 去重 → 无修饰键时 flag1/flag2/flag3 权限校验 →
//	sort.Slice(weight 降序: Alt/Option=0 < Ctrl=1 < Shift=2 < Cmd/Win/Super=3 < 99)
//	→ append(key) → strings.Join(prefix, "+")
//
// 入口 rax/rbx=binding, rcx=pred, dil/sil/r8b=flag1/2/3, 出口 string
func normalizeShortcutBindingWithOptions(binding string, pred *hotkeyBindingPredicate, flag1, flag2, flag3 bool) string {
	binding = strings.TrimSpace(binding)
	if binding == "" {
		return ""
	}

	parts := strings.Split(binding, "+")
	if len(parts) == 0 {
		return ""
	}

	// ---- 验证键部分 ----
	key := strings.TrimSpace(parts[len(parts)-1])
	if key == "" {
		return ""
	}

	isValid := false
	if pred != nil && pred.isValidKey != nil {
		isValid = pred.isValidKey(key)
	} else {
		// 缺省：用 canonicalSearchCategoryShortcutKey
		_, isValid = canonicalSearchCategoryShortcutKey(key)
	}

	if !isValid {
		if flag2 && len(parts) == 1 {
			if canon, ok := canonicalHotkeyModifier(key); ok {
				key = canon
				isValid = true
			}
		}
		if !isValid {
			return ""
		}
	}

	// ---- 处理修饰键前缀 ----
	prefixCount := len(parts) - 1
	prefix := make([]string, 0, prefixCount)
	seen := make(map[string]struct{}, prefixCount)

	for i := 0; i < prefixCount; i++ {
		mod := strings.TrimSpace(parts[i])
		canon, ok := canonicalHotkeyModifier(mod)
		if !ok {
			return ""
		}
		if _, exists := seen[canon]; !exists {
			seen[canon] = struct{}{}
			prefix = append(prefix, canon)
		}
	}

	// ---- 无修饰键时的权限校验 ----
	if len(prefix) == 0 && !flag3 {
		ok := false
		if flag1 && key == "PrintScreen" {
			ok = true
		}
		if !ok && flag2 && standaloneModifierNames[key] {
			ok = true
		}
		if !ok {
			return ""
		}
	}

	// ---- 权重排序（asm setg = 降序） ----
	sort.SliceStable(prefix, func(i, j int) bool {
		return shortcutModifierSortWeight(prefix[i]) > shortcutModifierSortWeight(prefix[j])
	})

	// ---- 拼装 ----
	all := append(prefix, key)
	return strings.Join(all, "+")
}
