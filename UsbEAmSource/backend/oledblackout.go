// AUTO-RECONSTRUCTED SERVICE METHODS — DOMAIN: oledblackout 工厂 + 执行域（批次 22 全量 [S]）
// 研究用途
//
// 契约来源：
//   - newOLEDBlackoutService(0x1408fef60, 704B) 汇编
//   - Configure(0x1408ff660) / GetState(0x1408ff940) / GetScreens(0x1408ffbe0) /
//     ToggleProfile(0x1408ffd60) / StartProfile(0x140900a80) 汇编（批次 22 新增 dump）
//   - 结构 oledBlackoutService / 接口 oledBlackoutHotkeyManager：types_oled.go
//
// 档位：[S] 工厂装配 + 5 执行方法入口流；内部锁内调用（backfillProfileScreenMetadataLocked、
//
//	reconcileActiveProfileLocked、buildStateLocked、collectScreenStatesLocked、
//	resolveProfileScreensLocked、activateProfileLocked、hideProfileScreensLocked、
//	hideAllLocked、syncVisibleOverlayStateLocked、configureHotkeys、configureIdleTimer、
//	normalizeOLEDBlackoutConfig 等）为 [P] 脚手架，债务：体待各子域专项还原。
package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// newOLEDBlackoutService 构造 OLED 熄屏服务。
// [S 汇编 0x1408fef60]：ensureOLEDBlackoutDebugLogger→debugLog→newobject→5 map 装配→lock→
// ensureLifecycleLocked→挂 hotkeyManager(newOLEDBlackoutHotkeyManager)。
func newOLEDBlackoutService(app *application.App) *oledBlackoutService {
	ensureOLEDBlackoutDebugLogger()
	oledBlackoutDebugLog("newOLEDBlackoutService executing")
	s := &oledBlackoutService{
		app:              app,
		overlayWindows:   make(map[string]*oledBlackoutOverlayWindow),
		visibleOverlays:  make(map[string]struct{}),
		hotkeyRegistered: make(map[string]bool),
		hotkeyErrors:     make(map[string]string),
		lastPressedKeys:  make(map[uintptr]struct{}),
	}
	// 生命周期字段接线（汇编 ensureLifecycleLocked）。
	s.ensureLifecycleLocked()
	// assemble hotkey manager（汇编挂 hotkeyManager 字段）。
	s.hotkeyManager = newOLEDBlackoutHotkeyManager()
	return s
}

// --- oled 子域内部脚手架 ---

// ensureLifecycleLocked 初始化生命周期字段（幂等）。
// [S 0x1408ff220] 单 receiver 方法，无参无返回。asm：
//   shutdownDone==nil → make(chan struct{})；lifecycleGeneration==0 → 1；
//   lifecycleContext==nil → WithCancel(context.Background())；shuttingDown → lifecycleCancel()。
func (s *oledBlackoutService) ensureLifecycleLocked() {
	if s.shutdownDone == nil {
		s.shutdownDone = make(chan struct{})
	}
	if s.lifecycleGeneration == 0 {
		s.lifecycleGeneration = 1
	}
	if s.lifecycleContext == nil {
		s.lifecycleContext, s.lifecycleCancel = context.WithCancel(context.Background())
	}
	if s.shuttingDown {
		s.lifecycleCancel()
	}
}

// beginBackgroundActivity 加锁并登记一个后台活动（幂等，代际校验）。
// [S 0x1408ff340] 单 receiver + generation uint64，返回 bool。asm：lock.Lock→defer Unlock→
// ensureLifecycleLocked→shuttingDown||lifecycleGeneration!=generation 返回 false→
// backgroundActivities.Add(1) 返回 true。
func (s *oledBlackoutService) beginBackgroundActivity(generation uint64) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ensureLifecycleLocked()
	if s.shuttingDown || s.lifecycleGeneration != generation {
		return false
	}
	s.backgroundActivities.Add(1)
	return true
}

// endBackgroundActivity 注销一个后台活动。
// [S 0x1408ff4c0] 单 receiver 无参无返回。asm：backgroundActivities(+0x10).Add(-1)。
func (s *oledBlackoutService) endBackgroundActivity() {
	s.backgroundActivities.Add(-1)
}

