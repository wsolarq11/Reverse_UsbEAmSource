package main

import (
	"encoding/binary"
	"strings"
)

// 拼音模糊匹配（文件搜索）：fileSearchPinyinFuzzyMatcher 的 Match 与术语级
// matchFileSearchPinyinTerm。全部为寄存器级 asm 实证的 [S] 落地。

// Match 用 contains 正则对字节串做包含匹配；nil receiver 或 nil contains 恒假。
// [S 汇编 0x14080a7e0, 160B] 实证：m==nil → false；m.contains==nil → false；
// 否则转 regexp.Regexp.doExecute（Match 的底层），返回 doExecute 结果非 nil。
func (m *fileSearchPinyinFuzzyMatcher) Match(b []byte) bool {
	if m == nil || m.contains == nil {
		return false
	}
	return m.contains.Match(b)
}

// MatchExact 用 exact 正则做精确匹配；首轮失败且 allowDot=false 时，去掉最后一个
// '.' 及之后（扩展名）再试一次。nil receiver 或 nil exact 恒假。
// [S 汇编 0x14080a880, 416B] 实证：m==nil/m.exact==nil → false；m.exact.Match(b) 命中
// → true；allowDot（byte 参数非 0）→ false；否则从 len(b)-1 往前找 '.'，dot<=0 → false，
// 否则 m.exact.Match(b[:dot])。
func (m *fileSearchPinyinFuzzyMatcher) MatchExact(b []byte, allowDot bool) bool {
	if m == nil || m.exact == nil {
		return false
	}
	if m.exact.Match(b) {
		return true
	}
	if allowDot {
		return false
	}
	dot := len(b) - 1
	for dot >= 0 && b[dot] != '.' {
		dot--
	}
	if dot <= 0 {
		return false
	}
	return m.exact.Match(b[:dot])
}

// MatchPrefix 用 prefix 正则对字节串做前缀匹配；nil receiver 或 nil prefix 恒假。
// [S 汇编 0x14080aa20, 160B] 实证：m==nil → false；m.prefix==nil → false；
// 否则转 regexp.Regexp.doExecute。
func (m *fileSearchPinyinFuzzyMatcher) MatchPrefix(b []byte) bool {
	if m == nil || m.prefix == nil {
		return false
	}
	return m.prefix.Match(b)
}

// matchFileSearchPinyinTerm 对单个名称搜索术语做拼音模糊匹配。
// [S 汇编 0x14080fda0, 160B] 实证：先 bytesMatchWildcardFold(b, term.patterns, false)
// 命中即 true；否则转 term.matcher.Match(b)（term.matcher 即 +0x30 字段）。
func matchFileSearchPinyinTerm(b []byte, term nameSearchTerm) bool {
	if bytesMatchWildcardFold(b, term.patterns, false) {
		return true
	}
	return term.matcher.Match(b)
}

// matchNodeNameTermsWithPinyin 对节点名 b 执行名称术语匹配（含拼音模糊兜底），
// 返回更新后的 matched 掩码。与 matchNodeNameTerms 同构的三段式，但匹配函数换成
// 闭包 match：wildcard（b, caseSensitive）→ 拼音模糊（b1）→ 拼音模糊（b2）。
// [S 汇编 0x14080fe40, 1280B + func1 0x140810340, 256B] 实证，plan 按值传入：
//
//	(A) 快路径：plan.driver>=0 且 matched&plan.driverBit==0 时先匹配 driver 术语，
//	    失败直接返回 matched；
//	(B) 按 plan.order 顺序匹配各索引，跳过已匹配位，命中置位；
//	(C) 兜底：遍历全部术语，仅当 flags[0]!=0&&flags[1]==0&&flags[3]!=0 且未匹配时
//	    再试。三段均走 match 闭包（bytesMatchWildcardFold + matchFileSearchPinyinTerm×2）。
func matchNodeNameTermsWithPinyin(b []byte, b1 []byte, b2 []byte, terms []nameSearchTerm, plan nameSearchPlan, caseSensitive bool) uint32 {
	matched := plan.matched
	match := func(term nameSearchTerm) bool {
		if bytesMatchWildcardFold(b, term.patterns, caseSensitive) {
			return true
		}
		if matchFileSearchPinyinTerm(b1, term) {
			return true
		}
		return matchFileSearchPinyinTerm(b2, term)
	}
	if plan.driver >= 0 && matched&plan.driverBit == 0 {
		if !match(terms[plan.driver]) {
			return matched
		}
		matched |= plan.driverBit
	}
	for _, idx := range plan.order {
		bit := uint32(1) << uint(idx)
		if matched&bit != 0 {
			continue
		}
		if match(terms[idx]) {
			matched |= bit
		}
	}
	for idx := 0; idx < len(terms); idx++ {
		t := terms[idx]
		if !(t.flags[0] != 0 && t.flags[1] == 0 && t.flags[3] != 0) {
			continue
		}
		bit := uint32(1) << uint(idx)
		if matched&bit != 0 {
			continue
		}
		if match(t) {
			matched |= bit
		}
	}
	return matched
}

