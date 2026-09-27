package main

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"sort"
	"strings"
	"sync"
	"time"
)

// 本文件承载 DXGI 输出复制缓存域（批次 54）：
// 缓存/缓存项方法族 + 缓存键 + 超时哨兵 + 三个入口 + 纯 Go 像素克隆。
// 深层 D3D11 / 桌面复制 / tone-map / 旋转 / 纹理换算体见
// screenshot_dxgi_deep_windows.go（[S-sig] 桩，批次 55）。

// errScreenshotDXGICaptureFrameTimeout 是桌面复制 AcquireNextFrame 超时哨兵。
// [S] 全局错误 iface 0x141bc3d90 → errorString "DXGI 捕获帧超时"（20B）。
var errScreenshotDXGICaptureFrameTimeout = errors.New("DXGI 捕获帧超时")

// screenshotDXGICaptureCacheReconcileStaleness 是缓存项判定过期的阈值。
// [S] asm 立即数 0x45d964b800 = 300_000_000_000 ns = 5 分钟。
const screenshotDXGICaptureCacheReconcileStaleness = 5 * time.Minute

// screenshotDXGICaptureCacheReconcileMaxEntries 是缓存项保留上限。
// [S] asm Reconcile 中常量 8。
const screenshotDXGICaptureCacheReconcileMaxEntries = 8

// screenshotDXGIOutputCaptureCache 是进程级 DXGI 输出捕获缓存。
// [S] 布局：mu(0x00) entries(0x08) closed(0x10)。
type screenshotDXGIOutputCaptureCache struct {
	mu      sync.Mutex
	entries map[string]*screenshotDXGIOutputCaptureEntry
	closed  bool
}

// screenshotDXGIOutputCaptureCacheGlobal 是包级单例（Shutdown 以全局符号取 receiver）。
// [S] asm 0x140773e8c 以 [rip+0x144de05] 取此全局（VA 0x140bc1c98）；
// 类型名与方法族同名，全局单例以 Global 后缀区分。
var screenshotDXGIOutputCaptureCacheGlobal = &screenshotDXGIOutputCaptureCache{}

// screenshotDXGIOutputCaptureCacheKey 生成缓存键。
// [S] ASM 0x140979820: "%d:%d:%s:%t:%d:%d:%d:%d:%d:%d" =
// AdapterIndex, OutputIndex, TrimSpace(DeviceName), HDR, ColorSpace, Rotation,
// Bounds.Min.X, Bounds.Min.Y, Bounds.Dx, Bounds.Dy。
func screenshotDXGIOutputCaptureCacheKey(d screenshotDisplayCaptureInfo) string {
	return fmt.Sprintf("%d:%d:%s:%t:%d:%d:%d:%d:%d:%d",
		d.AdapterIndex, d.OutputIndex, strings.TrimSpace(d.DeviceName),
		d.HDR, d.ColorSpace, d.Rotation,
		d.Bounds.Min.X, d.Bounds.Min.Y, d.Bounds.Dx(), d.Bounds.Dy())
}

// prepare 预热指定显示器（惰性创建缓存项）。
// [S] ASM 0x1409768e0: nil cache → "DXGI 输出复制缓存未初始化"；否则
// entryForDisplay(d) 仅取 error。
func (c *screenshotDXGIOutputCaptureCache) prepare(d screenshotDisplayCaptureInfo) error {
	if c == nil {
		return errors.New("DXGI 输出复制缓存未初始化")
	}
	_, err := c.entryForDisplay(d)
	return err
}

// entryForDisplay 返回（或惰性创建）指定显示器的缓存项。
// [S] ASM 0x140977020: nil cache → "DXGI 输出复制缓存未初始化"；
// rotation 0/>4 → "不支持的 DXGI 显示器旋转值 %d"；closed →
// "DXGI 输出复制缓存已关闭"；lazy makemap；mapaccess 命中返回；
// 否则 createEntryLocked。
func (c *screenshotDXGIOutputCaptureCache) entryForDisplay(d screenshotDisplayCaptureInfo) (*screenshotDXGIOutputCaptureEntry, error) {
	if c == nil {
		return nil, errors.New("DXGI 输出复制缓存未初始化")
	}
	if d.Rotation == 0 || d.Rotation > 4 {
		return nil, fmt.Errorf("不支持的 DXGI 显示器旋转值 %d", d.Rotation)
	}
	key := screenshotDXGIOutputCaptureCacheKey(d)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil, errors.New("DXGI 输出复制缓存已关闭")
	}
	if c.entries == nil {
		c.entries = make(map[string]*screenshotDXGIOutputCaptureEntry)
	}
	if e, ok := c.entries[key]; ok {
		return e, nil
	}
	return c.createEntryLocked(d)
}

