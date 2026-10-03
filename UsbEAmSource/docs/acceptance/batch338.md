# 批次 338 · 调度停止/连续性遗忘/空闲定时器/光标控制器 +4（FUNCS 3107）

## 目标

落地 4 个函数：`desktopWidgetScheduler.Stop`、
`oledBlackoutBrowserMediaContinuity.forget`、`oledBlackoutService.stopIdleTimerLocked`、
`windowsOLEDBlackoutCursorController.Close`。

## 基线 / 收口

| 指标 | 基线（batch 337 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3103 | 3107 |
| MARKED | 3103 | 3107 |
| S | 1499 | 1500 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1525 | 1528 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1536 | 1537 |
| USABLE | 1537 | 1538 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1537  FUNCS=3107  MARKED=3107  P=41  S-eq=1  S-inline=37  S-sig=1528  S=1500  USABLE=1538
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1500 + 37 + 1 + 1528 + 41 = 3107 = FUNCS`。
S 1499→1500（+1）、S-sig 1525→1528（+3）、FUNCS 3103→3107（+4）、FAITHFUL 1536→1537（+1）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 desktopWidgetScheduler.Stop [S-sig 0x1407acd20, 288B]

nil 早退；stopOnce(+0x2c) 关 stop；once(+0x20) 关 wake；NewTimer(5s) + selectgo 等待。

### 3.2 forget [S 0x1408fcb20, 288B]

nil 早退；key.processPath=ToLower(TrimSpace(processPath))；lock → delete(entries, key) → unlock。
（key = oledBlackoutBrowserMediaContinuityKey{windowHandle,processID,processPath}）

### 3.3 stopIdleTimerLocked [S-sig 0x140902440, 288B]

idleTimer(+0x128) 非空则 stopTimer + 置 nil；generation(+0x130)++；清空多个字段。

### 3.4 Close [S-sig 0x140914120, 288B]

closeOnce 保护 → 发 close 命令 → 等待响应。

## G4 独立复核

- `backend/desktopwidgets_scheduler.go`：+Stop [S-sig]。
- `backend/oledblackout.go`：+forget [S] +stopIdleTimerLocked [S-sig] +Close [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（1 [S] + 3 [S-sig]）。FUNCS 3107/4754 = 65.36%。下一批：oledBlackout
inputDismissGuardActiveLocked / pluginWindowService.CloseAudience / screenshotPreviewBounds。
