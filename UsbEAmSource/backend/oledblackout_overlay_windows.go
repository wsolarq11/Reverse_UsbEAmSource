// AUTO-RECONSTRUCTED — DOMAIN: oledblackout overlay (Win32 原生窗口侧)
// Source: UsbEAm_Launcher 1.0.3 (Go 1.25.12, PE64), disassembled from
// main.oledBlackoutNativeOverlayWindow* / main.oledBlackoutUseNativeOverlayWindow
// / main.createOLEDBlackoutNativeOverlayWindow / main.ensureOLEDBlackoutNativeOverlayWindowClass
// / main.oledBlackoutNativeOverlayWindowProc symbols. Tier markers:
//
//	[S VA]      — body fully translated from asm (pure geometry / math).
//	[S-sig VA]  — signature proven from asm; body is a faithful zero skeleton.
//	[P]         — signature not yet proven; zero skeleton + blocking reason.
//
// receiver *oledBlackoutNativeOverlayWindow 已在 types_oled.go 落地。
// 研究用途
package main

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

// [S 0x1409111c0] 编译期常量（Windows 构建标签下恒为 true）。
// asm：mov eax,1; ret。
func oledBlackoutUseNativeOverlayWindow() bool {
	return true
}

// [S-sig 0x1409111e0] 构造原生遮罩窗口并等待 ready。
// 序言 test rax（overlay 指针 nil 检查）+ 保存 rax + rdi/rsi/rcx/rbx（4×int = bounds application.Rect）；
// call ensureOLEDBlackoutNativeOverlayWindowClass；尾声返回 rax=指针 + rbx/rcx=error（3 字）。体骨架。
func createOLEDBlackoutNativeOverlayWindow(overlay *oledBlackoutOverlayWindow, bounds application.Rect) (*oledBlackoutNativeOverlayWindow, error) {
	_ = overlay
	_ = bounds
	return nil, nil
}

// [S-sig 0x140911440] 幂等注册原生窗口类（sync.Once）。
// 序言无参数保存（morestack_noctxt 空）；体：读全局 done 标志 test 非零即跳过，否则
// sync.Once.doSlow(&once, func1)（func1=0x1409f6180）；尾声 mov rax/rbx 两个相邻全局，
// 返回 2 字（窗口类原子 + 实例句柄）。体骨架。
func ensureOLEDBlackoutNativeOverlayWindowClass() (uintptr, uintptr) {
	return 0, 0
}

// [S-sig 0x1409114a0] 原生窗口消息线程：LockOSThread → 建窗 → 泵消息 → UnlockOSThread。无返回值。
func (w *oledBlackoutNativeOverlayWindow) run() {
	_ = w
}

// [S-sig 0x1409117a0] 注册类后 CreateWindow 建窗并应用 chrome。asm 尾部清零 2 字 = 返回 error。
func (w *oledBlackoutNativeOverlayWindow) createWindow() error {
	_ = w
	return nil
}

// [S-sig 0x140911a40] Win32 窗口过程。asm：4 实参(hwnd,msg,wparam,lparam)，返回 LRESULT。
func oledBlackoutNativeOverlayWindowProc(hwnd uintptr, msg uint32, wparam uintptr, lparam uintptr) uintptr {
	_, _, _, _ = hwnd, msg, wparam, lparam
	return 0
}

// [S-sig 0x140911c00] 分发窗口消息（switch on msg，0x200=WM_MOUSEMOVE 等）。返回 LRESULT。
// asm：receiver=RAX + hwnd=BX + msg=CX(uint32) + wparam=DI + lparam=SI。
func (w *oledBlackoutNativeOverlayWindow) handleMessage(hwnd uintptr, msg uint32, wparam uintptr, lparam uintptr) uintptr {
	_, _, _, _ = hwnd, msg, wparam, lparam
	_ = w
	return 0
}

// [S-sig 0x140911f80] 线程内 ShowWindow。asm 零清 2 字 = 返回 error。
func (w *oledBlackoutNativeOverlayWindow) Show() error {
	_ = w
	return nil
}

// [S-sig 0x140912000] 线程内隐藏窗口。返回 error。
func (w *oledBlackoutNativeOverlayWindow) Hide() error {
	_ = w
	return nil
}

// [S-sig 0x140912080] 关闭窗口（sync.Once 保证单次）。返回 error。
func (w *oledBlackoutNativeOverlayWindow) Close() error {
	_ = w
	return nil
}

// [S-sig 0x140912180] 线程内 SetForegroundWindow。无返回值。
func (w *oledBlackoutNativeOverlayWindow) Focus() {
	_ = w
}

// [S-sig 0x140912200] 更新 bounds 并转发 setBoundsOnThread。receiver + 4 整型实参，无返回值。
func (w *oledBlackoutNativeOverlayWindow) SetBounds(bounds application.Rect) {
	_ = w
	_ = bounds
}

// [S-sig 0x140912340] 返回原生窗口句柄（内部调用 handle()）。
func (w *oledBlackoutNativeOverlayWindow) NativeWindowHandle() uintptr {
	_ = w
	return 0
}

// [S-sig 0x140912380] 应用窗口 chrome（去除边框/置顶/透明等），收尾调用 oledBlackoutSuppressWindowBorder。
func (w *oledBlackoutNativeOverlayWindow) applyWindowChrome(hwnd uintptr) {
	_ = w
	_ = hwnd
}

// [S-sig 0x140912520] 消息线程上 ShowWindow 并置 visible 标志。无返回值。
func (w *oledBlackoutNativeOverlayWindow) showOnThread() {
	_ = w
}

// [S-sig 0x140912660] 消息线程上隐藏窗口并清 visible 标志。无返回值。
func (w *oledBlackoutNativeOverlayWindow) hideOnThread() {
	_ = w
}

// [S-sig 0x1409126e0] 消息线程上聚焦窗口。无返回值。
func (w *oledBlackoutNativeOverlayWindow) focusOnThread() {
	_ = w
}

// [S-sig 0x140912860] 消息线程上 SetWindowPos 应用 bounds。receiver + 4 整型实参，无返回值。
func (w *oledBlackoutNativeOverlayWindow) setBoundsOnThread(bounds application.Rect) {
	_ = w
	_ = bounds
}

// [S-sig 0x140912a00] WM_PAINT 处理：以黑色填充客户区。receiver + hdc 实参，无返回值。
func (w *oledBlackoutNativeOverlayWindow) paint(hdc uintptr) {
	_ = w
	_ = hdc
}

// [S-sig 0x140912c00] 加锁写入 hwnd 字段。无返回值。
func (w *oledBlackoutNativeOverlayWindow) setHandle(hwnd uintptr) {
	_ = w
	_ = hwnd
}

// [S-sig 0x140912ca0] 加锁读取 hwnd 字段并返回。
func (w *oledBlackoutNativeOverlayWindow) handle() uintptr {
	_ = w
	return 0
}

// [S-sig 0x140912dc0] 加锁写入 threadID 字段。无返回值。
func (w *oledBlackoutNativeOverlayWindow) setThreadID(id uintptr) {
	_ = w
	_ = id
}

// [S-sig 0x140912e60] 加锁读取 bounds 字段，以 4 整型返回 application.Rect。
func (w *oledBlackoutNativeOverlayWindow) Bounds() application.Rect {
	_ = w
	return application.Rect{}
}
