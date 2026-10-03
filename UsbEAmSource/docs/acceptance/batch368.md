# 批次 368 · 源路径判定/墓碑应用到堆快照/USN 跟随者路径/画布行读取 +4（FUNCS 3228）

## 目标

落地 4 个函数：`volumeIndexMappedReadProvider.IsSourcePath`、
`applyVolumeIndexTombstonesToHeapSnapshot`、`FileIndexService.usnFollowerPaths`、
`screenshotScrollingChunkedCanvas.Row`。

## 基线 / 收口

| 指标 | 基线（batch 367 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3224 | 3228 |
| MARKED | 3224 | 3228 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1641 | 1645 |
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
FAITHFUL=1541  FUNCS=3228  MARKED=3228  P=41  S-eq=1  S-inline=37  S-sig=1645  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1645 + 41 = 3228 = FUNCS`。
S-sig 1641→1645（+4）、FUNCS 3224→3228（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 IsSourcePath [S-sig 0x1407e48a0, 448B]

TrimSpace 空→false；lock → 源路径(+0xe8/+0xf0) → abs+Clean → EqualFold。

### 3.2 applyVolumeIndexTombstonesToHeapSnapshot [S-sig 0x1408073a0, 448B]

遍历 heap 快照 → tombstone map 命中 → 标记 flag(+0x16) bit2 → 返回 tombstone 计数。

### 3.3 usnFollowerPaths [S-sig 0x140818280, 448B]

volumeIndexPath×2 → 尾段匹配 → concatstring2 拼接。

### 3.4 Row [S-sig 0x14099f560, 448B]

nil/越界→fmt.Errorf；遍历 chunks 找覆盖 row 的 chunk → 计算行像素偏移返回。

## G4 独立复核

- `backend/filesearch_index_windows.go`：+IsSourcePath [S-sig] +applyVolumeIndexTombstonesToHeapSnapshot [S-sig]
  +usnFollowerPaths [S-sig]。
- `backend/screenshot_scroll_canvas_windows.go`：+screenshotScrollingChunkedCanvas.Row [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3228/4754 = 67.90%。下一批：filesearch 域
VolumeIndex.activeEntryCountLocked / VolumeIndex.overlayStatsLocked。
