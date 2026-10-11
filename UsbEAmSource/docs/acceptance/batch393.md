# Batch 393 — Checkpoint Write Chain 部分落地

**日期** 2025-01-16 | **文件** `backend/filesearch_index_windows.go` | **FUNCS** 3295 (UNMARKED=0)

## 落地函数

| 函数 | 标记 | 大小 | 类型 |
|------|------|------|------|
| `writeVolumeIndexNodesContext` | [S] | 1056B | 全量 asm→Go：24B LE IndexNode 分块序列化（0x2000×24=0x30000B 缓冲） |
| `writeVolumeIndexCheckpointSectionsContext` | [S] | 1024B | 全量 asm→Go：header+nodes+sorted+namePool 四段持久化，CreateTemp+replaceFileAtomically |
| `buildVolumeIndexCheckpointHeaderLocked` | [S] | 288B | S-sig→S 升级，签名修正为 `(v *VolumeIndex, entryCount,sortedCount,namePoolSize uint32) [64]byte` |
| `saveVolumeIndexCheckpointLockedContext` | [P] | 928B | 存根（编排器，分 sections/heap 两路） |
| `CheckpointSections` | [P] | 928B | 存根（`*volumeIndexMappedReadProvider` 方法） |

## 未落地

| 函数 | 原因 |
|------|------|
| `writeVolumeIndexCheckpointHeapContext` (2112B) | 体量极大（405 行 asm → ~200 行 Go，含 5 路错误清理、closeMappedReadProviderForReplaceLocked、replaceFileAtomically、buildNamePrefixBucketCounts、clearNameTrigramIndexLocked 等调用链），留待后续批次 |

## 计数变化

- **S**: 1564 → 1567 (+3)
- **S-sig**: 1648 → 1647 (−1：buildVolumeIndexCheckpointHeaderLocked 升级)
- **P**: 41 → 43 (+2)
- **总计**: 3293 → 3295 (+2)
- **三闸**: 编译 ✓ vet ✓ 测试 ✓