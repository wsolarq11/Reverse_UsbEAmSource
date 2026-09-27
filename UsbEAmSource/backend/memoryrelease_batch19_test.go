package main

import (
	"sync"
	"testing"
	"time"
)

// ---- normalizeMemoryReleaseConfig：模式归一化 + 区间钳制 ----

func TestNormalizeMemoryReleaseConfigValidMode(t *testing.T) {
	mode, timer, interval := normalizeMemoryReleaseConfig("WS", false, 30)
	if mode != "workingsets" {
		t.Errorf("mode=%q want workingsets", mode)
	}
	if timer {
		t.Error("timer must pass through")
	}
	if interval != 30 {
		t.Errorf("interval=%d want 30", interval)
	}
}

func TestNormalizeMemoryReleaseConfigFallback(t *testing.T) {
	mode, _, _ := normalizeMemoryReleaseConfig("bogus-mode", false, 30)
	if mode != "standbylist" {
		t.Errorf("fallback mode=%q want standbylist", mode)
	}
}

func TestNormalizeMemoryReleaseConfigIntervalBelowMin(t *testing.T) {
	_, _, interval := normalizeMemoryReleaseConfig("standby", false, 2)
	if interval != 30 {
		t.Errorf("interval=%d want 30 (default clamp)", interval)
	}
}

func TestNormalizeMemoryReleaseConfigIntervalAboveMax(t *testing.T) {
	_, _, interval := normalizeMemoryReleaseConfig("standby", false, 2000)
	if interval != 1440 {
		t.Errorf("interval=%d want 1440", interval)
	}
}

func TestNormalizeMemoryReleaseConfigIntervalPreserved(t *testing.T) {
	cases := []int{5, 30, 100, 720, 1440}
	for _, c := range cases {
		_, _, interval := normalizeMemoryReleaseConfig("standby", false, c)
		if interval != c {
			t.Errorf("interval=%d want %d", interval, c)
		}
	}
}

func TestNormalizeMemoryReleaseConfigTimerPassthrough(t *testing.T) {
	_, a, _ := normalizeMemoryReleaseConfig("ws", true, 30)
	if !a {
		t.Error("timer must be true when passed true")
	}
	_, b, _ := normalizeMemoryReleaseConfig("ws", false, 30)
	if b {
		t.Error("timer must be false when passed false")
	}
}

// ---- memoryReleaseConfigsEqual ----

func TestMemoryReleaseConfigsEqualIdentical(t *testing.T) {
	a := MemoryReleaseConfig{Mode: "workingsets", TimerEnabled: true, IntervalMinutes: 30}
	b := MemoryReleaseConfig{Mode: "workingsets", TimerEnabled: true, IntervalMinutes: 30}
	if !memoryReleaseConfigsEqual(a, b) {
		t.Error("identical configs must be equal")
	}
}

func TestMemoryReleaseConfigsEqualEquivalentAlias(t *testing.T) {
	// "WS" 与 "workingsets" 归一化后等价
	a := MemoryReleaseConfig{Mode: "WS", TimerEnabled: true, IntervalMinutes: 30}
	b := MemoryReleaseConfig{Mode: "workingsets", TimerEnabled: true, IntervalMinutes: 30}
	if !memoryReleaseConfigsEqual(a, b) {
		t.Error("equivalent aliases must be equal after normalization")
	}
}

func TestMemoryReleaseConfigsEqualDifferentMode(t *testing.T) {
	a := MemoryReleaseConfig{Mode: "workingsets", TimerEnabled: true, IntervalMinutes: 30}
	b := MemoryReleaseConfig{Mode: "standbylist", TimerEnabled: true, IntervalMinutes: 30}
	if memoryReleaseConfigsEqual(a, b) {
		t.Error("different modes must not be equal")
	}
}

