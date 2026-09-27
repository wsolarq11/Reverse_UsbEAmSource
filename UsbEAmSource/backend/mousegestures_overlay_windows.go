// AUTO-RECONSTRUCTED — DOMAIN: mousegestures (overlay window / render)
// Source: UsbEAm_Launcher 1.0.3 (Go 1.25.12, PE64), disassembled from
// main.(*mouseGestureOverlayWindow).* / main.mouseGestureOverlay* symbols.
// 画布类型经汇编实证为 *image.RGBA（+0x18 Stride，+0x20..+0x38 为 Rect.Min/Max）。
// Tier markers:
//
//	[S VA]      — body fully translated from asm (pure geometry / math).
//	[S-sig VA]  — signature proven from asm + types_gesture.go; body is a faithful zero
//	              skeleton (Win32 / GDI / chan / mutex deps)。「推断」= 颜色通道/宽度等
//	              形参顺序来自序言字节溢出模式，未逐条实证。
//	[P]         — 签名待实证：references an unlanded type; zero skeleton only.
//
// 研究用途
package main

import "image"

// *mouseGestureRuntimeSession —— 手势会话的 overlay 驱动（接收者与 runtime 文件同源）。
// [S-sig 0x1408e8100] updateGestureOverlay: 用轨迹点更新手势轨迹 overlay。
func (s *mouseGestureRuntimeSession) updateGestureOverlay(points []MouseGesturePoint, button string) {
}

// [S-sig 0x1408e8320] updateGestureLabelOverlay: 更新手势名标签 overlay。
func (s *mouseGestureRuntimeSession) updateGestureLabelOverlay(label string) {
}

// [S-sig 0x1408e8520] finishGestureOverlay: 结束手势 overlay 展示。
func (s *mouseGestureRuntimeSession) finishGestureOverlay() {
}

// [S-sig 0x1408e8600] hideGestureOverlay: 隐藏手势 overlay。
func (s *mouseGestureRuntimeSession) hideGestureOverlay() {
}

// [S-sig 0x1408e8660] closeGestureOverlay: 关闭并释放手势 overlay 窗口。
func (s *mouseGestureRuntimeSession) closeGestureOverlay() {
}

// [S-sig 0x1408e8720] createMouseGestureOverlayWindow: 创建 overlay 窗口并启动消息线程。
func createMouseGestureOverlayWindow() (*mouseGestureOverlayWindow, error) {
	return nil, nil
}

// [S-sig 0x1408e88c0] ensureMouseGestureOverlayWindowClass: 注册（惰性）overlay 窗口类。
func ensureMouseGestureOverlayWindowClass() {
}

// *mouseGestureOverlayWindow —— 无边框置顶 overlay 窗口。
// [S-sig 0x1408e8920] run: overlay 窗口消息循环。
func (w *mouseGestureOverlayWindow) run() error {
	return nil
}

// [S-sig 0x1408e8c20] createWindow: 在 UI 线程创建原生窗口。
func (w *mouseGestureOverlayWindow) createWindow() error {
	return nil
}

// [S-sig 0x1408e9000] handleMessage: 分发窗口消息（WM_PAINT / WM_NCHITTEST 等）。
func (w *mouseGestureOverlayWindow) handleMessage(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	return 0
}

// [S-sig 0x1408e9240] Update: 提交新 payload 触发重绘。
func (w *mouseGestureOverlayWindow) Update(payload mouseGestureOverlayPayload) {
}

// [S-sig 0x1408e9460] UpdateLabel: 仅更新标签（labelSnapshot）。
func (w *mouseGestureOverlayWindow) UpdateLabel(label string) {
}

// [S-sig 0x1408e9600] Fade: 触发淡出动画。
func (w *mouseGestureOverlayWindow) Fade() {
}

// [S-sig 0x1408e97a0] Hide: 立即隐藏窗口。
func (w *mouseGestureOverlayWindow) Hide() {
}

// [S-sig 0x1408e98c0] BringToTop: 将窗口置顶。
func (w *mouseGestureOverlayWindow) BringToTop() {
}

