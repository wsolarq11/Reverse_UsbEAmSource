// AUTO-RECONSTRUCTED — DOMAIN: link entry normalize
// 研究用途. [S 汇编实证 0x14088a940, 2016B]
package main

import "testing"

func TestNormalizeLinkEntries_NilInput(t *testing.T) {
	got := normalizeLinkEntries(nil, "default")
	if got != nil {
		t.Errorf("nil input should return nil, got %+v", got)
	}
}

func TestNormalizeLinkEntries_EmptyInput(t *testing.T) {
	got := normalizeLinkEntries([]LinkEntry{}, "default")
	if got != nil {
		t.Errorf("empty input should return nil, got %+v", got)
	}
}

func TestNormalizeLinkEntries_FiltersEmptyURL(t *testing.T) {
	entries := []LinkEntry{
		{URL: "  ", Name: "Spaces"},
		{URL: "", Name: "Empty"},
		{URL: "https://example.com", Name: "Valid"},
		{URL: "https://test.org", Name: "Test"},
	}
	got := normalizeLinkEntries(entries, "default")
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	if got[0].Name != "Valid" || got[1].Name != "Test" {
		t.Errorf("unexpected order: %+v", got)
	}
}

func TestNormalizeLinkEntries_NameFallbackToURL(t *testing.T) {
	entries := []LinkEntry{
		{URL: "https://example.com", Name: "  "},
	}
	got := normalizeLinkEntries(entries, "default")
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].Name != "https://example.com" {
		t.Errorf("Name should fallback to URL, got %q", got[0].Name)
	}
}

func TestNormalizeLinkEntries_PreservesFields(t *testing.T) {
	entries := []LinkEntry{
		{
			URL:            "https://example.com",
			Name:           "Example",
			TitleLocked:    true,
			Icon:           "custom.ico",
			Favorite:       true,
			LaunchCount:    5,
			LastLaunchedAt: "2024-01-01T00:00:00Z",
			Browser:        "chrome",
		},
	}
	got := normalizeLinkEntries(entries, "default")
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].Name != "Example" {
		t.Errorf("Name = %q", got[0].Name)
	}
	if !got[0].TitleLocked {
		t.Error("TitleLocked should be true")
	}
	if got[0].Icon != "custom.ico" {
		t.Errorf("Icon = %q", got[0].Icon)
	}
	if !got[0].Favorite {
		t.Error("Favorite should be true")
	}
	if got[0].LaunchCount != 5 {
		t.Errorf("LaunchCount = %d, want 5", got[0].LaunchCount)
	}
	if got[0].LastLaunchedAt != "2024-01-01T00:00:00Z" {
		t.Errorf("LastLaunchedAt = %q", got[0].LastLaunchedAt)
	}
	if got[0].Browser != "chrome" {
		t.Errorf("Browser = %q", got[0].Browser)
	}
	if got[0].URL != "https://example.com" {
		t.Errorf("URL = %q", got[0].URL)
	}
}

func TestNormalizeLinkEntries_IconModeUpload(t *testing.T) {
	entries := []LinkEntry{
		{
			URL:      "https://example.com",
			Name:     "Site",
			IconMode: "upload",
			IconData: "data:image/png;base64,abc",
			IconRef:  "ref123",
		},
	}
	got := normalizeLinkEntries(entries, "default")
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	// iconMode "upload" should match, preserve iconData/iconRef
	if got[0].IconMode != "upload" {
		t.Errorf("IconMode = %q, want upload", got[0].IconMode)
	}
	if got[0].IconData == "" || got[0].IconRef == "" {
		t.Error("upload mode should preserve iconData and iconRef")
	}
}

func TestNormalizeLinkEntries_IconModeNonUpload(t *testing.T) {
	// iconRef 非空时 normalizeLinkIconMode 返回 "upload"，这是汇编确凿行为。
	// 此测试验证当 iconRef 为空且 iconData 不匹配前缀时的 favicon 分支。
	entries := []LinkEntry{
		{
			URL:      "https://example.com",
			Name:     "Site",
			IconMode: "favicon",
			IconData: "other-data",
			IconRef:  "",
		},
	}
	got := normalizeLinkEntries(entries, "default")
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].IconMode != "favicon" {
		t.Errorf("IconMode = %q, want favicon", got[0].IconMode)
	}
	// non-upload mode must clear iconData and iconRef
	if got[0].IconData != "" {
		t.Errorf("non-upload mode should clear IconData, got %q", got[0].IconData)
	}
	if got[0].IconRef != "" {
		t.Errorf("non-upload mode should clear IconRef, got %q", got[0].IconRef)
	}
}

