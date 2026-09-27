package main

import (
	"encoding/binary"
	"testing"
)

func TestBytesMatchWildcardFoldEmptyPatterns(t *testing.T) {
	if !bytesMatchWildcardFold([]byte("anything"), nil, false) {
		t.Fatal("空 patterns 应返回 true")
	}
	if !bytesMatchWildcardFold(nil, [][]byte{}, false) {
		t.Fatal("空 b + 空 patterns 应返回 true")
	}
}

func TestBytesMatchWildcardFoldSinglePattern(t *testing.T) {
	// 单 pattern 退化为 bytesContainsFold，大小写折叠仅作用 b 侧。
	if !bytesMatchWildcardFold([]byte("Hello World"), [][]byte{[]byte("hello")}, false) {
		t.Fatal("折叠匹配应命中")
	}
	if bytesMatchWildcardFold([]byte("Hello World"), [][]byte{[]byte("hello")}, true) {
		t.Fatal("大小写敏感时不应命中")
	}
	if bytesMatchWildcardFold([]byte("abc"), [][]byte{[]byte("abcd")}, false) {
		t.Fatal("b 短于 pattern 应返回 false")
	}
}

func TestBytesMatchWildcardFoldSequential(t *testing.T) {
	// foo*bar*baz 顺序拼接匹配语义。
	if !bytesMatchWildcardFold([]byte("fooXbarYbaz"), [][]byte{[]byte("foo"), []byte("bar"), []byte("baz")}, false) {
		t.Fatal("顺序片段应全部命中")
	}
	if !bytesMatchWildcardFold([]byte("FooXBarYBaz"), [][]byte{[]byte("foo"), []byte("bar"), []byte("baz")}, false) {
		t.Fatal("大小写折叠顺序片段应命中")
	}
	if bytesMatchWildcardFold([]byte("FooXBarYBaz"), [][]byte{[]byte("foo"), []byte("bar"), []byte("baz")}, true) {
		t.Fatal("大小写敏感顺序片段不应命中")
	}
}

func TestBytesMatchWildcardFoldMissingMiddle(t *testing.T) {
	if bytesMatchWildcardFold([]byte("fooXquxYbaz"), [][]byte{[]byte("foo"), []byte("bar"), []byte("baz")}, false) {
		t.Fatal("中间片段缺失应返回 false")
	}
}

func TestBytesMatchWildcardFoldSkipEmptyPattern(t *testing.T) {
	if !bytesMatchWildcardFold([]byte("foobar"), [][]byte{[]byte("foo"), []byte{}, []byte("bar")}, false) {
		t.Fatal("空片段应被跳过")
	}
}

func TestBytesMatchWildcardFoldPartialConsumption(t *testing.T) {
	// 每个片段找「首个」匹配并消耗：foo 首次出现后，bar 须在其后出现。
	if !bytesMatchWildcardFold([]byte("foo1foo2bar"), [][]byte{[]byte("foo"), []byte("bar")}, false) {
		t.Fatal("首个 foo 后跟 bar 应命中")
	}
	// bar 只出现在第一个 foo 之前，消耗首个 foo 后找不到 bar → false。
	if bytesMatchWildcardFold([]byte("barfoo"), [][]byte{[]byte("foo"), []byte("bar")}, false) {
		t.Fatal("bar 在消耗 foo 后不存在，应返回 false")
	}
}

func makeNodeBytes(nodes []IndexNode) []byte {
	buf := make([]byte, 0, len(nodes)*24)
	for _, n := range nodes {
		var rec [24]byte
		binary.LittleEndian.PutUint64(rec[0:8], n.FRN)
		binary.LittleEndian.PutUint32(rec[8:12], n.NameOffset)
		binary.LittleEndian.PutUint32(rec[12:16], uint32(n.ParentIdx))
		binary.LittleEndian.PutUint32(rec[16:20], n.ModTime)
		binary.LittleEndian.PutUint16(rec[20:22], n.NameLen)
		binary.LittleEndian.PutUint16(rec[22:24], n.Flags)
		buf = append(buf, rec[:]...)
	}
	return buf
}

