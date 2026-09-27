// AUTO-RECONSTRUCTED — DOMAIN: command line split
// 研究用途. 实证依据 dump 反汇编（1440B，0x1408a7ee0）。
// 汇编确定性语义（Windows 命令行解析规则，逐条对齐）：
//   - 输入 TrimSpace；空 → nil（0x1408a7f05/0x7f0a）
//   - []rune 扫描；逐符文判定（0x1408a7fd1 起）
//   - 引号状态 esi：单引号 0x27 / 双引号 0x22 / 0 无；引号本身不写入输出（语法标记）
//   - 反斜杠 '\\'：单引号内作普通字符；否则计数连续反斜杠（0x1408a8297）
//     后随 '"'：每 2 反斜杠写 1 个并翻转双引号状态；奇数额外写 1 个字面 '"'（0x1408a83b6）
//     否则：原样写全部反斜杠（0x1408a8343）
//   - 空白 ' ' 0x20 / tab 0x09 / LF 0x0a / CR 0x0d / VT 0x0b 且不在引号内 → 提交 token（0x1408a8050）
//   - 其他 rune → strings.Builder 累积（0x1408a8145）
package main

import "strings"

// splitCommandLineArguments 按 Windows 命令行规则把参数字符串分割为 token 切片。
// [S 汇编 0x1408a7ee0, 1440B(0x5a0)]：TrimSpace→[]rune 扫描；引号/反斜杠/空白规则逐条对齐。
func splitCommandLineArguments(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	runes := []rune(s)
	res := make([]string, 0, 4)
	var b strings.Builder
	var quote rune // 0 / '\'' / '"'
	n := len(runes)
	for i := 0; i < n; {
		r := runes[i]
		switch {
		case r == '\\':
			if quote == '\'' {
				b.WriteRune('\\')
				i++
				continue
			}
			start := i
			for i < n && runes[i] == '\\' {
				i++
			}
			cnt := i - start
			if i < n && runes[i] == '"' {
				half := cnt / 2
				for k := 0; k < half; k++ {
					b.WriteByte('\\')
				}
				if cnt%2 == 1 {
					b.WriteByte('"')
				} else {
					if quote == '"' {
						quote = 0
					} else {
						quote = '"'
					}
				}
				i++
			} else {
				for k := 0; k < cnt; k++ {
					b.WriteByte('\\')
				}
			}
		case r == '\'' || r == '"':
			if quote == 0 {
				quote = r
			} else if quote == r {
				quote = 0
			}
			i++
		case quote == 0 && (r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\v'):
			if b.Len() > 0 {
				res = append(res, b.String())
				b.Reset()
			}
			i++
		default:
			b.WriteRune(r)
			i++
		}
	}
	if b.Len() > 0 {
		res = append(res, b.String())
	}
	return res
}
