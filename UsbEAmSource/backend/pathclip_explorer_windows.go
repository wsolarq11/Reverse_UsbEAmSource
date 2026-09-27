// pathclip_explorer_windows.go — 资源管理器启动/定位 + Windows 参数转义（逆向还原）
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)
// 基线：docs/goresym/source_funcs.txt + docs/goresym/pipeline/tmp/*.asm.txt
//
// 域内函数（批次 109 落地）：
//   quoteWindowsArgument          (0x1408ad280)  Unicode 感知 Windows 参数转义
//   revealPathInExplorer          (0x1408a8d20)  explorer.exe /select, 定位
//   startShellTargetViaExplorer   (0x1408aaee0)  Explorer ShellDispatch 启动（[S-sig] 体待 COM 链）

package main

import (
	"errors"
	"strings"

	ole "github.com/go-ole/go-ole"
)

// quoteWindowsArgument 按 MS CRT 规则转义 Windows 命令行参数（Unicode 感知）。
// [S 汇编 0x1408ad280, 1632B] 空 → `""`（@0x1408ad3c5，2B @0x1411cace8）；
// IndexAny(arg," \t\n\v\"")<0（@0x1408ad2c7，字符集 5B @0x140c35d0d）→ 原样返回；
// 否则 Builder.Grow(len+8) → 前缀 `"` → 逐 rune（decoderune @0x1408ad402）：
//
//	`\` 计数（@0x1408ad430）；`"` → Repeat(`\`,2*n+1)+`"`（@0x1408ad594）；其余 →
//	Repeat(`\`,n)+WriteRune（@0x1408ad467/568）；结尾残留反斜杠 Repeat(`\`,2*n)（@0x1408ad6d3）
//	→ 后缀 `"`（@0x1408ad843）。
func quoteWindowsArgument(arg string) string {
	if arg == "" {
		return `""`
	}
	if !strings.ContainsAny(arg, " \t\n\x0b\"") {
		return arg
	}
	var b strings.Builder
	b.Grow(len(arg) + 8)
	b.WriteByte('"')
	var backslashes int
	for _, r := range arg {
		switch r {
		case '\\':
			backslashes++
		case '"':
			b.WriteString(strings.Repeat(`\`, 2*backslashes+1))
			b.WriteRune(r)
			backslashes = 0
		default:
			if backslashes > 0 {
				b.WriteString(strings.Repeat(`\`, backslashes))
				backslashes = 0
			}
			b.WriteRune(r)
		}
	}
	if backslashes > 0 {
		b.WriteString(strings.Repeat(`\`, 2*backslashes))
	}
	b.WriteByte('"')
	return b.String()
}

// revealPathInExplorer 在资源管理器中定位路径（explorer /select,）。
// [S 汇编 0x1408a8d20, 128B] TrimSpace → quoteWindowsArgument（@0x1408ad38）→
// concatstring2("/select,",quoted)（@0x1408ad54，"/select," 8B @0x140c3c24c）→
// withShellApartment(shellExecuteProgram("","explorer.exe",args,""))（func1 @0x1408a8da0，
// program="explorer.exe" 12B @0x140c4b9fc）。
func revealPathInExplorer(path string) error {
	quoted := quoteWindowsArgument(strings.TrimSpace(path))
	args := "/select," + quoted
	return withShellApartment(func() error {
		return shellExecuteProgram("", "explorer.exe", args, "")
	})
}

// startShellTargetViaExplorer 经资源管理器 ShellDispatch 启动目标。
// [S 0x1408aaee0, 352B]：TrimSpace(shellTarget) 空 → errors.New("Shell 目标不能为空")
// （@0x1408aaf36，24B @0x140c651a8）；buildShellExecuteArguments(args)（@0x1408aaf7a）；
// withExplorerShellDispatch 闭包（func1 @0x1408ab040）：TrimSpace(workingDir) →
// shellExecuteByExplorer(dispatch, name, shellArgs, dir, "", 1)（verb 恒空 @0x1408ab0a4/0a7，
// show=1 @0x1408ab087）。
func startShellTargetViaExplorer(shellTarget string, args []string, workingDir string) error {
	name := strings.TrimSpace(shellTarget)
	if name == "" {
		return errors.New("Shell 目标不能为空")
	}
	shellArgs := buildShellExecuteArguments(args)
	return withExplorerShellDispatch(func(dispatch *ole.IDispatch) error {
		dir := strings.TrimSpace(workingDir)
		return shellExecuteByExplorer(dispatch, name, shellArgs, dir, "", 1)
	})
}
