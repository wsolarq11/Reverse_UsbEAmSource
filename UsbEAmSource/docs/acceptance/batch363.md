# 批次 363 · 工具栏状态规范化/配置比较交换/Chromium 根压缩/书签根名排序 +4（FUNCS 3208）

## 目标

落地 4 个函数：`normalizeScreenshotSelectionToolbarState`、
`launcherConfigStore.CompareAndSwap`、`compactChromiumRoots`、`sortedBookmarkRootNames`。

## 基线 / 收口

| 指标 | 基线（batch 362 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3204 | 3208 |
| MARKED | 3204 | 3208 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1621 | 1625 |
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
FAITHFUL=1541  FUNCS=3208  MARKED=3208  P=41  S-eq=1  S-inline=37  S-sig=1625  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1625 + 41 = 3208 = FUNCS`。
S-sig 1621→1625（+4）、FUNCS 3204→3208（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 normalizeScreenshotSelectionToolbarState [S-sig 0x1409ab460, 320B]

TrimSpace(name/desc) 空→默认；count clamp(1,18)。

### 3.2 launcherConfigStore.CompareAndSwap [S-sig 0x14089aec0, 416B]

CompareAndSwapPrepared → 大结构(0x126*8) 拷贝返回。

### 3.3 compactChromiumRoots [S-sig 0x140766a60, 448B]

makeslice → 遍历 roots TrimSpace 非空 → append。

### 3.4 sortedBookmarkRootNames [S-sig 0x140767e00, 448B]

map 空→nil；makeslice → 遍历收集 keys → sort.Strings。

## G4 独立复核

- `backend/screenshot_selection_toolbar_windows.go`：+normalizeScreenshotSelectionToolbarState [S-sig]。
- `backend/launcherconfig.go`：+launcherConfigStore.CompareAndSwap [S-sig]。
- `backend/bookmarks.go`：+compactChromiumRoots [S-sig] +sortedBookmarkRootNames [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3208/4754 = 67.48%。下一批：screenshot 域
screenshotSelectionToolbarBoundsForSelection / waitUntilReady。
