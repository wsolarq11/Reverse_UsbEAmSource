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

// Start 单次启动调度循环：once(+0x20) 内异步 go run()。
// [S 汇编 0x1407acb80, 96B]：读 once.done([rax+0x20])，未完成则 doSlow 携带闭包 func1；
// func1（0x1407acbe0）newobject 打包 gowrap1 捕获 scheduler，runtime.newproc 起 goroutine；
// gowrap1（0x1407acc60）尾调 desktopWidgetScheduler.run。幂等（once 保证只跑一次）。
func (s *desktopWidgetScheduler) Start() {
	s.once.Do(func() {
		go s.run()
	})
}

// run 调度循环主体。[S-sig 0x1407acec0]：签名经 Start.func1.gowrap1 尾调实证——
// 单 receiver、无参数、无返回。体待调度循环专项还原。
func (s *desktopWidgetScheduler) run() {
}

// Stop 停止调度器（stopOnce 关 stop 通道，once 关 wake，等 5s timer/done）。
// [S-sig 0x1407acd20, 288B]：nil 早退；stopOnce(+0x2c) 关 stop；once(+0x20) 关 wake；
// NewTimer(5s) + selectgo 等待。体待调度停止序列专项还原。
func (s *desktopWidgetScheduler) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() { close(s.stop) })
	s.once.Do(func() { close(s.wake) })
}
