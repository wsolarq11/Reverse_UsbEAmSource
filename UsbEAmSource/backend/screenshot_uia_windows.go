// AUTO-RECONSTRUCTED — DOMAIN: screenshot UIA / MSAA control-hit-testing
// Source: UsbEAm_Launcher 1.0.3 (Go 1.25.12, PE64), disassembled from
// main.screenshotUIA* / main.screenshotMSAA* / main.screenshotAccessible* symbols.
// Tier markers:
//
//	[S VA]      — body fully translated from asm (pure geometry / math).
//	[S-sig VA]  — signature proven from asm; body is a faithful zero skeleton
//	              (COM / IUIAutomation / IAccessible vtable 调用未还原).
//
// 研究用途
package main

import (
	"image"
	"time"
	"unsafe"
)

// [P] 序言 eax=x、ebx=y（dword）+ rcx..r11 七 qword + 4 栈槽，远超当前 4 形参；中间参数
// （session 透传等）语义未定。保持 [P]。
func screenshotControlBoundsAtPointThroughOverlaySessionWithTimeout(x, y int, session *screenshotWindowSelectionSession, timeout time.Duration) image.Rectangle {
	return image.Rectangle{}
}

// [P] 序言 eax=x、ebx=y（dword）+ rcx..r11 七 qword + 栈 byte，远超当前 5 形参；中间参数
// 语义未定。保持 [P]。
func screenshotControlBoundsAtPointWithWindowFallbackWithTimeoutAndPriority(x, y int, hwnd uintptr, timeout time.Duration, priority int) image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x1409aed00] 序言 test rsi（hwnd）后 rax/rbx/rcx/rdi=rect（4 int），并透传
// screenshotWindowBounds + screenshotRectMatchesBounds；返回 bool。体未还原。
func screenshotControlRectMatchesTargetWindow(rect image.Rectangle, hwnd uintptr) bool {
	return false
}

// [S 汇编 0x1409aedc0, 96B]：a/b 两 image.Rectangle。a 空或 b 空 → false；面积（Dx*Dy）
// <=0 → false；返回 bArea < aArea（偏好面积更小的 b）。
func screenshotPreferSmallerControlRect(a, b image.Rectangle) bool {
	if a.Empty() || b.Empty() {
		return false
	}
	aArea := a.Dx() * a.Dy()
	bArea := b.Dx() * b.Dy()
	if aArea <= 0 || bArea <= 0 {
		return false
	}
	return bArea < aArea
}

// [P] 序言 eax=x、ebx=y（dword）+ rcx..r10 六 qword + r11b（byte），远超当前 5 形参；中间
// 参数语义未定。保持 [P]。
func screenshotFallbackControlBoundsAtPointWithTimeoutAndPriority(x, y int, hwnd uintptr, timeout time.Duration, priority int) image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x1409af060] 序言 test rax（hwnd）后透传 LazyProc.Call（GetWindowRect）；尾迹
// 返回 4 int（image.Rectangle），失败返回空矩形。体未还原。
func screenshotOverlayWindowRect(hwnd uintptr) image.Rectangle { return image.Rectangle{} }

// [S-sig 0x1409af160] 汇编实证：ctrlRect/winBounds(image.Rectangle 各 4 寄存器)+x/y(int 栈)+
// ctrlFound/winFound(bool 栈) 参数；返回 bool。逻辑：winFound→true、ctrlFound→false、
// 否则 screenshotRectMatchesBounds(ctrlRect,winBounds) 或 screenshotPointNearWindowEdge(x,y,winBounds,10)。
func shouldPreferWindowOverControlAtPoint(ctrlRect, winBounds image.Rectangle, x, y int, ctrlFound, winFound bool) bool { return false }

