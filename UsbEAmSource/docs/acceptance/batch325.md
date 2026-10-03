# 批次 325 · Firefox 根/书签 EOF/覆盖层鼠标/覆盖层关闭 +4（FUNCS 3062）

## 目标

落地 4 个短函数：`resolveFirefoxProfileRoot`（Firefox 根解析）、
`ensureBookmarkJSONEOF`（书签 JSON EOF 校验）、
`captureScreenshotOverlayMouseInput`（覆盖层鼠标捕获）、
`oledBlackoutService.dismissOverlayScreenLocked`（覆盖层关闭）。

## 基线 / 收口

| 指标 | 基线（batch 324 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3058 | 3062 |
| MARKED | 3058 | 3062 |
| S | 1481 | 1482 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1498 | 1501 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1518 | 1519 |
| USABLE | 1519 | 1520 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1519  FUNCS=3062  MARKED=3062  P=41  S-eq=1  S-inline=37  S-sig=1501  S=1482  USABLE=1520
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1482 + 37 + 1 + 1501 + 41 = 3062 = FUNCS`。
S 1481→1482（+1）、S-sig 1498→1501（+3）、FUNCS 3058→3062（+4）、FAITHFUL 1518→1519（+1）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 resolveFirefoxProfileRoot [S 0x14076c420, 192B]

UserHomeDir（丢弃）→ Getenv(7 字符) → TrimSpace → 空则 ""；否则 Join(env, 常量7, 常量7)。

### 3.2 ensureBookmarkJSONEOF [S-sig 0x140767d00, 256B]

json.Decoder.Decode → err==io.EOF→nil；err!=nil→fmt.Errorf；否则 panic。

### 3.3 captureScreenshotOverlayMouseInput [S-sig 0x140973ce0, 192B]

LazyProc.Call；nil/失败 → nil；否则 SetCapture 闭包返回释放函数。

### 3.4 oledBlackoutService.dismissOverlayScreenLocked [S-sig 0x14090d320, 256B]

TrimSpace(key) 空则 panic；overlay map 非空则 Hide；mapdelete → syncVisibleOverlayStateLocked。

## G4 独立复核

- `backend/bookmarks.go`：+resolveFirefoxProfileRoot [S] +ensureBookmarkJSONEOF [S-sig]（+os import）。
- `backend/screenshot_services.go`：+captureScreenshotOverlayMouseInput [S-sig]。
- `backend/oledblackout.go`：+dismissOverlayScreenLocked [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（1 [S] + 3 [S-sig]）。FUNCS 3062/4754 = 64.41%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
