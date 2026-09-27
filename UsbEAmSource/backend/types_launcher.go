// AUTO-RECONSTRUCTED TYPES — DOMAIN: launcher
// 研究用途
package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/wailsapp/wails/v3/pkg/application"
	"os"
	"sync"
	"time"
)

type BootstrapSnapshot struct {
	Workspace WorkspaceLayout    `json:"workspace"`
	Languages []LanguageManifest `json:"languages"`
	Plugins   []PluginManifest   `json:"plugins"`
}

type LanguageManifest struct {
	Code       string                 `json:"code"`
	Name       string                 `json:"name"`
	NativeName string                 `json:"nativeName"`
	SourceFile string                 `json:"sourceFile"`
	Messages   map[string]interface{} `json:"messages,omitempty"`
}

type LauncherAdvertisementConfig struct {
	Enabled       bool   `json:"enabled"`
	ImageURL      string `json:"img"`
	ImageAssetURL string `json:"imageAssetUrl,omitempty"`
	URL           string `json:"url"`
	Seconds       int    `json:"sec"`
}

type LauncherBackgroundMetrics struct {
	Supported bool `json:"supported"`
	VirtualX  int  `json:"virtualX"`
	VirtualY  int  `json:"virtualY"`
	Width     int  `json:"width"`
	Height    int  `json:"height"`
}

type LauncherBackgroundSelectionResult struct {
	Preference BackgroundPreference `json:"preference"`
	FileName   string               `json:"fileName,omitempty"`
}

type LauncherConfig struct {
	Revision         string                 `json:"revision,omitempty"`
	Initialized      bool                   `json:"initialized,omitempty"`
	Version          int                    `json:"version"`
	Storage          StorageConfig          `json:"storage,omitempty"`
	Preferences      Preferences            `json:"preferences"`
	Apps             []AppEntry             `json:"apps"`
	SpeedDial        []LinkEntry            `json:"speedDial"`
	Bookmarks        BookmarkSection        `json:"bookmarks"`
	FileSearch       FileSearchConfig       `json:"fileSearch"`
	FileLocator      FileLocatorConfig      `json:"fileLocator"`
	Files            []FileEntry            `json:"files"`
	MemoryRelease    MemoryReleaseConfig    `json:"memoryRelease"`
	OLEDBlackout     OLEDBlackoutConfig     `json:"oledBlackout"`
	WindowManagement WindowManagementConfig `json:"windowManagement"`
	MouseGestures    MouseGestureConfig     `json:"mouseGestures"`
	TwoFactor        TwoFactorConfig        `json:"twoFactor"`
}

type LauncherConfigOptions struct {
	DefaultTagCatalog []TagCatalogItem
}

type LauncherIconResource struct {
	IconData string `json:"iconData,omitempty"`
	IconRef  string `json:"iconRef,omitempty"`
	IconURL  string `json:"iconUrl,omitempty"`
}

type LauncherLatestVersionState struct {
	Version       string                      `json:"version"`
	SourceURL     string                      `json:"sourceUrl"`
	CheckedAt     string                      `json:"checkedAt"`
	ReleaseURL    string                      `json:"releaseUrl,omitempty"`
	Update        LauncherUpdatePackageConfig `json:"update"`
	Advertisement LauncherAdvertisementConfig `json:"advertisement"`
}

type LauncherState struct {
	Workspace                WorkspaceLayout    `json:"workspace"`
	Languages                []LanguageManifest `json:"languages"`
	Plugins                  []PluginManifest   `json:"plugins"`
	Config                   LauncherConfig     `json:"config"`
	StartupTrayMode          bool               `json:"startupTrayMode"`
	ConfigOnly               bool               `json:"configOnly,omitempty"`
	ContentRuntimeSyncNeeded bool               `json:"contentRuntimeSyncNeeded,omitempty"`
	HotkeyRegistrationErrors map[string]string  `json:"hotkeyRegistrationErrors,omitempty"`
}

type LauncherUpdateInstallResult struct {
	Started bool   `json:"started"`
	Version string `json:"version"`
}

type LauncherUpdatePackageConfig struct {
	Available  bool   `json:"available"`
	PackageURL string `json:"packageUrl,omitempty"`
	SHA256     string `json:"sha256,omitempty"`
	Size       int64  `json:"size,omitempty"`
	Root       string `json:"root,omitempty"`
}

