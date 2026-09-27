// AUTO-RECONSTRUCTED — DOMAIN: ZIP archive extraction
// 研究用途。反汇编实证全部 11 个函数 + 3 个 deferwrap。
// 已解码 .rdata 错误串与 Windows 保留名检查表。
package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// ---- helpers ----

// hasReparsePoint 检查路径是否含 reparse point（符号链接/交接点）。
// 现在委托给 launcherUpdatePathHasReparsePoint（GetFileAttributes + FILE_ATTRIBUTE_REPARSE_POINT）。
// [S-inline 内联]：委托 launcherUpdatePathHasReparsePoint，error 时返 false。
func hasReparsePoint(p string) bool {
	b, err := launcherUpdatePathHasReparsePoint(p)
	if err != nil {
		return false
	}
	return b
}

// newExtractError 创建解压错误（对应反汇编 runtime.newobject + .rdata 字面量）。
// [S 汇编实证]：runtime.newobject + .rdata 字面量构造 extractError{msg}。
func newExtractError(msg string) error {
	return &extractError{msg: msg}
}

type extractError struct {
	msg string
}

// Error 返回解压错误消息。[S-inline 内联，语义确定]
func (e *extractError) Error() string { return e.msg }

// matchSplitArchiveExtension 是否以 .r##/.z## 结尾（分卷 RAR/ZIP 标记）。
// [S] 0x140817760, 128B：len>=5, byte[-4]=='.', byte[-3]→(r|z), byte[-2]/[-1]均为数字。
func matchSplitArchiveExtension(p string) bool {
	if len(p) < 5 {
		return false
	}
	if p[len(p)-4] != '.' {
		return false
	}
	c := p[len(p)-3]
	if c >= 'A' && c <= 'Z' {
		c += 0x20
	}
	if c != 'r' && c != 'z' {
		return false
	}
	if p[len(p)-2] < '0' || p[len(p)-2] > '9' {
		return false
	}
	if p[len(p)-1] < '0' || p[len(p)-1] > '9' {
		return false
	}
	return true
}

// ---- path safety ----

// isPathInsideDirectory 检查 p 是否在 dir 内（不可逃逸到父目录）。
// [S] 0x140750be0, 320B：Clean+Abs→Rel→检查 "." / ".." / 绝对路径。
func isPathInsideDirectory(p, dir string) bool {
	p, err := filepath.Abs(filepath.Clean(p))
	if err != nil {
		return false
	}
	dir, err = filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(dir, p)
	if err != nil {
		return false
	}
	if len(rel) == 1 && rel[0] == '.' {
		return true
	}
	if len(rel) == 2 && rel[0] == '.' && rel[1] == '.' {
		return false
	}
	if filepath.IsAbs(rel) {
		return false
	}
	if len(rel) >= 3 && rel[:3] == ".."+string(filepath.Separator) {
		return false
	}
	return true
}

// ---- remove (tree) ----

// removeArchiveExtractionRoot 删除解压根路径。
// [S] 0x140750900, 192B：Lstat→ModeDir→hasReparse→removeArchiveTreeNoFollow/os.Remove。
func removeArchiveExtractionRoot(root string) error {
	fi, err := os.Lstat(root)
	if err != nil {
		return err
	}
	if fi.Mode().IsDir() && !hasReparsePoint(root) {
		return removeArchiveTreeNoFollow(root)
	}
	return os.Remove(root)
}

// removeArchiveTreeNoFollow 递归删除目录树，不跟踪符号链接。
// [S] 0x1407509c0, 544B：Lstat→errors.Is(ErrNotExist)→ModeDir→ReadDir→递归→最终 Remove。
func removeArchiveTreeNoFollow(root string) error {
	fi, err := os.Lstat(root)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	if !fi.Mode().IsDir() {
		return os.Remove(root)
	}
	if hasReparsePoint(root) {
		return os.Remove(root)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		childPath := filepath.Join(root, entry.Name())
		if err := removeArchiveTreeNoFollow(childPath); err != nil {
			return err
		}
	}
	return os.Remove(root)
}

// ---- create directory ----

// createArchiveDirectory 递归创建相对路径目录，做 zip-slip 与 reparse 安全检查。
// [S] 0x1407502c0, 768B。
func createArchiveDirectory(relative, targetDir string) error {
	fullPath := filepath.Join(targetDir, relative)
	rel, err := filepath.Rel(targetDir, fullPath)
	if err != nil {
		return err
	}
	if len(rel) == 1 && rel[0] == '.' {
		return nil
	}

	parts := strings.Split(rel, string(filepath.Separator))
	cur := targetDir
	for _, part := range parts {
		subPath := filepath.Join(cur, part)
		fi, err := os.Lstat(subPath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				if mkErr := os.Mkdir(subPath, 0o700); mkErr != nil {
					return mkErr
				}
				cur = subPath
				continue
			}
			return err
		}
		if !fi.Mode().IsDir() {
			return newExtractError("解压目录被非目录或链接占用")
		}
		if hasReparsePoint(subPath) {
			return newExtractError("解压目录包含 reparse point")
		}
		cur = subPath
	}
	return nil
}

