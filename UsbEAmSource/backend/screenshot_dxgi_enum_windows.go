package main

import (
	"errors"
	"fmt"
	"image"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// iidIDXGIOutput6 用于 IDXGIOutput → IDXGIOutput6 的 QueryInterface。
var iidIDXGIOutput6 = windows.GUID{Data1: 0x068346e8, Data2: 0xaaec, Data3: 0x4b84, Data4: [8]byte{0xad, 0xd7, 0x13, 0x7f, 0x51, 0x3f, 0x77, 0xa1}}

// dxgiAdapterDesc 是 DXGI_ADAPTER_DESC（GetDesc 输出）。
type dxgiAdapterDesc struct {
	Description           [128]uint16
	VendorID              uint32
	DeviceID              uint32
	SubSysID              uint32
	Revision              uint32
	DedicatedVideoMemory  uintptr
	DedicatedSystemMemory uintptr
	SharedSystemMemory    uintptr
	AdapterLuidLow        uint32
	AdapterLuidHigh       uint32
}

// dxgiOutputDesc1 是 DXGI_OUTPUT_DESC1（IDXGIOutput6.GetDesc1 输出，148 字节）。
type dxgiOutputDesc1 struct {
	DeviceName            [32]uint16
	DesktopLeft           int32
	DesktopTop            int32
	DesktopRight          int32
	DesktopBottom         int32
	AttachedToDesktop     int32
	Rotation              uint32
	Monitor               uintptr
	BitsPerColor          uint32
	ColorSpace            uint32
	RedPrimary            [2]float32
	GreenPrimary          [2]float32
	BluePrimary           [2]float32
	WhitePoint            [2]float32
	MinLuminance          float32
	MaxLuminance          float32
	MaxFullFrameLuminance float32
}

// dxgiFactoryVtbl 覆盖 IDXGIFactory 前 8 个槽（EnumAdapters）。
type dxgiFactoryVtbl struct {
	QueryInterface          uintptr // 0x00
	AddRef                  uintptr // 0x08
	Release                 uintptr // 0x10
	SetPrivateData          uintptr // 0x18
	SetPrivateDataInterface uintptr // 0x20
	GetPrivateData          uintptr // 0x28
	GetParent               uintptr // 0x30
	EnumAdapters            uintptr // 0x38
}

// dxgiAdapterVtable 读取 IDXGIAdapter 对象首字段指向的 vtable。
// [S-inline] ASM 内联 `mov rdx,[rax]; mov rax,[rdx]` 模式。
func dxgiAdapterVtable(p uintptr) *dxgiAdapterVtbl {
	return *(**dxgiAdapterVtbl)(unsafe.Pointer(&p))
}

// dxgiFactoryVtable 读取 IDXGIFactory 对象首字段指向的 vtable。
// [S-inline] ASM 内联 `mov rdx,[rax]; mov rax,[rdx]` 模式。
func dxgiFactoryVtable(p uintptr) *dxgiFactoryVtbl {
	return *(**dxgiFactoryVtbl)(unsafe.Pointer(&p))
}

// dxgiOutput6Vtable 读取 IDXGIOutput6 对象首字段指向的 vtable。
// [S-inline] ASM 内联 `mov rdx,[rax]; mov rax,[rdx]` 模式。
func dxgiOutput6Vtable(p uintptr) *dxgiOutput6Vtbl {
	return *(**dxgiOutput6Vtbl)(unsafe.Pointer(&p))
}

// screenshotDXGIAdapterName 读取适配器描述名称。
// [S] ASM 0x14097eba0: adapter 零 → ("",nil)；vtable[8] GetDesc；
// HRESULT 非零 → "%s failed: HRESULT 0x%08X"；否则 TrimSpace(UTF16ToString(Description))。
func screenshotDXGIAdapterName(adapter uintptr) (string, error) {
	if adapter == 0 {
		return "", nil
	}
	var desc dxgiAdapterDesc
	hresult, _, _ := syscall.SyscallN(
		dxgiAdapterVtable(adapter).GetDesc,
		adapter,
		uintptr(unsafe.Pointer(&desc)),
	)
	if hresult != 0 {
		return "", fmt.Errorf("%s failed: HRESULT 0x%08X", "IDXGIAdapter.GetDesc", uint32(hresult))
	}
	return strings.TrimSpace(syscall.UTF16ToString(desc.Description[:])), nil
}

// queryScreenshotDXGIOutput6 通过 QueryInterface 查询 IDXGIOutput6。
// [S] ASM 0x14097f180: output 零 → "DXGI 输出为空"；vtable[0] QueryInterface；
// HRESULT 非零 → "%s failed: HRESULT 0x%08X"("IDXGIOutput6 unavailable")；
// output6 零 → "IDXGIOutput6 查询结果为空"。
func queryScreenshotDXGIOutput6(output uintptr) (uintptr, error) {
	if output == 0 {
		return 0, errors.New("DXGI 输出为空")
	}
	var output6 uintptr
	hresult, _, _ := syscall.SyscallN(
		dxgiUnknownVtable(output).queryInterface,
		output,
		uintptr(unsafe.Pointer(&iidIDXGIOutput6)),
		uintptr(unsafe.Pointer(&output6)),
	)
	if hresult != 0 {
		return 0, fmt.Errorf("%s failed: HRESULT 0x%08X", "IDXGIOutput6 unavailable", uint32(hresult))
	}
	if output6 == 0 {
		return 0, errors.New("IDXGIOutput6 查询结果为空")
	}
	return output6, nil
}

// screenshotDisplayCaptureInfoFromDesc 从 DXGI_OUTPUT_DESC1 构造捕获信息。
// [S] ASM 0x14097f320: AdapterName/DeviceName TrimSpace；Bounds 归一化；
// HDR = ColorSpace∈{12,13,14,16,18,19}；AdvancedColor = ColorSpace!=0 || BitsPerColor>8。
func screenshotDisplayCaptureInfoFromDesc(desc dxgiOutputDesc1, adapterIndex int, outputIndex int, adapterName string) screenshotDisplayCaptureInfo {
	adapterName = strings.TrimSpace(adapterName)
	deviceName := strings.TrimSpace(syscall.UTF16ToString(desc.DeviceName[:]))

	minX := int(desc.DesktopLeft)
	maxX := int(desc.DesktopRight)
	if maxX < minX {
		minX, maxX = maxX, minX
	}
	minY := int(desc.DesktopTop)
	maxY := int(desc.DesktopBottom)
	if maxY < minY {
		minY, maxY = maxY, minY
	}

	colorSpace := desc.ColorSpace
	hdr := colorSpace == 12 || colorSpace == 13 || colorSpace == 14 || colorSpace == 16 || colorSpace == 18 || colorSpace == 19
	advancedColor := colorSpace != 0 || desc.BitsPerColor > 8

	return screenshotDisplayCaptureInfo{
		AdapterIndex:          adapterIndex,
		OutputIndex:           outputIndex,
		AdapterName:           adapterName,
		DeviceName:            deviceName,
		Bounds:                image.Rect(minX, minY, maxX, maxY),
		Rotation:              desc.Rotation,
		AttachedToDesktop:     desc.AttachedToDesktop != 0,
		BitsPerColor:          desc.BitsPerColor,
		ColorSpace:            colorSpace,
		HDR:                   hdr,
		AdvancedColor:         advancedColor,
		MinLuminance:          desc.MinLuminance,
		MaxLuminance:          desc.MaxLuminance,
		MaxFullFrameLuminance: desc.MaxFullFrameLuminance,
	}
}

// screenshotDXGIOutputDisplayInfo 查询输出 IDXGIOutput6 并读取 GetDesc1。
// [S] ASM 0x14097ed40: queryScreenshotDXGIOutput6 → vtable[27] GetDesc1 →
// screenshotDisplayCaptureInfoFromDesc。
func screenshotDXGIOutputDisplayInfo(output uintptr, adapterIndex int, outputIndex int, adapterName string) (screenshotDisplayCaptureInfo, error) {
	output6, err := queryScreenshotDXGIOutput6(output)
	if err != nil {
		return screenshotDisplayCaptureInfo{}, err
	}
	defer releaseDXGIUnknown(output6)

	var desc dxgiOutputDesc1
	hresult, _, _ := syscall.SyscallN(
		dxgiOutput6Vtable(output6).GetDesc1,
		output6,
		uintptr(unsafe.Pointer(&desc)),
	)
	if hresult != 0 {
		return screenshotDisplayCaptureInfo{}, fmt.Errorf("%s failed: HRESULT 0x%08X", "IDXGIOutput6.GetDesc1", uint32(hresult))
	}
	return screenshotDisplayCaptureInfoFromDesc(desc, adapterIndex, outputIndex, adapterName), nil
}

// enumerateScreenshotDXGIAdapterDisplays 枚举适配器全部输出。
// [S] ASM 0x14097e7c0: make(cap 2)；循环 vtable[7] EnumOutputs；
// NOT_FOUND → (infos,nil)；其他 HRESULT → 错误；output 零 → continue；
// 逐输出 screenshotDXGIOutputDisplayInfo 并 append。
func enumerateScreenshotDXGIAdapterDisplays(adapter uintptr, adapterIndex int) ([]screenshotDisplayCaptureInfo, error) {
	name, err := screenshotDXGIAdapterName(adapter)
	if err != nil {
		return nil, err
	}
	infos := make([]screenshotDisplayCaptureInfo, 0, 2)
	for i := 0; ; i++ {
		var output uintptr
		hresult, _, _ := syscall.SyscallN(
			dxgiAdapterVtable(adapter).EnumOutputs,
			adapter,
			uintptr(i),
			uintptr(unsafe.Pointer(&output)),
		)
		if hresult != 0 {
			if hresult == 0x887a0002 { // DXGI_ERROR_NOT_FOUND
				return infos, nil
			}
			return nil, fmt.Errorf("%s failed: HRESULT 0x%08X", "IDXGIAdapter.EnumOutputs", uint32(hresult))
		}
		if output == 0 {
			continue
		}
		info, err := screenshotDXGIOutputDisplayInfo(output, adapterIndex, i, name)
		releaseDXGIUnknown(output)
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}
}

// enumerateScreenshotDisplayCaptureInfos 枚举全部适配器与输出。
// [S] ASM 0x14097e1c0: createDXGIFactory1 → queryDXGIFactory6 → make(cap 4)；
// 循环 vtable[7] EnumAdapters；逐适配器 enumerateScreenshotDXGIAdapterDisplays 并 append。
func enumerateScreenshotDisplayCaptureInfos() ([]screenshotDisplayCaptureInfo, error) {
	factory1, err := createDXGIFactory1()
	if err != nil {
		return nil, err
	}
	defer releaseDXGIUnknown(factory1)

	factory6, err := queryDXGIFactory6(factory1)
	if err != nil {
		return nil, err
	}
	defer releaseDXGIUnknown(factory6)

	infos := make([]screenshotDisplayCaptureInfo, 0, 4)
	for i := 0; ; i++ {
		var adapter uintptr
		hresult, _, _ := syscall.SyscallN(
			dxgiFactoryVtable(factory6).EnumAdapters,
			factory6,
			uintptr(i),
			uintptr(unsafe.Pointer(&adapter)),
		)
		if hresult != 0 {
			if hresult == 0x887a0002 { // DXGI_ERROR_NOT_FOUND
				return infos, nil
			}
			return nil, fmt.Errorf("%s failed: HRESULT 0x%08X", "IDXGIFactory.EnumAdapters", uint32(hresult))
		}
		if adapter == 0 {
			continue
		}
		displays, err := enumerateScreenshotDXGIAdapterDisplays(adapter, i)
		releaseDXGIUnknown(adapter)
		if err != nil {
			return nil, err
		}
		infos = append(infos, displays...)
	}
}
