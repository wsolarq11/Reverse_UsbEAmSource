// AUTO-RECONSTRUCTED — QR 码生成与 data URL 链（empirical asm）
// 研究用途
package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"golang.design/x/clipboard"
)

// clampQRCodeDimension 将用户指定尺寸收敛到 QR 渲染尺寸区间。
// [S 汇编 0x140931a76 内联] size<0xc0(192)→0x1a4(420)，size>0x200(512)→0x200，否则原值。
func clampQRCodeDimension(size int) int {
	if size < 0xc0 {
		return 0x1a4
	}
	if size > 0x200 {
		return 0x200
	}
	return size
}

// generateQRCodePNG 将文本编码为透明背景 QR 码 PNG。
// [S 汇编 0x140931a40, 608B] strings.TrimSpace 仅用于空内容判定（原始 data 参与编码）→
// 尺寸 clamp → gozxing QRCodeWriter.Encode(BarcodeFormat_QR_CODE=0xb) →
// renderTransparentQRCodeImage → png.Encode。
func generateQRCodePNG(data string, size int) ([]byte, error) {
	if strings.TrimSpace(data) == "" {
		return nil, errors.New("二维码内容不能为空")
	}
	width := clampQRCodeDimension(size)
	height := clampQRCodeDimension(size)
	matrix, err := (&qrcode.QRCodeWriter{}).Encode(data, gozxing.BarcodeFormat_QR_CODE, width, height, nil)
	if err != nil {
		return nil, fmt.Errorf("生成二维码失败: %w", err)
	}
	img := renderTransparentQRCodeImage(matrix)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("编码二维码图片失败: %w", err)
	}
	return buf.Bytes(), nil
}

// renderTransparentQRCodeImage 将 BitMatrix 渲染为透明背景的 *image.NRGBA。
// [S 汇编 0x140931ca0, 416B] 位=1（黑模块）→ NRGBA{0,0,0,0xff}；位=0 → NRGBA{0xff,0xff,0xff,0}。
// 位读取展开自 gozxing.BitMatrix.Get：bits[rowSize*y + x/32]>>(x&31)&1。
func renderTransparentQRCodeImage(matrix *gozxing.BitMatrix) *image.NRGBA {
	w := matrix.GetWidth()
	h := matrix.GetHeight()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if matrix.Get(x, y) {
				img.SetNRGBA(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 0xff})
			} else {
				img.SetNRGBA(x, y, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0})
			}
		}
	}
	return img
}

// generateQRCodeDataURL 生成二维码 PNG 的 base64 data URL。
// [S 汇编 0x140931e40, 192B] generateQRCodePNG 失败透传；空 PNG 返回空串；
// 否则 "data:image/png;base64," + base64.StdEncoding。
func generateQRCodeDataURL(data string, size int) (string, error) {
	pngBytes, err := generateQRCodePNG(data, size)
	if err != nil {
		return "", err
	}
	if len(pngBytes) == 0 {
		return "", nil
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes), nil
}

// copyQRCodeImageToClipboard 生成二维码 PNG 并写入剪贴板图片格式。
// [S 汇编 0x140931f00, 224B] initQRCodeClipboard 失败 → "初始化剪贴板失败: %w"；
// generateQRCodePNG 失败透传；否则 clipboard.Write(FmtImage, png)（丢弃返回 channel）。
func copyQRCodeImageToClipboard(data string, size int) error {
	if err := initQRCodeClipboard(); err != nil {
		return fmt.Errorf("初始化剪贴板失败: %w", err)
	}
	pngBytes, err := generateQRCodePNG(data, size)
	if err != nil {
		return err
	}
	clipboard.Write(clipboard.FmtImage, pngBytes)
	return nil
}
