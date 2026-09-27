// AUTO-RECONSTRUCTED — DOMAIN: app shell-open / working dir inference
// 研究用途. [S] looksLikeShellProtocolTarget(0x1408a8680), requiresShellOpen(0x1408a8580),
// shouldInferWorkingDir(0x14088cda0)。
package main

import "strings"

// looksLikeShellProtocolTarget：TrimSpace 后形如 <scheme>: 。[S]
func looksLikeShellProtocolTarget(path string) bool {
	p := strings.TrimSpace(path)
	i := strings.IndexByte(p, ':')
	if i <= 0 {
		return false
	}
	if !isAsciiLetter(p[0]) {
		return false
	}
	for j := 0; j < i; j++ {
		c := p[j]
		if !(isAsciiLetter(c) || isAsciiDigit(c) || c == '+' || c == '-' || c == '.') {
			return false
		}
	}
	return true
}

// isAsciiLetter 判定 ASCII 字母。[S 汇编实证，内联谓词]
func isAsciiLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isAsciiDigit 判定 ASCII 数字。[S 汇编实证，内联谓词]
func isAsciiDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// requiresShellOpen：shell 协议目标，或扩展名 in {.lnk,.url,.appref-ms} → true。[S]
func requiresShellOpen(path string) bool {
	if looksLikeShellProtocolTarget(path) {
		return true
	}
	p := strings.TrimSpace(path)
	base := p
	if i := strings.LastIndexAny(base, "/\\"); i >= 0 {
		base = base[i+1:]
	}
	if i := strings.LastIndexByte(base, '.'); i >= 0 {
		switch strings.ToLower(base[i+1:]) {
		case "lnk", "url", "appref-ms":
			return true
		}
	}
	return false
}

// shouldInferWorkingDir：路径像 filesystem 且不需 shell 打开 → true。[S]
func shouldInferWorkingDir(path string) bool {
	if !looksLikeFilesystemPath(path) {
		return false
	}
	return !requiresShellOpen(path)
}
