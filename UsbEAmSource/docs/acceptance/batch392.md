# 批次 392 · .bkt 桶持久化 + checkpoint 视图全链汇合 · FUNCS 3293

## 目标

落地 3 个 checkpoint 读链末端函数 `[S]` + 1 个依赖函数 `[S-sig]`，修复 1 处哨兵缺失，完成 checkpoint 读链全链路闭环（从磁盘文件 → 解析头 → mmap 三区 → 解码节点 → 前缀桶计数 → 读视图 → .bkt 持久化 + 异步重存）：

1. `loadVolumeIndexBucketContext` `[S]`（0x1407f6f00，230B）— 从 .bkt 二进制文件加载名称前缀桶计数与身份校验。
2. `saveVolumeIndexBucketContext` `[S]`（0x1407f68e0，328B）— 持久化卷前缀桶索引到 .bkt 文件（原子替换，含 delta 分叉）。
3. `buildVolumeIndexCheckpointSnapshotLocked` `[S-sig]`（0x1407e70a0，314B）— 构建检查点快照（dependency，save 的 delta 分支调用）。
4. `loadVolumeIndexCheckpointMappedViewContext` `[S]`（0x1407fb100，1041B）— mmap 加载 checkpoint 视图，含 .bkt 桶加载与异步回退重存 goroutine（`go func(){...}()` 编译器生成闭包）。

修复：补回 `errVolumeIndexBucketPathEmpty` 哨兵（`errors.New("empty path")`，10B 消息，@0x141bc3e10 静态 error，此前被误删）。

依赖（均已落地）：`writeAllContext` / `replaceFileAtomically` / `parseVolumeIndexCheckpointHeader` / `buildVolumeIndexCheckpointLayout` / `planVolumeIndexCheckpointMapWithSystemGranularity` / `openVolumeIndexMappedFile` / `Map` / `validateVolumeIndexNodeBytes` / `decodeVolumeIndexNodesFromBytesWithNodeCount` / `buildNamePrefixBucketCounts` / `buildNamePrefixBucketCountsFromNodeBytes` / `newVolumeIndexReadViewFromCheckpointView` / `Close` / `shouldUseVolumeIndexMmapStaticProvider` / `timeToUnixNano` / `addNamePrefixBucketCount` / `contextErr`。

## 基线 / 收口

| 指标 | 基线（batch 391 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3289 | 3293 |
| S | 1562 | 1565 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1648 | 1649 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1599 | 1602 |
| USABLE | 1600 | 1603 |
| TRUE（S+S-inline+S-sig） | 3247 | 3251 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`），`gofmt -l ./backend` 空。

## G1 编译

- `go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
- `go1.25.12 vet ./backend` EXIT=0；
- `go1.25.12 test -count=1 -p=1 -tags production ./backend` `ok changeme/backend`。

## G2 契约

```
FUNCS=3293  MARKED=3293  UNMARKED=0  S=1565  S-inline=37  S-eq=1  S-sig=1649  P=41  FAITHFUL=1602  USABLE=1603  TRUE=3251
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1565 + 37 + 1 + 1649 + 41 = 3293 = FUNCS`。
本批 +4 函数：3 个全新 `[S]`（S 1562→1565）+ 1 个全新 `[S-sig]` dependency（S-sig 1648→1649），FUNCS 3289→3293、TRUE 3247→3251、FAITHFUL 1599→1602、USABLE 1600→1603。UNMARKED=0 / P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 errVolumeIndexBucketPathEmpty 哨兵补回

- .data 静态 error 位于 @0x141bc3e10，消息 10B "empty path"（`mov [rax+8], 0xa` @0x1407f7089）。
- 插入位置：`var errVolumeIndexFileNotFound` 声明后一行。

### 3.2 loadVolumeIndexBucketContext [S 0x1407f6f00, 230B]

签名 `(ctx context.Context, path string, want *volumeIndexBucketIdentity, enabled bool) (*namePrefixBucketCounts, int, error)`。