type LauncherUpdateProgressState struct {
	Active          bool   `json:"active"`
	Stage           string `json:"stage"`
	DownloadedBytes int64  `json:"downloadedBytes"`
	TotalBytes      int64  `json:"totalBytes"`
	Percent         int    `json:"percent"`
	Error           string `json:"error,omitempty"`
	UpdatedAt       string `json:"updatedAt"`
}

type StartupState struct {
	Workspace    WorkspaceLayout    `json:"workspace"`
	Languages    []LanguageManifest `json:"languages"`
	Plugins      []PluginManifest   `json:"plugins"`
	ConfigExists bool               `json:"configExists"`
	Initialized  bool               `json:"initialized"`
}

type launcherAssetRef struct {
	ID          string
	URL         string
	Version     int64
	ContentType string
}

type launcherConfigConflictError struct {
	Expected string
	Actual   string
}

// Error 返回冲突错误消息。
// [S 汇编 0x140898a00]：fmt.Sprintf("%s: expected=%s actual=%s",
// "CONFIG_REVISION_CONFLICT" 24B @0x140c65118, Expected, Actual)。
func (e launcherConfigConflictError) Error() string {
	return fmt.Sprintf("%s: expected=%s actual=%s", "CONFIG_REVISION_CONFLICT", e.Expected, e.Actual)
}

type launcherConfigIconAssetCacheEntry struct {
	URL      string
	Sequence uint64
}

type launcherConfigIconDataError struct {
	Path   string
	Reason string
}

type launcherConfigIconExtractionResult struct {
	Extracted int
	Added     int
	Reused    int
	Removed   int
}

type launcherConfigIconHydrationError struct {
	Path string
	Ref  string
	Err  error
}

type launcherConfigIconLibrary struct {
	Version int                                 `json:"version"`
	Icons   map[string]launcherConfigIconRecord `json:"icons"`
}

type launcherConfigIconRecord struct {
	ContentType string `json:"contentType"`
	Data        string `json:"data"`
}

type launcherConfigIconSlot struct {
	Path string
	Data *string
	Ref  *string
	URL  *string
}

// launcherHotkeyBindings 字段顺序与 ABI 展平一致（汇编 0x140791400 实证）：
// 7 组 {string, bool} 依次排列，ScreenshotFeatureEnabled 位于末位。
type launcherHotkeyBindings struct {
	SummonSearch                  string
	SummonSearchEnabled           bool
	SummonOnly                    string
	SummonOnlyEnabled             bool
	Screenshot                    string
	ScreenshotEnabled             bool
	ScreenshotQRCode              string
	ScreenshotQRCodeEnabled       bool
	ScreenshotAllScreens          string
	ScreenshotAllScreensEnabled   bool
	ScreenshotScrolling           string
	ScreenshotScrollingEnabled    bool
	ScreenshotActiveWindow        string
	ScreenshotActiveWindowEnabled bool
	ScreenshotFeatureEnabled      bool
}

type launcherHotkeyRegistrationError struct {
	Action  string
	Binding string
	Err     error
}

type launcherStartupTaskActionsXML struct {
	Context string                     `xml:"Context,attr"`
	Exec    launcherStartupTaskExecXML `xml:"Exec"`
}

type launcherStartupTaskEnabledXML struct {
	Enabled bool   `xml:"Enabled"`
	Delay   string `xml:"Delay,omitempty"`
}

type launcherStartupTaskExecXML struct {
	Command          string `xml:"Command"`
	Arguments        string `xml:"Arguments,omitempty"`
	WorkingDirectory string `xml:"WorkingDirectory"`
}

type launcherStartupTaskIdleSettingsXML struct {
	Duration      string `xml:"Duration"`
	WaitTimeout   string `xml:"WaitTimeout"`
	StopOnIdleEnd bool   `xml:"StopOnIdleEnd"`
	RestartOnIdle bool   `xml:"RestartOnIdle"`
}

type launcherStartupTaskPrincipalXML struct {
	ID        string `xml:"id,attr"`
	UserID    string `xml:"UserId"`
	LogonType string `xml:"LogonType"`
	RunLevel  string `xml:"RunLevel"`
}

type launcherStartupTaskPrincipalsXML struct {
	Principal launcherStartupTaskPrincipalXML `xml:"Principal"`
}

