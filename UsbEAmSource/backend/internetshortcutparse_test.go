package main

import (
	"reflect"
	"testing"
)

func TestParseInternetShortcutValues_Basic(t *testing.T) {
	content := "[InternetShortcut]\r\nURL=https://example.com\r\nIconFile=C:\\x.ico\r\nIconIndex=1\r\n"
	got := parseInternetShortcutValues(content)
	want := map[string]string{
		"url":       "https://example.com",
		"iconfile":  "C:\\x.ico",
		"iconindex": "1",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseInternetShortcutValues_CommentsSections(t *testing.T) {
	content := "# top comment\n; another comment\n[Header]\nURL=https://x\n\n"
	got := parseInternetShortcutValues(content)
	if len(got) != 1 || got["url"] != "https://x" {
		t.Fatalf("got %#v, want url=https://x only", got)
	}
}

func TestParseInternetShortcutValues_TrimLower(t *testing.T) {
	content := "  URL  =  https://y  \nMiXeD=Value\n"
	got := parseInternetShortcutValues(content)
	if got["url"] != "https://y" {
		t.Fatalf("url = %q, want https://y", got["url"])
	}
	if got["mixed"] != "Value" {
		t.Fatalf("mixed = %q, want Value", got["mixed"])
	}
}

func TestParseInternetShortcutValues_NoEquals(t *testing.T) {
	// 无 '=' 的行跳过；值中含 '=' 取首个分隔符
	got := parseInternetShortcutValues("justatext\nURL=a=b\n")
	if got["url"] != "a=b" {
		t.Fatalf("url = %q, want a=b", got["url"])
	}
}

func TestParseInternetShortcutValues_Empty(t *testing.T) {
	if got := parseInternetShortcutValues(""); len(got) != 0 {
		t.Fatalf("empty -> %#v, want empty map", got)
	}
}
