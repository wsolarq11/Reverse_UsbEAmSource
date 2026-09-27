# 批次 123 验收（应用启动链 launchContext 签名升格）

日期：2026-09-23
子批次：startApplication-domain-refactor

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.358s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1213 / MARKED=1213 / S=904 / S-inline=34 / S-sig=274 / P=1 / UNMARKED=0`。

相对批次 122（1209/895/34/279/1）：FUNCS +4，S +9，S-sig −5。
（5 个 [S-sig] 升格 [S]：startApplicationUsingCurrentPrivileges、startApplicationViaExplorer、
startApplicationWindowsAsAdmin、retryLaunchAsAdminIfElevationRequired、resolveApplicationWorkingDirectory；
4 个新 [S]：startApplication、startApplicationWindows、startApplicationWindowsStandard、
startApplicationWithShellParent。）

## G3 逻辑等价

`backend/applaunch_windows.go` + `backend/process_launch_windows.go` 全链切 `launchContext` 签名，
字段偏移均来自 asm（entryType@0x20、shellTarget@0xa0、entry@0xb0、args@0xc0）：

- **startApplication**（0x1408a6a20, 512B）[S]：resolveLaunchableAppEntry → 空→"入口路径不能为空"
  → directory→openPathDirectory → resolveEffectiveAppLaunchPrivilege(privilegeMode,"standard")
  → startApplicationWindows。
- **startApplicationWindows**（0x1408a9880, 448B）[S]：admin→AsAdmin；
  followLauncher→retry(UsingCurrentPrivileges,ctx)；standard→Standard。
- **startApplicationWindowsStandard**（0x1408a9a40, 320B）[S]：isProcessElevated 分流
  ViaExplorer/UsingCurrentPrivileges，均尾接 retryLaunchAsAdminIfElevationRequired。
- **startApplicationWindowsAsAdmin**（0x1408a9b80, 384B）[S]：已提权透传 UsingCurrentPrivileges；
  否则闭包 shellExecuteProgram("runas",…)。
- **startApplicationUsingCurrentPrivileges**（0x1408a7020, 864B）[S]：resolveLaunchableAppEntry →
  空→"入口路径不能为空" → directory→openPathDirectory → 提权分 exec.Command/withShellApartment；
  未提权→ViaExplorer。
- **startApplicationViaExplorer**（0x1408a9400, 576B）[S]：空→"应用路径不能为空"
  （@0x140c65190，与 WithShellParent 同源）；withExplorerShellDispatch(func1 0x1408a9780)
  →shellExecuteByExplorer(…,"",1)；失败后提权→WithShellParent / 未提权→withShellApartment(func2
  0x1408a9640)→shellExecuteProgram。
- **startApplicationWithShellParent**（0x1408aa0e0, 448B）[S]：空→"应用路径不能为空"；
  requiresShellOpen 真→52B 混淆错误（[P] @0x140c59916 待取证）；否则 buildCreateProcessCommandLine
  →createProcessWithShellParentWithVisibility(name,cmdline,workingDir,false)。
- **retryLaunchAsAdminIfElevationRequired**（0x1408acd60, 416B）[S]：isElevationRequiredError 假
  →透传 err；真→AsAdmin(ctx)。
- **resolveApplicationWorkingDirectory**（0x1408a74c0, 288B）[S]：directory→""；entry 非空→entry；
  shellTarget 是文件系统路径→inferWorkingDir；否则 ""。

## G4 复验

`TestStartApplicationUsingCurrentPrivilegesEmptyAppName`、`TestStartApplicationEmptyAppName`（新增）、
`TestRetryLaunchAsAdminIfElevationRequiredPassthrough`、`TestResolveApplicationWorkingDirectory`
（切 ctx 签名）全绿。740 判据测试不变。

## 判定

四路 PASS，批次 123 闭环。启动链 9 函数全部 100% asm 直译升格 launchContext 签名，
消除 5 处 [S-sig] 平铺占位。遗留：WithShellParent 52B 混淆错误文案 [P] 待取证。
