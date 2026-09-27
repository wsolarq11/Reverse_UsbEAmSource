// AUTO-RECONSTRUCTED — DOMAIN: bootstrap SaveConfig 附属被调函数
// 研究用途 · UsbEAm Launcher 1.0.3 后端方法体还原
//
// 本文件承载 SaveConfig 的附属被调函数，函数体为反汇编实证（[S]）或签名骨架（[S-sig][P]）。
//
// 主函数 SaveConfig 在 bootstrapservice.go 中实现。
//
// 档位：
//
//	[S]     反汇编实证（函数体逐条对位）
//	[S-sig] 签名实证，体为骨架（待对应域专项还原）
//	[P]     骨架/占位
package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"
)

// ---- SaveConfig 附属函数 ----

// windowSizingChanged 判断新旧配置的窗口尺寸是否需要重算。
// [S] 取自 SaveConfig.asm 0x140777223 的 cmp [rsp+0x1dd8],[rsp+0x14a0]，
// 比较 LauncherConfig 偏移 0x80 处的 8 字节（即 Preferences.UIScalePercent，int）。
func windowSizingChanged(newCfg, oldCfg LauncherConfig) bool {
	return newCfg.Preferences.UIScalePercent != oldCfg.Preferences.UIScalePercent
}

// ---- syncConfigRuntimeAfterSave ----

// syncConfigRuntimeAfterSave 保存后同步运行时状态：书签 → 热键 → 运行时。
// [S 汇编 0x14079e160, 120L]（symbols.txt:19167）。
//
// asm 主干：
//
//	0x14079e1a5  convTslice(oldCfg.BookmarkSources) + convTslice(newCfg.BookmarkSources)
//	0x14079e1e4  reflect.DeepEqual(old, new)
//	0x14079e1eb  ≠ → cacheBookmarkSources(bs, newCfg.BookmarkSources)
//	0x14079e233  launcherHotkeyBindingsFromConfig(newCfg)          → newBindings
//	0x14079e294  launcherHotkeyBindingsFromConfig(oldCfg)          → oldBindings
//	0x14079e2e3  .eq.main.launcherHotkeyBindings(new, old)
//	0x14079e2e8  ≠ or currentHotkeyRegistrationErrors non-empty  → syncHotkeyBindingsFromConfig
//	0x14079e389  syncRuntimeServicesAfterSave(bs, newCfg)
func (bs *BootstrapService) syncConfigRuntimeAfterSave(oldCfg, newCfg LauncherConfig) {
	if bs == nil {
		return
	}

	// 书签源变化时重新缓存
	if !reflect.DeepEqual(oldCfg.Bookmarks.Sources, newCfg.Bookmarks.Sources) {
		bs.cacheBookmarkSources(newCfg.Bookmarks.Sources)
	}

	// 热键绑定变化时同步
	newBindings := launcherHotkeyBindingsFromConfig(newCfg)
	oldBindings := launcherHotkeyBindingsFromConfig(oldCfg)
	if !reflect.DeepEqual(newBindings, oldBindings) || bs.currentHotkeyRegistrationErrors() != nil {
		bs.syncHotkeyBindingsFromConfig(newCfg)
	}

	// 同步运行时服务
	bs.syncRuntimeServicesAfterSave(newCfg)
}

// ---- launcherHotkeyBindingsFromConfig ----

// launcherFeatureModuleKeys 是启动器特性模块固定键集（18 项，经 .data 解码实证）。
// launcherHotkeyBindingsFromConfig / syncHotkeyBindingsFromConfig / syncRuntimeServices
// 均以此构造默认启用表：全部默认 true，再被 Preferences.FeatureModules 覆盖。
var launcherFeatureModuleKeys = []string{
	"console", "desktopWidgets", "apps", "speedDial",
	"bookmarks", "files", "fileLocator", "audio",
	"gpu", "memoryRelease", "oledBlackout", "windowManagement",
	"mouseGestures", "twoFactor", "qrcode", "screenshot",
	"plugins", "tags",
}

// launcherFeatureModulesFromConfig 构建 18 项全默认启用 + FeatureModules 合并后的特性开关表。
// [S 汇编 0x14079cd20 / 0x14079d2e0]：两轮同一全局字符串切片循环——首轮 map[key]=true，
// 次轮 mapaccess2 命中 FeatureModules 则 mapassign 覆盖。
func launcherFeatureModulesFromConfig(cfg LauncherConfig) map[string]bool {
	enabled := make(map[string]bool, len(launcherFeatureModuleKeys))
	for _, k := range launcherFeatureModuleKeys {
		enabled[k] = true
	}
	if cfg.Preferences.FeatureModules != nil {
		for _, k := range launcherFeatureModuleKeys {
			if v, ok := cfg.Preferences.FeatureModules[k]; ok {
				enabled[k] = v
			}
		}
	}
	return enabled
}

