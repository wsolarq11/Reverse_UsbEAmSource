// AUTO-RECONSTRUCTED — DOMAIN: qrcode_windows / qrcode
// Source: UsbEAm_Launcher 1.0.3 (Go 1.25.12, PE64), disassembled from
// main.qrcode_* / main.qrCodeScreenSelectionSession.* / main.BootstrapService.*
// symbols. Tier markers:
//
//	[S VA]      — body fully translated from asm (pure geometry / math).
//	[S-sig VA]  — signature proven from asm; body is a faithful zero skeleton
//	              (Win32 / GDI / message-loop / state-machine deps).
//	[P]         — signature not proven from asm; zero-value skeleton only.
//
// 研究用途
package main

import (
	"image"
	"unsafe"

	"golang.design/x/clipboard"
)

// ===== qrcode.go domain =====

// [S 汇编 0x140931960, 96B] initQRCodeClipboard 失败 → ("", err)；否则
// string(clipboard.Read(FmtText))，nil error。asm：eax=0(FmtText) → clipboard.Read
// 返回 []byte(3 寄存器) → slicebytetostring（cap 丢弃）→ (rax/rbx=string, rcx/rdi=nil error)。
func readQRCodeTextFromClipboard() (string, error) {
	if err := initQRCodeClipboard(); err != nil {
		return "", err
	}
	return string(clipboard.Read(clipboard.FmtText)), nil
}

// [S 汇编 0x1409319c0, 128B] initQRCodeClipboard 失败 → 透传 err；否则
// clipboard.Write(FmtText, []byte(text))（丢弃返回 channel），return nil。
func writeQRCodeTextToClipboard(text string) error {
	if err := initQRCodeClipboard(); err != nil {
		return err
	}
	clipboard.Write(clipboard.FmtText, []byte(text))
	return nil
}

// ===== qrcode_windows.go domain =====

// [S-sig 0x1409356c0] 会话主循环：阻塞运行直到接受/取消/超时，产出结果。
// 序言仅保存 receiver；返回单个大结构（栈返回槽）。体为 Win32 消息泵，未翻译。
func (s *qrCodeScreenSelectionSession) runResult() qrCodeNativeSelectionResult {
	return qrCodeNativeSelectionResult{}
}

// [S-sig 0x140936080] 判断遮罩窗口关闭后是否应最终化：finalizeRequest 非空、
// 未取消、未超时。无参，返回 bool(零值)。
func (s *qrCodeScreenSelectionSession) shouldFinalizeAfterOverlayClosed() bool {
	return false
}

// [S-sig 0x140936100] 启动 HDR 预览升级（状态位 cmpxchg 0x460）。无参无返回值。
func (s *qrCodeScreenSelectionSession) startHDRPreviewUpgrade() {
}

// [S-sig 0x140936480] 完成 HDR 预览升级（状态位 cmpxchg 释放）。无参无返回值。
func (s *qrCodeScreenSelectionSession) finishHDRPreviewUpgrade() {
}

// [S-sig 0x140936520] 设置待处理的 HDR 预览快照。序言保存 rbx/rcx/rdi/rsi/r8/r9
// 共 6 个 qword = qrCodeScreenSnapshot（image.Rectangle 4 int + 2 指针）。
func (s *qrCodeScreenSelectionSession) setPendingHDRPreview(p qrCodeScreenSnapshot) {
}

// [S-sig 0x140936700] 取出并清空待处理的 HDR 预览快照。无参，返回 *qrCodeScreenSnapshot。
func (s *qrCodeScreenSelectionSession) takePendingHDRPreview() *qrCodeScreenSnapshot {
	return nil
}

// [S-sig 0x140936860] 创建选择窗口（注册类 + CreateWindowEx 链）。无显式参数，
// 返回 error。
func (s *qrCodeScreenSelectionSession) createWindow() error {
	return nil
}

// [S-sig 0x140936b60] 关闭会话（隐藏工具栏 + 释放窗口句柄）。无参无返回值。
func (s *qrCodeScreenSelectionSession) close() {
}

// [S-sig 0x140936ca0] 设置窗口句柄。序言保存 rbx = hwnd uintptr。
func (s *qrCodeScreenSelectionSession) setWindowHandle(hwnd uintptr) {
}

