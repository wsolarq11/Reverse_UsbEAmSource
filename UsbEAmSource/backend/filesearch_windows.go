// AUTO-RECONSTRUCTED — DOMAIN: filesearch runtime (filesearch_windows.go)
// Source: UsbEAm_Launcher 1.0.3 (Go 1.25.12, PE64), disassembled from
// main.filesearch_* / main.(*FileIndexService).* symbols.
// Tier markers:
//
//	[S VA]      — body fully translated from asm (trivial pure fn).
//	[S-sig VA]  — signature proven from asm (prologue arg regs + epilogue ret regs);
//	              body is a faithful zero-value skeleton.
//	[P]         — signature not proven; zero-value skeleton, blocker noted inline.
//
// 研究用途
package main

import (
	"context"
	"strings"
	"time"

	"golang.org/x/text/collate"
)

// ---- 未落地 receiver 类型（all_types.txt 未提取，此处最小落地） ----

// fileSearchFieldSet 字段投影集合。includes(0x140852400) 以 map 句柄直查（mapaccess），
// 实证为 map 类型；nil 集合视为「包含一切」。
type fileSearchFieldSet map[string]struct{}

// fileSearchMatchCollectors 按 MatchKind 分桶（0..8）收集候选的 9 槽数组。
// Release(0x140822860) 以 stride 72(0x48) 遍历 0..8，元素 = limitedFileSearchCandidateCollector。
type fileSearchMatchCollectors [9]limitedFileSearchCandidateCollector

// fileSearchJournalBudgetController 控制日志应用预算；RecordApply(0x14084dbc0) 直读 [0x0]
// 单 int64 字段（钳制 [0x40000,0x400000] 后写回）。字段名待实证。
type fileSearchJournalBudgetController struct {
	budgetBytes int64
}

// ---- perf / memory 诊断 ----

// [S-sig 0x14081f5a0] 是否启用文件搜索性能打点（读全局开关，返回 bool）。
func isFileSearchPerfEnabled() bool {
	return false
}

// [S-sig 0x14081f600] 打点 perf 日志（isFileSearchPerfEnabled 门控）。
// 序言 morestack 保存 rax/rbx/rcx/rdi/rsi = 5 字 = string(kind:rax,rbx) + string(query:rcx,rdi) + int(count:rsi)；
// 体：concatstring2 拼常量(19B)+kind，写 []interface{} 后调 log.Logger.output(level=2)。返回 void。
// 旧签名多出的 (d time.Duration, err error) 无寄存器对应，已订正。
func logFileSearchPerf(kind string, query string, count int) {
}

// [S-sig 0x14081f780] 采集进程内存计数（fileSearchProcessMemoryCounters，值返回）。
func collectFileSearchMemoryDiagnostics() fileSearchProcessMemoryCounters {
	return fileSearchProcessMemoryCounters{}
}

// [S-sig 0x14081f880] 带缓存的进程内存诊断采集（同返回 fileSearchProcessMemoryCounters）。
func cachedFileSearchMemoryDiagnostics() fileSearchProcessMemoryCounters {
	return fileSearchProcessMemoryCounters{}
}

// [S-sig 0x14081fce0] 读 PROCESS_MEMORY_COUNTERS（GetProcessMemoryInfo）并值返回。
func fileSearchProcessMemoryUsage() fileSearchProcessMemoryCounters {
	return fileSearchProcessMemoryCounters{}
}

// ---- collator / sort comparer ----

// [S-sig 0x14081fda0] 从全局 sync.Pool 取/建 collator，无参返回 *collate.Collator。
func acquireFileSearchCollator() *collate.Collator {
	return nil
}

// [S-sig 0x14081fe80] 释放排序比较器持有的 collator（放回池）。
func (c *fileSearchSortComparer) Release() {
}

// [S-sig 0x1407e13e0] 从 sync.Pool 取/建 nodePathCache 并配 sliceCache 容量。
// 签名经 asm 序言（rax=count int，test jle 判非正）与 Resolve 调用方实证；体未翻译。
func newNodePathCache(count int) *nodePathCache {
	return &nodePathCache{}
}

// [S-sig 0x140809280] 读视图下计算节点计数（含 defer 释放读视图）。返回 int。
func (idx *VolumeIndex) NodeCount() int {
	return 0
}

// [S-sig 0x140808f20] 按节点索引与 FRN 解析路径（含 defer）。参数 nodeIndex int32、
// frn uint64、cache *nodePathCache；返回 (path string, nodeIndex int32)。签名经
// Resolve 调用方寄存器实证；体未翻译（依赖读视图/名称池反解链）。
func (idx *VolumeIndex) ResolvePathByNodeIndexAndFRN(nodeIndex int32, frn uint64, cache *nodePathCache) (string, int32) {
	return "", 0
}

// [S 汇编实证 0x14081ff00, 416B] 解析候选路径（就地填 Path/NodeIndex，返回 Path）。
// c==nil → ""；pathResolved → Path；Index==nil → 置 pathResolved 返回 Path；否则查
// pathCaches[Index]，缺失则 newNodePathCache(Index.NodeCount()) 建缓存并写回；计时
// ResolvePathByNodeIndexAndFRN(NodeIndex,FRN,cache) → Path/NodeIndex；pathResolveDuration
// += time.Since(start)；pathResolved=true；返回 Path。
func (r *fileSearchCandidatePathResolver) Resolve(c *scoredFileSearchCandidate) string {
	if c == nil {
		return ""
	}
	if c.pathResolved {
		return c.Path
	}
	if c.Index == nil {
		c.pathResolved = true
		return c.Path
	}
	cache := r.pathCaches[c.Index]
	if cache == nil {
		cache = newNodePathCache(c.Index.NodeCount())
		r.pathCaches[c.Index] = cache
	}
	start := time.Now()
	path, nodeIndex := c.Index.ResolvePathByNodeIndexAndFRN(c.NodeIndex, c.FRN, cache)
	c.Path = path
	c.NodeIndex = nodeIndex
	r.pathResolveDuration += int64(time.Since(start))
	c.pathResolved = true
	return c.Path
}

// [S 汇编实证 0x1408200a0, 224B] 释放路径解析器缓存：r==nil → 返回；否则遍历
// pathCaches，逐项 returnNodePathCache(cache) 后 delete(pathCaches, idx)。
func (r *fileSearchCandidatePathResolver) Release() {
	if r == nil {
		return
	}
	for idx := range r.pathCaches {
		returnNodePathCache(r.pathCaches[idx])
		delete(r.pathCaches, idx)
	}
}

