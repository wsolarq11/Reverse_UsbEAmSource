package main

import "strings"

// sanitizeScreenshotPreviewResult 清洗截图预览结果：TrimSpace 三个 URL/数据字段与 path，
// Width/Height 钳制到 ≥0，Mode 经 normalizeScreenshotMode 归一，Cancelled 透传。
// [S 汇编 0x14099b3c0, 448B]：逐字段 TrimSpace / cmovl 钳制 / normalizeScreenshotMode。
func sanitizeScreenshotPreviewResult(r ScreenshotCaptureResult) ScreenshotCaptureResult {
	r.ImageData = strings.TrimSpace(r.ImageData)
	r.ImageURL = strings.TrimSpace(r.ImageURL)
	r.ThumbnailURL = strings.TrimSpace(r.ThumbnailURL)
	if r.Width < 0 {
		r.Width = 0
	}
	if r.Height < 0 {
		r.Height = 0
	}
	r.Mode = normalizeScreenshotMode(r.Mode)
	r.Path = strings.TrimSpace(r.Path)
	return r
}

// screenshotPreviewShowWindowFlags 返回截图预览窗口显示 flags。
// [S] ASM 0x14099e7a0: mov eax, 0x53; ret（返回常量 0x53=83）。
func screenshotPreviewShowWindowFlags() int {
	return 0x53
}

// screenshotPreviewShowWindowCommand 返回截图预览窗口 ShowWindow 命令。
// [S] ASM 0x14099e7c0: mov eax, 4; ret（SW_SHOWNOACTIVATE）。
func screenshotPreviewShowWindowCommand() int {
	return 4
}

// waitBeforeRevealNative 原生窗口显示前等待（定时器/条件）。
// [S-sig 0x140999de0, 256B]：全局计数 <=0 → shouldDisplayNative；否则 time.NewTimer + chanrecv1
// 后 shouldDisplayNative。体待预览窗口域专项还原。
func (s *screenshotPreviewWindowService) waitBeforeRevealNative() bool {
	return false
}

// shouldDisplayNative 判定是否应以原生窗口显示预览。
// [S-sig 0x1409998e0, 288B]：lock → 读字段(+0xa8/+0xb0/+0x28/+0xb1) 判定 → unlock。
// 体待预览窗口域专项还原。
func (s *screenshotPreviewWindowService) shouldDisplayNative(a, b interface{}) bool {
	_, _ = a, b
	return false
}

// screenshotPreviewBounds 计算截图预览窗口边界（基于屏幕尺寸钳位）。
// [S-sig 0x14099b660, 288B]：ScreenManager.GetAll → 宽高减 x/y 再减 0x12 → clamp。
// 体待预览窗口域专项还原。
func screenshotPreviewBounds(a interface{}, x, y int64) (int64, int64, int64, int64) {
	_, _, _ = a, x, y
	return 0, 0, 0, 0
}

// buildScreenshotPreviewWindowHTMLWithStateURL 构建预览窗口 HTML（JSON 状态 URL 注入）。
// [S-sig 0x14099b860, 256B]：TrimSpace → json.Marshal → 模板 Replace(json) → Replace("50")。
// 体待 HTML 模板常量专项还原。
func buildScreenshotPreviewWindowHTMLWithStateURL(a interface{}) string {
	_ = a
	return ""
}

// waitBeforeReveal 延迟后判定是否显示（全局 delay<=0 直接 shouldDisplay）。
// [S-sig 0x140999c60, 288B]：delay>0 则 NewTimer(delay) 阻塞后 shouldDisplay。
// 体待全局 delay 与 shouldDisplay 专项还原。
func (s *screenshotPreviewWindowService) waitBeforeReveal(a, b interface{}) bool {
	_, _ = a, b
	return false
}

// readThumbnailPNG 读取缩略图 PNG（持锁，asset.ReadBytes）。
// [S-sig 0x1409963c0, 352B]：lock(+0x8) → asset(+0xe0) 空→error →
// launcherAssetService.ReadBytes。体待 asset 路径常量专项还原。
func (s *screenshotPreviewWindowService) readThumbnailPNG() ([]byte, error) {
	_ = s
	return nil, nil
}

// shouldDisplay 判定是否应显示预览（持锁，字段匹配判定）。
// [S-sig 0x140999640, 352B]：lock(+0x8) → 字段(+0xa8/+0xb0/+0x18/+0x20) 匹配 →
// 返回 flag(+0xb1)==0。体待字段域专项还原。
func (s *screenshotPreviewWindowService) shouldDisplay(a, b, c interface{}) bool {
	_, _, _, _ = s, a, b, c
	return false
}

// waitUntilReady 等待预览就绪（timeout → selectgo → shouldDisplay）。
// [S-sig 0x140999a60, 416B]：timeout nil→false；time.NewTimer → selectgo →
// 超时→false；ready→shouldDisplay。体待就绪域专项还原。
func (s *screenshotPreviewWindowService) waitUntilReady(a, b, c, d interface{}) bool {
	_, _, _, _, _ = s, a, b, c, d
	return false
}

// screenshotPreviewUpdateSessionScript 更新预览会话脚本（Marshal 参数 + Sprintf 拼接）。
// [S-sig 0x14099b200, 448B]：Marshal(参数) → TrimSpace → Marshal → Sprintf。体待会话脚本域专项还原。
func screenshotPreviewUpdateSessionScript(a, b, c, d, e, f, g, h, i interface{}) string {
	_, _, _, _, _, _, _, _, _ = a, b, c, d, e, f, g, h, i
	return ""
}

// screenshotNativePreviewWindowProc 原生预览窗口消息处理（HashTrieMap 查找 → handleMessage/DefWindowProc）。
// [S-sig 0x14099c2c0, 448B]：HashTrieMap.Load 命中→handleMessage；否则 LazyProc.Call(DefWindowProc)。
// 体待消息分发域专项还原。
func screenshotNativePreviewWindowProc(a interface{}, msg uint32, w, l uintptr) uintptr {
	_, _, _, _ = a, msg, w, l
	return 0
}