// ---- ensure parent safe ----

// ensureArchiveParentPathSafe 验证父路径不含符号链接或 reparse point。
// [S] 0x1407505c0, 832B：Abs→VolName→strings.Split→逐段 Lstat→检查/跳过。
func ensureArchiveParentPathSafe(parentPath string) error {
	absPath, err := filepath.Abs(parentPath)
	if err != nil {
		return err
	}

	// 分解路径逐段检查
	vol := filepath.VolumeName(absPath)
	suffix := absPath[len(vol):]
	suffix = strings.TrimPrefix(suffix, string(filepath.Separator))
	if suffix == "" {
		return nil
	}

	parts := strings.Split(suffix, string(filepath.Separator))
	cur := vol + string(filepath.Separator)
	for _, part := range parts {
		if part == "" {
			continue
		}
		subPath := filepath.Join(cur, part)
		fi, err := os.Lstat(subPath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				cur = subPath
				continue
			}
			return err
		}
		if !fi.Mode().IsDir() {
			return newExtractError("解压父路径包含链接或非目录")
		}
		if hasReparsePoint(subPath) {
			return newExtractError("解压父路径包含 reparse point")
		}
		cur = subPath
	}
	return nil
}

// ---- Windows reserved name validation ----

var windowsReservedNames = map[string]bool{
	"con": true, "nul": true, "prn": true, "aux": true,
	"com1": true, "com2": true, "com3": true, "com4": true,
	"com5": true, "com6": true, "com7": true, "com8": true, "com9": true,
	"lpt1": true, "lpt2": true, "lpt3": true, "lpt4": true,
	"lpt5": true, "lpt6": true, "lpt7": true, "lpt8": true, "lpt9": true,
	"clock$": true, "con$": true, "conout$": true,
}

// isWindowsReservedName 检查名称是否为 Windows 保留名（含 superscript 变体 .rdata 实测）。
// [S] 反汇编 0x14074f9b5-0x14074fa00
func isWindowsReservedName(name string) bool {
	if windowsReservedNames[name] {
		return true
	}
	// 检查 UTF-8 superscript 变体 com²/com³/com¹ lpt²/lpt³/lpt¹
	if len(name) == 5 {
		prefix := name[:3]
		rest := name[3:]
		if (prefix == "com" || prefix == "lpt") && len(rest) == 2 &&
			rest[0] == 0xc2 && (rest[1] == 0xb2 || rest[1] == 0xb3 || rest[1] == 0xb9) {
			return true
		}
	}
	return false
}

// validateWindowsArchiveSegment 检查单个路径段是否合法。
// [S] 0x14074f500, 1408B：空/./..→空格/非法字→尾随点/空格→保留名。
func validateWindowsArchiveSegment(segment string) error {
	if len(segment) == 0 || segment == "." || segment == ".." {
		return newExtractError("目录段为空、保留或带尾随点/空格")
	}

	// 非法字符：<>:"/\|?* 和控制字符
	const invalidChars = `<>:"/\|?*`
	for i := 0; i < len(segment); {
		r := rune(segment[i])
		if r < 0x20 {
			return newExtractError("目录段包含 Windows 非法字符")
		}
		if r < 0x80 {
			if strings.IndexRune(invalidChars, r) >= 0 {
				return newExtractError("目录段包含 Windows 非法字符")
			}
			i++
		} else {
			_, size := utf8.DecodeRuneInString(segment[i:])
			i += size
		}
	}

	// 尾随点或空格
	if last := segment[len(segment)-1]; last == '.' || last == ' ' {
		return newExtractError("目录段为空、保留或带尾随点/空格")
	}

	// 检查段名和扩展名是否为 Windows 保留名
	lower := strings.ToLower(segment)
	if isWindowsReservedName(lower) {
		return newExtractError("目录段使用 Windows 保留名称")
	}
	// 去掉 . 之后的部分检查扩展名
	if extStart := strings.LastIndexByte(lower, '.'); extStart >= 0 {
		ext := lower[extStart+1:]
		if isWindowsReservedName(ext) {
			return newExtractError("目录段使用 Windows 保留名称")
		}
	}

	return nil
}

// ---- validate zip entry name ----