// [S 汇编 0x1409af2a0, 128B]：x/y 为点、rect 为矩形、threshold 为阈值（<0 钳 0）。rect 空或
// 点不在 rect 内 → false；否则返回点到 rect 边界最近距离 <= threshold。
func screenshotPointNearWindowEdge(x, y int, rect image.Rectangle, threshold int) bool {
	if rect.Empty() {
		return false
	}
	if x < rect.Min.X || x >= rect.Max.X || y < rect.Min.Y || y >= rect.Max.Y {
		return false
	}
	if threshold < 0 {
		threshold = 0
	}
	dx := x - rect.Min.X
	if right := rect.Max.X - x; right < dx {
		dx = right
	}
	dy := y - rect.Min.Y
	if bottom := rect.Max.Y - y; bottom < dy {
		dy = bottom
	}
	if dy < dx {
		dx = dy
	}
	return threshold >= dx
}

// [S 汇编 0x1409af320, 224B]：rect/bounds 两 image.Rectangle。rect 空或 bounds 空 → false；
// rect==bounds → true；rect 包含 bounds（Min<= 且 Max>=）→ true；否则四边 |差值|<=2 → true，
// 任一边 |差值|>2 → false。
func screenshotRectMatchesBounds(rect image.Rectangle, bounds image.Rectangle) bool {
	if rect.Empty() || bounds.Empty() {
		return false
	}
	if rect == bounds {
		return true
	}
	if rect.Min.X <= bounds.Min.X && rect.Max.X >= bounds.Max.X &&
		rect.Min.Y <= bounds.Min.Y && rect.Max.Y >= bounds.Max.Y {
		return true
	}
	if dx := rect.Min.X - bounds.Min.X; dx > 2 || dx < -2 {
		return false
	}
	if dy := rect.Min.Y - bounds.Min.Y; dy > 2 || dy < -2 {
		return false
	}
	if dx := rect.Max.X - bounds.Max.X; dx > 2 || dx < -2 {
		return false
	}
	if dy := rect.Max.Y - bounds.Max.Y; dy > 2 || dy < -2 {
		return false
	}
	return true
}

// [S 汇编 0x1409af400, 224B]：rect/bounds 两 image.Rectangle。rect 空或 bounds 空 →
// (空,false)；否则 rect.Intersect(bounds)，交集 Dx()<2 或 Dy()<2 → (空,false)；否则
// (交集,true)。
func normalizeScreenshotControlRect(rect, bounds image.Rectangle) (image.Rectangle, bool) {
	if rect.Empty() || bounds.Empty() {
		return image.Rectangle{}, false
	}
	r := rect.Intersect(bounds)
	if r.Dx() < 2 || r.Dy() < 2 {
		return image.Rectangle{}, false
	}
	return r, true
}

// [S 汇编 0x1409af4e0, 288B]：rect/controlType/bounds。normalize(rect,bounds) 失败 → false；
// controlType 不在 [50008,50033] → true；容器类 role（List/Menu/MenuBar/StatusBar/Tab/
// ToolBar/Tree/Group/DataGrid/Document/Window/Pane）→ 返回 !screenshotRectMatchesBounds(交集,
// bounds)（占满窗口边界则不采用）；其余 → true。switch 跳转表 0x1411e1ea0 26 项逐项解码。
func screenshotUIAControlRectUsable(rect image.Rectangle, controlType int, bounds image.Rectangle) bool {
	r, ok := normalizeScreenshotControlRect(rect, bounds)
	if !ok {
		return false
	}
	switch controlType {
	case 50008, 50009, 50010, 50017, 50018, 50021, 50023, 50026, 50028, 50030, 50032, 50033:
		return !screenshotRectMatchesBounds(r, bounds)
	default:
		return true
	}
}

// [S-sig 0x1409af600] 序言 eax=x、ebx=y、rcx=hwnd、rdi/rsi/r8/r9=fallback（4 int，透传
// normalizeScreenshotControlRect 作 bounds）；尾迹返回 normalize 结果 (rect, bool)。体未还原。
func screenshotWin32ChildControlBoundsAtPoint(x, y int, hwnd uintptr, fallback image.Rectangle) (image.Rectangle, bool) {
	return image.Rectangle{}, false
}

// [S-sig 0x1409af740] 序言 test rax（hwnd）后 ebx=x、ecx=y（dword 保存），透传
// ChildWindowFromPointEx；返回 uintptr。故 (hwnd, x, y) 顺序。体未还原。
func screenshotDeepestChildWindowAtPoint(hwnd uintptr, x, y int) uintptr { return 0 }

