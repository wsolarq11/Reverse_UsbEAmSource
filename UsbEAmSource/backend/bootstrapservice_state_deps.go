// AUTO-RECONSTRUCTED SERVICE METHODS — DOMAIN: state assembly dependencies
// 研究用途 · UsbEAm Launcher 1.0.3 后端方法体还原
//
// 本文件承载 bootstrapservice_state.go 与 bootstrapservice_lifecycle.go 的被调方。
// 档位标注：
//
//	[S]     反汇编实证体（本文件内已还原）
//	[S-sig] 仅签名实证，体待对应专项域还原（[P]）
//
// 这些符号在目标二进制中均存在（VA 见各条注释），在重建树中此前缺失或签名不符，
// 本轮按实证补齐签名以闭合状态装配链。
package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ---- 状态装配域被调方 ----

// isStartupTrayMode 读取启动托盘模式标志（锁保护）。
// [S 汇编 0x1407966e0, 256B]：
//
//	0x140796703  bs == nil → false（0x140796783 出口，eax=0）
//	0x140796713  lock(+0x540)（cmpxchg 快路径）+ defer unlock（deferwrap lea [rip+0x7f]）
//	0x140796761  读 startupTrayMode(+0x449) → [rsp+0xe]
//	0x140796771  defer 展开后 movzx eax,[rsp+0xe] 返回
func (bs *BootstrapService) isStartupTrayMode() bool {
	if bs == nil {
		return false
	}
	bs.lock.Lock()
	defer bs.lock.Unlock()
	return bs.startupTrayMode
}

// fileExists 判断路径是否存在且为普通文件。
// [S-sig 汇编 0x1407a1a00]：GetStartupState 调用点 0x140775b4f 以
// test rbx,rbx 判错、test al,al 取存在位 —— 返回 (bool, error)。
// [P] 体（是否区分目录、是否吞 ErrNotExist）待文件系统域专项实证。
func fileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil
}

// launcherConfigRevision 计算配置修订标识。
// [S-sig 汇编 0x1408771e0]：入参 LauncherConfig 走调用者栈，返回 string（rax=ptr, rbx=len）；
// 调用点 launcherStateFromCommittedConfigWithSnapshot 0x1407769e0 与
// launcherStateFromSavedConfig 0x140776c46 均把结果写入 cfg.Revision（ptr@+0, len@+8）。
// [P] 计算规则（内容哈希 / 单调序号 / 时间戳）待配置域专项实证；
// 当前返回入参原值，保证不破坏调用者已有修订标识。
func launcherConfigRevision(cfg LauncherConfig) string {
	return cfg.Revision
}

// cacheBookmarkSources 缓存书签源集合。
// [S 汇编 0x14079cb80, 107 行]：调用 normalizeBookmarkSources 规范化后，在锁保护下
// 用 reflect.DeepEqual 比较旧源与新源，无变化则早返；变化则写字段后调用 resetBookmarkIconCache。
func (bs *BootstrapService) cacheBookmarkSources(sources []BookmarkSource) {
	if bs == nil {
		return
	}
	normalized := normalizeBookmarkSources(sources)

	bs.lock.Lock()
	if reflect.DeepEqual(bs.bookmarkSources, normalized) {
		bs.lock.Unlock()
		return
	}
	bs.bookmarkSources = normalized
	bs.lock.Unlock()
	resetBookmarkIconCache()
}

