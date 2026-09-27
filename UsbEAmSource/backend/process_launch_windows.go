// process_launch_windows.go — 进程创建链（逆向还原）
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)
// 基线：docs/goresym/source_funcs.txt + docs/goresym/pipeline/tmp/*.asm.txt
//
// 域内函数（批次 112 落地）：
//   buildCreateProcessCommandLine (0x1408acf00)  exe+args → CreateProcess 命令行

package main

import (
	"errors"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

// buildCreateProcessCommandLine 组装 CreateProcess 命令行。
// [S 汇编 0x1408acf00, 288B] TrimSpace(exe)/TrimSpace(args)（@0x1408acf24/40）；
// exe 空 → 返回 args（@0x1408acf4d→0x1408acfdd）；否则
// Replace(exe,`"`,`\"`,-1)（@0x1408acf85，old=`"` 1B @0x1411cac88，new=`\"` 2B
// @0x140c336b5）→ concatstring3(nil,`"`,escaped,`"`)（@0x1408acfa4）；args 空 →
// 返回引号包裹的 exe；否则 concatstring3(nil,quoted,` `,args)（@0x1408acfd2，
// 空格 1B @0x1411ca5c8）。
func buildCreateProcessCommandLine(exe, args string) string {
	exe = strings.TrimSpace(exe)
	args = strings.TrimSpace(args)
	if exe == "" {
		return args
	}
	escaped := strings.Replace(exe, `"`, `\"`, -1)
	quoted := `"` + escaped + `"`
	if args == "" {
		return quoted
	}
	return quoted + " " + args
}

// isElevationRequiredError 判断错误是否为 ERROR_ELEVATION_REQUIRED(740)。
// [S 汇编 0x1408ace20, 224B] err nil → false（@0x1408ace43）；errors.Is(err,
// ERROR_ELEVATION_REQUIRED)（@0x1408ace60，target data 0x2e4=740 @0x1411cd5c0）真 →
// true；否则 errors.As(err,&errno)（@0x1408ace8e）且 errno==740（@0x1408ace9c）→
// true，否则 false。
func isElevationRequiredError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, windows.ERROR_ELEVATION_REQUIRED) {
		return true
	}
	var errno syscall.Errno
	return errors.As(err, &errno) && errno == windows.ERROR_ELEVATION_REQUIRED
}

// isProcessHandleElevated 判断进程句柄是否提权。
// [S 汇编 0x1408acb40, 288B] OpenProcessToken(process,0x8,&token)（@0x1408acb80，
// TOKEN_QUERY=8）失败 → (false, wrapWinError("打开进程令牌失败",err))（24B 消息
// @0x140c651c0）；成功 → defer token.Close()（func1 @0x1408acc60）→
// getTokenElevation(token)（@0x1408acbf1）。
func isProcessHandleElevated(process windows.Handle) (bool, error) {
	var token windows.Token
	if err := windows.OpenProcessToken(process, windows.TOKEN_QUERY, &token); err != nil {
		return false, wrapWinError("打开进程令牌失败", err)
	}
	defer token.Close()
	return getTokenElevation(token)
}

// openShellProcessForUnelevatedLaunch 打开未提权的系统外壳进程（explorer）。
// [S 汇编 0x1408ac980, 448B] GetShellWindow（@0x1408ac992）==0 → errors.New("未找到系统外壳进程")
// （27B @0x140c69f93）；GetWindowThreadProcessId（@0x1408ac9b6）err → wrapWinError("读取系统
// 外壳进程失败",err)（30B @0x140c6f0db），pid==0 → 同上 errors.New；OpenProcess(0x1080=
// PROCESS_QUERY_LIMITED_INFORMATION|PROCESS_CREATE_PROCESS,false,pid)（@0x1408ac9e0）err →
// wrapWinError("打开系统外壳进程失败",err)（30B @0x140c6f0f9）；isProcessHandleElevated err →
// CloseHandle + wrapWinError("读取系统外壳进程失败",err)；elevated 真 → CloseHandle +
// errors.New("当前外壳进程仍为管理员权限，无法降权启动；请检查 UAC 是否已关闭")（92B
// @0x140c93c33）；否则返回 (process, nil)。
func openShellProcessForUnelevatedLaunch() (windows.Handle, error) {
	shell := windows.GetShellWindow()
	if shell == 0 {
		return 0, errors.New("未找到系统外壳进程")
	}
	var pid uint32
	if _, err := windows.GetWindowThreadProcessId(shell, &pid); err != nil {
		return 0, wrapWinError("读取系统外壳进程失败", err)
	}
	if pid == 0 {
		return 0, errors.New("未找到系统外壳进程")
	}
	process, err := windows.OpenProcess(
		windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.PROCESS_CREATE_PROCESS,
		false, pid)
	if err != nil {
		return 0, wrapWinError("打开系统外壳进程失败", err)
	}
	elevated, err := isProcessHandleElevated(process)
	if err != nil {
		windows.CloseHandle(process)
		return 0, wrapWinError("读取系统外壳进程失败", err)
	}
	if elevated {
		windows.CloseHandle(process)
		return 0, errors.New("当前外壳进程仍为管理员权限，无法降权启动；请检查 UAC 是否已关闭")
	}
	return process, nil
}

