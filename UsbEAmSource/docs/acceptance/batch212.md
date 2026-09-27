# 批次 212 — 候选节点名匹配（非拼音）+ 搜索计划类型修正

## 基线 / 收口

| 指标 | 基线（批次 211 收口） | 收口（批次 212） |
|---|---|---|
| FUNCS | 2787 | **2788** |
| S | 1253 | **1254** |
| S-inline | 36 | 36 |
| S-sig | 1385 | 1385 |
| P | 113 | 113 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2674 | **2675**（56.27%） |
| §10 差集 | 54 | 54 |

SHA256 `8EF85A98319BB00F7675897BA75DDBC2F7FC873CCED566FC361E1021B9A43D44`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批内容

`backend/filesearch_index_windows.go` 落地候选节点名匹配（非拼音）1 函数 [S]，并顺带修正
搜索计划链路的类型签名偏差（asm 实证发现）。

### 新建 [S]（体完整翻译）

- `volumeIndexReadView.matchSearchCandidateNode`（0x1407ecca0，1096B）：非拼音签名候选路径。
  结构同 `matchSearchCandidateNodeWithPinyin`（批次 211）：`matchNodeNameTerms` 算 matched →
  driver 未命中淘汰 → mask==matched 命中 → `^matched&others==0` 淘汰 → 沿 `node.ParentIdx`
  父链上溯、`matchHierarchyTermWithPinyin` 补齐 others 剩余层级术语、墓碑后代淘汰，末了
  `mask==matched`。区别：无 nodeIndex、无 b1/b2 拼音别名，name 直接传。

### 类型签名修正（asm 实证，行为不变）

- `nameFrequencyIndex` 改为 `namePrefixBucketCounts` 的类型别名。asm 0x140810e60 实证
  `buildIndexSearchPlan` 的 idx 实参是 `volumeIndexReadView.prefixBuckets`（*namePrefixBucketCounts），
  前 0x410 字节布局与 nameFrequencyIndex 完全一致。
- `estimateSearchTermCandidateCount` 字段访问改为 firstByte/twoByte（unigram/bigram/bigramReady
  为旧命名）；`len(idx.twoByte)==0x10000` 作 bigram 就绪标志。
- `matchNodeNameTerms` 签名从 7 参数（b,terms,caseSensitive,matched,driverIndex,driverBit,order）
  收敛为 4 参数（b,terms,caseSensitive,plan），matched/driver/driverBit/order 改从 plan 读取。
  asm 0x1407e3ac0 实证：栈参数整体是 nameSearchPlan（56B），并非拆开的四个标量。

## 关键知悉

- `matchNodeNameTerms` 调用约定：b/terms/caseSensitive 走寄存器（AX/BX/CX、DI/SI/R8、R9b），
  plan（56B）整体在栈；内部三段式（driver 快路径 / order 循环 / flags[0]&&!flags[1]&&flags[3]
  兜底循环）与拼音版 `matchNodeNameTermsWithPinyin` 的 order/兜底两段完全同构。
- `matchSearchCandidateNode` 参数铺排：node IndexNode 按值展开占 AX/BX/CX/DI/SI/R8（FRN 在 AX
  为 qword），name 占 R9/R10/R11，terms/plan/tombstoneLookup/caseSensitive 依次入栈——与拼音版
  （nodeIndex 占 AX）的差异根因是这里候选索引不作为参数传入。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1254 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`。
- **G3 行为**：新增 `TestMatchSearchCandidateNode` 黄金用例覆盖 driver 未命中、mask==matched、
  others 全匹配但 mask 未全匹配、沿父链补齐层级术语、父链墓碑淘汰五分支，均 PASS；签名修正后
  既有测试（estimateSearchTermCandidateCount / matchNodeNameTerms 相关）全量回归 PASS。
- **G4 review**：`filesearch_index_windows.go`（1 函数 + 3 处签名修正）+ 测试；vet/test/build 复验通过。

## 遗留（下一批）

- §10 差集 54 文件。`matchSearchCandidateNode` 已落地，其调用方闭包
  `searchPinyinContextWithTombstones.func1`（0x140811c60，84 行，汇编已抽取到
  `docs/goresym/pipeline/tmp/searchPinyinContextFunc1.asm.txt`）现依赖全部就绪，可作下一步；
  再下一步 `searchPinyinContextWithTombstones`（0x140810e60，98 行，700 行汇编已抽取）。
  其余候选：`pluginupdate.go`、`pluginwindow.go`、`twofactor_provision_misc.go` 各存根、
  `transport.go`、`TestReminderNotification`（0x1407a7320）。