// attachLauncherConfigIconURLs 为配置内的图标资源附加 URL。
// [S asm 0x1408a1ec0, 240L] 遍历 cfg.Apps（stride 0x168）每个两图标槽：
//
//	Slot1（AutoIcon 控制）：IconData + IconRef → buildLauncherConfigIconResource("icon/app")
//	→ 写 IconRef/IconURL，清 IconData；无源仅清 IconData。
//	Slot2（无条件）：CustomIconRef/IconData → build → 写 CustomIconRef/IconURL，清 CustomIconData；
//	无源仅清 CustomIconData。
//	两槽的 IconURL 共享同一字段（slot2 覆盖 slot1）。
func (bs *BootstrapService) attachLauncherConfigIconURLs(cfg *LauncherConfig) {
	if bs == nil || cfg == nil {
		return
	}
	for i := range cfg.Apps {
		a := &cfg.Apps[i]

		// ---- Slot 1：AutoIcon 控制 ----
		if a.AutoIcon && launcherConfigIconSlotHasSource(a.IconData, a.IconRef) {
			res := bs.buildLauncherConfigIconResource("icon/app", a.IconRef, a.IconData)
			a.IconRef = res.IconRef
			a.IconURL = res.IconURL
		}
		// asm：IconData 恒清零（无论是否有源）
		a.IconData = ""

		// ---- Slot 2：CustomIcon（无条件，可覆盖 IconURL） ----
		if launcherConfigIconSlotHasSource(a.CustomIconData, a.CustomIconRef) {
			// asm：清零 IconURL 在先（清除旧槽1 URL）
			a.IconURL = ""
			res := bs.buildLauncherConfigIconResource("icon/app", a.CustomIconRef, a.CustomIconData)
			a.CustomIconRef = res.IconRef
			a.IconURL = res.IconURL
		}
		a.CustomIconData = ""
	}
	// 非 App 槽（SpeedDial/Bookmarks/TagCatalog/ConsoleItems/MouseGestures/OLEDBlackout/etc.）由
	// visitLauncherConfigIconSlots + namespace mapper 处理，当前 [P] 待 slots 签名对齐后扩充。
}

// attachLauncherBackgroundURL 为配置附加背景资源 URL。
// [S 汇编实证 0x140798600, 387L asm]（attachLauncherBackgroundURL.asm.txt）：
//
//	bs/cfg nil 守卫 → normalized = normalizeBackgroundPreference(cfg.Background)
//	→ TrimSpace(ImagePath) 空 → 写回 normalized 返回
//	→ svc = screenshotAssetService()（读 bs.assets）nil → 写回返回
//	→ cleaned = filepath.Clean(TrimSpace(ImagePath)) → os.Stat → err/IsDir/Size≤0 → 写回返回
//	→ modifiedAt = ModTime().UnixNano()（asm：itab fun[1]→nano 时间变换，bt 0x3f / imul 1e9 /
//	and 0x3fffffff / 常量 0xdd7b17f80=wallToInternal、0xa1b203eb3d1a0000=internalToUnix×1e9）
//	→ backgroundAssetLock.Lock + defer Unlock
//	→ 缓存命中（owner==svc && path==cleaned && size 相等 && modifiedAt 相等）
//	且 TrimSpace(URL) 非空且 svc.Exists(URL, "background/custom") → 复用缓存 URL 写回返回
//	→ 未命中：RegisterFile("background/custom", cleaned, launcherBackgroundContentTypeForPath(cleaned), 0)
//	→ err==nil 更新 backgroundAssetOwner/Path/Size/ModifiedAt/URL 与 normalized.ImageURL
//	→ 写回 normalized。
func (bs *BootstrapService) attachLauncherBackgroundURL(cfg *LauncherConfig) {
	if bs == nil || cfg == nil {
		return
	}
	normalized := normalizeBackgroundPreference(cfg.Preferences.Background)
	if strings.TrimSpace(normalized.ImagePath) == "" {
		cfg.Preferences.Background = normalized
		return
	}
	svc := bs.screenshotAssetService()
	if svc == nil {
		cfg.Preferences.Background = normalized
		return
	}
	cleaned := filepath.Clean(strings.TrimSpace(normalized.ImagePath))
	info, err := os.Stat(cleaned)
	if err != nil || info.IsDir() {
		cfg.Preferences.Background = normalized
		return
	}
	size := info.Size()
	if size <= 0 {
		cfg.Preferences.Background = normalized
		return
	}
	modifiedAt := info.ModTime().UnixNano()

	bs.backgroundAssetLock.Lock()
	defer bs.backgroundAssetLock.Unlock()

	if bs.backgroundAssetOwner == svc &&
		bs.backgroundAssetPath == cleaned &&
		bs.backgroundAssetSize == size &&
		bs.backgroundAssetModifiedAt == modifiedAt {
		if url := strings.TrimSpace(bs.backgroundAssetURL); url != "" && svc.Exists(url, "background/custom") {
			normalized.ImageURL = bs.backgroundAssetURL
			cfg.Preferences.Background = normalized
			return
		}
	}

	ref, err := svc.RegisterFile("background/custom", cleaned, launcherBackgroundContentTypeForPath(cleaned), 0)
	if err == nil {
		bs.backgroundAssetOwner = svc
		bs.backgroundAssetPath = cleaned
		bs.backgroundAssetSize = size
		bs.backgroundAssetModifiedAt = modifiedAt
		bs.backgroundAssetURL = ref.URL
		normalized.ImageURL = ref.URL
	}
	cfg.Preferences.Background = normalized
}

