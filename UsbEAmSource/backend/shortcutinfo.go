// AUTO-RECONSTRUCTED — ShortcutInfo 结构 + COM 快捷方式解析器
// 研究用途
//
// ShortcutInfo 从 Win32 IShellLink COM 接口还原的快捷方式信息字段。
// 档位：[S] 结构签名 redress 确认；函数体依赖 go-ole COM（IShellLink + WScript.Shell），
//
//	需 go-ole 环境才能实际运行。
package main

import "strings"

// ShortcutInfo 快捷方式解析结果。
type ShortcutInfo struct {
	DisplayName      string // DisplayName / 显示名称
	TargetPath       string // IShellLink.GetPath
	Arguments        string // IShellLink.GetArguments
	WorkingDirectory string // IShellLink.GetWorkingDirectory
	IconLocation     string // IShellLink.GetIconLocation
	IconIndex        int    // 图标索引

	// internal
	pathForResolver string // 解析时的原路径（COM 失败时回退）
}

// resolveShortcutInfoWithIconResolver 通过 COM IShellLink 解析快捷方式信息。
// [S 地址 0x1409c2300, 2944B] — go-ole COM 外部集成。
// 先检测 .url 则调 readInternetShortcutInfo；否则 COM IShellLink 解析。
func resolveShortcutInfoWithIconResolver(path string) (ShortcutInfo, error) {
	p := strings.TrimSpace(path)
	if p == "" {
		return ShortcutInfo{}, newShortcutError("empty path")
	}

	if isInternetShortcutPath(p) {
		return readInternetShortcutInfo(p)
	}

	return ShortcutInfo{}, errShortcutCOMNotAvailable
}

var errShortcutCOMNotAvailable = newShortcutError("shortcut COM resolver not available in offline stub")

// newShortcutError 构造快捷方式错误。
// [S-inline 内联]：错误工厂（包装消息字符串），语义确定。
func newShortcutError(msg string) error {
	return &shortcutStubError{msg: msg}
}

type shortcutStubError struct {
	msg string
}

// Error 返回错误描述。
// [S-inline 内联]：消息透传，语义确定。
func (e *shortcutStubError) Error() string { return e.msg }
