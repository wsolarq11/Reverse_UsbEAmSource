// AUTO-RECONSTRUCTED TYPES — DOMAIN: misc
// 研究用途
package main

import (
	"archive/zip"
	"modernc.org/sqlite"
	"sync"
	"time"
)

type DetectedBookmarkSource struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	BrowserName string `json:"browserName"`
	Profile     string `json:"profile"`
	Path        string `json:"path"`
}

type FileMetadataHit struct {
	Path           string `json:"path"`
	Size           int64  `json:"size"`
	SizeReady      bool   `json:"sizeReady"`
	ModifiedAt     string `json:"modifiedAt"`
	ModifiedAtUnix int64  `json:"modifiedAtUnix"`
}

type IndexNode struct {
	FRN        uint64
	NameOffset uint32
	ParentIdx  int32
	ModTime    uint32
	NameLen    uint16
	Flags      uint16
}

type ResolvedBookmarkItem struct {
	ID         string `json:"id"`
	SourceID   string `json:"sourceId"`
	SourceName string `json:"sourceName"`
	SourcePath string `json:"sourcePath"`
	Browser    string `json:"browser"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Folder     string `json:"folder"`
}

type ResolvedBookmarkSource struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Browser     string `json:"browser"`
	BrowserName string `json:"browserName"`
	Profile     string `json:"profile"`
	Path        string `json:"path"`
	Enabled     bool   `json:"enabled"`
	Exists      bool   `json:"exists"`
	ItemCount   int    `json:"itemCount"`
	Error       string `json:"error"`
}

type WebView2ProcessInfo struct {
	ProcessID               int     `json:"processId"`
	Kind                    string  `json:"kind"`
	WorkingSetBytes         uint64  `json:"workingSetBytes"`
	PrivateBytes            uint64  `json:"privateBytes"`
	WorkingSetMB            float64 `json:"workingSetMB"`
	PrivateMB               float64 `json:"privateMB"`
	HasJSFlags              bool    `json:"hasJsFlags"`
	HasDisableFeatures      bool    `json:"hasDisableFeatures"`
	HasRendererProcessLimit bool    `json:"hasRendererProcessLimit"`
	CommandLine             string  `json:"commandLine,omitempty"`
}

type WebView2ProcessSnapshot struct {
	UserDataDir          string                `json:"userDataDir,omitempty"`
	Processes            []WebView2ProcessInfo `json:"processes"`
	TotalWorkingSetBytes uint64                `json:"totalWorkingSetBytes"`
	TotalPrivateBytes    uint64                `json:"totalPrivateBytes"`
	TotalWorkingSetMB    float64               `json:"totalWorkingSetMB"`
	TotalPrivateMB       float64               `json:"totalPrivateMB"`
}

type WindowProcessPickResult struct {
	Path        string  `json:"path"`
	ProcessID   uint32  `json:"processId"`
	ProcessName string  `json:"processName"`
	DisplayName string  `json:"displayName"`
	Title       string  `json:"title,omitempty"`
	HWND        uintptr `json:"hwnd,omitempty"`
	IconData    string  `json:"iconData,omitempty"`
	IconURL     string  `json:"iconUrl,omitempty"`
}

type appIconInfo struct {
	FIcon    int32
	XHotspot uint32
	YHotspot uint32
	HbmMask  uintptr
	HbmColor uintptr
}

type fileMetadataResolveTask struct {
	Path string
	Name string
	Key  string
}

type indexSearchTerm struct {
	Original       string
	Parts          [][]uint8
	IsPath         bool
	IsDrive        bool
	AllowHierarchy bool
	AllowNameMatch bool
	PinyinFuzzy    *fileSearchPinyinFuzzyMatcher
}

type platformDesktopWidgetNotifier struct{}

type backupEntry struct {
	path    string
	name    string
	date    string
	modTime time.Time
}

type deliveryEntry struct {
	key string
	at  string
}

type endpointRef struct {
	id     string
	device *audioIMMDevice
}

type hashWriter interface {
	Sum([]uint8) []uint8
	Write([]uint8) (int, error)
}

type inlineIcon struct {
	ref     string
	metrics launcherConfigIconMetrics
}

type item struct {
	key   string
	entry *screenshotDXGIOutputCaptureEntry
	used  time.Time
}

type limitedFileSearchCandidateCollector struct {
	limit    int
	items    []scoredFileSearchCandidate
	heap     fileSearchCandidateHeap
	resolver *fileSearchCandidatePathResolver
}

type limitedSortedFileSearchCandidateCollector struct {
	limit    int
	items    []scoredFileSearchCandidate
	heap     fileSearchSortedCandidateHeap
	resolver *fileSearchCandidatePathResolver
	comparer *fileSearchSortComparer
}

type loadJob struct {
	root string
}

type loadResult struct {
	root  string
	idx   *VolumeIndex
	meta  fileSearchVolumeMeta
	err   error
	exist bool
}

type pendingPathRequest struct {
	id        string
	root      string
	nodeIndex int32
	frn       uint64
}

type persistTask struct {
	root string
	idx  *VolumeIndex
}

type platformDesktopWidgetAudioPlayer struct {
	playback sync.Mutex
}

type preparedArchiveEntry struct {
	file       *zip.File
	relative   string
	targetPath string
	directory  bool
}

type readResult struct {
	content []uint8
	err     error
}

type searchTermValue struct {
	value          string
	fromPathField  bool
	allowNameMatch bool
}

type sqliteBackuper interface {
	NewBackup(string) (*sqlite.Backup, error)
}

type versionTranslation struct {
	language uint16
	codePage uint16
}

type walEntry struct {
	endUSN int64
	raw    []uint8
}

type walkState struct {
	idx      int32
	path     string
	excluded bool
}

type windowsLauncherGlobalHotkeyManager struct {
	callback             func(string)
	commands             chan launcherGlobalHotkeyCommand
	ready                chan struct{}
	threadID             uint32
	keyboardHook         uintptr
	keyboardHookCallback uintptr
	keyboardHookBindings []launcherGlobalHotkeyRegistration
	printScreenKeyDown   bool
	closeOnce            sync.Once
}

type windowsOLEDBlackoutCursorController struct {
	commands  chan oledBlackoutCursorCommand
	ready     chan struct{}
	done      chan struct{}
	call      func(bool) (int32, error)
	lock      sync.Mutex
	closed    bool
	closeErr  error
	closeOnce sync.Once
}

type windowsOLEDBlackoutHotkeyManager struct {
	callback  func(oledBlackoutHotkeyEvent)
	commands  chan oledBlackoutHotkeyCommand
	ready     chan struct{}
	done      chan struct{}
	threadID  uint32
	closeOnce sync.Once
}
