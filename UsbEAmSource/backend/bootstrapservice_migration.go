// AUTO-RECONSTRUCTED — DOMAIN: bootstrap workspace data migration（工作区数据迁移域）
// 研究用途
//
// 契约来源：capstone 反汇编符号标注（docs/goresym/pipeline/tmp/*.asm.txt）
//   - MigrateConfig 0x14077c5e0 / 3296B（va_map_fixed2.txt:112）
//   - buildWorkspaceLayoutWithConfig 0x1407a1ac0 / 1696B（buildWorkspaceLayoutWithConfig.asm.txt）
//   - importLauncherBackgroundImage 0x140798e40 / 1632B（importLauncherBackgroundImage.asm.txt）
//   - copyLauncherBackgroundImageFile 0x1407994a0 / 672B（copyLauncherBackgroundImageFile.asm.txt）
//   - normalizeLauncherBackgroundImageExtension 0x1407997a0 / 256B（normalizeLauncherBackgroundImageExtension.asm.txt）
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

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

// buildWorkspaceLayoutWithConfig 依据配置构造工作区布局。
// [S 汇编实证 0x1407a1ac0, 1696B]（buildWorkspaceLayoutWithConfig.asm.txt）：
//
//	参序（8×string，前 3 组走寄存器 RAX-R8，后 5 组走栈）：
//	  root=参1 configFile=参2 pluginDir=参3 dataRoot/iconDir/indexDir/screenshotDir/webView2Dir=参4-8。
//	rootEff = TrimSpace(root)，空则 "."；normalizeStorageConfig 后 DataRoot 非空则覆盖 rootEff。
//	ConfigFile = TrimSpace(configFile)，空则 resolveLauncherConfigFilePath(resolveProcessWorkingDirectory())。
//	LanguageDir/AppLanguageDir = filepath.Join(filepath.Dir(pluginDir), "language")；PluginDir 原样直传。
//	IconDir/IndexDir/ScreenshotDir/WebView2Dir 各自非空优先，空则 filepath.Join(rootEff, 默认子目录)。
//	BackgroundDir = filepath.Join(rootEff, "background")。
//	默认子目录名（rodata 解码，单数形式）：icon/index/screenshot/webview2/background/language。
//
// 三处调用点交叉验证（寄存器映射一致）：
//	commit...WithWidgetsImpl.asm 0x140778f40：参1-3 = ws.Root/ws.ConfigFile/ws.PluginDir；
//	MigrateConfig.asm 0x14077c8d6：参1 = staged.targetPath，参2-3 = ws.ConfigFile/ws.PluginDir；
//	PreviewInitializationImportConfig.asm 0x14077c1e2：先 filepathlite.Dir 后装配参2-3。
func buildWorkspaceLayoutWithConfig(root, configFile, pluginDir, dataRoot, iconDir, indexDir, screenshotDir, webView2Dir string) WorkspaceLayout {
	rootEff := strings.TrimSpace(root)
	if rootEff == "" {
		rootEff = "."
	}
	sc := normalizeStorageConfig(dataRoot, iconDir, indexDir, screenshotDir, webView2Dir)
	if sc.DataRoot != "" {
		rootEff = sc.DataRoot
	}

	configPath := strings.TrimSpace(configFile)
	if configPath == "" {
		configPath = resolveLauncherConfigFilePath(resolveProcessWorkingDirectory())
	}

	icon := sc.IconDir
	if icon == "" {
		icon = filepath.Join(rootEff, "icon")
	}
	idx := sc.IndexDir
	if idx == "" {
		idx = filepath.Join(rootEff, "index")
	}
	shot := sc.ScreenshotDir
	if shot == "" {
		shot = filepath.Join(rootEff, "screenshot")
	}
	wv2 := sc.WebView2Dir
	if wv2 == "" {
		wv2 = filepath.Join(rootEff, "webview2")
	}
	language := filepath.Join(filepath.Dir(pluginDir), "language")

	return WorkspaceLayout{
		Root:           rootEff,
		ConfigFile:     configPath,
		LanguageDir:    language,
		AppLanguageDir: language,
		PluginDir:      pluginDir,
		IconDir:        icon,
		IndexDir:       idx,
		ScreenshotDir:  shot,
		WebView2Dir:    wv2,
		BackgroundDir:  filepath.Join(rootEff, "background"),
	}
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
// [S 汇编实证 0x140798e40, 1632B]（importLauncherBackgroundImage.asm.txt +
// ChooseLauncherBackgroundImage.asm 0x14077b3c9 交叉验证）：
//
//	签名 (sourcePath string) (string, error)，无 ws 参数——ws 由 bs.workspaceSnapshot() 取 BackgroundDir。
//	流程：TrimSpace(sourcePath) → filepath.Abs → os.Stat → FileInfo.IsDir 判目录；
//	ext = normalizeLauncherBackgroundImageExtension(filepath.Ext(abs))（空则"不支持的背景图片格式"）；
//	os.MkdirAll(BackgroundDir, 0o755) → dest = filepath.Join(BackgroundDir, "custom-background"+ext)；
//	copyLauncherBackgroundImageFile(abs, dest) → 装配 BackgroundPreference（Enabled=true、
//	ImagePath=dest、ReadabilityOverlayEnabled=true、ReadabilityOverlayOpacity=0.18、ImageOpacity=1.0）
//	经 normalizeBackgroundPreference 归一化后 attachLauncherBackgroundURL；成功返回 (filepath.Base(abs), nil)。
//	错误消息（rodata 解码）："背景图片路径不能为空" / "背景图片不能是目录" / "不支持的背景图片格式"。
func (bs *BootstrapService) importLauncherBackgroundImage(sourcePath string) (string, error) {
	p := strings.TrimSpace(sourcePath)
	if p == "" {
		return "", newExtractError("背景图片路径不能为空")
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", newExtractError("背景图片不能是目录")
	}

	ext := normalizeLauncherBackgroundImageExtension(filepath.Ext(abs))
	if ext == "" {
		return "", newExtractError("不支持的背景图片格式")
	}

	ws := bs.workspaceSnapshot()
	if err := os.MkdirAll(ws.BackgroundDir, 0o755); err != nil {
		return "", err
	}

	dest := filepath.Join(ws.BackgroundDir, "custom-background"+ext)
	if err := copyLauncherBackgroundImageFile(abs, dest); err != nil {
		return "", err
	}

	trueVal := true
	readabilityOpacity := 0.18
	imageOpacity := 1.0
	pref := BackgroundPreference{
		Enabled:                   true,
		ImagePath:                 dest,
		ReadabilityOverlayEnabled: &trueVal,
		ReadabilityOverlayOpacity: &readabilityOpacity,
		ImageOpacity:              &imageOpacity,
	}
	cfg := LauncherConfig{Preferences: Preferences{Background: normalizeBackgroundPreference(pref)}}
	bs.attachLauncherBackgroundURL(&cfg)

	return filepath.Base(abs), nil
}

// normalizeLauncherBackgroundImageExtension 归一化背景图扩展名。
// [S 汇编实证 0x1407997a0, 256B]（normalizeLauncherBackgroundImageExtension.asm.txt）：
// TrimSpace + ToLower 后按字面量分派：.bmp/.gif/.jpg/.png/.webp 原样，.jpeg 归一为 .jpg，其余空串。
// 立即数实证：0x706d622e=".bmp" 0x6669672e=".gif" 0x67706a2e=".jpg" 0x676e702e=".png"
// 0x65706a2e(+'g')=".jpeg" 0x6265772e(+'p')=".webp"。
func normalizeLauncherBackgroundImageExtension(ext string) string {
	switch strings.ToLower(strings.TrimSpace(ext)) {
	case ".bmp":
		return ".bmp"
	case ".gif":
		return ".gif"
	case ".jpg", ".jpeg":
		return ".jpg"
	case ".png":
		return ".png"
	case ".webp":
		return ".webp"
	}
	return ""
}

// copyLauncherBackgroundImageFile 复制背景图文件（源→目标）。
// [S 汇编实证 0x1407994a0, 672B]（copyLauncherBackgroundImageFile.asm.txt）：
// Clean(src)==Clean(dst) 短路返回 nil；os.OpenFile(src, O_RDONLY, 0) 后 defer Close；
// os.OpenFile(dst, O_WRONLY|O_CREATE|O_TRUNC, 0o644)；io.Copy(dst, src) 出错关闭 dst 返错；
// 成功返回 dst.Close() 的错误。
func copyLauncherBackgroundImageFile(src, dst string) error {
	if filepath.Clean(src) == filepath.Clean(dst) {
		return nil
	}
	srcFile, err := os.OpenFile(src, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		return err
	}
	return dstFile.Close()
}
