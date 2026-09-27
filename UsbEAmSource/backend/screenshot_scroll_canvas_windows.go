package main

import "image"

// Bounds 返回截图滚动画布的矩形边界。
// [S] ASM 0x14099f520: c==nil || width<=0 || height<=0 → 零矩形；
// 否则 image.Rect(0, 0, width, height)。
func (c *screenshotScrollingChunkedCanvas) Bounds() image.Rectangle {
	if c == nil || c.width <= 0 || c.height <= 0 {
		return image.Rectangle{}
	}
	return image.Rect(0, 0, c.width, c.height)
}
