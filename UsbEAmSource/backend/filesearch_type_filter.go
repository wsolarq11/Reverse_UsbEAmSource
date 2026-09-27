package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// normalizeFileSearchTypeFilterID 归一化类型过滤器 ID。
// [S] 0x140815840：ToLower(TrimSpace(s)) 空→""；逐 rune 只允许 a-z/0-9/'-'/'_'，含其他→""；否则返回 s。
func normalizeFileSearchTypeFilterID(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return ""
	}
	return s
}

// normalizeFileSearchTypeRules 归一化类型过滤规则列表（去重，最多 128 条）。
// [S] 0x140815900：nil→(nil,true)；逐条 normalizeFileSearchTypeRuleToken 无效则跳过；
// ToLower 后 map 去重，追加到 out；返回 (out,true)。
func normalizeFileSearchTypeRules(rules []string) ([]string, bool) {
	if rules == nil {
		return nil, true
	}
	out := make([]string, 0, len(rules))
	seen := make(map[string]struct{})
	n := len(rules)
	if n > 128 {
		n = 128
	}
	for i := 0; i < n; i++ {
		tok, ok := normalizeFileSearchTypeRuleToken(rules[i])
		if !ok {
			continue
		}
		tok = strings.ToLower(tok)
		if _, dup := seen[tok]; dup {
			continue
		}
		seen[tok] = struct{}{}
		out = append(out, tok)
	}
	return out, true
}

// normalizeFileSearchTypeRuleToken 归一化单条类型过滤规则 token。
// [S] 0x140815be0：TrimSpace 空→("",false)；rune 数>64→("",false)；含 '/','\\',':' 或控制字符→("",false)；
// ToLower；HasPrefix(".*") 时 rest=s[2:]（空→("",false)；rest 含 "*?[" 走 glob 校验，否则返回 (rest,isValid(rest))）；
// 否则含 "*?[" 走 glob 校验；否则 TrimLeft(".") 后空→("",false)，isValid 则返回 (s,true)。
func normalizeFileSearchTypeRuleToken(token string) (string, bool) {
	s := strings.TrimSpace(token)
	if s == "" {
		return "", false
	}
	if utf8.RuneCountInString(s) > 64 {
		return "", false
	}
	for _, r := range s {
		if r == '/' || r == '\\' || r == ':' {
			return "", false
		}
		if r <= 0xff && unicode.IsControl(r) {
			return "", false
		}
	}
	s = strings.ToLower(s)
	if strings.HasPrefix(s, ".*") {
		if len(s) == 2 {
			return "", false
		}
		rest := s[2:]
		if strings.IndexAny(rest, "*?[") >= 0 {
			if isValidFileSearchTypeGlob(s) {
				return s, true
			}
			return "", false
		}
		return rest, isValidNormalizedExtensionToken(rest)
	}
	if strings.IndexAny(s, "*?[") >= 0 {
		if isValidFileSearchTypeGlob(s) {
			return s, true
		}
		return "", false
	}
	s = strings.TrimLeft(s, ".")
	if s == "" {
		return "", false
	}
	if isValidNormalizedExtensionToken(s) {
		return s, true
	}
	return "", false
}

// isValidNormalizedExtensionToken 校验扩展名 token 合法性。
// [S] 0x140815ec0：空→false；逐 rune 只允许 a-z/0-9/'.'/'-'/'_'/'+'，其他→false；
// s[0]=='.' 或 s[len-1]=='.'→false；strings.Index(s,"..")>=0→false；否则 true。
func isValidNormalizedExtensionToken(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') ||
			r == '.' || r == '-' || r == '_' || r == '+' {
			continue
		}
		return false
	}
	if s[0] == '.' || s[len(s)-1] == '.' {
		return false
	}
	if strings.Index(s, "..") >= 0 {
		return false
	}
	return true
}

