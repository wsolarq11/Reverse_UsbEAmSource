// AUTO-RECONSTRUCTED TYPES — DOMAIN: workspace
// 研究用途
package main

import (
	"sync"
)

type WorkspaceLayout struct {
	Root           string `json:"root"`
	ConfigFile     string `json:"configFile"`
	LanguageDir    string `json:"languageDir"`
	AppLanguageDir string `json:"appLanguageDir,omitempty"`
	PluginDir      string `json:"pluginDir"`
	IconDir        string `json:"iconDir"`
	IndexDir       string `json:"indexDir"`
	ScreenshotDir  string `json:"screenshotDir"`
	WebView2Dir    string `json:"webview2Dir"`
	BackgroundDir  string `json:"backgroundDir"`
}

type workspaceDataManifestEntry struct {
	Size   int64
	SHA256 string
}

type workspaceDataMaintenanceGate struct {
	mu          sync.Mutex
	maintenance bool
	active      int
	idle        chan struct{}
}

type workspacePathIdentity struct {
	canonical string
	volume    uint64
	file      uint64
	valid     bool
}

type workspacePathInspection struct {
	canonical        string
	exists           bool
	identity         workspacePathIdentity
	ancestorIdentity workspacePathIdentity
}
