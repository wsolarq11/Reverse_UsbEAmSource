# 批次 109 验收（资源管理器启动/定位 + Windows 参数转义）

日期：2026-09-23
子批次：pathclip-explorer
目标：落地 `openWithDefaultHandler` [S-sig]→[S]、`quoteWindowsArgument` [S]、
`revealPathInExplorer` [S]、`startShellTargetViaExplorer` [S-sig] 签名。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 定向 test | `go test -tags production -count=1 -p=1 -run 'QuoteWindowsArgument' ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1183 / MARKED=1183 / S=874 / S-inline=34 / S-sig=274 / P=1 / UNMARKED=0`。

相对批次 108（1180/871/34/274/1/0）：3 个新函数（`quoteWindowsArgument`/`revealPathInExplorer`
`[S]`，`startShellTargetViaExplorer` `[S-sig]`）+ `openWithDefaultHandler` [S-sig]→[S]
（S +3，S-sig 净 0）。`UNMARKED=0` 保持。

## G3 逻辑等价

新文件 `backend/pathclip_explorer_windows.go`（3 函数）+ `bootstrap_pathclip.go` `openWithDefaultHandler`：

- **openWithDefaultHandler**（0x1408a8de0, 256B）：TrimSpace 空 → newPathError；
  `shouldOpenPathViaExplorer` 真 → `startShellTargetViaExplorer(path, nil, "")`；否则
  `withShellApartment(shellExecuteProgram("", path, "", ""))`。参数映射实证：rax/rbx=path、
  rcx/rdi/rsi=args(nil)、r8/r9=verb("")，func1 @0x1408a8ee0 反汇编为
  `shellExecuteProgram("",path,"","")`。
- **quoteWindowsArgument**（0x1408ad280, 1632B）：MS CRT argv 转义（Unicode 感知，逐 rune
  decoderune）。空 → `""`；`IndexAny(arg," \t\n\v\"")<0`（字符集 5B @0x140c35d0d）→ 原样；
  否则 Builder.Grow(len+8) → 前缀 `"` → `\` 计数、`"` 时 `Repeat(\`,2n+1)+`"`、其余
  `Repeat(\`,n)+WriteRune` → 结尾残留 `Repeat(\`,2n)` → 后缀 `"`。
- **revealPathInExplorer**（0x1408a8d20, 128B）：TrimSpace → quoteWindowsArgument →
  `concatstring2("/select,", quoted)`（"/select," 8B @0x140c3c24c）→
  `withShellApartment(shellExecuteProgram("", "explorer.exe", args, ""))`（func1 @0x1408a8da0，
  program "explorer.exe" 12B @0x140c4b9fc）。
- **startShellTargetViaExplorer**（0x1408aaee0, [S-sig]）：签名实证
  `(path string, args []string, verb string) error`；体含 buildShellExecuteArguments +
  withExplorerShellDispatch（COM 域，批次 110+ 落地）。

## G4 测试

`backend/pathclip_explorer_windows_test.go` 2 用例：

- `TestQuoteWindowsArgument`：13 例（空/原样/反斜杠不触发/空格包裹/引号转义/反斜杠后
  引号翻倍/结尾翻倍/Unicode）。
- `TestQuoteWindowsArgumentUnicode`：含空格 Unicode 包裹、无特殊字符 Unicode 原样。

全 PASS。转义结果与目标二进制 asm 逐分支核对（区别于 syscall.EscapeArg：quoteWindowsArgument
额外加首尾引号）。

## 判定

四路 PASS，批次 109 闭环。`openWithDefaultHandler` 完整落地（依赖链 5/5 齐），
`quoteWindowsArgument` 与 `revealPathInExplorer` 全 `[S]`；`startShellTargetViaExplorer`
签名确定、体留待 COM ShellDispatch 链（withExplorerShellDispatch/shellExecuteByExplorer）。
