# 批次 389 · 名称三元组索引读链（mmap + stream + 分派）+3 · FUNCS 3285

## 目标

落地 3 个名称三元组（trigram）索引读链函数，全部 `[S]`（新增，无存根升级）：

1. `loadVolumeNameTrigramIndexMappedContext` `[S]`（0x1407f8d60，1440B）— mmap 读链主入口。
2. `loadVolumeNameTrigramIndexContext` `[S]`（0x1407f9300，2112B）— 流式读链（含分块读取循环）。
3. `loadVolumeNameTrigramIndexIntoVolumeContext` `[S]`（0x1407f8940，1056B）— mmap/stream 分派 + 写入 idx.nameTrigram。
4. 新增哨兵 `errVolumeIndexFileNotFound`（"索引文件不存在"，.data @0x141bc3c60 实证）。
5. 新增分块常量 `volumeNameTrigramStreamChunkEntries=0x8000` / `volumeNameTrigramStreamChunkBytes=0x40000`。

依赖（均已落地）：`contextErr` / `shouldUseVolumeIndexMmapStaticProvider` / `openVolumeIndexMappedFile` /
`(*volumeIndexMappedFile).Map` / `(*volumeIndexMappedFile).Close` / `parseVolumeNameTrigramHeader` /
`volumeNameTrigramHeaderMatchesIndex` / `(*VolumeIndex).runtimeVersionSnapshot`。

## 基线 / 收口

