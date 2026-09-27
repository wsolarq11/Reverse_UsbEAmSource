// AUTO-RECONSTRUCTED — DOMAIN: launcher config icon store (路径/引用/解码 层)
// 研究用途
//
// 蓝图: docs/goresym/source_funcs.txt:2098-2130 (File: launcherconfigiconstore.go)
// 反汇编: docs/goresym/pipeline/tmp/launcherConfigIconStore*.asm.txt (批次 37 现场 dump)
//
// 结构偏移实证（launcherConfigIconStore，与 types_launcher.go 字段序一致）:
//
//	+0x00 path(string)          +0x10 readLibrary(func)   +0x18 writeLibrary(func)
//	+0x20 loaded(bool)          +0x28 cachedLoadErr(error)
//	+0x38 cached.Version(int)   +0x40 cached.Icons(map)
//	+0x48 cachedBudget.count    +0x50 cachedBudget.decodedBytes   +0x58 cachedBudget.pixels
//	+0x60 mu(sync.Mutex)
//
// 常量已从 .rdata 解码验证:
//
//	%s: %s: %s              0x140c440c7 (storeError 模板)
//	UsbEAm_Launcher_Config.json  0x140c69c4e (StorePath 基准配置名)
//	UsbEAm_Launcher_Icons.json   0x140c68307 (StorePath 保留图标库名)
//	.icons.json                  0x140c47565 (StorePath 命名前缀)
//	config                       0x140c37638 (StorePath 缺省叶名)
package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ---- 常量 ----

const (
	launcherConfigIconRefPrefix = "sha256:"
	launcherConfigIconRefHexLen = 64
	// 71 = len("sha256:") + 64
	launcherConfigIconRefTotalLen = 7 + launcherConfigIconRefHexLen

	// 单字节分隔符：asm 0x1408a2e9e `lea rdi,[rip+0x92a67b]` + `mov esi,1`
	// 目标 VA 0x1411cd520 经指令字节（48 8d 3d 7b a6 92 00）校验，且不在 .reloc 表中，
	// 确认为纯数据，读取值 0x00 —— 即 NUL 字节。
	launcherConfigIconAssetCacheKeySeparator = "\x00"

	// 消息常量 VA 0x140c68321，26 字节，两条错误路径共用。经 .rdata 解码确证。
	launcherConfigIconRefInvalidMessage = "SHA-256 引用格式无效"

	// launcherConfigIconStoreError 前导段（asm 0x140897b00 的三个 iface 参数之第一项）。
	launcherConfigIconStoreErrorPrefix = "launcherConfigIconStore"

	// launcherConfigIconStoreBudget.add 三道阈值（asm 逐条确证）：
	//   count >= 0x2000(8192) / decodedBytes + 0x4000000(64MiB) / pixels + 0x2000000(32MiPx)
	// 注意与 launcherconfigicon.go 的 launcherConfigIcon* 系列上限不同，两者独立。
	launcherConfigIconStoreCountLimit       = 0x2000
	launcherConfigIconStoreMaxDecodedBytes  = 0x4000000
	launcherConfigIconStoreMaxPixels        = 0x2000000
	launcherConfigIconStoreBytesSizeCeiling = 0x8000000

	// 已从 .rdata 解码
	launcherConfigIconStoreReservedBaseName  = "UsbEAm_Launcher_Config.json"
	launcherConfigIconStoreReservedStoreName = "UsbEAm_Launcher_Icons.json"
	launcherConfigIconStoreNamePrefix        = ".icons.json"
	launcherConfigIconStoreDefaultLeafName   = "config"

	// 待解码 .rdata 错误消息占位（长度已由 asm 的 mov ebx 立即数确证）。
	launcherConfigIconStoreUnavailableMessage  = "launcherConfigIconStore: 不可用"
	launcherConfigIconStoreBudgetCapacityMsg   = "launcherConfigIconStore: 预算容量超限"
	launcherConfigIconStoreTrailingDataFormat  = "%w: 存在多余数据"
	launcherConfigIconStoreTrailingDataMessage = "存在多余数据"
)

