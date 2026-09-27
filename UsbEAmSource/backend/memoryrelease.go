// AUTO-RECONSTRUCTED SERVICE METHODS — DOMAIN: memoryrelease (工厂 + 执行/调度域)
// 研究用途
//
// 契约来源（全部 [S] 汇编实证，VA 见各函数注释）：
//   - parseMemoryReleaseMode        0x1408d31e0  736B  模式规范化（16 条别名 → 5 个规范值）
//   - newMemoryReleaseService       0x1408d3560  192B  装配 + 能力探测
//   - GetState                      0x1408d3960  384B  lock + refresh + SchedulerActive 重算 + NextRunAt 清理
//   - RunNow                        0x1408d3b40  896B
//   - refreshPlatformStateLocked    0x1408d3ec0   96B
//   - currentTime                   0x1408d3f20   64B
//   - stopTimerLocked               0x1408d3f60  160B
//   - rescheduleLocked              0x1408d4000  608B  + func1 0x1408d4260 (64B)
//   - runLocked                     0x1408d43c0 2496B  + deferwrap1 0x1408d4d80 (96B)
//
// 结构体字段偏移实证（与 types_memoryrelease.go 字段序双向交叉印证）：
//
//	lock@0x00 config@0x08 moduleEnabled@0x28 state@0x30 timer@0x78
//	scheduleID@0x80 generation@0x88 shuttingDown@0x90 runningDone@0x98
//	execute@0xa0 supported@0xa8 processIsElevated@0xb0 now@0xb8
//
//	config.Mode@0x08+0x00(16B) config.TimerEnabled@0x18 config.IntervalMinutes@0x20
//	state.Supported@0x30 state.IsAdmin@0x31 state.Running@0x32 state.SchedulerActive@0x33
//	state.NextRunAt@0x38(+0x40) state.LastRunAt@0x48(+0x50)
//	state.LastMode@0x58(+0x60) state.LastError@0x68(+0x70)
package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sys/windows"
)

// memoryReleaseSupported 探针：本构建下内存释放始终受支持（编译期常量返回）。
// [S 汇编 0x1408d5000, 32B] 实证：`mov eax, 1; ret` —— 无分支、无调用、无参数依赖。
// 常量语义由编译器内联决定：该平台判定恒为 true，相当于 `return true`。
func memoryReleaseSupported() bool {
	return true
}

// openCurrentProcessToken 打开当前进程的访问令牌。
// [S 汇编 0x1408d5dc0, 160B] 实证流程：
//
//	access(eax) → OpenProcessToken(0xffffffffffffffff /* pseudo current process */, access, &token) →
//	失败 → fmt.Errorf（34B 消息 @VA 0x140c65c17）→ return (0, err)；
//	成功 → return (token, nil)。
func openCurrentProcessToken(access uint32) (windows.Token, error) {
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), access, &token); err != nil {
		return 0, fmt.Errorf("open current process token: %w", err)
	}
	return token, nil
}

// memoryReleaseProcessIsElevated 探针：当前进程是否以提升权限运行。
// [S 汇编 0x1408d5020, 160B] 实证流程：
//
//	openCurrentProcessToken(8 /* TOKEN_QUERY */) →
//	err 非空 → return false；
//	成功 → defer token.Close() → token.IsElevated()（golang.org/x/sys/windows.Token.IsElevated
//	@0x140193d40）→ 返回其结果。
func memoryReleaseProcessIsElevated() bool {
	token, err := openCurrentProcessToken(windows.TOKEN_QUERY)
	if err != nil {
		return false
	}
	defer token.Close()
	return token.IsElevated()
}

