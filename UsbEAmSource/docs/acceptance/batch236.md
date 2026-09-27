# 批次 236 — captureScreenshotWithLauncherVisibility 签名订正（1 升档）

## 基线 / 收口

| 指标 | 基线（批次 235 收口） | 收口（批次 236） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1440 | **1441** |
| P | 56 | **55** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2732 | **2733**（57.48%） |

SHA256 `1CBD4A981CC6F76729DD7AC90ACFED398FD547218F0A8ED4C8DC5177E2B314F8`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批内容

`backend/screenshot_windows.go`：`captureScreenshotWithLauncherVisibility` 升档 `[S-sig]`，
签名 `(service, hideLauncher bool) ScreenshotCaptureResult` → `(service, delay time.Duration,
hideLauncher bool, cb func() ([]byte,bool,error)) ([]byte,bool,error)`。

### 签名证据（0x1409b94c0）

morestack 保护 4 寄存器：rax、rbx、cl(byte)、rdi。逐参判定：

| 寄存器 | 类型 | 证据 |
|---|---|---|
| rax | service | `test rax,rax` + 读 `[rax+0x408]`（截图预览服务） |
| rbx | delay time.Duration | `test rbx,rbx; jle` 后 `rax=rbx; call time.Sleep` |
| cl | hideLauncher bool | 存 `[rsp+0x110]` 后参与 begin/end 分支 |
| rdi | cb func() ([]byte,bool,error) | `test rdi,rdi` 后 `rax=[rdi]; rdx=rdi; call rax` |

返回 6 字：内部 `call cb` 后把 (rax,rbx,rcx,dil,rsi,r8) 存 `[rsp+0x38/0x40/0x48/0x25/0x28/0x30]`
再逐字透传（第 140-145 行）。错误路径返回 `(nil, false, err)`（3 字清零 + bool=0 + error 两字）。
cb 实现 `captureScreenshotAreaPNGForServiceWithControlsAndLauncherVisibility.func1`（0x1409b7660）
尾声同样返回 6 字 (3 字, bool, error)，确认 6 字 = `([]byte, bool, error)`。

### 函数语义

隐藏启动器窗口 → `beginLauncherScreenshotCapture` 保存原状态 → `time.Sleep(delay)` →
调用 cb 取得截屏结果 → `endLauncherScreenshotCapture` 恢复窗口 → 透传 cb 结果。

## 关键知悉

- **6 字返回 = ([]byte, bool, error)**：旧注释"返回 3 字与 ScreenshotCaptureResult 16 字不符"
  误读了返回宽度；实为 []byte(3) + bool(1) + error(2) 共 6 字。
- **cb 是调用者闭包**：rdi 参数是 `func() ([]byte,bool,error)` 值（非内部 func1.1 清理闭包），
  内部 func1/func1.1 是 sync.Once 保护的隐藏/恢复回调。
- **联动解锁**：本订正同时厘清了 captureScreenshotAreaPNG 的返回 = ([]byte,bool,error)（透传），
  及 captureScreenshotWindowSelectionWithOptions 的返回形态（大结构体走栈，待后续批次）。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`vet ./backend` EXIT=0；
  `test ./backend` `ok changeme/backend`。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1441 / P=55 / UNMARKED=0`。
- **G3 行为**：纯签名订正（体空→体空），无行为变更；既有测试全量 PASS。
- **G4 review**：`screenshot_windows.go`（1 升档）。

## 遗留（下一批）

- P 已降至 55。截图域剩余复杂 P：captureScreenshotAreaPNG（7 参数，参数4/6/7 需 qrCodeNativeSelectionOptions
  字段映射）、captureScreenshotWindowSelectionWithOptions（5 参数 + WindowProcessPickResult 大结构返回）、
  newScreenshotWindowSelectionSession（快照 6 字 + 4 栈参数）。
- filelocator walkRoot/processFile（值传大结构体 prepared + 未还原聚合结构体 r9/r10）。
- oledblackout_windows 25 个（WinRT/COM 匿名上下文）。
