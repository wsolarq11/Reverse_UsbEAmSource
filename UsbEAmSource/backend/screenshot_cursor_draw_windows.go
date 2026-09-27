package main

import (
	"image"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

// procGdiFlush 是 gdi32.GdiFlush 的 LazyProc。
var procGdiFlush = gdi32DLL.NewProc("GdiFlush")

// screenshotSystemCursorSize 返回系统光标标准尺寸（SM_CXCURSOR/SM_CYCURSOR）。
// [S] ASM 0x140974e00: GetSystemMetrics(0xd)→宽、GetSystemMetrics(0xe)→高。
func screenshotSystemCursorSize() (int, int) {
	w := getSystemMetrics(0xd)
	h := getSystemMetrics(0xe)
	return w, h
}

// screenshotCursorBaseSizeFromRegistry 从 HKCU 光标配置读取基准尺寸。
// [S] ASM 0x140974ca0: OpenKey(HKCU,"Control Panel\Cursors",QUERY_VALUE) →
// GetIntegerValue("CursorBaseSize")；错误/0/大于 0x200 返回 0。
func screenshotCursorBaseSizeFromRegistry() uint64 {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Control Panel\Cursors`, registry.QUERY_VALUE)
	if err != nil {
		return 0
	}
	defer k.Close()
	v, _, err := k.GetIntegerValue(`CursorBaseSize`)
	if err != nil || v == 0 || v > 0x200 {
		return 0
	}
	return v
}

// screenshotCursorDrawSize 按图标尺寸等比缩放光标绘制尺寸。
// [S] ASM 0x140974b80: max(基准,w,h,系统宽,系统高)；w<=0||h<=0||w==h → (max,max)；
// h<w → (max, max(1,h*max/w))；否则 (max(1,w*max/h), max)。
func screenshotCursorDrawSize(w, h int) (int, int) {
	sysW, sysH := screenshotSystemCursorSize()
	base := int64(screenshotCursorBaseSizeFromRegistry())
	max := int64(0)
	for _, v := range []int64{base, int64(w), int64(h), int64(sysW), int64(sysH)} {
		if v > max {
			max = v
		}
	}
	if max <= 0 {
		return 0, 0
	}
	if w <= 0 || h <= 0 || w == h {
		return int(max), int(max)
	}
	if h < w {
		v := int64(h) * max / int64(w)
		if v <= 0 {
			v = 1
		}
		return int(max), int(v)
	}
	v := int64(w) * max / int64(h)
	if v <= 0 {
		v = 1
	}
	return int(v), int(max)
}

// screenshotCursorIconMetrics 加载光标图标信息并计算绘制尺寸。
// [S] ASM 0x140974ac0: loadAppIconInfo → screenshotCursorDrawSize；加载失败或
// 绘制尺寸无效时释放图标并返回零值 + false。
func screenshotCursorIconMetrics(hicon uintptr) (*appIconInfo, int, int, int, int, bool) {
	info, w, h, err := loadAppIconInfo(hicon)
	if err != nil {
		return nil, 0, 0, 0, 0, false
	}
	drawW, drawH := screenshotCursorDrawSize(w, h)
	if drawW <= 0 || drawH <= 0 {
		releaseAppIconInfo(info)
		return nil, 0, 0, 0, 0, false
	}
	return info, drawW, drawH, w, h, true
}

// forceScreenshotDIBAlphaOpaqueInRect 将矩形交集内的 BGRA alpha 通道置为不透明。
// [S] ASM 0x140974ea0: nil/非正尺寸/空 rect → false；Intersect((x0,y0,x1,y1),
// (0,0,width,height)) 空 → false；逐像素 dst[(y*width+x)*4+3]=0xFF → true。
func forceScreenshotDIBAlphaOpaqueInRect(bits unsafe.Pointer, width, height int, x0, y0, x1, y1 int) bool {
	if bits == nil || width <= 0 || height <= 0 || x0 >= x1 || y0 >= y1 {
		return false
	}
	r := image.Rect(x0, y0, x1, y1).Intersect(image.Rect(0, 0, width, height))
	if r.Empty() {
		return false
	}
	dst := unsafe.Slice((*byte)(bits), width*height*4)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		row := y * width * 4
		for x := r.Min.X; x < r.Max.X; x++ {
			dst[row+x*4+3] = 0xFF
		}
	}
	return true
}

// drawScreenshotCursorInfoOnDC 将光标图标绘制到 DIB 兼容 DC 并强制 alpha 不透明。
// [S] ASM 0x140974080: 前置 ok&&Cursor!=0&&Flags&1 → screenshotCursorIconMetrics
// → defer releaseAppIconInfo → 热点等比缩放 → 目标坐标 = 光标屏幕位 - 起点 - 缩放热点
// → 与位图区域求交 → DrawIconEx(hdc,目标X,目标Y,hIcon,drawW,drawH,0,0,3) →
// forceScreenshotDIBAlphaOpaqueInRect(bits,width,height,交集)。
func drawScreenshotCursorInfoOnDC(hdc uintptr, bits unsafe.Pointer, x0, y0, x1, y1 int, info screenshotCursorInfo, ok bool) bool {
	if !ok || info.Cursor == 0 || info.Flags&0x1 == 0 {
		return false
	}
	iconInfo, drawW, drawH, w, h, ok := screenshotCursorIconMetrics(info.Cursor)
	if !ok {
		return false
	}
	defer releaseAppIconInfo(iconInfo)

	dx := int(int32(iconInfo.XHotspot))
	if dx > 0 && drawW > 0 && w > 0 && w != drawW {
		dx = dx * drawW / w
	}
	dy := int(int32(iconInfo.YHotspot))
	if dy > 0 && drawH > 0 && h > 0 && h != drawH {
		dy = dy * drawH / h
	}

	targetX := int(info.ScreenPos.X) - x0 - dx
	targetY := int(info.ScreenPos.Y) - y0 - dy

	r, _, _ := procDrawIconEx.Call(hdc, uintptr(targetX), uintptr(targetY), info.Cursor, uintptr(drawW), uintptr(drawH), 0, 0, 3)
	if r == 0 {
		return false
	}
	return forceScreenshotDIBAlphaOpaqueInRect(bits, x1-x0, y1-y0, targetX, targetY, targetX+drawW, targetY+drawH)
}

// drawScreenshotCursorSnapshotOnDC 校验后委托 drawScreenshotCursorInfoOnDC。
// [S] ASM 0x140973fc0: hdc/bits 零或空 rect → false；否则透传 info+ok。
func drawScreenshotCursorSnapshotOnDC(hdc uintptr, bits unsafe.Pointer, x0, y0, x1, y1 int, info screenshotCursorInfo, ok bool) bool {
	if hdc == 0 || bits == nil || x1 <= x0 || y1 <= y0 {
		return false
	}
	return drawScreenshotCursorInfoOnDC(hdc, bits, x0, y0, x1, y1, info, ok)
}

// drawScreenshotCursorSnapshotOnRGBA 将光标快照绘制到 *image.RGBA。
// [S] ASM 0x140974560: nil/空/尺寸不匹配 → false；createQRCodeRGBACompatibleBitmap
// → defer deleteAppObject → createAppCompatibleDC → defer deleteAppDC →
// selectAppObject → defer restoreAppObject → drawScreenshotCursorSnapshotOnDC →
// GdiFlush → copyQRCodeDIBBitsToRGBA。
func drawScreenshotCursorSnapshotOnRGBA(img *image.RGBA, x0, y0, x1, y1 int, info screenshotCursorInfo, ok bool) bool {
	if img == nil || x0 >= x1 || y0 >= y1 {
		return false
	}
	if img.Rect.Dx() != x1-x0 || img.Rect.Dy() != y1-y0 {
		return false
	}
	bitmap, bits, err := createQRCodeRGBACompatibleBitmap(img)
	if err != nil {
		return false
	}
	defer deleteAppObject(bitmap)

	memDC, err := createAppCompatibleDC()
	if err != nil {
		return false
	}
	defer deleteAppDC(memDC)

	oldObj, err := selectAppObject(memDC, bitmap)
	if err != nil {
		return false
	}
	defer restoreAppObject(memDC, oldObj)

	if !drawScreenshotCursorSnapshotOnDC(memDC, bits, x0, y0, x1, y1, info, ok) {
		return false
	}
	procGdiFlush.Call()
	copyQRCodeDIBBitsToRGBA(bits, img)
	return true
}

// drawScreenshotCursorOnRGBA 现场捕获光标并绘制。
// [S] ASM 0x1409744a0: captureScreenshotCursorSnapshot → snapshot 变体。
func drawScreenshotCursorOnRGBA(img *image.RGBA, x0, y0, x1, y1 int) bool {
	info, ok := captureScreenshotCursorSnapshot()
	return drawScreenshotCursorSnapshotOnRGBA(img, x0, y0, x1, y1, info, ok)
}

// drawScreenshotCaptureCursorOnRGBA 按选项绘制已捕获或现场捕获的光标。
// [S] ASM 0x14096f100: !captureCursor → false；!cursor.ok → 现场捕获；
// 否则快照变体（ok 传 true）。
func drawScreenshotCaptureCursorOnRGBA(img *image.RGBA, x0, y0, x1, y1 int, captureCursor bool, cursor screenshotCursorSnapshot) bool {
	if !captureCursor {
		return false
	}
	if !cursor.ok {
		return drawScreenshotCursorOnRGBA(img, x0, y0, x1, y1)
	}
	return drawScreenshotCursorSnapshotOnRGBA(img, x0, y0, x1, y1, cursor.info, true)
}

// captureScreenRect 捕获矩形区域并叠加光标。
// [S] ASM 0x14096e840: captureScreenshotScreenRectGDI → 错误透传 →
// normalizeScreenshotScrollingRect → drawScreenshotCaptureCursorOnRGBA（返回值忽略）
// → 返回 (img, nil)。
func (screenshotGDIScreenCaptureBackend) captureScreenRect(rect image.Rectangle, options screenshotScreenCaptureOptions) (*image.RGBA, error) {
	img, err := captureScreenshotScreenRectGDI(rect.Min.X, rect.Min.Y, rect.Max.X, rect.Max.Y)
	if err != nil {
		return nil, err
	}
	norm := normalizeScreenshotScrollingRect(rect)
	drawScreenshotCaptureCursorOnRGBA(img, norm.Min.X, norm.Min.Y, norm.Max.X, norm.Max.Y, options.captureCursor, options.cursorSnapshot)
	return img, nil
}