// [S-sig 0x140936dc0] 返回当前窗口句柄（锁保护读取 hwnd 字段）。
func (s *qrCodeScreenSelectionSession) windowHandle() uintptr {
	return 0
}

// [S-sig 0x140936f00] 设置点击穿透。序言保存 bl = bool。
func (s *qrCodeScreenSelectionSession) setHitTestTransparent(v bool) {
}

// [S-sig 0x140937020] 返回是否点击穿透。无参，返回 bool。
func (s *qrCodeScreenSelectionSession) isHitTestTransparent() bool {
	return false
}

// [S-sig 0x140937160] 标记会话超时（置 timeoutTriggered）。无参无返回值。
func (s *qrCodeScreenSelectionSession) markTimedOut() {
}

// [S-sig 0x140937280] 返回是否已超时。无参，返回 bool。
func (s *qrCodeScreenSelectionSession) didTimeOut() bool {
	return false
}

// [S-sig 0x1409373c0] 显示选择窗口（ShowWindow 链）。无参无返回值。
func (s *qrCodeScreenSelectionSession) showSelectionWindow() {
}

// [S-sig 0x140937be0] 窗口消息分发。序言保存 rbx/rcx/rdi/rsi = (hwnd,msg,wparam,lparam)，
// msg 在 ecx 与 0x113(WM_TIMER)/0x20 比较；返回 uintptr(LRESULT)。
func (s *qrCodeScreenSelectionSession) handleMessage(hwnd uintptr, msg uint32, wparam, lparam uintptr) uintptr {
	return 0
}

// [S-sig 0x140939100] 开始右键动作（依据字段状态分派 applyRightButtonAction）。无参无返回值。
func (s *qrCodeScreenSelectionSession) beginRightButtonAction() {
}

// [S-sig 0x1409391c0] 应用右键动作（取消/抑制上下文菜单）。无参无返回值。
func (s *qrCodeScreenSelectionSession) applyRightButtonAction() {
}

// [S-sig 0x140939280] 轮询右键物理状态（GetAsyncKeyState 链）。无参无返回值。
func (s *qrCodeScreenSelectionSession) pollRightButtonState() {
}

// [S-sig 0x140939320] 更新右键物理状态。序言 test cl = down bool。
func (s *qrCodeScreenSelectionSession) updateRightButtonPhysicalState(down bool) {
}

// [S-sig 0x140939400] 取消当前选择（隐藏工具栏 + 关闭窗口）。无参无返回值。
func (s *qrCodeScreenSelectionSession) cancelSelection() {
}

// [S-sig 0x140939720] 取消并关闭会话。无参无返回值。
func (s *qrCodeScreenSelectionSession) cancelAndClose() {
}

// [S-sig 0x1409397c0] morestack 保存 rax(recv)/rbx；rbx 参数重载 rax 作 w32.InvalidateRect 的 hwnd；
// 无返回寄存器=void。
func (s *qrCodeScreenSelectionSession) handleHDRPreviewUpgrade(p uintptr) {
}

// [S-sig 0x140939a60] 将待处理 HDR 预览应用到最终图像。无显式参数，无返回值。
func (s *qrCodeScreenSelectionSession) applyPendingHDRPreviewForFinalImage() {
}

// [S-sig 0x140939be0] 确保最终图像具备 HDR 预览（字段状态判断 + 异步触发）。无参无返回值。
func (s *qrCodeScreenSelectionSession) ensureHDRPreviewForFinalImage() {
}

// [S-sig 0x140939d00] 返回 HDR 预览升级是否进行中。无参，返回 bool。
func (s *qrCodeScreenSelectionSession) isHDRPreviewUpgradeInProgress() bool {
	return false
}

// [S-sig 0x140939e40] 替换快照位图。序言保存 rbx/rcx = original/shaded 两个 *image.RGBA，
// 分别 call createQRCodeBitmapHandle。无返回值。
func (s *qrCodeScreenSelectionSession) replaceSnapshotBitmaps(original, shaded *image.RGBA) {
}

// [S-sig 0x14093a1e0] 释放快照位图（DeleteObject 链）。无参无返回值。
func (s *qrCodeScreenSelectionSession) releaseSnapshotBitmaps() {
}

