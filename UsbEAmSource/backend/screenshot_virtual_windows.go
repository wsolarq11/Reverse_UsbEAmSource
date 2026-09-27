package main

import (
	"errors"
	"fmt"
	"image"
)

// buildDarkenedQRCodeSelectionPreview 生成 42% 亮度的暗化预览图。
// [S] ASM 0x14095a200: nil → NewRGBA(零矩形)；否则 NewRGBA(img.Rect) +
// copy(Pix)；每像素 RGB 三通道 *0x2a 取低 16 位 *0x147af >> 0x17。
func buildDarkenedQRCodeSelectionPreview(img *image.RGBA) *image.RGBA {
	if img == nil {
		return image.NewRGBA(image.Rectangle{})
	}
	shaded := image.NewRGBA(img.Rect)
	copy(shaded.Pix, img.Pix)
	for i := 0; i+3 < len(shaded.Pix); i += 4 {
		shaded.Pix[i+0] = uint8((uint32(uint16(uint32(shaded.Pix[i+0])*0x2a)) * 0x147af) >> 0x17)
		shaded.Pix[i+1] = uint8((uint32(uint16(uint32(shaded.Pix[i+1])*0x2a)) * 0x147af) >> 0x17)
		shaded.Pix[i+2] = uint8((uint32(uint16(uint32(shaded.Pix[i+2])*0x2a)) * 0x147af) >> 0x17)
	}
	return shaded
}

// captureQRCodeVirtualScreenSnapshotGDIWithCursorSnapshot 捕获虚拟屏幕并叠加光标。
// [S] ASM 0x140971820: bounds 空 → "未检测到可用的屏幕区域"；GetDC(0) 失败 →
// "获取屏幕设备上下文失败"；CreateCompatibleDC/CreateDIBSection/SelectObject/
// BitBlt 失败分别 fmt.Errorf 包装；ok 时 drawScreenshotCursorSnapshotOnDC；
// buildRGBAFromDIBBits → buildDarkenedQRCodeSelectionPreview。
func captureQRCodeVirtualScreenSnapshotGDIWithCursorSnapshot(info screenshotCursorInfo, ok bool) (qrCodeScreenSnapshot, error) {
	bounds := qrCodeVirtualScreenBounds()
	if bounds.Empty() {
		return qrCodeScreenSnapshot{}, errors.New("未检测到可用的屏幕区域")
	}

	hdc, _, _ := procGetDC.Call(0)
	if hdc == 0 {
		return qrCodeScreenSnapshot{}, errors.New("获取屏幕设备上下文失败")
	}
	defer procReleaseDC.Call(0, hdc)

	memDC, err := createAppCompatibleDC()
	if err != nil {
		return qrCodeScreenSnapshot{}, fmt.Errorf("创建截图设备上下文失败: %w", err)
	}
	defer deleteAppDC(memDC)

	width := bounds.Dx()
	height := bounds.Dy()
	canvas, pBits, err := createAppIconCanvas(width, height)
	if err != nil {
		return qrCodeScreenSnapshot{}, fmt.Errorf("创建截图位图失败: %w", err)
	}
	defer deleteAppObject(canvas)

	oldObj, err := selectAppObject(memDC, canvas)
	if err != nil {
		return qrCodeScreenSnapshot{}, fmt.Errorf("选择截图位图失败: %w", err)
	}
	defer restoreAppObject(memDC, oldObj)

	if err := qrCodeBitBlt(memDC, 0, 0, uintptr(width), uintptr(height), hdc, uintptr(bounds.Min.X), uintptr(bounds.Min.Y), 0x40CC0020); err != nil {
		return qrCodeScreenSnapshot{}, fmt.Errorf("复制屏幕像素失败: %w", err)
	}

	if ok {
		drawScreenshotCursorSnapshotOnDC(memDC, pBits, bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y, info, ok)
	}

	original := buildRGBAFromDIBBits(pBits, width, height)
	shaded := buildDarkenedQRCodeSelectionPreview(original)
	return qrCodeScreenSnapshot{
		virtualBounds: bounds,
		original:      original,
		shaded:        shaded,
	}, nil
}

// captureQRCodeVirtualScreenSnapshotGDI 按需捕获光标并调用快照变体。
// [S] ASM 0x140971700: captureCursor → captureScreenshotCursorSnapshot；
// 否则零值 info + ok=false；返回快照（丢弃 error）。
func captureQRCodeVirtualScreenSnapshotGDI(captureCursor bool) qrCodeScreenSnapshot {
	info := screenshotCursorInfo{}
	ok := false
	if captureCursor {
		info, ok = captureScreenshotCursorSnapshot()
	}
	snapshot, _ := captureQRCodeVirtualScreenSnapshotGDIWithCursorSnapshot(info, ok)
	return snapshot
}

// selectScreenshotScreenCaptureBackendName 选择截图后端名。
// [S] ASM 0x14096e0a0: !preferHDR 或 "disabled" → "gdi"；"force" 命中首个
// 附着且边界有效的显示器 → "dxgi-hdr"；"auto" 额外要求 HDR；其余 "gdi"。
func selectScreenshotScreenCaptureBackendName(displays []screenshotDisplayCaptureInfo, mode string, preferHDR bool) string {
	if !preferHDR || mode == "disabled" {
		return "gdi"
	}
	if mode == "force" {
		for i := range displays {
			d := displays[i]
			if d.AttachedToDesktop && d.Bounds.Min.X < d.Bounds.Max.X && d.Bounds.Min.Y < d.Bounds.Max.Y {
				return "dxgi-hdr"
			}
		}
		return "gdi"
	}
	if mode == "auto" {
		for i := range displays {
			d := displays[i]
			if d.AttachedToDesktop && d.Bounds.Min.X < d.Bounds.Max.X && d.Bounds.Min.Y < d.Bounds.Max.Y && d.HDR {
				return "dxgi-hdr"
			}
		}
		return "gdi"
	}
	return "gdi"
}

// captureVirtualScreen 捕获虚拟屏幕快照并叠加光标。
// [S] ASM 0x14096e660: captureCursor && cursorSnapshot.ok → 快照变体
// （ok 传 true）；否则 captureQRCodeVirtualScreenSnapshotGDI(captureCursor)。
func (screenshotGDIScreenCaptureBackend) captureVirtualScreen(options screenshotScreenCaptureOptions) qrCodeScreenSnapshot {
	if options.captureCursor && options.cursorSnapshot.ok {
		snapshot, _ := captureQRCodeVirtualScreenSnapshotGDIWithCursorSnapshot(options.cursorSnapshot.info, true)
		return snapshot
	}
	return captureQRCodeVirtualScreenSnapshotGDI(options.captureCursor)
}
