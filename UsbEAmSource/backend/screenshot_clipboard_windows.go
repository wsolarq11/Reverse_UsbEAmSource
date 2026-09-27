// AUTO-RECONSTRUCTED — Screenshot clipboard Windows helpers (empirical asm)
// 研究用途
package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/w32"
	"golang.org/x/sys/windows"
)

// procRegisterClipboardFormatW 注册自定义剪贴板格式（复用 appicon_windows.go 的 user32DLL）。
var procRegisterClipboardFormatW = user32DLL.NewProc("RegisterClipboardFormatW")

// wrapWinError 将 Win32 错误码包装为带上下文消息的 error。
// [S 汇编 0x1408af3a0, 352B] err==nil 或 err==Errno(0)(ERROR_SUCCESS) 时仅返回
// errors.New(msg)，否则 fmt.Errorf("%s: %w", msg, err)（格式串 "%s: %w"）。
// 共享 helper：qrcode / clipboard / 文件操作等域复用。
func wrapWinError(msg string, err error) error {
	if err == nil {
		return errors.New(msg)
	}
	// ERROR_SUCCESS(0)：windows.Errno 即 syscall.Errno 的别名（x/sys/windows aliases.go）。
	if err == syscall.Errno(0) {
		return errors.New(msg)
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// globalAllocMoveable 用 GMEM_MOVEABLE(0x2) 分配全局堆内存。
// [S 汇编 0x1408ae9e0, 224B] size==0 → "剪贴板数据不能为空"；
// GlobalAlloc 返回 0 → wrapWinError("分配剪贴板内存失败")。
func globalAllocMoveable(size int) (w32.HGLOBAL, error) {
	if size == 0 {
		return 0, errors.New("剪贴板数据不能为空")
	}
	h := w32.GlobalAlloc(gmemMoveable, uint32(size))
	if h == 0 {
		return 0, wrapWinError("分配剪贴板内存失败", windows.GetLastError())
	}
	return h, nil
}

// globalLock 锁定全局堆内存并返回可写指针。
// [S 汇编 0x1408aeac0, 160B] GlobalLock 返回 0 → wrapWinError("锁定剪贴板内存失败")。
func globalLock(hMem w32.HGLOBAL) (unsafe.Pointer, error) {
	p := w32.GlobalLock(hMem)
	if p == nil {
		return nil, wrapWinError("锁定剪贴板内存失败", windows.GetLastError())
	}
	return p, nil
}

// registerScreenshotPNGClipboardFormat 注册 "PNG" 剪贴板格式。
// [S 汇编 0x140972ee0, 224B] UTF16FromString("PNG") → RegisterClipboardFormatW；
// 返回 0 → wrapWinError("注册 PNG 剪贴板格式失败")。
func registerScreenshotPNGClipboardFormat() (uint32, error) {
	name, err := windows.UTF16FromString("PNG")
	if err != nil {
		return 0, err
	}
	r, _, _ := procRegisterClipboardFormatW.Call(uintptr(unsafe.Pointer(&name[0])))
	if r == 0 {
		return 0, wrapWinError("注册 PNG 剪贴板格式失败", windows.GetLastError())
	}
	return uint32(r), nil
}

// openScreenshotClipboard 打开剪贴板，失败时以 5ms 间隔重试至多 40 次。
// [S 汇编 0x140972fc0, 224B] 40 次循环(每次失败 Sleep 5ms)后额外尝试一次，
// 仍失败 → wrapWinError("打开截图剪贴板失败")。
func openScreenshotClipboard() error {
	for i := 0; i < 40; i++ {
		if w32.OpenClipboard(0) {
			return nil
		}
		time.Sleep(5 * time.Millisecond)
	}
	if w32.OpenClipboard(0) {
		return nil
	}
	return wrapWinError("打开截图剪贴板失败", windows.GetLastError())
}

// setScreenshotClipboardBytes 将字节写入剪贴板指定格式。
// [S 汇编 0x1409730a0, 640B] 空数据 → "截图剪贴板数据不能为空"；
// globalAllocMoveable → globalLock → copy → GlobalUnlock → SetClipboardData。
func setScreenshotClipboardBytes(format uint32, data []byte) error {
	if len(data) == 0 {
		return errors.New("截图剪贴板数据不能为空")
	}
	hMem, err := globalAllocMoveable(len(data))
	if err != nil {
		return err
	}
	pMem, err := globalLock(hMem)
	if err != nil {
		return err
	}
	copy(unsafe.Slice((*byte)(pMem), len(data)), data)
	w32.GlobalUnlock(hMem)
	if w32.SetClipboardData(uint(format), w32.HANDLE(hMem)) == 0 {
		return wrapWinError("设置截图剪贴板格式失败", windows.GetLastError())
	}
	return nil
}

// buildScreenshotClipboardDIBV5 将 PNG 解码并构建 CF_DIBV5 字节。
// [S 汇编 0x140973380, 1888B] decodeScreenshotImageBytesWithBudget(png,"png",12) →
// 校验尺寸/像素字节上限 → 构造 124B BITMAPV5HEADER + BGRA 像素(自底向上)。
func buildScreenshotClipboardDIBV5(png []byte) ([]byte, error) {
	img, _, release, err := decodeScreenshotImageBytesWithBudget(png, "png", 12)
	if err != nil {
		return nil, fmt.Errorf("解析截图 PNG 失败: %w", err)
	}
	defer release()

	b := img.Bounds()
	w := b.Dx()
	h := b.Dy()
	if w <= 0 || h <= 0 {
		return nil, errors.New("截图尺寸不能为空")
	}
	if w > 0x7fffffff || h > 0x7fffffff {
		return nil, errors.New("截图尺寸超出 DIBV5 范围")
	}
	pixelBytes := uint64(w) * uint64(h) * 4
	if pixelBytes > 0xffffffff {
		return nil, errors.New("截图像素数据超出 DIBV5 范围")
	}

	buf := make([]byte, 124+int(pixelBytes))
	// BITMAPV5HEADER（124 字节）
	binary.LittleEndian.PutUint32(buf[0:], 124)         // bV5Size
	binary.LittleEndian.PutUint32(buf[4:], uint32(w))   // bV5Width
	binary.LittleEndian.PutUint32(buf[8:], uint32(h))   // bV5Height（正值 = 自底向上）
	binary.LittleEndian.PutUint16(buf[12:], 1)          // bV5Planes
	binary.LittleEndian.PutUint16(buf[14:], 32)         // bV5BitCount
	binary.LittleEndian.PutUint32(buf[40:], 0x00ff0000) // bV5RedMask
	binary.LittleEndian.PutUint32(buf[44:], 0x0000ff00) // bV5GreenMask
	binary.LittleEndian.PutUint32(buf[48:], 0x000000ff) // bV5BlueMask
	binary.LittleEndian.PutUint32(buf[52:], 0xff000000) // bV5AlphaMask
	binary.LittleEndian.PutUint32(buf[56:], 0x57696e20) // bV5CSType = 'Win '
	binary.LittleEndian.PutUint32(buf[108:], 4)         // bV5Intent = LCS_GM_GRAPHICS

	// BGRA 像素，自底向上：At 的 y 取 Max.Y-1-y 翻转，写入顺序 B,G,R,A(右移 8 位)。
	for y := 0; y < h; y++ {
		srcY := b.Max.Y - 1 - y
		for x := 0; x < w; x++ {
			r, g, bl, a := img.At(b.Min.X+x, srcY).RGBA()
			off := 124 + (y*w+x)*4
			buf[off+0] = byte(bl >> 8)
			buf[off+1] = byte(g >> 8)
			buf[off+2] = byte(r >> 8)
			buf[off+3] = byte(a >> 8)
		}
	}
	return buf, nil
}
