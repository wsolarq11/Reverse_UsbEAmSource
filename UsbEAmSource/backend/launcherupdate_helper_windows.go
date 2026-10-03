package main

import "golang.org/x/sys/windows"

// openLauncherUpdateReadOnlyHandle 以只读 + 顺序扫描打开启动器更新文件句柄。
// [S 汇编 0x1408c4d20, 128B]：UTF16PtrFromString 失败 → (0,err)；CreateFile(GENERIC_READ,
// SHARE_READ, nil, OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL|FILE_FLAG_SEQUENTIAL_SCAN, 0)。
func openLauncherUpdateReadOnlyHandle(path string) (windows.Handle, error) {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	return windows.CreateFile(
		ptr,
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL|windows.FILE_FLAG_SEQUENTIAL_SCAN,
		0,
	)
}

// openLauncherUpdateReadOnlyInheritedHandle 在只读句柄上置 HANDLE_FLAG_INHERIT 使子进程可继承。
// [S 汇编 0x1408c4da0, 160B]：openLauncherUpdateReadOnlyHandle 失败 → (0,err)；
// SetHandleInformation(handle, HANDLE_FLAG_INHERIT, HANDLE_FLAG_INHERIT) 失败 →
// CloseHandle(handle) 后 (0,err)；成功 → (handle,nil)。
func openLauncherUpdateReadOnlyInheritedHandle(path string) (windows.Handle, error) {
	handle, err := openLauncherUpdateReadOnlyHandle(path)
	if err != nil {
		return 0, err
	}
	if err := windows.SetHandleInformation(handle, windows.HANDLE_FLAG_INHERIT, windows.HANDLE_FLAG_INHERIT); err != nil {
		windows.CloseHandle(handle)
		return 0, err
	}
	return handle, nil
}