| 指标 | 基线（batch 388 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3282 | 3285 |
| S | 1555 | 1558 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1648 | 1648 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1592 | 1595 |
| USABLE | 1593 | 1596 |
| TRUE（S+S-inline+S-sig） | 3240 | 3243 |
| missing | 345 | 342 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet -tags production ./backend` EXIT=0；
`go1.25.12 test -count=1 -p=1 -tags production ./backend` `ok changeme/backend`。

## G2 契约

```
FUNCS=3285  MARKED=3285  UNMARKED=0  S=1558  S-inline=37  S-eq=1  S-sig=1648  P=41  FAITHFUL=1595  USABLE=1596  TRUE=3243
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1558 + 37 + 1 + 1648 + 41 = 3285 = FUNCS`。
本批 +3 全新函数（S 1555→1558），FUNCS 3282→3285、TRUE 3240→3243、missing 345→342、
UNMARKED=0 / P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 loadVolumeNameTrigramIndexMappedContext [S 0x1407f8d60]

签名 `(ctx context.Context, path string, idx *VolumeIndex) (*volumeNameTrigramIndex, error)`
（ctx=rax:rbx、path=rcx:rdi、idx=rsi）。

链路（错误 → Close 后返回）：

- idx==nil（@0x1407f8d95 test rsi + @0x1407f9276 加载 errVolumeIndexNodeBytesInvalid）→ (nil, Y)。
- contextErr(ctx)（@0x1407f8dc6）非 nil → (nil, cerr)。
- os.Stat(path)（@0x1407f8de4）失败：errors.Is(os.ErrNotExist)（@0x1407f8e18，哨兵 @0x141bc3990）
  → (nil, errVolumeIndexFileNotFound@0x141bc3c60)，否则 (nil, err)。
- fi.Size()<64（@0x1407f8e6e jl）→ (nil, Y)。
- openVolumeIndexMappedFile（@0x1407f8e88）失败 → (nil, err)。
- m.Map(0, size, 0, size)（@0x1407f8ee4）失败 → Close + (nil, err)。
- contextErr(ctx)（@0x1407f8f1a）非 nil → Close + (nil, cerr)。
- parseVolumeNameTrigramHeader(mapped[:64])（@0x1407f8f4b）失败 → Close + (nil, err)。
- !volumeNameTrigramHeaderMatchesIndex（@0x1407f8fe0 je）→ Close + (nil, Y)。
- len(mapped)<64+EntryCount*8（@0x1407f902b jge 失败）→ Close + (nil, Y)。
- 否则 mapped[0x40 : 0x40+EntryCount*8] 作 signatureBytes（@0x1407f9053..90a4 切片头），
  newobject{volumeNameTrigramIndex}（@0x1407f90c3，类型 @0x140bc5e00），
  version=runtimeVersionSnapshot（@0x1407f90b6）、signatureBytes 三字 @+0x20/0x28/0x30、
  mappedFile=@+0x38（@0x1407f9123/9127）→ (result, nil)。

### 3.2 loadVolumeNameTrigramIndexContext [S 0x1407f9300]

签名同 mmap。链路：

- idx==nil → (nil, Y)；os.OpenFile(path, O_RDONLY, 0)（@0x1407f938c）失败 → (nil, err)；
- defer f.Close()（@0x1407f93a2，闭包 @0x1407f9b40 deferwrap1）；
- contextErr（@0x1407f93e0）→ (nil, cerr)；f.Stat()（@0x1407f93f6）失败 → (nil, err)；
- io.ReadAtLeast(f, headerBuf[:], 64)（@0x1407f944a）失败 → (nil, Y)；
- parseVolumeNameTrigramHeader（@0x1407f9468）失败 → (nil, err)；
- !matchesIndex（@0x1407f9500）→ (nil, Y)；
- fi.Size()<64+EntryCount*8（@0x1407f9558）→ (nil, Y)；
- signatures=make([]uint64, EntryCount)（@0x1407f9579）、buf=make([]byte, 0x40000)（@0x1407f9590）；
- 循环（@0x1407f9828）：contextErr 检查 → n=min(剩余, 0x8000)（@0x1407f986f..987c）→
  io.ReadAtLeast(f, buf[:0x40000], n*8) 失败 → (nil, Y)（@0x1407f9905）→
  内循环（@0x1407f9a61）逐条 `mov r10,[buf+j*8]` LE uint64 写入 signatures[i+j]；
- 结束构造 {version: runtimeVersionSnapshot（@0x1407f99ba）, signatures@+0x08/0x10/0x18} → (result, nil)。

### 3.3 loadVolumeNameTrigramIndexIntoVolumeContext [S 0x1407f8940]

签名 `(ctx context.Context, idx *VolumeIndex, path string) error`
（ctx=rax:rbx、idx=rcx、path=rdi:rsi）。

- idx==nil（@0x1407f8967 test rcx）→ nil；TrimSpace(path)==""（@0x1407f8997 TrimSpace + @0x1407f89a0）→ nil；
- shouldUseVolumeIndexMmapStaticProvider（@0x1407f89af）true →
  loadVolumeNameTrigramIndexMappedContext(ctx, path, idx)（@0x1407f89e1）；
- err!=nil 且非 os.ErrNotExist / errVolumeIndexFileNotFound / errVolumeIndexNodeBytesInvalid
  （@0x1407f89fe..8a86，哨兵 @0x141bc3990/@0x141bc3c60/@0x141bc3c50）→
  contextErr(ctx)!=nil 则返回 cerr（@0x1407f8ac0），否则 logFileSearchPerf（@0x1407f8b36）后
  loadVolumeNameTrigramIndexContext(ctx, path, idx) 兜底（@0x1407f8b60）；false → 直接 stream（@0x1407f8b6d）；
- 汇合（@0x1407f8b97）：err!=nil → 返回 err；result!=nil →
  nameTrigramMu.Lock(+0x500) + nameTrigram=result(+0x508) + Unlock（@0x1407f8c4f..8cd7）；返回 nil。

## G4 独立复核

- `backend/filesearch_index_windows.go`：新增 `errVolumeIndexFileNotFound`、两个分块常量、
  三个 `[S]` 函数（复用 `volumeIndexCheckpointHeaderSize=64`、`volumeIndexFormatVersion`、
  `openVolumeIndexMappedFile`/`Map`/`Close`、`parseVolumeNameTrigramHeader`（pinyin 文件）、
  `volumeNameTrigramHeaderMatchesIndex`、`runtimeVersionSnapshot`、`contextErr`、`logFileSearchPerf`）。
- 字段布局与 asm 对齐：`volumeNameTrigramIndex.version@0x00 / signatures@0x08,0x10,0x18 /
  signatureBytes@0x20,0x28,0x30 / mappedFile@0x38 / closed@0x40`（mmap 写 signatureBytes、
  stream 写 signatures，两种填充互斥，与 asm 两处 newobject 初始化一致）。
- `VolumeIndex.nameTrigramMu@0x500 / nameTrigram@0x508` 与 @0x1407f8c50/8cb5 对齐。
- 全部 `[S]`，逻辑与 asm 逐地址对齐（含错误哨兵、contextErr 时机、切片边界、分块循环）；
  `go build/vet/test` 全仓通过。

## 移交（本轮收尾）

3 新函数（均 [S]）+ 1 哨兵 + 2 常量。FUNCS 3285/4754 = 69.10%，FAITHFUL 1595/4754 = 33.55%，
TRUE 3243/4754 = 68.22%。P=41 持平、UNMARKED=0 保持。
下一批：`loadVolumeIndexWithPathsContext` 0x1407f86e0（trigram 读链最外层入口，608B，
依赖本批三函数 + `loadVolumeIndexCheckpointMappedViewContext`）；checkpoint 读链
`loadVolumeIndexCheckpointMappedViewContext` 0x1407fb100 及其 `decodeVolumeIndexNodesFromBytesWithNodeCount`；
`P=41 → [S]` 转换。
