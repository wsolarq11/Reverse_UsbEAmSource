// AUTO-RECONSTRUCTED — DOMAIN: launcher config icon store (存储核心方法)
// 研究用途
//
// 本文件为 launcherconfigiconstore.go 的延续，按函数拓扑拆分以避免单文件过长。
// 函数排列遵守 grep ^func 即执行叙事（入口 → 攒 → 拆 → 平台层）。
package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// ---- 存储核心 ----

// Put 将图标数据 URL 解码、校验、写入图标库，返回其规范化引用。
// [S-sig asm 0x1408928e0] 控制流逐条确证：
//
//	nil receiver → 错误
//	→ decodeLauncherConfigIconDataURL(path, data, &stackBudget) → 失败即透传
//	→ newLauncherConfigIconRef(decoded) → 得 ref
//	→ mu.Lock() + defer Unlock
//	→ loadUnlocked() → 失败即透传
//	→ cached.Icons[ref] 已存在 → 直接返回 (ref, nil)
//	→ add(path, data) 预算校验 → 失败即 error
//	→ mapassign_faststr 写入 cached.Icons[ref] = {ContentType, Data}
//	→ writeUnlocked(cached) → 失败即 error
//	→ 返回 (ref, nil)
//
// 返回 (string, error)。成功返回新写入 ref，重复写入返回已存在的 ref（幂等，不重复写文件）。
func (s *launcherConfigIconStore) Put(path, data string) (string, error) {
	if s == nil {
		return "", errors.New("launcherConfigIconStore: nil receiver")
	}

	// 解码并预算校验（使用临时 budget 作为探测）
	var probeBudget launcherConfigIconBudget
	contentType, body, decoded, err := decodeLauncherConfigIconDataURL(path, data, &probeBudget)
	if err != nil {
		return "", err
	}
	_ = body

	ref := newLauncherConfigIconRef(decoded)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, _, _, err := s.loadUnlocked(); err != nil {
		return "", err
	}

	// 幂等：已存在则直接返回
	if _, ok := s.cached.Icons[ref]; ok {
		return ref, nil
	}

	// 预算校验
	if err := s.cachedBudget.add(path, data); err != nil {
		return "", err
	}

	// 写入缓存
	s.cached.Icons[ref] = launcherConfigIconRecord{
		ContentType: contentType,
		Data:        body,
	}

	// 持久化
	if err := s.writeUnlocked(s.cached); err != nil {
		return "", err
	}

	return ref, nil
}

// ExtractLauncherConfigIcons 提取配置中的内联图标数据到图标库（force=false 模式）。
// [S 汇编实证 0x140893700]：xor ecx,ecx; xor edi,edi; call externalizeLauncherConfigIcons
// → 透传 self, cfg, refSet, force=false（64B 薄透传）。
func (s *launcherConfigIconStore) ExtractLauncherConfigIcons(cfg LauncherConfig, refSet map[string]struct{}) (launcherConfigIconExtractionResult, error) {
	return s.externalizeLauncherConfigIcons(cfg, refSet, false)
}

// ExternalizeLauncherConfigIcons 外部化图标数据（force=true 模式）。
// [S 汇编实证 0x140893740]：mov edi,1; call externalizeLauncherConfigIcons
// → 透传 self, cfg, refSet, force=true（64B 薄透传）。
func (s *launcherConfigIconStore) ExternalizeLauncherConfigIcons(cfg LauncherConfig, refSet map[string]struct{}) (launcherConfigIconExtractionResult, error) {
	return s.externalizeLauncherConfigIcons(cfg, refSet, true)
}

