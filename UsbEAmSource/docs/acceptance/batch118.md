# 批次 118 验收（COM ShellDispatch 链）

日期：2026-09-23
子批次：com-explorer-dispatch
目标：落地 COM ShellDispatch 链 4 函数（[S-sig] 核心控制流实证，go-ole COM 依赖）。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1200 / MARKED=1200 / S=885 / S-inline=34 / S-sig=280 / P=1 / UNMARKED=0`。

相对批次 117（1196/885/34/276/1/0）：4 个新函数 [S-sig]。

## G3 逻辑等价

`backend/com_explorer_windows.go`（4 函数，核心 COM 链 asm 实证）：

- **getDesktopExplorerShellDispatch**（0x1408ab360, 2082B）：CreateObject("Shell.Application") →
  QueryInterface → Windows() → FindWindowSW(SWC_DESKTOP) → Item → Document → Application →
  返回 IDispatch。错误消息六段实测（初始化/集合/定位/视图/启动对象/调度接口）。
- **withDesktopExplorerShellDispatch**（0x1408ab2a0）：获取 → defer Release → fn(dispatch)。
- **withExplorerShellDispatch**（0x1408ab0e0）：LockOSThread + CoInitializeEx(0,6) +
  defer CoUninitialize/UnlockOSThread → 委托 withDesktopExplorerShellDispatch。
- **shellExecuteByExplorer**（0x1408abe60）：InvokeWithOptionalArgs("ShellExecute",
  DISPATCH_METHOD, [verb,appName,args,workingDir,show])；err →
  wrapWinError("委托 Explorer 启动应用失败")；result.Clear()。

方法名实测：Windows(7B)/FindWindowSW(12B)/Document(8B)/Application(11B)/ShellExecute(12B)。

## G4 测试

COM 函数依赖真实 Shell.Application 运行环境（离线无法执行 CreateObject），本批以编译 +
全量 test 门禁锁形状；错误消息字符串经 `read_gostring.py` 字节级解码对齐。

## 判定

四路 PASS，批次 118 闭环。COM ShellDispatch 链核心控制流 [S-sig] 实证。升格 [S] 的
剩余取证：FindWindowSW 的 VARIANT 手动布局（pvarLoc/pvarLocRoot/swClass/pHWND/
swfwOptions 五参精确字节），需 go-ole 实机逐 VARIANT 对齐。
