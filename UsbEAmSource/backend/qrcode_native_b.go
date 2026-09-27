package main

// 本文件承载 UsbEAm Launcher 1.0.3 (Go 1.25.12, PE64, module main) 中
// QRCode 截图标注/选区渲染域的 59 个函数（含 1 个同文件 helper）的反汇编
// 翻译。每个函数上方紧邻 [S]（体完整翻译）或 [S-sig]（仅签名已证实、体返回
// 零值）tier 标记。签名仅由 asm ABI 与既有类型推导，未臆造函数体。

import (
	"fmt"
	"image"
	"math"
)

// qrCodeSelectionResizeHandle 是选区 8 个缩放手柄的坐标/序号槽。
// [S] ASM 派生：{id byte; x,y int} 布局为 24 字节（id 后 7 字节对齐）。
type qrCodeSelectionResizeHandle struct {
	id   byte
	x, y int
}

// qrCodePaintBuffer 是 createQRCodePaintBuffer 返回的 40 字节画笔缓冲结构。
// [S-sig] ASM 派生：5 个 qword 字段语义（dc/canvas/回退句柄/标志）待确认。
type qrCodePaintBuffer struct {
	dc     uintptr
	canvas uintptr
	a, b   uintptr
	ok     bool
}

// [S 0x1409529e0]
// 计算一条线段标注的绘制包围盒：p0/p1 的包围盒向外扩 max(8, lineWidth+10)。
func qrCodeAnnotationLinePaintBounds(p0, p1 image.Point, lineWidth int) image.Rectangle {
	b := qrCodeAnnotationPointsBounds([]image.Point{p0, p1})
	pad := lineWidth + 10
	if pad < 8 {
		pad = 8
	}
	return b.Inset(-pad)
}

// [S-sig 0x140952ac0]
// GDI FillRect：以实心画刷填充矩形。签名由 morestack 保存
// (rax,rbx,rcx,rdi,rsi,r8b,r9b,r10b,r11b) 证实为 9 参数；体为 GDI 调用链。
func fillQRCodeRect(hdc uintptr, x0, y0, x1, y1 int, r, g, b, a byte) {}

// [S-sig 0x140952d20]
// GDI RoundRect。签名由 morestack (rax..r8 共 6 qword + 栈上 4 byte RGBA)
// 证实；radius 语义与栈字节 RGBA 顺序经 body 起始保存确认。
func drawQRCodeRoundedRect(hdc uintptr, x0, y0, x1, y1, radius int, r, g, b, a byte) {}

// [S-sig 0x140953220]
// GDI 滑块拇指绘制。签名由 morestack (rax,rbx,rcx,rdi,sil,r8b,r9b,r10b)
// 证实为 4 qword + 4 byte；x,y,radius 语义经 body 起始保存推断。
func drawQRCodeSliderThumb(hdc uintptr, x, y, radius int, r, g, b, a byte) {}

// [S-sig 0x140953740]
// GDI 标注完成遮罩。签名由 morestack (rax..r11 共 9 qword) 证实；
// 两个 image.Rectangle 的参数分组由 body 起始 9 个 qword 保存推断。
func drawQRCodeAnnotationFinishOverlay(hdc uintptr, r0, r1 image.Rectangle) {}

// [S-sig 0x140953d20]
// GDI 矩形边框。签名由 morestack (rax,rbx,rcx,rdi,rsi,r8b,r9b,r10b,r11b)
// 证实为 9 参数；体为 CreatePen/SelectObject/矩形描边调用链。
func drawQRCodeRectFrame(hdc uintptr, x0, y0, x1, y1 int, r, g, b, a byte) {}

// [S 0x140953fe0]
// 在 (x0,y0)-(x1,y1) 内居中放置 w×h 矩形，返回左上角。负数 w/h 时按
// asm 的 min/max 交换语义回退到更小坐标。
func centerRect(x0, y0, x1, y1, w, h int) (int, int) {
	left := x0 + (x1-x0-w)/2
	top := y0 + (y1-y0-h)/2
	if left+w < left {
		left = left + w
	}
	if top+h < top {
		top = top + h
	}
	return left, top
}

// [S-sig 0x140954040]
// GDI 带描边文字。签名由 morestack (rax,rbx,rcx;edi,esi,r8d,r9d;r10;r11d)
// 证实为 3 qword + 4 int + 1 qword + 1 int；text/rect 分组经 body 的
// newobject 与 4-int 打包推断，后两参数语义待确认。
func drawQRCodeAnnotationOutlinedText(img *image.RGBA, text string, rect image.Rectangle, a, b int) {}

