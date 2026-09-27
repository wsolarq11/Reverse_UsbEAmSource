# 批次 127 验收（开机自启域：schtasks 命令行创建任务）

日期：2026-09-23
子批次：startup-task-schtasks-create

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.382s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1225 / MARKED=1225 / S=914 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。

相对批次 126（1224/913/35/275/1）：FUNCS +1，S +1。

新增：`createLauncherStartupTaskUsingSchtasks` 0x1408afa60 [S]（新文件
`backend/startup_task_schtasks_windows.go`）。

## G3 逻辑等价

`createLauncherStartupTaskUsingSchtasks`（[S 汇编 0x1408afa60, 1126B]）：
- buildLauncherStartupTaskDefinitionXML(exe,args,userID,enabled,delaySeconds) err 直返；
- os.CreateTemp("", "usbeam-launcher-startup-task-*.xml")；defer os.Remove(tmpPath)
  （deferwrap1 0x1408aff40 已证 os.Remove）；
- Write err→Close+返 err；Close err→返 err；
- convertXMLFileToUTF16LE(tmpPath)；defer cleanup()；
- exec.Command("schtasks.exe","/Create","/TN","UsbEAm Launcher","/XML",utf16Path,"/F")
  CombinedOutput err→fmt.Errorf("创建开机启动任务失败: %w: %s",err,decodeWindowsCommandOutput(output))。

全部字符串明文解码：schtasks.exe / /Create / /TN / UsbEAm Launcher / /XML / /F /
usbeam-launcher-startup-task-*.xml / 创建开机启动任务失败: %w: %s。

## G4 复验

新增测试：空 exe / 空 userID 的错误传播（不触发文件系统与 schtasks 真实调用）。

## 判定

四路 PASS，批次 127 闭环。

## 下一步（批次 128 方向）

TaskScheduler COM 层专项：withTaskSchedulerService 0x1408b14e0（asm 已 dump，240 行，
go-ole coInitialize/CreateObject/QueryInterface/callTaskSchedulerMethod 链）→
callTaskSchedulerMethod 0x1408b1920 → callTaskSchedulerDispatchMethod 0x1408b1a80 →
putTaskSchedulerProperty 0x1408b1e00 → 之后回接 deleteLauncherStartupTask 0x1408affa0
（asm 已 dump，schtasks /Delete fallback + "使用 Task Scheduler API 删除...%v；使用
schtasks 删除...%w: %s" 117B 明文）与 createLauncherStartupTaskUsingTaskSchedulerAPI
0x1408b01c0 / applyLauncherStartupTaskSettings 0x1408b10a0。
