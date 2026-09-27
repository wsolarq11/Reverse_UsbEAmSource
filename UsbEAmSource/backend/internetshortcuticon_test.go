package main

import "testing"

func TestBuildInternetShortcutIconLocation_Basic(t *testing.T) {
	cases := []struct {
		file, idx, want string
	}{
		{"C:\\x.ico", "0", "C:\\x.ico,0"},
		{"C:\\x.ico", "3", "C:\\x.ico,3"},
		{"  C:\\x.ico  ", "  2  ", "C:\\x.ico,2"},
		{"C:\\x.ico", "abc", "C:\\x.ico,0"}, // Atoi 失败 → 0
		{"C:\\x.ico", "", "C:\\x.ico,0"},
	}
	for _, c := range cases {
		if got := buildInternetShortcutIconLocation(c.file, c.idx); got != c.want {
			t.Errorf("build(%q,%q) = %q, want %q", c.file, c.idx, got, c.want)
		}
	}
}

func TestBuildInternetShortcutIconLocation_Empty(t *testing.T) {
	if got := buildInternetShortcutIconLocation("   ", "0"); got != "" {
		t.Fatalf("empty file -> %q, want empty", got)
	}
}

func TestBuildInternetShortcutIconLocation_ContainsComma(t *testing.T) {
	// 含逗号时去引号（[P] 主路径）
	got := buildInternetShortcutIconLocation(`"C:\x.ico,0"`, "1")
	if got != "C:\\x.ico,0,1" {
		t.Fatalf("quoted-comma -> %q, want C:\\x.ico,0,1", got)
	}
}
