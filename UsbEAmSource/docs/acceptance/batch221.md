# 批次 221 — screenshot_windows 会话方法四 [P] 升档 [S-sig]

## 基线 / 收口

| 指标 | 基线（批次 220 收口） | 收口（批次 221） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1413 | **1417** |
| P | 83 | **79** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2705 | **2709**（56.98%） |

SHA256 `FEBC4D6E31A93AF90C050DB3161E858C118C31624EF6802A3F6533CFDBA77D55`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/screenshot_windows.go` 四个 `screenshotWindowSelectionSession` 方法 [P] 升档 [S-sig]，
均为「形参未逐寄存器实证」类阻断，逐寄存器对齐后参数分组唯一。

### invalidateWindowChange（0x1409bdda0）

`()` → `(hwnd uintptr)`。rbx 透传 `w32.InvalidateRect(hwnd,rect,false)` 第一实参；无返回值。

### invalidateInfoRectChange（0x1409be100）

`()` → `(hwnd uintptr, a, b image.Rectangle)`。rbx=hwnd（透传 InvalidateRect）、rcx/rdi/rsi/r8=a、
栈 [rsp+0xf0..0x108]=b；无返回值。

### shouldReuseControlHoverForWindow（0x1409be900）

`(hwnd uintptr)` → `(hwnd uintptr, x, y int, ctrlFound, winFound bool)`。
逻辑：recv[0x210]==hwnd 且 shouldReuseQRCodeControlHover(rect,point) 且
cachedControlHoverAllowedForPoint(hwnd,x,y,ctrlFound,winFound)。

### cachedControlHoverAllowedForPoint（0x1409bea00）

`(x,y int)` → `(hwnd uintptr, x, y int, ctrlFound, winFound bool)`。
逻辑：hover rect 非空后 `!shouldPreferWindowOverControlAtPoint(hoverRect,windowRect,x,y,ctrlFound,winFound)`。

## 关键知悉

- `shouldReuseQRCodeControlHover(rect, point)` 的 rect.Min 来自 recv[0x158]/[0x160]，rect.Max 来自
  本函数 x/y 实参（rect=从 hover 原点框到 point 的矩形），point 来自 recv[0x168]/[0x170]。
- `screenshotCurrentControlSelectionPreference`（0x140947040）返回 (bool,bool)，其两 bool 沿
  resolveControlHoverForWindow → shouldReuseControlHoverForWindow 透传为 ctrlFound/winFound。
- ctrlFound/winFound 语义与批次 220 的 shouldPreferWindowOverControlAtPoint 同名参数一致。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1417 / P=79 / UNMARKED=0`。
- **G3 行为**：纯签名修正（体零值→体零值），无行为变更；既有测试全量回归 PASS。
- **G4 review**：`screenshot_windows.go`（4 方法签名修正）；vet/test/build 复验通过。

## 遗留（下一批）

- screenshot_windows 域剩余 [P]：resolveControlHoverForWindow（参数 1 rbx 语义未确证，透传
  screenshotControlBoundsAtPointThroughOverlaySessionWithTimeout 的 r9）、
  resolveCaptureRectForWindowAtPoint、newScreenshotWindowSelectionSession、drawScreenshotInfoText、
  drawAppIconForPath、screenshotCandidateWindowAtPointWith、captureScreenshotAreaPNGForService*
  / captureScreenshotWindowSelectionWithOptions / captureScreenshotForegroundWindowWithProcessNameEx
  / captureScreenshotWithLauncherVisibility / finalizeScreenshotCaptureResult。
- P 已降至 79。下一批优先继续 screenshot_windows（resolveCaptureRectForWindowAtPoint 与已贯通
  hover 链同构），或转 oledblackout_windows（27 个，含大量匿名上下文闭包）。
