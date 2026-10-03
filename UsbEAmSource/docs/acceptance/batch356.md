# 批次 356 · 工具栏消息设置/书签节点查找/暗接缝伪影/滚轮输入构建 +4（FUNCS 3180）

## 目标

落地 4 个函数：`screenshotSelectionToolbarWindowService.SetMessages`、
`findFirefoxBookmarkNodeIDByGUID`、`screenshotScrollingRowLooksDarkSeamArtifact`、
`buildScreenshotMouseWheelInputs`。

## 基线 / 收口

| 指标 | 基线（batch 355 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3176 | 3180 |
| MARKED | 3176 | 3180 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1593 | 1597 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1541 | 1541 |
| USABLE | 1542 | 1542 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1541  FUNCS=3180  MARKED=3180  P=41  S-eq=1  S-inline=37  S-sig=1597  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1597 + 41 = 3180 = FUNCS`。
S-sig 1593→1597（+4）、FUNCS 3176→3180（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 SetMessages [S-sig 0x1409a6780, 352B]

lock(+0x8) → normalizeScreenshotSelectionToolbarMessages → 写入 field(+0x68)。

### 3.2 findFirefoxBookmarkNodeIDByGUID [S-sig 0x14076eac0, 384B]

遍历 map → node.guid(+0x78/+0x80) TrimSpace → EqualFold 匹配 guid → 返回 (key, true)。

### 3.3 screenshotScrollingRowLooksDarkSeamArtifact [S-sig 0x1409a3a60, 384B]

AverageRowLuma(y/y±1) → 负→false；纯黑邻域判定 → 亮度阈值。

### 3.4 buildScreenshotMouseWheelInputs [S-sig 0x1409a3120, 384B]

delta==0→nil；dir=±120；count=clamp(ceil(|delta|/120),1,16)；makeslice 填充。

## G4 独立复核

- `backend/screenshot_selection_toolbar_windows.go`：+SetMessages [S-sig]。
- `backend/bookmarks.go`：+findFirefoxBookmarkNodeIDByGUID [S-sig]。
- `backend/screenshot_scroll_windows.go`：+screenshotScrollingRowLooksDarkSeamArtifact [S-sig]
  +buildScreenshotMouseWheelInputs [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3180/4754 = 66.89%。下一批：launcher/filesearch 域
launcherConfigStore.Replace / selectSearchBigramKeys。
