// AUTO-RECONSTRUCTED SERVICE TESTS — DOMAIN: launcherasset (Batch A)
// 研究用途
// 表驱动覆盖 AGENTS 确定性子集：极值边界、非法状态、幂等、并发。
package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

func testAssetService() *launcherAssetService {
	return &launcherAssetService{
		entries:        map[string]launcherAssetEntry{},
		namespaceBytes: map[string]int64{},
		limits: launcherAssetLimits{
			maxEntries:        0, // 0 = 不限制（见 validateItemSize 语义）
			maxItemBytes:      100,
			maxNamespaceBytes: 0,
			maxTotalBytes:     0,
		},
	}
}

func TestValidateItemSize(t *testing.T) {
	s := testAssetService()
	cases := []struct {
		name    string
		size    int64
		wantErr bool
	}{
		{name: "negative", size: -1, wantErr: true},
		{name: "zero", size: 0, wantErr: true}, // 汇编: size<=0 → error
		{name: "within", size: 50, wantErr: false},
		{name: "at-limit", size: 100, wantErr: false},
		{name: "over-limit", size: 101, wantErr: true},
	}
	for _, c := range cases {
		err := s.validateItemSize(c.size)
		if (err != nil) != c.wantErr {
			t.Errorf("%s: validateItemSize(%d) err=%v wantErr=%v", c.name, c.size, err, c.wantErr)
		}
	}

	// 上限 <=0 → 汇编默认 64MB(0x4000000)
	s.limits.maxItemBytes = 0
	if err := s.validateItemSize(1024 * 1024); err != nil {
		t.Errorf("under 64MB default: unexpected err=%v", err)
	}
	if err := s.validateItemSize(64<<20 + 1); err == nil {
		t.Error("over 64MB default must fail")
	}
}

func TestValidateItemSizeDefaultLimit(t *testing.T) {
	s := testAssetService()
	s.limits.maxItemBytes = 0 // 汇编 cmovle → 64MB
	if err := s.validateItemSize((1 << 20)); err != nil {
		t.Errorf("1MB within 64MB default: unexpected err=%v", err)
	}
	if err := s.validateItemSize(65 << 20); err == nil {
		t.Error("65MB over 64MB default must fail")
	}
	s.limits.maxItemBytes = -5 // 负 limit 同样触发 64MB 默认
	if err := s.validateItemSize((1 << 20)); err != nil {
		t.Errorf("negative limit → 64MB default: unexpected err=%v", err)
	}
}

func TestRemoveEntryLocked(t *testing.T) {
	s := testAssetService()
	s.entries["a"] = launcherAssetEntry{namespace: "ns1", size: 30}
	s.namespaceBytes["ns1"] = 30
	s.totalBytes = 30

	if !s.removeEntryLocked("a") {
		t.Fatal("expected existing entry removed")
	}
	if _, ok := s.entries["a"]; ok {
		t.Error("entry still present")
	}
	if s.totalBytes != 0 {
		t.Errorf("totalBytes=%d want 0", s.totalBytes)
	}
	if _, ok := s.namespaceBytes["ns1"]; ok {
		t.Error("namespaceBytes should be dropped when reaching 0")
	}
	// 幂等：再次删除返回 false
	if s.removeEntryLocked("a") {
		t.Error("second remove should return false")
	}
}

func TestEvictOldestLocked(t *testing.T) {
	base := time.Now()
	s := testAssetService()
	s.entries["old"] = launcherAssetEntry{namespace: "ns1", accessedAt: base.Add(-time.Hour)}
	s.entries["new"] = launcherAssetEntry{namespace: "ns1", accessedAt: base}
	s.entries["newer"] = launcherAssetEntry{namespace: "ns2", accessedAt: base.Add(time.Hour)}

	// 空 filter：不限定命名空间 → 驱逐全局最老 old
	if !s.evictOldestLocked("") {
		t.Fatal("expected eviction")
	}
	if _, ok := s.entries["old"]; ok {
		t.Error("oldest should be evicted")
	}

	// filter=ns2：仅在 ns2 内选取 → newer（ns1 的 new 虽更老但因命名空间不符被跳过）
	if !s.evictOldestLocked("ns2") {
		t.Fatal("expected eviction within namespace")
	}
	if _, ok := s.entries["newer"]; ok {
		t.Error("newer should be evicted under ns2 filter")
	}
	if _, ok := s.entries["new"]; !ok {
		t.Error("ns1 entry must survive ns2-filtered eviction")
	}

	// 无匹配命名空间：无可驱逐对象
	if s.evictOldestLocked("nsX") {
		t.Error("no match in namespace must not evict")
	}

	// 空集合：不驱逐
	empty := testAssetService()
	if empty.evictOldestLocked("") {
		t.Error("empty should not evict")
	}
}

func TestEvictOldestLockedTieBreakByVersion(t *testing.T) {
	base := time.Now()
	s := testAssetService()
	// accessedAt 并列时 version 更小者胜（asm: cmp [rsp+0x168],rcx / jge 保留旧值）
	s.entries["v9"] = launcherAssetEntry{accessedAt: base, version: 9}
	s.entries["v3"] = launcherAssetEntry{accessedAt: base, version: 3}

	if !s.evictOldestLocked("") {
		t.Fatal("expected eviction")
	}
	if _, ok := s.entries["v3"]; ok {
		t.Error("smaller version should win the tie-break and be evicted")
	}
	if _, ok := s.entries["v9"]; !ok {
		t.Error("larger version must survive")
	}
}

