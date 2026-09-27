// AUTO-RECONSTRUCTED TESTS — gpu_windows.go 纯逻辑助手链
// 研究用途
package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestShouldCleanupMissingGPUPreferencePath(t *testing.T) {
	never := func(string) bool { return false }
	always := func(string) bool { return true }

	cases := []struct {
		name string
		path string
		f    func(string) bool
		want bool
	}{
		{"empty path", "", never, false},
		{"nil predicate", `C:\foo\bar.exe`, nil, false},
		{"relative path", `bar.exe`, never, false},
		{"dll not exe", `C:\foo\bar.dll`, never, false},
		{"predicate true", `C:\foo\bar.exe`, always, false},
		{"predicate false", `C:\foo\bar.exe`, never, true},
	}
	for _, c := range cases {
		if got := shouldCleanupMissingGPUPreferencePath(c.path, c.f); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}

func TestGPUPPreferencePathMayAccessFile(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{``, false},
		{`C:\Program Files\App.exe`, true},
		{`C:/Program Files/App.exe`, true},
		{`C:\Program Files\App.exe `, true},
		{`C:\Program Files\App.dll`, false},
		{`C:\Program Files\noext`, false},
		{`C:\dir.name\file`, false},
		{`\\server\share\app.exe`, false},
		{`\\?\C:\app.exe`, false},
		{`\??\C:\app.exe`, false},
		{`\\.\PhysicalDrive0`, false},
	}
	for _, c := range cases {
		if got := gpuPreferencePathMayAccessFile(c.path); got != c.want {
			t.Errorf("%q: got %v want %v", c.path, got, c.want)
		}
	}
}

