// pathclip_logic_windows_test.go — 路径/可执行目标判断链测试（批次 108）
package main

import (
	"testing"

	"golang.org/x/sys/windows"
)

func TestNewPathError(t *testing.T) {
	err := newPathError()
	if err == nil {
		t.Fatal("newPathError 应返回非 nil 错误")
	}
	if err.Error() != "路径不能为空" {
		t.Fatalf("错误消息异常：%q", err.Error())
	}
}

func TestIsWindowsAbsoluteFilesystemPath(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{`C:\foo`, true},
		{`C:/foo`, true},
		{`c:\foo`, true},
		{`\\server\share`, true},
		{`//server/share`, true},
		{` C:\foo `, true},
		{"", false},
		{"relative", false},
		{"1C:\\foo", false}, // 首字符非字母
		{"C:foo", false},    // 盘符后无斜杠
		{"C", false},
	}
	for _, c := range cases {
		if got := isWindowsAbsoluteFilesystemPath(c.in); got != c.want {
			t.Fatalf("isWindowsAbsoluteFilesystemPath(%q)=%v，期望 %v", c.in, got, c.want)
		}
	}
}

func TestHasKnownWindowsExecutableExtension(t *testing.T) {
	positive := []string{
		`C:\a.exe`, `C:\dir\a.bat`, `a.cmd`, `a.com`, `a.scr`, `a.vbs`, `a.vbe`,
		`a.ws`, `a.vb`, `a.ps1`, `a.msi`, `a.msp`, `a.msc`, `a.jse`, `a.wsf`,
		`a.wsh`, `a.hta`, `a.cpl`, `a.lnk`, `a.pif`, `a.url`, `a.appref-ms`,
		`C:\A.EXE`, // 大写经 ToLower 命中
	}
	for _, p := range positive {
		if !hasKnownWindowsExecutableExtension(p) {
			t.Fatalf("hasKnownWindowsExecutableExtension(%q) 应 true", p)
		}
	}
	negative := []string{
		"", "a.txt", "a.js", "a.dll", "a.exe.backup", `C:\a.txt`, "noextension",
	}
	for _, p := range negative {
		if hasKnownWindowsExecutableExtension(p) {
			t.Fatalf("hasKnownWindowsExecutableExtension(%q) 应 false", p)
		}
	}
}

func TestGetTokenElevation(t *testing.T) {
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token); err != nil {
		t.Skipf("OpenProcessToken 不可用：%v", err)
	}
	defer token.Close()
	elevated, err := getTokenElevation(token)
	if err != nil {
		t.Fatalf("getTokenElevation 报错：%v", err)
	}
	// 结果仅作类型与可调用性验证，不校验具体布尔值（依赖运行环境权限）。
	_ = elevated
}
