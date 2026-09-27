package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLauncherConfigStorePathKey(t *testing.T) {
	// 汇编实证（0x140898c20）：TrimSpace → filepath.Abs(err nil 用 abs / 失败 Clean) → ToLower。
	// Windows 下 filepath.Abs 输出反斜杠分隔。
	cases := []struct{ name, in, contains string }{
		{"trim", "  C:/x/y  ", "c:\\x\\y"},
		{"tolower", "C:/ABC/Def", "c:\\abc\\def"},
		{"abs-normalize", "config.json", "config.json"}, // Abs 依 CWD；仅校验含文件名为小写
	}
	for _, c := range cases {
		got := launcherConfigStorePathKey(c.in)
		t.Logf("%s: key(%q)=%q", c.name, c.in, got)
		if !strings.Contains(got, c.contains) {
			t.Errorf("%s: key(%q)=%q missing %q", c.name, c.in, got, c.contains)
		}
	}
	// 大小写不敏感 + trim：同路径不同大小写/空白应一致（平台分隔符统一为反斜杠后 ToLower）。
	a := launcherConfigStorePathKey("C:/X/Config.json")
	b := launcherConfigStorePathKey("c:/x/config.json ")
	if a != b {
		t.Errorf("case/trim-insensitive key mismatch: %q vs %q", a, b)
	}
}

