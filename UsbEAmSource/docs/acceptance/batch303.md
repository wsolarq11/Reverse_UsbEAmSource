# 批次 303 · 截图预览自动隐藏 + WinRT IAsyncOperation 取消 +2（FUNCS 2995）

## 目标

落地 2 个短函数：`screenshotPreviewWindowService.scheduleAutoHide`（自动隐藏调度）、
`oledBlackoutIAsyncOperation.Cancel`（WinRT 取消）。新建 `oledBlackoutIAsyncOperation` 类型。

## 基线 / 收口

| 指标 | 基线（batch 302 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2993 | 2995 |
| MARKED | 2993 | 2995 |
| S | 1453 | 1453 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1461 | 1463 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1490 | 1490 |
| USABLE | 1491 | 1491 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1490  FUNCS=2995  MARKED=2995  P=41  S-eq=1  S-inline=37  S-sig=1463  S=1453  USABLE=1491
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1453 + 37 + 1 + 1463 + 41 = 2995 = FUNCS`。
S-sig 1461→1463（+2）、FUNCS 2993→2995（+2）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 screenshotPreviewWindowService.scheduleAutoHide [S-sig 0x140999540, 160B]

`(d time.Duration)`。newobject 打包 func1（捕获 receiver + 参数），runtime.newproc 起 goroutine；
func1（0x1409995e0）定时器到点后 Hide。体待自动隐藏链专项。

### 3.2 oledBlackoutIAsyncOperation.Cancel [S-sig 0x140921960, 192B]

`()`。oledBlackoutQueryInterface 取 IAsyncInfo（nil/失败则返回），defer releaseComObject；
SyscallN(vtbl[9]@0x48 Cancel, this)。体待 QueryInterface 签名专项。新建类型
`oledBlackoutIAsyncOperation{vtbl *uintptr}`。

## G4 独立复核

- `backend/screenshot_services.go`：+scheduleAutoHide [S-sig]。
- `backend/oledblackout_windows.go`：+Cancel [S-sig]。
- `backend/types_oled.go`：+oledBlackoutIAsyncOperation 类型。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

2 个函数落地（2 [S-sig]）。FUNCS 2995/4754 = 63.00%。下一批：filesearch
SearchWithPaths 薄包装 + searchWithPathsContextMetrics 签名 / remoteicons
validateRemoteIconCachePath 路径校验。
