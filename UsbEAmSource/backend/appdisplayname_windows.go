// AUTO-RECONSTRUCTED SERVICE METHODS — DOMAIN: app display name (Win32 Version API)
// 研究用途
//
// 契约来源：
//   - 函数签名：redress types all -m -v
//   - symbols.main.bak：10 函数 VA 完整
//   - 反汇编：../work/disasm/dump/*.asm.txt（capstone 逐条标注）
//   - 行号蓝图：source_funcs.txt File: appdisplayname.go + appdisplayname_windows.go
//
// 档位：
//
//	[S] 汇编实证：resolveAppDisplayName(0x140748280)、resolveFileDisplayName(0x1407486c0)
//	    fallbackAppDisplayName(0x140748040)、queryVersionField(0x1407487a0)
//	    loadVersionInfo(0x140748a80)、queryVersionTranslations(0x140748c00)
//	    decodeVersionTranslations(0x140748ca0)、queryVersionString(0x140748f00)
//	    decodeVersionUTF16String(0x1407490a0)、queryVersionValueLocation(0x140749220)
//
// 验证：全 10 函数反汇编逐条追译
package main

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ========================================================================
// resolveAppDisplayName — 快捷方式解析 → file display name
// ========================================================================

// resolveAppDisplayName 解析文件/快捷方式的显示名。
// [S 汇编实证 0x140748280, 1088B]
// 返回 (name string, found bool)
func resolveAppDisplayName(path string) (string, bool) {
	p := strings.TrimSpace(path)
	if p == "" {
		return "", false
	}
	if !isShortcutFilePath(p) {
		return p, true
	}

	info, err := resolveShortcutInfoWithIconResolver(p)
	if err != nil {
		return p, true
	}

	display := strings.TrimSpace(info.DisplayName)
	if display == "" {
		return p, true
	}

	resolved := strings.TrimSpace(display)
	if resolved != "" {
		return resolved, true
	}

	// fallback
	fb := fallbackAppDisplayName(p)
	return fb, true
}

// ========================================================================
// resolveFileDisplayName — 从 Win32 Version 资源读取
// ========================================================================

// resolveFileDisplayName 从版本资源查询文件显示名（FileDescription / ProductName）。
// [S 汇编实证 0x1407486c0, 224B]
func resolveFileDisplayName(path string) (string, bool) {
	fields := []struct {
		name string
		len  int
	}{
		{"FileDescription", 15},
		{"ProductName", 11},
	}

	for _, f := range fields {
		result, found := queryVersionField(path, f.name)
		if found && result != "" {
			return result, true
		}
	}
	return "", false
}

// ========================================================================
// queryVersionField — 遍历所有翻译查询指定版本字段
// ========================================================================

// queryVersionField 遍历所有语言/代码页组合，查询 version 资源的指定字段。
// [S 汇编实证 0x1407487a0, 736B]
func queryVersionField(path string, fieldName string) (string, bool) {
	data, size, err := loadVersionInfo(path)
	if err != nil || data == nil {
		return "", false
	}

	// queryVersionTranslations → 解析翻译对
	translations, err := queryVersionTranslations(data, size)
	if err != nil {
		return "", false
	}

	// 遍历 translation 对，map 去重
	seen := make(map[string]struct{})
	for _, t := range translations {
		key := fmt.Sprintf("%04X%04X", t.lang, t.codepage)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		result, found := queryVersionString(data, size, t.lang, t.codepage, fieldName)
		if found && result != "" {
			return result, true
		}
	}

	return "", false
}

// ========================================================================
// versionTranslation 语言/代码页对
// ========================================================================

// versionTranslationPair 语言/代码页对
type versionTranslationPair struct {
	lang     uint16
	codepage uint16
}

// ========================================================================
// loadVersionInfo — 加载 Win32 Version 资源
// ========================================================================

// loadVersionInfo 调用 GetFileVersionInfoSizeW/GetFileVersionInfoW 加载版本资源。
// [S 汇编实证 0x140748a80, 384B]
// 返回 (dataBlock, dataSize, error)
func loadVersionInfo(path string) ([]byte, uint32, error) {
	p := strings.TrimSpace(path)
	if p == "" {
		return nil, 0, errors.New("loadVersionInfo: empty path")
	}

	pathPtr, err := windows.UTF16PtrFromString(p)
	if err != nil {
		return nil, 0, err
	}

	// GetFileVersionInfoSizeW
	var handle uintptr
	size, err := getFileVersionInfoSize(pathPtr, &handle)
	if err != nil || size == 0 {
		return nil, 0, err
	}

	// allocate buffer
	buf := make([]byte, size)

	// GetFileVersionInfoW
	if err := getFileVersionInfo(pathPtr, handle, size, unsafe.Pointer(&buf[0])); err != nil {
		return nil, 0, err
	}

	return buf, size, nil
}

// ========================================================================
// queryVersionTranslations — 查询并解码 Translations 块
// ========================================================================

// queryVersionTranslations 从版本资源解析语言/代码页对列表。
// [S 汇编实证 0x140748c00, 160B]
func queryVersionTranslations(data []byte, size uint32) ([]versionTranslationPair, error) {
	ptr, count, found := queryVersionValueLocation(data, size, "\\StringFileInfo\\Translation")
	if !found {
		return nil, errors.New("queryVersionTranslations: \\StringFileInfo\\Translation not found")
	}
	return decodeVersionTranslations(ptr, count)
}

// ========================================================================
// decodeVersionTranslations — 解码 Translation 块为 []versionTranslation
// ========================================================================

