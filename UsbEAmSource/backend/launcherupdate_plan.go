// AUTO-RECONSTRUCTED FUNCTIONS — DOMAIN: launcher update plan building / validation
// 研究用途
//
// 契约来源：
//   - 符号地址：symbols.main.bak（各函数地址见各自 [S] 标注）
//   - 行号蓝图：source_funcs.txt launcherupdate_plan.go L52-409
//
// 档位：[S] 汇编实证逐条追译 / [R] 标准模式还原
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ---- nonce ----

// newLauncherUpdateNonce 生成随机 nonce 的十六进制串。
// [S 汇编 0x1408c7da0]：分配 n 字节随机数 → crypto/rand.Read → hex 编码（2n 字符小写）。
// 参数 n 来自调用方（buildLauncherUpdatePlan 传 16：nonce 与 healthNonce 各一）。
func newLauncherUpdateNonce(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("生成随机 nonce 失败") // [S] err 路径 25B 静态消息
	}
	return hex.EncodeToString(buf), nil
}

// buildLauncherUpdatePlan 构建更新计划：安装 / 更新源 / 可执行文件的路径规范后，
// 生成双 nonce、收集文件清单、计算哈希、validate 后返回。
// [S 汇编 0x1408c7f40]：双 newLauncherUpdateNonce → TrimSpace+Clean+Abs 路径 →
// normalizeLauncherUpdatePackageRoot → collectLauncherUpdatePlanFiles →
// walk 结果转 plan.Files → validateLauncherUpdatePlan。
func buildLauncherUpdatePlan(installDir, updateRoot, executableName string) (*launcherUpdatePlan, error) {
	nonce, err := newLauncherUpdateNonce(16)
	if err != nil {
		return nil, err
	}
	healthNonce, err := newLauncherUpdateNonce(16)
	if err != nil {
		return nil, err
	}
	installDir = strings.TrimSpace(installDir)
	installDir = filepath.Clean(installDir)
	installDir, err = filepath.Abs(installDir)
	if err != nil {
		return nil, err
	}
	updateRoot = strings.TrimSpace(updateRoot)
	updateRoot = filepath.Clean(updateRoot)
	updateRoot, err = filepath.Abs(updateRoot)
	if err != nil {
		return nil, err
	}
	plan := &launcherUpdatePlan{
		SchemaVersion:  1,
		Nonce:          nonce,
		HealthNonce:    healthNonce,
		InstallDir:     installDir,
		UpdateRoot:     updateRoot,
		ExecutableName: executableName,
		PackageRoot:    normalizeLauncherUpdatePackageRoot(updateRoot),
		UpdateDir:      filepath.Join(installDir, ".usbeam-update-"+nonce[:8]),
		LogFile:        filepath.Join(installDir, ".update-log.txt"),
		JournalFile:    filepath.Join(installDir, ".update-journal.json"),
	}
	entries, err := collectLauncherUpdatePlanFiles(updateRoot)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		info, err := os.Stat(entry)
		if err != nil {
			return nil, err
		}
		h, err := hashLauncherUpdateFile(entry)
		if err != nil {
			return nil, err
		}
		rel := strings.TrimPrefix(entry, updateRoot)
		rel = strings.TrimPrefix(rel, `\`)
		rel = strings.TrimPrefix(rel, `/`)
		plan.Files = append(plan.Files, launcherUpdatePlanFile{
			Path:   rel,
			Size:   info.Size(),
			SHA256: h,
			Mode:   uint32(info.Mode().Perm()),
		})
	}
	plan.PackageSize = plan.computeTotalSize()
	plan.PackageSHA256 = "" // [S] 汇编中 PackageSHA256 字段留空
	if err := validateLauncherUpdatePlan(plan); err != nil {
		return nil, err
	}
	return plan, nil
}

// computeTotalSize 累加全部更新文件字节数。
// [S-sig]：签名实证；体为标准累加模式（[R]）。
func (plan *launcherUpdatePlan) computeTotalSize() int64 {
	var n int64
	for _, f := range plan.Files {
		n += f.Size
	}
	return n
}

// collectLauncherUpdatePlanFiles 递归收集 updateRoot 下全部文件（不含目录自身）。
// [S 汇编 0x1408c8700]：filepath.Walk → 跳过目录自身 → append 到结果切片。
func collectLauncherUpdatePlanFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

// hashLauncherUpdateFile 计算文件 SHA-256。
// [S] 真实语义在下方 0x1408c9000 实证版（303 行起）。

// marshalLauncherUpdatePlan JSON marshal 更新计划（缩进）。
// [S 汇编 0x1408c95a0]：json.MarshalIndent(plan, "", "  ")。
func marshalLauncherUpdatePlan(plan *launcherUpdatePlan) ([]byte, error) {
	return json.MarshalIndent(plan, "", "  ")
}

// unmarshalLauncherUpdatePlan JSON unmarshal 更新计划。
// [S 汇编 0x1408c96c0]：json.Unmarshal → 必要字段 TrimSpace 验证。
func unmarshalLauncherUpdatePlan(data []byte) (*launcherUpdatePlan, error) {
	var plan launcherUpdatePlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, err
	}
	if strings.TrimSpace(plan.Nonce) == "" {
		return nil, errors.New("launcherupdate: plan missing nonce")
	}
	return &plan, nil
}

// ---- validation ----

// validateLauncherUpdatePlan 校验更新计划全部字段合法性。
// [S 汇编 0x1408c9920]：逐字段非空/路径安全/数量边界检查。
func validateLauncherUpdatePlan(plan *launcherUpdatePlan) error {
	if plan == nil {
		return errors.New("launcherupdate: plan is nil")
	}
	if strings.TrimSpace(plan.Nonce) == "" {
		return errors.New("launcherupdate: plan nonce is empty")
	}
	if strings.TrimSpace(plan.InstallDir) == "" {
		return errors.New("launcherupdate: plan installDir is empty")
	}
	if strings.TrimSpace(plan.UpdateRoot) == "" {
		return errors.New("launcherupdate: plan updateRoot is empty")
	}
	if strings.TrimSpace(plan.ExecutableName) == "" {
		return errors.New("launcherupdate: plan executableName is empty")
	}
	if len(plan.Files) == 0 {
		return errors.New("launcherupdate: plan has no files")
	}
	seen := make(map[string]bool)
	for _, f := range plan.Files {
		if strings.TrimSpace(f.Path) == "" {
			return errors.New("launcherupdate: plan file path empty")
		}
		if seen[f.Path] {
			return fmt.Errorf("launcherupdate: duplicate file %q in plan", f.Path)
		}
		seen[f.Path] = true
		if f.Size < 0 {
			return fmt.Errorf("launcherupdate: negative size for %q", f.Path)
		}
		if f.SHA256 == "" {
			return fmt.Errorf("launcherupdate: missing hash for %q", f.Path)
		}
		if launcherUpdatePathProtected(f.Path) {
			return fmt.Errorf("launcherupdate: protected path %q in plan", f.Path)
		}
	}
	return nil
}

// validateLauncherUpdateNonce 验证 nonce：TrimSpace → hex 解码成功 → 解码字节数 == 期望长度。
// [S 汇编 0x1408ca680]：strings.TrimSpace → hex.DecodeString（任何大小写）→ 解码 len 与第二参数比较；
// 失败 → "nonce 编码或长度无效"（27B GBK）。
func validateLauncherUpdateNonce(nonce string, expectedLen int) error {
	decoded, err := hex.DecodeString(strings.TrimSpace(nonce))
	if err != nil || len(decoded) != expectedLen {
		return errors.New("nonce 编码或长度无效")
	}
	return nil
}

// validateLauncherUpdateAbsolutePath 校验更新目标绝对路径。
// [S 汇编 0x1408ca720]：
//
//	TrimSpace 为空或含 NUL → 错；反斜杠归一化 + 小写后依次检查保留前缀
//	"//"、"//?/"、"//./"、"/device/" → "不允许 UNC 或设备路径"；
//	filepath.Clean 后须 IsAbs 且与原串相等 → "路径必须是规范化绝对路径"；
//	卷名后路径含 ':'（ADS）→ "路径包含 ADS"；去前导分隔符后按 FieldsFunc 分段，
//	每段过 validateWindowsArchiveSegment；全部通过返回 nil。
func validateLauncherUpdateAbsolutePath(path string) error {
	p := strings.TrimSpace(path)
	if p == "" || strings.IndexRune(p, 0) >= 0 {
		return errors.New("路径为空")
	}
	lower := strings.ToLower(strings.ReplaceAll(p, "\\", "/"))
	if strings.HasPrefix(lower, "//") ||
		strings.HasPrefix(lower, "//?/") ||
		strings.HasPrefix(lower, "//./") ||
		strings.HasPrefix(lower, "/device/") {
		return errors.New("不允许 UNC 或设备路径")
	}
	c := filepath.Clean(p)
	if !filepath.IsAbs(c) || c != p {
		return errors.New("路径必须是规范化绝对路径")
	}
	rest := c[len(filepath.VolumeName(c)):]
	if strings.Contains(rest, ":") {
		return errors.New("路径包含 ADS")
	}
	rest = strings.TrimPrefix(rest, "\\")
	for _, seg := range strings.FieldsFunc(rest, func(r rune) bool {
		return r == '/' || r == '\\'
	}) {
		if err := validateWindowsArchiveSegment(seg); err != nil {
			return err
		}
	}
	return nil
}

// normalizeLauncherUpdateRelativePath 归一化相对路径并执行安全校验。
// [S 汇编 0x1408caac0]：TrimSpace(ReplaceAll(\→/)) 空/含 NUL/含 ':'/以 '/' 开头 → 错；
// 以 Windows 分隔符语义 Clean 后还原 '/'，结果须非 "."/".."/"../" 前缀且与原串相等 → 错；
// 分段数 ≤32 且 UTF-16 长度 ≤1024 → 错；每段 validateWindowsArchiveSegment；
// launcherUpdatePathProtected 命中 → 错。成功返回规范化路径。
func normalizeLauncherUpdateRelativePath(path string) (string, error) {
	p := strings.TrimSpace(strings.ReplaceAll(path, "\\", "/"))
	if p == "" || strings.IndexRune(p, 0) >= 0 || strings.Contains(p, ":") || strings.HasPrefix(p, "/") {
		return "", errors.New("相对路径为空或不安全")
	}
	c := filepath.Clean(strings.ReplaceAll(p, "/", "\\"))
	norm := strings.ReplaceAll(c, "\\", "/")
	if norm == "." || norm == ".." || strings.HasPrefix(norm, "../") || norm != p {
		return "", errors.New("相对路径未规范化或越界")
	}
	segs := strings.Split(norm, "/")
	if len(segs) > 0x20 {
		return "", errors.New("相对路径超过限制")
	}
	if utf16RuneLen(norm) > 0x400 {
		return "", errors.New("相对路径超过限制")
	}
	for _, seg := range segs {
		if err := validateWindowsArchiveSegment(seg); err != nil {
			return "", err
		}
	}
	if launcherUpdatePathProtected(norm) {
		return "", errors.New("路径指向受保护的用户数据")
	}
	return norm, nil
}

// utf16RuneLen 计算字符串 UTF-16 编码后的单元数（代理对计 2）。
// [S] normalizeLauncherUpdateRelativePath 中 utf16.Encode 结果长度比较用。
func utf16RuneLen(s string) int {
	n := 0
	for _, r := range s {
		if r > 0xffff {
			n += 2
		} else {
			n++
		}
	}
	return n
}

// launcherUpdatePathProtected 判断路径是否指向受保护的用户数据。
// [S 汇编 0x1408cae60]：反斜杠→正斜杠、TrimSpace、ToLower；Split("/") 后首段若为
// data/backup/plugin/update/profile_backup → 保护；Base 名若为
// usbeam_launcher_config / .usbeam-launcher-home → 保护；首段以 ".usbeam-update-" 开头 → 保护。
func launcherUpdatePathProtected(path string) bool {
	normalized := strings.ToLower(strings.TrimSpace(strings.ReplaceAll(path, "\\", "/")))
	segs := strings.Split(normalized, "/")
	if len(segs) == 0 {
		return false
	}
	switch segs[0] {
	case "data", "backup", "plugin", "update", "profile_backup":
		return true
	}
	base := strings.ToLower(filepath.Base(strings.ReplaceAll(normalized, "/", "\\")))
	if base == "usbeam_launcher_config" || base == ".usbeam-launcher-home" {
		return true
	}
	return len(segs[0]) >= 15 && segs[0][:15] == ".usbeam-update-"
}

// samePathFold 大小写不敏感的同路径比较。
// [S 汇编 0x1408cb060]：两侧 TrimSpace → filepath.Clean → filepath.Abs；
// 任一侧 Abs 失败 → false；否则 strings.EqualFold。
func samePathFold(a, b string) bool {
	pa, err1 := filepath.Abs(filepath.Clean(strings.TrimSpace(a)))
	pb, err2 := filepath.Abs(filepath.Clean(strings.TrimSpace(b)))
	if err1 != nil || err2 != nil {
		return false
	}
	return strings.EqualFold(pa, pb)
}

// normalizeLauncherUpdatePackageRoot 规范更新包根目录名。
// [S 汇编 0x1408b6b60]：strings.Replace(root,"\\","/",-1) → TrimSpace → 空/含"/"/"."/".." 返回空。
// 常量实证：old=0x1411cac40(0x5c "\\")，new=needle=0x140c3362f(0x2f "/")。
func normalizeLauncherUpdatePackageRoot(root string) string {
	root = strings.Replace(root, "\\", "/", -1)
	root = strings.TrimSpace(root)
	if root == "" || strings.Contains(root, "/") || root == "." || root == ".." {
		return ""
	}
	return root
}

// hashLauncherUpdateFile 计算文件的 SHA-256 十六进制串。
// [S 汇编 0x1408c9000]：os.Open → defer Close → Stat：非常规文件（Dir/Symlink/Device/
// NamedPipe/Socket/CharDevice/Irregular）→ "更新文件不是普通文件"；
// sha256 hash ← io.Copy；复制字节数 != Stat.Size → "更新文件读取长度发生变化"；
// hex 编码返回。
func hashLauncherUpdateFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return "", err
	}
	if fi.Mode()&(os.ModeDir|os.ModeSymlink|os.ModeDevice|os.ModeNamedPipe|os.ModeSocket|os.ModeCharDevice|os.ModeIrregular) != 0 {
		return "", errors.New("更新文件不是普通文件")
	}
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", err
	}
	if n != fi.Size() {
		return "", errors.New("更新文件读取长度发生变化")
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// compareLauncherUpdatePlanFiles 比较两个更新计划文件列表（依路径排序后逐项对比）。
// [S 汇编 0x1408cb120]：排序两份 Files → 逐项 Path/Size/SHA256 对比 → 报告差异。
func compareLauncherUpdatePlanFiles(a, b *launcherUpdatePlan) []string {
	var diffs []string
	am := make(map[string]launcherUpdatePlanFile)
	for _, f := range a.Files {
		am[f.Path] = f
	}
	bm := make(map[string]launcherUpdatePlanFile)
	for _, f := range b.Files {
		bm[f.Path] = f
	}
	for path, af := range am {
		bf, ok := bm[path]
		if !ok {
			diffs = append(diffs, fmt.Sprintf("only in first: %s", path))
			continue
		}
		if af.Size != bf.Size {
			diffs = append(diffs, fmt.Sprintf("size mismatch %s: %d vs %d", path, af.Size, bf.Size))
		}
		if af.SHA256 != bf.SHA256 {
			diffs = append(diffs, fmt.Sprintf("hash mismatch %s: %s vs %s", path, af.SHA256, bf.SHA256))
		}
	}
	for path := range bm {
		if _, ok := am[path]; !ok {
			diffs = append(diffs, fmt.Sprintf("only in second: %s", path))
		}
	}
	sort.Strings(diffs)
	return diffs
}
