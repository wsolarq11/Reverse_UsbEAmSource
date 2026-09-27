// AUTO-RECONSTRUCTED EXEC DOMAIN — DOMAIN: memoryrelease (系统内存列表释放执行链)
// 研究用途
//
// 契约来源（全部 [S] 汇编实证，VA 见各函数注释）：
//   - executeMemoryReleaseMode           0x1408d5120  160B  + func1 0x1408d51c0 (256B)
//   - resolveMemoryReleaseCommand        0x1408d52c0  448B
//   - withMemoryReleasePrivilege         0x1408d5480 1120B  + func1 0x1408d58e0 (恢复权限 defer)
//   - adjustMemoryReleaseTokenPrivileges 0x1408d5c00  448B
//
// SYSTEM_MEMORY_LIST_COMMAND 映射（resolveMemoryReleaseCommand 逐字节 cmp 实证）：
//   "standbylist"           → 4 (MemoryPurgeStandbyList)
//   "workingsets"           → 2
//   "modifiedpagelist"      → 3
//   "priority0standbylist"  → 5 (LowPriority)
package main

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// memoryReleaseError 是内存释放执行链的具名错误载体（16 字节 struct { msg string }）。
// [S] newobject(0x140b54380, 16B) 后写 [obj]=msg.ptr / [obj+8]=msg.len，以 itab
// 0x1411d3100（*memoryReleaseError 实现 error）返回。
type memoryReleaseError struct {
	msg string
}

// Error 返回错误消息。[S-inline 内联，语义确定]
func (e *memoryReleaseError) Error() string { return e.msg }

// resolveMemoryReleaseCommand 把规范模式名映射为 SYSTEM_MEMORY_LIST_COMMAND 命令值。
// [S 汇编 0x1408d52c0, 448B] 实证：
//
//	parseMemoryReleaseMode 失败 → (0, fmt.Errorf("不支持的内存释放模式: %s" /*34B*/, TrimSpace(mode)))
//	"standbylist"          → (4, nil)
//	"workingsets"          → (2, nil)
//	"modifiedpagelist"     → (3, nil)
//	"priority0standbylist" → (5, nil)
//	其余 → (0, fmt.Errorf(34B, TrimSpace(mode)))
func resolveMemoryReleaseCommand(mode string) (uint32, error) {
	normalized, ok := parseMemoryReleaseMode(mode)
	if !ok {
		return 0, fmt.Errorf("不支持的内存释放模式: %s", strings.TrimSpace(mode))
	}
	switch normalized {
	case "standbylist":
		return 4, nil
	case "workingsets":
		return 2, nil
	case "modifiedpagelist":
		return 3, nil
	case "priority0standbylist":
		return 5, nil
	}
	return 0, fmt.Errorf("不支持的内存释放模式: %s", strings.TrimSpace(mode))
}

// executeMemoryReleaseMode 执行一次系统内存列表释放。
// [S 汇编 0x1408d5120, 160B] 实证：
//
//	parseMemoryReleaseMode(mode) 失败 → normalized = "standbylist"（11B 回退）
//	cmd, err := resolveMemoryReleaseCommand(normalized)；err 非空 → 返回 err
//	withMemoryReleasePrivilege("SeProfileSingleProcessPrivilege" /*31B*/, func() error { ... })
func executeMemoryReleaseMode(mode string) error {
	normalized, ok := parseMemoryReleaseMode(mode)
	if !ok {
		normalized = "standbylist"
	}
	cmd, err := resolveMemoryReleaseCommand(normalized)
	if err != nil {
		return err
	}
	return withMemoryReleasePrivilege("SeProfileSingleProcessPrivilege", func() error {
		// [S func1 0x1408d51c0] NtSetSystemInformation(0x50 /*SystemMemoryListInformation*/, &cmd, 4)
		if err := windows.NtSetSystemInformation(windows.SystemMemoryListInformation, unsafe.Pointer(&cmd), 4); err != nil {
			if err == windows.STATUS_PRIVILEGE_NOT_HELD {
				return &memoryReleaseError{msg: "当前进程没有足够权限，请使用管理员身份启动后再试"}
			}
			return fmt.Errorf("执行内存释放失败: %w", err)
		}
		return nil
	})
}