// syncRuntimeServices 同步运行时服务。
// [S 汇编 0x14079f060, 276 行]：从当前配置的 Preferences.FeatureModules map 合并默认
// 启用表（18 项，经 .data 解码实证）后逐项传递给 5 个服务的 Configure 方法。
func (bs *BootstrapService) syncRuntimeServices() {
	if bs == nil {
		return
	}

	// 读当前配置（FeatureModules 合并见 launcherFeatureModulesFromConfig：18 项默认全 true，
	// 后被 Preferences.FeatureModules 覆盖）。
	cfg, err := bs.readCurrentConfig()
	if err != nil {
		return
	}
	enabledFeatures := launcherFeatureModulesFromConfig(cfg)

	// -- 缓存截图偏好 --
	bs.cacheScreenshotCapturePreferences(cfg.Preferences)

	// -- 逐服务配置 --

	// memoryRelease
	if bs.memoryRelease != nil {
		bs.memoryRelease.Configure(cfg.MemoryRelease, enabledFeatures["memoryRelease"])
	}

	// oledBlackout
	if bs.oledBlackout != nil {
		bs.oledBlackout.Configure(cfg.OLEDBlackout, enabledFeatures["oledBlackout"])
	}

	// windowManagement（同时检查截图 feature 状态）
	if bs.windowManagement != nil {
		windowMgtEnabled := enabledFeatures["windowManagement"] || enabledFeatures["screenshot"]
		bs.windowManagement.Configure(cfg.WindowManagement, windowMgtEnabled)
	}

	// mouseGestures
	if bs.mouseGestures != nil {
		bs.mouseGestures.Configure(cfg.MouseGestures, enabledFeatures["mouseGestures"])
	}

	// desktopWidgets（Open + ReconcilePendingDeletes + markDegraded + Configure 链）
	if bs.desktopWidgets != nil {
		deskEnabled := enabledFeatures["desktopWidgets"]
		bs.desktopWidgets.Open(deskEnabled)
		if err := bs.desktopWidgets.ReconcilePendingDeletes(); err != nil {
			bs.desktopWidgets.markDegraded(err)
		}
		bs.desktopWidgets.Configure(deskEnabled)
	}

	// -- 文件搜索运行时 --
	bs.syncFileSearchRuntime(enabledFeatures, cfg)

	// -- 光标包裹/热角守卫 --
	bs.syncCursorWrapHotCornerGuard(enabledFeatures, cfg.MouseGestures.HotCorners)
}

// ---- 同步运行时辅助 ----

// readCurrentConfig 读取当前配置。
// [S-sig] 辅助桩：从 configStore 读取当前配置。
func (bs *BootstrapService) readCurrentConfig() (LauncherConfig, error) {
	if bs == nil || bs.configStore == nil {
		return LauncherConfig{}, nil
	}
	cfg, err := bs.configStore.Read()
	if err != nil {
		return LauncherConfig{}, err
	}
	return cfg, nil
}

