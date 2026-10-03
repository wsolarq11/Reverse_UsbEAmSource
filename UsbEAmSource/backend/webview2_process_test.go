package main

import (
	"math"
	"testing"
)

func TestWebView2ProcessKindSortWeight(t *testing.T) {
	cases := []struct {
		kind string
		want int
	}{
		{"browser", 0},
		{"renderer", 1},
		{"gpu-process", 2},
		{"utility", 3},
		{"unknown", 4},
		{"", 4},
		{" Browser ", 0},
		{"RENDERER", 1},
		{"Gpu-Process", 2},
	}
	for _, c := range cases {
		if got := webView2ProcessKindSortWeight(c.kind); got != c.want {
			t.Errorf("webView2ProcessKindSortWeight(%q) = %d, want %d", c.kind, got, c.want)
		}
	}
}

func TestExtractWebView2ProcessType(t *testing.T) {
	cases := []struct {
		cmd  string
		want string
	}{
		{"", "browser"},
		{"no type here", "browser"},
		{"--type=renderer", "renderer"},
		{"--type=gpu-process --foo=bar", "gpu-process"},
		{`--type="utility"`, "utility"},
		{"--TYPE=Renderer", "renderer"},
		{"--type=browser", "browser"},
		{"--type= --x", ""}, // token 为空串时返回空
	}
	for _, c := range cases {
		if got := extractWebView2ProcessType(c.cmd); got != c.want {
			t.Errorf("extractWebView2ProcessType(%q) = %q, want %q", c.cmd, got, c.want)
		}
	}
}

func TestNormalizeWebView2ProcessKind(t *testing.T) {
	cases := []struct {
		kind, cmd, want string
	}{
		{"browser", "", "browser"},
		{"", "--type=renderer", "renderer"},
		{"", "no type", "browser"},
		{"", "", "browser"},
		{"utility", "", "utility"},
		{" Browser ", "", "browser"},
		{"gpu-process", "", "gpu-process"},
		{"", "--type=gpu-process", "gpu-process"},
	}
	for _, c := range cases {
		if got := normalizeWebView2ProcessKind(c.kind, c.cmd); got != c.want {
			t.Errorf("normalizeWebView2ProcessKind(%q, %q) = %q, want %q", c.kind, c.cmd, got, c.want)
		}
	}
}

func TestNormalizeWebView2ProcessInfo(t *testing.T) {
	const mb = 1024 * 1024
	info := WebView2ProcessInfo{
		ProcessID:               123,
		Kind:                    " renderer ",
		WorkingSetBytes:         2 * mb,
		PrivateBytes:            mb,
		CommandLine:             "--js-flags=--expose-gc --type=renderer",
		HasRendererProcessLimit: true,
	}
	got := normalizeWebView2ProcessInfo(info)

	if got.Kind != "renderer" {
		t.Errorf("Kind = %q, want renderer", got.Kind)
	}
	if math.Abs(got.WorkingSetMB-2.0) > 1e-9 {
		t.Errorf("WorkingSetMB = %v, want 2.0", got.WorkingSetMB)
	}
	if math.Abs(got.PrivateMB-1.0) > 1e-9 {
		t.Errorf("PrivateMB = %v, want 1.0", got.PrivateMB)
	}
	if !got.HasJSFlags {
		t.Errorf("HasJSFlags = false, want true (substring present)")
	}
	if got.HasDisableFeatures {
		t.Errorf("HasDisableFeatures = true, want false")
	}
	if !got.HasRendererProcessLimit {
		t.Errorf("HasRendererProcessLimit = false, want true (pre-set)")
	}
	if got.WorkingSetBytes != 2*mb || got.PrivateBytes != mb {
		t.Errorf("byte fields mutated: %d/%d", got.WorkingSetBytes, got.PrivateBytes)
	}
}

func TestNormalizeWebView2ProcessSnapshot(t *testing.T) {
	const mb = 1024 * 1024
	snap := WebView2ProcessSnapshot{
		Processes: []WebView2ProcessInfo{
			{ProcessID: 30, Kind: "utility", WorkingSetBytes: 3 * mb, PrivateBytes: 3 * mb},
			{ProcessID: 10, Kind: "browser", WorkingSetBytes: 1 * mb, PrivateBytes: 1 * mb},
			{ProcessID: 20, Kind: "renderer", WorkingSetBytes: 2 * mb, PrivateBytes: 2 * mb},
			{ProcessID: 40, Kind: "gpu-process", WorkingSetBytes: 4 * mb, PrivateBytes: 4 * mb},
		},
	}
	got := normalizeWebView2ProcessSnapshot(snap)

	// 权重升序：browser=0, renderer=1, gpu-process=2, utility=3
	wantKinds := []string{"browser", "renderer", "gpu-process", "utility"}
	for i, k := range wantKinds {
		if got.Processes[i].Kind != k {
			t.Fatalf("Processes[%d].Kind = %q, want %q", i, got.Processes[i].Kind, k)
		}
	}
	if got.TotalWorkingSetBytes != 10*mb || got.TotalPrivateBytes != 10*mb {
		t.Errorf("totals = %d/%d, want 10MB/10MB", got.TotalWorkingSetBytes, got.TotalPrivateBytes)
	}
	if math.Abs(got.TotalWorkingSetMB-10.0) > 1e-9 || math.Abs(got.TotalPrivateMB-10.0) > 1e-9 {
		t.Errorf("total MB = %v/%v, want 10/10", got.TotalWorkingSetMB, got.TotalPrivateMB)
	}
}

func TestNormalizeWebView2ProcessSnapshotNilAndTieBreak(t *testing.T) {
	got := normalizeWebView2ProcessSnapshot(WebView2ProcessSnapshot{})
	if got.Processes == nil {
		t.Errorf("Processes = nil, want empty non-nil slice")
	}
	if len(got.Processes) != 0 {
		t.Errorf("len(Processes) = %d, want 0", len(got.Processes))
	}

	// 权重相等时按 ProcessID 升序
	snap := WebView2ProcessSnapshot{
		Processes: []WebView2ProcessInfo{
			{ProcessID: 5, Kind: "renderer"},
			{ProcessID: 2, Kind: "renderer"},
		},
	}
	got = normalizeWebView2ProcessSnapshot(snap)
	if got.Processes[0].ProcessID != 2 || got.Processes[1].ProcessID != 5 {
		t.Errorf("tie-break order = %d,%d, want 2,5", got.Processes[0].ProcessID, got.Processes[1].ProcessID)
	}
}
