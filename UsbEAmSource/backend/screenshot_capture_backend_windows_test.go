package main

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestScreenshotDXGIBackendName(t *testing.T) {
	if got := (screenshotDXGIHDRScreenCaptureBackend{}).name(); got != "dxgi-hdr" {
		t.Fatalf("dxgi backend name = %q, want dxgi-hdr", got)
	}
}

func TestDefaultScreenshotScreenCaptureBackendDisabled(t *testing.T) {
	t.Setenv("USBEAM_SCREENSHOT_HDR_CAPTURE", "disabled")
	backend := defaultScreenshotScreenCaptureBackend()
	if got := backend.name(); got != "gdi" {
		t.Fatalf("disabled mode backend name = %q, want gdi", got)
	}
}

func TestBuildScreenshotDXGIVirtualScreenImageEmptyBounds(t *testing.T) {
	_, err := buildScreenshotDXGIVirtualScreenImage(nil, image.Rectangle{}, nil)
	if err == nil {
		t.Fatal("empty virtual bounds should error")
	}
}

func TestBuildScreenshotDXGIDesktopBoundsNoDisplays(t *testing.T) {
	_, err := buildScreenshotDXGIDesktopBoundsImage(nil, image.Rect(0, 0, 10, 10), func(screenshotDisplayCaptureInfo) (*image.RGBA, error) {
		return nil, nil
	})
	if err == nil {
		t.Fatal("no connected displays should error")
	}
}

func TestBuildScreenshotDXGIDesktopBoundsCompose(t *testing.T) {
	displays := []screenshotDisplayCaptureInfo{{
		Bounds:            image.Rect(0, 0, 10, 10),
		AttachedToDesktop: true,
		HDR:               true,
	}}
	src := image.NewRGBA(image.Rect(0, 0, 10, 10))
	draw.Draw(src, src.Bounds(), image.NewUniform(color.RGBA{R: 255, A: 255}), image.Point{}, draw.Src)

	img, err := buildScreenshotDXGIDesktopBoundsImage(displays, image.Rect(2, 2, 8, 8), func(d screenshotDisplayCaptureInfo) (*image.RGBA, error) {
		return src, nil
	})
	if err != nil {
		t.Fatalf("compose failed: %v", err)
	}
	if got := img.Rect; got != image.Rect(0, 0, 6, 6) {
		t.Fatalf("canvas rect = %v, want (0,0)-(6,6)", got)
	}
	c := img.RGBAAt(0, 0)
	if c.R != 255 || c.A != 255 {
		t.Fatalf("pixel (0,0) = %v, want opaque red", c)
	}
}

func TestCaptureScreenshotDXGIRegionImageEmpty(t *testing.T) {
	if _, err := captureScreenshotDXGIRegionImageWithOptions(nil, image.Rectangle{}, false); err == nil {
		t.Fatal("empty region should error")
	}
	if _, err := captureScreenshotDXGIRegionImageWithOptions(nil, image.Rect(0, 0, 10, 10), false); err == nil {
		t.Fatal("no connected displays should error")
	}
}

func TestCaptureScreenshotDXGIDisplayFrameWithFallbackUsingSDRNilFallback(t *testing.T) {
	if _, err := captureScreenshotDXGIDisplayFrameWithFallbackUsing(screenshotDisplayCaptureInfo{}, nil, nil); err == nil {
		t.Fatal("SDR display with nil GDI fallback should error")
	}
}
