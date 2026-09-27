package main

import (
	"os"
	"path/filepath"
	"testing"
)

// 批次 104：CF_HDROP 编排侧测试（格式注册 + 提交 smoke）。

func TestRegisterPreferredDropEffectClipboardFormat(t *testing.T) {
	format, err := registerPreferredDropEffectClipboardFormat()
	if err != nil {
		t.Fatal(err)
	}
	// RegisterClipboardFormatW 注册的格式 >= 0xC000，且非 0。
	if format == 0 || format < 0xC000 {
		t.Fatalf("注册格式异常：%d", format)
	}
}

func TestSetWindowsFileClipboard(t *testing.T) {
	// 用临时文件作为剪贴板文件源。
	dir := t.TempDir()
	p := filepath.Join(dir, "drop.txt")
	if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// 剪贴板可能被其他进程占用；占用时 openClipboardForFileOperation 重试后返回
	// "打开剪贴板失败"，此时跳过而非判失败。
	if err := setWindowsFileClipboard([]string{p}, dropEffectCopy); err != nil {
		if err.Error() == "打开剪贴板失败" {
			t.Skip("剪贴板被占用，跳过提交验证")
		}
		t.Fatalf("setWindowsFileClipboard: %v", err)
	}
}
