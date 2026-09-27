package main

import (
	"strings"
	"testing"
	"time"
)

// ---- parseMemoryReleaseMode：16 条别名全量表驱动 ----
// [S 0x1408d31e0, 736B] 跳转表按长度 2..22 分派 + 逐字节精确匹配。
// 期望值为汇编中 `lea rax,[rip+disp]` 解出的 5 个规范化常量：
//
//	standbylist            @0x140c47523 (11B)
//	workingsets            @0x140c475de (11B)
//	modifiedpagelist       @0x140c5524a (16B)
//	priority0standbylist   @0x140c5da07 (20B)
func TestParseMemoryReleaseModeAliasTable(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		// len 2
		{"ws", "workingsets", true},
		// len 7
		{"standby", "standbylist", true},
		// len 8
		{"modified", "modifiedpagelist", true},
		// len 9
		{"priority0", "priority0standbylist", true},
		// len 10
		{"priority-0", "priority0standbylist", true},
		{"workingset", "workingsets", true},
		// len 11
		{"standbylist", "standbylist", true},
		{"workingsets", "workingsets", true},
		{"working-set", "workingsets", true},
		{"lowpriority", "priority0standbylist", true},
		// len 12
		{"low-priority", "priority0standbylist", true},
		{"modifiedpage", "modifiedpagelist", true},
		// len 13
		{"modified-page", "modifiedpagelist", true},
		// len 16
		{"modifiedpagelist", "modifiedpagelist", true},
		// len 20
		{"priority0standbylist", "priority0standbylist", true},
		// len 22 —— 注意：**不**返回同名值，而是收敛到 20B 常量
		{"lowprioritystandbylist", "priority0standbylist", true},

		// 归一化：TrimSpace + ToLower
		{"  STANDBY  ", "standbylist", true},
		{"\tWorkingset\n", "workingsets", true},
		{"  WS", "workingsets", true},
		{"PRIORITY-0", "priority0standbylist", true},
		{"LowPriority", "priority0standbylist", true},

		// 非法：空 / 未覆盖长度 / 长度对但内容错
		{"", "", false},
		{"   ", "", false},
		{"w", "", false},                      // len 1 < 2（jump-table 下界）
		{"work", "", false},                   // len 4 未覆盖
		{"workin", "", false},                 // len 6 未覆盖
		{"standbyy", "", false},               // len 8 但非 modified
		{"modifie", "", false},                // len 7 但非 standby
		{"priority9", "", false},              // len 9 但非 priority0
		{"priority-1", "", false},             // len 10 但非 priority-0/workingset
		{"workingsetz", "", false},            // len 11 首字节 >'s' 但尾字节不符
		{"standbylistx", "", false},           // len 12 非 low-priority/modifiedpage
		{"modified-pages", "", false},         // len 14 未覆盖
		{"modifiedpagezzz", "", false},        // len 16 但内容不符
		{"modifiedpagelistx", "", false},      // len 17 未覆盖
		{"priority0standbylis", "", false},    // len 19 未覆盖
		{"priority0standbylistx", "", false},  // len 21 未覆盖
		{"lowprioritystandbylisz", "", false}, // len 23 > 0x14+2（越上界）
		{"lowprioritystandbylis", "", false},  // len 21 未覆盖
	}

	for _, c := range cases {
		got, ok := parseMemoryReleaseMode(c.in)
		if ok != c.ok {
			t.Errorf("parseMemoryReleaseMode(%q) ok=%v want %v", c.in, ok, c.ok)
			continue
		}
		if got != c.want {
			t.Errorf("parseMemoryReleaseMode(%q) = %q want %q", c.in, got, c.want)
		}
	}
}

