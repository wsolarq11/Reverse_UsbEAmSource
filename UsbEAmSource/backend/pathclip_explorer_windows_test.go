// pathclip_explorer_windows_test.go — 资源管理器启动 + 参数转义测试（批次 109）
package main

import "testing"

func TestQuoteWindowsArgument(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", `""`},
		{"abc", "abc"}, // 无特殊字符原样
		{`a\b`, `a\b`}, // 反斜杠本身不触发
		{`a\`, `a\`},   // 结尾反斜杠无特殊字符 → 原样
		{`\\server\share`, `\\server\share`},
		{"a b", `"a b"`},     // 空格 → 包裹
		{`a"b`, `"a\"b"`},    // 引号 → 包裹 + 转义
		{`a\"b`, `"a\\\"b"`}, // 反斜杠后引号 → 翻倍 + 转义
		{`a\ b`, `"a\ b"`},   // 反斜杠后空格 → 反斜杠保留
		{`a b\`, `"a b\\"`},  // 空格 + 结尾反斜杠 → 结尾翻倍
		{`C:\Program Files\App`, `"C:\Program Files\App"`},
		{`a"b"c`, `"a\"b\"c"`},
		{`trailing space `, `"trailing space "`},
	}
	for _, c := range cases {
		if got := quoteWindowsArgument(c.in); got != c.want {
			t.Fatalf("quoteWindowsArgument(%q)=%q，期望 %q", c.in, got, c.want)
		}
	}
}

func TestQuoteWindowsArgumentUnicode(t *testing.T) {
	got := quoteWindowsArgument("路径 with space")
	if got[0] != '"' || got[len(got)-1] != '"' {
		t.Fatalf("含空格 Unicode 参数应被引号包裹：%q", got)
	}
	if got := quoteWindowsArgument("纯路径"); got != "纯路径" {
		t.Fatalf("无特殊字符 Unicode 参数应原样返回：%q", got)
	}
}

// TestStartShellTargetViaExplorerEmpty 空目标分支（批次 120）：
// TrimSpace 空 → errors.New("Shell 目标不能为空")。非空分支触发真实 COM 链，离线不可执行。
func TestStartShellTargetViaExplorerEmpty(t *testing.T) {
	for _, in := range []string{"", "   ", "\t\n"} {
		err := startShellTargetViaExplorer(in, nil, "")
		if err == nil {
			t.Fatalf("startShellTargetViaExplorer(%q) 应返回错误", in)
		}
		if err.Error() != "Shell 目标不能为空" {
			t.Fatalf("startShellTargetViaExplorer(%q) 错误=%q，期望 %q", in, err.Error(), "Shell 目标不能为空")
		}
	}
}
