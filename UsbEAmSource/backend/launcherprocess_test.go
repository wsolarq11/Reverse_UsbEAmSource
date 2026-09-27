// launcherprocess_test.go — 启动上下文解析域测试（批次 121）
package main

import (
	"os"
	"path/filepath"
	"testing"
	"unsafe"
)

// TestLaunchContextSize 锁定结构体字节布局（360B=0x168，45 词）。
// 字段偏移经 resolveLaunchableAppEntry / startApplicationViaExplorer /
// startApplicationUsingCurrentPrivileges 三函数 asm 交叉实证。
func TestLaunchContextSize(t *testing.T) {
	if got := unsafe.Sizeof(launchContext{}); got != 360 {
		t.Fatalf("launchContext size=%d，期望 360", got)
	}
}

func TestShouldUseSavedShortcutResolution(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"", true},
		{"   ", true},
		{"foo.txt", false},    // 非快捷方式
		{"foo.exe", false},    // 非快捷方式
		{"missing.lnk", true}, // 快捷方式且文件不存在 → true
	}
	for _, c := range cases {
		if got := shouldUseSavedShortcutResolution(c.path); got != c.want {
			t.Fatalf("shouldUseSavedShortcutResolution(%q)=%v，期望 %v", c.path, got, c.want)
		}
	}
	// 存在的 .lnk → Stat 成功 → false
	dir := t.TempDir()
	lnk := filepath.Join(dir, "real.lnk")
	if err := os.WriteFile(lnk, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := shouldUseSavedShortcutResolution(lnk); got != false {
		t.Fatalf("存在 .lnk 应返回 false，got=%v", got)
	}
}

func TestResolveLaunchableAppEntryDirectory(t *testing.T) {
	ctx := launchContext{entryType: "directory", rawName: "C:\\x", shellTarget: "old"}
	got := resolveLaunchableAppEntry(ctx)
	if got.shellTarget != "old" {
		t.Fatalf("directory 条目应原样返回，shellTarget=%q", got.shellTarget)
	}
}

func TestResolveLaunchableAppEntryEmptyRawName(t *testing.T) {
	ctx := launchContext{entryType: "app", rawName: "  ", shellTarget: "old"}
	got := resolveLaunchableAppEntry(ctx)
	if got.shellTarget != "old" {
		t.Fatalf("空 rawName 应原样返回，shellTarget=%q", got.shellTarget)
	}
}

func TestResolveLaunchableAppEntryNonShortcutTarget(t *testing.T) {
	// shellTarget 非快捷方式 → shouldUseSavedShortcutResolution false → 原样返回
	ctx := launchContext{entryType: "app", rawName: "C:\\real.exe", shellTarget: "foo.txt"}
	got := resolveLaunchableAppEntry(ctx)
	if got.shellTarget != "foo.txt" {
		t.Fatalf("非快捷方式 shellTarget 应原样返回，got=%q", got.shellTarget)
	}
}

func TestResolveLaunchableAppEntrySavedBranch(t *testing.T) {
	ctx := launchContext{
		entryType:      "app",
		shellTarget:    "broken.lnk", // 失效快捷方式 → 走 saved 分支
		entry:          "   ",
		args:           nil,
		rawName:        "C:\\real.exe",
		rawEntry:       "C:\\real.lnk",
		rawCommandLine: `"C:\real.exe" --flag`,
	}
	got := resolveLaunchableAppEntry(ctx)
	if got.shellTarget != "C:\\real.exe" {
		t.Fatalf("shellTarget=%q，期望 C:\\real.exe", got.shellTarget)
	}
	if got.entry != "C:\\real.lnk" {
		t.Fatalf("entry=%q，期望 C:\\real.lnk（rawEntry 回退）", got.entry)
	}
	wantArgs := []string{`C:\real.exe`, "--flag"}
	if len(got.args) != 2 || got.args[0] != wantArgs[0] || got.args[1] != wantArgs[1] {
		t.Fatalf("args=%v，期望 %v（rawCommandLine 分割）", got.args, wantArgs)
	}
}

func TestBuildAppEntryWithDroppedFilesDirectory(t *testing.T) {
	ctx := launchContext{entryType: "directory", args: []string{"keep"}}
	got := buildAppEntryWithDroppedFiles([]string{"f1.txt"}, ctx)
	if len(got.args) != 1 || got.args[0] != "keep" {
		t.Fatalf("directory 应原样返回，args=%v", got.args)
	}
}

func TestBuildAppEntryWithDroppedFilesEmpty(t *testing.T) {
	ctx := launchContext{entryType: "app", args: []string{"keep"}}
	got := buildAppEntryWithDroppedFiles(nil, ctx)
	if len(got.args) != 1 || got.args[0] != "keep" {
		t.Fatalf("空 droppedFiles 应原样返回，args=%v", got.args)
	}
}

func TestBuildAppEntryWithDroppedFilesMerge(t *testing.T) {
	ctx := launchContext{
		entryType:   "app",
		shellTarget: "C:\\app.exe", // requiresShellOpen false → 不触发显式参数分支
		args:        nil,
	}
	got := buildAppEntryWithDroppedFiles([]string{"f1.txt", "f2.txt"}, ctx)
	if len(got.args) != 2 || got.args[0] != "f1.txt" || got.args[1] != "f2.txt" {
		t.Fatalf("args=%v，期望合并 droppedFiles [f1.txt f2.txt]", got.args)
	}
}
