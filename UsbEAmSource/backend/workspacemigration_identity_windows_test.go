package main

import "testing"

func TestNormalizeWorkspaceWindowsFinalPath(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"普通路径", `C:\foo\bar`, `C:\foo\bar`},
		{"设备长路径前缀", `\\?\C:\foo\bar`, `C:\foo\bar`},
		{"UNC 长路径前缀", `\\?\UNC\server\share\dir`, `\\server\share\dir`},
		{"设备前缀大小写不敏感", `\\?\c:\Foo`, `c:\Foo`},
		{"UNC 前缀大小写不敏感", `\\?\unc\Server\Share`, `\\Server\Share`},
		{"首尾空白去除", `  C:\foo  `, `C:\foo`},
		{"仅前缀无余量", `\\?\`, `.`},
		{"空串", ``, `.`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeWorkspaceWindowsFinalPath(tc.in); got != tc.want {
				t.Fatalf("normalizeWorkspaceWindowsFinalPath(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
