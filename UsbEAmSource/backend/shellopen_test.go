package main

import "testing"

func TestLooksLikeShellProtocolTarget(t *testing.T) {
	if !looksLikeShellProtocolTarget("shell:AppsFolder") || !looksLikeShellProtocolTarget("ms-settings:display") {
		t.Error("scheme-like must be true")
	}
	if looksLikeShellProtocolTarget("plain") || looksLikeShellProtocolTarget("") {
		t.Error("non-scheme must be false")
	}
}

func TestRequiresShellOpen(t *testing.T) {
	// [S 汇编] X: 单字母方案视为协议 → 也触发 shell。
	if !requiresShellOpen(`C:\a\x.lnk`) || !requiresShellOpen("shell:AppsFolder") ||
		!requiresShellOpen("b.url") || !requiresShellOpen("x.appref-ms") || !requiresShellOpen(`C:\app.exe`) {
		t.Error("shell-target + lnk/url/appref-ms + drive-scheme must be true")
	}
	if requiresShellOpen(`dir\plain`) || requiresShellOpen("plain") {
		t.Error("non-shell plain must be false")
	}
}

func TestShouldInferWorkingDir(t *testing.T) {
	// 无 drive-scheme 的 filesystem 路径（如相对但含分隔符）且非 shell 扩展 → 推断 workdir。
	if !shouldInferWorkingDir(`x\app.exe`) {
		t.Error("filesystem non-shell path must be true")
	}
	if shouldInferWorkingDir("shell:AppsFolder") || shouldInferWorkingDir("plain") {
		t.Error("shell target / non-path must be false")
	}
}
