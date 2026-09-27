# 批次 226 — finalizeScreenshotCaptureResult 解开 16 字大结构真实签名

## 基线 / 收口

| 指标 | 基线（批次 225 收口） | 收口（批次 226） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1425 | **1426** |
| P | 71 | **70** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2717 | **2718**（57.17%） |

SHA256 `B2FA8AD018B8CC9891F86477EE514EBD1DB9A201F1AB0228FDF39B66BC61BFB2`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/screenshot_windows.go`：`finalizeScreenshotCaptureResult` 解开真实签名（此前误作
16 字 `ScreenshotCaptureResult` 值传），补 `time` 导入。

### finalizeScreenshotCaptureResult（0x1409b8d80）

`(result ScreenshotCaptureResult) ScreenshotCaptureResult`（16 字值传，错误）
→ `(service *BootstrapService, png []byte, source string, captureTime time.Time, format string) ScreenshotCaptureResult`。

证据链：
- morestack 只保存 rax..r10 八寄存器（无 r11），故栈参数 = captureTime.loc(1) + format(2) = 3 字。
- rax=service（recv），rbx/rcx/rdi=png(3 字) 直接透传 `copyScreenshotPNGToClipboard`；
  rsi/r8=source(2 字)，r9/r10=captureTime.wall/ext；全部原样透传
  `BootstrapService.saveScreenshotCaptureResultWithFormat(png, source, captureTime, format)`。
- 返回 16 字 = `buildScreenshotResultFromPNG(png)` 结果 duffcopy 回填。
- 函数体内 time.Now/time.Since 为计时，captureTime 实为参数而非函数体计算。

## 关键知悉

- 16 字大结构值传（>9 字走栈单指针）的旧假设是错的：该函数实为 8 寄存器 + 3 栈 = 11 字参数，
  恰好等于 saveScreenshotCaptureResultWithFormat 的参数全集加 recv。
- 同类「序言 spill 8 寄存器」的大结构 [P] 存根，应优先检查是否透传给已知签名的下游方法，
  用下游签名反推参数分组。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1426 / P=70 / UNMARKED=0`。
- **G3 行为**：纯签名修正（体零值→体零值），无行为变更；既有测试全量回归 PASS。
- **G4 review**：`screenshot_windows.go`（1 函数签名订正 + time 导入）。

## 遗留（下一批）

- P 已降至 70。剩余分布：oledblackout_windows 25、screenshot_windows 12、launcherupdate_runtime 9、
  screenshot_uia_windows 8、screenshot_pin 6、screenshot_uia_worker_windows 4、
  filelocator_runtime 2、qrcode_windows 3、oledblackout_hotkey_windows 1。
- 可快速突破口：captureScreenshotWithLauncherVisibility 参数已确证 = (service, delay, hideLauncher,
  cb func()→6 字)，仅 cb 返回类型待逆推（调用方 func1/func1.1 可证）；screenshotCandidateWindowAtPointWith
  (x/y+slice+hwnd+rect 组合)；beginShutdown 返回 3 寄存器 int/error 组合。
