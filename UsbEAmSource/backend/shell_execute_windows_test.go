// shell_execute_windows_test.go — Shell 执行域测试（批次 107）
package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func TestUTF16PtrOrNil(t *testing.T) {
	p, err := utf16PtrOrNil("")
	if err != nil || p != nil {
		t.Fatalf("空串应返回 (nil,nil)，得到 (%v,%v)", p, err)
	}
	p, err = utf16PtrOrNil("   ")
	if err != nil || p != nil {
		t.Fatalf("全空白应返回 (nil,nil)，得到 (%v,%v)", p, err)
	}
	p, err = utf16PtrOrNil("  abc ")
	if err != nil || p == nil {
		t.Fatalf("非空应返回非 nil，得到 (%v,%v)", p, err)
	}
	if got := windows.UTF16PtrToString(p); got != "abc" {
		t.Fatalf("UTF16 内容异常：%q", got)
	}
}

func TestQuoteCmdArgument(t *testing.T) {
	cases := []struct{ in, want string }{
		{"abc", `"abc"`},
		{`a"b`, `"a""b"`},
		{"  abc  ", `"abc"`},
		{"", `""`},
	}
	for _, c := range cases {
		if got := quoteCmdArgument(c.in); got != c.want {
			t.Fatalf("quoteCmdArgument(%q)=%q，期望 %q", c.in, got, c.want)
		}
	}
}

func TestResolveCmdExePath(t *testing.T) {
	t.Setenv("ComSpec", `C:\Win\cmd.exe`)
	if got := resolveCmdExePath(); got != `C:\Win\cmd.exe` {
		t.Fatalf("ComSpec 命中异常：%q", got)
	}
	t.Setenv("ComSpec", "")
	t.Setenv("SystemRoot", `C:\Win`)
	want := filepath.Join(`C:\Win`, "System32", "cmd.exe")
	if got := resolveCmdExePath(); got != want {
		t.Fatalf("SystemRoot 拼接异常：%q 期望 %q", got, want)
	}
	t.Setenv("SystemRoot", "")
	if got := resolveCmdExePath(); got != "cmd.exe" {
		t.Fatalf("兜底异常：%q", got)
	}
}

func TestEncodePowerShellCommand(t *testing.T) {
	got := encodePowerShellCommand("abc")
	decoded, err := base64.StdEncoding.DecodeString(got)
	if err != nil {
		t.Fatalf("base64 解码失败：%v", err)
	}
	want := []byte{'a', 0, 'b', 0, 'c', 0}
	if !bytes.Equal(decoded, want) {
		t.Fatalf("UTF-16LE 字节异常：%v 期望 %v", decoded, want)
	}
}

func TestWithShellApartmentPassthrough(t *testing.T) {
	if err := withShellApartment(func() error { return nil }); err != nil {
		t.Fatalf("nil 透传异常：%v", err)
	}
	wantErr := errors.New("boom")
	if err := withShellApartment(func() error { return wantErr }); err != wantErr {
		t.Fatalf("错误透传异常：%v", err)
	}
}
