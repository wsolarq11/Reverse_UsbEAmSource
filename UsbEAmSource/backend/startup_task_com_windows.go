// startup_task_com_windows.go — 开机自启任务域：Task Scheduler COM 基础设施
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)
// 基线：docs/goresym/pipeline/tmp/{withTaskSchedulerService,callTaskSchedulerMethod,
// callTaskSchedulerDispatchMethod,putTaskSchedulerProperty}.asm.txt
//
// 本域函数：
//   withTaskSchedulerService            (0x1408b14e0) COM 服务上下文包装
//   callTaskSchedulerMethod             (0x1408b1920) IDispatch 方法调用（无返回值提取）
//   callTaskSchedulerDispatchMethod     (0x1408b1a80) IDispatch 方法调用（提取调度对象）
//   getTaskSchedulerDispatchProperty    (0x1408b1b60) IDispatch 属性读取（提取调度对象）
//   taskSchedulerDispatchFromVariant    (0x1408b1c20) VARIANT→*ole.IDispatch 所有权转移
//   putTaskSchedulerProperty            (0x1408b1e00) IDispatch 属性写入

package main

import (
	"fmt"
	"runtime"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// iidTaskService 是 ITaskService COM 接口 GUID。
// [S] 常量 {2FABA4C7-4DA9-4013-9697-20CC3FD40F85}（Microsoft 公开接口定义）。
// 注：asm 经 .data 重定位指针（mov rbx,[rip+0x135e982] @0x1408b167f）间接装载；
// 磁盘重定位占位字节非明文，按公开 IID 常量还原。
var iidTaskService = ole.NewGUID("{2FABA4C7-4DA9-4013-9697-20CC3FD40F85}")

// withTaskSchedulerService 在 Task Scheduler COM 服务上下文中执行 fn。
// [S 汇编 0x1408b14e0, 848B(0x350)]：
//
//	runtime.LockOSThread（@0x1408b150e）+ defer UnlockOSThread（deferprocStack @0x1408b152a）；
//	ole.CoInitialize(nil)（@0x1408b152f）err→"初始化 Task Scheduler COM 失败: %w"
//	  （@0x1408b1574，39B @0x140c7de4d）；defer CoUninitialize（@0x1408b15c0）；
//	oleutil.CreateObject("Schedule.Service")（@0x1408b15d1，16B @0x140c551ea）err→
//	  "创建 Task Scheduler 服务对象失败: %w"（@0x1408b1616，44B @0x140c8307d）；
//	  defer obj.Release（@0x1408b1668）；
//	obj.QueryInterface(iidTaskService)（@0x1408b168e）err→"获取 Task Scheduler 调度接口失败: %w"
//	  （@0x1408b16d3，44B @0x140c830a9）；defer svc.Release（@0x1408b1734）；
//	callTaskSchedulerMethod(svc,"Connect",nil)（@0x1408b1759，7B @0x140c3948f）err→
//	  "连接 Task Scheduler 服务失败: %w"（@0x1408b17a0，38B @0x140c7b304）；
//	fn(svc)（@0x1408b17e6）结果透传。
func withTaskSchedulerService(fn func(*ole.IDispatch) error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := ole.CoInitialize(0); err != nil {
		return fmt.Errorf("初始化 Task Scheduler COM 失败: %w", err)
	}
	defer ole.CoUninitialize()

	obj, err := oleutil.CreateObject("Schedule.Service")
	if err != nil {
		return fmt.Errorf("创建 Task Scheduler 服务对象失败: %w", err)
	}
	defer obj.Release()

	svc, err := obj.QueryInterface(iidTaskService)
	if err != nil {
		return fmt.Errorf("获取 Task Scheduler 调度接口失败: %w", err)
	}
	defer svc.Release()

	if err := callTaskSchedulerMethod(svc, "Connect", nil); err != nil {
		return fmt.Errorf("连接 Task Scheduler 服务失败: %w", err)
	}
	return fn(svc)
}

// callTaskSchedulerMethod 调用 Task Scheduler 调度方法，丢弃返回值。
// [S 汇编 0x1408b1920, 352B(0x160)]：
//
//	svc.InvokeWithOptionalArgs(name, DISPATCH_METHOD, args)（@0x1408b1963，edi=1）；
//	result 非 nil → ole.VariantClear(result)（@0x1408b19ac，deferwrap1 0x1408b1a20）；
//	err 透传（@0x1408b19ae）。
func callTaskSchedulerMethod(svc *ole.IDispatch, name string, args []interface{}) error {
	result, err := svc.InvokeWithOptionalArgs(name, ole.DISPATCH_METHOD, args)
	if result != nil {
		ole.VariantClear(result)
	}
	return err
}

// callTaskSchedulerDispatchMethod 调用返回调度对象的方法，提取 *ole.IDispatch。
// [S 汇编 0x1408b1a80, 128B(0x80)]：
//
//	svc.InvokeWithOptionalArgs(name, DISPATCH_METHOD, args)（@0x1408b1aac，edi=1）；
//	err 非 nil → result 非 nil 则 VariantClear（@0x1408b1ac6）+ 返回 (nil,err)；
//	err nil → taskSchedulerDispatchFromVariant(result,name)（@0x1408b1af4）透传。
func callTaskSchedulerDispatchMethod(svc *ole.IDispatch, name string, args []interface{}) (*ole.IDispatch, error) {
	result, err := svc.InvokeWithOptionalArgs(name, ole.DISPATCH_METHOD, args)
	if err != nil {
		if result != nil {
			ole.VariantClear(result)
		}
		return nil, err
	}
	return taskSchedulerDispatchFromVariant(result, name)
}

// getTaskSchedulerDispatchProperty 读取返回调度对象的属性，提取 *ole.IDispatch。
// [S 汇编 0x1408b1b60, 128B(0x80)]：
//
//	svc.InvokeWithOptionalArgs(name, DISPATCH_PROPERTYGET, nil)（@0x1408b1b86，edi=2）；
//	err 非 nil → result 非 nil 则 VariantClear（@0x1408b1ba0）+ 返回 (nil,err)；
//	err nil → taskSchedulerDispatchFromVariant(result,name)（@0x1408b1bce）透传。
func getTaskSchedulerDispatchProperty(svc *ole.IDispatch, name string) (*ole.IDispatch, error) {
	result, err := svc.InvokeWithOptionalArgs(name, ole.DISPATCH_PROPERTYGET, nil)
	if err != nil {
		if result != nil {
			ole.VariantClear(result)
		}
		return nil, err
	}
	return taskSchedulerDispatchFromVariant(result, name)
}

// taskSchedulerDispatchFromVariant 从 VARIANT 提取调度对象并转移所有权。
// [S 汇编 0x1408b1c20, 368B(0x170)]：
//
//	result nil → "%s 未返回调度对象"（@0x1408b1d3a/1d75，24B @0x140c65220）；
//	VT==VT_DISPATCH(9) → Val（@0x1408b1c49/53）否则 nil；
//	pdisp nil → VariantClear(result)（@0x1408b1cd9）+ "%s 返回的调度对象为空"
//	  （@0x1408b1d20，30B @0x140c6f117）；
//	pdisp 非 nil → VT=VT_EMPTY、Val=0 转移所有权（@0x1408b1c66/6b），返回 (pdisp,nil)。
func taskSchedulerDispatchFromVariant(result *ole.VARIANT, name string) (*ole.IDispatch, error) {
	if result == nil {
		return nil, fmt.Errorf("%s 未返回调度对象", name)
	}
	pdisp := result.ToIDispatch()
	if pdisp == nil {
		ole.VariantClear(result)
		return nil, fmt.Errorf("%s 返回的调度对象为空", name)
	}
	result.VT = ole.VT_EMPTY
	result.Val = 0
	return pdisp, nil
}

// putTaskSchedulerProperty 写入 Task Scheduler 调度对象属性。
// [S 汇编 0x1408b1e00, 640B(0x280)]：
//
//	svc.InvokeWithOptionalArgs(name, DISPATCH_PROPERTYPUT, []interface{}{value})
//	  （@0x1408b1e6c，edi=4，r8d=1 单参）；err → "设置计划任务属性 %s 失败: %w"
//	  （@0x1408b1f67，38B @0x140c7b32a）；result 非 nil → VariantClear（@0x1408b1ec1/1f95）。
func putTaskSchedulerProperty(svc *ole.IDispatch, name string, value interface{}) error {
	result, err := svc.InvokeWithOptionalArgs(name, ole.DISPATCH_PROPERTYPUT, []interface{}{value})
	if err != nil {
		return fmt.Errorf("设置计划任务属性 %s 失败: %w", name, err)
	}
	if result != nil {
		ole.VariantClear(result)
	}
	return nil
}
