// launcherprocess.go — 应用启动上下文域（逆向还原，跨平台）
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)
// 基线：docs/goresym/source_funcs.txt + docs/goresym/pipeline/tmp/*.asm.txt
//
// 域内函数（批次 121 落地）：
//   launchContext                        (0x0)      启动上下文结构体（360B，字段偏移 asm 实证）
//   shouldUseSavedShortcutResolution     (0x1408a7e60) 快捷方式失效回退判定
//   resolveLaunchableAppEntry            (0x1408a75e0) saved-shortcut 入口解析
//   resolveLaunchableAppEntryForExplicitArgs (0x1408a79e0) 显式参数入口解析（拖拽文件）
//
// 结构体字段偏移三函数交叉实证（resolveLaunchableAppEntry / startApplicationViaExplorer /
// startApplicationUsingCurrentPrivileges 一致）：entryType@0x20、shellTarget@0xa0、entry@0xb0、
// args@0xc0、rawName@0x138、rawEntry@0x148、rawCommandLine@0x158。
// 结构体 360B=0x168（45 词）：duffcopy+0x24c 复制 352B（44 词）+ 第 0 词单独 mov。
// 中间 30 词（0x00-0x1f、0x30-0x9f、0xd8-0x137）为启动元数据 / droppedFiles / privilege 等，
// 待 buildAppEntryWithDroppedFiles 完整取证。

package main

import (
	"os"
	"strings"
)

// launchContext 应用启动上下文。字段偏移仅来自 asm 交叉实证，禁止凭字段名臆测。
// 待取证区用 _ [N]uintptr 占位，锁定偏移；语义补齐于 buildAppEntryWithDroppedFiles 批次。
type launchContext struct {
	_              [4]uintptr  // 0x00-0x1f 待取证
	entryType      string      // 0x20-0x2f
	_              [14]uintptr // 0x30-0x9f 待取证
	shellTarget    string      // 0xa0-0xaf（appName；解析结果写回）
	entry          string      // 0xb0-0xbf（解析结果写回）
	args           []string    // 0xc0-0xcf（解析结果写回）
	_              [12]uintptr // 0xd8-0x137 待取证
	rawName        string      // 0x138-0x147（原始 appName）
	rawEntry       string      // 0x148-0x157（原始 entry，entry 空时回退）
	rawCommandLine string      // 0x158-0x167（原始命令行，args 空时回退）
}

// shouldUseSavedShortcutResolution 判定是否走「保存的快捷方式解析」回退。
// [S 汇编 0x1408a7e60, 128B]：TrimSpace(path)（@0x1408a7e73）空 → true（@0x1408a7eb3）；
// !isShortcutFilePath(name)（@0x1408a7e87）→ false（@0x1408a7eab）；
// os.Stat(name)（@0x1408a7e9a）err → true，否则 false（setne @0x1408a7ea2）。
func shouldUseSavedShortcutResolution(path string) bool {
	name := strings.TrimSpace(path)
	if name == "" {
		return true
	}
	if !isShortcutFilePath(name) {
		return false
	}
	_, err := os.Stat(name)
	return err != nil
}