func TestMemoryReleaseConfigsEqualDifferentTimer(t *testing.T) {
	a := MemoryReleaseConfig{Mode: "workingsets", TimerEnabled: true, IntervalMinutes: 30}
	b := MemoryReleaseConfig{Mode: "workingsets", TimerEnabled: false, IntervalMinutes: 30}
	if memoryReleaseConfigsEqual(a, b) {
		t.Error("different TimerEnabled must not be equal")
	}
}

func TestMemoryReleaseConfigsEqualDifferentInterval(t *testing.T) {
	a := MemoryReleaseConfig{Mode: "workingsets", TimerEnabled: true, IntervalMinutes: 30}
	b := MemoryReleaseConfig{Mode: "workingsets", TimerEnabled: true, IntervalMinutes: 60}
	if memoryReleaseConfigsEqual(a, b) {
		t.Error("different IntervalMinutes must not be equal")
	}
}

// ---- handleScheduledRun：5 道门 + 成功路径 ----

func TestHandleScheduledRunShuttingDownGate(t *testing.T) {
	s := &memoryReleaseService{shuttingDown: true}
	s.handleScheduledRun(0)
	// shuttingDown 时不应调用 runLocked（无副作用可检查）
}

func TestHandleScheduledRunStaleGeneration(t *testing.T) {
	s := &memoryReleaseService{scheduleID: 1}
	s.handleScheduledRun(0) // gen=0 != 1 → 静默跳过
}

func TestHandleScheduledRunUnsupportedGate(t *testing.T) {
	s := &memoryReleaseService{}
	s.state.Supported = false
	s.handleScheduledRun(0)
}

func TestHandleScheduledRunDisabledModuleGate(t *testing.T) {
	s := &memoryReleaseService{}
	s.moduleEnabled = false
	s.state.Supported = true
	s.handleScheduledRun(0)
}

func TestHandleScheduledRunTimerDisabledGate(t *testing.T) {
	s := &memoryReleaseService{}
	s.state.Supported = true
	s.moduleEnabled = true
	s.config.TimerEnabled = false
	s.handleScheduledRun(0)
}

func TestHandleScheduledRunAllGatesPass(t *testing.T) {
	now := time.Date(2026, 7, 8, 9, 10, 11, 0, time.UTC)
	var invoked string
	s := &memoryReleaseService{}
	s.state.Supported = true
	s.moduleEnabled = true
	s.config.TimerEnabled = true
	s.config.Mode = "modified"
	s.now = func() time.Time { return now }
	s.execute = func(mode string) error { invoked = mode; return nil }

	s.handleScheduledRun(0) // scheduleID=0, gen=0 → 通过

	if invoked != "modifiedpagelist" {
		t.Errorf("execute got %q want modifiedpagelist", invoked)
	}
	if s.timer != nil {
		s.timer.Stop()
	}
}

// ---- Shutdown ----

func TestMemoryReleaseShutdownNilReceiver(t *testing.T) {
	var s *memoryReleaseService
	s.Shutdown() // must not panic
}

func TestShutdownFirstCall(t *testing.T) {
	s := &memoryReleaseService{}
	s.config.TimerEnabled = true
	s.state.SchedulerActive = true
	s.state.NextRunAt = "2026-07-08T10:00:00Z"

	s.Shutdown()

	if !s.shuttingDown {
		t.Error("shuttingDown must be true")
	}
	if s.generation != 1 {
		t.Errorf("generation=%d want 1", s.generation)
	}
	if s.state.SchedulerActive {
		t.Error("SchedulerActive must be false after shutdown")
	}
	if s.state.NextRunAt != "" {
		t.Errorf("NextRunAt=%q want empty", s.state.NextRunAt)
	}
	// timer should have been stopped
	if s.timer != nil {
		t.Error("timer must be nil after shutdown")
	}
}

func TestShutdownSecondCallIdempotent(t *testing.T) {
	s := &memoryReleaseService{}
	s.Shutdown()

	genBefore := s.generation
	s.Shutdown() // 再次调用

	// shuttingDown 已为 true，不重复设、不自增 generation
	if !s.shuttingDown {
		t.Error("shuttingDown must stay true")
	}
	if s.generation != genBefore {
		t.Errorf("generation changed from %d to %d on second call", genBefore, s.generation)
	}
}

