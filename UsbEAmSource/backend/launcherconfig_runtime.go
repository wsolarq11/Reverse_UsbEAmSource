// AUTO-RECONSTRUCTED FUNCTION SKELETONS — DOMAIN: launcherconfig (normalize/load 装配层)
// 研究用途
//
// 契约来源：
//   - gap 清单 D:\Users\Administrator\TEMP\gap_launcherconfig.csv（Func,Sym,VA,Len）
//   - 签名证据 D:\Users\Administrator\TEMP\sum_launcherconfig.txt（params/rets/call）
//   - 类型定义 backend/types_launcher.go / types_config.go / types_app.go / types_bookmark.go
//
// 档位：
//   - [S-sig 0xVA]：签名经符号表/命名明确推断，体为零值。
//   - [P]：签名不确定（多寄存器参数未逐寄存器实证 / rets=[] 与命名推断冲突），体为零值。
//   - 已存在符号（loadLauncherConfigRuntimeUnlocked、loadLauncherConfigIfExistsUnlocked 同 VA 0x140876220、
//     idAllocator.Next）与编译器生成函数（funcN）不重复落地。
package main

import (
	"net/url"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// defaultFileLocatorFilterValue 返回默认文件定位过滤值串。
// [S-sig 0x140874760]：无参；汇编尾部 strings.Join 返回 string（rets=[rbx,rdi]）。
func defaultFileLocatorFilterValue() string {
	return ""
}

// loadOrCreateLauncherConfig 加载配置，缺失时按 options 创建默认并落盘。
// [S-sig 0x140875b40]：签名逐寄存器实证——(path string, opts []LauncherConfigOptions)(LauncherConfig,error)。
// 实证：调用点 0x1409c8661 传 (rax,rbx)=path、rcx/rdi/rsi=0 空切片；返回配置经栈 0x126 qword、error 经 rax。
func loadOrCreateLauncherConfig(path string, opts []LauncherConfigOptions) (LauncherConfig, error) {
	return LauncherConfig{}, nil
}

// loadLauncherConfigOrDefaultIfMissing 加载配置，缺失时返回默认配置。
// [S-sig 0x140875e00]：签名逐寄存器实证——(path string, opts []LauncherConfigOptions)(LauncherConfig,bool)。
// 实证：调用点 0x14076478a 传 (rax,rbx)=path、rcx=&opts[0]、rdi/rsi=1 单元素切片；返回配置经栈、bool 经 eax（0/1）。
func loadLauncherConfigOrDefaultIfMissing(path string, opts []LauncherConfigOptions) (LauncherConfig, bool) {
	return LauncherConfig{}, false
}

// stripLauncherConfigRuntimeFields 剥离配置中的运行时字段（图标槽）。
// [S-sig 0x1408775e0]：序言 test rax 空指针检查后传 visitLauncherConfigIconSlots，入参唯一=*LauncherConfig；无返回寄存器=void。
func stripLauncherConfigRuntimeFields(cfg *LauncherConfig) {
}

// readLauncherConfigFile 读取并反序列化配置文件。
// [S-sig 0x1408788c0]：签名逐寄存器实证——(path string, sourceTag string)(LauncherConfig,error)。
// 实证：调用点 0x1407ab554 传 (rax,rbx)=path、(rcx,rdi)=sourceTag（错误消息上下文标签）；返回配置经栈、error 经 rax。
func readLauncherConfigFile(path string, sourceTag string) (LauncherConfig, error) {
	return LauncherConfig{}, nil
}

// launcherConfigDocumentIsIconLibrary 判断配置文档是否为图标库形态。
// [S-sig 0x140879000]：序言 rax+rbx+rcx=[]byte 切片(3) 传入 encoding/json.Unmarshal；尾声 eax 置 0/1 返回 bool。
func launcherConfigDocumentIsIconLibrary(data []byte) bool {
	return false
}

// staticPopulateAppIconOptions 填充图标所用的静态选项（汇编 0x141BE9500 全局，8 qword）。
// 原始 qword：{0, 0, 0x100, 0x4, 0x141965580, 0x5, 0x5, 0}。
// 待取证：与现 AppIconOptions 布局（IconIndex/Namespace/Size/ImageList/CandsPtr/CandsLen）
// 的字段边界存在偏移差——按现布局 Namespace.len=0x100 为畸形串、CandsPtr=5 为非法指针，
// 与程序实际运行不符，疑为 AppIconOptions 布局重构偏差。保守按零值近似（IconIndex=0 已实证，
// 走系统图标路径，为 App 路径的主导分支）。
var staticPopulateAppIconOptions = AppIconOptions{}

// populateAppIcons 为配置中所有 AutoIcon 应用解析并填充图标数据。
// [S 汇编 0x140879ca0, 78 行]：
//
//	func populateAppIcons(cfg *LauncherConfig)（自由函数，非方法；nil 早退）；
//	遍历 cfg.Apps（[]AppEntry，偏移 0x4b8，元素 0x168）：
//	对 AutoIcon(+0x41) 为真且 IconDataVersion(+0x98)<3 的条目，
//	resolveAppIconDataWithOptions(Path(+0xa0/0xa8), 静态选项) → TrimSpace →
//	写 IconData(+0x48/0x50)，置 IconDataVersion=3。
func populateAppIcons(cfg *LauncherConfig) {
	if cfg == nil {
		return
	}
	for i := range cfg.Apps {
		app := &cfg.Apps[i]
		if !app.AutoIcon {
			continue
		}
		if app.IconDataVersion >= 3 {
			continue
		}
		data := resolveAppIconDataWithOptions(app.Path, staticPopulateAppIconOptions)
		app.IconData = strings.TrimSpace(data)
		app.IconDataVersion = 3
	}
}

// normalizeFileLocatorConfig 归一化文件定位配置。
// [S 汇编 0x14087a7a0, 285 行]：
//
//	过滤器归一化后空则补默认过滤器 {ID:"file-locator-default-non-text", Remark:"默认非文本排除",
//	Value:defaultFileLocatorFilterValue(), Type:"glob"}；
//	FileName/ContainsText 历史 flag=false、SearchRoot 历史 flag=true；
//	ActiveFilterID TrimSpace 后非空则按 EqualFold(TrimSpace(ID)) 匹配过滤器并取其原 ID（未命中置空）；
//	IncludeSubfolders 在 isZero 时回退默认(true)；Query/Mode/Scope/Root 各自归一化；
//	MaxSearchFileSizeMB clamp[32,4096]；FileNameMatchCase/ContainsTextMatchCase 直传。
func normalizeFileLocatorConfig(cfg FileLocatorConfig) FileLocatorConfig {
	filters := normalizeFileLocatorFilters(cfg.SavedFilters)
	if len(filters) == 0 {
		filters = []FileLocatorFilter{{
			ID:     "file-locator-default-non-text",
			Remark: "默认非文本排除",
			Value:  defaultFileLocatorFilterValue(),
			Type:   "glob",
		}}
	}
	fileNameHistory := normalizeFileLocatorHistoryEntries(cfg.FileNameHistory, false)
	containsTextHistory := normalizeFileLocatorHistoryEntries(cfg.ContainsTextHistory, false)
	searchRootHistory := normalizeFileLocatorHistoryEntries(cfg.SearchRootHistory, true)

	activeFilterID := strings.TrimSpace(cfg.ActiveFilterID)
	includeSubfolders := cfg.IncludeSubfolders
	if isZeroFileLocatorConfig(cfg) {
		includeSubfolders = true
	}
	if activeFilterID != "" {
		matched := ""
		for _, f := range filters {
			if strings.EqualFold(strings.TrimSpace(f.ID), activeFilterID) {
				matched = f.ID
				break
			}
		}
		activeFilterID = matched
	}

	maxSize := cfg.MaxSearchFileSizeMB
	if maxSize <= 0 {
		maxSize = 32
	} else if maxSize > 4096 {
		maxSize = 4096
	}

	return FileLocatorConfig{
		FileNameQuery:         strings.TrimSpace(cfg.FileNameQuery),
		FileNameMode:          normalizeFileLocatorSearchMode(cfg.FileNameMode),
		FileNameMatchCase:     cfg.FileNameMatchCase,
		FileNameHistory:       fileNameHistory,
		ContainsTextQuery:     strings.TrimSpace(cfg.ContainsTextQuery),
		ContainsTextMode:      normalizeFileLocatorSearchMode(cfg.ContainsTextMode),
		ContainsTextMatchCase: cfg.ContainsTextMatchCase,
		ContainsTextHistory:   containsTextHistory,
		BooleanScope:          normalizeFileLocatorBooleanScope(cfg.BooleanScope),
		SearchRoot:            strings.TrimSpace(cfg.SearchRoot),
		SearchRootHistory:     searchRootHistory,
		MaxSearchFileSizeMB:   maxSize,
		IncludeSubfolders:     includeSubfolders,
		ActiveFilterID:        activeFilterID,
		SavedFilters:          filters,
	}
}

// normalizeFileLocatorSearchMode 归一化搜索模式串。
// [S 0x14087ae80]：TrimSpace→ToLower 后按长度跳表白名单，命中返 canonical，否则 "plain"。
func normalizeFileLocatorSearchMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "bool", "boolean":
		return "boolean"
	case "word", "whole-word":
		return "wholeWord"
	case "regex", "regexp":
		return "regex"
	case "boolregex", "bool-regex", "booleanregex", "boolean-regex":
		return "booleanRegex"
	default:
		return "plain"
	}
}

