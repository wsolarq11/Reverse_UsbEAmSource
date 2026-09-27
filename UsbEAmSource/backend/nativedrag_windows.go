// AUTO-RECONSTRUCTED — DOMAIN: nativedrag
// Source: UsbEAm_Launcher 1.0.3 (Go 1.25.12, PE64), disassembled from
// main.nativeFileDrag* / main.(*nativeFileDrag*).* symbols. Tier markers:
//
//	[S VA]      — body fully translated from asm (trivial HRESULT returns / full
//	              QueryGetData FORMATETC validation).
//	[S-sig VA]  — signature proven from asm + ifce defs (types_windows.go);
//	              body is a faithful zero skeleton (Win32 / combridge deps).
//	[P]         — 签名待实证：register/stack layout is ambiguous; zero skeleton only.
//
// 研究用途
package main

import (
	"github.com/wailsapp/go-webview2/pkg/combridge"
	"github.com/wailsapp/wails/v3/pkg/w32"
)

// combridge IDataObject trampolines — dispatch onto nativeFileDragDataObjectCom.
// [S-sig 0x1408f9d60] GetData trampoline.
func nativeFileDragGetData(obj *combridge.ComObject[nativeFileDragDataObjectCom], pformatetc *w32.FORMATETC, pmedium *w32.STGMEDIUM) uintptr {
	return 0
}

// [S-sig 0x1408f9de0] GetDataHere trampoline.
func nativeFileDragGetDataHere(obj *combridge.ComObject[nativeFileDragDataObjectCom], pformatetc *w32.FORMATETC, pmedium *w32.STGMEDIUM) uintptr {
	return 0
}

// [S-sig 0x1408f9e60] QueryGetData trampoline.
func nativeFileDragQueryGetData(obj *combridge.ComObject[nativeFileDragDataObjectCom], pformatetc *w32.FORMATETC) uintptr {
	return 0
}

// [S-sig 0x1408f9ec0] GetCanonicalFormatEtc trampoline.
func nativeFileDragGetCanonicalFormatEtc(obj *combridge.ComObject[nativeFileDragDataObjectCom], pformetcIn, pformetcOut *w32.FORMATETC) uintptr {
	return 0
}

// [S-sig 0x1408f9f40] SetData trampoline.
func nativeFileDragSetData(obj *combridge.ComObject[nativeFileDragDataObjectCom], pformatetc *w32.FORMATETC, pmedium *w32.STGMEDIUM, fRelease int32) uintptr {
	return 0
}

// [S-sig 0x1408f9fc0] EnumFormatEtc trampoline.
func nativeFileDragEnumFormatEtc(obj *combridge.ComObject[nativeFileDragDataObjectCom], dwDirection uint32, ppenumFormatEtc **w32.IEnumFORMATETC) uintptr {
	return 0
}

// [S-sig 0x1408fa040] DAdvise trampoline.
func nativeFileDragDAdvise(obj *combridge.ComObject[nativeFileDragDataObjectCom], pformatetc *w32.FORMATETC, advf uint32, pAdvSink *w32.IAdviseSink, pdwConnection *uint32) uintptr {
	return 0
}

// [S-sig 0x1408fa0e0] DUnadvise trampoline.
func nativeFileDragDUnadvise(obj *combridge.ComObject[nativeFileDragDataObjectCom], dwConnection uint32) uintptr {
	return 0
}

// [S-sig 0x1408fa140] EnumDAdvise trampoline.
func nativeFileDragEnumDAdvise(obj *combridge.ComObject[nativeFileDragDataObjectCom], ppenumAdvise **w32.IEnumStatData) uintptr {
	return 0
}

// combridge IEnumFORMATETC trampolines — dispatch onto nativeFileDragEnumFormatEtcCom.
// [S-sig 0x1408fa1a0] Next trampoline.
func nativeFileDragEnumFormatEtcNext(obj *combridge.ComObject[nativeFileDragEnumFormatEtcCom], celt uint32, rgelt *w32.FORMATETC, pceltFetched *uint32) uintptr {
	return 0
}

// [S-sig 0x1408fa220] Skip trampoline.
func nativeFileDragEnumFormatEtcSkip(obj *combridge.ComObject[nativeFileDragEnumFormatEtcCom], celt uint32) uintptr {
	return 0
}

// [S-sig 0x1408fa280] Reset trampoline.
func nativeFileDragEnumFormatEtcReset(obj *combridge.ComObject[nativeFileDragEnumFormatEtcCom]) uintptr {
	return 0
}