// newMemoryReleaseService 构造并装配内存释放服务。
// [S 汇编 0x1408d3560]：newobject 装配 config（TimerEnabled=false, IntervalMinutes=30[0x1e]）+
// 若干文案字段 + `supported`/`processIsElevated` 两项能力探测 → 写 state.Supported/IsAdmin。
func newMemoryReleaseService() *memoryReleaseService {
	s := &memoryReleaseService{
		config: MemoryReleaseConfig{
			Mode:            "standbylist", // 11B 默认（汇编 [0x8]=ptr / [0x10]=0xb）
			TimerEnabled:    false,
			IntervalMinutes: 0x1e, // 30 分钟默认（汇编 0x1e）
		},
	}
	// [S] 三处能力注入（汇编 [0xa0]/[0xa8]/[0xb0] 各一次 funcval 装载，[0xb8]=time.Now）
	s.execute = executeMemoryReleaseMode
	s.supported = memoryReleaseSupported
	s.processIsElevated = memoryReleaseProcessIsElevated
	s.now = time.Now
	// [S] 两处能力探测（汇编 [0x30]/[0x31] 各一次调用）
	s.state.Supported = s.supported()
	s.state.IsAdmin = s.processIsElevated()
	return s
}

// parseMemoryReleaseMode 把外部传入的模式别名规范化为 5 个规范值之一。
// [S 汇编 0x1408d31e0, 736B] 返回 (规范化模式, 是否有效)。
//
// 汇编结构：TrimSpace → ToLower → `lea rcx,[rbx-2]; cmp rcx,0x14; ja default`
// （按**长度**索引跳转表，长度 2..22）→ 每条分支做逐字节 cmp/movabs 精确匹配。
//
// 全量模式表（16 条别名实证，含各行对应的返回常量 VA）：
//
//	len 2  "ws"                    → workingsets
//	len 7  "standby"               → standbylist
//	len 8  "modified"              → modifiedpagelist
//	len 9  "priority0"             → priority0standbylist
//	len 10 "priority-0"            → priority0standbylist
//	len 10 "workingset"            → workingsets
//	len 11 "standbylist"           → standbylist
//	len 11 "workingsets"           → workingsets
//	len 11 "working-set"           → workingsets
//	len 11 "lowpriority"           → priority0standbylist
//	len 12 "low-priority"          → priority0standbylist
//	len 12 "modifiedpage"          → modifiedpagelist
//	len 13 "modified-page"         → modifiedpagelist
//	len 16 "modifiedpagelist"      → modifiedpagelist
//	len 20 "priority0standbylist"  → priority0standbylist
//	len 22 "lowprioritystandbylist"→ priority0standbylist（注意：**不**返回同名值）
//
// 其余长度/取值 → ("", false)（汇编 default 出口 `xor eax,eax; xor ebx,ebx; xor ecx,ecx`）。
func parseMemoryReleaseMode(mode string) (string, bool) {
	lower := strings.ToLower(strings.TrimSpace(mode))
	switch len(lower) {
	case 2:
		// [S 0x1408d321d] cmp word ptr [rax],0x7377 → "ws"
		if lower == "ws" {
			return "workingsets", true
		}
	case 7:
		// [S 0x1408d322d] "stan"+"db"+"y"
		if lower == "standby" {
			return "standbylist", true
		}
	case 8:
		// [S 0x1408d3255] movabs 0x6465696669646f6d → "modified"
		if lower == "modified" {
			return "modifiedpagelist", true
		}
	case 9:
		// [S 0x1408d326e] "priority" + byte'0'
		if lower == "priority0" {
			return "priority0standbylist", true
		}
	case 10:
		// [S 0x1408d3298] "priority" + word 0x302d("-0")；否则 [0x1408d32b3] "workings"+"et"
		if lower == "priority-0" {
			return "priority0standbylist", true
		}
		if lower == "workingset" {
			return "workingsets", true
		}
	case 11:
		// [S 0x1408d32da] 按首字节与 's'(0x73) 二分：
		//   首字节 > 's' → [0x1408d3347] "working-"+"se"+"t" / "workings"+"et"+"s"
		//   首字节 ≤ 's' → [0x1408d32df] "lowprior"+"it"+"y" / "standbyl"+"is"+"t"
		switch lower {
		case "standbylist":
			return "standbylist", true
		case "workingsets", "working-set":
			return "workingsets", true
		case "lowpriority":
			return "priority0standbylist", true
		}
	case 12:
		// [S 0x1408d33a7] "low-prio"+"rity" → priority0；[0x1408d33c6] "modified"+"page"
		if lower == "low-priority" {
			return "priority0standbylist", true
		}
		if lower == "modifiedpage" {
			return "modifiedpagelist", true
		}
	case 13:
		// [S 0x1408d33e7] "modified"+"-pag"+"e"
		if lower == "modified-page" {
			return "modifiedpagelist", true
		}
	case 16:
		// [S 0x1408d340f] "modified"+"pagelist" → 返回 16B 常量 @0x140c5524a
		if lower == "modifiedpagelist" {
			return "modifiedpagelist", true
		}
	case 20:
		// [S 0x1408d3445] lea rbx,[0x140c5da07] + memequal(0x14)
		if lower == "priority0standbylist" {
			return "priority0standbylist", true
		}
	case 22:
		// [S 0x1408d345c] 比较 22B 常量 @0x140c61370("lowprioritystandbylist")，
		// 命中后仍返回 20B 常量 @0x140c5da07("priority0standbylist")。
		if lower == "lowprioritystandbylist" {
			return "priority0standbylist", true
		}
	}
	return "", false
}

