package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// validatePluginID 把底层插件 ID 校验错误翻译为用户可读的中文消息。
// [S 汇编 0x140923be0, 672B] 实证：Validate 通过→nil；errors.As 取 *ValidationError；
// Reason 1/3/4/5/6/默认走 errors.New，Reason 2/9 走 fmt.Errorf，其余透传原错误。
func validatePluginID(id string) error {
	if err := Validate(id); err != nil {
		var ve *ValidationError
		if !errors.As(err, &ve) {
			return err
		}
		switch ve.Reason {
		case 1:
			return errors.New("插件标识不能为空")
		case 2:
			return fmt.Errorf("插件标识长度不能超过 %d 字节", 64)
		case 3:
			return errors.New("插件标识只能使用 ASCII 字符")
		case 4:
			return errors.New("插件标识不能包含控制字符")
		case 5:
			return errors.New("插件标识包含不允许的路径字符")
		case 6:
			return errors.New("插件标识不能以点或空格结尾")
		case 9:
			return fmt.Errorf("插件标识不能使用 Windows 保留名称 %s", ve.ReservedName)
		default:
			return errors.New("插件标识只能包含字母、数字、点、下划线和连字符，且必须以字母或数字开头和结尾")
		}
	}
	return nil
}

// normalizeValidatedPluginID 先去除首尾空白，再校验；失败返回空串，成功返回规整后的 ID。
// [S 汇编 0x140923e80, 128B] 实证：strings.TrimSpace→validatePluginID；err!=nil→("",err)，否则 (trimmed,nil)。
func normalizeValidatedPluginID(id string) (string, error) {
	s := strings.TrimSpace(id)
	if err := validatePluginID(s); err != nil {
		return "", err
	}
	return s, nil
}

// readPluginFileBounded 以只读方式读取 path，字节数上限 maxBytes。
// [S 汇编 0x140923f00, 1312B] 实证：maxBytes<=0→错误；os.OpenFile(O_RDONLY,0)；
// Stat 后要求普通文件；Size<0 或 >maxBytes→错误；io.ReadAll(LimitReader(f,maxBytes+1)) 再复核长度。
func readPluginFileBounded(path string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, errors.New("插件文件读取上限无效")
	}
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("插件资源必须是普通文件")
	}
	if info.Size() < 0 || info.Size() > maxBytes {
		return nil, fmt.Errorf("插件资源超过 %d 字节上限", maxBytes)
	}
	data, err := io.ReadAll(io.LimitReader(f, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("插件资源超过 %d 字节上限", maxBytes)
	}
	return data, nil
}

// readPluginManifestBounded 读取并解码插件清单 JSON，校验 ID 与尾随内容。
// [S 汇编 0x140924480, 928B] 实证：readPluginFileBounded(path,0x40000)；
// json.NewDecoder(strings.NewReader(string(data)))；Decode(&m) 后第二次 Decode 必须为 io.EOF；
// 否则 nil→多值错误，非 EOF→fmt.Errorf 尾随错误；最后 validatePluginID(m.ID)。
func readPluginManifestBounded(path string) (PluginManifest, error) {
	data, err := readPluginFileBounded(path, 0x40000)
	if err != nil {
		return PluginManifest{}, err
	}
	var m PluginManifest
	dec := json.NewDecoder(strings.NewReader(string(data)))
	if err := dec.Decode(&m); err != nil {
		return PluginManifest{}, err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return PluginManifest{}, errors.New("插件清单只能包含一个 JSON 值")
		}
		return PluginManifest{}, fmt.Errorf("插件清单存在尾随内容: %w", err)
	}
	if err := validatePluginID(m.ID); err != nil {
		return PluginManifest{}, err
	}
	return m, nil
}

// isPathWithinPluginRoot 判断 path 是否落在 root 内（不逃逸到上级目录）。
// [S 汇编 0x140924820, 192B] 实证：filepath.Rel(root,path)；err!=nil 或 rel==".." 或 IsAbs(rel)
// 或 strings.HasPrefix(rel,"..\\") 均判 false。
func isPathWithinPluginRoot(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	if rel == ".." {
		return false
	}
	if filepath.IsAbs(rel) {
		return false
	}
	if strings.HasPrefix(rel, "..\\") {
		return false
	}
	return true
}

// resolvePluginPathSecurely 把插件内相对资源路径解析为可信的绝对路径。
// [S 汇编 0x1409248e0, 1792B] 实证：path/root 去空白，root 反斜杠先转正斜杠做绝对性检查，
// 再转回反斜杠 Clean；拒绝绝对 root、"."/".."、上级逃逸；Join 后逐段 Lstat 拒绝 symlink/reparse；
// 最后 EvalSymlinks 后复核在 root 内，并拒绝目录（allowDir=false）与非规则文件。
func resolvePluginPathSecurely(path, root string, allowDir bool) (string, error) {
	trimmedPath := strings.TrimSpace(path)
	trimmedRoot := strings.TrimSpace(strings.Replace(root, "\\", "/", -1))
	if trimmedPath == "" || trimmedRoot == "" {
		return "", errors.New("插件资源路径无效")
	}
	if strings.IndexByte(trimmedRoot, 0) >= 0 {
		return "", errors.New("插件资源路径无效")
	}
	if filepath.IsAbs(trimmedRoot) || filepath.VolumeName(trimmedRoot) != "" || trimmedRoot[0] == '/' {
		return "", errors.New("插件资源必须使用相对路径")
	}
	cleanedRoot := filepath.Clean(strings.ReplaceAll(trimmedRoot, "/", "\\"))
	if cleanedRoot == "." || cleanedRoot == ".." || strings.HasPrefix(cleanedRoot, "..\\") {
		return "", errors.New("不允许访问插件目录外的文件")
	}
	absPath, err := filepath.Abs(trimmedPath)
	if err != nil {
		return "", err
	}
	absClean := filepath.Clean(absPath)
	joined := filepath.Join(absClean, cleanedRoot)
	if !isPathWithinPluginRoot(absClean, joined) {
		return "", errors.New("不允许访问插件目录外的文件")
	}
	base := absClean
	for _, part := range strings.Split(cleanedRoot, "\\") {
		if part == "" || part == "." {
			continue
		}
		current := filepath.Join(base, part)
		info, err := os.Lstat(current)
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 || pluginPathHasReparsePoint(current) {
			return "", errors.New("插件资源路径不能包含符号链接或 reparse point")
		}
		base = current
	}
	absResolved, err := filepath.EvalSymlinks(absClean)
	if err != nil {
		return "", err
	}
	joinedResolved, err := filepath.EvalSymlinks(joined)
	if err != nil {
		return "", err
	}
	cleanAbs := filepath.Clean(absResolved)
	cleanJoined := filepath.Clean(joinedResolved)
	if !isPathWithinPluginRoot(cleanAbs, cleanJoined) {
		return "", errors.New("插件资源最终路径越界")
	}
	info, err := os.Stat(joinedResolved)
	if err != nil {
		return "", err
	}
	if info.IsDir() && !allowDir {
		return "", errors.New("插件资源不能是目录")
	}
	if !info.IsDir() && !info.Mode().IsRegular() {
		return "", errors.New("插件资源必须是普通文件")
	}
	return filepath.Clean(joinedResolved), nil
}