// [S 0x140954480]
// 数字标注尺寸：digits>2 → int(size*0.58) 最小 12；==2 → *0.68 最小 11；
// <=1 → *0.82 最小 10。
func qrCodeAnnotationNumberLabelSize(size, digits int) int {
	f := float64(size)
	var v int
	switch {
	case digits > 2:
		v = int(f * 0.58)
		if v < 12 {
			v = 12
		}
	case digits == 2:
		v = int(f * 0.68)
		if v < 11 {
			v = 11
		}
	default:
		v = int(f * 0.82)
		if v < 10 {
			v = 10
		}
	}
	return v
}

// [S 0x1409546c0]
// 数字标注描边宽度：int(size*0.08) 钳制到 [1,3]。
func qrCodeAnnotationNumberOutlineWidth(size int) int {
	w := int(float64(size) * 0.08)
	if w < 1 {
		w = 1
	}
	if w > 3 {
		w = 3
	}
	return w
}

// [S 0x1409547a0]
// 生成文字描边偏移集合：n 至少为 1；遍历 [-n,n]^2 内 dx^2+dy^2<=n^2+1 的
// 点，跳过 (0,0)。
func qrCodeAnnotationTextOutlineOffsets(n int) []image.Point {
	if n < 1 {
		n = 1
	}
	limit := n*n + 1
	out := make([]image.Point, 0, (2*n+1)*(2*n+1))
	for dy := -n; dy <= n; dy++ {
		for dx := -n; dx <= n; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			if dx*dx+dy*dy <= limit {
				out = append(out, image.Pt(dx, dy))
			}
		}
	}
	return out
}

// [S-sig 0x140954c20]
// 复杂标注描边绘制（含平滑胶囊内部闭包）。签名由 morestack 8 qword 证实；
// 7 个 int/uintptr 参数语义待确认。
func drawQRCodeAnnotationStrokeOnImageWithOffset(img *image.RGBA, a, b, c, d, e, f, g int) {}

// [S-sig 0x140955500]
// 标注马赛克绘制。签名由 morestack 8 qword 证实；参数语义待确认。
func drawQRCodeAnnotationMosaicOnImageWithOffset(img *image.RGBA, a, b, c, d, e, f, g int) {}

// [S-sig 0x1409557c0]
// 标注模糊绘制。签名由 morestack 8 qword 证实；参数语义待确认。
func drawQRCodeAnnotationBlurOnImageWithOffset(img *image.RGBA, a, b, c, d, e, f, g int) {}

// [S 0x140955a60]
// 画箭头头部：对 ArrowHeadPoints 的每个点，从 p1 到该点画平滑胶囊。
func drawArrowHeadOnRGBA(img *image.RGBA, p0, p1 image.Point, r, g, b, a byte, lineWidth int) {
	for _, p := range qrCodeAnnotationArrowHeadPoints(p0, p1, lineWidth) {
		drawSmoothCapsuleOnRGBA(img, p1.X, p1.Y, p.X, p.Y, r, g, b, a, lineWidth)
	}
}

// [S 0x140955ba0]
// 计算箭头头部两点：angle=atan2(dy,dx)，len=max(14, 4*max(1,lineWidth))，
// 两点 = p1 − (cos/sin(angle ± 0.5236216637818927) * len)。
func qrCodeAnnotationArrowHeadPoints(p0, p1 image.Point, lineWidth int) []image.Point {
	dy := float64(p1.Y - p0.Y)
	dx := float64(p1.X - p0.X)
	angle := math.Atan2(dy, dx)
	lw := lineWidth
	if lw < 1 {
		lw = 1
	}
	l := float64(14)
	if v := float64(4 * lw); v > l {
		l = v
	}
	const spread = 0.5236216637818927 // ~30°
	p := make([]image.Point, 2)
	p[0] = image.Pt(
		p1.X-int(math.Cos(angle+spread)*l),
		p1.Y-int(math.Sin(angle+spread)*l),
	)
	p[1] = image.Pt(
		p1.X-int(math.Cos(angle-spread)*l),
		p1.Y-int(math.Sin(angle-spread)*l),
	)
	return p
}

