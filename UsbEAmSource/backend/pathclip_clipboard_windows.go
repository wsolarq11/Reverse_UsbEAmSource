// AUTO-RECONSTRUCTED — 文件剪贴板（CF_HDROP）链：路径规范化 + dropEffect 解析 + 编排
// 研究用途
package main

import (
	"encoding/binary"
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/w32"
	"golang.org/x/sys/windows"
)

// dropFilesHeaderSize 是 DROPFILES 结构固定头大小（pFiles + pt + fNC + fWide）。
const dropFilesHeaderSize = 20

// cfHDrop 是 CF_HDROP 剪贴板格式常量。
const cfHDrop = 15

// Windows DROPEFFECT 常量（CF_HDROP Preferred DropEffect）。
const (
	dropEffectCopy = 1 // DROPEFFECT_COPY
	dropEffectMove = 2 // DROPEFFECT_MOVE
)

// normalizePathList 规范化路径列表：TrimSpace → filepath.Clean → ToLower 去重，
// 空路径跳过。返回 []string（无 error）。
// [S 汇编 0x14088c420, 704B] 空输入提前返回 nil；makeslice(cap=len(paths))；
// makemap(map[string]struct{}, len(paths))；循环内 TrimSpace→Clean→ToLower→map 去重→append。
func normalizePathList(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	result := make([]string, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		clean := filepath.Clean(p)
		key := strings.ToLower(clean)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, clean)
	}
	return result
}

// resolveClipboardDropEffect 将字符串解析为 Windows DROPEFFECT。
// [S 汇编 0x1408adbe0, 224B] TrimSpace→ToLower 后比较：
// "cut"/"move" → 2(DROPEFFECT_MOVE)，"copy" → 1(DROPEFFECT_COPY)，
// 其他 → errors.New("不支持的剪贴板文件操作")。
func resolveClipboardDropEffect(effect string) (uint32, error) {
	switch strings.ToLower(strings.TrimSpace(effect)) {
	case "cut":
		return dropEffectMove, nil
	case "copy":
		return dropEffectCopy, nil
	case "move":
		return dropEffectMove, nil
	default:
		return 0, errors.New("不支持的剪贴板文件操作")
	}
}

// setFileClipboard 设置文件剪贴板（CF_HDROP 编排入口）。
// [S 汇编 0x1408a9f00, 256B] normalizePathList 空结果 → errors.New("文件路径不能为空")；
// resolveClipboardDropEffect 失败透传；否则 setWindowsFileClipboard(normalized, effect)。
func setFileClipboard(paths []string, dropEffect string) error {
	normalized := normalizePathList(paths)
	if len(normalized) == 0 {
		return errors.New("文件路径不能为空")
	}
	effect, err := resolveClipboardDropEffect(dropEffect)
	if err != nil {
		return err
	}
	return setWindowsFileClipboard(normalized, effect)
}

// setWindowsFileClipboard 构建并提交 CF_HDROP / Preferred DropEffect 剪贴板数据。
// [S 汇编 0x1408adcc0] LockOSThread→openClipboardForFileOperation→defer CloseClipboard→
// EmptyClipboard→buildHDropClipboardData→buildDropEffectClipboardData→
// SetClipboardData(CF_HDROP)→registerPreferredDropEffectClipboardFormat→
// SetClipboardData(PreferredDropEffect)→显式 CloseClipboard（失败 wrapWinError）。
func setWindowsFileClipboard(paths []string, dropEffect uint32) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := openClipboardForFileOperation(); err != nil {
		return err
	}
	defer w32.CloseClipboard()
	if !w32.EmptyClipboard() {
		return wrapWinError("清空剪贴板失败", windows.GetLastError())
	}
	hDrop, err := buildHDropClipboardData(paths)
	if err != nil {
		return err
	}
	defer func() { w32.GlobalFree(hDrop) }()
	hEffect, err := buildDropEffectClipboardData(dropEffect)
	if err != nil {
		return err
	}
	defer func() { w32.GlobalFree(hEffect) }()
	if w32.SetClipboardData(cfHDrop, hDrop) == 0 {
		return wrapWinError("写入文件剪贴板失败", windows.GetLastError())
	}
	hDrop = 0 // 所有权转移给剪贴板，defer GlobalFree(0) 无害
	format, err := registerPreferredDropEffectClipboardFormat()
	if err != nil {
		return err
	}
	if w32.SetClipboardData(uint(format), hEffect) == 0 {
		return wrapWinError("写入剪贴板文件操作失败", windows.GetLastError())
	}
	hEffect = 0
	if !w32.CloseClipboard() {
		return wrapWinError("关闭剪贴板失败", windows.GetLastError())
	}
	return nil
}

