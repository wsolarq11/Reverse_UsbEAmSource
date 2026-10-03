package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"syscall"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/w32"
)

// procGetCursorInfo 是 user32.GetCursorInfo 的 LazyProc。
var procGetCursorInfo = user32DLL.NewProc("GetCursorInfo")

// procShowCursor 是 user32.ShowCursor 的 LazyProc。
var procShowCursor = user32DLL.NewProc("ShowCursor")

// procLoadCursorW 是 user32.LoadCursorW 的 LazyProc。
var procLoadCursorW = user32DLL.NewProc("LoadCursorW")

// procSetCursor 是 user32.SetCursor 的 LazyProc。
var procSetCursor = user32DLL.NewProc("SetCursor")

// procReleaseCapture 是 user32.ReleaseCapture 的 LazyProc。
var procReleaseCapture = user32DLL.NewProc("ReleaseCapture")

// screenshotCursorNativeInfo 对应 Win32 CURSORINFO（24B）。
// [S] currentScreenshotCursorInfo 0x140974a20: newobject 24B，cbSize=0x18，
// 字段偏移 cbSize@0/flags@4/hCursor@8/ptScreenPos@0x10。
type screenshotCursorNativeInfo struct {
	cbSize      uint32
	flags       uint32
	hCursor     uintptr
	ptScreenPos w32.POINT
}

// currentScreenshotCursorInfo 调用 GetCursorInfo 返回当前光标信息。
// [S] ASM 0x140974a20: cbSize=0x18；GetCursorInfo(&info)；返回
// (Size,Flags,Cursor,ScreenPos.X,ScreenPos.Y, r!=0)。
func currentScreenshotCursorInfo() (screenshotCursorInfo, bool) {
	var info screenshotCursorNativeInfo
	info.cbSize = 0x18
	r, _, _ := procGetCursorInfo.Call(uintptr(unsafe.Pointer(&info)))
	return screenshotCursorInfo{
		Size:      info.cbSize,
		Flags:     info.flags,
		Cursor:    info.hCursor,
		ScreenPos: info.ptScreenPos,
	}, r != 0
}

// captureScreenshotCursorSnapshot 仅当光标可见且句柄有效时返回光标快照。
// [S] ASM 0x140973f80: ok && Cursor!=0 && Flags&1(CURSOR_SHOWING) → info；
// 否则零值 + false。
func captureScreenshotCursorSnapshot() (screenshotCursorInfo, bool) {
	info, ok := currentScreenshotCursorInfo()
	if !ok || info.Cursor == 0 || info.Flags&0x1 == 0 {
		return screenshotCursorInfo{}, false
	}
	return info, true
}

// screenshotCursorCurrentlyShowing 判断光标当前是否可见。
// [S] ASM 0x140973c40: cbSize=0x18；GetCursorInfo；r==0(失败)→true（保守可见）；
// 否则 flags&1(CURSOR_SHOWING)。
func screenshotCursorCurrentlyShowing() bool {
	var info screenshotCursorNativeInfo
	info.cbSize = 0x18
	r, _, _ := procGetCursorInfo.Call(uintptr(unsafe.Pointer(&info)))
	if r == 0 {
		return true
	}
	return info.flags&0x1 != 0
}

// adjustScreenshotCursorVisibility 循环调用 ShowCursor 直至光标达到目标可见状态（上限 64 次）。
// [S] ASM 0x140973e00: 单参 visible bool，返回 error。循环内 ShowCursor(visible)；
// visible && r>=0 或 !visible && r<0 → nil；err 非零且非 Errno(0) → fmt.Errorf 带 %w；
// 64 次未达成 → fmt.Errorf 无参。
func adjustScreenshotCursorVisibility(visible bool) error {
	var v uintptr
	if visible {
		v = 1
	}
	for i := 0; i < 64; i++ {
		r, _, err := procShowCursor.Call(v)
		if visible {
			if int32(r) >= 0 {
				return nil
			}
		} else if int32(r) < 0 {
			return nil
		}
		if err != nil && err != syscall.Errno(0) {
			return fmt.Errorf("恢复截图覆盖层鼠标指针失败: %w", err)
		}
	}
	return fmt.Errorf("恢复截图覆盖层鼠标指针失败")
}

// ensureScreenshotCursorVisible 确保光标可见（隐藏时调 ShowCursor(true)）。
// [S] ASM 0x140973ba0: 无参无返回。screenshotCursorCurrentlyShowing() 为真则返回；
// 否则 adjustScreenshotCursorVisibility(true)。
func ensureScreenshotCursorVisible() {
	if screenshotCursorCurrentlyShowing() {
		return
	}
	_ = adjustScreenshotCursorVisibility(true)
}

// setScreenshotOverlayCursor 加载系统光标资源（MAKEINTRESOURCE）并设置为当前光标。
// [S 汇编 0x140973ae0, 192B]：ensureScreenshotCursorVisible()；
// LoadCursorW(0, uint16) 返回 0 → false；否则 SetCursor(hCursor) → true。
func setScreenshotOverlayCursor(cursor uint16) bool {
	ensureScreenshotCursorVisible()
	hCursor, _, _ := procLoadCursorW.Call(0, uintptr(cursor))
	if hCursor == 0 {
		return false
	}
	procSetCursor.Call(hCursor)
	return true
}