// [S 0x140955f00]
// 椭圆描边绘制：center=((x0+x1)/2,(y0+y1)/2)，rx=(x1-x0)/2，ry=(y1-y0)/2。
// dist²<=r_inner² 完全混合，>r_outer² 跳过，否则按覆盖度混合。
func drawEllipseOnRGBA(img *image.RGBA, x0, y0, x1, y1 int, r, g, b, a byte, lineWidth int) {
	if img == nil {
		return
	}
	cx := float64(x0+x1) / 2
	cy := float64(y0+y1) / 2
	rx := float64(x1-x0) / 2
	ry := float64(y1-y0) / 2
	if rx <= 0 || ry <= 0 {
		return
	}
	hw := float64(lineWidth) / 2
	rInner := hw - 0.5
	if rInner < 0 {
		rInner = 0
	}
	rOuter := hw + 0.5
	for py := y0; py <= y1; py++ {
		for px := x0; px <= x1; px++ {
			d := qrCodeEllipseStrokeDistance(float64(px)+0.5, float64(py)+0.5, cx, cy, rx, ry)
			switch {
			case d <= rInner:
				blendQRCodeRGBAPixel(img, px, py, r, g, b, a)
			case d <= rOuter:
				cov := byte(qrCodeEllipseStrokePixelCoverage(px, py, cx, cy, rx, ry, hw) * 255)
				if cov != 0 {
					blendQRCodeRGBAPixel(img, px, py, r, g, b, byte((int(a)*int(cov)+127)/255))
				}
			}
		}
	}
}

// [S-sig 0x140956360]
// 数字标注绘制。morestack 仅保存 rax(img)，其余大量栈参数（数值/坐标/颜色/
// 字体）未在 morestack 暴露；完整参数列表待确认。
func drawNumberOnRGBA(img *image.RGBA) {}

// [S-sig 0x140956820]
// 带描边文本块绘制。签名由 morestack (rax,bl,cl,dil,sil,r8) 证实为
// img + 4 byte(RGBA) + 1 int(fontSize)；栈上的 text/布局参数待确认。
func drawTextBlockOnRGBAWithOutline(img *image.RGBA, r, g, b, a byte, fontSize int) {}

// [S-sig 0x140956ce0]
// 文字掩膜区域绘制。签名由 morestack (rax,rbx,rcx,rdi,rsi,r8b,r9b,r10b,r11b)
// 证实为 5 qword + 4 byte；5 个 int/uintptr 语义待确认。
func drawTextMaskOnRGBARegion(img *image.RGBA, a, b, c, d int, r, g, b_, a_ byte) {}

// [S-sig 0x140957b40]
// 文字掩膜 DIB 位拷贝到 RGBA 区域。签名由 morestack 7 qword 证实；
// 7 个句柄/指针语义待确认。
func copyQRCodeTextMaskDIBBitsToRGBARegion(a, b, c, d, e, f, g uintptr) {}

// [S-sig 0x140957f00]
// 平滑胶囊绘制。签名由 morestack (rax,rbx,rcx,rdi,rsi,r8b,r9b,r10b,r11b +
// 栈 [0x1a8] lineWidth) 证实；内部含胶囊 SDF 闭包与退化圆回退。
func drawSmoothCapsuleOnRGBA(img *image.RGBA, x0, y0, x1, y1 int, r, g, b, a byte, lineWidth int) {}

// [S 0x140958da0]
// 胶囊像素覆盖度（4×4 超采样）：capsule SDF t=clamp(((x-x0)*dx+(y-y0)*dy)/len2,0,1)，
// 采样点到胶囊轴线距离平方 <= r2 计数，结果 ×(1/16)。
func qrCodeCapsulePixelCoverage(px, py int, x0, y0, dx, dy, len2, r2 float64) float64 {
	count := 0
	for row := 0; row < 4; row++ {
		sy := float64(py) + (float64(row)+0.5)*0.25 - 0.5
		for col := 0; col < 4; col++ {
			sx := float64(px) + (float64(col)+0.5)*0.25 - 0.5
			t := ((sx-x0)*dx + (sy-y0)*dy) / len2
			if t < 0 {
				t = 0
			} else if t > 1 {
				t = 1
			}
			cx := x0 + t*dx
			cy := y0 + t*dy
			ddx := sx - cx
			ddy := sy - cy
			if ddx*ddx+ddy*ddy <= r2 {
				count++
			}
		}
	}
	return float64(count) / 16.0
}

