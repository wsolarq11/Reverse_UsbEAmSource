package main

import (
	"fmt"
	"strings"
)

// ValidationError 是插件 ID 校验失败错误。Reason 是失败原因码（1-9），
// ReservedName 仅在命中 Windows 保留名（Reason 9）时填充。
// [S 汇编 0x1406db280] 布局实证：Reason uint8@0x00，ReservedName string@0x08（16 字节）。
type ValidationError struct {
	Reason       uint8
	ReservedName string
}

// Error 返回校验失败的可读消息。
// [S 汇编 0x1406dad80, 384B] 实证：nil 与默认分支→"plugin id validation failed"@0x140c6b331,27B；
// Reason 1-9 跳转表；Reason 2 走 fmt.Sprintf("plugin id exceeds %d bytes",64)、
// Reason 9 走 fmt.Sprintf("plugin id uses Windows reserved name %s",ReservedName)，其余返回固定串。
func (e *ValidationError) Error() string {
	if e == nil {
		return "plugin id validation failed"
	}
	switch e.Reason {
	case 1:
		return "plugin id is empty"
	case 2:
		return fmt.Sprintf("plugin id exceeds %d bytes", 64)
	case 3:
		return "plugin id contains non-ASCII characters"
	case 4:
		return "plugin id contains a control character"
	case 5:
		return "plugin id contains a path separator"
	case 6:
		return "plugin id ends with a dot or space"
	case 7:
		return "plugin id must begin and end with an ASCII letter or digit"
	case 8:
		return "plugin id contains a disallowed character"
	case 9:
		return fmt.Sprintf("plugin id uses Windows reserved name %s", e.ReservedName)
	default:
		return "plugin id validation failed"
	}
}

// Validate 校验插件 ID（局部路径段语义）。
// [S 汇编 0x1406daf00, 896B] 实证：len==0→Reason1；len>64→Reason2；字符<0x20 或 ==0x7f→Reason4、
// >=0x80→Reason3；尾字符'.'/' '→Reason6；"."/".."/IndexAny("/\\:")≥0→Reason5；
// 首尾非 ASCII 字母数字→Reason7；其余字符非[字母数字._-]→Reason8；
// 首个'.'前段大写后命中 AUX/CON/NUL/PRN/COM1-9/LPT1-9→Reason9(ReservedName=该段)；否则 nil。
func Validate(id string) error {
	if len(id) == 0 {
		return &ValidationError{Reason: 1}
	}
	if len(id) > 64 {
		return &ValidationError{Reason: 2}
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		if c < 0x20 || c == 0x7f {
			return &ValidationError{Reason: 4}
		}
		if c >= 0x80 {
			return &ValidationError{Reason: 3}
		}
	}
	if id[len(id)-1] == '.' || id[len(id)-1] == ' ' {
		return &ValidationError{Reason: 6}
	}
	if id == "." || id == ".." || strings.IndexAny(id, "/\\:") >= 0 {
		return &ValidationError{Reason: 5}
	}
	first := id[0]
	last := id[len(id)-1]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || (first >= '0' && first <= '9')) ||
		!((last >= 'a' && last <= 'z') || (last >= 'A' && last <= 'Z') || (last >= '0' && last <= '9')) {
		return &ValidationError{Reason: 7}
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '.' || c == '_' || c == '-' {
			continue
		}
		return &ValidationError{Reason: 8}
	}
	idx := strings.IndexByte(id, '.')
	segment := id
	if idx >= 0 {
		segment = id[:idx]
	}
	upper := strings.ToUpper(segment)
	if len(upper) == 3 {
		switch upper {
		case "AUX", "CON", "NUL", "PRN":
			return &ValidationError{Reason: 9, ReservedName: upper}
		}
	} else if len(upper) == 4 {
		if (upper[:3] == "COM" || upper[:3] == "LPT") && upper[3] >= '1' && upper[3] <= '9' {
			return &ValidationError{Reason: 9, ReservedName: upper}
		}
	}
	return nil
}
