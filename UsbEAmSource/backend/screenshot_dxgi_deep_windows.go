package main

import (
	"errors"
	"fmt"
	"image"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// 本文件承载 DXGI 深层 D3D11 / 桌面复制 / tone-map / 旋转 / 纹理换算桩。
// 批次 55：全部 [S-sig]/[P] 桩按 asm 直译落地为 [S]。
// 覆盖：输出目标打开、D3D11 设备创建、桌面复制、帧捕获（含 HDR tone-map）、
// 纹理/桌面矩形换算、像素旋转，以及浮点 LUT、sRGB 编码、DisplayConfig 解析链。

// ---------------------------------------------------------------------------
// D3D11 / DisplayConfig 常量与结构
// ---------------------------------------------------------------------------

// iidID3D11Texture2D 是 ID3D11Texture2D 的 IID（桌面复制帧资源 QueryInterface 目标）。
var iidID3D11Texture2D = windows.GUID{Data1: 0x6f15aaf2, Data2: 0xd208, Data3: 0x4e89, Data4: [8]byte{0x9a, 0xb4, 0x48, 0x95, 0x35, 0xd3, 0x4f, 0x9c}}

// d3d11Texture2DDesc 对应 D3D11_TEXTURE2D_DESC（11 字段，44 字节）。
type d3d11Texture2DDesc struct {
	Width          uint32
	Height         uint32
	MipLevels      uint32
	ArraySize      uint32
	Format         uint32
	SampleCount    uint32
	SampleQuality  uint32
	Usage          uint32
	BindFlags      uint32
	CPUAccessFlags uint32
	MiscFlags      uint32
}

// d3d11Box 对应 D3D11_BOX（6×uint32，24 字节）。
type d3d11Box struct {
	left   uint32
	top    uint32
	front  uint32
	right  uint32
	bottom uint32
	back   uint32
}

// d3d11MappedSubresource 对应 D3D11_MAPPED_SUBRESOURCE（16 字节）。
type d3d11MappedSubresource struct {
	pData      unsafe.Pointer
	rowPitch   uint32
	depthPitch uint32
}

// DisplayConfig 相关 DLL 入口（复用 appicon_windows.go 的 user32DLL）。
var (
	procGetDisplayConfigBufferSizes = user32DLL.NewProc("GetDisplayConfigBufferSizes")
	procQueryDisplayConfig          = user32DLL.NewProc("QueryDisplayConfig")
	procDisplayConfigGetDeviceInfo  = user32DLL.NewProc("DisplayConfigGetDeviceInfo")
)

// D3D11 功能级别 11_1 不支持哨兵（二进制 IAT 区消息不可读，占位可 errors.Is 比较）。
var errScreenshotD3D11FeatureLevel11_1Unsupported = errors.New("D3D11 功能级别 11_1 不支持")

// 浮点常量池（对齐 0x1411cd47c 起的 rodata 常量）。
const (
	screenshotDefaultSDRWhiteLevel = float32(2.5374999)
	screenshotToneMapFloor         = float32(0.001)
	screenshotToneMapHalf          = float32(0.5)
	screenshotDefaultSDRWhiteOut   = float32(0.55)
	screenshotToneMapOne           = float32(1.0)
	screenshotToneMapPeakFloor     = float32(1.001)
	screenshotSDRLinearLimit       = float32(1.01999998)
	screenshotDefaultRolloff       = float32(1.35)
	screenshotToneMapEight         = float32(8.0)
	screenshotToneMapSixteen       = float32(16.0)
	screenshotToneMapTwo           = float32(2.0)
	screenshotToneMapEighty        = float32(80.0)
	screenshotToneMapThousand      = float32(1000.0)
	screenshotToneMap4096          = float32(4096.0)
)

// screenshotDisplayConfigResolverGlobal 是显示器配置解析器（SDR 白点等）单例。
// [S] 目标全局 0x141096c38：值为空函数指针（二进制中恒为 nil）。
var screenshotDisplayConfigResolverGlobal func(screenshotDisplayCaptureInfo) (float32, bool)

// ---------------------------------------------------------------------------
// COM vtable 槽位读取（超出 IUnknown 前三槽的通用读取器）
// ---------------------------------------------------------------------------

// screenshotCOMVtableSlot 读取 COM 对象 vtable 指定槽位的函数指针。
// [S-inline] ASM 内联 `mov rdx,[rax]; mov rax,[rdx+off]` 模式。
func screenshotCOMVtableSlot(p uintptr, slot int) uintptr {
	if p == 0 {
		return 0
	}
	vtbl := *(**uintptr)(unsafe.Pointer(&p))
	if vtbl == nil {
		return 0
	}
	return *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(vtbl)) + uintptr(slot*8)))
}

// ---------------------------------------------------------------------------
// 浮点 LUT：half → float32 与 linear → sRGB byte
// ---------------------------------------------------------------------------

var (
	screenshotFloat16ToFloat32LookupOnce sync.Once
	screenshotFloat16ToFloat32LookupData []float32

	screenshotSRGBByteLookupOnce sync.Once
	screenshotSRGBByteLookupData []byte
)

// screenshotFloat16ToFloat32Lookup 惰性构建 65536 项 half→float32 查找表。
// [S] 构建闭包 0x1409f5c80：65536=0x10000 项，逐项 float16ToFloat32Uncached。
func screenshotFloat16ToFloat32Lookup() []float32 {
	screenshotFloat16ToFloat32LookupOnce.Do(func() {
		screenshotFloat16ToFloat32LookupData = make([]float32, 0x10000)
		for i := 0; i < 0x10000; i++ {
			screenshotFloat16ToFloat32LookupData[i] = screenshotFloat16ToFloat32Uncached(uint16(i))
		}
	})
	return screenshotFloat16ToFloat32LookupData
}

// screenshotSRGBByteLookup 惰性构建 4097 项 linear→sRGB byte 查找表。
// [S] 构建闭包 0x1409f5ae0：x=i/4096.0，sRGB 编码，round(x*255) 夹 [0,255]。
func screenshotSRGBByteLookup() []byte {
	screenshotSRGBByteLookupOnce.Do(func() {
		screenshotSRGBByteLookupData = make([]byte, 0x1001)
		for i := 0; i <= 0x1000; i++ {
			x := float64(i) / 4096.0
			var v float64
			if x <= 0.0031308 {
				v = 12.92 * x
			} else {
				v = 1.055*math.Pow(x, 1.0/2.4) - 0.055
			}
			y := int(math.Round(v * 255.0))
			if y < 0 {
				y = 0
			}
			if y > 255 {
				y = 255
			}
			screenshotSRGBByteLookupData[i] = byte(y)
		}
	})
	return screenshotSRGBByteLookupData
}

// screenshotFloat16ToFloat32Uncached 无缓存的 half→float32 转换。
// [S] ASM 0x14097dce0：标准 subnormal 归一化 + exp 偏移 112。
func screenshotFloat16ToFloat32Uncached(h uint16) float32 {
	sign := uint32(h&0x8000) << 16
	exp := uint32(h>>10) & 0x1f
	mant := uint32(h & 0x3ff)
	var bits uint32
	switch {
	case exp == 0:
		if mant == 0 {
			bits = sign
		} else {
			e := int32(1)
			for mant&0x400 == 0 {
				mant <<= 1
				e--
			}
			mant &= 0x3ff
			bits = sign | (uint32(e+112) << 23) | (mant << 13)
		}
	case exp == 31:
		bits = sign | 0x7f800000 | (mant << 13)
	default:
		bits = sign | ((exp + 112) << 23) | (mant << 13)
	}
	return math.Float32frombits(bits)
}

