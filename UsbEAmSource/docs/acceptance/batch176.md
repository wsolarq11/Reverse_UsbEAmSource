# 批次 176 验收（多文件 [P]→[S-sig] 签名实证批量·56 函数）

日期：2026-09-25
子批次：签名实证批量（Lead + 5 个并行 subagent，跨 9 个文件）

## 目标

把 9 个文件的 `[P]` 存根做**逐寄存器签名实证**（morestack 序言数参数宽度 + 函数体判类型），
签名可确证的升 `[S-sig]`，不可确证的保持 `[P]` 并订正阻断注释。全程不臆造签名。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go1.25.12 test -count=1 -p=1 ./backend` | ok changeme/backend 0.402s |

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2659 / S=1104 / S-inline=35 / S-sig=1405 / P=115 / UNMARKED=0`。

**真函数 = 1104 + 35 + 1405 = 2544 / 4754 = 53.5%**（相对批次 175 真函数 2488 **增 +56**；
P 171→**115**（-56），S-sig 1349→**1405**（+56））。

重建产物 SHA256：`06B551688111458EE66DC1087ECD5B1BA331CA9BFEC8181F2AA7BC94BA6B7EF4`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细（56 函数，跨 9 文件）

### Lead 直接实证（9 个）

| 文件 | 函数 | VA | 关键订正 |
|---|---|---|---|
| oledblackout_windows.go | oledBlackoutApplyWindowChrome | 0x140914580 | 1 接口参数 + InvokeSync 闭包，itab fun[0x130] |
| oledblackout_windows.go | oledBlackoutFocusWindow | 0x140914860 | 尾调 FocusWindowNow |
| oledblackout_windows.go | oledBlackoutFocusWindowNow | 0x1409148c0 | itab fun[0x130] |
| oledblackout_overlay_windows.go | ensureOLEDBlackoutNativeOverlayWindowClass | 0x140911440 | () → (uintptr, uintptr)，sync.Once.doSlow |
| filesearch_windows.go | logFileSearchPerf | 0x14081f600 | 8 参误作 5 参，订正 (kind,query string, count int) |
| mousegestures_actions_windows.go | mouseGestureWindowMovePosition | 0x1408e7c40 | int→int32（dword 运算证 32 位） |
| nativedrag_windows.go | sanitizeDirectoryShortcutName | 0x1408fb580 | (name string) → (name, path string) |
| audio_windows_runtime.go | (*audioIPropertyStore).GetValue | 0x140759a20 | (key,pv) error → (key) (audioPropVariant, error) |
| audio_windows_runtime.go | (*audioIAudioPolicyConfigFactory).SetPersistedDefaultAudioEndpoint | 0x14075ab20 | (policy) → (processID,flow,role,deviceID) |

### oledblackout_windows.go（7 个，subagent）

`oledBlackoutMediaPauseProcessExcluded`(0x14091c720)、`oledBlackoutStandaloneBrowserCandidateAtCursor`(0x1409152e0)、
`oledBlackoutCountVisiblePlayingCoarseAppWindows`(0x1409195e0)、`oledBlackoutRectHasVisibleAreaAfterOcclusion`(0x14091bb40)、
`oledBlackoutWindowTargetVisibleRects`(0x14091b8c0)、`oledBlackoutNewHString`(0x14091fc80)、`oledBlackoutHStringToString`(0x14091fe40)。

### screenshot_windows.go（11 个，subagent）

`pickWindowProcessFromScreenshotSelection`(0x1409ba720)、`run`(0x1409bb300)、`paint`(0x1409bd900)、
`resolveWindowAtPoint`(0x1409be420)、`drawHoverProcessInfo`(0x1409befa0)、`createScreenshotInfoFont`(0x1409bfae0)、
`loadBestAppIconForOverlay`(0x1409bfe40)、`buildWindowProcessPickResult`(0x1409c0120)、
`buildWindowProcessPickResultWithProcessName`(0x1409c02a0)、`uniqueNonEmptyScreenshotRects`(0x1409c0a00)、
`finalizeWindowAtPoint`(0x1409c0da0)。

