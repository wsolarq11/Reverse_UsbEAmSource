// applaunch_windows.go — 应用启动链（逆向还原，批次 123 升格 launchContext 签名）
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)
// 基线：docs/goresym/pipeline/tmp/startApplication*.asm.txt
//
// 本批将整个启动链从平铺参数升格为 launchContext 结构体签名，字段偏移均来自 asm：
//   entryType@0x20 / shellTarget@0xa0 / entry@0xb0 / args@0xc0。
//
//   startApplication                          (0x1408a6a20) 顶层：resolveLaunchableAppEntry → 分流
//   startApplicationWindows                   (0x1408a9880) 权限分派（admin/standard/followLauncher）
//   startApplicationWindowsStandard           (0x1408a9a40) standard：explorer/current + 740 重试
//   startApplicationWindowsAsAdmin            (0x1408a9b80) 提权(UAC runas)启动
//   startApplicationUsingCurrentPrivileges    (0x1408a7020) 当前权限启动主入口
//   startApplicationViaExplorer               (0x1408a9400) COM explorer 降权启动（回退壳）
//   startApplicationWithShellParent           (0x1408aa0e0) CreateProcess + shell parent 启动
//   retryLaunchAsAdminIfElevationRequired     (0x1408acd60) 740 错误重试提权

package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-ole/go-ole"
)

// startApplication 启动应用（顶层入口，权限模式优先）。
// [S 汇编 0x1408a6a20, 512B(0x200)]：resolveLaunchableAppEntry（@0x1408a6a79）；
// TrimSpace(shellTarget) 空 → "入口路径不能为空"（@0x1408a6bc1，24B @0x140c65178）；
// normalizeAppEntryType(entryType)=="directory"（@0x1408a6b24）→ openPathDirectory
// （@0x1408a6b54）；否则 resolveEffectiveAppLaunchPrivilege(privilegeMode,"standard")
// （@0x1408a6b80）→ startApplicationWindows（@0x1408a6bb3）。
func startApplication(privilegeMode string, ctx launchContext) error {
	ctx = resolveLaunchableAppEntry(ctx)
	name := strings.TrimSpace(ctx.shellTarget)
	if name == "" {
		return errors.New("入口路径不能为空")
	}
	if normalizeAppEntryType(ctx.entryType) == "directory" {
		return openPathDirectory(name)
	}
	privilege := resolveEffectiveAppLaunchPrivilege(privilegeMode, "standard")
	return startApplicationWindows(privilege, ctx)
}

// startApplicationWindows 按生效权限模式分派启动。
// [S 汇编 0x1408a9880, 448B(0x1c0)]：resolveEffectiveAppLaunchPrivilege(privilegeMode,"standard")
// （@0x1408a98b1）→ "admin"（@0x1408a98b6）→ startApplicationWindowsAsAdmin
// （@0x1408a9900）；"followLauncher"（@0x1408a990e）→ retryLaunchAsAdminIfElevationRequired(
// startApplicationUsingCurrentPrivileges,ctx)（@0x1408a9973/99b3）；否则
// startApplicationWindowsStandard（@0x1408a99f3）。
func startApplicationWindows(privilegeMode string, ctx launchContext) error {
	switch resolveEffectiveAppLaunchPrivilege(privilegeMode, "standard") {
	case "admin":
		return startApplicationWindowsAsAdmin(ctx)
	case "followLauncher":
		return retryLaunchAsAdminIfElevationRequired(startApplicationUsingCurrentPrivileges(ctx), ctx)
	default: // "standard"
		return startApplicationWindowsStandard(ctx)
	}
}

// startApplicationWindowsStandard standard 模式：explorer/current 启动后 740 重试。
// [S 汇编 0x1408a9a40, 320B(0x140)]：isProcessElevated（@0x1408a9a60）真 →
// retryLaunchAsAdminIfElevationRequired(startApplicationViaExplorer,ctx)
// （@0x1408a9a95/ad3）；否则 retryLaunchAsAdminIfElevationRequired(
// startApplicationUsingCurrentPrivileges,ctx)（@0x1408a9b13/b53）。
func startApplicationWindowsStandard(ctx launchContext) error {
	if isProcessElevated() {
		return retryLaunchAsAdminIfElevationRequired(startApplicationViaExplorer(ctx), ctx)
	}
	return retryLaunchAsAdminIfElevationRequired(startApplicationUsingCurrentPrivileges(ctx), ctx)
}

// startApplicationWindowsAsAdmin 提权(UAC runas)启动应用。
// [S 汇编 0x1408a9b80, 384B(0x180)]：isProcessElevated（@0x1408a9ba0）真 →
// startApplicationUsingCurrentPrivileges（@0x1408a9bd5）；否则闭包 func1（0x1408a9d00）：
// TrimSpace(shellTarget)（@0x1408a9bf3）、buildShellExecuteArguments(args)（@0x1408a9c20）、
// requiresShellOpen 真 → 清空 shellArgs（@0x1408a9c97 cmovne）、resolveApplicationWorkingDirectory
// （@0x1408a9d74）→ shellExecuteProgram("runas",…)（@0x1408a9dab）。
func startApplicationWindowsAsAdmin(ctx launchContext) error {
	if isProcessElevated() {
		return startApplicationUsingCurrentPrivileges(ctx)
	}
	name := strings.TrimSpace(ctx.shellTarget)
	shellArgs := buildShellExecuteArguments(ctx.args)
	if requiresShellOpen(name) {
		shellArgs = ""
	}
	return withShellApartment(func() error {
		workingDir := resolveApplicationWorkingDirectory(ctx)
		return shellExecuteProgram("runas", name, shellArgs, workingDir)
	})
}

