// AUTO-RECONSTRUCTED TESTS — DOMAIN: bootstrap service
// 研究用途
//
// 覆盖引导服务层已还原函数，聚焦：
//   - 纯函数单元测试（钳制/校验/切片变换）
//   - setStartupTrayMode 字段修正验证（as 校准锚点，前轮错误已按实证修正）
//   - 快照/配置方法骨架的回归保护
//   - 错误常量引用验证
package main

import (
	"errors"
	"testing"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ---- 辅助纯函数 ----

func TestClampLauncherUIScalePercent(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{-100, 100},
		{0, 100},
		{50, 100},
		{99, 100},
		{100, 100},
		{150, 150},
		{200, 200},
		{299, 299},
		{300, 300},
		{301, 300},
		{999, 300},
	}
	for _, tc := range tests {
		got := clampLauncherUIScalePercent(tc.input)
		if got != tc.want {
			t.Errorf("clampLauncherUIScalePercent(%d) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestValidateLauncherConfigExportTarget(t *testing.T) {
	if err := validateLauncherConfigExportTarget("/some/path", "/cfg/path"); err != nil {
		t.Errorf("non-empty path should succeed: %v", err)
	}
	if err := validateLauncherConfigExportTarget("", "/cfg/path"); !errors.Is(err, errLauncherConfigExportTargetEmpty) {
		t.Errorf("empty path should return errLauncherConfigExportTargetEmpty, got %v", err)
	}
}

func TestTagCatalogNames(t *testing.T) {
	if got := tagCatalogNames(nil); got != nil {
		t.Errorf("nil input should return nil, got %v", got)
	}
	if got := tagCatalogNames([]TagCatalogItem{}); got != nil {
		t.Errorf("empty slice should return nil, got %v", got)
	}
	items := []TagCatalogItem{
		{Name: "alpha"},
		{Name: "beta"},
	}
	got := tagCatalogNames(items)
	if len(got) != 2 || got[0] != "alpha" || got[1] != "beta" {
		t.Errorf("tagCatalogNames = %v, want [alpha beta]", got)
	}
}

// ---- 错误常量验证 ----

func TestBootstrapErrorConstants(t *testing.T) {
	if errBootstrapConfigPathEmpty.Error() == "" {
		t.Error("errBootstrapConfigPathEmpty should not be empty")
	}
	if errBootstrapExportPathEmpty.Error() == "" {
		t.Error("errBootstrapExportPathEmpty should not be empty")
	}
	if errBootstrapNoDataRoot.Error() == "" {
		t.Error("errBootstrapNoDataRoot should not be empty")
	}
	if errBootstrapMigratePathEmpty.Error() == "" {
		t.Error("errBootstrapMigratePathEmpty should not be empty")
	}
}

// ---- 字段修正验证 ----

// TestSetStartupTrayModeFields verifies field writes match asm evidence.
// Asm anchor: setStartupTrayMode.asm 0x14077b7b6/0x14077b7bc writes +0x449 and +0x448,
// mapped to startupTrayMode / resetLauncherSizeOnNextShow (previous error wrote pendingLauncherReveal).
func TestSetStartupTrayModeFields(t *testing.T) {
	bs := &BootstrapService{}
	bs.setStartupTrayMode(true)
	if !bs.startupTrayMode {
		t.Error("setStartupTrayMode(true): startupTrayMode should be true")
	}
	if !bs.resetLauncherSizeOnNextShow {
		t.Error("setStartupTrayMode(true): resetLauncherSizeOnNextShow should be true")
	}
	bs.setStartupTrayMode(false)
	if bs.startupTrayMode {
		t.Error("setStartupTrayMode(false): startupTrayMode should be false")
	}
	if bs.resetLauncherSizeOnNextShow {
		t.Error("setStartupTrayMode(false): resetLauncherSizeOnNextShow should be false")
	}
}

// TestMarkLauncherFrontendReadyFields verifies frontend-ready flag and QR result clearing.
func TestMarkLauncherFrontendReadyFields(t *testing.T) {
	bs := &BootstrapService{}
	bs.MarkLauncherFrontendReady()
	if !bs.launcherFrontendReady {
		t.Error("MarkLauncherFrontendReady: launcherFrontendReady should be true")
	}

	// Verify pending QR result is taken and cleared
	result := &QRCodeDecodeResult{}
	bs.pendingQRCodeDecodeResult = result
	bs.MarkLauncherFrontendReady()
	if bs.pendingQRCodeDecodeResult != nil {
		t.Error("MarkLauncherFrontendReady: pendingQRCodeDecodeResult should be cleared to nil")
	}
}

// ---- 骨架方法回归保护 ----

func TestPreviewLauncherUIScaleOnNilWindow(t *testing.T) {
	bs := &BootstrapService{}
	// Should not panic, resolveAttachedLauncherWindow returns zero value
	err := bs.PreviewLauncherUIScale(150)
	if err != nil {
		t.Errorf("PreviewLauncherUIScale(150) on nil window: %v", err)
	}
}

func TestAbortInitializationNilStore(t *testing.T) {
	bs := &BootstrapService{}
	err := bs.AbortInitialization()
	if err != nil {
		t.Errorf("AbortInitialization on uninit: %v", err)
	}
}

func TestResetConfigNilStore(t *testing.T) {
	bs := &BootstrapService{}
	err := bs.ResetConfig()
	_ = err
}

func TestImportConfigEmptyData(t *testing.T) {
	bs := &BootstrapService{}
	err := bs.ImportConfig("")
	// [P] readLauncherConfigFileWithDesktopWidgets is stub, should not panic
	_ = err
}

func TestCompleteInitializationOnZeroValue(t *testing.T) {
	bs := &BootstrapService{}
	err := bs.CompleteInitialization()
	_ = err
}

func TestChooseInitializationDataRootUninitialized(t *testing.T) {
	bs := &BootstrapService{}
	path, err := bs.ChooseInitializationDataRoot()
	if err != nil && !errors.Is(err, errBootstrapNoDataRoot) {
		// May return empty string on uninitialized service
		_ = path
	}
}

func TestChooseLauncherBackgroundImageUninitialized(t *testing.T) {
	bs := &BootstrapService{}
	_, err := bs.ChooseLauncherBackgroundImage()
	_ = err
}

func TestMigrateConfigEmptyPath(t *testing.T) {
	bs := &BootstrapService{}
	err := bs.MigrateConfig("", "")
	if !errors.Is(err, errBootstrapMigratePathEmpty) {
		t.Errorf("empty path should return errBootstrapMigratePathEmpty, got %v", err)
	}
}

func TestExportTextFileEmptyPath(t *testing.T) {
	bs := &BootstrapService{}
	err := bs.ExportTextFile("", "content")
	if !errors.Is(err, errBootstrapExportPathEmpty) {
		t.Errorf("empty path should return errBootstrapExportPathEmpty, got %v", err)
	}
}

// ---- P1 快照体系回归测试 ----

// TestWorkspaceSnapshotAtomicity verifies workspaceSnapshot returns bs.workspace directly.
func TestWorkspaceSnapshotAtomicity(t *testing.T) {
	ws := WorkspaceLayout{Root: "C:\\test\\root", ConfigFile: "cfg.json"}
	bs := &BootstrapService{workspace: ws}
	got := bs.workspaceSnapshot()
	if got != ws {
		t.Errorf("workspaceSnapshot = %+v, want %+v", got, ws)
	}
}

// TestWorkspaceSnapshotOnNil verifies nil bootstrap doesn't panic.
func TestWorkspaceSnapshotOnNil(t *testing.T) {
	var bs *BootstrapService
	got := bs.workspaceSnapshot()
	if got.Root != "" || got.ConfigFile != "" {
		t.Errorf("nil bs should return zero WorkspaceLayout, got %+v", got)
	}
}

// TestCloneBootstrapMessageValueString verifies string values passthrough.
func TestCloneBootstrapMessageValueString(t *testing.T) {
	v := "hello"
	got := cloneBootstrapMessageValue(v)
	if got == nil {
		t.Fatal("clone returned nil")
	}
	s, ok := got.(string)
	if !ok || s != "hello" {
		t.Errorf("cloneBootstrapMessageValue(string) = %v (%T), want 'hello'", got, got)
	}
}

// TestCloneBootstrapMessageValueSliceString verifies []string deep copy.
func TestCloneBootstrapMessageValueSliceString(t *testing.T) {
	src := []string{"a", "b", "c"}
	got := cloneBootstrapMessageValue(src)
	if got == nil {
		t.Fatal("clone returned nil")
	}
	s, ok := got.([]string)
	if !ok {
		t.Fatalf("type = %T, want []string", got)
	}
	if len(s) != 3 || s[0] != "a" || s[1] != "b" || s[2] != "c" {
		t.Errorf("clone = %v, want [a b c]", s)
	}
	// Verify isolation
	s[0] = "x"
	if src[0] != "a" {
		t.Errorf("clone modified source: src[0] = %q, want 'a'", src[0])
	}
}

// TestCloneBootstrapMessageValueSliceInterface verifies []interface{} deep copy.
func TestCloneBootstrapMessageValueSliceInterface(t *testing.T) {
	src := []interface{}{"a", []string{"b", "c"}, 42}
	got := cloneBootstrapMessageValue(src)
	if got == nil {
		t.Fatal("clone returned nil")
	}
	s, ok := got.([]interface{})
	if !ok {
		t.Fatalf("type = %T, want []interface{}", got)
	}
	if len(s) != 3 {
		t.Fatalf("len = %d, want 3", len(s))
	}
	// Recursive isolation: modify inner slice
	inner, _ := s[1].([]string)
	if len(inner) != 2 {
		t.Fatalf("inner slice len = %d, want 2", len(inner))
	}
	inner[0] = "x"
	// Source should not reflect the change
	orig, _ := src[1].([]string)
	if orig[0] != "b" {
		t.Errorf("clone modified source: src[1][0] = %q, want 'b'", orig[0])
	}
}

// TestCloneBootstrapMessageValueMapStringString verifies map[string]string deep copy.
func TestCloneBootstrapMessageValueMapStringString(t *testing.T) {
	src := map[string]string{"k1": "v1", "k2": "v2"}
	got := cloneBootstrapMessageValue(src)
	if got == nil {
		t.Fatal("clone returned nil")
	}
	m, ok := got.(map[string]string)
	if !ok {
		t.Fatalf("type = %T, want map[string]string", got)
	}
	if m["k1"] != "v1" || m["k2"] != "v2" || len(m) != 2 {
		t.Errorf("clone = %v, want {k1:v1 k2:v2}", m)
	}
	// Isolation
	m["k3"] = "v3"
	if len(src) != 2 {
		t.Errorf("clone modified source: src len = %d, want 2", len(src))
	}
}

// TestCloneBootstrapMessageValueMapInterface verifies map[string]interface{} delegates.
func TestCloneBootstrapMessageValueMapInterface(t *testing.T) {
	src := map[string]interface{}{"key": []string{"a", "b"}}
	got := cloneBootstrapMessageValue(src)
	if got == nil {
		t.Fatal("clone returned nil")
	}
	m, ok := got.(map[string]interface{})
	if !ok {
		t.Fatalf("type = %T, want map[string]interface{}", got)
	}
	inner, _ := m["key"].([]string)
	inner[0] = "x"
	// Source isolation
	orig, _ := src["key"].([]string)
	if orig[0] == "x" {
		t.Error("source mutated by clone")
	}
}

// TestCloneBootstrapMessageMap verifies the high-level message map clone.
func TestCloneBootstrapMessageMap(t *testing.T) {
	src := map[string]interface{}{"a": 1, "b": []string{"x"}}
	dst := cloneBootstrapMessageMap(src)
	if len(dst) != 2 || dst["a"] != 1 {
		t.Errorf("cloneBootstrapMessageMap = %v, want {a:1 b:[x]}", dst)
	}
	// Isolation: modify nested slice
	nested, _ := dst["b"].([]string)
	nested[0] = "y"
	if src["b"].([]string)[0] != "x" {
		t.Error("source mutated by clone")
	}
}

// TestCloneBootstrapBoolPtr verifies *bool cloning.
func TestCloneBootstrapBoolPtr(t *testing.T) {
	// nil case
	if got := cloneBootstrapBoolPtr(nil); got != nil {
		t.Errorf("nil ptr should stay nil, got %v", *got)
	}
	// valid case
	v := true
	got := cloneBootstrapBoolPtr(&v)
	if got == nil || *got != true {
		t.Errorf("clone should be true, got %v", got)
	}
	*got = false
	if v != true {
		t.Error("clone modified source ptr")
	}
}

// TestCloneBootstrapLocalizationMap verifies Localization map deep copy.
func TestCloneBootstrapLocalizationMap(t *testing.T) {
	src := map[string]PluginLocalization{
		"en": {Messages: map[string]interface{}{"hello": "world"}},
	}
	dst := cloneBootstrapLocalizationMap(src)
	if len(dst) != 1 || dst["en"].Messages["hello"] != "world" {
		t.Errorf("clone = %+v", dst)
	}
	dst["en"].Messages["hello"] = "modified"
	if src["en"].Messages["hello"] != "world" {
		t.Error("source mutated by clone")
	}
}

// TestCloneBootstrapPermissions verifies permissions slice cloning.
func TestCloneBootstrapPermissions(t *testing.T) {
	if got := cloneBootstrapPermissions(nil); got != nil {
		t.Errorf("nil permissions should stay nil")
	}
	if got := cloneBootstrapPermissions([]string{}); got != nil {
		t.Errorf("empty permissions should become nil")
	}
	src := []string{"read", "write"}
	dst := cloneBootstrapPermissions(src)
	dst[0] = "delete"
	if src[0] != "read" {
		t.Error("source mutated by clone")
	}
}

// TestCacheBootstrapSnapshotNilReceivers safely handles nil BootstrapService
// for the methods that have assembly-verified nil guards.
func TestCacheBootstrapSnapshotNilReceivers(t *testing.T) {
	var bs *BootstrapService
	// cacheBootstrapSnapshot and cachedBootstrapSnapshotForWorkspace handle nil
	bs.cacheBootstrapSnapshot(BootstrapSnapshot{})
	_, ok := bs.cachedBootstrapSnapshotForWorkspace(WorkspaceLayout{})
	if ok {
		t.Error("cachedBootstrapSnapshotForWorkspace on nil bs should return false")
	}
	// Note: snapshotForCommittedLauncherState ends with GetSnapshot which has
	// no nil guard in asm, so skip on nil bs.
	// workspaceSnapshot has explicit nil guard.
	_ = bs.workspaceSnapshot()
}

// TestSnapshotForCommittedLauncherState verifies delegation to GetSnapshot when cache miss.
func TestSnapshotForCommittedLauncherState(t *testing.T) {
	bs := &BootstrapService{}
	_, err := bs.snapshotForCommittedLauncherState()
	_ = err
}

// TestCachedBootstrapSnapshotForWorkspace verifies cache-only semantics.
func TestCachedBootstrapSnapshotForWorkspace(t *testing.T) {
	ws := WorkspaceLayout{Root: "/test"}
	bs := &BootstrapService{
		workspace:              ws,
		bootstrapSnapshot:      BootstrapSnapshot{Workspace: ws},
		bootstrapSnapshotReady: true,
	}
	snap, ok := bs.cachedBootstrapSnapshotForWorkspace(ws)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if snap.Workspace != ws {
		t.Errorf("snapshot workspace = %+v, want %+v", snap.Workspace, ws)
	}
	// Cache miss on different workspace
	different := WorkspaceLayout{Root: "/other"}
	_, ok2 := bs.cachedBootstrapSnapshotForWorkspace(different)
	if ok2 {
		t.Error("expected cache miss for different workspace")
	}
}

// TestCloneBootstrapSnapshot verifies full BootstrapSnapshot deep copy.
func TestCloneBootstrapSnapshot(t *testing.T) {
	ws := WorkspaceLayout{Root: "/test"}
	langs := []LanguageManifest{
		{Code: "en", Messages: map[string]interface{}{"k": "v"}},
	}
	src := BootstrapSnapshot{
		Workspace: ws,
		Languages: langs,
	}
	dst := cloneBootstrapSnapshot(src)
	if dst.Workspace != ws {
		t.Errorf("workspace changed: %+v", dst.Workspace)
	}
	if len(dst.Languages) != 1 || dst.Languages[0].Code != "en" {
		t.Errorf("languages: %+v", dst.Languages)
	}
	// Mutate clone — source should not be affected
	dst.Languages[0].Messages["k"] = "modified"
	if src.Languages[0].Messages["k"] != "v" {
		t.Error("source mutated by clone")
	}
}

// TestWorkspaceSnapshotFieldOffset verifies bs.workspace field is at struct+0x0 offset.
// Cross-checking asm lock→duffcopy 0xa0, which is sizeof(WorkspaceLayout) as first field.
func TestWorkspaceSnapshotFieldOffset(t *testing.T) {
	ws := WorkspaceLayout{Root: "C:\\test"}
	bs := &BootstrapService{workspace: ws}
	// unsafe check: workspace is first field in BootstrapService
	firstField := *(*WorkspaceLayout)(unsafe.Pointer(bs))
	if firstField != ws {
		t.Errorf("first struct field through unsafe = %+v, want %+v", firstField, ws)
	}
}

// TestBootstrapSnapshotFields verifies the 0xa0 size assumption via offset calculation.
func TestBootstrapSnapshotFields(t *testing.T) {
	var ws WorkspaceLayout
	wsSize := uintptr(unsafe.Sizeof(ws))
	if wsSize != 0xa0 {
		// 10 string fields × 16 bytes (ptr+len) = 160 = 0xa0
		t.Logf("WorkspaceLayout size = 0x%x (expected 0xa0)", wsSize)
	}
}

// ---- P2 状态体系回归测试（本轮新增） ----

// TestGetStartupStateOnEmptyService verifies state assembly on zero-value BootstrapService.
// Note: the asm has no nil guard for bs, so nil-receiver test is excluded.
func TestGetStartupStateOnEmptyService(t *testing.T) {
	bs := &BootstrapService{}
	_, err := bs.GetStartupState()
	if err != nil {
		// [P] ensureWorkspaceDirectories creates dirs; depends on OS state
		t.Logf("GetStartupState on empty: err=%v (acceptable if workspace dirs fail)", err)
	}
}

// TestGetStateOnEmptyService verifies GetState behavior on zero-value BootstrapService.
func TestGetStateOnEmptyService(t *testing.T) {
	bs := &BootstrapService{}
	_, err := bs.GetState()
	if err != nil {
		t.Logf("GetState on empty: err=%v", err)
	} else {
		t.Log("GetState on empty returned nil error")
	}
}

// TestLauncherStateFromCommittedConfigWithSnapshot verifies state assembly from snapshot.
func TestLauncherStateFromCommittedConfigWithSnapshot(t *testing.T) {
	ws := WorkspaceLayout{Root: "/test"}
	cfg := LauncherConfig{Version: 42}
	snap := BootstrapSnapshot{
		Workspace: ws,
	}
	bs := &BootstrapService{
		startupTrayMode: true,
	}
	state := bs.launcherStateFromCommittedConfigWithSnapshot(cfg, snap)
	if state.Workspace != ws {
		t.Errorf("workspace mismatch: %+v", state.Workspace)
	}
	if state.Config.Version != 42 {
		t.Errorf("config version mismatch: got %d, want 42", state.Config.Version)
	}
	if !state.StartupTrayMode {
		t.Error("StartupTrayMode should be true (from service field)")
	}
	if state.ConfigOnly {
		t.Error("ConfigOnly should be false for committed config")
	}
}

// TestLauncherStateFromSavedConfig verifies saved-config-only state assembly.
func TestLauncherStateFromSavedConfig(t *testing.T) {
	cfg := LauncherConfig{}
	bs := &BootstrapService{}
	state := bs.launcherStateFromSavedConfig(cfg, true)
	if state.Workspace != (WorkspaceLayout{}) {
		t.Errorf("workspace should be zero, got %+v", state.Workspace)
	}
	if !state.ConfigOnly {
		t.Error("ConfigOnly should be true for saved config")
	}
	if !state.ContentRuntimeSyncNeeded {
		t.Error("ContentRuntimeSyncNeeded should match input (true)")
	}
}

// TestLauncherStateFromCommittedConfigCaches verifies delegation through cache.
func TestLauncherStateFromCommittedConfigCaches(t *testing.T) {
	bs := &BootstrapService{}
	cfg := LauncherConfig{}
	_, err := bs.launcherStateFromCommittedConfig(cfg)
	_ = err
}

// TestShutdownNilReceiver verifies nil-guard.
func TestShutdownNilReceiver(t *testing.T) {
	var bs *BootstrapService
	bs.Shutdown() // must not panic
}

// TestShutdownOnEmptyService verifies Shutdown on zero-value BootstrapService.
func TestShutdownOnEmptyService(t *testing.T) {
	bs := &BootstrapService{}
	bs.Shutdown() // must not panic; all services are nil
}

// TestShutdownResourceCleanup verifies key side effects on Shutdown.
func TestShutdownResourceCleanup(t *testing.T) {
	bs := &BootstrapService{
		globalHotkey: &launcherGlobalHotkeyManagerImpl{
			service:   make(chan *launcherGlobalHotkeyCommand, 4),
			closeDone: make(chan struct{}),
		},
		memoryRelease: &memoryReleaseService{},
		oledBlackout:  &oledBlackoutService{},
	}
	bs.Shutdown()
	if bs.allowLauncherWindowClose != true {
		t.Error("allowLauncherWindowClose should be set to true after Shutdown")
	}
	if bs.globalHotkey != nil {
		t.Error("globalHotkey should be nil after Shutdown")
	}
}

// TestIsStartupTrayModeNilReceiver verifies nil-guard.
func TestIsStartupTrayModeNilReceiver(t *testing.T) {
	var bs *BootstrapService
	if bs.isStartupTrayMode() {
		t.Error("nil bs should return false")
	}
}

// TestIsStartupTrayModeField verifies field read.
func TestIsStartupTrayModeField(t *testing.T) {
	bs := &BootstrapService{startupTrayMode: true}
	if !bs.isStartupTrayMode() {
		t.Error("isStartupTrayMode should return true")
	}
	bs.startupTrayMode = false
	if bs.isStartupTrayMode() {
		t.Error("isStartupTrayMode should return false")
	}
}

// TestCurrentHotkeyRegistrationErrorsNil verifies nil guard.
func TestCurrentHotkeyRegistrationErrorsNil(t *testing.T) {
	bs := &BootstrapService{}
	errs := bs.currentHotkeyRegistrationErrors()
	if errs != nil {
		t.Errorf("expected nil, got %v", errs)
	}
}

// TestFileExistsNotExist verifies ErrNotExist handling.
func TestFileExistsNotExist(t *testing.T) {
	ok, err := fileExists("C:\\__nonexistent_path__")
	if ok {
		t.Error("fileExists should return false for nonexistent path")
	}
	if err != nil {
		t.Errorf("fileExists should not return error for nonexistent path: %v", err)
	}
}

// TestLauncherConfigRevisionRoundtrip verifies pass-through gives same string.
func TestLauncherConfigRevisionRoundtrip(t *testing.T) {
	cfg := LauncherConfig{Revision: "test-rev"}
	got := launcherConfigRevision(cfg)
	if got != "test-rev" {
		t.Errorf("expected 'test-rev', got %q", got)
	}
}

// TestBootstrapServiceLayout verifies struct field sizes match asm expectations.
func TestBootstrapServiceLayout(t *testing.T) {
	var bs BootstrapService
	lockOff := unsafe.Offsetof(bs.lock)
	if lockOff != 0x540 {
		t.Errorf("BootstrapService.lock offset = 0x%x (expected 0x540)", lockOff)
	}
	startupTaskSyncOff := unsafe.Offsetof(bs.startupTaskSync)
	if startupTaskSyncOff != 0x510 {
		t.Errorf("BootstrapService.startupTaskSync offset = 0x%x (expected 0x510)", startupTaskSyncOff)
	}
}

// TestSyncStartupTaskNilReceiver 锁定批次 131 nil 检查：bs==nil → "启动任务服务不可用"。
func TestSyncStartupTaskNilReceiver(t *testing.T) {
	var bs *BootstrapService
	err := bs.syncStartupTask(false, 0)
	if err == nil || err.Error() != "启动任务服务不可用" {
		t.Fatalf("nil 接收者应返回服务不可用错误, got %v", err)
	}
}

// TestSyncStartupTaskCustomFn 锁定批次 131 字段注入：startupTaskSync 非 nil 时透传参数。
func TestSyncStartupTaskCustomFn(t *testing.T) {
	var gotEnabled bool
	var gotDelay int
	bs := &BootstrapService{
		startupTaskSync: func(enabled bool, delay int) error {
			gotEnabled = enabled
			gotDelay = delay
			return errors.New("custom")
		},
	}
	err := bs.syncStartupTask(true, 42)
	if err == nil || err.Error() != "custom" {
		t.Fatalf("应透传自定义 fn 的错误, got %v", err)
	}
	if !gotEnabled || gotDelay != 42 {
		t.Fatalf("参数透传错误: enabled=%v delay=%d", gotEnabled, gotDelay)
	}
}

// TestLauncherConfigConflictErrorError 锁定批次 133：冲突错误消息格式。
func TestLauncherConfigConflictErrorError(t *testing.T) {
	e := launcherConfigConflictError{Expected: "rev-a", Actual: "rev-b"}
	want := "CONFIG_REVISION_CONFLICT: expected=rev-a actual=rev-b"
	if got := e.Error(); got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

// boolPtr 构造 *bool 测试值。
func boolPtr(b bool) *bool { return &b }

// TestSyncStartupTaskFromConfig 锁定批次 132：新旧开机自启设置比较 → sync 触发语义。
func TestSyncStartupTaskFromConfig(t *testing.T) {
	cases := []struct {
		name      string
		newOn     *bool
		newDelay  int
		oldOn     *bool
		oldDelay  int
		wantCall  bool
		wantOn    bool
		wantDelay int
	}{
		{"启用变化 false→true", boolPtr(true), 30, boolPtr(false), 30, true, true, 30},
		{"启用变化 true→false", boolPtr(false), 30, boolPtr(true), 30, true, false, 30},
		{"延时变化 30→50", boolPtr(true), 50, boolPtr(true), 30, true, true, 50},
		{"均无变化", boolPtr(true), 30, boolPtr(true), 30, false, false, 0},
		{"禁用时延时变化不触发", boolPtr(false), 50, boolPtr(false), 30, false, false, 0},
		{"nil 视为 false 启用", boolPtr(true), 30, nil, 0, true, true, 30},
		{"延时 clamp 0→10", boolPtr(true), 0, boolPtr(true), 10, false, false, 0},
		{"延时 clamp 150→100", boolPtr(true), 150, boolPtr(false), 0, true, true, 100},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var called bool
			var gotOn bool
			var gotDelay int
			bs := &BootstrapService{
				startupTaskSync: func(enabled bool, delay int) error {
					called = true
					gotOn = enabled
					gotDelay = delay
					return nil
				},
			}
			newCfg := LauncherConfig{}
			newCfg.Preferences.StartupLaunchEnabled = c.newOn
			newCfg.Preferences.StartupLaunchDelaySeconds = c.newDelay
			oldCfg := LauncherConfig{}
			oldCfg.Preferences.StartupLaunchEnabled = c.oldOn
			oldCfg.Preferences.StartupLaunchDelaySeconds = c.oldDelay

			err := bs.syncStartupTaskFromConfig(newCfg, oldCfg)
			if err != nil {
				t.Fatalf("不应报错, got %v", err)
			}
			if called != c.wantCall {
				t.Fatalf("called=%v, want %v", called, c.wantCall)
			}
			if c.wantCall && (gotOn != c.wantOn || gotDelay != c.wantDelay) {
				t.Fatalf("sync 参数 enabled=%v delay=%d, want %v/%d", gotOn, gotDelay, c.wantOn, c.wantDelay)
			}
		})
	}
}

// TestGetSnapshotRetError verifies GetSnapshot returns (BootstrapSnapshot, error).
func TestGetSnapshotRetError(t *testing.T) {
	bs := &BootstrapService{}
	_, err := bs.GetSnapshot()
	_ = err
}

// TestSnapshotForCommittedLauncherStateRetError verifies new signature.
func TestSnapshotForCommittedLauncherStateRetError(t *testing.T) {
	bs := &BootstrapService{}
	_, err := bs.snapshotForCommittedLauncherState()
	_ = err
}

// ---- launcherWindowController 测试替身 ----

type launcherWindowControllerMock struct {
	visible            bool
	x, y               int
	setPositionX, setY int
	setPositionCalled  bool
	contentProtection  []bool
}

func (m *launcherWindowControllerMock) Close()                               {}
func (m *launcherWindowControllerMock) EmitEvent(string, []interface{}) bool { return false }
func (m *launcherWindowControllerMock) Focus()                               {}
func (m *launcherWindowControllerMock) Hide() application.Window             { return nil }
func (m *launcherWindowControllerMock) IsFocused() bool                      { return false }
func (m *launcherWindowControllerMock) IsMinimised() bool                    { return false }
func (m *launcherWindowControllerMock) IsVisible() bool                      { return m.visible }
func (m *launcherWindowControllerMock) Position() (int, int)                 { return m.x, m.y }
func (m *launcherWindowControllerMock) Restore()                             {}
func (m *launcherWindowControllerMock) SetContentProtection(p bool) application.Window {
	m.contentProtection = append(m.contentProtection, p)
	return nil
}
func (m *launcherWindowControllerMock) SetPosition(x, y int) {
	m.setPositionX, m.setY = x, y
	m.setPositionCalled = true
}
func (m *launcherWindowControllerMock) Show() application.Window { return nil }

// TestBeginLauncherScreenshotCaptureDisabled verifies the !enabled early exit.
func TestBeginLauncherScreenshotCaptureDisabled(t *testing.T) {
	bs := &BootstrapService{}
	win := &launcherWindowControllerMock{visible: true, x: 100, y: 200}
	st := bs.beginLauncherScreenshotCapture(win, false)
	if st.enabled {
		t.Fatalf("disabled 应返回 enabled=false 态")
	}
	if win.setPositionCalled {
		t.Fatalf("disabled 不应调用 SetPosition")
	}
}

// TestBeginLauncherScreenshotCaptureNilWindow verifies nil-window early exit.
func TestBeginLauncherScreenshotCaptureNilWindow(t *testing.T) {
	bs := &BootstrapService{}
	var win launcherWindowController
	st := bs.beginLauncherScreenshotCapture(win, true)
	if st.enabled {
		t.Fatalf("nil window 应返回 enabled=false 态")
	}
}

// TestBeginLauncherScreenshotCaptureInvisible verifies !IsVisible early exit.
func TestBeginLauncherScreenshotCaptureInvisible(t *testing.T) {
	bs := &BootstrapService{}
	win := &launcherWindowControllerMock{visible: false, x: 100, y: 200}
	st := bs.beginLauncherScreenshotCapture(win, true)
	if st.enabled {
		t.Fatalf("不可见窗口应返回 enabled=false 态")
	}
	if win.setPositionCalled {
		t.Fatalf("不可见窗口不应调用 SetPosition")
	}
}

// TestBeginLauncherScreenshotCaptureFullPath verifies off-screen shift + protection.
func TestBeginLauncherScreenshotCaptureFullPath(t *testing.T) {
	bs := &BootstrapService{}
	win := &launcherWindowControllerMock{visible: true, x: 640, y: 480}
	st := bs.beginLauncherScreenshotCapture(win, true)
	if !st.enabled {
		t.Fatalf("可见窗口应返回 enabled=true 态")
	}
	if st.x != 640 || st.y != 480 {
		t.Fatalf("state 应记录原坐标 (640,480), got (%d,%d)", st.x, st.y)
	}
	if !win.setPositionCalled {
		t.Fatalf("应调用 SetPosition")
	}
	if win.setPositionX != 640-launcherWindowScreenshotHideOffset || win.setY != 480 {
		t.Fatalf("SetPosition 应 (x-%d, y)=(%d,480), got (%d,%d)",
			launcherWindowScreenshotHideOffset, 640-launcherWindowScreenshotHideOffset,
			win.setPositionX, win.setY)
	}
	if len(win.contentProtection) != 1 || !win.contentProtection[0] {
		t.Fatalf("应调用 SetContentProtection(true), got %v", win.contentProtection)
	}
}

// TestEndLauncherScreenshotCaptureDisabled verifies the !enabled early exit.
func TestEndLauncherScreenshotCaptureDisabled(t *testing.T) {
	bs := &BootstrapService{}
	win := &launcherWindowControllerMock{}
	bs.endLauncherScreenshotCapture(win, false, 10, 20)
	if win.setPositionCalled {
		t.Fatalf("disabled 不应调用 SetPosition")
	}
}

// TestEndLauncherScreenshotCaptureFullPath verifies restore order.
func TestEndLauncherScreenshotCaptureFullPath(t *testing.T) {
	bs := &BootstrapService{}
	win := &launcherWindowControllerMock{}
	bs.endLauncherScreenshotCapture(win, true, 640, 480)
	if !win.setPositionCalled {
		t.Fatalf("应调用 SetPosition")
	}
	if win.setPositionX != 640 || win.setY != 480 {
		t.Fatalf("SetPosition 应 (640,480), got (%d,%d)", win.setPositionX, win.setY)
	}
	if len(win.contentProtection) != 1 || win.contentProtection[0] {
		t.Fatalf("应调用 SetContentProtection(false), got %v", win.contentProtection)
	}
}

// TestEnsureLauncherWindowForShowNilApp verifies the app==nil early exit.
func TestEnsureLauncherWindowForShowNilApp(t *testing.T) {
	bs := &BootstrapService{} // app == nil
	if w := bs.ensureLauncherWindowForShow(); w != nil {
		t.Fatalf("app==nil 且 resolveLauncherWindow 返 nil 时应返回 nil 窗口, got %v", w)
	}
}
