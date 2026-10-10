// AUTO-RECONSTRUCTED TYPES — DOMAIN: screenshot
// 研究用途
package main

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/w32"
	"image"
	"io"
	"sync"
	"time"
	"unsafe"
)

type ScreenshotCaptureResult struct {
	ImageData    string `json:"imageData"`
	ImageURL     string `json:"imageUrl,omitempty"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Mode         string `json:"mode"`
	Path         string `json:"path"`
	Cancelled    bool   `json:"cancelled"`
}

type ScreenshotHistoryEntry struct {
	ID         string `json:"id"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Mode       string `json:"mode"`
	Path       string `json:"path"`
	CapturedAt int64  `json:"capturedAt"`
}

type ScreenshotImageRef struct {
	ImageURL  string `json:"imageUrl,omitempty"`
	Path      string `json:"path,omitempty"`
	ImageData string `json:"imageData,omitempty"`
}

type ScreenshotPinState struct {
	WindowName   string `json:"windowName"`
	Visible      bool   `json:"visible"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	ImageData    string `json:"imageData,omitempty"`
	ImageURL     string `json:"imageUrl,omitempty"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
	Scale        string `json:"scale"`
	Opacity      string `json:"opacity"`
	ClickThrough bool   `json:"clickThrough"`
	AutoShow     bool   `json:"autoShow"`
	Order        int64  `json:"order"`
}

type screenshotAccessibleVtbl struct {
	QueryInterface         uintptr
	AddRef                 uintptr
	Release                uintptr
	GetTypeInfoCount       uintptr
	GetTypeInfo            uintptr
	GetIDsOfNames          uintptr
	Invoke                 uintptr
	GetAccParent           uintptr
	GetAccChildCount       uintptr
	GetAccChild            uintptr
	GetAccName             uintptr
	GetAccValue            uintptr
	GetAccDescription      uintptr
	GetAccRole             uintptr
	GetAccState            uintptr
	GetAccHelp             uintptr
	GetAccHelpTopic        uintptr
	GetAccKeyboardShortcut uintptr
	GetAccFocus            uintptr
	GetAccSelection        uintptr
	GetAccDefaultAction    uintptr
	AccSelect              uintptr
	AccLocation            uintptr
	AccNavigate            uintptr
	AccHitTest             uintptr
	AccDoDefaultAction     uintptr
	PutAccName             uintptr
	PutAccValue            uintptr
}

type screenshotCursorInfo struct {
	Size      uint32
	Flags     uint32
	Cursor    uintptr
	ScreenPos w32.POINT
}

type screenshotDisplayCaptureInfo struct {
	AdapterIndex          int
	OutputIndex           int
	AdapterName           string
	DeviceName            string
	Bounds                image.Rectangle
	Rotation              uint32
	AttachedToDesktop     bool
	BitsPerColor          uint32
	ColorSpace            uint32
	HDR                   bool
	AdvancedColor         bool
	MinLuminance          float32
	MaxLuminance          float32
	MaxFullFrameLuminance float32
}

type screenshotGDIScreenCaptureBackend struct{}

type screenshotHDRToneMapOptions struct {
	SDRWhiteLevel    float32
	SDRWhiteOutput   float32
	MaxInputLevel    float32
	HighlightRolloff float32
}

type screenshotOleVariant struct {
	VT        uint16
	Reserved1 uint16
	Reserved2 uint16
	Reserved3 uint16
	Val       int64
}

