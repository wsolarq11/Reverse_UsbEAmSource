// appicon_windows.go — 图标提取/渲染/缓存 GDI 层（逆向还原）
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)
// 基线：docs/goresym/source_funcs.txt + work/disasm/appicon_asm/*.asm.txt

package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ========================================================================
// Windows API 代理（GDI32 + USER32 + SHELL32 + COMCTL32 延迟加载）
// ========================================================================

var (
	gdi32DLL    = windows.NewLazySystemDLL("gdi32.dll")
	user32DLL   = windows.NewLazySystemDLL("user32.dll")
	shell32DLL  = windows.NewLazySystemDLL("shell32.dll")
	comctl32DLL = windows.NewLazySystemDLL("comctl32.dll")

	procCreateCompatibleDC = gdi32DLL.NewProc("CreateCompatibleDC")
	procDeleteDC           = gdi32DLL.NewProc("DeleteDC")
	procDeleteObject       = gdi32DLL.NewProc("DeleteObject")
	procSelectObject       = gdi32DLL.NewProc("SelectObject")
	procCreateDIBSection   = gdi32DLL.NewProc("CreateDIBSection")
	procGetObjectW         = gdi32DLL.NewProc("GetObjectW")
	procGetIconInfo        = user32DLL.NewProc("GetIconInfo")
	procDrawIconEx         = user32DLL.NewProc("DrawIconEx")
	procDestroyIcon        = user32DLL.NewProc("DestroyIcon")

	procPrivateExtractIcons = shell32DLL.NewProc("PrivateExtractIconsW")
	procSHDefExtractIcon    = shell32DLL.NewProc("SHDefExtractIconW")
	procSHGetFileInfoW      = shell32DLL.NewProc("SHGetFileInfoW")
	procSHGetImageList      = shell32DLL.NewProc("SHGetImageList")

	procImageListGetIcon = comctl32DLL.NewProc("ImageList_GetIcon")
)

// iidImageList 是 IID_IImageList {46EB5926-582E-4017-9FDF-E8998DAA0950}。
// [S] resolve_lea_strings 解出的 SHGetImageList 第二参 riid。
var iidImageList = &windows.GUID{
	Data1: 0x46eb5926,
	Data2: 0x582e,
	Data3: 0x4017,
	Data4: [8]byte{0x9f, 0xdf, 0xe8, 0x99, 0x8d, 0xaa, 0x09, 0x50},
}

// appIconDataCache 是共享图标 dataURL 缓存（map[string]string）。
// [S] resolveAppIconDataWithOptions 0x1407497a0 与 resolveSystemIconDataWithOptions
// 0x14074aced 的 Load/Swap 目标同为 0x141C12740（单一全局 HashTrieMap）。
var appIconDataCache sync.Map

// defaultAppIconSizes 是提取候选尺寸列表。
// [S] extractAppResourceIconDataWithIndexAndOptions 0x14074a87e/0x14074a8f7 的
// 全局 rodata 列表 [0x100,0x80,0x40,0x30,0x20]（256/128/64/48/32）。
var defaultAppIconSizes = [...]uintptr{0x100, 0x80, 0x40, 0x30, 0x20}

// AppIconOptions 是图标解析选项。ABI 大小固定 64B：
//
//	qword0(0x00) IconIndex；qword1-2(0x08) Namespace；qword3(0x18) Size；
//	qword4(0x20) ImageList(iImageList)；qword5-6(0x28) 候选列表 ptr/len；
//	qword7(0x38) 保留。
//
// [S] resolveAppIconDataWithOptions 0x1407496bd 栈展开 8 qword；extract 族序言
// 偏移反推。qword7 未被 asm 读取，保留为不透明填充。
type AppIconOptions struct {
	IconIndex int32
	_         uint32
	Namespace string
	Size      int
	ImageList int
	CandsPtr  unsafe.Pointer
	CandsLen  int
	CandsCap  int
}

// winBitmap 对应 GDI BITMAP 结构（32B，GetObjectW cbBuffer=0x20）。
type winBitmap struct {
	bmType       int32
	bmWidth      int32
	bmHeight     int32
	bmWidthBytes int32
	bmPlanes     uint16
	bmBitsPixel  uint16
	bmBits       uintptr
}

// winBitmapInfoHeader 对应 BITMAPINFOHEADER（biSize=0x28=40）。
type winBitmapInfoHeader struct {
	biSize          uint32
	biWidth         int32
	biHeight        int32
	biPlanes        uint16
	biBitCount      uint16
	biCompression   uint32
	biSizeImage     uint32
	biXPelsPerMeter int32
	biYPelsPerMeter int32
	biClrUsed       uint32
	biClrImportant  uint32
}

type winBitmapInfo struct {
	bmiHeader winBitmapInfoHeader
	bmiColors [1]uint32
}

// ========================================================================
// GDI 薄包装函数
// ========================================================================

// createAppCompatibleDC 创建与屏幕兼容的 GDI 设备上下文。
// [S] ASM 0x14074c720: LazyProc(0 args: nil → CreateCompatibleDC) → HDC。
func createAppCompatibleDC() (uintptr, error) {
	r, _, _ := procCreateCompatibleDC.Call(0)
	if r == 0 {
		return 0, fmt.Errorf("CreateCompatibleDC failed")
	}
	return r, nil
}

// deleteAppDC 释放 GDI 设备上下文（nil 安全）。
// [S] ASM 0x14074c7a0: test rax → nil → ret; 否则 LazyProc(1 arg → DeleteDC)。
func deleteAppDC(hdc uintptr) {
	if hdc == 0 {
		return
	}
	procDeleteDC.Call(hdc)
}