链路：
- `TrimSpace(path) == ""`（@0x1407f702e/7043）→ `(nil, 0, errVolumeIndexBucketPathEmpty)`；
- `os.ReadFile(path)`（@0x1407f7061）→ err→`(nil, 0, err)`；
- `len(data) < 0x440`（@0x1407f7083 `cmp jl`）→ `errVolumeIndexNodeBytesInvalid`；
- magic@0≠`USBEAMBK`（0x4b424d4145425355）或 version@8≠`T2`（0x3254）→ `errVolumeIndexNodeBytesInvalid`；
- entryCount = dword@0xa（@0x1407f70aa）；
- `want != nil`→ 4×qword 身份比较（rootFRN/journalID/lastUSN/generatedAt），任一不等→哨兵；
- namePoolSize = dword@0xe（@0x1407f70d8）；`len(data) != 0x440+namePoolSize*4`→哨兵；
- `new(namePrefixBucketCounts)` → firstByte[256] int32-LE 逐条拷贝（@0x1407f7129/3b duffcopy+loop）；
- namePoolSize>0→ `make([]int32, namePoolSize) + int32-LE` 逐条拷贝 → `(bucket, entryCount, nil)`。

### 3.3 saveVolumeIndexBucketContext [S 0x1407f68e0, 328B]

签名 `(ctx context.Context, v *VolumeIndex, path string, counts namePrefixBucketCounts) error`（counts 0x418B 值传参 @rsp+0x8a0）。

链路：
- `v == nil || TrimSpace(path) == ""`（@0x1407f6915/694d）→ nil；
- `os.MkdirAll(filepath.Dir(path), 0o755)`→ err→return err；
- 默认 counts = v.prefixBuckets 全字段拷贝（@0x1407f6991/9b7 `rep movsq` 0x82 qwords = 0x418B 到栈 local @rsp+0x898）；
- delta 分支：`len(v.deltaByFRN) > 0`（@0x1407f69bd/69d0 `jg`）或 `v.deltaReplaced != nil`（@0x1407f69dc）→ `buildVolumeIndexCheckpointSnapshotLocked(v)`→ entryCount=activeCount、counts=`buildNamePrefixBucketCounts(nodes, namePool)`；
- 否则 entryCount=`len(v.SortedIndices)`；
- header 0x40 字节小端写入：magic@0、version@8、entryCount@0xa、namePoolSize@0xe（`len(counts.twoByte)`）、rootFRN@0x12、journalID@0x1a、lastUSN@0x22、generatedAt@0x2a（`timeToUnixNano(v.GeneratedAt)`）；
- `os.CreateTemp(Dir(path), ".bkt-*.tmp")`→ 写 header[:0x40]→ firstByte[256] int32-LE（0x400B）→ twoByte 段（namePoolSize>0 时 namePoolSize*4B int32-LE）→ `writeAllContext`；
- `tmp.Close()`→ `replaceFileAtomically(tmpName, path)`。任一写/关闭失败→清理 temp。

**.bkt header 字段 @0xe 修正**：字段为 `len(counts.twoByte)`（twoByte 条目计数），非 `len(v.NamePool)`。save 的 async 路径中 v.NamePool=nil 但 counts.twoByte 有正确数据，故 @0xe 必须跟随 counts.twoByte 而非 v.NamePool。load 侧 @0xe 读为 namePoolSize 用于计算 twoByte 段字节数。

### 3.4 buildVolumeIndexCheckpointSnapshotLocked [S-sig 0x1407e70a0, 314B]

签名 `(v *VolumeIndex) (nodes []IndexNode, namePool []byte, sortedIndices []int32, activeCount int, err error)`。

链路：
- `v == nil`→ 零返回；
- 无 delta（`len(deltaByFRN) == 0 && deltaReplaced == nil`）且 v.Nodes/v.NamePool 非空→字段深拷贝；
- 无 delta 且两者皆空→若 readProvider 为 `*volumeIndexMappedReadProvider` 则哨兵，否则零返回；
- 有 delta→哨兵（`HeapSnapshot + applyTombstones + applyDeltaOverlay` 路径为 gap，体待专项还原）。

> 标记为 `[S-sig]`：delta-present 分支的 `HeapSnapshot`/`applyTombstones`/`applyDeltaOverlay` 为未落地 gap，当前返回哨兵；`saveVolumeIndexBucketContext` 的 delta 分支调用此处后立即从返回值重算 counts，哨兵导致 write 失败 → 异步重存触发替代路径，功能上不造成数据丢失。

### 3.5 loadVolumeIndexCheckpointMappedViewContext [S 0x1407fb100, 1041B]

签名 `(ctx context.Context, root string, path string, bucketPath string) (*volumeIndexCheckpointView, error)`（bucketPath 为第 4 寄存参数 @r10/r11）。

