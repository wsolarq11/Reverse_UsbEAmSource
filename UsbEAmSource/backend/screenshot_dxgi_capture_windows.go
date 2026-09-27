package main

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
)

// 本文件承载 DXGI HDR 截图捕获域（批次 53）：
// 顶层虚拟屏幕/区域合成链 + 单显示器帧捕获与 GDI 兜底链。
// 深层 D3D11 / 桌面复制（captureScreenshotDXGIOutputFrameWithOptions /
// captureScreenshotDXGIOutputFrameRegionWithOptions / 捕获缓存）留待后续批次。

// buildScreenshotDXGIVirtualScreenImage 校验虚拟屏幕矩形后委托桌面合成。
// [S] ASM 0x14096f1a0: bounds.Empty() → "DXGI 虚拟屏幕区域为空"；否则
// 尾调用 buildScreenshotDXGIDesktopBoundsImage（全参透传）。
func buildScreenshotDXGIVirtualScreenImage(displays []screenshotDisplayCaptureInfo, bounds image.Rectangle, capture func(screenshotDisplayCaptureInfo) (*image.RGBA, error)) (*image.RGBA, error) {
	if bounds.Empty() {
		return nil, errors.New("DXGI 虚拟屏幕区域为空")
	}
	return buildScreenshotDXGIDesktopBoundsImage(displays, bounds, capture)
}

// buildScreenshotDXGIDesktopBoundsImage 将多个显示器帧合成到桌面边界画布。
// [S] ASM 0x14096f8e0: bounds 空 → "DXGI 输出区域为空"；capture nil →
// "DXGI 显示器捕获函数未初始化"；NewRGBA(0,0,w,h)；跳过未附着显示器；对
// 每个显示器求 inter=Bounds∩bounds，非空则 capture(display)，img2 空 →
// "DXGI 捕获显示器 %s 返回空图像"，失败 → "DXGI 捕获显示器 %s 失败: %w"；
// 按 (inter.Min-display.Min) 源偏移 + (inter.Min-bounds.Min) 目标偏移
// DrawMask(Src)；零命中 → "DXGI 未捕获到任何已连接显示器"。
func buildScreenshotDXGIDesktopBoundsImage(displays []screenshotDisplayCaptureInfo, bounds image.Rectangle, capture func(screenshotDisplayCaptureInfo) (*image.RGBA, error)) (*image.RGBA, error) {
	if bounds.Empty() {
		return nil, errors.New("DXGI 输出区域为空")
	}
	if capture == nil {
		return nil, errors.New("DXGI 显示器捕获函数未初始化")
	}

	img := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	drawn := 0
	for i := range displays {
		d := displays[i]
		if !d.AttachedToDesktop {
			continue
		}
		inter := d.Bounds.Intersect(bounds)
		if inter.Empty() {
			continue
		}
		img2, err := capture(d)
		if err != nil {
			return nil, fmt.Errorf("DXGI 捕获显示器 %s 失败: %w", screenshotDisplayCaptureLabel(d), err)
		}
		if img2 == nil || img2.Rect.Empty() {
			return nil, fmt.Errorf("DXGI 捕获显示器 %s 返回空图像", screenshotDisplayCaptureLabel(d))
		}

		offsetX := inter.Min.X - d.Bounds.Min.X
		offsetY := inter.Min.Y - d.Bounds.Min.Y
		dstX := inter.Min.X - bounds.Min.X
		dstY := inter.Min.Y - bounds.Min.Y
		w := min(inter.Dx(), img2.Rect.Dx()-offsetX)
		h := min(inter.Dy(), img2.Rect.Dy()-offsetY)
		if w <= 0 || h <= 0 {
			continue
		}
		draw.DrawMask(img, image.Rect(dstX, dstY, dstX+w, dstY+h), img2, img2.Rect.Min.Add(image.Pt(offsetX, offsetY)), nil, image.Point{}, draw.Src)
		drawn++
	}
	if drawn == 0 {
		return nil, errors.New("DXGI 未捕获到任何已连接显示器")
	}
	return img, nil
}

