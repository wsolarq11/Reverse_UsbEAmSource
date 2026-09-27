package main

import "image"

// windowManagementVirtualScreenBounds 返回虚拟屏幕矩形。
// [S] ASM 0x1409edbc0：四次 GetSystemMetrics（0x4c/0x4d/0x4e/0x4f =
// SM_X/Y/CX/CYVIRTUALSCREEN），尾段 edx=x+cx / edi=y+cy 合成
// image.Rect(x, y, x+cx, y+cy)。无回退分支（与 qrCodeVirtualScreenBounds 不同）。
func windowManagementVirtualScreenBounds() image.Rectangle {
	x := getSystemMetrics(0x4c) // SM_XVIRTUALSCREEN
	y := getSystemMetrics(0x4d) // SM_YVIRTUALSCREEN
	w := getSystemMetrics(0x4e) // SM_CXVIRTUALSCREEN
	h := getSystemMetrics(0x4f) // SM_CYVIRTUALSCREEN
	return image.Rect(x, y, x+w, y+h)
}