// [S-sig 0x1407e1560] 将 nodePathCache 放回 sync.Pool（Reset 后归还）。体未翻译。
func returnNodePathCache(cache *nodePathCache) {
}

// Get 按键读取缓存路径。key < 0x40000000 走 sliceCache 索引；>= 0x40000000 走 deltaCache map。
// [S 汇编 0x1407e1720, 192B]：cmp ebx,0x40000000（signed）分路；slice 路径 movsxd 符号扩展，
// 负键/越界返回 ("",false)，空字符串亦判不存在；map 路径 mapaccess2_fast32 直接透传 (v, ok)。
func (c *nodePathCache) Get(key int32) (string, bool) {
	if key >= 0x40000000 {
		v, ok := c.deltaCache[key]
		return v, ok
	}
	if key < 0 || key >= int32(len(c.sliceCache)) {
		return "", false
	}
	v := c.sliceCache[key]
	if v == "" {
		return "", false
	}
	return v, true
}

// Set 按键写入缓存路径。key < 0x40000000 写 sliceCache 索引；>= 0x40000000 写 deltaCache map。
// [S 汇编 0x1407e17e0, 259B]：map 路径 nil 则 makemap_small 建表后 mapassign_fast32；
// slice 路径 movsxd 符号扩展，负键/越界直接忽略返回；两路均带写屏障。
func (c *nodePathCache) Set(key int32, value string) {
	if key >= 0x40000000 {
		if c.deltaCache == nil {
			c.deltaCache = make(map[int32]string)
		}
		c.deltaCache[key] = value
		return
	}
	if key < 0 || key >= int32(len(c.sliceCache)) {
		return
	}
	c.sliceCache[key] = value
}

// Reset 重置缓存到指定容量：清空 sliceCache；容量不足则扩容 make([]string,n)，
// 否则截断长度；清空 deltaCache map。
// [S 汇编 0x1407e1620, 256B]：nil 返回；len!=0 则 memclrHasPointers(sliceCache)；
// n > cap 则 makeslice([]string,n)；len=n；deltaCache(+0x18) 非空则 mapclear。
func (c *nodePathCache) Reset(n int) {
	if c == nil {
		return
	}
	clear(c.sliceCache)
	if n > cap(c.sliceCache) {
		c.sliceCache = make([]string, n)
	} else {
		c.sliceCache = c.sliceCache[:n]
	}
	clear(c.deltaCache)
}

// [S 汇编实证 0x140820180, 320B] 比较两字符串：TrimSpace 后空值处理（双空=0，a 空=1，
// b 空=-1），否则 collator 非空走 CompareString、空走 strings.Compare；spec.Direction=="desc"
// 时结果取反。返回 int(-1/0/1)。
func (c *fileSearchSortComparer) compareText(a, b string) int {
	ta := strings.TrimSpace(a)
	tb := strings.TrimSpace(b)
	var result int
	switch {
	case len(ta) == 0 && len(tb) == 0:
		result = 0
	case len(ta) == 0:
		result = 1
	case len(tb) == 0:
		result = -1
	case c.collator != nil:
		result = c.collator.CompareString(ta, tb)
	default:
		result = strings.Compare(ta, tb)
	}
	if c.spec.Direction == "desc" {
		result = -result
	}
	return result
}

// [S 汇编实证 0x1408202c0, 864B] 三态比较两候选（按 spec.Key 分派字段优先级）。
// nil：a==b==nil→0；a==nil→1；b==nil→-1。路径经 resolver.Resolve（nil 时用 Path）TrimSpace。
// "path"：path→name→modifiedAt；"modifiedAt"：modifiedAt→name→path；默认：name→path→modifiedAt。
// 字符串比较走 compareText（含 desc），int64 比较（ModifiedAtUnix）独立做 desc 反转。
func (c *fileSearchSortComparer) compareCandidate(a, b *scoredFileSearchCandidate, resolver *fileSearchCandidatePathResolver) int {
	if a == nil {
		if b == nil {
			return 0
		}
		return 1
	}
	if b == nil {
		return -1
	}
	resolvePath := func(cand *scoredFileSearchCandidate) string {
		if resolver != nil {
			return strings.TrimSpace(resolver.Resolve(cand))
		}
		return strings.TrimSpace(cand.Path)
	}
	compareModified := func() int {
		var r int
		switch {
		case a.ModifiedAtUnix > b.ModifiedAtUnix:
			r = 1
		case a.ModifiedAtUnix < b.ModifiedAtUnix:
			r = -1
		}
		if c.spec.Direction == "desc" {
			r = -r
		}
		return r
	}
	switch c.spec.Key {
	case "path":
		if r := c.compareText(resolvePath(a), resolvePath(b)); r != 0 {
			return r
		}
		if r := c.compareText(a.Name, b.Name); r != 0 {
			return r
		}
		return compareModified()
	case "modifiedAt":
		if r := compareModified(); r != 0 {
			return r
		}
		if r := c.compareText(a.Name, b.Name); r != 0 {
			return r
		}
		return c.compareText(resolvePath(a), resolvePath(b))
	default:
		if r := c.compareText(a.Name, b.Name); r != 0 {
			return r
		}
		if r := c.compareText(resolvePath(a), resolvePath(b)); r != 0 {
			return r
		}
		return compareModified()
	}
}

// ---- heap（container/heap 接口） ----

// [S 0x140820680] 排序堆长度。
func (h *fileSearchSortedCandidateHeap) Len() int {
	return len(h.items)
}

// [S-sig 0x1408206a0] 排序堆比较（委托 comparer）。
func (h *fileSearchSortedCandidateHeap) Less(i, j int) bool {
	return false
}

// [S 0x140820780] 排序堆交换。
func (h *fileSearchSortedCandidateHeap) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
}

// [S-sig 0x1408209c0] 排序堆入堆（元素 scoredFileSearchCandidate，0x68 宽）。
func (h *fileSearchSortedCandidateHeap) Push(x interface{}) {
}

// [S-sig 0x140820b60] 排序堆出堆（返回 interface{} 装箱元素）。
func (h *fileSearchSortedCandidateHeap) Pop() interface{} {
	return nil
}

// [S 0x140821be0] 候选堆长度。
func (h *fileSearchCandidateHeap) Len() int {
	return len(h.items)
}

// [S-sig 0x140821c00] 候选堆比较（按 Score 降序）。
func (h *fileSearchCandidateHeap) Less(i, j int) bool {
	return false
}

// [S 0x140821cc0] 候选堆交换。
func (h *fileSearchCandidateHeap) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
}

// [S-sig 0x140821e60] 候选堆入堆。
func (h *fileSearchCandidateHeap) Push(x interface{}) {
}

