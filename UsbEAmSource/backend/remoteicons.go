package main

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// remoteIconContentTypeAllowed 判断远程图标的内容类型是否与允许值匹配（忽略大小写）。
// [S] ASM 0x140961fe0: TrimSpace → ParseMediaType，err!=nil→false，否则 EqualFold(mediatype, allowed)。
func remoteIconContentTypeAllowed(contentType, allowedType string) bool {
	mediatype, _, err := mime.ParseMediaType(strings.TrimSpace(contentType))
	if err != nil {
		return false
	}
	return strings.EqualFold(mediatype, allowedType)
}

// resolveRemoteIconCollectionName 解析远程图标集合显示名：从 collections 查 name，
// TrimSpace(集合 Name) 非空则返回之，否则回退返回 name。
// [S 汇编 0x1409661a0, 160B] 实证：mapaccess2_faststr 查 name；未命中→返回 name；
// 命中→TrimSpace(value) 非空→返回 TrimSpace(value)，否则返回 name。
func resolveRemoteIconCollectionName(name string, collections map[string]remoteIconifyCollectionRef) string {
	c, ok := collections[name]
	if !ok {
		return name
	}
	if trimmed := strings.TrimSpace(c.Name); trimmed != "" {
		return trimmed
	}
	return name
}

// isWindowsReservedRemoteIconSegment 判定 Windows 保留设备名段。
// [S 汇编 0x140961d60, 288B]：s=ToLower(TrimSpace(segment))；len==3 且 ∈{aux,con,nul,prn}
// （@0x140961d8b 首字符 >'c' 分叉 cmp word 字节）→ true；len==6 且 =="clock$"
// （@0x140961dd0 dword 0x636f6c63 + word 0x246b）→ true；len==4 且 (s[:3]=="com"||s[:3]=="lpt")
// 且 s[3]∈'1'..'9'（@0x140961e30）→ true；否则 false。
func isWindowsReservedRemoteIconSegment(segment string) bool {
	s := strings.ToLower(strings.TrimSpace(segment))
	switch s {
	case "aux", "con", "nul", "prn", "clock$":
		return true
	}
	if len(s) == 4 && (s[:3] == "com" || s[:3] == "lpt") {
		return s[3] >= '1' && s[3] <= '9'
	}
	return false
}

// isRemoteIconIDSegment 判定远程图标 ID 段合法性。
// [S 汇编 0x140961c80, 224B]：空或 len>maxLen（@0x140961ca0）→ false；首/末字符 '-'
// （@0x140961ca5/af）→ false；isWindowsReservedRemoteIconSegment（@0x140961cc6）→ false；
// 逐字节仅限 a-z/0-9/-（@0x140961ced 减 0x61 界 0x19、减 0x30 界 9、==0x2d），否则 false；
// 全通过 → true。
func isRemoteIconIDSegment(segment string, maxLen int) bool {
	if segment == "" || len(segment) > maxLen {
		return false
	}
	if segment[0] == '-' || segment[len(segment)-1] == '-' {
		return false
	}
	if isWindowsReservedRemoteIconSegment(segment) {
		return false
	}
	for i := 0; i < len(segment); i++ {
		c := segment[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			continue
		}
		return false
	}
	return true
}

// remoteIconURLMatchesTarget 判定响应 URL 与目标 URL 是否匹配。
// [S 汇编 0x1409623c0, 320B]：target/candidate 任一 nil（@0x1409623e3/e8）→ false；
// target.User!=nil（@0x1409623ea 偏移 0x20）→ false；!EqualFold(Scheme)（@0x140962409）→ false；
// !EqualFold(Host)（@0x140962440 偏移 0x28）→ false；EscapedPath 不等（@0x140962480 memequal）→ false；
// RawQuery 不等（@0x1409624a8 偏移 0x60/0x68 memequal）→ false；否则 true。
func remoteIconURLMatchesTarget(target, candidate *url.URL) bool {
	if target == nil || candidate == nil {
		return false
	}
	if target.User != nil {
		return false
	}
	if !strings.EqualFold(target.Scheme, candidate.Scheme) {
		return false
	}
	if !strings.EqualFold(target.Host, candidate.Host) {
		return false
	}
	if target.EscapedPath() != candidate.EscapedPath() {
		return false
	}
	return target.RawQuery == candidate.RawQuery
}

