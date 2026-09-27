// AUTO-RECONSTRUCTED FUNCTION SKELETONS — DOMAIN: filelocator (runtime/engine)
// 研究用途
//
// 契约来源：
//   - gap 清单 D:\Users\Administrator\TEMP\gap_filelocator.csv（Func,Sym,VA,Len）
//   - 签名证据 D:\Users\Administrator\TEMP\sum_filelocator.txt（params/rets/call）
//   - 类型定义 backend/types_filelocator.go（fileLocatorService/FileLocatorConfig 等）
//
// 档位：
//   - [S-sig 0xVA]：签名经符号表/接口/命名明确推断，体为零值。
//   - [P]：签名不确定（多寄存器参数未逐寄存器实证 / rets=[] 与命名推断冲突），体为零值。
//   - 仅落地 gap 清单中缺失的符号；已存在的装配方法（filelocator.go / bootstrapservice_state_deps.go）
//     与编译器生成函数（deferwrap/gowrap/funcN）不重复落地。
package main

import (
	"context"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
)

// ---- 搜索执行引擎（fileLocatorService 内部方法）----

// runSearch 执行一次搜索主循环（StartSearch 派生 goroutine 的体）。
// [S-sig 0x1407d0260] 汇编实证：recv(rax)+prepared(值,栈,约288B)+ctx(2 word)+generation(uint64) 参数；
// 返回 void（尾声无返回寄存器，非 error）。栈槽 [rsp+0x3b0]=prepared.roots.len（值传结构 @rbp+0x10 的 +0xf8），
// 主循环遍历 prepared.roots 对每根调 waitIfPaused(ctx,generation,time.Now()) 等；体未逐条翻译，保持零值。
func (s *fileLocatorService) runSearch(prepared fileLocatorPreparedSearch, ctx context.Context, generation uint64) {
}

// waitIfPaused 在暂停状态下阻塞等待，返回错误（非 bool）。
// [S 汇编 0x1407d09a0, 768B] 实证：recv(rax)+ctx(2 word)+generation(uint64)+startedAt(time.Time 3 word)
// 共 7 寄存器参数；返回 error（2 寄存器=type/data）。体：s.lock.Lock()+defer Unlock；
// 循环条件 generation==gen && state.Running && state.Paused && ctx.Err()==nil 时 pauseCond.Wait()
// 并 time.Since(startedAt)（duration 转秒存局部，未见下游读取，进度发布待确证）；
// 退出后：ctx.Err()!=nil 或 generation 变 → context.Canceled；state.Running → nil；否则 context.Canceled。
func (s *fileLocatorService) waitIfPaused(ctx context.Context, generation uint64, startedAt time.Time) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	for s.generation == generation && s.state.Running && s.state.Paused && ctx.Err() == nil {
		s.pauseCond.Wait()
		_ = time.Since(startedAt)
	}

	if ctx.Err() != nil {
		return context.Canceled
	}
	if s.generation != generation {
		return context.Canceled
	}
	if s.state.Running {
		return nil
	}
	return context.Canceled
}

// walkRoot 遍历单个搜索根目录（filepath.WalkDir 回调体）。
// [S-sig 0x1407d0d20] 汇编实证：recv(rax)+prepared(fileLocatorPreparedSearch 值传,栈 288B)+
// ctx(2 word:rbx/rcx)+generation(uint64:rdi)+path(string:rsi/r8)+progress(*fileLocatorSearchProgress:r9)+
// onFileDone(func(bool):r10)+startedAt(time.Time 3 word:栈)，共 recv+7 寄存器 + prepared/startedAt 两栈块。
// 体：os.Stat(path)→FileInfo.IsDir() 分支；非目录走 processFile；目录走 filepathlite.Clean(path)+
// filepath.WalkDir 递归（闭包捕获 progress/onFileDone/startedAt），未逐条翻译，保持零值。
func (s *fileLocatorService) walkRoot(prepared fileLocatorPreparedSearch, ctx context.Context, generation uint64, path string, progress *fileLocatorSearchProgress, onFileDone func(bool), startedAt time.Time) error {
	return nil
}

