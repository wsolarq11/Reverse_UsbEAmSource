package main

import (
	"os"
	"path/filepath"
	"testing"
)

// 反汇编实证语义的回归测试（launcherupdate_plan.go 域）

func TestLauncherUpdatePathProtected(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		// Split 首段命中表
		{`data\x.json`, true},
		{`backup\a`, true},
		{`plugin\b`, true},
		{`update\c`, true},
		{`profile_backup\d`, true},
		// Base 名命中
		{`C:\x\usbeam_launcher_config`, true},
		{`C:\x\.usbeam-launcher-home`, true},
		// 首段 .usbeam-update- 前缀（首目录段）
		{`.usbeam-update-abc`, true},
		{`.usbeam-update-abc\x`, true},
		{`C:\x\.usbeam-update-abc`, false}, // 首段为 c:，非保护
		// 不命中
		{`C:\x\apps`, false},
		{`C:\x\usbeam_launcher.exe`, false},
		{`normal.txt`, false},
		{"", false},
	}
	for _, c := range cases {
		if got := launcherUpdatePathProtected(c.path); got != c.want {
			t.Errorf("launcherUpdatePathProtected(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestSamePathFold(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	a := filepath.Join(wd, "Sub", "File.Txt")
	b := filepath.Join(wd, "sub", "file.txt")
	if !samePathFold(a, b) {
		t.Errorf("samePathFold(%q,%q) = false, want true", a, b)
	}
	if samePathFold(a, filepath.Join(wd, "other")) {
		t.Errorf("samePathFold different paths = true, want false")
	}
}

func TestValidateLauncherUpdateAbsolutePath(t *testing.T) {
	wd, _ := os.Getwd()
	ok := filepath.Join(wd, "updates")
	bad := []string{
		"",                               // 空
		"relative/path",                  // 非绝对
		filepath.Join(wd, "a") + `\..\b`, // 未规范化（含 ..）
		`\\server\share\x`,               // UNC
		filepath.Join(wd, "a:b"),         // ADS
	}
	if _, err := validateLauncherUpdateAbsolutePath(ok); err != nil {
		t.Errorf("valid abs path rejected: %v", err)
	}
	for _, p := range bad {
		if _, err := validateLauncherUpdateAbsolutePath(p); err == nil {
			t.Errorf("invalid abs path accepted: %q", p)
		}
	}
}

func TestValidateLauncherUpdateNonce(t *testing.T) {
	valid := "0123456789abcdef0123456789abcdef"
	if err := validateLauncherUpdateNonce(valid, 16); err != nil {
		t.Errorf("valid nonce rejected: %v", err)
	}
	for _, tc := range []struct {
		s string
		n int
	}{
		{valid, 15},
		{valid[:30], 16}, // 15 字节 hex 但期望 16
		{"zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz", 16},
		{"", 16},
	} {
		if err := validateLauncherUpdateNonce(tc.s, tc.n); err == nil {
			t.Errorf("invalid nonce accepted: %q/%d", tc.s, tc.n)
		}
	}
}

func TestNormalizeLauncherUpdateRelativePath(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{`a\b\c.txt`, `a/b/c.txt`, true},
		{`a/b/c.txt`, `a/b/c.txt`, true},
		{"", "", false},
		{`..\x`, "", false},
		{`.\x`, "", false},
		{`../x`, "", false},
		{`a:`, "", false},
		{`data/x`, "", false},            // 受保护
		{`updates\x`, `updates/x`, true}, // 注意 `\x` 是原始字符串
	}
	for _, c := range cases {
		got, err := normalizeLauncherUpdateRelativePath(c.in)
		if c.ok {
			if err != nil {
				t.Errorf("normalize(%q) unexpected err: %v", c.in, err)
			} else if got != c.want {
				t.Errorf("normalize(%q) = %q, want %q", c.in, got, c.want)
			}
		} else if err == nil {
			t.Errorf("normalize(%q) = %q, want error", c.in, got)
		}
	}
}

func TestHashLauncherUpdateFile(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(good, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	sum, err := hashLauncherUpdateFile(good)
	if err != nil {
		t.Fatalf("hash regular file: %v", err)
	}
	// sha256("hello") 已知值
	const want = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if sum != want {
		t.Errorf("hash = %s, want %s", sum, want)
	}
	if _, err := hashLauncherUpdateFile(filepath.Join(dir, "missing")); err == nil {
		t.Errorf("missing file: expected error")
	}
	// 目录视为非常规文件
	if _, err := hashLauncherUpdateFile(dir); err == nil {
		t.Errorf("directory: expected error")
	}
}

func TestUTF16RuneLen(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"abc", 3},
		{"汉字", 2}, // BMP 内每 rune 一单元
		{"a汉", 2},
		{"\U0001F600", 2}, // emoji 代理对
		{"\U0001F600a", 3},
	}
	for _, c := range cases {
		if got := utf16RuneLen(c.in); got != c.want {
			t.Errorf("utf16RuneLen(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}