type launcherStartupTaskSettingsXML struct {
	MultipleInstancesPolicy    string                             `xml:"MultipleInstancesPolicy"`
	DisallowStartIfOnBatteries bool                               `xml:"DisallowStartIfOnBatteries"`
	StopIfGoingOnBatteries     bool                               `xml:"StopIfGoingOnBatteries"`
	AllowHardTerminate         bool                               `xml:"AllowHardTerminate"`
	StartWhenAvailable         bool                               `xml:"StartWhenAvailable"`
	RunOnlyIfNetworkAvailable  bool                               `xml:"RunOnlyIfNetworkAvailable"`
	IdleSettings               launcherStartupTaskIdleSettingsXML `xml:"IdleSettings"`
	AllowStartOnDemand         bool                               `xml:"AllowStartOnDemand"`
	Enabled                    bool                               `xml:"Enabled"`
	Hidden                     bool                               `xml:"Hidden"`
	RunOnlyIfIdle              bool                               `xml:"RunOnlyIfIdle"`
	WakeToRun                  bool                               `xml:"WakeToRun"`
	ExecutionTimeLimit         string                             `xml:"ExecutionTimeLimit"`
	Priority                   *int                               `xml:"Priority,omitempty"`
}

type launcherStartupTaskTriggersXML struct {
	LogonTrigger launcherStartupTaskEnabledXML `xml:"LogonTrigger"`
}

type launcherStartupTaskXML struct {
	XMLName    xml.Name                         `xml:"Task"`
	Version    string                           `xml:"version,attr"`
	Xmlns      string                           `xml:"xmlns,attr"`
	Triggers   launcherStartupTaskTriggersXML   `xml:"Triggers"`
	Principals launcherStartupTaskPrincipalsXML `xml:"Principals"`
	Settings   launcherStartupTaskSettingsXML   `xml:"Settings"`
	Actions    launcherStartupTaskActionsXML    `xml:"Actions"`
}

type launcherUpdateHealthHello struct {
	Nonce    string                        `json:"nonce"`
	Identity launcherUpdateProcessIdentity `json:"identity"`
}

type launcherUpdateHelperChallenge struct {
	ParentIdentity launcherUpdateProcessIdentity `json:"parentIdentity"`
	ChildIdentity  launcherUpdateProcessIdentity `json:"childIdentity"`
	ChildNonce     string                        `json:"childNonce"`
	ParentNonce    string                        `json:"parentNonce"`
	Plan           []uint8                       `json:"plan"`
}

type launcherUpdateHelperHello struct {
	Identity   launcherUpdateProcessIdentity `json:"identity"`
	ChildNonce string                        `json:"childNonce"`
}

type launcherUpdateHelperReady struct {
	Proof string `json:"proof"`
}

type launcherUpdatePackageRange struct {
	Start int64
	End   int64
}

type launcherUpdatePlan struct {
	SchemaVersion  int                      `json:"schemaVersion"`
	Nonce          string                   `json:"nonce"`
	HealthNonce    string                   `json:"healthNonce"`
	ParentPID      int                      `json:"parentPid"`
	InstallDir     string                   `json:"installDir"`
	UpdateRoot     string                   `json:"updateRoot"`
	UpdateDir      string                   `json:"updateDir"`
	ExecutableName string                   `json:"executableName"`
	LogFile        string                   `json:"logFile"`
	JournalFile    string                   `json:"journalFile"`
	PackageSize    int64                    `json:"packageSize"`
	PackageSHA256  string                   `json:"packageSha256"`
	PackageRoot    string                   `json:"packageRoot"`
	Files          []launcherUpdatePlanFile `json:"files"`
	Deletes        []string                 `json:"deletes,omitempty"`
}

type launcherUpdatePlanFile struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
	Mode   uint32 `json:"mode"`
}

type launcherUpdateProcessIdentity struct {
	PID           uint32 `json:"pid"`
	ParentPID     uint32 `json:"parentPid"`
	CreatedAt     int64  `json:"createdAt"`
	SessionID     uint32 `json:"sessionId"`
	UserSID       string `json:"userSid"`
	IntegrityRID  uint32 `json:"integrityRid"`
	ImagePath     string `json:"imagePath"`
	ImageSHA256   string `json:"imageSha256"`
	VolumeSerial  uint32 `json:"volumeSerial"`
	FileIndexHigh uint32 `json:"fileIndexHigh"`
	FileIndexLow  uint32 `json:"fileIndexLow"`
}

