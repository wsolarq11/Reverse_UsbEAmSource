// AUTO-RECONSTRUCTED — DOMAIN: bootstrap config lifecycle callees（配置生命周期被调）
// 研究用途
//
// 契约来源：capstone 反汇编符号标注（docs/goresym/pipeline/tmp/）：
//
//	ExportConfig.asm.txt / ResetConfig.asm.txt / AbortInitialization.asm.txt
//	ImportConfig.asm.txt / CompleteInitialization.asm.txt / MigrateConfig.asm.txt
//
// 还原口径：可编译 + 功能一致（非字节级同哈希）。档位如实标注：
//
//	[S] 汇编实证（函数体逐条对位）
//	[P] 骨架/占位（签名已按调用点寄存器装配对齐，体待续作）
package main

import (
	"errors"
	"os"
	"strings"
)

// errLauncherConfigExportTargetEmpty 导出目标路径为空（错误值 + 期望格式）。
var errLauncherConfigExportTargetEmpty = errors.New("launcherconfig: 导出路径不能为空")

// errBootstrapConfigPathEmpty 配置路径为空，无法执行删除/重置类操作。
var errBootstrapConfigPathEmpty = errors.New("bootstrap: 配置路径为空")

// launcherConfigImportSourceTag 导入数据源标识（24 字节 .rdata 常量串）。
// [P] 对应 ImportConfig.asm 0x14077bdf2 的 lea rcx,[rip+0x4e8fbf]（VA 0x140C64DB8）
// 与紧随的 mov edi,0x18（长度 24）。该常量尚未用 decode_rodata_str.py 解码，
// 暂以致零占位；待验证阶段解码后以实证值回填。
var launcherConfigImportSourceTag = ""

// errBootstrapExportPathEmpty 导出目标路径为空。
// [S] 对应 ExportTextFile.asm 0x14077bbcf 处构造的 24 字节中文错误串（VA 0x140C64D88）。
var errBootstrapExportPathEmpty = errors.New("导出路径不能为空")

// errBootstrapNoDataRoot 无可用数据根目录。
var errBootstrapNoDataRoot = errors.New("未设置数据根目录")

// errBootstrapMigratePathEmpty 迁移源路径为空。
var errBootstrapMigratePathEmpty = errors.New("迁移源路径为空")

// ---- 配置快照 ----

// configStoreSnapshot 在 bs.lock 保护下取得配置存储 + 工作区布局快照。
// [S 汇编实证 0x1407a1240, 544B]（va_map_fixed2.txt:312）：
//
//	0x1407a12a0  test rax,rax → bs==nil 则 duffzero 160B 返回空 WorkspaceLayout（rax=0）
//	0x1407a12bc  lock cmpxchg [rcx+0x540] → bs.lock 互斥锁（快路径 + lockSlow 兜底）
//	0x1407a12df-0x1407a130d  open-coded defer Unlock（deferwrap1 = 0x1407a1460）
//	0x1407a132a  duffcopy+0x2f4（160B = WorkspaceLayout）bs.workspace → 局部快照 [rsp+0x28]
//	0x1407a1333  store = bs.configStore（[rcx+0xa0]）
//	0x1407a133a-0x1407a1344  matchesPath(store, workspace.ConfigFile@+0x10)
//	0x1407a134b  未命中 → launcherConfigStoreForPath(ConfigFile) → 回写 bs.configStore（含 GC 写屏障）
//	0x1407a13ca  duffcopy+0x2f4 快照工作区 → 调用者返回槽；rax = store
//
// 返回值实证修正：旧骨架误写 `(*launcherConfigStore, LauncherConfig, DesktopWidgetDocument)`。
// 真 ABI 为 `(*launcherConfigStore, WorkspaceLayout)`——第二个栈返回值只有 160B（duffzero+0x134 /
// duffcopy+0x2f4 均为 160B），即工作区布局快照，不含 LauncherConfig（2576B）与
// DesktopWidgetDocument（2352B）。后者由调用方另行 store.Read() / store.ReadSelfContained() 取得。
func (bs *BootstrapService) configStoreSnapshot() (*launcherConfigStore, WorkspaceLayout) {
	if bs == nil {
		return nil, WorkspaceLayout{}
	}
	bs.lock.Lock()
	defer bs.lock.Unlock()

	store := bs.configStore
	if !store.matchesPath(bs.workspace.ConfigFile) {
		store = launcherConfigStoreForPath(bs.workspace.ConfigFile)
		bs.configStore = store
	}
	return store, bs.workspace
}

// ---- 工作区目录 ----