// withMemoryReleasePrivilege 在启用指定权限后执行 fn，并在退出时恢复权限。
// [S 汇编 0x1408d5480, 1120B] 实证：
//
//	TrimSpace(name) 为空 → &memoryReleaseError{"缺少内存释放所需的系统权限名" /*42B*/}
//	fn == nil → &memoryReleaseError{"缺少内存释放原生操作" /*30B*/}
//	LockOSThread + defer UnlockOSThread
//	UTF16PtrFromString 失败 → fmt.Errorf("准备系统权限失败: %w" /*28B*/, err)
//	LookupPrivilegeValue 失败 → fmt.Errorf("查询系统权限失败: %w" /*28B*/, err)
//	openCurrentProcessToken(0x28) 失败 → 返回 err
//	defer token.Close()
//	adjustMemoryReleaseTokenPrivileges 失败 → fmt.Errorf("启用系统权限失败: %w" /*28B*/, err)
//	defer 恢复权限（失败时 errors.Join 原错误与 "恢复系统权限失败: %w" /*28B*/）
//	return fn()
func withMemoryReleasePrivilege(name string, fn func() error) (err error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return &memoryReleaseError{msg: "缺少内存释放所需的系统权限名"}
	}
	if fn == nil {
		return &memoryReleaseError{msg: "缺少内存释放原生操作"}
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	privName, e := windows.UTF16PtrFromString(name)
	if e != nil {
		return fmt.Errorf("准备系统权限失败: %w", e)
	}
	var luid windows.LUID
	if e := windows.LookupPrivilegeValue(nil, privName, &luid); e != nil {
		return fmt.Errorf("查询系统权限失败: %w", e)
	}
	token, e := openCurrentProcessToken(windows.TOKEN_ADJUST_PRIVILEGES | windows.TOKEN_QUERY)
	if e != nil {
		return e
	}
	defer token.Close()

	tp := windows.Tokenprivileges{
		PrivilegeCount: 1,
		Privileges: [1]windows.LUIDAndAttributes{{
			Luid:       luid,
			Attributes: windows.SE_PRIVILEGE_ENABLED,
		}},
	}
	var previous windows.Tokenprivileges
	var retLen uint32
	if e := adjustMemoryReleaseTokenPrivileges(token, &tp, &previous, &retLen); e != nil {
		return fmt.Errorf("启用系统权限失败: %w", e)
	}
	defer func() {
		if e := adjustMemoryReleaseTokenPrivileges(token, &previous, nil, nil); e != nil {
			wrapped := fmt.Errorf("恢复系统权限失败: %w", e)
			if err == nil {
				err = wrapped
			} else {
				err = errors.Join(err, wrapped)
			}
		}
	}()
	return fn()
}

// adjustMemoryReleaseTokenPrivileges 调整进程 token 权限。
// [S 汇编 0x1408d5c00, 448B] 实证：
//
//	newState == nil → &memoryReleaseError{"缺少系统权限状态" /*24B*/}
//	AdjustTokenPrivileges 失败 → 返回 err
//	GetLastError == ERROR_NOT_ALL_ASSIGNED → &memoryReleaseError{"当前进程不是管理员权限，无法启用内存释放所需特权" /*72B*/}
//	其余 → 返回 GetLastError（nil 或其余 Errno）
func adjustMemoryReleaseTokenPrivileges(token windows.Token, newState, previousState *windows.Tokenprivileges, retLen *uint32) error {
	if newState == nil {
		return &memoryReleaseError{msg: "缺少系统权限状态"}
	}
	var buflen uint32
	if previousState != nil {
		buflen = 16
	}
	if err := windows.AdjustTokenPrivileges(token, false, newState, buflen, previousState, retLen); err != nil {
		return err
	}
	le := windows.GetLastError()
	if le == windows.ERROR_NOT_ALL_ASSIGNED {
		return &memoryReleaseError{msg: "当前进程不是管理员权限，无法启用内存释放所需特权"}
	}
	return le
}