// createProcessWithShellParentWithVisibility 以系统外壳为父进程启动目标（降权）。
// [S 汇编 0x1408ac140, 1828B] openShellProcessForUnelevatedLaunch（@0x1408ac1ac）err → 直接返回；
// defer CloseHandle(parent)（deferwrap2 @0x1408ac920）；NewProcThreadAttributeList(1)
// （@0x1408ac1f3）err → wrapWinError("初始化进程属性列表失败")（33B @0x140c73ba0），
// defer attrs.Delete()（deferwrap1 @0x1408ac8c0）；attrs.Update(PROC_THREAD_ATTRIBUTE_PARENT_PROCESS
// =0x20000,&parent,8)（@0x1408ac305）err → wrapWinError("设置父进程属性失败")（27B
// @0x140c69f78）；utf16.Encode(cmd/cmdLine)+NUL（@0x1408ac331/3c5），utf16PtrOrNil(dir)
// （@0x1408ac445）err → 返回 err，UTF16FromString("winsta0\\default")（@0x1408ac467）err →
// 返回 err；StartupInfoEx：cb=0x70、Desktop=&desktop[0]、ProcThreadAttributeList=attrs.List()；
// hidden 真 → STARTF_USESHOWWINDOW|SW_HIDE + CREATE_NO_WINDOW|EXTENDED（0x8080000），假 →
// EXTENDED|CREATE_NEW_CONSOLE（0x80010）；CreateProcess（@0x1408ac580）err →
// wrapWinError("启动进程失败")（18B @0x140c59e94）；成功 → CloseHandle(Thread/Process)。
func createProcessWithShellParentWithVisibility(cmd, cmdLine, dir string, hidden bool) error {
	parent, err := openShellProcessForUnelevatedLaunch()
	if err != nil {
		return err
	}
	defer windows.CloseHandle(parent)

	attrs, err := windows.NewProcThreadAttributeList(1)
	if err != nil {
		return wrapWinError("初始化进程属性列表失败", err)
	}
	defer attrs.Delete()

	if err := attrs.Update(windows.PROC_THREAD_ATTRIBUTE_PARENT_PROCESS, unsafe.Pointer(&parent), unsafe.Sizeof(parent)); err != nil {
		return wrapWinError("设置父进程属性失败", err)
	}

	appName := utf16.Encode([]rune(cmd))
	appName = append(appName, 0)
	commandLine := utf16.Encode([]rune(cmdLine))
	commandLine = append(commandLine, 0)

	workingDir, err := utf16PtrOrNil(dir)
	if err != nil {
		return err
	}
	desktop, err := windows.UTF16FromString("winsta0\\default")
	if err != nil {
		return err
	}

	si := &windows.StartupInfoEx{}
	si.Cb = uint32(unsafe.Sizeof(*si))
	si.Desktop = &desktop[0]
	si.ProcThreadAttributeList = attrs.List()

	var flags uint32
	if hidden {
		si.Flags = windows.STARTF_USESHOWWINDOW
		si.ShowWindow = windows.SW_HIDE
		flags = windows.CREATE_NO_WINDOW | windows.EXTENDED_STARTUPINFO_PRESENT
	} else {
		flags = windows.EXTENDED_STARTUPINFO_PRESENT | windows.CREATE_NEW_CONSOLE
	}

	var pi windows.ProcessInformation
	if err := windows.CreateProcess(&appName[0], &commandLine[0], nil, nil, false, flags, nil, workingDir, &si.StartupInfo, &pi); err != nil {
		return wrapWinError("启动进程失败", err)
	}
	windows.CloseHandle(pi.Thread)
	windows.CloseHandle(pi.Process)
	return nil
}

// buildShellExecuteArguments 组装 ShellExecute 参数串（逐参数 Trim + Windows 转义 + 空格连接）。
// [S 汇编 0x1408ad020, 608B] len==0 → ""（@0x1408ad045）；make([]string,0,len)（@0x1408ad080，
// len>2 走 makeslice，否则栈缓冲）；逐 arg TrimSpace（@0x1408ad114），空跳过，非空
// quoteWindowsArgument（@0x1408ad137）→ append（@0x1408ad1e2）；strings.Join(quoted," ")
// （@0x1408ad22a，分隔符 " " 1B @0x1411ca5c8）。
func buildShellExecuteArguments(args []string) string {
	if len(args) == 0 {
		return ""
	}
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		if arg = strings.TrimSpace(arg); arg != "" {
			quoted = append(quoted, quoteWindowsArgument(arg))
		}
	}
	return strings.Join(quoted, " ")
}

// resolveApplicationWorkingDirectory 解析应用工作目录。
// [S 汇编 0x1408a74c0, 288B(0x120)]：normalizeAppEntryType(ctx.entryType)=="directory"
// （@0x1408a7520，9B "directory"）→ ""（@0x1408a7542）；TrimSpace(ctx.entry) 非空
// （@0x1408a7560）→ 原样返回；否则 TrimSpace(ctx.shellTarget) 且 looksLikeFilesystemPath
// （@0x1408a7589）→ inferWorkingDir（@0x1408a75a0）；否则 ""。
// 字段偏移实证：entryType@0x20、entry@0xb0、shellTarget@0xa0。
func resolveApplicationWorkingDirectory(ctx launchContext) string {
	if normalizeAppEntryType(ctx.entryType) == "directory" {
		return ""
	}
	if entry := strings.TrimSpace(ctx.entry); entry != "" {
		return entry
	}
	if target := strings.TrimSpace(ctx.shellTarget); looksLikeFilesystemPath(target) {
		return inferWorkingDir(target)
	}
	return ""
}
