// filesearch_usn_follower_windows.go — USN 追随者判定与元数据（逆向还原）
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)
// 基线：docs/goresym/source_funcs.txt + docs/goresym/pipeline/tmp/wv2/*.asm.txt
package main

import "strings"

// shouldUseLightweightUSNFollowerForResourceMode 判定资源模式是否走轻量 USN 追随者。
// [S 汇编 0x1408186c0, 160B]：m := ToLower(TrimSpace(resourceMode))；len==8 且字节 ==
// "balanced"（0x6465636e616c6162）→ true；len==12 且前 8 字节 == "memory-s"
// （0x732d79726f6d656d）且后 4 字节 == "aver"（0x72657661）→ true；否则 false。
func shouldUseLightweightUSNFollowerForResourceMode(resourceMode string) bool {
	m := strings.ToLower(strings.TrimSpace(resourceMode))
	return m == "balanced" || m == "memory-saver"
}

// shouldUseLightweightUSNFollowerForFileSearchRuntime 判定文件搜索运行时是否走轻量 USN 追随者。
// [S 汇编 0x140818760, 256B]：enabled 非 nil 且 !enabled["files"]（@0x140818788, 5B
// key @0x140c65b46，mapaccess1_faststr @0x1408187a0）→ false；normalized :=
// normalizeFileSearchConfig(cfg)（@0x1408187d3）；!normalized.Enabled
// （@0x1408187fb 偏移 0x00）→ false；否则 shouldUseLightweightUSNFollowerForResourceMode(
// normalized.ResourceMode)（@0x140818819，偏移 0x68）。
func shouldUseLightweightUSNFollowerForFileSearchRuntime(enabled map[string]bool, cfg FileSearchConfig) bool {
	if enabled != nil && !enabled["files"] {
		return false
	}
	normalized := normalizeFileSearchConfig(cfg)
	if !normalized.Enabled {
		return false
	}
	return shouldUseLightweightUSNFollowerForResourceMode(normalized.ResourceMode)
}
