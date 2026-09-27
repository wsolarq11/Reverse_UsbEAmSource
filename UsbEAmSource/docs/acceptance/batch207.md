# 批次 207 — filesearch 拼音模糊匹配域（Match/matchFileSearchPinyinTerm/matchNodeNameTermsWithPinyin + 结构修正）

## 基线 / 收口

| 指标 | 基线（批次 206 收口） | 收口（批次 207） |
|---|---|---|
| FUNCS | 2773 | **2776** |
| S | 1239 | **1242** |
| S-inline | 36 | 36 |
| S-sig | 1385 | 1385 |
| P | 113 | 113 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2660 | **2663**（56.02%） |
| §10 差集 | 54 | 54 |

SHA256 `ADBE298A702394145C08F4C3BF31B5DF2AA5196B291C0508D71E26D24DECCCD2`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B，`bash build.sh` 重建）。

## 本批内容

`backend/filesearch_pinyin_windows.go` 新建：3 函数 [S] 落地；`backend/filesearch_index_windows.go`
`nameSearchTerm` 结构修正。FUNCS +3 / S +3。

### 新建 [S]（体完整翻译）

- `fileSearchPinyinFuzzyMatcher.Match`（0x14080a7e0，160B）：`m==nil` 或 `m.contains==nil`
  → false；否则 `m.contains.Match(b)`（asm 实证转 regexp.Regexp.doExecute，返回结果非 nil 判真）。
- `matchFileSearchPinyinTerm`（0x14080fda0，160B）：
  `bytesMatchWildcardFold(b, term.patterns, false)` 命中 → true；否则 `term.matcher.Match(b)`。
- `matchNodeNameTermsWithPinyin`（0x14080fe40，1280B + func1 0x140810340，256B）：
  与 matchNodeNameTerms 同构三段式（快路径/order 顺序/兜底），匹配函数换成闭包
  `bytesMatchWildcardFold(b, patterns, caseSensitive)` → `matchFileSearchPinyinTerm(b1, term)`
  → `matchFileSearchPinyinTerm(b2, term)`。`plan nameSearchPlan` 按值传入。

### 结构修正（布局不变，语义补全）

`nameSearchTerm` 的 `flags [16]byte`（+0x28..0x38）拆为 `flags [8]byte`（+0x28..0x30）
+ `matcher *fileSearchPinyinFuzzyMatcher`（+0x30..0x38）。总大小仍 56B（0x38），
现有 flags[0..3] 访问（索引 <8）不受影响。

## 关键知悉

- `matchFileSearchPinyinTerm` 第二个调用 `Match` 的 receiver 取自 term +0x30（8 字节），
  `Match` 内部解引用 `[receiver]` 作 contains——证明 +0x30 是指向
  `fileSearchPinyinFuzzyMatcher` 的指针（contains 在结构 +0x00），而非 flags 后半。
- flags 仅前 4 字节（+0x28..0x2b）被 matchNodeNameTerms / matchNodeNameTermsWithPinyin
  访问；+0x2c..0x2f 保留，+0x30..0x37 为 matcher 指针。
- `matchNodeNameTermsWithPinyin` 签名经调用方 matchSearchCandidateNodeWithPinyin 反推确认：
  9 寄存器参数 = b/b1/b2 三个 []byte，栈参数 = terms（3 word）+ nameSearchPlan 按值（56B =
  7 word，mask/matched 打包在 +0x28 槽、driver@+0x38、driverBit@+0x40、order@+0x48/0x50/0x58）
  + caseSensitive（+0x60）。主函数读 matched@+0x2c（plan.matched）、driver@+0x38（plan.driver）、
  driverBit@+0x40（plan.driverBit）、order@+0x48/0x50（plan.order），与 nameSearchPlan 字段
  偏移逐一吻合。
- caseSensitive 只在闭包首段 bytesMatchWildcardFold 生效；matchFileSearchPinyinTerm 内部
  wildcard 固定 false（`xor r9d,r9d`）。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2776 / S=1242 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`。
- **G3 行为**：新增 `backend/filesearch_pinyin_windows_test.go`（黄金用例）验证三段式——
  order 顺序置位、兜底循环 flags 门控 + matcher 拼音命中置双位、快路径 driver 失败返回原值，
  均 PASS。
- **G4 review**：新建 `backend/filesearch_pinyin_windows.go`（3 函数）与对应测试；改
  `nameSearchTerm` 结构（`backend/filesearch_index_windows.go`）；vet/test/build 复验通过。

## 遗留（下一批）

- §10 差集 54 文件。filesearch 拼音模糊匹配域（Match/matchFileSearchPinyinTerm/
  matchNodeNameTermsWithPinyin）已闭环。候选：`classifyVolumePinyinMatch`（0x140810960，
  依赖 matchNodeNameTermsWithPinyin + matchSearchCandidateNodeWithPinyin）、`pluginupdate.go`、
  `pluginwindow.go`、`twofactor_provision_misc.go` 各存根、`transport.go`、
  `TestReminderNotification`（0x1407a7320）。
