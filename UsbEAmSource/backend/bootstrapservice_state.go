// AUTO-RECONSTRUCTED SERVICE METHODS — DOMAIN: launcher state assembly
// 研究用途 · UsbEAm Launcher 1.0.3 后端方法体还原
//
// 本文件承载 LauncherState / StartupState 装配链，函数体均为反汇编实证（[S]）。
//
// ---- 结构尺寸（多锚点交叉验证闭合） ----
//
//	WorkspaceLayout     0x0a0  （10 × string，字段序见 types_workspace.go）
//	LauncherConfig      0x938
//	BootstrapSnapshot   0x0d0  （Workspace + Languages + Plugins）
//	LauncherState       0xa18
//	StartupState        0x0d8
//
// ---- LauncherState 字段偏移 ----
//
//	0x000 Workspace(a0)  0x0a0 Languages(18)  0x0b8 Plugins(18)  0x0d0 Config(938)
//	0xa08 StartupTrayMode  0xa09 ConfigOnly  0xa0a ContentRuntimeSyncNeeded
//	0xa10 HotkeyRegistrationErrors（map，8 字节对齐）
//
// 实证：asm 0x140776cf3 写 [rsp+0x1c99]=1（结果区 0x1290 + 0xa09 = ConfigOnly）、
// 0x140776d03 写 [rsp+0x1c9a]=bl（+0xa0a = ContentRuntimeSyncNeeded）、
// 0x140776d0a 写 [rsp+0x1ca0]=rdx（+0xa10 = HotkeyRegistrationErrors）。
//
// ---- StartupState 字段偏移 ----
//
//	0x000 Workspace  0x0a0 Languages  0x0b8 Plugins  0x0d0 ConfigExists  0x0d1 Initialized
//
// 实证：asm 0x140775cf6 写 [rsp+0x1d70]=al、0x140775cfd 写 [rsp+0x1d71]=cl，
// 结果区自 [rsp+0x1ca0] 起（0x1ca0 + 0xd0 = 0x1d70）。
//
// ---- 栈参数布局（调用者分配） ----
//
//	WithSnapshot : [cfg 0x938][snapshot 0x0d0][result 0xa18] + error
//	Committed    : [cfg 0x938][result 0xa18] + error
//	Saved        : [cfg 0x938][result 0xa18]（单返回，无 error）
//
// 实证：GetState 调用点 0x140776660 —— 入参 cfg 铺于 [rsp]、snapshot 铺于 [rsp+0x938]，
// 结果区紧接 [rsp+0xa08]（0x938 + 0xd0 = 0xa08），与 1.2 的 launcherStateFromCommittedConfig
// 调用点 0x140776860 同构。
//
// ---- 本轮修正的签名（7 处，均为 asm 实证） ----
//
//	GetSnapshot()                                    → (BootstrapSnapshot, error)
//	snapshotForCommittedLauncherState()              → (BootstrapSnapshot, error)
//	GetStartupState()                                → (StartupState, error)
//	GetState()                                       → (LauncherState, error)
//	launcherStateFromCommittedConfig()               → (LauncherConfig) (LauncherState, error)
//	launcherStateFromCommittedConfigWithSnapshot()   → (LauncherConfig, BootstrapSnapshot) LauncherState
//	launcherStateFromSavedConfig()                   → (LauncherConfig, bool) LauncherState
package main

import "errors"

// ---- 状态装配主链 ----