// [S 0x140958f00]
// 椭圆描边有符号距离：|((x-cx)/rx)²+((y-cy)/ry)²−1|/hypot(2dx/rx²,2dy/ry²)。
// rx<=0||ry<=0 返回 MaxFloat64；hypot==0 返回 max(rx,ry)。
func qrCodeEllipseStrokeDistance(x, y, cx, cy, rx, ry float64) float64 {
	if rx <= 0 || ry <= 0 {
		return math.MaxFloat64
	}
	dx := x - cx
	dy := y - cy
	nx := dx / rx
	ny := dy / ry
	val := nx*nx + ny*ny - 1
	gx := 2 * dx / (rx * rx)
	gy := 2 * dy / (ry * ry)
	g := math.Hypot(gx, gy)
	if g == 0 {
		return math.Max(rx, ry)
	}
	return math.Abs(val) / g
}

// [S 0x140959080]
// 椭圆描边像素覆盖度（4×4 超采样）：采样点距离 <= strokeWidth 计数，×1/16。
func qrCodeEllipseStrokePixelCoverage(px, py int, cx, cy, rx, ry, strokeWidth float64) float64 {
	count := 0
	for row := 0; row < 4; row++ {
		sy := float64(py) + (float64(row)+0.5)*0.25 - 0.5
		for col := 0; col < 4; col++ {
			sx := float64(px) + (float64(col)+0.5)*0.25 - 0.5
			if qrCodeEllipseStrokeDistance(sx, sy, cx, cy, rx, ry) <= strokeWidth {
				count++
			}
		}
	}
	return float64(count) / 16.0
}

// [S 0x140959260]
// 平滑圆绘制：r_inner=max(0,r−0.8)，r_outer=r+0.8，bound=ceil(r+1.8)。
// dist²<=r_inner² 完全混合，>r_outer² 跳过，否则按覆盖度混合。
func drawSmoothCircleOnRGBA(img *image.RGBA, cx, cy, radius int, r, g, b, a byte) {
	if img == nil {
		return
	}
	rf := float64(radius)
	rInner := rf - 0.8
	if rInner < 0 {
		rInner = 0
	}
	rOuter := rf + 0.8
	bound := int(math.Ceil(rf + 1.8))
	for py := cy - bound; py <= cy+bound; py++ {
		for px := cx - bound; px <= cx+bound; px++ {
			dx := float64(px - cx)
			dy := float64(py - cy)
			d2 := dx*dx + dy*dy
			switch {
			case d2 <= rInner*rInner:
				blendQRCodeRGBAPixel(img, px, py, r, g, b, a)
			case d2 <= rOuter*rOuter:
				cov := byte(qrCodeCirclePixelCoverage(px, py, cx, cy, rf*rf) * 255)
				if cov != 0 {
					blendQRCodeRGBAPixel(img, px, py, r, g, b, byte((int(a)*int(cov)+127)/255))
				}
			}
		}
	}
}

// [S 0x1409596e0]
// 圆像素覆盖度（4×4 超采样）：sample = px + (row+0.5)*0.25 − 0.5，×1/16。
func qrCodeCirclePixelCoverage(px, py, cx, cy int, r2 float64) float64 {
	count := 0
	for row := 0; row < 4; row++ {
		sy := float64(py) + (float64(row)+0.5)*0.25 - 0.5
		for col := 0; col < 4; col++ {
			sx := float64(px) + (float64(col)+0.5)*0.25 - 0.5
			dx := sx - float64(cx)
			dy := sy - float64(cy)
			if dx*dx+dy*dy <= r2 {
				count++
			}
		}
	}
	return float64(count) / 16.0
}

// [S 0x1409597a0]
// 单像素混合：以新像素 alpha 为覆盖度与原像素做 coverage 混合。
func blendQRCodeRGBAPixel(img *image.RGBA, x, y int, r, g, b, a byte) {
	if img == nil {
		return
	}
	if x < img.Rect.Min.X || x >= img.Rect.Max.X || y < img.Rect.Min.Y || y >= img.Rect.Max.Y {
		return
	}
	off := img.PixOffset(x, y)
	pr, pg, pb, pa := img.Pix[off], img.Pix[off+1], img.Pix[off+2], img.Pix[off+3]
	nr, ng, nb, na := blendQRCodeRGBAByCoverage(pr, pg, pb, pa, r, g, b, a, a)
	img.Pix[off], img.Pix[off+1], img.Pix[off+2], img.Pix[off+3] = nr, ng, nb, na
}