// createEntryLocked 在持锁状态下创建缓存项。
// [S] ASM 0x140977480: closed → "DXGI 输出复制缓存已关闭"；key →
// openScreenshotDXGIOutputTarget → createScreenshotD3D11Device →
// duplicateScreenshotDXGIOutput → tone-map → label → 日志 → 建 entry →
// mapassign → 日志 → 返回；失败按 defer release 释放。
func (c *screenshotDXGIOutputCaptureCache) createEntryLocked(d screenshotDisplayCaptureInfo) (*screenshotDXGIOutputCaptureEntry, error) {
	if c.closed {
		return nil, errors.New("DXGI 输出复制缓存已关闭")
	}
	key := screenshotDXGIOutputCaptureCacheKey(d)
	output, output6, release, err := openScreenshotDXGIOutputTarget(d)
	if err != nil {
		return nil, err
	}
	defer release()
	device, err := createScreenshotD3D11Device(output)
	if err != nil {
		return nil, err
	}
	duplication, err := duplicateScreenshotDXGIOutput(output6, device.device, d.HDR)
	if err != nil {
		device.release()
		return nil, err
	}
	toneMap := screenshotHDRToneMapOptionsForDisplayWithSDRWhiteResolver(d, screenshotDisplayConfigResolverGlobal)
	label := screenshotDisplayCaptureLabel(d)
	screenshotHDRCaptureDebugLog("DXGI HDR tone mapping: display=%s options=%s", label, formatScreenshotHDRToneMapOptionsForDebug(toneMap))
	e := &screenshotDXGIOutputCaptureEntry{
		key:            key,
		displayLabel:   label,
		device:         device,
		duplication:    duplication,
		toneMapOptions: toneMap,
		lastUsed:       time.Now(),
	}
	c.entries[key] = e
	screenshotHDRCaptureDebugLog("DXGI 输出复制对象已缓存: display=%s", e.displayLabel)
	return e, nil
}

// remove 在 expected 匹配时删除缓存项并释放。
// [S] ASM 0x140977ae0: nil cache / 空 key → 返回；mapaccess 未命中 → 返回；
// expected 非 nil 且不等 → 返回；mapdelete；解锁后 entry.release()。
func (c *screenshotDXGIOutputCaptureCache) remove(key string, expected *screenshotDXGIOutputCaptureEntry) {
	if c == nil || len(key) == 0 {
		return
	}
	c.mu.Lock()
	e, ok := c.entries[key]
	if !ok {
		c.mu.Unlock()
		return
	}
	if expected != nil && e != expected {
		c.mu.Unlock()
		return
	}
	delete(c.entries, key)
	c.mu.Unlock()
	e.release()
}

// capture 缓存整帧捕获：失败且非超时哨兵时重建输出复制对象重试。
// [S] ASM 0x140976980: nil cache → "DXGI 输出复制缓存未初始化"；entryForDisplay；
// e.capture；超时哨兵直接返回；否则日志 + remove + 重建 entryForDisplay；
// 重建失败 → "DXGI 捕获失败: %v；重建输出复制对象失败: %w"；成功则 e2.capture。
func (c *screenshotDXGIOutputCaptureCache) capture(d screenshotDisplayCaptureInfo, requireFresh bool) (*image.RGBA, error) {
	if c == nil {
		return nil, errors.New("DXGI 输出复制缓存未初始化")
	}
	e, err := c.entryForDisplay(d)
	if err != nil {
		return nil, err
	}
	img, err := e.capture(requireFresh)
	if err == nil {
		return img, nil
	}
	if errors.Is(err, errScreenshotDXGICaptureFrameTimeout) {
		return nil, err
	}
	screenshotHDRCaptureDebugLog("DXGI 缓存捕获失败，重建输出复制对象: display=%s err=%v", screenshotDisplayCaptureLabel(d), err)
	c.remove(e.key, e)
	e2, err2 := c.entryForDisplay(d)
	if err2 != nil {
		return nil, fmt.Errorf("DXGI 捕获失败: %v；重建输出复制对象失败: %w", err, err2)
	}
	return e2.capture(requireFresh)
}

