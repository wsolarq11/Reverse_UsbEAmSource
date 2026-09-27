package main

import (
	"encoding/binary"
	"regexp"
	"testing"
)

func TestMatchNodeNameTermsWithPinyin(t *testing.T) {
	re := regexp.MustCompile("pin")
	matcher := &fileSearchPinyinFuzzyMatcher{contains: re}

	// 术语：flags[0]!=0 且 flags[3]!=0 表示兜底循环参与；matcher 指针 +0x30
	terms := []nameSearchTerm{
		{patterns: [][]byte{[]byte("alpha")}, flags: [8]byte{1, 0, 0, 1}, matcher: matcher}, // 兜底参与
		{patterns: [][]byte{[]byte("beta")}, flags: [8]byte{1, 0, 0, 1}, matcher: matcher},  // 兜底参与
	}
	plan := nameSearchPlan{
		matched:   0,
		driver:    -1,
		driverBit: 0,
		order:     []int{0},
	}

	b := []byte("ALPHA")
	// order 循环：idx0 匹配 wildcard（大小写不敏感）→ matched |= 1
	got := matchNodeNameTermsWithPinyin(b, []byte("nope1"), []byte("nope2"), terms, plan, false)
	if got != 1 {
		t.Fatalf("order loop expected matched=1, got %d", got)
	}

	// 兜底循环：order 空，两个术语都 flags[0]=1,flags[1]=0,flags[3]=1 参与；
	// wildcard 不命中，但 b1 拼音模糊 matcher 命中 "pin" → 两术语都置位。
	plan2 := nameSearchPlan{matched: 0, driver: -1, order: []int{}}
	got = matchNodeNameTermsWithPinyin([]byte("zzz"), []byte("xxpinyy"), []byte("nope"), terms, plan2, false)
	if got != 3 {
		t.Fatalf("fallback pinyin loop expected matched=3, got %d", got)
	}

	// 快路径失败：driver=0 且 b wildcard 不命中且两个拼音都不命中 → 返回 matched(0)
	plan3 := nameSearchPlan{matched: 0, driver: 0, driverBit: 1, order: []int{}}
	got = matchNodeNameTermsWithPinyin([]byte("zzz"), []byte("nope"), []byte("nope"), terms, plan3, false)
	if got != 0 {
		t.Fatalf("fast path fail expected matched=0, got %d", got)
	}
}

func TestClassifyVolumePinyinMatch(t *testing.T) {
	matcher := &fileSearchPinyinFuzzyMatcher{
		contains: regexp.MustCompile("abc"),
		exact:    regexp.MustCompile("^exact$"),
		prefix:   regexp.MustCompile("^pre"),
	}

	// 精确级：b1 通配符命中 exact → MatchExact 命中 → 3
	term := &nameSearchTerm{patterns: [][]byte{[]byte("exact")}, matcher: matcher}
	if got := classifyVolumePinyinMatch([]byte("exact"), []byte("zzz"), term, false); got != 3 {
		t.Fatalf("exact level expected 3, got %d", got)
	}

	// 前缀级：b1 通配符命中 pre* 的别名但 MatchExact 失败、MatchPrefix 命中 → 4
	term2 := &nameSearchTerm{patterns: [][]byte{[]byte("pre")}, matcher: matcher}
	if got := classifyVolumePinyinMatch([]byte("preabc"), []byte("zzz"), term2, false); got != 4 {
		t.Fatalf("prefix level expected 4, got %d", got)
	}

	// 组 B 前缀全等：MatchPrefix 失败但 patterns[0] 是 b1 前缀 → 前缀级 4
	term3 := &nameSearchTerm{patterns: [][]byte{[]byte("ab")}, matcher: matcher}
	if got := classifyVolumePinyinMatch([]byte("abc"), []byte("zzz"), term3, false); got != 4 {
		t.Fatalf("aliasB prefix expected 4, got %d", got)
	}

	// 其他级：contains 命中但一切精确/前缀/组 B 都不中 → 5
	term4 := &nameSearchTerm{patterns: [][]byte{[]byte("zz")}, matcher: matcher}
	if got := classifyVolumePinyinMatch([]byte("abc"), []byte("zzz"), term4, false); got != 5 {
		t.Fatalf("other level expected 5, got %d", got)
	}

	// 二次回退：b1 全不中，b2 命中 exact → 6
	if got := classifyVolumePinyinMatch([]byte("zzz"), []byte("exact"), term, false); got != 6 {
		t.Fatalf("second pass exact expected 6, got %d", got)
	}

	// nil 术语 / 空 pattern → 5
	if got := classifyVolumePinyinMatch([]byte("a"), []byte("b"), nil, false); got != 5 {
		t.Fatalf("nil term expected 5, got %d", got)
	}
	term5 := &nameSearchTerm{patterns: [][]byte{}, matcher: matcher}
	if got := classifyVolumePinyinMatch([]byte("a"), []byte("b"), term5, false); got != 5 {
		t.Fatalf("empty patterns expected 5, got %d", got)
	}
}

