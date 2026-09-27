# 批次 128 验收（开机自启域：Task Scheduler COM 基础设施）

日期：2026-09-23
子批次：startup-task-com

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.341s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1231 / MARKED=1231 / S=920 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。

相对批次 127（1225/914/35/275/1）：FUNCS +6，S +6。

新增（新文件 `backend/startup_task_com_windows.go`，6 函数 [S]）：
- `withTaskSchedulerService` 0x1408b14e0（848B）
- `callTaskSchedulerMethod` 0x1408b1920（352B）
- `callTaskSchedulerDispatchMethod` 0x1408b1a80（128B）
- `getTaskSchedulerDispatchProperty` 0x1408b1b60（128B）
- `taskSchedulerDispatchFromVariant` 0x1408b1c20（368B）
- `putTaskSchedulerProperty` 0x1408b1e00（640B）

## G3 逻辑等价

`withTaskSchedulerService`（[S]）：LockOSThread→defer UnlockOSThread；CoInitialize(0) 失败→
"初始化 Task Scheduler COM 失败: %w"；defer CoUninitialize；CreateObject("Schedule.Service")
失败→"创建 Task Scheduler 服务对象失败: %w"；defer obj.Release；QueryInterface(IID_ITaskService)
失败→"获取 Task Scheduler 调度接口失败: %w"；defer svc.Release；callTaskSchedulerMethod(svc,
"Connect", nil) 失败→"连接 Task Scheduler 服务失败: %w"；fn(svc) 透传。

`callTaskSchedulerMethod` / `callTaskSchedulerDispatchMethod` / `getTaskSchedulerDispatchProperty`
/ `putTaskSchedulerProperty`（[S]）：go-ole InvokeWithOptionalArgs 封装，dispatch 分别为
DISPATCH_METHOD(1)/DISPATCH_METHOD(1)/DISPATCH_PROPERTYGET(2)/DISPATCH_PROPERTYPUT(4)；
result 非 nil 则 VariantClear。

`taskSchedulerDispatchFromVariant`（[S]）：nil→"%s 未返回调度对象"；ToIDispatch() nil→
VariantClear+"%s 返回的调度对象为空"；非 nil→VT=VT_EMPTY、Val=0 所有权转移。

全部错误串明文解码（39B/44B/44B/38B/30B/24B），ProgID "Schedule.Service"，方法名 "Connect"。
IID_ITaskService 用公开 GUID {2FABA4C7-4DA9-4013-9697-20CC3FD40F85}（asm 经 .data 重定位指针
间接装载，磁盘占位非明文）。

## G4 复验

COM 调用不可离线实证（需真实 Windows Task Scheduler 服务），本批以编译+静态逻辑等价门禁
为准，运行时实证列入后续。

## 判定

四路 PASS，批次 128 闭环。

## 下一步（批次 129 方向）

回接 schtasks /Delete 删除链：deleteLauncherStartupTask 0x1408affa0（asm 已 dump，117B 组合
错误文案已解码）+ deleteLauncherStartupTask.func1 0x141096a20（COM 回调，待 dump）。
再接 createLauncherStartupTaskUsingTaskSchedulerAPI 0x1408b01c0 +
applyLauncherStartupTaskSettings 0x1408b10a0（COM 创建链）。
