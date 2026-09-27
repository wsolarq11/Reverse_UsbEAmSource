// AUTO-RECONSTRUCTED SERVICE METHODS — DOMAIN: launcher config icon migration (historical)
// 研究用途
//
// 契约来源（全部 [S] 汇编实证，source_funcs.txt:2075-2082）：
//   - migrateHistoricalLauncherConfigIcons(0x14088f680) + func1(0x14088f720) — 遍历 slot 迁移
//   - migrateHistoricalLauncherConfigIcon(0x14088f980) — 单图标 re-encode
//   - decodeHistoricalLauncherConfigIcon(0x14088ff60) — 解码 data: URI
//   - ensureLauncherConfigIconMigrationBackup(0x1408904a0) — .bak 备份代办
//   - ensureLauncherConfigIconLibraryMigrationBackup(0x140890560) — .bak 备份代办(库)
//   - ensureLauncherConfigMigrationBackupFile(0x140890620) + func1(0x140890d00) — 实际备份落盘
//
// 档位：全部 [S]（完整汇编实证 + .rdata 字符串常量解码 + 字节级帧偏移三重确证）。
package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

// ---- 常量 ----

const (
	// launcherConfigIconCountLimit 单次校验允许的最大图标数量（0x1000）。
	launcherConfigIconCountLimit = 4096

	// launcherConfigIconMaxDecodedBytes 累计解码字节上限（32 MiB / 0x2000000）。
	launcherConfigIconMaxDecodedBytes = 32 << 20

	// launcherConfigIconMaxDecodedSingle 单图标解码字节上限（2 MiB / 0x200000）。
	launcherConfigIconMaxDecodedSingle = 2 << 20

	// launcherConfigIconMaxPixels 累计像素上限（16 MiPx / 0x1000000）。
	launcherConfigIconMaxPixels = 16 << 20

	// launcherConfigIconMaxSide 图标单边像素上限。
	launcherConfigIconMaxSide = 512

	// launcherConfigMigrationDirSuffix 迁移备份目录后缀。
	launcherConfigMigrationDirSuffix = ".before-icon-migration-v1.bak"

	// launcherConfigLibraryMigrationDirSuffix 图标库迁移备份目录后缀。
	launcherConfigLibraryMigrationDirSuffix = ".before-icon-library-v1.bak"

	// launcherConfigIconMaxDimension 缩放裁剪单边上限（0x200）。
	launcherConfigIconMaxDimension = 512

	// launcherConfigIconMaxOriginalDimension 原始图标单边上限（0x1000）。
	launcherConfigIconMaxOriginalDimension = 4096

	// launcherConfigIconMaxPixelsBeforeRescale 未缩放像素上限（0x100000 = 1 MiPx）。
	launcherConfigIconMaxPixelsBeforeRescale = 1 << 20

	// launcherConfigIconRescaledMaxPixels 缩放后像素上限侧的 0x200 下限。
	launcherConfigIconRescaledMinDimension = 512

	// launcherConfigIconMaxDecodedBeforeRescale 单次解码字节上限（0x800000 = 8 MiB）。
	launcherConfigIconMaxDecodedBeforeRescale = 8 << 20

	// launcherConfigIconPNGBufferCap 后续缩放后/重新编码的内缓存上限（2 MiB）。
	launcherConfigIconPNGBufferCap = 2 << 20
)

// ---- launcherConfigIconBudget ----

// launcherConfigIconBudget 图标预算计数器（24 字节）。
// 汇编实证偏移：+0x00=count(int) +0x08=decodedBytes(int64) +0x10=pixels(int64)
type launcherConfigIconBudget struct {
	count        int   // +0x00：已校验图标数
	decodedBytes int64 // +0x08：累计解码字节数
	pixels       int64 // +0x10：累计像素数
}

