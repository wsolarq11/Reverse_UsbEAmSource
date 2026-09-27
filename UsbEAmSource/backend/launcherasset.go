// AUTO-RECONSTRUCTED SERVICE METHODS — DOMAIN: launcherasset (Batch A)
// 研究用途
//
// 契约来源：
//   - 方法签名：redress types all -m -v 对目标 exe 提取（真实签名，非臆造）
//   - 结构体字段：backend/types_launcher.go L454-482（与 all_types.txt L23183 一致）
//   - 行号蓝图：docs/goresym/source_funcs.txt File: launcherasset.go（L1912-1941）
//
// 还原口径（与交接一致）：可编译 + 功能一致（非字节级同哈希）。
//
// 方法体档位如实标注：
//
//	[S] 签名确证(redress) + 函数体经目标 exe 反汇编实证（gore 地址 + capstone）
//	[P] 签名确证 + 函数体据字段语义推断，标记依据；无对应反汇编（疑内联/待译）
//	[T] 签名确证 + 函数体反汇编后仍待逐条翻译（锁/IO/HTTP/多分支）
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// normalizeLauncherAssetNamespace 将 namespace 归一化为 URL 安全路径段。
// [汇编实证 0x140873ca0] 语义：按 "/" 切分；逐段过滤 "."(len==1) 与 ".."(len==2)；
// 其余每段经 pathEscapeLauncherAssetSegment 转义后以 "/" 重新 join。
// 空输入返回空串（0x140873d71）。[S 依据汇编实证]
func normalizeLauncherAssetNamespace(ns string) string {
	if ns == "" {
		return ""
	}
	parts := strings.Split(ns, "/")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "." || p == ".." {
			continue
		}
		out = append(out, pathEscapeLauncherAssetSegment(p))
	}
	return strings.Join(out, "/")
}

// pathEscapeLauncherAssetSegment 把 URL 资产段内的保留字符替换为 '-'。
// [汇编实证 0x140874200] 常量表（.rodata，RIP 相对逐条解引用）共 5 对、全 new 为 '-':
//
//	' '->'-'  '#'->'-'  '?'->'-'  '&'->'-'  '%'->'-'
//
// 由 10 个单字节 string 交替作为 old,new 输入 strings.NewReplacer 构造。[S 依据汇编实证]
func pathEscapeLauncherAssetSegment(seg string) string {
	return strings.NewReplacer(" ", "-", "#", "-", "?", "-", "&", "-", "%", "-").Replace(seg)
}

// buildLauncherAssetURL 构建资产 URL：/__usbeam_asset__/<ns>/<id>?v=<version>。
// [汇编实证 0x140873b00] 参数 (rax/rbx=ns, rcx/rdi=id, rsi=version int64)；
// 调用 normalizeLauncherAssetNamespace(ns) 归一化 ns，经 escape 处理 id；
// 6 段拼接为：
//
//	'/__usbeam_asset__/' + nrm(ns) + '/' + idEsc + '?v=' + FormatInt(version,10)
//
// 常量：'/__usbeam_asset__/'(0x12=18)、'/'(1)、'?v='(3)，均自 .rodata 实证；
// version 经 0x1400ad400（strconv.FormatInt，base 0xa=10）格式化为十进制。[S]
func buildLauncherAssetURL(namespace, id string, version int64) string {
	return "/__usbeam_asset__/" + normalizeLauncherAssetNamespace(namespace) + "/" + id + "?v=" + strconv.FormatInt(version, 10)
}

// isValidLauncherAssetIDHex 校验候选 ID 是否为合法的 launcherAssetID（恰 32 个十六进制字符）。
// [parseLauncherAssetRequest 实证 0x140873a47-0x140873aa4]：循环逐字符校验，达 0x20(32) 判成功，
// 每字符仅接受 '0'-'9'（-0x30<=9）或 'a'-'f'（-0x61<=5）；与 newLauncherAssetID（0x140874400，经
// hex 表 [rip+0x3e0c6e] 生成 32 字符）的形态互洽。[S 依据汇编实证]
func isValidLauncherAssetIDHex(id []byte) bool {
	if len(id) != 32 {
		return false
	}
	for _, c := range id {
		if !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

// newLauncherAssetID 生成随机 32-hex launcherAssetID。
// [汇编实证 0x140874400]：16 字节随机源（0x14025e5a0，Go 随机包装）+ hex 表
// [rip+0x3e0c6e]（'0123456789abcdef'）逐字节查表编码为 32 字符。
// 与 isValidLauncherAssetIDHex 互洽（生成值恒通过该校验）。[S 依据汇编实证]
func newLauncherAssetID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic("launcherasset: random source unavailable: " + err.Error())
	}
	return hex.EncodeToString(buf[:])
}