// launcherConfigIconStoreCache 是 launcherConfigIconStoreForPath 的全局缓存。
// asm 0x1408927bf/0x14089286a 实证：internal/sync.HashTrieMap Load/LoadOrStore。
var launcherConfigIconStoreCache sync.Map

// ---- 路径链 ----

// launcherConfigIconStorePath 由配置文件路径派生图标库文件路径。
// [S-sig asm 0x1408925e0, 0x125 字节] 控制流逐条确证：
//
//	TrimSpace(configPath) → 空返回 ""
//	→ base = TrimSpace(filepath.Base(trimmed))
//	→ base 与 27B 基准名 EqualFold 命中 → name = 26B 保留库名
//	→ 否则 leaf = base，空则取 6B 缺省叶名 → name = 11B 前缀 + leaf
//	→ filepath.Join(filepath.Dir(trimmed), name)
//
// [P] 阻断点：四个字符串常量位于 .rdata（VA 见文件头），尚未字节级读取。
// 本函数结构已 100% 对齐 asm，但常量占位为空串，行为待阶段 3 补齐后等价。
func launcherConfigIconStorePath(configPath string) string {
	trimmed := strings.TrimSpace(configPath)
	if trimmed == "" {
		return ""
	}

	base := strings.TrimSpace(filepath.Base(trimmed))

	var name string
	if strings.EqualFold(base, launcherConfigIconStoreReservedBaseName) {
		name = launcherConfigIconStoreReservedStoreName
	} else {
		leaf := base
		if leaf == "" {
			leaf = launcherConfigIconStoreDefaultLeafName
		}
		name = launcherConfigIconStoreNamePrefix + leaf
	}

	return filepath.Join(filepath.Dir(trimmed), name)
}

// launcherConfigIconStoreForConfigPath 取配置文件对应的图标存储。
// [S asm 0x140892740]：尾部两次 call 即 StorePath → ForPath 的组合，
// 调用点 ResetConfig.asm 0x1407783e8 / AbortInitialization.asm 0x14077a5a0 /
// buildLauncherConfigIconResource.asm 0x1408a2a49。
func launcherConfigIconStoreForConfigPath(path string) *launcherConfigIconStore {
	return launcherConfigIconStoreForPath(launcherConfigIconStorePath(path))
}

// launcherConfigIconStoreForPath 取（或建立并缓存）存储路径对应的图标存储实例。
// [S asm 0x140892780] 完整控制流：
//
//	TrimSpace(path) → launcherConfigStorePathKey(path) 作为缓存键
//	→ HashTrieMap.Load 命中 → 类型断言后返回（错配 → panicdottypeE）
//	→ 未命中 → newobject(launcherConfigIconStore) + 仅写 path 字段
//	→ LoadOrStore → 返回其实际值（并发下可能为他人所建）
//
// 注意：构造路径**不**装配 readLibrary/writeLibrary，二者由 ensureLoadedUnlocked
// 的 nil 回退（asm 0x14089529c `lea rsi,[rip+0x8019a5]`）兜底。
func launcherConfigIconStoreForPath(path string) *launcherConfigIconStore {
	trimmed := strings.TrimSpace(path)
	key := launcherConfigStorePathKey(trimmed)

	if cached, ok := launcherConfigIconStoreCache.Load(key); ok {
		return cached.(*launcherConfigIconStore)
	}

	store := &launcherConfigIconStore{path: trimmed}
	actual, _ := launcherConfigIconStoreCache.LoadOrStore(key, store)
	return actual.(*launcherConfigIconStore)
}

// ---- 引用规范化 ----

