// AUTO-RECONSTRUCTED — DOMAIN: screenshot window/area capture & window-selection overlay
// Source: UsbEAm_Launcher 1.0.3 (Go 1.25.12, PE64), disassembled from
// main.captureScreenshot* / main.screenshot* / main.windowProcessPick* symbols.
// Tier markers:
//
//	[S VA]      — body fully translated from asm (pure geometry / math).
//	[S-sig VA]  — signature proven from asm; body is a faithful zero skeleton
//	              (DXGI/GDI/窗口选择 UI 副作用未还原).
//
// 研究用途
package main

import (
	"image"
	"time"
)

// [P] 签名待实证：序言把 service/两个 bool/rect 打包进 [rsp+0x20] 栈结构经 rdi 传 captureScreenshotWithLauncherVisibility；
// 返回 rdi 单字与 []byte 三字不符，参数/返回均存歧义。保留 [P]。
func captureScreenshotAreaPNGForServiceWithControlsAndLauncherVisibility(service *BootstrapService, rect image.Rectangle, withControls, hideLauncher bool) []byte {
	return nil
}

// [S-sig 0x1409b8d80] 由 PNG/源标签/时间/格式最终化截屏结果。
// 汇编实证：morestack 保存 rax..r10 八寄存器；rax=service(recv)，rbx/rcx/rdi=png(3 字，
// 直接透传 copyScreenshotPNGToClipboard)，rsi/r8=source(2 字)，r9/r10=captureTime.wall/ext，
// 栈参数 = captureTime.loc(1)+format(2) = 3 字；返回 ScreenshotCaptureResult(16 字，
// buildScreenshotResultFromPNG 结果 duffcopy 回填)。旧存根 result 大结构 16 字全错，已订正。
func finalizeScreenshotCaptureResult(service *BootstrapService, png []byte, source string, captureTime time.Time, format string) ScreenshotCaptureResult {
	_, _, _, _, _ = service, png, source, captureTime, format
	return ScreenshotCaptureResult{}
}

// [S-sig 0x1409b94c0] 截屏期间隐藏启动器窗口、执行 cb 后恢复。汇编实证：rax=service、rbx=delay(time.Duration，
// test 后 call time.Sleep)、cl=hideLauncher(bool)、rdi=cb(func() ([]byte,bool,error)，test nil 后 call)；morestack
// 保护 4 寄存器。内部 beginLauncherScreenshotCapture 保存状态 → sleep(delay) → call cb → endLauncherScreenshotCapture
// 恢复；返回 cb 的 6 字透传 ([]byte,bool,error)，错误路径 (nil,false,err)。
func captureScreenshotWithLauncherVisibility(service *BootstrapService, delay time.Duration, hideLauncher bool, cb func() ([]byte, bool, error)) ([]byte, bool, error) {
	return nil, false, nil
}

// [S-sig 0x1409b9980] 取当前光标所在显示器矩形。无参。返回 image.Rectangle。
func currentMonitorRect() image.Rectangle { return image.Rectangle{} }

// [S-sig 0x1409b9a20] 取指定窗口所在显示器矩形。返回 image.Rectangle。
func currentMonitorRectByHandle(hwnd uintptr) image.Rectangle { return image.Rectangle{} }

// [S-sig 0x1409b9bc0] 由屏幕坐标取显示器矩形。返回 image.Rectangle。
func monitorFromScreenPoint(x, y int) image.Rectangle { return image.Rectangle{} }

// [S-sig 0x1409b9ce0] 由坐标枚举求显示器矩形。返回 image.Rectangle。
func currentMonitorRectByEnumeration(x, y int) image.Rectangle { return image.Rectangle{} }

// [S-sig 0x1409b9dc0] 枚举全部显示器矩形。返回 []image.Rectangle。
func enumerateMonitorRects() []image.Rectangle { return nil }