// normalizeFileLocatorBooleanScope 归一化布尔作用域串。
// [S 0x14087b060]：TrimSpace→ToLower；"line"/"lines"→"line"，否则 "file"。
func normalizeFileLocatorBooleanScope(scope string) string {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "line", "lines":
		return "line"
	default:
		return "file"
	}
}

// normalizeFileLocatorFilterType 归一化过滤器类型串。
// [S 0x14087b100]：TrimSpace→ToLower；bool 族→"boolean"，regex 族→"regex"，glob 族→"glob"，否则 "plain"。
func normalizeFileLocatorFilterType(typ string) string {
	switch strings.ToLower(strings.TrimSpace(typ)) {
	case "bool", "boolean":
		return "boolean"
	case "regex", "regexp":
		return "regex"
	case "glob", "mask", "wildcard":
		return "glob"
	default:
		return "plain"
	}
}

// normalizeFileLocatorFilters 归一化过滤器切片。
// [S-sig 0x14087b240]：命名 normalize...Filters → []FileLocatorFilter；rets 3 寄存器=slice。
func normalizeFileLocatorFilters(filters []FileLocatorFilter) []FileLocatorFilter {
	return nil
}

// normalizeFileLocatorHistoryEntries 归一化历史条目切片。
// [S 汇编 0x14087b7a0, 179 行]：
//
//	去空白、去空、按 map 去重、上限 12；normalizePath 时先反斜杠→斜杠再小写作为去重键，输出保留原去空白串。
//	签名修正：原 [S-sig] 误为 ([]string) []string，实为 ([]string, bool) []string（dil 入参）。
func normalizeFileLocatorHistoryEntries(entries []string, normalizePath bool) []string {
	if len(entries) == 0 {
		return nil
	}
	cap := len(entries)
	if cap > 12 {
		cap = 12
	}
	result := make([]string, 0, cap)
	seen := make(map[string]struct{})
	for _, entry := range entries {
		trimmed := strings.TrimSpace(entry)
		if trimmed == "" {
			continue
		}
		key := trimmed
		if normalizePath {
			key = strings.ToLower(strings.Replace(trimmed, "\\", "/", -1))
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, trimmed)
		if len(result) >= 12 {
			break
		}
	}
	return result
}

