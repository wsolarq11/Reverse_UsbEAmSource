// AUTO-RECONSTRUCTED — DOMAIN: mousegestures (win32 runtime / event queue / session)
// Source: UsbEAm_Launcher 1.0.3 (Go 1.25.12, PE64), disassembled from
// main.(*mouseGestureEventQueue).* / main.(*mouseGesturePlatformRuntime).* /
// main.(*mouseGestureRuntimeSession).* / main.mouseGesture* symbols.
// Tier markers:
//
//	[S VA]      — body fully translated from asm (pure geometry / math).
//	[S-sig VA]  — signature proven from asm + types_gesture.go; body is a faithful zero
//	              skeleton（SetWindowsHookEx / chan / mutex / timer deps）。「推断」= 形参
//	              类型来自字段偏移/命名，未经逐条序言反汇编实证。
//	[P]         — 签名待实证：references an unlanded type; zero skeleton only.
//
// 研究用途
package main

import (
	"image"
	"time"
)

// *mouseGestureEventQueue —— 无锁唤醒的事件队列（hook 事件入队/出队）。
// [S-sig 0x1408f0040] tryPushAndWake: 尝试入队并在首次成功时唤醒运行时。
func (q *mouseGestureEventQueue) tryPushAndWake(ev mouseGestureHookEvent) bool {
	return false
}

// [S-sig 0x1408f02e0] lockForPush: 加锁并返回是否有空闲槽可入队。
func (q *mouseGestureEventQueue) lockForPush(ev mouseGestureHookEvent) bool {
	return false
}

// [S-sig 0x1408f03c0] pushLocked: 在持锁状态下写入事件。
func (q *mouseGestureEventQueue) pushLocked(ev mouseGestureHookEvent) {
}

// [S-sig 0x1408f09a0] pop: 出队一个事件；空队列返回 (零值, false)。
func (q *mouseGestureEventQueue) pop() (mouseGestureHookEvent, bool) {
	return mouseGestureHookEvent{}, false
}

// *mouseGestureService —— 平台运行时的生命周期同步。
// [S-sig 0x1408f0c80] syncPlatformRuntime: 按当前配置同步（启动/停止）平台运行时。
func (s *mouseGestureService) syncPlatformRuntime() {
}

// [S-sig 0x1408f1860] stopPlatformRuntime: 停止平台运行时并等待退出。
func (s *mouseGestureService) stopPlatformRuntime() {
}

// [S-sig 0x1408f1980] stopPlatformRuntimeLocked: 持锁停止平台运行时。
func (s *mouseGestureService) stopPlatformRuntimeLocked() {
}

// [S-sig 0x1408f1b40] runtimeSnapshot: 返回运行时活跃状态快照。
func (s *mouseGestureService) runtimeSnapshot() MouseGestureState {
	return MouseGestureState{}
}

// [S-sig 0x1408f1d60] capturePausedSnapshot: 返回暂停捕获时的状态快照。
func (s *mouseGestureService) capturePausedSnapshot() MouseGestureState {
	return MouseGestureState{}
}

// *mouseGesturePlatformRuntime —— 底层钩子与事件泵。
// [S-sig 0x1408f1e80] run: 运行时主循环（处理 hook 事件与按钮回放）。
func (r *mouseGesturePlatformRuntime) run() error {
	return nil
}

// [S-sig 0x1408f25e0] startButtonReplayWorker: 启动按键回放 worker。
func (r *mouseGesturePlatformRuntime) startButtonReplayWorker() {
}

// [S-sig 0x1408f2a20] executeButtonReplay: 执行单个按键回放任务。
func (r *mouseGesturePlatformRuntime) executeButtonReplay(replay mouseGestureButtonReplay) {
}

// [S-sig 0x1408f2b20] stopWatcher: 停止运行时 watcher。
func (r *mouseGesturePlatformRuntime) stopWatcher() {
}

// [S-sig 0x1408f2bc0] wakeRuntime: 唤醒运行时处理挂起事件。
func (r *mouseGesturePlatformRuntime) wakeRuntime() {
}

// [S-sig 0x1408f2c80] processPending: 处理队列中挂起的事件。
func (r *mouseGesturePlatformRuntime) processPending() {
}

// [S-sig 0x1408f2f20] installHook: 安装 WH_MOUSE_LL 低级鼠标钩子。
func (r *mouseGesturePlatformRuntime) installHook() error {
	return nil
}

// [S-sig 0x1408f3260] acceptHookEvent: 判定是否接受该 hook 事件进入会话。
func (r *mouseGesturePlatformRuntime) acceptHookEvent(ev mouseGestureHookEvent) bool {
	return false
}

// [S-sig 0x1408f34c0] uninstallHook: 卸载低级鼠标钩子。
func (r *mouseGesturePlatformRuntime) uninstallHook() {
}

