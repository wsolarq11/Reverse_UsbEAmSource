# 批次 202 — filesearch 名称搜索计划域完整还原（6 函数 + 3 结构）

## 基线 / 收口

| 指标 | 基线（批次 201 收口） | 收口（批次 202） |
|---|---|---|
| FUNCS | 2763 | **2769** |
| S | 1224 | **1230** |
| S-inline | 36 | 36 |
| S-sig | 1389 | 1389 |
| P | 114 | 114 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2649 | 2655（55.8%） |
| §10 差集 | 54 | 54 |

SHA256 `27D52FEAFB342F72C01E180EA53D49D29E1F7BFC587B224CAAD3EBABEFD587AF`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,322,240 B，`bash build.sh` 重建）。

## 本批内容

`backend/filesearch_index_windows.go` 名称搜索计划域闭环（新落地 6 个 [S] 函数 + 3 个结构，
`matchNodeNameTerms` 由"当前目标"转正为真实 [S] 体）。全为寄存器级 asm 实证，无推断签名。

### 数据结构（3 个）

- `nameSearchTerm`（56B=0x38）：`_ [16]byte`（+0x00，上游构建本域不访问）+
  `patterns [][]byte`（+0x10，wildcard 片段已小写）+ `flags [16]byte`（+0x28）。
- `nameFrequencyIndex`（0x410）：`unigram [256]int32`（+0x00）+ `bigram *[65536]int32`
  （+0x400）+ `bigramReady uint64`（+0x408，==0x10000 表示 bigram 就绪）。
- `nameSearchPlan`（56B）：`mask/matched/others uint32` + `driver int` + `driverBit uint32` +
  `order []int`。

flags 语义经 asm 实证（非臆造）：`flags[0]||flags[1]` 判"模式术语"（走
`bytesMatchWildcardFold`），`flags[0]==0&&flags[1]==0` 判"字面量候选"，
`flags[2]` 置 others 掩码，`flags[3]` 为兜底循环的必需位。

### 函数（6 个，[S] 落地）

| 函数 | VA / 大小 | 关键逻辑 |
|---|---|---|
| `shouldPrioritizeSearchTerm` | 0x1407e19c0 / 704B | 五级比较：计数小优先 → 单片段优先 → 总长短优先 → 片段少优先 → 索引大优先 |
| `estimateSearchTermCandidateCount` | 0x1407e3100 / 320B | bigram（前 2 字符）/ unigram（首字符）查频表，命中 >0 返回，否则 fallback |
| `selectDriverTermIndex` | 0x1407e3240 / 352B | 跳过模式术语，取候选计数最小的字面量术语，无候选返 -1 |
| `buildOrderedNameTermIndices` | 0x1407e33a0 / 544B | 收集除 driver 外的字面量索引，`sort.SliceStable` 排序（闭包即 func1） |
| `buildIndexSearchPlan` | 0x1407e36e0 / 992B | mask=(1<<n)-1；模式术语命中置 matched，flags[2] 置 others；选 driver + order |
| `matchNodeNameTerms` | 0x1407e3ac0 / 928B | 三段：driver 快路径 → order 顺序 → 全部术语兜底（flags[0]&&!flags[1]&&flags[3]） |

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`（0.942s）；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2769 / S=1230 / S-inline=36 / S-sig=1389 / P=114 / UNMARKED=0`。
- **G3 行为**：六函数控制流逐寄存器追踪（`buildIndexSearchPlan.asm` 186 行 +
  `selectDriverTermIndex.asm` + `buildOrderedNameTermIndices.asm` + `estimateSearchTermCandidateCount.asm` +
  `shouldPrioritizeSearchTerm.asm` + `matchNodeNameTerms.asm`），bit 掩码计算
  `uint32(1)<<uint(i)`（idx≥32 自动归零）、mask 下溢 `(1<<n)-1`（n≥32→0xffffffff）、
  bigram 键 `c0<<8|c1`（16 位）均与 asm 逐条对齐。
- **G4 review**：仅改 `backend/filesearch_index_windows.go`（新增 import "sort" + 3 结构 + 6 函数）；
  无跨文件写重叠；未动既有 3 函数；vet/test/build 复验通过。

## 遗留（下一批）

- §10 差集 54 文件。desktopwidgets（calendar/clock）、filesearch 索引计划域均已闭环。
  候选：qrcode.go 2 个 [S-sig] 剪贴板存根、`matchNodeNameTermsWithPinyin`（0x14080fe40，
  filesearch_pinyin_windows.go 域）、`pluginupdate.go`、`pluginwindow.go`、`twofactor.go`、
  `transport.go`、`TestReminderNotification`（0x1407a7320，依赖 launcherWidgetStore.Read +
  normalize/validateDesktopReminderAudio 链）。