func TestPruneExpiredLocked(t *testing.T) {
	s := testAssetService()
	now := time.Now()
	s.entries["exp"] = launcherAssetEntry{expiresAt: now.Add(-time.Minute)}
	s.entries["live"] = launcherAssetEntry{expiresAt: now.Add(time.Hour)}
	s.entries["never"] = launcherAssetEntry{expiresAt: time.Time{}}

	s.pruneExpiredLocked(now)
	if _, ok := s.entries["expired"]; ok {
		t.Error("expired entry should be pruned")
	}
	if _, ok := s.entries["live"]; !ok {
		t.Error("live entry must remain")
	}
	if _, ok := s.entries["never"]; !ok {
		t.Error("zero-expiry entry must remain")
	}
}

func TestCurrentTime(t *testing.T) {
	fixed := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	s := &launcherAssetService{now: func() time.Time { return fixed }}
	if got := s.currentTime(); !got.Equal(fixed) {
		t.Errorf("currentTime=%v want %v", got, fixed)
	}
	s2 := &launcherAssetService{} // now nil → 真实时钟
	if got := s2.currentTime(); got.IsZero() {
		t.Error("nil now should fall back to real clock")
	}
}

func TestClearIdempotent(t *testing.T) {
	s := testAssetService()
	s.entries["a"] = launcherAssetEntry{namespace: "n", size: 5}
	s.namespaceBytes["n"] = 5
	s.totalBytes = 5
	s.next = 7

	s.Clear()
	s.Clear() // 幂等二次

	if len(s.entries) != 0 || len(s.namespaceBytes) != 0 || s.totalBytes != 0 {
		t.Errorf("Clear clean: entries=%d ns=%d total=%d", len(s.entries), len(s.namespaceBytes), s.totalBytes)
	}
	// 汇编实证：Clear 不清 next（仅 entries/namespaceBytes/totalBytes）
	if s.next != 7 {
		t.Errorf("next should be preserved, got %d", s.next)
	}
}

// 并发：并发 remove/lookup 不得出现数据竞争（-race 校验下）。
func TestConcurrentRemove(t *testing.T) {
	s := testAssetService()
	for i := 0; i < 100; i++ {
		s.entries[strconv.Itoa(i)] = launcherAssetEntry{namespace: "n", size: 1}
		s.namespaceBytes["n"]++
		s.totalBytes++
	}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			s.lock.Lock()
			s.removeEntryLocked(id)
			s.lock.Unlock()
		}(strconv.Itoa(i))
	}
	wg.Wait()
	if len(s.entries) != 0 {
		t.Errorf("after concurrent remove, %d entries remain", len(s.entries))
	}
}

// ---- 汇编实证新增：normalizeLauncherAssetNamespace / pathEscapeLauncherAssetSegment / buildLauncherAssetURL ----

func TestPathEscapeLauncherAssetSegment(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		// 汇编实证（0x140874200 .rodata 常量表）：' ' '#' '?' '&' '%' → '-'
		{"empty", "", ""},
		{"plain", "abc", "abc"},
		{"space", "a b", "a-b"},
		{"hash", "a#b", "a-b"},
		{"question", "a?b", "a-b"},
		{"amp", "a&b", "a-b"},
		{"percent", "a%b", "a-b"},
		{"slash-kept", "a/b", "a/b"}, // 汇编未替换 '/' 与 '\'
		{"backslash-kept", `a\b`, `a\b`},
		{"mixed", "a b#c?d&e%f", "a-b-c-d-e-f"},
	}
	for _, c := range cases {
		if got := pathEscapeLauncherAssetSegment(c.in); got != c.want {
			t.Errorf("%s: escape(%q)=%q want %q", c.name, c.in, got, c.want)
		}
	}
}

func TestNormalizeLauncherAssetNamespace(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		// 汇编实证（0x140873ca0）：split '/' + 过滤 '.'/'..' + 每段 escape + 再 join '/'
		{"empty", "", ""},
		{"single", "a", "a"},
		{"multi", "a/b/c", "a/b/c"},
		{"dot", "a/./b", "a/b"},
		{"dotdot", "a/../b", "a/b"},
		{"leading-slash", "/a/b", "/a/b"}, // "" 段非 "."/".." 保留（escape("")=""）
		{"escape-each-seg", "a b/c?d", "a-b/c-d"},
		{"all-empty", "///", "///"},
	}
	for _, c := range cases {
		if got := normalizeLauncherAssetNamespace(c.in); got != c.want {
			t.Errorf("%s: normalize(%q)=%q want %q", c.name, c.in, got, c.want)
		}
	}
}

func TestBuildLauncherAssetURL(t *testing.T) {
	// 汇编实证结构（0x140873b00）：'/__usbeam_asset__/' + nrm(ns) + '/' + id + '?v=' + FormatInt(version,10)
	// 第 3 参 version int64（rsi），经 0x1400ad400=strconv.FormatInt（base 0xa=10）格式化。
	cases := []struct {
		name, ns, id string
		version      int64
		want         string
	}{
		{"basic", "a/b", "x", 1, "/__usbeam_asset__/a/b/x?v=1"},
		{"zero-ver", "a", "b", 0, "/__usbeam_asset__/a/b?v=0"},
		{"big-ver", "k", "i", 16777216, "/__usbeam_asset__/k/i?v=16777216"},
		{"empty-ns-normalized", "", "i", 3, "/__usbeam_asset__//i?v=3"},
	}
	for _, c := range cases {
		if got := buildLauncherAssetURL(c.ns, c.id, c.version); got != c.want {
			t.Errorf("%s: buildURL(%q,%q,%d)=%q want %q", c.name, c.ns, c.id, c.version, got, c.want)
		}
	}
}

