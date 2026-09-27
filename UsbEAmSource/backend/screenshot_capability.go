// AUTO-RECONSTRUCTED — DOMAIN: screenshot route capability (HTTP 鉴权中间件)
// Source: UsbEAm_Launcher 1.0.3 (Go 1.25.12, PE64), disassembled from
// main.screenshotRouteCapability* / main.*ScreenshotCapability* symbols. Tier markers:
//
//	[S VA]      — body fully translated from asm (pure geometry / math).
//	[S-sig VA]  — signature proven from asm; body is a faithful zero skeleton
//	              (net/http 响应装配未还原).
//
// 研究用途
package main

import "net/http"

// [S-sig 0x14096cb40] 构造路由能力。形参经 asm 实证为两个 string（token/owner），
// 返回 40B 值结构 screenshotRouteCapability（{Token, Owner string, Generation int64}），
// Generation 由包级计数器填充。体未还原，返回零值。
func newScreenshotRouteCapability(token, owner string) screenshotRouteCapability {
	return screenshotRouteCapability{}
}

// [S-sig 0x14096cd60] 值接收者方法；判定 Token 是否为空。返回 bool。
func (c screenshotRouteCapability) Valid() bool { return false }

// [S-sig 0x14096ce20] 值接收者方法；对 *http.Request 校验能力（先调 Valid，再比对
// 请求头/查询中的 token）。r 经 r8 传递。返回 bool。
func (c screenshotRouteCapability) MatchesRequest(r *http.Request) bool { return false }

// [S-sig 0x14096d000] 值接收者方法；从能力派生鉴权 URL。返回 string。
func (c screenshotRouteCapability) URL() string { return "" }

// [S-sig 0x14096d3c0] 比对能力值。形参为两个 string（token/value，4 寄存器），
// 走 runtime.memequal 判等。返回 bool。
func screenshotCapabilityValueMatches(token, value string) bool { return false }

// [S-sig 0x14096d4c0] 向 ResponseWriter 写入能力响应头（固定 13B 头名）。rax 为 w，
// nil 直返。体未还原。
func applyScreenshotCapabilityResponseHeaders(w http.ResponseWriter) {}

// [S-sig 0x14096d760] 拒绝能力请求：写入拒绝状态后回填响应头。rax/rbx 为 w 接口两字。
func rejectScreenshotCapabilityRequest(w http.ResponseWriter) {}

// [S-sig 0x14096d7e0] 序言 spill rsi/rbx/rax/rdi = w 接口两字 + method string 两字；rcx 判空即 r 指针；
// cmp rsi, r.Method.len 后 memequal(r.Method.ptr, method.ptr, len)，错误路径 xor eax 返回 false。返回 bool。
func requireScreenshotRequestMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	return false
}

// [S-sig 0x14096d980] 序言 spill rax/rbx = w 接口两字；rcx 判空即 r 指针，读 [rcx+0x40] 为 r.Body(io.ReadCloser)；
// ifaceeq 判接口相等，尾声 movzx eax, byte[rsp+0x2e] 返回局部 bool。返回 bool。
func requireEmptyScreenshotRequestBody(w http.ResponseWriter, r *http.Request) bool {
	return false
}

// [S-sig 0x14096dd20] 序言 spill rax/rbx = w 接口两字；rcx 判空即 r 指针，读 [rcx+0x38] 为 r.Header 后
// MIMEHeader.Get 取 Content-Type；错误路径 xor eax 返回 false，成功 mov eax,1 返回 true。返回 bool。
func requireScreenshotJSONContentType(w http.ResponseWriter, r *http.Request) bool {
	return false
}
