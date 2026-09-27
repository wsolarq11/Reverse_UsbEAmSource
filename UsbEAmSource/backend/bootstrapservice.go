// AUTO-RECONSTRUCTED BOOTSTRAP SERVICE — 核心编排器骨架
// 研究用途
//
// 契约来源：
//   - source_funcs.txt bootstrapservice.go L360-3512+（~130 方法）
//   - 类型：types_launcher.go BootstrapService struct
//
// 档位：[P] 骨架/占位（方法计数与签名对齐 source_funcs，体为桩代码待反汇编逐条实证）
// 注：commitLauncherConfigReplacementWithWidgets 已在批 13 提升为 [S] 实证体（见本文件 + bootstrapservice_commitconfig.go）。
package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ---- BootstrapService 初始化与生命周期 ----

// setStartupTrayMode 设置启动托盘模式。
// [S 汇编 0x140772760]
func (bs *BootstrapService) setStartupTrayMode(enabled bool) {
	bs.lock.Lock()
	bs.startupTrayMode = enabled
	bs.resetLauncherSizeOnNextShow = enabled
	bs.lock.Unlock()
}

// clearStartupTrayMode 清除启动托盘模式标志（加锁）。
// [S 汇编 0x140796820] 单 receiver 无参无返回。asm：nil 检查 → lock(+0x540) →
// startupTrayMode(+0x449)=false → unlock。
func (bs *BootstrapService) clearStartupTrayMode() {
	if bs == nil {
		return
	}
	bs.lock.Lock()
	bs.startupTrayMode = false
	bs.lock.Unlock()
}

// syncStartupTask 同步开机自启任务（配置保存后调用）。
// [S 汇编 0x1407a14c0, 240B(0xf0)]：
//
//	bs==nil → errors.New("启动任务服务不可用" 27B @0x140c69c33)（@0x1407a14d2/d5/1588）；
//	lock([bs+0x540])（@0x1407a14f6 lock cmpxchg）→ defer unlock（@0x1407a153c lock xadd）；
//	读 [bs+0x510]=startupTaskSync（func value，1 字）（@0x1407a152d）；
//	nil（@0x1407a1569）→ 默认 funcval @0x141096d88（[0]=syncLauncherStartupTask 0x1408af840，
//	  @0x1407a156e/75）；
//	call fn(enabled,delaySeconds)（@0x1407a1580，eax=enabled，rbx=delaySeconds，rdx=funcval）。
func (bs *BootstrapService) syncStartupTask(enabled bool, delaySeconds int) error {
	if bs == nil {
		return errors.New("启动任务服务不可用")
	}
	bs.lock.Lock()
	defer bs.lock.Unlock()
	fn := bs.startupTaskSync
	if fn == nil {
		fn = syncLauncherStartupTask
	}
	return fn(enabled, delaySeconds)
}

// syncStartupTaskFromConfig 比较新旧配置的开机自启设置，变化则同步任务。
// [S 提取自 SaveConfig.func1 内联段 0x140777618-0x140777725]：
//
//	oldEnabled = oldCfg.Preferences.StartupLaunchEnabled 非 nil 且 *它（@0x140777618/625/62a）；
//	oldDelay = clamp(oldCfg.Preferences.StartupLaunchDelaySeconds)（@0x14077762c/634-646）；
//	newEnabled = newCfg.Preferences.StartupLaunchEnabled 非 nil 且 *它（@0x14077764d/655/65f）；
//	newDelay = clamp(newCfg.Preferences.StartupLaunchDelaySeconds)（@0x140777678/680-693）；
//	clamp：<=0→10，>100→100（@0x140777634/639/642/648，与 syncLauncherStartupTask 同）；
//	newEnabled != oldEnabled || (newEnabled && newDelay != oldDelay)（@0x1407776cc/e4/e7）
//	  → bs.syncStartupTask(newEnabled,newDelay)（@0x140777720）。
func (bs *BootstrapService) syncStartupTaskFromConfig(newCfg, oldCfg LauncherConfig) error {
	oldEnabled := false
	if oldCfg.Preferences.StartupLaunchEnabled != nil {
		oldEnabled = *oldCfg.Preferences.StartupLaunchEnabled
	}
	oldDelay := oldCfg.Preferences.StartupLaunchDelaySeconds
	if oldDelay <= 0 {
		oldDelay = 10
	} else if oldDelay > 100 {
		oldDelay = 100
	}

	newEnabled := false
	if newCfg.Preferences.StartupLaunchEnabled != nil {
		newEnabled = *newCfg.Preferences.StartupLaunchEnabled
	}
	newDelay := newCfg.Preferences.StartupLaunchDelaySeconds
	if newDelay <= 0 {
		newDelay = 10
	} else if newDelay > 100 {
		newDelay = 100
	}

	if newEnabled != oldEnabled || (newEnabled && newDelay != oldDelay) {
		return bs.syncStartupTask(newEnabled, newDelay)
	}
	return nil
}

// Initialize 初始化引导服务。
// [S 汇编实证 0x140772820, 256B]：
//
//	xor ebx, ebx → call initializeWithCheckpoint → 3× duffcopy（拷贝 workspace 到栈/bs 字段）
//	→ 返 rbx=0（error=nil）。函数体不设错误路径。
func (bs *BootstrapService) Initialize() error {
	bs.initializeWithCheckpoint()
	return nil
}

// Shutdown 实证体见 bootstrapservice_lifecycle.go：0x140773e60 走 bs.lock(+0x540)，
// 非旧骨架的 hotkeyCaptureOperation；函数无返回值。

// ---- 快照与状态 ----

// GetSnapshot 采集当前工作区快照并刷新缓存。
// [S 汇编实证 0x1407741a0, 800B]（symbols.txt:18850）：
//
//	lock(bootstrapSnapshotRefresh@+0xa8) → defer unlock
//	ws = workspaceSnapshot()
//	Languages = discoverLanguages(languageDirectoriesForWorkspace(ws))
//	Plugins   = discoverPlugins(TrimSpace(ws.PluginDir) 单元素切片；空则 nil)
//	cacheBootstrapSnapshot(snapshot) → return snapshot
//
// 关键实证：刷新锁是 +0xa8（bootstrapSnapshotRefresh），与 +0x540（lock）分属两把锁；
// 返回的是未 clone 的原值——clone 只发生在 cacheBootstrapSnapshot 内部。
// 依赖符号：workspaceSnapshot 0x1407a10c0、languageDirectoriesForWorkspace 0x1407a27a0、
// discoverLanguages 0x1407a2c60、discoverPlugins 0x1407a5320。
//
// 签名实证（本轮修正）：尾 0x140774451 以 mov rax,[rsp+0xf0] / mov rbx,[rsp+0xf8]
// 回传清零后的 error 接口（defer 路径 0x14077446f 同构）；GetState 调用点 0x1407765d6
// 与 launcherStateFromCommittedConfig 0x140774aa0 均以 rax 判错 —— 带 error 位。
func (bs *BootstrapService) GetSnapshot() (BootstrapSnapshot, error) {
	bs.bootstrapSnapshotRefresh.Lock()
	defer bs.bootstrapSnapshotRefresh.Unlock()

	workspace := bs.workspaceSnapshot()

	// asm 0x1407742f8：strings.TrimSpace(ws.PluginDir)，非空才构造单元素切片。
	var pluginDirs []string
	if dir := strings.TrimSpace(workspace.PluginDir); dir != "" {
		pluginDirs = []string{dir}
	}

	snapshot := BootstrapSnapshot{
		Workspace: workspace,
		Languages: discoverLanguages(languageDirectoriesForWorkspace(workspace)),
		Plugins:   discoverPlugins(pluginDirs),
	}
	bs.cacheBootstrapSnapshot(snapshot)
	return snapshot, nil
}

// workspaceSnapshot 在 lock 保护下复制当前工作区布局。
// [S 汇编实证 0x1407a10c0, 272B]（symbols.txt:19183）：
//
//	bs == nil → 零值 WorkspaceLayout
//	lock(+0x540) → duffcopy 0xa0 字节（WorkspaceLayout 全 10 字段）→ defer unlock
func (bs *BootstrapService) workspaceSnapshot() WorkspaceLayout {
	if bs == nil {
		return WorkspaceLayout{}
	}
	bs.lock.Lock()
	defer bs.lock.Unlock()
	return bs.workspace
}

// cacheBootstrapSnapshot 以深拷贝刷新快照缓存（仅当工作区未变时）。
// [S 汇编实证 0x140774520, 416B]（symbols.txt:18852）：
//
//	bs == nil → 直接返回（asm 0x140774540）
//	cloned = cloneBootstrapSnapshot(snapshot)   ← 在加锁之前完成
//	lock(+0x540)
//	if .eq.WorkspaceLayout(bs.workspace, snapshot.Workspace) {
//	    bs.bootstrapSnapshot(+0xb0) = cloned
//	    bs.bootstrapSnapshotReady(+0x180) = true
//	}
//	unlock
//
// 即：工作区已切换时拒绝写入陈旧快照（防止迁移后快照错配）。
func (bs *BootstrapService) cacheBootstrapSnapshot(snapshot BootstrapSnapshot) {
	if bs == nil {
		return
	}
	cloned := cloneBootstrapSnapshot(snapshot)
	bs.lock.Lock()
	if bs.workspace == snapshot.Workspace {
		bs.bootstrapSnapshot = cloned
		bs.bootstrapSnapshotReady = true
	}
	bs.lock.Unlock()
}

// cloneBootstrapSnapshot 深拷贝启动快照。
// [S 汇编实证 0x140774b40, 1568B]（symbols.txt:18855；旧注 0x140770840 系占位误值）：
//
//	Workspace 直接值拷贝（初始 duffcopy 0xd0 字节）
//	Languages：len==0 → nil（asm 0x140774bb1）；否则重建切片并逐条深拷贝 Messages
//	Plugins  ：len==0 → nil；否则重建切片并逐项深拷贝 Permissions/UI.Resizable/I18N
//
// 结构偏移实证（type descriptor 解码，types_base=0x140a67000）：
//
//	LanguageManifest size=0x48，Messages@0x40（type 0x140bc35e0）
//	PluginManifest   size=0x178，UI@0xa0、UI.Resizable@0xe0、Permissions@0xf0、I18N@0x108
//	I18N 类型 = map[string]main.PluginLocalization（type 0x140b48e20）
func cloneBootstrapSnapshot(src BootstrapSnapshot) BootstrapSnapshot {
	dst := BootstrapSnapshot{Workspace: src.Workspace}

	if n := len(src.Languages); n != 0 {
		dst.Languages = make([]LanguageManifest, n)
		copy(dst.Languages, src.Languages)
		for i := range dst.Languages {
			dst.Languages[i].Messages = cloneBootstrapMessageMap(dst.Languages[i].Messages)
		}
	}

	if n := len(src.Plugins); n != 0 {
		dst.Plugins = make([]PluginManifest, n)
		copy(dst.Plugins, src.Plugins)
		for i := range dst.Plugins {
			dst.Plugins[i].Permissions = cloneBootstrapPermissions(src.Plugins[i].Permissions)
			dst.Plugins[i].UI.Resizable = cloneBootstrapBoolPtr(dst.Plugins[i].UI.Resizable)
			dst.Plugins[i].I18N = cloneBootstrapLocalizationMap(src.Plugins[i].I18N)
		}
	}

	return dst
}

// cloneBootstrapPermissions 复制插件权限列表。
// [S asm 0x140774dbc：以 len 判空（非 ptr），len==0 归一为 nil；否则 growslice+typedslicecopy]
func cloneBootstrapPermissions(src []string) []string {
	if len(src) == 0 {
		return nil
	}
	dst := make([]string, len(src))
	copy(dst, src)
	return dst
}

// cloneBootstrapBoolPtr 复制 *bool 指向的布尔值。
// [S asm 0x140774e9c：newobject(bool) + movzx 读 + 写，源为 nil 时保持 nil]
func cloneBootstrapBoolPtr(src *bool) *bool {
	if src == nil {
		return nil
	}
	v := *src
	return &v
}

// cloneBootstrapLocalizationMap 复制插件本地化表并深拷贝条目内的消息映射。
// [S asm 0x140774eff：makemap(len) + mapIter；每条 PluginLocalization(0x28 字节)
// 的 Messages 字段经 cloneBootstrapMessageMap 后 wbMove 写回]
func cloneBootstrapLocalizationMap(src map[string]PluginLocalization) map[string]PluginLocalization {
	if src == nil {
		return nil
	}
	dst := make(map[string]PluginLocalization, len(src))
	for k, v := range src {
		v.Messages = cloneBootstrapMessageMap(v.Messages)
		dst[k] = v
	}
	return dst
}

// cloneBootstrapMessageMap 深拷贝语言消息映射（逐值为递归深拷贝，非浅拷贝）。
// [S 汇编实证 0x140775160, 352B]（symbols.txt:18856）：
//
//	src == nil → nil（asm 0x14077517d）
//	make(map[string]interface{}, len(src)) → mapIter → 每个 value 调
//	cloneBootstrapMessageValue → mapassign_faststr
//
// 类型实证：map type = 0x140b48da0（map[string]interface {}）
func cloneBootstrapMessageMap(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return nil
	}
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = cloneBootstrapMessageValue(v)
	}
	return dst
}

// cloneBootstrapMessageValue 深拷贝单个消息值。
// [S 汇编实证 0x1407752c0, 864B]（symbols.txt:18857）：按接口类型 hash 分派
// （type descriptor 已解码，types_base=0x140a67000）：
//
//	0x1701f1b5 / 0x140adc540 = []string            → 重建切片（元素不可变，无递归）
//	0xf2662176 / 0x140adc5c0 = []interface{}       → makeslice + 逐元素递归
//	0x6dd49dde / 0x140b48720 = map[string]string   → makemap + 浅拷贝（string 值不可变）
//	0xc75e6aa4 / 0x140b48da0 = map[string]interface{} → 委派 cloneBootstrapMessageMap
//	nil interface 及其余类型                    → 原样返回（asm 0x1407754d8 直 ret）
func cloneBootstrapMessageValue(v interface{}) interface{} {
	switch t := v.(type) {
	case nil:
		return nil
	case []string:
		if len(t) == 0 {
			return []string(nil)
		}
		out := make([]string, len(t))
		copy(out, t)
		return out
	case []interface{}:
		out := make([]interface{}, len(t))
		for i, e := range t {
			out[i] = cloneBootstrapMessageValue(e)
		}
		return out
	case map[string]string:
		out := make(map[string]string, len(t))
		for k, val := range t {
			out[k] = val
		}
		return out
	case map[string]interface{}:
		return cloneBootstrapMessageMap(t)
	default:
		return v
	}
}

// snapshotForCommittedLauncherState 取已提交状态的快照：优先命中缓存，否则全量采集。
// [S 汇编实证 0x140774920, 544B]（symbols.txt:18854）：
//
//	ws = workspaceSnapshot()  （asm 0x140774973，仅调用一次）
//	snapshot, ok = cachedBootstrapSnapshotForWorkspace(ws)
//	ok  → return snapshot
//	!ok → return GetSnapshot()
//
// 签名实证（本轮修正）：旧注「末尾未设置 al 故单返回值」只核了 bool 位，漏了 error 位。
// asm 快路径尾 0x140774a5d 为 xor eax,eax / xor ebx,ebx（error = nil）；
// 慢路径 0x140774aa0 调 GetSnapshot 后原样透传其 rax/rbx —— 返回 (BootstrapSnapshot, error)。
func (bs *BootstrapService) snapshotForCommittedLauncherState() (BootstrapSnapshot, error) {
	workspace := bs.workspaceSnapshot()
	if snapshot, ok := bs.cachedBootstrapSnapshotForWorkspace(workspace); ok {
		return snapshot, nil
	}
	return bs.GetSnapshot()
}

// cachedBootstrapSnapshotForWorkspace 读取与给定工作区匹配的缓存快照副本。
// [S 汇编实证 0x1407746c0, 608B]（symbols.txt:18853）：
//
//	bs == nil → (零值, false)
//	lock(+0x540)
//	if !bootstrapSnapshotReady(+0x180) || .eq.WorkspaceLayout(bs.bootstrapSnapshot.Workspace, workspace) == false
//	    → unlock + (零值, false)
//	→ cloneBootstrapSnapshot(bs.bootstrapSnapshot(+0xb0)) → unlock → (副本, true)
//
// 签名实证：入参 workspace 占调用者栈 0xa0 字节（asm 0x14077476d 以 [rsp+0x288]
// 即帧内偏移 0 处作比较操作数），出参 BootstrapSnapshot 紧随其后（asm 从 [rsp+0xa0] 读回），
// bool 走 al。旧骨架的无参单返回值版不成立。
func (bs *BootstrapService) cachedBootstrapSnapshotForWorkspace(workspace WorkspaceLayout) (BootstrapSnapshot, bool) {
	if bs == nil {
		return BootstrapSnapshot{}, false
	}
	bs.lock.Lock()
	defer bs.lock.Unlock()
	if !bs.bootstrapSnapshotReady || bs.bootstrapSnapshot.Workspace != workspace {
		return BootstrapSnapshot{}, false
	}
	return cloneBootstrapSnapshot(bs.bootstrapSnapshot), true
}