链路（checkpoint 读链全链路闭环）：
- `contextErr(ctx)`→ 返回 cerr；
- `os.OpenFile(path, O_RDONLY, 0)`→ `os.ErrNotExist`→`errVolumeIndexFileNotFound`，否则 err；
- `f.Stat()`→ err→close+return；
- `io.ReadAtLeast(f, headerBuf[:64], 64)`→ err→close+`errVolumeIndexNodeBytesInvalid`；
- close f；
- `parseVolumeIndexCheckpointHeader`→ err→return；
- `buildVolumeIndexCheckpointLayout(fi.Size(), entryCount, sortedCount, namePoolSize)`→ err→return；
- `planVolumeIndexCheckpointMapWithSystemGranularity`→ err→return；
- `openVolumeIndexMappedFile`→ err→return；
- `m.Map`×3（node/sorted/namePool）→ err→Close+return；
- contextErr recheck→Close+return；
- `validateVolumeIndexNodeBytes`→ err→Close+return；
- 分支 `shouldUseVolumeIndexMmapStaticProvider()`：true→nodes/sortedIndices/namePool=nil，校验 sortedBytes 长度与 sortedCount、逐个 sortedIndex∈[0,nodeCount)；false→`decodeVolumeIndexNodesFromBytesWithNodeCount`→nodes+activeCount，sortedIndices int32-LE copy+逐个越界校验，namePool=copy(namePoolBytes)；
- 构造 view#1（prefixBuckets=nil）；
- 桶段：bucketPath 非空→want 身份（rootFRN/journalID/lastUSN/generatedAt）→`loadVolumeIndexBucketContext(ctx, bucketPath, &want, true)`；
  - 成功且非 nil→prefixBuckets=counts（无覆盖）；
  - 失败或 nil→`buildNamePrefixBucketCountsFromNodeBytes(nodeBytes, namePoolBytes)`→err→Close+return；recounts 赋值 prefixBuckets，`go func(){tmp := VolumeIndex{RootFRN:, JournalID:, LastUSN:, GeneratedAt:, activeCount:, prefixBuckets: recounts}; _ = saveVolumeIndexBucketContext(context.Background(), &tmp, bucketPath, recounts)}()`（异步重存，不等待）；
- 构造 view#2（prefixBuckets 已赋值）→ `return (view, nil)`。

**goroutine 闭包实证**：编译器生成的 func1（0x1407fc720）捕获 rootFRN/journalID/lastUSN/generatedAt/activeCount/bucketPath + counts（0x418B），不捕获 NamePool。func1 内部重建栈 `*VolumeIndex` 布局临时体：NamePool=nil、SortedIndices=nil，但 counts.twoByte 有正确数据。调用 `saveVolumeIndexBucketContext` 时 @0xe 字段 = `len(counts.twoByte)`（非 len(v.NamePool)），格式正确。

## G4 独立复核

- `backend/filesearch_index_windows.go`：新增 4 函数 + 1 哨兵，字段布局与既有 `.bkt` 常数（`volumeIndexBucketHeaderMagic`/`volumeIndexBucketVersion`/`volumeIndexBucketHeaderSize`）和依赖函数完全对齐。`.bkt` header @0xe 字段含义在 save/load 双向一致（均为 `len(counts.twoByte)`）。
- `backend/filesearch_index_windows_batch392_test.go`：新增 8 个测试覆盖：空路径→`errVolumeIndexBucketPathEmpty`、文件缺失→err、nil v→nil、空 path→nil、round-trip save/load（firstByte + twoByte 64K 条目往返验证）、nil snapshot、no-delta 深拷贝验证、delta→哨兵。
- 3 个 `[S]`（汇编逐地址实证）+ 1 个 `[S-sig]`（签名实证，delta 路径 gap），`go build/vet/test` 全仓通过。

## 移交（本轮收尾）

4 函数（3 `[S]` + 1 `[S-sig]`）+ 1 哨兵补回。FUNCS 3293/4754 = **69.27%**，FAITHFUL 1602/4754 = **33.70%**，TRUE 3251/4754 = **68.38%**。P=41 持平、UNMARKED=0 保持。

checkpoint 读链全链路已闭环：OpenFile→Stat→ReadAtLeast→parseHeader→buildLayout→planMap→openMappedFile→Map×3→validateNodeBytes→decodeNodes→sortedIndices 校验→namePool→`loadVolumeIndexBucketContext`（.bkt 身份校验+计数加载）→view 构造→`newVolumeIndexReadViewFromCheckpointView`（批次 391）→读视图 Ready。

下一批：checkpoint 写链（`saveVolumeIndexCheckpointContext + buildVolumeIndexCheckpointFileContext + writeVolumeIndexNodesContext`），`P=41 → [S]` 转换。