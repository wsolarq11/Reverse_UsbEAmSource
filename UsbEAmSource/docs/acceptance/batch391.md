# 批次 391 · checkpoint 读链前缀桶计数 +3 · FUNCS 3289

## 目标

落地 3 个 checkpoint 读链前缀桶计数函数 `[S]`（新增，无存根升级），并修正 1 处既有 `[S]` 函数的行为偏差：

1. `buildNamePrefixBucketCounts` `[S]`（0x1407e1c80，608B）— 从已解析 `[]IndexNode` 构建名字前缀桶计数（in-memory 版）。
2. `buildNamePrefixBucketCountsFromNodeBytes` `[S]`（0x1407e1ee0，1344B）— 从序列化节点字节流构建前缀桶计数并统计 active 数。
3. `newVolumeIndexReadViewFromCheckpointView` `[S]`（0x1407e5260，1120B）— 从 checkpoint 视图构建读视图（prefixBuckets 复用或惰性构建）。

修正：`addNamePrefixBucketCount`（0x1407e2620）补 `twoByte.len==0` 空检查。

依赖（均已落地）：`IndexNode` / `namePrefixBucketCounts` / `volumeIndexCheckpointView` / `volumeIndexReadView` / `errVolumeIndexNodeBytesInvalid` / `binary.LittleEndian`。

## 基线 / 收口

