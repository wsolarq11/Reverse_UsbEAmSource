package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// volumeIndexMmapStaticDefault mmap 静态 provider 全局默认开关（环境变量空时生效）。
var volumeIndexMmapStaticDefault uint32

// volumeIndexGetSystemInfoProc 是 kernel32.GetSystemInfo 的延迟加载入口（LazyProc 名
// "GetSystemInfo"@0x140c4ded5，13B，全局 @0x141bc1930 实证）。
var volumeIndexGetSystemInfoProc = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetSystemInfo")

// errInvalidWALOffset 无效 WAL 偏移哨兵错误（f==nil 或 offset<0）。
// 精确消息字符串待取证（.data 静态 error 值 @0x14193c650 为 typeOff 编码，未解码）。
var errInvalidWALOffset = errors.New("invalid WAL offset")

// errVolumeIndexNodeBytesInvalid 卷索引节点字节流损坏/格式不兼容的统一哨兵。
var errVolumeIndexNodeBytesInvalid = errors.New("索引文件损坏或版本不兼容")

// errVolumeIndexFileNotFound 卷索引文件不存在的哨兵（.data 静态 error @0x141bc3c60 实证）。
var errVolumeIndexFileNotFound = errors.New("索引文件不存在")

// errVolumeIndexBucketPathEmpty bucket 路径为空的哨兵（.data 静态 error @0x141bc3e10，消息 10B "empty path" 经 0x1407f7089 mov [rax+8],0xa 实证）。
var errVolumeIndexBucketPathEmpty = errors.New("empty path")

