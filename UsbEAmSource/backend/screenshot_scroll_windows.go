package main

import (
	"image"
	"math"
)

// screenshot 滚动所需的 user32 LazyProc（依赖 appicon_windows.go 中的 user32DLL）。
var procGetAsyncKeyState = user32DLL.NewProc("GetAsyncKeyState")

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
