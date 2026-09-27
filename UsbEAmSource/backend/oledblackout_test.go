package main

import (
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestNewOLEDBlackoutService(t *testing.T) {
	s := newOLEDBlackoutService((*application.App)(nil))
	if s == nil {
		t.Fatal("factory must return non-nil")
	}
	// 装配：5 个 map 已建（汇编 makemaap×5）。
	if s.overlayWindows == nil || s.visibleOverlays == nil ||
		s.hotkeyRegistered == nil || s.hotkeyErrors == nil || s.lastPressedKeys == nil {
		t.Error("all five oled maps must be initialized")
	}
	// 生命周期上下文已接线。
	if s.lifecycleContext == nil || s.lifecycleCancel == nil {
		t.Error("lifecycle context must be wired")
	}
	// hotkeyManager 已挂 [P] stub。
	if s.hotkeyManager == nil {
		t.Error("hotkeyManager must be set")
	}
}