// captureScreenshotDXGIRegionImageWithOptions 将多个显示器的区域帧合成到目标画布。
// [S] ASM 0x14096f260: rect 空 → "DXGI 截图区域为空"；NewRGBA(0,0,w,h)；
// 跳过未附着显示器；inter=Bounds∩rect 非空则调用
// captureScreenshotDXGIDisplayRegionFrameWithFallback(display, inter,
// requireFresh)；img2 空 → "DXGI 捕获显示器 %s 区域 %v 返回空图像"，
// 失败 → "DXGI 捕获显示器 %s 区域 %v 失败: %w"；按 (inter.Min-rect.Min)
// 目标偏移 DrawMask(Src)；零命中 → "DXGI 未捕获到任何已连接显示器"。
func captureScreenshotDXGIRegionImageWithOptions(displays []screenshotDisplayCaptureInfo, rect image.Rectangle, requireFresh bool) (*image.RGBA, error) {
	if rect.Empty() {
		return nil, errors.New("DXGI 截图区域为空")
	}

	img := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	drawn := 0
	for i := range displays {
		d := displays[i]
		if !d.AttachedToDesktop {
			continue
		}
		inter := d.Bounds.Intersect(rect)
		if inter.Empty() {
			continue
		}
		img2, err := captureScreenshotDXGIDisplayRegionFrameWithFallback(d, inter, requireFresh)
		if err != nil {
			return nil, fmt.Errorf("DXGI 捕获显示器 %s 区域 %v 失败: %w", screenshotDisplayCaptureLabel(d), inter, err)
		}
		if img2 == nil || img2.Rect.Empty() {
			return nil, fmt.Errorf("DXGI 捕获显示器 %s 区域 %v 返回空图像", screenshotDisplayCaptureLabel(d), inter)
		}

		offsetX := inter.Min.X - rect.Min.X
		offsetY := inter.Min.Y - rect.Min.Y
		w := min(inter.Dx(), img2.Rect.Dx())
		h := min(inter.Dy(), img2.Rect.Dy())
		if w <= 0 || h <= 0 {
			continue
		}
		draw.DrawMask(img, image.Rect(offsetX, offsetY, offsetX+w, offsetY+h), img2, img2.Rect.Min, nil, image.Point{}, draw.Src)
		drawn++
	}
	if drawn == 0 {
		return nil, errors.New("DXGI 未捕获到任何已连接显示器")
	}
	return img, nil
}

// captureScreenshotDXGIDisplayRegionFrameWithFallback 捕获单显示器区域帧并 GDI 兜底。
// [S] ASM 0x14096ff00: inter=rect∩d.Bounds 空 → "DXGI 显示器区域为空"；
// SDR → debug 后 captureScreenshotScreenRectGDI(inter)；HDR 失败 → debug +
// GDI 兜底；GDI 再失败 → "DXGI 区域捕获失败: %v；GDI 区域兜底失败: %w"。
func captureScreenshotDXGIDisplayRegionFrameWithFallback(d screenshotDisplayCaptureInfo, rect image.Rectangle, requireFresh bool) (*image.RGBA, error) {
	inter := rect.Intersect(d.Bounds)
	if inter.Empty() {
		return nil, errors.New("DXGI 显示器区域为空")
	}

	if !d.HDR {
		screenshotHDRCaptureDebugLog("DXGI 跳过 SDR 显示器，使用 GDI 单屏区域捕获: display=%s rect=%v", screenshotDisplayCaptureLabel(d), inter)
		return captureScreenshotScreenRectGDI(inter.Min.X, inter.Min.Y, inter.Max.X, inter.Max.Y)
	}

	rel := inter.Sub(d.Bounds.Min)
	img, err := captureScreenshotDXGIOutputFrameRegionWithOptions(d, rel, requireFresh)
	if err == nil {
		return img, nil
	}
	screenshotHDRCaptureDebugLog("DXGI 捕获显示器 %s 区域 %v 失败，改用 GDI 区域兜底: %v", screenshotDisplayCaptureLabel(d), inter, err)
	img2, gdiErr := captureScreenshotScreenRectGDI(inter.Min.X, inter.Min.Y, inter.Max.X, inter.Max.Y)
	if gdiErr != nil {
		return nil, fmt.Errorf("DXGI 区域捕获失败: %v；GDI 区域兜底失败: %w", err, gdiErr)
	}
	return img2, nil
}

// captureScreenshotDXGIDisplayGDI 是单显示器 GDI 兜底（目标闭包
// captureScreenshotDXGIDisplayFrameWithFallback.func1）。
// [S] ASM 0x1409f5d60: 读取 display.Bounds 四个分量后调用
// captureScreenshotScreenRectGDI。
func captureScreenshotDXGIDisplayGDI(d screenshotDisplayCaptureInfo) (*image.RGBA, error) {
	b := d.Bounds
	return captureScreenshotScreenRectGDI(b.Min.X, b.Min.Y, b.Max.X, b.Max.Y)
}

