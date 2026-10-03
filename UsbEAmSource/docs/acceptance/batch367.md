# 批次 367 · 读视图获取/节点派生判定/运行时快照/overlay 检查点判定 +4（FUNCS 3224）

## 目标

落地 4 个函数：`VolumeIndex.acquireReadView`、`volumeIndexBaseNodeDescendsFrom`、
`VolumeIndex.runtimeSnapshot`、`VolumeIndex.ShouldCheckpointOverlay`。

## 基线 / 收口

| 指标 | 基线（batch 366 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3220 | 3224 |
| MARKED | 3220 | 3224 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1637 | 1641 |
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
FAITHFUL=1541  FUNCS=3224  MARKED=3224  P=41  S-eq=1  S-inline=37  S-sig=1641  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1641 + 41 = 3224 = FUNCS`。
S-sig 1637→1641（+4）、FUNCS 3220→3224（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 acquireReadView [S-sig 0x1407e5ac0, 448B]

RLock(+0x10) → acquireReadViewLocked → 组装读视图。

### 3.2 volumeIndexBaseNodeDescendsFrom [S-sig 0x140804720, 448B]

遍历 nodeAtIndex → 父链匹配 nodeIndex → bool。

### 3.3 runtimeSnapshot [S-sig 0x1407e78e0, 448B]

RLock → activeEntryCountLocked → 字段(+0x90/+0x98/+0xa0/+0xa8/+0xb0/+0xb8/+0x80/+0x88) 拷贝。

### 3.4 ShouldCheckpointOverlay [S-sig 0x1407e7dc0, 448B]

RLock → overlayStatsLocked → 计数比较 → bool。

## G4 独立复核

- `backend/filesearch_index_windows.go`：+acquireReadView [S-sig] +volumeIndexBaseNodeDescendsFrom [S-sig]
  +runtimeSnapshot [S-sig] +ShouldCheckpointOverlay [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3224/4754 = 67.82%。下一批：filesearch 域
VolumeIndex.activeEntryCountLocked / VolumeIndex.overlayStatsLocked。
