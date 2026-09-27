// archiveextract_open_windows.go — ZIP 提取：NOFOLLOW 文件打开
// 反汇编地址 0x140750d20，用 Win32 CreateFileW 直接实现 NOFOLLOW。
package main

import (
	"os"

	"golang.org/x/sys/windows"
)

// openArchiveRegularFileNoFollow 以 CREATE_NEW 方式打开文件并验证不是目录。
// 由 Win32 CreateFileW 原生保证不上读符号链接。
// [S 汇编 0x140750d20]：CreateFileW(CREATE_NEW)→GetFileInformationByHandle→目录校验→os.NewFile。
func openArchiveRegularFileNoFollow(path string, _ os.FileMode) (*os.File, error) {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}

	handle, err := windows.CreateFile(
		ptr,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		0,
		nil,
		windows.CREATE_NEW,
		windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return nil, err
	}

	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		windows.CloseHandle(handle)
		return nil, err
	}

	if info.FileAttributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0 {
		windows.CloseHandle(handle)
		return nil, newExtractError("解压目标不是普通文件")
	}

	file := os.NewFile(uintptr(handle), path)
	if file == nil {
		windows.CloseHandle(handle)
		return nil, newExtractError("无法包装解压目标句柄")
	}

	return file, nil
}