// cacheScreenshotCapturePreferences 缓存截图捕获偏好设置。
// [S 汇编 0x14079f640, 117 行]：接收 Preferences 参数，调 normalizePreferencesWithOptions 归一化后，
// 从结果中提取三个 *bool 字段（ScreenshotCaptureControls/cursor/SelectionConfirm），
// 锁保护写入 bs 缓存字段。*bool nil 处理：captureControls 默认 true，cursor/selectionConfirm 默认 false。
func (bs *BootstrapService) cacheScreenshotCapturePreferences(prefs Preferences) {
	if bs == nil {
		return
	}

	// 归一化 Preferences（opts = nil）
	normalized := normalizePreferencesWithOptions(prefs, nil)

	// 提取三个 *bool 字段，nil 处理
	captureControls := true
	if normalized.ScreenshotCaptureControls != nil {
		captureControls = *normalized.ScreenshotCaptureControls
	}
	captureCursor := false
	if normalized.ScreenshotCaptureCursor != nil {
		captureCursor = *normalized.ScreenshotCaptureCursor
	}
	selectionConfirm := false
	if normalized.ScreenshotSelectionConfirm != nil {
		selectionConfirm = *normalized.ScreenshotSelectionConfirm
	}

	bs.lock.Lock()
	bs.screenshotCaptureControlsSetting = captureControls
	bs.screenshotCaptureCursorSetting = captureCursor
	bs.screenshotSelectionConfirmSetting = selectionConfirm
	bs.screenshotCaptureSettingsReady = true
	bs.lock.Unlock()
}

// shouldWarmResidentFileSearchRuntime 判定是否应预热常驻文件搜索运行时。
// [S 汇编 0x14079f9a0, 0x120]：
//
//	enabled 非 nil 且 !enabled["files"] → false
//	normalized := normalizeFileSearchConfig(cfg)
//	!normalized.Enabled → false
//	normalized.ResourceMode != "resident"（len 8 常量）→ false
//	否则 true
func shouldWarmResidentFileSearchRuntime(enabled map[string]bool, cfg FileSearchConfig) bool {
	if enabled != nil && !enabled["files"] {
		return false
	}
	normalized := normalizeFileSearchConfig(cfg)
	if !normalized.Enabled {
		return false
	}
	return normalized.ResourceMode == "resident"
}

// syncFileSearchRuntime 同步文件搜索运行时。
// [S 汇编 0x14079f820, 0x154]：warm = shouldWarmResidentFileSearchRuntime(enabled, cfg.FileSearch) →
// 锁内读 fileIndex：!warm 清 pending；warm 且 idx==nil 置 pending（warm 且 idx!=nil 保持原样）→
// 解锁后 warm 且 idx!=nil → idx.scheduleFileSearchMaintenance(nil, true)。
func (bs *BootstrapService) syncFileSearchRuntime(enabled map[string]bool, cfg LauncherConfig) {
	if bs == nil {
		return
	}

	warm := shouldWarmResidentFileSearchRuntime(enabled, cfg.FileSearch)

	bs.lock.Lock()
	idx := bs.fileIndex
	if !warm {
		bs.pendingFileSearchResidentWarm = false
	} else if idx == nil {
		bs.pendingFileSearchResidentWarm = true
	}
	bs.lock.Unlock()

	if warm && idx != nil {
		idx.scheduleFileSearchMaintenance(nil, true)
	}
}

// syncCursorWrapHotCornerGuard 同步光标包裹/热角守卫。
// [S 汇编 0x14079fb60, 0x18a]：windowManagement 非 nil →
// normalizeHotCornerConfig(hotCorner) → guardPx 默认 0 →
// enabled["mouseGestures"] 且 normalized.Enabled 且 hasEnabledHotCornerRule → guardPx = TriggerSizePx →
// SetCursorWrapCornerGuard(guardPx)。
func (bs *BootstrapService) syncCursorWrapHotCornerGuard(features map[string]bool, hotCorner HotCornerConfig) {
	if bs == nil || bs.windowManagement == nil {
		return
	}

	normalized := normalizeHotCornerConfig(hotCorner)

	guardPx := 0
	if features["mouseGestures"] && normalized.Enabled && hasEnabledHotCornerRule(normalized) {
		guardPx = normalized.TriggerSizePx
	}
	bs.windowManagement.SetCursorWrapCornerGuard(guardPx)
}

