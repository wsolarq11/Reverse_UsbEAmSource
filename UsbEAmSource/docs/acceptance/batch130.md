# 批次 130 验收（开机自启域：顶层调度 + XML 模板取证）

日期：2026-09-23
子批次：startup-task-sync-template

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.396s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1235 / MARKED=1235 / S=925 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。

相对批次 129（1234/923/35/275/1）：FUNCS +1，S +2，S-sig -1。

新增：
- `syncLauncherStartupTask` 0x1408af840 [S]（`backend/startup_task_sync_windows.go`）

升级：
- `buildLauncherStartupTaskDefinitionXML` 0x1408b2080 [S-sig] → [S]（模板 duffcopy 取证闭合）

## G3 逻辑等价

`syncLauncherStartupTask`（[S]）：!enabled→deleteLauncherStartupTask；delaySeconds 夹取
[10,100]（<=0→10，>100→100）；os.Executable→filepath.Abs→currentLauncherStartupTaskUserID
→deleteLauncherStartupTask；args="--usbeam-startup-tray"（21B @0x140c5f818），enabled=false；
createLauncherStartupTaskUsingTaskSchedulerAPI 成功→nil；失败→
createLauncherStartupTaskUsingSchtasks 成功→nil；均失败→fmt.Errorf("使用 Task Scheduler API
创建开机启动任务失败: %v；%w", apiErr, schtasksErr)（65B @0x140c8f2d9）。

`buildLauncherStartupTaskDefinitionXML` 模板取证（duffcopy 源 @0x1411e4ea0 全字段解引用）：
- Version="1.2"（修正原 "1.0" 标准推断）；
- Settings.Enabled 恒 true（模板值，asm 未覆盖）、Settings.Hidden=enabled（asm @0x1408b2248
  写 def+0xda），修正原 "Enabled=enabled/Hidden=false" 反转；
- 其余字段与既有还原一致：Xmlns=2004/02/mit/task、LogonTrigger.Enabled=true、Principal.ID/
  Actions.Context="Author"、RunLevel="HighestAvailable"、MultipleInstancesPolicy="IgnoreNew"、
  IdleSettings(Duration="PT10M"/WaitTimeout="PT1H"/StopOnIdleEnd=true/RestartOnIdle=false)、
  AllowStartOnDemand=true、ExecutionTimeLimit="PT0S"、Priority=asm 覆盖 7。

## G4 复验

`startup_task_xml_windows_test.go` 新增 `TestBuildLauncherStartupTaskDefinitionXMLTemplateFields`
锁定 Version="1.2"、Enabled 恒 true、Hidden=enabled（enabled/disabled 两路断言）。

## 判定

四路 PASS，批次 130 闭环。

## 下一步（批次 131 方向）

- 开机自启域收尾：syncLauncherStartupTask 的上层调用点（热键/配置保存链路）核对签名与
  调用约定；quoteWindowsTaskActionCommand / launcherStartupTaskLogonDelay /
  launcherStartupTaskWorkingDirectory 若仍为 [P]/未标记需逐条实证。
- 其他域专项推进（按 HANDOFF 待移交清单）。
