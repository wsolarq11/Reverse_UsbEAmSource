package main

import "testing"

func TestNormalizeFileEntries_Basic(t *testing.T) {
	if got := normalizeFileEntries(nil); got != nil {
		t.Fatal("nil in -> nil out")
	}
	if got := normalizeFileEntries([]FileEntry{}); len(got) != 0 {
		t.Fatalf("empty in -> empty out, got %d", len(got))
	}
}

func TestNormalizeFileEntries_SkipEmptyPath(t *testing.T) {
	// Path 空条目被跳过
	got := normalizeFileEntries([]FileEntry{{Path: "  "}, {Path: "C:\\a\\b.txt"}})
	if len(got) != 1 {
		t.Fatalf("expected 1 kept, got %d", len(got))
	}
}

func TestNormalizeFileEntries_NameFallback(t *testing.T) {
	// Name 空 → 从 Path 推导 basename
	got := normalizeFileEntries([]FileEntry{{Path: "C:\\dir\\report.pdf"}})
	if got[0].Name != "report" {
		t.Fatalf("Name = %q, want report", got[0].Name)
	}
	// 字段 Trim
	got2 := normalizeFileEntries([]FileEntry{{Name: "  Doc  ", Path: "  C:\\x  "}})
	if got2[0].Name != "Doc" || got2[0].Path != "C:\\x" {
		t.Fatalf("not trimmed: %+v", got2[0])
	}
}

func TestNormalizeFileEntries_IDSeq(t *testing.T) {
	// ID 候选 [ID, Name, Path]，最后非空 = Path → 以 Path 为基
	got := normalizeFileEntries([]FileEntry{
		{Path: "/p/a"},
		{Path: "/p/a"},
		{Path: "/p/b"},
	})
	if len(got) != 3 {
		t.Fatalf("want 3, got %d", len(got))
	}
	if got[0].ID != "p-a" || got[1].ID != "p-a-2" || got[2].ID != "p-b" {
		t.Fatalf("IDs = [%s %s %s], want [p-a p-a-2 p-b]", got[0].ID, got[1].ID, got[2].ID)
	}
}

func TestNormalizeFileEntries_TagsClean(t *testing.T) {
	// Tags 经 cleanStringList：Trim、去空、去重（键不分大小写）
	got := normalizeFileEntries([]FileEntry{{Path: "/x", Tags: []string{" Tag ", "  ", "tag", "b"}}})
	if len(got[0].Tags) != 2 {
		t.Fatalf("Tags = %v, want 2 entries", got[0].Tags)
	}
	if got[0].Tags[0] != "Tag" || got[0].Tags[1] != "b" {
		t.Fatalf("Tags = %v, want [Tag b]", got[0].Tags)
	}
}