// [S-sig 0x14093a400] 绘制窗口内容（WM_PAINT 处理）。序言保存 rbx = hdc uintptr；
// 返回 bool(是否处理)。
func (s *qrCodeScreenSelectionSession) paint(hdc uintptr) bool {
	return false
}

// [S-sig 0x14093afa0] 返回当前活动选区矩形（依据控制选中/拖拽/角半径状态）。
// 无参，返回 image.Rectangle。
func (s *qrCodeScreenSelectionSession) activeSelectionRect() image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14093b0a0] 返回某点是否命中选区缩放手柄。序言保存 rbx/rcx = point；
// 返回 bool（调用点 test al）。
func (s *qrCodeScreenSelectionSession) selectionResizeHandleAt(p image.Point) bool {
	return false
}

// [S-sig 0x14093b180] 开始选区缩放。序言保存 rax(receiver) + rbx(handle,1字) + rcx/rdi(point 2字)；
// call selectionResizeHandleAt(p) 判命中；尾声 mov eax,1 / xor eax = 返回 bool。体骨架。
func (s *qrCodeScreenSelectionSession) beginSelectionResize(handle uint8, p image.Point) bool {
	return false
}

// [S-sig 0x14093b3c0] 更新选区缩放（依据 selectionResizeActive 状态）。无显式参数，
// 无返回值。
func (s *qrCodeScreenSelectionSession) updateSelectionResize() {
}

// [S-sig 0x14093b760] 完成选区缩放（清 resize 状态 + invalidate）。序言保存 rdi/rcx = point。
func (s *qrCodeScreenSelectionSession) finishSelectionResize(p image.Point) {
}

// [S-sig 0x14093ba20] morestack 保存 rax(recv)/rbx/rcx/rdi；rbx 透传 invalidateCornerRadiusControl 的 hwnd；
// rcx/rdi = point（透传 selectionResizeHandleAt）；无返回寄存器=void。
func (s *qrCodeScreenSelectionSession) updateSelectionResizeHover(hwnd uintptr, p image.Point) {
}

// [S-sig 0x14093bba0] morestack 保存 rax(recv)/rbx/rcx/rdi；rbx 装箱后进 syscall.LazyProc.Call（hwnd）；
// rcx/rdi = point（X/Y 与 slider rect 比较）；尾声 mov eax,1 / xor eax = bool。
func (s *qrCodeScreenSelectionSession) beginCornerRadiusDrag(hwnd uintptr, p image.Point) bool {
	return false
}

// [S-sig 0x14093bcc0] 更新圆角半径拖拽。morestack 保存 rax(recv)/rbx/rcx/rdi；rbx=hwnd，
// rcx/rdi=point(X/Y)；switch 表按状态选 WM 消息 ID(0x7f82..0x7f89) 送 LazyProc.Call；
// 两个返回路径均 xor eax 零值、无 mov eax,1，返回 void。
func (s *qrCodeScreenSelectionSession) updateCornerRadiusDrag(hwnd uintptr, p image.Point) {
	_, _ = hwnd, p
}

// [S-sig 0x14093bf80] 完成圆角半径拖拽（清 cornerRadiusDragging + invalidate）。
// 序言保存 rdi/rcx = point。
func (s *qrCodeScreenSelectionSession) finishCornerRadiusDrag(p image.Point) {
}

// [S-sig 0x14093c080] 返回圆角半径滑块命中矩形。无参，返回 image.Rectangle。
func (s *qrCodeScreenSelectionSession) cornerRadiusSliderHitRect() image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14093c220] 失效圆角控件区域。序言保存 rbx = prev rect。
func (s *qrCodeScreenSelectionSession) invalidateCornerRadiusControl(prev image.Rectangle) {
}

// [S-sig 0x14093c4a0] 失效圆角变更区域。序言保存 rbx/rcx/rdi = prev/cur rect 相关。
func (s *qrCodeScreenSelectionSession) invalidateCornerRadiusChange(prev, cur image.Rectangle) {
}

// [S-sig 0x14093c620] 失效圆角选区变更区域。序言保存 rbx = prev rect。
func (s *qrCodeScreenSelectionSession) invalidateCornerRadiusSelectionChange(prev image.Rectangle) {
}

// [S-sig 0x14093c900] 绘制选区缩放手柄。序言保存 rbx/rcx/rdi/rsi/r8 等 rect 参数；
// 无返回值。
func (s *qrCodeScreenSelectionSession) drawSelectionResizeHandles(hdc uintptr) {
}