// [S-sig 0x1408f3620] shouldSuppressHookEvent: 判定是否抑制该 hook 事件。
func (r *mouseGesturePlatformRuntime) shouldSuppressHookEvent(ev mouseGestureHookEvent) bool {
	return false
}

// [S-sig 0x1408f37e0] shouldQueueHookEvent: 判定是否将该 hook 事件入队（形参推断）。
func (r *mouseGesturePlatformRuntime) shouldQueueHookEvent(ev mouseGestureHookEvent, injected bool) bool {
	return false
}

// [S-sig 0x1408f3880] queueAndWakeHookEvent: 入队并唤醒运行时。
func (r *mouseGesturePlatformRuntime) queueAndWakeHookEvent(ev mouseGestureHookEvent) {
}

// [S-sig 0x1408f3920] noteQueueFailure: 记录一次入队失败（供诊断）。
func (r *mouseGesturePlatformRuntime) noteQueueFailure(ev mouseGestureHookEvent) {
}

// [S-sig 0x1408f39e0] shouldCaptureGestureButtonDown: 判定是否捕获该按键按下事件。
func (r *mouseGesturePlatformRuntime) shouldCaptureGestureButtonDown(ev mouseGestureHookEvent) bool {
	return false
}

// [S-sig 0x1408f3d60] runtimeConfig: 返回运行时当前配置。
func (r *mouseGesturePlatformRuntime) runtimeConfig() MouseGestureConfig {
	return MouseGestureConfig{}
}

// [S-sig 0x1408f3ee0] gestureRuntimeEnabled: 手势运行时是否启用。
func (r *mouseGesturePlatformRuntime) gestureRuntimeEnabled() bool {
	return false
}

// [S-sig 0x1408f4020] gestureTargetHWNDAtPoint: 点位目标窗口 HWND（跳过自身）。
func (r *mouseGesturePlatformRuntime) gestureTargetHWNDAtPoint(x, y int, skipSelf bool) uintptr {
	return 0
}

// [S-sig 0x1408f40a0] storeGestureTargetScope: 存储目标窗口作用域（hwnd + allowed）。
func (r *mouseGesturePlatformRuntime) storeGestureTargetScope(scope mouseGestureTargetScope) {
}

// [S-sig 0x1408f4140] mouseGestureHookEventFromWindows: 将原始 hook 数据转为事件结构。
func mouseGestureHookEventFromWindows(message uint32, wParam, lParam uintptr) mouseGestureHookEvent {
	return mouseGestureHookEvent{}
}

// *mouseGestureRuntimeSession —— 单次手势的识别与状态机。
// [S-sig 0x1408f4400] handleEvent: 会话事件入口（分发到按下/移动/抬起处理）。
func (s *mouseGestureRuntimeSession) handleEvent(ev mouseGestureHookEvent) {
}

// [S-sig 0x1408f49a0] handleGesture: 处理一个手势事件并推进识别。
func (s *mouseGestureRuntimeSession) handleGesture(ev mouseGestureHookEvent) {
}

// [S-sig 0x1408f5960] currentTime: 返回会话时钟（注入 now() 或 time.Now）。
func (s *mouseGestureRuntimeSession) currentTime() time.Time {
	return time.Time{}
}

// [S-sig 0x1408f59a0] sendDown: 发送按键按下（注入或平台默认）。
func (s *mouseGestureRuntimeSession) sendDown(button string) error {
	return nil
}

// [S-sig 0x1408f5a20] sendUp: 发送按键抬起。
func (s *mouseGestureRuntimeSession) sendUp(button string) error {
	return nil
}

// [S-sig 0x1408f5aa0] sendClick: 发送按键点击。
func (s *mouseGestureRuntimeSession) sendClick(button string) error {
	return nil
}

// [S-sig 0x1408f5b20] gestureTimeoutStatus: 返回手势是否处于超时敏感状态（形参推断）。
func (s *mouseGestureRuntimeSession) gestureTimeoutStatus() bool {
	return false
}

// [S-sig 0x1408f5c80] armGestureTimeout: 布防手势超时。
func (s *mouseGestureRuntimeSession) armGestureTimeout() {
}

// [S-sig 0x1408f5e60] scheduleGestureTimeoutWake: 调度超时唤醒。
func (s *mouseGestureRuntimeSession) scheduleGestureTimeoutWake() {
}

// [S-sig 0x1408f6000] stopGestureTimeout: 停止手势超时计时。
func (s *mouseGestureRuntimeSession) stopGestureTimeout() {
}

// [S-sig 0x1408f6100] handleTimeout: 处理手势超时。
func (s *mouseGestureRuntimeSession) handleTimeout() {
}