// normalizePreferencesWithOptions 规范化偏好设置（带选项）。
// [S-sig] 骨架桩。VA 0x14087cb00。签名为 (Preferences, *LauncherConfigOptions) Preferences。
func normalizePreferencesWithOptions(prefs Preferences, opts *LauncherConfigOptions) Preferences {
	_ = opts
	return prefs
}

// ---- 关闭链被调方 ----

// shutdownScreenshotCOMQueryWorkers 停止截图 COM 查询工作线程。
// [S-sig 汇编 0x1409b71a0]：Shutdown 调用点 0x140773e87，无实参、无返回值。
// [P] 体待截图 COM 域。
func shutdownScreenshotCOMQueryWorkers() {}

// screenshotDXGIOutputCaptureCache 及其 Close 已由批次 54 在
// screenshot_dxgi_cache_windows.go 落地（真类型 + 单例 + Close）。
// 此处不再保留占位，避免重复定义。

// ---- 服务关闭链（asm Shutdown 0x140773e60 逐项调用的 12 个实符号） ----
//
// 以下 Shutdown 在目标二进制中均为真实符号（调用点 VA 见 bootstrapservice_lifecycle.go），
// 重建树此前缺失，本轮按 [S-sig] 补齐签名以保证关闭链闭合；体待各专项域。
// 实证要点：asm 中每个调用的返回值一律不检查（调用后直接推进下一项），故均无返回值。

// Shutdown 停止 OLED 黑屏服务。[S-sig 调用点 0x140774082]
func (s *oledBlackoutService) Shutdown() {
	_ = s
}

// Shutdown 停止鼠标手势服务。[S-sig 调用点 0x1407740d3]
func (s *mouseGestureService) Shutdown() {}

// Shutdown 关闭截图钉图窗口服务。[S-sig 调用点 0x1407740f4]
func (s *screenshotPinWindowService) Shutdown() {}

// Shutdown 关闭截图预览窗口服务。[S-sig 调用点 0x140774110]
func (s *screenshotPreviewWindowService) Shutdown() {}

// Shutdown 关闭截图选区工具条服务。[S-sig 调用点 0x140774128]
func (s *screenshotSelectionToolbarWindowService) Shutdown() {}

// Shutdown 关闭插件窗口服务。[S-sig 调用点 0x14077413a]
func (s *pluginWindowService) Shutdown() {}

// AttachApp 对接 wails 应用（加锁写 app 字段）。
// [S 汇编 0x14092d4c0, 160B]：lock(+0x00).Lock + 写屏障写 app(+0x08) + Unlock。
func (s *pluginWindowService) AttachApp(app *application.App) {
	s.lock.Lock()
	s.app = app
	s.lock.Unlock()
}

// CloseAudience 关闭插件窗口观众（windows[id].audience.close）。
// [S-sig 0x140930080, 288B]：nil 早退；TrimSpace(id)；lock → windows(+0x10)[id] → unlock；
// win 非空且 audience(+0x88) 非空则调 close(+0x30)。体待 pluginManagedWindow 域专项还原。
func (s *pluginWindowService) CloseAudience(id string) {
	if s == nil {
		return
	}
	id = strings.TrimSpace(id)
	s.lock.Lock()
	win := s.windows[id]
	s.lock.Unlock()
	_ = win
}

// Close 关闭插件窗口（windows[id].audience.close，空 id 报错）。
// [S-sig 0x14092f500, 320B]：TrimSpace 空→error；lock → windows(+0x10)[id] → unlock；
// win 非空且 audience(+0x88) 非空则调 close(+0x30)。体待 pluginManagedWindow 域专项还原。
func (s *pluginWindowService) Close(id string) error {
	if s == nil {
		return nil
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}
	s.lock.Lock()
	win := s.windows[id]
	s.lock.Unlock()
	_ = win
	return nil
}

// store 存储插件窗口（windows[key] 已存在且相同则更新，shutting 则跳过）。
// [S-sig 0x140930300, 320B]：lock → shutting(+0x28)==0 且 windows(+0x10)[key]==win →
// mapassign → unlock。体待 pluginManagedWindow 域专项还原。
func (s *pluginWindowService) store(key string, win *pluginManagedWindow) {
	if s == nil {
		return
	}
	s.lock.Lock()
	if !s.shutting && s.windows[key] == win {
		s.windows[key] = win
	}
	s.lock.Unlock()
}

