# 批次 225 — screenshot_windows 两 [P] 升档 [S-sig]

## 基线 / 收口

| 指标 | 基线（批次 224 收口） | 收口（批次 225） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1423 | **1425** |
| P | 73 | **71** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2715 | **2717**（57.15%） |

SHA256 `32522D7887B40F39D2BBC61762EC3834A46AE4B1CEDE62A413DEA41F7FB31776`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/screenshot_windows.go` 两个 [P] 升档 [S-sig]。

### drawAppIconForPath（0x1409bfc00）

`(path string)` → `(path string, rect image.Rectangle)`。序言 rax/rbx=path(string 2 字) +
rdi/rsi/r8/r9=rect(4 int，cmp rdi,r8 / rsi,r9 判非空即直返)；返回 void。旧存根漏 rect，已订正。

### captureScreenshotForegroundWindowWithProcessNameEx（0x1409bab20）

`(processName string) ScreenshotCaptureResult` 原签名正确。序言 rax/rbx=processName(string 2 字)，
xor eax 后调 captureQRCodeVirtualScreenSnapshotWithCursor(无参，返回 6 字快照)；返回
ScreenshotCaptureResult(16 字)。

## 关键知悉

- captureScreenshotForegroundWindowWithProcessNameEx 原签名已正确（string→ScreenshotCaptureResult），
  仅「未逐寄存器实证」——对齐后升档。
- finalizeScreenshotCaptureResult 仍 [P]：序言 8 寄存器参数，与 ScreenshotCaptureResult 16 字不符，
  真实参数形态（8 字结构体或 8 独立参数）待解，非简单「ScreenshotCaptureResult→ScreenshotCaptureResult」。
- drawScreenshotInfoText 序言含 5+ 参数（hdc+string+rect+int32），保持 [P]。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1425 / P=71 / UNMARKED=0`。
- **G3 行为**：纯签名修正（体零值→体零值），无行为变更；既有测试全量回归 PASS。
- **G4 review**：`screenshot_windows.go`（2 函数签名修正）。

## 遗留（下一批）

- P 已降至 71。剩余分布：oledblackout_windows 25、screenshot_windows 13、launcherupdate_runtime 9、
  screenshot_uia_windows 8、screenshot_pin 6、screenshot_uia_worker_windows 4、
  filelocator_runtime 2、qrcode_windows 3、oledblackout_hotkey_windows 1。
- 可快速突破口：screenshot_windows 的 finalizeScreenshotCaptureResult（8 字参数形态待解）、
  captureScreenshotWithLauncherVisibility（service+hideLauncher，2 参+返回结构体）、
  screenshotCandidateWindowAtPointWith（x/y+slice+hwnd+rect 组合）。
