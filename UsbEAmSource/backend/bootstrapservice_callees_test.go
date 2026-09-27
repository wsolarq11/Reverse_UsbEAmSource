// AUTO-RECONSTRUCTED TESTS — bootstrapservice_callees 批次 41 落档函数
// 研究用途
package main

import "testing"

func TestLanguageDirectoriesForWorkspace(t *testing.T) {
	ws := WorkspaceLayout{LanguageDir: "  base  ", AppLanguageDir: "  app "}
	got := languageDirectoriesForWorkspace(ws)
	if len(got) != 1 || got[0] != "app" {
		t.Fatalf("AppLanguageDir 优先：got %v", got)
	}

	ws.AppLanguageDir = "   "
	got = languageDirectoriesForWorkspace(ws)
	if len(got) != 1 || got[0] != "base" {
		t.Fatalf("回退 LanguageDir：got %v", got)
	}

	ws.LanguageDir = "  "
	if got = languageDirectoriesForWorkspace(ws); got != nil {
		t.Fatalf("双空应返 nil：got %v", got)
	}
}

func TestIsSupportedAppPath(t *testing.T) {
	cases := map[string]bool{
		"a.exe":          true,
		"b.LNK":          true,
		"c.url":          true,
		"d.appref-ms":    true,
		"e.txt":          false,
		"dir/noext":      false,
		" C:\\x.exe ":    true,
		"folder":         false,
		"a.b.exe":        true,
		"dir.with.dot/f": false,
	}
	for in, want := range cases {
		if got := isSupportedAppPath(in); got != want {
			t.Errorf("isSupportedAppPath(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestLauncherConfigsEqual(t *testing.T) {
	a, b := LauncherConfig{}, LauncherConfig{}
	if !launcherConfigsEqual(a, b) {
		t.Fatal("零值配置应相等")
	}
	b.Apps = []AppEntry{{ID: "x"}}
	if launcherConfigsEqual(a, b) {
		t.Fatal("Apps 不同应不相等")
	}
}
