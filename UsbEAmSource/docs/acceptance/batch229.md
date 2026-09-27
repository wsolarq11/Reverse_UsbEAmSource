# 批次 229 — waitForScreenshotCOMQueryWorkers 补齐 (workers, deadline) 签名，worker 域清零

## 基线 / 收口

| 指标 | 基线（批次 228 收口） | 收口（批次 229） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1428 | **1429** |
| P | 68 | **67** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2720 | **2721**（57.23%） |

SHA256 `2FCAA00FCD350EB27C35D3C88CD2D6D2DEC11805AC96BB1811F27904EB9CCD14`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/screenshot_uia_worker_windows.go`：`waitForScreenshotCOMQueryWorkers` 补齐签名，
补 `time` 导入。该文件 [P] 至此清零。

### waitForScreenshotCOMQueryWorkers（0x1409b6f00）

`func waitForScreenshotCOMQueryWorkers()`（无参 void，错误）
→ `func waitForScreenshotCOMQueryWorkers(workers []*screenshotCOMQueryWorker, deadline time.Time) bool`。

证据链：morestack 保存 rax..r8 六寄存器 = workers slice(ptr/len/cap，`[rax+rcx*8]` 8 字节元素遍历)
+ deadline time.Time(wall/ext/loc，透传 `time.Until` 计算剩余时间)；返回 bool（mov eax,0/1）。

## 关键知悉

- 6 寄存器参数 = slice(3) + time.Time(3)，非「pool/timer/计数」——判别要点：后 3 字直接进
  `time.Until`（time.Time 入参），前 3 字有 `[ptr+idx*8]` 元素遍历（slice）。
- screenshot_uia_worker_windows.go 的 4 个 [P]（beginShutdown/workerExited/waitForScreenshotCOMQueryWorkers/
  pumpScreenshotCOMThreadMessages）已全部升档，该文件归零。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1429 / P=67 / UNMARKED=0`。
- **G3 行为**：纯签名订正（体零值→体零值），无行为变更；既有测试全量回归 PASS。
- **G4 review**：`screenshot_uia_worker_windows.go`（1 函数签名订正 + time 导入）。

## 遗留（下一批）

- P 已降至 67。剩余分布：oledblackout_windows 25、screenshot_windows 12、launcherupdate_runtime 9、
  screenshot_uia_windows 8、screenshot_pin 6、filelocator_runtime 2、qrcode_windows 3、
  oledblackout_hotkey_windows 1。
- 可快速突破口：captureScreenshotWithLauncherVisibility（参数已确证，仅 cb 返回 6 字类型待逆推）；
  screenshotCandidateWindowAtPointWith (x/y+slice+hwnd+2 回调)；screenshot_uia 三函数返回 8 寄存器。
