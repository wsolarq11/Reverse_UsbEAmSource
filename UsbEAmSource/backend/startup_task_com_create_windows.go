// startup_task_com_create_windows.go — 开机自启任务域：Task Scheduler COM 创建链
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)
// 基线：docs/goresym/pipeline/tmp/{createLauncherStartupTaskUsingTaskSchedulerAPI,
// applyLauncherStartupTaskSettings}.asm.txt
//
// 本域函数：
//   createLauncherStartupTaskUsingTaskSchedulerAPI (0x1408b01c0) COM 创建任务入口
//   applyLauncherStartupTaskSettings               (0x1408b10a0) Settings/IdleSettings 批量属性

package main

import (
	"errors"
	"fmt"
	"strings"

	ole "github.com/go-ole/go-ole"
)

// createLauncherStartupTaskUsingTaskSchedulerAPI 经 Task Scheduler COM API 创建开机自启任务。
// [S 汇编 0x1408b01c0, 176B(0xb0) + func1 0x1408b0420, 2620B(0xa3c)]：
//
//	exe=TrimSpace；空→errors.New("开机启动程序路径不能为空")（36B @0x140c784e3）；
//	userID=TrimSpace；空→errors.New("开机启动任务用户标识不能为空")（42B @0x140c80e94）；
//	workDir=launcherStartupTaskWorkingDirectory(exe)（@0x1408b0292）；
//	logonDelay=launcherStartupTaskLogonDelay(delaySeconds)（@0x1408b02a9）；
//	logonType=enabled?2:3（@0x1408b02fb/30b，S4U/InteractiveToken）；
//	withTaskSchedulerService(func1)。
//
// func1（回调，捕获 exe/args/userID/logonType/logonDelay/workDir/enabled）：
//
//	GetFolder(`\`)（@0x1408b0504）→"获取任务计划根目录失败: %w"；defer Release；
//	NewTask(0)（@0x1408b05e0）→"创建任务定义失败: %w"；defer Release；
//	Principal（@0x1408b0689）→"获取任务主体失败: %w"；defer Release；
//	putTaskSchedulerProperty(principal,"UserId",userID)/("RunLevel",1)/("LogonType",logonType)；
//	Triggers（@0x1408b07ea）→"获取任务触发器集合失败: %w"；defer Release；
//	Create(9)（@0x1408b08c1，TASK_TRIGGER_LOGON）→"创建登录触发器失败: %w"；defer Release；
//	putTaskSchedulerProperty(trigger,"Enabled",true)/("Delay",logonDelay)；
//	Actions（@0x1408b09f0）→"获取任务动作集合失败: %w"；defer Release；
//	Create(0)（@0x1408b0ac7，TASK_ACTION_EXEC）→"创建执行动作失败: %w"；defer Release；
//	putTaskSchedulerProperty(action,"Path",exe)；args 非空才 ("Arguments",TrimSpace(args))；
//	("WorkingDirectory",workDir)；
//	Settings（@0x1408b0c82）→"获取任务设置失败: %w"；defer Release；
//	applyLauncherStartupTaskSettings(settings,enabled)（@0x1408b0d25）；
//	RegisterTaskDefinition("UsbEAm Launcher",task,6,userID,"",logonType)（@0x1408b0e23，
//	  方法 22B @0x140c612d6，flags=6 TASK_CREATE_OR_UPDATE）→"注册开机启动任务失败: %w"。
func createLauncherStartupTaskUsingTaskSchedulerAPI(exe, args, userID string, enabled bool, delaySeconds int) error {
	exe = strings.TrimSpace(exe)
	if exe == "" {
		return errors.New("开机启动程序路径不能为空")
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return errors.New("开机启动任务用户标识不能为空")
	}
	workDir := launcherStartupTaskWorkingDirectory(exe)
	logonDelay := launcherStartupTaskLogonDelay(delaySeconds)
	logonType := 3 // TASK_LOGON_INTERACTIVE_TOKEN
	if enabled {
		logonType = 2 // TASK_LOGON_S4U
	}

	return withTaskSchedulerService(func(svc *ole.IDispatch) error {
		folder, err := callTaskSchedulerDispatchMethod(svc, "GetFolder", []interface{}{`\`})
		if err != nil {
			return fmt.Errorf("获取任务计划根目录失败: %w", err)
		}
		defer folder.Release()

		task, err := callTaskSchedulerDispatchMethod(svc, "NewTask", []interface{}{0})
		if err != nil {
			return fmt.Errorf("创建任务定义失败: %w", err)
		}
		defer task.Release()

		principal, err := getTaskSchedulerDispatchProperty(task, "Principal")
		if err != nil {
			return fmt.Errorf("获取任务主体失败: %w", err)
		}
		defer principal.Release()
		if err := putTaskSchedulerProperty(principal, "UserId", userID); err != nil {
			return err
		}
		if err := putTaskSchedulerProperty(principal, "RunLevel", 1); err != nil {
			return err
		}
		if err := putTaskSchedulerProperty(principal, "LogonType", logonType); err != nil {
			return err
		}

		triggers, err := getTaskSchedulerDispatchProperty(task, "Triggers")
		if err != nil {
			return fmt.Errorf("获取任务触发器集合失败: %w", err)
		}
		defer triggers.Release()
		trigger, err := callTaskSchedulerDispatchMethod(triggers, "Create", []interface{}{9})
		if err != nil {
			return fmt.Errorf("创建登录触发器失败: %w", err)
		}
		defer trigger.Release()
		if err := putTaskSchedulerProperty(trigger, "Enabled", true); err != nil {
			return err
		}
		if err := putTaskSchedulerProperty(trigger, "Delay", logonDelay); err != nil {
			return err
		}

		actions, err := getTaskSchedulerDispatchProperty(task, "Actions")
		if err != nil {
			return fmt.Errorf("获取任务动作集合失败: %w", err)
		}
		defer actions.Release()
		action, err := callTaskSchedulerDispatchMethod(actions, "Create", []interface{}{0})
		if err != nil {
			return fmt.Errorf("创建执行动作失败: %w", err)
		}
		defer action.Release()
		if err := putTaskSchedulerProperty(action, "Path", exe); err != nil {
			return err
		}
		if trimmedArgs := strings.TrimSpace(args); trimmedArgs != "" {
			if err := putTaskSchedulerProperty(action, "Arguments", trimmedArgs); err != nil {
				return err
			}
		}
		if err := putTaskSchedulerProperty(action, "WorkingDirectory", workDir); err != nil {
			return err
		}

		settings, err := getTaskSchedulerDispatchProperty(task, "Settings")
		if err != nil {
			return fmt.Errorf("获取任务设置失败: %w", err)
		}
		defer settings.Release()
		if err := applyLauncherStartupTaskSettings(settings, enabled); err != nil {
			return err
		}

		if err := callTaskSchedulerMethod(folder, "RegisterTaskDefinition", []interface{}{
			"UsbEAm Launcher", task, 6, userID, "", logonType,
		}); err != nil {
			return fmt.Errorf("注册开机启动任务失败: %w", err)
		}
		return nil
	})
}

// taskSchedulerProperty 描述一次 putTaskSchedulerProperty 调用的属性名与值。
type taskSchedulerProperty struct {
	name  string
	value interface{}
}

// applyLauncherStartupTaskSettings 批量写入任务 Settings 与 IdleSettings 属性。
// [S 汇编 0x1408b10a0, 1048B(0x418)]：
//
//	13 项 Settings 属性表（基础表 @0x1411e63a8，值由栈上常量覆盖；第 8 项 Hidden
//	  由 enabled 在 {0x1411f3720,0x1411f3728} 常量表中二选一）依次 putTaskSchedulerProperty；
//	getTaskSchedulerDispatchProperty(settings,"IdleSettings")（@0x1408b12a9，12B
//	  @0x140c4ba2c）→"获取任务空闲设置失败: %w"（@0x1408b12ee，34B @0x140c75a19）；
//	4 项 IdleSettings 属性依次 putTaskSchedulerProperty（@0x1408b1438）。
func applyLauncherStartupTaskSettings(settings *ole.IDispatch, enabled bool) error {
	props := []taskSchedulerProperty{
		{"MultipleInstances", 2},
		{"DisallowStartIfOnBatteries", false},
		{"StopIfGoingOnBatteries", false},
		{"AllowHardTerminate", false},
		{"StartWhenAvailable", false},
		{"RunOnlyIfNetworkAvailable", false},
		{"AllowDemandStart", true},
		{"Enabled", true},
		{"Hidden", enabled},
		{"RunOnlyIfIdle", false},
		{"WakeToRun", false},
		{"ExecutionTimeLimit", "PT0S"},
		{"Priority", 7},
	}
	for _, p := range props {
		if err := putTaskSchedulerProperty(settings, p.name, p.value); err != nil {
			return err
		}
	}

	idle, err := getTaskSchedulerDispatchProperty(settings, "IdleSettings")
	if err != nil {
		return fmt.Errorf("获取任务空闲设置失败: %w", err)
	}
	defer idle.Release()

	idleProps := []taskSchedulerProperty{
		{"IdleDuration", "PT10M"},
		{"WaitTimeout", "PT1H"},
		{"StopOnIdleEnd", true},
		{"RestartOnIdle", false},
	}
	for _, p := range idleProps {
		if err := putTaskSchedulerProperty(idle, p.name, p.value); err != nil {
			return err
		}
	}
	return nil
}
