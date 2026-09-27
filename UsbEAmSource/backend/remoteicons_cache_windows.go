// AUTO-RECONSTRUCTED — DOMAIN: remote icon cache reparse-point detection & atomic replace
// 研究用途。反汇编实证：
//
//	remoteIconPathHasReparsePoint 0x140966240, 256B
//	replaceRemoteIconCacheFile   0x140966340, 192B
//
// remoteIconPathHasReparsePoint:
//
//	TrimSpace → UTF16PtrFromString → GetFileAttributes
//	→ 双 errors.Is(fs.ErrNotExist, ERROR_PATH_NOT_FOUND) → (false, nil)
//	→ 其他 err → (false, err); 成功 → (attrs&REPARSE_POINT != 0, nil)
//
// replaceRemoteIconCacheFile:
//
//	UTF16PtrFromString(src) → UTF16PtrFromString(dst)
//	→ MoveFileEx(src, dst, MOVEFILE_REPLACE_EXISTING|MOVEFILE_WRITE_THROUGH)
package main

import (
	"errors"
	"io/fs"
	"strings"

	"golang.org/x/sys/windows"
)

// remoteIconPathHasReparsePoint 检查远程图标缓存路径是否为或位于 reparse point 上。
// 与 launcherUpdatePathHasReparsePoint 同构。
// [S 汇编 0x140966240, 256B]：TrimSpace → UTF16PtrFromString → GetFileAttributes；
// 双 errors.Is(fs.ErrNotExist, ERROR_PATH_NOT_FOUND) → (false, nil)。
func remoteIconPathHasReparsePoint(p string) (bool, error) {
	path := strings.TrimSpace(p)
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false, err
	}
	attrs, err := windows.GetFileAttributes(ptr)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) || errors.Is(err, windows.ERROR_PATH_NOT_FOUND) {
			return false, nil
		}
		return false, err
	}
	return attrs&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0, nil
}

// replaceRemoteIconCacheFile 原子替换远程图标缓存文件。
// [S 汇编 0x140966340, 192B]：UTF16PtrFromString×2 → MoveFileEx(REPLACE_EXISTING|WRITE_THROUGH)。
func replaceRemoteIconCacheFile(src, dst string) error {
	srcPtr, err := windows.UTF16PtrFromString(src)
	if err != nil {
		return err
	}
	dstPtr, err := windows.UTF16PtrFromString(dst)
	if err != nil {
		return err
	}
	return windows.MoveFileEx(srcPtr, dstPtr,
		windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH)
}
