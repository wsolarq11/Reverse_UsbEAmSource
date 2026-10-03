// AUTO-RECONSTRUCTED — DOMAIN: oledblackout windows（Win32 光标/窗口/几何/媒体辅助）
// Source: UsbEAm_Launcher 1.0.3 (Go 1.25.12, PE64), disassembled from
// main.oledBlackout* symbols（windows 域自由函数）。Tier markers:
//
//	[S VA]      — body fully translated from asm (pure geometry / math).
//	[S-sig VA]  — signature proven from asm; body is a faithful zero skeleton.
//	[P]         — signature not yet proven; zero skeleton + blocking reason.
//
// 说明：接收者为 windowsOLEDBlackoutCursorController / oledBlackoutAudioTopologyTracker /
//
//	GSMTC*/IAsyncOperation* 等方法因接收者类型未在 types_oled.go 落地而跳过（见报告）。
//
// 研究用途
package main

import (
	"errors"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// [S 0x1409134c0] Windows 构建恒支持。asm：mov eax,1; ret。
func oledBlackoutSupported() bool {
	return true
}

// [S-sig 0x1409134e0] 构造光标控制器（转发 newWindowsOLEDBlackoutCursorController）。
// asm：调用后 (rax=itab, rbx=data) 返回 oledBlackoutCursorController 接口。
func newOLEDBlackoutCursorController() oledBlackoutCursorController {
	return nil
}

// request 发送光标命令并等待响应（hide/show/close 三选一）。
// [S 0x140913f20] 单 receiver + hide/show/close bool + response chan error，返回 error。
// asm：nil receiver → errors.New("鼠标控制器不可用")；lock.Lock + defer Unlock；
// closed → 返回 closeErr；否则 make(chan error,1) + commands<-命令 + <-响应。
// 注：response 形参在 asm 中未被引用（内部总是 makechan），保留仅为签名对齐。
func (c *windowsOLEDBlackoutCursorController) request(hide, show, close bool, response chan error) error {
	_ = response
	if c == nil {
		return errors.New("鼠标控制器不可用")
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.closed {
		return c.closeErr
	}
	resp := make(chan error, 1)
	c.commands <- oledBlackoutCursorCommand{hide: hide, show: show, close: close, response: resp}
	return <-resp
}

// Hide 隐藏光标。 [S 0x1409140a0] request(true,false,false,nil)，返回 error 透传。
func (c *windowsOLEDBlackoutCursorController) Hide() error {
	return c.request(true, false, false, nil)
}

// Show 显示光标。 [S 0x1409140e0] request(false,true,false,nil)，返回 error 透传。
func (c *windowsOLEDBlackoutCursorController) Show() error {
	return c.request(false, true, false, nil)
}

// restoreOnThread 在 OLED 线程上恢复光标（按次数循环调用恢复闭包）。
// [S-sig 0x140913e60, 192B]：receiver nil 或 restore(+0x18) nil → newobject 错误返回；
// 非 nil 则循环 n 次间接调用 [restore]（闭包 code 于 [restore] 首字）。
// 体待恢复闭包字段与错误语义专项还原。
func (c *windowsOLEDBlackoutCursorController) restoreOnThread(n int) {
	_, _ = c, n
}

// [S-sig 0x140914420] 当前空闲秒数。序言无参数保存（params=[]，直接 newobject + LazyProc.Call 调 GetLastInputInfo）；
// 尾声 imul rax,rax,0xf4240（×1e6 纳秒化）+ xor ebx + xor ecx = 3×int64 返回。
// 三返回值语义（lastInput/tick/current 或 idle/上限/当前）仍待续写，体骨架。
func oledBlackoutCurrentIdleSeconds() (int64, int64, int64) {
	return 0, 0, 0
}

// [S-sig 0x140914580] 为原生窗口应用 chrome（扩展样式/置顶/透明）。
// 序言 morestack 保存 rax+rbx = application.Window 接口(itab+data)，test rax 判 nil 接口。
// 体：application.InvokeSync(func1 闭包)；func1(0x140914620) 捕获 itab+data 两字，
// 经 itab fun[0x130]（rax=data 作 receiver）取窗口句柄，test 判 0 后做
// GWL_EXSTYLE(-16)/GWL_STYLE(-20) 双路 SetWindowLong + SetWindowPos(0x37) + SuppressWindowBorder。
// 返回 void。
func oledBlackoutApplyWindowChrome(window application.Window) {
	_ = window
}

// [S-sig 0x1409147c0] 抑制窗口边框（DwmSetWindowAttribute / SetWindowLong）。单实参，无返回值。
func oledBlackoutSuppressWindowBorder(hwnd uintptr) {
	_ = hwnd
}

// [S-sig 0x140914860] 聚焦窗口（延迟到前台）。序言 morestack 保存 rax+rbx = application.Window
// 接口(itab+data)，test rax 判 nil；体尾调用 oledBlackoutFocusWindowNow(window)。返回 void。
func oledBlackoutFocusWindow(window application.Window) {
	_ = window
}

// [S-sig 0x1409148c0] 立即聚焦窗口。序言 morestack 保存 rax+rbx = application.Window 接口(itab+data)，
// test rax 判 nil；体经 itab fun[0x130]（rax=data 作 receiver）取窗口句柄后聚焦。返回 void。
func oledBlackoutFocusWindowNow(window application.Window) {
	_ = window
}

// [S-sig 0x140914aa0] 光标物理坐标。asm：尾调用 w32.GetCursorPos（返回 (x,y,ok)）。
func oledBlackoutCurrentCursorPhysicalPoint() (int, int, bool) {
	return 0, 0, false
}

// [S-sig 0x140914ac0] 当前按下的键盘键集合。asm：makemap_small + KeyboardVirtualKeys 遍历。
func oledBlackoutPressedKeyboardKeys() map[uintptr]struct{} {
	return nil
}

// [S-sig 0x140914ba0] 枚举当前按下的虚拟键码（GetAsyncKeyState 扫描）。返回切片。
func oledBlackoutKeyboardVirtualKeys() []uintptr {
	return nil
}

// [P] 空闲暂停是否激活（上下文判定）。asm：7 实参（匿名上下文结构体按值展开）；
// 首参匿名上下文类型未落地，签名待实证。
func oledBlackoutIdlePauseActiveContext(ctx interface{}) bool {
	_ = ctx
	return false
}

// [P] 观测浏览器媒体连续性（上下文判定）。asm：多实参 + 匿名上下文结构体；签名待实证。
func oledBlackoutObserveBrowserMediaContinuityContext(ctx interface{}) bool {
	_ = ctx
	return false
}

// [S-sig 0x1409152e0] 序言 morestack 保存 rax/rbx/rcx = candidates slice(3)；
// 循环元素 0x40=oledBlackoutWindowProcessCandidate，oledBlackoutWindowContainsPhysicalPoint(hwnd,x,y) 判光标命中；
// 尾声 9 寄存器展开返回 struct(hwnd/processID/3×string)+r11d(bool)。
func oledBlackoutStandaloneBrowserCandidateAtCursor(candidates []oledBlackoutWindowProcessCandidate) (oledBlackoutWindowProcessCandidate, bool) {
	_ = candidates
	return oledBlackoutWindowProcessCandidate{}, false
}

// [S-sig 0x1409157a0] 判断窗口是否包含物理坐标点。asm：3 实参(hwnd,x,y)，返回 bool。
func oledBlackoutWindowContainsPhysicalPoint(hwnd uintptr, x int, y int) bool {
	_, _, _ = hwnd, x, y
	return false
}

// [S-sig 0x1409158c0] 判断候选是否为独立浏览器窗口。
// 序言保存 8 字参数：rax(hwnd) + ebx dword(processID uint32) + rcx/rdi + rsi/r8 + r9/r10（3×string）。
// rcx/rdi 传 normalizeOLEDBlackoutMediaPauseExclusionProcessName 证实为 string；尾声 mov eax,1 / xor eax = bool。
// 3×string 字段顺序（processName/processPath/windowTitle）为语义推断，体骨架。
func oledBlackoutCandidateIsStandaloneBrowserWindow(hwnd uintptr, processID uint32, processName string, processPath string, windowTitle string) bool {
	_, _, _, _, _ = hwnd, processID, processName, processPath, windowTitle
	return false
}

// [P] 遍历活动渲染会话（DWM 会话）。asm：3 实参 + 匿名上下文；签名待实证。
func oledBlackoutWalkActiveRenderSessionsContext(ctx interface{}) bool {
	_ = ctx
	return false
}

// [P] 遍历渲染端点会话。asm：6 实参；签名待实证。
func oledBlackoutWalkRenderEndpointSessionsContext(ctx interface{}) bool {
	_ = ctx
	return false
}

// [P] 记住最近交互的独立浏览器候选。asm：6 实参（含 uint32 与字符串）；签名待实证。
func oledBlackoutRememberInteractedStandaloneBrowserCandidate(ctx interface{}, candidate *oledBlackoutWindowProcessCandidate) {
	_, _ = ctx, candidate
}

// [S-sig 0x140917cc0] 前台窗口是否全屏。asm：单实参(hwnd)，返回 bool。
func oledBlackoutForegroundWindowIsFullscreen(hwnd uintptr) bool {
	_ = hwnd
	return false
}

// [P] 前台播放是否激活（上下文判定）。asm：多实参 + 匿名上下文；签名待实证。
func oledBlackoutForegroundPlaybackActiveContext(ctx interface{}) bool {
	_ = ctx
	return false
}

// [P] 目标显示器是否有播放媒体窗口（上下文判定）。签名待实证。
func oledBlackoutTargetMonitorsHavePlayingMediaWindowContext(ctx interface{}) bool {
	_ = ctx
	return false
}

// [P] 收集目标屏幕窗口候选。asm：7 qword 参数（rax/rbx/rcx=[]oledBlackoutRect 显示器矩形 + rdi/rsi/r8/r9 匿名上下文字段），
// 直接透传 oledBlackoutCollectScreenWindowCandidates；后 4 参类型未落地，签名无法精确表达，阻断升档。
func oledBlackoutCollectTargetScreenWindowCandidates(screens []*application.Screen) []oledBlackoutWindowProcessCandidate {
	_ = screens
	return nil
}

// [P] 收集可见媒体窗口候选。签名待实证。
func oledBlackoutCollectVisibleMediaWindowCandidates(screens []*application.Screen) []oledBlackoutWindowProcessCandidate {
	_ = screens
	return nil
}

// [P] 收集屏幕窗口候选（核心枚举）。签名待实证。
func oledBlackoutCollectScreenWindowCandidates(screens []*application.Screen) []oledBlackoutWindowProcessCandidate {
	_ = screens
	return nil
}

// [S-sig 0x1409191c0] 候选是否包含粗粒度应用会话。
// 序言保存 rax/rbx = candidates slice(ptr,len)，session string 在 rdi/rsi（函数体 memequal 比较）；
// 尾声 mov eax,1 / xor eax = bool。体骨架。
func oledBlackoutCandidatesContainCoarseAppSession(candidates []oledBlackoutWindowProcessCandidate, session string) bool {
	_, _ = candidates, session
	return false
}

// [S-sig 0x140919480] 候选是否包含指定进程路径。
// 序言：rdi/rsi 传 strings.TrimSpace+ToLower（path string），rax/rbx = candidates slice；
// 尾声 memequal + mov eax,1 / xor eax = bool。体骨架。
func oledBlackoutCandidatesContainProcessPath(candidates []oledBlackoutWindowProcessCandidate, path string) bool {
	_, _ = candidates, path
	return false
}

// [S-sig 0x1409195e0] 序言 morestack 保存 rax/rbx/rcx/rdi/rsi = slice(3)+string(2)；
// 体：(rdi,rsi) TrimSpace+ToLower 判空；normalizeOLEDBlackoutMediaPauseExclusionProcessName 判浏览器名；
// 循环元素 0x40=oledBlackoutWindowProcessCandidate 比较 processPath；尾声 inc rbx 计数 mov rax,rbx = int。
func oledBlackoutCountVisiblePlayingCoarseAppWindows(candidates []oledBlackoutWindowProcessCandidate, processPath string) int {
	_, _ = candidates, processPath
	return 0
}

// [S-sig 0x140919c00] 记住浏览器媒体连续性。
// 序言 test rax + 保存 rax/rbx/rcx = candidates slice(3字)；尾声无返回值设置（void）。体骨架。
func oledBlackoutRememberBrowserMediaContinuity(candidates []oledBlackoutWindowProcessCandidate) {
	_ = candidates
}

// [P] 播放媒体候选上下文。签名待实证。
func oledBlackoutPlayingMediaCandidatesContext(ctx interface{}) bool {
	_ = ctx
	return false
}

// [P] 浏览器音频是否匹配可见播放窗口。签名待实证。
func oledBlackoutBrowserAudioMatchesVisiblePlayingWindows(candidates []oledBlackoutWindowProcessCandidate) bool {
	_ = candidates
	return false
}

// [P] 浏览器媒体连续性是否匹配候选。签名待实证。
func oledBlackoutBrowserMediaContinuityMatchesCandidates(candidates []oledBlackoutWindowProcessCandidate) bool {
	_ = candidates
	return false
}

// [S-sig 0x14091b260] 汇编实证：单实参 rax=candidates(*[]oledBlackoutWindowProcessCandidate，nil 直返)，
// EnumWindows 回调 oledBlackoutEnumTopLevelWindows.func1 写入；无返回值。
func oledBlackoutEnumTopLevelWindows(candidates *[]oledBlackoutWindowProcessCandidate) {
	_ = candidates
}

// [S-sig 0x14091b560] 汇编实证：rax=hwnd(uintptr)+rbx=cache(map[uintptr]oledBlackoutRect，透传
// mapaccess2_fast64 的 hmap) 参数；返回 (oledBlackoutRect, bool)（4×int32 寄存器 + esi 找到标志）。
// 旧存根误标 screen *application.Screen 且漏 bool 返回，已订正。
func oledBlackoutVisibleWindowRectForScreenPause(hwnd uintptr, cache map[uintptr]oledBlackoutRect) (oledBlackoutRect, bool) {
	_, _ = hwnd, cache
	return oledBlackoutRect{}, false
}

// [S-sig 0x14091b780] 窗口标题。asm：单实参(hwnd)，返回 string。
func oledBlackoutWindowTitle(hwnd uintptr) string {
	_ = hwnd
	return ""
}

// [S-sig 0x14091b8c0] 序言 morestack 保存 eax/ebx/ecx/edi(dword)/rsi/r8/r9 = rect(4×int32)+slice(3)；
// 体：test r8 判 slice.len==0 则 newobject 写单元素返回；否则 makeslice+遍历 rsi 元素 16 字节=oledBlackoutRect 交集合并；
// 尾声返回 []oledBlackoutRect。
func oledBlackoutWindowTargetVisibleRects(r oledBlackoutRect, targets []oledBlackoutRect) []oledBlackoutRect {
	_, _ = r, targets
	return nil
}

// [S-sig 0x14091bb40] 序言 morestack 保存 eax/ebx/ecx/edi(dword)/rsi/r8/r9 = rect(4×int32)+slice(3)；
// 体：cmp ecx,eax / edi,ebx 判 rect 面积；循环 rsi/r8 元素 16 字节=oledBlackoutRect，
// oledBlackoutRectSubtract 差集累加；尾声 setne al = bool。
func oledBlackoutRectHasVisibleAreaAfterOcclusion(r oledBlackoutRect, occluders []oledBlackoutRect) bool {
	_, _ = r, occluders
	return false
}

// [S-sig 0x14091bdc0] 矩形差集（a 减去 b 的可见区域）。
// 序言 eax/ebx/ecx/edi = a(4×int32)，esi/r8d/r9d/r10d = b(4×int32，cmp+cmovl/cmovg 钳制)；
// 尾声 newobject 写 4×int32 元素 + ebx=1 = 返回 []oledBlackoutRect。体骨架。
func oledBlackoutRectSubtract(a oledBlackoutRect, b oledBlackoutRect) []oledBlackoutRect {
	_, _ = a, b
	return nil
}

// [S-sig 0x14091c020] 窗口是否匹配目标显示器。
// 序言：rax=hwnd 传 oledBlackoutWindowMonitorRect（test sil 判返回 bool），rbx/rcx = rects slice；
// 比较循环后 mov eax,1 / xor eax = bool。体骨架。
func oledBlackoutWindowMatchesTargetMonitors(hwnd uintptr, rects []oledBlackoutRect) bool {
	_, _ = hwnd, rects
	return false
}

// [S-sig 0x14091c100] 窗口所在显示器矩形。asm：单实参(hwnd)，返回 oledBlackoutRect（4 整型）。
func oledBlackoutWindowMonitorRect(hwnd uintptr) oledBlackoutRect {
	_ = hwnd
	return oledBlackoutRect{}
}

// [S-sig 0x14091c240] 屏幕集合对应的显示器矩形集合。
// 序言保存 rax/rbx = screens slice(ptr,len)，test rbx 后 makeslice；尾声写 4×int32 元素并返回 []oledBlackoutRect。
// 体骨架。
func oledBlackoutMonitorRectsForScreens(screens []*application.Screen) []oledBlackoutRect {
	_ = screens
	return nil
}

// [S 0x14091c6c0] 两个显示器矩形是否在 2px 容差内逐边匹配。
// asm：a.left=eax,a.top=ebx,a.right=ecx,a.bottom=edi；b.left=esi,b.top=r8d,b.right=r9d,b.bottom=r10d；
// 逐边 abs(a-b) > 2 即 false，否则 true。
func oledBlackoutMonitorRectsMatch(a oledBlackoutRect, b oledBlackoutRect) bool {
	if diff32(a.left, b.left) > 2 {
		return false
	}
	if diff32(a.top, b.top) > 2 {
		return false
	}
	if diff32(a.right, b.right) > 2 {
		return false
	}
	return diff32(a.bottom, b.bottom) <= 2
}

// [S] |a-b|（int32 环绕语义，按 oledBlackoutMonitorRectsMatch 的 asm sub+cmovl 直译，提取为 helper）。
func diff32(a int32, b int32) int32 {
	d := a - b
	if d < 0 {
		return -d
	}
	return d
}

// [S-sig 0x14091c720] 序言 morestack 保存 eax(dword)/rbx/rcx/rdi/rsi/r8 = pid(uint32)+string(2)+slice(3)；
// 体：test rsi 判 slice.len==0；(rbx,rcx) 传 normalizePath/ToLower 与 normalizeProcessName，
// resolveAudioSessionProcessPath(pid) 兜底；循环 add rcx,0x68 元素大小 0x68=OLEDBlackoutMediaPauseExclusion；
// 尾声 mov eax,1/xor eax = bool。
func oledBlackoutMediaPauseProcessExcluded(pid uint32, processPath string, exclusions []OLEDBlackoutMediaPauseExclusion) bool {
	_, _, _ = pid, processPath, exclusions
	return false
}

// [P] 音频会话是否匹配前台。asm：eax(dword)+rbx/rcx(string) 寄存器参数，另 [rsp+0x58](dword)+[rsp+0x60..68](string)+2 bool 栈参数，
// 即 GSMTC session 结构体按寄存器+栈混合展开（PID + ID string + bool 标志）；该类型未落地，签名无法精确表达，阻断升档。
func oledBlackoutAudioSessionMatchesForeground(sessionID string, foregroundID string) bool {
	_, _ = sessionID, foregroundID
	return false
}

// [P] 音频会话是否匹配候选。asm：rax(qword)/ebx(dword)/rcx/rdi/rsi/r8/r9/r10 寄存器 + [rsp+0x58](dword)+[rsp+0x60..68](string)+2 bool 栈参数，
// 即 GSMTC session 结构体按寄存器+栈混合展开；体含浏览器进程名硬编码列表比较，类型未落地，签名无法精确表达，阻断升档。
func oledBlackoutAudioSessionMatchesCandidate(sessionID string, candidate string) bool {
	_, _ = sessionID, candidate
	return false
}

// [S-sig 0x14091cde0] 媒体会话来源是否包含 PID。
// 序言保存 rax/rbx = source string(2字)，test ecx = pid uint32；convT32(pid)+fmt.Sprintf+stringslite.Index；
// 尾声 setge cl/mov eax,ecx = bool。体骨架。
func oledBlackoutMediaSessionSourceContainsPID(source string, pid uint32) bool {
	_, _ = source, pid
	return false
}

// [S-sig 0x14091cea0] 媒体标题是否匹配窗口标题。
// 序言两次 call oledBlackoutNormalizeMediaComparableText（2×string：rax/rbx, rcx/rdi）；
// 尾声 stringslite.Index + setge dl = bool。体骨架。
func oledBlackoutMediaTitleMatchesWindowTitle(mediaTitle string, windowTitle string) bool {
	_, _ = mediaTitle, windowTitle
	return false
}

// [S-sig 0x14091cf80] 规范化媒体可比文本。asm：TrimSpace → ToLower → Fields → 重拼；返回 string。
func oledBlackoutNormalizeMediaComparableText(s string) string {
	_ = s
	return ""
}

// [P] 按路径统计活动音频会话（上下文判定）。asm：2 实参 + makemap_small；签名待实证。
func oledBlackoutActiveAudioSessionCountsByPathContext(ctx interface{}) bool {
	_ = ctx
	return false
}

// [P] 前台媒体会话暂停判定（上下文）。签名待实证。
func oledBlackoutForegroundMediaSessionPauseDecisionContext(ctx interface{}) bool {
	_ = ctx
	return false
}

// [P] 前台音频活动（上下文）。签名待实证。
func oledBlackoutForegroundAudioActiveContext(ctx interface{}) bool {
	_ = ctx
	return false
}

// [P] 任一候选是否有活动音频（上下文）。签名待实证。
func oledBlackoutAnyCandidateHasActiveAudioContext(ctx interface{}) bool {
	_ = ctx
	return false
}

// [P] 候选集合的媒体会话暂停判定（上下文）。签名待实证。
func oledBlackoutMediaSessionPauseDecisionForCandidatesContext(ctx interface{}) bool {
	_ = ctx
	return false
}

// [P] 媒体会话来源是否匹配候选（上下文）。签名待实证。
func oledBlackoutMediaSessionSourceMatchesCandidateContext(ctx interface{}) bool {
	_ = ctx
	return false
}

// [S-sig 0x14091ef40] 进程的 ApplicationUserModelID。
// 序言 test eax = pid uint32（0 值检查），OpenProcess(pid, 0x1000=QUERY_LIMITED_INFORMATION) 取句柄；
// 返回 rax/rbx = string，error 成功时 xmm15 清零 = nil。体骨架。
func oledBlackoutProcessApplicationUserModelID(pid uint32) (string, error) {
	_ = pid
	return "", nil
}

// [P] WinRT 上下文包装。签名待实证。
func oledBlackoutWithWinRTContext(ctx interface{}) bool {
	_ = ctx
	return false
}

// [P] 请求 MediaSessionManager（WinRT）。签名待实证。
func oledBlackoutRequestMediaSessionManagerContext(ctx interface{}) bool {
	_ = ctx
	return false
}

// [P] RoGetActivationFactory 封装。签名待实证。
func oledBlackoutRoGetActivationFactory(ctx interface{}) bool {
	_ = ctx
	return false
}

// [S-sig 0x14091fc80] 序言 morestack 保存 rax/rbx = string；UTF16PtrFromString 后 WindowsCreateString；
// 成功 xor ebx/ecx 返回 (rax=HSTRING, nil)；HRESULT<0 走 fmt.Errorf；test rbx 判 UTF16PtrFromString error。
func oledBlackoutNewHString(s string) (uintptr, error) {
	_ = s
	return 0, nil
}

// [S-sig 0x14091fde0] 释放 HSTRING。asm：单实参，无返回值（内部 WindowsDeleteString）。
func oledBlackoutDeleteHString(h uintptr) {
	_ = h
}

// [S-sig 0x14091fe40] 序言 morestack 保存 rax = HSTRING；test rax 判 0 返回 ("")；
// WindowsGetStringRawBuffer 取 UTF16 指针+长度，mul rcx×2 越界检查后 UTF16ToString；
// 只设 rax/rbx = string，无 error 返回。
func oledBlackoutHStringToString(h uintptr) string {
	_ = h
	return ""
}

// [P] COM QueryInterface 封装。签名待实证。
func oledBlackoutQueryInterface(ctx interface{}) bool {
	_ = ctx
	return false
}

// GetInt32 读取 WinRT IReference<int32>.get_Value。
// [S 汇编 0x140921060, 160B]：值槽清零 → SyscallN(vtbl[6]@0x30, this, &value) →
// HRESULT 有符号 <0 → (0,false)；否则读回 value 并置 bool=true 返回 (int32, bool)。
// 实证：receiver 首字段为 vtable 指针，this 以接口指针传入；bool 语义 = HRESULT 非负。
func (r *oledBlackoutIReference) GetInt32() (int32, bool) {
	var value int32
	getValue := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(r.vtbl)) + 0x30))
	hresult, _, _ := syscall.SyscallN(getValue, uintptr(unsafe.Pointer(r)), uintptr(unsafe.Pointer(&value)))
	if int32(hresult) < 0 {
		return 0, false
	}
	return value, true
}

