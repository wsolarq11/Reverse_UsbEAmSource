// AUTO-RECONSTRUCTED — DOMAIN: desktop widget scheduler 唤醒
// 研究用途。反汇编实证：
//
//	(*desktopWidgetScheduler)Wake 0x1407accc0, 96B
//
// 结构 desktopWidgetScheduler：types_desktopwidget.go（wake chan struct{} @ +0x08）。
package main

// Wake 非阻塞唤醒调度器：向 wake 通道发送空结构体，通道满则丢弃。
// [S 汇编 0x1407accc0, 96B] 实证：nil receiver→return；[rax+0x08] 取 wake chan；
// runtime.selectnbsend 非阻塞发送（返回值被忽略，即 select default 分支）。
func (s *desktopWidgetScheduler) Wake() {
	if s == nil {
		return
	}
	select {
	case s.wake <- struct{}{}:
	default:
	}
}