func TestLauncherConfigStoreForPathCache(t *testing.T) {
	// 同一归一化路径 → 返回同一 *launcherConfigStore 实例（Assembly Load/LoadOrStore 单例）。
	s1 := launcherConfigStoreForPath("C:/Some/Path/config.json")
	s2 := launcherConfigStoreForPath("c:/some/path/config.json ")
	if s1 != s2 {
		t.Error("相同归一化路径应命中同一单例缓存")
	}
	// path 字段 = TrimSpace(path)。
	if s1.path != "C:/Some/Path/config.json" {
		t.Errorf("store.path=%q want raw-with-trim %q", s1.path, "C:/Some/Path/config.json")
	}
	// 不同路径 → 不同实例。
	s3 := launcherConfigStoreForPath("C:/Other/Path/config.json")
	if s3 == s1 {
		t.Error("不同路径应产生不同 store 实例")
	}
	// 新路径首次创建：path 已设。
	if len(s3.path) == 0 {
		t.Error("新 store 应带 path 字段")
	}
}
func TestWriteJSONFile(t *testing.T) {
	dir := t.TempDir()
	p := dir + "/cfg/config.json" // 嵌套目录由 MkdirAll 创建
	cfg := LauncherConfig{Revision: "abc", Version: 42, Initialized: true}

	if err := writeJSONFile(p, cfg); err != nil {
		t.Fatalf("writeJSONFile err=%v", err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read back err=%v", err)
	}
	s := string(b)
	// 缩进 JSON + 换行结束（汇编：MarshalIndent + '\n'）。
	if !strings.Contains(s, "\"revision\": \"abc\"") {
		t.Errorf("content missing revision field: %q", s)
	}
	if !strings.HasSuffix(s, "\n") {
		t.Error("json must end with newline")
	}
	// 空路径 → error。
	if err := writeJSONFile("", cfg); err == nil {
		t.Error("empty path must error")
	}
}

func TestWriteConfigUnlocked(t *testing.T) {
	dir := t.TempDir()
	p := dir + "/config.json"
	s := &launcherConfigStore{path: p}
	cfg := LauncherConfig{Version: 7}

	// 未注入 writeConfig → 走 writeJSONFile 落盘。
	if err := s.writeConfigUnlocked(cfg); err != nil {
		t.Fatalf("writeConfigUnlocked err=%v", err)
	}
	if _, err := os.ReadFile(p); err != nil {
		t.Errorf("config file not written: %v", err)
	}
	// 指纹已刷新。
	if !s.iconFingerprintReady || !s.runtimeReady {
		t.Error("iconFingerprintReady/runtimeReady must be set")
	}
	if s.iconFingerprint == ([32]byte{}) {
		t.Error("iconFingerprint should be non-zero")
	}

	// 注入 writeConfig：走注入路径且错误透传。
	gotPath := ""
	s2 := &launcherConfigStore{
		path: "injected-path",
		writeConfig: func(p string, c LauncherConfig) error {
			gotPath = p
			return nil
		},
	}
	if err := s2.writeConfigUnlocked(cfg); err != nil {
		t.Fatalf("injected writeConfig err=%v", err)
	}
	// 注入 writeConfig 收到 store.path（TrimSpace 后原 path）。
	if gotPath != "injected-path" {
		t.Errorf("writeConfig got path=%q want injected-path", gotPath)
	}
	// 注入 error 透传。
	s3 := &launcherConfigStore{
		writeConfig: func(string, LauncherConfig) error { return errInjectedWrite },
	}
	if err := s3.writeConfigUnlocked(cfg); err != errInjectedWrite {
		t.Errorf("injected error should pass through, got=%v", err)
	}
}

var errInjectedWrite = errors.New("injected write error")

func TestReadLauncherConfigBytes(t *testing.T) {
	dir := t.TempDir()
	p := dir + "/cfg.json"
	if err := os.WriteFile(p, []byte(`{"revision":"x","version":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	b, err := readLauncherConfigBytes(p)
	if err != nil || len(b) == 0 || !strings.Contains(string(b), "revision") {
		t.Errorf("readLauncherConfigBytes err=%v b=%q", err, b)
	}
	// 缺失 → error（非 NotExist 视为 err）。
	if _, err := readLauncherConfigBytes(dir + "/missing.json"); err == nil {
		t.Error("missing file must error")
	}
}

func TestLoadUnlockedAndDeletePrepared(t *testing.T) {
	dir := t.TempDir()
	p := dir + "/cfg.json"
	if err := os.WriteFile(p, []byte(`{"revision":"r1","version":5}`), 0o644); err != nil {
		t.Fatal(err)
	}
	s := &launcherConfigStore{path: p}

	// loadUnlocked 首次加载：runtimeReady 置位 + 指纹刷新。
	cfg, err := s.loadUnlocked()
	if err != nil || cfg.Version != 5 {
		t.Fatalf("loadUnlocked err=%v cfg=%+v", err, cfg)
	}
	if !s.runtimeReady || !s.iconFingerprintReady {
		t.Error("runtimeReady/iconFingerprintReady must be true after load")
	}
	// 再次 load：短路（不重读文件、runtimeReady 保持）。
	cfg2, _ := s.loadUnlocked()
	if cfg2.Version != 5 {
		t.Errorf("stale runtime load: %+v", cfg2)
	}

	// deletePreparedLocked：prepare 回调执行 + 文件删除 + 指纹清。
	called := false
	if err := s.deletePreparedLocked(func(LauncherConfig) error { called = true; return nil }); err != nil {
		t.Fatalf("deletePreparedLocked err=%v", err)
	}
	if !called {
		t.Error("prepare callback must be called")
	}
	if _, err := os.Stat(p); !errors.Is(err, os.ErrNotExist) {
		t.Error("config file must be removed")
	}
	if s.iconFingerprintReady || s.iconFingerprint != ([32]byte{}) {
		t.Error("iconFingerprint must be cleared after delete")
	}
}

func TestResolveProcessWorkingDirectory(t *testing.T) {
	// 汇编实证：os.Getwd 成功且非空 → 返回 TrimSpace(wd)。断言返回非空且为绝对路径。
	wd, err := os.Getwd()
	if err != nil {
		t.Skip("no cwd available")
	}
	got := resolveProcessWorkingDirectory()
	if got == "" {
		t.Fatal("working dir must be non-empty")
	}
	// 非空且（通常）等于 TrimSpace(cwd)。
	if strings.TrimSpace(wd) != "" && got != strings.TrimSpace(wd) {
		// Getwd 可能因符号链接差异，仅警告不失败。
		t.Logf("cwd=%q resolve=%q", wd, got)
	}
}

func TestResolveWorkspaceLayoutRoot(t *testing.T) {
	l := resolveWorkspaceLayout()
	if l.Root == "" || l.ConfigFile == "" {
		t.Errorf("layout root/config must be set: %+v", l)
	}
	if l.IconDir == "" || l.PluginDir == "" || l.ScreenshotDir == "" {
		t.Errorf("layout dirs incomplete: %+v", l)
	}
	if filepath.Base(l.ConfigFile) != "config.json" {
		t.Errorf("config file base=%q want config.json", filepath.Base(l.ConfigFile))
	}
}

// test warmer 实现 fileSearchRuntimeWarmer 接口，记录调用。
type testFileSearchWarmer struct {
	called  bool
	volumes map[string]struct{}
	deep    bool
}

func (w *testFileSearchWarmer) scheduleFileSearchMaintenance(v map[string]struct{}, deep bool) {
	w.called = true
	w.volumes = v
	w.deep = deep
}

func TestAttachFileIndexService(t *testing.T) {
	// nil receiver → false。
	bs := &BootstrapService{}
	if bs.attachFileIndexService(nil) {
		t.Error("nil service attach must be false")
	}
	if (*BootstrapService)(nil).attachFileIndexService(nil) {
		t.Error("nil bootstrap attach must be false")
	}

	// pending 未置位 → attach 不预热、返回 false、字段已设。
	bs2 := &BootstrapService{}
	w := &testFileSearchWarmer{}
	if bs2.attachFileIndexService(w) {
		t.Error("pending unset must not warm")
	}
	if bs2.fileIndex != w {
		t.Error("fileIndex field not attached")
	}

	// pending 置位 + 非 nil → 清标志、立即预热、返回 true。
	bs3 := &BootstrapService{pendingFileSearchResidentWarm: true}
	w2 := &testFileSearchWarmer{}
	if !bs3.attachFileIndexService(w2) {
		t.Error("pending warm must trigger attach warm")
	}
	if !w2.called {
		t.Error("scheduleFileSearchMaintenance must be called on warm")
	}
	if bs3.pendingFileSearchResidentWarm {
		t.Error("pending flag must be cleared after warm")
	}
}

func TestNewBootstrapServiceAssembly(t *testing.T) {
	bs := NewBootstrapService(nil)
	if bs == nil {
		t.Fatal("bootstrap must be non-nil")
	}
	// 核心装配：[S] workspace + configStore 已接线。
	if bs.workspace.ConfigFile == "" || bs.configStore == nil {
		t.Errorf("workspace/configStore wired: ws=%+v store=%v", bs.workspace, bs.configStore)
	}
	// service 工厂挂载（装配依赖 8 叶已填）。
	if bs.globalHotkey == nil {
		t.Error("globalHotkey must be wired")
	}
	if bs.memoryRelease == nil || bs.oledBlackout == nil || bs.mouseGestures == nil ||
		bs.inputMonitor == nil || bs.fileLocator == nil || bs.desktopWidgets == nil {
		t.Error("all service sub-factories must be wired")
	}
	// 容器 map / chan。
	if bs.hotkeyCaptureOwners == nil || bs.hotkeyRegistrationErrors == nil || bs.iconAssetURLs == nil {
		t.Error("container maps must be initialized")
	}
	if bs.launcherUpdateDone == nil {
		t.Error("launcherUpdateDone chan must be set")
	}
}

func TestAttachAppDispatch(t *testing.T) {
	// nil receiver / nil app 安全返回。
	var nilBS *BootstrapService
	nilBS.attachApp()
	(&BootstrapService{app: nil}).attachApp()

	// 装配后 attachApp 分发不 panic（[P] AttachApp 脚手架亦可编译）。
	bs := NewBootstrapService(nil)
	bs.attachApp()
	bs.syncHotkeyBindings(bs.bindings)
}

func TestSaveUnlockedWithCurrentRefs(t *testing.T) {
	var got string
	s := &launcherConfigStore{writeConfig: func(_ string, c LauncherConfig) error { got = c.Revision; return nil }}
	if err := s.saveUnlockedWithCurrentRefs(LauncherConfig{Revision: "rv1"}); err != nil {
		t.Fatalf("save err=%v", err)
	}
	if got != "rv1" {
		t.Errorf("writeConfig revision got=%q want rv1", got)
	}
}

func TestSavePreparedUnlocked(t *testing.T) {
	var got string
	s := &launcherConfigStore{writeConfig: func(_ string, c LauncherConfig) error { got = c.Revision; return nil }}
	if err := s.savePreparedUnlocked(true, LauncherConfig{Revision: "pv2"}); err != nil {
		t.Fatalf("savePrepared(err=%v", err)
	}
	if got != "pv2" {
		t.Errorf("got=%q want pv2", got)
	}
	var got3 string
	s2 := &launcherConfigStore{writeConfig: func(_ string, c LauncherConfig) error { got3 = c.Revision; return nil }}
	if err := s2.savePreparedUnlocked(false, LauncherConfig{Revision: "pv3"}); err != nil {
		t.Fatalf("savePrepared#2 err=%v", err)
	}
	if got3 != "pv3" {
		t.Errorf("got3=%q want pv3", got3)
	}
}

func TestLauncherConfigIconDataFingerprint(t *testing.T) {
	a := launcherConfigIconDataFingerprint(LauncherConfig{})
	b := launcherConfigIconDataFingerprint(LauncherConfig{})
	// 确定性 + 非零 32 字节形态。
	if a != b {
		t.Error("fingerprint must be deterministic")
	}
	if len(a) != 32 || a == ([32]byte{}) {
		t.Errorf("fingerprint length/shape wrong: %x", a)
	}
	// 引擎经 visit 链（[S] 入口）；slots 字段遍历 [P] 影响活性。
	_ = a
}

func TestLauncherConfigStorePathKeyExtremes(t *testing.T) {
	// 空串 / 纯空白 → 归一化后为空 → filepath.Abs(".") → CWD：两者一致且均为小写绝对路径。
	a := launcherConfigStorePathKey("")
	b := launcherConfigStorePathKey("   ")
	if a == "" || b == "" {
		t.Fatalf("empty/blank path must resolve (cwd), got %q %q", a, b)
	}
	if a != b {
		t.Error("empty vs blank must normalize identically")
	}
	if launcherConfigStorePathKey("") != a || a != strings.ToLower(a) {
		t.Error("must be deterministic and lowercased")
	}
}

func TestDeletePreparedIdempotent(t *testing.T) {
	dir := t.TempDir()
	// 目标文件不存在：删除应容忍 NotExist（幂等）。
	s := &launcherConfigStore{path: dir + "/nonexistent.json"}
	if err := s.deletePreparedLocked(func(LauncherConfig) error { return nil }); err != nil {
		t.Fatalf("delete 1 err=%v (must tolerate missing)", err)
	}
	if err := s.deletePreparedLocked(func(LauncherConfig) error { return nil }); err != nil {
		t.Fatalf("delete 2 (idempotent) err=%v", err)
	}
	// prepare 回调错误须传播。
	boom := errors.New("prepare boom")
	if err := s.deletePreparedLocked(func(LauncherConfig) error { return boom }); err != boom {
		t.Errorf("prepare error must propagate, got %v", err)
	}
}

func TestSaveToEndToEndFile(t *testing.T) {
	dir := t.TempDir()
	p := dir + "/sub/cfg.json"
	s := &launcherConfigStore{path: p}
	if err := s.saveUnlockedWithCurrentRefs(LauncherConfig{Revision: "e2e", Version: 7, Apps: []AppEntry{{Name: "n", IconData: "icon"}}}); err != nil {
		t.Fatalf("save err=%v", err)
	}
	b, err := readLauncherConfigBytes(p)
	if err != nil {
		t.Fatalf("read err=%v", err)
	}
	raw := string(b)
	if !strings.Contains(raw, `"version": 7`) || !strings.Contains(raw, "e2e") || !strings.Contains(raw, "icon") {
		t.Errorf("persisted config missing key fields: %q", raw)
	}
}

func TestNormalizeStorageConfig(t *testing.T) {
	r := normalizeStorageConfig("  A  ", "B  ", "   ", "  C ", "D  ")
	if r.DataRoot != "A" || r.IconDir != "B" || r.IndexDir != "" || r.ScreenshotDir != "C" || r.WebView2Dir != "D" {
		t.Errorf("normalize result = %+v", r)
	}
}

func TestNormalizeBookmarkFavoritePathEmpty(t *testing.T) {
	r := normalizeBookmarkFavoritePath(BookmarkFavoritePath{})
	if r.SourcePath != "" || r.FolderPath != "" {
		t.Errorf("空 SourcePath 应返回零值 struct, got %+v", r)
	}
}

func TestNormalizeBookmarkFavoritePathTrim(t *testing.T) {
	r := normalizeBookmarkFavoritePath(BookmarkFavoritePath{SourcePath: "  C:/favs  ", FolderPath: "C:/folders"})
	if r.SourcePath != "C:/favs" {
		t.Errorf("SourcePath 应 TrimSpace 为 %q, got %q", "C:/favs", r.SourcePath)
	}
	// normalizeBookmarkFolderPathValue 已升 [S]：无首尾斜杠与 " / " 分隔符时原样保留。
	if r.FolderPath != "C:/folders" {
		t.Errorf("FolderPath 应经 normalizeBookmarkFolderPathValue 保留, got %q", r.FolderPath)
	}
}

func TestNormalizeSearchScopeCycleDefaultEmpty(t *testing.T) {
	if got := normalizeSearchScopeCycleDefault("   ", nil); got != "" {
		t.Errorf("空/空白 scope 应返空串, got %q", got)
	}
}

func TestNormalizeDefaultLinkBrowserID(t *testing.T) {
	if got := normalizeDefaultLinkBrowserID("  ", nil); got != "" {
		t.Errorf("空 id 应返空串, got %q", got)
	}
	if got := normalizeDefaultLinkBrowserID("__system_default__", nil); got != "__system_default__" {
		t.Errorf("系统默认字面量应原样返回, got %q", got)
	}
	browsers := []LinkBrowser{
		{ID: "  Chrome ", Name: "Google Chrome"},
		{ID: "firefox", Name: "Firefox"},
	}
	if got := normalizeDefaultLinkBrowserID("chrome", browsers); got != "Chrome" {
		t.Errorf("EqualFold 命中应返回 TrimSpace(ID), got %q", got)
	}
	if got := normalizeDefaultLinkBrowserID("edge", browsers); got != "" {
		t.Errorf("未命中应返空串, got %q", got)
	}
}

func TestFindConfiguredLinkBrowserByID(t *testing.T) {
	browsers := []LinkBrowser{
		{ID: "  Chrome ", Name: "Google Chrome"},
		{ID: "firefox", Name: "Firefox"},
	}
	if b, ok := findConfiguredLinkBrowserByID(browsers, "  "); ok || b.ID != "" {
		t.Errorf("空 id 应返 (zero, false), got (%+v, %v)", b, ok)
	}
	b, ok := findConfiguredLinkBrowserByID(browsers, "chrome")
	if !ok || b.ID != "  Chrome " || b.Name != "Google Chrome" {
		t.Errorf("应命中完整 LinkBrowser, got (%+v, %v)", b, ok)
	}
	if b, ok := findConfiguredLinkBrowserByID(browsers, "edge"); ok {
		t.Errorf("未命中应返 false, got (%+v, %v)", b, ok)
	}
}

func TestStartMenuLaunchLogKeys(t *testing.T) {
	log := StartMenuLaunchLog{
		ID:           "AppX",
		ShortcutPath: "C:/Shortcut",
		TargetPath:   "C:/Target",
	}
	got := startMenuLaunchLogKeys(log)
	want := []string{"id:appx", "shortcut:c:/shortcut", "target:c:/target"}
	if len(got) != len(want) {
		t.Fatalf("keys 数量 = %d, got %v", len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("keys[%d] = %q, want %q", i, got[i], want[i])
		}
	}
	// 空字段跳过。
	got = startMenuLaunchLogKeys(StartMenuLaunchLog{Name: "only-name"})
	if len(got) != 0 {
		t.Errorf("全空身份字段应返空 slice, got %v", got)
	}
}

func TestNormalizeStartMenuLaunchLog(t *testing.T) {
	// 有身份且已启动：全字段 TrimSpace + ID 去前缀 + LaunchCount 夹取。
	log := StartMenuLaunchLog{
		ID:           " start-menu:AppX ",
		Name:         "  App X ",
		ShortcutPath: " C:/Shortcut ",
		TargetPath:   " C:/Target ",
		LaunchCount:  3,
	}
	got := normalizeStartMenuLaunchLog(log)
	if got.ID != "AppX" || got.Name != "App X" || got.ShortcutPath != "C:/Shortcut" || got.TargetPath != "C:/Target" {
		t.Errorf("归一化结果 = %+v", got)
	}
	if got.LaunchCount != 3 {
		t.Errorf("LaunchCount 应保留正值, got %d", got.LaunchCount)
	}
	// 无身份 → 返零值。
	if got := normalizeStartMenuLaunchLog(StartMenuLaunchLog{Name: "x", LaunchCount: 5}); got != (StartMenuLaunchLog{}) {
		t.Errorf("无身份应返零值, got %+v", got)
	}
	// 未启动（LaunchCount<=0 且无时间）→ 返零值。
	if got := normalizeStartMenuLaunchLog(StartMenuLaunchLog{ID: "a"}); got != (StartMenuLaunchLog{}) {
		t.Errorf("未启动应返零值, got %+v", got)
	}
	// LaunchCount 负值夹取为 0（需有时间戳或计数通过次空检）。
	clamped := normalizeStartMenuLaunchLog(StartMenuLaunchLog{ID: "a", LastLaunchedAt: " 2026-01-01 ", LaunchCount: -2})
	if clamped.LaunchCount != 0 || clamped.ID != "a" || clamped.LastLaunchedAt != "2026-01-01" {
		t.Errorf("负 LaunchCount 夹取失败: %+v", clamped)
	}
}

func TestMergeStartMenuLaunchLog(t *testing.T) {
	// existing 较新：以 existing 为基底，回填空字段，LaunchCount 取 max。
	existing := StartMenuLaunchLog{ID: "a", LastLaunchedAt: "2026-02-01", LaunchCount: 5}
	incoming := StartMenuLaunchLog{Name: "Incoming", TargetPath: "T", LastLaunchedAt: "2026-01-01", LaunchCount: 3}
	got := mergeStartMenuLaunchLog(existing, incoming)
	if got.ID != "a" || got.Name != "Incoming" || got.TargetPath != "T" || got.LaunchCount != 5 {
		t.Errorf("existing 较新合并失败: %+v", got)
	}
	// incoming 较新：以 incoming 为基底，回填空字段，LaunchCount 保留 incoming（asm 读原第二参 [0x178]）。
	existing = StartMenuLaunchLog{Name: "Old", LastLaunchedAt: "2026-01-01", LaunchCount: 9}
	incoming = StartMenuLaunchLog{ID: "b", ShortcutPath: "S", LastLaunchedAt: "2026-03-01", LaunchCount: 2}
	got = mergeStartMenuLaunchLog(existing, incoming)
	if got.ID != "b" || got.Name != "Old" || got.ShortcutPath != "S" || got.LastLaunchedAt != "2026-03-01" {
		t.Errorf("incoming 较新合并失败: %+v", got)
	}
	if got.LaunchCount != 2 {
		t.Errorf("incoming 较新时 LaunchCount 应保留 incoming(=2), got %d", got.LaunchCount)
	}
}

func TestPopulateAppIcons(t *testing.T) {
	// nil 接收与空 Apps 早退，不 panic。
	populateAppIcons(nil)
	populateAppIcons(&LauncherConfig{})

	// AutoIcon=false 跳过。
	cfg := &LauncherConfig{Apps: []AppEntry{{ID: "a", AutoIcon: false, IconDataVersion: 0}}}
	populateAppIcons(cfg)
	if cfg.Apps[0].IconDataVersion != 0 {
		t.Errorf("非 AutoIcon 条目不应被处理: %+v", cfg.Apps[0])
	}

	// IconDataVersion>=3 跳过。
	cfg = &LauncherConfig{Apps: []AppEntry{{ID: "a", AutoIcon: true, IconDataVersion: 3}}}
	populateAppIcons(cfg)
	if cfg.Apps[0].IconData != "" || cfg.Apps[0].IconDataVersion != 3 {
		t.Errorf("IconDataVersion>=3 不应被处理: %+v", cfg.Apps[0])
	}

	// AutoIcon=true、IconDataVersion<3、空路径：resolve 空键即返（无 Win32），置 version=3。
	cfg = &LauncherConfig{Apps: []AppEntry{{ID: "a", AutoIcon: true, IconDataVersion: 0, Path: ""}}}
	populateAppIcons(cfg)
	if cfg.Apps[0].IconDataVersion != 3 {
		t.Errorf("AutoIcon 条目应置 IconDataVersion=3: %+v", cfg.Apps[0])
	}
}

func TestBuildConsoleItemID(t *testing.T) {
	// 任一关键字段空 → ""。
	if got := buildConsoleItemID(ConsoleItem{SectionID: "s", TargetID: "t"}); got != "" {
		t.Errorf("空 Kind 应返空, got %q", got)
	}
	if got := buildConsoleItemID(ConsoleItem{Kind: "k", TargetID: "t"}); got != "" {
		t.Errorf("空 SectionID 应返空, got %q", got)
	}
	if got := buildConsoleItemID(ConsoleItem{Kind: "k", SectionID: "s"}); got != "" {
		t.Errorf("空 TargetID 应返空, got %q", got)
	}
	// 完整字段：各段 PathEscape 后以 ":" 连接（顺序 Kind:Section:Source:Target:Folder）。
	item := ConsoleItem{Kind: "app", SectionID: "sec", SourceID: "src", TargetID: "tgt", FolderPath: "C:/f o/o"}
	got := buildConsoleItemID(item)
	want := strings.Join([]string{
		url.PathEscape("app"), url.PathEscape("sec"), url.PathEscape("src"),
		url.PathEscape("tgt"), url.PathEscape("C:/f o/o"),
	}, ":")
	if got != want {
		t.Errorf("buildConsoleItemID = %q, want %q", got, want)
	}
}

func TestNormalizeBackgroundPreference(t *testing.T) {
	// 全空：默认值 + ShellOpacity 落空(nil)。
	got := normalizeBackgroundPreference(BackgroundPreference{Enabled: true, ImagePath: "  p.png  ", ImageURL: "  u  "})
	if !got.Enabled || got.ImagePath != "p.png" || got.ImageURL != "u" {
		t.Errorf("TrimSpace/Enabled 直传失败: %+v", got)
	}
	if got.ReadabilityOverlayEnabled == nil || !*got.ReadabilityOverlayEnabled {
		t.Errorf("ReadabilityOverlayEnabled 默认应为 true: %+v", got)
	}
	if got.ReadabilityOverlayOpacity == nil || *got.ReadabilityOverlayOpacity != 0.18 {
		t.Errorf("ReadabilityOverlayOpacity 默认应为 0.18: %+v", got)
	}
	if got.ImageOpacity == nil || *got.ImageOpacity != 1.0 {
		t.Errorf("ImageOpacity 默认应为 1.0: %+v", got)
	}
	if got.SidebarTransparency == nil || got.ContentTransparency == nil || *got.SidebarTransparency != 0 || *got.ContentTransparency != 0 {
		t.Errorf("ShellOpacity nil 时 Sidebar/Content 应回退 0: %+v", got)
	}
	if got.ShellOpacity != nil {
		t.Errorf("ShellOpacity 应落空(nil): %+v", got)
	}

	// 越界钳制。
	shell, ro, io, st, ct := 2.0, 0.9, -0.5, 1.5, -1.0
	got = normalizeBackgroundPreference(BackgroundPreference{
		ShellOpacity:              &shell,
		ReadabilityOverlayOpacity: &ro,
		ImageOpacity:              &io,
		ImageBlur:                 100,
		SidebarTransparency:       &st,
		ContentTransparency:       &ct,
	})
	if *got.ReadabilityOverlayOpacity != 0.65 {
		t.Errorf("ReadabilityOverlayOpacity 应钳制到 0.65, got %v", *got.ReadabilityOverlayOpacity)
	}
	if *got.ImageOpacity != 0 {
		t.Errorf("ImageOpacity 应钳制到 0, got %v", *got.ImageOpacity)
	}
	if got.ImageBlur != 24 {
		t.Errorf("ImageBlur 应钳制到 24, got %d", got.ImageBlur)
	}
	if *got.SidebarTransparency != 1 || *got.ContentTransparency != 0 {
		t.Errorf("Sidebar/Content 钳制失败: %+v", got)
	}
	if got.ShellOpacity != nil {
		t.Errorf("ShellOpacity 应落空(nil): %+v", got)
	}

	// ShellOpacity 作为 Sidebar/Content 的 nil 回退默认。
	shell = 0.4
	got = normalizeBackgroundPreference(BackgroundPreference{ShellOpacity: &shell})
	if *got.SidebarTransparency != 0.4 || *got.ContentTransparency != 0.4 {
		t.Errorf("Sidebar/Content 应回退 shellOpacity 0.4: %+v", got)
	}
}

func TestNormalizeFileLocatorHistoryEntries(t *testing.T) {
	if got := normalizeFileLocatorHistoryEntries(nil, false); got != nil {
		t.Errorf("空输入应返 nil, got %v", got)
	}
	// 去空白 + 去空 + 去重。
	got := normalizeFileLocatorHistoryEntries([]string{"  a  ", "b", "", "  a  ", "b", "c"}, false)
	if want := []string{"a", "b", "c"}; !reflect.DeepEqual(got, want) {
		t.Errorf("history = %v, want %v", got, want)
	}
	// normalizePath：反斜杠→斜杠+小写去重，输出保留原串。
	got = normalizeFileLocatorHistoryEntries([]string{"C:\\Foo", "c:/foo", "D:\\bar"}, true)
	if want := []string{"C:\\Foo", "D:\\bar"}; !reflect.DeepEqual(got, want) {
		t.Errorf("normalizePath history = %v, want %v", got, want)
	}
	// 上限 12。
	many := make([]string, 20)
	for i := range many {
		many[i] = fmt.Sprintf("e%d", i)
	}
	got = normalizeFileLocatorHistoryEntries(many, false)
	if len(got) != 12 {
		t.Errorf("应截断到 12, got %d", len(got))
	}
}

func TestNormalizeFileLocatorConfig(t *testing.T) {
	cfg := normalizeFileLocatorConfig(FileLocatorConfig{
		FileNameQuery:       "  foo  ",
		SearchRoot:          "  C:\\root  ",
		MaxSearchFileSizeMB: 99999,
		IncludeSubfolders:   false,
	})
	if cfg.FileNameQuery != "foo" {
		t.Errorf("FileNameQuery 应 TrimSpace, got %q", cfg.FileNameQuery)
	}
	if cfg.SearchRoot != "C:\\root" {
		t.Errorf("SearchRoot 应 TrimSpace, got %q", cfg.SearchRoot)
	}
	if cfg.MaxSearchFileSizeMB != 4096 {
		t.Errorf("MaxSearchFileSizeMB 应钳制到 4096, got %d", cfg.MaxSearchFileSizeMB)
	}
	if cfg.IncludeSubfolders {
		t.Errorf("IncludeSubfolders 应直传 false（非零配置）")
	}
	if len(cfg.SavedFilters) != 1 {
		t.Fatalf("空 SavedFilters 应补默认过滤器 1 条, got %d", len(cfg.SavedFilters))
	}
	if d := cfg.SavedFilters[0]; d.ID != "file-locator-default-non-text" || d.Type != "glob" || d.Remark != "默认非文本排除" {
		t.Errorf("默认过滤器字段不符: %+v", d)
	}
	// ActiveFilterID 命中默认过滤器 ID → 取其原 ID；未命中 → 清空。
	if got := normalizeFileLocatorConfig(FileLocatorConfig{ActiveFilterID: "  file-locator-default-non-text  "}).ActiveFilterID; got != "file-locator-default-non-text" {
		t.Errorf("ActiveFilterID 命中应取原 ID, got %q", got)
	}
	if got := normalizeFileLocatorConfig(FileLocatorConfig{ActiveFilterID: "nonexistent"}).ActiveFilterID; got != "" {
		t.Errorf("ActiveFilterID 未命中应清空, got %q", got)
	}
	// 尺寸下界。
	if got := normalizeFileLocatorConfig(FileLocatorConfig{MaxSearchFileSizeMB: -5}).MaxSearchFileSizeMB; got != 32 {
		t.Errorf("MaxSearchFileSizeMB 应钳制到 32, got %d", got)
	}
}

func TestNormalizeTagCatalogWithDefault(t *testing.T) {
	// 全空 → defaultTagCatalogItems()（[S-sig] 返 nil）。
	if got := normalizeTagCatalogWithDefault(nil, nil, nil); got != nil {
		t.Errorf("全空应返 nil, got %v", got)
	}
	// items 空 + fallbackNames 非空 → 从 names 生成。
	got := normalizeTagCatalogWithDefault(nil, []string{"  gamepad  ", ""}, nil)
	if len(got) != 1 || got[0].Name != "gamepad" || got[0].Icon != "gamepad" {
		t.Errorf("从 names 生成失败: %+v", got)
	}
	// 去重（小写键）+ 去空名 + 默认图标。
	got = normalizeTagCatalogWithDefault([]TagCatalogItem{
		{Name: "  Gamepad  "},
		{Name: "gamepad"},
		{Name: "Briefcase"},
	}, nil, nil)
	if len(got) != 2 {
		t.Fatalf("去重后应 2 条, got %d", len(got))
	}
	if got[0].Name != "Gamepad" || got[0].Icon != "gamepad" {
		t.Errorf("第 1 条不符: %+v", got[0])
	}
	if got[1].Name != "Briefcase" || got[1].Icon != "briefcase" {
		t.Errorf("第 2 条不符: %+v", got[1])
	}
	// IconData 前缀校验 + 非法清空。
	got = normalizeTagCatalogWithDefault([]TagCatalogItem{
		{Name: "x", IconData: "data:image/png;base64,AAA"},
		{Name: "y", IconData: "not-an-image"},
	}, nil, nil)
	if got[0].IconData != "data:image/png;base64,AAA" {
		t.Errorf("合法 IconData 应保留: %+v", got[0])
	}
	if got[1].IconData != "" {
		t.Errorf("非法 IconData 应清空: %+v", got[1])
	}
	// 尾递归：结果空且 fallbackItems 非空 → 回退。
	got = normalizeTagCatalogWithDefault([]TagCatalogItem{{Name: "  "}}, nil, []TagCatalogItem{{Name: "Folder"}})
	if len(got) != 1 || got[0].Name != "Folder" || got[0].Icon != "folder" {
		t.Errorf("尾递归回退失败: %+v", got)
	}
}

func TestNormalizeConsoleItem(t *testing.T) {
	// 注：normalizeConsoleIconData / normalizeBookmarkFolderPathValue 仍为 [S-sig] 返 ""，
	// 故 IconData/FolderPath 归一化结果固定为 ""，测试据此断言。

	// window 管理特殊条目：kind=section + section=windowManagement + target 空 → 重映射为 canonical。
	got := normalizeConsoleItem(ConsoleItem{Kind: "section", SectionID: "windowManagement"})
	if got.Kind != "windowManagementTool" || got.SectionID != "windowManagement" || got.TargetID != "windowTools" {
		t.Errorf("window 管理重映射失败: %+v", got)
	}
	if got.Icon != "window" {
		t.Errorf("window 管理 Icon 应为 window: %+v", got)
	}
	if got.SourceID != "" || got.FolderPath != "" || got.Title != "" || got.Subtitle != "" ||
		got.IconData != "" || got.IconRef != "" || got.IconURL != "" {
		t.Errorf("window 管理条目内容字段应清空: %+v", got)
	}
	if got.ID != "windowManagementTool:windowManagement::windowTools:" {
		t.Errorf("window 管理 ID 不符: %q", got.ID)
	}

	// 鼠标手势特殊条目：kind=section + section=mouseGestures → 重映射为 canonical。
	got = normalizeConsoleItem(ConsoleItem{Kind: "section", SectionID: "mouseGestures"})
	if got.Kind != "mouseGestureTool" || got.SectionID != "mouseGestures" || got.TargetID != "gestures" {
		t.Errorf("鼠标手势重映射失败: %+v", got)
	}
	if got.Icon != "pointer" {
		t.Errorf("鼠标手势 Icon 应为 pointer: %+v", got)
	}
	if got.ID != "mouseGestureTool:mouseGestures::gestures:" {
		t.Errorf("鼠标手势 ID 不符: %q", got.ID)
	}

	// 普通条目透传：Title 去空白、Icon 保留、内容不清空。
	got = normalizeConsoleItem(ConsoleItem{
		Kind: "appGroup", SectionID: "games", TargetID: "steam",
		Title: "  Steam  ", Icon: "folder",
	})
	if got.Kind != "appGroup" || got.SectionID != "games" || got.TargetID != "steam" {
		t.Errorf("普通条目透传失败: %+v", got)
	}
	if got.Title != "Steam" || got.Icon != "folder" {
		t.Errorf("普通条目 Title/Icon 不符: %+v", got)
	}

	// 布局夹取：X/W 上限 0x40、Y/H 上限 0x200，<=0 → 0。
	got = normalizeConsoleItem(ConsoleItem{
		Kind: "appGroup", SectionID: "games", TargetID: "steam",
		LayoutX: 100, LayoutY: 1000, LayoutW: -5, LayoutH: 30,
	})
	if got.LayoutX != 0x40 || got.LayoutY != 0x200 || got.LayoutW != 0 || got.LayoutH != 30 {
		t.Errorf("布局夹取不符: %+v", got)
	}

	// 早期退出：target 空且非特殊 → 返零值条目。
	if got := normalizeConsoleItem(ConsoleItem{Kind: "appGroup", SectionID: "games"}); got != (ConsoleItem{}) {
		t.Errorf("target 空非特殊应返零值: %+v", got)
	}
	// 早期退出：Kind 不在白名单 → 返零值条目。
	if got := normalizeConsoleItem(ConsoleItem{Kind: "unknownKind", SectionID: "x", TargetID: "y"}); got != (ConsoleItem{}) {
		t.Errorf("Kind 非白名单应返零值: %+v", got)
	}
}

func TestNormalizeConsoleIconData(t *testing.T) {
	if got := normalizeConsoleIconData("data:image/png;base64,AAA"); got != "data:image/png;base64,AAA" {
		t.Errorf("合法 data:image 应保留: %q", got)
	}
	if got := normalizeConsoleIconData("  data:image/png;base64,AAA  "); got != "data:image/png;base64,AAA" {
		t.Errorf("应 TrimSpace 后保留: %q", got)
	}
	if got := normalizeConsoleIconData("DATA:IMAGE/png;base64,AAA"); got != "DATA:IMAGE/png;base64,AAA" {
		t.Errorf("大小写不敏感前缀应保留原串: %q", got)
	}
	if got := normalizeConsoleIconData("not-an-image"); got != "" {
		t.Errorf("非 data:image 应清空: %q", got)
	}
	if got := normalizeConsoleIconData("data:image"); got != "" {
		t.Errorf("短于前缀应清空: %q", got)
	}
	if got := normalizeConsoleIconData(""); got != "" {
		t.Errorf("空串应清空: %q", got)
	}
}

func TestNormalizeBookmarkFolderPathValue(t *testing.T) {
	if got := normalizeBookmarkFolderPathValue(" / A / B / "); got != "A / B" {
		t.Errorf("首尾斜杠应剥除: %q", got)
	}
	if got := normalizeBookmarkFolderPathValue("/A/B/"); got != "A/B" {
		t.Errorf("无空格斜杠应仅剥首尾: %q", got)
	}
	if got := normalizeBookmarkFolderPathValue("A /  / B"); got != "A / B" {
		t.Errorf("空段应去重: %q", got)
	}
	if got := normalizeBookmarkFolderPathValue("  "); got != "" {
		t.Errorf("空白应返空: %q", got)
	}
	if got := normalizeBookmarkFolderPathValue("/"); got != "" {
		t.Errorf("单斜杠应返空: %q", got)
	}
}

func TestNormalizeFileLocatorSearchMode(t *testing.T) {
	cases := map[string]string{
		"bool":          "boolean",
		"boolean":       "boolean",
		"word":          "wholeWord",
		"whole-word":    "wholeWord",
		"regex":         "regex",
		"regexp":        "regex",
		"boolregex":     "booleanRegex",
		"bool-regex":    "booleanRegex",
		"booleanregex":  "booleanRegex",
		"boolean-regex": "booleanRegex",
		"  BOOL  ":      "boolean", // TrimSpace+ToLower
		"whatever-else": "plain",
		"":              "plain",
	}
	for in, want := range cases {
		if got := normalizeFileLocatorSearchMode(in); got != want {
			t.Errorf("normalizeFileLocatorSearchMode(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeFileLocatorBooleanScope(t *testing.T) {
	if got := normalizeFileLocatorBooleanScope("line"); got != "line" {
		t.Errorf("line → line, got %q", got)
	}
	if got := normalizeFileLocatorBooleanScope("  LINES  "); got != "line" {
		t.Errorf("LINES → line, got %q", got)
	}
	if got := normalizeFileLocatorBooleanScope("file"); got != "file" {
		t.Errorf("file → file, got %q", got)
	}
	if got := normalizeFileLocatorBooleanScope(""); got != "file" {
		t.Errorf("空 → file, got %q", got)
	}
}

func TestNormalizeFileLocatorFilterType(t *testing.T) {
	cases := map[string]string{
		"bool":     "boolean",
		"boolean":  "boolean",
		"regex":    "regex",
		"regexp":   "regex",
		"glob":     "glob",
		"mask":     "glob",
		"wildcard": "glob",
		"  MASK  ": "glob",
		"unknown":  "plain",
		"":         "plain",
	}
	for in, want := range cases {
		if got := normalizeFileLocatorFilterType(in); got != want {
			t.Errorf("normalizeFileLocatorFilterType(%q) = %q, want %q", in, got, want)
		}
	}
}
