// AUTO-RECONSTRUCTED — DOMAIN: app entry normalize
// 研究用途. [S 汇编实证 0x140889780, 4544B]
package main

import "testing"

func TestNormalizeAppEntries_FiltersEmptyName(t *testing.T) {
	entries := []AppEntry{
		{Name: "  "},
		{Name: "App1", Path: "C:\\apps\\app1.exe"},
		{Name: ""},
		{Name: "App2", Path: "C:\\apps\\app2.exe"},
	}
	got := normalizeAppEntries(entries)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	if got[0].Name != "App1" || got[1].Name != "App2" {
		t.Errorf("unexpected order: %+v", got)
	}
}

func TestNormalizeAppEntries_PreservesFields(t *testing.T) {
	entries := []AppEntry{
		{
			Name:           "TestApp",
			EntryType:      "app",
			Path:           "C:\\apps\\test.exe",
			Icon:           "test.ico",
			Favorite:       true,
			LaunchCount:    5,
			LastLaunchedAt: "2024-01-01T00:00:00Z",
		},
	}
	got := normalizeAppEntries(entries)
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].Name != "TestApp" {
		t.Errorf("Name = %q, want TestApp", got[0].Name)
	}
	if got[0].EntryType != "app" {
		t.Errorf("EntryType = %q, want app", got[0].EntryType)
	}
	if got[0].Path != "C:\\apps\\test.exe" {
		t.Errorf("Path = %q", got[0].Path)
	}
	if got[0].Icon != "test.ico" {
		t.Errorf("Icon = %q", got[0].Icon)
	}
	if !got[0].Favorite {
		t.Error("Favorite should be true")
	}
	if got[0].LaunchCount != 5 {
		t.Errorf("LaunchCount = %d, want 5", got[0].LaunchCount)
	}
}

func TestNormalizeAppEntries_EntryTypeDirectory(t *testing.T) {
	entries := []AppEntry{
		{
			Name:      "MyDir",
			EntryType: "directory",
			Path:      "C:\\mydir",
		},
	}
	got := normalizeAppEntries(entries)
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].EntryType != "directory" {
		t.Errorf("EntryType = %q, want directory", got[0].EntryType)
	}
	// directory should clear shortcut fields
	if got[0].ShortcutMode != "" {
		t.Errorf("directory ShortcutMode should be empty, got %q", got[0].ShortcutMode)
	}
}

func TestNormalizeAppEntries_EmptyInput(t *testing.T) {
	got := normalizeAppEntries(nil)
	if got != nil {
		t.Errorf("nil input should return nil, got %+v", got)
	}
	got = normalizeAppEntries([]AppEntry{})
	if got != nil {
		t.Errorf("empty input should return nil, got %+v", got)
	}
}

func TestNormalizeAppEntries_IconDataPrefix(t *testing.T) {
	entries := []AppEntry{
		{
			Name:     "IconApp",
			Path:     "C:\\app.exe",
			IconData: "data:image/png;base64,abc123",
		},
	}
	got := normalizeAppEntries(entries)
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	// iconData matching "data:image/" prefix should not be treated as upload icon
	// when autoIcon is false and iconRef is empty
	if got[0].IconData != "data:image/png;base64,abc123" {
		t.Errorf("IconData = %q", got[0].IconData)
	}
}
