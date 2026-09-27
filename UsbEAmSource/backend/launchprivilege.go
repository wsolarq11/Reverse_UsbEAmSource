// AUTO-RECONSTRUCTED — DOMAIN: app launch privilege mode canonical
// 研究用途. [S 汇编 0x140882680, 512B]：TrimSpace+ToLower 后按长度跳转表（len 5..15）分派枚举。
package main

import "strings"

// normalizeAppLaunchPrivilegeMode 规整启动权限模式。
// [S 汇编 0x140882680, 512B(0x200)]：TrimSpace+ToLower（@0x140882697/a0）；len-5 跳转表
// （@0x1408826a5/ba，表 @0x140c37da0，len<5 或 >15 → ""）；逐枚举 memequal：
//
//	admin/runas/administrator → "admin"（5B @0x140c35cef）
//	standard/normal/unelevated → "standard"（8B @0x140c3c20c）
//	follow/launcher/followlauncher/follow-launcher → "followLauncher"（14B @0x140c56758）
//	default → "default"（7B @0x140c39338）
//	其余（含空）→ ""（@0x14088283a）。
func normalizeAppLaunchPrivilegeMode(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "admin", "runas", "administrator":
		return "admin"
	case "standard", "normal", "unelevated":
		return "standard"
	case "follow", "launcher", "followlauncher", "follow-launcher":
		return "followLauncher"
	case "default":
		return "default"
	default:
		return ""
	}
}

// resolveEffectiveAppLaunchPrivilege 解析生效的启动权限模式。
// [S 汇编 0x140882880, 416B(0x1a0)]：normalizeAppLaunchPrivilegeMode(configured)
// （@0x1408828a1）为 admin/standard/followLauncher → 直接返回；否则
// normalizeAppLaunchPrivilegeMode(fallback)（@0x140882939）为三者之一 → 返回；
// 否则默认 "standard"（@0x1408829b9，8B @0x140c3c20c）。
func resolveEffectiveAppLaunchPrivilege(configured, fallback string) string {
	p := normalizeAppLaunchPrivilegeMode(configured)
	switch p {
	case "admin", "standard", "followLauncher":
		return p
	}
	p = normalizeAppLaunchPrivilegeMode(fallback)
	switch p {
	case "admin", "standard", "followLauncher":
		return p
	}
	return "standard"
}
