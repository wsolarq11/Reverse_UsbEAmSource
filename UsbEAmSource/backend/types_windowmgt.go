// AUTO-RECONSTRUCTED TYPES — DOMAIN: windowmgt
// 研究用途
package main

import (
	"sync"
)

type WindowFullscreenSnapshot struct {
	HWND      uintptr              `json:"hwnd,omitempty"`
	ProcessID uint32               `json:"processId,omitempty"`
	Rect      WindowManagementRect `json:"rect,omitempty"`
	Style     uintptr              `json:"style,omitempty"`
	ExStyle   uintptr              `json:"exStyle,omitempty"`
	TopMost   bool                 `json:"topMost,omitempty"`
}

type WindowManagementInfo struct {
	HWND               uintptr              `json:"hwnd,omitempty"`
	WindowText         string               `json:"windowText,omitempty"`
	WindowClass        string               `json:"windowClass,omitempty"`
	WindowStyle        string               `json:"windowStyle,omitempty"`
	WindowExStyle      string               `json:"windowExStyle,omitempty"`
	WindowStyleValue   string               `json:"windowStyleValue,omitempty"`
	WindowExStyleValue string               `json:"windowExStyleValue,omitempty"`
	WindowID           uintptr              `json:"windowId,omitempty"`
	ParentHWND         uintptr              `json:"parentHwnd,omitempty"`
	ParentText         string               `json:"parentText,omitempty"`
	ParentClass        string               `json:"parentClass,omitempty"`
	ThreadID           uint32               `json:"threadId,omitempty"`
	ProcessID          uint32               `json:"processId,omitempty"`
	ProcessName        string               `json:"processName,omitempty"`
	ProcessPath        string               `json:"processPath,omitempty"`
	WindowRect         WindowManagementRect `json:"windowRect,omitempty"`
	ClientRect         WindowManagementRect `json:"clientRect,omitempty"`
}

type WindowManagementRect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type WindowManagementSize struct {
	Width      int  `json:"width,omitempty"`
	Height     int  `json:"height,omitempty"`
	ClientArea bool `json:"clientArea,omitempty"`
}

type WindowManagementState struct {
	Supported            bool                   `json:"supported"`
	Platform             string                 `json:"platform"`
	ModuleEnabled        bool                   `json:"moduleEnabled"`
	Config               WindowManagementConfig `json:"config"`
	Target               WindowManagementTarget `json:"target,omitempty"`
	TargetInfo           WindowManagementInfo   `json:"targetInfo,omitempty"`
	TargetValid          bool                   `json:"targetValid"`
	DisplayCount         int                    `json:"displayCount"`
	CursorWrapActive     bool                   `json:"cursorWrapActive"`
	TargetTopMost        bool                   `json:"targetTopMost"`
	TargetBorderless     bool                   `json:"targetBorderless"`
	TargetFullscreen     bool                   `json:"targetFullscreen"`
	TargetOpacityPercent int                    `json:"targetOpacityPercent,omitempty"`
	LastError            string                 `json:"lastError,omitempty"`
}

type WindowManagementTarget struct {
	HWND        uintptr `json:"hwnd,omitempty"`
	ProcessID   uint32  `json:"processId,omitempty"`
	ProcessName string  `json:"processName,omitempty"`
	Path        string  `json:"path,omitempty"`
	Title       string  `json:"title,omitempty"`
	DisplayName string  `json:"displayName,omitempty"`
	IconData    string  `json:"iconData,omitempty"`
	IconRef     string  `json:"iconRef,omitempty"`
	IconURL     string  `json:"iconUrl,omitempty"`
}

type windowManagementMonitorInfo struct {
	Size    uint32
	Monitor windowManagementRECT
	Work    windowManagementRECT
	Flags   uint32
}

type windowManagementOpacitySnapshot struct {
	HWND       uintptr
	ProcessID  uint32
	ExStyle    uintptr
	WasLayered bool
	ColorKey   uint32
	Alpha      uint8
	Flags      uint32
}

type windowManagementPoint struct {
	X int32
	Y int32
}

type windowManagementRECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type WindowManagementConfig struct {
	CursorWrapHorizontalEnabled bool                       `json:"cursorWrapHorizontalEnabled,omitempty"`
	CursorWrapVerticalEnabled   bool                       `json:"cursorWrapVerticalEnabled,omitempty"`
	Target                      WindowManagementTarget     `json:"target,omitempty"`
	OpacityPercent              int                        `json:"opacityPercent"`
	Resolution                  WindowManagementSize       `json:"resolution,omitempty"`
	BorderlessSnapshots         []WindowFullscreenSnapshot `json:"borderlessSnapshots,omitempty"`
	FullscreenSnapshots         []WindowFullscreenSnapshot `json:"fullscreenSnapshots,omitempty"`
	opacityPercentSet           bool
}

type windowManagementConfigAlias struct {
	CursorWrapHorizontalEnabled bool                       `json:"cursorWrapHorizontalEnabled,omitempty"`
	CursorWrapVerticalEnabled   bool                       `json:"cursorWrapVerticalEnabled,omitempty"`
	Target                      WindowManagementTarget     `json:"target,omitempty"`
	OpacityPercent              int                        `json:"opacityPercent"`
	Resolution                  WindowManagementSize       `json:"resolution,omitempty"`
	BorderlessSnapshots         []WindowFullscreenSnapshot `json:"borderlessSnapshots,omitempty"`
	FullscreenSnapshots         []WindowFullscreenSnapshot `json:"fullscreenSnapshots,omitempty"`
	opacityPercentSet           bool
}

type windowManagementService struct {
	operationLock       sync.Mutex
	lock                sync.Mutex
	config              WindowManagementConfig
	moduleEnabled       bool
	shuttingDown        bool
	persistConfig       func(WindowManagementConfig)
	cursorStop          chan struct{}
	cursorDone          chan struct{}
	cursorActive        bool
	cursorGuardPx       int
	opacitySnapshots    map[uintptr]windowManagementOpacitySnapshot
	managedSnapshots    map[uintptr]WindowFullscreenSnapshot
	borderlessSnapshots []WindowFullscreenSnapshot
	fullscreenSnapshots []WindowFullscreenSnapshot
	lastError           string
}