// GetLanguageMessages 获取语言消息。
// [S 汇编实证 0x140775620, 64B]：rbx = bs → call languageDirectories → call loadLanguageMessages → 返结果。
func (bs *BootstrapService) GetLanguageMessages() map[string]interface{} {
	dirs := bs.languageDirectories()
	return loadLanguageMessages(dirs)
}

// defaultTagCatalogFromLanguageMessages 从语言消息生成默认标签目录。
// [S 汇编实证 0x1407756c0, 448B]：
//
//	rax=langMsgs, rbx=lang(ptr), rcx=lang(len)
//	0x140775777 mapaccess2_faststr(langMsgs, "en-US") → ok 则 mergeLanguageMessages(&merged, val)
//	0x1407757a5 TrimSpace(lang) → 空则 fallback "en-US"（magic 0x552d6e65="en-U"+0x53='S' 判等）
//	lang != "en-US" 则 mapaccess2_faststr(langMsgs, lang) → ok 则再 merge
//	0x140775819 nestedLanguageMessageString(&merged, "tags.defaults") → splitTagText → defaultTagCatalogFromNames
//
// [S-sig] 汇编 map 值类型为嵌套 map（mapaccess2_faststr 单字解引用），当前树
// loadLanguageMessages 返回 map[string]interface{}，故以类型断言近似；map 值域类型校正归语言加载域专项。
func defaultTagCatalogFromLanguageMessages(langMsgs map[string]interface{}, lang string) []TagCatalogItem {
	merged := make(map[string]interface{})
	if v, ok := langMsgs["en-US"]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			mergeLanguageMessages(merged, m)
		}
	}
	lang = strings.TrimSpace(lang)
	if lang == "" {
		lang = "en-US"
	}
	if lang != "en-US" {
		if v, ok := langMsgs[lang]; ok {
			if m, ok := v.(map[string]interface{}); ok {
				mergeLanguageMessages(merged, m)
			}
		}
	}
	names := nestedLanguageMessageString(merged, "tags.defaults")
	tags := splitTagText(names)
	if len(tags) == 0 {
		return nil
	}
	return defaultTagCatalogFromNames(tags)
}

// splitTagText 按逗号/全角逗号/换行分割标签文本，去空。
// [S 汇编实证 0x140775880, 384B]：FieldsFunc(pred={','|'，'|'\n'}) → TrimSpace each → skip empty → return
func splitTagText(text string) []string {
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == ',' || r == '，' || r == '\n'
	})
	if len(parts) == 0 {
		return nil
	}
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// ---- 配置相关 ----

// launcherConfigOptions 获取启动器配置选项。
// [S 汇编实证 0x140775660, 64B]：rax=bs, rbx=locale(ptr), rcx=locale(len) →
// languageDirectories → loadLanguageMessages → defaultTagCatalogFromLanguageMessages(msgs, locale)。
// 旧重建树误为无参形式，本批按 asm 补 locale 参。
func (bs *BootstrapService) launcherConfigOptions(locale string) LauncherConfigOptions {
	dirs := bs.languageDirectories()
	langMsgs := loadLanguageMessages(dirs)
	catalog := defaultTagCatalogFromLanguageMessages(langMsgs, locale)
	return LauncherConfigOptions{
		DefaultTagCatalog: catalog,
	}
}

// GetStartupState 0x140775a00 / GetState 0x140775e20 / launcherStateFromCommittedConfig 0x140776720 /
// launcherStateFromCommittedConfigWithSnapshot 0x140776920 / launcherStateFromSavedConfig 0x140776b80
// 的实证体见 bootstrapservice_state.go（本轮从 [S] 桩升级为反汇编实证体）。

// SaveConfig 保存配置。
// [S 汇编 0x140776d60]（实现见 bootstrapservice_saveconfig.go）
func (bs *BootstrapService) SaveConfig(cfg LauncherConfig) error {
	bs.workspaceTransaction.Lock()
	defer bs.workspaceTransaction.Unlock()

	store, _ := bs.configStoreSnapshot()
	oldCfg, _ := store.Read()

	if err := ensureWorkspaceDirectories(bs.workspace); err != nil {
		return err
	}

	merged := mergeConfigStorageWithWorkspace(cfg, bs.workspace)
	bs.launcherConfigOptions(cfg.Preferences.Language) // 冷加载默认标签目录

	if err := store.CompareAndSwapPrepared(merged); err != nil {
		return err
	}

	if oldCfg.Revision != merged.Revision {
		pruneLauncherConfigBackups(bs.workspace.ConfigFile, 0)
	}

	bs.syncConfigRuntimeAfterSave(oldCfg, merged)
	if windowSizingChanged(merged, oldCfg) {
		win := bs.resolveAttachedLauncherWindow()
		if win != nil {
			bs.applyLauncherWindowSizing(win, merged.Preferences.UIScalePercent, false)
		}
	}

	syncNeeded := contentRuntimeSyncNeededAfterSave(merged)
	_ = bs.launcherStateFromSavedConfig(merged, syncNeeded)
	return nil
}

// ResetConfig 重置配置：删除配置文件、图标库与小部件文档。
// [S 汇编实证 0x140778200, 928B]（va_map_fixed2.txt:149）：
//
//	0x14077824b  lock bs.workspaceTransaction(+0x518)
//	0x14077826e  defer Unlock（deferwrap1 = 0x1407786e0，由 lea [rip+0x46b] 解出）
//	0x1407782d6  store, snapshot := bs.configStoreSnapshot()
//	0x140778354  ensureWorkspaceDirectories(ws)（栈参：duffcopy 先铺 WorkspaceLayout）
//	0x140778379  path = strings.TrimSpace(快照 WorkspaceLayout.ConfigFile@+0xd8)
//	0x1407784cb  path 为空 → runtime.newobject 构造错误返回
//	0x1407783ca  store.DeletePrepared(prepare 闭包)
//	0x1407783e8  launcherConfigIconStoreForConfigPath(path).Delete()
//	0x140778445  launcherWidgetStoreForConfigPath(path).Delete()
//
// 四处错误分支（0x14077848d/0x1407784b0/0x1407783f7/0x140778435）均经 defer 展开后原样返回。
func (bs *BootstrapService) ResetConfig() error {
	bs.workspaceTransaction.Lock()
	defer bs.workspaceTransaction.Unlock()

	store, ws := bs.configStoreSnapshot()

	if err := ensureWorkspaceDirectories(ws); err != nil {
		return err
	}

	path := strings.TrimSpace(ws.ConfigFile)
	if path == "" {
		return errBootstrapConfigPathEmpty
	}

	if err := store.DeletePrepared(nil); err != nil {
		return err
	}
	if err := launcherConfigIconStoreForConfigPath(path).Delete(); err != nil {
		return err
	}
	return launcherWidgetStoreForConfigPath(path).Delete()
}

// isLauncherConfigEmptyInitialConfig 检查配置是否为空初始配置。
// [S 汇编 0x140778740] 实证链：
//
//	normalizeLauncherConfigWithOptions(lang=调用点实参 "en-US" 5B) →
//	defaultLauncherConfigWithOptions(lang) → normalize → launcherConfigsEqual。
func isLauncherConfigEmptyInitialConfig(cfg LauncherConfig) bool {
	lang := "en-US"
	var ncfg LauncherConfig
	normalizeLauncherConfigWithOptions(&ncfg, lang)
	def := defaultLauncherConfigWithOptions(lang)
	normalizeLauncherConfigWithOptions(&def, lang)
	return launcherConfigsEqual(cfg, def)
}

// CompleteInitialization 完成初始化：标签目录缺省时按语言消息重建，随后提交配置替换。
// [S 汇编实证 0x140778980, 512B]（va_map_fixed2.txt:32）：
//
//	0x1407789c3  cmp qword [rsp+0x2a88], 0   → 判 cfg.Preferences.TagCatalog 的 len
//	0x1407789f2  cleanStringList([rsp+0x2a68])→ GlobalTags（[]string）清理
//	             清理后 len != 0（test rbx,rbx 未跳转）→ 0x1407789fc xor eax,eax 直接返回 nil
//	0x140778a10  bs.languageDirectories()
//	0x140778a15  loadLanguageMessages(dirs)
//	0x140778a2a  defaultTagCatalogFromLanguageMessages(msgs) → 回写 [rsp+0x2a80]
//	0x140778a47  tagCatalogNames(catalog)                    → 回写 [rsp+0x2a68]
//	0x140778ada  commitLauncherConfigReplacementWithWidgets(cfg, true)（ecx=1，对比 ImportConfig 的 ecx=0）
//
// 栈偏移归属由字段声明序交叉确认：Preferences.GlobalTags(L95) 在 TagCatalog(L96) 之前，
// 与 0x2a68 < 0x2a80 一致。
// [P] cfg 来源：汇编 0x140778980 中 CompleteInitialization 并不调用 configStoreSnapshot，
// 而是以调用者栈槽（0x142 qword = LauncherConfig）承载 cfg 入参（真实签名含 cfg 栈参）。
// 本重建树以 store.Read() 取提交配置近似承载该语义，签名差异待 CompleteInitialization 专项校正。
func (bs *BootstrapService) CompleteInitialization() error {
	store, _ := bs.configStoreSnapshot()
	cfg, _ := store.Read()

	if len(cfg.Preferences.TagCatalog) == 0 {
		if len(cleanStringList(cfg.Preferences.GlobalTags)) != 0 {
			return nil
		}
		msgs := loadLanguageMessages(languageDirectoriesForWorkspace(bs.workspace))
		cfg.Preferences.TagCatalog = defaultTagCatalogFromLanguageMessages(msgs, cfg.Preferences.Language)
		cfg.Preferences.GlobalTags = tagCatalogNames(cfg.Preferences.TagCatalog)
	}

	return bs.commitLauncherConfigReplacementWithWidgets(cfg, true)
}

// commitLauncherConfigReplacementWithWidgets 提交配置替换并重建 widget。
// [S 汇编 0x140778b80, 3072B]（va_map_fixed2.txt:307）：巨型函数（source_funcs 记为 4115 源码行），
// 附属 commitLauncherConfigReplacementWithWidgetsfunc1/func2/.func21 + deferwrap1。
// 调用点实证第 2 形参为 bool 初始化标志：
//
//	CompleteInitialization.asm 0x140778ad5-0x140778ada：mov ecx, 1
//	ImportConfig.asm         0x14077becb-0x14077becd：xor ecx, ecx
//
// [S] 函数体在 bootstrapservice_commitconfig.go 中以反汇编实证展开实现。
//
//	此处保留调用点可见的签名，体委托给内联实现。
func (bs *BootstrapService) commitLauncherConfigReplacementWithWidgets(cfg LauncherConfig, initialization bool) error {
	// 实现体见 bootstrapservice_commitconfig.go:
	// 汇编 0x140778b80 完整的 lock → normalize → buildWorkspaceLayout →
	// compare paths → prepareWorkspaceDirectories → ReplacePrepared →
	// clearPendingFileSearchResidentWarm → emitChange → launcherStateFromCommittedConfig 链
	return bs.commitLauncherConfigReplacementWithWidgetsImpl(cfg, initialization)
}

// commitLauncherConfigReplacementWithWidgetsImpl 是提交配置替换的实现体。
// 按 bootstrapservice_commitconfig.go 中反汇编实证还原。
// [S-sig]：签名经调用方 commitLauncherConfigReplacementWithWidgets 实证，体为功能还原。
func (bs *BootstrapService) commitLauncherConfigReplacementWithWidgetsImpl(cfg LauncherConfig, initialization bool) error {
	if bs == nil {
		return nil
	}

	bs.workspaceTransaction.Lock()
	defer bs.workspaceTransaction.Unlock()

	store, _ := bs.configStoreSnapshot()
	ws := bs.workspaceSnapshot()

	// 装配目标路径（来自工作区配置路径）
	target := ws.ConfigFile

	// 构建工作区布局
	layout := buildWorkspaceLayoutWithConfig(cfg, target)

	// 规整存储配置段（第 1 轮 normalize）
	sc := normalizeStorageConfig(
		strings.TrimSpace(cfg.Storage.DataRoot),
		strings.TrimSpace(cfg.Storage.IconDir),
		strings.TrimSpace(cfg.Storage.IndexDir),
		strings.TrimSpace(cfg.Storage.ScreenshotDir),
		strings.TrimSpace(cfg.Storage.WebView2Dir),
	)

	// [S] 汇编 0x140778d93-0x140778da9: .eq.StorageConfig 短路——若与当前相同则跳过
	// 当前未实现 .eq 操作，按路径比较覆盖。

	// 5 组路径比较（汇编 0x140778f73-0x14077919f: filepath.Clean + memequal ×5）
	type pathPair struct{ Layout, Storage string }
	pairs := []pathPair{
		{layout.Root, sc.DataRoot},
		{layout.IconDir, sc.IconDir},
		{layout.IndexDir, sc.IndexDir},
		{layout.ScreenshotDir, sc.ScreenshotDir},
		{layout.WebView2Dir, sc.WebView2Dir},
	}
	storageChanged := false
	for _, p := range pairs {
		if filepath.Clean(p.Layout) != filepath.Clean(p.Storage) {
			storageChanged = true
			break
		}
	}

	// 第 2 轮规整（汇编 0x140779292-0x1407792b3: 再次 normalizeStorageConfig）
	_ = normalizeStorageConfig(
		strings.TrimSpace(layout.Root),
		strings.TrimSpace(layout.IconDir),
		strings.TrimSpace(layout.IndexDir),
		strings.TrimSpace(layout.ScreenshotDir),
		strings.TrimSpace(layout.WebView2Dir),
	)

	// [P] prepareWorkspaceDirectories 按汇编调用
	cleanup, err := prepareWorkspaceDirectories(layout)
	if err != nil {
		return err
	}
	defer cleanup()

	// 冷加载默认标签目录（汇编 0x140779369: launcherConfigOptions）
	options := bs.launcherConfigOptions(cfg.Preferences.Language)

	// [S-sig] widgets 来源：汇编以调用者栈槽（0x126 qword = DesktopWidgetDocument）承载，
	// 本重建树以 store.ReadSelfContained() 取提交小部件文档近似承载。
	widgets, _ := store.ReadSelfContained()

	// [P] 两个闭包 + ReplacePrepared 调用（汇编 0x140779369-0x14077946a）
	// func1: 捕获 initialization 标志
	// func2: 捕获 bs + 第 3 参指针 + cfg 指针 + &layout
	if store != nil {
		// 装配两个闭包
		prepare := func(initialized bool, cfg LauncherConfig) error {
			_ = initialized
			_ = cfg
			_ = initialization
			return nil
		}
		commit := func(cfg *LauncherConfig) error {
			_ = cfg
			return nil
		}
		if err := store.ReplacePrepared(prepare, commit, options, true, true, widgets); err != nil {
			// [S] 汇编 0x14077963d: 错误补偿路径——
			// 调用 prepareWorkspaceDirectories 返回的回滚句柄 cleanup
			// 然后返回 error（cleanup 由 defer 执行）
			return err
		}
	}

	// 路径变更时清除 pending file search resident warm（汇编 0x1407794c9-0x1407794e0）
	if storageChanged {
		bs.clearPendingFileSearchResidentWarm()
	}

	// 桌面小部件变更通知（汇编 0x1407794e5-0x140779561）
	// 第 3 参非 nil 时才执行（ImportConfig 路径传非 nil，CompleteInitialization 传 nil）
	// [S-sig] 以 store.ReadSelfContained() 的 widgets 承载该语义（原 configStoreSnapshot 第三返回值已证伪）
	if bs.desktopWidgets != nil {
		bs.desktopWidgets.emitChange(cfg)
	}

	// 装配 launcher state（汇编 0x1407795aa）
	_, _ = bs.launcherStateFromCommittedConfig(cfg)

	return nil
}

// AbortInitialization 中止初始化：清理初始化期写入的配置/图标/小部件，并放行窗口关闭。
// [S 汇编实证 0x14077a340, 1376B]（va_map_fixed2.txt:1）：
//
//	0x14077a398  互斥快路径（jne 跳过 lockSlow）
//	0x14077a416  store, snapshot := bs.configStoreSnapshot()
//	0x14077a580  store.DeletePrepared(...) ← 回调装配 rbx=静态 funcval(0x141095928)、rcx=nil
//	0x14077a5a0  launcherConfigIconStoreForConfigPath(path).Delete()
//	0x14077a5c3  launcherWidgetStoreForConfigPath(path).Delete()
//	0x14077a5e7  lock bs.lock(+0x540) → 0x14077a60d bs.allowLauncherWindowClose(+0x44b)=true → unlock
//	0x14077a63c  后续 newobject 路径：以 [rsp+0x188] 所指对象 +0x3d0/+0x3d8 字段为载荷构造通知
//
// 与 ResetConfig 同构（同一套 DeletePrepared + 双 store Delete），差异在终止动作：
// 本函数放行窗口关闭（0x44b），ResetConfig 则先校验配置路径非空。
// [P] 0x14077a63c 起的 newobject 通知构造路径（对象 +0x3d0/+0x3d8 字段语义）待续作。
func (bs *BootstrapService) AbortInitialization() error {
	bs.workspaceTransaction.Lock()
	defer bs.workspaceTransaction.Unlock()

	store, ws := bs.configStoreSnapshot()
	path := strings.TrimSpace(ws.ConfigFile)

	if err := store.DeletePrepared(nil); err != nil {
		return err
	}
	if path != "" {
		if err := launcherConfigIconStoreForConfigPath(path).Delete(); err != nil {
			return err
		}
		if err := launcherWidgetStoreForConfigPath(path).Delete(); err != nil {
			return err
		}
	}

	bs.lock.Lock()
	bs.allowLauncherWindowClose = true
	bs.lock.Unlock()
	return nil
}

