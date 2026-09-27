# 批次 209 — 拼音别名查询链（aliasesForNode/pinyinAliasesForNode/matchHierarchyTermWithPinyin）

## 基线 / 收口

| 指标 | 基线（批次 208 收口） | 收口（批次 209） |
|---|---|---|
| FUNCS | 2779 | **2782** |
| S | 1245 | **1248** |
| S-inline | 36 | 36 |
| S-sig | 1385 | 1385 |
| P | 113 | 113 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2666 | **2669**（56.14%） |
| §10 差集 | 54 | 54 |

SHA256 `7C8EAFC81B3FC2F0713F9F8B63FAC53CBB292FFA01492DE8FBDBB3A538FF2AF0`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B，`bash build.sh` 重建）。

## 本批内容

`backend/filesearch_pinyin_windows.go` 扩展：3 函数 [S] 落地。FUNCS +3 / S +3。

### 新建 [S]（体完整翻译）

- `volumePinyinIndex.aliasesForNode`（0x14080df20，800B）：
  在 `m.records`（每记录 32B）内对 `nodeIndex` 做无符号 lower_bound 二分，命中后按记录头
  12B 切出 `pinyinFull` / `pinyinInitials`。入口护栏 `m==nil || nodeIndex<0 ||
  nodeIndex>=0x40000000 → (nil,nil,false)`；`len(m.aliases) < fullOffset+fullLen+initLen →
  (nil,nil,false)`。
- `volumeIndexReadView.pinyinAliasesForNode`（0x14080f900，608B）：
  `nodeIndex>=0x40000000` 走 deltaNodes 覆盖（取 `delta.pinyinFull/pinyinInitials`，
  `len(deltaNodes)<=deltaIdx` 或两别名皆空 → false）；否则转 `v.pinyin.aliasesForNode`，
  `v.pinyin==nil → false`。
- `volumeIndexReadView.matchHierarchyTermWithPinyin`（0x14080fb60，576B）：
  先 `bytesMatchWildcardFold(b, term.patterns, caseSensitive)` 命中即 true；否则取拼音别名，
  `!ok → false`；依次 `matchFileSearchPinyinTerm(full, term)`、`matchFileSearchPinyinTerm(initials, term)`。

## 关键知悉

- `volumePinyinIndex` 字段偏移经 asm 复验：`records` @ +0x28（ptr）/ +0x30（len）/ +0x38（cap），
  `aliases` @ +0x40（ptr）/ +0x48（len）/ +0x50（cap）——与 `identity`（40B=0x28）后紧跟两
  切片的布局一致。记录头小端 12B：+0x00 nodeIndex u32、+0x04 fullOffset u32、+0x08 fullLen u16、
  +0x0a initLen u16；别名区以 `fullOffset` 为起点，先 full（fullLen）后 initials（initLen）连续存放。
- 二分用**无符号**比较（`jae`）：`recNodeIndex < uint32(nodeIndex)` 时 lo=mid+1，否则 hi=mid；
  循环结束 lo 即首个 `>= nodeIndex` 的记录下标，再核对 `records[lo].nodeIndex == nodeIndex`。
- `volumeIndexDeltaNode` 字段偏移经 asm 复验：`pinyinFull` @ +0x40（3 word）、`pinyinInitials`
  @ +0x58（3 word）——与结构里 node(0x18)+name(0x18)+bigram(0x8)+trigram(0x8) 后的位置一致。
- `matchHierarchyTermWithPinyin` 里 `test r9b` 检查的是 `pinyinAliasesForNode` 返回的 **ok 标志**
  （call 覆盖了 r9 寄存器），并非 `caseSensitive`（caseSensitive 已在首次 bytesMatchWildcardFold
  调用时被 r9d 复用）。故"拼音兜底"仅在别名存在时进行，与 caseSensitive 无关。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2782 / S=1248 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`。
- **G3 行为**：新增 `TestVolumePinyinAliasesForNode` 黄金用例覆盖二分命中、首/末记录、缺配、
  越界、nil 护栏与 delta 覆盖，均 PASS。
- **G4 review**：扩展 `backend/filesearch_pinyin_windows.go`（3 函数）+ 测试；vet/test/build
  复验通过。

## 遗留（下一批）

- §10 差集 54 文件。拼音别名查询链已闭环（aliasesForNode/pinyinAliasesForNode/
  matchHierarchyTermWithPinyin），为 `matchSearchCandidateNodeWithPinyin`（0x140810440）清除了
  拼音侧依赖。剩余依赖：`IsTombstonedDescendant`（0x1407f1a60，49 行）、`nodeAtIndex`
  （0x1407f2240，24 行）、`nodeName`（0x1407f1e40，29 行），均属 filesearch_index_windows.go
  蓝图未落地段。候选：`searchPinyinContextWithTombstones`（0x140810e60）、`pluginupdate.go`、
  `pluginwindow.go`、`twofactor_provision_misc.go` 各存根、`transport.go`、
  `TestReminderNotification`（0x1407a7320）。
