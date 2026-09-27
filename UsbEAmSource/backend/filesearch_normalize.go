// AUTO-RECONSTRUCTED — DOMAIN: filesearch config normalization
// 研究用途 · UsbEAm Launcher 1.0.3 后端方法体还原
package main

import (
	"os"
	"strings"
	"unicode/utf8"
)

// normalizeFileSearchResourceMode 归一化文件搜索资源模式（闭合 LegacyMode→ResourceMode 映射）。
// [S 汇编 0x14087a5c0, 0x1e0]：
//
//	mode := ToLower(TrimSpace(resourceMode))；空则回退 ToLower(TrimSpace(legacyMode))。
//	hot/fast/resident/performance → "resident"；
//	cold/memory/low-memory/low_memory/mem-saver/mem_saver → "memory-saver"；
//	其余（含 balanced）→ "balanced"。
func normalizeFileSearchResourceMode(resourceMode, legacyMode string) string {
	mode := strings.ToLower(strings.TrimSpace(resourceMode))
	if mode == "" {
		mode = strings.ToLower(strings.TrimSpace(legacyMode))
	}
	switch mode {
	case "hot", "fast", "resident", "performance":
		return "resident"
	case "cold", "memory", "low-memory", "low_memory", "mem-saver", "mem_saver":
		return "memory-saver"
	default:
		return "balanced"
	}
}

// normalizeFileSearchConfig 归一化文件搜索配置（14 字段原始配置 → 10 字段归一化配置）。
// [S 汇编 0x140879e00, 0x340] 实证流程：
//
//	Volumes = normalizeVolumeRoots(Volumes)；len==0 时回退 normalizeVolumeRoots(LegacyRoots)
//	MaxResults != 200 时钳制为 200（实测 cmovne，恒 200）
//	IgnoredDirectories 非 nil → normalizeFileSearchIgnoredDirectoryRules，否则 default
//	FileTypeFilters 非 nil → normalizeFileSearchTypeFilters，否则 default
//	RecentItemsEnabled = new(bool)：nil → true，否则取 *cfg
//	RecentItemsLimit：<=0→24；<10→10；>128→128；否则原样
//	RecentItems = normalizeFileSearchRecentItems
//	ResourceMode = normalizeFileSearchResourceMode(ResourceMode, LegacyMode)
//	Enabled / PinyinSearchEnabled 原样透传
//
// 依赖的 6 个 helper 未落地（见下方 [P] 存根）；Enabled 与 ResourceMode 归一化路径已闭合，
// 满足 shouldWarmResidentFileSearchRuntime 的判定需求。
func normalizeFileSearchConfig(cfg FileSearchConfig) fileSearchConfigNormalized {
	volumes := normalizeVolumeRoots(cfg.Volumes)
	if len(volumes) == 0 {
		volumes = normalizeVolumeRoots(cfg.LegacyRoots)
	}

	maxResults := cfg.MaxResults
	if maxResults != 200 {
		maxResults = 200
	}

	var ignoredDirectories []string
	if cfg.IgnoredDirectories != nil {
		ignoredDirectories = normalizeFileSearchIgnoredDirectoryRules(cfg.IgnoredDirectories)
	} else {
		ignoredDirectories = defaultFileSearchIgnoredDirectoryRules()
	}

	var typeFilters []FileSearchTypeFilter
	if cfg.FileTypeFilters != nil {
		typeFilters = normalizeFileSearchTypeFilters(cfg.FileTypeFilters)
	} else {
		typeFilters = defaultFileSearchTypeFilters()
	}

	recentItemsEnabled := new(bool)
	if cfg.RecentItemsEnabled != nil {
		*recentItemsEnabled = *cfg.RecentItemsEnabled
	} else {
		*recentItemsEnabled = true
	}

	recentItemsLimit := cfg.RecentItemsLimit
	switch {
	case recentItemsLimit <= 0:
		recentItemsLimit = 24
	case recentItemsLimit < 10:
		recentItemsLimit = 10
	case recentItemsLimit > 128:
		recentItemsLimit = 128
	}

	return fileSearchConfigNormalized{
		Enabled:             cfg.Enabled,
		PinyinSearchEnabled: cfg.PinyinSearchEnabled,
		RecentItemsEnabled:  recentItemsEnabled,
		RecentItemsLimit:    recentItemsLimit,
		RecentItems:         normalizeFileSearchRecentItems(cfg.RecentItems),
		Volumes:             volumes,
		MaxResults:          maxResults,
		IgnoredDirectories:  ignoredDirectories,
		ResourceMode:        normalizeFileSearchResourceMode(cfg.ResourceMode, cfg.LegacyMode),
		FileTypeFilters:     typeFilters,
	}
}