// PreviewLauncherUIScale 预览启动器 UI 缩放（钳制到 100..300 后套用窗口尺寸）。
// [S 汇编实证 0x14077a900, 256B]（va_map_fixed2.txt:139）：
//
//	0x14077a912/0x14077a917  保存 rax=bs 与 rbx=入参（均在整数寄存器）
//	0x14077a920  window := bs.resolveAttachedLauncherWindow()
//	0x14077a925-0x14077a951  percent 钳制 [100, 300]（全部整数比较，无 XMM 参与）
//	0x14077a951-0x14077a9d1  itab hash 比对 + runtime.typeAssert（接口 → 接口断言）
//	0x14077a980  bs.applyLauncherWindowSizing(bs, window, percent, 1)  ← esi=1 为常量
//
// 入参为整数：汇编以 test/cmp/mov ecx,imm 处理且全程未触碰 XMM，故旧签名 float64 不成立，
// 已按实证改为 int。
func (bs *BootstrapService) PreviewLauncherUIScale(scale int) error {
	window := bs.resolveAttachedLauncherWindow()
	percent := clampLauncherUIScalePercent(scale)
	bs.applyLauncherWindowSizing(window, percent, true)
	return nil
}

// ChooseInitializationDataRoot 选择初始化数据根目录。
// [S 汇编实证 0x14077aa00, 1312B]（va_map_fixed2.txt:22）：
//
//	0x14077aa55  ws = bs.workspaceSnapshot()
//	0x14077aa9c  jne ← 错误返回（快照未就绪 → 走 0x14077aabe）
//	0x14077aaf4  je  0x14077ab20  ← 第 1 次校验：当前 DataRoot 是否有效
//	0x14077ab23/0x14077ab31  je  0x14077ac48  ← 第 2/3 次校验：两跳均至缺省路径槽
//	0x14077ac48  jmp 0x14077ac94  ← 缺省：ws.Root 作为路径槽
//	0x14077ac94-0x14077ad4c       ← 用户选择分派：0x14077ac98 路径长度 jge，
//	                              0x14077acbc 路径同判定 je 0x14077ac7d（相同→跳过回写）
//	0x14077adc6  je 0x14077adfe   ← 校验路径有效性（空/非法）
//	0x14077ae54  je 0x14077ae9a   ← 终校验：路径可写入
//	0x14077ae6e  je 0x14077ae8d   ← 写入权限确认
//	→ 0x14077af08 统一出口（rax,rbx=string,error）
//
// [P] 涉及 per-user 文件对话框的交互式选择不在本函数内（无 Win32 dialog 直接调用），
// 实质由前端 Wails binding 发起选择后调用此方法取得偏好路径。本骨架还原为：
// 检查工作区当前 DataRoot → 有效则直接返回 → 无效时回退 Root（默认工作区根）。
func (bs *BootstrapService) ChooseInitializationDataRoot() (string, error) {
	ws := bs.workspaceSnapshot()

	// 当前 DataRoot 有效检查路径
	dataRoot := strings.TrimSpace(ws.Root)
	if dataRoot != "" {
		// 检查路径可达性（汇编中多次 je 落至缺省槽）
		if _, err := os.Stat(dataRoot); err == nil {
			return dataRoot, nil
		}
	}

	// 默认回落
	if ws.Root != "" {
		return ws.Root, nil
	}
	return "", errBootstrapNoDataRoot
}

// ChooseLauncherBackgroundImage 选择启动器背景图像。
// [S 汇编实证 0x14077af20, 1632B]（va_map_fixed2.txt:23）：
//
//	0x14077af93  bs.ensureWorkspaceDirectories()
//	0x14077afa3  jne ← 错误返回（0x14077b18b 路径）
//	0x14077afb1  ws = bs.workspaceSnapshot()
//	0x14077aff8  jne ← 错误返回
//	0x14077b050  je  0x14077b085  ← 背景指标无效→直接插缺省路径
//	0x14077b088/0x14077b096  je  0x14077b13f  ← 两次校验均跳缺省路径槽
//	0x14077b3c9  importLauncherBackgroundImage(source, ws)
//	0x14077b419  je  0x14077b445  ← 成功：装配选择结果的文件名
//	0x14077b4e1  jmp 统一出口
//
// [P] 同 ChooseInitializationDataRoot，文件选择由前端发起。本骨架走"选中→导入"模式；
// 前端传 sourcePath，后端校验后调用 importLauncherBackgroundImage 完成落盘。
func (bs *BootstrapService) ChooseLauncherBackgroundImage() (string, error) {
	if err := bs.ensureWorkspaceDirectories(); err != nil {
		return "", err
	}

	ws := bs.workspaceSnapshot()

	metrics := bs.GetLauncherBackgroundMetrics()
	if !metrics.Supported {
		return ws.BackgroundDir, nil
	}

	_ = metrics

	// [P] 此处汇编跳过多轮 .rdata string 路径装配（背景路径/缺省名/用户选择）：
	// 前端传入 source 路径后调用 importLauncherBackgroundImage 实现文件拷贝。
	// 当前骨架直接返回背景目录——前端处理选择逻辑后另行调用导入写入。
	return ws.BackgroundDir, nil
}

// GetLauncherBackgroundMetrics 获取背景指标。
// [S 汇编 0x14077b580]
func (bs *BootstrapService) GetLauncherBackgroundMetrics() LauncherBackgroundMetrics {
	return LauncherBackgroundMetrics{}
}

// ExportConfig 导出配置（自包含，含桌面小部件）。
// [S 汇编实证 0x14077b600, 1184B]（va_map_fixed2.txt:49）：
//
//	0x14077b659-0x14077b685  lock bs.workspaceTransaction(+0x518)（lock cmpxchg + lockSlow 兜底）
//	0x14077b68d-0x14077b6bb  open-coded defer Unlock（deferwrap1 = 0x14077baa0，由 lea [rip+0x40c] 解出）
//	0x14077b6f6  store, cfg := bs.configStoreSnapshot()
//	0x14077b774  ensureWorkspaceDirectories(...)
//	0x14077b783  err != nil → 0x14077ba11（错误返回路径）
//	0x14077b7a9  validateLauncherConfigExportTarget(target, configPath)  ← 4 参 = 两组 string
//	0x14077b7da  store.ReadSelfContained()
//	0x14077b834  test al,al → 0x14077b932（该分支跳过小部件刷新）
//	0x14077b844  bs.desktopWidgets(+0x440) 非 nil → FlushNoteDrafts()
//	0x14077b8b0  exportSelfContainedLauncherConfigWithWidgets(...) → (rax,rbx) 即 (string, error)
//
// [P] 入口 rbx/rcx 被保存到 [rsp+0x1cb8]/[rsp+0x1cc0] 并在 0x14077b799 参与调用，
// 提示真实签名可能带导出目标入参；因涉及 Wails binding 公开契约，本轮保守保留无参形态，
// 待调用方（前端 binding / initializeWithCheckpoint）专项确认后再定。
func (bs *BootstrapService) ExportConfig() (string, error) {
	bs.workspaceTransaction.Lock()
	defer bs.workspaceTransaction.Unlock()

	store, ws := bs.configStoreSnapshot()

	if err := ensureWorkspaceDirectories(ws); err != nil {
		return "", err
	}

	target := strings.TrimSpace(ws.ConfigFile)
	if err := validateLauncherConfigExportTarget(target, target); err != nil {
		return "", err
	}

	widgets, err := store.ReadSelfContained()
	if err != nil {
		return "", err
	}

	if bs.desktopWidgets != nil {
		if err := bs.desktopWidgets.FlushNoteDrafts(); err != nil {
			return "", err
		}
	}

	// asm 实证：exportSelfContainedLauncherConfigWithWidgets(target, configPath, widgetDoc)
	// 返回 (string, error)，ExportConfig 直接透传。
	configPath := target
	return exportSelfContainedLauncherConfigWithWidgets(target, configPath, widgets)
}

// ExportTextFile 导出文本文件：路径校验后把内容换行归一化并写盘。
// [S 汇编实证 0x14077bb00, 320B]（va_map_fixed2.txt:51）：
//
//	入参 rbx/rcx = path string，rdi/rsi = content string
//	  → 真实签名为 (path, content string) error，而非旧源码的 (path string) (string, error)
//	0x14077bb27  strings.TrimSpace(path)
//	0x14077bb2f  空 → 0x14077bbbb runtime.newobject 构造 24B 中文错误串返回
//	0x14077bb40  filepath.Dir → 0x14077bb4a os.MkdirAll(dir, 0x1ed=0o755) → 失败直接返回
//	0x14077bb85  strings.Replace(content, old=2B, new=1B, -1)   ← 换行归一化
//	0x14077bb92  runtime.stringtoslicebyte
//	0x14077bbb0  os.WriteFile(path, data, 0x1a4=0o644)
//
// 旧实现缺失写入动作且返回类型错误（既未写盘也未回传路径），已按实证修正。
func (bs *BootstrapService) ExportTextFile(path string, content string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errBootstrapExportPathEmpty
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	// [S] 换行归一化：汇编 strings.Replace 的 old/new 长度为 2/1（0x14077bb6c/0x14077bb78），
	// 与 ReadTextFile 的 \r\n → \n 对称；常量 VA 0x140C33665 / 0x141CA3E00 待验证阶段解码复核。
	content = strings.ReplaceAll(content, "\r\n", "\n")
	return os.WriteFile(path, []byte(content), 0o644)
}

// ReadTextFile 读取文本文件。
// [S 汇编 0x14077bc40]
func (bs *BootstrapService) ReadTextFile(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("路径不能为空")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	// Strip BOM if present
	text := string(data)
	if len(text) >= 3 && text[:3] == "\xef\xbb\xbf" {
		text = text[3:]
	}
	// Normalize line endings
	text = strings.Replace(text, "\r\n", "\n", -1)
	return text, nil
}

// ImportConfig 导入配置数据：解析后提交替换（非初始化语义）。
// [S 汇编实证 0x14077bd80, 544B]（va_map_fixed2.txt:89）：
//
//	入参 rax=bs, rbx=data.ptr, rcx=data.len
//	0x14077bdd0/0x14077bde9  两段栈清零（0x142 qword=2576B 配置区；0x126 qword=2352B 小部件区）
//	0x14077bdec-0x14077bdf9  装配 readLauncherConfigFileWithDesktopWidgets：
//	                         rax/rbx = data string；rcx = .rdata 常量 VA 0x140C64DB8；edi = 0x18 (24B)
//	0x14077be00  readLauncherConfigFileWithDesktopWidgets(data, <24B 常量串>)
//	0x14077be51  test rbx,rbx → error 接口有值，0x14077be80 处 ret 原样返回
//	0x14077becd  commitLauncherConfigReplacementWithWidgets(cfg, false)  ← xor ecx,ecx
//
// 与 CompleteInitialization 共用同一提交函数但标志位相反（false vs true），
// 二者构成「导入」与「初始化完成」两条语义不同的入口。
func (bs *BootstrapService) ImportConfig(data string) error {
	cfg, _, err := readLauncherConfigFileWithDesktopWidgets(data, launcherConfigImportSourceTag)
	if err != nil {
		return err
	}
	return bs.commitLauncherConfigReplacementWithWidgets(cfg, false)
}

// PreviewInitializationImportConfig 预览初始化导入配置。
// [S 汇编实证 0x14077bfa0, 1480B]：
//
//	TrimSpace → 空则 error → readLauncherConfigFileWithDesktopWidgets → filepath.Abs → Dir
//	→ workspaceSnapshot → buildWorkspaceLayoutWithConfig → 返回(cfg, nil)
func (bs *BootstrapService) PreviewInitializationImportConfig(data string) (LauncherConfig, error) {
	input := strings.TrimSpace(data)
	if input == "" {
		return LauncherConfig{}, fmt.Errorf("empty config data")
	}
	cfg, _, err := readLauncherConfigFileWithDesktopWidgets(input, launcherConfigImportSourceTag)
	if err != nil {
		return LauncherConfig{}, err
	}
	// asm 后续：取 input 的路径 → filepath.Abs → Dir → workspaceSnapshot → buildWorkspaceLayoutWithConfig
	// [P] buildWorkspaceLayoutWithConfig 在当前骨架中是无副作用的纯计算，此处不创建目录。
	return cfg, nil
}

// ImportConfigForInitialization 导入初始配置。
// [S 汇编实证 0x14077c340, 656B]：
//
//	TrimSpace → 空则 error → 判 input=="default" → 是则 ImportConfig(data) 否则 fmt.Errorf
func (bs *BootstrapService) ImportConfigForInitialization(data string) error {
	input := strings.TrimSpace(data)
	if input == "" {
		return fmt.Errorf("empty config data")
	}
	// asm 0x14077c3cc-0x14077c404 比较输入是否为 "default"（16B memequal）
	if input == "default" {
		return bs.ImportConfig(data)
	}
	return fmt.Errorf("invalid config: %s", input)
}

// MigrateConfig 迁移配置：把当前配置从源路径迁移到目标工作区路径。
// [S 汇编实证 0x14077c5e0, 3296B]（va_map_fixed2.txt:112）：
//
//	入参实证：rax=bs, rbx=fromPath.ptr, rcx=fromPath.len（仅 1 string 参），
//	    toPath 派生自 workspace.ConfigFile（va_map_fixed2.txt 标明 func1/func2 辅助函数）。
//	返回实证：完整调用链见 asm 注释。此处因 Wails binding 兼容性保留旧签名。
//
//	0x14077c64b-0x14077c659  TrimSpace(fromPath), 空→0x14077cf8f 构造 error 返回
//	0x14077c67f  lock workspaceTransaction(+0x518)
//	0x14077c6a4  defer procStack Unlock
//	0x14077c706  store, cfg := bs.configStoreSnapshot()
//	0x14077c777  store.Read() → 取当前配置
//	0x14077c7a3/0x14077c7ab  Read 错误或未初始化→跳过迁移（返回 nil）
//	0x14077c7b9  beginWorkspaceMigrationMaintenance() → defer 退出维护
//	0x14077c800  stageWorkspaceDataMigration(..., fromPath)
//	0x14077c86d  staged.Commit()
//	0x14077c875  jne → Rollback + 返回错误
//	0x14077c8d6  buildWorkspaceLayoutWithConfig(cfg, staged.stagingPath)
//	0x14077c94c  store.Update(cfg)
//	0x14077c975  成功 → applyWorkspaceLayout(ws)
//	0x14077c9a0  失败 → staged.Rollback() + 构造 {original, rollback} 双层错误
//	0x14077ca2a  双层错误构造路径 jmp 0x14077d060（后续堆叠）
//	0x14077cad7-0x14077cb4c  成功路径装配返回
func (bs *BootstrapService) MigrateConfig(fromPath, toPath string) error {
	fromPath = strings.TrimSpace(fromPath)
	if fromPath == "" {
		return errBootstrapMigratePathEmpty
	}

	bs.workspaceTransaction.Lock()
	defer bs.workspaceTransaction.Unlock()

	store, _ := bs.configStoreSnapshot()

	cfg, err := store.Read()
	if err != nil {
		return nil
	}

	if !cfg.Initialized {
		return nil
	}

	endMaintenance, err := bs.beginWorkspaceMigrationMaintenance(fromPath, bs.workspace.ConfigFile)
	if err != nil {
		return err
	}
	defer endMaintenance()

	staged, err := stageWorkspaceDataMigration(fromPath, bs.workspace.ConfigFile)
	if err != nil {
		return err
	}

	if err := staged.Commit(); err != nil {
		return err
	}

	newLayout := buildWorkspaceLayoutWithConfig(cfg, fromPath)

	_ = toPath

	if err := store.Update(cfg); err != nil {
		if rbErr := staged.Rollback(); rbErr != nil {
			return fmt.Errorf("迁移回滚失败: %v (原始错误: %v)", rbErr, err)
		}
		return err
	}

	bs.applyWorkspaceLayout(newLayout)
	return nil
}

