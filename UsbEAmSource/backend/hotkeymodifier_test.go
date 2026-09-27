package main

import "testing"

func TestCanonicalHotkeyModifier(t *testing.T) {
	for in, want := range map[string]string{
		"CTRL": "Ctrl", "Control": "Ctrl", "ctrl": "Ctrl",
		"alt": "Alt", "ALT": "Alt", "option": "Alt",
		"shift": "Shift", "SHIFT": "Shift",
		"win": "Win", "windows": "Win",
		"meta": "Win", "super": "Win",
		"cmd": "Ctrl", "command": "Ctrl",
	} {
		got, ok := canonicalHotkeyModifier(in)
		if !ok || got != want {
			t.Errorf("%q -> %q,%v want %q", in, got, ok, want)
		}
	}
	if _, ok := canonicalHotkeyModifier("weird"); ok {
		t.Error("unknown modifier must return false")
	}
	if got, ok := canonicalHotkeyModifier("   ctrl  "); !ok || got != "Ctrl" {
		t.Errorf("trim not applied: %q,%v", got, ok)
	}
}
