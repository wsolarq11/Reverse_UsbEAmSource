// startup_task_schtasks_windows.go — 开机自启任务域：schtasks 命令行创建/删除
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)
// 基线：docs/goresym/pipeline/tmp/{createLauncherStartupTaskUsingSchtasks,deleteLauncherStartupTask,
// createSchtasks_deferwrap1}.asm.txt
//
// 本域函数：
//   createLauncherStartupTaskUsingSchtasks (0x1408afa60) schtasks /Create 创建任务
//   deleteLauncherStartupTask            (0x1408affa0) schtasks /Delete 删除任务

package main

import (
	"fmt"
	"os"
	"os/exec"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// createLauncherStartupTaskUsingSchtasks 经 schtasks.exe 创建开机自启任务。
// [S 汇编 0x1408afa60, 1126B(0x466)]：
//
//	buildLauncherStartupTaskDefinitionXML(exe,args,userID,enabled,delaySeconds)（@0x1408afaa9）
//	  err→return err（@0x1408afab1/fe94）；
//	os.CreateTemp("", "usbeam-launcher-startup-task-*.xml")（@0x1408afad9，前缀 34B
//	  @0x140c759b3）err→return err（@0x1408afae3/fe78）；
//	tmpPath=file.Name()（@0x1408afaf1/f4/fb）；defer os.Remove(tmpPath)（deferwrap1 0x1408aff40，
//	  @0x1408afb0d/34）；
//	file.Write(xmlBytes) err→file.Close()+return err（@0x1408afb53/63/fe18）；
//	file.Close() err→return err（@0x1408afb74/83/fde3）；
//	convertXMLFileToUTF16LE(tmpPath)（@0x1408afb93）err→return err（@0x1408afba3/fdae）；
//	defer cleanup()（@0x1408afba9/b1）；
//	exec.Command("schtasks.exe","/Create","/TN","UsbEAm Launcher","/XML",utf16Path,"/F")
//	  （@0x1408afc88，命令名 12B @0x140c4ba20，任务名 15B @0x140c52cbf）；
//	CombinedOutput()（@0x1408afc8d）err→fmt.Errorf("创建开机启动任务失败: %w: %s",err,
//	  decodeWindowsCommandOutput(output))（@0x1408afca8/d22，格式 38B @0x140c7b2de）；
//	成功→nil。
func createLauncherStartupTaskUsingSchtasks(exe, args, userID string, enabled bool, delaySeconds int) error {
	xmlBytes, err := buildLauncherStartupTaskDefinitionXML(exe, args, userID, enabled, delaySeconds)
	if err != nil {
		return err
	}
	tmpFile, err := os.CreateTemp("", "usbeam-launcher-startup-task-*.xml")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)
	if _, err := tmpFile.Write(xmlBytes); err != nil {
		tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	utf16Path, cleanup, err := convertXMLFileToUTF16LE(tmpPath)
	if err != nil {
		return err
	}
	defer cleanup()
	cmd := exec.Command("schtasks.exe", "/Create", "/TN", "UsbEAm Launcher", "/XML", utf16Path, "/F")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("创建开机启动任务失败: %w: %s", err, decodeWindowsCommandOutput(output))
	}
	return nil
}

// deleteLauncherStartupTask 删除开机自启任务：优先 Task Scheduler COM API，失败回退 schtasks。
// [S 汇编 0x1408affa0, 512B(0x200)]：
//
//	withTaskSchedulerService(回调)（@0x1408affc2，闭包 0x141096a20 → func1 0x1409f7e80）
//	  成功→return nil（@0x1408affca/18a）；
//	exec.Command("schtasks.exe","/Delete","/TN","UsbEAm Launcher","/F")（@0x1408b006b）；
//	CombinedOutput()（@0x1408b0070）err nil→return nil（@0x1408b0078/17d）；
//	decodeWindowsCommandOutput（@0x1408b0088）→isTaskNotFoundMessage（@0x1408b0097）
//	  真→return nil（@0x1408b00a2/170）；
//	否则 fmt.Errorf("使用 Task Scheduler API 删除开机启动任务失败: %v；使用 schtasks 删除开机
//	  启动任务失败: %w: %s",apiErr,cmdErr,msg)（@0x1408b0162，格式 117B @0x140c950a7）。
//
// 回调 func1（0x1409f7e80，[S] 实证）：
//
//	callTaskSchedulerDispatchMethod(svc,"GetFolder",[]interface{}{`\`})（@0x1409f7ee2，9B
//	  @0x140c3f796）err→fmt.Errorf("获取任务计划根目录失败: %w")（@0x1409f7f20，37B
//	  @0x140c79dff）；defer folder.Release()（@0x1409f7f38）；
//	oleutil.CallMethod(folder,"DeleteTask","UsbEAm Launcher",0)（@0x1409f7f8e，10B
//	  @0x140c442e3）err→fmt.Errorf("删除开机启动任务失败: %w")（@0x1409f8006，34B
//	  @0x140c76167）；成功→nil。
func deleteLauncherStartupTask() error {
	apiErr := withTaskSchedulerService(func(svc *ole.IDispatch) error {
		folder, err := callTaskSchedulerDispatchMethod(svc, "GetFolder", []interface{}{`\`})
		if err != nil {
			return fmt.Errorf("获取任务计划根目录失败: %w", err)
		}
		defer folder.Release()

		if _, err := oleutil.CallMethod(folder, "DeleteTask", "UsbEAm Launcher", 0); err != nil {
			return fmt.Errorf("删除开机启动任务失败: %w", err)
		}
		return nil
	})
	if apiErr == nil {
		return nil
	}

	cmd := exec.Command("schtasks.exe", "/Delete", "/TN", "UsbEAm Launcher", "/F")
	output, cmdErr := cmd.CombinedOutput()
	if cmdErr == nil {
		return nil
	}
	msg := decodeWindowsCommandOutput(output)
	if isTaskNotFoundMessage(msg) {
		return nil
	}
	return fmt.Errorf("使用 Task Scheduler API 删除开机启动任务失败: %v；使用 schtasks 删除开机启动任务失败: %w: %s", apiErr, cmdErr, msg)
}