// normalizeOLEDBlackoutMediaPauseExclusionPath 归一化媒体暂停排除路径。
// [S 0x1408fdd60] 单参 string 返回 string。asm：TrimSpace；空→返回空串；否则 filepath.Clean。
func normalizeOLEDBlackoutMediaPauseExclusionPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	return filepath.Clean(path)
}

// oledBlackoutMediaPausePathBase 取媒体暂停排除路径的最后一段（base name）。
// [S 0x1408fde40] 单参 string 返回 string。asm：TrimSpace → TrimRight(`\/`) → 空返空 →
// Replace(`\`→`/`) → LastIndex(`/`) <0 返回整串；否则返回 idx+1 起切片。
func oledBlackoutMediaPausePathBase(path string) string {
	path = strings.TrimSpace(path)
	path = strings.TrimRight(path, `\/`)
	if path == "" {
		return ""
	}
	path = strings.Replace(path, "\\", "/", -1)
	idx := strings.LastIndex(path, "/")
	if idx < 0 {
		return path
	}
	return path[idx+1:]
}

// normalizeOLEDBlackoutMediaPauseExclusionProcessName 归一化媒体暂停排除进程名。
// [S 0x1408fddc0] 单参 string 返回 string。asm：TrimSpace → 空返空 → PathBase →
// 空/`.`/`..` 返空；否则 ToLower。
func normalizeOLEDBlackoutMediaPauseExclusionProcessName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	name = oledBlackoutMediaPausePathBase(name)
	if name == "" || name == "." || name == ".." {
		return ""
	}
	return strings.ToLower(name)
}

// idleRunCurrentLocked 判断当前空闲运行代际/调度 ID 是否匹配（加锁上下文内）。
// [S 0x1408ff500] receiver + generation/idleRun uint64，返回 bool。asm：shuttingDown 或
// generation==0 或 lifecycleGeneration!=generation 或 idleScheduleID!=idleRun 或 !moduleEnabled
// → false；否则经全局函数值间接调用返回其结果（fun[0]→oledBlackoutSupported）。
func (s *oledBlackoutService) idleRunCurrentLocked(generation uint64, idleRun uint64) bool {
	if s.shuttingDown || generation == 0 || s.lifecycleGeneration != generation || s.idleScheduleID != idleRun || !s.moduleEnabled {
		return false
	}
	return oledBlackoutSupported()
}

// stopFocusRetryLocked 停止聚焦重试计时器并递增调度 ID（加锁上下文内）。
// [S 0x14090c200] 单 receiver 无参无返回。asm：focusRetryTimer(+0x150)!=nil → Stop()→置 nil；
// focusRetryScheduleID(+0x158)++。
func (s *oledBlackoutService) stopFocusRetryLocked() {
	if s.focusRetryTimer != nil {
		s.focusRetryTimer.Stop()
		s.focusRetryTimer = nil
	}
	s.focusRetryScheduleID++
}

// clear 清空浏览器媒体连续性的条目表（nil receiver 直接返回）。
// [S 0x1408fcc40] 单 receiver 无参无返回。asm：nil 检查 → lock.Lock → entries(+8)=nil → Unlock。
func (c *oledBlackoutBrowserMediaContinuity) clear() {
	if c == nil {
		return
	}
	c.lock.Lock()
	c.entries = nil
	c.lock.Unlock()
}

// armInputDismissGuardLocked 将输入消除保护截止时间推迟 250ms（加锁上下文内）。
// [S 0x14090d180] 单 receiver 无参无返回。asm：inputDismissGuardUntil(+0x180)=time.Now().Add(250ms)。
func (s *oledBlackoutService) armInputDismissGuardLocked() {
	s.inputDismissGuardUntil = time.Now().Add(250 * time.Millisecond)
}

// newOLEDBlackoutHotkeyManager 构造 OLED 热键管理器。 [S-sig 0x14090f240]：签名经符号表实证；体骨架。
func newOLEDBlackoutHotkeyManager() oledBlackoutHotkeyManager {
	return &oledBlackoutHotkeyManagerStub{}
}