// [S-sig 0x1409ba200] 序言 spill rdi/rcx/rbx/rax = rect 四字(image.Rectangle)，调 qrCodeVirtualScreenBounds 后
// cmovg 取 min/max，再 image.Rectangle.Intersect 与工作区求交。返回 image.Rectangle。
func normalizeCurrentScreenCaptureRect(rect image.Rectangle) image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x1409ba2e0] 序言 spill rax(x)/rbx(y)/rcx(rects.ptr)/rdi(rects.len)，遍历时调 squaredDistanceFromPointToRectangle；
// 尾声返回 rax/rbx/rcx/rdi 四字(命中 rect) + esi=1(找到)/esi=0(未找到)。返回 (image.Rectangle, bool)。
func chooseMonitorRectForPoint(x, y int, rects []image.Rectangle) (image.Rectangle, bool) {
	return image.Rectangle{}, false
}

// [S 0x1409ba480] 点到半开矩形 [x0,x1)×[y0,y1) 的平方距离：点在区间内分量取 0，
// 越界取越界量的平方。以下为 asm 直译。
func squaredDistanceFromPointToRectangle(x, y, x0, y0, x1, y1 int) int {
	var dx int
	if x0 < x1 {
		if x < x0 {
			dx = x0 - x
		} else if x > x1-1 {
			dx = x - x1 + 1
		}
	}
	var dy int
	if y0 < y1 {
		if y < y0 {
			dy = y0 - y
		} else if y > y1-1 {
			dy = y - y1 + 1
		}
	}
	return dx*dx + dy*dy
}

// [P] 签名待实证：带选项的窗口选择截屏（序言含 service 指针 + 两个 bool + excludedPID）。
// 返回（PNG + 选中进程 + error）未逐寄存器实证。
func captureScreenshotWindowSelectionWithOptions(service *BootstrapService, captureControls, processPickOnly bool, excludedPID uint32) ([]byte, WindowProcessPickResult, error) {
	return nil, WindowProcessPickResult{}, nil
}

// [S-sig 0x1409ba720] 序言仅 spill eax(uint32=excludedPID)；写 newSession 栈 [rsp+4]=excludedPID、[rsp+8]=0(service 空)，
// 调 run 后 test rsi(error)/test dil(cancelled) 分支，成功路径 add rsi,0x90 从 session.selectedProcess duffcopy 构造结果。
// 返回 WindowProcessPickResult（14 字经栈）。
func pickWindowProcessFromScreenshotSelection(excludedPID uint32) WindowProcessPickResult {
	return WindowProcessPickResult{}
}

// [S-sig 0x1409bab20] 带进程名的前台窗口截屏。
// 汇编实证：rax/rbx=processName(string 2 字)，序言后 xor eax 调 captureQRCodeVirtualScreenSnapshotWithCursor
// (无参，返回 6 字快照)；返回 ScreenshotCaptureResult(16 字)。
func captureScreenshotForegroundWindowWithProcessNameEx(processName string) ScreenshotCaptureResult {
	_ = processName
	return ScreenshotCaptureResult{}
}

// [S-sig 0x1409bae00] 前台窗口进程名（无参）。返回 string。
func screenshotForegroundProcessName() string { return "" }

// [S-sig 0x1409bae40] 指定窗口的进程名。返回 string。
func screenshotProcessNameForWindow(hwnd uintptr) string { return "" }

// [S-sig 0x1409bae80] 前台可截屏窗口句柄（无参）。返回 uintptr。
func screenshotForegroundCapturableWindow() uintptr { return 0 }

// [S-sig 0x1409baf60] 由基准句柄求前台可截屏窗口。返回 uintptr。
func screenshotForegroundCapturableWindowFromBase(hwnd uintptr) uintptr { return 0 }

// [S-sig 0x1409bb040] 注册窗口选择窗口类（幂等，无参无返回值）。
func ensureScreenshotWindowSelectionClass() {}

// [P] 签名待实证：序言 spill rax/rbx/rcx/rdi/rsi/r8 共 6 个参数寄存器(service+snapshot+选项)，非零参；
// 具体参数类型无法唯一确定。保留 [P]。
func newScreenshotWindowSelectionSession() *screenshotWindowSelectionSession { return nil }

