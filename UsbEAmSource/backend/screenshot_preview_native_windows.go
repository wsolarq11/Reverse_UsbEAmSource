// AUTO-RECONSTRUCTED — DOMAIN: screenshot native preview window 基础访问器
// 研究用途。反汇编实证：
//
//	setHandle  0x14099db40, 160B（lock 加锁 → [rdx+0x18]=h → 解锁）
//	setThreadID 0x14099dd00, 160B（lock 加锁 → [rdx+0x20]=id → 解锁）
//	handle     0x14099dbe0, 192B（lock.Lock + defer Unlock → 读 [rcx+0x18]）
//
// 结构 screenshotNativePreviewWindow：types_screenshot.go（lock @+0x10、hwnd @+0x18、
// threadID @+0x20）。
package main

import (
	"errors"
	"image"
	"math"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/w32"
)

// 自定义窗口消息（WM_APP 偏移）：显示 0x86a1、隐藏 0x86a2、关闭 0x86a3、透明度 0x86a5。
const (
	screenshotNativePreviewShowMessage    = 0x86a1
	screenshotNativePreviewHideMessage    = 0x86a2
	screenshotNativePreviewCloseMessage   = 0x86a3
	screenshotNativePreviewOpacityMessage = 0x86a5
)

// setHandle 加锁写入窗口句柄。[S 汇编 0x14099db40, 160B]：cmpxchg [rdx+0x10] 加锁 →
// [rdx+0x18]=h → xadd -1 解锁。
func (s *screenshotNativePreviewWindow) setHandle(h uintptr) {
	s.lock.Lock()
	s.hwnd = h
	s.lock.Unlock()
}

// setThreadID 加锁写入窗口线程 ID。[S 汇编 0x14099dd00, 160B]：cmpxchg [rdx+0x10] 加锁 →
// [rdx+0x20]=id → xadd -1 解锁。
func (s *screenshotNativePreviewWindow) setThreadID(id uintptr) {
	s.lock.Lock()
	s.threadID = id
	s.lock.Unlock()
}

// handle 加锁读取窗口句柄。[S 汇编 0x14099dbe0, 192B]：cmpxchg 加锁 → defer Unlock →
// 读 [rcx+0x18] 返回。
func (s *screenshotNativePreviewWindow) handle() uintptr {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.hwnd
}

// Hide 异步隐藏窗口：PostMessage 隐藏消息（0x86a2）。无句柄时直接返回 nil。
// [S 汇编 0x14099cf60, 160B]：handle()==0→(nil,nil)；否则 newobject 打包 {hwnd,0x86a2,0,0}
// → PostMessageW LazyProc.Call(4 arg) → 返回 nil error。
func (s *screenshotNativePreviewWindow) Hide() error {
	hwnd := s.handle()
	if hwnd == 0 {
		return nil
	}
	w32.PostMessage(hwnd, screenshotNativePreviewHideMessage, 0, 0)
	return nil
}

// Show 异步显示窗口：PostMessage 显示消息（0x86a1）。无句柄时返回错误"原生截图预览窗口不可用"。
// [S 汇编 0x14099cea0, 192B]：handle()==0→errors.New(33B 字符串)；否则 PostMessageW(hwnd,0x86a1,0,0)
// → 返回 nil error。
func (s *screenshotNativePreviewWindow) Show() error {
	hwnd := s.handle()
	if hwnd == 0 {
		return errors.New("原生截图预览窗口不可用")
	}
	w32.PostMessage(hwnd, screenshotNativePreviewShowMessage, 0, 0)
	return nil
}

// Close 幂等关闭窗口：closeOnce 内 PostMessage 关闭消息（0x86a3）。无句柄时直接返回 nil。
// [S 汇编 0x14099d000, 128B]：handle()==0→nil；否则 closeOnce(+0xdc).Do(func1{PostMessage(hwnd,0x86a3)}) → nil。
func (s *screenshotNativePreviewWindow) Close() error {
	hwnd := s.handle()
	if hwnd == 0 {
		return nil
	}
	s.closeOnce.Do(func() {
		w32.PostMessage(hwnd, screenshotNativePreviewCloseMessage, 0, 0)
	})
	return nil
}

