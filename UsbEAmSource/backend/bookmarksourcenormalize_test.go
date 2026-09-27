package main

import "testing"

func TestInferBookmarkSourceName(t *testing.T) {
	cases := []struct {
		a, b, c, want string
	}{
		{"Chrome", "", "C:\\nope", "Chrome"},
		{"Chrome", "Chromium", "C:\\x", "Chrome / Chromium"},
		{"", "Edge", "C:\\y", "Edge"},
		{"", "", `C:\profiles\data`, "profiles"},
		{"", "", `C:\`, ""},
	}
	for _, tc := range cases {
		if got := inferBookmarkSourceName(tc.a, tc.b, tc.c); got != tc.want {
			t.Errorf("inferBookmarkSourceName(%q,%q,%q) = %q, want %q", tc.a, tc.b, tc.c, got, tc.want)
		}
	}
}

func TestNormalizeBookmarkSources_Basic(t *testing.T) {
	// 空输入 → nil
	if got := normalizeBookmarkSources(nil); got != nil {
		t.Fatal("nil in -> nil out")
	}
	if got := normalizeBookmarkSources([]BookmarkSource{}); len(got) != 0 {
		t.Fatalf("empty in -> empty out, got %d", len(got))
	}
}

func TestNormalizeBookmarkSources_SkipEmptyFields(t *testing.T) {
	// 全空字段条目被跳过
	got := normalizeBookmarkSources([]BookmarkSource{{}, {Name: "  Foo "}})
	if len(got) != 1 {
		t.Fatalf("expected 1 kept, got %d", len(got))
	}
	// 字段被 TrimSpace
	if got[0].Name != "Foo" {
		t.Fatalf("Name = %q, want Foo", got[0].Name)
	}
	// ID 由 allocator 生成（首次裸 slug）
	if got[0].ID != "foo" {
		t.Fatalf("ID = %q, want foo", got[0].ID)
	}
}

func TestNormalizeBookmarkSources_IDSeq(t *testing.T) {
	got := normalizeBookmarkSources([]BookmarkSource{
		{Name: "Same", Browser: "B", Path: "SamePath"},
		{Name: "Same", Browser: "B", Path: "SamePath"},
		{Name: "Diff", Browser: "B", Path: "OtherPath"},
	})
	if len(got) != 3 {
		t.Fatalf("want 3 entries, got %d", len(got))
	}
	if got[0].ID != "samepath" || got[1].ID != "samepath-2" || got[2].ID != "otherpath" {
		t.Fatalf("IDs = [%s %s %s], want [samepath samepath-2 otherpath]", got[0].ID, got[1].ID, got[2].ID)
	}
}

func TestNormalizeBookmarkSources_Enabled(t *testing.T) {
	// 未启用且无 ID → 置启用并推导 Name
	got := normalizeBookmarkSources([]BookmarkSource{{Browser: " Edge ", Path: "C:\\pf", Enabled: false}})
	if len(got) != 1 {
		t.Fatalf("want 1, got %d", len(got))
	}
	if !got[0].Enabled {
		t.Fatal("expected Enabled=true after derive")
	}
	if got[0].Name != "Edge" {
		t.Fatalf("derived Name = %q, want Edge", got[0].Name)
	}
	// 已启用则保留状态
	got2 := normalizeBookmarkSources([]BookmarkSource{{Name: " X ", Enabled: true}})
	if !got2[0].Enabled {
		t.Fatal("expected enabled retained")
	}
}

func TestNormalizeBookmarkSources_TrimAll(t *testing.T) {
	got := normalizeBookmarkSources([]BookmarkSource{{Name: "  A  ", Browser: "  B  ", Path: "  P  "}})
	if got[0].Name != "A" || got[0].Browser != "B" || got[0].Path != "P" {
		t.Fatalf("fields not trimmed: %+v", got[0])
	}
}