// [S-sig 0x1409bb300] 序言仅 spill rax(接收者)；LockOSThread 后 createWindow/showSelectionWindow 循环 LazyProc.Call 泵消息，
// 尾声返回 rax/rbx/rcx=session.resultPNG([]byte 三字) + edi=[rsp+0x27](bool,取消/超时标志) + rsi/r8=session.err(error 两字)。
// 选中进程经 session.selectedProcess 字段而非返回值传递。返回 ([]byte, bool, error)。
func (s *screenshotWindowSelectionSession) run() ([]byte, bool, error) {
	return nil, false, nil
}

// [S-sig 0x1409bb9a0] 无参；读 [rax+0x101]/[rax+0x118..0x130]/[rax+0x60]/[rax+0x50] 判状态，
// 末尾调 didTimeOut 后 xor eax,1 取反，或各分支 xor eax 返回 false。返回 bool。
func (s *screenshotWindowSelectionSession) shouldFinalizeAfterOverlayClosed() bool { return false }

// [S-sig 0x1409bba20] 无参；判 pendingCaptureRect 非空后 encodeScreenshotImagePNG，test rdi 判错误；
// 成功路径 xor eax/ebx 返回 nil error，失败路径 mov rax,rdi/mov rbx,rsi 返回 error 两字。返回 error。
func (s *screenshotWindowSelectionSession) finalizePendingCaptureRect() error { return nil }

// [S-sig 0x1409bbae0] 触发窗口截屏已接受回调（once 幂等）。无参无返回值。
func (n *screenshotWindowCaptureAcceptedNotifier) Emit() {}

// [S-sig 0x1409bbb40] 无参；UTF16PtrFromString 构造窗口类名；成功返回 (rax=type,rbx=data) 即 error 接口，
// 失败路径 xor eax/ebx 返回 nil error。返回 error。
func (s *screenshotWindowSelectionSession) createWindow() error { return nil }

// [S-sig 0x1409bbd80] 序言仅 spill rax(接收者)，读 [rax+0x38]/[rax+0x40] 后 LazyProc.Call 销毁窗口；尾声无返回值。void。
func (s *screenshotWindowSelectionSession) close() {}

// [S-sig 0x1409bbf20] 记录选择窗口句柄（hwnd 字段）。
func (s *screenshotWindowSelectionSession) setWindowHandle(hwnd uintptr) {}

// [S-sig 0x1409bc040] 返回选择窗口句柄。
func (s *screenshotWindowSelectionSession) windowHandle() uintptr { return 0 }

// [S-sig 0x1409bc180] 设置命中测试透明（hitTestTransparent 字段）。
func (s *screenshotWindowSelectionSession) setHitTestTransparent(v bool) {}

// [S-sig 0x1409bc2a0] 是否命中测试透明。返回 bool。
func (s *screenshotWindowSelectionSession) isHitTestTransparent() bool { return false }

// [S-sig 0x1409bc3e0] 标记已超时（timeoutTriggered 字段）。
func (s *screenshotWindowSelectionSession) markTimedOut() {}

// [S-sig 0x1409bc500] 是否已超时。返回 bool。
func (s *screenshotWindowSelectionSession) didTimeOut() bool { return false }

// [S-sig 0x1409bc640] 无参；调 windowHandle 判空后 ShowWindow/SetCursor 等 LazyProc.Call，尾声无返回值。void。
func (s *screenshotWindowSelectionSession) showSelectionWindow() {}

// [S-sig 0x1409bc7c0] 设置全局活动选择会话。无返回值。
func setActiveScreenshotWindowSelectionSession(session *screenshotWindowSelectionSession) {}

// [S-sig 0x1409bc900] 返回全局活动选择会话。
func activeScreenshotWindowSelectionSession() *screenshotWindowSelectionSession { return nil }

// [S-sig 0x1409bca20] 窗口选择窗口过程（标准 WndProc 签名）。
func screenshotWindowSelectionProc(hwnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	return 0
}

