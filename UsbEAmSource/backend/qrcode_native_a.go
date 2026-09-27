// AUTO-RECONSTRUCTED — DOMAIN: qrcode native screen-selection / annotation
// Source: UsbEAm_Launcher 1.0.3 (Go 1.25.12, PE64), disassembled from
// main.qrcode_* symbols. Tier markers:
//
//	[S VA]      — body fully translated from asm (pure geometry / math).
//	[S-sig VA]  — signature proven from asm; body is a faithful zero skeleton
//	              (Win32 / GDI / syscall.LazyProc / package-global deps).
//
// 研究用途
package main

import (
	"image"
	"image/color"
	"strconv"
	"strings"
	"time"
)

// [S-sig 0x1409341e0] orchestrates captureQRCodesFromNativeSelectionResultWithOptions
// and post-processes the returned qrCodeNativeSelectionResult (validity / finalize /
// PNG path). Heavily Win32/GDI-adjacent; body not translated.
func captureQRCodesFromNativeSelection(options qrCodeNativeSelectionOptions) qrCodeNativeSelectionResult {
	return qrCodeNativeSelectionResult{}
}

// [S-sig 0x140934500] reserves a session, captures preview snapshot, builds the
// selection window, and blocks until accepted/finalized. Returns the large result
// by value (stack) plus error. Body not translated.
func captureQRCodesFromNativeSelectionResultWithOptions(options qrCodeNativeSelectionOptions) (qrCodeNativeSelectionResult, error) {
	return qrCodeNativeSelectionResult{}, nil
}

// [S-sig 0x140934ac0] captures the virtual-screen snapshot backing the preview,
// branching between GDI-with-cursor / plain GDI / options-based paths.
func captureQRCodeSelectionPreviewSnapshot(options qrCodeNativeSelectionOptions) qrCodeScreenSnapshot {
	return qrCodeScreenSnapshot{}
}

// [S-sig 0x140934d60] captures the final selection rect and encodes it as PNG.
// 4 register args = image.Rectangle. Returns ([]byte, error).
func captureQRCodeFinalSelectionPNG(selection image.Rectangle) ([]byte, error) {
	return nil, nil
}

// [S-sig 0x140934ec0] registers the selection window class once (sync.Once) and
// returns any registration error.
func ensureQRCodeSelectionWindowClass() error {
	return nil
}

// [S-sig 0x140934f20] constructor: 6 register args = qrCodeScreenSnapshot, plus
// options on stack. Allocates session, creates bitmaps/DCs, sets defaults.
func newQRCodeScreenSelectionSession(snapshot qrCodeScreenSnapshot, options qrCodeNativeSelectionOptions) (*qrCodeScreenSelectionSession, error) {
	return nil, nil
}

// [S-sig 0x140935400] wraps CreateHBITMAPFromImage for a non-empty *image.RGBA.
func createQRCodeBitmapHandle(img *image.RGBA) (uintptr, error) {
	return 0, nil
}

// [S-sig 0x140935520] creates an app-compatible memory DC and selects a bitmap.
// Returns (hdc, previous, error).
func createQRCodeBitmapDC(bitmap uintptr) (uintptr, uintptr, error) {
	return 0, 0, nil
}

// [S-sig 0x140936ac0] calls SetWindowDisplayAffinity(hwnd, WDA_EXCLUDEFROMCAPTURE=0x11).
func applyQRCodeSelectionCaptureExclusion(hwnd uintptr) {
}

// [S-sig 0x1409375a0] atomically reserves the shared selection-session slot under a
// package mutex; returns true iff the slot was free.
func tryReserveQRCodeSelectionSession() bool {
	return false
}

// [S-sig 0x1409376e0] promotes a reserved session to the active slot; fails if the
// slot is not reserved or already active.
func activateReservedQRCodeSelectionSession(session *qrCodeScreenSelectionSession) bool {
	return false
}

// [S-sig 0x140937860] clears the active session slot when it still points at session.
func finishQRCodeSelectionSession(session *qrCodeScreenSelectionSession) {
}

// [S-sig 0x1409379e0] returns the currently active selection session (nil if none).
func activeQRCodeSelectionSession() *qrCodeScreenSelectionSession {
	return nil
}

// [S-sig 0x140937b00] window proc: dispatches to the active session's handleMessage,
// else falls back to DefWindowProc via syscall.LazyProc.Call.
func qrCodeSelectionWindowProc(hwnd uintptr, msg uint32, wparam, lparam uintptr) uintptr {
	return 0
}

