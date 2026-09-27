// startup_task_sync_windows.go — 开机自启任务域：顶层同步调度
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)
// 基线：docs/goresym/pipeline/tmp/syncLauncherStartupTask.asm.txt
//
// 本域函数（批次 130 落地）：
//   syncLauncherStartupTask (0x1408af840) 删除/创建调度入口
//
// currentLauncherStartupTaskUserID 已在 startup_task_xml_windows.go [S] 落地；
// launcherStartedForStartupTray / prepareLauncherStartupTrayProcess 已在 shellverb.go [S] 落地。

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// syncLauncherStartupTask 按 enabled 同步开机自启任务：禁用删除，启用（重新）创建。
// [S 汇编 0x1408af840, 464B(0x1d0)]：
//
//	!enabled → deleteLauncherStartupTask()（@0x1408af860/20）；
//	delaySeconds 夹取到 [10,100]（<=0→10 @0x1408af86d，>100→100 @0x1408af87a）；
//	os.Executable()（@0x1408af884）err→return err；
//	filepath.Abs(exe)（@0x1408af893）err→return err；
//	currentLauncherStartupTaskUserID()（@0x1408af8b3）err→return err；
//	deleteLauncherStartupTask()（@0x1408af8d3）err→return err；
//	args="--usbeam-startup-tray"（21B @0x140c5f818），enabled=false；
//	createLauncherStartupTaskUsingTaskSchedulerAPI(exeAbs,args,userID,false,delaySeconds)
//	  （@0x1408af911）成功→nil；
//	失败→createLauncherStartupTaskUsingSchtasks(exeAbs,args,userID,false,delaySeconds)
//	  （@0x1408af951）成功→nil；
//	均失败→fmt.Errorf("使用 Task Scheduler API 创建开机启动任务失败: %v；%w",
//	  apiErr,schtasksErr)（@0x1408af9c1，格式 65B @0x140c8f2d9）。
func syncLauncherStartupTask(enabled bool, delaySeconds int) error {
	if !enabled {
		return deleteLauncherStartupTask()
	}
	if delaySeconds <= 0 {
		delaySeconds = 10
	} else if delaySeconds > 100 {
		delaySeconds = 100
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exeAbs, err := filepath.Abs(exe)
	if err != nil {
		return err
	}
	userID, err := currentLauncherStartupTaskUserID()
	if err != nil {
		return err
	}
	if err := deleteLauncherStartupTask(); err != nil {
		return err
	}

	args := "--usbeam-startup-tray"
	apiErr := createLauncherStartupTaskUsingTaskSchedulerAPI(exeAbs, args, userID, false, delaySeconds)
	if apiErr == nil {
		return nil
	}
	if err := createLauncherStartupTaskUsingSchtasks(exeAbs, args, userID, false, delaySeconds); err == nil {
		return nil
	} else {
		return fmt.Errorf("使用 Task Scheduler API 创建开机启动任务失败: %v；%w", apiErr, err)
	}
}
