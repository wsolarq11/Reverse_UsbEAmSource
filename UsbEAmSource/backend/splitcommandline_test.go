package main

import (
	"reflect"
	"testing"
)

func TestSplitCommandLineArguments_Empty(t *testing.T) {
	if got := splitCommandLineArguments(""); got != nil {
		t.Fatalf("empty -> %v, want nil", got)
	}
	if got := splitCommandLineArguments("   "); got != nil {
		t.Fatalf("blank -> %v, want nil", got)
	}
}

func TestSplitCommandLineArguments_Basic(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"a b c", []string{"a", "b", "c"}},
		{"a  b  c", []string{"a", "b", "c"}},
		{"--flag value", []string{"--flag", "value"}},
	}
	for _, c := range cases {
		if got := splitCommandLineArguments(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("split(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestSplitCommandLineArguments_Quotes(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{`"C:\Program Files\app.exe" arg`, []string{`C:\Program Files\app.exe`, "arg"}},
		{`say "hello world"`, []string{"say", "hello world"}},
		{`'a b' c`, []string{"a b", "c"}},
		{`"" empty`, []string{"empty"}},
	}
	for _, c := range cases {
		if got := splitCommandLineArguments(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("split(%q) = %#v, want %#v", c.in, got, c.want)
		}
	}
}

func TestSplitCommandLineArguments_Backslash(t *testing.T) {
	// 反斜杠后引号：偶数反斜杠 → 翻转引号，写一半；奇数 → 写一个字面引号
	cases := []struct {
		in   string
		want []string
	}{
		{`a\b c`, []string{`a\b`, "c"}}, // 反斜杠未后随引号 → 保留
		{`x\"y`, []string{`x"y`}},       // 奇数 \ 后 " → 字面引号
		{`x\\"y`, []string{"x\\y"}},     // 偶数 \ + " toggle（写 1 \）
		{`"a\"b"`, []string{`a"b`}},     // 引号内转义引号
	}
	for _, c := range cases {
		if got := splitCommandLineArguments(c.in); !reflect.DeepEqual(got, c.want) {
			t.Errorf("split(%q) = %#v, want %#v", c.in, got, c.want)
		}
	}
}

func TestSplitCommandLineArguments_Mixed(t *testing.T) {
	got := splitCommandLineArguments(`--name "My App" --path C:\dir\f.bin`)
	want := []string{"--name", "My App", "--path", `C:\dir\f.bin`}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}
