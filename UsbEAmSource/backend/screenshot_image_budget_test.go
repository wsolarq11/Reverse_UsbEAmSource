package main

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 验证 Reserve 的预算占用/一次性释放语义。
func TestReserveScreenshotImageBudget(t *testing.T) {
	b := &screenshotImageMemoryBudget{limit: 1000}

	release, err := b.Reserve(400)
	if err != nil {
		t.Fatalf("Reserve(400) err = %v", err)
	}
	if b.used != 400 {
		t.Fatalf("used after Reserve = %d, want 400", b.used)
	}

	// 释放一次后 used 归零。
	release()
	if b.used != 0 {
		t.Fatalf("used after release = %d, want 0", b.used)
	}

	// sync.Once 保证二次释放不再扣减（不会变成负数）。
	release()
	if b.used != 0 {
		t.Fatalf("used after double release = %d, want 0", b.used)
	}

	// 超限：size > limit。
	if _, err := b.Reserve(1001); err == nil {
		t.Fatal("Reserve(1001) should fail, got nil")
	}

	// 超限：used + size > limit。
	if _, err := b.Reserve(700); err != nil {
		t.Fatalf("Reserve(700) err = %v", err)
	}
	if _, err := b.Reserve(301); err == nil {
		t.Fatal("Reserve(301) should fail when used=700 limit=1000")
	}

	// nil receiver。
	if _, err := (*screenshotImageMemoryBudget)(nil).Reserve(1); err == nil {
		t.Fatal("nil receiver Reserve should fail")
	}
}

// 验证配置校验的分支与估算公式。
func TestValidateScreenshotImageConfig(t *testing.T) {
	cfg := screenshotImageWorkConfig

	// 正常：宽高/步长/通道均在限内，estimated = stride + channels*w*h。
	est, err := validateScreenshotImageConfig(nil, 100, 100, 400, 4, cfg)
	if err != nil {
		t.Fatalf("validate ok case err = %v", err)
	}
	if want := int64(400 + 4*100*100); est != want {
		t.Fatalf("estimated = %d, want %d", est, want)
	}

	// stride <= 0。
	if _, err := validateScreenshotImageConfig(nil, 100, 100, 0, 4, cfg); err == nil {
		t.Fatal("stride<=0 should fail")
	}

	// stride > maxBytes。
	if _, err := validateScreenshotImageConfig(nil, 100, 100, cfg.maxBytes+1, 4, cfg); err == nil {
		t.Fatal("stride>maxBytes should fail")
	}

	// 宽高非法。
	if _, err := validateScreenshotImageConfig(nil, 0, 100, 400, 4, cfg); err == nil {
		t.Fatal("width<=0 should fail")
	}
	if _, err := validateScreenshotImageConfig(nil, cfg.maxWidth+1, 100, 400, 4, cfg); err == nil {
		t.Fatal("width>maxWidth should fail")
	}

	// 像素超限（10000*5000 = 5e7 > maxPixels = 4e7，宽高均在 32768 内）。
	if _, err := validateScreenshotImageConfig(nil, 10000, 5000, 400, 4, cfg); err == nil {
		t.Fatal("pixels>maxPixels should fail")
	}

	// 工作集超限（stride + channels*w*h > maxWorkSet）。
	if _, err := validateScreenshotImageConfig(nil, 100, 100, cfg.maxWorkSet, 4, cfg); err == nil {
		t.Fatal("estimated>maxWorkSet should fail")
	}
}

// 验证受限读取：正常读取 + 超限拒绝 + 目录拒绝。
func TestReadScreenshotImageFileLimited(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.png")

	data := []byte("not-a-real-png-but-small")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := readScreenshotImageFileLimited(path)
	if err != nil {
		t.Fatalf("read err = %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("read bytes = %q, want %q", got, data)
	}

	// 目录应报错（IsDir 分支）。
	if _, err := readScreenshotImageFileLimited(dir); err == nil {
		t.Fatal("directory should fail")
	}

	// 空文件（size<=0）应报错。
	empty := filepath.Join(dir, "empty.bin")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readScreenshotImageFileLimited(empty); err == nil {
		t.Fatal("empty file should fail")
	}
}

// 验证从 PNG 构建元数据（scrolling 跳过预算预留，普通模式走全局 Reserve 并立即释放）。
func TestBuildScreenshotResultMetadataFromPNG(t *testing.T) {
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 3, 5))
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}

	meta, err := buildScreenshotResultMetadataFromPNG(buf.Bytes(), "")
	if err != nil {
		t.Fatalf("build err = %v", err)
	}
	if meta.Width != 3 || meta.Height != 5 {
		t.Fatalf("meta = %dx%d, want 3x5", meta.Width, meta.Height)
	}
	if meta.Mode != normalizeScreenshotMode("") {
		t.Fatalf("mode = %q, want %q", meta.Mode, normalizeScreenshotMode(""))
	}

	// 空 PNG 返回零值 + nil error。
	meta, err = buildScreenshotResultMetadataFromPNG(nil, "")
	if err != nil {
		t.Fatalf("empty png err = %v, want nil", err)
	}
	if meta.Width != 0 || meta.Height != 0 || meta.Mode != "" {
		t.Fatalf("empty png meta = %+v, want zero", meta)
	}

	// 非 PNG 字节应报错（DecodeConfig 失败）。
	if _, err := buildScreenshotResultMetadataFromPNG([]byte("definitely-not-png"), ""); err == nil {
		t.Fatal("non-png should fail")
	} else if !strings.Contains(err.Error(), "解析截图内容失败") {
		t.Fatalf("error = %v, want 解析截图内容失败 prefix", err)
	}
}