func TestIsValidLauncherAssetIDHex(t *testing.T) {
	hex32 := "0123456789abcdef0123456789abcdef"
	cases := []struct {
		name string
		id   string
		want bool
	}{
		{"valid-32", hex32, true},
		{"valid-allzero", "00000000000000000000000000000000", true},
		{"valid-max", "ffffffffffffffffffffffffffffffff", true},
		{"short-31", "0123456789abcdef0123456789abcde", false},    // 31 位
		{"long-33", hex32 + "0", false},                           // 33 位
		{"empty", "", false},                                      // 0 位
		{"upper-case", "0123456789ABCDEF0123456789ABCDEF", false}, // parse 仅接受小写 a-f
		{"with-g", "0123456789abcdef0123456789abcdeg", false},
		{"with-slash", "0123456789abcdef0123456789abcde/", false},
		{"with-space", "0123456789abcdef0123456789abcde ", false},
	}
	for _, c := range cases {
		if got := isValidLauncherAssetIDHex([]byte(c.id)); got != c.want {
			t.Errorf("%s: isValidLauncherAssetIDHex(%q)=%v want %v", c.name, c.id, got, c.want)
		}
	}
}

// ---- ensureCapacityLocked：容量预算校验 + 三层顺序驱逐（汇编实证 0x140872540） ----
// 签名实证：(namespace string, itemSize int64) —— 第一参进 namespaceBytes(off 0x10) 查表。

func TestEnsureCapacityDefaultsAndSizeCheck(t *testing.T) {
	s := &launcherAssetService{ // limits 全 0 → 走汇编默认
		entries:        map[string]launcherAssetEntry{},
		namespaceBytes: map[string]int64{},
		limits:         launcherAssetLimits{},
	}
	if err := s.ensureCapacityLocked("", -1); err == nil {
		t.Error("itemSize<=0 must fail")
	}
	if err := s.ensureCapacityLocked("", 0); err == nil {
		t.Error("itemSize==0 must fail")
	}
	// 64MB 默认最大单项
	if err := s.ensureCapacityLocked("", 64<<20); err != nil {
		t.Errorf("64MB at default maxItemBytes should pass: %v", err)
	}
	if err := s.ensureCapacityLocked("", 64<<20+1); err == nil {
		t.Error(">64MB over default maxItemBytes must fail")
	}
	// 空：无驱逐也需通过
	if err := s.ensureCapacityLocked("", 1); err != nil {
		t.Errorf("small item should pass: %v", err)
	}
}

func TestEnsureCapacityEvictsEntries(t *testing.T) {
	base := time.Now()
	s := testAssetService()
	s.limits.maxEntries = 2
	s.entries["a"] = launcherAssetEntry{accessedAt: base.Add(-time.Hour)}
	s.entries["b"] = launcherAssetEntry{accessedAt: base}
	s.entries["c"] = launcherAssetEntry{accessedAt: base.Add(time.Hour)}
	// 3 条 > maxEntries=2 → 驱逐到 2
	if err := s.ensureCapacityLocked("", 1); err != nil {
		t.Fatalf("expect no error after eviction: %v", err)
	}
	if len(s.entries) > 2 {
		t.Errorf("entries=%d want <=2", len(s.entries))
	}
}

func TestEnsureCapacityNamespaceScope(t *testing.T) {
	base := time.Now()
	s := testAssetService()
	s.limits.maxEntries = 10         // 不因条目数驱逐
	s.limits.maxNamespaceBytes = 100 // 命名空间预算触发第一层驱逐
	s.namespaceBytes["ns1"] = 60
	s.totalBytes = 100 // ns1(60) + ns2(40)
	s.entries["a"] = launcherAssetEntry{namespace: "ns1", accessedAt: base.Add(-time.Hour), size: 60}
	s.entries["b"] = launcherAssetEntry{namespace: "ns2", accessedAt: base, size: 40}

	// ns1 占 60，新增 50 → 60+50>100 → 驱逐 ns1 内最老(a)；跨命名空间的 b 不受影响
	if err := s.ensureCapacityLocked("ns1", 50); err != nil {
		t.Fatalf("ensureCapacity under namespace budget should pass: %v", err)
	}
	if _, ok := s.entries["a"]; ok {
		t.Error("oldest within ns1 must be evicted")
	}
	if _, ok := s.entries["b"]; !ok {
		t.Error("entry of another namespace must be kept")
	}
	if s.namespaceBytes["ns1"]+50 > 100 {
		t.Errorf("ns1 bytes=%d +50 exceeds budget 100", s.namespaceBytes["ns1"])
	}
}

func TestEnsureCapacityTotalBytesEvicts(t *testing.T) {
	base := time.Now()
	s := testAssetService()
	s.limits.maxTotalBytes = 100
	s.totalBytes = 90
	s.entries["old"] = launcherAssetEntry{accessedAt: base.Add(-time.Hour), size: 90}
	s.entries["new"] = launcherAssetEntry{accessedAt: base, size: 20}
	// 新增 item 20 → totalBytes 90+20 > 100 → 须驱逐
	if err := s.ensureCapacityLocked("", 20); err != nil {
		t.Fatalf("ensureCapacity should evict to fit: %v", err)
	}
	if s.totalBytes+20 > 100 {
		t.Errorf("totalBytes=%d after ensure, +20 exceeds 100", s.totalBytes)
	}
}

