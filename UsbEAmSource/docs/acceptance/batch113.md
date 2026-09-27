# 批次 113 验收（进程创建链：句柄提权判定 + 外壳进程打开）

日期：2026-09-23
子批次：process-launch-shellparent
目标：落地 `isProcessHandleElevated` [S] + `openShellProcessForUnelevatedLaunch` [S]
（createProcessWithShellParentWithVisibility 依赖链中间两环）。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 定向 test | `go test -tags production -count=1 -p=1 -run 'BuildCreateProcessCommandLine|IsElevationRequiredError|IsProcessHandleElevated|OpenShellProcess' ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1189 / MARKED=1189 / S=882 / S-inline=34 / S-sig=272 / P=1 / UNMARKED=0`。

相对批次 112（1187/880/34/272/1/0）：2 个新函数均 `[S]`。`UNMARKED=0` 保持。

## G3 逻辑等价

`backend/process_launch_windows.go`（2 函数）：

- **isProcessHandleElevated**（0x1408acb40, 288B）：OpenProcessToken(process, TOKEN_QUERY, &token)
  失败 → (false, wrapWinError("打开进程令牌失败",err))（24B @0x140c651c0）；成功 → defer
  token.Close() → getTokenElevation(token)。
- **openShellProcessForUnelevatedLaunch**（0x1408ac980, 448B）：GetShellWindow==0 → errors.New
  ("未找到系统外壳进程")（27B @0x140c69f93）；GetWindowThreadProcessId err → wrapWinError
  ("读取系统外壳进程失败")（30B @0x140c6f0db），pid==0 → 同上 errors.New；OpenProcess
  (PROCESS_QUERY_LIMITED_INFORMATION|PROCESS_CREATE_PROCESS=0x1080, false, pid) err →
  wrapWinError("打开系统外壳进程失败")（30B @0x140c6f0f9）；isProcessHandleElevated err →
  CloseHandle + wrapWinError；elevated → CloseHandle + errors.New("当前外壳进程仍为管理员权限，
  无法降权启动；请检查 UAC 是否已关闭")（92B @0x140c93c33）；否则 (process, nil)。

## G4 测试

`backend/process_launch_windows_test.go` 新增 2 用例：

- `TestIsProcessHandleElevatedCurrent`：当前进程伪句柄冒烟（无 explorer 环境则 skip）。
- `TestOpenShellProcessForUnelevatedLaunch`：外壳进程冒烟（无 shell 则 skip）。

全 PASS。

## 判定

四路 PASS，批次 113 闭环。`isProcessHandleElevated` 与 `openShellProcessForUnelevatedLaunch`
全 `[S]`；GetWindowThreadProcessId 用 x/sys v0.46.0 新签名 `(tid uint32, err error)`
（err 即 tid==0），OpenProcess 权限 0x1080 实测为 QUERY_LIMITED|CREATE_PROCESS。
createProcessWithShellParentWithVisibility 最后依赖 NewProcThreadAttributeList 链（批次 114+）。
