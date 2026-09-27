// AUTO-RECONSTRUCTED — DOMAIN: config icon data error
// 研究用途. [S 汇编实证 0x14088d660, 224B]：fmt.Sprintf("%s: %s: %s", code, Path, Reason)；
// nil receiver 返回 code 本身。code=CONFIG_ICON_DATA_INVALID（.rodata 解码确凿）。
package main

import "fmt"

const launcherConfigIconDataCode = "CONFIG_ICON_DATA_INVALID"

// Error 返回错误描述。
// [S 汇编 0x14088d660, 224B(0xe0)]：nil 收者 → code；否则 fmt.Sprintf("%s: %s: %s", code, Path, Reason)。
func (e *launcherConfigIconDataError) Error() string {
	if e == nil {
		return launcherConfigIconDataCode
	}
	return fmt.Sprintf("%s: %s: %s", launcherConfigIconDataCode, e.Path, e.Reason)
}