// deleteAppObject 释放 GDI 对象（nil 安全）。
// [S] ASM 0x14074ca20: 与 deleteAppDC 结构相同。
func deleteAppObject(hgdiobj uintptr) {
	if hgdiobj == 0 {
		return
	}
	procDeleteObject.Call(hgdiobj)
}

// selectAppObject 将 GDI 对象选入 DC，返回原对象。失败返回 0+error。
// [S] ASM 0x14074caa0: LazyProc(2 arg → SelectObject); 返回值 = 0 或 -1 时
// error 串为 "SelectObject failed"（无 %d 格式化，已复核 rodata）。
func selectAppObject(hdc, hgdiobj uintptr) (uintptr, error) {
	r, _, _ := procSelectObject.Call(hdc, hgdiobj)
	if r == 0 || r == ^uintptr(0) {
		return 0, fmt.Errorf("SelectObject failed")
	}
	return r, nil
}

// restoreAppObject 恢复 DC 的 GDI 对象（nil/零安全，不检查返回值）。
// [S] ASM 0x14074cb60: test rax/rbx → nil → ret; 否则 LazyProc(2 arg)。
func restoreAppObject(hdc, oldObj uintptr) {
	if hdc == 0 || oldObj == 0 {
		return
	}
	procSelectObject.Call(hdc, oldObj)
}

// ========================================================================
// COM 释放辅助
// ========================================================================

// releaseComObject 通过 COM vtable Release 释放接口指针。
// [S] ASM 0x14074d820: if ptr==nil || *ptr==nil || (*ptr).Release==nil → ret
// 否则 syscall.Syscall(release, 1, ptr, 0, 0)。
func releaseComObject(ptr unsafe.Pointer) {
	if ptr == nil {
		return
	}
	ptrPtr := *(*unsafe.Pointer)(ptr)
	if ptrPtr == nil {
		return
	}
	vtable := *(*[4]uintptr)(ptrPtr)
	release := vtable[2] // IUnknown.Release = vtable offset 2
	if release == 0 {
		return
	}
	syscall.Syscall(release, 1, uintptr(ptr), 0, 0) //nolint:staticcheck // SA1019
}

// ========================================================================
// 图标信息加载
// ========================================================================

// loadAppIconInfo 调用 GetIconInfo 并计算位图尺寸。
// [S] ASM 0x14074bd80: GetIconInfo(hicon,&info) → appIconBitmapSize(&info)。
// 返回 (info,w,h,err)；GetIconInfo 失败或尺寸失败时释放并返回 nil。
func loadAppIconInfo(hicon uintptr) (*appIconInfo, int, int, error) {
	var info appIconInfo
	r, _, _ := procGetIconInfo.Call(hicon, uintptr(unsafe.Pointer(&info)))
	if r == 0 {
		return nil, 0, 0, fmt.Errorf("GetIconInfo failed")
	}
	w, h, err := appIconBitmapSize(&info)
	if err != nil {
		releaseAppIconInfo(&info)
		return nil, 0, 0, err
	}
	return &info, w, h, nil
}

// releaseAppIconInfo 释放 ICONINFO 内嵌位图并清零。
// [S] ASM 0x14074be80: nil → ret; DeleteObject(HbmMask); DeleteObject(HbmColor);
// 清零两个句柄槽。
func releaseAppIconInfo(info *appIconInfo) {
	if info == nil {
		return
	}
	if info.HbmMask != 0 {
		procDeleteObject.Call(info.HbmMask)
		info.HbmMask = 0
	}
	if info.HbmColor != 0 {
		procDeleteObject.Call(info.HbmColor)
		info.HbmColor = 0
	}
}

// appIconBitmapSize 通过 GetObjectW 读取颜色位图/掩码位图尺寸。
// [S] ASM 0x14074bf60: 优先 HbmColor；失败回退 HbmMask。掩码位图且
// HbmColor==0 时高度减半。返回 (w,h,err)。
func appIconBitmapSize(info *appIconInfo) (int, int, error) {
	if info == nil {
		return 0, 0, fmt.Errorf("icon info is nil")
	}
	var bm winBitmap
	if info.HbmColor != 0 {
		r, _, _ := procGetObjectW.Call(info.HbmColor, uintptr(unsafe.Sizeof(bm)), uintptr(unsafe.Pointer(&bm)))
		if r == 0 {
			return 0, 0, fmt.Errorf("GetObject for color bitmap failed")
		}
		return int(bm.bmWidth), absInt(int(bm.bmHeight)), nil
	}
	if info.HbmMask == 0 {
		return 0, 0, fmt.Errorf("icon has no bitmap")
	}
	r, _, _ := procGetObjectW.Call(info.HbmMask, uintptr(unsafe.Sizeof(bm)), uintptr(unsafe.Pointer(&bm)))
	if r == 0 {
		return 0, 0, fmt.Errorf("GetObject for mask bitmap failed")
	}
	h := absInt(int(bm.bmHeight))
	if info.HbmColor == 0 {
		h /= 2
	}
	return int(bm.bmWidth), h, nil
}

// absInt 返回整数绝对值。[S-inline 内联，语义确定]
func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// ========================================================================
// 图标渲染管线
// ========================================================================

