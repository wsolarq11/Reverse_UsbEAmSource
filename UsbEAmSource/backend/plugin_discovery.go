// AUTO-RECONSTRUCTED PLUGIN DISCOVERY — 插件发现域（批次 194 落地）
// 研究用途
//
// 档位：[S] 全汇编实证（discoverPluginsUnlocked 依赖链）。
//
//	normalizePluginIconFile (0x14092bb20)、resolvePluginLocalIconURL (0x14092c040)、
//	mergeManifestLocalizations (0x1407a6380)、loadPluginLocalizations (0x1407a5ec0)、
//	parseLanguageContent (0x1407a3bc0)、parseLanguageScanner (0x1407a3e00)、
//	decodeLanguageValue (0x1407a4760)、setNestedLanguageMessage (0x1407a4940)
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// normalizePluginIconFile 规范化插件图标文件名（去空白、正斜杠化、拒绝目录与逃逸、限 svg/png）。
// [S 汇编 0x14092bb20, 384B]：TrimSpace 空则 "" → "\"→"/" → 含 "/" 则 "" →
// path.Clean(icon)!=icon 则 "" → 扩展名 ToLower 仅 ".svg"/".png" 通过。
func normalizePluginIconFile(icon string) string {
	icon = strings.TrimSpace(icon)
	if icon == "" {
		return ""
	}
	icon = strings.ReplaceAll(icon, "\\", "/")
	if strings.Index(icon, "/") >= 0 {
		return ""
	}
	if path.Clean(icon) != icon {
		return ""
	}
	ext := strings.ToLower(path.Ext(icon))
	if ext != ".svg" && ext != ".png" {
		return ""
	}
	return icon
}

// resolvePluginLocalIconURL 解析插件本地图标 URL（相对 host 路径）。
// [S 汇编 0x14092c040, 480B]：normalizePluginIconFile → SourceDir/Icon 空则 "" →
// filepath.Join(SourceDir, icon) Stat 出错或目录则 "" → fmt.Sprintf("/plugin-host/%s/%s",
// url.PathEscape(IconURL), url.PathEscape(icon))。
func resolvePluginLocalIconURL(m PluginManifest) string {
	icon := normalizePluginIconFile(m.Icon)
	srcDir := strings.TrimSpace(m.SourceDir)
	if icon == "" || srcDir == "" {
		return ""
	}
	iconURL := strings.TrimSpace(m.IconURL)
	if iconURL == "" {
		return ""
	}
	info, err := os.Stat(filepath.Join(srcDir, icon))
	if err != nil || info.IsDir() {
		return ""
	}
	return fmt.Sprintf("/plugin-host/%s/%s", url.PathEscape(iconURL), url.PathEscape(icon))
}

// mergeManifestLocalizations 合并清单语言条目（key 去空白、空键兜底 "en-US"、Messages 空则初始化）。
// [S 汇编 0x1407a6380, 384B]：遍历 src map → TrimSpace(key) 空则 "en-US" →
// Messages==nil 则 make → dst[key]=loc。
func mergeManifestLocalizations(dst map[string]PluginLocalization, src map[string]PluginLocalization) {
	for key, loc := range src {
		key = strings.TrimSpace(key)
		if key == "" {
			key = "en-US"
		}
		if loc.Messages == nil {
			loc.Messages = make(map[string]interface{})
		}
		dst[key] = loc
	}
}

// loadPluginLocalizations 从插件 "language" 目录解析 .ini 语言文件，回填 I18N。
// [S 汇编 0x1407a5ec0, 1728B]：mergeManifestLocalizations(result, m.I18N) →
// resolvePluginPathSecurely(SourceDir,"language",true) 失败返回 → ReadDir 失败返回 →
// 逐项：目录跳过、非 .ini（EqualFold 扩展名）跳过 → resolve 路径 + readPluginFileBounded(0x80000) →
// parseLanguageContent → lang 空跳过 → result[lang] 以 nestedLanguageMessageString 覆盖
// Name/Description（"plugin.name"/"plugin.description"）→ mergeLanguageMessages → 写回。
func loadPluginLocalizations(m PluginManifest) map[string]PluginLocalization {
	result := make(map[string]PluginLocalization)
	mergeManifestLocalizations(result, m.I18N)
	langDir, err := resolvePluginPathSecurely(strings.TrimSpace(m.SourceDir), "language", true)
	if err != nil {
		return result
	}
	entries, err := os.ReadDir(langDir)
	if err != nil {
		return result
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.EqualFold(path.Ext(name), ".ini") {
			continue
		}
		filePath, err := resolvePluginPathSecurely(m.SourceDir, filepath.Join("language", name), false)
		if err != nil {
			continue
		}
		data, err := readPluginFileBounded(filePath, 0x80000)
		if err != nil {
			continue
		}
		lang, _, _, _, messages, err := parseLanguageContent(data, filePath)
		if err != nil {
			continue
		}
		if strings.TrimSpace(lang) == "" {
			continue
		}
		loc := result[lang]
		if s := nestedLanguageMessageString(messages, "plugin.name"); s != "" {
			loc.Name = s
		}
		if s := nestedLanguageMessageString(messages, "plugin.description"); s != "" {
			loc.Description = s
		}
		if loc.Messages == nil {
			loc.Messages = make(map[string]interface{})
		}
		mergeLanguageMessages(loc.Messages, messages)
		result[lang] = loc
	}
	return result
}

