// AUTO-RECONSTRUCTED STUBS — Screenshot domain missing package-level helpers
// 研究用途
package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// beginWorkspaceDataOperation 开启一次工作区数据操作。
// [S 汇编 0x1409ef7c0, 128B] 实证：
//
//	bs==nil → runtime.newobject 分配 24B 错误 "工作区维护不可用"（@0x140C65700）并返 (nil, error)；
//	否则 `(&bs.workspaceDataMaintenance).begin()`（字段 @+0x520，`add rax, 0x520`）原样透传 (gate, error)。
func (bs *BootstrapService) beginWorkspaceDataOperation() (*workspaceDataMaintenanceGate, error) {
	if bs == nil {
		return nil, errors.New("工作区维护不可用")
	}
	return bs.workspaceDataMaintenance.begin()
}

// begin 进入一次操作：nil 守卫 → lock → active++ → unlock。
// [S 汇编 0x1409ef000] 返回 receiver 自身以便 defer 关闭。
func (g *workspaceDataMaintenanceGate) begin() (*workspaceDataMaintenanceGate, error) {
	if g == nil {
		return nil, errors.New("工作区维护不可用")
	}
	g.mu.Lock()
	g.active++
	g.mu.Unlock()
	return g, nil
}

// End 结束一次操作（defer 关闭）：引用计数归零时唤醒等待者。
// [S-sig]：begin() [S 0x1409ef000] 的逆操作，签名经 beginWorkspaceDataOperation(0x1409ef7c0) 实证。
func (g *workspaceDataMaintenanceGate) End() {
	if g == nil {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.active--
	if g.active == 0 && g.idle != nil {
		close(g.idle)
		g.idle = nil
	}
}

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