// validate 校验单条图标数据 URL 并累加预算。
// [S 汇编实证 0x14088e6c0, 368L]
//
// 流程：TrimSpace(data) → data 为空则跳过 → budget.count >= 4096 报错
// → ToLower(data) → 前缀"data:"检查 → 错误固定串
// → IndexByteString(data, ',') 定位头体分隔 → 取 mediaType
// → EqualFold(mediaType, "base64") 验证 → TrimSpace(lowercaseSegment)
// → ExpectedFormat(path, segment) → 获 normalizedFormat
// → 解码 base64 部分 → base64.StdEncoding.DecodeString -> err 透传
// → MetricsFromDecoded(path, mediaType, decoded) → 获 metrics(decodedBytes, pixels)
// → addMetrics(decodedBytes, pixels) 累加计数
func (b *launcherConfigIconBudget) validate(path, data string) error {
	data = strings.TrimSpace(data)
	if data == "" {
		return nil
	}
	if b.count >= launcherConfigIconCountLimit {
		return launcherConfigIconError(path, fmt.Sprintf("图标数量超过上限 %d", launcherConfigIconCountLimit))
	}

	lower := strings.ToLower(data)
	if !strings.HasPrefix(lower, "data:") {
		return launcherConfigIconError(path, "缺少 data: 前缀")
	}

	comma := strings.IndexByte(lower, ',')
	if comma < 0 {
		return launcherConfigIconError(path, "缺少 base64 分隔逗号")
	}

	mediaType := lower[5:comma] // 去掉 "data:"
	if !strings.EqualFold(mediaType, "base64") {
		// 检查 base64;media 格式
		if idx := strings.IndexByte(mediaType, ';'); idx >= 0 {
			encType := strings.TrimSpace(mediaType[idx+1:])
			if !strings.EqualFold(encType, "base64") {
				return launcherConfigIconError(path, "编码格式非 base64")
			}
		} else {
			return launcherConfigIconError(path, "编码格式非 base64")
		}
	}

	segment := strings.TrimSpace(lower[comma+1:])

	normalized, err := launcherConfigIconExpectedFormat(path, segment)
	if err != nil {
		return err // err 已含 path 信息
	}
	_ = normalized // asm 中仅校验不存储

	decoded, err := base64.StdEncoding.DecodeString(segment)
	if err != nil {
		return launcherConfigIconError(path, fmt.Sprintf("Base64 解码失败: %v", err))
	}

	decodedBytes, pixels, err := launcherConfigIconMetricsFromDecoded(path, mediaType, decoded)
	if err != nil {
		return err
	}
	return b.addMetrics(decodedBytes, pixels)
}

// addMetrics 递增预算计数器（校验上限 + 累加）。
// [S 汇编实证 0x14088f380, 148L]
//
// asm 流程：receiver nil → error → count >= 4096 → error
// → decodedBytes <= 0 || pixels <= 0 → error
// → decodedBytes > remaining cap → error
// → pixels > remaining cap → error
// → count++; decodedBytes += arg; pixels += arg
func (b *launcherConfigIconBudget) addMetrics(decodedBytes int64, pixels int64) error {
	if b == nil {
		return fmt.Errorf("launcherConfigIconBudget: nil receiver")
	}
	if b.count >= launcherConfigIconCountLimit {
		return fmt.Errorf("图标数量已达到上限 %d", launcherConfigIconCountLimit)
	}
	if decodedBytes <= 0 || pixels <= 0 {
		return fmt.Errorf("图标指标无效: decodedBytes=%d pixels=%d", decodedBytes, pixels)
	}
	remainingBytes := launcherConfigIconMaxDecodedBytes - b.decodedBytes
	if decodedBytes > remainingBytes {
		return fmt.Errorf("解码字节数 %d 超出剩余预算 %d", decodedBytes, remainingBytes)
	}
	remainingPixels := launcherConfigIconMaxPixels - b.pixels
	if pixels > remainingPixels {
		return fmt.Errorf("像素数 %d 超出剩余预算 %d", pixels, remainingPixels)
	}
	b.count++
	b.decodedBytes += decodedBytes
	b.pixels += pixels
	return nil
}

