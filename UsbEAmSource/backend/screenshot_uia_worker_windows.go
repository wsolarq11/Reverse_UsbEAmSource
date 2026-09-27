// AUTO-RECONSTRUCTED — DOMAIN: screenshot COM query worker pool
// Source: UsbEAm_Launcher 1.0.3 (Go 1.25.12, PE64), disassembled from
// main.screenshotCOMQueryWorkerPool* / main.*ScreenshotCOMQueryWorker* symbols.
// Tier markers:
//
//	[S VA]      — body fully translated from asm (pure geometry / math).
//	[S-sig VA]  — signature proven from asm; body is a faithful zero skeleton
//	              (worker 调度 / 超时 / 毒丸机制未还原).
//
// 研究用途
package main

import "time"

// [S-sig 0x1409b33c0] 用默认配置构造工作池（name 为池名，构造默认
// screenshotCOMQueryWorkerPoolConfig 后转入 WithConfig）。返回 *screenshotCOMQueryWorkerPool。
func newScreenshotCOMQueryWorkerPool(name string) *screenshotCOMQueryWorkerPool { return nil }

// [S-sig 0x1409b34c0] newobject 分配 pool；序言从栈 [rsp+0x40..] duffcopy 拷贝 config 结构到 pool 字段，
// 证明 cfg 为值传递（大结构经栈，非指针）。返回 rax 指针。
func newScreenshotCOMQueryWorkerPoolWithConfig(cfg screenshotCOMQueryWorkerPoolConfig) *screenshotCOMQueryWorkerPool {
	return nil
}

// [S-sig 0x1409b3820] 提交查询请求并等待结果。payload 经 execute 回调执行。返回
// screenshotCOMQueryResult。体未还原。
func (p *screenshotCOMQueryWorkerPool) Query(payload screenshotCOMQueryPayload) screenshotCOMQueryResult {
	return screenshotCOMQueryResult{}
}

// [S-sig 0x1409b44a0] morestack 序言保存 7 槽逐寄存器实证（rax=接收者 + 6 非接收者参数）：
//   - ebx/ecx = 2×int32（存 dword，写 req+0x10/+0x14 = payload.point.X/Y）
//   - rdi = uintptr（写 req+0x18 = payload.targetWindow）
//   - sil = uint8（写 req+0x08 = priority）
//   - r8/r9 = error 接口（itab/data；0x1409b44e7 test r8,r8 判 nil，nil 则 fmt.Errorf 构造默认
//     busyError，写 req+0x20/+0x28 = busyError 字段）
//   返回三寄存器：成功路径 rax=req 指针 + rbx/rcx=0（(req,nil)）；busy/shutdown 路径 rax=0 +
//   rbx/rcx=error（(nil,err)）。故返回 (*screenshotCOMQueryRequest, error)。体未还原。
func (p *screenshotCOMQueryWorkerPool) enqueue(x, y int32, targetWindow uintptr, priority uint8, busyError error) (*screenshotCOMQueryRequest, error) {
	return nil, nil
}

// [S-sig 0x1409b4960] 序言仅 spill rax(接收者)，无 worker 参数；正文 makechan + newobject 内部创建 worker，
// 尾声 mov rax,[rsp+0x18] 返回 worker 指针。返回 *screenshotCOMQueryWorker。
func (p *screenshotCOMQueryWorkerPool) startWorkerLocked() *screenshotCOMQueryWorker {
	return nil
}

// [S-sig 0x1409b4b00] 序言 spill rax(接收者)/rbx(worker 指针)，构造闭包转 newproc；尾声无返回值。void。
func (p *screenshotCOMQueryWorkerPool) runWorker(worker *screenshotCOMQueryWorker) {}

// [S-sig 0x1409b5260] 接收者 rax 读 [rax+0x50] 为 pool.execute 回调并间接调用；defer 分支恢复 7 寄存器
// 恰为 screenshotCOMQueryResult{rect 4字, controlType, err 2字}。返回 screenshotCOMQueryResult。
func (p *screenshotCOMQueryWorkerPool) executeRequest(req *screenshotCOMQueryRequest) screenshotCOMQueryResult {
	return screenshotCOMQueryResult{}
}

