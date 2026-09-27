package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// UsbEAm Launcher 1.0.3 — 重建源码入口
// 研究用途
// 参考：source_funcs.txt main.go 的函数布局重建
// [S-sig 0x1408d1be0]：Wails v3 入口（application.New→NewBootstrapService→attach→Run）。
func main() {
	// 1. 创建应用
	app := application.New(application.Options{
		Name:        "UsbEAm Launcher",
		Description: "Windows desktop launcher and local utility toolbox.",
		Mac:         application.MacOptions{},
	})

	// 2. 创建 BootstrapService（核心编排器）
	bs := NewBootstrapService(app)

	// 3. 初始化并附加子服务
	bs.initializeWithCheckpoint()
	bs.attachApp()
	bs.attachFileIndexService(nil)
	bs.attachWindow()

	// 4. 运行应用
	if err := app.Run(); err != nil {
		panic(err)
	}
}

// NewBootstrapService 创建启动器核心服务并装配依赖子系统。
// [S 汇编实证 0x140771d20]：resolveWorkspaceLayout → launcherConfigStoreForPath → 各 service 工厂
// 挂到 Bootstrap 字段 + 容器 map/chans 初始化。
func NewBootstrapService(app *application.App) *BootstrapService {
	workspace := resolveWorkspaceLayout()
	store := launcherConfigStoreForPath(workspace.ConfigFile)

	return &BootstrapService{
		app:                      app,
		workspace:                workspace,
		configStore:              store,
		globalHotkey:             newLauncherGlobalHotkeyManager(),
		hotkeyCaptureOwners:      make(map[string]struct{}),
		hotkeyRegistrationErrors: make(map[string]string),
		memoryRelease:            newMemoryReleaseService(),
		oledBlackout:             newOLEDBlackoutService(app),
		mouseGestures:            newMouseGestureService(),
		inputMonitor:             newInputMonitorService(),
		fileLocator:              newFileLocatorService(),
		desktopWidgets:           newDesktopWidgetService(app, workspace.ConfigFile),
		iconAssetURLs:            make(map[string]launcherConfigIconAssetCacheEntry),
		workspaceTransaction:     sync.Mutex{},
		lock:                     sync.Mutex{},
		bootstrapSnapshotRefresh: sync.Mutex{},
		backgroundAssetLock:      sync.Mutex{},
		iconAssetLock:            sync.Mutex{},
		launcherUpdateDone:       make(chan struct{}),
	}
}

// resolveProcessWorkingDirectory 解析进程工作目录。
// [S 汇编实证 0x1407a2480]：os.Getwd() 成功且非空返回；否则 filepath.Abs(".") 成功且非空返回；否则空串。
func resolveProcessWorkingDirectory() string {
	if wd, err := os.Getwd(); err == nil {
		if p := strings.TrimSpace(wd); p != "" {
			return p
		}
	}
	if abs, err := filepath.Abs("."); err == nil {
		if p := strings.TrimSpace(abs); p != "" {
			return p
		}
	}
	return ""
}

// resolveWorkspaceLayout 确定工作区布局。
// [S 汇编实证 0x1407a15e0]：resolveProcessWorkingDirectory 作为 Root 基座（对齐汇编 Getwd 优先调用）。
func resolveWorkspaceLayout() WorkspaceLayout {
	root := resolveProcessWorkingDirectory()
	if root == "" {
		if execPath, err := os.Executable(); err == nil {
			root = filepath.Dir(execPath)
		}
	}
	return WorkspaceLayout{
		Root:          root,
		ConfigFile:    filepath.Join(root, "config.json"),
		LanguageDir:   filepath.Join(root, "languages"),
		PluginDir:     filepath.Join(root, "plugins"),
		IconDir:       filepath.Join(root, "icons"),
		IndexDir:      filepath.Join(root, "index"),
		ScreenshotDir: filepath.Join(root, "screenshots"),
		WebView2Dir:   filepath.Join(root, "webview2"),
		BackgroundDir: filepath.Join(root, "backgrounds"),
	}
}

