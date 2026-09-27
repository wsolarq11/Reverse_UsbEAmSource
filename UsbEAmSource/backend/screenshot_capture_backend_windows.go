package main

import (
	"errors"
	"image"
)

// 本文件承载截图后端收口链：默认后端选择 + 按 options 捕获入口 + DXGI 后端
// 接口方法（转发到 DXGI 捕获核心，核心函数体在后续批次落地的
// buildScreenshotDXGIVirtualScreenImage / drawScreenshotCaptureCursorOnRGBA
// 等域中）。

// defaultScreenshotScreenCaptureBackend 选择默认截图后端。
// [S] ASM 0x14096de60: resolveScreenshotHDRCaptureMode() == "disabled" →
// GDI；!screenshotDXGIHDRCaptureAvailable() → GDI；
// enumerateScreenshotDisplayCaptureInfos() err != nil → debug log + GDI；
// selectScreenshotScreenCaptureBackendName(displays, mode, true)；
// logScreenshotHDRCaptureBackendDecision(mode, name, displays)；
// name != "dxgi-hdr" → GDI；否则 debug log 后返回
// &screenshotDXGIHDRScreenCaptureBackend{displays: displays}。
func defaultScreenshotScreenCaptureBackend() screenshotScreenCaptureBackend {
	mode := resolveScreenshotHDRCaptureMode()
	if mode == "disabled" {
		return screenshotGDIScreenCaptureBackend{}
	}
	if !screenshotDXGIHDRCaptureAvailable() {
		return screenshotGDIScreenCaptureBackend{}
	}
	displays, err := enumerateScreenshotDisplayCaptureInfos()
	if err != nil {
		screenshotHDRCaptureDebugLog("枚举 DXGI 显示器失败，使用 GDI 后端: %v", err)
		return screenshotGDIScreenCaptureBackend{}
	}
	name := selectScreenshotScreenCaptureBackendName(displays, mode, true)
	logScreenshotHDRCaptureBackendDecision(mode, name, displays)
	if name != "dxgi-hdr" {
		return screenshotGDIScreenCaptureBackend{}
	}
	screenshotHDRCaptureDebugLog("启用 DXGI HDR 截图后端: mode=%s displays=%d", mode, len(displays))
	return &screenshotDXGIHDRScreenCaptureBackend{displays: displays}
}

// captureScreenshotVirtualScreenSnapshotWithOptions 按 options 捕获虚拟屏幕。
// [S] ASM 0x14096e280: defaultScreenshotScreenCaptureBackend().captureVirtualScreen(options)。
func captureScreenshotVirtualScreenSnapshotWithOptions(options screenshotScreenCaptureOptions) qrCodeScreenSnapshot {
	return defaultScreenshotScreenCaptureBackend().captureVirtualScreen(options)
}

// captureScreenshotScreenRectWithOptions 按 options 捕获屏幕矩形。
// [S] ASM 0x14096e420: defaultScreenshotScreenCaptureBackend().captureScreenRect(rect, options)。
func captureScreenshotScreenRectWithOptions(rect image.Rectangle, options screenshotScreenCaptureOptions) (*image.RGBA, error) {
	return defaultScreenshotScreenCaptureBackend().captureScreenRect(rect, options)
}

// name 返回 DXGI HDR 后端名。
// [S] selectScreenshotScreenCaptureBackendName 0x14096e0a0 命中分支返回
// rodata "dxgi-hdr"（8 字节）；defaultScreenshotScreenCaptureBackend 0x14096de60
// 以 name=="dxgi-hdr" 作为 dxgi 后端判据。
func (screenshotDXGIHDRScreenCaptureBackend) name() string {
	return "dxgi-hdr"
}