// isZeroFileLocatorConfig 判断配置是否为零值。
// [S-sig 0x14087bac0]：命名 is...Config → bool；入参 cfg 按值（汇编按指针）。
func isZeroFileLocatorConfig(cfg FileLocatorConfig) bool {
	return false
}

// defaultTagCatalogItems 返回默认标签目录项切片。
// [S-sig 0x14087bc20]：命名 default...Items → []TagCatalogItem；rets 3 寄存器=slice。
func defaultTagCatalogItems() []TagCatalogItem {
	return nil
}

// normalizeTagCatalogWithDefault 归一化标签目录并合并默认项。
// [S 汇编 0x14087c340, 263 行]：
//
//	签名修正：原 [P] 误为 (items,fallbackNames) 二参，实为三切片
//	(items []TagCatalogItem, fallbackNames []string, fallbackItems []TagCatalogItem) []TagCatalogItem。
//	items 空且 fallbackNames 非空 → defaultTagCatalogFromNames；仍空则 src=fallbackItems（再空 → defaultTagCatalogItems）；
//	按小写 Name 去重，Name/Icon/IconData/IconRef/IconURL TrimSpace，IconData 非 "data:image/" 前缀清空，
//	Icon 空则 defaultTagCatalogIcon(Name)；结果空且 fallbackItems 非空 → 尾递归。
func normalizeTagCatalogWithDefault(items []TagCatalogItem, fallbackNames []string, fallbackItems []TagCatalogItem) []TagCatalogItem {
	if len(items) == 0 && len(fallbackNames) != 0 {
		items = defaultTagCatalogFromNames(fallbackNames)
	}
	var src []TagCatalogItem
	switch {
	case len(items) != 0:
		src = items
	case len(fallbackItems) != 0:
		src = fallbackItems
	default:
		return defaultTagCatalogItems()
	}

	result := make([]TagCatalogItem, 0, len(src))
	seen := make(map[string]struct{})
	for _, it := range src {
		name := strings.TrimSpace(it.Name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		icon := strings.TrimSpace(it.Icon)
		iconData := strings.TrimSpace(it.IconData)
		iconRef := strings.TrimSpace(it.IconRef)
		if !strings.HasPrefix(strings.ToLower(iconData), "data:image/") {
			iconData = ""
		}
		if icon == "" {
			icon = defaultTagCatalogIcon(name)
		}
		seen[key] = struct{}{}
		result = append(result, TagCatalogItem{
			Name:     name,
			Icon:     icon,
			IconData: iconData,
			IconRef:  iconRef,
			IconURL:  strings.TrimSpace(it.IconURL),
		})
	}
	if len(result) == 0 {
		if len(fallbackItems) != 0 {
			return normalizeTagCatalogWithDefault(fallbackItems, nil, nil)
		}
		return defaultTagCatalogItems()
	}
	return result
}

// normalizeBackgroundPreference 归一化背景偏好配置。
// [S 汇编 0x14087eae0, 202 行]：
//
//	ShellOpacity clamp[0,1] 后作为 Sidebar/ContentTransparency 的 nil 回退默认，本身不落输出（置 nil）；
//	ReadabilityOverlayEnabled nil→true；ReadabilityOverlayOpacity clamp[0,0.65] nil→0.18；
//	ImageOpacity clamp[0,1] nil→1.0；ImageBlur clamp[0,24]；Sidebar/ContentTransparency clamp[0,1] nil→shellOpacity。
//	字符串 ImagePath/ImageURL TrimSpace，Layout 走 normalizeBackgroundLayout，Enabled 直传。
func normalizeBackgroundPreference(pref BackgroundPreference) BackgroundPreference {
	shellOpacity := 0.0
	if pref.ShellOpacity != nil {
		shellOpacity = *pref.ShellOpacity
		if shellOpacity < 0 {
			shellOpacity = 0
		} else if shellOpacity > 1 {
			shellOpacity = 1
		}
	}
	readabilityOverlayEnabled := true
	if pref.ReadabilityOverlayEnabled != nil {
		readabilityOverlayEnabled = *pref.ReadabilityOverlayEnabled
	}
	readabilityOverlayOpacity := 0.18
	if pref.ReadabilityOverlayOpacity != nil {
		readabilityOverlayOpacity = *pref.ReadabilityOverlayOpacity
		if readabilityOverlayOpacity < 0 {
			readabilityOverlayOpacity = 0
		} else if readabilityOverlayOpacity > 0.65 {
			readabilityOverlayOpacity = 0.65
		}
	}
	imageOpacity := 1.0
	if pref.ImageOpacity != nil {
		imageOpacity = *pref.ImageOpacity
		if imageOpacity < 0 {
			imageOpacity = 0
		} else if imageOpacity > 1 {
			imageOpacity = 1
		}
	}
	imageBlur := pref.ImageBlur
	if imageBlur < 0 {
		imageBlur = 0
	} else if imageBlur > 24 {
		imageBlur = 24
	}
	sidebarTransparency := shellOpacity
	if pref.SidebarTransparency != nil {
		sidebarTransparency = *pref.SidebarTransparency
		if sidebarTransparency < 0 {
			sidebarTransparency = 0
		} else if sidebarTransparency > 1 {
			sidebarTransparency = 1
		}
	}
	contentTransparency := shellOpacity
	if pref.ContentTransparency != nil {
		contentTransparency = *pref.ContentTransparency
		if contentTransparency < 0 {
			contentTransparency = 0
		} else if contentTransparency > 1 {
			contentTransparency = 1
		}
	}
	return BackgroundPreference{
		Enabled:                   pref.Enabled,
		ImagePath:                 strings.TrimSpace(pref.ImagePath),
		ImageURL:                  strings.TrimSpace(pref.ImageURL),
		Layout:                    normalizeBackgroundLayout(pref.Layout),
		ReadabilityOverlayEnabled: &readabilityOverlayEnabled,
		ReadabilityOverlayOpacity: &readabilityOverlayOpacity,
		ImageOpacity:              &imageOpacity,
		ImageBlur:                 imageBlur,
		SidebarTransparency:       &sidebarTransparency,
		ContentTransparency:       &contentTransparency,
	}
}

// normalizeBackgroundLayout 归一化背景布局串。
// [S-sig 0x14087ee40]：命名 normalize...Layout → string；1 string 入 / 1 string 出。
func normalizeBackgroundLayout(layout string) string {
	return ""
}

// normalizeConsoleItems 归一化控制台条目切片。
// [S-sig 0x14087f160]：命名 normalize...Items → []ConsoleItem；rets 3 寄存器=slice。
func normalizeConsoleItems(items []ConsoleItem) []ConsoleItem {
	return nil
}

// normalizeConsoleItem 归一化单条控制台条目。
// [S 汇编 0x14087f560, 0xe0=3648 字节]：
//
//	kind=normalizeConsoleItemKind(Kind)；section/target/source=trimStringLimit(各,0x40/0x104/0x104)；
//	分类用有效 target（target 空时回落 section）把 window 管理/鼠标手势两类特殊条目重映射到
//	canonical kind/section/target；kind/section/target 任一空 → 返零值条目；
//	再归一化 folderPath/title(0xa0)/subtitle(0xdc)/icon(0x40)/iconData/iconRef(TrimSpace)/iconURL(TrimSpace)；
//	两类 canonical 条目清空 SourceID/FolderPath/Title/Subtitle/IconData/IconRef/IconURL 且 Icon 置
//	"window"/"pointer"；布局四项 clampLayout；最后 ID=buildConsoleItemID(结果)。
func normalizeConsoleItem(item ConsoleItem) ConsoleItem {
	kind := normalizeConsoleItemKind(item.Kind)
	sectionID := trimStringLimit(item.SectionID, 0x40)
	targetID := trimStringLimit(item.TargetID, 0x104)

	// 分类用有效 target：targetID 为空时回落到 sectionID。
	effTarget := targetID
	if effTarget == "" {
		effTarget = sectionID
	}

	switch {
	case kind == "section" && effTarget == "windowManagement":
		kind, sectionID, targetID = "windowManagementTool", "windowManagement", "windowTools"
	case kind == "windowManagementTool" && sectionID == "windowManagement" && targetID != "windowTools":
		kind, sectionID, targetID = "windowManagementTool", "windowManagement", "windowTools"
	case kind == "section" && effTarget == "mouseGestures":
		kind, sectionID, targetID = "mouseGestureTool", "mouseGestures", "gestures"
	case kind == "mouseGestureTool" && sectionID == "mouseGestures" && targetID != "gestures":
		kind, sectionID, targetID = "mouseGestureTool", "mouseGestures", "gestures"
	}

	if kind == "" || sectionID == "" || targetID == "" {
		return ConsoleItem{}
	}

	sourceID := trimStringLimit(item.SourceID, 0x104)
	folderPath := normalizeBookmarkFolderPathValue(item.FolderPath)
	title := trimStringLimit(item.Title, 0xa0)
	subtitle := trimStringLimit(item.Subtitle, 0xdc)
	icon := trimStringLimit(item.Icon, 0x40)
	iconData := normalizeConsoleIconData(item.IconData)
	iconRef := strings.TrimSpace(item.IconRef)
	iconURL := strings.TrimSpace(item.IconURL)

	clearContent := false
	switch {
	case kind == "windowManagementTool" && sectionID == "windowManagement" && targetID == "windowTools":
		icon, clearContent = "window", true
	case kind == "mouseGestureTool" && sectionID == "mouseGestures" && targetID == "gestures":
		icon, clearContent = "pointer", true
	}
	if clearContent {
		sourceID, folderPath, title, subtitle, iconData, iconRef, iconURL = "", "", "", "", "", "", ""
	}

	result := ConsoleItem{
		Kind:       kind,
		SectionID:  sectionID,
		TargetID:   targetID,
		SourceID:   sourceID,
		FolderPath: folderPath,
		Title:      title,
		Subtitle:   subtitle,
		Icon:       icon,
		IconData:   iconData,
		IconRef:    iconRef,
		IconURL:    iconURL,
		LayoutX:    min(max(item.LayoutX, 0), 0x40),
		LayoutY:    min(max(item.LayoutY, 0), 0x200),
		LayoutW:    min(max(item.LayoutW, 0), 0x40),
		LayoutH:    min(max(item.LayoutH, 0), 0x200),
	}
	result.ID = buildConsoleItemID(result)
	return result
}

// normalizeConsoleItemKind 归一化控制台条目类型串。
// [S 汇编 0x1408803a0, 0x215]：TrimSpace 后按长度跳表做大小写敏感白名单匹配，
// 命中返 TrimSpace 原串，未命中返 ""。白名单见下。
func normalizeConsoleItemKind(kind string) string {
	kind = strings.TrimSpace(kind)
	switch kind {
	case "section", "appGroup", "audioCard", "qrCodeTool", "catalogItem",
		"desktopWidget", "screensTool", "mouseGestureTool", "memoryReleaseMode",
		"gpuPreferenceEntry", "oledBlackoutProfile", "windowManagementTool":
		return kind
	default:
		return ""
	}
}

// buildConsoleItemID 构造控制台条目 ID。
// [S 汇编 0x1408805e0, 154 行]：
//
//	Kind/SectionID/TargetID 任一空 → 返 ""；
//	parts = [Kind, SectionID, SourceID, TargetID, FolderPath]；
//	各段 url.PathEscape（net/url.escape mode=2=encodePathSegment）→ strings.Join(":", ...)。
func buildConsoleItemID(item ConsoleItem) string {
	if item.Kind == "" || item.SectionID == "" || item.TargetID == "" {
		return ""
	}
	parts := []string{item.Kind, item.SectionID, item.SourceID, item.TargetID, item.FolderPath}
	escaped := make([]string, 0, len(parts))
	for _, p := range parts {
		escaped = append(escaped, url.PathEscape(p))
	}
	return strings.Join(escaped, ":")
}

// normalizeConsoleIconData 归一化控制台图标数据串。
// [S 0x1408808a0]：TrimSpace 后小写比对 "data:image/" 前缀（11 字节 memequal），
// 命中返回 TrimSpace 原串，否则返回空串。
func normalizeConsoleIconData(data string) string {
	s := strings.TrimSpace(data)
	if strings.HasPrefix(strings.ToLower(s), "data:image/") {
		return s
	}
	return ""
}

// trimStringLimit 先去空白，再按 rune 上限截断。
// [S 汇编 0x140880920, 0xdc]：TrimSpace；s=="" 或 limit<=0 → ""；
// rune 数 <= limit → 返 TrimSpace 原串；否则取前 limit 个 rune 回串。
func trimStringLimit(s string, limit int) string {
	s = strings.TrimSpace(s)
	if s == "" || limit <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	return string([]rune(s)[:limit])
}

// normalizeStartMenuLaunchLogs 归一化开始菜单启动日志切片。
// [S-sig 0x140880a40]：命名 normalize...Logs → []StartMenuLaunchLog；rets 3 寄存器=slice。
func normalizeStartMenuLaunchLogs(logs []StartMenuLaunchLog) []StartMenuLaunchLog {
	return nil
}

// normalizeStartMenuLaunchLog 归一化单条启动日志。
// [S 汇编 0x140881400, 166 行]：
//
//	log.ID = TrimPrefix(TrimSpace(ID), "start-menu:")
//	log.Name / ShortcutPath / TargetPath / LastLaunchedAt = TrimSpace(各)
//	LaunchCount < 0 → 0（cmovl 夹取非负）
//	(ID 空 且 ShortcutPath 空 且 TargetPath 空) → 返零值
//	(LaunchCount <= 0 且 LastLaunchedAt 空) → 返零值
//	否则返归一化后的 log
//
// 汇编实证：11B 前缀 memequal(0x140c4754f) 解码 "start-menu:"（TrimPrefix）；
// 首空检 (ID/ShortcutPath/TargetPath 三者皆空) + 次空检 (LaunchCount<=0 且 LastLaunchedAt 空)，
// 两空检均 duffzero 返零态；结构体字段偏移与 types_app.go:128 逐一对齐。
func normalizeStartMenuLaunchLog(log StartMenuLaunchLog) StartMenuLaunchLog {
	log.ID = strings.TrimPrefix(strings.TrimSpace(log.ID), "start-menu:")
	log.Name = strings.TrimSpace(log.Name)
	log.ShortcutPath = strings.TrimSpace(log.ShortcutPath)
	log.TargetPath = strings.TrimSpace(log.TargetPath)
	if log.LaunchCount < 0 {
		log.LaunchCount = 0
	}
	log.LastLaunchedAt = strings.TrimSpace(log.LastLaunchedAt)

	if log.ID == "" && log.ShortcutPath == "" && log.TargetPath == "" {
		return StartMenuLaunchLog{}
	}
	if log.LaunchCount <= 0 && log.LastLaunchedAt == "" {
		return StartMenuLaunchLog{}
	}
	return log
}

// startMenuLaunchLogKeys 提取启动日志的身份键。
// [S 汇编 0x140881700, 138 行]：
//
//	keys := make([]string, 0, 3)
//	ID 非空 → append "id:"+ToLower(ID)
//	ShortcutPath 非空 → append "shortcut:"+ToLower(ShortcutPath)
//	TargetPath 非空 → append "target:"+ToLower(TargetPath)
//	return keys
//
// 汇编实证：makeslice(0,3) 预分配；三字段按序判空后 concatstring2(前缀, ToLower(field))，
// 前缀字面量 read_gostring 解码 id:(3B)/shortcut:(9B)/target:(7B)；返 (ptr,len,cap) 3 字 slice。
func startMenuLaunchLogKeys(log StartMenuLaunchLog) []string {
	keys := make([]string, 0, 3)
	if log.ID != "" {
		keys = append(keys, "id:"+strings.ToLower(log.ID))
	}
	if log.ShortcutPath != "" {
		keys = append(keys, "shortcut:"+strings.ToLower(log.ShortcutPath))
	}
	if log.TargetPath != "" {
		keys = append(keys, "target:"+strings.ToLower(log.TargetPath))
	}
	return keys
}

// mergeStartMenuLaunchLog 合并启动日志条目。
// [S 汇编 0x140881940, 147 行]：
//
//	existing.LaunchCount = max(existing.LaunchCount, incoming.LaunchCount)  // 读原 incoming
//	if incoming.LastLaunchedAt > existing.LastLaunchedAt → swap(existing, incoming)
//	existing 的空 ID/Name/ShortcutPath/TargetPath 依次从 incoming 回填
//	return existing
//
// 汇编实证：两 88B StartMenuLaunchLog 栈传（S1@[0xe0] S2@[0x138]），返单 88B 结构（[0x190]）；
// cmpstring(B.LastLaunchedAt, A.LastLaunchedAt)>0 触发整结构 swap；LaunchCount max 读 [0x178]（原第二参）；
// 尾 cmpstring 对 LastLaunchedAt 的 max 为冗余空操作（swap 已保证较晚者为基底）。
func mergeStartMenuLaunchLog(existing, incoming StartMenuLaunchLog) StartMenuLaunchLog {
	if existing.LaunchCount < incoming.LaunchCount {
		existing.LaunchCount = incoming.LaunchCount
	}
	if incoming.LastLaunchedAt > existing.LastLaunchedAt {
		existing, incoming = incoming, existing
	}
	if existing.ID == "" {
		existing.ID = incoming.ID
	}
	if existing.Name == "" {
		existing.Name = incoming.Name
	}
	if existing.ShortcutPath == "" {
		existing.ShortcutPath = incoming.ShortcutPath
	}
	if existing.TargetPath == "" {
		existing.TargetPath = incoming.TargetPath
	}
	return existing
}

// extractDragLaunchAppIDsFromLegacyRules 从旧拖拽规则提取应用 ID 列表。
// [S-sig 0x140881c20]：命名 extract...AppIDs → []string；params 3 寄存器=slice。
func extractDragLaunchAppIDsFromLegacyRules(rules []DragLaunchRule) []string {
	return nil
}

// createIdentifierKeySet 由 ID 列表构造去重键集合。
// [S-sig 0x140882280]：命名 create...KeySet → map[string]struct{}（call=runtime.makemap）。
func createIdentifierKeySet(ids []string) map[string]struct{} {
	return nil
}

// cleanIdentifierList 清洗 ID 列表（去空/去重/保序）。
// [S-sig 0x140882380]：命名 clean...List → []string；rets 3 寄存器=slice。
func cleanIdentifierList(ids []string) []string {
	return nil
}

// normalizeHTTPProxy 归一化 HTTP 代理串。
// [S-sig 0x140882a20]：命名 normalize...Proxy → string。
func normalizeHTTPProxy(proxy string) string {
	return ""
}

// normalizeWebViewMemoryReleaseStrategy 归一化内存释放策略串。
// [S-sig 0x140882b80]：命名 normalize...Strategy → string；1 string 入 / 1 string 出。
func normalizeWebViewMemoryReleaseStrategy(strategy string) string {
	return ""
}

// normalizeBookmarkFavoritePath 归一化书签收藏路径配置。
// [S 汇编 0x140882ee0, 44 行]：
//
//	source := strings.TrimSpace(p.SourcePath)
//	if source == "" → 返 BookmarkFavoritePath{}（零值）
//	p.FolderPath = normalizeBookmarkFolderPathValue(p.FolderPath)
//	return BookmarkFavoritePath{SourcePath: source, FolderPath: p.FolderPath}
//
// 汇编实证：rax/rbx=SourcePath(ptr,len) → TrimSpace；rcx/rdi=FolderPath(ptr,len)
// → normalizeBookmarkFolderPathValue(FolderPath)；返 (trimmed SourcePath, normalized FolderPath) 四字。
func normalizeBookmarkFavoritePath(p BookmarkFavoritePath) BookmarkFavoritePath {
	source := strings.TrimSpace(p.SourcePath)
	if source == "" {
		return BookmarkFavoritePath{}
	}
	return BookmarkFavoritePath{
		SourcePath: source,
		FolderPath: normalizeBookmarkFolderPathValue(p.FolderPath),
	}
}

// normalizeLinkBrowserKind 归一化链接浏览器类型串。
// [S-sig 0x140882f80]：命名 normalize...Kind → string。
func normalizeLinkBrowserKind(kind string) string {
	return ""
}

// inferLinkBrowserKindFromPath 由路径推断浏览器类型串。
// [S-sig 0x140883060]：命名 infer...Kind → string。
func inferLinkBrowserKindFromPath(path string) string {
	return ""
}

// inferLinkBrowserName 由路径与类型推断浏览器名。
// [S-sig 0x140883160]：命名 infer...Name → string；rets 2 寄存器=string。
func inferLinkBrowserName(path string, kind string) string {
	return ""
}

// normalizeLinkBrowsers 归一化链接浏览器切片。
// [S-sig 0x140883340]：命名 normalize...Browsers → []LinkBrowser；rets 3 寄存器=slice。
func normalizeLinkBrowsers(browsers []LinkBrowser) []LinkBrowser {
	return nil
}

// normalizeDefaultLinkBrowserID 归一化默认浏览器 ID。
// [S 汇编 0x140883920, 92 行]：
//
//	id = strings.TrimSpace(id)；空 → 返 ""
//	id == "__system_default__"（18B 字面量 memequal 实证）→ 返原字面量
//	遍历 browsers：bid := TrimSpace(b.ID)；EqualFold(bid, id) 命中 → 返 bid
//	未命中 → 返 ""
//
// 汇编实证：0x12=18 长度判等 + memequal(0x140c59bfa) 命中 "__system_default__"；
// 逐元素 0x58=88B（LinkBrowser 布局）取 [rcx]=ID，TrimSpace 后 EqualFold，命中返 TrimSpace(ID)。
func normalizeDefaultLinkBrowserID(id string, browsers []LinkBrowser) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	if id == "__system_default__" {
		return "__system_default__"
	}
	for _, b := range browsers {
		bid := strings.TrimSpace(b.ID)
		if strings.EqualFold(bid, id) {
			return bid
		}
	}
	return ""
}

