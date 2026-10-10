# 批次 386 · filesearch checkpoint 映射规划（三段 + 系统粒度）+2（FUNCS 3281）

## 目标

落地 2 个 `[S]` 函数（均新增），补齐 checkpoint 映射规划链：

1. `planVolumeIndexCheckpointMap` `[S]`（0x1407fcfe0，736B）— 三段（nodes/sorted/names）逐段页对齐规划。
2. `planVolumeIndexCheckpointMapWithSystemGranularity` `[S]`（0x1407fd2c0，576B）— GetSystemInfo 取分配粒度后转发。

## 基线 / 收口

| 指标 | 基线（batch 385 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3279 | 3281 |
| MARKED | 3279 | 3281 |
| S | 1549 | 1551 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1651 | 1651 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1586 | 1588 |
| USABLE | 1587 | 1589 |
| TRUE（S+S-inline+S-sig） | 3237 | 3239 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。
`go fmt ./backend` 干净。

## G2 契约

```
FUNCS=3281  MARKED=3281  UNMARKED=0  S=1551  S-inline=37  S-eq=1  S-sig=1651  P=41  FAITHFUL=1588  USABLE=1589  TRUE=3239
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1551 + 37 + 1 + 1651 + 41 = 3281 = FUNCS`。
FUNCS 3279→3281（+2）、S 1549→1551（+2）、TRUE 3237→3239（+2）、S-sig=1651/P=41/UNMARKED=0 保持。
`list_missing.js` 未落地数 348→346（−2），两函数退出缺失清单。

## G3 行为（asm 逐地址实证）

### 3.1 planVolumeIndexCheckpointMap [S 0x1407fcfe0]

签名 `(layout volumeIndexCheckpointLayout, granularity int64) ([3]volumeIndexMappedSectionPlan, error)`
（layout 8×int64 拆寄存器 rax..r10，granularity=r11，@0x1407fd032 保存 r11）。

- 逐段调用 planVolumeIndexMappedSection（@0x1407fd071/0b3/100）：
  - nodes：offset=NodeOffset、size=NodeBytes（@0x1407fd056/05e）；
  - sorted：offset=SortedOffset、size=SortedBytes（@0x1407fd093/09b）；
  - names：offset=NameOffset、size=NameBytes（@0x1407fd0dd/0e5）；
  - 三者 granularity=r11、limit=FileSize（@0x1407fd066/0a3/0ed rdi=[rsp+0xd0]）。
- 任一段 err（@0x1407fd076/0c0/105 test rsi）→ 返回 (零, err)（@0x1407fd1ff/1cb/133）。
- 成功组装三段 (MapOffset,MapBytes,SectionOffset,SectionBytes)（@0x1407fd13f..1b9 栈写
  [rsp+0x70..0xc8]，共 96B=3×32B），返回 (out, nil)（@0x1407fd1c1）。

### 3.2 planVolumeIndexCheckpointMapWithSystemGranularity [S 0x1407fd2c0]

签名 `(layout volumeIndexCheckpointLayout) ([3]volumeIndexMappedSectionPlan, error)`。

- newobject{volumeIndexSystemInfo}（@0x1407fd33d，类型 @rip+0x3f81bc）→ 保存 [rsp+0x110]。
- newobject{变参切片}（@0x1407fd359）→ 字段0=&info（@0x1407fd36d）。
- GetSystemInfo LazyProc.Call(1 arg=&info)（@0x1407fd385，全局 @0x141bc1930 → LazyProc
  @0x141bcfbc0 Name="GetSystemInfo"@0x140c4ded5,13B 实证）。
- info.AllocationGranularity(+0x28)<=0（@0x1407fd392/396 jg 失败）→ errVolumeIndexNodeBytesInvalid
  （@0x1407fd3ba，全局 @0x141bc3c50）。
- 否则转发 planVolumeIndexCheckpointMap(layout, granularity)（@0x1407fd411，r11=granularity），
  duffcopy 三段返回（@0x1407fd42b/44e/471）。

## G4 独立复核

- `backend/filesearch_index_windows.go`：新增两函数 + 全局 `volumeIndexGetSystemInfoProc`
  （kernel32.GetSystemInfo LazyProc）；复用已落地 `planVolumeIndexMappedSection`、
  `volumeIndexCheckpointLayout`（8 字段，types_filesearch.go line 252 已定义）、
  `volumeIndexMappedSectionPlan`（4 字段，line 272）、`volumeIndexSystemInfo`
  （AllocationGranularity@+0x28，line 316）、`errVolumeIndexNodeBytesInvalid`。
- `ExpectedBytes` 字段在规划阶段未被读取（asm @0x1407fd032 保存 r10 后无读取，实证）。
- 全部 `[S]` 标记，逻辑与 asm 逐地址对齐（含 struct 拆寄存器、LazyProc.Call 变参、idiv 页对齐）；
  `go build/vet/test` 全仓通过。

## 移交（本轮收尾）

2 函数落地（均 [S]）。FUNCS 3281/4754 = 69.02%，FAITHFUL 1588/4754 = 33.40%，
TRUE 3239/4754 = 68.13%。未落地文件差集仍 30、P=41 持平、UNMARKED=0 保持。
下一批：`buildVolumeIndexCheckpointLayout` 0x1407fc820（依赖 parseVolumeIndexCheckpointHeader
0x1407fc9a0，产 layout 供本批两函数）；`loadVolumeIndexCheckpointMappedViewContext` 0x1407fb100
（已见其调用链：OpenFile→Stat→ReadAtLeast 0x40→parseHeader→buildLayout→planMap→openMapped→Map×3）；
`parseVolumeNameTrigramHeader` 0x1407f9ba0 / `volumeNameTrigramHeaderMatchesIndex` 0x1407f9ca0；
`P=41 → [S]` 转换。