// isValidFileSearchTypeGlob 校验文件类型 glob 模式。
// [S] 0x140815fe0：空→false；IndexAny("(){}|+^$\\")>=0→false；逐字节：
// '/','\\',']'→false；'[' 解析字符类（可选 '!'/'^' 前缀，找 ']' 过程中禁 '/','\\'，
// 未闭合→false，类内容交给 isValidASCIICharClass），其余普通字符继续。
func isValidFileSearchTypeGlob(s string) bool {
	if s == "" {
		return false
	}
	if strings.IndexAny(s, "(){}|+^$\\") >= 0 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '/' || c == '\\' || c == ']' {
			return false
		}
		if c != '[' {
			continue
		}
		start := i + 1
		if start >= len(s) {
			return false
		}
		if s[start] == '!' || s[start] == '^' {
			start++
		}
		if start >= len(s) {
			return false
		}
		j := start
		for j < len(s) {
			if s[j] == ']' {
				break
			}
			if s[j] == '/' || s[j] == '\\' {
				return false
			}
			j++
		}
		if j >= len(s) {
			return false
		}
		if !isValidASCIICharClass(s[i+1 : j]) {
			return false
		}
		i = j
	}
	return true
}

// isValidASCIICharClass 校验 glob 字符类内部（不含方括号）。
// [S] 0x140816180：空→false；可选前导 '!'/'^'（仅此一个字符→false）；逐字节：
// 非可打印 ASCII(0x20-0x7e) 或 '/','\\'→false；'-' 须位于中间且两端为字母数字且升序，否则 false。
func isValidASCIICharClass(s string) bool {
	if s == "" {
		return false
	}
	if s[0] == '!' || s[0] == '^' {
		if len(s) == 1 {
			return false
		}
		s = s[1:]
	}
	isAlnum := func(b byte) bool {
		return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c > 0x7e {
			return false
		}
		if c == '/' || c == '\\' {
			return false
		}
		if c != '-' {
			continue
		}
		if i == 0 || i == len(s)-1 {
			return false
		}
		prev := s[i-1]
		next := s[i+1]
		if !isAlnum(prev) || !isAlnum(next) {
			return false
		}
		if next < prev {
			return false
		}
	}
	return true
}

// classifyNormalizedFileSearchTypeRule 分类归一化规则。
// [S] 0x1408162c0：rule=="?*.[rz][0-9][0-9]"→3（分卷）；IndexAny("*?[")>=0→2（glob）；
// Index(".")>=0→1（复合扩展名）；否则 0（简单扩展名）；均原样回传 rule。
func classifyNormalizedFileSearchTypeRule(rule string) (int, string) {
	if rule == "?*.[rz][0-9][0-9]" {
		return 3, rule
	}
	if strings.IndexAny(rule, "*?[") >= 0 {
		return 2, rule
	}
	if strings.Index(rule, ".") >= 0 {
		return 1, rule
	}
	return 0, rule
}