// [S-sig 0x1408fa2c0] Clone trampoline.
func nativeFileDragEnumFormatEtcClone(obj *combridge.ComObject[nativeFileDragEnumFormatEtcCom], ppenum **w32.IEnumFORMATETC) uintptr {
	return 0
}

// *nativeFileDragDataObject — IDataObject implementation.
// [S-sig 0x1408fa320] GetData: 依据 CfFormat 构造 STGMEDIUM（CF_HDROP / DWORD 偏好格式）。
func (d *nativeFileDragDataObject) GetData(pformatetc *w32.FORMATETC, pmedium *w32.STGMEDIUM) uintptr {
	return 0
}

// [S 0x1408fa480] GetDataHere: 无条件返回 E_NOTIMPL。
func (d *nativeFileDragDataObject) GetDataHere(pformatetc *w32.FORMATETC, pmedium *w32.STGMEDIUM) uintptr {
	return 0x80004001 // E_NOTIMPL
}

// [S 0x1408fa4a0] QueryGetData: 逐项校验 FORMATETC（nil / DwAspect / Tymed / CfFormat）。
func (d *nativeFileDragDataObject) QueryGetData(pformatetc *w32.FORMATETC) uintptr {
	if pformatetc == nil {
		return 0x80070057 // E_INVALIDARG
	}
	if pformatetc.DwAspect != 0 && pformatetc.DwAspect != 1 {
		return 0x8004006b // DV_E_DVASPECT
	}
	if pformatetc.Tymed != 0 && pformatetc.Tymed&1 == 0 {
		return 0x80040069 // DV_E_TYMED (TYMED_HGLOBAL == 1)
	}
	if pformatetc.CfFormat == 0x000f || pformatetc.CfFormat == d.preferredDropEffectFormat {
		return 0 // S_OK
	}
	return 0x80040064 // DV_E_FORMATETC
}

// [S-sig 0x1408fa500] GetCanonicalFormatEtc: 输出清零 FORMATETC（Ptd=nil, Lindex=-1）。
func (d *nativeFileDragDataObject) GetCanonicalFormatEtc(pformetcIn, pformetcOut *w32.FORMATETC) uintptr {
	return 0
}

// [S 0x1408fa5a0] SetData: 无条件返回 E_NOTIMPL。
func (d *nativeFileDragDataObject) SetData(pformatetc *w32.FORMATETC, pmedium *w32.STGMEDIUM, fRelease int32) uintptr {
	return 0x80004001 // E_NOTIMPL
}

// [S-sig 0x1408fa5c0] EnumFormatEtc: 创建 *nativeFileDragFormatEnumerator 并回填 ppenumFormatEtc。
func (d *nativeFileDragDataObject) EnumFormatEtc(dwDirection uint32, ppenumFormatEtc **w32.IEnumFORMATETC) uintptr {
	return 0
}

// [S 0x1408fa680] DAdvise: 清零 *pdwConnection，返回 OLE_E_ADVISENOTSUPPORTED。
func (d *nativeFileDragDataObject) DAdvise(pformatetc *w32.FORMATETC, advf uint32, pAdvSink *w32.IAdviseSink, pdwConnection *uint32) uintptr {
	if pdwConnection != nil {
		*pdwConnection = 0
	}
	return 0x80040003 // OLE_E_ADVISENOTSUPPORTED
}

// [S 0x1408fa6a0] DUnadvise: 无条件返回 OLE_E_ADVISENOTSUPPORTED。
func (d *nativeFileDragDataObject) DUnadvise(dwConnection uint32) uintptr {
	return 0x80040003 // OLE_E_ADVISENOTSUPPORTED
}

// [S 0x1408fa6c0] EnumDAdvise: 清零 *ppenumAdvise，返回 OLE_E_ADVISENOTSUPPORTED。
func (d *nativeFileDragDataObject) EnumDAdvise(ppenumAdvise **w32.IEnumStatData) uintptr {
	if ppenumAdvise != nil {
		*ppenumAdvise = nil
	}
	return 0x80040003 // OLE_E_ADVISENOTSUPPORTED
}

// *nativeFileDragFormatEnumerator — IEnumFORMATETC implementation.
// [S-sig 0x1408fa720] Next: 按 index 复制 celt 个 FORMATETC 到 rgelt，回填 pceltFetched。
func (e *nativeFileDragFormatEnumerator) Next(celt uint32, rgelt *w32.FORMATETC, pceltFetched *uint32) uintptr {
	return 0
}