// remoteIconResponseMatchesTarget 判定 HTTP 响应与目标 URL 匹配。
// [S 汇编 0x140962300, 192B]：resp/Request/URL 任一 nil（@0x140962313/320/32a，偏移 0x80/0x10）
// → false；url.Parse(urlStr)（@0x140962337）err!=nil 或 nil（@0x140962340/45）→ false；
// 否则 remoteIconURLMatchesTarget(resp.Request.URL, parsed)（@0x140962368）。
func remoteIconResponseMatchesTarget(resp *http.Response, urlStr string) bool {
	if resp == nil || resp.Request == nil || resp.Request.URL == nil {
		return false
	}
	parsed, err := url.Parse(urlStr)
	if err != nil || parsed == nil {
		return false
	}
	return remoteIconURLMatchesTarget(resp.Request.URL, parsed)
}

// remoteIconPathWithinRoot 判定 path 是否位于 root 之内。
// [S 汇编 0x140965480, 288B]：Clean(root)/Clean(path)（@0x1409654a1/ba）→ Rel(cleanRoot,cleanPath)
// （@0x1409654cf）；err!=nil（@0x1409654d7）→ (false,err)；rel==".."（@0x1409654df word 0x2e2e）
// 或前缀 `..\`（@0x140965510 memequal 3B）→ (false,nil)；IsAbs(rel)（@0x14096552f）→ (false,nil)；
// 否则 (true,nil)。
func remoteIconPathWithinRoot(root, path string) (bool, error) {
	cleanRoot := filepath.Clean(root)
	cleanPath := filepath.Clean(path)
	rel, err := filepath.Rel(cleanRoot, cleanPath)
	if err != nil {
		return false, err
	}
	if rel == ".." || strings.HasPrefix(rel, `..\`) {
		return false, nil
	}
	if filepath.IsAbs(rel) {
		return false, nil
	}
	return true, nil
}

// knownRemoteIconProviders 远程图标供应商白名单（Iconify provider 前缀）。
// 汇编实证：.data 段全局 []string{len=7,cap=7}，元素 lucide/tabler/ph/carbon/
// material-symbols/mingcute/simple-icons（@0x140961b69 mov rdx,[rip+0x1264dd0] 线性 memequal 查找）。
var knownRemoteIconProviders = []string{
	"lucide",
	"tabler",
	"ph",
	"carbon",
	"material-symbols",
	"mingcute",
	"simple-icons",
}

// splitRemoteIconID 拆分远程图标 ID（"provider:icon" 形式）。
// [S 汇编 0x1409619e0, 672B]：空或 TrimSpace!=id（@0x140961a0f/20）→ ("","",false)；
// SplitN(id,":",2)（@0x140961a80 genSplit sep ":"）len!=2 → false；
// TrimSpace(parts[0/1])!=原值（@0x140961aa0/e3）→ false；ToLower 两段（@0x140961b22/40）；
// isRemoteIconIDSegment(provider,64)（@0x140961b60）→ false；线性 memequal 查 provider 白名单
// （@0x140961c01 循环）未命中 → false；isRemoteIconIDSegment(icon,256)（@0x140961bc0）→ false；
// 否则 (provider,icon,true)。
func splitRemoteIconID(id string) (provider, icon string, ok bool) {
	if id == "" || strings.TrimSpace(id) != id {
		return "", "", false
	}
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	if strings.TrimSpace(parts[0]) != parts[0] || strings.TrimSpace(parts[1]) != parts[1] {
		return "", "", false
	}
	provider = strings.ToLower(parts[0])
	icon = strings.ToLower(parts[1])
	if !isRemoteIconIDSegment(provider, 64) {
		return "", "", false
	}
	known := false
	for _, p := range knownRemoteIconProviders {
		if p == provider {
			known = true
			break
		}
	}
	if !known {
		return "", "", false
	}
	if !isRemoteIconIDSegment(icon, 256) {
		return "", "", false
	}
	return provider, icon, true
}

// remoteIconCachePath 计算远程图标缓存相对路径。
// [S 汇编 0x140961880, 352B]：root=TrimSpace(cacheRoot) 空（@0x1409618c0）→ ""；
// splitRemoteIconID(iconID)（@0x1409618e3）!ok → ""；否则
// filepath.Join(root,"remote-icons",provider,icon+".svg")（@0x140961980 Join 4 元素，
// 元素1 常量 "remote-icons" 12B，元素3 concatstring2(icon,".svg")）。
func remoteIconCachePath(cacheRoot, iconID string) string {
	root := strings.TrimSpace(cacheRoot)
	if root == "" {
		return ""
	}
	provider, icon, ok := splitRemoteIconID(iconID)
	if !ok {
		return ""
	}
	return filepath.Join(root, "remote-icons", provider, icon+".svg")
}

// humanizeRemoteIconName 将远程图标 ID 的图标名人性化。
// [S 汇编 0x140965ec0, 736B]：splitRemoteIconID(name)（@0x140965eea）!ok → TrimSpace(name)
// （@0x140966080）；否则 NewReplacer("-"," ","_"," ","/"," ")（@0x140965f3f 六常量）
// Replace(icon) → Fields → 逐词 ToUpper(f[:1])+f[1:]（@0x1409660c3/100）→ Join(" ")
// （@0x140966159）。
func humanizeRemoteIconName(name string) string {
	_, icon, ok := splitRemoteIconID(name)
	if !ok {
		return strings.TrimSpace(name)
	}
	replacer := strings.NewReplacer("-", " ", "_", " ", "/", " ")
	replaced := replacer.Replace(icon)
	fields := strings.Fields(replaced)
	for i, f := range fields {
		fields[i] = strings.ToUpper(f[:1]) + f[1:]
	}
	return strings.Join(fields, " ")
}

// errRemoteIconReaderNil 远程图标读取器为空错误（asm 全局 errorString，24B "远程图标内容为空"）。
var errRemoteIconReaderNil = errors.New("远程图标内容为空")

// errRemoteIconContentTooLarge 远程图标内容超限错误（asm 全局 errorString，36B "远程图标内容超过大小限制"）。
var errRemoteIconContentTooLarge = errors.New("远程图标内容超过大小限制")

// readRemoteIconContent 读取远程图标内容并强制大小上限。
// [S 汇编 0x140961e80, 352B]：r==nil（@0x140961ea0）→ errRemoteIconReaderNil；
// limit<=0（@0x140961eac）→ errRemoteIconContentTooLarge；否则
// io.ReadAll(&io.LimitedReader{R:r,N:limit+1})（@0x140961ec1/305）err 透传；
// int64(len(data))>limit（@0x140961f20）→ errRemoteIconContentTooLarge；否则 (data,nil)。
func readRemoteIconContent(r io.Reader, limit int64) ([]byte, error) {
	if r == nil {
		return nil, errRemoteIconReaderNil
	}
	if limit <= 0 {
		return nil, errRemoteIconContentTooLarge
	}
	data, err := io.ReadAll(&io.LimitedReader{R: r, N: limit + 1})
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, errRemoteIconContentTooLarge
	}
	return data, nil
}

// remoteIconSVGURLReferencesLocal 判定 SVG 内所有 url(...) 引用是否均为本地片段引用。
// [S 汇编 0x140963a00, 384B]：循环 Index(rest,"url(")（@0x140963a51，常量 4B）；未找到→true
// （@0x140963b3b）；IndexByte(after,')')（@0x140963aac）<0→false；ref=Trim(TrimSpace(after[:j]),`"'`)
// （@0x140963ae4/af5）；len<2 或 ref[0]!='#'（@0x140963b00/06）→false；否则继续 after[j+1:]。
func remoteIconSVGURLReferencesLocal(data string) bool {
	rest := data
	for {
		i := strings.Index(rest, "url(")
		if i < 0 {
			return true
		}
		after := rest[i+4:]
		j := strings.IndexByte(after, ')')
		if j < 0 {
			return false
		}
		ref := strings.Trim(strings.TrimSpace(after[:j]), `"'`)
		if len(ref) < 2 || ref[0] != '#' {
			return false
		}
		rest = after[j+1:]
	}
}

// remoteIconFailureTTL 远程图标失败缓存有效期（asm 常量 0x8bb2c97000 ns = 600s = 10 分钟）。
const remoteIconFailureTTL = 10 * time.Minute

// remoteIconFailureMap 远程图标失败缓存（provider → 失败时间戳）。
// 汇编实证：全局 map + 全局 []string（LRU 顺序，容量上限 0x800=2048）+ 全局 sync.Mutex。
var remoteIconFailureMap = make(map[string]time.Time)

// remoteIconFailureList 远程图标失败缓存 LRU 顺序列表（provider 名）。
var remoteIconFailureList []string

// remoteIconFailureMutex 保护失败缓存的互斥锁。
var remoteIconFailureMutex sync.Mutex

// cacheRemoteIconFailure 缓存远程图标失败时间戳（LRU，容量 2048）。
// [S 汇编 0x1409628e0, 1184B]：splitRemoteIconID(id)（@0x140962933）!ok→返回；加锁（@0x14096294b）
// mapaccess2 命中（@0x1409629c1）→ 线性删除 list 旧 provider（@0x140962bad）；len>=2048
// （@0x140962a2f）→ list=list[1:] 且 mapdelete(list[0])（@0x140962a0e）；mapassign
// （@0x140962aa1）value=timestamp；append provider 到 list（@0x140962b40）。
func cacheRemoteIconFailure(id string, timestamp time.Time) {
	provider, _, ok := splitRemoteIconID(id)
	if !ok {
		return
	}
	remoteIconFailureMutex.Lock()
	defer remoteIconFailureMutex.Unlock()
	if _, exists := remoteIconFailureMap[provider]; exists {
		for i, k := range remoteIconFailureList {
			if k == provider {
				remoteIconFailureList = append(remoteIconFailureList[:i], remoteIconFailureList[i+1:]...)
				break
			}
		}
	}
	if len(remoteIconFailureList) >= 2048 {
		oldest := remoteIconFailureList[0]
		remoteIconFailureList = remoteIconFailureList[1:]
		delete(remoteIconFailureMap, oldest)
	}
	remoteIconFailureMap[provider] = timestamp
	remoteIconFailureList = append(remoteIconFailureList, provider)
}

// isRemoteIconFailureCached 判定 provider 是否在失败缓存内且未过期。
// [S 汇编 0x140962500, 992B]：加锁；mapaccess2（@0x1409625d9）未命中→false；timestamp.IsZero
// （@0x140962610）或 now.Sub(timestamp)>=TTL（@0x140962640/64f）→ mapdelete+线性删除 list
// （@0x140962700）→false；否则 true。
func isRemoteIconFailureCached(key string, now time.Time) bool {
	remoteIconFailureMutex.Lock()
	defer remoteIconFailureMutex.Unlock()
	ts, ok := remoteIconFailureMap[key]
	if !ok {
		return false
	}
	if ts.IsZero() || now.Sub(ts) >= remoteIconFailureTTL {
		delete(remoteIconFailureMap, key)
		for i, k := range remoteIconFailureList {
			if k == key {
				remoteIconFailureList = append(remoteIconFailureList[:i], remoteIconFailureList[i+1:]...)
				break
			}
		}
		return false
	}
	return true
}

// deleteRemoteIconFailure 从失败缓存删除 provider。
// [S 汇编 0x140962d80, 480B]：加锁；mapdelete_faststr（@0x140962de6）+ 线性删除 list 元素
// （@0x140962e22/e46）；解锁。
func deleteRemoteIconFailure(key string) {
	remoteIconFailureMutex.Lock()
	defer remoteIconFailureMutex.Unlock()
	delete(remoteIconFailureMap, key)
	for i, k := range remoteIconFailureList {
		if k == key {
			remoteIconFailureList = append(remoteIconFailureList[:i], remoteIconFailureList[i+1:]...)
			break
		}
	}
}

// SearchRemoteIcons 搜索远程图标（BootstrapService 入口）。
// [S-sig 0x14095f6c0, 256B]：搜索远程图标候选并返回结果集。
// 体待远程图标搜索域专项还原。
func (bs *BootstrapService) SearchRemoteIcons(a interface{}) interface{} {
	_ = a
	return nil
}
