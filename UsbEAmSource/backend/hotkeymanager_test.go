package main

import "testing"

func TestNewLauncherGlobalHotkeyManager(t *testing.T) {
	m := newLauncherGlobalHotkeyManager()
	if m == nil {
		t.Fatal("factory must return non-nil")
	}
	// 接口可调（[P] 装配期实现）。
	if err := m.Update(launcherHotkeyBindings{}); err != nil {
		t.Errorf("Update err=%v", err)
	}
	if err := m.Close(); err != nil {
		t.Errorf("Close err=%v", err)
	}
	// 幂等二次 Close。
	if err := m.Close(); err != nil {
		t.Errorf("second Close err=%v", err)
	}
}
