# 批次 111 验收（命令行打开路径：直启 / 提权 runas / 降权 stub）

日期：2026-09-23
子批次：pathclip-commandline
目标：落地 `openPathCommandLine` [S-sig]→[S]、`openPathCommandLineAdmin` [S-sig]→[S]、
`openUnelevatedCommandLine` [S-sig] 签名。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 定向 test | `go test -tags production -count=1 -p=1 -run 'ResolveOpenLocationDirectory|OpenPathDirectoryEmpty|OpenPathCommandLine' ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1185 / MARKED=1185 / S=878 / S-inline=34 / S-sig=272 / P=1 / UNMARKED=0`。

相对批次 110（1184/876/34/273/1/0）：1 个新函数 `openUnelevatedCommandLine` `[S-sig]`；
`openPathCommandLine`、`openPathCommandLineAdmin` [S-sig]→[S]（S +2，S-sig 净 -1）。`UNMARKED=0` 保持。

## G3 逻辑等价

`bootstrap_pathclip.go`（3 函数）：

- **openPathCommandLine**（0x1408a8ac0, 288B）：resolveOpenLocationDirectory 空 →
  `errors.New("位置不能为空")`；`isProcessElevated` 真 → `openUnelevatedCommandLine(dir)`；
  假 → `startDetachedCommand([resolveCmdExePath(), "/k", "pushd", dir])`（4 元素 slice
  @0x1408a8b76；"/k" 2B @0x140c336b3、`pushd` 5B @0x140c35d03）。
- **openPathCommandLineAdmin**（0x1408a8be0, 512B）：resolveOpenLocationDirectory 空 →
  `errors.New("位置不能为空")`；`withShellApartment(shellExecuteProgram("runas", cmd,
  "/k pushd "+quoteCmdArgument(dir), ""))`（func1 @0x1408a8c80；verb "runas" 5B
  @0x140c35d08、前缀 "/k pushd " 9B @0x140c3f784）。
- **openUnelevatedCommandLine**（0x1408aa000, [S-sig]）：签名实证 `(dir string) error`；体含
  buildCreateProcessCommandLine + createProcessWithShellParentWithVisibility（进程创建链，
  批次 112+ 落地）。

## G4 测试

`backend/bootstrap_pathclip_test.go` 新增 2 用例：

- `TestOpenPathCommandLineEmpty`：空路径 → `位置不能为空`。
- `TestOpenPathCommandLineAdminEmpty`：空路径 → `位置不能为空`。

全 PASS（非空路径分支会真实启动 cmd.exe/runas，副作用跳过）。

## 判定

四路 PASS，批次 111 闭环。`openPathCommandLineAdmin` 完整落地（依赖链 6/6 齐：
resolveOpenLocationDirectory/resolveCmdExePath/quoteCmdArgument/shellExecuteProgram/
withShellApartment），`openPathCommandLine` 完整落地（除 openUnelevatedCommandLine 依赖）。
错误消息实测为 `位置不能为空`（区别于 newPathError 的 `路径不能为空`）。