// drawAppIconToRGBA 通过 DrawIconEx 将图标绘入 32bpp DIB 并转 RGBA。
// [S] ASM 0x14074c180: 校验 hicon/w/h → CreateCompatibleDC → createAppIconCanvas
// → selectAppObject → DrawIconEx(9 arg, flags=第4参) → buildRGBAFromDIBBits。
// flags 由调用方指定（3=DI_NORMAL, 1=DI_MASK）。
func drawAppIconToRGBA(hicon uintptr, width, height int, flags uint32) (*image.RGBA, error) {
	if hicon == 0 || width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid icon or size")
	}
	hdc, err := createAppCompatibleDC()
	if err != nil {
		return nil, err
	}
	defer deleteAppDC(hdc)
	hbitmap, pBits, err := createAppIconCanvas(width, height)
	if err != nil {
		return nil, err
	}
	defer deleteAppObject(hbitmap)
	old, err := selectAppObject(hdc, hbitmap)
	if err != nil {
		return nil, err
	}
	defer restoreAppObject(hdc, old)
	r, _, _ := procDrawIconEx.Call(hdc, 0, 0, hicon, uintptr(width), uintptr(height), 0, 0, uintptr(flags))
	if r == 0 {
		return nil, fmt.Errorf("DrawIconEx failed")
	}
	return buildRGBAFromDIBBits(pBits, width, height), nil
}

// createAppIconCanvas 创建 32bpp 自顶向下 DIB Section。
// [S] ASM 0x14074c820: CreateDIBSection(hdc=0, &bmi, DIB_RGB_COLORS, &pBits, 0, 0)。
// 返回 (hbitmap, pBits, err)。无 hdc 参数（asm 中 hdc 槽恒为 0）。
func createAppIconCanvas(width, height int) (uintptr, unsafe.Pointer, error) {
	var bmi winBitmapInfo
	bmi.bmiHeader.biSize = uint32(unsafe.Sizeof(bmi.bmiHeader)) // 0x28
	bmi.bmiHeader.biWidth = int32(width)
	bmi.bmiHeader.biHeight = int32(-height) // 负值 = 自顶向下
	bmi.bmiHeader.biPlanes = 1
	bmi.bmiHeader.biBitCount = 32
	bmi.bmiHeader.biCompression = 0 // BI_RGB
	bmi.bmiHeader.biSizeImage = uint32(width * height * 4)
	var pBits unsafe.Pointer
	hbmp, _, _ := procCreateDIBSection.Call(0, uintptr(unsafe.Pointer(&bmi)), 0, uintptr(unsafe.Pointer(&pBits)), 0, 0)
	if hbmp == 0 || pBits == nil {
		return 0, nil, fmt.Errorf("CreateDIBSection failed")
	}
	return hbmp, pBits, nil
}

// buildRGBAFromDIBBits 将 32bpp BGRA DIB 位转换为 *image.RGBA。
// [S] ASM 0x14074cc00: image.NewRGBA(Rect(0,0,w,h))；pBits/尺寸非法直接返回空图；
// 逐像素 BGRA→RGBA。
func buildRGBAFromDIBBits(pBits unsafe.Pointer, width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	if pBits == nil || width <= 0 || height <= 0 {
		return img
	}
	pixels := unsafe.Slice((*byte)(pBits), width*height*4)
	for y := 0; y < height; y++ {
		row := y * width * 4
		for x := 0; x < width; x++ {
			i := row + x*4
			img.SetRGBA(x, y, color.RGBA{R: pixels[i+2], G: pixels[i+1], B: pixels[i], A: pixels[i+3]})
		}
	}
	return img
}

// applyAppIconMaskAlpha 按单色掩码设置目标 alpha。
// [S] ASM 0x14074cda0: 掩码 RGB 全零＝绘制区域（无可见 alpha 时 A=0xFF）；
// 掩码非零＝非绘制区域（A=0）。语义与旧实现相反。
func applyAppIconMaskAlpha(dst, mask *image.RGBA) {
	if dst == nil || mask == nil {
		return
	}
	r := dst.Bounds().Intersect(mask.Bounds())
	hasAlpha := appIconHasVisibleAlpha(dst)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			c := dst.RGBAAt(x, y)
			m := mask.RGBAAt(x, y)
			if m.R == 0 && m.G == 0 && m.B == 0 {
				if !hasAlpha {
					c.A = 0xFF
				}
			} else {
				c.A = 0
			}
			dst.SetRGBA(x, y, c)
		}
	}
}

// appIconHasVisibleAlpha 判断是否存在 8 位 alpha 非零像素。
// [S] ASM 0x14074cf80: 逐像素 test al（A != 0）即返回 true。
func appIconHasVisibleAlpha(img *image.RGBA) bool {
	if img == nil {
		return false
	}
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.RGBAAt(x, y).A != 0 {
				return true
			}
		}
	}
	return false
}

// normalizeAppIconPNGWithSize 解码 PNG → 裁剪可见区域 → 缩放至正方形 → 重编码。
// [S] ASM 0x14074d040: png.Decode → NewRGBA + DrawMask → visibleIconBounds →
// 裁剪 DrawMask → fitIconToSquare(targetSize<=0?256:targetSize) → png.Encode。
// 返回 (data, ok)。
func normalizeAppIconPNGWithSize(pngData []byte, targetSize int) ([]byte, bool) {
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, false
	}
	b := img.Bounds()
	canvas := image.NewRGBA(b)
	draw.DrawMask(canvas, b, img, b.Min, nil, image.Point{}, draw.Src)

	vb, ok := visibleIconBounds(canvas)
	if !ok {
		return nil, false
	}
	cropped := image.NewRGBA(image.Rect(0, 0, vb.Dx(), vb.Dy()))
	draw.DrawMask(cropped, cropped.Bounds(), canvas, vb.Min, nil, image.Point{}, draw.Src)

	if targetSize <= 0 {
		targetSize = 256
	}
	square := fitIconToSquare(cropped, targetSize)

	var buf bytes.Buffer
	if err := png.Encode(&buf, square); err != nil {
		return nil, false
	}
	return buf.Bytes(), true
}