// release 释放插件窗口（gen/identity 匹配才删除 windows[key]）。
// [S-sig 0x1409301a0, 352B]：lock → windows(+0x10)[key] 匹配 gen(+0x88)+identity(+0x90) →
// mapdelete → unlock。体待 pluginManagedWindow 域专项还原。
func (s *pluginWindowService) release(key string, gen uintptr, identity interface{}) {
	if s == nil {
		return
	}
	s.lock.Lock()
	if win := s.windows[key]; win != nil {
		_, _, _ = key, gen, identity
		delete(s.windows, key)
	}
	s.lock.Unlock()
}

// resolvePluginWindowScreen 解析插件窗口所在屏幕（按名字匹配 ScreenManager）。
// [S-sig 0x1409316c0, 352B]：screenManager(+0x310).GetAll → 名字空→GetPrimary；
// EqualFold(TrimSpace) 匹配 → 返回 screen。体待屏幕域专项还原。
func resolvePluginWindowScreen(a interface{}, name string) interface{} {
	_, _ = a, name
	return nil
}

// CloseAudienceOwned 关闭观众拥有的插件窗口（持锁，EqualFold 匹配 owner 后 close）。
// [S-sig 0x14092fa20, 416B]：TrimSpace×2 → lock(+0x0) → windows(+0x10)[owner] →
// window(+0x88) nil→nil；EqualFold → close 回调。体待窗口域专项还原。
func (s *pluginWindowService) CloseAudienceOwned(a, b string) error {
	_, _, _ = s, a, b
	return nil
}

// resolvePluginPackageURL 解析插件包 URL。
// [S-sig 0x14092c520, 224B]：TrimSpace 分支 → resolvePluginCatalogAssetURL / normalizePluginPackageFile。
// 体待 catalog 域专项还原。
func resolvePluginPackageURL(a, b, c, d, e string) string {
	_, _, _, _, _ = a, b, c, d, e
	return ""
}

// resolvePluginWindowBounds 解析插件窗口边界。
// [S-sig 0x140931820, 224B]：覆盖值非空则透传；否则居中/钳位算术（4 字返回）。
// 体待 window 域专项还原。
func resolvePluginWindowBounds(a interface{}, b, c int64) (int64, int64, int64, int64) {
	_, _, _ = a, b, c
	return 0, 0, 0, 0
}

// Shutdown 停止输入监视服务。[S-sig 调用点 0x140774149]
func (s *inputMonitorService) Shutdown() {}

// Shutdown 停止文件定位服务。[S-sig 调用点 0x140774158]
func (s *fileLocatorService) Shutdown() {}

// Shutdown 停止桌面小部件服务。[S-sig 调用点 0x140774167]
func (s *desktopWidgetService) Shutdown() {}

// normalizePluginWindowLogicalName 归一化插件窗口逻辑名（去空白、空则回退、截断 128 rune）。
// [S 汇编 0x140930f00, 288B]：TrimSpace(name) 空→TrimSpace(fallback)；仍空→默认（6B）；
// countrunes>128→截断到 128 rune。
func normalizePluginWindowLogicalName(name, fallback string) string {
	s := strings.TrimSpace(name)
	if s == "" {
		s = strings.TrimSpace(fallback)
	}
	if s == "" {
		s = "window"
	}
	runes := []rune(s)
	if len(runes) > 128 {
		runes = runes[:128]
	}
	return string(runes)
}

// launcherConfigSnapshotsEqual 比较两个配置快照是否相等（字段 + revision）。
// [S-sig 0x14089cc00, 288B]：bool 标志不等→false；标志皆 false→true；字段字节比较 →
// launcherConfigRevision 两次→revision 相等则 memequal。体待配置快照结构专项还原。
func launcherConfigSnapshotsEqual(a, b interface{}, fa, fb bool) bool {
	_, _, _, _ = a, b, fa, fb
	return false
}
