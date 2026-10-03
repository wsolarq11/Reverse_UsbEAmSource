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

// SetValue 写 IPropertyStore 属性值。
// [S 汇编 0x14086de40, 256B]：SyscallN(vtbl[0x30] SetValue, this, key, propvariant)
// → HRESULT <0 → fmt.Errorf；否则 nil。
func (s *launcherAppIdentityIPropertyStore) SetValue(key, propvariant uintptr) error {
	setValue := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(s.vtbl)) + 0x30))
	hresult, _, _ := syscall.SyscallN(setValue, uintptr(unsafe.Pointer(s)), key, propvariant)
	if int32(hresult) < 0 {
		return fmt.Errorf("IPropertyStore.SetValue failed: 0x%x", uint32(hresult))
	}
	return nil
}

// launcherAppIdentityIPersistFile Win32 IPersistFile 接口指针包装。
type launcherAppIdentityIPersistFile struct {
	vtbl *uintptr
}

// Load 加载 IPersistFile（文件名 + 模式）。
// [S 汇编 0x14086dc40, 256B]：SyscallN(vtbl[0x28] Load, this, fileName, mode)
// → HRESULT <0 → fmt.Errorf；否则 nil。
func (f *launcherAppIdentityIPersistFile) Load(fileName uintptr, mode uint32) error {
	load := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(f.vtbl)) + 0x28))
	hresult, _, _ := syscall.SyscallN(load, uintptr(unsafe.Pointer(f)), fileName, uintptr(mode))
	if int32(hresult) < 0 {
		return fmt.Errorf("IPersistFile.Load failed: 0x%x", uint32(hresult))
	}
	return nil
}

// Save 保存 IPersistFile（文件名 + 是否记住）。
// [S 汇编 0x14086dd40, 256B]：SyscallN(vtbl[0x30] Save, this, fileName, remember)
// → HRESULT <0 → fmt.Errorf；否则 nil。
func (f *launcherAppIdentityIPersistFile) Save(fileName uintptr, remember bool) error {
	save := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(f.vtbl)) + 0x30))
	rem := uintptr(0)
	if remember {
		rem = 1
	}
	hresult, _, _ := syscall.SyscallN(save, uintptr(unsafe.Pointer(f)), fileName, rem)
	if int32(hresult) < 0 {
		return fmt.Errorf("IPersistFile.Save failed: 0x%x", uint32(hresult))
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

// resolveLauncherNotificationIconPath 解析启动器通知图标路径（环境变量→缓存目录→拼接）。
// [S-sig 0x14086bda0, 320B]：Getenv → UserCacheDir → filepath.Join(3 段)。体待常量专项还原。
func resolveLauncherNotificationIconPath() (string, error) {
	return "", nil
}

// launcherAppIdentityLPWSTRPropVariant 构建 AppID 的 LPWSTR PROPVARIANT。
// [S-sig 0x14086db00, 320B]：UTF16FromString → PROPVARIANT{vt:0x1f, pwszVal}。体待结构专项还原。
func launcherAppIdentityLPWSTRPropVariant(a string) interface{} {
	_ = a
	return nil
}

// setCurrentProcessExplicitAppUserModelID 设置当前进程显式 AppUserModelID。
// [S-sig 0x14086c620, 352B]：TrimSpace 空→error；UTF16PtrFromString →
// LazyProc.Call(SetCurrentProcessExplicitAppUserModelID, 1) → 失败 fmt.Errorf。
// 体待错误文案专项还原。
func setCurrentProcessExplicitAppUserModelID(a string) error {
	_ = a
	return nil
}

// createLauncherShellLink 创建启动器 Shell 链接（CoCreateInstance）。
// [S-sig 0x14086d780, 352B]：CoCreateInstance(5) → 失败 fmt.Errorf。体待 CLSID 专项还原。
func createLauncherShellLink() (interface{}, error) {
	return nil, nil
}

// resolveLauncherStartMenuShortcutPath 解析启动器开始菜单快捷方式路径。
// [S-sig 0x14086cb00, 416B]：Getenv(7)→TrimSpace→非空返回；UserConfigDir→err→fmt.Errorf；
// filepath.join(dir, "Microsoft", "Windows", "Start Menu", "Programs", ...)。体待路径段专项还原。
func resolveLauncherStartMenuShortcutPath() (string, error) {
	return "", nil
}

// setShortcutAppUserModelID 设置快捷方式 AppUserModelID（TrimSpace 校验 → COM 写入）。
// [S-sig 0x14086cca0, 416B]：TrimSpace×3 → path/id 空→error →
// withLauncherAppIdentityCOM。体待 COM 域专项还原。
func setShortcutAppUserModelID(a, b, c string) error {
	_, _, _ = a, b, c
	return nil
}