func TestNodeAtAndNodeAtIndex(t *testing.T) {
	node := IndexNode{FRN: 100, NameOffset: 5, ParentIdx: -1, ModTime: 200, NameLen: 3, Flags: 1}

	// nodeBytes 序列化路径
	vBytes := volumeIndexReadView{nodeBytes: makeNodeBytes([]IndexNode{node})}
	if got := vBytes.nodeAt(0); got != node {
		t.Fatalf("nodeBytes path got %+v, want %+v", got, node)
	}

	// nodes 切片路径
	vNodes := volumeIndexReadView{nodes: []IndexNode{node}}
	if got := vNodes.nodeAt(0); got != node {
		t.Fatalf("nodes path got %+v, want %+v", got, node)
	}

	// delta 覆盖：nodeIndex>=0x40000000
	delta := volumeIndexDeltaNode{node: IndexNode{FRN: 42}}
	vDelta := volumeIndexReadView{deltaNodes: []volumeIndexDeltaNode{delta}}
	if got := vDelta.nodeAtIndex(0x40000000); got.FRN != 42 {
		t.Fatalf("delta override got FRN=%d, want 42", got.FRN)
	}

	// 越界 / 负值 → 零值
	if got := vNodes.nodeAtIndex(5); got.FRN != 0 {
		t.Fatalf("out-of-range expected zero, got FRN=%d", got.FRN)
	}
	if got := vNodes.nodeAtIndex(-1); got.FRN != 0 {
		t.Fatalf("negative expected zero, got FRN=%d", got.FRN)
	}
	if got := (volumeIndexReadView{}).nodeAtIndex(0x40000000); got.FRN != 0 {
		t.Fatalf("empty delta expected zero, got FRN=%d", got.FRN)
	}
}

func TestNodeName(t *testing.T) {
	// 普通路径：namePool 切片
	v := volumeIndexReadView{namePool: []byte("helloworld")}
	if got := v.nodeName(IndexNode{NameOffset: 5, NameLen: 5}); string(got) != "world" {
		t.Fatalf("namePool slice got %q, want %q", got, "world")
	}
	if got := v.nodeName(IndexNode{NameOffset: 8, NameLen: 5}); got != nil {
		t.Fatalf("out-of-range namePool expected nil, got %q", got)
	}

	// delta 路径：deltaByFRN 命中
	delta := volumeIndexDeltaNode{node: IndexNode{FRN: 99}, name: []byte("dname")}
	vDelta := volumeIndexReadView{
		deltaNodes: []volumeIndexDeltaNode{delta},
		deltaByFRN: map[uint64]int32{99: 0x40000000},
	}
	deltaNode := IndexNode{FRN: 99, NameOffset: 0xFFFFFFFF}
	if got := vDelta.nodeName(deltaNode); string(got) != "dname" {
		t.Fatalf("deltaByFRN hit got %q, want %q", got, "dname")
	}

	// delta 路径：线性扫描 fallback
	vScan := volumeIndexReadView{deltaNodes: []volumeIndexDeltaNode{delta}}
	if got := vScan.nodeName(deltaNode); string(got) != "dname" {
		t.Fatalf("linear scan got %q, want %q", got, "dname")
	}

	// Flags&2 跳过
	deltaSkip := volumeIndexDeltaNode{node: IndexNode{FRN: 99, Flags: 2}, name: []byte("skip")}
	vSkip := volumeIndexReadView{
		deltaNodes: []volumeIndexDeltaNode{deltaSkip},
		deltaByFRN: map[uint64]int32{99: 0x40000000},
	}
	if got := vSkip.nodeName(deltaNode); got != nil {
		t.Fatalf("flags&2 skip expected nil, got %q", got)
	}
}