// refreshPlatformStateLocked 在锁内重跑两项能力探测并回写 state 头部布尔。
// [S 汇编 0x1408d3ec0, 96B] 实证：
//
//	if s.supported != nil        { s.state.Supported(@0x30) = s.supported() }
//	if s.processIsElevated != nil{ s.state.IsAdmin(@0x31)   = s.processIsElevated() }
//
// 两处均为 nil 守卫下的间接调用（`mov rdx,[rax+0xa8]; test rdx,rdx; je skip; call [rdx]`）。
func (s *memoryReleaseService) refreshPlatformStateLocked() {
	if s.supported != nil {
		s.state.Supported = s.supported()
	}
	if s.processIsElevated != nil {
		s.state.IsAdmin = s.processIsElevated()
	}
}

// currentTime 返回服务时钟当前时间（注入时钟优先，否则 time.Now）。
// [S 汇编 0x1408d3f20, 64B]：`mov rdx,[rax+0xb8]; test rdx,rdx; je → call time.Now`。
func (s *memoryReleaseService) currentTime() time.Time {
	if s.now != nil {
		return s.now()
	}
	return time.Now()
}

// stopTimerLocked 在锁内停止并丢弃当前定时器，同时推进调度代号。
// [S 汇编 0x1408d3f60, 160B] 实证：
//
//	if s.timer != nil { s.timer.Stop(); s.timer = nil }
//	s.scheduleID++          ← **无条件执行**（汇编 `inc qword ptr [rax+0x80]` 在所有分支汇合后）
//
// 注意：Stop() 内部的 `cmp byte ptr [rcx+8],0; je panic` 是 Go 运行时对未初始化
// Timer 的 panic 路径（time.Timer.initTimer），非业务分支。
func (s *memoryReleaseService) stopTimerLocked() {
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	s.scheduleID++
}

