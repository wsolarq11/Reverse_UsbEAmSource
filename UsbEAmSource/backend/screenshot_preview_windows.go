package main

// screenshotPreviewShowWindowFlags 返回截图预览窗口显示 flags。
// [S] ASM 0x14099e7a0: mov eax, 0x53; ret（返回常量 0x53=83）。
func screenshotPreviewShowWindowFlags() int {
	return 0x53
}

// screenshotPreviewShowWindowCommand 返回截图预览窗口 ShowWindow 命令。
// [S] ASM 0x14099e7c0: mov eax, 4; ret（SW_SHOWNOACTIVATE）。
func screenshotPreviewShowWindowCommand() int {
	return 4
}

// waitBeforeRevealNative 原生窗口显示前等待（定时器/条件）。
// [S-sig 0x140999de0, 256B]：全局计数 <=0 → shouldDisplayNative；否则 time.NewTimer + chanrecv1
// 后 shouldDisplayNative。体待预览窗口域专项还原。
func (s *screenshotPreviewWindowService) waitBeforeRevealNative() bool {
	return false
}

// shouldDisplayNative 判定是否应以原生窗口显示预览。
// [S-sig 0x1409998e0, 288B]：lock → 读字段(+0xa8/+0xb0/+0x28/+0xb1) 判定 → unlock。
// 体待预览窗口域专项还原。
func (s *screenshotPreviewWindowService) shouldDisplayNative(a, b interface{}) bool {
	_, _ = a, b
	return false
}

// screenshotPreviewBounds 计算截图预览窗口边界（基于屏幕尺寸钳位）。
// [S-sig 0x14099b660, 288B]：ScreenManager.GetAll → 宽高减 x/y 再减 0x12 → clamp。
// 体待预览窗口域专项还原。
func screenshotPreviewBounds(a interface{}, x, y int64) (int64, int64, int64, int64) {
	_, _, _ = a, x, y
	return 0, 0, 0, 0
}
