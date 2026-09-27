// AUTO-RECONSTRUCTED — DOMAIN: plugin path reparse-point detection
// 研究用途。反汇编实证：
//
//	pluginPathHasReparsePoint 0x140923b40, 160B
//
// 简化 bool 版：UTF16FromString → GetFileAttributes
// error 时保守返 true（防止操作 reparse 点出错）。
package main

import (
	"syscall"

	"golang.org/x/sys/windows"
)

// pluginPathHasReparsePoint 检查路径是否含 reparse point。
// 保守实现：若无法获取属性则返回 true。
// [S 汇编 0x140923b40, 160B]：UTF16FromString → GetFileAttributes；error 时保守返 true。
func pluginPathHasReparsePoint(p string) bool {
	utf16, err := syscall.UTF16FromString(p)
	if err != nil || len(utf16) <= 1 {
		return true
	}
	ptr := &utf16[0]
	attrs, err := windows.GetFileAttributes(ptr)
	if err != nil {
		return true
	}
	return attrs&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0
}
