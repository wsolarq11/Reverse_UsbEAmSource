package main

import (
	"errors"
	"fmt"
	"image"
	"syscall"
)

// GDI 屏幕捕获所需的 user32/gdi32 LazyProc（依赖 appicon_windows.go 中的
// user32DLL / gdi32DLL 全局 DLL 句柄）。
var (
	procGetDC            = user32DLL.NewProc("GetDC")
	procReleaseDC        = user32DLL.NewProc("ReleaseDC")
	procBitBlt           = gdi32DLL.NewProc("BitBlt")
	procGetSystemMetrics = user32DLL.NewProc("GetSystemMetrics")
)

// getSystemMetrics 调用 user32.GetSystemMetrics 并将 uintptr 返回值按有符号
// 整数解释（asm 0x14095a060 的 jle/jg 分支证明返回值参与有符号比较）。
// [S] ASM 0x14095a060: GetSystemMetrics LazyProc(1 arg)。
func getSystemMetrics(index int) int {
	r, _, _ := procGetSystemMetrics.Call(uintptr(index))
	return int(r)
}

// qrCodeBitBlt 包装 GDI BitBlt。
// [S] ASM 0x14095ebe0: dstW<=0 || dstH<=0 → nil；否则 LazyProc.Call(9 args)。
// r1!=0 → nil；lastErr==nil || errors.Is(lastErr, syscall.Errno(0)) →
// "BitBlt 失败"；否则返回 lastErr。
func qrCodeBitBlt(dstHDC, dstX, dstY, dstW, dstH, srcHDC, srcX, srcY, rop uintptr) error {
	if dstW <= 0 || dstH <= 0 {
		return nil
	}
	r1, _, lastErr := procBitBlt.Call(dstHDC, dstX, dstY, dstW, dstH, srcHDC, srcX, srcY, rop)
	if r1 != 0 {
		return nil
	}
	if lastErr == nil || errors.Is(lastErr, syscall.Errno(0)) {
		return errors.New("BitBlt 失败")
	}
	return lastErr
}

// qrCodeVirtualScreenBounds 返回虚拟屏幕矩形。
// [S] ASM 0x14095a060: GetSystemMetrics(SM_X/Y/CX/CYVIRTUALSCREEN)；宽或高
// <=0 时回退到 (SM_XSCREEN, SM_YSCREEN, 0, 0)；最后对 (x,y,x+w,y+h) 做
// min/max 规范化。
func qrCodeVirtualScreenBounds() image.Rectangle {
	x := getSystemMetrics(0x4c) // SM_XVIRTUALSCREEN
	y := getSystemMetrics(0x4d) // SM_YVIRTUALSCREEN
	w := getSystemMetrics(0x4e) // SM_CXVIRTUALSCREEN
	h := getSystemMetrics(0x4f) // SM_CYVIRTUALSCREEN
	if w <= 0 || h <= 0 {
		x = getSystemMetrics(0x0) // SM_XSCREEN
		y = getSystemMetrics(0x1) // SM_YSCREEN
		w = 0
		h = 0
	}
	x0 := x
	x1 := x + w
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	y0 := y
	y1 := y + h
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	return image.Rect(x0, y0, x1, y1)
}

// normalizeScreenshotScrollingRect 规范化滚动截图矩形：先按轴排序 Min/Max，
// 空矩形返回零矩形，否则与虚拟屏幕边界求交。
// [S] ASM 0x1409a2c60: cmovg 交换各轴；Min.X>=Max.X || Max.Y<=Min.Y →
// 零矩形；否则 image.Rectangle.Intersect(qrCodeVirtualScreenBounds)。
func normalizeScreenshotScrollingRect(rect image.Rectangle) image.Rectangle {
	if rect.Min.X > rect.Max.X {
		rect.Min.X, rect.Max.X = rect.Max.X, rect.Min.X
	}
	if rect.Min.Y > rect.Max.Y {
		rect.Min.Y, rect.Max.Y = rect.Max.Y, rect.Min.Y
	}
	if rect.Empty() {
		return image.Rectangle{}
	}
	return rect.Intersect(qrCodeVirtualScreenBounds())
}

// captureScreenshotScreenRectGDI 用 GDI BitBlt 捕获屏幕矩形。
// [S] ASM 0x1409722e0: normalize → GetDC(0) → CreateCompatibleDC →
// CreateDIBSection → SelectObject → BitBlt → BGRA→RGBA。
func captureScreenshotScreenRectGDI(x0, y0, x1, y1 int) (*image.RGBA, error) {
	rect := normalizeScreenshotScrollingRect(image.Rect(x0, y0, x1, y1))
	if rect.Empty() {
		return nil, errors.New("截图区域为空")
	}

	hdc, _, _ := procGetDC.Call(0)
	if hdc == 0 {
		return nil, errors.New("获取屏幕设备上下文失败")
	}
	defer procReleaseDC.Call(0, hdc)

	memDC, err := createAppCompatibleDC()
	if err != nil {
		return nil, fmt.Errorf("创建截图设备上下文失败: %w", err)
	}
	defer deleteAppDC(memDC)

	width := rect.Dx()
	height := rect.Dy()
	canvas, pBits, err := createAppIconCanvas(width, height)
	if err != nil {
		return nil, fmt.Errorf("创建长截图位图失败: %w", err)
	}
	defer deleteAppObject(canvas)

	oldObj, err := selectAppObject(memDC, canvas)
	if err != nil {
		return nil, fmt.Errorf("选择长截图位图失败: %w", err)
	}
	defer restoreAppObject(memDC, oldObj)

	if err := qrCodeBitBlt(memDC, 0, 0, uintptr(width), uintptr(height), hdc, uintptr(rect.Min.X), uintptr(rect.Min.Y), 0x40CC0020); err != nil {
		return nil, fmt.Errorf("复制长截图像素失败: %w", err)
	}

	return buildRGBAFromDIBBits(pBits, width, height), nil
}

// name 返回后端名称。
// [S] selectScreenshotScreenCaptureBackendName 0x14096e0a0 的 "disabled"
// 分支返回 rodata "gdi"（3 字节）。
func (screenshotGDIScreenCaptureBackend) name() string {
	return "gdi"
}
