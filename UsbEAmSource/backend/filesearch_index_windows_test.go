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

func TestDecodeVolumeIndexNodesFromBytesWithNodeCount(t *testing.T) {
	nodes := []IndexNode{
		{FRN: 100, NameOffset: 0, ParentIdx: -1, ModTime: 200, NameLen: 3, Flags: 1},
		{FRN: 101, NameOffset: 3, ParentIdx: 0, ModTime: 201, NameLen: 4, Flags: 0},
		{FRN: 102, NameOffset: 7, ParentIdx: 0, ModTime: 202, NameLen: 2, Flags: 2},
	}
	got, active, err := decodeVolumeIndexNodesFromBytesWithNodeCount(makeNodeBytes(nodes), 3)
	if err != nil {
		t.Fatalf("decode err=%v, want nil", err)
	}
	if len(got) != 3 {
		t.Fatalf("len=%d, want 3", len(got))
	}
	for i := range nodes {
		if got[i] != nodes[i] {
			t.Fatalf("node[%d]=%+v, want %+v", i, got[i], nodes[i])
		}
	}
	// Flags&2 的节点不计入 active
	if active != 2 {
		t.Fatalf("active=%d, want 2", active)
	}

	// 空流：makeslice(0,0) 返回非 nil 空切片（与 asm 无条件 makeslice 一致）
	if g, a, err := decodeVolumeIndexNodesFromBytesWithNodeCount(nil, 0); err != nil || len(g) != 0 || a != 0 {
		t.Fatalf("empty decode got (%v,%d,%v), want (len 0,0,nil)", g, a, err)
	}

	// len%24!=0
	if _, _, err := decodeVolumeIndexNodesFromBytesWithNodeCount(make([]byte, 25), 2); err != errVolumeIndexNodeBytesInvalid {
		t.Fatalf("len%%24!=0 err=%v, want sentinel", err)
	}

	// nodeCount < derived
	if _, _, err := decodeVolumeIndexNodesFromBytesWithNodeCount(makeNodeBytes(nodes), 2); err != errVolumeIndexNodeBytesInvalid {
		t.Fatalf("nodeCount<derived err=%v, want sentinel", err)
	}

	// ParentIdx >= nodeCount（显式 nodeCount 校验）
	bad := []IndexNode{{FRN: 1, ParentIdx: 5, Flags: 0}}
	if _, _, err := decodeVolumeIndexNodesFromBytesWithNodeCount(makeNodeBytes(bad), 3); err != errVolumeIndexNodeBytesInvalid {
		t.Fatalf("parentIdx>=nodeCount err=%v, want sentinel", err)
	}

	// ParentIdx < -1
	bad2 := []IndexNode{{FRN: 1, ParentIdx: -2, Flags: 0}}
	if _, _, err := decodeVolumeIndexNodesFromBytesWithNodeCount(makeNodeBytes(bad2), 3); err != errVolumeIndexNodeBytesInvalid {
		t.Fatalf("parentIdx<-1 err=%v, want sentinel", err)
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

func TestAddNamePrefixBucketCountSkipsEmptyTwoByte(t *testing.T) {
	// twoByte 为空（nil）时只累加 firstByte，bigram 累加应被跳过（不 panic）。
	var c namePrefixBucketCounts
	addNamePrefixBucketCount(&c, "foo", 1)
	if c.firstByte['f'] != 1 {
		t.Fatalf("firstByte['f']=%d, want 1", c.firstByte['f'])
	}
	if c.firstByte['o'] != 0 {
		t.Fatalf("firstByte['o']=%d, want 0 (非首字节)", c.firstByte['o'])
	}
}

func TestBuildNamePrefixBucketCounts(t *testing.T) {
	namePool := []byte("foobar")
	nodes := []IndexNode{
		{FRN: 100, NameOffset: 0, ParentIdx: -1, NameLen: 3, Flags: 1}, // "foo"
		{FRN: 101, NameOffset: 3, ParentIdx: 0, NameLen: 3, Flags: 0},  // "bar"
		{FRN: 102, NameOffset: 0, ParentIdx: 0, NameLen: 3, Flags: 2},  // "foo" deleted
	}
	got := buildNamePrefixBucketCounts(nodes, namePool)
	if got.firstByte['f'] != 1 {
		t.Fatalf("firstByte['f']=%d, want 1", got.firstByte['f'])
	}
	if got.firstByte['b'] != 1 {
		t.Fatalf("firstByte['b']=%d, want 1", got.firstByte['b'])
	}
	if got.firstByte['o'] != 0 {
		t.Fatalf("firstByte['o']=%d, want 0 (非首字节)", got.firstByte['o'])
	}
	if len(got.twoByte) != 0 {
		t.Fatalf("twoByte len=%d, want 0 (nodes<=10000 不预分配)", len(got.twoByte))
	}
}

func TestBuildNamePrefixBucketCountsPreallocatesTwoByte(t *testing.T) {
	// >10000 节点触发 twoByte 预分配（即使全 deleted，循环内不访问 namePool）。
	nodes := make([]IndexNode, 10001)
	for i := range nodes {
		nodes[i] = IndexNode{Flags: 2}
	}
	got := buildNamePrefixBucketCounts(nodes, nil)
	if len(got.twoByte) != 0x10000 {
		t.Fatalf("twoByte len=%d, want 0x10000", len(got.twoByte))
	}
}

func TestBuildNamePrefixBucketCountsFromNodeBytes(t *testing.T) {
	namePool := []byte("foobar")
	nodes := []IndexNode{
		{FRN: 100, NameOffset: 0, ParentIdx: -1, NameLen: 3, Flags: 1}, // "foo"
		{FRN: 101, NameOffset: 3, ParentIdx: 0, NameLen: 3, Flags: 0},  // "bar"
		{FRN: 102, NameOffset: 0, ParentIdx: 0, NameLen: 3, Flags: 2},  // "foo" deleted
	}
	got, active, err := buildNamePrefixBucketCountsFromNodeBytes(makeNodeBytes(nodes), namePool)
	if err != nil {
		t.Fatalf("err=%v, want nil", err)
	}
	if active != 2 {
		t.Fatalf("active=%d, want 2", active)
	}
	if got.firstByte['f'] != 1 {
		t.Fatalf("firstByte['f']=%d, want 1", got.firstByte['f'])
	}
	if got.firstByte['b'] != 1 {
		t.Fatalf("firstByte['b']=%d, want 1", got.firstByte['b'])
	}

	if _, _, err := buildNamePrefixBucketCountsFromNodeBytes(make([]byte, 25), namePool); err != errVolumeIndexNodeBytesInvalid {
		t.Fatalf("len%%24!=0 err=%v, want sentinel", err)
	}

	bad := []IndexNode{{FRN: 1, ParentIdx: 5, Flags: 0, NameOffset: 0, NameLen: 0}}
	if _, _, err := buildNamePrefixBucketCountsFromNodeBytes(makeNodeBytes(bad), namePool); err != errVolumeIndexNodeBytesInvalid {
		t.Fatalf("parentIdx>=derived err=%v, want sentinel", err)
	}

	bad2 := []IndexNode{{FRN: 1, ParentIdx: -2, Flags: 0}}
	if _, _, err := buildNamePrefixBucketCountsFromNodeBytes(makeNodeBytes(bad2), namePool); err != errVolumeIndexNodeBytesInvalid {
		t.Fatalf("parentIdx<-1 err=%v, want sentinel", err)
	}

	bad3 := []IndexNode{{FRN: 1, ParentIdx: -1, Flags: 0, NameOffset: 100, NameLen: 10}}
	if _, _, err := buildNamePrefixBucketCountsFromNodeBytes(makeNodeBytes(bad3), namePool); err != errVolumeIndexNodeBytesInvalid {
		t.Fatalf("namePool overflow err=%v, want sentinel", err)
	}
}

func TestNewVolumeIndexReadViewFromCheckpointView(t *testing.T) {
	if got := newVolumeIndexReadViewFromCheckpointView(nil); got.root != "" || got.prefixBuckets != nil {
		t.Fatalf("nil cv expected zero view, got %+v", got)
	}

	namePool := []byte("foobar")
	nodes := []IndexNode{
		{FRN: 100, NameOffset: 0, ParentIdx: -1, NameLen: 3, Flags: 1},
		{FRN: 101, NameOffset: 3, ParentIdx: 0, NameLen: 3, Flags: 0},
	}

	pb := &namePrefixBucketCounts{}
	cv1 := &volumeIndexCheckpointView{
		root:          "C:\\vol",
		nodes:         nodes,
		nodeBytes:     makeNodeBytes(nodes),
		sortedIndices: []int32{0, 1},
		sortedBytes:   []byte{0, 0, 0, 0},
		namePool:      nil,
		namePoolBytes: namePool,
		activeCount:   7,
		prefixBuckets: pb,
	}
	got1 := newVolumeIndexReadViewFromCheckpointView(cv1)
	if got1.root != "C:\\vol" {
		t.Fatalf("root=%q, want C:\\vol", got1.root)
	}
	if len(got1.nodeBytes) != 0 {
		t.Fatalf("nodes 非空应清零 nodeBytes, got len=%d", len(got1.nodeBytes))
	}
	if string(got1.namePool) != "foobar" {
		t.Fatalf("namePool=%q, want foobar (fallback)", got1.namePool)
	}
	if got1.activeCount != 7 {
		t.Fatalf("activeCount=%d, want 7", got1.activeCount)
	}
	if got1.prefixBuckets != pb {
		t.Fatal("prefixBuckets 应复用传入指针")
	}

	cv2 := &volumeIndexCheckpointView{
		root:          "C:\\vol2",
		nodes:         nodes,
		nodeBytes:     makeNodeBytes(nodes),
		namePool:      namePool,
		namePoolBytes: []byte("unused"),
		activeCount:   3,
	}
	got2 := newVolumeIndexReadViewFromCheckpointView(cv2)
	if got2.prefixBuckets == nil {
		t.Fatal("prefixBuckets 应被构建")
	}
	if got2.prefixBuckets.firstByte['f'] != 1 {
		t.Fatalf("prefixBuckets.firstByte['f']=%d, want 1", got2.prefixBuckets.firstByte['f'])
	}
	if string(got2.namePool) != "foobar" {
		t.Fatalf("namePool=%q, want foobar (不回退)", got2.namePool)
	}

	cv3 := &volumeIndexCheckpointView{
		root:      "C:\\vol3",
		nodeBytes: makeNodeBytes(nodes),
		namePool:  namePool,
	}
	got3 := newVolumeIndexReadViewFromCheckpointView(cv3)
	if len(got3.nodeBytes) == 0 {
		t.Fatal("nodes 空应保留 nodeBytes")
	}
}
