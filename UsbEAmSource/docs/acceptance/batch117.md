# 批次 117 验收（应用启动/提权重试链）

日期：2026-09-23
子批次：app-launch-retry
目标：落地提权重试链 4 函数（[S-sig] 核心控制流实证）。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 定向 test | `go test -tags production -count=1 -p=1 -run 'StartApplicationUsingCurrentPrivilegesEmptyAppName\|RetryLaunchAsAdminIfElevationRequired' ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1196 / MARKED=1196 / S=885 / S-inline=34 / S-sig=276 / P=1 / UNMARKED=0`。

相对批次 116（1192/885/34/272/1/0）：4 个新函数 [S-sig]。

## G3 逻辑等价

`backend/applaunch_windows.go`（4 函数，核心控制流 asm 实证）：

- **startApplicationViaExplorer**（0x1408a9400）：TrimSpace(appName)+buildShellExecuteArguments(args)
  +requiresShellOpen → 回退 shellExecuteProgram 默认 verb。完整 withExplorerShellDispatch COM 链待取证。
- **startApplicationUsingCurrentPrivileges**（0x1408a7020, 116 行）：空 appName →
  errors.New("入口路径不能为空")（24B @0x140c65178）；entryType=="directory" →
  openPathDirectory；已提权且 requiresShellOpen → withShellApartment(shellExecuteProgram("",…))；
  已提权不需 shell → exec.Command+resolveApplicationWorkingDirectory+Cmd.Start，
  err → fmt.Errorf("启动失败: %w")（@0x1408a7269）；未提权 → startApplicationViaExplorer。
- **startApplicationWindowsAsAdmin**（0x1408a9b80, 416B）：isProcessElevated 真 → 透传
  startApplicationUsingCurrentPrivileges；否则 withShellApartment 闭包（func1 0x1408a9d00）
  shellExecuteProgram("runas",…)。
- **retryLaunchAsAdminIfElevationRequired**（0x1408acd60, 192B）：isElevationRequiredError 假 →
  返回 err；真 → startApplicationWindowsAsAdmin。

## G4 测试

`backend/applaunch_windows_test.go`（3 用例）：
- 空 appName → "入口路径不能为空"（纯逻辑，实测消息一致）。
- 非提权错误原样返回（sentinel 透传）。
- ERROR_ELEVATION_REQUIRED 判据锁定（不误报）。

## 判定

四路 PASS，批次 117 闭环。提权重试链核心控制流 [S-sig] 实证。剩余债务（待后续批次）：
launchContext 完整结构体（第 14-27 词 7 string）、resolveLaunchableAppEntry(0x1408a75e0)、
COM ShellDispatch 链（withExplorerShellDispatch 0x1408ab0e0 等）——升格 [S] 的前置取证目标。