// captureVirtualScreen 捕获虚拟屏幕；DXGI 失败时回退 GDI。
// [S] ASM 0x14096e920: captureVirtualScreenDXGI err != nil → debug log
// （"DXGI 虚拟屏幕捕获失败，回退 GDI: %v"）+ GDI 回退；否则返回 DXGI 快照。
func (b screenshotDXGIHDRScreenCaptureBackend) captureVirtualScreen(options screenshotScreenCaptureOptions) qrCodeScreenSnapshot {
	snapshot, err := b.captureVirtualScreenDXGI(options)
	if err != nil {
		screenshotHDRCaptureDebugLog("DXGI 虚拟屏幕捕获失败，回退 GDI: %v", err)
		return screenshotGDIScreenCaptureBackend{}.captureVirtualScreen(options)
	}
	return snapshot
}

// captureScreenRect 捕获屏幕矩形；DXGI 失败时回退 GDI。
// [S] ASM 0x14096eb40: captureScreenRectDXGI err != nil → debug log
// （"DXGI 区域捕获失败，回退 GDI: rect=%v err=%v"）+ GDI 回退；否则返回 (img, nil)。
func (b screenshotDXGIHDRScreenCaptureBackend) captureScreenRect(rect image.Rectangle, options screenshotScreenCaptureOptions) (*image.RGBA, error) {
	img, err := b.captureScreenRectDXGI(rect, options)
	if err != nil {
		screenshotHDRCaptureDebugLog("DXGI 区域捕获失败，回退 GDI: rect=%v err=%v", rect, err)
		return screenshotGDIScreenCaptureBackend{}.captureScreenRect(rect, options)
	}
	return img, nil
}

// captureVirtualScreenDXGI 是 DXGI 虚拟屏幕捕获核心。
// [S] ASM 0x14096ed00: qrCodeVirtualScreenBounds 空 →
// "未检测到可用的屏幕区域"；buildScreenshotDXGIVirtualScreenImage(displays,
// bounds, captureScreenshotDXGIDisplayFrameWithFallback) 失败透传；
// drawScreenshotCaptureCursorOnRGBA 叠光标；buildDarkenedQRCodeSelectionPreview
// 生成暗化图；返回 qrCodeScreenSnapshot{bounds, original, shaded}。
func (b screenshotDXGIHDRScreenCaptureBackend) captureVirtualScreenDXGI(options screenshotScreenCaptureOptions) (qrCodeScreenSnapshot, error) {
	bounds := qrCodeVirtualScreenBounds()
	if bounds.Empty() {
		return qrCodeScreenSnapshot{}, errors.New("未检测到可用的屏幕区域")
	}
	img, err := buildScreenshotDXGIVirtualScreenImage(b.displays, bounds, captureScreenshotDXGIDisplayFrameWithFallback)
	if err != nil {
		return qrCodeScreenSnapshot{}, err
	}
	drawScreenshotCaptureCursorOnRGBA(img, bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y, options.captureCursor, options.cursorSnapshot)
	shaded := buildDarkenedQRCodeSelectionPreview(img)
	return qrCodeScreenSnapshot{virtualBounds: bounds, original: img, shaded: shaded}, nil
}

// captureScreenRectDXGI 是 DXGI 屏幕矩形捕获核心。
// [S] ASM 0x14096ef60: normalizeScreenshotScrollingRect(rect) 空 →
// "截图区域为空"；captureScreenshotDXGIRegionImageWithOptions(displays,
// norm, options.requireFreshDXGIFrame) 失败透传；drawScreenshotCaptureCursorOnRGBA
// 叠光标；返回 (img, nil)。
func (b screenshotDXGIHDRScreenCaptureBackend) captureScreenRectDXGI(rect image.Rectangle, options screenshotScreenCaptureOptions) (*image.RGBA, error) {
	norm := normalizeScreenshotScrollingRect(rect)
	if norm.Empty() {
		return nil, errors.New("截图区域为空")
	}
	img, err := captureScreenshotDXGIRegionImageWithOptions(b.displays, norm, options.requireFreshDXGIFrame)
	if err != nil {
		return nil, err
	}
	drawScreenshotCaptureCursorOnRGBA(img, norm.Min.X, norm.Min.Y, norm.Max.X, norm.Max.Y, options.captureCursor, options.cursorSnapshot)
	return img, nil
}