// 规范化后的取值必须落在 5 个规范常量集合内（防未来误加分支）。
func TestParseMemoryReleaseModeCanonicalSet(t *testing.T) {
	allowed := map[string]bool{
		"standbylist":          true,
		"workingsets":          true,
		"modifiedpagelist":     true,
		"priority0standbylist": true,
	}
	for _, in := range []string{
		"ws", "standby", "modified", "priority0", "priority-0", "workingset",
		"standbylist", "workingsets", "working-set", "lowpriority",
		"low-priority", "modifiedpage", "modified-page", "modifiedpagelist",
		"priority0standbylist", "lowprioritystandbylist",
	} {
		got, ok := parseMemoryReleaseMode(in)
		if !ok {
			t.Errorf("%q should be accepted", in)
			continue
		}
		if !allowed[got] {
			t.Errorf("%q → %q not in canonical set", in, got)
		}
	}
}

// 长度边界：跳转表覆盖 len-2 ∈ [0,0x14] 即 len ∈ [2,22]。
func TestParseMemoryReleaseModeLengthBounds(t *testing.T) {
	if _, ok := parseMemoryReleaseMode("w"); ok {
		t.Error("len 1 must fall outside jump table lower bound")
	}
	if _, ok := parseMemoryReleaseMode(strings.Repeat("x", 23)); ok {
		t.Error("len 23 must fall outside jump table upper bound")
	}
	if _, ok := parseMemoryReleaseMode(strings.Repeat("x", 22)); ok {
		t.Error("len 22 with wrong content must not match")
	}
}

// ---- currentTime：注入时钟优先 ----
func TestMemoryReleaseCurrentTimeInjection(t *testing.T) {
	fixed := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
	s := &memoryReleaseService{now: func() time.Time { return fixed }}
	if got := s.currentTime(); !got.Equal(fixed) {
		t.Errorf("currentTime=%v want %v", got, fixed)
	}

	// now 为 nil 时回退真实时钟（只断言非零且接近当下，保持确定性）
	s2 := &memoryReleaseService{}
	got := s2.currentTime()
	if got.IsZero() {
		t.Error("currentTime with nil now must fall back to time.Now")
	}
	if delta := time.Since(got); delta < 0 || delta > time.Minute {
		t.Errorf("fallback clock drifted: delta=%v", delta)
	}
}

// ---- stopTimerLocked：scheduleID 无条件自增 ----
func TestStopTimerLockedAdvancesScheduleID(t *testing.T) {
	s := &memoryReleaseService{}
	s.scheduleID = 7
	s.stopTimerLocked() // timer == nil 分支
	if s.scheduleID != 8 {
		t.Errorf("scheduleID=%d want 8 (must be unconditional)", s.scheduleID)
	}
	if s.timer != nil {
		t.Error("timer must stay nil")
	}

	s.timer = time.NewTimer(time.Hour)
	defer s.timer.Stop()
	s.scheduleID = 41
	s.stopTimerLocked()
	if s.scheduleID != 42 {
		t.Errorf("scheduleID=%d want 42", s.scheduleID)
	}
	if s.timer != nil {
		t.Error("timer must be cleared to nil after stop")
	}
}

// ---- refreshPlatformStateLocked：nil 守卫 + 回写 state 头部 ----
func TestRefreshPlatformStateLocked(t *testing.T) {
	called := 0
	s := &memoryReleaseService{}
	s.supported = func() bool { called++; return true }
	s.processIsElevated = func() bool { called++; return true }
	s.refreshPlatformStateLocked()
	if called != 2 {
		t.Errorf("probes called %d times want 2", called)
	}
	if !s.state.Supported || !s.state.IsAdmin {
		t.Errorf("state head not refreshed: %+v", s.state)
	}

	// 两个探针为 nil → 不 panic、不改写
	s2 := &memoryReleaseService{}
	s2.refreshPlatformStateLocked()
	if s2.state.Supported || s2.state.IsAdmin {
		t.Error("nil probes must not fabricate capability flags")
	}

	// 探针返回 false → 覆盖既有 true
	s3 := &memoryReleaseService{}
	s3.state.Supported = true
	s3.state.IsAdmin = true
	s3.supported = func() bool { return false }
	s3.processIsElevated = func() bool { return false }
	s3.refreshPlatformStateLocked()
	if s3.state.Supported || s3.state.IsAdmin {
		t.Error("probe result must overwrite prior state")
	}
}