// [S 0x14093e000] returns true when a hover control should be re-evaluated: the point
// moved more than 4px in either axis, the previous sample time is zero, or >=90ms
// elapsed since the previous sample.
func shouldRefreshQRCodeControlHover(prev, cur image.Point, at, now time.Time) bool {
	dx := cur.X - prev.X
	dy := cur.Y - prev.Y
	if dx > 4 || dx < -4 || dy > 4 || dy < -4 {
		return true
	}
	if at.IsZero() {
		return true
	}
	return now.Sub(at) >= 90*time.Millisecond
}

// [S 0x14093e120] returns true when a cached hover rect can be reused: the point is
// strictly inside rect and the rect area is <= 240000.
func shouldReuseQRCodeControlHover(rect image.Rectangle, point image.Point) bool {
	if point.X <= rect.Min.X || point.X >= rect.Max.X || point.Y <= rect.Min.Y || point.Y >= rect.Max.Y {
		return false
	}
	return rect.Dx()*rect.Dy() <= 240000
}

// [S 0x14093e860] intersects the cached hover rect with a control rect, provided both
// rects are non-empty and the point lies inside the hover rect; else zero rect.
func qrCodeControlSelectionFromCachedHover(hoverRect image.Rectangle, point image.Point, controlRect image.Rectangle) image.Rectangle {
	if hoverRect.Dx() <= 0 || hoverRect.Dy() <= 0 {
		return image.Rectangle{}
	}
	if controlRect.Dx() <= 0 || controlRect.Dy() <= 0 {
		return image.Rectangle{}
	}
	if !point.In(hoverRect) {
		return image.Rectangle{}
	}
	return hoverRect.Intersect(controlRect)
}

// [S-sig 0x140940760] hides an overlay window via ShowWindow(hwnd, SW_HIDE) and
// SetWindowPos(hwnd, 0, 0, 0, 0, 0, SWP_NOMOVE|...=0x97).
func hideScreenshotOverlayWindow(hwnd uintptr) {
}

// [S-sig 0x140946ce0] normalizes clipboard annotation text through a
// strings.NewReplacer with 4 old->new pairs. The exact pairs are literal string
// constants in .rodata (undecoded); signature proven, body not translated.
func normalizeQRCodeAnnotationClipboardText(s string) string {
	return ""
}

// [S-sig 0x140947040] reads two system metrics via syscall.LazyProc.Call
// (indices 0x11 and 0x12) and returns whether each is negative (int16 < 0).
func screenshotCurrentControlSelectionPreference() (bool, bool) {
	return false, false
}

// [S-sig 0x140947100] trims and case-insensitively matches name against a package
// global []string font-family list, returning the canonical match or a default.
func normalizedQRCodeAnnotationFontFamily(name string) string {
	return ""
}

// [S-sig 0x140947200] returns the font family following current in the package
// global list (wrapping around), else the default.
func nextQRCodeAnnotationFontFamily(current string) string {
	return ""
}

// [S 0x140947ec0] returns a copy of the stroke with start, end and every point
// translated by (dx, dy).
func offsetQRCodeAnnotationStroke(s qrCodeAnnotationStroke, dx, dy int) qrCodeAnnotationStroke {
	s.start.X += dx
	s.start.Y += dy
	s.end.X += dx
	s.end.Y += dy
	if len(s.points) > 0 {
		pts := make([]image.Point, len(s.points))
		copy(pts, s.points)
		for i := range pts {
			pts[i].X += dx
			pts[i].Y += dy
		}
		s.points = pts
	}
	return s
}

// [S-sig 0x1409481a0] translates a stroke by (dx, dy) and clamps it against a limit
// rect (6 register args = dx, dy, limit.Min.X, limit.Min.Y, limit.Max.X, limit.Max.Y).
// Complex; signature defensible, body not translated.
func translateQRCodeAnnotationStroke(s qrCodeAnnotationStroke, dx, dy int, limit image.Rectangle) qrCodeAnnotationStroke {
	return qrCodeAnnotationStroke{}
}

// [S 0x140948640] clamps a text-editor anchor point into the limit rect: computes the
// raw editor rect, clamps it (margin 8), and applies the resulting min-corner delta
// to the point.
func clampQRCodeAnnotationTextEditorPoint(p image.Point, limit image.Rectangle) image.Point {
	if limit.Dx() <= 0 || limit.Dy() <= 0 {
		return p
	}
	raw, _ := qrCodeAnnotationTextEditorRawRectForPointWithLimit(p, limit)
	clamped := clampQRCodeAnnotationRect(raw, limit, 8)
	return image.Point{X: p.X + (clamped.Min.X - raw.Min.X), Y: p.Y + (clamped.Min.Y - raw.Min.Y)}
}

