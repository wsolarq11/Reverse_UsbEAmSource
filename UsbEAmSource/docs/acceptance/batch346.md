# 批次 346 · Esc 消除/光标控制器构造与循环/插件窗口关闭/链接浏览器校验 +5（FUNCS 3140）

## 目标

落地 5 个函数：`oledBlackoutService.handleOverlayDismissOnEscape`、
`newWindowsOLEDBlackoutCursorController`、`windowsOLEDBlackoutCursorController.run`、
`pluginWindowService.Close`、`validateExplicitLinkBrowserResolvedTarget`。

## 基线 / 收口

| 指标 | 基线（batch 345 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3135 | 3140 |
| MARKED | 3135 | 3140 |
| S | 1501 | 1503 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1555 | 1558 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1538 | 1540 |
| USABLE | 1539 | 1541 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1540  FUNCS=3140  MARKED=3140  P=41  S-eq=1  S-inline=37  S-sig=1558  S=1503  USABLE=1541
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1503 + 37 + 1 + 1558 + 41 = 3140 = FUNCS`。
S 1501→1503（+2）、S-sig 1555→1558（+3）、FUNCS 3135→3140（+5）、FAITHFUL 1538→1540（+2）、
USABLE 1539→1541（+2）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 handleOverlayDismissOnEscape [S-sig 0x14090af20, 320B]

lock → 字段(+0x88/+0x38/+0x3a) 判定 → dismissOverlayForKeyboardInputLocked。

### 3.2 newWindowsOLEDBlackoutCursorController [S 0x140913520, 320B]

makechan(commands/ready/done) → newobject 填充 call(+0x18) → newproc(goroutine) →
chanrecv1(ready) → 返回 controller。字段顺序与 `windowsOLEDBlackoutCursorController` 对齐。

### 3.3 run [S 0x1409136c0, 512B]

LockOSThread → defer UnlockOSThread → defer close(done) → close(ready) →
循环 <-commands 分发。

### 3.4 pluginWindowService.Close [S-sig 0x14092f500, 320B]

TrimSpace 空→error；lock → windows(+0x10)[id] → unlock；win 非空且 audience(+0x88)
非空则调 close(+0x30)。

### 3.5 validateExplicitLinkBrowserResolvedTarget [S-sig 0x1408aab60, 320B]

TrimSpace 空→error；isWindowsAbsoluteFilesystemPath → classifyAutomaticWindowsPath +
isDisallowedExplicitLinkBrowserExecutable。

## G4 独立复核

- `backend/oledblackout.go`：+handleOverlayDismissOnEscape [S-sig]。
- `backend/oledblackout_windows.go`：+newWindowsOLEDBlackoutCursorController [S] +run [S]。
- `backend/bootstrapservice_state_deps.go`：+pluginWindowService.Close [S-sig]。
- `backend/linkbrowser.go`：+validateExplicitLinkBrowserResolvedTarget [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

5 个函数落地（2 [S] + 3 [S-sig]）。FUNCS 3140/4754 = 66.06%。下一批：launcher 域
resolveLauncherNotificationIconPath / launcherAppIdentityLPWSTRPropVariant。
