package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestLauncherStartupTaskLogonDelay(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{0, "PT10S"},
		{-5, "PT10S"},
		{5, "PT5S"},
		{50, "PT50S"},
		{100, "PT100S"},
		{150, "PT100S"},
	}
	for _, c := range cases {
		if got := launcherStartupTaskLogonDelay(c.in); got != c.want {
			t.Errorf("delaySeconds=%d: got %q, want %q", c.in, got, c.want)
		}
	}
}

func TestLauncherStartupTaskWorkingDirectory(t *testing.T) {
	if got := launcherStartupTaskWorkingDirectory(`C:\tools\app.exe`); got != `C:\tools` {
		t.Fatalf("绝对路径应取父目录, got %q", got)
	}
	// 相对路径（无目录）回退程序目录。
	want := filepath.Dir(os.Args[0])
	if got := launcherStartupTaskWorkingDirectory("app.exe"); got != want {
		t.Fatalf("相对路径应回退程序目录, got %q want %q", got, want)
	}
}

func TestQuoteWindowsTaskActionCommand(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"   ", ""},
		{`"C:\a.exe"`, `"C:\a.exe"`},
		{`C:\a.exe`, `"C:\a.exe"`},
		{`C:\a"b.exe`, `"C:\a\"b.exe"`},
	}
	for _, c := range cases {
		if got := quoteWindowsTaskActionCommand(c.in); got != c.want {
			t.Errorf("input %q: got %q, want %q", c.in, got, c.want)
		}
	}
}

func TestEncodeUTF16LEWithBOM(t *testing.T) {
	got := encodeUTF16LEWithBOM("A")
	want := []byte{0xff, 0xfe, 0x41, 0x00}
	if !bytes.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestEncodeUTF16LEWithBOMEmpty(t *testing.T) {
	got := encodeUTF16LEWithBOM("")
	if !bytes.Equal(got, []byte{0xff, 0xfe}) {
		t.Fatalf("空串应仅 BOM, got %v", got)
	}
}

func TestCurrentLauncherStartupTaskUserID(t *testing.T) {
	id, err := currentLauncherStartupTaskUserID()
	if err == nil && id == "" {
		t.Fatal("成功时不应返回空标识")
	}
}

func TestBuildLauncherStartupTaskDefinitionXMLEmptyExe(t *testing.T) {
	_, err := buildLauncherStartupTaskDefinitionXML("  ", "", "user", true, 5)
	if err == nil || err.Error() != "开机启动程序路径不能为空" {
		t.Fatalf("空 exe 应返回入口路径校验错误, got %v", err)
	}
}

func TestBuildLauncherStartupTaskDefinitionXMLEmptyUserID(t *testing.T) {
	_, err := buildLauncherStartupTaskDefinitionXML(`C:\tools\app.exe`, "", " ", true, 5)
	if err == nil || err.Error() != "开机启动任务用户标识不能为空" {
		t.Fatalf("空 userID 应返回用户标识校验错误, got %v", err)
	}
}

func TestBuildLauncherStartupTaskDefinitionXML(t *testing.T) {
	b, err := buildLauncherStartupTaskDefinitionXML(`C:\tools\app.exe`, "-arg", "S-1-5-18", true, 5)
	if err != nil {
		t.Fatalf("构建失败: %v", err)
	}
	if !bytes.HasPrefix(b, []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")) {
		t.Fatalf("缺少 XML 声明头: %q", b[:min(len(b), 60)])
	}
	if !bytes.Contains(b, []byte("<Task")) {
		t.Fatalf("缺少 <Task> 根元素")
	}
	if !bytes.Contains(b, []byte("PT5S")) {
		t.Fatalf("缺少 logon delay PT5S")
	}
	if !bytes.Contains(b, []byte("S4U")) {
		t.Fatalf("enabled 时 LogonType 应为 S4U")
	}
	if !bytes.Contains(b, []byte("app.exe")) {
		t.Fatalf("Command 应包含 app.exe")
	}
}

func TestBuildLauncherStartupTaskDefinitionXMLDisabled(t *testing.T) {
	b, err := buildLauncherStartupTaskDefinitionXML(`C:\tools\app.exe`, "", "u", false, 0)
	if err != nil {
		t.Fatalf("构建失败: %v", err)
	}
	if !bytes.Contains(b, []byte("InteractiveToken")) {
		t.Fatalf("禁用时 LogonType 应为 InteractiveToken")
	}
}

// TestBuildLauncherStartupTaskDefinitionXMLTemplateFields 锁定批次 130 模板取证：
// Version="1.2"（非 1.0）、Settings.Enabled 恒 true、Hidden=enabled。
func TestBuildLauncherStartupTaskDefinitionXMLTemplateFields(t *testing.T) {
	enabled, err := buildLauncherStartupTaskDefinitionXML(`C:\tools\app.exe`, "", "u", true, 0)
	if err != nil {
		t.Fatalf("构建失败: %v", err)
	}
	if !bytes.Contains(enabled, []byte(`version="1.2"`)) {
		t.Fatalf("Version 应为 1.2, got %q", enabled[:min(len(enabled), 160)])
	}
	if !bytes.Contains(enabled, []byte("<Enabled>true</Enabled>")) {
		t.Fatalf("enabled 时 Settings.Enabled 应为 true")
	}
	if !bytes.Contains(enabled, []byte("<Hidden>true</Hidden>")) {
		t.Fatalf("enabled 时 Settings.Hidden 应为 true")
	}

	disabled, err := buildLauncherStartupTaskDefinitionXML(`C:\tools\app.exe`, "", "u", false, 0)
	if err != nil {
		t.Fatalf("构建失败: %v", err)
	}
	if !bytes.Contains(disabled, []byte("<Enabled>true</Enabled>")) {
		t.Fatalf("禁用时 Settings.Enabled 仍应为 true")
	}
	if !bytes.Contains(disabled, []byte("<Hidden>false</Hidden>")) {
		t.Fatalf("禁用时 Settings.Hidden 应为 false")
	}
}

func TestConvertXMLFileToUTF16LE(t *testing.T) {
	src := filepath.Join(t.TempDir(), "task.xml")
	if err := os.WriteFile(src, []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<Task/>"), 0o600); err != nil {
		t.Fatal(err)
	}
	tmpPath, cleanup, err := convertXMLFileToUTF16LE(src)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
	defer cleanup()
	if filepath.Base(tmpPath) != "task.xml.utf16.xml" {
		t.Fatalf("临时路径后缀错误: %q", tmpPath)
	}
	data, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("读取临时文件失败: %v", err)
	}
	if len(data) < 2 || data[0] != 0xff || data[1] != 0xfe {
		t.Fatalf("缺少 UTF-16LE BOM: %v", data[:min(len(data), 4)])
	}
	// 声明应已被替换为 UTF-16（UTF-16LE 编码）。
	utf16le := []byte{'U', 0, 'T', 0, 'F', 0, '-', 0, '1', 0, '6', 0}
	if !bytes.Contains(data, utf16le) {
		t.Fatalf("声明未替换为 UTF-16")
	}
	cleanup()
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Fatalf("cleanup 应删除临时文件")
	}
}