// [S 0x140948740] parses "#RRGGBB" (optional '#', trimmed) into RGBA components.
// NOTE: asm returns full-width registers for r/g/b (strconv.ParseUint uint64 result)
// and a=0xffffffff; the exact integer widths are unproven — uint8 assumed to match
// the color.RGBA domain.
func parseQRCodeAnnotationHexColor(s string) (r, g, b, a uint8, ok bool) {
	s = strings.TrimSpace(s)
	if len(s) > 0 && s[0] == '#' {
		s = s[1:]
	}
	if len(s) != 6 {
		return 0, 0, 0, 0, false
	}
	rv, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil {
		return 0, 0, 0, 0, false
	}
	gv, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		return 0, 0, 0, 0, false
	}
	bv, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		return 0, 0, 0, 0, false
	}
	return uint8(rv), uint8(gv), uint8(bv), 0xff, true
}

// [S 0x14094c700] unions two rects, treating an empty operand as identity.
func unionQRCodeAnnotationRects(a, b image.Rectangle) image.Rectangle {
	if a.Dx() <= 0 || a.Dy() <= 0 {
		return b
	}
	if b.Dx() <= 0 || b.Dy() <= 0 {
		return a
	}
	return a.Union(b)
}

// [S 0x14094cbc0] computes the text-editor rect for a point/limit, clamped with
// margin 8 against the (possibly defaulted) limit returned by the raw helper.
func qrCodeAnnotationTextEditorRectForPoint(p image.Point, limit image.Rectangle) image.Rectangle {
	raw, lim := qrCodeAnnotationTextEditorRawRectForPointWithLimit(p, limit)
	return clampQRCodeAnnotationRect(raw, lim, 8)
}

// [S 0x14094cc40] computes an un-clamped editor rect for a point, defaulting an empty
// limit to a 676x148 rect anchored at the point, with width clamped to [360,660] and
// height 132 (top edge 36 above the point, or 146 above when the limit's bottom is
// tight). Returns (rect, limit) so callers can reuse the effective limit.
func qrCodeAnnotationTextEditorRawRectForPointWithLimit(p image.Point, limit image.Rectangle) (image.Rectangle, image.Rectangle) {
	if limit.Dx() <= 0 || limit.Dy() <= 0 {
		limit = image.Rect(p.X, p.Y, p.X+676, p.Y+148)
	}
	w := limit.Dx() - 16
	if w < 360 {
		w = 360
	}
	if w > 660 {
		w = 660
	}
	left := p.X - 52
	right := left + w
	top := p.Y + 36
	if limit.Max.Y-8 < p.Y+168 {
		top = p.Y - 146
	}
	bottom := top + 132
	return image.Rect(left, top, right, bottom), limit
}

// [S-sig 0x14094cd20] computes an editor rect that avoids a preview rect, iterating
// over candidate offsets and clamping with margin 8. Complex; signature defensible,
// body not translated.
func qrCodeAnnotationTextEditorRectAvoidingPreview(p image.Point, limit, preview image.Rectangle) image.Rectangle {
	return image.Rectangle{}
}

// [S 0x14094d120] clamps rect into limit with the given margin on each side, keeping
// the rect's size. Empty rect or limit returns rect unchanged.
func clampQRCodeAnnotationRect(rect image.Rectangle, limit image.Rectangle, margin int) image.Rectangle {
	if rect.Dx() <= 0 || rect.Dy() <= 0 || limit.Dx() <= 0 || limit.Dy() <= 0 {
		return rect
	}
	w := rect.Dx()
	h := rect.Dy()

	lowX := limit.Min.X + margin
	highX := limit.Max.X - margin - w
	if highX < lowX {
		lowX = limit.Min.X
		highX = limit.Max.X - w
		if highX < lowX {
			highX = lowX
		}
	}
	lowY := limit.Min.Y + margin
	highY := limit.Max.Y - margin - h
	if highY < lowY {
		lowY = limit.Min.Y
		highY = limit.Max.Y - h
		if highY < lowY {
			highY = lowY
		}
	}

	minX := rect.Min.X
	if minX < lowX {
		minX = lowX
	}
	if minX > highX {
		minX = highX
	}
	minY := rect.Min.Y
	if minY < lowY {
		minY = lowY
	}
	if minY > highY {
		minY = highY
	}
	return image.Rect(minX, minY, minX+w, minY+h)
}

