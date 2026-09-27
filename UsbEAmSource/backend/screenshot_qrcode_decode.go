// AUTO-RECONSTRUCTED — 二维码解码链（gozxing 多/单码解码 + 结果转换）
// 研究用途
package main

import (
	"encoding/base64"
	"fmt"
	"image"
	"math"
	"strings"

	"github.com/makiuchi-d/gozxing"
	multiqrcode "github.com/makiuchi-d/gozxing/multi/qrcode"
	"github.com/makiuchi-d/gozxing/qrcode"
)

// buildQRCodeDecodeResultFromPNG 从 PNG 字节构建二维码解码结果（QRCodeDecodeResult）。
// [S 汇编 0x140931fe0, 1056B] 空 png → 零值；decodeScreenshotImageBytesWithBudget(png,"png",12)
// 失败 → fmt.Errorf("解析截图内容失败: %w")；defer release；decodeQRCodesFromImage 得 entries；
// ImageData = "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)（concatstring2 @0x140932180）；
// ImageWidth/Height = img.Bounds().Dx()/Dy()（sub rcx,rdx / sub rdi,rbx @0x1409321d8）；
// len(entries)==1 时 SelectedEntryKey = buildQRCodeDecodedEntryKey(...)（cmp rdx,1 @0x1409321e6）。
func buildQRCodeDecodeResultFromPNG(png []byte) (QRCodeDecodeResult, error) {
	if len(png) == 0 {
		return QRCodeDecodeResult{}, nil
	}
	img, _, release, err := decodeScreenshotImageBytesWithBudget(png, "png", 12)
	if err != nil {
		return QRCodeDecodeResult{}, fmt.Errorf("解析截图内容失败: %w", err)
	}
	defer release()
	entries := decodeQRCodesFromImage(img)
	bounds := img.Bounds()
	dataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
	var key string
	if len(entries) == 1 {
		e := entries[0]
		key = buildQRCodeDecodedEntryKey(e.Text, e.Format, e.Selectable, e.MarkerX, e.MarkerY, e.BoundsLeft, e.BoundsTop, e.BoundsRight, e.BoundsBottom)
	}
	return QRCodeDecodeResult{
		ImageData:        dataURL,
		ImageWidth:       bounds.Dx(),
		ImageHeight:      bounds.Dy(),
		Entries:          entries,
		SelectedEntryKey: key,
	}, nil
}

// 包级 reader 单例（对应目标二进制全局 0x140c140e0 经指针链取 receiver，
// DecodeMultiple/Decode 依赖内部 decoder，必须经 New* 初始化）。
var (
	qrMultiReaderSingleton = multiqrcode.NewQRCodeMultiReader().(*multiqrcode.QRCodeMultiReader)
	qrReaderSingleton      = qrcode.NewQRCodeReader().(*qrcode.QRCodeReader)
)

// tryDecodeMultipleQRCodes 尝试多二维码解码（QRCodeMultiReader.DecodeMultiple）。
// [S 汇编 0x140932640, 160B] 构建 hints 参数栈帧后调 DecodeMultiple；
// err 非空 → 返回 (nil, err)；results 空 → 返回 (nil, nil)；否则透传。
func tryDecodeMultipleQRCodes(bitmap *gozxing.BinaryBitmap, hints map[gozxing.DecodeHintType]interface{}) ([]*gozxing.Result, error) {
	results, err := qrMultiReaderSingleton.DecodeMultiple(bitmap, hints)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results, nil
}

// decodeQRCodesFromImage 从图像解码二维码（先多码，空则单码回退）。
// [S 汇编 0x140932400, 576B] NewBinaryBitmapFromImage → hints{TRY_HARDER:true} →
// tryDecodeMultipleQRCodes → img.Bounds() → convertQRCodeResults；entries 空则
// QRCodeReader.Decode 单码回退 → 单结果 convertQRCodeResults。
func decodeQRCodesFromImage(img image.Image) []QRCodeDecodedEntry {
	bitmap, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil
	}
	hints := map[gozxing.DecodeHintType]interface{}{
		gozxing.DecodeHintType_TRY_HARDER: true,
	}
	results, _ := tryDecodeMultipleQRCodes(bitmap, hints)
	bounds := img.Bounds()
	if entries := convertQRCodeResults(results, bounds); len(entries) > 0 {
		return entries
	}
	result, err := qrReaderSingleton.Decode(bitmap, hints)
	if err != nil || result == nil {
		return nil
	}
	return convertQRCodeResults([]*gozxing.Result{result}, bounds)
}

