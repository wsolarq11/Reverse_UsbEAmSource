// process_launch_windows_test.go — 进程创建链测试（批次 112）
package main

import (
	"errors"
	"syscall"
	"testing"

	"golang.org/x/sys/windows"
)

func TestBuildCreateProcessCommandLine(t *testing.T) {
	cases := []struct {
		exe, args string
		want      string
	}{
		{"C:\\Windows\\System32\\cmd.exe", "", `"C:\Windows\System32\cmd.exe"`},
		{"cmd.exe", "/k pushd X", `"cmd.exe" /k pushd X`},
		{"", "/k pushd X", "/k pushd X"}, // exe 空 → 返回 args
		{"  cmd.exe  ", "  /k  ", `"cmd.exe" /k`},
		{`a"b.exe`, "", `"a\"b.exe"`}, // 内部引号转义
	}
	for _, c := range cases {
		if got := buildCreateProcessCommandLine(c.exe, c.args); got != c.want {
			t.Fatalf("buildCreateProcessCommandLine(%q,%q)=%q，期望 %q", c.exe, c.args, got, c.want)
		}
	}
}

func TestIsElevationRequiredError(t *testing.T) {
	if isElevationRequiredError(nil) {
		t.Fatal("nil 应返回 false")
	}
	if !isElevationRequiredError(windows.ERROR_ELEVATION_REQUIRED) {
		t.Fatal("ERROR_ELEVATION_REQUIRED 应返回 true")
	}
	if !isElevationRequiredError(syscall.Errno(740)) {
		t.Fatal("syscall.Errno(740) 应返回 true")
	}
	if isElevationRequiredError(syscall.Errno(2)) {
		t.Fatal("syscall.Errno(2) 应返回 false")
	}
	if isElevationRequiredError(errors.New("x")) {
		t.Fatal("普通 error 应返回 false")
	}
}

func TestIsProcessHandleElevatedCurrent(t *testing.T) {
	// 冒烟：当前进程伪句柄应能开令牌（不断言提权态，避免 CI 环境差异）。
	if _, err := isProcessHandleElevated(windows.CurrentProcess()); err != nil {
		t.Skipf("当前进程令牌不可用：%v", err)
	}
}

func TestOpenShellProcessForUnelevatedLaunch(t *testing.T) {
	h, err := openShellProcessForUnelevatedLaunch()
	if err != nil {
		t.Skipf("外壳进程不可用（可能无 explorer shell）：%v", err)
	}
	defer windows.CloseHandle(h)
	if h == 0 {
		t.Fatal("成功时应返回非零句柄")
	}
}

func TestBuildShellExecuteArguments(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want string
	}{
		{"空", nil, ""},
		{"空 slice", []string{}, ""},
		{"单参数", []string{"a"}, "a"},
		{"空格分隔", []string{"a", "b", "c"}, "a b c"},
		{"含空格参数", []string{"a", "b c", "d"}, `a "b c" d`},
		{"空参数跳过", []string{"a", "", "  ", "b"}, "a b"},
		{"首尾空白修剪", []string{"  a  ", "b"}, "a b"},
	}
	for _, c := range cases {
		if got := buildShellExecuteArguments(c.in); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestResolveApplicationWorkingDirectory(t *testing.T) {
	cases := []struct {
		name      string
		entryType string
		entry     string
		appName   string
		want      string
	}{
		{"目录条目", "directory", "x", "y", ""},
		{"目录条目大小写", "Directory", "x", "y", ""},
		{"entry 非空", "app", "C:\\work\\dir", "app.exe", "C:\\work\\dir"},
		{"entry 空白修剪", "app", "  C:\\work  ", "app.exe", "C:\\work"},
		{"appName 推断", "app", "", "C:\\tools\\app.exe", "C:\\tools"},
		{"appName 无路径", "app", "", "app.exe", ""},
		{"全空", "app", "", "", ""},
	}
	for _, c := range cases {
		ctx := launchContext{entryType: c.entryType, entry: c.entry, shellTarget: c.appName}
		if got := resolveApplicationWorkingDirectory(ctx); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}