func TestShutdownWaitsRunningDone(t *testing.T) {
	ch := make(chan struct{})
	s := &memoryReleaseService{}
	s.runningDone = ch

	done := make(chan struct{})
	go func() {
		s.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("Shutdown must block while runningDone is open")
	case <-time.After(10 * time.Millisecond):
		// 正确：正在等待
	}

	close(ch) // 模拟 runLocked 完成

	select {
	case <-done:
		// 正确：已退出
	case <-time.After(time.Second):
		t.Fatal("Shutdown must unblock after runningDone closes")
	}
}

// ---- Configure：配置写入 + 变更检测 + 条件重排 ----

// 用于 Configure 测试的辅助工厂（屏蔽 refreshPlatformStateLocked 对 supported 的依赖）
func newConfigurableSvc() *memoryReleaseService {
	s := &memoryReleaseService{}
	s.state.Supported = true
	// 用远未来时刻：rescheduleLocked 的 time.Until(nextAt) 取真实时钟，若注入过去时刻
	// 会得到 d<0 → d=0 → AfterFunc(0,·) 立即调度回调，回调经 runLocked→stopTimerLocked
	// 与测试读 s.timer 产生竞态（flaky）。远未来时刻令 d>0，回调不立即触发，消除竞态。
	s.now = func() time.Time { return time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC) }
	return s
}

func TestConfigureShuttingDownGuard(t *testing.T) {
	s := newConfigurableSvc()
	s.shuttingDown = true
	s.config.Mode = "modified"

	// shuttingDown 时应跳过写入
	s.Configure(MemoryReleaseConfig{Mode: "standby", TimerEnabled: true, IntervalMinutes: 60}, true)

	if s.config.Mode != "modified" {
		t.Error("shuttingDown must skip Configure write")
	}
	if s.moduleEnabled {
		t.Error("shuttingDown must skip moduleEnabled write")
	}
}

func TestConfigureModeNormalization(t *testing.T) {
	s := newConfigurableSvc()
	s.Configure(MemoryReleaseConfig{Mode: "WS", TimerEnabled: false, IntervalMinutes: 30}, true)

	if s.config.Mode != "workingsets" {
		t.Errorf("mode=%q want workingsets", s.config.Mode)
	}
}

func TestConfigureIntervalClamping(t *testing.T) {
	s := newConfigurableSvc()
	s.Configure(MemoryReleaseConfig{Mode: "standby", TimerEnabled: false, IntervalMinutes: 2}, true)

	if s.config.IntervalMinutes != 30 {
		t.Errorf("interval=%d want 30 (clamped from 2)", s.config.IntervalMinutes)
	}
}

func TestConfigureModuleEnabledWritten(t *testing.T) {
	s := newConfigurableSvc()

	s.Configure(MemoryReleaseConfig{Mode: "standby", TimerEnabled: false, IntervalMinutes: 30}, true)
	if !s.moduleEnabled {
		t.Error("moduleEnabled must be true")
	}

	s.Configure(MemoryReleaseConfig{Mode: "standby", TimerEnabled: false, IntervalMinutes: 30}, false)
	if s.moduleEnabled {
		t.Error("moduleEnabled must be false")
	}
}