// rescheduleLocked 在锁内按当前配置重排下一次定时运行。
// [S 汇编 0x1408d4000, 608B] 实证（now 经 (rbx,rcx,rdi) 三寄存器传入 time.Time）：
//
//	s.stopTimerLocked()
//	active := (!shuttingDown && state.Supported && moduleEnabled) ? config.TimerEnabled : false
//	s.state.SchedulerActive(@0x33) = active
//	if !active || state.Running { s.state.NextRunAt = ""; return }
//	interval := time.Duration(config.IntervalMinutes) * time.Minute   ← imul 0xdf8475800(=60e9ns)
//	nextAt   := now.Add(interval)
//	s.state.NextRunAt = nextAt.Format(time.RFC3339)                   ← 25B @0x140c66a5b
//	gen      := s.scheduleID
//	d        := time.Until(nextAt); if d < 0 { d = 0 }                ← `cmovl rdx,0`
//	s.timer  = time.AfterFunc(d, func(){ s.handleScheduledRun(gen) })
//
// 闭包捕获实证：funcval = {fn:0x1408d4260, s@+8, gen@+0x10}（rescheduleLocked.func1）。
func (s *memoryReleaseService) rescheduleLocked(now time.Time) {
	s.stopTimerLocked()

	active := false
	if !s.shuttingDown && s.state.Supported && s.moduleEnabled {
		active = s.config.TimerEnabled
	}
	s.state.SchedulerActive = active
	if !active || s.state.Running {
		s.state.NextRunAt = ""
		return
	}

	interval := time.Duration(s.config.IntervalMinutes) * time.Minute
	nextAt := now.Add(interval)
	s.state.NextRunAt = nextAt.Format(time.RFC3339)

	gen := s.scheduleID
	d := time.Until(nextAt)
	if d < 0 {
		d = 0
	}
	s.timer = time.AfterFunc(d, func() { s.handleScheduledRun(gen) })
}

// GetState 返回状态快照（含平台探测刷新与调度态重算）。
// [S 汇编 0x1408d3960, 384B] 实证：
//
//	lock → defer unlock（open-coded，闭包 @0x1408d3ae0）→ refreshPlatformStateLocked()
//	active := (!shuttingDown && state.Supported && moduleEnabled) ? config.TimerEnabled : false
//	s.state.SchedulerActive = active
//	if !active && !state.Running { s.state.NextRunAt = "" }
//	return s.state                      ← 0x48 字节经**栈**返回
//
// 注意：并非纯 getter —— 每次调用都会重跑能力探测并重算 SchedulerActive。
func (s *memoryReleaseService) GetState() MemoryReleaseState {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.refreshPlatformStateLocked()

	active := false
	if !s.shuttingDown && s.state.Supported && s.moduleEnabled {
		active = s.config.TimerEnabled
	}
	s.state.SchedulerActive = active
	if !active && !s.state.Running {
		s.state.NextRunAt = ""
	}
	return s.state
}

// RunNow 立即执行内存释放（Wails Binding BootstrapService.RunMemoryRelease 的落点）。
// [S 汇编 0x1408d3b40, 896B] 实证：
//
//	normalized, ok := parseMemoryReleaseMode(mode)
//	if !ok {
//	    return s.GetState(), fmt.Errorf("不支持的内存释放模式: %s", strings.TrimSpace(mode))
//	}
//	return s.runLocked(normalized)
//
// 返回约定实证（两条出口一致）：
//   - MemoryReleaseState 经**栈**返回（`mov [rsp+0x110]` 起 5 段 movups = 0x48 字节，
//     偏移换算到调用者 RunMemoryRelease 的 [rsp+0]）；
//   - error 经 **rax/rbx 寄存器**返回（fmt.Errorf 结果原样保留至 ret）。
func (s *memoryReleaseService) RunNow(mode string) (MemoryReleaseState, error) {
	normalized, ok := parseMemoryReleaseMode(mode)
	if !ok {
		// [S 0x1408d3cde] 错误路径：GetState → TrimSpace(mode) → convTstring →
		// fmt.Errorf(34B @0x140c75bf5, 1 个参数) → 搬 state → ret
		return s.GetState(), fmt.Errorf("不支持的内存释放模式: %s", strings.TrimSpace(mode))
	}
	return s.runLocked(normalized)
}

