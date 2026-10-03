package main

import (
	"image"
	"math"
	"time"
)

// screenshot 滚动所需的 user32 LazyProc（依赖 appicon_windows.go 中的 user32DLL）。
var procGetAsyncKeyState = user32DLL.NewProc("GetAsyncKeyState")

// screenshotScrollingCancelState 是滚动捕获的取消判定状态：pressed 记录上一次右键
// 按下状态，armed 记录"等待右键释放"的挂起标志（按下取消键后 arm，释放后 disarm）。
type screenshotScrollingCancelState struct {
	pressed bool
	armed   bool
}

// shouldCancelScreenshotScrolling 依右键状态推进取消判定状态机。
// [S 汇编 0x1409a32a0, 128B]：nil→false；read 得 (pressed,clicked)；
// armed：state.pressed=pressed，pressed→false，否则 armed=false 且 false；
// clicked→state.pressed=pressed 且 true；pressed→!state.pressed（上升沿）且
// state.pressed=pressed；否则 false。
func shouldCancelScreenshotScrolling(state *screenshotScrollingCancelState) bool {
	if state == nil {
		return false
	}
	pressed, clicked := readScreenshotScrollingRightButtonState()
	if state.armed {
		state.pressed = pressed
		if pressed {
			return false
		}
		state.armed = false
		return false
	}
	if clicked {
		state.pressed = pressed
		return true
	}
	cancel := false
	if pressed {
		cancel = !state.pressed
	}
	state.pressed = pressed
	return cancel
}

// waitScreenshotScrollingCancelable 在超时窗内轮询取消状态（sleep 步长上限 16ms）。
// [S 汇编 0x1409a3380, 256B]：timeout<=0 → shouldCancel；先查 shouldCancel→true；
// deadline=now.Add(timeout)；循环 time.Until<=0 → shouldCancel；sleep=min(remaining,16ms)；
// 每次 sleep 后 shouldCancel→true。
func waitScreenshotScrollingCancelable(timeout time.Duration, state *screenshotScrollingCancelState) bool {
	if timeout <= 0 {
		return shouldCancelScreenshotScrolling(state)
	}
	if shouldCancelScreenshotScrolling(state) {
		return true
	}
	deadline := time.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return shouldCancelScreenshotScrolling(state)
		}
		if remaining > 16*time.Millisecond {
			remaining = 16 * time.Millisecond
		}
		time.Sleep(remaining)
		if shouldCancelScreenshotScrolling(state) {
			return true
		}
	}
}

// readScreenshotScrollingRightButtonState 读取鼠标右键状态（GetAsyncKeyState，VK_RBUTTON=2）。
// [S] ASM 0x1409a3320：Call(2)；`bt eax,0xf; setb al` 取 bit15 → 当前是否按下（第 1 返回值）；
// `and ebx,1` 取 bit0 → 自上次调用后是否按过（第 2 返回值）。返回 (bool, bool)。
func readScreenshotScrollingRightButtonState() (bool, bool) {
	state, _, _ := procGetAsyncKeyState.Call(2) // VK_RBUTTON
	return state&0x8000 != 0, state&1 != 0
}

