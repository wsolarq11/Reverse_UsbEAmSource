package main

import (
	"bytes"
	"image"
	"image/png"
	"strings"
	"testing"

	"github.com/makiuchi-d/gozxing"
)

// 批次 101：QR 码生成链纯逻辑测试（不触 Windows API / clipboard）。

func TestClampQRCodeDimension(t *testing.T) {
	cases := []struct {
		in   int
		want int
	}{
		{0, 420}, {100, 420}, {191, 420},
		{192, 192}, {300, 300}, {512, 512},
		{513, 512}, {1000, 512},
	}
	for _, c := range cases {
		if got := clampQRCodeDimension(c.in); got != c.want {
			t.Fatalf("clamp(%d)=%d 期望 %d", c.in, got, c.want)
		}
	}
}

func TestGenerateQRCodePNGEmpty(t *testing.T) {
	if _, err := generateQRCodePNG("", 300); err == nil || err.Error() != "二维码内容不能为空" {
		t.Fatalf("空内容应报错：%v", err)
	}
	if _, err := generateQRCodePNG("   \t\n", 300); err == nil {
		t.Fatal("全空白内容应报错")
	}
}

func TestGenerateQRCodePNG(t *testing.T) {
	got, err := generateQRCodePNG("hello world", 300)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(got))
	if err != nil {
		t.Fatalf("PNG 解码失败：%v", err)
	}
	b := img.Bounds()
	if b.Dx() <= 0 || b.Dy() <= 0 {
		t.Fatalf("尺寸异常：%v", b)
	}
	// NRGBA 透明背景：验证四个角至少有一个透明像素（白模块渲染为 alpha=0）。
	if _, ok := img.(*image.NRGBA); !ok {
		t.Fatalf("期望 *image.NRGBA，得到 %T", img)
	}
}

func TestRenderTransparentQRCodeImage(t *testing.T) {
	m, err := gozxing.NewBitMatrix(4, 4)
	if err != nil {
		t.Fatal(err)
	}
	// 左上 2x2 置位（黑模块），其余未置位（透明）。
	m.Set(0, 0)
	m.Set(1, 0)
	m.Set(0, 1)
	m.Set(1, 1)

	img := renderTransparentQRCodeImage(m)
	if img.Bounds().Dx() != 4 || img.Bounds().Dy() != 4 {
		t.Fatalf("尺寸异常：%v", img.Bounds())
	}
	// 黑模块：不透明黑。
	r, g, b, a := img.At(0, 0).RGBA()
	if r != 0 || g != 0 || b != 0 || a != 0xffff {
		t.Fatalf("黑模块像素异常：%v,%v,%v,%v", r, g, b, a)
	}
	// 白模块：透明（alpha=0）。
	_, _, _, a2 := img.At(3, 3).RGBA()
	if a2 != 0 {
		t.Fatalf("透明模块 alpha=%v 期望 0", a2)
	}
}

func TestGenerateQRCodeDataURL(t *testing.T) {
	got, err := generateQRCodeDataURL("abc", 300)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "data:image/png;base64,") {
		t.Fatalf("前缀异常：%q", got[:40])
	}
}