// compileFileSearchRuleMatcher 编译类型过滤规则为匹配器。
// [S] 0x1408163c0：normalizeFileSearchTypeRules 失败→error；空→{empty:true,signature:"all"}；
// 否则按 classify 填充 simpleByHash(FNV-1a 小写 hash)/compoundByLastByte(末字节小写)/splitArchiveFast/glob；
// glob 规则逐条 fileSearchGlobToAnchoredRegexp 包装 "(?:re)"，Join("|") 后
// 拼 "(?i)^(?:"+join+")$" 用 regexp.Compile；最后签名 fileSearchRuleMatcherSignature(normalized,m)。
func compileFileSearchRuleMatcher(rules []string) (*fileSearchRuleMatcher, error) {
	normalized, ok := normalizeFileSearchTypeRules(rules)
	if !ok {
		return nil, fmt.Errorf("文件类型规则无效")
	}
	if len(normalized) == 0 {
		return &fileSearchRuleMatcher{
			empty:     true,
			signature: "all",
		}, nil
	}
	m := &fileSearchRuleMatcher{
		simpleByHash:       make(map[uint64][][]uint8),
		compoundByLastByte: make(map[uint8][][]uint8),
	}
	simpleSeen := make(map[string]struct{})
	compoundSeen := make(map[string]struct{})
	globSeen := make(map[string]struct{})
	var globRules []string
	for _, rule := range normalized {
		class, r := classifyNormalizedFileSearchTypeRule(rule)
		switch class {
		case 0:
			if _, dup := simpleSeen[r]; dup {
				continue
			}
			simpleSeen[r] = struct{}{}
			h := uint64(0xcbf29ce484222325)
			for i := 0; i < len(r); i++ {
				c := r[i]
				if c >= 'A' && c <= 'Z' {
					c += 0x20
				}
				h ^= uint64(c)
				h *= 0x100000001b3
			}
			m.simpleByHash[h] = append(m.simpleByHash[h], []uint8(r))
			m.simpleCount++
		case 1:
			if _, dup := compoundSeen[r]; dup {
				continue
			}
			compoundSeen[r] = struct{}{}
			key := r[len(r)-1]
			if key >= 'A' && key <= 'Z' {
				key += 0x20
			}
			m.compoundByLastByte[key] = append(m.compoundByLastByte[key], []uint8("."+r))
			m.compoundCount++
		case 2:
			if _, dup := globSeen[r]; dup {
				continue
			}
			globSeen[r] = struct{}{}
			globRules = append(globRules, r)
			m.globCount++
		case 3:
			m.splitArchiveFast = true
		}
	}
	if len(globRules) > 0 {
		anchored := make([]string, 0, len(globRules))
		for _, g := range globRules {
			re, err := fileSearchGlobToAnchoredRegexp(g)
			if err != nil {
				return nil, err
			}
			anchored = append(anchored, "(?:"+re+")")
		}
		pattern := "(?i)^(?:" + strings.Join(anchored, "|") + ")$"
		glob, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("编译文件类型 glob 失败: %w", err)
		}
		m.glob = glob
	}
	m.signature = fileSearchRuleMatcherSignature(normalized, m)
	return m, nil
}

// fileSearchRuleMatcherSignature 计算匹配器签名（排序规则 + 统计字段的 sha256 前 8 字节 hex）。
// [S] 0x140816fe0：copy+sort.Strings；json.Marshal 结构体（失败则 strings.Join(sorted,"|")）；
// sha256.Sum256 取前 8 字节 hex 编码（16 字符）。
func fileSearchRuleMatcherSignature(rules []string, m *fileSearchRuleMatcher) string {
	sorted := make([]string, len(rules))
	copy(sorted, rules)
	sort.Strings(sorted)
	payload := struct {
		Rules            []string
		SimpleCount      int
		CompoundCount    int
		GlobCount        int
		SplitArchiveFast bool
	}{
		Rules:            sorted,
		SimpleCount:      m.simpleCount,
		CompoundCount:    m.compoundCount,
		GlobCount:        m.globCount,
		SplitArchiveFast: m.splitArchiveFast,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		data = []byte(strings.Join(sorted, "|"))
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:8])
}

// Allow 判断文件名是否命中类型过滤规则。
// [S] 0x140817280：m==nil 或 !m.empty→true（放行）；isDir 或 name==""→false；
// matchSimpleExtension||matchCompoundExtension||(splitArchiveFast&&matchSplitArchiveExtension)
// ||(glob!=nil&&glob.MatchString(name))。
func (m *fileSearchRuleMatcher) Allow(name string, stemLen int, isDir bool) bool {
	if m == nil || !m.empty {
		return true
	}
	if isDir || name == "" {
		return false
	}
	if m.matchSimpleExtension(name, stemLen) {
		return true
	}
	if m.matchCompoundExtension(name, stemLen) {
		return true
	}
	if m.splitArchiveFast && matchSplitArchiveExtension(name) {
		return true
	}
	if m.glob != nil && m.glob.MatchString(name) {
		return true
	}
	return false
}

