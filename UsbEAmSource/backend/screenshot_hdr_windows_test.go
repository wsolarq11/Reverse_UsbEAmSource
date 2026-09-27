package main

import "testing"

func TestParseScreenshotBool(t *testing.T) {
	truthy := []string{"1", "on", "yes", "true", " ON ", "Yes"}
	falsy := []string{"0", "no", "off", "false", "Off"}
	invalid := []string{"", "maybe", "2", "tru"}

	for _, v := range truthy {
		b, ok := parseScreenshotBool(v)
		if !ok || !b {
			t.Errorf("parseScreenshotBool(%q) = (%v,%v), want (true,true)", v, b, ok)
		}
	}
	for _, v := range falsy {
		b, ok := parseScreenshotBool(v)
		if !ok || b {
			t.Errorf("parseScreenshotBool(%q) = (%v,%v), want (false,true)", v, b, ok)
		}
	}
	for _, v := range invalid {
		b, ok := parseScreenshotBool(v)
		if ok || b {
			t.Errorf("parseScreenshotBool(%q) = (%v,%v), want (false,false)", v, b, ok)
		}
	}
}

func TestParseScreenshotHDRCaptureMode(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"0", "disabled"}, {"no", "disabled"}, {"off", "disabled"},
		{"false", "disabled"}, {"disable", "disabled"}, {"disabled", "disabled"},
		{"1", "auto"}, {"on", "auto"}, {"hdr", "auto"},
		{"yes", "auto"}, {"auto", "auto"}, {"true", "auto"},
		{"dxgi", "force"}, {"force", "force"}, {"always", "force"},
		{"  Off  ", "disabled"}, {"", "auto"}, {"   ", "auto"},
		{"unknown", "disabled"},
	}
	for _, c := range cases {
		if got := parseScreenshotHDRCaptureMode(c.in); got != c.want {
			t.Errorf("parseScreenshotHDRCaptureMode(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestResolveScreenshotHDRCaptureMode(t *testing.T) {
	t.Setenv("USBEAM_SCREENSHOT_HDR_CAPTURE", "force")
	if got := resolveScreenshotHDRCaptureMode(); got != "force" {
		t.Errorf("resolveScreenshotHDRCaptureMode() = %q, want force", got)
	}
	t.Setenv("USBEAM_SCREENSHOT_HDR_CAPTURE", "")
	if got := resolveScreenshotHDRCaptureMode(); got != "auto" {
		t.Errorf("resolveScreenshotHDRCaptureMode() = %q, want auto", got)
	}
}