// [S-sig 0x1409af820] 序言 test rax（hwnd）后 ebx=x、ecx=y（打包 POINT），透传
// LazyProc.Call；返回 uintptr。体未还原。
func screenshotChildWindowFromClientPoint(hwnd uintptr, x, y int) uintptr { return 0 }

// [S-sig 0x1409af8c0] 序言 eax=x、ebx=y（dword）、rcx=hwnd、rdi/rsi/r8/r9=bounds（4 int，透传
// normalizeScreenshotControlRect）、r10=timeout、r11b=priority（byte）；尾迹两路返回 (rect, bool)。
// 体未还原。
func screenshotMSAAControlBoundsAtPointWithPriority(x, y int, hwnd uintptr, bounds image.Rectangle, timeout time.Duration, priority uint8) (image.Rectangle, bool) {
	return image.Rectangle{}, false
}

// [S-sig 0x1409af9e0] 序言 eax=x、ebx=y（dword）、rcx=hwnd、rdi=timeout（test+cmovle 默认
// 120ms）、sil=priority（byte）；调 Query 后尾迹返回 6 寄存器（rect 4 + error 2，丢弃
// controlType）。体未还原。
func screenshotMSAAElementBoundsAtPointWithQueryTimeoutAndPriority(x, y int, hwnd uintptr, timeout time.Duration, priority uint8) (image.Rectangle, error) {
	return image.Rectangle{}, nil
}

// [S-sig 0x1409afb20] 序言 eax=x、ebx=y、rcx=hwnd；test rcx 后分支透传 FromWindow/Global，
// 尾迹返回 (rect, error)。体未还原。
func screenshotMSAAElementBoundsAtPointOnCOMThread(x, y int, hwnd uintptr) (image.Rectangle, error) {
	return image.Rectangle{}, nil
}

// [S-sig 0x1409afba0] 序言 eax=x、ebx=y、rcx=hwnd（透传 AccessibleObjectFromWindow）；尾迹
// 返回 6 寄存器 (rect, error)。体未还原。
func screenshotMSAAElementBoundsAtPointFromWindow(x, y int, hwnd uintptr) (image.Rectangle, error) {
	return image.Rectangle{}, nil
}

// [S-sig 0x1409affa0] 序言 eax=x、ebx=y、rcx=hwnd（test 判空决定祖先过滤）；尾迹返回
// (rect, error)。体未还原。
func screenshotMSAAElementBoundsAtPointGlobal(x, y int, hwnd uintptr) (image.Rectangle, error) {
	return image.Rectangle{}, nil
}

// [S-sig 0x1409b0520] 序言 test rax（hwnd）后构造对象；尾迹两路：成功 (对象, nil error)、
// 失败构造 errorString（len 0x17）+ lea itab，故返回 (*screenshotAccessible, error)。体未还原。
func screenshotAccessibleObjectFromWindow(hwnd uintptr) (*screenshotAccessible, error) {
	return nil, nil
}

// [P] 序言 rax=acc、ebx=x、ecx=y（dword）；尾迹返回 VARIANT 结构（4 word + qword）与 error
// 组合，当前 (bool, rect) 与之不符。保持 [P]。
func screenshotAccessibleHitTest(acc *screenshotAccessible, x, y int) (bool, image.Rectangle) {
	return false, image.Rectangle{}
}

// [S-sig 0x1409b09c0] 无参；体内 xor eax(al=0) 后调用 initializeScreenshotCOMThreadMode(false)
// 并丢弃 (func(),error) 返回。无返回值。体未还原。
func initializeScreenshotCOMSTAWorkerThread() {}

// [S-sig 0x1409b0a00] COM 线程模式初始化（APARTMENTTHREADED|DISABLE_OLE1DDE）。
// 汇编实证：morestack 仅保存 al（单 bool 形参 allowChangedMode）；LockOSThread 后
// CoInitializeEx(NULL,6) 经 LazyProc.Call；HRESULT>=0(S_OK/S_FALSE) → 构造 func2 闭包
// {func2, success bool}（func2=CoUninitialize 清理，读 [rdx+8] bool 决定是否反初始化）
// 返回 (func,nil)；RPC_E_CHANGED_MODE(0x80010106) 且 allowChangedMode → 同成功；
// 否则 fmt.Errorf(HRESULT) 返回 (nil,err)。
func initializeScreenshotCOMThreadMode(allowChangedMode bool) (func(), error) {
	_ = allowChangedMode
	return nil, nil
}

