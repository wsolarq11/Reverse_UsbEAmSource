package main

import "testing"

func TestDecodeInternetShortcutTextBounded_Empty(t *testing.T) {
	if got := decodeInternetShortcutTextBounded(nil); got != "" {
		t.Fatalf("nil -> %q, want empty", got)
	}
	if got := decodeInternetShortcutTextBounded([]byte{}); got != "" {
		t.Fatalf("empty -> %q, want empty", got)
	}
}

func TestDecodeInternetShortcutTextBounded_UTF8(t *testing.T) {
	// UTF-8 BOM 剥除
	got := decodeInternetShortcutTextBounded([]byte{0xEF, 0xBB, 0xBF, 'H', 'i'})
	if got != "Hi" {
		t.Fatalf("utf8-bom -> %q, want Hi", got)
	}
	// 无 BOM 原样
	if got := decodeInternetShortcutTextBounded([]byte("plain")); got != "plain" {
		t.Fatalf("no-bom -> %q, want plain", got)
	}
}

func TestDecodeInternetShortcutTextBounded_UTF16(t *testing.T) {
	// LE BOM FF FE + 'a' 'b'
	le := []byte{0xFF, 0xFE, 'a', 0x00, 'b', 0x00}
	if got := decodeInternetShortcutTextBounded(le); got != "ab" {
		t.Fatalf("utf16-le -> %q, want ab", got)
	}
	// BE BOM FE FF + 'A' 'B'
	be := []byte{0xFE, 0xFF, 0x00, 'A', 0x00, 'B'}
	if got := decodeInternetShortcutTextBounded(be); got != "AB" {
		t.Fatalf("utf16-be -> %q, want AB", got)
	}
}

func TestDecodeInternetShortcutTextBounded_TooBig(t *testing.T) {
	big := make([]byte, 0x40001)
	if got := decodeInternetShortcutTextBounded(big); got != "" {
		t.Fatalf("too-big -> %q, want empty", got)
	}
}

func TestDecodeUTF16ShortcutText(t *testing.T) {
	// LE "Hi"
	got := decodeUTF16ShortcutText([]byte{'H', 0x00, 'i', 0x00}, false)
	if got != "Hi" {
		t.Fatalf("le -> %q, want Hi", got)
	}
	// BE "Hi"
	got2 := decodeUTF16ShortcutText([]byte{0x00, 'H', 0x00, 'i'}, true)
	if got2 != "Hi" {
		t.Fatalf("be -> %q, want Hi", got2)
	}
	// 短于 2 → 空
	if got3 := decodeUTF16ShortcutText([]byte{0x00}, false); got3 != "" {
		t.Fatalf("short -> %q, want empty", got3)
	}
	// null 码元跳过：'a' + null(LE: 0x00 0x00) + 'b'（这里的 null 是 U+0000，2 字节 0x00 0x00）
	got4 := decodeUTF16ShortcutText([]byte{'a', 0x00, 0x00, 0x00, 'b', 0x00}, false)
	if got4 != "ab" {
		t.Fatalf("null-skip -> %q, want ab", got4)
	}
}