// resolveLauncherConfigFilePath 解析配置文件路径。
// [S 汇编 0x1407a1920, 0xae] 证伪纠正：旧体 `return layout.ConfigFile` 错误。
// 实义（逐指令）：
//
//	os.Getenv("USBEAM_LAUNCHER_CONFIG")（22B @0x140c611ce）→ strings.TrimSpace
//	→ 非空：filepath.Abs(v)，err==nil 返 abs，否则返原 trimmed
//	→ 空：filepath.Join(layout.Root, "UsbEAm_Launcher_Config.json")（27B @0x140c69c4e）
func resolveLauncherConfigFilePath(layout WorkspaceLayout) string {
	if v := strings.TrimSpace(os.Getenv("USBEAM_LAUNCHER_CONFIG")); v != "" {
		if abs, err := filepath.Abs(v); err == nil {
			return abs
		}
		return v
	}
	return filepath.Join(layout.Root, "UsbEAm_Launcher_Config.json")
}

// initializeWithCheckpoint 初始化并检查点恢复
// 参考 source_funcs.txt: Lines: 377 to 426
// initializeWithCheckpoint 初始化并检查点恢复。
// [S 汇编实证 0x140772920]：workspaceSnapshot → ensureWorkspaceDirectories → 读配置（存在则当日
// 备份）→ 初始化后负载安排（错误恢复/日志为 [P] 子域，此处为装配链骨架）。
func (bs *BootstrapService) initializeWithCheckpoint() {
	if bs == nil {
		return
	}
	if err := ensureWorkspaceDirectories(bs.workspace); err != nil {
		// [P] 目录创建失败路径（recovery/迁移待续）
		return
	}
	_, ok, err := loadLauncherConfigIfExists(bs.workspace.ConfigFile)
	if err == nil && ok {
		_, _ = backupLauncherConfigForToday(bs.workspace.ConfigFile, 0, time.Now())
	}
	_ = bs.workspaceSnapshot()
	// [P] 其余初始化（checkpoint/迁移/状态缓存）待字节级续作。
}

// attachApp 附加应用项。
// [S 汇编实证 0x140773400]：持锁分发各非 nil service.AttachApp(app) + syncHotkeyBindings；
// oled/desktop 已装配挂载；screenshot/plugin 未装配（nil）跳过（[P] 各子域对接待续）。
func (bs *BootstrapService) attachApp() {
	if bs == nil || bs.app == nil {
		return
	}
	if bs.oledBlackout != nil {
		bs.oledBlackout.AttachApp(bs.app)
	}
	if bs.desktopWidgets != nil {
		bs.desktopWidgets.AttachApp(bs.app)
	}
	bs.syncHotkeyBindings(bs.bindings)
}

// syncHotkeyBindings 同步热键绑定。
// [S 汇编 0x140791400, 447 行]：入参 bindings 为已从配置提取的原始绑定（调用方负责构建）；
// normalize → 锁保护写所有 bs 热键字段 → globalHotkey.Update(bindings) →
// 错误存入 hotkeyRegistrationErrors → syncAppHotkeyBindings(bs)。
//
// 汇编签名实证：rax=bs，随后 22 标量实参（9 寄存器 + 13 栈槽）为 launcherHotkeyBindings
// 结构体按 ABI 展平，非多标量参数。
func (bs *BootstrapService) syncHotkeyBindings(bindings launcherHotkeyBindings) {
	if bs == nil {
		return
	}

	bindings = normalizeLauncherHotkeyBindings(bindings)

	bs.lock.Lock()
	bs.bindings = bindings
	bs.hotkey = bindings.SummonSearch
	bs.hotkeyEnabled = bindings.SummonSearchEnabled
	bs.summonOnlyHotkey = bindings.SummonOnly
	bs.summonOnlyHotkeyEnabled = bindings.SummonOnlyEnabled
	bs.screenshotHotkey = bindings.Screenshot
	bs.screenshotHotkeyEnabled = bindings.ScreenshotEnabled
	bs.screenshotQRCodeHotkey = bindings.ScreenshotQRCode
	bs.screenshotQRCodeHotkeyEnabled = bindings.ScreenshotQRCodeEnabled
	bs.screenshotAllScreensHotkey = bindings.ScreenshotAllScreens
	bs.screenshotAllScreensHotkeyEnabled = bindings.ScreenshotAllScreensEnabled
	bs.screenshotScrollingHotkey = bindings.ScreenshotScrolling
	bs.screenshotScrollingHotkeyEnabled = bindings.ScreenshotScrollingEnabled
	bs.screenshotActiveWindowHotkey = bindings.ScreenshotActiveWindow
	bs.screenshotActiveWindowHotkeyEnabled = bindings.ScreenshotActiveWindowEnabled
	bs.screenshotFeatureEnabled = bindings.ScreenshotFeatureEnabled
	bs.lock.Unlock()

	// 通知全局热键管理器
	if bs.globalHotkey != nil {
		if err := bs.globalHotkey.Update(bindings); err != nil {
			bs.lock.Lock()
			bs.hotkeyRegistrationErrors = buildLauncherHotkeyRegistrationErrors(map[string]string{
				"global": err.Error(),
			})
			bs.lock.Unlock()
		}
	}

	// 同步应用级热键绑定
	bs.syncAppHotkeyBindings(bs.app)
}

