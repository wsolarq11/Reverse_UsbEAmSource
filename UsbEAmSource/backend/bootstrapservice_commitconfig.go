// AUTO-RECONSTRUCTED — commitLauncherConfigReplacementWithWidgets 附属被调函数
// 研究用途 · UsbEAm Launcher 1.0.3 后端方法体还原
//
// 本文件承载主函数 commitLauncherConfigReplacementWithWidgets 的四个被调方：
//
//	prepareWorkspaceDirectories          [S-sig] 创建 WorkspaceLayout 各目录，返回回滚句柄
//	clearPendingFileSearchResidentWarm   [S]     汇编 53L 全量实证
//	launcherConfigStore.ReplacePrepared  [S-sig] 配置原子替换（6 参形态）
//	desktopWidgetService.emitChange      [S-sig] 桌面小部件变更通知
//
// 契约来源：
//
//	pipeline/tmp/commitLauncherConfigReplacementWithWidgets.asm.txt（完整 549 行）
//	pipeline/tmp/CompleteInitialization.asm.txt（第 1 调用点）
//	pipeline/tmp/ImportConfig.asm.txt（第 2 调用点）
//	symbols.main.bak L2147-2151（5 个符号：主函数 + func1 + func2 + func2.1 + deferwrap1）
//
// ---- 汇编全景（549L / 3072B，帧 sub rsp,0x2890）----
//
//	Prologue  L1-37   : 栈守卫 → 保 rax=bs / rbx=可选指针 / cl=initialization →
//	                    零化 [rsp+0x31e0] 0x142 qwords → xmm15 零化
//	Lock      L26-37  : lock cmpxchg [bs+0x518](workspaceTransaction) → 慢路径 lockSlow →
//	                    装配 open-coded defer（[rsp+0x1e18]=deferwrap1, +0x1e20=&mutex）→
//	                    defer_armed([rsp+0x135f])=1
//	Snapshot  L38-63  : newobject(typeof LauncherConfig) → configStoreSnapshot(bs) →
//	                    rax=*launcherConfigStore 存 [rsp+0x13d8]；
//	                    LauncherConfig / DesktopWidgetDocument 走栈返回
//	Storage   L64-151 : normalizeStorageConfig(cfg.Storage 五字段) → [rsp+0x13e0] →
//	                    .eq.main.StorageConfig 比对 → 不等则重新规整装配
//	Layout    L152-168: buildWorkspaceLayoutWithConfig(cfg, target) → [rsp+0x1d68]
//	PathsCmp  L179-290: 5 轮 filepathlite.Clean + runtime.memequal + xor 1 →
//	                    storageChanged 合成位写 [rsp+0x135e]
//	Dirs      L291-358: 第 2 轮 normalizeStorageConfig(layout 五字段) →
//	                    prepareWorkspaceDirectories(layout) → rax 存 [rsp+0x1e10]（回滚句柄）
//	Options   L359-363: launcherConfigOptions(bs) → DefaultTagCatalog → [rsp+0x2840] +
//	                    byte[+0x10]=1（slice 头 + 标志）
//	Closures  L364-394: func1（捕获 initialization byte）→ [rsp+0x1390]；
//	                    func2（捕获 bs / 第 3 参指针 / cfg 指针 / &layout）→ [rsp+0x2858]
//	WidgetDoc L395-401: 零化 [rsp+0x28a0] → rep movsq 0x126 qwords（2352B）+
//	                    byte[+0x10]=1
//	Replace   L401-416: launcherConfigStore.ReplacePrepared(
//	                        rax=store, rbx=&func1, rcx=&func2, rdi=&options,
//	                        esi=1, r8=1, [栈]=2352B 结构)
//	PostRepl  L417-431: rax≠0（error 非 nil）→ 补偿路径 0x14077963d；
//	                    storageChanged → clearPendingFileSearchResidentWarm
//	WidgetNtf L432-448: rdx(第 3 参)==nil → 跳过；bs.desktopWidgets==nil → 跳过；
//	                    notifier(+0x30)==nil → 跳过 selectnbsend；
//	                    否则 runtime.selectnbsend 后再调 emitChange
//	State     L449-474: launcherStateFromCommittedConfig(bs, cfg) → 结果存 [rsp+0x13a0]
//	Return    L474-484: defer 展开（unlock）→ rax/rbx = error → add rsp / pop rbp / ret
//	ErrPlain  L486-508: 补偿路径：先调 [rsp+0x1e10] 回滚句柄 → 展开 defer → 返错
//	ErrPre    L509-525: prepareWorkspaceDirectories 失败：直接展开 defer 返错
//	                    （**不调回滚句柄** —— 目录未建成，无可回滚）
//	Morestack L533-548: runtime.morestack_noctxt → 回跳入口
//
// ---- 调用点差异实证（两处 ecx 相反）----
//
//	CompleteInitialization.asm 0x140778ad0-0x140778ada:
//	    mov rax, rdx        ; bs
//	    xor ebx, ebx        ; 第 3 参 = nil
//	    mov ecx, 1          ; initialization = true
//	    call commitLauncherConfigReplacementWithWidgets
//
//	ImportConfig.asm 0x14077bec0-0x14077becd:
//	    mov rax, [rsp+0x30e0]  ; bs
//	    mov rbx, rdx           ; 第 3 参 = 非 nil
//	    xor ecx, ecx           ; initialization = false
//	    call commitLauncherConfigReplacementWithWidgets
//
// 第 3 参的 nil 守卫落在主函数 0x1407794e5，故其语义为「可选指针」：
// nil 时跳过桌面小部件变更通知（CompleteInitialization 路径），
// 非 nil 时走完整的 notifier 通道 + emitChange（ImportConfig 路径）。
// 本重建树以 configStoreSnapshot 的第三返回值承载该语义（见主函数体）。
//
// 还原口径：可编译 + 功能一致（非字节级同哈希）。档位标注：
//
//	[S]     反汇编实证体（函数体逐条对位）
//	[S-sig] 签名实证，体待对应专项域还原
//	[P]     骨架/占位
package main