// [S-sig 0x1409bcb00] 序言 spill rax(接收者)/rbx(hwnd)/rdi(wParam)/rsi(lParam)，cmp ecx 判 WM_* 系列(msg)；
// 尾声 xor eax 返回 0 或 mov eax,LRESULT。返回 uintptr。
func (s *screenshotWindowSelectionSession) handleMessage(hwnd uintptr, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	return 0
}

// [S-sig 0x1409bd580] 序言 spill rax(接收者)/rbx，写 [rax+0x50]=1 置 cancelled，gcWriteBarrier 写选中进程字段；无返回值。void。
func (s *screenshotWindowSelectionSession) beginRightButtonCancel() {}

// [S-sig 0x1409bd740] 序言 spill rbx(临时)/rax(接收者)，写 [rax+0x50]=1 置 cancelled 并清 hover 相关字段；无返回值。void。
func (s *screenshotWindowSelectionSession) cancelAndClose() {}

// [S-sig 0x1409bd900] 序言 spill rax(接收者)/rbx(hwnd)；rbx 作为 BeginPaint/EndPaint 的 hwnd 参数，
// 成功后读 session.snapshot 与 PAINTSTRUCT 求交并 qrCodeBitBlt/绘制边框；所有返回路径 mov rax,[rsp+0x50](恒 0) 返回 uintptr。
func (s *screenshotWindowSelectionSession) paint(hwnd uintptr) uintptr { return 0 }

// [S-sig 0x1409bdda0] 汇编实证：recv+rbx=hwnd(uintptr，透传 w32.InvalidateRect(hwnd,rect,false)) 参数；
// 无返回值。
func (s *screenshotWindowSelectionSession) invalidateWindowChange(hwnd uintptr) {}

// [S-sig 0x1409be100] 汇编实证：recv+rbx=hwnd(uintptr，透传 w32.InvalidateRect)+a/b(image.Rectangle
// 各 4 word，rcx/rdi/rsi/r8 与栈 [rsp+0xf0..0x108]) 参数；无返回值。
func (s *screenshotWindowSelectionSession) invalidateInfoRectChange(hwnd uintptr, a, b image.Rectangle) {}

// [S-sig 0x1409be360] 序言 test rbx 判空(hwnd)，mov rax,rbx 后调 screenshotWindowBounds 求边框，
// 再 image.Rectangle.Intersect 与工作区求交；零值分支 xor eax 清四字。返回 image.Rectangle。
func (s *screenshotWindowSelectionSession) resolveWindowRectForWindow(hwnd uintptr) image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x1409be420] 序言 spill rax(接收者)/rbx(x)/rcx(y)；x+=snapshot[0]、y+=snapshot[8] 平移窗口内坐标，
// windowHandle 取 hwnd 后调 screenshotTopWindowAtPoint(x,y,hwnd) 透传单字返回。返回 uintptr。
func (s *screenshotWindowSelectionSession) resolveWindowAtPoint(x, y int) uintptr { return 0 }

// [P] 签名待实证：解析窗口的控件悬停。形参/返回未逐寄存器实证。
func (s *screenshotWindowSelectionSession) resolveControlHoverForWindow(hwnd uintptr) {}

// [S-sig 0x1409be900] 汇编实证：recv+rbx=hwnd+rcx/rdi=x/y(int)+sil/r8b=ctrlFound/winFound(bool) 参数；
// 返回 bool。逻辑：recv[0x210]==hwnd 且 shouldReuseQRCodeControlHover(rect,point) 且
// cachedControlHoverAllowedForPoint(hwnd,x,y,ctrlFound,winFound)。
func (s *screenshotWindowSelectionSession) shouldReuseControlHoverForWindow(hwnd uintptr, x, y int, ctrlFound, winFound bool) bool {
	_, _, _, _ = hwnd, x, y, ctrlFound
	_ = winFound
	return false
}