// ---- 未落地 helper 存根（[P]，真实 VA + 阻断原因）----

// normalizeVolumeRoots 归一化卷根路径（去重）。
// [S 汇编 0x14088c6e0, 0x2c0]：len==0→nil；make([]string,0,len)+make(map[string]struct{})；
// 逐条 normalizeVolumeRoot 空跳过；去重键=strings.ToLower(norm)；键重复跳过；append 原 norm。
func normalizeVolumeRoots(roots []string) []string {
	if len(roots) == 0 {
		return nil
	}
	out := make([]string, 0, len(roots))
	seen := make(map[string]struct{})
	for _, root := range roots {
		norm := normalizeVolumeRoot(root)
		if norm == "" {
			continue
		}
		key := strings.ToLower(norm)
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, norm)
	}
	return out
}

// normalizeFileSearchRecentItems 归一化最近项列表（去重 + 上限 128）。
// [S 汇编 0x14087a140, 0x480]：len==0→nil；cap=min(len,128)；
// 逐条 TrimSpace(Path) 空跳过；RuneCount>0x7fff 跳过；去重键=ToLower(Replace(path,"/","\\",-1))；
// 键重复跳过；Name=trimStringLimit(Name,512)；append {Path:trimmed, Name, IsDirectory}；
// append 后 len>127 终止。
func normalizeFileSearchRecentItems(items []FileSearchRecentItem) []FileSearchRecentItem {
	if len(items) == 0 {
		return nil
	}
	limit := len(items)
	if limit > 128 {
		limit = 128
	}
	out := make([]FileSearchRecentItem, 0, limit)
	seen := make(map[string]struct{})
	for _, item := range items {
		path := strings.TrimSpace(item.Path)
		if path == "" || utf8.RuneCountInString(path) > 0x7fff {
			continue
		}
		key := strings.ToLower(strings.Replace(path, "/", "\\", -1))
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, FileSearchRecentItem{
			Path:        path,
			Name:        trimStringLimit(item.Name, 512),
			IsDirectory: item.IsDirectory,
		})
		if len(out) > 127 {
			break
		}
	}
	return out
}

// defaultFileSearchTypeFilterIDs 内置类型过滤器 ID 集合（mapaccess2 的 map 指针来自全局
// 0x141c0f6f8，初始 nil，由 defaultFileSearchTypeFilters 填充）。
var defaultFileSearchTypeFilterIDs map[string]struct{}

