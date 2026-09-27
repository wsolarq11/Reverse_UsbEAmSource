// AUTO-RECONSTRUCTED — screenshot_png_fast 回归测试
// 研究用途
package main

import (
	"bytes"
	"image"
	"image/png"
	"testing"
)

// TestEncodeScreenshotPNGOpaque：全 alpha=255 时像素应原样往返。
func TestEncodeScreenshotPNGOpaque(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 5, 3))
	for i := range src.Pix {
		src.Pix[i] = byte(i * 13)
	}
	for i := 3; i < len(src.Pix); i += 4 {
		src.Pix[i] = 0xff
	}
	encoded, err := encodeScreenshotRGBAWithKlauspostPNG((*screenshotRGBAImageRowSource)(src))
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := png.Decode(bytes.NewReader(encoded))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	nrgba := decoded.(*image.NRGBA)
	if nrgba.Bounds() != src.Bounds() {
		t.Fatalf("bounds: got %v want %v", nrgba.Bounds(), src.Bounds())
	}
	if !bytes.Equal(nrgba.Pix, src.Pix) {
		t.Fatalf("pixels mismatch")
	}
}

// TestEncodeScreenshotPNGAlpha：部分 alpha 应反预乘为 straight RGBA。
func TestEncodeScreenshotPNGAlpha(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 3, 1))
	// premultiplied 输入，alpha 分别为 0 / 100 / 255。
	src.Pix[0], src.Pix[1], src.Pix[2], src.Pix[3] = 0x00, 0x00, 0x00, 0x00
	src.Pix[4], src.Pix[5], src.Pix[6], src.Pix[7] = 0x32, 0x64, 0x96, 0x64 // 50,100,150 premul a=100
	src.Pix[8], src.Pix[9], src.Pix[10], src.Pix[11] = 0x40, 0x80, 0xc0, 0xff

	encoded, err := encodeScreenshotRGBAWithKlauspostPNG((*screenshotRGBAImageRowSource)(src))
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	decoded, err := png.Decode(bytes.NewReader(encoded))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	nrgba := decoded.(*image.NRGBA)

	want := []byte{
		0x00, 0x00, 0x00, 0x00,
		unpremulScreenshotPNG(0x32, 0x64), unpremulScreenshotPNG(0x64, 0x64), unpremulScreenshotPNG(0x96, 0x64), 0x64,
		0x40, 0x80, 0xc0, 0xff,
	}
	if !bytes.Equal(nrgba.Pix, want) {
		t.Fatalf("pixels: got %x want %x", nrgba.Pix, want)
	}
}

// TestWriteScreenshotPNGChunk：chunk 头/CRC 与标准 CRC32 IEEE 一致。
func TestWriteScreenshotPNGChunk(t *testing.T) {
	var buf bytes.Buffer
	err := writeScreenshotPNGChunk(&buf, []byte("hello"), [4]byte{'t', 'E', 's', 't'})
	if err != nil {
		t.Fatalf("chunk: %v", err)
	}
	want := []byte{0, 0, 0, 5, 't', 'E', 's', 't', 'h', 'e', 'l', 'l', 'o'}
	if !bytes.Equal(buf.Bytes()[:13], want) {
		t.Fatalf("header/data: got %x", buf.Bytes()[:13])
	}
	if len(buf.Bytes()) != 17 {
		t.Fatalf("len: got %d want 17", len(buf.Bytes()))
	}
}