// externalizeLauncherConfigIcons 外部化配置中的内嵌图标数据。
// [S-sig asm 0x1408937a0, 0x14f8 字节帧]
//
// 入参：(s *launcherConfigIconStore, cfg LauncherConfig, refSet map[string]struct{}, force bool)
// 流程：validateLauncherConfigIconData → visitLauncherConfigIconSlots 逐槽处理 →
//
//	出槽后 normalizeLauncherConfigIconRefSet(refSet) → Prune 掉不在此集合的 Icons
//	→ mu.Lock → loadUnlocked → 已存在图标的复用计数
//	→ 新图标 decodeLauncherConfigIconDataURL + budget.add + 写入 Icons map
//	→ writeUnlocked → 返回统计
func (s *launcherConfigIconStore) externalizeLauncherConfigIcons(cfg LauncherConfig, refSet map[string]struct{}, force bool) (launcherConfigIconExtractionResult, error) {
	if s == nil {
		return launcherConfigIconExtractionResult{}, errors.New("launcherConfigIconStore: nil receiver")
	}

	// [P] launcherConfigIconStorePath 的 4 个 .rdata 常量尚未解码，
	// 路径为空时退化为无操作（兼容测试环境与未初始化存储）。
	if strings.TrimSpace(s.path) == "" {
		return launcherConfigIconExtractionResult{}, nil
	}

	// 校验全部图标预算
	validateLauncherConfigIconData(&cfg)

	// 收集 ref 集合与提取统计
	accumulatedRefs := make(map[string]struct{})
	var added, reused int
	var accum []launcherConfigIconSlot // 存有内联数据的槽

	visitLauncherConfigIconSlots(&cfg, func(slot launcherConfigIconSlot) bool {
		// 处理外部引用（无内联数据但有 Ref）
		if slot.Ref != nil && *slot.Ref != "" {
			accumulatedRefs[*slot.Ref] = struct{}{}
		}

		// 处理内联数据
		if slot.Data == nil || *slot.Data == "" {
			return false
		}
		trimmed := strings.TrimSpace(*slot.Data)
		if trimmed == "" {
			return false
		}

		// 解码并加入累加列表
		accum = append(accum, slot)
		return false
	})

	// 规范化外部引用集（合并 refSet 参数 + accumulatedRefs）
	normalizedRefs := make(map[string]struct{})
	for ref := range refSet {
		if ref != "" {
			normalizedRefs[ref] = struct{}{}
		}
	}
	for ref := range accumulatedRefs {
		normalizedRefs[ref] = struct{}{}
	}

	// 如果无内联数据且 force=false → 只做 prune 不做新写入
	if len(accum) == 0 && !force {
		return launcherConfigIconExtractionResult{}, nil
	}

	// 加锁并加载
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, _, _, err := s.loadUnlocked(); err != nil {
		return launcherConfigIconExtractionResult{}, err
	}

	// Prune 阶段：删除 Icons 中不在 normalizedRefs 中的项
	var removed int
	for ref := range s.cached.Icons {
		if _, ok := normalizedRefs[ref]; !ok {
			delete(s.cached.Icons, ref)
			removed++
		}
	}

	// 提取阶段：逐个处理内联数据
	for _, slot := range accum {
		if slot.Data == nil || *slot.Data == "" {
			continue
		}
		trimmed := strings.TrimSpace(*slot.Data)
		if trimmed == "" {
			continue
		}

		// 解码（用栈 budget 做逐条校验，与 Put 一致）
		var probeBudget launcherConfigIconBudget
		contentType, body, decoded, err := decodeLauncherConfigIconDataURL(slot.Path, trimmed, &probeBudget)
		if err != nil {
			continue
		}

		ref := newLauncherConfigIconRef(decoded)

		// 检查是否已存在（幂等）
		if _, exists := s.cached.Icons[ref]; exists {
			reused++
			continue
		}

		// 存储预算校验
		if err := s.cachedBudget.add(slot.Path, trimmed); err != nil {
			break
		}

		// 写入缓存
		s.cached.Icons[ref] = launcherConfigIconRecord{
			ContentType: contentType,
			Data:        body,
		}
		added++
	}

	// 持久化
	if added > 0 || removed > 0 {
		if err := s.writeUnlocked(s.cached); err != nil {
			return launcherConfigIconExtractionResult{}, err
		}
	}

	return launcherConfigIconExtractionResult{
		Extracted: added,
		Added:     added,
		Reused:    reused,
		Removed:   removed,
	}, nil
}