// findConfiguredLinkBrowserByID 按 ID 查找已配置浏览器。
// [S 汇编 0x140883aa0, 142 行]：
//
//	id = strings.TrimSpace(id)；空 → 返 (LinkBrowser{}, false)
//	遍历 browsers：bid := TrimSpace(b.ID)；EqualFold(bid, id) 命中 → 返 (b, true)
//	未命中 → 返 (LinkBrowser{}, false)
//
// 汇编实证：返 (LinkBrowser 0x58=88B 栈结构, bool eax)；命中 eax=1 + duffcopy 全量拷贝元素，
// 未命中/空 eax=0 + duffzero 清零返回缓冲。元素 0x58=88B 布局取 [rcx]=ID（types_app.go:75）。
func findConfiguredLinkBrowserByID(browsers []LinkBrowser, id string) (LinkBrowser, bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return LinkBrowser{}, false
	}
	for _, b := range browsers {
		bid := strings.TrimSpace(b.ID)
		if strings.EqualFold(bid, id) {
			return b, true
		}
	}
	return LinkBrowser{}, false
}

// normalizeBookmarkFolderPathValue 归一化书签文件夹路径值。
// [S 0x140883d00]：TrimSpace → 去前缀 "/" → TrimSpace → 去后缀 "/" → TrimSpace，
// 再按 " / " 分割、逐段 TrimSpace 去空、按 " / " 重接。
func normalizeBookmarkFolderPathValue(path string) string {
	s := strings.TrimSpace(path)
	s = strings.TrimPrefix(s, "/")
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "/")
	s = strings.TrimSpace(s)
	parts := strings.Split(s, " / ")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return strings.Join(out, " / ")
}

