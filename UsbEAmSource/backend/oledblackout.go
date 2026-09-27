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
	"path/filepath"
	"strings"
	"time"

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