// schedulerActiveFor 复刻 GetState / rescheduleLocked 共用的调度态判据
// （汇编两处同构：`!shuttingDown && state.Supported && moduleEnabled` → config.TimerEnabled）。
func schedulerActiveFor(s *memoryReleaseService) bool {
	if !s.shuttingDown && s.state.Supported && s.moduleEnabled {
		return s.config.TimerEnabled
	}
	return false
}

// ---- GetState：平台刷新 + 调度态重算 + NextRunAt 清理 ----
func TestGetStateRecomputesSchedulerActive(t *testing.T) {
	base := func() *memoryReleaseService {
		s := &memoryReleaseService{}
		s.config.TimerEnabled = true
		s.moduleEnabled = true
		s.state.Supported = true
		return s
	}

	// 全条件满足 → SchedulerActive 跟随 TimerEnabled
	s := base()
	st := s.GetState()
	if !st.SchedulerActive {
		t.Error("SchedulerActive should follow config.TimerEnabled when all gates pass")
	}
	if schedulerActiveFor(s) != st.SchedulerActive {
		t.Error("GetState and rescheduleLocked must agree on the gate")
	}

	// shuttingDown 压制
	s = base()
	s.shuttingDown = true
	s.state.NextRunAt = "2006-01-02T15:04:05Z"
	st = s.GetState()
	if st.SchedulerActive {
		t.Error("shuttingDown must suppress SchedulerActive")
	}
	if st.NextRunAt != "" {
		t.Error("NextRunAt must be cleared when scheduler inactive")
	}

	// moduleEnabled 关闭压制
	s = base()
	s.moduleEnabled = false
	if s.GetState().SchedulerActive {
		t.Error("disabled module must suppress SchedulerActive")
	}

	// Supported 为假压制
	s = base()
	s.state.Supported = false
	if s.GetState().SchedulerActive {
		t.Error("unsupported platform must suppress SchedulerActive")
	}

	// Running 期间保留 NextRunAt（汇编 `!active && !Running` 才清）
	s = base()
	s.state.NextRunAt = "2026-01-02T03:04:05Z"
	s.state.Running = true
	st = s.GetState()
	if st.NextRunAt != "2026-01-02T03:04:05Z" {
		t.Errorf("NextRunAt must survive while Running: %q", st.NextRunAt)
	}
}

// ---- rescheduleLocked：NextRunAt 用 RFC3339 格式化 ----
func TestRescheduleLockedFormatsRFC3339(t *testing.T) {
	s := &memoryReleaseService{}
	s.config.TimerEnabled = true
	s.config.IntervalMinutes = 30
	s.moduleEnabled = true
	s.state.Supported = true

	now := time.Date(2026, 5, 6, 7, 8, 9, 0, time.UTC)
	s.rescheduleLocked(now)
	if s.timer != nil {
		s.timer.Stop()
	}

	want := now.Add(30 * time.Minute).Format(time.RFC3339)
	if s.state.NextRunAt != want {
		t.Errorf("NextRunAt=%q want %q (RFC3339)", s.state.NextRunAt, want)
	}
	if _, err := time.Parse(time.RFC3339, s.state.NextRunAt); err != nil {
		t.Errorf("NextRunAt not RFC3339-parseable: %v", err)
	}
	if !s.state.SchedulerActive {
		t.Error("SchedulerActive should be true when rescheduling an active timer")
	}
	if s.timer == nil {
		t.Error("timer must be armed via time.AfterFunc")
	}
	if s.scheduleID != 1 {
		t.Errorf("scheduleID=%d want 1 (stopTimerLocked ran once)", s.scheduleID)
	}
}

