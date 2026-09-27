# 批次 217 — filelocator 四 [P] 升档 [S-sig]（流式行匹配 + 文本文档 + 布尔解析 + 邻近区间）

## 基线 / 收口

| 指标 | 基线（批次 216 收口） | 收口（批次 217） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1393 | **1397** |
| P | 103 | **99** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2685 | **2689**（56.56%） |

SHA256 `42AB6442E798F212CDBD0E7365F472A93F3BFC297DA48E11D0A7037E33D881F8`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/filelocator_runtime.go` 四个 [P] 存根经汇编逐寄存器实证后升档 [S-sig]（签名修正，
体未逐条翻译保持零值）。

### 升档 [S-sig]（签名实证修正）

- `fileLocatorStreamingLineMatch`（0x1407d78e0）：签名 `(matcher, line) bool` →
  `(line string, matcher *fileLocatorStringMatcher) ([]FileLocatorTextRange, int)`。
  汇编实证：line(string:rax/rbx)+matcher(*fileLocatorStringMatcher:rcx)；返回 slice 3 word + int:rdi。
  体 matcher+0x58 布尔表达式存在时走 HighlightRanges(line)（int=max(len,1)），否则走 Ranges(line)。
- `newFileLocatorTextDocument`（0x1407d7be0）：返回 `*fileLocatorTextDocument` → `fileLocatorTextDocument`（值）。
  汇编实证：text(string) 参数，返回 8 word 值（栈返回）；结构体实测 64B = text(2)+lines(3)+runes(3)，
  较旧定义多 runes [][]rune 字段。体 Replace(\r\n→\n)→genSplit(\n)→makeslice([][]rune) 每行 stringtoslicerune。
- `parseFileLocatorBooleanExpression`（0x1407dac60）：签名 `(query)` → `(query, booleanScope, matchCase bool)`。
  汇编实证：query(string)+booleanScope(string)+matchCase(bool)；parser 结构体扩展为
  tokens+pos+booleanScope+matchCase（较旧定义多 2 字段）。体 tokenizeFileLocatorBoolean(query) 后递归下降。
- `fileLocatorProximityRanges`（0x1407dc640）：签名 `(left,right,distance) []` →
  `(node fileLocatorBooleanNode, distance int, active bool) ([]FileLocatorTextRange, bool)`。
  汇编实证：node 接口 2 word+distance(int)+active(bool)；递归 left/right 透传 distance/active，结果按
  distance 合并去重。返回值含 bool（active 聚合标志）。

## 关键知悉

- `fileLocatorTextDocument` 补全 runes [][]rune 字段后实测 8 word=64B，与返回值栈布局对齐；此字段
  供后续 matchFileLocatorLineBased/BooleanPerLine/BooleanAcrossFile 逐行匹配使用。
- `fileLocatorBooleanParser` 补全 booleanScope(string)+matchCase(bool) 字段后对齐 0x1407dac60 装配布局
  （tokens@0x00、pos@0x18、booleanScope@0x20、matchCase@0x30）。
- `fileLocatorProximityRanges` 参数 3 active(bool) 与递归返回 bool 同构（`test dl/dil`），印证邻近
  区间的"激活→聚合"语义；当前存根 `(left,right,distance)` 的 7 参数形态为误判。
- `matchFileLocatorContentStreamContext`（0x1407d5d80）初勘确认 ctx(2)+matcher(1)+maxSize(int)+r(1)
  四参数（含 jle/nil 检查），返回 slice 结构待专项。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1397 / P=99 / UNMARKED=0`。
- **G3 行为**：纯签名修正（体零值→体零值），无行为变更；既有测试全量回归 PASS。
- **G4 review**：`filelocator_runtime.go`（4 函数签名修正 + 2 结构体扩展）；vet/test/build 复验通过。

## 遗留（下一批）

- §10 差集 54 文件。filelocator 域剩余 [P]：walkRoot、processFile、MatchContent 族
  （matchFileLocatorLineBased/BooleanPerLine/BooleanAcrossFile/buildFileLocatorLineMatch 均为 9 寄存器）、
  matchFileLocatorContentStreamContext。P 已跌破 100。
- walkRoot/processFile 的 2 个指针实参类型待 WalkDir 闭包 func1（0x1407d10c0）捕获链确证。
- searchPinyinContextWithTombstones（0x140810e60）含排序/合并/去重三段仍待专项。