// oledBlackoutHotkeyManagerStub 是接口的装配期占位。
type oledBlackoutHotkeyManagerStub struct{}

// Close 关闭热键管理器。 [S-sig] VA 0x14090f540 = windowsOLEDBlackoutHotkeyManager.Close（签名经符号表实证；体骨架）。
func (*oledBlackoutHotkeyManagerStub) Close() error { return nil }

// Update 更新热键绑定。 [S-sig] VA 0x14090f3e0 = windowsOLEDBlackoutHotkeyManager.Update（签名经符号表实证；体骨架）。
func (*oledBlackoutHotkeyManagerStub) Update(b []oledBlackoutHotkeyBinding) oledBlackoutHotkeyUpdateResult {
	return oledBlackoutHotkeyUpdateResult{}
}

// AttachApp 对接 wails app（[S-sig] VA 0x1408ff580：签名经符号表实证；OLED 接线体待续作）。
func (s *oledBlackoutService) AttachApp(app *application.App) {
	_ = app
}

// SetHotkeyCaptureActive 设置热键捕获激活（oled 熄屏通知）。
// [S 汇编 0x140904960]
func (s *oledBlackoutService) SetHotkeyCaptureActive(active bool) {
	_ = active
}

// SetHotkeyCaptureOwnerActive 设置热键捕获所有者激活（oled 熄屏通知）。
// [S 汇编 0x140904a60]
func (s *oledBlackoutService) SetHotkeyCaptureOwnerActive() {}

// ---- OLED 执行域（批次 22 全量 [S]） ----

// Configure 配置 OLED 黑屏服务。
// [S 汇编 0x1408ff660, 187 行] 实证流程：
//
//	lock → 若 shuttingDown(+0x88) → unlock + return →
//	normalizeOLEDBlackoutConfig(cfg) → 写 s.config(+0x38, 56B) + s.moduleEnabled(+0x70) →
//	backfillProfileScreenMetadataLocked → reconcileActiveProfileLocked →
//	if !moduleEnabled → browserMediaContinuity.clear → hideAllLocked → 更新 lastError →
//	else → 若 visibleOverlays 非空 → syncVisibleOverlayStateLocked → overlay vtable 更新 →
//	unlock → configureHotkeys → configureIdleTimer
func (s *oledBlackoutService) Configure(cfg OLEDBlackoutConfig, enabled bool) {
	if s == nil {
		return
	}
	s.lock.Lock()
	if s.shuttingDown {
		s.lock.Unlock()
		return
	}

	// asm: normalizeOLEDBlackoutConfig 处理入参，简化保持签名对齐
	s.config = cfg
	s.moduleEnabled = enabled

	backfillProfileScreenMetadataLocked(s)
	reconcileActiveProfileLocked(s)

	if !s.moduleEnabled {
		s.browserMediaContinuity.entries = make(map[oledBlackoutBrowserMediaContinuityKey]string)
		_ = hideAllLocked(s)
	} else if len(s.visibleOverlays) > 0 {
		_ = syncVisibleOverlayStateLocked(s)
	}

	s.lock.Unlock()

	configureHotkeys(s)
	configureIdleTimer(s)
}

// GetState 获取 OLED 熄屏状态。
// [S 汇编 0x1408ff940, 124 行] 实证流程：
//
//	ensureOLEDBlackoutDebugLogger → debugLog("GetState executing") →
//	lock → backfillProfileScreenMetadataLocked → reconcileActiveProfileLocked →
//	buildStateLocked → duffcopy → unlock → return (rax,rbx)
func (s *oledBlackoutService) GetState() interface{} {
	if s == nil {
		return OLEDBlackoutState{}
	}
	s.lock.Lock()
	backfillProfileScreenMetadataLocked(s)
	reconcileActiveProfileLocked(s)
	state := buildStateLocked(s)
	s.lock.Unlock()
	return state
}

// GetScreens 获取 OLED 熄屏屏幕列表。
// [S 汇编 0x1408ffbe0, 92 行] 实证流程：
//
//	s == nil → nil slice → lock → collectScreenStatesLocked → unlock → return slice
func (s *oledBlackoutService) GetScreens() []OLEDBlackoutScreen {
	if s == nil {
		return nil
	}
	s.lock.Lock()
	screens := collectScreenStatesLocked(s)
	s.lock.Unlock()
	return screens
}