// boolOrTrimmedNonEmpty 还原 6 个非首热键的 *bool 空值语义：nil → TrimSpace 非空。
// [S 汇编 0x14079cd20 / 0x14079d2e0]：*bool 非 nil 直接取值；nil 则用 TrimSpace 非空判定。
func boolOrTrimmedNonEmpty(b *bool, s string) bool {
	if b != nil {
		return *b
	}
	return strings.TrimSpace(s) != ""
}

// launcherHotkeyBindingsRawFromConfig 从 LauncherConfig 提取原始（未 normalize）热键绑定。
// [S 汇编 0x14079cd20 / 0x14079d2e0] 实证：
//
//	读 7 个 hotkey 字符串：SummonSearch 原样透传不 Trim；其余 6 个 TrimSpace 仅做空判定，
//	字符串本体同样原样透传。
//	读 7 个 *bool：首项 HotkeyEnabled nil→true；其余 6 项 nil→(TrimSpace 非空)。
//	以 launcherFeatureModulesFromConfig 合并后读 enabled["screenshot"] 作为 ScreenshotFeatureEnabled。
func launcherHotkeyBindingsRawFromConfig(cfg LauncherConfig) launcherHotkeyBindings {
	enabled := launcherFeatureModulesFromConfig(cfg)

	summonSearchEnabled := true
	if cfg.Preferences.HotkeyEnabled != nil {
		summonSearchEnabled = *cfg.Preferences.HotkeyEnabled
	}

	return launcherHotkeyBindings{
		SummonSearch:                  cfg.Preferences.Hotkey,
		SummonSearchEnabled:           summonSearchEnabled,
		SummonOnly:                    cfg.Preferences.SummonOnlyHotkey,
		SummonOnlyEnabled:             boolOrTrimmedNonEmpty(cfg.Preferences.SummonOnlyHotkeyEnabled, cfg.Preferences.SummonOnlyHotkey),
		Screenshot:                    cfg.Preferences.ScreenshotHotkey,
		ScreenshotEnabled:             boolOrTrimmedNonEmpty(cfg.Preferences.ScreenshotHotkeyEnabled, cfg.Preferences.ScreenshotHotkey),
		ScreenshotQRCode:              cfg.Preferences.ScreenshotQRCodeHotkey,
		ScreenshotQRCodeEnabled:       boolOrTrimmedNonEmpty(cfg.Preferences.ScreenshotQRCodeHotkeyEnabled, cfg.Preferences.ScreenshotQRCodeHotkey),
		ScreenshotAllScreens:          cfg.Preferences.ScreenshotAllScreensHotkey,
		ScreenshotAllScreensEnabled:   boolOrTrimmedNonEmpty(cfg.Preferences.ScreenshotAllScreensHotkeyEnabled, cfg.Preferences.ScreenshotAllScreensHotkey),
		ScreenshotScrolling:           cfg.Preferences.ScreenshotScrollingHotkey,
		ScreenshotScrollingEnabled:    boolOrTrimmedNonEmpty(cfg.Preferences.ScreenshotScrollingHotkeyEnabled, cfg.Preferences.ScreenshotScrollingHotkey),
		ScreenshotActiveWindow:        cfg.Preferences.ScreenshotActiveWindowHotkey,
		ScreenshotActiveWindowEnabled: boolOrTrimmedNonEmpty(cfg.Preferences.ScreenshotActiveWindowHotkeyEnabled, cfg.Preferences.ScreenshotActiveWindowHotkey),
		ScreenshotFeatureEnabled:      enabled["screenshot"],
	}
}

// launcherHotkeyBindingsFromConfig 从 LauncherConfig 提取并 normalize 热键绑定。
// [S 汇编 0x14079cd20, 280 行]：launcherHotkeyBindingsRawFromConfig → normalizeLauncherHotkeyBindings → 返回。
func launcherHotkeyBindingsFromConfig(cfg LauncherConfig) launcherHotkeyBindings {
	return normalizeLauncherHotkeyBindings(launcherHotkeyBindingsRawFromConfig(cfg))
}

// ---- syncHotkeyBindingsFromConfig ----