// decodeVersionTranslations 把 VerQueryValue 返回的 Translation 数据解析为语言/代码页对。
// [S 汇编实证 0x140748ca0, 608B]
func decodeVersionTranslations(ptr unsafe.Pointer, count int) ([]versionTranslationPair, error) {
	if ptr == nil || count < 4 || count%4 != 0 {
		return nil, errors.New("decodeVersionTranslations: invalid translation block")
	}

	// Use unsafe to read uint16 pairs
	n := count / 4
	result := make([]versionTranslationPair, n)
	for i := 0; i < n; i++ {
		p := (*[1 << 20]uint16)(ptr)[i*2 : i*2+2]
		result[i] = versionTranslationPair{
			lang:     p[0],
			codepage: p[1],
		}
	}
	return result, nil
}

// ========================================================================
// queryVersionString — 查询指定语言/编码的版本字段值
// ========================================================================

// queryVersionString 查询特定语言/代码页下的版本字段。
// [S 汇编实证 0x140748f00, 416B]
func queryVersionString(data []byte, size uint32, lang, codepage uint16, fieldName string) (string, bool) {
	// 格式："\\StringFileInfo\\%04X%04X\\fieldName"
	subBlock := fmt.Sprintf("\\StringFileInfo\\%04X%04X\\%s", lang, codepage, fieldName)

	ptr, count, found := queryVersionValueLocation(data, size, subBlock)
	if !found {
		return "", false
	}
	utf16Str := decodeVersionUTF16String(ptr, count)
	result := strings.TrimSpace(utf16Str)
	return result, true
}

// ========================================================================
// decodeVersionUTF16String — 解码 UTF-16 版本字符串为 Go string
// ========================================================================

// decodeVersionUTF16String 把 VerQueryValue 返回的 UTF-16 数据转为 Go string。
// [S 汇编实证 0x1407490a0, 384B]
func decodeVersionUTF16String(ptr unsafe.Pointer, count int) string {
	if ptr == nil || count <= 0 {
		return ""
	}
	// Read as []uint16 up to count/2 code points
	n := count / 2
	if n == 0 {
		return ""
	}
	p := (*[1 << 20]uint16)(ptr)[:n]
	return syscall.UTF16ToString(p)
}

// ========================================================================
// queryVersionValueLocation — 查询 Version 资源中子块位置与大小
// ========================================================================

// queryVersionValueLocation 调用 VerQueryValueW 查询子块地址。
// [S 汇编实证 0x140749220, 608B]
// 返回 (unsafe.Pointer, sizeInBytes, found)
func queryVersionValueLocation(data []byte, size uint32, subBlock string) (unsafe.Pointer, int, bool) {
	if data == nil || size == 0 {
		return nil, 0, false
	}

	subPtr, err := windows.UTF16PtrFromString(subBlock)
	if err != nil {
		return nil, 0, false
	}

	var buf unsafe.Pointer
	var bufLen uint32

	// VerQueryValueW
	// The syscall: VerQueryValueW(pBlock, lpSubBlock, lplpBuffer, puLen) → BOOL
	err = verQueryValue(
		unsafe.Pointer(&data[0]),
		subPtr,
		&buf,
		&bufLen,
	)
	if err != nil {
		return nil, 0, false
	}

	if buf == nil || bufLen == 0 {
		return nil, 0, false
	}

	return buf, int(bufLen), true
}

// ========================================================================
// Windows API 代理（version.dll 延迟加载）
// ========================================================================

var (
	versionDLL                  = windows.NewLazySystemDLL("version.dll")
	procGetFileVersionInfoSizeW = versionDLL.NewProc("GetFileVersionInfoSizeW")
	procGetFileVersionInfoW     = versionDLL.NewProc("GetFileVersionInfoW")
	procVerQueryValueW          = versionDLL.NewProc("VerQueryValueW")
)

// getFileVersionInfoSize 封装 GetFileVersionInfoSizeW。
// [S-sig] Win32 syscall 包装（version.DLL），签名实证，体标准 LazyProc.Call 模式。
func getFileVersionInfoSize(lptstrFilename *uint16, lpdwHandle *uintptr) (uint32, error) {
	r, _, err := procGetFileVersionInfoSizeW.Call(
		uintptr(unsafe.Pointer(lptstrFilename)),
		uintptr(unsafe.Pointer(lpdwHandle)),
	)
	if r == 0 {
		return 0, err
	}
	return uint32(r), nil
}

// getFileVersionInfo 封装 GetFileVersionInfoW。
// [S-sig] Win32 syscall 包装（version.DLL），签名实证，体标准 LazyProc.Call 模式。
func getFileVersionInfo(lptstrFilename *uint16, dwHandle uintptr, dwLen uint32, lpData unsafe.Pointer) error {
	r, _, _ := procGetFileVersionInfoW.Call(
		uintptr(unsafe.Pointer(lptstrFilename)),
		dwHandle,
		uintptr(dwLen),
		uintptr(lpData),
	)
	if r == 0 {
		return errors.New("GetFileVersionInfoW failed")
	}
	return nil
}

// verQueryValue 封装 VerQueryValueW。
// [S-sig] Win32 syscall 包装（version.DLL），签名实证，体标准 LazyProc.Call 模式。
func verQueryValue(pBlock unsafe.Pointer, lpSubBlock *uint16, lplpBuffer *unsafe.Pointer, puLen *uint32) error {
	r, _, _ := procVerQueryValueW.Call(
		uintptr(pBlock),
		uintptr(unsafe.Pointer(lpSubBlock)),
		uintptr(unsafe.Pointer(lplpBuffer)),
		uintptr(unsafe.Pointer(puLen)),
	)
	if r == 0 {
		return errors.New("VerQueryValueW failed")
	}
	return nil
}