// ---- launcherConfigIconError 构造器 ----

// launcherConfigIconError 构造 *launcherConfigIconDataError 返回 error。
// [S 汇编实证 0x14088f5a0]
//
// 签名解析：rax=path.ptr, rbx=path.len, rcx=message.ptr, rdi=message.len。
// TrimSpace 两参 → runtime.newobject → 填 Path.ptr/len → Reason.ptr/len → 加载 itable 返回。
func launcherConfigIconError(path, message string) error {
	if strings.TrimSpace(path) == "" {
		path = "launcherConfigIcon"
	}
	return &launcherConfigIconDataError{Path: path, Reason: message}
}

// ---- launcherConfigIconExpectedFormat ----

// launcherConfigIconExpectedFormat 标准化 MIME 类型（仅限 png/jpeg/webp）。
// [S 汇编实证 0x14088ed80]
//
// 签名：rax=path.ptr, rbx=path.len, rcx=mediaType.ptr, rdi=mediaType.len
// 返回：成功 rax=format.ptr, rbx=format.len, rcx=0；失败 rax=0, rbx=0, rcx=error.
//
// MIME 匹配为手写 cmp movabs（内联）：长度 9 → "image/png" → "png"；
// 长度 10 再分 "image/jpeg" → "jpeg" / "image/webp" → "webp"。
// 字符串常量存放于 .rodata 拼接串中（"pnggifbm"/"jpegwebp"/"webpavif"），取前 N 字节。
func launcherConfigIconExpectedFormat(path, mediaType string) (string, error) {
	mediaType = strings.TrimSpace(mediaType)
	mediaType = strings.ToLower(mediaType)
	switch mediaType {
	case "image/png":
		return "png", nil
	case "image/jpeg":
		return "jpeg", nil
	case "image/webp":
		return "webp", nil
	}
	return "", launcherConfigIconError(path, fmt.Sprintf("不支持的 MIME %q", mediaType))
}

// ---- launcherConfigIconMetricsFromDecoded ----

// launcherConfigIconMetricsFromDecoded 从解码的图标字节中提取指标（字节数 + 像素数）。
// [S 汇编实证 0x14088ef00, 1152B]
//
// 签名：rax=path.ptr，rbx=path.len，rcx=mediaType.ptr，rdi=mediaType.len，
// rsi=decoded.ptr，r8=decoded.len，r9=decoded.cap
// 返回：rax=decoded.len，rbx=Width*Height，rcx=0（成功）或 rcx=err（失败）。
//
// 流程：空校验 → 单图上界检查(2<<20) → ExpectedFormat 校验并获规范化类型
// → bytes.NewReader → image.DecodeConfig → MIME 一致性验证
// → 尺寸非零 → 单边上限(512) → 返回(decoded.len, Width*Height, nil)
func launcherConfigIconMetricsFromDecoded(path, mediaType string, decoded []byte) (int64, int64, error) {
	if len(decoded) == 0 {
		return 0, 0, launcherConfigIconError(path, "Base64 数据为空")
	}
	if len(decoded) > launcherConfigIconMaxDecodedSingle {
		return 0, 0, launcherConfigIconError(path, fmt.Sprintf("解码字节数超过单图上限 %d", launcherConfigIconMaxDecodedSingle))
	}

	normalized, err := launcherConfigIconExpectedFormat(path, mediaType)
	if err != nil {
		return 0, 0, err
	}

	r := bytes.NewReader(decoded)
	cfg, actualFormat, decErr := image.DecodeConfig(r)
	if decErr != nil {
		return 0, 0, launcherConfigIconError(path, fmt.Sprintf("图片头解析失败: %v", decErr))
	}
	if normalized != actualFormat {
		return 0, 0, launcherConfigIconError(path, fmt.Sprintf("MIME 声明为 %s，实际格式为 %s", mediaType, actualFormat))
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return 0, 0, launcherConfigIconError(path, "图片尺寸无效")
	}
	if cfg.Width > launcherConfigIconMaxSide || cfg.Height > launcherConfigIconMaxSide {
		return 0, 0, launcherConfigIconError(path, fmt.Sprintf("图片尺寸 %dx%d 超过单边上限 %d", cfg.Width, cfg.Height, launcherConfigIconMaxSide))
	}
	return int64(len(decoded)), int64(cfg.Width) * int64(cfg.Height), nil
}