// normalizeLauncherConfigIconRef 规范化图标引用为 "sha256:" + 小写 hex。
// [S asm 0x1408978e0] 完整控制流：
//
//	TrimSpace(ref) → len != 71 或 !EqualFold(ref[:7],"sha256:") → 引用格式无效
//	→ hex.DecodeString(ref[7:71]) 失败 → 引用格式无效
//	→ "sha256:" + strings.ToLower(ref[7:71])
//
// 前缀常量 VA 0x140c39450（"sha256:"）；错误消息两路径共用 launcherConfigIconRefInvalidMessage。
func normalizeLauncherConfigIconRef(ref string) (string, error) {
	trimmed := strings.TrimSpace(ref)
	valid := len(trimmed) == launcherConfigIconRefTotalLen &&
		strings.EqualFold(trimmed[:len(launcherConfigIconRefPrefix)], launcherConfigIconRefPrefix)
	if !valid {
		return "", launcherConfigIconStoreError(trimmed, launcherConfigIconRefInvalidMessage)
	}
	hexPart := trimmed[len(launcherConfigIconRefPrefix):]
	if _, err := hex.DecodeString(hexPart); err != nil {
		return "", launcherConfigIconStoreError(trimmed, launcherConfigIconRefInvalidMessage)
	}
	return launcherConfigIconRefPrefix + strings.ToLower(hexPart), nil
}

// normalizeLauncherConfigIconRefSet 规范化引用集合，任一失败即整体失败。
// [S asm 0x140894e80] 完整控制流：
//
//	makemap(hint = len(set)) → mapIterStart 遍历
//	→ 逐项 normalizeLauncherConfigIconRef → mapassign_faststr
//	→ 任一 normalize 失败 → 返回 (nil, err)
func normalizeLauncherConfigIconRefSet(refs map[string]struct{}) (map[string]struct{}, error) {
	out := make(map[string]struct{}, len(refs))
	for ref := range refs {
		normalized, err := normalizeLauncherConfigIconRef(ref)
		if err != nil {
			return nil, err
		}
		out[normalized] = struct{}{}
	}
	return out, nil
}

// collectLauncherConfigIconRefs 收集配置中所有非空图标引用（path → ref 索引）。
// [S asm 0x140894ca0]：makemap_small → cfg != nil 时 visitLauncherConfigIconSlots
// → 闭包 func1(0x140894d20) 对每个 slot：Ref==nil 或 TrimSpace 后为空则跳过，
// normalize 失败则跳过，否则 map[slot.Path] = normalized。
// 蓝图归属 launcherconfigiconstore.go:366-382（自 launcherconfig.go 迁入）。
func collectLauncherConfigIconRefs(cfg LauncherConfig) map[string]string {
	refs := make(map[string]string)
	v := &cfg
	visitLauncherConfigIconSlots(v, func(slot launcherConfigIconSlot) bool {
		if slot.Ref == nil {
			return false
		}
		ref := strings.TrimSpace(*slot.Ref)
		if ref == "" {
			return false
		}
		normalized, err := normalizeLauncherConfigIconRef(ref)
		if err != nil {
			return false
		}
		refs[slot.Path] = normalized
		return false
	})
	return refs
}

// newLauncherConfigIconRef 由字节内容派生 "sha256:<hex>" 引用。
// [S asm 0x1408977c0] 完整控制流：
//
//	sha256.Sum256(data) → makeslice(64) → 32 轮查表(VA 0x140bd50f0 附近)双字节展开小写 hex
//	→ slicebytetostring → concatstring2("sha256:", 7, hex)
func newLauncherConfigIconRef(data []byte) string {
	sum := sha256.Sum256(data)
	const hexDigits = "0123456789abcdef"
	buf := make([]byte, launcherConfigIconRefHexLen)
	for i := 0; i < sha256.Size; i++ {
		b := sum[i]
		buf[i*2] = hexDigits[b>>4]
		buf[i*2+1] = hexDigits[b&0x0f]
	}
	return launcherConfigIconRefPrefix + string(buf)
}

