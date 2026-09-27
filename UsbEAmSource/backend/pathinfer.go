// AUTO-RECONSTRUCTED — DOMAIN: app working-dir inference
// 研究用途. [S 汇编]：looksLikeFilesystemPath(0x14088cd00) + inferWorkingDir(0x14088cc20)。
package main

import (
	"path/filepath"
	"strings"
)

// looksLikeFilesystemPath：含 '/' 或 '\' 或绝对路径 → true。[S]
func looksLikeFilesystemPath(s string) bool {
	if strings.ContainsRune(s, '/') || strings.ContainsRune(s, '\\') {
		return true
	}
	return filepath.IsAbs(s)
}

// inferWorkingDir：Trim + 是路径 → 取 dir（LastIndexAny `/\`）→ TrimRight 尾分隔符；无分隔返回空。[S]
func inferWorkingDir(path string) string {
	p := strings.TrimSpace(path)
	if p == "" || !looksLikeFilesystemPath(p) {
		return ""
	}
	i := strings.LastIndexAny(p, "/\\")
	if i <= 0 {
		return ""
	}
	return strings.TrimRight(p[:i], "/\\")
}