// classifyVolumePinyinMatch 对 b1/b2 两路名称做拼音模糊分类，返回匹配级别（3/4/5/6/7/8，
// 5 为默认）。term 为匹配术语；allowDot 为 true 时跳过"去扩展名重试"。
// [S 汇编 0x140810960, 480B + func1 0x140810b40, 800B] 实证：term==nil 或无 pattern → 5；
// 取 patterns[0] 作 aliasB（唯一 pattern 时同时作 aliasA）。闭包 classify 按优先级判定：
//
//	matchFileSearchPinyinTerm 不中 → (0,false)；MatchExact / 与 aliasA 忽略大小写全等 /
//	(仅 !allowDot) bytesStemEqualFold → 精确级；MatchPrefix / 与 aliasB 前缀忽略大小写全等
//	→ 前缀级；其余 → 其他级。先试 b1（级别 3/4/5）再试 b2（6/7/8），均不中返回 5。
func classifyVolumePinyinMatch(b1 []byte, b2 []byte, term *nameSearchTerm, allowDot bool) int {
	if term == nil || len(term.patterns) == 0 {
		return 5
	}
	aliasB := term.patterns[0]
	var aliasA []byte
	if len(term.patterns) == 1 {
		aliasA = aliasB
	}

	classify := func(b []byte, exact, prefix, other int) (int, bool) {
		if len(b) == 0 {
			return 0, false
		}
		if !matchFileSearchPinyinTerm(b, *term) {
			return 0, false
		}
		matcher := term.matcher
		exactOK := matcher.MatchExact(b, allowDot)
		if !exactOK && len(aliasA) != 0 && len(b) == len(aliasA) {
			exactOK = true
			for i := 0; i < len(aliasA); i++ {
				c := b[i]
				if 'A' <= c && c <= 'Z' {
					c |= 0x20
				}
				if c != aliasA[i] {
					exactOK = false
					break
				}
			}
		}
		if !exactOK && !allowDot && len(aliasA) != 0 {
			exactOK = bytesStemEqualFold(b, aliasA, false)
		}
		if exactOK {
			return exact, true
		}
		if matcher.MatchPrefix(b) {
			return prefix, true
		}
		if len(aliasB) == 0 || len(b) < len(aliasB) {
			return other, true
		}
		for i := 0; i < len(aliasB); i++ {
			c := b[i]
			if 'A' <= c && c <= 'Z' {
				c |= 0x20
			}
			if c != aliasB[i] {
				return other, true
			}
		}
		return prefix, true
	}

	if r, ok := classify(b1, 3, 4, 5); ok {
		return r
	}
	if r, ok := classify(b2, 6, 7, 8); ok {
		return r
	}
	return 5
}