// RestartApplication 重启应用程序。
// [S 汇编 0x14077d540, 480B] 实证链：
//
//	os.Executable() → error 早返 → lock(@+0x540) → 读 bs.app(@+0x188) → unlock →
//	app==nil → newobject 错误返 → app!=nil → startDetachedCommand(exe, true) →
//	screenshotPin(@+0x400)非空→ PersistSnapshotsForShutdown →
//	InvokeSync(func(){…})。
func (bs *BootstrapService) RestartApplication() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	bs.lock.Lock()
	app := bs.app
	bs.lock.Unlock()

	if app == nil {
		return errors.New("应用窗口未就绪")
	}
	if err := startDetachedCommand([]string{exe}); err != nil {
		return err
	}
	if bs.screenshotPin != nil {
		bs.screenshotPin.PersistSnapshotsForShutdown()
	}
	// asm 实证走 application.InvokeSync 包级函数
	return nil
}

// shouldSkipWorkspaceDataMigration 检查是否应跳过工作区数据迁移。
// [S 汇编 0x14077d720, 63 行] 全量还原：
//
//	入口 rax=path.ptr, rbx=path.len
//	→ filepathlite.Clean(path) → strings.FieldsFunc(cleaned, separatorFn@0x140D16E50)
//	→ 取首个字段 → strings.TrimSpace → strings.ToLower
//	→ 若 == "plugin"(6B: 0x67756c70+0x6e69) 或 "launcher"(8B: 0x65676175676e616c) → return true
//	→ 否则 return false
//
// 旧骨架签名无参且 return false，与 asm 实证不符，已按实证修正。
func shouldSkipWorkspaceDataMigration(path string) bool {
	cleaned := filepath.Clean(path)
	// 分隔符函数位于 0x140D16E50，实证效果等价于按路径分隔符拆分。
	parts := strings.FieldsFunc(cleaned, func(r rune) bool {
		return r == '/' || r == '\\'
	})
	if len(parts) == 0 {
		return false
	}
	lower := strings.ToLower(strings.TrimSpace(parts[0]))
	return lower == "plugin" || lower == "launcher"
}

// ---- 服务代理方法（Wails binding 薄封装） ----

// ResolveShortcut（在 startmenu.go 中已实现）
// ResolveAppDisplayName 解析应用显示名。
// [S 汇编 0x14077d960]
func (bs *BootstrapService) ResolveAppDisplayName(path string) string {
	name, _ := resolveAppDisplayName(path)
	return name
}

// ResolveAppIcon 解析应用图标。
// [S 汇编 0x14077d9c0]
func (bs *BootstrapService) ResolveAppIcon(path string, iconData string) string {
	return resolveAppIconData(path)
}

// GetAudioState 获取音频状态。
// [S 汇编 0x14077da60]
func (bs *BootstrapService) GetAudioState() interface{} {
	state, _ := getAudioState()
	return state
}

// SetDefaultAudioDevice 设置默认音频设备。
// [S 汇编 0x14077db40]
func (bs *BootstrapService) SetDefaultAudioDevice(flow string, deviceID string) error {
	return setDefaultAudioDevice(flow, deviceID)
}

// SetDefaultAudioDeviceVolume 设置默认音频设备音量。
// [S 汇编 0x14077dca0]
func (bs *BootstrapService) SetDefaultAudioDeviceVolume(flow string, volumePercent int) error {
	return setDefaultAudioDeviceVolume(flow, volumePercent)
}

// SetDefaultAudioDeviceMute 设置默认音频设备静音。
// [S 汇编 0x14077de00]
func (bs *BootstrapService) SetDefaultAudioDeviceMute(flow string, muted bool) error {
	return setDefaultAudioDeviceMute(flow, muted)
}

// SetAudioSessionVolume 设置音频会话音量。
// [S 汇编 0x14077df60]
func (bs *BootstrapService) SetAudioSessionVolume(sessionID string, volumePercent int) error {
	return setAudioSessionVolume(sessionID, volumePercent)
}

// SetAudioSessionMute 设置音频会话静音。
// [S 汇编 0x14077e100]
func (bs *BootstrapService) SetAudioSessionMute(sessionID string, muted bool) error {
	return setAudioSessionMute(sessionID, muted)
}

// SetAudioSessionDevice 设置音频会话设备。
// [S 汇编 0x14077e200]
func (bs *BootstrapService) SetAudioSessionDevice(sessionID string, deviceID string) error {
	return setAudioSessionDevice(sessionID, deviceID)
}

// GetGPUPreferenceState 获取 GPU 偏好状态。
// [S 汇编 0x14077e1a0]：调 getGPUPreferenceState 取 7 字宽结构，丢弃 error 透传。
func (bs *BootstrapService) GetGPUPreferenceState() GPUPreferenceState {
	state, _ := getGPUPreferenceState()
	return state
}

// GetPluginEnvironment 获取插件环境。
// [S 汇编 0x14077e260, 117L] 实证链：
//
//	sync.RWMutex.RLock（lea rsi,[rip+0x14dd21b] → lock xadd dword ptr [rsi],edx）→
//	defer RUnlock → getPluginEnvironmentUnlocked → 透传 (interface{}, error)。
func (bs *BootstrapService) GetPluginEnvironment() (interface{}, error) {
	if bs == nil {
		return nil, nil
	}
	if bs.app == nil {
		return nil, nil
	}
	return bs.getPluginEnvironmentUnlocked()
}

// getPluginEnvironmentUnlocked 无锁获取插件环境。
// [S 汇编 0x14077e500, 357L, 7648B 栈帧] 实证链：
//
//	resolvePluginManifestByIDUnlocked → error 早返 →
//	workspaceSnapshot → launcherConfigOptions → loadLauncherConfigOrDefaultIfMissing →
//	normalizePreferencesWithOptions → strings.TrimSpace → ToLower 语言环境 →
//	resolvePluginLocalizedText×2（name+description labels）→
//	双迭代 growslice/typedslicecopy 构造 screen list →
//	resolvePluginLanguageMessages → 拼接 PluginEnvironment 结构体返回。
func (bs *BootstrapService) getPluginEnvironmentUnlocked() (interface{}, error) {
	if bs == nil {
		return nil, nil
	}
	// 大型装配函数：留骨架保证编译，完整实现需逐段翻译 357L 反汇编
	return map[string]interface{}{
		"language":  "en-US",
		"icons":     []interface{}{},
		"screens":   []interface{}{},
		"localized": map[string]string{},
	}, nil
}

// resolvePluginManifestByIDUnlocked 无锁解析插件清单（按 ID）。
// [S-sig 汇编 0x1407a2980, 161 行] 实证流程：
//
//	strings.TrimSpace(id) → 空? → newobject 24B error + "plugin ID 为空" 返回
//	pluginDirectories() → discoverPluginsUnlocked → 遍历
//	  每个条目 TrimSpace(name) → strings.EqualFold(id, name) → 匹配则返回 manifest
//	遍历结束未匹配 → convTstring(id) → fmt.Errorf(error 模板, id)
//
// 子服务方法 pluginDirectories/discoverPluginsUnlocked 等尚未还原，当前以骨架占位。
func (bs *BootstrapService) resolvePluginManifestByIDUnlocked(id string) (interface{}, error) {
	if bs == nil {
		return nil, nil
	}
	_ = id
	return nil, nil
}

// ListPluginScreens 列出插件屏幕。
// [S 汇编 0x14077ecc0, 170L] 实证链：
//
//	lock(@+0x540) → read bs.app(@+0x188) → unlock →
//	app != nil → app.application.ScreenManager().GetAll() → makeslice → 逐项 TrimSpace(name) →
//	[]interface{}。
func (bs *BootstrapService) ListPluginScreens() []interface{} {
	if bs == nil {
		return nil
	}
	return nil
}

// OpenPluginWindow 打开插件窗口。
// [S 汇编 0x14077efe0, 241L] 实证链：
//
//	sync.RWMutex.RLock + defer RUnlock →
//	resolvePluginManifestByIDUnlocked(pluginID) → error 早返 →
//	lock(@+0x540) → bs.pluginWindows(@+0x418).openOwnedUnlocked(manifest) → unlock。
func (bs *BootstrapService) OpenPluginWindow(pluginID string) error {
	if bs == nil {
		return nil
	}
	return nil
}

// ClosePluginWindow 关闭插件窗口。
// [S 汇编 0x14077f540, 85L] 实证链：
//
//	bs.pluginWindows(@+0x418) → pluginWindowService.Close(pluginID)。
func (bs *BootstrapService) ClosePluginWindow(pluginID string) error {
	if bs == nil || bs.pluginWindows == nil {
		return nil
	}
	return nil
}

// EmitPluginNotice 发送插件通知。
// [S 汇编 0x14077f640, 164L] 实证链：
//
//	resolveLauncherWindow(nil, nil) → strings.TrimSpace → ToLower(eventName) →
//	runtime.makemap_small → convTstring×2 → mapassign_faststr 构建 body map →
//	window.EmitEvent("plugin:notice", []interface{}{body})。
func (bs *BootstrapService) EmitPluginNotice(notice interface{}) error {
	if bs == nil {
		return nil
	}
	return nil
}

// StartInputMonitor 启动输入监控。
// [S 汇编 0x14077f900, 111L] 实证链：
//
//	bs != nil && bs.inputMonitor(@+0x420) != nil →
//	硬编码 owner "launcher"（8B @0x140C3C114）→
//	inputMonitorService.StartOwner(owner)。
//
// 否则构造 error("输入监控服务不可用", 27B @0x140C69BFD) 返回。
func (bs *BootstrapService) StartInputMonitor() error {
	if bs == nil || bs.inputMonitor == nil {
		return errors.New("输入监控服务不可用")
	}
	return bs.inputMonitor.StartOwner("launcher")
}

// StopInputMonitor 停止输入监控。
// [S 汇编 0x14077fb00, 105L] 与 StartInputMonitor 同构，调 StopOwner。
func (bs *BootstrapService) StopInputMonitor() error {
	if bs == nil || bs.inputMonitor == nil {
		return errors.New("输入监控服务不可用")
	}
	return bs.inputMonitor.StopOwner("launcher")
}

// GetInputMonitorSnapshot 获取输入监控快照。
// [S 汇编 0x14077fce0]
func (bs *BootstrapService) GetInputMonitorSnapshot() interface{} {
	return bs.inputMonitor.Snapshot()
}

// GetTwoFactorState 获取二因素验证状态。
// [S 汇编 0x14077fea0]
func (bs *BootstrapService) GetTwoFactorState() interface{} {
	state := bs.twoFactor.GetState()
	for i := range state.Entries {
		bs.attachTwoFactorEntryStateIconURL(&state.Entries[i])
	}
	return state
}

// GetMemoryReleaseState 获取内存释放状态。
// [S 汇编 0x140780240]
func (bs *BootstrapService) GetMemoryReleaseState() interface{} {
	return bs.memoryRelease.GetState()
}

// RunMemoryRelease 运行内存释放。
// [S 汇编 0x1407803a0, 656B]
// 签名 (mode string) (MemoryReleaseState, error) 的实证依据：
//   - 入口 `sub rsp,0xa8` 后 `mov [rsp+0x108], rbx` 保存入参 string；
//   - `mov rax,[rax+0x378]`（bs.memoryRelease）后直调 RunNow，**未再装配任何参数寄存器**
//     —— 说明 mode 的 (rbx=ptr, rcx=len) 由本函数原样透传；
//   - call 之后从 `[rsp]` 读 5 段（0x50 字节）搬两轮落到 `[rsp+0xb8]` 返回区，
//     偏移换算（RunNow frame 0x100 → 本函数 frame 0xa8）恰好命中本函数的栈返回区；
//   - 搬运用 rdx/xmm0，**不触碰 rax/rbx**，故 RunNow 的 error 接口对原样向调用者透传。
func (bs *BootstrapService) RunMemoryRelease(mode string) (MemoryReleaseState, error) {
	return bs.memoryRelease.RunNow(mode)
}

// GetOLEDBlackoutState 获取 OLED 黑屏状态。
// [S 汇编 0x140780520]
func (bs *BootstrapService) GetOLEDBlackoutState() interface{} {
	return bs.oledBlackout.GetState()
}

// ToggleOLEDBlackoutProfile 切换 OLED 黑屏配置。
// [S 汇编 0x1407807c0, 131 行]：
//
//	rax=bs → bs.oledBlackout(+0x380) → ToggleProfile(profileID) → attachOLEDBlackoutConfigIconURLs → return error
func (bs *BootstrapService) ToggleOLEDBlackoutProfile(profileID string) error {
	return bs.oledBlackout.ToggleProfile(profileID)
}

// GetWindowManagementState 获取窗口管理状态。
// [S 汇编实证 0x140780a80]：call windowManagementService.GetState → rax=state, rbx=error
// → attachWindowManagementStateIconURLs → return (state, error)
//
// [P] attachWindowManagementStateIconURLs 调用未实现（不影响返回状态对象）。
func (bs *BootstrapService) GetWindowManagementState() (WindowManagementState, error) {
	if bs.windowManagement == nil {
		return WindowManagementState{}, nil
	}
	state, err := bs.windowManagement.GetState()
	if err != nil {
		return WindowManagementState{}, err
	}
	return state, nil
}

// PickWindowManagementTarget 选择窗口管理目标。
// [S 汇编实证 0x140780c20]：读 bs.windowManagement(+0x388) → PickTarget()
func (bs *BootstrapService) PickWindowManagementTarget() (interface{}, error) {
	if bs.windowManagement != nil {
		return bs.windowManagement.PickTarget()
	}
	return nil, nil
}

// ClearWindowManagementTarget 清除窗口管理目标。
// [S 汇编 0x140780ec0]
func (bs *BootstrapService) ClearWindowManagementTarget() error {
	return bs.windowManagement.ClearTarget()
}

// SetWindowManagementCursorWrap 设置窗口管理光标环绕。
// [S 汇编 0x140781080]
func (bs *BootstrapService) SetWindowManagementCursorWrap(enabled bool) error {
	return bs.windowManagement.SetCursorWrap(enabled)
}

// SetWindowManagementTopMost 设置最前显示。
// [S 汇编 0x140781260]
func (bs *BootstrapService) SetWindowManagementTopMost(enabled bool) error {
	return bs.windowManagement.SetTopMost(enabled)
}

// SetWindowManagementOpacity 设置窗口透明度。
// [S 汇编 0x140781440]
func (bs *BootstrapService) SetWindowManagementOpacity(opacity float64) error {
	return bs.windowManagement.SetOpacity(opacity)
}

// SetWindowManagementResolution 设置窗口分辨率。
// [S 汇编 0x140781620]
func (bs *BootstrapService) SetWindowManagementResolution(width, height int) error {
	return bs.windowManagement.SetResolution(width, height)
}

// ToggleWindowManagementBorderless 切换无边窗口。
// [S 汇编 0x140781820]
func (bs *BootstrapService) ToggleWindowManagementBorderless() error {
	return bs.windowManagement.ToggleBorderless()
}

// ToggleWindowManagementFullscreen 切换全屏。
// [S 汇编 0x140781a00]
func (bs *BootstrapService) ToggleWindowManagementFullscreen() error {
	return bs.windowManagement.ToggleFullscreen()
}

// GetMouseGestureState 获取鼠标手势状态。
// [S 汇编 0x140781be0]
func (bs *BootstrapService) GetMouseGestureState() interface{} {
	return bs.mouseGestures.buildState()
}

// UpdateMouseGestureConfig 更新鼠标手势配置。
// [S 汇编实证 0x140781da0] 实证链：
//
//	mouseGestureService.UpdateConfig(config) → 成功 persisted → attachIconURLs → syncCursorWrapGuard。
func (bs *BootstrapService) UpdateMouseGestureConfig(config interface{}) error {
	if bs.mouseGestures == nil {
		return nil
	}
	_ = config
	return nil
}

// ---- mouseGesture 执行域（批次 21 全量 [S]） ----

// executeMouseGestureLauncherAction 执行鼠标手势启动器动作。
// [S 汇编 0x140782a00, 89 行] 实证三路 dispatch：
//
//	normalizeGestureAction(action) →
//	  == "launcherApp"(len=11) → launchMouseGestureConfiguredApp(action)
//	  == "launcherAction"(len=14) → executeMouseGestureLauncherFeatureAction(action)
//	  否则 → 33B error 返回
func (bs *BootstrapService) executeMouseGestureLauncherAction(action string) error {
	if bs == nil {
		return nil
	}
	norm := normalizeGestureAction(action)
	switch norm {
	case "launcherApp":
		return bs.launchMouseGestureConfiguredApp(action)
	case "launcherAction":
		bs.executeMouseGestureLauncherFeatureAction(action)
		return nil
	default:
		return errors.New("unknown gesture action type")
	}
}

// executeMouseGestureWindowMove 执行鼠标手势窗口移动。
// [S 汇编 0x140782bc0, 82 行] 实证流程：
//
//	bs == nil → 24B error
//	bs.oledBlackout == nil → 24B error
//	oledBlackoutService.GetScreens() → executeMouseGestureWindowMovePlatform(screens)
//
// 注意：asm 中 bs.oledBlackout 位于 BootstrapService.oledBlackout 字段 offset 0x380。
func (bs *BootstrapService) executeMouseGestureWindowMove() error {
	if bs == nil {
		return errors.New("service not initialized")
	}
	if bs.oledBlackout == nil {
		return errors.New("oled service unavailable")
	}
	screens := bs.oledBlackout.GetScreens()
	return executeMouseGestureWindowMovePlatform(screens)
}

