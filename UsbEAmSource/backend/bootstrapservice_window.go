// AUTO-RECONSTRUCTED — DOMAIN: bootstrap launcher window sizing（窗口尺寸域）
// 研究用途
//
// 契约来源：capstone 反汇编符号标注（docs/goresym/pipeline/tmp/）
//   - PreviewLauncherUIScale.asm.txt（0x14077a900, 256B）
//
// 还原口径：可编译 + 功能一致（非字节级同哈希）。档位如实标注：
//
//	[S] 汇编实证（函数体逐条对位）
//	[P] 骨架/占位（签名已对齐，体待续作）
package main

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

// clampLauncherUIScalePercent 把 UI 缩放百分比钳制到 [100, 300]。
// [S 汇编实证 0x14077a925-0x14077a951]：
//
//	test rcx,rcx / jg      → rcx <= 0    → ecx = 0x64 (100)
//	cmp  rcx,0x64 / jge    → 0 < rcx<100 → ecx = 0x64 (100)
//	cmp  rcx,0x12c / jle   → rcx > 300   → ecx = 0x12c (300)
//
// 汇编全程为整数比较（test/cmp/mov ecx,imm）且无任何 XMM 参与，
// 故调用方 PreviewLauncherUIScale 的入参是整数百分比，而非源码旧注的 float64。
func clampLauncherUIScalePercent(percent int) int {
	if percent < 100 {
		return 100
	}
	if percent > 300 {
		return 300
	}
	return percent
}

// resolveAttachedLauncherWindow 取得当前挂接的启动器窗口。
// [S 汇编 0x14079bf40, 55 行]：
//
//	0x14079bf6f  lock cmpxchg（bs.lock@+0x540 快路径）
//	0x14079bfbb  rax=bs.launcherWindow(+0x190), rcx=bs.launcherWindow(+0x198)  → 读窗口接口
//	0x14079bfd3  byte [rsp+0xf]=0 → defer 退出后返 (rax, rbx)
//	0x14079bfe0  call rax（defer 展开触发 runtime.deferreturn）
func (bs *BootstrapService) resolveAttachedLauncherWindow() application.Window {
	bs.lock.Lock()
	defer bs.lock.Unlock()
	return bs.launcherWindow
}

// applyLauncherWindowSizing 对窗口套用尺寸/缩放。
// [S-sig] 骨架。VA 0x14079c8a0 / 736B（va_map_fixed2.txt:236）。
// 调用点实证见 PreviewLauncherUIScale.asm 0x14077a96a-0x14077a980 的寄存器装配：
//
//	rax = bs            (rsp+0x48 回填)
//	rbx = 断言后 itab   rcx = 断言后 data       → 第 2 形参为窗口接口
//	rdi = 钳制后百分比                          → 第 3 形参 int
//	esi = 1                                     → 第 4 形参 bool
func (bs *BootstrapService) applyLauncherWindowSizing(window application.Window, percent int, preview bool) {
	if bs == nil {
		return
	}
	_ = window
	_ = clampLauncherUIScalePercent(percent)
	_ = preview
}

// showPendingLauncherReveal 检查并执行待处理的启动器揭示。
// [S 汇编 0x140796120, 76 行] 实证流程：
//
//	window==nil → return
//	lock(+0x540) → 读 pendingLauncherReveal(+0x44a) → 清 false →
//	unlock → 若原 pendingLauncherReveal 为 true → showLauncherWindow(bs, window, false, true)
//
// 汇编实证：rax=bs, rbx=window.TYPE, rcx=window.DATA
func (bs *BootstrapService) showPendingLauncherReveal(window interface{}) {
	if window == nil {
		return
	}
	bs.lock.Lock()
	reveal := bs.pendingLauncherReveal
	bs.pendingLauncherReveal = false
	bs.lock.Unlock()
	if reveal {
		bs.showLauncherWindow(window, false, true)
	}
}

// launcherScreenshotCaptureState 记录启动器窗口截屏前的原始状态，供截屏结束后恢复。
// 5 字宽（window 接口 2 字 + enabled bool 1 字 + x,y int 各 1 字），
// 与 beginLauncherScreenshotCapture 返回寄存器 (rax,rbx,rcx,rdi,rsi) 逐字对齐。
type launcherScreenshotCaptureState struct {
	window  launcherWindowController
	enabled bool
	x       int
	y       int
}