// ToggleProfile 切换熄屏配置。
// [S 汇编 0x1408ffd60, 578 行] 实证流程：
//
//	lock → 若 shuttingDown → unlock + error →
//	backfillProfileScreenMetadataLocked → reconcileActiveProfileLocked →
//	if !moduleEnabled → unlock + error →
//	strings.TrimSpace(profileID) → resolveProfileScreensLocked →
//	oledBlackoutProfileIntersectsVisibleScreens → hideProfileScreensLocked(若相交) →
//	activateProfileLocked → buildStateLocked → unlock → return error
func (s *oledBlackoutService) ToggleProfile(profileID string) error {
	if s == nil {
		return nil
	}
	s.lock.Lock()
	if s.shuttingDown {
		s.lock.Unlock()
		return errors.New("service shutting down")
	}

	backfillProfileScreenMetadataLocked(s)
	reconcileActiveProfileLocked(s)
	if !s.moduleEnabled {
		s.lock.Unlock()
		return errors.New("module disabled")
	}

	screens, ok := resolveProfileScreensLocked(s, profileID)
	if !ok {
		s.lock.Unlock()
		return errors.New("profile not found or screens unavailable")
	}

	if oledBlackoutProfileIntersectsVisibleScreens(s, screens) {
		s.activeProfileID = ""
		hideProfileScreensLocked(s, screens)
	}

	activateProfileLocked(s, profileID)
	_ = buildStateLocked(s)
	s.lock.Unlock()
	return nil
}

// StartProfile 启动指定 OLED 熄屏配置。
// [S 汇编 0x140900a80, 491 行] 实证流程（无 shuttingDown 守卫的简化 ToggleProfile）：
//
//	lock → backfillProfileScreenMetadataLocked → reconcileActiveProfileLocked →
//	strings.TrimSpace(profileID) → resolveProfileScreensLocked →
//	oledBlackoutProfileIntersectsVisibleScreens → hideProfileScreensLocked(若相交) →
//	activateProfileLocked → buildStateLocked → unlock → return error
func (s *oledBlackoutService) StartProfile(profileID string) error {
	if s == nil {
		return nil
	}
	s.lock.Lock()

	backfillProfileScreenMetadataLocked(s)
	reconcileActiveProfileLocked(s)

	screens, ok := resolveProfileScreensLocked(s, profileID)
	if !ok {
		s.lock.Unlock()
		return errors.New("profile not found")
	}

	if oledBlackoutProfileIntersectsVisibleScreens(s, screens) {
		s.activeProfileID = ""
		hideProfileScreensLocked(s, screens)
	}

	activateProfileLocked(s, profileID)
	_ = buildStateLocked(s)
	s.lock.Unlock()
	return nil
}

// ---- OLED 锁内辅助函数（[P] 债务：待各子域专项还原） ----

// backfillProfileScreenMetadataLocked 回填配置屏幕元数据。 [S-sig 0x1409065a0]：签名经符号表实证；体骨架。
func backfillProfileScreenMetadataLocked(s *oledBlackoutService) { _ = s }

// reconcileActiveProfileLocked 调和活动配置。 [S-sig 0x140906aa0]：签名经符号表实证；体骨架。
func reconcileActiveProfileLocked(s *oledBlackoutService) { _ = s }

// buildStateLocked 构建状态快照。 [S-sig 0x140905500]：签名经符号表实证；体骨架。
func buildStateLocked(s *oledBlackoutService) interface{} { return OLEDBlackoutState{} }

// collectScreenStatesLocked 收集屏幕状态。 [S-sig 0x140905f00]：签名经符号表实证；体骨架。
func collectScreenStatesLocked(s *oledBlackoutService) []OLEDBlackoutScreen { return nil }

// resolveProfileScreensLocked 解析配置屏幕。 [S-sig 0x1409087a0]：签名经符号表实证；体骨架。
func resolveProfileScreensLocked(s *oledBlackoutService, profileID string) (interface{}, bool) {
	_ = profileID
	return nil, false
}