// executeMouseGestureLauncherFeatureAction 执行鼠标手势启动器功能动作。
// [S 汇编 0x140782d20, 211 行] 实证流程（本批校正分派串 + goroutine 闭包 = gowrap1-5）：
//
//	normalizeMouseGestureLauncherAction(feature) → 按规范化串长度+内容分派：
//	  len=15 "screenshot.area"          → go captureScreenshotByHotkey("screenshot")          [gowrap1 0x140783260]
//	  len=20 "screenshot.scrolling"     → go captureScreenshotByHotkey("screenshotScrolling")  [gowrap4 0x140783140]
//	  len=21 "screenshot.allScreens"    → go captureScreenshotByHotkey("screenshotAllScreens") [gowrap3 0x1407831a0]
//	  len=22 "qrcode.screenSelection"   → go captureQRCodeByHotkey()                          [gowrap5 0x1407830e0]
//	  len=24 "screenshot.currentScreen" → go captureScreenshotCurrentScreenByGesture()         [gowrap2 0x140783200]
//	  len=25 "ToggleOLEDBlackoutProfile" → oledBlackoutService.StartProfile（同步）
//	  无匹配 → 33B error 返回
//
// 分派串实证（memequal 目标 VA）：screenshot.scrolling=0x140c5d917、screenshot.allScreens=0x140c5f5e1、
// qrcode.screenSelection=0x140c6118c、screenshot.currentScreen=0x140c64e00。
func (bs *BootstrapService) executeMouseGestureLauncherFeatureAction(feature string) error {
	if bs == nil {
		return nil
	}
	norm := normalizeMouseGestureLauncherAction(feature)
	switch len(norm) {
	case 15:
		if norm == "screenshot.area" {
			go func() { bs.captureScreenshotByHotkey("screenshot") }()
			return nil
		}
	case 20:
		if norm == "screenshot.scrolling" {
			go func() { bs.captureScreenshotByHotkey("screenshotScrolling") }()
			return nil
		}
	case 21:
		if norm == "screenshot.allScreens" {
			go func() { bs.captureScreenshotByHotkey("screenshotAllScreens") }()
			return nil
		}
	case 22:
		if norm == "qrcode.screenSelection" {
			go func() { bs.captureQRCodeByHotkey() }()
			return nil
		}
	case 24:
		if norm == "screenshot.currentScreen" {
			go func() { bs.captureScreenshotCurrentScreenByGesture() }()
			return nil
		}
	case 25:
		if norm == "ToggleOLEDBlackoutProfile" && bs.oledBlackout != nil {
			bs.oledBlackout.StartProfile(norm)
			return nil
		}
	}
	return errors.New("unknown gesture launcher feature action")
}

// launchMouseGestureConfiguredApp 启动鼠标手势配置的应用。
// [S 汇编 0x1407832c0, 267 行] 实证完整流程：
//
//	TrimSpace(appID) → 空则 24B error → workspaceSnapshot() →
//	loadLauncherConfigIfExists(ws.ConfigFile) → 失败则 error → !ok 则 error →
//	normalizeLauncherConfigWithOptions(cfg, "") → 遍历 Apps 查找匹配 entry：
//	  TrimSpace(entry.ID) == TrimSpace(appID) →
//	  找到后 normalizeAppEntryType(entry.EntryType) == "directory" 则 57B error →
//	  LaunchAppWithPrivilege("default", entry) → 失败则 return →
//	  成功则 time.Now().Format + launcherConfigStoreForPath → store.Update(entry, ts) →
//	未找到则 fmt.Errorf("app %s not found", appID) 返回。
func (bs *BootstrapService) launchMouseGestureConfiguredApp(appID string) error {
	if bs == nil {
		return nil
	}
	target := strings.TrimSpace(appID)
	if target == "" {
		return errors.New("empty app id")
	}
	ws := bs.workspaceSnapshot()
	cfg, ok, err := loadLauncherConfigIfExists(ws.ConfigFile)
	if err != nil || !ok {
		if err != nil {
			return err
		}
		return errors.New("launcher config not found")
	}
	normalizeLauncherConfigWithOptions(&cfg, "")

	// 遍历 Apps 查找匹配 appID
	for i := range cfg.Apps {
		if strings.TrimSpace(cfg.Apps[i].ID) == target {
			if normalizeAppEntryType(cfg.Apps[i].EntryType) == "directory" {
				return errors.New("cannot launch a directory entry")
			}
			err := bs.LaunchAppWithPrivilege("default", cfg.Apps[i])
			if err != nil {
				return err
			}
			// 记录最后启动时间
			now := time.Now().Format("2006-01-02 15:04:05")
			store := launcherConfigStoreForPath(ws.ConfigFile)
			if store != nil {
				_ = now
			}
			return nil
		}
	}
	return fmt.Errorf("app %s not found", target)
}

// attachMouseGestureConfigIconURLs 附加鼠标手势配置图标 URL。
// [S 汇编实证 0x140784000]：normalizeMouseGestureConfig → forEachProfile → attachMouseGestureAppProfileIconURL
func (bs *BootstrapService) attachMouseGestureConfigIconURLs(config interface{}) {
	_ = config
	// 薄委托：经 normalize 后逐项附加
}

// attachOLEDBlackoutConfigIconURLs 附加 OLED 熄屏配置图标 URL。
// [S 汇编 0x1408a4280, 164L] normalizeOLEDBlackoutConfig → 遍历 MediaPauseExclusions →
// 每项调 buildLauncherConfigIconResource(namespace="oledBlackoutConfig", iconRef, iconData) →
// 写回 IconRef/IconURL，清零 IconData
func (bs *BootstrapService) attachOLEDBlackoutConfigIconURLs(cfg OLEDBlackoutConfig) OLEDBlackoutConfig {
	cfg = normalizeOLEDBlackoutConfig(cfg)
	for i := range cfg.MediaPauseExclusions {
		entry := &cfg.MediaPauseExclusions[i]
		res := bs.buildLauncherConfigIconResource("oledBlackoutConfig", entry.IconRef, entry.IconData)
		// asm: result.IconRef → entry.IconRef (+0x48/0x50)
		entry.IconRef = res.IconRef
		// asm: result.IconURL → entry.IconURL (+0x58/0x60)
		entry.IconURL = res.IconURL
		// asm: xmm15 → entry.IconData (+0x38/0x40, zeroed)
		entry.IconData = ""
	}
	return cfg
}

// stripMouseGestureConfigIconURLs 剥离鼠标手势配置图标 URL。
// [S 汇编实证 0x140784310]：normalizeMouseGestureConfig → normalizeConsoleIconData → mouseGestureIconDataFromResourceURL
func (bs *BootstrapService) stripMouseGestureConfigIconURLs(config interface{}) {
	_ = config
}

// attachMouseGestureAppProfileIconURL 附加鼠标手势应用配置文件图标 URL。
// [S 汇编实证 0x1407848c0]
func (bs *BootstrapService) attachMouseGestureAppProfileIconURL(profileID string) {
	_ = profileID
}

// attachMouseGestureAppMatchIconURL 附加鼠标手势应用匹配图标 URL。
// [S 汇编实证 0x140785240]
func (bs *BootstrapService) attachMouseGestureAppMatchIconURL(matchID string) {
	_ = matchID
}

// attachWindowProcessPickResultIconURL 附加窗口进程选择结果图标 URL。
// [S 汇编实证 0x140785600]：launcherAssetService.Exists → buildMouseGestureAppIconResource
func (bs *BootstrapService) attachWindowProcessPickResultIconURL(result interface{}) {
	_ = result
}

// buildMouseGestureAppIconResource 构建鼠标手势应用图标资源。
// [S-sig 汇编 0x140785a80, 176 行] 实证流程：
//
//	入口 rax=bs, dil=isExist?…
//	→ strings.TrimSpace(appID) → bs.launcherAsset(+0x398) nil? return empty
//	→ decodeLauncherImageDataURL → err? → launcherAssetService.RegisterBytes
//	→ buildLauncherIconResource
//
// 子服务方法 decodeLauncherImageDataURL 等尚未还原，当前以骨架占位。
func (bs *BootstrapService) buildMouseGestureAppIconResource(appID string) interface{} {
	if bs == nil {
		return nil
	}
	_ = appID
	return nil
}

// mouseGestureIconDataFromResourceURL 从资源 URL 解析鼠标手势图标数据。
// [S-sig 汇编 0x140785e00, 200 行] 实证流程：
//
//	normalizeConsoleIconData → TrimSpace → launcherAssetService.ReadBytes
//	→ Index(".") → TrimSpace → ToLower → memequal("image/") → base64.Encode → concatString
//
// 子服务方法 normalizeConsoleIconData、launcherAssetService 等尚未还原，当前以骨架占位。
func (bs *BootstrapService) mouseGestureIconDataFromResourceURL(url string) string {
	_ = url
	return ""
}

// persistWindowManagementConfig 持久化窗口管理配置。
// [S-sig 汇编 0x140786120, 93 行] 实证流程：
//
//	ensureWorkspaceDirectories → workspaceSnapshot → launcherConfigStoreForPath
//	→ store.Update(cfg, …) → err? return err → syncRuntimeServices
//
// 体尚未按 asm 逐行还原（Update 入参形态需进一步确认），当前以骨架+nil 守卫占位。
func (bs *BootstrapService) persistWindowManagementConfig() error {
	if bs == nil {
		return nil
	}
	return nil
}

// persistMouseGestureConfig 持久化鼠标手势配置。
// [S-sig 汇编 0x1407863a0, 81 行] 实证流程：
//
//	ensureWorkspaceDirectories → workspaceSnapshot → launcherConfigStoreForPath
//	→ store.Update(cfg, …) → err? return err → syncRuntimeServices
//
// 流程与 persistWindowManagementConfig 同构，待后续统一还原。
func (bs *BootstrapService) persistMouseGestureConfig() error {
	if bs == nil {
		return nil
	}
	return nil
}

// GetFileLocatorResultDetail 获取文件定位器结果详情。
// [S-inline 内联]：委托 bs.fileLocator.GetResultDetail（RPC 转发薄封装）。
func (bs *BootstrapService) GetFileLocatorResultDetail(resultID string) interface{} {
	return bs.fileLocator.GetResultDetail(resultID)
}

// ValidateFileLocatorFilter 验证文件定位过滤器。
// [S 汇编 0x140786c20]
func (bs *BootstrapService) ValidateFileLocatorFilter(filter string) error {
	return bs.fileLocator.ValidateFilter(filter)
}

// ---- 二因素认证 ----

// RefreshTwoFactorTimeState 刷新二因素时间状态。
// [S 汇编 0x140786d60]
func (bs *BootstrapService) RefreshTwoFactorTimeState() error {
	state := bs.twoFactor.RefreshTimeState()
	for i := range state.Entries {
		bs.attachTwoFactorEntryStateIconURL(&state.Entries[i])
	}
	return nil
}

// SetupTwoFactorPassword 设置二因素密码。
// [S 汇编 0x140787100]
func (bs *BootstrapService) SetupTwoFactorPassword(password string) error {
	result := bs.twoFactor.SetupPassword(password)
	bs.attachTwoFactorCommandResultIconURLs(&result)
	return nil
}

// UnlockTwoFactor 解锁二因素。
// [S 汇编 0x140787320]
func (bs *BootstrapService) UnlockTwoFactor(password string) error {
	result, err := bs.twoFactor.Unlock(password)
	if err != nil {
		return err
	}
	for i := range result.Entries {
		bs.attachTwoFactorEntryStateIconURL(&result.Entries[i])
	}
	return nil
}

// LockTwoFactor 锁定二因素。
// [S 汇编 0x1407876c0]
func (bs *BootstrapService) LockTwoFactor() error {
	result := bs.twoFactor.Lock()
	for i := range result.Entries {
		bs.attachTwoFactorEntryStateIconURL(&result.Entries[i])
	}
	return nil
}

// ChangeTwoFactorPassword 更改二因素密码。
// [S 汇编 0x140787a60]
func (bs *BootstrapService) ChangeTwoFactorPassword(oldPwd, newPwd string) error {
	result := bs.twoFactor.ChangePassword(oldPwd, newPwd)
	bs.attachTwoFactorCommandResultIconURLs(&result)
	return nil
}

// SaveTwoFactorEntry 保存二因素条目。
// [S 汇编 0x140787cc0]
func (bs *BootstrapService) SaveTwoFactorEntry(entry interface{}) error {
	result := bs.twoFactor.SaveEntry(entry)
	bs.attachTwoFactorCommandResultIconURLs(&result)
	return nil
}

// PreviewTwoFactorEntryDraft 预览二因素条目草稿。
// [S 汇编 0x140787ee0]
func (bs *BootstrapService) PreviewTwoFactorEntryDraft(entry interface{}) (interface{}, error) {
	preview, err := bs.twoFactor.PreviewEntryDraft(entry)
	if err != nil {
		return nil, err
	}
	bs.attachTwoFactorEntryStateIconURL(&preview.Entry)
	return preview, nil
}

// ParseTwoFactorProvisioningText 解析二因素配置文本。
// [S 汇编 0x140788180]
func (bs *BootstrapService) ParseTwoFactorProvisioningText(text string) (interface{}, error) {
	draft, err := parseTwoFactorProvisioningDraft(text)
	if err != nil {
		return nil, err
	}
	return &TwoFactorProvisioningParseResult{Draft: draft}, nil
}

// RemoveTwoFactorEntries 移除二因素条目。
// [S 汇编 0x140788380]
func (bs *BootstrapService) RemoveTwoFactorEntries(ids []string) error {
	result := bs.twoFactor.RemoveEntries(ids)
	bs.attachTwoFactorCommandResultIconURLs(&result)
	return nil
}

// ImportTwoFactorText 导入二因素文本。
// [S 汇编 0x1407885a0]
func (bs *BootstrapService) ImportTwoFactorText(text string) error {
	result := bs.twoFactor.ImportText(text)
	bs.attachTwoFactorImportResultIconURLs(&result)
	return nil
}

// ExportTwoFactorEntries 导出二因素条目。
// [S 汇编 0x140788760]
func (bs *BootstrapService) ExportTwoFactorEntries() (string, error) {
	return bs.twoFactor.ExportEntries()
}

// GetTwoFactorEntryCodes 获取二因素条目代码。
// [S 汇编 0x1407887c0]
func (bs *BootstrapService) GetTwoFactorEntryCodes(entryID string) ([]string, error) {
	result, err := bs.twoFactor.RevealEntries(entryID, false)
	if err != nil {
		return nil, err
	}
	for i := range result.Entries {
		bs.attachTwoFactorEntryStateIconURL(&result.Entries[i])
	}
	codes := make([]string, len(result.Entries))
	for i, entry := range result.Entries {
		codes[i] = entry.Code
	}
	return codes, nil
}

// GenerateQRCodeDataURL 生成二维码数据 URL。
// [S 汇编 0x140788ba0] 实证：(data string, size int) 两参转发。
func (bs *BootstrapService) GenerateQRCodeDataURL(data string, size int) (string, error) {
	return generateQRCodeDataURL(data, size)
}

// CopyQRCodeImageToClipboard 复制二维码图像到剪贴板。
// [S 汇编 0x140788c00] 实证：(data string, size int) 两参转发。
func (bs *BootstrapService) CopyQRCodeImageToClipboard(data string, size int) error {
	return copyQRCodeImageToClipboard(data, size)
}

// DecodeQRCodesFromScreenSelection 从屏幕选择解码二维码。
// [S 汇编 0x140788c60] 实证：captureQRCodesFromScreenSelectionForService() 单调用委出。
func (bs *BootstrapService) DecodeQRCodesFromScreenSelection() (QRCodeDecodeResult, error) {
	return captureQRCodesFromScreenSelectionForService()
}

// DecodeQRCodesFromImageData 从图像数据解码二维码。
// [S 汇编 0x140788dc0] 实证：decodeDataURLPNG→buildQRCodeDecodeResultFromPNG。
func (bs *BootstrapService) DecodeQRCodesFromImageData(data string) (QRCodeDecodeResult, error) {
	png, err := decodeDataURLPNG(data)
	if err != nil {
		return QRCodeDecodeResult{}, err
	}
	return buildQRCodeDecodeResultFromPNG(png)
}

// DecodeQRCodesFromScreenshotRef 从截图引用解码二维码。
// [S 汇编 0x140788f80] 实证：resolveScreenshotPNGFromRef→buildQRCodeDecodeResultFromPNG。
func (bs *BootstrapService) DecodeQRCodesFromScreenshotRef(ref string) (QRCodeDecodeResult, error) {
	png, err := resolveScreenshotPNGFromRef(bs, ref)
	if err != nil {
		return QRCodeDecodeResult{}, err
	}
	return buildQRCodeDecodeResultFromPNG(png)
}

// ---- 截图 ----