// [S-sig 0x14094d4e0] draws small annotation helper text (info font, TextOut via
// LazyProc). Args: hdc, text string, x, y, r, g, b. Returns an int status.
func drawQRCodeAnnotationSmallText(hdc uintptr, text string, x, y int, r, g, b uint8) int {
	return 0
}

// [S-sig 0x14094d800] creates a temporary measure HDC (GetDC via LazyProc) and runs
// fn(hdc), restoring the previous object. Callback signature not fully proven.
func withQRCodeAnnotationTextMeasureHDC(fn func(hdc uintptr) int) int {
	return 0
}

// [S-sig 0x14094dbc0] returns the visible rune range of text fitting within maxWidth,
// anchored at startIndex, measuring via hdc (0 = estimate). Complex binary search;
// signature defensible, body not translated.
func qrCodeAnnotationTextVisibleRange(hdc uintptr, text string, maxWidth, startIndex int) (int, int) {
	return 0, 0
}

// [S-sig 0x14094e0c0] returns the caret rune index nearest the given x coordinate.
// Complex binary search; body not translated.
func qrCodeAnnotationTextIndexForX(hdc uintptr, text string, maxWidth, startIndex, x int) int {
	return 0
}

// [S-sig 0x14094e320] measures text width via GetTextExtentPoint32 (LazyProc) on hdc,
// falling back to a rune-count estimate when hdc==0 or text is empty.
func measureQRCodeAnnotationTextWidth(hdc uintptr, text string) int {
	return 0
}

// [S-sig 0x14094e460] dispatches a stroke to the matching draw routine by tool name.
func drawQRCodeAnnotationStroke(hdc uintptr, stroke qrCodeAnnotationStroke) {
}

// [S-sig 0x14094ea80] creates a GDI pen for a stroke color/width (CreatePen via
// LazyProc). Color bytes arrive in separate registers; exact grouping ambiguous.
func createQRCodeAnnotationPen(r, g, b, a uint8, lineWidth int) uintptr {
	return 0
}

// [S-sig 0x14094eba0] draws a line (MoveToEx + LineTo via LazyProc).
func drawQRCodeAnnotationLine(hdc uintptr, start, end image.Point, color color.RGBA, lineWidth int) {
}

// [S-sig 0x14094ecc0] draws a rectangle outline (Rectangle via LazyProc).
func drawQRCodeAnnotationRect(hdc uintptr, rect image.Rectangle, color color.RGBA, lineWidth int) {
}

// [S-sig 0x14094eec0] draws an ellipse outline; computes radii with floating point
// (Ellipse via LazyProc).
func drawQRCodeAnnotationEllipse(hdc uintptr, rect image.Rectangle, color color.RGBA, lineWidth int) {
}

// [S-sig 0x14094f420] draws an arrow: a line plus two head segments from
// qrCodeAnnotationArrowHeadPoints.
func drawQRCodeAnnotationArrow(hdc uintptr, start, end image.Point, color color.RGBA, lineWidth int) {
}

// [S-sig 0x14094f520] draws a pen path through points (Polyline via LazyProc).
func drawQRCodeAnnotationPenPath(hdc uintptr, points []image.Point) {
}

// [S-sig 0x14094f6a0] draws annotation text with an explicit color (font + TextOut).
func drawQRCodeAnnotationTextWithColor(hdc uintptr, stroke qrCodeAnnotationStroke, color color.RGBA) {
}

// [S-sig 0x14094fe00] convenience wrapper: forwards the stroke to
// qrCodeAnnotationTextBoundsForHDC.
func qrCodeAnnotationTextW32Rect(hdc uintptr, stroke qrCodeAnnotationStroke) image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x14094fe80] measures text size via GDI and derives a padded bounds rect.
func qrCodeAnnotationTextBoundsForHDC(hdc uintptr, stroke qrCodeAnnotationStroke) image.Rectangle {
	return image.Rectangle{}
}

