// shell_execute_windows.go — Shell 执行域（逆向还原）
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)
// 基线：docs/goresym/source_funcs.txt + docs/goresym/pipeline/tmp/*.asm.txt
//
// 域内函数（批次 107 落地）：
//   shellExecuteProgram (0x1408aec40)    ShellExecuteExW 编排
//   showShellProperties (0x1408aeea0)    SHObjectProperties 编排
//   withShellApartment (0x1408af020)     COM 单线程公寓 + CoInitializeEx/CoUninitialize
//   resolveCmdExePath   (0x1408af1c0)    ComSpec/SystemRoot/cmd.exe 兜底链
//   utf16PtrOrNil       (0x1408af2a0)    TrimSpace 空串→nil，否则 UTF16 首元素指针
//   quoteCmdArgument    (0x1408af320)    CMD 引号转义（"" 表示字面引号）
//   startPowerShellEncodedCommand (0x1408af500)  PowerShell -EncodedCommand 启动
//   encodePowerShellCommand       (0x1408af5e0)  UTF-16LE → base64

package main

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ========================================================================
// Windows API 代理（OLE32 + SHELL32 延迟加载；shell32DLL 复用 appicon_windows.go）
// ========================================================================

var (
	ole32DLL               = windows.NewLazySystemDLL("ole32.dll")
	procCoInitializeEx     = ole32DLL.NewProc("CoInitializeEx")
	procCoUninitialize     = ole32DLL.NewProc("CoUninitialize")
	procShellExecuteExW    = shell32DLL.NewProc("ShellExecuteExW")
	procSHObjectProperties = shell32DLL.NewProc("SHObjectProperties")
)

// SHELLEXECUTEINFO fMask 位（ShellExecuteExW）。
const (
	seeMaskInvokeIDList uint32 = 0x0C
	seeMaskNoAsync      uint32 = 0x100
	seeMaskFlagNoUI     uint32 = 0x400
)

// SHObjectProperties 对象类型（shell32.SHObjectProperties）。
const (
	shopFilePath   uintptr = 2 // SHOP_FILEPATH
	shopVolumeGuid uintptr = 4 // SHOP_VOLUMEGUID
)

// CoInitializeEx dwCoInit 位。
const (
	coinItApartmentThreaded uintptr = 0x2
	coinItDisableOle1DDE    uintptr = 0x4
)

// shellExecuteProgram 用 ShellExecuteExW 执行程序。
// [S 汇编 0x1408aec40, 608B] UTF16FromString(program)（@0x1408aec84）→ lpFile；
// utf16PtrOrNil(verb/args/workingDir)（@0x1408aecb8/aece0/aed03）→ lpVerb/lpParameters/lpDirectory；
// SHELLEXECUTEINFOW cbSize=0x70（movabs 0x50c00000070 @0x1408aed25 → cbSize=0x70/fMask=0x50c），
// nShow=1（@0x1408aed89）；ShellExecuteExW 返回 0 → wrapWinError("调用系统操作失败")。
func shellExecuteProgram(verb, program, args, workingDir string) error {
	program16, err := syscall.UTF16FromString(program)
	if err != nil {
		return err
	}
	verb16, err := utf16PtrOrNil(verb)
	if err != nil {
		return err
	}
	args16, err := utf16PtrOrNil(args)
	if err != nil {
		return err
	}
	dir16, err := utf16PtrOrNil(workingDir)
	if err != nil {
		return err
	}
	sei := &shellExecuteInfo{
		cbSize:       uint32(unsafe.Sizeof(shellExecuteInfo{})),
		fMask:        seeMaskInvokeIDList | seeMaskNoAsync | seeMaskFlagNoUI,
		lpVerb:       verb16,
		lpFile:       &program16[0],
		lpParameters: args16,
		lpDirectory:  dir16,
		nShow:        1, // SW_SHOWNORMAL
	}
	ret, _, callErr := procShellExecuteExW.Call(uintptr(unsafe.Pointer(sei)))
	if ret == 0 {
		return wrapWinError("调用系统操作失败", callErr)
	}
	return nil
}

// showShellProperties 用 SHObjectProperties 显示文件/卷属性窗口。
// [S 汇编 0x1408aeea0, 384B] VolumeName（@0x1408aeec0）；卷根判定 TrimRight(path,`\/`)+EqualFold
// （@0x1408aef00/0x1408aef0f）→ SHOP_VOLUMEGUID(4) 且 name=vol，否则 SHOP_FILEPATH(2) 且 name=path；
// UTF16FromString(name)（@0x1408aef47）；Call(0,objType,&utf16[0],0)（@0x1408aefb7）；
// 返回 0 → wrapWinError("打开属性窗口失败")。
func showShellProperties(path string) error {
	vol := filepath.VolumeName(path)
	objType := shopFilePath
	name := path
	if vol != "" {
		trimmed := strings.TrimRight(path, `\/`)
		if strings.EqualFold(vol, trimmed) {
			objType = shopVolumeGuid
			name = vol
		}
	}
	name16, err := syscall.UTF16FromString(name)
	if err != nil {
		return err
	}
	ret, _, callErr := procSHObjectProperties.Call(0, objType, uintptr(unsafe.Pointer(&name16[0])), 0)
	if ret == 0 {
		return wrapWinError("打开属性窗口失败", callErr)
	}
	return nil
}

