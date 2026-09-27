// AUTO-RECONSTRUCTED — DOMAIN: hotkey binding canonicalize
// 研究用途. [S 汇编实证 0x1408887c0 前半]：TrimSpace → Split("+")（分隔符 .ro 解码 "＋"）→
// 逐段 canonicalHotkeyModifier，任一非修饰键 → (nil,false)；否则去重保留顺序。
package main

import "strings"

// canonicalizeShortcutModifiers 归一快捷方式修饰键序列。
// [S 汇编实证 0x1408887c0 前半]：TrimSpace → Split("+") → 逐段 canonicalHotkeyModifier → 去重保留顺序。
func canonicalizeShortcutModifiers(binding string) ([]string, bool) {
	binding = strings.TrimSpace(binding)
	if binding == "" {
		return nil, false
	}
	parts := strings.Split(binding, "+")
	out := make([]string, 0, len(parts))
	seen := make(map[string]bool, len(parts))
	for _, p := range parts {
		canon, ok := canonicalHotkeyModifier(p)
		if !ok {
			return nil, false
		}
		if !seen[canon] {
			seen[canon] = true
			out = append(out, canon)
		}
	}
	return out, true
}
