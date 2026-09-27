package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestReadInternetShortcutFile_Content(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x.url")
	content := "[InternetShortcut]\nURL=https://example.com\n"
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := readInternetShortcutFile(p)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if string(got) != content {
		t.Fatalf("content = %q, want %q", got, content)
	}
}

func TestReadInternetShortcutFile_NotFound(t *testing.T) {
	_, err := readInternetShortcutFile(filepath.Join(t.TempDir(), "nope.url"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestReadInternetShortcutFile_TooLarge(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "big.url")
	big := make([]byte, 0x40010)
	if err := os.WriteFile(p, big, 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := readInternetShortcutFile(p)
	if err == nil {
		t.Fatal("expected error for oversized file")
	}
	if !errors.Is(err, errInternetShortcutTooLarge) {
		t.Fatalf("err = %v, want errInternetShortcutTooLarge", err)
	}
}