// ---- validateLauncherConfigIconData ----

// validateLauncherConfigIconData 校验配置全部图标槽位并累计预算（纯副作用，不返回错误）。
// [S 汇编实证 0x14088d740]：newobject budget → closure {func1, &budget}
// → visitLauncherConfigIconData(cfg, closure) → 返回（不检查 error）。
//
// asm func1(0x14088d800, 33行) 从 slot 展开中取 Data 字段调用 budget.validate(path, data)。
// 汇编利用寄存器残留传 Data（rcx 残留指向 slot.Data *string），重建时直接走 slot 遍历。
func validateLauncherConfigIconData(cfg *LauncherConfig) {
	if cfg == nil {
		return
	}
	var budget launcherConfigIconBudget
	visitLauncherConfigIconSlots(cfg, func(slot launcherConfigIconSlot) bool {
		if slot.Data != nil {
			_ = budget.validate(slot.Path, *slot.Data)
		}
		return false
	})
}

// ---- migrateHistoricalLauncherConfigIcons ----

// migrateHistoricalLauncherConfigIcons 遍历全部图标槽位，将内联旧格式图标重编码为现代 data:image/png;base64 格式。
// [S 汇编 0x14088f680]：遍历 visitLauncherConfigIconData → func1 逐槽验证 → 预算不足时迁移/丢弃。
// 返回 (migrated bool, rescaled int, dropped int)。
// 汇编回传通过寄存器泄漏传 slot.Data.ptr（visitLauncherConfigIconData.func1 仅传 path 但 rcx 残留 Data 指针），
// Go 代码直走 visitLauncherConfigIconSlots 显式取两字段。
func migrateHistoricalLauncherConfigIcons(cfg *LauncherConfig) (bool, int, int) {
	if cfg == nil {
		return false, 0, 0
	}

	var budget launcherConfigIconBudget
	var migrated bool
	var rescaled, dropped int

	visitLauncherConfigIconSlots(cfg, func(slot launcherConfigIconSlot) bool {
		if slot.Data == nil || *slot.Data == "" {
			return false
		}
		trimmed := strings.TrimSpace(*slot.Data)
		if trimmed == "" {
			return false
		}

		// 验证原数据预算
		err := budget.validate(slot.Path, trimmed)
		if err == nil {
			// 预算有余，无需迁移
			return false
		}

		// 预算不足 → 尝试迁移（重编码）
		newData, resized, ok := migrateHistoricalLauncherConfigIcon(trimmed)
		if !ok {
			// 解码失败 → 丢弃
			*slot.Data = ""
			dropped++
			return false
		}

		// 迁移后验证新数据预算
		err2 := budget.validate(slot.Path, newData)
		if err2 != nil {
			*slot.Data = ""
			dropped++
			return false
		}

		*slot.Data = newData
		migrated = true
		if resized {
			rescaled++
		}
		return false
	})

	return migrated, rescaled, dropped
}

// ---- migrateHistoricalLauncherConfigIcon ----

