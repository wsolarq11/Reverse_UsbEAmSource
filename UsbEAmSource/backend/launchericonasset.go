// AUTO-RECONSTRUCTED — DOMAIN: launcher icon asset
// 研究用途
package main

import (
	"strings"
	"unsafe"
)

// defaultAppIconOptions 返回默认图标解析选项。
// [S-inline] 原 exe 无独立符号：rodata 0x141be9500（56B = 7 qword）在 ResolveAppIconResource 与
// attachWindowManagementTargetIconURL 两处被逐 qword 内联复制到栈，此处提取为 DRY 辅助函数。
//
//	IconIndex=0（经 ecx 寄存器置零传入）、Namespace=""、
//	Size=256、ImageList=4（SHIL_JUMBO）、
//	CandsPtr=&defaultAppIconSizes[0]（0x141965580）、CandsLen=5、CandsCap=5。
func defaultAppIconOptions() AppIconOptions {
	return AppIconOptions{
		Size:      256,
		ImageList: 4,
		CandsPtr:  unsafe.Pointer(&defaultAppIconSizes[0]),
		CandsLen:  len(defaultAppIconSizes),
		CandsCap:  cap(defaultAppIconSizes[:]),
	}
}

// ResolveAppIconResource 解析应用图标资源。
// [S asm 0x1408a17e0, 320B] 完整控制流：
//
//	压默认 AppIconOptions（rodata 0x141be9500, 56B）→
//	resolveAppIconDataWithOptions(path, opts) → iconData
//	→ buildLauncherIconResource("icon/app", iconData)。
func (bs *BootstrapService) ResolveAppIconResource(path string) LauncherIconResource {
	iconData := resolveAppIconDataWithOptions(path, defaultAppIconOptions())
	return bs.buildLauncherIconResource("icon/app", iconData)
}

// attachWindowManagementTargetIconURL 给窗口管理目标附加图标 URL。
// [S asm 0x1408a45c0, 800B] 完整控制流：
//
//	normalizeWindowManagementTarget(target) →
//	buildLauncherConfigIconResource("icon/window-management", target.IconRef, target.IconData)
//	→ 写回 IconRef/IconURL，清零 IconData
//	→ 若 !launcherConfigIconSlotHasSource("", IconRef) 且 IconURL=="" 且 TrimSpace(Path)!=""
//	  → resolveAppIconDataWithOptions(Path, 默认 opts) → iconData
//	  → buildLauncherIconResource("icon/window-management", iconData) → 写回 IconURL。
func (bs *BootstrapService) attachWindowManagementTargetIconURL(target WindowManagementTarget) WindowManagementTarget {
	target = normalizeWindowManagementTarget(target)
	res := bs.buildLauncherConfigIconResource("icon/window-management", target.IconRef, target.IconData)
	target.IconRef = res.IconRef
	target.IconURL = res.IconURL
	target.IconData = ""

	if !launcherConfigIconSlotHasSource("", target.IconRef) && target.IconURL == "" {
		if strings.TrimSpace(target.Path) != "" {
			iconData := resolveAppIconDataWithOptions(target.Path, defaultAppIconOptions())
			res := bs.buildLauncherIconResource("icon/window-management", iconData)
			target.IconURL = res.IconURL
		}
	}
	return target
}

// attachWindowManagementStateIconURLs 给窗口管理状态的 Config.Target 与 Target 附加图标 URL。
// [S asm 0x1408a48e0, 416B] 完整控制流：
//
//	attachWindowManagementTargetIconURL(state.Config.Target)（target @ +0x28）
//	→ attachWindowManagementTargetIconURL(state.Target)（target @ +0x100）。
func (bs *BootstrapService) attachWindowManagementStateIconURLs(state WindowManagementState) WindowManagementState {
	state.Config.Target = bs.attachWindowManagementTargetIconURL(state.Config.Target)
	state.Target = bs.attachWindowManagementTargetIconURL(state.Target)
	return state
}

// persistAndRefreshWindowManagementState 持久化并刷新窗口管理状态。
// [S asm 0x1408a4a80, 1184B] 完整控制流：
//
//	priorErr != nil → attachWindowManagementStateIconURLs(state) + priorErr
//	persistWindowManagementConfig(state.Config) err != nil → attach(state) + err
//	windowManagement.GetState() err != nil → attach(state) + err
//	否则 → attach(st) + nil。
func (bs *BootstrapService) persistAndRefreshWindowManagementState(state WindowManagementState, priorErr error) (WindowManagementState, error) {
	if priorErr != nil {
		return bs.attachWindowManagementStateIconURLs(state), priorErr
	}
	if err := bs.persistWindowManagementConfig(state.Config); err != nil {
		return bs.attachWindowManagementStateIconURLs(state), err
	}
	st, err := bs.windowManagement.GetState()
	if err != nil {
		return bs.attachWindowManagementStateIconURLs(state), err
	}
	return bs.attachWindowManagementStateIconURLs(st), nil
}