import (
	"errors"
	"os"
	"strings"
)

// ---- prepareWorkspaceDirectories ----

// prepareWorkspaceDirectories 创建 WorkspaceLayout 中各目录，返回回滚句柄。
// [S 汇编 0x1409f3760, source_funcs.txt:4886 记为 24 行]。
//
// 调用点 commitLauncherConfigReplacementWithWidgets.asm 0x140779333：
//
//	0x14077930c-0x14077932a  duffcopy 铺入参（WorkspaceLayout 0xa0）
//	0x140779333               call prepareWorkspaceDirectories
//	0x140779340               test rbx,rbx   ← error.data 判空
//	0x140779343               jne 0x1407796c3 ← 失败路径（展开 defer 直接返错）
//	0x140779349               mov [rsp+0x1e10], rax ← 成功路径保存回滚句柄
//
// 回滚句柄的调用形态实证（0x14077963d-0x140779658，即 ReplacePrepared 失败后的补偿路径）：
//
//	mov rdx, [rsp+0x1e10]   ; 句柄
//	mov rax, [rdx]          ; funcval 代码指针
//	call rax
//
// 即 rax 是 Go funcval 指针（[0]=代码入口，rdx=捕获环境）—— 返回类型为 func()。
//
// 关键语义（与 AGENTS 第 10 条「跨域补偿」对应）：
//   - prepareWorkspaceDirectories **自身失败**时不调回滚句柄（0x140779343 分支直达 defer 展开）：
//     目录尚未建成，无可回滚；
//   - 只有 **ReplacePrepared 失败**时（0x14077963d）才调回滚句柄：目录已建成但配置未提交，
//     需回滚已创建的目录。
//
// [S] 汇编 0x1409f3760：仅遍历 6 个目录
// （Root/IconDir/IndexDir/ScreenshotDir/WebView2Dir/BackgroundDir，栈偏移 0x00/0x50/0x60/0x70/0x80/0x90）。
// created=make([]string,0,6)；cleanup(func1 0x1409f3b80)=逆序 os.Remove(created[i])。
// 每目录：TrimSpace 空则 continue；os.Stat 非 ErrNotExist 错误→cleanup()+return(nil,err)；
// ErrNotExist→append(created,trimmed)；os.MkdirAll(trimmed,0755) 错误→cleanup()+return(nil,err)。
func prepareWorkspaceDirectories(ws WorkspaceLayout) (func(), error) {
	dirs := []string{
		ws.Root,
		ws.IconDir,
		ws.IndexDir,
		ws.ScreenshotDir,
		ws.WebView2Dir,
		ws.BackgroundDir,
	}
	created := make([]string, 0, 6)
	cleanup := func() {
		for i := len(created) - 1; i >= 0; i-- {
			os.Remove(created[i])
		}
	}
	for _, d := range dirs {
		trimmed := strings.TrimSpace(d)
		if trimmed == "" {
			continue
		}
		if _, err := os.Stat(trimmed); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				created = append(created, trimmed)
			} else {
				cleanup()
				return nil, err
			}
		}
		if err := os.MkdirAll(trimmed, 0o755); err != nil {
			cleanup()
			return nil, err
		}
	}
	return cleanup, nil
}

// ---- clearPendingFileSearchResidentWarm ----

// clearPendingFileSearchResidentWarm 清除 pending 文件搜索常驻温标记。
// [S 汇编 0x14079fac0, 53L]（va_map_fixed2.txt:304）全量反汇编实证：
//
//	0x14079fac0  cmp rsp,[r14+0x10]        ; 栈守卫
//	0x14079fac6  push rbp / mov rbp,rsp / sub rsp,0x20
//	0x14079face  test rax,rax              ; bs == nil 快返回
//	0x14079fad1  je 0x14079fb3a
//	0x14079fad4  lea rcx,[rax+0x540]       ; &bs.lock
//	0x14079fade  xor eax,eax / mov esi,1
//	0x14079fae5  lock cmpxchg [rdx+0x540],esi   ; 快路径加锁
//	0x14079faf4  jne 0x14079fb12
//	0x14079fb03  call sync.Mutex.lockSlow       ; 慢路径
//	0x14079fb12  mov byte ptr [rdx+0x508],0     ; pendingFileSearchResidentWarm = false
//	0x14079fb1b  mov ebx,0xffffffff
//	0x14079fb20  lock xadd [rdx+0x540],ebx      ; 解锁
//	0x14079fb2f  call sync.Mutex.unlockSlow
func (bs *BootstrapService) clearPendingFileSearchResidentWarm() {
	if bs == nil {
		return
	}
	bs.lock.Lock()
	bs.pendingFileSearchResidentWarm = false
	bs.lock.Unlock()
}

