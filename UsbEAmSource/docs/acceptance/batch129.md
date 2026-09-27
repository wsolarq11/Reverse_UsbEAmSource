# 批次 129 验收（开机自启域：Task Scheduler COM 创建/删除链）

日期：2026-09-23
子批次：startup-task-com-create-delete

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.396s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1234 / MARKED=1234 / S=923 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。

相对批次 128（1231/920/35/275/1）：FUNCS +3，S +3。

新增：
- `deleteLauncherStartupTask` 0x1408affa0 [S]（`backend/startup_task_schtasks_windows.go`）
- `createLauncherStartupTaskUsingTaskSchedulerAPI` 0x1408b01c0 [S]（`backend/startup_task_com_create_windows.go`）
- `applyLauncherStartupTaskSettings` 0x1408b10a0 [S]（`backend/startup_task_com_create_windows.go`）

## G3 逻辑等价

`deleteLauncherStartupTask`（[S]）：withTaskSchedulerService(回调) 成功→nil；否则 schtasks
/Delete /TN "UsbEAm Launcher" /F fallback；CombinedOutput 成功→nil；isTaskNotFoundMessage→nil；
否则 fmt.Errorf("使用 Task Scheduler API 删除开机启动任务失败: %v；使用 schtasks 删除开机启动
任务失败: %w: %s", apiErr, cmdErr, msg)（117B @0x140c950a7）。回调 func1 0x1409f7e80 [S]：
GetFolder(`\`)→"获取任务计划根目录失败: %w"；oleutil.CallMethod(folder,"DeleteTask",
"UsbEAm Launcher",0)→"删除开机启动任务失败: %w"。

`createLauncherStartupTaskUsingTaskSchedulerAPI`（[S]）：exe/userID TrimSpace 空校验（36B/42B
errors.New）；workDir=launcherStartupTaskWorkingDirectory(exe)；logonDelay=
launcherStartupTaskLogonDelay(delaySeconds)；logonType=enabled?2:3；withTaskSchedulerService
(func1)。func1 0x1408b0420 [S] 完整 COM 链：GetFolder(`\`)→NewTask(0)→Principal(
UserId/RunLevel=1/LogonType)→Triggers.Create(9 TASK_TRIGGER_LOGON, Enabled=true, Delay=
logonDelay)→Actions.Create(0 TASK_ACTION_EXEC, Path=exe, Arguments=TrimSpace(args) 非空才设,
WorkingDirectory=workDir)→Settings(applyLauncherStartupTaskSettings)→
RegisterTaskDefinition("UsbEAm Launcher", task, 6 TASK_CREATE_OR_UPDATE, userID, "",
logonType)。10 条错误串全明文解码（获取任务计划根目录/创建任务定义/获取任务主体/获取任务
触发器集合/创建登录触发器/获取任务动作集合/创建执行动作/获取任务设置/获取任务空闲设置/
注册开机启动任务 失败: %w）。

`applyLauncherStartupTaskSettings`（[S]）：13 项 Settings 属性表（MultipleInstances=2、
DisallowStartIfOnBatteries/StopIfGoingOnBatteries/AllowHardTerminate/StartWhenAvailable/
RunOnlyIfNetworkAvailable=false、AllowDemandStart/Enabled=true、Hidden=enabled、
RunOnlyIfIdle/WakeToRun=false、ExecutionTimeLimit="PT0S"、Priority=7）→
IdleSettings 4 项（IdleDuration="PT10M"、WaitTimeout="PT1H"、StopOnIdleEnd=true、
RestartOnIdle=false）。属性表基础段 @0x1411e63a8（13 个属性名）+ 栈上常量覆盖；
Hidden 由 enabled 在 {0x1411f3720=0, 0x1411f3728=1} 常量表二选一。

## G4 复验

COM 调用不可离线实证（需真实 Windows Task Scheduler 服务），本批以编译+静态逻辑等价门禁
为准；空 exe/空 userID 错误路径已加单测覆盖（`startup_task_com_create_windows_test.go`）。

## 判定

四路 PASS，批次 129 闭环。

## 下一步（批次 130 方向）

- `syncLauncherStartupTask` 0x1408af840（asm 已 dump，顶层调度：com 路径 vs schtasks 路径选择）。
- `buildLauncherStartupTaskDefinitionXML` 模板取证：duffcopy 源 @0x1411e4ea8 各字段精确
  解引用（当前为 Task Scheduler 标准语义推断 [P]）。
