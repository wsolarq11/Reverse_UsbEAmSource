// AUTO-RECONSTRUCTED — DOMAIN: mousegestures (win32 action / monitor helpers)
// Source: UsbEAm_Launcher 1.0.3 (Go 1.25.12, PE64), disassembled from
// main.sendMouseGesture* / main.executeMouseGesture*Platform / main.mouseGesture* symbols.
// Tier markers:
//
//	[S VA]      — body fully translated from asm (pure geometry / math).
//	[S-sig VA]  — signature proven from asm + types_gesture.go; body is a faithful zero
//	              skeleton (SendInput / monitor / window deps). 「推断」= 类型来自命名/字段偏移。
//	[P]         — 签名待实证：register/stack layout is ambiguous; zero skeleton only.
//
// 研究用途
package main

// [S-sig 0x1408e5c40] sendMouseGestureButtonClick: 向当前目标发送一次按键点击。
func sendMouseGestureButtonClick(button string) error {
	return nil
}

// [S-sig 0x1408e5d40] sendMouseGestureButtonDown: 向当前目标发送按键按下。
func sendMouseGestureButtonDown(button string) error {
	return nil
}

// [S-sig 0x1408e5e00] sendMouseGestureButtonUp: 向当前目标发送按键抬起。
func sendMouseGestureButtonUp(button string) error {
	return nil
}

// [S-sig 0x1408e5ec0] mouseGestureButtonInputFlags: 按键名 → MOUSEEVENTF_* 标志组合。
func mouseGestureButtonInputFlags(button string) uint32 {
	return 0
}

// [S-sig 0x1408e5fe0] executeMouseGestureHotkeyPlatform: 通过 SendInput 发送热键组合。
func executeMouseGestureHotkeyPlatform(action GestureAction) error {
	return nil
}

// [S-sig 0x1408e6080] executeMouseGestureTextInputPlatform: 通过 SendInput 输入文本。
func executeMouseGestureTextInputPlatform(action GestureAction) error {
	return nil
}

// [S-sig 0x1408e6100] executeMouseGestureTaskSwitchPlatform: 触发任务切换（Alt-Tab 等）。
func executeMouseGestureTaskSwitchPlatform(action GestureAction) error {
	return nil
}

// [S-sig 0x1408e6160] sendMouseGestureInputs: 批量发送 mouseGestureInput 序列。
func sendMouseGestureInputs(inputs []mouseGestureInput) error {
	return nil
}

// [S-sig 0x1408e6360] buildMouseGestureTextInputInputs: 由文本构造键盘输入序列。
func buildMouseGestureTextInputInputs(text string) []mouseGestureInput {
	return nil
}

// [S-sig 0x1408e65a0] buildMouseGestureHotkeyInputs: 由热键串构造键盘输入序列。
func buildMouseGestureHotkeyInputs(hotkey string) []mouseGestureInput {
	return nil
}

// [S-sig 0x1408e6b00] mouseGestureVirtualKey: 键名 → 虚拟键码。
func mouseGestureVirtualKey(name string) uint32 {
	return 0
}

// [S-sig 0x1408e6c40] executeMouseGestureWindowCommandPlatform: 执行窗口命令（最小化/最大化/关闭等）。
func executeMouseGestureWindowCommandPlatform(action GestureAction) error {
	return nil
}

// [S-sig 0x1408e7a00] mouseGestureMonitorHandleFromScreenID: 屏幕 ID → HMONITOR。
func mouseGestureMonitorHandleFromScreenID(screenID string) uintptr {
	return 0
}

// [S-sig 0x1408e7aa0] mouseGestureMonitorInfo: 屏幕 ID → 屏幕边界信息（返回类型推断）。
func mouseGestureMonitorInfo(screenID string) MouseGestureScreenBounds {
	return MouseGestureScreenBounds{}
}

// [S-sig 0x1408e7c40] 窗口移动位置钳制（纯几何叶子函数，无 morestack 序言）。
// 参数：6 寄存器(eax/ebx/ecx/edi/esi/r8d) + 4 栈槽([rsp+8..0x14]) = 10×int32（dword 运算证实 32 位）；
// 体：sub/lea/cmp/jg 矩形边界钳制；尾声 mov eax,ecx + mov ebx,r8d = 返回 (int32,int32)。
// 参数语义归属（矩形/屏幕/移动增量组合）无法从 asm 唯一确定，命名保留占位线索。
func mouseGestureWindowMovePosition(x, y, w, h, wx, wy, ww, wh, m0, m1 int32) (int32, int32) {
	return 0, 0
}

// [S-sig 0x1408e7ca0] activateMouseGestureActionTargetPlatform: 激活目标窗口（前台/焦点）。
func activateMouseGestureActionTargetPlatform(target MouseGestureActionTarget) {
}

// [S-sig 0x1408e7ea0] mouseGestureRootWindow: 获取（或创建）手势根窗口句柄。
func mouseGestureRootWindow() uintptr {
	return 0
}

// [S-sig 0x1408e7fa0] mouseGestureShowWindow: 显示/隐藏指定窗口。
func mouseGestureShowWindow(hwnd uintptr, show bool) {
}