// stableLauncherAssetID 对 (a, b string, c []byte) 三段做 SHA-256 并取前 16 字节编码为 32 小写 hex。
// [Ghidra 反编译精确（0x140873f80）] C 伪代码实证签名 3 参（2 string + 1 []byte，stacktrace struct{8,8};{8,8};{8,8,8}），
// 写序固定：Write(a) → Write(0x00) → Write(b) → Write(0x00) → Write(c)；
// 经 crypto/internal/fips140/sha256.Digest.Reset/Write/Sum 得到 SHA-256 摘要，前 16 字节逐字节 hex
// 查表（[rip+0x3e0c6e]）编为恰 32 字符。[S 依据 Ghidra 反编译]
func stableLauncherAssetID(a, b string, c []byte) string {
	h := sha256.New()
	h.Write([]byte(a))
	h.Write([]byte{0})
	h.Write([]byte(b))
	h.Write([]byte{0})
	h.Write(c)
	return hex.EncodeToString(h.Sum(nil)[:16])
}

// sha256HexPrefix 对输入做 SHA-256 并取其前 16 字节编码为 32 小写 hex ——
// stableLauncherAssetID（0x140873f80）的核心：反汇编实证其经
// crypto/internal/fips140/sha256 的 Digest.Reset/Write/Sum（0x140a0c1e0/0x140a0c2e0/0x140a0c5a0）
// 得到摘要，再逐字节 hex 查表编为恰 32 字符。与 newLauncherAssetID 的随机形态互为对照：
// stable（确定性）= sha256HexPrefix，random（随机）= newLauncherAssetID。[S 依据汇编实证]
func sha256HexPrefix(seed []byte) string {
	sum := sha256.Sum256(seed)
	return hex.EncodeToString(sum[:16])
}

// launcherAssetPrefix 是 launcher 资产请求的统一前缀。
const launcherAssetPrefix = "/__usbeam_asset__/"

// parseLauncherAssetRequest 把 /__usbeam_asset__/ 前缀的请求路径解析为资产标识。
// [Ghidra 反编译精确（0x140873760）] C 伪代码语义：
//   - memequal 前缀(0x12) → 截断（无前缀时按 raw 继续）
//   - strings.Trim(p,"/")、Replace("\\"→"/")、TrimSpace
//   - genSplit("/")；任一空/./.. 段 → 整体失败
//   - 最后段必须恰 32-hex（isValidLauncherAssetIDHex）
//   - namespace = path.Clean(join 前段) 去首尾 '/'
//   - 第 2 参 requestVersion：TrimSpace + strconv.ParseInt(10,64)，须 >0
//   - 返回 (namespace, id string, version int64, ok bool) 四值（48 字节聚合）
//
// [S 汇编 0x140873760]（[G] Ghidra 反编译实证，替换此前近似签名）：
// 前缀剥离 → 路径段校验 → 末段 id hex 校验 → namespace 装配 → version 解析。
func parseLauncherAssetRequest(requestPath, requestVersion string) (namespace, id string, version int64, ok bool) {
	p := requestPath
	if len(p) >= 0x12 && strings.HasPrefix(p, launcherAssetPrefix) {
		p = p[0x12:]
	}
	if p == "" {
		return "", "", 0, false
	}
	s0 := strings.TrimSpace(strings.Trim(strings.Replace(p, "\\", "/", -1), "/"))
	parts := strings.Split(s0, "/")
	for _, seg := range parts {
		if seg == "" || (len(seg) == 1 && seg[0] == '.') || (len(seg) == 2 && seg[0] == '.' && seg[1] == '.') {
			return "", "", 0, false
		}
	}
	if len(parts) < 2 {
		return "", "", 0, false
	}
	id = parts[len(parts)-1]
	if !isValidLauncherAssetIDHex([]byte(id)) {
		return "", "", 0, false
	}
	namespace = strings.Trim(path.Clean("/"+strings.Join(parts[:len(parts)-1], "/")), "/")
	if namespace == "" {
		return "", "", 0, false
	}
	vs := strings.TrimSpace(requestVersion)
	v, err := strconv.ParseInt(vs, 10, 64)
	if err != nil || v <= 0 {
		return "", "", 0, false
	}
	return namespace, id, v, true
}

