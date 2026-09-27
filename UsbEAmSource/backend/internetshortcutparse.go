// AUTO-RECONSTRUCTED — DOMAIN: internet shortcut values parse
// 研究用途. 实证依据 dump 反汇编（1056B，0x14086a5a0）。
// 汇编确定性语义（逐条对齐）：
//   - makemap_small 建空 map（0x14086a5ca）
//   - 按 '\n' 切行（strings.genSplit, 0x14086a5f9）
//   - 每行：去尾 '\r'(0x0d, 0x14086a66b) → TrimSpace（0x14086a6c2）→ 判首字符：
//     ';'(0x3b) / '#'(0x23) 注释行 → 跳过（0x14086a700/0x743）
//     '['(0x5b) 分区头 → 检查尾 ']'（0x14086a76f），不入 map（0x14086a925）
//     否则 → internal.Cut(line, '=')（0x14086a806）取键值对
//     key = TrimSpace → ToLower（0x14086a843/0x848），val = TrimSpace（0x14086a892）
//     → mapassign_faststr 写入（0x14086a8c0）
package main

import (
	"strings"
)

// parseInternetShortcutValues 解析 .url/.lnk 内容为键值 map（注释/分区头忽略）。
// [S 汇编 0x14086a5a0, 1056B]：按 '\n' 切行；去 '\r'→TrimSpace；';'/'#'/'[' 跳过；
// strings.Cut('=') 取键值；key ToLower、val TrimSpace → mapassign_faststr。
func parseInternetShortcutValues(content string) map[string]string {
	out := make(map[string]string)
	for _, line := range strings.Split(content, "\n") {
		// 去尾 '\r'（CRLF）
		if strings.HasSuffix(line, "\r") {
			line = line[:len(line)-1]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 注释 / 分区头忽略
		if line[0] == ';' || line[0] == '#' || line[0] == '[' {
			continue
		}
		if k, v, ok := strings.Cut(line, "="); ok {
			k = strings.ToLower(strings.TrimSpace(k))
			v = strings.TrimSpace(v)
			out[k] = v
		}
	}
	return out
}
