// AUTO-RECONSTRUCTED SERVICE METHODS — DOMAIN: bootstrap lifecycle
// 研究用途 · UsbEAm Launcher 1.0.3 后端方法体还原
//
// 本文件承载 BootstrapService 的生命周期撤销链，体为反汇编实证（[S]）。
//
// ---- BootstrapService 字段偏移（本轮多锚点闭合，含既有实证值） ----
//
//	0x0a8 bootstrapSnapshotRefresh   0x0b0 bootstrapSnapshot   0x180 bootstrapSnapshotReady
//	0x1a0 globalHotkey（接口：itab@0x1a0 / data@0x1a8）
//	0x370 twoFactor                  0x378 memoryRelease        0x380 oledBlackout
//	0x388 windowManagement           0x390 mouseGestures        0x398 assets
//	0x3a0 backgroundAssetLock        0x3a8 backgroundAssetOwner  0x3b0 backgroundAssetPath
//	0x3c0 backgroundAssetSize        0x3c8 backgroundAssetModifiedAt  0x3d0 backgroundAssetURL
//	0x3e0 iconAssetLock              0x3e8 iconAssetOwner       0x3f0 iconAssetURLs
//	0x3f8 iconAssetSequence          0x400 screenshotPin         0x408 screenshotPreview
//	0x410 screenshotSelectionToolbar 0x418 pluginWindows         0x420 inputMonitor
//	0x428 fileIndex（接口，16B）      0x438 fileLocator          0x440 desktopWidgets
//	0x448 resetLauncherSizeOnNextShow  0x449 startupTrayMode     0x44a pendingLauncherReveal
//	0x44b allowLauncherWindowClose     0x44c launcherWindowDestroyedToTray
//	0x44d launcherFrontendReady        0x44e launcherSidebarPinnedOpen
//	0x4f8 pendingQRCodeDecodeResult    0x518 workspaceTransaction
//	0x520 workspaceDataMaintenance     0x540 lock
//
// 闭合验证：assets@0x398 与 screenshotPin@0x400 之间的 backgroundAsset* / iconAsset*
// 连续区（含 sync.Mutex、string、int64、map、uint64）逐字段尺寸累加正好落在 0x398→0x400，
// 与 fileIndex 接口占 16B（0x428→0x438）互为交叉锚点。
package main

// Shutdown 撤销全部装配服务并关闭窗口关闭闸门。
// [S 汇编 0x140773e60, 832B]（symbols.txt:18849）。
//
// asm 主干（三条前置关闭 + 一次锁内字段快照 + 13 个 nil 守卫调用）：
//
//	0x140773e82  ShutdownLauncherUpdateTasks()
//	0x140773e87  shutdownScreenshotCOMQueryWorkers()
//	0x140773e94  screenshotDXGIOutputCaptureCache.Close()
//	0x140773eb0  lock(+0x540)（cmpxchg 快路径）
//	0x140773ed9  读取字段快照：globalHotkey(+0x1a0/+0x1a8)、memoryRelease(+0x378)、
//	             oledBlackout(+0x380)、windowManagement(+0x388)、mouseGestures(+0x390)、
//	             assets(+0x398)、screenshotPin(+0x400)、screenshotPreview(+0x408)、
//	             screenshotSelectionToolbar(+0x410)、pluginWindows(+0x418)、
//	             inputMonitor(+0x420)、fileLocator(+0x438)、desktopWidgets(+0x440)
//	0x140773f57  bs.globalHotkey = nil（itab/data 双字）+ write barrier
//	0x140773fb2  allowLauncherWindowClose(+0x44b) = true
//	0x140773fc0  unlock(+0x540)（lock xadd -1）
//	0x14077401c  globalHotkey.Close()（接口方法 fun[0]：mov rcx,[itab+0x18]; call rcx）
//	0x140774052  memoryRelease.Shutdown()      0x140774082  oledBlackout.Shutdown()
//	0x1407740ad  windowManagement.Shutdown()   0x1407740d3  mouseGestures.Shutdown()
//	0x1407740f4  screenshotPin.Shutdown()      0x140774110  screenshotPreview.Shutdown()
//	0x140774128  screenshotSelectionToolbar.Shutdown()
//	0x14077413a  pluginWindows.Shutdown()      0x140774149  inputMonitor.Shutdown()
//	0x140774158  fileLocator.Shutdown()        0x140774167  desktopWidgets.Shutdown()
//	0x140774179  assets.Clear()
//
// 签名实证：函数末尾为 add rsp,0x88 / pop rbp / ret，未设置 rax/rbx —— 无返回值。
// 旧骨架的 error 位与 hotkeyCaptureOperation 锁均为占位误值，本轮按实证纠正。
// 各服务 Shutdown 的 error 返回值在 asm 中一律不检查（调用后直接推进下一项）。
func (bs *BootstrapService) Shutdown() {
	if bs == nil {
		return
	}
	bs.ShutdownLauncherUpdateTasks()
	shutdownScreenshotCOMQueryWorkers()
	screenshotDXGIOutputCaptureCacheGlobal.Close()

	bs.lock.Lock()
	globalHotkey := bs.globalHotkey
	memoryRelease := bs.memoryRelease
	oledBlackout := bs.oledBlackout
	windowManagement := bs.windowManagement
	mouseGestures := bs.mouseGestures
	assets := bs.assets
	screenshotPin := bs.screenshotPin
	screenshotPreview := bs.screenshotPreview
	screenshotSelectionToolbar := bs.screenshotSelectionToolbar
	pluginWindows := bs.pluginWindows
	inputMonitor := bs.inputMonitor
	fileLocator := bs.fileLocator
	desktopWidgets := bs.desktopWidgets

	bs.globalHotkey = nil
	bs.allowLauncherWindowClose = true
	bs.lock.Unlock()

	if globalHotkey != nil {
		_ = globalHotkey.Close()
	}
	if memoryRelease != nil {
		memoryRelease.Shutdown()
	}
	if oledBlackout != nil {
		oledBlackout.Shutdown()
	}
	if windowManagement != nil {
		windowManagement.Shutdown()
	}
	if mouseGestures != nil {
		mouseGestures.Shutdown()
	}
	if screenshotPin != nil {
		screenshotPin.Shutdown()
	}
	if screenshotPreview != nil {
		screenshotPreview.Shutdown()
	}
	if screenshotSelectionToolbar != nil {
		screenshotSelectionToolbar.Shutdown()
	}
	if pluginWindows != nil {
		pluginWindows.Shutdown()
	}
	if inputMonitor != nil {
		inputMonitor.Shutdown()
	}
	if fileLocator != nil {
		fileLocator.Shutdown()
	}
	if desktopWidgets != nil {
		desktopWidgets.Shutdown()
	}
	if assets != nil {
		assets.Clear()
	}
}
