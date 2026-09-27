// AUTO-RECONSTRUCTED TYPES — DOMAIN: gesture
// 研究用途
package main

import (
	"image"
	"sync"
	"sync/atomic"
	"time"
)

type GestureAction struct {
	Kind           string   `json:"kind,omitempty"`
	Hotkey         string   `json:"hotkey,omitempty"`
	KeySequence    []string `json:"keySequence,omitempty"`
	Text           string   `json:"text,omitempty"`
	Path           string   `json:"path,omitempty"`
	Args           []string `json:"args,omitempty"`
	WorkingDir     string   `json:"workingDir,omitempty"`
	URL            string   `json:"url,omitempty"`
	WindowCommand  string   `json:"windowCommand,omitempty"`
	WindowMove     string   `json:"windowMove,omitempty"`
	ScreenID       string   `json:"screenId,omitempty"`
	ScreenIndex    int      `json:"screenIndex,omitempty"`
	ScreenName     string   `json:"screenName,omitempty"`
	LauncherAction string   `json:"launcherAction,omitempty"`
	AppID          string   `json:"appId,omitempty"`
	ProfileID      string   `json:"profileId,omitempty"`
}

type GestureAppMatch struct {
	Path                string `json:"path,omitempty"`
	ProcessName         string `json:"processName,omitempty"`
	WindowTitleContains string `json:"windowTitleContains,omitempty"`
	IconData            string `json:"iconData,omitempty"`
	IconRef             string `json:"iconRef,omitempty"`
	IconURL             string `json:"iconUrl,omitempty"`
}

type GestureAppProfile struct {
	ID                 string            `json:"id"`
	Order              int               `json:"order,omitempty"`
	Enabled            bool              `json:"enabled"`
	InheritGlobal      bool              `json:"inheritGlobal"`
	GlobalRulePriority string            `json:"globalRulePriority,omitempty"`
	Blacklisted        bool              `json:"blacklisted,omitempty"`
	Match              GestureAppMatch   `json:"match,omitempty"`
	Matches            []GestureAppMatch `json:"matches,omitempty"`
	Name               string            `json:"name,omitempty"`
	DisplayName        string            `json:"displayName,omitempty"`
	IconData           string            `json:"iconData,omitempty"`
	IconRef            string            `json:"iconRef,omitempty"`
	IconURL            string            `json:"iconUrl,omitempty"`
	Rules              []GestureRule     `json:"rules,omitempty"`
}

type GesturePattern struct {
	Button     string   `json:"button,omitempty"`
	Directions []string `json:"directions,omitempty"`
	Modifier   string   `json:"modifier,omitempty"`
}

type GestureProfile struct {
	Enabled bool          `json:"enabled,omitempty"`
	Rules   []GestureRule `json:"rules,omitempty"`
}

type GestureRule struct {
	ID      string         `json:"id"`
	Order   int            `json:"order,omitempty"`
	Enabled bool           `json:"enabled"`
	Name    string         `json:"name,omitempty"`
	Gesture GesturePattern `json:"gesture,omitempty"`
	Action  GestureAction  `json:"action,omitempty"`
}

type HotCornerConfig struct {
	Enabled               bool            `json:"enabled,omitempty"`
	TriggerSizePx         int             `json:"triggerSizePx,omitempty"`
	TriggerDelayMs        int             `json:"triggerDelayMs,omitempty"`
	CooldownMs            int             `json:"cooldownMs,omitempty"`
	AutoDisableFullscreen bool            `json:"autoDisableFullscreen,omitempty"`
	Corners               []HotCornerRule `json:"corners,omitempty"`
}

type HotCornerRule struct {
	Corner  string `json:"corner,omitempty"`
	Enabled bool   `json:"enabled"`
	Hotkey  string `json:"hotkey,omitempty"`
}

type MouseGestureActionTarget struct {
	HWND        uintptr
	ProcessID   uint32
	Path        string
	ProcessName string
	Title       string
	Activate    bool
}

type MouseGestureConfig struct {
	Enabled    bool                `json:"enabled,omitempty"`
	Settings   GestureSettings     `json:"settings,omitempty"`
	Global     GestureProfile      `json:"global,omitempty"`
	Apps       []GestureAppProfile `json:"apps,omitempty"`
	HotCorners HotCornerConfig     `json:"hotCorners,omitempty"`
}

type MouseGesturePoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type MouseGestureScreenBounds struct {
	ID     string `json:"id,omitempty"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type MouseGestureState struct {
	Loading           bool               `json:"loading"`
	Supported         bool               `json:"supported"`
	Platform          string             `json:"platform"`
	ModuleEnabled     bool               `json:"moduleEnabled"`
	RuntimeEnabled    bool               `json:"runtimeEnabled"`
	HotCornersEnabled bool               `json:"hotCornersEnabled"`
	RuntimeActive     bool               `json:"runtimeActive"`
	RuntimeWanted     bool               `json:"runtimeWanted"`
	Config            MouseGestureConfig `json:"config"`
	LastAction        string             `json:"lastAction,omitempty"`
	LastGesture       string             `json:"lastGesture,omitempty"`
	LastError         string             `json:"lastError,omitempty"`
}

type MouseGestureTarget struct {
	HWND        uintptr `json:"hwnd,omitempty"`
	ProcessID   uint32  `json:"processId,omitempty"`
	Path        string  `json:"path,omitempty"`
	ProcessName string  `json:"processName,omitempty"`
	Title       string  `json:"title,omitempty"`
}

type mouseGestureInput struct {
	Type uint32
	MI   mouseGestureMouseInput
}

type mouseGestureMouseInput struct {
	DX        int32
	DY        int32
	MouseData uint32
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

type GestureSettings struct {
	GestureButtons           []string `json:"gestureButtons"`
	StartDistancePx          int      `json:"startDistancePx"`
	StartTimeoutMs           int      `json:"startTimeoutMs"`
	StopTimeoutMs            int      `json:"stopTimeoutMs"`
	AllowDiagonal            bool     `json:"allowDiagonal,omitempty"`
	AutoDisableFullscreen    bool     `json:"autoDisableFullscreen,omitempty"`
	TargetWindowUnderPointer bool     `json:"targetWindowUnderPointer,omitempty"`
	ShowTrail                bool     `json:"showTrail,omitempty"`
	ShowGestureName          bool     `json:"showGestureName,omitempty"`
	FadeAfterExecution       bool     `json:"fadeAfterExecution,omitempty"`
	startDistancePxSet       bool
	startTimeoutMsSet        bool
	stopTimeoutMsSet         bool
}

type mouseGestureButtonReplay struct {
	service    *mouseGestureService
	generation uint64
	button     string
	sendClick  func(string) error
}

type mouseGestureEventQueue struct {
	lock  sync.Mutex
	items []mouseGestureHookEvent
	head  int
	size  int
}

type mouseGestureHookEvent struct {
	message    uint32
	x          int
	y          int
	button     string
	modifier   string
	down       bool
	up         bool
	move       bool
	wheel      bool
	injected   bool
	suppressed bool
}

type mouseGestureOverlayLabelSnapshot struct {
	label string
	image *image.RGBA
}

type mouseGestureOverlayPayload struct {
	points      []MouseGesturePoint
	button      string
	label       string
	labelDirty  bool
	bounds      image.Rectangle
	labelBounds image.Rectangle
	alpha       uint8
	scale       float64
}

type mouseGestureOverlayWindow struct {
	ready          chan error
	done           chan struct{}
	lock           sync.Mutex
	closeOnce      sync.Once
	hwnd           uintptr
	threadID       uintptr
	payload        mouseGestureOverlayPayload
	pendingPayload *mouseGestureOverlayPayload
	updatePending  bool
	redrawing      bool
	alpha          uint8
	visible        bool
	labelSnapshot  mouseGestureOverlayLabelSnapshot
}

type mouseGesturePlatformRuntime struct {
	service           *mouseGestureService
	events            *mouseGestureEventQueue
	ready             chan error
	config            MouseGestureConfig
	moduleEnabled     bool
	gestureEnabled    bool
	configReady       bool
	threadID          uint32
	mouseHook         uintptr
	hookRef           uintptr
	generation        uint64
	wake              func() error
	timeoutEpoch      atomic.Uint64
	timeoutPending    atomic.Bool
	queueFailure      atomic.Pointer[mouseGestureHookEvent]
	stopping          atomic.Bool
	targetScope       atomic.Pointer[mouseGestureTargetScope]
	targetHWNDAtPoint func(int, int, bool) uintptr
	buttonReplays     chan mouseGestureButtonReplay
	buttonReplayStop  chan struct{}
	buttonReplayDone  chan struct{}
}

type mouseGestureRuntimeSession struct {
	state             string
	button            string
	start             MouseGesturePoint
	points            []MouseGesturePoint
	startedAt         time.Time
	lastMovedAt       time.Time
	target            MouseGestureTarget
	targetResolved    bool
	suppressed        bool
	replayOnRelease   bool
	currentLabel      string
	currentCorner     string
	triggeredCorner   string
	candidateCorner   string
	candidateSince    time.Time
	lastTriggeredAt   map[string]time.Time
	screens           []MouseGestureScreenBounds
	lastScreenRefresh time.Time
	lastOverlayUpdate time.Time
	lastObservedHWND  uintptr
	overlay           *mouseGestureOverlayWindow
	labelOverlay      *mouseGestureOverlayWindow
	runtimeHandle     *mouseGesturePlatformRuntime
	generation        uint64
	timeoutTimer      *time.Timer
	timeoutDeadline   time.Time
	now               func() time.Time
	sendButtonDown    func(string) error
	sendButtonUp      func(string) error
	sendButtonClick   func(string) error
	targetHWNDAtPoint func(int, int, bool) uintptr
	targetFromHWND    func(uintptr) MouseGestureTarget
}

type mouseGestureService struct {
	lock                   sync.Mutex
	runtimeLifecycle       sync.Mutex
	config                 MouseGestureConfig
	moduleEnabled          bool
	runtimeStop            chan struct{}
	runtimeDone            chan struct{}
	runtimeMarkStopping    func()
	runtimeGeneration      uint64
	runtimeActive          bool
	runtimeActions         sync.WaitGroup
	shuttingDown           bool
	lastError              string
	lastAction             string
	lastGesture            string
	suppressUpMask         atomic.Uint32
	gestureInProgress      atomic.Bool
	capturePaused          bool
	observeGesture         atomic.Bool
	observeHotCorners      atomic.Bool
	observeGestureTargets  atomic.Bool
	launcherActionExecutor func(GestureAction) error
	windowMoveExecutor     func(GestureAction, MouseGestureActionTarget) error
	platformRuntimeSync    func()
}

type mouseGestureTargetScope struct {
	hwnd    uintptr
	allowed bool
}