func TestEnsureCapacityExhaustedError(t *testing.T) {
	s := testAssetService()
	s.limits.maxTotalBytes = 100
	s.totalBytes = 100 // 人为构造：总字节已满且无可驱逐条目
	// 第二层驱逐：100+20>100 → evictOldestLocked("") 无候选 → error
	if err := s.ensureCapacityLocked("", 20); err == nil {
		t.Error("unresolvable total budget must return error")
	}
}

func TestEnsureCapacityNamespaceExhaustedError(t *testing.T) {
	s := testAssetService()
	s.limits.maxEntries = 10
	s.limits.maxNamespaceBytes = 50
	s.namespaceBytes["ns1"] = 80 // 超预算但没有 ns1 条目可驱逐
	if err := s.ensureCapacityLocked("ns1", 10); err == nil {
		t.Error("unresolvable namespace budget must return error")
	}
}

func TestNewLauncherAssetID(t *testing.T) {
	// 结构性不变量（随机函数，非确定值断言）：
	// 长度恒 32、恒为合法 hex ID、多次生成不恒相同（随机性）。
	id1 := newLauncherAssetID()
	if len(id1) != 32 {
		t.Errorf("newLauncherAssetID() len=%d want 32", len(id1))
	}
	if !isValidLauncherAssetIDHex([]byte(id1)) {
		t.Errorf("generated id %q failed hex check", id1)
	}
	// 有小数度的 16 字节熵，20 个样本不应全部相同。
	same := true
	prev := id1
	for i := 0; i < 20; i++ {
		id := newLauncherAssetID()
		if len(id) != 32 {
			t.Fatalf("iteration %d len=%d want 32", i, len(id))
		}
		if id != prev {
			same = false
		}
		prev = id
	}
	if same {
		t.Error("20 consecutive newLauncherAssetID() all identical; expected randomness")
	}
}

func TestParseLauncherAssetRequest(t *testing.T) {
	id32 := "0123456789abcdef0123456789abcdef"
	upper := "0123456789ABCDEF0123456789ABCDEF"
	cases := []struct {
		name, path, ver string
		ok              bool
		ns, id          string
		verOK           int64
	}{
		{"no-prefix", "/other/ns/x", "1", false, "", "", 0},
		{"empty-rest", "/__usbeam_asset__/", "1", false, "", "", 0},
		{"single-seg", "/__usbeam_asset__/only", "1", false, "", "", 0},
		{"basic", "/__usbeam_asset__/ns/" + id32, "1", true, "ns", id32, 1},
		{"nested-ns", "/__usbeam_asset__/a/b/" + id32, "1", true, "a/b", id32, 1},
		{"dot-abort", "/__usbeam_asset__/a/./b/" + id32, "1", false, "", "", 0},
		{"dotdot-abort", "/__usbeam_asset__/a/../b/" + id32, "1", false, "", "", 0},
		{"empty-seg-abort", "/__usbeam_asset__//ns/" + id32, "1", true, "ns", id32, 1},
		{"bad-id-short", "/__usbeam_asset__/ns/abcd", "1", false, "", "", 0},
		{"bad-id-nonhex", "/__usbeam_asset__/ns/" + id32[:31] + "g", "1", false, "", "", 0},
		{"bad-id-upper", "/__usbeam_asset__/ns/" + upper, "1", false, "", "", 0},
		{"prefix-not-eq", "/__usbeam_asset__X/ns/" + id32, "1", true, "__usbeam_asset__X/ns", id32, 1},
		{"bad-version", "/__usbeam_asset__/ns/" + id32, "abc", false, "", "", 0},
		{"zero-version", "/__usbeam_asset__/ns/" + id32, "0", false, "", "", 0},
		{"version-ok", "/__usbeam_asset__/ns/" + id32, "42", true, "ns", id32, 42},
		{"nested-ver", "/__usbeam_asset__/a/b/" + id32, "9", true, "a/b", id32, 9},
	}
	for _, c := range cases {
		ns, id, v, ok := parseLauncherAssetRequest(c.path, c.ver)
		if ok != c.ok || ns != c.ns || id != c.id || v != c.verOK {
			t.Errorf("%s: parse(%q,%q)=(%q,%q,%d,%v) want (%q,%q,%d,%v)", c.name, c.path, c.ver, ns, id, v, ok, c.ns, c.id, c.verOK, c.ok)
		}
	}
}

func TestExists(t *testing.T) {
	id32 := "0123456789abcdef0123456789abcdef"
	s := testAssetService()
	s.entries[id32] = launcherAssetEntry{namespace: "ns", id: id32, size: 1}
	s.entries["ffffffffffffffffffffffffffffffff"] = launcherAssetEntry{namespace: "a/b", id: "ffffffffffffffffffffffffffffffff", size: 2}

	cases := []struct {
		name, url, ns string
		want          bool
	}{
		{"match", buildLauncherAssetURL("ns", id32, 1), "", true},
		{"match-with-ns", buildLauncherAssetURL("ns", id32, 1), "ns", true},
		{"match-after-normalize", buildLauncherAssetURL("a/b", "ffffffffffffffffffffffffffffffff", 2), "", true},
		{"norm-no-match-ns", buildLauncherAssetURL("a/b", "ffffffffffffffffffffffffffffffff", 2), "other", false},
		{"ns-mismatch", buildLauncherAssetURL("ns", id32, 1), "other", false},
		{"missing-id", buildLauncherAssetURL("ns", "00000000000000000000000000000000", 1), "", false},
		{"empty-url", "", "", false},
	}
	for _, c := range cases {
		if got := s.Exists(c.url, c.ns); got != c.want {
			t.Errorf("%s: Exists(%q,%q)=%v want %v", c.name, c.url, c.ns, got, c.want)
		}
	}
	// nil receiver
	if got := (*launcherAssetService)(nil).Exists(buildLauncherAssetURL("ns", id32, 1), ""); got {
		t.Error("nil receiver Exists must be false")
	}
}