func TestShouldReplaceGPUPreferenceEntry(t *testing.T) {
	entry := func(value int, icon, path string, settingsN int) GPUPreferenceEntry {
		return GPUPreferenceEntry{
			PreferenceValue: value,
			Icon:            icon,
			Path:            path,
			Settings:        make([]GPUPreferenceSetting, settingsN),
		}
	}

	cases := []struct {
		name      string
		existing  GPUPreferenceEntry
		candidate GPUPreferenceEntry
		want      bool
	}{
		{"candidate higher value", entry(1, "", `a`, 1), entry(2, "", `b`, 1), true},
		{"candidate lower value", entry(2, "", `a`, 1), entry(1, "", `b`, 1), false},
		{"candidate gains icon", entry(1, "", `a`, 1), entry(1, "x", `b`, 1), true},
		{"candidate gains path", entry(1, "", "", 1), entry(1, "", `b`, 1), true},
		{"more settings wins", entry(1, "", `a`, 1), entry(1, "", `b`, 3), true},
		{"fewer settings loses", entry(1, "", `a`, 3), entry(1, "", `b`, 1), false},
	}
	for _, c := range cases {
		if got := shouldReplaceGPUPreferenceEntry(c.existing, c.candidate); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}

func TestBuildGPUPreferenceEntryEarlyExit(t *testing.T) {
	if _, ok := buildGPUPreferenceEntry("", ""); ok {
		t.Error("empty path should not build")
	}
	if _, ok := buildGPUPreferenceEntry("notepad.exe", ""); ok {
		t.Error("bare name should not build")
	}
}

type fakeFileInfo struct{ isDir bool }

func (f fakeFileInfo) Name() string       { return "" }
func (f fakeFileInfo) Size() int64        { return 0 }
func (f fakeFileInfo) Mode() os.FileMode  { return 0 }
func (f fakeFileInfo) ModTime() time.Time { return time.Time{} }
func (f fakeFileInfo) IsDir() bool        { return f.isDir }
func (f fakeFileInfo) Sys() interface{}   { return nil }

func TestValidateGPUPreferenceTargetPathWithResolver(t *testing.T) {
	noop := func(p string) (string, error) { return p, nil }

	cases := []struct {
		name            string
		path            string
		resolveShortcut func(string) (string, error)
		resolveSymlinks func(string) (string, error)
		stat            func(string) (os.FileInfo, error)
		wantPath        string
		wantErrSubstr   string
	}{
		{
			name:          "empty path",
			path:          "",
			wantErrSubstr: "显卡调度目标路径不能为空",
		},
		{
			name:          "bare name not filesystem",
			path:          "notepad.exe",
			wantErrSubstr: "显卡调度只支持本地程序路径",
		},
		{
			name:          "dll rejected",
			path:          `C:\foo\bar.dll`,
			wantErrSubstr: "显卡调度只支持 .exe 或指向 .exe 的 .lnk",
		},
		{
			name: "lnk resolves to exe",
			path: `C:\foo\bar.lnk`,
			resolveShortcut: func(p string) (string, error) {
				return `C:\foo\bar.exe`, nil
			},
			wantPath: `C:\foo\bar.exe`,
		},
		{
			name: "chained lnk rejected",
			path: `C:\foo\bar.lnk`,
			resolveShortcut: func(p string) (string, error) {
				return `C:\foo\bar2.lnk`, nil
			},
			wantErrSubstr: "显卡调度不支持链式快捷方式",
		},
		{
			name: "symlink resolves to dll",
			path: `C:\foo\bar.exe`,
			resolveSymlinks: func(p string) (string, error) {
				return `C:\foo\actual.dll`, nil
			},
			wantErrSubstr: "显卡调度目标必须是实际 .exe 文件",
		},
		{
			name: "directory rejected",
			path: `C:\foo\bar.exe`,
			stat: func(p string) (os.FileInfo, error) {
				return fakeFileInfo{isDir: true}, nil
			},
			wantErrSubstr: "显卡调度目标不能是目录",
		},
		{
			name: "file accepted",
			path: `C:\foo\bar.exe`,
			stat: func(p string) (os.FileInfo, error) {
				return fakeFileInfo{isDir: false}, nil
			},
			wantPath: `C:\foo\bar.exe`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := validateGPUPreferenceTargetPathWithResolver(
				c.path, c.resolveShortcut, c.resolveSymlinks, c.stat,
			)
			if c.wantErrSubstr != "" {
				if err == nil {
					t.Fatalf("want error containing %q, got nil (path=%q)", c.wantErrSubstr, got)
				}
				if !strings.Contains(err.Error(), c.wantErrSubstr) {
					t.Fatalf("want error containing %q, got %q", c.wantErrSubstr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.wantPath {
				t.Errorf("got path %q want %q", got, c.wantPath)
			}
		})
	}

	// resolveSymlinks nil 分支（默认注入 noop 覆盖到非 nil 分支已验证）：
	if _, err := validateGPUPreferenceTargetPathWithResolver(
		`C:\foo\bar.exe`, nil, nil,
		func(p string) (os.FileInfo, error) { return fakeFileInfo{isDir: false}, nil },
	); err != nil {
		t.Fatalf("nil resolveSymlinks: %v", err)
	}
	_ = noop
}

func TestGPUPPreferenceFileExists(t *testing.T) {
	if gpuPreferenceFileExists("") {
		t.Error("empty path should be false")
	}

	dir := t.TempDir()
	if gpuPreferenceFileExists(dir) {
		t.Error("directory should be false")
	}

	f, err := os.CreateTemp(dir, "probe-*.exe")
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()
	f.Close()
	defer os.Remove(name)
	if !gpuPreferenceFileExists(name) {
		t.Errorf("existing file %q should be true", name)
	}
}

// 批次 156：registry 入口的「前置校验早退」分支不触注册表，可安全做黄金用例。
// 三条入口都在任何 Win32 I/O 之前因空/非法路径返回，验证错误串逐字一致。

func TestAddGPUPreferenceEntryInvalidPath(t *testing.T) {
	_, err := addGPUPreferenceEntry("")
	if err == nil {
		t.Fatal("addGPUPreferenceEntry(\"\") 应报错")
	}
	if !strings.Contains(err.Error(), "显卡调度目标路径不能为空") {
		t.Errorf("addGPUPreferenceEntry 错误串不符: %q", err.Error())
	}
}

func TestSaveGPUPreferenceEntryInvalidPath(t *testing.T) {
	_, err := saveGPUPreferenceEntry("", nil)
	if err == nil {
		t.Fatal("saveGPUPreferenceEntry(\"\", nil) 应报错")
	}
	if !strings.Contains(err.Error(), "显卡调度目标路径不能为空") {
		t.Errorf("saveGPUPreferenceEntry 错误串不符: %q", err.Error())
	}
}

func TestRemoveGPUPreferenceEntryEmptyPath(t *testing.T) {
	_, err := removeGPUPreferenceEntry("")
	if err == nil {
		t.Fatal("removeGPUPreferenceEntry(\"\") 应报错")
	}
	if !strings.Contains(err.Error(), "显卡调度目标路径不能为空") {
		t.Errorf("removeGPUPreferenceEntry 错误串不符: %q", err.Error())
	}
}

// 批次 157：gpu_pick_windows.go 的「零值早退」分支不触任何 Win32 I/O，
// 可安全做黄金用例，验证错误串/返回逐字一致。

func TestResolveProcessNameByPIDZero(t *testing.T) {
	_, err := resolveProcessNameByPID(0)
	if err == nil {
		t.Fatal("resolveProcessNameByPID(0) 应报错")
	}
	if !strings.Contains(err.Error(), "进程 ID 无效") {
		t.Errorf("resolveProcessNameByPID 错误串不符: %q", err.Error())
	}
}

func TestResolveWindowProcessPathZeroHWND(t *testing.T) {
	path, pid, err := resolveWindowProcessPath(0)
	if err == nil {
		t.Fatal("resolveWindowProcessPath(0) 应报错")
	}
	if !strings.Contains(err.Error(), "窗口句柄无效") {
		t.Errorf("resolveWindowProcessPath 错误串不符: %q", err.Error())
	}
	if path != "" || pid != 0 {
		t.Errorf("resolveWindowProcessPath(0) 应返回零值: path=%q pid=%d", path, pid)
	}
}

func TestGetWindowAncestorZeroHWND(t *testing.T) {
	if got := getWindowAncestor(0, 3); got != 0 {
		t.Errorf("getWindowAncestor(0, 3) 应返回 0，got %#x", got)
	}
}
