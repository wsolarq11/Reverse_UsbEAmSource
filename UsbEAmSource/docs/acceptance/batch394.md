# Batch 394 — Checkpoint Write Chain: saveVolumeIndexCheckpointLockedContext [S] + writeVolumeIndexCheckpointHeapContext [P] 存根

**日期** 2025-01-17 | **文件** `backend/filesearch_index_windows.go` | **FUNCS** 3298 (UNMARKED=0)

## 落地函数

| 函数 | 标记 | 大小 | 类型 |
|------|------|------|------|
| `saveVolumeIndexCheckpointLockedContext` | [S] | 928B | P→S 升级：编排器，读提供者类型分派 + delta/Nodes 空检查 + sections→HeapSnapshot 两路决策。签名修正为 `(ctx context.Context, v *VolumeIndex, path string) error`，调用点为 `CheckpointSections()` / `IsSourcePath()` / `buildVolumeIndexCheckpointSnapshotLocked()`。asm 实证 175 行全量落地。 |
| `writeVolumeIndexCheckpointHeapContext` | [P] | 2112B | 新增存根：堆快照持久化（5 路错误清理 + closeMappedReadProviderForReplaceLocked + replaceFileAtomically + 11 字段更新 + clearNameTrigramIndexLocked + buildNamePrefixBucketCounts）。签名 `(ctx context.Context, v *VolumeIndex, path string, nodes []IndexNode, namePool []byte, sortedIndices []int32, activeCount int) error`。405 行 asm 待后续批次还原。 |
| `writeVolumeIndexNodesContext` | [S] | 1056B | 修复：编译损坏恢复（单行注释转义） |
| `writeVolumeIndexCheckpointSectionsContext` | [S] | 1024B | 修复：编译损坏恢复（单行注释转义） |

## 修复

- **编译损坏**：`filesearch_index_windows.go` 第 2476 行曾为单一巨型字符串（3861 字节），`writeVolumeIndexNodesContext` + `writeVolumeIndexCheckpointSectionsContext` 两个函数被压缩为 `\\n` 字面量注释行，导致 Go 编译器完全不可见。本批修复拆分为正常多行格式 + 删除重复损坏副本。

## 计数变化

- **S**: 1567 → 1569 (+2: saveVolumeIndexCheckpointLockedContext P→S, writeVolumeIndexCheckpointHeapContext 新增 P 计入总数)
- **S-sig**: 1648 → 1648 (→)
- **P**: 43 → 43 (→：saveVolumeIndexCheckpointLockedContext −1 P, writeVolumeIndexCheckpointHeapContext +1 P)
- **总计**: 3295 → 3298 (+3)
- **三闸**: build ✓ vet ✓ test ✓