# 批次 119 验收（Shell 动词 + 启动托盘进程域）

日期：2026-09-23
子批次：shellverb-startup-tray
目标：落地 Shell 动词分派 + 启动托盘进程 4 函数（3 [S] + 1 [S-sig]）。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 定向 test | `go test -tags production -count=1 -p=1 -run 'LauncherStartedForStartupTray\|InvokeShellVerbEmpty' ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1204 / MARKED=1204 / S=888 / S-inline=34 / S-sig=281 / P=1 / UNMARKED=0`。

相对批次 118（1200/885/34/280/1/0）：4 个新函数（3 [S] + 1 [S-sig]）。

## G3 逻辑等价

`backend/shellverb.go`（4 函数）：

- **launcherStartedForStartupTray**（0x1408af720, 192B）[S]：遍历 args 逐项 TrimSpace+
  EqualFold("--usbeam-startup-tray")，命中 true。
- **prepareLauncherStartupTrayProcess**（0x1408af7e0, 96B）[S]：os.Executable → filepath.Abs →
  filepath.Dir → os.Chdir。
- **invokeShellVerb**（0x1408a9de0, 208B）[S]：TrimSpace 双参 → 空则 errors.New("路径或动作不能为空")
  → EqualFold("properties")→showShellProperties，否则 invokeShellVerbWithPowerShell。
- **invokeShellVerbWithPowerShell**（0x1408ada40, 350 行）[S-sig]：单引号转义（Replace '→''）+ 模板拼接
  实证；完整 PowerShell 脚本模板待取证，以 Shell.Application ShellExecute 脚本等价承载。

## G4 测试

`backend/shellverb_test.go`（8 用例）：托盘标记大小写/空白/多参数/空；invokeShellVerb 空参错误消息。
全 PASS。

## 判定

四路 PASS，批次 119 闭环。这是「往下做 20 个批次」（批次 100–119）的收尾轮。