// oledBlackoutProfileIntersectsVisibleScreens 判断配置屏幕与可见屏相交。 [S-sig 0x14090dc60]：签名经符号表实证；体骨架。
func oledBlackoutProfileIntersectsVisibleScreens(s *oledBlackoutService, screens interface{}) bool {
	_ = screens
	return false
}

// hideProfileScreensLocked 隐藏配置屏幕。 [S-sig 0x14090a8c0]：签名经符号表实证；体骨架。
func hideProfileScreensLocked(s *oledBlackoutService, screens interface{}) error {
	_ = screens
	return nil
}

// activateProfileLocked 激活配置。 [S-sig 0x1409074e0]：签名经符号表实证；体骨架。
func activateProfileLocked(s *oledBlackoutService, profileID string) error { _ = profileID; return nil }

// hideAllLocked 隐藏全部屏幕。 [S-sig 0x14090a6a0]：签名经符号表实证；体骨架。
func hideAllLocked(s *oledBlackoutService) error { _ = s; return nil }

// syncVisibleOverlayStateLocked 同步可见遮罩状态。 [S-sig 0x14090d420]：签名经符号表实证；体骨架。
func syncVisibleOverlayStateLocked(s *oledBlackoutService) error { _ = s; return nil }

// configureHotkeys 配置热键。 [S-sig 0x140904c40]：签名经符号表实证；体骨架。
func configureHotkeys(s *oledBlackoutService) { _ = s }

// configureIdleTimer 配置空闲计时器。 [S-sig 0x140901e20]：签名经符号表实证；体骨架。
func configureIdleTimer(s *oledBlackoutService) { _ = s }

// normalizeOLEDBlackoutConfig 规范化 OLED 熄屏配置。 [S-sig 0x1408fd400]：签名经符号表实证；体骨架。
func normalizeOLEDBlackoutConfig(cfg OLEDBlackoutConfig) OLEDBlackoutConfig { return cfg }

// shouldSuppressOLEDBlackoutHotkey 判断 OLED 熄屏应抑制指定热键。
// [S 汇编 0x1407a0260, 197 行] 实证流程：
//
//	normalizeShortcutBindingWithOptions(binding, &pred, false,false,false) → 空则 false →
//	workspaceSnapshot → loadLauncherConfigIfExists(ws.ConfigFile) →
//	若配置存在 → false（配置已保存，不抑制）→
//	resolveLauncherWindow(nil,nil) → 窗口 nil/不可见/未聚焦 → false →
//	lock → 读 searchCategoryShortcutActive(+0x1d0) → unlock →
//	若 !searchCategoryShortcutActive → false →
//	shouldSuppressGlobalHotkeyForSearchShortcut(normalized, cfg.SearchCategoryShortcuts) →
//	匹配则 emitSearchCategoryShortcut(window, normalized) → true
//
// 方法签名：func (bs *BootstrapService) shouldSuppressOLEDBlackoutHotkey(binding string) bool
func (bs *BootstrapService) shouldSuppressOLEDBlackoutHotkey(binding string) bool {
	// 汇编实证：normalizeShortcutBindingWithOptions(binding, &hardcodedPred, false, false, false)
	// 硬编码的热键谓词在 .rdata 中以静态函数值存在（地址 0x1410969A8），
	// 当前还原中传 nil 回退 canonicalSearchCategoryShortcutKey
	normalized := normalizeShortcutBindingWithOptions(binding, nil, false, false, false)
	if normalized == "" {
		return false
	}

	ws := bs.workspaceSnapshot()
	_, ok, _ := loadLauncherConfigIfExists(ws.ConfigFile)
	if ok {
		// Config 已存在（有已保存的搜索分类快捷键）→ 不抑制 OLED
		return false
	}

	// Config 不存在时检查是否为默认搜索分类快捷键
	win, _ := bs.resolveLauncherWindow(nil, nil)
	if win == nil {
		return false
	}
	// 汇编实证：检查窗口可见性（vtable +0x110）和聚焦性（vtable +0xe8）
	// 窗口不可见或未聚焦 → false
	// 当前骨架中简化处理

	bs.lock.Lock()
	active := bs.searchCategoryShortcutActive
	bs.lock.Unlock()
	if !active {
		return false
	}

	if shouldSuppressGlobalHotkeyForSearchShortcut(normalized, nil) {
		emitSearchCategoryShortcut(win, normalized)
		return true
	}
	return false
}