// 调度未激活 / 正在运行 → 清空 NextRunAt 且不武装定时器。
func TestRescheduleLockedInactivePaths(t *testing.T) {
	now := time.Date(2026, 5, 6, 7, 8, 9, 0, time.UTC)

	s := &memoryReleaseService{}
	s.state.Supported = true
	s.moduleEnabled = true
	s.config.TimerEnabled = false // 未启用
	s.state.NextRunAt = "stale"
	s.rescheduleLocked(now)
	if s.timer != nil {
		t.Error("timer must not be armed when TimerEnabled=false")
	}
	if s.state.NextRunAt != "" {
		t.Errorf("NextRunAt must be cleared, got %q", s.state.NextRunAt)
	}
	if s.state.SchedulerActive {
		t.Error("SchedulerActive must be false")
	}

	s2 := &memoryReleaseService{}
	s2.state.Supported = true
	s2.moduleEnabled = true
	s2.config.TimerEnabled = true
	s2.state.Running = true // 运行中不重排
	s2.state.NextRunAt = "stale"
	s2.rescheduleLocked(now)
	if s2.timer != nil {
		t.Error("timer must not be armed while Running")
	}
	if s2.state.NextRunAt != "" {
		t.Errorf("NextRunAt must be cleared while Running, got %q", s2.state.NextRunAt)
	}
}

// ---- runLocked：四条前置错误分支 + 正常收尾 ----
func TestRunLockedGateErrors(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 9, 10, 0, time.UTC)
	newSvc := func() *memoryReleaseService {
		s := &memoryReleaseService{}
		s.moduleEnabled = true
		s.state.Supported = true
		s.now = func() time.Time { return now }
		return s
	}

	// shuttingDown → 返回快照 state + 30B 错误
	s := newSvc()
	s.shuttingDown = true
	st, err := s.runLocked("standbylist")
	if err == nil || err.Error() != "内存释放服务正在关闭" {
		t.Errorf("shuttingDown err=%v", err)
	}
	if st.Supported != s.state.Supported {
		t.Error("shuttingDown path must return a real state snapshot")
	}

	// !moduleEnabled → 返回零值 state + 33B 错误
	s = newSvc()
	s.moduleEnabled = false
	st, err = s.runLocked("standbylist")
	if err == nil || err.Error() != "内存释放功能当前已关闭" {
		t.Errorf("moduleEnabled err=%v", err)
	}
	if st != (MemoryReleaseState{}) {
		t.Errorf("disabled path must return zero state, got %+v", st)
	}

	// !Supported → 返回零值 state + 39B 错误
	s = newSvc()
	s.state.Supported = false
	st, err = s.runLocked("standbylist")
	if err == nil || err.Error() != "当前平台不支持内存释放功能" {
		t.Errorf("unsupported err=%v", err)
	}
	if st != (MemoryReleaseState{}) {
		t.Errorf("unsupported path must return zero state, got %+v", st)
	}

	// 已在运行 → 返回快照 + 27B 错误
	s = newSvc()
	s.state.Running = true
	st, err = s.runLocked("standbylist")
	if err == nil || err.Error() != "内存释放正在执行中" {
		t.Errorf("running err=%v", err)
	}
	if st.Running != true {
		t.Error("running path must return the live snapshot")
	}

	// 模式非法（runLocked 内部**二次验证**）
	s = newSvc()
	st, err = s.runLocked("no-such-mode")
	if err == nil || !strings.Contains(err.Error(), "不支持的内存释放模式") {
		t.Errorf("bad mode err=%v", err)
	}
	if st != s.state {
		t.Error("bad mode path must return current state")
	}
}

// 正常路径：generation/scheduleID 推进、LastMode 归一化、LastRunAt 用 RFC3339、错误落 LastError。
func TestRunLockedSuccessPath(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 9, 10, 0, time.UTC)
	var invoked string
	s := &memoryReleaseService{}
	s.moduleEnabled = true
	s.state.Supported = true
	s.now = func() time.Time { return now }
	s.execute = func(mode string) error { invoked = mode; return nil }

	st, err := s.runLocked("WS") // 别名 → 归一化
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if invoked != "workingsets" {
		t.Errorf("execute got %q want normalized %q", invoked, "workingsets")
	}
	if st.LastMode != "workingsets" {
		t.Errorf("LastMode=%q want workingsets", st.LastMode)
	}
	if st.LastRunAt != now.Format(time.RFC3339) {
		t.Errorf("LastRunAt=%q want RFC3339 of injected clock", st.LastRunAt)
	}
	if st.LastError != "" {
		t.Errorf("LastError=%q want empty on success", st.LastError)
	}
	if st.Running {
		t.Error("Running must be false after completion")
	}
	if s.generation != 1 {
		t.Errorf("generation=%d want 1", s.generation)
	}
	if s.runningDone != nil {
		t.Error("runningDone must be released")
	}
	if s.timer != nil {
		s.timer.Stop()
	}
}

