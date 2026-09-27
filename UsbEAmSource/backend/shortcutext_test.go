package main

import "testing"

func TestShortcutPath(t *testing.T) {
	if !isShortcutFilePath(`C:\d\l\x.lnk`) || !isShortcutFilePath("x.URL") || !isShortcutFilePath("a.lnk") {
		t.Error("lnk/url must be shortcut")
	}
	if isShortcutFilePath("x.exe") || isShortcutFilePath("dir") || isShortcutFilePath("") {
		t.Error("exe/dir/plain must not")
	}
}
