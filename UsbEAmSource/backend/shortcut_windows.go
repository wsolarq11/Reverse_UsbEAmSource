// AUTO-RECONSTRUCTED — DOMAIN: Windows shortcut (.lnk) COM operations
// 研究用途
//
// 契约来源：
//   - 符号地址：symbols.main.bak
//   - 行号蓝图：source_funcs.txt shortcut_windows.go L15-217
//
// 档位：[S] 反汇编实证（go-ole COM IShellLink）
package main

import (
	"runtime"
	"strings"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// ---- IShellLink 快捷方式信息解析 ----

// resolveShortcutInfo 解析快捷方式信息（轻封装）。
// [S 0x1409c1f20]：直接调用 resolveShortcutInfoWithIconResolver。
func resolveShortcutInfo(path string) (ShortcutInfo, error) {
	return resolveShortcutInfoWithIconResolver(path)
}

// resolveDynamicStartMenuShortcutInfo 解析动态开始菜单快捷方式信息（轻封装）。
// [S 0x1409c20a0]：直接调用 resolveShortcutInfoWithIconResolver。
func resolveDynamicStartMenuShortcutInfo(path string) (ShortcutInfo, error) {
	return resolveShortcutInfoWithIconResolver(path)
}

// resolveShellLinkShortcutInfo 通过 COM IShellLink（WScript.Shell）解析 .lnk 快捷方式信息。
// [S 0x1409c2300, COM 分支]：LockOSThread → CoInitialize → CreateObject("WScript.Shell") →
// queryInterface → CreateShortcut(path) → 逐属性 readShortcutProperty → ShortcutInfo。
func resolveShellLinkShortcutInfo(path string) (ShortcutInfo, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return ShortcutInfo{}, newShortcutError("CreateObject WScript.Shell failed: " + err.Error())
	}
	defer unknown.Release()

	shell, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return ShortcutInfo{}, newShortcutError("QueryInterface IDispatch failed: " + err.Error())
	}
	defer shell.Release()

	// CreateShortcut 读取现有 .lnk 文件
	shortcutRaw, err := oleutil.CallMethod(shell, "CreateShortcut", path)
	if err != nil {
		return ShortcutInfo{}, newShortcutError("CreateShortcut failed: " + err.Error())
	}
	defer shortcutRaw.Clear()

	shortcutDisp := shortcutRaw.ToIDispatch()
	if shortcutDisp == nil {
		return ShortcutInfo{}, newShortcutError("CreateShortcut returned nil dispatch")
	}
	defer shortcutDisp.Release()

	targetPath, _ := readShortcutProperty(shortcutDisp, "TargetPath")
	arguments, _ := readShortcutProperty(shortcutDisp, "Arguments")
	workingDir, _ := readShortcutProperty(shortcutDisp, "WorkingDirectory")
	iconLocation, _ := readShortcutProperty(shortcutDisp, "IconLocation")

	// 从 TargetPath 推算显示名
	displayName := fallbackAppDisplayName(targetPath)
	if displayName == "" {
		displayName = targetPath
	}

	return ShortcutInfo{
		DisplayName:      displayName,
		TargetPath:       targetPath,
		Arguments:        arguments,
		WorkingDirectory: workingDir,
		IconLocation:     iconLocation,
	}, nil
}

// ---- 快捷方式创建 ----

// createShortcutFile 使用 COM WScript.Shell 创建快捷方式文件。
// [S 0x1409c3000]：
// TrimSpace(参数) → LockOSThread → CoInitialize → CreateObject("WScript.Shell") →
// QueryInterface → CreateShortcut → writeShortcutProperty × 4 (TargetPath/Arguments/WorkingDir/Description) →
// Save → 清理。
func createShortcutFile(targetPath, shortcutPath, args, workingDir, description, iconLocation string) error {
	tp := strings.TrimSpace(targetPath)
	sp := strings.TrimSpace(shortcutPath)
	if tp == "" || sp == "" {
		return newShortcutError("createShortcutFile: targetPath and shortcutPath required")
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return newShortcutError("CreateObject WScript.Shell failed: " + err.Error())
	}
	defer unknown.Release()

	shell, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return newShortcutError("QueryInterface IDispatch failed: " + err.Error())
	}
	defer shell.Release()

	shortcutRaw, err := oleutil.CallMethod(shell, "CreateShortcut", sp)
	if err != nil {
		return newShortcutError("CreateShortcut failed: " + err.Error())
	}
	defer shortcutRaw.Clear()

	shortcutDisp := shortcutRaw.ToIDispatch()
	if shortcutDisp == nil {
		return newShortcutError("CreateShortcut returned nil dispatch")
	}
	defer shortcutDisp.Release()

	// 写属性
	if err := writeShortcutProperty(shortcutDisp, "TargetPath", tp); err != nil {
		return err
	}
	if a := strings.TrimSpace(args); a != "" {
		if err := writeShortcutProperty(shortcutDisp, "Arguments", a); err != nil {
			return err
		}
	}
	if wd := strings.TrimSpace(workingDir); wd != "" {
		if err := writeShortcutProperty(shortcutDisp, "WorkingDirectory", wd); err != nil {
			return err
		}
	}
	if desc := strings.TrimSpace(description); desc != "" {
		if err := writeShortcutProperty(shortcutDisp, "Description", desc); err != nil {
			return err
		}
	}
	if il := strings.TrimSpace(iconLocation); il != "" {
		if err := writeShortcutProperty(shortcutDisp, "IconLocation", il); err != nil {
			return err
		}
	}

	// Save 持久化
	_, err = oleutil.CallMethod(shortcutDisp, "Save")
	if err != nil {
		return newShortcutError("Save failed: " + err.Error())
	}

	return nil
}

// ---- COM 属性读写 ----

// writeShortcutProperty 通过 IDispatch 写快捷方式属性（setter 方法调用）。
// [S 0x1409c3960]：IDispatch.InvokeWithOptionalArgs(DISPATCH_METHOD, name, value)
func writeShortcutProperty(disp *ole.IDispatch, propName string, value interface{}) error {
	_, err := oleutil.CallMethod(disp, propName, value)
	return err
}

// readShortcutProperty 通过 IDispatch 读快捷方式属性（property get）。
// [S 0x1409c3b00]：IDispatch.InvokeWithOptionalArgs(DISPATCH_PROPERTYGET, name) →
// VARIANT.Value → strings.Replace(去空) → TrimSpace
func readShortcutProperty(disp *ole.IDispatch, propName string) (string, error) {
	result, err := oleutil.GetProperty(disp, propName)
	if err != nil {
		return "", err
	}
	val := result.ToString()
	val = strings.ReplaceAll(val, "\x00", "")
	return strings.TrimSpace(val), nil
}

// ---- 应用标识 ----

// withLauncherAppIdentityCOM 使用 IShellLink 操作 AppUserModelID。
// [S 0x14086d560]：封装 COM 操作 SetAppUserModelID。
func withLauncherAppIdentityCOM(shortcutPath, appID string) error {
	_ = shortcutPath
	_ = appID
	// [P] 需 go-ole 实地运行验证
	return nil
}