// restoreScreenshotCursorAfterOverlay 覆盖层关闭后恢复光标：ReleaseCapture → 恢复箭头光标
// （IDC_ARROW=0x7f00）→ 光标仍隐藏则调整到可见。
// [S 汇编 0x140973be0, 96B]：ReleaseCapture()（0 参）；setScreenshotOverlayCursor(0x7f00)；
// !screenshotCursorCurrentlyShowing() → adjustScreenshotCursorVisibility(true)。
func restoreScreenshotCursorAfterOverlay() {
	procReleaseCapture.Call()
	setScreenshotOverlayCursor(0x7f00)
	if !screenshotCursorCurrentlyShowing() {
		_ = adjustScreenshotCursorVisibility(true)
	}
}

// setScreenshotCursorPosition 移动鼠标到屏幕坐标 (x,y)。
// [S 汇编 0x1409a2e20, 256B]：SetCursorPos(x,y) 返回非 0 → nil；
// 否则 err 为 nil 或 errors.Is(err, syscall.Errno(0)) → "设置鼠标位置失败"，其余透传 err。
func setScreenshotCursorPosition(x, y int) error {
	r1, _, err := procSetCursorPos.Call(uintptr(x), uintptr(y))
	if r1 != 0 {
		return nil
	}
	if err == nil || errors.Is(err, syscall.Errno(0)) {
		return errors.New("设置鼠标位置失败")
	}
	return err
}

// copyRGBAToQRCodeDIBBits 将 *image.RGBA 拷贝为紧凑 32bpp BGRA DIB 位。
// [S] ASM 0x140957760: 逐像素 RGBA→BGRA；目标 stride=width*4；源偏移
// =img.Stride*(y-Min.Y)+(x-Min.X)*4；nil 参数直接返回。
func copyRGBAToQRCodeDIBBits(img *image.RGBA, pBits unsafe.Pointer) {
	if img == nil || pBits == nil {
		return
	}
	w := img.Rect.Dx()
	h := img.Rect.Dy()
	dst := unsafe.Slice((*byte)(pBits), w*h*4)
	for y := 0; y < h; y++ {
		srcRow := img.PixOffset(img.Rect.Min.X, img.Rect.Min.Y+y)
		dstRow := y * w * 4
		for x := 0; x < w; x++ {
			s := srcRow + x*4
			d := dstRow + x*4
			dst[d+0] = img.Pix[s+2]
			dst[d+1] = img.Pix[s+1]
			dst[d+2] = img.Pix[s+0]
			dst[d+3] = img.Pix[s+3]
		}
	}
}

// copyQRCodeDIBBitsToRGBA 将紧凑 32bpp BGRA DIB 位写回 *image.RGBA。
// [S] ASM 0x140957980: 逐像素 BGRA→RGBA（SetRGBA 到 Min.X+x/Min.Y+y）；
// 源 stride=width*4；nil 参数直接返回。
func copyQRCodeDIBBitsToRGBA(pBits unsafe.Pointer, img *image.RGBA) {
	if pBits == nil || img == nil {
		return
	}
	w := img.Rect.Dx()
	h := img.Rect.Dy()
	src := unsafe.Slice((*byte)(pBits), w*h*4)
	for y := 0; y < h; y++ {
		srcRow := y * w * 4
		for x := 0; x < w; x++ {
			s := srcRow + x*4
			img.SetRGBA(img.Rect.Min.X+x, img.Rect.Min.Y+y, color.RGBA{
				R: src[s+2],
				G: src[s+1],
				B: src[s+0],
				A: src[s+3],
			})
		}
	}
}

// createQRCodeRGBACompatibleBitmap 创建与图像尺寸一致的 32bpp 自顶向下
// DIB Section 并填充 BGRA 像素。
// [S] ASM 0x140957520: img nil 或空 → "创建截图标注文字画布失败: 图像为空"；
// CreateDIBSection(0,&bmi,0,&pBits,0,0) 失败 → "创建截图标注文字画布失败"；
// 成功后 copyRGBAToQRCodeDIBBits(img, pBits)。
func createQRCodeRGBACompatibleBitmap(img *image.RGBA) (uintptr, unsafe.Pointer, error) {
	if img == nil || img.Rect.Min.X >= img.Rect.Max.X || img.Rect.Min.Y >= img.Rect.Max.Y {
		return 0, nil, errors.New("创建截图标注文字画布失败: 图像为空")
	}
	width := img.Rect.Dx()
	height := img.Rect.Dy()
	var bmi winBitmapInfo
	bmi.bmiHeader.biSize = uint32(unsafe.Sizeof(bmi.bmiHeader))
	bmi.bmiHeader.biWidth = int32(width)
	bmi.bmiHeader.biHeight = int32(-height)
	bmi.bmiHeader.biPlanes = 1
	bmi.bmiHeader.biBitCount = 32
	bmi.bmiHeader.biCompression = 0
	bmi.bmiHeader.biSizeImage = uint32(width * height * 4)
	var pBits unsafe.Pointer
	hbmp, _, _ := procCreateDIBSection.Call(0, uintptr(unsafe.Pointer(&bmi)), 0, uintptr(unsafe.Pointer(&pBits)), 0, 0)
	if hbmp == 0 || pBits == nil {
		return 0, nil, errors.New("创建截图标注文字画布失败")
	}
	copyRGBAToQRCodeDIBBits(img, pBits)
	return hbmp, pBits, nil
}