// [S-sig 0x140822000] 候选堆出堆。
func (h *fileSearchCandidateHeap) Pop() interface{} {
	return nil
}

// ---- 收集器 ----

// [S-sig 0x140820c20] 新建带排序比较器的限量候选收集器。
func newLimitedSortedFileSearchCandidateCollector(limit int, comparer *fileSearchSortComparer) *limitedSortedFileSearchCandidateCollector {
	return nil
}

// [S-sig 0x140820f60] 释放限量排序收集器。
func (c *limitedSortedFileSearchCandidateCollector) Release() {
}

// [S-sig 0x1408210c0] 加入候选到限量排序收集器。
func (c *limitedSortedFileSearchCandidateCollector) Add(cand *scoredFileSearchCandidate) {
}

// [S-sig 0x140821340] 返回排序后的命中（[]FileSearchHit，元素 0xb0 宽）。
func (c *limitedSortedFileSearchCandidateCollector) SortedHits() []FileSearchHit {
	return nil
}

// [S-sig 0x140821b00] 按需回填候选修改时间（unix 秒）。
func hydrateFileSearchCandidateModifiedAt(c *scoredFileSearchCandidate) {
}

// [S-sig 0x1408220c0] 释放限量候选收集器。
func (c *limitedFileSearchCandidateCollector) Release() {
}

// [S-sig 0x1408221c0] 加入候选到限量收集器。
func (c *limitedFileSearchCandidateCollector) Add(cand *scoredFileSearchCandidate) {
}

// [S-sig 0x1408223e0] 返回排序命中（内部排序 + 转换，[]FileSearchHit）。
func (c *limitedFileSearchCandidateCollector) sortedHits() []FileSearchHit {
	return nil
}

// [S-sig 0x140822860] 释放 9 槽匹配收集器。
func (m *fileSearchMatchCollectors) Release() {
}

// [S-sig 0x1408228c0] 按 MatchKind 分桶加入轻量匹配。
func (m *fileSearchMatchCollectors) Add(match volumeSearchMatch) {
}

// [S-sig 0x1408229a0] 归并 9 桶排序命中为轻量命中（[]FileSearchHit）。
func (m *fileSearchMatchCollectors) SortedLightweightHits() []FileSearchHit {
	return nil
}

// [S-sig 0x140822bc0] 候选转 FileSearchHit（填 ID/Name/Path/Ext/ModifiedAt/Score）。
func (c *limitedFileSearchCandidateCollector) toFileSearchHit(cand *scoredFileSearchCandidate) FileSearchHit {
	return FileSearchHit{}
}

// ---- 服务构造 / 持久化 ----

// [S-sig 0x140822de0] 新建文件索引服务（装配 bootstrap/锁/映射/chans）。
func NewFileIndexService(bootstrap *BootstrapService) *FileIndexService {
	return nil
}

// [S-sig 0x1408232a0] 持久化循环（订阅信号，串行落盘）。
func (s *FileIndexService) runPersistenceLoop() {
}

// [S-sig 0x140823620] 检查并保存待写索引；返回是否有工作。
func (s *FileIndexService) checkAndSaveIndexes(now time.Time) bool {
	return false
}

// [S-sig 0x1408236a0] 检查并保存待写索引候选。
func (s *FileIndexService) checkAndSavePendingIndexes(now time.Time) {
}

// [S-sig 0x140823720] 检查并保存 overlay 合并；返回是否有工作。
func (s *FileIndexService) checkAndSaveOverlayMerges(now time.Time) bool {
	return false
}

// [S-sig 0x140823880] 检查并保存待办 overlay 合并。
func (s *FileIndexService) checkAndSavePendingOverlayMerges(now time.Time) {
}

// [S-sig 0x140823a20] 保存卷索引候选（checkpoint 落盘）。
func (s *FileIndexService) saveVolumeIndexCandidates(roots []string, now time.Time) {
}

// [S-sig 0x140824ac0] 是否应立即持久化卷索引。
func shouldPersistVolumeIndexNow(meta fileSearchVolumeMeta, now time.Time) bool {
	return false
}

// [S-sig 0x140824c20] 是否应立即合并卷 overlay。
func shouldMergeVolumeIndexOverlayNow(meta fileSearchVolumeMeta, now time.Time) bool {
	return false
}

// [S-sig 0x140824de0] checkpoint 候选成本分级。
func classifyCheckpointCandidateCost(c fileSearchCheckpointCandidate) int {
	return 0
}

// [S-sig 0x140824ee0] overlay 合并候选成本分级。
func classifyOverlayMergeCandidateCost(c fileSearchCheckpointCandidate) int {
	return 0
}

// [S-sig 0x140824f40] checkpoint WAL 大小（字节）。
func fileSearchCheckpointWALSize(c fileSearchCheckpointCandidate) int64 {
	return 0
}

// [S-sig 0x140824fc0] checkpoint 候选优先级。
func prioritizeCheckpointCandidate(c fileSearchCheckpointCandidate) int {
	return 0
}

// [S-sig 0x140825380] overlay 合并候选优先级。
func prioritizeOverlayMergeCandidate(c fileSearchCheckpointCandidate) int {
	return 0
}

// [S-sig 0x140825580] 入队卷索引 checkpoint。
func (s *FileIndexService) queueVolumeIndexCheckpoint(root string, c fileSearchCheckpointCandidate) {
}

// [S-sig 0x140825700] 入队卷索引 overlay 合并。
func (s *FileIndexService) queueVolumeIndexOverlayMerge(root string, c fileSearchCheckpointCandidate) {
}

// [S-sig 0x140825880] 快照 checkpoint 候选来源（root + *VolumeIndex）。
func (s *FileIndexService) snapshotCheckpointCandidateSources() []fileSearchCheckpointCandidateSource {
	return nil
}

// [S-sig 0x140825c60] 收集 checkpoint 候选。
func (s *FileIndexService) collectCheckpointCandidates(sources []fileSearchCheckpointCandidateSource, now time.Time) []fileSearchCheckpointCandidate {
	return nil
}

// [S-sig 0x140826220] 收集 overlay 合并候选。
func (s *FileIndexService) collectOverlayMergeCandidates(sources []fileSearchCheckpointCandidateSource, now time.Time) []fileSearchCheckpointCandidate {
	return nil
}

// [S-sig 0x140826a40] 排空卷索引 checkpoint 候选（返回 root 列表）。
func (s *FileIndexService) drainVolumeIndexCheckpointCandidates(now time.Time) []string {
	return nil
}

// [S-sig 0x140826aa0] 排空卷索引 overlay 合并候选（返回 root 列表）。
func (s *FileIndexService) drainVolumeIndexOverlayMergeCandidates(now time.Time) []string {
	return nil
}