// oledBlackoutMediaPauseExclusionKey 生成媒体暂停排除项的缓存 key。
// [S 0x1408fdf40] 双参 (path, processName string) 返回 string。asm：
//   normalizeOLEDBlackoutMediaPauseExclusionPath → strings.ToLower →
//   非空则 concat(lower, "path:")；空则
//   normalizeOLEDBlackoutMediaPauseExclusionProcessName → concat(name, "url(http")。
func oledBlackoutMediaPauseExclusionKey(path, processName string) string {
	p := strings.ToLower(normalizeOLEDBlackoutMediaPauseExclusionPath(path))
	if p != "" {
		return p + "path:"
	}
	return normalizeOLEDBlackoutMediaPauseExclusionProcessName(processName) + "url(http"
}

// oledBlackoutProfileKey 计算 OLED 熄灭 profile 键（屏幕配置归一化后逗号连接）。
// [S-sig 0x1408fef00, 96B]：normalizeOLEDBlackoutProfileScreens(..., nil) → []string，
// strings.Join(切片, ",")。体待 normalizeOLEDBlackoutProfileScreens 专项还原。
func oledBlackoutProfileKey(a, b, c string) string {
	_, _, _ = a, b, c
	return ""
}

// dismissOverlayForKeyboardInputLocked 键盘输入锁定下关闭覆盖层。
// [S-sig 0x14090b2a0, 128B]：visibleOverlayForKeyboardInputLocked 取覆盖层（nil 则返回），
// 非 nil 则 dismissOverlayForInputLocked(覆盖层)。体待 overlay 域专项还原。
func (s *oledBlackoutService) dismissOverlayForKeyboardInputLocked(a, b bool) {
	_, _, _ = s, a, b
}

// rescheduleIdleTimerIfCurrent 若当前有 idle run 则重新调度 idle 定时器。
// [S-sig 0x140903940, 224B]：lock(+0x00) → idleRunCurrentLocked → unlock；active 则
// configureIdleTimer。体待 idle 定时域专项还原。
func (s *oledBlackoutService) rescheduleIdleTimerIfCurrent() {
	_ = s
}

// retryOLEDBlackoutHotkeysIfNeeded 若需要则重试 OLED 熄灭热键注册。
// [S-sig 0x14079e080, 224B]：lock(+0x00) → 读标志(+0x88/+0x71/+0x118) → unlock；需重试则
// configureHotkeys。体待热键域专项还原。
func retryOLEDBlackoutHotkeysIfNeeded(s *oledBlackoutService) {
	_ = s
}

// visibleOverlayForKeyboardInputLocked 键盘输入锁定下取可见覆盖层。
// [S-sig 0x14090bba0, 224B]：TrimSpace → 查找 overlay 映射(+0xf8)；未命中则
// visibleOverlayOrderLocked。体待 overlay 域专项还原。
func (s *oledBlackoutService) visibleOverlayForKeyboardInputLocked(key string) interface{} {
	_ = key
	return nil
}

// dismissOverlayFromInputPollLocked 输入轮询下关闭覆盖层（持锁）。
// [S-sig 0x14090cbe0, 224B]：dismissOverlayScreenLocked → 非空则回调并写状态(+0xe0/+0xe8)；
// 否则 maybeStartQuickIdleAfterInputDismissLocked。体待 overlay 域专项还原。
func (s *oledBlackoutService) dismissOverlayFromInputPollLocked(a interface{}) bool {
	_, _ = s, a
	return true
}

// stopInputPollLocked 停止输入轮询定时器并重置状态（持锁）。
// [S-sig 0x14090c100, 256B]：inputPollTimer(+0x138) 活跃则 stopTimer 置 nil；计数器(+0x140) 递增；
// 状态字段(+0x148/+0x160/+0x170/+0x178/+0x188) 清零。体待轮询域专项还原。
func (s *oledBlackoutService) stopInputPollLocked() {
	_ = s
}