func TestSha256HexPrefix(t *testing.T) {
	// sha256("abc") = ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
	// 取前 16 字节 hex = 32 个字符（stable ID 形态，SHA-256 前 128 位）。
	if got := sha256HexPrefix([]byte("abc")); got != "ba7816bf8f01cfea414140de5dae2223" {
		t.Errorf("sha256HexPrefix(abc)=%q want ba7816bf8f01cfea414140de5dae2223", got)
	}
	// 确定性：重复调用结果一致。
	if sha256HexPrefix([]byte("x")) != sha256HexPrefix([]byte("x")) {
		t.Error("sha256HexPrefix must be deterministic")
	}
	// 合法 SHAPE：恒为 32 小写 hex（可通过 isValidLauncherAssetIDHex）。
	if !isValidLauncherAssetIDHex([]byte(sha256HexPrefix([]byte("stable-ns")))) {
		t.Error("sha256HexPrefix output must be a valid 32-hex launcherAssetID")
	}
}

func TestAddEntryLocked(t *testing.T) {
	// 新增：entries 写入、ns 计数、total 计数。
	s := testAssetService()
	e1 := launcherAssetEntry{id: "id1", namespace: "ns", size: 10}
	s.addEntryLocked(e1)
	if v, ok := s.entries["id1"]; !ok || v.id != e1.id || v.namespace != e1.namespace || v.size != e1.size {
		t.Errorf("entry not stored or mismatched: %+v", v)
	}
	if s.namespaceBytes["ns"] != 10 {
		t.Errorf("namespaceBytes[ns]=%d want 10", s.namespaceBytes["ns"])
	}
	if s.totalBytes != 10 {
		t.Errorf("totalBytes=%d want 10", s.totalBytes)
	}

	// 同 ns 多条目：namespace 累加。
	s.addEntryLocked(launcherAssetEntry{id: "id2", namespace: "ns", size: 3})
	if s.namespaceBytes["ns"] != 13 {
		t.Errorf("namespaceBytes[ns]=%d want 13", s.namespaceBytes["ns"])
	}
	if s.totalBytes != 13 {
		t.Errorf("totalBytes=%d want 13", s.totalBytes)
	}

	// 覆盖：同 id 换新 size，回退旧值再累加新。
	s.addEntryLocked(launcherAssetEntry{id: "id1", namespace: "ns", size: 7})
	// 旧 id1 10 回退，新 7 累加：13 - 10 + 7 = 10
	if s.totalBytes != 10 {
		t.Errorf("cover totalBytes=%d want 10", s.totalBytes)
	}
	if s.namespaceBytes["ns"] != 10 {
		t.Errorf("cover namespaceBytes[ns]=%d want 10", s.namespaceBytes["ns"])
	}

	// 跨 ns 覆盖：旧 ns 计数回退归零则删除。
	s3 := testAssetService()
	s3.addEntryLocked(launcherAssetEntry{id: "x", namespace: "n1", size: 5})
	s3.addEntryLocked(launcherAssetEntry{id: "x", namespace: "n2", size: 5})
	if s3.totalBytes != 5 {
		t.Errorf("cross-ns cover totalBytes=%d want 5", s3.totalBytes)
	}
	if _, ok := s3.namespaceBytes["n1"]; ok {
		t.Error("n1 should be dropped after reverting to 0")
	}

	// 懒建：nil map 也可 add（makemap_small 语义）。
	raw := &launcherAssetService{}
	raw.addEntryLocked(launcherAssetEntry{id: "y", namespace: "n", size: 1})
	if len(raw.entries) != 1 || raw.totalBytes != 1 {
		t.Errorf("lazy map init failed: entries=%d total=%d", len(raw.entries), raw.totalBytes)
	}
}