// currentTime 返回服务时钟当前时间；now 为零值时回退真实时钟。
// [签名确证: func(*launcherAssetService) currentTime() time.Time] [S]
func (s *launcherAssetService) currentTime() time.Time {
	if s.now == nil {
		return time.Now()
	}
	return s.now()
}

// resolvedLimits 返回当前生效预算。
// [S-inline] 字段 s.limits 直取（无独立符号，疑内联；语义确证）。
func (s *launcherAssetService) resolvedLimits() launcherAssetLimits {
	return s.limits
}

// validateItemSize 校验单项字节数不超 maxItemBytes 且非负。
// [签名确证: func(int64) error]
// 依据汇编（0x140872440）：
//
//	rcx=[rax+0x38] 取 limits.maxItemBytes；若 rcx<=0 则默认 0x4000000(64MB)；
//	rbx(size)<=0 → 构造 error（负值）；rbx<=rcx → 通过；否则 → 构造错误（含 size 与限额）。
//
// [S 依据汇编实证]
func (s *launcherAssetService) validateItemSize(size int64) error {
	limit := s.limits.maxItemBytes
	if limit <= 0 {
		limit = 64 << 20 // 0x4000000，汇编 cmovle 默认
	}
	if size <= 0 {
		return errors.New("launcherasset: non-positive item size")
	}
	if size <= limit {
		return nil
	}
	return errors.New("launcherasset: item exceeds maxItemBytes")
}

// Clear 重新分配 entries/namespaceBytes 并清零 totalBytes。
// [S-sig 0x140871fc0]（签名确证）：重新分配 entries(off+8)、namespaceBytes(off+0x10)，
// 置 totalBytes(off+0x18)=0；**未重置 next(off+0x20)**，故保留 next。
func (s *launcherAssetService) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.entries = make(map[string]launcherAssetEntry)
	s.namespaceBytes = make(map[string]int64)
	s.totalBytes = 0
	// 汇编未触及 next —— 保持原值
}

// removeEntryLocked 移除指定 entry 并回退字节计数。
// [S-inline] 符号表 0 命中(幽灵/内联)，删除逻辑被 pruneExpiredLocked/evictOldestLocked/addEntryLocked
// 各自内联(反汇编实证 0x1408720e0/0x140872c00/0x1408728c0 均含相同 inline 删除体)。此处提取为 helper
// 以便共享；totalBytes 减法后 asm 含 clamp-to-0 逻辑（jge 0x140872350 / 0x1408730a4），防止竞态漂移。
func (s *launcherAssetService) removeEntryLocked(id string) bool {
	e, ok := s.entries[id]
	if !ok {
		return false
	}
	delete(s.entries, id)
	s.totalBytes -= e.size
	if s.totalBytes < 0 {
		s.totalBytes = 0
	}
	if nb, ok := s.namespaceBytes[e.namespace]; ok {
		if nb <= e.size {
			delete(s.namespaceBytes, e.namespace)
		} else {
			s.namespaceBytes[e.namespace] = nb - e.size
		}
	}
	return true
}

// evictOldestLocked 在 namespaceFilter 限定的范围内驱逐 accessedAt 最老的条目，返回是否执行。
// [S] 反汇编实证 0x140872c00（310L asm）。遍历 entries map；namespaceFilter 非空时按 e.namespace
// 筛选（先比长度[cmp r10,r13]再做 memequal 0x140006280，不同则跳过）；命中集合内取 accessedAt
// 最小（时间比较 0x1400fa040/0x1400fa1c0），并列时以 version 更小者胜（cmp [rsp+0x168],rcx / jge
// 保留旧值）。找到后内联删除体（delete + totalBytes clamp-to-0 + namespaceBytes 递减/兜底删除）。
// 无候选返回 false。栈布局校验：entry 副本基线 [rsp+0x148]，namespace@+0x10=[0x158]、
// version@+0x20=[0x168]、accessedAt@+0x80=[0x1c8] —— 三处与 asm 使用点零误差吻合。
func (s *launcherAssetService) evictOldestLocked(namespaceFilter string) bool {
	var victim string
	var oldest time.Time
	var oldestVersion int64
	have := false
	for id, e := range s.entries {
		if namespaceFilter != "" && e.namespace != namespaceFilter {
			continue
		}
		if have && !e.accessedAt.Before(oldest) {
			if !e.accessedAt.Equal(oldest) || e.version >= oldestVersion {
				continue
			}
		}
		have = true
		oldest = e.accessedAt
		oldestVersion = e.version
		victim = id
	}
	if !have {
		return false
	}
	s.removeEntryLocked(victim)
	return true
}