// resolveLaunchableAppEntry 解析启动入口（saved-shortcut 回退）。
// [S 汇编 0x1408a75e0, 1024B]：normalizeAppEntryType(entryType)=="directory"（@0x1408a7680，9B
// "directory"）→ 原样返回；TrimSpace(rawName)（@0x1408a772c，rawName@0x138）空 → 原样返回；
// !shouldUseSavedShortcutResolution(TrimSpace(shellTarget))（@0x1408a7754/59，shellTarget@0xa0）
// → 原样返回；否则 shellTarget=TrimSpace(rawName)（@0x1408a7795，写回 0xa0）；TrimSpace(entry)
// （@0x1408a77c0，entry@0xb0）空 → entry=TrimSpace(rawEntry)（@0x1408a77da，rawEntry@0x148）；
// cleanStringList(args)（@0x1408a7807，args@0xc0）空 → args=splitCommandLineArguments(
// rawCommandLine)（@0x1408a7821，rawCommandLine@0x158）。
func resolveLaunchableAppEntry(ctx launchContext) launchContext {
	if normalizeAppEntryType(ctx.entryType) == "directory" {
		return ctx
	}
	name := strings.TrimSpace(ctx.rawName)
	if name == "" {
		return ctx
	}
	if !shouldUseSavedShortcutResolution(strings.TrimSpace(ctx.shellTarget)) {
		return ctx
	}
	ctx.shellTarget = name
	if strings.TrimSpace(ctx.entry) == "" {
		ctx.entry = strings.TrimSpace(ctx.rawEntry)
	}
	if len(cleanStringList(ctx.args)) == 0 {
		ctx.args = splitCommandLineArguments(ctx.rawCommandLine)
	}
	return ctx
}

// resolveLaunchableAppEntryForExplicitArgs 解析显式参数入口（拖拽文件）。
// [S 汇编 0x1408a79e0, 1152B]：normalizeAppEntryType(entryType)=="directory"（@0x1408a7a83）
// → 原样返回；先 resolveLaunchableAppEntry（@0x1408a7b53）；!requiresShellOpen(TrimSpace(
// shellTarget))（@0x1408a7ba3/a8）→ 返回 resolveLaunchableAppEntry 结果；TrimSpace(rawName)
// （@0x1408a7bc5，rawName@0x138）空 → 返回 resolveLaunchableAppEntry 结果；否则 shellTarget=
// TrimSpace(rawName)（@0x1408a7c13，写回 0xa0）；TrimSpace(entry)（@0x1408a7c33，entry@0xb0）
// 空 → entry=TrimSpace(rawEntry)（@0x1408a7c4d，rawEntry@0x148）；cleanStringList(args)
// （@0x1408a7c7a，args@0xc0）空 → args=splitCommandLineArguments(rawCommandLine)
// （@0x1408a7c95，rawCommandLine@0x158）。
func resolveLaunchableAppEntryForExplicitArgs(ctx launchContext) launchContext {
	if normalizeAppEntryType(ctx.entryType) == "directory" {
		return ctx
	}
	resolved := resolveLaunchableAppEntry(ctx)
	if !requiresShellOpen(strings.TrimSpace(resolved.shellTarget)) {
		return resolved
	}
	if strings.TrimSpace(ctx.rawName) == "" {
		return resolved
	}
	resolved.shellTarget = strings.TrimSpace(ctx.rawName)
	if strings.TrimSpace(resolved.entry) == "" {
		resolved.entry = strings.TrimSpace(ctx.rawEntry)
	}
	if len(cleanStringList(resolved.args)) == 0 {
		resolved.args = splitCommandLineArguments(ctx.rawCommandLine)
	}
	return resolved
}

// buildAppEntryWithDroppedFiles 用拖拽文件构建启动入口。
// [S 汇编 0x1408a6c20, 1024B(0x400)]：normalizeAppEntryType(entryType)=="directory"
// （@0x1408a6cc8）→ 原样返回；normalizePathList(droppedFiles)（@0x1408a6d74）空 →
// 原样返回（@0x1408a6d83）；resolveLaunchableAppEntryForExplicitArgs（@0x1408a6dd3）→
// resolved.args = cleanStringList(append(resolved.args, paths...))（growslice @0x1408a6e47、
// typedslicecopy @0x1408a6eab、cleanStringList @0x1408a6ec8）。
func buildAppEntryWithDroppedFiles(droppedFiles []string, ctx launchContext) launchContext {
	if normalizeAppEntryType(ctx.entryType) == "directory" {
		return ctx
	}
	paths := normalizePathList(droppedFiles)
	if len(paths) == 0 {
		return ctx
	}
	resolved := resolveLaunchableAppEntryForExplicitArgs(ctx)
	resolved.args = cleanStringList(append(resolved.args, paths...))
	return resolved
}