// [S-sig 0x140826b00] 调度持久化延续（信号/定时）。
func (s *FileIndexService) schedulePersistenceContinuation(now time.Time) {
}

// [S-sig 0x140826c20] checkpoint 是否可重试。
func (s *FileIndexService) checkpointRetryReady(root string, now time.Time) bool {
	return false
}

// [S-sig 0x140826de0] overlay 合并是否可重试。
func (s *FileIndexService) overlayMergeRetryReady(root string, now time.Time) bool {
	return false
}

// [S-sig 0x140826fa0] 标记 checkpoint 尝试。
func (s *FileIndexService) markCheckpointAttempt(root string, now time.Time) {
}

// [S-sig 0x140827280] 标记 checkpoint 成功。
func (s *FileIndexService) markCheckpointSuccess(root string, now time.Time) {
}

// [S-sig 0x1408276a0] 标记 checkpoint 失败。
func (s *FileIndexService) markCheckpointFailure(root string, err error) {
}

// [S-sig 0x1408279e0] 标记 overlay 合并尝试。
func (s *FileIndexService) markOverlayMergeAttempt(root string, now time.Time) {
}

// [S-sig 0x140827cc0] 标记 overlay 合并成功。
func (s *FileIndexService) markOverlayMergeSuccess(root string, now time.Time) {
}

// [S-sig 0x1408280e0] 标记 overlay 合并失败。
func (s *FileIndexService) markOverlayMergeFailure(root string, err error) {
}

// [S-sig 0x140828420] 按需请求 overlay 合并。
func (s *FileIndexService) requestOverlayMergeIfNeeded(root string) {
}

// ---- 静态 mmap 重载 ----

// [S-sig 0x140828560] 入队静态 mmap 重载。
func (s *FileIndexService) queueStaticMmapReload(root string) {
}

// [S-sig 0x140828740] 锁内判定是否应执行静态 mmap 重载。
func (s *FileIndexService) shouldRunStaticMmapReloadLocked(root string, now time.Time) bool {
	return false
}

// [S-sig 0x140828cc0] 下一次 restatic 唤醒时间。
func (s *FileIndexService) nextRestaticWakeLocked(now time.Time) time.Time {
	return time.Time{}
}

// [S-sig 0x140829300] 标记 restatic 尝试。
func (s *FileIndexService) markRestaticAttempt(root string, now time.Time) {
}

// [S-sig 0x140829520] 标记 restatic 推迟。
func (s *FileIndexService) markRestaticDeferred(root string, now time.Time) {
}

// [S-sig 0x140829680] 标记 restatic 成功。
func (s *FileIndexService) markRestaticSuccess(root string, now time.Time) {
}

// [S-sig 0x1408298c0] 标记 restatic 失败。
func (s *FileIndexService) markRestaticFailure(root string, err error) {
}

// [S-sig 0x140829b40] 停止可选定时器。
func stopOptionalTimer(t *time.Timer) {
}

// [S-sig 0x140829c20] 重置可选定时器。
func resetOptionalTimer(t *time.Timer, d time.Duration) {
}

// [S-sig 0x140829da0] 重排 restatic 定时器。
func (s *FileIndexService) rescheduleRestaticTimer(root string) {
}

// [S-sig 0x140829f00] 锁内修剪 restatic 状态。
func (s *FileIndexService) pruneRestaticStateLocked(now time.Time) {
}

// [S-sig 0x14082a160] 检查并重载静态 mmap 索引。
func (s *FileIndexService) checkAndReloadStaticMmapIndexes(now time.Time) {
}

// [S-sig 0x14082a6e0] 是否应 checkpoint 卷索引 WAL。
func shouldCheckpointVolumeIndexWAL(meta fileSearchVolumeMeta) bool {
	return false
}

// [S-sig 0x14082a780] 卷索引 WAL 是否存在。
func volumeIndexWALExists(path string) bool {
	return false
}

// ---- 运行时卸载 ----

// [S-sig 0x14082a800] 是否应卸载文件搜索运行时。
func shouldUnloadFileSearchRuntime(meta fileSearchVolumeMeta, now time.Time) bool {
	return false
}

// [S-sig 0x14082a920] 按资源模式返回空闲卸载延迟。
func fileSearchIdleUnloadDelayForMode(mode string) time.Duration {
	return 0
}

// [S-sig 0x14082a9c0] 该资源模式是否使用 mmap 静态 provider。
func shouldUseVolumeIndexMmapStaticProviderForResourceMode(mode string) bool {
	return false
}

// [S-sig 0x14082aa60] 检查并卸载空闲运行时；返回是否卸载。
func (s *FileIndexService) checkAndUnloadIdleRuntime(now time.Time) bool {
	return false
}

// [S-sig 0x14082afc0] 检查并卸载内存压力运行时。
func (s *FileIndexService) checkAndUnloadMemoryPressureRuntime(now time.Time) {
}

// [S-sig 0x14082b720] 查询系统内存状态（Win32 GlobalMemoryStatusEx）。
func queryFileSearchSystemMemoryStatus() fileSearchProcessMemoryCounters {
	return fileSearchProcessMemoryCounters{}
}

// ---- 服务生命周期 ----

// [S-sig 0x14082b840] 服务启动（装配运行时/观测者）。
func (s *FileIndexService) ServiceStartup() {
}

// [S-sig 0x14082ba00] 服务关闭（停观测者/持久化/释放）。
func (s *FileIndexService) ServiceShutdown() {
}

// [S-sig 0x14082be80] 释放运行时资源。
func (s *FileIndexService) ReleaseRuntimeResources() {
}

// [S-sig 0x14082c260] 获取文件搜索状态快照。
func (s *FileIndexService) GetFileSearchState() FileSearchState {
	return FileSearchState{}
}

// [S-sig 0x14082c960] 获取文件搜索状态（可取消）。
func (s *FileIndexService) GetFileSearchStateCancellable(ctx context.Context) (FileSearchState, error) {
	return FileSearchState{}, nil
}

// [S-sig 0x14082d0a0] 刷新卷（可取消）。
func (s *FileIndexService) RefreshVolumesCancellable(ctx context.Context) error {
	return nil
}

// [S-sig 0x14082d480] 预热文件搜索运行时（可取消）。
func (s *FileIndexService) WarmFileSearchRuntimeCancellable(ctx context.Context) error {
	return nil
}

// [S-sig 0x14082d860] 重建文件索引。
func (s *FileIndexService) RebuildFileIndex() {
}

// [S-sig 0x14082dc00] 重建文件索引（可取消）。
func (s *FileIndexService) RebuildFileIndexCancellable(ctx context.Context) error {
	return nil
}

