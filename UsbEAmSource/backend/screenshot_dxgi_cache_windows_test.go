package main

import (
	"image"
	"testing"
)

func TestCloneRGBAImageNil(t *testing.T) {
	if got := cloneRGBAImage(nil); got != nil {
		t.Fatalf("cloneRGBAImage(nil) = %v, want nil", got)
	}
}

func TestCloneRGBAImageDeepCopy(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 2, 1))
	src.Pix[0], src.Pix[1], src.Pix[2], src.Pix[3] = 1, 2, 3, 4
	dst := cloneRGBAImage(src)
	if dst == nil {
		t.Fatal("cloneRGBAImage returned nil")
	}
	if &dst.Pix[0] == &src.Pix[0] {
		t.Fatal("clone shares Pix backing array")
	}
	if dst.Rect != src.Rect {
		t.Fatalf("rect = %v, want %v", dst.Rect, src.Rect)
	}
	dst.Pix[0] = 99
	if src.Pix[0] != 1 {
		t.Fatalf("src.Pix[0] = %d, want 1 (aliased)", src.Pix[0])
	}
}

func TestCloneRGBARegionEmpty(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 4, 4))
	got := cloneRGBARegion(src, image.Rect(0, 0, 0, 0))
	if got.Rect != image.Rect(0, 0, 0, 0) {
		t.Fatalf("rect = %v, want empty", got.Rect)
	}
}

func TestCloneRGBARegionClip(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for i := range src.Pix {
		src.Pix[i] = 1
	}
	src.Pix[20] = 7 // src[1,1] = 1*16 + 1*4 = 20
	got := cloneRGBARegion(src, image.Rect(1, 1, 3, 3))
	if got.Rect != image.Rect(0, 0, 2, 2) {
		t.Fatalf("rect = %v, want (0,0)-(2,2)", got.Rect)
	}
	if got.Stride != 8 {
		t.Fatalf("stride = %d, want 8", got.Stride)
	}
	if got.Pix[0] != 7 {
		t.Fatalf("dst[0,0] = %d, want 7 (src[1,1])", got.Pix[0])
	}
}

func TestCloneRGBARegionOutOfBounds(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 4, 4))
	got := cloneRGBARegion(src, image.Rect(-1, -1, 2, 2))
	if got.Rect != image.Rect(0, 0, 2, 2) {
		t.Fatalf("rect = %v, want (0,0)-(2,2)", got.Rect)
	}
}

func TestScreenshotDXGIOutputCaptureCacheKey(t *testing.T) {
	d := screenshotDisplayCaptureInfo{
		AdapterIndex: 1,
		OutputIndex:  2,
		DeviceName:   " display ",
		HDR:          true,
		ColorSpace:   3,
		Rotation:     4,
		Bounds:       image.Rect(10, 20, 40, 50),
	}
	got := screenshotDXGIOutputCaptureCacheKey(d)
	want := "1:2:display:true:3:4:10:20:30:30"
	if got != want {
		t.Fatalf("key = %q, want %q", got, want)
	}
}

func TestCaptureOutputFrameNotAttached(t *testing.T) {
	d := screenshotDisplayCaptureInfo{}
	if _, err := captureScreenshotDXGIOutputFrameWithOptions(d, false); err == nil || err.Error() != "DXGI 输出未连接到桌面" {
		t.Fatalf("err = %v, want 未连接", err)
	}
}

func TestCaptureOutputFrameRegionBounds(t *testing.T) {
	d := screenshotDisplayCaptureInfo{
		AttachedToDesktop: true,
		Bounds:            image.Rect(0, 0, 100, 100),
	}
	if _, err := captureScreenshotDXGIOutputFrameRegionWithOptions(d, image.Rect(50, 50, 150, 150), false); err == nil || err.Error() != "DXGI 输出区域越界: (50,50)-(150,150)" {
		t.Fatalf("err = %v, want 越界", err)
	}
	if _, err := captureScreenshotDXGIOutputFrameRegionWithOptions(d, image.Rect(0, 0, 0, 0), false); err == nil || err.Error() != "DXGI 输出区域为空" {
		t.Fatalf("err = %v, want 空", err)
	}
}

func TestCacheLifecycleNilSafe(t *testing.T) {
	var c *screenshotDXGIOutputCaptureCache
	c.Close()
	c.Reconcile(nil)
	c.remove("", nil)
	if err := c.prepare(screenshotDisplayCaptureInfo{}); err == nil || err.Error() != "DXGI 输出复制缓存未初始化" {
		t.Fatalf("prepare err = %v", err)
	}
	if _, err := c.capture(screenshotDisplayCaptureInfo{}, false); err == nil || err.Error() != "DXGI 输出复制缓存未初始化" {
		t.Fatalf("capture err = %v", err)
	}
}

func TestEntryNilSafe(t *testing.T) {
	var e *screenshotDXGIOutputCaptureEntry
	if _, err := e.capture(false); err == nil || err.Error() != "DXGI 输出复制缓存项为空" {
		t.Fatalf("capture err = %v", err)
	}
	if _, err := e.captureRegion(image.Rect(0, 0, 1, 1), false); err == nil || err.Error() != "DXGI 输出复制缓存项为空" {
		t.Fatalf("captureRegion err = %v", err)
	}
	e.release()
	if !e.lastUsedSnapshot().IsZero() {
		t.Fatalf("nil entry lastUsed should be zero")
	}
}

func TestEntryClosedGuard(t *testing.T) {
	e := &screenshotDXGIOutputCaptureEntry{closed: true}
	if _, err := e.capture(false); err == nil || err.Error() != "DXGI 输出复制缓存项已关闭" {
		t.Fatalf("capture err = %v, want 已关闭", err)
	}
	if _, err := e.captureRegion(image.Rect(0, 0, 1, 1), false); err == nil || err.Error() != "DXGI 输出复制缓存项已关闭" {
		t.Fatalf("captureRegion err = %v", err)
	}
}

func TestOrientationRotation1Identity(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	got, err := orientScreenshotDXGIFrame(img, 1)
	if err != nil || got != img {
		t.Fatalf("rotation 1 should be identity")
	}
	if _, err := orientScreenshotDXGIFrame(img, 0); err == nil || err.Error() != "不支持的 DXGI 显示器旋转值 0" {
		t.Fatalf("rotation 0 err = %v", err)
	}
	if _, err := orientScreenshotDXGIFrame(img, 5); err == nil {
		t.Fatalf("rotation 5 should error")
	}
	if _, err := orientScreenshotDXGIFrame(nil, 1); err == nil || err.Error() != "DXGI 捕获帧为空" {
		t.Fatalf("nil img err = %v", err)
	}
}