// [S-sig 0x1408f6260] handleQueueFailure: 处理事件队列失败。
func (s *mouseGestureRuntimeSession) handleQueueFailure(ev mouseGestureHookEvent) {
}

// [S-sig 0x1408f6320] shutdown: 关闭会话并释放资源。
func (s *mouseGestureRuntimeSession) shutdown() {
}

// [S-sig 0x1408f63a0] handleGestureMove: 处理手势移动（追加轨迹点）。
func (s *mouseGestureRuntimeSession) handleGestureMove(ev mouseGestureHookEvent) {
}

// [S-sig 0x1408f6ce0] shouldUpdateGestureOverlay: 是否应更新 overlay。
func (s *mouseGestureRuntimeSession) shouldUpdateGestureOverlay() bool {
	return false
}

// [S-sig 0x1408f6da0] resetGestureSession: 重置会话状态。
func (s *mouseGestureRuntimeSession) resetGestureSession() {
}

// [S-sig 0x1408f6f00] awaitGestureButtonRelease: 等待手势按键抬起。
func (s *mouseGestureRuntimeSession) awaitGestureButtonRelease(button string) bool {
	return false
}

// [S-sig 0x1408f70a0] completeAwaitingGestureButtonRelease: 完成按键抬起等待。
func (s *mouseGestureRuntimeSession) completeAwaitingGestureButtonRelease() {
}

// [S-sig 0x1408f71a0] queueGestureButtonClickReplay: 排队一次按键点击回放。
func (s *mouseGestureRuntimeSession) queueGestureButtonClickReplay(button string) {
}

// [S-sig 0x1408f73c0] cancelGestureWithRestore: 取消手势并恢复抑制状态。
func (s *mouseGestureRuntimeSession) cancelGestureWithRestore() {
}

// [S-sig 0x1408f75a0] finishGesture: 完成手势并执行命中动作。
func (s *mouseGestureRuntimeSession) finishGesture() {
}

// [S-sig 0x1408f7d40] finishGestureTarget: 完成手势并作用于目标窗口（形参推断）。
func (s *mouseGestureRuntimeSession) finishGestureTarget(target MouseGestureTarget) {
}

// [S-sig 0x1408f8100] currentMatchedGestureLabel: 当前命中手势的标签。
func (s *mouseGestureRuntimeSession) currentMatchedGestureLabel() string {
	return ""
}

// [S-sig 0x1408f8460] observeGestureTargetScope: 观测目标窗口作用域。
func (s *mouseGestureRuntimeSession) observeGestureTargetScope() {
}

// [S-sig 0x1408f86a0] ensureGestureTargetResolved: 确保目标窗口已解析。
func (s *mouseGestureRuntimeSession) ensureGestureTargetResolved() {
}

// [S-sig 0x1408f87e0] gestureTargetHWNDAtPoint: 会话级点位目标窗口 HWND。
func (s *mouseGestureRuntimeSession) gestureTargetHWNDAtPoint(x, y int, skipSelf bool) uintptr {
	return 0
}

// [S-sig 0x1408f8860] gestureTargetAtPoint: 会话级点位目标。
func (s *mouseGestureRuntimeSession) gestureTargetAtPoint(x, y int, skipSelf bool) MouseGestureTarget {
	return MouseGestureTarget{}
}

// [S-sig 0x1408f89e0] handleHotCorner: 处理热区角触发（形参推断）。
func (s *mouseGestureRuntimeSession) handleHotCorner(corner string) {
}

// [S-sig 0x1408f91e0] executeHotCornerRule: 执行热区角规则（发送热键等）。
func executeHotCornerRule(rule HotCornerRule) error {
	return nil
}

// [S-sig 0x1408f9440] mouseGestureRuntimeTargetHWNDAtPoint: 运行时点位目标 HWND。
func mouseGestureRuntimeTargetHWNDAtPoint(x, y int, skipSelf bool) uintptr {
	return 0
}

// [S-sig 0x1408f94c0] mouseGestureTargetFromHWND: 由 HWND 解析手势目标。
func mouseGestureTargetFromHWND(hwnd uintptr) MouseGestureTarget {
	return MouseGestureTarget{}
}

// [S-sig 0x1408f96e0] mouseGestureWindowFromPoint: 点位所在窗口 HWND。
func mouseGestureWindowFromPoint(x, y int) uintptr {
	return 0
}

// [S-sig 0x1408f9760] mouseGestureForegroundWindowFullscreen: 前台窗口是否全屏。
func mouseGestureForegroundWindowFullscreen() bool {
	return false
}

// [S-sig 0x1408f98e0] mouseGestureScreenBoundsFromRects: 由监视器矩形列表合成屏幕边界（形参推断）。
func mouseGestureScreenBoundsFromRects(rects []image.Rectangle) []MouseGestureScreenBounds {
	return nil
}