// [S-sig 0x1409b5500] 接收者 rax 为互斥锁（lock cmpxchg [rax]）；尾声 movzx ebx, byte[rsp+0x16] 返回 bool、
// mov rax,[rsp+0x18] 返回 req 指针。返回 (*screenshotCOMQueryRequest, bool)。
func (p *screenshotCOMQueryWorkerPool) takeNextRequest() (*screenshotCOMQueryRequest, bool) {
	return nil, false
}

// [S-sig 0x1409b5840] 序言 spill rax(接收者)/rbx(req)/rcx(result.controlType)，result 其余 6 字在 rdi..r11；
// 尾声 mov eax, ecx 返回单字节 bool。返回 bool。
func (p *screenshotCOMQueryWorkerPool) completeRequest(req *screenshotCOMQueryRequest, result screenshotCOMQueryResult) bool {
	return false
}

// [S-sig 0x1409b5a20] 序言 spill rax(接收者)/rbx(req)；defer 路径 movzx eax, byte[rsp+0x2e] 返回局部 bool。返回 bool。
func (p *screenshotCOMQueryWorkerPool) cancelPendingRequest(req *screenshotCOMQueryRequest) bool {
	return false
}

// [S-sig 0x1409b5d00] 序言 spill rax(接收者)/rbx(req)；defer 路径 movzx eax, byte[rsp+0x2e] 返回局部 bool。返回 bool。
func (p *screenshotCOMQueryWorkerPool) timeoutRunningRequest(req *screenshotCOMQueryRequest) bool {
	return false
}

// [S-sig 0x1409b5ee0] 序言 spill rsi/rbx(worker 指针)/rax(接收者)，写 [rbx+0x10] 毒丸标志；两返回路径均无返回值。void。
func (p *screenshotCOMQueryWorkerPool) poisonWorkerLocked(worker *screenshotCOMQueryWorker) {}

// [S-sig 0x1409b6080] worker 退出回调（可能带错误）。
// 汇编实证：序言 spill rax(recv)+rbx(worker 指针,[rbx]=key)+rcx/rdi(err error 两字，L75 test rsi=itab
// 判 nil)；mapdelete_fast64 用 [rbx] 作 key 删 worker 映射；返回 void。
func (p *screenshotCOMQueryWorkerPool) workerExited(worker *screenshotCOMQueryWorker, err error) {
	_, _ = worker, err
}

// [S-sig 0x1409b6800] 开始关闭 worker 池并收集被关闭的 worker。
// 汇编实证：序言仅 spill rax(recv)；尾声返回 rax/rbx/rcx 三寄存器 = slice(ptr/len/cap) 而非 int/error
// 组合——[rsp+0xb8]=rdi(ptr)、[rsp+0x48]=r8(len)、[rsp+0x70]=rsi(cap)，循环 growslice 后
// [rdi+r8*8-8]=worker 指针 8 字节元素写入。返回 []*screenshotCOMQueryWorker。
func (p *screenshotCOMQueryWorkerPool) beginShutdown() []*screenshotCOMQueryWorker {
	_ = p
	return nil
}

// [S-sig 0x1409b6f00] 等待所有 COM 查询 worker 退出（带截止时间）。
// 汇编实证：morestack 保存 rax..r8 六寄存器 = workers slice(ptr/len/cap，[rax+rcx*8] 8 字节元素遍历)
// + deadline time.Time(wall/ext/loc，透传 time.Until)；返回 bool（mov eax,0/1）。
func waitForScreenshotCOMQueryWorkers(workers []*screenshotCOMQueryWorker, deadline time.Time) bool {
	_, _ = workers, deadline
	return false
}

// [S-sig 0x1409b7080] 无参；GetMessage/TranslateMessage/DispatchMessage 泵送，cmp [rcx+8],0x12 判 WM_QUIT，
// 收到 WM_QUIT 返回 false（xor eax），否则返回 true（mov eax,1）。返回 bool。
func pumpScreenshotCOMThreadMessages() bool { return false }