// normalizeAppSortMode 归一化应用排序模式串。
// [S-sig 0x140883fc0]：命名 normalize...Mode → string。
func normalizeAppSortMode(mode string) string {
	return ""
}

// normalizeSpeedDialSortMode 归一化速拨排序模式串。
// [S-sig 0x1408840a0]：命名 normalize...Mode → string。
func normalizeSpeedDialSortMode(mode string) string {
	return ""
}

// normalizeSearchHistoryEntries 归一化搜索历史条目切片。
// [S-sig 0x140884180]：命名 normalize...HistoryEntries → []string；params 3 寄存器=slice。
func normalizeSearchHistoryEntries(entries []string) []string {
	return nil
}

// normalizeSearchHistoryEntry 归一化单条搜索历史（Unicode 归一）。
// [S-sig 0x140884460]：命名 normalize...Entry → string；call=x/text/unicode/norm。
func normalizeSearchHistoryEntry(entry string) string {
	return ""
}

// searchHistoryIdentity 计算搜索历史条目标识串。
// [S-sig 0x140884540]：命名 ...Identity → string；rets=[r9] 单寄存器。
func searchHistoryIdentity(entry string) string {
	return ""
}

// normalizeAppCategoryOrder 归一化应用分类顺序。
// [S-sig 0x1408845e0]：命名 normalize...Order → []string；rets 3 寄存器=slice。
func normalizeAppCategoryOrder(order []string) []string {
	return nil
}