// [S-sig 0x14093cd40] 绘制控件悬停高亮。序言保存 rbx = hdc；无返回值。
func (s *qrCodeScreenSelectionSession) drawControlHover(hdc uintptr) {
}

// [S-sig 0x14093ce40] 绘制窗口悬停高亮。序言保存 rbx = hdc；无返回值。
func (s *qrCodeScreenSelectionSession) drawWindowHover(hdc uintptr) {
}

// [S-sig 0x14093cf40] 绘制选区尺寸标签。序言保存 rbx/rcx/rdi/rsi/r8 等 rect 参数；
// 无返回值。
func (s *qrCodeScreenSelectionSession) drawSelectionSizeLabel(hdc uintptr) {
}

// [S-sig 0x14093dca0] 汇编实证：recv+sess(*screenshotWindowSelectionSession:rbx)+hwnd(uintptr:rcx)+p(image.Point:rdi/rsi)；
// 返回 image.Rectangle。rbx 透传 resolveControlSelectionForWindowAtPoint→
// screenshotControlBoundsAtPointThroughOverlaySessionWithTimeout 的 session 指针，类型已确证。
func (s *qrCodeScreenSelectionSession) resolveControlHoverForWindow(sess *screenshotWindowSelectionSession, hwnd uintptr, p image.Point) image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14093df20] morestack 仅保存 rax(recv)/rbx；rbx test nil 后 mov rax,rbx 传 screenshotWindowBounds(hwnd)；
// 返回 image.Rectangle（4 寄存器），失败路径 xor eax/rbx/rcx/rdi = 零 rect。
func (s *qrCodeScreenSelectionSession) resolveWindowHoverForWindow(hwnd uintptr) image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14093e160] 判断某窗口的悬停控件是否可复用缓存。序言保存 rbx/rcx/rdi/rsi
// 等参数；返回 bool。
func (s *qrCodeScreenSelectionSession) shouldReuseControlHoverForWindow(hwnd uintptr, p image.Point) bool {
	return false
}

// [S-sig 0x14093e240] 判断缓存悬停是否适用于某点。无参（读字段），返回 bool。
func (s *qrCodeScreenSelectionSession) cachedControlHoverAllowedForPoint(p image.Point) bool {
	return false
}

// [S-sig 0x14093e320] 汇编实证：recv+sess(*screenshotWindowSelectionSession:rbx)+p(image.Point:rcx/rdi)；
// 返回 image.Rectangle。rbx 透传 resolveControlSelectionForWindowAtPoint 的 session 指针，类型已确证。
func (s *qrCodeScreenSelectionSession) resolveControlSelectionAtPoint(sess *screenshotWindowSelectionSession, p image.Point) image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14093e3c0] 汇编实证：recv+sess(*screenshotWindowSelectionSession:rbx)+hwnd(uintptr:rcx)+
// p(image.Point:rdi/rsi)+force(bool:r8b)；返回 image.Rectangle。rbx 透传
// screenshotControlBoundsAtPointThroughOverlaySessionWithTimeout 的 session 指针，类型已确证。
func (s *qrCodeScreenSelectionSession) resolveControlSelectionForWindowAtPoint(sess *screenshotWindowSelectionSession, hwnd uintptr, p image.Point, force bool) image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14093e6e0] 汇编实证：recv+sess(*screenshotWindowSelectionSession:rbx)+p(image.Point:rcx/rdi)+
// controlRect(image.Rectangle:rsi/r8/r9/r10) 参数；返回 image.Rectangle。rbx 透传
// resolveControlSelectionAtPoint 的 session 指针，rsi..r10 透传
// qrCodeControlSelectionFromCachedHover 的 controlRect，类型已确证。
func (s *qrCodeScreenSelectionSession) resolveControlClickSelectionAtPoint(sess *screenshotWindowSelectionSession, p image.Point, controlRect image.Rectangle) image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14093e920] 返回某点所在的窗口句柄。序言保存 rbx/rcx = point；返回 uintptr。
func (s *qrCodeScreenSelectionSession) resolveWindowAtPoint(p image.Point) uintptr {
	return 0
}

