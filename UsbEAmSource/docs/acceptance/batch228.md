# 批次 228 — workerExited 补齐 err error 参数

## 基线 / 收口

| 指标 | 基线（批次 227 收口） | 收口（批次 228） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1427 | **1428** |
| P | 69 | **68** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2719 | **2720**（57.21%） |

SHA256 `B208F9EC0E9055A8F57175A037002FD583C2AE53BF6CA677A0BFAE304B47A647`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/screenshot_uia_worker_windows.go`：`workerExited` 补齐 `err error` 参数。

### workerExited（0x1409b6080）

`func (p *screenshotCOMQueryWorkerPool) workerExited(worker *screenshotCOMQueryWorker)`
→ `func (p *screenshotCOMQueryWorkerPool) workerExited(worker *screenshotCOMQueryWorker, err error)`。

证据链：序言 spill rax(recv)+rbx(worker 指针，`[rbx]`=key)+rcx/rdi(err error 两字)；L75
`test rsi` 判 err itab 是否 nil（即 err==nil）；mapdelete_fast64 用 `[rbx]` 作 key 删 worker 映射；
返回 void。

## 关键知悉

- 参数寄存器多出的 2 字（rcx/rdi）若其一被 `test …nil` + 条件跳转，且另一被当作配对值透传，
  大概率是 `error` 接口（itab+data）——workerExited 正是如此。
- 同文件 workerExited/beginShutdown 已双双升档，screenshot_uia_worker 仅剩 waitForScreenshotCOMQueryWorkers 1 个 [P]。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1428 / P=68 / UNMARKED=0`。
- **G3 行为**：纯签名订正（体零值→体零值），无行为变更；既有测试全量回归 PASS。
- **G4 review**：`screenshot_uia_worker_windows.go`（1 函数签名订正）。

## 遗留（下一批）

- P 已降至 68。剩余分布：oledblackout_windows 25、screenshot_windows 12、launcherupdate_runtime 9、
  screenshot_uia_windows 8、screenshot_pin 6、filelocator_runtime 2、qrcode_windows 3、
  oledblackout_hotkey_windows 1、screenshot_uia_worker_windows 1。
- 可快速突破口：captureScreenshotWithLauncherVisibility（参数已确证，仅 cb 返回 6 字类型待逆推）；
  screenshotCandidateWindowAtPointWith (x/y+slice+hwnd+2 回调)；screenshot_uia 三函数返回 8 寄存器。
