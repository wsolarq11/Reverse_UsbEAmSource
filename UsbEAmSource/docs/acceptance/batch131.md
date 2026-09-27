# 批次 131 验收（开机自启域：配置提交链枢纽 syncStartupTask）

日期：2026-09-23
子批次：startup-task-sync-method

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.409s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1236 / MARKED=1236 / S=926 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。

相对批次 130（1235/925/35/274/1）：FUNCS +1，S +1。

新增：
- `(*BootstrapService).syncStartupTask` 0x1407a14c0 [S]（`backend/bootstrapservice.go`）

## G3 逻辑等价

`syncStartupTask(enabled bool, delaySeconds int) error`（[S]，240B/0xf0）：
- bs==nil → `errors.New("启动任务服务不可用")`（27B @0x140c69c33）；
- `lock([bs+0x540])`（cmpxchg @0x1407a14f6）→ defer unlock（xadd @0x1407a153c）；
- 读 `[bs+0x510]=startupTaskSync`（func value，1 字）；
- nil → 默认 funcval @0x141096d88（`[0]=syncLauncherStartupTask 0x1408af840`，
  `[8]=0x1409f7080`，`[0x10]=0x1409f70a0`，`[0x18]=unregisterLauncherHotkey`）；
- call fn(enabled, delaySeconds)（@0x1407a1580，eax=enabled，rbx=delaySeconds，rdx=funcval）。

**关键取证（本轮新增）**：
- Go 1.25 `func(bool,int) error` 的 `unsafe.Sizeof == 8`（1 字），非历史认知的 2 字；
  故结构体字段 `startupTaskSync` 单字指针语义与 asm `[bs+0x510]` 一致。
- `unsafe.Offsetof` 实测：`startupTaskSync=0x510`、`workspaceTransaction=0x518`、
  `workspaceDataMaintenance=0x520`、`lock=0x540`，与 asm 完全吻合（结构体布局无漂移）。
- xref 扫描（.text 段 `E8 rel32`）：syncStartupTask 共 8 处调用——
  SaveConfig 主函数 1 处、SaveConfig.func1.1 1 处、ResetConfig.func1/.1 2 处、
  commitLauncherConfigReplacementWithWidgets 主函数 4 处。均属配置提交/重置链。

## G4 复验

`bootstrapservice_test.go` 新增：
- `TestSyncStartupTaskNilReceiver`（nil → "启动任务服务不可用"）；
- `TestSyncStartupTaskCustomFn`（注入 startupTaskSync 验证参数透传 enabled=true/delay=42）；
- `TestBootstrapServiceLayout` 增补 startupTaskSync offset=0x510 断言。

## 判定

四路 PASS，批次 131 闭环。

## 下一步（批次 132 方向）

- 配置提交链接入：SaveConfig / ResetConfig / commitLauncherConfigReplacementWithWidgets
  三函数体需接入 syncStartupTask 调用（8 处调用点），逐点核对 enabled/delaySeconds 实参
  来源（配置字段），补齐当前函数体缺失的调用。
