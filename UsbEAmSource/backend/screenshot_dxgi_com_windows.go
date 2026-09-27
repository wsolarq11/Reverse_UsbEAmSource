package main

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// DXGI 工厂接口标识与 QueryInterface/Release vtable 偏移约定。
var (
	iidIDXGIFactory1 = windows.GUID{Data1: 0x770aae78, Data2: 0xf26f, Data3: 0x4dba, Data4: [8]byte{0xa8, 0x29, 0x25, 0x3c, 0x83, 0xd1, 0xb3, 0x87}}
	iidIDXGIFactory6 = windows.GUID{Data1: 0xc1b6694f, Data2: 0xff09, Data3: 0x44a9, Data4: [8]byte{0xb0, 0x3c, 0x77, 0x90, 0x0a, 0x0a, 0x1d, 0x17}}
)

// dxgiUnknownVtbl 是 IUnknown vtable 前三个槽位。
type dxgiUnknownVtbl struct {
	queryInterface uintptr
	addRef         uintptr
	release        uintptr
}

// dxgiUnknownVtable 读取 COM 对象首字段指向的 vtable。
// [S-inline] ASM 内联 `mov rdx,[rax]; mov rax,[rdx]` 模式。
func dxgiUnknownVtable(p uintptr) *dxgiUnknownVtbl {
	return *(**dxgiUnknownVtbl)(unsafe.Pointer(&p))
}

// createDXGIFactory1 创建 IDXGIFactory1。
// [S] ASM 0x14085da00: Find 失败返回 err；SyscallN(CreateDXGIFactory1, &iid, &factory)；
// HRESULT 非零 → "%s failed: HRESULT 0x%08X"；factory 零 → errors.New。
func createDXGIFactory1() (uintptr, error) {
	if err := procCreateDXGIFactory1.Find(); err != nil {
		return 0, err
	}
	var factory uintptr
	hresult, _, _ := syscall.SyscallN(
		procCreateDXGIFactory1.Addr(),
		uintptr(unsafe.Pointer(&iidIDXGIFactory1)),
		uintptr(unsafe.Pointer(&factory)),
	)
	if hresult != 0 {
		return 0, fmt.Errorf("%s failed: HRESULT 0x%08X", "CreateDXGIFactory1", uint32(hresult))
	}
	if factory == 0 {
		return 0, errors.New("CreateDXGIFactory1 did not return a factory")
	}
	return factory, nil
}

// queryDXGIFactory6 通过 QueryInterface 查询 IDXGIFactory6。
// [S] ASM 0x14085db80: factory 零 → "DXGI factory is nil"；vtable[0] QueryInterface；
// HRESULT 非零 → "%s failed: HRESULT 0x%08X"("IDXGIFactory6 unavailable")；
// factory6 零 → "IDXGIFactory6 query returned nil"。
func queryDXGIFactory6(factory uintptr) (uintptr, error) {
	if factory == 0 {
		return 0, errors.New("DXGI factory is nil")
	}
	var factory6 uintptr
	hresult, _, _ := syscall.SyscallN(
		dxgiUnknownVtable(factory).queryInterface,
		factory,
		uintptr(unsafe.Pointer(&iidIDXGIFactory6)),
		uintptr(unsafe.Pointer(&factory6)),
	)
	if hresult != 0 {
		return 0, fmt.Errorf("%s failed: HRESULT 0x%08X", "IDXGIFactory6 unavailable", uint32(hresult))
	}
	if factory6 == 0 {
		return 0, errors.New("IDXGIFactory6 query returned nil")
	}
	return factory6, nil
}

// releaseDXGIUnknown 释放 DXGI COM 对象（vtable[2] Release）。
// [S] ASM 0x14085dd20: p/vtable/release 任一零 → 返回；否则 SyscallN(release, p)。
func releaseDXGIUnknown(p uintptr) {
	if p == 0 {
		return
	}
	vtable := dxgiUnknownVtable(p)
	if vtable == nil || vtable.release == 0 {
		return
	}
	syscall.SyscallN(vtable.release, p)
}

// screenshotCOMRelease 释放 COM 对象（vtable[2] Release）。
// [S] ASM 0x14097df80: p/vtable/release 任一零 → 返回；否则 SyscallN(release, p)。
func screenshotCOMRelease(p uintptr) {
	if p == 0 {
		return
	}
	vtable := dxgiUnknownVtable(p)
	if vtable == nil || vtable.release == 0 {
		return
	}
	syscall.SyscallN(vtable.release, p)
}

// screenshotCOMQueryInterface 通过 vtable[0] QueryInterface 查询接口。
// [S] ASM 0x14097e000: p 零或 iid 零 → "COM 对象或 IID 为空"；qi=vtable[0]；
// SyscallN(qi, p, iid, &result)；HRESULT 非零 → "%s failed: HRESULT 0x%08X"
// ("COM QueryInterface")；result 零 → "COM QueryInterface 结果为空"。
func screenshotCOMQueryInterface(p uintptr, iid *windows.GUID) (uintptr, error) {
	if p == 0 || iid == nil {
		return 0, errors.New("COM 对象或 IID 为空")
	}
	var result uintptr
	qi := uintptr(0)
	if vtable := dxgiUnknownVtable(p); vtable != nil {
		qi = vtable.queryInterface
	}
	hresult, _, _ := syscall.SyscallN(qi, p, uintptr(unsafe.Pointer(iid)), uintptr(unsafe.Pointer(&result)))
	if hresult != 0 {
		return 0, fmt.Errorf("%s failed: HRESULT 0x%08X", "COM QueryInterface", uint32(hresult))
	}
	if result == 0 {
		return 0, errors.New("COM QueryInterface 结果为空")
	}
	return result, nil
}
