# 批次 307 · 滚动帧分析/画布保存/多路径搜索 +3（FUNCS 3008）

## 目标

落地 3 个短函数：`analyzeScreenshotScrollingFrameProgress`（帧进度分析）、
`stopScreenshotScrollingChunkedCanvasAndSave`（分块画布保存）、
`VolumeIndex.SearchWithPaths`（多路径搜索薄包装）。

## 基线 / 收口

| 指标 | 基线（batch 306 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3005 | 3008 |
| MARKED | 3005 | 3008 |
| S | 1456 | 1456 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1470 | 1473 |
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
FAITHFUL=1493  FUNCS=3008  MARKED=3008  P=41  S-eq=1  S-inline=37  S-sig=1473  S=1456  USABLE=1494
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1456 + 37 + 1 + 1473 + 41 = 3008 = FUNCS`。
S-sig 1470→1473（+3）、FUNCS 3005→3008（+3）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 analyzeScreenshotScrollingFrameProgress [S-sig 0x1409a41e0, 128B]

screenshotFramesAreSimilar(a,b) → bool，交 resolveScreenshotScrollingAppendMatchWithTarget(a,b,c)，
返回 (结果, 相似 bool)。

### 3.2 stopScreenshotScrollingChunkedCanvasAndSave [S-sig 0x1409a0740, 128B]

尺寸 (w,h) <=0 则返回空结果，否则 encodeScreenshotScrollingChunkedCanvas 编码。

### 3.3 VolumeIndex.SearchWithPaths [S-sig 0x1407eaa80, 128B]

薄包装：参数重排后 searchWithPathsContextMetrics(v, 0, 0, a, b, c, d, e)。

## G4 独立复核

- `backend/screenshot_scroll_windows.go`：+analyzeScreenshotScrollingFrameProgress
  +stopScreenshotScrollingChunkedCanvasAndSave [S-sig]。
- `backend/filesearch_index_windows.go`：+SearchWithPaths [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（3 [S-sig]）。FUNCS 3008/4754 = 63.27%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
