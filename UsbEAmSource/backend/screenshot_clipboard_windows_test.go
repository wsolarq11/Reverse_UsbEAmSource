package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"syscall"
	"testing"
)

// 批次 100：clipboard 域纯逻辑函数测试（不触 Windows API）。

func TestWrapWinError(t *testing.T) {
	if got := wrapWinError("m", nil); got == nil || got.Error() != "m" {
		t.Fatalf("nil err: got %v", got)
	}
	if got := wrapWinError("m", syscall.Errno(0)); got == nil || got.Error() != "m" {
		t.Fatalf("Errno(0) 应等同 nil：got %v", got)
	}
	if got := wrapWinError("m", syscall.Errno(2)); got == nil || got.Error() == "m" {
		t.Fatalf("非零 errno 应包装：got %v", got)
	}
}

func TestGlobalAllocMoveableEmpty(t *testing.T) {
	if _, err := globalAllocMoveable(0); err == nil {
		t.Fatal("size=0 应报错且不触 GlobalAlloc")
	}
}

func TestBuildScreenshotClipboardDIBV5(t *testing.T) {
	// 2x2 全不透明图像，避免 PNG 透明像素编码歧义。
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})     // 红
	img.Set(1, 0, color.RGBA{0, 255, 0, 255})     // 绿
	img.Set(0, 1, color.RGBA{0, 0, 255, 255})     // 蓝
	img.Set(1, 1, color.RGBA{255, 255, 255, 255}) // 白
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}

	dib, err := buildScreenshotClipboardDIBV5(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(dib) != 124+2*2*4 {
		t.Fatalf("len=%d 期望 %d", len(dib), 124+16)
	}
	if binary.LittleEndian.Uint32(dib[0:]) != 124 {
		t.Fatalf("bV5Size=%d", binary.LittleEndian.Uint32(dib[0:]))
	}
	if binary.LittleEndian.Uint32(dib[4:]) != 2 {
		t.Fatalf("bV5Width=%d", binary.LittleEndian.Uint32(dib[4:]))
	}
	if binary.LittleEndian.Uint32(dib[8:]) != 2 {
		t.Fatalf("bV5Height=%d", binary.LittleEndian.Uint32(dib[8:]))
	}
	if got := binary.LittleEndian.Uint16(dib[12:]); got != 1 {
		t.Fatalf("bV5Planes=%d", got)
	}
	if got := binary.LittleEndian.Uint16(dib[14:]); got != 32 {
		t.Fatalf("bV5BitCount=%d", got)
	}
	if got := binary.LittleEndian.Uint32(dib[56:]); got != 0x57696e20 {
		t.Fatalf("bV5CSType=%#x", got)
	}

	// 自底向上：row0=图像底行(蓝,白)，row1=图像顶行(红,绿)，BGRA 顺序。
	want := [4][4]byte{
		{255, 0, 0, 255},     // 蓝 → B,G,R,A
		{255, 255, 255, 255}, // 白
		{0, 0, 255, 255},     // 红
		{0, 255, 0, 255},     // 绿
	}
	for i := 0; i < 4; i++ {
		off := 124 + i*4
		for j := 0; j < 4; j++ {
			if dib[off+j] != want[i][j] {
				t.Fatalf("像素 %d 字节 %d = %d 期望 %d（全图 %v）",
					i, j, dib[off+j], want[i][j], dib[124:])
			}
		}
	}
}
