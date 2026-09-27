package main

import (
	"image"
	"testing"
	"unsafe"
)

func TestForceScreenshotDIBAlphaOpaqueInRect(t *testing.T) {
	const w, h = 4, 3
	buf := make([]byte, w*h*4)
	if !forceScreenshotDIBAlphaOpaqueInRect(unsafe.Pointer(&buf[0]), w, h, 1, 1, 4, 3) {
		t.Fatal("expected true")
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			off := (y*w+x)*4 + 3
			if x >= 1 && y >= 1 {
				if buf[off] != 0xFF {
					t.Fatalf("pixel (%d,%d) alpha = %d, want 0xFF", x, y, buf[off])
				}
			} else if buf[off] != 0 {
				t.Fatalf("pixel (%d,%d) alpha = %d, want 0", x, y, buf[off])
			}
		}
	}
}

func TestForceScreenshotDIBAlphaOpaqueInRectRejects(t *testing.T) {
	buf := make([]byte, 16)
	if forceScreenshotDIBAlphaOpaqueInRect(nil, 1, 1, 0, 0, 1, 1) {
		t.Fatal("nil bits should be rejected")
	}
	if forceScreenshotDIBAlphaOpaqueInRect(unsafe.Pointer(&buf[0]), 0, 1, 0, 0, 1, 1) {
		t.Fatal("non-positive width should be rejected")
	}
	if forceScreenshotDIBAlphaOpaqueInRect(unsafe.Pointer(&buf[0]), 2, 2, 3, 3, 5, 5) {
		t.Fatal("rect outside bitmap should be rejected")
	}
}

func TestDrawScreenshotCaptureCursorDisabled(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	if drawScreenshotCaptureCursorOnRGBA(img, 0, 0, 8, 8, false, screenshotCursorSnapshot{}) {
		t.Fatal("captureCursor=false should return false")
	}
}

func TestDrawScreenshotCursorSnapshotSizeMismatch(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	if drawScreenshotCursorSnapshotOnRGBA(img, 0, 0, 4, 8, screenshotCursorInfo{}, true) {
		t.Fatal("size mismatch should return false")
	}
	if drawScreenshotCursorSnapshotOnRGBA(nil, 0, 0, 8, 8, screenshotCursorInfo{}, true) {
		t.Fatal("nil image should return false")
	}
}