### qrcode_windows.go（6 个，subagent）

`handleHDRPreviewUpgrade`(0x1409397c0)、`updateSelectionResizeHover`(0x14093ba20)、
`beginCornerRadiusDrag`(0x14093bba0)、`resolveWindowHoverForWindow`(0x14093df20)、
`annotationToolbarState`(0x140940e40)、`drawAnnotationStrokesOnPaintBuffer`(0x1409498c0)。

### launcherupdate_runtime.go（13 个，subagent）

`InstallLauncherUpdate`(0x1408b3560)、`CancelLauncherUpdate`(0x1408b4120)、
`fetchLauncherRemoteConfigWithContext`(0x1408b4440)、`attachLauncherAdvertisementImageAsset`(0x1408b4d40)、
`parseLauncherRemoteConfigText`(0x1408b5320)、`normalizeLauncherUpdatePackageConfig`(0x1408b65a0)、
`normalizeLauncherAdvertisementConfig`(0x1408b67c0)、`setLauncherUpdateProgressForTask`(0x1408b6d00)、
`setLauncherUpdateProgressState`(0x1408b6e00)、`failLauncherUpdateProgressForTask`(0x1408b7200)、
`hashLauncherUpdatePackageFile`(0x1408bc180)、`locateLauncherUpdateContentRoot`(0x1408bc6a0)、
`validateLauncherUpdateContentRoot`(0x1408bcb80)。

### screenshot_uia_windows.go（10 个，subagent）

`screenshotMSAAControlBoundsAtPointWithPriority`(0x1409af8c0)、`screenshotMSAAElementBoundsAtPointWithQueryTimeoutAndPriority`(0x1409af9e0)、
`screenshotMSAAElementBoundsAtPointOnCOMThread`(0x1409afb20)、`screenshotMSAAElementBoundsAtPointFromWindow`(0x1409afba0)、
`screenshotMSAAElementBoundsAtPointGlobal`(0x1409affa0)、`screenshotUIAElementBoundsAtPointWithQueryTimeoutAndPriority`(0x1409b1000)、
`screenshotUIAElementBoundsAtPointOnCOMThread`(0x1409b1140)、`screenshotUIAElementBoundsAtPointWithAutomation`(0x1409b1320)、
`screenshotUIARawViewWalker`(0x1409b1b40)、`readScreenshotUIAElementHitInfo`(0x1409b1ca0)。

## G4 关键实证结论

1. **参数宽度以 morestack 保存为准**：多处旧签名参数个数/顺序错误，如 `logFileSearchPerf` 旧 8 参实为 5 参、
   `sanitizeDirectoryShortcutName` 漏 `path` 参数、`SetPersistedDefaultAudioEndpoint` 旧单参实为 4 参、
   `screenshotUIAElementBoundsAtPointWithAutomation` 参数顺序反了。均以序言寄存器实证订正。
2. **int32 判别**：`mouseGestureWindowMovePosition` 用 dword 运算（sub ecx,eax）证 32 位，旧 `int` 订正为 `int32`。
3. **返回值结构经栈返回**：`GetValue` 的 PROPVARIANT 24B 经栈返回区写回，旧签名把输出误作入参 `pv`。
4. **不臆造**：filelocator_runtime.go 22 个、以及各域 COM/WinRT/匿名结构体/多寄存器栈混合展开的存根，
   因类型无法唯一确证保持 `[P]` 并订正注释（P 存根仍含约 23 个 screenshot_uia/pin 域 + 22 个 filelocator 域等）。

## 残留 / 未落地（不触及）

- `[P]` 存根余 115，主要阻断三类：① WinRT/COM 类型未落地（oledblackout_windows.go 27 个 GSMTC/IAsyncOperation、
  screenshot_uia 深层 11 个）；② 参数多寄存器+栈混合展开无法唯一拆分（filelocator_runtime.go 22 个、launcherupdate
  下载族 10 个）；③ 匿名上下文/回调结构体（oledblackout_windows.go 若干）。留待各专项批次。
- 未落地原始文件差集 100 个（§10），本批次不触及。
