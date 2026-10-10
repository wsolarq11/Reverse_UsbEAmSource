# 批次 383 · filesearch 排序索引定位 + 映射读视图租约 +2（FUNCS 3273）

## 目标

filesearch 域落地 2 个 `[S]` 函数（均从缺失清单直接还原，非存根升级）：

1. `volumeIndexReadView.sortedIndexAt` `[S]`（0x1407f2460，544B）。
2. `volumeIndexMappedReadProvider.Acquire` `[S]`（0x1407e4100，512B）。

## 基线 / 收口

| 指标 | 基线（batch 382 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3271 | 3273 |
| MARKED | 3271 | 3273 |
| S | 1537 | 1539 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1655 | 1655 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1574 | 1576 |
| USABLE | 1575 | 1577 |
| TRUE（S+S-inline+S-sig） | 3229 | 3231 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

```
FUNCS=3273  MARKED=3273  UNMARKED=0  S=1539  S-inline=37  S-eq=1  S-sig=1655  P=41  FAITHFUL=1576  USABLE=1577  TRUE=3231
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1539 + 37 + 1 + 1655 + 41 = 3273 = FUNCS`。
FUNCS 3271→3273（+2）、S 1537→1539（+2）、TRUE 3229→3231（+2）、UNMARKED=0/P=41 保持。
`list_missing.js` 未落地数 356→354（−2），两函数退出缺失清单。

## G3 行为（asm 逐地址实证）

### 3.1 volumeIndexReadView.sortedIndexAt [S 0x1407f2460]

签名 `(v volumeIndexReadView, index int) (int32, bool)`（rax=index，接收者经栈传递；@0x1407f2480
`test rax; jl` 判 index<0）。

- index<0（@0x1407f2483）→ 返回 `(-1,false)`（@0x1407f263a eax=0xffffffff、ebx=0）。
- sortedIndices 非空（[rsp+0x148]≠0，@0x1407f2494）：
  - index≥len（@0x1407f24a3 `jge`）→ `(-1,false)`（@0x1407f2522）；
  - 否则读 `sortedIndices[index]`（@0x1407f24b1 `mov eax,[rcx+rax*4]`）→ nodeIndex；
  - nodeIndex<0（@0x1407f24b6 `jl`）→ `(nodeIndex,false)`。
- sortedIndices 空 → sortedBytes 分支（@0x1407f2532）：
  - len==0 或 len%4≠0（@0x1407f2540/2549）→ `(-1,false)`；
  - index≥len/4（@0x1407f2563）→ `(-1,false)`；
  - 否则小端读 `sortedBytes[index*4:]` int32（@0x1407f25a0）→ nodeIndex；
  - nodeIndex<0（@0x1407f25a5）→ `(nodeIndex,false)`。
- nodeIndex≥0：count = nodeBytes 非空 ? len(nodeBytes)/24（magic 0xaaaaaaaaaaaaaaab，@0x1407f24e7）:
  len(nodes)（@0x1407f24ff）；返回 `(nodeIndex, count>nodeIndex)`（@0x1407f2509 `setg`）。

### 3.2 volumeIndexMappedReadProvider.Acquire [S 0x1407e4100]

签名 `(p *volumeIndexMappedReadProvider) volumeIndexReadLease`（rax=接收者）。

- nil（@0x1407e4153 `test rax; je`）→ 零 lease（@0x1407e42c4）。
- Mutex lock（@0x1407e4170 cmpxchg → @0x1407e4180 lockSlow）。
- closed（+0x111，@0x1407e418d `cmp byte[rcx+0x111],0`）==0 即已关闭：
  - 锁内 duffcopy 复制 view（+8，@0x1407e419a/41ad）→ unlock（@0x1407e41b8）→
    返回 lease{view}（release 保持 nil，@0x1407e41cd-420c 复制）。
- 否则（@0x1407e421e）active(+0x108)++ → 取 release(+0x100，@0x1407e4225) → unlock →
  复制 view → 返回 lease{view,release}（@0x1407e425b-42aa）。

## G4 独立复核

- `backend/filesearch_index_windows.go`：+`sortedIndexAt`（复用已落地 `nodeBytes`/`nodes`/
  `sortedIndices`/`sortedBytes` 字段，无新依赖）；+`Acquire`（满足 `volumeIndexReadProvider`
  接口，复用 `mu`/`view`/`release`/`active`/`closed` 字段）。
- 两函数均用 `[S]` 标记，逻辑与 asm 逐地址对齐（含 magic 除 24、cmpxchg/xadd 锁快慢路径、
  duffcopy 视图复制）；`go build/vet/test` 全仓通过；`gofmt -l ./backend` 无差异。

## 移交（本轮收尾）

2 函数落地（均 [S]）。FUNCS 3273/4754 = 68.84%，FAITHFUL 1576/4754 = 33.15%，
TRUE 3231/4754 = 67.96%。未落地文件差集仍 30、P=41 持平、UNMARKED=0 保持。
下一批：filesearch 写链上下文（`writeVolumeIndexSortedIndicesContext` 0x1407f61e0 /
`writeVolumeNameTrigramSignaturesContext` 0x1407f7cc0 / `saveVolumeIndexMetaContext`
0x1407f7f00 / `writeAllContext` 0x1407f6420，四者互依，需先定 `writeAllContext` 签名与
writer 类型）；`volumeIndexReadView.resolvePath` 0x1407f11e0（依赖 resolvePathWithTombstones）；
remoteicons 域（`validateRemoteIconCachePath`/`validateRemoteIconTemporaryFile`/
`writeRemoteIconCache`）；`P=41 → [S]` 转换。
