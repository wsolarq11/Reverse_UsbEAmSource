# UsbEAm Launcher 1.0.3 — 逆向还原工程 · 技术交接文档

> 目的：把已完成的取证/还原进展、方法、环境与工具固化，让**新会话**可以无缝续接"后端服务方法体逐函数重建"这一独立大工程。
> 目标口径（用户已确认，2026-09-20 升级为**总进度 100%**）：**4,754 蓝图函数 100% 还原、110 个未落地原始文件 100% 落地、`UNMARKED=0` 且 `P=0`**；同时全程保持**可编译 + 功能一致**（非字节级同哈希）。
> 研究用途，版权（©2026 DOGFIGHT360）合规。

---

## 1. 一句话现状（批次 254 · oledblackout 短函数 +4 · 2026-09-27）

前端层已**字节级完整还原**并验证；Go 后端**类型层（507 个结构体）已全量还原且能编译**；函数/方法体还原已推进到**批次 54**，后端非测试代码 **25,198 行 / 125 个 .go 文件**，测试 **54 文件 / 5,866 行**。**批次 40 已闭环（appicon_windows.go 全量 [S]）；批次 41 已闭环（bootstrapservice_callees.go 34 未标清零）；批次 42 已闭环（screenshot_png_fast.go 自定义 PNG 编码器 14 函数 [S]）；批次 43 已闭环（screenshot image_budget 内存预算/受限读取/解码链 8 函数 [S]）；批次 44 已闭环（screenshot_capture_windows GDI 捕获核心 6 函数 [S]）；批次 45 已闭环（screenshot_cursor_windows 光标快照数据搬运层 5 函数 [S]）；批次 46 已闭环（screenshot_cursor_draw_windows 光标绘制收口链 11 函数 [S]）；批次 47 已闭环（screenshot_virtual_windows GDI 虚拟屏捕获链 5 函数 [S]）；批次 48 已闭环（screenshot_hdr_windows DXGI/HDR 模式解析与入口探测 6 函数 [S]）；批次 49 已闭环（screenshot_dxgi_debug_windows DXGI 调试格式化链 4 函数 [S]）；批次 50 已闭环（screenshot_dxgi_com_windows DXGI COM 薄包装 3 函数 [S] + 1 [S-inline]）；批次 51 已闭环（screenshot_dxgi_enum_windows DXGI 枚举链 6 函数 [S] + 3 [S-inline]）；批次 52 已闭环（screenshot_capture_backend_windows 收口链 6 函数 [S] + 2 [P] DXGI 捕获核心入口）；批次 53 已闭环（screenshot_dxgi_capture_windows DXGI 捕获域 10 函数 [S] + 3 深层 [S-sig]，两个 [P] 核心入口转 [S]）；批次 54 已闭环（screenshot_dxgi_cache_windows DXGI 输出复制缓存域 19 函数 [S] + screenshot_dxgi_deep_windows 10 深层 [S-sig] + COM Release/QueryInterface 2 函数 [S]）。**

**门禁活体实测（2026-09-26，批次 208）**：`go1.25.12 build ./backend` / `go vet ./backend` / `go test ./backend` 三项 **EXIT=0**（`ok changeme/backend`）。**批次 202 落地 filesearch 名称搜索计划域 6 函数 [S]**（`shouldPrioritizeSearchTerm`/`estimateSearchTermCandidateCount`/`selectDriverTermIndex`/`buildOrderedNameTermIndices`/`buildIndexSearchPlan`/`matchNodeNameTerms` + 3 结构），`matchNodeNameTerms` 转正；**批次 203 落地 qrcode 剪贴板读写链 2 函数 [S-sig]→[S]**；**批次 204 落地 twofactor 供给 URI 图标编解码 2 函数 [S-sig]→[S]**；**批次 205 落地 twofactor 供给 URI 标签构造 buildTwoFactorLabel [S-sig]→[S]**（修正 issuer 固定为 "Steam" 大写）；**批次 206 落地 filesearch 排序比较器域 4 函数 [S]**（Resolve/Release/compareText/compareCandidate，compareCandidate 由 [P] 转正并订正 int 三态签名）；**批次 207 落地 filesearch 拼音模糊匹配域 3 函数 [S]**（fileSearchPinyinFuzzyMatcher.Match + matchFileSearchPinyinTerm + matchNodeNameTermsWithPinyin，并修正 nameSearchTerm 的 +0x30 为 matcher 指针字段）；**批次 208 落地 filesearch 拼音分类域 3 函数 [S]**（MatchExact + MatchPrefix + classifyVolumePinyinMatch，classifyVolumePinyinMatch 的 r9 参数反推确认为 *nameSearchTerm）；**批次 209 落地 filesearch 拼音别名查询链 3 函数 [S]**（volumePinyinIndex.aliasesForNode 二分查记录 + volumeIndexReadView.pinyinAliasesForNode delta 覆盖 + matchHierarchyTermWithPinyin 层级匹配收口）；**批次 210 落地 filesearch 节点访问与墓碑链 4 函数 [S]**（nodeAt 双路径取节点 + nodeAtIndex 护栏 + nodeName 名称切片 + IsTombstonedDescendant 墓碑后代判定，新增 volumeIndexTombstoneLookup 结构）；**批次 211 落地候选节点拼音匹配主函数 matchSearchCandidateNodeWithPinyin [S]**（四段式：普通术语匹配 + driver 淘汰 + mask/others 判定 + 沿父链补齐层级术语）；**批次 212 落地非拼音候选匹配 matchSearchCandidateNode [S]，并修正 nameFrequencyIndex→namePrefixBucketCounts 别名、estimateSearchTermCandidateCount 字段、matchNodeNameTerms 签名收敛为 (b,terms,caseSensitive,plan)**；**批次 213 落地 filelocator 运行时两个 [P] 升档 [S]（acquireFileLocatorContentScan 签名 bool→error + 单缓冲信号量 select 体；waitIfPaused 签名 (ctx)bool→(ctx,generation,startedAt)error + 暂停等待循环体，均汇编逐寄存器实证）**；**批次 214 落地 filelocator 运行时两个 [P] 升档 [S-sig]（runSearch 签名 (prepared 指针,ctx)error→(prepared 值,ctx,generation)void；finishSearch 签名 (cancelled)→(generation,startedAt,lastError,cancelled)，均汇编逐寄存器实证，体留待专项）**；**批次 233 落地 screenshotAccessibleLocation 补 VARIANT 参数**；**批次 234 落地 screenshot_pin 域 bounds 签名链订正（storeSnapshotBounds/storeSnapshotLocked/createPinnedWindow/registerPinnedWindow/createWebviewPinnedWindow/buildScreenshotPinWindowHTML 六 [P] 升档 [S-sig] + createScreenshotNativePinWindow 签名订正 (pinName,pinGeneration,image)→(bounds application.Rect)）**；**批次 235 落地 qrcode 标注动作链 hwnd 订正（applyAnnotationToolbarAction [P] 升档 [S-sig] + handleAnnotationToolbarActions/undoAnnotationAndInvalidate/invalidateAnnotationDirtyRect 三签名补 hwnd uintptr）**；**批次 236 落地 captureScreenshotWithLauncherVisibility [P] 升档 [S-sig]（(service,hideLauncher)→(service,delay,hideLauncher,cb func() ([]byte,bool,error)) 返回 ([]byte,bool,error)）**；**批次 237 落地 initializeScreenshotCOMThreadMode [P] 升档 [S-sig]（()→(allowChangedMode bool)(func(),error)）**；**批次 238 落地 screenshotNativePinDrawOverlay [P] 升档 [S-sig]（(img,bounds)→(img,bounds,showCloseButton,showOpacityPanel,showHighlight bool,tick int,opacity float64)）**；**批次 239 落地 screenshotCOMQueryWorkerPool.enqueue [P] 升档 [S-sig]（(req)→(x,y int32,targetWindow uintptr,priority uint8,busyError error)(*screenshotCOMQueryRequest,error)）**；**批次 240 落地 probeLauncherUpdatePackageRange [P] 升档 [S-sig]（(ctx,url)→(ctx,access LauncherNetworkAccess,url)(bool,int64,error)）**；**批次 241 落地 downloadLauncherUpdatePackageRange [P] 升档 [S-sig]（(ctx,url,r,dir)→(ctx,access,url,file *os.File,r)error）**；**批次 242 落地 downloadLauncherUpdatePackageSequentially [P] 升档 [S-sig]（(ctx,pkg,dir)→(ctx,taskID,access,path,url,sha256,expectedSize)(int64,error)）**；**批次 244 落地 launcherupdate 下载链三函数签名订正（downloadLauncherUpdatePackageWithWorkspace/concurrently 双 [P] 升档 [S-sig]，均返回 (int64,error)；修正批 242 sequentially 漏 concurrent bool+root string 两参，workspace 实为 WorkspaceLayout 值传）**；**批次 245 落地 launcherupdate 域 [P] 全清零（fetch/prepare/runLauncherUpdateTask/beginLauncherUpdateTask 四升档 [S-sig]，isLauncherVersionUpdateAvailableText 订正双参；beginLauncherUpdateTask 实为 (ctx)→(ctx,taskID,done func(),error) 非 (version)→(int64,error)）**；**批次 246 订正 normalizeLauncherUpdatePackageRoot 函数体（旧错误 Clean/Abs，新 asm 直译 Replace("\\","-")+TrimSpace+空/含-/. /.. 判空）**；**批次 247 落地 oledblackout ghost [P] 清零（删除脚手架 newOLEDLifecycleContext，新增 ensureLifecycleLocked 方法 [S]，newOLEDBlackoutService 调用点改为 s.ensureLifecycleLocked()）**；**批次 248 落地已存在文件短函数批量 [S] +11（IndexNode.IsDirectory/IsDeleted、oledBlackoutService.beginBackgroundActivity/endBackgroundActivity、closeLauncherStartupDebugLog、windowsOLEDBlackoutCursorController.request/Hide/Show、screenshotCursorCurrentlyShowing/adjustScreenshotCursorVisibility/ensureScreenshotCursorVisible + procShowCursor）**；**批次 249 落地 oledblackout 域短函数批量 [S] +9（normalizeOLEDBlackoutMediaPauseExclusionPath/PathBase/ProcessName、idleRunCurrentLocked、stopFocusRetryLocked、browserMediaContinuity.clear、armInputDismissGuardLocked、volumeIndexReadLease.Release、GetPinnedScreenshotStates）并订正 normalizeLauncherUpdatePackageRoot 的 `\`→`/` 常量（batch 246 误读 `-`）**；**批次 250 落地 BootstrapService/ShellVerb 短函数 +4（InvokeShellVerb、emitScreenshotCaptureAccepted [S-sig]、screenshotCaptureAcceptedNotifier.Emit、clearStartupTrayMode）**；**批次 251 落地 BootstrapService/FileSearch 短函数 +3（finishScreenshotHotkeyCapture、timeToUnixNano、ensureWorkspaceDirectories 方法改名订正 [S-sig]→[S]）**；**批次 252 落地 filesearch 域 +3 [S]（addNamePrefixBucketCount、volumeIndexMappedFile.Close、volumeNameTrigramIndex.close）**；**批次 253 落地 filesearch WAL 截断 + inputmonitor 平台拥有关闭链 +5（truncateOpenWALAtValidOffset、advancePlatformGeneration、stopPlatformThread [S-sig]、stopPlatformOwned、closePlatformOwned）**；**批次 254 落地 oledblackout 短函数 +4 [S]（oledBlackoutMediaPauseExclusionKey、parseOLEDBlackoutScreenNumber、sortOLEDBlackoutScreens、oledBlackoutReadInputSnapshot）**；真函数口径 2786（58.60%），未落地文件差集 57。明细见 `docs/acceptance/batch196.md` … `batch254.md`。

**当前进度分母（本次校准新增，此后一律以此为准）**：

| 指标 | 实测值 | 目标（总进度 100%） | 取证方式 |
|---|---|---|---|
| 蓝图函数项 | 4,754 | 4,754 | `docs/goresym/source_funcs.txt` 中 `Lines: a to b (n)` 条目计数 |
| 蓝图源文件数 | 145 | 145 | 同文件 `^File: ` 条目计数 |
| 原始源码规模 | ≈104,374 行 | ≈104,374 行 | 每文件最大行号求和（闭包共享父函数区间，属上界估计） |
| 已重建函数 | 2893 | 4,754 | `bash tools/count_funcs.sh` 实测（批次 278 后） |
| 真函数（S+S-inline+S-sig） | 2852 | 4,754 | 同上，**批次 278 达 59.99%** |
| 文件覆盖 | 106/144 | **100%（144/144）** | backend 非测试文件名与蓝图 `File:` 清单逐个对名 |
| 未落地原始文件 | 38 | **0** | 同上差集（活体实测 2026-09-30 批次 278 重跑，较 §10 的 57 已减 19），清单见 §10 |
| UNMARKED | 0 | **0** | `bash tools/count_funcs.sh` 实测 |
| [P] 存根 | 41 | **0** | 同上（批次 278 持平） |

**⚠️ 口径纪律（本文件历史数字曾三度失真）**：§1 曾长期写「批次 34 / `FUNCS=900`」，与正文实际进度不符；批次 39 记录的 `S=552/S-sig=97/P=46/UNMARKED=259` 与活体实测不符，且分项相加 955 ≠ `FUNCS=969`（自相矛盾）。**任何批次记录落笔前必须先跑 `bash tools/count_funcs.sh` 取活体数字，禁止抄上一批的数字改一改。**

**已闭环域（批次 19 起累加，批次 35–39 新增 5 域）**：
- memoryrelease 执行/调度域：13 函数全部 [S]
- SQLite 域：4 函数全部 [S]
- mouseGesture 执行域：4 函数全部 [S]
- OLEDBlackout 执行域：6 方法全部 [S]（+shouldSuppressOLEDBlackoutHotkey）
- **Screenshot 截图历史子域：批次 27 闭环**
- **Screenshot 体/辅助域：~26 函数已升级为 [S]**
- **hotkey 三角域**（批次 24）
- **hotkey 抑制域**（批次 25）
- **launcher asset aux 子域**（批次 28）
- **launcher asset icon 中枢域**（批次 29）
- **attach*IconURLs 集群**（批次 30）
- **launcherasset 驱逐子域（批次 31 闭环）**
- **visitLauncherConfigIconSlots 全遍历（批次 32 闭环）**：回调签名 `func(launcherConfigIconSlot) bool`（9 区段全覆盖，12 格式串解码，643L asm 实证）
- **launcherasset 读取/服务链（批次 33 新闭环）**：`ReadBytes`(261L) + `Exists`(142L) + `ServeAssetRequest`(536L) + `registerStable`(404L) 四函数签名重写与体重建
- **launcherasset 文件注册链（批次 34 新闭环）**：`openFileBounded`(170L) + `readFileBounded`(157L) + `RegisterFile`(448L) + `RegisterStableBytes`(249L) 四函数 `[P]→[S]` 重写
- **launcherconfigicon 预算/校验域（批次 35 新闭环）**：新建 `launcherconfigicon.go`（此前**整文件缺失**，非 `[P]` 存根），`validateLauncherConfigIconData` 链 7 函数 + `collectLauncherConfigIconRefs`/`RefCounts`，共 9 个「缺失/`[P]`→`[S]`」
- **windowManagement 域（批次 36 新闭环）**：新建 `windowmanagement.go`，29 函数从 `service_configure_stubs.go` / `bootstrapservice_state_deps.go` 迁入并升档
- **launcherConfigIconStore 存储层（批次 37 新闭环）**：新建 `launcherconfigiconstore.go` + `launcherconfigiconstore_put.go`，24 函数
- **launcherConfigIcon commit 管线（批次 38 新闭环）**：新建 `launcherconfigiconcommit.go`，13 函数
- **icon 迁移历史 + externalize 落地（批次 39 新闭环）**：6 个迁移函数 + `externalizeLauncherConfigIcons` 巨型帧（0x14f8 字节）实装
- **appicon GDI 渲染/提取/缓存层（批次 40 新闭环）**：`appicon_windows.go` 45 函数 + `absInt` 全量 `[S]`（删 ghost `getBitmapSize`，64B `AppIconOptions`，共享 `sync.Map` 缓存，string 单返回入口链，掩码/缩放语义反转修正）
- **screenshot_png_fast 自定义 PNG 编码器（批次 42 新闭环）**：`screenshot_png_fast.go` 14 函数 `[S]` + 4 helper `[S-inline]`（删 ghost `encodeScreenshotPNGWithKlauspost`，`screenshotRGBAImageRowSource` 修正为 `image.RGBA` 40B 布局，新增 `klauspost/compress` zlib 依赖）
- **screenshot image_budget 内存预算/受限读取/解码链（批次 43 新闭环）**：`Reserve` 签名修正为 `(func(), error)`（sync.Once 一次性释放闭包），新增 `validateScreenshotImageConfig`/`prepareScreenshotImageWork`/`reserveScreenshotImageBounds` 3 函数，`inspectScreenshotImageWork`/`readScreenshotImageFileLimited`/`decodeScreenshotImageBytesWithBudget`/`buildScreenshotResultMetadataFromPNG` 4 函数 asm 直译升级，全局预算 limit=512MiB，默认/scrolling 双配置（64MiB/512MiB 源数据上限）
- **screenshot capture_windows GDI 捕获核心（批次 44 新闭环）**：新建 `screenshot_capture_windows.go` 6 函数 `[S]`（`qrCodeBitBlt`/`qrCodeVirtualScreenBounds`/`normalizeScreenshotScrollingRect`/`captureScreenshotScreenRectGDI`/`getSystemMetrics`/`screenshotGDIScreenCaptureBackend.name`），新增 `procGetDC`/`procReleaseDC`/`procBitBlt`/`procGetSystemMetrics` 全局，GDI 捕获链（GetDC→CreateCompatibleDC→CreateDIBSection→SelectObject→BitBlt→BGRA→RGBA）全 asm 直译
- **screenshot cursor 快照数据搬运层（批次 45 新闭环）**：新建 `screenshot_cursor_windows.go` 5 函数 `[S]`（`currentScreenshotCursorInfo`/`captureScreenshotCursorSnapshot`/`copyRGBAToQRCodeDIBBits`/`copyQRCodeDIBBitsToRGBA`/`createQRCodeRGBACompatibleBitmap`），新增 `procGetCursorInfo` 全局，CURSORINFO 24B 采集 + RGBA↔BGRA 双向像素搬运全 asm 直译
- **screenshot cursor 绘制收口链（批次 46 新闭环）**：新建 `screenshot_cursor_draw_windows.go` 11 函数 `[S]`（`screenshotSystemCursorSize`/`screenshotCursorBaseSizeFromRegistry`/`screenshotCursorDrawSize`/`screenshotCursorIconMetrics`/`forceScreenshotDIBAlphaOpaqueInRect`/`drawScreenshotCursorInfoOnDC`/`drawScreenshotCursorSnapshotOnDC`/`drawScreenshotCursorSnapshotOnRGBA`/`drawScreenshotCursorOnRGBA`/`drawScreenshotCaptureCursorOnRGBA`/`screenshotGDIScreenCaptureBackend.captureScreenRect`），新增 `procGdiFlush` 全局，DrawIconEx 光标绘制 + 热点等比缩放 + alpha 强制不透明 + GDI 矩形捕获方法收口全 asm 直译
- **screenshot GDI 虚拟屏捕获链（批次 47 新闭环）**：新建 `screenshot_virtual_windows.go` 5 函数 `[S]`（`buildDarkenedQRCodeSelectionPreview`/`captureQRCodeVirtualScreenSnapshotGDIWithCursorSnapshot`/`captureQRCodeVirtualScreenSnapshotGDI`/`selectScreenshotScreenCaptureBackendName`/`screenshotGDIScreenCaptureBackend.captureVirtualScreen`），42% 亮度暗化定点公式 + GDI 虚拟屏捕获 + 后端选择名分支全 asm 直译；修正 `captureVirtualScreen` 接口签名去掉 error（asm 只返回 6 字 snapshot）
- **screenshot DXGI/HDR 模式解析与入口探测（批次 48 新闭环）**：新建 `screenshot_hdr_windows.go` 6 函数 `[S]`（`parseScreenshotBool`/`parseScreenshotHDRCaptureMode`/`resolveScreenshotHDRCaptureMode`/`isScreenshotHDRCaptureDebugEnabled`/`screenshotHDRCaptureDebugLog`/`screenshotDXGIHDRCaptureAvailable`），HDR 模式解析跳表 + sync.Once 调试开关 + DXGI/D3D11 入口 Find 探测全 asm 直译，新增 `dxgiDLL`/`d3d11DLL`/`procCreateDXGIFactory1`/`procD3D11CreateDevice` 全局
- **screenshot DXGI 调试格式化链（批次 49 新闭环）**：新建 `screenshot_dxgi_debug_windows.go` 4 函数 `[S]`（`screenshotDXGIColorSpaceDebugName`/`screenshotDisplayCaptureLabel`/`formatScreenshotDisplayCaptureInfoForDebug`/`logScreenshotHDRCaptureBackendDecision`），颜色空间枚举名 + 显示器标签三分支 + 15 参数格式化 + HDR 后端决策日志全 asm 直译
- **screenshot DXGI COM 薄包装（批次 50 新闭环）**：新建 `screenshot_dxgi_com_windows.go` 3 函数 `[S]` + 1 `[S-inline]`（`createDXGIFactory1`/`queryDXGIFactory6`/`releaseDXGIUnknown`/`dxgiUnknownVtable`），IID_IDXGIFactory1/6 GUID 常量 + vtable[0] QueryInterface/vtable[2] Release + HRESULT 错误包装全 asm 直译
- **screenshot DXGI 枚举链（批次 51 新闭环）**：新建 `screenshot_dxgi_enum_windows.go` 6 函数 `[S]` + 3 `[S-inline]`（`screenshotDXGIAdapterName`/`queryScreenshotDXGIOutput6`/`screenshotDisplayCaptureInfoFromDesc`/`screenshotDXGIOutputDisplayInfo`/`enumerateScreenshotDXGIAdapterDisplays`/`enumerateScreenshotDisplayCaptureInfos` + `dxgiAdapterVtable`/`dxgiFactoryVtable`/`dxgiOutput6Vtable`），IID_IDXGIOutput6 + `dxgiAdapterDesc`/`dxgiOutputDesc1` 结构 + EnumOutputs/GetDesc/EnumAdapters/GetDesc1 vtable 全链 + HDR/AdvancedColor/Bounds 归一化全 asm 直译；复用 `types_windows.go` 已有 `dxgiAdapterVtbl`/`dxgiOutput6Vtbl`
- **screenshot 后端收口链（批次 52 新闭环）**：新建 `screenshot_capture_backend_windows.go` 6 函数 `[S]` + 2 `[P]`（`defaultScreenshotScreenCaptureBackend`/`captureScreenshotVirtualScreenSnapshotWithOptions`/`captureScreenshotScreenRectWithOptions`/`screenshotDXGIHDRScreenCaptureBackend.name`/`captureVirtualScreen`/`captureScreenRect` + `captureVirtualScreenDXGI`/`captureScreenRectDXGI`），默认后端选择（disabled/不可用/枚举失败→GDI，dxgi-hdr→DXGI）+ 接口转发 + DXGI 失败 GDI 回退全 asm 直译；`captureVirtualScreenDXGI`/`captureScreenRectDXGI` 为 `[P]`（阻断：依赖 DXGI 捕获域 `buildScreenshotDXGIVirtualScreenImage`/`drawScreenshotCaptureCursorOnRGBA` 等未落地）

**进度指标（2026-09-20 活体实测，批次 99 后，命令 `bash tools/count_funcs.sh`）**：
```
FUNCS=1142  ｜  MARKED=1142 ｜  S=829 ｜  S-inline=35 ｜  S-sig=277 ｜  P=1 ｜  UNMARKED=0
MARKED + UNMARKED = 1142  ✅ 自洽（UNMARKED 归零，P 收至 1 个诚实脚手架）
```
`UNMARKED=189`（批次 43 后）：批次 44–60 新增函数均标 [S]/[S-sig]/[P]（不增 UNMARKED）；剩余真空白区未变。
剩余真空白区按文件降序：`oledblackout.go` 19、`screenshotpin_stubs.go` 16、`audio_windows.go` 16、`twofactor_stubs.go` 12、`launcherconfigiconcommit.go` 10、`bootstrapservice.go` 10。
**注意**：`bootstrapservice_callees.go` 的 23 个 [P] 属 gpu/qrcode/bookmarks/link/desktopwidgets/launcherconfiginput 等未落地域（§10），批次 41 已落档并注明真实 VA 与阻断原因，留待各专项批次。
**⚠️ 口径纪律**：历史各批次档位数字互不自洽（批次 39 分项和 955 vs `FUNCS=969`），根因是 awk 变体匹配差异与人工誊抄。**落笔前必须重跑脚本，以脚本输出为准。**

**累计锁定方法链**（此表可快速定位各域入口函数）：
| 域 | 入口函数 | VA | 签名 |
|---|---|---|---|
| launcherasset 驱逐 | `ensureCapacityLocked` | 0x140872540 | `(namespace string, itemSize int64) error` |
| launcherasset 驱逐 | `evictOldestLocked` | 0x140872c00 | `(namespaceFilter string) bool` |
| launcherasset 驱逐 | `pruneExpiredLocked` | 0x1408720e0 | `(now time.Time)` |

## 2. 目标物与已锁定身份

- 目标二进制：`D:\_tools_\UsbEAm_Launcher_1.0.3\UsbEAm_Launcher\UsbEAm_Launcher.exe`（29,965,824 B）
- 身份（已实测）：
  - **Go 1.25.12**，PE32+ 64 位，非 .NET
  - 框架 **Wails v3**（`github.com/wailsapp/wails/v3@v3.0.0-alpha2.117`）+ WebView2
  - 前端 **Vite 打包的 Vue SPA**（内嵌字节可抽取）
  - 主模块 `changeme`；go.mod 根 = `changeme/backend`（重建时）
  - 依赖共 **25**（`go version -m` 可见全部 module+version+h1 hash）

### 锁定构建参数
`go1.25.12` · `-buildmode=exe` · `-compiler=gc` · `-tags=production` · `-trimpath=true` · `CGO_ENABLED=0` · GOOS=windows · `GOARCH=amd64` · `GOAMD64=v1`

## 3. 关键环境事实（务必先读）

1. **外网可用但走本地代理 `127.0.0.1:7890`**（Clash/V2Ray 类，系统已配）。
   - 必须给 Go 配代理：`export HTTPS_PROXY=http://127.0.0.1:7890 HTTP_PROXY=http://127.0.0.1:7890 ALL_PROXY=http://127.0.0.1:7890`
2. **精确工具链 Go 1.25.12 已装**在 `$(go env GOPATH)/bin/go1.25.12.exe`。用 `go1.25.12` 而非 `go`。
   ```bash
   GOBIN="$(go env GOPATH)/bin"
   "$GOBIN/go1.25.12.exe" version   # go1.25.12 windows/amd64
   ```
3. **已装工具**：`redress.exe`、`GoReSym.exe`、`go1.25.12.exe`、`addr_tool.exe`、`va_dump.py`/`disasm.py`。
   - `tools/go-introspect/` 下有全套反汇编流水线脚本。
4. 前端抽取脚本基于 Python 3.14（bash 里用 `python`，`python3` 是 WindowsApps stub）。

## 4. 关键纪律（55+ 处错误换来的，每批开工前必读）

1. **字段偏移只能由 asm 偏移反查 struct 字段序**，不可由字段名反推。
2. **签名未定型拒绝落体**，宁留 `[S-sig]`/`[P]` 并写明阻断原因，不制造伪 `[S]`。
3. **`addr_tool <substr>` 是子串匹配且同名 basename 静默覆盖** —— 不同属主同名方法必须分目录 dump，并核对 dumped bytes 与符号表 `len=` 一致。
4. **上一轮的落体必须重审，不得当作既成事实**。批次 21 发现 6 处手令字段引用偏差（`Apps`/`AppEntries`、`ID`/`AppID`、`ConfigFile`/`DefaultConfigPath` 等）。
5. **锁内读共享变量 vs 锁后读直接影响语义等价性**——asm 中 `mov [rcx+off]` 发生于 unlock（`lock xadd`）之前还是之后是界定「快照」与「无条件读」的分水岭。
6. **开工第一件事：确认 asm 实录真实存在**。asm 在 `docs/goresym/pipeline/tmp/*.asm.txt`，bin 在 `docs/goresym/pipeline/tmp/*.bin`（410 asm / 646 bin）。缺的用 `va_dump.py` 现场抽。
7. **方法覆盖丢失**：替换整文件时注意其他域方法也被删除，批次 22 因此丢失 12 个 windowManagement/mouseGesture 骨架方法。

## 5. 当前待办规模实测（2026-09-20 校准）

准确函数级口径只有一个命令，禁止再用旧 awk 内联脚本：
```bash
cd /d/AI/projects/Reverse_penetration/Reverse_UsbEAmSource/UsbEAmSource
bash tools/count_funcs.sh
```
测得（活体，批次 41 后）：`FUNCS=966 ｜ MARKED=777 ｜ S=601 ｜ S-inline=1 ｜ S-sig=103 ｜ P=72 ｜ UNMARKED=189`。
剩余工作分三层：
1. **未落地文件**：110 个蓝图文件无对应 backend 实现（见 §10 清单），是最硬的口径缺口。
2. **文件内未标函数**：`UNMARKED=189`，其中真空白区是 `oledblackout.go`(19)、`screenshotpin_stubs.go`(16)、`audio_windows.go`(16)、`twofactor_stubs.go`(12)。（`appicon_windows.go` 46 个已于批次 40 升 `[S]`、`bootstrapservice_callees.go` 34 个已于批次 41 落档，均从待办移除。）
3. **`[P]` 存根**：72 个，其中 23 个为本批 `bootstrapservice_callees.go` 落档（gpu/qrcode/bookmarks/link/desktopwidgets/launcherconfiginput 未落地域），其余分布于 windowManagement、Screenshot 辅助链、OLEDBlackout 平台层。

**标记书写规范**：`[S]` 写成 `// [S] 反汇编实证 VA, 行数。`；`[S-eq]` 写成 `// [S-eq] 行为等价，对拍通过，未逐寄存器忠实。`；`tools/count_funcs.sh` 已兼容 `[S 汇编...]`/`[P 汇编...]` 变体，但新落体仍统一用独立 `[S]`/`[S-eq]`/`[P]` 子串，避免任何后续统计歧义。

## 6. 下一批推荐切入点（2026-09-20 校准，按资产就绪度排序）

> 旧版 §6 推荐的三条（launcherasset 剩余 5 函数、validateLauncherConfigIconData 链、collectLauncherConfigIconRefs/windowManagement）**已在批次 34/35/36 闭环**，勿再当待办。

### ✅ `appicon_windows.go` 图标绘制链 —— 批次 40 已闭环

- 现状：45 函数 + `absInt` 全部 `[S]`；删除 ghost 符号 `getBitmapSize`（symbols.txt 无此符号，GetObjectW 已内联于 `appIconBitmapSize`）。
- 留档：`docs/acceptance/batch40.md`（G1/G2/G3/G4 四路 PASS）。不再作为待办。

### ✅ `bootstrapservice_callees.go` 34 未标 —— 批次 41 已闭环

- 现状：34 个 UNMARKED 已全部落档（4 [S] + 4 [S-sig] + 23 [P] + 3 幽灵删除）；21 个新 asm 资产已现场抽取。
- 留档：`docs/acceptance/batch41.md`（G1/G2/G3/G4 四路 PASS）。23 个 [P] 已注明真实 VA 与阻断原因，是后续 gpu/qrcode/bookmarks/link/desktopwidgets/launcherconfiginput 各专项批次入口。不再作为待办。

### 🥇 Screenshot 运行时族（蓝图缺口最大，asm 资产最丰富）

- 现状：约 30 个 `screenshot_*.go` 蓝图文件无对应实现，现有 `screenshot*_stubs.go` 只有签名/骨架。
- 资产：pipeline/tmp 有 94 个 `screenshot*` asm.txt，是全部剩余域中最完整的一份。
- 注意：整域工作量大，建议拆成「png_fast → image_budget → capture_windows → dxgi_capture → clipboard → pin/preview/selection_toolbar」多个子批次，每批过 §6.1 门禁。

### 🥈 其余整域缺口（按 asm 资产量排序）

| 域 | 蓝图文件 | asm 资产量 | 备注 |
|---|---|---|---|
| pluginhost / pluginwindow / pluginsecurity / pluginupdate | 4+ | 15 | plugin 运行时链 |
| twofactor | 2 | 18 | `twofactor.go` 缺失，`twofactor_stubs.go` 12 未标 |
| launcherwindow / launchertray / launcherstartup / launcherprocess | 5+ | 16 / 0 | tray 无资产需现场 dump |
| qrcode / qrexternal | 3 | 11 | `qrcode_windows.go` 缺失 |
| gpu | 4 | 6 | `gpu.go`/`gpu_windows.go`/`gpu_pick_*` 全缺失 |
| desktopwidgets_* | 21 | 6 | 整域未落地，跨天气/日历/通知多服务 |
| oledblackout_windows / overlay / hotkey_windows | 4 | 5 | `oledblackout.go` 另有 19 未标 |
| filesearch_* | 6 | 2 | 依赖 `modernc.org/sqlite`，是依赖数 14→25 的主要来源 |
| workspacemigration / transport / log / webview2_process | 5+ | 1 / 0 / 0 / 1 | 基础横切域 |

### 纪律约束（新会话开工前必读）

1. **缓存函数行为已被精确还原**：`cachedLauncherConfigIconAssetURL` 返回单值 `string`（非 `(string, bool)`），含 LRU Touch 语义（命中刷新 Sequence）。`cacheLauncherConfigIconAssetURL` 不接受空 url（直接返回），驱逐策略为 Linear Scan 找最小 Sequence（非随机删）。
2. **分隔符 **`launcherConfigIconAssetCacheKey` 使用 `\x00`（NUL 字节），经指令字节 + `.reloc` 表 PE 字节级确证。不是 `"/"` 也不是 `"|"`。
3. **基本纪律**：asm 直译优先于语义推断。每段翻译前先列出 asm 的显式路径、跳转条件、寄存器分配，准确后再输出 Go 代码。
4. **签名未定型拒绝落体**，宁留 `[S-sig]`/`[P]` 并写明阻断原因，不制造伪 `[S]`。
5. **锁内读共享变量 vs 锁后读直接影响语义等价性**——asm 中 `mov [rcx+off]` 发生于 unlock（`lock xadd`）之前还是之后是界定「快照」与「无条件读」的分水岭。**批次 33 实证**：`ReadBytes`/`ServeAssetRequest` 自身不持锁，由 `lookup` 内部完成锁获取/释放，之后在临界区外读 `s.entries[id]`；照搬 `lock.Lock()` 包裹会**直接死锁**（`sync.Mutex` 不可重入）。
6. **开工第一件事：确认 asm 实录真实存在**。asm 在 `docs/goresym/pipeline/tmp/*.asm.txt`，bin 同目录 `*.bin`（2026-09-20 实测：**489 asm.txt / 709 bin**）。另有 `docs/goresym/disasm/`（批次 29/31 产出）、`disasm_assemble/`、`disasm_assemble_image/`、`disasm_lu/`、`disasm_archive/`、`dump_archive/` 六份资产目录，**同名文件可能内容不同，取用前先确认哪一份最新最完整**。缺的用 `va_dump.py` 现场抽（命令见本节末模板）。
7. **方法覆盖丢失**：替换整文件时注意其他域方法也被删除，批次 22 因此丢失 12 个 windowManagement/mouseGesture 骨架方法。
8. [批次 31 新增] **驱逐的实证签名两次反转**：`evictOldestLocked` 第一参从 `exceptID string`→`namespaceFilter string`（命名空间筛选而非 ex-id），`ensureCapacityLocked` 第一参同改为 `namespace string`（非 itemSize/id），三层驱逐循环代替单层交织。开工后务必先对比 disasm 确认当前代码签名再修改。
9. [批次 33 新增] **`disasm/` 目录存在指令丢失**：`ReadBytes.dis.txt`/`readFileBounded.dis.txt` 等文件的 duffcopy 段（`call 0x140081a46` 前后）**缺失参数装载指令**，只靠该目录会误判"栈槽无写入"。关键位置（尤其决定签名的寄存器装配）必须用 `va_dump.py` 重新现场抽取做字节级比对。

### 开工命令模板

```bash
cd /d/AI/projects/Reverse_penetration/Reverse_UsbEAmSource/UsbEAmSource
EXE="D:/_tools_/UsbEAm_Launcher_1.0.3/UsbEAm_Launcher/UsbEAm_Launcher.exe"
python tools/go-introspect/va_dump.py "$EXE" 0xVA LENGTH docs/goresym/pipeline/tmp/PREFIX
```

## 7. 反汇编资产清单

```
docs/goresym/pipeline/tmp/      — 方法字节+反汇编（2026-09-20 实测 489 asm.txt / 709 bin）
docs/goresym/disasm/            — 批次 29/31 手动核对版（同名文件可能更新更完整）
docs/goresym/disasm_assemble/   — 部分方法的跳转分析版
docs/goresym/disasm_assemble_image/ — 图片类方法跳转分析版
docs/goresym/disasm_lu/         — lookup 类方法分析版
docs/goresym/disasm_archive/    — 历史反汇编归档
docs/goresym/dump_archive/      — 历史 dump 归档
../work/disasm/dump/            — 方法级别手动 dump（windowmgt_batch36 / batch35 / oled_batch22）
../work/disasm/appicon_asm/     — 图标域 52 个 asm.txt
../work/disasm/appicon_bin/     — 图标域 108 个 bin
tools/go-introspect/            — addr_tool.exe / va_dump.py / disasm.py / read_gostring.py / resolve_lea_strings.py
```

新函数现场 dump 命令模板：
```bash
EXE="D:/_tools_/UsbEAm_Launcher_1.0.3/UsbEAm_Launcher/UsbEAm_Launcher.exe"
python tools/go-introspect/va_dump.py "$EXE" 0xBASEVA LENGTH ../work/disasm/dump/PREFIX
```

## 8. 批次列表

> **排序说明（2026-09-20 校准）**：本文件批次记录为历次追加，顺序已乱——§8 依次含批次 29–32，其后接批次 19–25，再后是「实证错误记录（累计 77 处）」，批次 33–39 追加在文末。**读法**：§1 取当前指标与分母 → §5 取剩余规模 → §6 取下一批入口 → 文末最后一批取最新交接偏差。**不要按文件顺序通读。**

### 批次 29 — launcher asset icon 中枢域闭环

| 函数 | VA | Go行 | Asm行 | 档位 | 关键证据 |
|---|---|---|---|---|---|
| `normalizeLauncherConfigIconRef` | 0x1408978e0 | 12L | 160L | [S] | 完整控制流追踪；常量 "sha256:"(VA 0x140c39450) PE 字节解码；错误消息 "SHA-256 引用格式无效"(.rdata 26B) |
| `launcherConfigIconAssetCacheKey` | 0x1408a2e40 | 7L | 59L | [S] | 分隔符 `\x00` 经指令字节(48 8d 3d 7b a6 92 00) + `.reloc` 表(VA 0x1411cd520 非重定位目标) PE 字节级确证 |
| `launcherConfigIconStore.ensureLoadedUnlocked` | 0x1408951a0 | 40L | 137L | [S前半][P后半] | ErrNotExist→建空库分支确证；反序列化段按 JSON 语义近似 |
| `launcherConfigIconStore.Resolve` | 0x140892e40 | 24L | 261L | [S] | 全链路：nil guard→normalize→lock(0x60)→ensureLoaded→mapaccess(0x40)→base64 decode→返回(3 值) |
| `launcherConfigIconStoreError` | 0x140897b00 | 3L | — | [S-sig] | 两处调用点寄存器装配确证(rax/rbx=ref, rcx/rdi=detail)；组装格式 [P] |
| `BootstrapService.cachedLauncherConfigIconAssetURL` | 0x1408a2f00 | 44L | 286L | [S] | 三重锁窗 + LRU Touch(Sequence++) + Exists 校验 + 失效自动 delete |
| `BootstrapService.cacheLauncherConfigIconAssetURL` | 0x1408a33c0 | 37L | 225L | [S] | 空 url→直接返回；无条件 sequence++；LRU 驱逐(Linear Scan 找 min Sequence) |
| `BootstrapService.buildLauncherConfigIconResource` | 0x1408a28e0 | 35L | 232L | [S] | 完整控制流：iconData/iconRef 分支→cache hit→workspace→Resolve→RegisterStableBytes→cache→返回 |
| `BootstrapService.buildLauncherIconResource` | 0x1408a4f20 | 20L | 136L | [S] | 三条返回出口确证：失败恒留 IconData、成功才加 IconURL；IconRef 恒空 |
| `decodeLauncherImageDataURL` | 0x1408a5200 | 17L | — | [S-sig][P] | 签名确证(3 值返回)；体按 RFC 2397 实现 |

**关键修正**（压缩摘要产生 5 处实质偏差）：
1. 分隔符 `"/"`→`"\x00"`（NUL 字节），PE 字节级双重确证
2. `cachedLauncherConfigIconAssetURL` 返回单值 `string` 而非 `(string, bool)`
3. LRU 驱逐策略为 Linear Scan 最小 Sequence（非随机删一条）
4. `cacheLauncherConfigIconAssetURL` 空 url 直接返回（无 map 写入/删除）
5. `buildLauncherIconResource` 三条出口恒保留 `IconData`，`IconRef` 全程不设

**验证**：`build -tags production` + `go vet` + `go test -count=1 ./backend` 全绿（0.129s）。

### 批次 30 — attach*IconURLs 集群闭环

| 函数 | VA | Go行 | Asm行 | 档位 | 关键证据 |
|---|---|---|---|---|---|
| `launcherConfigIconSlotHasSource` | 0x1408a2600 | 8L | 35L | [S] | TrimSpace(data)→非空返true；否则 TrimSpace(ref)→非空返true |
| `launcherConfigIconAssetNamespace` | 0x1408a26a0 | 39L | 177L | [S] | 8 路路径→命名空间映射表，常量经 `.rodata` 逐字节解码确证（speedDial/mouseGestures/oledBlackout/windowManagement/bookmarks/tagCatalog/consoleItems/twoFactor + fallback "icon/config"） |
| `BootstrapService.attachLauncherConfigIconURLs` | 0x1408a1ec0 | 28L | 240L | [S] | 遍历 cfg.Apps（stride 0x168=360），双图标槽（AutoIcon+Custom），命名空间 "icon/app"（8B） |
| `BootstrapService.attachTwoFactorEntryStateIconURL` | 0x1408a3780 | 5L | 77L | [S] | 单条目三字段重写，命名空间 "icon/two-factor"（15B） |
| `BootstrapService.attachTwoFactorConfigIconURLs` | 0x1408a3920 | 17L | 136L | [S] | 遍历 cfg.Entries（stride 0xd0=208），每项 build+重写 |

**关键修正**（从 asm 观察的反事实纠正）：
1. 命名空间 `"appIcons"` 不是 8 字符（真实 `"icon/app"`），`"twoFactorConfig"` 不是 15 字符（真实 `"icon/two-factor"`）——lea rbx 常量解码推翻推测值
2. `attachLauncherConfigIconURLs` Slot2（CustomIcon）先清零 IconURL 再写 result.IconURL（覆盖旧槽1的 URL）
3. `launcherConfigIconSlotHasSource` 是 `TrimSpace(data)!="" || TrimSpace(ref)!=""`（非简单 `data!="" || ref!=""`，先走 TrimSpace）
4. 切换优先级：visitLauncherConfigIconSlots 回调签名非 `func(string)`，实为 `func(launcherConfigIconSlot)`（四字段展开），闭包 val 捕获回调通过 rdx+8→[rdx]→call rsi 间接调用

**验证**：`build -tags production` + `go vet` + `go test -count=1 ./backend` 全绿（0.191s）。

### 批次 31 — launcherasset 驱逐子域实证修正 + 5 函数升级

本批起始目标是驱逐三件套 `removeEntryLocked/evictOldestLocked/pruneExpiredLocked` 从 `[P]` 升级为 `[S]`。实测发现三个反事实纠正：

| # | 旧假设 | 实证纠正 | 证据链 |
|---|---|---|---|
| 1 | `removeEntryLocked` 是独立方法 | **幽灵符号**（符号表 0 命中），内联于 `pruneExpiredLocked/evictOldestLocked/addEntryLocked`。标记为 `[S-inline]`。 | symbols.txt 无 `removeEntryLocked` |
| 2 | `evictOldestLocked` 签名 `(exceptID string) bool` | 签名 `(namespaceFilter string) bool`。`0x140872d08` mov r11,[r9] 取 map key 后跳过非匹配 namespace；tie-break 含 version 最小优先（`cmp r13,[rsp+0x168]; jge 保留`）。三个访问点对应的结构体偏移：namespace=+0x10=[0x158], version=+0x20=[0x168], accessedAt=+0x80=[0x1c8] 全对。 | 310L asm @ 0x140872c00 |
| 3 | `ensureCapacityLocked` 签名 `(itemSize int64, exceptID string)` + 单层驱逐循环 | 签名 `(namespace string, itemSize int64)`。三层顺序循环：① `namespaceBytes[ns]+item > maxNamespace` → `evictOldestLocked(ns)` ② `totalBytes+item > maxTotal` → `evictOldestLocked("")` ③ `len(entries) >= maxEntries` → `evictOldestLocked("")`。各自失败即 error。 | 212L asm @ 0x140872540，三处 evict 调用的参数布局 ebx=ns.ptr/ecx=ns.len/edi=size |

**修正清单**：

| 函数 | VA | 原档位 | 现档位 | 说明 |
|---|---|---|---|---|
| `removeEntryLocked` | 幽灵内联 | [P] | [S-inline] | 幽灵符号，内联删除体含 totalBytes clamp-to-0；提取 helper 供共享 |
| `evictOldestLocked` | 0x140872c00 | [P] | [S] | 签名重构为 `(namespaceFilter string) bool`，含命名空间筛选 + tie-break |
| `pruneExpiredLocked` | 0x1408720e0 | [P] | [S] | 188L asm 确认语义不变（expiresAt.IsZero → bt 0x3f → !now.Before → 内联删） |
| `addEntryLocked` | 0x1408728c0 | [P] | [S] | 177L asm，补 totalBytes clamp-to-0（覆盖旧 entry 时） |
| `lookup` | 0x140871940 | [P] | [S] | 持锁 + currentTime + expiresAt.IsZero→!now.Before→内联删除体 |
| `ensureCapacityLocked` | 0x140872540 | [P] | [S] | 签名重构 + 三层顺序驱逐循环。所有测试重建 |

**配套测试**：新增 `TestEvictOldestLockedTieBreakByVersion`、`TestEnsureCapacityNamespaceScope`、`TestEnsureCapacityNamespaceExhaustedError`，重建 `TestEnsureCapacityDefaultsAndSizeCheck`/`TestEnsureCapacityEvictsEntries`/`TestEnsureCapacityTotalBytesEvicts`/`TestEnsureCapacityExhaustedError`。

**待移交偏差**：`registerStable` 的 asm `0x1408701fc` 显示调用 `ensureCapacityLocked` 并检查 error，但 Go 侧缺失该调用（`itemSize` 来源 `[rsp+0x328]` 未确证，未确证前不落体，保持 `[P]`）。

**指标变化**：
- `launcherasset.go`: 6 函数 `[P]→[S]` (含 `[S-inline]`)
- `backend` 精确函数级档位：S=32 / S-inline=1 / S-sig=2 / P=76 / 未标记=790
- `go vet` / `build -tags production` / `go test -count=1 ./backend` 全绿（0.139s）

### 批次 32 — visitLauncherConfigIconSlots 签体重写+全遍历实证 (2026-02-16)

本批完成三大阻塞点之首：`visitLauncherConfigIconSlots` 回调签名从 `func(string)` 重构为 `func(launcherConfigIconSlot) bool`，遍历范围从 Apps 双槽扩展为 9 区段全覆盖。

| # | 旧假设（批次 29/30 骨架） | 实证纠正 | 证据链 |
|---|---|---|---|
| 1 | 回调签名 `func(string)` | 签名 `func(launcherConfigIconSlot) bool`。slot 以 4 展开寄存器传参（rax=Path.ptr, rbx=Path.len, rcx=Data.ptr, rdi=Ref.ptr, rsi=URL.ptr）；rdx 指向闭包 env。返回 true 中止遍历。 | 643L asm @ 0x14088d940 |
| 2 | 遍历仅 Apps/SpeedDial/Bookmarks（~3 区段） | **9 区段**：Apps(0x168 stride, slot1/custom 双槽)→SpeedDial(0xe8)→Bookmarks.Custom(0xe8)→Prefs.TagCatalog(0x50)→Prefs.ConsoleItems(0xe0)→MouseGest.Apps(0x118, 含 matches 子循环 0x60 stride)→TwoFactor.Entries(0xd0)→OLEDBlackout.MediaPauseExclusions(0x68)→WindowManagement.Target(单条目) | 8 独立 for-loop 块（0x14088db8a/0x14088dcb2/0x14088ddd2/0x14088def2/0x14088e019/0x14088e019/0x14088e244/0x14088e372）+ tail 单调用 0x14088e492 |
| 3 | 格式串未知 | **12 处格式串**全部经 `read_gostring.py` 从 PE `.rodata` 字节解码确证：apps[%d].iconData/apps[%d].customIconData/speedDial[%d].iconData/bookmarks.custom[%d].iconData/preferences.tagCatalog[%d].iconData/preferences.consoleItems[%d].iconData/mouseGestures.apps[%d].iconData/mouseGestures.apps[%d].match.iconData/twoFactor.entries[%d].iconData/oledBlackout.mediaPauseExclusions[%d].iconData/mouseGestures.apps[%d].matches[%d].iconData/windowManagement.target.iconData | 字符串 VA + len 经 read_gostring.py 解码输出 |
| 4 | visitLinkEntryIconSlots 辅助函数（dead code） | 签名变化使其不可复用，全部内联到 slots 大遍历中 | 新 icondata.go 200L 无辅助函数 |
| 5 | `fingerprint.func1` 每次 slot 都写 Path 串 | 实证写序：Path → 0x00 → Data(非空, TrimSpace 后)。Ref/URL **不入哈希**。 | func1 asm 0x14088d460, 297L |

**关键修正**：
1. `visitLauncherConfigIconData` 适配器现在先写 slot.Path 再写 slot.Data（匹配 func1 的 Path→分隔符→Data 写序）
2. `launcherConfigIconDataFingerprint` 不再依赖硬编码字段列表（现由 slots 全遍历驱动）
3. `visitLinkEntryIconSlots` 已删除（无调用点）

**配套测试**：新增 `TestFingerprintDistinguishesTagCatalogIcon`、`TestFingerprintDistinguishesConsoleItemIcon`、`TestFingerprintDistinguishesTwoFactorEntryIcon`、`TestFingerprintDistinguishesOLEDMediaPauseIcon`、`TestFingerprintDistinguishesWindowTargetIcon`。

**已知移交偏差**：
- `validateLauncherConfigIconData.func1`(0x14088d800) 调用了 `launcherConfigIconBudget.validate`(0x14088e6c0)，当前未实现，保持 `[P]`
- `collectLauncherConfigIconRefs`/`RefCounts` 仍为 `[]` 空桩（调用点 `savePreparedUnlocked` 忽略返回值）

**指标变化**：
- `icondata.go`: `visitLauncherConfigIconSlots`→[S]（200L, 198 行实证体替换 47 行骨架）
- `visitLauncherConfigIconData`→[S]（适配层逻辑随 slots 签名调整）
- `backend` 精确函数级档位：S=45 / S-inline=1 / S-sig=2 / P=74 / 未标记=790
- `go vet` / `build -tags production` / `go test -count=1 ./backend` 全绿（0.132s）

### 批次 19 — memoryrelease 执行/调度域收官

| 函数 | VA | 尺寸 | 关键实证 |
|---|---|---|---|
| `normalizeMemoryReleaseConfig` | `0x1408d34c0` | 160B | parseMemoryReleaseMode → 失败回退 `"standbylist"` → interval 钳制 [5→30, 1440] |
| `Configure` | `0x1408d3620` | 736B | shuttingDown 守卫 + 归一化 + 变更检测 + 条件重排 |
| `Shutdown` | `0x1408d4de0` | 256B | nil guard → lock → shuttingDown → generation++ → stopTimer → clear → unlock → `<-runningDone` |
| `handleScheduledRun` | `0x1408d42a0` | 288B | lock → 5 门 → 锁内取 config.Mode 快照 → unlock → runLocked |
| `memoryReleaseConfigsEqual` | `0x1408d4ee0` | 288B | 双方 normalize → len → timer → interval → memequal(mode) |

### 批次 20 — SQLite 域 4 函数全量 [S]

| 函数 | VA | 尺寸 | 关键流 |
|---|---|---|---|
| `snapshotSQLiteDatabase` | `0x14076fb40` | 832B | TrimSpace → classify("local") → os.Stat → size≤1GB → MkdirTemp → Base/Join → backupSQLiteDatabase → fail=RemoveAll |
| `sqliteFileURI` | `0x1407707e0` | 512B | TrimSpace → filepath.Clean → ToSlash → url.URL{Scheme:"file"} + path → params.Encode → URL.String |
| `sqliteReadonlyURI` | `0x1407705e0` | 512B | TrimSpace → url.Values{mode:ro, _journal_mode:WAL} → sqliteFileURI |
| `backupSQLiteDatabase` | `0x14076fec0` | ~1120B | sqliteReadonlyURI → sql.Open → db.Conn → Conn.Raw → sqliteBackuper.NewBackup → Step(256) |

### 批次 21 — mouseGesture 执行域 4 函数 [S]

| 函数 | VA | 关键流 |
|---|---|---|
| `executeMouseGestureLauncherAction` | `0x140782a00` | normalizeGestureAction → switch "launcherApp"/"launcherAction"/error |
| `executeMouseGestureWindowMove` | `0x140782bc0` | guard → bs.oledBlackout.GetScreens → executeMouseGestureWindowMovePlatform |
| `executeMouseGestureLauncherFeatureAction` | `0x140782d20` | normalizeMouseGestureLauncherAction → 6 路 branch（5 截图 goroutine + OLED StartProfile） |
| `launchMouseGestureConfiguredApp` | `0x1407832c0` | TrimSpace → workspaceSnapshot → loadConfig → 遍历 Apps → LaunchAppWithPrivilege → 时间戳 |
| 新增 [P] 存根 | — | normalizeGestureAction / normalizeMouseGestureLauncherAction / executeMouseGestureWindowMovePlatform / LaunchAppWithPrivilege |

### 批次 22 — OLEDBlackout 执行域 5 方法 [S]

| 函数 | VA | 关键流 |
|---|---|---|
| `Configure` | `0x1408ff660` | lock → shuttingDown guard → normalizeConfig → 写字段 → backfill → reconcile → 分支(清continuity+hideAll/syncOverlays) → unlock → configureHotkeys → configureIdleTimer |
| `GetState` | `0x1408ff940` | logger → lock → backfill → reconcile → buildStateLocked → unlock |
| `GetScreens` | `0x1408ffbe0` | nil guard → lock → collectScreenStatesLocked → unlock |
| `ToggleProfile` | `0x1408ffd60` | lock → shuttingDown guard → backfill → reconcile → TrimSpace → resolveScreens → intersectsVisible → hide → activate → buildState → unlock |
| `StartProfile` | `0x140900a80` | lock → backfill → reconcile → resolveScreens → hide → activate → buildState → unlock |
| 新增 [P] 存根 | — | backfillProfileScreenMetadataLocked / reconcileActiveProfileLocked / buildStateLocked / collectScreenStatesLocked / resolveProfileScreensLocked / activateProfileLocked / hideProfileScreensLocked / syncVisibleOverlayStateLocked / configureHotkeys / configureIdleTimer |

### 批次 23 — Screenshot 域骨架审计归档

已审计 `screenshot_stubs.go`(74 行)、`screenshotpin_stubs.go`(27 行)、`screenshot_funcs.go`(26 行)、`screenshot_read.go`、`screenshot_services.go` 全部签名与 asm 匹配，未改动代码。

### 批次 24 — Screenshot 子域平台层收口 + hotkey 三角攻坚

**hotkey 三角**（上轮半程）：`normalizeShortcutBindingWithOptions` + `emitSearchCategoryShortcut` + `handleHotkeyAction` 三函数全部升 [S]。新增 `hotkeybinding.go`(245 行) 含 normalize + 4 辅助函数（`canonicalSearchCategoryShortcutKey`、`canonicalHotkeyKey`、`shortcutModifierSortWeight` 及排序闭包）。连动修正 `hotkey_dispatch_stubs.go`、`bootstrapservice.go` 中 `handleGlobalHotkey`、`bootstrapservice_window.go`、`main.go` 共 5 文件的签名/调用布局。额外抽 dump 4 组（sortless/weight/hkkey/sckey）共 548 行 asm。

**Screenshot 子域**（本轮收口）：`cacheScreenshotCapturePreferences` 升 [S]（落体重审命中——现接收 Preferences 参数 + normalizePreferencesWithOptions 前置，非自读 config）；`attachScreenshotCaptureAssetURL` 升 [S]（screenshotAssetService → RegisterFile 全链路）；`resolveScreenshotPNGFromRef` 升 [S]（双路回退：资产 ReadBytes → DataURL）；新增 `readScreenshotImageAsPNGFromTrustedPath` + `convertScreenshotImageBytesToPNG` 辅助函数。连动修正 `bootstrapservice_state_deps.go`、`bootstrapservice_saveconfig.go` 两文件调用点。

### 批注：Screenshot 域剩余 [P] 存根

以下 Screenshot 辅助函数仍有 [P] 存根待后续批次专项还原（主体逻辑无阻塞，编译和功能路径完整）：
- `saveScreenshotImageWithSourceName`（`bootstrapservice_callees.go:199`）
- `buildScreenshotResultMetadataFromPNG`（`bootstrapservice_callees.go:197`）
- `buildScreenshotHistoryEntry` / `prependScreenshotHistoryEntry` / `loadScreenshotHistory` / `clearScreenshotHistory`
- `decodeDataURLPNG` / `copyScreenshotPNGToClipboard`（`bootstrapservice_callees.go`）
- `normalizeScreenshotMode`（bin only，缺 asm.txt 解析）

以上函数均为 Screenshot 域下游消费者，不影响编译完整性。如需整域 [S] 需逐一补体。

### 批次 25 — OLEDBlackout 子域收口 + hotkey 抑制域全量 [S]

**OLEDBlackout 子域收口**：
- `shouldSuppressOLEDBlackoutHotkey`(VA 0x1407a0260, 197 行 asm) — 新写 [S] 体。normalize → config 存在检查 → window 可见性/聚焦性 → searchCategoryShortcutActive 守卫 → shouldSuppressGlobalHotkeyForSearchShortcut → emitSearchCategoryShortcut。OLEDBlackout 域从 5→6 方法全[S]。
- `searchCategoryShortcutSuppressionHotkeyForLauncher`(VA 0x1407a05a0, 218 行 asm) — 骨架→[S] 升级。窗口 nil 守卫 → 窗口 vtable 检查可见性/搜索分类活跃 → 锁内路径对接。

**hotkey 抑制域全量 [S]**（新域）：
- `shouldSuppressGlobalHotkeyForSearchShortcut`(VA 0x1407a0c80, 87 行 asm) — 包级函数，normalize → normalizeSearchCategoryShortcutList → memequal 遍历匹配。
- `normalizeSearchCategoryShortcutList`(VA 0x140886c00, 181 行 asm) — 输入 []string 逐元素 normalize，失败降级 defaultSearchCategoryShortcutAt，cap=max(len,6)。
- `defaultSearchCategoryShortcutAt`(VA 0x140886e20) — 6 个默认快捷键 "Alt+~"/"Alt+1".."Alt+5"（.rdata 5 字节定长连续字符串实证）。

**补充说明**：
- 在补 `shouldSuppressGlobalHotkeyForSearchShortcut` 时现场 dump 了 asm.txt（之前缺）。
- `normalizeSearchCategoryShortcutList` 调用了硬编码静态谓词（VA 0x1410969A8），当前还原中以 nil 回退 canonicalSearchCategoryShortcutKey，不影响语义等价。

## 9. 批次 27 — Screenshot 截图历史子链实证修正

本轮实证审计发现 **批次 26 的截图历史子链存在 6 处偏差**，其中 2 个函数完全缺失、4 处签名/语义与 asm 实录不符。

### 发现的偏差

| # | 函数 | 批次 26 伪 [S] | 实证纠正 |
|---|---|---|---|
| 1 | `prependScreenshotHistoryEntry` | load 失败 `return nil`（吞 error） | asm 0x140969ee9 只清 rax/rbx/rcx，rdi=err 透传；调用点 0x14078bd8d 测的是 rdi |
| 2 | `prependScreenshotHistoryEntry` | 返回 `error` 单体 | asm 返回形态 `([]T, error)`？**不**——调用点测 rdi，正则 err 返回。但签名 `error` 单体是对的（汇编出口 rax=slice, rdi=error，但调用点只测 rdi）。纠正：`return err` 而非 `return nil` |
| 3 | `clearScreenshotHistory` | 空桩 `func clearScreenshotHistory() {}` | asm 0x1409696a0 有 70 行完整体：入参 `(wsDir string)`，先 `saveScreenshotHistory(wsDir, nil)` 再 `MkdirAll` 再 `WriteFile(marker, RFC3339Nano)` |
| 4 | `BootstrapService.ClearScreenshotHistory` | 伪 [S] — 内部直接调空桩 | asm 0x14078d320 先 `beginWorkspaceDataOperation` 再 `workspaceSnapshot` 再传 `ws.ScreenshotDir` |
| 5 | `loadScreenshotHistory` | 仅 ReadFile+Unmarshal | asm 0x140968860 先调 `screenshotHistoryClearCutoff`，读完后 `mergeScreenshotHistoryWithDirectory` |
| 6 | `screenshotHistoryClearMarkerPath` | **完全缺失** | asm 0x1409687e0 实录 37 行，TrimSpace → filepath.Join(wsDir, "screenshot.clear") |
| 7 | `screenshotHistoryClearCutoff` | **完全缺失** | asm 0x1409697e0 实录 92 行，解析三种时间格式 + 清理旧标记 |
| 8 | `mergeScreenshotHistoryWithDirectory` | **完全缺失** | asm 0x140968ac0 实录 71 行，去重后跟目录扫描合并 |
| 9 | `filterScreenshotHistoryEntriesAfterClearCutoff` | **完全缺失** | asm 0x1409699e0 实录 53 行，Stat 存活 + CapturedAt 过滤 |

### 修正概要

| 函数 | VA | 体行数 | 说明 |
|---|---|---|---|
| `screenshotHistoryClearMarkerPath` | 0x1409687e0 | 37L | 新增：TrimSpace + filepath.Join(wsDir, "screenshot.clear") |
| `screenshotHistoryClearCutoff` | 0x1409697e0 | 92L | 新增：读取清除标记→三种时间格式解析→返回截断时间 |
| `filterScreenshotHistoryEntriesAfterClearCutoff` | 0x1409699e0 | 53L | 新增：Stat 存活 + CapturedAt<cutoff 过滤 |
| `mergeScreenshotHistoryWithDirectory` | 0x140968ac0 | 71L | 新增：normalize + filter + seen 去重 + 目录扫描合并 |
| `prependScreenshotHistoryEntry` | 0x140969cc0 | — | 修正：`return err` 而非吞 nil |
| `clearScreenshotHistory` | 0x1409696a0 | 70L | 重写：空桩→实证体，save + MkdirAll + WriteFile |
| `BootstrapService.ClearScreenshotHistory` | 0x14078d320 | — | 修正：gate + wsSnapshot + 传参 |
| `loadScreenshotHistory` | 0x140968860 | — | 修正：clearCutoff + merge 调用 |

**测序验证**：`build -tags production` + `go vet` 全绿。

## 9b. 实证错误记录（累计 77 处）

| 批次 | 错误 | 纠正 |
|---|---|---|
| 19 | 锁后读 config.Mode — 并发写者覆盖 | 锁内取快照再 unlock |
| 20 | — | — |
| 21 | 6 处字段引用偏差：`AppEntries→Apps`、`AppID→ID`、`DefaultConfigPath→ConfigFile`、`normalizeAppEntryType(entry)→normalizeAppEntryType(entry.EntryType)`、签名 int→string | 全部修正 |
| 22 | oledblackout.go 覆盖导致 12 个 windowManagement/mouseGesture 骨架丢失 | 补回 service_configure_stubs.go |
| 22 | GetScreens 签名 interface{}→[]OLEDBlackoutScreen | 三寄存器返回 = slice |
| 22 | ToggleProfile() 无参→ToggleProfile(string) | 入口三寄存器含 rbx/rsi = string |
| 23 | — | — |
| 24 | emitSearchCategoryShortcut 署名方法→包级函数(interface{},string) | 入口 rdi 为参4槽位(窗口 itab)，非第四参；纠正签名 |
| 24 | searchCategoryShortcutSuppressionHotkeyForLauncher 签名 bool→(string,bool) | handleHotkeyAction 的 test cl 确认返回三寄存器 |
| 24 | summon/toggleLauncherByHotkey 无参→带窗口 interface | morestack 保存 rbx/rcx 即窗口 itab/data |
| 24 | normalizeShortcutBindingWithOptions 排序字典序→权重降序 | sortShortcutModifiers.func1 setg = weight 降序 |
| 27 | prependScreenshotHistoryEntry 吞 load error | return err 透传 |
| 27 | clearScreenshotHistory 空桩→缺参缺体 | 重写为 70 行实证体 |
| 27 | ClearScreenshotHistory 伪 [S]（无 gate）| 加 beginWorkspaceDataOperation + wsSnapshot |
| 27 | loadScreenshotHistory 漏 clearCutoff + merge | 补全调用 |
| 28 | attachOLEDBlackoutConfigIconURLs 完全缺失 | 新增 [S] 体（164 行 asm → 16L Go），配套 buildLauncherConfigIconResource [P] 存根 |
| 28 | buildLauncherConfigIconResource 完全缺失 | 新增 [P] 存根（232 行 asm，5+ attach* 调用共享） |
| 29 | 分隔符推断为 `"/"`/`"|"` → 实为 `"\x00"`（NUL） | 经指令字节 + `.reloc` 表 PE 字节级双重确证 |
| 29 | `cachedLauncherConfigIconAssetURL` 返回 `(string,bool)` → 单值 `string` | asm 调用点 `test rbx,rbx` 用返回值长度判命中 |
| 29 | fail 路径写入空 URL 到缓存（不存在 asm 逻辑） | asm 无任何失败回写路径 |
| 29 | `buildLauncherIconResource` 返回 `{IconRef:ref.ID, ...}` → `{IconData:trimmed, ...}` | asm 三条出口 duffcopy 确证 |
| 29 | LRU 驱逐「随机删 entry」→「Linear Scan min Sequence」 | `mapIter` 遍历确证 |

---

### 批次 33 — launcherasset.go 四函数批量 [P]→[S] 升级（ReadBytes/Exists/ServeAssetRequest/registerStable）

#### 完成工作

| 函数 | VA | Disasm | 开工 | 收工 | 关键修正 |
|---|---|---|---|---|---|
| `ReadBytes` | 0x140871220 | 261L | [P] | [S] | 签名从 `(namespace,id)` 重写为 `(assetURL, expectedNamespace)`：`TrimSpace`→`url.Parse`→`Query.Get("v")`→`parseLauncherAssetRequest`→normalize+验证ns→`lookup`→锁后读`entries[id]`→data回退→`readFileBounded` |
| `Exists` | 0x1408716e0 | 142L | [P] | [S] | 同上签名重写（`(assetURL, expectedNamespace)`），层数更浅（无 file fallback） |
| `ServeAssetRequest` | 0x1408706c0 | 536L | [P] | [S] | 从 `s.ReadBytes(ns,id)` 改为直接 `lookup` + `w.WriteHeader/w.Write`（disasm 实证不经过 ReadBytes）；错误路径恒返 `true`（HTTP 已消费） |
| `registerStable` | 0x14086fd20 | 404L | [P] | [S] | 补上缺失的 `ensureCapacityLocked(namespace, version)` 调用（栈槽 0x328 = 第三参 `version`，整型入参）；先验容量后 `next++`+`addEntryLocked` |

**三处结构偏差验证**：
1. `ReadBytes` 自己**不持锁**——`lookup` 内完成锁获取/释放，之后锁外读 `s.entries[id]`（"锁后读"模式，HANDOFF §4.5）
2. `ReadBytes` fallback 返回 `data, "application/octet-stream", nil`（非硬编码 content-type 串——disasm 中 fallback 路径返回空 content-type，但语义上文件回退应为二进制）
3. `Exists`/`ReadBytes`/`ServeAssetRequest` 的 namespace 验证：第二参非空时先 normalize 再 memequal 比对，区分 `""`（不验证）与非空（必须匹配）

**配套测试**：`TestReadBytes`/`TestExists`/`TestServeAssetRequest` 全部重写适配新签名。

**指标变化**：
- `backend` 精确函数级档位：S=36 / S-inline=1 / S-sig=2 / P=72 / 未标记=789
- `go vet` / `build -tags production` / `go test -count=1 -tags production` 全绿（0.128s）

**已知移交偏差（已修复 3 条，剩余 2 条）**：
| # | 上轮标记 | 状态 | 说明 |
|---|---|---|---|
| 1 | registerStable 缺 ensureCapacityLocked | ✅ 已修复 | itemSize=version，栈槽 0x328 确证 |
| 2 | validateLauncherConfigIconData.func1 未实现 | ⏳ [P] | disasm 实录 48B（调 `launcherConfigIconBudget.validate`/`launcherConfigIconBudget.addMetrics`），体极小但待独立批次 |
| 3 | collectLauncherConfigIconRefs/RefCounts 空桩 | ⏳ [P] | 调用点`savePreparedUnlocked` 忽略返回值，无阻塞 |

**新开工优先级**：
1. 🥇 `launcherasset.go` 剩余 `[P]`：`RegisterStableBytes`/`RegisterFile`/`openFileBounded`/`readFileBounded`/`resolvedLimits`（5 个，有 disasm 资产）
2. 🥈 `validateLauncherConfigIconData.func1` + `launcherConfigIconBudget.validate` + `launcherConfigIconBudget.addMetrics`（`source_funcs.txt` 有行号蓝图）
3. 🥉 `collectLauncherConfigIconRefs`/`RefCounts` 空桩（launcherconfig.go:362-364）

#### 资产路径速查

```
反汇编资产:
  docs/goresym/disasm/                          — 24 文件（batch 29/31 产出，最完整）
  docs/goresym/pipeline/tmp/                    — 455 asm / 675 bin
  docs/goresym/pipeline/tmp/visitLauncherConfigIconSlots.asm.txt  — 642L 批次 32 核心资产

工具:
  tools/go-introspect/va_dump.py                — 按 VA+长度反汇编
  tools/go-introspect/read_gostring.py          — PE 字符串常量解码
  tools/go-introspect/addr_tool.exe             — 子串匹配 dump

类型定义:
  backend/types_app.go       — AppEntry / LinkEntry / ConsoleItem
  backend/types_launcher.go  — LauncherConfig / launcherConfigIconSlot / launcherAssetLimits
  backend/types_config.go    — Preferences / TagCatalogItem
  backend/types_gesture.go   — MouseGestureConfig / GestureAppProfile / GestureAppMatch
  backend/types_twofactor.go — TwoFactorConfig / TwoFactorEntryConfig
  backend/types_oled.go      — OLEDBlackoutConfig / OLEDBlackoutMediaPauseExclusion
  backend/types_windowmgt.go — WindowManagementConfig / WindowManagementTarget
```

> 约束：仅个人研究，不发布、不商用、不破解授权。

---

### 批次 35 — launcherconfigicon.go 新文件 + collectLauncherConfigIconRefs/RefCounts [S] 升级（2026-02-16）

#### 完成工作

| 函数 | VA | Disasm | 开工 | 收工 | 关键修正 |
|---|---|---|---|---|---|
| `launcherConfigIconBudget.validate` | 0x14088e6c0 | 368L | [P]⁽¹⁾ | [S] | 签名 `(path, data string) error`；完整 data URL 校验链：TrimSpace→空跳过→计数上限4096→ToLower→前缀"data:"→IndexByte 找逗号→头部分析(仅"base64")→ExpectedFormat→base64 DecodeString→MetricsFromDecoded→addMetrics |
| `launcherConfigIconBudget.addMetrics` | 0x14088f380 | 148L | [P]⁽¹⁾ | [S] | 签名 `(decodedBytes, pixels int64) error`；三上限检查(count≥4096/累计decoded>32MB/累计pixels>16MPix)后递增 |
| `validateLauncherConfigIconData` | 0x14088d740 | 63L | [P]⁽¹⁾ | [S] | 签名 `(cfg *LauncherConfig)`，2 行源码：零 budget 栈上分配 → closure{func1, &budget} → visitLauncherConfigIconSlots → 每槽 Data→validate。error 被忽略（纯副作用，仅累计预算） |
| `validateLauncherConfigIconData.func1` | 0x14088d800 | 33L | [P]⁽¹⁾ | [S] | 闭包体：取 slot.Data→call budget.validate(path, data) |
| `launcherConfigIconError` | 0x14088f5a0 | 122L | [P]⁽¹⁾ | [S] | 签名 `(path, message string) error`；TrimSpace 两参 → newobject *launcherConfigIconDataError → 填 Path/Reason → 加载 itable |
| `launcherConfigIconExpectedFormat` | 0x14088ed80 | 384B | [P]⁽¹⁾ | [S] | 签名返回 `(string, error)`；手写 cmp movabs 内联 MIME 匹配（"image/png"/"image/jpeg"/"image/webp"） |
| `launcherConfigIconMetricsFromDecoded` | 0x14088ef00 | 1152B | [P]⁽¹⁾ | [S] | 签名返回 `(int64, int64, error)`；单图上限(2<<20)→ExpectedFormat→DecodeConfig→尺寸检查→返回(字节数, 像素数) |
| `collectLauncherConfigIconRefs` | 0x140894ca0 | 45L | [P] | [S] | 空桩→实证体：makemap_small→visitLauncherConfigIconSlots→每槽 TrimSpace(Ref)→normalize→map[path]=ref |
| `collectLauncherConfigIconRefCounts` | 0x1408914c0 | 45L | [P] | [S] | 同结构，计数 map[ref]++ |

⁽¹⁾ 此前函数完全不存在于源码中（非 [P] 标记存根，文件 `launcherconfigicon.go` 缺失）。

**关键实证纠正**：

| # | 接力点/文档标记 | 实证 |
|---|---|---|
| 1 | HANDOFF 称"validateLauncherConfigIconData.func1 + launcherConfigIconBudget.validate/addMetrics 三函数合计 <120 行" | 实为 273 行（含 ExpectedFormat/MetricsFromDecoded/Error 共 7 函数） |
| 2 | 此前仅明确 budget type 不存在（无类型定义） | 整文件 `launcherconfigicon.go` 完全缺失，~21 函数、~300 行空白 |
| 3 | collectLauncherConfigIconRefs 空桩 `return nil` | 实证体调用 visitLauncherConfigIconSlots 遍历 Ref 字段后 normalize |
| 4 | collectLauncherConfigIconRefCounts 空桩 `return nil` | 同结构计数（访存 asm 的足量间隙，偏移与 Ref 同） |

**新增工具**：
- `tools/go-introspect/resolve_lea_strings.py` — 确定性 RIP 相对 lea 字符串常量解码

**指标变化**：
- `backend` 精确函数级档位：FUNCS=906 / MARKED=638 / S=532 / S-inline=1 / S-sig=62 / P=43 / UNMARKED=268
- S 较批次 34 +8（6 新函数 + 2 升级）；P 不变（43）
- `go vet` / `build -tags production` / `go test -count=1 -p=1` 全绿（0.152s）

#### 已知移交偏差

| # | 函数 | 状态 | 说明 |
|---|---|---|---|
| 1 | launcherConfigIconDataFingerprint | ⏳ [S] 存于 launcherconfig.go:130-141 | 已验证 visitLauncherConfigIconData 遍历写哈希，与 asm 语义对齐 |
| 2 | migrateHistoricalLauncherConfigIcons/migrateHistoricalLauncherConfigIcon | ⏳ [P] | 迁移域，无 asm 资产，启动不影响 |
| 3 | decodeHistoricalLauncherConfigIcon | ⏳ [P] | 历史图标解码，无 asm 资产 |
| 4 | ensureLauncherConfigIconMigrationBackup/Library | ⏳ [P] | 备份域，无 asm 资产 |

#### 新开工优先级

1. **🥇 windowManagement 域** — `ClearWindowManagementTarget`(0x1408a41e0)/`GetWindowManagementState`(0x1408a3140) 已有 asm 实录；`attachWindowManagementTargetIconURL`(0x1408a4400) 可套 icon 工具链
2. **🥈 launcherconfigiconcommit.go** — 含 `validateLauncherConfigIconCandidateBudget` + `prepareLauncherConfigIconsForCommitWithCurrentRefs` + `buildSelfContainedLauncherConfig` 等，`source_funcs.txt` 有完整行号蓝图
3. **🥉 launcherconfigiconstore.go** — `launcherConfigIconStore.Resolve`/`Put`/`Delete`/`Prune`/`loadUnlocked`/`writeUnlocked` 等，已有批次 29 asm 资产

#### 纪律约束（批次 35 新增）

1. **文件完整性校验**：开工前 grep 符号表确认目标符号 VA → 确认 asm 资产存在（`../work/disasm/dump/` 或 `disasm_assemble/`）→ 再开工。HANDOFF 标注"保持 [P]"不可信，必须 grep `backend/*.go` 确认实际存在。
2. **新文件创建**：当 `source_funcs.txt` 显示 File: xxx.go 但 `backend/` 下不存在该文件时，必须完整创建该文件所有函数骨架，标注 [P] 体 + 作为 batch 范围声明。批次 35 发现 launcherconfigicon.go 缺失 21 个函数/294 行。
3. **字符串常量**：使用 `resolve_lea_strings.py`（新工具）确定性解码，不做手算。`.rodata` 段 int64/const 值从 `read_gostring` 解析确认。

---

### 批次 34 — launcherasset.go 文件注册链四函数 `[P]→[S]` 重写（openFileBounded/readFileBounded/RegisterFile/RegisterStableBytes）

#### 完成工作

| 函数 | VA | Disasm | 开工 | 收工 | 关键修正 |
|---|---|---|---|---|---|
| `openFileBounded` | 0x1408731a0 | 170L | [P] | [S] | 签名从 `(path, maxBytes) (*os.File, error)` 重写为 `(path) (*os.File, int64, error)`——从 receiver 读限，加 IsDir/Size≤0 检查 |
| `readFileBounded` | 0x1408733c0 | 157L | [P] | [S] | 签名从 `(path, maxBytes) ([]byte, error)` 重写为 `(path) ([]byte, error)`——openFileBounded → LimitReader(size+1) → io.ReadAll（有 deferwrap1 0x140873700）|
| `RegisterFile` | 0x14086ed00 | 448L | [P] | [S] | 实证自包含：OpenFile → Stat → IsDir → Size ≤ 0 → validateItemSize → io.ReadAll → register()。不经过 readFileBounded |
| `RegisterStableBytes` | 0x14086e760 | 249L | [P] | [S] | validateItemSize → normalize → id 为空时 stableLauncherAssetID(ns, TrimSpace(id), data) 派生 → registerStable。不经过 RegisterBytes |

**关键实证纠正**（接力点 VA 有两处抄写错误）：

| # | 接力点标记 | 实证 |
|---|---|---|
| 1 | `RegisterFile @ 0x14086eff0` | 真实 VA = `0x14086ed00`；0x14086eff0 仅为函数体内 `mov edi,1` 指令 |
| 2 | `RegisterStableBytes @ 0x14086e640` | 真实 VA = `0x14086e760`；0x14086e640 仅为函数体内 `call validateItemSize` 前 |
| 3 | 接力点称 "RegisterFile 经 readFileBounded/RegisterBytes" | 实证：直接 OpenFile+Stat+io.ReadAll+register()，完全不经过 readFileBounded |
| 4 | 接力点称 "RegisterStableBytes 空 id 经 sha256HexPrefix→RegisterBytes" | 实证：id 空时经 `stableLauncherAssetID(ns, TrimSpace(id), data)` 派生 → `registerStable` |

**配套测试**：`TestOpenFileBounded`/`TestReadFileBounded` 重写适配新签名；`TestRegisterFile`/`TestRegisterStableBytes` 保留。

**其他修整**：`ReadBytes`/`ServeAssetRequest` 中 `readFileBounded`/`openFileBounded` 调用点适配新签名；`ServeAssetRequest` 文件回退路径从空写入改正为 `io.Copy(w, file)`。

**指标变化**：
- `backend` 精确函数级档位：S=524 / S-inline=1 / S-sig=62 / P=43 / 未标记=270
- `go vet` / `build -tags production` / `go test -count=1 -p=1` 全绿（0.127s）

#### 新增工具
- `tools/count_funcs.awk` — 修正版 awk 统计脚本（匹配 `[S]`、`[S 汇编...]`、`[P]`、`[P 汇编...]` 等变体）
- `tools/count_funcs.sh` — 同功能的 bash 封装

#### 已知移交偏差（已修复 4 条，剩余 2 条）

| # | 上轮标记 | 状态 | 说明 |
|---|---|---|---|
| 1 | registerStable 缺 ensureCapacityLocked | ✅ 批次 33 已修复 | itemSize=version，栈槽 0x328 确证 |
| 2 | openFileBounded/readFileBounded/RegisterFile/RegisterStableBytes 标记 [P] | ✅ 批次 34 已修复 | 全量重写为 [S] |
| 3 | **validateLauncherConfigIconData.func1 未实现** | ⏳ [P] | disasm 实录 48B（调 `launcherConfigIconBudget.validate`/`launcherConfigIconBudget.addMetrics`），体极小<120 行 |
| 4 | **collectLauncherConfigIconRefs/RefCounts 空桩** | ⏳ [P] | 调用点`savePreparedUnlocked` 忽略返回值，无阻塞 |

#### 新开工优先级

1. **🥇 validateLauncherConfigIconData.func1 + launcherConfigIconBudget.validate/addMetrics** — 三函数合计 <120 行，有 disasm 和 source_funcs 行号蓝图
2. **🥇 windowManagement 域** — `ClearWindowManagementTarget`/`GetWindowManagementState` 已有 asm 实录；`attachWindowManagementTargetIconURL`(0x1408a4400) 可套接 icon 工具链
3. **🥉 collectLauncherConfigIconRefs/RefCounts 空桩** — launcherconfig.go:362-364，无阻塞

#### 纪律约束（与 HANDOFF §6 同源，批次 34 新增）

1. **统计脚本不再依赖 `[S]` 独立子串**：`tools/count_funcs.sh` 已改版，开工可直接用
2. **VA 必须从符号表确认后方可信任**：disasm 文件名和接力点手抄 VA 都可能不准——开工第一事先查 `symbols.txt` 对应行
3. **RegisterStableBytes 空 id 推导入参**：disasm 实证第三参为 `data` 切片本身（非 namespace 子串/切片前 N 字节），开工务必确认新会话中 `data` 类型与寄存器

---

### 批次 36 — windowManagement 域 `[P]→[S]` 升级 + 新文件 `windowmanagement.go`（2026-09-13）

#### 完成工作

| 函数 | VA | Disasm | 开工 | 收工 | 关键修正 |
|---|---|---|---|---|---|
| `windowManagementService.Configure` | 0x1409dfda0 | 640B | [S-sig]⁽¹⁾ | [S] | 372L 体：双锁(operationLock+lock) → normalizeConfig → 写字段 → reconcileRuntime |
| `windowManagementService.GetState` | 0x1409e0480 | 512B | [P]⁽¹⁾ | [S] | 完整控制流：operationLock → reconcileRuntime → buildState → return (WindowManagementState, error) |
| `windowManagementService.SetTarget` | 0x1409e0680 | 1536B | [P]⁽¹⁾ | [S-sig][P] | 321L 体，deferwrap 双锁上下文保留 |
| `windowManagementService.ClearTarget` | 0x1409e0c80 | 1248B | [P]⁽¹⁾ | [S][P] | 实证三寄存器 0x88/0x8/0xf1 双锁检测，persistConfig 调用 |
| `windowManagementService.UpdateConfig` | 0x1409e1160 | 320B | [P]⁽¹⁾ | [S] | currentModuleEnabled → Configure → buildState → return |
| `windowManagementService.SetCursorWrap` | 0x1409e12a0 | 928B | [P]⁽¹⁾ | [S-sig][P] | 13L 体，配置更新+reconcileRuntime |
| `windowManagementService.SetTopMost` | 0x1409e1640 | 1504B | [P]⁽¹⁾ | [S-sig][P] | 20L 体，平台函数桩 |
| `windowManagementService.SetOpacity` | 0x1409e1c20 | 2496B | [P]⁽¹⁾ | [S-sig][P] | 47L 体，透明度快照链 |
| `windowManagementService.SetResolution` | 0x1409e25e0 | 1856B | [P]⁽¹⁾ | [S-sig][P] | 24L 体 |
| `windowManagementService.ToggleBorderless` | 0x1409e2d20 | 4128B | [P]⁽¹⁾ | [S-sig][P] | 389L 体，全屏快照链 |
| `windowManagementService.ToggleFullscreen` | 0x1409e3d40 | 4128B | [P]⁽¹⁾ | [S-sig][P] | 330L 体 |
| `windowManagementService.PickTarget` | 0x1409e4d60 | 1056B | [P]⁽¹⁾ | [S-sig][P] | 16L 体 |
| `windowManagementService.Shutdown` | 0x1409e5180 | 480B | [P]⁽¹⁾ | [S] | 完整 control flow：双锁→shuttingDown guard→ restoreAllManagedTargets→stopCursorWrap→setLastError |
| `windowManagementService.captureManagedSnapshot` | 0x1409e5360 | 960B | [P]⁽¹⁾ | [S] | mapaccess2_fast64 + duffcopy snapshot 结构 |
| `windowManagementService.rememberManagedSnapshot` | 0x1409e5720 | 320B | [P]⁽¹⁾ | [S] | nil hwnd guard → lock → makemap_small → mapassign + xmm 赋值 |
| `windowManagementService.restoreManagedTarget` | 0x1409e5860 | 4544B | [P]⁽¹⁾ | [S-sig][P] | 205L 体，平台层窗口操作 |
| `windowManagementService.restoreAllManagedTargets` | 0x1409e6a20 | 1888B | [P]⁽¹⁾ | [S-sig][P] | 26L 体 |
| `windowManagementService.currentModuleEnabled` | 0x1409e7180 | 288B | [P]⁽¹⁾ | [S] | lock → movzx s+0xf0 → unlock → return |
| `windowManagementService.currentValidTarget` | 0x1409e72a0 | 800B | [P]⁽¹⁾ | [S] | 4 路 error exit itab实证，windowManagementValidateTarget 调用 |
| `windowManagementService.reconcileRuntime` | 0x1409e75c0 | 352B | [P]⁽¹⁾ | [S] | 锁内 duffcopy config → startCursorWrap/stopCursorWrap 分支 |
| `windowManagementService.buildState` | 0x1409e7720 | 3264B | [P]⁽¹⁾ | [S-sig][P] | 65L 体，窗口枚举+信息收集 |
| `windowManagementService.setLastError` | 0x1409e83e0 | 448B | [P]⁽¹⁾ | [S] | 完整 control flow：err==nil→clear, err!=nil→Error()+gcWriteBarrier |
| `WindowManagementConfig.UnmarshalJSON` | — | — | [P] | [S-sig][P] | 112L 行号蓝图，JSON 别名 |
| `normalizeWindowManagementConfig` | — | — | [P] | [S-sig][P] | 57L 行号蓝图 |
| `normalizeWindowManagementTarget` | — | — | [P] | [S-sig][P] | 80L 行号蓝图 |
| `upsertWindowFullscreenSnapshot` | 0x1409e85a0 | 1888B | [P] | [S-sig][P] | 18L 体 |
| `removeWindowFullscreenSnapshot` | 0x1409e8d00 | 1024B | [P] | [S-sig][P] | 11L 体，HWND 去重 |
| `windowManagementValidateTarget` | 0x1409e9200 | — | [P] | [S-sig][P] | 平台层桩（待 windowmanagement_windows.go） |
| `startCursorWrap` / `stopCursorWrap` | 0x1409ecaa0 | — | [P] | [S-sig][P] | 14+16L 体 |

⁽¹⁾ 此前存于 `service_configure_stubs.go`（9 个 `[P]` 骨架）和 `bootstrapservice_state_deps.go`（`Shutdown` 桩）。全部迁移到 `windowmanagement.go`。

#### 关键修正（4 处）
1. **`GetState` 签名 `interface{}` → `(WindowManagementState, error)`**：asm 实测三段返回（rax=state, rbx=err, duffcopy）。`BootstrapService.GetWindowManagementState` 同步调整。
2. **`PickTarget` 签名 `(interface{}, error)` → `(WindowManagementTarget, error)`**：类型化，不返回裸接口。
3. **`Shutdown` 无返回值**（非 `error`）——asm `test rbx,rbx` 检查 err 不走返回值。
4. **`currentValidTarget` 四路 error exit**：shuttingDown/moduleDisabled/noTarget/invalidWindow 对应 4 个不同 `newobject` itab。

#### 移交偏差（无新增）
- `attachWindowManagementStateIconURLs` / `attachWindowManagementTargetIconURL`：未实现，`GetWindowManagementState` 调用点暂移除
- `cursorWrapLoop`：未实现

#### 指标变化
- **FUNCS=935**（+29，从 906）
- **S=542**（+10，从 532）
- **S-sig=82**（+20，从 62）
- **P=51**（+8，从 43）
- **UNMARKED=259**（-9，从 268）

---

### 批次 37 — launcherconfigiconstore.go 全量创建（icon 存储层闭环）（2026-09-13）

#### 完成工作

新建 `backend/launcherconfigiconstore.go`（496 行）+ `backend/launcherconfigiconstore_put.go`（294 行），合计 24 个函数：

| 函数 | VA | 标记 | 说明 |
|---|---|---|---|
| `launcherConfigIconStorePath` | 0x1408925e0 | [S] | 4 个 .rdata 常量解码确证：`UsbEAm_Launcher_Config.json`(27B) / `UsbEAm_Launcher_Icons.json`(26B) / `.icons.json`(11B) / `config`(6B) |
| `launcherConfigIconStoreForConfigPath` | 0x140892740 | [S] | 尾调 StorePath + ForPath |
| `launcherConfigIconStoreForPath` | 0x140892780 | [S] | **实证修正**：用 `sync.Map` 全局缓存（非裸构造） |
| `normalizeLauncherConfigIconRef` | 0x1408978e0 | [S] | 已有，复核其签名 |
| `normalizeLauncherConfigIconRefSet` | 0x140894e80 | [S] | map 遍历规范化 |
| `collectLauncherConfigIconRefs` | 0x140894ca0 | [S] | 从 `launcherconfig.go` 迁入（蓝图归属确证） |
| `newLauncherConfigIconRef` | 0x1408977c0 | [S] | sha256 + 手写 hex 表 |
| `launcherConfigIconStoreError` | 0x140897b00 | [S] | **实证修正**：模板 `%s: %s: %s`（10B），之前错误记为 `"launcherConfigIconStore: %s: %s"` |
| `launcherConfigIconAssetCacheKey` | 0x1408a2e40 | [S] | 已存在 |
| `decodeLauncherConfigIconDataURL` | 0x1408973a0 | [S-sig] | `data[5:逗号]` 段上找 `;`（非 `header[:逗号]`） |
| `(*storeBudget).add` | 0x1408965e0 | [S] | 三道阈值 8192/64MiB/32MiPx |
| `(*store).Delete` | 0x1408934c0 | [S-sig] | os.Remove + 清缓存 |
| `ensureLauncherConfigIconJSONEOF` | 0x140897a00 | [S-sig] | Decode + errors.Is(io.EOF) |
| `(*store).ensureLoadedUnlocked` | 0x1408951a0 | [S] | 已有 |
| `(*store).Resolve` | 0x140892e40 | [S] | 已有 |
| `(*store).Put` | 0x1408928e0 | [S-sig] | `[P]` 体预算校验段 |
| `(*store).ExtractLauncherConfigIcons` | 0x140893700 | [S-sig] | 参数透传 |
| `(*store).ExternalizeLauncherConfigIcons` | 0x140893740 | [S-sig] | 参数透传 |
| `(*store).externalizeLauncherConfigIcons` | 0x1408937a0 | [S-sig][P] | 巨型帧(0x14f8)，体未还原 |
| `(*store).Prune` | 0x1408948e0 | [S-sig] | set diff + 删除 |
| `(*store).loadUnlocked` | 0x140894fa0 | [S-sig] | 委派给 ensureLoadedUnlocked |
| `(*store).writeUnlocked` | 0x140895a80 | [S-sig] | JSON 序列化+写入+清缓 |
| `validateLauncherConfigIconLibrary` | 0x140895e80 | [S-sig] | 版本/条目/内容检查 |
| `readLauncherConfigIconStoreBytes` | 0x1408968a0 | [S-sig] | 128MiB 上限 |
| `writeLauncherConfigIconLibraryFile` | 0x140896de0 | [S-sig] | JSON 落盘 |

#### 符号冲突消解
- `iconstore_service.go` 转迁移桩（原有 4 函数全迁）
- `bootstrapservice_config.go` 摘除 `launcherConfigIconStoreForConfigPath` + `(*store).Delete` 两个 `[P]` 骨架
- `launcherconfig.go` 摘除 `collectLauncherConfigIconRefs`（蓝图归属确认为 `launcherconfigiconstore.go`）

#### 实测关键修正（3 处）

1. **`launcherConfigIconStoreForPath` 全局缓存**：现有 `[P]` 骨架的 `&launcherConfigIconStore{path: ...}` 语义不等价，asm 实证走 `internal/sync.HashTrieMap.Load`/`LoadOrStore` → 重写为 `sync.Map` 全局缓存。

2. **`launcherConfigIconStoreError` 模板**：asm 的 `mov ebx, 0xa` 确证 10 字节模板 + 3 个 iface 参数。`.rdata` VA `0x140c440c7` 解码得 `%s: %s: %s`，现实现 `"launcherConfigIconStore: %s: %s"` 的长度（31B vs 10B）与参数数（2 vs 3）均不匹配 → 修正并记入实证错误表。

3. **`StorePath` 4 个 .rdata 常量**：手算 VA 时两处进位错误（前缀 0x140cc4b65→实际 0x140c47565，缺省叶名 0x140cd7638→实际 0x140c37638），用 `resolve_consts.py` 脚本化计算修正。

#### 移交偏差
- `externalizeLauncherConfigIcons` 巨型帧未还原：帧 0x14f8 字节包含 `validateLauncherConfigIconData` 前置 + 9 字段×0x126 qword 复制 + 随机种子初始化 + json.Encoder 写入，完整体需独立批次
- `cursorWrapLoop` / `attachWindowManagementStateIconURLs` 等批次 36 移交偏差未动

#### 指标变化
- **FUNCS=952**（+17，从 935）
- **S=546**（+4，从 542）
- **S-sig=97**（+15，从 82）
- **P=49**（-2，从 51）
- **UNMARKED=259**（持平）
- `go vet` / `build -tags production` / `go test -count=1 -p=1` **全绿**（0.145s）

---

### 批次 38 — launcherconfigiconcommit.go 全量创建（icon commit 管线闭环）（2026-09-13）

#### 完成工作

新建 `backend/launcherconfigiconcommit.go`（274 行，13 函数），从 `source_funcs.txt:2083-2097` 蓝图还原。从 `launcherconfig.go` 迁移摘除 2 个 `[P]` 骨架。补了 2 个缺失的 asm dump（`prepareLauncherConfigIconsForCommit` 0x140890ec0 + `prepareLauncherConfigIconsForCommitWithCurrentRefs` 0x140891120）。

| 函数 | VA | 标记 | 说明 |
|---|---|---|---|
| `(*launcherConfigIconHydrationError).Error` | — | [S] | 类型已在 `types_launcher.go` |
| `(*launcherConfigIconHydrationError).Unwrap` | — | [S] | 类型已在 `types_launcher.go` |
| `(*launcherConfigIconCommitPlan).finish` | 0x140891da0 | [S] | 14L，委派 syncLauncherConfigBeforeIconPrune |
| `collectLauncherConfigIconRefCounts` | 0x1408914c0 | [S] | 从 launcherconfig.go 迁入（同原有 [S] 实现） |
| `collectLauncherConfigIconRefCounts.func1` | — | [S] | 闭包同体 |
| `prepareLauncherConfigIconsForCommit` | 0x140890ec0 | [S] | 121 asm 行，委派 withCurrentRefs |
| `prepareLauncherConfigIconsForCommitWithCurrentRefs` | 0x140891120 | [S] | **220L 主体**：collect → hasInline → budget → externalize → recollect |
| `validateLauncherConfigIconCandidateBudget` | 0x1408916a0 | [S-sig] | 14L 体 + func1(50L)，骨架待 externalizeLauncherConfigIcons 还原后完善 |
| `validateLauncherConfigIconCandidateBudget.func1` | 0x140891880 | [S-sig] | 预算随机采样闭包 |
| `syncLauncherConfigBeforeIconPrune` | 0x140891ea0 | [S] | 14L，orphan set → store.Prune |
| `buildSelfContainedLauncherConfig` | 0x140891fa0 | [S-sig] | 37L 体 + func1(36L)，骨架 |
| `buildSelfContainedLauncherConfig.func1` | 0x140892220 | [S-sig] | slot 内联闭包 |
| `launcherConfigHasInlineIconData` | 0x1408924e0 | [S] | 11L + func1(4L)，visit 遍历 slot.Data |
| `launcherConfigHasInlineIconData.func1` | 0x140892540 | [S] | slot.Data 非空检查 |

#### 符号冲突消解
- `launcherconfig.go` 摘除 `prepareLauncherConfigIconsForCommitWithCurrentRefs` + `collectLauncherConfigIconRefCounts` 两个函数（移至 `launcherconfigiconcommit.go`）
- `collectLauncherConfigIconRefs` 此前已在批次 37 迁至 `launcherconfigiconstore.go`——本批次做了最终确认

#### 指标变化
- **FUNCS=963**（+11，从 952；+13 新文件 −2 摘除）
- **S=546**（+6：Error/Unwrap/finish/collectRefCounts/hasInline/hasInlineFunc1）
- **S-sig=97**（+4：validateCandidateBudget + func1 + buildSelfContained + func1）
- **P=47**（−2：从 launcherconfig.go 摘除 2 个 [P] 骨架）
- **UNMARKED=259**（持平）
- `go vet` / `build -tags production` / `go test -count=1 -p=1` **全绿**（0.125s）

#### 移交偏差
- `launcherConfigIconStorePath` 4 个 .rdata 常量未解码——`externalizeLauncherConfigIcons` 路径空时退化 no-op（兼容测试）
- `cursorWrapLoop` / `attachWindowManagementStateIconURLs` 未还原——不影响本批次
- `validateLauncherConfigIconCandidateBudget` 体为骨架，完整预算校验依赖 `externalizeLauncherConfigIcons` 还原后的实际阈值逻辑

---

### 批次 39 — launcherconfigicon.go 迁移历史函数 + externalizeLauncherConfigIcons 还原（2026-09-13）

#### 完成工作

**目标 A — launcherconfigicon.go 8 个迁移历史函数**

新建 6 个包级函数（从 `source_funcs.txt:2075-2082` 蓝图还原），含现场 asm dump（8 个 VA 区间共 ~2200 字节）逐条翻译：

| 函数 | VA | 标记 | 说明 |
|---|---|---|---|
| `migrateHistoricalLauncherConfigIcons` | 0x14088f680 | [S] | 遍历全部槽位，预算超限时重编码/丢弃（回传 migrated, rescaled, dropped） |
| `migrateHistoricalLauncherConfigIcon` | 0x14088f980 | [S] | 单图标 data:URI → 解码 → 缩放(≤512) → PNG 重编码 → base64 |
| `decodeHistoricalLauncherConfigIcon` | 0x14088ff60 | [S] | 校验 `data:image/{png/jpeg/webp};base64,` 前缀 → base64 解码 |
| `ensureLauncherConfigIconMigrationBackup` | 0x1408904a0 | [S] | 委托 `.before-icon-migration-v1.bak` 备份 |
| `ensureLauncherConfigIconLibraryMigrationBackup` | 0x140890560 | [S] | 委托 `.before-icon-library-v1.bak` 备份 |
| `ensureLauncherConfigMigrationBackupFile` | 0x140890620 | [S] | 实际落盘：Lstat → MkdirAll → OpenFile(O_EXCL) → io.Copy → Sync |
| `ensureLauncherConfigMigrationBackupFile.func1` | (内联 defer) | [S] | 关闭文件 + 未提交时 Remove |

新增依赖 `golang.org/x/image/draw`（CatmullRom 缩放核，原 asm 调用 `draw.Kernel.Scale`）。

**目标 B — externalizeLauncherConfigIcons 完整体还原**

`backend/launcherconfigiconstore_put.go` 原有 100 行 `[P]` 骨架替换为 140 行 `[S-sig]` 实装（asm 0x1408937a0 帧 0x14f8 字节逐条确证）：

流程：`validateLauncherConfigIconData` → `visitLauncherConfigIconSlots` 收集内联 → `normalizeLauncherConfigIconRefSet`（入参 refSet + 现有 Ref 归并）→ 加锁 → `loadUnlocked` → Prune 不在此集合的 Icons → `decodeLauncherConfigIconDataURL` 逐条解码 → 预算校验 → 写入 `cached.Icons` → `writeUnlocked` 持久化。`ExtractLauncherConfigIcons` / `ExternalizeLauncherConfigIcons` 签名为新增 `refSet map[string]struct{}` 参数对齐 asm 实证。

#### 符号冲突消解
- `launcherconfigiconstore_put.go` `externalizeLauncherConfigIcons` 从 `[P]` 骨架升级为 `[S-sig]`
- `launcherconfigiconcommit.go` 调用点 `ExternalizeLauncherConfigIcons(cfg)` → `ExternalizeLauncherConfigIcons(cfg, nil)`

#### 指标变化
- **FUNCS=969**（+6，从 963）
- **S=552**（+6：6 个新迁移函数）
- **S-sig=97**（持平：externalize 从 [P] 骨架→[S-sig] 实装，计数不变）
- **P=46**（−1：externalizeLauncherConfigIcons 升级出 P，无新 P 引入）
- **UNMARKED=259**（持平）
- `go vet` / `build -tags production` / `go test -count=1 -p=1` **全绿**（0.125s，`TestConfigureChangeTriggersReschedule` 已知预存波动）

> **批次 39 指标校正（2026-09-20）**：上面批次 39 的 `S=552/S-sig=97/P=46/UNMARKED=259` 与活体实测不符且分项相加 955 ≠ `FUNCS=969`。实测为 `S=551/S-sig=100/P=49/UNMARKED=268`。原因是历次档位由不同 awk 变体与人工誊抄产生。本段保留原始记录不改写，**以 §1/§5 的活体数字为准**。

---

## 10. 未落地原始文件清单（38 个，2026-09-30 批次 278 实测）

> 生成方式：`source_funcs.txt` 的 `^File: ` 清单（去除 `<autogenerated>`）与 `backend/` 非测试 `.go` 文件名做差集。新会话开工前请重跑：
> ```powershell
> $orig = Select-String docs\goresym\source_funcs.txt -Pattern '^File: (.+?)\s*$' | ForEach-Object { $_.Matches[0].Groups[1].Value }
> $have = Get-ChildItem backend -Filter *.go | Where-Object { $_.Name -notmatch '_test\.go$' } | ForEach-Object Name
> $orig | Where-Object { $_ -ne '<autogenerated>' -and $have -notcontains $_ } | Sort-Object
> ```

```
audio.go
bookmarks_firefox.go
bookmarksqlite.go
bookmarktitle.go
desktopwidgets_bootstrap.go
desktopwidgets_scheduler.go
desktopwidgets_weather_openmeteo.go
desktopwidgets_weather_qweather.go
desktopwidgets_weather_weatherapi.go
filesearch_usn_follower_windows.go
launcherappidentity_windows.go
launcherbackground_windows.go
launcherconfigstore.go
launcherglobalhotkey_windows.go
launchericonasset.go
launcherinstance_windows.go
launchernetwork.go
launcherprocess_windows.go
launcherstartup_windows.go
launcherstartup.go
launcherupdate_helper_windows.go
launcherwindow.go
log.go
memoryrelease_windows.go
pluginupdate.go
pluginwindow.go
qrcode.go
screenshot_displayconfig_windows.go
screenshot_image_budget.go
screenshot_preview_native_windows.go
screenshot_preview.go
screenshot_rotation_windows.go
screenshot_selection_toolbar.go
screenshot_tonemap_windows.go
screenshot.go
transport.go
twofactor.go
workspacemaintenance.go
```

## 11. 产物与验收记录状态（2026-09-20 实测）

| 项 | 状态 | 说明 |
|---|---|---|
| `artifacts/backend_root.exe` | 最新（2026-09-13 16:43，20.9 MB） | 批次 39 之后最近一次编译 |
| `artifacts/backend_service.exe` | 2026-09-13 15:55（20.9 MB） | 批次 39 中途产物 |
| `artifacts/UsbEAm_Launcher_rebuilt.exe` | **最新**（2026-09-20 批次 43，19.8 MB） | 批次 43 之后 `bash build.sh` 重建；SHA256 `C7EC7E01C23014D2326833FF779FA34F5A15EDA940D838446D52832D9BEACF06` |
| `docs/acceptance/` | 批次 40 已补 | `batch-A.md` + `batch40.md`；批次 B/C/D 及 19–39 仍无四路验收记录，与 REBUILD_EXECUTION §6.1 冲突 |

**验收补账义务**：批次 40 起，每批完工必须写 `docs/acceptance/<batch>.md`（G1 编译 / G2 契约 / G3 行为 / G4 独立复核），未写 = 该批未完成。历史批次不必回头补写，但不得继续缺位。

---

## 12. 批次 40 记录（appicon GDI 渲染/提取/缓存层）

- **目标**：`backend/appicon_windows.go` 46 函数全 UNMARKED 起步，52 个 asm 资产逐函数直译。
- **产出**：45 函数 + `absInt` 全部 `[S]`；删除 ghost 符号 `getBitmapSize`（symbols.txt 无此符号，GetObjectW 已内联于 `appIconBitmapSize`）。
- **签名修正核心**：
  - `AppIconOptions` 24B→64B（`IconIndex int32` + pad + `Namespace string` + `Size int` + `ImageList int` + `CandsPtr unsafe.Pointer` + `CandsLen int` + 8B 保留）。
  - 全局缓存 `appIconDataCache sync.Map` 新增（`resolveAppIconDataWithOptions` 与 `resolveSystemIconDataWithOptions` 的 Load/Swap 目标同址 `0x141C12740`）。
  - 入口链 10 函数 `(string,error)` → 单 `string`；`loadAppIconInfo` 改 `(hicon uintptr)(*appIconInfo,int,int,error)`；`appIconBitmapSize` 加 error；`createAppIconCanvas` 去 hdc 参数返回 3 值；`buildRGBAFromDIBBits` 单返回 unsafe.Pointer；`normalizeAppIconPNGWithSize`/`visibleIconBounds` 改 bool 双返回；`fitIconToSquare` 最近邻缩放。
  - `applyAppIconMaskAlpha` 掩码语义反转（RGB 全零=绘制区域）；`appIconHasVisibleAlpha` 改 `A != 0`。
  - `getSystemIconIndexWithAttributes` 3 参（useFileAttributes，0x4010/0x4000 回退重试）；`loadPrivateExtractedAppIcon`/`loadShellDefinedAppIcon`/`loadImageListIcon` 参数类型对齐；`buildResourceAppIconCacheKey` 改 `(url,a,b int64)`；`resolveFileSearch*` 签名对齐。
- **外部调用点修正**：`bootstrapservice.go:1019`（`data, err :=` → `data :=`）；`startmenu_icon_windows.go` 6 处（`if data, err := X(); err == nil && data != ""` → `if data := X(); data != ""`，`IconIndex: uint32` → `int32`）。
- **测试重写**：22 测试函数对齐新签名（mask 反转、scale 语义、bool 双返回、nil panic、A!=0）。
- **门禁**：`go build -tags production -trimpath` / `go vet` / `go test -count=1 -p=1` 三项 EXIT=0。
- **artifact**：`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`（19,726,848 B，SHA256 `1D18BFEE3B5CEEE4A36A4AAAC1107ADC10A82E65B61E293D30F1EB7F45A66AAE`）。
- **验收**：`docs/acceptance/batch40.md`（G1/G2/G3/G4 四路 PASS，含 5 处 G4 自纠）。
- **实证常量**：全局候选尺寸 `[0x100,0x80,0x40,0x30,0x20]`（rodata `0x141be9520` 解码）；`IID_IImageList` GUID `{46EB5926-582E-4017-9FDF-E8998DAA0950}`；SHIL_JUMBO=4（iImageList==0 时）。

---

## 13. 给新会话：从这里继续（策略已升级为 L0→L1→L2 分层）

**先读 `docs/STRATEGY.md`**——它是还原策略的唯一真相源，L0→L1→L2 分层、`[S-eq]` 标记、对拍验证范式都在里面。本节的推进顺序服从它。

**仓库已 git 化（2026-09-27）**：远端 `https://github.com/wsolarq11/Reverse_UsbEAmSource`（公开，2026-09-30 起常驻 public），首次提交 `d0246a7` 含全部源码 + 文档 + 反汇编资产 + 策略。新会话工作流：`git pull` 拉最新 → 按 §6 开工 → 每批收尾 commit + push → CI 自动验证。

**本地 git 传输配置（必读，否则 push 报 `schannel: failed to receive handshake`）**：本机 git 直连与默认 schannel 走系统代理均失败，已写入仓库级配置：`http.sslBackend=openssl`、`http.proxy=http://127.0.0.1:7890`、`http.postBuffer=524288000`。换新机器或重装后若 push 报 schannel 握手失败，重跑这三条 `git config` 即可。CI 定义在 `.github/workflows/ci.yml`（windows-latest + go1.25.12，build/vet/test 三项，push 到 main 自动触发）。

**CI 验证惯例（2026-09-30 起，用户定规）**：仓库**常驻 public**。凡是需要触发 CI 的，统一**先确认/转为 public 再跑**；禁止先以 private 跑 CI 发现跑不通（`job not started (account payments failed)`——私有仓 windows runner 受账号计费限额，job 直接不启动）再转 public 重跑。公开仓库 Actions 免费无限分钟，不受计费影响。若仓库意外回到 private，先 `gh repo edit --visibility public --accept-visibility-change-consequences` 再触发 CI。

批次 43 已闭环，screenshot image_budget 内存预算链全量 `[S]`，门禁三连绿，artifact 已重建。你接手时不要从零判断，直接按下面顺序推进。

**第一步，校准基线（5 分钟内）**：先跑 `bash tools/count_funcs.sh` 取活体数字，再跑 `go build -tags production -trimpath ./backend`、`go vet -tags production ./backend`、`go test -count=1 -p=1 -tags production ./backend` 三项确认 EXIT=0。文档里任何历史数字若与此冲突，以活体实测为准，不要抄文档。

**第二步，按 §6 顺序开工**：`bootstrapservice_callees.go`（34 未标）已在批次 41 清零。Screenshot 运行时族继续推进——批次 42 已闭环 `png_fast`、批次 43 已闭环 `image_budget`、批次 44 已闭环 `capture_windows` GDI 核心、批次 45 已闭环 `cursor` 数据搬运层、批次 46 已闭环 `cursor_draw` 绘制收口、批次 47 已闭环 `capture_virtual` GDI 虚拟屏捕获、批次 48 已闭环 `dxgi_hdr_parse` 模式解析与入口探测、批次 49 已闭环 `dxgi_debug` 调试格式化链、批次 50 已闭环 `dxgi_com` COM 薄包装，下一子批次建议 `dxgi_enum`（`enumerateScreenshotDXGIAdapterDisplays`/`enumerateScreenshotDisplayCaptureInfos` + `captureScreenshotVirtualScreenSnapshotWithOptions`/`defaultScreenshotScreenCaptureBackend` 收口，全部 asm 已 dump 到 pipeline/tmp），之后 `clipboard`，每批单独过门禁。

**四条不改的铁律**：① 签名未定型拒绝落体，宁留 `[S-sig]`/`[P]` 并写明阻断原因；② asm 直译优先于语义推断，锁内读与锁后读严格区分（照搬 `lock.Lock()` 包裹可能直接死锁，见 §6 纪律 5）；③ L1 落体必须带对拍测试，无测试不算 L1（只能算 `[S-sig]`）；④ 长尾停 L1、核心才升 L2，禁止对所有函数无差别追 `[S]`。所有文件改动走编辑工具，不得命令行改文件。

**每批收尾五件套**：写 `docs/acceptance/<batch>.md`（G1/G2/G3/G4）→ 跑 `count_funcs.sh` 更新 §1 指标 → 在 HANDOFF 尾部追加批次记录 → `git add -A && git commit && git push` 触发 CI → 盯 `gh run list` 确认 CI 三绿。缺任何一件，该批不算完成。

当前进度锚点：`FUNCS=1026 / MARKED=837 / UNMARKED=189`，函数覆盖约 22%，未落地文件 110 个（§10）。
**目标口径已升级为总进度 100%**：4,754 蓝图函数 100% 还原、110 未落地文件 100% 落地、`UNMARKED=0` 且 `P=0`；全程保持可编译 + 功能一致，不是字节级同哈希。**判定完成的标准改为：`count_funcs.sh` 显示 `FUNCS≥4,754`、`UNMARKED=0`、`P=0`，且 §10 差集为空。**

---

## 14. 批次 41 记录（bootstrapservice_callees.go 34 未标落档）

- **目标**：`backend/bootstrapservice_callees.go` 34 个 UNMARKED 函数全部落档。
- **资产**：pipeline/tmp 已有 12 个 + `va_dump.py` 现场抽取 21 个（VA 全部来自 `symbols.txt`，长度来自相邻符号差，未手抄文档 VA）。新资产前缀：`pickWindowProcessForService`、`captureQRCodesFromScreenSelectionForService`、`buildQRCodeDecodeResultFromPNG`、`BootstrapService.languageDirectories`、`loadLanguageMessages`、`discoverLanguages`、`languageDirectoriesForWorkspace`、`buildBookmarkState`、`resolveBookmarkPageTitleWithNetwork`、`openLinkWithPreferences`、`isSupportedAppPath`、`buildDesktopCalendarMonth`、`buildDesktopWorldClockSnapshot`、`sendLauncherUpdateHealthFromEnvironment`、`startDetachedCommand`、`screenshotPinWindowService.PersistSnapshotsForShutdown`、`defaultLauncherConfigWithOptions`、`normalizeLauncherConfigWithOptions`、`launcherConfigsEqual`、`detectBookmarkSources`、`detectLinkBrowsers`。
- **产出**：4 [S] + 4 [S-sig] + 23 [P] + 3 幽灵删除。
  - [S]：`languageDirectoriesForWorkspace`(0x1407a27a0)、`isSupportedAppPath`(0x14088ce00)、`launcherConfigsEqual`(0x14088bcc0)、`resolveScreenshotPNGFromRef`(0x140799a00，恢复批次 24 丢失标记)。
  - [S-sig]：`startDetachedCommand`(0x1408a8480，首参实为 `[]string` 非 string)、`screenshotPinWindowService.PersistSnapshotsForShutdown`(0x140981580，只落锁/标志语义)、`BootstrapService.loadLinkPreferences`(0x140769c60，包级改方法)、`resolveBookmarkPageTitleWithNetwork`(0x1407709e0，首参实为 string 非 WorkspaceLayout 整结构)。
  - [P]：GPU×6、QR×4、语言×2（loadLanguageMessages/discoverLanguages）、书签×2（buildBookmarkState/detectBookmarkSources）、链接×2（openLinkWithPreferences/detectLinkBrowsers）、桌面×3（buildDesktopCalendarMonth/buildDesktopWorldClockSnapshot/getLauncherBackgroundMetrics）、更新×1（sendLauncherUpdateHealthFromEnvironment）、配置×2（defaultLauncherConfigWithOptions/normalizeLauncherConfigWithOptions）、ResolveBookmarkIconResource。均属 §10 未落地域，已注明真实 VA 与阻断原因。
  - 幽灵删除：包级 `languageDirectories`（symbols.txt 无）、包级 `setHotkeyCaptureLease`（仅方法存在，方法已实现于 bootstrapservice.go:2479）、`parseLauncherImageDataURLHeader`（内联于 decodeLauncherImageDataURL@0x1408a5200）。
- **测试**：新增 `backend/bootstrapservice_callees_test.go`（3 用例）。
- **门禁**：`go build -tags production -trimpath` / `go vet` / `go test -count=1 -p=1` 三项 EXIT=0。
- **artifact**：`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`（19,726,848 B，SHA256 `B468005DD5B03F135D00FFC218BE85D8792BACD7DEEDA57DCED67CAA39A26C63`）。
- **验收**：`docs/acceptance/batch41.md`（G1/G2/G3/G4 四路 PASS）。

**待移交偏差**（批次 41）：
1. `startDetachedCommand` 变长切片边界 `args[:len(args)-1]` 存疑，`RestartApplication` 调用点应改传 `[]string{exe}`——留待 body 还原时一并处理。
2. `resolveBookmarkPageTitleWithNetwork` 首参疑为 `ws.ConfigFile`（string），调用点 `ResolveBookmarkTitle` 当前传整个 `ws`——需还原 `bookmarkPageTitleCandidates`/`fetchBookmarkPageTitle` 后纠正。
3. `PersistSnapshotsForShutdown` 的 `snapshotMirror`/`saveSnapshots` 未接入（screenshot_pin_windows.go 未落地）。
4. 23 个 [P] 函数均已落档并注明真实 VA，是后续 gpu/qrcode/bookmarks/link/desktopwidgets/launcherconfiginput 各专项批次的首选入口。

---

## 15. 批次 42 记录（screenshot_png_fast 自定义 PNG 编码器）

- **目标**：`backend/screenshot_png_fast.go` 蓝图 14 函数全量 asm 直译。
- **资产**：pipeline/tmp 已有 14 个（`encodeScreenshotPNGRowSource`、`writeScreenshotPNGIHDR`、`writeScreenshotPNGFilteredRows`、`selectScreenshotPNGFilter`、`filterScreenshotPNGPaeth`、`predictScreenshotPNGPaeth`、`convertScreenshotRGBARowToPNG`、`writeScreenshotPNGChunk`、`screenshotPNGIDATChunkWriter.Write/.Close/.flush`、`screenshotRGBAImageRowSource.Bounds/.Row`、`encodeScreenshotRGBAWithKlauspostPNG`），VA 全部来自 `symbols.txt`，长度来自相邻符号差。
- **产出**：14 [S] + 4 [S-inline]（`unpremulScreenshotPNG`/`absScreenshotPNGInt`/`absScreenshotPNGByte`/`sumAbsScreenshotPNG`）+ 1 幽灵删除（`encodeScreenshotPNGWithKlauspost`，其旧实现误用标准库 `image/png`）。
- **类型/签名修正**：
  - `screenshotRGBAImageRowSource`：`{source *image.RGBA}`(8B) → `image.RGBA` 底层(40B，Pix@0/Stride@0x18/Rect@0x20)，与 `Bounds`/`Row` asm 偏移 `[rax+0x20]…[rax+0x38]` 一致。
  - `encodeScreenshotRGBAWithKlauspostPNG`：`(*image.RGBA)` → `(*screenshotRGBAImageRowSource)`；调用点 2 处（`screenshot_read.go:185`、`screenshot_funcs.go:83`）改指针转换。
- **错误字符串 8 个全部 rodata 解码**：`0x140C614FC`(22B)/`0x140C61512`(22B)/`0x140C6C020`(28B)/`0x140C6D837`(29B)/`0x140C6A452`(27B)/`0x140C6C03C`(28B)/`0x140C885E2`(50B)/`0x140C63056`(23B)。
- **依赖**：新增 `github.com/klauspost/compress v1.18.4`（zlib.NewWriterLevelDict level=1，模块缓存已有）。
- **测试**：新增 `backend/screenshot_png_fast_test.go`（3 用例：全 alpha 往返 / 部分 alpha 反预乘 / chunk 头与 CRC）。标准库解码为 `*image.NRGBA`（straight alpha）验证通过。
- **门禁**：`go build -tags production -trimpath` / `go vet` / `go test` 三项 EXIT=0。
- **artifact**：`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`（19,757,056 B，SHA256 `FB24F265F03DD131597FFB4AA9D129090FC2F608863F933EDFAE66CFDDCDF216`）。
- **验收**：`docs/acceptance/batch42.md`（G1/G2/G3/G4 四路 PASS）。

**待移交偏差**（批次 42）：
1. `writeScreenshotPNGFilteredRows` 中具体类型 itab 断言分支（asm 0x140995265，Fun[0] 返回值被丢弃）为编译器接口优化，实现以 memmove 等价覆盖，已在函数注释注明。
2. 4 个 helper（`unpremulScreenshotPNG`/`absScreenshotPNGInt`/`absScreenshotPNGByte`/`sumAbsScreenshotPNG`）是 asm 内联逻辑抽取，非独立符号，标 [S-inline] 不占 UNMARKED。

---

## 16. 批次 43 记录（screenshot image_budget 内存预算/受限读取/解码链）

- **目标**：`screenshotImageMemoryBudget.Reserve` / `readScreenshotImageFileLimited` / `decodeScreenshotImageBytesWithBudget` / `validateScreenshotImageConfig` / `inspectScreenshotImageWork` / `prepareScreenshotImageWork` / `reserveScreenshotImageBounds` 全量 asm 直译。
- **资产**：pipeline/tmp 已抽取 `screenshotImageMemoryBudget_Reserve` / `Reserve.func2` / `Reserve.func2.1` / `validateScreenshotImageConfig` / `inspectScreenshotImageWork` / `prepareScreenshotImageWork` / `decodeScreenshotImageBytesWithBudget` / `readScreenshotImageFileLimited` / `reserveScreenshotImageBounds` / `buildScreenshotResultMetadataFromPNG`（现场补抽 8 个）。
- **签名修正（关键）**：
  - `Reserve(size int64) (func(), error)`：返回 `sync.Once` 包裹的一次性释放闭包（`used -= size`，钳制到 0）；`size<=0` 返回空函数；nil receiver 报错。
  - `buildScreenshotResultMetadataFromPNG(png []byte, mode string)`：删除多余 `budget` 参数（Reserve 用全局实例）；`len(png)==0` 返回零值 + nil error。
  - `decodeScreenshotImageBytesWithBudget(data []byte, expectedFormat string, channels int64) (image.Image, string, func(), error)`：readExternal 传 `(raw, "", 12)`，解码失败先 `release()` 再报错。
  - `inspectScreenshotImageWork(data, expectedFormat, channels, cfg) (color.Model, int, int, string, int64, error)`：`expectedFormat` 空则跳过格式白名单；`EqualFold(format, lower(trim(expected)))`。
- **全局常量（rodata 解码）**：
  - 全局预算 `limit = 0x20000000`（512 MiB，main.init 0x140747328 静态写 `[0x141c5b508]`）。
  - 默认配置 0x141965500：`{64MiB, 32768, 32768, 4e7, 512MiB}`。
  - scrolling 配置 0x141965540：`{512MiB, 32768, 50000, 1.2e8, 512MiB}`。
  - 错误字符串 10 个全解码（reserve74 / reserve_nil30 / readfile_size43 / readfile_short51 / val43a / val37 / val43b / val53 / val30 / inspect46 / inspect28 / inspect39 / decode22）。
- **validate 估算**：`estimated = stride + channels*width*height`，五段校验（源数据 → 尺寸 → 总像素除法防溢出 → 像素/通道 → 工作集除法防溢出）。
- **调用点修正**：`readExternalScreenshotImageFromPath`（`defer release()`）；`bootstrapservice.go:1914`（删 budget 参数）。
- **测试**：新增 `backend/screenshot_image_budget_test.go`（4 用例：Reserve 语义 / validate 边界 / 受限读取 / build 元数据），全 PASS。
- **门禁**：`go build -tags production -trimpath` / `go vet` / `go test` 三项 EXIT=0。
- **artifact**：`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`（19,757,056 B，SHA256 `C7EC7E01C23014D2326833FF779FA34F5A15EDA940D838446D52832D9BEACF06`）。
- **验收**：`docs/acceptance/batch43.md`（G1/G2/G3/G4 四路 PASS）。

**待移交偏差**（批次 43）：
1. `Reserve` 的闭包在 Go 1.25 ABI 下表现为单字 funcval（调用时从 `[funcval]` 取 code），源码用标准 `func()` 闭包 + `sync.Once` 表达，编译器生成等价布局。
2. `validateScreenshotImageConfig` 的 `colorModel` 参数 unused（asm 中 itab/data 仅透传零值），保留参数以对齐 ABI 签名。
3. `buildScreenshotResultMetadataFromPNG` 非 scrolling 路径 Reserve 成功后立即 `release()`（元数据构建不长期占用预算），属 asm 实证语义。

---

## 17. 批次 44 记录（screenshot capture_windows GDI 捕获核心）

- **目标**：新建 `backend/screenshot_capture_windows.go`，落地 GDI 屏幕矩形捕获链核心 6 函数。
- **资产**：pipeline/tmp 已抽取 `captureScreenshotVirtualScreenSnapshotWithOptions` / `captureScreenshotScreenRectWithOptions` / `screenshotGDIScreenCaptureBackend_captureVirtualScreen` / `screenshotGDIScreenCaptureBackend_captureScreenRect` / `captureScreenshotScreenRectGDI` / `qrCodeBitBlt` / `normalizeScreenshotScrollingRect` / `drawScreenshotCaptureCursorOnRGBA` / `rectgdi_deferwrap1..4` + 现场补抽 `qrCodeVirtualScreenBounds` / `drawScreenshotCursorOnRGBA` / `drawScreenshotCursorSnapshotOnRGBA` / `defaultScreenshotScreenCaptureBackend` / `selectScreenshotScreenCaptureBackendName`。
- **产出**：6 [S]（`qrCodeBitBlt`/`getSystemMetrics`/`qrCodeVirtualScreenBounds`/`normalizeScreenshotScrollingRect`/`captureScreenshotScreenRectGDI`/`screenshotGDIScreenCaptureBackend.name`）。新增全局 `procGetDC`/`procReleaseDC`/`procBitBlt`/`procGetSystemMetrics`。
- **签名/语义（asm 实证）**：
  - `qrCodeBitBlt(dstHDC,dstX,dstY,dstW,dstH,srcHDC,srcX,srcY,rop uintptr) error`：`dstW<=0||dstH<=0` → nil；sentinel = `syscall.Errno(0)`（itab kind=12 Uintptr/size=8，.rdata 0x1411cd520 值 0）。
  - `qrCodeVirtualScreenBounds() image.Rectangle`：SM_X/Y/CX/CYVIRTUALSCREEN（0x4c..0x4f），宽或高 `<=0` 回退 `(SM_XSCREEN,SM_YSCREEN,0,0)`，最终 min/max 规范化。
  - `normalizeScreenshotScrollingRect(rect) image.Rectangle`：cmovg 轴排序 → `Empty()` → 零矩形 → `Intersect(qrCodeVirtualScreenBounds())`。
  - `captureScreenshotScreenRectGDI(x0,y0,x1,y1 int) (*image.RGBA, error)`：`GetDC(0)` 失败报 33B 无 %w；错误链 `创建截图设备上下文失败: %w`(37B)/`创建长截图位图失败: %w`(31B)/`选择长截图位图失败: %w`(31B)/`复制长截图像素失败: %w`(31B)；ROP `0x40CC0020`（SRCCOPY|CAPTUREBLT）。
  - `name()` 返回 `"gdi"`（selectScreenshotScreenCaptureBackendName 0x14096e0a0 "disabled" 分支 rodata 0x140c33d31）。
- **defer 语义**：asm 的 4 个 deferwrap（ReleaseDC(0,hdc) / DeleteDC(memDC) / DeleteObject(canvas) / SelectObject 恢复）分别对应 `defer procReleaseDC.Call(0,hdc)`、`defer deleteAppDC`、`defer deleteAppObject`、`defer restoreAppObject`，LIFO 顺序由 Go 源码 `defer` 语句自然保证，未手工复制 deferwrap。
- **错误字符串 6+1 全 rodata 解码**：`0x140c5a032`(18B)/`0x140c740e9`(33B)/`0x140c7a005`(37B)/`0x140c711d8`(31B)/`0x140c711f7`(31B)/`0x140c71216`(31B)/`0x140c4e145`(13B)。
- **门禁**：`go build ./backend` / `go vet ./backend` 两项 EXIT=0，`gofmt -l` 空。
- **artifact**：`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`（19,757,568 B，SHA256 `ED082A47ED3362F133CEB36F66CE9AC392DB290347128A91927EFA3C807C1780`）。
- **验收**：`docs/acceptance/batch44.md`（G1/G2/G3/G4 四路 PASS）。

**待移交偏差**（批次 44）：
1. `screenshotGDIScreenCaptureBackend.captureScreenRect`（0x14096e840）与光标绘制链（`drawScreenshotCaptureCursorOnRGBA`/`drawScreenshotCursorOnRGBA`/`drawScreenshotCursorSnapshotOnRGBA`/`captureScreenshotCursorSnapshot`/`drawScreenshotCursorSnapshotOnDC`）留待批次 45；后三者依赖 qrcode 域 `createQRCodeRGBACompatibleBitmap`(0x140957520)/`copyQRCodeDIBBitsToRGBA`(0x140957980)/`drawScreenshotCursorInfoOnDC`(0x140974080)，本批次已 dump 全部 asm 到 pipeline/tmp。
2. `qrCodeBitBlt` 的 `errors.Is(lastErr, syscall.Errno(0))` 是 asm 实证（target 为 Uintptr kind 且值 0），语义上零 errno 与 nil 均转 `"BitBlt 失败"`。
3. `captureScreenshotScreenRectGDI` 顶部先调 `normalizeScreenshotScrollingRect` 再判空（非直接按原始坐标判空），与旧摘要描述相反，本批次已按 asm 修正。

---

## 18. 批次 45 记录（screenshot cursor 快照数据搬运层）

- **目标**：新建 `backend/screenshot_cursor_windows.go`，落地光标快照采集与 DIB/RGBA 双向像素搬运 5 函数。
- **资产**：pipeline/tmp 现场补抽 `captureScreenshotCursorSnapshot` / `drawScreenshotCursorSnapshotOnDC` / `drawScreenshotCursorInfoOnDC` / `createQRCodeRGBACompatibleBitmap` / `copyQRCodeDIBBitsToRGBA` / `copyRGBAToQRCodeDIBBits` / `currentScreenshotCursorInfo`。
- **产出**：5 [S]（`currentScreenshotCursorInfo`/`captureScreenshotCursorSnapshot`/`copyRGBAToQRCodeDIBBits`/`copyQRCodeDIBBitsToRGBA`/`createQRCodeRGBACompatibleBitmap`）。新增全局 `procGetCursorInfo`。
- **签名/语义（asm 实证）**：
  - `currentScreenshotCursorInfo() (screenshotCursorInfo, bool)`：CURSORINFO 24B（cbSize=0x18），`GetCursorInfo(&info)`，返回 `(Size,Flags,Cursor,ScreenPos, r!=0)`。
  - `captureScreenshotCursorSnapshot() (screenshotCursorInfo, bool)`：`ok && Cursor!=0 && Flags&1(CURSOR_SHOWING)` 才返回，否则零值 + false。
  - `copyRGBAToQRCodeDIBBits(img *image.RGBA, pBits unsafe.Pointer)`：RGBA→BGRA；目标紧凑 stride=width*4；源偏移 `Stride*(y-Min.Y)+(x-Min.X)*4`。
  - `copyQRCodeDIBBitsToRGBA(pBits unsafe.Pointer, img *image.RGBA)`：BGRA→RGBA；`SetRGBA(Min.X+x, Min.Y+y, {R:+2,G:+1,B:+0,A:+3})`。
  - `createQRCodeRGBACompatibleBitmap(img) (uintptr, unsafe.Pointer, error)`：img nil/空 → 50B 错误；CreateDIBSection 失败 → 36B 错误；成功 `copyRGBAToQRCodeDIBBits`。
- **错误字符串 2 个全 rodata 解码**：`0x140c8857e`(50B "创建截图标注文字画布失败: 图像为空")/`0x140c788f7`(36B "创建截图标注文字画布失败")。
- **测试**：新增 `backend/screenshot_cursor_windows_test.go`（2 用例：RGBA↔BGRA 往返 / nil 与空图报错），全 PASS。
- **门禁**：`go build -tags production -trimpath` / `go vet` / `go test` 三项 EXIT=0，`gofmt -l` 空。
- **artifact**：`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`（19,757,056 B，SHA256 `A20BF72A8E0575BC79E277FF3EDC06CAB5F1920A259DE6CCC925344531259E0E`）。
- **验收**：`docs/acceptance/batch45.md`（G1/G2/G3/G4 四路 PASS）。

**待移交偏差**（批次 45）：
1. 光标绘制收口链（`screenshotCursorIconMetrics`/`forceScreenshotDIBAlphaOpaqueInRect`/`drawScreenshotCursorInfoOnDC`/`drawScreenshotCursorSnapshotOnDC`/`drawScreenshotCursorSnapshotOnRGBA`/`drawScreenshotCursorOnRGBA`/`drawScreenshotCaptureCursorOnRGBA` + `screenshotGDIScreenCaptureBackend.captureScreenRect`）留待批次 46，全部 asm 已 dump 到 pipeline/tmp。
2. `drawScreenshotCursorInfoOnDC`（0x140974080）依赖 `screenshotCursorIconMetrics`(0x140974ac0) 与 `forceScreenshotDIBAlphaOpaqueInRect`(0x140974ea0)，批次 46 需先分析这两个。
3. `copyRGBAToQRCodeDIBBits`/`copyQRCodeDIBBitsToRGBA` 的 unsafe.Slice 越界检查（asm 的 panicIndex/panicunsafeslicelen 分支）由 Go 编译器自动生成，源码以 `unsafe.Slice` + 索引表达等价。

---

## 19. 批次 46 记录（screenshot cursor 绘制收口链）

- **目标**：新建 `backend/screenshot_cursor_draw_windows.go`，落地光标绘制收口链 + GDI 矩形捕获方法 11 函数。
- **资产**：pipeline/tmp 现场补抽 `screenshotCursorIconMetrics` / `forceScreenshotDIBAlphaOpaqueInRect` / `screenshotCursorDrawSize` / `screenshotSystemCursorSize` / `screenshotCursorBaseSizeFromRegistry`；已有 `drawScreenshotCursorInfoOnDC` / `drawScreenshotCursorSnapshotOnDC` / `drawScreenshotCursorSnapshotOnRGBA` / `drawScreenshotCursorOnRGBA` / `drawScreenshotCaptureCursorOnRGBA` / `screenshotGDIScreenCaptureBackend_captureScreenRect`。
- **产出**：11 [S]（`screenshotSystemCursorSize`/`screenshotCursorBaseSizeFromRegistry`/`screenshotCursorDrawSize`/`screenshotCursorIconMetrics`/`forceScreenshotDIBAlphaOpaqueInRect`/`drawScreenshotCursorInfoOnDC`/`drawScreenshotCursorSnapshotOnDC`/`drawScreenshotCursorSnapshotOnRGBA`/`drawScreenshotCursorOnRGBA`/`drawScreenshotCaptureCursorOnRGBA`/`screenshotGDIScreenCaptureBackend.captureScreenRect`）。新增全局 `procGdiFlush`。
- **签名/语义（asm 实证）**：
  - `screenshotSystemCursorSize() (int, int)`：GetSystemMetrics(13/14)。
  - `screenshotCursorBaseSizeFromRegistry() uint64`：`HKCU\Control Panel\Cursors` → `CursorBaseSize`，错误/0/大于 0x200 返回 0。
  - `screenshotCursorDrawSize(w,h int) (int,int)`：max(基准,w,h,系统宽,系统高) 等比缩放并钳制 ≥1。
  - `screenshotCursorIconMetrics(hicon) (*appIconInfo,int,int,int,int,bool)`：加载失败/绘制尺寸无效释放图标。
  - `forceScreenshotDIBAlphaOpaqueInRect(bits,w,h,x0,y0,x1,y1) bool`：交集内 alpha 置 0xFF。
  - `drawScreenshotCursorInfoOnDC(hdc,bits,x0,y0,x1,y1,info,ok) bool`：前置三条件 → 热点缩放 → DrawIconEx(diFlags=3) → alpha 强制。
  - `drawScreenshotCursorSnapshotOnDC`/`drawScreenshotCursorSnapshotOnRGBA`/`drawScreenshotCursorOnRGBA`/`drawScreenshotCaptureCursorOnRGBA`/`captureScreenRect`：全链 asm 直译。
- **注册表字符串 2 个全 rodata 解码**：`0x140c5f97d`(21B "Control Panel\Cursors")/`0x140c507f2`(14B "CursorBaseSize")。
- **测试**：新增 `backend/screenshot_cursor_draw_windows_test.go`（4 用例：alpha 强制/拒绝分支/捕获禁用/尺寸不匹配），全 PASS。
- **门禁**：`go build -tags production -trimpath` / `go vet` / `go test` 三项 EXIT=0，`gofmt -l` 空。
- **artifact**：`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`（19,758,080 B，SHA256 `A7D64D4B816FEC77F1FAD5FE05278C96E3068967D7EA81DC986A21DD7236718D`）。
- **验收**：`docs/acceptance/batch46.md`（G1/G2/G3/G4 四路 PASS）。

**待移交偏差**（批次 46）：
1. GDI 虚拟屏捕获（`screenshotGDIScreenCaptureBackend.captureVirtualScreen` + `captureScreenshotVirtualScreenSnapshotWithOptions`/`captureScreenshotScreenRectWithOptions` + `defaultScreenshotScreenCaptureBackend`/`selectScreenshotScreenCaptureBackendName`）留待批次 47，全部 asm 已 dump 到 pipeline/tmp。
2. `drawScreenshotCursorInfoOnDC` 中 `procDrawIconEx.Call` 复用 `appicon_windows.go` 的 x/sys/windows LazyProc，asm 目标 0x1401935a0 同为 x/sys/windows.LazyProc.Call，ABI 一致。
3. 热点缩放 `int(int32(iconInfo.XHotspot))` 保留 asm 的有符号解释（jle 比较）；实际热点 < 256 无溢出差异。

---

## 20. 批次 47 记录（screenshot GDI 虚拟屏捕获链）

- **目标**：新建 `backend/screenshot_virtual_windows.go`，落地 GDI 虚拟屏捕获 + 暗化预览 + 后端选择名 5 函数。
- **资产**：pipeline/tmp 现场补抽 `captureQRCodeVirtualScreenSnapshotGDI`（0x140971700, 288B）与 `buildDarkenedQRCodeSelectionPreview`（0x14095a200, 352B）；已有 `captureQRCodeVirtualScreenSnapshotGDIWithCursorSnapshot`（0x140971820, 2208B）/`selectScreenshotScreenCaptureBackendName`（0x14096e0a0）/`screenshotGDIScreenCaptureBackend_captureVirtualScreen`（0x14096e660）。
- **产出**：5 [S]（`buildDarkenedQRCodeSelectionPreview`/`captureQRCodeVirtualScreenSnapshotGDIWithCursorSnapshot`/`captureQRCodeVirtualScreenSnapshotGDI`/`selectScreenshotScreenCaptureBackendName`/`screenshotGDIScreenCaptureBackend.captureVirtualScreen`）。
- **签名/语义（asm 实证）**：
  - `buildDarkenedQRCodeSelectionPreview(*image.RGBA) *image.RGBA`：nil → NewRGBA(零矩形)；否则 NewRGBA(img.Rect)+copy(Pix)，RGB 三通道 `uint8((uint16(v*42) * 83887) >> 23)`（42% 亮度）。
  - `captureQRCodeVirtualScreenSnapshotGDIWithCursorSnapshot(screenshotCursorInfo, bool) (qrCodeScreenSnapshot, error)`：bounds 空 / GetDC 失败 / CreateCompatibleDC / CreateDIBSection / SelectObject / BitBlt 六段错误；ok 时 drawScreenshotCursorSnapshotOnDC；buildRGBAFromDIBBits → 暗化预览。
  - `captureQRCodeVirtualScreenSnapshotGDI(bool) qrCodeScreenSnapshot`：按 captureCursor 决定现场捕获或零值。
  - `selectScreenshotScreenCaptureBackendName([]screenshotDisplayCaptureInfo, string, bool) string`："gdi"/"dxgi-hdr" 分支。
  - `captureVirtualScreen(screenshotScreenCaptureOptions) qrCodeScreenSnapshot`：快照变体或无光标捕获。
- **错误字符串 6 个全 rodata 解码**：`0x140c740a7`(33B "未检测到可用的屏幕区域") / `0x140c740e9`(33B "获取屏幕设备上下文失败") / `0x140c7a005`(37B "创建截图设备上下文失败: %w") / `0x140c6bf08`(28B "创建截图位图失败: %w") / `0x140c6bf24`(28B "选择截图位图失败: %w") / `0x140c6bf40`(28B "复制屏幕像素失败: %w")。
- **接口修正**：`types_screenshot.go` `captureVirtualScreen` 去掉 error（asm 两分支均只返回 6 字 snapshot；error 寄存器不参与，调用方也只取 6 字）。`captureScreenRect` 保留 error。
- **格式债务**：`types_screenshot.go` CRLF→LF + package/import 空行 + struct 字段对齐（66 行 gofmt diff，纯格式）。
- **测试**：新增 `backend/screenshot_virtual_windows_test.go`（2 用例：暗化公式/后端选择名 7 分支），全 PASS。
- **门禁**：`go build -tags production -trimpath` / `go vet` / `go test` 三项 EXIT=0，`gofmt -l` 空。
- **artifact**：`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`（19,757,568 B，SHA256 `FCA407D9926F1DC0311F89DAA14FA0D594B2FE1D370524888B54C841DC880BFF`）。
- **验收**：`docs/acceptance/batch47.md`（G1/G2/G3/G4 四路 PASS）。

**待移交偏差**（批次 47）：
1. dxgi/HDR 域（`screenshotDXGIHDRCaptureAvailable`/`resolveScreenshotHDRCaptureMode`/`parseScreenshotHDRCaptureMode`/`enumerateScreenshotDisplayCaptureInfos`/`logScreenshotHDRCaptureBackendDecision`/`screenshotHDRCaptureDebugLog` + `captureScreenshotVirtualScreenSnapshotWithOptions`/`defaultScreenshotScreenCaptureBackend` 收口）留待批次 48，全部 asm 已 dump 到 pipeline/tmp。
2. `defaultScreenshotScreenCaptureBackend`（0x14096de60）依赖 `enumerateScreenshotDisplayCaptureInfos` 与 `screenshotHDRCaptureDebugLog`，需与 dxgi 枚举链同批落地。
3. `captureScreenshotScreenRectWithOptions`（未在本批 dump）依赖 `captureScreenshotVirtualScreenSnapshotWithOptions` 的后端选择结果，需并入批次 48 或 49。

---

## 21. 批次 48 记录（screenshot DXGI/HDR 模式解析与入口探测）

- **目标**：新建 `backend/screenshot_hdr_windows.go`，落地 HDR 模式解析 + 调试开关 + DXGI/D3D11 入口探测 6 函数。
- **资产**：pipeline/tmp 现场补抽 `resolveScreenshotHDRCaptureMode`（0x140971240, 64B）/`parseScreenshotHDRCaptureMode`（0x140971280, 448B）/`parseScreenshotBool`（0x140971620, 224B）/`isScreenshotHDRCaptureDebugEnabled`（0x1409715c0, 96B）+ 其闭包 func1（0x1409f5d00, 96B）/`screenshotHDRCaptureDebugLog`（0x140971440, 288B）/`screenshotDXGIHDRCaptureAvailable`（0x140971160, 224B）。
- **产出**：6 [S]（`parseScreenshotBool`/`parseScreenshotHDRCaptureMode`/`resolveScreenshotHDRCaptureMode`/`isScreenshotHDRCaptureDebugEnabled`/`screenshotHDRCaptureDebugLog`/`screenshotDXGIHDRCaptureAvailable`）。新增全局 `dxgiDLL`/`d3d11DLL`/`procCreateDXGIFactory1`/`procD3D11CreateDevice` + `screenshotHDRCaptureDebugOnce`/`screenshotHDRCaptureDebug`。
- **签名/语义（asm 实证）**：
  - `parseScreenshotBool(string) (bool, bool)`：truthy/falsy/invalid 三分类。
  - `parseScreenshotHDRCaptureMode(string) string`：len 1-8 跳表 + 默认分支。
  - `resolveScreenshotHDRCaptureMode() string`：`Getenv("USBEAM_SCREENSHOT_HDR_CAPTURE")`。
  - `isScreenshotHDRCaptureDebugEnabled() bool`：sync.Once + 闭包 `Getenv("USBEAM_SCREENSHOT_HDR_DEBUG")`。
  - `screenshotHDRCaptureDebugLog(string, ...any)`：开关 + `"[screenshot-hdr] "` 前缀。
  - `screenshotDXGIHDRCaptureAvailable() bool`：`CreateDXGIFactory1`/`D3D11CreateDevice` Find。
- **字符串/符号实证**：`USBEAM_SCREENSHOT_HDR_CAPTURE`@0x140c6d789(29B) / `USBEAM_SCREENSHOT_HDR_DEBUG`@0x140c6a710(27B) / `[screenshot-hdr] `@0x140c574a6(17B) / `DXGI 工厂不可用: %v`@0x140c65520(24B) / `D3D11 设备创建入口不可用: %v`@0x140c79fe0(37B) / `CreateDXGIFactory1`@文件0xC582B6 / `D3D11CreateDevice`@0xC55A0F / `dxgi.dll`@0xC3A834 / `d3d11.dll`@0xC3DE13。
- **测试**：新增 `backend/screenshot_hdr_windows_test.go`（3 用例：布尔解析/HDR 模式解析 19 分支/env 解析），全 PASS。
- **门禁**：`go build -tags production -trimpath` / `go vet` / `go test` 三项 EXIT=0，`gofmt -l` 空。
- **artifact**：`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`（19,759,104 B，SHA256 `224FA934E81FBEC092857AE5755DAB21D0C70D8D273ACDC43E95FFC118B88524`）。
- **验收**：`docs/acceptance/batch48.md`（G1/G2/G3/G4 四路 PASS）。

**待移交偏差**（批次 48）：
1. dxgi 枚举链（`enumerateScreenshotDisplayCaptureInfos`/`formatScreenshotDisplayCaptureInfoForDebug`/`screenshotDXGIColorSpaceDebugName`/`logScreenshotHDRCaptureBackendDecision` + `captureScreenshotVirtualScreenSnapshotWithOptions`/`defaultScreenshotScreenCaptureBackend` 收口）留待批次 49，全部 asm 已 dump 到 pipeline/tmp。
2. `enumerateScreenshotDisplayCaptureInfos`（0x14097e1c0, 1344B）是 dxgi 枚举核心，含 deferwrap，需单独分析。
3. `defaultScreenshotScreenCaptureBackend`（0x14096de60）依赖本批已落地的 `screenshotHDRCaptureDebugLog`/`screenshotDXGIHDRCaptureAvailable`/`selectScreenshotScreenCaptureBackendName`，已具备收口前置条件。

---

## 22. 批次 49 记录（screenshot DXGI 调试格式化链）

- **目标**：新建 `backend/screenshot_dxgi_debug_windows.go`，落地 DXGI 显示器调试格式化 + HDR 后端决策日志 4 函数。
- **资产**：pipeline/tmp 现场补抽 `screenshotDXGIColorSpaceDebugName`（0x140970ec0, 672B）/`screenshotDisplayCaptureLabel`（0x1409706c0, 352B）/`formatScreenshotDisplayCaptureInfoForDebug`（0x140970b60, 864B）/`logScreenshotHDRCaptureBackendDecision`（0x140970820, 832B）。
- **产出**：4 [S]（`screenshotDXGIColorSpaceDebugName`/`screenshotDisplayCaptureLabel`/`formatScreenshotDisplayCaptureInfoForDebug`/`logScreenshotHDRCaptureBackendDecision`）。
- **签名/语义（asm 实证）**：
  - `screenshotDXGIColorSpaceDebugName(uint32) string`：0/12/13/14/16/18/19 七个分支 + UNKNOWN。
  - `screenshotDisplayCaptureLabel(screenshotDisplayCaptureInfo) string`：DeviceName → AdapterName#OutputIndex → adapter=%d output=%d 三分支。
  - `formatScreenshotDisplayCaptureInfoForDebug(screenshotDisplayCaptureInfo) string`：15 参数 fmt.Sprintf。
  - `logScreenshotHDRCaptureBackendDecision(string, string, []screenshotDisplayCaptureInfo)`：开关 + needHDR 首循环 + 汇总/逐条日志。
- **字符串实证**：格式 173B@0x140c95f33 / `%s#%d`@0x140c35bbe / `adapter=%d output=%d`@0x140c5dae3 / `DXGI HDR 后端选择...`@0x140c8ecd9 / `DXGI 显示器: %s`@0x140c5a044 / 8 个颜色空间名。
- **测试**：新增 `backend/screenshot_dxgi_debug_windows_test.go`（3 用例），全 PASS。
- **门禁**：`go build -tags production -trimpath` / `go vet` / `go test` 三项 EXIT=0，`gofmt -l` 空。
- **artifact**：`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`（19,759,616 B，SHA256 `AC5A210562908863780EDFE568B44825774C2458583E26307D3C8AEF9A117FE1`）。
- **验收**：`docs/acceptance/batch49.md`（G1/G2/G3/G4 四路 PASS）。

**待移交偏差**（批次 49）：
1. dxgi COM 薄包装与枚举链（`createDXGIFactory1`/`queryDXGIFactory6`/`releaseDXGIUnknown`/`enumerateScreenshotDXGIAdapterDisplays`/`enumerateScreenshotDisplayCaptureInfos` + `captureScreenshotVirtualScreenSnapshotWithOptions`/`defaultScreenshotScreenCaptureBackend` 收口）留待批次 50，全部 asm 已 dump 到 pipeline/tmp。
2. `createDXGIFactory1`（0x14085da00, 384B）依赖 `procCreateDXGIFactory1.Find` + `SyscallN` + `fmt.Errorf`，是 COM 工厂创建薄包装。
3. `enumerateScreenshotDXGIAdapterDisplays`（0x14097e7c0, 992B）是 dxgi 枚举的适配器层，含 COM vtable 调用。

---

## 23. 批次 50 记录（screenshot DXGI COM 薄包装）

- **目标**：新建 `backend/screenshot_dxgi_com_windows.go`，落地 DXGI 工厂创建 / IDXGIFactory6 查询 / COM 释放 3 函数 + vtable 读取辅助。
- **资产**：pipeline/tmp 现场补抽 `createDXGIFactory1`（0x14085da00, 384B）/`queryDXGIFactory6`（0x14085db80, 416B）/`releaseDXGIUnknown`（0x14085dd20, 128B）。
- **产出**：3 [S] + 1 [S-inline]（`createDXGIFactory1`/`queryDXGIFactory6`/`releaseDXGIUnknown`/`dxgiUnknownVtable`）。新增 IID 常量 `iidIDXGIFactory1`/`iidIDXGIFactory6` + vtable 结构 `dxgiUnknownVtbl`。
- **签名/语义（asm 实证）**：
  - `createDXGIFactory1() (uintptr, error)`：Find → SyscallN(CreateDXGIFactory1) → HRESULT/零值错误三分支。
  - `queryDXGIFactory6(uintptr) (uintptr, error)`：nil 检查 → vtable[0] QueryInterface → HRESULT/零值错误三分支。
  - `releaseDXGIUnknown(uintptr)`：三层零检查 → vtable[2] Release。
- **字符串/GUID 实证**：IID1 `770AAE78-F26F-4DBA-A829-253C83D1B387`@0x141962a10 / IID6 `C1B6694F-FF09-44A9-B03C-77900A0A1D17`@0x141962a20 / `CreateDXGIFactory1 did not return a factory`@0x140c81f53 / `DXGI factory is nil`@0x140c5bbd0 / `IDXGIFactory6 query returned nil`@0x140c724a6 / `IDXGIFactory6 unavailable`@0x140c66b23。
- **vet 处理**：vtable 读取从 `*(*uintptr)(unsafe.Pointer(p))` 改为 `*(**dxgiUnknownVtbl)(unsafe.Pointer(&p))` 类型化读取，消除 unsafeptr 误报。
- **测试**：新增 `backend/screenshot_dxgi_com_windows_test.go`（2 用例：IID 字段断言/nil 安全），全 PASS。
- **门禁**：`go build -tags production -trimpath` / `go vet` / `go test` 三项 EXIT=0，`gofmt -l` 空。
- **artifact**：`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`（19,759,616 B，SHA256 `8CE27D12C6373912E87DAEE96DD27A399EF2564D7818A506DC12003AF00BC3BB`）。
- **验收**：`docs/acceptance/batch50.md`（G1/G2/G3/G4 四路 PASS）。

**待移交偏差**（批次 50）：
1. dxgi 枚举链（`enumerateScreenshotDXGIAdapterDisplays`/`enumerateScreenshotDisplayCaptureInfos` + `screenshotDXGIAdapterName`/`captureScreenshotVirtualScreenSnapshotWithOptions`/`defaultScreenshotScreenCaptureBackend` 收口）留待批次 51，全部 asm 已 dump 到 pipeline/tmp。
2. `enumerateScreenshotDXGIAdapterDisplays`（0x14097e7c0, 992B）与 `enumerateScreenshotDisplayCaptureInfos`（0x14097e1c0, 1344B）是 dxgi 枚举核心，含 COM vtable 调用与 deferwrap。
3. 本批 COM 薄包装已就绪，`defaultScreenshotScreenCaptureBackend`（0x14096de60）依赖链全部满足，可下一批收口。

---

## 24. 批次 51 记录（screenshot DXGI 枚举链）

- **目标**：新建 `backend/screenshot_dxgi_enum_windows.go`，落地 DXGI 适配器/输出枚举链 6 函数 + 3 vtable 读取辅助。
- **资产**：pipeline/tmp 现场补抽 `screenshotDXGIAdapterName`（0x14097eba0, 416B）/`queryScreenshotDXGIOutput6`（0x14097f180, 416B）/`screenshotDisplayCaptureInfoFromDesc`（0x14097f320, 553B）/`screenshotDXGIOutputDisplayInfo`（0x14097ed40, 1024B）/`enumerateScreenshotDXGIAdapterDisplays`（0x14097e7c0, 992B）/`enumerateScreenshotDisplayCaptureInfos`（0x14097e1c0, 1344B）。
- **产出**：6 [S] + 3 [S-inline]（`screenshotDXGIAdapterName`/`queryScreenshotDXGIOutput6`/`screenshotDisplayCaptureInfoFromDesc`/`screenshotDXGIOutputDisplayInfo`/`enumerateScreenshotDXGIAdapterDisplays`/`enumerateScreenshotDisplayCaptureInfos` + `dxgiAdapterVtable`/`dxgiFactoryVtable`/`dxgiOutput6Vtable`）。新增 IID 常量 `iidIDXGIOutput6` + 结构 `dxgiAdapterDesc`（0x130）/`dxgiOutputDesc1`（0x98）/`dxgiFactoryVtbl`；复用 `types_windows.go` 已有 `dxgiAdapterVtbl`/`dxgiOutput6Vtbl`（后者已含 `GetDesc1`@槽 27）。
- **签名/语义（asm 实证）**：
  - `screenshotDXGIAdapterName(uintptr) (string, error)`：nil → `("",nil)`；vtable[8] GetDesc；HRESULT 错误包装；`TrimSpace(UTF16ToString(Description))`。
  - `queryScreenshotDXGIOutput6(uintptr) (uintptr, error)`：nil 检查 → vtable[0] QueryInterface → HRESULT/零值三分支。
  - `screenshotDisplayCaptureInfoFromDesc(dxgiOutputDesc1, int, int, string) screenshotDisplayCaptureInfo`：trim + Bounds min/max 归一化 + HDR∈{12,13,14,16,18,19} + AdvancedColor(CS!=0 || BPC>8)。
  - `screenshotDXGIOutputDisplayInfo(uintptr, int, int, string) (screenshotDisplayCaptureInfo, error)`：QueryInterface → defer Release → vtable[27] GetDesc1 → FromDesc。
  - `enumerateScreenshotDXGIAdapterDisplays(uintptr, int) ([]screenshotDisplayCaptureInfo, error)`：cap 2；EnumOutputs 循环；NOT_FOUND → `(infos,nil)`。
  - `enumerateScreenshotDisplayCaptureInfos() ([]screenshotDisplayCaptureInfo, error)`：factory1 → factory6 → cap 4；EnumAdapters 循环；逐适配器合并 displays。
- **vtable 偏移全链确认**：`QueryInterface@0x00` / `EnumOutputs@0x38` / `GetDesc@0x40` / `EnumAdapters@0x38` / `GetDesc1@0xd8`。关键修正：`IDXGIOutput` 实为 **12** 方法（含 `SetDisplaySurface`），故 `GetDesc1` 在槽 27（0xd8），与 asm `mov rax,[rdx+0xd8]` 精确一致（winapi 0.3.9 手册字段顺序对但人工计数曾漏 `SetDisplaySurface`）。
- **字符串/GUID 实证**：`IID_IDXGIOutput6 = 068346E8-AAEC-4B84-ADD7-137F513F77A1`@0x141962ad0 / `"IDXGIFactory.EnumAdapters"`@0x140c66d30 / `"IDXGIAdapter.EnumOutputs"`@0x140c65550 / `"IDXGIAdapter.GetDesc"`@0x140c5d953 / `"IDXGIOutput6.GetDesc1"`@0x140c5f9a7 / `"IDXGIOutput6 unavailable"`@0x140c65580 / `"DXGI 输出为空"`@0x140c574b7 / `"IDXGIOutput6 查询结果为空"`@0x140c71273。
- **vet 处理**：vtable 读取统一 `*(**T)(unsafe.Pointer(&p))` 类型化；`dxgiAdapterVtbl`/`dxgiOutput6Vtbl` 复用 `types_windows.go` 现有定义（避免 redeclare），`dxgiFactoryVtbl` 本批新增。
- **测试**：新增 `backend/screenshot_dxgi_enum_windows_test.go`（8 用例：IID/两结构布局/FromDesc 语义/AdvancedColor 双分支/nil 安全），全 PASS。
- **门禁**：`go build -tags production -trimpath` / `go vet` / `go test` 三项 EXIT=0，`gofmt -l` 空。
- **artifact**：`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`（19,759,616 B，SHA256 `00D61287D1A30D0EF09264EA8A116E881A1F31506790027F17E17F9C05AC68EF`）。
- **验收**：`docs/acceptance/batch51.md`（G1/G2/G3/G4 四路 PASS）。

**待移交偏差**（批次 51）：
1. dxgi 收口链（`captureScreenshotVirtualScreenSnapshotWithOptions`/`defaultScreenshotScreenCaptureBackend`）留待批次 52，asm 已 dump 到 pipeline/tmp。
2. `captureScreenshotVirtualScreenSnapshotWithOptions`（0x14096e280）与 `defaultScreenshotScreenCaptureBackend`（0x14096de60）是 Screenshot 运行时族的收口入口，本批枚举链已满足其依赖。
3. `enumerateScreenshotDisplayCaptureInfos` 已是 dxgi 枚举的顶层出口，可被 `defaultScreenshotScreenCaptureBackend` 直接调用。

---

## 25. 批次 52 记录（screenshot 后端收口链）

- **目标**：新建 `backend/screenshot_capture_backend_windows.go`，落地默认后端选择 + 按 options 捕获入口 + DXGI 后端接口方法 6 函数 + DXGI 捕获核心入口 2 存根。
- **资产**：pipeline/tmp 现场补抽 `defaultScreenshotScreenCaptureBackend`（0x14096de60, 5739B）/`captureScreenshotVirtualScreenSnapshotWithOptions`（0x14096e280, 4215B）/`captureScreenshotScreenRectWithOptions`（0x14096e420, 192B）/`screenshotDXGIHDRScreenCaptureBackend.captureVirtualScreen`（0x14096e920, 576B）/`captureScreenRect`（0x14096eb40, 512B）/`captureVirtualScreenDXGI`（0x14096ed00, 672B）/`captureScreenRectDXGI`（0x14096ef60, 512B）。
- **产出**：6 [S] + 2 [P]（`defaultScreenshotScreenCaptureBackend`/`captureScreenshotVirtualScreenSnapshotWithOptions`/`captureScreenshotScreenRectWithOptions`/`screenshotDXGIHDRScreenCaptureBackend.name`/`captureVirtualScreen`/`captureScreenRect` + `captureVirtualScreenDXGI`/`captureScreenRectDXGI`）。
- **签名/语义（asm 实证）**：
  - `defaultScreenshotScreenCaptureBackend() screenshotScreenCaptureBackend`：disabled/不可用/枚举失败→GDI；name!="dxgi-hdr"→GDI；否则 `&screenshotDXGIHDRScreenCaptureBackend{displays}`。
  - `captureScreenshotVirtualScreenSnapshotWithOptions(screenshotScreenCaptureOptions) qrCodeScreenSnapshot`：`backend.captureVirtualScreen(options)`。
  - `captureScreenshotScreenRectWithOptions(image.Rectangle, screenshotScreenCaptureOptions) (*image.RGBA, error)`：`backend.captureScreenRect(rect, options)`。
  - `(screenshotDXGIHDRScreenCaptureBackend) name() string`：`"dxgi-hdr"`。
  - `(screenshotDXGIHDRScreenCaptureBackend) captureVirtualScreen(...) qrCodeScreenSnapshot`：DXGI 失败→debug log+GDI 回退。
  - `(screenshotDXGIHDRScreenCaptureBackend) captureScreenRect(...) (*image.RGBA, error)`：DXGI 失败→debug log+GDI 回退；成功→`(img,nil)`。
  - `captureVirtualScreenDXGI`/`captureScreenRectDXGI`：[P] 存根（见下）。
- **值接收者确认**：`screenshotDXGIHDRScreenCaptureBackend` 为 24 字节 slice header；morestack 保存 rax/rbx/rcx 三寄存器（displays ptr/len/cap），options 溢出到栈 `[rsp+0xb0]`；接口 itab 指针包装由 Go 自动生成。
- **字符串实证**：`枚举 DXGI 显示器失败，使用 GDI 后端: %v`@0x140c88f4d(51B) / `启用 DXGI HDR 截图后端: mode=%s displays=%d`@0x140c879f9(49B) / `DXGI 虚拟屏幕捕获失败，回退 GDI: %v`@0x140c8512a(46B) / `DXGI 区域捕获失败，回退 GDI: rect=%v err=%v`@0x140c899e6(52B) / `"disabled"`=`0x64656c6261736964` / `"dxgi-hdr"`=`0x7264682d69677864`。
- **测试**：新增 `backend/screenshot_capture_backend_windows_test.go`（3 用例：dxgi name/disabled→gdi/核心存根返回错误），全 PASS。
- **门禁**：`go build -tags production -trimpath` / `go vet` / `go test` 三项 EXIT=0，`gofmt -l` 空。
- **artifact**：`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`（19,759,616 B，SHA256 `967661CA359640A0472101746B1283EE20DEEE5D271FAF85D457A5969021841C`）。
- **验收**：`docs/acceptance/batch52.md`（G1/G2/G3/G4 四路 PASS）。

**待移交偏差**（批次 52）：
1. DXGI 捕获域（`captureVirtualScreenDXGI`/`captureScreenRectDXGI` 函数体 + 依赖的 `buildScreenshotDXGIVirtualScreenImage`（0x14096f1a0）/`drawScreenshotCaptureCursorOnRGBA`（0x14096f100）/`captureScreenshotDXGIRegionImageWithOptions`/`buildScreenshotDXGIDesktopBoundsImage`/`captureScreenshotDXGIDisplayFrameWithFallback` 系列）留待批次 53，全部 asm 已 dump 到 `docs/goresym/pipeline/tmp/*.asm.txt`。
2. `captureVirtualScreenDXGI`（0x14096ed00, 672B）调用 `qrCodeVirtualScreenBounds`+`buildScreenshotDXGIVirtualScreenImage`+`drawScreenshotCaptureCursorOnRGBA`+`buildDarkenedQRCodeSelectionPreview`，其中两个依赖未落地。
3. `captureScreenRectDXGI`（0x14096ef60, 512B）是 DXGI 矩形捕获核心，依赖 region 捕获链未落地。
4. 截图运行时族剩余：`prewarmScreenshotHDRCaptureAsync`（0x14096e4e0）/`prewarmScreenshotHDRCaptureAsync.func1`（0x14096e580）已 dump，可与 DXGI 捕获域同批或紧随其后。

---

## 26. 批次 53 记录（screenshot DXGI 捕获域）

- **目标**：新建 `backend/screenshot_dxgi_capture_windows.go`，落地 DXGI 捕获域顶层合成/兜底链，把 `captureVirtualScreenDXGI` / `captureScreenRectDXGI` 两个 `[P]` 存根转 `[S]`。
- **资产**：pipeline/tmp 已有 `_b53_buildVSImage` / `_b53_desktopBounds` / `_b53_regionImage` / `_b53_regionFrameFallback` / `_b53_frameFallback` / `_b53_frameFallbackUsing`；本批补抽 `captureScreenshotDXGIOutputFrame`（0x140975fc0, 96B）、`captureScreenshotDXGIDisplayFrameWithFallback.func1`（0x1409f5d60）、`captureScreenshotOverlayMouseInput.func1`（0x1409f5ce0）。
- **产出**：10 `[S]` + 3 `[S-sig]` + 2 `[P]→[S]`（`buildScreenshotDXGIVirtualScreenImage`/`buildScreenshotDXGIDesktopBoundsImage`/`captureScreenshotDXGIRegionImageWithOptions`/`captureScreenshotDXGIDisplayRegionFrameWithFallback`/`captureScreenshotDXGIDisplayGDI`/`captureScreenshotDXGIDisplayFrameWithFallback`/`captureScreenshotDXGIDisplayFrameWithFallbackUsing`/`captureScreenshotDXGIOutputFrame`/`prewarmScreenshotHDRCaptureAsync`/`prewarmScreenshotHDRCaptureWorker`；`captureScreenshotDXGIOutputFrameWithOptions`/`captureScreenshotDXGIOutputFrameRegionWithOptions`/`prewarmScreenshotDXGIOutputFrameCache` 为 `[S-sig]` 桩；`captureVirtualScreenDXGI`/`captureScreenRectDXGI` 转 `[S]`）。
- **签名/语义（asm 实证）**：
  - `captureVirtualScreenDXGI(options) (qrCodeScreenSnapshot, error)`：bounds 空 → `未检测到可用的屏幕区域`；合成 → 叠光标 → 暗化 → 快照。
  - `captureScreenRectDXGI(rect, options) (*image.RGBA, error)`：normalize 空 → `截图区域为空`；region 合成 → 叠光标 → `(img,nil)`。
  - `buildScreenshotDXGIVirtualScreenImage(displays, bounds, capture) (*image.RGBA, error)`：bounds 空 → `DXGI 虚拟屏幕区域为空`；否则尾调用 desktopBounds。
  - `buildScreenshotDXGIDesktopBoundsImage(...)`：bounds 空/capture nil 校验 + 多显示器合成，源偏移 `(inter.Min-display.Min)`、目标偏移 `(inter.Min-bounds.Min)`，`draw.DrawMask(Src)`。
  - `captureScreenshotDXGIRegionImageWithOptions(displays, rect, requireFresh)`：rect 空校验 + 区域合成，目标偏移 `(inter.Min-rect.Min)`，`draw.DrawMask(Src)`。
  - `captureScreenshotDXGIDisplayRegionFrameWithFallback(d, rect, requireFresh)`：`inter=rect∩d.Bounds`；SDR→GDI；HDR→region 捕获失败后 GDI 兜底。
  - `captureScreenshotDXGIDisplayFrameWithFallback(d)`：Using 变体传 `captureScreenshotDXGIOutputFrame` + `captureScreenshotDXGIDisplayGDI`。
  - `captureScreenshotDXGIDisplayFrameWithFallbackUsing(d, dxgiCapture, gdiFallback)`：SDR→gdiFallback；HDR→dxgiCapture 失败后 gdiFallback。
  - `captureScreenshotDXGIOutputFrame(d)`：尾调用 `captureScreenshotDXGIOutputFrameWithOptions(d, false)`。
  - `prewarmScreenshotHDRCaptureAsync()`：disabled/不可用短路；否则 `go` worker(mode)。
  - `prewarmScreenshotHDRCaptureWorker(mode)`：枚举失败 debug；name!="dxgi-hdr" 返回；否则预热缓存。
- **关键偏差**：本批发现 §25 记载「`drawScreenshotCaptureCursorOnRGBA` 未落地」已陈旧——该函数早在批次 46 已是 `[S]` 且 asm 吻合（0x14096f100），本批直接复用，未重复落地。
- **字符串实证**（rip 算术 `lea_addr+7+disp` 修正后）：`未检测到可用的屏幕区域`@0x140c740a7(33B)；`截图区域为空`@0x140c5a032(18B)；`DXGI 虚拟屏幕区域为空`@0x140c6d732(29B)；`DXGI 输出区域为空`@0x140c62fcc(23B)；`DXGI 显示器捕获函数未初始化`@0x140c7b4cc(38B)；`DXGI 未捕获到任何已连接显示器`@0x140c80166(41B)；`DXGI 截图区域为空`@0x140c62fb5(23B)；`DXGI 显示器区域为空`@0x140c6848d(26B)；`GDI 单屏兜底函数未初始化`@0x140c75e9d(34B)；`DXGI HDR 预热枚举显示器失败: %v`@0x140c7f2fa(40B)；其余 12 条 fmt/debug 格式串与 VA 见 acceptance/batch53.md §G3。
- **测试**：改写 `backend/screenshot_capture_backend_windows_test.go`（6 用例全确定性，含 DrawMask 合成像素断言），全 PASS。
- **门禁**：`go build -tags production -trimpath ./backend` / `go vet ./backend` / `go test -count=1 -p=1` 三项 EXIT=0。
- **artifact**：`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`（19,759,616 B，SHA256 `F3D0D9B809A46C2C88795DB0231A121606A2E3C2D86E671C97E11C69452A1A6F`）。
- **验收**：`docs/acceptance/batch53.md`（G1/G2/G3/G4 四路 PASS）。

**待移交偏差**（批次 53 → 54 已收口）：

1. ~~深层 D3D11/桌面复制链~~ → 批次 54 已落地纯 Go 缓存/入口层 + 方法族；深层体（D3D11 设备/桌面复制/tone-map/旋转/纹理换算）仍留待批次 55（`[S-sig]` 桩）。
2. ~~`bootstrapservice_state_deps.go` 占位类型~~ → 批次 54 已删除占位，真类型 `screenshotDXGIOutputCaptureCache` 落地到 `screenshot_dxgi_cache_windows.go`（全局单例改名 `screenshotDXGIOutputCaptureCacheGlobal`，与类型名区分）。
3. ~~3 个 `[S-sig]` 桩返回 pending error~~ → 批次 54 已转 `[S]`（三入口 + 缓存链完整语义）。

## 27. 批次 54 记录（screenshot DXGI 输出复制缓存域）

- **目标**：新建 `backend/screenshot_dxgi_cache_windows.go`（纯 Go 缓存/缓存项方法族 + 缓存键 + 超时哨兵 + 三入口 + 像素克隆）与 `backend/screenshot_dxgi_deep_windows.go`（10 个深层 `[S-sig]` 桩）；`screenshot_dxgi_com_windows.go` 追加 `screenshotCOMRelease`/`screenshotCOMQueryInterface`；删除 `bootstrapservice_state_deps.go` 占位类型。
- **资产**：pipeline/tmp `b54_*.asm.txt`（entry 方法族、cache 方法族、COM helpers、cloneRGBAImage/cloneRGBARegion、三个入口、orient/texture-rect）。
- **产出**：19 `[S]`（cache 文件）+ 10 `[S-sig]`（deep 文件）+ 2 `[S]`（com 追加）；3 个入口从 `[S-sig]` 升 `[S]`；删除 3 个旧 `[S-sig]` 桩 + 1 个占位 Close。
- **关键实证修正**（批次 53 前的误记，本批按 asm 定论）：
  - 缓存重试语义：`test al; je` 证明**非超时哨兵**错误才 `remove` + 重建重试；超时哨兵直接 `(nil, err)` 透传（正常"无桌面更新"）。
  - 区域边界检查：`rect.Intersect(image.Rect(0,0,Dx,Dy)) != rect` → `DXGI 输出区域越界: %v`（不是 `Intersect(d.Bounds)`）。
  - `updateLastFrameLocked` 条件：`img==nil || len(img.Pix)==0`（asm 读 `[img+8]`=Pix.len，非 Pix 指针）。
  - 类型/全局同名冲突：Go 不允许类型与变量同名，全局单例定名 `screenshotDXGIOutputCaptureCacheGlobal`（VA 0x140bc1c98，Shutdown 0x140773e8c `mov rax,[rip+0x144de05]`）。
- **字符串实证**：20 条错误/日志串（`DXGI 捕获帧超时`20B / `DXGI 输出复制缓存未初始化`35B / `DXGI 输出复制缓存已关闭`32B / `DXGI 输出复制缓存项为空`32B / `DXGI 输出复制缓存项已关闭`35B / `DXGI 输出未连接到桌面`29B / `DXGI 输出区域为空`23B / `DXGI 输出区域越界: %v`27B / `不支持的 DXGI 显示器旋转值 %d`39B / `DXGI 旋转后帧尺寸不匹配: frame=%v display=%v rotation=%d`65B / `DXGI 旋转后区域尺寸不匹配: frame=%v rect=%v`53B / `DXGI 输出复制对象预热失败: display=%s err=%v`54B / `DXGI 缓存捕获失败，重建输出复制对象: display=%s err=%v`69B / `DXGI 捕获失败: %v；重建输出复制对象失败: %w`58B / `DXGI 缓存区域捕获失败，重建输出复制对象: display=%s rect=%v err=%v`83B / `DXGI 区域捕获失败: %v；重建输出复制对象失败: %w`64B / `DXGI 本次无桌面更新，复用上一帧: display=%s`56B / `DXGI 区域捕获本次无桌面更新，复用上一帧: display=%s rect=%v`76B / `DXGI HDR tone mapping: display=%s options=%s`44B / `DXGI 输出复制对象已缓存: display=%s`44B / `COM 对象或 IID 为空`24B / `COM QueryInterface 结果为空`31B / `DXGI 帧为空`20B）详见 acceptance/54.md §G3。
- **测试**：新建 `backend/screenshot_dxgi_cache_windows_test.go`（12 用例全确定性），全 PASS。
- **门禁**：`go build -tags production -trimpath ./backend` / `go vet ./backend` / `go test -count=1 -p=1 ./backend` 三项 EXIT=0。
- **artifact**：`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`（19,760,128 B，SHA256 `1071F2FF61F645D374D5AD5F25A6FEE46AD614F4F74438EEB64DE83FA361BF6C`）。
- **验收**：`docs/acceptance/54.md`（G1/G2/G3/G4 四路 PASS）。

**待移交偏差**（批次 54）：

1. 10 个深层 `[S-sig]` 桩留待批次 55：`openScreenshotDXGIOutputTarget` / `createScreenshotD3D11Device` / `createScreenshotD3D11DeviceWithFeatureLevels` / `(*screenshotD3D11Device).release` / `duplicateScreenshotDXGIOutput` / `captureScreenshotDXGIFrameRegion` / `screenshotHDRToneMapOptionsForDisplayWithSDRWhiteResolver` / `formatScreenshotHDRToneMapOptionsForDebug` / `screenshotDXGITextureRectForDesktopRect` / `orientScreenshotDXGIFrame`（rotation 2/3/4 体）。
2. `screenshotDisplayConfigResolverGlobal` 占位 `uintptr`（目标全局 0x141096c38），字段布局待批次 55。
3. `orientScreenshotDXGIFrame` 仅 rotation==1 恒等路径 + 0/>4 报错；rotation 2/3/4 像素旋转待批次 55。

> 以上 3 项已于批次 55 全部收口（见 §28）。

## 28. 批次 55 记录（screenshot DXGI 深层 D3D11/桌面复制/tone-map/旋转）

- **目标**：`backend/screenshot_dxgi_deep_windows.go` 的 10 个 `[S-sig]` 桩全量 asm 直译转 `[S]`，连带 `screenshotDisplayConfigResolverGlobal` 占位 `uintptr` → `func(screenshotDisplayCaptureInfo) (float32, bool)` 及全部依赖 helper（float16/sRGB LUT、tone-map 数学、DisplayConfig 解析链、旋转/纹理换算）。
- **资产**：pipeline/tmp `b54_*.asm.txt`（10 桩 + captureOnce + orient + openTarget 闭包链）、`b55_*.asm.txt`（tone-map / display-config 解析链）。
- **产出**：deep 文件 45 函数（31 `[S]` + 4 `[S-inline]` + 10 `[S-sig]→[S]`）；`screenshotDisplayConfigResolverGlobal` 改型为 `func(screenshotDisplayCaptureInfo) (float32, bool)`（二进制恒 nil，cache 调用点 117 行透传不变）；`screenshot_dxgi_cache_windows_test.go` 修正 1 处断言字符串。
- **关键实证修正（本批按 asm 定论）**：
  - `orientScreenshotDXGIFrame` nil/空帧错误串是 **`DXGI 捕获帧为空`（20B，str @0x140c5db0b）**，非批次 54 误记的 `DXGI 帧为空`（14B）；已实测二进制字节 `44 58 47 49 20 e6 8d 95 e8 8e b7 e5 b8 a7 e4 b8 ba e7 a9 ba`。
  - 重试日志（超时/黑帧两路）**共用同一格式串** `DXGI 捕获疑似黑帧，准备重试: attempt=%d/%d`（0x140c8a2dc，超时 0x14097aaf3 与黑帧 0x14097aa7d 两处 rip 算术同落一址，编译器去重）。
  - tone-map `epsilon` 常量是 **1.0**（0x1411cd490），非 0.001；m=1.0 恒成立（退化 `(1+peak)/(1+peak)`）；refWhite=clamp(0.5*(1/peak²+1),0.001,1.0)；sdrWhiteOutput 夹上界 1.0（默认 0.55）。
  - `queryScreenshotDisplayConfigSDRWhiteLevelForDisplay`：sdrWhite=raw/1000.0（除数 @0x1411cd4b0），nits=sdrWhite×**80.0**（乘数 @0x1411cd4a8，仅日志）；raw==0 或 sdrWhite≤0/NaN → `忽略无效` 并返回 `(0,false)`。
  - `openScreenshotDXGIOutputTarget` 首返回值是 **adapter**（EnumAdapters 结果，保活供 D3D11CreateDevice），非原始 IDXGIOutput；原始 output 在 queryOutput6 后立即释放。
- **D3D11 vtable 槽位（实测）**：CreateTexture2D@0x28(槽5)、Map@0x70(槽14)/Unmap@0x78(槽15)、CopySubresourceRegion@0x170(槽46)/CopyResource@0x178(槽47)、GetDesc@0x50(槽10)；DuplicateOutput1@0xd0(槽26)；AcquireNextFrame@0x40(槽8)/ReleaseFrame@0x70(槽14)。
- **DisplayConfig 解析链（死代码，解析器 nil，忠实复现）**：user32.dll `GetDisplayConfigBufferSizes`/`QueryDisplayConfig`/`DisplayConfigGetDeviceInfo`；source 名 Type=1 Size=0x54、SDR 白点 Type=11 Size=0x18；活动路径 ERROR_INSUFFICIENT_BUFFER(122) 重试 4 次。
- **测试**：`screenshot_dxgi_cache_windows_test.go` 修正 `TestOrientationRotation1Identity` nil 断言（`DXGI 帧为空`→`DXGI 捕获帧为空`），12 用例全 PASS。
- **门禁**：`go build -tags production -trimpath ./backend` / `go vet ./backend` / `go test -count=1 -p=1 ./backend` 三项 EXIT=0。
- **artifact**：`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`（19,760,128 B，SHA256 `6C1D17B0214F329DFB98A10170CE0BA9D30DE49D0F709C0058B17772DBC689AC`）。
- **验收**：`docs/acceptance/55.md`（G1/G2/G3/G4 四路 PASS）。

**待移交偏差**（批次 55）：无。DXGI 截图深层域（openTarget/设备/桌面复制/帧捕获/tone-map/旋转/纹理换算/DisplayConfig）已全部落地，`screenshotDisplayConfigResolverGlobal` 二进制恒 nil（无 setter），DisplayConfig 解析链为忠实死代码。

## 29. 批次 56 记录（S-sig 补体 + 资产就绪度实测校准）

- **目标**：按推荐补体 `windowmanagement.go` 的 [S-sig]，中途实测校准后转向「有 asm 资产但仍是 [P]/[S-sig]」的真实低垂果实。
- **产出**：4 个 [S-sig]→[S]（`upsertWindowFullscreenSnapshot` / `removeWindowFullscreenSnapshot` / `ExtractLauncherConfigIcons` / `ExternalizeLauncherConfigIcons`），`S-sig` 102→98，`S` 742→746，其余口径不变。
- **关键实证修正（本批 asm 定论）**：
  - `upsertWindowFullscreenSnapshot`（0x1409e85a0）**真实签名是 2 参数** `(snapshots []WindowFullscreenSnapshot, snap WindowFullscreenSnapshot)`，非之前 3 参数 `(snapshots, hwnd, snap)`。依据：morestack 仅存 rax/rbx/rcx 3 寄存器 → snap 全在栈，`rdi = [rsp+0x2b0]` = snap.HWND（首个栈参数）。语义：`snap.HWND==0` → 内联去重（跳过 HWND==0 + seen map，等价 `remove(snapshots,0)`）；`!=0` → `remove(snapshots, snap.HWND)` + append。
  - `removeWindowFullscreenSnapshot`（0x1409e8d00）两遍实现（第一遍过滤 hwnd、第二遍 seen 去重 + 跳过 0），一遍等价实现语义一致，已升 [S]。
  - `ExtractLauncherConfigIcons`（0x140893700）/`ExternalizeLauncherConfigIcons`（0x140893740）均 64B 薄透传（xor edi / mov edi,1 → call externalizeLauncherConfigIcons），已升 [S]。
- **资产就绪度实测校准（纠正 §6 两处失实）**：
  - `twofactor` 域 17 个 asm 资产（13 个 `BootstrapService` 入口 + 4 个 attach）**已全部 [S]**（bootstrapservice.go 1225-1747 / twofactor_helpers.go）；剩余 twoFactorService 核心（SetupPassword 730L / ChangePassword 625L / ExportEntries 1315L 等 ~90 函数）**无 asm 资产**（仅 source_funcs 行号），需现场 dump。§6 表格「twofactor 18 asm」系历史 asm 数，非可补体缺口。
  - `windowmanagement.go` 22 个 [S-sig] 中 17+ 依赖未落地的 `windowmanagement_windows.go` 平台层（GetInfo 915L / SetTopMost 753L / SetResolution 603L / CaptureSnapshot 540L / EnterBorderless 542L / EnterFullscreen 502L / RestoreSnapshot 465L / ValidateTarget 700L），该 8 平台函数在 source_funcs 仅行号、无反汇编。剩余 normalize×2 + UnmarshalJSON 亦仅行号蓝图。
  - **真实低垂果实（交叉分析：pipeline/tmp 677 asm × backend [P]/[S-sig]）**：`launcherconfigiconstore` 域 6 个（`launcherConfigIconDataFingerprint` / `ExtractLauncherConfigIcons`✓ / `ExternalizeLauncherConfigIcons`✓ / `loadUnlocked` / `writeUnlocked` / `validateLauncherConfigIconLibrary`）+ `bootstrapservice` 域 15 个（`resolveBookmarkPageTitleWithNetwork` / `loadLinkPreferences` / `startDetachedCommand` / `PersistSnapshotsForShutdown` / `configStoreSnapshot` / `syncHotkeyBindingsFromConfig` / `syncRuntimeServicesAfterSave` / `attachLauncherBackgroundURL` / `syncFileSearchRuntime` / `syncCursorWrapHotCornerGuard` / `defaultTagCatalogFromLanguageMessages` / `CompleteInitialization` / `AbortInitialization` / `GetWindowManagementState` / `executeMouseGestureLauncherFeatureAction`）+ `main.go` 1 个（`initializeWithCheckpoint`）。
- **门禁**：`go build -tags production -trimpath ./backend` / `go vet ./backend` / `go test -count=1 -p=1 ./backend` 三项 EXIT=0。

**待移交偏差**（批次 56）：下一批推荐从 §6 修正为「launcherconfigiconstore 剩余 4 个（loadUnlocked/writeUnlocked/validateLauncherConfigIconLibrary/launcherConfigIconDataFingerprint，有 asm，阻断点主要是 .rdata 字符串常量待 decode）」，其次 bootstrapservice 15 个有 asm 的 [P]/[S-sig]。

### 批次 56 续（launcherconfigiconstore 域 3 签名修正 + 还原，S-sig 95）

- **产出**：`loadUnlocked` / `writeUnlocked` / `validateLauncherConfigIconLibrary` 三个 [S-sig]→[S]，`S-sig` 98→95，`S` 746→749。`launcherConfigIconDataFingerprint` 已 [S]（体含 [P] 标注，不影响计数），无需改动。
- **关键签名修正（本批 asm 定论，均为当前重建代码签名错误）**：
  - `loadUnlocked`（0x140894fa0, 108L）真实签名 **4 值** `(int, map[string]launcherConfigIconRecord, *launcherConfigIconStoreBudget, error)`，非单值 error。依据：成功路径 rax=[s+0x38]=Version、rbx=副本 map、rcx=newobject(24B budget 快照)、rdi/rsi=0；错误路径 rax=1、rbx=makemap_small 空 map、rcx=0。体：ensureLoadedUnlocked → makemap(hint) + mapiter 逐项拷贝 Icons + newobject 快照 cachedBudget 三字段（0x48/0x50/0x58）。3 处调用点（Put/externalize/Prune）改为 4 值接收。
  - `writeUnlocked`（0x140895a80, 47L）真实签名 `(lib launcherConfigIconLibrary) error`，非无参方法 `() error`。依据：morestack 存 3 寄存器（s, lib.Version, lib.Icons）；错误消息解码 = `图标库路径不能为空`(27B)。体：validate(lib)→(budget,err) → writeLibrary（nil 回退 writeLauncherConfigIconLibraryFile）→ 成功后 s.cached=lib、s.cachedBudget=*budget、loaded=true、cachedLoadErr=nil。3 处调用点改为 `writeUnlocked(s.cached)`。
  - `validateLauncherConfigIconLibrary`（0x140895e80, 39L）真实签名 `(*launcherConfigIconStoreBudget, error)`，非单值 error。依据：返回 (rax=budget 指针, rbx=error.itab, rcx=error.data)。体：Version!=1 → count>8192 → sort.Strings(refs) → 逐项 normalize 一致性 / ContentType TrimSpace+ToLower 一致 / Data TrimSpace 非空 / budget.add(`icons[<ref>(`, `data:<ct>;base64,<data>`) / newRef==ref。拼接常量解码：`icons[`(6B) `(`(1B) `data:`(5B) `;base64,`(8B)。
- **门禁**：build / vet / test 三项 EXIT=0。

### 批次 56 续二（bootstrapservice 域现实校准 + startDetachedCommand，S-sig 94）

- **产出**：`startDetachedCommand` [S-sig]→[S]，`S-sig` 95→94，`S` 749→750。
- **关键签名修正**：`startDetachedCommand`（0x1408a8480, 256B）真实签名 `([]string) error`，非 `(string, bool) error`。依据：morestack 仅存 3 寄存器（slice ptr/len/cap）。体：len==0 → `命令不能为空`(18B)；`exec.Command(args[0], args[1:]...)` → `Cmd.Start()` → 失败 `fmt.Errorf("启动失败: %w", err)`(16B)。调用点 `RestartApplication` 改传 `[]string{exe}`。
- **bootstrapservice 域 15 个实测校准（纠正 goal 交叉分析失实）**：
  - **已 [S]（无需补体，5 个）**：`defaultTagCatalogFromLanguageMessages` / `CompleteInitialization` / `AbortInitialization` / `GetWindowManagementState` / `executeMouseGestureLauncherFeatureAction`。其注释块顶部均 [S 汇编实证]，仅体尾 [P] 标注（count_funcs.sh 取注释块首个 marker，故计 [S]）。
  - **依赖未落地域（无法独立直译，9 个）**：`resolveBookmarkPageTitleWithNetwork`（bookmarks.go）、`loadLinkPreferences`（linkbrowser*.go）、`PersistSnapshotsForShutdown`（snapshotMirror/saveSnapshots）、`configStoreSnapshot`（desktopwidget 2352B 文档）、`syncHotkeyBindingsFromConfig`（依赖 syncHotkeyBindings@0x140791400 复杂链 + 9 字段 TrimSpace/map 构建）、`syncRuntimeServicesAfterSave`（runtime 域）、`attachLauncherBackgroundURL`（asset 背景域 387 行）、`syncFileSearchRuntime`（shouldWarmResidentFileSearchRuntime 未落地）、`syncCursorWrapHotCornerGuard`（normalizeHotCornerConfig/hasEnabledHotCornerRule 未落地）。
  - 上述 9 个的 [P]/[S-sig] 根因是**依赖未落地 helper/域**，非"有 asm 即可直译"，需先落地各自依赖域。
- **门禁**：build / vet / test 三项 EXIT=0。

### 批次 57（热键/文件搜索/热角三签名补体）

**基线**：`FUNCS=1118 / S=750 / S-sig=94 / P=72 / UNMARKED=189`，SHA256 `D513D0D168A70FF9B2133398D1704B3D12DD0DDE0D39B581971C30F3A4727C5F`。
**收口**：`FUNCS=1127 / S=759 / S-sig=92 / P=74 / UNMARKED=189`，SHA256 `41D12D445069AF29B4A791BCAB2FC3F1793B95050A5BCCC4E2E1FC4A3D2BEDA2`（`artifacts/UsbEAm_Launcher_rebuilt.exe`）。

**本批落地（续三 1–3 项）**：

1. **`syncHotkeyBindingsFromConfig` 链（[S-sig]→[S]）**
   - 实证 `syncHotkeyBindings@0x140791400` 真实签名：`(bs *BootstrapService, bindings launcherHotkeyBindings)`——结构体按 ABI 展平为 22 标量（9 寄存器 + 13 栈槽），非无参/多标量。
   - 修正 `launcherHotkeyBindings` 结构字段顺序：`ScreenshotFeatureEnabled` 移至末位（7 组 {string,bool} + 末位 bool），对齐 asm 展平序。
   - 新增 `launcherFeatureModulesFromConfig`（18 键全默认 true + `Preferences.FeatureModules` 合并，经 `.data` 解码实证两轮同一全局切片）+ `launcherHotkeyBindingsRawFromConfig`（7 hotkey 字符串：SummonSearch 原样透传不 Trim，其余 6 个 TrimSpace 仅空判定；首 *bool nil→true，其余 6 个 nil→TrimSpace 非空；`enabled["screenshot"]` 作 ScreenshotFeatureEnabled）+ `boolOrTrimmedNonEmpty`。
   - `launcherHotkeyBindingsFromConfig` = raw + normalize；`syncHotkeyBindingsFromConfig` = `bs.syncHotkeyBindings(raw)`（normalize 在 syncHotkeyBindings 内部）。
   - `syncHotkeyBindings()` 无参 → `syncHotkeyBindings(bindings launcherHotkeyBindings)`；调用点：attachApp/setHotkeyCaptureLease 传 `bs.bindings`（缓存态），GetState 传 `launcherHotkeyBindingsRawFromConfig(cfg)`。
   - `normalizeLauncherHotkeyBindings` 修正为 `[S-sig 0x140790ee0]`（原误标 `[S 0x14077e0a0]`）：体依赖 `normalizeLauncherHotkeyBindingValue@0x14087f020`（1255 行，未落地），暂恒等。
   - **顺带修正** `syncRuntimeServices` 默认启用表：原混合 true/false → 全 true + FeatureModules 合并（复用 `launcherFeatureModulesFromConfig`；asm 0x14079f124 首轮 `mapassign` 恒写 1 实证）。

2. **`syncFileSearchRuntime`（[S-sig]→[S]）**
   - 签名 `()` → `(enabled map[string]bool, cfg LauncherConfig)`（cfg 0x938 字节栈传值，非指针）。
   - 新增 `shouldWarmResidentFileSearchRuntime`：`enabled` 非 nil 且 `!enabled["files"]`→false；`!cfg.Enabled`→false；`cfg.ResourceMode != "resident"`（len 8 常量）→false。
   - 体：warm 判定 → 锁内 `!warm` 清 pending / `warm && idx==nil` 置 pending（`warm && idx!=nil` 保持）→ 解锁后 `warm && idx!=nil` 则 `idx.scheduleFileSearchMaintenance(nil, true)`。
   - **阻断残留**：`normalizeFileSearchConfig@0x140879e00` 未落地，暂视为恒等直接读 `ResourceMode`（LegacyMode→ResourceMode 映射待该域专项）。

3. **`syncCursorWrapHotCornerGuard`（[S-sig]→[S]）**
   - 签名 `(features)` → `(features map[string]bool, hotCorner HotCornerConfig)`（hotCorner 取自 `cfg.MouseGestures.HotCorners`）。
   - 新增 `normalizeHotCornerConfig`：TriggerSizePx 0→3/<0→1/>64→64；TriggerDelayMs <0→0/>10000→10000；CooldownMs 0→300/<0→0/>60000→60000；AutoDisableFullscreen = incoming || isZero(incoming)；Corners 走 normalizeHotCornerRules。
   - 新增 `normalizeHotCornerRules`：默认 4 角基底（topLeft/topRight/bottomLeft/bottomRight；bottom 两角默认 Win/Win+D）→ 入参逐项合并（normalizeHotCorner 空则跳过）→ 按默认顺序重建 len=4。
   - 新增 `hasEnabledHotCornerRule`：normalize 后遍历 Corners，任一 `Enabled && TrimSpace(Hotkey) != ""`。
   - 体：`features["mouseGestures"]`（len 13 常量实证）且 normalized.Enabled 且 hasEnabledHotCornerRule → `SetCursorWrapCornerGuard(TriggerSizePx)`，否则 0。
   - **阻断残留**：`normalizeHotCorner@0x1408dc520` + `normalizeMouseGestureHotkey@0x1408dc800` 未落地，暂恒等（角名/热键规范化待鼠标手势域专项）。

**门禁**：build / vet / test 三项 EXIT=0；`bash build.sh` 完成。

### 批次 58（热键/热角/文件搜索三深层 helper 补体）

**基线**：`FUNCS=1127 / S=759 / S-sig=92 / P=74 / UNMARKED=189`，SHA256 `41D12D445069AF29B4A791BCAB2FC3F1793B95050A5BCCC4E2E1FC4A3D2BEDA2`。
**收口**：`FUNCS=1136 / S=765 / S-sig=91 / P=78 / UNMARKED=189`，SHA256 `9A47DAD5855EE842705712183AFD14F88D2F23760C39138C4500F75D59595301`（`artifacts/UsbEAm_Launcher_rebuilt.exe`）。

**本批落地（续三 1–3 项全部闭合）**：

1. **`normalizeLauncherHotkeyBindingValue` [S]（0x14087f020, 0x140）+ `normalizeLauncherHotkeyBindings` [S-sig]→[S]**
   - 新增包级静态谓词 `launcherHotkeyBindingPredicate`（asm 0x1410969a0 = 16B 双函数值 `{canonicalHotkeyKey, canonicalSearchCategoryShortcutKey}`；`normalizeShortcutBindingWithOptions` 只取偏移 0 的 canonicalHotkeyKey bool 语义）。
   - helper 语义：`normalizeShortcutBindingWithOptions(value, pred, false, false, false)`；flag2 时重算 `(value, pred, true, false, false)`；空→""；!flag1→result；set==nil→result；set 已含→""；否则插入并返回。
   - 父函数：SummonSearch 空则回退 `"Ctrl+Alt+S"`（len 0xa 常量 0x140c44077 解码）；构建 `map[string]struct{}` 去重集，`SummonSearchEnabled && 非空` 写入；SummonOnly 以 `(SummonOnlyEnabled, false)` 调 helper，5 组截图热键以 `(ScreenshotFeatureEnabled && XEnabled, true)` 调 helper；Enabled 各标志原样透传。
   - 修正 HANDOFF 批 57 误记「1255 行」——真实体量 0x140（320B，104 asm 行）；1255 系 source_funcs.txt 行号转写错误。
2. **`normalizeFileSearchConfig` [S]（0x140879e00, 0x340）+ `normalizeFileSearchResourceMode` [S]（0x14087a5c0, 0x1e0）**
   - 闭合 LegacyMode→ResourceMode 映射：`hot/fast/resident/performance`→`resident`；`cold/memory/low-memory/low_memory/mem-saver/mem_saver`→`memory-saver`；其余（含 balanced）→`balanced`；空 ResourceMode 回退 LegacyMode（TrimSpace+ToLower）。
   - 新增 `fileSearchConfigNormalized`（10 字段，无 Legacy*；asm 实证输出偏移 Enabled 0x00/ResourceMode 0x68/FileTypeFilters 0x78）。
   - config 体：Volumes 空则回退 LegacyRoots；MaxResults `!=200` 钳 200（cmovne 实证恒 200）；RecentItemsEnabled nil→true；RecentItemsLimit `<=0→24 / <10→10 / >128→128`；6 个 helper 未落地标 [P]。
   - `shouldWarmResidentFileSearchRuntime` 改调 `normalizeFileSearchConfig`（asm 0x14079f9a0 实证：`enabled["files"]`→`normalized.Enabled`→`ResourceMode=="resident"` 三重判定）。
3. **`normalizeHotCorner` [P]→[S]（0x1408dc520, 0x2e0）+ `normalizeMouseGestureHotkey` [P]→[S]（0x1408dc800, 0xe0）**
   - 角名：TrimSpace→ToLower 后按长度分派字节比较，接受 4 角名的缩写/连字符/词序颠倒共 19 变体 → camelCase；未命中 ""。
   - 热键：TrimSpace→空""；`EqualFold(trimmed,"win"/"windows")`→`"Win"`；否则 `normalizeShortcutBindingWithOptions(trimmed, pred, true, true, true)`。
   - `normalizeHotCornerRules` 两处 [P] 引用随之转 [S]。

**门禁**：build / vet / test 三项 EXIT=0；新增 `hotkey_normalize_test.go`（6 测试函数锁定新 [S] 行为）。

### 批次 59（configStoreSnapshot 签名实证修正 + matchesPath 落地）

**基线**：`FUNCS=1136 / S=765 / S-sig=91 / P=78 / UNMARKED=189`，SHA256 `9A47DAD5855EE842705712183AFD14F88D2F23760C39138C4500F75D59595301`。
**收口**：`FUNCS=1137 / S=767 / S-sig=91 / P=77 / UNMARKED=189`，SHA256 `5070E271466330C1971DCAB616F21F63C7DD695B2C84C23F106C6713088F1292`（`artifacts/UsbEAm_Launcher_rebuilt.exe`）。

**本批落地（§29 剩余 6 项第 1 项闭合）**：

1. **`configStoreSnapshot`（[P]→[S]，签名证伪修正）**
   - 真 ABI 实证：`(*launcherConfigStore, WorkspaceLayout)`，非旧骨架误写的三值
     `(*launcherConfigStore, LauncherConfig, DesktopWidgetDocument)`。第二栈返回值仅 160B
     （`duffzero+0x134` 与 `duffcopy+0x2f4` 均为 160B = WorkspaceLayout），不含 LauncherConfig
     （2576B=0x142 qword）与 DesktopWidgetDocument（2352B=0x126 qword）。
   - 体：nil guard → lock `bs.lock(+0x540)` → defer Unlock → 快照 `bs.workspace` →
     `store = bs.configStore(+0xa0)` → `!store.matchesPath(ConfigFile)` 则 `launcherConfigStoreForPath`
     重建并回写（含 GC 写屏障）→ 返回 `(store, workspace快照)`。
   - 新增 helper `launcherConfigStore.matchesPath@0x140898ca0` [S]：nil→false；
     `PathKey(s.path)==PathKey(path)`（长度 + memequal）。
   - 连带修正 7 处调用点（cfg/widgets 改由 `store.Read()` / `store.ReadSelfContained()` 承载，
     标注 [S-sig]；`CompleteInitialization` asm 不调用 configStoreSnapshot，cfg 实为栈参，[P] 留待专项）。

**门禁**：build / vet / test 三项 EXIT=0；`bash build.sh` 完成。

### 批次 60（bookmark 标题解析链 6 函数落地）

**基线**：`FUNCS=1137 / S=767 / S-sig=91 / P=77 / UNMARKED=189`，SHA256 `5070E271466330C1971DCAB616F21F63C7DD695B2C84C23F106C6713088F1292`。
**收口**：`FUNCS=1142 / S=773 / S-sig=90 / P=77 / UNMARKED=189`，SHA256 `5D29655EA095FDBFE7DFE97CA35A1439EFDA20F8E3D0E6F2F68D608BCE48B1B2`（`artifacts/UsbEAm_Launcher_rebuilt.exe`）。

**本批落地（§29 剩余项第 2 项闭合）**：

1. **bookmarks.go 依赖域落地**：`normalizeBookmarkPageTitle`(0x140771cc0) /
   `normalizeBookmarkPageTitleURL`(0x140770ec0) / `bookmarkPageTitleCandidates`(0x140770b40) /
   `extractBookmarkPageTitle`(0x140771820) / `fetchBookmarkPageTitle`(0x140771000) 全 [S]。
2. **`resolveBookmarkPageTitleWithNetwork`（[S-sig]→[S]）**：签名证伪为
   `(access LauncherNetworkAccess, url string) (string, error)`（首参为网络访问接口，非 WorkspaceLayout）。
3. **调用点纠正**：`ResolveBookmarkTitle` 改传 `newLauncherNetworkAccess(ws.ConfigFile)` 的返回值
   `access`，而非 `ws`。
4. **新增依赖**：`golang.org/x/net v0.56.0`（与目标 exe buildinfo dep 版本一致）。

**门禁**：build / vet / test 三项 EXIT=0；`bash build.sh` 完成。

### 批次 61（launcherConfigOptions locale 签名校正 + loadLinkPreferences 依赖勘察）

**基线**：`FUNCS=1142 / S=773 / S-sig=90 / P=77 / UNMARKED=189`，SHA256 `5D29655EA095FDBFE7DFE97CA35A1439EFDA20F8E3D0E6F2F68D608BCE48B1B2`。
**收口**：`FUNCS=1142 / S=773 / S-sig=90 / P=77 / UNMARKED=189`（计数持平——仅签名/体校正，无档位迁移），SHA256 `B4031691AD878D86AC4D5B158B691141778EE87A9C983C06BD2003B85CF6709E`（`artifacts/UsbEAm_Launcher_rebuilt.exe`）。

**本批落地（§29 剩余项第 1 项 `loadLinkPreferences` 的前置依赖签名校正）**：

1. **`launcherConfigOptions(locale string)` 签名校正**：asm 0x140775660 实证双参（bs + locale），
   旧无参形式系误重建；locale 透传 `defaultTagCatalogFromLanguageMessages`。
2. **`defaultTagCatalogFromLanguageMessages(langMsgs, lang string)` 完整体落地**：asm 0x1407756c0
   实证「en-US 先 merge → TrimSpace/fallback en-US → lang 非 en-US 再 merge → tags.defaults →
   splitTagText → defaultTagCatalogFromNames」，替换旧扁平回退实现。
3. **3 调用点 locale 传递**：SaveConfig / CompleteInitialization / commitConfig 传
   `cfg.Preferences.Language`（asm 从调用者栈槽取，非 rip 常量）。
4. **`loadLinkPreferences` 桩签名校正**：`(Preferences, error)`（非 interface{}），并记 asm 体依赖链。

**门禁**：build / vet / test 三项 EXIT=0；`bash build.sh` 完成。

### 批次 62（syncRuntimeServicesAfterSave 比较 helper 落地 + 体长纠偏）

**基线**：`FUNCS=1142 / S=773 / S-sig=90 / P=77 / UNMARKED=189`，SHA256 `B4031691AD878D86AC4D5B158B691141778EE87A9C983C06BD2003B85CF6709E`。
**收口**：`FUNCS=1144 / S=775 / S-sig=90 / P=77 / UNMARKED=189`，SHA256 `721BEC7F8B2D7121BD707EA6BB5DDADD0CBE3081B7CA1656BC5E83EC39739156`。

**本批落地（§29 剩余项 `syncRuntimeServicesAfterSave` 的前置 helper）**：

1. **`screenshotCapturePreferencesEqual@0x14079d7e0`（160B）[S]**：比较截图捕获三元组
   （Controls 默认 true / Cursor、SelectionConfirm 默认 false）。
2. **`fileSearchRuntimeConfigsEqual@0x14079d880`（384B）[S]**：normalizeFileSearchConfig → reflect.DeepEqual。
3. **体长纠偏**：`syncRuntimeServicesAfterSave` 240B→3182B(0xc6e)（8 服务差异同步编排器），
   `PersistSnapshotsForShutdown` 224B→14784B(0x39c0)。

**门禁**：build / vet / test 三项 EXIT=0；`bash build.sh` 完成。

### 批次 65（bootstrapservice_callees.go 23 [P] 存根清查 + 签名证伪校正）

**基线/收口**：`FUNCS=1144 / S=775 / S-sig=90 / P=77 / UNMARKED=189`（无 [S] 落地，无标记漂移），
SHA256 `7A45117FE4CDD5D5EC9252BEC9D36F1765EE6822ECBA780EFEB3DA8E202D5C0C`。

**本批结论**：

1. **23 个 [P] 存根全部「域阻断」**（gpu/qrcode/语言/bookmarks/linkbrowser/desktopwidgets/launcherconfiginput/
   gesture/app启动/launcherupdate/screenshot_pin），无独立直译快赢——清存根须按域落地。
2. **`LaunchAppWithPrivilege@0x1408a60c0` 签名证伪**：第二形参非 `int flags` 系 AppEntry 结构/单元素切片；
   asm 体链已全解码，依赖 loadAppLaunchPrivilegeDefault/resolveEffectiveAppLaunchPrivilege/startApplication
   （app 启动域未落地）。调用点 bootstrapservice.go:1523 传 `7` 亦待校正。
3. **新解码 `retryOLEDBlackoutHotkeysIfNeeded@0x14079e080`（832B 包级）**：控制流已全解码（见 65.md G4），
   属 oledblackout 域（批 71 首目标）。

**路线建议**：批 71–78 UNMARKED 清空（从 oledblackout.go 起）比 [P] 清存根更快收敛。

### 批次 71（oledblackout.go UNMARKED 清空 — 调试日志域）

**基线**：`FUNCS=1144 / S=775 / S-sig=90 / P=77 / UNMARKED=189`。
**收口**：`FUNCS=1146 / S=778 / S-sig=91 / P=77 / UNMARKED=187`，SHA256 `6DADC1264154E8569937F53F3D7291A54F423EA95F9D533AA557F1572E89AC6D`。

**本批落地（oledblackout 调试日志子域，4 函数）**：

1. **`ensureOLEDBlackoutDebugLogger@0x14090ee60`（50B）[S]**：sync.Once.Do(resolve...)。
2. **`parseOLEDBlackoutDebugBool@0x14090f160`（191B）[S]**：0/1/no/on/off/yes/true/false 映射。
3. **`oledBlackoutDebugLog@0x14090eb80`（640B）[S]**：变参日志行落盘（签名由 `(string)` 修正为 `(format string, args ...interface{})`）。
4. **`resolveOLEDBlackoutDebugConfiguration@0x14090eea0`（992B）[S-sig]**：环境变量部分逐条对位，CLI 覆盖扫描近似。

新增 `oledblackout_debug.go`；移除 `oledblackout.go` 两处空存根。UNMARKED 189→187，S 775→778。

### 批次 75（launcherconfigiconcommit.go UNMARKED 清空 — 错误类型方法）

**基线**：`FUNCS=1146 / S=778 / S-sig=91 / P=77 / UNMARKED=187`。
**收口**：`FUNCS=1146 / S=780 / S-sig=91 / P=77 / UNMARKED=185`，SHA256 `76225AEFBCF77A9B39C1262002DD65DE5AD7912911A1E5794542EFD82D88F3E8`。

**本批落地（launcherConfigIconHydrationError 两方法，2 函数）**：

1. **`Error@0x140890d80`（262B）[S]**：nil 收者 → "配置图标水合失败"；否则
   `fmt.Sprintf("配置图标水合失败: %s: %s: %v", TrimSpace(Path), TrimSpace(Ref), Err)`。
   旧体（`"launcherConfigIconHydrationError: path=%q ref=%q err=%v"` 无 TrimSpace/无 nil 守卫）证伪纠正。
2. **`Unwrap@0x140890ea0`（21B）[S]**：nil 收者 → nil，否则 e.Err。

结构偏移实证：Path@0x0 / Ref@0x10 / Err@0x20。本文件 UNMARKED 清零（原 2 个全落地）。
UNMARKED 187→185，S 778→780。

### 批次 76（UNMARKED 清空 — 标记卫生 + 2 处签名/体证伪纠正）

**基线**：`FUNCS=1146 / S=780 / S-sig=91 / P=77 / UNMARKED=185`。
**收口**：`FUNCS=1146 / S=786 / S-sig=91 / P=77 / UNMARKED=179`，SHA256 `C7E50F123BAE73EE4823D1E3E5FA5CE3C1EF655F2D139F68D24F32AB2F84A829`。

**本批落地（6 个 [S]）**：

1. **`canonicalHotkeyModifier@0x140889560`（544B）[S]**：证伪纠正——旧体 meta/super→"Meta"、
   command/cmd/cmdkey→"command"、捏造 al/sh/cntrl/cmdkey 均错。实证映射 Win{os,win,windows,meta,super}/
   Alt{alt,option,optionoralt}/Ctrl{ctrl,control,cmd,command,cmdorctrl}/Shift{shift}。
2. **`normalizeShortcutMode@0x14088a780`（288B）[S]**：签名证伪纠正——`(mode string)` →
   `(mode string, fallback string)`。调用点 appentrynormalize.go:118 补 `e.ShortcutPath`（[推断]）。
3. `launcherConfigIconDataError.Error`（224B）/ `slugify`（576B）/ `cleanStringList`（704B）/
   `normalizeShortcutPath`（96B）[S]：补即时标记（体已正确）。

**关键发现**：多个 UNMARKED 实为已重建体，[S] 标记置于文件头被 `package`/`import`/空行重置，
`count_funcs.sh` 判 UNMARKED。补「紧贴 func 的即时标记」即可诚实收敛，无需新重建。

### 批次 77（UNMARKED 清空 — 批量补即时标记 7 叶子）

**基线**：`FUNCS=1146 / S=786 / S-sig=91 / P=77 / UNMARKED=179`。
**收口**：`FUNCS=1146 / S=793 / S-sig=91 / P=77 / UNMARKED=172`，SHA256 `995684F0BDA998B0C1AACFFC39BBB199FB89A1CFBA2796649CECC7439AF77BCF`。

**本批落地（7 个 [S]，纯标记补齐，体未改）**：

- filterDragLaunchAppIDsByAppEntries(0x140881f60) / inferBookmarkSourceName(0x140766dc0) /
  normalizeLinkIconMode(0x14088b120) / inferName(0x14088ca40) / normalizeAppLaunchPrivilegeMode(0x140882680) /
  normalizeLinkEntries(0x14088a940) / normalizeFileEntries(0x14088b880)。

均为既有重建体补「紧贴 func」即时标记；inferName 与 normalizeAppLaunchPrivilegeMode 的 [P] 边例在标记注释保留。
UNMARKED 179→172，S 786→793。

### 批次 78（UNMARKED 清空 — 续批量补即时标记 6 叶子）

**基线**：`FUNCS=1146 / S=793 / S-sig=91 / P=77 / UNMARKED=172`。
**收口**：`FUNCS=1146 / S=799 / S-sig=91 / P=77 / UNMARKED=166`，SHA256 `42BA3494B11EFC89A39745A1CE201F46464C8F1AFF961D6DD97ED2655351152F`。

**本批落地（6 个 [S]）**：canonicalizeShortcutModifiers(0x1408887c0) / clampRetention /
launcherUpdatePlan.computeTotalSize / isAsciiLetter / isAsciiDigit / shortcutPathExt。

标记卫生路径的可清对象已基本清完。批 71–78 累计：UNMARKED 189→166（-23），S 775→799（+24）。

### 批次 79（bookmarkIconCacheDomain 证伪纠正 + 4 标记补齐）

**基线**：`FUNCS=1146 / S=799 / S-inline=13 / S-sig=91 / P=77 / UNMARKED=166`。
**收口**：`FUNCS=1146 / S=801 / S-inline=14 / S-sig=93 / P=77 / UNMARKED=161`，
SHA256 `F69D60A60C36C4A598C65F974788C14AFC04F09C6A2BC75EB3DD17F5D1F0A2E8`。

**证伪纠正**：bookmarkIconCacheDomain（0x140764480）旧体手写 "www." 剥离+末两段拼接，
反汇编实证真实体用 `publicsuffix.EffectiveTLDPlusOne` 取 eTLD+1，已改体并引入
`golang.org/x/net/publicsuffix` 依赖。

**标记补齐/纠偏**：applyFileDelete[S] / applyFileReplace[S] / copyLauncherUpdateFile[S-sig] /
isAlpha[S-inline] / computeTotalSize 由 [S] 纠为 [S-sig]（[R] 标准累加模式）。

**关键教训**：重建档位不止 [S]/[S-sig]/[P]，还有 [R]（标准模式还原）/ [F]（功能实现）/ [G]（Ghidra 反编译）。
标记卫生须逐函数核对档位，不得见 header=[S] 即标 [S]。

### 批次 80（normalizeBookmarkSources [S] + startmenu 4 helper 补标）

**基线**：`FUNCS=1146 / S=801 / S-inline=14 / S-sig=93 / P=77 / UNMARKED=161`。
**收口**：`FUNCS=1146 / S=802 / S-inline=18 / S-sig=93 / P=77 / UNMARKED=156`，
SHA256 `8CDC14843EAD7DC18C6C70896C53E733DBF8B59B029B3E6B37812E242DA50E5F`。

**本批落地**：normalizeBookmarkSources[S]（符号 0x14088b260，1568B 逐条对位）；
defaultStartMenuAppResolvers / invokeStartMenuNameResolver / resolveStartMenuScanDisplayName /
composeStartMenuScanRoots 四 helper [S-inline]（内联于 startmenu 主循环，无独立符号）。

### 批次 81（idAllocator/splitCommandLine/parseLauncherAssetRequest/Clear 补标）

**基线**：`FUNCS=1146 / S=802 / S-inline=18 / S-sig=93 / P=77 / UNMARKED=156`。
**收口**：`FUNCS=1146 / S=805 / S-inline=18 / S-sig=94 / P=77 / UNMARKED=152`，
SHA256 `80D18817E31C09EA980F69F0C4AB92BB99B492917C98D090189D3AF455C979D6`。

**本批落地**：idAllocator.Next[S]（0x14088cee0，480B）；splitCommandLineArguments[S]（0x1408a7ee0，1440B）；
parseLauncherAssetRequest[S]（0x140873760，[G] Ghidra 实证）；launcherAssetService.Clear[S-sig]（0x140871fc0）。

### 批次 82（networkaccess [S-sig] + absInt [S-inline] + exit [S-sig]）

**基线**：`FUNCS=1146 / S=805 / S-inline=18 / S-sig=94 / P=77 / UNMARKED=152`。
**收口**：`FUNCS=1146 / S=805 / S-inline=19 / S-sig=98 / P=77 / UNMARKED=147`，
SHA256 `96C4DB8DC941262FDB595D3589B02169C050E90405C55C2F7440CC35A4043AD2`。

**本批落地**：networkaccess 4 方法 [S-sig]（第二参确证为 timeout int64=5s，非 maxBytes）；
absInt [S-inline]；workspaceDataMaintenanceGate.exit [S-sig]。

**标记未落地（待复核）**：
- msToDuration：体 `*time.Millisecond` 与注释「asm 实证 3000→3s/5000→5s/其他→0」冲突，需反汇编复核。
- encodeJPEG：placeholder（返回 "JPEG编码暂未实现"），属 [P]。

### 批次 83（msToDuration 证伪纠正 + 5 叶子补标）

**基线**：`FUNCS=1146 / S=805 / S-inline=19 / S-sig=98 / P=77 / UNMARKED=147`。
**收口**：`FUNCS=1146 / S=807 / S-inline=20 / S-sig=101 / P=77 / UNMARKED=141`，
SHA256 `CA9A1B4A2EEF8C56FE9CEAC84D48073C41F3872C72D2650353411DD88F149E8A`。

**证伪纠正**：msToDuration 旧体 `*time.Millisecond` 对非 3000/5000 值返回 ms ms，实为 0。
反汇编实证 switch（3000→3s / 5000→5s / 其他→0），已改体。疑点清单已销号。

**标记补齐**：newExtractError[S] / extractError.Error[S-inline] / getFileVersionInfoSize[S-sig] /
getFileVersionInfo[S-sig] / verQueryValue[S-sig]。

### 批次 84（internetshortcut/pluginpath/remoteicons 5 叶子补标）

**基线**：`FUNCS=1146 / S=807 / S-inline=20 / S-sig=101 / P=77 / UNMARKED=141`。
**收口**：`FUNCS=1146 / S=812 / S-inline=20 / S-sig=101 / P=77 / UNMARKED=136`，
SHA256 `3EB4843F697B268A2909906462E8CF71F97A155962DE9EB83F11737E54E867B9`。

**本批落地（5 个 [S]，含符号+文件头汇编实证）**：decodeInternetShortcutTextBounded(0x14086ab60) /
decodeUTF16ShortcutText(0x14086ada0) / pluginPathHasReparsePoint(0x140923b40) /
remoteIconPathHasReparsePoint(0x140966240) / replaceRemoteIconCacheFile(0x140966340)。

### 批次 85（internetshortcutparse/read + launcherupdate_security 4 叶子补标）

**基线**：`FUNCS=1146 / S=812 / S-inline=20 / S-sig=101 / P=77 / UNMARKED=136`。
**收口**：`FUNCS=1146 / S=816 / S-inline=20 / S-sig=101 / P=77 / UNMARKED=132`，
SHA256 `05E257635F28055F6ECBC8AE9688AAD0C164B860D7C8D9EF888A2C6934A62DE4`。

**本批落地（4 个 [S]）**：parseInternetShortcutValues(0x14086a5a0) / readInternetShortcutFile(0x14086a1a0) /
launcherUpdatePathHasReparsePoint(0x1408cb340) / replaceLauncherUpdateMetadataFile(0x1408cb440)。

### 批次 86（internetshortcuticon/archiveextract_open/hasReparsePoint/unique 4 叶子补标）

**基线**：`FUNCS=1146 / S=816 / S-inline=20 / S-sig=101 / P=77 / UNMARKED=132`。
**收口**：`FUNCS=1146 / S=818 / S-inline=22 / S-sig=101 / P=77 / UNMARKED=128`，
SHA256 `66326D8B96A7D0916338424A27E16A754766A35058E4A0744976E2EBF3129DE5`。

**本批落地**：buildInternetShortcutIconLocation[S]（0x14086a9c0）/ openArchiveRegularFileNoFollow[S]（0x140750d20）/
hasReparsePoint[S-inline] / uniqueLauncherBindingValues[S-inline]。

**方法学修正**：本轮重写 survey（`tools/list_unmarked.awk`，严格对齐 count_funcs.sh 逻辑），
剔除多行注释标记误判，得到真实 UNMARKED 清单（含 return nil 存根）。

### 批次 87（automaticpath_windows 5 叶子补标）

**基线**：`FUNCS=1146 / S=818 / S-inline=22 / S-sig=101 / P=77 / UNMARKED=128`。
**收口**：`FUNCS=1146 / S=823 / S-inline=22 / S-sig=101 / P=77 / UNMARKED=123`，
SHA256 `0C25F9AEAEBE7BD03DFDBF25C7474547C926602E1CAADD113FEB6F51263E84F7`。

**本批落地（5 个 [S]）**：classifyAutomaticWindowsPathLexically(0x14075b8e0) /
classifyAutomaticWindowsPath(0x14075b180) / normalizeAutomaticWindowsReparseDestination(0x14075bae0) /
automaticWindowsDriveType(0x14075bcc0) / automaticWindowsPathAttributes(0x14075bd40)。

### 批次 88（bookmarkicon 3 内联叶子补标）

**基线**：`FUNCS=1146 / S=823 / S-inline=22 / S-sig=101 / P=77 / UNMARKED=123`。
**收口**：`FUNCS=1146 / S=823 / S-inline=25 / S-sig=101 / P=77 / UNMARKED=120`，
SHA256 `710ED497A20305F31EAC5535F31E8FA1952098FEABB48BD7C45F4B8D3223F19D`。

**本批落地（3 个 [S-inline]）**：resolveBookmarkIconSource（内联于 0x14075c120）/
tryGetOrReuseFaviconDBEntry / openNewFaviconDBEntry（内联于 0x14075f5e0）。

### 批次 89（launcherconfigiconcommit 8 叶子标记卫生）

**基线**：`FUNCS=1146 / S=823 / S-inline=25 / S-sig=101 / P=77 / UNMARKED=120`。
**收口**：`FUNCS=1146 / S=827 / S-inline=25 / S-sig=105 / P=77 / UNMARKED=112`，
SHA256 `A5EE0048EB1AD0CEE4B207A2792970909203F35A323C2A825DABC6116FF4E491`。

**本批落地（8 个标记卫生修复）**：finish/collectLauncherConfigIconRefCounts/
prepareLauncherConfigIconsForCommit/prepareLauncherConfigIconsForCommitWithCurrentRefs → [S]；
validateLauncherConfigIconCandidateBudget/syncLauncherConfigBeforeIconPrune/
buildSelfContainedLauncherConfig/launcherConfigHasInlineIconData → [S-sig]。
根因：注释块与 func 间空行重置 count_funcs.sh 的 last → 移除 8 处空行。

### 批次 90（normalizeAppEntries/main/bootstrapservice 9 叶子 + resolveLauncherConfigFilePath 证伪纠正）

**基线**：`FUNCS=1146 / S=827 / S-inline=25 / S-sig=105 / P=77 / UNMARKED=112`。
**收口**：`FUNCS=1146 / S=829 / S-inline=29 / S-sig=109 / P=77 / UNMARKED=102`，
SHA256 `492258E661D563F0200B75E7B062BDC82038C8331C351D2424CB44B178FDB2FA`。

**本批落地**：resolveLauncherConfigFilePath[S 证伪纠正]/normalizeAppEntries[S]（0x140889780）/
main.main[S-sig]/commitLauncherConfigReplacementWithWidgetsImpl[S-sig]/
CopyScreenshot{Image,Ref}ToClipboard[S-sig]/GetFileLocatorResultDetail[S-inline]/
screenshot{AnnotationLineWidth,CornerRadius}Setting[S-inline]/collectLauncherHotkeyRegistrationErrors[S-inline]。

**证伪纠正**：resolveLauncherConfigFilePath 旧体 `return layout.ConfigFile` 错误；实为
`Getenv("USBEAM_LAUNCHER_CONFIG")→TrimSpace→Abs / 空→Join(Root,"UsbEAm_Launcher_Config.json")`。

### 批次 91（shortcutinfo/End/launcherconfig 6 标记）

**基线**：`FUNCS=1146 / S=829 / S-inline=29 / S-sig=109 / P=77 / UNMARKED=102`。
**收口**：`FUNCS=1146 / S=829 / S-inline=31 / S-sig=113 / P=77 / UNMARKED=96`（首破 100），
SHA256 `A8E67E9459031343597B4669CC419421D26E43C65A5F39272B6E3E52F3F4FA12`。

**本批落地**：newShortcutError/shortcutStubError.Error[S-inline] /
workspaceDataMaintenanceGate.End[S-sig] / validateTwoFactorStoredConfig[S-sig 0x1409c6740] /
normalizeLauncherConfigForSave[S-sig 0x140877420] / validateChangedIconData[S-sig 0x140899000]。

### 批次 92（encodeJPEG 实现 + twofactor 12 签名标记）

**基线**：`FUNCS=1146 / S=829 / S-inline=31 / S-sig=113 / P=77 / UNMARKED=96`。
**收口**：`FUNCS=1146 / S=829 / S-inline=32 / S-sig=125 / P=77 / UNMARKED=83`，
SHA256 `A9B9EA5D5689CFE2FD41F3383FD13A02FD81CC096C8996EB8C9CDB697F329A41`。

**本批落地**：encodeJPEG[S-inline 真实现 jpeg.Encode Quality 92]（asm 0x140967f80 证质量 0x5c）/
twoFactorService 12 方法[S-sig]（符号 + call target 实证，见 acceptance/92.md 表）。

### 批次 93（inputmonitor/service_configure/pathclip/hotkey 18 签名标记）

**基线**：`FUNCS=1146 / S=829 / S-inline=32 / S-sig=125 / P=77 / UNMARKED=83`。
**收口**：`FUNCS=1146 / S=829 / S-inline=33 / S-sig=142 / P=77 / UNMARKED=65`，
SHA256 `26381FBFB057D66F318555DDB2F10755CDA98FE58208C2902C5B527DEEB89112`。

**本批落地**：inputMonitorService 3[S-sig]/mouseGestureService 7[S-sig]/
openWithDefaultHandler+openPath*+setFileClipboard 5[S-sig]/newPathError[S-inline 内联实证]/
remove/addLauncherAppHotkeyBindings 2[S-sig]。签名全部经符号表 + [S] 调用方 call target 实证。

### 批次 94（filelocator/desktopwidget/hotkey_dispatch 16 签名标记）

**基线**：`FUNCS=1146 / S=829 / S-inline=33 / S-sig=142 / P=77 / UNMARKED=65`。
**收口**：`FUNCS=1146 / S=829 / S-inline=34 / S-sig=157 / P=77 / UNMARKED=49`（首破 50），
SHA256 `46B94737A39047B5F34C81852AAFB98C1EC2620D8C8D63E3E57C9D09248898C9`。

**本批落地**：fileLocatorService 7[S-sig]/desktopWidgetService 7[S-sig]/
screenshotHotkeyCapturePreferences[S-sig 0x140795300]/emitLauncherEvent[S-inline 内联]。
签名全部经符号表实证。

### 批次 95（audio/screenshotpin/oledblackout 49 标记 —— UNMARKED 归零）

**基线**：`FUNCS=1146 / S=829 / S-inline=34 / S-sig=157 / P=77 / UNMARKED=49`。
**收口**：`FUNCS=1146 / S=829 / S-inline=34 / S-sig=203 / P=80 / UNMARKED=0`（归零里程碑），
SHA256 `6C006ADFBE230A694F15642A36709D090295949B2DEA383DCFDC6E75E3E5BE0F`。

**本批落地**：audio_windows 16[S-sig]/screenshotpin 16[S-sig]/oledblackout 14[S-sig]+3[P]。
签名全部经符号表实证；体骨架（WASAPI COM / Wails v3 window 域待落地）。

### 批次 96（[P] 存根升格 [S-sig]：callees 26 + filesearch 6）

**基线**：`FUNCS=1146 / S=829 / S-inline=34 / S-sig=203 / P=80 / UNMARKED=0`。
**收口**：`FUNCS=1146 / S=829 / S-inline=34 / S-sig=235 / P=48 / UNMARKED=0`，
SHA256 `F3DCF42B5A012E8D83A3E40229885B0869D7C2B590883DE6E40EAF6C9D61FCD0`。

**本批落地**：bootstrapservice_callees 26 [P]→[S-sig]（GPU/QR/语言/书签/桌面小部件/手势等，见
acceptance/96.md）+ filesearch_normalize 6 [P]→[S-sig]。LaunchAppWithPrivilege 因「签名证伪」保留 [P]。

### 批次 97（[P] 存根升格 [S-sig]：config/migration/state_deps/window/toast/lifecycle 23）

**基线**：`FUNCS=1146 / S=829 / S-inline=34 / S-sig=235 / P=48 / UNMARKED=0`。
**收口**：`FUNCS=1146 / S=829 / S-inline=34 / S-sig=258 / P=25 / UNMARKED=0`，
SHA256 `C0C5B18513DD1E4634478D5878BD58E95B749CCDF9AEA903BF7458AE94BFBCD1`。

**本批落地**：config 10 / migration 5 / state_deps 3 / window 1 / toast 2 / lifecycle 2 共 23 [P]→[S-sig]
（签名经 VA + 调用点寄存器实证，见 acceptance/97.md 表）。

### 批次 98（[P] 升格：windowmanagement 8 + desktopwidget 3 + screenshot 1 + launcherasset 1 + oledblackout 1）

**基线**：`FUNCS=1146 / S=829 / S-inline=34 / S-sig=258 / P=25 / UNMARKED=0`。
**收口**：`FUNCS=1146 / S=829 / S-inline=35 / S-sig=271 / P=11 / UNMARKED=0`，
SHA256 `69F04FD1A3B0F7990DD89BCD0F87F0C26C6FD2CF5464C086F706DAA46AD3CF43`。

**本批落地**：windowmanagement 8 平台函数 [S-sig]（0x1409e9340/0x1409e9a80/0x1409ea8e0/0x1409eaae0/
0x1409eaf00/0x1409eb240/0x1409eb9c0/0x1409ec180）+ desktopwidget 3 [S-sig]（0x1407bfd20/0x1407c5040/
0x1407b3380）+ screenshot_stubs resolveGeneratedScreenshotSaveFormat [S-sig] 0x140967a20 +
launcherasset resolvedLimits [S-inline] + oledblackout AttachApp [S-sig] 0x1408ff580。

### 批次 99（[P] 清尾 + 手势分派串校正 + LaunchAppWithPrivilege 签名证伪纠正）

**基线**：`FUNCS=1146 / S=829 / S-inline=35 / S-sig=271 / P=11 / UNMARKED=0`。
**收口**：`FUNCS=1142 / S=829 / S-inline=35 / S-sig=277 / P=1 / UNMARKED=0`，
SHA256 `CC54A40A2C19EC831846A1D1B4B174C2CADAA181DE1BEE1D59D44B3F0D63D57F`。

**本批落地**（三处实证纠正，详见 acceptance/99.md）：
1. **LaunchAppWithPrivilege 签名证伪纠正**：[P]→[S-sig] 0x1408a60c0，签名由 `(appID string, flags int)`
   纠正为 `(privilege string, entry AppEntry)`；调用点实证传 `("default", entry)`（0x140c39338 "default"）。
2. **executeMouseGestureLauncherFeatureAction 分派串校正**：[S] 4 处分派串错误（screenshot.captureall/
   captureactivewindow/capturecurrentscreen/captureselectedwindow 全错），经 gowrap1-5 asm + memequal
   目标 VA 解码纠正为 screenshot.scrolling/screenshot.allScreens/qrcode.screenSelection/
   screenshot.currentScreen；删 5 个 `mouseGestureCapture*` 重建脚手架（真实系 gowrap 闭包）。
3. **新增 captureQRCodeByHotkey** [S-sig] 0x140794da0（qrcode.screenSelection 分派目标）。
4. **接口实现类 [P]→[S-sig]**（4）：launcherGlobalHotkeyManagerImpl.Close/Update =
   windowsLauncherGlobalHotkeyManager.Close 0x14089da00 / Update 0x14089d8e0；
   oledBlackoutHotkeyManagerStub.Close/Update = windowsOLEDBlackoutHotkeyManager.Close 0x14090f540 /
   Update 0x14090f3e0。

### 批次 99 后（下一步移交清单）

**当前基线**：`FUNCS=1142 / S=829 / S-inline=35 / S-sig=277 / P=1 / UNMARKED=0`。

**剩余 [P]（1 个）**：`newOLEDLifecycleContext`（oledblackout.go）——重建脚手架，真实二进制为
`oledBlackoutService.ensureLifecycleLocked` 0x1408ff220（方法非 helper，签名不同，诚实保留 [P]；
补体需按 ensureLifecycleLocked 方法还原，而非 helper）。

**遗留一致性提示**：resolveWorkspaceLayout 体用 `"config.json"`，而 asm 实证配置基准名为
`"UsbEAm_Launcher_Config.json"`（launcherconfigiconstore.go 保留名 0x140c69c4e 已证）。
tests 断言 `config.json`，属待取证的一致性差（可对 resolveWorkspaceLayout 0x1407a15e0 复核）。
另：encodeScreenshotJPEGFromPNG [S] 体缺 screenshotImageToOpaqueRGBA 去 alpha 步骤，待审计。

**`loadLinkPreferences` 体仍 [S-sig]，剩余阻断（按落地顺序）**：

1. **语言加载域类型校正**（`loadLinkPreferences` 前置）：asm 中 `langMsgs` 为
   `map[string]map[string]interface{}`（mapaccess2_faststr 单字解引用），`mergeLanguageMessages` /
   `nestedLanguageMessageString` 传 `&map` 而非按值 map——当前树 `map[string]interface{}` 系误建，
   需专项校正 `loadLanguageMessages` / `GetLanguageMessages` 返回类型及连锁调用点。
2. **`loadLauncherConfigOrDefaultIfMissing`(0x140875e00, 736B)**：返回结构未映射
   （asm 复制 0x126 qword 与 LauncherConfig 0x142 不一致，疑似含 DesktopWidgetDocument 子结构）。
3. **`normalizePreferencesWithOptions`(0x14087cb00, 8160B)**：[P] 巨型规范化函数，需独立批次。
4. **linkbrowser*.go 域**（detectLinkBrowsers 22KB / openLinkWithPreferences 13KB / OpenLink 等）：
   link 域整体未落地。

**其余 §29 项（62–64，依赖不同域）**：

1. `syncRuntimeServicesAfterSave`（[S-sig]）——2 比较 helper 已 [S]
   （screenshotCapturePreferencesEqual / fileSearchRuntimeConfigsEqual），体仍缺：
   `retryOLEDBlackoutHotkeysIfNeeded`(0x14079e080)、feature 位图 diff 语义、8 服务 Configure 编排
   （服务字段名可由 syncRuntimeServices [S] 复用）。
2. `PersistSnapshotsForShutdown`（[S-sig]）——缺 `snapshotMirror`(20896B)/`saveSnapshots`（screenshot_pin_windows.go）。
3. `attachLauncherBackgroundURL`（[P]，0x140798600, 2016B）——缺 launcher asset 背景域。

**注意**：`loadLinkPreferences` 体、`syncRuntimeServicesAfterSave`、`PersistSnapshotsForShutdown`、
`attachLauncherBackgroundURL` 的 asm 资产均在 `docs/goresym/pipeline/tmp/`，但各自依赖未落地 helper/域；
补体前须先落地依赖，或按 §10 归入专项批次。不要重复做交叉分析——结论已定。

### 批次 100（screenshot clipboard 剪贴板写入链还原）

**基线**：`FUNCS=1142 / S=829 / S-inline=35 / S-sig=277 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1149 / S=836 / S-inline=35 / S-sig=277 / P=1 / UNMARKED=0`，
SHA256 `C27699ED49994FD1A721461C570C69F48DF29BB8EC5416C47BBC39ADB7B93198`。

**本批落地**（详见 acceptance/batch100.md）：
1. 新建 `backend/screenshot_clipboard_windows.go`，落地 7 个 `[S]`：
   `wrapWinError` 0x1408af3a0（共享 helper，`%s: %w` 格式串，ERROR_SUCCESS 判空）、
   `globalAllocMoveable` 0x1408ae9e0、`globalLock` 0x1408aeac0、
   `registerScreenshotPNGClipboardFormat` 0x140972ee0（RegisterClipboardFormatW("PNG")）、
   `openScreenshotClipboard` 0x140972fc0（40 次重试 + 5ms 间隔）、
   `setScreenshotClipboardBytes` 0x1409730a0、`buildScreenshotClipboardDIBV5` 0x140973380
   （124B BITMAPV5HEADER + BGRA 自底向上）。
2. `writeScreenshotPNGToClipboard`（screenshot_funcs.go）由旧简化体（w32 直写 CF_DIBV5）
   重构为完整 asm 链：register → buildDIBV5 → LockOSThread → open → EmptyClipboard →
   set(PNG) → set(CF_DIBV5)；错误按 PNG/CF_DIBV5 分支组合 + defer CloseClipboard 失败写回。
3. 字符串常量全部经 resolve_lea_strings.py 确定性解码（"PNG"/"png"/9 条中文错误消息/`%s: %w`），
   BITMAPV5HEADER 模板 0x1411df5e4 的 124B 字段值全部 dump 实证（掩码/CSType/Intent）。

### 批次 100 后（下一步移交清单）

**当前基线**：`FUNCS=1149 / S=836 / S-inline=35 / S-sig=277 / P=1 / UNMARKED=0`。

**剩余 [P]（1 个）**：`newOLEDLifecycleContext`（同 batch 99 移交，未动）。

**已解锁依赖**：`wrapWinError` / `globalAllocMoveable` / `globalLock` 三个共享 helper 现已 [S]，
供 qrcode / filesearch / 文件操作等域复用（此前这些域因缺 Win32 helper 而阻断）。

**下一批方向（按 asm 就绪度 + 依赖闭环）**：
1. screenshot 域收尾：`screenshot_scroll_windows.go` / `screenshot_pin.go`（依赖已闭环）。
2. 独立整域启动：qrcode_windows.go（329 蓝图函数 / 81 asm）——globalAllocMoveable/globalLock/
   wrapWinError 现已就绪，可落地 qrcode clipboard/文件操作链。
3. filesearch_windows.go（321 / 73）、twofactor.go（140 / 30）等整域。

### 批次 101（QR 码生成与剪贴板复制链 + 原生依赖引入）

**基线**：`FUNCS=1149 / S=836 / S-inline=35 / S-sig=277 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1152 / S=841 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`，
SHA256 `C52957B29C4C7A53EBF8C7DD7B0DAB9E7FDBA67B8050898EF0432C515120E49C`。

**本批落地**（详见 acceptance/batch101.md）：
1. 引入原生依赖：`github.com/makiuchi-d/gozxing v0.1.1`（QR 编码）、
   `golang.design/x/clipboard v0.8.0`（cgo-free 剪贴板，`Write(Format,[]byte) <-chan struct{}`
   旧版 API，与 asm 两参调用一致；v0.9.0 已改 `Write(ctx,...)` 故锁 v0.8.0）。
2. 新建 `backend/screenshot_qrcode_windows.go`，落地 5 个 `[S]`：
   `clampQRCodeDimension`（size<192→420 / >512→512）、
   `generateQRCodePNG` 0x140931a40（TrimSpace 空判定 + gozxing Encode + 透明渲染 + png.Encode）、
   `renderTransparentQRCodeImage` 0x140931ca0（BitMatrix→NRGBA，黑模块不透明/白模块透明）、
   `generateQRCodeDataURL` 0x140931e40（base64 data URL）、
   `copyQRCodeImageToClipboard` 0x140931f00（clipboard.Write(FmtImage)）。
3. `initQRCodeClipboard` 空壳补体（sync.Once + clipboard.Init + 缓存 error）。
4. 签名证伪纠正：`GenerateQRCodeDataURL` / `CopyQRCodeImageToClipboard` 两方法由
   `(data string)` 校正为 `(data string, size int)`（asm 三字转发实证）。

### 批次 101 后（下一步移交清单）

**当前基线**：`FUNCS=1152 / S=841 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。

**剩余 [P]（1 个）**：`newOLEDLifecycleContext`（同 batch 99 移交，未动）。

**已解锁依赖**：`wrapWinError` / `globalAllocMoveable` / `globalLock`（批 100）+
`generateQRCodePNG` / `renderTransparentQRCodeImage` / `initQRCodeClipboard`（批 101）现已 [S]；
gozxing + clipboard 两原生库已引入，qrcode 域编码/剪贴板基础设施闭环。

**下一批方向（按 asm 就绪度 + 依赖闭环）**：
1. qrcode 域续链：`decodeQRCodesFromImage` / `buildQRCodeDecodeResultFromPNG`
   （0x140931fe0/0x140932400，gozxing 解码端，依赖已闭环）→ 完成 QR 读/写双向。
2. 文件剪贴板链：`setFileClipboard` 0x1408a9f00 → `normalizePathList` 0x14088c420 /
   `resolveClipboardDropEffect` 0x1408adbe0 / `setWindowsFileClipboard` 0x1408adcc0
   （globalAllocMoveable/globalLock/wrapWinError 已就绪）。
3. screenshot scroll_canvas 域（16 函数已 dump asm，类型定义已存在 types_screenshot.go）。
4. filesearch_windows.go / twofactor.go 等整域。

### 批次 102（文件剪贴板链前置纯逻辑 + SetFileClipboard 签名纠正）

**基线**：`FUNCS=1152 / S=841 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1155 / S=844 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`，
SHA256 `7CD5EB58B5FE7E076BC58DA52BE12C9AFB0239CE19E88F2C365EE655337A6CEC`。

**本批落地**（详见 acceptance/batch102.md）：
1. 新建 `backend/pathclip_clipboard_windows.go`，落地 3 个 `[S]` + 1 个 `[S-sig]`：
   `normalizePathList` 0x14088c420（TrimSpace→Clean→ToLower 去重，返回 []string 无 error）、
   `resolveClipboardDropEffect` 0x1408adbe0（cut/move→2、copy→1，否则
   `不支持的剪贴板文件操作`）、`setFileClipboard` 0x1408a9f00（编排：空路径报错→
   dropEffect 解析→setWindowsFileClipboard）、`setWindowsFileClipboard` 0x1408adcc0 骨架。
2. 签名证伪纠正：`BootstrapService.SetFileClipboard` 由 `(paths []string)` 校正为
   `(paths []string, dropEffect string)`（asm 五字转发实证）。

### 批次 102 后（下一步移交清单）

**当前基线**：`FUNCS=1155 / S=844 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。

**下一批（批次 103）方向**：补体 `setWindowsFileClipboard` 0x1408adcc0（640B + 3 闭包
func1 0x1408ae200 / func2 0x1408ae1a0 / func3 0x1408ae140），依赖链：
`openClipboardForFileOperation` 0x1408ae240（224B）、`buildHDropClipboardData` 0x1408ae320
（608B + func1 0x1408ae580）、`buildDropEffectClipboardData` 0x1408ae5e0（384B + func1
0x1408ae760）。`wrapWinError`/`globalAllocMoveable`/`globalLock`（批 100）已就绪。

### 批次 103（CF_HDROP 数据构建链）

**基线**：`FUNCS=1155 / S=844 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1158 / S=847 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`，
SHA256 `482AD4B80EAB26A4FC28876652F1A50E832DAED0925410D2A931E433DAA27A5F`。

**本批落地**（详见 acceptance/batch103.md）：
1. `encodeHDropPaths` 0x1408ae7c0（路径 → UTF-16 DROPFILES 块，双 null 终止，`[]uint16`）。
2. `buildHDropClipboardData` 0x1408ae320（GlobalAlloc(len*2+20) + DROPFILES 头 pFiles=20/fWide=1
   + 路径块，defer GlobalUnlock，GlobalLock 失败 GlobalFree）。
3. `buildDropEffectClipboardData` 0x1408ae5e0（GlobalAlloc(4) + 写 uint32 dropEffect）。

### 批次 103 后（下一步移交清单）

**当前基线**：`FUNCS=1158 / S=847 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。

**下一批（批次 104）方向**：补体 CF_HDROP 编排侧：
`openClipboardForFileOperation` 0x1408ae240（OpenClipboard 100 次重试 + sleep(10ms)）、
`registerPreferredDropEffectClipboardFormat` 0x1408aeb60（RegisterClipboardFormatW）、
`setWindowsFileClipboard` 0x1408adcc0（LockOSThread + defer 链 + SetClipboardData(CF_HDROP)
+ SetClipboardData(PreferredDropEffect) + wrapWinError 编排）。落地后 `setFileClipboard` 全链 `[S]`。
依赖：`w32.OpenClipboard` / `w32.CloseClipboard` / `w32.SetClipboardData` 已就绪
（screenshot_clipboard_windows.go），但需确认 `Preferred DropEffect` 格式注册名。

### 批次 104（CF_HDROP 编排侧）

**基线**：`FUNCS=1158 / S=847 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1160 / S=850 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`，
SHA256 `5031B48A254E51AB6D83D3E8173CA04D20042F33D40817CDA727CF5E387F5BBB`。

**本批落地**（详见 acceptance/batch104.md）：
1. `openClipboardForFileOperation` 0x1408ae240（OpenClipboard 100 次重试 + sleep(10ms) +
   最后再试一次，失败 `打开剪贴板失败`）。
2. `registerPreferredDropEffectClipboardFormat` 0x1408aeb60（注册 `Preferred DropEffect`
   格式，失败 `注册剪贴板文件操作格式失败`）。
3. `setWindowsFileClipboard` 0x1408adcc0 补体（LockOSThread + EmptyClipboard +
   SetClipboardData(CF_HDROP=15) + SetClipboardData(PreferredDropEffect) + CloseClipboard，
   错误消息：`清空剪贴板失败`/`写入文件剪贴板失败`/`写入剪贴板文件操作失败`/`关闭剪贴板失败`）。
   `setFileClipboard` 全链 `[S]`。

### 批次 104 后（下一步移交清单）

**当前基线**：`FUNCS=1160 / S=850 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。

**下一批（批次 105）方向**（候选，二选一）：
- **QR 解码链**：`buildQRCodeDecodeResultFromPNG` 0x140931fe0（stub 在
  bootstrapservice_callees.go:66）、`decodeQRCodesFromImage` 0x140932400、
  `tryDecodeMultipleQRCodes` 0x140932640（1888B）、`convertQRCodeResults` 0x1409326e0、
  `buildQRCodeDecodedEntryFromPoints` 0x140932da0、`buildQRCodeDecodedEntryKey` 0x140933720。
  `decodeQRCodesFromImage` 调 `gozxing.NewBinaryBitmapFromImage` +
  `qrcode.QRCodeReader.Decode`。
- **shellExecuteProgram** 0x1408aec40（已见部分 asm，UTF16FromString + utf16PtrOrNil +
  ShellExecuteW 链）。

### 批次 105（二维码解码链）

**基线**：`FUNCS=1160 / S=850 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1165 / S=853 / S-inline=35 / S-sig=276 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch105.md，新文件 screenshot_qrcode_decode.go）：
1. `tryDecodeMultipleQRCodes` 0x140932640（QRCodeMultiReader.DecodeMultiple 包装，err/空判）。
2. `decodeQRCodesFromImage` 0x140932400（NewBinaryBitmapFromImage → hints{TRY_HARDER} →
   多码 → img.Bounds() → convertQRCodeResults；空则单码 QRCodeReader.Decode 回退）。
3. `convertQRCodeResults` 0x1409326e0（TrimSpace + 格式映射 switch 表 18 分支 →
   BarcodeFormat.String() → key 去重 map[string]struct{} → []QRCodeDecodedEntry）。
4. `buildQRCodeDecodedEntryKey` 0x140933720 `[S-sig]`（!selectable 实证 strings.Join(...,":")；
   selectable fmt.Sprintf 坐标格式串推断）。
5. `buildQRCodeDecodedEntryFromPoints` 0x140932da0 `[S-sig]`（points<3 / bounds 无效 → 空）。

### 批次 105 后（下一步移交清单）

**当前基线**：`FUNCS=1165 / S=853 / S-inline=35 / S-sig=276 / P=1 / UNMARKED=0`。

**下一批（批次 106）方向**（优先补 QR 链收尾，再转 shellExecuteProgram）：
- **buildQRCodeDecodedEntryFromPoints 坐标公式** 0x140932da0（2432B，ResultPoint.GetX/GetY
  计算 marker 与 bounds，需读全 470 行 asm）。
- **buildQRCodeDecodedEntryKey selectable 格式串** 0x140933720（fmt.Sprintf 多段拼接实证）。
- **buildQRCodeDecodeResultFromPNG** 0x140931fe0（1056B，返回 `QRCodeDecodeResult` 大结构：
  ImageData=data URL + Entries；需连带修正 `DecodeQRCodesFromImageData`/`DecodeQRCodesFromScreenshotRef`
  的签名——从结果提取 Text 列表）。
- **shellExecuteProgram** 0x1408aec40（UTF16FromString + utf16PtrOrNil + ShellExecuteW 链）。

### 批次 106（二维码解码链收尾 + 返回类型修正）

**基线**：`FUNCS=1165 / S=853 / S-inline=35 / S-sig=276 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1165 / S=855 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch106.md）：
1. `buildQRCodeDecodedEntryFromPoints` 0x140932da0 补体 `[S-sig]`→`[S]`
   （marker=p0+(p2-p1) 外推 + 包围盒 min/max + 归一化 (v-Min)/Max；Selectable 恒 false）。
2. `buildQRCodeDecodeResultFromPNG` 0x140931fe0 补体 `[S-sig]`→`[S]`
   （返回 `QRCodeDecodeResult`：ImageData=data URL + ImageWidth/Height + Entries +
    SelectedEntryKey；错误 `解析截图内容失败: %w`）。
3. **返回类型修正**：4 个方法 `([]string,error)` → `(QRCodeDecodeResult,error)`
   （`DecodeQRCodesFromScreenSelection` / `DecodeQRCodesFromImageData` /
   `DecodeQRCodesFromScreenshotRef` / `captureQRCodesFromScreenSelectionForService`）。
   `DecodeQRCodesFromImageData`（0x140788dc0）asm 实证大结构返回（duffcopy 栈搬运）。
4. **reader 单例修正**：gozxing `QRCodeReader`/`QRCodeMultiReader` 须经 `New*` 初始化
   （否则 GetDecoder nil panic），改包级单例匹配目标全局 0x140c140e0。

### 批次 106 后（下一步移交清单）

**当前基线**：`FUNCS=1165 / S=855 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。

**下一批（批次 107）方向**：
- **captureQRCodesFromScreenSelectionForService** 0x140933c40（跨 screenshot 域，[S-sig]
  遗留；调 emitQRCodeDecoded 0x140799d80 → 屏幕选择捕获 + 解码）。
- **shellExecuteProgram** 0x1408aec40（UTF16FromString + utf16PtrOrNil + ShellExecuteW 链）。

### 批次 107（Shell 执行域 8 函数）

**基线**：`FUNCS=1165 / S=855 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1173 / S=863 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch107.md，`backend/shell_execute_windows.go` 新建）：
1. `shellExecuteProgram` 0x1408aec40（ShellExecuteExW，参数序 verb,program,args,workingDir；
   cbSize=0x70/fMask=0x50c）。
2. `showShellProperties` 0x1408aeea0（SHObjectProperties，SHOP_FILEPATH/SHOP_VOLUMEGUID）。
3. `withShellApartment` 0x1408af020（LockOSThread + CoInitializeEx(0,6) + defer CoUninitialize）。
4. `resolveCmdExePath` 0x1408af1c0（ComSpec → SystemRoot\System32 → cmd.exe）。
5. `utf16PtrOrNil` 0x1408af2a0（TrimSpace 空→nil，否则 &u[0]）。
6. `quoteCmdArgument` 0x1408af320（CMD `""` 转义 + 引号包裹）。
7. `startPowerShellEncodedCommand` 0x1408af500（-NoProfile -Sta -EncodedCommand）。
8. `encodePowerShellCommand` 0x1408af5e0（utf16.Encode 不含 NUL → 小端 []byte → base64）。

**关键实测**：`resolveCmdExePath` 环境变量名是 `"ComSpec"`；`encodePowerShellCommand` 用
`unicode/utf16.Encode`（不含 NUL 终止）；`shellExecuteProgram` 用 ShellExecuteExW（非 ShellExecuteW）。

### 批次 107 后（下一步移交清单）

**当前基线**：`FUNCS=1173 / S=863 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。

**下一批（批次 108）方向**：
- **captureQRCodesFromScreenSelectionForService** 0x140933c40（跨 screenshot 域，[S-sig]
  遗留；调 emitQRCodeDecoded 0x140799d80 → 屏幕选择捕获 + 解码）。
- **startDetachedCommand 调用域扩展**：`openWithDefaultHandler` 0x1408a8de0（[S-sig]，
  ShellExecuteW 链现可用 shellExecuteProgram 复用）。
- **launcherStartedForStartupTray** 0x1408af720 起（开机自启域，含 schtasks / TaskScheduler API）。

### 批次 108（路径/可执行目标判断链 + 提权判定）

**基线**：`FUNCS=1173 / S=863 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1180 / S=871 / S-inline=34 / S-sig=274 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch108.md，`backend/pathclip_logic_windows.go` 新建；
`bootstrap_pathclip.go` newPathError 补体）：
1. `shouldOpenPathViaExplorer` 0x1408a8f20（isProcessElevated && isWindowsExecutableOpenTarget）。
2. `isWindowsExecutableOpenTarget` 0x1408a8f80（SaferiIsExecutableFileType 优先，回退扩展名）。
3. `saferIsExecutableFileType` 0x1408a9000（advapi32.SaferiIsExecutableFileType）。
4. `hasKnownWindowsExecutableExtension` 0x1408a9100（23 项可执行扩展名，含 .appref-ms）。
5. `isWindowsAbsoluteFilesystemPath` 0x1408a9320（UNC / 盘符判定）。
6. `getTokenElevation` 0x1408accc0（GetTokenInformation TokenElevation）。
7. `isProcessElevated` 0x1408ad900（OpenProcessToken + getTokenElevation）。
8. `newPathError`（[S-inline]→[S]，`errors.New("路径不能为空")`）。

**关键实测**：`saferIsExecutableFileType` 用 `SaferiIsExecutableFileType`（非 AssocQueryString）；
扩展名集合 23 项；`newPathError` 是 errorString 对象（非 nil）。

### 批次 108 后（下一步移交清单）

**当前基线**：`FUNCS=1180 / S=871 / S-inline=34 / S-sig=274 / P=1 / UNMARKED=0`。

**下一批（批次 109）方向**：
- **openWithDefaultHandler** 0x1408a8de0（[S-sig]→[S]，依赖链 now 齐：shouldOpenPathViaExplorer
  + startShellTargetViaExplorer + shellExecuteProgram + withShellApartment）。
- **startShellTargetViaExplorer** 0x1408aaee0（openWithDefaultHandler 的 explorer 分支）。
- **revealPathInExplorer** 0x1408a8d20 / **startApplicationViaExplorer** 0x1408a9400
  （资源管理器启动/定位域）。
- **launcherStartedForStartupTray** 0x1408af720 起（开机自启域，含 schtasks / TaskScheduler API）。

### 批次 109（资源管理器启动/定位 + Windows 参数转义）

**基线**：`FUNCS=1180 / S=871 / S-inline=34 / S-sig=274 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1183 / S=874 / S-inline=34 / S-sig=274 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch109.md，`backend/pathclip_explorer_windows.go` 新建；
`bootstrap_pathclip.go` openWithDefaultHandler 补体）：
1. `openWithDefaultHandler` 0x1408a8de0（[S-sig]→[S]：shouldOpenPathViaExplorer →
   startShellTargetViaExplorer(path,nil,"") 或 withShellApartment(shellExecuteProgram("",path,"",""))）。
2. `quoteWindowsArgument` 0x1408ad280（[S]：MS CRT argv 转义，Unicode 感知，加首尾引号）。
3. `revealPathInExplorer` 0x1408a8d20（[S]：explorer /select, 定位）。
4. `startShellTargetViaExplorer` 0x1408aaee0（[S-sig]：签名 (path string, args []string, verb string) error）。

**关键实测**：quoteWindowsArgument 加首尾引号（区别于 syscall.EscapeArg）；openWithDefaultHandler
的 explorer 分支传 (path, nil, "")；startShellTargetViaExplorer 第 2 参数是 []string。

### 批次 109 后（下一步移交清单）

**当前基线**：`FUNCS=1183 / S=874 / S-inline=34 / S-sig=274 / P=1 / UNMARKED=0`。

**下一批（批次 110）方向**：
- **buildShellExecuteArguments** 0x1408ad020（2272B，[]string → 参数组装；startShellTargetViaExplorer
  依赖）。
- **withExplorerShellDispatch** 0x1408ab0e0 / **getDesktopExplorerShellDispatch** 0x1408ab360 /
  **withDesktopExplorerShellDispatch** 0x1408ab2a0 / **shellExecuteByExplorer** 0x1408abe60
  （COM ShellDispatch 链，startShellTargetViaExplorer 完整落地依赖）。
- **revealPathInExplorer 域扩展**：openPathDirectory 0x1408a88a0（依赖 resolveOpenLocationDirectory
  + revealPathInExplorer，后者已落地）。
- **launcherStartedForStartupTray** 0x1408af720 起（开机自启域，含 schtasks / TaskScheduler API）。

### 批次 110（路径定位目录解析 + 资源管理器打开目录）

**基线**：`FUNCS=1183 / S=874 / S-inline=34 / S-sig=274 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1184 / S=876 / S-inline=34 / S-sig=273 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch110.md，`bootstrap_pathclip.go`）：
1. `resolveOpenLocationDirectory` 0x1408a87c0（[S]：目录返身/文件返父/不存在走 looksLikeFilesystemPath）。
2. `openPathDirectory` 0x1408a88a0（[S-sig]→[S]：Stat 成功 IsDir→shellExecuteProgram，非目录→
   revealPathInExplorer；Stat 失败→resolveOpenLocationDirectory→shellExecuteProgram）。

**关键实测**：openPathDirectory 三态分支；错误分支均 errors.New("路径不能为空")；
resolveOpenLocationDirectory 用 internal/filepathlite.Dir。

### 批次 110 后（下一步移交清单）

**当前基线**：`FUNCS=1184 / S=876 / S-inline=34 / S-sig=273 / P=1 / UNMARKED=0`。

**下一批（批次 111）方向**：
- **buildShellExecuteArguments** 0x1408ad020（2272B，[]string → 参数组装；startShellTargetViaExplorer
  依赖）。
- **withExplorerShellDispatch** 0x1408ab0e0 / **getDesktopExplorerShellDispatch** 0x1408ab360 /
  **withDesktopExplorerShellDispatch** 0x1408ab2a0 / **shellExecuteByExplorer** 0x1408abe60
  （COM ShellDispatch 链，startShellTargetViaExplorer 完整落地依赖）。
- **openPathCommandLine** 0x1408a8ac0（[S-sig]，依赖 resolveOpenLocationDirectory + isProcessElevated
  + openUnelevatedCommandLine + resolveCmdExePath + startDetachedCommand）。
- **openPathCommandLineAdmin** 0x1408a8be0（[S-sig]，依赖 resolveCmdExePath + quoteCmdArgument +
  shellExecuteProgram + quoteWindowsArgument + withShellApartment）。
- **launcherStartedForStartupTray** 0x1408af720 起（开机自启域，含 schtasks / TaskScheduler API）。

### 批次 111（命令行打开路径）

**基线**：`FUNCS=1184 / S=876 / S-inline=34 / S-sig=273 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1185 / S=878 / S-inline=34 / S-sig=272 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch111.md，`bootstrap_pathclip.go`）：
1. `openPathCommandLine` 0x1408a8ac0（[S-sig]→[S]：提权→openUnelevatedCommandLine，非提权→
   startDetachedCommand([cmd,"/k","pushd",dir])）。
2. `openPathCommandLineAdmin` 0x1408a8be0（[S-sig]→[S]：runas + "/k pushd "+quoteCmdArgument(dir)）。
3. `openUnelevatedCommandLine` 0x1408aa000（[S-sig]：签名 (dir string) error）。

**关键实测**：错误消息 `位置不能为空`（区别于 newPathError 的 `路径不能为空`）；
openPathCommandLineAdmin verb="runas"；openPathCommandLine 非提权分支用 "/k pushd"。

### 批次 111 后（下一步移交清单）

**当前基线**：`FUNCS=1185 / S=878 / S-inline=34 / S-sig=272 / P=1 / UNMARKED=0`。

**下一批（批次 112）方向**：
- **buildCreateProcessCommandLine** 0x1408acf00 / **createProcessWithShellParentWithVisibility**
  0x1408ac140（进程创建链，openUnelevatedCommandLine 完整落地依赖）。
- **buildShellExecuteArguments** 0x1408ad020（2272B，[]string → 参数组装）。
- **withExplorerShellDispatch** 0x1408ab0e0 / **getDesktopExplorerShellDispatch** 0x1408ab360 /
  **withDesktopExplorerShellDispatch** 0x1408ab2a0 / **shellExecuteByExplorer** 0x1408abe60
  （COM ShellDispatch 链，startShellTargetViaExplorer 完整落地依赖）。
- **launcherStartedForStartupTray** 0x1408af720 起（开机自启域，含 schtasks / TaskScheduler API）。

### 批次 112（进程创建链：命令行组装 + 提权错误判定）

**基线**：`FUNCS=1185 / S=878 / S-inline=34 / S-sig=272 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1187 / S=880 / S-inline=34 / S-sig=272 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch112.md，`backend/process_launch_windows.go` 新建）：
1. `buildCreateProcessCommandLine` 0x1408acf00（[S]：TrimSpace + 引号转义 + 空格拼接）。
2. `isElevationRequiredError` 0x1408ace20（[S]：errors.Is/As 判定 ERROR_ELEVATION_REQUIRED=740）。

**关键实测**：CreateProcess 命令行 = `"escaped exe" + " " + args`，exe 空返回 args；
提权错误 740（0x2e4）。

### 批次 112 后（下一步移交清单）

**当前基线**：`FUNCS=1187 / S=880 / S-inline=34 / S-sig=272 / P=1 / UNMARKED=0`。

**下一批（批次 113）方向**：
- **createProcessWithShellParentWithVisibility** 0x1408ac140（1828B，explorer 作 parent 的
  CreateProcess 链，依赖 openShellProcessForUnelevatedLaunch 0x1408ac980 +
  isProcessHandleElevated 0x1408acb40 + NewProcThreadAttributeList）。
- **retryLaunchAsAdminIfElevationRequired** 0x1408acd60（依赖 isElevationRequiredError +
  shellExecuteProgram）。
- **buildShellExecuteArguments** 0x1408ad020（2272B，[]string → 参数组装）。
- **withExplorerShellDispatch** 0x1408ab0e0 / **getDesktopExplorerShellDispatch** 0x1408ab360 /
  **withDesktopExplorerShellDispatch** 0x1408ab2a0 / **shellExecuteByExplorer** 0x1408abe60
  （COM ShellDispatch 链，startShellTargetViaExplorer 完整落地依赖）。
- **launcherStartedForStartupTray** 0x1408af720 起（开机自启域，含 schtasks / TaskScheduler API）。

### 批次 113（句柄提权判定 + 外壳进程打开）

**基线**：`FUNCS=1187 / S=880 / S-inline=34 / S-sig=272 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1189 / S=882 / S-inline=34 / S-sig=272 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch113.md，`backend/process_launch_windows.go`）：
1. `isProcessHandleElevated` 0x1408acb40（[S]：OpenProcessToken + getTokenElevation）。
2. `openShellProcessForUnelevatedLaunch` 0x1408ac980（[S]：GetShellWindow + OpenProcess(0x1080)
  + isProcessHandleElevated 降权外壳进程）。

**关键实测**：GetWindowThreadProcessId v0.46.0 新签名 (tid,err)；OpenProcess 权限 0x1080=
QUERY_LIMITED|CREATE_PROCESS；错误消息三条（未找到/打开失败/读取失败/已提权）。

### 批次 113 后（下一步移交清单）

**当前基线**：`FUNCS=1189 / S=882 / S-inline=34 / S-sig=272 / P=1 / UNMARKED=0`。

**下一批（批次 114）方向**：
- **createProcessWithShellParentWithVisibility** 0x1408ac140（1828B，explorer 作 parent 的
  CreateProcess 链：NewProcThreadAttributeList + UpdateProcThreadAttribute + STARTUPINFOEX +
  CreateProcess + defer CloseHandle；openUnelevatedCommandLine 完整落地最后一环）。
- **retryLaunchAsAdminIfElevationRequired** 0x1408acd60（8 行源码，依赖 isElevationRequiredError
  + startApplicationWindowsAsAdmin 0x1408a9b80）。
- **buildShellExecuteArguments** 0x1408ad020（2272B，[]string → 参数组装）。
- **withExplorerShellDispatch** 0x1408ab0e0 / **getDesktopExplorerShellDispatch** 0x1408ab360 /
  **withDesktopExplorerShellDispatch** 0x1408ab2a0 / **shellExecuteByExplorer** 0x1408abe60
  （COM ShellDispatch 链，startShellTargetViaExplorer 完整落地依赖）。
- **launcherStartedForStartupTray** 0x1408af720 起（开机自启域，含 schtasks / TaskScheduler API）。

### 批次 114（explorer 作父进程的 CreateProcess 全链闭环）

**基线**：`FUNCS=1189 / S=882 / S-inline=34 / S-sig=272 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1190 / S=884 / S-inline=34 / S-sig=271 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch114.md）：
1. `createProcessWithShellParentWithVisibility` 0x1408ac140（[S]：NewProcThreadAttributeList +
  Update(PROC_THREAD_ATTRIBUTE_PARENT_PROCESS) + StartupInfoEx + CreateProcess）。
2. `openUnelevatedCommandLine` 0x1408aa000（[S-sig]→[S]：resolveCmdExePath + "/k pushd " +
  quoteCmdArgument + buildCreateProcessCommandLine + createProcessWithShellParentWithVisibility）。

**关键实测**：hidden 真 → CREATE_NO_WINDOW|EXTENDED + SW_HIDE，假 → EXTENDED|CREATE_NEW_CONSOLE；
defer 顺序 attrs.Delete → CloseHandle(parent)；desktop="winsta0\default"；成功即 CloseHandle(pi)。

### 批次 114 后（下一步移交清单）

**当前基线**：`FUNCS=1190 / S=884 / S-inline=34 / S-sig=271 / P=1 / UNMARKED=0`。

**下一批（批次 115）方向**：
- **retryLaunchAsAdminIfElevationRequired** 0x1408acd60（8 行源码，依赖 isElevationRequiredError
  + startApplicationWindowsAsAdmin 0x1408a9b80，后者含 buildShellExecuteArguments 链）。
- **buildShellExecuteArguments** 0x1408ad020（2272B，[]string → 参数组装；startApplicationWindowsAsAdmin
  与 startShellTargetViaExplorer 共同依赖）。
- **withExplorerShellDispatch** 0x1408ab0e0 / **getDesktopExplorerShellDispatch** 0x1408ab360 /
  **withDesktopExplorerShellDispatch** 0x1408ab2a0 / **shellExecuteByExplorer** 0x1408abe60
  （COM ShellDispatch 链，startShellTargetViaExplorer 完整落地依赖）。
- **launcherStartedForStartupTray** 0x1408af720 起（开机自启域，含 schtasks / TaskScheduler API）。

### 批次 115（ShellExecute 参数组装）

**基线**：`FUNCS=1190 / S=884 / S-inline=34 / S-sig=271 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1191 / S=885 / S-inline=34 / S-sig=271 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch115.md，`backend/process_launch_windows.go`）：
1. `buildShellExecuteArguments` 0x1408ad020（[S]，608B：TrimSpace + quoteWindowsArgument +
  strings.Join 空格；实为 608B 而非此前估算的 2272B，后者含 quoteWindowsArgument 0x1408ad280）。

### 批次 115 后（下一步移交清单）

**当前基线**：`FUNCS=1191 / S=885 / S-inline=34 / S-sig=271 / P=1 / UNMARKED=0`。

**下一批（批次 116）方向**：
- **startApplicationWindowsAsAdmin** 0x1408a9b80（416B，依赖 buildShellExecuteArguments 已齐 +
  startApplicationUsingCurrentPrivileges + requiresShellOpen + withShellApartment）。
- **retryLaunchAsAdminIfElevationRequired** 0x1408acd60（8 行源码，依赖 isElevationRequiredError +
  startApplicationWindowsAsAdmin）。
- **startApplicationUsingCurrentPrivileges** 0x1408a9f60（提权/降权分流主入口）。
- **withExplorerShellDispatch** 0x1408ab0e0 / **getDesktopExplorerShellDispatch** 0x1408ab360 /
  **withDesktopExplorerShellDispatch** 0x1408ab2a0 / **shellExecuteByExplorer** 0x1408abe60
  （COM ShellDispatch 链，startShellTargetViaExplorer 完整落地依赖）。
- **launcherStartedForStartupTray** 0x1408af720 起（开机自启域，含 schtasks / TaskScheduler API）。

### 批次 116（应用工作目录解析）

**基线**：`FUNCS=1191 / S=885 / S-inline=34 / S-sig=271 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1192 / S=885 / S-inline=34 / S-sig=272 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch116.md，`backend/process_launch_windows.go`）：
1. `resolveApplicationWorkingDirectory` 0x1408a74c0（[S-sig]，288B：entryType=="directory"→""；
  entry 非空→entry；appName 路径→inferWorkingDir）。

**关键取证**：launchContext 结构体第 28 词=appName、第 30 词=entry、第 32 词=args；第 12 词
=entryType；第 14-27 词（7 string）待取证。startApplicationWindowsAsAdmin 与
startApplicationUsingCurrentPrivileges 共享该结构体（第 8+ 词栈参数）。

### 批次 116 后（下一步移交清单）

**当前基线**：`FUNCS=1192 / S=885 / S-inline=34 / S-sig=272 / P=1 / UNMARKED=0`。

**下一批（批次 117）方向**：
- **startApplicationWindowsAsAdmin** 0x1408a9b80（416B：isProcessElevated 分流 → 透传
  startApplicationUsingCurrentPrivileges 或 withShellApartment(shellExecuteProgram("runas",…))）。
- **retryLaunchAsAdminIfElevationRequired** 0x1408acd60（isElevationRequiredError → 透传
  startApplicationWindowsAsAdmin 或返回 err）。
- **startApplicationUsingCurrentPrivileges** 0x1408a7020（116 行：resolveLaunchableAppEntry +
  normalizeAppEntryType + openPathDirectory + isProcessElevated + requiresShellOpen +
  os/exec.Command + resolveApplicationWorkingDirectory + Cmd.Start）。
- **withExplorerShellDispatch** 0x1408ab0e0 / **getDesktopExplorerShellDispatch** 0x1408ab360 /
  **withDesktopExplorerShellDispatch** 0x1408ab2a0 / **shellExecuteByExplorer** 0x1408abe60
  （COM ShellDispatch 链，startShellTargetViaExplorer 完整落地依赖）。
- **launcherStartedForStartupTray** 0x1408af720 起（开机自启域，含 schtasks / TaskScheduler API）。

### 批次 117（应用启动/提权重试链）

**基线**：`FUNCS=1192 / S=885 / S-inline=34 / S-sig=272 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1196 / S=885 / S-inline=34 / S-sig=276 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch117.md，`backend/applaunch_windows.go`）：
1. `startApplicationViaExplorer` 0x1408a9400（[S-sig]，COM explorer 链回退壳）。
2. `startApplicationUsingCurrentPrivileges` 0x1408a7020（[S-sig]，主入口核心控制流）。
3. `startApplicationWindowsAsAdmin` 0x1408a9b80（[S-sig]，UAC runas 提权分流）。
4. `retryLaunchAsAdminIfElevationRequired` 0x1408acd60（[S-sig]，740 重试提权）。

**关键取证**：空 appName 错误消息 = "入口路径不能为空"（24B @0x140c65178）；错误模板
"启动失败: %w"（16B @0x140c551ba）；startApplicationUsingCurrentPrivileges 控制流：
空→directory→(提权? shell/exec)→未提权 explorer。startApplicationWindowsAsAdmin 闭包
func1 用 verb="runas"。

### 批次 117 后（下一步移交清单）

**当前基线**：`FUNCS=1196 / S=885 / S-inline=34 / S-sig=276 / P=1 / UNMARKED=0`。

**下一批（批次 118）方向**：
- **withExplorerShellDispatch** 0x1408ab0e0 / **getDesktopExplorerShellDispatch** 0x1408ab360 /
  **withDesktopExplorerShellDispatch** 0x1408ab2a0 / **shellExecuteByExplorer** 0x1408abe60
  （COM ShellDispatch 链：startShellTargetViaExplorer 完整落地，升格 startApplicationViaExplorer）。
- **resolveLaunchableAppEntry** 0x1408a75e0（launchContext 第 14-27 词 7 string 字段取证，
  升格 startApplicationUsingCurrentPrivileges 为 [S]）。
- **launcherStartedForStartupTray** 0x1408af720 起（开机自启域，含 schtasks / TaskScheduler API）。
- **invokeShellVerb** 0x1408a7020 关联 / **invokeShellVerbWithPowerShell** 0x1408a9e40 起（Shell 动词链）。

### 批次 118（COM ShellDispatch 链）

**基线**：`FUNCS=1196 / S=885 / S-inline=34 / S-sig=276 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1200 / S=885 / S-inline=34 / S-sig=280 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch118.md，`backend/com_explorer_windows.go`）：
1. `getDesktopExplorerShellDispatch` 0x1408ab360（[S-sig]，定位桌面 Explorer Shell.Application）。
2. `withDesktopExplorerShellDispatch` 0x1408ab2a0（[S-sig]，获取→defer Release→fn）。
3. `withExplorerShellDispatch` 0x1408ab0e0（[S-sig]，STA+CoInitialize+委托）。
4. `shellExecuteByExplorer` 0x1408abe60（[S-sig]，Invoke ShellExecute 动词）。

**关键取证**：COM 链 = Windows→FindWindowSW(SWC_DESKTOP)→Item→Document→Application→
QueryInterface；方法名 Windows(7)/FindWindowSW(12)/Document(8)/Application(11)/ShellExecute(12)；
错误消息六段字节级解码。

### 批次 118 后（下一步移交清单）

**当前基线**：`FUNCS=1200 / S=885 / S-inline=34 / S-sig=280 / P=1 / UNMARKED=0`。

**下一批（批次 119）方向**：
- **resolveLaunchableAppEntry** 0x1408a75e0（launchContext 第 14-27 词 7 string 字段取证，
  升格 startApplicationUsingCurrentPrivileges 为 [S]）。
- **launcherStartedForStartupTray** 0x1408af720 / **prepareLauncherStartupTrayProcess** 0x1408af7e0
  （启动托盘进程域）。
- **invokeShellVerb** 0x1408a9de0 / **invokeShellVerbWithPowerShell** 0x1408ada40（Shell 动词链，
  invokeShellVerbWithPowerShell 源码 350 行）。
- **startShellTargetViaExplorer**（把 shellExecuteByExplorer + withExplorerShellDispatch 串起，
  升格 startApplicationViaExplorer 为 [S]）。

### 批次 119（Shell 动词 + 启动托盘进程域）

**基线**：`FUNCS=1200 / S=885 / S-inline=34 / S-sig=280 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1204 / S=888 / S-inline=34 / S-sig=281 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch119.md，`backend/shellverb.go`）：
1. `launcherStartedForStartupTray` 0x1408af720（[S]，参数含 --usbeam-startup-tray 判定）。
2. `prepareLauncherStartupTrayProcess` 0x1408af7e0（[S]，切到可执行文件目录）。
3. `invokeShellVerb` 0x1408a9de0（[S]，properties→属性窗口，其余 PowerShell fallback）。
4. `invokeShellVerbWithPowerShell` 0x1408ada40（[S-sig]，单引号转义 + 模板拼接实证）。

**关键取证**：错误消息 "路径或动作不能为空"（27B @0x140c69f42）；托盘标记
"--usbeam-startup-tray"（21B @0x140c5f818）；动词 "properties"（10B @0x140c441ad）。

### 批次 119 后（下一步移交清单）

**当前基线**：`FUNCS=1204 / S=888 / S-inline=34 / S-sig=281 / P=1 / UNMARKED=0`。
（批次 100–119 已收口。）

**下一批（批次 120 起）方向**：
- **resolveLaunchableAppEntry** 0x1408a75e0 + **resolveLaunchableAppEntryForExplicitArgs**
  （launchContext 第 14-27 词 7 string 字段取证，升格 startApplicationUsingCurrentPrivileges 为 [S]）。
- **invokeShellVerbWithPowerShell** 完整 350 行脚本模板取证（升格 [S]）。
- **syncLauncherStartupTask** 0x1408af840 / **createLauncherStartupTaskUsingSchtasks** /
  **deleteLauncherStartupTask** / **withTaskSchedulerService** 0x1408b14e0（开机自启任务域，
  schtasks + TaskScheduler COM）。
- **startShellTargetViaExplorer**（串起 shellExecuteByExplorer + withExplorerShellDispatch，
  升格 startApplicationViaExplorer 为 [S]）。

### 批次 120（COM ShellExecute 参数序修正 + startShellTargetViaExplorer 升格）

**基线**：`FUNCS=1204 / S=888 / S-inline=34 / S-sig=281 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1204 / S=890 / S-inline=34 / S-sig=279 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch120.md）：
1. `shellExecuteByExplorer` 0x1408abe60（[S-sig]→[S]）：参数序从 batch118 的
   `(verb, appName, args, workingDir, show)` 修正为 `(file, args, dir, verb, show)`，
   对齐 ShellExecute(sFile, vArguments, vDirectory, vOperation, vShow) MSDN 语义。
2. `startShellTargetViaExplorer` 0x1408aaee0（[S-sig]→[S]）：TrimSpace 空 →
   "Shell 目标不能为空"（24B @0x140c651a8）；buildShellExecuteArguments +
   withExplorerShellDispatch 闭包 → shellExecuteByExplorer(dispatch, name, shellArgs,
   dir, "", 1)（verb 恒空，show=1）。

**关键实测**：两独立调用点 asm 寄存器映射一致，推翻 batch118 参数序命名 ——
startShellTargetViaExplorer.func1 (0x1408ab040) 与 startApplicationViaExplorer.func1
(0x1408a9780) 均为 rbx/rcx=file、rdi/rsi=args、r8/r9=dir、r10/r11=verb(空)、[rsp]=show(1)。
`startShellTargetViaExplorer` 第三参实为 workingDir（func1 内 TrimSpace 后作 dir 传参），
非 batch109 注释的 verb。

### 批次 120 后（下一步移交清单）

**当前基线**：`FUNCS=1204 / S=890 / S-inline=34 / S-sig=279 / P=1 / UNMARKED=0`。

**下一批（批次 121 起）方向**：
- **resolveLaunchableAppEntry** 0x1408a75e0 + **resolveLaunchableAppEntryForExplicitArgs**
  0x1408a79e0（launchContext 结构体字段取证，已 dump asm：resolveLaunchableAppEntry.bin/
  resolveLaunchableAppEntryForExplicitArgs.bin 各 1024/1152B，升格 startApplicationUsingCurrentPrivileges）。
- **invokeShellVerbWithPowerShell** 0x1408ada40 完整 350 行脚本模板取证（升格 [S]）。
- **syncLauncherStartupTask** 0x1408af840 / **createLauncherStartupTaskUsingSchtasks**
  0x1408afa60 / **deleteLauncherStartupTask** 0x1408affa0 / **withTaskSchedulerService**
  0x1408b14e0（开机自启任务域，schtasks + TaskScheduler COM）。
- **startApplicationViaExplorer** 0x1408a9400 升格 [S]（asm 已 dump，依赖 launchContext
  结构体 + resolveApplicationWorkingDirectory 结构体签名）。

### 批次 121（launchContext 领域模型立案 + 入口解析函数）

**基线**：`FUNCS=1204 / S=890 / S-inline=34 / S-sig=279 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1207 / S=893 / S-inline=34 / S-sig=279 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch121.md，领域模型变更，走澄清→商榷→立案→执行→复验）：
1. `launchContext` 结构体（360B=0x168，45 词）：字段偏移三函数 asm 交叉实证 —
   entryType@0x20、shellTarget@0xa0、entry@0xb0、args@0xc0、rawName@0x138、rawEntry@0x148、
   rawCommandLine@0x158。中间 30 词 `_ [N]uintptr` 占位（0x00-0x1f / 0x30-0x9f / 0xd8-0x137，
   待 buildAppEntryWithDroppedFiles 取证）。`TestLaunchContextSize` 断言 size==360 锁布局。
2. `shouldUseSavedShortcutResolution` 0x1408a7e60（[S]，128B）：TrimSpace 空→true；
   !isShortcutFilePath→false；os.Stat err→true。
3. `resolveLaunchableAppEntry` 0x1408a75e0（[S]，1024B）：directory/空 rawName/非快捷方式
   三早退 → saved 分支写回 shellTarget/entry/args。
4. `resolveLaunchableAppEntryForExplicitArgs` 0x1408a79e0（[S]，1152B）：先
   resolveLaunchableAppEntry → requiresShellOpen 判定 → 显式 args 分支写回。

**duffcopy 长度公式（实证）**：duffcopy 每条 14B（MOVUPS 3 + ADDQ 4 × 2），64 条；
偏移 0x24c=588=42 条 → 复制 22×16=352B。用于后续结构体大小取证。

### 批次 121 后（下一步移交清单）

**当前基线**：`FUNCS=1207 / S=893 / S-inline=34 / S-sig=279 / P=1 / UNMARKED=0`。

**下一批（批次 122 起）方向**：
- **buildAppEntryWithDroppedFiles** 0x1408a6c20（1024B，asm 已 dump）：补齐 launchContext
  中间 30 词（droppedFiles/privilege/icon 等）语义，把结构体 padding 替换为命名字段。
- **startApplication** 0x1408a6a20（512B，asm 已 dump）：切 launchContext 签名（升格 [S]）。
- **startApplicationUsingCurrentPrivileges** 0x1408a7020 切 launchContext 签名。
- **startApplicationViaExplorer** 0x1408a9400 切 launchContext 签名（asm 已 dump）。
- **invokeShellVerbWithPowerShell** 0x1408ada40 模板取证（duffcopy 源 @0x1409e6ec0，duffcopy+0x292）。
- **开机自启任务域**：syncLauncherStartupTask 0x1408af840 / createLauncherStartupTaskUsingSchtasks
  0x1408afa60 / deleteLauncherStartupTask 0x1408affa0（asm 已 dump，schtasks + TaskScheduler COM）。

### 批次 122（启动权限归一化 + 拖拽文件入口构建）

**基线**：`FUNCS=1207 / S=893 / S-inline=34 / S-sig=279 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1209 / S=895 / S-inline=34 / S-sig=279 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch122.md）：
1. `normalizeAppLaunchPrivilegeMode` 0x140882680 修正 [P]→[S]：补全枚举（len 跳转表实证）—
   admin/runas/administrator→"admin"；standard/normal/unelevated→"standard"；
   follow/launcher/followlauncher/follow-launcher→"followLauncher"；default→"default"；
   空/未知→""（修正旧版「空→default」错误）。
2. `resolveEffectiveAppLaunchPrivilege` 0x140882880（[S]，416B）：configured→fallback→"standard" 三阶回退。
3. `buildAppEntryWithDroppedFiles` 0x1408a6c20（[S]，1024B）：directory/空 paths 早退 →
   resolveLaunchableAppEntryForExplicitArgs + args 合并 cleanStringList。

### 批次 123（应用启动链 launchContext 签名升格）

**基线**：`FUNCS=1209 / S=895 / S-inline=34 / S-sig=279 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1213 / S=904 / S-inline=34 / S-sig=274 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch123.md）：
启动链 9 函数全切 launchContext 签名（字段偏移 entryType@0x20/shellTarget@0xa0/entry@0xb0/args@0xc0）：
- `startApplication` 0x1408a6a20 [S]（顶层分派）
- `startApplicationWindows` 0x1408a9880 [S]（admin/standard/followLauncher 分派）
- `startApplicationWindowsStandard` 0x1408a9a40 [S]
- `startApplicationWindowsAsAdmin` 0x1408a9b80 [S]（runas）
- `startApplicationUsingCurrentPrivileges` 0x1408a7020 [S]（升格）
- `startApplicationViaExplorer` 0x1408a9400 [S]（升格，func1 0x1408a9780/func2 0x1408a9640）
- `startApplicationWithShellParent` 0x1408aa0e0 [S]（CreateProcess 链）
- `retryLaunchAsAdminIfElevationRequired` 0x1408acd60 [S]（升格）
- `resolveApplicationWorkingDirectory` 0x1408a74c0 [S]（升格）

**遗留**：WithShellParent 52B 混淆错误文案（@0x140c59916）[P] 待取证。

### 批次 124（开机自启域：命令输出解码 + 任务缺失判定）

**基线**：`FUNCS=1213 / S=904 / S-inline=34 / S-sig=274 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1216 / S=907 / S-inline=34 / S-sig=274 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch124.md）：
新文件 `backend/startup_task_windows.go`：
- `decodeWindowsBytes` 0x1408b2c40 [S]（MultiByteToWideChar + utf16.Decode）
- `decodeWindowsCommandOutput` 0x1408b2b20 [S]（UTF-8→ACP→GBK 三级回退）
- `isTaskNotFoundMessage` 0x1408b2e60 [S]（6 子串判定，"cannot find" 明文，5 混淆子串 [P]）

**遗留**：schtasks 混淆字符串（任务名/命令名/格式串/判定子串）待统一破解混淆机制。

### 批次 125（开机自启域：XML 定义辅助函数）

**基线**：`FUNCS=1216 / S=907 / S-inline=34 / S-sig=274 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1221 / S=912 / S-inline=34 / S-sig=274 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch125.md）：
新文件 `backend/startup_task_xml_windows.go`：
- `launcherStartupTaskLogonDelay` 0x1408b24c0 [S]（PT%dS）
- `launcherStartupTaskWorkingDirectory` 0x1408b2540 [S]（Dir 推断 + os.Args[0] 回退）
- `quoteWindowsTaskActionCommand` 0x1408b2960 [S]
- `encodeUTF16LEWithBOM` 0x1408b2600 [S]
- `currentLauncherStartupTaskUserID` 0x1408b2a60 [S]（"无法获取当前用户标识"）

### 批次 126（开机自启域：XML 定义主函数 + 字符串混淆假设撤销）

**基线**：`FUNCS=1221 / S=912 / S-inline=34 / S-sig=274 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1224 / S=913 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch126.md）：
- `buildLauncherStartupTaskDefinitionXML` 0x1408b2080 [S-sig]
- `convertXMLFileToUTF16LE` 0x1408b2740 [S]
- `intPtr` [S-inline]（Settings.Priority=7 内联填充）

**重大修正：本 EXE 无字符串混淆**。早前多个批次把字符串标注为 `[P] 混淆` 属
**地址计算错误**：内联 `pefile.get_offset_from_rva` 对 `VirtualSize<SizeOfRawData` 的节
映射错误，读到相邻 runtime 字符串片段。改用 `read_gostring.py`（遍历节用
`max(vsize,rawsz)` 判断）后全部复核为明文。同步修正 batch 123/124 两处：
`startApplicationWithShellParent` 52B → "无法通过安全的 Explorer Shell 启动该目标"；
`isTaskNotFoundMessage` 5 子串 → not found / cannot find the file specified / 找不到 /
不存在 / 系统找不到指定的文件。

**⚠️ 纪律更新**：所有字符串常量一律用 `python tools/go-introspect/read_gostring.py
<exe> <VA:len>` 解码；lea 目标地址必须 `python -c "print(hex(下一条地址 + disp))"`
精确重算，禁止心算/内联 pefile。

### 批次 127（开机自启域：schtasks 命令行创建任务）

**基线**：`FUNCS=1224 / S=913 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1225 / S=914 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch127.md）：
新文件 `backend/startup_task_schtasks_windows.go`：
- `createLauncherStartupTaskUsingSchtasks` 0x1408afa60 [S]（CreateTemp→Write→Close→
  convertXMLFileToUTF16LE→schtasks /Create 全链 + 双 defer 清理）。

### 批次 128（开机自启域：Task Scheduler COM 基础设施）

**基线**：`FUNCS=1225 / S=914 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1231 / S=920 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch128.md）：
新文件 `backend/startup_task_com_windows.go`（6 函数 [S]）：
- `withTaskSchedulerService` 0x1408b14e0（LockOSThread/CoInitialize/CreateObject/
  QueryInterface(IID_ITaskService)/Connect 链 + 4 defer）
- `callTaskSchedulerMethod` 0x1408b1920（InvokeWithOptionalArgs DISPATCH_METHOD + VariantClear）
- `callTaskSchedulerDispatchMethod` 0x1408b1a80（DISPATCH_METHOD + taskSchedulerDispatchFromVariant）
- `getTaskSchedulerDispatchProperty` 0x1408b1b60（DISPATCH_PROPERTYGET + FromVariant）
- `taskSchedulerDispatchFromVariant` 0x1408b1c20（VT_DISPATCH→*IDispatch 所有权转移）
- `putTaskSchedulerProperty` 0x1408b1e00（DISPATCH_PROPERTYPUT 单参 + VariantClear）

IID_ITaskService 用公开 GUID {2FABA4C7-4DA9-4013-9697-20CC3FD40F85}（asm 经 .data 重定位
指针间接装载，磁盘占位非明文）。

### 批次 129（开机自启域：Task Scheduler COM 创建/删除链）

**基线**：`FUNCS=1231 / S=920 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1234 / S=923 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch129.md）：
- `deleteLauncherStartupTask` 0x1408affa0 [S]（`startup_task_schtasks_windows.go`）：
  COM 删除优先、schtasks /Delete fallback、isTaskNotFoundMessage→nil、117B 组合错误；
  回调 func1 0x1409f7e80 [S]（GetFolder→DeleteTask）。
- `createLauncherStartupTaskUsingTaskSchedulerAPI` 0x1408b01c0 [S]（新文件
  `startup_task_com_create_windows.go`）：TrimSpace 空校验 + withTaskSchedulerService(func1)。
  func1 0x1408b0420 [S] 完整 COM 创建链：GetFolder→NewTask→Principal(UserId/RunLevel=1/
  LogonType=enabled?2:3)→Triggers.Create(9, Enabled=true, Delay=logonDelay)→Actions.Create(0,
  Path/Arguments/WorkingDirectory)→Settings→RegisterTaskDefinition("UsbEAm Launcher", task, 6,
  userID, "", logonType)。
- `applyLauncherStartupTaskSettings` 0x1408b10a0 [S]（同文件）：13 项 Settings 属性 +
  IdleSettings 4 项（IdleDuration="PT10M"/WaitTimeout="PT1H"/StopOnIdleEnd=true/
  RestartOnIdle=false）。Settings 表 @0x1411e63a8，Hidden=enabled 由常量表
  {0x1411f3720=0, 0x1411f3728=1} 二选一。

### 批次 129 后（下一步移交清单）

**当前基线**：`FUNCS=1234 / S=923 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。
- **顶层**：syncLauncherStartupTask 0x1408af840（asm 已 dump，COM 路径 vs schtasks 路径调度）。
- **buildLauncherStartupTaskDefinitionXML 模板取证**：duffcopy 源 @0x1411e4ea8 各字段精确
  解引用（当前为 Task Scheduler 标准语义推断 [P]）。

### 批次 130（开机自启域：顶层调度 + XML 模板取证）

**基线**：`FUNCS=1234 / S=923 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1235 / S=925 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch130.md）：
- `syncLauncherStartupTask` 0x1408af840 [S]（新文件 `startup_task_sync_windows.go`）：
  !enabled→delete；delaySeconds 夹取 [10,100]；Executable→Abs→currentLauncherStartupTaskUserID
  →delete→COM 创建→schtasks fallback→65B 组合错误。args="--usbeam-startup-tray"，enabled=false。
- `buildLauncherStartupTaskDefinitionXML` 0x1408b2080 [S-sig]→[S]：模板 @0x1411e4ea0 全字段
  解引用闭合。修正两处标准推断错误：Version="1.2"（非 1.0）；Settings.Enabled 恒 true、
  Settings.Hidden=enabled（原 "Enabled=enabled/Hidden=false" 反转）。

### 批次 130 后（下一步移交清单）

**当前基线**：`FUNCS=1235 / S=925 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。
- 开机自启域收尾：syncLauncherStartupTask 上层调用点（配置保存/热键链路）签名与调用约定核对。
- 其余域专项推进（按 HANDOFF 待移交清单）。

### 批次 131（开机自启域：配置提交链枢纽 syncStartupTask）

**基线**：`FUNCS=1235 / S=925 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1236 / S=926 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch131.md）：
- `(*BootstrapService).syncStartupTask` 0x1407a14c0 [S]：bs==nil→"启动任务服务不可用"（27B
  @0x140c69c33）；lock(+0x540)→读 startupTaskSync(+0x510，func value 1 字)→nil 时默认 funcval
  @0x141096d88（fn=syncLauncherStartupTask 0x1408af840）→call fn(enabled,delaySeconds)。
- 取证：Go 1.25 `func(bool,int) error` sizeof==8（1 字）；startupTaskSync/lock offset 实测
  0x510/0x540 与 asm 一致；xref 8 处调用（SaveConfig 主 1、SaveConfig.func1.1 1、
  ResetConfig.func1/.1 2、commitLauncherConfigReplacementWithWidgets 主 4）。

### 批次 131 后（下一步移交清单）

**当前基线**：`FUNCS=1236 / S=926 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。
- 配置提交链接入：SaveConfig / ResetConfig / commitLauncherConfigReplacementWithWidgets
  需接入 syncStartupTask 调用（8 处调用点），核对 enabled/delaySeconds 实参来源。

### 批次 132（开机自启域：配置比较同步 helper + SaveConfig prepare 取证）

**基线**：`FUNCS=1236 / S=926 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1237 / S=927 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch132.md）：
- `(*BootstrapService).syncStartupTaskFromConfig` [S]（提取自 SaveConfig.func1 内联段
  0x140777618-0x140777725）：读 Preferences.StartupLaunchEnabled(*bool,nil→false) +
  StartupLaunchDelaySeconds(int)，delay clamp[10,100]，触发条件 `newEnabled != oldEnabled ||
  (newEnabled && newDelay != oldDelay)` → syncStartupTask(newEnabled,newDelay)。
- 取证：字段来源锁定 types_config.go L45-46（StartupLaunchEnabled/StartupLaunchDelaySeconds）；
  SaveConfig.func1 prepare 回调全景（空初始校验→sync→18 键特性 map→桌面小部件暂停回滚）。

### 批次 132 后（下一步移交清单）

**当前基线**：`FUNCS=1237 / S=927 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。
- CompareAndSwapPrepared 真实签名重建：asm @0x140777082 为 6 参形态（store、cfg、prepare
  func()、options、两枚 int），现 [S-sig] 简化为单参；需还原 prepare 回调参数。
- SaveConfig / ResetConfig 完整重建：接入 func1/func1.1 prepare 回调（含 syncStartupTaskFromConfig）。

### 批次 133（配置存储域：冲突错误 Error + CompareAndSwapPrepared 签名取证）

**基线**：`FUNCS=1237 / S=927 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1238 / S=928 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch133.md）：
- `launcherConfigConflictError.Error` [S]（0x140898a00）：`fmt.Sprintf("%s: expected=%s
  actual=%s", "CONFIG_REVISION_CONFLICT" 24B @0x140c65118, Expected, Actual)`。
- 取证：CompareAndSwapPrepared 真实签名（revision string + prepare funcval + options +
  withIcons + saveWidgets + cfg 栈参）；体 TrimSpace→CAS 闭包（0x14089b280）→ReplacePrepared；
  CAS 闭包语义 = Revision 比对→冲突错误（期望空/不等→launcherConfigConflictError）。

### 批次 133 后（下一步移交清单）

**当前基线**：`FUNCS=1238 / S=928 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。
- ReplacePrepared 域专项：按 0x140899980 反汇编还原 prepare/commit 闭包真实签名（asm 实证
  prepare 接收 initialized bool 首参），随后 CompareAndSwapPrepared 签名升级。
- SaveConfig / ResetConfig 完整重建（接入 func1 prepare 回调 + CompareAndSwapPrepared 6 参调用）。

### 批次 134（配置存储域：ReplacePrepared prepare 签名校正）

**基线**：`FUNCS=1238 / S=928 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1238 / S=928 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`（签名校正，无新增函数）。

**本批落地**（详见 acceptance/batch134.md）：
- `ReplacePrepared.prepare` 签名校正：`func(bool, LauncherConfig) error`（原单参）。
  asm 实证 prepare 首参 al=initialized（@0x140899c14 前 al=loadUnlocked 返回标志；func1/CAS
  闭包入口 test al，al==0 → "配置尚未初始化"）。体内以 cfg.Initialized 承载（待 loadUnlocked
  签名专项对齐）。
- 取证：ReplacePrepared 完整流程（lock store.mu+0x3c → loadUnlocked → prepare →
  validateTwoFactorStoredConfig → commit → savePreparedUnlocked）。

### 批次 134 后（下一步移交清单）

**当前基线**：`FUNCS=1238 / S=928 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。
- CompareAndSwapPrepared 签名升级：6 参 + 体 TrimSpace→CAS 闭包→ReplacePrepared；需先厘清
  commit 参数真实类型（func1 值传 cfg vs ReplacePrepared commit 指针）。
- loadUnlocked 真实签名专项：asm 显示 3 返回值（cfg+bool+error），现简化为 2 返回值。

### 批次 135（二因素域：validateTwoFactorStoredConfig 签名校正 + 体落地）

**基线**：`FUNCS=1238 / S=928 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1239 / S=929 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch135.md）：
- `validateTwoFactorStoredConfig` [S]（0x1409c6740）：签名校正为
  `(TwoFactorPasswordConfig, []TwoFactorEntryConfig)`；体 KDF/Salt/Verifier base64 校验 +
  entries 循环。sentinel `errTwoFactorDataCorrupted = "TWO_FACTOR_DATA_CORRUPTED"`。
- `validateTwoFactorStoredEntry` [S-sig]（0x1409c6ce0）：签名实证（值传 0xd0B），体骨架。
- 错误串全解码（password.kdf/salt/verifier 无效、password 配置不完整、存在条目但密码配置缺失、
  entries[%d]）。

### 批次 135 后（下一步移交清单）

**当前基线**：`FUNCS=1239 / S=929 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。
- `validateTwoFactorStoredEntry` 体落地（[S-sig]→[S]，0x1409c6ce0, 0x560B）。
- `normalizeTwoFactorPasswordConfig`（0x1409c7240）/ `normalizeTwoFactorEntryConfigs`
  （0x1409c7420）/ `normalizeTwoFactorEntryConfig`（0x1409c77a0）落地，补齐二因素保存规范化链。
- 挂起：CompareAndSwapPrepared 签名升级、loadUnlocked 真实签名专项（见批次 134 后清单）。

### 批次 136（二因素域：validateTwoFactorStoredEntry 体落地）

**基线**：`FUNCS=1239 / S=929 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1239 / S=930 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch136.md）：
- `validateTwoFactorStoredEntry` [S-sig]→[S]（0x1409c6ce0）：Kind(totp/steam)、
  Algorithm(SHA1/256/512)、steam/totp Digits+Period 区间、SecretNonce 12B、SecretCiphertext
  [16,65536]B 校验，全部 %w 包裹 errTwoFactorDataCorrupted。

### 批次 136 后（下一步移交清单）

**当前基线**：`FUNCS=1239 / S=930 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。
- `normalizeTwoFactorPasswordConfig`（0x1409c7240, 0x1e0B）/ `normalizeTwoFactorEntryConfigs`
  （0x1409c7420, 0x380B）/ `normalizeTwoFactorEntryConfig`（0x1409c77a0）落地。
- `sanitizeTwoFactorKind`（0x1409c7ca0）/ `sanitizeTwoFactorEntryIcon`（0x1409c7d40）落地。
- 挂起：CompareAndSwapPrepared 签名升级、loadUnlocked 真实签名专项（见批次 134 后清单）。

### 批次 137（二因素域：sanitizeTwoFactorKind / sanitizeTwoFactorAlgorithm）

**基线**：`FUNCS=1239 / S=930 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1241 / S=932 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch137.md）：
- `sanitizeTwoFactorKind` [S]（0x1409c7ca0）：steam/steamguard→"steam"，其余 "totp"。
- `sanitizeTwoFactorAlgorithm` [S]（0x1409c83e0）：steam 强制 "SHA1"；否则 SHA256/SHA512
  原样返回，其余 "SHA1"。新建 `backend/twofactor_sanitize.go`。

### 批次 137 后（下一步移交清单）

**当前基线**：`FUNCS=1241 / S=932 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。
- `normalizeTwoFactorPasswordConfig`（0x1409c7240, 0x1e0B）/ `normalizeTwoFactorEntryConfigs`
  （0x1409c7420, 0x380B）/ `normalizeTwoFactorEntryConfig`（0x1409c77a0）落地。
- `sanitizeTwoFactorEntryIcon`（0x1409c7d40）/ `sanitizeTwoFactorEntryIconData`（0x1409c7e00）落地。
- `looksLikeTwoFactorBase32Payload`（0x1409c8320）落地。
- 挂起：CompareAndSwapPrepared 签名升级、loadUnlocked 真实签名专项（见批次 134 后清单）。

### 批次 138（二因素域：looksLikeTwoFactorBase32Payload）

**基线**：`FUNCS=1241 / S=932 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1242 / S=933 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch138.md）：
- `looksLikeTwoFactorBase32Payload` [S]（0x1409c8320）：base32 字符集（A-Z/a-z/2-7/=）
  逐 rune 判定，空串 false。

### 批次 138 后（下一步移交清单）

**当前基线**：`FUNCS=1242 / S=933 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。
- `normalizeTwoFactorPasswordConfig`（0x1409c7240, 0x1e0B）/ `normalizeTwoFactorEntryConfigs`
  （0x1409c7420, 0x380B）/ `normalizeTwoFactorEntryConfig`（0x1409c77a0）落地。
- `sanitizeTwoFactorEntryIcon`（0x1409c7d40）/ `sanitizeTwoFactorEntryIconData`（0x1409c7e00）落地。
- 挂起：CompareAndSwapPrepared 签名升级、loadUnlocked 真实签名专项（见批次 134 后清单）。

### 批次 139（二因素域：sanitizeTwoFactorEntryIcon）

**基线**：`FUNCS=1242 / S=933 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1243 / S=934 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch139.md）：
- `sanitizeTwoFactorEntryIcon` [S]（0x1409c7d40）：steam 强制 "steam"；icon TrimSpace 后非空
  原样返回，空则 "lock"。

### 批次 120–139 目标闭环

**基线**：`FUNCS=1243 / S=934 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。
20 批次（120–139）全部验收通过，build/vet/test 全绿，UNMARKED=0 持续保持，P=1
（`newOLEDLifecycleContext`，oledblackout.go，与本轮无关）。

### 批次 140（二因素域：规范化链收尾）

**基线**：`FUNCS=1243 / S=934 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1254 / S=945 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch140.md，均 [S]，位于 `backend/twofactor_normalize.go`）：
- 规范化链 4：`normalizeTwoFactorConfig`（0x1409c65a0）、`normalizeTwoFactorPasswordConfig`
  （0x1409c7240，KDF 默认 `argon2id-v1`）、`normalizeTwoFactorEntryConfigs`（0x1409c7420，
  ToLower(ID) 去重）、`normalizeTwoFactorEntryConfig`（0x1409c77a0，steam digits=5/period=30、
  非 steam digits 4..10/period 5..300 钳制）。
- 图标链 4：`sanitizeTwoFactorEntryIconData`（0x1409c7e00）、`normalizeTwoFactorOpaqueIconPayload`
  （0x1409c7f60）、`normalizeTwoFactorOpaqueIconPayloadWithDecoder`（0x1409c8120，decoder 签名
  `func(string)([]byte,error)` + `http.DetectContentType`）、`resolveTwoFactorDisplayName`（0x1409c84e0）。
- 解码 3：`stripTwoFactorSecretGrouping`（0x1409d51a0）、`decodeTwoFactorBase32Secret`（0x1409d56c0）、
  `decodeTwoFactorBase64Secret`（0x1409d5720，RawStd→RawURL→URL→Std 四变体容错 + `empty secret`）。

### 批次 141（二因素域：密码 / 加解密 / 密钥派生链）

**基线**：`FUNCS=1254 / S=945 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1261 / S=952 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch141.md，均 [S]，位于 `backend/twofactor_crypto.go`）：
- 密钥派生/校验 4：`deriveTwoFactorKey`（0x1409d2320，argon2.IDKey 3/32768/4/32）、
  `buildTwoFactorPasswordVerifier`（0x1409d2400，sha256("usbeam-two-factor-password:"+key)+base64）、
  `isTwoFactorPasswordConfigured`（0x1409d1da0）、`verifyTwoFactorPassword`（0x1409d1e60，
  KDF 须 `argon2id-v1` + salt 16B/verifier 32B base64 + 常量比较）。
- 加解密 3：`encryptTwoFactorSecret`（0x1409d2560，AES-GCM）、`randomBytes`（0x1409d9760）、
  `zeroTwoFactorBytes`（0x1409d97e0）。

**既有缺陷修正**：`launcherconfig.go` `validateTwoFactorStoredConfig` KDF 判定
`"organi2d-v1"` → `"argon2id-v1"`（活体 movabs `0x6469326e6f677261` 实证），连带 [S] 头长度
`0x594B→0x5a0`、`0x53bB→0x560` 校正；`twofactor_validate_test.go` 3 处 KDF 固定值同步。

### 批次 142（二因素域：密钥解密 / 存储态归一化 / 遗留修复链）

**基线**：`FUNCS=1261 / S=952 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1266 / S=957 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch142.md，均 [S]）：
- `decryptTwoFactorSecret`（0x1409d2780，`backend/twofactor_crypto.go`，AES-GCM 解密，与批次 141
  encrypt 形成 round-trip）。
- `backend/twofactor_secret.go` 4：`normalizeStoredTwoFactorSecretBytes`（0x1409d47c0）、
  `repairLegacySteamSecretBytes`（0x1409d48a0，RawStd/RawURL base64 候选 + base32 严格判定 +
  decodeTwoFactorBase32Secret 修复）、`normalizeTwoFactorSecretText`（0x1409d4600，Replacer 删除
  空格/\t/\r/\n）、`looksLikeStrictTwoFactorBase32Secret`（0x1409d4b60，[A-Z]∪[2-7]）。

### 批次 143（二因素域：TOTP/HOTP/Steam 验证码生成链）

**基线**：`FUNCS=1266 / S=957 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1271 / S=962 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch143.md，均 [S]，位于 `backend/twofactor_code.go`）：
- `newTwoFactorHMAC`（0x1409d3400，`sanitizeTwoFactorAlgorithm(algorithm,"totp")` 后按
  "SHA256"("SHA2"+"56")/"SHA512"("SHA5"+"12")/默认 SHA1 选择 `hmac.New`）。
- `buildHOTPValue`（0x1409d3280，RFC 4226 动态截断：8B big-endian counter → HMAC → Sum →
  `offset=sum[len-1]&0xf` → `Uint32(sum[offset:offset+4])&0x7fffffff`）。
- `buildStandardTwoFactorCode`（0x1409d2f80，`mod=10^digits` → `hotp%uint32(mod)` →
  `fmt.Sprintf("%0*d", digits, code)`）。
- `buildSteamTwoFactorCode`（0x1409d30c0，`buildHOTPValue(secret,counter,"SHA1")` → 5 轮
  `v%26` 映射 Steam 字母表 `23456789BCDFGHJKMNPQRTVWXY`(26B @0x141bc6980)，`v/=26`）。
- `buildTwoFactorCode`（0x1409d2c80，顶层分发：period clamp（steam 或越界→30，否则 [5,300]）→
  `counter=now.Unix()/period` → `remaining=period-(unix%period)`（≤0 置 period）→ steam 走
  buildSteamTwoFactorCode，否则 digits clamp([4,10]→6) + sanitizeTwoFactorAlgorithm 后
  buildStandardTwoFactorCode；返回 `(code string, remaining int64, err error)`）。

**逻辑等价校验**：`buildHOTPValue` 通过 RFC 4226 附录 D 全部 10 组向量（counter 0–9，
decimal 31-bit + 6-digit 双列）；`buildStandardTwoFactorCode` 通过 RFC 6238 T=1 官方 8-digit
`94287082`；`buildSteamTwoFactorCode` counter=0 → `GG5F5`；`buildTwoFactorCode` 验证 remaining
与 period/digits clamp。

### 批次 144（二因素域：provision secret 解码链）

**基线**：`FUNCS=1271 / S=962 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1283 / S=974 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch144.md，12 个均 [S]，位于 `backend/twofactor_provisioning.go`）：

- secret-decode 主链 9 个：`decodeTwoFactorSecret`（0x1409d4440，extract 成功取提取文本否则
  原始 text → normalizeTwoFactorSecretText 空报错 → 按 decoder 优先级序列依次解码首个成功）/
  `resolveTwoFactorSecretDecoders`（0x1409d4c20，按 kind/文本特征选 base32/base64/hex 三解码器
  优先级）／`extractTwoFactorProvisioningSecretText`（0x1409d5200，按 scheme=steam/steamguard/
  otpauth 分流取密钥文本）／`resolveSteamProvisioningSecretText`（0x1409d55a0，query→host→path
  三级兜底）／`provisioningUsesSteamSharedSecret`（0x1409d4ea0）／`shouldPreferSteamBase64Secret`
  （0x1409d4e00，shared_secret 或含 `+/=_-` → 优先 base64）／`looksLikeTwoFactorBase32Secret`
  （0x1409d5020）／`looksLikeTwoFactorHexSecret`（0x1409d50e0）／`decodeTwoFactorHexSecret`
  （0x1409d58a0）。
- parse/query 依赖 3 个：`parseTwoFactorProvisioningURL`（0x1409d6ba0，steam:// 裸密钥 host 转
  path 容错后 url.Parse）／`parseTwoFactorRawQuery`（0x1409d7480，先 `+`→`%2B` 再 ParseQuery 保留
  字面 '+'）／`firstNonEmptyQueryValue`（0x1409d83e0，遍历 names 取首非空值）。

**逻辑等价校验**：11 个测试函数覆盖 hex/base32/base64 三种解码、URI 提取（otpauth/steam）、
query '+' 保留、decoder 优先级（totp 首 base32、steam hex 首 hex、steam base32 首 base32）、
错误分支（空文本/乱码）。全量 `go test ./backend` PASS。

### 批次 145（二因素域：provision label/int 辅助纯函数）

**基线**：`FUNCS=1283 / S=974 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1286 / S=977 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

**本批落地**（详见 acceptance/batch145.md，3 个均 [S]，位于 `backend/twofactor_provisioning.go`）：

- `decodeProvisioningLabel`（0x1409d8260，TrimSpace → 去前导 '/' → 空则 "" → url.PathUnescape
  (mode=2) 成功取 TrimSpace 结果，失败回退去 '/' 原文）。
- `splitProvisioningLabel`（0x1409d8300，TrimSpace → 空则 ("","") → strings.Cut(s, ":")
  (0x1411cac58) → 未找到 ("",s)，找到 (TrimSpace(before), TrimSpace(after))）。
- `parseTwoFactorOptionalInt`（0x1409d84c0，strconv.Atoi(TrimSpace(s))，err 则 0）。

**逻辑等价校验**：3 个测试函数覆盖 label 解码（去 '/'、PathUnescape、%ZZ 回退）、':' 拆分
（多冒号取首、无冒号回退）、可选 int（数字/空白/非数字/负）。全量 `go test ./backend` PASS。

**既有 flaky 修复**：全量 `go test ./backend` 曾偶发 `TestConfigureChangeTriggersReschedule`
（memoryrelease 域）失败，根因 `newConfigurableSvc()` 注入过去时刻 + `rescheduleLocked` 的
`time.Until` 真实时钟 → `AfterFunc(0,·)` 立即回调与读 timer 竞态。修复：`newConfigurableSvc()`
注入时刻改为远未来 `2030-01-01`（被测代码 asm 实证，不改；仅测试侧注入值），修复后全量
`go test ./backend -count=1` 连续 3 次 PASS。

### 批次 146（二因素域：twoFactorService 私有辅助方法签名落地）

**基线**：`FUNCS=1286 / S=977 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1302 / S=977 / S-inline=35 / S-sig=289 / P=1 / UNMARKED=0`，
SHA256 `40BFBCC05063D126FE4171E490F4AD81CFDF2E12C0E7EF4DCF761B97F2B71470`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

**本批落地**（详见 acceptance/batch146.md，16 个均 [S-sig]，receiver `*twoFactorService`，
位于 `backend/twofactor_service_helpers.go`）：`buildState`（0x1409d0f40）、
`loadUnlockedConfig`（0x1409d0c80，返回 LauncherConfig+[]byte+uint64+error 四值）、
`buildLockedState`（0x1409d15e0）、`ParseProvisioningText`（0x1409cd160）、
`resolveTwoFactorExportEntryIcon`（0x1409cfc00）、`setSessionKeyForOperation`（0x1409d9900，
gen 在前 key 在后）、`beginSessionOperation`（0x1409d9860）、`sessionKeySnapshot`
（0x1409d9bc0）、`clearSessionKey`（0x1409d9e00）、`clearSessionKeyIfGeneration`（0x1409d9f00）、
`sessionOperationMatches`（0x1409da0e0）、`sessionGenerationMatches`（0x1409da220）、
`previewTimeState`（0x1409da360）、`resolvePreviewCurrentTime`（0x1409da5a0，返回
TwoFactorTimeState+time.Time）、`resolveCurrentTime`（0x1409da820，force bool → 三值）、
`fetchInternetTime`（0x1409db420）。

**签名实证要点**：`loadUnlockedConfig` 返回内存类大结构 `LauncherConfig`（0x126 qword 透传）
+ 会话密钥 `[]byte`(AX/BX/CX) + 代际 `uint64`(DI) + `error`(SI/R8)；`resolveCurrentTime`/
`resolvePreviewCurrentTime` 额外返回 `time.Time`（AX/BX/CX 三寄存器）。体为恒返零值骨架
（TOTP/密码加解密/会话密钥/存储域待整域落地，无臆造逻辑）。全量 `go test ./backend` PASS
（ok 1.451s）。

### 批次 147（qrcode 域：原生屏幕选区/标注 — 上集）

**基线**：`FUNCS=1302 / S=977 / S-inline=35 / S-sig=289 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1361 / S=990 / S-inline=35 / S-sig=335 / P=1 / UNMARKED=0`，
SHA256 `ABEE109245F3C3773A88A9F97B6EF8597911BBBB261C954F68D99339DA25A832`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

**本批落地**（详见 acceptance/batch147.md，59 个，位于 `backend/qrcode_native_a.go`）：13 [S]
（纯几何/数学）+ 46 [S-sig]（Win32/GDI/LazyProc/包级全局依赖）。代表性 [S]：
`shouldRefreshQRCodeControlHover`（4px/90ms）、`shouldReuseQRCodeControlHover`、
`qrCodeControlSelectionFromCachedHover`、`offsetQRCodeAnnotationStroke`、
`unionQRCodeAnnotationRects`、`clampQRCodeAnnotationRect`、`qrCodeAnnotationPointsBounds`、
`qrCodeAnnotationTextEstimatedBounds`。代表性 [S-sig]：`captureQRCodesFromNativeSelection`
（0x1409341e0）、`captureQRCodesFromNativeSelectionResultWithOptions`（0x140934500）、
`newQRCodeScreenSelectionSession`（0x140934f20）、`qrCodeSelectionWindowProc`（0x140937b00）、
`drawQRCodeAnnotationStroke`（0x14094e460）及 draw 系 GDI 函数。

**签名存疑点（已在代码注释标注，待后续补证）**：`parseQRCodeAnnotationHexColor` /
`averageQRCodeAnnotationColor` 返回整型宽度（假定 uint8）、`createQRCodeAnnotationPen` 颜色
字节寄存器分组、`withQRCodeAnnotationTextMeasureHDC` 回调签名、`normalizeQRCodeAnnotationClipboardText`
4 组 Replacer 字面量未解码、`drawQRCodeAnnotationSmallText` 颜色字节序。

### 批次 148（二因素域：包级 URI/entry 构建链签名落地）

**基线**：`FUNCS=1361 / S=990 / S-inline=35 / S-sig=335 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1373 / S=990 / S-inline=35 / S-sig=347 / P=1 / UNMARKED=0`，
SHA256 `D4967527DCA71B84F53E0ED0EE3D240576644D2FCB6D30C18B5C6B99893FFE7E`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

**本批落地**（详见 acceptance/batch148.md，12 个均 [S-sig]，位于 `backend/twofactor_provision_uri.go`）：
build URI 链 `buildTwoFactorProvisioningURI`（0x1409d8520）/`buildSteamProvisioningURI`
（0x1409d8d80）/`buildTOTPProvisioningURI`（0x1409d8660）、`buildTwoFactorLabel`
（0x1409d9360）、图标编解码 `encode/decodeTwoFactorProvisioningIcon`（0x1409d9100/0x1409d92a0）、
`buildTwoFactorDraftFromConfig`（0x1409d7060）、`buildOTPAuthEntryFromProvisioning`
（0x1409d79e0）/`buildSteamEntryFromProvisioning`（0x1409d7540，均 (u *url.URL, query url.Values)
→ (Config, []byte, error)）、`buildTwoFactorDraftPreviewEntry`（0x1409d3c60，三寄存器字经反查
敲定为 now/offset/sampleCount）、`buildTwoFactorEntryStateWithCode`（0x1409d1a40）、
`buildTwoFactorEntryMetadataStates`（0x1409d17c0）。

### 批次 149（qrcode 域：原生屏幕选区/标注 — 下集）

**基线**：`FUNCS=1373 / S=990 / S-inline=35 / S-sig=347 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1432 / S=1021 / S-inline=35 / S-sig=375 / P=1 / UNMARKED=0`，
SHA256 `2ECEDABB37022071C98D73611E32B24662F5B2EFC59FB41E54316EA7FB8B4FAF`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

**⚠️ 30% 达成**：本批将函数覆盖推至 `1432 / 4,754 = 30.12%`，越过 30% 目标线。

**本批落地**（详见 acceptance/batch149.md，59 个，位于 `backend/qrcode_native_b.go`）：31 [S] +
28 [S-sig]。代表性 [S]：`qrCodeAnnotationLinePaintBounds`、`centerRect`、
`qrCodeAnnotationArrowHeadPoints`、`qrCodeCapsulePixelCoverage`/`qrCodeEllipseStrokeDistance`/
`qrCodeEllipseStrokePixelCoverage`/`qrCodeCirclePixelCoverage`/`qrCodeRoundedCornerMaskCoverage`、
`resizeQRCodeSelectionRect`、`buildQRCodeSelectionDirtyRectsWithSizeLabels`、
`blendQRCodeRGBAPixel`/`blendQRCodeRGBAByCoverage`、`fillRGBA`、`subtractQRCodeSelectionRect`、
`qrCodeSelectionSizeLabelText`。`captureQRCodeVirtualScreenSnapshotWithCursor` 提升为 [S]。

**UNMARKED 修复**：子代理 helper `mulByte` 原无 tier 标记致 `UNMARKED=1`，已内联至两处调用点
并删除该函数，恢复 `UNMARKED=0`。

### 批次 150（二因素域：normalize/parse/time 包级函数签名落地）

**基线**：`FUNCS=1432 / S=1021 / S-inline=35 / S-sig=375 / P=1 / UNMARKED=0`。
**收口**：`FUNCS=1444 / S=1021 / S-inline=35 / S-sig=387 / P=1 / UNMARKED=0`，
SHA256 `710B065F58671123F3BB9948DD51F7D255C0F2BC98EE44B56DAFE8B1E3A3CB93`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

**本批落地**（详见 acceptance/batch150.md，12 个均 [S-sig]，位于 `backend/twofactor_provision_misc.go`）：
`buildTwoFactorDuplicateKey`（0x1409d4080）、`allocateTwoFactorEntryID`（0x1409d58e0）、
`normalizeTwoFactorEntryDraft`（0x1409d3640，→ Config+[]byte+error）、
`normalizeTwoFactorEntryMetadata`（0x1409d37c0）、`filterTwoFactorEntriesByIDs`（0x1409d9480）、
`parseTwoFactorProvisioningToken`（0x1409d67e0）、`parseTwoFactorImportText`（0x1409d5c40）、
`extractTwoFactorImportTokens`（0x1409d5f80）、`doTwoFactorTimeRequest`（0x1409dd4e0，
requester 实证 LauncherNetworkAccess）、`fetchTwoFactorTimeSample`（0x1409dc600）、
`isTrustedTwoFactorTimeURL`（0x1409dd720）、`selectTwoFactorTimeSamples`（0x1409dbbe0，
3 条 ret 确认 (cache,error)）。

**二因素域收口**：provision label/URI/entry/parse/import/time 链包级函数已全部落地为 [S]/[S-sig]。

### 后续移交清单（批次 151+）

- 二因素域剩余（parse 链大函数，需先厘清 entry/draft 栈返回布局与 UUID/时间依赖）：
  parseTwoFactorProvisioningToken 0x1409d67e0 / parseTwoFactorProvisioningDraft 0x1409d6dc0 /
  buildTwoFactorDraftFromConfig 0x1409d7060 / buildSteamEntryFromProvisioning 0x1409d7540 /
  buildOTPAuthEntryFromProvisioning 0x1409d79e0；build URI 链（buildTwoFactorProvisioningURI
  0x1409d8520 / buildTOTPProvisioningURI 0x1409d8660 / buildSteamProvisioningURI 0x1409d8d80）。
- `twoFactorService` 12 个大方法（SetupPassword 730L / ChangePassword 625L / ExportEntries 1315L /
  其余 import/export/provisioning），无现成 asm 资产，需现场 dump 0x1409c7ca0 之后的 service
  方法后按依赖链推进。
- 配置存储域挂起：CompareAndSwapPrepared 签名升级（6 参）、loadUnlocked 真实签名专项
  （asm 3 返回值）、func1 prepare 闭包 commit 类型链条厘清。
- 大域备选：filesearch（620 蓝图函数）/ screenshot 剩余（580）/ qrcode（380）/ launcher-system
  （450）/ desktopwidgets（300）。

### 批次 153（gpu 域纯逻辑链：11 函数 [S] 反汇编实证还原）

**基线**：`FUNCS=2626 / S=1041 / S-inline=35 / S-sig=1353 / P=197 / UNMARKED=0`（真函数 2429 = 51.1%）。
**收口**：`FUNCS=2637 / S=1052 / S-inline=35 / S-sig=1353 / P=197 / UNMARKED=0`（真函数 2440 = 51.3%），
SHA256 `5189E3DFEBD9D9243A5EE017E1A7F5BBE6F89A36C66D76AC19842D20D82793E1`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

**本批落地**（详见 acceptance/batch153.md，11 个均 [S]，位于 `backend/gpu.go`）：
`normalizeGPUPreferenceMode`（0x14085a0e0，长度跳表 @0x1411e0940 19 项 → systemDefault/powerSaving/
highPerformance/unknown）、`gpuPreferenceCanonicalSettingKey`（0x14085a360，4 已知键小写化）、
`gpuPreferenceNormalizeSettingValue`（0x14085a4e0，GpuPreference/AutoHDREnable 语义规范化）、
`normalizeGPUPreferenceRegistrySettings`（0x14085a700，ToLower 键去重 + IndexAny("=;") 过滤 +
gpupreference 前移）、`ensureGPUPreferenceRegistrySettings`（0x14085ac40，**前插** {GpuPreference,"0"}，
asm typedslicecopy dst+0x20 右移证得）、`gpuPreferenceModeFromSettings`（0x14085ade0，Atoi 三分支）、
`parseGPUPreferenceRegistryValue`（0x14085af40，Split(";")+Cut("=") 解析）、
`serializeGPUPreferenceRegistrySettings`（0x14085b240，Key=Value; 拼接）、
`normalizeGPUPreferencePath`（0x14085b600，ReplaceAll "/"→"\\" + filepath.Clean）、
`gpuPreferencePathKey`（0x14085b6a0，路径小写键）、`parseGPUPickDebugBool`（0x14085bd40，返回 (bool,bool)）。

**关键实证**：`ensureGPUPreferenceRegistrySettings` 为前插（PREPEND）而非尾插——asm `growslice` 以
newobject 字面量为 1 元素旧 slice、`typedslicecopy` dst=`新指针+0x20` 把旧元素右移一位，对应
`append([]GPUPreferenceSetting{{...}}, s...)`；首项为原始大小写 `GpuPreference`（未过规范化）。
全量 `go test ./backend` PASS（ok 0.402s），GPU 行为黄金用例 8 组全绿。

**下一批**：gpu_windows.go（getGPUPreferenceState 0x14085dda0 等 7 入口，调用本批纯函数）+ gpu_pick_debug.go
剩余（gpuPickDebugLog / ensureGPUPickDebugLogger / resolveGPUPickDebugConfiguration，全局态 + 文件 IO）。

### 批次 154（gpu_pick_debug.go 全量落地：调试日志器 + 配置解析链）

**基线**：`FUNCS=2637 / S=1052 / S-inline=35 / S-sig=1353 / P=197 / UNMARKED=0`（真函数 2440 = 51.3%）。
**收口**：`FUNCS=2640 / S=1055 / S-inline=35 / S-sig=1353 / P=197 / UNMARKED=0`（真函数 2443 = 51.4%），
SHA256 `C341F124FE5CCE88FF2067BE04A51CC0E101CE884D077CB7B9B52254455D4F04`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

**本批落地**（详见 acceptance/batch154.md，3 具名函数 [S]，`backend/gpu_pick_debug.go`）：
`resolveGPUPickDebugConfiguration`（0x14085ba20，返回 (bool,string)，`USBEAM_GPU_PICK_DEBUG_FILE`/
`USBEAM_GPU_PICK_DEBUG` 环境变量 + os.Args[1:] 三分支解析）、`ensureGPUPickDebugLogger`
（0x14085b9e0，sync.Once + 闭包 0x1409f7380：Root 基座路径解析 → MkdirAll(0o755) →
OpenFile(0x441,0o644) → 初始会话日志）、`gpuPickDebugLog`（0x14085b700，ensure → enabled/file
双判 → Sprintf 链 → mu.Lock → defer Unlock → file.Write）。gpu_pick_debug.go 蓝图 6 条目
全部覆盖（+deferwrap1/func1 两编译器生成体）。

**关键实证**：全局态为单例结构体 @0x141c13400（64B：once@0x00 12B / enabled@0x0c /
path@0x10 / file@0x20 / err@0x28 / mu@0x38）；sync.Once 12B 布局证得 enabled 落 0x0c；
`resolveGPUPickDebugConfiguration` 返回值 (bool,string) 由 AX=bool、BX/CX=string 寄存器对敲定。
全量 `go test ./backend` PASS（ok 0.402s），配置解析黄金用例 8 组全绿。

**下一批**：gpu_windows.go（getGPUPreferenceState 0x14085dda0 / addGPUPreferenceEntry 0x14085e920 /
saveGPUPreferenceEntry 0x14085f340 / removeGPUPreferenceEntry 0x14085fde0 /
cleanupMissingGPUPreferenceEntries 0x140860760 / resolveGPUPreferenceAdapterInfo 0x14085d240 /
resolveDXGIAdapterNameByPreference 0x14085d340，Win32 注册表/DXGI，调用批次 153/154 纯函数）+ 排序闭包
`getGPUPreferenceState.sortGPUPreferenceEntries.func1`（0x14085e720）。

> **批次 155–175 记录见 `docs/acceptance/batch155.md`…`batch175.md`（HANDOFF 正文未逐批转录，仅 §1 指标已同步）。**

### 批次 176（多文件 [P]→[S-sig] 签名实证批量：56 函数）

**基线**：`FUNCS=2659 / S=1104 / S-inline=35 / S-sig=1349 / P=171 / UNMARKED=0`（真函数 2488 = 52.4%）。
**收口**：`FUNCS=2659 / S=1104 / S-inline=35 / S-sig=1405 / P=115 / UNMARKED=0`（真函数 2544 = 53.5%），
SHA256 `06B551688111458EE66DC1087ECD5B1BA331CA9BFEC8181F2AA7BC94BA6B7EF4`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

**本批落地**（详见 acceptance/batch176.md，9 文件 56 函数 [P]→[S-sig]，Lead + 5 并行 subagent）：
- Lead 9：oledblackout_windows.go 3 + oledblackout_overlay_windows.go 1 + filesearch_windows.go 1 +
  mousegestures_actions_windows.go 1 + nativedrag_windows.go 1 + audio_windows_runtime.go 2。
- oledblackout_windows.go 7、screenshot_windows.go 11、qrcode_windows.go 6、
  launcherupdate_runtime.go 13、screenshot_uia_windows.go 10（各 subagent）。
- 关键订正：`logFileSearchPerf` 8 参→5 参、`sanitizeDirectoryShortcutName` 补 path 参、
  `SetPersistedDefaultAudioEndpoint` 单参→4 参、`GetValue` 输出 PROPVARIANT 误作入参、
  `mouseGestureWindowMovePosition` int→int32、`screenshotUIAElementBoundsAtPointWithAutomation` 参数序反。

**关键实证**：参数宽度一律以 morestack 序言保存的寄存器为准（dword 运算证 32 位、栈返回区证结构返回）；
签名无法唯一确证的一律保持 `[P]` 并订正注释（filelocator_runtime.go 22 个、screenshot_uia/pin 域 23 个、
launcherupdate 下载族 10 个、oledblackout WinRT/COM 27 个等），不臆造。全量 build/vet/test EXIT=0。

**下一批**：降 P 主攻 filelocator_runtime.go（22 个，需先落地 FileLocator 匹配器/流上下文类型）与
screenshot_uia 深层（11 个，WinRT COM 类型）；落地 100 个未映射原始文件（§10）。

### 批次 177（小文件落地 + 体错/签名订正：8 函数新增 + 4 处订正）

**基线**：`FUNCS=2659 / S=1104 / S-inline=35 / S-sig=1405 / P=115 / UNMARKED=0`（真函数 2544 = 53.5%）。
**收口**：`FUNCS=2667 / S=1111 / S-inline=35 / S-sig=1406 / P=115 / UNMARKED=0`（真函数 2552 = 53.7%），
SHA256 `615C8D00CDCE7F0BB2CEB2E8C739DCF68A45B11CBFD44A7B9D109E0E52ADFB01`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

**本批落地**（详见 acceptance/batch177.md，6 文件新建）：
- 新建 launcherinstance.go / windowmanagement_windows.go / launcheraltmenu_windows.go /
  launcherglobalhotkey.go / webview2_options.go / linkbrowser.go。
- 8 函数新增：`launcherSecondInstanceShouldReveal` [S]、`windowManagementVirtualScreenBounds` [S]、
  `suppressLauncherWindowsAltMenuMessage` [S]、`launcherHotkeyRegistrationError.Error/Unwrap` [S]×2、
  `resolveLauncherWebViewUserDataPath` [S]、`buildLauncherBrowserArgs`/`loadLauncherBrowserArgs` [S-sig]×2。

**关键订正（4 处，不改变 FUNCS 计数）**：
- `normalizeAudioFlowName` 旧体 `(string,error)` + playback/record/all 分支为臆造 → 订正为 `string` 单返回
  （input/capture→"capture"、output/render→"render"、其余→""，.rodata 常量字节级确证）。
- `getLauncherBackgroundMetrics` [S-sig] 空骨架 → [S] 完整体（windowManagementVirtualScreenBounds 宽高判定 + 栈 40B 返回区）。
- `detectLinkBrowsers` 签名 `()` → `() []StartMenuApp`；`ScanLinkBrowsers` 签名 `()` → `() []StartMenuApp` 并迁入 linkbrowser.go。

**差集口径校准（本轮新增）**：实测未落地文件差集 79（非 §1 旧值 100 / §10 旧值 110）→ 本轮 73；
79 文件对应 697 个真正未落地顶层函数（backend 无同名定义），454 个已按域重组散落到别处文件名。
FUNCS 缺口（→4754）的核心是 697 个未落地函数 + 闭包/方法，而非文件名对齐。

**关键教训**：已落地 `[S]` 函数可能混杂臆造体（normalizeAudioFlowName 是实例），后续需抽检历史 `[S]` 标记，
优先复核「错误/解析/归一化」类纯逻辑函数的 switch 分支与返回类型是否与 .rodata 常量一致。

**下一批**：主攻 FUNCS 缺口——落地真正未落地的 697 个顶层函数（按域：filesearch_index_windows.go 224、
twofactor.go 140、launcherprocess_windows.go 89、launcherupdate.go 78 为最大四个整域）；
同时抽检历史 [S] 臆造体（normalizeAudioFlowName 同型错误）。

### 批次 178（链接浏览器/工作区路径/配置校验纯逻辑：13 函数新增）

**基线**：`FUNCS=2667 / S=1111 / S-inline=35 / S-sig=1406 / P=115 / UNMARKED=0`（真函数 2552 = 53.7%）。
**收口**：`FUNCS=2680 / S=1120 / S-inline=35 / S-sig=1410 / P=115 / UNMARKED=0`（真函数 2565 = 53.9%），
SHA256 `BF09B2818A3F254A4A4D4748BD60799EEF2B500F1735A6062B65AD891F44F18A`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

**本批落地**（详见 acceptance/batch178.md，7 文件新建）：
- 新建 desktopwidgets_clock.go / desktopwidgets_store.go / launcherupdate.go / linkbrowser_windows.go /
  launchertray.go / workspacemigration_identity.go / launcherconfiginput.go。
- 9 [S]：`desktopTimezoneOffset`（UTC±HH:MM）、`desktopWidgetStoreError`、`(*launcherUpdateUserError).Error`、
  `resolveRegisteredLinkBrowserPath`/`readRegistryString`/`resolveExistingExecutablePath`/`buildEnvLinkBrowserPath`
  （链接浏览器 App Paths 注册表 + 环境变量解析链）、`workspacePathsEqual`、`validateLauncherConfigInMemoryStructure`。
- 4 [S-sig]：`setupLauncherTray`（Wails SystemTray 菜单构建）、`validateLauncherConfigJSONStructure`、
  `validateLauncherConfigInMemoryBudget`、`validateLauncherConfigForPersistence`。

**关键发现（1 处，未改）**：`validateLauncherConfigIconData` 现落地签名 `(cfg *LauncherConfig)`（无返回）存疑——
`validateLauncherConfigForPersistence` 汇编按值传 2352B 大结构体并透传返回值，疑似应为 `(cfg LauncherConfig) error`，
待单独核验其 asm 订正。

**口径再澄清**：差集 73→66 是文件名对齐指标；本轮新建的 `launcherupdate.go` 是 78 函数大文件，但只落地其中
`(*launcherUpdateUserError).Error` 1 个函数。真实进度以 FUNCS 为准（2667→2680，+13）。

**下一批**：落地 desktopwidgets_audio 域（subagent 2888f722 上下文耗尽未产出，normalizeDesktopReminderAudio 等
6 顶层函数仍待落地）+ filesearch_type_filter/ignore 域（subagent 7a883441 结果待并入）；订正
`validateLauncherConfigIconData` 签名。

### 批次 179（文件搜索类型过滤 + 忽略规则域全量落地：23 函数新增 + 1 升档）

**基线**：`FUNCS=2680 / S=1120 / S-inline=35 / S-sig=1410 / P=115 / UNMARKED=0`（真函数 2565 = 53.9%）。
**收口**：`FUNCS=2703 / S=1144 / S-inline=35 / S-sig=1409 / P=115 / UNMARKED=0`（真函数 2588 = 54.4%），
SHA256 `F9007863FE244E607F76C24A83F03A4B1816CE41F8F3939DE7E3D0597749DC74`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

**本批落地**（详见 acceptance/batch179.md，3 文件新建）：
- 忽略域 filesearch_ignore_windows.go 8 [S]（subagent 7a883441）：normalizeFileSearchIgnoredDirectory{Rule,Path,MatchPath}、
  isFileSearchIgnoredDirectoryPathRule、fileSearchPathWithinIgnoredDirectory、newFileSearchIgnoreMatcher、
  HasRules、IsIgnoredDirectory（双参数 name,path）。
- 类型过滤域 filesearch_type_filter.go 13 [S]（subagent 7a883441）：normalizeFileSearchTypeFilterID/Rules/RuleToken、
  isValidNormalizedExtensionToken/Glob/ASCIICharClass、classifyNormalizedFileSearchTypeRule、
  compileFileSearchRuleMatcher（FNV-1a + glob 正则 + 分卷快路径）、fileSearchRuleMatcherSignature（sha256）、
  Allow/matchSimpleExtension/matchCompoundExtension、fileSearchGlobToAnchoredRegexp。
- 音频域 desktopwidgets_audio.go 2 [S]（Lead）：normalizeDesktopReminderAudio、validateDesktopReminderAudio。

**升档**：`normalizeVolumeRoot`（launcherconfig_runtime.go）[S-sig] 空骨架 → [S] 完整体
（TrimSpace→VolumeName 非 "X:" 返 ""→ToUpper(vol)+"\\"），使忽略域卷根特判分支恢复生效。

**关键订正**：IsIgnoredDirectory 签名 Lead 初判单参数 `(path)`，subagent 经 asm 实证订正为双参数 `(name, path)`；
normalizeFileSearchIgnoredDirectoryPath 补卷根 EqualFold 特判 + TrimRight 尾段。并行落地同域发生写范围重叠，
Lead 初版被 subagent 正确版覆盖，以 asm 实证为准。

**下一批**：还原 defaultFileSearchTypeFilters（≈11KB 静态表）+ defaultFileSearchIgnoredDirectoryRules；
normalizeFileSearchIgnoredDirectoryRules/normalizeFileSearchTypeFilters 由 identity stub 升真实归一化；
订正 validateLauncherConfigIconData 签名。

### 批次 180（归一化链升档 + 提醒音频播放链落地：4 函数新增 + 3 升档）

**基线**：`FUNCS=2703 / S=1144 / S-inline=35 / S-sig=1409 / P=115 / UNMARKED=0`（真函数 2588 = 54.4%）。
**收口**：`FUNCS=2707 / S=1151 / S-inline=35 / S-sig=1406 / P=115 / UNMARKED=0`（真函数 2592 = 54.5%），
SHA256 `EFABF9C3CDD7511242BA319809E0A655D6B180F9F798E1A9F3F7287C8FD8B5BE`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

**本批落地**（详见 acceptance/batch180.md，1 文件新建）：
- 提醒音频播放链 desktopwidgets_audio_windows.go 3 [S]（新建）：playDesktopReminderSound（winmm PlaySoundW）、
  playDesktopReminderSystemSound（回退 user32 MessageBeep）、playDesktopReminderAudioSequence（custom 分支 + 250ms 间隔）。
- 音频定义转换 desktopwidgets_audio.go 1 [S] + 1 类型：desktopReminderAudioFromDefinition + DesktopReminderAudioDefinition
  （Name@0x88/Path@0x98/Volume@0xa8，前置 0x88 字节占位）。

**升档**（filesearch_normalize.go 3 处 identity→[S]）：normalizeFileSearchIgnoredDirectoryRules（绝对/相对双键去重）、
defaultFileSearchIgnoredDirectoryRules（node_modules + os.TempDir + system volume information）、
normalizeFileSearchTypeFilters（ID 空/"all" 跳过 + seen 去重 + Label 64 rune 截断 + enabled 规则）。

**关键订正**：playDesktopReminderSound/SystemSound 的 errors.Is 哨兵为 `windows.ERROR_SUCCESS`（接口 data@0x1411cd520=0x0），
非先前误判 ERROR_INSUFFICIENT_BUFFER（0x7a 为相邻变量）。PlaySound flags：system=0x210002（SND_ALIAS）、custom=0x220002（SND_FILENAME）。

**下一批**：还原 defaultFileSearchTypeFilters（≈11KB 静态表）；desktopwidgets_audio 剩余接口方法
（(*platformDesktopWidgetAudioPlayer)Play / deliverNotification / TestReminderNotification 及其闭包）；
订正 validateLauncherConfigIconData 签名。

### 批次 181（文件搜索归一化链收口：3 处 [S-sig]→[S] 升档）

**基线**：`FUNCS=2707 / S=1151 / S-inline=35 / S-sig=1406 / P=115 / UNMARKED=0`（真函数 2592）。
**收口**：`FUNCS=2707 / S=1154 / S-inline=35 / S-sig=1403 / P=115 / UNMARKED=0`（真函数 2592，升档不改总数），
SHA256 `E136965D8192D9111CD35EBC692E8C88B68FD3E29619496FC8C6739FDB37574F`（`bash build.sh` 重建）。

**本批升档**（详见 acceptance/batch181.md）：
- defaultFileSearchTypeFilters（0x140812580，≈11KB）：return nil 骨架 → 完整 6 过滤器表。
  duffcopy 静态模板 0x1411e5d70 得 ID∈{image,document,video,audio,archive,executable}、Label 全空、Enabled 全 true；
  6 次 growslice（num=64/66/33/39/48/50，元素 string kind=24）逐区解析 300 条扩展名规则；
  archive 含 5 条分卷 glob（*.7z.* 等 + ?*.[rz][0-9][0-9]）。解析脚本 tools/parse_deffilters.py 保留为证据。
- normalizeVolumeRoots（0x14088c6e0）：恒等 → 完整（normalizeVolumeRoot 逐条 + ToLower 去重 + append 原值）。
- normalizeFileSearchRecentItems（0x14087a140）：恒等 → 完整（TrimSpace 空跳过 + RuneCount>32767 跳过 +
  ToLower(Replace(path,"/","\\",-1)) 去重 + Name=trimStringLimit(Name,512) + 上限 128）。

**关键订正/知悉**：defaultFileSearchTypeFilters 返回 6 项（非先前假设的更多项）；archive 规则含 glob 与分卷
特判字符串，与 classifyNormalizedFileSearchTypeRule case 3 精确对应。filesearch 归一化链自此无遗留 stub。

**下一批**：desktopwidgets_audio 剩余接口方法（(*platformDesktopWidgetAudioPlayer)Play / deliverNotification /
TestReminderNotification 及闭包）；订正 validateLauncherConfigIconData 签名（asm 确证按值传 cfg 2352B 栈参数）；
继续落地 truly-missing 顶层函数（697 个）以推进 FUNCS。

### 批次 182（内存释放执行链闭环：4 函数 + 1 错误类型 + 工厂装配修正）

**基线**：`FUNCS=2707 / S=1154 / S-inline=35 / S-sig=1403 / P=115 / UNMARKED=0`（真函数 2592）。
**收口**：`FUNCS=2712 / S=1158 / S-inline=36 / S-sig=1403 / P=115 / UNMARKED=0`（真函数 2597 = 54.6%），
SHA256 `9A2172D378A74D72F85A450F17BCE5F348FDC4CB5A9887EEA1CEB232DFAFD04E`（`bash build.sh` 重建）。

**本批落地**（详见 acceptance/batch182.md，1 文件新建 memoryrelease_exec.go）：
- executeMemoryReleaseMode（0x1408d5120）：parse 失败回退 "standbylist" → resolveMemoryReleaseCommand →
  withMemoryReleasePrivilege("SeProfileSingleProcessPrivilege", func1)。func1 闭包 NtSetSystemInformation(0x50,&cmd,4)，
  STATUS_PRIVILEGE_NOT_HELD（itab 0x1411d2e00/data 0x1411cd478）→ "当前进程没有足够权限…"（72B），其余 → "执行内存释放失败: %w"（28B）。
- resolveMemoryReleaseCommand（0x1408d52c0）：5 个 SYSTEM_MEMORY_LIST_COMMAND 逐字节 cmp 映射
  standbylist→4 / workingsets→2 / modifiedpagelist→3 / priority0standbylist→5；非法 → "不支持的内存释放模式: %s"（34B）。
- withMemoryReleasePrivilege（0x1408d5480）：空名/空 fn 具名错误（42B/30B）+ LockOSThread + UTF16PtrFromString +
  LookupPrivilegeValue + openCurrentProcessToken(0x28) + AdjustTokenPrivileges + defer 恢复（失败 errors.Join）+ fn()。
  三个 %w 包装文案 "准备/查询/启用系统权限失败: %w"（各 28B）+ 恢复文案 "恢复系统权限失败: %w"（28B）。
- adjustMemoryReleaseTokenPrivileges（0x1408d5c00）：newState nil → "缺少系统权限状态"（24B）；buflen=previousState?16:0；
  AdjustTokenPrivileges 失败原样返回；GetLastError==ERROR_NOT_ALL_ASSIGNED → "当前进程不是管理员权限…"（72B）；其余返回 GetLastError。
- memoryReleaseError（16B {msg string}，newobject 0x140b54380）+ Error() [S-inline]。

**工厂装配修正**（memoryrelease.go newMemoryReleaseService）：补齐 `config.Mode="standbylist"` / `s.execute` / `s.now=time.Now`
三处遗漏（asm 0x1408d3560 逐字段对照，此前 execute/now 未装配导致运行期功能断链）。

**关键订正/知悉**：newMemoryReleaseService 的 funcval 装载地址曾因 RIP 相对进位误读（0x1408d6e20 实为
0x141096e20，normalizeGestureButton 内标签 vs time.Now funcval）；确认 now 字段=time.Now、execute=executeMemoryReleaseMode。
resolveMemoryReleaseCommand 返回约定 (rax=cmd,rbx=err.type,rcx=err.data)，execute 侧以 `test rbx,rbx` 判 err。

**下一批**：desktopwidgets_audio 剩余接口方法（(*platformDesktopWidgetAudioPlayer)Play / deliverNotification /
TestReminderNotification 及闭包）；订正 validateLauncherConfigIconData 签名（asm 确证按值传 cfg 2352B 栈参数）；
继续落地 truly-missing 顶层函数（693 个）以推进 FUNCS。

### 批次 183（首页组件提醒音播放/下发链闭环：2 方法 + 闭包）

**基线**：`FUNCS=2712 / S=1158 / S-inline=36 / S-sig=1403 / P=115 / UNMARKED=0`（真函数 2597）。
**收口**：`FUNCS=2714 / S=1160 / S-inline=36 / S-sig=1403 / P=115 / UNMARKED=0`（真函数 2599 = 54.7%），
SHA256 `D7DF6AC7C96C9E517DBDF644FA087F91705DDE0A9F211CBBAA7B9B5F0B090E71`（`bash build.sh` 重建）。

**本批落地**（详见 acceptance/batch183.md，desktopwidgets_audio.go 追加）：
- (*platformDesktopWidgetAudioPlayer)Play（0x1407a6ae0，544B）：normalize + validate 失败即返回；name=="custom" 时
  os.Stat 失败或 info.Mode()&os.ModeType!=0（0x8f280000=ModeType）→ "desktopWidgets.reminderAudioUnavailable"（39B）；
  否则 go func1 异步播放。func1（0x1407a6d00）playback.Lock + defer Unlock + playDesktopReminderAudioSequence，
  失败 log.Printf "播放首页组件提醒音频失败: %v"（40B）。
- (*desktopWidgetService)deliverNotification（0x1407a6f20，736B）：s==nil → notificationFailed（33B）；
  notifier 非空 → Notify(title,message,audio!=nil)，否则 "通知服务不可用"（21B）；audio!=nil 时 Play(*audio)，
  否则 "音频服务不可用"（21B）；notifErr 非空 → log.Printf "发送首页组件通知失败: %v"（34B）+ notificationFailed；
  playErr 非空 → log.Printf "播放首页组件提醒音频失败: %v"（40B）+ reminderAudioFailed（34B）；否则 nil。

**关键订正/知悉**：Play 的 custom 文件类型检查非 `info.IsDir()`，而是 `info.Mode()&os.ModeType!=0`
（itab Fun[2]=Mode() 返回 FileMode，test eax,0x8f280000 逐位分解 = ModeDir|ModeSymlink|ModeNamedPipe|ModeSocket|
ModeDevice|ModeCharDevice|ModeIrregular）。deliverNotification 第 3 参为 `*desktopReminderAudio`（可 nil），
Notify 的 urgent 由 `audio!=nil` 派生（setne 确证）。

**下一批**：落地 `launcherWidgetStore.Read`（0x1407c0420）与 `normalizeDesktopWidgetTitle`（0x1407bef80），
随后闭环 `TestReminderNotification`（0x1407a7320，依赖二者 + map 查找 "reminder"）；订正 validateLauncherConfigIconData
签名（asm 确证按值传 cfg 2352B 栈参数）；继续落地 truly-missing 顶层函数（691 个）以推进 FUNCS。

### 批次 184（标题归一化 + 文件搜索字节折叠匹配：3 函数）

**基线**：`FUNCS=2714 / S=1160 / S-inline=36 / S-sig=1403 / P=115 / UNMARKED=0`（真函数 2599）。
**收口**：`FUNCS=2717 / S=1163 / S-inline=36 / S-sig=1403 / P=115 / UNMARKED=0`（真函数 2602 = 54.7%），
SHA256 `848B5876B06BAC3C185320C4FB5332BCABD70270E8DC193C7EAF297B9FBD6AEC`（`bash build.sh` 重建）。

**本批落地**（详见 acceptance/batch184.md）：
- normalizeDesktopWidgetTitle（0x1407bef80，576B，desktopwidgets_service.go 新建）：TrimSpace + []rune 截断 160 +
  空标题按 kind 回退 9 个英文默认标题（Note/Clock/Timer/Weather/Calendar/Reminder/Stopwatch/World Clock/Widget）。
- bytesContainsFold（0x1407e1260，224B，filesearch_index_windows.go 新建）：大小写折叠的字节子串查找。
- bytesStemEqualFold（0x1407e1340，224B，同文件）：比较 b 主体（最后一个 '.' 之前）与 sub。

**关键知悉**：bytes* 折叠函数第三个参数为 `caseSensitive bool`（false=折叠 true=敏感），经调用点
matchNodeNameTerms（0x1407e3ac0）确证 bool 透传（r9b 保存/恢复，`test r9b,r9b; jne` 跳过折叠）；折叠仅作用于
b（haystack）侧，sub（needle）由调用方保证已小写（asm 只 `or c,0x20` b 侧字节）。

**下一批**：落地 `bytesMatchWildcardFold`（0x1407e3e60，176 行，同文件通配符匹配）；随后闭环存储层依赖链
（launcherWidgetStore.Read/loadUnlocked/readDesktopWidgetStoreBytes/normalizeDesktopWidgetDocument，
desktopwidgets_store.go）；再闭环 TestReminderNotification（normalizeDesktopWidgetTitle 已落地）；继续落地
truly-missing 顶层函数（688 个）以推进 FUNCS。

### 批次 185（启动器更新错误链修正 + 上下文错误升档：2 升档 + 1 订正）

**基线**：`FUNCS=2717 / S=1163 / S-inline=36 / S-sig=1403 / P=115 / UNMARKED=0`（真函数 2602）。
**收口**：`FUNCS=2717 / S=1165 / S-inline=36 / S-sig=1401 / P=115 / UNMARKED=0`（真函数 2602 = 54.7%），
SHA256 `8518D3B53A9F0A930EBC7946A53FB0C096CB1C080E5BE65D112D96C65E931F17`（`bash build.sh` 重建）。

**本批内容**（详见 acceptance/batch185.md）：
- 订正 launcherUpdateUserError.Error：原误用指针接收者（对应 autogenerated 符号），改为值接收者
  `func (e launcherUpdateUserError) Error() string { return string(e) }`（值接收者 0x1408b7f00 + 指针包装
  0x140a05c40 共同确证 string 别名）。
- 升档 launcherUpdateContextError（0x1408b7f20）：ctx==nil 或 ctx.Err()==nil→nil；否则
  errors.New(TrimSpace("launcherUpdate.cancelled" /*24B*/))。
- 升档 normalizeLauncherUpdateContextError（0x1408b7fa0）：e=launcherUpdateContextError(ctx) 非空→e；
  errors.Is(err, context.Canceled)→"launcherUpdate.cancelled"；否则 return err。

**下一批**：落地 `bytesMatchWildcardFold`（0x1407e3e60，176 行，同文件通配符匹配）；随后闭环存储层依赖链
（launcherWidgetStore.Read/loadUnlocked/readDesktopWidgetStoreBytes/normalizeDesktopWidgetDocument，
desktopwidgets_store.go）；再闭环 TestReminderNotification（normalizeDesktopWidgetTitle 已落地）；继续落地
truly-missing 顶层函数（688 个）以推进 FUNCS。

### 批次 186（首页组件归一化链 + 存储文档归一化：4 函数 → [S]）

**基线**：`FUNCS=2717 / S=1165 / S-inline=36 / S-sig=1401 / P=115 / UNMARKED=0`（真函数 2602）。
**收口**：`FUNCS=2721 / S=1169 / S-inline=36 / S-sig=1401 / P=115 / UNMARKED=0`（真函数 2606 = 54.8%），
SHA256 `131E08CCA0653D9B376B6CC69880FB7DC05E13209AEFBCAB9E8AD1078AF40C97`（`bash build.sh` 重建）。

**本批内容**（详见 acceptance/batch186.md，desktopwidgets_service.go + desktopwidgets_store.go）：
- `normalizeDesktopReminderScheduleKind`（0x1407bf740）：water/sedentary → "interval"，否则 TrimSpace 原样。
- `normalizeDesktopReminderTimezone`（0x1407bf220）：kind=="reminder" 且 cfg.Reminder 非空才归一化；拷贝 def，
  ScheduleKind 归一 + Preset→"custom" + TimezoneID→"Local" + 音频归一化，cfg.Reminder=&def。
- `normalizeDesktopWidgetConfig`（0x1407bf5a0）：先走 reminder 归一，kind=="clock" 时 HourFormat 归一到 "12"/"24"。
- `normalizeDesktopWidgetDocument`（0x1407bfec0）：Version==0→1；九个 map 字段 nil→空 map；遍历 Widgets 归一
  Config、遍历 Reminders 归一 ScheduleKind。

**下一批**：闭环存储层读取链 `launcherWidgetStore.Read`（0x1407c0420，已解析：nil→"首页组件存储不可用"）/
`loadUnlocked`（0x1407c2440，依赖 readDesktopWidgetStoreBytes/cloneDesktopWidgetDocument/validateDesktopWidgetDocument/
ensureDesktopWidgetJSONEOF）；`bytesMatchWildcardFold`（0x1407e3e60，176 行）；继续落地 truly-missing 顶层函数（684 个）。

### 批次 187（组件存储读取链底层：3 函数 → [S]）

**基线**：`FUNCS=2721 / S=1169 / S-inline=36 / S-sig=1401 / P=115 / UNMARKED=0`（真函数 2606）。
**收口**：`FUNCS=2724 / S=1172 / S-inline=36 / S-sig=1401 / P=115 / UNMARKED=0`（真函数 2609 = 54.9%），
SHA256 `A1D0AD9BAB7AC3693CCFFCB127C3F9149D490E7C603D32CFE0F2F4B8E63BC7AA`（`bash build.sh` 重建）。

**本批内容**（详见 acceptance/batch187.md，desktopwidgets_store.go，loadUnlocked 依赖链底三层）：
- `readDesktopWidgetStoreBytes`（0x1407c4280）：读文件 32MB 上限 + 普通文件校验 + 超限校验。
- `ensureDesktopWidgetJSONEOF`（0x1407c4e20）：JSON 末尾校验，多个顶层值/尾部无效分别报错。
- `cloneDesktopWidgetDocument`（0x1407c34e0）：JSON 序列化深拷贝，失败回退零值文档。

**下一批**：落地 `validateDesktopWidgetDocument`（0x1407c3940，2368B，多 map 遍历校验）后闭环
`launcherWidgetStore.Read`（0x1407c0420）/`loadUnlocked`（0x1407c2440）；`bytesMatchWildcardFold`（0x1407e3e60，
176 行）；继续落地 truly-missing 顶层函数（681 个）。

### 批次 188（组件存储读取链闭环：2 [S] + 1 [S-sig]）

**基线**：`FUNCS=2724 / S=1172 / S-inline=36 / S-sig=1401 / P=115 / UNMARKED=0`（真函数 2609）。
**收口**：`FUNCS=2727 / S=1174 / S-inline=36 / S-sig=1402 / P=115 / UNMARKED=0`（真函数 2612 = 55.0%），
SHA256 `10F7943CA8774DBF67B8EDEE18DE704F745AC39977C9F60FAD931EA55190122B`（`bash build.sh` 重建）。

**本批内容**（详见 acceptance/batch188.md，desktopwidgets_store.go 读取链闭环）：
- `loadUnlocked`（0x1407c2440）[S]：读文件→结构校验→JSON 解析→末尾校验→归一化→文档校验→缓存，
  返回 (bool,doc,error)，三态缓存字段写入点逐一对齐 asm。
- `Read`（0x1407c0420）[S]：nil→"首页组件存储不可用"，加锁 defer 解锁，调 loadUnlocked 返回 (exists,err)。
- `validateDesktopWidgetDocument`（0x1407c3940）[S-sig]：签名还原，多 map 校验体待逐项字节级还原。

**下一批**：`validateDesktopWidgetDocument` 升档 [S-sig]→[S]；写入侧 `launcherWidgetStore.Ensure`
（0x1407c06e0）/`writeDesktopWidgetDocumentFile`（0x1407c47c0）；`bytesMatchWildcardFold`（0x1407e3e60，176 行）；
继续落地 truly-missing 顶层函数（678 个）。

### 批次 189（组件存储写入链闭环：2 [S] + 1 [S-sig]）

**基线**：`FUNCS=2727 / S=1174 / S-inline=36 / S-sig=1402 / P=115 / UNMARKED=0`（真函数 2612）。
**收口**：`FUNCS=2730 / S=1176 / S-inline=36 / S-sig=1403 / P=115 / UNMARKED=0`（真函数 2615 = 55.0%），
SHA256 `9F2C5732CA904E77EB533197669BCF4C1699AAF9D056E3EAA38276FC287B5D3E`（`bash build.sh` 重建）。

**本批内容**（详见 acceptance/batch189.md，desktopwidgets_store.go 写入链闭环）：
- `Ensure`（0x1407c06e0）[S]：nil→"首页组件存储不可用"，加锁 defer 解锁，不存在则零值+UTC RFC3339Nano 写盘。
- `writeUnlocked`（0x1407c3260）[S]：normalize→validate→writeDocument（默认 writeDesktopWidgetDocumentFile）→缓存。
- `writeDesktopWidgetDocumentFile`（0x1407c47c0）[S-sig]：原子写结构还原（MkdirAll 0o755 / MarshalIndent /
  CreateTemp / rename 重试 ≤5 次 + 50ms），重试条件 e1/e2 待动态确认。

**下一批**：`writeDesktopWidgetDocumentFile` 与 `validateDesktopWidgetDocument` 升档 [S]；写入侧
`launcherWidgetStore.Update`（0x1407c1100，1600B）/`ReplaceWithRollback`（0x1407c17a0，3136B）；
`bytesMatchWildcardFold`（0x1407e3e60，176 行）；继续落地 truly-missing 顶层函数（678 个）。

### 批次 190（组件存储 CRUD 完整闭环：Update/ReplaceWithRollback 升档 [S]，原子写升档 [S]）

**基线**：`FUNCS=2730 / S=1176 / S-inline=36 / S-sig=1403 / P=115 / UNMARKED=0`（真函数 2615）。
**收口**：`FUNCS=2732 / S=1179 / S-inline=36 / S-sig=1402 / P=115 / UNMARKED=0`（真函数 2617 = 55.1%），
SHA256 `060D52BE6A3B13E7BB62592874C27AB278DDD148805343443167B00F418078E7`（`bash build.sh` 重建）。

**本批内容**（详见 acceptance/batch190.md，desktopwidgets_store.go CRUD 闭环）：
- `Update`（0x1407c1100）[S]：深拷贝→mutate 回调→Revision+1→UTC 时间戳→writeUnlocked→返回副本。
- `ReplaceWithRollback`（0x1407c17a0 + func1@0x1407c1e60）[S]：归一化+校验→writeDocument→缓存→返回回滚闭包。
- `writeDesktopWidgetDocumentFile`（0x1407c47c0）[S-sig]→[S]：原子写完整还原（含 rename 重试条件 os.ErrPermission/os.ErrExist）。

**下一批**：`validateDesktopWidgetDocument` 升档 [S]（多 map 校验逐条还原）；`launcherWidgetStore.Delete`
升档 [S]；`bytesMatchWildcardFold`（0x1407e3e60，176 行）；继续落地 truly-missing 顶层函数（678 个）。

### 批次 191（组件存储全方法 [S] 闭环：Delete + validateDesktopWidgetDocument 升档）

**基线**：`FUNCS=2732 / S=1179 / S-inline=36 / S-sig=1402 / P=115 / UNMARKED=0`（真函数 2617）。
**收口**：`FUNCS=2732 / S=1181 / S-inline=36 / S-sig=1400 / P=115 / UNMARKED=0`（真函数 2617 = 55.0%），
SHA256 `731D447F3326FDD4C4230A06ACD1343FF000E20C12D4E7D6BED72DBFEF97D654`（`bash build.sh` 重建）。

**本批内容**（详见 acceptance/batch191.md，组件存储全方法 [S]）：
- `launcherWidgetStore.Delete`（0x1407c0d20）[S-sig]→[S]：nil/空路径错误 + os.Remove + 缓存重置默认文档。
- `validateDesktopWidgetDocument`（0x1407c3940）[S-sig]→[S]：2368B 校验链完整还原（版本/修订号/组件数上限、
  Widgets 白名单 8 值 + Title/Revision/Config 校验、Notes 字节+rune+累计、StopwatchLaps 单条+累计、终 Marshal
  ≤0x2000000 → validateLauncherConfigJSONStructure）。

**下一批**：`bytesMatchWildcardFold`（0x1407e3e60，176 行）；`TestReminderNotification`（0x1407a7320）；
`validateLauncherConfigIconData`（by-value 2352B）；`desktopCalendarStringList`（0x1407aa820）；
`pluginid.go`（2 funcs）；继续落地 truly-missing 顶层函数（678 个）与 §10 差集 60 文件。

### 批次 192（pluginid + pluginsecurity 全落地：8 函数 [S]）

**基线**：`FUNCS=2732 / S=1181 / S-inline=36 / S-sig=1400 / P=115 / UNMARKED=0`（真函数 2617）。
**收口**：`FUNCS=2740 / S=1189 / S-inline=36 / S-sig=1400 / P=115 / UNMARKED=0`（真函数 2625 = 55.2%），
SHA256 `252F195D356BF353D3C50CBCEF724920653D0BC307FE49F7130C053C225DCB22`（`bash build.sh` 重建）。

**本批内容**（详见 acceptance/batch192.md，插件 ID/资源路径安全链闭环）：
- `pluginid.go`（原 `changeme/internal/pluginid` 并入 main，2 funcs）：`ValidationError{Reason;ReservedName}` +
  `Error()` + `Validate()`（Reason 1–9 校验顺序逐条 asm 对齐，22 组临时用例全通过后删除）。
- `pluginsecurity.go`（6 funcs 全部 [S]）：`validatePluginID`（Reason→中文消息）、`normalizeValidatedPluginID`、
  `readPluginFileBounded`（O_RDONLY + 256KB 上限）、`readPluginManifestBounded`（JSON 解码 + 尾随内容拒 +
  validatePluginID）、`isPathWithinPluginRoot`（Rel 越界判定）、`resolvePluginPathSecurely`（绝对化 + 逐段
  Lstat 拒 symlink/reparse + EvalSymlinks 复核 + 目录/普通文件终判）。

**下一批**：`pluginhost.go`（12 funcs）、`pluginupdate.go`、`pluginwindow.go` 插件域未落地文件；
`bytesMatchWildcardFold`（0x1407e3e60，176 行）；`TestReminderNotification`（0x1407a7320）；
`validateLauncherConfigIconData`；`desktopCalendarStringList`（0x1407aa820）；
继续落地 truly-missing 顶层函数（678 个）与 §10 差集 58 文件。

### 批次 193（pluginhost 全落地：10 函数 [S] + 3 依赖）

**基线**：`FUNCS=2740 / S=1189 / S-inline=36 / S-sig=1400 / P=115 / UNMARKED=0`（真函数 2625）。
**收口**：`FUNCS=2753 / S=1201 / S-inline=36 / S-sig=1400 / P=116 / UNMARKED=0`（真函数 2637 = 55.5%），
SHA256 `23823CC4C32AEF4452BD35DEB8360F55CA8D1DCEBB9D44C59F6AC2E8D3D03810`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

**本批落地**（详见 acceptance/batch193.md）：
- `pluginhost.go` 新建 10 函数全部 [S]：`ServeHTTP`（0x140921b80，RLock + 405/400/404/500 路由）、
  `resolvePluginManifestUnlocked`（0x140922300）、`parsePluginRequestPath`（0x1409225c0）、
  `resolvePluginAssetPath`（0x140922800）、`isHTMLPluginAsset`（0x1409229e0）、`servePluginHTML`（0x140922aa0）、
  `injectPluginRuntime`（0x140922f40）、`resolvePluginBaseHref`（0x140923840）、`buildPluginRuntimeSnippet`
  （0x140923a00）、`urlPathEscape`（0x140923ae0）。
- `plugin_runtime_snippet.html` 新建（40370B 模板，2 个 `%s` + 26 个 `%%`，`//go:embed` 内嵌，与 rodata
  @0x140d0c43a 字节一致；`& < > " '` 五对 HTML 转义 Replacer）。
- `bootstrapservice_callees.go` 追加 3 依赖：`pluginDirectories` [S]（0x1407a2860）、
  `resolvePluginLanguageMessages` [S]（0x1407a4f80）、`discoverPluginsUnlocked` [P]（0x1407a54a0，2432B，
  插件发现域专项，当前返回空切片）。

**关键知悉**：`ServeHTTPdeferwrap1`（0x1409222a0）为 `defer mu.RUnlock()` 编译器自动生成包装，不手写；
`resolvePluginBaseHref` 的 Rel 参数序为 `filepath.Rel(srcDir, path)`（asm 0x14092388c 实证，非反序）；
全局插件发现锁 `pluginDiscoveryLock sync.RWMutex` 由 discoverPlugins（0x1407a535f）与 ServeHTTP（0x140921bdc）共享。

**下一批**：`discoverPluginsUnlocked` 升级 [S]（os.ReadDir + sort + 清单解析链 readPluginManifestBounded/
normalizePluginIconFile/resolvePluginLocalIconURL/loadPluginLocalizations）；`pluginupdate.go`、`pluginwindow.go`
插件域未落地文件；`bytesMatchWildcardFold` 0x1407e3e60、`TestReminderNotification` 0x1407a7320、
`validateLauncherConfigIconData`、`desktopCalendarStringList` 0x1407aa820；继续落地 truly-missing 顶层函数
（678 个）与 §10 差集 57 文件。

### 批次 194（插件发现域全落地：8 函数 [S] + discoverPluginsUnlocked 升 [S]）

**基线**：`FUNCS=2753 / S=1201 / S-inline=36 / S-sig=1400 / P=116 / UNMARKED=0`（真函数 2637）。
**收口**：`FUNCS=2761 / S=1210 / S-inline=36 / S-sig=1400 / P=115 / UNMARKED=0`（真函数 2646 = 55.7%），
SHA256 `EA917E6F8DCDE7536A97B89ED47F76343A9E88C4764E1F5A5E8F9008AE04C82B`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,279,744 B，`bash build.sh` 重建）。

**本批落地**（详见 acceptance/batch194.md）：
- `plugin_discovery.go` 新建 8 函数全部 [S]：`normalizePluginIconFile`（0x14092bb20）、
  `resolvePluginLocalIconURL`（0x14092c040）、`mergeManifestLocalizations`（0x1407a6380）、
  `loadPluginLocalizations`（0x1407a5ec0）、`parseLanguageContent`（0x1407a3bc0）、
  `parseLanguageScanner`（0x1407a3e00，2334B）、`decodeLanguageValue`（0x1407a4760）、
  `setNestedLanguageMessage`（0x1407a4940）。
- `bootstrapservice_callees.go`：`discoverPluginsUnlocked`（0x1407a54a0，2432B）[P]→[S]
  （ReadDir 逐目录 → 跳过非目录/symlink/`.` 前缀 → plugin.json 解析 → Entry 兜底 index.html →
  SourceDir/Installed → normalizePluginIconFile + readPluginFileBounded(1<<20) 验证 →
  resolvePluginLocalIconURL → loadPluginLocalizations → ToLower(TrimSpace(ID)) 去重 → sort.Slice 按 ID）。
  新增 import os/sort。
- `nestedLanguageMessageString` 补 TrimSpace(key)/TrimSpace(part)+空 part 返回 ""（对齐 asm 0x1407a6500）。

**关键知悉**：parseLanguageScanner 的 `EqualFold(section,"meta")` 门控 code/name/native_name 三键分派，
但 meta 键值对同时落入 `setNestedLanguageMessage(messages, section+"."+key, value)`（asm jmp 0x1407a450f）；
locale 目录常量 `language`、扩展名 `.ini`、段落 `meta`、key 字面量 code/name/native_name、兜底 `en-US`、
格式 `/plugin-host/%s/%s`、BOM `\xef\xbb\xbf` 全部 resolve_lea_targets.py 解码确证。

**下一批**：`pluginupdate.go`、`pluginwindow.go` 插件域未落地文件（normalizePluginIconFile/
resolvePluginLocalIconURL 已从 pluginupdate.go 蓝图迁出落地）；`bytesMatchWildcardFold` 0x1407e3e60、
`TestReminderNotification` 0x1407a7320、`validateLauncherConfigIconData`、
`desktopCalendarStringList` 0x1407aa820；继续落地 truly-missing 顶层函数（678 个）与 §10 差集 57 文件。

### 批次 195（文件搜索通配符折叠匹配：bytesMatchWildcardFold [S]）

**基线**：`FUNCS=2761 / S=1210 / S-inline=36 / S-sig=1400 / P=115 / UNMARKED=0`（真函数 2646）。
**收口**：`FUNCS=2762 / S=1211 / S-inline=36 / S-sig=1400 / P=115 / UNMARKED=0`（真函数 2647 = 55.7%），
SHA256 `D9A9879424FB2E100AC6B782A9B6E5168EBD3EB9CBF46AD8B034498FC5927CD0`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,279,744 B，`bash build.sh` 重建）。

**本批落地**（详见 acceptance/batch195.md）：
- `bytesMatchWildcardFold`（0x1407e3e60，480B）[S] 追加到 `filesearch_index_windows.go`
  （蓝图 filesearch_index_windows.go Lines 936–1112，176 行源码）。签名
  `(b []byte, patterns [][]byte, caseSensitive bool) bool` 经调用点 matchNodeNameTerms
  （0x1407e3ac0）三处 call 交叉确证。
- 语义：len(patterns)==0 → true；len==1 → bytesContainsFold 退化；否则顺序遍历 patterns，
  空片段跳过，每片段在 b 中找首个匹配起点（折叠仅作用 b 侧），找到则 b=b[k+len(p):] 继续，
  找不到 → false；全部匹配 → true。
- 新增 `filesearch_index_windows_test.go` 6 用例（空 patterns/单 pattern 折叠与敏感/顺序拼接/
  中间缺失/空片段跳过/首匹配消耗语义）全 PASS。

**下一批**：同文件 `matchNodeNameTerms`（0x1407e3ac0，928B，本批调用者）与
`matchNodeNameTermsWithPinyin`（0x14080fe40）；`pluginupdate.go`、`pluginwindow.go`；
`TestReminderNotification` 0x1407a7320、`validateLauncherConfigIconData`、
`desktopCalendarStringList` 0x1407aa820；继续落地 truly-missing 顶层函数（677 个）与 §10 差集 57 文件。

### 批次 196（screenshot UIA 纯几何辅助 4 函数 [S]）

**基线**：`FUNCS=2762 / S=1211 / S-inline=36 / S-sig=1400 / P=115 / UNMARKED=0`（真函数 2647）。
**收口**：`FUNCS=2762 / S=1215 / S-inline=36 / S-sig=1397 / P=114 / UNMARKED=0`（真函数 2648 = 55.7%），
SHA256 `9DC98F73905D4AD5F8D512D4DC79C8979B5E786BA64D6F0CED11350CFE354DF6`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,279,744 B，`bash build.sh` 重建）。

**本批落地**（screenshot_uia_windows.go，详见 acceptance/batch196.md）：
- `screenshotRectMatchesBounds`（0x1409af320，224B）[S-sig]→[S]：rect 空或 bounds 空 → false；
  rect==bounds → true；rect 包含 bounds → true；否则四边 |差值|<=2 → true，任一边 >2 → false。
- `screenshotUIAElementHitContainsPoint`（0x1409b2c60，128B）[P]→[S]：**签名修正为 8 槽**
  `(rect image.Rectangle, controlType uint32, ok bool, x, y int) bool`（esi/r8b 经调用点 0x1409b2aaa/
  0x1409b2ab1 与 readScreenshotUIAElementHitInfo 返回 `(rect,uint32,bool,error)` 交叉确证）；
  体：rect.Dx()<2 || rect.Dy()<2 → false；否则 rect.Min.X<=x && x<rect.Max.X && rect.Min.Y<=y &&
  y<rect.Max.Y（controlType/ok 未用）。
- `screenshotPreferSmallerControlRect`（0x1409aedc0，96B）[S-sig]→[S]：a/b 空 → false；面积
  （Dx*Dy）<=0 → false；返回 bArea < aArea（偏好面积更小）。
- `screenshotPointNearWindowEdge`（0x1409af2a0，128B）[S-sig]→[S]：rect 空或点不在 rect 内 →
  false；threshold<0 钳 0；返回点到 rect 四边最近距离 <= threshold。

**关键知悉**：`screenshotUIAElementHitContainsPoint` 旧 [P] 存根签名 `(rect, x, y int)` 是错的
（少 controlType/ok 两槽），本批依据调用点实参寄存器 + readScreenshotUIAElementHitInfo 返回交叉
实证后修正为 8 槽，未臆造。

**下一批**：继续 screenshot_uia_windows.go 相邻纯几何 `normalizeScreenshotControlRect`
（0x1409af400）、`screenshotUIAControlRectUsable`（0x1409af4e0）；`matchNodeNameTerms`、
`pluginupdate.go`、`pluginwindow.go`；§10 差集 57 文件。

### 批次 197（screenshot UIA 控件矩形规范化/可用性 2 函数 [S]）

**基线**：`FUNCS=2762 / S=1215 / S-inline=36 / S-sig=1397 / P=114 / UNMARKED=0`（真函数 2648）。
**收口**：`FUNCS=2762 / S=1217 / S-inline=36 / S-sig=1395 / P=114 / UNMARKED=0`（真函数 2648 = 55.7%），
SHA256 `04005C3C5544C7E298CC4071DB1B129C3CA9DD080229E5E3FCF4112D20A4C525`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,279,744 B，`bash build.sh` 重建）。

**本批落地**（screenshot_uia_windows.go，详见 acceptance/batch197.md）：
- `normalizeScreenshotControlRect`（0x1409af400，224B）[S-sig]→[S]：rect/bounds 空 → (空,false)；
  否则 rect.Intersect(bounds)，交集 Dx()<2 || Dy()<2 → (空,false)；否则 (交集,true)。
- `screenshotUIAControlRectUsable`（0x1409af4e0，288B）[S-sig]→[S]：normalize 失败 → false；
  controlType 不在 [50008,50033] → true；容器类 role（List/Menu/MenuBar/StatusBar/Tab/ToolBar/
  Tree/Group/DataGrid/Document/Window/Pane）→ 返回 !screenshotRectMatchesBounds(交集,bounds)；
  其余 → true。switch 跳转表 0x1411e1ea0 26 项经 va_read 原始字节逐项解码（CASE_MATCH=0x1409af545
  / RET_TRUE=0x1409af56c），未臆造分派。

**关键知悉**：纯升档（S-sig→S），真函数总数与 P 不变；UIA 容器类 role 判定依据标准
ControlTypeId 常量（50008=List … 50033=Pane）。

**下一批**：screenshot_uia_windows.go 剩余 [S-sig]/[P] 多为 COM/IUIAutomation vtable 调用
（createScreenshotUIAutomationOnCOMThread、readScreenshotUIAElementHitInfo、
screenshotUIASelectableAncestorHitInfoAtPoint 等），需先还原 COM 包装后成批落地；`matchNodeNameTerms`
（0x1407e3ac0）、`pluginupdate.go`、`pluginwindow.go`、`TestReminderNotification` 0x1407a7320、
`desktopCalendarStringList` 0x1407aa820；§10 差集 57 文件。

### 批次 198（qrexternal 二维码外部 URL 域 4 函数 [S] + 独立成文件）

**基线**：`FUNCS=2762 / S=1217 / S-inline=36 / S-sig=1395 / P=114 / UNMARKED=0`（真函数 2648）。
**收口**：`FUNCS=2762 / S=1221 / S-inline=36 / S-sig=1391 / P=114 / UNMARKED=0`（真函数 2648 = 55.7%），
SHA256 `18C1130FB400012F8C2FE93B8B7B4AD19A4680669E1F534E5589E0D291BA9538`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,279,744 B，`bash build.sh` 重建）。

**本批落地**（`backend/qrexternal.go` 新建，从 qrcode_windows.go 拆出 4 存根升 [S]，详见 acceptance/batch198.md）：
- `OpenQRCodeExternalURL`（0x14095eda0，128B）[S-sig]→[S]：normalize 后 err 非 nil 原样返回，
  否则 `openWithDefaultHandler(normalized)`（复用 bootstrap_pathclip.go）。
- `normalizeQRCodeExternalURL`（0x14095ee20，1792B）[S-sig]→[S]：9 类中文错误消息逐条 va_read 实证；
  TrimSpace 空/首尾空白 → 错误1；不安全文本 → 错误2；双重 url.PathUnescape（mode=encodePathSegment，
  对应 asm net/url.unescape mode=2）至多 2 次 → 错误3/4；url.Parse 失败或 Opaque 非空 → 错误9；
  Scheme=lower(trim) 非 http/https → 错误5；User 非 nil → 错误8；host=trim(Hostname()) 空 → 错误7；
  net.ParseIP 命中 → ip.String()，否则 idna.Lookup.ToASCII（err 或 trim 后空 → 错误6）；
  port=Port()；host 含 ":"（IPv6）→ "[host]" 或 "[host]:port"，否则 ToLower(host) + 可选 ":port"；
  返回 u.String()。
- `containsUnsafeQRCodeURLText`（0x14095f520，224B）[S-sig]→[S]：strings.Contains(s,"\\") 命中 → true；
  否则按 rune 判 r<=0x1f || (r>=0x7f && r<=0x9f)；256B 位图 @0x14196c560 65 个 &1 命中解码，
  rune>0xff 视为安全。
- `qrExternalURLInvalidError`（0x14095f600，192B）[S-sig]→[S]：TrimSpace 空 → 「二维码链接无效」；
  errors.New("QR_EXTERNAL_URL_INVALID: "+s)（前缀 25B ASCII @0x140c66cfe）。

**关键知悉**：idna Profile 全局变量指针 `mov [rip+0x1264444]` 精算为 0x141bc3560（此前
0x140bb9560 是把 disp32 加错位）→ 0x141be4440；其 options 布尔位 `00 01 01 01 00 00`
（transitional=false、useSTD3Rules/checkHyphens/checkJoiners=true）与 mapping=validateAndMap
（0x14071dda0）、fromPuny=validateFromPunycode（0x14071ea60）、bidirule=ValidString（0x14071cde0）
精确对应 idna.Lookup，非 Punycode/Display/Registration。纯升档（S-sig→S），真函数总数与 P 不变。

**下一批**：qrcode.go 域仍有 2 个 [S-sig] 存根（readQRCodeTextFromClipboard/writeQRCodeTextToClipboard，
Win32 剪贴板链）；screenshot_uia_windows.go 剩余 COM/IUIAutomation vtable 包装；`matchNodeNameTerms`
（0x1407e3ac0）、`pluginupdate.go`、`pluginwindow.go`、`TestReminderNotification` 0x1407a7320、
`desktopCalendarStringList` 0x1407aa820；§10 差集 56 文件。

### 批次 199（desktopwidgets_calendar.go desktopCalendarStringList [S]）

**基线**：`FUNCS=2762 / S=1221 / S-inline=36 / S-sig=1391 / P=114 / UNMARKED=0`（真函数 2648）。
**收口**：`FUNCS=2763 / S=1222 / S-inline=36 / S-sig=1391 / P=114 / UNMARKED=0`（真函数 2649 = 55.7%），
SHA256 `A3DD9DB9448D27603281E6A8FB1840464E617FFCD82CBDC6832BBDF9337D3DD9`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,312,000 B，`bash build.sh` 重建）。

**本批落地**（`backend/desktopwidgets_calendar.go` 新建，1 函数 [S]，详见 acceptance/batch199.md）：
- `desktopCalendarStringList`（0x1407aa820，320B）[S]：`func desktopCalendarStringList(l *list.List) []string`；
  l==nil 或 l.Len()==0 → nil；make([]string,0,l.Len())；遍历 l.Front()..Next()，e.Value.(string) 断言
  成功且非空 → append；返回切片。

**关键知悉**：参数 `*list.List`（48B = root.Element 40B + len int@+0x28）、`Element`（40B =
next/prev/list@+0x00/+0x08/+0x10 + Value any@+0x18 type/+0x20 data）与 container/list 精确吻合；
`Solar/Lunar.GetFestivals()`（lunar-go v1.4.6 源码实证）返回 `*list.List`，本函数是它的转 []string 收口。
string 类型描述符经 lea 目标重算为 0x140ae9e00（Size=16/Kind=24），早前手算 0x140c29e00 系进位错误已纠正。
纯新增 [S]，真函数 +1，P 不变。

**下一批**：desktopwidgets_calendar.go 还剩 buildDesktopCalendarMonth 主函数（0x1407a9f60）未落地（仍在差集）；
其签名经调用点 0x1407a9ac0 实证为 `(year int, month int, now time.Time)` 返回 `DesktopCalendarMonth`（56B=7 寄存器），
体含 lunar-go 农历/节气/节日逐日计算，独立成批。qrcode.go 域 2 个 [S-sig] 剪贴板存根、`matchNodeNameTerms`
（0x1407e3ac0）、`pluginupdate.go`、`pluginwindow.go`、`twofactor.go`、`transport.go`；§10 差集 56 文件。

### 批次 200（buildDesktopCalendarMonth [S] 完整还原 + GetDesktopCalendarMonth 签名修正）

**基线**：`FUNCS=2763 / S=1222 / S-inline=36 / S-sig=1391 / P=114 / UNMARKED=0`（真函数 2649）。
**收口**：`FUNCS=2763 / S=1223 / S-inline=36 / S-sig=1390 / P=114 / UNMARKED=0`（真函数 2649 = 55.7%），
SHA256 `42CDAC7DB025558CC0D0717ABC9D8F1883112263150BA994BB3A017E9C05283D`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,322,240 B，`bash build.sh` 重建）。

**本批落地**（desktopwidgets_calendar.go 域完整闭环，差集 -1，详见 acceptance/batch200.md）：
- `buildDesktopCalendarMonth`（0x1407a9f60，2240B）[S-sig]→[S]：
  `func buildDesktopCalendarMonth(year int, month int, now time.Time) (DesktopCalendarMonth, error)`；
  year∈[1901,2100]/month∈[1,12] 否则 `errors.New("desktopWidgets.calendarRange")`；now.IsZero()→time.Now()；
  dayCount=time.Date(year,month+1,0,12,...).Day()；serverNow=now.Format(RFC3339Nano)；make([]DesktopCalendarDay,0,dayCount)；
  逐日 NewSolar→NewLunarFromSolar，Date=fmt.Sprintf("%04d-%02d-%02d")、Weekday=time.Date(...,12,...).Weekday()、
  LunarYear=GetYearInGanZhi()、LunarMonth=GetMonthInChinese()（month<0→"闰"+MONTH[-month]）、LunarDay=GetDayInChinese()、
  SolarTerm=GetJieQi()、双节日列表=desktopCalendarStringList(GetFestivals())；返回 (DesktopCalendarMonth,nil)。
- `GetDesktopCalendarMonth`（0x1407a9ac0）签名修正：`(year int, month int) DesktopCalendarMonth`
  （原误标 `() interface{}`），asm 实证接收 year/month、time.Now()、返回 7 寄存器结构值。
- 依赖新增 `github.com/6tail/lunar-go v1.4.6`（buildinfo 实证 h1:APCXi1PC3Q7gZt6RJyug/ZdZcwX2qOkzIsZIcjCQdHY=）。

**关键知悉**：Weekday 的 magic 除法（0xc22e450672894ab7 除 86400 + 0x2492492492492493 除 7）经 Python
逐位模拟（含 rcr 进位）证明等价 `(absSec/86400+3)%7`，即 `time.Weekday()`；Lunar 结构字段偏移
（yearGanIndex@+0x30/yearZhiIndex@+0x38/month@+0x08/day@+0x10）与 lunar-go v1.4.6 Lunar.go 逐字段对齐。
行为活体实证：2024-02 = 29 天，02-01 癸卯年腊月廿二周四、02-10 正月初一春节、02-24 正月十五元宵节、
越界返回 calendarRange error（临时 go test 已删）。

**下一批**：desktopwidgets_clock.go 还剩 buildDesktopWorldClockSnapshot 主函数（0x1407aa960，1984B）[S-sig]
未落地（仍在差集），调用点 GetDesktopWorldClockSnapshot（0x1407a9be0）同模式需签名修正。qrcode.go 域 2 个
[S-sig] 剪贴板存根、`matchNodeNameTerms`（0x1407e3ac0）、`pluginupdate.go`、`pluginwindow.go`、`twofactor.go`、
`transport.go`；§10 差集 55 文件。

### 批次 201（buildDesktopWorldClockSnapshot [S] 完整还原 + GetDesktopWorldClockSnapshot 签名修正）

**基线**：`FUNCS=2763 / S=1223 / S-inline=36 / S-sig=1390 / P=114 / UNMARKED=0`（真函数 2649）。
**收口**：`FUNCS=2763 / S=1224 / S-inline=36 / S-sig=1389 / P=114 / UNMARKED=0`（真函数 2649 = 55.7%），
SHA256 `ABD47A9F955BB60D6C60C68B67C9750869159B68B50B885268AD89D73A46E5BE`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,322,240 B，`bash build.sh` 重建）。

**本批落地**（desktopwidgets_clock.go 域完整闭环，差集 -1，详见 acceptance/batch201.md）：
- `buildDesktopWorldClockSnapshot`（0x1407aa960，1984B）[S-sig]→[S]：
  `func buildDesktopWorldClockSnapshot(timezones []DesktopWorldClockZone, now time.Time) (DesktopWorldClockSnapshot, error)`；
  len>24 → `errors.New("desktopWidgets.tooManyTimezones")`；now.IsZero()→time.Now()；serverNow=now.Format(RFC3339Nano)；
  make([]DesktopWorldClockTime,0,len) + make(map[string]struct{},len) 去重；逐时区 id=TrimSpace(TimezoneID)，
  空/已见跳过，LoadLocation 失败 → `fmt.Errorf("desktopWidgets.invalidTimezone: %s", id)`；localNow=now.In(loc)；
  _,offset=Zone()；LocalTime=Format(RFC3339)；Offset=desktopTimezoneOffset(offset)；OffsetSeconds=offset；DST=IsDST()；
  Label=TrimSpace(zone.Label)；返回 (DesktopWorldClockSnapshot,nil)。
- `GetDesktopWorldClockSnapshot`（0x1407a9be0）签名修正：`(timezones []DesktopWorldClockZone) DesktopWorldClockSnapshot`
  （原误标 `() interface{}`），asm 实证接收 timezones slice、time.Now()、返回 5 寄存器结构值。

**关键知悉**：timezones 参数为 `[]DesktopWorldClockZone`（32B 元素 = TimezoneID/Label 两 string），非 []string；
LoadLocation 失败格式串 `desktopWidgets.invalidTimezone: %s`（34B @0x140c75595）与 tooManyTimezones（31B @0x140c70d7c）
均经 lea 目标重算（Python 精确加法）实证。行为活体实证：Asia/Shanghai → UTC+08:00/28800/DST=false/RFC3339 正确；
重复时区去重 + 空时区跳过；无效时区与 25 时区均返回对应 error（临时 go test 已删）。

**下一批**：desktopwidgets 系列（calendar/clock）已闭环，§10 差集 54 文件。候选：`matchNodeNameTerms`
（0x1407e3ac0）、qrcode.go 2 个 [S-sig] 剪贴板存根、`pluginupdate.go`、`pluginwindow.go`、`twofactor.go`、
`transport.go`、`TestReminderNotification`（0x1407a7320）。

### 批次 202（filesearch 名称搜索计划域完整还原：6 函数 [S] + 3 结构）

**基线**：`FUNCS=2763 / S=1224 / S-inline=36 / S-sig=1389 / P=114 / UNMARKED=0`（真函数 2649）。
**收口**：`FUNCS=2769 / S=1230 / S-inline=36 / S-sig=1389 / P=114 / UNMARKED=0`（真函数 2655 = 55.8%），
SHA256 `27D52FEAFB342F72C01E180EA53D49D29E1F7BFC587B224CAAD3EBABEFD587AF`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,322,240 B，`bash build.sh` 重建）。

**本批落地**（`backend/filesearch_index_windows.go` 名称搜索计划域闭环，详见 acceptance/batch202.md）：
- 3 结构：`nameSearchTerm`（56B，`_ [16]byte` + `patterns [][]byte` + `flags [16]byte`）、
  `nameFrequencyIndex`（0x410，`unigram [256]int32` + `bigram *[65536]int32` + `bigramReady uint64`）、
  `nameSearchPlan`（56B，`mask/matched/others uint32` + `driver int` + `driverBit uint32` + `order []int`）。
- 6 函数 [S]：`shouldPrioritizeSearchTerm`（0x1407e19c0，五级比较）、`estimateSearchTermCandidateCount`
  （0x1407e3100，bigram/unigram 查频表）、`selectDriverTermIndex`（0x1407e3240）、
  `buildOrderedNameTermIndices`（0x1407e33a0，sort.SliceStable）、`buildIndexSearchPlan`（0x1407e36e0）、
  `matchNodeNameTerms`（0x1407e3ac0，三段：driver 快路径→order 顺序→全部术语兜底）。

**关键知悉**：术语元素步长 0x38（56B）经六处 asm 实证；`uint32(1)<<uint(i)` 处理 idx≥32 自动归零，
与 asm `shl + sbb + and` 一致；mask=`uint32(1)<<uint(n)-1` 在 n≥32 下溢为全 1，匹配 asm `lea r12d,[r13-1]`；
bigram 键 `c0<<8|c1`（16 位截断）与 asm `movzx edx,dx` 一致。术语 +0x00/+0x0f 保留字段语义本域不访问
（上游 buildNamePrefixBucketCounts 构建），未臆造。

**下一批**：§10 差集 54 文件。desktopwidgets（calendar/clock）、filesearch 索引计划域均已闭环。
候选：qrcode.go 2 个 [S-sig] 剪贴板存根、`matchNodeNameTermsWithPinyin`（0x14080fe40，
filesearch_pinyin_windows.go 域）、`pluginupdate.go`、`pluginwindow.go`、`twofactor.go`、
`transport.go`、`TestReminderNotification`（0x1407a7320，依赖 launcherWidgetStore.Read +
normalize/validateDesktopReminderAudio 链）。

### 批次 203（qrcode 剪贴板读写链 [S-sig]→[S]）

**基线**：`FUNCS=2769 / S=1230 / S-inline=36 / S-sig=1389 / P=114 / UNMARKED=0`（真函数 2655）。
**收口**：`FUNCS=2769 / S=1232 / S-inline=36 / S-sig=1387 / P=114 / UNMARKED=0`（真函数 2655 = 55.8%），
SHA256 `C664417959FF917CBCFAF2CC72577EC561840F1A4C9CEF7DEA296310C9FA952E`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B，`bash build.sh` 重建）。

**本批落地**（`backend/qrcode_windows.go` "qrcode.go domain" 段 2 存根升档，详见 acceptance/batch203.md）：
- `readQRCodeTextFromClipboard`（0x140931960，96B）[S-sig]→[S]：`initQRCodeClipboard` 失败 → ("",err)；
  否则 `string(clipboard.Read(clipboard.FmtText))`，nil error。
- `writeQRCodeTextToClipboard`（0x1409319c0，128B）[S-sig]→[S]：`initQRCodeClipboard` 失败 → 透传 err；
  否则 `clipboard.Write(clipboard.FmtText, []byte(text))`（丢弃 channel），return nil。

**关键知悉**：FmtText=iota 0（asm eax=0 与 clipboard.go `FmtText Format = iota` 一致）；
golang.design/x/clipboard v0.8.0 API `func Read(t Format) []byte` / `func Write(t Format, buf []byte) <-chan struct{}`
与 asm 完全对齐（Read 3 word 返回、Write 1 word 返回丢弃）；stringtoslicebyte/slicebytetostring 为
`[]byte(text)` / `string(data)` 的编译器零拷贝路径。

**下一批**：§10 差集 54 文件。desktopwidgets（calendar/clock）、filesearch 索引计划域、qrcode 剪贴板
读写链均已闭环。候选：`matchNodeNameTermsWithPinyin`（0x14080fe40，filesearch_pinyin_windows.go 域）、
`pluginupdate.go`、`pluginwindow.go`、`twofactor.go`、`transport.go`、
`TestReminderNotification`（0x1407a7320，依赖 launcherWidgetStore.Read +
normalize/validateDesktopReminderAudio 链）。

### 批次 204（twofactor 供给 URI 图标编解码 [S-sig]→[S]）

**基线**：`FUNCS=2769 / S=1232 / S-inline=36 / S-sig=1387 / P=114 / UNMARKED=0`（真函数 2655）。
**收口**：`FUNCS=2769 / S=1234 / S-inline=36 / S-sig=1385 / P=114 / UNMARKED=0`（真函数 2655 = 55.8%），
SHA256 `79D49F3A62BEFA1721AD14F65700AFE33DD69E589592AEC554BB191341F70697`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B，`bash build.sh` 重建）。

**本批落地**（`backend/twofactor_provision_uri.go` 2 存根升档，详见 acceptance/batch204.md）：
- `encodeTwoFactorProvisioningIcon`（0x1409d9100，416B）[S-sig]→[S]：sanitize("totp",iconData) 空 → ""；
  HasPrefix(lower,"base64:") → "base64:"+TrimSpace(s[7:])；HasPrefix(lower,"data:image/") → Cut(s,",")
  且 found 且 Contains(ToLower(before),";base64") → "base64:"+TrimSpace(after)；其余 ""。
- `decodeTwoFactorProvisioningIcon`（0x1409d92a0，192B）[S-sig]→[S]：TrimSpace 空 → ""；否则
  sanitize("totp", s)（asm 两分支同目标，合并单一路径）。

**关键知悉**：五处字符串字面量 va_read 逐字节实证：`"totp"` 4B @0x140c34986、`"base64:"` 7B
@0x140c39568、`"data:image/"` 11B @0x140c47539、`","` 1B @0x1411cac20、`";base64"` 7B
@0x140c39457。encode/decode 均依赖 `sanitizeTwoFactorEntryIconData`（已 [S]，twofactor_normalize.go）。

**下一批**：§10 差集 54 文件。desktopwidgets（calendar/clock）、filesearch 索引计划域、qrcode 剪贴板
读写链、twofactor 供给 URI 图标编解码均已闭环。候选：`buildTwoFactorLabel`（0x1409d9360）、
`matchNodeNameTermsWithPinyin`（0x14080fe40，filesearch_pinyin_windows.go 域）、`pluginupdate.go`、
`pluginwindow.go`、`twofactor_provision_misc.go` 各存根、`transport.go`、
`TestReminderNotification`（0x1407a7320）。

### 批次 205（twofactor 供给 URI 标签构造 [S-sig]→[S]）

**基线**：`FUNCS=2769 / S=1234 / S-inline=36 / S-sig=1385 / P=114 / UNMARKED=0`（真函数 2655）。
**收口**：`FUNCS=2769 / S=1235 / S-inline=36 / S-sig=1384 / P=114 / UNMARKED=0`（真函数 2655 = 55.8%），
SHA256 `8279C159F1E606A63D88AE86C7BAD3C4699BBD7D5E94F52375EFAA79C30CAA2F`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B，`bash build.sh` 重建）。

**本批落地**（`backend/twofactor_provision_uri.go`，详见 acceptance/batch205.md）：
- `buildTwoFactorLabel`（0x1409d9360，288B）[S-sig]→[S]：issuer/accountName/name 三 TrimSpace；
  steam → issuer="Steam"（大写）；双非空 → "issuer:account"；accountName 非空 → accountName；
  issuer 非空 → issuer；否则 name。

**关键知悉（重要修正）**：`"Steam"` 5B @0x140c35d94（raw_hex `537465616d`，大写 S）与 sanitize
域小写 `"steam"` @0x140c35d8f 是两个不同字面量；原注释「固定为 'steam'」有误，已修正。分隔符
`":"` 1B @0x1411cac58。

**下一批**：§10 差集 54 文件。desktopwidgets（calendar/clock）、filesearch 索引计划域、qrcode 剪贴板
读写链、twofactor 供给 URI（图标编解码 + 标签构造）均已闭环。候选：
`matchNodeNameTermsWithPinyin`（0x14080fe40，filesearch_pinyin_windows.go 域）、`pluginupdate.go`、
`pluginwindow.go`、`twofactor_provision_misc.go` 各存根、`transport.go`、
`TestReminderNotification`（0x1407a7320）。

### 批次 206（filesearch 排序比较器域闭环：compareText/compareCandidate/Resolve/Release）

**基线**：`FUNCS=2769 / S=1235 / S-inline=36 / S-sig=1384 / P=114 / UNMARKED=0`（真函数 2655）。
**收口**：`FUNCS=2773 / S=1239 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`（真函数 2660 = 55.9%），
SHA256 `AF2600E9B8EC4126EEAEF39EF711A6766F8F6991E97AF661764631021108E588`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B，`bash build.sh` 重建）。

**本批落地**（`backend/filesearch_windows.go` 排序比较器子域闭环，详见 acceptance/batch206.md）：
- 4 函数升档 [S]：`fileSearchCandidatePathResolver.Resolve`（0x14081ff00，签名 void→`string`）、
  `fileSearchCandidatePathResolver.Release`（0x1408200a0，遍历 pathCaches + returnNodePathCache + delete）、
  `fileSearchSortComparer.compareText`（0x140820180，TrimSpace + 空值处理 + collator/cmpstring + desc 反转）、
  `fileSearchSortComparer.compareCandidate`（0x1408202c0，**[P]→[S]**，签名 `bool`+2 参 → `int`+3 参加 resolver；
  spec.Key 三态分派 path/modifiedAt/默认的字段优先级链）。
- 4 依赖存根新增 [S-sig]：`newNodePathCache`（0x1407e13e0）、`VolumeIndex.NodeCount`（0x140809280）、
  `VolumeIndex.ResolvePathByNodeIndexAndFRN`（0x140808f20）、`returnNodePathCache`（0x1407e1560）。

**关键知悉**：`scoredFileSearchCandidate` 字段偏移 asm 实证（Name@0x20、FRN@0x30、ModifiedAtUnix@0x40、
Path@0x50/0x58、pathResolved@0x60）；`fileSearchSortComparer` spec.Key@0x00/spec.Direction@0x10/collator@0x20；
`fileSearchCandidatePathResolver` pathCaches@0x00/pathResolveDuration@0x08。Key 字面量内联小端常量
（"path"=0x68746170、"desc"=0x63736564、"modifiedAt"=0x6465696669646f6d+0x7441）。compareCandidate/compareText
均返回 int 三态（-1/0/1），原 [P] 存根误标 bool 已订正。

**下一批**：§10 差集 54 文件。filesearch 排序比较器子域已闭环，路径解析链（newNodePathCache/
ResolvePathByNodeIndexAndFRN/returnNodePathCache/NodeCount）仍 [S-sig] 待专项。候选：
`matchNodeNameTermsWithPinyin`（0x14080fe40）、`pluginupdate.go`、`pluginwindow.go`、
`twofactor_provision_misc.go` 各存根、`transport.go`、`TestReminderNotification`（0x1407a7320）。

### 批次 207（filesearch 拼音模糊匹配域：Match/matchFileSearchPinyinTerm/matchNodeNameTermsWithPinyin + 结构修正）

**基线**：`FUNCS=2773 / S=1239 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`（真函数 2660）。
**收口**：`FUNCS=2776 / S=1242 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`（真函数 2663 = 56.02%），
SHA256 `ADBE298A702394145C08F4C3BF31B5DF2AA5196B291C0508D71E26D24DECCCD2`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B，`bash build.sh` 重建）。

**本批落地**（`backend/filesearch_pinyin_windows.go` 新建，详见 acceptance/batch207.md）：
- `fileSearchPinyinFuzzyMatcher.Match`（0x14080a7e0，160B）[S]：nil receiver / nil contains → false；
  否则 `contains.Match(b)`（asm 实证转 regexp.Regexp.doExecute）。
- `matchFileSearchPinyinTerm`（0x14080fda0，160B）[S]：`bytesMatchWildcardFold(b, term.patterns, false)`
  命中 → true；否则 `term.matcher.Match(b)`。
- `matchNodeNameTermsWithPinyin`（0x14080fe40，1280B+func1 256B）[S]：与 matchNodeNameTerms 同构
  三段式，匹配函数换成闭包（wildcard + matchFileSearchPinyinTerm×2）；`plan nameSearchPlan` 按值传入。
- `nameSearchTerm` 结构修正：`flags [16]byte` 拆为 `flags [8]byte`（+0x28..0x30）+
  `matcher *fileSearchPinyinFuzzyMatcher`（+0x30..0x38），总大小仍 56B（0x38）。

**关键知悉**：`matchFileSearchPinyinTerm` 的 Match receiver 取自 term+0x30（8 字节），Match 内部
解引用 `[receiver]` 作 contains（结构 +0x00）——证明 +0x30 是 matcher 指针而非 flags 后半；flags
仅前 4 字节（+0x28..0x2b）被访问，+0x2c..0x2f 保留。`matchNodeNameTermsWithPinyin` 签名经调用方
matchSearchCandidateNodeWithPinyin 反推：9 寄存器=b/b1/b2，栈=terms+nameSearchPlan 按值（56B）+
caseSensitive，plan 字段偏移（matched@+0x04/driver@+0x10/driverBit@+0x18/order@+0x20）与主函数
读位逐一吻合。caseSensitive 仅闭包首段 wildcard 生效，matchFileSearchPinyinTerm 固定 false。

**下一批**：§10 差集 54 文件。filesearch 拼音模糊匹配域（Match/matchFileSearchPinyinTerm/
matchNodeNameTermsWithPinyin）已闭环。候选：`classifyVolumePinyinMatch`（0x140810960）、
`pluginupdate.go`、`pluginwindow.go`、`twofactor_provision_misc.go` 各存根、`transport.go`、
`TestReminderNotification`（0x1407a7320）。

### 批次 208（filesearch 拼音分类域：MatchExact/MatchPrefix/classifyVolumePinyinMatch）

**基线**：`FUNCS=2776 / S=1242 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`（真函数 2663）。
**收口**：`FUNCS=2779 / S=1245 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`（真函数 2666 = 56.08%），
SHA256 `09071D1036AC2D99EC7237D7ACB5893DCFB2629FDBC48ADE98C62A5F698181BF`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B，`bash build.sh` 重建）。

**本批落地**（`backend/filesearch_pinyin_windows.go` 扩展，详见 acceptance/batch208.md）：
- `fileSearchPinyinFuzzyMatcher.MatchExact`（0x14080a880，416B）[S]：exact 正则命中 → true；
  allowDot → false；否则去掉最后一个 '.' 及之后扩展名重试。
- `fileSearchPinyinFuzzyMatcher.MatchPrefix`（0x14080aa20，160B）[S]：prefix 正则匹配。
- `classifyVolumePinyinMatch`（0x140810960，480B+func1 800B）[S]：三级分类（精确 3/6、前缀 4/7、
  其他 5/8），先试 b1 再试 b2。

**关键知悉**：classifyVolumePinyinMatch 的 r9 参数反推确认为 **`*nameSearchTerm`**——`[r9+0x10]` 即
patterns.ptr（`[r9+0x10]+0/8/0x10` 三重解引用得 patterns[0] 的 ptr/len/cap）、`[r9+0x18]` 即
patterns.len、`[r9+0x30]` 即 matcher，与 nameSearchTerm 字段偏移逐一吻合。组 A（唯一别名）仅在
`len(patterns)==1` 时取 patterns[0]，组 B 恒取 patterns[0]。精确级三段兜底（正则精确 / 与唯一 pattern
忽略大小写全等 / !allowDot 时 bytesStemEqualFold），前缀级两路（MatchPrefix / b 前缀==aliasB），
其余归 other。两处忽略大小写全等只折叠 b（aliasA/aliasB 已小写），内联实现。

**下一批**：§10 差集 54 文件。filesearch 拼音模糊匹配 + 分类域已闭环。候选：
`volumeIndexReadView.matchSearchCandidateNodeWithPinyin`（0x140810440，依赖 classify 与
matchNodeNameTermsWithPinyin 均已就绪）、`searchPinyinContextWithTombstones`（0x140810e60）、
`pluginupdate.go`、`pluginwindow.go`、`twofactor_provision_misc.go` 各存根、`transport.go`、
`TestReminderNotification`（0x1407a7320）。

### 批次 209（filesearch 拼音别名查询链：aliasesForNode/pinyinAliasesForNode/matchHierarchyTermWithPinyin）

**基线**：`FUNCS=2779 / S=1245 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`（真函数 2666）。
**收口**：`FUNCS=2782 / S=1248 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`（真函数 2669 = 56.14%），
SHA256 `7C8EAFC81B3FC2F0713F9F8B63FAC53CBB292FFA01492DE8FBDBB3A538FF2AF0`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B，`bash build.sh` 重建）。

**本批落地**（`backend/filesearch_pinyin_windows.go` 扩展，详见 acceptance/batch209.md）：
- `volumePinyinIndex.aliasesForNode`（0x14080df20，800B）[S]：records（每记录 32B）内对
  nodeIndex 无符号 lower_bound 二分，命中后按记录头 12B（nodeIndex u32 / fullOffset u32 /
  fullLen u16 / initLen u16）从 m.aliases 切出 pinyinFull / pinyinInitials。
- `volumeIndexReadView.pinyinAliasesForNode`（0x14080f900，608B）[S]：nodeIndex>=0x40000000 走
  deltaNodes 覆盖（取 delta.pinyinFull/pinyinInitials），否则转 v.pinyin.aliasesForNode。
- `volumeIndexReadView.matchHierarchyTermWithPinyin`（0x14080fb60，576B）[S]：先 wildcard 匹配，
  未命中取拼音别名，依次 matchFileSearchPinyinTerm(full/initials, term)。

**关键知悉**：volumePinyinIndex 字段偏移复验 records @+0x28/0x30/0x38、aliases @+0x40/0x48/0x50
（identity 40B=0x28 后紧跟两切片）；记录头小端 12B，别名区 full 在前 initials 在后连续存放；二分用
**无符号**比较（jae）。volumeIndexDeltaNode 的 pinyinFull @+0x40、pinyinInitials @+0x58 复验一致。
`matchHierarchyTermWithPinyin` 里 `test r9b` 检查的是 pinyinAliasesForNode 返回的 **ok 标志**（call
覆盖 r9 寄存器），非 caseSensitive——拼音兜底仅在别名存在时进行。

**下一批**：§10 差集 54 文件。拼音别名查询链已闭环，为 `matchSearchCandidateNodeWithPinyin`
（0x140810440）清除了拼音侧依赖；剩余依赖 IsTombstonedDescendant（0x1407f1a60，49 行）、
nodeAtIndex（0x1407f2240，24 行）、nodeName（0x1407f1e40，29 行）均属 filesearch_index_windows.go
蓝图未落地段。候选：`searchPinyinContextWithTombstones`（0x140810e60）、`pluginupdate.go`、
`pluginwindow.go`、`twofactor_provision_misc.go` 各存根、`transport.go`、
`TestReminderNotification`（0x1407a7320）。

### 批次 210（filesearch 节点访问与墓碑链：nodeAt/nodeAtIndex/nodeName/IsTombstonedDescendant）

**基线**：`FUNCS=2782 / S=1248 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`（真函数 2669）。
**收口**：`FUNCS=2786 / S=1252 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`（真函数 2673 = 56.23%），
SHA256 `5AB200AFBBCE88F146CEEE787D252E401E265C18F138AB9B1E4470B0345C0CB2`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B，`bash build.sh` 重建）。

**本批落地**（`backend/filesearch_index_windows.go` 扩展 + `types_filesearch.go` 新结构，
详见 acceptance/batch210.md）：
- `volumeIndexReadView.nodeAt`（0x1407f2060，480B）[S]：nodeBytes 非空时从 24B 小端序列化节点解析
  IndexNode，否则直接取 v.nodes[i]。
- `volumeIndexReadView.nodeAtIndex`（0x1407f2240，544B）[S]：nodeIndex>=0x40000000 走 deltaNodes
  覆盖；越界/负值零值；节点数 = nodeBytes.len/24 或 nodes.len。
- `volumeIndexReadView.nodeName`（0x1407f1e40，544B）[S]：NameOffset==0xFFFFFFFF 走 delta
  （deltaByFRN 命中或线性扫描，Flags&2 跳过）；否则 namePool 切片。
- `volumeIndexTombstoneLookup.IsTombstonedDescendant`（0x1407f1a60，992B）[S]：沿父链墓碑判定，
  迭代 + memo 记忆化，delta 节点递归 ParentIdx。
- `volumeIndexTombstoneLookup` 结构（0xe8B）：嵌入 volumeIndexReadView + memo map[int32]bool @+0xe0。

**关键知悉**：nodeAt 的 1638 行源码为虚高（bounds check 展开），实际 480B；nodeBytes 字段小端
FRN@+0/NameOffset@+8/ParentIdx@+0xc/ModTime@+0x10/NameLen@+0x14/Flags@+0x16。nodeAtIndex 的节点数
用 magic 除法 0xaaaaaaaaaaaaaaab（除以 24）。nodeName 的 delta 值需 >=0x40000000（减偏移得索引）。
IsTombstonedDescendant 的 path 栈 cap 8，回填阶段沿 path 写 memo。

**下一批**：§10 差集 54 文件。`matchSearchCandidateNodeWithPinyin`（0x140810440）的全部依赖现已就绪
（matchNodeNameTermsWithPinyin / classifyVolumePinyinMatch / IsTombstonedDescendant / nodeAtIndex /
nodeName 均已落地），可收口该主函数。候选：`searchPinyinContextWithTombstones`（0x140810e60）、
`pluginupdate.go`、`pluginwindow.go`、`twofactor_provision_misc.go` 各存根、`transport.go`、
`TestReminderNotification`（0x1407a7320）。

### 批次 211（候选节点拼音匹配主函数：matchSearchCandidateNodeWithPinyin）

**基线**：`FUNCS=2786 / S=1252 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`（真函数 2673）。
**收口**：`FUNCS=2787 / S=1253 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`（真函数 2674 = 56.25%），
SHA256 `482D45B50EE1B8E34417796FE9184522DB1C20C036671A72C6EF8EB46039F414`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

**本批落地**（`backend/filesearch_pinyin_windows.go` 扩展，详见 acceptance/batch211.md）：
- `volumeIndexReadView.matchSearchCandidateNodeWithPinyin`（0x140810440，1312B）[S]：四段式——
  matchNodeNameTermsWithPinyin 算普通术语 matched；driver 未命中→false；mask==matched→true、
  ^matched&others==0→false；否则从 node.ParentIdx 沿父链上溯，matchHierarchyTermWithPinyin 补齐
  others 剩余层级术语，墓碑后代淘汰，末了返回 mask==matched。

**关键知悉**：签名 14 参数，nodeIndex（AX）在实现中未使用，起始父索引取自 node.ParentIdx（DI）；
IndexNode 其余字段亦未使用。nameSearchPlan 字段口径 mask uint32@+0x00 / others uint32@+0x08 /
driver int@+0x10 / driverBit uint32@+0x18 复验一致。内层遍历 `for idx := range terms`（非 order），
bit 用 uint32(1)<<idx（idx>=32 由 x86 cmp/sbb/and 归零），命中 matched|=bit、required&=^bit。

**下一批**：§10 差集 54 文件。调用方 `searchPinyinContextWithTombstones`（0x140810e60，98 行）及其
闭包 `.func1`（0x140811c60，84 行）已抽取汇编到
`docs/goresym/pipeline/tmp/searchPinyinContext.asm.txt`（700 行），是下一步候选。候选：
`classifyVolumePinyinMatch`（已落地）、`pluginupdate.go`、`pluginwindow.go`、
`twofactor_provision_misc.go` 各存根、`transport.go`、`TestReminderNotification`（0x1407a7320）。

### 批次 212（候选节点名匹配非拼音 + 计划类型修正：matchSearchCandidateNode）

**基线**：`FUNCS=2787 / S=1253 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`（真函数 2674）。
**收口**：`FUNCS=2788 / S=1254 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`（真函数 2675 = 56.27%），
SHA256 `8EF85A98319BB00F7675897BA75DDBC2F7FC873CCED566FC361E1021B9A43D44`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

**本批落地**（`backend/filesearch_index_windows.go` 扩展，详见 acceptance/batch212.md）：
- `volumeIndexReadView.matchSearchCandidateNode`（0x1407ecca0，1096B）[S]：非拼音候选路径，
  结构同拼音版（matchNodeNameTerms 算 matched → driver 淘汰 → mask==matched → ^matched&others==0
  淘汰 → 沿父链补齐 others → 墓碑淘汰 → mask==matched），但无 nodeIndex、无 b1/b2，name 直接传。
- 类型签名修正（asm 实证，行为不变）：nameFrequencyIndex 改为 namePrefixBucketCounts 别名；
  estimateSearchTermCandidateCount 字段改 firstByte/twoByte；matchNodeNameTerms 签名从 7 参数
  收敛为 (b,terms,caseSensitive,plan)。

**关键知悉**：matchNodeNameTerms 调用约定 = b/terms/caseSensitive 走寄存器、plan（56B）整体在栈；
matchSearchCandidateNode 参数铺排 = node 按值展开占 AX/BX/CX/DI/SI/R8、name 占 R9/R10/R11，
terms/plan/tombstoneLookup/caseSensitive 依次入栈（与拼音版 nodeIndex 占 AX 的差异根因）。

**下一批**：§10 差集 54 文件。闭包 `searchPinyinContextWithTombstones.func1`（0x140811c60，84 行，
汇编已抽取 `docs/goresym/pipeline/tmp/searchPinyinContextFunc1.asm.txt`）依赖已全部就绪，可作下一步；
再下一步 `searchPinyinContextWithTombstones`（0x140810e60，98 行）。候选：`pluginupdate.go`、
`pluginwindow.go`、`twofactor_provision_misc.go` 各存根、`transport.go`、
`TestReminderNotification`（0x1407a7320）。

### 批次 213（filelocator 运行时两 [P] 升档 [S]：acquireFileLocatorContentScan + waitIfPaused）

**基线**：`FUNCS=2788 / S=1254 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`（真函数 2675）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1385 / P=111 / UNMARKED=0`（真函数 2677 = 56.31%），
SHA256 `00EFA6EA8B4E4BB78EC7509303EE08752BF9E873C2C0A2BC8371A10FF3281B93`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/filelocator_runtime.go` 两 [P] 升档 [S] + `filelocator_runtime_test.go` 新增，
详见 acceptance/batch213.md）：
- `(*fileLocatorService).acquireFileLocatorContentScan`（0x1407d23e0，512B）[S]：签名 bool→error。
  尾声双寄存器返回即 error 接口；s==nil 返回的全局 error 经 .data 0x141bc4520 → "context canceled"
  实证为 context.Canceled。体：s.lock.Lock() 懒初始化单缓冲信号量 s.contentScan，取 ch 后 Unlock；
  ctx==nil 阻塞发送返回 nil；否则 select { ch<-struct{}{}: nil; <-ctx.Done(): ctx.Err() }。
- `(*fileLocatorService).waitIfPaused`（0x1407d09a0，768B）[S]：签名
  (ctx)bool → (ctx, generation uint64, startedAt time.Time) error。汇编实证 7 寄存器参数、error 双寄存器返回。
  体：s.lock.Lock()+defer Unlock；循环条件 generation==gen && state.Running && state.Paused &&
  ctx.Err()==nil 时 pauseCond.Wait()+time.Since(startedAt)；退出后 ctx.Err()!=nil 或 generation 变 →
  context.Canceled，state.Running → nil，否则 context.Canceled。

**关键知悉**：fileLocatorService 字段偏移经汇编对齐验证与 types_filelocator.go 定义一致
（generation@+0x20、state.Running@+0x118、contentScan@+0x1e8，FileLocatorConfig 实测 0xf0=240B），
此前怀疑的结构体布局偏差不成立。acquireFileLocatorContentScan 单缓冲信号量语义：首次 acquire 占满，
二次 acquire 阻塞至 release 或 ctx 取消。waitIfPaused 的 time.Since(startedAt) 结果存 [rsp+0x158]
未见下游读取，按 `_ =` 保守处理。

**下一批**：§10 差集 54 文件。searchPinyinContextWithTombstones（0x140810e60，98 行）与其闭包
.func1（0x140811c60，84 行）依赖已就绪，但含排序/合并/去重三段（3398B 汇编）需专项完整翻译。
filelocator 域剩余 [P]：runSearch（0x1407d0260，返回值 error 存疑→实为 void，prepared 值传 + limit
参数待精确）、walkRoot、processFile、finishSearch（8 参数）、prepareFileLocatorSearch、
buildFileLocatorPathFilter 等。

### 批次 214（filelocator 运行时两 [P] 升档 [S-sig]：runSearch + finishSearch 签名实证）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1385 / P=111 / UNMARKED=0`（真函数 2677）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1387 / P=109 / UNMARKED=0`（真函数 2679 = 56.35%），
SHA256 `2F4D421B7D6C1BE0F6579D7290E4AAE84F4E094FF7E963FF04D47C7C66EE04F2`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/filelocator_runtime.go` 两 [P] 升档 [S-sig]，仅签名修正，体留待专项，
详见 acceptance/batch214.md）：
- `(*fileLocatorService).runSearch`（0x1407d0260）：签名 `(prepared 指针,ctx)error`
  → `(prepared fileLocatorPreparedSearch, ctx context.Context, generation uint64)`。
  recv+ctx(2)+generation 共 4 寄存器，prepared 值传（288B 栈 @rbp+0x10）；返回 void（非 error）。
  体主循环遍历 prepared.roots 对每根调 waitIfPaused(ctx,generation,time.Now()) 后 walkRoot。
- `(*fileLocatorService).finishSearch`（0x1407d2a00）：签名 `(cancelled bool)`
  → `(generation uint64, startedAt time.Time, lastError string, cancelled bool)`。
  recv+generation+startedAt(3)+lastError(2)+cancelled(bool) 共 8 寄存器；返回 void。
  参数 5/6 经 strings.TrimSpace 调用实证为 string（非初判 int）。

**关键知悉**：runSearch 尾声调 finishSearch(generation,startedAt,"",true) 走取消路径；finishSearch
参数 5/6 初判 int 经 asm 内 TrimSpace 反推为 string。waitIfPaused 的 startedAt 实参在
runSearch/processFile 内均为 time.Now() 结果（3 word），印证批次 213 口径。

**下一批**：§10 差集 54 文件。walkRoot（0x1407d0d20，prepared 值传 + 7 寄存器 + startedAt 栈参，
返回 error 非 void——WalkDir 尾调用返回 error）、processFile（0x1407d15a0，9 寄存器 + prepared 值 +
startedAt 栈参，含 fs.FileInfo 接口实参）的 2 个指针实参类型需沿 WalkDir 闭包 func1（0x1407d10c0）
捕获链确证后落地。searchPinyinContextWithTombstones（0x140810e60）含排序/合并/去重三段仍待专项。

### 批次 215（filelocator 运行时两 [P] 升档 [S-sig]：publishProgress + prepareFileLocatorSearch）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1387 / P=109 / UNMARKED=0`（真函数 2679）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1390 / P=106 / UNMARKED=0`（真函数 2682 = 56.41%），
SHA256 `94FBC3CC507E6F3541692DBF597377CB4FF3D11DA717FB93C41D197160FA3F66`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/filelocator_runtime.go` 三 [P] 升档 [S-sig]，签名修正，体留待专项，
详见 acceptance/batch215.md）：
- `(*fileLocatorService).publishProgress`（0x1407d2620）：签名 `()` → `(generation uint64, startedAt time.Time)`。
  recv+generation+startedAt(3) 共 5 寄存器，返回 void。体：generation/Running 校验 + 复制 state 快照 +
  time.Since(startedAt) 算 elapsed。
- `prepareFileLocatorSearch`（0x1407d2ea0）：返回 `*fileLocatorPreparedSearch` → `fileLocatorPreparedSearch`（值）。
  request 值经栈传递（morestack 无寄存器保存）；返回 prepared 值（返回值结构体 @[rsp+0x3f8]，
  request 字段 duffcopy 自 normalized request），非指针。
- `buildFileLocatorPathFilter`（0x1407d3400）：签名 `(cfg) *fileLocatorPathFilter` →
  `(cfg, remark, value, include, typeStr string) (*fileLocatorPathFilter, error)`。
  cfg 值栈传 + 4 string（8 寄存器）；返回 filter 指针 + error 双寄存器（非单指针）。remark 未使用
  （dead 参数，FileLocatorFilter.Remark 命名推断）。

**关键知悉**：prepareFileLocatorSearch 返回值结构体 @[rsp+0x3f8] 逐字段对齐
fileLocatorPreparedSearch 定义（request@0、roots@0xf0、fileNameMatcher@0x108、contentMatcher@0x110、
pathFilter@0x118），印证批次 213 的 240B FileLocatorConfig 口径。MatchContent（0x1407d4ee0）初勘
确认 recv+8 参数（非 content string），体内检查 mode 字符串（"boolean" 0x6c6f6f62 实证）并尾调
matchFileLocatorBooleanAcrossFile，完整签名待专项。

**下一批**：§10 差集 54 文件。filelocator 域剩余 [P]：MatchContent（recv+8 参数）、
matchFileLocatorContentStreamContext、fileLocatorStreamingLineMatch、
fileLocatorBooleanAcrossFile 族、readFileLocatorPreview 等。walkRoot/processFile 指针实参类型待闭包链确证。

### 批次 216（filelocator 三 [P] 升档 [S-sig]：collectPositiveMatchers + newTermMatcher + readTextFromHandle）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1390 / P=106 / UNMARKED=0`（真函数 2682）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1393 / P=103 / UNMARKED=0`（真函数 2685 = 56.48%），
SHA256 `968A4D5C8F58035F6EDFE08D410B8D6F52E236FEB4AA20ADE8B8F0B9141FAEC7`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/filelocator_runtime.go` 三 [P] 升档 [S-sig]，签名修正，体留待专项，
详见 acceptance/batch216.md）：
- `collectFileLocatorPositiveMatchers`（0x1407dcf60）：`(node, out)` → `(node, positive bool, out)`。
  node(接口)+positive(bool)+out(*[]fileLocatorTermMatcher) 共 4 寄存器，返回 void。NOT 节点 `xor ecx,1`
  翻转极性；项节点 positive 时 append 16B 接口元素到 out。
- `newFileLocatorTermMatcher`（0x1407dd1c0）：`(query, mode)` → `(query, mode, matchCase bool)`。
  query/mode(string)+matchCase(bool)。体 TrimSpace+ToLower(mode) 后 "like"→wildcardToFileLocatorRegex，
  其余→compileFileLocatorRegex(regex,matchCase)。
- `readFileLocatorTextFromHandle`（0x1407da360）：`(r, maxSize)` → `(r, matcher, maxSize)`。
  r(io.Reader)+matcher(*fileLocatorStringMatcher)+maxSize(int)；返回 (string,error)。matcher 存入上下文
  对象 +0x08 字段。

**关键知悉**：out 元素 16B=接口（`fileLocatorTermMatcher`）对齐 `positiveMatchers` 定义；
readFileLocatorTextFromHandle 上下文对象 [0x08]=matcher、[0x10]=maxSize+1。parseFileLocatorBooleanExpression
（0x1407dac60）初勘确认 query+2 额外 qword+1 bool 参数，parser 结构体扩展 3 字段待专项。

**下一批**：§10 差集 54 文件。filelocator 域剩余 [P]：parseFileLocatorBooleanExpression（parser 扩展）、
fileLocatorProximityRanges、MatchContent 族（9 寄存器）、matchFileLocatorContentStreamContext、
fileLocatorStreamingLineMatch、newFileLocatorTextDocument。walkRoot/processFile 指针实参类型待闭包链确证。

### 批次 217（filelocator 四 [P] 升档 [S-sig]：流式行匹配 + 文本文档 + 布尔解析 + 邻近区间）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1393 / P=103 / UNMARKED=0`（真函数 2685）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1397 / P=99 / UNMARKED=0`（真函数 2689 = 56.56%），
SHA256 `42AB6442E798F212CDBD0E7365F472A93F3BFC297DA48E11D0A7037E33D881F8`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/filelocator_runtime.go` 四 [P] 升档 [S-sig]，签名修正 + 2 结构体扩展，
详见 acceptance/batch217.md）：
- `fileLocatorStreamingLineMatch`（0x1407d78e0）：`(matcher, line) bool` → `(line, matcher) ([]FileLocatorTextRange, int)`。
- `newFileLocatorTextDocument`（0x1407d7be0）：返回指针 → 值；`fileLocatorTextDocument` 补 runes [][]rune（8 word=64B）。
- `parseFileLocatorBooleanExpression`（0x1407dac60）：`(query)` → `(query, booleanScope, matchCase bool)`；
  `fileLocatorBooleanParser` 补 booleanScope/matchCase 字段。
- `fileLocatorProximityRanges`（0x1407dc640）：`(left,right,distance) []` → `(node, distance, active) ([]FileLocatorTextRange, bool)`。

**关键知悉**：fileLocatorTextDocument 补全 runes 后 8 word=64B 对齐返回栈布局；
fileLocatorBooleanParser 补全后对齐 tokens@0/pos@0x18/booleanScope@0x20/matchCase@0x30；
fileLocatorProximityRanges 参数 3 active(bool) 与递归返回 bool 同构。matchFileLocatorContentStreamContext
（0x1407d5d80）初勘确认 ctx(2)+matcher+maxSize(int)+r 四参数（jle/nil 检查）。

**下一批**：§10 差集 54 文件。filelocator 域剩余 [P]：walkRoot、processFile、MatchContent 族（9 寄存器）、
matchFileLocatorContentStreamContext。P 已跌破 100。walkRoot/processFile 指针实参类型待闭包链确证。

### 批次 218（内容匹配链五 [P] 升档 [S-sig] + matchFileLocatorContentContext 签名纠错）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1397 / P=99 / UNMARKED=0`（真函数 2689）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1403 / P=93 / UNMARKED=0`（真函数 2695 = 56.69%），
SHA256 `4FA8D63618BA7C5F7A1E34E89EE778E58830D97191FE2C98E2F83434F08EEE8B`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/filelocator_runtime.go`，五 [P] 升档 + 一 [S-sig] 纠错，详见 acceptance/batch218.md）：
- `MatchContent`：`(content string)` → `(doc fileLocatorTextDocument)`（recv+doc 8 字值传递）。
- `matchFileLocatorContentContext`：签名纠错 `(ctx,matcher,text)` → `(ctx,path,maxSize,matcher)`（6 寄存器）。
- `matchFileLocatorLineBased`：`(doc, m)`（第 9 参 r11=matcher）。
- `matchFileLocatorBooleanPerLine` / `matchFileLocatorBooleanAcrossFile`：`(doc, expr)`（第 9 参 r11=matcher+0x58）。
- `buildFileLocatorLineMatch`：`(doc, index)`（doc.lines[index] 转 rune 后截取 before/after）。

**关键知悉**：内容匹配调用链 processFile→matchFileLocatorContentContext(ctx,path,maxSize,matcher)→
newFileLocatorTextDocument→MatchContent(doc)→按 mode 分发 BooleanAcrossFile/PerLine(doc,expr) 或
LineBased(doc,matcher)→buildFileLocatorLineMatch(doc,index)。fileLocatorTextDocument 8 word（64B）全链
寄存器直传，与批次 217 补全的 runes 字段自洽。

**下一批**：filelocator 域剩余 [P] 仅 walkRoot（0x1407d0d20）、processFile（0x1407d15a0）两个，
其 a/b 指针实参指向 runSearch 栈上的结果累加器（[0]=count、[0x48]/[0x50]=paths slice 头）与双层闭包。
之后转向 windowmanagement/screenshot/oledblackout 等 Windows 域（匿名上下文结构体展开，需逐结构落地）。

### 批次 219（qrcode 框选会话八 [P] 升档 [S-sig]）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1403 / P=93 / UNMARKED=0`（真函数 2695）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1411 / P=85 / UNMARKED=0`（真函数 2703 = 56.86%），
SHA256 `5415D6667011C7F2A45B5EFDE1478B6518B1974463F82E946E40DFA21F29D74A`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/qrcode_windows.go`，八 [P] 升档，详见 acceptance/batch219.md）：
- `resolveControlHoverForWindow`：`(hwnd,p)` → `(sess, hwnd, p)`。
- `resolveControlSelectionAtPoint`：`(p)` → `(sess, p)`。
- `resolveControlSelectionForWindowAtPoint`：`(hwnd,p)` → `(sess, hwnd, p, force bool)`。
- `resolveControlClickSelectionAtPoint`：`(hwnd,p)` → `(sess, p, controlRect)`。
- `prepareSelectionResult`：`(hint uintptr)` → `(sess)`，返回 void。
- `resolveClickLikeControlSelection`：`(hint uintptr)` → `(sess)`。
- `drawAnnotationOverlay`：`(hdc,rect)` → `(hdc, rect, bits unsafe.Pointer)`。
- `applyRoundedSelectionPreview`：`(img,rect)` → `(pBits unsafe.Pointer, selRect, previewRect)`。

**关键知悉**：`sess *screenshotWindowSelectionSession` 沿 `resolveControl* →
screenshotControlBoundsAtPointThroughOverlaySessionWithTimeout` 链透传，类型唯一确定。void 判定统一依据
「全部返回路径无返回寄存器设置」，不用 `xor eax` 判别（void 函数尾声既可能 xor eax 也可能裸 ret）。

**下一批**：qrcode 域剩 updateCornerRadiusDrag / confirmAnnotationEditing（返回单寄存器零值，void/bool
无法判别）与 applyAnnotationToolbarAction（action 结构体首字段 + rbx 语义未确证）。之后优先
screenshot_uia_windows 的 session 透传链（与 qrcode 同构，可复用「session 指针类型贯通」手法）。

### 批次 220（screenshot_uia 两 [P] 升档 [S-sig]）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1411 / P=85 / UNMARKED=0`（真函数 2703）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1413 / P=83 / UNMARKED=0`（真函数 2705 = 56.90%），
SHA256 `5056F736359D969BB1293F133B16286C1E6480390FCB1820C2CB9E1EB8AF551A`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/screenshot_uia_windows.go`，两 [P] 升档，详见 acceptance/batch220.md）：
- `shouldPreferWindowOverControlAtPoint`：`(x,y,hwnd)` → `(ctrlRect, winBounds image.Rectangle, x, y int, ctrlFound, winFound bool)`。
- `screenshotPreferUIAElementHitInfo`：`(a,b image.Rectangle)` → `(a image.Rectangle, aControlType uint32, aOK bool, b image.Rectangle, bControlType uint32, bOK bool)`。

**关键知悉**：readScreenshotUIAElementHitInfo 返回 (rect,uint32,bool,error) 的 8 寄存器布局，被
screenshotPreferUIAElementHitInfo 参数分组复用印证。screenshotUIADeepestElementHitInfoAtPoint 返回
(unsafe.Pointer,image.Rectangle,bool) 已确证（尾迹 6 寄存器），但参数含 depth/state/栈 hit 累加器
匿名类型，保持 [P]。initializeScreenshotCOMThreadMode 参数确证 bool，返回装箱指针具体类型未定，保持 [P]。

**下一批**：screenshot_uia 域剩 9 个 [P]（三函数同构中间参数 + VARIANT 返回 + 装箱指针 + 匿名累加器），
暂缓。优先转 oledblackout_windows（27 个，多为绘制类，参数分组较规则）或 screenshot_windows（15 个）。

### 批次 221（screenshot_windows 会话方法四 [P] 升档 [S-sig]）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1413 / P=83 / UNMARKED=0`（真函数 2705）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1417 / P=79 / UNMARKED=0`（真函数 2709 = 56.98%），
SHA256 `FEBC4D6E31A93AF90C050DB3161E858C118C31624EF6802A3F6533CFDBA77D55`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/screenshot_windows.go`，四 [P] 升档，详见 acceptance/batch221.md）：
- `invalidateWindowChange`：`()` → `(hwnd uintptr)`。
- `invalidateInfoRectChange`：`()` → `(hwnd uintptr, a, b image.Rectangle)`。
- `shouldReuseControlHoverForWindow`：`(hwnd)` → `(hwnd, x, y int, ctrlFound, winFound bool)`。
- `cachedControlHoverAllowedForPoint`：`(x,y)` → `(hwnd, x, y int, ctrlFound, winFound bool)`。

**关键知悉**：shouldReuseQRCodeControlHover 的 rect.Min 来自 recv[0x158]/[0x160]，rect.Max 来自本函数
x/y 实参，point 来自 recv[0x168]/[0x170]。screenshotCurrentControlSelectionPreference 返回 (bool,bool)
沿链透传为 ctrlFound/winFound。

**下一批**：screenshot_windows 域剩 resolveControlHoverForWindow（参数 1 rbx 透传
screenshotControlBoundsAtPointThroughOverlaySessionWithTimeout 的 r9，语义未定）、
resolveCaptureRectForWindowAtPoint 等。继续 screenshot_windows 或转 oledblackout_windows（27 个）。

### 批次 222（oledblackout 两 [P] 升档 [S-sig]）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1417 / P=79 / UNMARKED=0`（真函数 2709）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1419 / P=77 / UNMARKED=0`（真函数 2711 = 57.02%），
SHA256 `744749C5F9C6B9914F1AECCF35B83F5B3836D50B136793CE31FC48929A41C761`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/oledblackout_windows.go`，两 [P] 升档，详见 acceptance/batch222.md）：
- `oledBlackoutEnumTopLevelWindows`：原签名正确，确证单指针 + void 后升档。
- `oledBlackoutVisibleWindowRectForScreenPause`：`(hwnd, screen) oledBlackoutRect` →
  `(hwnd uintptr, cache map[uintptr]oledBlackoutRect) (oledBlackoutRect, bool)`。

**关键知悉**：oledBlackoutVisibleWindowRectForScreenPause 的 map 是 hwnd→rect 缓存（value 16 字节值
类型），rbx 透传 mapaccess2_fast64 的 hmap，返回 5 寄存器（4×int32 + bool）。oledBlackoutBrowserAudio
/MediaContinuityMatchesCandidates 实参含 slice+string+额外 qword，未唯一确定，保持 [P]。

**下一批**：oledblackout 域剩 25 个 [P]（约 17 个 ctx 匿名闭包 + GSMTC session 结构体 + 匿名上下文
字段）。优先攻坚 screenshotControlBoundsAtPointThroughOverlaySessionWithTimeout 完整签名，可一次贯通
screenshot_windows 的 resolveControlHoverForWindow / resolveCaptureRectForWindowAtPoint 与
screenshot_uia 三同构函数。

### 批次 223（screenshot_pin/launcherupdate 两 [P] 升档 [S-sig]）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1419 / P=77 / UNMARKED=0`（真函数 2711）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1421 / P=75 / UNMARKED=0`（真函数 2713 = 57.07%），
SHA256 `09BAB4DA2C1B96FB2EB816372C47B9C8366DD31C498454CE5AA7ECE5BA3D5ADD`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（详见 acceptance/batch223.md）：
- `screenshotPinNormalizeSnapshotBounds`：`(x,y,w,h)` → `(x,y,w,h,minW,minH int)`（返回 4 int 不变）。
- `finishLauncherUpdateTask`：`(taskID)` → `(taskID int64, done chan struct{})`。

**关键知悉**：chan 元素类型未定类 [P] 应优先查对应 receiver 结构体字段反向确证（本批
launcherUpdateDone chan struct{} 即此例）。probeLauncherUpdatePackageRange 第三参已收敛为接口
(itab[0x18] 间接调用)，非第二 string 或标量。

**下一批**：可快速突破口 = 同类「chan 类型未定」[P] 查结构体字段；screenshot_pin 的
storeSnapshotBounds 缺 r9b bool（语义 = storeSnapshotLocked 透传标志，需先攻 storeSnapshotLocked）；
beginLauncherUpdateTask 6 寄存器返回形态待解。

### 批次 224（oledblackout_hotkey 两 [P] 升档 [S-sig]）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1421 / P=75 / UNMARKED=0`（真函数 2713）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1423 / P=73 / UNMARKED=0`（真函数 2715 = 57.11%），
SHA256 `2827DB44A4277896C094BE49EB7AB62292344EB3624BBBF94069DFB02051B6C3`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/oledblackout_hotkey_windows.go`，两 [P] 升档，详见 acceptance/batch224.md）：
- `buildOLEDBlackoutHotkeyRegistrations`：原签名正确，序言逐寄存器对齐后升档。
- `failedOLEDBlackoutHotkeyUpdateResult`：原签名正确，序言逐寄存器对齐后升档。

**关键知悉**：两函数原签名已正确，阻断仅为「未逐寄存器实证」——对齐后参数分组唯一。
applyOLEDBlackoutHotkeyBindings 仍 [P]（首参 windowsOLEDBlackoutHotkeyManager 未落地，interface{}
占位 2 字与 asm 首参 1 字指针不匹配）。

**下一批**：captureScreenshotWithLauncherVisibility 参数已确证 = (service, delay, hideLauncher, cb func()→6 字)，
仅 cb 返回类型待逆推；screenshotCandidateWindowAtPointWith (x/y+slice+hwnd+2 回调)；
screenshot_uia 三函数返回 8 寄存器；qrcode 仅剩 applyAnnotationToolbarAction。

### 批次 233（screenshotAccessibleLocation 补 VARIANT 参数）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1432 / P=64 / UNMARKED=0`（真函数 2724）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1433 / P=63 / UNMARKED=0`（真函数 2725 = 57.30%），
SHA256 `75DF02BB96204D47C0A8FA4BFEC812B5381D6EE608532B0F3F9A72315F22486A`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/screenshot_uia_windows.go`，详见 acceptance/batch233.md）：
- `screenshotAccessibleLocation`：补第二形参 `screenshotOleVariant`（值传 5 寄存器字段展开）。

**关键知悉**：VARIANT 值传 = 4 个 uint16 各占独立寄存器低 16 位 + 1 个 int64 占第 5 寄存器，
非打包 2 word；识别关键在序言 `mov word ptr [rsp+x], bx/cx/di/si` 的 16 位保存。

**下一批**：filelocator walkRoot/processFile（值传大结构体 prepared + 额外 generation/ctx）；
screenshotAccessibleHitTest（3 参数但 9 字返回）；storeSnapshotBounds（r9b 死参数，bool 未读）。

### 批次 234（screenshot_pin bounds 签名链订正）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1433 / P=63 / UNMARKED=0`（真函数 2725）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1439 / P=57 / UNMARKED=0`（真函数 2731 = 57.45%），
SHA256 `E777DD37779D676AB74BFBAD3D469C9328DBEDBB1E325B192F3772246663ACC9`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/screenshot_pin.go` + `backend/screenshot_pin_native_windows.go`，详见 acceptance/batch234.md）：
- `storeSnapshotBounds`：`(w, bounds application.Rect, persist bool)`（确认正确）。
- `storeSnapshotLocked`：`(view screenshotPinnedWindowView, bounds application.Rect, persist bool)`（view 栈传大结构体）。
- `createPinnedWindow`：`(w, bounds application.Rect, persist bool) error`。
- `registerPinnedWindow`：`(w, bounds application.Rect, persist bool) bool`。
- `createWebviewPinnedWindow`：`(w, bounds application.Rect, persist bool) error`。
- `buildScreenshotPinWindowHTML`：`(name, imageURL string, w, h int, scale float64) string`。
- `createScreenshotNativePinWindow`：签名订正 `(service, pin, pinName string, pinGeneration uint64, image *image.RGBA)` → `(service, pin, bounds application.Rect)`。

**关键知悉**：createPinnedWindow 的 rcx/rdi/rsi/r8 是 `application.Rect{X,Y,Width,Height}`（透传
createScreenshotNativePinWindow 存对象 0x58/0x60 不钳制、0x68/0x70 钳 ≥1），非旧注释臆测的
(pinName,pinGeneration,image)；据此订正整条 bounds 链。storeSnapshotLocked 的 view 大结构体走
调用者栈传递（duffcopy 至 [rsp]，被调者 [rsp+0x268..0x310] 读字段），persist（bl）为死参数。

**下一批**：filelocator walkRoot/processFile（值传大结构体 prepared + 额外 generation/ctx）；
screenshot_uia_windows accessibleHitTest（3 参数但 9 字返回）；qrcode applyAnnotationToolbarAction。

### 批次 235（qrcode 标注动作链 hwnd 订正）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1439 / P=57 / UNMARKED=0`（真函数 2731）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1440 / P=56 / UNMARKED=0`（真函数 2732 = 57.46%），
SHA256 `0706E3CA489F83C4DB140D3C59D9C58BB7DC3EBB68065076D743FD5780B03115`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/qrcode_windows.go`，详见 acceptance/batch235.md）：
- `applyAnnotationToolbarAction`：`[P]` → `[S-sig]`，补 `hwnd uintptr` 首参（action 值传走栈）。
- `handleAnnotationToolbarActions`：补 `hwnd uintptr` 参数。
- `undoAnnotationAndInvalidate`：补 `hwnd uintptr` 参数。
- `invalidateAnnotationDirtyRect`：补 `hwnd uintptr` 首参（rect 在 rcx/rdi/rsi/r8）。

**关键知悉**：hwnd 自 `qrCodeSelectionWindowProc`(rax=hwnd) → `handleMessage`(rbx=hwnd) →
`handleAnnotationToolbarActions` → `applyAnnotationToolbarAction` 一路透传，是标注动作链第二参数；
旧注释"rbx 语义未确证"与 invalidateAnnotationDirtyRect"rbx/rcx/rdi/rsi = rect"均误读。

**下一批**：filelocator walkRoot/processFile（值传大结构体 prepared + ctx/generation/root/
results/startedAt，runSearch 调用点已读，待逐参定类型）；screenshot_uia_windows accessibleHitTest。

### 批次 236（captureScreenshotWithLauncherVisibility 签名订正）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1440 / P=56 / UNMARKED=0`（真函数 2732）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1441 / P=55 / UNMARKED=0`（真函数 2733 = 57.48%），
SHA256 `1CBD4A981CC6F76729DD7AC90ACFED398FD547218F0A8ED4C8DC5177E2B314F8`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/screenshot_windows.go`，详见 acceptance/batch236.md）：
- `captureScreenshotWithLauncherVisibility`：`[P]` → `[S-sig]`，签名
  `(service, hideLauncher bool) ScreenshotCaptureResult` → `(service, delay time.Duration,
  hideLauncher bool, cb func() ([]byte,bool,error)) ([]byte,bool,error)`。

**关键知悉**：morestack 保护 4 寄存器（rax=service、rbx=delay、cl=hideLauncher、rdi=cb）；
返回 6 字 = cb 透传 = `([]byte,bool,error)`，非旧注释臆测的 ScreenshotCaptureResult(16 字)。
cb 实现为调用者闭包（captureScreenshotAreaPNG.func1），内部 func1/func1.1 是 sync.Once 保护的
隐藏/恢复回调。本订正联动厘清 captureScreenshotAreaPNG 返回 = ([]byte,bool,error)。

**下一批**：captureScreenshotAreaPNG（7 参数，参数4/6/7 需 qrCodeNativeSelectionOptions 字段映射）；
captureScreenshotWindowSelectionWithOptions（5 参数 + WindowProcessPickResult 大结构返回）；
newScreenshotWindowSelectionSession（快照 6 字 + 4 栈参数）；filelocator walkRoot/processFile。

### 批次 237（initializeScreenshotCOMThreadMode 签名订正）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1441 / P=55 / UNMARKED=0`（真函数 2733）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1442 / P=54 / UNMARKED=0`（真函数 2734 = 57.50%），
SHA256 `83012FB171D4C9AE5804985724CF8132EC4C8B07B173B4AF294E28363B60C282`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/screenshot_uia_windows.go`，详见 acceptance/batch237.md）：
- `initializeScreenshotCOMThreadMode`：`[P]` → `[S-sig]`，签名
  `()` → `(allowChangedMode bool) (func(), error)`。
- `initializeScreenshotCOMSTAWorkerThread` 注释订正（"透传"→"调用并丢弃 (func(),error) 返回"）。

**关键知悉**：morestack 仅保存 al（单 bool 形参）；成功路径 newobject 构造 `{func2, success}`
闭包（func2=CoUninitialize 清理，读 `[rdx+8]` bool 决定是否反初始化）返回 (func,nil)；
RPC_E_CHANGED_MODE(0x80010106) 且 allowChangedMode → 同成功；否则 fmt.Errorf(HRESULT) 返回
(nil,err)。错误路径 rax=0x141096ad0（全局空函数表首元素 func1=ret 空函数，空清理哨兵）。

**下一批**：captureScreenshotAreaPNG（参数4/6/7 需 qrCodeNativeSelectionOptions 字段映射）；
captureScreenshotWindowSelectionWithOptions / newScreenshotWindowSelectionSession；
resolveControlHoverForWindow / screenshotAccessibleHitTest；filelocator walkRoot/processFile。

### 批次 238（screenshotNativePinDrawOverlay 签名订正）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1442 / P=54 / UNMARKED=0`（真函数 2734）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1443 / P=53 / UNMARKED=0`（真函数 2735 = 57.53%），
SHA256 `14a69e1f4f7cbe9e83853963064996b220d9e4f9606766672b84360c0c89a05c`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/screenshot_pin_native_windows.go`，一 [P] 升档，详见 acceptance/batch238.md）：
- `screenshotNativePinDrawOverlay`：`[P]` → `[S-sig]`，签名
  `(img, bounds)` → `(img *image.RGBA, bounds application.Rect, showCloseButton, showOpacityPanel, showHighlight bool, tick int, opacity float64)`。

**关键知悉**：morestack 序言保存 10 槽逐寄存器实证——r8b/r9b/r10b 为三路布尔门控
（close 按钮 / opacity 面板 / 高亮二次绘制，均 `test dl,dl; je` 存在性判定，非 color 分量）；
r11 为动画 tick（`test r11,r11; jle` + `bt r11d,0; jb` 作「正且偶」判定，高亮边框闪烁）；
xmm0 为 opacity（透传 OpacityThumbRect）。三条 ret 均无返回寄存器 = void。旧注释「r11 疑为
tick/边框厚度、语义未定」已解。

**下一批**：captureScreenshotAreaPNG（参数4/6/7 需 qrCodeNativeSelectionOptions 字段映射）；
captureScreenshotWindowSelectionWithOptions / newScreenshotWindowSelectionSession；
resolveControlHoverForWindow / screenshotAccessibleHitTest；launcherupdate_runtime.go 9 个 [P]；
filelocator walkRoot/processFile。

### 批次 239（screenshotCOMQueryWorkerPool.enqueue 签名订正）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1443 / P=53 / UNMARKED=0`（真函数 2735）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1444 / P=52 / UNMARKED=0`（真函数 2736 = 57.55%），
SHA256 `cedc2c3fc2adca50cedf21ec3f2ac7e3c1a6a70e81dc2c125d2f6cb784f04987`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/screenshot_uia_worker_windows.go`，一 [P] 升档，详见 acceptance/batch239.md）：
- `(*screenshotCOMQueryWorkerPool).enqueue`：`[P]` → `[S-sig]`，签名
  `(req *screenshotCOMQueryRequest)` → `(x, y int32, targetWindow uintptr, priority uint8, busyError error) (*screenshotCOMQueryRequest, error)`。

**关键知悉**：入参不是单一 req 指针，而是把 payload.point.X/Y（ebx/ecx int32）、
targetWindow（rdi uintptr）、priority（sil uint8）、busyError（r8/r9 error）作 6 个独立标量传入，
体内 newobject 组装请求对象（id 取 pool.nextRequestID，state=0，events=makechan(2)）。
返回三寄存器：成功 `(req,nil)`、busy/shutdown `(nil,err)` = `(*screenshotCOMQueryRequest, error)`。
旧注释「参数数量与结构超出单 req 指针，无法唯一确定」已解。

**下一批**：captureScreenshotAreaPNG（参数4/6/7 需 qrCodeNativeSelectionOptions 字段映射）；
captureScreenshotWindowSelectionWithOptions / newScreenshotWindowSelectionSession；
resolveControlHoverForWindow / screenshotAccessibleHitTest；launcherupdate_runtime.go 9 个 [P]；
filelocator walkRoot/processFile。

### 批次 240（probeLauncherUpdatePackageRange 签名订正）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1444 / P=52 / UNMARKED=0`（真函数 2736）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1445 / P=51 / UNMARKED=0`（真函数 2737 = 57.57%），
SHA256 `df43216244c069d13fb1cc49cca4452861ab2ab40ce1e163370a32cf453e18a9`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/launcherupdate_runtime.go`，一 [P] 升档，详见 acceptance/batch240.md）：
- `probeLauncherUpdatePackageRange`：`[P]` → `[S-sig]`，签名
  `(ctx context.Context, url string) (int64, error)` →
  `(ctx context.Context, access LauncherNetworkAccess, url string) (bool, int64, error)`。

**关键知悉**：rcx/rdi 是 LauncherNetworkAccess 接口（itab+data，Do 方法经 itab fun[0]=+0x18 调用
`Do(data, req, maxBytes)`），rsi/r8 是 url（ptr/len，r8 为长度非标量）。返回四寄存器：AL=bool
（Content-Range 确定态，仅 206+total>0 为 1）、BX=int64（total）、CX/DI=error。旧签名多漏
access 参数且漏 bool 返回。

**下一批**：captureScreenshotAreaPNG（参数4/6/7 需 qrCodeNativeSelectionOptions 字段映射）；
captureScreenshotWindowSelectionWithOptions / newScreenshotWindowSelectionSession；
resolveControlHoverForWindow / screenshotAccessibleHitTest；launcherupdate_runtime.go 8 个 [P]；
filelocator walkRoot/processFile。

### 批次 241（downloadLauncherUpdatePackageRange 签名订正）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1445 / P=51 / UNMARKED=0`（真函数 2737）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1446 / P=50 / UNMARKED=0`（真函数 2738 = 57.59%），
SHA256 `e8d5bbe189b7dd73054bc07eb69d17a1ccaf6e20907d0b360737e6425306bde1`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/launcherupdate_runtime.go`，一 [P] 升档 + 新增 `os` import，详见 acceptance/batch241.md）：
- `downloadLauncherUpdatePackageRange`：`[P]` → `[S-sig]`，签名
  `(ctx, url, r launcherUpdatePackageRange, dir string) error` →
  `(ctx, access LauncherNetworkAccess, url, file *os.File, r launcherUpdatePackageRange) error`。

**关键知悉**：9 槽逐寄存器实证——rcx/rdi=access（itab+data，Do 经 fun[0]=+0x18）、rsi/r8=url、
r9=file（os.File.WriteAt 接收者）、r10/r11=r（Start/End，cmp r10,r11 求 len=End-Start+1）。
旧签名漏 access、误把 dir string 当下载目标（实为已打开的 file，体内直接 WriteAt 落盘）。

**下一批**：captureScreenshotAreaPNG（参数4/6/7 需 qrCodeNativeSelectionOptions 字段映射）；
captureScreenshotWindowSelectionWithOptions / newScreenshotWindowSelectionSession；
resolveControlHoverForWindow / screenshotAccessibleHitTest；launcherupdate_runtime.go 7 个 [P]；
filelocator walkRoot/processFile。

### 批次 242（downloadLauncherUpdatePackageSequentially 签名订正）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1446 / P=50 / UNMARKED=0`（真函数 2738）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1447 / P=49 / UNMARKED=0`（真函数 2739 = 57.61%），
SHA256 `485eae6a500c19ac30cd7d3796e6894bb61eb3c1f8bb29c3ec3e13f5b5f9500c`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/launcherupdate_runtime.go`，一 [P] 升档，详见 acceptance/batch242.md）：
- `downloadLauncherUpdatePackageSequentially`：`[P]` → `[S-sig]`，签名
  `(ctx, pkg LauncherUpdatePackageConfig, dir string) (int64, error)` →
  `(ctx, taskID int64, access LauncherNetworkAccess, path, url, sha256 string, expectedSize int64) (int64, error)`。

**关键知悉**：8 槽 + stack 5 槽逐寄存器实证——rbx/rcx=ctx、rdi=taskID（setProgress 第二参）、
rsi/r8=access（Do 经 fun[0]=+0x18）、r9/r10=path（os.OpenFile name）、stack 上 url/sha256/expectedSize
三个独立参。旧签名把 pkg 整体压结构体、dir 当唯一目标——实为 5 个独立标量/字符串参。

**下一批**：captureScreenshotAreaPNG（参数4/6/7 需 qrCodeNativeSelectionOptions 字段映射）；
captureScreenshotWindowSelectionWithOptions / newScreenshotWindowSelectionSession；
resolveControlHoverForWindow / screenshotAccessibleHitTest；launcherupdate_runtime.go 6 个 [P]；
filelocator walkRoot/processFile。

### 批次 243（filelocator walkRoot/processFile 签名订正）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1447 / P=49 / UNMARKED=0`（真函数 2739）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1449 / P=47 / UNMARKED=0`（真函数 2741 = 57.65%），
SHA256 `8e78592bda4dde01bb5d8daacc2d19c9f1b9431d8503a0c881e67a35003b4a11`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/filelocator_runtime.go` 双 [P] 升档 + `backend/types_filelocator.go` 新增 1 类型，详见 acceptance/batch243.md）：
- `walkRoot`：`[P]` → `[S-sig]`，签名
  `(prepared *fileLocatorPreparedSearch, root string, ctx context.Context) error` →
  `(prepared fileLocatorPreparedSearch, ctx context.Context, generation uint64, path string, progress *fileLocatorSearchProgress, onFileDone func(bool), startedAt time.Time) error`。
- `processFile`：`[P]` → `[S-sig]`，签名
  `(prepared *fileLocatorPreparedSearch, path string, ctx context.Context) (bool, error)` →
  `(prepared fileLocatorPreparedSearch, ctx context.Context, generation uint64, path string, info os.FileInfo, progress *fileLocatorSearchProgress, onFileDone func(bool), startedAt time.Time) error`。
- 新增 `fileLocatorSearchProgress` 累加器类型（0x78 字节，计数器 + results 切片 + truncated，字段级语义 [P]）。

**关键知悉**：walkRoot 是 recv+7 寄存器（morestack 存 rax..r10），processFile 是 recv+8（存 rax..r11）。
`prepared` 为**值传大结构**（288B 走栈，非指针）。processFile 返回 **error 非 (bool,error)**（各尾迹仅清
rax/rbx 两字）。`onFileDone` 是 func(bool) 闭包（processFile 尾迹 `mov rdx,[0x3f8]; mov rcx,[rdx]; mov eax,1; call rcx`）。
runSearch 调用点（0x1407d0543）实锤：r9=&progress、r10=&onFileDone 闭包，闭包捕获 &[0x230] 状态 + receiver。

**下一批**：launcherupdate_runtime.go 6 个 [P]（该族与 downloadLauncherUpdatePackageWithWorkspace
栈参耦合，须先落地调用方全形参）；截图域（captureScreenshotAreaPNG / captureScreenshotWindowSelectionWithOptions /
newScreenshotWindowSelectionSession / resolveControlHoverForWindow / screenshotAccessibleHitTest）；
oledblackout_windows（25）；nativedrag createTemporaryDirectoryShortcut；§10 差集 54 文件。

### 批次 244（launcherupdate 下载链三函数签名订正）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1449 / P=47 / UNMARKED=0`（真函数 2741）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1451 / P=45 / UNMARKED=0`（真函数 2743 = 57.69%），
SHA256 `b8f2a593da6f09623d4e0dab7415795db624aee4f9ef0def1cb08db0a888187c`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/launcherupdate_runtime.go`，2 升档 + 1 修正，详见 acceptance/batch244.md）：
- `downloadLauncherUpdatePackageWithWorkspace`：`[P]` → `[S-sig]`，
  `(pkg LauncherUpdatePackageConfig) error` → `(workspace WorkspaceLayout, ctx, taskID, path, concurrent bool, url, sha256, totalSize, root) (int64, error)`。
- `downloadLauncherUpdatePackageConcurrently`：`[P]` → `[S-sig]`，
  `(ctx, pkg, dir) (int64, error)` → `(ctx, taskID, access, total int64, path, concurrent bool, url, sha256, totalSize, root) (int64, error)`。
- `downloadLauncherUpdatePackageSequentially`：订正批 242 签名（栈参实为 8 槽），补 `concurrent bool` 与 `root string`。

**关键知悉**：调用链 prepare(0x1408b8160)→download(0x1408b8f20)→concurrently(0x1408ba9a0)/sequentially(0x1408b9a60)
三方互证。`workspace` 是 `WorkspaceLayout` **值传**（栈首参 160B，ConfigFile@+0x10 传 newLauncherNetworkAccess），
非 LauncherUpdatePackageConfig 打包体。download 栈参 8 槽（concurrent+url+sha256+totalSize+root）经 4 xmmword
透传 concurrently（frame 0x118，栈首参 @0x128）/sequentially（frame 0x1d0，栈首参 @0x1e0）。批 242 的 sequentially
漏 concurrent bool 与 root string 两参。download/concurrently 均返回 `(int64, error)`。

**下一批**：launcherupdate_runtime.go 剩余 4 个 [P]（runLauncherUpdateTask 7 参 /
fetchLauncherRemoteConfigWithWorkspace——workspace 已证 WorkspaceLayout，可顺藤解码 /
beginLauncherUpdateTask 6 寄存器返回 / prepareLauncherUpdatePackageWithWorkspace——本次已从调用点反推全形参）；
截图域；oledblackout_windows（25）；nativedrag createTemporaryDirectoryShortcut；§10 差集 54 文件。

### 批次 245（launcherupdate 域 [P] 全清零）

**基线**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1451 / P=45 / UNMARKED=0`（真函数 2743）。
**收口**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1455 / P=41 / UNMARKED=0`（真函数 2747 = 57.78%），
SHA256 `5f8132cd7b71df5dd500cbdd69294521ee435b3ab638ecff737f388512cf751d`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/launcherupdate_runtime.go`，4 升档 + 1 双参订正，详见 acceptance/batch245.md）：
- `fetchLauncherRemoteConfigWithWorkspace`：`[P]`→`[S-sig]`，`()`→`(workspace WorkspaceLayout, ctx) (LauncherConfig, error)`。
- `prepareLauncherUpdatePackageWithWorkspace`：`[P]`→`[S-sig]`，`(pkg) (string, error)`→`(workspace, ctx, taskID, executableName, concurrent, url, sha256, totalSize, root) (string, int64, error)`。
- `runLauncherUpdateTask`：`[P]`→`[S-sig]`，`(ctx, version)`→`(ctx, taskID, done func(), version)`。
- `beginLauncherUpdateTask`：`[P]`→`[S-sig]`，`(version) (int64, error)`→`(ctx) (context.Context, int64, func(), error)`。
- `isLauncherVersionUpdateAvailableText`：订正双参 `(version, remoteVersion string) bool`。

**关键知悉**：beginLauncherUpdateTask 第二参是 **ctx 非 version**——InstallLauncherUpdate（0x1408b35a0）传
`context.Background()` 全局 itab/data，version 经闭包捕获（0x1408b366d 段）传 runLauncherUpdateTask goroutine。
beginLauncherUpdateTask 返回 6 寄存器 = `(context.Context, int64, func(), error)`（成功 AX/BX=WithCancel ctx、
CX=taskID、DI=done 闭包、SI/R8=error）。runLauncherUpdateTask 的 rsi=done func()（尾声 0x1408b3d70 无参 call）。
isLauncherVersionUpdateAvailableText 实为双参（0x1408b38e9 设 AX/BX=version + CX/DI=远程版本）。

**下一批**：剩余 [P] 集中 oledblackout_windows（~25）/screenshot_windows（~7）/screenshot_uia_windows（~5-6）/
mousegestures（1）/nativedrag_windows（1）/oledblackout（1 ghost）；补缺失函数 `normalizeLauncherUpdatePackageRoot`
（源码 602-613 行）；§10 差集 54 文件；FUNCS 缺口 697 个未落地顶层函数 + 闭包/方法。

### 批次 246（normalizeLauncherUpdatePackageRoot 函数体订正）

**基线/收口**：计数不变 `FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1455 / P=41 / UNMARKED=0`
（该函数本已标记 [S] 计入），SHA256 `d03e41f104dc2a5d2a6d67298fbe7c697584754ebd0c19a964a781dd3220839e`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批订正**（`backend/launcherupdate_plan.go:332`，详见 acceptance/batch246.md）：
- 旧（错误 [S]）：`TrimSpace → filepath.Clean → filepath.Abs`。
- 新（asm 直译 0x1408b6b60）：`strings.Replace(root,"\\","-",-1) → TrimSpace → 空/含"-"/"."/".." 返回空`。
- 常量实证：old=0x1411cac40(0x5c `"\\"`)、new=needle=0x140c3362f(0x2d `"-"`)。

**关键知悉**：该函数曾标 [S] 但函数体错误（Clean/Abs 与真实 asm 不符），本轮以 va_dump.py 提取完整
asm（0x1408b6b60→0x1408b6c40）逐条直译订正。normalizeLauncherUpdatePackageRoot 语义：更新包根目录名
须为单个目录名（拒绝路径分隔符 `\`→`-` 后检测、拒绝 `"."`/`".."`/空）。

**下一批**：P=41（oledblackout_windows 25 / screenshot_windows 7 / screenshot_uia_windows 6 /
mousegestures 1 / nativedrag_windows 1 / oledblackout 1 ghost）；§10 差集 57 文件；FUNCS 缺口 697。

### 批次 247（oledblackout ghost [P] 清零 → ensureLifecycleLocked [S]）

**基线/收口**：`FUNCS=2788 / S=1256→1257 / S-inline=36 / S-sig=1455 / P=41→40 / UNMARKED=0`
（真函数 2747→2748 = 57.80%），SHA256 `457410aa5600be56a7ecc0c127bbce0bec9e86d5f7c53ff5133ba1284acbbc85`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（`backend/oledblackout.go`，详见 acceptance/batch247.md）：
- 删除 ghost 脚手架 `newOLEDLifecycleContext`（`[P]`，无独立符号，实为方法）。
- 新增 `(*oledBlackoutService).ensureLifecycleLocked`（`[S] 0x1408ff220`，单 receiver 无参无返回）。
- `newOLEDBlackoutService` 调用点改为 `s.ensureLifecycleLocked()`（对齐 0x1408fef60 流程）。

**关键知悉**：ensureLifecycleLocked 幂等初始化生命周期字段——shutdownDone==nil → make(chan struct{})；
lifecycleGeneration==0 → 1；lifecycleContext==nil → context.WithCancel(Background)；shuttingDown!=0 →
lifecycleCancel()。字段偏移对齐 types_oled.go：[0x88]=shuttingDown、[0x90/0x98]=lifecycleContext、
[0xa0]=lifecycleCancel、[0xa8]=lifecycleGeneration、[0xb0]=shutdownDone。

**下一批**：P=40（oledblackout_windows 25 / screenshot_windows 7 / screenshot_uia_windows 6 /
mousegestures 1 / nativedrag_windows 1）。oledblackout `*Context` 系 [P] 为闭包展开（捕获结构体 +
方法字段），批量解码需先定捕获结构体布局；`createTemporaryDirectoryShortcut`（nativedrag）为
0 寄存器参数 + 大结构体返回；§10 差集 57 文件；FUNCS 缺口 697。

### 批次 248（已存在文件短函数批量 [S] 落地 +11 函数）

**基线/收口**：`FUNCS=2788→2799 / S=1257→1268 / S-inline=36 / S-sig=1455 / P=40 / UNMARKED=0`
（真函数 2748→2759 = 58.03%），SHA256 `aea815511cc97fadaeffe66cd6321c668a752325cae5b606ee7155c716d8a11b`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

**本批落地**（7 个已存在文件，详见 acceptance/batch248.md）：
- `filesearch_index_windows.go`：`(*IndexNode).IsDirectory`（Flags&1）、`IsDeleted`（Flags&2）。
- `oledblackout.go`：`beginBackgroundActivity(generation uint64) bool`（锁+defer+代际校验+WaitGroup.Add(1)）、
  `endBackgroundActivity`（Add(-1)）。
- `main.go`：`closeLauncherStartupDebugLog(*os.File)`（nil 检查 + Close）。
- `oledblackout_windows.go`：`request(hide,show,close bool,response chan error) error`（nil 错误
  "鼠标控制器不可用" + 锁 + closed + makechan + commands<- + <-响应）、`Hide`/`Show`（request 包装）。
- `screenshot_cursor_windows.go`：`screenshotCursorCurrentlyShowing`（GetCursorInfo 失败保守 true）、
  `adjustScreenshotCursorVisibility(visible bool) error`（64 次 ShowCursor 循环 + Errno(0) 哨兵）、
  `ensureScreenshotCursorVisible`（已可见则返回否则 adjust(true)）+ `procShowCursor` LazyProc。

**关键知悉**：本批首次用 `tools/aggregate_gap.py`（修复 load_landed 接收者正则，支持 `(var *Type)` 带变量名
形式）重生成 `gap_aggregate.txt`（2084 缺失，其中顶层非闭包 918），替换过时的 diff_funcs/diff_truly_missing
（后者仍列出已落地的 launcherinstance/shellverb 函数，且统计口径漏闭包）。落地前一律 grep backend 复核。

**下一批**：P=40（oledblackout_windows 25 / screenshot_windows 7 / screenshot_uia_windows 6 /
mousegestures 1 / nativedrag_windows 1）。继续按 `gap_aggregate.txt`（长度升序）落地已存在文件短函数；
`restoreScreenshotCursorAfterOverlay`（0x140973be0）、`Close`（0x140914120）等 cursor/oled 域剩余函数待续；
FUNCS 缺口约 1955（2799/4754 = 58.88%）。

### 批次 249（oledblackout 域短函数批量 [S] +9 + normalizeLauncherUpdatePackageRoot 常量订正）

**基线/收口**：`FUNCS=2799→2808 / S=1268→1277 / S-inline=36 / S-sig=1455 / P=40 / UNMARKED=0`
（真函数 2759→2768 = 58.22%），SHA256 `d2d808b497d49a1ee576605f3ce77bd20f5718535c918343d4fc7fa6c3e4313d`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,322,240 B）。

**本批落地**（3 文件 +9 [S]，详见 acceptance/batch249.md）：
- `oledblackout.go`（+7，新增 import path/filepath/strings/time）：normalizeOLEDBlackoutMediaPauseExclusionPath
  （TrimSpace→空→filepath.Clean）、oledBlackoutMediaPausePathBase（TrimSpace→TrimRight(`\/`)→空→Replace(`\`→`/`)
  →LastIndex(`/`)<0 返整串否则 idx+1 切片）、normalizeOLEDBlackoutMediaPauseExclusionProcessName（TrimSpace→空→
  PathBase→空/`.`/`..` 返空→ToLower）、idleRunCurrentLocked（5 条件→false 否则全局函数值间接调用）、
  stopFocusRetryLocked（focusRetryTimer!=nil→Stop→nil；scheduleID++）、browserMediaContinuity.clear（锁+entries=nil）、
  armInputDismissGuardLocked（inputDismissGuardUntil=Now().Add(250ms)）。
- `filesearch_index_windows.go`：volumeIndexReadLease.Release（幂等 released 置位 + release 回调）。
- `bootstrapservice.go`：GetPinnedScreenshotStates（screenshotPin nil→nil 否则 ListStates）。

**本批订正**（`launcherupdate_plan.go` normalizeLauncherUpdatePackageRoot，batch 246 常量误读）：
- 旧 `\`→`-` + Contains(`-`)；新 `\`→`/` + Contains(`/`)。
- 实证：0x140c3362f 复核为 `/`（0x2f）非 `-`（0x2d）；PathBase 的 Replace-new 与 LastIndex 分隔符同为该地址。

**关键知悉**：va_read 读 RIP 相对地址务必 64 位整体相加（本轮发现多处「忘进位」把 .rdata 地址算进
.text 段读成 UTF-8 乱码，0x1411cac40 vs 0x140c8ac40 即典型高位进位差）。oledBlackoutService 偏移复核：
moduleEnabled(+0x70)/shuttingDown(+0x88)/lifecycleGeneration(+0xa8)/idleScheduleID(+0x130)/
focusRetryTimer(+0x150)/focusRetryScheduleID(+0x158)/inputDismissGuardUntil(+0x180)。

**下一批**：P=40。dismissOverlayForKeyboardInputLocked 依赖 visibleOverlayForKeyboardInputLocked(672B)+
dismissOverlayForInputLocked(2176B)；oledBlackoutReadInputSnapshot 依赖平台快照函数表（.data 全局函数值
0x141bc1b30..0x141bc1b48）。继续按 gap_aggregate.txt 长度升序落地；FUNCS 2808/4754 = 59.07%。

### 批次 250（BootstrapService/ShellVerb 短函数 +4）

**基线/收口**：`FUNCS=2808→2812 / S=1277→1280 / S-inline=36 / S-sig=1455→1456 / P=40 / UNMARKED=0`
（真函数 2768→2772 = 58.31%），SHA256 `253923d58e5b5aa7c1422bc63aa94c82f889caa52172bb5465e84dab53f3b34d`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,322,240 B）。

**本批落地**（2 文件 +4，详见 acceptance/batch250.md）：
- `shellverb.go`：`(*BootstrapService).InvokeShellVerb(verb,path string) error` [S 0x1408a6920]
  = 参数重排后尾调用 invokeShellVerb（该自由函数已落地）。
- `bootstrapservice.go`：`emitScreenshotCaptureAccepted` [S-sig 0x140797080]（签名实证，体为事件分发链）；
  `(*screenshotCaptureAcceptedNotifier).Emit` [S 0x140796fe0]（nil 检查 + once.Do(func1)，func1 解引用
  n.service）；`clearStartupTrayMode` [S 0x140796820]（lock(+0x540) + startupTrayMode(+0x449)=false + unlock）。

**关键知悉**：screenshotCaptureAcceptedNotifier={service *BootstrapService, once sync.Once}（[0]=service、
[8]=once）。InvokeShellVerb 签名由 shellverb_test.go 的 `invokeShellVerb("", "x")`/`("open", "  ")` 用例锁定
`(verb, path string) error`。applyLauncherWindowSizingForShow 依赖 consumeLauncherDefaultSizeReset +
loadLauncherUIScalePercent（均未落地），LaunchApp 为栈传大结构体 + LaunchAppWithPrivilege 尾调用。

**下一批**：P=40。继续按 gap_aggregate.txt 长度升序落地已存在文件短函数；FUNCS 2812/4754 = 59.15%。

### 批次 251（BootstrapService/FileSearch 短函数 +3 + ensureWorkspaceDirectories 改名订正）

**基线/收口**：`FUNCS=2812→2814 / S=1280→1283 / S-inline=36 / S-sig=1456→1455 / P=40 / UNMARKED=0`
（真函数 2772→2774 = 58.35%），SHA256 `9a16bed6d9f6c6b3502293f42ec6a14440ec9583baa9ad143e644684a04602d5`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,322,240 B）。

**本批落地**（2 文件 +2 [S] +1 转正，详见 acceptance/batch251.md）：
- `hotkey_dispatch_stubs.go`：finishScreenshotHotkeyCapture [S 0x140795260]（nil 检查 + lock(+0x540) +
  screenshotHotkeyActive(+0x27a)=false + unlock，独立符号，与 deferwrap1 同语义）。
- `filesearch_index_windows.go`（新增 import time）：timeToUnixNano [S 0x1407f8640] = t.UnixNano()。

**本批订正**（`bootstrapservice_config.go`）：ensureWorkspaceDirectoriesForService 改名
ensureWorkspaceDirectories（符号实证 symbols.txt:19195），[S-sig]→[S]，体修正为
`ensureWorkspaceDirectories(bs.workspaceSnapshot())`（asm 0x1407a26a0 先加锁快照再传参，旧版直取字段无锁）；
调用点 ChooseLauncherBackgroundImage 同步改名。

**关键知悉**：timeToUnixNano 为 time.Time.UnixNano 直包装（asm 常量 wallToInternal=0xdd7b17f80、
internalToUnix=0xa1b203eb3d1a0000、1e9=0x3b9aca00 与 Go 1.25 time 包一致）。

**下一批**：P=40。继续按 gap_aggregate.txt 长度升序落地已存在文件短函数；FUNCS 2814/4754 = 59.19%。

### 批次 252（filesearch 域内存映射关闭链 + 前缀桶计数 +3 [S]）

**基线/收口**：`FUNCS=2814→2817 / S=1283→1286 / S-inline=36 / S-sig=1455 / P=40 / UNMARKED=0`
（真函数 2774→2777 = 58.42%），SHA256 `6de4d49f118647832e8738a9ee9e04f30860a2bf1dbadb55e73a7187f7a6cbda`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,322,240 B）。

**本批落地**（filesearch_index_windows.go 新增 import windows，详见 acceptance/batch252.md）：
- addNamePrefixBucketCount [S 0x1407e2620]（首字节折叠小写 + firstByte/twoByte 桶累加）。
- volumeIndexMappedFile.Close [S 0x1407fd880]（倒序 UnmapViewOfFile + CloseHandle + file.Close，
  仅保留首个错误）。
- volumeNameTrigramIndex.close [S 0x1407e1920]（幂等关闭：取 mappedFile → 清 slice/指针 → Close）。

**关键知悉**：volumeIndexMappedFile={file(+0x00),mapping(+0x08),views(+0x10)}；
volumeNameTrigramIndex={version,signatures,signatureBytes,mappedFile(+0x38),closed(+0x40)}。
Close 的「仅保留首个错误」由 asm `je` 方向反推锁定。

**下一批**：P=40。继续按 gap_aggregate.txt 长度升序落地已存在文件短函数；FUNCS 2817/4754 = 59.26%。

### 批次 253（filesearch WAL 截断 + inputmonitor 平台拥有关闭链 +5）

**基线/收口**：`FUNCS=2817→2822 / S=1286→1290 / S-inline=36 / S-sig=1455→1456 / P=40 / UNMARKED=0`
（真函数 2777→2782 = 58.52%），SHA256 `3e62c33c96d25bb7a2bb0799d3cf53adaa3b835a8d06e6b1b554229dfc342cce`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,322,240 B）。

**本批落地**（2 文件 +5，详见 acceptance/batch253.md）：
- filesearch_index_windows.go（新增 import errors/os）：truncateOpenWALAtValidOffset [S 0x1408008e0]
  （nil/负偏移 → errInvalidWALOffset 占位 → Truncate → Sync）。
- inputmonitor.go：advancePlatformGeneration [S 0x140861b60]（lock + platformGeneration.Add(1)）；
  stopPlatformThread [S-sig 0x140866120]（1785B 大函数留待专项）；stopPlatformOwned [S 0x140861aa0] /
  closePlatformOwned [S 0x140861b00]（override 优先，否则停线程）。

**关键知悉**：truncateOpenWALAtValidOffset 哨兵错误静态值（.data @0x14193c650=0x30）为 Go 1.25
typeOff 编码，字符串多次定位落空，按「不伪造」纪律留占位。stop/closePlatformOwned 的 asm 偏移
（+0x108/+0x110）与字段声明存在 8 字节错位（疑 closeOnce 尺寸差异），落地以字段名分配偏移。
inputMonitorService 偏移锚点：lock(+0x08)、done(+0xa0)、platformGeneration(+0xd0)。

**下一批**：P=40。gap_aggregate.txt 已重新生成（total missing=2053/top-level=887）。继续按长度升序
落地已存在文件短函数；FUNCS 2822/4754 = 59.36%。

### 批次 254（oledblackout 短函数 +4 [S]）

**基线/收口**：`FUNCS=2822→2826 / S=1290→1294 / S-inline=36 / S-sig=1456 / P=40 / UNMARKED=0`
（真函数 2782→2786 = 58.60%）。`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

**本批落地**（1 文件 +4，全部 [S]，详见 acceptance/batch254.md）：
- oledblackout.go（新增 import sort/strconv）：
  oledBlackoutMediaPauseExclusionKey [S 0x1408fdf40]（normalizePath → ToLower → 非空 concat "path:"
  / 空 normalizeProcessName concat "url(http"，两常量 rodata 解码实证）；
  parseOLEDBlackoutScreenNumber [S 0x140906420]（收集 '0'-'9' → Atoi，空/错则 0）；
  sortOLEDBlackoutScreens [S 0x140906280]（[]*OLEDBlackoutScreen，编号升序，相同则 Name 字符串序）；
  oledBlackoutReadInputSnapshot [S 0x14090be40]（经两全局函数值调 cursor/keys，返回 (x,y,ok,keys)）。

**关键知悉**：sortOLEDBlackoutScreens 元素为 8 字节指针（`[]*OLEDBlackoutScreen`），比较 Name
（元素 +0x10=ptr/+0x18=len）。oledBlackoutReadInputSnapshot 的两全局函数值 @0x141BC1B30/
@0x141BC1B38 → funcval fn=oledBlackoutCurrentCursorPhysicalPoint@0x140914aa0 /
oledBlackoutPressedKeyboardKeys@0x140914ac0，落地以直接调用等价还原。oledBlackoutProfileKey
（依赖 normalizeOLEDBlackoutProfileScreens 2240B）与 dismissOverlayForKeyboardInputLocked
（依赖 visibleOverlayForKeyboardInputLocked/dismissOverlayForInputLocked 方法链）留待专项。

**下一批**：P=40。gap_aggregate.txt top-level=883。继续按长度升序落地已存在文件短函数；
下一批优先 windowmanagement_windows.go 的 wrap/rect 短函数链（128–192B）。
FUNCS 2826/4754 = 59.44%。

### 批次 255（windowmanagement_windows 短函数 +5 [S]）

**基线/收口**：`FUNCS=2826→2831 / S=1294→1299 / S-inline=36 / S-sig=1456 / P=40 / UNMARKED=0`
（真函数 2786→2791 = 58.71%）。`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

**本批落地**（1 文件 +5，全部 [S]，详见 acceptance/batch255.md）：
- windowmanagement_windows.go（新增 import unsafe + procGetWindowLongPtrW/procGetClientRect）：
  windowManagementPointInCornerGuard [S 0x1409ed9a0]（radius<=0||right<=left||top>=bottom→false，
  margin=clamp(radius+4,0,96) 截到 width/height，返回 (左||右)&&(上||下)带）；
  windowManagementWrapTargetX [S 0x1409eda40]（y∈[Top,Bottom) 且 Right>Left，direction→
  max(best,Right-1)/min(best,Left)，返回 best∓3）；
  windowManagementWrapTargetY [S 0x1409edb00]（x∈[Left,Right) 且 Bottom>Top，对称）；
  windowManagementGetWindowLongPtr [S 0x1409ee320]（显式 Find 预加载，失败→nil→Call panic，
  [2]uintptr{hwnd,index} 后 Call 返回 r1）；
  windowManagementGetClientRectValue [S 0x1409ee780]（GetClientRect(hwnd,&rect)，r1==0→
  (0,0,0,0)，否则 (Left,Top,Right-Left,Bottom-Top) int32→int）。

**关键知悉**：WrapTargetX/Y 命名交叉——X 函数输入垂直 y 返回水平 x，Y 函数反之（WrappedCursorPoint
调用点实证）。windowManagementRECT=标准 RECT（Left/Top/Right/Bottom，16B int32）。GetWindowLongPtr
的 Find+cmovne 全局值2=0x140B3A7A0（procGetWindowLongPtrW）、全局值1=0（nil 回退），非 A/W 切换。

**遗留订正**：windowManagementVirtualScreenBounds 现返 image.Rectangle（4×int64），asm 实返
4×int32（32 位 add/lea）；与调用方 getLauncherBackgroundMetrics（bootstrapservice_callees.go @312）
联动订正留专项批次。windowManagementStyleNames 依赖未落地 FlagNames/ExStyleNames，跳过。

**下一批**：P=40。继续按 gap_aggregate.txt 长度升序落地 windowmanagement_windows.go 剩余短函数
（GetClassName 256B / MaybeWrapCursor 288B / EnumDisplayMonitorProc 320B / GetCursorPoint 320B 等）。
FUNCS 2831/4754 = 59.55%。

### 批次 256（截图/桌面小组件短函数 +7 [S]）

**基线/收口**：`FUNCS=2831→2838 / S=1299→1306 / S-inline=36 / S-sig=1456 / P=40 / UNMARKED=0`
（真函数 2791→2798 = 58.86%）。`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

**本批落地**（4 新文件 +7，全部 [S]，详见 acceptance/batch256.md）：
- screenshot_preview_windows.go：screenshotPreviewShowWindowFlags [S 0x14099e7a0]（return 0x53=83）、
  screenshotPreviewShowWindowCommand [S 0x14099e7c0]（return 4=SW_SHOWNOACTIVATE）。
- screenshot_selection_toolbar_windows.go：screenshotSelectionToolbarShowWindowCommand
  [S 0x1409ad200]（return 8=SW_SHOWNA）、screenshotSelectionToolbarShowWindowFlags [S 0x1409ad220]
  （return 0x53=83）。
- desktopwidgets_weather.go：desktopWeatherInt [S 0x1407c9640]（Atoi(TrimSpace(s)) →
  (int,error)）、(*desktopWidgetWeatherService)Shutdown [S 0x1407c5180]（nil/`cancel==nil` 守卫后 s.cancel()）。
- screenshot_scroll_canvas_windows.go：(*screenshotScrollingChunkedCanvas)Bounds [S 0x14099f520]
  （nil/宽高<=0 → 零矩形，否则 image.Rect(0,0,width,height)）。

**关键知悉**：ShowWindowFlags 两处均返 0x53=83（0x40|0x10|0x02|0x01，非标准 WS_*/WS_EX_* 组合，
按 asm 直译字面常量）。Shutdown 偏移 cancel@+0x28、Bounds 偏移 width@+0x00/height@+0x08 与
types 声明吻合。本批重跑 aggregate_gap.py 刷新缺口：total missing=2040、top-level=874。

**下一批**：P=40。继续按 gap_aggregate.txt 长度升序落地 windowmanagement_windows.go 剩余短函数
（GetClassName 256B / MaybeWrapCursor 288B / EnumDisplayMonitorProc 320B / GetCursorPoint 320B）
及 desktopwidgets_weather.go 的 desktopWeatherFloat（128B）。FUNCS 2838/4754 = 59.70%。

### 批次 257（输入/远程图标/工作区迁移/天气浮点 +4 [S]）

**基线/收口**：`FUNCS=2838→2842 / S=1306→1310 / S-inline=36 / S-sig=1456 / P=40 / UNMARKED=0`
（真函数 2798→2802 = 58.94%）。`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

**本批落地**（+4 [S]，1 追加 + 3 新文件，详见 acceptance/batch257.md）：
- desktopwidgets_weather.go：desktopWeatherFloat [S 0x1407c95c0]（ParseFloat(s,64) + NaN/上界
  MaxFloat64/负值三层守卫 → 越界返 0）。
- inputmonitor_windows.go：inputMonitorXButtonLabel [S 0x1408695e0]（HIWORD==1→"x1"、==2→"x2"、
  否则 "x"+FormatInt(HIWORD,10)）。
- remoteicons.go：remoteIconContentTypeAllowed [S 0x140961fe0]（ParseMediaType(TrimSpace) err→false、
  否则 EqualFold(mediatype, allowed)）。
- workspacemigration_identity_windows.go：workspaceMigrationPathIsReparse [S 0x1409f4660]
  （UTF16PtrFromString→GetFileAttributes 任一步 err→false、否则 attrs&REPARSE_POINT(0x400)!=0）。

**关键知悉**：RIP-relative 目标地址初算多进 0x100000（`0x869605+0x3ca0a6=0xC336AB` 误算 `0xD096AB`），
读到垃圾后按「disp32 加到下一条指令地址」重算修正为 0x140C336xx/0x1411CD8xx。desktopWeatherFloat
下界常量 = 负零（-0），语义 `f < 0 → 0`。workspaceMigrationPathIsReparse 与 pluginPathIsReparse/
launcherUpdatePathHasReparsePoint 同族（本函数 error 时返 false）。timeFromWindowsTick 涉及
time.Time wall/monotonic 内部布局，留待专项。

**下一批**：P=40。继续按 gap_aggregate.txt 长度升序落地 windowmanagement_windows.go 剩余短函数
（GetClassName 256B / MaybeWrapCursor 288B / EnumDisplayMonitorProc 320B / GetCursorPoint 320B /
GetWindowText 320B）。FUNCS 2842/4754 = 59.78%。

### 批次 258（窗口管理 user32 薄封装 +4 [S]）

**基线/收口**：`FUNCS=2842→2846 / S=1310→1314 / S-inline=36 / S-sig=1456 / P=40 / UNMARKED=0`
（真函数 2802→2806 = 59.02%）。`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

**本批落地**（+4 [S]，全部追加到 windowmanagement_windows.go，详见 acceptance/batch258.md）：
- windowManagementGetClassName [S 0x1409ee220]（GetClassNameW.Call(hwnd,&buf[0],256) →
  TrimSpace(UTF16ToString(buf[:n]))；新增 procGetClassNameW）。
- windowManagementGetWindowText [S 0x1409ee0e0]（GetWindowTextW 两阶段取长度再取内容；
  新增 procGetWindowTextW）。
- windowManagementGetCursorPoint [S 0x1409edfa0]（GetCursorPos.Call(&pt)；成功→(X,Y,nil)；
  失败→lastErr==nil→errors.New("未知错误")→fmt.Errorf("读取鼠标位置失败: %w")；复用
  gpu_pick_windows.go 的 procGetCursorPos）。
- windowManagementEnumDisplayMonitorProc [S 0x1409ede60]（EnumDisplayMonitors 回调，dwData 为
  &[]windowManagementRECT，收集有效矩形；签名 dwData unsafe.Pointer）。

**关键知悉**：procGetCursorPos 已在 gpu_pick_windows.go:20 定义，首版重复声明被 build 拦下
（删本文件重复项复用已有）。EnumDisplayMonitorProc 首版 dwData uintptr + unsafe.Pointer(dwData)
触发 vet "possible misuse of unsafe.Pointer"，改 dwData unsafe.Pointer 后 vet 绿（asm 无差）。
GetCursorPoint 失败链实测字符串 "未知错误"(12B) + "读取鼠标位置失败: %w"(28B)。

**下一批**：P=40。windowmanagement_windows.go 剩余：MaybeWrapCursor 288B（依赖 WrappedCursorPoint
992B，需同批或先落地）、GetWindowRect 384B、MonitorRectForWindow 512B。FUNCS 2846/4754 = 59.87%。

### 批次 259（窗口矩形 + 快照归属校验 +2 [S]，含 258 忠实度订正）

**基线/收口**：`FUNCS=2846→2848 / S=1314→1316 / S-inline=36 / S-sig=1456 / P=40 / UNMARKED=0`
（真函数 2806→2808 = 59.07%）。`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

**本批落地**（+2 [S]，均追加到 windowmanagement_windows.go，详见 acceptance/batch259.md）：
- windowManagementGetWindowRect [S 0x1409ee600]（GetWindowRect.Call(hwnd,&rect)；成功→
  (Left,Top,Right,Bottom,nil)；失败→"未知错误"→"读取窗口位置失败: %w"）。
- windowManagementValidateSnapshotOwner [S 0x1409ea780]（IsWindow + GetWindowThreadProcessId
  双 proc 校验 pid；失败文案 "目标窗口已失效" / "目标窗口句柄已被其他进程复用"）。
新增 proc：procIsWindow / procGetWindowRect / procGetWindowThreadProcessId。

**批次 258 忠实度订正**：GetCursorPoint 失败判据原只写 `lastErr == nil`，漏 `syscall.Errno(0)`
分支；对照同库已固化 getCursorScreenPoint(0x14085cf00) 的 asm（cmp ErrnoItab + ifaceeq）订正为
`lastErr == nil || errors.Is(lastErr, syscall.Errno(0))`。

**关键纪律**：proc 身份一律以 LazyProc.Name 内存实证为准，不做语义推测。本批实证：
IsWindow(0x141BD1B40) / GetWindowThreadProcessId(0x141BD1BC0) / GetWindowRect(0x141BD1EC0)；
并预留 SetLastError(0x141BD1A40) / SetWindowLongW(0x141BD1D00) / SetWindowLongPtrW(0x141BD1C80)。

**RIP 进位第三次踩坑（已固化自检）**：`0x1409EA7BF + 0x011D7609 = 0x141BC1DC8`，前次误算
0x1411D1DC8。纪律：disp32 加下一条指令地址、逐位进位，读出后用 len 合理性自检（首读出
len=5381816026 即错址信号）。

**下一批**：FlagNames 544B / ExStyleNames 384B / StyleNames 128B（名字枚举链）；SetWindowLongPtr
544B（三 proc 已实证）；MaybeWrapCursor 288B（依赖 WrappedCursorPoint 992B）。
FUNCS 2848/4754 = 59.90%。

### 批次 260（窗口样式名枚举链 +3 [S]）

**基线/收口**：`FUNCS=2848→2851 / S=1316→1319 / S-inline=36 / S-sig=1456 / P=40 / UNMARKED=0`
（真函数 2808→2811 = 59.13%）。`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

**本批落地**（+3 [S]，均追加到 windowmanagement_windows.go，详见 acceptance/batch260.md）：
- windowManagementFlagNames [S 0x1409eede0]（value==0→""；`Flag==0 || value&Flag!=Flag` 跳过；
  `strings.Join(names," | ")`）。
- windowManagementStyleNames [S 0x1409eebe0]（17 条 WS_* 切片字面量）。
- windowManagementExStyleNames [S 0x1409eec60]（8 条 WS_EX_* 切片字面量）。
新增类型 windowManagementFlagName{Flag uint32; Name string}（stride 24B）。

**表内容全部内存实证**：StyleNames 17 条 / ExStyleNames 8 条，flag 值与 Win32 常量逐条吻合；
分隔符 " | "（3B @0x140C33CEC）。

**结构忠实度订正**：首版写成包级 var 数组 + `table[:]`，但 asm 显示两表是**每次调用在栈上构建**
（StyleNames 走 duffcopy 复制 408B、ExStyleNames 走 duffzero+逐条填入 192B），已改为切片字面量
直传，结构对齐二进制。

**RIP 进位第四次踩坑 → 纪律治本**：`0x1409EEC00 + 0x7F7610` 应为 0x1411E6210（手算漏进位成
0x141E6210）。此后一律用 `[long]` 显式加法并回显 computed 地址 + 校验读出内容合理性
（名字须为可打印 ASCII、len∈1..96）；本批 ExStyle 4 条错址即靠"读出非 WS_EX_ 前缀垃圾"发现。

**下一批**：SetWindowLongPtr 544B（SetLastError/SetWindowLongW/SetWindowLongPtrW 三 proc 已实证，
格式串 "更新窗口样式失败: %w"）；DisplayRects 384B；MaybeWrapCursor 288B（依赖 WrappedCursorPoint）。
FUNCS 2851/4754 = 59.97%。

### 批次 261（SetWindowLongPtr +1 [S]）

**基线/收口**：`FUNCS=2851→2852 / S=1319→1320 / S-inline=36 / S-sig=1456 / P=40 / UNMARKED=0`
（真函数 2811→2812 = 59.19%）。`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

**本批落地**：windowManagementSetWindowLongPtr [S 0x1409ee3e0]（LockOSThread + defer
UnlockOSThread；SetLastError.Call(0) 预清；SetWindowLongPtrW 为首选、Find 失败退回 SetWindowLongW；
Call(hwnd,index,newValue)）。

**关键发现：同一 Errno(0) 判据在本域内语义相反（须分别落地）**
- GetCursorPoint / GetWindowRect：`Errno(0)` → **失败**，替换为 "未知错误" 后返回 error。
- SetWindowLongPtr：`Errno(0)` → **成功**，直接返回 nil（r1 是「旧值」，0 合法，改由 lastErr 判定）。
两处 asm 分支结构不同（前者构造 errors.New，后者 `test rcx,rcx; je` 走成功返回）。禁止"同域同判据"复用。

**全部事实内存实证**：
- defer 目标 funcval @0x141096E10 → code=0x1400525E0 = **runtime.UnlockOSThread**。
- 三 proc 槽：0x141BC1DA8=SetLastError、0x141BC1DF0=SetWindowLongPtrW、0x141BC1E00=SetWindowLongW。
- Errno itab `0x1409ee4f5+7+0x7e47eb = 0x1411D2CE0`，与批次 259 GetCursorPoint 算出的一致（交叉验证通过）。
- 格式串 `更新窗口样式失败: %w`（28B @0x140C6C1E0）。
- 全量扫 `x/sys@v0.46.0/windows/*.go` 确认**无 SetLastError**（只有 GetLastError），故新增
  `kernel32DLL = windows.NewLazySystemDLL("kernel32.dll")` + `procSetLastError`。

**下一批**：DisplayRects 384B；MaybeWrapCursor 288B（依赖 WrappedCursorPoint 992B，须先落地）；
MonitorRectForWindow 512B；targetFromWindowProcessPick 416B。FUNCS 2852/4754 = 60.00%。

### 批次 262（MonitorRectForWindow +1 [S] + VirtualScreenBounds 返回类型订正）

**基线/收口**：`FUNCS=2852→2853 / S=1320→1321 / S-inline=36 / S-sig=1456 / P=40 / UNMARKED=0`
（真函数 2812→2813 = 59.24%）。`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

**遗留订正（HANDOFF 挂账项已清）**：`windowManagementVirtualScreenBounds` 原返回 `image.Rectangle`
（4×int64），**实为 4×int32**。两条独立证据：① 自身尾部 asm `add edx,ecx` / `lea edi,[rax+rbx]`
全 32 位；② 调用方 getLauncherBackgroundMetrics(0x140874520) `sub ecx,eax` + `movsxd rcx,ecx`
（32 位减法后符号扩展，若 int64 不会有 movsxd）。已改为返回 `windowManagementRECT`，
同步订正 bootstrapservice_callees.go（唯一调用点），并移除随之失效的 `image` 导入。

**本批落地**：windowManagementMonitorRectForWindow [S 0x1409ee840]（MonitorFromWindow(hwnd,2) →
hMonitor==0 → errors.New("无法定位目标窗口所在显示器")；GetMonitorInfoW(hMonitor,&mi{Size:40})；
失败→"未知错误"→"读取显示器边界失败: %w"；成功→Monitor 矩形四值）。
新增 proc：procMonitorFromWindow / procGetMonitorInfoW。

**proc 身份内存实证**：0x141BC1DB0=GetSystemMetrics、0x141BC1E60=EnumDisplayMonitors、
0x141BC1E50=MonitorFromWindow、0x141BC1E58=GetMonitorInfoW。
**Errno itab 第三次交叉验证**：`0x1409ee910+7+0x7e43d0 = 0x1411D2CE0`，与 259/261 完全一致。

**工具纪律**：disp32 一律从 .bin 原始字节按 `next+disp` 解码，不人工读十六进制；新增 PE 段表解析
用于定性地址归属（本批确认 0x141C5AB60 属 .data，即 init 期写入的包级 funcval 槽）。

**未落地（留下一批专项）**：DisplayRects 384B（回调经全局 funcval 0x141C5AB60 传递，构造方式待实证）；
targetFromWindowProcessPick 416B（大结构栈展开待逐字段核对）；MaybeWrapCursor 288B（依赖
WrappedCursorPoint 992B）。FUNCS 2853/4754 = 60.02%。

### 批次 263（DisplayRects +1 [S]：EnumDisplayMonitors 回调范式首落地）

**基线/收口**：`FUNCS=2853→2854 / S=1321→1322 / S-inline=36 / S-sig=1456 / P=40 / UNMARKED=0`
（真函数 2813→2814 = 59.29%）。`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

**本批落地**：windowManagementDisplayRects [S 0x1409edce0]（SM_CMONITORS → count=max(r1,1) →
makeslice(0,count) → EnumDisplayMonitors(0,0,callback,&rects) → r1!=0&&len!=0 返回 rects；
否则 VirtualScreenBounds 回退单元素或 nil）。新增 proc：procEnumDisplayMonitors。

**上一批挂起项已闭环（callback 构造方式）**：
① PE 段表证明槽 0x141C5AB60 属 `.data` 的 **BSS 区**（已初始化区止于 0x141C0F800），
即 init 期写入 → 包级 `syscall.NewCallback` 形态；
② 二进制含 `syscall.compileCallback`(0x14007c660)，main_init 与 oledBlackoutEnumTopLevelWindows
均调用之，参照 win_enumTopLevelWindows.asm.txt 确认范式为 funcval → compileCallback → args → Call。
据此落地包级变量 `windowManagementEnumDisplayMonitorCallback`。

**重要验证经验**：`syscall.NewCallback` 的参数约束是**运行时**校验（init 期由 compileCallback
反射检查），非法签名会让**全部**测试失败，而 build/vet 均无法发现。本批靠 `go test` 通过证明
签名合法（4 实参皆 uintptr 宽度 + 单 uintptr 返回）。**结论：test 门禁不可省，它兜住 init 期崩溃类缺陷。**

**数据流闭环**：批次 258 的回调侧 `*rects = append(...)` 与本批调用侧
`uintptr(unsafe.Pointer(&rects))` 配对，构成完整「传入-填充-回读」链路。

**下一批**：targetFromWindowProcessPick 416B（大结构栈展开待核对）；MaybeWrapCursor 288B
（依赖 WrappedCursorPoint 992B）。FUNCS 2854/4754 = 60.04%。

### 批次 264（短函数 +3 [S]：滚动右键状态 + WebView2 检视辅助）

**基线/收口**：`FUNCS=2854→2857 / S=1322→1325 / S-inline=36 / S-sig=1456 / P=40 / UNMARKED=0`
（真函数 2814→2817 = 59.34%）。`go1.25.12 build/vet/test -tags production ./backend` 全 EXIT=0。

**本批落地（+3 [S]）**：
1. readScreenshotScrollingRightButtonState [S 0x1409a3320]：GetAsyncKeyState(VK_RBUTTON=2)，
   bit15=当前按下 + bit0=上次按过，返回 (bool,bool)。新 proc procGetAsyncKeyState。文件 screenshot_scroll_windows.go（新建）。
2. resolveWebView2ProcessInspectUserDataDir [S 0x1409df7e0]：TrimSpace → filepath.Abs，返回 (string,error)。
3. resolveWebView2ProcessInspectHostExeName [S 0x1409df840]：os.Executable → Base → TrimSpace，
   空/错回落默认名 "UsbEAm_Launcher.exe"（19B @0x140C5BD00，两分支同址）。文件 webview2_process_windows.go（新建）。

**proc 实证**：右键 proc 槽 0x141BC1CE8 = GetAsyncKeyState。

**本批后即转入整洁移交**：过程产物（pipeline/tmp + disasm*/dump_archive，共 4850 文件）移出版本控制、
CI 门禁统一 -tags production、push、远端 CI、交接文档。详见 STRUCTURE.md / 本文件「整洁与移交」节。

**下一批（新会话续接）**：targetFromWindowProcessPick 416B；MaybeWrapCursor 288B（依赖 WrappedCursorPoint 992B）。
FUNCS 2857/4754 = 60.10%。

### 批次 265（windowmanagement 光标环绕链 +3 [S]）

**基线/收口**：`FUNCS=2857→2860 / S=1325→1328 / S-inline=36 / S-sig=1456 / P=40 / UNMARKED=0`
（真函数 2817→2820 = 59.32%；FAITHFUL 1361→1364）。`go1.25.12 build/vet/test -tags production ./backend` 全 EXIT=0。

**本批落地（+3 [S]）**：
1. windowManagementWrappedCursorPoint [S 0x1409ed5c0]（992B）：`(x,y int32, monitors []RECT,
   wrapX,wrapY bool, guardPx int)(int32,int32,bool)`。扫描首命中 monitor → 有效性检查 →
   corner-guard → 水平 wrap（左边缘 WrapTargetX true=最右 best-3 / 右边缘 false=最左 best+3）→
   垂直 wrap（上边缘 WrapTargetY true=最下 / 下边缘 false=最上）。**关键：水平与垂直均基于原始
   (x,y)，相邻 monitor 命中跳过 wrap，ok=hOK|vOK。**
2. windowManagementMaybeWrapCursor [S 0x1409ed4a0]（288B）：`(wrapX,wrapY bool, monitors []RECT,
   guardPx int) bool`。GetCursorPoint 失败→false；WrappedCursorPoint ok=false→false；否则
   SetCursorPos(newX,newY) 无条件 true（返回值丢弃）。
3. targetFromWindowProcessPick [S 0x1409eea40]（416B）：`(path string, pid uint32, processName,
   displayName, title string, hwnd uintptr, iconData string) WindowManagementTarget`。
   duffzero 128B；Path/ProcessName/DisplayName 均 TrimSpace；ProcessName 空→filepath.Base(Path)；
   DisplayName 空→TrimSpace(title)→ProcessName 两级回退；**Title 保留原始值**；IconRef/IconURL 空串。

**新增 proc**：procSetCursorPos（user32DLL）。**新增 import**：path/filepath。
**proc 身份内存实证**：0x141BC1DC0 = SetCursorPos（`0x1409ed53d+0x11d4883` Python 显式计算，
与既有 proc 槽族 0x141BC1xxx 自洽）。
**黄金对拍**：新建 `windowmanagement_windows_test.go`（WrappedCursorPoint 10 例 +
targetFromWindowProcessPick 4 例，期望值取自 asm 显式路径 + 已落地 WrapTargetX/Y best∓3，
非凭空造值），全 PASS。

**下一批**：cursorWrapLoop / cursorWrapLoop.func1 / deferwrap1（0x1409ecde0 / 0x1409ed360 /
0x1409ed440，time.NewTicker 8ms + selectgo 两 case + mutex 临界区闭包）；gap_aggregate 剩余短函数。
FUNCS 2860/4754 = 60.16%。

### 批次 266（windowmanagement 光标环绕线程链 +3 [S] 升档）

**基线/收口**：`FUNCS=2860→2860 / S=1328→1331 / S-inline=36 / S-sig=1456→1453 / P=40 / UNMARKED=0`
（真函数 2820→2820 = 59.32%；FAITHFUL 1364→1367）。`go1.25.12 build/vet/test -tags production ./backend` 全 EXIT=0。

**本批落地（+3 [S]，均为 [S-sig] 升档，FUNCS 不变）**：
1. startCursorWrap [S 0x1409ecaa0]（480B）：`(horizontal, vertical bool)`。无 nil 检查；lock 后
   cursorActive 非 0 → 显式 Unlock return；makechan×2（chan struct{} 无缓冲）得 stop/done →
   s.cursorStop/s.cursorDone/cursorActive=true → Unlock → `go func(){ s.cursorWrapLoop(h,v,stop,done) }()`
   （newobject 5 捕获闭包 + newproc(gowrap1 0x1409ecc80)）。
2. stopCursorWrap [S 0x1409ecce0]（256B）：`()`。无 nil 检查、无 cursorActive 前置检查（幂等）；
   锁内读 stop/done 后 movups xmm15 清 cursorStop+cursorDone（0x100..0x108）+ byte[0x110]=0 →
   Unlock → stop!=nil 则 closechan；done!=nil 则 chanrecv1（`<-done` 阻塞至收尾）。
3. cursorWrapLoop [S 0x1409ecde0]（1248B）：`(horizontal, vertical bool, stop, done chan struct{})`。
   defer 链：close(done)（deferwrap1 0x1409ed440）→ func1（0x1409ed360，lock 内若 s.cursorStop==stop
   则清 cursorStop/cursorDone/cursorActive，竞态守卫）→ ticker.Stop（deferwrap2 0x1409ed300）。
   ticker=8ms（0x7a1200）；selectgo 双 recv case（ticker.C / stop，block=1）；锁内 duffcopy 拷 config
   （0xe0=224B）+ 读 cursorGuardPx(+0x118)/moduleEnabled(+0xf0)。判断链：!moduleEnabled→return；
   !wrapX&&!wrapY→return；getSystemMetrics(0x50=SM_CMONITORS)<=1→return；time.Since(lastWrap)<140ms
   （0x8583b00）→continue；time.Since(lastDisplayRefresh)>=500ms（0x1dcd6500）或 monitors 空→刷新
   DisplayRects+now；MaybeWrapCursor 真→lastWrap=now。**死参数 h/v**（morestack spill bl/cl 后从未
   读取，循环初始化 xor ebx/ecx 覆盖，函数体从 config 实时值取 wrapX/wrapY）。

**新增 import**：time。**proc 身份内存实证**：0x141BC1DB0 = GetSystemMetrics（0x1409ed084 读
`[rip+0x11d4d25]`，Python `hex(0x1409ed08b+0x11d4d25)=0x141BC1DB0`）。
**黄金测试**：`windowmanagement_windows_test.go` 追加生命周期 2 例（start/二次 start 短路/stop 清三态；
未启动 stop 幂等 no-op。moduleEnabled 默认 false → cursorWrapLoop 首次 tick 或 stop 关闭即 return，
不触达系统调用，确定性安全），全 PASS。

**下一批**：gap_aggregate.txt 剩余短函数落地。P=40。FUNCS 2860/4754 = 60.16%。

### 批次 267（workspacemigration 身份域 NOFOLLOW 文件打开链 +4 [S] 新增）

**基线/收口**：`FUNCS=2860→2864 / MARKED=2860→2864 / S=1331→1335 / S-inline=36 /
S-sig=1453 / P=40 / UNMARKED=0`（真函数 2820→2824 = 59.40%；FAITHFUL 1367→1371）。
`go1.25.12 build/vet/test -tags production ./backend` 全 EXIT=0。

**本批落地（+4 [S]，均为完全缺失→新增，非升档，FUNCS +4）**：
1. normalizeWorkspaceWindowsFinalPath [S 0x1409f4d00]（288B）：`(path string) string`。
   TrimSpace→ToLower；前缀 `\\?\unc\`(8B) → `\\`+trimmed[8:]；前缀 `\\?\`(4B) →
   trimmed[4:]；否则原样；最终 `internal/filepathlite.Clean`（0x1401154c0）。前缀判断大小写
   不敏感但拼接用原始 trimmed。语义：`\\?\C:\x`→`C:\x`，`\\?\UNC\srv\share`→`\\srv\share`。
2. openWorkspaceMigrationSourceFileNoFollow [S 0x1409f4e20]（96B）：`(path string)(*os.File,error)`。
   `ecx=0x80000000`(GENERIC_READ)、`edi=3`(OPEN_EXISTING) → regular，返回值透传。
3. openWorkspaceMigrationTargetFileNoFollow [S 0x1409f4e80]（96B）：`ecx=0x40000000`(GENERIC_WRITE)、
   `edi=1`(CREATE_NEW) → regular，返回值透传。
4. openWorkspaceMigrationRegularFileWindows [S 0x1409f4ee0]（480B）：`(path string, access uint32,
   disposition uint32)(*os.File,error)`。UTF16PtrFromString → CreateFile（sharemode =
   OPEN_EXISTING?7(READ|WRITE|DELETE):0、attrs=0x200080=FILE_ATTRIBUTE_NORMAL|
   FILE_FLAG_OPEN_REPARSE_POINT）→ GetFileInformationByHandle → `test attrs,0x410`
   (REPARSE_POINT|DIRECTORY) 命中 → CloseHandle + 错误 A；`cmp handle,-1` 守卫后
   `os.newFile(handle,path,0)`，nil → CloseHandle + 错误 B。

**错误类型身份实证（复用 extractError，非新建）**：错误对象 = `newobject(0x140b54380)`
（16B `struct{msg string}`，ptrdata=8）+ `[+0]=msg ptr`/`[+8]=msg len`；itab 0x1411d3100，
Error 方法（fun[0]=0x140090240）= `mov rcx,[rax]; mov rbx,[rax+8]; ret`。与
`openArchiveRegularFileNoFollow`（0x140750d20）错误构造**同址交叉验证**（newobject type
`0x140750dcc+0x4035b4=0x140B54380`、itab `0x140750dea+0xa82316=0x1411D3100`），故直接复用
backend 既有 `newExtractError`。字符串内存实证：错误 A `迁移文件最终句柄不是普通文件`
@0x140c81380（14 字符=42B=0x2a，asm 实写 `[rax+8]=0x2a`）；错误 B `无法包装迁移文件句柄`
@0x140c6fb0d（10 字符=30B=0x1e，实写 `[rax+8]=0x1e`）——len 字段与中文字符×3 精确吻合，
交叉确认这**不是 code 字段而是 string.len**。

**新增 import**：os / path/filepath / strings（+ 既有 x/sys/windows）。
**调用地址实证**：filepathlite.Clean 0x1401154c0、windows.UTF16PtrFromString 0x140194220、
windows.CreateFile 0x140195d20、windows.GetFileInformationByHandle 0x140197060、
windows.CloseHandle 0x140195860、os.newFile 0x140129ea0。
**黄金测试**：新建 `workspacemigration_identity_windows_test.go`，normalize 8 例（普通路径 /
设备前缀 / UNC 前缀保留 `\\` 头 / 大小写不敏感前缀×2 / 首尾空白 / 仅前缀 / 空串→`.`），
期望值取自 asm 显式路径（常量 0x5c636e755c3f5c5c / 0x5c3f5c5c / 0x140c3365d 解码），
非凭空造值，全 PASS。四函数此前 backend 内零调用点（grep 证实），无回归面。

**下一批**：`workspacemigration_identity.go` 剩余 3 函数（inspectWorkspaceMigrationPath 1696B /
validateWorkspaceMigrationExistingChain 320B / sameWorkspacePathInspection 384B）依赖本批
已落地的 normalize + NOFOLLOW 打开链，可成批落地。P=40。FUNCS 2864/4754 = 60.24%。

### 批次 268（workspaceMigrationPathIsReparse 签名订正 + validateWorkspaceMigrationExistingChain +1 [S]）

**基线/收口**：`FUNCS=2864→2865 / MARKED=2864→2865 / S=1335→1336 / S-inline=36 /
S-sig=1453 / P=40 / UNMARKED=0`（真函数 2824→2825 = 59.42%；FAITHFUL 1371→1372）。
`go1.25.12 build/vet/test -tags production ./backend` 全 EXIT=0。

**签名订正（关键）**：`workspaceMigrationPathIsReparse` 真实签名 `(path string) (bool, error)`，
batch 267 误落单 `bool`。完整 dump 0x1409f4660 四返回路径实证（UTF16 err / GetFileAttributes
err 两路 `xor eax,eax; ret` 返 `(false, err)`；成功 `bt eax,0xa; setb al` + `xor ebx,ebx;
xor ecx,ecx` 返 `(attrs&0x400!=0, nil)`），调用点 0x1409f42ea `test rbx,rbx` 检查 err 佐证。
已订正为 `(bool, error)`，err 分支不再吞错。

**本批落地（+1 [S]）**：
- validateWorkspaceMigrationExistingChain [S 0x1409f42a0]（320B）：`(path string) error`。
  循环 Clean(path) → os.Lstat err 直接返回 → workspaceMigrationPathIsReparse err 直接返回、
  true 则 `fmt.Errorf("迁移路径不能经过符号链接、目录联接或 reparse point: %s", cleaned)`
  （格式串 @0x140c90f6e，48B，`0x1409f436f+0x29cbff`）→ parent=Dir(cleaned)，parent==cleaned
  则 nil，否则 path=parent 续环。

**全仓整洁（本轮收尾动作）**：工作树干净（git status 空），13 个忽略项均为 pipeline/tmp、
disasm、__pycache__、backend.exe、artifacts 等合法临时产物，无死代码残留；修正 batch 267
遗留签名缺陷。

**下一批**：`inspectWorkspaceMigrationPath`（1696B）+ `sameWorkspacePathInspection`（384B）
依赖 `workspacePathInspection`（104B 布局，字段含 string×2 + bool×N，asm 已初探 A@0x80/
B@0xe8 间距 0x68）与 `workspaceMigrationIdentityForExisting`（1472B，尚未 dump）。**必须先落地
`workspaceMigrationIdentityForExisting` 定型结构，再三函数同批落地**（铁律①签名未定型拒绝落体）。
P=40。FUNCS 2865/4754 = 60.26%。

### 批次 269（workspacemigration_identity 域三函数收口：identityForExisting + inspect + same）

**基线/收口**：`FUNCS=2865→2868 / MARKED=2865→2868 / S=1336→1339 / S-inline=36 /
S-sig=1453 / P=40 / UNMARKED=0`（真函数 2825→2828 = 59.49%；FAITHFUL 1372→1375）。
`go build/vet/test -tags production ./backend` 全 EXIT=0（test `ok changeme/backend`）。

**结构体定型（本批关键前置，复用 types_workspace.go 既有声明，未新建）**：
`.eq.main.workspacePathIdentity`（0x140a0ace0）与 `.eq.main.workspacePathInspection`
（0x140a0ac20）逐字段实证，与 types_workspace.go 已有声明**布局一致**：
- `workspacePathIdentity`（40B）：canonical string@0x00(16B) + volume uint64@0x10 +
  file uint64@0x18 + valid bool@0x20。
- `workspacePathInspection`（104B）：canonical string@0x00(16B) + exists bool@0x10 +
  identity@0x18(40B) + ancestorIdentity@0x40(40B)。exists 后 Go 自动 pad 7B 对齐。

**本批落地（+3 [S]）**：
1. `workspaceMigrationIdentityForExisting` [S 0x1409f46e0]（0x5C0=1472B）：
   `(path string, _ os.FileInfo) (workspacePathIdentity, error)`。UTF16PtrFromString →
   CreateFile(access=0x80 FILE_READ_ATTRIBUTES, share=7 READ|WRITE|DELETE,
   disposition=OPEN_EXISTING, attrs=0x2200000=BACKUP_SEMANTICS|OPEN_REPARSE_POINT) →
   defer CloseHandle → GetFileInformationByHandle → `bt attrs,0xa`(REPARSE_POINT) 命中 →
   `newExtractError("迁移路径最终句柄指向 reparse point")`（@0x140c832b9，44B，`0x1409f4825+0x28ea94`）→
   GetFinalPathNameByHandle 循环（初始栈 buf 1024，`size<=n` 则 `make([]uint16,n+1)` 重试）→
   `normalizeWorkspaceWindowsFinalPath(string(utf16.Decode(buf[:n])))` →
   identity{canonical, volume=VolumeSerialNumber, file=FileIndexHigh<<32|FileIndexLow, valid=true}。
   **Filetime align=4 偏移实证**（x/sys ByHandleFileInformation）：VolumeSerialNumber@+0x1c、
   FileIndexHigh@+0x2c、FileIndexLow@+0x30（asm 读 0x850/0x860/0x864 = info 基址 0x834 加偏移）。
   **info 为死参数**：asm 序言不保存 rcx/rdi，直接被 `mov ecx,7`/`xor edi,edi` 覆盖，
   源码保留参数但未使用（`_ os.FileInfo`）。
2. `inspectWorkspaceMigrationPath` [S 0x1409f3c00]（0x6A0=1696B）：
   `(path string) (workspacePathInspection, error)`。Abs→Clean；循环 Lstat(current)：
   成功→break 得 info；`errors.Is(err, os.ErrNotExist)`→Dir(current)==current 则
   `fmt.Errorf("找不到迁移路径的现有父目录: %s", cleaned)`（@0x140c822af，43B，`0x1409f3ef7+0x28e3b8`），
   否则 append(Base(current)) 后 current=Dir(current)；其他 err 直接返回 →
   validateWorkspaceMigrationExistingChain(current) → workspaceMigrationIdentityForExisting(current, info) →
   从尾到头 Join 缺失段回 identity.canonical 后 Clean → exists=(缺失段数==0)；
   identity 仅 exists 时填，ancestorIdentity 恒填。
3. `sameWorkspacePathInspection` [S 0x1409f43e0]（0x180=384B）：
   `(a, b workspacePathInspection) bool`。exists 不等→false；exists 时比较 identity 的
   valid/volume/file；否则比较 ancestorIdentity 的 valid/volume/file 且 workspacePathsEqual(canonical)。

**错误类型复用实证**：reparse 错误 newobject type `0x140b54380` + itab `0x1411d3100`，
与 openWorkspaceMigrationRegularFileWindows 三处错误**同址**（`0x1409f4811+0x15fb6f` /
`0x1409f484a+0x7de8b6`），故复用 backend 既有 `newExtractError`（archiveextract.go），非新建类型。

**新增 import**：workspacemigration_identity.go 增 errors；workspacemigration_identity_windows.go
增 unicode/utf16。**调用地址实证**：os.Lstat 0x14012d9a0、errors.Is 0x1400904a0、
filepathlite.Dir/Base 0x140116440/0x140116320、path/filepath.join 0x1401cd660、
path/filepath.abs 0x1401cd5e0、windows.GetFinalPathNameByHandle 0x140197160（x/sys 别名，
zsyscall 0x140197160）、unicode/utf16.decode 0x1400e08e0、runtime.slicerunetostring 0x1400607c0。
**黄金测试**：本批无新增测试，既有 normalize 8 例仍 PASS，三函数后端内零调用点（grep 证实），
无回归面。

**下一批**：workspace migration 域剩余骨架升档 —— `validateWorkspaceMigrationRoots`
（0x1409efac0）、`workspacePathContains`（0x1409f0160）、`waitWorkspaceMigrationDrain`
（0x1409ef500）、`removeOwnedWorkspaceMigrationDirectory`（0x1409f34c0）四函数未落地 +
`stageWorkspaceDataMigration`（0x1409f0240）/`stagedWorkspaceData.Commit`（0x1409f2460）/
`stagedWorkspaceData.Rollback`（0x1409f3100）三 [S-sig] 骨架待体。P=40。FUNCS 2868/4754 = 60.33%。

### 批次 270（workspacemigration 域 12 函数落地：gate/排空/校验/暂存/提交/回滚）

**基线/收口**：`FUNCS=2868→2871 / MARKED=2868→2871 / S=1339→1349 / S-inline=36 /
S-sig=1453→1446 / P=40 / UNMARKED=0`（真函数 2828→2831 = 59.55%；FAITHFUL 1375→1385）。
`go build/vet/test -tags production ./backend` 全 EXIT=0（test `ok changeme/backend`）。

**落地（+12 [S]，-7 [S-sig]，+3 FUNCS）**：
1. `workspaceDataMaintenanceGate.begin` [S 0x1409ef000]：`(func(), error)`，nil→noop；
   maintenance→全局 error；active==0→make idle；active++；返回 `once.Do(func2.1)` 退出闭包
   （func2.1 0x1409ef1e0 = 递减 active / 归零 close(idle) 并清空）。
2. `workspaceDataMaintenanceGate.enter` [S 0x1409ef2a0]：`(chan struct{}, func(), error)`，
   nil→(closed chan,noop,nil)；maintenance→(nil,nil,全局 error)；maintenance=true；
   idleChan = active==0 ? make+close : g.idle（不回写 g.idle）；返回 `once.Do(func2.1)`。
3. `waitWorkspaceMigrationDrain` [S 0x1409ef500]：idle==nil→nil；select ctx.Done→
   `errors.Join(errWorkspaceDrainTimeout, ctx.Err())` / idle→nil。
4. `beginWorkspaceDataOperation` [S 0x1409ef7c0]：bs==nil→"工作区服务不可用"；否则
   `(&bs.workspaceDataMaintenance).begin()` 透明透传 `(func(), error)`。
5. `beginWorkspaceMigrationMaintenance` [S 0x1409ef840]（0 参）：bs==nil→"工作区服务不可用"；
   app!=nil→`errWorkspaceAppActive`；enter→err 透传；WithTimeout(15s)→waitDrain→超时 exitFn()
   后返 err；成功返 exitFn。
6. `validateWorkspaceMigrationRoots` [S 0x1409efac0]：TrimSpace 空/Abs/Clean/workspacePathsEqual/
   workspacePathContains 双向/inspect 源/目标/最终路径三向重叠/同一文件对象 六段校验。
7. `workspacePathContains` [S 0x1409f0160]：Rel err/`.`/`..`/`..\` 前缀→false；否则 !IsAbs。
8. `stageWorkspaceDataMigration` [S 0x1409f0240]：validate→Stat 源→inspect→目标空校验→
   MkdirAll 父级→rand 命名 staging→Mkdir 0700→卷一致性→Walk 拷贝→verify(staging,false)→
   verify(source,true)→身份复核→构造 staged。
9. `verifyWorkspaceDataManifestWithPolicy` [S 0x1409f1ba0] + func1 [S 0x1409f1da0]：
   visited map→Walk→清单完整性；func1 walkErr 在前、strict skip、isReparse、Mode&ModeType、
   relKey ReplaceAll、sha256 校验。
10. `stagedWorkspaceData.Commit` [S 0x1409f2460]：verify×2→身份复核×4→Stat/ReadDir/Remove
    目标→Rename→committed 标志→提交后身份复核。
11. `stagedWorkspaceData.Rollback` [S 0x1409f3100]：!committed→removeOwned(staging)；
    否则身份复核→verify→RemoveAll→targetWasEmpty 恢复。
12. `removeOwnedWorkspaceMigrationDirectory` [S 0x1409f34c0]：inspect→身份匹配才 RemoveAll。

**订正**：`prepareWorkspaceDirectories` [S-sig]→[S]（6 目录 `{Root,IconDir,IndexDir,
ScreenshotDir,WebView2Dir,BackgroundDir}` + 逆序清理闭包 func1 0x1409f3b80）；
`openWorkspaceMigrationTargetFileNoFollow` 签名补 perm 参数（asm 0x1409f4e93 覆盖为
GENERIC_WRITE 忽略，源码 `_ os.FileMode`）；`MigrateConfig` 两调用点（begin 0 参、
buildLayout 用 staged.targetPath）+ 7 处 `defer op()/gate()`；删除独立 `End` 方法与
bootstrapservice_migration.go 6 骨架、screenshot_stubs.go 3 骨架。

**错误类型实证**：3 个包级 error 接口（errWorkspaceMigrationInProgress 0x141BC3E00 /
errWorkspaceAppActive 0x141BC3E10 / errWorkspaceDrainTimeout 0x141BC4050，均 extractError
itab 0x1411d3100）；"工作区服务不可用" 为每次 newobject（24B @0x140C65700）。

**下一批**：workspace migration 域剩余骨架 `buildWorkspaceLayoutWithConfig`（0x1407a1ac0）
与 `importLauncherBackgroundImage`（0x140798e40）[S-sig] 升档，或转入其他专项域。
P=40。FUNCS 2871/4754 = 60.39%。

### 批次 271（buildWorkspaceLayoutWithConfig + importLauncherBackgroundImage 升档 + 2 helper）

**基线/收口**：`FUNCS=2871→2873 / MARKED=2871→2873 / S=1349→1353 / S-inline=36 /
S-sig=1446→1444 / P=40 / UNMARKED=0`（真函数 2831→2833 = 59.59%；FAITHFUL 1385→1389）。
`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

**落地（+2 [S-sig]→[S]，+2 新 [S]，-2 [S-sig]，+2 FUNCS）**：
1. `buildWorkspaceLayoutWithConfig` [S 0x1407a1ac0, 1696B]：签名订正为 8×string
   （root/configFile/pluginDir + Storage 五字段，前 3 组走寄存器后 5 组走栈）；
   rootEff=TrimSpace(root) 空则 "."，normalizeStorageConfig 后 DataRoot 非空覆盖 rootEff；
   ConfigFile 空则 `resolveLauncherConfigFilePath(resolveProcessWorkingDirectory())`；
   LanguageDir=AppLanguageDir=`filepath.Join(filepath.Dir(pluginDir),"language")`，PluginDir 原样；
   Icon/Index/Screenshot/WebView2 各自非空优先，空则 `filepath.Join(rootEff, 单数子目录)`；
   BackgroundDir=`filepath.Join(rootEff,"background")`。
   子目录名 rodata 解码（**单数**，与 resolveWorkspaceLayout 的复数不同）：
   icon/index/screenshot/webview2/background/language。
2. `importLauncherBackgroundImage` [S 0x140798e40, 1632B]：签名订正为
   `(sourcePath string)(string,error)`（ws 参数删除，改由 `bs.workspaceSnapshot()` 取 BackgroundDir，
   ChooseLauncherBackgroundImage.asm 0x14077b3c9 只传 source 单参交叉验证）；
   流程 TrimSpace→Abs→Stat→IsDir 判目录→`normalizeLauncherBackgroundImageExtension(filepath.Ext(abs))`
   →MkdirAll(BackgroundDir,0o755)→`Join(BackgroundDir,"custom-background"+ext)`→
   `copyLauncherBackgroundImageFile`→装配 BackgroundPreference（Enabled/ReadabilityOverlayEnabled=true、
   ReadabilityOverlayOpacity=0.18、ImageOpacity=1.0）→`normalizeBackgroundPreference`→
   `attachLauncherBackgroundURL`→返回 `(filepath.Base(abs), nil)`。
   错误消息（rodata 解码）："背景图片路径不能为空"(30B) / "背景图片不能是目录"(27B) /
   "不支持的背景图片格式"(30B)。
3. `normalizeLauncherBackgroundImageExtension` [S 0x1407997a0, 256B]（新）：TrimSpace+ToLower 分派，
   .bmp/.gif/.jpg/.png/.webp 原样，.jpeg→.jpg，其余空串（立即数 0x706d622e/0x6669672e/
   0x67706a2e/0x676e702e/0x65706a2e(+'g')/0x6265772e(+'p') 实证）。
4. `copyLauncherBackgroundImageFile` [S 0x1407994a0, 672B]（新）：`filepath.Clean(src)==Clean(dst)`
   短路 nil；OpenFile(src,O_RDONLY,0) defer Close；OpenFile(dst,O_WRONLY|O_CREATE|O_TRUNC,0o644)；
   `io.Copy` 出错关 dst 返错，成功返 `dst.Close()`。

**订正**：`resolveLauncherConfigFilePath` 签名 `(WorkspaceLayout)string`→`(string)string`
（buildWorkspaceLayoutWithConfig.asm 0x1407a1b93 调用点只传单 string，且为
resolveProcessWorkingDirectory 返回值）；`buildWorkspaceLayoutWithConfig` 两调用点改 8 参
（commit...WithWidgetsImpl 0x140778f40 用 ws.Root/ws.ConfigFile/ws.PluginDir + cfg.Storage 五字段；
MigrateConfig 0x14077c8d6 用 staged.targetPath + ws.ConfigFile/ws.PluginDir + 空五字段，
并补接 `configStoreSnapshot` 第二返回值 ws）。

**下一批**：workspace migration 域已闭环（buildWorkspaceLayoutWithConfig 是迁移链最后一环，
批次 270 的 12 函数 + 本批 2 函数全 [S]）。可转入 launcher asset 背景域
（`attachLauncherBackgroundURL` 0x140798600 [S-sig] 387 行 asm、
`launcherBackgroundContentTypeForPath` 0x1407998a0 256B）或其它专项域。
P=40。FUNCS 2873/4754 = 60.43%。

### 批次 272（attachLauncherBackgroundURL 升档 + launcherBackgroundContentTypeForPath 新增 + launcherasset 签名订正）

**基线/收口**：`FUNCS=2873→2874 / MARKED=2873→2874 / S=1353→1355 / S-inline=36 /
S-sig=1444→1443 / P=40 / UNMARKED=0`（真函数 2833→2834 = 59.61%；FAITHFUL 1389→1391）。
`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

**落地（+1 [S-sig]→[S]，+1 新 [S]，-1 [S-sig]，+1 FUNCS）**：
1. `attachLauncherBackgroundURL` [S 0x140798600, 387L]：bs/cfg nil 守卫 →
   normalizeBackgroundPreference → TrimSpace(ImagePath) 空则写回返回 → screenshotAssetService()
   nil 返回 → Clean(TrimSpace(ImagePath)) → os.Stat err/IsDir/Size≤0 返回 → ModTime().UnixNano()
   → backgroundAssetLock 缓存命中（owner/path/size/modifiedAt 相等 + Exists(URL,
   "background/custom")）复用 URL 返回 → 否则 RegisterFile("background/custom", cleaned,
   contentTypeForPath(cleaned), 0) → err==nil 更新缓存五字段 + ImageURL → 写回 normalized。
2. `launcherBackgroundContentTypeForPath` [S 0x1407998a0, 352B]（新）：filepath.Ext → ToLower →
   立即数分派 .bmp/.gif/.png → image/{bmp,gif,png}、.jpg/.jpeg → image/jpeg、.webp → image/webp、
   其余 application/octet-stream。

**订正（launcherasset 核心签名，汇编实证，均已是 [S] 无档位迁移）**：
- `RegisterFile` 0x14086ed00：`(namespace,id,path,version)` → `(namespace,path,contentType,ttl)`。
  开 path（第2参）、contentType 派生（mime.TypeByExtension 回落 octet-stream）、ttl 直传；
  **不预读文件**，filePath 存 entry（ReadBytes 时 readFileBounded 回读）。
- `RegisterBytes` 0x14086e220：`(namespace,id,data,version)` → `(namespace,contentType,data,ttl)`。
  contentType 派生（http.DetectContentType 回落 octet-stream）、mallocgc+memmove 复制 data。
- `register` 0x14086f680：`(namespace,id,data,version)` →
  `(namespace,contentType,data,filePath,size,ttl)`。id 恒 newLauncherAssetID 查重生成、
  version 恒 next、expiresAt=now+ttl（ttl==0→600s、ttl<0→零值永不过期）。
- 截图域三调用点（attachScreenshotAssetURL/attachScreenshotCaptureAssetURL/
  attachScreenshotThumbnailAssetURL）：namespace "screenshot/current"（0x140c59c66 18B）、
  第2参 = trimmed 全路径（非 basename）、第3参 = screenshotContentTypeForPath、第4参 = 600s。

**错误消息实证**（rodata 解码）："读取资源文件失败: %w"(28B) / "资源文件不能是目录"(27B) /
"资源文件不能为空"(24B) / "资源内容不能为空"(24B) / "资源命名空间不能为空"(30B) /
"资源服务尚未初始化"(27B)。

**下一批**：`attachScreenshotThumbnailAssetURL` 汇编实走 buildScreenshotThumbnailPNGFromData/
FromPath（非 RegisterFile），属截图域后续专项。launcher asset 背景域已闭环。P=40。
FUNCS 2874/4754 = 60.45%。

### 批次 273（webview2_process.go 整文件落地 +6 [S] +1 [P] 存根）

**基线/收口**：`FUNCS=2874→2881 / MARKED=2874→2881 / S=1355→1361 / S-inline=36 /
S-sig=1443 / P=40→41 / UNMARKED=0`（真函数 2834→2840 = 59.74%；FAITHFUL 1391→1397）。
`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

**落地（+6 [S]，+1 [P] 存根，+7 FUNCS）**：
1. `webView2ProcessKindSortWeight` [S 0x1409deac0, 256B]：TrimSpace+ToLower，browser=0/renderer=1/
   gpu-process=2/utility=3/其余=4。
2. `extractWebView2ProcessType` [S 0x1409de960, 352B]：ToLower→Index("--type=")，未命中 "browser"；
   命中取 token 后 TrimSpace+Trim(`"`)。
3. `normalizeWebView2ProcessKind` [S 0x1409de840, 288B]：双参 (kind,commandLine)，空串 extract，
   ""/"browser" 归一 "browser"。
4. `normalizeWebView2ProcessInfo` [S 0x1409de500, 832B]：kind 归一 + MB 换算（÷1024²）+
   三 bool（HasJSFlags/HasDisableFeatures/HasRendererProcessLimit）已置位保持否则 Contains。
5. `normalizeWebView2ProcessSnapshot` [S 0x1409dde00, 1472B]：nil 置空 + 逐元素归一累计 +
   sort.SliceStable（权重升序，ProcessID 平局升序）+ 总量 MB。
6. `InspectWebView2Processes` [S 0x1409ddc60, 416B]：nil 守卫→workspaceSnapshot().WebView2Dir
   （偏移 +0x80）→inspectWebView2Processes。
7. `inspectWebView2Processes` [P 0x1409debc0, 3104B]：平台层进程枚举存根（签名已实证，体待专项）。

**常量实证**："--type="（7B @0x140C395A0）、"browser"（7B @0x140C39599）、
`"`（1B cutset @0x1411CAC88）、"--js-flags="(11B @0x140C47699)、
"--disable-features="(19B @0x140C5BCED)、"--renderer-process-limit="(25B @0x140C66E11)、
0.0009765625（1/1024 @0x1411CD608，MB 换算连乘两次）。

**下一批**：`inspectWebView2Processes`（3104B）平台层枚举专项（CreateToolhelp32Snapshot/
OpenProcess/读命令行链），或转其它未落地文件差集（43 个）。P=41。FUNCS 2881/4754 = 60.60%。
未落地文件差集 44→43。

### 批次 274（screenshot_pin_windows.go 整文件落地 +2 [S]）

**基线/收口**：`FUNCS=2881→2883 / MARKED=2881→2883 / S=1361→1363 / S-inline=36 /
S-sig=1443 / P=41 / UNMARKED=0`（真函数 2840→2842 = 59.78%；FAITHFUL 1397→1399）。
`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

**落地（+2 [S]，+2 FUNCS）**：
1. `screenshotPinSetLayeredWindowOpacity` [S 0x1409944a0, 256B]：`(hwnd uintptr, opacity float64)`
   void；hwnd==0 返回；clamp 到 [0.2, 1.0]（0.2 @0x1411CD658、1.0 @0x1411CD6D0）；
   opacity<1.0 → alpha=max(1, int(opacity*255))，否则 255；SetLayeredWindowAttributes
   (hwnd, 0, alpha, LWA_ALPHA=2)（proc @0x141BC1CB0）。
2. `screenshotPinSuppressWindowBorder` [S 0x1409945a0, 160B]：`(hwnd uintptr)` void；hwnd==0 返回；
   DwmSetWindowAttribute(hwnd, DWMWA_BORDER_COLOR=0x22, &DWMWA_COLOR_NONE(0xfffffffe), 4)
   （proc @0x141BC2658）。

**新增全局**：`dwmapiDLL` + `procDwmSetWindowAttribute` + `procSetLayeredWindowAttributes`
（user32DLL 复用 appicon_windows.go）。

**下一批**：`inspectWebView2Processes`（3104B）平台层枚举专项，或转其它未落地文件差集（42 个）。
P=41 持平。FUNCS 2883/4754 = 60.64%。未落地文件差集 43→42。

### 批次 275（desktopwidgets_notification_windows.go 整文件落地 +2 [S]）

**基线/收口**：`FUNCS=2883→2885 / MARKED=2883→2885 / S=1363→1365 / S-inline=36 /
S-sig=1443 / P=41 / UNMARKED=0`（真函数 2842→2844 = 59.82%；FAITHFUL 1399→1401）。
`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

**落地（+2 [S]，+2 FUNCS）**：
1. `normalizeDesktopWidgetNotificationText` [S 0x1407aca60, 288B]：`(text string, maxLen int,
   fallback string) string`；TrimSpace→Map(控制字符折叠空格，闭包 0x1409f7ba0)→Fields→
   Join(" " @0x1411CA5C8)→[]rune 截断到 maxLen→空串回退 fallback。
2. `showDesktopWidgetNotification` [S 0x1407ac960, 256B]：`(title, body string, silent bool)`；
   title→normalize(120,"UsbEAm Launcher")、body→normalize(240,"提醒时间已到")、
   silent?sound="silent":"ms-winsoundevent:Notification.Default"、showLauncherNotificationWithAudio。

**依赖订正**：`showLauncherNotificationWithAudio`（0x1408cfd20）序言 6 寄存器 = 3×string，
签名订正 `(title, message, sound string) error`（原二参存根为错，仍 [S-sig] 骨架）。

**下一批**：`inspectWebView2Processes`（3104B）平台层枚举专项，或转其它未落地文件差集（41 个）。
P=41 持平。FUNCS 2885/4754 = 60.69%。未落地文件差集 42→41。

### 批次 276（desktopwidgets_secrets_windows.go 整文件落地 +2 [S]）

**基线/收口**：`FUNCS=2885→2887 / MARKED=2885→2887 / S=1365→1367 / S-inline=36 /
S-sig=1443 / P=41 / UNMARKED=0`（真函数 2844→2846 = 59.87%；FAITHFUL 1401→1403）。
`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`，含 DPAPI roundtrip）。

**落地（+2 [S]，+2 FUNCS）**：
1. `protectDesktopWidgetSecret` [S 0x1407b1660, 960B]：`(secret string) (string, error)`；
   空串→"凭据不能为空"（18B @0x140C59CD2）；CryptProtectData(flags=CRYPTPROTECT_UI_FORBIDDEN=1)
   → LocalFree → 复制 → base64.StdEncoding；defer clear/LocalFree/clear。
2. `unprotectDesktopWidgetSecret` [S 0x1407b1b40, 1024B]：`(encrypted string) (string, error)`；
   DecodeString err||len==0→"受保护凭据格式无效"（27B @0x140C69C84）；CryptUnprotectData(flags=1)
   → LocalFree → 复制 → string；defer 逆序 clear/LocalFree/clear。

**实证**：DataBlob{Size,Data} 布局、CryptProtectData/CryptUnprotectData 双调用 @0x1401953E0/
0x1401955A0、LocalFree @0x140198400、clear=memclrNoHeapPointers（deferwrap 0x1407B1AE0/1A20）、
CryptUnprotectData name 参数 **uint16。

**下一批**：`inspectWebView2Processes`（3104B）平台层枚举专项，或转其它未落地文件差集（40 个）。
P=41 持平。FUNCS 2887/4754 = 60.73%。未落地文件差集 41→40。

### 批次 277（desktopwidgets_configio.go 整文件落地 +4 [S]）

**基线/收口**：`FUNCS=2887→2891 / MARKED=2887→2891 / S=1367→1371 / S-inline=36 /
S-sig=1443 / P=41 / UNMARKED=0`（真函数 2846→2850 = 59.95%；FAITHFUL 1403→1407）。
`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

**落地（+4 [S]，+4 FUNCS）**：
1. `extractDesktopWidgetDocument` [S 0x1407ab8c0, 1344B]：`(data []byte)(DesktopWidgetDocument,bool,error)`；
   Unmarshal 失败→err；无"desktopWidgets"键或 TrimSpace 空或=="null"→(零值,false,nil)；
   Decoder.Decode 失败→desktopWidgetStoreError("desktopWidgets","解析失败: %v")；EOF 失败→
   desktopWidgetStoreError("desktopWidgets",err.Error())；normalize+validate 失败→err；成功→(doc,true,nil)。
2. `launcherConfigDocumentIsWidgetLibrary` [S 0x1407abe00, 768B]：`(data []byte) bool`；
   需含"widgets"且不含 15 白名单键（initialized/storage/preferences/apps/speedDial/bookmarks/
   fileSearch/fileLocator/files/memoryRelease/oledBlackout/windowManagement/mouseGestures/twoFactor/
   desktopWidgets）。
3. `portableDesktopWidgetDocument` [S 0x1407ac4a0, 384B]：clone 后 3×makemap_small 清空
   WeatherCache/NotificationDeliveries/ProtectedSecrets（偏移 +0x50/+0x58/+0x60）。
4. `desktopWidgetDocumentForImport` [S 0x1407ac620, 448B]：portable 后从 clone 恢复
   ProtectedSecrets（+0x60），其余设备字段仍清空。

**下一批**：`inspectWebView2Processes`（3104B）平台层枚举专项，或转其它未落地文件差集（39 个）。
P=41 持平。FUNCS 2891/4754 = 60.81%。未落地文件差集 40→39。

### 批次 278（desktopwidgets_notification.go 整文件落地 + Notify 方法 +2 [S]）

**基线/收口**：`FUNCS=2891→2893 / MARKED=2891→2893 / S=1371→1373 / S-inline=36 /
S-sig=1443 / P=41 / UNMARKED=0`（真函数 2850→2852 = 59.99%；FAITHFUL 1407→1409）。
`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

**落地（+2 [S]，+2 FUNCS）**：
1. `platformDesktopWidgetNotifierNotify` [S VA=0]：`(title, body string, silent bool)` 空实现，
   Windows 构建被编译期死代码消除（Notify 直连 showDesktopWidgetNotification 不经由本函数）。
2. `(platformDesktopWidgetNotifier) Notify` [S 0x1407ac8e0, 34B]：`(title, message string,
   silent bool) error`；rax/rbx=title、rcx/rdi=message、rsi=silent 透传 showDesktopWidgetNotification
   （call 0x1407ac960）后返回 nil；闭合 deliverNotification → notifier.Notify 调用链。

**下一批**：`inspectWebView2Processes`（3104B）平台层枚举专项，或转其它未落地文件差集（38 个）。
P=41 持平。FUNCS 2893/4754 = 60.85%。未落地文件差集 39→38。

### 批次 279（launchericonasset.go gap 差集闭合 +4 [S] +1 [S-inline]）

**基线/收口**：`FUNCS=2893→2898 / MARKED=2893→2898 / S=1373→1377 / S-inline=36→37 /
S-sig=1443 / P=41 / UNMARKED=0`（真函数 2852→2857 = 60.10%；FAITHFUL 1409→1414）。
`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

**落地（+4 [S]，+1 [S-inline]，+5 FUNCS）**：
1. `ResolveAppIconResource` [S 0x1408a17e0, 320B]：`(path string) LauncherIconResource`；
   默认 opts（rodata 0x141be9500）→ `resolveAppIconDataWithOptions` →
   `buildLauncherIconResource("icon/app", iconData)`。
2. `attachWindowManagementTargetIconURL` [S 0x1408a45c0, 800B]：normalize →
   `buildLauncherConfigIconResource("icon/window-management", IconRef, IconData)` →
   写回 IconRef/IconURL + 清零 IconData；回退（!HasSource && IconURL=="" && TrimSpace(Path)!=""）
   → `resolveAppIconDataWithOptions` → `buildLauncherIconResource` → IconURL。
3. `attachWindowManagementStateIconURLs` [S 0x1408a48e0, 416B]：两次 attachTarget
   （Config.Target @+0x28、Target @+0x100）。
4. `persistAndRefreshWindowManagementState` [S 0x1408a4a80, 1184B]：priorErr 短路 →
   `persistWindowManagementConfig(state.Config)` → `windowManagement.GetState()` → 分支 attachState。
5. `defaultAppIconOptions` [S-inline]：rodata 0x141be9500 内联复制的 DRY 提取
   （Size=256、ImageList=4、CandsPtr=&defaultAppIconSizes[0]、CandsLen/CandsCap=5）。

**连带**：`AppIconOptions.CandsCap` padding→int（+0x38，ABI 64B 不变）；
`persistWindowManagementConfig` 签名补全 cfg 入参（persist.asm duffcopy state.Config 入栈确证）。

**下一批**：`inspectWebView2Processes`（3104B）平台层枚举专项，或转其它未落地文件差集（37 个）。
P=41 持平。FUNCS 2898/4754 = 60.96%。未落地文件差集 38→37。