// [S-sig 0x1408e9940] Close: 关闭窗口并停止消息线程。
func (w *mouseGestureOverlayWindow) Close() {
}

// [S-sig 0x1408e9b40] queuePayload: 入队待应用 payload（覆盖 pendingPayload）。
func (w *mouseGestureOverlayWindow) queuePayload(payload mouseGestureOverlayPayload) {
}

// [S-sig 0x1408e9ce0] applyPendingPayloadOnThread: 在线程上应用 pending payload。
func (w *mouseGestureOverlayWindow) applyPendingPayloadOnThread() {
}

// [S-sig 0x1408e9e40] applyPayloadOnThread: 在线程上应用指定 payload。
func (w *mouseGestureOverlayWindow) applyPayloadOnThread(payload mouseGestureOverlayPayload) {
}

// [S-sig 0x1408ea060] hideOnThread: 在线程上隐藏窗口。
func (w *mouseGestureOverlayWindow) hideOnThread() {
}

// [S-sig 0x1408ea1e0] stepFadeOnThread: 在线程上执行单步淡出。
func (w *mouseGestureOverlayWindow) stepFadeOnThread() {
}

// [S-sig 0x1408ea2c0] redraw: 触发一次重绘。
func (w *mouseGestureOverlayWindow) redraw() {
}

// [S-sig 0x1408eaa80] finishRedraw: 结束重绘（清 redrawing 标记）。
func (w *mouseGestureOverlayWindow) finishRedraw() {
}

// [S-sig 0x1408eaba0] render: 将 payload 渲染到画布。
func (w *mouseGestureOverlayWindow) render(canvas *image.RGBA) {
}

// [S-sig 0x1408eae00] paint: WM_PAINT 处理（blit 到窗口 DC）。
func (w *mouseGestureOverlayWindow) paint() {
}

// [S-sig 0x1408eafc0] setThreadID: 记录窗口消息线程 ID。
func (w *mouseGestureOverlayWindow) setThreadID(id uintptr) {
}

// [S-sig 0x1408eb060] setHandle: 记录窗口 HWND。
func (w *mouseGestureOverlayWindow) setHandle(hwnd uintptr) {
}

// [S-sig 0x1408eb100] handle: 返回窗口 HWND。
func (w *mouseGestureOverlayWindow) handle() uintptr {
	return 0
}

// [S-sig 0x1408e8e40] mouseGestureOverlayWindowProc: overlay 窗口消息回调（WinProc）。
func mouseGestureOverlayWindowProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	return 0
}

// [S-sig 0x1408eaf00] mouseGestureOverlayBringToTop: 将指定 HWND 置顶。
func mouseGestureOverlayBringToTop(hwnd uintptr) {
}

// [S-sig 0x1408eb220] buildMouseGestureOverlayPayload: 由轨迹点/按键构建 overlay payload。
func buildMouseGestureOverlayPayload(points []MouseGesturePoint, button string) mouseGestureOverlayPayload {
	return mouseGestureOverlayPayload{}
}

// [S-sig 0x1408eb760] buildMouseGestureLabelOverlayPayload: 由标签构建标签 overlay payload。
func buildMouseGestureLabelOverlayPayload(label string) mouseGestureOverlayPayload {
	return mouseGestureOverlayPayload{}
}

// [S-sig 0x1408eb9e0] mouseGestureOverlayVisibleTrailPoints: 裁剪出可见轨迹点。
func mouseGestureOverlayVisibleTrailPoints(points []MouseGesturePoint) []MouseGesturePoint {
	return nil
}

// [S-sig 0x1408ebfc0] mouseGestureOverlayBounds: 计算 overlay 包围矩形。
func mouseGestureOverlayBounds(payload mouseGestureOverlayPayload) image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x1408ec1c0] mouseGestureOverlayLabelBounds: 计算标签包围矩形。
func mouseGestureOverlayLabelBounds(payload mouseGestureOverlayPayload) image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x1408ec300] mouseGestureOverlayLabelBoundsForScreen: 按屏幕计算标签包围矩形（形参推断）。
func mouseGestureOverlayLabelBoundsForScreen(payload mouseGestureOverlayPayload, bounds MouseGestureScreenBounds) image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x1408ec820] mouseGestureOverlayScreenForPoint: 点坐标 → 所在屏幕边界（形参推断）。
func mouseGestureOverlayScreenForPoint(x, y int) MouseGestureScreenBounds {
	return MouseGestureScreenBounds{}
}