func makePinyinRecord(nodeIndex, fullOffset uint32, fullLen, initLen uint16) []byte {
	rec := make([]byte, 32)
	binary.LittleEndian.PutUint32(rec[0:], nodeIndex)
	binary.LittleEndian.PutUint32(rec[4:], fullOffset)
	binary.LittleEndian.PutUint16(rec[8:], fullLen)
	binary.LittleEndian.PutUint16(rec[10:], initLen)
	return rec
}

func TestVolumePinyinAliasesForNode(t *testing.T) {
	// records: node 10 -> full@0 len3 / init len2；node 20 -> full@5 len2 / init len1
	recs := append(makePinyinRecord(10, 0, 3, 2), makePinyinRecord(20, 5, 2, 1)...)
	aliases := []byte("fulinXYZ")
	m := &volumePinyinIndex{records: recs, aliases: aliases}

	full, initials, ok := m.aliasesForNode(10)
	if !ok || string(full) != "ful" || string(initials) != "in" {
		t.Fatalf("node 10 got full=%q initials=%q ok=%v", full, initials, ok)
	}
	full, initials, ok = m.aliasesForNode(20)
	if !ok || string(full) != "XY" || string(initials) != "Z" {
		t.Fatalf("node 20 got full=%q initials=%q ok=%v", full, initials, ok)
	}

	// 缺配 / 越界 / 护栏
	if _, _, ok = m.aliasesForNode(15); ok {
		t.Fatalf("missing node 15 expected ok=false")
	}
	if _, _, ok = m.aliasesForNode(25); ok {
		t.Fatalf("out-of-range node 25 expected ok=false")
	}
	if _, _, ok = (*volumePinyinIndex)(nil).aliasesForNode(0); ok {
		t.Fatalf("nil receiver expected ok=false")
	}
	if _, _, ok = m.aliasesForNode(-1); ok {
		t.Fatalf("negative node expected ok=false")
	}
	if _, _, ok = m.aliasesForNode(0x40000000); ok {
		t.Fatalf("delta-range node expected ok=false")
	}

	// aliases 不足：node 20 需 end=8，改用短别名
	short := &volumePinyinIndex{records: recs, aliases: []byte("fulinX")}
	if _, _, ok = short.aliasesForNode(20); ok {
		t.Fatalf("short aliases expected ok=false")
	}

	// delta 覆盖路径：nodeIndex>=0x40000000 走 deltaNodes
	delta := volumeIndexDeltaNode{pinyinFull: []byte("pf"), pinyinInitials: []byte("pi")}
	v := volumeIndexReadView{deltaNodes: []volumeIndexDeltaNode{delta}}
	full, initials, ok = v.pinyinAliasesForNode(0x40000000)
	if !ok || string(full) != "pf" || string(initials) != "pi" {
		t.Fatalf("delta override got full=%q initials=%q ok=%v", full, initials, ok)
	}
	if _, _, ok = (volumeIndexReadView{}).pinyinAliasesForNode(0x40000000); ok {
		t.Fatalf("empty deltaNodes expected ok=false")
	}
	vEmpty := volumeIndexReadView{deltaNodes: []volumeIndexDeltaNode{{}}}
	if _, _, ok = vEmpty.pinyinAliasesForNode(0x40000000); ok {
		t.Fatalf("empty aliases in delta expected ok=false")
	}

	// 普通路径：pinyin 转发
	vIdx := volumeIndexReadView{pinyin: m}
	if full, initials, ok = vIdx.pinyinAliasesForNode(10); !ok || string(full) != "ful" {
		t.Fatalf("pinyin forward got full=%q initials=%q ok=%v", full, initials, ok)
	}
	if _, _, ok = (volumeIndexReadView{}).pinyinAliasesForNode(10); ok {
		t.Fatalf("nil pinyin expected ok=false")
	}
}