// validateZipEntryName 验证 ZIP 条目名安全且符合 Windows 约束。
// [S] 0x14074ee60, 1696B：\→/→前导/→..→path.Clean→split→validateWindowsArchiveSegment→UTF-16 长度。
func validateZipEntryName(name string, allowDir bool) (string, error) {
	if !utf8.ValidString(name) {
		return "", newExtractError("ZIP 条目名不是有效 UTF-8")
	}

	// 检查空名
	if len(name) == 0 {
		return "", newExtractError("ZIP 条目名空")
	}

	// \ → /
	cleaned := strings.Replace(name, "\\", "/", -1)

	// 前导 /
	if cleaned[0] == '/' {
		return "", fmt.Errorf("ZIP 条目名以 / 开头: %s", name)
	}

	// 检查 //
	if strings.Contains(cleaned, "//") {
		return "", fmt.Errorf("ZIP 条目名含 //: %s", name)
	}

	// path.Clean→路径越界检查
	cleanPath := path.Clean(cleaned)
	if cleanPath == "." || cleanPath == ".." || strings.HasPrefix(cleanPath, "../") {
		return "", fmt.Errorf("ZIP 条目名路径越界: %s", name)
	}

	// 检查各段
	segments := strings.Split(cleaned, "/")
	for _, seg := range segments {
		if seg == "" {
			continue
		}
		if err := validateWindowsArchiveSegment(seg); err != nil {
			return "", fmt.Errorf("ZIP 条目名无效: %s: %w", name, err)
		}
	}

	// UTF-16 长度检查
	if utf8.RuneCountInString(cleaned) >= 256 {
		return "", fmt.Errorf("ZIP 条目名 UTF-16 长度超限: %s", name)
	}

	return cleaned, nil
}

// ---- extractArchiveRegularFile ----

// extractArchiveRegularFile 解压单个 ZIP 条目为普通文件。
// [S] 0x14074fa80, 1920B：Dir→ensureArchiveParentPathSafe→Open→openArchiveRegularFileNoFollow→CopyN→大小校验→Sync。
func extractArchiveRegularFile(file *zip.File, targetPath string, targetDir string, buf []byte) error {
	parentDir := filepath.Dir(targetPath)
	if err := ensureArchiveParentPathSafe(parentDir); err != nil {
		return err
	}

	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	f, err := openArchiveRegularFileNoFollow(targetPath, 0o644)
	if err != nil {
		return err
	}
	fileClosed := false
	defer func() {
		if !fileClosed && f != nil {
			f.Close()
			os.Remove(targetPath)
		}
	}()

	written, err := io.CopyBuffer(f, rc, buf)
	if err != nil && err != io.EOF {
		// 写入出错
		fileClosed = true
		f.Close()
		os.Remove(targetPath)
		return fmt.Errorf("解压文件实际大小超过限制: %w", err)
	}

	if uint64(written) != file.UncompressedSize64 {
		fileClosed = true
		f.Close()
		os.Remove(targetPath)
		return fmt.Errorf("压缩包文件实际大小与目录记录不一致: %s", file.Name)
	}

	if err := f.Sync(); err != nil {
		fileClosed = true
		f.Close()
		os.Remove(targetPath)
		return err
	}

	fileClosed = true
	return f.Close()
}

// ---- prepareArchiveEntries ----

// prepareArchiveEntries 验证 ZIP 条目并准备解压列表。
// [S] 0x14074e2e0, 2944B：Clean→Abs→IsAbs→makeslice→遍历→validateZipEntryName。
func prepareArchiveEntries(entries []*zip.File, targetDir string) ([]preparedArchiveEntry, error) {
	if len(entries) == 0 {
		return nil, newExtractError("ZIP 文件内无条目")
	}

	absDir, err := filepath.Abs(filepath.Clean(targetDir))
	if err != nil {
		return nil, err
	}
	if !filepath.IsAbs(absDir) {
		return nil, newExtractError("解压目标必须是绝对路径")
	}

	result := make([]preparedArchiveEntry, 0, len(entries))
	for _, f := range entries {
		validName, err := validateZipEntryName(f.Name, f.FileInfo().IsDir())
		if err != nil {
			return nil, err
		}
		result = append(result, preparedArchiveEntry{
			file:       f,
			relative:   validName,
			targetPath: filepath.Join(absDir, validName),
			directory:  f.FileInfo().IsDir(),
		})
	}
	return result, nil
}

// ---- extractZipArchive ----

// extractZipArchive 解压 ZIP 到 targetDir。
// [S] 0x14074d8a0, 2464B：OpenFile→Stat→NewReader→prepare→createDir→逐条 extract/createDirectory。
func extractZipArchive(path, targetDir string) error {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if fi.Size() <= 0 || fi.Mode().IsDir() {
		return newExtractError("压缩包大小或文件类型无效")
	}

	zr, err := zip.NewReader(f, fi.Size())
	if err != nil {
		return err
	}
	entries := zr.File

	prepared, err := prepareArchiveEntries(entries, targetDir)
	if err != nil {
		return err
	}

	// 确保目标目录存在
	tdFi, err := os.Lstat(targetDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		parentDir := filepath.Dir(targetDir)
		if err := ensureArchiveParentPathSafe(parentDir); err != nil {
			return err
		}
		if err := os.Mkdir(targetDir, 0o700); err != nil {
			return err
		}
	} else if !tdFi.IsDir() {
		return newExtractError("解压目标不是普通文件")
	}

	buf := make([]byte, 32*1024)
	for _, entry := range prepared {
		if entry.directory {
			if err := createArchiveDirectory(entry.relative, targetDir); err != nil {
				return err
			}
		} else {
			if err := extractArchiveRegularFile(entry.file, entry.targetPath, targetDir, buf); err != nil {
				return err
			}
		}
	}
	return nil
}