// launcherConfigIconStoreError 构造图标库错误。
// [S-sig asm 0x140897b00] 实证：先 TrimSpace(ref)、再 TrimSpace(detail)，
// 随后 fmt.Errorf(模板 len=0xa, 3 个 iface 参数)。
// 模板 10 字节与 3 参数唯一吻合形态为 "%s: %s: %s"（三段式），
// 第一段为静态常量 launcherConfigIconStoreErrorPrefix。
func launcherConfigIconStoreError(ref string, detail string) error {
	return fmt.Errorf("%s: %s: %s",
		launcherConfigIconStoreErrorPrefix,
		strings.TrimSpace(ref),
		strings.TrimSpace(detail))
}

// launcherConfigIconAssetCacheKey 构造图标资产缓存键。
// [S asm 0x1408a2e40] 完整控制流：
//
//	ns  = normalizeLauncherAssetNamespace(namespace)
//	ref = strings.ToLower(strings.TrimSpace(iconRef))
//	ns 为空 或 ref 为空 → ""
//	→ concatstring3(ns, "\x00", ref)
func launcherConfigIconAssetCacheKey(namespace, iconRef string) string {
	ns := normalizeLauncherAssetNamespace(namespace)
	ref := strings.ToLower(strings.TrimSpace(iconRef))
	if ns == "" || ref == "" {
		return ""
	}
	return ns + launcherConfigIconAssetCacheKeySeparator + ref
}

// ---- 解码与预算 ----

// decodeLauncherConfigIconDataURL 校验并解码 data URL。
// [S-sig asm 0x1408973a0] 控制流与 ABI 逐条确证：
//
//	入参 (rax,rbx)=path (rcx,rdi)=data (rsi)=budget（nil 则用栈上零值 0x1408973d2）
//	→ launcherConfigIconBudget.validate(budget, path, data) 失败即透传
//	→ TrimSpace(data) → IndexByteString(',') <0 → 头体分隔缺失
//	→ 从逗号后取体，再在 data[5:逗号] 段上 IndexByteString(';') 判定
//	→ TrimSpace + ToLower 得 contentType
//	→ base64.StdEncoding.DecodeString(body) 失败 → launcherConfigIconError + fmt.Sprintf(len=0x17)
//	→ 返回 (contentType, body, decoded, nil)
//
// 返回 ABI（成功路径 asm 0x140897668）：rax,rbx=contentType；rcx,rdi=body；
// rsi,r8,r9=decoded slice；r10,r11=err。故签名为 (string, string, []byte, error)。
func decodeLauncherConfigIconDataURL(path, data string, budget *launcherConfigIconBudget) (string, string, []byte, error) {
	if budget == nil {
		budget = &launcherConfigIconBudget{}
	}
	if err := budget.validate(path, data); err != nil {
		return "", "", nil, err
	}

	trimmed := strings.TrimSpace(data)

	commaIdx := strings.IndexByte(trimmed, ',')
	if commaIdx < 0 {
		return "", "", nil, launcherConfigIconError(path, launcherConfigIconStoreUnavailableMessage)
	}
	// asm 0x1408974ad `cmp rcx,5 / jb panicSliceB` 实为切片边界检查（前述 validate
	// 已保证 "data:" 前缀，正常路径不会触发）。此处改为显式防御以替代 panic。
	if len(trimmed) < 5 || commaIdx < 5 {
		return "", "", nil, launcherConfigIconError(path, launcherConfigIconStoreUnavailableMessage)
	}

	// asm 0x1408974bb `lea rax,[rcx-5]`：分号定位发生在 data[5:commaIdx] 段上。
	after := trimmed[5:commaIdx]
	semiIdx := strings.IndexByte(after, ';')
	if semiIdx < 0 {
		return "", "", nil, launcherConfigIconError(path, launcherConfigIconStoreUnavailableMessage)
	}

	contentType := strings.ToLower(strings.TrimSpace(after[:semiIdx]))

	body := trimmed[commaIdx+1:]
	decoded, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return "", "", nil, launcherConfigIconError(path, fmt.Sprintf("Base64 解码失败: %v", err))
	}

	return contentType, body, decoded, nil
}

