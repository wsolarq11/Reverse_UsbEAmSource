// AUTO-RECONSTRUCTED TYPES — DOMAIN: inputmonitor
// 研究用途
package main

import (
	"sync"
	"sync/atomic"
	"time"
)

type InputMonitorEvent struct {
	ID         uint64 `json:"id"`
	Timestamp  int64  `json:"timestamp"`
	Device     string `json:"device"`
	Type       string `json:"type"`
	Action     string `json:"action"`
	Key        string `json:"key,omitempty"`
	Code       string `json:"code,omitempty"`
	VKCode     uint32 `json:"vkCode,omitempty"`
	ScanCode   uint32 `json:"scanCode,omitempty"`
	Button     string `json:"button,omitempty"`
	WheelDelta int    `json:"wheelDelta,omitempty"`
	WheelAxis  string `json:"wheelAxis,omitempty"`
	X          int    `json:"x,omitempty"`
	Y          int    `json:"y,omitempty"`
	Injected   bool   `json:"injected,omitempty"`
	Repeat     bool   `json:"repeat,omitempty"`
	AltKey     bool   `json:"altKey,omitempty"`
	CtrlKey    bool   `json:"ctrlKey,omitempty"`
	ShiftKey   bool   `json:"shiftKey,omitempty"`
	MetaKey    bool   `json:"metaKey,omitempty"`
}

type InputMonitorSnapshot struct {
	Supported       bool                `json:"supported"`
	Platform        string              `json:"platform"`
	Running         bool                `json:"running"`
	Keyboard        bool                `json:"keyboard"`
	Mouse           bool                `json:"mouse"`
	ThreadHealthy   bool                `json:"threadHealthy"`
	HookHealthy     bool                `json:"hookHealthy"`
	StartedAt       string              `json:"startedAt,omitempty"`
	LastError       string              `json:"lastError,omitempty"`
	LastEventID     uint64              `json:"lastEventId"`
	EarliestEventID uint64              `json:"earliestEventId"`
	DroppedEvents   uint64              `json:"droppedEvents"`
	Gap             bool                `json:"gap"`
	BufferSize      int                 `json:"bufferSize"`
	BufferLimit     int                 `json:"bufferLimit"`
	Events          []InputMonitorEvent `json:"events"`
	MonitorStrategy string              `json:"monitorStrategy,omitempty"`
}

type InputMonitorStartRequest struct {
	Keyboard bool `json:"keyboard,omitempty"`
	Mouse    bool `json:"mouse,omitempty"`
}

type inputMonitorKeyboardHookEvent struct {
	vkCode      uint32
	scanCode    uint32
	flags       uint32
	time        uint32
	dwExtraInfo uintptr
}

type inputMonitorMouseHookEvent struct {
	point       inputMonitorPoint
	mouseData   uint32
	flags       uint32
	time        uint32
	dwExtraInfo uintptr
}

type inputMonitorPlatformCommand struct {
	close  bool
	result chan error
}

type inputMonitorPoint struct {
	x int32
	y int32
}

type inputMonitorRawEvent struct {
	generation uint64
	message    uint32
	keyboard   bool
	keyEvent   inputMonitorKeyboardHookEvent
	mouseEvent inputMonitorMouseHookEvent
}

type inputMonitorService struct {
	lifecycleMu                sync.Mutex
	lock                       sync.Mutex
	supported                  bool
	running                    bool
	keyboard                   bool
	mouse                      bool
	startedAt                  time.Time
	lastError                  string
	nextID                     uint64
	events                     []InputMonitorEvent
	eventHead                  int
	eventSize                  int
	droppedEvents              uint64
	rawDropped                 atomic.Uint64
	pressedKeys                map[uint32]bool
	maxEvents                  int
	commandChan                chan inputMonitorPlatformCommand
	ready                      chan error
	done                       chan struct{}
	threadID                   uint32
	keyboardHook               uintptr
	mouseHook                  uintptr
	keyboardHookRef            uintptr
	mouseHookRef               uintptr
	platformGeneration         atomic.Uint64
	platformHealthManaged      bool
	platformThreadAlive        bool
	platformMessageLoopHealthy bool
	platformKeyboardInstalled  bool
	platformMouseInstalled     bool
	closeOnce                  sync.Once
	owners                     map[string]InputMonitorStartRequest
	closed                     bool
	startOverride              func(InputMonitorStartRequest) error
	stopOverride               func() error
	closeOverride              func() error
}