// captureRegion 缓存区域捕获：失败且非超时哨兵时重建输出复制对象重试。
// [S] ASM 0x140976c60: 与 capture 对称，区域日志与重建失败文案不同。
func (c *screenshotDXGIOutputCaptureCache) captureRegion(d screenshotDisplayCaptureInfo, rect image.Rectangle, requireFresh bool) (*image.RGBA, error) {
	if c == nil {
		return nil, errors.New("DXGI 输出复制缓存未初始化")
	}
	e, err := c.entryForDisplay(d)
	if err != nil {
		return nil, err
	}
	img, err := e.captureRegion(rect, requireFresh)
	if err == nil {
		return img, nil
	}
	if errors.Is(err, errScreenshotDXGICaptureFrameTimeout) {
		return nil, err
	}
	screenshotHDRCaptureDebugLog("DXGI 缓存区域捕获失败，重建输出复制对象: display=%s rect=%v err=%v", screenshotDisplayCaptureLabel(d), rect, err)
	c.remove(e.key, e)
	e2, err2 := c.entryForDisplay(d)
	if err2 != nil {
		return nil, fmt.Errorf("DXGI 区域捕获失败: %v；重建输出复制对象失败: %w", err, err2)
	}
	return e2.captureRegion(rect, requireFresh)
}

// Reconcile 修剪缓存：删除不在 wanted 集合或过期（>5min 未用）的项，
// 超过 8 项时按 lastUsed 升序裁剪到 8 项；被删项在解锁后统一释放。
// [S] ASM 0x140977c40。
func (c *screenshotDXGIOutputCaptureCache) Reconcile(displays []screenshotDisplayCaptureInfo) {
	if c == nil {
		return
	}
	wanted := make(map[string]struct{}, len(displays))
	for i := range displays {
		d := displays[i]
		if !d.AttachedToDesktop || d.Bounds.Empty() {
			continue
		}
		if d.Rotation == 0 || d.Rotation > 4 {
			continue
		}
		wanted[screenshotDXGIOutputCaptureCacheKey(d)] = struct{}{}
	}

	now := time.Now()
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	var collected []*screenshotDXGIOutputCaptureEntry
	for key, e := range c.entries {
		if _, ok := wanted[key]; ok {
			if now.Sub(e.lastUsedSnapshot()) <= screenshotDXGICaptureCacheReconcileStaleness {
				continue
			}
		}
		delete(c.entries, key)
		collected = append(collected, e)
	}

	if len(c.entries) > screenshotDXGICaptureCacheReconcileMaxEntries {
		type agedEntry struct {
			key      string
			entry    *screenshotDXGIOutputCaptureEntry
			lastUsed time.Time
		}
		aged := make([]agedEntry, 0, len(c.entries))
		for key, e := range c.entries {
			aged = append(aged, agedEntry{key: key, entry: e, lastUsed: e.lastUsedSnapshot()})
		}
		sort.Slice(aged, func(i, j int) bool {
			return aged[i].lastUsed.Before(aged[j].lastUsed)
		})
		for i := 0; len(c.entries) > screenshotDXGICaptureCacheReconcileMaxEntries && i < len(aged); i++ {
			delete(c.entries, aged[i].key)
			collected = append(collected, aged[i].entry)
		}
	}
	c.mu.Unlock()

	for _, e := range collected {
		e.release()
	}
}

// Close 关闭缓存：置 closed，清空 map，解锁后释放全部缓存项。
// [S] ASM 0x140978780。
func (c *screenshotDXGIOutputCaptureCache) Close() {
	if c == nil {
		return
	}
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	entries := make([]*screenshotDXGIOutputCaptureEntry, 0, len(c.entries))
	for key, e := range c.entries {
		entries = append(entries, e)
		delete(c.entries, key)
	}
	c.mu.Unlock()
	for _, e := range entries {
		e.release()
	}
}

