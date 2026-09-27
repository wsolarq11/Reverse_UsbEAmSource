// AUTO-RECONSTRUCTED — BootstrapService path/link/clipboard 方法
// 研究用途
//
// 原始契约：source_funcs.txt path_launcher_windows.go / fileclipboard.go 等
// 档位：[S] 反汇编实证
package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// OpenPath 打开路径。
// [S 汇编 0x1408a6760]
func (bs *BootstrapService) OpenPath(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return newPathError()
	}
	return openWithDefaultHandler(path)
}

// OpenPathDirectory 打开路径目录。
// [S 汇编 0x1408a6800]
func (bs *BootstrapService) OpenPathDirectory(path string) error {
	return openPathDirectory(path)
}

// OpenPathCommandLine 在命令行中打开路径。
// [S 汇编 0x1408a6860]
func (bs *BootstrapService) OpenPathCommandLine(path string) error {
	return openPathCommandLine(path)
}

// OpenPathCommandLineAdmin 以管理员身份在命令行中打开路径。
// [S 汇编 0x1408a68c0]
func (bs *BootstrapService) OpenPathCommandLineAdmin(path string) error {
	return openPathCommandLineAdmin(path)
}

// SetFileClipboard 设置文件剪贴板。
// [S 汇编 0x1408a69a0] 实证：(paths []string, dropEffect string) 两参转发。
func (bs *BootstrapService) SetFileClipboard(paths []string, dropEffect string) error {
	return setFileClipboard(paths, dropEffect)
}

// SetSearchCategoryShortcutActive 设置搜索类别快捷方式激活。
// [S 汇编 0x140790e40]
func (bs *BootstrapService) SetSearchCategoryShortcutActive(active bool) {
	bs.lock.Lock()
	bs.searchCategoryShortcutActive = active
	bs.lock.Unlock()
}

// ---- 子服务包级函数（ShellExecuteW 域，签名经符号表 + [S] 调用方实证）----

// openWithDefaultHandler 用默认处理器打开路径。
// [S 汇编 0x1408a8de0, 256B] TrimSpace 空 → newPathError；shouldOpenPathViaExplorer 真 →
// startShellTargetViaExplorer(path,nil,"")（rax/rbx=path, rcx/rdi/rsi=args, r8/r9=workingDir 空）；
// 否则 withShellApartment(shellExecuteProgram("",path,"",""))（闭包 func1 @0x1408a8ee0）。
func openWithDefaultHandler(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return newPathError()
	}
	if shouldOpenPathViaExplorer(path) {
		return startShellTargetViaExplorer(path, nil, "")
	}
	return withShellApartment(func() error {
		return shellExecuteProgram("", path, "", "")
	})
}

// openPathDirectory 在资源管理器中打开目录。
// [S 汇编 0x1408a88a0, 544B] TrimSpace 空 → newPathError；os.Stat（@0x1408a88d3）成功 →
// IsDir 真 → withShellApartment(shellExecuteProgram("",path,"",""))（func1 @0x1408a8a60），
// 非目录 → revealPathInExplorer（@0x1408a898c）；Stat 失败 → resolveOpenLocationDirectory 空 →
// newPathError，非空 → withShellApartment(shellExecuteProgram("",dir,"",""))（func2 @0x1408a8a00）。
func openPathDirectory(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return newPathError()
	}
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			return withShellApartment(func() error {
				return shellExecuteProgram("", path, "", "")
			})
		}
		return revealPathInExplorer(path)
	}
	dir := resolveOpenLocationDirectory(path)
	if dir == "" {
		return newPathError()
	}
	return withShellApartment(func() error {
		return shellExecuteProgram("", dir, "", "")
	})
}

// resolveOpenLocationDirectory 解析路径的「定位目录」。
// [S 汇编 0x1408a87c0, 224B] TrimSpace 空 → ""；os.Stat（@0x1408a87ef）成功 →
// IsDir 真 → 返回 path，非目录 → filepath.Dir（@0x1408a8852）；Stat 失败 →
// looksLikeFilesystemPath 假 → ""，真 → filepath.Dir（@0x1408a8816）。
func resolveOpenLocationDirectory(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			return path
		}
		return filepath.Dir(path)
	}
	if looksLikeFilesystemPath(path) {
		return filepath.Dir(path)
	}
	return ""
}

// openPathCommandLine 在命令行中打开路径（非提权直启 / 提权走 explorer 降权）。
// [S 汇编 0x1408a8ac0, 288B] resolveOpenLocationDirectory 空 → errors.New("位置不能为空")；
// isProcessElevated（@0x1408a8af3）真 → openUnelevatedCommandLine；假 →
// startDetachedCommand([cmd,"/k","pushd",dir])（4 元素 slice @0x1408a8b76，
// "/k" 2B @0x140c336b3 / "pushd" 5B @0x140c35d03）。
func openPathCommandLine(path string) error {
	dir := resolveOpenLocationDirectory(path)
	if dir == "" {
		return errors.New("位置不能为空")
	}
	if isProcessElevated() {
		return openUnelevatedCommandLine(dir)
	}
	return startDetachedCommand([]string{resolveCmdExePath(), "/k", "pushd", dir})
}

// openPathCommandLineAdmin 以管理员在命令行打开路径（runas + pushd）。
// [S 汇编 0x1408a8be0, 512B] resolveOpenLocationDirectory 空 → errors.New("位置不能为空")；
// withShellApartment(shellExecuteProgram("runas", cmd, "/k pushd "+quoteCmdArgument(dir), ""))
// （func1 @0x1408a8c80，verb="runas" 5B @0x140c35d08，前缀 "/k pushd " 9B @0x140c3f784）。
func openPathCommandLineAdmin(path string) error {
	dir := resolveOpenLocationDirectory(path)
	if dir == "" {
		return errors.New("位置不能为空")
	}
	return withShellApartment(func() error {
		cmd := resolveCmdExePath()
		args := "/k pushd " + quoteCmdArgument(dir)
		return shellExecuteProgram("runas", cmd, args, "")
	})
}

// openUnelevatedCommandLine 降权启动命令行（explorer 作 shell parent）。
// [S 汇编 0x1408aa000, 224B] resolveCmdExePath（@0x1408aa020）→ "/k pushd "+
// quoteCmdArgument(dir)（@0x1408aa060，前缀 9B @0x140c3f784）→
// buildCreateProcessCommandLine(cmd,args)（@0x1408aa075）→
// createProcessWithShellParentWithVisibility(cmd,cmdLine,dir,false)（@0x1408aa09a，r9=0）。
func openUnelevatedCommandLine(dir string) error {
	cmd := resolveCmdExePath()
	args := "/k pushd " + quoteCmdArgument(dir)
	cmdLine := buildCreateProcessCommandLine(cmd, args)
	return createProcessWithShellParentWithVisibility(cmd, cmdLine, dir, false)
}

// setFileClipboard 已迁至 pathclip_clipboard_windows.go（批次 102 落地）。

// newPathError 构造路径错误。 [S 内联于 OpenPath(0x1408a6760)]：asm 实证为
// errors.New("路径不能为空")（newobject + errorString{s: 18B 消息 @0x140c59e5e}）。
func newPathError() error {
	return errors.New("路径不能为空")
}
