package main

import "testing"

// 黄金用例期望值来源：windowManagementWrappedCursorPoint（0x1409ed5c0）的 asm 显式路径
// + 已落地 windowManagementWrapTargetX/Y 的 best∓3 逻辑（batch255/256 实证），非凭空造值。

func TestWindowManagementWrappedCursorPoint(t *testing.T) {
	single := []windowManagementRECT{{Left: 0, Top: 0, Right: 1920, Bottom: 1080}}
	dual := []windowManagementRECT{
		{Left: 0, Top: 0, Right: 1920, Bottom: 1080},
		{Left: 1920, Top: 0, Right: 3840, Bottom: 1080},
	}
	cases := []struct {
		name         string
		x, y         int32
		monitors     []windowManagementRECT
		wrapX, wrapY bool
		guardPx      int
		wantX, wantY int32
		wantOK       bool
	}{
		// 中间点：不命中 corner guard、不触发任何 wrap → 原值。
		{"interior", 960, 540, single, true, true, 100, 960, 540, false},
		// 左上角 (0,0)：corner guard 命中 → (x,y,false)，wrap 不执行。
		{"corner-guard", 0, 0, single, true, true, 100, 0, 0, false},
		// 左边缘 (0,540)：guard 不命中；x<=left，探测 (-1,540) 不在 monitor → WrapTargetX 最右 = 1919-3。
		{"left-edge-wrap", 0, 540, single, true, true, 100, 1916, 540, true},
		// 右边缘 (1919,540)：x>=right-1，探测 (1920,540) 不在 monitor → WrapTargetX 最左 = 0+3。
		{"right-edge-wrap", 1919, 540, single, true, true, 100, 3, 540, true},
		// 上边缘 (960,0)：y<=top，探测 (960,-1) 不在 monitor → WrapTargetY 最下 = 1079-3。
		{"top-edge-wrap", 960, 0, single, true, true, 100, 960, 1076, true},
		// 下边缘 (960,1079)：y>=bottom-1，探测 (960,1080) 不在 monitor → WrapTargetY 最上 = 0+3。
		{"bottom-edge-wrap", 960, 1079, single, true, true, 100, 960, 3, true},
		// 双屏右边缘 (1919,540)：探测 (1920,540) 在第二屏 → 不 wrap，保持原值。
		{"dual-adjacent-no-wrap", 1919, 540, dual, true, true, 100, 1919, 540, false},
		// 双屏左边缘 (0,540)：探测 (-1,540) 不在任何屏 → wrap 到最右 3840-1-3。
		{"dual-left-edge-wrap", 0, 540, dual, true, true, 100, 3836, 540, true},
		// wrapX/wrapY 均关闭：任何点均原值返回。
		{"both-disabled", 0, 0, single, false, false, 100, 0, 0, false},
		// 空 monitors：扫描未命中 → 原值 false。
		{"empty-monitors", 10, 20, nil, true, true, 100, 10, 20, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gx, gy, ok := windowManagementWrappedCursorPoint(c.x, c.y, c.monitors, c.wrapX, c.wrapY, c.guardPx)
			if gx != c.wantX || gy != c.wantY || ok != c.wantOK {
				t.Fatalf("got (%d,%d,%v), want (%d,%d,%v)", gx, gy, ok, c.wantX, c.wantY, c.wantOK)
			}
		})
	}
}