// convertQRCodeResults 将 gozxing 结果转换为 QRCodeDecodedEntry 列表（含格式名与坐标）。
// [S 汇编 0x1409326e0, 1728B] nil results → nil；makeslice(cap=len(results))；
// 逐 result：nil 跳过 → TrimSpace(GetText) 空跳过 → buildQRCodeDecodedEntryFromPoints
// 算坐标 → 填充 Text/Format（format 用 GetBarcodeFormat().String() 映射）→
// buildQRCodeDecodedEntryKey → map[string]struct{} 去重 → append entry。
func convertQRCodeResults(results []*gozxing.Result, bounds image.Rectangle) []QRCodeDecodedEntry {
	if len(results) == 0 {
		return nil
	}
	out := make([]QRCodeDecodedEntry, 0, len(results))
	seen := make(map[string]struct{})
	for _, r := range results {
		if r == nil {
			continue
		}
		text := strings.TrimSpace(r.GetText())
		if text == "" {
			continue
		}
		entry := buildQRCodeDecodedEntryFromPoints(r.GetResultPoints(), bounds)
		entry.Text = text
		entry.Format = r.GetBarcodeFormat().String()
		key := buildQRCodeDecodedEntryKey(
			entry.Text, entry.Format, entry.Selectable,
			entry.MarkerX, entry.MarkerY,
			entry.BoundsLeft, entry.BoundsTop, entry.BoundsRight, entry.BoundsBottom,
		)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, entry)
	}
	return out
}

// buildQRCodeDecodedEntryKey 构建去重 key（":" 分隔 format/text，可选坐标）。
// [S-sig 汇编 0x140933720] !selectable 分支实证 strings.Join([]string{format,text},":")
// （0x140933b99 call strings.Join，sep=":" 由 read_gostring 0x140c6c000:1 实证）；
// selectable 分支 fmt.Sprintf 坐标格式串为推断（功能等价去重语义，格式串待取证）。
func buildQRCodeDecodedEntryKey(text, format string, selectable bool, markerX, markerY, boundsLeft, boundsTop, boundsRight, boundsBottom float64) string {
	if !selectable {
		return strings.Join([]string{format, text}, ":")
	}
	return fmt.Sprintf("%s:%s:%v:%v:%v:%v:%v:%v",
		format, text, markerX, markerY, boundsLeft, boundsTop, boundsRight, boundsBottom)
}

// buildQRCodeDecodedEntryFromPoints 从二维码定位点计算条目坐标（Marker + Bounds）。
// [S 汇编 0x140932da0, 2432B] points<3 或 bounds 宽/高<=0 → 空 entry（Selectable 恒 false，
// [rsp+0x100] 仅 duffzero 清零后 3 处 movzx 读取，无任何写入）。
// 坐标公式实证：marker = p0 + (p2 - p1)（subsd/addsd 链）；包围盒 = min/max(三点+marker)
// （minsd 链 + pxor 符号翻转=math.Max 的 NaN 语义）；归一化 = (v - Min) / Max
// （subsd [rsp+0x60]=Min.X → divsd [rsp+0x40]=Max.X；图像 bounds 通常 Min=(0,0)）。
func buildQRCodeDecodedEntryFromPoints(points []gozxing.ResultPoint, bounds image.Rectangle) QRCodeDecodedEntry {
	if len(points) < 3 || bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		return QRCodeDecodedEntry{}
	}
	p0, p1, p2 := points[0], points[1], points[2]
	markerX := p0.GetX() + (p2.GetX() - p1.GetX())
	markerY := p0.GetY() + (p2.GetY() - p1.GetY())
	minX := math.Min(math.Min(p0.GetX(), p1.GetX()), math.Min(p2.GetX(), markerX))
	maxX := math.Max(math.Max(p0.GetX(), p1.GetX()), math.Max(p2.GetX(), markerX))
	minY := math.Min(math.Min(p0.GetY(), p1.GetY()), math.Min(p2.GetY(), markerY))
	maxY := math.Max(math.Max(p0.GetY(), p1.GetY()), math.Max(p2.GetY(), markerY))
	normX := func(v float64) float64 { return (v - float64(bounds.Min.X)) / float64(bounds.Max.X) }
	normY := func(v float64) float64 { return (v - float64(bounds.Min.Y)) / float64(bounds.Max.Y) }
	return QRCodeDecodedEntry{
		MarkerX:      normX(markerX),
		MarkerY:      normY(markerY),
		BoundsLeft:   normX(minX),
		BoundsTop:    normY(minY),
		BoundsRight:  normX(maxX),
		BoundsBottom: normY(maxY),
		Selectable:   false,
	}
}
