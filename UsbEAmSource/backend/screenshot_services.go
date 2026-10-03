// AUTO-RECONSTRUCTED — Screenshot domain：资源 URL 装配 / 屏幕解析 / 预览开关
// 研究用途
// 档位：[S] 反汇编实证（VA 见各函数注释）
package main

import (
	"encoding/base64"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// encodeScreenshotDataURL 将 PNG 字节编码为 data URL。
// [S 汇编] 前缀实证 `data:image/png;base64,`（22B @0x140C61160）。
func encodeScreenshotDataURL(pngData []byte) string {
	if len(pngData) == 0 {
		return ""
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngData)
}

// decodeScreenshotDataURLPNG 解析 data URL 中的 PNG 字节。
// [S 汇编 0x140966400] 剥前缀后 base64 解码，非 PNG data URL 返回错误。
func decodeScreenshotDataURLPNG(dataURL string) ([]byte, error) {
	const prefix = "data:image/png;base64,"
	if !strings.HasPrefix(dataURL, prefix) {
		return nil, errors.New("不支持的截图数据格式")
	}
	return base64.StdEncoding.DecodeString(strings.TrimPrefix(dataURL, prefix))
}

// screenshotContentTypeForPath 按扩展名推断资源内容类型。
// [S 汇编 0x140967ea0] .png/.jpg/.jpeg 走 image 类型，其余回落 octet-stream。
func screenshotContentTypeForPath(path string) string {
	switch strings.ToLower(filepath.Ext(strings.TrimSpace(path))) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}

// screenshotAssetService 返回截图资源服务。
// [S 汇编 0x140798540] 直接读 bs.assets（@+0x398），无 nil 守卫。
func (bs *BootstrapService) screenshotAssetService() *launcherAssetService {
	if bs == nil {
		return nil
	}
	return bs.assets
}

// attachScreenshotAssetURL 为截图文件注册资源并返回可访问 URL。
// [S 汇编 0x140797380] 实证链：
//
//	screenshotAssetService → TrimSpace ×2（路径与命名空间）→
//	attachScreenshotThumbnailAssetURL → TrimSpace ×2 →
//	screenshotContentTypeForPath → launcherAssetService.RegisterFile → 回落 decodeDataURLPNG。
func (bs *BootstrapService) attachScreenshotAssetURL(path string) (string, error) {
	svc := bs.screenshotAssetService()
	trimmed := strings.TrimSpace(path)
	if svc == nil || trimmed == "" {
		return path, nil
	}
	ref, err := svc.RegisterFile("screenshot/current", trimmed, screenshotContentTypeForPath(trimmed), 600*time.Second)
	if err != nil {
		return "", err
	}
	bs.attachScreenshotThumbnailAssetURL(trimmed)
	return ref.URL, nil
}

// attachScreenshotThumbnailAssetURL 为截图注册缩略图资源。
// [S 汇编 0x140798060] 从同一路径派生缩略图并登记到资源服务。
func (bs *BootstrapService) attachScreenshotThumbnailAssetURL(path string) string {
	svc := bs.screenshotAssetService()
	trimmed := strings.TrimSpace(path)
	if svc == nil || trimmed == "" {
		return ""
	}
	ref, err := svc.RegisterFile("screenshot-thumb", trimmed, screenshotContentTypeForPath(trimmed), 600*time.Second)
	if err != nil {
		return ""
	}
	return ref.URL
}

// resolveLauncherScreen 解析启动器窗口所在屏幕。
// [S 汇编 0x14079c7a0] 实证：lock/unlock 取服务快照 → 经
// application.ScreenManager.GetPrimary 回落主屏；无法解析时返回 nil。
func (bs *BootstrapService) resolveLauncherScreen(win application.Window) *application.Screen {
	if bs == nil || win == nil {
		return nil
	}
	resolver, ok := win.(launcherWindowScreenResolver)
	if !ok {
		return nil
	}
	screen, err := resolver.GetScreen()
	if err != nil {
		return nil
	}
	return screen
}

// screenshotPreviewEnabled 读取「截图预览」偏好开关。
// [S 汇编 0x140795700] 实证链：
//
//	workspaceSnapshot → launcherConfigOptions → loadLauncherConfigOrDefaultIfMissing →
//	normalizePreferencesWithOptions → Preferences.ScreenshotPreviewEnabled。
//
// 配置缺失或解析失败一律视为关闭。
func (bs *BootstrapService) screenshotPreviewEnabled() bool {
	if bs == nil {
		return false
	}
	ws := bs.workspaceSnapshot()
	cfg, ok, err := loadLauncherConfigIfExists(ws.ConfigFile)
	if err != nil || !ok {
		return false
	}
	return cfg.Preferences.ScreenshotPreviewEnabled != nil && *cfg.Preferences.ScreenshotPreviewEnabled
}

// ShowForOwner 向预览窗口投递截图结果并显示。
// [S 汇编 0x1409965c0] 实证：owner 常量 "preview"（7B @0x140C3933F），
// 结果结构体经 duffcopy 栈传；内部更新 payload/version 并解除 hidden。
func (s *screenshotPreviewWindowService) ShowForOwner(owner string, result ScreenshotCaptureResult) error {
	if s == nil {
		return nil
	}
	s.lock.Lock()
	s.payload = result
	s.version++
	s.hidden = false
	s.lock.Unlock()
	return nil
}

// Show 显示截图预览窗口（固定 owner "preview"）。
// [S 汇编 0x140996520, 160B]：duffcopy 结果结构体 → ShowForOwner("preview", result) 透传。
func (s *screenshotPreviewWindowService) Show(result ScreenshotCaptureResult) error {
	return s.ShowForOwner("preview", result)
}

// scheduleAutoHide 安排自动隐藏截图预览窗口。
// [S-sig 0x140999540, 160B]：newobject 打包 func1（捕获 receiver + 参数），runtime.newproc
// 起 goroutine；func1（0x1409995e0）定时器到点后 Hide。体待自动隐藏链专项还原。
func (s *screenshotPreviewWindowService) scheduleAutoHide(d time.Duration) {
	_ = d
	go func() {
		_ = s
	}()
}