// withShellApartment 在 STA（单线程公寓）内执行 fn。
// [S 汇编 0x1408af020, 320B] LockOSThread（@0x1408af048）→ CoInitializeEx(0,6)（@0x1408af087，
// 6=COINIT_APARTMENTTHREADED|COINIT_DISABLE_OLE1DDE）；hr∈{S_OK(0),S_FALSE(1)} → defer CoUninitialize
// （@0x1408af097/af100）；defer UnlockOSThread（@0x1408af112）；fn() 结果透传（@0x1408af0d6）。
func withShellApartment(fn func() error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	hr, _, _ := procCoInitializeEx.Call(0, coinItApartmentThreaded|coinItDisableOle1DDE)
	if hr == 0 || hr == 1 { // S_OK 或 S_FALSE
		defer procCoUninitialize.Call()
	}
	return fn()
}

// resolveCmdExePath 解析 cmd.exe 路径（ComSpec → SystemRoot\System32 → cmd.exe 兜底）。
// [S 汇编 0x1408af1c0, 224B] os.Getenv("ComSpec")+TrimSpace（@0x1408af1e0）非空即返回；
// 否则 os.Getenv("SystemRoot")+TrimSpace（@0x1408af200）非空 → filepath.Join(...,"System32","cmd.exe")
// （@0x1408af262）；最终兜底返回 "cmd.exe"（@0x1408af26d）。
func resolveCmdExePath() string {
	if s := strings.TrimSpace(os.Getenv("ComSpec")); s != "" {
		return s
	}
	if s := strings.TrimSpace(os.Getenv("SystemRoot")); s != "" {
		return filepath.Join(s, "System32", "cmd.exe")
	}
	return "cmd.exe"
}

// utf16PtrOrNil 空/全空白串返回 nil 指针，否则 UTF-16 首元素指针。
// [S 汇编 0x1408af2a0, 128B] TrimSpace（@0x1408af2b3）空 → (nil,nil)；UTF16FromString（@0x1408af2c0）
// 返回 &utf16[0]（rax=slice.ptr）。
func utf16PtrOrNil(s string) (*uint16, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	u, err := syscall.UTF16FromString(s)
	if err != nil {
		return nil, err
	}
	return &u[0], nil
}

// quoteCmdArgument CMD 参数引号包裹与转义（双引号内 "" 表示字面引号）。
// [S 汇编 0x1408af320, 128B] TrimSpace（@0x1408af333）→ Replace(s,`"`,`""`,-1)（@0x1408af358，
// old=`"` 1B/new=`""` 2B/r9=-1）→ concatstring3(`"`,s,`"`)（@0x1408af377）。
func quoteCmdArgument(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Replace(s, `"`, `""`, -1)
	return `"` + s + `"`
}

// startPowerShellEncodedCommand 用 PowerShell -EncodedCommand 启动命令。
// [S 汇编 0x1408af500, 224B] encodePowerShellCommand（@0x1408af517）→
// startDetachedCommand(["powershell.exe","-NoProfile","-Sta","-EncodedCommand",encoded])（@0x1408af5a3）。
func startPowerShellEncodedCommand(command string) error {
	encoded := encodePowerShellCommand(command)
	return startDetachedCommand([]string{
		"powershell.exe", "-NoProfile", "-Sta", "-EncodedCommand", encoded,
	})
}

// encodePowerShellCommand UTF-16LE（含 NUL 终止）→ base64（PowerShell -EncodedCommand 编码）。
// [S 汇编 0x1408af5e0, 320B] stringtoslicerune（@0x1408af60d）→ utf16.Encode（@0x1408af612）→
// makeslice([]byte,2*len)（@0x1408af664，rcx=rbx+rbx @0x1408af617）→ 逐 uint16 小端写 2 字节
// （mov word @0x1408af688）→ base64.StdEncoding.EncodeToString（@0x1408af6cf）。
func encodePowerShellCommand(command string) string {
	utf16s := utf16.Encode([]rune(command))
	buf := make([]byte, 2*len(utf16s))
	for i, u := range utf16s {
		buf[2*i] = byte(u)
		buf[2*i+1] = byte(u >> 8)
	}
	return base64.StdEncoding.EncodeToString(buf)
}