// visibleIconBounds 计算首个/末个 alpha 非零像素包围盒（max 含 +1）。
// [S] ASM 0x14074d3a0: 逐像素 A!=0 收紧 min/max；未找到返回 (Rect{},false)。
// 无 nil 检查（nil 直接 panic，对齐 asm）。
func visibleIconBounds(img *image.RGBA) (image.Rectangle, bool) {
	b := img.Bounds()
	minX, minY := b.Max.X, b.Max.Y
	maxX, maxY := b.Min.X, b.Min.Y
	found := false
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.RGBAAt(x, y).A != 0 {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
				found = true
			}
		}
	}
	if !found {
		return image.Rectangle{}, false
	}
	return image.Rect(minX, minY, maxX+1, maxY+1), true
}

// fitIconToSquare 最近邻缩放到 size×size 并居中。
// [S] ASM 0x14074d520: scale=size/max(w,h)；newW/newH=int(dim*scale+0.5)，<=0 钳 1；
// offset=(size-new)/2；srcX=int(dstX*w/newW)+Min.X；srcY 同理。
func fitIconToSquare(img *image.RGBA, size int) *image.RGBA {
	if img == nil || size <= 0 {
		return nil
	}
	b := img.Bounds()
	w := b.Dx()
	h := b.Dy()
	if w <= 0 || h <= 0 {
		return nil
	}
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	maxDim := w
	if h > maxDim {
		maxDim = h
	}
	scale := float64(size) / float64(maxDim)
	newW := int(float64(w)*scale + 0.5)
	if newW <= 0 {
		newW = 1
	}
	newH := int(float64(h)*scale + 0.5)
	if newH <= 0 {
		newH = 1
	}
	offsetX := (size - newW) / 2
	offsetY := (size - newH) / 2
	for dstY := 0; dstY < newH; dstY++ {
		srcY := int(float64(dstY)*float64(h)/float64(newH)) + b.Min.Y
		for dstX := 0; dstX < newW; dstX++ {
			srcX := int(float64(dstX)*float64(w)/float64(newW)) + b.Min.X
			dst.SetRGBA(offsetX+dstX, offsetY+dstY, img.RGBAAt(srcX, srcY))
		}
	}
	return dst
}

// ========================================================================
// 缓存键构建
// ========================================================================

// isPrivateExtractIconCandidate 判断路径是否 .dll/.exe/.ico（忽略大小写）。
// [S] ASM 0x14074b200: TrimSpace → 尾向前扫描到 . 或 / \ → ToLower(ext) →
// 比较 ".dll"(0x6c6c642e)/".exe"(0x6578652e)/".ico"(0x6f63692e)。
func isPrivateExtractIconCandidate(path string) bool {
	p := strings.TrimSpace(path)
	ext := ""
	for i := len(p) - 1; i >= 0; i-- {
		c := p[i]
		if c == '\\' || c == '/' {
			break
		}
		if c == '.' {
			ext = p[i:]
			break
		}
	}
	ext = strings.ToLower(ext)
	return ext == ".dll" || ext == ".exe" || ext == ".ico"
}

// normalizeAppIconCacheKey 清理缓存键：TrimSpace → Clean → ToLower。
// [S] ASM 0x140749d00: strings.TrimSpace → filepathlite.Clean → strings.ToLower。
func normalizeAppIconCacheKey(key string) string {
	return strings.ToLower(filepath.Clean(strings.TrimSpace(key)))
}

// normalizeFileSearchIconExtension 规范化扩展名：TrimSpace → ToLower → 去首点。
// [S] ASM 0x140749d60: TrimSpace → ToLower → len>0 && [0]=='.' 则 [1:]。
func normalizeFileSearchIconExtension(ext string) string {
	ext = strings.TrimSpace(ext)
	ext = strings.ToLower(ext)
	if len(ext) > 0 && ext[0] == '.' {
		ext = ext[1:]
	}
	return ext
}

// namespaceAppIconCacheKey 用命名空间前缀拼接缓存键。
// [S] ASM 0x140749c20: TrimSpace(key)==空 → ""；TrimSpace(ns)==空 → 返回原始 key；
// 否则 concat(TrimSpace(ns), "/", 原始 key)。
func namespaceAppIconCacheKey(key, namespace string) string {
	if strings.TrimSpace(key) == "" {
		return ""
	}
	if strings.TrimSpace(namespace) == "" {
		return key
	}
	return strings.TrimSpace(namespace) + "/" + key
}

// buildPathAppIconCacheKey 构建路径缓存键。
// [S] ASM 0x14074a0e0: normalizeAppIconCacheKey → 空则 ""；modtime!=0 →
// fmt.Sprintf("path:%s#%d", key, modtime)；modtime==0 → "path:"+key。
func buildPathAppIconCacheKey(path string, modtime int64) string {
	key := normalizeAppIconCacheKey(path)
	if key == "" {
		return ""
	}
	if modtime != 0 {
		return fmt.Sprintf("path:%s#%d", key, modtime)
	}
	return "path:" + key
}