// [S-sig 0x1409bea00] 汇编实证：recv+rbx=hwnd+rcx/rdi=x/y(int)+sil/r8b=ctrlFound/winFound(bool) 参数；
// 返回 bool。逻辑：hover rect 非空后 !shouldPreferWindowOverControlAtPoint(hoverRect,windowRect,x,y,ctrlFound,winFound)。
func (s *screenshotWindowSelectionSession) cachedControlHoverAllowedForPoint(hwnd uintptr, x, y int, ctrlFound, winFound bool) bool {
	_, _, _, _ = hwnd, x, y, ctrlFound
	_ = winFound
	return false
}

// [P] 签名待实证：解析点处窗口的截屏矩形。返回 image.Rectangle。形参未逐寄存器实证。
func (s *screenshotWindowSelectionSession) resolveCaptureRectForWindowAtPoint(hwnd uintptr, x, y int) image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x1409bed80] 序言仅 spill rax(接收者)，hover 从 [rax+0x1a0](session.hoverProcess) duffcopy 读取而非形参；
// 按控件尺寸算矩形后 image.Rectangle.Intersect。返回 image.Rectangle。
func (s *screenshotWindowSelectionSession) resolveInfoRectForHover() image.Rectangle {
	return image.Rectangle{}
}

// [S-sig 0x1409befa0] 序言 spill rax(接收者)/rbx(hdc)；test rdx(hdc) 判空，resolveInfoRectForHover 求矩形，
// duffcopy session.hoverProcess 后 windowProcessPickResultHasIdentity 判身份并 LazyProc.Call 绘制；无返回值。void。
func (s *screenshotWindowSelectionSession) drawHoverProcessInfo(hdc uintptr) {}

// [P] 签名待实证：绘制截屏信息文本。形参未逐寄存器实证。
func drawScreenshotInfoText() {}

// [S-sig 0x1409bfae0] 序言 spill eax/ebx 两 dword(int32)；neg eax 作 LOGFONT.lfHeight、写 ebx 作 lfWeight，
// StringToUTF16 构造字体名后 CreateFontIndirect 返回 HFONT。返回 uintptr。
func createScreenshotInfoFont(size, weight int32) uintptr { return 0 }

// [S-sig 0x1409bfc00] 按路径在给定矩形内绘制应用图标。
// 汇编实证：rax/rbx=path(string 2 字) + rdi/rsi/r8/r9=rect(image.Rectangle 4 int，cmp rdi,r8/rsi,r9
// 判非空即直返)；返回 void。旧存根漏 rect 参数，已订正。
func drawAppIconForPath(path string, rect image.Rectangle) {
	_, _ = path, rect
}

// [S-sig 0x1409bfe40] 序言 spill rax/rbx(path string)/rcx(size int)；cmovle 将 size 下限钳到 0x24，
// 依次 loadPrivateExtractedAppIcon/getSystemIconIndexWithAttributes/loadAssociatedAppIcon 求图标，
// 返回 rax 单字(图标句柄 HICON)。返回 uintptr。
func loadBestAppIconForOverlay(path string, size int) uintptr { return 0 }

// [S-sig 0x1409c0120] 序言 spill rax/rbx(path string)/ecx(pid uint32)；TrimSpace/Clean/Base 规范化 path 后
// 调 buildWindowProcessPickResultWithProcessName 并 duffcopy 构造结果。返回 WindowProcessPickResult（14 字经栈）。
func buildWindowProcessPickResult(path string, pid uint32) WindowProcessPickResult {
	return WindowProcessPickResult{}
}

// [S-sig 0x1409c02a0] 序言 spill rax/rbx(path string)/ecx(pid uint32)/rdi/rsi(processName string)；
// resolveAppDisplayName/resolveAppIconDataWithOptions 补全显示名与图标，duffcopy 构造结果。返回 WindowProcessPickResult（14 字经栈）。
func buildWindowProcessPickResultWithProcessName(path string, pid uint32, processName string) WindowProcessPickResult {
	return WindowProcessPickResult{}
}