| 指标 | 基线（batch 390 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3286 | 3289 |
| S | 1559 | 1562 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1648 | 1648 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1596 | 1599 |
| USABLE | 1597 | 1600 |
| TRUE（S+S-inline+S-sig） | 3244 | 3247 |
| missing | 341 | 338 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`），`gofmt -l ./backend` 空。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；
`go1.25.12 test -count=1 -p=1 -tags production ./backend` `ok changeme/backend`。

## G2 契约

```
FUNCS=3289  MARKED=3289  UNMARKED=0  S=1562  S-inline=37  S-eq=1  S-sig=1648  P=41  FAITHFUL=1599  USABLE=1600  TRUE=3247
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1562 + 37 + 1 + 1648 + 41 = 3289 = FUNCS`。
本批 +3 全新函数（S 1559→1562），FUNCS 3286→3289、TRUE 3244→3247、missing 341→338、
UNMARKED=0 / P=41 保持。`addNamePrefixBucketCount` 为函数体修正，不改计数。

## G3 行为（asm 逐地址实证）

### 3.1 addNamePrefixBucketCount 修正（0x1407e2620）

原实现缺 `twoByte.len==0` 空检查，twoByte 为 nil 时会在 `len(name)>=2` 分支 panic。

asm 实证（@0x1407e2655..0x1407e2663）：

```
0x1407e2655: mov rcx, [rax+0x408]   ; twoByte.len（+0x400 为 slice header 起始）
0x1407e2660: test rcx, rcx
0x1407e2663: jne 0x1407e266b        ; twoByte.len != 0 才做 bigram
0x1407e2665: ret                    ; twoByte.len == 0 直接返回（跳过 bigram）
```

修正后：`len(name) < 2 || len(c.twoByte) == 0` 时返回，否则 `twoByte[(c0<<8)|c1] += count`。

### 3.2 buildNamePrefixBucketCounts [S 0x1407e1c80]

签名 `(nodes []IndexNode, namePool []byte) namePrefixBucketCounts`
（nodes=rax/ptr rbx/len rcx/cap，namePool=rdi/ptr rsi/len r8/cap）。

链路：

- `len(nodes)>0x2710`（@0x1407e1cf1 `cmp rbx,0x2710 jle`）→ 预分配
  `twoByte = make([]int32, 0x10000, 0x10000)`（@0x1407e1d09 makeslice）；
- 循环（@0x1407e1d55 `cmp rbx,rax jle`）逐节点（步长 24B = rax*3*8），
  `Flags&2`（@0x1407e1d82 `test r9w,2 jne`）跳过 deleted；
- 否则 `namePool[NameOffset:NameOffset+NameLen]`（@0x1407e1da3 `jb panicSliceAcap`，
  cap 界）→ `addNamePrefixBucketCount(&counts, name, 1)`（@0x1407e1de0 call）。
- 无 error 返回（in-memory 节点越界直接 panic，非哨兵）。

### 3.3 buildNamePrefixBucketCountsFromNodeBytes [S 0x1407e1ee0]

签名 `(nodeBytes []byte, namePool []byte) (namePrefixBucketCounts, int, error)`
（nodeBytes=rax/ptr rbx/len rcx/cap，namePool=rdi/ptr rsi/len r8/cap）。

链路：

- `len%24!=0`（@0x1407e1f6a 魔数 0xaaaaaaaaaaaaaaab 除 24 后 `cmp rbx,r12 jne`）
  → `(零,0,errVolumeIndexNodeBytesInvalid)`；
- `derived = len/24`；`derived>0x2710`（@0x1407e1fa2）→ 预分配 twoByte；
- 逐 24B 解析（FRN@0 8B / NameOffset@8 4B / ParentIdx@12 4B / ModTime@16 4B /
  NameLen@20 2B / Flags@22 2B，小端原生读）；
- `ParentIdx<-1`（@0x1407e2146 `cmp r8d,-1`）或 `ParentIdx>=derived`
  （@0x1407e2165 `cmp r8d,r11d`）→ 哨兵；
- `Flags&2`（@0x1407e218f `test r8w,2`）跳过；
- `NameOffset+NameLen > len(namePool)`（@0x1407e21b1 `cmp r8,rcx jb`）→ 哨兵；
- `addNamePrefixBucketCount(&counts, name, 1)`（@0x1407e2200 call），active 累加器
  （rdx）递增；终返回 `(counts, active, nil)`。

### 3.4 newVolumeIndexReadViewFromCheckpointView [S 0x1407e5260]

签名 `(cv *volumeIndexCheckpointView) volumeIndexReadView`。

链路：

- `cv==nil`（@0x1407e52a0 `test rax,rax je`）→ 零值；
- `prefixBuckets!=nil`（@0x1407e52a9 `cmp [rax+0x120],0 je`）→ 直接路径：
  namePool 空则回退 namePoolBytes（@0x1407e52cc `test rcx,rcx`），nodes 非空则
  nodeBytes 清零（@0x1407e530a `cmovne rdi/rcx/r8,0`）；
- `prefixBuckets==nil` → `pb := new(namePrefixBucketCounts)`（@0x1407e546d newobject）
  → `*pb = buildNamePrefixBucketCounts(cv.nodes, cv.namePool)`（@0x1407e5500 call），
  此路径 namePool 不回退（@0x1407e5649 直拷 `cv.namePool`）；
- `nameTrigram/nameBigram/pinyin/tombstones/deltaNodes/deltaByFRN/deltaReplaced`
  保持零值（duffzero 清零，@0x1407e5293）。

## G4 独立复核

- `backend/filesearch_index_windows.go`：新增 3 函数，字段布局与既有
  `nodeAt`/`decodeVolumeIndexNodesFromBytesWithNodeCount`/`addNamePrefixBucketCount`
  完全一致（IndexNode 24B 小端；namePrefixBucketCounts = firstByte[256]int32 +
  twoByte []int32；volumeIndexCheckpointView 字段偏移与 asm 0xe8/0xf0/0xf8/0x120 等对齐）。
  修正 `addNamePrefixBucketCount` 补 twoByte 空检查，与下游 `buildIndexSearchPlan`
  （`len(twoByte)==0x10000` 判 bigram 就绪）语义一致。
- `backend/filesearch_index_windows_test.go`：新增 6 个测试覆盖 round-trip、
  twoByte 空跳过、>10000 预分配、len%24!=0、ParentIdx 越界两类、namePool 越界、
  nil cv、prefixBuckets 复用/惰性构建、namePool 回退/不回退、nodeBytes 清零/保留。
- 全部 `[S]`，逻辑与 asm 逐地址对齐；`go build/vet/test` 全仓通过。

## 移交（本轮收尾）

3 新函数（[S]）+ 1 行为修正。FUNCS 3289/4754 = 69.19%，FAITHFUL 1599/4754 = 33.63%，
TRUE 3247/4754 = 68.30%。P=41 持平、UNMARKED=0 保持。
下一批：`loadVolumeIndexCheckpointMappedViewContext` 0x1407fb100（含 goroutine 闭包
func1/gowrap1 与 `loadVolumeIndexBucketContext` 0x1407f6f00）；`P=41 → [S]` 转换。