// buildResourceAppIconCacheKey 构建资源图标缓存键。
// [S] ASM 0x14074a1e0: normalizeAppIconCacheKey → "resource:%s"；a!=0 →
// fmt.Sprintf("%s#%d", ...)；b>0 → fmt.Sprintf("%s@%d", ...)。
func buildResourceAppIconCacheKey(url string, a, b int64) string {
	key := normalizeAppIconCacheKey(url)
	if key == "" {
		return ""
	}
	result := fmt.Sprintf("resource:%s", key)
	if a != 0 {
		result = fmt.Sprintf("%s#%d", result, a)
	}
	if b > 0 {
		result = fmt.Sprintf("%s@%d", result, b)
	}
	return result
}

// resolveAppIconLastWriteUnixNano 返回文件最后修改时间（纳秒）。
// [S] ASM 0x14074a380: os.Stat + ModTime().UnixNano()；错误返回 0。
func resolveAppIconLastWriteUnixNano(path string) int64 {
	st, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return st.ModTime().UnixNano()
}

// ========================================================================
// 入口解析链（全部单 string 返回）
// ========================================================================

// resolveAppIconData 解析路径图标为 dataURL。
// [S] ASM 0x140749480: resolveAppIconDataWithOptions(path, AppIconOptions{})。
func resolveAppIconData(path string) string {
	return resolveAppIconDataWithOptions(path, AppIconOptions{})
}

// resolveDynamicStartMenuAppIconData 解析动态开始菜单图标。
// [S] ASM 0x140749500: classifyAutomaticWindowsPath → "local" 才继续。
func resolveDynamicStartMenuAppIconData(path, _ string) string {
	result, _ := classifyAutomaticWindowsPath(path, nil)
	if result != classLocal {
		return ""
	}
	return resolveAppIconDataWithOptions(path, AppIconOptions{})
}

// resolveDynamicStartMenuAppIconDataWithIndex 同上，带图标索引。
// [S] ASM 0x1407495c0: classifyAutomaticWindowsPath → "local" → indexed 解析。
func resolveDynamicStartMenuAppIconDataWithIndex(path string, index uint32) string {
	result, _ := classifyAutomaticWindowsPath(path, nil)
	if result != classLocal {
		return ""
	}
	return resolveAppIconDataWithOptions(path, AppIconOptions{IconIndex: int32(index)})
}

