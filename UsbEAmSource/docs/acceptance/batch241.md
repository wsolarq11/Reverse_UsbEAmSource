# 批次 241 — downloadLauncherUpdatePackageRange 签名订正（1 升档）

## 基线 / 收口

| 指标 | 基线（批次 240 收口） | 收口（批次 241） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1445 | **1446** |
| P | 51 | **50** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2737（57.57%） | **2738**（57.59%） |

SHA256 `e8d5bbe189b7dd73054bc07eb69d17a1ccaf6e20907d0b360737e6425306bde1`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/launcherupdate_runtime.go`：`downloadLauncherUpdatePackageRange` 升档 `[S-sig]`，签名
`(ctx context.Context, url string, r launcherUpdatePackageRange, dir string) error` →
`(ctx context.Context, access LauncherNetworkAccess, url string, file *os.File, r launcherUpdatePackageRange) error`。
新增 `"os"` import。

### 签名证据（0x1408bb800）

morestack 序言（第 413-433 行）保存 9 槽，逐寄存器实证：

| 项 | 类型 | 证据 |
|---|---|---|
| rax/rbx | ctx context.Context | NewRequestWithContext 首参透传 |
| rcx/rdi | access LauncherNetworkAccess（itab/data） | 0x1408bbb0c 读 `[itab+0x18]`=Do 方法槽 |
| rsi/r8 | url string（ptr/len） | NewRequestWithContext 第三参透传 |
| r9 | file *os.File | 0x1408bbdb9 作 `os.File.WriteAt` 接收者 |
| r10/r11 | r launcherUpdatePackageRange（Start/End） | 0x1408bb860 `cmp r10,r11`，len=End-Start+1 |

返回 rax/rbx = error(2)，成功路径 xmm15 清零（nil）。故返回 `error`。

### 关键纠正

旧注释「rsi/r9/r10/r11 与 range 结构体字段归属无法唯一确定」已解：完整形参为
`(ctx, access, url, file, r)`——旧签名既漏 `access` 参数，又误把 `dir string` 当作下载目标
（实为已打开的 `file *os.File`，体内直接 `WriteAt` 落盘，无任何 open/mkdir 路径）。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`vet ./backend` EXIT=0；
  `test ./backend` `ok changeme/backend`。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1446 / P=50 / UNMARKED=0`。
- **G3 行为**：纯签名订正（体空→体空零骨架 `return nil`），无调用方，无行为变更；
  既有测试全量 PASS。
- **G4 review**：`launcherupdate_runtime.go`（1 升档 + 1 import）。

## 遗留（下一批）

- P 已降至 50。截图域：captureScreenshotAreaPNG（7 参数）、
  captureScreenshotWindowSelectionWithOptions、newScreenshotWindowSelectionSession、
  resolveControlHoverForWindow、screenshotAccessibleHitTest。
- launcherupdate_runtime.go 剩余 7 个 [P]：runLauncherUpdateTask（7 参）、
  fetchLauncherRemoteConfigWithWorkspace、beginLauncherUpdateTask（6 寄存器返回）、
  prepareLauncherUpdatePackageWithWorkspace、downloadLauncherUpdatePackageWithWorkspace、
  downloadLauncherUpdatePackageSequentially/Concurrently。
- filelocator walkRoot/processFile；oledblackout_windows 25；nativedrag createTemporaryDirectoryShortcut。
