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
