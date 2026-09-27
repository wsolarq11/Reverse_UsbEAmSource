// AUTO-RECONSTRUCTED TYPES — DOMAIN: config
// 研究用途
package main

type BackgroundPreference struct {
	Enabled                   bool     `json:"enabled,omitempty"`
	ImagePath                 string   `json:"imagePath,omitempty"`
	ImageURL                  string   `json:"imageUrl,omitempty"`
	Layout                    string   `json:"layout,omitempty"`
	ReadabilityOverlayEnabled *bool    `json:"readabilityOverlayEnabled,omitempty"`
	ReadabilityOverlayOpacity *float64 `json:"readabilityOverlayOpacity,omitempty"`
	ImageOpacity              *float64 `json:"imageOpacity,omitempty"`
	ImageBlur                 int      `json:"imageBlur,omitempty"`
	SidebarTransparency       *float64 `json:"sidebarTransparency,omitempty"`
	ContentTransparency       *float64 `json:"contentTransparency,omitempty"`
	ShellOpacity              *float64 `json:"shellOpacity,omitempty"`
}

type ConfigImportPreview struct {
	SourceFile      string          `json:"sourceFile"`
	SourceDirectory string          `json:"sourceDirectory"`
	Workspace       WorkspaceLayout `json:"workspace"`
}

type ConfigMigrationResult struct {
	OldConfig       string `json:"oldConfig"`
	NewConfig       string `json:"newConfig"`
	OldDataRoot     string `json:"oldDataRoot,omitempty"`
	NewDataRoot     string `json:"newDataRoot,omitempty"`
	RestartRequired bool   `json:"restartRequired"`
}

type InitializationImportRequest struct {
	SourceFile string `json:"sourceFile"`
	Mode       string `json:"mode"`
}