// 执行失败 → LastError 记录；代数不匹配（陈旧代）→ 不刷新 LastRunAt/LastError。
func TestRunLockedExecutionErrorRecorded(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 9, 10, 0, time.UTC)
	s := &memoryReleaseService{}
	s.moduleEnabled = true
	s.state.Supported = true
	s.now = func() time.Time { return now }
	s.execute = func(string) error { return errMemoryReleaseFake }

	st, err := s.runLocked("standbylist")
	if err == nil || err != errMemoryReleaseFake {
		t.Fatalf("err=%v want the execute error propagated", err)
	}
	if !strings.Contains(st.LastError, "fake release failure") {
		t.Errorf("LastError=%q must carry execute error text", st.LastError)
	}
	if s.timer != nil {
		s.timer.Stop()
	}
}

var errMemoryReleaseFake = &memoryReleaseFakeError{}

type memoryReleaseFakeError struct{}

func (*memoryReleaseFakeError) Error() string { return "fake release failure" }

// ---- RunNow：签名与错误载体 ----
// RunNow 返回 (MemoryReleaseState, error)：state 走栈、error 走寄存器。
func TestRunNowUnsupportedModeReturnsStateAndError(t *testing.T) {
	s := &memoryReleaseService{}
	s.state.Supported = true
	s.state.IsAdmin = true

	st, err := s.RunNow("definitely-not-a-mode")
	if err == nil {
		t.Fatal("expected error for unsupported mode")
	}
	if !strings.Contains(err.Error(), "不支持的内存释放模式") {
		t.Errorf("err=%v", err)
	}
	// 错误文本携带 TrimSpace 后的原始入参，格式为 "不支持的内存释放模式: %s"（冒号+单空格）
	if !strings.Contains(err.Error(), ": definitely-not-a-mode") {
		t.Errorf("err must embed ': '+trimmed mode: %v", err)
	}
	// 错误路径返回的是 GetState 的结果（此处 state 未被平台探测改写）
	if !st.Supported {
		t.Error("error path must return the live state snapshot")
	}
}

func TestRunNowTrimsModeInErrorMessage(t *testing.T) {
	s := &memoryReleaseService{}
	_, err := s.RunNow("   bogus   ")
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "   bogus   ") {
		t.Errorf("error must embed TrimSpace(mode), got %q", err.Error())
	}
	if !strings.HasSuffix(err.Error(), "bogus") {
		t.Errorf("error must end with trimmed mode, got %q", err.Error())
	}
}

// RunNow 合法模式 → 委托 runLocked，并传入**归一化**后的模式。
func TestRunNowDelegatesNormalizedMode(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 9, 10, 0, time.UTC)
	var invoked string
	s := &memoryReleaseService{}
	s.moduleEnabled = true
	s.state.Supported = true
	s.now = func() time.Time { return now }
	s.execute = func(mode string) error { invoked = mode; return nil }

	if _, err := s.RunNow("  LowPriority  "); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if invoked != "priority0standbylist" {
		t.Errorf("execute got %q want priority0standbylist", invoked)
	}
	if s.timer != nil {
		s.timer.Stop()
	}
}

// 未启用调度时应无定时器残留（避免后续测试被 AfterFunc 的 goroutine 污染）。
func TestRunLockedLeavesNoTimerWhenSchedulerDisabled(t *testing.T) {
	now := time.Date(2026, 6, 7, 8, 9, 10, 0, time.UTC)
	s := &memoryReleaseService{}
	s.moduleEnabled = true
	s.state.Supported = true
	s.config.TimerEnabled = false
	s.now = func() time.Time { return now }
	s.execute = func(string) error { return nil }

	if _, err := s.runLocked("workingsets"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if s.timer != nil {
		t.Error("no timer should be armed when TimerEnabled=false")
	}
	if s.state.SchedulerActive {
		t.Error("SchedulerActive must be false")
	}
}