// captureScreenshotDXGIDisplayFrameWithFallback 捕获单显示器整帧并 GDI 兜底。
// [S] ASM 0x14096fea0: 传 dxgiCapture=captureScreenshotDXGIOutputFrame、
// gdiFallback=captureScreenshotDXGIDisplayGDI 给 Using 变体。
func captureScreenshotDXGIDisplayFrameWithFallback(d screenshotDisplayCaptureInfo) (*image.RGBA, error) {
	return captureScreenshotDXGIDisplayFrameWithFallbackUsing(d, captureScreenshotDXGIOutputFrame, captureScreenshotDXGIDisplayGDI)
}

// captureScreenshotDXGIDisplayFrameWithFallbackUsing 是带兜底的单显示器帧捕获核心。
// [S] ASM 0x140970320: SDR → 检查 gdiFallback 后 debug + gdiFallback(d)；
// HDR → 检查 dxgiCapture 后调用；失败且无 gdiFallback →
// "DXGI 捕获失败且 GDI 兜底未初始化: %w"；否则 debug +
// gdiFallback(d)；GDI 再失败 → "DXGI 捕获失败: %v；GDI 兜底失败: %w"。
func captureScreenshotDXGIDisplayFrameWithFallbackUsing(d screenshotDisplayCaptureInfo, dxgiCapture, gdiFallback func(screenshotDisplayCaptureInfo) (*image.RGBA, error)) (*image.RGBA, error) {
	if !d.HDR {
		if gdiFallback == nil {
			return nil, errors.New("GDI 单屏兜底函数未初始化")
		}
		screenshotHDRCaptureDebugLog("DXGI 跳过 SDR 显示器，使用 GDI 单屏捕获: display=%s", screenshotDisplayCaptureLabel(d))
		return gdiFallback(d)
	}

	if dxgiCapture == nil {
		return nil, errors.New("DXGI 显示器捕获函数未初始化")
	}
	img, err := dxgiCapture(d)
	if err == nil {
		return img, nil
	}
	if gdiFallback == nil {
		return nil, fmt.Errorf("DXGI 捕获失败且 GDI 兜底未初始化: %w", err)
	}
	screenshotHDRCaptureDebugLog("DXGI 捕获显示器 %s 失败，改用 GDI 单屏兜底: %v", screenshotDisplayCaptureLabel(d), err)
	img, gdiErr := gdiFallback(d)
	if gdiErr != nil {
		return nil, fmt.Errorf("DXGI 捕获失败: %v；GDI 兜底失败: %w", err, gdiErr)
	}
	return img, nil
}

// captureScreenshotDXGIOutputFrame 捕获单显示器整帧（无强制刷新）。
// [S] ASM 0x140975fc0: 尾调用 captureScreenshotDXGIOutputFrameWithOptions(d, false)。
func captureScreenshotDXGIOutputFrame(d screenshotDisplayCaptureInfo) (*image.RGBA, error) {
	return captureScreenshotDXGIOutputFrameWithOptions(d, false)
}

// prewarmScreenshotHDRCaptureAsync 异步预热 DXGI HDR 捕获缓存。
// [S] ASM 0x14096e4e0: mode==disabled 或 !screenshotDXGIHDRCaptureAvailable()
// 直接返回；否则 go func1(mode)。
func prewarmScreenshotHDRCaptureAsync() {
	mode := resolveScreenshotHDRCaptureMode()
	if mode == "disabled" {
		return
	}
	if !screenshotDXGIHDRCaptureAvailable() {
		return
	}
	go prewarmScreenshotHDRCaptureWorker(mode)
}

// prewarmScreenshotHDRCaptureWorker 是预热 goroutine 体（目标闭包
// prewarmScreenshotHDRCaptureAsync.func1）。
// [S] ASM 0x14096e580: 枚举失败 → "DXGI HDR 预热枚举显示器失败: %v"；
// selectScreenshotScreenCaptureBackendName(displays, mode, true) != "dxgi-hdr"
// 返回；否则 prewarmScreenshotDXGIOutputFrameCache(displays)。
func prewarmScreenshotHDRCaptureWorker(mode string) {
	displays, err := enumerateScreenshotDisplayCaptureInfos()
	if err != nil {
		screenshotHDRCaptureDebugLog("DXGI HDR 预热枚举显示器失败: %v", err)
		return
	}
	if selectScreenshotScreenCaptureBackendName(displays, mode, true) != "dxgi-hdr" {
		return
	}
	prewarmScreenshotDXGIOutputFrameCache(displays)
}

// ---- 深层 D3D11 / 桌面复制 / 缓存入口已由批次 54 落地 ----
// captureScreenshotDXGIOutputFrameWithOptions /
// captureScreenshotDXGIOutputFrameRegionWithOptions /
// prewarmScreenshotDXGIOutputFrameCache 见
// screenshot_dxgi_cache_windows.go；深层桩见 screenshot_dxgi_deep_windows.go。