// sampleScreenshotFrameAverageDiffWithGrid 用网格采样两帧的 RGB 平均差异（8 位分量，
// 曼哈顿差 /3 取整）；尺寸收缩到两图交集与 height 上限，网格步长 = 尺寸/网格数（下限 1）。
// [S 汇编 0x1409a6040, 768B]：width=min(a.Dx,b.Dx)、height=min(aMaxY-aMinY,bMaxY-bMinY,height)；
// stepY=max(1,height/gridX)、stepX=max(1,width/gridY)；双循环逐点 RGBAAt 取 |ΔR|+|ΔG|+|ΔB|/3 累加；
// 无采样点返回 0x3fffffffffffffff，否则 sum/count。
func sampleScreenshotFrameAverageDiffWithGrid(a *image.RGBA, aMinY int, b *image.RGBA, bMinY int, height int, gridX, gridY int) int {
	width := a.Rect.Dx()
	if w := b.Rect.Dx(); w < width {
		width = w
	}
	h := a.Rect.Max.Y - aMinY
	if h2 := b.Rect.Max.Y - bMinY; h2 < h {
		h = h2
	}
	if height > h {
		height = h
	}
	if width <= 0 || height <= 0 {
		return math.MaxInt64 >> 1
	}
	stepY := height / gridX
	if stepY < 1 {
		stepY = 1
	}
	stepX := width / gridY
	if stepX < 1 {
		stepX = 1
	}
	var sum, count int
	for y := 0; y < height; y += stepY {
		for x := 0; x < width; x += stepX {
			ca := a.RGBAAt(a.Rect.Min.X+x, aMinY+y)
			cb := b.RGBAAt(b.Rect.Min.X+x, bMinY+y)
			dr := int(ca.R) - int(cb.R)
			if dr < 0 {
				dr = -dr
			}
			dg := int(ca.G) - int(cb.G)
			if dg < 0 {
				dg = -dg
			}
			db := int(ca.B) - int(cb.B)
			if db < 0 {
				db = -db
			}
			sum += (dr + dg + db) / 3
			count++
		}
	}
	if count == 0 {
		return math.MaxInt64 >> 1
	}
	return sum / count
}

// screenshotFramesAreSimilar 判定两帧是否相似：尺寸一致且网格平均 RGB 差异 <= 2。
// [S 汇编 0x1409a4140, 160B]：nil→false；Dx 或 Dy 不等→false；
// sampleScreenshotFrameAverageDiffWithGrid(a,Min.Y,b,Min.Y,Dy,42,80) <= 2。
func screenshotFramesAreSimilar(a, b *image.RGBA) bool {
	if a == nil || b == nil {
		return false
	}
	if a.Rect.Dx() != b.Rect.Dx() || a.Rect.Dy() != b.Rect.Dy() {
		return false
	}
	return sampleScreenshotFrameAverageDiffWithGrid(a, a.Rect.Min.Y, b, b.Rect.Min.Y, a.Rect.Dy(), 42, 80) <= 2
}

// rankScreenshotScrollingAppendCandidate 排名滚动追加候选（纯算术评分，无调用）。
// [S 汇编 0x1409a5e40, 128B]：score>=1<<50→0x3fffffffffffffff（@0x1409a5e4d）；
// score=score*1000+a*10（@0x1409a5e4f/56/5e）；d<=0||e<=0→score（@0x1409a5e64/69）；
// s=max(a+b,0)（@0x1409a5e6f/77）；diff=max(e-s,0)（@0x1409a5e7b/81）；delta=|diff-d|
// （@0x1409a5e88/8b/94）；score+=2*delta（@0x1409a5e98）。
func rankScreenshotScrollingAppendCandidate(a, b, score, d, e int64) int64 {
	if score >= 1<<50 {
		return 0x3fffffffffffffff
	}
	score = score*1000 + a*10
	if d <= 0 || e <= 0 {
		return score
	}
	s := a + b
	if s < 0 {
		s = 0
	}
	diff := e - s
	if diff < 0 {
		diff = 0
	}
	delta := diff - d
	if delta < 0 {
		delta = -delta
	}
	return score + 2*delta
}

// analyzeScreenshotScrollingFrameProgress 分析滚动帧进度（相似性 + 追加匹配）。
// [S-sig 0x1409a41e0, 128B]：screenshotFramesAreSimilar(a,b) → bool，交
// resolveScreenshotScrollingAppendMatchWithTarget(a,b,c)，返回 (结果, 相似 bool)。
// 体待 resolveScreenshotScrollingAppendMatchWithTarget 专项还原。
func analyzeScreenshotScrollingFrameProgress(a, b *image.RGBA, c int64) (int64, bool) {
	similar := screenshotFramesAreSimilar(a, b)
	_ = c
	return 0, similar
}