// [S-sig 0x1408ec940] renderMouseGestureOverlay: 将 payload 绘制到画布（编排各绘制原语）。
func renderMouseGestureOverlay(canvas *image.RGBA, payload mouseGestureOverlayPayload) {
}

// [S-sig 0x1408ecf60] mouseGestureOverlayDrawLabel: 在画布绘制标签文本（形参推断）。
func mouseGestureOverlayDrawLabel(canvas *image.RGBA, label string, bounds image.Rectangle) {
}

// [S-sig 0x1408ed660] mouseGestureOverlayDrawText: 在画布绘制文本（形参推断）。
func mouseGestureOverlayDrawText(canvas *image.RGBA, text string, x, y int) {
}

// [S-sig 0x1408ee060] mouseGestureOverlayMeasureTextWidth: 测量文本像素宽度（形参推断）。
func mouseGestureOverlayMeasureTextWidth(canvas *image.RGBA, text string) int {
	return 0
}

// [S-sig 0x1408ee620] mouseGestureOverlayCreateCanvasBitmap: 按现有画布 bounds 创建/缩放位图。
func mouseGestureOverlayCreateCanvasBitmap(canvas *image.RGBA) *image.RGBA {
	return nil
}

// [S-sig 0x1408ee8a0] mouseGestureOverlayBlendTextDIBToRGBA: 将文本 DIB 混合进 RGBA 画布（形参推断）。
func mouseGestureOverlayBlendTextDIBToRGBA(canvas *image.RGBA, dib []byte) {
}

// [S-sig 0x1408eeb60] mouseGestureOverlayDrawPolylineStroke: 折线描边（颜色/宽度形参推断）。
func mouseGestureOverlayDrawPolylineStroke(canvas *image.RGBA, points []MouseGesturePoint, color uint32, width int) {
}

// [S-sig 0x1408eecc0] mouseGestureOverlaySimplifyStrokePoints: 抽稀轨迹点（最小间距过滤）。
func mouseGestureOverlaySimplifyStrokePoints(points []MouseGesturePoint, minDistance int) []MouseGesturePoint {
	return nil
}

// [S-sig 0x1408eef20] mouseGestureOverlayDrawMaxStroke: 绘制主轨迹（最大宽度描边，形参推断）。
func mouseGestureOverlayDrawMaxStroke(canvas *image.RGBA, points []MouseGesturePoint, color uint32) {
}

// [S-sig 0x1408ef4a0] mouseGestureOverlaySetMaxPixel: 以最大值写像素（形参推断）。
func mouseGestureOverlaySetMaxPixel(canvas *image.RGBA, x, y int, color uint32) {
}

// [S-sig 0x1408ef660] mouseGestureOverlayFillRoundedRect: 填充圆角矩形（形参推断）。
func mouseGestureOverlayFillRoundedRect(canvas *image.RGBA, bounds image.Rectangle, radius int, color uint32) {
}

// [S-sig 0x1408ef900] mouseGestureOverlayBlendPixel: 将像素按 alpha 混合入画布（形参推断）。
func mouseGestureOverlayBlendPixel(canvas *image.RGBA, x, y int, color uint32) {
}

// [S-sig 0x1408efae0] mouseGestureOverlayPremultiplyAlpha: 将画布像素就地预乘 alpha。
func mouseGestureOverlayPremultiplyAlpha(canvas *image.RGBA) {
}

// [S-sig 0x1408efcc0] mouseGestureOverlayEstimateTextWidth: 估算文本像素宽度（形参推断）。
func mouseGestureOverlayEstimateTextWidth(text string) int {
	return 0
}