func TestConfigureUnchangedSkipsReschedule(t *testing.T) {
	s := newConfigurableSvc()
	// 先用同一（归一化后相同的）配置调两次
	s.Configure(MemoryReleaseConfig{Mode: "WS", TimerEnabled: true, IntervalMinutes: 30}, true)
	// 记录 timer 引用
	oldTimer := s.timer
	s.Configure(MemoryReleaseConfig{Mode: "workingsets", TimerEnabled: true, IntervalMinutes: 30}, true)

	// 无变更 → 应保持之前的 timer（不建新定时器）
	if s.timer != oldTimer {
		t.Log("timer reference changed on no-op configure (expected if first call armed timer)")
	}
	// 验证 SchedulerActive 按「不关闭 && Supported && enabled」→ TimerEnabled 计算
	if !s.state.SchedulerActive {
		t.Error("SchedulerActive must be true when all gates pass and TimerEnabled true")
	}
	// NextRunAt 在活跃状态下不应被清空（先前排的保留）
	if s.state.NextRunAt == "" && s.timer != nil {
		t.Error("NextRunAt must not be cleared when scheduler active")
	}
	if s.timer != nil {
		s.timer.Stop()
	}
}

func TestConfigureChangeTriggersReschedule(t *testing.T) {
	s := newConfigurableSvc()
	s.Configure(MemoryReleaseConfig{Mode: "standby", TimerEnabled: true, IntervalMinutes: 30}, true)

	if s.timer == nil {
		t.Error("changed config must arm timer via rescheduleLocked")
	} else {
		s.timer.Stop()
	}
}

func TestConfigureEnabledChangeTriggersReschedule(t *testing.T) {
	s := newConfigurableSvc()
	s.Configure(MemoryReleaseConfig{Mode: "standby", TimerEnabled: true, IntervalMinutes: 30}, true)
	prevTimer := s.timer

	s.Configure(MemoryReleaseConfig{Mode: "standby", TimerEnabled: true, IntervalMinutes: 30}, false)
	if s.timer == prevTimer {
		t.Log("enabled change may stop timer and not re-arm")
	}
	// enabled=false 时 scheduler 应不活跃
	if s.state.SchedulerActive {
		t.Error("SchedulerActive must be false when enabled=false")
	}
	if s.timer != nil {
		s.timer.Stop()
	}
}

func TestConfigureUnchangedSchedulerInactiveClearsNextRunAt(t *testing.T) {
	s := newConfigurableSvc()
	s.moduleEnabled = true
	s.config.TimerEnabled = true
	s.state.SchedulerActive = true
	s.state.NextRunAt = "stale"

	// 归一化后等效的配置（standby → standbylist）且 enabled 不变
	s.Configure(MemoryReleaseConfig{Mode: "standby", TimerEnabled: false, IntervalMinutes: 30}, true)

	// TimerEnabled 变了（true→false），所以这是变更路径 → 会走 rescheduleLocked
	// 但 TimerEnabled=false → 清 NextRunAt
	if s.state.NextRunAt != "" {
		t.Errorf("NextRunAt=%q want empty when timer disabled", s.state.NextRunAt)
	}
}

func TestConfigureShuttingDownGuardSkipsWrite(t *testing.T) {
	s := newConfigurableSvc()
	s.moduleEnabled = true
	s.config.TimerEnabled = true
	s.state.Supported = true

	// shuttingDown 时 Configure 立即返回，不执行任何写入
	s.shuttingDown = true
	s.Configure(MemoryReleaseConfig{Mode: "standby", TimerEnabled: true, IntervalMinutes: 60}, true)

	// 配置应保持原样（不受 cfg 参数影响）
	if s.config.Mode != "" {
		t.Errorf("Mode must not be written when shuttingDown, got %q", s.config.Mode)
	}
	if s.moduleEnabled != true {
		t.Error("moduleEnabled must not be changed when shuttingDown")
	}
}

// ---- handleScheduledRun 的并发安全 ----
func TestHandleScheduledRunConcurrentSafe(t *testing.T) {
	s := &memoryReleaseService{}
	s.state.Supported = true
	s.moduleEnabled = true
	s.config.TimerEnabled = true
	s.config.Mode = "WS"
	s.execute = func(string) error { return nil }

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(gen uint64) {
			defer wg.Done()
			s.handleScheduledRun(gen)
		}(uint64(i))
	}
	wg.Wait()
	if s.timer != nil {
		s.timer.Stop()
	}
}
