// AUTO-RECONSTRUCTED — DOMAIN: internet shortcut (.url) read + parse → ShortcutInfo
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)
//
// 函数来源：
//
//	readInternetShortcutInfo 0x140869b00  (800B)
package main

import "strings"

// readInternetShortcutInfo 读取 .url 快捷方式文件，解析其内容并返回 ShortcutInfo。
// [S 0x140869b00]
func readInternetShortcutInfo(path string) (ShortcutInfo, error) {
	data, err := readInternetShortcutFile(path)
	if err != nil {
		return ShortcutInfo{}, err
	}
	text := decodeInternetShortcutTextBounded(data)
	vals := parseInternetShortcutValues(text)
	// 取 URL 键值作为显示名
	url, _ := vals["url"]
	url = strings.TrimSpace(url)
	return ShortcutInfo{DisplayName: url}, nil
}