// 黄金用例期望值来源：targetFromWindowProcessPick（0x1409eea40）的 asm 显式路径：
// Path/ProcessName/DisplayName 均 TrimSpace；ProcessName 空时回退 filepath.Base(Path)；
// DisplayName 空时先回退 TrimSpace(title) 再回退 ProcessName；Title 保留原始值。
func TestTargetFromWindowProcessPick(t *testing.T) {
	cases := []struct {
		name                     string
		path                     string
		pid                      uint32
		processName, displayName string
		title                    string
		hwnd                     uintptr
		iconData                 string
		wantPath, wantName       string
		wantTitle, wantDisplay   string
		wantHWND                 uintptr
		wantPID                  uint32
		wantIcon                 string
	}{
		{
			name: "all-populated", path: " C:\\app\\x.exe ", pid: 42,
			processName: "  app  ", displayName: "", title: "  Title  ",
			hwnd: 0x1234, iconData: "icon",
			wantPath: "C:\\app\\x.exe", wantName: "app",
			wantTitle: "  Title  ", wantDisplay: "Title",
			wantHWND: 0x1234, wantPID: 42, wantIcon: "icon",
		},
		{
			name: "name-empty-fallback-base", path: "  C:\\dir\\app.exe  ", pid: 1,
			processName: "   ", displayName: "", title: "",
			hwnd: 0, iconData: "",
			wantPath: "C:\\dir\\app.exe", wantName: "app.exe",
			wantTitle: "", wantDisplay: "app.exe",
			wantHWND: 0, wantPID: 1, wantIcon: "",
		},
		{
			name: "display-fallback-to-name", path: "  C:\\dir\\app.exe  ", pid: 7,
			processName: "  app  ", displayName: "   ", title: "   ",
			hwnd: 0, iconData: "",
			wantPath: "C:\\dir\\app.exe", wantName: "app",
			wantTitle: "   ", wantDisplay: "app",
			wantHWND: 0, wantPID: 7, wantIcon: "",
		},
		{
			name: "all-empty", path: "", pid: 0,
			processName: "", displayName: "", title: "",
			hwnd: 0, iconData: "",
			wantPath: "", wantName: "",
			wantTitle: "", wantDisplay: "",
			wantHWND: 0, wantPID: 0, wantIcon: "",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := targetFromWindowProcessPick(c.path, c.pid, c.processName, c.displayName, c.title, c.hwnd, c.iconData)
			if got.HWND != c.wantHWND || got.ProcessID != c.wantPID ||
				got.ProcessName != c.wantName || got.Path != c.wantPath ||
				got.Title != c.wantTitle || got.DisplayName != c.wantDisplay ||
				got.IconData != c.wantIcon || got.IconRef != "" || got.IconURL != "" {
				t.Fatalf("got %+v", got)
			}
		})
	}
}

// startCursorWrap/stopCursorWrap 的 channel 生命周期黄金用例。
// moduleEnabled 默认 false → cursorWrapLoop 首次 tick（或 stop 关闭）即 return，不触达
// GetSystemMetrics/MaybeWrapCursor 系统调用，测试确定性且安全。
func TestWindowManagementCursorWrapLifecycle(t *testing.T) {
	s := &windowManagementService{}
	s.startCursorWrap(true, true)

	if !s.cursorActive {
		t.Fatal("startCursorWrap: cursorActive = false, want true")
	}
	if s.cursorStop == nil || s.cursorDone == nil {
		t.Fatal("startCursorWrap: cursorStop/cursorDone 未创建")
	}

	// 二次 start 应被 cursorActive 短路（不覆盖既有 channel）。
	prevStop := s.cursorStop
	s.startCursorWrap(true, true)
	if s.cursorStop != prevStop {
		t.Fatal("startCursorWrap: 重复启动覆盖了 cursorStop")
	}

	s.stopCursorWrap()

	if s.cursorActive {
		t.Fatal("stopCursorWrap: cursorActive = true, want false")
	}
	if s.cursorStop != nil || s.cursorDone != nil {
		t.Fatal("stopCursorWrap: cursorStop/cursorDone 未清空")
	}
}

// stopCursorWrap 未启动时幂等：cursorStop/cursorDone 为 nil，跳过 close/recv，不 panic。
func TestWindowManagementStopCursorWrapIdempotent(t *testing.T) {
	s := &windowManagementService{}
	s.stopCursorWrap()
	if s.cursorActive || s.cursorStop != nil || s.cursorDone != nil {
		t.Fatal("stopCursorWrap on fresh service 应为 no-op")
	}
}