func TestNormalizeLinkEntries_IconFallback(t *testing.T) {
	entries := []LinkEntry{
		{URL: "https://example.com", Name: "NoIcon", Icon: ""},
		{URL: "https://test.com", Name: "HasIcon", Icon: "  "},
		{URL: "https://keep.com", Name: "KeepIcon", Icon: "myicon.ico"},
	}
	got := normalizeLinkEntries(entries, "fallback-icon")
	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
	if got[0].Icon != "fallback-icon" {
		t.Errorf("empty Icon should fallback to default, got %q", got[0].Icon)
	}
	if got[1].Icon != "fallback-icon" {
		t.Errorf("whitespace Icon should fallback to default, got %q", got[1].Icon)
	}
	if got[2].Icon != "myicon.ico" {
		t.Errorf("non-empty Icon should be kept, got %q", got[2].Icon)
	}
}

func TestNormalizeLinkEntries_IconURLTrimmed(t *testing.T) {
	entries := []LinkEntry{
		{URL: "https://example.com", Name: "Site", IconURL: "  https://icons.example.com/fav.ico  "},
	}
	got := normalizeLinkEntries(entries, "default")
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].IconURL != "https://icons.example.com/fav.ico" {
		t.Errorf("IconURL should be trimmed, got %q", got[0].IconURL)
	}
}

func TestNormalizeLinkEntries_CleanStringList(t *testing.T) {
	entries := []LinkEntry{
		{
			URL:  "https://example.com",
			Name: "Site",
			Args: []string{"  --verbose  ", "", "--output", "  ", "--verbose"},
			Tags: []string{"  web  ", "  ", "web", "important"},
		},
	}
	got := normalizeLinkEntries(entries, "default")
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	// Args should be cleaned: trim, remove empty, dedup by lowercase
	if len(got[0].Args) != 2 {
		t.Errorf("Args count = %d, want 2 (--verbose, --output after dedup)", len(got[0].Args))
	}
	// Tags should be cleaned
	if len(got[0].Tags) != 2 {
		t.Errorf("Tags count = %d, want 2 (web, important)", len(got[0].Tags))
	}
}

func TestNormalizeLinkEntries_NegativeLaunchCount(t *testing.T) {
	entries := []LinkEntry{
		{URL: "https://example.com", Name: "Site", LaunchCount: -5},
	}
	got := normalizeLinkEntries(entries, "default")
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].LaunchCount != 0 {
		t.Errorf("negative LaunchCount should be 0, got %d", got[0].LaunchCount)
	}
}

func TestNormalizeLinkEntries_LastLaunchedAtTrimmed(t *testing.T) {
	entries := []LinkEntry{
		{URL: "https://example.com", Name: "Site", LastLaunchedAt: "  2024-01-01T00:00:00Z  "},
	}
	got := normalizeLinkEntries(entries, "default")
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].LastLaunchedAt != "2024-01-01T00:00:00Z" {
		t.Errorf("LastLaunchedAt should be trimmed, got %q", got[0].LastLaunchedAt)
	}
}

func TestNormalizeLinkEntries_EmptyURLWithWhitespaceBrowser(t *testing.T) {
	entries := []LinkEntry{
		{URL: "https://site.com", Name: "Site", Browser: "  edge  "},
	}
	got := normalizeLinkEntries(entries, "default")
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].Browser != "edge" {
		t.Errorf("Browser should be trimmed, got %q", got[0].Browser)
	}
}

func TestNormalizeLinkEntries_IconModeChain(t *testing.T) {
	// Test the full icon mode chain: iconRef non-empty -> upload, else iconData prefix -> upload
	entries := []LinkEntry{
		{URL: "https://a.com", Name: "A", IconRef: "ref", IconData: "data:image/png;abc"},
		{URL: "https://b.com", Name: "B", IconRef: "", IconData: "data:image/png;abc"},
		{URL: "https://c.com", Name: "C", IconRef: "", IconData: "other"},
		{URL: "https://d.com", Name: "D", IconRef: "", IconData: ""},
	}
	got := normalizeLinkEntries(entries, "default")
	if len(got) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(got))
	}
	// A: iconRef non-empty -> upload mode, preserve data
	if got[0].IconMode != "upload" || got[0].IconData == "" || got[0].IconRef == "" {
		t.Error("A should be upload mode with preserved data")
	}
	// B: iconRef empty, iconData matches prefix -> upload mode, but iconRef empty
	if got[1].IconMode != "upload" || got[1].IconData == "" {
		t.Error("B should be upload mode with iconData preserved")
	}
	if got[1].IconRef != "" {
		t.Error("B iconRef should be empty")
	}
	// C: no match -> favicon mode, clear iconData/iconRef
	if got[2].IconMode != "favicon" || got[2].IconData != "" || got[2].IconRef != "" {
		t.Error("C should be favicon mode with cleared data")
	}
	// D: empty -> favicon mode
	if got[3].IconMode != "favicon" {
		t.Error("D should be favicon mode")
	}
}
