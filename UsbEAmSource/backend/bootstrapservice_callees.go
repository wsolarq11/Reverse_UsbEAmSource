// AUTO-RECONSTRUCTED BOOTSTRAP CALLEES — 包级基础函数桩
// 研究用途
//
// 档位：混合。
//
//	[S]     反汇编实证体（批次 29 起）：buildLauncherConfigIconResource (0x1408a28e0)、
//	        cachedLauncherConfigIconAssetURL (0x1408a2f00)、cacheLauncherConfigIconAssetURL (0x1408a33c0)、
//	        buildLauncherIconResource (0x1408a4f20)
//	[P]     其余桩代码，待反汇编逐条实证还原
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

// ---- GPU 偏好 ----
// 本组属 gpu.go / gpu_pick_windows.go 蓝图（§10 未落地文件）。asm 已现场抽取。
// getGPUPreferenceState / addGPUPreferenceEntry / saveGPUPreferenceEntry /
// removeGPUPreferenceEntry / cleanupMissingGPUPreferenceEntries 及助手链
// 已迁至 gpu_windows.go（批次 156 落地）。
// [P] 阻断原因：GPU 偏好域专项，依赖未落地助手链
// （resolveWindowProcessPath / resolveProcessPathByPID 等），需独立批次还原。

// pickWindowProcessForService 从屏幕点选择窗口进程（拖放拾取主流程）。
// [S-sig 汇编 0x14085be20] 依赖未落地 beginLauncherScreenshotCapture /
// endLauncherScreenshotCapture / resolveLauncherWindow / showLauncherWindow 与
// [P] pickWindowProcessFromScreenshotSelection，暂以正确签名占位：
// launcherPID → ensureGPUPickDebugLogger → resolveLauncherWindow(非空时) →
// beginLauncherScreenshotCapture + defer 恢复 → 180ms 睡眠 →
// pickWindowProcessFromScreenshotSelection → 结果分支（失败/取消/成功）。
func pickWindowProcessForService(bs *BootstrapService) (WindowProcessPickResult, error) {
	return WindowProcessPickResult{}, nil
}

// ---- QR 码 ----
// 本组属 qrcode.go 蓝图（§10 未落地文件）。asm 已现场抽取。

// copyQRCodeImageToClipboard / generateQRCodeDataURL 已迁至 screenshot_qrcode_windows.go（批次 101 落地）。

// ---- 截图 ----

// captureQRCodesFromScreenSelectionForService 从屏幕选择捕获并解码二维码。
// [S-sig 汇编 0x140933c40] qrcode + screenshot 跨域专项，助手链未落地。
func captureQRCodesFromScreenSelectionForService() (QRCodeDecodeResult, error) {
	return QRCodeDecodeResult{}, nil
}

// ---- 语言 ----

// BootstrapService.languageDirectories 获取语言目录（方法版）。
// [S 汇编 0x1407a2720, 128B]：workspaceSnapshot() 取布局 → duffcopy 传值 →
// languageDirectoriesForWorkspace(ws) 返回 []string。
// 包级 languageDirectories 为幽灵符号（symbols.txt 无 main.languageDirectories），已删除。
func (bs *BootstrapService) languageDirectories() []string {
	return languageDirectoriesForWorkspace(bs.workspaceSnapshot())
}

// loadLanguageMessages 加载语言消息 map。
// [S-sig 汇编 0x1407a3300, 928B] 语言加载域专项（log.go 蓝图未落地），体待还原。
func loadLanguageMessages(dirs []string) map[string]interface{} {
	return make(map[string]interface{})
}

// discoverLanguages 发现语言清单。
// [S-sig 汇编 0x1407a2c60, 1568B] 语言加载域专项，体待还原。
func discoverLanguages(dirs []string) []LanguageManifest { return nil }

// discoverPlugins 在 RWMutex 保护下发现插件清单。
// [S 汇编实证 0x1407a5320, 288B]（symbols.txt:19213）：全局 RWMutex RLock
// （asm 0x1407a535f lock xadd）→ 原样透传 rax/rbx/rcx 调 discoverPluginsUnlocked
// （0x1407a54a0, 2432B）→ 三字宽结果（[]PluginManifest）→ RUnlock。
// 寄存器仅承载一个 slice，故签名是单参 []string——旧桩 (WorkspaceLayout, map) 不成立。
func discoverPlugins(dirs []string) []PluginManifest { return nil }