// runLocked 在锁保护下执行一次内存释放（含运行态编排与结果回写）。
// [S 汇编 0x1408d43c0, 2496B] 实证流程：
//
//	lock → refreshPlatformStateLocked()
//	if shuttingDown        { unlock → return s.state, errors.New("内存释放服务正在关闭" /*30B*/) }
//	if !moduleEnabled      { unlock → return MemoryReleaseState{}, errors.New("内存释放功能当前已关闭" /*33B*/) }
//	if !state.Supported    { unlock → return MemoryReleaseState{}, errors.New("当前平台不支持内存释放功能" /*39B*/) }
//	if  state.Running      { unlock → return s.state, errors.New("内存释放正在执行中" /*27B*/) }
//
//	normalized, ok := parseMemoryReleaseMode(mode)        ← 对入参**再验证一次**
//	if !ok { unlock → return s.state, fmt.Errorf("不支持的内存释放模式: %s", TrimSpace(mode)) }
//
//	gen := ++s.generation; ch := make(chan struct{})
//	state.Running=true; state.LastMode=normalized; state.NextRunAt=""; state.LastError=""
//	s.runningDone = ch; s.stopTimerLocked(); **unlock**（手工解锁，此时尚无 defer）
//
//	err := s.execute(normalized)                          ← 锁外执行
//	now := s.currentTime()
//	lock; defer unlock（open-coded，闭包 @0x1408d4d80 = 纯 Unlock）
//	if s.runningDone == ch { close(ch); s.runningDone = nil }
//	if !shuttingDown && s.generation == gen {             ← 路径 A：正常收尾
//	    s.refreshPlatformStateLocked()
//	    state.Running = false
//	    state.LastMode = normalized
//	    state.LastRunAt = now.Format(time.RFC3339)
//	    if err == nil { state.LastError = "" } else { state.LastError = err.Error() }
//	    s.rescheduleLocked(now)
//	} else {                                              ← 路径 B：陈旧代/已关闭
//	    state.Running = false; state.SchedulerActive = false; state.NextRunAt = ""
//	    // 不刷新 LastRunAt / LastError，不重排定时器
//	}
//	return s.state, err
//
// 返回约定与 RunNow 同构：state 走栈、error 走 rax/rbx。
func (s *memoryReleaseService) runLocked(mode string) (MemoryReleaseState, error) {
	s.lock.Lock()

	s.refreshPlatformStateLocked()

	if s.shuttingDown {
		s.lock.Unlock()
		return s.state, errors.New("内存释放服务正在关闭")
	}
	if !s.moduleEnabled {
		s.lock.Unlock()
		return MemoryReleaseState{}, errors.New("内存释放功能当前已关闭")
	}
	if !s.state.Supported {
		s.lock.Unlock()
		return MemoryReleaseState{}, errors.New("当前平台不支持内存释放功能")
	}
	if s.state.Running {
		s.lock.Unlock()
		return s.state, errors.New("内存释放正在执行中")
	}

	normalized, ok := parseMemoryReleaseMode(mode)
	if !ok {
		s.lock.Unlock()
		return s.state, fmt.Errorf("不支持的内存释放模式: %s", strings.TrimSpace(mode))
	}

	s.generation++
	gen := s.generation
	ch := make(chan struct{})

	s.state.Running = true
	s.state.LastMode = normalized
	s.state.NextRunAt = ""
	s.state.LastError = ""
	s.runningDone = ch

	s.stopTimerLocked()
	s.lock.Unlock()

	// 锁外执行实际释放动作。
	var execErr error
	if s.execute != nil {
		execErr = s.execute(normalized)
	}
	now := s.currentTime()

	s.lock.Lock()
	defer s.lock.Unlock()

	if s.runningDone == ch {
		close(ch)
		s.runningDone = nil
	}

	if !s.shuttingDown && s.generation == gen {
		s.refreshPlatformStateLocked()
		s.state.Running = false
		s.state.LastMode = normalized
		s.state.LastRunAt = now.Format(time.RFC3339)
		if execErr == nil {
			s.state.LastError = ""
		} else {
			s.state.LastError = execErr.Error()
		}
		s.rescheduleLocked(now)
	} else {
		s.state.Running = false
		s.state.SchedulerActive = false
		s.state.NextRunAt = ""
	}

	return s.state, execErr
}

