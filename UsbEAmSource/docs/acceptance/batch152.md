# 批次 152 验收（[P]→[S-sig] 全真转换：真函数口径突破 50%）

日期：2026-09-25
子批次：P-to-Ssig（签名逐寄存器实证，纠正 317 个 [P] 存根中的 120 个）

## 目标校正

批次 151 曾以 `FUNCS=2626` 声称达成 50%，但其中 317 个 `[P]`（签名未实证的零值存根）属
「统计口径水分」。本批按「必须全真」要求，将 50% 门槛改为**真函数口径**：
`真函数 = [S] + [S-inline] + [S-sig] ≥ 2377`，即 `[P]` 不计入进度。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath -buildmode=exe ./backend` | EXIT=0 |
| vet | `go vet ./backend` | EXIT=0 |
| 全量 test | `go test -count=1 -p=1 ./backend` | ok (0.528s) |
| gofmt | 本批 19 个转换文件 `gofmt -l` | 空（历史遗留文件未格式化与本批无关） |

## G2 覆盖（真函数口径）

`bash tools/count_funcs.sh`：`FUNCS=2626 / S=1041 / S-inline=35 / S-sig=1353 / P=197 / UNMARKED=0`。

**真函数 = 1041 + 35 + 1353 = 2429 / 4754 = 51.1%**，超过 50% 门槛 2377 共 52 个。

相对批次 151（S-sig=1233 / P=317，真函数 2309 = 48.6%）：**S-sig +120，P −120，真函数 +120**。

## G3 签名实证（[P]→[S-sig] 转换明细）

| 转换组 | 转换数 | 残留 [P] |
|---|---|---|
| screenshot_pin_native_windows.go | 27 | 1 |
| screenshot_pin.go | 18 | 7 |
| screenshot_uia_windows.go | 16 | 21 |
| screenshot_windows.go | 15 | 26 |
| oledblackout_windows.go | 11 | 37 |
| screenshot_uia_worker_windows.go | 10 | 4 |
| filelocator_runtime.go | 8 | 25 |
| audio_windows_runtime.go | 4 | 9 |
| screenshot_capability.go | 3 | 0 |
| launcherupdate_runtime.go | 3 | 23 |
| launcherconfig_runtime.go | 2 | 16 |
| filesearch_windows.go | 1 | 2 |
| oledblackout_overlay_windows.go | 1 | 1 |
| qrcode_windows.go | 1 | 17 |
| 合计 | **120** | 197 |

关键签名修正（推翻原 [P] 占位假设，均以寄存器 ABI + call 目标交叉验证）：

- `startWorkerLocked` 无参返回 `*screenshotCOMQueryWorker`（原误传 worker 参数）
- `executeRequest` 返回 `screenshotCOMQueryResult`（defer 分支恢复 7 寄存器 = rect4+controlType1+err2）
- `chooseMonitorRectForPoint` 返回 `(image.Rectangle, bool)`
- `compareText` 返回 `int`（cmpstring 返回码证伪原 bool）
- `resolveProcessDisplayName`/`resolveProcessIconData` 返回 `void→string`
- `handleMessage` 补 `hwnd` 参数（selection 版 4 参 `(hwnd,msg,wParam,lParam)`）
- `resolveInfoRectForHover` 无参（hover 读 session 字段，非形参）
- `buildDarkenedScreenshotPreview` 形参 `*image.RGBA`（读 RGBA.Rect 四字，非 qrCodeScreenSnapshot）
- `screenshotWindowFromAccessible` 返回 `(uintptr,bool)`（mov ebx,1 布尔模式，非 error）
- `oledBlackoutCurrentIdleSeconds` 返回 `() (int64,int64,int64)`（imul rax,0xf4240 纳秒换算）

## 剩余 197 个 [P] 的诚实归因

均为「现有 asm 证据无法唯一确定 Go 签名」的真硬骨头，非充数：

1. 16 字大结构体按值传参/返回走栈帧（`ScreenshotCaptureResult`/`WindowProcessPickResult`/`LauncherConfig`），
   寄存器无法唯一分解字段。
2. COM/UIA/MSAA vtable 方法（`*AtPointWithTimeoutAndPriority`/`*HitTest`/`AccessibleObjectFromWindow` 族），
   多参+多值返回，timeout/priority/session/role 定序难。
3. 首参接口 vs 多独立参数不可区分（收集器/枚举类）。
4. 栈传参（`add`/`mouseGestureWindowMovePosition` 10 参/`createTemporaryDirectoryShortcut` 读 `[rsp+...]`）。

这些需先落地对应结构体类型（栈帧偏移映射）或 COM 接口 vtable 布局，才能转 `[S-sig]`。

## 判定

四路 PASS，**50% 里程碑以真函数口径达成**（真函数 2429 ≥ 2377）。`P=197` 为诚实存根，
不再计入进度；其清零（`P=0`）是 100% 目标，需结构体/COM 类型落地后继续。

## 工具新增

`tools/p_funcs.py`：精确复刻 count_funcs.awk 的 last 状态机，提取所有 `[P]` 存根并反查
symbols.txt 得 VA（支持 `(s *Type)` 指针接收者），供 batch_sig_summary/batch_sig_evidence
生成签名证据。