type launcherUpdateTransaction struct {
	SchemaVersion int                             `json:"schemaVersion"`
	Nonce         string                          `json:"nonce"`
	State         string                          `json:"state"`
	InstallDir    string                          `json:"installDir"`
	StagingDir    string                          `json:"stagingDir"`
	BackupDir     string                          `json:"backupDir"`
	JournalFile   string                          `json:"journalFile"`
	Executable    string                          `json:"executable"`
	HealthNonce   string                          `json:"healthNonce"`
	Files         []launcherUpdateTransactionFile `json:"files"`
	UpdatedAt     string                          `json:"updatedAt"`
}

type launcherUpdateTransactionFile struct {
	Path         string `json:"path"`
	Delete       bool   `json:"delete,omitempty"`
	Existed      bool   `json:"existed,omitempty"`
	Status       string `json:"status"`
	ExpectedSize int64  `json:"expectedSize,omitempty"`
	ExpectedHash string `json:"expectedSha256,omitempty"`
}

type launcherWindowLayoutSnapshot struct {
	Width     int
	Height    int
	RelativeX int
	RelativeY int
}

type BootstrapService struct {
	workspace                           WorkspaceLayout
	configStore                         *launcherConfigStore
	bootstrapSnapshotRefresh            sync.Mutex
	bootstrapSnapshot                   BootstrapSnapshot
	bootstrapSnapshotReady              bool
	app                                 *application.App
	launcherWindow                      application.Window
	globalHotkey                        launcherGlobalHotkeyManager
	hotkeyCapture                       bool
	hotkeyCaptureLegacyRef              int
	hotkeyCaptureOwners                 map[string]struct{}
	hotkeyCaptureOperation              sync.Mutex
	searchCategoryShortcutActive        bool
	hotkey                              string
	hotkeyEnabled                       bool
	summonOnlyHotkey                    string
	summonOnlyHotkeyEnabled             bool
	screenshotHotkey                    string
	screenshotHotkeyEnabled             bool
	screenshotQRCodeHotkey              string
	screenshotQRCodeHotkeyEnabled       bool
	screenshotAllScreensHotkey          string
	screenshotAllScreensHotkeyEnabled   bool
	screenshotScrollingHotkey           string
	screenshotScrollingHotkeyEnabled    bool
	screenshotActiveWindowHotkey        string
	screenshotActiveWindowHotkeyEnabled bool
	screenshotFeatureEnabled            bool
	screenshotHotkeyActive              bool
	screenshotSaveFormatOverride        string
	screenshotCaptureControlsSetting    bool
	screenshotSelectionConfirmSetting   bool
	screenshotCaptureCursorSetting      bool
	screenshotCaptureSettingsReady      bool
	screenshotAnnotationLineWidth       int
	screenshotCornerRadius              int
	bindings                            launcherHotkeyBindings
	hotkeyRegistrationErrors            map[string]string
	bookmarkSources                     []BookmarkSource
	twoFactor                           *twoFactorService
	memoryRelease                       *memoryReleaseService
	oledBlackout                        *oledBlackoutService
	windowManagement                    *windowManagementService
	mouseGestures                       *mouseGestureService
	assets                              *launcherAssetService
	backgroundAssetLock                 sync.Mutex
	backgroundAssetOwner                *launcherAssetService
	backgroundAssetPath                 string
	backgroundAssetSize                 int64
	backgroundAssetModifiedAt           int64
	backgroundAssetURL                  string
	iconAssetLock                       sync.Mutex
	iconAssetOwner                      *launcherAssetService
	iconAssetURLs                       map[string]launcherConfigIconAssetCacheEntry
	iconAssetSequence                   uint64
	screenshotPin                       *screenshotPinWindowService
	screenshotPreview                   *screenshotPreviewWindowService
	screenshotSelectionToolbar          *screenshotSelectionToolbarWindowService
	pluginWindows                       *pluginWindowService
	inputMonitor                        *inputMonitorService
	fileIndex                           fileSearchRuntimeWarmer
	fileLocator                         *fileLocatorService
	desktopWidgets                      *desktopWidgetService
	resetLauncherSizeOnNextShow         bool
	startupTrayMode                     bool
	pendingLauncherReveal               bool
	allowLauncherWindowClose            bool
	launcherWindowDestroyedToTray       bool
	launcherFrontendReady               bool
	launcherSidebarPinnedOpen           bool
	lastLauncherHotkeyAction            string
	lastLauncherHotkeyAt                time.Time
	launcherUpdateProgress              LauncherUpdateProgressState
	launcherUpdateCancel                func()
	launcherUpdateTaskID                int64
	launcherUpdateDir                   string
	launcherUpdateDone                  chan struct{}
	launcherUpdateShuttingDown          bool
	launcherUpdateNotificationShown     bool
	pendingQRCodeDecodeResult           *QRCodeDecodeResult
	verticalMaximizeSnapshot            *launcherWindowLayoutSnapshot
	pendingFileSearchResidentWarm       bool
	startupTaskSync                     func(bool, int) error
	workspaceTransaction                sync.Mutex
	workspaceDataMaintenance            workspaceDataMaintenanceGate
	lock                                sync.Mutex
}