// [S-sig 0x1409b0be0] 序言 test rax（acc）后透传 LazyProc.Call（GetWindowFromAccessible）；
// 尾迹 xor eax/xor ebx（(0,false)）与 mov ebx,1（(hwnd,true)），故返回 (uintptr, bool)。体未还原。
func screenshotWindowFromAccessible(acc *screenshotAccessible) (uintptr, bool) { return 0, false }

// [S-sig 0x1409b0ca0] 取可访问元素位置。汇编实证：rax=acc + bx/cx/di/si 四 16 位（VARIANT
// VT/Reserved1..3）+ r8=qword（Val int64），即值传 screenshotOleVariant（5 寄存器字段展开）；
// 返回 6 寄存器 (image.Rectangle 4 int, error 2 word)，错误路径 rect 清零 + itab/data。
func screenshotAccessibleLocation(acc *screenshotAccessible, v screenshotOleVariant) (image.Rectangle, error) {
	_, _ = acc, v
	return image.Rectangle{}, nil
}

// [S-sig 0x1409b0f80] 清空 VARIANT 包装（screenshotOleVariant）。无返回值。
func clearScreenshotOleVariant(v *screenshotOleVariant) {}

// [S-sig 0x1409b1000] 序言 eax=x、ebx=y、rcx=timeout（cmovle 默认 120ms）、dil=priority
// （byte，xor edi 置 hwnd=0）；调 Query 尾迹返回 7 寄存器 (rect, controlType, error)。体未还原。
func screenshotUIAElementBoundsAtPointWithQueryTimeoutAndPriority(x, y int, timeout time.Duration, priority uint8) (image.Rectangle, uint32, error) {
	return image.Rectangle{}, 0, nil
}

// [S-sig 0x1409b1140] 序言 eax=x、ebx=y（无 hwnd）；createScreenshotUIAutomationOnCOMThread
// 后透传 WithAutomation，尾迹返回 7 寄存器 (rect, controlType, error)。体未还原。
func screenshotUIAElementBoundsAtPointOnCOMThread(x, y int) (image.Rectangle, uint32, error) {
	return image.Rectangle{}, 0, nil
}

// [S-sig 0x1409b1320] 序言 eax=x、ebx=y、rcx=automation（参数顺序与旧注释相反）；尾迹返回
// 7 寄存器 (rect, controlType, error)。体未还原。
func screenshotUIAElementBoundsAtPointWithAutomation(x, y int, automation *screenshotUIAutomation) (image.Rectangle, uint32, error) {
	return image.Rectangle{}, 0, nil
}

// [S-sig 0x1409b1b40] 序言 test rax（automation）后透传 SyscallN（get_RawViewWalker）；返回
// unsafe.Pointer（失败 nil）。体未还原。
func screenshotUIARawViewWalker(automation *screenshotUIAutomation) unsafe.Pointer { return nil }

// [S-sig 0x1409b1ca0] 序言 test rax（element）后透传 SyscallN（get_CurrentBoundingRectangle）；
// 尾迹返回 8 寄存器 (rect, controlType, ok, error)。体未还原。
func readScreenshotUIAElementHitInfo(element unsafe.Pointer) (image.Rectangle, uint32, bool, error) {
	return image.Rectangle{}, 0, false, nil
}

// [P] 序言 rax=walker、rbx=element、ecx=x、edi=y、rsi/r8 额外 2 槽 + 栈大结构；返回 8 寄存器，
// 超出当前 4 形参/3 返回。保持 [P]。
func screenshotUIADeepestElementHitInfoAtPoint(walker, element unsafe.Pointer, x, y int) (unsafe.Pointer, image.Rectangle, bool) {
	return nil, image.Rectangle{}, false
}