// capture 捕获缓存项整帧（可选强制刷新），超时且可复用时回退上一帧。
// [S] ASM 0x140978a80。
func (e *screenshotDXGIOutputCaptureEntry) capture(requireFresh bool) (*image.RGBA, error) {
	if e == nil {
		return nil, errors.New("DXGI 输出复制缓存项为空")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed || e.device == nil || e.duplication == nil {
		return nil, errors.New("DXGI 输出复制缓存项已关闭")
	}
	e.lastUsed = time.Now()
	img, err := captureScreenshotDXGIFrameRegion(e.device, uintptr(e.duplication), e.toneMapOptions, image.Rectangle{})
	if err != nil {
		if errors.Is(err, errScreenshotDXGICaptureFrameTimeout) && e.lastFrame != nil && !requireFresh {
			screenshotHDRCaptureDebugLog("DXGI 本次无桌面更新，复用上一帧: display=%s", e.displayLabel)
			return cloneRGBAImage(e.lastFrame), nil
		}
		return nil, err
	}
	e.updateLastFrameLocked(img)
	return img, nil
}

// captureRegion 捕获缓存项指定区域，超时且可复用时回退上一帧区域。
// [S] ASM 0x140978ea0: 成功路径不更新 lastFrame（区域捕获不缓存整帧）。
func (e *screenshotDXGIOutputCaptureEntry) captureRegion(rect image.Rectangle, requireFresh bool) (*image.RGBA, error) {
	if e == nil {
		return nil, errors.New("DXGI 输出复制缓存项为空")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed || e.device == nil || e.duplication == nil {
		return nil, errors.New("DXGI 输出复制缓存项已关闭")
	}
	e.lastUsed = time.Now()
	img, err := captureScreenshotDXGIFrameRegion(e.device, uintptr(e.duplication), e.toneMapOptions, rect)
	if err != nil {
		if errors.Is(err, errScreenshotDXGICaptureFrameTimeout) && e.lastFrame != nil && !requireFresh {
			screenshotHDRCaptureDebugLog("DXGI 区域捕获本次无桌面更新，复用上一帧: display=%s rect=%v", e.displayLabel, rect)
			return cloneRGBARegion(e.lastFrame, rect), nil
		}
		return nil, err
	}
	return img, nil
}

// release 释放缓存项：置 closed，释放桌面复制对象、设备、lastFrame 预算。
// [S] ASM 0x140979360。
func (e *screenshotDXGIOutputCaptureEntry) release() {
	if e == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return
	}
	e.closed = true
	if e.duplication != nil {
		screenshotCOMRelease(uintptr(e.duplication))
		e.duplication = nil
	}
	if e.device != nil {
		e.device.release()
		e.device = nil
	}
	e.lastFrame = nil
	if e.lastFrameFree != nil {
		e.lastFrameFree()
		e.lastFrameFree = nil
	}
}

// updateLastFrameLocked 在持锁状态下更新缓存的上一次整帧。
// [S] ASM 0x140979560: img 空/Pix 空 → 清空 lastFrame 并释放预算；
// 否则 Reserve(len(Pix))，失败返回；克隆后替换 lastFrame/lastFrameFree。
func (e *screenshotDXGIOutputCaptureEntry) updateLastFrameLocked(img *image.RGBA) {
	if img == nil || len(img.Pix) == 0 {
		e.lastFrame = nil
		if e.lastFrameFree != nil {
			e.lastFrameFree()
			e.lastFrameFree = nil
		}
		return
	}
	free, err := screenshotImageMemoryBudgetGlobal.Reserve(int64(len(img.Pix)))
	if err != nil {
		return
	}
	clone := cloneRGBAImage(img)
	if e.lastFrameFree != nil {
		e.lastFrameFree()
	}
	e.lastFrame = clone
	e.lastFrameFree = free
}

// lastUsedSnapshot 返回缓存项最近使用时间快照。
// [S] ASM 0x1409796a0: nil → 零 time.Time；否则锁内拷贝 lastUsed。
func (e *screenshotDXGIOutputCaptureEntry) lastUsedSnapshot() time.Time {
	if e == nil {
		return time.Time{}
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.lastUsed
}

// cloneRGBAImage 深拷贝整张 RGBA 图像。
// [S] ASM 0x140957320: nil → nil；NewRGBA(img.Rect)；copy(dst.Pix, img.Pix)。
func cloneRGBAImage(img *image.RGBA) *image.RGBA {
	if img == nil {
		return nil
	}
	dst := image.NewRGBA(img.Rect)
	copy(dst.Pix, img.Pix)
	return dst
}

// cloneRGBARegion 深拷贝 RGBA 图像中与 rect 相交的区域。
// [S] ASM 0x1409573a0: img nil 或 rect 空 → NewRGBA(0,0,0,0)；
// inter=rect∩img.Rect；NewRGBA(0,0,inter.Dx,inter.Dy)；
// DrawMask(dst, dst.Rect, img, inter.Min, nil, zero, Src)。
func cloneRGBARegion(img *image.RGBA, rect image.Rectangle) *image.RGBA {
	if img == nil || rect.Empty() {
		return image.NewRGBA(image.Rect(0, 0, 0, 0))
	}
	inter := rect.Intersect(img.Rect)
	dst := image.NewRGBA(image.Rect(0, 0, inter.Dx(), inter.Dy()))
	draw.DrawMask(dst, dst.Rect, img, inter.Min, nil, image.Point{}, draw.Src)
	return dst
}

// captureScreenshotDXGIOutputFrameWithOptions 捕获单显示器整帧（可选强制刷新）。
// [S] ASM 0x140976020: 未附着/空边界 → "DXGI 输出未连接到桌面"；缓存捕获；
// orient；尺寸不匹配 → "DXGI 旋转后帧尺寸不匹配: frame=%v display=%v rotation=%d"。
func captureScreenshotDXGIOutputFrameWithOptions(d screenshotDisplayCaptureInfo, requireFresh bool) (*image.RGBA, error) {
	if !d.AttachedToDesktop || d.Bounds.Empty() {
		return nil, errors.New("DXGI 输出未连接到桌面")
	}
	img, err := screenshotDXGIOutputCaptureCacheGlobal.capture(d, requireFresh)
	if err != nil {
		return nil, err
	}
	img, err = orientScreenshotDXGIFrame(img, d.Rotation)
	if err != nil {
		return nil, err
	}
	if img.Rect.Dx() != d.Bounds.Dx() || img.Rect.Dy() != d.Bounds.Dy() {
		return nil, fmt.Errorf("DXGI 旋转后帧尺寸不匹配: frame=%v display=%v rotation=%d", img.Rect, d.Bounds, d.Rotation)
	}
	return img, nil
}

// captureScreenshotDXGIOutputFrameRegionWithOptions 捕获单显示器区域帧。
// [S] ASM 0x140976260: 未附着/空边界 → "DXGI 输出未连接到桌面"；rect 空 →
// "DXGI 输出区域为空"；rect 越界 → "DXGI 输出区域越界: %v"；纹理矩形换算；
// 缓存区域捕获；orient；尺寸不匹配 → "DXGI 旋转后区域尺寸不匹配: frame=%v rect=%v"。
func captureScreenshotDXGIOutputFrameRegionWithOptions(d screenshotDisplayCaptureInfo, rect image.Rectangle, requireFresh bool) (*image.RGBA, error) {
	if !d.AttachedToDesktop || d.Bounds.Empty() {
		return nil, errors.New("DXGI 输出未连接到桌面")
	}
	if rect.Empty() {
		return nil, errors.New("DXGI 输出区域为空")
	}
	if rect.Intersect(image.Rect(0, 0, d.Bounds.Dx(), d.Bounds.Dy())) != rect {
		return nil, fmt.Errorf("DXGI 输出区域越界: %v", rect)
	}
	textureRect, err := screenshotDXGITextureRectForDesktopRect(d.Rotation, d.Bounds.Dx(), d.Bounds.Dy(), rect)
	if err != nil {
		return nil, err
	}
	img, err := screenshotDXGIOutputCaptureCacheGlobal.captureRegion(d, textureRect, requireFresh)
	if err != nil {
		return nil, err
	}
	img, err = orientScreenshotDXGIFrame(img, d.Rotation)
	if err != nil {
		return nil, err
	}
	if img.Rect.Dx() != rect.Dx() || img.Rect.Dy() != rect.Dy() {
		return nil, fmt.Errorf("DXGI 旋转后区域尺寸不匹配: frame=%v rect=%v", img.Rect, rect)
	}
	return img, nil
}

// prewarmScreenshotDXGIOutputFrameCache 预热输出帧捕获缓存。
// [S] ASM 0x1409766c0: Reconcile(displays)；对每个附着且非空且 HDR 的显示器
// prepare；失败 → "DXGI 输出复制对象预热失败: display=%s err=%v"。
func prewarmScreenshotDXGIOutputFrameCache(displays []screenshotDisplayCaptureInfo) {
	screenshotDXGIOutputCaptureCacheGlobal.Reconcile(displays)
	for i := range displays {
		d := displays[i]
		if !d.AttachedToDesktop || d.Bounds.Empty() || !d.HDR {
			continue
		}
		if err := screenshotDXGIOutputCaptureCacheGlobal.prepare(d); err != nil {
			screenshotHDRCaptureDebugLog("DXGI 输出复制对象预热失败: display=%s err=%v", screenshotDisplayCaptureLabel(d), err)
		}
	}
}