func TestReadFileBounded(t *testing.T) {
	s := testAssetService()

	dir := t.TempDir()
	small := dir + "\\small.bin"
	if err := os.WriteFile(small, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	// 小文件 → 读回内容。
	got, err := s.readFileBounded(small)
	if err != nil || string(got) != "hello" {
		t.Errorf("read small file err=%v got=%q want hello", err, got)
	}
	// 不存在 → 报错。
	if _, err := s.readFileBounded(dir + "\\missing.bin"); err == nil {
		t.Error("missing file must error")
	}
}

func TestLookup(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	fixed := func() time.Time { return now }

	// 命中 + accessedAt 刷新。
	s := testAssetService()
	s.now = fixed
	s.entries["id-a"] = launcherAssetEntry{id: "id-a", namespace: "ns", accessedAt: now.Add(-time.Hour)}
	if !s.lookup("ns", "id-a") {
		t.Error("existing unexpired entry should be found")
	}
	if got := s.entries["id-a"].accessedAt; !got.Equal(now) {
		t.Errorf("accessedAt=%v want refreshed %v", got, now)
	}

	// 过期 → 未命中。
	s2 := testAssetService()
	s2.now = fixed
	s2.entries["id-exp"] = launcherAssetEntry{id: "id-exp", namespace: "ns", expiresAt: now.Add(-time.Minute)}
	if s2.lookup("ns", "id-exp") {
		t.Error("expired entry must not be found")
	}

	// 未超期未来 expires → 命中。
	s3 := testAssetService()
	s3.now = fixed
	s3.entries["id-fut"] = launcherAssetEntry{id: "id-fut", namespace: "ns", expiresAt: now.Add(time.Hour)}
	if !s3.lookup("ns", "id-fut") {
		t.Error("future-expiry entry should be found")
	}

	// 缺失 → 未命中。
	if testAssetService().lookup("ns", "zzz0000000000000000000000000000") {
		t.Error("missing id must not be found")
	}

	// nil receiver。
	if (*launcherAssetService)(nil).lookup("ns", "id-a") {
		t.Error("nil receiver lookup must be false")
	}
}

func TestOpenFileBounded(t *testing.T) {
	s := testAssetService()
	// 测试用固定 limits，不依赖默认值。
	// openFileBounded 从 s.limits.maxItemBytes 读限（默认 64MB）。
	dir := t.TempDir()
	fp := dir + "\\data.bin"
	if err := os.WriteFile(fp, []byte("content-here"), 0o644); err != nil {
		t.Fatal(err)
	}

	f, size, err := s.openFileBounded(fp)
	if err != nil {
		t.Fatalf("open small file err=%v", err)
	}
	if size != 12 {
		t.Errorf("size=%d want 12", size)
	}
	_ = f.Close()

	// 缺失 → 报错。
	if _, _, err := s.openFileBounded(dir + "\\missing.bin"); err == nil {
		t.Error("missing open must error")
	}

	// 目录 → IsDir 拒绝。
	baddir := dir + "\\sub"
	if err := os.MkdirAll(baddir, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.openFileBounded(baddir); err == nil {
		t.Error("directory open must error")
	}

	// 空文件 → 拒绝。
	empty := dir + "\\empty.bin"
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.openFileBounded(empty); err == nil {
		t.Error("empty file open must error")
	}

	// 超限 → 拒绝（设 limits.maxItemBytes=4）。
	s.limits.maxItemBytes = 4
	if _, _, err := s.openFileBounded(fp); err == nil {
		t.Error("over-limit open must error")
	}
}

func TestRegisterBytes(t *testing.T) {
	s := testAssetService()

	// 基本注册。
	ref, err := s.RegisterBytes("ns", "id3", []byte("abc"), 7)
	if err != nil {
		t.Fatalf("RegisterBytes err=%v", err)
	}
	if ref.ID != "id3" || ref.Version != 7 {
		t.Errorf("ref=%+v want ID=id3 version=7", ref)
	}
	if want := "/__usbeam_asset__/ns/id3?v=1"; ref.URL != want {
		t.Errorf("ref.URL=%q want %q", ref.URL, want)
	}
	e, ok := s.entries["id3"]
	if !ok || e.size != 3 || string(e.data) != "abc" {
		t.Errorf("entry not recorded: ok=%v entry=%+v", ok, e)
	}
	if s.next != 1 {
		t.Errorf("next=%d want 1", s.next)
	}
	if s.namespaceBytes["ns"] != 3 {
		t.Errorf("namespaceBytes[ns]=%d want 3", s.namespaceBytes["ns"])
	}

	// 空 id → 自动生成合法 32-hex。
	ref2, err := s.RegisterBytes("ns", "", []byte("x"), 1)
	if err != nil {
		t.Fatalf("empty-id err=%v", err)
	}
	if len(ref2.ID) != 32 || !isValidLauncherAssetIDHex([]byte(ref2.ID)) {
		t.Errorf("auto id=%q not 32-hex", ref2.ID)
	}

	// 越限（maxItemBytes=100）→ error。
	if _, err := s.RegisterBytes("ns", "big", make([]byte, 101), 1); err == nil {
		t.Error("over-item-limit must error")
	}

	// nil receiver → error。
	if _, err := (*launcherAssetService)(nil).RegisterBytes("ns", "x", []byte("a"), 1); err == nil {
		t.Error("nil service must error")
	}
}

func TestRegisterFile(t *testing.T) {
	s := testAssetService()
	dir := t.TempDir()
	fp := dir + "\\asset.txt"
	content := []byte("file-content")
	if err := os.WriteFile(fp, content, 0o644); err != nil {
		t.Fatal(err)
	}

	ref, err := s.RegisterFile("ns", "idf", fp, 9)
	if err != nil {
		t.Fatalf("RegisterFile err=%v", err)
	}
	if ref.ID != "idf" || ref.Version != 9 {
		t.Errorf("ref=%+v want ID=idf v=9", ref)
	}
	e, ok := s.entries["idf"]
	if !ok || string(e.data) != string(content) || e.size != int64(len(content)) {
		t.Errorf("entry not recorded: ok=%v data=%q size=%d", ok, string(e.data), e.size)
	}

	// 缺失文件 → error。
	if _, err := s.RegisterFile("ns", "id2", dir+"\\missing.bin", 1); err == nil {
		t.Error("missing file RegisterFile must error")
	}
}

func TestReadBytes(t *testing.T) {
	id32 := "0123456789abcdef0123456789abcdef"
	s := testAssetService()
	if _, err := s.RegisterBytes("ns", id32, []byte("payload"), 5); err != nil {
		t.Fatal(err)
	}

	url := buildLauncherAssetURL("ns", id32, 5)
	data, ct, err := s.ReadBytes(url, "")
	if err != nil || string(data) != "payload" {
		t.Errorf("ReadBytes err=%v data=%q", err, data)
	}
	_ = ct

	// 期望命名空间不匹配 → error。
	if _, _, err := s.ReadBytes(url, "other"); err == nil {
		t.Error("ns mismatch must error")
	}

	// 缺失 id → error。
	missingID := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	url2 := buildLauncherAssetURL("ns", missingID, 5)
	if _, _, err := s.ReadBytes(url2, ""); err == nil {
		t.Error("missing id must error")
	}

	// nil receiver → error。
	if _, _, err := (*launcherAssetService)(nil).ReadBytes(url, ""); err == nil {
		t.Error("nil service must error")
	}
}

func TestRegisterStableBytes(t *testing.T) {
	s := testAssetService()

	// id 空 → 确定性派生 32-hex stable id。
	ref1, err := s.RegisterStableBytes("ns", "", []byte("d"), 1)
	if err != nil {
		t.Fatalf("empty-id stable err=%v", err)
	}
	ref2, err := s.RegisterStableBytes("ns", "", []byte("d2"), 1)
	if err != nil {
		t.Fatalf("empty-id stable#2 err=%v", err)
	}
	if len(ref1.ID) != 32 || !isValidLauncherAssetIDHex([]byte(ref1.ID)) {
		t.Errorf("stable auto id=%q not 32-hex", ref1.ID)
	}
	want := stableLauncherAssetID("ns", "", []byte("ns"))
	if ref1.ID != want || ref2.ID != want {
		t.Errorf("stable ids (%s,%s) want %s", ref1.ID, ref2.ID, want)
	}

	// 给定 32-hex id → registerStable 通过。
	id32 := "0123456789abcdef0123456789abcdef"
	ref3, err := s.RegisterStableBytes("ns", id32, []byte("x"), 2)
	if err != nil || ref3.ID != id32 {
		t.Errorf("explicit-id ref=%+v err=%v", ref3, err)
	}

	// nil receiver → error。
	if _, err := (*launcherAssetService)(nil).RegisterStableBytes("ns", "x", []byte("a"), 1); err == nil {
		t.Error("nil service must error")
	}
}

func TestServeAssetRequest(t *testing.T) {
	s := testAssetService()
	id32 := "0123456789abcdef0123456789abcdef"
	if _, err := s.RegisterBytes("ns", id32, []byte("serve-body"), 1); err != nil {
		t.Fatal(err)
	}

	// 命中 → 200 + body。
	req := httptest.NewRequest("GET", "http://x/__usbeam_asset__/ns/"+id32+"?v=1", nil)
	rec := httptest.NewRecorder()
	if !s.ServeAssetRequest(rec, req) {
		t.Fatal("expected handled (true)")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("code=%d want 200", rec.Code)
	}
	if rec.Body.String() != "serve-body" {
		t.Errorf("body=%q want serve-body", rec.Body.String())
	}

	// 非资产路径 → 已处理（HTTP 错误响应）。
	req2 := httptest.NewRequest("GET", "http://example.com/other", nil)
	if !s.ServeAssetRequest(httptest.NewRecorder(), req2) {
		t.Error("non-asset path must still return true (handler consumed)")
	}

	// 未注册资产 → 已处理（HTTP 错误响应）。
	req3 := httptest.NewRequest("GET", "http://example.com/__usbeam_asset__/ns/ffffffffffffffffffffffffffffffff", nil)
	if !s.ServeAssetRequest(httptest.NewRecorder(), req3) {
		t.Error("unknown asset must still return true (handler consumed)")
	}
}

func TestStableLauncherAssetID(t *testing.T) {
	// Ghidra 确证 3 参签名 (a,b string, c []byte)：写序 Write(a),0x00,Write(b),0x00,Write(c)，
	// SHA-256 前 16 字节 hex。验证：stableLauncherAssetID("ns","id",nil)
	// = sha256HexPrefix("ns\x00\x00")（c 空，第三个分隔符之后无内容，实际两分隔符）。
	a := stableLauncherAssetID("ns", "id", nil)
	if len(a) != 32 || !isValidLauncherAssetIDHex([]byte(a)) {
		t.Errorf("stable id %q not 32-hex shape", a)
	}
	// 与独立 SHA-256 引擎对照（写序 a,0x00,b,0x00,c）：
	exp := sha256HexPrefix(append(append(append([]byte("ns"), 0), []byte("id")...), 0))
	if a != exp {
		t.Errorf("stable id %q != engine %q (write order a,0,b,0,c)", a, exp)
	}
	// 确定性：相同输入 → 相同 id。
	if b := stableLauncherAssetID("ns", "id", nil); b != a {
		t.Errorf("stable not deterministic: %q vs %q", a, b)
	}
	// 输入变化 → id 变化（b 段不同）。
	if stableLauncherAssetID("ns", "other", nil) == a {
		t.Error("different segments must yield different stable id")
	}
	// []byte 第三参参与哈希：非空 c 改变指纹。
	if stableLauncherAssetID("ns", "id", []byte("blob")) == a {
		t.Error("c []byte segment must alter stable id")
	}
}

func TestRegisterStableBody(t *testing.T) {
	s := testAssetService()
	id32 := "0123456789abcdef0123456789abcdef"

	// 非 32-hex 稳定 id → error。
	if _, err := s.registerStable("ns", "short", 1); err == nil {
		t.Error("non-32-hex id must error")
	}

	// 新 id → 建立 entry + ref。
	ref, err := s.registerStable("ns", id32, 5)
	if err != nil {
		t.Fatalf("new stable err=%v", err)
	}
	if ref.ID != id32 || ref.Version != 5 {
		t.Errorf("ref=%+v want id %s v5", ref, id32)
	}
	e, ok := s.entries[id32]
	if !ok || e.namespace != "ns" || e.version != 5 {
		t.Errorf("entry=%+v ok=%v", e, ok)
	}
	if e.expiresAt.IsZero() {
		t.Error("expiresAt must be set")
	}

	// 重复同 id → 复用（next 不变、仍成功）。
	next0 := s.next
	ref2, err := s.registerStable("ns", id32, 5)
	if err != nil || ref2.ID != id32 {
		t.Errorf("re-register err=%v", err)
	}
	if s.next != next0 {
		t.Errorf("next changed across stable re-register: %d → %d", next0, s.next)
	}

	// nil receiver → error。
	if _, err := (*launcherAssetService)(nil).registerStable("ns", id32, 1); err == nil {
		t.Error("nil service must error")
	}
}

func TestLookupExpanded(t *testing.T) {
	// 过期→条目被删除并回退计数（Ghidra C 语义）。
	s := testAssetService()
	s.now = func() time.Time { return time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC) }
	s.totalBytes = 100
	s.namespaceBytes["ns"] = 50
	s.entries["exp"] = launcherAssetEntry{id: "exp", namespace: "ns", size: 40}
	s.entries["exp"] = launcherAssetEntry{id: "exp", namespace: "ns", size: 40, expiresAt: s.now().Add(-48 * time.Hour)}
	if s.lookup("ns", "exp") {
		t.Error("expired must not be found")
	}
	if _, ok := s.entries["exp"]; ok {
		t.Error("expired entry must be removed")
	}
	if s.totalBytes != 60 {
		t.Errorf("totalBytes=%d want 60", s.totalBytes)
	}
	if s.namespaceBytes["ns"] != 10 {
		t.Errorf("namespaceBytes[ns]=%d want 10", s.namespaceBytes["ns"])
	}

	// namespace 不匹配 → 未命中且不删除。
	s2 := testAssetService()
	s2.now = func() time.Time { return time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC) }
	s2.entries["x"] = launcherAssetEntry{id: "x", namespace: "ns", accessedAt: time.Time{}}
	if s2.lookup("other", "x") {
		t.Error("ns mismatch must not be found")
	}
	if _, ok := s2.entries["x"]; !ok {
		t.Error("ns mismatch must not delete entry")
	}
}