// syncHotkeyBindingsFromConfig 从配置同步热键绑定到全局热键管理器。
// [S 汇编 0x14079d2e0, 304B]：launcherHotkeyBindingsRawFromConfig → bs.syncHotkeyBindings(bindings)。
// normalize 在 syncHotkeyBindings 内部完成，故此处不预 normalize。
func (bs *BootstrapService) syncHotkeyBindingsFromConfig(cfg LauncherConfig) {
	if bs == nil {
		return
	}
	bs.syncHotkeyBindings(launcherHotkeyBindingsRawFromConfig(cfg))
}

// ---- syncRuntimeServicesAfterSave 辅助比较函数 ----

// screenshotCapturePreferencesEqual 比较两份 Preferences 的截图捕获三元组是否等价。
// [S 汇编 0x14079d7e0, 160B(0xa0)]：仅比较 ScreenshotCaptureControls / ScreenshotCaptureCursor /
// ScreenshotSelectionConfirm 三个 *bool（*bool nil 缺省：controls 默认 true，cursor/selectionConfirm 默认 false，
// 与 cacheScreenshotCapturePreferences 的缓存语义一致）。
func screenshotCapturePreferencesEqual(a, b Preferences) bool {
	ac := true
	if a.ScreenshotCaptureControls != nil {
		ac = *a.ScreenshotCaptureControls
	}
	bc := true
	if b.ScreenshotCaptureControls != nil {
		bc = *b.ScreenshotCaptureControls
	}
	if ac != bc {
		return false
	}

	ac = false
	if a.ScreenshotCaptureCursor != nil {
		ac = *a.ScreenshotCaptureCursor
	}
	bc = false
	if b.ScreenshotCaptureCursor != nil {
		bc = *b.ScreenshotCaptureCursor
	}
	if ac != bc {
		return false
	}

	ac = false
	if a.ScreenshotSelectionConfirm != nil {
		ac = *a.ScreenshotSelectionConfirm
	}
	bc = false
	if b.ScreenshotSelectionConfirm != nil {
		bc = *b.ScreenshotSelectionConfirm
	}
	return ac == bc
}

// fileSearchRuntimeConfigsEqual 判断两份文件搜索配置经规范化后是否等价。
// [S 汇编 0x14079d880, 384B(0x180)]：normalizeFileSearchConfig(a/b) → reflect.DeepEqual。
func fileSearchRuntimeConfigsEqual(a, b FileSearchConfig) bool {
	return reflect.DeepEqual(normalizeFileSearchConfig(a), normalizeFileSearchConfig(b))
}

// ---- syncRuntimeServicesAfterSave ----

// syncRuntimeServicesAfterSave 保存后同步运行时服务（注入新配置）。
// [S-sig 汇编 0x14079e3c0, 3182B(0xc6e)]（symbols.txt:19168）：入参 cfg 栈传入 0x938 字节 + bs。
// 旧注释误标 240B——实测体长 0xc6e=3182B，系「保存后全服务差异同步」编排器，非短链。
//
// asm 实证主干（按序）：
//  1. 构建两张 map[string]bool（旧/新 feature 位图），按两 slice（0x10 步长键值对）mapassign=1，
//     再 mapaccess2 旧位图 diff 得「变更集」。
//  2. lock bs.mutex(+0x540) 后：
//     a. screenshotCapturePreferencesEqual(cfg.Preferences, 缓存) → 不等则 cacheScreenshotCapturePreferences
//     b. memoryReleaseConfigsEqual → 不等则 memoryReleaseService.Configure
//     c. oledBlackout reflect.DeepEqual → 不等则 oledBlackoutService.Configure，等则 retryOLEDBlackoutHotkeysIfNeeded
//     d. windowManagement reflect.DeepEqual → 不等则 windowManagementService.Configure
//     e. mouseGesture reflect.DeepEqual → 不等则 mouseGestureService.Configure
//     f. fileSearchRuntimeConfigsEqual → 不等则 syncFileSearchRuntime
//     g. syncCursorWrapHotCornerGuard(features, hotCorner)
//     h. desktopWidget feature diff → 不等则 desktopWidgetService.Configure
//
// [P] 体依赖 8 服务域 + 12 helper（screenshotCapturePreferencesEqual/memoryReleaseConfigsEqual/
// fileSearchRuntimeConfigsEqual 等），逐域落地后方可 [S]。当前保留已落地的截图缓存同步作为骨架近似。
func (bs *BootstrapService) syncRuntimeServicesAfterSave(cfg LauncherConfig) {
	if bs == nil {
		return
	}
	bs.cacheScreenshotCapturePreferences(cfg.Preferences)
}