// discoverPluginsUnlocked 无锁发现插件清单。
// [S 汇编 0x1407a54a0, 2432B]：遍历 dirs → ReadDir 失败跳过 → 逐项：
// 非目录或 symlink 跳过 → TrimSpace 名空或以 "." 开头跳过 →
// resolvePluginPathSecurely(dir,name,true) → resolvePluginPathSecurely(resolved,"plugin.json",false) →
// readPluginManifestBounded → TrimSpace(Entry) 空则 "index.html" →
// resolvePluginPathSecurely(resolved,Entry,false) 失败跳过 → SourceDir=resolved/Installed=true →
// Icon=normalizePluginIconFile 非空则校验可读(1MB)否则清空 → IconURL=resolvePluginLocalIconURL →
// I18N=loadPluginLocalizations → key=ToLower(TrimSpace(ID)) 去重 → 按 ID 排序返回。
func discoverPluginsUnlocked(dirs []string) []PluginManifest {
	seen := make(map[string]PluginManifest)
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
				continue
			}
			name := strings.TrimSpace(entry.Name())
			if strings.HasPrefix(name, ".") {
				continue
			}
			resolved, err := resolvePluginPathSecurely(dir, name, true)
			if err != nil {
				continue
			}
			manifestPath, err := resolvePluginPathSecurely(resolved, "plugin.json", false)
			if err != nil {
				continue
			}
			manifest, err := readPluginManifestBounded(manifestPath)
			if err != nil {
				continue
			}
			if e := strings.TrimSpace(manifest.Entry); e == "" {
				manifest.Entry = "index.html"
			}
			if _, err := resolvePluginPathSecurely(resolved, manifest.Entry, false); err != nil {
				continue
			}
			manifest.SourceDir = resolved
			manifest.Installed = true
			manifest.Icon = normalizePluginIconFile(manifest.Icon)
			if manifest.Icon != "" {
				iconPath, err := resolvePluginPathSecurely(resolved, manifest.Icon, false)
				if err != nil {
					manifest.Icon = ""
				} else if _, err := readPluginFileBounded(iconPath, 1<<20); err != nil {
					manifest.Icon = ""
				}
			}
			manifest.IconURL = resolvePluginLocalIconURL(manifest)
			manifest.I18N = loadPluginLocalizations(manifest)
			key := strings.ToLower(strings.TrimSpace(manifest.ID))
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = manifest
		}
	}
	result := make([]PluginManifest, 0, len(seen))
	for _, m := range seen {
		result = append(result, m)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

// BootstrapService.pluginDirectories 获取插件目录。
// [S 汇编 0x1407a2860, 288B]：workspaceSnapshot() 取布局 → TrimSpace(PluginDir@+0x40) →
// 空则 nil，否则 []string{dir}（与 GetSnapshot 0x1407742f8 内联逻辑一致）。
func (bs *BootstrapService) pluginDirectories() []string {
	dir := strings.TrimSpace(bs.workspaceSnapshot().PluginDir)
	if dir == "" {
		return nil
	}
	return []string{dir}
}

// resolvePluginLanguageMessages 解析插件的语言消息 map（en-US 兜底 + 目标语言覆盖）。
// [S 汇编 0x1407a4f80, 448B]：TrimSpace(lang) 空则 "en-US" → 新建 map →
// 先 merge I18N["en-US"].Messages，再 lang!="en-US" 时 merge I18N[lang].Messages。
func resolvePluginLanguageMessages(manifest PluginManifest, lang string) map[string]interface{} {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		lang = "en-US"
	}
	messages := make(map[string]interface{})
	if loc, ok := manifest.I18N["en-US"]; ok {
		mergeLanguageMessages(messages, loc.Messages)
	}
	if lang != "en-US" {
		if loc, ok := manifest.I18N[lang]; ok {
			mergeLanguageMessages(messages, loc.Messages)
		}
	}
	return messages
}

// languageDirectoriesForWorkspace 从工作区布局取语言目录。
// [S 汇编 0x1407a27a0, 192B]：TrimSpace(AppLanguageDir@+0x30) → 空则 TrimSpace(LanguageDir@+0x20)
// → 仍空 return nil → 否则 []string{dir}。
func languageDirectoriesForWorkspace(ws WorkspaceLayout) []string {
	dir := strings.TrimSpace(ws.AppLanguageDir)
	if dir == "" {
		dir = strings.TrimSpace(ws.LanguageDir)
	}
	if dir == "" {
		return nil
	}
	return []string{dir}
}

// ---- 标签目录辅助 ----

// mergeLanguageMessages 深合并语言消息 map。
// [S 汇编实证 0x1407a5140, 480B]：迭代 src map → 逐 key 合并到 dst（嵌套递归 string map）。
func mergeLanguageMessages(dst map[string]interface{}, src map[string]interface{}) {
	for k, v := range src {
		if vm, ok := v.(map[string]interface{}); ok {
			if dm, ok := dst[k].(map[string]interface{}); ok {
				mergeLanguageMessages(dm, vm)
			} else {
				dst[k] = vm
			}
		} else {
			dst[k] = v
		}
	}
}