func TestRegisterBody(t *testing.T) {
	// register 内部壳（Ghidra 反编译 0x14086f680）：id 空自动生成查重 / next 递增 /
	// entry 装配（data/size/TTL 600s）/ ensureCapacity 出错分支 / buildURL=next。
	s := testAssetService()

	// 显式 id + 数据：entry 落表、data/size/namespace、next 递增、URL ?v=next。
	ref, err := s.register("ns", "reg1", []byte("payload"), 3)
	if err != nil {
		t.Fatalf("register err=%v", err)
	}
	if ref.ID != "reg1" || ref.Version != 3 {
		t.Errorf("ref=%+v want id reg1 v3", ref)
	}
	if want := "/__usbeam_asset__/ns/reg1?v=1"; ref.URL != want {
		t.Errorf("ref.URL=%q want %q", ref.URL, want)
	}
	e, ok := s.entries["reg1"]
	if !ok || e.namespace != "ns" || e.size != 7 || string(e.data) != "payload" || e.version != 3 {
		t.Errorf("entry=%+v ok=%v", e, ok)
	}
	if s.next != 1 {
		t.Errorf("next=%d want 1", s.next)
	}

	// 空 id → 自动生成合法 32-hex（查重唯一）。
	id0 := ref.ID
	ref2, err := s.register("ns", "", []byte("x"), 1)
	if err != nil {
		t.Fatalf("empty-id register err=%v", err)
	}
	if len(ref2.ID) != 32 || !isValidLauncherAssetIDHex([]byte(ref2.ID)) || ref2.ID == id0 {
		t.Errorf("auto id=%q invalid or collides %q", ref2.ID, id0)
	}

	// TTL：expiresAt ≈ now + 600s（微秒级容忍，验证非零且未来）。
	if d := time.Until(s.entries["reg1"].expiresAt); d < 500*time.Second || d > 700*time.Second {
		t.Errorf("expiresAt=%v (%v from now), want ~600s TTL", s.entries["reg1"].expiresAt, d)
	}

	// ensureCapacity 越限 → error（据 limits.maxTotalBytes 触发驱逐耗尽）。
}

func TestRegisterBodyCapacityError(t *testing.T) {
	// register 内 ensureCapacityLocked 失败路径：itemSize>maxTotalBytes → 直接拒绝（驱逐也无效）。
	s := testAssetService()
	s.limits.maxTotalBytes = 5
	if _, err := s.register("ns", "", make([]byte, 10), 1); err == nil {
		t.Error("capacity exceeded must error")
	}

	// nil receiver。
	if _, err := (*launcherAssetService)(nil).register("ns", "x", []byte("a"), 1); err == nil {
		t.Error("nil service must error")
	}
}
