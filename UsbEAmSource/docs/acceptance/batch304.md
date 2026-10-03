# 批次 304 · 窗口相对位置/托盘显示/光标恢复/带 ctx 读取 +4（FUNCS 2999）

## 目标

落地 4 个短函数：`resolveLauncherWindowRelativePosition`（居中/钳位算术）、
`BootstrapService.showLauncherFromTray`（托盘显示）、
`windowsOLEDBlackoutCursorController.restoreOnThread`（光标恢复）、
`fileLocatorContextReader.Read`（带 ctx 读取）。

## 基线 / 收口

| 指标 | 基线（batch 303 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2995 | 2999 |
| MARKED | 2995 | 2999 |
| S | 1453 | 1455 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1463 | 1465 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1490 | 1492 |
| USABLE | 1491 | 1493 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1492  FUNCS=2999  MARKED=2999  P=41  S-eq=1  S-inline=37  S-sig=1465  S=1455  USABLE=1493
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1455 + 37 + 1 + 1465 + 41 = 2999 = FUNCS`。
S 1453→1455（+2）、S-sig 1463→1465（+2）、FUNCS 2995→2999（+4）、FAITHFUL 1490→1492（+2）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 resolveLauncherWindowRelativePosition [S 0x1408d0a80, 80B]

6 int + 1 bool 参数，返回 (x, y)。center=true：x/y = max((屏-窗)>>1, 0)；
center=false：x/y = clamp(屏-窗, 0, 上限)，上限为负置 0，屏<=0 取上限。

### 3.2 BootstrapService.showLauncherFromTray [S 0x140795ce0, 80B]

ensureLauncherWindowForShow 取窗口（interface 断言解包）→
showLauncherWindow(window, setWindow=true, show=false)。

### 3.3 windowsOLEDBlackoutCursorController.restoreOnThread [S-sig 0x140913e60, 192B]

receiver/restore(+0x18) nil → 错误返回；非 nil 循环 n 次间接调用恢复闭包。

### 3.4 fileLocatorContextReader.Read [S-sig 0x140a04f20, 192B]

receiver nil → panicwrap；解包 reader(+0x00:8)/ctx(+0x10:8) 两接口，ctx 与读取参数重组后
间接调用底层 Read。

## G4 独立复核

- `backend/bootstrapservice_window.go`：+resolveLauncherWindowRelativePosition [S]。
- `backend/hotkey_dispatch_stubs.go`：+showLauncherFromTray [S]。
- `backend/oledblackout_windows.go`：+restoreOnThread [S-sig]。
- `backend/types_filelocator.go`：+fileLocatorContextReader.Read [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（2 [S] + 2 [S-sig]）。FUNCS 2999/4754 = 63.08%。下一批：filesearch
SearchWithPaths 薄包装 + searchWithPathsContextMetrics 签名 / remoteicons
validateRemoteIconCachePath 路径校验。
