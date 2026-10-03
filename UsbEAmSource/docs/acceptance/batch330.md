# 批次 330 · 线性 alpha 转字节/启动应用/媒体连续性提交 +3（FUNCS 3077）

## 目标

落地 3 个短函数：`screenshotLinearAlphaToByte`（线性 alpha 转字节）、
`BootstrapService.LaunchApp`（启动应用）、
`oledBlackoutService.commitMediaContinuityIfCurrent`（媒体连续性提交）。

## 基线 / 收口

| 指标 | 基线（batch 329 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3074 | 3077 |
| MARKED | 3074 | 3077 |
| S | 1491 | 1492 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1504 | 1506 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1528 | 1529 |
| USABLE | 1529 | 1530 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1529  FUNCS=3077  MARKED=3077  P=41  S-eq=1  S-inline=37  S-sig=1506  S=1492  USABLE=1530
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1492 + 37 + 1 + 1506 + 41 = 3077 = FUNCS`。
S 1491→1492（+1）、S-sig 1504→1506（+2）、FUNCS 3074→3077（+3）、FAITHFUL 1528→1529（+1）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 screenshotLinearAlphaToByte [S 0x14097de80, 256B]

NaN/<=0→0；>=1.0→255；否则 *255 后 round-to-nearest-even → clamp [0,255]。

### 3.2 BootstrapService.LaunchApp [S-sig 0x1408a6020, 160B]

duffcopy 参数 → LaunchAppWithPrivilege(..., "normal")。

### 3.3 oledBlackoutService.commitMediaContinuityIfCurrent [S-sig 0x140903a20, 256B]

lock(+0x00) → idleRunCurrentLocked → 真则 replace(+0x20, arg)。

## G4 独立复核

- `backend/screenshot_dxgi_deep_windows.go`：+screenshotLinearAlphaToByte [S]。
- `backend/applaunch_windows.go`：+BootstrapService.LaunchApp [S-sig]。
- `backend/oledblackout.go`：+commitMediaContinuityIfCurrent [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（1 [S] + 2 [S-sig]）。FUNCS 3077/4754 = 64.72%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
