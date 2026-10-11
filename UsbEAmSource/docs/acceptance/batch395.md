# 批次 395 验收 — writeVolumeIndexCheckpointHeapContext [P→S]

**批次**：395  
**日期**：2026-12-12  
**类型**：函数升级（P→S）  
**目标**：`writeVolumeIndexCheckpointHeapContext`（0x1407f4fa0，2112B）从 [P] 存根升级为 [S]，405 行 asm 实证。

## G1 编译门禁

```
go build ./backend  → EXIT=0
go vet ./backend    → EXIT=0（无新增）
```

## G2 契约门禁

```
node tools/count_funcs.js
→ FUNCS=3298, S=1570, S-inline=37, S-eq=1, S-sig=1648, P=42, UNMARKED=0
→ TRUE=3255/4754 = 68.47%
→ ΔP: 43→42 (−1)，ΔS: 1569→1570 (+1)
```

## G3 行为实证

| 步骤 | asm 偏移 | Go 对应 |
|---|---|---|
| contextErr 守卫 | @0x1407f5005 | `contextErr(ctx)` |
| nodes ParentIdx 校验 | @0x1407f502c..5083 | 逐节点 `p := int(nodes[i].ParentIdx); p < -1 \|\| p >= nodeCount` |
| MkdirAll(dir, 0o755) | @0x1407f5094 | `os.MkdirAll(dir, 0o755)` |
| CreateTemp | @0x1407f50b3 | `os.CreateTemp(dir, ".data-*.tmp")` |
| buildVolumeIndexCheckpointHeaderLocked | @0x1407f5147 | `buildVolumeIndexCheckpointHeaderLocked(v, nodeCount, sortedCount, namePoolSize)` |
| writeAllContext(header) | @0x1407f51b4 | `writeAllContext(ctx, tmp, header[:])` ，5 路错误清理 |
| writeVolumeIndexNodesContext | @0x1407f51e5 | `writeVolumeIndexNodesContext(ctx, tmp, nodes)` |
| writeVolumeIndexSortedIndicesContext | @0x1407f522a | 同上 |
| contextErr 重检 | @0x1407f5248 | `contextErr(ctx)` |
| writeAllContext(namePool) | @0x1407f528d | `writeAllContext(ctx, tmp, namePool)` |
| tmp.Close | @0x1407f52a6 | `tmp.Close()` |
| closeMappedReadProviderForReplaceLocked | @0x1407f52cc | `closeMappedReadProviderForReplaceLocked(v, path)` |
| replaceFileAtomically | @0x1407f52fa | `replaceFileAtomically(tmpName, path)` |
| 字段更新 Nodes=snapshot | @0x1407f5309.. | `v.Nodes = nodes` + 10 个字段 |
| v.activeCount=activeCount | @0x1407f53.. | `v.activeCount = activeCount` |
| deltaByFRN/deltaReplaced/deltaNodes=nil | @0x1407f53.. | 清零 |
| buildNamePrefixBucketCounts | @0x1407f5483 | `v.prefixBuckets = buildNamePrefixBucketCounts(nodes, namePool)` |
| clearNameTrigramIndexLocked | @0x1407f5509 | `v.clearNameTrigramIndexLocked()` |

5 路错误清理路径均对照 asm 还原：任一步失败则 Close+Remove 临时文件。

## G4 独立复核

- 所有依赖项（`buildVolumeIndexCheckpointHeaderLocked` [S]、`writeVolumeIndexNodesContext` [S]、`writeVolumeIndexSortedIndicesContext` [S]、`closeMappedReadProviderForReplaceLocked` [S-sig]、`clearNameTrigramIndexLocked` [S-sig]、`buildNamePrefixBucketCounts` [S]）均已落地且能编译。
- `deltaNodes` 字段在 `VolumeIndex` struct 中存在（`types_filesearch.go:459`），清零语义与 asm mov [v+0x540], 0 对齐。
- prefixBuckets 字段为值类型 `namePrefixBucketCounts`（`types_filesearch.go:448`），赋值语义与 asm rep movsq 0x82 qwords 拷贝对齐。

## 交付

- [S] 函数体写入 `backend/filesearch_index_windows.go:1395-1473`
- P=43→42，S=1569→1570
- UNMARKED=0 保持