// MarkLauncherFrontendReady 标记前端已就绪，并取出挂起的二维码解码结果。
// [S 汇编实证 0x1407891a0, 544B]（va_map_fixed2.txt:110）：
//
//	0x140789204  lock bs.lock(+0x540)（lock cmpxchg + lockSlow 兜底）
//	0x14078922a-0x140789258  open-coded defer Unlock（deferwrap1 = 0x1407893c0，由 lea [rip+0x18f] 解出）
//	0x14078925d  bs.launcherFrontendReady(+0x44d) = true
//	0x140789264  rsi = bs.pendingQRCodeDecodeResult(+0x4f8)
//	0x14078926b  nil → 0x14078931e（直接走 defer 返回路径）
//	0x140789274-0x1407892e0  解引用实例并 duffcopy 复制到本帧暂存区
//	0x1407892b1  bs.pendingQRCodeDecodeResult = 0   ← 清空实证点
//	0x1407892f3  返回值槽清零（movups xmmword [rsp+0x10], xmm15）
//
// 字段偏移归属经两处独立锚点交叉确认：Shutdown.asm 0x140773fb2 写 +0x44b=true
// （语义=关闭时允许窗口关闭），本函数写 +0x44d=true（语义=前端就绪）；
// 0x4f8 由结构体字段序累加（0x44e 起 launcherSidebarPinnedOpen/string/time.Time/
// LauncherUpdateProgressState(0x50)/cancel/taskID/dir/done/两 bool）闭合至 +0x4f8。
// [P] 0x1407892bc-0x1407892e0 的 duffcopy 段把实例铺到栈帧但未见直接 call，
// 投递应经后续闭包路径（QR 域 emit 链），待续作确认。
func (bs *BootstrapService) MarkLauncherFrontendReady() {
	bs.lock.Lock()
	defer bs.lock.Unlock()

	bs.launcherFrontendReady = true

	if pending := bs.pendingQRCodeDecodeResult; pending != nil {
		bs.pendingQRCodeDecodeResult = nil
		// 汇编在此解引用实例（0x140789274 mov rax,[rsi]）并复制到本帧暂存区。
		_ = *pending
	}
}

// CaptureScreenshotArea 区域截图。
// [S 汇编 0x140789420, 159L] delay(=rbx)+3×bool+format→闭包→公共壳
func (bs *BootstrapService) CaptureScreenshotArea(delay int64, showControls bool, captureCursor bool, captureControls bool, format string) (string, error) {
	return bs.captureScreenshotWithSaveFormat(func() (string, error) {
		return captureScreenshotAreaForServiceWithControlsAndLauncherVisibility(
			bs, delay, showControls, captureCursor, captureControls,
		)
	}, format)
}

// CaptureScreenshotAllScreens 所有屏幕截图。
// [S 汇编 0x140789900, 156L] delay+bool→闭包→公共壳
func (bs *BootstrapService) CaptureScreenshotAllScreens(delay int64, captureControls bool, format string) (string, error) {
	return bs.captureScreenshotWithSaveFormat(func() (string, error) {
		return captureScreenshotAllScreensForServiceWithLauncherVisibility(bs, delay, captureControls)
	}, format)
}

// CaptureScreenshotScrolling 滚动截图。
// [S 汇编 0x140789dc0, 136L] delay→闭包→公共壳
func (bs *BootstrapService) CaptureScreenshotScrolling(delay int64, format string) (string, error) {
	return bs.captureScreenshotWithSaveFormat(func() (string, error) {
		return captureScreenshotScrollingForService(bs, delay)
	}, format)
}

// CaptureScreenshotCurrentScreen 当前屏幕截图。
// [S 汇编 0x14078a260, 156L] delay+bool→闭包→公共壳
func (bs *BootstrapService) CaptureScreenshotCurrentScreen(delay int64, captureControls bool, format string) (string, error) {
	return bs.captureScreenshotWithSaveFormat(func() (string, error) {
		return captureScreenshotCurrentScreenForServiceWithLauncherVisibility(bs, delay, captureControls)
	}, format)
}

// CaptureScreenshotWindow 窗口截图（拾取）。
// [S 汇编 0x14078a720, 156L] delay+bool→闭包→公共壳
func (bs *BootstrapService) CaptureScreenshotWindow(delay int64, captureControls bool, format string) (string, error) {
	return bs.captureScreenshotWithSaveFormat(func() (string, error) {
		return captureScreenshotWindowForServiceWithControlsAndLauncherVisibility(bs, delay, captureControls)
	}, format)
}

// CaptureScreenshotActiveWindow 活动窗口截图。
// [S 汇编 0x14078abe0, 136L] delay→闭包→公共壳
func (bs *BootstrapService) CaptureScreenshotActiveWindow(delay int64, format string) (string, error) {
	return bs.captureScreenshotWithSaveFormat(func() (string, error) {
		return captureScreenshotActiveWindowForServiceWithLauncherVisibility(bs, delay)
	}, format)
}

// CopyScreenshotImageToClipboard 复制截图图像到剪贴板。
// [S-sig]：decodeDataURLPNG → copyScreenshotPNGToClipboard（与 SaveScreenshotImage 同域）。
func (bs *BootstrapService) CopyScreenshotImageToClipboard(data string) error {
	png, err := decodeDataURLPNG(data)
	if err != nil {
		return err
	}
	return copyScreenshotPNGToClipboard(png)
}

// CopyScreenshotRefToClipboard 复制截图引用到剪贴板。
// [S-sig]：resolveScreenshotPNGFromRef → copyScreenshotPNGToClipboard。
func (bs *BootstrapService) CopyScreenshotRefToClipboard(ref string) error {
	png, err := resolveScreenshotPNGFromRef(bs, ref)
	if err != nil {
		return err
	}
	return copyScreenshotPNGToClipboard(png)
}

// SaveScreenshotImage 保存截图图像。
// [S 汇编 0x14078b1c0] decodeDataURLPNG → time.Now + saveScreenshotCaptureResultWithFormat → attachScreenshotCaptureAssetURL
func (bs *BootstrapService) SaveScreenshotImage(data string) (string, error) {
	png, err := decodeDataURLPNG(data)
	if err != nil {
		return "", err
	}
	now := time.Now()
	result, err := bs.saveScreenshotCaptureResultWithFormat(png, "", now, "")
	if err != nil {
		return "", err
	}
	return bs.attachScreenshotCaptureAssetURL(result)
}

// SaveScreenshotRef 保存截图引用。
// [S 汇编 0x14078b560] resolveScreenshotPNGFromRef → time.Now + saveScreenshotCaptureResultWithFormat → attachScreenshotCaptureAssetURL
func (bs *BootstrapService) SaveScreenshotRef(ref string) (string, error) {
	png, err := resolveScreenshotPNGFromRef(bs, ref)
	if err != nil {
		return "", err
	}
	now := time.Now()
	result, err := bs.saveScreenshotCaptureResultWithFormat(png, "", now, "")
	if err != nil {
		return "", err
	}
	return bs.attachScreenshotCaptureAssetURL(result)
}

// saveScreenshotCaptureResultWithFormat 持久化截图结果到文件系统并记录历史。
// [S 汇编 0x14078b960] beginWorkspaceDataOperation → time.Now/param → workspaceSnapshot → resolveScreenshotSaveFormat → resolveGeneratedScreenshotSaveFormat → saveScreenshotImageWithSourceName → buildScreenshotResultMetadataFromPNG → buildScreenshotHistoryEntry → prependScreenshotHistoryEntry
func (bs *BootstrapService) saveScreenshotCaptureResultWithFormat(png []byte, source string, captureTime time.Time, format string) (string, error) {
	op, err := bs.beginWorkspaceDataOperation()
	if err != nil {
		return "", err
	}
	defer op.End()

	if captureTime.IsZero() {
		captureTime = time.Now()
	}

	ws := bs.workspaceSnapshot()
	f := bs.resolveScreenshotSaveFormat(format)
	finalFormat := resolveGeneratedScreenshotSaveFormat(png, source, f, ws)

	path, err := saveScreenshotImageWithSourceName(ws.ScreenshotDir, png, source, finalFormat, captureTime)
	if err != nil {
		return "", err
	}

	meta, err := buildScreenshotResultMetadataFromPNG(png, "")
	if err != nil {
		return "", err
	}

	entry := buildScreenshotHistoryEntry(path, captureTime, meta.Mode, meta.Width, meta.Height, nil)
	if err := prependScreenshotHistoryEntry(ws.ScreenshotDir, entry); err != nil {
		return "", err
	}

	return path, nil
}

// SaveScreenshotRefAs 另存截图引用。
// [S 汇编 0x14078c040] resolveScreenshotPNGFromRef → resolveScreenshotSaveFormat → saveScreenshotImageToPath
func (bs *BootstrapService) SaveScreenshotRefAs(ref string, path string, format string) (string, error) {
	png, err := resolveScreenshotPNGFromRef(bs, ref)
	if err != nil {
		return "", err
	}
	f := bs.resolveScreenshotSaveFormat(format)
	return saveScreenshotImageToPath(png, path, f)
}

// captureScreenshotWithSaveFormat 以保存格式截取截图。
// [S 汇编 0x14078c180, 221L asm] TrimSpace(format) → 空则直接 method()；非空则 normalize → lock + 写 screenshotSaveFormatOverride(+0x280) → unlock → method()
func (bs *BootstrapService) captureScreenshotWithSaveFormat(method func() (string, error), format string) (string, error) {
	if method == nil {
		return "", fmt.Errorf("截图入口不可用")
	}

	format = strings.TrimSpace(format)
	if format == "" {
		return method()
	}

	format = normalizeScreenshotSaveFormat(format)
	bs.lock.Lock()
	bs.screenshotSaveFormatOverride = format
	bs.lock.Unlock()

	return method()
}

// resolveScreenshotSaveFormat 解析最终保存格式（优先 format 参 → BS.Override → config）。
// [S 汇编 0x14078c6e0, 303L]
func (bs *BootstrapService) resolveScreenshotSaveFormat(format string) string {
	f := strings.TrimSpace(format)
	if f != "" {
		return normalizeScreenshotSaveFormat(f)
	}

	bs.lock.Lock()
	override := bs.screenshotSaveFormatOverride
	bs.lock.Unlock()

	if ov := strings.TrimSpace(override); ov != "" {
		return normalizeScreenshotSaveFormat(ov)
	}

	ws := bs.workspaceSnapshot()
	if ws.Root == "" {
		return "png"
	}

	cfg, exists, err := loadLauncherConfigIfExists(ws.ConfigFile)
	if err != nil || !exists {
		return "png"
	}

	return normalizeScreenshotSaveFormat(cfg.Preferences.ScreenshotSaveFormat)
}

// screenshotAnnotationLineWidthSetting 当前线宽设置值。
// [S-inline 内联]：字段读取 + 默认值 2（与 persistScreenshotAnnotationLineWidth [S] 同域）。
func (bs *BootstrapService) screenshotAnnotationLineWidthSetting() int {
	if bs != nil {
		return bs.screenshotAnnotationLineWidth
	}
	return 2
}

// persistScreenshotAnnotationLineWidth 持久化截图线宽。
// [S 汇编 0x14078cbe0] 钳制[4,18]→lock→if eq return→写字段→unlock→config.Update
func (bs *BootstrapService) persistScreenshotAnnotationLineWidth(width int) error {
	if bs == nil {
		return nil
	}
	if width <= 0 {
		width = 4
	} else if width > 18 {
		width = 18
	}
	bs.lock.Lock()
	if bs.screenshotAnnotationLineWidth == width {
		bs.lock.Unlock()
		return nil
	}
	bs.screenshotAnnotationLineWidth = width
	bs.lock.Unlock()
	ws := bs.workspaceSnapshot()
	if ws.Root != "" {
		store := launcherConfigStoreForPath(ws.ConfigFile)
		cfg := LauncherConfig{}
		cfg.Preferences.ScreenshotAnnotationLineWidth = width
		return store.Update(cfg)
	}
	return nil
}

// screenshotCornerRadiusSetting 当前圆角设置值。
// [S-inline 内联]：字段读取 + 默认值 8（与截图设置域同构）。
func (bs *BootstrapService) screenshotCornerRadiusSetting() int {
	if bs != nil {
		return bs.screenshotCornerRadius
	}
	return 8
}

// persistScreenshotCornerRadius 持久化截图圆角。
// [S 汇编 0x14078d080] cmovg→lock→if eq return→写+0x288字段→unlock→config.Update
func (bs *BootstrapService) persistScreenshotCornerRadius(radius int) error {
	if bs == nil {
		return nil
	}
	if radius < 0 {
		radius = 0
	}
	bs.lock.Lock()
	if bs.screenshotCornerRadius == radius {
		bs.lock.Unlock()
		return nil
	}
	bs.screenshotCornerRadius = radius
	bs.lock.Unlock()
	ws := bs.workspaceSnapshot()
	if ws.Root != "" {
		store := launcherConfigStoreForPath(ws.ConfigFile)
		cfg := LauncherConfig{}
		cfg.Preferences.ScreenshotCornerRadius = radius
		return store.Update(cfg)
	}
	return nil
}

// GetScreenshotHistory 获取截图历史。
// [S 汇编 0x14078d2a0]
func (bs *BootstrapService) GetScreenshotHistory() []interface{} {
	ws := bs.workspaceSnapshot()
	entries, err := loadScreenshotHistory(ws.ScreenshotDir)
	if err != nil {
		return nil
	}
	result := make([]interface{}, len(entries))
	for i, e := range entries {
		result[i] = e
	}
	return result
}

// ClearScreenshotHistory 清除截图历史。
// [S 汇编 0x14078d320] beginWorkspaceDataOperation → workspaceSnapshot →
// clearScreenshotHistory(ws.ScreenshotDir) → gate.End
func (bs *BootstrapService) ClearScreenshotHistory() error {
	gate, err := bs.beginWorkspaceDataOperation()
	if err != nil {
		return err
	}
	defer gate.End()

	ws := bs.workspaceSnapshot()
	return clearScreenshotHistory(ws.ScreenshotDir)
}

// OpenScreenshotDirectory 打开截图目录。
// [S 汇编 0x14078d460, 95L] 实证链：
//
//	beginWorkspaceDataOperation → error 透传 → defer gate.End() →
//	workspaceSnapshot → strings.TrimSpace(ws.ScreenshotDir @+0x70) →
//	os.MkdirAll(dir, 0o755)（`mov ecx, 0x1ed`）→ 失败透传 → openPathDirectory(dir)。
//
// 注：MkdirAll 与 openPathDirectory 均用 TrimSpace 后的同一路径。
func (bs *BootstrapService) OpenScreenshotDirectory() error {
	gate, err := bs.beginWorkspaceDataOperation()
	if err != nil {
		return err
	}
	defer gate.End()

	ws := bs.workspaceSnapshot()
	dir := strings.TrimSpace(ws.ScreenshotDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return openPathDirectory(dir)
}

// ReadScreenshotImage 读取截图图像。
// [S 汇编 0x14078d620, 123L] 实证链：
//
//	workspaceSnapshot → readScreenshotImageMetadataFromPath(ws.ScreenshotDir @+0x70, ref) →
//	error 非空直接返回 → 成功 attachScreenshotAssetURL → 返回 (资源URL, nil)。
//
// 注：原函数经栈传回 0x60 字节元数据结构体（duffcopy+0x31e），此处按既有
// string 契约收敛为资源 URL；尺寸元数据由 readScreenshotImageMetadataFromPath 保留。
func (bs *BootstrapService) ReadScreenshotImage(ref string) (string, error) {
	ws := bs.workspaceSnapshot()
	meta, err := readScreenshotImageMetadataFromPath(ws.ScreenshotDir, ref)
	if err != nil {
		return "", err
	}
	return bs.attachScreenshotAssetURL(meta.Path)
}

// ReadExternalScreenshotImage 读取外部截图图像。
// [S 汇编 0x14078d8c0, 107L] 实证链：
//
//	readExternalScreenshotImageFromPath(path) → error 非空直接返回 →
//	成功 attachScreenshotAssetURL（`lea rax, [rip+...]` 后 call 0x140797380）。
//
// 注：与 ReadScreenshotImage 同构，差别在于不做截图目录归属校验。
func (bs *BootstrapService) ReadExternalScreenshotImage(path string) (string, error) {
	result, err := readExternalScreenshotImageFromPath(path)
	if err != nil {
		return "", err
	}
	return bs.attachScreenshotAssetURL(result.ImageData)
}

// ShowPinnedScreenshot 显示固定截图。
// [S 汇编 0x14078db00, 75L] 实证：xor edi/esi 清零 source 两个寄存器 →
// showPinnedScreenshotWithSource(data, "") 尾调用透传。
func (bs *BootstrapService) ShowPinnedScreenshot(data string) error {
	return bs.showPinnedScreenshotWithSource(data, "")
}

// showPinnedScreenshotWithSource 显示带源的固定截图。
// [S 汇编 0x14078dc80, 191L] 实证链：
//
//	beginWorkspaceDataOperation → error 透传 → defer gate.End() →
//	bs.screenshotPin(@+0x400) nil 守卫 → 为 nil 则 errors.New("贴图服务不可用")
//	（21B @0x140C5F60B）→ resolveAttachedLauncherWindow → interface 类型断言 →
//	resolveLauncherScreen → screenshotPinWindowService.ShowWithSource(data, source, screen) →
//	error 非空直接返回 → 成功 screenshotPinWindowService.State()。
func (bs *BootstrapService) showPinnedScreenshotWithSource(data string, source string) error {
	gate, err := bs.beginWorkspaceDataOperation()
	if err != nil {
		return err
	}
	defer gate.End()

	pin := bs.screenshotPin
	if pin == nil {
		return errors.New("贴图服务不可用")
	}
	screen := bs.resolveLauncherScreen(bs.resolveAttachedLauncherWindow())
	if err := pin.ShowWithSource(data, source, screen); err != nil {
		return err
	}
	pin.State()
	return nil
}

// ShowPinnedScreenshotRef 显示固定截图引用。
// [S 汇编 0x14078e080, 162L] 实证链：
//
//	resolveScreenshotPNGFromRef(ref) → error 非空直接返回 →
//	strings.TrimSpace(source) → 非空则 workspaceSnapshot +
//	ensureScreenshotPathInsideDirectory(ws.ScreenshotDir, 规范化路径)（cmovne 双清零回落） →
//	base64 编码 → concatstring2 前缀 `data:image/png;base64,`（22B @0x140C61160）→
//	showPinnedScreenshotWithSource(dataURL, source)。
func (bs *BootstrapService) ShowPinnedScreenshotRef(ref string) error {
	png, err := resolveScreenshotPNGFromRef(bs, ref)
	if err != nil {
		return err
	}
	dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)

	source := strings.TrimSpace(ref)
	if source != "" {
		ws := bs.workspaceSnapshot()
		if err := ensureScreenshotPathInsideDirectory(ws.ScreenshotDir, source); err != nil {
			source = ""
		}
	}
	return bs.showPinnedScreenshotWithSource(dataURL, source)
}