// [S-sig 0x14082dfc0] 搜索已索引文件（非窗口，全量）。
func (s *FileIndexService) SearchIndexedFiles(query string) (FileSearchResult, error) {
	return FileSearchResult{}, nil
}

// [S-sig 0x14082f020] 解析已索引文件路径（可取消）。
func (s *FileIndexService) ResolveIndexedFilePathsCancellable(ctx context.Context, requests []FileSearchPathRequest) ([]FileSearchPathHit, error) {
	return nil, nil
}

// [S-sig 0x14082ff20] 解析文件图标（按路径）。
func (s *FileIndexService) ResolveFileIcon(path string) (string, error) {
	return "", nil
}

// [S-sig 0x14082ff80] 解析文件图标（带扩展名/目录提示）。
func (s *FileIndexService) ResolveFileIconWithHints(path string, extension string, isDirectory bool) (string, error) {
	return "", nil
}

// [S-sig 0x1408301c0] 解析文件元数据。
func (s *FileIndexService) ResolveFileMetadata(id string) (FileSearchHit, error) {
	return FileSearchHit{}, nil
}

// [S-sig 0x140830240] 构建文件元数据命中。
func buildFileMetadataHit(id string, root string, nodeIndex int32) FileSearchHit {
	return FileSearchHit{}
}

// [S-sig 0x1408303e0] 解析文件元数据命中列表。
func resolveFileMetadataHits(requests []FileSearchPathRequest) []FileSearchHit {
	return nil
}

// [S-sig 0x140831380] 合并文件元数据命中。
func mergeFileMetadataHits(a, b []FileSearchHit) []FileSearchHit {
	return nil
}

// [S-sig 0x1408314c0] 按 stat 解析文件元数据命中。
func resolveFileMetadataHitsByStat(requests []FileSearchPathRequest) []FileSearchHit {
	return nil
}

// [S-sig 0x1408317c0] 从目录解析文件元数据命中。
func resolveFileMetadataHitsFromDirectory(requests []FileSearchPathRequest) []FileSearchHit {
	return nil
}

// ---- 上下文 / 快照编排 ----

// [S-sig 0x140832320] 确保已配置上下文。
func (s *FileIndexService) ensureConfiguredContext() error {
	return nil
}

// [S-sig 0x140832480] 确保搜索上下文已配置。
func (s *FileIndexService) ensureConfiguredForSearchContext() error {
	return nil
}

// [S-sig 0x1408325e0] 确保可查询快照（按搜索词）。
func (s *FileIndexService) ensureQueryableSnapshotForSearchTermsContext(terms []string) error {
	return nil
}

// [S-sig 0x140832bc0] 确保文件搜索状态快照上下文。
func (s *FileIndexService) ensureFileSearchStateSnapshotContext() error {
	return nil
}

// [S-sig 0x140833c40] 确保已配置快照上下文。
func (s *FileIndexService) ensureConfiguredSnapshotContext() error {
	return nil
}

// [S-sig 0x1408358a0] 确保按目标配置上下文模式。
func (s *FileIndexService) ensureConfiguredContextModeWithTargets(targets []string) error {
	return nil
}

// [S-sig 0x140837540] 当前启动器配置。
func (s *FileIndexService) currentLauncherConfig() FileSearchConfig {
	return FileSearchConfig{}
}

// [S-sig 0x140837800] 克隆文件搜索目标根。
func cloneFileSearchTargetRoots(roots []string) []string {
	return nil
}

// [S-sig 0x140837940] 合并文件搜索目标根（去重）。
func mergeFileSearchTargetRoots(roots ...[]string) []string {
	return nil
}

// [S-sig 0x140837a80] 当前文件搜索配置（归一化）。
func (s *FileIndexService) currentFileSearchConfig() fileSearchConfigNormalized {
	return fileSearchConfigNormalized{}
}

// [S-sig 0x140837e40] 从索引构造卷元数据。
func (s *FileIndexService) volumeMetaFromIndex(root string, idx *VolumeIndex) fileSearchVolumeMeta {
	return fileSearchVolumeMeta{}
}

// [S-sig 0x140838180] 锁内加载可查询卷快照。
func (s *FileIndexService) loadQueryableVolumeSnapshotsLockedContext(ctx context.Context, targets []string) error {
	return nil
}

// [S-sig 0x14083a800] 锁内应用卷状态。
func (s *FileIndexService) applyVolumeStateLockedContext(ctx context.Context, targets []string) error {
	return nil
}

// [S-sig 0x14083d5a0] 锁内释放运行时资源。
func (s *FileIndexService) releaseRuntimeResourcesLocked() {
}

// [S-sig 0x14083dda0] 锁内持久化索引。
func (s *FileIndexService) persistIndexesLocked() error {
	return nil
}

// [S-sig 0x14083e220] checkpoint 后详细重载静态 mmap 索引。
func (s *FileIndexService) reloadStaticMmapIndexAfterCheckpointDetailed(root string, idx *VolumeIndex) {
}

// [S-sig 0x14083e980] 标记文件搜索已使用。
func (s *FileIndexService) markFileSearchUsed(now time.Time) {
}

// [S-sig 0x14083ea60] 开始文件搜索查询。
func (s *FileIndexService) beginFileSearchQuery() {
}

// [S-sig 0x14083eac0] 结束文件搜索查询。
func (s *FileIndexService) finishFileSearchQuery() {
}

// [S-sig 0x14083eb40] 是否有活跃文件搜索查询。
func (s *FileIndexService) hasActiveFileSearchQuery() bool {
	return false
}

// [S-sig 0x14083ebe0] 最近是否使用过文件搜索。
func (s *FileIndexService) recentlyUsedFileSearch(now time.Time, window time.Duration) bool {
	return false
}

// [S-sig 0x14083ed20] 文件搜索最近使用时间。
func fileSearchRecentlyUsedAt() time.Time {
	return time.Time{}
}

// [S-sig 0x14083ee60] 锁内快照（构造 FileSearchState）。
func (s *FileIndexService) snapshotLocked(now time.Time) FileSearchState {
	return FileSearchState{}
}

// [S-sig 0x14083fd00] 聚合拼音统计。
func aggregateFileSearchPinyinStats() int {
	return 0
}

// [S-sig 0x14083fec0] 锁内快照文件搜索卷列表。
func (s *FileIndexService) snapshotFileSearchVolumesLocked() []FileSearchVolume {
	return nil
}

// [S-sig 0x140840260] 查询目标运行时状态。
func (s *FileIndexService) queryRuntimeStatusForTargets(targets []string) bool {
	return false
}

