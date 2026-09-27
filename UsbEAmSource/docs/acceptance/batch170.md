# 批次 170 验收（filelocator_runtime.go：两个路径匹配器构造器 [P]→[S-sig]/[S]）

日期：2026-09-25
子批次：文件定位器路径匹配器构造链（newFileLocatorPathMatcher / newFileLocatorGlobPathMatcher）

## 目标

`newFileLocatorPathMatcher`（0x1407d3c00）、`newFileLocatorGlobPathMatcher`（0x1407d3cc0）
两个 `[P]` 升真函数级：前者升 [S-sig]（签名坐实 + arg 顺序订正），后者升 [S]（完整 asm 体）。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go test ./backend` | ok (2.0s) |
| 黄金用例 | `TestNewFileLocatorGlobPathMatcher` | PASS |

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2659 / S=1101 / S-inline=35 / S-sig=1344 / P=179 / UNMARKED=0`。

**真函数 = 1101 + 35 + 1344 = 2480 / 4754 = 52.2%**（相对批次 169 **增 +2**；P 181→**179**，
S 1100→1101、S-sig 1343→1344，两函数由 [P] 升为真函数级）。

重建产物 SHA256：`AADEA2E07D2B160203B74AA19148A595B8CD1600F5113FC39D6969AAD59E4AE4`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细（2 函数，`backend/filelocator_runtime.go`）

| 函数 | VA | 旧 | 新 |
|---|---|---|---|
| newFileLocatorPathMatcher | 0x1407d3c00 | [P] (filterType, pattern) | [S-sig] (pattern string, filterType string) *fileLocatorStringMatcher |
| newFileLocatorGlobPathMatcher | 0x1407d3cc0 | [P] (pattern) | [S] (pattern string) *fileLocatorStringMatcher（完整体） |

## G4 关键实证结论

1. **arg 顺序颠倒订正（newFileLocatorPathMatcher）**：0x1407d3c22 `mov rax,rcx; mov rbx,rdi` 把
   (rcx,rdi)=第 2 串传给 `normalizeFileLocatorFilterType`，0x1407d3c3b 把 (rax,rbx)=第 1 串传给
   `newFileLocatorGlobPathMatcher`——故第 1 串是 pattern、第 2 串是 filterType，旧树 `(filterType, pattern)`
   颠倒。调用点 0x1407d35ac（buildFileLocatorPathFilter 内）亦显示 String1=pattern、String2=filterType。
2. **newFileLocatorGlobPathMatcher 完整体**：TrimSpace 后 `test rbx`（空串返 nil）→
   `compileFileLocatorPathGlobMatcher` 返回 (rax=compiled, rbx=err)，`test rbx` 非零返 nil →
   `runtime.newobject` 分配 fileLocatorStringMatcher 后按偏移写字段：`[rax]=query.ptr`、`[rax+8]=query.len`、
   `[rax+0x20]=mode.ptr`、`[rax+0x28]=mode.len=5`、`[rax+0x48]=regex`。字段偏移与
   types_filelocator.go L174 的 fileLocatorStringMatcher（query@0x00 / mode@0x20 / regex@0x48）逐一吻合。
3. **mode 常量**：glob 路径匹配器的 mode = `"regex"`（5B@0x140C33F54，hex `7265676578`）——glob 经
   `fileLocatorPathGlobToRegex` 转正则后按 regex 模式承载，与字符串匹配器结构一致。
4. **返回类型确认**：newobject 的类型描述符 + 字段偏移（query/mode/regex）证明返回类型为
   `*fileLocatorStringMatcher`，非独立的 fileLocatorPathMatcher 接口（旧树推断正确）。

## 残留 / 未落地（不触及）

- `newFileLocatorPathMatcher` 非 glob 分支调 `newFileLocatorStringMatcherInternal`（0x1407d3c76）时
  传入 8 寄存器（3 string + 2 bool），与现有 [S-sig] `(query,mode,matchCase)`（5 寄存器）不符——该
  [S-sig] 签名疑似缺失 booleanScope+wholeWord 两参，待下一批 asm 专项校正后 newFileLocatorPathMatcher
  才可升 [S]。
- filelocator_runtime.go 剩余 23 个 [P]（runSearch/walkRoot/processFile 等引擎方法）待逐寄存器实证。