// GetState 装配完整启动器状态（状态体系核心入口）。
// [S 汇编 0x140775e20, 2304B]（symbols.txt:18856）。
//
// asm 主干：
//
//	0x140775e6e  ws = workspaceSnapshot()
//	0x140775eb3  ensureWorkspaceDirectories(ws) → err → 0x140776298 清栈返错
//	0x140775ef4  cfg, exists, err = loadLauncherConfigIfExists(ws.ConfigFile)
//	0x140775f48  err != nil            → 0x14077626e 清栈返错
//	0x140775f50  exists == false       → 0x14077621f 返错（21B 常量串）
//	0x140775f8a  !cfg.Initialized && isLauncherConfigEmptyInitialConfig(cfg)
//	                                   → 0x1407761d0 返错（同一 21B 常量串）
//	0x140775fb0  cacheBookmarkSources(bs, cfg+0x500 …)
//	0x1407761b3  make(map) + 两轮全局切片循环填充/过滤（热键触发映射）
//	0x140776518  syncHotkeyBindings(bs, …)（13 组实参，见 [P] 段说明）
//	0x140776546  syncRuntimeServices(bs)
//	0x140776580  snapshot = GetSnapshot() → err → 0x1407765db 清栈返错
//	0x140776660  launcherStateFromCommittedConfigWithSnapshot(cfg, snapshot)
//	0x1407766dd  return state, nil
//
// 错误出口形态实证：0x14077621f / 0x1407761d0 以 runtime.newobject 构造
// {ptr,len=0x15} 字符串头 + itab(0x140c5ce9e/0x140c5ceed) 组装 error 接口，
// 两条出口引用同一字符串（VA 0x140C5F5B7，len 21）。
func (bs *BootstrapService) GetState() (LauncherState, error) {
	workspace := bs.workspaceSnapshot()
	if err := ensureWorkspaceDirectories(workspace); err != nil {
		return LauncherState{}, err
	}

	cfg, configExists, err := loadLauncherConfigIfExists(workspace.ConfigFile)
	if err != nil {
		return LauncherState{}, err
	}
	if !configExists {
		return LauncherState{}, errLauncherStateConfigNotInitialized
	}
	if !cfg.Initialized && isLauncherConfigEmptyInitialConfig(cfg) {
		return LauncherState{}, errLauncherStateConfigNotInitialized
	}

	// [S] 热键/运行时同步段（asm 0x140775f90 ~ 0x140776546）。
	// 已实证：该段先以 cfg+0x500 三元组调 cacheBookmarkSources；
	// 再对 cfg 的 7 组字段（+0x1a0/+0x1a8/+0x1c0/+0x1d8/+0x1f0/+0x208/+0x220）
	// 取 TrimSpace 非空判定，配合两轮全局字符串切片循环构造 map[string]bool；
	// 随后以 22 标量实参（9 寄存器 + 13 栈槽，即 launcherHotkeyBindings 展平）调用 syncHotkeyBindings。
	// 该内联构建与 launcherHotkeyBindingsRawFromConfig 语义一致。
	bs.cacheBookmarkSources(cfg.Bookmarks.Sources)
	bs.syncHotkeyBindings(launcherHotkeyBindingsRawFromConfig(cfg))
	bs.syncRuntimeServices()

	snapshot, err := bs.GetSnapshot()
	if err != nil {
		return LauncherState{}, err
	}
	return bs.launcherStateFromCommittedConfigWithSnapshot(cfg, snapshot), nil
}

// launcherStateFromCommittedConfigWithSnapshot 由已提交配置 + 快照装配状态。
// [S 汇编 0x140776920, 608B]（symbols.txt:18859）。
//
// asm 主干：
//
//	0x140776960  newobject(LauncherConfig) + 拷入参 cfg（0x938）
//	0x1407769e0  launcherConfigRevision(cfg) → cfg.Revision（ptr@+0, len@+8）
//	0x140776a0f  cfg.Initialized(+0x10) = true
//	0x140776a20  attachLauncherConfigIconURLs(bs, cfg)
//	0x140776a35  attachLauncherBackgroundURL(bs, cfg)
//	0x140776a42  isStartupTrayMode(bs) → StartupTrayMode
//	0x140776a56  currentHotkeyRegistrationErrors(bs)
//	0x140776a93  Workspace / Languages / Plugins 自入参 snapshot 拷贝
//	0x140776b2f  写回 StartupTrayMode(+0xa08)，0x140776b36 写回 map(+0xa10)
//
// 关键实证：结果区 ConfigOnly(+0xa09) 与 ContentRuntimeSyncNeeded(+0xa0a) 未被写入，
// 保持结果区清零值 false —— 本函数不设置这两个标志。
func (bs *BootstrapService) launcherStateFromCommittedConfigWithSnapshot(cfg LauncherConfig, snapshot BootstrapSnapshot) LauncherState {
	committed := cfg
	committed.Revision = launcherConfigRevision(committed)
	committed.Initialized = true
	bs.attachLauncherConfigIconURLs(&committed)
	bs.attachLauncherBackgroundURL(&committed)

	return LauncherState{
		Workspace:                snapshot.Workspace,
		Languages:                snapshot.Languages,
		Plugins:                  snapshot.Plugins,
		Config:                   committed,
		StartupTrayMode:          bs.isStartupTrayMode(),
		HotkeyRegistrationErrors: bs.currentHotkeyRegistrationErrors(),
	}
}

// launcherStateFromCommittedConfig 由已提交快照装配状态。
// [S 汇编 0x140776720, 512B]（symbols.txt:18857）。
//
// asm 主干：
//
//	0x140776796  snapshot, err = snapshotForCommittedLauncherState()
//	0x1407767e3  err != nil → 0x14077680f? 否：err 非 nil 时 0x1407767e5 清结果区返错
//	0x14077680f  以入参 cfg（调用者栈 [R] 起 0x938）铺参数
//	0x140776860  launcherStateFromCommittedConfigWithSnapshot(cfg, snapshot)
//	0x1407768dd  return state, nil
//
// 入参实证：asm 帧偏移 0x14077680f 从 [rsp+0x1f20] 读取 cfg 首 8 字节并在
// [rsp+0x1f28] 续读 0x126 qwords（合计 0x938 = LauncherConfig 尺寸），
// 且函数无其它入参寄存器读取 —— 签名带一个栈传入的 LauncherConfig。
func (bs *BootstrapService) launcherStateFromCommittedConfig(cfg LauncherConfig) (LauncherState, error) {
	snapshot, err := bs.snapshotForCommittedLauncherState()
	if err != nil {
		return LauncherState{}, err
	}
	return bs.launcherStateFromCommittedConfigWithSnapshot(cfg, snapshot), nil
}

