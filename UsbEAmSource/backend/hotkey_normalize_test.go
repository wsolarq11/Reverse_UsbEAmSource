package main

import "testing"

func TestNormalizeHotCorner(t *testing.T) {
	cases := map[string]string{
		"lt": "topLeft", "topleft": "topLeft", "lefttop": "topLeft",
		"top-left": "topLeft", "left-top": "topLeft",
		"rt": "topRight", "topright": "topRight", "righttop": "topRight",
		"top-right": "topRight", "right-top": "topRight",
		"lb": "bottomLeft", "bottomleft": "bottomLeft", "leftbottom": "bottomLeft",
		"bottom-left": "bottomLeft", "left-bottom": "bottomLeft",
		"rb": "bottomRight", "bottomright": "bottomRight", "rightbottom": "bottomRight",
		"bottom-right": "bottomRight", "right-bottom": "bottomRight",
		"  Top-Left  ": "topLeft", // TrimSpace + ToLower
		"unknown":      "", "": "",
	}
	for in, want := range cases {
		if got := normalizeHotCorner(in); got != want {
			t.Errorf("normalizeHotCorner(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeMouseGestureHotkey(t *testing.T) {
	for _, in := range []string{"win", "Win", "WIN", "windows", "Windows", "  win  "} {
		if got := normalizeMouseGestureHotkey(in); got != "Win" {
			t.Errorf("normalizeMouseGestureHotkey(%q) = %q, want Win", in, got)
		}
	}
	if got := normalizeMouseGestureHotkey(""); got != "" {
		t.Errorf("empty -> %q, want empty", got)
	}
	if got := normalizeMouseGestureHotkey("Ctrl+Alt+S"); got != "Ctrl+Alt+S" {
		t.Errorf("Ctrl+Alt+S -> %q, want Ctrl+Alt+S", got)
	}
}

func TestNormalizeFileSearchResourceMode(t *testing.T) {
	cases := map[string]string{
		"hot": "resident", "fast": "resident", "resident": "resident", "performance": "resident",
		"cold": "memory-saver", "memory": "memory-saver", "low-memory": "memory-saver",
		"low_memory": "memory-saver", "mem-saver": "memory-saver", "mem_saver": "memory-saver",
		"balanced": "balanced", "unknown": "balanced", "": "balanced",
	}
	for in, want := range cases {
		if got := normalizeFileSearchResourceMode(in, ""); got != want {
			t.Errorf("normalizeFileSearchResourceMode(%q) = %q, want %q", in, got, want)
		}
	}
	if got := normalizeFileSearchResourceMode("", "resident"); got != "resident" {
		t.Errorf("legacy fallback = %q, want resident", got)
	}
	if got := normalizeFileSearchResourceMode("", "  FAST  "); got != "resident" {
		t.Errorf("legacy fallback trim = %q, want resident", got)
	}
}

func TestNormalizeLauncherHotkeyBindingValue(t *testing.T) {
	set := make(map[string]struct{})
	if got := normalizeLauncherHotkeyBindingValue("Ctrl+S", true, set, false); got != "Ctrl+S" {
		t.Errorf("first = %q, want Ctrl+S", got)
	}
	if got := normalizeLauncherHotkeyBindingValue("Ctrl+S", true, set, false); got != "" {
		t.Errorf("dup = %q, want empty", got)
	}
	if got := normalizeLauncherHotkeyBindingValue("Ctrl+S", false, set, false); got != "Ctrl+S" {
		t.Errorf("flag1=false = %q, want Ctrl+S", got)
	}
	if got := normalizeLauncherHotkeyBindingValue("Ctrl+S", true, nil, false); got != "Ctrl+S" {
		t.Errorf("nil set = %q, want Ctrl+S", got)
	}
	if got := normalizeLauncherHotkeyBindingValue("", true, set, false); got != "" {
		t.Errorf("empty = %q, want empty", got)
	}
}

func TestNormalizeLauncherHotkeyBindings(t *testing.T) {
	b := normalizeLauncherHotkeyBindings(launcherHotkeyBindings{
		SummonSearch:                  "Ctrl+Alt+S",
		SummonSearchEnabled:           true,
		SummonOnly:                    "Alt+S",
		SummonOnlyEnabled:             true,
		Screenshot:                    "Ctrl+Shift+S",
		ScreenshotEnabled:             true,
		ScreenshotQRCode:              "Ctrl+Shift+Q",
		ScreenshotQRCodeEnabled:       true,
		ScreenshotAllScreens:          "Ctrl+Shift+A",
		ScreenshotAllScreensEnabled:   true,
		ScreenshotScrolling:           "Ctrl+Shift+C",
		ScreenshotScrollingEnabled:    true,
		ScreenshotActiveWindow:        "Ctrl+Shift+W",
		ScreenshotActiveWindowEnabled: true,
		ScreenshotFeatureEnabled:      true,
	})
	if b.SummonSearch != "Ctrl+Alt+S" {
		t.Errorf("SummonSearch = %q", b.SummonSearch)
	}
	if b.SummonOnly != "Alt+S" {
		t.Errorf("SummonOnly = %q", b.SummonOnly)
	}
	// 截图热键按 Shift(2) > Ctrl(1) 权重降序。
	if b.Screenshot != "Shift+Ctrl+S" {
		t.Errorf("Screenshot = %q, want Shift+Ctrl+S", b.Screenshot)
	}
	if !b.ScreenshotFeatureEnabled || !b.ScreenshotEnabled || !b.SummonOnlyEnabled {
		t.Error("enabled flags must be passed through")
	}
}

func TestNormalizeLauncherHotkeyBindingsFallback(t *testing.T) {
	b := normalizeLauncherHotkeyBindings(launcherHotkeyBindings{})
	if b.SummonSearch != "Ctrl+Alt+S" {
		t.Errorf("fallback SummonSearch = %q, want Ctrl+Alt+S", b.SummonSearch)
	}
	if b.SummonOnly != "" {
		t.Errorf("zero SummonOnly = %q, want empty", b.SummonOnly)
	}
}