// [S-sig 0x140840820] 搜索用活跃索引。
func (s *FileIndexService) activeIndexesForSearch(targets []string) []*VolumeIndex {
	return nil
}

// [S-sig 0x140840c00] 聚合卷索引 overlay 统计。
func aggregateVolumeIndexOverlayStats() volumeIndexOverlayStats {
	return volumeIndexOverlayStats{}
}

// [S-sig 0x140840e00] checkpoint 运行时快照（锁内）。
func (s *FileIndexService) checkpointRuntimeSnapshotLocked(now time.Time) {
}

// [S-sig 0x140841080] merge 运行时快照（锁内）。
func (s *FileIndexService) mergeRuntimeSnapshotLocked(now time.Time) {
}

// [S-sig 0x140841300] restatic 运行时快照（锁内）。
func (s *FileIndexService) restaticRuntimeSnapshotLocked(now time.Time) {
}

// [S-sig 0x140841580] journal 运行时快照（锁内）。
func (s *FileIndexService) journalRuntimeSnapshotLocked(root string) {
}

// [S-sig 0x140841b80] 静态 mmap 是否活跃。
func fileSearchStaticMmapActive(meta fileSearchVolumeMeta) bool {
	return false
}

// [S-sig 0x140841c60] 锁内开始预热运行时。
func (s *FileIndexService) beginWarmRuntimeLocked(targets []string) {
}

// [S-sig 0x140842140] 锁内标记预热卷完成。
func (s *FileIndexService) markWarmRuntimeVolumeDoneLocked(root string) {
}

// ---- 维护 / 拼音 ----

// [S-sig 0x140842240] 调度文件搜索维护（实现 fileSearchRuntimeWarmer）。
func (s *FileIndexService) scheduleFileSearchMaintenance(targets map[string]struct{}, all bool) {
}

// [S-sig 0x140842280] 按模式调度文件搜索维护。
func (s *FileIndexService) scheduleFileSearchMaintenanceMode(mode string) {
}

// [S-sig 0x140842500] 文件搜索维护循环。
func (s *FileIndexService) runFileSearchMaintenanceLoop() {
}

// [S-sig 0x140842620] 取文件搜索维护工作；返回是否有工作。
func (s *FileIndexService) takeFileSearchMaintenanceWork() bool {
	return false
}

// [S-sig 0x1408427c0] 单次文件搜索维护。
func (s *FileIndexService) runFileSearchMaintenanceOnce() {
}

// [S-sig 0x140842c60] 为查询准备拼音能力。
func (s *FileIndexService) prepareFileSearchPinyinCapabilityForQuery() {
}

// [S-sig 0x140843100] 同步文件搜索拼音能力。
func (s *FileIndexService) syncFileSearchPinyinCapability() {
}

// [S-sig 0x140844080] 拼音工作循环。
func (s *FileIndexService) runFileSearchPinyinWorker() {
}

// [S-sig 0x140844bc0] 锁内缓存快照。
func (s *FileIndexService) cacheSnapshotLocked(snap FileSearchState) {
}

// [S-sig 0x140844ce0] 设置快照缓存。
func (s *FileIndexService) setSnapshotCache(snap FileSearchState) {
}

// [S-sig 0x140844e40] 缓存的进度快照。
func (s *FileIndexService) cachedProgressSnapshot() FileSearchState {
	return FileSearchState{}
}

// ---- 签名 ----

// [S-sig 0x1408450a0] 文件搜索签名（配置+卷的指纹）。
func fileSearchSignature() string {
	return ""
}

// [S-sig 0x1408452c0] 文件搜索配置签名。
func fileSearchConfigSignature() string {
	return ""
}

// ---- NTFS / USN ----

// [S-sig 0x140845760] 发现 NTFS 卷（返回卷根列表）。
func discoverNTFSVolumes() []string {
	return nil
}

// [S-sig 0x140845d00] 查询卷信息（GetVolumeInformation）。
func queryVolumeInformation(root string) (string, error) {
	return "", nil
}

// [S-sig 0x140845e00] 枚举卷条目（USN 遍历）。
func enumerateVolumeEntriesContext(ctx context.Context, root string, idx *VolumeIndex) error {
	return nil
}

// [S-sig 0x140846b40] 打开卷句柄。
func openVolumeHandle(root string) (uintptr, error) {
	return 0, nil
}

// [S-sig 0x140846be0] 查询或创建 USN 日志。
func queryOrCreateUSNJournal(handle uintptr) error {
	return nil
}

// [S-sig 0x140846ec0] 查询卷日志状态。
func queryVolumeJournalState(handle uintptr) uint64 {
	return 0
}

// [S-sig 0x140847120] 获取根引用号（FRN）。
func getRootReferenceNumber(handle uintptr) (uint64, error) {
	return 0, nil
}

// ---- 词项 / 排序 / 打分 ----

// [S-sig 0x140847360] 拆分文件搜索词项。
func splitFileSearchTerms(query string, exact bool) []string {
	return nil
}

// [S-sig 0x140847500] 拆分文件搜索查询字段值。
func splitFileSearchQueryFieldValues(query string, exact bool) []string {
	return nil
}

// [S-sig 0x140847740] 从字段值构建搜索词项。
func buildFileSearchTermsFromValues(values []string) []string {
	return nil
}

// [S-sig 0x140847ca0] 从词项提取目标根。
func fileSearchTargetRootsFromTerms(terms []string) []string {
	return nil
}

// [S-sig 0x140847e20] 比较候选排名（分数/名称/FRN）。
func compareFileSearchCandidateRank(a, b *scoredFileSearchCandidate) int {
	return 0
}

// [S-sig 0x140848020] 取文件扩展名（目录返回空；从末尾扫 \/. 并剥前导点）。
func fileSearchHitExtension(name string, isDir bool) string {
	return ""
}

// [S-sig 0x1408480e0] 构建打分候选。
func buildScoredFileSearchCandidate(root string, idx *VolumeIndex, nodeIndex int32, name string, frn uint64, isDir bool, modifiedAt int64, score float64) scoredFileSearchCandidate {
	return scoredFileSearchCandidate{}
}

// [S-sig 0x140848680] 构建文件搜索条目 ID。
func buildFileSearchEntryID(root string, frn uint64) string {
	return ""
}

// ---- 观测者 / journal ----

// [S-sig 0x1408488c0] 锁内同步观测者。
func (s *FileIndexService) syncWatchersLocked(targets []string) {
}

// [S-sig 0x140849040] 锁内按目标同步观测者。
func (s *FileIndexService) syncWatchersLockedForTargets(targets []string) {
}