// [S 0x140950000] estimates a text stroke's bounds from trimmed rune count, font size,
// and measured textWidth (no GDI calls).
func qrCodeAnnotationTextEstimatedBounds(stroke qrCodeAnnotationStroke) image.Rectangle {
	text := strings.TrimSpace(stroke.text)
	if text == "" {
		return image.Rectangle{}
	}
	n := len([]rune(text))
	base := stroke.fontSize
	if base < 12 {
		base = 12
	}
	w := n * base
	if w < 80 {
		w = 80
	}
	if w < stroke.textWidth {
		w = stroke.textWidth
	}
	h := base + 8
	if h < 18 {
		h = 18
	}
	return image.Rect(stroke.start.X, stroke.start.Y, stroke.start.X+w, stroke.start.Y+h)
}

// [S-sig 0x1409500e0] measures text width with a fresh DC and font. Takes the stroke
// on the stack; body not translated.
func qrCodeAnnotationTextMeasuredWidth(stroke qrCodeAnnotationStroke) int {
	return 0
}

// [S-sig 0x140950660] measures text extent via GetTextExtentPoint32 on hdc.
// Returns (width, height, ok).
func measureQRCodeAnnotationTextSizeForHDC(hdc uintptr, text string) (int, int, bool) {
	return 0, 0, false
}

// [S-sig 0x1409507a0] creates a GDI font (CreateFontIndirect via LazyProc) with the
// given size/weight/family/quality.
func createQRCodeAnnotationFontWithQuality(fontSize, weight int, fontFamily string, quality uint8) uintptr {
	return 0
}

// [S-sig 0x140950900] creates the number-annotation font (fixed family, weight 0x2bc).
func createQRCodeAnnotationNumberFont(fontSize int, quality uint8) uintptr {
	return 0
}

// [S-sig 0x140950a20] draws a numbered annotation (mosaic + number text).
func drawQRCodeAnnotationNumber(hdc uintptr, stroke qrCodeAnnotationStroke) {
}

// [S-sig 0x140951800] pixelates a bitmap region (mosaic) into cells, averaging via
// averageQRCodeAnnotationColor. Args: bitmap *image.RGBA, rect, cellSize.
func drawQRCodeAnnotationMosaic(bitmap *image.RGBA, rect image.Rectangle, cellSize int) {
}

// [S-sig 0x140951b20] blurs a bitmap region in-place (box blur).
func drawQRCodeAnnotationBlur(bitmap *image.RGBA, rect image.Rectangle, radius int) {
}

// [S 0x140951e00] averages the RGBA of a bitmap region, returning (r, g, b, a).
// NOTE: asm returns r/g/b as full-width registers from 32-bit division and
// a=0xffffffff; exact widths unproven, uint8 assumed for the color domain.
func averageQRCodeAnnotationColor(img *image.RGBA, x0, y0, x1, y1 int) (r, g, b, a uint8) {
	var sr, sg, sb, n uint32
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			c := img.RGBAAt(x, y)
			sr += uint32(c.R)
			sg += uint32(c.G)
			sb += uint32(c.B)
			n++
		}
	}
	if n == 0 {
		return 0, 0, 0, 0xff
	}
	return uint8(sr / n), uint8(sg / n), uint8(sb / n), 0xff
}

// [S-sig 0x140951f40] draws the selection outline for a stroke (dashed border).
func drawQRCodeAnnotationSelection(hdc uintptr, stroke qrCodeAnnotationStroke) {
}

// [S-sig 0x140952300] returns the bounds of a stroke based on its tool:
// pen/line/arrow -> point bounds, text -> estimated bounds, number -> font-size based
// radius. Depends on qrCodeAnnotationArrowHeadPoints (not reconstructed) and float
// math; body not translated.
func qrCodeAnnotationStrokeBounds(stroke qrCodeAnnotationStroke) image.Rectangle {
	return image.Rectangle{}
}

// [S 0x1409527e0] returns the bounding rect of a point set, grown by one pixel
// (maxX+1, maxY+1); empty input yields a zero rect.
func qrCodeAnnotationPointsBounds(points []image.Point) image.Rectangle {
	if len(points) == 0 {
		return image.Rectangle{}
	}
	minX, minY := points[0].X, points[0].Y
	maxX, maxY := minX, minY
	for i := 1; i < len(points); i++ {
		p := points[i]
		if p.X < minX {
			minX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	return image.Rect(minX, minY, maxX+1, maxY+1)
}

// [S-sig 0x140952880] returns stroke bounds inflated by a tool-dependent paint margin
// (line width + 10, or fontSize/2 + 8 for text/number). Depends on
// qrCodeAnnotationStrokeBounds; body not translated.
func qrCodeAnnotationStrokePaintBounds(stroke qrCodeAnnotationStroke) image.Rectangle {
	return image.Rectangle{}
}