// launcherWindowScreenshotHideOffset 截屏时把启动器窗口左移的像素数（移出屏幕）。
// 汇编 `add rbx,-0x186a0`（0x186a0=100000）实证为字面量。
const launcherWindowScreenshotHideOffset = 100000

// beginLauncherScreenshotCapture 开始截屏：把启动器窗口移出屏幕并开启内容保护。
// [S 汇编 0x140796aa0, 108 行]：
//
//	if !enabled || win==nil → 返零态
//	if !win.IsVisible()    → 返零态
//	x,y := win.Position()
//	win.SetContentProtection(true)
//	win.SetPosition(x-100000, y)
//	return {window:win, enabled:true, x:x, y:y}
//
// 汇编实证：itab 方法序 fun[6]=IsVisible / fun[7]=Position /
// fun[9]=SetContentProtection / fun[10]=SetPosition（launcherWindowController，
// types_launcher.go:570）。SetPosition 首参减 100000（`add rbx,-0x186a0`）。
func (bs *BootstrapService) beginLauncherScreenshotCapture(win launcherWindowController, enabled bool) launcherScreenshotCaptureState {
	if !enabled || win == nil {
		return launcherScreenshotCaptureState{}
	}
	if !win.IsVisible() {
		return launcherScreenshotCaptureState{}
	}
	x, y := win.Position()
	win.SetContentProtection(true)
	win.SetPosition(x-launcherWindowScreenshotHideOffset, y)
	return launcherScreenshotCaptureState{window: win, enabled: true, x: x, y: y}
}

// endLauncherScreenshotCapture 结束截屏：把启动器窗口恢复到原位置并关闭内容保护。
// [S 汇编 0x140796c20, 68 行]：
//
//	if !enabled || win==nil → return
//	win.SetPosition(x, y)
//	win.SetContentProtection(false)
//
// 汇编实证：fun[10]=SetPosition(x,y) 先于 fun[9]=SetContentProtection(false)。
func (bs *BootstrapService) endLauncherScreenshotCapture(win launcherWindowController, enabled bool, x, y int) {
	if !enabled || win == nil {
		return
	}
	win.SetPosition(x, y)
	win.SetContentProtection(false)
}

// HideLauncherWindow 隐藏启动器窗口。
// [S-sig 0x14079a220, 192B]：resolveLauncherWindow 取窗口（nil 则返回）；经 itab 查找
// hideLauncherWindowToTray 方法调用。体待窗口隐藏链专项还原。
func (bs *BootstrapService) HideLauncherWindow() {
	_ = bs
}

// HideScreenshotPreview 隐藏截图预览窗口。
// [S-sig 0x14079be80, 192B]：mu(+0x540) 保护读 screenshotPreview(+0x408)，非空则
// screenshotPreviewWindowService.Hide。体待字段名精确对位专项还原。
func (bs *BootstrapService) HideScreenshotPreview() {
	_ = bs
}

// applyLauncherWindowSizingForShow 显示前应用启动器窗口尺寸。
// [S-sig 0x1407968c0, 160B]：window nil 返回；consumeLauncherDefaultSizeReset → bool、
// loadLauncherUIScalePercent → int，交 applyLauncherWindowSizing(window,percent,bool)。
// 体待两个依赖函数专项还原。
func (bs *BootstrapService) applyLauncherWindowSizingForShow(window application.Window) {
	_ = window
}

// consumeLauncherDefaultSizeReset 消费「下次显示重置尺寸」标志（读旧值并清空）。
// [S 汇编 0x140796960, 224B]：mu(+0x540).Lock + defer Unlock → 读 resetLauncherSizeOnNextShow(+0x448)
// 置 false，返回旧值。
func (bs *BootstrapService) consumeLauncherDefaultSizeReset() bool {
	bs.lock.Lock()
	defer bs.lock.Unlock()
	old := bs.resetLauncherSizeOnNextShow
	bs.resetLauncherSizeOnNextShow = false
	return old
}

