# 批次 315 · 关闭到托盘钩子/idle 定时重调度 +2（FUNCS 3031）

## 目标

落地 2 个短函数：`registerLauncherCloseToTrayHook`（关闭到托盘钩子注册）、
`oledBlackoutService.rescheduleIdleTimerIfCurrent`（idle 定时重调度）。

## 基线 / 收口

| 指标 | 基线（batch 314 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3029 | 3031 |
| MARKED | 3029 | 3031 |
| S | 1471 | 1471 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1479 | 1481 |
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
FAITHFUL=1508  FUNCS=3031  MARKED=3031  P=41  S-eq=1  S-inline=37  S-sig=1481  S=1471  USABLE=1509
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1471 + 37 + 1 + 1481 + 41 = 3031 = FUNCS`。
S-sig 1479→1481（+2）、FUNCS 3029→3031（+2）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 registerLauncherCloseToTrayHook [S-sig 0x1408d0500, 224B]

nil 返回；newobject 闭包捕获 (回调, 参数, bs) → 间接调用 bs 钩子注册函数（+0x160）。

### 3.2 oledBlackoutService.rescheduleIdleTimerIfCurrent [S-sig 0x140903940, 224B]

lock(+0x00) → idleRunCurrentLocked → unlock；active 则 configureIdleTimer。

## G4 独立复核

- `backend/bootstrapservice_window.go`：+registerLauncherCloseToTrayHook [S-sig]。
- `backend/oledblackout.go`：+rescheduleIdleTimerIfCurrent [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

2 个函数落地（2 [S-sig]）。FUNCS 3031/4754 = 63.76%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
