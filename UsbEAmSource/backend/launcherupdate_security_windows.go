// AUTO-RECONSTRUCTED — DOMAIN: launcher update security / reparse-point detection
// 研究用途。反汇编实证：
//
//	launcherUpdatePathHasReparsePoint 0x1408cb340, 256B
//	replaceLauncherUpdateMetadataFile 0x1408cb440, 192B
//
// launcherUpdatePathHasReparsePoint:
//
//	TrimSpace → UTF16PtrFromString → GetFileAttributes
//	→ 双 errors.Is(fs.ErrNotExist, ERROR_PATH_NOT_FOUND) → (false, nil)
//	→ 其他 err → (false, err); 成功 → (attrs&REPARSE_POINT != 0, nil)
//
// replaceLauncherUpdateMetadataFile:
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

// launcherUpdatePathHasReparsePoint 检查路径是否为或位于 reparse point 上。
// 比简化版 hasReparsePoint（仅 os.ModeSymlink）更精确：
// 使用 GetFileAttributesW + FILE_ATTRIBUTE_REPARSE_POINT 位检测。
// [S 汇编 0x1408cb340, 256B]：TrimSpace → UTF16PtrFromString → GetFileAttributes；
// 双 errors.Is(fs.ErrNotExist, ERROR_PATH_NOT_FOUND) → (false, nil)。
func launcherUpdatePathHasReparsePoint(p string) (bool, error) {
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

// replaceLauncherUpdateMetadataFile 原子替换目标文件（先写临时文件后调用此函数完成替换）。
// [S 汇编 0x1408cb440, 192B]：UTF16PtrFromString×2 → MoveFileEx(REPLACE_EXISTING|WRITE_THROUGH)。
func replaceLauncherUpdateMetadataFile(src, dst string) error {
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