// normalizeSpeedDialCategoryOrder 归一化速拨分类顺序。
// [S-sig 0x1408853a0]：命名 normalize...Order → []string；rets 3 寄存器=slice。
func normalizeSpeedDialCategoryOrder(order []string) []string {
	return nil
}

// normalizeSpeedDialExpandedGroups 归一化速拨展开分组。
// [S-sig 0x140885da0]：命名 normalize...Groups → []string；rets 3 寄存器=slice。
func normalizeSpeedDialExpandedGroups(groups []string) []string {
	return nil
}

// normalizeAppExpandedGroups 归一化应用展开分组。
// [S-sig 0x1408863a0]：命名 normalize...Groups → []string；rets 3 寄存器=slice。
func normalizeAppExpandedGroups(groups []string) []string {
	return nil
}

// normalizeSearchCategoryOrder 归一化搜索分类顺序。
// [S-sig 0x140887000]：命名 normalize...Order → []string；params/rets 3 寄存器=slice。
func normalizeSearchCategoryOrder(order []string) []string {
	return nil
}

// normalizeSearchCategoryOrderID 归一化搜索分类顺序 ID 串。
// [S-sig 0x140887520]：命名 normalize...OrderID → string。
func normalizeSearchCategoryOrderID(id string) string {
	return ""
}

