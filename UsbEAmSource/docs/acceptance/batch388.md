# 批次 388 · 三字索引头解析 + 匹配校验（S-sig→S 升级，FUNCS 3282）

## 目标

将 2 个 `[S-sig]` 存根升级为 `[S]`，打通 trigram 读链的头部解析与身份匹配：

1. `parseVolumeNameTrigramHeader` `[S]`（0x1407f9ba0，640B）— 存根升级（64B 头解析 + 校验）。
2. `volumeNameTrigramHeaderMatchesIndex` `[S]`（0x1407f9ca0，448B）— 存根升级（8 字段逐项校验）。
3. 新增结构体 `volumeNameTrigramHeader`（types_filesearch.go）。

## 基线 / 收口

| 指标 | 基线（batch 387 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3282 | 3282 |
| S | 1553 | 1555 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1650 | 1648 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1590 | 1592 |
| USABLE | 1591 | 1593 |
| TRUE（S+S-inline+S-sig） | 3240 | 3240 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。
`go fmt ./backend` 干净。

## G2 契约

```
FUNCS=3282  MARKED=3282  UNMARKED=0  S=1555  S-inline=37  S-eq=1  S-sig=1648  P=41  FAITHFUL=1592  USABLE=1593  TRUE=3240
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1555 + 37 + 1 + 1648 + 41 = 3282 = FUNCS`。
本批为质量升级（无新增函数）：S 1553→1555（+2，两存根升级）、S-sig 1650→1648（−2）、
FUNCS/TRUE/missing 持平、FAITHFUL 1590→1592（+2）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 parseVolumeNameTrigramHeader [S 0x1407f9ba0]

签名 `(b []byte) (volumeNameTrigramHeader, error)`（rax=ptr、rbx=len、rcx=cap）。

二进制布局（64B，LE）：magic "UITG"(+0)、version 1(+4)、IndexVersion(+8)、EntryCount(+0xc)、
RootFRN(+0x10)、JournalID(+0x18)、LastUSN(+0x20)、GeneratedAt.Unix()(+0x28)、
LastMutationAt.UnixNano()(+0x30)、保留(+0x38)。

校验：len<64（@0x1407f9bc0 jl）、magic≠"UITG"（@0x1407f9bca，0x47544955）、version≠1
（@0x1407f9bd2）→ errVolumeIndexNodeBytesInvalid（逐路径 rip 验证 @0x141bc3c50）；
否则字段直拷（@0x1407f9bfd..c34，IndexVersion/EntryCount 存 eax/ebx，RootFRN 存 rcx，
JournalID/LastUSN/GeneratedAt/LastMutationAt 保留 rdi/rsi/r8/r9）。

与已落地 `buildVolumeNameTrigramHeaderLocked` 双向一致（写读同布局）。

### 3.2 volumeNameTrigramHeaderMatchesIndex [S 0x1407f9ca0]

签名 `(header volumeNameTrigramHeader, idx *VolumeIndex) bool`（header 8 字段拆寄存器
rax..r9：IndexVersion=eax、EntryCount=ebx、RootFRN=rcx、JournalID=rdi、LastUSN=rsi、
GeneratedAt=r8、LastMutationAt=r9；idx=r10 @0x1407f9cd3 test）。

校验顺序（任一失败 → false）：
- idx==nil（@0x1407f9cd6 je）；
- IndexVersion≠0x20003（@0x1407f9cdc cmp）；
- baseEntryCountLocked()≠int64(EntryCount)（@0x1407f9cf2 call + @0x1407f9d00 cmp）；
- idx.RootFRN(+0x30)≠RootFRN（@0x1407f9d13）；
- idx.JournalID(+0x80)≠（@0x1407f9d22）/ idx.LastUSN(+0x88)≠（@0x1407f9d30）；
- idx.GeneratedAt.Unix()≠GeneratedAt（@0x1407f9d80，.Unix() wall/ext 分派内联 @0x1407f9d49..77：
  wall 单调位 bt+add rdx,rdx+shr 0x1f+59453308800 或 ext 直取，再 −62135596800）；
- 否则 timeToUnixNano(idx.LastMutationAt)==LastMutationAt（@0x1407f9da4 call + @0x1407f9dae sete）。

## G4 独立复核

- `backend/types_filesearch.go`：新增 `volumeNameTrigramHeader`（7 字段，与 asm 返回寄存器
  逐一对应：IndexVersion/EntryCount uint32、RootFRN/JournalID uint64、LastUSN/GeneratedAt/
  LastMutationAt int64）。
- `backend/filesearch_pinyin_windows.go`：`parseVolumeNameTrigramHeader` 存根升级（复用
  `volumeIndexCheckpointHeaderSize`/`volumeNameTrigramHeaderMagic`/`volumeNameTrigramHeaderVersion`
  /`errVolumeIndexNodeBytesInvalid` 同包符号）。
- `backend/filesearch_index_windows.go`：`volumeNameTrigramHeaderMatchesIndex` 存根升级
  （复用 `volumeIndexFormatVersion`/`timeToUnixNano`；`VolumeIndex` 字段 RootFRN@+0x30、
  JournalID@+0x80、LastUSN@+0x88、GeneratedAt@+0x90、LastMutationAt@+0xa8 与 asm 逐地址对齐）。
- 全部 `[S]` 标记，逻辑与 asm 逐地址对齐（含 time.Time wall/ext 分派内联、LE 布局）；
  `go build/vet/test` 全仓通过。

## 移交（本轮收尾）

2 存根升级（均 [S]）+ 1 结构体新增。FUNCS 3282/4754 = 69.04%（本批持平），
FAITHFUL 1592/4754 = 33.49%，TRUE 3240/4754 = 68.15%。P=41 持平、UNMARKED=0 保持。
下一批：`loadVolumeIndexCheckpointMappedViewContext` 0x1407fb100（全链已实证，依赖本批与
b385/b386/b387 全链函数）；`validateVolumeIndexNodeBytes` / `parseVolumeIndexCheckpointNodes`
（节点流解析）；`loadVolumeNameTrigramIndexMappedContext` 0x1407f8d60（trigram 读链主入口，
依赖本批两函数）；`P=41 → [S]` 转换。
