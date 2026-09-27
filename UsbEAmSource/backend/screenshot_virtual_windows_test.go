package main

import (
	"image"
	"testing"
)

func darkenQRCodePreviewByte(v uint8) uint8 {
	return uint8((uint32(uint16(uint32(v)*0x2a)) * 0x147af) >> 0x17)
}

func TestBuildDarkenedQRCodeSelectionPreview(t *testing.T) {
	if got := buildDarkenedQRCodeSelectionPreview(nil); !got.Rect.Empty() {
		t.Fatalf("nil should produce empty rect, got %v", got.Rect)
	}
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.Pix = []uint8{255, 128, 64, 255, 0, 42, 255, 128}
	got := buildDarkenedQRCodeSelectionPreview(img)
	for i := 0; i+3 < len(got.Pix); i += 4 {
		for o := 0; o < 3; o++ {
			want := darkenQRCodePreviewByte(img.Pix[i+o])
			if got.Pix[i+o] != want {
				t.Fatalf("pixel %d channel %d = %d, want %d", i/4, o, got.Pix[i+o], want)
			}
		}
		if got.Pix[i+3] != img.Pix[i+3] {
			t.Fatalf("alpha channel should be preserved")
		}
	}
	if darkenQRCodePreviewByte(255) != 107 {
		t.Fatalf("darken(255) = %d, want 107", darkenQRCodePreviewByte(255))
	}
}

func TestSelectScreenshotScreenCaptureBackendName(t *testing.T) {
	attached := screenshotDisplayCaptureInfo{
		AttachedToDesktop: true,
		Bounds:            image.Rect(0, 0, 1920, 1080),
	}
	hdr := attached
	hdr.HDR = true

	cases := []struct {
		name      string
		displays  []screenshotDisplayCaptureInfo
		mode      string
		preferHDR bool
		want      string
	}{
		{"no prefer", nil, "auto", false, "gdi"},
		{"disabled", nil, "disabled", true, "gdi"},
		{"force attached", []screenshotDisplayCaptureInfo{attached}, "force", true, "dxgi-hdr"},
		{"force detached", []screenshotDisplayCaptureInfo{{}}, "force", true, "gdi"},
		{"auto hdr", []screenshotDisplayCaptureInfo{hdr}, "auto", true, "dxgi-hdr"},
		{"auto sdr", []screenshotDisplayCaptureInfo{attached}, "auto", true, "gdi"},
		{"unknown mode", []screenshotDisplayCaptureInfo{hdr}, "other", true, "gdi"},
	}
	for _, c := range cases {
		if got := selectScreenshotScreenCaptureBackendName(c.displays, c.mode, c.preferHDR); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}