// [S-sig 0x140849460] 锁内停止卷观测者。
func (s *FileIndexService) stopVolumeWatcherLocked(root string) {
}

// [S-sig 0x140849580] 解析文件搜索条目 FRN（"root:frn" 或裸 hex）。
func parseFileSearchEntryFRN(id string) (uint64, bool) {
	return 0, false
}

// [S-sig 0x140849660] 锁内开始 journal 转换。
func (s *FileIndexService) beginJournalTransitionLocked(root string) {
}

// [S-sig 0x140849740] 锁内结束 journal 转换。
func (s *FileIndexService) finishJournalTransitionLocked(root string) {
}

// [S-sig 0x140849840] 锁内停止卷观测者并等待。
func (s *FileIndexService) stopVolumeWatcherAndWaitLocked(root string) {
}

// [S-sig 0x140849b20] 锁内启动卷观测者。
func (s *FileIndexService) startVolumeWatcherLocked(root string) {
}

// [S-sig 0x14084a160] 卷观测者循环。
func (s *FileIndexService) runVolumeWatcher(root string) {
}

// [S-sig 0x14084b040] 卷日志应用循环。
func (s *FileIndexService) runVolumeJournalApplier(root string) {
}

// [S-sig 0x14084c520] 记录卷日志读取。
func (s *FileIndexService) recordVolumeJournalRead(root string, nextUSN int64) {
}

// [S-sig 0x14084c940] 记录卷日志游标。
func (s *FileIndexService) recordVolumeJournalCursor(root string) {
}

// [S-sig 0x14084ce20] 更新卷日志运行时。
func (s *FileIndexService) updateVolumeJournalRuntime(root string) {
}

// [S-sig 0x14084d440] 清空卷日志运行时。
func (s *FileIndexService) clearVolumeJournalRuntime(root string) {
}

// [S-sig 0x14084d820] 待处理日志应用延迟。
func (s *FileIndexService) fileSearchJournalApplyDelayForPending(pending int) time.Duration {
	return 0
}

// [S-sig 0x14084d8c0] 待处理日志应用延迟（按时间）。
func (s *FileIndexService) fileSearchJournalApplyDelayForPendingAt(pending int, since time.Time) time.Duration {
	return 0
}

// [S-sig 0x14084da20] 是否应冲刷文件搜索日志待处理。
func (s *FileIndexService) shouldFlushFileSearchJournalPending(state fileSearchJournalRuntimeState) bool {
	return false
}

// [S-sig 0x14084dbc0] 记录一次应用（更新预算）。
func (c *fileSearchJournalBudgetController) RecordApply(pendingBytes int64, applyCount int64, recovering bool) {
}

// [S-sig 0x14084dc60] 调整日志应用预算（钳制 [0x40000,0x400000]）。
func adjustFileSearchJournalApplyBudgetBytes(budget int64, pendingBytes int64, threshold int64, recovering bool) int64 {
	return 0
}

// [S-sig 0x14084dd40] 待处理日志应用预算字节。
func (s *FileIndexService) fileSearchJournalApplyBudgetBytesForPending(pending int) int64 {
	return 0
}

// [S-sig 0x14084de20] 记录卷观测者错误。
func (s *FileIndexService) recordVolumeWatcherError(root string, err error) {
}

// [S-sig 0x14084e0c0] 调度卷日志恢复。
func (s *FileIndexService) scheduleVolumeJournalRecovery(root string) {
}

// [S-sig 0x14084e7c0] 是否为可恢复的卷日志错误。
func isRecoverableVolumeJournalError(err error) bool {
	return false
}

// [S-sig 0x14084e8c0] 收集相关卷日志变更。
func collectRelevantVolumeJournalChanges(records []uint8, ignore fileSearchIgnoreMatcher) []volumeIndexJournalChange {
	return nil
}

// [S-sig 0x14084f560] 读取卷日志批次。
func readVolumeJournalBatch(handle uintptr, buf []uint8) fileSearchJournalBatch {
	return fileSearchJournalBatch{}
}

// [S-sig 0x14084f760] 合并文件搜索日志变更（去重/合并）。
func coalesceFileSearchJournalChanges(changes []volumeIndexJournalChange) []volumeIndexJournalChange {
	return nil
}

// [S-sig 0x14084fc20] 应用卷日志批次。
func (s *FileIndexService) applyVolumeJournalBatches(root string, batches []fileSearchJournalBatch) {
}

// [S-sig 0x140850d40] 锁内重算聚合状态。
func (s *FileIndexService) recalculateAggregateStateLocked(root string) {
}

// [S-sig 0x140850f40] 锁内同步卷运行时统计。
func (s *FileIndexService) syncVolumeRuntimeStatsLocked(root string) {
}

// [S-sig 0x1408511c0] 锁内更新卷元数据。
func (s *FileIndexService) updateVolumeMetaLocked(root string, meta fileSearchVolumeMeta) {
}

// [S-sig 0x140851400] 文件搜索索引文件大小。
func fileSearchIndexFileSize(path string) int64 {
	return 0
}

// [S-sig 0x140851480] 卷索引路径。
func (s *FileIndexService) volumeIndexPath(root string) string {
	return ""
}

// [S-sig 0x140851560] 卷索引路径列表。
func (s *FileIndexService) volumeIndexPaths(root string) []string {
	return nil
}

// [S-sig 0x140851660] 卷索引名（根路径 → 索引名）。
func volumeIndexName(root string) string {
	return ""
}

// [S-sig 0x1408516e0] 锁内修剪卷索引。
func (s *FileIndexService) pruneVolumeIndexesLocked() {
}

// [S-sig 0x140851d20] 锁内清理遗留缓存。
func (s *FileIndexService) cleanupLegacyCachesLocked() {
}

// ---- 窗口 / 格式化 / 归一化 ----

// [S-sig 0x140852000] 格式化 unix 时间（0 → 空串；返回 string）。
func formatUnixTime(t uint32) string {
	return ""
}

// [S-sig 0x140852060] 构建文件搜索窗口结果。
func buildFileSearchWindowResult(hits []FileSearchHit, totalMatchCount int, offset int, limit int, nextOffset int, hasMore bool) FileSearchWindowResult {
	return FileSearchWindowResult{}
}

// [S-sig 0x140852320] 归一化文件搜索字段（→ fileSearchFieldSet）。
func normalizeFileSearchFields(fields []string) fileSearchFieldSet {
	return nil
}

// [S-sig 0x140852400] 字段集合是否包含某字段（nil 集合视为全含）。
func (s fileSearchFieldSet) includes(field string) bool {
	return false
}

// [S-sig 0x140852480] 归一化文件搜索排序提示。
func normalizeFileSearchSortHint(hint string) string {
	return ""
}