// [S 0x1409599c0]
// 以 r,g,b,a 填充 rect 内所有像素（逐像素 coverage 混合）。
func fillRGBA(img *image.RGBA, rect image.Rectangle, r, g, b, a byte) {
	if img == nil {
		return
	}
	rect = rect.Intersect(img.Rect)
	for py := rect.Min.Y; py < rect.Max.Y; py++ {
		for px := rect.Min.X; px < rect.Max.X; px++ {
			blendQRCodeRGBAPixel(img, px, py, r, g, b, a)
		}
	}
}

// [S 0x140959ee0]
// 选区窗口矩形：r 平移 (−offX,−offY) 后与 s 求交。
func qrCodeSelectionWindowRect(r image.Rectangle, offX, offY int, s image.Rectangle) image.Rectangle {
	return r.Sub(image.Pt(offX, offY)).Intersect(s)
}

// [S 0x140959f80]
// 带光标选项的虚拟屏幕快照：仅设置 options.captureCursor，其余字段零值，
// 转发到既有后端 captureScreenshotVirtualScreenSnapshotWithOptions。
func captureQRCodeVirtualScreenSnapshotWithCursor(captureCursor bool) qrCodeScreenSnapshot {
	return captureScreenshotVirtualScreenSnapshotWithOptions(screenshotScreenCaptureOptions{
		captureCursor: captureCursor,
	})
}

// [S-sig 0x14095a360]
// 选区 PNG 编码。签名由 body 起始 (rax=img, rbx..rsi=rect) 证实为
// img + image.Rectangle；体经 image.Rectangle.Intersect + klauspost/png 编码，
// 返回 ([]byte, error)。
func encodeQRCodeSelectionPNG(img *image.RGBA, rect image.Rectangle) ([]byte, error) {
	return nil, nil
}

// [S 0x14095a5a0]
// 生成选区 8 个缩放手柄（4 角 + 4 边中点），midX=(Min.X+Max.X−1)/2。
func qrCodeSelectionResizeHandlePositions(rect image.Rectangle) [8]qrCodeSelectionResizeHandle {
	midX := (rect.Min.X + rect.Max.X - 1) / 2
	midY := (rect.Min.Y + rect.Max.Y - 1) / 2
	return [8]qrCodeSelectionResizeHandle{
		{id: 0, x: rect.Min.X, y: rect.Min.Y},
		{id: 1, x: rect.Max.X - 1, y: rect.Min.Y},
		{id: 2, x: rect.Max.X - 1, y: rect.Max.Y - 1},
		{id: 3, x: rect.Min.X, y: rect.Max.Y - 1},
		{id: 4, x: midX, y: rect.Min.Y},
		{id: 5, x: rect.Max.X - 1, y: midY},
		{id: 6, x: midX, y: rect.Max.Y - 1},
		{id: 7, x: rect.Min.X, y: midY},
	}
}

// [S 0x14095a6a0]
// 返回距 p 10px 内最近手柄 id（1-8），无则 0。
func qrCodeSelectionResizeHandleAtPoint(rect image.Rectangle, p image.Point) uint8 {
	best := uint8(0)
	bestD := 100 // 10px 平方
	for _, h := range qrCodeSelectionResizeHandlePositions(rect) {
		dx := h.x - p.X
		dy := h.y - p.Y
		d := dx*dx + dy*dy
		if d <= bestD {
			bestD = d
			best = h.id + 1
		}
	}
	return best
}

// [S-sig 0x14095a840]
// 选区缩放。签名由 morestack (rax..r10 共 8 qword + r11b + 2 个栈 int) 证实；
// 语义为 (rect,bounds,handle,mouseX,mouseY)，体含 Intersect 与 8 方向 clamp。
func resizeQRCodeSelectionRect(rect, bounds image.Rectangle, handle uint8, mouseX, mouseY int) image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14095aa20]
// 光标位置：GetCursorPos + ScreenToClient。签名由 body (rax=hwnd，调用 w32
// GetCursorPos/ScreenToClient) 证实；返回 (image.Point, bool)。
func qrCodeSelectionCursorPoint(hwnd uintptr) (image.Point, bool) {
	return image.Point{}, false
}