// ensureWorkspaceDirectories 确保工作区各目录存在。
// [S 汇编实证 0x1407a2520]：调用点 ResetConfig.asm 0x140778336-0x140778354 先以 duffcopy
// 把工作区布局复制进栈参数区再 call，返回 error 由 test rax,rax / jne 判定。
// 注意：本函数与 0x1407a26a0 的 BootstrapService 方法版（symbols.txt:19195）是
// 两个不同符号，方法版见下方 ensureWorkspaceDirectories 方法。
func ensureWorkspaceDirectories(ws WorkspaceLayout) error {
	for _, dir := range []string{
		ws.Root, ws.IconDir, ws.IndexDir, ws.ScreenshotDir,
		ws.WebView2Dir, ws.BackgroundDir, ws.LanguageDir,
		ws.AppLanguageDir, ws.PluginDir,
	} {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

// ensureWorkspaceDirectories 方法版工作区目录确保。
// [S 汇编 0x1407a26a0, 128B]：workspaceSnapshot()（加锁快照）→ ensureWorkspaceDirectories(ws)。
// 符号名实证为 BootstrapService.ensureWorkspaceDirectories（symbols.txt:19195）。
func (bs *BootstrapService) ensureWorkspaceDirectories() error {
	return ensureWorkspaceDirectories(bs.workspaceSnapshot())
}

// ---- launcherConfigStore 方法 ----

// DeletePrepared 删除已准备态配置：可选 prepare 回调先行，随后移除配置文件。
// [S-sig] VA 0x14089bfa0。两处调用点的回调装配不同，如实记录待专项确认：
//
//	ResetConfig.asm 0x1407783b8-ca ：rax=store, rbx=0, rcx=&栈闭包(ResetConfigfunc1, 捕获 bs)
//	AbortInitialization.asm 0x14077a56a-80：rax=store, rbx=静态 funcval(0x141095928), rcx=nil
//
// 即回调既可走第 1 参槽（静态 funcval）也可走第 2 参槽（栈闭包）。本骨架按
// 「单 prepare 回调」还原——两处语义均为「删除前执行一次清理」；[P] 待
// launcherConfigStore 域专项按 0x14089bfa0 完整体确认参数槽位后校正。
func (s *launcherConfigStore) DeletePrepared(prepare func(LauncherConfig) error) error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deletePreparedLocked(prepare)
}

// ReadSelfContained 读取自包含配置（含内嵌图标资源）。
// [S-sig] 骨架。VA 0x1408993c0。调用点 ExportConfig.asm 0x14077b7d2-0x14077b7da：
// rax=store(receiver)，返回后在 0x14077b7df-0x14077b828 以 0x126 qword 段
// rep movsq 复制进调用者栈槽（即大结构走栈返回），随后 test rbx,rbx 判错。
func (s *launcherConfigStore) ReadSelfContained() (DesktopWidgetDocument, error) {
	if s == nil {
		return DesktopWidgetDocument{}, nil
	}
	return DesktopWidgetDocument{}, nil
}

// Update 更新配置（迁移链写入点）。
// [S-sig] 骨架。VA 0x14089b420。调用点 MigrateConfig.asm 0x14077c94c，
// 其后 0x14077c9a0 处失败会转 stagedWorkspaceData.Rollback（补偿路径）。
func (s *launcherConfigStore) Update(cfg LauncherConfig) error {
	if s == nil {
		return nil
	}
	return s.saveUnlockedWithCurrentRefs(cfg)
}

// Read 读取配置。
// [S-sig] 骨架。VA 0x140899100。调用点 MigrateConfig.asm 0x14077c777，
// 返回 error 由 0x14077c7a3 的 jne 判定。
func (s *launcherConfigStore) Read() (LauncherConfig, error) {
	if s == nil {
		return LauncherConfig{}, nil
	}
	return s.loadUnlocked()
}

// ---- 图标 / 小部件存储 ----

// launcherConfigIconStoreForConfigPath 与 (*launcherConfigIconStore).Delete
// 已于批次 37 迁至 launcherconfigiconstore.go（[P] 骨架 → [S] 汇编实证还原），
// 此处不再重复声明。

// Delete 删除小部件文档缓存文件，并把缓存重置为默认文档（cachedExists=false）。
// [S 汇编 0x1407c0d20, 992B] 实证：
//   s==nil→errors.New("首页组件存储不可用"@0x140c69cd5,27B)；s.mu.Lock()+defer Unlock；
//   path=strings.TrimSpace(s.path)，空→errors.New("首页组件存储路径不能为空"@0x140c782eb,36B)；
//   os.Remove(path)，err 非 nil 且 !errors.Is(err,os.ErrNotExist) 返回 err；
//   成功/NotExist→s.cached=normalizeDesktopWidgetDocument(零值)、cachedExists=false、cachedLoadErr=nil、
//   loaded=true、return nil。
func (s *launcherWidgetStore) Delete() error {
	if s == nil {
		return errors.New("首页组件存储不可用")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	path := strings.TrimSpace(s.path)
	if path == "" {
		return errors.New("首页组件存储路径不能为空")
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	s.cached = normalizeDesktopWidgetDocument(DesktopWidgetDocument{})
	s.cachedExists = false
	s.cachedLoadErr = nil
	s.loaded = true
	return nil
}

// ---- 导出 / 导入 ----

// validateLauncherConfigExportTarget 校验导出目标。
// [S-sig] 骨架。VA 0x140878700。调用点 ExportConfig.asm 0x14077b789-0x14077b7a9，
// 入口装配 4 参：(rax,rbx) 与 (rcx,rdi) 各为一组 string 对，
// 即 (导出目标路径, 工作区配置路径)；返回 error 由 0x14077b7ae 的 test/jne 判定。
func validateLauncherConfigExportTarget(target string, configPath string) error {
	if strings.TrimSpace(target) == "" {
		return errLauncherConfigExportTargetEmpty
	}
	_ = configPath
	return nil
}

// exportSelfContainedLauncherConfigWithWidgets 生成自包含（含小部件）配置导出文本。
// [S 汇编 0x1407ac7e0, 256B/52 行]
// 实证链：
//
//	入口 rax/rbx=target string, rcx/rdi=configPath string, stack=DesktopWidgetDocument
//	  → validateLauncherConfigExportTarget(target, configPath)
//	  → strings.TrimSpace(target)
//	  → 复制 widgetDoc 到栈参数区
//	  → writeSelfContainedLauncherConfigWithWidgets(target, configPath, widgetDoc)
//	  → 透传其 (string, error) 返回
//
// 旧骨架误将前 4 参写为 (cfg, widgets, target, configPath) —— 无 LauncherConfig 形参，
// rax/rbx 与 rcx/rdi 均为 string 对。
func exportSelfContainedLauncherConfigWithWidgets(
	target string,
	configPath string,
	widgetDoc DesktopWidgetDocument,
) (string, error) {
	if err := validateLauncherConfigExportTarget(target, configPath); err != nil {
		return "", err
	}
	target = strings.TrimSpace(target)
	return writeSelfContainedLauncherConfigWithWidgets(target, configPath, widgetDoc)
}

// writeSelfContainedLauncherConfigWithWidgets 执行自包含配置的写入。
// [S-sig] 骨架。VA 0x1407ac100，被 exportSelfContainedLauncherConfigWithWidgets 尾调。
// 实证入参：rax/rbx=target string（已 TrimSpace），rcx/rdi=configPath string，
// stack=DesktopWidgetDocument struct。返回 (string, error) 由 exportSelfContained 直接透传。
// 待后续实证还原写入逻辑。
func writeSelfContainedLauncherConfigWithWidgets(
	target string,
	configPath string,
	widgetDoc DesktopWidgetDocument,
) (string, error) {
	_ = target
	_ = configPath
	_ = widgetDoc
	return "", nil
}

// readLauncherConfigFileWithDesktopWidgets 从字节数据解析配置与小部件文档。
// [S-sig] 骨架。VA 0x1407ab260。调用点 ImportConfig.asm 0x14077bdec-0x14077be00：
//
//	rax = data.ptr, rbx = data.len          → 第 1 形参 string
//	lea rcx,[rip+0x4e8fbf] → VA 0x140C64DB8 → 第 2 形参（.rdata 常量串起始）
//	edi = 0x18 (24)                          → 第 2 形参长度为 24 字节
//
// 故第 2 形参是 24 字节的 .rdata 常量串；返回 error 由 0x14077be51 的 test rbx,rbx 判定。
// [P] 该 24 字节常量尚未解码（需 decode_rodata_str.py 跑 0x140C64DB8），
// 暂以空串占位，待验证阶段解码后回填。
func readLauncherConfigFileWithDesktopWidgets(data string, sourceTag string) (LauncherConfig, DesktopWidgetDocument, error) {
	_ = data
	_ = sourceTag
	return LauncherConfig{}, DesktopWidgetDocument{}, nil
}

// tagCatalogNames 取标签目录的名称列表。
// [S-sig] 骨架。VA 0x14087c920。调用点 CompleteInitialization.asm 0x140778a47，
// 入参 (rax,rbx,rcx) 为 defaultTagCatalogFromLanguageMessages 的 3 word 返回值。
func tagCatalogNames(catalog []TagCatalogItem) []string {
	if len(catalog) == 0 {
		return nil
	}
	names := make([]string, 0, len(catalog))
	for _, item := range catalog {
		names = append(names, item.Name)
	}
	return names
}