// pruneExpiredLocked 清除过期的条目（expiresAt 非零且 now.After(expiresAt)）。
// [S] 反汇编实证 0x1408720e0（188L asm）。mapiter 遍历；expiresAt.IsZero 检查(bt 0x3f)；
// 未零则 time.After 判断；过期则内联删除体(delete+totalBytes clamp-to-0+namespaceBytes 递减/删除)。
func (s *launcherAssetService) pruneExpiredLocked(now time.Time) {
	for id, e := range s.entries {
		if !e.expiresAt.IsZero() && now.After(e.expiresAt) {
			s.removeEntryLocked(id)
		}
	}
}

// ---- 以下为仅落签名/占位骨架（T 档：函数体待还原，签名 100% 来自 redress） ----

// RegisterBytes 对外注册入口，把内存数据注册为 launcher 资产并返回引用。
// [signature确证 func(string,string,[]uint8,int64) (launcherAssetRef,error)]
// 反汇编调用链：RegisterBytes → register（0x14086e463 call 0x14086f680），故本入口仅做参数校验
// 后委托内部 register 完成锁内装配（newID 查重 / ensureCapacityLocked / next++ / entry TTL=600s /
// addEntryLocked / buildURL(next)），语义与 register 组件对齐，避免平行重复。[S 依据 Ghidra 反编译]
func (s *launcherAssetService) RegisterBytes(namespace, id string, data []uint8, version int64) (launcherAssetRef, error) {
	if s == nil {
		return launcherAssetRef{}, errors.New("launcherasset: nil service")
	}
	if err := s.validateItemSize(int64(len(data))); err != nil {
		return launcherAssetRef{}, err
	}
	return s.register(namespace, id, data, version)
}

// RegisterFile 把磁盘文件注册为资产。
// [S] 反汇编实证 0x14086ed00（448L asm）。
// 实证流程：
//
//	normalizeLauncherAssetNamespace → TrimSpace(id) → os.OpenFile(path, O_RDONLY, 0)
//	→ File.Stat → IsDir(rax.fun[4]) → Size ≤ 0 → Size > maxItemBytes → validateItemSize
//	→ duffcopy 构造面参 → register(ns, id, data, version)
//
// 注意：RegisterFile 直接 OpenFile 读取完整内容后走 register，不经过 readFileBounded。
// data 为 os.OpenFile → Stat → 校验通过后 io.ReadAll 读入（asm 无显式 ReadAll，由 register
// 接收 []byte 完成——filePath 不存入 entry）。
func (s *launcherAssetService) RegisterFile(namespace, id, path string, version int64) (launcherAssetRef, error) {
	if s == nil {
		return launcherAssetRef{}, errors.New("launcherasset: nil service")
	}
	ns := normalizeLauncherAssetNamespace(namespace)
	id = strings.TrimSpace(id)
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return launcherAssetRef{}, err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return launcherAssetRef{}, err
	}
	if fi.IsDir() {
		return launcherAssetRef{}, errors.New("launcherasset: path is a directory")
	}
	size := fi.Size()
	if size <= 0 {
		return launcherAssetRef{}, errors.New("launcherasset: file is empty")
	}
	if err := s.validateItemSize(size); err != nil {
		return launcherAssetRef{}, err
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return launcherAssetRef{}, err
	}
	return s.register(ns, id, data, version)
}