// linearFloatToSRGBByte 线性浮点 → sRGB 字节（查表）。
// [S] ASM 0x14097dd60：NaN/≤0→0；夹 [0,1]；idx=int(x*4096+0.5) 夹 [0,4096]。
func linearFloatToSRGBByte(x float32) uint8 {
	if !(x > 0) {
		return 0
	}
	if x > 1.0 {
		x = 1.0
	}
	idx := int(x*4096.0 + 0.5)
	if idx < 0 {
		idx = 0
	}
	if idx > 4096 {
		idx = 4096
	}
	return screenshotSRGBByteLookup()[idx]
}

// linearAlphaToByte 线性 alpha → 字节。
// [S] ASM 0x14097de80：NaN/≤0→0；≥1→255；否则 round(x*255) 夹 [0,255]。
func linearAlphaToByte(x float32) uint8 {
	if !(x > 0) {
		return 0
	}
	if x >= 1.0 {
		return 255
	}
	y := int(math.Round(float64(x * 255.0)))
	if y < 0 {
		y = 0
	}
	if y > 255 {
		y = 255
	}
	return uint8(y)
}

// [S-inline] 浮点夹 [0,1]。
func clamp01f(x float32) float32 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

// [S-inline] 从 data 偏移处读 little-endian uint16（half）。
func screenshotHalfAt(data unsafe.Pointer, off int) uint16 {
	return *(*uint16)(unsafe.Pointer(uintptr(data) + uintptr(off)))
}

// ---------------------------------------------------------------------------
// DXGI 格式调试名
// ---------------------------------------------------------------------------

// screenshotDXGIFormatDebugName 生成 DXGI 格式调试名（含数值）。
// [S] ASM 0x14097c780：按格式号 Sprintf 名称(%d)。
func screenshotDXGIFormatDebugName(format uint32) string {
	switch format {
	case 0xa:
		return fmt.Sprintf("R16G16B16A16_FLOAT(%d)", format)
	case 0x57:
		return fmt.Sprintf("B8G8R8A8_UNORM(%d)", format)
	case 0x58:
		return fmt.Sprintf("B8G8R8X8_UNORM(%d)", format)
	case 0x5b:
		return fmt.Sprintf("B8G8R8A8_UNORM_SRGB(%d)", format)
	case 0x5d:
		return fmt.Sprintf("B8G8R8X8_UNORM_SRGB(%d)", format)
	default:
		return fmt.Sprintf("UNKNOWN(%d)", format)
	}
}