// nestedLanguageMessageString 按点号分隔键路径逐层取语言消息字符串。
// [S 汇编实证 0x1407a6500, 352B]：TrimSpace(key) → strings.Split(key, ".") → 逐层
// TrimSpace(part) 空则 "" → mapaccess → 末层 string 断言 TrimSpace 返回。
func nestedLanguageMessageString(langMsgs map[string]interface{}, key string) string {
	parts := strings.Split(strings.TrimSpace(key), ".")
	current := langMsgs
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return ""
		}
		v, ok := current[part]
		if !ok {
			return ""
		}
		if i == len(parts)-1 {
			if s, ok := v.(string); ok {
				return strings.TrimSpace(s)
			}
			return ""
		}
		if m, ok := v.(map[string]interface{}); ok {
			current = m
		} else {
			return ""
		}
	}
	return ""
}

// defaultTagCatalogFromNames 从标签名切片生成默认 TagCatalogItem 列表。
// [S 汇编实证 0x14087be00, 544B]：遍历 names → TrimSpace → 跳空 → 按索引取图标 → 构造 TagCatalogItem。
func defaultTagCatalogFromNames(names []string) []TagCatalogItem {
	if len(names) == 0 {
		return nil
	}
	result := make([]TagCatalogItem, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		icon := defaultTagCatalogIconForIndex(len(result), name)
		result = append(result, TagCatalogItem{
			Name: name,
			Icon: icon,
		})
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// defaultTagCatalogIcon 标签名到预设图标的 switch 映射。
// [S 汇编实证 0x14087c020, 640B]：TrimSpace → ToLower → 比较各已知名（UTF-8 魔数）→ 返 icon 串。
func defaultTagCatalogIcon(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "gamepad":
		return "gamepad"
	case "briefcase":
		return "briefcase"
	case "games":
		return "games"
	case "folder":
		return "folder"
	case "code":
		return "code"
	case "dev":
		return "dev"
	case "multimedia":
		return "multimedia"
	default:
		return "bookmark"
	}
}

// defaultTagCatalogIconForIndex 按索引取标签图标（含全局表路由）。
// [S 汇编实证 0x14087c2a0, 160B]：若 index 在全局图标表范围内，取表项；否则 fallback 到 defaultTagCatalogIcon。
func defaultTagCatalogIconForIndex(index int, name string) string {
	defaultIcons := []string{"gamepad", "briefcase", "wrench", "users", "code", "file"}
	if index >= 0 && index < len(defaultIcons) {
		icon := strings.TrimSpace(defaultIcons[index])
		if icon != "" {
			return icon
		}
	}
	return defaultTagCatalogIcon(name)
}

// ---- 背景 ----

// getLauncherBackgroundMetrics 返回启动器背景指标。
// [S 汇编 0x140874520]：windowManagementVirtualScreenBounds()（rax/rbx/rcx/rdi=Min.X/Min.Y/Max.X/Max.Y）
// → width=Max.X-Min.X、height=Max.Y-Min.Y；宽或高 <=0 返回零值结构，否则
// Supported=true、VirtualX/Y=Min.X/Min.Y、Width/Height=宽高。返回结构经栈 40B 展开。
func getLauncherBackgroundMetrics() LauncherBackgroundMetrics {
	bounds := windowManagementVirtualScreenBounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y
	if width <= 0 || height <= 0 {
		return LauncherBackgroundMetrics{}
	}
	return LauncherBackgroundMetrics{
		Supported: true,
		VirtualX:  bounds.Min.X,
		VirtualY:  bounds.Min.Y,
		Width:     width,
		Height:    height,
	}
}

// ---- 书签 ----

// buildBookmarkState 构建书签状态。
// [S-sig 汇编 0x140764e40, 2848B] bookmarks.go 域专项（§10 未落地），体待还原。
func buildBookmarkState(ws WorkspaceLayout) interface{} { return nil }

// ---- 链接 ----

// loadLinkPreferences 加载链接偏好。
// [S-sig 汇编 0x140769c60, 544B] 方法形式；真返回 (Preferences, error)（0x88 qword=1088B 返回体 + 16B 错误）。
// 体：workspaceSnapshot → launcherConfigOptions("en-US" 5B@0x140c35be6) →
// loadLauncherConfigOrDefaultIfMissing(ws.ConfigFile, &opts, 1) → 按 bool 分支
// （!ok → normalizePreferencesWithOptions(cfg.Preferences, &opts)；ok → 零值 Preferences）。
// [P] 体依赖 loadLauncherConfigOrDefaultIfMissing(0x140875e00,736B) /
// normalizePreferencesWithOptions(0x14087cb00,8160B) / linkbrowser*.go，待还原。
func (bs *BootstrapService) loadLinkPreferences() (Preferences, error) { return Preferences{}, nil }

// openLinkWithPreferences 按偏好打开链接。
// [S-sig 汇编 0x140768fc0, 1600B] link 域专项（linkbrowser*.go 未落地），体待还原。
func openLinkWithPreferences(prefs interface{}, entries interface{}) error { return nil }

// isSupportedAppPath 判断路径是否为受支持的应用扩展名。
// [S 汇编 0x14088ce00, 224B]：TrimSpace → 自末尾反扫取最后一个 '.' 之后扩展名
// （遇 '\\' 或 '/' 先于 '.' 则无扩展名）→ ToLower → 匹配 .exe/.lnk/.url/.appref-ms。
// 魔数（little-endian）0x6578652e=.exe、0x6b6e6c2e=.lnk、0x6c72752e=.url、0x2d6665727070612e+0x736d=.appref-ms。
func isSupportedAppPath(path string) bool {
	trimmed := strings.TrimSpace(path)
	ext := ""
	for i := len(trimmed) - 1; i >= 0; i-- {
		c := trimmed[i]
		if c == '\\' || c == '/' {
			break
		}
		if c == '.' {
			ext = trimmed[i:]
			break
		}
	}
	switch strings.ToLower(ext) {
	case ".exe", ".lnk", ".url", ".appref-ms":
		return true
	}
	return false
}

// ---- 截图辅助 ----

// resolveScreenshotPNGFromRef 解析截图引用为 PNG 字节。
// [S 汇编 0x140799a00]（批次 24 已升 [S]，本批恢复丢失标记）：资产 ReadBytes → DataURL 双路回退。
func resolveScreenshotPNGFromRef(bs *BootstrapService, ref string) ([]byte, error) {
	if bs == nil {
		return nil, errors.New("截图服务不可用")
	}

	r := strings.TrimSpace(ref)
	if r == "" {
		return nil, errors.New("截图引用为空")
	}

	// 资产模式：从 ScreenshotAssetService 读取
	svc := bs.screenshotAssetService()
	if svc != nil {
		data, contentType, err := svc.ReadBytes(r, "screenshot")
		if err == nil && len(data) > 0 {
			if strings.EqualFold(strings.TrimSpace(contentType), "image/png") {
				return data, nil
			}
			return convertScreenshotImageBytesToPNG(data)
		}
	}

	// DataURL 模式
	if strings.HasPrefix(r, "data:image/") {
		return decodeDataURLPNG(r)
	}

	return nil, errors.New("无法解析截图引用")
}

// ---- Desktop Widget ----

// ---- 更新 ----

// sendLauncherUpdateHealthFromEnvironment 从环境发送更新健康状态。
// [S-sig 汇编 0x1408c24c0, 1056B] launcherupdate 域专项，体待还原。
func sendLauncherUpdateHealthFromEnvironment() (string, error) { return "", nil }

// ---- 热键 ----

// setHotkeyCaptureLease 为幽灵符号（symbols.txt 仅 main.BootstrapService.setHotkeyCaptureLease@0x1407906e0，
// 方法已实现于 bootstrapservice.go:2479）。包级桩已删除。

// ResolveBookmarkIconResource 解析书签图标资源。
// [S-sig 汇编 0x1408a1920] bookmarks 域专项（bookmarks*.go 未落地），体待还原。
func (bs *BootstrapService) ResolveBookmarkIconResource(url string) (interface{}, error) {
	return nil, nil
}

// startDetachedCommand 启动分离的命令（不等待退出）。
// [S 汇编实证 0x1408a8480, 256B]：
//
//	入口 3 寄存器 = []string(ptr/len/cap)，非 (string, bool)。morestack 仅存 rax/rbx/rcx。
//	len(args)==0 → 错误 "命令不能为空"(18B)
//	→ exec.Command(args[0], args[1:]...) → Cmd.Start()
//	→ Start 失败 → fmt.Errorf("启动失败: %w", err)（模板 16B）
//	→ 成功返回 nil
//
// 签名实证：asm `mov rdx,[rax]; mov r8,[rax+8]` 读的是 slice 首元素字符串头。
func startDetachedCommand(args []string) error {
	if len(args) == 0 {
		return errors.New("命令不能为空")
	}
	cmd := exec.Command(args[0], args[1:]...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动失败: %w", err)
	}
	return nil
}

// PersistSnapshotsForShutdown 持久化关闭前快照。
// [S-sig 汇编 0x140981580, 14784B(0x39c0)]：nil guard → lock(+0x0) → restoreSuppressed(+0x52) 为假则解锁直返 →
// 为真则 shutting(+0x50)=true、shutdownPersisted(+0x53)=true → 解锁 → snapshotMirror → saveSnapshots。
// 旧注释误标 224B——实测体长 0x39c0=14784B（snapshotMirror 0x140984f40 为下一符号）。
// [P] 体：snapshotMirror(0x140984f40,20896B)/saveSnapshots(0x14098a0e0)（screenshot_pin 域未落地）暂缺，
// 本批只落锁/标志语义。
func (s *screenshotPinWindowService) PersistSnapshotsForShutdown() {
	if s == nil {
		return
	}
	s.lock.Lock()
	if !s.restoreSuppressed {
		s.lock.Unlock()
		return
	}
	s.shutting = true
	s.shutdownPersisted = true
	s.lock.Unlock()
	// TODO [P]: snapshotMirror()/saveSnapshots(...) 待 screenshot_pin 域还原后接入。
}

// defaultLauncherConfigWithOptions 生成带选项的缺省配置。
// [S-sig 汇编 0x140874980, 4544B] launcherconfiginput.go 域专项（§10 未落地，巨型帧），体待还原。
func defaultLauncherConfigWithOptions(lang string) LauncherConfig { return LauncherConfig{} }

// normalizeLauncherConfigWithOptions 规范化配置（填充缺省值）。
// [S-sig 汇编 0x140879380, 1984B] launcherconfiginput.go 域专项（§10 未落地），体待还原。
func normalizeLauncherConfigWithOptions(cfg *LauncherConfig, lang string) {}

// launcherConfigsEqual 比较两个配置是否相等。
// [S 汇编 0x14088bcc0, 192B]：json.Marshal(a) 出错返 false → json.Marshal(b) 出错返 false →
// 长度相等（cmp rbx,rcx）→ memequal 字节比较。等价于 bytes.Equal(jsonA, jsonB)。
func launcherConfigsEqual(a, b LauncherConfig) bool {
	aj, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bj, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return bytes.Equal(aj, bj)
}

// ---- 鼠标手势辅助函数（批次 21 引入） ----

// normalizeGestureAction 规范化手势动作字符串。
// [S-sig 汇编 0x1408da100]：体待专项还原，当前返回原串。
func normalizeGestureAction(action string) string {
	_ = action
	return action
}

// normalizeMouseGestureLauncherAction 规范化鼠标手势启动器动作字符串。
// [S-sig 汇编 0x1408dafe0]：体待专项还原，当前返回原串。
func normalizeMouseGestureLauncherAction(feature string) string {
	_ = feature
	return feature
}

// executeMouseGestureWindowMovePlatform 平台级鼠标手势窗口移动。
// [S-sig 汇编 0x1408e7000]：依赖 OLED 屏幕数据+窗口移动参数，体待平台层还原。
func executeMouseGestureWindowMovePlatform(screens interface{}) error {
	_ = screens
	return nil
}

// ---- BootstrapService 辅助函数（批次 21 引入） ----

// LaunchAppWithPrivilege 以指定权限启动应用。
// [S-sig 汇编 0x1408a60c0, 512B(0x200)]：签名经 asm 实证——形参1 = string（rax/rbx 装载，
// 传 normalizeAppLaunchPrivilegeMode）；形参2 = AppEntry 结构（栈槽 [rsp+0x458] 首 qword +
// duffcopy 续段），非 `int flags`。调用点实证传 ("default", entry)：
// 0x140783687 lea rbx,[rip+0x4b5caa]=0x140c39338（7B "default"）。
// asm 体：normalizeAppEntries([]AppEntry{entry}) → 空则返 24B error "入口路径不能为空" →
// normalizeAppLaunchPrivilegeMode(privilege)（magic 0x61666564+0x6c75+0x74 = "default" 判等）→
// 为 ""/"default" 则 resolveEffectiveAppLaunchPrivilege(mode, loadAppLaunchPrivilegeDefault()) →
// startApplication(normalized[0], privilege)。
// [P] 体依赖 loadAppLaunchPrivilegeDefault(0x1408a65a0)/resolveEffectiveAppLaunchPrivilege(0x140882880)/
// startApplication(0x1408a6a20)（app 启动域未落地），effective 解析与启动步待该域续作。
func (bs *BootstrapService) LaunchAppWithPrivilege(privilege string, entry AppEntry) error {
	entries := normalizeAppEntries([]AppEntry{entry})
	if len(entries) == 0 {
		return errors.New("入口路径不能为空")
	}
	mode := normalizeAppLaunchPrivilegeMode(privilege)
	if mode == "" || mode == "default" {
		// [P] 未落地：mode = resolveEffectiveAppLaunchPrivilege(mode, bs.loadAppLaunchPrivilegeDefault())
	}
	// [P] 未落地：startApplication(entries[0], mode)（app 启动域）
	return nil
}

// ---- BootstrapService 配置图标资源 ----

// launcherIconAssetVersion 图标资产注册版本常量。
// [S asm] buildLauncherConfigIconResource 0x1408a2a86 与 buildLauncherIconResource 0x1408a4fef
// 均为 `mov r11, 0xffffffffc4653600`，即 int64(-1000000000)。
const launcherIconAssetVersion int64 = -1000000000

// buildLauncherConfigIconResource 构建配置图标资源。
// [S 汇编 0x1408a28e0, 232L] 完整控制流（寄存器级逐条追踪 + 返回值构造验证）：
//
//	TrimSpace(iconData) 非空 → buildLauncherIconResource(namespace, trimmedData)
//	ref = TrimSpace(iconRef)
//	bs == nil 或 bs.assets == nil（0x398）→ {IconRef: ref}
//	ref 为空 → {}
//	cachedLauncherConfigIconAssetURL 命中 → {IconRef: ref, IconURL: cached}
//	configFile = TrimSpace(workspaceSnapshot().ConfigFile)（结构偏移 +0x10）
//	configFile 为空 → {IconRef: ref}
//	Resolve 失败 → {IconRef: ref}
//	RegisterStableBytes(namespace, contentType, data, -1000000000) 失败 → {IconRef: ref}
//	url = TrimSpace(ref.URL) → cacheLauncherConfigIconAssetURL → {IconRef: ref, IconURL: url}
//
// 注：asm 中无「失败路径回写缓存」逻辑；IconRef 恒为 TrimSpace(iconRef)，
// 不取 assetRef.ID（assetRef 仅 URL 字段被消费）。
func (bs *BootstrapService) buildLauncherConfigIconResource(namespace string, iconRef string, iconData string) LauncherIconResource {
	if trimmedData := strings.TrimSpace(iconData); trimmedData != "" {
		return bs.buildLauncherIconResource(namespace, trimmedData)
	}

	iconRef = strings.TrimSpace(iconRef)
	if bs == nil || bs.assets == nil {
		return LauncherIconResource{IconRef: iconRef}
	}
	if iconRef == "" {
		return LauncherIconResource{}
	}

	if url := bs.cachedLauncherConfigIconAssetURL(namespace, iconRef); url != "" {
		return LauncherIconResource{IconRef: iconRef, IconURL: url}
	}

	configFile := strings.TrimSpace(bs.workspaceSnapshot().ConfigFile)
	if configFile == "" {
		return LauncherIconResource{IconRef: iconRef}
	}

	data, contentType, err := launcherConfigIconStoreForConfigPath(configFile).Resolve(iconRef)
	if err != nil {
		return LauncherIconResource{IconRef: iconRef}
	}

	ref, err := bs.assets.RegisterStableBytes(namespace, contentType, data, launcherIconAssetVersion)
	if err != nil {
		return LauncherIconResource{IconRef: iconRef}
	}

	url := strings.TrimSpace(ref.URL)
	bs.cacheLauncherConfigIconAssetURL(namespace, iconRef, url)
	return LauncherIconResource{IconRef: iconRef, IconURL: url}
}

// cachedLauncherConfigIconAssetURL 取缓存中的图标资产 URL，并校验资产仍在系统中。
// [S asm 0x1408a2f00, 286L] 完整控制流：
//
//	bs == nil 或 bs.assets == nil → ""
//	key 为空 → ""
//	持 iconAssetLock(0x3e0)：owner(0x3e8) != assets 或 map(0x3f0) == nil
//	  → 重建 map 并把 sequence(0x3f8) 归零
//	→ url = TrimSpace(map[key].URL) → 解锁
//	url 为空 → ""
//	!assets.Exists(namespace, url)   // 用 url 作 id 查询
//	  → 重新持锁；若 owner 一致且 map[key].URL == url → mapdelete；解锁；返回 ""
//	Exists 命中 → 重新持锁；若 owner 一致且 map[key].URL == url
//	  → sequence++ 并回写 {URL, Sequence}（LRU Touch）；解锁；返回 url
//
// 返回单值 string（asm 0x1408a29e7 以返回值 len 判命中），非 (string, bool)。
func (bs *BootstrapService) cachedLauncherConfigIconAssetURL(namespace, iconRef string) string {
	if bs == nil || bs.assets == nil {
		return ""
	}
	key := launcherConfigIconAssetCacheKey(namespace, iconRef)
	if key == "" {
		return ""
	}
	assets := bs.assets

	bs.iconAssetLock.Lock()
	if bs.iconAssetOwner != assets || bs.iconAssetURLs == nil {
		bs.iconAssetOwner = assets
		bs.iconAssetURLs = make(map[string]launcherConfigIconAssetCacheEntry)
		bs.iconAssetSequence = 0
	}
	url := strings.TrimSpace(bs.iconAssetURLs[key].URL)
	bs.iconAssetLock.Unlock()

	if url == "" {
		return ""
	}

	if !assets.Exists(url, namespace) {
		bs.iconAssetLock.Lock()
		if bs.iconAssetOwner == assets {
			if entry, ok := bs.iconAssetURLs[key]; ok && entry.URL == url {
				delete(bs.iconAssetURLs, key)
			}
		}
		bs.iconAssetLock.Unlock()
		return ""
	}

	bs.iconAssetLock.Lock()
	if bs.iconAssetOwner == assets {
		if entry, ok := bs.iconAssetURLs[key]; ok && entry.URL == url {
			sequence := bs.iconAssetSequence + 1
			bs.iconAssetSequence = sequence
			bs.iconAssetURLs[key] = launcherConfigIconAssetCacheEntry{URL: entry.URL, Sequence: sequence}
		}
	}
	bs.iconAssetLock.Unlock()
	return url
}

// launcherConfigIconAssetCacheLimit 图标资产缓存条目上限。
// [S asm 0x1408a35a5] `cmp rsi, 0x1000`，rsi 取 hmap.count（map 首字段）。
const launcherConfigIconAssetCacheLimit = 4096

// cacheLauncherConfigIconAssetURL 把图标资产 URL 写入缓存。
// [S asm 0x1408a33c0, 225L] 完整控制流：
//
//	bs == nil 或 bs.assets == nil → 返回
//	key = launcherConfigIconAssetCacheKey(namespace, iconRef)
//	url = TrimSpace(url)
//	key 为空 或 url 为空 → 直接返回（asm 0x1408a346a 无任何 map 写入/删除）
//	持 iconAssetLock：owner 不一致或 map 为 nil → 重建 map + sequence 归零
//	→ sequence++（无条件，先于存在性判定）
//	→ 若 key 不存在且 map 长度 >= 4096：mapIter 遍历找 Sequence 最小者（LRU）并删除
//	→ map[key] = {URL: url, Sequence: sequence}
func (bs *BootstrapService) cacheLauncherConfigIconAssetURL(namespace, iconRef, url string) {
	if bs == nil || bs.assets == nil {
		return
	}
	key := launcherConfigIconAssetCacheKey(namespace, iconRef)
	url = strings.TrimSpace(url)
	if key == "" || url == "" {
		return
	}
	assets := bs.assets

	bs.iconAssetLock.Lock()
	defer bs.iconAssetLock.Unlock()

	if bs.iconAssetOwner != assets || bs.iconAssetURLs == nil {
		bs.iconAssetOwner = assets
		bs.iconAssetURLs = make(map[string]launcherConfigIconAssetCacheEntry)
		bs.iconAssetSequence = 0
	}

	bs.iconAssetSequence++
	sequence := bs.iconAssetSequence

	if _, exists := bs.iconAssetURLs[key]; !exists && len(bs.iconAssetURLs) >= launcherConfigIconAssetCacheLimit {
		evictedKey, found := "", false
		var oldest uint64
		for candidate, entry := range bs.iconAssetURLs {
			if !found || entry.Sequence < oldest {
				evictedKey, oldest, found = candidate, entry.Sequence, true
			}
		}
		if found {
			delete(bs.iconAssetURLs, evictedKey)
		}
	}

	bs.iconAssetURLs[key] = launcherConfigIconAssetCacheEntry{URL: url, Sequence: sequence}
}

// buildLauncherIconResource 构建启动器图标资源（iconData 回退路径）。
// [S asm 0x1408a4f20, 136L] 完整控制流（三个返回出口的寄存器构造均已逐条追踪）：
//
//	trimmed = TrimSpace(iconData)（第 2 个 string 参数；第 1 个为 namespace）
//	bs == nil 或 bs.assets == nil（0x398）→ {IconData: trimmed}
//	trimmed 为空 → {IconData: trimmed}
//	data, contentType, err = decodeLauncherImageDataURL(trimmed)；err != nil → {IconData: trimmed}
//	assets.RegisterStableBytes(namespace, contentType, data, -1000000000)；err != nil → {IconData: trimmed}
//	成功 → {IconData: trimmed, IconURL: assetRef.URL}
//
// 三条出口均保留 IconData，IconRef 恒为空（assetRef.ID 不参与返回构造）。
func (bs *BootstrapService) buildLauncherIconResource(namespace string, iconData string) LauncherIconResource {
	trimmed := strings.TrimSpace(iconData)
	if bs == nil || bs.assets == nil {
		return LauncherIconResource{IconData: trimmed}
	}
	if trimmed == "" {
		return LauncherIconResource{IconData: trimmed}
	}

	data, contentType, err := decodeLauncherImageDataURL(trimmed)
	if err != nil {
		return LauncherIconResource{IconData: trimmed}
	}

	ref, err := bs.assets.RegisterStableBytes(namespace, contentType, data, launcherIconAssetVersion)
	if err != nil {
		return LauncherIconResource{IconData: trimmed}
	}

	return LauncherIconResource{IconData: trimmed, IconURL: ref.URL}
}

// decodeLauncherImageDataURL 解码图标 data URL 为原始字节与内容类型。
// [S-sig asm 0x1408a5200] 签名经调用点寄存器装配确证：
// 入参 rax/rbx = dataURL；返回 rax/rbx/rcx = []byte，rdi/rsi = contentType，r8/r9 = error。
// [P] 函数体未反汇编，按 data URL（RFC 2397）规范实现，待该域专项升级为 [S]。
func decodeLauncherImageDataURL(dataURL string) ([]byte, string, error) {
	if !strings.HasPrefix(dataURL, "data:") {
		return nil, "", errors.New("decodeLauncherImageDataURL: 缺少 data: 前缀")
	}
	comma := strings.IndexByte(dataURL, ',')
	if comma < 0 {
		return nil, "", errors.New("decodeLauncherImageDataURL: 缺少数据分隔符")
	}
	header, payload := dataURL[len("data:"):comma], dataURL[comma+1:]
	contentType, isBase64 := "", false
	for _, part := range strings.Split(header, ";") {
		part = strings.TrimSpace(part)
		switch {
		case part == "":
		case strings.EqualFold(part, "base64"):
			isBase64 = true
		case !strings.Contains(part, "="):
			contentType = part
		}
	}
	if !isBase64 {
		return []byte(payload), contentType, nil
	}
	data, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, "", fmt.Errorf("decodeLauncherImageDataURL: base64: %w", err)
	}
	return data, contentType, nil
}

