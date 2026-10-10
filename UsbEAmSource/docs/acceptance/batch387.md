# 批次 387 · filesearch checkpoint 头解析 + 布局构建 +2（FUNCS 3282）

## 目标

落地 2 个 `[S]` 函数，打通 checkpoint 读链的头部解析与布局规划：

1. `parseVolumeIndexCheckpointHeader` `[S]`（0x1407fc9a0，608B）— 新增（64B 头解析 + 校验 + time.Unix）。
2. `buildVolumeIndexCheckpointLayout` `[S]`（0x1407fc820，384B）— `[S-sig]` 存根升级（偏移算术 + 溢出检查）。

## 基线 / 收口

| 指标 | 基线（batch 386 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3281 | 3282 |
| MARKED | 3281 | 3282 |
| S | 1551 | 1553 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1651 | 1650 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1588 | 1590 |
| USABLE | 1589 | 1591 |
| TRUE（S+S-inline+S-sig） | 3239 | 3240 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。
`go fmt ./backend` 干净。

## G2 契约

```
FUNCS=3282  MARKED=3282  UNMARKED=0  S=1553  S-inline=37  S-eq=1  S-sig=1650  P=41  FAITHFUL=1590  USABLE=1591  TRUE=3240
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1553 + 37 + 1 + 1650 + 41 = 3282 = FUNCS`。
FUNCS 3281→3282（+1，parse 新增）、S 1551→1553（+2，buildLayout 存根升级）、S-sig 1651→1650（−1，buildLayout）、
TRUE 3239→3240（+1）、UNMARKED=0/P=41 保持。`list_missing.js` 未落地数 346→345（−1，parse 退出缺失清单）。

## G3 行为（asm 逐地址实证）

### 3.1 parseVolumeIndexCheckpointHeader [S 0x1407fc9a0]

签名 `(b []byte) (volumeIndexCheckpointHeader, error)`（@0x1407fb2ec 调用：rax=ptr、rbx=0x40、rcx=0x40）。

二进制布局（64B，LE）：
| 偏移 | 字段 | 证据 |
|---|---|---|
| +0x0 | magic "UIDX"（0x58444955） | @0x1407fc9ce cmp dword |
| +0x4 | version 0x20003 | @0x1407fc9db cmp dword |
| +0x8 | entryCount uint32 | @0x1407fc9f6 mov r9d |
| +0xc | sortedCount uint32 | @0x1407fc9fa mov r10d |
| +0x10 | rootFRN uint64 | @0x1407fc9e8 mov rdx |
| +0x18 | journalID uint64 | @0x1407fca4f mov r12 |
| +0x20 | lastUSN int64 | @0x1407fca53 mov r13 |
| +0x28 | GeneratedAt.Unix() int64 | @0x1407fca65 add r15 |
| +0x30 | namePoolSize uint64 | @0x1407fca57 mov r11 |

校验顺序（任一失败 → errVolumeIndexNodeBytesInvalid @0x141bc3c50，逐路径 rip 验证）：
len<64（@0x1407fc9c8 jl）、magic≠"UIDX"（@0x1407fc9d4）、version≠0x20003（@0x1407fc9e2）、
rootFRN==0（@0x1407fc9ed test）、entryCount≠sortedCount（@0x1407fca00 cmp r10d,r9d）。
成功：generatedAt=time.Unix(sec,0)——wall=0、ext=sec+62135596800(=unixToInternal)、
loc=time.Local（@0x1407fca5b..a8d：r15=0xe7791f700=62135596800，r8=@rip+0x13c6250）。

### 3.2 buildVolumeIndexCheckpointLayout [S 0x1407fc820]

签名 `(fileSize int64, entryCount uint32, sortedCount uint32, namePoolSize uint64) (volumeIndexCheckpointLayout, error)`
（@0x1407fb3b8 调用：rax=fileSize（FileInfo.Size() 返回）、rbx=entryCount、rcx=sortedCount、rdi=namePoolSize）。

- namePoolSize>MaxInt64（@0x1407fc849 cmp rdi,0x7fffffffffffffff ja）→ 哨兵。
- nodeBytes=entryCount*24（@0x1407fc852 *3 → @0x1407fc858 shl 3）、sortedBytes=sortedCount*4（@0x1407fc85e lea rsi*4）。
- expected=namePoolSize+nodeBytes+sortedBytes+0x40（@0x1407fc862/866 lea r10）。
- expected<0x40（@0x1407fc872 jl）或 expected<nodeBytes+0x40（@0x1407fc880 jl）→ 哨兵
  （第三道 nameOffset<=expected @0x1407fc885 因 namePoolSize>=0 恒真，语义等价省略）。
- fileSize<expected（@0x1407fc8c0 cmp rax,r10 jge 失败）→ 哨兵。
- 成功布局（@0x1407fc8fa..924 栈写）：FileSize=fileSize、NodeOffset=0x40、
  NodeBytes=nodeBytes、SortedOffset=nodeBytes+0x40、SortedBytes=sortedBytes、
  NameOffset=nodeBytes+sortedBytes+0x40、NameBytes=namePoolSize、ExpectedBytes=expected。

## G4 独立复核

- `backend/filesearch_index_windows.go`：新增 `parseVolumeIndexCheckpointHeader` + 升级
  `buildVolumeIndexCheckpointLayout`；新增 import `math`；新增常量
  `volumeIndexCheckpointHeaderMagic="UIDX"`/`volumeIndexCheckpointHeaderSize=64`。
- 复用已落地 `volumeIndexCheckpointHeader`（types_filesearch.go:616，字段名/类型逐一对应）、
  `volumeIndexCheckpointLayout`（:252）、`IndexNode`（types_misc.go:28，24B 实证）、
  `volumeIndexFormatVersion=0x20003`、`errVolumeIndexNodeBytesInvalid`。
- `buildVolumeIndexCheckpointLayout` 返回类型由 interface{} 修正为具体签名
  （调用者 loadVolumeIndexCheckpointMappedViewContext 待落地，无现有调用者，安全改签名）。
- 全部 `[S]` 标记，逻辑与 asm 逐地址对齐（含 LE 二进制布局、time.Unix 内部 ext 编码、
  溢出检查、偏移算术）；`go build/vet/test` 全仓通过。

## 移交（本轮收尾）

2 函数落地（均 [S]）。FUNCS 3282/4754 = 69.04%，FAITHFUL 1590/4754 = 33.45%，
TRUE 3240/4754 = 68.15%。未落地文件差集仍 30、P=41 持平、UNMARKED=0 保持。
下一批：`loadVolumeIndexCheckpointMappedViewContext` 0x1407fb100（已见全链：OpenFile→Stat→
ReadAtLeast 0x40→parseHeader（本批）→buildLayout（本批）→planMapWithSystemGranularity（b386）→
openMapped/Map（b385）→validateVolumeIndexNodeBytes）；`validateVolumeIndexNodeBytes` /
`parseVolumeIndexCheckpointNodes`（节点流解析）；`parseVolumeNameTrigramHeader` 0x1407f9ba0 /
`volumeNameTrigramHeaderMatchesIndex` 0x1407f9ca0（trigram 读链）；`P=41 → [S]` 转换。