// [S-sig 0x1409c05c0] 序言 mov [rsp+0x68],rax 保存 hwnd 后调 screenshotWindowTitle + TrimSpace/Clean，
// duffzero/duffcopy 构造结果。返回 WindowProcessPickResult（16 字经栈）。
func withWindowProcessPickHandle(hwnd uintptr) WindowProcessPickResult {
	return WindowProcessPickResult{}
}

// [S-sig 0x1409c0680] 窗口标题。返回 string。
func screenshotWindowTitle(hwnd uintptr) string { return "" }

// [S-sig 0x1409c07c0] 拾取结果是否具备身份信息（进程/标题）。返回 bool。
func windowProcessPickResultHasIdentity(r WindowProcessPickResult) bool { return false }

// [S-sig 0x1409c0820] 拾取结果标题。返回 string。
func windowProcessPickTitle(r WindowProcessPickResult) string { return "" }

// [S-sig 0x1409c0940] 拾取结果副标题。返回 string。
func windowProcessPickSubtitle(r WindowProcessPickResult) string { return "" }

// [S-sig 0x1409c0a00] 序言 spill rax/rbx/rcx(rects []image.Rectangle)+rdi/rsi/r8/r9(clipRect image.Rectangle)；
// makeslice 分配结果，逐矩形 image.Rectangle.Intersect 求交后 makemap/mapaccess2/mapassign 去重，growslice 追加。
// 返回 []image.Rectangle。
func uniqueNonEmptyScreenshotRects(rects []image.Rectangle, clipRect image.Rectangle) []image.Rectangle {
	return nil
}

// [S-sig 0x1409c0da0] 序言 spill rax(接收者)/rbx(hwnd)/rcx(x)/rdi(y)；判 processPickOnly 走 resolveProcessInfoForWindow
// 或 resolveCaptureRectForWindowAtPoint；成功 xor eax/ebx 返回 nil error，失败 newobject 构造 error 返回 (itab,data)。返回 error。
func (s *screenshotWindowSelectionSession) finalizeWindowAtPoint(hwnd uintptr, x, y int) error {
	return nil
}

// [S-sig 0x1409c1200] 序言 duffzero 清结果，test rbx 判空(hwnd)，GetWindowThreadProcessId(hwnd) 取 PID，
// 再截图窗口标题 + TrimSpace 构造结果。返回 WindowProcessPickResult（16 字经栈）。
func (s *screenshotWindowSelectionSession) resolveProcessInfoForWindow(hwnd uintptr) WindowProcessPickResult {
	return WindowProcessPickResult{}
}

// [S-sig 0x1409c1540] test rax 判空(*image.RGBA)，读 [rax+0x20..0x38] 为 RGBA.Rect 四字后 image.NewRGBA(rect)；
// 复制像素逐通道变暗。参数为 *image.RGBA，返回 *image.RGBA。
func buildDarkenedScreenshotPreview(src *image.RGBA) *image.RGBA { return nil }

// [S-sig 0x1409c16a0] 窗口边界。返回 image.Rectangle。
func screenshotWindowBounds(hwnd uintptr) image.Rectangle { return image.Rectangle{} }

// [S-sig 0x1409c18c0] 窗口是否可截屏。返回 bool。
func isScreenshotWindowCapturable(hwnd uintptr) bool { return false }

// [S-sig 0x1409c19c0] 点处顶层窗口句柄。返回 uintptr。
func screenshotTopWindowAtPoint(x, y int) uintptr { return 0 }

// [S-sig 0x1409c1ac0] 点处全部窗口句柄。返回 []uintptr。
func enumerateScreenshotWindowsAtPoint(x, y int) []uintptr { return nil }

// [P] 签名待实证：序言 spill r8/r9/r10/rdi/rcx/rsi/rbx/rax/r11 共 8 个非接收者参数寄存器(候选窗口列表+过滤谓词)，
// 远超存根 (x,y)；参数结构无法唯一确定。保留 [P]。
func screenshotCandidateWindowAtPointWith(x, y int) uintptr { return 0 }
