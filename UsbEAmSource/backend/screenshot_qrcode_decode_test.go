package main

import (
	"image"
	"strings"
	"testing"

	"github.com/makiuchi-d/gozxing"
)

// 批次 105：二维码解码链纯逻辑测试（key 构建 + 结果转换/去重）。

func TestBuildQRCodeDecodedEntryKey(t *testing.T) {
	if got := buildQRCodeDecodedEntryKey("hello", "QR_CODE", false, 0, 0, 0, 0, 0, 0); got != "QR_CODE:hello" {
		t.Fatalf("非 selectable key = %q 期望 %q", got, "QR_CODE:hello")
	}
	// selectable key 应含坐标值。
	got := buildQRCodeDecodedEntryKey("x", "QR_CODE", true, 1.5, 2.5, 3, 4, 5, 6)
	if got == "" || got[0:8] != "QR_CODE:" {
		t.Fatalf("selectable key 异常：%q", got)
	}
}

func newTestResult(text string) *gozxing.Result {
	p := gozxing.NewResultPoint(10, 20)
	pts := []gozxing.ResultPoint{p, p, p}
	return gozxing.NewResult(text, nil, pts, gozxing.BarcodeFormat_QR_CODE)
}

func TestConvertQRCodeResultsEmpty(t *testing.T) {
	if got := convertQRCodeResults(nil, image.Rect(0, 0, 100, 100)); got != nil {
		t.Fatalf("nil results 应返回 nil，得到 %v", got)
	}
	if got := convertQRCodeResults([]*gozxing.Result{}, image.Rect(0, 0, 100, 100)); got != nil {
		t.Fatalf("空 results 应返回 nil，得到 %v", got)
	}
}

func TestConvertQRCodeResults(t *testing.T) {
	entries := convertQRCodeResults(
		[]*gozxing.Result{newTestResult("  hello  "), newTestResult("   ")},
		image.Rect(0, 0, 100, 100),
	)
	if len(entries) != 1 {
		t.Fatalf("应只剩 1 个有效 entry，得到 %d", len(entries))
	}
	if entries[0].Text != "hello" {
		t.Fatalf("Text 未 TrimSpace：%q", entries[0].Text)
	}
	if entries[0].Format != "QR_CODE" {
		t.Fatalf("Format 映射错误：%q", entries[0].Format)
	}
}

func TestConvertQRCodeResultsDedup(t *testing.T) {
	entries := convertQRCodeResults(
		[]*gozxing.Result{newTestResult("dup"), newTestResult("dup")},
		image.Rect(0, 0, 100, 100),
	)
	if len(entries) != 1 {
		t.Fatalf("重复 entry 应去重，得到 %d", len(entries))
	}
}

func TestBuildQRCodeDecodeResultFromPNGEmpty(t *testing.T) {
	got, err := buildQRCodeDecodeResultFromPNG(nil)
	if err != nil {
		t.Fatalf("空 PNG 不应报错：%v", err)
	}
	if got.ImageData != "" || len(got.Entries) != 0 {
		t.Fatalf("空 PNG 应返回零值结果，得到 %+v", got)
	}
}

func TestBuildQRCodeDecodeResultFromPNGInvalid(t *testing.T) {
	_, err := buildQRCodeDecodeResultFromPNG([]byte("not a png"))
	if err == nil {
		t.Fatal("无效 PNG 应报错")
	}
	if !strings.Contains(err.Error(), "解析截图内容失败") {
		t.Fatalf("错误消息应含前缀，得到 %q", err.Error())
	}
}

func TestBuildQRCodeDecodeResultFromPNGRoundtrip(t *testing.T) {
	dataURL, err := generateQRCodeDataURL("hello-qr", 300)
	if err != nil {
		t.Fatal(err)
	}
	png, err := decodeDataURLPNG(dataURL)
	if err != nil {
		t.Fatal(err)
	}
	result, err := buildQRCodeDecodeResultFromPNG(png)
	if err != nil {
		t.Fatalf("解码失败：%v", err)
	}
	if !strings.HasPrefix(result.ImageData, "data:image/png;base64,") {
		t.Fatalf("ImageData 前缀异常：%q", result.ImageData[:40])
	}
	if len(result.Entries) == 0 {
		t.Fatal("应解码出至少一个二维码")
	}
	if result.Entries[0].Text != "hello-qr" {
		t.Fatalf("解码文本异常：%q", result.Entries[0].Text)
	}
	if result.ImageWidth <= 0 || result.ImageHeight <= 0 {
		t.Fatalf("图像尺寸异常：%dx%d", result.ImageWidth, result.ImageHeight)
	}
}
