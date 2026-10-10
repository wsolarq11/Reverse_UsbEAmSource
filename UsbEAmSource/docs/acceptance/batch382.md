# 批次 382 · filesearch 搜索键收集 + 持久化路径 + 节点校验 + 三元组头 +7（FUNCS 3271）

## 目标

filesearch 域一次性落地 7 个 `[S]` 函数（其中 2 个由 `[S-sig]` 存根升级为真函数）：

1. `collectSearchTermBigramKeys` `[S]`（0x1407e27c0，576B）。
2. `selectSearchBigramKeys` `[S]`（0x1407e2a00，384B，原 `[S-sig]` 存根升级）。
3. `collectSearchTermTrigramKeys` `[S]`（0x1407e2ca0，736B）。
4. `selectSearchTrigramKeys` `[S]`（0x1407e2f80，384B，原 `[S-sig]` 存根升级）。
5. `newVolumeIndexPersistencePaths` `[S]`（0x1407e1020，576B）。
6. `validateVolumeIndexNodeBytes` `[S]`（0x1407e2420，512B）。
7. `buildVolumeNameTrigramHeaderLocked` `[S]`（0x1407f66e0，512B）。

## 基线 / 收口

| 指标 | 基线（batch 381 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3266 | 3271 |
| MARKED | 3266 | 3271 |
| S | 1530 | 1537 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1657 | 1655 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1567 | 1574 |
| USABLE | 1568 | 1575 |
| TRUE（S+S-inline+S-sig） | 3224 | 3229 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

```
FUNCS=3271  MARKED=3271  UNMARKED=0  S=1537  S-inline=37  S-eq=1  S-sig=1655  P=41  FAITHFUL=1574  USABLE=1575  TRUE=3229
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1537 + 37 + 1 + 1655 + 41 = 3271 = FUNCS`。
FUNCS 3266→3271（+5）、S 1530→1537（+7）、S-sig 1657→1655（−2，两存根升级为真）、
TRUE 3224→3229（+5）、UNMARKED=0/P=41 保持。`list_missing.js` 未落地数 361→356（−5），七函数退出缺失清单。

## G3 行为（asm 逐地址实证）

### 3.1 collectSearchTermBigramKeys [S 0x1407e27c0]

签名 `(term nameSearchTerm) []uint16`（rax=term 首字，@0x1407e27c0 起）。

- `flags[0]||flags[1]` 非零（模式术语，@0x1407e27da/7e2 检查 [rsp+0x108]/[rsp+0x109]）→ 返回 nil（@0x1407e2895）。
- 遍历 `patterns`（[]byte 步长 0x18，@0x1407e28b5 判 `[rdx+8]==2` 即 len==2）；
  两字节 ASCII 大写→小写（@0x1407e28c3/28f5 `[c-0x41]<=0x19 → or 0x20`）；
  组合 `uint16(c0)<<8|c1`（@0x1407e2907/290e）。
- `mapaccess2`（@0x1407e292e）去重；`mapassign`（@0x1407e2983）后 append（初始 cap=2，@0x1407e2804 makeslice ecx=2）。

### 3.2 selectSearchBigramKeys [S 0x1407e2a00]

签名 `(terms []nameSearchTerm) []uint16`（rax/rcx=ptr/len，@0x1407e2a1a）。

- 遍历 terms（步长 0x38，@0x1407e2a3c `add r8,0x38`）→ `collectSearchTermBigramKeys`（@0x1407e2ad2）。
- keys 空（`test rbx,rbx` @0x1407e2ad7）→ 跳过；result 为 nil 或 len(keys)>len(result)
  （@0x1407e2b0e `cmp rbx,rsi; jg`）→ result=keys。返回最长键组。

### 3.3 collectSearchTermTrigramKeys [S 0x1407e2ca0]

签名 `(term nameSearchTerm) []uint32`。

- `flags[0]||flags[1]` 非零 → nil（@0x1407e2d8a）。
- 遍历 patterns，`len(p)>=3`（@0x1407e2daa `cmp r8,3; jl`）；对每个 3 字节滑动窗口
  （@0x1407e2e17/7e20 边界 `r11=r10+2`）三字节大写→小写（@0x1407e2e58/6c/91）；
  组合 `uint32(c0)<<16|c1<<8|c2`（@0x1407e2ea3/2eaa/2eaf）。
- `mapaccess2_fast32`/`mapassign_fast32`（@0x1407e2ec8/2f00）去重；append（cap=8，@0x1407e2ce2 makeslice ecx=8）。

### 3.4 selectSearchTrigramKeys [S 0x1407e2f80]

签名 `(terms []nameSearchTerm) []uint32`。结构与 3.2 同构（@0x1407e3052 调 collectSearchTermTrigramKeys），返回最长键组。

### 3.5 newVolumeIndexPersistencePaths [S 0x1407e1020]

