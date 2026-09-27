# 批次 132 验收（开机自启域：配置比较同步 helper + SaveConfig prepare 取证）

日期：2026-09-23
子批次：startup-task-config-compare

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.188s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1237 / MARKED=1237 / S=927 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。

相对批次 131（1236/926/35/274/1）：FUNCS +1，S +1。

新增：
- `(*BootstrapService).syncStartupTaskFromConfig` [S]（`backend/bootstrapservice.go`）

## G3 逻辑等价

`syncStartupTaskFromConfig(newCfg, oldCfg LauncherConfig) error`（[S]，提取自 SaveConfig.func1
内联段 0x140777618-0x140777725）：
- 读 `Preferences.StartupLaunchEnabled *bool`（nil→false）与 `StartupLaunchDelaySeconds int`；
- delay clamp [10,100]（<=0→10，>100→100，@0x140777634/639/642/648）；
- 触发条件 `newEnabled != oldEnabled || (newEnabled && newDelay != oldDelay)`
  （@0x1407776cc/e4/e7）→ `bs.syncStartupTask(newEnabled, newDelay)`（@0x140777720）。

**配置字段来源取证（本轮新增）**：func1 prepare 回调读取的 `*bool @+0xb0` / `int @+0xb8`
对应 `types_config.go` 的 `Preferences.StartupLaunchEnabled` / `StartupLaunchDelaySeconds`。

**SaveConfig.func1 prepare 回调全景（取证，待批次 133+ 重建）**：
1. 入参 al==0 或 `isLauncherConfigEmptyInitialConfig(newCfg)` → `errors.New("配置尚未初始化，
   请先完成初始化向导" 51B @0x140c88d4f)`；
2. 比较新旧开机自启设置 → syncStartupTaskFromConfig（本批落地）；
3. 两轮 18 键特性模块 map 构建（mapassign_faststr，@0x140777782-0x140777ba7）；
4. desktopWidgets(+0x440) 非 nil 且 FeatureModules["desktopWidgets"] 变化 →
   `pauseRunningTasksWithRollback`（@0x140777c2f），返回 []func() 回滚句柄。

## G4 复验

`bootstrapservice_test.go` 新增 `TestSyncStartupTaskFromConfig`（8 用例：启用变化/延时变化/
无变化/禁用延时不触发/nil 视为 false/延时 clamp 0→10、150→100）。

## 判定

四路 PASS，批次 132 闭环。

## 下一步（批次 133 方向）

- CompareAndSwapPrepared 真实签名重建：asm @0x140777082 显示 6 参形态（store、cfg、
  prepare func()、options、两枚 int），现 [S-sig] 简化为单参；需还原 prepare 回调参数。
- SaveConfig / ResetConfig 完整重建：接入 func1/func1.1 prepare 回调（含 syncStartupTaskFromConfig）。
