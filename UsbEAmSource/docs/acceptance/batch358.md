# 批次 358 · 读提供者关闭/三字索引清除/checkpoint 布局/删除应用 +4（FUNCS 3188）

## 目标

落地 4 个函数：`closeVolumeIndexReadProvider`、`VolumeIndex.clearNameTrigramIndexLocked`、
`buildVolumeIndexCheckpointLayout`、`VolumeIndex.ApplyDelete`。

## 基线 / 收口

| 指标 | 基线（batch 357 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3184 | 3188 |
| MARKED | 3184 | 3188 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1601 | 1605 |
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
FAITHFUL=1541  FUNCS=3188  MARKED=3188  P=41  S-eq=1  S-inline=37  S-sig=1605  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1605 + 41 = 3188 = FUNCS`。
S-sig 1601→1605（+4）、FUNCS 3184→3188（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 closeVolumeIndexReadProvider [S-sig 0x1407e6f20, 384B]

provider(+0x520/+0x528) nil→nil；查找 → 回调 close → 置零。

### 3.2 clearNameTrigramIndexLocked [S-sig 0x1407e9120, 384B]

lock(+0x500/+0x510) → 字段(+0x508/+0x518) 置零 → trigramIndex.close。

### 3.3 buildVolumeIndexCheckpointLayout [S-sig 0x1407fc820, 384B]

count*24 + other*4 + 0x40 溢出检查 → layout。

### 3.4 ApplyDelete [S-sig 0x140802ea0, 384B]

lock(+0x18) → RLock → canApplyDeleteLocked → applyDeleteLocked>0 → markDirtyLocked。

## G4 独立复核

- `backend/filesearch_index_windows.go`：+closeVolumeIndexReadProvider [S-sig]
  +clearNameTrigramIndexLocked [S-sig] +buildVolumeIndexCheckpointLayout [S-sig]
  +ApplyDelete [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3188/4754 = 67.06%。下一批：filesearch 域
baseSubtreeActiveCountsLocked / ResolvePathByNodeIndex。
