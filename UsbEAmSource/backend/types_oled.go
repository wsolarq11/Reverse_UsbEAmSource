// AUTO-RECONSTRUCTED TYPES — DOMAIN: oled
// 研究用途
package main

import (
	"context"
	"github.com/wailsapp/wails/v3/pkg/application"
	"sync"
	"time"
)

type OLEDBlackoutConfig struct {
	ExitTriggers         OLEDBlackoutExitTriggers          `json:"exitTriggers,omitempty"`
	MediaPauseExclusions []OLEDBlackoutMediaPauseExclusion `json:"mediaPauseExclusions,omitempty"`
	Profiles             []OLEDBlackoutProfileConfig       `json:"profiles,omitempty"`
}

type OLEDBlackoutExitTriggers struct {
	Escape    bool `json:"escape,omitempty"`
	MouseMove bool `json:"mouseMove,omitempty"`
	AnyKey    bool `json:"anyKey,omitempty"`
}

type OLEDBlackoutMediaPauseExclusion struct {
	Enabled     bool   `json:"enabled"`
	ProcessName string `json:"processName,omitempty"`
	Path        string `json:"path,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	IconData    string `json:"iconData,omitempty"`
	IconRef     string `json:"iconRef,omitempty"`
	IconURL     string `json:"iconUrl,omitempty"`
}

type OLEDBlackoutProfileConfig struct {
	ID            string   `json:"id"`
	ScreenIDs     []string `json:"screenIds,omitempty"`
	ScreenIndexes []int    `json:"screenIndexes,omitempty"`
	ScreenNames   []string `json:"screenNames,omitempty"`
	Hotkey        string   `json:"hotkey,omitempty"`
	IdleEnabled   bool     `json:"idleEnabled,omitempty"`
	IdleMinutes   int      `json:"idleMinutes,omitempty"`
	QuickIdle     bool     `json:"quickIdle,omitempty"`
	MediaPause    bool     `json:"mediaPause,omitempty"`
}

type OLEDBlackoutProfileState struct {
	ID                string   `json:"id"`
	ScreenIDs         []string `json:"screenIds"`
	ScreenIndexes     []int    `json:"screenIndexes,omitempty"`
	ScreenNames       []string `json:"screenNames,omitempty"`
	Hotkey            string   `json:"hotkey,omitempty"`
	IdleEnabled       bool     `json:"idleEnabled,omitempty"`
	IdleMinutes       int      `json:"idleMinutes,omitempty"`
	QuickIdle         bool     `json:"quickIdle,omitempty"`
	MediaPause        bool     `json:"mediaPause,omitempty"`
	Registered        bool     `json:"registered"`
	RegistrationError string   `json:"registrationError,omitempty"`
	MissingScreenIDs  []string `json:"missingScreenIds,omitempty"`
	Active            bool     `json:"active"`
	PartiallyActive   bool     `json:"partiallyActive,omitempty"`
}

type OLEDBlackoutScreen struct {
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

type OLEDBlackoutState struct {
	Loading         bool                       `json:"loading"`
	Supported       bool                       `json:"supported"`
	Enabled         bool                       `json:"enabled"`
	Config          OLEDBlackoutConfig         `json:"config"`
	ActiveProfileID string                     `json:"activeProfileId,omitempty"`
	CursorHidden    bool                       `json:"cursorHidden"`
	LastError       string                     `json:"lastError,omitempty"`
	DebugLogPath    string                     `json:"debugLogPath,omitempty"`
	Screens         []OLEDBlackoutScreen       `json:"screens"`
	Profiles        []OLEDBlackoutProfileState `json:"profiles"`
}

type oledBlackoutHotkeyBinding struct {
	ProfileID string
	Hotkey    string
}

type oledBlackoutHotkeyEvent struct {
	ProfileID string
	Hotkey    string
}

type oledBlackoutHotkeyUpdateResult struct {
	RegisteredProfileIDs []string
	Errors               map[string]string
}

type oledBlackoutIdlePauseTargetScreen struct {
	ID             string
	Bounds         application.Rect
	PhysicalBounds application.Rect
}

type oledBlackoutAudioSessionScanKey struct {
	topologyGeneration uint64
	sessionID          string
	processID          uint32
	systemSounds       bool
}

type oledBlackoutBrowserMediaContinuity struct {
	lock    sync.Mutex
	entries map[oledBlackoutBrowserMediaContinuityKey]string
}

type oledBlackoutBrowserMediaContinuityKey struct {
	windowHandle uintptr
	processID    uint32
	processPath  string
}

type oledBlackoutCursorCommand struct {
	hide     bool
	show     bool
	close    bool
	response chan error
}

type oledBlackoutCursorController interface {
	Close() error
	Hide() error
	Show() error
}

type oledBlackoutHotkeyCommand struct {
	bindings     []oledBlackoutHotkeyBinding
	closeManager bool
	result       chan oledBlackoutHotkeyUpdateResult
}

type oledBlackoutHotkeyManager interface {
	Close() error
	Update([]oledBlackoutHotkeyBinding) oledBlackoutHotkeyUpdateResult
}

type oledBlackoutHotkeyRegistration struct {
	profileID  string
	binding    string
	id         int32
	modifiers  uint32
	virtualKey uint32
}

type oledBlackoutLastInputInfo struct {
	cbSize uint32
	dwTime uint32
}

type oledBlackoutMonitorInfo struct {
	cbSize  uint32
	monitor oledBlackoutRect
	work    oledBlackoutRect
	flags   uint32
	device  [32]uint16
}

type oledBlackoutNativeOverlayWindow struct {
	overlay   any
	ready     chan error
	done      chan struct{}
	lock      sync.Mutex
	hwnd      uintptr
	threadID  uintptr
	bounds    application.Rect
	visible   bool
	closeOnce sync.Once
}

type oledBlackoutOverlayWindow struct {
	service      any
	screenID     string
	bounds       application.Rect
	window       application.Window
	nativeWindow *oledBlackoutNativeOverlayWindow
}

type oledBlackoutRect struct {
	left   int32
	top    int32
	right  int32
	bottom int32
}

type oledBlackoutService struct {
	lock                   sync.Mutex
	hotkeyOperationLock    sync.Mutex
	backgroundActivities   sync.WaitGroup
	browserMediaContinuity oledBlackoutBrowserMediaContinuity
	app                    *application.App
	config                 OLEDBlackoutConfig
	moduleEnabled          bool
	hotkeyCapture          bool
	hotkeyCaptureLegacyRef int
	hotkeyCaptureOwners    map[string]struct{}
	shuttingDown           bool
	lifecycleContext       context.Context
	lifecycleCancel        func()
	lifecycleGeneration    uint64
	shutdownDone           chan struct{}
	activeProfileID        string
	cursorHidden           bool
	cursorController       oledBlackoutCursorController
	lastError              string
	overlayWindows         map[string]*oledBlackoutOverlayWindow
	visibleOverlays        map[string]struct{}
	hotkeyManager          oledBlackoutHotkeyManager
	hotkeyRegistered       map[string]bool
	hotkeyErrors           map[string]string
	shouldSuppressHotkey   func(string) bool
	idleTimer              *time.Timer
	idleScheduleID         uint64
	inputPollTimer         *time.Timer
	inputPollScheduleID    uint64
	inputPollActive        bool
	focusRetryTimer        *time.Timer
	focusRetryScheduleID   uint64
	lastCursorX            int
	lastCursorY            int
	lastCursorValid        bool
	lastPressedKeys        map[uintptr]struct{}
	inputDismissGuardUntil time.Time
	autoActivatedProfileID string
	quickIdleProfileID     string
	quickIdleWakeAt        time.Time
	quickIdleActiveAt      time.Time
}

type oledBlackoutWindowProcessCandidate struct {
	hwnd         uintptr
	processID    uint32
	processPath  string
	processAUMID string
	windowTitle  string
}

// oledBlackoutIReference WinRT IReference<int32> 接口指针包装。
// asm 实证（0x140921060 GetInt32）：首字段（+0x00）为 vtable 指针，
// this 直接以接口指针传入 get_Value；get_Value 位于 vtbl[6]（+0x30）。
type oledBlackoutIReference struct {
	vtbl *uintptr
}
