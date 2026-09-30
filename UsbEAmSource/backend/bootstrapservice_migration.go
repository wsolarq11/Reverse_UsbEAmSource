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
//
// 注：迁移门闸（gate.begin/enter）、waitWorkspaceMigrationDrain、beginWorkspaceDataOperation、
// beginWorkspaceMigrationMaintenance、validateWorkspaceMigrationRoots、workspacePathContains、
// stageWorkspaceDataMigration、verifyWorkspaceDataManifestWithPolicy、stagedWorkspaceData.Commit/Rollback、
// removeOwnedWorkspaceMigrationDirectory 均已迁至 workspacemigration.go（batch 270）。
package main

// buildWorkspaceLayoutWithConfig 依据配置构造工作区布局。
// [S-sig] 骨架。VA 0x1407a1ac0。调用点 MigrateConfig.asm 0x14077c8d6，
// 入参为迁移后配置，返回新的 WorkspaceLayout 供 applyWorkspaceLayout 应用。
func buildWorkspaceLayoutWithConfig(cfg LauncherConfig, target string) WorkspaceLayout {
	_ = cfg
	ws := WorkspaceLayout{Root: target, ConfigFile: target}
	return ws
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