// normalizeFileSearchTypeFilters 归一化类型过滤器列表。
// [S 汇编 0x140815220]：nil→defaultFileSearchTypeFilters()；空→空 slice；
// 上限 32 条；逐条 normalizeFileSearchTypeFilterID 空或 "all" 跳过；seen[id] 去重；
// normalizeFileSearchTypeRules 失败跳过；label=TrimSpace 超 64 rune 截断；
// enabled=原值，若 id 非内置且 label 空→false，若 enabled 且 rules 空→false。
func normalizeFileSearchTypeFilters(filters []FileSearchTypeFilter) []FileSearchTypeFilter {
	if filters == nil {
		return defaultFileSearchTypeFilters()
	}
	if len(filters) == 0 {
		return []FileSearchTypeFilter{}
	}
	out := make([]FileSearchTypeFilter, 0, len(filters))
	seen := make(map[string]struct{})
	n := len(filters)
	if n > 32 {
		n = 32
	}
	for i := 0; i < n; i++ {
		f := filters[i]
		id := normalizeFileSearchTypeFilterID(f.ID)
		if id == "" || id == "all" {
			continue
		}
		if _, dup := seen[id]; dup {
			continue
		}
		rules, ok := normalizeFileSearchTypeRules(f.Rules)
		if !ok {
			continue
		}
		label := strings.TrimSpace(f.Label)
		if utf8.RuneCountInString(label) > 64 {
			label = string([]rune(label)[:64])
		}
		enabled := f.Enabled
		if _, isDefault := defaultFileSearchTypeFilterIDs[id]; !isDefault && label == "" {
			enabled = false
		}
		if enabled && len(rules) == 0 {
			enabled = false
		}
		seen[id] = struct{}{}
		out = append(out, FileSearchTypeFilter{
			ID:      id,
			Label:   label,
			Enabled: enabled,
			Rules:   rules,
		})
	}
	return out
}

// defaultFileSearchTypeFilters 默认类型过滤器表（6 项：图片/文档/视频/音频/压缩包/可执行）。
// [S 汇编 0x140812580, 0x2ca0]：6 次 growslice 构建各 filter 的 Rules 扩展名列表 +
// newobject 分配 6×64B 静态模板 + duffcopy 复制。ID/Label/Enabled 经 duffcopy 源
// (0x1411e5d70) 字节确证：ID∈{image,document,video,audio,archive,executable}，Label 全空，
// Enabled 全 true。Rules 经 6 个 growslice 区域 lea+len 字节级解析（tools/parse_deffilters.py）。
func defaultFileSearchTypeFilters() []FileSearchTypeFilter {
	return []FileSearchTypeFilter{
		{ID: "image", Enabled: true, Rules: []string{
			"jpg", "jpeg", "jpe", "jfif", "pjpeg", "pjp", "png", "apng", "gif", "webp",
			"bmp", "dib", "svg", "svgz", "heic", "heif", "avif", "tif", "tiff", "ico",
			"icns", "cur", "raw", "dng", "cr2", "cr3", "nef", "nrw", "arw", "srf",
			"sr2", "raf", "orf", "rw2", "pef", "srw", "kdc", "dcr", "mrw", "x3f",
			"psd", "psb", "ai", "eps", "ps", "jxl", "jp2", "j2k", "jpf", "jpm",
			"mj2", "tga", "targa", "dds", "exr", "hdr", "pic", "pict", "pct", "wmf",
			"emf", "jxr", "hdp", "wdp",
		}},
		{ID: "document", Enabled: true, Rules: []string{
			"pdf", "xps", "oxps", "doc", "docx", "docm", "dot", "dotx", "dotm", "xls",
			"xlsx", "xlsm", "xlsb", "xlt", "xltx", "xltm", "xlam", "csv", "tsv", "ppt",
			"pptx", "pptm", "pot", "potx", "potm", "pps", "ppsx", "ppsm", "odt", "ott",
			"ods", "ots", "odp", "otp", "rtf", "txt", "text", "md", "markdown", "epub",
			"mobi", "azw", "azw3", "fb2", "djvu", "djv", "chm", "pages", "numbers", "key",
			"tex", "latex", "bib", "json", "xml", "yaml", "yml", "html", "htm", "mhtml",
			"mht", "log", "ini", "cfg", "conf", "nfo",
		}},
		{ID: "video", Enabled: true, Rules: []string{
			"mp4", "m4v", "mkv", "webm", "avi", "mov", "qt", "wmv", "flv", "f4v",
			"mpg", "mpeg", "mpe", "m2v", "mpv", "ts", "m2ts", "mts", "3gp", "3g2",
			"ogv", "rm", "rmvb", "vob", "asf", "divx", "xvid", "hevc", "h265", "h264",
			"avchd", "swf", "mxf",
		}},
		{ID: "audio", Enabled: true, Rules: []string{
			"mp3", "flac", "wav", "wave", "w64", "m4a", "aac", "ogg", "oga", "opus",
			"wma", "alac", "aiff", "aif", "aifc", "ape", "mka", "mid", "midi", "amr",
			"ac3", "eac3", "dts", "m3u", "m3u8", "pls", "cue", "cda", "ra", "ram",
			"au", "snd", "caf", "voc", "tta", "wv", "dsf", "dff", "mpc",
		}},
		{ID: "archive", Enabled: true, Rules: []string{
			"zip", "zipx", "7z", "rar", "tar", "tgz", "tbz", "tbz2", "txz", "tlz",
			"gz", "gzip", "bz", "bz2", "bzip2", "xz", "z", "zst", "tzst", "lz",
			"lzma", "lzo", "br", "cab", "arj", "lzh", "lha", "ace", "arc", "pea",
			"wim", "swm", "esd", "iso", "udf", "img", "cpio", "xar", "sit", "sitx",
			"hqx", "sea", "pak", "*.7z.*", "*.zip.*", "*.rar.*", "*.tar.*", "?*.[rz][0-9][0-9]",
		}},
		{ID: "executable", Enabled: true, Rules: []string{
			"exe", "com", "msi", "msp", "msu", "appx", "appxbundle", "msix", "msixbundle", "appxupload",
			"msixupload", "dmg", "pkg", "mpkg", "deb", "rpm", "run", "bin", "appimage", "flatpak",
			"snap", "apk", "xapk", "apks", "apkm", "aab", "ipa", "jar", "war", "ear",
			"bat", "cmd", "ps1", "psm1", "psd1", "vbs", "vbe", "js", "jse", "wsf",
			"wsh", "sh", "bash", "zsh", "fish", "command", "scr", "cpl", "appref-ms", "application",
		}},
	}
}

