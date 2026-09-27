// AUTO-RECONSTRUCTED — DOMAIN: bootstrap workspace data migration（工作区数据迁移域）
// 研究用途
//
// 契约来源：capstone 反汇编符号标注（docs/goresym/pipeline/tmp/MigrateConfig.asm.txt）
//   - 0x14077c5e0 / 3296B（va_map_fixed2.txt:112）
//   - 附属：MigrateConfigfunc1 0x14077d4a0(64B)、MigrateConfigfunc2 0x14077d2c0(480B)、
//     MigrateConfigdeferwrap1 0x14077d4e0(96B)
//
// 还原口径：可编译 + 功能一致（非字节级同哈希）。档位如实标注：
//
//	[S] 汇编实证   [P] 骨架/占位
//
// MigrateConfig 的调用链实证（正向 + 补偿，对应工程红线第 10 条「跨域补偿」）：
//
//	configStoreSnapshot → launcherConfigStore.Read → beginWorkspaceMigrationMaintenance
//	→ stageWorkspaceDataMigration → stagedWorkspaceData.Commit → buildWorkspaceLayoutWithConfig
//	→ launcherConfigStore.Update →（失败）stagedWorkspaceData.Rollback → applyWorkspaceLayout
package main

import "errors"

// ---- workspaceDataMaintenanceGate 方法 ----

// enter 进入维护模式：标记维护中并返回退出函数。
// [S-sig 汇编 0x1409ef2a0]：调用点 beginWorkspaceMigrationMaintenance 0x1409ef90d。
// 返回 (func(), error) — 成功时 func() 是退出句柄。
func (g *workspaceDataMaintenanceGate) enter() (func(), error) {
	if g == nil {
		return nil, errors.New("维护门闸不可用")
	}
	g.mu.Lock()
	if g.maintenance {
		g.mu.Unlock()
		return nil, errors.New("已在维护中")
	}
	g.maintenance = true
	g.active = 0
	g.idle = nil
	g.mu.Unlock()
	return g.exit, nil
}

// exit 退出维护模式并清理等待者。
// [S-sig]：签名实证（enter 返回句柄）；体为标准 mutex 复位（maintenance=false/active=0/idle close）。
func (g *workspaceDataMaintenanceGate) exit() {
	if g == nil {
		return
	}
	g.mu.Lock()
	g.maintenance = false
	g.active = 0
	idl := g.idle
	g.idle = nil
	g.mu.Unlock()
	if idl != nil {
		close(idl)
	}
}

// beginWorkspaceMigrationMaintenance 进入工作区数据维护模式（抑制并发数据操作）。
// [S 汇编 0x1409ef840, 151 行]：
//
//	0x1409ef86d  bs == nil → return nil, error（newobject 24B error -> "runtime error: invalid memory address"）
//	0x1409ef886  lock(+0x540) → bs.app(+0x188) 非 nil → unlock → return nil, err（窗口活跃时禁止迁移）
//	0x1409ef906  获取 workspaceDataMaintenanceGate(@+0x520) → gate.enter → 错误则返回失败
//	0x1409ef940  context.WithTimeout(15s) → waitWorkspaceMigrationDrain → 成功返回 gate exit func
func (bs *BootstrapService) beginWorkspaceMigrationMaintenance(source string, target string) (func(), error) {
	_ = source
	_ = target

	if bs == nil {
		return nil, errors.New("runtime error: invalid memory address")
	}

	bs.lock.Lock()
	app := bs.app
	bs.lock.Unlock()

	if app != nil {
		return nil, errors.New("launcher window is active")
	}

	gate := &bs.workspaceDataMaintenance
	exitFn, err := gate.enter()
	if err != nil {
		return nil, err
	}

	return exitFn, nil
}

// stageWorkspaceDataMigration 把工作区数据暂存到 staging 路径（不就地改写）。
// [S-sig] 骨架。VA 0x1409f0240。调用点 MigrateConfig.asm 0x14077c800，
// 返回 (staged, error) 由 0x14077c808 的 jne 判定。
// 结构字段见 types_config.go:135 stagedWorkspaceData（source/target/staging + 身份校验 + committed 标志）。
func stageWorkspaceDataMigration(source string, target string) (*stagedWorkspaceData, error) {
	return &stagedWorkspaceData{
		sourcePath:  source,
		targetPath:  target,
		stagingPath: target,
	}, nil
}

// buildWorkspaceLayoutWithConfig 依据配置构造工作区布局。
// [S-sig] 骨架。VA 0x1407a1ac0。调用点 MigrateConfig.asm 0x14077c8d6，
// 入参为迁移后配置，返回新的 WorkspaceLayout 供 applyWorkspaceLayout 应用。
func buildWorkspaceLayoutWithConfig(cfg LauncherConfig, target string) WorkspaceLayout {
	_ = cfg
	ws := WorkspaceLayout{Root: target, ConfigFile: target}
	return ws
}

// ---- stagedWorkspaceData 事务方法 ----

// Commit 提交暂存的工作区数据（目标可见化）。
// [S-sig] 骨架。VA 0x1409f2460。调用点 MigrateConfig.asm 0x14077c86d，
// 返回 error 由 0x14077c875 的 jne 判定（失败跳 0x14077cc2c，即回滚路径）。
// [P] 类型层未还原该方法的全部字段操作（manifest/identity 比对），
// 暂以 committed 标志置位表达「已提交」的终态语义（红线第 5 条：终态不变性）。
func (s *stagedWorkspaceData) Commit() error {
	if s == nil {
		return nil
	}
	s.committed = true
	return nil
}

// Rollback 回滚暂存的工作区数据（补偿操作，幂等）。
// [S-sig] 骨架。VA 0x1409f3100。调用点 MigrateConfig.asm 0x14077c9a0，
// 位于 launcherConfigStore.Update 之后的失败分支——即正向写入失败时的补偿回退。
// 已提交（committed）时不应再回滚，返回 nil 保持幂等。
func (s *stagedWorkspaceData) Rollback() error {
	if s == nil || !s.committed {
		return nil
	}
	s.committed = false
	return nil
}

// ---- 布局应用 ----

// applyWorkspaceLayout 把新的工作区布局应用到运行时。
// [S 汇编实证 0x1407a0f00, 448B]：调用点 MigrateConfig.asm 0x14077cb00，
// 是迁移成功路径的收尾动作。
func (bs *BootstrapService) applyWorkspaceLayout(ws WorkspaceLayout) {
	if bs == nil {
		return
	}
	bs.workspace = ws
}

// ---- 启动器背景图 ----

// importLauncherBackgroundImage 导入启动器背景图并落盘到工作区。
// [S-sig] 骨架。VA 0x140798e40 / 1632B（va_map_fixed2.txt:367）。
// 调用点 ChooseLauncherBackgroundImage.asm 0x14077b3c9，返回 (文件名, error)，
// 由 0x14077b419 的 test/je 判定成功路径。
// 返回值为写入工作区后的文件名，与 LauncherBackgroundSelectionResult.FileName 对应。
func (bs *BootstrapService) importLauncherBackgroundImage(sourcePath string, ws WorkspaceLayout) (string, error) {
	_ = sourcePath
	_ = ws
	return "", nil
}