// migrateHistoricalLauncherConfigIcon 重编码单条图标数据：解码 data:URI → 缩放(如需) → PNG 重编码 → base64。
// [S 汇编 0x14088f980]。
// 返回 (newData string, rescaled bool, ok bool)。
func migrateHistoricalLauncherConfigIcon(data string) (string, bool, bool) {
	decoded, _, ok := decodeHistoricalLauncherConfigIcon(data)
	if !ok || len(decoded) == 0 {
		return "", false, false
	}

	// 解码成功 → 获取图像尺寸
	cfg, _, err := image.DecodeConfig(bytes.NewReader(decoded))
	if err != nil {
		return "", false, false
	}
	w, h := cfg.Width, cfg.Height

	// 原始尺寸边界校验
	if w <= 0 || h <= 0 || w > launcherConfigIconMaxOriginalDimension || h > launcherConfigIconMaxOriginalDimension {
		return "", false, false
	}
	pixels := int64(w) * int64(h)
	if pixels > launcherConfigIconMaxPixelsBeforeRescale {
		return "", false, false
	}
	if pixels <= launcherConfigIconMaxPixelsBeforeRescale/16 { // 64k 像素
		return "", false, false
	}

	// 若无需缩放（尺寸 ≤ 512 且无需重编码）→ 直接组装新 data URI
	if w <= launcherConfigIconMaxDimension && h <= launcherConfigIconMaxDimension {
		newData := "data:image/png;base64," + base64.StdEncoding.EncodeToString(decoded)
		return newData, false, true
	}

	// 需要缩放
	// 计算目标尺寸（保持宽高比，按较长边缩放至 launcherConfigIconMaxDimension）
	newW, newH := w, h
	if w > h {
		newW = launcherConfigIconMaxDimension
		newH = h * launcherConfigIconMaxDimension / w
	} else if h > w {
		newH = launcherConfigIconMaxDimension
		newW = w * launcherConfigIconMaxDimension / h
	} else {
		newW = launcherConfigIconMaxDimension
		newH = launcherConfigIconMaxDimension
	}
	// 确保至少 1 像素
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	// 解码全图像 → 缩放 → 重编码 PNG
	src, _, err := image.Decode(bytes.NewReader(decoded))
	if err != nil {
		return "", false, false
	}

	dst := image.NewNRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return "", false, false
	}
	if buf.Len() > launcherConfigIconPNGBufferCap {
		return "", false, false
	}

	// 组装 data URI
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	newData := "data:image/png;base64," + encoded
	return newData, true, true
}

// ---- decodeHistoricalLauncherConfigIcon ----