// startApplicationUsingCurrentPrivileges 以当前权限启动应用（主入口）。
// [S 汇编 0x1408a7020, 864B(0x360)]：resolveLaunchableAppEntry（@0x1408a7073）；
// TrimSpace(shellTarget) 空 → "入口路径不能为空"（@0x1408a7328，24B @0x140c65178）；
// normalizeAppEntryType(entryType)=="directory" → openPathDirectory（@0x1408a7154）；
// isProcessElevated 真且 requiresShellOpen 假 → exec.Command(name,args...)+Dir+Start
// （@0x1408a71b5/71f3/7224），err → fmt.Errorf("启动失败: %w")（@0x1408a7269）；
// isProcessElevated 真且 requiresShellOpen 真 → withShellApartment(shellExecuteProgram("",…))
// （func1 @0x1408a7380）；isProcessElevated 假 → startApplicationViaExplorer
// （@0x1408a731a）。
func startApplicationUsingCurrentPrivileges(ctx launchContext) error {
	ctx = resolveLaunchableAppEntry(ctx)
	name := strings.TrimSpace(ctx.shellTarget)
	if name == "" {
		return errors.New("入口路径不能为空")
	}
	if normalizeAppEntryType(ctx.entryType) == "directory" {
		return openPathDirectory(name)
	}
	if isProcessElevated() {
		if !requiresShellOpen(name) {
			cmd := exec.Command(name, ctx.args...)
			cmd.Dir = resolveApplicationWorkingDirectory(ctx)
			if err := cmd.Start(); err != nil {
				return fmt.Errorf("启动失败: %w", err)
			}
			return nil
		}
		return withShellApartment(func() error {
			n := strings.TrimSpace(ctx.shellTarget)
			a := buildShellExecuteArguments(ctx.args)
			if requiresShellOpen(n) {
				a = ""
			}
			workingDir := resolveApplicationWorkingDirectory(ctx)
			return shellExecuteProgram("", n, a, workingDir)
		})
	}
	return startApplicationViaExplorer(ctx)
}

// startApplicationViaExplorer 经 explorer 上下文启动应用（COM ShellDispatch 降权链）。
// [S 汇编 0x1408a9400, 576B(0x240)]：TrimSpace(shellTarget) 空 → "应用路径不能为空"
// （@0x1408a95fb，24B @0x140c65190）；buildShellExecuteArguments（@0x1408a9463）；
// requiresShellOpen 真 → 清空 shellArgs；withExplorerShellDispatch 闭包（func1 0x1408a9780）
// → shellExecuteByExplorer(dispatch,name,shellArgs,workingDir,"",1)；成功 → nil；失败 →
// isProcessElevated 真 → startApplicationWithShellParent（@0x1408a9573）；否则
// withShellApartment 闭包（func2 0x1408a9640）→ shellExecuteProgram("",…)（@0x1408a9761）。
func startApplicationViaExplorer(ctx launchContext) error {
	name := strings.TrimSpace(ctx.shellTarget)
	if name == "" {
		return errors.New("应用路径不能为空")
	}
	shellArgs := buildShellExecuteArguments(ctx.args)
	if requiresShellOpen(name) {
		shellArgs = ""
	}
	if err := withExplorerShellDispatch(func(dispatch *ole.IDispatch) error {
		workingDir := resolveApplicationWorkingDirectory(ctx)
		return shellExecuteByExplorer(dispatch, name, shellArgs, workingDir, "", 1)
	}); err == nil {
		return nil
	}
	if isProcessElevated() {
		return startApplicationWithShellParent(ctx)
	}
	return withShellApartment(func() error {
		n := strings.TrimSpace(ctx.shellTarget)
		a := buildShellExecuteArguments(ctx.args)
		if requiresShellOpen(n) {
			a = ""
		}
		workingDir := resolveApplicationWorkingDirectory(ctx)
		return shellExecuteProgram("", n, a, workingDir)
	})
}

// startApplicationWithShellParent 经 shell parent 继承降权启动（CreateProcess 链）。
// [S 汇编 0x1408aa0e0, 448B(0x1c0)]：TrimSpace(shellTarget) 空 → "应用路径不能为空"
// （@0x1408aa251，24B @0x140c65190）；buildShellExecuteArguments（@0x1408aa143）；
// resolveApplicationWorkingDirectory（@0x1408aa193）；requiresShellOpen 真 →
// "无法通过安全的 Explorer Shell 启动该目标"（@0x1408aa1d8，52B @0x140c89916）；
// 否则 buildCreateProcessCommandLine（@0x1408aa215）→
// createProcessWithShellParentWithVisibility(name,cmdline,workingDir,false)（@0x1408aa243）。
func startApplicationWithShellParent(ctx launchContext) error {
	name := strings.TrimSpace(ctx.shellTarget)
	if name == "" {
		return errors.New("应用路径不能为空")
	}
	shellArgs := buildShellExecuteArguments(ctx.args)
	workingDir := resolveApplicationWorkingDirectory(ctx)
	if requiresShellOpen(name) {
		return errors.New("无法通过安全的 Explorer Shell 启动该目标")
	}
	cmdline := buildCreateProcessCommandLine(name, shellArgs)
	return createProcessWithShellParentWithVisibility(name, cmdline, workingDir, false)
}

// retryLaunchAsAdminIfElevationRequired 检测 740 提权错误并重试提权启动。
// [S 汇编 0x1408acd60, 416B(0x1a0)]：isElevationRequiredError(err)（@0x1408acd89）假 →
// 返回 err（@0x1408acdce）；真 → startApplicationWindowsAsAdmin(ctx)（@0x1408acdc0）。
func retryLaunchAsAdminIfElevationRequired(err error, ctx launchContext) error {
	if !isElevationRequiredError(err) {
		return err
	}
	return startApplicationWindowsAsAdmin(ctx)
}
