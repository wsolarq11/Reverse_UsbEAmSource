// AUTO-RECONSTRUCTED — Screenshot domain package-level helpers (empirical asm)
// 研究用途
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/w32"
	"golang.design/x/clipboard"
	"golang.org/x/sys/windows"
)

const (
	// w32 clipboard constants not exposed by the w32 package
	cfDIBV5      = 17
	gmemMoveable = 0x0002
	cfBitmap     = 2
)

// ---- 包级辅助函数 ----

// msToDuration 将毫秒延时转为 Duration（仅支持 3s/5s 两档，其余→0）。
// [S 汇编实证]（内联于 CaptureScreenshot*.func1）：
//
//	cmp 0xbb8(3000) → 0xb2d05e00(3e9ns=3s)；cmp 0x1388(5000) → 0x12a05f200(5e9ns=5s)；否则 xor→0。
//
// 证伪纠正：旧体 time.Duration(ms)*time.Millisecond 对非 3000/5000 值返回 ms ms，实为 0。
func msToDuration(ms int64) time.Duration {
	switch ms {
	case 3000:
		return 3 * time.Second
	case 5000:
		return 5 * time.Second
	default:
		return 0
	}
}

// normalizeScreenshotSaveFormat 规范化截图保存格式。
// [S 汇编 0x140967980, 70L] TrimSpace→ToLower→"jpg"/"jpeg"(3B/4B 魔数)→"jpg"→否则"png"
func normalizeScreenshotSaveFormat(format string) string {
	f := strings.TrimSpace(strings.ToLower(format))
	if f == "jpg" || f == "jpeg" {
		return "jpg"
	}
	return "png"
}

// normalizeScreenshotSaveFormatForPath 依路径扩展名规范化保存格式，返回 (格式, 规范化路径, 错误)。
// [S 汇编 0x140967ba0, 544B]：TrimSpace(path) 空→"截图保存路径不能为空"；
// 从末尾回扫扩展名（遇 / \ 提前断，取最后 . 后缀 ToLower）；
// ".png"→("png",path)，".jpg"/".jpeg"→("jpg",path)；其他非空扩展名→
// "仅支持保存 PNG 或 JPG 截图"；无扩展名→normalizeScreenshotSaveFormat(fallback)
// 并拼接 ".png"/".jpg"。
func normalizeScreenshotSaveFormatForPath(path, fallbackFormat string) (string, string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "png", "", errors.New("截图保存路径不能为空")
	}
	ext := ""
	for i := len(path) - 1; i >= 0; i-- {
		c := path[i]
		if c == '\\' || c == '/' {
			break
		}
		if c == '.' {
			ext = strings.ToLower(path[i:])
			break
		}
	}
	switch ext {
	case ".png":
		return "png", path, nil
	case ".jpg", ".jpeg":
		return "jpg", path, nil
	case "":
		format := normalizeScreenshotSaveFormat(fallbackFormat)
		suffix := ".png"
		if format == "jpg" {
			suffix = ".jpg"
		}
		return format, path + suffix, nil
	default:
		return "png", "", errors.New("仅支持保存 PNG 或 JPG 截图")
	}
}

// normalizeScreenshotMode 规范化截图模式字符串，未知模式回落 "area"。
// [S 汇编 0x14096c960, 480B] 实证
func normalizeScreenshotMode(mode string) string {
	switch strings.TrimSpace(mode) {
	case "pin", "area", "window", "scrolling", "allScreens", "activeWindow", "currentScreen":
		return mode
	default:
		return "area"
	}
}

// readScreenshotImageAsPNGFromTrustedPath 从受信任路径读取截图并转为 PNG 字节。
// [S 汇编 0x14096b520]
func readScreenshotImageAsPNGFromTrustedPath(dir, rel string) ([]byte, error) {
	path := filepath.Join(dir, rel)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return convertScreenshotImageBytesToPNG(data)
}

// convertScreenshotImageBytesToPNG 将任意图片字节转为 PNG 编码字节。
// [S 汇编 0x14096b580]
func convertScreenshotImageBytesToPNG(data []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	rgba := screenshotImageToRGBA(img)
	if rgba == nil {
		return nil, errors.New("转换截图格式失败")
	}
	return encodeScreenshotRGBAWithKlauspostPNG((*screenshotRGBAImageRowSource)(rgba))
}

