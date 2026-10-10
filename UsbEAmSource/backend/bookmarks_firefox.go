// AUTO-RECONSTRUCTED BOOKMARKS (Firefox/Chromium source kind) — DOMAIN: bookmark source detect
// 研究用途. 逐寄存器还原（批次 375），asm 资产见 work/disasm/dump/b375_*.asm.txt。
package main

import (
	"path/filepath"
	"strings"
)

// resolveBookmarkSourceKind 解析书签源引擎类型（"chromium"/"firefox"/空）。
// [S 汇编 0x14076ad20, 448B(0x1c0)]：先 ToLower(TrimSpace(path)) 匹配
// edge/brave/chrome/vivaldi/chromium（chromium 系统一归 "chromium"）、firefox 归 "firefox"；
// 否则 Base(TrimSpace(source)) ToLower 匹配 "bookmarks"→"chromium"、"places.sqlite"→"firefox"；否则空。
func resolveBookmarkSourceKind(source, path string) string {
	switch strings.ToLower(strings.TrimSpace(path)) {
	case "edge", "brave", "chrome", "vivaldi", "chromium":
		return "chromium"
	case "firefox":
		return "firefox"
	}
	switch strings.ToLower(strings.TrimSpace(filepath.Base(strings.TrimSpace(source)))) {
	case "bookmarks":
		return "chromium"
	case "places.sqlite":
		return "firefox"
	}
	return ""
}

// resolveBookmarkBrowserKind 解析书签浏览器种类名。
// [S 汇编 0x14076aee0, 512B(0x200)]：ToLower(TrimSpace(path)) 非空→返回；kind=="firefox"→"firefox"；
// ToLower(TrimSpace(name)) 依次 Index "edge"/"brave"/"vivaldi"/"chromium"/"chrome" 命中返回；
// 否则返回 TrimSpace(kind)。
func resolveBookmarkBrowserKind(kind, name, path string) string {
	if p := strings.ToLower(strings.TrimSpace(path)); p != "" {
		return p
	}
	if kind == "firefox" {
		return "firefox"
	}
	n := strings.ToLower(strings.TrimSpace(name))
	switch {
	case strings.Contains(n, "edge"):
		return "edge"
	case strings.Contains(n, "brave"):
		return "brave"
	case strings.Contains(n, "vivaldi"):
		return "vivaldi"
	case strings.Contains(n, "chromium"):
		return "chromium"
	case strings.Contains(n, "chrome"):
		return "chrome"
	}
	return strings.TrimSpace(kind)
}

// describeBookmarkSourceDescriptor 描述书签源：返回 (kind, browserKind, name, resolvedPath)。
// [S 汇编 0x14076a9a0, 0x340B]：kind=resolveBookmarkSourceKind；firefox/chromium 走
// describeFirefox/ChromiumBookmarkPath + resolveBookmarkBrowserKind；否则
// browserKind=TrimSpace(path)、name=inferName(Dir(TrimSpace(source)))、resolvedPath 空。
func describeBookmarkSourceDescriptor(source, path string) (kind, browserKind, name, resolvedPath string) {
	kind = resolveBookmarkSourceKind(source, path)
	switch kind {
	case "firefox":
		name, resolvedPath = describeFirefoxBookmarkPath(source)
		browserKind = resolveBookmarkBrowserKind(kind, name, path)
	case "chromium":
		name, resolvedPath = describeChromiumBookmarkPath(source)
		browserKind = resolveBookmarkBrowserKind(kind, name, path)
	default:
		browserKind = strings.TrimSpace(path)
		name = inferName(filepath.Dir(strings.TrimSpace(source)))
	}
	return
}
