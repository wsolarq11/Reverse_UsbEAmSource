# 批次 318 · mmap load 判定/映射段规划/排序索引重建 +3（FUNCS 3040）

## 目标

落地 3 个短函数：`shouldUseVolumeIndexMmapLoad`（mmap load 判定）、
`planVolumeIndexMappedSection`（映射段规划）、
`rebuildVolumeIndexSortedIndices`（排序索引重建）。

## 基线 / 收口

| 指标 | 基线（batch 317 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3037 | 3040 |
| MARKED | 3037 | 3040 |
| S | 1472 | 1473 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1486 | 1488 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1509 | 1510 |
| USABLE | 1510 | 1511 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1510  FUNCS=3040  MARKED=3040  P=41  S-eq=1  S-inline=37  S-sig=1488  S=1473  USABLE=1511
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1473 + 37 + 1 + 1488 + 41 = 3040 = FUNCS`。
S 1472→1473（+1）、S-sig 1486→1488（+2）、FUNCS 3037→3040（+3）、FAITHFUL 1509→1510（+1）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 shouldUseVolumeIndexMmapLoad [S 0x1407fa0e0, 256B]

Getenv(28 字符名) → TrimSpace → ToLower → "0"/"no"/"off"/"false"/"disable"/"disabled" → false；
其他（含空）→ true。

### 3.2 planVolumeIndexMappedSection [S-sig 0x1407fcf00, 224B]

参数合法性/溢出检查 → 页对齐起始 → 4 int64 + 错误串返回。

### 3.3 rebuildVolumeIndexSortedIndices [S-sig 0x140807dc0, 256B]

makeslice([]int32,n) → 递增填充 → sort.Slice（比较闭包）。

## G4 独立复核

- `backend/filesearch_index_windows.go`：+shouldUseVolumeIndexMmapLoad [S]
  +planVolumeIndexMappedSection +rebuildVolumeIndexSortedIndices [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（1 [S] + 2 [S-sig]）。FUNCS 3040/4754 = 63.95%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
