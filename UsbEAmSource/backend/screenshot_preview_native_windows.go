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

import "github.com/wailsapp/wails/v3/pkg/w32"

// 自定义窗口消息（WM_APP 偏移）：隐藏 0x86a2、关闭 0x86a3。
const (
	screenshotNativePreviewHideMessage  = 0x86a2
	screenshotNativePreviewCloseMessage = 0x86a3
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
