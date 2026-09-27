package main

import (
	"reflect"
	"testing"
)

// 批次 102：文件剪贴板链纯逻辑测试（normalizePathList / resolveClipboardDropEffect / setFileClipboard）。

func TestNormalizePathListEmpty(t *testing.T) {
	if got := normalizePathList(nil); got != nil {
		t.Fatalf("nil 输入应返回 nil，得到 %v", got)
	}
	if got := normalizePathList([]string{}); got != nil {
		t.Fatalf("空切片应返回 nil，得到 %v", got)
	}
}

func TestNormalizePathList(t *testing.T) {
	got := normalizePathList([]string{
		"  C:\\Users\\A\\..\\B  ",
		"c:\\users\\b", // 与上一条 ToLower 后重复
		"   ",          // 空（TrimSpace 后），跳过
		"C:\\temp\\file.txt",
	})
	want := []string{
		"C:\\Users\\B", // TrimSpace + Clean(..) → C:\Users\B
		"C:\\temp\\file.txt",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizePathList = %#v 期望 %#v", got, want)
	}
}

func TestResolveClipboardDropEffect(t *testing.T) {
	cases := []struct {
		in     string
		want   uint32
		wantOK bool
	}{
		{"copy", dropEffectCopy, true},
		{"COPY", dropEffectCopy, true},
		{"  copy  ", dropEffectCopy, true},
		{"cut", dropEffectMove, true},
		{"move", dropEffectMove, true},
		{"Move", dropEffectMove, true},
		{"link", 0, false},
		{"", 0, false},
		{"unknown", 0, false},
	}
	for _, c := range cases {
		got, err := resolveClipboardDropEffect(c.in)
		if c.wantOK {
			if err != nil || got != c.want {
				t.Fatalf("resolve(%q)=%d,%v 期望 %d,nil", c.in, got, err, c.want)
			}
		} else if err == nil || err.Error() != "不支持的剪贴板文件操作" {
			t.Fatalf("resolve(%q) 应报错「不支持的剪贴板文件操作」，得到 %v", c.in, err)
		}
	}
}

func TestSetFileClipboardEmptyPaths(t *testing.T) {
	err := setFileClipboard(nil, "copy")
	if err == nil || err.Error() != "文件路径不能为空" {
		t.Fatalf("空路径应报错「文件路径不能为空」，得到 %v", err)
	}
	// 全空路径（TrimSpace 后全空）也应报错。
	err = setFileClipboard([]string{"  ", "\t"}, "copy")
	if err == nil || err.Error() != "文件路径不能为空" {
		t.Fatalf("全空路径应报错「文件路径不能为空」，得到 %v", err)
	}
}

func TestSetFileClipboardBadEffect(t *testing.T) {
	err := setFileClipboard([]string{"C:\\file.txt"}, "link")
	if err == nil || err.Error() != "不支持的剪贴板文件操作" {
		t.Fatalf("非法 dropEffect 应报错「不支持的剪贴板文件操作」，得到 %v", err)
	}
}

func TestSetFileClipboardOK(t *testing.T) {
	// setWindowsFileClipboard 为 stub（返回 nil），此路径验证编排正确。
	if err := setFileClipboard([]string{"C:\\file.txt"}, "copy"); err != nil {
		t.Fatalf("合法路径+effect 应成功：%v", err)
	}
}