// [S 0x1408fa8c0] Skip: 前进 celt 个元素；越界时置 index=len 并返回 S_FALSE。
func (e *nativeFileDragFormatEnumerator) Skip(celt uint32) uintptr {
	if celt == 0 {
		return 0 // S_OK
	}
	remaining := len(e.formats) - e.index
	if remaining < 0 {
		remaining = 0
	}
	if int64(celt) > int64(remaining) {
		e.index = len(e.formats)
		return 1 // S_FALSE
	}
	e.index += int(celt)
	return 0 // S_OK
}

// [S 0x1408fa900] Reset: index 归零，返回 S_OK。
func (e *nativeFileDragFormatEnumerator) Reset() uintptr {
	e.index = 0
	return 0 // S_OK
}

// [S-sig 0x1408fa920] Clone: 基于当前 formats/index 复制出新的枚举器并回填 ppenum。
func (e *nativeFileDragFormatEnumerator) Clone(ppenum **w32.IEnumFORMATETC) uintptr {
	return 0
}

// *BootstrapService — 启动原生文件拖拽入口。
// [S-sig 0x1408fa9a0] StartNativeFileDrag: 规范化路径后启动 IDropSource 风格拖拽。
func (b *BootstrapService) StartNativeFileDrag(paths []string) error {
	return nil
}

// [S-sig 0x1408fab80] StartAppEntryNativeDrag: 从 AppEntry 解析路径并复用 StartNativeFileDrag。
func (b *BootstrapService) StartAppEntryNativeDrag(entry AppEntry) error {
	return nil
}

// [S-sig 0x1408faea0] normalizeNativeFileDragPaths: 校验/规整路径列表（存在性、绝对化）。
func normalizeNativeFileDragPaths(paths []string) ([]string, error) {
	return nil, nil
}

// [P] 签名待实证：大帧 + 序言仅露出 [rsp+0x1d8]/[rsp+0x1e0] 两个栈槽且为大结构返回，
// 形参与返回契约无法从序言唯一确定；返回 (shortcutPath, error) 为推断。
func createTemporaryDirectoryShortcut(dirPath, name string) (string, error) {
	return "", nil
}

// [S-sig 0x1408fb580] 净化目录快捷方式名。
// 序言 morestack 保存 rax/rbx/rcx/rdi = 4 字 = name(rax,rbx) + path(rcx,rdi) 两 string；
// 体：TrimSpace(name) 空则 inferName(path)；Replacer 替换非法字符 + Trim；空则返回默认名(18B)。
// 返回 rax/rbx = string。旧签名漏 path 参数，已订正。
func sanitizeDirectoryShortcutName(name string, path string) string {
	return ""
}

// [S-sig 0x1408fb700] startNativeFileDrag: 解析偏好 drop 效果格式并启动拖拽循环。
func startNativeFileDrag(paths []string) error {
	return nil
}

// [S-sig 0x1408fbac0] ensureNativeFileDragOleInitialised: 惰性调用 OleInitialize。
func ensureNativeFileDragOleInitialised() {
}

// [S-sig 0x1408fbb00] resolveNativeFileDragPreferredEffectFormat: 注册/解析偏好 drop 效果剪贴板格式。
func resolveNativeFileDragPreferredEffectFormat() uint16 {
	return 0
}

// [S-sig 0x1408fbb60] registerNativeFileDragClipboardFormat: 注册命名剪贴板格式。
func registerNativeFileDragClipboardFormat(name string) uint16 {
	return 0
}

// [S-sig 0x1408fbcc0] createNativeFileDragEnumFormatEtcAtIndex: 以指定 index 构建枚举器。
func createNativeFileDragEnumFormatEtcAtIndex(formats []w32.FORMATETC, index int) *nativeFileDragFormatEnumerator {
	return nil
}

// [S-sig 0x1408fbe60] buildNativeFileDragHDropHandle: 构造 DROPFILES（HDROP）全局内存句柄。
func buildNativeFileDragHDropHandle(paths []string) uintptr {
	return 0
}

// [S-sig 0x1408fc1c0] buildNativeFileDragDWORDHandle: 构造含单个 DWORD 的全局内存句柄。
func buildNativeFileDragDWORDHandle(value uint32) uintptr {
	return 0
}

// [S-sig 0x1408fc440] buildNativeFileDragWidePathList: 拼接以双 NUL 结尾的宽字符路径列表。
func buildNativeFileDragWidePathList(paths []string) []uint16 {
	return nil
}
