// appicon_windows_test.go — 图标 GDI 层测试（纯函数部分，无需 HWND/HDC）
// 研究用途。覆盖：buildRGBAFromDIBBits → applyAppIconMaskAlpha → appIconHasVisibleAlpha →
// visibleIconBounds → fitIconToSquare → normalizeAppIconPNGWithSize

package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
	"unsafe"
)

// ========================================================================
// buildRGBAFromDIBBits
// ========================================================================

func TestBuildRGBAFromDIBBits_Valid(t *testing.T) {
	w, h := 4, 3
	dib := make([]byte, w*h*4)
	// 填充 BGRA 数据（{B, G, R, A}）
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*w*4 + x*4
			dib[idx] = byte(x * 32)       // B
			dib[idx+1] = byte(y * 32)     // G
			dib[idx+2] = byte(128 + x*16) // R
			dib[idx+3] = 255              // A
		}
	}

	img := buildRGBAFromDIBBits(unsafe.Pointer(&dib[0]), w, h)
	if img == nil {
		t.Fatal("expected non-nil image")
	}
	bounds := img.Bounds()
	if bounds.Dx() != w || bounds.Dy() != h {
		t.Fatalf("expected %dx%d, got %dx%d", w, h, bounds.Dx(), bounds.Dy())
	}
	// 角像素 (0,0)：B=0,G=0,R=128,A=255
	got := img.RGBAAt(0, 0)
	if got.R != 128 || got.G != 0 || got.B != 0 || got.A != 255 {
		t.Fatalf("unexpected pixel (0,0): %v", got)
	}
}

func TestBuildRGBAFromDIBBits_NilData(t *testing.T) {
	img := buildRGBAFromDIBBits(nil, 4, 4)
	if img == nil {
		t.Fatal("expected non-nil empty image")
	}
	if img.Bounds().Dx() != 4 || img.Bounds().Dy() != 4 {
		t.Fatalf("expected 4x4 empty image, got %v", img.Bounds())
	}
}

func TestBuildRGBAFromDIBBits_InvalidDim(t *testing.T) {
	img := buildRGBAFromDIBBits(unsafe.Pointer(&[]byte{0, 0, 0, 0}[0]), 0, 4)
	if img == nil {
		t.Fatal("expected non-nil empty image")
	}
}

// ========================================================================
// appIconHasVisibleAlpha（A != 0 判定）
// ========================================================================

func TestAppIconHasVisibleAlpha_True(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.SetRGBA(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 128})
	if !appIconHasVisibleAlpha(img) {
		t.Fatal("expected visible alpha (semi-transparent)")
	}
}

func TestAppIconHasVisibleAlpha_True_Opaque(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.SetRGBA(0, 0, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if !appIconHasVisibleAlpha(img) {
		t.Fatal("expected visible alpha (A != 0)")
	}
}

func TestAppIconHasVisibleAlpha_False_FullyTransparent(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.SetRGBA(0, 0, color.RGBA{R: 0, G: 0, B: 0, A: 0})
	if appIconHasVisibleAlpha(img) {
		t.Fatal("expected no visible alpha (A == 0)")
	}
}

func TestAppIconHasVisibleAlpha_Nil(t *testing.T) {
	if appIconHasVisibleAlpha(nil) {
		t.Fatal("nil should return false")
	}
}

// ========================================================================
// visibleIconBounds（返回 (Rect, bool)；nil panic 对齐 asm）
// ========================================================================

func TestVisibleIconBounds_NonEmpty(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	// 仅右下有像素
	img.SetRGBA(7, 8, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	bounds, ok := visibleIconBounds(img)
	if !ok {
		t.Fatal("expected ok")
	}
	if bounds.Min.X != 7 || bounds.Min.Y != 8 || bounds.Max.X != 8 || bounds.Max.Y != 9 {
		t.Fatalf("unexpected bounds: %v", bounds)
	}
}

func TestVisibleIconBounds_FullyTransparent(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	_, ok := visibleIconBounds(img)
	if ok {
		t.Fatal("expected not-ok for transparent image")
	}
}

func TestVisibleIconBounds_Nil(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on nil image")
		}
	}()
	visibleIconBounds(nil)
}

// ========================================================================
// fitIconToSquare（最近邻缩放，非居中平铺）
// ========================================================================

func TestFitIconToSquare_Scale(t *testing.T) {
	// 源 4x6：maxDim=6，scale=8/6，newW=5，newH=8，offsetX=1，offsetY=0
	src := image.NewRGBA(image.Rect(0, 0, 4, 6))
	src.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})

	square := fitIconToSquare(src, 8)
	if square == nil {
		t.Fatal("expected non-nil result")
	}
	bounds := square.Bounds()
	if bounds.Dx() != 8 || bounds.Dy() != 8 {
		t.Fatalf("expected 8x8, got %dx%d", bounds.Dx(), bounds.Dy())
	}
	// src(0,0) → dst(1,0)
	if square.RGBAAt(1, 0) != src.RGBAAt(0, 0) {
		t.Fatal("source not scaled correctly")
	}
}

