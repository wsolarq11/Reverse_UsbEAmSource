# 批次 115 验收（ShellExecute 参数组装）

日期：2026-09-23
子批次：shell-execute-args
目标：落地 `buildShellExecuteArguments` [S]（[]string → ShellExecute 参数串）。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 定向 test | `go test -tags production -count=1 -p=1 -run 'BuildShellExecuteArguments' ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1191 / MARKED=1191 / S=885 / S-inline=34 / S-sig=271 / P=1 / UNMARKED=0`。

相对批次 114（1190/884/34/271/1/0）：1 个新函数 `buildShellExecuteArguments` `[S]`。

## G3 逻辑等价

`backend/process_launch_windows.go`（1 函数）：

- **buildShellExecuteArguments**（0x1408ad020, 608B）：len==0 → ""；make([]string,0,len(args))
  （len>2 走 makeslice，否则栈缓冲）；逐 arg TrimSpace，空跳过，非空 quoteWindowsArgument
  → append；strings.Join(quoted, " ")（分隔符空格 1B @0x1411ca5c8）。

## G4 测试

`backend/process_launch_windows_test.go` 新增 `TestBuildShellExecuteArguments`（7 用例）：
空/单参数/空格分隔/含空格参数转义/空参数跳过/首尾空白修剪。全 PASS。

## 判定

四路 PASS，批次 115 闭环。`buildShellExecuteArguments` 全 `[S]`，补齐 startApplicationWindowsAsAdmin
与 startShellTargetViaExplorer 的公共依赖。
