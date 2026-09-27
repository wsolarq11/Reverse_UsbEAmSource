package main

// screenshotSelectionToolbarShowWindowCommand 返回选区工具栏 ShowWindow 命令。
// [S] ASM 0x1409ad200: mov eax, 8; ret（SW_SHOWNA）。
func screenshotSelectionToolbarShowWindowCommand() int {
	return 8
}

// screenshotSelectionToolbarShowWindowFlags 返回选区工具栏显示 flags。
// [S] ASM 0x1409ad220: mov eax, 0x53; ret（返回常量 0x53=83）。
func screenshotSelectionToolbarShowWindowFlags() int {
	return 0x53
}