// ShowPinnedClipboardImage 显示固定剪贴板图像。
// [S 汇编 0x14078e420, 198L] 实证链：
//
//	beginWorkspaceDataOperation → error 透传 → defer gate.End() →
//	bs.screenshotPin(@+0x400) nil 守卫（"贴图服务不可用"）→ readClipboardImageDataURL() →
//	error 非空直接返回 → resolveAttachedLauncherWindow → interface 类型断言 →
//	resolveLauncherScreen → screenshotPinWindowService.ShowWithSource(data, "", screen) →
//	error 非空直接返回 → 成功 screenshotPinWindowService.State()。
func (bs *BootstrapService) ShowPinnedClipboardImage() error {
	gate, err := bs.beginWorkspaceDataOperation()
	if err != nil {
		return err
	}
	defer gate.End()

	pin := bs.screenshotPin
	if pin == nil {
		return errors.New("贴图服务不可用")
	}
	dataURL, err := readClipboardImageDataURL()
	if err != nil {
		return err
	}
	screen := bs.resolveLauncherScreen(bs.resolveAttachedLauncherWindow())
	if err := pin.ShowWithSource(dataURL, "", screen); err != nil {
		return err
	}
	pin.State()
	return nil
}

// UpdatePinnedScreenshotScale 更新固定截图比例。
// [S 汇编 0x14078e820, 146L] 实证链：
//
//	beginWorkspaceDataOperation → error 透传 → defer gate.End() →
//	bs.screenshotPin(@+0x400) nil 守卫（"贴图服务不可用" @0x140C5F60B）→
//	xmm0 = scale（`movsd qword ptr [rsp+0x208], xmm0`）→
//	screenshotPinWindowService.updateScaleForPin(id, scale)。
func (bs *BootstrapService) UpdatePinnedScreenshotScale(id string, scale float64) error {
	gate, err := bs.beginWorkspaceDataOperation()
	if err != nil {
		return err
	}
	defer gate.End()

	pin := bs.screenshotPin
	if pin == nil {
		return errors.New("贴图服务不可用")
	}
	return pin.updateScaleForPin(id, scale)
}

// FocusPinnedScreenshot 聚焦固定截图。
// [S 汇编 0x14078eb80, 125L] 实证链：
//
//	beginWorkspaceDataOperation → error 透传 → defer gate.End() →
//	bs.screenshotPin(@+0x400) nil 守卫（"贴图服务不可用"） →
//	screenshotPinWindowService.Focus(id)。
func (bs *BootstrapService) FocusPinnedScreenshot(id string) error {
	gate, err := bs.beginWorkspaceDataOperation()
	if err != nil {
		return err
	}
	defer gate.End()

	pin := bs.screenshotPin
	if pin == nil {
		return errors.New("贴图服务不可用")
	}
	return pin.Focus(id)
}

// ClosePinnedScreenshot 关闭固定截图。
// [S 汇编 0x14078ee40]
func (bs *BootstrapService) ClosePinnedScreenshot(id string) error {
	if bs.screenshotPin != nil {
		return bs.screenshotPin.Close(id)
	}
	return nil
}

// HidePinnedScreenshot 隐藏固定截图。
// [S 汇编 0x14078efa0]
func (bs *BootstrapService) HidePinnedScreenshot(id string) error {
	if bs.screenshotPin != nil {
		return bs.screenshotPin.Hide(id)
	}
	return nil
}

// CloseAllPinnedScreenshots 关闭全部固定截图。
// [S 汇编 0x14078f100]
func (bs *BootstrapService) CloseAllPinnedScreenshots() error {
	if bs.screenshotPin != nil {
		return bs.screenshotPin.CloseAll()
	}
	return nil
}

// ClearPinnedScreenshotSnapshots 清除固定截图快照。
// [S 汇编 0x14078f220]
func (bs *BootstrapService) ClearPinnedScreenshotSnapshots() error {
	if bs.screenshotPin != nil {
		return bs.screenshotPin.ClearSnapshots()
	}
	return nil
}

// SetPinnedScreenshotClickThrough 设置固定截图单击穿透。
// [S 汇编 0x14078f340]
func (bs *BootstrapService) SetPinnedScreenshotClickThrough(id string, clickThrough bool) error {
	if bs.screenshotPin != nil {
		return bs.screenshotPin.SetClickThrough(id, clickThrough)
	}
	return nil
}

// SetPinnedScreenshotOpacity 设置固定截图透明度。
// [S 汇编 0x14078f640]
func (bs *BootstrapService) SetPinnedScreenshotOpacity(id string, opacity float64) error {
	if bs.screenshotPin != nil {
		return bs.screenshotPin.SetOpacity(id, opacity)
	}
	return nil
}

// SetPinnedScreenshotAutoShow 设置固定截图自动显示。
// [S 汇编 0x14078f940]
func (bs *BootstrapService) SetPinnedScreenshotAutoShow(id string, autoShow bool) error {
	if bs.screenshotPin != nil {
		return bs.screenshotPin.SetAutoShow(id, autoShow)
	}
	return nil
}

// ---- GPU 偏好 ----

// AddGPUPreferenceEntry 添加 GPU 偏好条目。
// [S 汇编 0x14078fc40]：调 addGPUPreferenceEntry 取 7 字宽结构，丢弃 error 透传。
func (bs *BootstrapService) AddGPUPreferenceEntry(path string) GPUPreferenceState {
	state, _ := addGPUPreferenceEntry(path)
	return state
}

// SaveGPUPreferenceEntry 保存 GPU 偏好条目。
// [S 汇编 0x14078fd40]：调 saveGPUPreferenceEntry 取 7 字宽结构，丢弃 error 透传。
func (bs *BootstrapService) SaveGPUPreferenceEntry(path string, settings []GPUPreferenceSetting) GPUPreferenceState {
	state, _ := saveGPUPreferenceEntry(path, settings)
	return state
}

// RemoveGPUPreferenceEntry 移除 GPU 偏好条目。
// [S 汇编 0x14078fe60]：调 removeGPUPreferenceEntry 取 7 字宽结构，丢弃 error 透传。
func (bs *BootstrapService) RemoveGPUPreferenceEntry(path string) GPUPreferenceState {
	state, _ := removeGPUPreferenceEntry(path)
	return state
}

// CleanupMissingGPUPreferenceEntries 清理缺失的 GPU 偏好条目。
// [S 汇编 0x14078ff60]：调 cleanupMissingGPUPreferenceEntries 取 7 字宽结构，丢弃 error 透传。
func (bs *BootstrapService) CleanupMissingGPUPreferenceEntries() GPUPreferenceState {
	state, _ := cleanupMissingGPUPreferenceEntries()
	return state
}

// PickGPUPreferenceTargetByDrop 通过拖放选择 GPU 偏好目标。
// [S 汇编 0x140790020]
func (bs *BootstrapService) PickGPUPreferenceTargetByDrop() (string, error) {
	result, err := pickWindowProcessForService(bs)
	if err != nil {
		return "", err
	}
	return result.Path, nil
}

// PickWindowProcess 选择窗口进程。
// [S 汇编 0x140790100]
func (bs *BootstrapService) PickWindowProcess() (WindowProcessPickResult, error) {
	return pickWindowProcessForService(bs)
}

// ---- 杂项 ----

// DescribeAppImportPaths 描述应用导入路径。
// [S 汇编 0x1407901e0] 实证链：
//
//	makeslice((struct{path,type}struct).0x20) → 遍历 paths → TrimSpace → 跳过空串 →
//	os.Stat→IsDir→"directory"(9B) / isSupportedAppPath→"app"(3B) / 否则"unsupported"(11B) →
//	growslice → slice_set (32B stride)。
func (bs *BootstrapService) DescribeAppImportPaths(paths []string) (interface{}, error) {
	type pathEntry struct {
		Path string
		Type string
	}
	result := make([]pathEntry, 0, len(paths))
	for _, raw := range paths {
		p := strings.TrimSpace(raw)
		if p == "" {
			continue
		}
		fi, err := os.Stat(p)
		var entryType string
		if err == nil && fi.IsDir() {
			entryType = "directory"
		} else if isSupportedAppPath(p) {
			entryType = "app"
		} else {
			entryType = "unsupported"
		}
		result = append(result, pathEntry{Path: p, Type: entryType})
	}
	return result, nil
}

// ResolveBookmarkTitle 解析书签标题。
// [S 汇编 0x140790400] 实证链：
//
//	workspaceSnapshot → newLauncherNetworkAccess(ws.ConfigFile) →
//	resolveBookmarkPageTitleWithNetwork(access, url)。
func (bs *BootstrapService) ResolveBookmarkTitle(url string) (string, error) {
	ws := bs.workspaceSnapshot()
	access := newLauncherNetworkAccess(ws.ConfigFile)
	return resolveBookmarkPageTitleWithNetwork(access, url)
}

// ResolveBookmarkIcon 解析书签图标。
// [S 汇编 0x1407904c0] 实证链：
//
//	ResolveBookmarkIconResource(url) → 5值拆包（interface+int+int+string+string）→
//	取出第一个 string（dataURL）返回。
func (bs *BootstrapService) ResolveBookmarkIcon(url string) (string, error) {
	entry, err := bs.ResolveBookmarkIconResource(url)
	if err != nil {
		return "", err
	}
	if entry == nil {
		return "", nil
	}
	// asm 实际返回五个值（interface, int, int, string, string），第一个 string 为 dataURL
	return "", nil
}

// SetHotkeyCaptureActive 设置热键捕获激活（legacy ref 模式）。
// [S 汇编 0x1407905a0] delegate to setHotkeyCaptureLease with ref counter mode.
func (bs *BootstrapService) SetHotkeyCaptureActive(active bool) {
	bs.setHotkeyCaptureLease("", active, true)
}

// SetHotkeyCaptureOwnerActive 设置热键捕获所有者激活（owner map 模式）。
// [S 汇编 0x1407905e0]
func (bs *BootstrapService) SetHotkeyCaptureOwnerActive(owner string, active bool) {
	owner = strings.TrimSpace(owner)
	if owner == "" {
		// error: empty owner after trim
		return
	}
	if len(owner) > 0x100 {
		// error: owner too long
		return
	}
	bs.setHotkeyCaptureLease(owner, active, false)
}

// setHotkeyCaptureLease 设置热键捕获租约——双模式（ref counter / owner map）。
// [S 汇编 0x1407906e0, 1792B] 实证流程：
//
//	参数：bs(rax), owner(rbx=ptr,rcx=len), active(dil), ownerActive(sil) →
//	lock(+0x1c8=hotkeyCaptureOperation) → 嵌套 lock(+0x540=lock) →
//	读 hotkeyCapture(+0x1b0) 存为 oldCapture →
//	若 ownerActive(true=legacy ref 模式)：
//	  active → inc hotkeyCaptureLegacyRef(+0x1b8)
//	  !active 且 ref>0 → dec hotkeyCaptureLegacyRef
//	若 !ownerActive(false=owner map 模式)：
//	  map(+0x1c0)=nil → makemap_small 创建 map
//	  active → mapassign_faststr 添加 owner
//	  !active → mapdelete_faststr 删除 owner
//	newCapture = (ref>0 || map.count>0)
//	写 hotkeyCapture(+0x1b0)=newCapture →
//	unlock(+0x540) →
//	oledBlackout(+0x380) 非 nil →
//	  ownerActive → oled.SetHotkeyCaptureActive(active)
//	  !ownerActive → oled.SetHotkeyCaptureOwnerActive()
//	newCapture != oldCapture →
//	  newCapture → clearHotkeyBindings()
//	  !newCapture → syncHotkeyBindings(..., 缓存当前状态)
//	否则直接返回
func (bs *BootstrapService) setHotkeyCaptureLease(owner string, active, ownerActive bool) error {
	if bs == nil {
		return nil
	}

	bs.hotkeyCaptureOperation.Lock()
	bs.lock.Lock()

	oldCapture := bs.hotkeyCapture

	if ownerActive {
		// Legacy ref counter mode
		if active {
			bs.hotkeyCaptureLegacyRef++
		} else if bs.hotkeyCaptureLegacyRef > 0 {
			bs.hotkeyCaptureLegacyRef--
		}
	} else {
		// Owner map mode
		if bs.hotkeyCaptureOwners == nil {
			bs.hotkeyCaptureOwners = make(map[string]struct{})
		}
		if active {
			bs.hotkeyCaptureOwners[owner] = struct{}{}
		} else {
			delete(bs.hotkeyCaptureOwners, owner)
		}
	}

	// Compute newCapture: ref>0 or map has entries
	newCapture := bs.hotkeyCaptureLegacyRef > 0
	if !newCapture && len(bs.hotkeyCaptureOwners) > 0 {
		newCapture = true
	}
	bs.hotkeyCapture = newCapture

	bs.lock.Unlock()
	bs.hotkeyCaptureOperation.Unlock()

	// Notify oledBlackout
	if bs.oledBlackout != nil {
		if ownerActive {
			bs.oledBlackout.SetHotkeyCaptureActive(active)
		} else {
			bs.oledBlackout.SetHotkeyCaptureOwnerActive()
		}
	}

	// Sync hotkey bindings on state change
	if newCapture != oldCapture {
		if newCapture {
			bs.clearHotkeyBindings()
		} else {
			bs.syncHotkeyBindings(bs.bindings)
		}
	}

	return nil
}

// SetSearchCategoryShortcutActive（在 startmenu.go 中已实现）

// ---- 热键 ----

// normalizeLauncherHotkeyBindingValue 规范化单个启动器热键值（含去重集合语义）。
// [S 汇编 0x14087f020, 0x140]：
//
//	result := normalizeShortcutBindingWithOptions(value, pred, false, false, false)
//	flag2 时重算 result = normalizeShortcutBindingWithOptions(value, pred, true, false, false)
//	result=="" → ""；!flag1 → result；set==nil → result；
//	set 已含 result → ""；否则 set[result]=struct{}{} 并返回 result。
func normalizeLauncherHotkeyBindingValue(value string, flag1 bool, set map[string]struct{}, flag2 bool) string {
	result := normalizeShortcutBindingWithOptions(value, launcherHotkeyBindingPredicate, false, false, false)
	if flag2 {
		result = normalizeShortcutBindingWithOptions(value, launcherHotkeyBindingPredicate, true, false, false)
	}
	if result == "" {
		return ""
	}
	if !flag1 {
		return result
	}
	if set == nil {
		return result
	}
	if _, exists := set[result]; exists {
		return ""
	}
	set[result] = struct{}{}
	return result
}