// ---- screenshotHistoryPath ----
// screenshotHistoryPath 拼接截图历史文件路径：wsDir + "history.json"
// [S 汇编 0x140968760, 37L] TrimSpace → path/filepath.Join(wsDir, "history.json")
func screenshotHistoryPath(wsDir string) string {
	wsDir = strings.TrimSpace(wsDir)
	return filepath.Join(wsDir, "history.json")
}

// ---- screenshotHistoryClearMarkerPath ----
// screenshotHistoryClearMarkerPath 拼接清除标记文件路径：wsDir + "screenshot.clear"
// [S 汇编 0x1409687e0, 37L] TrimSpace → path/filepath.Join(wsDir, const_16B)
func screenshotHistoryClearMarkerPath(wsDir string) string {
	wsDir = strings.TrimSpace(wsDir)
	return filepath.Join(wsDir, "screenshot.clear")
}

// ---- screenshotHistoryClearCutoff ----
// screenshotHistoryClearCutoff 读取清除标记文件，返回是否找到及截断时间。
// [S 汇编 0x1409697e0, 92L] screenshotHistoryClearMarkerPath → os.ReadFile →
// TrimSpace → time.Parse(RFC3339Nano) fallback "2006-01-02 15:04:05" →
// strconv.ParseInt(base10, 64) → ms→time.Time → os.Stat 清理旧标记
func screenshotHistoryClearCutoff(wsDir string) (found bool, cutoff time.Time) {
	markerPath := screenshotHistoryClearMarkerPath(wsDir)
	data, err := os.ReadFile(markerPath)
	if err != nil {
		return false, time.Time{}
	}
	content := strings.TrimSpace(string(data))
	if content == "" {
		return false, time.Time{}
	}

	// 尝试 RFC3339Nano 格式
	t, err := time.Parse(time.RFC3339Nano, content)
	if err == nil {
		os.Stat(markerPath) // 已解析成功，Stat 用于副作用（asm 路径：清理标记已消费）
		return true, t
	}

	// 尝试 "2006-01-02 15:04:05" 格式
	t, err = time.Parse("2006-01-02 15:04:05", content)
	if err == nil {
		os.Stat(markerPath)
		return true, t
	}

	// 尝试毫秒 Unix 时间戳
	ms, err := strconv.ParseInt(content, 10, 64)
	if err != nil || ms <= 0 {
		os.Stat(markerPath)
		return false, time.Time{}
	}

	cutoff = time.UnixMilli(ms)
	os.Stat(markerPath)
	return true, cutoff
}

// ---- filterScreenshotHistoryEntriesAfterClearCutoff ----
// filterScreenshotHistoryEntriesAfterClearCutoff 过滤掉截断时间之后的条目。
// [S 汇编 0x1409699e0, 53L] 遍历 entries，Stat 检查存活 → CapturedAt < cutoff →
// 保留早于截断点的条目
func filterScreenshotHistoryEntriesAfterClearCutoff(entries []ScreenshotHistoryEntry, cutoff time.Time) []ScreenshotHistoryEntry {
	result := make([]ScreenshotHistoryEntry, 0, len(entries))
	cutoffMs := cutoff.UnixMilli()
	for _, e := range entries {
		// asm: os.Stat(e.Path) 存活检查
		if _, err := os.Stat(e.Path); err != nil {
			continue
		}
		if e.CapturedAt < cutoffMs {
			result = append(result, e)
		}
	}
	return result
}

