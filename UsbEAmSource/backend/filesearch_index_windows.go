package main

import (
	"encoding/binary"
	"errors"
	"os"
	"sort"
	"time"

	"golang.org/x/sys/windows"
)

// errInvalidWALOffset 无效 WAL 偏移哨兵错误（f==nil 或 offset<0）。
// 精确消息字符串待取证（.data 静态 error 值 @0x14193c650 为 typeOff 编码，未解码）。
var errInvalidWALOffset = errors.New("invalid WAL offset")

// bytesContainsFold 在字节切片 b 中查找子串 sub；caseSensitive=false 时对 b 的
// 字节做 ASCII 大写→小写折叠（sub 由调用方保证已小写）。
// [S 汇编 0x1407e1260, 224B] 实证：
//
//	len(sub)==0 → true；len(b)<len(sub) → false；first=sub[0]；
//	双循环：外 i∈[0,len(b)-len(sub)]，内 j∈[1,len(sub))；匹配成功 → true，全失败 → false。
//	折叠（test r9b/r9b；jne 跳过）：'A'<=c<='Z' 时 c|=0x20。
func bytesContainsFold(b, sub []byte, caseSensitive bool) bool {
	if len(sub) == 0 {
		return true
	}
	if len(b) < len(sub) {
		return false
	}
	first := sub[0]
	for i := 0; i <= len(b)-len(sub); i++ {
		c := b[i]
		if !caseSensitive && 'A' <= c && c <= 'Z' {
			c |= 0x20
		}
		if c == first {
			j := 1
			for j < len(sub) {
				cj := b[i+j]
				if !caseSensitive && 'A' <= cj && cj <= 'Z' {
					cj |= 0x20
				}
				if cj != sub[j] {
					break
				}
				j++
			}
			if j == len(sub) {
				return true
			}
		}
	}
	return false
}

// bytesStemEqualFold 比较 b 的"主体"（最后一个 '.' 之前的部分）与 sub；主体长度
// 须等于 len(sub)。caseSensitive=false 时对 b 做 ASCII 大写折叠。
// [S 汇编 0x1407e1340, 224B] 实证：
//
//	从 len(b)-1 往前找 '.'，dot<=0 → false；len(sub)!=dot → false；
//	否则逐字节比较 b[0:dot] 与 sub，全等 → true，任一不等 → false。
func bytesStemEqualFold(b, sub []byte, caseSensitive bool) bool {
	dot := len(b) - 1
	for dot >= 0 && b[dot] != '.' {
		dot--
	}
	if dot <= 0 || len(sub) != dot {
		return false
	}
	for i := 0; i < dot; i++ {
		c := b[i]
		if !caseSensitive && 'A' <= c && c <= 'Z' {
			c |= 0x20
		}
		if c != sub[i] {
			return false
		}
	}
	return true
}