// newVolumeIndexPersistencePaths 从索引根路径派生 6 个持久化文件路径。
// [S 汇编 0x1407e1020, 576B] 实证：从末尾回扫找最后一个 `\`/`/`/`.`；
// 命中 `.` 且其索引 > 最后一个 `\`/`/` 索引时 base=root[:dot]（去扩展名），否则 base=root。
// CheckpointPath 恒为完整 root；其余 5 个 = base + ".meta"/".wal"/".tri"/".bkt"/".pyn"
// （后缀经 .rdata 常量 0x140c35c68/0x140c3480a/0x140c3480e/0x140c34812/0x140c34816 实证）。
func newVolumeIndexPersistencePaths(root string) volumeIndexPersistencePaths {
	base := root
	if dot := strings.LastIndexByte(root, '.'); dot > strings.LastIndexAny(root, `/\`) {
		base = root[:dot]
	}
	return volumeIndexPersistencePaths{
		CheckpointPath:  root,
		MetaPath:        base + ".meta",
		WALPath:         base + ".wal",
		NameTrigramPath: base + ".tri",
		BucketPath:      base + ".bkt",
		PinyinPath:      base + ".pyn",
	}
}

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
//  1. 双方候选计数都 >0 且不等 → 计数更小者优先（return countJ > countI）；
//  2. 单片段（len(patterns)==1）vs 多片段 → 单片段优先；
//  3. 全部片段字节总长不等 → 总长更短者优先；
//  4. 片段数不等 → 片段更少者优先；
//  5. 平局 → 索引更大者优先（return j > i）。
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

// collectSearchTermBigramKeys 收集单个术语 patterns 中的双字（bigram）键。
// [S 汇编 0x1407e27c0, 576B] 实证：flags[0]||flags[1] 非零（模式术语）→ 返回 nil；
// 遍历 patterns（[]byte 步长 0x18），len(p)==2 时取前两字节 ASCII 大写→小写
// 组合为 uint16 key（c0<<8|c1），map 去重后 append（结果初始 cap=2）。
func collectSearchTermBigramKeys(term nameSearchTerm) []uint16 {
	if term.flags[0] != 0 || term.flags[1] != 0 {
		return nil
	}
	seen := make(map[uint16]struct{})
	keys := make([]uint16, 0, 2)
	for _, p := range term.patterns {
		if len(p) != 2 {
			continue
		}
		c0 := p[0]
		if 'A' <= c0 && c0 <= 'Z' {
			c0 |= 0x20
		}
		c1 := p[1]
		if 'A' <= c1 && c1 <= 'Z' {
			c1 |= 0x20
		}
		key := uint16(c0)<<8 | uint16(c1)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	return keys
}

// selectSearchBigramKeys 从所有术语中选出最长的一组双字键。
// [S 汇编 0x1407e2a00, 384B] 实证：遍历 terms(+0x38 结构) → collectSearchTermBigramKeys；
// keys 空则跳过；result 为 nil 或 len(keys)>len(result) 时 result=keys。
func selectSearchBigramKeys(terms []nameSearchTerm) []uint16 {
	var result []uint16
	for _, term := range terms {
		keys := collectSearchTermBigramKeys(term)
		if len(keys) == 0 {
			continue
		}
		if result == nil || len(keys) > len(result) {
			result = keys
		}
	}
	return result
}

// collectSearchTermTrigramKeys 收集单个术语 patterns 中的三字（trigram）键。
// [S 汇编 0x1407e2ca0, 736B] 实证：flags[0]||flags[1] 非零（模式术语）→ 返回 nil；
// 遍历 patterns，len(p)>=3 时对每个 3 字节滑动窗口（i∈[0,len-3)）取三字节
// ASCII 大写→小写组合为 uint32 key（c0<<16|c1<<8|c2），map 去重后 append（cap=8）。
func collectSearchTermTrigramKeys(term nameSearchTerm) []uint32 {
	if term.flags[0] != 0 || term.flags[1] != 0 {
		return nil
	}
	seen := make(map[uint32]struct{})
	keys := make([]uint32, 0, 8)
	for _, p := range term.patterns {
		for i := 0; i+2 < len(p); i++ {
			c0 := p[i]
			if 'A' <= c0 && c0 <= 'Z' {
				c0 |= 0x20
			}
			c1 := p[i+1]
			if 'A' <= c1 && c1 <= 'Z' {
				c1 |= 0x20
			}
			c2 := p[i+2]
			if 'A' <= c2 && c2 <= 'Z' {
				c2 |= 0x20
			}
			key := uint32(c0)<<16 | uint32(c1)<<8 | uint32(c2)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			keys = append(keys, key)
		}
	}
	return keys
}

// selectSearchTrigramKeys 从所有术语中选出最长的一组三字键。
// [S 汇编 0x1407e2f80, 384B] 实证：遍历 terms(+0x38 结构) → collectSearchTermTrigramKeys；
// keys 空则跳过；result 为 nil 或 len(keys)>len(result) 时 result=keys。
func selectSearchTrigramKeys(terms []nameSearchTerm) []uint32 {
	var result []uint32
	for _, term := range terms {
		keys := collectSearchTermTrigramKeys(term)
		if len(keys) == 0 {
			continue
		}
		if result == nil || len(keys) > len(result) {
			result = keys
		}
	}
	return result
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

// validateVolumeIndexNodeBytes 校验节点字节流与名字池大小的一致性。
// [S 汇编 0x1407e2420, 512B] 实证：len%24!=0 或 namePoolSize<0 → errVolumeIndexNodeBytesInvalid；
// 逐节点（24B）解析并校验 ParentIdx∈[-1,nodeCount) 与 NameOffset+NameLen<=namePoolSize，
// 任一越界 → errVolumeIndexNodeBytesInvalid；全部通过 → nil。
func validateVolumeIndexNodeBytes(nodeBytes []byte, namePoolSize int) error {
	if len(nodeBytes)%24 != 0 || namePoolSize < 0 {
		return errVolumeIndexNodeBytesInvalid
	}
	nodeCount := len(nodeBytes) / 24
	for i := 0; i < nodeCount; i++ {
		rec := nodeBytes[i*24 : i*24+24]
		parentIdx := int32(binary.LittleEndian.Uint32(rec[12:16]))
		if parentIdx < -1 || parentIdx >= int32(nodeCount) {
			return errVolumeIndexNodeBytesInvalid
		}
		nameOffset := int(binary.LittleEndian.Uint32(rec[8:12]))
		nameLen := int(binary.LittleEndian.Uint16(rec[20:22]))
		if nameOffset+nameLen > namePoolSize {
			return errVolumeIndexNodeBytesInvalid
		}
	}
	return nil
}

// decodeVolumeIndexNodesFromBytesWithNodeCount 从序列化节点字节流解析 []IndexNode。
// [S 汇编 0x1407fcc00, 768B] 实证：签名 (nodeBytes []byte, nodeCount int)
// ([]IndexNode, int, error)。len%24!=0（@0x1407fcc39 魔数除 24 后 cmp rbx,r9）→ 哨兵；
// nodeCount < len/24（@0x1407fcc40 cmp rdi,rdx jl）→ 哨兵；否则 makeslice(len/24)。
// 逐 24B 解析（FRN@0 8B、NameOffset@8 4B、ParentIdx@12 4B、ModTime@16 4B、
// NameLen@20 2B、Flags@22 2B，小端）；ParentIdx<-1 或 ParentIdx>=nodeCount
// （@0x1407fcde2..e18，cmp esi,r9d jg）→ 哨兵；Flags&2==0（@0x1407fccf3 test r10w,2）
// 计入 active 计数；返回 (nodes, active, nil)。
func decodeVolumeIndexNodesFromBytesWithNodeCount(nodeBytes []byte, nodeCount int) ([]IndexNode, int, error) {
	if len(nodeBytes)%24 != 0 {
		return nil, 0, errVolumeIndexNodeBytesInvalid
	}
	derived := len(nodeBytes) / 24
	if nodeCount < derived {
		return nil, 0, errVolumeIndexNodeBytesInvalid
	}
	nodes := make([]IndexNode, derived, derived)
	active := 0
	for i := 0; i < derived; i++ {
		rec := nodeBytes[i*24 : i*24+24]
		node := IndexNode{
			FRN:        binary.LittleEndian.Uint64(rec[0:8]),
			NameOffset: binary.LittleEndian.Uint32(rec[8:12]),
			ParentIdx:  int32(binary.LittleEndian.Uint32(rec[12:16])),
			ModTime:    binary.LittleEndian.Uint32(rec[16:20]),
			NameLen:    binary.LittleEndian.Uint16(rec[20:22]),
			Flags:      binary.LittleEndian.Uint16(rec[22:24]),
		}
		if node.ParentIdx < -1 || node.ParentIdx >= int32(nodeCount) {
			return nil, 0, errVolumeIndexNodeBytesInvalid
		}
		nodes[i] = node
		if node.Flags&2 == 0 {
			active++
		}
	}
	return nodes, active, nil
}

// sortedIndexAt 按排序位置取节点索引（sortedIndices 优先，缺省回退 sortedBytes）。
// [S 汇编 0x1407f2460, 544B] 实证：index<0→(-1,false)；sortedIndices 非空→越界返回
// (-1,false)，否则读 sortedIndices[index]；空→sortedBytes 空或 len%4!=0 或越界→(-1,false)，
// 否则按小端读 sortedBytes[index*4:] 的 int32。nodeIndex<0→(nodeIndex,false)；
// 否则返回 (nodeIndex, nodeCount>nodeIndex)（nodeCount=nodeBytes/24 或 len(nodes)）。
func (v volumeIndexReadView) sortedIndexAt(index int) (int32, bool) {
	if index < 0 {
		return -1, false
	}
	var nodeIndex int32
	if len(v.sortedIndices) != 0 {
		if index >= len(v.sortedIndices) {
			return -1, false
		}
		nodeIndex = v.sortedIndices[index]
	} else {
		if len(v.sortedBytes) == 0 || len(v.sortedBytes)%4 != 0 {
			return -1, false
		}
		if index >= len(v.sortedBytes)/4 {
			return -1, false
		}
		nodeIndex = int32(binary.LittleEndian.Uint32(v.sortedBytes[index*4:]))
	}
	if nodeIndex < 0 {
		return nodeIndex, false
	}
	var count int64
	if len(v.nodeBytes) != 0 {
		count = int64(len(v.nodeBytes) / 24)
	} else {
		count = int64(len(v.nodes))
	}
	return nodeIndex, count > int64(nodeIndex)
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
// [S 0x1407e2620, 160B] (c, name string, count int32)。asm：c==nil（@0x1407e262d
// test rax,rax je）或 len(name)==0（@0x1407e2632 test rcx,rcx je）返回 → 首字节
// 折叠小写（@0x1407e2640 cmp dil,0x19 ja / or edx,0x20）→ firstByte[c0]+=count
// （@0x1407e264c add [rax+rdi*4],esi）→ len(name)<2（@0x1407e264f cmp rcx,2 jl）
// 或 twoByte.len==0（@0x1407e2655 mov rcx,[rax+0x408] / @0x1407e2660 test rcx,rcx
// jne）返回 → 否则 twoByte[(c0<<8)|c1]+=count（@0x1407e2695 add [rax+rdx*4],esi）。
func addNamePrefixBucketCount(c *namePrefixBucketCounts, name string, count int32) {
	if c == nil || len(name) == 0 {
		return
	}
	c0 := name[0]
	if c0 >= 'A' && c0 <= 'Z' {
		c0 += 'a' - 'A'
	}
	c.firstByte[c0] += count
	if len(name) < 2 || len(c.twoByte) == 0 {
		return
	}
	c1 := name[1]
	if c1 >= 'A' && c1 <= 'Z' {
		c1 += 'a' - 'A'
	}
	c.twoByte[int(c0)<<8|int(c1)] += count
}

// buildNamePrefixBucketCounts 从已解析节点切片构建名字前缀桶计数。
// [S 汇编 0x1407e1c80, 608B] 实证：签名 (nodes []IndexNode, namePool []byte)
// namePrefixBucketCounts。len(nodes)>0x2710（@0x1407e1cf1 cmp rbx,0x2710 jle）→
// 预分配 twoByte=make([]int32,0x10000)（@0x1407e1d09 makeslice）；遍历 nodes
// （@0x1407e1d55 cmp rbx,rax jle），Flags&2（@0x1407e1d82 test r9w,2 jne）跳过
// deleted；否则 namePool[NameOffset:NameOffset+NameLen]（@0x1407e1da3 jb
// panicSliceAcap，cap 界）→ addNamePrefixBucketCount(&counts, name, 1)。
func buildNamePrefixBucketCounts(nodes []IndexNode, namePool []byte) namePrefixBucketCounts {
	var counts namePrefixBucketCounts
	if len(nodes) > 0x2710 {
		counts.twoByte = make([]int32, 0x10000, 0x10000)
	}
	for i := range nodes {
		n := nodes[i]
		if n.Flags&2 != 0 {
			continue
		}
		off := int(n.NameOffset)
		name := namePool[off : off+int(n.NameLen)]
		addNamePrefixBucketCount(&counts, string(name), 1)
	}
	return counts
}

// buildNamePrefixBucketCountsFromNodeBytes 从序列化节点字节流构建名字前缀桶计数。
// [S 汇编 0x1407e1ee0, 1344B] 实证：签名 (nodeBytes []byte, namePool []byte)
// (namePrefixBucketCounts, int, error)。len%24!=0（@0x1407e1f6a 魔数除 24 后
// cmp rbx,r12 jne）→ (零,0,errVolumeIndexNodeBytesInvalid)；derived=len/24；
// derived>0x2710（@0x1407e1fa2）→ 预分配 twoByte；逐 24B 解析（FRN@0/NameOffset@8/
// ParentIdx@12/ModTime@16/NameLen@20/Flags@22，小端）；ParentIdx<-1（@0x1407e2146
// cmp r8d,-1）或 ParentIdx>=derived（@0x1407e2165 cmp r8d,r11d）→ 错误；Flags&2
// （@0x1407e218f test r8w,2）跳过；namePool 越界（@0x1407e21b1 cmp r8,rcx jb）→
// 错误；addNamePrefixBucketCount(&counts, name, 1) 累加；返回 (counts, active, nil)。
func buildNamePrefixBucketCountsFromNodeBytes(nodeBytes []byte, namePool []byte) (namePrefixBucketCounts, int, error) {
	var counts namePrefixBucketCounts
	if len(nodeBytes)%24 != 0 {
		return counts, 0, errVolumeIndexNodeBytesInvalid
	}
	derived := len(nodeBytes) / 24
	if derived > 0x2710 {
		counts.twoByte = make([]int32, 0x10000, 0x10000)
	}
	active := 0
	for off := 0; off < len(nodeBytes); off += 24 {
		rec := nodeBytes[off : off+24]
		node := IndexNode{
			FRN:        binary.LittleEndian.Uint64(rec[0:8]),
			NameOffset: binary.LittleEndian.Uint32(rec[8:12]),
			ParentIdx:  int32(binary.LittleEndian.Uint32(rec[12:16])),
			ModTime:    binary.LittleEndian.Uint32(rec[16:20]),
			NameLen:    binary.LittleEndian.Uint16(rec[20:22]),
			Flags:      binary.LittleEndian.Uint16(rec[22:24]),
		}
		if node.ParentIdx < -1 || node.ParentIdx >= int32(derived) {
			return counts, 0, errVolumeIndexNodeBytesInvalid
		}
		if node.Flags&2 != 0 {
			continue
		}
		nameOff := int(node.NameOffset)
		nameLen := int(node.NameLen)
		if nameOff+nameLen > len(namePool) {
			return counts, 0, errVolumeIndexNodeBytesInvalid
		}
		name := namePool[nameOff : nameOff+nameLen]
		addNamePrefixBucketCount(&counts, string(name), 1)
		active++
	}
	return counts, active, nil
}

// newVolumeIndexReadViewFromCheckpointView 从 checkpoint 视图构建读视图。
// [S 汇编 0x1407e5260, 1120B] 实证：签名 (cv *volumeIndexCheckpointView)
// volumeIndexReadView。cv==nil（@0x1407e52a0 test rax,rax je）→ 零值。
// prefixBuckets!=nil（@0x1407e52a9 cmp [rax+0x120],0 je）→ 直接路径：namePool 空则
// 回退 namePoolBytes（@0x1407e52cc test rcx,rcx），nodes 非空则 nodeBytes 清零
// （@0x1407e530a cmovne）。prefixBuckets==nil → new(namePrefixBucketCounts)
// （@0x1407e546d newobject）→ *pb = buildNamePrefixBucketCounts(nodes, namePool)
// （@0x1407e5500 call），此路径 namePool 不回退（@0x1407e5649 直拷 cv.namePool）。
// nameTrigram/nameBigram/pinyin/tombstones/delta 等字段保持零值（duffzero 清零）。
func newVolumeIndexReadViewFromCheckpointView(cv *volumeIndexCheckpointView) volumeIndexReadView {
	if cv == nil {
		return volumeIndexReadView{}
	}
	if cv.prefixBuckets != nil {
		namePool := cv.namePool
		if len(namePool) == 0 {
			namePool = cv.namePoolBytes
		}
		nodeBytes := cv.nodeBytes
		if len(cv.nodes) != 0 {
			nodeBytes = nil
		}
		return volumeIndexReadView{
			root:          cv.root,
			nodes:         cv.nodes,
			nodeBytes:     nodeBytes,
			sortedIndices: cv.sortedIndices,
			sortedBytes:   cv.sortedBytes,
			namePool:      namePool,
			activeCount:   cv.activeCount,
			prefixBuckets: cv.prefixBuckets,
		}
	}
	pb := new(namePrefixBucketCounts)
	*pb = buildNamePrefixBucketCounts(cv.nodes, cv.namePool)
	nodeBytes := cv.nodeBytes
	if len(cv.nodes) != 0 {
		nodeBytes = nil
	}
	return volumeIndexReadView{
		root:          cv.root,
		nodes:         cv.nodes,
		nodeBytes:     nodeBytes,
		sortedIndices: cv.sortedIndices,
		sortedBytes:   cv.sortedBytes,
		namePool:      cv.namePool,
		activeCount:   cv.activeCount,
		prefixBuckets: pb,
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

// MarkPersistedAt 标记索引已持久化（加写锁转发 markPersistedAtLocked）。
// [S 汇编 0x1407e8f00, 224B]：mu.Lock + defer Unlock → markPersistedAtLocked(t)。
func (v *VolumeIndex) MarkPersistedAt(t time.Time) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.markPersistedAtLocked(t)
}

// IsDirty 读锁下返回 (脏标志, 变更计数)。
// [S 汇编 0x1407e8840, 224B]：mu.RLock + defer RUnlock → 读 dirty(+0xe0)/changeCaught(+0xe4)。
func (v *VolumeIndex) IsDirty() (bool, uint32) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.dirty, v.changeCaught
}

// usesMappedReadProvider 判定读 provider 是否为映射实现。
// [S 汇编 0x1407e7660, 224B]：nil 返回 false；mu.RLock + defer RUnlock →
// 比较 readProvider(+0x520) 具体类型 itab 是否为映射 provider。
func (v *VolumeIndex) usesMappedReadProvider() bool {
	if v == nil {
		return false
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.readProvider != nil && reflect.TypeOf(v.readProvider) == reflect.TypeOf((*volumeIndexMappedReadProvider)(nil))
}

// runtimeVersionSnapshot 读锁下返回运行时版本号。
// [S 汇编 0x1407e92a0, 224B]：nil 返回 0；mu.RLock + defer RUnlock → 读 runtimeVersion(+0x590)。
func (v *VolumeIndex) runtimeVersionSnapshot() uint64 {
	if v == nil {
		return 0
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.runtimeVersion
}

// ApplyDelete 应用 journal 删除条目到 overlay 映射。
// [S-sig 0x140806b60, 224B]：entry 解析 + mapassign_fast64 写 overlay(+0x210)。
// 体待 journal entry 专项还原。
func (p *volumeIndexJournalPendingState) ApplyDelete(a, b interface{}, c bool) {
	_, _, _ = p, a, b
	_ = c
}

// ApplyJournalChanges 应用 journal 变更（薄包装：默认参数转发 ApplyJournalChangesDetailedResult）。
// [S-sig 0x140804b80, 160B]：参数重排后 ApplyJournalChangesDetailedResult。
func (v *VolumeIndex) ApplyJournalChanges(a, b interface{}, c, d bool) interface{} {
	_, _, _, _ = a, b, c, d
	return nil
}

// SearchWithPathsContext 按多路径搜索（薄包装转发 searchWithPathsContextMetrics）。
// [S-sig 0x140808160, 160B]：参数重排后 searchWithPathsContextMetrics（8 参数 + 9 字返回）。
func (v *VolumeIndex) SearchWithPathsContext(a, b, c, d interface{}, e bool) interface{} {
	_, _, _, _, _ = a, b, c, d, e
	return nil
}

// shouldUseVolumeIndexMmapStaticProvider 判定是否启用 mmap 静态 provider（环境变量）。
// [S 汇编 0x1407e77a0, 256B]：Getenv(39 字符名) → TrimSpace → ToLower → 按长度分支
// "0"/"no"/"off"/"false"/"disable"/"disabled" → false；空 → 全局默认 !=0；其他 → true。
func shouldUseVolumeIndexMmapStaticProvider() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("USBEAM_FILE_INDEX_MMAP_STATIC")))
	switch v {
	case "":
		return volumeIndexMmapStaticDefault != 0
	case "0", "no", "off", "false", "disable", "disabled":
		return false
	default:
		return true
	}
}

// shouldUseVolumeIndexMmapLoad 判定是否启用 mmap load（环境变量）。
// [S 汇编 0x1407fa0e0, 256B]：Getenv(28 字符名) → TrimSpace → ToLower →
// "0"/"no"/"off"/"false"/"disable"/"disabled" → false；其他（含空）→ true。
func shouldUseVolumeIndexMmapLoad() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("USBEAM_FILE_INDEX_MMAP_LOAD")))
	switch v {
	case "0", "no", "off", "false", "disable", "disabled":
		return false
	default:
		return true
	}
}

// planVolumeIndexMappedSection 规划索引映射段（页对齐算术）。
// [S 汇编 0x1407fcf00, 224B] 实证：签名 (offset, size, granularity, limit int64)
// 返回 (alignedOffset, alignedSize, inPageOffset, sectionSize int64, error)。
// offset<0 || size<0 || granularity<=0 || limit<0（@0x1407fcf00..f12）→ 哨兵错误；
// size==0（@0x1407fcf2e je）→ (align(offset), 0, offset-align, 0, nil)；
// size!=0：offset+size 溢出或 limit<offset+size（@0x1407fcf37/40）→ 哨兵错误；
// alignedOffset = floor(offset/granularity)*granularity（@0x1407fcf67 idiv）；
// inPageOffset = offset - alignedOffset（@0x1407fcf6e）；
// alignedSize = size + inPageOffset（@0x1407fcf71）；inPageOffset>alignedSize 溢出 → 哨兵；
// 否则返回 (alignedOffset, alignedSize, inPageOffset, size, nil)。
func planVolumeIndexMappedSection(offset, size, granularity, limit int64) (int64, int64, int64, int64, error) {
	if offset < 0 || size < 0 || granularity <= 0 || limit < 0 {
		return 0, 0, 0, 0, errVolumeIndexNodeBytesInvalid
	}
	if size == 0 {
		alignedOffset := offset / granularity * granularity
		inPageOffset := offset - alignedOffset
		return alignedOffset, 0, inPageOffset, 0, nil
	}
	if offset+size < offset || limit < offset+size {
		return 0, 0, 0, 0, errVolumeIndexNodeBytesInvalid
	}
	alignedOffset := offset / granularity * granularity
	inPageOffset := offset - alignedOffset
	alignedSize := size + inPageOffset
	if inPageOffset > alignedSize {
		return 0, 0, 0, 0, errVolumeIndexNodeBytesInvalid
	}
	return alignedOffset, alignedSize, inPageOffset, size, nil
}

// planVolumeIndexCheckpointMap 规划 checkpoint 三个段（nodes/sorted/names）的映射视图。
// [S 汇编 0x1407fcfe0, 736B] 实证：签名 (layout volumeIndexCheckpointLayout, granularity int64)
// 返回 ([3]volumeIndexMappedSectionPlan, error)。逐段调用 planVolumeIndexMappedSection：
// nodes=(NodeOffset,NodeBytes)、sorted=(SortedOffset,SortedBytes)、names=(NameOffset,NameBytes)，
// 三者 limit 均取 layout.FileSize（@0x1407fd066/0a3/0ed rdi=[rsp+0xd0]）、granularity 取 r11；
// 任一段 err（@0x1407fd076/0c0/105 test rsi）→ 返回 (零, err)；否则组装三段
// (MapOffset,MapBytes,SectionOffset,SectionBytes)（@0x1407fd13f..1b9 栈写）。
func planVolumeIndexCheckpointMap(layout volumeIndexCheckpointLayout, granularity int64) ([3]volumeIndexMappedSectionPlan, error) {
	var out [3]volumeIndexMappedSectionPlan
	alignedOffset, alignedSize, inPageOffset, sectionBytes, err := planVolumeIndexMappedSection(layout.NodeOffset, layout.NodeBytes, granularity, layout.FileSize)
	if err != nil {
		return out, err
	}
	out[0] = volumeIndexMappedSectionPlan{MapOffset: alignedOffset, MapBytes: alignedSize, SectionOffset: inPageOffset, SectionBytes: sectionBytes}
	alignedOffset, alignedSize, inPageOffset, sectionBytes, err = planVolumeIndexMappedSection(layout.SortedOffset, layout.SortedBytes, granularity, layout.FileSize)
	if err != nil {
		return out, err
	}
	out[1] = volumeIndexMappedSectionPlan{MapOffset: alignedOffset, MapBytes: alignedSize, SectionOffset: inPageOffset, SectionBytes: sectionBytes}
	alignedOffset, alignedSize, inPageOffset, sectionBytes, err = planVolumeIndexMappedSection(layout.NameOffset, layout.NameBytes, granularity, layout.FileSize)
	if err != nil {
		return out, err
	}
	out[2] = volumeIndexMappedSectionPlan{MapOffset: alignedOffset, MapBytes: alignedSize, SectionOffset: inPageOffset, SectionBytes: sectionBytes}
	return out, nil
}

// planVolumeIndexCheckpointMapWithSystemGranularity 用系统分配粒度规划 checkpoint 映射视图。
// [S 汇编 0x1407fd2c0, 576B] 实证：签名 (layout volumeIndexCheckpointLayout)
// 返回 ([3]volumeIndexMappedSectionPlan, error)。newobject{volumeIndexSystemInfo}
// （@0x1407fd33d）→ GetSystemInfo LazyProc.Call(1 arg=&info)（@0x1407fd385，全局 @0x141bc1930
// Name="GetSystemInfo"）；info.AllocationGranularity(+0x28)<=0（@0x1407fd396 jg 失败）→
// errVolumeIndexNodeBytesInvalid；否则 planVolumeIndexCheckpointMap(layout, granularity)
// （@0x1407fd411，r11=granularity）。
func planVolumeIndexCheckpointMapWithSystemGranularity(layout volumeIndexCheckpointLayout) ([3]volumeIndexMappedSectionPlan, error) {
	var info volumeIndexSystemInfo
	volumeIndexGetSystemInfoProc.Call(uintptr(unsafe.Pointer(&info)))
	if info.AllocationGranularity <= 0 {
		return [3]volumeIndexMappedSectionPlan{}, errVolumeIndexNodeBytesInvalid
	}
	return planVolumeIndexCheckpointMap(layout, int64(info.AllocationGranularity))
}

// rebuildVolumeIndexSortedIndices 重建排序索引切片。
// [S-sig 0x140807dc0, 256B]：makeslice([]int32,n) → 递增填充 → sort.Slice（比较闭包）。
// 体待排序比较域专项还原。
func rebuildVolumeIndexSortedIndices(n int, a, b interface{}) []int32 {
	_, _, _ = n, a, b
	return nil
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

// ApplyJournalChangesDetailed 应用日志变更（详细结果转发）。
// [S-sig 0x140804c20, 192B]：多参数转发 ApplyJournalChangesDetailedResult。
func (s *VolumeIndex) ApplyJournalChangesDetailed(a, b, c, d, e, f interface{}) interface{} {
	_, _, _, _, _, _ = a, b, c, d, e, f
	return nil
}

// releaseLease 释放映射读取租约（锁内 active--，归零且 closing 时调用 release）。
// [S 汇编 0x1407e4440, 288B]：nil 早退；Mutex lock；active>0→active--；active==0 且
// closing&&!closed→取 release 置 nil + closed=true；unlock；release 非空则调用。
func (p *volumeIndexMappedReadProvider) releaseLease() {
	if p == nil {
		return
	}
	p.mu.Lock()
	if p.active > 0 {
		p.active--
	}
	var rel func()
	if p.active == 0 && p.closing && !p.closed {
		rel = p.release
		p.release = nil
		p.closed = true
	}
	p.mu.Unlock()
	if rel != nil {
		rel()
	}
}

// Acquire 获取映射读视图租约（持锁计数；已关闭返回无 release 的静态视图）。
// [S 汇编 0x1407e4100, 512B] 实证：nil→零 lease；Mutex lock；closed(+0x111)→锁内复制 view
// 后 unlock 返回 lease{view}（release 保持 nil）；否则 active(+0x108)++ → 锁内取
// release(+0x100) → unlock → 复制 view → lease{view,release}。
func (p *volumeIndexMappedReadProvider) Acquire() volumeIndexReadLease {
	if p == nil {
		return volumeIndexReadLease{}
	}
	p.mu.Lock()
	if p.closed {
		view := p.view
		p.mu.Unlock()
		return volumeIndexReadLease{view: view}
	}
	p.active++
	rel := p.release
	p.mu.Unlock()
	return volumeIndexReadLease{view: p.view, release: rel}
}

// CanReloadStaticMmapCheckpoint 判定能否重载静态 mmap 检查点。
// [S-sig 0x1407e8980, 288B]：RLock → defer RUnlock；读字段（+0xe0）；overlayStatsLocked
// 结果判空。体待 overlayStatsLocked 落地。
func (s *VolumeIndex) CanReloadStaticMmapCheckpoint() bool {
	_ = s
	return false
}

// buildVolumeIndexCheckpointHeaderLocked 构建 checkpoint 64B 头（magic "UIDX"/版本 0x20003/字段拷贝）。
// [S 汇编 0x1407f4e80, 288B] 实证：签名 (v *VolumeIndex, entryCount uint32,
// sortedCount uint32, namePoolSize uint32) [64]byte。
// 二进制布局：magic "UIDX" @+0（copy 4B，@0x1407f4eee）、version 0x20003 @+4
// （@0x1407f4efd lea rcx/mov dword）、entryCount @+8（@0x1407f4f14）、
// sortedCount @+0xc（@0x1407f4f1c）、rootFRN=v.RootFRN(+0x30) @+0x10
// （@0x1407f4f2c）、journalID=v.JournalID(+0x80) @+0x18（@0x1407f4f34）、
// lastUSN=v.LastUSN(+0x88) @+0x20（@0x1407f4f3c）、generatedAtUnix=
// v.GeneratedAt.Unix() @+0x28（@0x1407f4f55 timeToUnixNano 内联拆解）、
// namePoolSize @+0x30（@0x1407f4f6f）、保留 @+0x38 为零。
func buildVolumeIndexCheckpointHeaderLocked(v *VolumeIndex, entryCount, sortedCount uint32, namePoolSize uint32) [64]byte {
	var h [64]byte
	copy(h[0:4], volumeIndexCheckpointHeaderMagic)
	binary.LittleEndian.PutUint32(h[4:8], volumeIndexFormatVersion)
	binary.LittleEndian.PutUint32(h[8:12], entryCount)
	binary.LittleEndian.PutUint32(h[12:16], sortedCount)
	binary.LittleEndian.PutUint64(h[16:24], v.RootFRN)
	binary.LittleEndian.PutUint64(h[24:32], v.JournalID)
	binary.LittleEndian.PutUint64(h[32:40], uint64(v.LastUSN))
	binary.LittleEndian.PutUint64(h[40:48], uint64(v.GeneratedAt.Unix()))
	binary.LittleEndian.PutUint64(h[48:56], uint64(namePoolSize))
	// h[56:64] 保留为零
	return h
}

// CheckpointSections 返回映射读提供器的 checkpoint 段（节点/排序/名字池 + ok）。
// [S-eq 0x1407e4560, 928B]：无独立 asm dump，签名与语义由调用方实证交叉验证
// （saveVolumeIndexCheckpointLockedContext @0x1407f4bb5 call → 9 寄存器返回
// nodes.ptr/len/cap + sorted.ptr/len/cap + names.ptr/len/cap + stack bool ok）。
// 语义：持锁读取 p.view 的 nodeBytes/sortedBytes/namePool 原始字节，有效则返 (sec, true)。
func (p *volumeIndexMappedReadProvider) CheckpointSections() (volumeIndexCheckpointSections, bool) {
	if p == nil {
		return volumeIndexCheckpointSections{}, false
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.view.nodeBytes == nil {
		return volumeIndexCheckpointSections{}, false
	}
	return volumeIndexCheckpointSections{
		nodes:  p.view.nodeBytes,
		sorted: p.view.sortedBytes,
		names:  p.view.namePool,
	}, true
}

// saveVolumeIndexCheckpointLockedContext 持锁保存 checkpoint（两条路径：sections 直写 / heap 快照写入）。
// [S 汇编 0x1407f4ae0, 928B] 实证：contextErr→readProvider 非 Mapped→堆路径直落；
// readProvider 为 Mapped→再查 deltaByFRN(len>0)/deltaReplaced(≠nil)/Nodes(len>0)/
// SortedIndices(len>0)/NamePool(len>0)→任一项命中即堆路径，否则 CheckpointSections→
// !ok 或 !IsSourcePath→堆路径，否则 sections 直写路径。
func saveVolumeIndexCheckpointLockedContext(ctx context.Context, v *VolumeIndex, path string) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	// 若 readProvider 非 *volumeIndexMappedReadProvider（或无），直接走堆路径。
	rp, isMapped := v.readProvider.(*volumeIndexMappedReadProvider)
	if !isMapped || rp == nil {
		nodes, namePool, sortedIndices, activeCount, err := buildVolumeIndexCheckpointSnapshotLocked(v)
		if err != nil {
			return err
		}
		return writeVolumeIndexCheckpointHeapContext(ctx, v, path, nodes, namePool, sortedIndices, activeCount)
	}
	// readProvider 为 Mapped：有 delta 或内存数据则强制堆路径。
	if len(v.deltaByFRN) > 0 || v.deltaReplaced != nil ||
		len(v.Nodes) > 0 || len(v.SortedIndices) > 0 || len(v.NamePool) > 0 {
		nodes, namePool, sortedIndices, activeCount, err := buildVolumeIndexCheckpointSnapshotLocked(v)
		if err != nil {
			return err
		}
		return writeVolumeIndexCheckpointHeapContext(ctx, v, path, nodes, namePool, sortedIndices, activeCount)
	}
	// 尝试 sections 路径：要求 CheckpointSections 成功且路径匹配源文件。
	sections, ok := rp.CheckpointSections()
	if !ok || !rp.IsSourcePath(path) {
		nodes, namePool, sortedIndices, activeCount, err := buildVolumeIndexCheckpointSnapshotLocked(v)
		if err != nil {
			return err
		}
		return writeVolumeIndexCheckpointHeapContext(ctx, v, path, nodes, namePool, sortedIndices, activeCount)
	}
	// sections 直写路径。
	nodeCount := uint32(len(sections.nodes) / 24)
	sortedCount := uint32(len(sections.sorted) / 4)
	namePoolSize := uint32(len(sections.names))
	header := buildVolumeIndexCheckpointHeaderLocked(v, nodeCount, sortedCount, namePoolSize)
	return writeVolumeIndexCheckpointSectionsContext(ctx, path, header, sections.nodes, sections.sorted, sections.names)
}

// writeVolumeIndexCheckpointHeapContext 堆快照持久化 checkpoint（头+节点+排序+名字池+字段更新）。
// [S 汇编 0x1407f4fa0, 2112B] 实证：签名 (ctx context.Context, v *VolumeIndex, path string,
// nodes []IndexNode, namePool []byte, sortedIndices []int32, activeCount int) error。
// 流程：(1) contextErr 守卫；(2) 遍历 nodes 校验每个 node.ParentIdx∈[-1, len(nodes))；
// (3) MkdirAll→CreateTemp→buildVolumeIndexCheckpointHeaderLocked(v, len(nodes),
// len(sortedIndices), len(namePool))→writeAllContext(header[0x40])→
// writeVolumeIndexNodesContext(nodes)→writeVolumeIndexSortedIndicesContext(sortedIndices)→
// contextErr 重检→writeAllContext(namePool)→tmp.Close→
// closeMappedReadProviderForReplaceLocked→replaceFileAtomically；
// (4) 字段更新：v.Nodes=nodes、v.SortedIndices=sortedIndices、v.NamePool=namePool、
// v.activeCount=activeCount、delta 全清零→buildNamePrefixBucketCounts→v.prefixBuckets→
// clearNameTrigramIndexLocked。
// 5 路错误清理：任一步失败则 os.Remove(tmpName)+（若 tmp 仍打开则 Close）后返回错误。
func writeVolumeIndexCheckpointHeapContext(ctx context.Context, v *VolumeIndex, path string, nodes []IndexNode, namePool []byte, sortedIndices []int32, activeCount int) error {
	// (1) contextErr 守卫（@0x1407f5005 contextErr + @0x1407f500d test jne）。
	if err := contextErr(ctx); err != nil {
		return err
	}
	// (2) 遍历 nodes 校验每个 node.ParentIdx∈[-1, len(nodes))（@0x1407f502c..5083）。
	nodeCount := len(nodes)
	for i := 0; i < nodeCount; i++ {
		p := int(nodes[i].ParentIdx)
		if p < -1 || p >= nodeCount {
			return errVolumeIndexNodeBytesInvalid
		}
	}
	// (3) 写入管线：MkdirAll → CreateTemp → header → nodes → sortedIndices → namePool → Close。
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".data-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// header（@0x1407f5147 buildVolumeIndexCheckpointHeaderLocked）。
	header := buildVolumeIndexCheckpointHeaderLocked(v, uint32(nodeCount), uint32(len(sortedIndices)), uint32(len(namePool)))
	if err := writeAllContext(ctx, tmp, header[:]); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	// nodes（@0x1407f51e5 writeVolumeIndexNodesContext）。
	if err := writeVolumeIndexNodesContext(ctx, tmp, nodes); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	// sortedIndices（@0x1407f522a writeVolumeIndexSortedIndicesContext）。
	if err := writeVolumeIndexSortedIndicesContext(ctx, tmp, sortedIndices); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	// contextErr 重检（@0x1407f5248 contextErr + @0x1407f5250 jne→清理）。
	if err := contextErr(ctx); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	// namePool（@0x1407f528d writeAllContext）。
	if err := writeAllContext(ctx, tmp, namePool); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	// Close（@0x1407f52a6 os.file.close）。
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	// 关闭旧映射提供者（@0x1407f52cc closeMappedReadProviderForReplaceLocked）。
	closeMappedReadProviderForReplaceLocked(v, path)
	// 原子替换（@0x1407f52fa replaceFileAtomically）。
	if err := replaceFileAtomically(tmpName, path); err != nil {
		return err
	}
	// (4) 字段更新（@0x1407f5309..551c，heapFlag 常为 1，径直执行）。
	v.Nodes = nodes
	v.SortedIndices = sortedIndices
	v.NamePool = namePool
	v.activeCount = activeCount
	v.deltaByFRN = nil
	v.deltaReplaced = nil
	v.deltaNodes = nil
	// 重建前缀桶（@0x1407f5483 buildNamePrefixBucketCounts）。
	v.prefixBuckets = buildNamePrefixBucketCounts(nodes, namePool)
	// 清除旧三字索引（@0x1407f5509 clearNameTrigramIndexLocked）。
	v.clearNameTrigramIndexLocked()
	return nil
}

// 名称三元组（trigram）索引头二进制格式常量（64 字节头，SSOT 见 parseVolumeNameTrigramHeader）。
const (
	volumeNameTrigramHeaderMagic   = "UITG"
	volumeNameTrigramHeaderVersion = 1
	volumeIndexFormatVersion       = 0x20003
	maxVolumeIndexEntryCount       = 0xffffffff
)

// 名称三元组索引流式读取分块常量（每批最多 0x8000 条 = 0x40000 字节，实证自 0x1407f9300）。
const (
	volumeNameTrigramStreamChunkEntries = 0x8000
	volumeNameTrigramStreamChunkBytes   = 0x40000
)

// 卷索引 checkpoint 头二进制格式常量（64 字节头，SSOT 见 parseVolumeIndexCheckpointHeader）。
const (
	volumeIndexCheckpointHeaderMagic = "UIDX"
	volumeIndexCheckpointHeaderSize  = 64
)

// 名称前缀桶索引（.bkt）二进制格式常量（0x440 字节头，SSOT 见 load/saveVolumeIndexBucketContext）。
const (
	volumeIndexBucketHeaderMagic = uint64(0x4b424d4145425355) // "USBEAMBK"
	volumeIndexBucketVersion     = uint16(0x3254)             // "T2"
	volumeIndexBucketHeaderSize  = 0x440
	volumeIndexBucketNameLen     = 0xa // CreateTemp 模式 ".bkt-*.tmp" 的随机名长度
)

// loadVolumeIndexBucketContext 从 .bkt 文件加载名称前缀桶计数与身份校验。
// [S 汇编 0x1407f6f00, 230B] 实证：签名 (ctx context.Context, path string,
// want *volumeIndexBucketIdentity, enabled bool) (*namePrefixBucketCounts, int, error)。
// TrimSpace(path)==""（@0x1407f702e strings.TrimSpace + @0x1407f7043 test rcx,rcx）→
// (nil, 0, errVolumeIndexBucketPathEmpty)；os.ReadFile(path)（@0x1407f7061）→
// err!=nil（@0x1407f706d jne）→ (nil, 0, err)；len(data)<0x440（@0x1407f7083 cmp jl）→
// (nil, 0, errVolumeIndexNodeBytesInvalid)；magic@0≠USBEAMBK 或 version@8≠T2（@0x1407f7098/9c）→
// (nil, 0, errVolumeIndexNodeBytesInvalid)；entryCount=dword@0xa（@0x1407f70aa）；want!=nil
// （@0x1407f70b6 test rdx）→ 4×qword 身份比较，任一不等→哨兵；namePoolSize=dword@0xe（@0x1407f70d8 test），
// len(data)!=0x440+namePoolSize*4（@0x1407f70fc jne）→哨兵；new(namePrefixBucketCounts)
// （@0x1407f710a）→ 逐 int32 LE 解析 firstByte（@0x1407f7129/3b duffcopy+loop）、
// namePoolSize>0（@0x1407f715c test）→ twoByte=make([]int32,namePoolSize)+int32-LE 逐条拷贝
// → return (bucket, entryCount, nil)（@0x1407f719c）。
func loadVolumeIndexBucketContext(ctx context.Context, path string, want *volumeIndexBucketIdentity, enabled bool) (*namePrefixBucketCounts, int, error) {
	_ = enabled
	if strings.TrimSpace(path) == "" {
		return nil, 0, errVolumeIndexBucketPathEmpty
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}
	if len(data) < volumeIndexBucketHeaderSize {
		return nil, 0, errVolumeIndexNodeBytesInvalid
	}
	if binary.LittleEndian.Uint64(data[0:8]) != volumeIndexBucketHeaderMagic {
		return nil, 0, errVolumeIndexNodeBytesInvalid
	}
	if binary.LittleEndian.Uint16(data[8:10]) != volumeIndexBucketVersion {
		return nil, 0, errVolumeIndexNodeBytesInvalid
	}
	entryCount := int(binary.LittleEndian.Uint32(data[0xa:0xe]))
	if want != nil {
		if want.rootFRN != binary.LittleEndian.Uint64(data[0x12:0x1a]) ||
			want.journalID != binary.LittleEndian.Uint64(data[0x1a:0x22]) ||
			want.lastUSN != int64(binary.LittleEndian.Uint64(data[0x22:0x2a])) ||
			want.generatedAt != int64(binary.LittleEndian.Uint64(data[0x2a:0x32])) {
			return nil, 0, errVolumeIndexNodeBytesInvalid
		}
	}
	namePoolSize := int(binary.LittleEndian.Uint32(data[0xe:0x12]))
	if len(data) != volumeIndexBucketHeaderSize+namePoolSize*4 {
		return nil, 0, errVolumeIndexNodeBytesInvalid
	}
	bucket := new(namePrefixBucketCounts)
	for i := 0; i < 256; i++ {
		bucket.firstByte[i] = int32(binary.LittleEndian.Uint32(data[0x40+i*4 : 0x44+i*4]))
	}
	if namePoolSize > 0 {
		bucket.twoByte = make([]int32, namePoolSize)
		for i := 0; i < namePoolSize; i++ {
			bucket.twoByte[i] = int32(binary.LittleEndian.Uint32(data[volumeIndexBucketHeaderSize+i*4 : volumeIndexBucketHeaderSize+(i+1)*4]))
		}
	}
	return bucket, entryCount, nil
}

// saveVolumeIndexBucketContext 持久化卷前缀桶索引到 .bkt 文件（原子替换）。
// [S 汇编 0x1407f68e0, 328B] 实证：签名 (ctx context.Context, v *VolumeIndex, path string,
// counts namePrefixBucketCounts) error（counts 0x418 字节值传参 @rsp+0x8a0）。
// v==nil（@0x1407f6915 test rcx,je）或 TrimSpace(path)==""（@0x1407f694d）→ nil；
// os.MkdirAll(filepath.Dir(path), 0o755)（@0x1407f6983，ecx=0x1ed）→ err→return err；
// 默认 counts = v.prefixBuckets 全字段拷贝（@0x1407f6991/9b7 rep movsq 0x82 qwords 拷贝
// 0x418 字节到栈 local @rsp+0x898）；delta 分支：len(v.deltaByFRN)>0（@0x1407f69bd/69d0 jg）
// 或 v.deltaReplaced!=0（@0x1407f69dc）→ buildVolumeIndexCheckpointSnapshotLocked(v)
// （@0x1407f69e6）→ entryCount=activeCount、counts=buildNamePrefixBucketCounts(nodes,namePool)；
// 否则 entryCount=len(v.SortedIndices)（@0x1407f69d5）。header 0x40 字节小端写入 magic@0、
// version@8、entryCount@0xa、namePoolSize（len(counts.twoByte) 当 twoByte.ptr≠nil 否则 0）@0xe、
// rootFRN@0x12、journalID@0x1a、lastUSN@0x22、generatedAt@0x2a=timeToUnixNano(v.GeneratedAt)。
// os.CreateTemp(Dir(path), ".bkt-*.tmp")（@0x1407f6a74，edi=0xa）→ 写 header[:0x40]
// → firstByte[256] int32-LE 写 0x400 字节 → twoByte 段（namePoolSize>0 时 namePoolSize*4 字节
// int32-LE）→ close → replaceFileAtomically(temp,path)。任一写/关闭失败则清理临时文件后返回错误。
func saveVolumeIndexBucketContext(ctx context.Context, v *VolumeIndex, path string, counts namePrefixBucketCounts) error {
	if v == nil || strings.TrimSpace(path) == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	entryCount := len(v.SortedIndices)
	if len(v.deltaByFRN) > 0 || v.deltaReplaced != nil {
		nodes, namePool, _, activeCount, err := buildVolumeIndexCheckpointSnapshotLocked(v)
		if err != nil {
			return err
		}
		entryCount = activeCount
		counts = buildNamePrefixBucketCounts(nodes, namePool)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".bkt-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	var header [0x40]byte
	binary.LittleEndian.PutUint64(header[0:8], volumeIndexBucketHeaderMagic)
	binary.LittleEndian.PutUint16(header[8:10], volumeIndexBucketVersion)
	binary.LittleEndian.PutUint32(header[0xa:0xe], uint32(entryCount))
	namePoolSize := uint32(len(counts.twoByte))
	binary.LittleEndian.PutUint32(header[0xe:0x12], namePoolSize)
	binary.LittleEndian.PutUint64(header[0x12:0x1a], v.RootFRN)
	binary.LittleEndian.PutUint64(header[0x1a:0x22], v.JournalID)
	binary.LittleEndian.PutUint64(header[0x22:0x2a], uint64(v.LastUSN))
	binary.LittleEndian.PutUint64(header[0x2a:0x32], uint64(timeToUnixNano(v.GeneratedAt)))
	if err := writeAllContext(ctx, tmp, header[:]); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	firstByteBuf := make([]byte, 0x400)
	for i := 0; i < 256; i++ {
		binary.LittleEndian.PutUint32(firstByteBuf[i*4:], uint32(counts.firstByte[i]))
	}
	if err := writeAllContext(ctx, tmp, firstByteBuf); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if namePoolSize > 0 {
		twoByteBuf := make([]byte, int(namePoolSize)*4)
		for i := uint32(0); i < namePoolSize; i++ {
			binary.LittleEndian.PutUint32(twoByteBuf[i*4:], uint32(counts.twoByte[i]))
		}
		if err := writeAllContext(ctx, tmp, twoByteBuf); err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
			return err
		}
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return replaceFileAtomically(tmpName, path)
}

// buildVolumeIndexCheckpointSnapshotLocked 构建检查点快照（持锁，返回节点/名字池/排序索引/活动计数）。
// [S-sig 汇编 0x1407e70a0, 314B] 实证：签名 (v *VolumeIndex) (nodes []IndexNode, namePool []byte,
// sortedIndices []int32, activeCount int, err error)。v==nil→(零,零,nil)；无 delta
// （len(deltaByFRN)==0 && deltaReplaced==0）且 v.Nodes 非空或 v.NamePool 非空→字段深拷贝；
// 无 delta 且两者皆空→若 readProvider 为 MappedReadProvider 则走 HeapSnapshot 路径（gap），否则零返回；
// 有 delta→走 HeapSnapshot + applyTombstones + applyDeltaOverlay 路径（gap，体待专项还原）。
func buildVolumeIndexCheckpointSnapshotLocked(v *VolumeIndex) (nodes []IndexNode, namePool []byte, sortedIndices []int32, activeCount int, err error) {
	if v == nil {
		return nil, nil, nil, 0, nil
	}
	if len(v.deltaByFRN) == 0 && v.deltaReplaced == nil {
		if len(v.Nodes) > 0 || len(v.NamePool) > 0 {
			nodes = make([]IndexNode, len(v.Nodes))
			copy(nodes, v.Nodes)
			namePool = make([]byte, len(v.NamePool))
			copy(namePool, v.NamePool)
			sortedIndices = make([]int32, len(v.SortedIndices))
			copy(sortedIndices, v.SortedIndices)
			return nodes, namePool, sortedIndices, v.activeCount, nil
		}
		if _, ok := v.readProvider.(*volumeIndexMappedReadProvider); ok {
			return nil, nil, nil, 0, errVolumeIndexNodeBytesInvalid
		}
		return nil, nil, nil, 0, nil
	}
	return nil, nil, nil, 0, errVolumeIndexNodeBytesInvalid
}

// loadVolumeIndexCheckpointMappedViewContext mmap 方式加载 checkpoint 视图（含可选 .bkt 桶加载）。
// [S 汇编 0x1407fb100, 1041B] 实证：签名 (ctx context.Context, root string, path string,
// bucketPath string) (*volumeIndexCheckpointView, error)（bucketPath 为第 4 寄存参数 @r10/r11）。
// 流程：contextErr→return；os.OpenFile(path,O_RDONLY,0)→err 分派 errVolumeIndexFileNotFound
// 或原始 err；f.Stat()→err→close+return；io.ReadAtLeast(f,headerBuf[:64],64)→err→close+
// errVolumeIndexNodeBytesInvalid；close f；parseVolumeIndexCheckpointHeader→err→return；
// buildVolumeIndexCheckpointLayout(fi.Size(),entryCount,sortedCount,namePoolSize)→err→return；
// planVolumeIndexCheckpointMapWithSystemGranularity→err→return；
// openVolumeIndexMappedFile→err→return；m.Map×3 node/sorted/namePool→err→Close+return；
// contextErr recheck→Close+return；validateVolumeIndexNodeBytes→err→Close+return。
// 分支 shouldUseVolumeIndexMmapStaticProvider：true→nodes/sortedIndices/namePool=nil、
// 校验 sortedBytes 长度与 sortedCount、逐个 sortedIndex∈[0,nodeCount)；false→
// decodeVolumeIndexNodesFromBytesWithNodeCount→nodes+activeCount→
// sortedIndices int32-LE copy+逐个越界校验、namePool=copy(namePoolBytes)。
// 构造 view#1（prefixBuckets=nil）。桶段：bucketPath 非空→want 身份→
// loadVolumeIndexBucketContext；成功且非 nil→prefixBuckets=counts；否则回退重建
// buildNamePrefixBucketCountsFromNodeBytes→err→Close+return；recounts 赋值 prefixBuckets、
// go func(){_ = saveVolumeIndexBucketContext(context.Background(), &tmp, bucketPath, recounts)}()。
// 构造 view#2（prefixBuckets 已赋值）→ return (view, nil)。
func loadVolumeIndexCheckpointMappedViewContext(ctx context.Context, root string, path string, bucketPath string) (*volumeIndexCheckpointView, error) {
	if cerr := contextErr(ctx); cerr != nil {
		return nil, cerr
	}
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errVolumeIndexFileNotFound
		}
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	var headerBuf [volumeIndexCheckpointHeaderSize]byte
	if _, err := io.ReadAtLeast(f, headerBuf[:], volumeIndexCheckpointHeaderSize); err != nil {
		f.Close()
		return nil, errVolumeIndexNodeBytesInvalid
	}
	f.Close()

	header, err := parseVolumeIndexCheckpointHeader(headerBuf[:])
	if err != nil {
		return nil, err
	}
	layout, err := buildVolumeIndexCheckpointLayout(fi.Size(), header.entryCount, header.sortedCount, header.namePoolSize)
	if err != nil {
		return nil, err
	}
	plans, err := planVolumeIndexCheckpointMapWithSystemGranularity(layout)
	if err != nil {
		return nil, err
	}
	m, err := openVolumeIndexMappedFile(path)
	if err != nil {
		return nil, err
	}
	nodeBytes, err := m.Map(plans[0].MapOffset, plans[0].MapBytes, plans[0].SectionOffset, plans[0].SectionBytes)
	if err != nil {
		m.Close()
		return nil, err
	}
	sortedBytes, err := m.Map(plans[1].MapOffset, plans[1].MapBytes, plans[1].SectionOffset, plans[1].SectionBytes)
	if err != nil {
		m.Close()
		return nil, err
	}
	namePoolBytes, err := m.Map(plans[2].MapOffset, plans[2].MapBytes, plans[2].SectionOffset, plans[2].SectionBytes)
	if err != nil {
		m.Close()
		return nil, err
	}
	if cerr := contextErr(ctx); cerr != nil {
		m.Close()
		return nil, cerr
	}
	if err := validateVolumeIndexNodeBytes(nodeBytes, int(header.namePoolSize)); err != nil {
		m.Close()
		return nil, err
	}

	var nodes []IndexNode
	var sortedIndices []int32
	var namePool []byte
	activeCount := 0

	if shouldUseVolumeIndexMmapStaticProvider() {
		if len(sortedBytes)%4 != 0 {
			m.Close()
			return nil, errVolumeIndexNodeBytesInvalid
		}
		sortedCount := len(sortedBytes) / 4
		if int(header.sortedCount) != sortedCount {
			m.Close()
			return nil, errVolumeIndexNodeBytesInvalid
		}
		nodeCount := len(nodeBytes) / 24
		for i := 0; i < sortedCount; i++ {
			idx := int32(binary.LittleEndian.Uint32(sortedBytes[i*4:]))
			if idx < 0 || int(idx) >= nodeCount {
				m.Close()
				return nil, errVolumeIndexNodeBytesInvalid
			}
		}
	} else {
		var derr error
		nodes, activeCount, derr = decodeVolumeIndexNodesFromBytesWithNodeCount(nodeBytes, len(nodeBytes)/24)
		if derr != nil {
			m.Close()
			return nil, derr
		}
		if len(sortedBytes)%4 != 0 {
			m.Close()
			return nil, errVolumeIndexNodeBytesInvalid
		}
		sortedIndices = make([]int32, len(sortedBytes)/4)
		nodeCount := len(nodeBytes) / 24
		for i := range sortedIndices {
			sortedIndices[i] = int32(binary.LittleEndian.Uint32(sortedBytes[i*4:]))
			if sortedIndices[i] < 0 || int(sortedIndices[i]) >= nodeCount {
				m.Close()
				return nil, errVolumeIndexNodeBytesInvalid
			}
		}
		namePool = append([]byte(nil), namePoolBytes...)
	}

	cv := &volumeIndexCheckpointView{
		root:          root,
		rootFRN:       header.rootFRN,
		journalID:     header.journalID,
		lastUSN:       header.lastUSN,
		generatedAt:   header.generatedAt,
		layout:        layout,
		mappedFile:    m,
		nodes:         nodes,
		nodeBytes:     nodeBytes,
		sortedIndices: sortedIndices,
		sortedBytes:   sortedBytes,
		namePool:      namePool,
		namePoolBytes: namePoolBytes,
		activeCount:   activeCount,
	}

	var prefixBuckets *namePrefixBucketCounts
	if bucketPath != "" {
		want := volumeIndexBucketIdentity{
			rootFRN:     header.rootFRN,
			journalID:   header.journalID,
			lastUSN:     header.lastUSN,
			generatedAt: timeToUnixNano(header.generatedAt),
		}
		counts, _, lerr := loadVolumeIndexBucketContext(ctx, bucketPath, &want, true)
		if lerr != nil || counts == nil {
			recounts, _, rerr := buildNamePrefixBucketCountsFromNodeBytes(nodeBytes, namePoolBytes)
			if rerr != nil {
				m.Close()
				return nil, rerr
			}
			prefixBuckets = &recounts
			go func() {
				var tmp VolumeIndex
				tmp.RootFRN = header.rootFRN
				tmp.JournalID = header.journalID
				tmp.LastUSN = header.lastUSN
				tmp.GeneratedAt = header.generatedAt
				tmp.activeCount = activeCount
				tmp.prefixBuckets = recounts
				_ = saveVolumeIndexBucketContext(context.Background(), &tmp, bucketPath, recounts)
			}()
		} else {
			prefixBuckets = counts
		}
	}
	cv.prefixBuckets = prefixBuckets
	return cv, nil
}

// buildVolumeNameTrigramHeaderLocked 构建名称三元组索引头（调用方持锁）。
// [S 汇编 0x1407f66e0, 512B] 实证：entryCount∉[0,0xffffffff] → 返回 64B 零头；
// 否则写 magic "UITG"(+0)、version 1(+4)、indexVersion 0x20003(+8)、entryCount(+0xc)、
// rootFRN(+0x10)、journalID(+0x18)、lastUSN(+0x20)、GeneratedAt.Unix()(+0x28)、
// LastMutationAt.UnixNano()(+0x30)、保留(+0x38)。
func buildVolumeNameTrigramHeaderLocked(idx *VolumeIndex, entryCount int) [64]byte {
	var h [64]byte
	if entryCount < 0 || entryCount > maxVolumeIndexEntryCount {
		return h
	}
	copy(h[0:4], volumeNameTrigramHeaderMagic)
	binary.LittleEndian.PutUint32(h[4:8], volumeNameTrigramHeaderVersion)
	binary.LittleEndian.PutUint32(h[8:12], volumeIndexFormatVersion)
	binary.LittleEndian.PutUint32(h[12:16], uint32(entryCount))
	binary.LittleEndian.PutUint64(h[16:24], idx.RootFRN)
	binary.LittleEndian.PutUint64(h[24:32], idx.JournalID)
	binary.LittleEndian.PutUint64(h[32:40], uint64(idx.LastUSN))
	binary.LittleEndian.PutUint64(h[40:48], uint64(idx.GeneratedAt.Unix()))
	binary.LittleEndian.PutUint64(h[48:56], uint64(idx.LastMutationAt.UnixNano()))
	// h[56:64] 保留为零。
	return h
}

// stopUSNFollowerLocked 停止指定卷的 USN follower（map 删除 + 关闭 channel + CancelIoEx）。
// [S-sig 0x14081b4e0, 288B]：normalizeVolumeRoot → map 查找 → 删除 → closechan →
// CancelIoEx。体待 USN follower 结构字段名专项还原。
func (s *FileIndexService) stopUSNFollowerLocked(a, b interface{}) {
	_, _ = a, b
}

// Close 关闭映射读取提供者（closing 置位，归零时调用 release）。
// [S 汇编 0x1407e4300, 320B]：nil/closed→(nil)；closing=true；active>0→(nil)；
// 否则取 release 置 nil + closed=true；unlock 后 release 非空则调用。
func (p *volumeIndexMappedReadProvider) Close() error {
	if p == nil {
		return nil
	}
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closing = true
	if p.active > 0 {
		p.mu.Unlock()
		return nil
	}
	rel := p.release
	p.release = nil
	p.closed = true
	p.mu.Unlock()
	if rel != nil {
		rel()
	}
	return nil
}

// OverlayStats 返回 overlay 统计（RLock 包装 overlayStatsLocked）。
// [S-sig 0x1407e7c20, 320B]：RLock(+0x10) → defer RUnlock → overlayStatsLocked。
// 体待 overlayStatsLocked 落地。
func (s *VolumeIndex) OverlayStats() interface{} {
	_ = s
	return nil
}

// overlayStatsLocked 计算 overlay 统计（持锁，5 值返回）。
// [S-sig 0x1407e7fe0, 320B]：baseEntryCountLocked → 迭代节点计数 → 比率计算。
// 体待 overlay 统计结构专项还原。
func (s *VolumeIndex) overlayStatsLocked() interface{} {
	_ = s
	return nil
}

// lockUSNFollowerSidecar 锁定 USN follower 侧车（全局 map LoadOrStore mutex）。
// [S-sig 0x140818440, 320B]：usnFollowerPaths → HashTrieMap LoadOrStore → lock。
// 体待侧车锁结构专项还原。
func (s *FileIndexService) lockUSNFollowerSidecar(a, b interface{}) interface{} {
	_, _ = a, b
	return nil
}

// normalizeFileSearchUSNFollowerMeta 归一化 USN follower 元数据（默认值填充）。
// [S-sig 0x140818580, 320B]：enabled==0→1；normalizeVolumeRoot；limit<字段→字段；
// timestamp==0→time.Now()。体待元数据结构字段名专项还原。
func normalizeFileSearchUSNFollowerMeta(a interface{}) interface{} {
	_ = a
	return nil
}

// CloseReadProvider 关闭读提供者（持写锁，清理索引）。
// [S-sig 0x1407e6d60, 352B]：RLock → closeVolumeIndexReadProvider →
// clearNameTrigramIndexLocked → pinyinIndex(+0x538).Close + 置零 → 标志(+0x530)=false。
// 体待索引结构专项还原。
func (v *VolumeIndex) CloseReadProvider() error {
	_ = v
	return nil
}

// writeAllContext 将 p 全部写入 w，每次写前检查 ctx 是否已取消。
// [S 汇编 0x1407f6420, 352B] 实证：签名 (ctx context.Context, w io.Writer, p []byte) error。
// 循环 for len(p)>0：contextErr(ctx)!=nil→return err；w.Write(p)（itab+0x18 方法槽、
// data@+0x60 为 receiver）→ err!=nil→return；n<=0→return io.ErrShortWrite（全局
// {0x1411d3100,0x141bc4610}→errorString{0x140c47b7f,11}="short write"）；
// n>len(p) 则 runtime.panicSliceB（编译器边界检查）；否则 p=p[n:] 推进；循环结束返回 nil。
func writeAllContext(ctx context.Context, w io.Writer, p []byte) error {
	for len(p) > 0 {
		if err := contextErr(ctx); err != nil {
			return err
		}
		n, err := w.Write(p)
		if err != nil {
			return err
		}
		if n <= 0 {
			return io.ErrShortWrite
		}
		p = p[n:]
	}
	return nil
}

// writeVolumeIndexSortedIndicesContext 将排序索引（[]int32）按小端分块写入。
// [S 汇编 0x1407f61e0, 576B] 实证：签名 (ctx context.Context, w io.Writer, indices []int32) error。
// len==0→nil；makeslice []byte(0x20000) 复用缓冲；外层游标循环每块 ≤0x8000 个索引；
// 内层把 int32 逐个小端写入 buf[i*4:]（x86 下等价 memcpy）；writeAllContext(ctx,w,buf[:chunk*4])。
func writeVolumeIndexSortedIndicesContext(ctx context.Context, w io.Writer, indices []int32) error {
	if len(indices) == 0 {
		return nil
	}
	buf := make([]byte, 0x20000)
	for cursor := 0; cursor < len(indices); {
		if err := contextErr(ctx); err != nil {
			return err
		}
		chunk := len(indices) - cursor
		if chunk > 0x8000 {
			chunk = 0x8000
		}
		for i := 0; i < chunk; i++ {
			binary.LittleEndian.PutUint32(buf[i*4:], uint32(indices[cursor+i]))
		}
		if err := writeAllContext(ctx, w, buf[:chunk*4]); err != nil {
			return err
		}
		cursor += chunk
	}
	return nil
}

// writeVolumeNameTrigramSignaturesContext 将卷名三元组签名（[]uint64）按小端分块写入。
// [S 汇编 0x1407f7cc0, 576B] 实证：签名 (ctx context.Context, w io.Writer, signatures []uint64) error。
// 与 sortedIndices 同构：缓冲 0x40000、块 ≤0x8000 个 uint64、每元素 8 字节小端。
func writeVolumeNameTrigramSignaturesContext(ctx context.Context, w io.Writer, signatures []uint64) error {
	if len(signatures) == 0 {
		return nil
	}
	buf := make([]byte, 0x40000)
	for cursor := 0; cursor < len(signatures); {
		if err := contextErr(ctx); err != nil {
			return err
		}
		chunk := len(signatures) - cursor
		if chunk > 0x8000 {
			chunk = 0x8000
		}
		for i := 0; i < chunk; i++ {
			binary.LittleEndian.PutUint64(buf[i*8:], signatures[cursor+i])
		}
		if err := writeAllContext(ctx, w, buf[:chunk*8]); err != nil {
			return err
		}
		cursor += chunk
	}
	return nil
}

// buildVolumeIndexPersistenceMetaLocked 构建卷索引持久化元数据（持锁）。
// [S 汇编 0x1407f6580, 352B] 实证：签名 (v *VolumeIndex, now time.Time) volumeIndexPersistenceMeta。
// Version=1、IndexVersion=0x20003（movabs 0x2000300000001 写入前 8 字节）；
// RootFRN=v.RootFRN(+0x30)、JournalID=v.JournalID(+0x80)、LastUSN=v.LastUSN(+0x88)；
// 三个时间字段分别 timeToUnixNano(v.GeneratedAt@+0x90 / v.LastMutationAt@+0xa8 / now)。
func buildVolumeIndexPersistenceMetaLocked(v *VolumeIndex, now time.Time) volumeIndexPersistenceMeta {
	return volumeIndexPersistenceMeta{
		Version:                1,
		IndexVersion:           0x20003,
		RootFRN:                v.RootFRN,
		JournalID:              v.JournalID,
		LastUSN:                v.LastUSN,
		GeneratedAtUnixNano:    timeToUnixNano(v.GeneratedAt),
		LastMutationAtUnixNano: timeToUnixNano(v.LastMutationAt),
		LastSavedAtUnixNano:    timeToUnixNano(now),
	}
}

// errReplaceFileAtomically 原子替换耗尽重试后的兜底错误。
// 消息字符串位于 .data BSS 静态区（运行时初始化），非 .rdata 常量，未直接读取；
// 该分支仅在 lastErr 全程为 nil 时可达（循环首轮 rename 失败即赋值，实际不可达）。
var errReplaceFileAtomically = errors.New("replace file atomically failed")

// replaceFileAtomically 原子替换 src→dst（Windows rename 不覆盖已存在文件，
// 故先备份 dst 到临时文件再替换，失败回滚）。最多 8 轮、逐轮 25ms 退避。
// [S 汇编 0x1407f8140, 1280B] 实证：签名 (src, dst string) error。
// 每轮：rename(src,dst) 成功→nil；失败 lastErr=err。CreateTemp(dir, "."+base+"-*.bak")
// 备份名；close→remove 临时占位→rename(dst,tmpName) 备份；rename(src,dst) 正式替换，
// 失败则 rename(tmpName,dst) 回滚，回滚失败返回 fmt.Errorf("替换 %q 失败且回滚旧文件失败
// （新文件保留在 %q，旧文件保留在 %q）: replace=%v rollback=%w", dst, src, tmpName, err, rbErr)；
// 成功则 remove 备份并返回 nil。rename(dst,tmpName) 失败且非 os.ErrNotExist 时 lastErr=err。
func replaceFileAtomically(src, dst string) error {
	var lastErr error
	for i := 0; i < 8; i++ {
		if err := os.Rename(src, dst); err == nil {
			return nil
		} else {
			lastErr = err
		}
		tmp, err := os.CreateTemp(filepath.Dir(dst), "."+filepath.Base(dst)+"-*.bak")
		if err != nil {
			if i < 7 {
				time.Sleep(time.Duration(i+1) * 25 * time.Millisecond)
			}
			continue
		}
		tmpName := tmp.Name()
		if err := tmp.Close(); err != nil {
			_ = os.Remove(tmpName)
			if i < 7 {
				time.Sleep(time.Duration(i+1) * 25 * time.Millisecond)
			}
			continue
		}
		if err := os.Remove(tmpName); err != nil {
			if i < 7 {
				time.Sleep(time.Duration(i+1) * 25 * time.Millisecond)
			}
			continue
		}
		if err := os.Rename(dst, tmpName); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				lastErr = err
			}
			if i < 7 {
				time.Sleep(time.Duration(i+1) * 25 * time.Millisecond)
			}
			continue
		}
		if err := os.Rename(src, dst); err != nil {
			if rbErr := os.Rename(tmpName, dst); rbErr != nil {
				return fmt.Errorf("替换 %q 失败且回滚旧文件失败（新文件保留在 %q，旧文件保留在 %q）: replace=%v rollback=%w",
					dst, src, tmpName, err, rbErr)
			}
			lastErr = err
			if i < 7 {
				time.Sleep(time.Duration(i+1) * 25 * time.Millisecond)
			}
			continue
		}
		_ = os.Remove(tmpName)
		return nil
	}
	if lastErr != nil {
		return lastErr
	}
	return errReplaceFileAtomically
}

// saveVolumeIndexMetaContext 持久化卷索引元数据（原子替换）。
// [S 汇编 0x1407f7f00, 576B] 实证：签名 (ctx context.Context, path string, meta volumeIndexPersistenceMeta) error。
// TrimSpace(path) 空→nil；contextErr；MkdirAll(Dir(path),0o755)；json.Marshal(meta)；
// CreateTemp(Dir(path), ".meta-*.tmp")；Write(data)；Close；replaceFileAtomically(tmp, path)。
// 任一失败清理临时文件后返回错误。
func saveVolumeIndexMetaContext(ctx context.Context, path string, meta volumeIndexPersistenceMeta) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	if err := contextErr(ctx); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".meta-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return replaceFileAtomically(tmpName, path)
}

// openVolumeIndexMappedFile 打开卷索引映射文件（OpenFile + CreateFileMapping）。
// [S 汇编 0x1407fd500, 352B]：os.OpenFile(path, O_RDONLY, 0)（@0x1407fd520）；
// err!=nil 且 errors.Is(err, os.ErrNotExist)（@0x1407fd607）→ (nil,nil)，否则 (nil,err)；
// windows.CreateFileMapping(f.Fd(), nil, PAGE_READONLY=2, 0, 0, nil)（@0x1407fd569，rcx=2）；
// err!=nil → f.Close() → (nil,err)；否则 newobject{volumeIndexMappedFile}，
// m.file=f、m.mapping=handle（@0x1407fd5d0/5d8）→ (m, nil)。
func openVolumeIndexMappedFile(path string) (*volumeIndexMappedFile, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	mapping, err := windows.CreateFileMapping(windows.Handle(f.Fd()), nil, windows.PAGE_READONLY, 0, 0, nil)
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	return &volumeIndexMappedFile{file: f, mapping: uintptr(mapping)}, nil
}

// Map 映射索引文件 [mapOffset, mapOffset+mapBytes) 到内存视图，返回
// 视图内 [sectionOffset, sectionOffset+sectionBytes) 的切片。
// [S 汇编 0x1407fd660, 544B] 实证：签名 (m, mapOffset int64, mapBytes int64,
// sectionOffset int64, sectionBytes int64) ([]byte, error)。
// m==nil 或 mapping==0（@0x1407fd672/680）→ errVolumeIndexNodeBytesInvalid；
// mapBytes<=0 或 sectionBytes<=0（@0x1407fd689/68e）→ (nil,nil)；
// mapOffset<0 或 sectionOffset<0（@0x1407fd6a5/6aa）→ errVolumeIndexNodeBytesInvalid；
// windows.MapViewOfFile(mapping, FILE_MAP_READ=4, mapOffset>>32, mapOffset, mapBytes)
// （@0x1407fd6f3）；err→(nil,err)、addr==0→errVolumeIndexNodeBytesInvalid；
// views 追加 addr（@0x1407fd77c，Close 逆序 Unmap 用）；
// mapBytes < sectionOffset+sectionBytes（@0x1407fd7a8 jge 失败）→ errVolumeIndexNodeBytesInvalid；
// 否则 unsafe.Slice(addr, mapBytes)[sectionOffset : sectionOffset+sectionBytes : mapBytes]
// （@0x1407fd7d8 ptr=base+sectionOffset、len=sectionBytes、cap=mapBytes-sectionOffset）。
func (m *volumeIndexMappedFile) Map(mapOffset, mapBytes, sectionOffset, sectionBytes int64) ([]byte, error) {
	if m == nil || m.mapping == 0 {
		return nil, errVolumeIndexNodeBytesInvalid
	}
	if mapBytes <= 0 || sectionBytes <= 0 {
		return nil, nil
	}
	if mapOffset < 0 || sectionOffset < 0 {
		return nil, errVolumeIndexNodeBytesInvalid
	}
	addr, err := windows.MapViewOfFile(windows.Handle(m.mapping), windows.FILE_MAP_READ, uint32(mapOffset>>32), uint32(mapOffset), uintptr(mapBytes))
	if err != nil {
		return nil, err
	}
	if addr == 0 {
		return nil, errVolumeIndexNodeBytesInvalid
	}
	m.views = append(m.views, addr)
	if sectionOffset+sectionBytes < sectionOffset || mapBytes < sectionOffset+sectionBytes {
		return nil, errVolumeIndexNodeBytesInvalid
	}
	data := unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(nil), addr)), int(mapBytes))
	return data[int(sectionOffset):int(sectionOffset+sectionBytes):int(mapBytes)], nil
}

// UpdateMeta 更新卷索引元数据（持锁，字段 + 版本号递增）。
// [S-sig 0x140807f40, 352B]：lock(+0x18) → RLock → 字段(+0x80/+0x88) 更新 →
// version(+0x590)++。体待字段类型专项还原。
func (v *VolumeIndex) UpdateMeta(a, b interface{}) {
	_, _ = a, b
}

// shouldPersistUSNFollowerIdleMeta 判定是否应持久化 USN 跟随者空闲元数据。
// [S-sig 0x14081d200, 352B]：idleThreshold<=0→true；否则 lastPersist.Add(threshold+30s)
// 与 now 比较。体待时间常量专项还原。
func shouldPersistUSNFollowerIdleMeta(a, b interface{}, idleThreshold int64) bool {
	_, _, _ = a, b, idleThreshold
	return true
}

// closeVolumeIndexReadProvider 关闭卷索引读提供者（置 nil + 回调关闭）。
// [S-sig 0x1407e6f20, 384B]：provider(+0x520/+0x528) nil→nil；查找 → 回调 close →
// 置零。体待提供者域专项还原。
func closeVolumeIndexReadProvider(v *VolumeIndex) error {
	_ = v
	return nil
}

// clearNameTrigramIndexLocked 清除名称三字索引（持锁，置 nil + close）。
// [S-sig 0x1407e9120, 384B]：lock(+0x500/+0x510) → 字段(+0x508/+0x518) 置零 →
// trigramIndex.close。体待索引域专项还原。
func (v *VolumeIndex) clearNameTrigramIndexLocked() {
	_ = v
}

// parseVolumeIndexCheckpointHeader 解析卷索引 checkpoint 64B 头。
// [S 汇编 0x1407fc9a0, 608B] 实证：签名 (b []byte) (volumeIndexCheckpointHeader, error)。
// 二进制布局：magic "UIDX"(+0)、version 0x20003(+4)、entryCount(+8)、sortedCount(+0xc)、
// rootFRN(+0x10)、journalID(+0x18)、lastUSN(+0x20)、GeneratedAt.Unix()(+0x28)、
// namePoolSize(+0x30)、保留(+0x38)。校验顺序：len<64（@0x1407fc9c8 jl）、
// magic≠"UIDX"（@0x1407fc9ce）、version≠0x20003（@0x1407fc9db）、
// rootFRN==0（@0x1407fc9ed test）、entryCount≠sortedCount（@0x1407fca00 cmp r10d,r9d）→
// errVolumeIndexNodeBytesInvalid；否则 generatedAt=time.Unix(sec,0)（@0x1407fca5b
// ext=[+0x28]+62135596800=unixToInternal、wall=0、loc=time.Local @rip+0x13c6250）。
func parseVolumeIndexCheckpointHeader(b []byte) (volumeIndexCheckpointHeader, error) {
	if len(b) < volumeIndexCheckpointHeaderSize {
		return volumeIndexCheckpointHeader{}, errVolumeIndexNodeBytesInvalid
	}
	if string(b[0:4]) != volumeIndexCheckpointHeaderMagic {
		return volumeIndexCheckpointHeader{}, errVolumeIndexNodeBytesInvalid
	}
	if binary.LittleEndian.Uint32(b[4:8]) != volumeIndexFormatVersion {
		return volumeIndexCheckpointHeader{}, errVolumeIndexNodeBytesInvalid
	}
	rootFRN := binary.LittleEndian.Uint64(b[16:24])
	if rootFRN == 0 {
		return volumeIndexCheckpointHeader{}, errVolumeIndexNodeBytesInvalid
	}
	entryCount := binary.LittleEndian.Uint32(b[8:12])
	sortedCount := binary.LittleEndian.Uint32(b[12:16])
	if entryCount != sortedCount {
		return volumeIndexCheckpointHeader{}, errVolumeIndexNodeBytesInvalid
	}
	return volumeIndexCheckpointHeader{
		rootFRN:      rootFRN,
		journalID:    binary.LittleEndian.Uint64(b[24:32]),
		lastUSN:      int64(binary.LittleEndian.Uint64(b[32:40])),
		generatedAt:  time.Unix(int64(binary.LittleEndian.Uint64(b[40:48])), 0),
		entryCount:   entryCount,
		sortedCount:  sortedCount,
		namePoolSize: binary.LittleEndian.Uint64(b[48:56]),
	}, nil
}

// buildVolumeIndexCheckpointLayout 构建卷索引 checkpoint 布局（偏移算术）。
// [S 汇编 0x1407fc820, 384B] 实证：签名 (fileSize int64, entryCount uint32,
// sortedCount uint32, namePoolSize uint64) (volumeIndexCheckpointLayout, error)。
// namePoolSize>MaxInt64（@0x1407fc849 cmp rdi,0x7fffffffffffffff ja）→ 哨兵；
// nodeBytes=entryCount*24（@0x1407fc854 *3 后 shl 3）、sortedBytes=sortedCount*4
// （@0x1407fc85e lea rsi*4）；expected=namePoolSize+nodeBytes+sortedBytes+0x40
// （@0x1407fc862/866）；expected<0x40（@0x1407fc872）或 expected<nodeBytes+0x40
// （@0x1407fc880）→ 哨兵（第三道 nameOffset<=expected @0x1407fc885 因 namePoolSize>=0
// 恒真，语义等价省略）；fileSize<expected（@0x1407fc8c0 cmp rax,r10 jge）→ 哨兵；
// 否则布局 FileSize=fileSize、NodeOffset=0x40、NodeBytes=nodeBytes、
// SortedOffset=nodeBytes+0x40、SortedBytes=sortedBytes、NameOffset=nodeBytes+sortedBytes+0x40、
// NameBytes=namePoolSize、ExpectedBytes=expected（@0x1407fc8fa..924 栈写）。
func buildVolumeIndexCheckpointLayout(fileSize int64, entryCount, sortedCount uint32, namePoolSize uint64) (volumeIndexCheckpointLayout, error) {
	if namePoolSize > math.MaxInt64 {
		return volumeIndexCheckpointLayout{}, errVolumeIndexNodeBytesInvalid
	}
	nodeBytes := int64(entryCount) * int64(unsafe.Sizeof(IndexNode{}))
	sortedBytes := int64(sortedCount) * int64(unsafe.Sizeof(int32(0)))
	expectedBytes := int64(namePoolSize) + nodeBytes + sortedBytes + volumeIndexCheckpointHeaderSize
	if expectedBytes < volumeIndexCheckpointHeaderSize || expectedBytes < nodeBytes+volumeIndexCheckpointHeaderSize {
		return volumeIndexCheckpointLayout{}, errVolumeIndexNodeBytesInvalid
	}
	if fileSize < expectedBytes {
		return volumeIndexCheckpointLayout{}, errVolumeIndexNodeBytesInvalid
	}
	return volumeIndexCheckpointLayout{
		FileSize:      fileSize,
		NodeOffset:    volumeIndexCheckpointHeaderSize,
		NodeBytes:     nodeBytes,
		SortedOffset:  nodeBytes + volumeIndexCheckpointHeaderSize,
		SortedBytes:   sortedBytes,
		NameOffset:    nodeBytes + sortedBytes + volumeIndexCheckpointHeaderSize,
		NameBytes:     int64(namePoolSize),
		ExpectedBytes: expectedBytes,
	}, nil
}

// ApplyDelete 应用删除（持锁，canApplyDeleteLocked → applyDeleteLocked → markDirty）。
// [S-sig 0x140802ea0, 384B]：lock(+0x18) → RLock → canApplyDeleteLocked →
// applyDeleteLocked>0 → markDirtyLocked。体待删除域专项还原。
func (v *VolumeIndex) ApplyDelete(a interface{}) int {
	_, _ = v, a
	return 0
}

// baseSubtreeActiveCountsLocked 取子树活动计数（持锁，带缓存）。
// [S-sig 0x140803fc0, 384B]：缓存 field(+0x558) 命中→返回 field(+0x550/+0x560)；
// 否则 buildVolumeIndexBaseSubtreeActiveCounts。体待计数域专项还原。
func (v *VolumeIndex) baseSubtreeActiveCountsLocked(a interface{}) (interface{}, bool) {
	_, _ = v, a
	return nil, false
}

// ResolvePathByNodeIndex 按节点索引解析路径（acquireReadView → resolvePath）。
// [S-sig 0x140808d40, 384B]：acquireReadView → volumeIndexReadView.resolvePath(nodeIndex)。
// 体待路径解析域专项还原。
func (v *VolumeIndex) ResolvePathByNodeIndex(a interface{}, b interface{}) string {
	_, _, _ = v, a, b
	return ""
}

// volumeNameTrigramHeaderMatchesIndex 校验三字索引头是否匹配索引。
// [S 汇编 0x1407f9ca0, 448B] 实证：签名 (header volumeNameTrigramHeader, idx *VolumeIndex) bool
// （header 8 字段拆寄存器 rax..r9、idx=r10 @0x1407f9cd3 test）。校验顺序：
// idx==nil（@0x1407f9cd6 je）→ false；IndexVersion≠0x20003（@0x1407f9cdc cmp）→ false；
// baseEntryCountLocked()≠EntryCount（@0x1407f9cf2 call + @0x1407f9d00 cmp）→ false；
// idx.RootFRN(+0x30)≠header.RootFRN（@0x1407f9d13）→ false；
// idx.JournalID(+0x80)≠（@0x1407f9d22）/ idx.LastUSN(+0x88)≠（@0x1407f9d30）→ false；
// idx.GeneratedAt.Unix()≠header.GeneratedAt（@0x1407f9d80，wall/ext 分派内联 @0x1407f9d49..77）→ false；
// 否则 timeToUnixNano(idx.LastMutationAt)==header.LastMutationAt（@0x1407f9da4 call + sete）。
func volumeNameTrigramHeaderMatchesIndex(header volumeNameTrigramHeader, idx *VolumeIndex) bool {
	if idx == nil {
		return false
	}
	if header.IndexVersion != volumeIndexFormatVersion {
		return false
	}
	if idx.baseEntryCountLocked() != int64(header.EntryCount) {
		return false
	}
	if idx.RootFRN != header.RootFRN {
		return false
	}
	if idx.JournalID != header.JournalID || idx.LastUSN != header.LastUSN {
		return false
	}
	if idx.GeneratedAt.Unix() != header.GeneratedAt {
		return false
	}
	return timeToUnixNano(idx.LastMutationAt) == header.LastMutationAt
}

// loadVolumeNameTrigramIndexMappedContext 以 mmap 方式加载名称三元组索引。
// [S 汇编 0x1407f8d60, 1440B] 实证：签名 (ctx context.Context, path string,
// idx *VolumeIndex) (*volumeNameTrigramIndex, error)。
// idx==nil（@0x1407f8d95 test rsi）→ (nil, errVolumeIndexNodeBytesInvalid)；
// contextErr(ctx)!=nil（@0x1407f8dc6）→ (nil, cerr)；
// os.Stat(path)（@0x1407f8de4）失败且 errors.Is(os.ErrNotExist)（@0x1407f8e18）→
// (nil, errVolumeIndexFileNotFound)，否则 (nil, err)；
// fi.Size()<64（@0x1407f8e6e jl）→ (nil, errVolumeIndexNodeBytesInvalid)；
// openVolumeIndexMappedFile（@0x1407f8e88）失败 → (nil, err)；
// m.Map(0, size, 0, size)（@0x1407f8ee4）失败 → Close + (nil, err)；
// 再次 contextErr（@0x1407f8f1a）→ Close + (nil, cerr)；
// parseVolumeNameTrigramHeader(mapped[:64])（@0x1407f8f4b）失败 → Close + (nil, err)；
// !volumeNameTrigramHeaderMatchesIndex（@0x1407f8fe0）→ Close + (nil, errVolumeIndexNodeBytesInvalid)；
// len(mapped)<64+EntryCount*8（@0x1407f902b）→ Close + (nil, errVolumeIndexNodeBytesInvalid)；
// 否则 mapped[0x40:0x40+EntryCount*8] 作 signatureBytes、mappedFile=m、
// version=runtimeVersionSnapshot → (result, nil)。
func loadVolumeNameTrigramIndexMappedContext(ctx context.Context, path string, idx *VolumeIndex) (*volumeNameTrigramIndex, error) {
	if idx == nil {
		return nil, errVolumeIndexNodeBytesInvalid
	}
	if cerr := contextErr(ctx); cerr != nil {
		return nil, cerr
	}
	fi, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errVolumeIndexFileNotFound
		}
		return nil, err
	}
	fileSize := fi.Size()
	if fileSize < volumeIndexCheckpointHeaderSize {
		return nil, errVolumeIndexNodeBytesInvalid
	}
	m, err := openVolumeIndexMappedFile(path)
	if err != nil {
		return nil, err
	}
	mapped, err := m.Map(0, fileSize, 0, fileSize)
	if err != nil {
		_ = m.Close()
		return nil, err
	}
	if cerr := contextErr(ctx); cerr != nil {
		_ = m.Close()
		return nil, cerr
	}
	header, err := parseVolumeNameTrigramHeader(mapped[:volumeIndexCheckpointHeaderSize])
	if err != nil {
		_ = m.Close()
		return nil, err
	}
	if !volumeNameTrigramHeaderMatchesIndex(header, idx) {
		_ = m.Close()
		return nil, errVolumeIndexNodeBytesInvalid
	}
	entryBytes := int64(header.EntryCount) * 8
	if int64(len(mapped)) < int64(volumeIndexCheckpointHeaderSize)+entryBytes {
		_ = m.Close()
		return nil, errVolumeIndexNodeBytesInvalid
	}
	return &volumeNameTrigramIndex{
		version:        idx.runtimeVersionSnapshot(),
		signatureBytes: mapped[volumeIndexCheckpointHeaderSize : volumeIndexCheckpointHeaderSize+int(entryBytes)],
		mappedFile:     m,
	}, nil
}

// loadVolumeNameTrigramIndexContext 以流式读取方式加载名称三元组索引。
// [S 汇编 0x1407f9300, 2112B] 实证：签名 (ctx context.Context, path string,
// idx *VolumeIndex) (*volumeNameTrigramIndex, error)。
// idx==nil（@0x1407f9360 test rsi）→ (nil, errVolumeIndexNodeBytesInvalid)；
// os.OpenFile(path, O_RDONLY, 0)（@0x1407f938c）失败 → (nil, err)；
// defer f.Close()（@0x1407f93a2，deferwrap1=0x1407f9b40）；
// contextErr（@0x1407f93e0）→ (nil, cerr)；f.Stat()（@0x1407f93f6）失败 → (nil, err)；
// io.ReadAtLeast(f, headerBuf[:], 64)（@0x1407f944a）失败 → (nil, errVolumeIndexNodeBytesInvalid)；
// parseVolumeNameTrigramHeader（@0x1407f9468）失败 → (nil, err)；
// !volumeNameTrigramHeaderMatchesIndex（@0x1407f9500）→ (nil, errVolumeIndexNodeBytesInvalid)；
// fi.Size()<64+EntryCount*8（@0x1407f9558）→ (nil, errVolumeIndexNodeBytesInvalid)；
// signatures=make([]uint64, EntryCount)、buf=make([]byte, 0x40000)；
// 循环（@0x1407f9828）：contextErr 检查 → 本批 min(剩余, 0x8000) 条 →
// io.ReadAtLeast(f, buf, n*8) 失败 → (nil, errVolumeIndexNodeBytesInvalid) →
// 逐条 LE uint64 拷入 signatures[i+j]；
// 结束构造 {version: runtimeVersionSnapshot, signatures: signatures}。
func loadVolumeNameTrigramIndexContext(ctx context.Context, path string, idx *VolumeIndex) (*volumeNameTrigramIndex, error) {
	if idx == nil {
		return nil, errVolumeIndexNodeBytesInvalid
	}
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if cerr := contextErr(ctx); cerr != nil {
		return nil, cerr
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	var headerBuf [volumeIndexCheckpointHeaderSize]byte
	if _, err := io.ReadAtLeast(f, headerBuf[:], volumeIndexCheckpointHeaderSize); err != nil {
		return nil, errVolumeIndexNodeBytesInvalid
	}
	header, err := parseVolumeNameTrigramHeader(headerBuf[:])
	if err != nil {
		return nil, err
	}
	if !volumeNameTrigramHeaderMatchesIndex(header, idx) {
		return nil, errVolumeIndexNodeBytesInvalid
	}
	if fi.Size() < int64(volumeIndexCheckpointHeaderSize)+int64(header.EntryCount)*8 {
		return nil, errVolumeIndexNodeBytesInvalid
	}
	signatures := make([]uint64, int(header.EntryCount))
	buf := make([]byte, volumeNameTrigramStreamChunkBytes)
	total := int(header.EntryCount)
	for i := 0; i < total; {
		if cerr := contextErr(ctx); cerr != nil {
			return nil, cerr
		}
		n := total - i
		if n > volumeNameTrigramStreamChunkEntries {
			n = volumeNameTrigramStreamChunkEntries
		}
		if _, err := io.ReadAtLeast(f, buf[:volumeNameTrigramStreamChunkBytes], n*8); err != nil {
			return nil, errVolumeIndexNodeBytesInvalid
		}
		for j := 0; j < n; j++ {
			signatures[i+j] = binary.LittleEndian.Uint64(buf[j*8 : j*8+8])
		}
		i += n
	}
	return &volumeNameTrigramIndex{
		version:    idx.runtimeVersionSnapshot(),
		signatures: signatures,
	}, nil
}

// loadVolumeNameTrigramIndexIntoVolumeContext 加载名称三元组索引到卷索引（mmap/stream 分派）。
// [S 汇编 0x1407f8940, 1056B] 实证：签名 (ctx context.Context, idx *VolumeIndex,
// path string) error。idx==nil（@0x1407f8967 test rcx）→ nil；
// TrimSpace(path)==""（@0x1407f8997 TrimSpace + @0x1407f89a0 test rbx）→ nil；
// shouldUseVolumeIndexMmapStaticProvider（@0x1407f89af）true →
// loadVolumeNameTrigramIndexMappedContext(ctx, path, idx)；err!=nil 且非 os.ErrNotExist /
// errVolumeIndexFileNotFound / errVolumeIndexNodeBytesInvalid（@0x1407f89fe..8a86）→
// contextErr(ctx)!=nil 则返回 cerr（@0x1407f8ac0），否则打点 logFileSearchPerf 后
// loadVolumeNameTrigramIndexContext(ctx, path, idx) 兜底；false → 直接 stream；
// 最终 err!=nil → 返回 err；result!=nil → nameTrigramMu.Lock(+0x500) +
// nameTrigram=result(+0x508) + Unlock（@0x1407f8c4f..8cd7）；返回 nil。
func loadVolumeNameTrigramIndexIntoVolumeContext(ctx context.Context, idx *VolumeIndex, path string) error {
	if idx == nil {
		return nil
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	var result *volumeNameTrigramIndex
	var err error
	if shouldUseVolumeIndexMmapStaticProvider() {
		result, err = loadVolumeNameTrigramIndexMappedContext(ctx, path, idx)
		if err != nil && !errors.Is(err, os.ErrNotExist) && !errors.Is(err, errVolumeIndexFileNotFound) && !errors.Is(err, errVolumeIndexNodeBytesInvalid) {
			if cerr := contextErr(ctx); cerr != nil {
				return cerr
			}
			logFileSearchPerf("trigram mmap load failed path=%q err=%v fallback=stream", fmt.Sprintf("%q %v", path, err), 2)
			result, err = loadVolumeNameTrigramIndexContext(ctx, path, idx)
		}
	} else {
		result, err = loadVolumeNameTrigramIndexContext(ctx, path, idx)
	}
	if err != nil {
		return err
	}
	if result != nil {
		idx.nameTrigramMu.Lock()
		idx.nameTrigram = result
		idx.nameTrigramMu.Unlock()
	}
	return nil
}

// writeVolumeIndexNodesContext 将 []IndexNode 按 24B 小端格式分块写入（0x2000 块/0x30000 缓冲）。
// [S 汇编 0x1407f5dc0, 1056B] 实证：签名 (ctx context.Context, w io.Writer, nodes []IndexNode) error。
// len(nodes)==0→nil；makeslice([]byte,0x30000)；外层游标循环：contextErr(ctx)→err；
// chunk=min(剩余,0x2000)、chunkBytes=chunk*24；内层逐节点序列化 24B：FRN qword LE@+0、
// NameOffset dword LE@+8、ParentIdx dword LE@+12、ModTime dword LE@+16、
// NameLen word LE@+20、Flags word LE@+22；writeAllContext(ctx,w,buf[:chunkBytes])；
// err!=nil→return；cursor+=chunk 推进；终 return nil。
func writeVolumeIndexNodesContext(ctx context.Context, w io.Writer, nodes []IndexNode) error {
	if len(nodes) == 0 {
		return nil
	}
	buf := make([]byte, 0x30000)
	var cursor int
	for cursor < len(nodes) {
		if err := contextErr(ctx); err != nil {
			return err
		}
		chunk := len(nodes) - cursor
		if chunk > 0x2000 {
			chunk = 0x2000
		}
		chunkBytes := chunk * 24
		for i := 0; i < chunk; i++ {
			node := nodes[cursor+i]
			off := i * 24
			binary.LittleEndian.PutUint64(buf[off:off+8], node.FRN)
			binary.LittleEndian.PutUint32(buf[off+8:off+12], uint32(node.NameOffset))
			binary.LittleEndian.PutUint32(buf[off+12:off+16], uint32(node.ParentIdx))
			binary.LittleEndian.PutUint32(buf[off+16:off+20], uint32(node.ModTime))
			binary.LittleEndian.PutUint16(buf[off+20:off+22], node.NameLen)
			binary.LittleEndian.PutUint16(buf[off+22:off+24], node.Flags)
		}
		if err := writeAllContext(ctx, w, buf[:chunkBytes]); err != nil {
			return err
		}
		cursor += chunk
	}
	return nil
}

// writeVolumeIndexCheckpointSectionsContext 持久化 checkpoint 文件（头+节点+排序+名字池四段）。
// [S 汇编 0x1407f57e0, 1024B] 实证：签名 (ctx context.Context, path string, header [64]byte,
// nodes []byte, sorted []byte, namePool []byte) error。contextErr→return；
// MkdirAll(filepath.Dir(path), 0o755)→err→return；
// os.CreateTemp(filepath.Dir(path), ".data-*.tmp")→err→return；
// 写头 writeAllContext(ctx, tmp, header[:0x40])→err 清理；写节点字节
// writeAllContext(ctx,tmp, nodes)→err 清理；写排序 writeAllContext(ctx,tmp,sorted)
// →err 清理；写名字池 writeAllContext(ctx,tmp,namePool)→err 清理；
// tmp.Close()→err 清理后 os.Remove 返回；replaceFileAtomically(tmpName, path)→return err。
// 任一写失败→close+Remove 清理暂存文件后返回错误。
func writeVolumeIndexCheckpointSectionsContext(ctx context.Context, path string, header [64]byte, nodes []byte, sorted []byte, namePool []byte) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".data-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if err := writeAllContext(ctx, tmp, header[:]); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := writeAllContext(ctx, tmp, nodes); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := writeAllContext(ctx, tmp, sorted); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := writeAllContext(ctx, tmp, namePool); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return replaceFileAtomically(tmpName, path)
}

// [S-sig 0x140807140, 416B]：缓存 map 命中→返回；entry → 递归 parent → filepath.join 缓存。
// 体待日志待定域专项还原。
func (s *volumeIndexJournalPendingState) resolvePath(a, b interface{}) string {
	_, _, _ = s, a, b
	return ""
}

// baseEntryCountLocked 返回索引基础条目计数（持锁，缓存 field(+0x40)）。
// [S-sig 0x1407e8120, 448B]：field(+0x40) 非空→返回；读视图(+0x520/+0x528) 空→0；
// volumeIndexMappedReadProvider.Acquire → 计数。体待读视图域专项还原。
func (v *VolumeIndex) baseEntryCountLocked() int64 {
	_ = v
	return 0
}

// acquireReadView 获取索引读视图（RWMutex.RLock → acquireReadViewLocked）。
// [S-sig 0x1407e5ac0, 448B]：RLock(+0x10) → acquireReadViewLocked → 组装读视图。
// 体待读视图域专项还原。
func (v *VolumeIndex) acquireReadView() interface{} {
	_ = v
	return nil
}

// runtimeSnapshot 返回索引运行时快照（RLock → activeEntryCountLocked → 字段拷贝）。
// [S-sig 0x1407e78e0, 448B]：RLock → activeEntryCountLocked → 字段(+0x90/+0x98/+0xa0/+0xa8/+0xb0/+0xb8/+0x80/+0x88) 拷贝。体待快照域专项还原。
func (v *VolumeIndex) runtimeSnapshot() interface{} {
	_ = v
	return nil
}

// ShouldCheckpointOverlay 判定是否应做 overlay 检查点（RLock → overlayStatsLocked）。
// [S-sig 0x1407e7dc0, 448B]：RLock → overlayStatsLocked → 计数比较 → bool。
// 体待 overlay 统计域专项还原。
func (v *VolumeIndex) ShouldCheckpointOverlay(a interface{}, b float64) bool {
	_, _, _ = v, a, b
	return false
}

// volumeIndexBaseNodeDescendsFrom 判定基础节点是否从某节点派生（nodeAtIndex 遍历父链）。
// [S-sig 0x140804720, 448B]：遍历 nodeAtIndex → 父链匹配 nodeIndex → bool。
// 体待节点域专项还原。
func volumeIndexBaseNodeDescendsFrom(a, b interface{}, nodeIndex int) bool {
	_, _, _ = a, b, nodeIndex
	return false
}

// IsSourcePath 判定路径是否等于映射读提供器的源路径（abs+Clean+EqualFold）。
// [S-sig 0x1407e48a0, 448B]：TrimSpace 空→false；lock → 源路径(+0xe8/+0xf0) → abs+Clean →
// EqualFold(path abs+Clean)。体待源路径域专项还原。
func (p *volumeIndexMappedReadProvider) IsSourcePath(path string) bool {
	_, _ = p, path
	return false
}

// applyVolumeIndexTombstonesToHeapSnapshot 将墓碑应用到堆快照（标记 flag bit 2）。
// [S-sig 0x1408073a0, 448B]：遍历 heap 快照 → tombstone map 命中 → 标记 flag(+0x16) bit2 →
// 返回 tombstone 计数。体待墓碑域专项还原。
func applyVolumeIndexTombstonesToHeapSnapshot(a interface{}, b interface{}, c, d interface{}) int {
	_, _, _, _ = a, b, c, d
	return 0
}

// usnFollowerPaths 计算 USN 跟随者路径对（volumeIndexPath → 拼接）。
// [S-sig 0x140818280, 448B]：volumeIndexPath×2 → 尾段匹配 → concatstring2 拼接。
// 体待 USN 域专项还原。
func (s *FileIndexService) usnFollowerPaths(a, b interface{}) (string, string) {
	_, _, _ = s, a, b
	return "", ""
}

// Acquire 获取堆读视图（字段拷贝组装）。
// [S-sig 0x140a05100, 448B]：nil→panic；字段(+0x20/+0x28/+0x38/+0x40/+0x48/+0x68/+0x70/+0x78/+0x50/+0x58/+0x60) 拷贝。体待堆读视图域专项还原。
func (p *volumeIndexHeapReadProvider) Acquire() interface{} {
	_ = p
	return nil
}

// closeMappedReadProviderForReplaceLocked 关闭映射读提供器以替换（持锁，IsSourcePath → HeapSnapshot → close）。
// [S-sig 0x1407f5be0, 480B]：TrimSpace(path) 空→return；IsSourcePath→匹配→HeapSnapshot →
// clearNameTrigramIndexLocked → closeVolumeIndexReadProvider。体待替换域专项还原。
func closeMappedReadProviderForReplaceLocked(v *VolumeIndex, path string) {
	_, _ = v, path
}

// validateVolumePinyinSections 校验卷拼音 sections 边界（段边界 + ASCII 校验）。
// [S-sig 0x14080c220, 480B]：遍历 sections 校验边界 + 段字符 ASCII → 非法→error。
// 体待拼音 sections 域专项还原。
func validateVolumePinyinSections(a, b, c, d, e, f, g interface{}) error {
	_, _, _, _, _, _, _ = a, b, c, d, e, f, g
	return nil
}
