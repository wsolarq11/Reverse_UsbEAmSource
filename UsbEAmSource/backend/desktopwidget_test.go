package main

import (
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestNewDesktopWidgetService(t *testing.T) {
	s := newDesktopWidgetService((*application.App)(nil), "C:/cfg.json")
	if s == nil {
		t.Fatal("factory must return non-nil")
	}
	// store 已装配（configPath 存入 store.path）。
	if s.store == nil || s.store.path != "C:/cfg.json" {
		t.Errorf("store=%+v want path=C:/cfg.json", s.store)
	}
	// weather 已装配。
	if s.weather == nil {
		t.Error("weather must be wired")
	}
	if s.weather.inflight == nil || s.weather.providerErrors == nil {
		t.Error("weather maps must be initialized")
	}
	// drafts map。
	if s.drafts == nil {
		t.Error("drafts map must be initialized")
	}
}