签名 `(root string) volumeIndexPersistencePaths`（rax/rcx=ptr/len）。

- 从末尾回扫最后一个 `\`/`/`/`.`（@0x1407e1065/1070/1076 逐字节比较）；命中 `.` 且其索引
  大于最后一个 `\`/`/`（@0x1407e107c 起比较）→ `base=root[:dot]`，否则 `base=root`。
- `CheckpointPath = root`（@0x1407e1133 起）；其余五字段 = base + 后缀：
  `.meta`（5B @0x1407e114c）、`.wal`（4B）、`.tri`（4B）、`.bkt`（4B）、`.pyn`（4B）
  （`.rdata` 常量 @0x140c35c68/0x140c3480a/0x140c3480e/0x140c34812/0x140c34816 实证）。

### 3.6 validateVolumeIndexNodeBytes [S 0x1407e2420]

签名 `(nodeBytes []byte, namePoolSize int) error`（rax/rbx/rcx=ptr/len/cap、rdi=namePoolSize）。

- `len%24!=0 || namePoolSize<0`（@0x1407e2450 magic 除 24 判余、@0x1407e2455 test rdi）→
  全局错误（@0x1407e245e）。
- 逐节点（24B）解析（字段偏移 @0x1407e24bd/4cb/4da/4ea/4f9/50a = FRN/NameOffset/ParentIdx/ModTime/NameLen/Flags）：
  - `ParentIdx < -1`（@0x1407e2548）→ 错误；`ParentIdx >= nodeCount`（@0x1407e2565）→ 错误；
  - `NameOffset+NameLen > namePoolSize`（@0x1407e2590 add/2593 cmp rdi）→ 错误。
- 四类失败均返回同一全局错误（四路 rip 引用 @0x1407e245e/254e/256a/259c 全指向
  0x141bc3c50 的 `errors.New`），错误串经 .rdata 实证为 `索引文件损坏或版本不兼容`（36B）。
- 全通过 → nil。

### 3.7 buildVolumeNameTrigramHeaderLocked [S 0x1407f66e0]

签名 `(idx *VolumeIndex, entryCount int) [64]byte`（rax=idx、rbx=entryCount）。

- `entryCount<0 || entryCount>0xffffffff`（@0x1407f6728 test rbx/jl、@0x1407f6732 cmp rbx,0xffffffff）→
  返回 64B 零头（@0x1407f6737）。
- 否则写头（64B，字段布局 @0x1407f6799 起）：
  magic `"UITG"`（0x47544955 @0x1407f6799）、version 1（@0x1407f67a2）、
  indexVersion 0x20003（@0x1407f67ab）、entryCount（@0x1407f67b4）、
  rootFRN（idx+0x30 @0x1407f67b8）、journalID（idx+0x80 @0x1407f67c1）、
  lastUSN（idx+0x88 @0x1407f67cd）、GeneratedAt.Unix()（idx+0x90/+0x98 内联 wall/ext
  转 unixSec @0x1407f67d9-6812）、LastMutationAt.UnixNano()（idx+0xa8 经 timeToUnixNano
  @0x1407f6817-6831）、保留零（@0x1407f6837）。

## G4 独立复核

- `backend/filesearch_index_windows.go`：+7 函数（全 [S]）；`backend/filesearch_windows.go`：−2
  `[S-sig]` 存根（selectSearchBigramKeys/selectSearchTrigramKeys，体迁至 index 文件真实现）。
- 新增 `errVolumeIndexNodeBytesInvalid`（消息串与 .rdata 实证一致）；新增头部格式常量
  `volumeNameTrigramHeaderMagic/volumeNameTrigramHeaderVersion/volumeIndexFormatVersion/
  maxVolumeIndexEntryCount`。
- 复用已落地 `timeToUnixNano`/`nameSearchTerm`/`volumeIndexPersistencePaths`/`IndexNode`，
  无新依赖；`go build/vet/test` 全仓通过；`gofmt -l ./backend` 无差异。

## 移交（本轮收尾）

7 函数落地（均 [S]，2 存根升级）。FUNCS 3271/4754 = 68.80%，FAITHFUL 1574/4754 = 33.11%，
TRUE 3229/4754 = 67.92%。未落地文件差集仍 30、P=41 持平、UNMARKED=0 保持。
下一批：remoteicons 域剩余（`validateRemoteIconCachePath` 0x1409643a0 / `validateRemoteIconTemporaryFile`
0x1409652a0 / `writeRemoteIconCache` 0x140963b80）；plugin update catalog 域
（`mergeLocalAndRemotePluginManifest`/`remoteCatalogEntryToManifest`/
`validateAndNormalizeRemoteCatalogEntries`）；`directLauncherNetworkAccess.newHTTPClient`/
`configBackedLauncherNetworkAccess.newHTTPClient` 体（需先定 client +0x28 字段与 Transport.Clone 装配）；
`P=41 → [S]` 转换。
