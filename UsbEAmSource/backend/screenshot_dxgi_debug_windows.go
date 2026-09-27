package main

import (
	"fmt"
	"strings"
)

// screenshotDXGIColorSpaceDebugName 返回 DXGI 颜色空间枚举的调试名。
// [S] ASM 0x140970ec0: switch 0/12/13/14/16/18/19 → 命名；其余 UNKNOWN。
func screenshotDXGIColorSpaceDebugName(v uint32) string {
	switch v {
	case 0:
		return fmt.Sprintf("RGB_FULL_G22_NONE_P709(%d)", v)
	case 12:
		return fmt.Sprintf("RGB_FULL_G2084_NONE_P2020(%d)", v)
	case 13:
		return fmt.Sprintf("YCBCR_STUDIO_G2084_LEFT_P2020(%d)", v)
	case 14:
		return fmt.Sprintf("RGB_STUDIO_G2084_NONE_P2020(%d)", v)
	case 16:
		return fmt.Sprintf("YCBCR_STUDIO_G2084_TOP_P2020(%d)", v)
	case 18:
		return fmt.Sprintf("YCBCR_STUDIO_GHLG_TOP_P2020(%d)", v)
	case 19:
		return fmt.Sprintf("YCBCR_FULL_GHLG_TOP_P2020(%d)", v)
	}
	return fmt.Sprintf("UNKNOWN(%d)", v)
}

// screenshotDisplayCaptureLabel 生成显示器调试标签。
// [S] ASM 0x1409706c0: DeviceName 非空 → 返回；否则 AdapterName#OutputIndex；
// 否则 adapter=%d output=%d。
func screenshotDisplayCaptureLabel(info screenshotDisplayCaptureInfo) string {
	if t := strings.TrimSpace(info.DeviceName); t != "" {
		return t
	}
	if t := strings.TrimSpace(info.AdapterName); t != "" {
		return fmt.Sprintf("%s#%d", t, info.OutputIndex)
	}
	return fmt.Sprintf("adapter=%d output=%d", info.AdapterIndex, info.OutputIndex)
}

// formatScreenshotDisplayCaptureInfoForDebug 格式化显示器捕获信息。
// [S] ASM 0x140970b60: 15 参数 fmt.Sprintf，label 走 screenshotDisplayCaptureLabel，
// 颜色空间走 screenshotDXGIColorSpaceDebugName。
func formatScreenshotDisplayCaptureInfoForDebug(info screenshotDisplayCaptureInfo) string {
	return fmt.Sprintf(
		"label=%q adapter=%d output=%d adapterName=%q device=%q attached=%t bounds=%v rotation=%d bits=%d colorSpace=%s hdr=%t advanced=%t luminance[min=%.2f max=%.2f fullFrame=%.2f]",
		screenshotDisplayCaptureLabel(info),
		info.AdapterIndex,
		info.OutputIndex,
		strings.TrimSpace(info.AdapterName),
		strings.TrimSpace(info.DeviceName),
		info.AttachedToDesktop,
		info.Bounds,
		info.Rotation,
		info.BitsPerColor,
		screenshotDXGIColorSpaceDebugName(info.ColorSpace),
		info.HDR,
		info.AdvancedColor,
		info.MinLuminance,
		info.MaxLuminance,
		info.MaxFullFrameLuminance,
	)
}

// logScreenshotHDRCaptureBackendDecision 输出 HDR 后端选择决策日志。
// [S] ASM 0x140970820: 开关关闭 → 返回；先找首个附着且边界有效且 HDR 的显示器，
// 记 needHDR；输出后端选择汇总 + 逐显示器详情。
func logScreenshotHDRCaptureBackendDecision(mode string, backend string, displays []screenshotDisplayCaptureInfo) {
	if !isScreenshotHDRCaptureDebugEnabled() {
		return
	}
	needHDR := false
	for _, info := range displays {
		if info.AttachedToDesktop &&
			info.Bounds.Max.X > info.Bounds.Min.X &&
			info.Bounds.Max.Y > info.Bounds.Min.Y &&
			info.HDR {
			needHDR = true
			break
		}
	}
	screenshotHDRCaptureDebugLog("DXGI HDR 后端选择: mode=%s backend=%s displays=%d needHDR=%t", mode, backend, len(displays), needHDR)
	for _, info := range displays {
		screenshotHDRCaptureDebugLog("DXGI 显示器: %s", formatScreenshotDisplayCaptureInfoForDebug(info))
	}
}
