// pathclip_logic_windows.go — 路径/可执行目标判断链（逆向还原）
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)
// 基线：docs/goresym/source_funcs.txt + docs/goresym/pipeline/tmp/*.asm.txt
//
// 域内函数（批次 108 落地）：
//   shouldOpenPathViaExplorer        (0x1408a8f20)  提权 && 可执行目标
//   isWindowsExecutableOpenTarget    (0x1408a8f80)  SaferiIsExecutableFileType 优先，回退扩展名
//   saferIsExecutableFileType        (0x1408a9000)  advapi32.SaferiIsExecutableFileType
//   hasKnownWindowsExecutableExtension(0x1408a9100)  Windows 可执行扩展名集合
//   isWindowsAbsoluteFilesystemPath  (0x1408a9320)  UNC / 盘符路径判定
//   getTokenElevation                (0x1408accc0)  GetTokenInformation(TokenElevation)
//   isProcessElevated                (0x1408ad900)  OpenProcessToken + getTokenElevation

package main

import (
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// advapi32DLL 及 SaferiIsExecutableFileType 代理。
var (
	advapi32DLL                    = windows.NewLazySystemDLL("advapi32.dll")
	procSaferiIsExecutableFileType = advapi32DLL.NewProc("SaferiIsExecutableFileType")
)

// shouldOpenPathViaExplorer 判定是否应经资源管理器打开路径（提权进程打开可执行目标）。
// [S 汇编 0x1408a8f20, 96B] isProcessElevated（@0x1408a8f38）非提权 → false；
// 提权 → isWindowsExecutableOpenTarget(path)（@0x1408a8f52）。
func shouldOpenPathViaExplorer(path string) bool {
	if !isProcessElevated() {
		return false
	}
	return isWindowsExecutableOpenTarget(path)
}

// isWindowsExecutableOpenTarget 判定路径是否为 Windows 可执行打开目标。
// [S 汇编 0x1408a8f80, 128B] TrimSpace → isWindowsAbsoluteFilesystemPath 非真 → false；
// saferIsExecutableFileType（@0x1408a8fb5）err==nil（test rbx @0x1408a8fba）→ 返回其 bool；
// err 非空 → 回退 hasKnownWindowsExecutableExtension（@0x1408a8fc9）。
func isWindowsExecutableOpenTarget(path string) bool {
	path = strings.TrimSpace(path)
	if !isWindowsAbsoluteFilesystemPath(path) {
		return false
	}
	if ok, err := saferIsExecutableFileType(path); err == nil {
		return ok
	}
	return hasKnownWindowsExecutableExtension(path)
}

// saferIsExecutableFileType 用 advapi32.SaferiIsExecutableFileType 判定可执行文件类型。
// [S 汇编 0x1408a9000, 256B] LazyProc.Find（@0x1408a9023）失败 → (false,err)；
// UTF16FromString（@0x1408a9040）→ Call(&utf16[0],0)（@0x1408a90aa）→ 返回 ret!=0。
func saferIsExecutableFileType(path string) (bool, error) {
	if err := procSaferiIsExecutableFileType.Find(); err != nil {
		return false, err
	}
	path16, err := syscall.UTF16FromString(path)
	if err != nil {
		return false, err
	}
	ret, _, _ := procSaferiIsExecutableFileType.Call(uintptr(unsafe.Pointer(&path16[0])), 0)
	return ret != 0, nil
}

// hasKnownWindowsExecutableExtension 判定扩展名是否在 Windows 可执行扩展名集合。
// [S 汇编 0x1408a9100, 544B] TrimSpace → 从尾部扫 '.'（遇 '\\'/'/' 中止 @0x1408a912e/9133）→
// ToLower（@0x1408a9160）→ len 3/4/10 三档比较（编译器二分树）。
func hasKnownWindowsExecutableExtension(path string) bool {
	path = strings.TrimSpace(path)
	ext := ""
	for i := len(path) - 1; i >= 0; i-- {
		c := path[i]
		if c == '\\' || c == '/' {
			break
		}
		if c == '.' {
			ext = path[i:]
			break
		}
	}
	switch strings.ToLower(ext) {
	case ".vb", ".ws",
		".bat", ".cmd", ".com", ".cpl", ".lnk", ".pif", ".scr", ".url",
		".vbe", ".vbs", ".exe", ".hta", ".jse", ".msc", ".msi", ".msp",
		".ps1", ".wsc", ".wsf", ".wsh",
		".appref-ms":
		return true
	}
	return false
}

// isWindowsAbsoluteFilesystemPath 判定 UNC 或盘符绝对路径。
// [S 汇编 0x1408a9320, 224B] TrimSpace 空 → false；len>=2 且 "\\\\" 或 "//"（cmp word
// 0x5c5c/0x2f2f @0x1408a934f/9366）→ true；len>=3 且 [A-Za-z]:[\\/]（@0x1408a9387 起）。
func isWindowsAbsoluteFilesystemPath(path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}
	if len(path) >= 2 {
		if path[0] == '\\' && path[1] == '\\' {
			return true
		}
		if path[0] == '/' && path[1] == '/' {
			return true
		}
	}
	if len(path) >= 3 {
		c := path[0]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			if path[1] == ':' {
				return path[2] == '\\' || path[2] == '/'
			}
		}
	}
	return false
}

// getTokenElevation 读取令牌的 TokenElevation 值。
// [S 汇编 0x1408accc0, 160B] GetTokenInformation(token,0x14,&elevation,4,&size)（@0x1408accf2，
// ebx=0x14=TokenElevation）失败 → wrapWinError("读取令牌信息失败")；成功 → elevation!=0。
func getTokenElevation(token windows.Token) (bool, error) {
	var elevation uint32
	var size uint32
	err := windows.GetTokenInformation(token, windows.TokenElevation, (*byte)(unsafe.Pointer(&elevation)), 4, &size)
	if err != nil {
		return false, wrapWinError("读取令牌信息失败", err)
	}
	return elevation != 0, nil
}

// isProcessElevated 判定当前进程是否以提升权限运行。
// [S 汇编 0x1408ad900, 224B] OpenProcessToken(-1,0x8,&token)（@0x1408ad940）失败 → false；
// defer token.Close()；getTokenElevation（@0x1408ad980）→ elevated && err==nil
// （sete cl / and ecx,eax @0x1408ad988/98b）。
func isProcessElevated() bool {
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token); err != nil {
		return false
	}
	defer token.Close()
	elevated, err := getTokenElevation(token)
	return elevated && err == nil
}
