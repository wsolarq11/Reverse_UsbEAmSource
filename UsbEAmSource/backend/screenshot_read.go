// AUTO-RECONSTRUCTED — Screenshot domain：本地截图读取 / 解码 / 路径校验链
// 研究用途
// 档位：[S] 反汇编实证（VA 见各函数注释；常量均已按 HANDOFF 第 7 节方法从 .rdata 解码）
package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// screenshotImageMetadata 是本地截图文件读出的元数据。
// 栈布局实证（0x14096ae20 成功路径，结构体基址 rsp+0x88）：
//
//	+0x00..0x30  ScreenshotImageRef 静态模板（0x1409613F0 处 48B duffcopy 源）
//	+0x30 Width   (rsp+0xb8)
//	+0x38 Height  (rsp+0xc0)
//	+0x40 Mode    (rsp+0xc8, 16B, 保持零值)
//	+0x50 Path    (rsp+0xd8)
//	总 0x60 字节 —— 与 duffcopy+0x31e 的复制跨度一致
type screenshotImageMetadata struct {
	Ref    ScreenshotImageRef
	Width  int
	Height int
	Mode   string
	Path   string
}

// isSupportedScreenshotImageExtension 判断路径是否为受支持的截图扩展名。
// [S 汇编 0x140967dc0] 仅接受 .png / .jpg / .jpeg（大小写不敏感）。
func isSupportedScreenshotImageExtension(path string) bool {
	switch strings.ToLower(filepath.Ext(strings.TrimSpace(path))) {
	case ".png", ".jpg", ".jpeg":
		return true
	}
	return false
}

