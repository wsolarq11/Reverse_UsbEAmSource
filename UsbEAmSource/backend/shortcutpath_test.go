package main

import "testing"

func TestNormalizeShortcutPath(t *testing.T) {
	if got := normalizeShortcutPath("  c:/x.exe  ", "d:/y.lnk"); got != "c:/x.exe" {
		t.Errorf("own -> %q", got)
	}
	if got := normalizeShortcutPath("", "d:/y.lnk"); got != "d:/y.lnk" {
		t.Errorf("fallback lnk -> %q", got)
	}
	if got := normalizeShortcutPath("", "z.exe"); got != "" {
		t.Errorf("fallback non-shortcut -> %q", got)
	}
}