type Preferences struct {
	ThemeMode                              string               `json:"themeMode"`
	UIScalePercent                         int                  `json:"uiScalePercent,omitempty"`
	HTTPProxy                              string               `json:"httpProxy,omitempty"`
	AdDisplayEnabled                       *bool                `json:"adDisplayEnabled,omitempty"`
	AppLaunchPrivilegeDefault              string               `json:"appLaunchPrivilegeDefault,omitempty"`
	AppLaunchHideToTrayEnabled             *bool                `json:"appLaunchHideToTrayEnabled,omitempty"`
	StartupLaunchEnabled                   *bool                `json:"startupLaunchEnabled,omitempty"`
	StartupLaunchDelaySeconds              int                  `json:"startupLaunchDelaySeconds,omitempty"`
	DynamicStartMenuEnabled                *bool                `json:"dynamicStartMenuEnabled,omitempty"`
	DynamicDesktopEnabled                  *bool                `json:"dynamicDesktopEnabled,omitempty"`
	WebViewMemorySaverV8LimitEnabled       *bool                `json:"webviewMemorySaverV8LimitEnabled,omitempty"`
	WebViewMemorySaverEdgeExtrasDisabled   *bool                `json:"webviewMemorySaverEdgeExtrasDisabled,omitempty"`
	WebViewMemorySaverRendererLimitEnabled *bool                `json:"webviewMemorySaverRendererLimitEnabled,omitempty"`
	WebViewMemoryReleaseStrategy           string               `json:"webviewMemoryReleaseStrategy,omitempty"`
	WebViewDestroyOnTrayHideEnabled        *bool                `json:"webviewDestroyOnTrayHideEnabled,omitempty"`
	ConfigBackupLimit                      int                  `json:"configBackupLimit,omitempty"`
	Language                               string               `json:"language"`
	Background                             BackgroundPreference `json:"background,omitempty"`
	Hotkey                                 string               `json:"hotkey"`
	HotkeyEnabled                          *bool                `json:"hotkeyEnabled,omitempty"`
	SummonOnlyHotkey                       string               `json:"summonOnlyHotkey,omitempty"`
	SummonOnlyHotkeyEnabled                *bool                `json:"summonOnlyHotkeyEnabled,omitempty"`
	ScreenshotHotkey                       string               `json:"screenshotHotkey"`
	ScreenshotHotkeyEnabled                *bool                `json:"screenshotHotkeyEnabled,omitempty"`
	ScreenshotQRCodeHotkey                 string               `json:"screenshotQRCodeHotkey"`
	ScreenshotQRCodeHotkeyEnabled          *bool                `json:"screenshotQRCodeHotkeyEnabled,omitempty"`
	ScreenshotAllScreensHotkey             string               `json:"screenshotAllScreensHotkey"`
	ScreenshotAllScreensHotkeyEnabled      *bool                `json:"screenshotAllScreensHotkeyEnabled,omitempty"`
	ScreenshotScrollingHotkey              string               `json:"screenshotScrollingHotkey,omitempty"`
	ScreenshotScrollingHotkeyEnabled       *bool                `json:"screenshotScrollingHotkeyEnabled,omitempty"`
	ScreenshotActiveWindowHotkey           string               `json:"screenshotActiveWindowHotkey"`
	ScreenshotActiveWindowHotkeyEnabled    *bool                `json:"screenshotActiveWindowHotkeyEnabled,omitempty"`
	ScreenshotHotkeysInitialized           bool                 `json:"screenshotHotkeysInitialized,omitempty"`
	ScreenshotCaptureControls              *bool                `json:"screenshotCaptureControlsEnabled,omitempty"`
	ScreenshotCaptureCursor                *bool                `json:"screenshotCaptureCursorEnabled,omitempty"`
	ScreenshotSelectionConfirm             *bool                `json:"screenshotSelectionConfirmEnabled,omitempty"`
	ScreenshotPreviewEnabled               *bool                `json:"screenshotPreviewEnabled,omitempty"`
	ScreenshotSoundEnabled                 *bool                `json:"screenshotSoundEnabled,omitempty"`
	ScreenshotSaveFormat                   string               `json:"screenshotSaveFormat,omitempty"`
	ScreenshotHistoryDisplayLimit          int                  `json:"screenshotHistoryDisplayLimit,omitempty"`
	ScreenshotAnnotationLineWidth          int                  `json:"screenshotAnnotationLineWidth,omitempty"`
	ScreenshotCornerRadius                 int                  `json:"screenshotCornerRadius"`
	SearchCategoryShortcuts                []string             `json:"searchCategoryShortcuts"`
	DefaultSearchCategory                  string               `json:"defaultSearchCategory"`
	SearchScope                            string               `json:"searchScope,omitempty"`
	SearchScopeCycle                       []string             `json:"searchScopeCycle,omitempty"`
	SearchScopeCycleDefault                string               `json:"searchScopeCycleDefault,omitempty"`
	MatchCase                              bool                 `json:"matchCase"`
	PinyinSearchEnabled                    *bool                `json:"pinyinSearchEnabled,omitempty"`
	SearchHistory                          []string             `json:"searchHistory,omitempty"`
	SearchCategoryOrder                    []string             `json:"searchCategoryOrder"`
	NavigationVisibleModules               []string             `json:"navigationVisibleModules"`
	FeatureModules                         map[string]bool      `json:"featureModules,omitempty"`
	BookmarkFavoritePath                   BookmarkFavoritePath `json:"bookmarkFavoritePath"`
	LinkBrowsers                           []LinkBrowser        `json:"linkBrowsers,omitempty"`
	DefaultLinkBrowserID                   string               `json:"defaultLinkBrowserId,omitempty"`
	GlobalTags                             []string             `json:"globalTags"`
	TagCatalog                             []TagCatalogItem     `json:"tagCatalog,omitempty"`
	SpeedDialTileSize                      int                  `json:"speedDialTileSize,omitempty"`
	SpeedDialSortMode                      string               `json:"speedDialSortMode,omitempty"`
	SpeedDialCategoryOrder                 []string             `json:"speedDialCategoryOrder,omitempty"`
	SpeedDialExpandedGroups                []string             `json:"speedDialExpandedGroups,omitempty"`
	AppSortMode                            string               `json:"appSortMode,omitempty"`
	AppRecentLimit                         int                  `json:"appRecentLimit,omitempty"`
	AppCategoryOrder                       []string             `json:"appCategoryOrder,omitempty"`
	AppExpandedGroups                      []string             `json:"appExpandedGroups,omitempty"`
	DynamicStartMenuLaunchHistory          []StartMenuLaunchLog `json:"dynamicStartMenuLaunchHistory,omitempty"`
	DragLaunchAppIDs                       []string             `json:"dragLaunchAppIds,omitempty"`
	ConsoleItems                           []ConsoleItem        `json:"consoleItems,omitempty"`
	LegacyDragLaunchRules                  []DragLaunchRule     `json:"dragLaunchRules,omitempty"`
}

type StorageConfig struct {
	DataRoot      string `json:"dataRoot,omitempty"`
	IconDir       string `json:"iconDir,omitempty"`
	IndexDir      string `json:"indexDir,omitempty"`
	ScreenshotDir string `json:"screenshotDir,omitempty"`
	WebView2Dir   string `json:"webview2Dir,omitempty"`
}

type TagCatalogItem struct {
	Name     string `json:"name"`
	Icon     string `json:"icon,omitempty"`
	IconData string `json:"iconData,omitempty"`
	IconRef  string `json:"iconRef,omitempty"`
	IconURL  string `json:"iconUrl,omitempty"`
}

type stagedIcon struct {
	slot        launcherConfigIconSlot
	ref         string
	contentType string
	payload     string
	dataURL     string
}

type stagedWorkspaceData struct {
	sourcePath           string
	targetPath           string
	stagingPath          string
	manifest             map[string]workspaceDataManifestEntry
	sourceInspection     workspacePathInspection
	targetInspection     workspacePathInspection
	targetParentIdentity workspacePathIdentity
	stagingIdentity      workspacePathIdentity
	committedIdentity    workspacePathIdentity
	committed            bool
	targetWasEmpty       bool
}