// screenshotDXGIFormatListForDebug 格式化格式列表为 "[a, b]"。
// [S] ASM 0x14097c540：空→[]；否则逐项 join ", "。
func screenshotDXGIFormatListForDebug(formats []uint32) string {
	if len(formats) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(formats))
	for _, f := range formats {
		parts = append(parts, screenshotDXGIFormatDebugName(f))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// ---------------------------------------------------------------------------
// 捕获区域 / 纹理尺寸 / 旋转换算
// ---------------------------------------------------------------------------

// resolveScreenshotDXGIFrameCopyBounds 求捕获区域与纹理的交集。
// [S] ASM 0x14097c180：宽高≤0→空；rect 空→整帧；交为空→越界错误。
func resolveScreenshotDXGIFrameCopyBounds(rect image.Rectangle, desc d3d11Texture2DDesc) (image.Rectangle, error) {
	if desc.Width == 0 || desc.Height == 0 {
		return image.Rectangle{}, errors.New("DXGI 捕获纹理为空")
	}
	if rect.Empty() {
		return image.Rect(0, 0, int(desc.Width), int(desc.Height)), nil
	}
	full := image.Rect(0, 0, int(desc.Width), int(desc.Height))
	inter := rect.Intersect(full)
	if inter.Empty() {
		return image.Rectangle{}, fmt.Errorf("DXGI 区域不在捕获纹理内: rect=%v texture=%v", rect, full)
	}
	return inter, nil
}

// screenshotDXGITextureSizeForDesktop 由旋转值换算纹理尺寸。
// [S] ASM 0x14099e7e0：宽高≤0→无效；rotation 0/>4→不支持；2/4 交换。
func screenshotDXGITextureSizeForDesktop(rotation uint32, w, h int) (int, int, error) {
	if w <= 0 || h <= 0 {
		return 0, 0, errors.New("DXGI 桌面尺寸无效")
	}
	if rotation == 0 || rotation > 4 {
		return 0, 0, fmt.Errorf("不支持的 DXGI 显示器旋转值 %d", rotation)
	}
	if rotation == 2 || rotation == 4 {
		return h, w, nil
	}
	return w, h, nil
}

// screenshotDXGITextureRectForDesktopRect 将桌面矩形换算为纹理矩形。
// [S] ASM 0x14099e920：Canon 后校验桌面越界，按旋转映射，再校验纹理越界。
func screenshotDXGITextureRectForDesktopRect(rotation uint32, dx, dy int, rect image.Rectangle) (image.Rectangle, error) {
	r := rect.Canon()
	if r.Empty() || r != r.Intersect(image.Rect(0, 0, dx, dy)) {
		return image.Rectangle{}, fmt.Errorf("DXGI 桌面区域越界: %v", rect)
	}
	if rotation == 0 || rotation > 4 {
		return image.Rectangle{}, fmt.Errorf("不支持的 DXGI 显示器旋转值 %d", rotation)
	}
	var out image.Rectangle
	switch rotation {
	case 1:
		out = r
	case 2:
		out = image.Rect(r.Min.Y, dx-r.Max.X, r.Max.Y, dx-r.Min.X)
	case 3:
		out = image.Rect(dx-r.Max.X, dy-r.Max.Y, dx-r.Min.X, dy-r.Min.Y)
	case 4:
		out = image.Rect(dy-r.Max.Y, r.Min.X, dy-r.Min.Y, r.Max.X)
	}
	tw, th, _ := screenshotDXGITextureSizeForDesktop(rotation, dx, dy)
	if out.Empty() || out != out.Intersect(image.Rect(0, 0, tw, th)) {
		return image.Rectangle{}, fmt.Errorf("DXGI 纹理区域越界: %v", out)
	}
	return out, nil
}

// orientScreenshotDXGIFrame 按旋转方向旋转帧像素。
// [S] ASM 0x14099ee00：1 恒等；2 顺 90；3 180；4 逆 90；双循环逐像素重排。
func orientScreenshotDXGIFrame(img *image.RGBA, rotation uint32) (*image.RGBA, error) {
	if img == nil || img.Rect.Empty() {
		return nil, errors.New("DXGI 捕获帧为空")
	}
	if rotation == 0 || rotation > 4 {
		return nil, fmt.Errorf("不支持的 DXGI 显示器旋转值 %d", rotation)
	}
	if rotation == 1 {
		return img, nil
	}
	w := img.Rect.Dx()
	h := img.Rect.Dy()
	newW, newH := w, h
	if rotation == 2 || rotation == 4 {
		newW, newH = h, w
	}
	out := image.NewRGBA(image.Rect(0, 0, newW, newH))
	for dstRow := 0; dstRow < newH; dstRow++ {
		for dstCol := 0; dstCol < newW; dstCol++ {
			var srcRow, srcCol int
			switch rotation {
			case 2:
				srcRow = h - 1 - dstCol
				srcCol = dstRow
			case 3:
				srcRow = h - 1 - dstRow
				srcCol = w - 1 - dstCol
			case 4:
				srcRow = dstCol
				srcCol = w - 1 - dstRow
			}
			srcOff := srcRow*img.Stride + srcCol*4
			dstOff := dstRow*out.Stride + dstCol*4
			copy(out.Pix[dstOff:dstOff+4], img.Pix[srcOff:srcOff+4])
		}
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// 像素拷贝 / 近黑检测
// ---------------------------------------------------------------------------

// copyScreenshotDXGIBGRA BGRA → RGBA 拷贝（可选 alpha 交换为 0xff）。
// [S] ASM 0x14097c980：行跨度校验；逐像素 BGRA→RGBA。
func copyScreenshotDXGIBGRA(data unsafe.Pointer, rowPitch, depthPitch int, swapAlpha bool, desc d3d11Texture2DDesc) (*image.RGBA, error) {
	if rowPitch < int(desc.Width)*4 {
		return nil, errors.New("DXGI BGRA 行跨度小于图像宽度")
	}
	w := int(desc.Width)
	h := int(desc.Height)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		src := (*[1 << 30]byte)(data)[y*rowPitch : y*rowPitch+w*4]
		dst := img.Pix[y*img.Stride : y*img.Stride+w*4]
		for x := 0; x < w; x++ {
			dst[x*4+0] = src[x*4+2]
			dst[x*4+1] = src[x*4+1]
			dst[x*4+2] = src[x*4+0]
			if swapAlpha {
				dst[x*4+3] = src[x*4+3]
			} else {
				dst[x*4+3] = 0xff
			}
		}
	}
	return img, nil
}

// copyScreenshotDXGIRGBAFloat16 FP16 → RGBA 拷贝（SDR 快路径 + HDR tone-map 路径）。
// [S] ASM 0x14097cc00：整体 SDR 判定；否则逐像素 tone-map。
func copyScreenshotDXGIRGBAFloat16(data unsafe.Pointer, rowPitch, depthPitch int, opts screenshotHDRToneMapOptions, desc d3d11Texture2DDesc) (*image.RGBA, error) {
	if rowPitch < int(desc.Width)*8 {
		return nil, errors.New("DXGI FP16 行跨度小于图像宽度")
	}
	w := int(desc.Width)
	h := int(desc.Height)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	lut := screenshotFloat16ToFloat32Lookup()

	if screenshotDXGIMappedFloat16WithinSDRLinearRange(data, rowPitch, depthPitch, desc) {
		screenshotHDRCaptureDebugLog("DXGI 映射纹理位于 SDR 线性范围，直接转换")
		for y := 0; y < h; y++ {
			base := y * rowPitch
			dst := img.Pix[y*img.Stride:]
			for x := 0; x < w; x++ {
				off := base + x*8
				r := lut[screenshotHalfAt(data, off+0)]
				g := lut[screenshotHalfAt(data, off+2)]
				b := lut[screenshotHalfAt(data, off+4)]
				a := lut[screenshotHalfAt(data, off+6)]
				dst[x*4+0] = linearFloatToSRGBByte(r)
				dst[x*4+1] = linearFloatToSRGBByte(g)
				dst[x*4+2] = linearFloatToSRGBByte(b)
				dst[x*4+3] = linearAlphaToByte(a)
			}
		}
		return img, nil
	}

	rt := newScreenshotHDRToneMapRuntime(opts)
	for y := 0; y < h; y++ {
		base := y * rowPitch
		dst := img.Pix[y*img.Stride:]
		for x := 0; x < w; x++ {
			off := base + x*8
			r := lut[screenshotHalfAt(data, off+0)]
			g := lut[screenshotHalfAt(data, off+2)]
			b := lut[screenshotHalfAt(data, off+4)]
			a := lut[screenshotHalfAt(data, off+6)]
			if r <= screenshotSDRLinearLimit && g <= screenshotSDRLinearLimit && b <= screenshotSDRLinearLimit {
				dst[x*4+0] = linearFloatToSRGBByte(r)
				dst[x*4+1] = linearFloatToSRGBByte(g)
				dst[x*4+2] = linearFloatToSRGBByte(b)
			} else {
				rr, gg, bb := screenshotHDRToneMapRGBWithRuntime(r, g, b, rt)
				dst[x*4+0] = linearFloatToSRGBByte(rr)
				dst[x*4+1] = linearFloatToSRGBByte(gg)
				dst[x*4+2] = linearFloatToSRGBByte(bb)
			}
			dst[x*4+3] = linearAlphaToByte(a)
		}
	}
	return img, nil
}

// screenshotDXGIMappedTextureToRGBAWithToneMap 按格式分派像素拷贝。
// [S] ASM 0x14097c360：nil/0 校验；按 Format 分派；不支持格式报错。
func screenshotDXGIMappedTextureToRGBAWithToneMap(data unsafe.Pointer, rowPitch, depthPitch int, opts screenshotHDRToneMapOptions, desc d3d11Texture2DDesc) (*image.RGBA, error) {
	if data == nil || rowPitch == 0 || depthPitch == 0 || desc.Width == 0 || desc.Height == 0 {
		return nil, errors.New("DXGI 捕获纹理为空")
	}
	switch desc.Format {
	case 0xa:
		return copyScreenshotDXGIRGBAFloat16(data, rowPitch, depthPitch, opts, desc)
	case 0x57:
		return copyScreenshotDXGIBGRA(data, rowPitch, depthPitch, true, desc)
	case 0x58:
		return copyScreenshotDXGIBGRA(data, rowPitch, depthPitch, false, desc)
	case 0x5b:
		return copyScreenshotDXGIBGRA(data, rowPitch, depthPitch, true, desc)
	case 0x5d:
		return copyScreenshotDXGIBGRA(data, rowPitch, depthPitch, false, desc)
	}
	return nil, fmt.Errorf("暂不支持的 DXGI 截图格式: %s", screenshotDXGIFormatDebugName(desc.Format))
}

// [S-inline] 采样网格尺寸（每维最多 16）。
func screenshotSampleGrid(width, height int) (sampleW, sampleH int) {
	sampleW = width
	if sampleW > 16 {
		sampleW = 16
	}
	sampleH = height
	if sampleH > 16 {
		sampleH = 16
	}
	return
}

// screenshotDXGIMappedTextureLooksNearlyBlack 映射纹理近黑判定分派。
// [S] ASM 0x14097d4c0：按 Format 分派 Float16/BGRA 近黑判定。
func screenshotDXGIMappedTextureLooksNearlyBlack(data unsafe.Pointer, rowPitch, depthPitch int, desc d3d11Texture2DDesc) bool {
	if data == nil || rowPitch == 0 || depthPitch == 0 {
		return false
	}
	switch desc.Format {
	case 0xa:
		return screenshotDXGIMappedFloat16LooksNearlyBlack(data, rowPitch, depthPitch, int(desc.Width), int(desc.Height))
	case 0x57, 0x58, 0x5b, 0x5d:
		return screenshotDXGIMappedBGRALooksNearlyBlack(data, rowPitch, depthPitch, int(desc.Width), int(desc.Height))
	}
	return false
}

// screenshotDXGIMappedFloat16LooksNearlyBlack FP16 映射纹理近黑判定。
// [S] ASM 0x14097d780：采样网格，dark=R/G/B≤3.0，sampled*995<=dark*1000。
func screenshotDXGIMappedFloat16LooksNearlyBlack(data unsafe.Pointer, rowPitch, depthPitch, width, height int) bool {
	if data == nil || rowPitch < width*8 {
		return false
	}
	sampleW, sampleH := screenshotSampleGrid(width, height)
	lut := screenshotFloat16ToFloat32Lookup()
	sampled, dark := 0, 0
	for y := 0; y < sampleH; y++ {
		row := (height - 1) * y / max(1, sampleH-1)
		for x := 0; x < sampleW; x++ {
			col := (width - 1) * x / max(1, sampleW-1)
			off := row*rowPitch + col*8
			r := lut[screenshotHalfAt(data, off+0)]
			g := lut[screenshotHalfAt(data, off+2)]
			b := lut[screenshotHalfAt(data, off+4)]
			sampled++
			if r <= 3.0 && g <= 3.0 && b <= 3.0 {
				dark++
			}
		}
	}
	return sampled*995 <= dark*1000
}

// screenshotDXGIMappedBGRALooksNearlyBlack BGRA 映射纹理近黑判定。
// [S] ASM 0x14097d580：dark=R/G/B 字节≤3，sampled*995<=dark*1000。
func screenshotDXGIMappedBGRALooksNearlyBlack(data unsafe.Pointer, rowPitch, depthPitch, width, height int) bool {
	if data == nil || rowPitch < width*4 {
		return false
	}
	sampleW, sampleH := screenshotSampleGrid(width, height)
	sampled, dark := 0, 0
	for y := 0; y < sampleH; y++ {
		row := (height - 1) * y / max(1, sampleH-1)
		for x := 0; x < sampleW; x++ {
			col := (width - 1) * x / max(1, sampleW-1)
			off := row*rowPitch + col*4
			r := *(*byte)(unsafe.Pointer(uintptr(data) + uintptr(off+2)))
			g := *(*byte)(unsafe.Pointer(uintptr(data) + uintptr(off+1)))
			b := *(*byte)(unsafe.Pointer(uintptr(data) + uintptr(off+0)))
			sampled++
			if r <= 3 && g <= 3 && b <= 3 {
				dark++
			}
		}
	}
	return sampled*995 <= dark*1000
}

// screenshotDXGIMappedFloat16WithinSDRLinearRange 判定 FP16 映射纹理整体位于 SDR 线性范围。
// [S] ASM 0x14097d220：任一采样 NaN 或 >1.01999998 → false。
func screenshotDXGIMappedFloat16WithinSDRLinearRange(data unsafe.Pointer, rowPitch, depthPitch int, desc d3d11Texture2DDesc) bool {
	if data == nil || rowPitch == 0 || depthPitch == 0 {
		return false
	}
	if rowPitch < int(desc.Width)*8 {
		return false
	}
	w := int(desc.Width)
	h := int(desc.Height)
	sampleW, sampleH := screenshotSampleGrid(w, h)
	lut := screenshotFloat16ToFloat32Lookup()
	for y := 0; y < sampleH; y++ {
		row := (h - 1) * y / max(1, sampleH-1)
		for x := 0; x < sampleW; x++ {
			col := (w - 1) * x / max(1, sampleW-1)
			off := row*rowPitch + col*8
			r := lut[screenshotHalfAt(data, off+0)]
			g := lut[screenshotHalfAt(data, off+2)]
			b := lut[screenshotHalfAt(data, off+4)]
			if !(r <= screenshotSDRLinearLimit) || !(g <= screenshotSDRLinearLimit) || !(b <= screenshotSDRLinearLimit) {
				return false
			}
		}
	}
	return true
}

// screenshotRGBAImageLooksNearlyBlack RGBA 图像近黑判定。
// [S] ASM 0x14097da60：dark=R/G/B≤3，sampled*995<=dark*1000。
func screenshotRGBAImageLooksNearlyBlack(img *image.RGBA) bool {
	if img == nil || img.Rect.Empty() {
		return false
	}
	w := img.Rect.Dx()
	h := img.Rect.Dy()
	sampleW, sampleH := screenshotSampleGrid(w, h)
	sampled, dark := 0, 0
	for y := 0; y < sampleH; y++ {
		row := (h - 1) * y / max(1, sampleH-1)
		for x := 0; x < sampleW; x++ {
			col := (w - 1) * x / max(1, sampleW-1)
			off := row*img.Stride + col*4
			sampled++
			if img.Pix[off+0] <= 3 && img.Pix[off+1] <= 3 && img.Pix[off+2] <= 3 {
				dark++
			}
		}
	}
	return sampled*995 <= dark*1000
}

// ---------------------------------------------------------------------------
// HDR tone-map 数学
// ---------------------------------------------------------------------------

// screenshotHDRToneMapRuntime 是 tone-map 运行时参数（peak、B、A、sdrWhiteLevel）。
type screenshotHDRToneMapRuntime struct {
	peak          float32
	b             float32
	a             float32
	sdrWhiteLevel float32
}

// [S] clampScreenshotHDRToneMapOptions 应用共享 clamp 规则。
// sdrWhiteLevel NaN/≤0→1.0；sdrWhiteOutput NaN/≤0/≥1→0.55；
// maxInputLevel NaN 或 sdrWhiteLevel>=maxInputLevel→8.0；highlightRolloff NaN/≤0→1.35。
func clampScreenshotHDRToneMapOptions(opts screenshotHDRToneMapOptions) screenshotHDRToneMapOptions {
	s := opts.SDRWhiteLevel
	if !(s > 0) {
		s = 1.0
	}
	o := opts.SDRWhiteOutput
	if !(o > 0) || !(o < 1.0) {
		o = screenshotDefaultSDRWhiteOut
	}
	m := opts.MaxInputLevel
	if !(s < m) {
		m = screenshotToneMapEight
	}
	r := opts.HighlightRolloff
	if !(r > 0) {
		r = screenshotDefaultRolloff
	}
	return screenshotHDRToneMapOptions{SDRWhiteLevel: s, SDRWhiteOutput: o, MaxInputLevel: m, HighlightRolloff: r}
}

// newScreenshotHDRToneMapRuntime 由 tone-map 参数构造运行时。
// [S] ASM 0x1409ad9..: peak=max(maxInput/sdrWhite,1.001)；m=1.0（退化）；
// refWhite=clamp(0.5*(1/peak²+1),0.001,1.0)；B=1/m；A=sdrWhiteOutput/refWhite。
func newScreenshotHDRToneMapRuntime(opts screenshotHDRToneMapOptions) screenshotHDRToneMapRuntime {
	c := clampScreenshotHDRToneMapOptions(opts)
	peak := c.MaxInputLevel / c.SDRWhiteLevel
	if peak < screenshotToneMapPeakFloor {
		peak = screenshotToneMapPeakFloor
	}
	m := float32(1.0)
	refWhite := 0.5 * (1.0/(peak*peak) + 1.0) / m
	if refWhite < screenshotToneMapFloor {
		refWhite = screenshotToneMapFloor
	}
	if refWhite > screenshotToneMapOne {
		refWhite = screenshotToneMapOne
	}
	return screenshotHDRToneMapRuntime{
		peak:          peak,
		b:             1.0 / m,
		a:             c.SDRWhiteOutput / refWhite,
		sdrWhiteLevel: c.SDRWhiteLevel,
	}
}

// screenshotHDRToneMapRGBWithRuntime 对单像素 RGB 应用 tone-map。
// [S] ASM 0x1409ada..: 夹 NaN/≤0；peakRGB；q=peakRGB/sdrWhite；
// mapped=q*(q/peak²+1)/(q+1)；scale=A*B*mapped 夹 [0,1]；scale/=peakRGB；逐通道乘夹。
func screenshotHDRToneMapRGBWithRuntime(r, g, b float32, rt screenshotHDRToneMapRuntime) (float32, float32, float32) {
	if !(r > 0) {
		r = 0
	}
	if !(g > 0) {
		g = 0
	}
	if !(b > 0) {
		b = 0
	}
	peakRGB := r
	if g > peakRGB {
		peakRGB = g
	}
	if b > peakRGB {
		peakRGB = b
	}
	if !(peakRGB > 0) {
		return 0, 0, 0
	}
	scale := float32(0)
	if rt.b > 0 && rt.a > 0 {
		q := peakRGB / rt.sdrWhiteLevel
		mapped := q * (q/(rt.peak*rt.peak) + 1.0) / (q + 1.0)
		scale = rt.a * rt.b * mapped
		scale = clamp01f(scale)
		scale /= peakRGB
	}
	return clamp01f(r * scale), clamp01f(g * scale), clamp01f(b * scale)
}

// screenshotHDRToneMapRuntimeRGBToSRGBBytes tone-map 后转 sRGB 字节。
// [S] ASM 0x1409adaa..: toneMapRGBWithRuntime 后 linearFloatToSRGBByte×3。
func screenshotHDRToneMapRuntimeRGBToSRGBBytes(r, g, b float32, rt screenshotHDRToneMapRuntime) (uint8, uint8, uint8) {
	r, g, b = screenshotHDRToneMapRGBWithRuntime(r, g, b, rt)
	return linearFloatToSRGBByte(r), linearFloatToSRGBByte(g), linearFloatToSRGBByte(b)
}

// parseScreenshotHDRToneMapEnvFloat 解析 tone-map 环境变量浮点。
// [S] ASM 0x1409adc00：TrimSpace(Getenv)；空→(0,false)；解析失败→日志+(0,false)。
func parseScreenshotHDRToneMapEnvFloat(name string) (float32, bool) {
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(v, 32)
	if err != nil {
		screenshotHDRCaptureDebugLog("忽略无效 HDR tone mapping 环境变量: %s=%q", name, v)
		return 0, false
	}
	return float32(f), true
}

// applyScreenshotHDRToneMapEnvOverrides 应用环境变量覆盖。
// [S] ASM 0x1409adb40：SDR_WHITE_LEVEL / SDR_WHITE_OUTPUT 两个 env 覆盖。
func applyScreenshotHDRToneMapEnvOverrides(opts *screenshotHDRToneMapOptions) {
	if v, ok := parseScreenshotHDRToneMapEnvFloat("USBEAM_SCREENSHOT_HDR_SDR_WHITE_LEVEL"); ok {
		opts.SDRWhiteLevel = v
	}
	if v, ok := parseScreenshotHDRToneMapEnvFloat("USBEAM_SCREENSHOT_HDR_SDR_WHITE_OUTPUT"); ok {
		opts.SDRWhiteOutput = v
	}
}

// screenshotHDRToneMapOptionsForDisplayWithSDRWhiteResolver 计算 HDR tone-map 参数。
// [S] ASM 0x1409ad8e0：HDR 时解析 SDR 白点；maxInput 由亮度推导；env 覆盖 + clamp。
func screenshotHDRToneMapOptionsForDisplayWithSDRWhiteResolver(d screenshotDisplayCaptureInfo, resolver func(screenshotDisplayCaptureInfo) (float32, bool)) screenshotHDRToneMapOptions {
	sdrWhiteLevel := float32(1.0)
	if d.HDR {
		sdrWhiteLevel = screenshotDefaultSDRWhiteLevel
		if resolver != nil {
			if v, ok := resolver(d); ok {
				if v < screenshotDefaultSDRWhiteLevel {
					screenshotHDRCaptureDebugLog("DXGI SDR white level 低于下限，使用默认值: %.3f -> %.3f", v, screenshotDefaultSDRWhiteLevel)
					sdrWhiteLevel = screenshotDefaultSDRWhiteLevel
				} else {
					sdrWhiteLevel = v
				}
			}
		}
	}

	maxLum := float32(math.Abs(float64(d.MaxLuminance)))
	if a := float32(math.Abs(float64(d.MaxFullFrameLuminance))); a > maxLum {
		maxLum = a
	}
	maxInputLevel := screenshotToneMapEight
	if maxLum > screenshotToneMapEighty {
		maxInputLevel = min(screenshotToneMapSixteen, max(screenshotToneMapTwo, maxLum/screenshotToneMapEighty))
	}

	opts := screenshotHDRToneMapOptions{
		SDRWhiteLevel:    sdrWhiteLevel,
		SDRWhiteOutput:   screenshotDefaultSDRWhiteOut,
		MaxInputLevel:    maxInputLevel,
		HighlightRolloff: screenshotDefaultRolloff,
	}
	applyScreenshotHDRToneMapEnvOverrides(&opts)
	return clampScreenshotHDRToneMapOptions(opts)
}

// formatScreenshotHDRToneMapOptionsForDebug 格式化 tone-map 参数用于调试日志。
// [S] ASM 0x1409ad660：clamp 后 peak/refWhite，格式 5 字段。
func formatScreenshotHDRToneMapOptionsForDebug(opts screenshotHDRToneMapOptions) string {
	c := clampScreenshotHDRToneMapOptions(opts)
	peak := c.MaxInputLevel / c.SDRWhiteLevel
	if peak < screenshotToneMapPeakFloor {
		peak = screenshotToneMapPeakFloor
	}
	refWhite := 0.5 * (1.0/(peak*peak) + 1.0)
	if refWhite < screenshotToneMapFloor {
		refWhite = screenshotToneMapFloor
	}
	if refWhite > screenshotToneMapOne {
		refWhite = screenshotToneMapOne
	}
	return fmt.Sprintf("sdrWhiteLevel=%.3f sdrWhiteOutput=%.3f maxInputLevel=%.3f highlightRolloff=%.3f referenceWhiteOutput=%.3f",
		c.SDRWhiteLevel, c.SDRWhiteOutput, c.MaxInputLevel, c.HighlightRolloff, refWhite)
}

// ---------------------------------------------------------------------------
// 输出目标打开 / D3D11 设备创建
// ---------------------------------------------------------------------------

// openScreenshotDXGIOutputTarget 打开指定输出的 IDXGIOutput6 与复制源适配器。
// [S] ASM 0x140979a40：createDXGIFactory1→queryDXGIFactory6→EnumAdapters→
// EnumOutputs→queryScreenshotDXGIOutput6；release 闭包释放 output6→adapter→factory1。
func openScreenshotDXGIOutputTarget(d screenshotDisplayCaptureInfo) (adapter uintptr, output6 uintptr, release func(), err error) {
	factory1, err := createDXGIFactory1()
	if err != nil {
		return 0, 0, nil, err
	}
	factory6, err := queryDXGIFactory6(factory1)
	if err != nil {
		releaseDXGIUnknown(factory1)
		return 0, 0, nil, err
	}

	var adapterPtr uintptr
	hresult, _, _ := syscall.SyscallN(
		dxgiFactoryVtable(factory6).EnumAdapters,
		factory6,
		uintptr(d.AdapterIndex),
		uintptr(unsafe.Pointer(&adapterPtr)),
	)
	releaseDXGIUnknown(factory6)
	if hresult != 0 {
		releaseDXGIUnknown(factory1)
		return 0, 0, nil, fmt.Errorf("%s failed: HRESULT 0x%08X", "IDXGIFactory.EnumAdapters", uint32(hresult))
	}
	if adapterPtr == 0 {
		releaseDXGIUnknown(factory1)
		return 0, 0, nil, errors.New("DXGI 适配器为空")
	}

	var output uintptr
	hresult, _, _ = syscall.SyscallN(
		dxgiAdapterVtable(adapterPtr).EnumOutputs,
		adapterPtr,
		uintptr(d.OutputIndex),
		uintptr(unsafe.Pointer(&output)),
	)
	if hresult != 0 {
		releaseDXGIUnknown(adapterPtr)
		releaseDXGIUnknown(factory1)
		return 0, 0, nil, fmt.Errorf("%s failed: HRESULT 0x%08X", "IDXGIAdapter.EnumOutputs", uint32(hresult))
	}
	if output == 0 {
		releaseDXGIUnknown(adapterPtr)
		releaseDXGIUnknown(factory1)
		return 0, 0, nil, errors.New("DXGI 输出为空")
	}

	output6, err = queryScreenshotDXGIOutput6(output)
	releaseDXGIUnknown(output)
	if err != nil {
		releaseDXGIUnknown(adapterPtr)
		releaseDXGIUnknown(factory1)
		return 0, 0, nil, err
	}

	release = func() {
		releaseDXGIUnknown(output6)
		releaseDXGIUnknown(adapterPtr)
		releaseDXGIUnknown(factory1)
	}
	return adapterPtr, output6, release, nil
}

// createScreenshotD3D11Device 从输出适配器创建 D3D11 设备（功能级别协商）。
// [S] ASM 0x140979fa0：先 4 级（11_1..10_0），11_1 不支持哨兵则降 3 级重试。
func createScreenshotD3D11Device(adapter uintptr) (*screenshotD3D11Device, error) {
	dev, err := createScreenshotD3D11DeviceWithFeatureLevels(adapter, []uint32{0xb100, 0xb000, 0xa100, 0xa000})
	if err != nil && errors.Is(err, errScreenshotD3D11FeatureLevel11_1Unsupported) {
		return createScreenshotD3D11DeviceWithFeatureLevels(adapter, []uint32{0xb000, 0xa100, 0xa000})
	}
	return dev, err
}

// createScreenshotD3D11DeviceWithFeatureLevels 按功能级别列表创建设备。
// [S] ASM 0x14097a080：nil/空校验；D3D11CreateDevice 10 参；E_INVALIDARG+11_1→哨兵。
func createScreenshotD3D11DeviceWithFeatureLevels(adapter uintptr, levels []uint32) (*screenshotD3D11Device, error) {
	if adapter == 0 {
		return nil, errors.New("D3D11 适配器为空")
	}
	if len(levels) == 0 {
		return nil, errors.New("D3D11 feature level 列表为空")
	}
	if err := procD3D11CreateDevice.Find(); err != nil {
		return nil, err
	}
	var device, context unsafe.Pointer
	var featureLevel uint32
	hresult, _, _ := syscall.SyscallN(
		procD3D11CreateDevice.Addr(),
		adapter,
		uintptr(0), // D3D_DRIVER_TYPE_UNKNOWN
		uintptr(0), // Software
		uintptr(0), // Flags
		uintptr(unsafe.Pointer(&levels[0])),
		uintptr(len(levels)),
		uintptr(7), // D3D11_SDK_VERSION
		uintptr(unsafe.Pointer(&device)),
		uintptr(unsafe.Pointer(&featureLevel)),
		uintptr(unsafe.Pointer(&context)),
	)
	if hresult == 0x80070057 && levels[0] == 0xb100 { // E_INVALIDARG
		return nil, errScreenshotD3D11FeatureLevel11_1Unsupported
	}
	if hresult != 0 {
		return nil, fmt.Errorf("%s failed: HRESULT 0x%08X", "D3D11CreateDevice", uint32(hresult))
	}
	if device == nil || context == nil {
		if device != nil {
			screenshotCOMRelease(uintptr(device))
		}
		if context != nil {
			screenshotCOMRelease(uintptr(context))
		}
		return nil, errors.New("D3D11 创建设备结果为空")
	}
	return &screenshotD3D11Device{device: device, context: context, featureLevel: featureLevel}, nil
}

// release 释放 D3D11 设备与上下文。
// [S] ASM 0x140979425：nil 安全，先 context 后 device。
func (d *screenshotD3D11Device) release() {
	if d == nil {
		return
	}
	if d.context != nil {
		screenshotCOMRelease(uintptr(d.context))
		d.context = nil
	}
	if d.device != nil {
		screenshotCOMRelease(uintptr(d.device))
		d.device = nil
	}
}

// ---------------------------------------------------------------------------
// 桌面复制
// ---------------------------------------------------------------------------

// duplicateScreenshotDXGIOutput 在输出上创建桌面复制对象（HDR 优先 FP16）。
// [S] ASM 0x14097a520：hdr 先 FP16-only，失败降级 [FP16,BGRA]。
func duplicateScreenshotDXGIOutput(output uintptr, device unsafe.Pointer, hdr bool) (unsafe.Pointer, error) {
	if hdr {
		dup, err := duplicateScreenshotDXGIOutputWithFormats(output, device, []uint32{0xa})
		if err == nil {
			screenshotHDRCaptureDebugLog("DXGI 输出复制优先使用 FP16 格式")
			return dup, nil
		}
		screenshotHDRCaptureDebugLog("DXGI FP16-only 输出复制失败，降级尝试兼容格式列表: %v", err)
	}
	return duplicateScreenshotDXGIOutputWithFormats(output, device, []uint32{0xa, 0x57})
}

// duplicateScreenshotDXGIOutputWithFormats 按格式列表创建桌面复制对象。
// [S] ASM 0x14097a660：nil/空校验；DuplicateOutput1@0xd0；结果空→错误；成功日志。
func duplicateScreenshotDXGIOutputWithFormats(output uintptr, device unsafe.Pointer, formats []uint32) (unsafe.Pointer, error) {
	if output == 0 || device == nil {
		return nil, errors.New("DXGI 输出或 D3D11 设备为空")
	}
	if len(formats) == 0 {
		return nil, errors.New("DXGI 输出复制格式列表为空")
	}
	var dup unsafe.Pointer
	hresult, _, _ := syscall.SyscallN(
		screenshotCOMVtableSlot(output, 26), // IDXGIOutput6.DuplicateOutput1 @0xd0
		output,
		uintptr(device),
		uintptr(0),
		uintptr(unsafe.Pointer(&formats[0])),
		uintptr(len(formats)),
		uintptr(unsafe.Pointer(&dup)),
	)
	if hresult != 0 {
		return nil, fmt.Errorf("%s failed: HRESULT 0x%08X", "IDXGIOutput5.DuplicateOutput1", uint32(hresult))
	}
	if dup == nil {
		return nil, errors.New("DXGI 输出复制对象为空")
	}
	screenshotHDRCaptureDebugLog("DXGI 输出复制格式列表: %s", screenshotDXGIFormatListForDebug(formats))
	return dup, nil
}

// ---------------------------------------------------------------------------
// 帧捕获
// ---------------------------------------------------------------------------

// captureScreenshotDXGIFrameRegion 重试循环捕获一帧（或指定区域）。
// [S] ASM 0x14097a920：3 次；超时哨兵或近黑帧 → 日志 + 重试；3 次耗尽 → 超时哨兵。
func captureScreenshotDXGIFrameRegion(device *screenshotD3D11Device, output uintptr, opts screenshotHDRToneMapOptions, rect image.Rectangle) (*image.RGBA, error) {
	for attempt := 0; attempt < 3; attempt++ {
		img, err := captureScreenshotDXGIFrameRegionOnce(device, output, opts, rect)
		if err != nil {
			if errors.Is(err, errScreenshotDXGICaptureFrameTimeout) {
				screenshotHDRCaptureDebugLog("DXGI 捕获疑似黑帧，准备重试: attempt=%d/%d", attempt+1, 3)
				continue
			}
			return nil, err
		}
		if screenshotRGBAImageLooksNearlyBlack(img) {
			screenshotHDRCaptureDebugLog("DXGI 捕获疑似黑帧，准备重试: attempt=%d/%d", attempt+1, 3)
			continue
		}
		if attempt > 0 {
			screenshotHDRCaptureDebugLog("DXGI 黑帧重试后取得有效画面: attempts=%d", attempt+1)
		}
		return img, nil
	}
	return nil, errScreenshotDXGICaptureFrameTimeout
}

// captureScreenshotDXGIFrameRegionOnce 单次捕获一帧（或指定区域）。
// [S] ASM 0x14097ac40：AcquireNextFrame 循环→QueryInterface 纹理→GetDesc→
// 区域交集→staging 纹理→Copy→Map→近黑判定→像素拷贝。
func captureScreenshotDXGIFrameRegionOnce(device *screenshotD3D11Device, output uintptr, opts screenshotHDRToneMapOptions, rect image.Rectangle) (*image.RGBA, error) {
	if device == nil || device.context == nil || output == 0 {
		return nil, errors.New("DXGI 捕获上下文为空")
	}

	var frameInfo [48]byte // DXGI_OUTDUPL_FRAME_INFO
	var resource uintptr
	acquired := false
	for attempt := 0; attempt < 3; attempt++ {
		hresult, _, _ := syscall.SyscallN(
			screenshotCOMVtableSlot(output, 8), // AcquireNextFrame @0x40
			output,
			uintptr(100), // 100ms
			uintptr(unsafe.Pointer(&frameInfo)),
			uintptr(unsafe.Pointer(&resource)),
		)
		if hresult == 0 {
			acquired = true
			break
		}
		if hresult == 0x887a0027 { // DXGI_ERROR_ACCESS_LOST
			continue
		}
		return nil, fmt.Errorf("%s failed: HRESULT 0x%08X", "IDXGIOutputDuplication.AcquireNextFrame", uint32(hresult))
	}
	if !acquired || resource == 0 {
		return nil, errScreenshotDXGICaptureFrameTimeout
	}
	defer func() {
		syscall.SyscallN(screenshotCOMVtableSlot(output, 14), output) // ReleaseFrame @0x70
	}()
	defer releaseDXGIUnknown(resource)

	texture, err := screenshotCOMQueryInterface(resource, &iidID3D11Texture2D)
	if err != nil {
		return nil, err
	}
	defer releaseDXGIUnknown(texture)

	var desc d3d11Texture2DDesc
	syscall.SyscallN(
		screenshotCOMVtableSlot(texture, 10), // GetDesc @0x50
		texture,
		uintptr(unsafe.Pointer(&desc)),
	)
	if desc.Width == 0 || desc.Height == 0 {
		return nil, errors.New("DXGI 捕获纹理为空")
	}

	inter, err := resolveScreenshotDXGIFrameCopyBounds(rect, desc)
	if err != nil {
		return nil, err
	}

	screenshotHDRCaptureDebugLog("DXGI 捕获纹理: size=%dx%d format=%s array=%d mips=%d sample=%d usage=%d bind=0x%X cpu=0x%X misc=0x%X",
		desc.Width, desc.Height, screenshotDXGIFormatDebugName(desc.Format), desc.ArraySize, desc.MipLevels,
		desc.SampleCount, desc.Usage, desc.BindFlags, desc.CPUAccessFlags, desc.MiscFlags)

	stagingDesc := desc
	stagingDesc.Width = uint32(inter.Dx())
	stagingDesc.Height = uint32(inter.Dy())
	stagingDesc.Usage = 3                // D3D11_USAGE_STAGING
	stagingDesc.CPUAccessFlags = 0x20000 // D3D11_CPU_ACCESS_READ

	var staging uintptr
	hresult, _, _ := syscall.SyscallN(
		screenshotCOMVtableSlot(uintptr(device.device), 5), // CreateTexture2D @0x28
		uintptr(device.device),
		uintptr(unsafe.Pointer(&stagingDesc)),
		uintptr(0),
		uintptr(unsafe.Pointer(&staging)),
	)
	if hresult != 0 {
		return nil, fmt.Errorf("%s failed: HRESULT 0x%08X", "ID3D11Device.CreateTexture2D", uint32(hresult))
	}
	if staging == 0 {
		return nil, errors.New("DXGI staging 纹理为空")
	}
	defer releaseDXGIUnknown(staging)

	ctx := uintptr(device.context)
	fullFrame := inter.Min == image.Pt(0, 0) && inter.Max == image.Pt(int(desc.Width), int(desc.Height))
	if fullFrame {
		syscall.SyscallN(
			screenshotCOMVtableSlot(ctx, 47), // CopyResource @0x178
			ctx,
			staging,
			texture,
		)
	} else {
		box := d3d11Box{
			left:   uint32(inter.Min.X),
			top:    uint32(inter.Min.Y),
			front:  0,
			right:  uint32(inter.Max.X),
			bottom: uint32(inter.Max.Y),
			back:   1,
		}
		syscall.SyscallN(
			screenshotCOMVtableSlot(ctx, 46), // CopySubresourceRegion @0x170
			ctx,
			staging, 0, 0, 0, 0,
			texture, 0,
			uintptr(unsafe.Pointer(&box)),
		)
	}

	screenshotHDRCaptureDebugLog("DXGI 区域复制纹理: source=%v size=%dx%d", inter, inter.Dx(), inter.Dy())

	var mapped d3d11MappedSubresource
	hresult, _, _ = syscall.SyscallN(
		screenshotCOMVtableSlot(ctx, 14), // Map @0x70
		ctx,
		staging,
		uintptr(0), // Subresource
		uintptr(1), // D3D11_MAP_READ
		uintptr(0), // MapFlags
		uintptr(unsafe.Pointer(&mapped)),
	)
	if hresult != 0 {
		return nil, fmt.Errorf("%s failed: HRESULT 0x%08X", "ID3D11DeviceContext.Map", uint32(hresult))
	}
	defer func() {
		syscall.SyscallN(screenshotCOMVtableSlot(ctx, 15), ctx, staging, uintptr(0)) // Unmap @0x78
	}()

	screenshotHDRCaptureDebugLog("DXGI 映射纹理: rowPitch=%d depthPitch=%d format=%s",
		mapped.rowPitch, mapped.depthPitch, screenshotDXGIFormatDebugName(stagingDesc.Format))

	if screenshotDXGIMappedTextureLooksNearlyBlack(mapped.pData, int(mapped.rowPitch), int(mapped.depthPitch), stagingDesc) {
		return nil, errScreenshotDXGICaptureFrameTimeout
	}

	return screenshotDXGIMappedTextureToRGBAWithToneMap(mapped.pData, int(mapped.rowPitch), int(mapped.depthPitch), opts, stagingDesc)
}

// ---------------------------------------------------------------------------
// DisplayConfig 解析链（解析器全局为 nil，此链为忠实复现的死代码）
// ---------------------------------------------------------------------------

// normalizeScreenshotDisplayConfigDeviceName 规范化设备名：TrimSpace + ToUpper。
// [S] ASM 0x140975f80：TrimSpace 后 ToUpper。
func normalizeScreenshotDisplayConfigDeviceName(name string) string {
	return strings.ToUpper(strings.TrimSpace(name))
}

// queryScreenshotDisplayConfigSourceName 查询源设备 GDI 名称。
// [S] ASM 0x140975d20：DisplayConfigGetDeviceInfo(GET_SOURCE_NAME)。
func queryScreenshotDisplayConfigSourceName(adapterID windows.LUID, id uint32) (string, error) {
	if err := procDisplayConfigGetDeviceInfo.Find(); err != nil {
		return "", err
	}
	var info displayconfigSourceDeviceName
	info.Header.Type = 1 // DISPLAYCONFIG_DEVICE_INFO_GET_SOURCE_NAME
	info.Header.Size = 0x54
	info.Header.AdapterID = adapterID
	info.Header.ID = id
	hresult, _, _ := syscall.SyscallN(procDisplayConfigGetDeviceInfo.Addr(), uintptr(unsafe.Pointer(&info.Header)))
	if hresult != 0 {
		return "", fmt.Errorf("%s failed: HRESULT 0x%08X", "DisplayConfigGetDeviceInfo", uint32(hresult))
	}
	return strings.TrimSpace(syscall.UTF16ToString(info.ViewGDIDeviceName[:])), nil
}

// queryScreenshotDisplayConfigTargetSDRWhiteLevel 查询目标 SDR 白点（原始值）。
// [S] ASM 0x140975e60：DisplayConfigGetDeviceInfo(GET_SDR_WHITE_LEVEL)。
func queryScreenshotDisplayConfigTargetSDRWhiteLevel(adapterID windows.LUID, id uint32) (uint32, error) {
	if err := procDisplayConfigGetDeviceInfo.Find(); err != nil {
		return 0, err
	}
	var info displayconfigSDRWhiteLevel
	info.Header.Type = 11 // DISPLAYCONFIG_DEVICE_INFO_GET_SDR_WHITE_LEVEL
	info.Header.Size = 0x18
	info.Header.AdapterID = adapterID
	info.Header.ID = id
	hresult, _, _ := syscall.SyscallN(procDisplayConfigGetDeviceInfo.Addr(), uintptr(unsafe.Pointer(&info.Header)))
	if hresult != 0 {
		return 0, fmt.Errorf("%s failed: HRESULT 0x%08X", "DisplayConfigGetDeviceInfo", uint32(hresult))
	}
	return info.SDRWhiteLevel, nil
}

// queryScreenshotDisplayConfigActivePaths 查询活动显示路径。
// [S] ASM 0x140975620：GetDisplayConfigBufferSizes → QueryDisplayConfig；
// ERROR_INSUFFICIENT_BUFFER 重试（最多 4 次）。
func queryScreenshotDisplayConfigActivePaths() ([]displayconfigPathInfo, []displayconfigModeInfo, error) {
	if err := procGetDisplayConfigBufferSizes.Find(); err != nil {
		return nil, nil, err
	}
	if err := procQueryDisplayConfig.Find(); err != nil {
		return nil, nil, err
	}
	for attempt := 0; attempt < 4; attempt++ {
		var numPath, numMode uint32
		hresult, _, _ := syscall.SyscallN(
			procGetDisplayConfigBufferSizes.Addr(),
			uintptr(1), // QDC_ALL_PATHS
			uintptr(unsafe.Pointer(&numPath)),
			uintptr(unsafe.Pointer(&numMode)),
		)
		if hresult != 0 {
			return nil, nil, fmt.Errorf("%s failed: %d", "GetDisplayConfigBufferSizes", uint32(hresult))
		}
		if numPath == 0 {
			return nil, nil, nil
		}
		if numMode == 0 {
			numMode = 1
		}
		paths := make([]displayconfigPathInfo, numPath)
		modes := make([]displayconfigModeInfo, numMode)
		var topologyID uint32
		hresult, _, _ = syscall.SyscallN(
			procQueryDisplayConfig.Addr(),
			uintptr(1), // QDC_ALL_PATHS
			uintptr(unsafe.Pointer(&numPath)),
			uintptr(unsafe.Pointer(&paths[0])),
			uintptr(unsafe.Pointer(&numMode)),
			uintptr(unsafe.Pointer(&modes[0])),
			uintptr(unsafe.Pointer(&topologyID)),
		)
		if hresult == 122 { // ERROR_INSUFFICIENT_BUFFER
			continue
		}
		if hresult != 0 {
			return nil, nil, fmt.Errorf("%s failed: %d", "QueryDisplayConfig", uint32(hresult))
		}
		return paths[:numPath], modes[:numMode], nil
	}
	return nil, nil, errors.New("QueryDisplayConfig 缓冲不足，重试次数耗尽")
}

// findScreenshotDisplayConfigPathForDisplay 在路径中按规范化源名匹配显示器。
// [S] ASM 0x140975980：逐路径 querySourceName 并规范化比较，命中即返回。
func findScreenshotDisplayConfigPathForDisplay(paths []displayconfigPathInfo, deviceName string) (displayconfigPathInfo, bool) {
	if len(paths) == 0 {
		return displayconfigPathInfo{}, false
	}
	normalized := normalizeScreenshotDisplayConfigDeviceName(deviceName)
	for _, path := range paths {
		sourceName, err := queryScreenshotDisplayConfigSourceName(path.SourceInfo.AdapterID, path.SourceInfo.ID)
		if err != nil {
			screenshotHDRCaptureDebugLog("读取 Windows DisplayConfig source name 失败: sourceID=%d err=%v", path.SourceInfo.ID, err)
			continue
		}
		if normalizeScreenshotDisplayConfigDeviceName(sourceName) == normalized {
			return path, true
		}
	}
	return displayconfigPathInfo{}, false
}

// queryScreenshotDisplayConfigSDRWhiteLevelForDisplay 查询显示器 SDR 白点。
// [S] ASM 0x140975000：TrimSpace(displayName)→活动路径→匹配路径→目标白点；
// sdrWhite=raw/1000，nits=sdrWhite*80（仅日志）。
func queryScreenshotDisplayConfigSDRWhiteLevelForDisplay(displayName string, adapterName string) (float32, bool) {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return 0, false
	}
	paths, _, err := queryScreenshotDisplayConfigActivePaths()
	if err != nil {
		screenshotHDRCaptureDebugLog("读取 Windows DisplayConfig 活动路径失败: display=%q err=%v", displayName, err)
		return 0, false
	}
	path, ok := findScreenshotDisplayConfigPathForDisplay(paths, displayName)
	if !ok {
		screenshotHDRCaptureDebugLog("未匹配到 Windows DisplayConfig 路径: display=%q paths=%d", displayName, len(paths))
		return 0, false
	}
	sourceName, err := queryScreenshotDisplayConfigSourceName(path.SourceInfo.AdapterID, path.SourceInfo.ID)
	if err != nil {
		screenshotHDRCaptureDebugLog("读取 Windows DisplayConfig source name 失败: sourceID=%d err=%v", path.SourceInfo.ID, err)
		return 0, false
	}
	raw, err := queryScreenshotDisplayConfigTargetSDRWhiteLevel(path.TargetInfo.AdapterID, path.TargetInfo.ID)
	if err != nil {
		screenshotHDRCaptureDebugLog("读取 Windows SDR White Level 失败: display=%q source=%q err=%v", displayName, sourceName, err)
		return 0, false
	}
	sdrWhite := float32(raw) / screenshotToneMapThousand
	if !(sdrWhite > 0) {
		screenshotHDRCaptureDebugLog("忽略无效 Windows SDR White Level: display=%q source=%q raw=%d", displayName, sourceName, raw)
		return 0, false
	}
	nits := sdrWhite * screenshotToneMapEighty
	screenshotHDRCaptureDebugLog("读取 Windows SDR White Level: display=%q source=%q raw=%d sdrWhiteLevel=%.3f nits=%.1f",
		displayName, sourceName, raw, sdrWhite, nits)
	return sdrWhite, true
}