// normalizeMemoryReleaseConfig 对内存释放配置执行归一化：模式规范化、定时器间隔钳制。
// [S 汇编 0x1408d34c0, 160B] 实证：
//
//	parseMemoryReleaseMode(mode) →
//	  !ok → fallback "standbylist"（11B @0x140c47523）
//	timerEnabled 直通、不参与归一化
//	intervalMinutes 钳制：
//	  < 5  → 设为 30（默认值 0x1e）
//	  > 1440 → 设为 1440（0x5a0）
//	  5..1440 保留
//	返回 (归一化模式, timerEnabled, 钳制后的 interval)
func normalizeMemoryReleaseConfig(mode string, timerEnabled bool, intervalMinutes int) (string, bool, int) {
	normalized, ok := parseMemoryReleaseMode(mode)
	if !ok {
		normalized = "standbylist"
	}
	if intervalMinutes < 5 {
		intervalMinutes = 30
	} else if intervalMinutes > 1440 {
		intervalMinutes = 1440
	}
	return normalized, timerEnabled, intervalMinutes
}

// memoryReleaseConfigsEqual 比较两份配置是否等效（双方均经 normalize 归一化）。
// [S 汇编 0x1408d4ee0, 288B] 实证：
//
//	normalize(a.Mode, a.TimerEnabled, a.IntervalMinutes) → (normA, _, clampedA)
//	normalize(b.Mode, b.TimerEnabled, b.IntervalMinutes) → (normB, _, clampedB)
//	len(normA) != len(normB) → false
//	a.TimerEnabled != b.TimerEnabled → false
//	clampedA != clampedB → false
//	len 相等 → runtime.memequal(normA.ptr, normB.ptr, len) → 结果
func memoryReleaseConfigsEqual(a, b MemoryReleaseConfig) bool {
	modeA, timerA, intervalA := normalizeMemoryReleaseConfig(a.Mode, a.TimerEnabled, a.IntervalMinutes)
	modeB, timerB, intervalB := normalizeMemoryReleaseConfig(b.Mode, b.TimerEnabled, b.IntervalMinutes)
	if len(modeA) != len(modeB) {
		return false
	}
	if timerA != timerB {
		return false
	}
	if intervalA != intervalB {
		return false
	}
	return modeA == modeB
}

// Configure 配置内存释放服务（含归一化与变更检测）。
// [S 汇编 0x1408d3620, 736B] 实证流程：
//
//	lock →
//	if shuttingDown { return }  ← 跳过所有配置写入
//	normalizeMemoryReleaseConfig(cfg.Mode, cfg.TimerEnabled, cfg.IntervalMinutes)
//	保存旧 config 与旧 moduleEnabled
//	写入归一化后的新 config
//	moduleEnabled = enabled
//	refreshPlatformStateLocked()
//	memoryReleaseConfigsEqual(oldCfg, newCfg) && enabled == oldModuleEnabled
//	  ├─ 真（未变更）：仅重算 SchedulerActive + 条件清 NextRunAt
//	  └─ 假（已变更）：currentTime → rescheduleLocked
//
//	defer s.lock.Unlock() 由 open-coded 优化路径经 Configure.deferwrap1 调用。
//
// [S 汇编 0x1408d3900, 96B] deferwrap1 = `s.lock.Unlock()` 的闭包装配。
func (s *memoryReleaseService) Configure(cfg MemoryReleaseConfig, enabled bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.shuttingDown {
		return
	}

	// 归一化新配置
	normMode, timerEnabled, clampedInterval := normalizeMemoryReleaseConfig(cfg.Mode, cfg.TimerEnabled, cfg.IntervalMinutes)

	// 捕获旧值用于比较
	oldMode := s.config.Mode
	oldTimerEnabled := s.config.TimerEnabled
	oldInterval := s.config.IntervalMinutes
	oldModuleEnabled := s.moduleEnabled

	// 写入新值
	s.config.Mode = normMode
	s.config.TimerEnabled = timerEnabled
	s.config.IntervalMinutes = clampedInterval
	s.moduleEnabled = enabled

	s.refreshPlatformStateLocked()

	// 变更检测：若新、旧配置归一化后等效且 enabled 未变 → 仅重算调度态，不重排
	oldCfg := MemoryReleaseConfig{Mode: oldMode, TimerEnabled: oldTimerEnabled, IntervalMinutes: oldInterval}
	newCfg := MemoryReleaseConfig{Mode: normMode, TimerEnabled: timerEnabled, IntervalMinutes: clampedInterval}

	if memoryReleaseConfigsEqual(oldCfg, newCfg) && enabled == oldModuleEnabled {
		active := false
		if !s.shuttingDown && s.state.Supported && s.moduleEnabled {
			active = s.config.TimerEnabled
		}
		s.state.SchedulerActive = active
		if !active && !s.state.Running {
			s.state.NextRunAt = ""
		}
		return
	}

	// 配置或 enabled 已变化 → 重排
	now := s.currentTime()
	s.rescheduleLocked(now)
}