// [S-sig 0x14093e9a0] 失效控件悬停变更区域。序言保存 rbx = prev rect。
func (s *qrCodeScreenSelectionSession) invalidateControlHoverChange(prev image.Rectangle) {
}

// [S-sig 0x14093ebe0] 失效窗口悬停变更区域。序言保存 rbx = prev rect。
func (s *qrCodeScreenSelectionSession) invalidateWindowHoverChange(prev image.Rectangle) {
}

// [S-sig 0x14093ee20] 失效选区变更区域。序言保存 rbx/rcx/rdi/rsi/r8 = prev/cur rect。
func (s *qrCodeScreenSelectionSession) invalidateSelectionChange(prev, cur image.Rectangle) {
}

// [S-sig 0x14093f420] 失效选区尺寸标签变更区域。序言保存 rbx = prev rect。
func (s *qrCodeScreenSelectionSession) invalidateSelectionSizeLabelChange(prev image.Rectangle) {
}

// [S-sig 0x14093f700] 汇编实证：recv+sess(*screenshotWindowSelectionSession:rbx) 参数；
// rbx 透传 resolveClickLikeControlSelection 的 session 指针，类型已确证；全返回路径无返回寄存器=void。
func (s *qrCodeScreenSelectionSession) prepareSelectionResult(sess *screenshotWindowSelectionSession) {
}

// [S-sig 0x14093f880] 汇编实证：recv+sess(*screenshotWindowSelectionSession:rbx) 参数；
// rbx 透传 resolveControlSelectionAtPoint 的 session 指针，类型已确证；返回 image.Rectangle。
func (s *qrCodeScreenSelectionSession) resolveClickLikeControlSelection(sess *screenshotWindowSelectionSession) image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14093fa40] 准备区域选区结果。序言保存 rbx/rcx/rdi/rsi = rect(4 int)。
// 无返回值。
func (s *qrCodeScreenSelectionSession) prepareAreaSelectionResult(rect image.Rectangle) {
}

// [S-sig 0x14093fce0] 最终化已准备的请求。序言保存 cl = confirm bool；返回 error。
func (s *qrCodeScreenSelectionSession) finalizePreparedRequest(confirm bool) error {
	return nil
}

// [S-sig 0x140940080] 开始标注编辑（置 editing + 初始化悬停/控制状态）。无参无返回值。
func (s *qrCodeScreenSelectionSession) beginAnnotationEditing() {
}

// [S-sig 0x1409403e0] 确认注解编辑。morestack 保存 rax(recv)/rbx；rbx test nil 后传
// hideScreenshotOverlayWindow(hwnd uintptr)，故 p=hwnd；两个返回路径均 xor eax 零值、无 mov eax,1，
// 返回 void（非 bool）。
func (s *qrCodeScreenSelectionSession) confirmAnnotationEditing(p uintptr) {
	_ = p
}

// [S-sig 0x1409407e0] 处理最终化请求（转发 finalizePreparedRequest）。无参无返回值。
func (s *qrCodeScreenSelectionSession) handleFinalizeRequest() {
}

// [S-sig 0x1409408c0] 显示标注工具栏。序言保存 rbx 参数（sessionID）；无返回值。
func (s *qrCodeScreenSelectionSession) showAnnotationToolbar() {
}

// [S-sig 0x140940c00] 同步标注工具栏（状态推送到窗口服务）。无参无返回值。
func (s *qrCodeScreenSelectionSession) syncAnnotationToolbar() {
}

// [S-sig 0x140940d40] 隐藏标注工具栏。无参无返回值。
func (s *qrCodeScreenSelectionSession) hideAnnotationToolbar() {
}

// [S-sig 0x140940de0] 返回标注工具栏边界矩形（activeSelectionRect + 偏移）。无参，
// 返回 image.Rectangle。
func (s *qrCodeScreenSelectionSession) annotationToolbarBounds() image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x140940e40] morestack 仅保存 rax(recv)；无参数；返回值经栈返回槽 [rsp+0x68..0x9c]
// 写回 screenshotSelectionToolbarState 结构（字段语义不展开，类型确证）。
func (s *qrCodeScreenSelectionSession) annotationToolbarState() screenshotSelectionToolbarState {
	return screenshotSelectionToolbarState{}
}

