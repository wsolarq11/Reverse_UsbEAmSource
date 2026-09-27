package main

import "testing"

func TestNewMouseGestureService(t *testing.T) {
	s := newMouseGestureService()
	if s == nil {
		t.Fatal("factory must return non-nil")
	}
	// 双 runtime chan 已建。
	if s.runtimeStop == nil || s.runtimeDone == nil {
		t.Error("runtime chans must be initialized")
	}
	// [S] 默认 GestureSettings 装配（start/stop 距离/超时）。
	if s.config.Settings.StartDistancePx != 150 || s.config.Settings.StartTimeoutMs != 300 ||
		s.config.Settings.StopTimeoutMs != 500 {
		t.Errorf("defaults got %+v", s.config.Settings)
	}
}