// loadUnlocked 在持锁前提下加载并快照当前图标库数据。
// [S 汇编实证 0x140894fa0, 108L] 完整控制流：
//
//	makemap_small → 空 map（错误路径返回值，rbx）
//	→ ensureLoadedUnlocked() → error 非 nil → 返回 (1, 空map, nil, err)
//	→ newobject(launcherConfigIconStoreBudget) 快照 cachedBudget 三字段（0x48/0x50/0x58）
//	→ makemap(hint=len(cached.Icons)) + mapiter 逐项拷贝 Icons（mapassign_faststr + record 拷贝）
//	→ 返回 (cached.Version, Icons 副本, budget 快照, nil)
//
// 返回 4 值（Version, Icons 副本, budget 快照, error）以便释放锁后安全使用快照。
func (s *launcherConfigIconStore) loadUnlocked() (int, map[string]launcherConfigIconRecord, *launcherConfigIconStoreBudget, error) {
	empty := make(map[string]launcherConfigIconRecord)
	if err := s.ensureLoadedUnlocked(); err != nil {
		return 1, empty, nil, err
	}

	budget := &launcherConfigIconStoreBudget{
		count:        s.cachedBudget.count,
		decodedBytes: s.cachedBudget.decodedBytes,
		pixels:       s.cachedBudget.pixels,
	}

	icons := make(map[string]launcherConfigIconRecord, len(s.cached.Icons))
	for ref, rec := range s.cached.Icons {
		icons[ref] = rec
	}

	return s.cached.Version, icons, budget, nil
}

// writeUnlocked 在持锁前提下持久化当前图标库。
// [S 汇编实证 0x140895a80, 47L] 完整控制流：
//
//	nil receiver 或 TrimSpace(path) 空 → 错误 "图标库路径不能为空"(27B)
//	→ validateLauncherConfigIconLibrary(lib) → 返回 (budget 快照, err)
//	→ err 非 nil → 透传；budget 非 nil → 写回 s.cachedBudget
//	→ writeLibrary 为 nil 时回退 writeLauncherConfigIconLibraryFile
//	→ writeLibrary(path, lib) → 失败即透传
//	→ s.cached = lib; s.loaded = true; s.cachedLoadErr = nil
//
// 签名实证：morestack 存 3 寄存器（s, lib.Version, lib.Icons），非无参方法。
func (s *launcherConfigIconStore) writeUnlocked(lib launcherConfigIconLibrary) error {
	if s == nil || strings.TrimSpace(s.path) == "" {
		return errors.New("图标库路径不能为空")
	}

	budget, err := validateLauncherConfigIconLibrary(lib)
	if err != nil {
		return err
	}

	writeFn := s.writeLibrary
	if writeFn == nil {
		writeFn = writeLauncherConfigIconLibraryFile
	}

	if err := writeFn(s.path, lib); err != nil {
		return err
	}

	s.cached = lib
	if budget != nil {
		s.cachedBudget = *budget
	}
	s.loaded = true
	s.cachedLoadErr = nil
	return nil
}

// Prune 从图标库中移除集合中不存在的引用。
// [S-sig asm 0x1408948e0] 实证控制流：
//
//	nil receiver → error（len=0x15 常量）
//	→ normalizeLauncherConfigIconRefSet(refs) → 失败即返回该 error
//	→ mu.Lock() + defer Unlock
//	→ loadUnlocked() → 失败即返回
//	→ 遍历 s.cached.Icons，对每项用 mapaccess2_faststr 检查是否在 normalized 集合中
//	→ 不在集合中 → mapdelete_faststr + counter++
//	→ counter > 0 → writeUnlocked()
//	→ 返回 (counter, nil)
func (s *launcherConfigIconStore) Prune(refs map[string]struct{}) (int, error) {
	if s == nil {
		return 0, errors.New("launcherConfigIconStore: nil receiver")
	}

	normalized, err := normalizeLauncherConfigIconRefSet(refs)
	if err != nil {
		return 0, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, _, _, err := s.loadUnlocked(); err != nil {
		return 0, err
	}

	var removed int
	for ref := range s.cached.Icons {
		if _, ok := normalized[ref]; !ok {
			delete(s.cached.Icons, ref)
			removed++
		}
	}

	if removed > 0 {
		if err := s.writeUnlocked(s.cached); err != nil {
			return removed, err
		}
	}

	return removed, nil
}

// ---- 文件层 ----

// readLauncherConfigIconStoreBytes 读取图标库文件全部字节。
// [S-sig asm 0x1408968a0] 实证控制流：
//
//	TrimSpace(path) → 空 → ""（非错误）
//	→ os.OpenFile(path, O_RDONLY, 0) → 错误即返回
//	→ os.File.Stat → FileInfo
//	→ Mode().IsDir() → 目录错误
//	→ Size() > 128 MiB (0x8000000) → 体积超上限
//	→ io.ReadAll(file) → 返回 data
func readLauncherConfigIconStoreBytes(path string) ([]byte, error) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return nil, nil
	}

	f, err := os.OpenFile(trimmed, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, fmt.Errorf("读取图标库: %s 是目录", trimmed)
	}

	if info.Size() > launcherConfigIconStoreBytesSizeCeiling {
		return nil, fmt.Errorf("图标库文件体积超出上限: %d > %d", info.Size(), launcherConfigIconStoreBytesSizeCeiling)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// writeLauncherConfigIconLibraryFile 将图标库写入 JSON 文件。
// [S-sig asm 0x140896de0] 实证控制流：
//
//	TrimSpace(path) → 空 → error
//	→ json.Marshal(lib)
//	→ os.WriteFile(path, data, 0644)
func writeLauncherConfigIconLibraryFile(path string, lib launcherConfigIconLibrary) error {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return errors.New("launcherConfigIconStore: 写入路径为空")
	}

	data, err := json.Marshal(lib)
	if err != nil {
		return err
	}

	return os.WriteFile(trimmed, data, 0644)
}

