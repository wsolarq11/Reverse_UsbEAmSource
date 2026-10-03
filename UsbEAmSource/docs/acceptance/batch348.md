# 批次 348 · 滚动画布裁剪/像素混合/接缝修复/纯黑裁剪 +4（FUNCS 3148）

## 目标

落地 4 个函数：`screenshotScrollingChunkedCanvas.trimBottom`、`blendScreenshotScrollingRow`、
`repairScreenshotScrollingSeamArtifacts`、`trimScreenshotScrollingAppendStart`。

## 基线 / 收口

| 指标 | 基线（batch 347 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3144 | 3148 |
| MARKED | 3144 | 3148 |
| S | 1503 | 1503 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1562 | 1566 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1540 | 1540 |
| USABLE | 1541 | 1541 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1540  FUNCS=3148  MARKED=3148  P=41  S-eq=1  S-inline=37  S-sig=1566  S=1503  USABLE=1541
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1503 + 37 + 1 + 1566 + 41 = 3148 = FUNCS`。
S-sig 1562→1566（+4）、FUNCS 3144→3148（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 trimBottom [S-sig 0x14099fdc0, 320B]

height clamp → 循环 chunk 释放/收缩（releaseMemory）。

### 3.2 blendScreenshotScrollingRow [S-sig 0x1409a3d60, 320B]

遍历 x → RGBAAt(x-1,y)/RGBAAt(x+1,y) → 均值 → SetRGBA(x,y)。

### 3.3 repairScreenshotScrollingSeamArtifacts [S-sig 0x1409a3820, 320B]

y∈[n-2,n+2] → screenshotScrollingRowLooksDarkSeamArtifact → blendScreenshotScrollingRow。

### 3.4 trimScreenshotScrollingAppendStart [S-sig 0x1409a3ea0, 320B]

avail clamp → 遍历 screenshotScrollingRowLooksPureBlack 跳过。

## G4 独立复核

- `backend/screenshot_scroll_windows.go`：+trimBottom [S-sig] +blendScreenshotScrollingRow [S-sig]
  +repairScreenshotScrollingSeamArtifacts [S-sig] +trimScreenshotScrollingAppendStart [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3148/4754 = 66.22%。下一批：screenshot 域
normalizeScreenshotSelectionToolbarState / 其余 352B 候选。
