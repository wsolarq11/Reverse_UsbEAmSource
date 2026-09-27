// AUTO-RECONSTRUCTED — screenshot_png_fast.go 蓝图（14 函数，批次 42）
// 自定义 PNG 快速编码器：unpremultiply + filter 选择 + IDAT 流式 chunk 写出。
// 全部 asm 直译：docs/goresym/pipeline/tmp/{encodeScreenshotPNGRowSource,
// writeScreenshotPNGIHDR,writeScreenshotPNGFilteredRows,selectScreenshotPNGFilter,
// filterScreenshotPNGPaeth,predictScreenshotPNGPaeth,convertScreenshotRGBARowToPNG,
// writeScreenshotPNGChunk,screenshotPNGIDATChunkWriter.*,screenshotRGBAImageRowSource.*,
// encodeScreenshotRGBAWithKlauspostPNG}.asm.txt
// 研究用途
package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"image"
	"io"

	"github.com/klauspost/compress/zlib"
)

// screenshotPNGFilterBuffers 5 个 filter 的临时缓冲。
type screenshotPNGFilterBuffers struct {
	filters [5][]byte
}

// [S 汇编 0x1409946e0] Bounds 返回行源边界；nil 接收者返零矩形。
func (s *screenshotRGBAImageRowSource) Bounds() image.Rectangle {
	if s == nil {
		return image.Rectangle{}
	}
	return s.Rect
}

// [S 汇编 0x140994720] Row 返回第 y 行像素（宽度*4 字节）。
func (s *screenshotRGBAImageRowSource) Row(y int) ([]byte, error) {
	if s == nil || s.Rect.Min.X >= s.Rect.Max.X || s.Rect.Min.Y >= s.Rect.Max.Y {
		return nil, errors.New("PNG 编码图像为空")
	}
	if y < s.Rect.Min.Y || y >= s.Rect.Max.Y {
		return nil, fmt.Errorf("PNG 行 %d 超出图像范围", y)
	}
	relY := y - s.Rect.Min.Y
	width := s.Rect.Max.X - s.Rect.Min.X
	start := relY * s.Stride
	end := start + width*4
	return s.Pix[start:end], nil
}

// [S 汇编 0x140995fc0] predictScreenshotPNGPaeth Paeth 预测器（asm 直译，含绝对值分支）。
func predictScreenshotPNGPaeth(a, b, c int) int {
	p := a + b - c
	pa := absScreenshotPNGInt(p - a)
	pb := absScreenshotPNGInt(p - b)
	pc := absScreenshotPNGInt(p - c)
	if pa <= pb && pa <= pc {
		return a
	}
	if pb <= pc {
		return b
	}
	return c
}

// [S 汇编 0x140995e20] filterScreenshotPNGPaeth 对一行应用 Paeth filter，返回代价。
func filterScreenshotPNGPaeth(dst, src, prev []byte) int {
	cost := 0
	for i := 0; i < len(src); i++ {
		var left, upLeft byte
		if i >= 4 {
			left = src[i-4]
			upLeft = prev[i-4]
		}
		up := prev[i]
		p := predictScreenshotPNGPaeth(int(left), int(up), int(upLeft))
		v := src[i] - byte(p)
		dst[i] = v
		cost += absScreenshotPNGByte(v)
	}
	return cost
}

// [S 汇编 0x140995620] convertScreenshotRGBARowToPNG 将 premultiplied RGBA 转 straight RGBA。
func convertScreenshotRGBARowToPNG(dst, src []byte) {
	for len(src) >= 4 && len(dst) >= 4 {
		a := src[3]
		switch {
		case a == 0:
			dst[0], dst[1], dst[2], dst[3] = 0, 0, 0, 0
		case a == 0xff:
			if &dst[0] != &src[0] {
				copy(dst[:4], src[:4])
			}
		default:
			dst[0] = unpremulScreenshotPNG(src[0], a)
			dst[1] = unpremulScreenshotPNG(src[1], a)
			dst[2] = unpremulScreenshotPNG(src[2], a)
			dst[3] = a
		}
		src = src[4:]
		dst = dst[4:]
	}
}

// [S-inline 汇编 0x1409956a2] unpremulScreenshotPNG 将 premultiplied 分量反转为 straight（convert 内联）。
func unpremulScreenshotPNG(c, a byte) byte {
	if a == 0 {
		return 0
	}
	if a == 0xff {
		return c
	}
	v := (uint32(c)*0xff00 + uint32(a)*257 - 1) / (uint32(a) * 257)
	if v > 0xff {
		return 0xff
	}
	return byte(v)
}