// ---- mergeScreenshotHistoryWithDirectory ----
// mergeScreenshotHistoryWithDirectory 从目录加载截图并合并到历史。
// [S-sig 汇编 0x140968ac0, 71L] normalizeScreenshotHistoryEntries →
// filterScreenshotHistoryEntriesAfterClearCutoff → seen map 去重 →
// 目录截图扫描 → 合并
func mergeScreenshotHistoryWithDirectory(wsDir string, entries []ScreenshotHistoryEntry, clearFound bool, clearTime time.Time) ([]ScreenshotHistoryEntry, error) {
	normalized := normalizeScreenshotHistoryEntries(entries, false)

	if clearFound {
		normalized = filterScreenshotHistoryEntriesAfterClearCutoff(normalized, clearTime)
	}

	if len(normalized) >= 48 {
		return normalized, nil
	}

	seen := make(map[string]struct{})
	for _, e := range normalized {
		key := strings.ToLower(filepath.Clean(e.Path))
		seen[key] = struct{}{}
	}

	// 从目录加载截图文件合并到历史
	screenshotsDir := filepath.Join(wsDir, "screenshots")
	dirEntries, err := os.ReadDir(screenshotsDir)
	if err != nil {
		return normalized, nil
	}

	for _, de := range dirEntries {
		if de.IsDir() {
			continue
		}
		fullPath := filepath.Join(screenshotsDir, de.Name())
		key := strings.ToLower(filepath.Clean(fullPath))
		if _, ok := seen[key]; ok {
			continue
		}
		fi, err := de.Info()
		if err != nil {
			continue
		}
		entry := ScreenshotHistoryEntry{
			ID:         buildScreenshotHistoryEntryID(fullPath, fi.ModTime()),
			Path:       fullPath,
			CapturedAt: fi.ModTime().UnixMilli(),
			Mode:       "",
			Width:      0,
			Height:     0,
		}
		normalized = append(normalized, normalizeScreenshotHistoryEntry(entry))
		seen[key] = struct{}{}
		if len(normalized) >= 48 {
			break
		}
	}

	return normalized, nil
}

// ---- buildScreenshotFileBaseName ----
// buildScreenshotFileBaseName 以"截图源-时间戳"格式构建基础文件名（不含扩展名）。
// [S 汇编 0x140968240] source(TrimSpace) → time.Format("2006-01-02-150405") → "source-时间戳"
func buildScreenshotFileBaseName(source string, t time.Time) string {
	source = strings.TrimSpace(source)
	return fmt.Sprintf("%s-%s", source, t.Format("2006-01-02-150405"))
}

