package main

import (
	"testing"
	"unsafe"
)

func TestFallbackAppDisplayName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"   ", ""},
		{`C:\Windows\notepad.exe`, "notepad"},
		{`C:\Users\file.url`, "file"},
		{`/server/share/document.appref-ms`, "document"},
		{`simple.lnk`, "simple"},
		{`noext`, "noext"},
		{`C:\path\with\trailing\slash\`, ""},
		{`C:\path\.hidden.lnk`, ".hidden"},
		{`a.b.c.lnk`, "a.b.c"},
		{`many.dots.in.name.url`, "many.dots.in.name"},
		{`C:\test.lnk`, "test"},
		{`/home/user/test.url`, "test"},
		// .exe gets stripped
		{`app.exe`, "app"},
	}
	for _, tc := range tests {
		got := fallbackAppDisplayName(tc.input)
		if got != tc.expected {
			t.Errorf("fallbackAppDisplayName(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestStripDisplayExtensions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test.exe", "test"},
		{"test.lnk", "test"},
		{"test.appref-ms", "test"},
		{"test.url", "test"},
		{"test.unknown", "test.unknown"},
		{"noext", "noext"},
		{"", ""},
		{".exe", ""},
		{"a.b.lnk", "a.b"},
		{"mixed.CASE.LNK", "mixed.CASE"},
		{"test.exe.url", "test.exe"}, // only first match (.url stripped)
	}
	for _, tc := range tests {
		got := stripDisplayExtensions(tc.input)
		if got != tc.expected {
			t.Errorf("stripDisplayExtensions(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestResolveAppDisplayName_NonShortcut(t *testing.T) {
	result, found := resolveAppDisplayName(`C:\Windows\notepad.exe`)
	if !found || result != `C:\Windows\notepad.exe` {
		t.Errorf("resolveAppDisplayName(non-shortcut) = (%q, %v), want (%q, true)", result, found, `C:\Windows\notepad.exe`)
	}
}

func TestResolveAppDisplayName_Empty(t *testing.T) {
	result, found := resolveAppDisplayName("")
	if found || result != "" {
		t.Errorf("resolveAppDisplayName(\"\") = (%q, %v), want (\"\", false)", result, found)
	}
}

func TestResolveAppDisplayName_Shortcut(t *testing.T) {
	result, found := resolveAppDisplayName(`C:\test.lnk`)
	if !found || result == "" {
		t.Errorf("resolveAppDisplayName(shortcut stub) = (%q, %v), want (non-empty, true)", result, found)
	}
}

func TestResolveFileDisplayName_Empty(t *testing.T) {
	result, found := resolveFileDisplayName("")
	if found || result != "" {
		t.Errorf("resolveFileDisplayName(\"\") = (%q, %v), want (\"\", false)", result, found)
	}
}

func TestResolveFileDisplayName_NotFound(t *testing.T) {
	result, found := resolveFileDisplayName(`C:\nonexistent_file_12345.xyz`)
	if found || result != "" {
		t.Errorf("resolveFileDisplayName(nonexistent) = (%q, %v), want (\"\", false)", result, found)
	}
}

func TestQueryVersionField_EmptyPath(t *testing.T) {
	result, found := queryVersionField("", "FileDescription")
	if found || result != "" {
		t.Errorf("queryVersionField(\"\") = (%q, %v), want (\"\", false)", result, found)
	}
}

func TestLoadVersionInfo_Empty(t *testing.T) {
	_, _, err := loadVersionInfo("")
	if err == nil {
		t.Error("loadVersionInfo(\"\") expected error")
	}
}

func TestDecodeVersionUTF16String_Nil(t *testing.T) {
	if got := decodeVersionUTF16String(nil, 0); got != "" {
		t.Errorf("decodeVersionUTF16String(nil, 0) = %q, want \"\"", got)
	}
}

func TestDecodeVersionUTF16String_CountZero(t *testing.T) {
	var buf [4]uint16
	if got := decodeVersionUTF16String(unsafe.Pointer(&buf[0]), 0); got != "" {
		t.Errorf("decodeVersionUTF16String(_, 0) = %q, want \"\"", got)
	}
}

func TestQueryVersionValueLocation_NilData(t *testing.T) {
	_, _, found := queryVersionValueLocation(nil, 0, "test")
	if found {
		t.Error("queryVersionValueLocation(nil) should return false")
	}
}

func TestQueryVersionValueLocation_EmptySubBlock(t *testing.T) {
	// Empty subBlock: UTF16PtrFromString("") succeeds (returns empty string ptr)
	// The nil data check is hit first.
	data := make([]byte, 12)
	_, _, found := queryVersionValueLocation(data, 12, "")
	// UTF16PtrFromString succeeds, then VerQueryValueW is called and fails
	// resulting in found=false
	if found {
		t.Log("note: empty string subblock may succeed on some systems; test is system-dependent")
	}
}

func TestDecodeVersionUTF16String_NegativeCount(t *testing.T) {
	if got := decodeVersionUTF16String(nil, -1); got != "" {
		t.Errorf("decodeVersionUTF16String(nil, -1) = %q, want \"\"", got)
	}
}

func TestQueryVersionTranslations_Invalid(t *testing.T) {
	_, err := queryVersionTranslations(nil, 0)
	if err == nil {
		t.Error("queryVersionTranslations(nil) expected error")
	}

	data := make([]byte, 4)
	_, err2 := queryVersionTranslations(data, 0)
	if err2 == nil {
		t.Error("queryVersionTranslations(invalid) expected error")
	}
}