// [S 汇编 0x1409957c0] selectScreenshotPNGFilter 选择代价最小的 PNG filter。
func selectScreenshotPNGFilter(src, prev []byte, b *screenshotPNGFilterBuffers, count int) (int, []byte) {
	n := len(src)
	if n > len(b.filters[0]) {
		n = len(b.filters[0])
	}
	copy(b.filters[0][:n], src[:n])
	best := 0
	bestCost := sumAbsScreenshotPNG(src[:n])
	if bestCost == 0 {
		return 0, b.filters[0][:n]
	}
	if count > 1 {
		cost := 0
		for i := 0; i < n; i++ {
			var left byte
			if i >= 4 {
				left = src[i-4]
			}
			v := src[i] - left
			b.filters[1][i] = v
			cost += absScreenshotPNGByte(v)
		}
		if cost < bestCost {
			best, bestCost = 1, cost
		}
	}
	if count > 2 {
		cost := 0
		for i := 0; i < n; i++ {
			v := src[i] - prev[i]
			b.filters[2][i] = v
			cost += absScreenshotPNGByte(v)
		}
		if cost < bestCost {
			best, bestCost = 2, cost
		}
	}
	if count > 3 {
		cost := 0
		for i := 0; i < n; i++ {
			var left byte
			if i >= 4 {
				left = src[i-4]
			}
			avg := (int(left) + int(prev[i])) >> 1
			v := src[i] - byte(avg)
			b.filters[3][i] = v
			cost += absScreenshotPNGByte(v)
		}
		if cost < bestCost {
			best, bestCost = 3, cost
		}
	}
	if count > 4 {
		paethCost := filterScreenshotPNGPaeth(b.filters[4][:n], src[:n], prev[:n])
		if paethCost < bestCost {
			best = 4
		}
	}
	return best, b.filters[best][:n]
}

// [S 汇编 0x140996020] writeScreenshotPNGChunk 写一个 PNG chunk（长度+类型+数据+CRC32）。
func writeScreenshotPNGChunk(w io.Writer, data []byte, chunkType [4]byte) error {
	if len(data) > 0xffffffff {
		return fmt.Errorf("PNG 数据块过大: %d", len(data))
	}
	var hdr [8]byte
	binary.BigEndian.PutUint32(hdr[0:4], uint32(len(data)))
	copy(hdr[4:8], chunkType[:])
	if _, err := w.Write(hdr[:]); err != nil {
		return err
	}
	if len(data) > 0 {
		if _, err := w.Write(data); err != nil {
			return err
		}
	}
	crc := crc32.NewIEEE()
	crc.Write(chunkType[:])
	crc.Write(data)
	var sum [4]byte
	binary.BigEndian.PutUint32(sum[:], crc.Sum32())
	if _, err := w.Write(sum[:]); err != nil {
		return err
	}
	return nil
}

// [S 汇编 0x140994ba0] writeScreenshotPNGIHDR 写 IHDR chunk。
func writeScreenshotPNGIHDR(w io.Writer, x0, y0, x1, y1 int) error {
	width := x1 - x0
	height := y1 - y0
	if width <= 0 || height <= 0 {
		return errors.New("PNG 图像尺寸无效")
	}
	if uint64(width) > 0xffffffff || uint64(height) > 0xffffffff {
		return errors.New("PNG 图像尺寸超出限制")
	}
	var hdr [13]byte
	binary.BigEndian.PutUint32(hdr[0:4], uint32(width))
	binary.BigEndian.PutUint32(hdr[4:8], uint32(height))
	hdr[8] = 8 // bit depth
	hdr[9] = 6 // color type RGBA
	return writeScreenshotPNGChunk(w, hdr[:], [4]byte{'I', 'H', 'D', 'R'})
}

// [S 汇编 0x140995300] Write 缓冲写入 IDAT 数据，满时自动 flush。
func (w *screenshotPNGIDATChunkWriter) Write(p []byte) (int, error) {
	if w == nil || w.writer == nil || w.closed {
		return 0, errors.New("PNG IDAT 写入器不可用")
	}
	total := len(p)
	for len(p) > 0 {
		if w.limit-len(w.buffer) <= 0 {
			if err := w.flush(); err != nil {
				return 0, err
			}
		}
		n := w.limit - len(w.buffer)
		if n > len(p) {
			n = len(p)
		}
		w.buffer = append(w.buffer, p[:n]...)
		p = p[n:]
	}
	return total, nil
}

// [S 汇编 0x140995540] Close 标记关闭并 flush 剩余数据。
func (w *screenshotPNGIDATChunkWriter) Close() error {
	if w == nil || w.closed {
		return nil
	}
	w.closed = true
	return w.flush()
}