// validateLauncherConfigIconLibrary 校验图标库结构完整性并累加预算。
// [S 汇编实证 0x140895e80, 39L] 完整控制流：
//
//	lib.Version != 1 → 版本不兼容错误（fmt.Sprintf len=0x15）
//	→ len(lib.Icons) > 8192 (0x2000) → 条目数超限错误（fmt.Sprintf len=0x21）
//	→ 收集 refs 到 slice（count<=2 用栈缓冲，否则 makeslice）+ mapIterStart
//	→ sort.Strings(refs) → newobject(budget) → 遍历排序后 refs：
//	  ① normalizeLauncherConfigIconRef(ref) 失败或 != ref → 引用未规范化错误
//	  ② TrimSpace(ContentType) 空或 ToLower 不一致 → 内容类型错误（len=0x26）
//	  ③ TrimSpace(Data) 空 → 数据为空错误（len=0x2a）
//	  ④ budget.add("icons[ref(", "data:contentType;base64,data") 累加，失败即透传
//	  ⑤ newLauncherConfigIconRef(base64解码(Data)) != ref → 引用不匹配错误（len=0x24）
//	→ 返回 (budget 快照, nil)
//
// 签名实证：writeUnlocked 以 3 寄存器调本函数（Version, Icons），
// 返回 (rax=budget 指针, rbx=error.itab, rcx=error.data)，非单 error。
func validateLauncherConfigIconLibrary(lib launcherConfigIconLibrary) (*launcherConfigIconStoreBudget, error) {
	if lib.Version != 1 {
		return nil, launcherConfigIconStoreError("", fmt.Sprintf("版本不兼容: %d", lib.Version))
	}
	if len(lib.Icons) > launcherConfigIconStoreCountLimit {
		return nil, launcherConfigIconStoreError("", fmt.Sprintf("图标条目数 %d 超过上限 %d", len(lib.Icons), launcherConfigIconStoreCountLimit))
	}

	refs := make([]string, 0, len(lib.Icons))
	for ref := range lib.Icons {
		refs = append(refs, ref)
	}
	sort.Strings(refs)

	budget := &launcherConfigIconStoreBudget{}
	for _, ref := range refs {
		rec := lib.Icons[ref]

		normalized, err := normalizeLauncherConfigIconRef(ref)
		if err != nil {
			return nil, err
		}
		if normalized != ref {
			return nil, launcherConfigIconStoreError(ref, "引用未规范化")
		}

		contentType := strings.TrimSpace(rec.ContentType)
		if contentType == "" || strings.ToLower(contentType) != contentType {
			return nil, launcherConfigIconStoreError(ref, "内容类型无效")
		}

		data := strings.TrimSpace(rec.Data)
		if data == "" {
			return nil, launcherConfigIconStoreError(ref, "数据为空")
		}

		path := "icons[" + ref + "("
		dataURL := "data:" + contentType + ";base64," + rec.Data
		if err := budget.add(path, dataURL); err != nil {
			return nil, err
		}

		decoded, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			return nil, launcherConfigIconStoreError(ref, "Base64 解码失败")
		}
		if newLauncherConfigIconRef(decoded) != ref {
			return nil, launcherConfigIconStoreError(ref, "引用与数据不匹配")
		}
	}
	return budget, nil
}
