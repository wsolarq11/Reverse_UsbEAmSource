// AUTO-RECONSTRUCTED — DOMAIN: hotkey modifier canonical
// 研究用途. [S 汇编实证 0x140889560, 544B]：TrimSpace+ToLower 后按常量识别修饰键→标准名。
// 分支覆盖：ctrl/alt/shift/win/meta/super/os/option；command 组返回名待 .rodata 精确（标 [P]）。
package main

import "strings"

// canonicalHotkeyModifier 将修饰键归一到标准名。
// [S 汇编 0x140889560, 544B(0x220)]：TrimSpace+ToLower 后按别名映射（.rodata 返回串 Win/Ctrl/Alt/Shift 实证）：
//
//	os/win/windows/meta/super            → "Win"
//	alt/option/optionoralt               → "Alt"
//	ctrl/control/cmd/command/cmdorctrl   → "Ctrl"
//	shift                                → "Shift"
//	其余 → ("", false)
func canonicalHotkeyModifier(s string) (string, bool) {
	m := strings.ToLower(strings.TrimSpace(s))
	switch m {
	case "os", "win", "windows", "meta", "super":
		return "Win", true
	case "alt", "option", "optionoralt":
		return "Alt", true
	case "ctrl", "control", "cmd", "command", "cmdorctrl":
		return "Ctrl", true
	case "shift":
		return "Shift", true
	default:
		return "", false
	}
}
