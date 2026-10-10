# 批次 385 · filesearch 映射读链（open + Map + 页对齐规划）+3（FUNCS 3279）

## 目标

落地 3 个 `[S]` 函数（1 新增 + 2 存根升级），打通卷索引 mmap 读链的下半段：

1. `openVolumeIndexMappedFile` `[S]`（0x1407fd500，352B）— `[S-sig]` 存根升级（OpenFile + CreateFileMapping）。
2. `(*volumeIndexMappedFile).Map` `[S]`（0x1407fd660，544B）— 新增（MapViewOfFile + unsafe.Slice 切片构造）。
3. `planVolumeIndexMappedSection` `[S]`（0x1407fcf00，224B）— `[S-sig]` 存根升级（页对齐算术，返回 error 修正 string）。

## 基线 / 收口

| 指标 | 基线（batch 384 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3278 | 3279 |
| MARKED | 3278 | 3279 |
| S | 1546 | 1549 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1653 | 1651 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1583 | 1586 |
| USABLE | 1584 | 1587 |
| TRUE（S+S-inline+S-sig） | 3236 | 3237 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。
`go fmt ./backend` 干净（无残留未格式化文件）。

## G2 契约

```
FUNCS=3279  MARKED=3279  UNMARKED=0  S=1549  S-inline=37  S-eq=1  S-sig=1651  P=41  FAITHFUL=1586  USABLE=1587  TRUE=3237
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1549 + 37 + 1 + 1651 + 41 = 3279 = FUNCS`。
FUNCS 3278→3279（+1，Map 新增）、S 1546→1549（+3）、S-sig 1653→1651（−2，两存根升级）、
TRUE 3236→3237（+1）、UNMARKED=0/P=41 保持。`list_missing.js` 未落地数 349→348（−1，Map 退出缺失清单）。

## G3 行为（asm 逐地址实证）

### 3.1 openVolumeIndexMappedFile [S 0x1407fd500]

签名 `(path string) (*volumeIndexMappedFile, error)`。`os.OpenFile(path, O_RDONLY, 0)`
（@0x1407fd520，rcx=0=O_RDONLY、rdi=0）；err!=nil 且 `errors.Is(err, os.ErrNotExist)`
（@0x1407fd607，目标全局 @0x141bc3c50→os.ErrNotExist）→ (nil,nil)，否则 (nil,err)；
`f.Fd()`（@0x1407fd542..554 经 DisassociateIOCP 取 sysfd @+0x10）；
`windows.CreateFileMapping(handle, nil, 2=PAGE_READONLY, 0, 0, nil)`（@0x1407fd569，rcx=2）；
err!=nil → `f.Close()`（@0x1407fd58b os.file.close）→ (nil,err)；否则 newobject
（@0x1407fd5ae）→ m.file=f、m.mapping=handle（@0x1407fd5d0/5d8）→ (m, nil)。

### 3.2 volumeIndexMappedFile.Map [S 0x1407fd660]

签名 `(m, mapOffset, mapBytes, sectionOffset, sectionBytes int64) ([]byte, error)`。

- m==nil 或 m.mapping==0（@0x1407fd672/680）→ errVolumeIndexNodeBytesInvalid（@0x1407fd814，全局 @0x141bc3c50）。
- mapBytes<=0 或 sectionBytes<=0（@0x1407fd689/68e）→ (nil,nil)（@0x1407fd693）。
- mapOffset<0 或 sectionOffset<0（@0x1407fd6a5/6aa）→ errVolumeIndexNodeBytesInvalid（@0x1407fd6af）。
- `windows.MapViewOfFile(mapping, 4=FILE_MAP_READ, mapOffset>>32, mapOffset低32, mapBytes)`
  （@0x1407fd6f3，参数拆分 @0x1407fd6de..6eb）；err→(nil,err)（@0x1407fd801）、
  addr==0→errVolumeIndexNodeBytesInvalid（@0x1407fd7e6）。