// stopScreenshotScrollingChunkedCanvasAndSave 停止滚动分块画布并保存。
// [S-sig 0x1409a0740, 128B]：尺寸 (w,h) <=0 则返回空结果，否则
// encodeScreenshotScrollingChunkedCanvas 编码。体待编码链专项还原。
func stopScreenshotScrollingChunkedCanvasAndSave(a, b *image.RGBA, c, d int64, e bool) ([]byte, error) {
	_, _, _, _, _ = a, b, c, d, e
	return nil, nil
}

// refineScreenshotScrollingChunkedAppendFromForSeam 优化滚动追加起点（接缝对齐）。
// [S-sig 0x14099ff00, 192B]：nil/尺寸非法/边界条件则透传 c；否则
// tailImage 取尾部，非空则 refineScreenshotScrollingAppendFromForSeam。体待 tailImage 专项。
func refineScreenshotScrollingChunkedAppendFromForSeam(a, b, c interface{}) interface{} {
	_, _, _ = a, b, c
	return c
}

// screenshotNativePreviewWindow.decodeImage 解码预览图像（透传 decode）。
// [S-sig 0x14099dda0, 192B]：duffcopy 参数结构 → screenshotNativePreviewDecodeImage。
// 体待解码链专项还原。
func (w *screenshotNativePreviewWindow) decodeImage(a interface{}) interface{} {
	_, _ = w, a
	return nil
}

// screenshotScrollingTopSkipLooksStable 判定滚动顶部跳过是否稳定（帧差 <=10）。
// [S-sig 0x1409a5d60, 224B]：nil/n<=0 守卫 → min(lenA,lenB,n,240) → 帧差采样
// sampleScreenshotFrameAverageDiffWithGrid(a,b,n,12,32) <= 0xa。
func screenshotScrollingTopSkipLooksStable(a, b interface{}, n int64) bool {
	_, _, _ = a, b, n
	return false
}

// screenshotPreviewSetLayeredWindowOpacity 设置预览窗口分层透明度。
// [S-sig 0x14099e6a0, 256B]：hwnd nil 守卫 → clamp opacity [0,1] → *255 截断 → SetLayeredWindowAttributes。
// 体待预览窗口 Win32 域专项还原。
func screenshotPreviewSetLayeredWindowOpacity(hwnd uintptr, opacity float64) {
	_, _ = hwnd, opacity
}

// waitScreenshotScrollingFrameReady 等待滚动帧就绪（超时循环 + 取消检查）。
// [S-sig 0x1409a3480, 256B]：shouldCancelScreenshotScrolling → deadline 循环 → time.Sleep(min(1s))。
// 体待滚动捕获域专项还原。
func waitScreenshotScrollingFrameReady(a, b interface{}) bool {
	_, _ = a, b
	return false
}

// resolveBottomTrim 解析滚动分块画布底部裁剪行数。
// [S-sig 0x14099fcc0, 256B]：自底向上扫描行 → screenshotScrollingRowLooksPureBlack 判黑。
// 体待滚动分块域专项还原。
func (c *screenshotScrollingChunkedCanvas) resolveBottomTrim() int {
	return 0
}

// captureScreenshotScrollingFrame 捕获滚动帧（接口方法转发）。
// [S-sig 0x1409a2d40, 224B]：backend nil → defaultScreenshotScreenCaptureBackend；
// 调用 backend[+0x18] 接口方法转发参数。
// 体待滚动捕获域专项还原。
func captureScreenshotScrollingFrame(a, b, c, d, e, f interface{}) interface{} {
	_, _, _, _, _, _ = a, b, c, d, e, f
	return nil
}

// buildScreenshotThumbnailPNGFromPath 从路径构建截图缩略图 PNG。
// [S-sig 0x14096b200, 256B]：读图 → 缩放 → PNG 编码。
// 体待缩略图域专项还原。
func buildScreenshotThumbnailPNGFromPath(a interface{}) (interface{}, error) {
	_ = a
	return nil, nil
}

// encodeScreenshotScrollingChunkedCanvas 编码滚动分块画布。
// [S-sig 0x1409a0640, 256B]：分块画布 → 编码输出。
// 体待滚动分块域专项还原。
func encodeScreenshotScrollingChunkedCanvas(a interface{}) interface{} {
	_ = a
	return nil
}
