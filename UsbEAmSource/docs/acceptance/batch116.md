# 批次 116 验收（应用工作目录解析）

日期：2026-09-23
子批次：app-launch-workdir
目标：落地 `resolveApplicationWorkingDirectory` [S-sig]（应用启动工作目录解析）。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 定向 test | `go test -tags production -count=1 -p=1 -run 'ResolveApplicationWorkingDirectory' ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1192 / MARKED=1192 / S=885 / S-inline=34 / S-sig=272 / P=1 / UNMARKED=0`。

相对批次 115（1191/885/34/271/1/0）：1 个新函数 `resolveApplicationWorkingDirectory` `[S-sig]`。

## G3 逻辑等价

`backend/process_launch_windows.go`（1 函数）：

- **resolveApplicationWorkingDirectory**（0x1408a74c0, 288B）：normalizeAppEntryType(entryType)
  =="directory"（9B 精确匹配 @0x1408a7520）→ ""；entry=TrimSpace 非空（@0x1408a7560）→
  entry；appName=TrimSpace 且 looksLikeFilesystemPath（@0x1408a7589）→ inferWorkingDir(appName)
  （@0x1408a75a0）；否则 ""。

标注：完整 launchContext 结构体（第 8+ 词栈参数，含 resolveLaunchableAppEntry 上下文，
第 14-27 词 7 个 string 字段）未完全还原（待取证），此处按最小三参签名落地核心逻辑。

## G4 测试

`backend/process_launch_windows_test.go` 新增 `TestResolveApplicationWorkingDirectory`（7 用例）：
目录条目/大小写/entry 非空/空白修剪/appName 推断/无路径/全空。全 PASS。

## 判定

四路 PASS，批次 116 闭环。`resolveApplicationWorkingDirectory` 核心逻辑 [S-sig] 实证；
完整 launchContext 结构体为批次 117+ 的 startApplicationWindowsAsAdmin /
startApplicationUsingCurrentPrivileges 全 [S] 升级的前置取证目标。
