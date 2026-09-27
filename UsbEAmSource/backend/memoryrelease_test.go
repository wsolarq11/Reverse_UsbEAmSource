package main

import "testing"

func TestNewMemoryReleaseService(t *testing.T) {
	s := newMemoryReleaseService()
	if s == nil {
		t.Fatal("factory must return non-nil")
	}
	// 配置装配：[S] honest interval 默认 30 分钟、TimerEnabled=false。
	if s.config.IntervalMinutes != 30 {
		t.Errorf("interval=%d want 30", s.config.IntervalMinutes)
	}
	if s.config.TimerEnabled {
		t.Error("timer must default disabled")
	}
	// 能力探测已接线。
	if s.supported == nil || s.processIsElevated == nil {
		t.Error("capability probes must be wired")
	}
	// 初始 state 由探测结果填充：
	//   Supported 恒为 true（memoryReleaseSupported 汇编为 `mov eax,1; ret` 编译期常量）；
	//   IsAdmin 来自真实 Windows 令牌探测（Token.IsElevated），结果随运行权限而定，
	//   因此断言「state 与探测结果一致」而非固定值。
	if !s.state.Supported {
		t.Error("state.Supported should be true (memoryReleaseSupported returns constant true)")
	}
	if s.state.IsAdmin != s.processIsElevated() {
		t.Errorf("state.IsAdmin=%v must equal processIsElevated()=%v", s.state.IsAdmin, s.processIsElevated())
	}
}