func TestMatchHierarchyTermWithPinyin(t *testing.T) {
	recs := append(makePinyinRecord(10, 0, 3, 2), makePinyinRecord(20, 5, 2, 1)...)
	aliases := []byte("fulinXYZ")
	m := &volumePinyinIndex{records: recs, aliases: aliases}
	v := volumeIndexReadView{pinyin: m}

	// wildcard 命中即 true
	termWild := &nameSearchTerm{patterns: [][]byte{[]byte("ful")}}
	if !v.matchHierarchyTermWithPinyin(10, []byte("ful"), *termWild, false) {
		t.Fatalf("wildcard match expected true")
	}

	// wildcard 不命中，拼音别名 full 命中 matcher
	matcher := &fileSearchPinyinFuzzyMatcher{contains: regexp.MustCompile("ful")}
	termPin := &nameSearchTerm{patterns: [][]byte{[]byte("zz")}, matcher: matcher}
	if !v.matchHierarchyTermWithPinyin(10, []byte("ful"), *termPin, false) {
		t.Fatalf("pinyin full match expected true")
	}

	// 别名存在但都不命中 → false
	matcher2 := &fileSearchPinyinFuzzyMatcher{contains: regexp.MustCompile("zzz")}
	termPin2 := &nameSearchTerm{patterns: [][]byte{[]byte("zz")}, matcher: matcher2}
	if v.matchHierarchyTermWithPinyin(10, []byte("nope"), *termPin2, false) {
		t.Fatalf("no alias match expected false")
	}

	// 无别名（pinyin nil）→ false
	if (volumeIndexReadView{}).matchHierarchyTermWithPinyin(10, []byte("ful"), *termPin, false) {
		t.Fatalf("nil pinyin expected false")
	}
}

func TestMatchSearchCandidateNodeWithPinyin(t *testing.T) {
	mkTerm := func(pattern string, flags [8]byte) nameSearchTerm {
		return nameSearchTerm{patterns: [][]byte{[]byte(pattern)}, flags: flags}
	}

	// 场景 1：driver 未命中 → false
	terms1 := []nameSearchTerm{mkTerm("zzz", [8]byte{}), mkTerm("foo", [8]byte{})}
	plan1 := nameSearchPlan{mask: 0x3, driver: 0, driverBit: 0x1, others: 0x2}
	if (volumeIndexReadView{}).matchSearchCandidateNodeWithPinyin(0, IndexNode{}, nil, false, []byte("foo"), nil, nil, terms1, plan1) {
		t.Fatalf("driver miss expected false")
	}

	// 场景 2：mask == matched → true
	terms2 := []nameSearchTerm{mkTerm("foo", [8]byte{})}
	plan2 := nameSearchPlan{mask: 0x1, driver: -1, order: []int{0}}
	if !(volumeIndexReadView{}).matchSearchCandidateNodeWithPinyin(0, IndexNode{}, nil, false, []byte("foo"), nil, nil, terms2, plan2) {
		t.Fatalf("mask==matched expected true")
	}

	// 场景 3：~matched & others == 0 → false
	terms3 := []nameSearchTerm{mkTerm("foo", [8]byte{}), mkTerm("bar", [8]byte{})}
	plan3 := nameSearchPlan{mask: 0x3, driver: -1, others: 0x1, order: []int{0}}
	if (volumeIndexReadView{}).matchSearchCandidateNodeWithPinyin(0, IndexNode{}, nil, false, []byte("foo"), nil, nil, terms3, plan3) {
		t.Fatalf("others done but mask incomplete expected false")
	}

	// 场景 4：沿父链匹配层级术语 → true
	node0 := IndexNode{FRN: 0, NameOffset: 0, ParentIdx: 1, NameLen: 3}
	node1 := IndexNode{FRN: 1, NameOffset: 3, ParentIdx: -1, NameLen: 3}
	v := volumeIndexReadView{nodes: []IndexNode{node0, node1}, namePool: []byte("foobar")}
	tl := &volumeIndexTombstoneLookup{memo: map[int32]bool{}}
	terms4 := []nameSearchTerm{mkTerm("foo", [8]byte{}), mkTerm("bar", [8]byte{2: 1})}
	plan4 := nameSearchPlan{mask: 0x3, driver: -1, others: 0x2, order: []int{0}}
	if !v.matchSearchCandidateNodeWithPinyin(0, node0, tl, false, []byte("foo"), nil, nil, terms4, plan4) {
		t.Fatalf("hierarchy term matched on parent expected true")
	}

	// 场景 5：父链命中墓碑 → false
	tlTomb := &volumeIndexTombstoneLookup{
		volumeIndexReadView: volumeIndexReadView{tombstones: map[int32]struct{}{1: {}}},
		memo:                map[int32]bool{},
	}
	if v.matchSearchCandidateNodeWithPinyin(0, node0, tlTomb, false, []byte("foo"), nil, nil, terms4, plan4) {
		t.Fatalf("parent tombstoned expected false")
	}
}