func TestFitIconToSquare_Empty(t *testing.T) {
	if fitIconToSquare(nil, 8) != nil {
		t.Fatal("nil source should return nil")
	}
}

func TestFitIconToSquare_ZeroSize(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 4, 4))
	if fitIconToSquare(src, 0) != nil {
		t.Fatal("zero size should return nil")
	}
}

// ========================================================================
// applyAppIconMaskAlpha（掩码 RGB 全零 = 绘制区域）
// ========================================================================

func TestApplyAppIconMaskAlpha_Simple(t *testing.T) {
	dst := image.NewRGBA(image.Rect(0, 0, 3, 3))
	mask := image.NewRGBA(image.Rect(0, 0, 3, 3))
	for y := 0; y < 3; y++ {
		for x := 0; x < 3; x++ {
			dst.SetRGBA(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
			mask.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	// 掩码全零 → 绘制区域（dst 已有 alpha，保留 255）
	mask.SetRGBA(1, 1, color.RGBA{R: 0, G: 0, B: 0, A: 0})

	applyAppIconMaskAlpha(dst, mask)

	// (1,1) 掩码全零 → 绘制区域 → alpha 保留 255
	if a := dst.RGBAAt(1, 1).A; a != 255 {
		t.Fatalf("drawn pixel alpha should remain 255, got %d", a)
	}
	// (0,0) 掩码非零 → 非绘制区域 → alpha = 0
	if a := dst.RGBAAt(0, 0).A; a != 0 {
		t.Fatalf("non-drawn pixel alpha should be 0, got %d", a)
	}
}

func TestApplyAppIconMaskAlpha_NoVisibleAlpha(t *testing.T) {
	// dst 无 alpha（全 0），掩码全零区域应置 A=0xFF
	dst := image.NewRGBA(image.Rect(0, 0, 2, 2))
	mask := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			dst.SetRGBA(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 0})
			mask.SetRGBA(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 0})
		}
	}
	applyAppIconMaskAlpha(dst, mask)
	if a := dst.RGBAAt(0, 0).A; a != 0xFF {
		t.Fatalf("expected opaque alpha 255, got %d", a)
	}
}

func TestApplyAppIconMaskAlpha_Nil(t *testing.T) {
	// 不应 panic
	applyAppIconMaskAlpha(nil, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	applyAppIconMaskAlpha(image.NewRGBA(image.Rect(0, 0, 2, 2)), nil)
	applyAppIconMaskAlpha(nil, nil)
}

// ========================================================================
// normalizeAppIconPNGWithSize（返回 ([]byte, bool)）
// ========================================================================

func TestNormalizeAppIconPNGWithSize_SameSize(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 128, G: 64, B: 32, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	result, ok := normalizeAppIconPNGWithSize(buf.Bytes(), 8)
	if !ok {
		t.Fatal("expected ok")
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestNormalizeAppIconPNGWithSize_InvalidData(t *testing.T) {
	_, ok := normalizeAppIconPNGWithSize(nil, 8)
	if ok {
		t.Fatal("expected not-ok for nil data")
	}
}

func TestNormalizeAppIconPNGWithSize_ZeroSizeDefaults(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 9, G: 8, B: 7, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	result, ok := normalizeAppIconPNGWithSize(buf.Bytes(), 0)
	if !ok {
		t.Fatal("expected ok for zero size (defaults to 256)")
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

// ========================================================================
// releaseComObject（nil/零安全）
// ========================================================================

func TestReleaseComObject_Nil(t *testing.T) {
	// should not panic
	releaseComObject(nil)
}

func TestReleaseComObject_NilPtr(t *testing.T) {
	var iface uintptr
	releaseComObject(unsafe.Pointer(&iface))
}

// ========================================================================
// deleteAppDC / deleteAppObject（nil 安全）
// ========================================================================

func TestDeleteAppDC_Zero(t *testing.T) {
	deleteAppDC(0)
}

func TestDeleteAppObject_Zero(t *testing.T) {
	deleteAppObject(0)
}

// ========================================================================
// invalid args
// ========================================================================

func TestDrawAppIconToRGBA_InvalidArgs(t *testing.T) {
	_, err := drawAppIconToRGBA(0, 0, 0, 0)
	if err == nil {
		t.Fatal("expected error for zero args")
	}
}