// [S-sig 0x140852580] 归一化文件搜索排序方向。
func normalizeFileSearchSortDirection(dir string) string {
	return ""
}

// [S-sig 0x140852600] 归一化文件搜索排序规格。
func normalizeFileSearchSortSpec(key string, dir string) fileSearchSortSpec {
	return fileSearchSortSpec{}
}

// [S-sig 0x1408526a0] 是否应解析文件搜索窗口路径。
func shouldResolveFileSearchWindowPaths(fields fileSearchFieldSet) bool {
	return false
}

// [S-sig 0x140852740] 应用已解析文件搜索路径。
func applyResolvedFileSearchPaths(hits []FileSearchHit, paths map[string]string) []FileSearchHit {
	return nil
}

// [S-sig 0x140852fe0] 应用文件搜索字段投影。
func applyFileSearchFieldProjection(hits []FileSearchHit, fields fileSearchFieldSet) []FileSearchHit {
	return nil
}

// [S-sig 0x140853340] 水合文件搜索窗口路径。
func (s *FileIndexService) hydrateFileSearchWindowPaths(hits []FileSearchHit) {
}

// [S 0x140853840] 取 context 错误（nil context → nil）。
func contextErr(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

// [S-sig 0x1408538e0] 锁上下文（配置锁 + 快照锁）。
func (s *FileIndexService) lockContext(ctx context.Context) error {
	return nil
}

// [S-sig 0x140853b80] 文件搜索索引搜索并发度。
func fileSearchIndexSearchConcurrency() int {
	return 0
}

// [S-sig 0x140853c00] 并行搜索各索引候选。
func searchFileSearchIndexesCandidatesParallel(ctx context.Context, indexes []*VolumeIndex, terms []string) []fileSearchIndexCandidateSearchResult {
	return nil
}

// ---- 可取消搜索 API（前端 Wails 绑定面） ----

// [S-sig 0x140854720] 搜索已索引文件（可取消）。
func (s *FileIndexService) SearchIndexedFilesCancellable(ctx context.Context, query string) (FileSearchResult, error) {
	return FileSearchResult{}, nil
}

// [S-sig 0x140854860] 带类型匹配器搜索已索引文件。
func (s *FileIndexService) searchIndexedFilesWithTypeMatcher(ctx context.Context, query string) (FileSearchResult, error) {
	return FileSearchResult{}, nil
}

// [S-sig 0x140856620] 窗口查询（可取消）。
func (s *FileIndexService) SearchIndexedFilesWindowCancellable(ctx context.Context, query string, offset int, limit int, nodeIndex int, fields []string, sortDirection string) (FileSearchWindowResult, error) {
	return FileSearchWindowResult{}, nil
}

// [S-sig 0x1408567e0] 窗口字段查询（可取消）。
func (s *FileIndexService) SearchIndexedFilesWindowFieldsCancellable(ctx context.Context, query string, offset int, limit int, nodeIndex int, fields []string, sortDirection string) (FileSearchWindowResult, error) {
	return FileSearchWindowResult{}, nil
}

// [S-sig 0x1408569c0] 排序窗口查询（可取消）。
func (s *FileIndexService) SearchIndexedFilesSortedWindowQueryCancellable(ctx context.Context, query string, sortKey string, sortDirection string, offset int, limit int, nodeIndex int, fields []string) (FileSearchWindowResult, error) {
	return FileSearchWindowResult{}, nil
}

// [S-sig 0x140856c00] 过滤排序窗口查询（可取消）。
func (s *FileIndexService) SearchIndexedFilesFilteredSortedWindowQueryCancellable(ctx context.Context, query string, sortKey string, sortDirection string, offset int, limit int, nodeIndex int, fields []string, typeFilters []string) (FileSearchWindowResult, error) {
	return FileSearchWindowResult{}, nil
}

// [S-sig 0x140857040] 带类型匹配器的排序窗口查询。
func (s *FileIndexService) searchIndexedFilesSortedWindowQueryWithTypeMatcher(ctx context.Context, query string, sortKey string, sortDirection string, offset int, limit int, nodeIndex int, fields []string, typeFilters []string) (FileSearchWindowResult, error) {
	return FileSearchWindowResult{}, nil
}

// [S-sig 0x140859400] 窗口查询（可取消）。
func (s *FileIndexService) SearchIndexedFilesWindowQueryCancellable(ctx context.Context, query string, offset int, limit int, nodeIndex int, fields []string, sortDirection string) (FileSearchWindowResult, error) {
	return FileSearchWindowResult{}, nil
}

// [S-sig 0x140859620] 过滤窗口查询（可取消）。
func (s *FileIndexService) SearchIndexedFilesFilteredWindowQueryCancellable(ctx context.Context, query string, offset int, limit int, nodeIndex int, fields []string, sortDirection string, typeFilters []string) (FileSearchWindowResult, error) {
	return FileSearchWindowResult{}, nil
}

// [S-sig 0x140859a40] 带类型匹配器的窗口查询。
func (s *FileIndexService) searchIndexedFilesWindowQueryWithTypeMatcher(ctx context.Context, query string, offset int, limit int, nodeIndex int, fields []string, sortDirection string, typeFilters []string) (FileSearchWindowResult, error) {
	return FileSearchWindowResult{}, nil
}

// [S-sig 0x140859f20] 解析文件元数据（可取消）。
func (s *FileIndexService) ResolveFileMetadataCancellable(ctx context.Context, requests []FileSearchPathRequest) ([]FileSearchHit, error) {
	return nil, nil
}

// nameBigramSignature 计算名称双字签名：逐对字符小写化（A-Z→a-z），去重相邻重复，
// 每对经混合哈希后置位 64 位位图。len<2 返回 0。
// [S 汇编 0x1407e26c0, 256B]：len<2→0；循环 bigram 小写→去重→(h>>8)^h*0x45d9f3b→
// (h>>16)^h→bts 位图。
func nameBigramSignature(s string) uint64 {
	if len(s) < 2 {
		return 0
	}
	var sig uint64
	var last uint16
	hasLast := false
	for i := 0; i+1 < len(s); i++ {
		a := s[i]
		b := s[i+1]
		if a >= 'A' && a <= 'Z' {
			a |= 0x20
		}
		if b >= 'A' && b <= 'Z' {
			b |= 0x20
		}
		bg := uint16(a)<<8 | uint16(b)
		if hasLast && last == bg {
			continue
		}
		h := uint32(bg)
		h = ((h >> 8) ^ h) * 0x45d9f3b
		h = (h >> 16) ^ h
		sig |= 1 << h
		last = bg
		hasLast = true
	}
	return sig
}