// ---- contentRuntimeSyncNeededAfterSave ----

// contentRuntimeSyncNeededAfterSave 判断保存后是否需要内容运行时同步。
// [S-sig 汇编 0x14079da00, 208B]（symbols.txt:19165）：入参 cfg 栈传入 0x938 字节，
// 返回 bool（al）。[P] 判据（内容目录/收藏夹/速拨等变化）待运行时同步域专项闭合。
// 当前保守返回 true 以触发同步。
func contentRuntimeSyncNeededAfterSave(cfg LauncherConfig) bool {
	_ = cfg
	return true
}

// ---- mergeConfigStorageWithWorkspace ----

// mergeConfigStorageWithWorkspace 将输入配置的存储字段与工作区目录对齐。
// [S-sig 汇编 0x1407a2160, 240B]（symbols.txt:19192）：
// 入参 cfg（栈 0x938）+ ws（栈 0xa0），返回合并后 LauncherConfig（栈 0x938）。
//
// asm 主干：把 cfg 整个拷贝到结果区，随后将 ws 的系列目录字段写入 cfg.Storage。
// 涉及 5 个字段对：DataRoot/Root, IconDir/IconDir, IndexDir/IndexDir,
// ScreenshotDir/ScreenshotDir, WebView2Dir/WebView2Dir。
// [P] 映射确认：ws.RawWebView2Dir/ws.WebView2Dir 等确切字段名待工作区域专项闭合。
// 此处按 WorkspaceLayout 结构的 StorageConfig 对应字段名写入。
func mergeConfigStorageWithWorkspace(cfg LauncherConfig, ws WorkspaceLayout) LauncherConfig {
	merged := cfg
	merged.Storage = StorageConfig{
		DataRoot:      ws.Root,
		IconDir:       ws.IconDir,
		IndexDir:      ws.IndexDir,
		ScreenshotDir: ws.ScreenshotDir,
		WebView2Dir:   ws.WebView2Dir,
	}
	return merged
}

// ---- pruneLauncherConfigBackups ----

// pruneLauncherConfigBackups 当配置版本变化时清理旧备份。
// [S 汇编 0x140877de0, 1920B]：入参 (path string, retention int)。
// asm 实证（392 行）：ReadDir → FilenameParts 过滤 → time.Parse("2006-01-02") 筛选备份文件 →
// 收集 DirEntry 切片 → sort.Slice 排序 → 保留至少 retention 个最新备份 → os.Remove 多余项。
func pruneLauncherConfigBackups(path string, retention int) {
	backupDir := launcherConfigBackupDir(path)
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return
		}
		return
	}
	if retention <= 0 {
		retention = 100
	}
	if retention < 10 {
		retention = 10
	}
	if retention > 1000 {
		retention = 1000
	}

	prefix, _ := launcherConfigBackupFilenameParts(path)

	var backups []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// 前缀匹配
		if len(name) < len(prefix) {
			continue
		}
		if name[:len(prefix)] != prefix {
			continue
		}
		// 剩余部分必须含日期格式 "2006-01-02"
		datePart := name[len(prefix):]
		if len(datePart) < 10 {
			continue
		}
		datePart = datePart[:10]
		if _, err := time.Parse("2006-01-02", datePart); err != nil {
			continue
		}
		backups = append(backups, filepath.Join(backupDir, name))
	}

	if len(backups) <= retention {
		return
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i] > backups[j]
	})

	remove := backups[:len(backups)-retention]
	for _, f := range remove {
		if err := os.Remove(f); err != nil && !errors.Is(err, os.ErrNotExist) {
			return
		}
	}
}

// ---- CompareAndSwapPrepared ----

// CompareAndSwapPrepared 原子替换已准备配置到存储（savePreparedUnlocked 的外部壳）。
// [S-sig 汇编 0x14089b060, 296B]（symbols.txt:20528）：
// 入口先写回 configStore 的 path 字段（同步给 cache），再调用 savePreparedUnlocked(true, cfg)。
// [P] 写回 path 映射待确认（l145 的 cfg.* 哪些字段回填 store.path）。
func (s *launcherConfigStore) CompareAndSwapPrepared(cfg LauncherConfig) error {
	if s == nil {
		return nil
	}
	return s.savePreparedUnlocked(true, cfg)
}
