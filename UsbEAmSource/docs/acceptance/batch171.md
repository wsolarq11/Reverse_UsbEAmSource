# 批次 171 验收（filelocator_runtime.go：路径过滤器解析链 + 字符串匹配器内部构造签名订正）

日期：2026-09-25
子批次：parseFileLocatorFilterRule / splitFileLocatorFilterRuleType 升 [S]；newFileLocatorPathMatcher 升 [S]；
newFileLocatorStringMatcherInternal 签名订正

## 目标

文件定位器「路径过滤器」解析链落地：`parseFileLocatorFilterRule`（0x1407d3840）与
`splitFileLocatorFilterRuleType`（0x1407d3a40）从 [P]/[S-sig] 升 [S]（完整 asm 体），
并连带把 `newFileLocatorPathMatcher`（0x1407d3c00）[S-sig]→[S]、`newFileLocatorStringMatcherInternal`
（0x1407d4420）签名订正（补 booleanScope+wholeWord 两参）。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go test ./backend` | ok (0.4s) |
| 黄金用例 | `TestSplitFileLocatorFilterRuleType` / `TestParseFileLocatorFilterRule` | PASS |

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2659 / S=1104 / S-inline=35 / S-sig=1342 / P=178 / UNMARKED=0`。

**真函数 = 1104 + 35 + 1342 = 2481 / 4754 = 52.2%**（相对批次 170 **增 +1**；P 179→**178**，
S 1101→1104、S-sig 1344→1342——parseFileLocatorFilterRule [P]→[S]，splitFileLocatorFilterRuleType 与
newFileLocatorPathMatcher 由 [S-sig]→[S]）。

重建产物 SHA256：`06D926AD798F19B5AC9987994F7BC6E3AA4F55E9FD8C24A96C2C9F432FB63BC3`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细（4 函数，`backend/filelocator_runtime.go`）

| 函数 | VA | 旧 | 新 |
|---|---|---|---|
| parseFileLocatorFilterRule | 0x1407d3840 | [P] (rule)(string,string) | [S] (rule, filterType string)(include bool, typ, pattern string) |
| splitFileLocatorFilterRuleType | 0x1407d3a40 | [S-sig] (rule) string | [S] (rule string)(typ, pattern string, ok bool) |
| newFileLocatorPathMatcher | 0x1407d3c00 | [S-sig] (pattern, filterType) | [S] 完整体（glob 分支 + internal 分支） |
| newFileLocatorStringMatcherInternal | 0x1407d4420 | [S-sig] (query,mode,matchCase) | [S-sig] (query,mode,matchCase,booleanScope,wholeWord) |

## G4 关键实证结论

1. **parseFileLocatorFilterRule 双入参 + 三返回**：调用点 0x1407d3530 传 (rax,rbx)=rule、(rcx,rdi)=
   filterType；返回 eax=include bool、(rbx,rcx)=typ、(rdi,rsi)=pattern（`[S]` 体含 '!'/'-'→exclude、
   '+'→include 前缀剥离 + TrimSpace）。
2. **splitFileLocatorFilterRuleType 三返回**：`test sil; je` 判定 ok bool；首字符 ToLower 后按
   0x62='b'/0x67='g'/0x70='p'/0x72='r' 分发，canonical 常量逐一解码——b→"boolean"(7B)、
   g→"glob"(4B)、p→"plain"(5B)、r→"regex"(5B)，与 normalizeFileLocatorFilterType 白名单一致。
3. **newFileLocatorPathMatcher 完整体**：normalizeFileLocatorFilterType(filterType)=="glob" 走
   newFileLocatorGlobPathMatcher(pattern)，否则 newFileLocatorStringMatcherInternal(pattern,
   filterType, false, "file", false)——0x1407d3c64 esi=0(matchCase)、0x1407d3c66 r8="file"(4B)、
   0x1407d3c73 r10d=0(wholeWord)。
4. **newFileLocatorStringMatcherInternal 签名补全**：序言 spill rax/sil/r10b/r9/r8/rdi/rcx 共 8 寄存器
   = query(2)+mode(2)+matchCase(1)+booleanScope(2)+wholeWord(1)；旧 [S-sig] 漏 booleanScope+wholeWord，
   本批订正（体仍 [S-sig]，mode 分发 regex/plain/boolean 待专项）。

## 残留 / 未落地（不触及）

- newFileLocatorStringMatcherInternal 体（TrimSpace→ToLower→normalize mode/scope→newobject 装配→mode
  分发）已解到结构，regex 分支调 compileFileLocatorRegex、plain/boolean 分支待逐条，下一批升 [S]。
- filelocator_runtime.go 剩余 22 个 [P]（runSearch/walkRoot/processFile 等引擎方法）待逐寄存器实证。
