# 批次 216 — filelocator 三 [P] 升档 [S-sig]（收集正项匹配器 + 词项匹配器构造 + 句柄读文本）

## 基线 / 收口

| 指标 | 基线（批次 215 收口） | 收口（批次 216） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1390 | **1393** |
| P | 106 | **103** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2682 | **2685**（56.48%） |

SHA256 `968A4D5C8F58035F6EDFE08D410B8D6F52E236FEB4AA20ADE8B8F0B9141FAEC7`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/filelocator_runtime.go` 三个 [P] 存根经汇编逐寄存器实证后升档 [S-sig]（签名修正，
体未逐条翻译保持零值）。

### 升档 [S-sig]（签名实证修正）

- `collectFileLocatorPositiveMatchers`（0x1407dcf60）：签名 `(node, out)` → `(node, positive bool, out)`。
  汇编实证：node(接口 2 word:rax/rbx)+positive(bool:cl)+out(*[]fileLocatorTermMatcher:rdi) 共 4 寄存器，
  返回 void。体按 itab hash 分发二元/非/项节点，递归 left/right（非节点 `xor ecx,1` 翻转 positive），
  项节点 positive 时 append 到 out（元素 `shl rcx,4`=16B 接口），未逐条翻译。
- `newFileLocatorTermMatcher`（0x1407dd1c0）：签名 `(query, mode)` → `(query, mode, matchCase bool)`。
  汇编实证：query(string:rax/rbx)+mode(string:rcx/rdi)+matchCase(bool:sil)；返回 fileLocatorTermMatcher
  接口。体 TrimSpace+ToLower(mode) 后与 "like"（`0x656b696c`）比对，like→wildcardToFileLocatorRegex，
  其余→compileFileLocatorRegex(regex,matchCase)，未逐条翻译。
- `readFileLocatorTextFromHandle`（0x1407da360）：签名 `(r, maxSize)` → `(r, matcher, maxSize)`。
  汇编实证：r(io.Reader 2 word)+matcher(*fileLocatorStringMatcher:rcx, nil 检查)+maxSize(int:rdi, jle 检查)；
  返回 (string,error)。matcher 经 newobject 存入上下文对象 +0x08 字段；体 io.ReadAll→
  isTextLikeFileLocatorContent→decodeFileLocatorText，超限/非文本走 error 路径，未逐条翻译。

## 关键知悉

- `collectFileLocatorPositiveMatchers` 的 out 元素为 16B 接口（`fileLocatorTermMatcher`），与
  `fileLocatorBooleanExpression.positiveMatchers []fileLocatorTermMatcher` 定义一致；NOT 节点
  `xor ecx,1` 翻转极性，印证 boolean 表达式的负项过滤语义。
- `readFileLocatorTextFromHandle` 的上下文对象字段 [0x08]=matcher、[0x10]=maxSize+1，经 convT
  转接口后喂 io.ReadAll，印证"句柄→受 matcher 约束的读取"链。
- `parseFileLocatorBooleanExpression`（0x1407dac60）初勘确认 query(string)+2 额外 qword+1 bool
  参数，parser 结构体扩展为 tokens+pos+2 字段+bool（较当前 tokens+pos 定义多 3 字段），完整签名
  待专项（需沿 tokenize 后续递归路径确证 2 额外 qword 语义）。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1393 / P=103 / UNMARKED=0`。
- **G3 行为**：纯签名修正（体零值→体零值），无行为变更；既有测试全量回归 PASS。
- **G4 review**：`filelocator_runtime.go`（3 函数签名修正）；vet/test/build 复验通过。

## 遗留（下一批）

- §10 差集 54 文件。filelocator 域剩余 [P]：parseFileLocatorBooleanExpression（parser 结构体扩展
  3 字段）、fileLocatorProximityRanges（递归遍历节点树，接口+节点指针+bool）、MatchContent 族
  （matchFileLocatorLineBased/BooleanPerLine/BooleanAcrossFile/buildFileLocatorLineMatch 均为 9 寄存器）、
  matchFileLocatorContentStreamContext、fileLocatorStreamingLineMatch、newFileLocatorTextDocument。
- walkRoot/processFile 的 2 个指针实参类型待 WalkDir 闭包 func1（0x1407d10c0）捕获链确证。
- searchPinyinContextWithTombstones（0x140810e60）含排序/合并/去重三段仍待专项。