// parseLanguageContent 解析插件语言文件（薄封装：构造 bufio.Scanner 后交给 parseLanguageScanner）。
// [S 汇编 0x1407a3bc0, 576B]：bufio.NewScanner(bytes.NewReader(data))（split=ScanLines,
// maxTokenSize=64KB）→ parseLanguageScanner(scanner, path) 透传返回。
func parseLanguageContent(data []byte, path string) (lang, name, nativeName, reserved string, messages map[string]interface{}, err error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	return parseLanguageScanner(scanner, path)
}

// parseLanguageScanner 逐行扫描 .ini 语言文件，提取 [meta] 元数据与嵌套消息。
// [S 汇编 0x1407a3e00, 2334B]：lang 默认 = TrimSuffix(Base(path), Ext) → Scan 循环：
// 去 BOM/TrimSpace → 空行与 ";" "#" 注释跳过 → "[x]" 段落行更新 section →
// Cut(line,"=") 无 "=" 跳过 → key/value TrimSpace → decodeLanguageValue(value) →
// EqualFold(section,"meta") 则按 key（ToLower）分派 code/name/native_name →
// section 非空则 setNestedLanguageMessage(messages, section+"."+key, value)。
func parseLanguageScanner(scanner *bufio.Scanner, path string) (lang, name, nativeName, reserved string, messages map[string]interface{}, err error) {
	base := filepath.Base(path)
	lang = strings.TrimSuffix(base, filepath.Ext(base))
	messages = make(map[string]interface{})
	var section string
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimPrefix(line, "\ufeff")
		line = strings.TrimSpace(line)
		if line == "" || line[0] == ';' || line[0] == '#' {
			continue
		}
		if line[0] == '[' && line[len(line)-1] == ']' {
			section = strings.TrimSpace(line[1 : len(line)-1])
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = decodeLanguageValue(strings.TrimSpace(value))
		if strings.EqualFold(section, "meta") {
			switch strings.ToLower(key) {
			case "code":
				lang = value
			case "name":
				name = value
			case "native_name":
				nativeName = value
			}
		}
		if section != "" {
			setNestedLanguageMessage(messages, section+"."+key, value)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", "", "", nil, err
	}
	return lang, name, nativeName, reserved, messages, nil
}

// decodeLanguageValue 解码语言值中的反斜杠转义（\n \r \t \\ 及未知转义原样保留）。
// [S 汇编 0x1407a4760, 480B]：Index(value,"\\")<0 则原样返回 → Builder 逐 rune：
// 遇 "\" 置转义态，次字符按 n/r/t/\\ 解码，未知转义补 "\" 保留原字符。
func decodeLanguageValue(value string) string {
	if strings.Index(value, "\\") < 0 {
		return value
	}
	var b strings.Builder
	b.Grow(len(value))
	escaped := false
	for i := 0; i < len(value); {
		r, size := utf8.DecodeRuneInString(value[i:])
		i += size
		if !escaped {
			if r == '\\' {
				escaped = true
			} else {
				b.WriteRune(r)
			}
			continue
		}
		escaped = false
		switch r {
		case 'n':
			b.WriteRune('\n')
		case 'r':
			b.WriteRune('\r')
		case 't':
			b.WriteRune('\t')
		case '\\':
			b.WriteRune('\\')
		default:
			b.WriteRune('\\')
			b.WriteRune(r)
		}
	}
	if escaped {
		b.WriteRune('\\')
	}
	return b.String()
}

// setNestedLanguageMessage 按点号键路径把字符串写入嵌套消息 map（中间层自动建 map）。
// [S 汇编 0x1407a4940, 576B]：Split(TrimSpace(key),".") → 逐层：part 空则返回 →
// 末层 cur[part]=value；中间层非 map 则建新 map 下钻。
func setNestedLanguageMessage(messages map[string]interface{}, key string, value string) {
	parts := strings.Split(strings.TrimSpace(key), ".")
	cur := messages
	for i := 0; i < len(parts); i++ {
		part := strings.TrimSpace(parts[i])
		if part == "" {
			return
		}
		if i == len(parts)-1 {
			cur[part] = value
			return
		}
		if m, ok := cur[part].(map[string]interface{}); ok {
			cur = m
		} else {
			m := make(map[string]interface{})
			cur[part] = m
			cur = m
		}
	}
}