// ---- launcherConfigStore.ReplacePrepared ----

// ReplacePrepared 原子替换已准备态配置（含标签目录选项与桌面小部件文档载荷）。
// [S-sig 汇编 0x140899980, source_funcs.txt:2150 记为 255 行]。
//
// 调用点 commitLauncherConfigReplacementWithWidgets.asm 0x140779442-0x14077946a 寄存器装配：
//
//	mov rax, [rsp+0x13d8]   ; rax = *launcherConfigStore（receiver）
//	lea rbx, [rsp+0x1390]   ; rbx = &func1 闭包（捕获 initialization byte）
//	lea rcx, [rsp+0x2858]   ; rcx = &func2 闭包（捕获 bs / 第 3 参 / cfg 指针 / &layout）
//	lea rdi, [rsp+0x2840]   ; rdi = &LauncherConfigOptions（DefaultTagCatalog slice 头 + 标志）
//	mov esi, 1              ; 第 4 参 bool = true
//	mov r8, rsi             ; 第 5 参 bool = true
//	[栈 0x0..]              ; 第 6 参：rep movsq 0x126 qwords = 2352B 的结构 + byte[+0x10]=1
//	call launcherConfigStore.ReplacePrepared
//
// 第 6 参 2352B 与 DesktopWidgetDocument 尺寸一致（见 bootstrapservice_config.go:49 的
// 尺寸实证注记），且其 byte[+0x10]=1 标志位由 0x1407792fb 写入。
//
// [P] 体待 launcherConfigStore 域按 0x140899980 专项还原。当前实现保持调用链形状：
//
//	func1（prepare 阶段）：入参为待提交配置，返回错误可中止替换；
//	func2（commit 阶段）：替换完成后回调（用于运行时状态同步）。
//
// [S-sig 签名校正（批次 134）] prepare 首参为 initialized bool——
// 汇编 0x140899c14 call prepare 前 al=loadUnlocked 返回的初始化标志；
// func1/CAS 闭包入口 test al，al==0 → "配置尚未初始化"。
func (s *launcherConfigStore) ReplacePrepared(
	prepare func(bool, LauncherConfig) error,
	commit func(*LauncherConfig) error,
	options LauncherConfigOptions,
	withIcons bool,
	saveWidgets bool,
	widgets DesktopWidgetDocument,
) error {
	if s == nil {
		return nil
	}
	_ = options
	_ = withIcons
	_ = saveWidgets
	_ = widgets

	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.loadUnlocked()
	if err != nil {
		return err
	}
	if prepare != nil {
		// [S-sig] initialized 以 cfg.Initialized 承载（asm al=loadUnlocked 独立返回，
		// 语义同"配置已初始化"标志；待 loadUnlocked 签名专项校正后对齐）。
		if err := prepare(cfg.Initialized, cfg); err != nil {
			return err
		}
	}
	if commit != nil {
		if err := commit(&cfg); err != nil {
			return err
		}
	}
	return s.savePreparedUnlocked(withIcons, cfg)
}

// ---- desktopWidgetService.emitChange ----

// emitChange 发出桌面小部件变更通知。
// [S-sig 汇编 0x1407beb20]。
//
// 调用点 commitLauncherConfigReplacementWithWidgets.asm 0x140779535-0x140779554：
//
//	mov rax, [r10+0x440]    ; rax = bs.desktopWidgets（receiver）
//	mov rsi, [rdx+8]        ; rsi = 第 3 参的 +8 字段
//	xor ecx,ecx / mov rdi,rcx
//	lea r8, [rip+0x4c2bbe]  ; r8 = 类型描述符
//	mov r9d, 8              ; r9 = 载荷宽度 8B
//	call desktopWidgetService.emitChange
//
// 前置条件（0x140779506-0x140779520）：service.notifier(+0x30) 非 nil 时，
// 先对 notifier(+0x8) 执行 runtime.selectnbsend 做**非阻塞**通道通知。
//
// [P] 体待 desktopwidget notify/event 域专项还原。当前为空壳：
// 通道侧的非阻塞通知已由调用方（主函数体）承担，本方法只保留调用形状。
func (s *desktopWidgetService) emitChange(payload interface{}) {
	if s == nil {
		return
	}
	_ = payload
	// [P] 推测语义：把载荷封装为 app 事件向前端广播。
	// 待 0x1407beb20 反汇编专项确认事件名与载荷类型后回填。
}