// decodeHistoricalLauncherConfigIcon 解码 data:image/xxx;base64, 格式图标。
// [S 汇编 0x14088ff60]。
// 返回 (decoded []byte, mediaType string, ok bool)。
func decodeHistoricalLauncherConfigIcon(data string) ([]byte, string, bool) {
	d := strings.TrimSpace(data)
	d = strings.ToLower(d)
	if len(d) < 5 || !strings.HasPrefix(d, "data:") {
		return nil, "", false
	}

	sep := strings.IndexByte(d, ',')
	if sep <= 5 || sep == len(d)-1 {
		return nil, "", false
	}

	mediaSpec := d[len("data:"):sep]
	// 检查 ;base64
	semiIdx := strings.LastIndex(mediaSpec, ";")
	if semiIdx < 0 {
		return nil, "", false
	}
	encPart := strings.TrimSpace(mediaSpec[semiIdx+1:])
	if !strings.EqualFold(encPart, "base64") {
		return nil, "", false
	}

	// 媒体类型 = mediaSpec 分号之前的部分
	mediaType := strings.TrimSpace(mediaSpec[:semiIdx])
	mediaType = strings.ToLower(mediaType)
	switch mediaType {
	case "image/png", "image/jpeg", "image/webp":
		// 有效类型
	default:
		return nil, "", false
	}

	// 确保 base64 数据无空白（连贯检查）
	base64Data := d[sep+1:]
	for _, b := range []byte(base64Data) {
		if b <= 0x20 || b == 0x7f {
			return nil, "", false
		}
		// 只接受 base64 合法字符
		if !((b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') || b == '+' || b == '/' || b == '=') {
			return nil, "", false
		}
	}

	decoded, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil || len(decoded) == 0 {
		return nil, "", false
	}
	if len(decoded) > launcherConfigIconMaxDecodedBeforeRescale {
		return nil, "", false
	}

	return decoded, mediaType, true
}

// ---- ensureLauncherConfigIconMigrationBackup ----

// ensureLauncherConfigIconMigrationBackup 确保配置图标迁移前备份文件存在。
// [S 汇编 0x1408904a0]：TrimSpace(path).bak 后缀 → ensureLauncherConfigMigrationBackupFile。
func ensureLauncherConfigIconMigrationBackup(path string) error {
	backupPath := strings.TrimSpace(path) + launcherConfigMigrationDirSuffix
	migBackupFmt := "历史配置图标迁移备份失败: %w"
	return ensureLauncherConfigMigrationBackupFile(path, backupPath, migBackupFmt)
}

// ---- ensureLauncherConfigIconLibraryMigrationBackup ----

// ensureLauncherConfigIconLibraryMigrationBackup 确保图标库迁移前备份文件存在。
// [S 汇编 0x140890560]：TrimSpace(path) + .bak 后缀 → ensureLauncherConfigMigrationBackupFile。
func ensureLauncherConfigIconLibraryMigrationBackup(path string) error {
	backupPath := strings.TrimSpace(path) + launcherConfigLibraryMigrationDirSuffix
	libBackupFmt := "独立图标库迁移备份失败: %w"
	return ensureLauncherConfigMigrationBackupFile(path, backupPath, libBackupFmt)
}

// ---- ensureLauncherConfigMigrationBackupFile ----

// ensureLauncherConfigMigrationBackupFile 备份迁移前文件。
// [S 汇编 0x140890620, 帧 0xc8]：
// path == backupPath → err("当前配置路径不能为空", 30B)；
// !os.Lstat(backupPath) (exist) + errors.Is(ENOENT) → 已存在且非普通文件 → err fmt；
// else os.Lstat(path) → errors.Is(ENOENT) → ok(nil)；
// os.MkdirAll(dir, 0755)；
// os.OpenFile(f, O_RDWR|O_CREATE|O_EXCL, 0666) → errors.Is(EEXIST) → os.Lstat → 非普通文件 → err；
// io.CopyBuffer(f, portableReader) → os.File.Sync → os.File.Close。
func ensureLauncherConfigMigrationBackupFile(path, backupPath, errFmt string) error {
	if path == "" || path == backupPath {
		return fmt.Errorf("当前配置路径不能为空")
	}

	// 若 backupPath 已存在且是普通文件 → 不需要备份
	if fi, err := os.Lstat(backupPath); err == nil {
		if fi.Mode().IsRegular() {
			return nil
		}
		return fmt.Errorf(errFmt, fmt.Errorf("历史图标迁移备份路径不是普通文件: %s", backupPath))
	}

	// 若原配置也不存在 → 无需备份
	if _, err := os.Lstat(path); err != nil {
		return nil // ENOENT → 配置也不存在
	}

	// 创建备份目录
	dir := filepath.Dir(backupPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf(errFmt, err)
	}

	// 创建备份文件（O_RDWR|O_CREATE|O_EXCL）
	f, err := os.OpenFile(backupPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			// 并发竞争：另一个进程已建
			fi, lerr := os.Lstat(backupPath)
			if lerr != nil {
				return fmt.Errorf(errFmt, lerr)
			}
			if !fi.Mode().IsRegular() {
				return fmt.Errorf(errFmt, fmt.Errorf("历史图标迁移备份路径不是普通文件: %s", backupPath))
			}
			return nil // 已存在且是普通文件 → ok
		}
		return fmt.Errorf(errFmt, err)
	}
	defer func() {
		if f != nil {
			f.Close()
		}
	}()

	// 读取原配置 -> 副本到 backupFile
	cfgFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf(errFmt, err)
	}
	defer cfgFile.Close()

	if _, err := io.Copy(f, cfgFile); err != nil {
		return fmt.Errorf(errFmt, err)
	}

	if err := f.Sync(); err != nil {
		return fmt.Errorf(errFmt, err)
	}

	// 提交成功
	f.Close()
	f = nil
	return nil
}
