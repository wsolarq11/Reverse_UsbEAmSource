// AUTO-RECONSTRUCTED FUNCTION — DOMAIN: app display name (fallback path, no Win32)
// 研究用途
//
// 契约来源：
//   - 函数签名：redress types all -m -v
//   - 行号蓝图：source_funcs.txt File: appdisplayname.go L5-29
//   - 反汇编：../work/disasm/dump/fallbackAppDisplayName.asm.txt (0x140748040, 576B)
//
// 档位：[S] 汇编实证（反汇编逐条追译）
package main

import "strings"

// fallbackAppDisplayName 通过纯字符串解析得出文件显示名（不上 Win32 Version API）。
// 语义：TrimSpace → 反斜杠归一化 → 取末段 → 扩展名校验链 → 剥离。
// [S 汇编实证 0x140748040]
func fallbackAppDisplayName(path string) string {
	p := strings.TrimSpace(path)
	if p == "" {
		return ""
	}
	p = strings.ReplaceAll(p, "\\", "/")
	if idx := strings.LastIndex(p, "/"); idx >= 0 {
		p = p[idx+1:]
	}
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	return stripDisplayExtensions(p)
}

// stripDisplayExtensions 检测并剥离已知扩展名。
// .rdata 常量验证的扩展名链（依序）：.exe(4B) .lnk(4B) .appref-ms(10B) .url(4B)
// [S 汇编实证：runtime.memequal × 4]
func stripDisplayExtensions(name string) string {
	lower := strings.ToLower(name)
	n := len(lower)

	if n >= 4 && lower[n-4:] == ".exe" {
		name = name[:n-4]
	} else if n >= 4 && lower[n-4:] == ".lnk" {
		name = name[:n-4]
	} else if n >= 10 && lower[n-10:] == ".appref-ms" {
		name = name[:n-10]
	} else if n >= 4 && lower[n-4:] == ".url" {
		name = name[:n-4]
	}

	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	return name
}