// [S 汇编 0x1409955a0] flush 把缓冲写出为 IDAT chunk。
func (w *screenshotPNGIDATChunkWriter) flush() error {
	if w == nil || w.writer == nil {
		return nil
	}
	if len(w.buffer) > 0 {
		if err := writeScreenshotPNGChunk(w.writer, w.buffer, [4]byte{'I', 'D', 'A', 'T'}); err != nil {
			return err
		}
		w.buffer = w.buffer[:0]
	}
	return nil
}

// [S 汇编 0x140994d00] writeScreenshotPNGFilteredRows 逐行过滤并写入 zlib 流。
func writeScreenshotPNGFilteredRows(zw io.Writer, src screenshotPNGRowSource) error {
	if zw == nil || src == nil {
		return errors.New("PNG 行编码器未初始化")
	}
	b := src.Bounds()
	width := b.Max.X - b.Min.X
	height := b.Max.Y - b.Min.Y
	if width <= 0 || height <= 0 {
		return errors.New("PNG 图像尺寸无效")
	}
	rowBytes := width * 4
	var fb screenshotPNGFilterBuffers
	for i := range fb.filters {
		fb.filters[i] = make([]byte, rowBytes)
	}
	prev := make([]byte, rowBytes)
	cur := make([]byte, rowBytes)
	out := make([]byte, rowBytes+1)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		row, err := src.Row(y)
		if err != nil {
			return err
		}
		if len(row) != rowBytes {
			return fmt.Errorf("PNG 第 %d 行长度无效：实际 %d，期望 %d", y, len(row), rowBytes)
		}
		hasPartial := false
		for i := 3; i < len(row); i += 4 {
			if row[i] != 0xff {
				hasPartial = true
				break
			}
		}
		var current []byte
		if hasPartial {
			convertScreenshotRGBARowToPNG(cur, row)
			current = cur
		} else {
			current = row
		}
		ft, filtered := selectScreenshotPNGFilter(current, prev, &fb, 5)
		out[0] = byte(ft)
		copy(out[1:], filtered)
		if _, err := zw.Write(out); err != nil {
			return err
		}
		copy(prev, current)
	}
	return nil
}

// [S 汇编 0x140994880] encodeScreenshotPNGRowSource 完整 PNG 编码。
func encodeScreenshotPNGRowSource(src screenshotPNGRowSource) ([]byte, error) {
	if src == nil {
		return nil, errors.New("PNG 编码图像为空")
	}
	b := src.Bounds()
	if b.Min.X >= b.Max.X || b.Max.Y <= b.Min.Y {
		return nil, errors.New("PNG 编码图像为空")
	}
	var buf bytes.Buffer
	buf.Write(screenshotPNGSignature)
	if err := writeScreenshotPNGIHDR(&buf, b.Min.X, b.Min.Y, b.Max.X, b.Max.Y); err != nil {
		return nil, err
	}
	idat := &screenshotPNGIDATChunkWriter{
		writer: &buf,
		buffer: make([]byte, 0, 1<<20),
		limit:  1 << 20,
	}
	zw, err := zlib.NewWriterLevelDict(idat, 1, nil)
	if err != nil {
		return nil, err
	}
	if err := writeScreenshotPNGFilteredRows(zw, src); err != nil {
		zw.Close()
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	if err := idat.Close(); err != nil {
		return nil, err
	}
	if err := writeScreenshotPNGChunk(&buf, nil, [4]byte{'I', 'E', 'N', 'D'}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// [S 汇编 0x140994640] encodeScreenshotRGBAWithKlauspostPNG 入口。
func encodeScreenshotRGBAWithKlauspostPNG(src *screenshotRGBAImageRowSource) ([]byte, error) {
	if src == nil || src.Rect.Min.X >= src.Rect.Max.X || src.Rect.Max.Y <= src.Rect.Min.Y {
		return nil, errors.New("PNG 编码图像为空")
	}
	return encodeScreenshotPNGRowSource(src)
}

// screenshotPNGSignature PNG 文件签名（8 字节，rodata 0x140b5c850 区段）。
var screenshotPNGSignature = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}

// [S-inline 汇编 0x140995fc0] absScreenshotPNGInt Paeth 预测器内联绝对值。
func absScreenshotPNGInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// [S-inline 汇编 0x140995e20] absScreenshotPNGByte 循环内联字节绝对值（int8 取负）。
func absScreenshotPNGByte(v byte) int {
	x := int(int8(v))
	if x < 0 {
		return -x
	}
	return x
}

// [S-inline 汇编 0x1409957c0] sumAbsScreenshotPNG filter0 代价累加内联。
func sumAbsScreenshotPNG(b []byte) int {
	sum := 0
	for _, v := range b {
		sum += absScreenshotPNGByte(v)
	}
	return sum
}