func TestIsTombstonedDescendant(t *testing.T) {
	// 父链：5→4→3→1→-1，2→1→-1
	nodes := []IndexNode{
		{FRN: 0, ParentIdx: -1},
		{FRN: 1, ParentIdx: -1},
		{FRN: 2, ParentIdx: 1},
		{FRN: 3, ParentIdx: 1},
		{FRN: 4, ParentIdx: 3},
		{FRN: 5, ParentIdx: 4},
	}
	tl := &volumeIndexTombstoneLookup{
		volumeIndexReadView: volumeIndexReadView{
			nodes:      nodes,
			tombstones: map[int32]struct{}{3: {}},
		},
		memo: map[int32]bool{},
	}

	if !tl.IsTombstonedDescendant(5) {
		t.Fatal("node 5 → 4 → 3 (tombstone) expected true")
	}
	if !tl.memo[5] || !tl.memo[4] {
		t.Fatal("path nodes 5/4 should be memoized true")
	}
	if tl.IsTombstonedDescendant(2) {
		t.Fatal("node 2 → 1 → -1 (root) expected false")
	}

	// 护栏
	if tl.IsTombstonedDescendant(-1) {
		t.Fatal("negative node expected false")
	}
	empty := &volumeIndexTombstoneLookup{memo: map[int32]bool{}}
	if empty.IsTombstonedDescendant(0) {
		t.Fatal("empty tombstones expected false")
	}

	// delta 覆盖：deltaIdx 的 node.ParentIdx 是 tombstone
	delta := volumeIndexDeltaNode{node: IndexNode{ParentIdx: 3}}
	tlDelta := &volumeIndexTombstoneLookup{
		volumeIndexReadView: volumeIndexReadView{
			nodes:      nodes,
			tombstones: map[int32]struct{}{3: {}},
			deltaNodes: []volumeIndexDeltaNode{delta},
		},
		memo: map[int32]bool{},
	}
	if !tlDelta.IsTombstonedDescendant(0x40000000) {
		t.Fatal("delta node → parent 3 (tombstone) expected true")
	}
}

func TestMatchSearchCandidateNode(t *testing.T) {
	mkTerm := func(pattern string, flags [8]byte) nameSearchTerm {
		return nameSearchTerm{patterns: [][]byte{[]byte(pattern)}, flags: flags}
	}

	// 场景 1：driver 未命中 → false
	terms1 := []nameSearchTerm{mkTerm("zzz", [8]byte{}), mkTerm("foo", [8]byte{})}
	plan1 := nameSearchPlan{mask: 0x3, driver: 0, driverBit: 0x1, others: 0x2}
	if (volumeIndexReadView{}).matchSearchCandidateNode(IndexNode{}, []byte("foo"), terms1, plan1, nil, false) {
		t.Fatal("driver miss expected false")
	}

	// 场景 2：mask == matched → true
	terms2 := []nameSearchTerm{mkTerm("foo", [8]byte{})}
	plan2 := nameSearchPlan{mask: 0x1, driver: -1, order: []int{0}}
	if !(volumeIndexReadView{}).matchSearchCandidateNode(IndexNode{}, []byte("foo"), terms2, plan2, nil, false) {
		t.Fatal("mask==matched expected true")
	}

	// 场景 3：~matched & others == 0 → false
	terms3 := []nameSearchTerm{mkTerm("foo", [8]byte{}), mkTerm("bar", [8]byte{})}
	plan3 := nameSearchPlan{mask: 0x3, driver: -1, others: 0x1, order: []int{0}}
	if (volumeIndexReadView{}).matchSearchCandidateNode(IndexNode{}, []byte("foo"), terms3, plan3, nil, false) {
		t.Fatal("others done but mask incomplete expected false")
	}

	// 场景 4：沿父链匹配层级术语 → true
	node0 := IndexNode{FRN: 0, NameOffset: 0, ParentIdx: 1, NameLen: 3}
	node1 := IndexNode{FRN: 1, NameOffset: 3, ParentIdx: -1, NameLen: 3}
	v := volumeIndexReadView{nodes: []IndexNode{node0, node1}, namePool: []byte("foobar")}
	tl := &volumeIndexTombstoneLookup{memo: map[int32]bool{}}
	terms4 := []nameSearchTerm{mkTerm("foo", [8]byte{}), mkTerm("bar", [8]byte{2: 1})}
	plan4 := nameSearchPlan{mask: 0x3, driver: -1, others: 0x2, order: []int{0}}
	if !v.matchSearchCandidateNode(node0, []byte("foo"), terms4, plan4, tl, false) {
		t.Fatal("hierarchy term matched on parent expected true")
	}

	// 场景 5：父链命中墓碑 → false
	tlTomb := &volumeIndexTombstoneLookup{
		volumeIndexReadView: volumeIndexReadView{tombstones: map[int32]struct{}{1: {}}},
		memo:                map[int32]bool{},
	}
	if v.matchSearchCandidateNode(node0, []byte("foo"), terms4, plan4, tlTomb, false) {
		t.Fatal("parent tombstoned expected false")
	}
}
