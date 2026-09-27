package main

import "testing"

func TestNormalizeShortcutMode(t *testing.T) {
	if got := normalizeShortcutMode("Resolved", ""); got != "resolved" {
		t.Errorf("Resolved -> %q", got)
	}
	if got := normalizeShortcutMode("SHORTCUT", ""); got != "shortcut" {
		t.Errorf("SHORTCUT -> %q", got)
	}
	if got := normalizeShortcutMode("  resolved  ", ""); got != "resolved" {
		t.Errorf("trimmed resolved -> %q", got)
	}
	if got := normalizeShortcutMode("maximized", ""); got != "" {
		t.Errorf("other -> %q", got)
	}
	if got := normalizeShortcutMode("", "x.lnk"); got != "shortcut" {
		t.Errorf("fallback path -> %q", got)
	}
	if got := normalizeShortcutMode("", ""); got != "" {
		t.Errorf("empty -> %q", got)
	}
}