type launcherAssetEntry struct {
	id          string
	namespace   string
	version     int64
	contentType string
	data        []uint8
	filePath    string
	size        int64
	createdAt   time.Time
	accessedAt  time.Time
	expiresAt   time.Time
}

type launcherAssetLimits struct {
	maxEntries        int
	maxItemBytes      int64
	maxNamespaceBytes int64
	maxTotalBytes     int64
}

type launcherAssetService struct {
	lock           sync.Mutex
	entries        map[string]launcherAssetEntry
	namespaceBytes map[string]int64
	totalBytes     int64
	next           int64
	now            func() time.Time
	limits         launcherAssetLimits
}

type launcherConfigIconMetrics struct {
	decodedBytes int
	pixels       int
}

type launcherConfigIconStore struct {
	path          string
	readLibrary   func(string) ([]uint8, error)
	writeLibrary  func(string, launcherConfigIconLibrary) error
	loaded        bool
	cachedLoadErr error
	cached        launcherConfigIconLibrary
	cachedBudget  launcherConfigIconStoreBudget
	mu            sync.Mutex
}

type launcherConfigIconStoreBudget struct {
	count        int
	decodedBytes int
	pixels       int
}

type launcherConfigJSONFrame struct {
	kind  int32
	count int
}

type launcherConfigStore struct {
	path                 string
	writeConfig          func(string, LauncherConfig) error
	runtimeReady         bool
	iconFingerprint      [32]uint8
	iconFingerprintReady bool
	mu                   sync.Mutex
}

type launcherGlobalHotkeyCommand struct {
	bindings     launcherHotkeyBindings
	closeManager bool
	result       chan error
}

type launcherGlobalHotkeyManager interface {
	Close() error
	Update(launcherHotkeyBindings) error
}

type launcherGlobalHotkeyRegistration struct {
	action     string
	binding    string
	id         int32
	modifiers  uint32
	virtualKey uint32
}

type launcherUpdateLogger struct {
	file *os.File
}

type launcherUpdateProgressWriter struct {
	ctx     context.Context
	written int64
	write   func(int64)
}

type launcherWidgetStore struct {
	path          string
	mu            sync.Mutex
	loaded        bool
	cachedExists  bool
	cached        DesktopWidgetDocument
	cachedLoadErr error
	writeDocument func(string, DesktopWidgetDocument) error
}

type launcherWindowController interface {
	Close()
	EmitEvent(string, []interface{}) bool
	Focus()
	Hide() application.Window
	IsFocused() bool
	IsMinimised() bool
	IsVisible() bool
	Position() (int, int)
	Restore()
	SetContentProtection(bool) application.Window
	SetPosition(int, int)
	Show() application.Window
}

type launcherWindowLayoutController interface {
	Close()
	EmitEvent(string, []interface{}) bool
	GetScreen() (*application.Screen, error)
	Hide() application.Window
	IsFullscreen() bool
	IsMaximised() bool
	Maximise() application.Window
	Minimise() application.Window
	RelativePosition() (int, int)
	Restore()
	SetMaxSize(int, int) application.Window
	SetMinSize(int, int) application.Window
	SetRelativePosition(int, int) application.Window
	SetSize(int, int) application.Window
	Size() (int, int)
}

type launcherWindowScreenResolver interface {
	GetScreen() (*application.Screen, error)
}

type launcherWindowSizer interface {
	GetScreen() (*application.Screen, error)
	IsFullscreen() bool
	IsMaximised() bool
	Maximise() application.Window
	RelativePosition() (int, int)
	Restore()
	SetMaxSize(int, int) application.Window
	SetMinSize(int, int) application.Window
	SetRelativePosition(int, int) application.Window
	SetSize(int, int) application.Window
	Size() (int, int)
}

type launcherWindowTrayController interface {
	Close()
	EmitEvent(string, []interface{}) bool
	Hide() application.Window
}
