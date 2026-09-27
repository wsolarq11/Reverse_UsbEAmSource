package main

import (
	"log"
	"os"
	"strings"
	"sync"

	"golang.org/x/sys/windows"
)

// DXGI / D3D11 HDR 捕获入口（依赖 appicon_windows.go 的 DLL 加载风格）。
var (
	dxgiDLL  = windows.NewLazySystemDLL("dxgi.dll")
	d3d11DLL = windows.NewLazySystemDLL("d3d11.dll")

	procCreateDXGIFactory1 = dxgiDLL.NewProc("CreateDXGIFactory1")
	procD3D11CreateDevice  = d3d11DLL.NewProc("D3D11CreateDevice")

	screenshotHDRCaptureDebugOnce sync.Once
	screenshotHDRCaptureDebug     bool
)

// parseScreenshotBool 解析布尔字符串，返回 (value, ok)。
// [S] ASM 0x140971620: TrimSpace+ToLower；"1"/"on"/"yes"/"true" → (true,true)；
// "0"/"no"/"off"/"false" → (false,true)；否则 (false,false)。
func parseScreenshotBool(v string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "on", "yes", "true":
		return true, true
	case "0", "no", "off", "false":
		return false, true
	}
	return false, false
}

// parseScreenshotHDRCaptureMode 解析 HDR 捕获模式字符串。
// [S] ASM 0x140971280: TrimSpace+ToLower；disabled 集合 → "disabled"；auto 集合
// → "auto"；force 集合 → "force"；默认 trim 空 → "auto"，否则 "disabled"。
func parseScreenshotHDRCaptureMode(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "0", "no", "off", "false", "disable", "disabled":
		return "disabled"
	case "1", "on", "hdr", "yes", "auto", "true":
		return "auto"
	case "dxgi", "force", "always":
		return "force"
	}
	if strings.TrimSpace(v) == "" {
		return "auto"
	}
	return "disabled"
}

// resolveScreenshotHDRCaptureMode 从环境变量解析 HDR 捕获模式。
// [S] ASM 0x140971240: Getenv("USBEAM_SCREENSHOT_HDR_CAPTURE") → parse。
func resolveScreenshotHDRCaptureMode() string {
	return parseScreenshotHDRCaptureMode(os.Getenv("USBEAM_SCREENSHOT_HDR_CAPTURE"))
}

// isScreenshotHDRCaptureDebugEnabled 惰性读取调试开关。
// [S] ASM 0x1409715c0: sync.Once 闭包 Getenv("USBEAM_SCREENSHOT_HDR_DEBUG")；
// trim 空 → 保持 false；parseScreenshotBool ok → 赋值 value；否则 true。
func isScreenshotHDRCaptureDebugEnabled() bool {
	screenshotHDRCaptureDebugOnce.Do(func() {
		v := strings.TrimSpace(os.Getenv("USBEAM_SCREENSHOT_HDR_DEBUG"))
		if v == "" {
			return
		}
		if b, ok := parseScreenshotBool(v); ok {
			screenshotHDRCaptureDebug = b
		} else {
			screenshotHDRCaptureDebug = true
		}
	})
	return screenshotHDRCaptureDebug
}

// screenshotHDRCaptureDebugLog 输出 HDR 捕获调试日志。
// [S] ASM 0x140971440: 开关关闭 → 返回；否则输出 "[screenshot-hdr] "+format。
func screenshotHDRCaptureDebugLog(format string, args ...any) {
	if !isScreenshotHDRCaptureDebugEnabled() {
		return
	}
	log.Printf("[screenshot-hdr] "+format, args...)
}

// screenshotDXGIHDRCaptureAvailable 检查 DXGI/D3D11 HDR 捕获入口是否可用。
// [S] ASM 0x140971160: CreateDXGIFactory1.Find 失败 → "DXGI 工厂不可用: %v" +
// false；D3D11CreateDevice.Find 失败 → "D3D11 设备创建入口不可用: %v" + false；
// 否则 true。
func screenshotDXGIHDRCaptureAvailable() bool {
	if err := procCreateDXGIFactory1.Find(); err != nil {
		screenshotHDRCaptureDebugLog("DXGI 工厂不可用: %v", err)
		return false
	}
	if err := procD3D11CreateDevice.Find(); err != nil {
		screenshotHDRCaptureDebugLog("D3D11 设备创建入口不可用: %v", err)
		return false
	}
	return true
}