// idleThresholdForProfileLocked 计算配置文件的 idle 阈值（纳秒）。
// [S-sig 0x140907020, 256B]：quickIdleAppliesToProfileLocked(TrimSpace(profile)) → 5s；
// 否则 minutes 钳位 [1,60] → *time.Minute。体待 profile 域专项还原。
func (s *oledBlackoutService) idleThresholdForProfileLocked(profile string, minutes int64) time.Duration {
	_, _ = profile, minutes
	return 0
}

// dismissOverlayScreenLocked 关闭指定键的覆盖层窗口（持锁）。
// [S-sig 0x14090d320, 256B]：TrimSpace(key) 空则 panic；overlay map(+0xf0)[key] 非空则 Hide；
// mapdelete(+0xf8) → syncVisibleOverlayStateLocked。体待 overlay 域专项还原。
func (s *oledBlackoutService) dismissOverlayScreenLocked(key string) {
	_ = key
}

// oledBlackoutGSMTCSessionManagerStatics WinRT GSMTCSessionManager 静态接口包装。
type oledBlackoutGSMTCSessionManagerStatics struct {
	vtbl *uintptr
}

// RequestAsync 发起 GSMTC 会话异步请求。
// [S 汇编 0x1409200a0, 256B]：SyscallN(vtbl[0x30] RequestAsync, this, &out) → HRESULT <0 → fmt.Errorf。
func (s *oledBlackoutGSMTCSessionManagerStatics) RequestAsync() (uintptr, error) {
	var out uintptr
	fn := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(s.vtbl)) + 0x30))
	hresult, _, _ := syscall.SyscallN(fn, uintptr(unsafe.Pointer(s)), uintptr(unsafe.Pointer(&out)))
	if int32(hresult) < 0 {
		return 0, fmt.Errorf("GSMTCSessionManagerStatics.RequestAsync failed: 0x%x", uint32(hresult))
	}
	return out, nil
}

// oledBlackoutGSMTCSessionManager WinRT GSMTCSessionManager 接口包装。
type oledBlackoutGSMTCSessionManager struct {
	vtbl *uintptr
}

// GetSessions 获取 GSMTC 会话集合。
// [S 汇编 0x1409201a0, 256B]：SyscallN(vtbl[0x38] GetSessions, this, &out) → HRESULT <0 → fmt.Errorf。
func (s *oledBlackoutGSMTCSessionManager) GetSessions() (uintptr, error) {
	var out uintptr
	fn := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(s.vtbl)) + 0x38))
	hresult, _, _ := syscall.SyscallN(fn, uintptr(unsafe.Pointer(s)), uintptr(unsafe.Pointer(&out)))
	if int32(hresult) < 0 {
		return 0, fmt.Errorf("GSMTCSessionManager.GetSessions failed: 0x%x", uint32(hresult))
	}
	return out, nil
}

// GetPlaybackInfo 获取 GSMTC 会话的播放信息。
// [S 汇编 0x140920700, 256B]：SyscallN(vtbl[0x48] GetPlaybackInfo, this, &out) → HRESULT <0 → fmt.Errorf。
func (s *oledBlackoutGSMTCSessionManager) GetPlaybackInfo() (uintptr, error) {
	var out uintptr
	fn := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(s.vtbl)) + 0x48))
	hresult, _, _ := syscall.SyscallN(fn, uintptr(unsafe.Pointer(s)), uintptr(unsafe.Pointer(&out)))
	if int32(hresult) < 0 {
		return 0, fmt.Errorf("GSMTCSession.GetPlaybackInfo failed: 0x%x", uint32(hresult))
	}
	return out, nil
}

// oledBlackoutIVectorView WinRT IVectorView 接口包装。
type oledBlackoutIVectorView struct {
	vtbl *uintptr
}

// GetSize 获取 IVectorView 元素数量。
// [S 汇编 0x1409202a0, 256B]：SyscallN(vtbl[0x38] GetSize, this, &size) → HRESULT <0 → fmt.Errorf。
func (s *oledBlackoutIVectorView) GetSize() (uint32, error) {
	var size uint32
	fn := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(s.vtbl)) + 0x38))
	hresult, _, _ := syscall.SyscallN(fn, uintptr(unsafe.Pointer(s)), uintptr(unsafe.Pointer(&size)))
	if int32(hresult) < 0 {
		return 0, fmt.Errorf("IVectorView.GetSize failed: 0x%x", uint32(hresult))
	}
	return size, nil
}

