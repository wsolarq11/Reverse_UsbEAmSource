// AUTO-RECONSTRUCTED — DOMAIN: link icon mode normalize
// 研究用途. [S 汇编实证 0x14088b120, 320B]
//
// 调用参数（来自 normalizeLinkEntries 调用点）：
//
//	IconMode（TrimSpace+ToLower → 用于"upload"匹配检测，但[P]结果冗余）
//	IconData（TrimSpace+ToLower → 长度≥11 时与 11B 常量 memequal 判定上传图标）
//	IconRef（TrimSpace → 非空直接判定为上传图标）
//
// 返回值：
//
//	IconRef 非空                 → "upload"（6B）
//	IconRef 空 且 IconData 前缀匹配 11B 常量 → "upload"
//	否则                         → "favicon"（7B）
//
// 注：iconMode == "upload" 的检查在汇编中不影响结果结构（两个分支返回值相同），[P]编译器优化残留。
//
// 常量经 .rodata 解码确凿：
//   - 6B 返回串：upload（0x140C376B6）
//   - 7B 返回串：favicon（0x140C39449）
//   - 11B 比较常量：data:image/（0x140C47539，rod_vstr 解码确凿）
package main

import "strings"

// launcherLinkIconModeUploadPrefix 是 IconData 前缀常量，用于判定上传图标。
// .rodata 解码确凿：0x140C47539 → "data:image/"（11B）。
const launcherLinkIconModeUploadPrefix = "data:image/"

// normalizeLinkIconMode 规整链接条目的图标模式。
// [S 汇编 0x14088b120, 320B(0x140)]：IconRef 非空或 IconData 前缀 "data:image/" → "upload"；否则 "favicon"。
func normalizeLinkIconMode(iconMode, iconData, iconRef string) string {
	_ = strings.ToLower(strings.TrimSpace(iconMode)) // [S] 汇编检查"upload"但结果冗余
	iconRef = strings.TrimSpace(iconRef)
	iconData = strings.ToLower(strings.TrimSpace(iconData))

	var isUpload bool
	if iconRef != "" {
		isUpload = true
	} else {
		isUpload = len(iconData) >= 11 && iconData[:11] == launcherLinkIconModeUploadPrefix
	}

	if isUpload {
		return "upload"
	}
	return "favicon"
}
