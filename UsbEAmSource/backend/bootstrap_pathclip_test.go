// bootstrap_pathclip_test.go — 路径定位/目录打开测试（批次 110）
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveOpenLocationDirectory(t *testing.T) {
	if got := resolveOpenLocationDirectory(""); got != "" {
		t.Fatalf("空串应返回空：%q", got)
	}

	dir := t.TempDir()
	// 目录 → 返回本身
	if got := resolveOpenLocationDirectory(dir); got != dir {
		t.Fatalf("目录应返回本身：%q 期望 %q", got, dir)
	}

	// 文件 → 返回父目录
	f := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := resolveOpenLocationDirectory(f); got != dir {
		t.Fatalf("文件应返回父目录：%q 期望 %q", got, dir)
	}

	// 不存在但含分隔符 → Dir
	nonexist := filepath.Join(dir, "nope", "x.txt")
	if got := resolveOpenLocationDirectory(nonexist); got != filepath.Dir(nonexist) {
		t.Fatalf("不存在路径应返回 Dir：%q 期望 %q", got, filepath.Dir(nonexist))
	}

	// 不存在且无分隔符 → ""
	if got := resolveOpenLocationDirectory("no_such_thing_12345"); got != "" {
		t.Fatalf("无分隔符不存在路径应返回空：%q", got)
	}
}

func TestOpenPathDirectoryEmpty(t *testing.T) {
	err := openPathDirectory("")
	if err == nil || err.Error() != "路径不能为空" {
		t.Fatalf("空路径应返回 newPathError：%v", err)
	}
}

func TestOpenPathCommandLineEmpty(t *testing.T) {
	err := openPathCommandLine("")
	if err == nil || err.Error() != "位置不能为空" {
		t.Fatalf("空路径应返回位置错误：%v", err)
	}
}

func TestOpenPathCommandLineAdminEmpty(t *testing.T) {
	err := openPathCommandLineAdmin("")
	if err == nil || err.Error() != "位置不能为空" {
		t.Fatalf("空路径应返回位置错误：%v", err)
	}
}