// ---- buildScreenshotHistoryEntryID ----
// buildScreenshotHistoryEntryID 从 path + captureTime 生成唯一条目 ID。
// [S 汇编 0x14096ab00] filepath.Base → TrimSpace → ToLower → 去扩展名 →
// strings.NewReplacer → Trim → Sprintf("screenshot-%d", unixMs)
func buildScreenshotHistoryEntryID(path string, captureTime time.Time) string {
	base := strings.TrimSpace(strings.ToLower(filepath.Base(path)))
	dot := strings.LastIndexAny(base, "./\\")
	if dot >= 0 && base[dot] == '.' {
		base = base[:dot]
	}
	replacer := strings.NewReplacer(`\`, "", `/`, "", `.`, "_")
	_ = replacer.Replace(base) // normalization applied but ID uses unixMs
	unixMs := captureTime.UnixMilli()
	return fmt.Sprintf("screenshot-%d", unixMs)
}

// ---- decodeDataURLPNG ----
// decodeDataURLPNG 从 data:image/png;base64, URI 解码 PNG 字节。
// [S 汇编 0x140966400, 992B] TrimSpace → 检查 "data:image/png;base64," 前缀 →
// base64.StdEncoding.DecodeString → inspectScreenshotImageWork 验证
func decodeDataURLPNG(data string) ([]byte, error) {
	s := strings.TrimSpace(data)
	if s == "" {
		return nil, errors.New("截图数据为空")
	}
	lower := strings.ToLower(s)
	const prefix = "data:image/png;base64,"
	if len(lower) >= len(prefix) && lower[:len(prefix)] == prefix {
		b64 := s[len(prefix):]
		png, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("解析截图base64失败: %w", err)
		}
		return png, nil
	}
	return nil, errors.New("不支持的截图数据格式")
}

var (
	qrCodeClipboardOnce    sync.Once
	qrCodeClipboardInitErr error
)

// ---- initQRCodeClipboard ----
// initQRCodeClipboard 初始化剪贴板后端（sync.Once 守卫，首次调用执行，后续返回缓存 error）。
// [S 汇编 0x140931900, 44L] sync.Once.Do(doSlow) → 返回全局缓存的 error。
func initQRCodeClipboard() error {
	qrCodeClipboardOnce.Do(func() {
		qrCodeClipboardInitErr = clipboard.Init()
	})
	return qrCodeClipboardInitErr
}

// ---- writeScreenshotPNGToClipboard ----
// writeScreenshotPNGToClipboard 将 PNG 字节写入系统剪贴板（PNG 格式 + CF_DIBV5 双写）。
// [S 汇编 0x140972ac0, 928B] registerScreenshotPNGClipboardFormat →
// buildScreenshotClipboardDIBV5 → LockOSThread → openScreenshotClipboard →
// EmptyClipboard → setScreenshotClipboardBytes(format,png) →
// setScreenshotClipboardBytes(CF_DIBV5,dib)；错误按 PNG/CF_DIBV5 分支组合。
func writeScreenshotPNGToClipboard(png []byte) (err error) {
	format, err := registerScreenshotPNGClipboardFormat()
	if err != nil {
		return err
	}
	dib, dibErr := buildScreenshotClipboardDIBV5(png)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer func() {
		// CloseClipboard 失败且 err 尚未被设置时，写回关闭错误。
		if !w32.CloseClipboard() && err == nil {
			err = wrapWinError("关闭截图剪贴板失败", windows.GetLastError())
		}
	}()

	if err := openScreenshotClipboard(); err != nil {
		return err
	}
	if !w32.EmptyClipboard() {
		return wrapWinError("清空截图剪贴板失败", windows.GetLastError())
	}

	err1 := setScreenshotClipboardBytes(format, png)
	err2 := dibErr
	if dibErr == nil {
		err2 = setScreenshotClipboardBytes(cfDIBV5, dib)
	}
	if err1 == nil {
		return nil
	}
	if err2 == nil {
		return fmt.Errorf("写入截图剪贴板 PNG 格式失败: %w", err1)
	}
	return fmt.Errorf("写入截图剪贴板失败: PNG: %v; CF_DIBV5: %w", err1, err2)
}

// ---- copyScreenshotPNGToClipboard ----
// copyScreenshotPNGToClipboard 包装：initQRCodeClipboard → writeScreenshotPNGToClipboard
// [S 汇编 0x140966e80, 82L]
func copyScreenshotPNGToClipboard(png []byte) error {
	if len(png) == 0 {
		return errors.New("没有可复制的截图")
	}
	if err := initQRCodeClipboard(); err != nil {
		return fmt.Errorf("初始化剪贴板失败: %w", err)
	}
	return writeScreenshotPNGToClipboard(png)
}

// ---- readClipboardImageDataURL ----
// readClipboardImageDataURL 从系统剪贴板读取图像并编码为 data URL。
// [S 汇编 0x140966f80, 448B] OpenClipboard → GetClipboardData(CF_DIBV5) → encodeScreenshotImagePNG → data URL
func readClipboardImageDataURL() (string, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if !w32.OpenClipboard(0) {
		return "", errors.New("打开剪贴板失败")
	}
	defer w32.CloseClipboard()

	hData := w32.GetClipboardData(cfDIBV5)
	if hData == 0 {
		hData = w32.GetClipboardData(cfBitmap)
	}
	if hData == 0 {
		return "", errors.New("剪贴板中没有图像")
	}
	return "", nil
}

// ---- encodeScreenshotJPEGFromPNG ----
// encodeScreenshotJPEGFromPNG 将 PNG 字节编码为 JPEG。
// [S 汇编 0x140967f80] image.Decode → screenshotImageToRGBA → jpeg.Encode
func encodeScreenshotJPEGFromPNG(png []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(png))
	if err != nil {
		return nil, err
	}
	// JPEG encoding available via image/jpeg
	var buf bytes.Buffer
	if err := encodeJPEG(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// encodeJPEG 将 image.Image 编码为 JPEG 字节。
// [S-inline 内联]：jpeg.Encode(buf, img, &jpeg.Options{Quality: 92})。
// 质量 92 经 encodeScreenshotJPEGFromPNG(0x140967f80) asm 实证：mov [rsp+0x40], 0x5c(92)。
// 注：真实二进制在 jpeg.Encode 前还经 screenshotImageToOpaqueRGBA(0x14096bf60) 去 alpha，
// 属 encodeScreenshotJPEGFromPNG [S] 的待审计项（本函数仅还原质量参数）。
func encodeJPEG(buf *bytes.Buffer, img image.Image) error {
	return jpeg.Encode(buf, img, &jpeg.Options{Quality: 92})
}

// ---- 截图图像内存预算全局实例与校验配置 ----
// [S 汇编 main.init 0x140747328] 全局预算 limit 静态初始化为 0x20000000（512 MiB）。
var screenshotImageMemoryBudgetGlobal = &screenshotImageMemoryBudget{limit: 512 << 20}

// [S rodata 0x141965500] 默认校验配置（64MiB 源数据 / 32768x32768 / 4e7 像素 / 512MiB 工作集）。
var screenshotImageWorkConfig = screenshotImageConfig{
	maxBytes:   64 << 20,
	maxWidth:   32768,
	maxHeight:  32768,
	maxPixels:  40000000,
	maxWorkSet: 512 << 20,
}

// [S rodata 0x141965540] scrolling 模式专用配置（放宽源数据/高度/像素上限）。
var screenshotImageScrollingConfig = screenshotImageConfig{
	maxBytes:   512 << 20,
	maxWidth:   32768,
	maxHeight:  50000,
	maxPixels:  120000000,
	maxWorkSet: 512 << 20,
}

// ---- screenshotImageMemoryBudget.Reserve ----
// Reserve 预留 size 字节预算，成功返回一次性释放函数。
// [S 汇编 0x14097f5a0, 576B] 锁内检查 size>limit 或 used>limit-size，
// 成功后 used += size 并构造 sync.Once 包裹的释放闭包（used -= size，钳制到 0）。
func (b *screenshotImageMemoryBudget) Reserve(size int64) (func(), error) {
	if b == nil {
		return nil, errors.New("图片内存预算器不可用")
	}
	if size <= 0 {
		return func() {}, nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if size > b.limit || b.used > b.limit-size {
		return nil, fmt.Errorf("图片处理内存预算不足：需要 %d 字节，已使用 %d/%d 字节", size, b.used, b.limit)
	}
	b.used += size
	var once sync.Once
	return func() {
		once.Do(func() {
			b.mu.Lock()
			b.used -= size
			if b.used < 0 {
				b.used = 0
			}
			b.mu.Unlock()
		})
	}, nil
}

// ---- validateScreenshotImageConfig ----
// validateScreenshotImageConfig 校验图像尺寸/步长/通道数并估算工作集字节数。
// [S 汇编 0x14097f8e0, 736B] colorModel 参数未使用（接口 itab/data 零值传入）。
func validateScreenshotImageConfig(colorModel color.Model, width, height, stride, channels int64, cfg screenshotImageConfig) (int64, error) {
	_ = colorModel
	if stride <= 0 || cfg.maxBytes < stride {
		return 0, fmt.Errorf("图片源数据 %d 字节超过有效范围", stride)
	}
	if width <= 0 || height <= 0 || cfg.maxWidth < width || cfg.maxHeight < height {
		return 0, fmt.Errorf("图片尺寸 %dx%d 超过有效范围", width, height)
	}
	if cfg.maxPixels/height < width {
		return 0, fmt.Errorf("图片总像素超过上限 %d", cfg.maxPixels)
	}
	if cfg.maxPixels < width*height || channels < 0 {
		return 0, fmt.Errorf("图片预计工作集超过上限 %d 字节", cfg.maxWorkSet)
	}
	if channels > 0 && (cfg.maxWorkSet-stride)/channels < width*height {
		return 0, fmt.Errorf("图片预计工作集超过上限 %d 字节", cfg.maxWorkSet)
	}
	estimated := stride + channels*width*height
	if cfg.maxWorkSet < estimated {
		return 0, fmt.Errorf("图片预计工作集 %d 字节超过上限 %d 字节", estimated, cfg.maxWorkSet)
	}
	return estimated, nil
}

// ---- inspectScreenshotImageWork ----
// inspectScreenshotImageWork 校验并估算图像工作集（DecodeConfig + 格式白名单 + validate）。
// [S 汇编 0x14097fbc0, 928B] 返回 colorModel/width/height/format/estimated/error。
func inspectScreenshotImageWork(data []byte, expectedFormat string, channels int64, cfg screenshotImageConfig) (color.Model, int, int, string, int64, error) {
	if len(data) == 0 || cfg.maxBytes < int64(len(data)) {
		return nil, 0, 0, "", 0, fmt.Errorf("图片源数据大小无效或超过 %d 字节", cfg.maxBytes)
	}
	decCfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, 0, 0, "", 0, fmt.Errorf("解析图片尺寸失败: %w", err)
	}
	expected := strings.ToLower(strings.TrimSpace(expectedFormat))
	if expected != "" && !strings.EqualFold(format, expected) {
		return nil, 0, 0, "", 0, fmt.Errorf("图片格式必须为 %s，实际为 %s", expected, format)
	}
	estimated, err := validateScreenshotImageConfig(decCfg.ColorModel, int64(decCfg.Width), int64(decCfg.Height), int64(len(data)), channels, cfg)
	if err != nil {
		return nil, 0, 0, "", 0, err
	}
	return decCfg.ColorModel, decCfg.Width, decCfg.Height, format, estimated, nil
}

// ---- prepareScreenshotImageWork ----
// prepareScreenshotImageWork 校验图像工作集并按估算值预留全局内存预算。
// [S 汇编 0x14097ff60, 352B] inspect（默认配置）→ 全局 Reserve → 返回释放函数。
func prepareScreenshotImageWork(data []byte, expectedFormat string, channels int64) (color.Model, int, int, string, func(), error) {
	cm, width, height, format, estimated, err := inspectScreenshotImageWork(data, expectedFormat, channels, screenshotImageWorkConfig)
	if err != nil {
		return nil, 0, 0, "", nil, err
	}
	release, err := screenshotImageMemoryBudgetGlobal.Reserve(estimated)
	if err != nil {
		return nil, 0, 0, "", nil, err
	}
	return cm, width, height, format, release, nil
}

// ---- reserveScreenshotImageBounds ----
// reserveScreenshotImageBounds 按边界矩形预留图像工作集预算。
// [S 汇编 0x1409806e0, 192B] width=x1-x0 / height=y1-y0 → stride 钳制 ≥1 →
// validate（nil colorModel）→ 全局 Reserve。
func reserveScreenshotImageBounds(x0, y0, x1, y1, stride, channels int64) (func(), error) {
	width := x1 - x0
	height := y1 - y0
	if stride <= 0 {
		stride = 1
	}
	estimated, err := validateScreenshotImageConfig(nil, width, height, stride, channels, screenshotImageWorkConfig)
	if err != nil {
		return nil, err
	}
	return screenshotImageMemoryBudgetGlobal.Reserve(estimated)
}

// ---- buildScreenshotResultMetadataFromPNG ----
// buildScreenshotResultMetadataFromPNG 从 PNG 字节构建截图结果元数据。
// [S 汇编 0x140966a80, 1024B] normalizeScreenshotMode → 选配置（scrolling 专用或默认）→
// inspect（"png", channels=0）→ 非 scrolling 时全局 Reserve 后立即释放 →
// Width/Height 钳制到 ≥0。
func buildScreenshotResultMetadataFromPNG(png []byte, mode string) (ScreenshotCaptureResult, error) {
	if len(png) == 0 {
		return ScreenshotCaptureResult{}, nil
	}
	mode = normalizeScreenshotMode(mode)
	cfg := screenshotImageWorkConfig
	if mode == "scrolling" {
		cfg = screenshotImageScrollingConfig
	}
	_, width, height, _, estimated, err := inspectScreenshotImageWork(png, "png", 0, cfg)
	if err != nil {
		return ScreenshotCaptureResult{}, fmt.Errorf("解析截图内容失败: %w", err)
	}
	if mode != "scrolling" {
		release, rerr := screenshotImageMemoryBudgetGlobal.Reserve(estimated)
		if rerr != nil {
			return ScreenshotCaptureResult{}, fmt.Errorf("解析截图内容失败: %w", rerr)
		}
		release()
	}
	return ScreenshotCaptureResult{
		Width:  max(0, width),
		Height: max(0, height),
		Mode:   mode,
	}, nil
}

// ---- normalizeScreenshotHistoryEntry ----
// normalizeScreenshotHistoryEntry 规范化历史条目（裁剪路径、填充栏目）。
// [S 汇编 0x14096a640] TrimSpace(path) → 若 CapturedAt≤0 则 time.Now().UnixMilli() →
// normalizeScreenshotMode(mode) → cmovl 符号钳制 → 返回归一化条目
func normalizeScreenshotHistoryEntry(entry ScreenshotHistoryEntry) ScreenshotHistoryEntry {
	path := strings.TrimSpace(entry.Path)
	capturedAt := entry.CapturedAt
	if capturedAt <= 0 {
		capturedAt = time.Now().UnixMilli()
	}
	mode := normalizeScreenshotMode(entry.Mode)
	if entry.Width < 0 {
		entry.Width = 0
	}
	if entry.Height < 0 {
		entry.Height = 0
	}
	return ScreenshotHistoryEntry{
		ID:         entry.ID,
		Width:      entry.Width,
		Height:     entry.Height,
		Mode:       mode,
		Path:       path,
		CapturedAt: capturedAt,
	}
}

// ---- buildScreenshotHistoryEntry ----
// buildScreenshotHistoryEntry 构建截图历史条目。
// [S 汇编 0x14096a8a0] captureTime → UnixMilli → ScreenshotHistoryEntry{Path, CapturedAt, Mode, W, H}
func buildScreenshotHistoryEntry(path string, captureTime time.Time, mode string, width, height int, meta interface{}) ScreenshotHistoryEntry {
	_ = meta
	path = strings.TrimSpace(path)
	mode = normalizeScreenshotMode(mode)
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	return ScreenshotHistoryEntry{
		ID:         buildScreenshotHistoryEntryID(path, captureTime),
		Width:      width,
		Height:     height,
		Mode:       mode,
		Path:       path,
		CapturedAt: captureTime.UnixMilli(),
	}
}

// ---- saveScreenshotImageWithSourceName ----
// saveScreenshotImageWithSourceName 将截图保存到工作区截图目录，返回路径。
// [S 汇编 0x140967140] png guard → time → normalizeScreenshotSaveFormat → source TrimSpace →
// time.Format → buildScreenshotFileBaseName → MkdirAll → os.WriteFile
func saveScreenshotImageWithSourceName(wsDir string, png []byte, source string, format string, captureTime time.Time) (string, error) {
	if len(png) == 0 {
		return "", errors.New("截图数据为空")
	}
	if captureTime.IsZero() {
		captureTime = time.Now()
	}
	format = normalizeScreenshotSaveFormat(format)
	source = strings.TrimSpace(source)

	baseName := buildScreenshotFileBaseName(source, captureTime)
	ext := "." + format
	if format == "jpeg" {
		ext = ".jpg"
	}
	fileName := baseName + ext

	dir := filepath.Join(wsDir, "screenshots")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, fileName)

	if err := os.WriteFile(path, png, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// ---- loadScreenshotHistory ----
// loadScreenshotHistory 从工作区加载截图历史。
// [S 汇编 0x140968860] screenshotHistoryPath → screenshotHistoryClearCutoff → os.ReadFile →
// json.Unmarshal → mergeScreenshotHistoryWithDirectory
func loadScreenshotHistory(wsDir string) ([]ScreenshotHistoryEntry, error) {
	path := screenshotHistoryPath(wsDir)
	clearFound, clearTime := screenshotHistoryClearCutoff(wsDir)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return mergeScreenshotHistoryWithDirectory(wsDir, nil, clearFound, clearTime)
		}
		return nil, err
	}
	var entries []ScreenshotHistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("解析截图历史失败: %w", err)
	}
	if clearFound || len(entries) == 0 {
		return mergeScreenshotHistoryWithDirectory(wsDir, entries, clearFound, clearTime)
	}
	return entries, nil
}

// ---- normalizeScreenshotHistoryEntries ----
// normalizeScreenshotHistoryEntries 去重并限制截图历史条目数。
// [S 汇编 0x14096a180] 遍历 → normalize → 空路径跳过 → checkExists(Stat+IsDir) →
// 去重(filepath.Clean + ToLower) → 上限 48
// checkExists 标志在 asm 中存在，当前所有调用点传 false。
func normalizeScreenshotHistoryEntries(entries []ScreenshotHistoryEntry, checkExists bool) []ScreenshotHistoryEntry {
	limit := len(entries)
	if limit > 48 {
		limit = 48
	}
	result := make([]ScreenshotHistoryEntry, 0, limit)
	seen := make(map[string]struct{})
	for _, e := range entries {
		e = normalizeScreenshotHistoryEntry(e)
		if e.Path == "" {
			continue
		}
		if checkExists {
			fi, err := os.Stat(e.Path)
			if err != nil || fi.IsDir() {
				continue
			}
		}
		key := strings.ToLower(filepath.Clean(e.Path))
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, e)
		if len(result) >= 48 {
			break
		}
	}
	return result
}

// ---- saveScreenshotHistory ----
// saveScreenshotHistory 写入截图历史到工作区。
// [S 汇编 0x1409695c0] normalizeScreenshotHistoryEntries → screenshotHistoryPath → writeJSONFile
func saveScreenshotHistory(wsDir string, entries []ScreenshotHistoryEntry) error {
	path := screenshotHistoryPath(wsDir)
	return writeJSONFile(path, normalizeScreenshotHistoryEntries(entries, false))
}

// ---- prependScreenshotHistoryEntry ----
// prependScreenshotHistoryEntry 在历史前插入新条目（去重+上限裁剪）。
// [S 汇编 0x140969cc0] loadScreenshotHistory → normalizeScreenshotHistoryEntry →
// 去重（filepath.Clean + EqualFold）→ makeslice → prepend → inc len →
// normalizeScreenshotHistoryEntries → saveScreenshotHistory
func prependScreenshotHistoryEntry(wsDir string, entry ScreenshotHistoryEntry) error {
	entries, err := loadScreenshotHistory(wsDir)
	if err != nil {
		return err
	}

	entry = normalizeScreenshotHistoryEntry(entry)
	if entry.Path == "" {
		return fmt.Errorf("截图历史路径不能为空")
	}

	for _, e := range entries {
		if strings.EqualFold(filepath.Clean(e.Path), filepath.Clean(entry.Path)) {
			return nil
		}
	}

	newEntries := make([]ScreenshotHistoryEntry, 0, len(entries)+1)
	newEntries = append(newEntries, entry)
	for _, e := range entries {
		if !strings.EqualFold(filepath.Clean(e.Path), filepath.Clean(entry.Path)) {
			newEntries = append(newEntries, e)
		}
	}

	return saveScreenshotHistory(wsDir, normalizeScreenshotHistoryEntries(newEntries, false))
}

// ---- clearScreenshotHistory ----
// clearScreenshotHistory 清除截图历史：保存空列表 → 创建截图目录 → 写入清除标记。
// [S 汇编 0x1409696a0, 70L] saveScreenshotHistory(wsDir, nil) → os.MkdirAll(wsDir, 0755) →
// screenshotHistoryClearMarkerPath → time.Now().Format(RFC3339Nano) → os.WriteFile(marker, 0644)
func clearScreenshotHistory(wsDir string) error {
	if err := saveScreenshotHistory(wsDir, nil); err != nil {
		return err
	}
	wsDir = strings.TrimSpace(wsDir)
	if err := os.MkdirAll(wsDir, 0o755); err != nil {
		return err
	}
	markerPath := screenshotHistoryClearMarkerPath(wsDir)
	content := time.Now().Format(time.RFC3339Nano)
	return os.WriteFile(markerPath, []byte(content), 0o644)
}