// attachFileIndexService 附加文件索引运行时预热服务。
// [S 汇编实证 0x1407739e0]：持 @0x540 锁 → bs.fileIndex=idx → 若非 nil 且 pendingFileSearchResidentWarm
// 置位 → 清标志并立即 scheduleFileSearchMaintenance 预热（返回 true）；否则直接返回 false。
func (bs *BootstrapService) attachFileIndexService(idx fileSearchRuntimeWarmer) bool {
	if bs == nil {
		return false
	}
	bs.lock.Lock()
	bs.fileIndex = idx
	warm := false
	if idx != nil && bs.pendingFileSearchResidentWarm {
		bs.pendingFileSearchResidentWarm = false
		warm = true
	}
	bs.lock.Unlock()
	if warm {
		idx.scheduleFileSearchMaintenance(nil, true)
		return true
	}
	return false
}

// attachWindow 附加窗口。
// [S 汇编 0x140773b40, 576B] 实证流程：
//
//	entry: rax=bs, rbx=app(itab), rcx=launcherWindow(data) →
//	lock(+0x540) → bs.app(+0x190)=rbx → bs.launcherWindow(+0x198)写入gcWriteBarrier →
//	launcherWindowDestroyedToTray(+0x44c)=false → 写 word 0x100 到 +0x44d
//	(= launcherFrontendReady=false + launcherSidebarPinnedOpen=true) →
//	unlock → newobject 装配闭包（含 bs/app/launcherWindow）→ call app[0x138] →
//	检查 launcherWindow 接口：nil → showPendingLauncherReveal；非 nil → 类型断言验证后同样进入
//
// signature via asm: (bs, app interface{}, launcherWindowData interface{}) ()
// 当前保留无参形态（Wails 接线层初始化后调用），体按 asm 字段语义对齐。
func (bs *BootstrapService) attachWindow() {
	if bs == nil {
		return
	}
	bs.lock.Lock()
	bs.launcherWindowDestroyedToTray = false
	bs.launcherFrontendReady = false
	bs.launcherSidebarPinnedOpen = true
	bs.lock.Unlock()

	// asm 0x140773c1a: newobject 装配 itab → call app[0x138]
	// 当前保持委托骨架。

	// asm 0x140773cda: check launcherWindow nil → showPendingLauncherReveal(bs, window)
	// 当前 attachWindow 未显式捕获 window 参，透传 nil 骨架。
	var window interface{}
	bs.showPendingLauncherReveal(window)
}

// closeLauncherStartupDebugLog 关闭启动期调试日志文件句柄。
// [S 0x1408d2f00] 单参 *os.File 无返回。asm：test rax 判 nil→return；否则 [rax]=file 内部指针
// → call os.file.close（等价 _ = f.Close()）。
func closeLauncherStartupDebugLog(f *os.File) {
	if f == nil {
		return
	}
	_ = f.Close()
}