// openClipboardForFileOperation 以 10ms 间隔重试 OpenClipboard 最多 100 次，再最后尝试一次。
// [S 汇编 0x1408ae240, 224B] 循环 i<0x64，i>0 时 time.Sleep(0x989680=10ms)；
// OpenClipboard 成功（非 0）→ 返回 nil；循环后最后尝试一次 OpenClipboard；
// 仍失败 → wrapWinError("打开剪贴板失败")。
func openClipboardForFileOperation() error {
	for i := 0; i < 100; i++ {
		if i > 0 {
			time.Sleep(10 * time.Millisecond)
		}
		if w32.OpenClipboard(0) {
			return nil
		}
	}
	if w32.OpenClipboard(0) {
		return nil
	}
	return wrapWinError("打开剪贴板失败", windows.GetLastError())
}

// registerPreferredDropEffectClipboardFormat 注册 "Preferred DropEffect" 剪贴板格式。
// [S 汇编 0x1408aeb60, 352B] UTF16FromString("Preferred DropEffect") →
// RegisterClipboardFormatW；返回 0 → wrapWinError("注册剪贴板文件操作格式失败")。
func registerPreferredDropEffectClipboardFormat() (uint32, error) {
	name, err := syscall.UTF16FromString("Preferred DropEffect")
	if err != nil {
		return 0, err
	}
	r, _, _ := procRegisterClipboardFormatW.Call(uintptr(unsafe.Pointer(&name[0])))
	if r == 0 {
		return 0, wrapWinError("注册剪贴板文件操作格式失败", windows.GetLastError())
	}
	return uint32(r), nil
}

// encodeHDropPaths 将路径列表编码为 DROPFILES 的 UTF-16 路径块（\0 分隔，双 \0 终止）。
// [S 汇编 0x1408ae7c0, 544B] makeslice([]uint16,0,256)；逐路径 TrimSpace→空跳过→
// syscall.UTF16FromString（含 null）→append；空结果 → errors.New("文件路径不能为空")；
// 否则 append(0) 补双 null 终止。
func encodeHDropPaths(paths []string) ([]uint16, error) {
	buf := make([]uint16, 0, 256)
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		u16, err := syscall.UTF16FromString(p)
		if err != nil {
			return nil, err
		}
		buf = append(buf, u16...)
	}
	if len(buf) == 0 {
		return nil, errors.New("文件路径不能为空")
	}
	return append(buf, 0), nil
}

// buildHDropClipboardData 构建 CF_HDROP 剪贴板数据（DROPFILES 头 + UTF-16 路径块）。
// [S 汇编 0x1408ae320, 608B] encodeHDropPaths 失败透传；globalAllocMoveable(len(u16)*2+20)
// → globalLock（失败 GlobalFree 后透传）；defer GlobalUnlock；写 DROPFILES 头
// （pFiles=20, pt=0, fNC=0, fWide=1）→ 从偏移 20 拷贝路径 UTF-16 → 返回 hMem。
func buildHDropClipboardData(paths []string) (w32.HGLOBAL, error) {
	u16, err := encodeHDropPaths(paths)
	if err != nil {
		return 0, err
	}
	hMem, err := globalAllocMoveable(len(u16)*2 + dropFilesHeaderSize)
	if err != nil {
		return 0, err
	}
	ptr, err := globalLock(hMem)
	if err != nil {
		w32.GlobalFree(hMem)
		return 0, err
	}
	defer w32.GlobalUnlock(hMem)
	header := unsafe.Slice((*byte)(ptr), dropFilesHeaderSize)
	binary.LittleEndian.PutUint32(header[0:4], dropFilesHeaderSize) // pFiles = 20
	binary.LittleEndian.PutUint32(header[16:20], 1)                 // fWide = 1（UTF-16）
	dst := unsafe.Slice((*uint16)(unsafe.Add(ptr, dropFilesHeaderSize)), len(u16))
	copy(dst, u16)
	return hMem, nil
}

// buildDropEffectClipboardData 构建 Preferred DropEffect 剪贴板数据（4 字节 uint32）。
// [S 汇编 0x1408ae5e0, 384B] globalAllocMoveable(4) → globalLock（失败 GlobalFree 后透传）；
// defer GlobalUnlock；写 dropEffect → 返回 hMem。
func buildDropEffectClipboardData(dropEffect uint32) (w32.HGLOBAL, error) {
	hMem, err := globalAllocMoveable(4)
	if err != nil {
		return 0, err
	}
	ptr, err := globalLock(hMem)
	if err != nil {
		w32.GlobalFree(hMem)
		return 0, err
	}
	defer w32.GlobalUnlock(hMem)
	*(*uint32)(ptr) = dropEffect
	return hMem, nil
}