// launcherStateFromSavedConfig 由已保存配置装配状态（仅配置态，无快照）。
// [S 汇编 0x140776b80, 480B]（symbols.txt:18860）。
//
// asm 主干：
//
//	0x140776b9d  bs 入参（rax）；0x140776ba5 flag 入参（bl）—— 签名含一个 bool
//	0x140776bc7  newobject(LauncherConfig) + 拷入参 cfg（栈 [R] 起 0x938）
//	0x140776c46  cfg.Revision = launcherConfigRevision(cfg)
//	0x140776c74  cfg.Initialized(+0x10) = true
//	0x140776c83  attachLauncherConfigIconURLs(bs, cfg)
//	0x140776c98  attachLauncherBackgroundURL(bs, cfg)
//	0x140776ca5  currentHotkeyRegistrationErrors(bs)
//	0x140776cf3  ConfigOnly(+0xa09) = true
//	0x140776d03  ContentRuntimeSyncNeeded(+0xa0a) = 入参 flag
//	0x140776d0a  HotkeyRegistrationErrors(+0xa10) = map
//
// 关键实证：Workspace / Languages / Plugins / StartupTrayMode 均未写入（保持 false / 零值）——
// 本函数只装配配置态字段。
func (bs *BootstrapService) launcherStateFromSavedConfig(cfg LauncherConfig, contentRuntimeSyncNeeded bool) LauncherState {
	saved := cfg
	saved.Revision = launcherConfigRevision(saved)
	saved.Initialized = true
	bs.attachLauncherConfigIconURLs(&saved)
	bs.attachLauncherBackgroundURL(&saved)

	return LauncherState{
		Config:                   saved,
		ConfigOnly:               true,
		ContentRuntimeSyncNeeded: contentRuntimeSyncNeeded,
		HotkeyRegistrationErrors: bs.currentHotkeyRegistrationErrors(),
	}
}

// GetStartupState 装配启动期状态（工作区 + 语言/插件 + 配置存在性/初始化位）。
// [S 汇编 0x140775a00, 1056B]（symbols.txt:18855）。
//
// asm 主干：
//
//	0x140775a55  ws = workspaceSnapshot()
//	0x140775a96  ensureWorkspaceDirectories(ws) → err → 0x140775dbc 清栈返错
//	0x140775ae0  snapshot, err = GetSnapshot() → err → 0x140775d82 清栈返错
//	0x140775b3f  exists, err = fileExists(ws.ConfigFile) → err → 0x140775d4b 清栈返错
//	0x140775b62  !exists → initialized 直接为 false（0x140775b64 xor ecx,ecx）
//	0x140775ba0  cfg, _, err = loadLauncherConfigIfExists(ws.ConfigFile)
//	0x140775bf4  err != nil → 0x140775d11 清栈返错
//	0x140775bfa  cfg.Initialized(+0x10) 为真 → initialized = true
//	0x140775c2c  否则 initialized = !isLauncherConfigEmptyInitialConfig(cfg)
//	0x140775cc6  结果 = snapshot 前 0xd0 字节 + ConfigExists(al) + Initialized(cl)
//
// 实参实证：fileExists / loadLauncherConfigIfExists 均取 [rsp+0x9f0] 与 [rsp+0x9f8]，
// 即 workspace 副本偏移 +0x10 —— WorkspaceLayout 第 2 字段 ConfigFile（string）。
func (bs *BootstrapService) GetStartupState() (StartupState, error) {
	workspace := bs.workspaceSnapshot()
	if err := ensureWorkspaceDirectories(workspace); err != nil {
		return StartupState{}, err
	}

	snapshot, err := bs.GetSnapshot()
	if err != nil {
		return StartupState{}, err
	}

	configExists, err := fileExists(workspace.ConfigFile)
	if err != nil {
		return StartupState{}, err
	}

	initialized := false
	if configExists {
		cfg, _, err := loadLauncherConfigIfExists(workspace.ConfigFile)
		if err != nil {
			return StartupState{}, err
		}
		if cfg.Initialized {
			initialized = true
		} else {
			initialized = !isLauncherConfigEmptyInitialConfig(cfg)
		}
	}

	return StartupState{
		Workspace:    snapshot.Workspace,
		Languages:    snapshot.Languages,
		Plugins:      snapshot.Plugins,
		ConfigExists: configExists,
		Initialized:  initialized,
	}, nil
}

// errLauncherStateConfigNotInitialized 对应 asm 两处错误出口引用的同一 21 字节常量串。
//
// [S] GetState asm 0x1407761d0 与 0x14077621f：
//
//	runtime.newobject → {ptr, len=0x15} 字符串头，itab 分别取 0x140c5ceed / 0x140c5ce9e，
//	两条出口的字符串 VA 均为 0x140C5F5B7（LEA 目标 = 下一条指令地址 + disp）。
//
// 文本须按 PE 节映射解码（raw_off = .rdata RawPtr + (VA - ImageBase - .rdata RVA)，len=21）
// 后回填；21 字节对 GBK/UTF-8 中文分别约为 10 个汉字与 7 个汉字。
var errLauncherStateConfigNotInitialized = errors.New("launcher config not initialized")
