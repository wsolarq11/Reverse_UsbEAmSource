# 批次 345 · 屏幕收集/屏幕索引/非目标隐藏/输入轮询调度 +4（FUNCS 3135）

## 目标

落地 4 个函数：`oledBlackoutService.collectScreensLocked`、`oledBlackoutScreenIndex`、
`oledBlackoutService.hideNonTargetWindowsLocked`、`oledBlackoutService.scheduleInputPollLocked`。

## 基线 / 收口

| 指标 | 基线（batch 344 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3131 | 3135 |
| MARKED | 3131 | 3135 |
| S | 1501 | 1501 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1551 | 1555 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1538 | 1538 |
| USABLE | 1539 | 1539 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1538  FUNCS=3135  MARKED=3135  P=41  S-eq=1  S-inline=37  S-sig=1555  S=1501  USABLE=1539
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1501 + 37 + 1 + 1555 + 41 = 3135 = FUNCS`。
S-sig 1551→1555（+4）、FUNCS 3131→3135（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 collectScreensLocked [S-sig 0x140906140, 320B]

app(+0x30)/screenManager(+0x310) 空→nil；GetAll → 过滤 nil → sortOLEDBlackoutScreens。

### 3.2 oledBlackoutScreenIndex [S-sig 0x14090a340, 320B]

TrimSpace(name) → 遍历屏幕 → 匹配返回 index+1（1-based）。

### 3.3 hideNonTargetWindowsLocked [S-sig 0x14090abe0, 320B]

遍历 overlay(+0xf0) → 非目标→Hide + mapdelete(+0xf8)。

### 3.4 scheduleInputPollLocked [S-sig 0x14090c2a0, 320B]

ensureLifecycleLocked → AfterFunc → 条件不符 stopTimer。

## G4 独立复核

- `backend/oledblackout.go`：+collectScreensLocked [S-sig] +oledBlackoutScreenIndex [S-sig]
  +hideNonTargetWindowsLocked [S-sig] +scheduleInputPollLocked [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3135/4754 = 65.95%。下一批：oledBlackout 域
handleOverlayDismissOnEscape / newWindowsOLEDBlackoutCursorController。