// [S-sig 0x140940fe0] 处理标注工具栏动作通道（select 循环）。序言 rax=recv、rbx=hwnd(uintptr)；
// morestack 保护 2 寄存器。hwnd 透传 applyAnnotationToolbarAction；无返回值。
func (s *qrCodeScreenSelectionSession) handleAnnotationToolbarActions(hwnd uintptr) {
}

// [S-sig 0x140941180] 汇编实证：rax=recv、rbx=hwnd(uintptr，nil 检查后透传 undoAnnotationAndInvalidate/
// invalidateAnnotationDirtyRect)、action(screenshotSelectionToolbarAction) 值传走栈（[rsp+0x148]=SessionID
// 与 recv[0x438] 比较、[rsp+0x150/0x158]=Action string TrimSpace 后分发）。morestack 保护 2 寄存器。无返回值。
// hwnd 自 qrCodeSelectionWindowProc(hwnd) → handleMessage → handleAnnotationToolbarActions 一路透传。
func (s *qrCodeScreenSelectionSession) applyAnnotationToolbarAction(hwnd uintptr, action screenshotSelectionToolbarAction) {
}

// [S-sig 0x140941c80] 处理标注指针按下。序言保存 rbx/rcx/rdi = point 相关；无返回值。
func (s *qrCodeScreenSelectionSession) handleAnnotationPointerDown(p image.Point) {
}

// [S-sig 0x1409434c0] 处理标注指针移动。序言保存 rbx/rcx/rdi = point 相关；无返回值。
func (s *qrCodeScreenSelectionSession) handleAnnotationPointerMove(p image.Point) {
}

// [S-sig 0x140943fe0] 处理标注指针抬起。序言保存 rbx = point 相关；无返回值。
func (s *qrCodeScreenSelectionSession) handleAnnotationPointerUp(p image.Point) {
}

// [S-sig 0x140944860] 处理标注键盘按下。序言 ecx 与 0x1b(ESC)/0xd(Enter) 比较 = keyCode。
func (s *qrCodeScreenSelectionSession) handleAnnotationKeyDown(keyCode uint32) {
}

// [S-sig 0x140945e20] 处理标注字符输入。序言 ecx 与 0x20/0x7f 比较 = ch uint32。
func (s *qrCodeScreenSelectionSession) handleAnnotationChar(ch uint32) {
}

// [S-sig 0x140945f20] 返回标注文本选区范围 (caret, anchor)。无参，返回 (int, int)。
func (s *qrCodeScreenSelectionSession) annotationTextSelectionRange() (int, int) {
	return 0, 0
}

// [S-sig 0x140946000] 替换标注文本选区。序言保存 rbx = text string(ptr,len) 首部。
func (s *qrCodeScreenSelectionSession) replaceAnnotationTextSelection(text string) {
}

// [S-sig 0x140946640] 向后删除标注文本。无参无返回值。
func (s *qrCodeScreenSelectionSession) deleteAnnotationTextBackward() {
}

// [S-sig 0x1409466c0] 向前删除标注文本。无参无返回值。
func (s *qrCodeScreenSelectionSession) deleteAnnotationTextForward() {
}

// [S-sig 0x1409467c0] 左移标注文本光标。序言保存 bl = extend bool。
func (s *qrCodeScreenSelectionSession) moveAnnotationTextCaretLeft(extend bool) {
}

// [S-sig 0x140946940] 右移标注文本光标。序言保存 bl = extend bool。
func (s *qrCodeScreenSelectionSession) moveAnnotationTextCaretRight(extend bool) {
}

// [S-sig 0x140946ac0] 返回选中的标注文本。无参，返回 string。
func (s *qrCodeScreenSelectionSession) selectedAnnotationText() string {
	return ""
}

// [S-sig 0x140946be0] 复制选中的标注文本到剪贴板。无参无返回值。
func (s *qrCodeScreenSelectionSession) copySelectedAnnotationText() {
}

// [S-sig 0x140946c20] 剪切选中的标注文本到剪贴板。无参无返回值。
func (s *qrCodeScreenSelectionSession) cutSelectedAnnotationText() {
}

// [S-sig 0x140946c80] 从剪贴板粘贴标注文本。无参无返回值。
func (s *qrCodeScreenSelectionSession) pasteAnnotationText() {
}

