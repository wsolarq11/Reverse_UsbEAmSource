# 批次 390 · checkpoint 读链节点流解析 +1 · FUNCS 3286

## 目标

落地 1 个 checkpoint 读链叶子函数 `[S]`（新增，无存根升级）：

1. `decodeVolumeIndexNodesFromBytesWithNodeCount` `[S]`（0x1407fcc00，768B）— 从序列化节点字节流解析 `[]IndexNode` 并统计活动条目数。

依赖（均已落地）：`IndexNode` 类型 / `errVolumeIndexNodeBytesInvalid` / `binary.LittleEndian`。

## 基线 / 收口

| 指标 | 基线（batch 389 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3285 | 3286 |
| S | 1558 | 1559 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1648 | 1648 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1595 | 1596 |
| USABLE | 1596 | 1597 |
| TRUE（S+S-inline+S-sig） | 3243 | 3244 |
| missing | 342 | 341 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet -tags production ./backend` EXIT=0；
`go1.25.12 test -count=1 -p=1 -tags production ./backend` `ok changeme/backend`。

## G2 契约

```
FUNCS=3286  MARKED=3286  UNMARKED=0  S=1559  S-inline=37  S-eq=1  S-sig=1648  P=41  FAITHFUL=1596  USABLE=1597  TRUE=3244
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1559 + 37 + 1 + 1648 + 41 = 3286 = FUNCS`。
本批 +1 全新函数（S 1558→1559），FUNCS 3285→3286、TRUE 3243→3244、missing 342→341、
UNMARKED=0 / P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 decodeVolumeIndexNodesFromBytesWithNodeCount [S 0x1407fcc00]

签名 `(nodeBytes []byte, nodeCount int) ([]IndexNode, int, error)`
（nodeBytes = ptr@rax / len@rbx / cap@rcx，nodeCount=rdi）。

链路（错误 → (nil, 0, 哨兵)）：

- `len%24!=0`（@0x1407fcc39 魔数 0xaaaaaaaaaaaaaaab 除 24 后 `cmp rbx,r9 jne`）→ 哨兵；
- `nodeCount < len/24`（@0x1407fcc40 `cmp rdi,rdx jl`，rdx=len/24）→ 哨兵；
- `makeslice([]IndexNode, derived, derived)`（@0x1407fcc69，derived=len/24）；
- 循环（@0x1407fcd0a..ce3f）逐 24B 解析（FRN@0 8B / NameOffset@8 4B / ParentIdx@12 4B /
  ModTime@16 4B / NameLen@20 2B / Flags@22 2B，均小端原生读）：
  - 越界护栏按 cap（@0x1407fcd22 `cmp rdx,r11 jb`，rdx=cap，与调用方 3 词切片传递一致）；
  - ParentIdx<-1（@0x1407fcde7 `cmp r9d,-1 jl`）或 ParentIdx>=nodeCount（@0x1407fce05
    `cmp esi,r9d jg`，esi=显式 nodeCount 低 32 位）→ 哨兵；
  - Flags&2==0（@0x1407fccf3 `test r10w,2`）计入 active 计数（r8 累加器）；
- 终返回 (nodes, active, nil)（@0x1407fce3f：rax=ptr、rbx=len、rcx=cap、rdi=r8=active、
  rsi/r8=0）。

## G4 独立复核

- `backend/filesearch_index_windows.go`：新增 `decodeVolumeIndexNodesFromBytesWithNodeCount`，
  字段布局与既有 `nodeAt`/`validateVolumeIndexNodeBytes` 完全一致（FRN/NameOffset/ParentIdx/
  ModTime/NameLen/Flags 24B 小端）。ParentIdx 校验沿用 `validateVolumeIndexNodeBytes` 同款
  `parentIdx < -1 || parentIdx >= int32(nodeCount)`，但上限用**显式 nodeCount 参数**（非 len/24
  派生值），与 asm `cmp esi,r9d` 一致。active 计数语义对齐 `IsDeleted()`（Flags&2）。
- `backend/filesearch_index_windows_test.go`：新增 `TestDecodeVolumeIndexNodesFromBytesWithNodeCount`
  覆盖 round-trip、空流（非 nil 空切片）、len%24!=0、nodeCount<derived、ParentIdx 越界两类。
- 全部 `[S]`，逻辑与 asm 逐地址对齐；`go build/vet/test` 全仓通过。

## 移交（本轮收尾）

1 新函数（[S]）。FUNCS 3286/4754 = 69.12%，FAITHFUL 1596/4754 = 33.57%，
TRUE 3244/4754 = 68.24%。P=41 持平、UNMARKED=0 保持。
下一批：`buildNamePrefixBucketCountsFromNodeBytes` 0x1407e1ee0 与
`newVolumeIndexReadViewFromCheckpointView` 0x1407e5260（checkpoint 读链上游依赖），
随后 `loadVolumeIndexCheckpointMappedViewContext` 0x1407fb100（含 goroutine 闭包
func1/gowrap1 与 `loadVolumeIndexBucketContext` 0x1407f6f00）；`P=41 → [S]` 转换。
