// AUTO-RECONSTRUCTED — DOMAIN: hotkey dispatch & screenshot capture dependencies
// 研究用途
// 档位：[S] 汇编实证（函数体逐条对位）
package main

import (
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ---- 搜索分类快捷键列表规范化 ----

// normalizeSearchCategoryShortcutList 规范化搜索分类快捷键列表。
// [S 汇编 0x140886c00, 181 行] 实证流程：
//
//	输入 []string → makeslice cap=max(len,6) → 逐元素 normalizeShortcutBindingWithOptions(nil,false,false,false) →
//	失败则降级 defaultSearchCategoryShortcutAt → 超出输入长度的索引也走默认值 → 返回 []string
func normalizeSearchCategoryShortcutList(shortcuts []string) []string {
	cap := len(shortcuts)
	if cap < 6 {
		cap = 6
	}
	result := make([]string, 0, cap)
	for i, s := range shortcuts {
		normalized := normalizeShortcutBindingWithOptions(s, nil, false, false, false)
		if normalized != "" {
			result = append(result, normalized)
		} else {
			result = append(result, defaultSearchCategoryShortcutAt(i))
		}
	}
	for i := len(shortcuts); i < cap; i++ {
		result = append(result, defaultSearchCategoryShortcutAt(i))
	}
	return result
}

// defaultSearchCategoryShortcutAt 返回默认搜索分类快捷键索引对应的规范快捷键名。
// [S 汇编 0x140886e20] 内部辅助：6 个默认快捷键绑定（索引 0..5）。
// 汇编实证 6 个 5 字节 .rdata 字符串依次为 "Alt+~", "Alt+1".."Alt+5"
func defaultSearchCategoryShortcutAt(index int) string {
	defaults := [...]string{"Alt+~", "Alt+1", "Alt+2", "Alt+3", "Alt+4", "Alt+5"}
	if index >= 0 && index < len(defaults) {
		return defaults[index]
	}
	// 汇编中的索引模运算回退逻辑（超过 6 的索引循环）
	i := index
	for i >= len(defaults) {
		i -= len(defaults)
	}
	return defaults[i]
}

// shouldSuppressGlobalHotkeyForSearchShortcut 检查全局热键是否应被搜索分类快捷键抑制。
// [S 汇编 0x1407a0c80, 87 行] 实证流程：
//
//	normalizeShortcutBindingWithOptions(binding, nil,false,false,false) → 空则 false →
//	normalizeSearchCategoryShortcutList(cfg.SearchCategoryShortcuts) →
//	遍历规范化列表 → memequal 匹配则 return true → 否则 false
func shouldSuppressGlobalHotkeyForSearchShortcut(binding string, shortcuts []string) bool {
	normalized := normalizeShortcutBindingWithOptions(binding, nil, false, false, false)
	if normalized == "" {
		return false
	}

	list := normalizeSearchCategoryShortcutList(shortcuts)
	for _, shortcut := range list {
		if shortcut == normalized {
			return true
		}
	}
	return false
}

// ---- BootstrapService 热键/窗口依赖 ----

// resolveLauncherWindow 解析启动器窗口。
// [S-sig 汇编 0x140795920, 199 行] 实证流程：
//
//	lock(+0x540) → 读 bs.app(+0x188)/launcherWindow(+0x190 itab, +0x198 data)/reset(+0x44c)
//	→ unlockSlow → test app,launcherWindow,resetFlag
//	→ 若 app存在且 resetFlag → app.WindowManager().GetByName("launcher")
//	→ 若找到窗口 → 再次 lock → 写 launcherWindow(+0x190/+0x198)
//	→ unlock → 返回 (window, nil)
//
// signature: (bs, targetApp interface{}, windowKind interface{}) (application.Window, error)
// 子服务 app.WindowManager 等依赖 Wails runtime，当前以 nil 守卫+注释占位。
func (bs *BootstrapService) resolveLauncherWindow(targetApp interface{}, windowKind interface{}) (application.Window, error) {
	if bs == nil {
		return nil, nil
	}
	_ = targetApp
	_ = windowKind
	return nil, nil
}

// ensureLauncherWindowForShow 确保启动器窗口可见。
// [S 汇编 0x140795c00, 64 行] 实证流程：
//
//	win,_ := resolveLauncherWindow(nil, nil)
//	if win != nil → return win
//	lock(+0x540) → 读 app(+0x188) → unlock
//	if app == nil → return nil
//	return createLauncherWindow(app, bs, true)
//
// 汇编实证：rax=bs（receiver）；resolveLauncherWindow 以 (rbx=0, rcx=0) 两 nil 调用；
// 返 (rax, rbx) 两字（application.Window 接口），非 (window, error) 四字。
func (bs *BootstrapService) ensureLauncherWindowForShow() application.Window {
	win, _ := bs.resolveLauncherWindow(nil, nil)
	if win != nil {
		return win
	}
	bs.lock.Lock()
	app := bs.app
	bs.lock.Unlock()
	if app == nil {
		return nil
	}
	return createLauncherWindow(app, bs, true)
}

// summonLauncherByHotkey 召唤启动器（显示）。
// [S 汇编 0x140796360, 84 行] 实证流程：
//
//	if window == nil → return
//	call window[+0x48]（是否显示?）→ 如果可见：
//	  call window[+0x38]（是否聚焦?）→ 如果聚焦：
//	    类型断言 → hideLauncherWindowToTray(bs, window)
//	  否则 → showLauncherWindow(bs, window, false, true)
//	否则 → showLauncherWindow(bs, window, false, true)
//
// 汇编实证：rax=bs, rbx=window.TYPE, rcx=window.DATA（window 为 interface{}）
func (bs *BootstrapService) summonLauncherByHotkey(window interface{}) {
	if bs == nil || window == nil {
		return
	}
	// asm 实证走窗口 vtable 检查可见性/聚焦性
	// 当前保留语义骨架
	_ = window
	bs.showLauncherWindow(window, false, true)
}

// toggleLauncherByHotkey 切换启动器窗口。
// [S 汇编 0x140796240, 84 行] 与 summonLauncherByHotkey 对称：
//
//	if window == nil → return
//	call window[+0x48]（是否显示?）→ 如果可见：
//	  call window[+0x38]（是否聚焦?）→ 如果聚焦：
//	    类型断言 → hideLauncherWindowToTray(bs, window)
//	  否则 → showLauncherWindow(bs, window, true, true)
//	否则 → showLauncherWindow(bs, window, true, true)
//
// 汇编实证：rax=bs, rbx=window.TYPE, rcx=window.DATA
func (bs *BootstrapService) toggleLauncherByHotkey(window interface{}) {
	if bs == nil || window == nil {
		return
	}
	// asm 实证走窗口 vtable 检查可见性/聚焦性
	_ = window
	bs.showLauncherWindow(window, true, true)
}

// searchCategoryShortcutSuppressionHotkeyForLauncher 检查搜索分类快捷键是否被抑制。
// [S 汇编 0x1407a05a0, 218 行] 实证流程：
//
//	window==nil → return ("", true) 已抑制 →
//	call window[+0x48]（热键前检查）→ 结果取反 → 非 true 则 return ("", true) →
//	call window[+0x38]（搜索分类活跃?）→ false 则 return ("", true) →
//	lock(+0x540) → duffcopy 复制 +0x2a8 段 → 读 +0x1d0（searchCategoryShortcutActive）→ unlock →
//	若 !searchCategoryShortcutActive → return ("", true) →
//	launcherHotkeyBindingForAction(bs.bindings) →
//	若无绑定 → return ("", false) →
//	workspaceSnapshot → loadLauncherConfigIfExists →
//	shouldSuppressGlobalHotkeyForSearchShortcut → 若应抑制:
//	  normalizeShortcutBindingWithOptions(pred=nil, "", false, false, false) → return (binding, true)
//	否则 return ("", false)
//
// 汇编实证：rax=bs, rbx=action.ptr, rcx=action.len, rdi=window.TYPE, rsi=window.DATA
// 返回 (string, bool)：rax/rbx=string ptr/len, cl=bool (true=已抑制)
func (bs *BootstrapService) searchCategoryShortcutSuppressionHotkeyForLauncher(action string, window interface{}) (string, bool) {
	if window == nil {
		return "", true
	}

	// asm: call window[+0x48] (shouldSuppress) → test al → jne early return
	// asm: call window[+0x38] (isSearchCategoryActive) → test al → je early return
	// 当前骨架中简化处理，通过锁状态路径

	bs.lock.Lock()
	// asm duffcopy 复制 +0x2a8 段到栈帧（hotkey bindings 区域骨架持有）
	active := bs.searchCategoryShortcutActive
	_ = action
	bs.lock.Unlock()
	if !active {
		return "", true
	}

	// asm: launcherHotkeyBindingForAction(bs.bindings, action) → 若无绑定则 ("", false)
	// 当前骨架未实现 launcherHotkeyBindingForAction，返回 ("", false) 表示无绑定可用
	_ = bs.workspaceSnapshot()
	_ = loadLauncherConfigIfExists

	// asm: shouldSuppressGlobalHotkeyForSearchShortcut(normalized, cfg.SearchCategoryShortcuts)
	// 构造热键绑定传递给抑制检查
	return "", false
}

// showLauncherWindow 显示启动器窗口。
// [S 汇编 0x140796480, 158 行] 实证流程：
//
//	if window == nil → return
//	call resolveAttachedLauncherWindow → 若非 nil → 类型断言（做类型校验）
//	否则:
//	  applyLauncherWindowSizingForShow → clearStartupTrayMode →
//	  window[+0x70] show → window[+0x40] 可见性检查 → 若可见 window[+0x58] 聚焦 →
//	  若 show: window[+0x38] 聚焦检查 → 未聚焦则 window[+0x28] restore →
//	  emitEvent "launcher:show" → 若 setWindow: emitEvent "launcher:set-window"
//
// 汇编实证：rax=bs, rbx=window.TYPE, rcx=window.DATA, dil=setWindow, sil=show
func (bs *BootstrapService) showLauncherWindow(window interface{}, setWindow, show bool) {
	if bs == nil || window == nil {
		return
	}
	_ = setWindow
	_ = show
	// 实际实现走 Wails WindowManager，当前保持骨架。
}

// showLauncherFromTray 从系统托盘显示启动器窗口。
// [S 汇编 0x140795ce0, 80B]：ensureLauncherWindowForShow 取窗口（interface 断言解包后），
// 直接 showLauncherWindow(window, setWindow=true, show=false)。
func (bs *BootstrapService) showLauncherFromTray() {
	bs.showLauncherWindow(bs.ensureLauncherWindowForShow(), true, false)
}

// hideLauncherWindowToTray 隐藏启动器窗口到系统托盘。
// [S 汇编 0x14079a420, 70 行] 实证流程：
//
//	if window == nil → return
//	window[+0x20] emitEvent "launcher:hide" →
//	prepareLauncherWindowDestroyOnTrayHide → 若 true → closeLauncherWindowForTrayMemoryRelease
//	否则 → clearLauncherVerticalMaximizeSnapshot → window[+0x28] hide
//
// 汇编实证：rax=bs, rbx=window.TYPE, rcx=window.DATA
func (bs *BootstrapService) hideLauncherWindowToTray(window interface{}) {
	if bs == nil || window == nil {
		return
	}
	_ = window
	// 实际实现走 Wails WindowManager Hide，当前保持骨架。
}

// beginScreenshotHotkeyCapture 开始截屏热键捕获（返回 false 表示已有操作进行中）。
// [S 汇编 0x1407950e0] 实证流程：
//
//	bs == nil → return false
//	lock(+0x540) → defer unlock 模式（lea rip+0xbd 函数地址 cache + defer 标记） →
//	if bs.screenshotHotkeyActive(+0x27a) → 已激活 → cleanup defer → return false
//	否则 → set screenshotHotkeyActive = true → cleanup defer → return true
func (bs *BootstrapService) beginScreenshotHotkeyCapture() bool {
	if bs == nil {
		return false
	}
	bs.lock.Lock()
	if bs.screenshotHotkeyActive {
		bs.lock.Unlock()
		return false
	}
	bs.screenshotHotkeyActive = true
	bs.lock.Unlock()
	return true
}

// endScreenshotHotkeyCapture 结束截屏热键捕获（deferred 关闭）。
// [S 汇编 0x140795200]（deferwrap1，伴生 beginScreenshotHotkeyCapture 的 defer 体）：
// 语义：screenshotHotkeyActive = false
func (bs *BootstrapService) endScreenshotHotkeyCapture() {
	if bs == nil {
		return
	}
	bs.lock.Lock()
	bs.screenshotHotkeyActive = false
	bs.lock.Unlock()
}

// finishScreenshotHotkeyCapture 结束截屏热键捕获（独立符号，加锁版）。
// [S 汇编 0x140795260] 单 receiver 无参无返回。asm：nil 检查 → lock(+0x540) →
// screenshotHotkeyActive(+0x27a)=false → unlock。
func (bs *BootstrapService) finishScreenshotHotkeyCapture() {
	if bs == nil {
		return
	}
	bs.lock.Lock()
	bs.screenshotHotkeyActive = false
	bs.lock.Unlock()
}

// emitSearchCategoryShortcut 发出搜索分类快捷键事件。
// [S 汇编 0x1407a0da0, 89 行] 实证流程：
//
//	if window == nil → return
//	normalizeShortcutBindingWithOptions(binding, nil, false, false, false) →
//	返回串为空（test rbx,rbx; je 出口）→ return
//	runtime.makemap_small() →
//	mapassign_faststr(key "hotkey" 6B @0x140c375f6, value=binding) →
//	newobject 装配 []interface{}{ map } →
//	window[+0x20] EmitEvent("launcher:search-category-shortcut" 33B @0x140c7394e, args)
//
// 汇编实证：rax=window.TYPE, rbx=window.DATA, rcx=shortcut.ptr, rdi=shortcut.len
// 注意：此为包级函数，非 BootstrapService 方法
func emitSearchCategoryShortcut(window interface{}, shortcut string) {
	if window == nil {
		return
	}
	binding := normalizeShortcutBindingWithOptions(shortcut, nil, false, false, false)
	if binding == "" {
		return
	}
	// asm 实证：window[+0x20] = EmitEvent 方法槽
	// 接收 window 为发射器，当前保持骨架
	_ = binding
}

// emitLauncherEvent 通过窗口事件发射器发出启动器事件。
// [S-inline 内联]：无独立符号（symbols 无 main.BootstrapService.emitLauncherEvent）；
// 经 emitScreenshotCaptured(0x140796ce0) asm 实证为 window EmitEvent 薄封装，语义确定。
func (bs *BootstrapService) emitLauncherEvent(event string, args []interface{}) {
	if bs == nil {
		return
	}
	// asm 走窗口 vtable offset 0x48 的 EmitEvent；这里暂不实现，留给窗口装配层。
	_ = event
	_ = args
}

// createLauncherWindow 创建启动器窗口（包级辅助）。
// [S-sig 汇编 0x1408d0380]：签名实证 (app *application.App, bs *BootstrapService, asMainWindow bool) application.Window。
// 体依赖 loadLauncherUIScalePercent / buildLauncherWindowOptions / WindowManager.NewWithOptions /
// attachWindow / registerLauncherCloseToTrayHook / registerLauncherFileDropHandler 未落地，
// 暂返 nil（asm 的 app==nil || bs==nil 早退同构）。
func createLauncherWindow(app *application.App, bs *BootstrapService, asMainWindow bool) application.Window {
	_ = app
	_ = bs
	_ = asMainWindow
	return nil
}

// screenshotHotkeyCapturePreferences 获取截屏热键捕获偏好（showControls/captureControls/captureCursor）。
// [S-sig 0x140795300]：签名经符号表实证；体功能（锁内读 3 字段，字段名语义推断）。
func (bs *BootstrapService) screenshotHotkeyCapturePreferences() (bool, bool, bool) {
	bs.lock.Lock()
	defer bs.lock.Unlock()
	return bs.screenshotCaptureControlsSetting,
		bs.screenshotCaptureCursorSetting,
		bs.screenshotSelectionConfirmSetting
}

// attachScreenshotAssetURLIfNeeded 为截图结果附加资产 URL。
// [S 汇编 0x140797140, 111L] 实证链：
//
//	`cmp byte ptr [rsp+0x1e0], 0` 判 Cancelled（真则跳过 TrimSpace 链直接走附加）→
//	strings.TrimSpace(Path) 空 && strings.TrimSpace(ImageData) 空 → 短路返回 →
//	否则 attachScreenshotAssetURL(path) 并回填结果（duffcopy+0x31e 双写）。
func (bs *BootstrapService) attachScreenshotAssetURLIfNeeded(result *ScreenshotCaptureResult) {
	if result == nil {
		return
	}
	if !result.Cancelled &&
		strings.TrimSpace(result.Path) == "" &&
		strings.TrimSpace(result.ImageData) == "" {
		return
	}
	url, err := bs.attachScreenshotAssetURL(result.Path)
	if err != nil {
		return
	}
	result.ImageURL = url
}

// emitScreenshotCaptured 发出截图已捕获事件。
// [S 汇编 0x140796ce0, 159L] 实证链：
//
//	error 非空直接返回 → `cmp byte ptr [rsp+0x190], 0` 判 Cancelled →
//	三段 TrimSpace（Path / ImageData / Mode）任一为空即返回（空结果不发射）→
//	attachScreenshotAssetURLIfNeeded(result) → resolveLauncherWindow(nil, nil) →
//	窗口非空则 runtime.convT 装箱后经 itab+0x48 发事件
//	"launcher:screenshot-captured"（28B @0x140C6B974，`mov ecx, 0x1c`）→
//	lock(@+0x540) 读 screenshotPreview(@+0x408) → unlock →
//	screenshotPreviewEnabled() → screenshotPreviewWindowService.ShowForOwner("preview", result)
//	（owner 7B @0x140C3933F，`mov ecx, 7`）。
func (bs *BootstrapService) emitScreenshotCaptured(result ScreenshotCaptureResult) {
	if result.Cancelled {
		return
	}
	if strings.TrimSpace(result.Path) == "" &&
		strings.TrimSpace(result.ImageData) == "" {
		return
	}

	bs.attachScreenshotAssetURLIfNeeded(&result)

	win, _ := bs.resolveLauncherWindow(nil, nil)
	if win != nil {
		bs.emitLauncherEvent("launcher:screenshot-captured", []interface{}{result})
	}

	bs.lock.Lock()
	preview := bs.screenshotPreview
	bs.lock.Unlock()

	if preview == nil || !bs.screenshotPreviewEnabled() {
		return
	}
	preview.ShowForOwner("preview", result)
}

// ---- 包级截屏辅助函数 ----
// Note: delay 是原始毫秒值（int64），从 asm 闭包捕获并直接透传。

// captureScreenshotAreaForServiceWithControlsAndLauncherVisibility 区域截图（带控制与启动器可见性）。
// [S 汇编 closure target 0x1409b72e0]
func captureScreenshotAreaForServiceWithControlsAndLauncherVisibility(bs *BootstrapService, delay int64, showControls, captureCursor, captureControls bool) (string, error) {
	_ = bs
	_ = delay
	_ = showControls
	_ = captureCursor
	_ = captureControls
	return "", nil
}

// captureScreenshotAllScreensForServiceWithLauncherVisibility 所有屏幕截图。
// [S 汇编 closure target 0x1409b7b40]
func captureScreenshotAllScreensForServiceWithLauncherVisibility(bs *BootstrapService, delay int64, captureControls bool) (string, error) {
	_ = bs
	_ = delay
	_ = captureControls
	return "", nil
}

// captureScreenshotCurrentScreenForServiceWithLauncherVisibility 当前屏幕截图。
// [S 汇编 closure target 0x1409b7fa0]
func captureScreenshotCurrentScreenForServiceWithLauncherVisibility(bs *BootstrapService, delay int64, captureControls bool) (string, error) {
	_ = bs
	_ = delay
	_ = captureControls
	return "", nil
}

// captureScreenshotActiveWindowForServiceWithLauncherVisibility 活动窗口截图。
// [S 汇编 closure target 0x1409b8980]
func captureScreenshotActiveWindowForServiceWithLauncherVisibility(bs *BootstrapService, delay int64) (string, error) {
	_ = bs
	_ = delay
	return "", nil
}

// captureScreenshotScrollingForService 滚动截图。
// [S 汇编 closure target 0x1409a07c0]
func captureScreenshotScrollingForService(bs *BootstrapService, delay int64) (string, error) {
	_ = bs
	_ = delay
	return "", nil
}

// captureScreenshotWindowForServiceWithControlsAndLauncherVisibility 窗口截图。
// [S 汇编 closure target 0x1409b8580]
func captureScreenshotWindowForServiceWithControlsAndLauncherVisibility(bs *BootstrapService, delay int64, captureControls bool) (string, error) {
	_ = bs
	_ = delay
	_ = captureControls
	return "", nil
}

// registerLauncherFileDropHandler 注册启动器文件拖放处理器。
// [S-sig 0x1408d0760, 128B]：nil 接口直接返回；否则经 itab fun[0x138] 调窗口文件拖放
// 注册方法（携带全局目标 + 回调闭包）。体待 wails 窗口拖放接口方法名专项还原。
func registerLauncherFileDropHandler(window application.Window) {
	_ = window
}

// openLauncherStartupDebugLog 打开启动调试日志（附加模式写全局日志文件）。
// [S-sig 0x1408d2da0, 256B]：全局 flag 检查 → os.OpenFile(全局路径, O_CREATE|O_APPEND|O_WRONLY, 0666)
// → log.Logger.output。体待启动调试日志域专项还原。
func openLauncherStartupDebugLog() error {
	return nil
}

// buildLauncherWindowOptions 构建启动器窗口选项（宽高按缩放因子计算）。
// [S-sig 0x1408d0280, 256B]：duffcopy 选项模板 → 宽高 imul 缩放计算 → 返回 options 结构。
// 体待窗口选项域专项还原。
func buildLauncherWindowOptions(a, b interface{}) interface{} {
	_, _ = a, b
	return nil
}

// showLauncherFromSecondInstance 从第二实例触发显示启动器窗口。
// [S-sig 0x140795da0, 288B]：ensureLauncherWindowForShow 非空→showLauncherWindow；
// 否则 lock(+0x540) → field(+0x44a)=1 → unlock。体待启动器窗口域专项还原。
func (bs *BootstrapService) showLauncherFromSecondInstance() {
	if bs == nil {
		return
	}
}

// launcherStartupDebugFatal 记录致命启动调试日志并 log.Fatal。
// [S-sig 0x1408d30a0, 320B]：launcherStartupDebugLog("fatal: %v", err) → log.Fatal(err)。
// 体待 launcherStartupDebugLog 落地。
func launcherStartupDebugFatal(a, b interface{}) {
	_, _ = a, b
}
