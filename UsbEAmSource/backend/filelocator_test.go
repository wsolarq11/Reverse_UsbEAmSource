package main

import "testing"

func TestNewFileLocatorService(t *testing.T) {
	s := newFileLocatorService()
	if s == nil {
		t.Fatal("factory must return non-nil")
	}
	// 装配：detailByPath map、contentScan chan 已建。
	if s.detailByPath == nil {
		t.Error("detailByPath map must be initialized")
	}
	if s.contentScan == nil {
		t.Error("contentScan chan must be initialized")
	}
	// state 默认配置 + 结果上限 0x1f4=500。
	if s.state.ResultLimit != 0x1f4 {
		t.Errorf("ResultLimit=%d want 500", s.state.ResultLimit)
	}
	// Request 来自 defaultFileLocatorConfig（结构已装配，即使字符串为 [P] 近似）。
	if s.state.Request.MaxSearchFileSizeMB != 500 {
		t.Errorf("default maxSearchFileSizeMB=%d want 500", s.state.Request.MaxSearchFileSizeMB)
	}
}

func TestNewFileLocatorGlobPathMatcher(t *testing.T) {
	if m := newFileLocatorGlobPathMatcher(""); m != nil {
		t.Errorf("空 pattern 应返 nil, got %+v", m)
	}
	if m := newFileLocatorGlobPathMatcher("   "); m != nil {
		t.Errorf("纯空白 pattern 应返 nil, got %+v", m)
	}
	m := newFileLocatorGlobPathMatcher("  *.go  ")
	if m == nil {
		t.Fatal("非空 glob 应返非 nil")
	}
	if m.query != "*.go" {
		t.Errorf("query=%q want \"*.go\"（TrimSpace 生效）", m.query)
	}
	if m.mode != "regex" {
		t.Errorf("mode=%q want \"regex\"", m.mode)
	}
}

func TestSplitFileLocatorFilterRuleType(t *testing.T) {
	cases := []struct {
		in      string
		typ     string
		pattern string
		ok      bool
	}{
		{"g:*.go", "glob", "*.go", true},
		{"b:x", "boolean", "x", true},
		{"p:abc", "plain", "abc", true},
		{"r:^foo", "regex", "^foo", true},
		{"G:*.txt", "glob", "*.txt", true}, // 首字符大小写不敏感
		{"g:", "glob", "", true},
		{"*.go", "", "*.go", false},
		{"x:foo", "", "x:foo", false},
		{"", "", "", false},
		{"  g:*.h  ", "glob", "*.h", true}, // 整体 TrimSpace
	}
	for _, c := range cases {
		typ, pattern, ok := splitFileLocatorFilterRuleType(c.in)
		if typ != c.typ || pattern != c.pattern || ok != c.ok {
			t.Errorf("split(%q)=%q,%q,%v want %q,%q,%v", c.in, typ, pattern, ok, c.typ, c.pattern, c.ok)
		}
	}
}

func TestParseFileLocatorFilterRule(t *testing.T) {
	cases := []struct {
		rule, ft string
		include  bool
		typ, pat string
	}{
		{"*.go", "glob", true, "glob", "*.go"},
		{"!*.go", "glob", false, "glob", "*.go"},
		{"-*.go", "glob", false, "glob", "*.go"},
		{"+*.go", "glob", true, "glob", "*.go"},
		{"g:*.go", "plain", true, "glob", "*.go"}, // 前缀覆盖类型
		{"", "glob", true, "glob", ""},
		{"  !g:*.md  ", "plain", false, "glob", "*.md"},
	}
	for _, c := range cases {
		include, typ, pat := parseFileLocatorFilterRule(c.rule, c.ft)
		if include != c.include || typ != c.typ || pat != c.pat {
			t.Errorf("parse(%q,%q)=%v,%q,%q want %v,%q,%q", c.rule, c.ft, include, typ, pat, c.include, c.typ, c.pat)
		}
	}
}