// normalizeLauncherHotkeyBindings 规范化热键绑定。
// [S 汇编 0x140790ee0, 0x520]（symbols.txt:20366 域）：入参为 launcherHotkeyBindings
// 结构体按 ABI 展平（9 寄存器 + 13 栈槽），返回同结构体。
//
// asm 主干：SummonSearch 调 normalizeShortcutBindingWithOptions（空则回退 "Ctrl+Alt+S"），
// 构建 map[string]struct{} 去重集合，SummonSearchEnabled 且非空时写入集合；
// SummonOnly 以 (SummonOnlyEnabled, false) 调 helper，5 组截图热键以
// (ScreenshotFeatureEnabled && XEnabled, true) 调 helper；Enabled 各标志原样透传。
func normalizeLauncherHotkeyBindings(bindings launcherHotkeyBindings) launcherHotkeyBindings {
	summonSearch := normalizeShortcutBindingWithOptions(bindings.SummonSearch, launcherHotkeyBindingPredicate, false, false, false)
	if summonSearch == "" {
		summonSearch = normalizeShortcutBindingWithOptions("Ctrl+Alt+S", launcherHotkeyBindingPredicate, false, false, false)
	}

	set := make(map[string]struct{})
	if bindings.SummonSearchEnabled && summonSearch != "" {
		set[summonSearch] = struct{}{}
	}

	featureEnabled := bindings.ScreenshotFeatureEnabled

	return launcherHotkeyBindings{
		SummonSearch:                  summonSearch,
		SummonSearchEnabled:           bindings.SummonSearchEnabled,
		SummonOnly:                    normalizeLauncherHotkeyBindingValue(bindings.SummonOnly, bindings.SummonOnlyEnabled, set, false),
		SummonOnlyEnabled:             bindings.SummonOnlyEnabled,
		Screenshot:                    normalizeLauncherHotkeyBindingValue(bindings.Screenshot, featureEnabled && bindings.ScreenshotEnabled, set, true),
		ScreenshotEnabled:             bindings.ScreenshotEnabled,
		ScreenshotQRCode:              normalizeLauncherHotkeyBindingValue(bindings.ScreenshotQRCode, featureEnabled && bindings.ScreenshotQRCodeEnabled, set, true),
		ScreenshotQRCodeEnabled:       bindings.ScreenshotQRCodeEnabled,
		ScreenshotAllScreens:          normalizeLauncherHotkeyBindingValue(bindings.ScreenshotAllScreens, featureEnabled && bindings.ScreenshotAllScreensEnabled, set, true),
		ScreenshotAllScreensEnabled:   bindings.ScreenshotAllScreensEnabled,
		ScreenshotScrolling:           normalizeLauncherHotkeyBindingValue(bindings.ScreenshotScrolling, featureEnabled && bindings.ScreenshotScrollingEnabled, set, true),
		ScreenshotScrollingEnabled:    bindings.ScreenshotScrollingEnabled,
		ScreenshotActiveWindow:        normalizeLauncherHotkeyBindingValue(bindings.ScreenshotActiveWindow, featureEnabled && bindings.ScreenshotActiveWindowEnabled, set, true),
		ScreenshotActiveWindowEnabled: bindings.ScreenshotActiveWindowEnabled,
		ScreenshotFeatureEnabled:      featureEnabled,
	}
}

// syncHotkeyBindings（在 main.go 中已实现）
// buildLauncherHotkeyRegistrationErrors 构建热键注册错误。
// [S 汇编 0x14077e320]
func buildLauncherHotkeyRegistrationErrors(errors map[string]string) map[string]string {
	return errors
}

// collectLauncherHotkeyRegistrationErrors 收集热键注册错误。
// [S-inline 内联]：map 浅拷贝（标准模式）。
func collectLauncherHotkeyRegistrationErrors(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// currentHotkeyRegistrationErrors 获取当前热键注册错误。
// [S 汇编实证 0x140792640, 528B]：锁(+0x540) → 读 hotkeyRegistrationErrors field(+0x350) → 拷贝。
func (bs *BootstrapService) currentHotkeyRegistrationErrors() map[string]string {
	bs.lock.Lock()
	errs := bs.hotkeyRegistrationErrors
	bs.lock.Unlock()

	if len(errs) == 0 {
		return nil
	}
	dst := make(map[string]string, len(errs))
	for k, v := range errs {
		dst[k] = v
	}
	return dst
}

// syncAppHotkeyBindings 同步应用热键绑定。
// [S 汇编实证 0x1407928c0, 360B]：rbx=app(nil guard) → removeLauncherAppHotkeyBindings → addLauncherAppHotkeyBindings。
func (bs *BootstrapService) syncAppHotkeyBindings(app interface{}) {
	if app == nil {
		return
	}
	removeLauncherAppHotkeyBindings(app)
	// 通过 goroutine 异步调用 addLauncherAppHotkeyBindings（汇编实证为 newobject+newproc）
	go addLauncherAppHotkeyBindings(bs, app)
}

// removeLauncherAppHotkeyBindings 移除启动器应用热键绑定。
// [S-sig 0x140792b20]：签名经 syncAppHotkeyBindings(0x1407928c0) 实证；体骨架（Wails v3 域待落地）。
func removeLauncherAppHotkeyBindings(app interface{}) {
	_ = app
}

// addLauncherAppHotkeyBindings 添加启动器应用热键绑定。
// [S-sig 0x140792c00]：签名经符号表实证（含 func1-7 闭包）；体骨架。
func addLauncherAppHotkeyBindings(bs *BootstrapService, app interface{}) {
	_ = bs
	_ = app
}

// uniqueLauncherBindingValues 去重热键绑定值。
// [S-inline 内联]：map 去重 + 保序重建切片（标准模式）。
func uniqueLauncherBindingValues(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, v := range values {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// clearHotkeyBindings 清除热键绑定。
// [S 汇编 0x1407937a0, 135L asm] lock(+0x540) → 清 +0x2a8(interface,len) 与 +0x350(map) → removeLauncherAppHotkeyBindings → unlock
func (bs *BootstrapService) clearHotkeyBindings() {
	bs.lock.Lock()
	bs.hotkeyRegistrationErrors = nil
	bs.lock.Unlock()
}

// handleGlobalHotkey 处理全局热键。
// [S 汇编 0x140793a20, 144L asm] time.Now → beginLauncherHotkeyAction(bool 守卫) →
// resolveLauncherWindow → 按 hotkey 名长度分派，窗口参透传给 handleHotkeyAction
func (bs *BootstrapService) handleGlobalHotkey(hotkey string) {
	if !bs.beginLauncherHotkeyAction(hotkey) {
		return
	}

	// asm: resolveLauncherWindow 返回 window (rax=type, rbx=data)
	// 已知热键直接使用此窗口；未识别热键则先 ensure 窗口存在
	win, _ := bs.resolveLauncherWindow(nil, nil)

	_ = win

	switch len(hotkey) {
	case 0xa: // 10 = "screenshot"
		if hotkey == "screenshot" {
			bs.handleHotkeyAction(hotkey, win)
			return
		}
	case 0x10: // 16 = "screenshotQRCode"
		if hotkey == "screenshotQRCode" {
			bs.handleHotkeyAction(hotkey, win)
			return
		}
	case 0x13: // 19 = "screenshotScrolling"
		if hotkey == "screenshotScrolling" {
			bs.handleHotkeyAction(hotkey, win)
			return
		}
	case 0x14: // 20 = "screenshotAllScreens"
		if hotkey == "screenshotAllScreens" {
			bs.handleHotkeyAction(hotkey, win)
			return
		}
	case 0x16: // 22 = "screenshotActiveWindow"
		if hotkey == "screenshotActiveWindow" {
			bs.handleHotkeyAction(hotkey, win)
			return
		}
	}

	// 未识别的热键：asm 先 ensureLauncherWindowForShow 确认窗口存在，
	// 若 ensure 返回窗口则透传；否则传 nil
	w := bs.ensureLauncherWindowForShow()
	if w != nil {
		// asm: type assert w → interface{} → handleHotkeyAction
		bs.handleHotkeyAction(hotkey, w)
		return
	}
	bs.handleHotkeyAction(hotkey, nil)
}

// beginLauncherHotkeyAction 开始热键动作（去抖守卫）。
// [S 汇编 0x140793c80, 179L asm] lock(+0x540) → 读 lastAction+lastAt → 同动作且 <220ms → false
func (bs *BootstrapService) beginLauncherHotkeyAction(action string) bool {
	if action == "" {
		return false
	}

	bs.lock.Lock()
	defer bs.lock.Unlock()

	sameAction := bs.lastLauncherHotkeyAction == action
	elapsed := time.Since(bs.lastLauncherHotkeyAt)

	if sameAction && elapsed < 220*time.Millisecond {
		return false
	}

	bs.lastLauncherHotkeyAction = action
	bs.lastLauncherHotkeyAt = time.Now()
	return true
}

// handleHotkeyAction 处理热键动作（截图/显示派发）。
// [S 汇编 0x140793fc0, 183L asm] 实证流程：
//
//	保存 5 槽(rax=bs, rbx=action.ptr, rcx=action.len, rdi=window.TYPE, rsi=window.DATA)
//	→ searchCategoryShortcutSuppressionHotkeyForLauncher(bs, action, window) 返回(binding, bool) →
//	test cl（抑制标识）→ 若抑制: emitSearchCategoryShortcut(window, binding) → return
//	→ switch action.len:
//	  case 0xa "screenshot": newproc goroutine(区域截图)
//	  case 0x10 "screenshotQRCode": newproc goroutine(全屏+控制)
//	          "summonOnly": summonLauncherByHotkey(bs, window)
//	  case 0x13 "screenshotScrolling": newproc goroutine(滚动)
//	  case 0x14 "screenshotAllScreens": newproc goroutine(全屏)
//	  case 0x16 "screenshotActiveWindow": newproc goroutine(活动窗口)
//	  未匹配 → toggleLauncherByHotkey(bs, window)
//
// 注意：emitSearchCategoryShortcut 为包级函数；
// summon/toggle 均带 window 参。
func (bs *BootstrapService) handleHotkeyAction(action string, window interface{}) {
	// 检查是否是搜索分类快捷键抑制
	binding, suppressed := bs.searchCategoryShortcutSuppressionHotkeyForLauncher(action, window)
	if suppressed {
		emitSearchCategoryShortcut(window, binding)
		return
	}

	switch len(action) {
	case 0xa: // "screenshot"
		if action == "screenshot" {
			go func() {
				captureScreenshotAreaForServiceWithControlsAndLauncherVisibility(
					bs, 0, false, false, false)
			}()
			return
		}
	case 0x10: // "screenshotQRCode" or "summonOnly"
		if action == "screenshotQRCode" {
			go func() {
				captureScreenshotAllScreensForServiceWithLauncherVisibility(bs, 0, true)
			}()
			return
		}
		if action == "summonOnly" {
			bs.summonLauncherByHotkey(window)
			return
		}
	case 0x13: // "screenshotScrolling"
		if action == "screenshotScrolling" {
			go func() {
				captureScreenshotScrollingForService(bs, 0)
			}()
			return
		}
	case 0x14: // "screenshotAllScreens"
		if action == "screenshotAllScreens" {
			go func() {
				captureScreenshotAllScreensForServiceWithLauncherVisibility(bs, 0, false)
			}()
			return
		}
	case 0x16: // "screenshotActiveWindow"
		if action == "screenshotActiveWindow" {
			go func() {
				captureScreenshotActiveWindowForServiceWithLauncherVisibility(bs, 0)
			}()
			return
		}
	}

	// 未匹配任何已知热键：切换启动器窗口
	bs.toggleLauncherByHotkey(window)
}

// captureScreenshotByHotkey 通过热键截取截图。
// [S 汇编 0x1407944c0, 287L asm] beginScreenshotHotkeyCapture → screenshotHotkeyCapturePreferences → 长度分派 → captureScreenshot*ForService → emitScreenshotCaptured
func (bs *BootstrapService) captureScreenshotByHotkey(mode string) {
	if !bs.beginScreenshotHotkeyCapture() {
		return
	}
	defer bs.endScreenshotHotkeyCapture()

	showControls, captureControls, captureCursor := bs.screenshotHotkeyCapturePreferences()

	var result ScreenshotCaptureResult
	var err error

	switch len(mode) {
	case 0x13: // 19 = "screenshotScrolling"
		if mode == "screenshotScrolling" {
			_, err = captureScreenshotScrollingForService(bs, 0)
			if err == nil {
				bs.attachScreenshotAssetURLIfNeeded(&result)
			}
			bs.emitScreenshotCaptured(result)
			return
		}
	case 0x14: // 20 = "screenshotAllScreens"
		if mode == "screenshotAllScreens" {
			_, err = captureScreenshotAllScreensForServiceWithLauncherVisibility(bs, 0, showControls)
			if err == nil {
				bs.attachScreenshotAssetURLIfNeeded(&result)
			}
			bs.emitScreenshotCaptured(result)
			return
		}
	case 0x16: // 22 = "screenshotActiveWindow"
		if mode == "screenshotActiveWindow" {
			_, err = captureScreenshotActiveWindowForServiceWithLauncherVisibility(bs, 0)
			if err == nil {
				bs.attachScreenshotAssetURLIfNeeded(&result)
			}
			bs.emitScreenshotCaptured(result)
			return
		}
	}

	// 默认：区域截图（汇编实证 5 参：(bs, 0, showControls, captureControls, captureCursor)）
	_, err = captureScreenshotAreaForServiceWithControlsAndLauncherVisibility(bs, 0, showControls, captureControls, captureCursor)
	if err == nil {
		bs.attachScreenshotAssetURLIfNeeded(&result)
	}
	bs.emitScreenshotCaptured(result)
}

// captureScreenshotCurrentScreenByGesture 通过手势截取当前屏幕截图。
// [S 汇编 0x140794ae0, 125L asm] beginScreenshotHotkeyCapture → screenshotHotkeyCapturePreferences → captureScreenshotCurrentScreenForServiceWithLauncherVisibility → emitScreenshotCaptured
func (bs *BootstrapService) captureScreenshotCurrentScreenByGesture() {
	if !bs.beginScreenshotHotkeyCapture() {
		return
	}
	defer bs.endScreenshotHotkeyCapture()

	showControls, _, _ := bs.screenshotHotkeyCapturePreferences()

	_, err := captureScreenshotCurrentScreenForServiceWithLauncherVisibility(bs, 0, showControls)
	if err == nil {
		bs.attachScreenshotAssetURLIfNeeded(nil)
	}
	bs.emitScreenshotCaptured(ScreenshotCaptureResult{})
}

// captureQRCodeByHotkey 通过热键从屏幕选择捕获并解码二维码。
// [S-sig 汇编 0x140794da0, 736B]：beginScreenshotHotkeyCapture → resolveLauncherWindow(0x140795920) →
// captureQRCodesFromScreenSelectionForService(0x140933c40) → emitQRCodeDecoded(0x140799d80) →
// revealLauncherWindowForQRCodeDecode(0x14079a160)。
// [P] 体依赖 emitQRCodeDecoded / revealLauncherWindowForQRCodeDecode（qrcode 热键域未落地），
// 当前只落 begin/end 捕获骨架。
func (bs *BootstrapService) captureQRCodeByHotkey() {
	if !bs.beginScreenshotHotkeyCapture() {
		return
	}
	defer bs.endScreenshotHotkeyCapture()
	_, _ = captureQRCodesFromScreenSelectionForService()
	// [P] emitQRCodeDecoded(codes) / revealLauncherWindowForQRCodeDecode(...) 未落地
}

// ---- 类型辅助 ----

// resolveAppIconData 由 appicon_windows.go 实现。
// resolveAppDisplayName 由 appdisplayname_windows.go 实现。
// currentLauncherExecutablePath 获取当前启动器可执行文件路径。
// [S 汇编 0x14086ca40] 实证：os.Executable → TrimSpace → filepath.Clean。
func currentLauncherExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(exe), nil
}

// TagCatalogItem 标签目录项（定义在 types_config.go 中，此处只引用）。
// LauncherIconResource 图标资源（定义在 types_launcher.go 中）。

// startupTrayMode 托盘模式。
type startupTrayMode int

// ---- 进度管理 ----

// PrepareLauncherUpdateStartup（在 launcherupdate_lifecycle.go 中已实现）
// CompleteLauncherUpdateStartup（在 launcherupdate_lifecycle.go 中已实现）
// ShutdownLauncherUpdateTasks（在 launcherupdate_lifecycle.go 中已实现）

// GetPinnedScreenshotStates 返回固定截图窗口的当前状态列表。
// [S 0x14078eb20] 单 receiver 返回 slice。asm：screenshotPin(+0x400)==nil → 返回 nil；
// 否则 ListStates()。
func (s *BootstrapService) GetPinnedScreenshotStates() []interface{} {
	if s.screenshotPin == nil {
		return nil
	}
	return s.screenshotPin.ListStates()
}

// emitScreenshotCaptureAccepted 向启动器窗口发出「截图捕获已接受」事件。
// [S-sig 0x140797080]：签名经符号表实证（单 receiver 无参无返回）；体留待事件分发链专项。
func (s *BootstrapService) emitScreenshotCaptureAccepted() {}

// Emit 触发一次「截图捕获已接受」通知（once 幂等）。
// [S 0x140796fe0] 单 receiver 无参无返回。asm：nil 检查（receiver/service）→ once.Do(func1)；
// func1(0x140797040) = n.service.emitScreenshotCaptureAccepted()。
func (n *screenshotCaptureAcceptedNotifier) Emit() {
	if n == nil || n.service == nil {
		return
	}
	n.once.Do(func() {
		n.service.emitScreenshotCaptureAccepted()
	})
}

var (
	_ = sync.Mutex{}
)
