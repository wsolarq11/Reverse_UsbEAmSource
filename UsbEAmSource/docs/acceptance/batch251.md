# 批次 251 · BootstrapService/FileSearch 短函数 +3（含 ensureWorkspaceDirectories 方法改名订正）

## 目标

延续「已存在文件短函数」策略，落地 `hotkey_dispatch_stubs.go` + `filesearch_index_windows.go` 的
缺失函数；同时订正 `ensureWorkspaceDirectoriesForService` 方法命名（符号实证应为 `ensureWorkspaceDirectories`）。

## 基线 / 收口

| 指标 | 基线（batch 250 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2812 | 2814 |
| MARKED | 2812 | 2814 |
| S | 1280 | 1283 |
| S-inline | 36 | 36 |
| S-sig | 1456 | 1455 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2772（58.31%） | 2774（58.35%） |

构建：`bash build.sh` 成功，产物 `artifacts/UsbEAm_Launcher_rebuilt.exe`（21,322,240 B），
SHA256 `9a16bed6d9f6c6b3502293f42ec6a14440ec9583baa9ad143e644684a04602d5`。
`go build/vet/test ./backend` 全 EXIT=0。

## 本批落地（+2 FUNCS / +3 S / -1 S-sig）

### backend/hotkey_dispatch_stubs.go（+1 [S]）

1. `(*BootstrapService).finishScreenshotHotkeyCapture()` `[S 0x140795260]`
   = nil 检查 → lock(+0x540) → screenshotHotkeyActive(+0x27a)=false → unlock。
   （与 endScreenshotHotkeyCapture/deferwrap1 语义相同，但为独立符号。）

### backend/filesearch_index_windows.go（+1 [S]，新增 import time）

2. `timeToUnixNano(t time.Time) int64` `[S 0x1407f8640]`
   = `t.UnixNano()`（asm 为 time.Time.UnixNano 内联：wall 单调钟位 bit63 分派，
   wallToInternal + nsec 部分 + internalToUnix 偏移）。

## 本批订正（方法命名 + 体修正）

`backend/bootstrapservice_config.go` 的 `ensureWorkspaceDirectoriesForService`：

- 符号实证（symbols.txt:19195）= `main.BootstrapService.ensureWorkspaceDirectories`，
  落地名 `...ForService` 与符号不符，致 aggregate_gap.py 误判该方法缺失。
- 改名 `(*BootstrapService).ensureWorkspaceDirectories()`，档位 [S-sig]→[S]。
- 体修正：旧 `ensureWorkspaceDirectories(bs.workspace)`（直取字段、无锁）→
  新 `ensureWorkspaceDirectories(bs.workspaceSnapshot())`（asm 0x1407a26a0 实证：先 call
  workspaceSnapshot 加锁快照，再 duffcopy 大结构体传参）。
- 调用点 `bootstrapservice.go` ChooseLauncherBackgroundImage 同步改名。

## 关键知悉

- `finishScreenshotHotkeyCapture`(0x140795260) 与 `beginScreenshotHotkeyCapture.deferwrap1`
  (0x140795200) 是两个相邻符号，语义同为「加锁清零 screenshotHotkeyActive」。
- `timeToUnixNano` 是 time.Time.UnixNano 的直包装；asm 常量 0xdd7b17f80（wallToInternal）、
  0xa1b203eb3d1a0000（internalToUnix）、0x3b9aca00（1e9）与 Go 1.25 time 包内部一致。

## 下一批

P=40 不变。继续按 `gap_aggregate.txt` 长度升序落地已存在文件短函数；FUNCS 2814/4754 = 59.19%。
