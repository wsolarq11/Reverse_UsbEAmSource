// AUTO-RECONSTRUCTED — DOMAIN: twofactor sanitize
// 研究用途
package main

import "strings"

// sanitizeTwoFactorKind 归一化二因素种类。
// [S 汇编实证 0x1409c7ca0, 0x77B]：TrimSpace→ToLower；=="steam"(5B)/"steamguard"(10B)
// 返回 "steam"，其余一律 "totp"。
func sanitizeTwoFactorKind(kind string) string {
	k := strings.ToLower(strings.TrimSpace(kind))
	if k == "steam" || k == "steamguard" {
		return "steam"
	}
	return "totp"
}

// sanitizeTwoFactorAlgorithm 归一化二因素算法（受 kind 约束）。
// [S 汇编实证 0x1409c83e0, 0xc8B]：kind 经 sanitizeTwoFactorKind 为 "steam" → 强制 "SHA1"；
// 否则 algorithm TrimSpace→ToUpper 后 "SHA256"/"SHA512" 原样返回，其余 "SHA1"。
func sanitizeTwoFactorAlgorithm(algorithm, kind string) string {
	if sanitizeTwoFactorKind(kind) == "steam" {
		return "SHA1"
	}
	a := strings.ToUpper(strings.TrimSpace(algorithm))
	if a == "SHA256" {
		return "SHA256"
	}
	if a == "SHA512" {
		return "SHA512"
	}
	return "SHA1"
}

// looksLikeTwoFactorBase32Payload 判断字符串是否只含 base32 字符集。
// [S 汇编实证 0x1409c8320, 0x9dB]：TrimSpace 后逐 rune 校验 A-Z/a-z/2-7/=，空串返回 false，
// 全部合法返回 true。
func looksLikeTwoFactorBase32Payload(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
		case r >= 'a' && r <= 'z':
		case r >= '2' && r <= '7':
		case r == '=':
		default:
			return false
		}
	}
	return true
}

// sanitizeTwoFactorEntryIcon 归一化二因素条目图标标识。
// [S 汇编实证 0x1409c7d40, 0x77B]：kind 归一化为 "steam" → 返回 "steam"；否则 icon TrimSpace
// 后非空原样返回，空则返回 "lock"。返回字面量 "steam" 5B @0x140c35d8f、"lock" 4B @0x140c3498a。
func sanitizeTwoFactorEntryIcon(kind, icon string) string {
	if sanitizeTwoFactorKind(kind) == "steam" {
		return "steam"
	}
	i := strings.TrimSpace(icon)
	if i == "" {
		return "lock"
	}
	return i
}
