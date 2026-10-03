# 批次 329 · GSMTC TryGetMediaPropertiesAsync/GetResults/GetStatus +3（FUNCS 3074）

## 目标

落地 3 个 GSMTC/WinRT COM 接口方法：`oledBlackoutGSMTCSession.TryGetMediaPropertiesAsync`、
`oledBlackoutIAsyncOperation.GetResults`、`oledBlackoutIAsyncInfo.GetStatus`。
并修正 `GetPlaybackInfo` 的 receiver 类型为 `oledBlackoutGSMTCSession`。

## 基线 / 收口

| 指标 | 基线（batch 328 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3071 | 3074 |
| MARKED | 3071 | 3074 |
| S | 1488 | 1491 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1504 | 1504 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1525 | 1528 |
| USABLE | 1526 | 1529 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1528  FUNCS=3074  MARKED=3074  P=41  S-eq=1  S-inline=37  S-sig=1504  S=1491  USABLE=1529
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1491 + 37 + 1 + 1504 + 41 = 3074 = FUNCS`。
S 1488→1491（+3）、FUNCS 3071→3074（+3）、FAITHFUL 1525→1528（+3）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 oledBlackoutGSMTCSession.TryGetMediaPropertiesAsync [S 0x140920ce0, 256B]

SyscallN(vtbl[0x38] TryGetMediaPropertiesAsync, this, &out) → HRESULT <0 → fmt.Errorf。

### 3.2 oledBlackoutIAsyncOperation.GetResults [S 0x140921360, 256B]

SyscallN(vtbl[0x40] GetResults, this, &out) → HRESULT <0 → fmt.Errorf。

### 3.3 oledBlackoutIAsyncInfo.GetStatus [S 0x140921a80, 256B]

SyscallN(vtbl[0x38] GetStatus, this, &status) → HRESULT <0 → fmt.Errorf。

## G4 独立复核

- `backend/oledblackout.go`：+oledBlackoutGSMTCSession 类型 +TryGetMediaPropertiesAsync
  +oledBlackoutIAsyncInfo 类型 +GetStatus +GetResults；GetPlaybackInfo receiver 修正为
  oledBlackoutGSMTCSession（类型名对齐 source_funcs.txt）。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（3 [S]）。FUNCS 3074/4754 = 64.66%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