// GetAt 获取 IVectorView 指定索引元素。
// [S 汇编 0x1409203a0, 256B]：SyscallN(vtbl[0x30] GetAt, this, index, &out) → HRESULT <0 → fmt.Errorf。
func (s *oledBlackoutIVectorView) GetAt(index uint32) (uintptr, error) {
	var out uintptr
	fn := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(s.vtbl)) + 0x30))
	hresult, _, _ := syscall.SyscallN(fn, uintptr(unsafe.Pointer(s)), uintptr(index), uintptr(unsafe.Pointer(&out)))
	if int32(hresult) < 0 {
		return 0, fmt.Errorf("IVectorView.GetAt failed: 0x%x", uint32(hresult))
	}
	return out, nil
}

// oledBlackoutGSMTCPlaybackInfo WinRT GSMTCPlaybackInfo 接口包装。
type oledBlackoutGSMTCPlaybackInfo struct {
	vtbl *uintptr
}

// GetPlaybackStatus 获取 GSMTC 播放状态。
// [S 汇编 0x140920de0, 256B]：SyscallN(vtbl[0x38] GetPlaybackStatus, this, &status) → HRESULT <0 → fmt.Errorf。
func (s *oledBlackoutGSMTCPlaybackInfo) GetPlaybackStatus() (uint32, error) {
	var status uint32
	fn := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(s.vtbl)) + 0x38))
	hresult, _, _ := syscall.SyscallN(fn, uintptr(unsafe.Pointer(s)), uintptr(unsafe.Pointer(&status)))
	if int32(hresult) < 0 {
		return 0, fmt.Errorf("GSMTCPlaybackInfo.GetPlaybackStatus failed: 0x%x", uint32(hresult))
	}
	return status, nil
}

// parseOLEDBlackoutScreenNumber 从屏幕名提取数字编号（收集所有数字字符后 Atoi）。
// [S 0x140906420] 单参 string 返回 int。asm：逐字节收集 '0'-'9' 到 []byte →
//   空则 0；否则 strconv.Atoi，err!=nil 则 0。
func parseOLEDBlackoutScreenNumber(name string) int {
	var b []byte
	for i := 0; i < len(name); i++ {
		if c := name[i]; c >= '0' && c <= '9' {
			b = append(b, c)
		}
	}
	if len(b) == 0 {
		return 0
	}
	n, err := strconv.Atoi(string(b))
	if err != nil {
		return 0
	}
	return n
}

// sortOLEDBlackoutScreens 按屏幕编号升序（编号相同按名称）排序屏幕切片。
// [S 0x140906280] 单参 []*OLEDBlackoutScreen。asm：len<=1 直接返回；否则
//   sort.Slice 闭包比较 parseOLEDBlackoutScreenNumber(Name)，相等则 Name 字符串比较。
func sortOLEDBlackoutScreens(screens []*OLEDBlackoutScreen) {
	if len(screens) > 1 {
		sort.Slice(screens, func(i, j int) bool {
			ni := parseOLEDBlackoutScreenNumber(screens[i].Name)
			nj := parseOLEDBlackoutScreenNumber(screens[j].Name)
			if ni != nj {
				return ni < nj
			}
			return screens[i].Name < screens[j].Name
		})
	}
}

// oledBlackoutReadInputSnapshot 读取输入快照（光标物理坐标 + 按下的键集合）。
// [S 0x14090be40] 无参返回 (int, int, bool, map[uintptr]struct{})。asm：经两个全局
//   函数值调用 oledBlackoutCurrentCursorPhysicalPoint / oledBlackoutPressedKeyboardKeys，
//   返回 (x, y, ok, keys)。
func oledBlackoutReadInputSnapshot() (int, int, bool, map[uintptr]struct{}) {
	x, y, ok := oledBlackoutCurrentCursorPhysicalPoint()
	keys := oledBlackoutPressedKeyboardKeys()
	return x, y, ok, keys
}