// [S 0x14095aaa0]
// 构建选区脏矩形：a=prev∩bounds，b=cur∩bounds；a==b 返回 nil，否则拼接
// subtract(a,b)、subtract(b,a)、a/b 的边框矩形，并过滤到 bounds 内。
func buildQRCodeSelectionDirtyRects(prev, cur, bounds image.Rectangle) []image.Rectangle {
	a := prev.Intersect(bounds)
	b := cur.Intersect(bounds)
	if a == b {
		return nil
	}
	out := make([]image.Rectangle, 0, 8)
	out = append(out, subtractQRCodeSelectionRect(a, b)...)
	out = append(out, subtractQRCodeSelectionRect(b, a)...)
	out = append(out, buildQRCodeSelectionBorderRects(a)...)
	out = append(out, buildQRCodeSelectionBorderRects(b)...)
	return filterQRCodeSelectionRectangles(out, bounds)
}

// [S-sig 0x14095af40]
// 带尺寸标签的脏矩形构建。签名由 morestack 8 qword + 栈 4 int 证实为
// (prev,cur,bounds) 三个 image.Rectangle；体在 DirtyRects 基础上附尺寸标签矩形。
func buildQRCodeSelectionDirtyRectsWithSizeLabels(prev, cur, bounds image.Rectangle) []image.Rectangle {
	return nil
}

// [S 0x14095b2a0]
// 选区尺寸标签文本："%dx%d"，空选区返回空串。
func qrCodeSelectionSizeLabelText(r image.Rectangle) string {
	if r.Empty() {
		return ""
	}
	return fmt.Sprintf("%dx%d", r.Dx(), r.Dy())
}

// [S 0x14095b380]
// 尺寸标签矩形：width=len(text)*8+16，height=21，交由 InfoPlacementRect 放置。
func qrCodeSelectionSizeLabelRect(selection, bounds image.Rectangle) image.Rectangle {
	if selection.Empty() || bounds.Empty() {
		return image.Rectangle{}
	}
	width := len(qrCodeSelectionSizeLabelText(selection))*8 + 16
	height := 21
	return qrCodeSelectionInfoPlacementRect(selection, bounds, width, height)
}

// [S-sig 0x14095b4a0]
// 圆角布局。签名由 asm 证实为 3 个 image.Rectangle 入参，返回 160 字节
// (5×image.Rectangle) 栈结构；asm 无法区分 [5]image.Rectangle 与具名结构，
// 暂以 [5]image.Rectangle 表达。
func qrCodeSelectionCornerRadiusLayout(a, b, c image.Rectangle) [5]image.Rectangle {
	return [5]image.Rectangle{}
}

// [S-sig 0x14095ba40]
// 信息放置矩形。签名由 morestack (9 qword + 栈 int) 证实为
// (selection,bounds,width,height)；体最终 clampQRCodeSelectionSizeLabelRect。
func qrCodeSelectionInfoPlacementRect(selection, bounds image.Rectangle, width, height int) image.Rectangle {
	return image.Rectangle{}
}

// [S 0x14095bca0]
// 圆角半径拇指 X：Min.X + int(Dx*radius/maxRadius)。
func qrCodeCornerRadiusThumbX(rect image.Rectangle, radius, maxRadius int) int {
	if maxRadius == 0 || rect.Dx() == 0 {
		return rect.Min.X
	}
	return rect.Min.X + int(float64(rect.Dx())*float64(radius)/float64(maxRadius))
}

// [S 0x14095bd80]
// 由点反推圆角半径：int(clamp(px,Min.X,Max.X)−Min.X 的比例 × maxRadius)，钳到 [0,maxRadius]。
func qrCodeCornerRadiusFromPoint(rect image.Rectangle, px, maxRadius int) int {
	if rect.Dx() <= 0 || maxRadius <= 0 {
		return 0
	}
	x := px
	if x < rect.Min.X {
		x = rect.Min.X
	}
	if x > rect.Max.X {
		x = rect.Max.X
	}
	v := int(float64(x-rect.Min.X) * float64(maxRadius) / float64(rect.Dx()))
	if v < 0 {
		v = 0
	}
	if v > maxRadius {
		v = maxRadius
	}
	return v
}