// ensureScreenshotPathInsideDirectory 校验 rel 落在 dir 之内且扩展名受支持。
// [S 汇编 0x14096c5c0, 411B] 实证链：
//
//	filepath.Abs(TrimSpace(dir)) → filepath.Abs(TrimSpace(rel)) → filepath.Rel →
//	拒绝 "." / ".." / `..\`（3B 魔数 @0x140C33B4E）/ filepath.IsAbs →
//	isSupportedScreenshotImageExtension → 否则报错。
//
// 不变量：越界判定用 Rel 结果的前缀而非字符串包含，故 `..\foo` 与绝对路径均被拒。
func ensureScreenshotPathInsideDirectory(dir, rel string) error {
	absDir, err := filepath.Abs(strings.TrimSpace(dir))
	if err != nil {
		return err
	}
	absPath, err := filepath.Abs(strings.TrimSpace(rel))
	if err != nil {
		return err
	}
	inside, err := filepath.Rel(absDir, absPath)
	if err != nil {
		return err
	}
	if inside == "." || inside == ".." || strings.HasPrefix(inside, `..\`) || filepath.IsAbs(inside) {
		return errors.New("截图路径不在截图目录内")
	}
	if !isSupportedScreenshotImageExtension(absPath) {
		return errors.New("仅支持读取 PNG 或 JPG 截图")
	}
	return nil
}

// readScreenshotImageMetadataFromPath 读取截图尺寸元数据。
// [S 汇编 0x14096ae20, 806B] 实证链：
//
//	TrimSpace(rel) → 空则 "截图路径不能为空"(0x140C654D8, 24B) →
//	ensureScreenshotPathInsideDirectory(dir, trimmed) →
//	os.OpenFile(trimmed, O_RDONLY|0, 0)（ecx/edi 双零）→ defer Close →
//	image.DecodeConfig → 失败 fmt.Errorf("解析截图尺寸失败: %w", ...)（@0x140C6BEEC, 28B）→
//	成功按 max(0, ·) 钳制宽高并回填 Path。
func readScreenshotImageMetadataFromPath(dir, rel string) (screenshotImageMetadata, error) {
	trimmed := strings.TrimSpace(rel)
	if trimmed == "" {
		return screenshotImageMetadata{}, errors.New("截图路径不能为空")
	}
	if err := ensureScreenshotPathInsideDirectory(dir, trimmed); err != nil {
		return screenshotImageMetadata{}, err
	}
	f, err := os.OpenFile(trimmed, os.O_RDONLY, 0)
	if err != nil {
		return screenshotImageMetadata{}, err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return screenshotImageMetadata{}, fmt.Errorf("解析截图尺寸失败: %w", err)
	}
	return screenshotImageMetadata{
		Width:  max(0, cfg.Width),
		Height: max(0, cfg.Height),
		Path:   trimmed,
	}, nil
}

// readScreenshotImageFileLimited 受限读取截图文件字节。
// [S 汇编 0x140980260, 1056B] TrimSpace → os.OpenFile(O_RDONLY,0) → defer Close →
// Stat → IsDir/size≤0/size>maxBytes 均报错 → io.LimitedReader(N=maxBytes+1) 读全 →
// 读后 size 复核（不一致报"文件大小发生变化"）。
func readScreenshotImageFileLimited(path string) ([]byte, error) {
	f, err := os.OpenFile(strings.TrimSpace(path), os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if info.IsDir() || info.Size() <= 0 || screenshotImageWorkConfig.maxBytes < info.Size() {
		return nil, fmt.Errorf("图片文件大小无效或超过 %d 字节", screenshotImageWorkConfig.maxBytes)
	}
	data, err := io.ReadAll(&io.LimitedReader{R: f, N: screenshotImageWorkConfig.maxBytes + 1})
	if err != nil {
		return nil, err
	}
	if info.Size() != int64(len(data)) {
		return nil, errors.New("读取图片时文件大小发生变化，请重试")
	}
	return data, nil
}

// decodeScreenshotImageBytesWithBudget 在内存预算内解码图片字节。
// [S 汇编 0x1409800c0, 416B] prepare（默认配置 + 全局 Reserve）→ image.Decode；
// 解码失败时先释放预算再报错，成功返回 (img, format, release, error)。
func decodeScreenshotImageBytesWithBudget(data []byte, expectedFormat string, channels int64) (image.Image, string, func(), error) {
	_, _, _, _, release, err := prepareScreenshotImageWork(data, expectedFormat, channels)
	if err != nil {
		return nil, "", nil, err
	}
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		release()
		return nil, "", nil, fmt.Errorf("解析图片失败: %w", err)
	}
	return img, format, release, nil
}

// screenshotImageToRGBA 将任意解码图像转为 *image.RGBA。
// [S 汇编 0x14096be00] 已是 *image.RGBA 则直接返回，否则重绘到 RGBA 画布。
func screenshotImageToRGBA(img image.Image) *image.RGBA {
	if img == nil {
		return nil
	}
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}
	b := img.Bounds()
	dst := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Set(x, y, img.At(x, y))
		}
	}
	return dst
}

// buildScreenshotResultFromPNG 从 PNG 字节构造截图结果（含尺寸与 dataURL）。
// [S 汇编 0x1409667e0] decodeConfig → 填 Width/Height/ImageData。
func buildScreenshotResultFromPNG(png []byte) ScreenshotCaptureResult {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(png))
	result := ScreenshotCaptureResult{
		ImageData: encodeScreenshotDataURL(png),
	}
	if err != nil {
		return result
	}
	result.Width = cfg.Width
	result.Height = cfg.Height
	return result
}

// readExternalScreenshotImageFromPath 读取外部（非截图目录）图片文件。
// [S 汇编 0x14096c100, 1280B] 实证链：
//
//	TrimSpace → 空则 "图片路径不能为空"(0x140C654F0, 24B) →
//	readScreenshotImageFileLimited → decodeScreenshotImageBytesWithBudget
//	（失败 fmt.Errorf("解析图片失败: %w", @0x140C6148E, 22B)）→
//	screenshotImageToRGBA → encodeScreenshotRGBAWithKlauspostPNG
//	（失败 fmt.Errorf("转换图片失败: %w", @0x140C614A4, 22B)）→
//	buildScreenshotResultFromPNG
func readExternalScreenshotImageFromPath(path string) (ScreenshotCaptureResult, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ScreenshotCaptureResult{}, errors.New("图片路径不能为空")
	}
	raw, err := readScreenshotImageFileLimited(trimmed)
	if err != nil {
		return ScreenshotCaptureResult{}, err
	}
	img, _, release, err := decodeScreenshotImageBytesWithBudget(raw, "", 12)
	if err != nil {
		return ScreenshotCaptureResult{}, fmt.Errorf("解析图片失败: %w", err)
	}
	defer release()
	png, err := encodeScreenshotRGBAWithKlauspostPNG((*screenshotRGBAImageRowSource)(screenshotImageToRGBA(img)))
	if err != nil {
		return ScreenshotCaptureResult{}, fmt.Errorf("转换图片失败: %w", err)
	}
	return buildScreenshotResultFromPNG(png), nil
}

// screenshotImageToOpaqueRGBA 将图像转为不透明 RGBA（NewRGBA + DrawMask）。
// [S-sig 0x14096bf60, 416B]：img nil/越界→nil；bounds→子矩形→image.NewRGBA →
// DrawMask×2（不透明填充 + 绘制）。体待掩码域专项还原。
func screenshotImageToOpaqueRGBA(a, b interface{}) interface{} {
	_, _ = a, b
	return nil
}
