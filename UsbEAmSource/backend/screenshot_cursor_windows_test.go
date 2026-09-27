package main

import (
	"image"
	"image/color"
	"testing"
	"unsafe"
)

func TestCopyQRCodeDIBBitsRoundTrip(t *testing.T) {
	src := image.NewRGBA(image.Rect(3, 5, 9, 11))
	for y := 3; y < 11; y++ {
		for x := 5; x < 9; x++ {
			src.SetRGBA(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: uint8(x * y), A: 0xAB})
		}
	}
	w := src.Rect.Dx()
	h := src.Rect.Dy()
	buf := make([]byte, w*h*4)
	copyRGBAToQRCodeDIBBits(src, unsafe.Pointer(&buf[0]))

	dst := image.NewRGBA(image.Rect(3, 5, 9, 11))
	copyQRCodeDIBBitsToRGBA(unsafe.Pointer(&buf[0]), dst)

	for y := 3; y < 11; y++ {
		for x := 5; x < 9; x++ {
			a := src.RGBAAt(x, y)
			b := dst.RGBAAt(x, y)
			if a != b {
				t.Fatalf("pixel (%d,%d) mismatch: got %v want %v", x, y, b, a)
			}
		}
	}
}

func TestCreateQRCodeRGBACompatibleBitmapNil(t *testing.T) {
	if _, _, err := createQRCodeRGBACompatibleBitmap(nil); err == nil {
		t.Fatal("expected error for nil image")
	}
	empty := image.NewRGBA(image.Rect(0, 0, 0, 0))
	if _, _, err := createQRCodeRGBACompatibleBitmap(empty); err == nil {
		t.Fatal("expected error for empty image")
	}
}