// normalizeFileSearchIgnoredDirectoryRules 归一化忽略目录规则（去重）。
// [S 汇编 0x1407e0440]：nil→nil；make([]string,0,len)+make(map[string]struct{})；
// 逐条 normalizeFileSearchIgnoredDirectoryRule 空跳过；绝对路径规则去重键=normalizeFileSearchIgnoredDirectoryMatchPath，
// 否则 ToLower；键空/重复跳过；否则 seen[key]=struct{} 并 append 原 norm。
func normalizeFileSearchIgnoredDirectoryRules(rules []string) []string {
	if rules == nil {
		return nil
	}
	out := make([]string, 0, len(rules))
	seen := make(map[string]struct{})
	for _, rule := range rules {
		norm := normalizeFileSearchIgnoredDirectoryRule(rule)
		if norm == "" {
			continue
		}
		var key string
		if isFileSearchIgnoredDirectoryPathRule(norm) {
			key = normalizeFileSearchIgnoredDirectoryMatchPath(norm)
		} else {
			key = strings.ToLower(norm)
		}
		if key == "" {
			continue
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, norm)
	}
	return out
}

// defaultFileSearchIgnoredDirectoryRules 默认忽略目录规则。
// [S 汇编 0x1407e03c0]：normalizeFileSearchIgnoredDirectoryRules(["node_modules"@0x140c4b9a8,
// TrimSpace(os.TempDir()), "system volume information"@0x140c66ad8])。
func defaultFileSearchIgnoredDirectoryRules() []string {
	return normalizeFileSearchIgnoredDirectoryRules([]string{
		"node_modules",
		strings.TrimSpace(os.TempDir()),
		"system volume information",
	})
}