// RegisterStableBytes 稳定注册：id 空时由 stableLauncherAssetID(namespace, TrimSpace(id), data[:16]) 派生。
// [S] 反汇编实证 0x14086e760（249L asm）。
// 实证流程：
//
//	validateItemSize(len(data)) → normalize → id 为空时 mallocgc 造 []byte(namespace) 作为 seed →
//	stableLauncherAssetID(namespace, TrimSpace(id), [16]byte) → duffcopy →
//	registerStable(namespace, id, version)
//
// 注意：RegisterStableBytes **不经过** RegisterBytes/register——直接走 registerStable。
func (s *launcherAssetService) RegisterStableBytes(namespace, id string, data []uint8, version int64) (launcherAssetRef, error) {
	if s == nil {
		return launcherAssetRef{}, errors.New("launcherasset: nil service")
	}
	if err := s.validateItemSize(int64(len(data))); err != nil {
		return launcherAssetRef{}, err
	}
	ns := normalizeLauncherAssetNamespace(namespace)
	if id == "" {
		// asm 造 []byte(ns) → stableLauncherAssetID(ns, id, [16]byte)
		id = stableLauncherAssetID(ns, strings.TrimSpace(id), []byte(ns))
	}
	return s.registerStable(ns, id, version)
}

// ReadBytes 读取资产数据。
// [S] 反汇编实证 0x140871220, 261L。完整控制流：
//   - TrimSpace(assetURL) → url.Parse → Query.Get("v") → parseLauncherAssetRequest(path, v)
//   - normalizeLauncherAssetNamespace(expectedNamespace) 若非空则与解析出的 ns 比对(memequal)
//   - lookup(ns, id) → 命中 → 锁后读 entries[id]；data 非空返 data/ct；否则 readFileBounded 回退
//
// 锁后读模式（HANDOFF §4.5）：lookup 内持锁校验+刷新 accessedAt 后释放，本函数在临界区外读 entries。
func (s *launcherAssetService) ReadBytes(assetURL, expectedNamespace string) ([]byte, string, error) {
	if s == nil {
		return nil, "", errors.New("launcherasset: nil service")
	}
	u, err := url.Parse(strings.TrimSpace(assetURL))
	if err != nil || u == nil {
		return nil, "", errors.New("launcherasset: invalid asset url")
	}
	v := u.Query().Get("v")
	ns, id, _, ok := parseLauncherAssetRequest(u.Path, v)
	if !ok {
		return nil, "", errors.New("launcherasset: invalid asset url")
	}
	if expectedNamespace != "" {
		expectNS := normalizeLauncherAssetNamespace(expectedNamespace)
		if expectNS != "" && expectNS != ns {
			return nil, "", errors.New("资源命名空间不匹配")
		}
	}
	if !s.lookup(ns, id) {
		return nil, "", errors.New("资源不存在或已过期")
	}
	e, ok := s.entries[id]
	if !ok {
		return nil, "", errors.New("资源不存在或已过期")
	}
	if len(e.data) > 0 {
		return e.data, e.contentType, nil
	}
	data, err := s.readFileBounded(e.filePath)
	if err != nil {
		return nil, "", fmt.Errorf("读取资源文件失败: %w", err)
	}
	return data, "application/octet-stream", nil
}

// Exists 判断资产是否存在。
// [S] 反汇编实证 0x1408716e0, 142L。完整控制流：
//   - TrimSpace → url.Parse → Query.Get("v") → parseLauncherAssetRequest
//   - normalizeLauncherAssetNamespace(expectedNamespace) 非空比对
//   - lookup → 返回 bool
//
// 锁由 lookup 内部管理（锁后读模式）。
func (s *launcherAssetService) Exists(assetURL, expectedNamespace string) bool {
	if s == nil {
		return false
	}
	u, err := url.Parse(strings.TrimSpace(assetURL))
	if err != nil || u == nil {
		return false
	}
	v := u.Query().Get("v")
	ns, id, _, ok := parseLauncherAssetRequest(u.Path, v)
	if !ok {
		return false
	}
	if expectedNamespace != "" {
		expectNS := normalizeLauncherAssetNamespace(expectedNamespace)
		if expectNS != "" && expectNS != ns {
			return false
		}
	}
	return s.lookup(ns, id)
}

