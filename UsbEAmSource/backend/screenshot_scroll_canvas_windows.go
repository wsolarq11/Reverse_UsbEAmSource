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

// releaseMemory 释放单个 chunk：先调用 release 回调（若存在）并置空，再清空 image.Pix。
// [S 汇编 0x14099f840, 160B]：nil→return；release!=nil→调用+置 nil；image!=nil→Pix 清零（+0/+8/+0x10）。
func (c *screenshotScrollingCanvasChunk) releaseMemory() {
	if c == nil {
		return
	}
	if c.release != nil {
		c.release()
		c.release = nil
	}
	if c.image != nil {
		c.image.Pix = nil
	}
}

// Release 释放画布全部 chunk。[S 汇编 0x14099f7c0, 128B]：nil→return；遍历 chunks 逐个 releaseMemory。
func (s *screenshotScrollingChunkedCanvas) Release() {
	if s == nil {
		return
	}
	for _, c := range s.chunks {
		c.releaseMemory()
	}
}

// ReleaseRow 释放覆盖 row 的 chunk（仅当 row 是该 chunk 的最后一行）。
// [S 汇编 0x14099f720, 160B]：遍历 chunks 找 image.Rect.Min.Y<=row<Max.Y 的 chunk；
// 命中且 row==Max.Y-1 时 releaseMemory。
func (s *screenshotScrollingChunkedCanvas) ReleaseRow(row int) {
	var chunk *screenshotScrollingCanvasChunk
	if s != nil {
		for _, c := range s.chunks {
			if c == nil || c.image == nil {
				continue
			}
			if c.image.Rect.Min.Y > row || row >= c.image.Rect.Max.Y {
				continue
			}
			chunk = c
			break
		}
	}
	if chunk == nil || chunk.image == nil {
		return
	}
	if row == chunk.image.Rect.Max.Y-1 {
		chunk.releaseMemory()
	}
}

// newScreenshotScrollingChunkedCanvas 构造滚动 chunked 画布（矩形校验 → 单 chunk）。
// [S-sig 0x14099f120, 416B]：rect 空/越界→error；面积>0x7270e00→error；
// newScreenshotScrollingCanvasChunk → 组装 canvas。体待 chunk 域专项还原。
func newScreenshotScrollingChunkedCanvas(a, b interface{}) interface{} {
	_, _ = a, b
	return nil
}