// normalizeNavigationVisibleModules 归一化可见导航模块列表。
// [S-sig 0x140887e40]：命名 normalize...Modules → []string；params/rets 3 寄存器=slice。
func normalizeNavigationVisibleModules(modules []string) []string {
	return nil
}

// normalizeSearchScope 归一化搜索范围串。
// [S-sig 0x140888180]：命名 normalize...Scope → string。
func normalizeSearchScope(scope string) string {
	return ""
}

// normalizeSearchScopeCycle 归一化搜索范围循环列表。
// [S-sig 0x1408883c0]：命名 normalize...Cycle → []string；call=runtime.makeslice。
func normalizeSearchScopeCycle(cycle []string) []string {
	return nil
}

// normalizeSearchScopeCycleDefault 归一化搜索范围循环默认值。
// [S 汇编 0x1408886a0, 89 行]：
//
//	scope := strings.TrimSpace(defaultScope)
//	if scope == "" → 返 ""
//	normalized := normalizeSearchScope(scope)
//	for _, c := range normalizeSearchScopeCycle(cycle) { if c == normalized → 返 normalized }
//	return ""
//
// 汇编实证：rax/rbx=defaultScope → TrimSpace；normalizeSearchScope(trimmed) 得 normalized；
// normalizeSearchScopeCycle(cycle) 得 []string（16B/元素）→ 逐元素 cmp len + memequal 判等；
// 命中返 normalized，未命中返空串。
func normalizeSearchScopeCycleDefault(defaultScope string, cycle []string) string {
	scope := strings.TrimSpace(defaultScope)
	if scope == "" {
		return ""
	}
	normalized := normalizeSearchScope(scope)
	for _, c := range normalizeSearchScopeCycle(cycle) {
		if c == normalized {
			return normalized
		}
	}
	return ""
}

// normalizeVolumeRoot 归一化卷根路径。
// [S 汇编 0x14088c9a0]：TrimSpace 空→""；vol=VolumeName 非 "X:"（len!=2 或 [1]!=':'）→""；
// 否则 strings.ToUpper(vol)+"\\"（分隔符 @0x1411cac40）。
func normalizeVolumeRoot(root string) string {
	root = strings.TrimSpace(root)
	if root == "" {
		return ""
	}
	vol := filepath.VolumeName(root)
	if len(vol) != 2 || vol[1] != ':' {
		return ""
	}
	return strings.ToUpper(vol) + "\\"
}
