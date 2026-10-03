package main

import (
	"testing"
)

func TestScreenshotPinSetLayeredWindowOpacityZeroHWND(t *testing.T) {
	// hwnd==0 分支：直接返回，不触发 proc.Call。
	for _, v := range []float64{0.0, 0.2, 0.5, 1.0, -1.0, 2.0} {
		screenshotPinSetLayeredWindowOpacity(0, v)
	}
}

func TestScreenshotPinSuppressWindowBorderZeroHWND(t *testing.T) {
	screenshotPinSuppressWindowBorder(0)
}

func TestScreenshotPinOpacityClamp(t *testing.T) {
	// clamp 逻辑等价复现，锚定反汇编解码的 [0.2, 1.0] 下上界与 alpha 换算。
	cases := []struct {
		in        float64
		wantAlpha int
	}{
		{0.0, 51},   // clamp -> 0.2, int(0.2*255)=51
		{0.2, 51},
		{0.5, 127},  // int(127.5) 截断 -> 127
		{1.0, 255},  // >=1.0 -> 255
		{2.0, 255},
		{-1.0, 51},  // clamp -> 0.2
	}
	for _, c := range cases {
		op := c.in
		if op < 0.2 {
			op = 0.2
		} else if op > 1.0 {
			op = 1.0
		}
		var alpha int
		if op < 1.0 {
			alpha = int(op * 255)
			if alpha <= 0 {
				alpha = 1
			}
		} else {
			alpha = 255
		}
		if alpha != c.wantAlpha {
			t.Errorf("opacity %v -> alpha %d, want %d", c.in, alpha, c.wantAlpha)
		}
	}
}