// bytesMatchWildcardFold 在字节切片 b 中按通配符片段顺序匹配；patterns 为按 '*' 分割后的
// 片段序列（调用方保证每片段已小写）。caseSensitive=false 时对 b 做 ASCII 大写折叠
// （sub 侧由调用方保证小写，折叠仅作用 b 侧）。
// [S 汇编 0x1407e3e60, 480B] 实证：
//
//	len(patterns)==0 → true；len(patterns)==1 → bytesContainsFold(b, patterns[0], caseSensitive)；
//	否则顺序遍历 patterns：空片段跳过（continue）；每片段在 b 中找首个匹配起点 k
//	（内层逐字节比较，折叠仅作用 b 侧），找到则 b = b[k+len(p):] 继续下一片段，
//	找不到 → false；全部片段匹配成功 → true。
func bytesMatchWildcardFold(b []byte, patterns [][]byte, caseSensitive bool) bool {
	if len(patterns) == 0 {
		return true
	}
	if len(patterns) == 1 {
		return bytesContainsFold(b, patterns[0], caseSensitive)
	}
	for _, p := range patterns {
		if len(p) == 0 {
			continue
		}
		matched := false
		for k := 0; k <= len(b)-len(p); k++ {
			j := 0
			for j < len(p) {
				cb := b[k+j]
				if !caseSensitive && 'A' <= cb && cb <= 'Z' {
					cb |= 0x20
				}
				if cb != p[j] {
					break
				}
				j++
			}
			if j == len(p) {
				b = b[k+len(p):]
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

// ---- 名称搜索计划域（文件搜索索引匹配） ----
// buildIndexSearchPlan 先计算匹配掩码 / driver 术语 / 匹配顺序，matchNodeNameTerms
// 再据此对候选节点名做 wildcard 匹配。全部为寄存器级 asm 实证的 [S] 落地。

// nameSearchTerm 名称搜索术语（56 字节，asm 元素步长 0x38）。
// 布局经 estimateSearchTermCandidateCount / selectDriverTermIndex /
// buildOrderedNameTermIndices / buildIndexSearchPlan / matchNodeNameTerms /
// shouldPrioritizeSearchTerm / matchFileSearchPinyinTerm 七处寄存器级 asm 实证：
//
//	+0x00..+0x0f 保留（上游 buildNamePrefixBucketCounts 等构建，本域不访问）
//	+0x10..+0x27 patterns [][]byte（wildcard 片段，调用方保证已小写）
//	+0x28..+0x2f flags [8]byte：
//	  [0]=+0x28 与 [1]=+0x29 组合区分"模式术语"与"字面量候选术语"：
//	    buildIndexSearchPlan 以 flags[0]||flags[1] 判模式术语并走 bytesMatchWildcardFold；
//	    selectDriverTermIndex/buildOrderedNameTermIndices 以 flags[0]==0&&flags[1]==0 判候选。
//	  [2]=+0x2a 可选位（buildIndexSearchPlan 据此置 others 掩码）。
//	  [3]=+0x2b 必需位（matchNodeNameTerms 兜底循环判 flags[0]!=0&&flags[1]==0&&flags[3]!=0）。
//	  [4..7]=+0x2c..+0x2f 保留（本域不访问）。
//	+0x30..+0x37 matcher *fileSearchPinyinFuzzyMatcher（拼音模糊匹配器，
//	  matchFileSearchPinyinTerm 在 wildcard 未命中时转交 matcher.Match 兜底）。
type nameSearchTerm struct {
	_        [16]byte
	patterns [][]byte
	flags    [8]byte
	matcher  *fileSearchPinyinFuzzyMatcher
}

// nameFrequencyIndex 是 namePrefixBucketCounts 的别名。buildIndexSearchPlan 的
// driver/order 选择用它估算术语候选计数：前 0x400 字节 firstByte[256]int32 作 unigram，
// twoByte 切片作 bigram 表（len(twoByte)==0x10000 表示 bigram 就绪）。asm 0x1407e3100
// 实证布局与 namePrefixBucketCounts 完全一致（调用点传的是 volumeIndexReadView.prefixBuckets）。
type nameFrequencyIndex = namePrefixBucketCounts

// nameSearchPlan 名称搜索计划（56 字节）。buildIndexSearchPlan 的返回结构，
// matchNodeNameTerms 消费其 matched / driver / driverBit / order 字段。
type nameSearchPlan struct {
	mask      uint32
	matched   uint32
	others    uint32
	_         uint32
	driver    int
	driverBit uint32
	_         uint32
	order     []int
}

// shouldPrioritizeSearchTerm 判断术语 i 是否应排在术语 j 之前（sort less 语义）。
// [S 汇编 0x1407e19c0, 704B] 实证，按优先级依次比较：
//
//	1. 双方候选计数都 >0 且不等 → 计数更小者优先（return countJ > countI）；
//	2. 单片段（len(patterns)==1）vs 多片段 → 单片段优先；
//	3. 全部片段字节总长不等 → 总长更短者优先；
//	4. 片段数不等 → 片段更少者优先；
//	5. 平局 → 索引更大者优先（return j > i）。
func shouldPrioritizeSearchTerm(terms []nameSearchTerm, i, j, countI, countJ int) bool {
	if countI > 0 && countJ > 0 && countI != countJ {
		return countJ > countI
	}
	ti := terms[i]
	tj := terms[j]
	multiI := len(ti.patterns) > 1
	multiJ := len(tj.patterns) > 1
	if multiI != multiJ {
		return !multiI
	}
	sumI := 0
	for _, p := range ti.patterns {
		sumI += len(p)
	}
	sumJ := 0
	for _, p := range tj.patterns {
		sumJ += len(p)
	}
	if sumI != sumJ {
		return sumJ < sumI
	}
	if len(ti.patterns) != len(tj.patterns) {
		return len(ti.patterns) < len(tj.patterns)
	}
	return j > i
}

// estimateSearchTermCandidateCount 估算术语 termIdx 的候选节点数；无法估算时返回
// fallback。 [S 汇编 0x1407e3100, 320B] 实证：
//
//	termIdx 越界 → fallback；取第一个非空 pattern p；
//	len(p)>=2 且 idx.bigramReady==0x10000 → 用 p[0..2]（ASCII 大写→小写）查 bigram，
//	命中 >0 返回该计数；len(p)!=0 → 用 p[0] 查 unigram，命中 >0 返回该计数；否则 fallback。
func estimateSearchTermCandidateCount(terms []nameSearchTerm, idx *nameFrequencyIndex, fallback int, termIdx int) int {
	if termIdx < 0 || termIdx >= len(terms) {
		return fallback
	}
	var p []byte
	for _, c := range terms[termIdx].patterns {
		if len(c) != 0 {
			p = c
			break
		}
	}
	if len(p) >= 2 && idx != nil && len(idx.twoByte) == 0x10000 {
		c0 := int(p[0])
		if 'A' <= p[0] && p[0] <= 'Z' {
			c0 |= 0x20
		}
		c1 := int(p[1])
		if 'A' <= p[1] && p[1] <= 'Z' {
			c1 |= 0x20
		}
		if n := idx.twoByte[c0<<8|c1]; n > 0 {
			return int(n)
		}
	}
	if len(p) != 0 && idx != nil {
		c0 := int(p[0])
		if 'A' <= p[0] && p[0] <= 'Z' {
			c0 |= 0x20
		}
		if n := idx.firstByte[c0]; n > 0 {
			return int(n)
		}
	}
	return fallback
}

// selectDriverTermIndex 选出 driver 术语索引（候选计数最小的字面量术语），无候选返
// 回 -1。 [S 汇编 0x1407e3240, 352B] 实证：跳过 flags[0]||flags[1] 的模式术语；
// 首个字面量术语直接当选，其后用 shouldPrioritizeSearchTerm 择优替换。
func selectDriverTermIndex(terms []nameSearchTerm, idx *nameFrequencyIndex, fallback int) int {
	best := -1
	for i := range terms {
		if terms[i].flags[0] != 0 || terms[i].flags[1] != 0 {
			continue
		}
		cur := estimateSearchTermCandidateCount(terms, idx, fallback, i)
		if best == -1 {
			best = i
			continue
		}
		bestCount := estimateSearchTermCandidateCount(terms, idx, fallback, best)
		if shouldPrioritizeSearchTerm(terms, i, best, cur, bestCount) {
			best = i
		}
	}
	return best
}

// buildOrderedNameTermIndices 收集除 driverIdx 外的字面量术语索引，并按
// shouldPrioritizeSearchTerm 稳定排序。 [S 汇编 0x1407e33a0, 544B + func1 0x1407e35c0,
// 288B] 实证：make([]int,0,len(terms))；跳过 i==driverIdx 及 flags[0]||flags[1]；
// sort.SliceStable 的 less 对 result[a]/result[b] 分别估算候选计数后比较。
func buildOrderedNameTermIndices(terms []nameSearchTerm, idx *nameFrequencyIndex, fallback int, driverIdx int) []int {
	result := make([]int, 0, len(terms))
	for i := range terms {
		if i == driverIdx || terms[i].flags[0] != 0 || terms[i].flags[1] != 0 {
			continue
		}
		result = append(result, i)
	}
	sort.SliceStable(result, func(a, b int) bool {
		ia := result[a]
		ib := result[b]
		ca := estimateSearchTermCandidateCount(terms, idx, fallback, ia)
		cb := estimateSearchTermCandidateCount(terms, idx, fallback, ib)
		return shouldPrioritizeSearchTerm(terms, ia, ib, ca, cb)
	})
	return result
}

// buildIndexSearchPlan 构建名称搜索计划。 [S 汇编 0x1407e36e0, 992B] 实证：
//
//	mask = uint32(1)<<uint(n) - 1（n>=32 时下溢为全 1）；循环对每个术语：
//	flags[0]||flags[1] 为模式术语 → bytesMatchWildcardFold 命中则 matched|=1<<i；
//	flags[2]!=0 → others|=1<<i。随后 driver=selectDriverTermIndex(...)，driver>=0 时
//	driverBit=1<<driver；order=buildOrderedNameTermIndices(...)。
func buildIndexSearchPlan(terms []nameSearchTerm, b []byte, idx *nameFrequencyIndex, fallback int, caseSensitive bool) nameSearchPlan {
	n := len(terms)
	plan := nameSearchPlan{
		mask:   uint32(1)<<uint(n) - 1,
		driver: -1,
	}
	for i := range terms {
		t := terms[i]
		if t.flags[0] != 0 || t.flags[1] != 0 {
			if bytesMatchWildcardFold(b, t.patterns, caseSensitive) {
				plan.matched |= uint32(1) << uint(i)
			}
		}
		if t.flags[2] != 0 {
			plan.others |= uint32(1) << uint(i)
		}
	}
	plan.driver = selectDriverTermIndex(terms, idx, fallback)
	if plan.driver >= 0 {
		plan.driverBit = uint32(1) << uint(plan.driver)
	}
	plan.order = buildOrderedNameTermIndices(terms, idx, fallback, plan.driver)
	return plan
}

// matchNodeNameTerms 对节点名 b 执行名称术语匹配，返回更新后的 matched 掩码。
// [S 汇编 0x1407e3ac0, 928B] 实证，三段式：
//
//	(A) 快路径：driver>=0 且 matched&driverBit==0 时先匹配 driver 术语，失败直接返回；
//	(B) 按 order 顺序匹配各索引，跳过已匹配位，命中置位；
//	(C) 兜底：遍历全部术语，仅当 flags[0]!=0&&flags[1]==0&&flags[3]!=0 且未匹配时
//	    再试。三段都调用 bytesMatchWildcardFold(b, term.patterns, caseSensitive)。
func matchNodeNameTerms(b []byte, terms []nameSearchTerm, caseSensitive bool, plan nameSearchPlan) uint32 {
	matched := plan.matched
	if plan.driver >= 0 && matched&plan.driverBit == 0 {
		if !bytesMatchWildcardFold(b, terms[plan.driver].patterns, caseSensitive) {
			return matched
		}
		matched |= plan.driverBit
	}
	for _, idx := range plan.order {
		bit := uint32(1) << uint(idx)
		if matched&bit != 0 {
			continue
		}
		if bytesMatchWildcardFold(b, terms[idx].patterns, caseSensitive) {
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
		if bytesMatchWildcardFold(b, t.patterns, caseSensitive) {
			matched |= bit
		}
	}
	return matched
}

// [S 汇编 0x1407ecca0, 1096B] 实证：候选节点名匹配（非拼音，searchSignatureCandidatesWithTombstones 路径）。
// matched 先由 matchNodeNameTerms 计算；driver 未命中 → false；mask==matched → true；
// ^matched&others==0 → false；否则从 node.ParentIdx 沿父链上溯，逐个节点用
// matchHierarchyTermWithPinyin 补齐剩余层级术语，墓碑后代直接淘汰，最后返回 mask==matched。
func (v volumeIndexReadView) matchSearchCandidateNode(
	node IndexNode,
	name []byte,
	terms []nameSearchTerm,
	plan nameSearchPlan,
	tombstoneLookup *volumeIndexTombstoneLookup,
	caseSensitive bool,
) bool {
	matched := matchNodeNameTerms(name, terms, caseSensitive, plan)
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

// [S 汇编 0x1407f2060, 480B] 实证：按索引取节点。
// nodeBytes 非空时从序列化字节（每节点 24B）解析 IndexNode；否则直接取 v.nodes[i]。
func (v volumeIndexReadView) nodeAt(nodeIndex int) IndexNode {
	if len(v.nodeBytes) != 0 {
		off := nodeIndex * 24
		rec := v.nodeBytes[off : off+24]
		return IndexNode{
			FRN:        binary.LittleEndian.Uint64(rec[0:8]),
			NameOffset: binary.LittleEndian.Uint32(rec[8:12]),
			ParentIdx:  int32(binary.LittleEndian.Uint32(rec[12:16])),
			ModTime:    binary.LittleEndian.Uint32(rec[16:20]),
			NameLen:    binary.LittleEndian.Uint16(rec[20:22]),
			Flags:      binary.LittleEndian.Uint16(rec[22:24]),
		}
	}
	return v.nodes[nodeIndex]
}

// [S 汇编 0x1407f2240, 544B] 实证：带护栏取节点。
// nodeIndex>=0x40000000 走 deltaNodes 覆盖；越界/负值返回零值。
func (v volumeIndexReadView) nodeAtIndex(nodeIndex int32) IndexNode {
	if nodeIndex >= 0x40000000 {
		deltaIdx := int(nodeIndex - 0x40000000)
		if deltaIdx < len(v.deltaNodes) {
			return v.deltaNodes[deltaIdx].node
		}
		return IndexNode{}
	}
	if nodeIndex < 0 {
		return IndexNode{}
	}
	var count int64
	if len(v.nodeBytes) != 0 {
		count = int64(len(v.nodeBytes) / 24)
	} else {
		count = int64(len(v.nodes))
	}
	if count > int64(nodeIndex) {
		return v.nodeAt(int(nodeIndex))
	}
	return IndexNode{}
}

// [S 汇编 0x1407f1e40, 544B] 实证：取节点名。
// NameOffset==0xFFFFFFFF 表示 delta 节点：先查 deltaByFRN 再线性扫 deltaNodes；
// 否则从 namePool[NameOffset : NameOffset+NameLen] 切片。
func (v volumeIndexReadView) nodeName(node IndexNode) []byte {
	if node.NameOffset == 0xFFFFFFFF {
		if deltaIdx, ok := v.deltaByFRN[node.FRN]; ok && deltaIdx >= 0x40000000 {
			i := int(deltaIdx - 0x40000000)
			if i < len(v.deltaNodes) && v.deltaNodes[i].node.Flags&2 == 0 {
				return v.deltaNodes[i].name
			}
		}
		for i := len(v.deltaNodes) - 1; i >= 0; i-- {
			if v.deltaNodes[i].node.FRN == node.FRN && v.deltaNodes[i].node.Flags&2 == 0 {
				return v.deltaNodes[i].name
			}
		}
		return nil
	}
	off := int(node.NameOffset)
	end := off + int(node.NameLen)
	if len(v.namePool) < end {
		return nil
	}
	return v.namePool[off:end]
}

// [S 汇编 0x1407f1a60, 992B] 实证：沿父链判定是否为墓碑后代（迭代 + memo 记忆化）。
// delta 节点递归到 ParentIdx；普通节点先查 memo 缓存，未命中则沿 ParentIdx 链上溯，
// 命中 tombstones 或缓存的祖先即得出结果，并把路径上所有节点回填 memo。
func (t *volumeIndexTombstoneLookup) IsTombstonedDescendant(nodeIndex int32) bool {
	if t == nil || len(t.tombstones) == 0 {
		return false
	}
	if nodeIndex < 0 {
		return false
	}
	if nodeIndex >= 0x40000000 {
		deltaIdx := int(nodeIndex - 0x40000000)
		if deltaIdx < len(t.deltaNodes) {
			return t.IsTombstonedDescendant(t.deltaNodes[deltaIdx].node.ParentIdx)
		}
		return false
	}
	if v, ok := t.memo[nodeIndex]; ok {
		return v
	}
	path := make([]int32, 0, 8)
	result := false
	for cur := nodeIndex; cur >= 0; {
		if v, ok := t.memo[cur]; ok {
			result = v
			break
		}
		path = append(path, cur)
		if _, ok := t.tombstones[cur]; ok {
			result = true
			break
		}
		cur = t.nodeAtIndex(cur).ParentIdx
	}
	for _, idx := range path {
		t.memo[idx] = result
	}
	return result
}

// [S 0x1407e78a0] 目录标志位判定：Flags&1 != 0（叶子函数，movzx word [rax+0x16] & 1）。
func (n *IndexNode) IsDirectory() bool {
	return n.Flags&1 != 0
}

// [S 0x1407e78c0] 删除标志位判定：Flags&2 != 0（叶子函数，movzx word [rax+0x16] test 2 setne）。
func (n *IndexNode) IsDeleted() bool {
	return n.Flags&2 != 0
}

// Release 释放读租约（幂等：nil 或已释放则直接返回）。
// [S 0x1407e6d00] 单 receiver 无参无返回。asm：released(+0xf0) 已置位 → 返回；
// 否则置位 → release(+0xe8)!=nil → 调用回调。
func (l *volumeIndexReadLease) Release() {
	if l == nil || l.released {
		return
	}
	l.released = true
	if l.release != nil {
		l.release()
	}
}

// timeToUnixNano 将 time.Time 转为 Unix 纳秒。
// [S 0x1407f8640] time.Time 单参返回 int64。asm：time.Time.UnixNano 内联
// （wall 单调钟位 bit63 分派：wallToInternal + nsec 部分 + internalToUnix 偏移）。
func timeToUnixNano(t time.Time) int64 {
	return t.UnixNano()
}

// addNamePrefixBucketCount 累加名字 1-2 字节前缀桶计数（大小写折叠为小写）。
// [S 0x1407e2620] (c, name string, count int32)。asm：nil/空名返回 → 首字节折叠小写
// → firstByte[c0]+=count → len>=2 时 twoByte[(c0<<8)|c1]+=count。
func addNamePrefixBucketCount(c *namePrefixBucketCounts, name string, count int32) {
	if c == nil || len(name) == 0 {
		return
	}
	c0 := name[0]
	if c0 >= 'A' && c0 <= 'Z' {
		c0 += 'a' - 'A'
	}
	c.firstByte[c0] += count
	if len(name) >= 2 {
		c1 := name[1]
		if c1 >= 'A' && c1 <= 'Z' {
			c1 += 'a' - 'A'
		}
		c.twoByte[int(c0)<<8|int(c1)] += count
	}
}

// Close 关闭内存映射文件（逆序解映射 views、关闭映射句柄与底层文件，保留首个错误）。
// [S 0x1407fd880] 单 receiver 返回 error。asm：nil → nil；倒序遍历 views 逐个
// UnmapViewOfFile（err!=nil 且 firstErr==nil 才累积）→ views=nil → mapping!=0 则
// CloseHandle（同上累积）→ mapping=0 → file!=nil 则 file.Close()（同上累积）→ file=nil → 返回 firstErr。
func (m *volumeIndexMappedFile) Close() error {
	if m == nil {
		return nil
	}
	var firstErr error
	for i := len(m.views) - 1; i >= 0; i-- {
		if v := m.views[i]; v != 0 {
			if err := windows.UnmapViewOfFile(v); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	m.views = nil
	if m.mapping != 0 {
		if err := windows.CloseHandle(windows.Handle(m.mapping)); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	m.mapping = 0
	if m.file != nil {
		if err := m.file.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	m.file = nil
	return firstErr
}

// close 关闭三角索引（幂等：nil / 已关闭 / 无映射文件则返回）。
// [S 0x1407e1920] 单 receiver 无参无返回。asm：nil||closed(+0x40)||mappedFile(+0x38)==nil
// → 返回；否则 closed=true → 取 mappedFile → mappedFile=nil、signatureBytes=nil → mappedFile.Close()。
func (t *volumeNameTrigramIndex) close() {
	if t == nil || t.closed || t.mappedFile == nil {
		return
	}
	t.closed = true
	mf := t.mappedFile
	t.mappedFile = nil
	t.signatureBytes = nil
	mf.Close()
}

// truncateOpenWALAtValidOffset 在有效偏移截断打开的 WAL 文件。
// [S 0x1408008e0] (f *os.File, offset int64) error。asm：f==nil||offset<0 → errInvalidWALOffset；
// f.Truncate(offset) err!=nil → 返回；否则 f.Sync()。
func truncateOpenWALAtValidOffset(f *os.File, offset int64) error {
	if f == nil || offset < 0 {
		return errInvalidWALOffset
	}
	if err := f.Truncate(offset); err != nil {
		return err
	}
	return f.Sync()
}

// markPersistedAtLocked 标记索引已持久化（持锁）：t 为零值则取 now；清 dirty/changeCaught，记 LastSavedAt。
// [S 汇编 0x1407e9040, 224B]：sec()/nsec() 判零 → 零值则 time.Now；dirty(+0xe0)=false、
// changeCaught(+0xe4)=0、LastSavedAt(+0xc0/+0xc8/+0xd0)=t。
func (v *VolumeIndex) markPersistedAtLocked(t time.Time) {
	if t.IsZero() {
		t = time.Now()
	}
	v.dirty = false
	v.changeCaught = 0
	v.LastSavedAt = t
}

// markDirtyLocked 标记索引脏（持锁）：n==0 直接返回；置 dirty、累加 changeCaught、记 LastMutationAt、
// runtimeVersion 递增并跳过 0。
// [S 汇编 0x1408072e0, 192B]：ebx==0→ret；dirty(+0xe0)=true、changeCaught(+0xe4)+=n、
// LastMutationAt(+0xa8/+0xb0/+0xb8)=now、runtimeVersion(+0x590) 自增（旧值 -1 时覆盖为 1）。
func (v *VolumeIndex) markDirtyLocked(n int32) {
	if n == 0 {
		return
	}
	v.dirty = true
	v.changeCaught += uint32(n)
	v.LastMutationAt = time.Now()
	v.runtimeVersion++
	if v.runtimeVersion == 0 {
		v.runtimeVersion = 1
	}
}

// ClearDirty 加写锁标记索引已持久化。[S 汇编 0x1407e8e00, 160B]：mu.Lock + defer Unlock →
// markPersistedAtLocked(time.Now())。
func (v *VolumeIndex) ClearDirty() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.markPersistedAtLocked(time.Now())
}

// EntryCount 读锁下返回活动条目计数。[S 汇编 0x1407e7b00, 192B]：mu.RLock（lock xadd
// [rax+0x10] readerCount）+ defer RUnlock → activeEntryCountLocked() 透传返回 int。
func (v *VolumeIndex) EntryCount() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.activeEntryCountLocked()
}

// SearchWithPaths 按多路径搜索（薄包装：context 传 nil 调 searchWithPathsContextMetrics）。
// [S-sig 0x1407eaa80, 128B]：参数重排后 searchWithPathsContextMetrics(v, 0, 0, a, b, c, d, e)。
// 体待 searchWithPathsContextMetrics（8 寄存器参数 + 9 字返回）专项还原。
func (v *VolumeIndex) SearchWithPaths(a, b, c, d interface{}, e bool) interface{} {
	_, _, _, _, _ = a, b, c, d, e
	return nil
}

// activeEntryCountLocked 持锁下统计活动条目（读视图 + activeVolumeIndexEntryCountForView）。
// [S-sig 0x1407e8340, 448B]：readProvider(+0x520) 为空则用默认 provider 取读视图，
// defer 释放视图；打包 activeCount(+0xd8)/delta(+0x548..+0x588) 字段传给
// activeVolumeIndexEntryCountForView；返回 int。体待读视图域专项还原。
func (v *VolumeIndex) activeEntryCountLocked() int {
	return 0
}

// Release 释放 journal pending 状态（幂等）。
// [S-sig 0x140806820, 160B]：released(+0x128) 未置位则置位 + callback(+0x120) 非空调用；
// 随后清零 view 尾部(+0x110..0x130) 与 overlay(+0x130)。体待精确字段布局专项还原。
func (s *volumeIndexJournalPendingState) Release() {
	if !s.lease.released {
		s.lease.released = true
		if s.lease.release != nil {
			s.lease.release()
		}
	}
}
