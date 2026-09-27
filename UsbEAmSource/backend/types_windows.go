// AUTO-RECONSTRUCTED TYPES — DOMAIN: windows
// 研究用途
package main

import (
	"github.com/wailsapp/go-webview2/pkg/combridge"
	"github.com/wailsapp/wails/v3/pkg/w32"
	"golang.org/x/sys/windows"
)

type displayconfigDeviceInfoHeader struct {
	Type      uint32
	Size      uint32
	AdapterID windows.LUID
	ID        uint32
}

type displayconfigModeInfo struct {
	InfoType  uint32
	ID        uint32
	AdapterID windows.LUID
	Info      [6]uint64
}

type displayconfigPathInfo struct {
	SourceInfo displayconfigPathSourceInfo
	TargetInfo displayconfigPathTargetInfo
	Flags      uint32
}

type displayconfigPathSourceInfo struct {
	AdapterID   windows.LUID
	ID          uint32
	ModeInfoIdx uint32
	StatusFlags uint32
}

type displayconfigPathTargetInfo struct {
	AdapterID        windows.LUID
	ID               uint32
	ModeInfoIdx      uint32
	OutputTechnology uint32
	Rotation         uint32
	Scaling          uint32
	RefreshRate      displayconfigRational
	ScanLineOrdering uint32
	TargetAvailable  int32
	StatusFlags      uint32
}

type displayconfigRational struct {
	Numerator   uint32
	Denominator uint32
}

type displayconfigSDRWhiteLevel struct {
	Header        displayconfigDeviceInfoHeader
	SDRWhiteLevel uint32
}

type displayconfigSourceDeviceName struct {
	Header            displayconfigDeviceInfoHeader
	ViewGDIDeviceName [32]uint16
}

type dxgiAdapter struct {
	Vtbl *dxgiAdapterVtbl
}

type dxgiAdapterVtbl struct {
	QueryInterface          uintptr
	AddRef                  uintptr
	Release                 uintptr
	SetPrivateData          uintptr
	SetPrivateDataInterface uintptr
	GetPrivateData          uintptr
	GetParent               uintptr
	EnumOutputs             uintptr
	GetDesc                 uintptr
	CheckInterfaceSupport   uintptr
}

type dxgiIUnknown struct {
	Vtbl *dxgiIUnknownVtbl
}

type dxgiIUnknownVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
}

type dxgiOutput6 struct {
	Vtbl *dxgiOutput6Vtbl
}

type dxgiOutput6Vtbl struct {
	QueryInterface                uintptr
	AddRef                        uintptr
	Release                       uintptr
	SetPrivateData                uintptr
	SetPrivateDataInterface       uintptr
	GetPrivateData                uintptr
	GetParent                     uintptr
	GetDesc                       uintptr
	GetDisplayModeList            uintptr
	FindClosestMatchingMode       uintptr
	WaitForVBlank                 uintptr
	TakeOwnership                 uintptr
	ReleaseOwnership              uintptr
	GetGammaControlCapabilities   uintptr
	SetGammaControl               uintptr
	GetGammaControl               uintptr
	SetDisplaySurface             uintptr
	GetDisplaySurfaceData         uintptr
	GetFrameStatistics            uintptr
	GetDisplayModeList1           uintptr
	FindClosestMatchingMode1      uintptr
	GetDisplaySurfaceData1        uintptr
	DuplicateOutput               uintptr
	SupportsOverlays              uintptr
	CheckOverlaySupport           uintptr
	CheckOverlayColorSpaceSupport uintptr
	DuplicateOutput1              uintptr
	GetDesc1                      uintptr
	CheckHardwareComposition      uintptr
}

// dxgiFactory6Vtbl 覆盖 IDXGIFactory 全链至 IDXGIFactory6（EnumAdapterByGpuPreference）。
type dxgiFactory6Vtbl struct {
	QueryInterface                uintptr // 0x00
	AddRef                        uintptr // 0x08
	Release                       uintptr // 0x10
	SetPrivateData                uintptr // 0x18
	SetPrivateDataInterface       uintptr // 0x20
	GetPrivateData                uintptr // 0x28
	GetParent                     uintptr // 0x30
	EnumAdapters                  uintptr // 0x38
	MakeWindowAssociation         uintptr // 0x40
	GetWindowAssociation          uintptr // 0x48
	CreateSwapChain               uintptr // 0x50
	CreateSoftwareAdapter         uintptr // 0x58
	EnumAdapters1                 uintptr // 0x60
	IsCurrent                     uintptr // 0x68
	IsWindowedStereoEnabled       uintptr // 0x70
	CreateSwapChainForHwnd        uintptr // 0x78
	CreateSwapChainForCoreWindow  uintptr // 0x80
	GetSharedResourceAdapterLuid  uintptr // 0x88
	RegisterStereoStatusWindow    uintptr // 0x90
	RegisterStereoStatusEvent     uintptr // 0x98
	UnregisterStereoStatus        uintptr // 0xa0
	RegisterOcclusionStatusWindow uintptr // 0xa8
	RegisterOcclusionStatusEvent  uintptr // 0xb0
	UnregisterOcclusionStatus     uintptr // 0xb8
	CreateSwapChainForComposition uintptr // 0xc0
	GetCreationFlags              uintptr // 0xc8
	EnumAdapterByLuid             uintptr // 0xd0
	EnumWarpAdapter               uintptr // 0xd8
	CheckFeatureSupport           uintptr // 0xe0
	EnumAdapterByGpuPreference    uintptr // 0xe8
}

type shFileInfo struct {
	HIcon         uintptr
	IIcon         int32
	DwAttributes  uint32
	SzDisplayName [260]uint16
	SzTypeName    [80]uint16
}

type nativeFileDragDataObject struct {
	combridge.IUnknownImpl
	paths                     []string
	preferredDropEffectFormat uint16
	formats                   []w32.FORMATETC
}

type nativeFileDragDataObjectCom interface {
	DAdvise(*w32.FORMATETC, uint32, *w32.IAdviseSink, *uint32) uintptr
	DUnadvise(uint32) uintptr
	EnumDAdvise(**w32.IEnumStatData) uintptr
	EnumFormatEtc(uint32, **w32.IEnumFORMATETC) uintptr
	GetCanonicalFormatEtc(*w32.FORMATETC, *w32.FORMATETC) uintptr
	GetData(*w32.FORMATETC, *w32.STGMEDIUM) uintptr
	GetDataHere(*w32.FORMATETC, *w32.STGMEDIUM) uintptr
	QueryGetData(*w32.FORMATETC) uintptr
	SetData(*w32.FORMATETC, *w32.STGMEDIUM, int32) uintptr
}

type nativeFileDragEnumFormatEtcCom interface {
	Clone(**w32.IEnumFORMATETC) uintptr
	Next(uint32, *w32.FORMATETC, *uint32) uintptr
	Reset() uintptr
	Skip(uint32) uintptr
}

type nativeFileDragFormatEnumerator struct {
	combridge.IUnknownImpl
	formats []w32.FORMATETC
	index   int
}

type shellExecuteInfo struct {
	cbSize       uint32
	fMask        uint32
	hwnd         uintptr
	lpVerb       *uint16
	lpFile       *uint16
	lpParameters *uint16
	lpDirectory  *uint16
	nShow        int32
	hInstApp     uintptr
	lpIDList     uintptr
	lpClass      *uint16
	hkeyClass    uintptr
	dwHotKey     uint32
	hIcon        uintptr
	hProcess     uintptr
}
