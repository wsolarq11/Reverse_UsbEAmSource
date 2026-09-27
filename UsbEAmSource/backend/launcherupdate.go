package main

// launcherUpdateUserError 是启动器更新流程的用户可读错误（string 别名）。
type launcherUpdateUserError string

// Error 返回错误文本（string 别名，值接收者直接返回底层串）。
// [S 汇编 0x1408b7f00]：值接收者 0x1408b7f00（ABI 包装 mov [rsp+8],rax; ret），
// 指针接收者 0x140a05c40 为 Go 自动生成的解引用包装（test rax,rax → panicwrap；
// rbx=[e+8]=len、rax=[e]=ptr），二者共同确证 launcherUpdateUserError 为 string 别名。
func (e launcherUpdateUserError) Error() string {
	return string(e)
}