// ServeAssetRequest 处理资产 HTTP 请求：解析 /__usbeam_asset__/<ns>/<id> 路径（parse），
// lookup 查命中后写 200 + Content-Type（若有）与 body。返回 true 表示已处理。
// [S] 反汇编实证 0x1408706c0, 536L。独立于 ReadBytes：直接 lookup → w.WriteHeader/w.Write.
// 锁由 lookup 内部管理（锁后读模式）。
func (s *launcherAssetService) ServeAssetRequest(w http.ResponseWriter, r *http.Request) bool {
	if s == nil || w == nil || r == nil || r.URL == nil {
		return false
	}
	v := r.URL.Query().Get("v")
	ns, id, _, ok := parseLauncherAssetRequest(r.URL.Path, v)
	if !ok {
		http.Error(w, "invalid asset path", http.StatusNotFound)
		return true
	}
	if !s.lookup(ns, id) {
		http.Error(w, "asset not found", http.StatusNotFound)
		return true
	}
	e, ok := s.entries[id]
	if !ok {
		http.Error(w, "asset not found", http.StatusNotFound)
		return true
	}
	if e.contentType != "" {
		w.Header().Set("Content-Type", e.contentType)
	}
	w.WriteHeader(http.StatusOK)
	if len(e.data) > 0 {
		_, _ = w.Write(e.data)
		return true
	}
	// 空 data → 文件回退。
	file, _, err := s.openFileBounded(e.filePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return true
	}
	defer file.Close()
	_, _ = io.Copy(w, file)
	return true
}

// addEntryLocked 把 entry 写入 entries（key=id）并维护 namespaceBytes/totalBytes。
// [S] 反汇编实证 0x1408728c0（177L asm）。懒建两个 map（runtime.makemap_small）；
// entry 已存在(覆盖)时先回退旧值的 totalBytes/namespaceBytes（clamp-to-0）；
// 再写 entries[id]=e（mapassign_faststr）、namespaceBytes[ns]+=size、totalBytes+=size。
func (s *launcherAssetService) addEntryLocked(e launcherAssetEntry) {
	if s.entries == nil {
		s.entries = make(map[string]launcherAssetEntry)
	}
	if s.namespaceBytes == nil {
		s.namespaceBytes = make(map[string]int64)
	}
	if old, exists := s.entries[e.id]; exists {
		s.totalBytes -= old.size
		if s.totalBytes < 0 {
			s.totalBytes = 0
		}
		if nb, ok := s.namespaceBytes[old.namespace]; ok {
			if nb <= old.size {
				delete(s.namespaceBytes, old.namespace)
			} else {
				s.namespaceBytes[old.namespace] = nb - old.size
			}
		}
	}
	s.entries[e.id] = e
	s.namespaceBytes[e.namespace] += e.size
	s.totalBytes += e.size
}

// register 内部注册：把内存数据 nums 注册为 launcher 资产并返回引用。
// [Ghidra 反编译精确（0x14086f680）] C 伪代码语义；与 registerStable 同构但内容为"数据注册"且 id 自动生成：
//
//	LOCK → currentTime → pruneExpiredLocked
//	→ id 为空时循环 newLauncherAssetID + mapaccess2 查重至唯一（显式 id 直接使用）
//	→ ensureCapacityLocked（容量校验失败 → (ref,err)）
//	→ 成功路径 next++ → 构造 entry（createdAt=accessedAt=now；expiresAt=now+TTL，TTL 缺失默认 600s）
//	→ addEntryLocked → buildLauncherAssetURL(namespace, id, next) → (launcherAssetRef, nil)
//	返回 72 字节聚合 = (launcherAssetRef, error)。[S 依据 Ghidra 反编译]
func (s *launcherAssetService) register(namespace, id string, data []byte, version int64) (launcherAssetRef, error) {
	if s == nil {
		return launcherAssetRef{}, errors.New("launcherasset: nil service")
	}
	now := s.currentTime()
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pruneExpiredLocked(now)
	if id == "" {
		// register C：自动生成 id 时 mapaccess2 查重，直至唯一；显式 id 不重生成（保稳定注册确定性）。
		for {
			id = newLauncherAssetID()
			if _, exists := s.entries[id]; !exists {
				break
			}
		}
	}
	ns := normalizeLauncherAssetNamespace(namespace)
	if err := s.ensureCapacityLocked(ns, int64(len(data))); err != nil {
		return launcherAssetRef{}, err
	}
	s.next++
	entry := launcherAssetEntry{
		id:         id,
		namespace:  ns,
		version:    version,
		data:       data,
		size:       int64(len(data)),
		createdAt:  now,
		accessedAt: now,
		expiresAt:  now.Add(600 * time.Second), // register C：TTL 默认 600_000_000_000ns = 10min
	}
	s.addEntryLocked(entry)
	url := buildLauncherAssetURL(namespace, id, s.next) // register C：buildURL 第三参 = next 序号
	return launcherAssetRef{ID: id, URL: url, Version: version}, nil
}

