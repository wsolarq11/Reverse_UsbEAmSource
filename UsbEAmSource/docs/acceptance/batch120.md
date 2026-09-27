# 批次 120 验收（COM ShellExecute 参数序修正 + startShellTargetViaExplorer 升格）

日期：2026-09-23
子批次：shell-execute-signature
目标：修正 shellExecuteByExplorer 参数序（batch118 落体命名错误），并升格 startShellTargetViaExplorer 为 [S]。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.186s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1204 / MARKED=1204 / S=890 / S-inline=34 / S-sig=279 / P=1 / UNMARKED=0`。

相对批次 119（1204/888/34/281/1/0）：S +2，S-sig −2（两个 [S-sig] 升格 [S]，函数总数不变）。

## G3 逻辑等价

`backend/com_explorer_windows.go` + `backend/pathclip_explorer_windows.go`：

- **shellExecuteByExplorer**（0x1408abe60）[S-sig]→[S]：参数序从
  `(dispatch, verb, appName, args, workingDir, show)` 修正为
  `(dispatch, file, args, dir, verb, show)`，InvokeWithOptionalArgs 参数数组序
  `[file, args, dir, verb, show]`，对齐 ShellExecute(sFile, vArguments, vDirectory,
  vOperation, vShow) MSDN 语义。
- **startShellTargetViaExplorer**（0x1408aaee0, 352B）[S-sig]→[S]：TrimSpace(shellTarget)
  空 → `errors.New("Shell 目标不能为空")`（24B @0x140c651a8）；
  buildShellExecuteArguments(args) → withExplorerShellDispatch 闭包（func1 @0x1408ab040）
  → TrimSpace(workingDir) → shellExecuteByExplorer(dispatch, name, shellArgs, dir, "", 1)
  （verb 恒空 @0x1408ab0a4/0a7，show=1 @0x1408ab087）。

## G4 参数序交叉实证

两独立调用点 asm 寄存器映射一致，推翻 batch118 的 `(verb, appName, args, workingDir)` 序：

| 调用点 | rbx/rcx | rdi/rsi | r8/r9 | r10/r11 | [rsp] |
|---|---|---|---|---|---|
| startShellTargetViaExplorer.func1 (0x1408ab040) | file(name) | args(shellArgs) | dir(TrimSpace(workingDir)) | verb(空) | show=1 |
| startApplicationViaExplorer.func1 (0x1408a9780) | file(name) | args(shellArgs) | dir(resolveApplicationWorkingDirectory) | verb(空) | show=1 |

## 判定

四路 PASS，批次 120 闭环。shellExecuteByExplorer 参数序经双调用点交叉实证修正，
startShellTargetViaExplorer 升格 [S]。