// [S 0x14095be80]
// 生成 4 个圆角子矩形：r=min(radius, min(Dx,Dy)/4)，取四个角 r×r 区域，
// 经 filterQRCodeSelectionRectangles 过滤到 rect 内。
func qrCodeRoundedCornerRects(rect image.Rectangle, radius int) []image.Rectangle {
	if rect.Empty() || radius <= 0 {
		return nil
	}
	d := rect.Dx()
	if dy := rect.Dy(); dy < d {
		d = dy
	}
	r := d / 4
	if radius < r {
		r = radius
	}
	if r <= 0 {
		return nil
	}
	rects := [4]image.Rectangle{
		image.Rect(rect.Min.X, rect.Min.Y, rect.Min.X+r, rect.Min.Y+r),
		image.Rect(rect.Max.X-r, rect.Min.Y, rect.Max.X, rect.Min.Y+r),
		image.Rect(rect.Min.X, rect.Max.Y-r, rect.Min.X+r, rect.Max.Y),
		image.Rect(rect.Max.X-r, rect.Max.Y-r, rect.Max.X, rect.Max.Y),
	}
	return filterQRCodeSelectionRectangles(rects[:], rect)
}

// [S-sig 0x14095c040]
// 构建圆角掩膜。签名由 morestack (仅 rax=radius int) 证实；返回 *image.Alpha，
// 体为超采样圆角 alpha 生成。
func buildQRCodeRoundedCornerMask(radius int) *image.Alpha {
	return nil
}

// [S 0x14095c3c0]
// 圆角掩膜覆盖度：返回 px,py 在 rect 内对应的掩膜 alpha 值。
// 返回值语义：0=rect 外，-1=实心/掩膜无效，0-255=圆角区 alpha。
func qrCodeRoundedCornerMaskCoverage(px, py int, rect image.Rectangle, mask *image.Alpha) int {
	if px < rect.Min.X || px >= rect.Max.X || py < rect.Min.Y || py >= rect.Max.Y {
		return 0
	}
	if mask == nil || mask.Rect.Max.X <= mask.Rect.Min.X || mask.Rect.Min.Y >= mask.Rect.Max.Y {
		return -1
	}
	w := mask.Rect.Max.X - mask.Rect.Min.X
	h := mask.Rect.Max.Y - mask.Rect.Min.Y
	r := w
	if h < r {
		r = h
	}
	if r <= 0 {
		return -1
	}
	leftInner := rect.Min.X + r
	rightInner := rect.Max.X - r
	topInner := rect.Min.Y + r
	bottomInner := rect.Max.Y - r
	if px >= leftInner && px < rightInner {
		return -1
	}
	if py >= topInner && py < bottomInner {
		return -1
	}
	sx := px - rect.Min.X
	if px >= rightInner {
		sx = rect.Max.X - px - 1
	}
	sy := py - rect.Min.Y
	if py >= bottomInner {
		sy = rect.Max.Y - py - 1
	}
	mx := mask.Rect.Min.X + sx
	my := mask.Rect.Min.Y + sy
	if mx < mask.Rect.Min.X || mx >= mask.Rect.Max.X || my < mask.Rect.Min.Y || my >= mask.Rect.Max.Y {
		return 0
	}
	return int(mask.Pix[my*mask.Stride+mx])
}

// [S-sig 0x14095c520]
// 圆角选区边框绘制（图像域）。签名由 morestack 9 qword 证实为
// img + rect + bounds；体经 blendQRCodeRGBAPixel 逐点描边。
func drawQRCodeRoundedSelectionBorderOnImage(img *image.RGBA, rect, bounds image.Rectangle) {}

// [S-sig 0x14095cd20]
// 应用圆角（将掩膜乘入图像 alpha）。签名由 morestack 2 qword (img, mask) 证实；
// 返回 bool 表示是否实际应用。
func applyQRCodeRoundedCorners(img *image.RGBA, mask *image.Alpha) bool {
	return false
}

// [S 0x14095d080]
// 覆盖度混合两个颜色：out = (c1*coverage + c2*(255−coverage))/255，逐通道。
func blendQRCodeRGBAByCoverage(r1, g1, b1, a1, r2, g2, b2, a2 byte, coverage byte) (byte, byte, byte, byte) {
	cov := int(coverage)
	inv := 255 - cov
	return byte((int(r1)*cov + int(r2)*inv) / 255),
		byte((int(g1)*cov + int(g2)*inv) / 255),
		byte((int(b1)*cov + int(b2)*inv) / 255),
		byte((int(a1)*cov + int(a2)*inv) / 255)
}

