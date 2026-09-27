package main

import "testing"

func TestVisitLauncherConfigIconData(t *testing.T) {
	called := 0
	visit := func(string) { called++ }

	// nil cfg 安全。
	visitLauncherConfigIconData(nil, visit)
	// 非 nil cfg → visitLauncherConfigIconSlots（[P] 骨架 no-op，不 panic）。
	cfg := &LauncherConfig{Revision: "x"}
	visitLauncherConfigIconData(cfg, visit)
	visitLauncherConfigIconData(cfg, nil)
	_ = called
}

func TestFingerprintDistinguishesConfigIcons(t *testing.T) {
	empty := LauncherConfig{}
	withIcon := LauncherConfig{Apps: []AppEntry{{Name: "a", IconData: "data:image/png;base64,XXX"}}}
	fpEmpty := launcherConfigIconDataFingerprint(empty)
	fpIcon := launcherConfigIconDataFingerprint(withIcon)
	// [S] visit 遍历 Apps 图标段 → 带图标配置指纹应不同（行为真实化）。
	if fpEmpty == fpIcon {
		t.Error("config with App icon data must yield different fingerprint")
	}
	// 确定性保持。
	if fpEmpty != launcherConfigIconDataFingerprint(empty) {
		t.Error("empty fingerprint must be deterministic")
	}
}

func TestFingerprintDistinguishesSpeedDialIcon(t *testing.T) {
	empty := LauncherConfig{}
	withSD := LauncherConfig{SpeedDial: []LinkEntry{{IconData: "data:image/png;base64,SD"}}}
	if launcherConfigIconDataFingerprint(empty) == launcherConfigIconDataFingerprint(withSD) {
		t.Error("SpeedDial IconData must differ fingerprint")
	}
}

func TestFingerprintDistinguishesBookmarkIcon(t *testing.T) {
	empty := LauncherConfig{}
	// 闭包实证(fingerprint.func1) 只写 Path+Data 到 hash；Ref/URL 不入哈希。
	// 故用 IconData 而非 IconRef 来测差异。
	withBM := LauncherConfig{Bookmarks: BookmarkSection{Custom: []LinkEntry{{IconData: "bm:data"}}}}
	if launcherConfigIconDataFingerprint(empty) == launcherConfigIconDataFingerprint(withBM) {
		t.Error("Bookmarks IconData must differ fingerprint")
	}
}

func TestFingerprintDistinguishesTagCatalogIcon(t *testing.T) {
	empty := LauncherConfig{}
	withTC := LauncherConfig{Preferences: Preferences{TagCatalog: []TagCatalogItem{{IconData: "tc:data"}}}}
	if launcherConfigIconDataFingerprint(empty) == launcherConfigIconDataFingerprint(withTC) {
		t.Error("TagCatalog IconData must differ fingerprint")
	}
}

func TestFingerprintDistinguishesConsoleItemIcon(t *testing.T) {
	empty := LauncherConfig{}
	withCI := LauncherConfig{Preferences: Preferences{ConsoleItems: []ConsoleItem{{IconData: "ci:data"}}}}
	if launcherConfigIconDataFingerprint(empty) == launcherConfigIconDataFingerprint(withCI) {
		t.Error("ConsoleItem IconData must differ fingerprint")
	}
}

func TestFingerprintDistinguishesTwoFactorEntryIcon(t *testing.T) {
	empty := LauncherConfig{}
	withTF := LauncherConfig{TwoFactor: TwoFactorConfig{Entries: []TwoFactorEntryConfig{{IconData: "tf:data"}}}}
	if launcherConfigIconDataFingerprint(empty) == launcherConfigIconDataFingerprint(withTF) {
		t.Error("TwoFactorEntry IconData must differ fingerprint")
	}
}

func TestFingerprintDistinguishesOLEDMediaPauseIcon(t *testing.T) {
	empty := LauncherConfig{}
	withOM := LauncherConfig{OLEDBlackout: OLEDBlackoutConfig{MediaPauseExclusions: []OLEDBlackoutMediaPauseExclusion{{IconData: "om:data"}}}}
	if launcherConfigIconDataFingerprint(empty) == launcherConfigIconDataFingerprint(withOM) {
		t.Error("OLED MediaPauseExclusion IconData must differ fingerprint")
	}
}

func TestFingerprintDistinguishesWindowTargetIcon(t *testing.T) {
	empty := LauncherConfig{}
	withWT := LauncherConfig{WindowManagement: WindowManagementConfig{Target: WindowManagementTarget{IconData: "wt:data"}}}
	if launcherConfigIconDataFingerprint(empty) == launcherConfigIconDataFingerprint(withWT) {
		t.Error("WindowManagement Target IconData must differ fingerprint")
	}
}
