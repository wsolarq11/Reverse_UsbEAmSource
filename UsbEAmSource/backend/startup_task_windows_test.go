package main

import (
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

func TestDecodeWindowsCommandOutputEmpty(t *testing.T) {
	if got := decodeWindowsCommandOutput(nil); got != "" {
		t.Fatalf("空输入应返回空串, got %q", got)
	}
	if got := decodeWindowsCommandOutput([]byte{}); got != "" {
		t.Fatalf("空切片应返回空串, got %q", got)
	}
}

func TestDecodeWindowsCommandOutputValidUTF8(t *testing.T) {
	got := decodeWindowsCommandOutput([]byte("  hello  "))
	if got != "hello" {
		t.Fatalf("UTF-8 有效应 TrimSpace 原样返回, got %q", got)
	}
}

// GBK 中文 "找不到" = D5 D2 B2 BB B5 BD。非 UTF-8，应回退 ACP/936 解码。
func TestDecodeWindowsCommandOutputGBK(t *testing.T) {
	gbk := []byte{0xd5, 0xd2, 0xb2, 0xbb, 0xb5, 0xbd}

	// 确定性：GBK(936) 解码路径不依赖系统代码页，任何环境都必须解出"找不到"。
	if s, ok := decodeWindowsBytes(gbk, 936); !ok || s != "找不到" {
		t.Fatalf("GBK(936) 解码失败, got %q ok=%v", s, ok)
	}

	// decodeWindowsCommandOutput 忠实还原原 asm 的 GetACP→936 顺序：
	// 中文系统(ACP=936)命中 GBK；非中文系统 ACP 优先，不强断言具体值。
	if windows.GetACP() == 936 {
		if got := decodeWindowsCommandOutput(gbk); got != "找不到" {
			t.Fatalf("ACP=936 时应解码为找不到, got %q", got)
		}
	}
}

func TestIsTaskNotFoundMessageCannotFind(t *testing.T) {
	if !isTaskNotFoundMessage("ERROR: The system cannot find the file specified.") {
		t.Fatal("含 cannot find 应判为任务不存在")
	}
	if !isTaskNotFoundMessage("  CANNOT FIND  ") {
		t.Fatal("大小写与空白应归一化后命中")
	}
}

func TestIsTaskNotFoundMessageNegative(t *testing.T) {
	for _, m := range []string{"", "成功: 已删除计划任务", "success"} {
		if isTaskNotFoundMessage(m) {
			t.Fatalf("不应判为任务不存在: %q", m)
		}
	}
}

func TestIsTaskNotFoundMessageLowercased(t *testing.T) {
	if !isTaskNotFoundMessage(strings.ToUpper("cannot find")) {
		t.Fatal("ToLower 后应命中 cannot find")
	}
}
