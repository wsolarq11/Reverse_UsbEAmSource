package main

import "golang.org/x/sys/windows"

// workspaceMigrationPathIsReparse 判断路径是否为 reparse point（符号链接/挂载点）。
// [S] ASM 0x1409f4660: UTF16PtrFromString → GetFileAttributes，任一步 err→false，
// 否则 attrs&FILE_ATTRIBUTE_REPARSE_POINT(0x400，bit10)!=0。
func workspaceMigrationPathIsReparse(path string) bool {
	ptr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false
	}
	attrs, err := windows.GetFileAttributes(ptr)
	if err != nil {
		return false
	}
	return attrs&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0
}