// clearLauncherVerticalMaximizeSnapshot 清空启动器垂直最大化快照。
// [S 汇编 0x14079b8c0, 224B]：mu(+0x540).Lock + defer Unlock → verticalMaximizeSnapshot(+0x500) 置 nil。
func (bs *BootstrapService) clearLauncherVerticalMaximizeSnapshot() {
	bs.lock.Lock()
	defer bs.lock.Unlock()
	bs.verticalMaximizeSnapshot = nil
}

// SetLauncherSidebarPinnedOpen 设置启动器侧栏固定打开。
// [S-sig 0x14079aac0, 224B]：resolveLauncherWindow → 非空则 typeAssert 后
// applyLauncherSidebarPinnedWindowSizing(window, pinned)。体待该依赖专项还原。
func (bs *BootstrapService) SetLauncherSidebarPinnedOpen(pinned bool) {
	w, _ := bs.resolveLauncherWindow(nil, nil)
	_ = w
	_ = pinned
}

// shouldCloseLauncherWindow 判定是否应关闭启动器窗口。
// [S 汇编 0x14079a2e0, 224B]：mu(+0x540).Lock + defer Unlock → 读 allowLauncherWindowClose(+0x44b)。
func (bs *BootstrapService) shouldCloseLauncherWindow() bool {
	bs.lock.Lock()
	defer bs.lock.Unlock()
	return bs.allowLauncherWindowClose
}

// registerLauncherCloseToTrayHook 注册启动器关闭到托盘钩子。
// [S-sig 0x1408d0500, 224B]：nil 返回；newobject 闭包捕获 (回调, 参数, bs) → 间接调用
// bs 钩子注册函数（+0x160）。体待钩子注册域专项还原。
func registerLauncherCloseToTrayHook(bs *BootstrapService, callback, arg interface{}) {
	_, _, _ = bs, callback, arg
}

// resolveLauncherWindowRelativePosition 计算启动器窗口相对位置（居中或钳位）。
// [S 汇编 0x1408d0a80, 80B]：6 int + 1 bool 参数，返回 (x, y)。
// center=true：x=max((screenW-windowW)/2,0)、y=max((screenH-windowH)/2,0)（算术右移向下取整）；
// center=false：x=clamp(screenW-windowW,0,maxX)、y=clamp(screenH-windowH,0,maxY)，
// 其中 maxX/maxY 为负时对应分量置 0，screenW/screenH<=0 时直接取 maxX/maxY。
func resolveLauncherWindowRelativePosition(screenW, screenH, windowW, windowH, maxX, maxY int, center bool) (int, int) {
	if center {
		x := 0
		if screenW > 0 {
			x = (screenW - windowW) >> 1
			if x < 0 {
				x = 0
			}
		}
		y := 0
		if screenH > 0 {
			y = (screenH - windowH) >> 1
			if y < 0 {
				y = 0
			}
		}
		return x, y
	}
	var x int
	if screenW > 0 {
		x = screenW - windowW
		if x < 0 {
			x = 0
		}
		if maxX < 0 {
			x = 0
		} else if maxX <= x {
			x = maxX
		}
	} else {
		x = maxX
	}
	var y int
	if screenH > 0 {
		y = screenH - windowH
		if y < 0 {
			y = 0
		}
		if maxY < 0 {
			y = 0
		} else if maxY <= y {
			y = maxY
		}
	} else {
		y = maxY
	}
	return x, y
}

// revealLauncherWindowForQRCodeDecode 为二维码解码显示启动器窗口。
// [S-sig 0x14079a160, 192B]：接口两方法判定（+0x48/+0x38）→ 满足则 showLauncherWindow(false,true)。
// 体待二维码解码域专项还原。
func (bs *BootstrapService) revealLauncherWindowForQRCodeDecode(a, b interface{}) {
	_, _ = a, b
}

// ApplyLauncherWindowLayoutAction 应用启动器窗口布局动作。
// [S-sig 0x14079a9e0, 224B]：resolveLauncherWindow → 类型断言查找 action 处理器 →
// applyLauncherWindowLayoutAction。体待窗口布局域专项还原。
func (bs *BootstrapService) ApplyLauncherWindowLayoutAction(a, b interface{}) {
	_, _ = a, b
}