// ensureCapacityLocked 校验并逐条驱逐，确保条目数/总字节预算能容纳 itemSize 的新条目。
// [反汇编实证 0x140872540]：
//   - limits 默认值：maxEntries<=0→0x1000(4096)、maxItemBytes<=0→0x4000000(64MB)、
//     maxNamespaceBytes<=0→0x10000000(256MB)、maxTotalBytes<=0→0x20000000(512MB)（各自 cmovle）。
//   - itemSize(rdi)<=0 或 >maxItemBytes(rsi) → error(0x1408726c2)；itemSize>maxNamespaceBytes(r8)
//     → error(0x14087265e)；itemSize>maxTotalBytes(r9) → error(0x1408725fa)。
//   - 三层顺序驱逐循环（非交织），各自失败即返回 error：
//     ① namespaceBytes[ns]+itemSize > maxNamespace → evictOldestLocked(ns)   [0x14087272b]
//     ② totalBytes+itemSize > maxTotal          → evictOldestLocked("")    [0x1408727c2]
//     ③ len(entries) >= maxEntries              → evictOldestLocked("")    [0x140872812]
//   - 全通过 → 返回 (nil) [0x140872877 xor eax,eax]。
//
// 签名实证：Go ABI (rax=s, rbx/rcx=string, rdi=int64)；0x140872751 把 string 参数并入
// mapaccess1 查 s.namespaceBytes(off 0x10)，故第一参语义为 namespace（非 id/exceptID）。
// [S] 反汇编实证 0x140872540（212L asm）。
func (s *launcherAssetService) ensureCapacityLocked(namespace string, itemSize int64) error {
	maxEntries := s.limits.maxEntries
	if maxEntries <= 0 {
		maxEntries = 0x1000
	}
	maxItem := s.limits.maxItemBytes
	if maxItem <= 0 {
		maxItem = 0x4000000
	}
	maxNamespace := s.limits.maxNamespaceBytes
	if maxNamespace <= 0 {
		maxNamespace = 0x10000000
	}
	maxTotal := s.limits.maxTotalBytes
	if maxTotal <= 0 {
		maxTotal = 0x20000000
	}
	if itemSize <= 0 || itemSize > maxItem {
		return errors.New("launcherasset: item exceeds maxItemBytes")
	}
	if itemSize > maxNamespace || itemSize > maxTotal {
		return fmt.Errorf("launcherasset: item %d exceeds capacity", itemSize)
	}
	for s.namespaceBytes[namespace]+itemSize > maxNamespace {
		if !s.evictOldestLocked(namespace) {
			return errors.New("launcherasset: namespace capacity exhausted")
		}
	}
	for s.totalBytes+itemSize > maxTotal {
		if !s.evictOldestLocked("") {
			return errors.New("launcherasset: capacity exhausted")
		}
	}
	for int64(len(s.entries)) >= int64(maxEntries) {
		if !s.evictOldestLocked("") {
			return errors.New("launcherasset: capacity exhausted")
		}
	}
	return nil
}

// lookup 深查资产：若 entries 存在且未过期则视为命中，并刷新 accessedAt。
// [S] 反汇编实证 0x140871940。持锁 → currentTime → map 按 id 读、namespace 匹配、
// 过期判定(expiresAt.IsZero→bt 0x3f→!now.Before→内联删除体)、accessedAt 回写(mapassign_faststr)。
func (s *launcherAssetService) lookup(namespace, id string) bool {
	if s == nil {
		return false
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	e, ok := s.entries[id]
	if !ok {
		return false
	}
	// Ghidra C(0x140871940)：定位后还需 namespace 匹配（memequal 比较）。
	if e.namespace != normalizeLauncherAssetNamespace(namespace) {
		return false
	}
	now := s.currentTime()
	if !e.expiresAt.IsZero() && !now.Before(e.expiresAt) {
		// Ghidra C：过期则删除条目并回退 totalBytes/namespaceBytes 后返回 0。
		s.removeEntryLocked(id)
		return false
	}
	e.accessedAt = now
	s.entries[id] = e
	return true
}

// openFileBounded 打开文件、校验非目录、超限，返回文件对象与大小。
// [S] 反汇编实证 0x1408731a0（170L asm）。
// 签名实证：成功出口 rax=*File, rbx=Size(), rcx=0, rdi=0 → 三返回 (file, size, nil)。
// 调用点 readFileBounded 以 test rcx,rcx 判错，以 rbx 作 LimitReader 上限 → 交叉印证。
// maxBytes 从 s.limits.maxItemBytes 读取（≤0 时 cmovle 0x4000000）。
// 调用链：os.OpenFile → File.Stat → IsDir → Size ≤0 → Size > maxItemBytes → 全部通过则返回。
func (s *launcherAssetService) openFileBounded(path string) (*os.File, int64, error) {
	limit := s.limits.maxItemBytes
	if limit <= 0 {
		limit = 64 << 20 // 0x4000000，asm cmovle 默认
	}
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, 0, err
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, 0, err
	}
	if fi.IsDir() {
		f.Close()
		return nil, 0, errors.New("launcherasset: path is a directory")
	}
	size := fi.Size()
	if size <= 0 {
		f.Close()
		return nil, 0, errors.New("launcherasset: file is empty")
	}
	if size > limit {
		f.Close()
		return nil, 0, fmt.Errorf("launcherasset: file %d bytes exceeds bounded %d", size, limit)
	}
	return f, size, nil
}

