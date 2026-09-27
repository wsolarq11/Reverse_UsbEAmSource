// AUTO-RECONSTRUCTED — DOMAIN: oledblackout overlay (wails window 侧)
// Source: UsbEAm_Launcher 1.0.3 (Go 1.25.12, PE64), disassembled from
// main.oledBlackoutOverlayWindow.* symbols. Tier markers:
//
//	[S VA]      — body fully translated from asm (pure geometry / math).
//	[S-sig VA]  — signature proven from asm; body is a faithful zero skeleton.
//	[P]         — signature not yet proven; zero skeleton + blocking reason.
//
// receiver *oledBlackoutOverlayWindow 已在 types_oled.go 落地。
// 研究用途
package main

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

// [S-sig 0x140910ae0] 从屏幕的物理边界计算遮罩矩形。
// asm：receiver=RAX，arg=RBX(*application.Screen)，读 +0x68/+0x70/+0x78/+0x80
// (=PhysicalBounds.X/Y/Width/Height) 并以 4 整型返回 application.Rect。
func (w *oledBlackoutOverlayWindow) boundsForScreen(screen *application.Screen) application.Rect {
	_ = w
	_ = screen
	return application.Rect{}
}

// [S-sig 0x140910ba0] 写入 bounds 字段并（若已建原生窗口）转发 setBoundsOnThread。
// asm：receiver=RAX + 4 整型实参(bounds)，无返回值。
func (w *oledBlackoutOverlayWindow) SetBounds(bounds application.Rect) {
	_ = w
	_ = bounds
}

// [S-sig 0x140910ce0] 显示遮罩；优先原生窗口，否则走 wails 窗口。返回 error。
func (w *oledBlackoutOverlayWindow) Show() error {
	_ = w
	return nil
}

// [S-sig 0x140910e00] 隐藏遮罩。返回 error。
func (w *oledBlackoutOverlayWindow) Hide() error {
	_ = w
	return nil
}

// [S-sig 0x140910e80] 聚焦遮罩窗口。无返回值。
func (w *oledBlackoutOverlayWindow) Focus() {
	_ = w
}

// [S-sig 0x140910f00] 关闭遮罩窗口。返回 error。
func (w *oledBlackoutOverlayWindow) Close() error {
	_ = w
	return nil
}

// [S-sig 0x140910f80] 返回原生窗口句柄（无原生窗口时回退 wails 窗口句柄）。
func (w *oledBlackoutOverlayWindow) NativeWindowHandle() uintptr {
	_ = w
	return 0
}

// [S-sig 0x140911000] 从 service 解绑并释放资源。无返回值。
func (w *oledBlackoutOverlayWindow) releaseFromService() {
	_ = w
}