// parseLauncherImageDataURLHeader 为幽灵符号（symbols.txt 无此符号，头解析已内联进
// decodeLauncherImageDataURL@0x1408a5200），已删除。

// ---- 配置图标槽位辅助 ----

// launcherConfigIconSlotHasSource 判断图标槽位是否包含可用源（data 或 ref 任一 TrimSpace 非空）。
// [S asm 0x1408a2600, 35L]：先 TrimSpace(data) → 非空返 true；否则 TrimSpace(ref) → 非空返 true。
func launcherConfigIconSlotHasSource(data, ref string) bool {
	trimData := strings.TrimSpace(data)
	if trimData != "" {
		return true
	}
	trimRef := strings.TrimSpace(ref)
	return trimRef != ""
}

// launcherConfigIconAssetNamespace 根据路径前缀返回资产命名空间。
// [S asm 0x1408a26a0, 177L]：按长匹配前缀表将 launcherConfigIconSlot 的 Path 映射到命名空间字符串。
// 路径长度不足的长度的条目不参与匹配。
func launcherConfigIconAssetNamespace(path string) string {
	// 匹配按 λ 长度降序排列（长前缀优先匹配）
	type entry struct{ prefix, ns string }
	entries := []entry{
		{"preferences.consoleItems[", "icon/console"},
		{"preferences.tagCatalog[", "icon/tag"},
		{"mouseGestures.apps[", "icon/mouse-gesture-app"},
		{"oledBlackout.mediaPauseExclusions[", "icon/oled-blackout"},
		{"windowManagement.target.iconData", "icon/window-management"},
		{"twoFactor.entries[", "icon/two-factor"},
		{"bookmarks.custom[", "icon/bookmark/custom"},
		{"speedDial[", "icon/speed-dial"},
	}
	for _, e := range entries {
		if len(path) >= len(e.prefix) && path[:len(e.prefix)] == e.prefix {
			return e.ns
		}
	}
	return "icon/config"
}

// ---- 书签来源 ----

// detectBookmarkSources 探测书签来源。
// [S-sig 汇编 0x14076a380, 1152B] bookmarks_firefox.go 域专项（§10 未落地），体待还原。
func detectBookmarkSources() {}

// ---- 链接浏览器 ----

// detectLinkBrowsers 探测链接浏览器。
// [S-sig 汇编 0x1408d0b60, 3040B] 返回 []StartMenuApp（尾声 rax/rbx/rcx=rsi/rdi/rbx 三字 slice），
// 体待还原（注册表探测 + 路径解析，巨型帧）。
func detectLinkBrowsers() []StartMenuApp { return nil }