// [S-sig 0x140946ea0] 返回某点的标注文本光标索引。序言保存 rbx/rcx = point；返回 int。
func (s *qrCodeScreenSelectionSession) annotationTextIndexAtPoint(p image.Point) int {
	return 0
}

// [S-sig 0x140947340] 根据点位置更新标注文本字号。序言保存 rbx = point 相关。
func (s *qrCodeScreenSelectionSession) updateAnnotationTextFontSizeFromPoint(p image.Point) {
}

// [S-sig 0x1409474e0] 循环切换标注文本字体族。无参无返回值。
func (s *qrCodeScreenSelectionSession) cycleAnnotationTextFontFamily() {
}

// [S-sig 0x1409475a0] 新建标注笔画。序言保存 rbx/rcx = tool string(ptr,len)；返回 *qrCodeAnnotationStroke。
func (s *qrCodeScreenSelectionSession) newAnnotationStroke(tool string, p image.Point) *qrCodeAnnotationStroke {
	return nil
}

// [S-sig 0x140947840] 撤销标注并失效。序言 rax=recv、rbx=hwnd(uintptr)；无返回值。
func (s *qrCodeScreenSelectionSession) undoAnnotationAndInvalidate(hwnd uintptr) {
}

// [S-sig 0x140947ae0] 最终化标注文本（将输入 stroke 转正式笔画）。无参无返回值。
func (s *qrCodeScreenSelectionSession) finalizeAnnotationText() {
}

// [S-sig 0x140948880] 汇编实证：recv+hdc(uintptr:rbx，透传 drawQRCodeAnnotationStroke)+rect(image.Rectangle:rcx/rdi/rsi/r8)+
// bits(unsafe.Pointer:r9，透传 drawAnnotationStrokesOnPaintBuffer) 参数；无返回值。
func (s *qrCodeScreenSelectionSession) drawAnnotationOverlay(hdc uintptr, rect image.Rectangle, bits unsafe.Pointer) {
}

// [S-sig 0x140949200] 汇编实证：recv+pBits(unsafe.Pointer:rbx，透传 buildRGBAFromDIBBits)+
// selRect(image.Rectangle:rcx/rdi/rsi/r8)+previewRect(image.Rectangle:栈，透传 qrCodeRoundedCornerRects) 参数；
// 无返回值。旧存根误标 img *image.RGBA 且漏 previewRect，已订正。
func (s *qrCodeScreenSelectionSession) applyRoundedSelectionPreview(pBits unsafe.Pointer, selRect, previewRect image.Rectangle) {
}

// [S-sig 0x140949740] 返回圆角掩膜（缓存半径）。序言保存 rbx = radius int；返回 *image.Alpha。
func (s *qrCodeScreenSelectionSession) roundedCornerMask(radius int) *image.Alpha {
	return nil
}

// [S-sig 0x140949800] 返回内圆角掩膜（缓存半径）。序言保存 rbx = radius int；返回 *image.Alpha。
func (s *qrCodeScreenSelectionSession) roundedCornerInnerMask(radius int) *image.Alpha {
	return nil
}

// [S-sig 0x1409498c0] morestack 保存 rax(recv)/rbx/rcx/rdi/rsi/r8；rbx test nil 后传 buildRGBAFromDIBBits 的 pBits(unsafe.Pointer)；
// rcx/rdi/rsi/r8 = rect（Min.X/Min.Y/Max.X/Max.Y）；尾声 mov eax,1 / xor eax = bool。
func (s *qrCodeScreenSelectionSession) drawAnnotationStrokesOnPaintBuffer(bits unsafe.Pointer, rect image.Rectangle) bool {
	return false
}

// [S-sig 0x140949e40] 判断绘制矩形内是否有标注笔画。序言保存 rbx/rcx/rdi/rsi = rect；
// 返回 bool。
func (s *qrCodeScreenSelectionSession) hasAnnotationStrokeInPaintRect(rect image.Rectangle) bool {
	return false
}

// [S-sig 0x14094a160] 绘制标注文本编辑器。序言保存 rbx = hdc；无返回值。
func (s *qrCodeScreenSelectionSession) drawAnnotationTextEditor(hdc uintptr) {
}

// [S-sig 0x14094bae0] 失效标注文本区域。序言保存 rbx/rcx/rdi/rsi/r8 = rect 相关。
func (s *qrCodeScreenSelectionSession) invalidateAnnotationTextAreaWith(rect image.Rectangle) {
}

