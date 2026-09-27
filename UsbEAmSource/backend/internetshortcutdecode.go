// AUTO-RECONSTRUCTED — DOMAIN: internet shortcut text decode
// 研究用途. 实证依据 dump 反汇编：
//
//	decodeInternetShortcutTextBounded（576B, 0x14086ab60）：UTF-8 BOM(EF BB BF)/UTF-16 LE(FF FE)
//	/UTF-16 BE(FE FF) 探测后剥除 BOM；超过 0x40000 字节返回空。
//	decodeUTF16ShortcutText（800B, 0x14086ada0）：逐 2 字节经字节序回调累积为 uint16 码元，
//	null 码元跳过，最后 unicode/utf16.decode + slicerunetostring（0x14086b0b6/0x14086b13）。
package main

import (
	"unicode/utf16"
)

// decodeInternetUicon 从 UTF 字节序列解码 .url 快捷方式文本（剥除 BOM）。
// [S 汇编 0x14086ab60, 576B]：UTF-8/UTF-16 LE/BE BOM 探测剥除；>0x40000 字节返回空。
func decodeInternetShortcutTextBounded(b []byte) string {
	if len(b) > 0x40000 {
		return ""
	}
	if len(b) == 0 {
		return ""
	}
	// UTF-8 BOM: EF BB BF
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		return string(b[3:])
	}
	// UTF-16 LE BOM: FF FE
	if len(b) >= 2 && b[0] == 0xFF && b[1] == 0xFE {
		return decodeUTF16ShortcutText(b[2:], false)
	}
	// UTF-16 BE BOM: FE FF
	if len(b) >= 2 && b[0] == 0xFE && b[1] == 0xFF {
		return decodeUTF16ShortcutText(b[2:], true)
	}
	return string(b)
}

// decodeUTF16ShortcutText 把 UTF-16 字节序列解码为字符串（可指定字节序；null 码元跳过）。
// [S 汇编 0x14086ada0, 800B]：逐 2 字节经字节序回调累积 uint16；null 跳过；unicode/utf16.Decode。
func decodeUTF16ShortcutText(b []byte, bigEndian bool) string {
	if len(b) < 2 {
		return ""
	}
	var units []uint16
	for i := 0; i+1 < len(b); i += 2 {
		var u uint16
		if bigEndian {
			u = uint16(b[i])<<8 | uint16(b[i+1])
		} else {
			u = uint16(b[i]) | uint16(b[i+1])<<8
		}
		if u == 0 {
			continue
		}
		units = append(units, u)
	}
	if len(units) == 0 {
		return ""
	}
	return string(utf16.Decode(units))
}
