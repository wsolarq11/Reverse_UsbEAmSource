# 批次 309 · 滚动接缝优化/预览解码 +2（FUNCS 3013）

## 目标

落地 2 个短函数：`refineScreenshotScrollingChunkedAppendFromForSeam`（接缝优化）、
`screenshotNativePreviewWindow.decodeImage`（预览解码透传）。

## 基线 / 收口

| 指标 | 基线（batch 308 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3011 | 3013 |
| MARKED | 3011 | 3013 |
| S | 1459 | 1459 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1473 | 1475 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1496 | 1496 |
| USABLE | 1497 | 1497 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1496  FUNCS=3013  MARKED=3013  P=41  S-eq=1  S-inline=37  S-sig=1475  S=1459  USABLE=1497
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1459 + 37 + 1 + 1475 + 41 = 3013 = FUNCS`。
S-sig 1473→1475（+2）、FUNCS 3011→3013（+2）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 refineScreenshotScrollingChunkedAppendFromForSeam [S-sig 0x14099ff00, 192B]

nil/尺寸非法/边界条件则透传；否则 tailImage 取尾部，非空则 refineScreenshotScrollingAppendFromForSeam。

### 3.2 screenshotNativePreviewWindow.decodeImage [S-sig 0x14099dda0, 192B]

duffcopy 参数结构 → screenshotNativePreviewDecodeImage 透传。

## G4 独立复核

- `backend/screenshot_scroll_windows.go`：+refineScreenshotScrollingChunkedAppendFromForSeam
  +decodeImage [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

2 个函数落地（2 [S-sig]）。FUNCS 3013/4754 = 63.38%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
