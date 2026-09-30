package main

// screenshot 滚动所需的 user32 LazyProc（依赖 appicon_windows.go 中的 user32DLL）。
var procGetAsyncKeyState = user32DLL.NewProc("GetAsyncKeyState")

// readScreenshotScrollingRightButtonState 读取鼠标右键状态（GetAsyncKeyState，VK_RBUTTON=2）。
// [S] ASM 0x1409a3320：Call(2)；`bt eax,0xf; setb al` 取 bit15 → 当前是否按下（第 1 返回值）；
// `and ebx,1` 取 bit0 → 自上次调用后是否按过（第 2 返回值）。返回 (bool, bool)。
func readScreenshotScrollingRightButtonState() (bool, bool) {
	state, _, _ := procGetAsyncKeyState.Call(2) // VK_RBUTTON
	return state&0x8000 != 0, state&1 != 0
}