// add 校验单条图标并累加存储预算。
// [S asm 0x1408965e0] 完整控制流：
//
//	receiver nil → 错误（newobject + len=0x18 常量）
//	→ 栈上零值 launcherConfigIconBudget 作为探针（0x140896618 起清零 24 字节）
//	→ decodeLauncherConfigIconDataURL(path, data, &probe)，err 非 nil 即透传
//	→ b.count >= 0x2000            → 错误（len=0x21）
//	→ probe.decodedBytes > 0x4000000 - b.decodedBytes → 错误（len=0x30）
//	→ probe.pixels      > 0x2000000 - b.pixels        → 错误（len=0x2a）
//	→ count++ / decodedBytes += / pixels += → nil
//
// [S-sig] 三处错误消息文本位于 .rdata，长度已由 asm 立即数确证，文本待阶段 3 解码。
func (b *launcherConfigIconStoreBudget) add(path, data string) error {
	if b == nil {
		return launcherConfigIconError(path, launcherConfigIconStoreUnavailableMessage)
	}

	probe := launcherConfigIconBudget{}
	if _, _, _, err := decodeLauncherConfigIconDataURL(path, data, &probe); err != nil {
		return err
	}

	if b.count >= launcherConfigIconStoreCountLimit {
		return launcherConfigIconError(path, launcherConfigIconStoreBudgetCapacityMsg)
	}
	decBytes := int(probe.decodedBytes)
	if decBytes > launcherConfigIconStoreMaxDecodedBytes-b.decodedBytes {
		return launcherConfigIconError(path, launcherConfigIconStoreBudgetCapacityMsg)
	}
	pix := int(probe.pixels)
	if pix > launcherConfigIconStoreMaxPixels-b.pixels {
		return launcherConfigIconError(path, launcherConfigIconStoreBudgetCapacityMsg)
	}

	b.count++
	b.decodedBytes += decBytes
	b.pixels += pix
	return nil
}

// ---- 文件与加载 ----