type screenshotPinnedWindowSnapshot struct {
	WindowName   string `json:"windowName"`
	ImageData    string `json:"imageData,omitempty"`
	ImageURL     string `json:"imageUrl,omitempty"`
	Path         string `json:"path,omitempty"`
	ContentKey   string `json:"contentKey,omitempty"`
	X            int    `json:"x"`
	Y            int    `json:"y"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Scale        string `json:"scale"`
	Opacity      string `json:"opacity"`
	ClickThrough bool   `json:"clickThrough"`
	AutoShow     bool   `json:"autoShow"`
	Order        int64  `json:"order"`
}

type screenshotPreviewWindowUpdate struct {
	Version   int64  `json:"version"`
	ImageURL  string `json:"imageUrl,omitempty"`
	ImageData string `json:"imageData,omitempty"`
	ReadyURL  string `json:"readyURL"`
	OpenURL   string `json:"openURL,omitempty"`
}

type screenshotRouteCapability struct {
	Token      string
	Owner      string
	Generation int64
}

type screenshotSelectionToolbarAction struct {
	SessionID int64  `json:"sessionId"`
	Action    string `json:"action"`
	Tool      string `json:"tool,omitempty"`
	Color     string `json:"color,omitempty"`
	LineWidth int    `json:"lineWidth,omitempty"`
	Preview   bool   `json:"preview,omitempty"`
	Open      bool   `json:"open,omitempty"`
}

type screenshotSelectionToolbarMessages struct {
	LanguageCode   string `json:"languageCode"`
	ResolvedTheme  string `json:"resolvedTheme"`
	UIScalePercent int    `json:"uiScalePercent"`
	ToolbarLabel   string `json:"toolbarLabel"`
	Pointer        string `json:"pointer"`
	Delete         string `json:"delete"`
	Rect           string `json:"rect"`
	Ellipse        string `json:"ellipse"`
	Line           string `json:"line"`
	Arrow          string `json:"arrow"`
	Pen            string `json:"pen"`
	Text           string `json:"text"`
	Number         string `json:"number"`
	Mosaic         string `json:"mosaic"`
	Blur           string `json:"blur"`
	DragToolbar    string `json:"dragToolbar"`
	Color          string `json:"color"`
	CustomColor    string `json:"customColor"`
	Eyedropper     string `json:"eyedropper"`
	ConfirmColor   string `json:"confirmColor"`
	LineWidth      string `json:"lineWidth"`
	CornerRadius   string `json:"cornerRadius"`
	Undo           string `json:"undo"`
	Redo           string `json:"redo"`
	Clear          string `json:"clear"`
	Confirm        string `json:"confirm"`
	Cancel         string `json:"cancel"`
	Processing     string `json:"processing"`
}

type screenshotSelectionToolbarState struct {
	SessionID       int64  `json:"sessionId,string"`
	Tool            string `json:"tool"`
	Color           string `json:"color"`
	LineWidth       int    `json:"lineWidth"`
	CanUndo         bool   `json:"canUndo"`
	CanRedo         bool   `json:"canRedo"`
	CanClear        bool   `json:"canClear"`
	ColorPickerOpen bool   `json:"colorPickerOpen"`
	Processing      bool   `json:"processing"`
}

type screenshotUIAutomationVtbl struct {
	QueryInterface              uintptr
	AddRef                      uintptr
	Release                     uintptr
	CompareElements             uintptr
	CompareRuntimeIds           uintptr
	GetRootElement              uintptr
	ElementFromHandle           uintptr
	ElementFromPoint            uintptr
	GetFocusedElement           uintptr
	GetRootElementBuildCache    uintptr
	ElementFromHandleBuildCache uintptr
	ElementFromPointBuildCache  uintptr
	GetFocusedElementBuildCache uintptr
	CreateTreeWalker            uintptr
	ControlViewWalker           uintptr
	ContentViewWalker           uintptr
	RawViewWalker               uintptr
	RawViewCondition            uintptr
	ControlViewCondition        uintptr
	ContentViewCondition        uintptr
	CreateCacheRequest          uintptr
	CreateTrueCondition         uintptr
	CreateFalseCondition        uintptr
	CreatePropertyCondition     uintptr
}

type screenshotAccessible struct {
	lpVtbl *screenshotAccessibleVtbl
}

type screenshotCOMQueryEvent struct {
	requestID        uint64
	workerGeneration uint64
	started          bool
	result           screenshotCOMQueryResult
}

type screenshotCOMQueryPayload struct {
	point        gpuPreferenceScreenPoint
	targetWindow uintptr
}

type screenshotCOMQueryRequest struct {
	id               uint64
	priority         uint8
	payload          screenshotCOMQueryPayload
	busyError        error
	state            uint8
	workerGeneration uint64
	events           chan screenshotCOMQueryEvent
}

type screenshotCOMQueryResult struct {
	rect        image.Rectangle
	controlType uint32
	err         error
}

type screenshotCOMQueryWorker struct {
	generation uint64
	done       chan struct{}
	poisoned   bool
	request    *screenshotCOMQueryRequest
}

type screenshotCOMQueryWorkerPool struct {
	mu                  sync.Mutex
	name                string
	maxWorkers          int
	maxPendingFinal     int
	timeoutThreshold    int
	cooldown            int64
	pumpInterval        int64
	initializeThread    func() (func(), error)
	pumpMessages        func() bool
	execute             func(screenshotCOMQueryPayload) screenshotCOMQueryResult
	now                 func() time.Time
	wake                chan struct{}
	nextRequestID       uint64
	nextGeneration      uint64
	currentGeneration   uint64
	workers             map[uint64]*screenshotCOMQueryWorker
	pendingFinal        []*screenshotCOMQueryRequest
	pendingHover        *screenshotCOMQueryRequest
	consecutiveTimeouts int
	cooldownUntil       time.Time
	shuttingDown        bool
}

type screenshotCOMQueryWorkerPoolConfig struct {
	name               string
	maxWorkers         int
	maxPendingFinal    int
	timeoutThreshold   int
	cooldown           int64
	pumpInterval       int64
	initializeThread   func() (func(), error)
	pumpThreadMessages func() bool
	execute            func(screenshotCOMQueryPayload) screenshotCOMQueryResult
	now                func() time.Time
}

type screenshotCaptureAcceptedNotifier struct {
	service *BootstrapService
	once    sync.Once
}

type screenshotCursorSnapshot struct {
	info screenshotCursorInfo
	ok   bool
}

type screenshotD3D11Device struct {
	device       unsafe.Pointer
	context      unsafe.Pointer
	featureLevel uint32
}

type screenshotDXGIHDRScreenCaptureBackend struct {
	displays []screenshotDisplayCaptureInfo
	fallback screenshotGDIScreenCaptureBackend
}

type screenshotDXGIOutputCaptureEntry struct {
	mu             sync.Mutex
	key            string
	displayLabel   string
	device         *screenshotD3D11Device
	duplication    unsafe.Pointer
	toneMapOptions screenshotHDRToneMapOptions
	lastFrame      *image.RGBA
	lastFrameFree  func()
	lastUsed       time.Time
	closed         bool
}

type screenshotImageMemoryBudget struct {
	mu    sync.Mutex
	limit int64
	used  int64
}

// screenshotImageConfig 是截图图像解码/编码的内存与尺寸校验配置。
// [S rodata 0x141965500 默认 / 0x141965540 scrolling] 五个 int64 字段，
// validateScreenshotImageConfig 按字段序从栈上读取。
type screenshotImageConfig struct {
	maxBytes   int64
	maxWidth   int64
	maxHeight  int64
	maxPixels  int64
	maxWorkSet int64
}

type screenshotInput struct {
	inputType uint32
	mouse     screenshotMouseInput
}

type screenshotMouseInput struct {
	dx        int32
	dy        int32
	mouseData uint32
	flags     uint32
	time      uint32
	extraInfo uintptr
}

type screenshotNativePinWindow struct {
	service       any
	pin           any
	pinName       string
	pinGeneration uint64
	image         *image.RGBA
	imageRelease  func()
	ready         chan error
	lock          sync.Mutex
	hwnd          uintptr
	threadID      uintptr
	bounds        application.Rect
	visible       bool
	opacity       float64
	clickThrough  bool
	hovered       bool
	panelVisible  bool
	closePressed  bool
	opacityDrag   bool
	highlightTick int
	closeOnce     sync.Once
}

type screenshotNativePreviewWindow struct {
	service   *screenshotPreviewWindowService
	ready     chan error
	lock      sync.Mutex
	hwnd      uintptr
	threadID  uintptr
	bounds    application.Rect
	payload   ScreenshotCaptureResult
	version   int64
	image     *image.RGBA
	opacity   float64
	visible   bool
	themeDark bool
	closeOnce sync.Once
}

type screenshotPNGIDATChunkWriter struct {
	writer io.Writer
	buffer []uint8
	limit  int
	closed bool
}

type screenshotPNGRowReleaser interface {
	ReleaseRow(int)
}

type screenshotPNGRowSource interface {
	Bounds() image.Rectangle
	Row(int) ([]uint8, error)
}

type screenshotPinWindowService struct {
	lock              sync.Mutex
	app               *application.App
	assets            *launcherAssetService
	root              string
	windows           map[string]*screenshotPinnedWindow
	snapshots         map[string]screenshotPinnedWindowSnapshot
	contentWindows    map[string]string
	lastWindowName    string
	shutting          bool
	skipPersist       bool
	restoreSuppressed bool
	shutdownPersisted bool
	nextOrder         int64
	nextGeneration    uint64
	restoreOnce       sync.Once
}

type screenshotPinnedWindow struct {
	lock         sync.RWMutex
	generation   uint64
	stateVersion uint64
	closed       bool
	contentKey   string
	name         string
	imageData    string
	imageURL     string
	thumbnailURL string
	sourcePath   string
	window       application.Window
	nativeWindow *screenshotNativePinWindow
	imageWidth   int
	imageHeight  int
	scale        float64
	opacity      float64
	clickThrough bool
	autoShow     bool
	order        int64
}

type screenshotPinnedWindowView struct {
	generation   uint64
	stateVersion uint64
	closed       bool
	contentKey   string
	name         string
	imageData    string
	imageURL     string
	thumbnailURL string
	sourcePath   string
	window       application.Window
	nativeWindow *screenshotNativePinWindow
	imageWidth   int
	imageHeight  int
	scale        float64
	opacity      float64
	clickThrough bool
	autoShow     bool
	order        int64
}

type screenshotPreviewWindowService struct {
	createMu     sync.Mutex
	lock         sync.Mutex
	app          *application.App
	window       application.Window
	nativeWindow *screenshotNativePreviewWindow
	payload      ScreenshotCaptureResult
	version      int64
	shutting     bool
	hidden       bool
	bounds       application.Rect
	ready        chan struct{}
	assets       *launcherAssetService
	capability   screenshotRouteCapability
}

// screenshotRGBAImageRowSource 与 image.RGBA 同布局（Pix@0 / Stride@0x18 / Rect@0x20），
// 是自定义 Bounds/Row 方法的行源类型（asm 0x1409946e0/0x140994720 实证偏移）。
type screenshotRGBAImageRowSource image.RGBA

type screenshotScreenCaptureBackend interface {
	captureScreenRect(image.Rectangle, screenshotScreenCaptureOptions) (*image.RGBA, error)
	captureVirtualScreen(screenshotScreenCaptureOptions) qrCodeScreenSnapshot
	name() string
}

type screenshotScreenCaptureOptions struct {
	captureCursor         bool
	cursorSnapshot        screenshotCursorSnapshot
	requireFreshDXGIFrame bool
}

type screenshotScrollingAppendCandidate struct {
	topSkip int
	overlap int
	score   int64
}

type screenshotScrollingCanvasChunk struct {
	image   *image.RGBA
	release func()
}

type screenshotScrollingCaptureOutput struct {
	pngData          []uint8
	cancelled        bool
	truncated        bool
	truncationReason string
}

type screenshotScrollingChunkedCanvas struct {
	width  int
	height int
	chunks []*screenshotScrollingCanvasChunk
}

type screenshotScrollingScrollController struct {
	total int64
	step  int64
}

type screenshotSelectionOverlayHitTestSession interface {
	setHitTestTransparent(bool)
}

type screenshotSelectionToolbarWindowService struct {
	lifecycle                  sync.Mutex
	lock                       sync.Mutex
	app                        *application.App
	window                     application.Window
	sessionID                  int64
	state                      screenshotSelectionToolbarState
	messages                   screenshotSelectionToolbarMessages
	actions                    chan screenshotSelectionToolbarAction
	terminalActions            chan screenshotSelectionToolbarAction
	notify                     func(bool) bool
	version                    int64
	capability                 screenshotRouteCapability
	removeConfirmEventListener func()
	removeCancelEventListener  func()
	shutting                   bool
	lastKeepVisibleAt          time.Time
}

type screenshotUIAutomation struct {
	lpVtbl *screenshotUIAutomationVtbl
}

type screenshotWindowCaptureAcceptedNotifier struct {
	callback func()
	once     sync.Once
}

type screenshotWindowSelectionSession struct {
	snapshot           qrCodeScreenSnapshot
	shadedBitmap       uintptr
	shadedDC           uintptr
	shadedPrevious     uintptr
	borderBrush        uintptr
	cancelled          bool
	resultPNG          []uint8
	err                error
	sourceProcessName  string
	selectedProcess    WindowProcessPickResult
	captureControls    bool
	processPickOnly    bool
	excludedPID        uint32
	accepted           *screenshotWindowCaptureAcceptedNotifier
	hitTestTransparent bool
	rightCancelPending bool
	pendingCaptureRect image.Rectangle
	hoverRect          image.Rectangle
	hoverControlRect   image.Rectangle
	hoverControlPoint  image.Point
	hoverControlAt     time.Time
	hoverProcess       WindowProcessPickResult
	hoverProcessHwnd   uintptr
	stateMutex         sync.Mutex
	hwnd               uintptr
	timeoutTriggered   bool
}
