# 批次 306 · 重定向策略/OLED 覆盖层 +4（FUNCS 3005）

## 目标

落地 4 个短函数：`directLauncherNetworkAccess.DoWithRedirectPolicy`、
`configBackedLauncherNetworkAccess.DoWithRedirectPolicy`（重定向策略 wrapper）、
`oledBlackoutProfileKey`（profile 键）、`oledBlackoutService.dismissOverlayForKeyboardInputLocked`
（键盘锁定覆盖层关闭）。

## 基线 / 收口

| 指标 | 基线（batch 305 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3001 | 3005 |
| MARKED | 3001 | 3005 |
| S | 1456 | 1456 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1466 | 1470 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1493 | 1493 |
| USABLE | 1494 | 1494 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1493  FUNCS=3005  MARKED=3005  P=41  S-eq=1  S-inline=37  S-sig=1470  S=1456  USABLE=1494
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1456 + 37 + 1 + 1470 + 41 = 3005 = FUNCS`。
S-sig 1466→1470（+4）、FUNCS 3001→3005（+4）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 directLauncherNetworkAccess.DoWithRedirectPolicy [S-sig 0x140a05b40, 128B]

wrapper 转发：参数重排后调用 0x1408a5d60 具体实现（redirect policy 接口方法）。

### 3.2 configBackedLauncherNetworkAccess.DoWithRedirectPolicy [S-sig 0x140a05980, 160B]

wrapper 转发：interface 解包后调用具体实现。

### 3.3 oledBlackoutProfileKey [S-sig 0x1408fef00, 96B]

normalizeOLEDBlackoutProfileScreens(..., nil) → []string，strings.Join(切片, ",")。

### 3.4 oledBlackoutService.dismissOverlayForKeyboardInputLocked [S-sig 0x14090b2a0, 128B]

visibleOverlayForKeyboardInputLocked 取覆盖层（nil 则返回），非 nil 则 dismissOverlayForInputLocked。

## G4 独立复核

- `backend/networkaccess.go`：+DoWithRedirectPolicy ×2 [S-sig]。
- `backend/oledblackout.go`：+oledBlackoutProfileKey +dismissOverlayForKeyboardInputLocked [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3005/4754 = 63.22%。下一批：filesearch
SearchWithPaths 薄包装 + searchWithPathsContextMetrics 签名 / remoteicons
validateRemoteIconCachePath 路径校验。