// Shutdown 关闭内存释放服务：标记关闭、停定时器、等待运行中任务完成。
// [S 汇编 0x1408d4de0, 256B] 实证流程：
//
//	if s == nil { return }    ← test rax,rax; je ret
//	lock →
//	if !shuttingDown {
//	    shuttingDown = true
//	    generation++
//	    stopTimerLocked()
//	    state.SchedulerActive = false
//	    state.NextRunAt = ""    ← ptr+len 均清 0（含 gc barrier）
//	}
//	runningDone = s.runningDone(@0x98)
//	unlock（手工解锁，无 defer）→
//	if runningDone != nil { <-runningDone }  ← 阻塞到正在执行的释放完成
//
// 注意：解锁后才等待 runningDone，避免死锁（runLocked 需拿锁写 state）。
func (s *memoryReleaseService) Shutdown() {
	if s == nil {
		return
	}
	s.lock.Lock()

	if !s.shuttingDown {
		s.shuttingDown = true
		s.generation++
		s.stopTimerLocked()
		s.state.SchedulerActive = false
		s.state.NextRunAt = ""
	}

	runningDone := s.runningDone
	s.lock.Unlock()

	if runningDone != nil {
		<-runningDone
	}
}

// handleScheduledRun 定时器到期回调：校验调度代号与平台状态后执行一次释放。
// [S 汇编 0x1408d42a0, 288B] 实证流程：
//
//	lock →
//	if shuttingDown         { unlock; return }
//	if scheduleID != gen    { unlock; return }
//	if !state.Supported     { unlock; return }
//	if !moduleEnabled       { unlock; return }
//	if !config.TimerEnabled { unlock; return }
//	mode := config.Mode     ← **锁内取快照**（asm 0x1408d431b/431f 在 unlock 前读 ptr/len）
//	unlock →
//	s.runLocked(mode)
//
// 注意读取时机：asm 证实 Mode 的 ptr/len 在 `lock xadd`（unlock）**之前**装入
// (rbx,rdx)，解锁后经寄存器透传给 runLocked。若在解锁后再读 s.config.Mode，
// 并发写者可能已替换该字段，行为不等价。
func (s *memoryReleaseService) handleScheduledRun(gen uint64) {
	s.lock.Lock()

	if s.shuttingDown {
		s.lock.Unlock()
		return
	}
	if s.scheduleID != gen {
		s.lock.Unlock()
		return
	}
	if !s.state.Supported {
		s.lock.Unlock()
		return
	}
	if !s.moduleEnabled {
		s.lock.Unlock()
		return
	}
	if !s.config.TimerEnabled {
		s.lock.Unlock()
		return
	}

	mode := s.config.Mode
	s.lock.Unlock()
	s.runLocked(mode)
}

// 保留行：memoryReleaseSupported / openCurrentProcessToken / memoryReleaseProcessIsElevated / newMemoryReleaseService / parseMemoryReleaseMode
// 保留行：refreshPlatformStateLocked / currentTime / stopTimerLocked / rescheduleLocked / GetState / RunNow / runLocked
// ---- 以下为批次 18 已落体函数的分界标识，批次 19 新函数已插入上方 ----
