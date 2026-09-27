# 批次 208 — filesearch 拼音分类域（MatchExact/MatchPrefix/classifyVolumePinyinMatch）

## 基线 / 收口

| 指标 | 基线（批次 207 收口） | 收口（批次 208） |
|---|---|---|
| FUNCS | 2776 | **2779** |
| S | 1242 | **1245** |
| S-inline | 36 | 36 |
| S-sig | 1385 | 1385 |
| P | 113 | 113 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2663 | **2666**（56.08%） |
| §10 差集 | 54 | 54 |

SHA256 `09071D1036AC2D99EC7237D7ACB5893DCFB2629FDBC48ADE98C62A5F698181BF`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B，`bash build.sh` 重建）。

## 本批内容

`backend/filesearch_pinyin_windows.go` 扩展：3 函数 [S] 落地。FUNCS +3 / S +3。

### 新建 [S]（体完整翻译）

- `fileSearchPinyinFuzzyMatcher.MatchExact`（0x14080a880，416B）：
  `m==nil || m.exact==nil → false`；`m.exact.Match(b)` 命中 → true；`allowDot`（byte 参数非 0）
  → false；否则从 `len(b)-1` 往前找 `.`，`dot<=0 → false`，否则 `m.exact.Match(b[:dot])`
  （去掉最后一个 `.` 及之后的扩展名重试）。
- `fileSearchPinyinFuzzyMatcher.MatchPrefix`（0x14080aa20，160B）：
  `m==nil || m.prefix==nil → false`；否则转 `regexp.Regexp.doExecute`（前缀正则匹配）。
- `classifyVolumePinyinMatch`（0x140810960，480B + func1 0x140810b40，800B）：
  `term==nil || len(term.patterns)==0 → 5`；取 `patterns[0]` 作 aliasB，唯一 pattern 时同时作
  aliasA；闭包 classify 按优先级判定：matchFileSearchPinyinTerm → 精确（MatchExact / 与 aliasA
  忽略大小写全等 / `!allowDot` 时 bytesStemEqualFold）→ 前缀（MatchPrefix / 与 aliasB 前缀忽略
  大小写全等）→ 其他；先试 b1（3/4/5）再试 b2（6/7/8），均不中返回 5。

## 关键知悉

- `classifyVolumePinyinMatch` 的 r9 参数经 asm 反推确认为 **`*nameSearchTerm`**：`[r9+0x10]` =
  patterns.ptr（指向 `[][]byte` 底层数组，取 `[r9+0x10]+0/8/0x10` 得 patterns[0] 的
  ptr/len/cap）、`[r9+0x18]` = patterns.len、`[r9+0x30]` = matcher，与 nameSearchTerm 字段偏移
  逐一吻合。此前 8 字节 gap 疑点即 patterns[0] 的 slice 头三重解引用，非独立结构。
- 组 A（唯一别名）= `patterns[0]` 仅当 `len(patterns)==1`；组 B = `patterns[0]` 恒成立。
  主函数 `cmp [r9+0x18], 1` 后 jne 分支将组 A 三词清零，组 B 三词保留——对应 aliasA 置空、
  aliasB 恒取首 pattern。
- 分类语义：闭包三参数 `exact/prefix/other` 分别映射 3/4/5（b1 路）与 6/7/8（b2 路）；
  精确级含三段兜底（正则精确 / 与唯一 pattern 忽略大小写全等 / `!allowDot` 时去扩展名词干
  相等），前缀级含 MatchPrefix 与"b 前缀 == aliasB"两路，其余归 other。
- 两处"忽略大小写全等"均只对 b 做 ASCII `'A'..'Z'` 折叠（aliasA/aliasB 由调用方保证已小写），
  与 bytesStemEqualFold 的折叠一致；与标准库 bytes.EqualFold（双端折叠）不同，故内联实现。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2779 / S=1245 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`。
- **G3 行为**：新增 `TestClassifyVolumePinyinMatch` 黄金用例验证五路——精确级 3、MatchPrefix
  前缀级 4、aliasB 前缀全等 4、其他级 5、二次回退 b2 精确 6、nil 术语/空 pattern 5，均 PASS。
- **G4 review**：扩展 `backend/filesearch_pinyin_windows.go`（3 函数）+ 测试；vet/test/build
  复验通过。

## 遗留（下一批）

- §10 差集 54 文件。filesearch 拼音模糊匹配 + 分类域已闭环（Match/MatchExact/MatchPrefix/
  matchFileSearchPinyinTerm/matchNodeNameTermsWithPinyin/classifyVolumePinyinMatch）。
  候选：`volumeIndexReadView.matchSearchCandidateNodeWithPinyin`（0x140810440，依赖 classify 与
  matchNodeNameTermsWithPinyin 均已就绪）、`searchPinyinContextWithTombstones`（0x140810e60）、
  `pluginupdate.go`、`pluginwindow.go`、`twofactor_provision_misc.go` 各存根、`transport.go`、
  `TestReminderNotification`（0x1407a7320）。
