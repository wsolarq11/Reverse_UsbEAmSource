// AUTO-RECONSTRUCTED — DOMAIN: launcher AppUserModelID 身份配置
// 研究用途。反汇编实证：
//
//	ensureLauncherAppIdentity     0x14086b100, 96B
//	configureLauncherAppIdentity  0x14086b160, 2650B(0xa5a)
//	ensureLauncherAppIdentity.func1 0x1409f71a0
//
// ensureLauncherAppIdentity 是 sync.Once 单次包装：首次调用触发 configureLauncherAppIdentity
// 把 AppUserModelID 配置缓存到全局 string，后续调用直接返回缓存。
package main

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

// launcherAppIdentityIPropertyStore WinRT IPropertyStore 接口指针包装。
// asm 实证（0x14086e160 Commit）：首字段（+0x00）为 vtable 指针；
// Commit 位于 vtbl[7]（+0x38）经 SyscallN 调用。
type launcherAppIdentityIPropertyStore struct {
	vtbl *uintptr
}

// Commit 提交 WinRT IPropertyStore 属性变更。
// [S 汇编 0x14086e160, 192B]：SyscallN(vtbl[7]@0x38 Commit, this) → HRESULT 有符号 <0
// → fmt.Errorf 包装 HRESULT；否则 (nil)。
func (s *launcherAppIdentityIPropertyStore) Commit() error {
	commit := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(s.vtbl)) + 0x38))
	hresult, _, _ := syscall.SyscallN(commit, uintptr(unsafe.Pointer(s)))
	if int32(hresult) < 0 {
		return fmt.Errorf("IPropertyStore.Commit failed: 0x%x", uint32(hresult))
	}
	return nil
}

// launcherAppIdentityOnce 保护 AppUserModelID 配置的单次执行。
// asm 实证：once.done 位于全局 [rip+0x13efe7b]，与 launcherAppIdentityValue 独立成对。
var launcherAppIdentityOnce sync.Once

// launcherAppIdentityValue 缓存 configureLauncherAppIdentity 的返回值（AppUserModelID 字符串）。
// asm 实证：ptr 存全局 [0x140c10870]、len 存 [0x140c10878]。
var launcherAppIdentityValue string

// ensureLauncherAppIdentity 确保进程 AppUserModelID 已配置，返回配置结果。
// [S 汇编 0x14086b100, 96B] 实证：读全局 once.done；未完成则 sync.Once.doSlow(闭包)；
// 闭包（0x1409f71a0）调用 configureLauncherAppIdentity 并写全局缓存；尾读全局
// [0x140c10870]/[0x140c10878] 两字返回 string。
func ensureLauncherAppIdentity() string {
	launcherAppIdentityOnce.Do(func() {
		launcherAppIdentityValue = configureLauncherAppIdentity()
	})
	return launcherAppIdentityValue
}

// configureLauncherAppIdentity 配置进程级 AppUserModelID（通知图标文件、Toast 注册、
// Start Menu 快捷方式 AppID 等），返回最终 AppUserModelID 字符串。
// [S-sig 0x14086b160, 2650B]：签名经 ensureLauncherAppIdentity.func1 调用点实证——
// 无参数、返回 2 寄存器 = string。体调用链 ensureLauncherNotificationIconFile →
// ensureLauncherToastRegistry → setCurrentProcessExplicitAppUserModelID →
// ensureLauncherStartMenuShortcutWithAppID，体待 AppUserModelID 域专项还原。
func configureLauncherAppIdentity() string {
	return ""
}
