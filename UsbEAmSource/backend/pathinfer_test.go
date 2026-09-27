package main

import "testing"

func TestLooksLikeFilesystemPath(t *testing.T) {
	if !looksLikeFilesystemPath(`C:\a`) || !looksLikeFilesystemPath("a/b") || !looksLikeFilesystemPath("/a") {
		t.Error("path-like must be true")
	}
	if looksLikeFilesystemPath("abc") || looksLikeFilesystemPath("") {
		t.Error("non-path must be false")
	}
}

func TestInferWorkingDir(t *testing.T) {
	cases := map[string]string{
		`C:\a\b\app.exe`: `C:\a\b`,
		`/usr/bin/x`:     "/usr/bin",
		"a/b/c":          "a/b",
		"relative":       "",
		"":               "",
	}
	for in, want := range cases {
		if got := inferWorkingDir(in); got != want {
			t.Errorf("inferWorkingDir(%q)=%q want %q", in, got, want)
		}
	}
}
