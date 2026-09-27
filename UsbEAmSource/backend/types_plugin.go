// AUTO-RECONSTRUCTED TYPES — DOMAIN: plugin
// 研究用途
package main

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"sync"
)

type PluginEnvironment struct {
	PluginID         string                 `json:"pluginId"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	APIVersion       string                 `json:"apiVersion"`
	ThemeMode        string                 `json:"themeMode"`
	ResolvedTheme    string                 `json:"resolvedTheme"`
	Language         string                 `json:"language"`
	FallbackLanguage string                 `json:"fallbackLanguage"`
	UIScalePercent   int                    `json:"uiScalePercent"`
	Permissions      []string               `json:"permissions"`
	UI               PluginUIConfig         `json:"ui"`
	Messages         map[string]interface{} `json:"messages,omitempty"`
}

type PluginInstallResult struct {
	State     LauncherState            `json:"state"`
	Catalog   PluginUpdateCatalogState `json:"catalog"`
	Installed []string                 `json:"installed,omitempty"`
	Updated   []string                 `json:"updated,omitempty"`
	Skipped   []string                 `json:"skipped,omitempty"`
}

type PluginLocalization struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Messages    map[string]interface{} `json:"messages,omitempty"`
}

type PluginManifest struct {
	ID              string                        `json:"id"`
	Name            string                        `json:"name"`
	Description     string                        `json:"description"`
	Version         string                        `json:"version"`
	MinAppVersion   string                        `json:"minAppVersion,omitempty"`
	Entry           string                        `json:"entry"`
	Category        string                        `json:"category"`
	APIVersion      string                        `json:"apiVersion,omitempty"`
	Icon            string                        `json:"icon,omitempty"`
	IconURL         string                        `json:"iconUrl,omitempty"`
	UI              PluginUIConfig                `json:"ui,omitempty"`
	Permissions     []string                      `json:"permissions,omitempty"`
	I18N            map[string]PluginLocalization `json:"i18n,omitempty"`
	SourceDir       string                        `json:"sourceDir"`
	Installed       bool                          `json:"installed"`
	RemoteAvailable bool                          `json:"remoteAvailable,omitempty"`
	RemoteVersion   string                        `json:"remoteVersion,omitempty"`
	PackageFile     string                        `json:"packageFile,omitempty"`
	PackageURL      string                        `json:"packageUrl,omitempty"`
	PackageSHA256   string                        `json:"packageSha256,omitempty"`
	PackageSize     int64                         `json:"packageSize,omitempty"`
	UpdateAvailable bool                          `json:"updateAvailable,omitempty"`
}

type PluginRemoteCatalog struct {
	SchemaVersion int                        `json:"schemaVersion"`
	GeneratedAt   string                     `json:"generatedAt,omitempty"`
	BaseURL       string                     `json:"baseUrl,omitempty"`
	Plugins       []PluginRemoteCatalogEntry `json:"plugins"`
	CDNPurgeURLs  []string                   `json:"cdnPurgeUrls,omitempty"`
}

type PluginRemoteCatalogEntry struct {
	ID            string                        `json:"id"`
	Name          string                        `json:"name,omitempty"`
	Description   string                        `json:"description,omitempty"`
	Version       string                        `json:"version,omitempty"`
	MinAppVersion string                        `json:"minAppVersion,omitempty"`
	Entry         string                        `json:"entry,omitempty"`
	Category      string                        `json:"category,omitempty"`
	APIVersion    string                        `json:"apiVersion,omitempty"`
	Icon          string                        `json:"icon,omitempty"`
	UI            PluginUIConfig                `json:"ui,omitempty"`
	Permissions   []string                      `json:"permissions,omitempty"`
	I18N          map[string]PluginLocalization `json:"i18n,omitempty"`
	PackageFile   string                        `json:"package,omitempty"`
	PackageURL    string                        `json:"packageUrl,omitempty"`
	PackageSHA256 string                        `json:"sha256,omitempty"`
	PackageSize   int64                         `json:"size,omitempty"`
}

type PluginScreenInfo struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Index     int     `json:"index"`
	X         int     `json:"x"`
	Y         int     `json:"y"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	IsPrimary bool    `json:"isPrimary"`
	Scale     float32 `json:"scale"`
}

type PluginUIConfig struct {
	Theme       string `json:"theme,omitempty"`
	DefaultMode string `json:"defaultMode,omitempty"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	MinWidth    int    `json:"minWidth,omitempty"`
	MinHeight   int    `json:"minHeight,omitempty"`
	Resizable   *bool  `json:"resizable,omitempty"`
	Frameless   bool   `json:"frameless,omitempty"`
	Fullscreen  bool   `json:"fullscreen,omitempty"`
	FluidWidth  bool   `json:"fluidWidth,omitempty"`
}

type PluginUpdateCatalogState struct {
	CatalogURL       string           `json:"catalogUrl"`
	BaseURL          string           `json:"baseUrl"`
	GeneratedAt      string           `json:"generatedAt,omitempty"`
	CheckedAt        string           `json:"checkedAt"`
	Plugins          []PluginManifest `json:"plugins"`
	UpdateCount      int              `json:"updateCount"`
	InstallableCount int              `json:"installableCount"`
}

type PluginWindowOpenRequest struct {
	Name         string `json:"name,omitempty"`
	Entry        string `json:"entry,omitempty"`
	Title        string `json:"title,omitempty"`
	Mode         string `json:"mode,omitempty"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	MinWidth     int    `json:"minWidth,omitempty"`
	MinHeight    int    `json:"minHeight,omitempty"`
	Resizable    *bool  `json:"resizable,omitempty"`
	Frameless    *bool  `json:"frameless,omitempty"`
	Fullscreen   *bool  `json:"fullscreen,omitempty"`
	AlwaysOnTop  bool   `json:"alwaysOnTop,omitempty"`
	HideOnEscape *bool  `json:"hideOnEscape,omitempty"`
	ScreenID     string `json:"screenId,omitempty"`
	Query        string `json:"query,omitempty"`
}

type PluginWindowState struct {
	Name        string `json:"name"`
	PluginID    string `json:"pluginId"`
	Entry       string `json:"entry"`
	Title       string `json:"title"`
	Mode        string `json:"mode"`
	Visible     bool   `json:"visible"`
	Focused     bool   `json:"focused"`
	Frameless   bool   `json:"frameless"`
	Fullscreen  bool   `json:"fullscreen"`
	AlwaysOnTop bool   `json:"alwaysOnTop"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	ScreenID    string `json:"screenId,omitempty"`
	URL         string `json:"url"`
}

type pluginRuntimeMeta struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Category    string                 `json:"category"`
	APIVersion  string                 `json:"apiVersion"`
	UI          PluginUIConfig         `json:"ui"`
	Permissions []string               `json:"permissions"`
	Messages    map[string]interface{} `json:"messages,omitempty"`
}

type PluginHostService struct {
	bootstrap *BootstrapService
}

type pluginManagedWindow struct {
	name        string
	logicalName string
	owner       string
	pluginID    string
	entry       string
	mode        string
	url         string
	title       string
	frameless   bool
	fullscreen  bool
	alwaysOnTop bool
	window      application.Window
}

type pluginWindowReservation struct {
	name       string
	owner      string
	pluginID   string
	generation uint64
}

type pluginWindowService struct {
	lock       sync.Mutex
	app        *application.App
	windows    map[string]*pluginManagedWindow
	creating   map[string]*pluginWindowReservation
	generation uint64
	shutting   bool
}
