# 批次 211 — 候选节点拼音匹配主函数（matchSearchCandidateNodeWithPinyin）

## 基线 / 收口

| 指标 | 基线（批次 210 收口） | 收口（批次 211） |
|---|---|---|
| FUNCS | 2786 | **2787** |
| S | 1252 | **1253** |
| S-inline | 36 | 36 |
| S-sig | 1385 | 1385 |
| P | 113 | 113 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2673 | **2674**（56.25%） |
| §10 差集 | 54 | 54 |

SHA256 `482D45B50EE1B8E34417796FE9184522DB1C20C036671A72C6EF8EB46039F414`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批内容

`backend/filesearch_pinyin_windows.go` 落地候选节点拼音匹配主函数 1 个 [S]。FUNCS +1 / S +1。
该函数是 filesearch 拼音搜索链路的主入口，依赖全链已在批次 209/210 就绪。

### 新建 [S]（体完整翻译）

- `volumeIndexReadView.matchSearchCandidateNodeWithPinyin`（0x140810440，1312B）：四段式——
  1. `matchNodeNameTermsWithPinyin` 计算普通术语 matched；
  2. driver 术语未命中 → false；
  3. `mask==matched` → true，`^matched&others==0` → false；
  4. 从 `node.ParentIdx` 沿父链上溯，逐个节点用 `matchHierarchyTermWithPinyin` 补齐
     others 剩余层级术语，墓碑后代（`IsTombstonedDescendant`）直接淘汰，末了返回
     `mask==matched`。

## 关键知悉

- 签名 14 参数（1 值接收者 + 9 显式 + 5 栈）：`nodeIndex int32`（AX）、`node IndexNode`
  （BX/CX/DI/SI/R8/R9 按值展开）、`tombstoneLookup *volumeIndexTombstoneLookup`（R10）、
  `caseSensitive bool`（R11b）、`b/b1/b2 []byte`、`terms []nameSearchTerm`、
  `plan nameSearchPlan`。asm 实证：`nodeIndex`（AX）在实现中未被使用，起始父索引取自
  `node.ParentIdx`（DI）；IndexNode 其余字段（FRN/NameOffset/ModTime/NameLen/Flags）亦未使用。
- `nameSearchPlan` 字段口径复验：mask uint32@+0x00、others uint32@+0x08、driver int@+0x10、
  driverBit uint32@+0x18；`^matched&others` 为 uint32 位运算。
- 内层遍历 `for idx := range terms`（非 plan.order），bit 用 `uint32(1)<<idx`（idx>=32 时
  x86 以 `cmp/sbb/and` 归零，与 Go 移位溢出语义一致）；命中则 `matched|=bit`、
  `required&=^bit`，未命中恢复旧掩码。
- 语义闭环：`buildIndexSearchPlan` 中 `flags[2]!=0` 的术语进入 `others`（层级术语，需沿父链），
  本函数的第 4 段正是补全这些位；`matchNodeNameTermsWithPinyin` 只匹配普通术语（driver/order/兜底）。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2787 / S=1253 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`。
- **G3 行为**：新增 `TestMatchSearchCandidateNodeWithPinyin` 黄金用例覆盖 driver 未命中、
  mask==matched、others 全匹配但 mask 未全匹配、沿父链补齐层级术语、父链墓碑淘汰五分支，均 PASS。
- **G4 review**：仅 `backend/filesearch_pinyin_windows.go`（1 函数）+ 测试；vet/test/build 复验通过。

## 遗留（下一批）

- §10 差集 54 文件。`matchSearchCandidateNodeWithPinyin` 的调用方
  `searchPinyinContextWithTombstones`（0x140810e60，98 行）及其闭包
  `.func1`（0x140811c60，84 行）是下一步候选，且已抽取汇编到
  `docs/goresym/pipeline/tmp/searchPinyinContext.asm.txt`（700 行）。候选：
  `classifyVolumePinyinMatch`（0x140810960，已落地）、`pluginupdate.go`、
  `pluginwindow.go`、`twofactor_provision_misc.go` 各存根、`transport.go`、
  `TestReminderNotification`（0x1407a7320）。
