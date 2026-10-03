# 批次 351 · 全量写上下文/持久化元数据/映射文件打开/元数据更新 +4（FUNCS 3160）

## 目标

落地 4 个函数：`writeAllContext`、`buildVolumeIndexPersistenceMetaLocked`、
`openVolumeIndexMappedFile`、`VolumeIndex.UpdateMeta`。

## 基线 / 收口

| 指标 | 基线（batch 350 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3156 | 3160 |
| MARKED | 3156 | 3160 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1573 | 1577 |
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
FAITHFUL=1541  FUNCS=3160  MARKED=3160  P=41  S-eq=1  S-inline=37  S-sig=1577  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1577 + 41 = 3160 = FUNCS`。
S-sig 1573→1577（+4）、FUNCS 3156→3160（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 writeAllContext [S-sig 0x1407f6420, 352B]

contextErr → writeFunc(+0x18) 循环写 → io.EOF。

### 3.2 buildVolumeIndexPersistenceMetaLocked [S-sig 0x1407f6580, 352B]

timeToUnixNano × 3 → 组装 meta{version,name,...}。

### 3.3 openVolumeIndexMappedFile [S-sig 0x1407fd500, 352B]

OpenFile → ErrNotExist→(nil,nil)；CreateFileMapping；handle==0→close。

### 3.4 UpdateMeta [S-sig 0x140807f40, 352B]

lock(+0x18) → RLock → 字段(+0x80/+0x88) 更新 → version(+0x590)++。

## G4 独立复核

- `backend/filesearch_index_windows.go`：+writeAllContext [S-sig]
  +buildVolumeIndexPersistenceMetaLocked [S-sig] +openVolumeIndexMappedFile [S-sig]
  +UpdateMeta [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3160/4754 = 66.47%。下一批：filesearch 域
shouldPersistUSNFollowerIdleMeta / inputMonitorService.Stop。
