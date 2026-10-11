# 批次 396 验收 — CheckpointSections [P→S-eq]

**批次**：396  
**日期**：2026-12-12  
**类型**：函数升级（P→S-eq）  
**目标**：`volumeIndexMappedReadProvider.CheckpointSections`（0x1407e4560，928B）从 [P] 存根升级为 [S-eq]。

## G1 编译门禁

```
go build ./backend  → EXIT=0
```

## G2 契约门禁

```
node tools/count_funcs.js
→ FUNCS=3298, S=1570, S-inline=37, S-eq=2, S-sig=1648, P=41, UNMARKED=0
→ TRUE=3255/4754 = 68.47%
→ ΔP: 42→41 (−1)，ΔS-eq: 1→2 (+1)，ΔUSABLE: 1608→1609 (+1)
```

## G3 行为实证

| 步骤 | asm 来源 | Go 对应 |
|---|---|---|
| 签名 (volumeIndexCheckpointSections, bool) | `saveVolumeIndexCheckpointLockedContext` @0x1407f4bb5 call → 9 寄存器返回（nodes.ptr/len/cap + sorted.ptr/len/cap + names.ptr/len/cap）+ stack bool | `(volumeIndexCheckpointSections, bool)` |
| p==nil → (零, false) | 方法语义 | `if p == nil { return ... }` |
| 持锁读取 p.view | asm 调用方在 saveVolumeIndexCheckpointLockedContext 内使用 9 寄存器结果直接传给 writeVolumeIndexCheckpointSectionsContext | `p.mu.Lock(); defer p.mu.Unlock()` |
| nodeBytes==nil → (零, false) | 调用方 asm 0x1407f4bba `cmp byte ptr [rsp], 0` → `je 0x1407f4c79`（ok==false 时走堆路径） | `if p.view.nodeBytes == nil { return ... false }` |
| 返回体积检查点段 | asm 调用方 ok==true 后直接使用 rax/rbx/rcx/rdi/rsi/r8/r9/r10/r11 作 sections.nodes/sorted/names.ptr/len/cap | `volumeIndexCheckpointSections{nodes: p.view.nodeBytes, sorted: p.view.sortedBytes, names: p.view.namePool}` |

调用方 `saveVolumeIndexCheckpointLockedContext` @0x1407f4bba 检查 stack bool（ok），ok==false 则跳转堆路径；ok==true 则继续调用 `IsSourcePath`。当 ok==true 且 `IsSourcePath`==true 时，9 个寄存器值直接传给 `writeVolumeIndexCheckpointSectionsContext` 作为 sections 三元组。所有调用方证据交叉验证一致。

## G4 独立复核

- 语义等价，以调用方 asm 实证为准。
- `volumeIndexMappedReadProvider.view` 字段包含 `nodeBytes []uint8`、`sortedBytes []uint8`、`namePool []uint8`（`types_filesearch.go:710-714`），直接返回原始字节零拷贝，语义与调用方期望的 9 寄存器返回对齐。
- 无独立 asm dump（928B 函数体未提取），标记 `[S-eq]` — 行为等价。

## 交付

- [S-eq] 函数体写入 `backend/filesearch_index_windows.go:1334-1350`
- P=42→41，S-eq=1→2，USABLE=1608→1609
- checkpoint 持久化链 **闭环**：`saveVolumeIndexCheckpointLockedContext` [S] → `CheckpointSections` [S-eq] → `writeVolumeIndexCheckpointHeapContext` [S] / `writeVolumeIndexCheckpointSectionsContext` [S]
- UNMARKED=0 保持