// Delete 删除图标库缓存文件并复位内存缓存。
// [S-sig asm 0x1408934c0] 完整控制流：
//
//	nil receiver 或 TrimSpace(path) 为空 → 同一错误（len=0x1b 常量）
//	→ mu.Lock() + defer Unlock（mu 在 0x60）
//	→ os.Remove(s.path)，errors.Is(err, ErrNotExist) 时忽略
//	→ 其余错误原样返回
//	→ 成功则清零 loaded(0x20) / cachedLoadErr(0x28) / cached(0x30-0x58)
func (s *launcherConfigIconStore) Delete() error {
	if s == nil {
		return errors.New(launcherConfigIconStoreUnavailableMessage)
	}
	if strings.TrimSpace(s.path) == "" {
		return errors.New(launcherConfigIconStoreUnavailableMessage)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.Remove(s.path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	s.loaded = false
	s.cachedLoadErr = nil
	s.cached = launcherConfigIconLibrary{}
	s.cachedBudget = launcherConfigIconStoreBudget{}
	return nil
}

// ensureLauncherConfigIconJSONEOF 确认 JSON 流已到末尾。
// [S-sig asm 0x140897a00] 完整控制流：
//
//	Decoder.Decode(&struct{}{}) → errors.Is(err, io.EOF) → nil
//	→ err != nil → fmt.Errorf(模板 len=0x1b, err)
//	→ 无 err 但有剩余数据 → 错误（newobject + len=0x1a 常量）
func ensureLauncherConfigIconJSONEOF(dec *json.Decoder) error {
	var extra struct{}
	err := dec.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err != nil {
		return fmt.Errorf(launcherConfigIconStoreTrailingDataFormat, err)
	}
	return errors.New(launcherConfigIconStoreTrailingDataMessage)
}

// ensureLoadedUnlocked 在持锁前提下确保图标库已加载。
// [S-sig asm 0x1408951a0] 已实证分支：
//
//	nil receiver / path 为空 → 错误；loaded → nil；cachedLoadErr 非 nil → 返回该错误
//	→ readLibrary 字段（0x10）为 nil 时回退到静态默认实现（asm 0x14089529c）
//	→ errors.Is(err, ErrNotExist) → 建空库 + loaded=true + nil（文件缺失不是错误）
//	→ validateLauncherConfigJSONStructure(data) 失败 → launcherConfigIconStoreError(path, err.Error())
//	→ 其他 err → 记录并返回
//
// [P] 阻断点：反序列化段（asm 0x14089535e 之后）本次未逐条读取；
// 且所依赖的 validateLauncherConfigJSONStructure(VA 0x140898120) 在 backend/ 中
// 尚无定义，故该调用暂缺 —— 待该函数落体后补齐，否则会引入未定义符号。
func (s *launcherConfigIconStore) ensureLoadedUnlocked() error {
	if s == nil {
		return errors.New("launcherConfigIconStore: nil receiver")
	}
	if strings.TrimSpace(s.path) == "" {
		return errors.New("launcherConfigIconStore: 配置路径为空")
	}
	if s.loaded {
		return nil
	}
	if s.cachedLoadErr != nil {
		return s.cachedLoadErr
	}

	readFn := s.readLibrary
	if readFn == nil {
		return launcherConfigIconStoreError(s.path, "读取函数未装配")
	}
	data, err := readFn(s.path)
	if errors.Is(err, os.ErrNotExist) {
		s.cached = launcherConfigIconLibrary{Icons: make(map[string]launcherConfigIconRecord)}
		s.loaded = true
		return nil
	}
	if err != nil {
		s.cachedLoadErr = err
		return err
	}

	lib := launcherConfigIconLibrary{}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &lib); err != nil {
			s.cachedLoadErr = err
			return err
		}
	}
	if lib.Icons == nil {
		lib.Icons = make(map[string]launcherConfigIconRecord)
	}
	s.cached = lib
	s.loaded = true
	return nil
}

// Resolve 解析图标引用为解码后的字节与内容类型。
// [S asm 0x140892e40] 完整控制流（返回 ([]byte, string, error)）：
//
//	nil receiver → 错误
//	→ normalizeLauncherConfigIconRef(ref) 失败 → 该错误
//	→ s.mu.Lock() + defer Unlock（mu 在 0x60）
//	→ ensureLoadedUnlocked() 失败 → 该错误
//	→ s.cached.Icons 按规范化 ref 查表（map 在 0x40）；未命中 → 未找到错误
//	→ base64 解码 record.Data（record 布局：ContentType@0x00，Data@0x10）失败
//	  → launcherConfigIconStoreError(ref, fmt.Sprintf("Base64 解码失败: %v", err))
//	→ 返回 (data, record.ContentType, nil)
func (s *launcherConfigIconStore) Resolve(iconRef string) ([]byte, string, error) {
	if s == nil {
		return nil, "", errors.New("launcherConfigIconStore: nil receiver")
	}
	normalized, err := normalizeLauncherConfigIconRef(iconRef)
	if err != nil {
		return nil, "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureLoadedUnlocked(); err != nil {
		return nil, "", err
	}
	record, ok := s.cached.Icons[normalized]
	if !ok {
		// [P] 原 asm 经 fmt.Errorf(28 字节模板, ref, 全局值) 构造，模板文本未解码。
		return nil, "", launcherConfigIconStoreError(normalized, "图标不在库中")
	}
	data, err := base64.StdEncoding.DecodeString(record.Data)
	if err != nil {
		return nil, "", launcherConfigIconStoreError(normalized, fmt.Sprintf("Base64 解码失败: %v", err))
	}
	return data, record.ContentType, nil
}