// hideOnThread 在窗口线程把窗口移出屏幕（X+100000）并清 visible 标志。
// [S 汇编 0x14099d320, 128B]：handle()==0→return；SetWindowPos(hwnd,TOPMOST,X+100000,Y,W,H,0x50)
// → visible(+0xd8)=false。
func (s *screenshotNativePreviewWindow) hideOnThread() {
	hwnd := s.handle()
	if hwnd == 0 {
		return
	}
	w32.SetWindowPos(hwnd, w32.HWND_TOPMOST,
		s.bounds.X+100000, s.bounds.Y, s.bounds.Width, s.bounds.Height,
		w32.SWP_NOACTIVATE|w32.SWP_SHOWWINDOW)
	s.visible = false
}

// SetOpacity 设置预览窗口透明度（clamp 到 [0,1]），经 PostMessage 0x86a5 携带指针通知窗口线程。
// [S 汇编 0x14099d100, 288B]：handle()==0→errors.New("原生截图预览窗口不可用")；
// 否则 <0→0、>1→1，PostMessageW(hwnd,0x86a5,0,&opacity) → nil。
func (s *screenshotNativePreviewWindow) SetOpacity(opacity float64) error {
	hwnd := s.handle()
	if hwnd == 0 {
		return errors.New("原生截图预览窗口不可用")
	}
	if opacity < 0 {
		opacity = 0
	} else if opacity > 1 {
		opacity = 1
	}
	w32.PostMessage(hwnd, screenshotNativePreviewOpacityMessage, 0, uintptr(unsafe.Pointer(&opacity)))
	return nil
}

// showOnThread 在窗口线程显示窗口：SetWindowPos(0x53) → ShowWindow(SW_SHOWNOACTIVATE) →
// visible=true → redraw → UpdateWindow。
// [S 汇编 0x14099d220, 256B]：handle()==0→return；SetWindowPos(hwnd,TOPMOST,bounds,0x53)；
// ShowWindow(hwnd,4)；visible(+0xd8)=true；redraw()；UpdateWindow(hwnd)。
func (s *screenshotNativePreviewWindow) showOnThread() {
	hwnd := s.handle()
	if hwnd == 0 {
		return
	}
	w32.SetWindowPos(hwnd, w32.HWND_TOPMOST,
		s.bounds.X, s.bounds.Y, s.bounds.Width, s.bounds.Height,
		w32.SWP_SHOWWINDOW|w32.SWP_NOACTIVATE|w32.SWP_NOMOVE|w32.SWP_NOSIZE)
	w32.ShowWindow(hwnd, w32.SW_SHOWNOACTIVATE)
	s.visible = true
	s.redraw()
	w32.UpdateWindow(hwnd)
}

// redraw 重绘预览窗口内容。[S-sig 0x14099d3a0, 1152B] 待专项还原（依赖 image 渲染链）。
func (s *screenshotNativePreviewWindow) redraw() {
	_ = s
}

// screenshotNativePreviewImageRect 计算图像在 maxW×maxH 内等比居中缩放的矩形。
// [S 汇编 0x14099e580, 288B]：img==nil||maxW<=0||maxH<=0||img 尺寸<=0 → 零矩形；
// scale=min(maxW/imgW,maxH/imgH)<=0 → 零矩形；w/h=max(1,int(imgW/H*scale))；
// image.Rect((maxW-w)/2,(maxH-h)/2, +w,+h)。
func screenshotNativePreviewImageRect(img *image.RGBA, maxW, maxH int) image.Rectangle {
	if img == nil || maxW <= 0 || maxH <= 0 {
		return image.Rectangle{}
	}
	iw := img.Rect.Dx()
	ih := img.Rect.Dy()
	if iw <= 0 || ih <= 0 {
		return image.Rectangle{}
	}
	scale := math.Min(float64(maxW)/float64(iw), float64(maxH)/float64(ih))
	if scale <= 0 {
		return image.Rectangle{}
	}
	w := int(float64(iw) * scale)
	if w < 1 {
		w = 1
	}
	h := int(float64(ih) * scale)
	if h < 1 {
		h = 1
	}
	return image.Rect((maxW-w)/2, (maxH-h)/2, (maxW-w)/2+w, (maxH-h)/2+h)
}
