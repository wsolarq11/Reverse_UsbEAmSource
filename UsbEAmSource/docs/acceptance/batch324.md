# 批次 324 · idle 阈值/newHTTPClient/显式浏览器启动 +3（FUNCS 3058）

## 目标

落地 3 个短函数：`oledBlackoutService.idleThresholdForProfileLocked`（idle 阈值）、
`configBackedLauncherNetworkAccess.newHTTPClient`（HTTP 客户端构建）、
`startLinkWithExplicitBrowser`（显式浏览器启动）。

## 基线 / 收口

| 指标 | 基线（batch 323 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3055 | 3058 |
| MARKED | 3055 | 3058 |
| S | 1481 | 1481 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1495 | 1498 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1518 | 1518 |
| USABLE | 1519 | 1519 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1518  FUNCS=3058  MARKED=3058  P=41  S-eq=1  S-inline=37  S-sig=1498  S=1481  USABLE=1519
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1481 + 37 + 1 + 1498 + 41 = 3058 = FUNCS`。
S-sig 1495→1498（+3）、FUNCS 3055→3058（+3）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 oledBlackoutService.idleThresholdForProfileLocked [S-sig 0x140907020, 256B]

quickIdleAppliesToProfileLocked(TrimSpace(profile)) → 5s；否则 minutes 钳位 [1,60] → *time.Minute。

### 3.2 configBackedLauncherNetworkAccess.newHTTPClient [S-sig 0x1408a5960, 192B]

newTransport → 错误则返回；newobject(http.Client) → 写 Transport/Timeout。

### 3.3 startLinkWithExplicitBrowser [S-sig 0x140769a60, 192B]

buildLinkCommand → len<1 则 panicSliceB；命令[1:] → startExplicitLinkBrowser。

## G4 独立复核

- `backend/oledblackout.go`：+idleThresholdForProfileLocked [S-sig]。
- `backend/networkaccess.go`：+newHTTPClient [S-sig]。
- `backend/linkbrowser.go`：+startLinkWithExplicitBrowser [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（3 [S-sig]）。FUNCS 3058/4754 = 64.32%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