// matchSimpleExtension 匹配简单扩展名（最后一个点之后的部分）。
// [S] 0x140817420：simpleByHash 空→false；末尾向前找 '.'（无/末位→false）；
// dot<stemLen 取 name[dot+1:]，否则取 name[0:]；空→false；FNV-1a 小写 hash 查表，逐字节忽略大小写比较。
func (m *fileSearchRuleMatcher) matchSimpleExtension(name string, stemLen int) bool {
	if len(m.simpleByHash) == 0 {
		return false
	}
	dot := -1
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			dot = i
			break
		}
	}
	if dot < 0 || dot == len(name)-1 {
		return false
	}
	extStart := dot + 1
	if dot >= stemLen {
		extStart = 0
	}
	ext := name[extStart:]
	if ext == "" {
		return false
	}
	h := uint64(0xcbf29ce484222325)
	for i := 0; i < len(ext); i++ {
		c := ext[i]
		if c >= 'A' && c <= 'Z' {
			c += 0x20
		}
		h ^= uint64(c)
		h *= 0x100000001b3
	}
	for _, candidate := range m.simpleByHash[h] {
		if len(candidate) != len(ext) {
			continue
		}
		eq := true
		for k := 0; k < len(ext); k++ {
			ca := ext[k]
			if ca >= 'A' && ca <= 'Z' {
				ca += 0x20
			}
			cb := candidate[k]
			if cb >= 'A' && cb <= 'Z' {
				cb += 0x20
			}
			if ca != cb {
				eq = false
				break
			}
		}
		if eq {
			return true
		}
	}
	return false
}

// matchCompoundExtension 匹配复合扩展名（末字节索引 + 后缀比较）。
// [S] 0x1408175e0：compoundByLastByte 空或 name==""→false；末字节小写查表；
// len(name)<len(ext) 跳过；offset=len(name)-len(ext)，offset<stemLen 时取 0；
// name[offset:offset+len(ext)] 与 ext 逐字节忽略大小写比较。
func (m *fileSearchRuleMatcher) matchCompoundExtension(name string, stemLen int) bool {
	if len(m.compoundByLastByte) == 0 || name == "" {
		return false
	}
	key := name[len(name)-1]
	if key >= 'A' && key <= 'Z' {
		key += 0x20
	}
	for _, ext := range m.compoundByLastByte[key] {
		if len(name) < len(ext) {
			continue
		}
		offset := len(name) - len(ext)
		if offset < stemLen {
			offset = 0
		}
		eq := true
		for k := 0; k < len(ext); k++ {
			ca := name[offset+k]
			if ca >= 'A' && ca <= 'Z' {
				ca += 0x20
			}
			cb := ext[k]
			if cb >= 'A' && cb <= 'Z' {
				cb += 0x20
			}
			if ca != cb {
				eq = false
				break
			}
		}
		if eq {
			return true
		}
	}
	return false
}

// fileSearchGlobToAnchoredRegexp 将文件类型 glob 转为正则片段。
// [S] 0x1408177e0：isValidFileSearchTypeGlob 失败→fmt.Errorf("非法文件类型 glob: %s")；
// '*'→".*"（连续折叠）、'?'→"."、"[...]"→字符类（前导 '!'/'^'→'^'，'\\' 与 ']' 转义）、
// 普通字符 utf8.DecodeRuneInString 后 regexp.QuoteMeta；非法 UTF-8→error，未闭合→error。
func fileSearchGlobToAnchoredRegexp(glob string) (string, error) {
	if !isValidFileSearchTypeGlob(glob) {
		return "", fmt.Errorf("非法文件类型 glob: %s", glob)
	}
	var b []byte
	for i := 0; i < len(glob); {
		switch glob[i] {
		case '*':
			for i < len(glob) && glob[i] == '*' {
				i++
			}
			b = append(b, '.', '*')
		case '?':
			b = append(b, '.')
			i++
		case '[':
			j := strings.IndexByte(glob[i+1:], ']')
			if j < 0 {
				return "", fmt.Errorf("未闭合字符类")
			}
			class := glob[i+1 : i+1+j]
			b = append(b, '[')
			if len(class) > 0 && (class[0] == '!' || class[0] == '^') {
				b = append(b, '^')
				class = class[1:]
			}
			for k := 0; k < len(class); k++ {
				if class[k] == '\\' || class[k] == ']' {
					b = append(b, '\\')
				}
				b = append(b, class[k])
			}
			b = append(b, ']')
			i = i + 1 + j + 1
		default:
			r, size := utf8.DecodeRuneInString(glob[i:])
			if r == utf8.RuneError && size == 1 {
				return "", fmt.Errorf("文件类型 glob 不是有效 UTF-8")
			}
			b = append(b, regexp.QuoteMeta(string(r))...)
			i += size
		}
	}
	return string(b), nil
}