// Cancel 取消 WinRT IAsyncOperation（经 IAsyncInfo.Cancel）。
// [S-sig 0x140921960, 192B]：oledBlackoutQueryInterface 取 IAsyncInfo（nil/失败则返回），
// defer releaseComObject；SyscallN(vtbl[9]@0x48 Cancel, this)。体待 QueryInterface 签名专项还原。
func (r *oledBlackoutIAsyncOperation) Cancel() {
	_ = r
}

// newWindowsOLEDBlackoutCursorController 构造 Windows 光标控制器（chan × 3 + 起协程 + 等 ready）。
// [S 汇编 0x140913520, 320B]：makechan(commands/ready/done) → newobject 填充 call(+0x18) →
// newproc(goroutine) → chanrecv1(ready) → 返回 controller。
func newWindowsOLEDBlackoutCursorController(call func(bool) (int32, error)) *windowsOLEDBlackoutCursorController {
	c := &windowsOLEDBlackoutCursorController{
		commands: make(chan oledBlackoutCursorCommand),
		ready:    make(chan struct{}),
		done:     make(chan struct{}),
		call:     call,
	}
	go c.run()
	<-c.ready
	return c
}

// run 光标控制器命令循环（LockOSThread + close(ready) + 分发 hide/restore/close）。
// [S 汇编 0x1409136c0, 512B]：LockOSThread → defer UnlockOSThread → defer close(done) →
// close(ready) → 循环 <-commands 分发 hideOnThread/restoreOnThread/close。
// 体待命令分发专项还原。
func (c *windowsOLEDBlackoutCursorController) run() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer close(c.done)
	close(c.ready)
	for range c.commands {
	}
}

// handleOverlayDismissOnAnyKey 处理任意键关闭遮罩（持锁，守卫判定）。
// [S-sig 0x14090b0c0, 384B]：lock(+0x0) → field(+0x88)/field(+0x3a) 判定 →
// inputDismissGuardActiveLocked → dismissOverlayForKeyboardInputLocked。体待键盘域专项还原。
func (s *oledBlackoutService) handleOverlayDismissOnAnyKey(a, b interface{}) bool {
	_, _, _ = s, a, b
	return false
}

// findMatchingVisibleProfileIDLocked 查找匹配可见屏幕的配置 ID（持锁，遍历 profiles）。
// [S-sig 0x14090d840, 384B]：profiles(+0xf8) 空→""；collectScreensLocked →
// oledBlackoutProfileMatchesVisibleScreens → TrimSpace(profile.id)。体待配置域专项还原。
func (s *oledBlackoutService) findMatchingVisibleProfileIDLocked() string {
	_ = s
	return ""
}
