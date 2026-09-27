// AUTO-RECONSTRUCTED TYPES — DOMAIN: qrcode
// 研究用途
package main

import (
	"image"
	"image/color"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type QRCodeDecodeResult struct {
	ImageData                    string               `json:"imageData"`
	ImageWidth                   int                  `json:"imageWidth"`
	ImageHeight                  int                  `json:"imageHeight"`
	Entries                      []QRCodeDecodedEntry `json:"entries"`
	Cancelled                    bool                 `json:"cancelled"`
	SelectedEntryKey             string               `json:"selectedEntryKey,omitempty"`
	TriggeredFromLauncherVisible bool                 `json:"triggeredFromLauncherVisible,omitempty"`
	TriggeredFromLauncherFocused bool                 `json:"triggeredFromLauncherFocused,omitempty"`
}

type QRCodeDecodedEntry struct {
	Text         string  `json:"text"`
	Format       string  `json:"format"`
	Selectable   bool    `json:"selectable"`
	MarkerX      float64 `json:"markerX"`
	MarkerY      float64 `json:"markerY"`
	BoundsLeft   float64 `json:"boundsLeft"`
	BoundsTop    float64 `json:"boundsTop"`
	BoundsRight  float64 `json:"boundsRight"`
	BoundsBottom float64 `json:"boundsBottom"`
}

type qrCodeAnnotationStroke struct {
	id         int
	tool       string
	start      image.Point
	end        image.Point
	points     []image.Point
	color      color.RGBA
	lineWidth  int
	fontSize   int
	fontFamily string
	text       string
	textWidth  int
	number     int
}

type qrCodeCaptureAcceptedNotifier struct {
	callback func()
	once     sync.Once
}

type qrCodeNativeSelectionOptions struct {
	snapToControls        bool
	snapToWindows         bool
	confirmBeforeFinish   bool
	annotationToolbar     *screenshotSelectionToolbarWindowService
	annotationLineWidth   int
	cornerRadius          int
	cornerRadiusLabel     string
	captureCursor         bool
	fastGDIPreview        bool
	prewarmHDRCapture     bool
	freezeInitialFrame    bool
	waitForHDRPreview     bool
	cursorSnapshot        screenshotCursorSnapshot
	onAccepted            func()
	onLineWidthChanged    func(int)
	onCornerRadiusChanged func(int)
}

type qrCodeNativeSelectionResult struct {
	pngData           []uint8
	cancelled         bool
	needsFinalize     bool
	annotated         bool
	sourceProcessName string
	selection         image.Rectangle
	targetWindow      uintptr
	cursorSnapshot    screenshotCursorSnapshot
}

type qrCodeScreenSelectionSession struct {
	snapshot                    qrCodeScreenSnapshot
	originalBitmap              uintptr
	originalDC                  uintptr
	originalPrevious            uintptr
	shadedBitmap                uintptr
	shadedDC                    uintptr
	shadedPrevious              uintptr
	borderBrush                 uintptr
	cancelled                   bool
	resultPNG                   []uint8
	err                         error
	options                     qrCodeNativeSelectionOptions
	sourceProcessName           string
	selectedWindow              uintptr
	hitTestTransparent          bool
	rightButtonDownHandled      bool
	rightButtonSuppressContext  bool
	rightButtonCancelPending    bool
	dragging                    bool
	startPoint                  image.Point
	currentPoint                image.Point
	selection                   image.Rectangle
	selectionResizeHover        uint8
	selectionResizeActive       uint8
	selectionResizeOrigin       image.Rectangle
	selectionResizeRadiusOrigin int
	cornerRadius                int
	cornerRadiusHover           bool
	cornerRadiusDragging        bool
	cornerRadiusDragOrigin      int
	cornerRadiusMask            *image.Alpha
	cornerRadiusMaskRadius      int
	cornerRadiusInnerMask       *image.Alpha
	cornerRadiusInnerMaskRadius int
	controlSelection            image.Rectangle
	controlSelectionActive      bool
	hoverWindowRect             image.Rectangle
	hoverControlRect            image.Rectangle
	hoverControlWindow          uintptr
	hoverControlPoint           image.Point
	hoverControlAt              time.Time
	confirmBeforeFinish         bool
	editing                     bool
	annotationFinishing         bool
	annotationTool              string
	annotationColor             color.RGBA
	annotationColorPickerOpen   bool
	annotationLineWidth         int
	annotationFontSize          int
	annotationFontFamily        string
	annotationStrokes           []qrCodeAnnotationStroke
	annotationRedo              []qrCodeAnnotationStroke
	annotationContentTouched    bool
	annotationDraft             *qrCodeAnnotationStroke
	annotationTextInput         *qrCodeAnnotationStroke
	annotationText              strings.Builder
	annotationTextEditorPoint   image.Point
	annotationTextMoving        bool
	annotationTextSizing        bool
	annotationTextSelecting     bool
	annotationTextCaret         int
	annotationTextAnchor        int
	annotationTextMoveStart     image.Point
	annotationTextMoveSource    image.Point
	annotationTextAnchorStart   image.Point
	selectedStrokeID            int
	movingStrokeID              int
	moveStartPoint              image.Point
	moveSourceStroke            qrCodeAnnotationStroke
	annotationToolbar           *screenshotSelectionToolbarWindowService
	toolbarActions              chan screenshotSelectionToolbarAction
	toolbarTerminalActions      chan screenshotSelectionToolbarAction
	toolbarSessionID            int64
	nextStrokeID                int
	nextNumber                  int
	annotationEditing           atomic.Bool
	accepted                    *qrCodeCaptureAcceptedNotifier
	stateMutex                  sync.Mutex
	hwnd                        uintptr
	timeoutTriggered            bool
	upgradingPreview            bool
	pendingHDRPreview           *qrCodeScreenSnapshot
	finalizeRequest             *qrCodeNativeSelectionResult
}

type qrCodeScreenSnapshot struct {
	virtualBounds image.Rectangle
	original      *image.RGBA
	shaded        *image.RGBA
}
