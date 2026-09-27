// shellverb.go — Shell 动词 + 启动托盘进程域（逆向还原）
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)
// 基线：docs/goresym/source_funcs.txt + docs/goresym/pipeline/tmp/*.asm.txt
//
// 域内函数（批次 119 落地）：
//   launcherStartedForStartupTray     (0x1408af720) 参数含 --usbeam-startup-tray 判定
//   prepareLauncherStartupTrayProcess  (0x1408af7e0) 切到可执行文件目录
//   invokeShellVerb                    (0x1408a9de0) shell 动词分派（properties→属性窗口）
//   invokeShellVerbWithPowerShell      (0x1408ada40) PowerShell fallback（[S-sig]）

package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// launcherStartedForStartupTray 判定命令行参数是否含启动托盘标记。
// [S 汇编 0x1408af720, 192B]：遍历 args（rax/rbx），逐项 TrimSpace（@0x1408af76b）+
// EqualFold(arg, "--usbeam-startup-tray")（@0x1408af780，21B @0x140c5f818）；命中 → true，
// 否则 false。
func launcherStartedForStartupTray(args []string) bool {
	for _, arg := range args {
		if strings.EqualFold(strings.TrimSpace(arg), "--usbeam-startup-tray") {
			return true
		}
	}
	return false
}

// prepareLauncherStartupTrayProcess 切换到可执行文件所在目录。
// [S 汇编 0x1408af7e0, 96B]：os.Executable（@0x1408af7ee）err → 返回 err（@0x1408af7f6）；
// filepath.Abs(exe)（@0x1408af7f9）err → 返回 err（@0x1408af803）；filepath.Dir(abs)
// （@0x1408af811）→ os.Chdir(dir)（@0x1408af816）。
func prepareLauncherStartupTrayProcess() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	abs, err := filepath.Abs(exe)
	if err != nil {
		return err
	}
	return os.Chdir(filepath.Dir(abs))
}

// invokeShellVerb 分派 shell 动词：properties → 属性窗口，其余 PowerShell fallback。
// [S 汇编 0x1408a9de0, 208B]：TrimSpace(verb)/TrimSpace(path)（@0x1408a9e01/1a）；verb 或
// path 空 → errors.New("路径或动作不能为空")（@0x1408a9e2e，27B @0x140c69f42）；
// EqualFold(verb,"properties")（@0x1408a9e72，10B @0x140c441ad）→ showShellProperties(path)
// （@0x1408a9e85）；否则 invokeShellVerbWithPowerShell(verb,path)（@0x1408a9ea4）。
func invokeShellVerb(verb, path string) error {
	verb = strings.TrimSpace(verb)
	path = strings.TrimSpace(path)
	if verb == "" || path == "" {
		return errors.New("路径或动作不能为空")
	}
	if strings.EqualFold(verb, "properties") {
		return showShellProperties(path)
	}
	return invokeShellVerbWithPowerShell(verb, path)
}

// invokeShellVerbWithPowerShell 用 PowerShell 执行 shell 动词（fallback）。
// [S-sig 0x1408ada40, 350 行]：单引号转义实证 — strings.Replace(verb, "'", "”", -1)
// （@0x1408ada96, old="'" 1B/new="”" 2B）+ Replace(path)（@0x1408adb00）+
// concatstring3 拼接 PowerShell 模板（@0x1408adab8/22）。完整 PowerShell 脚本模板
// （duffcopy 源 @0x1411e46c0 起）待取证，此处以 Shell.Application ShellExecute 脚本等价承载。
func invokeShellVerbWithPowerShell(verb, path string) error {
	escVerb := strings.Replace(verb, "'", "''", -1)
	escPath := strings.Replace(path, "'", "''", -1)
	script := "& { $s = New-Object -ComObject Shell.Application; " +
		"$s.ShellExecute('" + escPath + "', '', '', '" + escVerb + "', 1) }"
	return startPowerShellEncodedCommand(script)
}

// InvokeShellVerb 执行文件壳层动词（BootstrapService 方法，直接转发 invokeShellVerb）。
// [S 0x1408a6920] 单 receiver + (verb,path string) 返回 error。asm：参数重排后尾调用 invokeShellVerb。
func (s *BootstrapService) InvokeShellVerb(verb, path string) error {
	return invokeShellVerb(verb, path)
}
