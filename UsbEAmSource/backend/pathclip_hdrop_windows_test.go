package main

import (
	"encoding/binary"
	"strings"
	"syscall"
	"testing"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/w32"
)

// 批次 103：CF_HDROP 数据构建函数测试（encodeHDropPaths 纯逻辑 + GlobalAlloc 结构写入）。

func utf16ToGo(u []uint16) string {
	if len(u) > 0 && u[len(u)-1] == 0 {
		u = u[:len(u)-1]
	}
	return syscall.UTF16ToString(u)
}

func TestEncodeHDropPaths(t *testing.T) {
	got, err := encodeHDropPaths([]string{"C:\\A", "  C:\\B  ", "  "})
	if err != nil {
		t.Fatal(err)
	}
	// 双 null 终止：最后一个元素后跟 null。
	if got[len(got)-1] != 0 {
		t.Fatalf("末尾应为 null 终止")
	}
	// 将 \0 替换为 '|' 以便同时检查多个路径（DROPFILES 用 \0 分隔）。
	var sb strings.Builder
	for _, c := range got {
		if c == 0 {
			sb.WriteByte('|')
		} else {
			sb.WriteRune(rune(c))
		}
	}
	s := sb.String()
	if !strings.Contains(s, "C:\\A") || !strings.Contains(s, "C:\\B") {
		t.Fatalf("路径缺失：%q", s)
	}
}

func TestEncodeHDropPathsEmpty(t *testing.T) {
	if _, err := encodeHDropPaths(nil); err == nil || err.Error() != "文件路径不能为空" {
		t.Fatalf("nil 应报错「文件路径不能为空」：%v", err)
	}
	if _, err := encodeHDropPaths([]string{"  ", "\t"}); err == nil {
		t.Fatal("全空白路径应报错")
	}
}

func TestBuildHDropClipboardData(t *testing.T) {
	hMem, err := buildHDropClipboardData([]string{"C:\\file.txt"})
	if err != nil {
		t.Fatal(err)
	}
	if hMem == 0 {
		t.Fatal("hMem 不应为 0")
	}
	defer w32.GlobalFree(hMem)

	p := w32.GlobalLock(hMem)
	if p == nil {
		t.Fatal("GlobalLock 失败")
	}
	defer w32.GlobalUnlock(hMem)

	header := unsafe.Slice((*byte)(p), dropFilesHeaderSize)
	if got := binary.LittleEndian.Uint32(header[0:4]); got != dropFilesHeaderSize {
		t.Fatalf("pFiles=%d 期望 %d", got, dropFilesHeaderSize)
	}
	if got := binary.LittleEndian.Uint32(header[16:20]); got != 1 {
		t.Fatalf("fWide=%d 期望 1", got)
	}
	// 路径 UTF-16 从偏移 20 开始。
	tail := unsafe.Slice((*uint16)(unsafe.Add(p, dropFilesHeaderSize)), 20)
	if s := utf16ToGo(tail); !strings.Contains(s, "C:\\file.txt") {
		t.Fatalf("路径数据缺失：%q", s)
	}
}

func TestBuildDropEffectClipboardData(t *testing.T) {
	hMem, err := buildDropEffectClipboardData(dropEffectCopy)
	if err != nil {
		t.Fatal(err)
	}
	defer w32.GlobalFree(hMem)

	p := w32.GlobalLock(hMem)
	if p == nil {
		t.Fatal("GlobalLock 失败")
	}
	defer w32.GlobalUnlock(hMem)

	if got := *(*uint32)(p); got != dropEffectCopy {
		t.Fatalf("dropEffect=%d 期望 %d", got, dropEffectCopy)
	}
}