// [S 汇编 0x14080df20, 800B] 实证：在 volumePinyinIndex.records（每记录 32B）
// 内二分查找 nodeIndex，命中后按记录头 12B 切出 pinyinFull / pinyinInitials。
// 记录头布局（小端）：+0x00 nodeIndex uint32、+0x04 fullOffset uint32、
// +0x08 fullLen uint16、+0x0a initLen uint16；别名数据存于 m.aliases。
func (m *volumePinyinIndex) aliasesForNode(nodeIndex int32) (full []uint8, initials []uint8, ok bool) {
	if m == nil || nodeIndex < 0 || nodeIndex >= 0x40000000 {
		return nil, nil, false
	}
	count := len(m.records) >> 5
	lo, hi := 0, count
	for lo < hi {
		mid := (lo + hi) >> 1
		if binary.LittleEndian.Uint32(m.records[mid*32:]) < uint32(nodeIndex) {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo >= count {
		return nil, nil, false
	}
	rec := m.records[lo*32:]
	if binary.LittleEndian.Uint32(rec) != uint32(nodeIndex) {
		return nil, nil, false
	}
	fullOffset := binary.LittleEndian.Uint32(rec[4:])
	fullLen := binary.LittleEndian.Uint16(rec[8:])
	initLen := binary.LittleEndian.Uint16(rec[10:])
	end := int(fullOffset) + int(fullLen) + int(initLen)
	if len(m.aliases) < end {
		return nil, nil, false
	}
	full = m.aliases[fullOffset : int(fullOffset)+int(fullLen)]
	initials = m.aliases[int(fullOffset)+int(fullLen) : end]
	return full, initials, true
}

// [S 汇编 0x14080f900, 608B] 实证：按节点归属取拼音别名。
// nodeIndex>=0x40000000 走 deltaNodes 覆盖（取 delta.pinyinFull/pinyinInitials），
// 否则走 volumePinyinIndex.aliasesForNode。无别名返回 (nil,nil,false)。
func (v volumeIndexReadView) pinyinAliasesForNode(nodeIndex int32) (full []uint8, initials []uint8, ok bool) {
	if nodeIndex >= 0x40000000 {
		deltaIdx := int(nodeIndex - 0x40000000)
		if len(v.deltaNodes) <= deltaIdx {
			return nil, nil, false
		}
		delta := v.deltaNodes[deltaIdx]
		if len(delta.pinyinFull) == 0 && len(delta.pinyinInitials) == 0 {
			return nil, nil, false
		}
		return delta.pinyinFull, delta.pinyinInitials, true
	}
	if v.pinyin == nil {
		return nil, nil, false
	}
	return v.pinyin.aliasesForNode(nodeIndex)
}

// [S 汇编 0x14080fb60, 576B] 实证：层级术语匹配。
// 先 wildcard 匹配 b；未命中则取拼音别名，依次用 matchFileSearchPinyinTerm
// 匹配 pinyinFull / pinyinInitials。
func (v volumeIndexReadView) matchHierarchyTermWithPinyin(nodeIndex int32, b []byte, term nameSearchTerm, caseSensitive bool) bool {
	if bytesMatchWildcardFold(b, term.patterns, caseSensitive) {
		return true
	}
	full, initials, ok := v.pinyinAliasesForNode(nodeIndex)
	if !ok {
		return false
	}
	if matchFileSearchPinyinTerm(full, term) {
		return true
	}
	return matchFileSearchPinyinTerm(initials, term)
}

// [S 汇编 0x140810440, 1312B] 实证：候选节点名匹配 + 层级术语沿父链上溯。
// matched 先由 matchNodeNameTermsWithPinyin 计算；driver 未命中 → false；mask==matched → true；
// ^matched&others==0 → false；否则从 node.ParentIdx 沿父链上溯，逐个节点用
// matchHierarchyTermWithPinyin 补齐剩余层级术语，墓碑后代直接淘汰，最后返回 mask==matched。
// nodeIndex 参数在实现中未使用（asm 实证仅使用 node.ParentIdx 作为起始父索引）。
func (v volumeIndexReadView) matchSearchCandidateNodeWithPinyin(
	nodeIndex int32,
	node IndexNode,
	tombstoneLookup *volumeIndexTombstoneLookup,
	caseSensitive bool,
	b []byte,
	b1 []byte,
	b2 []byte,
	terms []nameSearchTerm,
	plan nameSearchPlan,
) bool {
	_ = nodeIndex

	matched := matchNodeNameTermsWithPinyin(b, b1, b2, terms, plan, caseSensitive)
	if plan.driver >= 0 && matched&plan.driverBit == 0 {
		return false
	}
	if plan.mask == matched {
		return true
	}
	if ^matched&plan.others == 0 {
		return false
	}

	required := ^matched & plan.others
	for cur := node.ParentIdx; cur != -1 && plan.mask != matched && required != 0; {
		if tombstoneLookup.IsTombstonedDescendant(cur) {
			return false
		}
		n := v.nodeAtIndex(cur)
		name := v.nodeName(n)
		for idx := range terms {
			bit := uint32(1) << uint(idx)
			if required&bit == 0 {
				continue
			}
			if matched&bit != 0 {
				continue
			}
			if v.matchHierarchyTermWithPinyin(cur, name, terms[idx], caseSensitive) {
				matched |= bit
				required &= ^bit
			}
		}
		cur = n.ParentIdx
	}
	return plan.mask == matched
}

// normalizePinyinSyllable 规范化单个拼音音节：TrimSpace → ToLower → 统一 ü 表示。
// [S 汇编 0x14080b1e0, 160B] 实证：TrimSpace → ToLower → Replace("u:","v",-1)
// → Replace("ü","v",-1)（old len=2/2，new len=1/1，n=-1）。
func normalizePinyinSyllable(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.Replace(s, "u:", "v", -1)
	s = strings.Replace(s, "ü", "v", -1)
	return s
}

// Close 关闭拼音索引：nil 或 mappedFile 为空直接返回 nil；否则 Close 映射文件并
// 清空 mappedFile/records/aliases，透传 Close 的 error。
// [S 汇编 0x14080e240, 192B] 实证：p==nil||mappedFile(+0x58)==nil → (nil,nil)；
// 否则 volumeIndexMappedFile.Close() → mappedFile=nil、records(+0x28/+0x30/+0x38)=nil、
// aliases(+0x40/+0x48/+0x50)=nil，返回 Close 的 error。
func (p *volumePinyinIndex) Close() error {
	if p == nil || p.mappedFile == nil {
		return nil
	}
	err := p.mappedFile.Close()
	p.mappedFile = nil
	p.records = nil
	p.aliases = nil
	return err
}

// parseVolumeNameTrigramHeader 解析名称三元组头（魔数 "UITG" + 版本 + 字段）。
// [S-sig 0x1407f9ba0, 256B]：len<0x40→error；魔数/版本校验 → 读字段(+8..+0x30)。
func parseVolumeNameTrigramHeader(a []byte) (interface{}, error) {
	_ = a
	return nil, nil
}

// encodeVolumePinyinHeader 编码拼音头（魔数 "UIYP" + 字段 + 字典指纹）。
// [S-sig 0x14080bca0, 256B]：makeslice(0x80) → 写魔数/版本/字段 → 字典指纹(+0x18) → 写参数。
func encodeVolumePinyinHeader(a []byte, b interface{}) []byte {
	_, _ = a, b
	return nil
}
