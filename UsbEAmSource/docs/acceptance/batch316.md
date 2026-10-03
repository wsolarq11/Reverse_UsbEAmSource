# 批次 316 · OLED 热键重试/覆盖层可见性/轮询关闭 +3（FUNCS 3034）

## 目标

落地 3 个短函数：`retryOLEDBlackoutHotkeysIfNeeded`（热键重试）、
`oledBlackoutService.visibleOverlayForKeyboardInputLocked`（键盘锁定覆盖层）、
`oledBlackoutService.dismissOverlayFromInputPollLocked`（轮询关闭覆盖层）。

## 基线 / 收口

| 指标 | 基线（batch 315 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3031 | 3034 |
| MARKED | 3031 | 3034 |
| S | 1471 | 1471 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1481 | 1484 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1508 | 1508 |
| USABLE | 1509 | 1509 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1508  FUNCS=3034  MARKED=3034  P=41  S-eq=1  S-inline=37  S-sig=1484  S=1471  USABLE=1509
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1471 + 37 + 1 + 1484 + 41 = 3034 = FUNCS`。
S-sig 1481→1484（+3）、FUNCS 3031→3034（+3）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 retryOLEDBlackoutHotkeysIfNeeded [S-sig 0x14079e080, 224B]

lock(+0x00) → 读标志(+0x88/+0x71/+0x118) → unlock；需重试则 configureHotkeys。

### 3.2 oledBlackoutService.visibleOverlayForKeyboardInputLocked [S-sig 0x14090bba0, 224B]

TrimSpace → 查找 overlay 映射(+0xf8)；未命中则 visibleOverlayOrderLocked。

### 3.3 oledBlackoutService.dismissOverlayFromInputPollLocked [S-sig 0x14090cbe0, 224B]

dismissOverlayScreenLocked → 非空则回调并写状态(+0xe0/+0xe8)；否则 maybeStartQuickIdle。

## G4 独立复核

- `backend/oledblackout.go`：+retryOLEDBlackoutHotkeysIfNeeded
  +visibleOverlayForKeyboardInputLocked +dismissOverlayFromInputPollLocked [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（3 [S-sig]）。FUNCS 3034/4754 = 63.82%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
