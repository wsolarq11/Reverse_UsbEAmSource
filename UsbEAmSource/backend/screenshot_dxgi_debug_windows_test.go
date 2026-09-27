package main

import (
	"image"
	"strings"
	"testing"
)

func TestScreenshotDXGIColorSpaceDebugName(t *testing.T) {
	cases := []struct {
		in   uint32
		want string
	}{
		{0, "RGB_FULL_G22_NONE_P709(0)"},
		{12, "RGB_FULL_G2084_NONE_P2020(12)"},
		{13, "YCBCR_STUDIO_G2084_LEFT_P2020(13)"},
		{14, "RGB_STUDIO_G2084_NONE_P2020(14)"},
		{16, "YCBCR_STUDIO_G2084_TOP_P2020(16)"},
		{18, "YCBCR_STUDIO_GHLG_TOP_P2020(18)"},
		{19, "YCBCR_FULL_GHLG_TOP_P2020(19)"},
		{7, "UNKNOWN(7)"},
	}
	for _, c := range cases {
		if got := screenshotDXGIColorSpaceDebugName(c.in); got != c.want {
			t.Errorf("screenshotDXGIColorSpaceDebugName(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestScreenshotDisplayCaptureLabel(t *testing.T) {
	info := screenshotDisplayCaptureInfo{
		AdapterIndex: 3,
		OutputIndex:  5,
		AdapterName:  "  Intel(R) UHD  ",
		DeviceName:   "\\\\.\\DISPLAY1",
	}
	if got := screenshotDisplayCaptureLabel(info); got != info.DeviceName {
		t.Errorf("label with device = %q, want %q", got, info.DeviceName)
	}

	info.DeviceName = "   "
	if got := screenshotDisplayCaptureLabel(info); got != "Intel(R) UHD#5" {
		t.Errorf("label with adapter = %q, want Intel(R) UHD#5", got)
	}

	info.AdapterName = "  "
	if got := screenshotDisplayCaptureLabel(info); got != "adapter=3 output=5" {
		t.Errorf("label fallback = %q, want adapter=3 output=5", got)
	}
}

func TestFormatScreenshotDisplayCaptureInfoForDebug(t *testing.T) {
	info := screenshotDisplayCaptureInfo{
		AdapterIndex:          1,
		OutputIndex:           2,
		AdapterName:           "adp",
		DeviceName:            "dev",
		Bounds:                image.Rect(0, 0, 1920, 1080),
		Rotation:              0,
		AttachedToDesktop:     true,
		BitsPerColor:          8,
		ColorSpace:            0,
		HDR:                   true,
		AdvancedColor:         true,
		MinLuminance:          0.5,
		MaxLuminance:          270,
		MaxFullFrameLuminance: 270,
	}
	got := formatScreenshotDisplayCaptureInfoForDebug(info)
	if !strings.Contains(got, "label=\"dev\"") {
		t.Errorf("format missing label: %s", got)
	}
	if !strings.Contains(got, "adapter=1") || !strings.Contains(got, "output=2") {
		t.Errorf("format missing indices: %s", got)
	}
	if !strings.Contains(got, "hdr=true") || !strings.Contains(got, "advanced=true") {
		t.Errorf("format missing flags: %s", got)
	}
	if !strings.Contains(got, "min=0.50") {
		t.Errorf("format missing luminance: %s", got)
	}
}