// readFileBounded 带上限读取文件；先调 openFileBounded 打开校验，再用 LimitReader 读。
// [S] 反汇编实证 0x1408733c0（157L asm）。
// 帧首即调 openFileBounded(0x1408731a0)；成功路径 defer file.Close()（有 deferwrap1 0x140873700）；
// io.LimitReader(file, size+1) → io.ReadAll → ([]byte, error) 返回。
// 调用链：openFileBounded → defer file.Close → LimitReader(size+1) → ReadAll。
func (s *launcherAssetService) readFileBounded(path string) ([]byte, error) {
	file, size, err := s.openFileBounded(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// asm 中 LimitReader 第二个参数为 size+1（计至文件尾后一字节，使 ReadAll 读完整文件）
	r := io.LimitReader(file, size+1)
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// registerStable 内部稳定注册。
// [S] 反汇编实证 0x14086fd20, 404L。完整控制流：
//   - 传入 id 须恰 32-hex（循环校验 0-9/a-f，0x14086fe48 前，无效即构造 error 返回）
//   - 持锁 + currentTime + pruneExpiredLocked
//   - entries[id] 已存在且 ns/id 匹配（memequal）→ 仅刷新 expiresAt（now+TTL）/accessedAt，返回既有 URL
//   - 不存在 → ensureCapacityLocked(namespace, version) 检查 → next++ → 构造 entry（expiresAt=now+TTL）
//     → addEntryLocked → buildURL
//   - TTL 默认 600_000_000_000ns = 10 分钟（next 为空时同值 cmov）
//
// 栈槽 0x328 itemSize 确证为第三参 version（汇编 0x14087020c mov rdi,[rsp+0x328]）。
func (s *launcherAssetService) registerStable(namespace, id string, version int64) (launcherAssetRef, error) {
	if s == nil {
		return launcherAssetRef{}, errors.New("launcherasset: nil service")
	}
	if !isValidLauncherAssetIDHex([]byte(id)) {
		return launcherAssetRef{}, errors.New("launcherasset: invalid stable asset id")
	}
	now := s.currentTime()
	ttl := 600 * time.Second // 10 分钟默认过期
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pruneExpiredLocked(now)
	ns := normalizeLauncherAssetNamespace(namespace)
	if e, ok := s.entries[id]; ok && e.namespace == ns {
		// 已注册：刷新过期/访问时间并复用既有 URL。
		e.expiresAt = now.Add(ttl)
		e.accessedAt = now
		s.entries[id] = e
		return launcherAssetRef{ID: id, URL: buildLauncherAssetURL(namespace, id, version), Version: version}, nil
	}
	// 条目不存在或 ns 不匹配：先验容量再注册。
	if err := s.ensureCapacityLocked(namespace, version); err != nil {
		return launcherAssetRef{}, err
	}
	s.next++
	entry := launcherAssetEntry{
		id:         id,
		namespace:  ns,
		version:    version,
		size:       0,
		createdAt:  now,
		accessedAt: now,
		expiresAt:  now.Add(ttl),
	}
	s.addEntryLocked(entry)
	return launcherAssetRef{ID: id, URL: buildLauncherAssetURL(namespace, id, version), Version: version}, nil
}