// [S 0x14095d180]
// 将 rect 钳制到 bounds 内。
func clampQRCodeSelectionSizeLabelRect(rect, bounds image.Rectangle) image.Rectangle {
	if rect.Empty() || bounds.Empty() {
		return image.Rectangle{}
	}
	rect = rect.Intersect(bounds)
	if rect.Empty() {
		return image.Rectangle{}
	}
	return rect
}

// [S 0x14095d2c0]
// rect 减 other：不相交返回 {rect}，否则返回最多 4 条裁剪带。
func subtractQRCodeSelectionRect(rect, other image.Rectangle) []image.Rectangle {
	r := rect.Intersect(other)
	if r.Empty() {
		return []image.Rectangle{rect}
	}
	out := make([]image.Rectangle, 0, 4)
	if rect.Min.Y < r.Min.Y {
		out = append(out, image.Rect(rect.Min.X, rect.Min.Y, rect.Max.X, r.Min.Y))
	}
	if r.Max.Y < rect.Max.Y {
		out = append(out, image.Rect(rect.Min.X, r.Max.Y, rect.Max.X, rect.Max.Y))
	}
	if rect.Min.X < r.Min.X {
		out = append(out, image.Rect(rect.Min.X, r.Min.Y, r.Min.X, r.Max.Y))
	}
	if r.Max.X < rect.Max.X {
		out = append(out, image.Rect(r.Max.X, r.Min.Y, rect.Max.X, r.Max.Y))
	}
	return out
}

// [S 0x14095d540]
// 边框矩形：空/宽<=4/高<=4 返回 {rect}，否则返回 4 条 2px 边带。
func buildQRCodeSelectionBorderRects(rect image.Rectangle) []image.Rectangle {
	if rect.Empty() || rect.Dx() <= 4 || rect.Dy() <= 4 {
		return []image.Rectangle{rect}
	}
	return []image.Rectangle{
		image.Rect(rect.Min.X, rect.Min.Y, rect.Max.X, rect.Min.Y+2),
		image.Rect(rect.Min.X, rect.Max.Y-2, rect.Max.X, rect.Max.Y),
		image.Rect(rect.Min.X, rect.Min.Y, rect.Min.X+2, rect.Max.Y),
		image.Rect(rect.Max.X-2, rect.Min.Y, rect.Max.X, rect.Max.Y),
	}
}

// [S 0x14095d700]
// 过滤矩形：与 bounds 求交、去空、去重。
func filterQRCodeSelectionRectangles(rects []image.Rectangle, bounds image.Rectangle) []image.Rectangle {
	out := make([]image.Rectangle, 0, len(rects))
	seen := make(map[image.Rectangle]struct{}, len(rects))
	for _, r := range rects {
		r = r.Intersect(bounds)
		if r.Empty() {
			continue
		}
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		out = append(out, r)
	}
	return out
}

// [S-sig 0x14095da60]
// GDI 选区边框。签名由 body 起始 (rax=hdc, rbx..rsi=rect, r8=color) 证实；
// 体为 RECT + 画刷填充调用链。
func drawQRCodeSelectionBorder(hdc uintptr, rect image.Rectangle, color int) {}

// [S-sig 0x14095dbc0]
// GDI 圆角选区边框。签名经 morestack 证实为 hdc + rect + radius + color。
func drawQRCodeRoundedSelectionBorder(hdc uintptr, rect image.Rectangle, radius, color int) {}

// [S-sig 0x14095e0e0]
// GDI 虚线选区边框。签名经 morestack 证实为 hdc + rect + color；体为虚线画刷。
func drawQRCodeDashedSelectionBorder(hdc uintptr, rect image.Rectangle, color int) {}

// [S-sig 0x14095e840]
// 创建画笔缓冲（DC + 兼容位图 + canvas）。签名由 morestack 4 qword (rect)
// 证实；返回 40 字节 qrCodePaintBuffer。
func createQRCodePaintBuffer(rect image.Rectangle) qrCodePaintBuffer {
	return qrCodePaintBuffer{}
}

// [S-sig 0x14095eaa0]
// 冲刷画笔缓冲到 DC。签名由 morestack (rax..r8 共 6 qword + r9b) 证实为
// hdc + rect + bitmap + ok；体为 BitBlt(SRCCOPY)。
func flushQRCodePaintBuffer(hdc uintptr, rect image.Rectangle, bitmap uintptr, ok bool) {}