// resolveAppIconDataWithOptions 带选项的图标解析（含全局缓存）。
// [S] ASM 0x1407496a0: buildAppIconCacheLookup → Load(cacheKey) 命中即返；
// 未命中 extractAppIconDataForLookup → 非空则 Swap。
func resolveAppIconDataWithOptions(path string, opts AppIconOptions) string {
	cacheKey, cleanPath, assocIndex, sysIndex, systemOnly := buildAppIconCacheLookup(path, opts)
	if cacheKey == "" {
		return ""
	}
	if v, ok := appIconDataCache.Load(cacheKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	data := extractAppIconDataForLookup(cacheKey, cleanPath, assocIndex, sysIndex, systemOnly, opts)
	if data != "" {
		appIconDataCache.Store(cacheKey, data)
	}
	return data
}

// buildAppIconCacheLookup 根据路径与选项计算缓存键及提取参数。
// [S] ASM 0x140749900: TrimSpace+Clean → "."/空 返回全零；iconIndex!=0 或
// private 候选 → resolveAppIconLastWriteUnixNano → buildResourceAppIconCacheKey
// → namespace；否则 getSystemIconIndexWithAttributes(path,0,false)，错误回退
// buildPathAppIconCacheKey，成功且 idx>=0 → fmt.Sprintf("%06d",idx) → namespace，
// idx<0 → 空键。返回 (cacheKey, cleanPath, assocIndex, sysIndex, systemOnly)。
func buildAppIconCacheLookup(path string, opts AppIconOptions) (cacheKey, cleanPath string, assocIndex, sysIndex int, systemOnly bool) {
	key := strings.TrimSpace(path)
	key = filepath.Clean(key)
	if key == "." || key == "" {
		return "", "", 0, 0, false
	}

	if opts.IconIndex != 0 || isPrivateExtractIconCandidate(key) {
		modtime := resolveAppIconLastWriteUnixNano(key)
		rk := buildResourceAppIconCacheKey(key, int64(opts.IconIndex), modtime)
		return namespaceAppIconCacheKey(rk, opts.Namespace), key, int(opts.IconIndex), 0, false
	}

	idx, err := getSystemIconIndexWithAttributes(key, 0, false)
	if err != nil {
		modtime := resolveAppIconLastWriteUnixNano(key)
		pk := buildPathAppIconCacheKey(key, modtime)
		return namespaceAppIconCacheKey(pk, opts.Namespace), key, 0, 0, false
	}
	if idx >= 0 {
		sysKey := fmt.Sprintf("%06d", idx)
		return namespaceAppIconCacheKey(sysKey, opts.Namespace), key, 0, idx, true
	}
	return namespaceAppIconCacheKey("", opts.Namespace), key, 0, idx, true
}

// getSystemIconIndexWithAttributes 通过 SHGetFileInfoW 取系统图标索引。
// [S] ASM 0x14074b580: 5 arg 数组 SHGetFileInfoW；useFileAttributes 决定
// flagsCalc/indexFlags（0x4010:0x4000，0:0x80 回退）；失败重试一次（仅
// !useFileAttributes）；sfi[+8] int32 <0 → "invalid icon index"。
func getSystemIconIndexWithAttributes(path string, flags uint32, useFileAttributes bool) (int, error) {
	var sfi [0x2b8]byte
	utf16Path, err := windows.UTF16PtrFromString(path)
	if err != nil || utf16Path == nil {
		return 0, fmt.Errorf("invalid path")
	}
	flagsCalc := flags
	indexFlags := uint32(0x4000)
	if useFileAttributes {
		if flagsCalc == 0 {
			flagsCalc = 0x80
		}
		indexFlags = 0x4010
	}
	r, _, _ := procSHGetFileInfoW.Call(
		uintptr(unsafe.Pointer(utf16Path)),
		uintptr(flagsCalc),
		uintptr(unsafe.Pointer(&sfi[0])),
		uintptr(len(sfi)),
		uintptr(indexFlags),
	)
	if r == 0 && !useFileAttributes {
		// 重试：按文件属性回退。
		flagsCalc = 0x80
		indexFlags = 0x4010
		r, _, _ = procSHGetFileInfoW.Call(
			uintptr(unsafe.Pointer(utf16Path)),
			uintptr(flagsCalc),
			uintptr(unsafe.Pointer(&sfi[0])),
			uintptr(len(sfi)),
			uintptr(indexFlags),
		)
	}
	if r == 0 {
		return 0, fmt.Errorf("SHGetFileInfoW failed")
	}
	iImage := *(*int32)(unsafe.Pointer(&sfi[8]))
	if iImage < 0 {
		return 0, fmt.Errorf("invalid icon index")
	}
	return int(iImage), nil
}

// ========================================================================
// 图标句柄加载
// ========================================================================

// loadPrivateExtractedAppIcon 通过 PrivateExtractIconsW 提取图标。
// [S] ASM 0x14074b2c0: 8 arg 数组 PrivateExtractIconsW(path,index,cx=size,
// cy=size,&hicon,pid=0,nIcons=1,flags=0)；0/0xffffffff 失败。
func loadPrivateExtractedAppIcon(path string, index uint32, size uint32) (uintptr, error) {
	utf16Path, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	var hicon uintptr
	r, _, _ := procPrivateExtractIcons.Call(
		uintptr(unsafe.Pointer(utf16Path)),
		uintptr(int32(index)),
		uintptr(size),
		uintptr(size),
		uintptr(unsafe.Pointer(&hicon)),
		0,
		1,
		0,
	)
	if r == 0 || r == 0xffffffff {
		return 0, fmt.Errorf("PrivateExtractIconsW failed")
	}
	return hicon, nil
}

// loadShellDefinedAppIcon 通过 SHDefExtractIconW 提取图标。
// [S] ASM 0x14074b420: 6 arg 数组 SHDefExtractIconW(path,index,uFlags=0,
// &hicon,phiconSmall=0,nIconSize=flags)。
func loadShellDefinedAppIcon(path string, index uint32, flags uint16) (uintptr, error) {
	utf16Path, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	var hicon uintptr
	r, _, _ := procSHDefExtractIcon.Call(
		uintptr(unsafe.Pointer(utf16Path)),
		uintptr(int32(index)),
		0,
		uintptr(unsafe.Pointer(&hicon)),
		0,
		uintptr(flags),
	)
	if r != 0 || hicon == 0 {
		return 0, fmt.Errorf("SHDefExtractIconW failed")
	}
	return hicon, nil
}

// loadImageListIcon 从系统图像列表取图标句柄。
// [S] ASM 0x14074b7a0: ImageList_GetIcon(imagelist, idx, 0)（3 arg）。
func loadImageListIcon(imagelist uintptr, iconIndex int) (uintptr, error) {
	r, _, _ := procImageListGetIcon.Call(imagelist, uintptr(iconIndex), 0)
	if r == 0 {
		return 0, fmt.Errorf("ImageList_GetIcon failed")
	}
	return r, nil
}

// loadAssociatedAppIcon 通过 SHGetFileInfoW 取关联图标句柄。
// [S] ASM 0x14074b860: SHGetFileInfoW(path, dwFileAttributes=(flags&0x10?0x80:0),
// &sfi, 0x2b8, uFlags=flags)；返回 sfi.HIcon。
func loadAssociatedAppIcon(path string, flags uint32) (uintptr, error) {
	utf16Path, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	var sfi [0x2b8]byte
	dwAttr := uint32(0)
	if flags&0x10 != 0 {
		dwAttr = 0x80
	}
	r, _, _ := procSHGetFileInfoW.Call(
		uintptr(unsafe.Pointer(utf16Path)),
		uintptr(dwAttr),
		uintptr(unsafe.Pointer(&sfi[0])),
		uintptr(len(sfi)),
		uintptr(flags),
	)
	if r == 0 {
		return 0, fmt.Errorf("SHGetFileInfoW failed")
	}
	return *(*uintptr)(unsafe.Pointer(&sfi[0])), nil
}

// ========================================================================
// 图标句柄 → PNG/dataURL
// ========================================================================

// renderAppIconHandle 渲染图标为彩色 + 掩码 RGBA。
// [S] ASM 0x14074bb60: loadAppIconInfo → drawAppIconToRGBA(hicon,w,h,3)；
// HbmMask!=0 时 drawAppIconToRGBA(hicon,w,h,1)。
func renderAppIconHandle(hicon uintptr) (*image.RGBA, *image.RGBA, error) {
	info, w, h, err := loadAppIconInfo(hicon)
	if err != nil {
		return nil, nil, err
	}
	defer releaseAppIconInfo(info)
	img, err := drawAppIconToRGBA(hicon, w, h, 3)
	if err != nil {
		return nil, nil, err
	}
	var mask *image.RGBA
	if info.HbmMask != 0 {
		mask, err = drawAppIconToRGBA(hicon, w, h, 1)
		if err != nil {
			return nil, nil, err
		}
	}
	return img, mask, nil
}

// encodeAppIconHandleToPNG 将图标句柄编码为 PNG。
// [S] ASM 0x14074ba60: renderAppIconHandle → mask!=nil 则 applyAppIconMaskAlpha
// → png.Encoder.Encode(&buf, img) → buf.Bytes()。
func encodeAppIconHandleToPNG(hicon uintptr) ([]byte, error) {
	img, mask, err := renderAppIconHandle(hicon)
	if err != nil {
		return nil, err
	}
	if mask != nil {
		applyAppIconMaskAlpha(img, mask)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// saveAppIconHandleAsDataURLWithSize 将图标句柄保存为 base64 dataURL。
// [S] ASM 0x14074b980: encodeAppIconHandleToPNG；失败 ""；size>0 且归一化成功
// 则用归一化数据；拼 "data:image/png;base64," 前缀。
func saveAppIconHandleAsDataURLWithSize(hicon uintptr, size int) string {
	pngData, err := encodeAppIconHandleToPNG(hicon)
	if err != nil {
		return ""
	}
	if size > 0 {
		if normalized, ok := normalizeAppIconPNGWithSize(pngData, size); ok {
			pngData = normalized
		}
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngData)
}

// ========================================================================
// 提取调度链（全部单 string 返回）
// ========================================================================

// extractAppResourceIconDataWithIndexAndOptions 从 PE 资源提取图标。
// [S] ASM 0x14074a800: iconIndex!=0 → shell 候选循环（loadShellDefinedAppIcon，
// flags=元素低 16 位）；然后 isPrivateExtractIconCandidate → private 候选循环
// （loadPrivateExtractedAppIcon，size=元素低 32 位）；CandsLen==0 用全局尺寸表。
// 每次成功 save → DestroyIcon；非空即返。
func extractAppResourceIconDataWithIndexAndOptions(path string, iconIndex int32, opts AppIconOptions) string {
	var cands []uintptr
	if opts.CandsLen == 0 {
		cands = defaultAppIconSizes[:]
	} else {
		cands = unsafe.Slice((*uintptr)(opts.CandsPtr), opts.CandsLen)
	}

	if iconIndex != 0 {
		for _, v := range cands {
			hicon, err := loadShellDefinedAppIcon(path, uint32(iconIndex), uint16(v))
			if err != nil || hicon == 0 {
				continue
			}
			data := saveAppIconHandleAsDataURLWithSize(hicon, opts.Size)
			procDestroyIcon.Call(hicon)
			if data != "" {
				return data
			}
		}
	}

	if !isPrivateExtractIconCandidate(path) {
		return ""
	}
	for _, v := range cands {
		hicon, err := loadPrivateExtractedAppIcon(path, uint32(iconIndex), uint32(v))
		if err != nil || hicon == 0 {
			continue
		}
		data := saveAppIconHandleAsDataURLWithSize(hicon, opts.Size)
		procDestroyIcon.Call(hicon)
		if data != "" {
			return data
		}
	}
	return ""
}

// extractSystemIconDataByIndexWithOptions 按系统图标索引提取。
// [S] ASM 0x14074ae40: iconIndex<0 → ""；SHGetImageList(iImageList=ImageList
// 或 0→4, riid, &imagelist)；失败 ""；loadImageListIcon → saveAppIconHandleAsDataURLWithSize(Size)
// → DestroyIcon → Release(imagelist)。
func extractSystemIconDataByIndexWithOptions(iconIndex int, opts AppIconOptions) string {
	if iconIndex < 0 {
		return ""
	}
	var imagelist uintptr
	iImageList := opts.ImageList
	if iImageList == 0 {
		iImageList = 4 // SHIL_JUMBO
	}
	hr, _, _ := procSHGetImageList.Call(
		uintptr(iImageList),
		uintptr(unsafe.Pointer(iidImageList)),
		uintptr(unsafe.Pointer(&imagelist)),
	)
	if hr != 0 || imagelist == 0 {
		return ""
	}
	defer releaseComObject(unsafe.Pointer(&imagelist))

	hicon, err := loadImageListIcon(imagelist, iconIndex)
	if err != nil || hicon == 0 {
		return ""
	}
	defer procDestroyIcon.Call(hicon)

	return saveAppIconHandleAsDataURLWithSize(hicon, opts.Size)
}

// extractSystemIconDataForPathWithOptions 按路径取系统图标索引后提取。
// [S] ASM 0x14074ab00: getSystemIconIndexWithAttributes(path,0,false) →
// extractSystemIconDataByIndexWithOptions。
func extractSystemIconDataForPathWithOptions(path string, opts AppIconOptions) string {
	idx, err := getSystemIconIndexWithAttributes(path, 0, false)
	if err != nil {
		return ""
	}
	return extractSystemIconDataByIndexWithOptions(idx, opts)
}

// resolveSystemIconDataWithOptions 带缓存的系统图标解析。
// [S] ASM 0x14074ac00: idx<0 → ""；key=fmt.Sprintf("sys:%d",idx) →
// namespaceAppIconCacheKey → Load 命中即返；miss extractSystemIconDataByIndexWithOptions
// → 非空且 key 非空则 Swap。
func resolveSystemIconDataWithOptions(iconIndex int32, opts AppIconOptions) string {
	if iconIndex < 0 {
		return ""
	}
	key := fmt.Sprintf("sys:%d", iconIndex)
	cacheKey := namespaceAppIconCacheKey(key, opts.Namespace)
	if v, ok := appIconDataCache.Load(cacheKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	data := extractSystemIconDataByIndexWithOptions(int(iconIndex), opts)
	if data != "" && cacheKey != "" {
		appIconDataCache.Store(cacheKey, data)
	}
	return data
}

// extractAssociatedAppIconDataWithIndexAndOptions 提取关联图标（资源→系统→关联）。
// [S] ASM 0x14074a560: extractAppResource... 非空即返；iconIndex!=0 则 ""；
// extractSystemIconDataForPathWithOptions 非空即返；loadAssociatedAppIcon(path,0x100)
// 失败回退 0x110；save(Size) → DestroyIcon。
func extractAssociatedAppIconDataWithIndexAndOptions(path string, iconIndex int32, opts AppIconOptions) string {
	if data := extractAppResourceIconDataWithIndexAndOptions(path, iconIndex, opts); data != "" {
		return data
	}
	if iconIndex != 0 {
		return ""
	}
	if data := extractSystemIconDataForPathWithOptions(path, opts); data != "" {
		return data
	}
	hicon, err := loadAssociatedAppIcon(path, 0x100)
	if err != nil || hicon == 0 {
		hicon, err = loadAssociatedAppIcon(path, 0x110)
	}
	if err != nil || hicon == 0 {
		return ""
	}
	data := saveAppIconHandleAsDataURLWithSize(hicon, opts.Size)
	procDestroyIcon.Call(hicon)
	return data
}

// extractAppIconDataForLookup 提取调度：系统优先 → 关联回退。
// [S] ASM 0x14074a420: systemOnly && sysIndex>=0 → extractSystemIconDataByIndexWithOptions
// 非空即返；否则无条件 extractAssociatedAppIconDataWithIndexAndOptions。
// 首参 cacheKey 在 asm 中未被读取（保留为 ABI 占位）。
func extractAppIconDataForLookup(_ string, cleanPath string, assocIndex, sysIndex int, systemOnly bool, opts AppIconOptions) string {
	if systemOnly && sysIndex >= 0 {
		if data := extractSystemIconDataByIndexWithOptions(sysIndex, opts); data != "" {
			return data
		}
	}
	return extractAssociatedAppIconDataWithIndexAndOptions(cleanPath, int32(assocIndex), opts)
}

// ========================================================================
// 文件搜索图标解析
// ========================================================================

// resolveFileSearchIconDataWithHints 带提示的文件搜索图标解析。
// [S] ASM 0x140749dc0: trimmed=TrimSpace(path)；ext=normalizeFileSearchIconExtension(hints)；
// hints 空且 trimmed 非空则尾扫描 trimmed 找 ext；isFolder → resolveFileSearchTypeIconData("",true)；
// 否则 ext 再 normalize 后匹配 dll/exe/ico/lnk/url/appref-ms → resolveAppIconDataWithOptions(trimmed)，
// 未匹配 → resolveFileSearchTypeIconData(ext,false)。
func resolveFileSearchIconDataWithHints(path string, hints string, isFolder bool) string {
	trimmed := strings.TrimSpace(path)
	ext := normalizeFileSearchIconExtension(hints)
	if ext == "" && trimmed != "" {
		// 尾扫描兜底：取最后一个 '.' 之后的部分（遇 / \ 中止）。
		for i := len(trimmed) - 1; i >= 0; i-- {
			c := trimmed[i]
			if c == '\\' || c == '/' {
				break
			}
			if c == '.' {
				ext = trimmed[i:]
				break
			}
		}
		ext = normalizeFileSearchIconExtension(ext)
	}
	if isFolder {
		return resolveFileSearchTypeIconData("", true)
	}
	norm := normalizeFileSearchIconExtension(ext)
	switch norm {
	case "dll", "exe", "ico", "lnk", "url", "appref-ms":
		if trimmed == "" {
			return ""
		}
		return resolveAppIconDataWithOptions(trimmed, AppIconOptions{})
	}
	return resolveFileSearchTypeIconData(ext, false)
}

// resolveFileSearchTypeIconData 通过扩展名获取类型图标 dataURL。
// [S] ASM 0x140749fe0: folder→("folder",0x10)；否则 normalize 扩展名，空→("file",0x80)，
// 非空→("placeholder."+ext,0x80)；getSystemIconIndexWithAttributes(...,true)；
// 错误吞掉返回 ""；成功 resolveSystemIconDataWithOptions(idx, 全零 opts)。
func resolveFileSearchTypeIconData(ext string, isFolder bool) string {
	var placeholder string
	var flags uint32
	if isFolder {
		placeholder = "folder"
		flags = 0x10
	} else {
		ne := normalizeFileSearchIconExtension(ext)
		if ne == "" {
			placeholder = "file"
		} else {
			placeholder = "placeholder." + ne
		}
		flags = 0x80
	}
	idx, err := getSystemIconIndexWithAttributes(placeholder, flags, true)
	if err != nil {
		return ""
	}
	return resolveSystemIconDataWithOptions(int32(idx), AppIconOptions{})
}
