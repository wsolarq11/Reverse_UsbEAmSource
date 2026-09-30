// AUTO-RECONSTRUCTED STUBS — Screenshot domain missing package-level helpers
// 研究用途
//
// 注（batch 270）：beginWorkspaceDataOperation / workspaceDataMaintenanceGate.begin/enter/End
// 已迁至 workspacemigration.go 并按汇编订正签名（begin→(func(), error)，
// beginWorkspaceDataOperation→(func(), error) 透明透传）。本文件保留截图域其余辅助函数。
package main

import (
	"os"
	"path/filepath"
	"strings"
)

// attachScreenshotCaptureAssetURL 为截图路径注册资产 URL。
// [S 汇编 0x140797ba0] 实证：screenshotAssetService → attachScreenshotThumbnailAssetURL →
// TrimSpace 路径 → screenshotContentTypeForPath → launcherAssetService.RegisterFile。
func (bs *BootstrapService) attachScreenshotCaptureAssetURL(path string) (string, error) {
	svc := bs.screenshotAssetService()
	if svc == nil {
		return path, nil
	}

	// 先注册缩略图
	bs.attachScreenshotThumbnailAssetURL(path)

	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return path, nil
	}

	ref, err := svc.RegisterFile("screenshot", filepath.Base(trimmed), trimmed, 0)
	if err != nil {
		return "", err
	}
	return ref.URL, nil
}

// resolveGeneratedScreenshotSaveFormat resolves the final save format. [S-sig] VA 0x140967a20
func resolveGeneratedScreenshotSaveFormat(png []byte, source, format string, ws WorkspaceLayout) string {
	_ = png
	_ = source
	_ = ws
	return normalizeScreenshotSaveFormat(format)
}

// saveScreenshotImageToPath saves PNG to an explicit path.
// [S 汇编 0x14096e6c0] os.WriteFile(path, png, 0644) → err guard → return path
func saveScreenshotImageToPath(png []byte, path, format string) (string, error) {
	_ = format
	if err := os.WriteFile(path, png, 0644); err != nil {
		return "", err
	}
	return path, nil
}