// [S-sig 0x1409b2540] 序言 rax=walker（解引用 [rax] 取 vtable，cmp [rdx+0x20]=
// GetFirstChildElement）、rbx=element（test 非空）；尾迹成功 (element,nil) / 失败构造
// errorString；故 (walker, element unsafe.Pointer) (unsafe.Pointer, error)。体未还原。
func screenshotUIAFirstChildElement(walker, element unsafe.Pointer) (unsafe.Pointer, error) {
	return nil, nil
}

// [S-sig 0x1409b26a0] 序言同 FirstChild（cmp [rdx+0x30]=GetNextSiblingElement）；返回
// (unsafe.Pointer, error)。体未还原。
func screenshotUIANextSiblingElement(walker, element unsafe.Pointer) (unsafe.Pointer, error) {
	return nil, nil
}

// [S-sig 0x1409b2800] 序言同 FirstChild（cmp [rdx+0x18]=GetParentElement）；返回
// (unsafe.Pointer, error)。体未还原。
func screenshotUIAParentElement(walker, element unsafe.Pointer) (unsafe.Pointer, error) {
	return nil, nil
}

// [P] 序言 rax=walker、rbx=element、ecx=x、edi=y（4 槽）；返回 8 寄存器（element+rect+int+
// bool+int），超出当前 3 返回。保持 [P]。
func screenshotUIASelectableAncestorHitInfoAtPoint(walker, element unsafe.Pointer, x, y int) (unsafe.Pointer, image.Rectangle, bool) {
	return nil, image.Rectangle{}, false
}

// [S 汇编 0x1409b2c60, 128B]：8 槽签名——rect（4 int）+ controlType（uint32）+ ok（bool）+
// x + y。esi/r8b 经调用点 0x1409b2aaa/0x1409b2ab1 与 readScreenshotUIAElementHitInfo 返回
// (rect,uint32,bool,error) 交叉确证。体：rect.Dx()<2 || rect.Dy()<2 → false；否则
// rect.Min.X<=x && x<rect.Max.X && rect.Min.Y<=y && y<rect.Max.Y（controlType/ok 未用）。
func screenshotUIAElementHitContainsPoint(rect image.Rectangle, controlType uint32, ok bool, x, y int) bool {
	if rect.Dx() < 2 || rect.Dy() < 2 {
		return false
	}
	return rect.Min.X <= x && x < rect.Max.X && rect.Min.Y <= y && y < rect.Max.Y
}

// [S-sig 0x1409b2ce0] 汇编实证：a(image.Rectangle:rax/rbx/rcx/rdi)+aControlType(uint32:esi)+
// aOK(bool:r8b)+b(image.Rectangle+bControlType+bOK 栈) 参数；返回 bool。逻辑：controlType 分类
// （0xc358 基址跳转表）+ rect 比较，未逐条翻译。
func screenshotPreferUIAElementHitInfo(a image.Rectangle, aControlType uint32, aOK bool, b image.Rectangle, bControlType uint32, bOK bool) bool { return false }

// [S-sig 0x1409b3080] 在 COM 线程上创建 IUIAutomation 包装（CoCreateInstance）。返回
// *screenshotUIAutomation。体未还原。
func createScreenshotUIAutomationOnCOMThread() *screenshotUIAutomation { return nil }

// [S-sig 0x1409b31c0] 释放 IUIAutomation 包装（rax 为指针，nil 直返）。
func releaseScreenshotUIAutomation(p *screenshotUIAutomation) {}

// [S-sig 0x1409b3240] 释放 UIA 元素（无对应落地类型，以 unsafe.Pointer 承载）。
func releaseScreenshotUIAElement(p unsafe.Pointer) {}

// [S-sig 0x1409b32c0] 释放 IAccessible 包装（rax 为指针，nil 直返）。
func releaseScreenshotAccessible(p *screenshotAccessible) {}

// [S-sig 0x1409b3340] 释放 UIA 树遍历器（无对应落地类型，以 unsafe.Pointer 承载）。
func releaseScreenshotUIATreeWalker(p unsafe.Pointer) {}
