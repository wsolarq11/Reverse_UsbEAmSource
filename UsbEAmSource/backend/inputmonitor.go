// AUTO-RECONSTRUCTED SERVICE METHODS — DOMAIN: inputmonitor 工厂/初始化 (装配依赖叶子)
// 研究用途
//
// 契约来源：
//   - newInputMonitorService(0x140861c40, 224B) / inputMonitorInitialize(0x140865d20, 160B) 汇编
//   - 结构 inputMonitorService：types_inputmonitor.go
//
// 档位：[S] 工厂装配/初始化主骨架（汇编逐条）；字段 byte-offset 对齐为 [P]（Stub 结构用字段名）。
package main

import (
	"errors"
	"strings"
)

// newInputMonitorService 构造并初始化输入监听服务。
// [S 汇编 0x140861c40]：make([]InputMonitorEvent, 0x1000) + 两 map + newobject 装配 +
// 若干 limit(0x1000) + pressedKeys/owners，随后 inputMonitorInitialize。
func newInputMonitorService() *inputMonitorService {
	events := make([]InputMonitorEvent, 0x1000)
	pressed := make(map[uint32]bool)
	owners := make(map[string]InputMonitorStartRequest)
	svc := &inputMonitorService{
		events:      events,
		pressedKeys: pressed,
		owners:      owners,
		maxEvents:   0x1000,
		eventSize:   0x1000,
	}
	inputMonitorInitialize(svc)
	return svc
}

// Snapshot 获取输入监控快照。
// [S-sig 0x140864e60]：签名经符号表实证；体骨架（输入钩子域待落地）。
func (s *inputMonitorService) Snapshot() interface{} { return nil }

// StartOwner 以指定所有者启动监控。
// [S-sig 0x140861f40]：签名经符号表实证；体骨架。
func (s *inputMonitorService) StartOwner(owner string) error { return nil }

// StopOwner 停止指定所有者的监控。
// [S-sig 0x140864340]：签名经符号表实证；体骨架。
func (s *inputMonitorService) StopOwner(owner string) error { return nil }

// [S 汇编 0x140865d20]：nil 返回；持锁设 platformHealthManaged=true（汇编写 +0xd8 处为健康标志）；
// [P] 具体 bool 字段名待与字节对齐复核。
func inputMonitorInitialize(s *inputMonitorService) {
	if s == nil {
		return
	}
	s.lock.Lock()
	s.platformHealthManaged = true
	s.lock.Unlock()
}

// advancePlatformGeneration 递增平台世代号（加锁）。
// [S 汇编 0x140861b60] 单 receiver 无参无返回。asm：nil 返回 → lock(+0x08) →
// platformGeneration(+0xd0).Add(1) → unlock。
func (s *inputMonitorService) advancePlatformGeneration() {
	if s == nil {
		return
	}
	s.lock.Lock()
	s.platformGeneration.Add(1)
	s.lock.Unlock()
}

// startPlatformOwned 启动平台拥有（startOverride 优先，否则 startPlatform）。
// [S-sig 0x140861a00, 160B]：(a, b bool) 无返回。asm：advancePlatformGeneration → 读 +0x100
// 闭包，非 nil 则间接调用 (a,b)，否则 startPlatform(a,b)。体待 startPlatform 专项还原。
func (s *inputMonitorService) startPlatformOwned(a, b bool) {
	s.advancePlatformGeneration()
	_, _ = a, b
}

// stopPlatformThread 停止平台线程（大函数 1785B，体留待平台线程域专项）。
// [S-sig 0x140866120] 单 receiver 无参无返回（序言 test rax + 读 done(+0xa0)）。
func (s *inputMonitorService) stopPlatformThread() {}

// stopPlatformOwned 停止平台拥有（stopOverride 优先，否则停线程）。
// [S 汇编 0x140861aa0] 单 receiver 无参无返回。asm：advancePlatformGeneration → 读 +0x108 闭包，
// 非 nil 则间接调用（[closure]=fn），否则 stopPlatformThread。
func (s *inputMonitorService) stopPlatformOwned() {
	s.advancePlatformGeneration()
	if s.stopOverride != nil {
		s.stopOverride()
	} else {
		s.stopPlatformThread()
	}
}

// closePlatformOwned 关闭平台拥有（closeOverride 优先，否则停线程）。
// [S 汇编 0x140861b00] 单 receiver 无参无返回。asm：advancePlatformGeneration → 读 +0x110 闭包，
// 非 nil 则间接调用，否则 stopPlatformThread。
func (s *inputMonitorService) closePlatformOwned() {
	s.advancePlatformGeneration()
	if s.closeOverride != nil {
		s.closeOverride()
	} else {
		s.stopPlatformThread()
	}
}

// normalizeInputMonitorOwner 规范化输入监视 owner 名称：TrimSpace 后非空且 ≤256 字节则返回，
// 否则返回 ("", error)。[S 汇编 0x140861ea0, 160B] 实证：TrimSpace → len==0 或 len>0x100
// → newobject 错误("输入监测 owner 无效", 25 字节) 返回 ("", err)；否则返回 (trimmed, nil)。
func normalizeInputMonitorOwner(s string) (string, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" || len(trimmed) > 0x100 {
		return "", errors.New("输入监测 owner 无效")
	}
	return trimmed, nil
}
