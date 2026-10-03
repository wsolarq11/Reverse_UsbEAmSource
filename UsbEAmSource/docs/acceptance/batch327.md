# 批次 327 · GSMTC 接口方法 RequestAsync/GetSessions/GetSize +3（FUNCS 3068）

## 目标

落地 3 个 GSMTC/WinRT COM 接口方法：`oledBlackoutGSMTCSessionManagerStatics.RequestAsync`、
`oledBlackoutGSMTCSessionManager.GetSessions`、`oledBlackoutIVectorView.GetSize`。

## 基线 / 收口

| 指标 | 基线（batch 326 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3065 | 3068 |
| MARKED | 3065 | 3068 |
| S | 1482 | 1485 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1504 | 1504 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1519 | 1522 |
| USABLE | 1520 | 1523 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1522  FUNCS=3068  MARKED=3068  P=41  S-eq=1  S-inline=37  S-sig=1504  S=1485  USABLE=1523
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1485 + 37 + 1 + 1504 + 41 = 3068 = FUNCS`。
S 1482→1485（+3）、FUNCS 3065→3068（+3）、FAITHFUL 1519→1522（+3）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 oledBlackoutGSMTCSessionManagerStatics.RequestAsync [S 0x1409200a0, 256B]

SyscallN(vtbl[0x30] RequestAsync, this, &out) → HRESULT <0 → fmt.Errorf。

### 3.2 oledBlackoutGSMTCSessionManager.GetSessions [S 0x1409201a0, 256B]

SyscallN(vtbl[0x38] GetSessions, this, &out) → HRESULT <0 → fmt.Errorf。

### 3.3 oledBlackoutIVectorView.GetSize [S 0x1409202a0, 256B]

SyscallN(vtbl[0x38] GetSize, this, &size) → HRESULT <0 → fmt.Errorf。

## G4 独立复核

- `backend/oledblackout.go`：+3 类型（vtbl 包装）+RequestAsync +GetSessions +GetSize [S]
  （+fmt/syscall/unsafe import）。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（3 [S]）。FUNCS 3068/4754 = 64.53%。下一批：GSMTC 系列剩余 COM 方法
（GetPlaybackInfo/GetPlaybackStatus/GetStatus/GetResults/GetAt/TryGetMediaPropertiesAsync）。
