# 批次 227 — beginShutdown 返回实为 worker 切片而非 int/error

## 基线 / 收口

| 指标 | 基线（批次 226 收口） | 收口（批次 227） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1426 | **1427** |
| P | 70 | **69** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2718 | **2719**（57.19%） |

SHA256 `C189CBCD161B915A5995B181A19FE83A236F2C7AC4EA5B1E7137380A86298285`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/screenshot_uia_worker_windows.go`：`beginShutdown` 的「3 寄存器返回」实为 slice 三字，
解开此前「int/error 组合无法唯一确定」的阻断。

### beginShutdown（0x1409b6800）

`func (p *screenshotCOMQueryWorkerPool) beginShutdown()`（void，错误）
→ `func (p *screenshotCOMQueryWorkerPool) beginShutdown() []*screenshotCOMQueryWorker`。

证据链：尾声返回 rax/rbx/rcx 三寄存器，来源为 `[rsp+0xb8]=rdi(ptr)`、`[rsp+0x48]=r8(len)`、
`[rsp+0x70]=rsi(cap)`；循环体内 growslice 增长后 `[rdi+r8*8-8]=worker 指针`（8 字节元素写），
配合 gcWriteBarrier2；`[r9+0x18]` 遍历 worker 链表取下一个 worker。故返回为收集到的 worker 切片。

## 关键知悉

- 「3 寄存器返回」≠ 一定是 `(int64/int, error)`：slice 也是 3 字（ptr/len/cap）。判别要点是
  返回槽三处来源是否分别对应切片头部三字段，且体循环是否有 `[ptr+len*8-8]` 的 8 字节元素写。
- worker 链表节点 `[node+0x18]` 指向下一 worker，元素类型确认为 `*screenshotCOMQueryWorker`。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1427 / P=69 / UNMARKED=0`。
- **G3 行为**：纯签名修正（体零值→体零值），无行为变更；既有测试全量回归 PASS。
- **G4 review**：`screenshot_uia_worker_windows.go`（1 函数签名订正）。

## 遗留（下一批）

- P 已降至 69。剩余分布：oledblackout_windows 25、screenshot_windows 12、launcherupdate_runtime 9、
  screenshot_uia_windows 8、screenshot_pin 6、screenshot_uia_worker_windows 3、
  filelocator_runtime 2、qrcode_windows 3、oledblackout_hotkey_windows 1。
- 可快速突破口：captureScreenshotWithLauncherVisibility 参数已确证 = (service, delay, hideLauncher, cb func()→6 字)，
  仅 cb 返回类型待逆推；workerExited 序言 3 非接收者参数（rbx/rdi/rcx）待分组；
  screenshotCandidateWindowAtPointWith (x/y+slice+hwnd+2 回调)。