// processFile 处理单个文件（stat/判重/内容匹配/写进度）。
// [S-sig 0x1407d15a0] 汇编实证：recv(rax)+prepared(fileLocatorPreparedSearch 值传,栈 288B)+
// ctx(2:rbx/rcx)+generation(rdi)+path(string:rsi/r8)+info(os.FileInfo 接口:r9/r10)+
// progress(*fileLocatorSearchProgress:r11)+onFileDone(func(bool),栈)+startedAt(time.Time 3 word,栈)，
// 共 recv+8 寄存器；返回 error（2 寄存器，旧桩 (bool,error) 误判——各尾迹仅清 rax/rbx 两字）。
// 体：waitIfPaused→pathFilter.Allows→fileName/contentMatcher→matchFileLocatorContentContext→
// 写 progress 计数器与 results 切片→onFileDone(true)，未逐条翻译，保持零值。
func (s *fileLocatorService) processFile(prepared fileLocatorPreparedSearch, ctx context.Context, generation uint64, path string, info os.FileInfo, progress *fileLocatorSearchProgress, onFileDone func(bool), startedAt time.Time) error {
	return nil
}

// acquireFileLocatorContentScan 获取内容扫描信号量（chan struct{} 单缓冲），可被 ctx 取消。
// [S 汇编 0x1407d23e0, 512B] 实证：recv(rax)+ctx(2 word) 参数；返回 error（2 寄存器=type/data，
// 非单 bool）。s==nil → context.Canceled（.data 0x141bc4520 指向 "context canceled" 16B 字符串，
// 实证非自定义错误）；s.lock.Lock() 后 contentScan（汇编偏移 +0x1e8，结构体定义字段待补）为 nil 时
// make(chan struct{},1)；取 ch 并 Unlock；ctx==nil → ch<-struct{}{} 阻塞发送返回 nil；
// 否则 select { case ch<-struct{}{}: nil; case <-ctx.Done(): ctx.Err() }。
func (s *fileLocatorService) acquireFileLocatorContentScan(ctx context.Context) error {
	if s == nil {
		return context.Canceled
	}
	s.lock.Lock()
	if s.contentScan == nil {
		s.contentScan = make(chan struct{}, 1)
	}
	ch := s.contentScan
	s.lock.Unlock()

	if ctx == nil {
		ch <- struct{}{}
		return nil
	}

	select {
	case ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// releaseFileLocatorContentScan 释放内容扫描信号量。
// [S-sig 0x1407d2580]：params=[rax]（仅 receiver）、rets=[]（void），无参无返回。
func (s *fileLocatorService) releaseFileLocatorContentScan() {
}

// publishProgress 发布进度快照。
// [S-sig 0x1407d2620] 汇编实证：recv(rax)+generation(uint64)+startedAt(time.Time 3 word) 共 5 寄存器；
// 返回 void。体：generation 校验 + Running 校验 + 复制 state 快照 + time.Since(startedAt) 算 elapsed 秒，
// 未逐条翻译，保持零值。
func (s *fileLocatorService) publishProgress(generation uint64, startedAt time.Time) {
}

// finishSearch 收尾搜索（置完成/取消态、清理 generation）。
// [S-sig 0x1407d2a00] 汇编实证：recv(rax)+generation(uint64)+startedAt(time.Time 3 word)+lastError(string 2 word)+cancelled(bool)
// 共 8 寄存器（含 r10b）；返回 void。体：time.Now().Format + time.Since(startedAt) + TrimSpace(lastError) 组装终态
// 并清理 recv.cancel（+0x1e0），未逐条翻译，保持零值。
func (s *fileLocatorService) finishSearch(generation uint64, startedAt time.Time, lastError string, cancelled bool) {
}

// ---- 搜索准备 ----

// prepareFileLocatorSearch 由 FileLocatorConfig 构建 fileLocatorPreparedSearch。
// [S-sig 0x1407d2ea0] 汇编实证：request(值,栈)+无寄存器参数；返回 (fileLocatorPreparedSearch 值,栈 0x120)+error(2 寄存器)。
// 返回 prepared 为值（返回值结构体 @[rsp+0x3f8]，request 字段 duffcopy 自 normalized request），非指针；
// 错误路径返回 nil prepared + error。体：normalizeFileLocatorConfig→splitFileLocatorRoots→
// newFileLocatorStringMatcherInternal(fileName/content)→newFileLocatorPathFilter 组装，未逐条翻译，保持零值。
func prepareFileLocatorSearch(request FileLocatorConfig) (fileLocatorPreparedSearch, error) {
	return fileLocatorPreparedSearch{}, nil
}

// ---- 路径过滤器 ----

// newFileLocatorPathFilter 构造空路径过滤器。
// [S-sig 0x1407d3280]：New 命名 → *fileLocatorPathFilter；rets=[] 判定为工具遗漏。
func newFileLocatorPathFilter() *fileLocatorPathFilter {
	return &fileLocatorPathFilter{}
}

// buildFileLocatorPathFilter 从配置构建路径过滤器（include/exclude 规则）。
// [S-sig 0x1407d3400] 汇编实证：cfg(值,栈)+remark/value/include/typeStr(4 string,8 寄存器) 参数；
// 返回 (*fileLocatorPathFilter, error)（filter 指针 + error 2 寄存器，非单指针）。remark 未在体内使用
// （dead 参数，由 FileLocatorFilter.Remark 命名推断）；体：splitFileLocatorFilterEntries→
// normalizeFileLocatorFilterType→parseFileLocatorFilterRule→newFileLocatorPathMatcher 循环，未逐条翻译。
func buildFileLocatorPathFilter(cfg FileLocatorConfig, remark, value, include, typeStr string) (*fileLocatorPathFilter, error) {
	return nil, nil
}

// parseFileLocatorFilterRule 解析单条过滤规则。
// [S 0x1407d3840]：签名逐寄存器实证——(rule string, filterType string)(include bool, typ string, pattern string)。
// 体：normalize 类型 → TrimSpace rule → 前缀 '!'/'-' 记 exclude(include=false) 并剥前缀、'+' 记 include →
// splitFileLocatorFilterRuleType 命中则覆盖 (type,pattern)，否则 pattern=rule。返回 (include, typ, pattern)。
func parseFileLocatorFilterRule(rule string, filterType string) (include bool, typ string, pattern string) {
	typ = normalizeFileLocatorFilterType(filterType)
	pattern = strings.TrimSpace(rule)
	include = true
	if pattern != "" {
		switch pattern[0] {
		case '!', '-':
			include = false
			pattern = strings.TrimSpace(pattern[1:])
		case '+':
			pattern = strings.TrimSpace(pattern[1:])
		}
	}
	if t, p, ok := splitFileLocatorFilterRuleType(pattern); ok {
		typ, pattern = t, p
	}
	return include, typ, pattern
}

// splitFileLocatorFilterRuleType 从规则串中切分 "X:" 类型前缀。
// [S 0x1407d3a40]：TrimSpace 后 rule[1]!=':' 或长度不足返 ("", rule, false)；首字符 ToLower 后
// b→"boolean"/g→"glob"/p→"plain"/r→"regex"，返回 (type, TrimSpace(rule[2:]), true)。
func splitFileLocatorFilterRuleType(rule string) (typ string, pattern string, ok bool) {
	s := strings.TrimSpace(rule)
	if len(s) < 2 || s[1] != ':' {
		return "", s, false
	}
	switch strings.ToLower(s[:1]) {
	case "b":
		typ = "boolean"
	case "g":
		typ = "glob"
	case "p":
		typ = "plain"
	case "r":
		typ = "regex"
	default:
		return "", s, false
	}
	return typ, strings.TrimSpace(s[2:]), true
}

// newFileLocatorPathMatcher 按过滤器类型构造路径匹配器。
// [S 0x1407d3c00]：normalizeFileLocatorFilterType(filterType)=="glob" 走 newFileLocatorGlobPathMatcher(pattern)，
// 否则 newFileLocatorStringMatcherInternal(pattern, filterType, false, "file", false)。
func newFileLocatorPathMatcher(pattern string, filterType string) *fileLocatorStringMatcher {
	if normalizeFileLocatorFilterType(filterType) == "glob" {
		return newFileLocatorGlobPathMatcher(pattern)
	}
	return newFileLocatorStringMatcherInternal(pattern, filterType, false, "file", false)
}

// newFileLocatorGlobPathMatcher 按 glob 通配构造路径匹配器。
// [S 0x1407d3cc0]：TrimSpace 后为空返 nil → compileFileLocatorPathGlobMatcher 失败返 nil →
// newobject 构造 fileLocatorStringMatcher{query:trimmed, mode:"regex", regex:compiled}。
func newFileLocatorGlobPathMatcher(pattern string) *fileLocatorStringMatcher {
	p := strings.TrimSpace(pattern)
	if p == "" {
		return nil
	}
	compiled, err := compileFileLocatorPathGlobMatcher(p)
	if err != nil {
		return nil
	}
	return &fileLocatorStringMatcher{
		query: p,
		mode:  "regex",
		regex: compiled,
	}
}

// Allows 判断路径是否被过滤器放行。
// [S-sig 0x1407d3da0]：命名 Allows → bool；入参 path string。
func (f *fileLocatorPathFilter) Allows(path string) bool {
	return true
}

// ShouldExclude 判断路径是否应被排除。
// [S-sig 0x1407d3ea0]：命名 Should... → bool；入参 path string。
func (f *fileLocatorPathFilter) ShouldExclude(path string) bool {
	return false
}

// splitFileLocatorRoots 切分多根目录串为路径切片。
// [S-sig 0x1407d3f80]：命名 split...Roots → []string；rets 3 寄存器=slice。
func splitFileLocatorRoots(roots string) []string {
	return nil
}

// splitFileLocatorFilterEntries 切分过滤值串为条目切片。
// [S-sig 0x1407d4260]：params 2 寄存器=value string；尾部 runtime.makeslice 返回 3 寄存器=[]string（cap 复用 len 未显式写出，探测仅 2 寄存器）。
func splitFileLocatorFilterEntries(value string) []string {
	return nil
}

// ---- 字符串匹配器 ----

// newFileLocatorStringMatcherInternal 构造字符串匹配器（query/mode/matchCase/booleanScope/wholeWord）。
// [S-sig 0x1407d4420]：签名逐寄存器实证——序言 spill rax/sil/r10b/r9/r8/rdi/rcx 共 8 寄存器 =
// query(2)+mode(2)+matchCase(1)+booleanScope(2)+wholeWord(1)。体：TrimSpace→ToLower→normalize mode/scope→
// newobject 装配 fileLocatorStringMatcher 后按 mode 分发（regex/plain/boolean），待专项 [S]。
func newFileLocatorStringMatcherInternal(query string, mode string, matchCase bool, booleanScope string, wholeWord bool) *fileLocatorStringMatcher {
	return nil
}

// normalizeFileLocatorMatcherMode 归一化匹配模式串。
// [S-sig 0x1407d4900]：命名 normalize...Mode → string；1 string 入 / 1 string 出。
func normalizeFileLocatorMatcherMode(mode string) string {
	return ""
}

// MatchString 判断字符串是否命中（等价 regexp.MatchString 语义）。
// [S-sig 0x1407d4a60]：命名 MatchString → bool；入参 s string。
func (m *fileLocatorStringMatcher) MatchString(s string) bool {
	return false
}

// Ranges 返回命中区间切片。
// [S-sig 0x1407d4d00]：接口 fileLocatorTermMatcher.Ranges(string) []FileLocatorTextRange 固定。
func (m *fileLocatorStringMatcher) Ranges(s string) []FileLocatorTextRange {
	return nil
}

// MatchContent 对整段内容执行匹配，返回行命中结果。
// [S-sig 0x1407d4ee0] 汇编实证：recv(matcher:rax)+doc(fileLocatorTextDocument 8 word:rbx..r11 值传递) 参数；
// 返回 []FileLocatorLineMatch。体：按 mode 分发 boolean(across/per-line) 走 expr=matcher+0x58，
// 非 boolean 走 matchFileLocatorLineBased(matcher)，未逐条翻译。
func (m *fileLocatorStringMatcher) MatchContent(doc fileLocatorTextDocument) []FileLocatorLineMatch {
	return nil
}

// ---- 内容匹配引擎 ----

// matchFileLocatorContentContext 上下文可取消的内容匹配入口。
// [S-sig 0x1407d53e0] 汇编实证：ctx(context.Context:rax/rbx)+path(string:rcx/rdi)+maxSize(int:rsi,默认 1MB)+
// matcher(*fileLocatorStringMatcher:r8) 参数；返回 ([]FileLocatorLineMatch, error)。
// 体：os.OpenFile(path)→Stat 判尺寸→newFileLocatorTextDocument→MatchContent，未逐条翻译。
func matchFileLocatorContentContext(ctx context.Context, path string, maxSize int, matcher *fileLocatorStringMatcher) ([]FileLocatorLineMatch, error) {
	return nil, nil
}

// matchFileLocatorContentStreamContext 流式内容匹配（边读边匹配）。
// [S-sig 0x1407d5d80] 汇编实证：ctx(context.Context:rax/rbx)+f(*os.File:rcx,nil 检查)+maxSize(int:rdi,jle 检查)+
// matcher(*fileLocatorStringMatcher:rsi) 参数；返回 []FileLocatorLineMatch（nil 时只清 rax/rbx 两字）。
// 体：matcher 非 boolean 或 booleanScope≠"file" 时 makeslice(0x1000)+os.File.ReadAt 流式读入后逐行匹配；
// boolean+file 走 across-file 路径（本函数返回 nil），未逐条翻译。
func matchFileLocatorContentStreamContext(ctx context.Context, f *os.File, maxSize int, matcher *fileLocatorStringMatcher) []FileLocatorLineMatch {
	return nil
}

// fileLocatorStreamingLineMatch 流式模式下对单行执行匹配。
// [S-sig 0x1407d78e0] 汇编实证：line(string:rax/rbx)+matcher(*fileLocatorStringMatcher:rcx) 参数；
// 返回 ([]FileLocatorTextRange, int)（slice 3 word + int:rdi）。体：matcher+0x58 布尔表达式存在时走
// HighlightRanges(line)（int=命中数 max(len,1)），否则走 Ranges(line)，未逐条翻译。
func fileLocatorStreamingLineMatch(line string, matcher *fileLocatorStringMatcher) ([]FileLocatorTextRange, int) {
	return nil, 0
}

// fileLocatorContextError 构造/归一化上下文取消错误。
// [S-sig 0x1407d7a40]：params 2 寄存器=ctx 接口；尾声 xor eax,ebx 置 nil，返回 error 接口(2 寄存器)。
func fileLocatorContextError(ctx context.Context) error {
	return nil
}

// canFastRejectContent 判断内容是否可被快速拒判（缺必要子串）。
// [S-sig 0x1407d7ae0]：命名 can... → bool；入参 content string。
func (m *fileLocatorStringMatcher) canFastRejectContent(content string) bool {
	return false
}

// [S 0x1407d7be0] fileLocatorTextDocument 文本文档（内容匹配的行切分预处理产物）；实测 8 word=64B。
type fileLocatorTextDocument struct {
	text  string
	lines []string
	runes [][]rune
}

// newFileLocatorTextDocument 由原始文本构造文本文档。
// [S-sig 0x1407d7be0] 汇编实证：text(string:rax/rbx) 参数；返回 fileLocatorTextDocument 值（8 word，栈返回），
// 非指针。体：strings.Replace(\r\n→\n)→strings.genSplit(\n)→makeslice([][]rune) 每行 stringtoslicerune，
// 未逐条翻译。
func newFileLocatorTextDocument(text string) fileLocatorTextDocument {
	return fileLocatorTextDocument{}
}

// matchFileLocatorLineBased 逐行匹配（无布尔表达式）。
// [S-sig 0x1407d7dc0] 汇编实证：doc(fileLocatorTextDocument 8 word 值)+m(*fileLocatorStringMatcher:r11) 参数；
// 返回 []FileLocatorLineMatch。体：逐行 Ranges(line) 命中后 buildFileLocatorLineMatch，未逐条翻译。
func matchFileLocatorLineBased(doc fileLocatorTextDocument, m *fileLocatorStringMatcher) []FileLocatorLineMatch {
	return nil
}

// matchFileLocatorBooleanPerLine 布尔表达式逐行匹配。
// [S-sig 0x1407d81e0] 汇编实证：doc(fileLocatorTextDocument 8 word 值)+expr(*fileLocatorBooleanExpression:r11) 参数；
// 返回 []FileLocatorLineMatch。体：逐行 HighlightRanges(line) 命中后 buildFileLocatorLineMatch，未逐条翻译。
func matchFileLocatorBooleanPerLine(doc fileLocatorTextDocument, expr *fileLocatorBooleanExpression) []FileLocatorLineMatch {
	return nil
}

// matchFileLocatorBooleanAcrossFile 布尔表达式跨行匹配。
// [S-sig 0x1407d86c0] 汇编实证：doc(fileLocatorTextDocument 8 word 值)+expr(*fileLocatorBooleanExpression:r11) 参数；
// 返回 []FileLocatorLineMatch。体：跨行聚合邻近区间后 buildFileLocatorLineMatch，未逐条翻译。
func matchFileLocatorBooleanAcrossFile(doc fileLocatorTextDocument, expr *fileLocatorBooleanExpression) []FileLocatorLineMatch {
	return nil
}

// buildFileLocatorLineMatch 构造单条行命中记录。
// [S-sig 0x1407d8ce0] 汇编实证：doc(fileLocatorTextDocument 8 word 值)+index(int:r11) 参数；
// 返回 FileLocatorLineMatch（大结构栈返回）。体：doc.lines[index] 转 rune 后按 ranges 截取 before/after，
// 未逐条翻译。
func buildFileLocatorLineMatch(doc fileLocatorTextDocument, index int) FileLocatorLineMatch {
	return FileLocatorLineMatch{}
}

// compactFileLocatorHitLine 压缩命中行为摘要预览串。
// [S-sig 0x1407d9400]：params 5 寄存器=text string(2)+ranges slice(3)；rets 2 寄存器=string。
func compactFileLocatorHitLine(text string, ranges []FileLocatorTextRange) string {
	return ""
}

// readFileLocatorTextFromHandle 从读取句柄读出文本（带尺寸上限）。
// [S-sig 0x1407da360] 汇编实证：r(io.Reader 接口 2 word:rax/rbx)+matcher(*fileLocatorStringMatcher:rcx)+
// maxSize(int:rdi) 参数；返回 (string, error)。matcher 经 newobject 存入上下文对象 +0x08 字段；
// 体：io.ReadAll→isTextLikeFileLocatorContent→decodeFileLocatorText，超限/非文本走 error 路径，未逐条翻译。
func readFileLocatorTextFromHandle(r io.Reader, matcher *fileLocatorStringMatcher, maxSize int) (string, error) {
	return "", nil
}

// isTextLikeFileLocatorContent 判断内容是否为文本（非二进制）。
// [S-sig 0x1407da540]：命名 is... → bool；入参 data []byte。
func isTextLikeFileLocatorContent(data []byte) bool {
	return false
}

// decodeFileLocatorText 将字节解码为文本串。
// [S-sig 0x1407da6e0]：params 3 寄存器=[]byte、rets 2 寄存器=string。
func decodeFileLocatorText(data []byte) string {
	return ""
}

// decodeFileLocatorUTF16 将 UTF-16 字节解码为文本串。
// [S-sig 0x1407da960]：params 3 寄存器=[]byte、rets 2 寄存器=string。
func decodeFileLocatorUTF16(data []byte) string {
	return ""
}

// ---- 布尔表达式解析 ----

// [S 0x1407dac60] fileLocatorBooleanParser 布尔表达式递归下降解析器（token 流 + 游标 + 作用域）。
type fileLocatorBooleanParser struct {
	tokens       []fileLocatorToken
	pos          int
	booleanScope string
	matchCase    bool
}

// parseFileLocatorBooleanExpression 解析查询串为布尔表达式。
// [S-sig 0x1407dac60] 汇编实证：query(string:rax/rbx)+booleanScope(string:rcx/rdi)+matchCase(bool:sil)
// 参数；返回 (*fileLocatorBooleanExpression, error)。体：tokenizeFileLocatorBoolean(query) 后装配
// parser（tokens@0x00+pos@0x18+booleanScope@0x20+matchCase@0x30）递归下降，未逐条翻译。
func parseFileLocatorBooleanExpression(query string, booleanScope string, matchCase bool) (*fileLocatorBooleanExpression, error) {
	return nil, nil
}

// tokenizeFileLocatorBoolean 将查询串切分为 token 流。
// [S-sig 0x1407dae80]：命名 tokenize → []fileLocatorToken；rets 3 寄存器=slice。
func tokenizeFileLocatorBoolean(query string) []fileLocatorToken {
	return nil
}

// parseOr 解析 or 层级。
// [S-sig 0x1407dba80]：receiver 唯一(rax 直用)；返回 2 寄存器=fileLocatorBooleanNode 接口（nil 置 type+data）。
func (p *fileLocatorBooleanParser) parseOr() fileLocatorBooleanNode {
	return nil
}

// parseAnd 解析 and 层级。
// [S-sig 0x1407dbbc0]：receiver 唯一(rax 直用)；返回 2 寄存器=fileLocatorBooleanNode 接口。
func (p *fileLocatorBooleanParser) parseAnd() fileLocatorBooleanNode {
	return nil
}

// parseNear 解析 near(邻近) 层级。
// [S-sig 0x1407dbdc0]：receiver 唯一(rax 直用)；返回 2 寄存器=fileLocatorBooleanNode 接口。
func (p *fileLocatorBooleanParser) parseNear() fileLocatorBooleanNode {
	return nil
}

// parseUnary 解析一元 not 层级。
// [S-sig 0x1407dc000]：receiver 唯一(rax 直解引用 parser.pos)；返回 2 寄存器=fileLocatorBooleanNode 接口。
func (p *fileLocatorBooleanParser) parseUnary() fileLocatorBooleanNode {
	return nil
}

// parsePrimary 解析基本项（词项/括号）。
// [S-sig 0x1407dc0e0]：receiver 唯一(rax 直解引用)；返回 2 寄存器=fileLocatorBooleanNode 接口。
func (p *fileLocatorBooleanParser) parsePrimary() fileLocatorBooleanNode {
	return nil
}

// ---- 布尔节点求值 ----

// Eval 词项节点求值。
// [S-sig 0x1407dc3c0]：接口 fileLocatorBooleanNode.Eval(string) bool 固定。
func (n *fileLocatorTermNode) Eval(s string) bool {
	return false
}

// Eval not 节点求值。
// [S-sig 0x1407dc440]：接口 fileLocatorBooleanNode.Eval(string) bool 固定。
func (n *fileLocatorNotNode) Eval(s string) bool {
	return false
}

// Eval 二元节点求值。
// [S-sig 0x1407dc4c0]：接口 fileLocatorBooleanNode.Eval(string) bool 固定。
func (n *fileLocatorBinaryNode) Eval(s string) bool {
	return false
}

// Eval near 节点求值。
// [S-sig 0x1407dc5c0]：接口 fileLocatorBooleanNode.Eval(string) bool 固定。
func (n *fileLocatorNearNode) Eval(s string) bool {
	return false
}

// ---- 区间计算 ----

// fileLocatorProximityRanges 计算邻近命中区间。
// [S-sig 0x1407dc640] 汇编实证：node(fileLocatorBooleanNode 接口 2 word:rax/rbx)+distance(int:rcx)+
// active(bool:rdi) 参数；返回 ([]FileLocatorTextRange, bool)。体：递归 left/right（透传 distance/active），
// 结果按 distance 合并去重，未逐条翻译。
func fileLocatorProximityRanges(node fileLocatorBooleanNode, distance int, active bool) ([]FileLocatorTextRange, bool) {
	return nil, false
}

// HighlightRanges 计算布尔表达式的整体高亮区间。
// [S-sig 0x1407dcdc0]：命名 HighlightRanges → []FileLocatorTextRange；rets 3 寄存器=slice。
func (e *fileLocatorBooleanExpression) HighlightRanges(s string) []FileLocatorTextRange {
	return nil
}

// collectFileLocatorPositiveMatchers 收集表达式中的正项匹配器。
// [S-sig 0x1407dcf60] 汇编实证：node(接口 2 word:rax/rbx)+positive(bool:cl)+out(*[]fileLocatorTermMatcher:rdi)
// 共 4 寄存器；返回 void。体：按 itab hash 分发二元/非/项节点，递归 left/right（非节点 xor positive 翻转），
// 项节点 positive 时 append 到 out（元素 16B=接口），未逐条翻译。
func collectFileLocatorPositiveMatchers(node fileLocatorBooleanNode, positive bool, out *[]fileLocatorTermMatcher) {
}

// newFileLocatorTermMatcher 按查询与模式构造词项匹配器。
// [S-sig 0x1407dd1c0] 汇编实证：query(string:rax/rbx)+mode(string:rcx/rdi)+matchCase(bool:sil) 参数；
// 返回 fileLocatorTermMatcher 接口(2 word)。体：TrimSpace+ToLower(mode)，"like"→wildcardToFileLocatorRegex，
// 其余→compileFileLocatorRegex(regex,matchCase)，未逐条翻译。
func newFileLocatorTermMatcher(query string, mode string, matchCase bool) fileLocatorTermMatcher {
	return nil
}

// ---- 词项匹配器 ----

// Exists 判断字符串是否含词项。
// [S-sig 0x1407dd3e0]：接口 fileLocatorTermMatcher.Exists(string) bool 固定。
func (m *fileLocatorPlainMatcher) Exists(s string) bool {
	return false
}

// Ranges 返回词项命中区间。
// [S-sig 0x1407dd4a0]：接口 fileLocatorTermMatcher.Ranges(string) []FileLocatorTextRange 固定。
func (m *fileLocatorPlainMatcher) Ranges(s string) []FileLocatorTextRange {
	return nil
}

// Exists 判断字符串是否含正则命中。
// [S-sig 0x1407dd520]：接口 fileLocatorTermMatcher.Exists(string) bool 固定。
func (m *fileLocatorRegexMatcher) Exists(s string) bool {
	return false
}

// Ranges 返回正则命中区间。
// [S-sig 0x1407dd5e0]：接口 fileLocatorTermMatcher.Ranges(string) []FileLocatorTextRange 固定。
func (m *fileLocatorRegexMatcher) Ranges(s string) []FileLocatorTextRange {
	return nil
}

// Exists 判断字符串是否含整词命中。
// [S-sig 0x1407dd680]：接口 fileLocatorTermMatcher.Exists(string) bool 固定。
func (m *fileLocatorWholeWordMatcher) Exists(s string) bool {
	return false
}

// Ranges 返回整词命中区间。
// [S-sig 0x1407dd6e0]：接口 fileLocatorTermMatcher.Ranges(string) []FileLocatorTextRange 固定。
func (m *fileLocatorWholeWordMatcher) Ranges(s string) []FileLocatorTextRange {
	return nil
}

// ---- 正则/通配/glob 编译 ----

// compileFileLocatorRegex 编译正则匹配器。
// [S-sig 0x1407ddae0]：命名 compile...Regex → (*regexp.Regexp, error)；params 2 寄存器=pattern。
func compileFileLocatorRegex(pattern string) (*regexp.Regexp, error) {
	return nil, nil
}

// wildcardToFileLocatorRegex 将通配模式转正则串。
// [S-sig 0x1407ddf40]：命名 wildcardTo...Regex → string。
func wildcardToFileLocatorRegex(pattern string) string {
	return ""
}

// fileLocatorContainsWildcard 判断模式是否含通配符。
// [S-sig 0x1407ddfe0]：命名 Contains... → bool；rets=[rax] 单寄存器=bool。
func fileLocatorContainsWildcard(pattern string) bool {
	return false
}

// compileFileLocatorWildcardMatcher 编译通配匹配器。
// [S-sig 0x1407de060]：命名 compile...Matcher → (*regexp.Regexp, error)；rets 2 寄存器。
func compileFileLocatorWildcardMatcher(pattern string) (*regexp.Regexp, error) {
	return nil, nil
}

// compileFileLocatorPathGlobMatcher 编译路径 glob 匹配器。
// [S-sig 0x1407de6a0]：命名 compile...Matcher → (*regexp.Regexp, error)；rets 2 寄存器。
func compileFileLocatorPathGlobMatcher(pattern string) (*regexp.Regexp, error) {
	return nil, nil
}

// fileLocatorPathGlobToRegex 将路径 glob 转正则串。
// [S-sig 0x1407de7c0]：命名 PathGlobToRegex → string；rets 2 寄存器=string。
func fileLocatorPathGlobToRegex(glob string) string {
	return ""
}

// normalizeFileLocatorGlobPattern 归一化 glob 模式。
// [S-sig 0x1407dee60]：命名 normalize...Pattern → string。
func normalizeFileLocatorGlobPattern(pattern string) string {
	return ""
}

// findPlainFileLocatorRanges 查找普通子串命中区间。
// [S-sig 0x1407def80]：params 4 寄存器=haystack(2)+needle(2)，rets 3 寄存器=slice。
func findPlainFileLocatorRanges(haystack string, needle string, matchCase bool) []FileLocatorTextRange {
	return nil
}

// fileLocatorByteRangesToRuneRanges 将字节区间转为 rune 区间。
// [S-sig 0x1407df300]：命名 ByteRangesToRuneRanges → []FileLocatorTextRange。
func fileLocatorByteRangesToRuneRanges(text string, byteRanges []FileLocatorTextRange) []FileLocatorTextRange {
	return nil
}

// mergeFileLocatorRanges 合并重叠/相邻区间。
// [S-sig 0x1407df5e0]：params 3 寄存器=slice，rets 3 寄存器=slice。
func mergeFileLocatorRanges(ranges []FileLocatorTextRange) []FileLocatorTextRange {
	return nil
}

// cloneFileLocatorLineMatches 深拷贝行命中切片。
// [S-sig 0x1407df860]：命名 clone → []FileLocatorLineMatch；params 2 寄存器=slice。
func cloneFileLocatorLineMatches(matches []FileLocatorLineMatch) []FileLocatorLineMatch {
	return nil
}

// normalizeFileLocatorPath 归一化路径（trim/大小写折叠）。
// [S-sig 0x1407dffc0]：命名 normalize...Path → string。
func normalizeFileLocatorPath(path string) string {
	return ""
}

// sameFileLocatorPath 判断两路径归一化后是否相同。
// [S-sig 0x1407e0020]：命名 same... → bool；两路径串入参。
func sameFileLocatorPath(a string, b string) bool {
	return false
}

// truncateFileLocatorSummaryPreview 截断摘要预览串。
// [S-sig 0x1407e00c0]：命名 truncate...Preview → string。
func truncateFileLocatorSummaryPreview(preview string) string {
	return ""
}

// isFileLocatorWordRune 判断 rune 是否为词构成字符。
// [S-sig 0x1407e01a0]：命名 is...WordRune → bool。
func isFileLocatorWordRune(r rune) bool {
	return false
}

// resolveFileLocatorElapsedMilliseconds 由起止时间戳求耗时毫秒。
// [S-sig 0x1407e0260]：命名 resolve...ElapsedMilliseconds → int64；rets=[rax] 单寄存器。
func resolveFileLocatorElapsedMilliseconds(startedAt string, finishedAt string) int64 {
	return 0
}