- views 追加 addr（@0x1407fd77c，Close 逆序 Unmap 用）。
- mapBytes < sectionOffset+sectionBytes（@0x1407fd7a8 jge 失败）→ errVolumeIndexNodeBytesInvalid（@0x1407fd7ad）。
- 成功返回 unsafe.Slice(addr, mapBytes)[sectionOffset : sectionOffset+sectionBytes : mapBytes]
  （@0x1407fd7d8 ptr=base+sectionOffset、len=sectionBytes、cap=mapBytes-sectionOffset）。

调用点实证（扫描 .text 全量 call）：trigram 链 @0x1407f8ee4 传 (0,size,0,size)；
checkpoint 链 @0x1407fb525/573/5c9 传 3 个 `volumeIndexMappedSectionPlan`（各 4×int64）。
`planVolumeIndexMappedSection` 输出 (alignedOffset, alignedSize, inPageOffset, size)，
Map 的 mapOffset=alignedOffset、mapBytes=alignedSize、sectionOffset=inPageOffset、
sectionBytes=size，切片 len=size、cap=alignedSize-inPageOffset=size，恒自洽。

### 3.3 planVolumeIndexMappedSection [S 0x1407fcf00]

签名 `(offset, size, granularity, limit int64) (alignedOffset, alignedSize, inPageOffset, sectionSize int64, error)`。

- offset<0 || size<0 || granularity<=0 || limit<0（@0x1407fcf00..f12）→ errVolumeIndexNodeBytesInvalid。
- size==0（@0x1407fcf2e je）→ (floor(offset/g)*g, 0, offset−align, 0, nil)（@0x1407fcfa3）。
- size!=0：offset+size 溢出（@0x1407fcf37 jg）或 limit<offset+size（@0x1407fcf40）→ 哨兵；
  alignedOffset=floor(offset/granularity)*granularity（@0x1407fcf67 idiv）；
  inPageOffset=offset−alignedOffset（@0x1407fcf6e）；alignedSize=size+inPageOffset（@0x1407fcf71）；
  inPageOffset>alignedSize 溢出（@0x1407fcf75）→ 哨兵；
  返回 (alignedOffset, alignedSize, inPageOffset, size, nil)（@0x1407fcf94..9f）。

## G4 独立复核

- `backend/filesearch_index_windows.go`：新增 `Map` + 升级 `openVolumeIndexMappedFile`/
  `planVolumeIndexMappedSection`；新增 import `unsafe`；复用已落地 `volumeIndexMappedFile`
  结构（types_filesearch.go，file/mapping/views 字段布局与 asm 逐地址对齐）、
  `errVolumeIndexNodeBytesInvalid`（"索引文件损坏或版本不兼容"，全局 @0x141bc3c50 实证）。
- `planVolumeIndexMappedSection` 返回类型由 string 修正为 error（无调用者，安全改签名）。
- 全部 `[S]` 标记，逻辑与 asm 逐地址对齐（含 MapViewOfFile/CreateFileMapping 参数寄存器映射、
  idiv 页对齐、unsafe.Slice 三索引切片构造）；`go build/vet/test` 全仓通过。

## 移交（本轮收尾）

3 函数落地（均 [S]）。FUNCS 3279/4754 = 68.97%，FAITHFUL 1586/4754 = 33.36%，
TRUE 3237/4754 = 68.09%。未落地文件差集仍 30、P=41 持平、UNMARKED=0 保持。
下一批：`planVolumeIndexCheckpointMap` 0x1407fcfe0 / `planVolumeIndexCheckpointMapWithSystemGranularity`
0x1407fd2c0（依赖已落地的 planVolumeIndexMappedSection + GetSystemInfo 分配粒度）；
`loadVolumeNameTrigramIndexMappedContext` 0x1407f8d60（依赖 parseVolumeNameTrigramHeader
0x1407f9ba0 / volumeNameTrigramHeaderMatchesIndex 0x1407f9ca0）；`P=41 → [S]` 转换。
