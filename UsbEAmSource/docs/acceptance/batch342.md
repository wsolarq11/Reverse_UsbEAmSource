# 批次 342 · 映射提供者关闭/overlay 统计/侧车锁 +4（FUNCS 3123）

## 目标

落地 4 个函数：`volumeIndexMappedReadProvider.Close`、`VolumeIndex.OverlayStats`、
`VolumeIndex.overlayStatsLocked`、`FileIndexService.lockUSNFollowerSidecar`。

## 基线 / 收口

| 指标 | 基线（batch 341 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3119 | 3123 |
| MARKED | 3119 | 3123 |
| S | 1500 | 1501 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1540 | 1543 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1537 | 1538 |
| USABLE | 1538 | 1539 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1538  FUNCS=3123  MARKED=3123  P=41  S-eq=1  S-inline=37  S-sig=1543  S=1501  USABLE=1539
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1501 + 37 + 1 + 1543 + 41 = 3123 = FUNCS`。
S 1500→1501（+1）、S-sig 1540→1543（+3）、FUNCS 3119→3123（+4）、FAITHFUL 1537→1538（+1）、
USABLE 1538→1539（+1）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 Close [S 0x1407e4300, 320B]

nil/closed→(nil)；closing=true；active>0→(nil)；否则取 release 置 nil + closed=true；
unlock 后 release 非空则调用。字段名与 `volumeIndexMappedReadProvider` 逐一对齐。

### 3.2 OverlayStats [S-sig 0x1407e7c20, 320B]

RLock(+0x10) → defer RUnlock → overlayStatsLocked。

### 3.3 overlayStatsLocked [S-sig 0x1407e7fe0, 320B]

baseEntryCountLocked → 迭代节点计数 → 比率计算。

### 3.4 lockUSNFollowerSidecar [S-sig 0x140818440, 320B]

usnFollowerPaths → HashTrieMap LoadOrStore → lock。

## G4 独立复核

- `backend/filesearch_index_windows.go`：+Close [S] +OverlayStats [S-sig]
  +overlayStatsLocked [S-sig] +lockUSNFollowerSidecar [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（1 [S] + 3 [S-sig]）。FUNCS 3123/4754 = 65.69%。下一批：filesearch 域
normalizeFileSearchUSNFollowerMeta / searchWithPathsContextMetrics。