// [S-sig 0x14094bc80] 失效标注脏矩形。序言 rax=recv、rbx=hwnd(uintptr，test nil 后 [rsp+0x78])、
// rcx/rdi/rsi/r8=rect(Min.X/Min.Y/Max.X/Max.Y，cmp rcx,rsi 与 rdi,r8 判空)；无返回值。
func (s *qrCodeScreenSelectionSession) invalidateAnnotationDirtyRect(hwnd uintptr, rect image.Rectangle) {
}

// [S-sig 0x14094bdc0] 返回标注文本失效矩形。无参，返回 image.Rectangle。
func (s *qrCodeScreenSelectionSession) annotationTextInvalidationRect() image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14094bfa0] 返回标注内容失效矩形（所有笔画并集）。无参，返回 image.Rectangle。
func (s *qrCodeScreenSelectionSession) annotationContentInvalidationRect() image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14094c160] 返回标注瞬态失效矩形（悬停/拖拽状态）。无参，返回 image.Rectangle。
func (s *qrCodeScreenSelectionSession) annotationTransientInvalidationRect() image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14094c480] 失效标注笔画变更区域。序言保存 rbx = prev rect。
func (s *qrCodeScreenSelectionSession) invalidateAnnotationStrokeChange(prev image.Rectangle) {
}

// [S-sig 0x14094c7a0] 返回标注文本编辑器矩形（含位置/边界计算）。无参，返回 image.Rectangle。
func (s *qrCodeScreenSelectionSession) annotationTextEditorRect() image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14094cac0] 返回标注文本编辑器边界矩形。无参，返回 image.Rectangle。
func (s *qrCodeScreenSelectionSession) annotationTextEditorBounds() image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14094cb20] 返回标注文本编辑器位置。无参，返回 image.Point。
func (s *qrCodeScreenSelectionSession) annotationTextEditorPosition() image.Point {
	return image.Point{}
}

// [S-sig 0x14094d1e0] 返回标注文本拖拽手柄矩形（editor rect +8/+0x34 偏移）。
// 无参，返回 image.Rectangle。
func (s *qrCodeScreenSelectionSession) annotationTextDragHandleRect() image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14094d240] 返回标注文本确认按钮矩形。无参，返回 image.Rectangle。
func (s *qrCodeScreenSelectionSession) annotationTextConfirmRect() image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14094d2a0] 返回标注文本输入矩形。无参，返回 image.Rectangle。
func (s *qrCodeScreenSelectionSession) annotationTextInputRect() image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14094d300] 返回标注文本颜色色板矩形。无参，返回 image.Rectangle。
func (s *qrCodeScreenSelectionSession) annotationTextSwatchRect() image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14094d360] 返回标注文本字号滑块矩形。无参，返回 image.Rectangle。
func (s *qrCodeScreenSelectionSession) annotationTextSliderRect() image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14094d420] 返回标注文本字号滑块命中矩形。无参，返回 image.Rectangle。
func (s *qrCodeScreenSelectionSession) annotationTextSliderHitRect() image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14094d460] 返回标注文本字体矩形。无参，返回 image.Rectangle。
func (s *qrCodeScreenSelectionSession) annotationTextFontRect() image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x140952020] 返回某点命中的标注笔画。序言保存 rbx/rcx = point；返回 *qrCodeAnnotationStroke。
func (s *qrCodeScreenSelectionSession) annotationStrokeAt(p image.Point) *qrCodeAnnotationStroke {
	return nil
}

// [S-sig 0x1409548e0] 将标注后的选区编码为 PNG。无显式参数（读字段），返回 ([]byte, error)。
func (s *qrCodeScreenSelectionSession) encodeAnnotatedSelectionPNG() ([]byte, error) {
	return nil, nil
}

// [S-sig 0x140959b60] 准备窗口结果（记录 sourceProcessName 等）。序言保存 rbx = hwnd uintptr。
func (s *qrCodeScreenSelectionSession) prepareWindowResult(hwnd uintptr) {
}

// [S-sig 0x140959e80] 触发接受通知（sync.Once 保证单次）。无参无返回值。
func (n *qrCodeCaptureAcceptedNotifier) Emit() {
}
