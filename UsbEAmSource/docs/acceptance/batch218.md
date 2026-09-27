# 批次 218 — 内容匹配链五 [P] 升档 [S-sig] + 一 [S-sig] 签名纠错（doc 8 字值传递）

## 基线 / 收口

| 指标 | 基线（批次 217 收口） | 收口（批次 218） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1397 | **1403** |
| P | 99 | **93** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2689 | **2695**（56.69%） |

SHA256 `4FA8D63618BA7C5F7A1E34E89EE778E58830D97191FE2C98E2F83434F08EEE8B`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

突破点：`matchFileLocatorContentContext`（0x1407d53e0）完整 dump 显示其先 `newFileLocatorTextDocument`
（返回 8 word 值）再把 doc 原样透传给 `MatchContent`——从而一举确证 MatchContent 族的「8 字值传递」形态。
`backend/filelocator_runtime.go` 五个 [P] 升档 [S-sig] + 一个 [S-sig] 签名纠错。

### 升档 / 纠错

- `fileLocatorStringMatcher.MatchContent`（0x1407d4ee0）：`(content string)` → `(doc fileLocatorTextDocument)`。
  汇编实证：recv(matcher:rax)+doc(fileLocatorTextDocument 8 word:rbx..r11 值传递)。按 mode 分发
  boolean(across/per-line)→expr=matcher+0x58；非 boolean→matchFileLocatorLineBased(matcher)。
- `matchFileLocatorContentContext`（0x1407d53e0）签名纠错：`(ctx, matcher, text)` →
  `(ctx, path string, maxSize int, matcher)`。汇编实证 6 寄存器：ctx(2)+path(string:rcx/rdi，喂 os.OpenFile)+
  maxSize(int:rsi，默认 1MB)+matcher(r8)。返回 ([]FileLocatorLineMatch, error)。
- `matchFileLocatorLineBased`（0x1407d7dc0）：`(m, doc*fileLocatorTextDocument)` →
  `(doc fileLocatorTextDocument, m *fileLocatorStringMatcher)`。第 9 参 r11=matcher 本身（无布尔表达式分支）。
- `matchFileLocatorBooleanPerLine`（0x1407d81e0）：→ `(doc, expr *fileLocatorBooleanExpression)`。第 9 参 r11=matcher+0x58。
- `matchFileLocatorBooleanAcrossFile`（0x1407d86c0）：→ `(doc, expr *fileLocatorBooleanExpression)`。第 9 参 r11=matcher+0x58。
- `buildFileLocatorLineMatch`（0x1407d8ce0）：`(lineNumber,line,ranges)` → `(doc fileLocatorTextDocument, index int)`。
  汇编实证：doc.lines[index]（`shl rdx,4` 16B 元素）转 rune 后按 ranges 截取 before/after；返回 FileLocatorLineMatch
  大结构（栈返回）。

## 关键知悉

- 内容匹配调用链：processFile → matchFileLocatorContentContext(ctx,path,maxSize,matcher) →
  newFileLocatorTextDocument(text) → matcher.MatchContent(doc) → 按 mode 分发
  matchFileLocatorBooleanAcrossFile/BooleanPerLine(doc,expr) 或 matchFileLocatorLineBased(doc,matcher) →
  buildFileLocatorLineMatch(doc,index)。
- fileLocatorTextDocument 作为 8 word（64B）值在整条链上寄存器直传（不装箱），与批次 217 补全的
  runes 字段（8 word 布局）完全自洽。
- 三 match 函数第 9 参不同：LineBased 收 matcher，BooleanPerLine/AcrossFile 收 expr（matcher+0x58）。
- matchFileLocatorContentContext 旧 [S-sig] 注释「params 5=ctx+matcher+text」系误判，本批纠正为 6 参数；
  processFile 调用点（L152 前 6 寄存器 rax..r8）与之对齐。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1403 / P=93 / UNMARKED=0`。
- **G3 行为**：纯签名修正（体零值→体零值），无行为变更；既有测试全量回归 PASS。
- **G4 review**：`filelocator_runtime.go`（6 函数签名修正）；vet/test/build 复验通过。

## 遗留（下一批）

- filelocator 域剩余 [P]：walkRoot（0x1407d0d20）、processFile（0x1407d15a0）——两者 a/b 指针实参
  指向 runSearch 栈上的结果累加器（[0]=count、[0x48]/[0x50]=paths slice 头）与双层闭包，待闭包链专项确证。
- §10 差集 54 文件。P 已降至 93。后续转向 windowmanagement/screenshot/oledblackout 等 Windows 域
  （多为匿名上下文结构体展开，需逐结构落地）。
- searchPinyinContextWithTombstones（0x140810e60）含排序/合并/去重三段仍待专项。
