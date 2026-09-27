// com_explorer_windows.go — COM ShellDispatch 链（逆向还原）
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)
// 基线：docs/goresym/source_funcs.txt + docs/goresym/pipeline/tmp/*.asm.txt
//
// 域内函数（批次 118 落地，[S-sig] 核心控制流实证，go-ole COM 依赖）：
//   getDesktopExplorerShellDispatch    (0x1408ab360) 定位桌面 Explorer 的 Shell.Application
//   withDesktopExplorerShellDispatch   (0x1408ab2a0) 获取→defer Release→fn(dispatch)
//   withExplorerShellDispatch          (0x1408ab0e0) STA + CoInitialize + 委托
//   shellExecuteByExplorer             (0x1408abe60) IDispatch ShellExecute 动词
//
// 注：FindWindowSW 的 VARIANT 手动布局（pvarLoc/pvarLocRoot/swClass/pHWND/swfwOptions）
// 尚未逐字节取证，此处以 oleutil.CallMethod 高层语义等价实现，升格 [S] 需逐 VARIANT 对齐。

package main

import (
	"errors"
	"runtime"
	"syscall"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// getDesktopExplorerShellDispatch 定位桌面 Explorer 的 Shell.Application 调度接口。
// [S-sig 0x1408ab360, 2082B]：核心链实证 —
//
//	oleutil.CreateObject("Shell.Application")（@0x1408ab386，17B @0x140c573a7）→
//	QueryInterface（@0x1408ab3e3）→ Windows()（@0x1408ab4b5，7B "Windows"）→
//	FindWindowSW（@0x1408ab681，12B "FindWindowSW"）→ Document（@0x1408ab7c0，8B "Document"）→
//	Application（@0x1408ab900，11B "Application"）→ QueryInterface 返回 IDispatch。
//	错误消息实测：初始化/获取集合/定位/视图/启动对象/调度接口 六段。
func getDesktopExplorerShellDispatch() (*ole.IDispatch, error) {
	unknown, err := oleutil.CreateObject("Shell.Application")
	if err != nil {
		return nil, wrapWinError("初始化 Explorer Shell 失败", err)
	}
	defer unknown.Release()

	shell, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, wrapWinError("获取 Explorer Shell 调度接口失败", err)
	}
	defer shell.Release()

	windowsVar, err := oleutil.CallMethod(shell, "Windows")
	if err != nil {
		return nil, wrapWinError("获取 Explorer 窗口集合失败", err)
	}
	windows := windowsVar.ToIDispatch()
	if windows == nil {
		return nil, wrapWinError("获取 Explorer 窗口集合失败", errors.New("nil shell windows dispatch"))
	}
	defer windows.Release()

	// FindWindowSW 定位桌面窗口：SWC_DESKTOP(8) / SWFO_NEEDDISPATCH(1)。
	// [S-sig] VARIANT 精确布局待取证，此处按高层语义等价传参。
	hwndVar, err := oleutil.CallMethod(windows, "FindWindowSW", nil, nil, int32(8), int32(0), int32(1))
	if err != nil {
		return nil, wrapWinError("定位桌面 Explorer 窗口失败", err)
	}
	hwnd := hwndVar.Val

	itemVar, err := oleutil.CallMethod(windows, "Item", hwnd)
	if err != nil {
		return nil, wrapWinError("定位桌面 Explorer 窗口失败", err)
	}
	item := itemVar.ToIDispatch()
	if item == nil {
		return nil, wrapWinError("定位桌面 Explorer 窗口失败", errors.New("nil window item dispatch"))
	}
	defer item.Release()

	viewVar, err := oleutil.CallMethod(item, "Document")
	if err != nil {
		return nil, wrapWinError("获取桌面 Explorer 视图失败", err)
	}
	view := viewVar.ToIDispatch()
	if view == nil {
		return nil, wrapWinError("获取桌面 Explorer 视图失败", errors.New("nil document dispatch"))
	}
	defer view.Release()

	appVar, err := oleutil.CallMethod(view, "Application")
	if err != nil {
		return nil, wrapWinError("获取桌面 Explorer 启动对象失败", err)
	}
	app := appVar.ToIDispatch()
	if app == nil {
		return nil, wrapWinError("获取桌面 Explorer 启动对象失败", errors.New("nil application dispatch"))
	}
	return app, nil
}

// withDesktopExplorerShellDispatch 获取桌面 Explorer 调度接口并执行 fn。
// [S-sig 0x1408ab2a0, 128B]：getDesktopExplorerShellDispatch（@0x1408ab2c9）err 非 nil →
// 返回 err（@0x1408ab2d3）；否则 defer dispatch.Release()（@0x1408ab314）+ fn(dispatch)
// （@0x1408ab2fb）结果透传。
func withDesktopExplorerShellDispatch(fn func(*ole.IDispatch) error) error {
	dispatch, err := getDesktopExplorerShellDispatch()
	if err != nil {
		return err
	}
	defer dispatch.Release()
	return fn(dispatch)
}

// withExplorerShellDispatch 在 STA 内获取 Explorer 调度接口并执行 fn。
// [S-sig 0x1408ab0e0, 352B]：LockOSThread（@0x1408ab108）→ CoInitializeEx(0,6)
// （@0x1408ab147）；hr∉{0,1} → wrapWinError("初始化 Explorer Shell 失败")（@0x1408ab1eb）；
// defer CoUninitialize（deferwrap1 @0x1408ab240）+ UnlockOSThread（@0x1408ab1a7）；
// withDesktopExplorerShellDispatch(fn)（@0x1408ab183）。
func withExplorerShellDispatch(fn func(*ole.IDispatch) error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	hr, _, _ := procCoInitializeEx.Call(0, coinItApartmentThreaded|coinItDisableOle1DDE)
	if hr != 0 && hr != 1 {
		return wrapWinError("初始化 Explorer Shell 失败", syscall.Errno(hr))
	}
	defer procCoUninitialize.Call()
	return withDesktopExplorerShellDispatch(fn)
}

// shellExecuteByExplorer 经桌面 Explorer 调度接口执行 ShellExecute。
// [S 0x1408abe60, 520B]：IDispatch.InvokeWithOptionalArgs("ShellExecute",
// DISPATCH_METHOD, file, args, dir, verb, show)（@0x1408abf95，方法名 12B
// @0x140c4ba14，5 参）；err → wrapWinError("委托 Explorer 启动应用失败")
// （@0x1408ac013，34B @0x140c75991）；result 非 nil → VariantClear（@0x1408abfe7）。
//
// 参数序经两个独立调用点 asm 交叉实证（批次 120）：
//
//	startShellTargetViaExplorer.func1    (0x1408ab040)：rbx/rcx=file, rdi/rsi=args,
//	    r8/r9=dir, r10/r11=verb(空), [rsp]=show(1)
//	startApplicationViaExplorer.func1    (0x1408a9780)：rbx/rcx=file, rdi/rsi=args,
//	    r8/r9=dir(resolveApplicationWorkingDirectory), r10/r11=verb(空), [rsp]=show(1)
//
// 与 ShellExecute(sFile, vArguments, vDirectory, vOperation, vShow) MSDN 语义一致。
func shellExecuteByExplorer(dispatch *ole.IDispatch, file, args, dir, verb string, show int) error {
	result, err := dispatch.InvokeWithOptionalArgs("ShellExecute", ole.DISPATCH_METHOD,
		[]interface{}{file, args, dir, verb, show})
	if err != nil {
		return wrapWinError("委托 Explorer 启动应用失败", err)
	}
	if result != nil {
		result.Clear()
	}
	return nil
}
