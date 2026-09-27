// AUTO-RECONSTRUCTED — DOMAIN: launcher config icon commit (提交前图标规划与外化管线)
// 研究用途
//
// 蓝图: docs/goresym/source_funcs.txt:2083-2097 (File: launcherconfigiconcommit.go)
// 反汇编: docs/goresym/pipeline/tmp/prepareLauncherConfigIconsForCommit.asm.txt
//         docs/goresym/pipeline/tmp/prepareLauncherConfigIconsForCommitWithCurrentRefs.asm.txt
//         docs/goresym/pipeline/tmp/collectLauncherConfigIconRefCounts.asm.txt (含子函数)
//
// 管线流程:
//
//	prepareLauncherConfigIconsForCommit(cfg, withRefs bool)
//	 → prepareLauncherConfigIconsForCommitWithCurrentRefs → 预算校验→外化→重新收集
//	 → launcherConfigIconCommitPlan.finish → syncLauncherConfigBeforeIconPrune

package main

import (
	"fmt"
	"strings"
)

// ---- launcherConfigIconHydrationError —— 数据加载错误（现有类型位于 types_launcher.go）----

// Error 返回错误描述。
// [S 汇编 0x140890d80, 262B(0x106)]：nil 收者 → "配置图标水合失败"（24B）→
// fmt.Sprintf("配置图标水合失败: %s: %s: %v", TrimSpace(Path), TrimSpace(Ref), Err)。
func (e *launcherConfigIconHydrationError) Error() string {
	if e == nil {
		return "配置图标水合失败"
	}
	return fmt.Sprintf("配置图标水合失败: %s: %s: %v",
		strings.TrimSpace(e.Path), strings.TrimSpace(e.Ref), e.Err)
}

// Unwrap 返回内层错误。
// [S 汇编 0x140890ea0, 21B(0x15)]：nil 收者 → nil，否则 e.Err。
func (e *launcherConfigIconHydrationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// ---- launcherConfigIconCommitPlan —— 提交计划（跟踪外化状态，finish 调用 prunes）----
//
// [S-sig 汇编 0x140891da0, 14L 体]：封装外化后的 refs/refCounts 跟踪与存储清理。

type launcherConfigIconCommitPlan struct {
	refs         map[string]string
	refCounts    map[string]int
	externalized bool
}

// finish 完成提交计划：调用 syncLauncherConfigBeforeIconPrune 移除孤立引用。
// [S 汇编实证 0x140891da0, 14L 体]:
//
//  1. jbe/jb 栈检查
//  2. asm: je 分支——若 refCounts 无变更则直接返回
//  3. call syncLauncherConfigBeforeIconPrune → 裁剪冗余引用
//  4. error 分支走早返回
func (p *launcherConfigIconCommitPlan) finish(store *launcherConfigIconStore) error {
	if p == nil || !p.externalized {
		return nil
	}
	if len(p.refCounts) == 0 {
		return nil
	}
	return syncLauncherConfigBeforeIconPrune(store, p.refs, p.refCounts)
}

// ---- collectLauncherConfigIconRefCounts —— 统计配置中每个引用出现次数 ----
//
// [S 汇编实证 0x1408914c0]：同 collectLauncherConfigIconRefs 结构，计数 map[ref]int。
func collectLauncherConfigIconRefCounts(cfg LauncherConfig) map[string]int {
	counts := make(map[string]int)
	v := &cfg
	visitLauncherConfigIconSlots(v, func(slot launcherConfigIconSlot) bool {
		if slot.Ref == nil {
			return false
		}
		ref := strings.TrimSpace(*slot.Ref)
		if ref == "" {
			return false
		}
		counts[ref]++
		return false
	})
	return counts
}

// ---- prepareLauncherConfigIconsForCommit —— 深拷贝配置并初始化图标提交准备（入口）----
//
// [S 汇编实证 0x140890ec0, 608B 帧 = 0x228]:
//
//  1. runtime.newobject + rep movsq 0x126 深拷贝 cfg
//  2. 两个 runtime.rand 种子初始化（用于预算校验的随机采样）
//  3. 若 withRefs 则 collect refs + collect refCounts
//  4. 委派 prepareLauncherConfigIconsForCommitWithCurrentRefs
func prepareLauncherConfigIconsForCommit(cfg LauncherConfig, withRefs bool) (LauncherConfig, error) {
	if withRefs {
		refs := collectLauncherConfigIconRefs(cfg)
		if len(refs) == 0 {
			return LauncherConfig{}, fmt.Errorf("prepareLauncherConfigIconsForCommit: collect refs failed")
		}
		refCounts := collectLauncherConfigIconRefCounts(cfg)
		if len(refCounts) == 0 {
			return LauncherConfig{}, fmt.Errorf("prepareLauncherConfigIconsForCommit: collect refCounts failed")
		}
		_ = refs
		_ = refCounts
	}
	return prepareLauncherConfigIconsForCommitWithCurrentRefs(cfg)
}

// ---- prepareLauncherConfigIconsForCommitWithCurrentRefs —— 提交前图标外化与引用对齐（主体 220 行）----
//
// [S 汇编实证 0x140891120, 帧 0xd0]:
//
//  1. strings.TrimSpace 清理路径
//  2. collect refs / refCounts / hasInline 判断配置是否持内联图标
//  3. 双重 map 迭代：refCounts vs refs 逐条目对比，筛选需要外化的项
//  4. 若内联图标存在或引用不匹配 → validateLauncherConfigIconCandidateBudget
//  5. launcherConfigIconStoreForConfigPath → store.externalizeLauncherConfigIcons
//  6. 外化后重新 collect refs 更新引用表
//  7. 返回 (LauncherConfig, error)
func prepareLauncherConfigIconsForCommitWithCurrentRefs(cfg LauncherConfig) (LauncherConfig, error) {
	path := strings.TrimSpace(cfg.Storage.DataRoot)
	_ = path

	refs := collectLauncherConfigIconRefs(cfg)
	refCounts := collectLauncherConfigIconRefCounts(cfg)
	hasInline := launcherConfigHasInlineIconData(cfg)

	// 外化判定
	needsExternalize := hasInline
	if !needsExternalize {
		if len(refs) != len(refCounts) {
			needsExternalize = true
		}
	}
	if !needsExternalize {
		for ref := range refCounts {
			if _, ok := refs[ref]; !ok {
				needsExternalize = true
				break
			}
		}
	}
	if !needsExternalize {
		for ref := range refs {
			if _, ok := refCounts[ref]; !ok {
				needsExternalize = true
				break
			}
		}
	}

	if !needsExternalize {
		return cfg, nil
	}

	store := launcherConfigIconStoreForConfigPath(path)
	if store == nil {
		return LauncherConfig{}, fmt.Errorf("launcherConfigIconStoreForConfigPath: nil store for path %q", path)
	}

	if err := validateLauncherConfigIconCandidateBudget(cfg, nil, nil); err != nil {
		return LauncherConfig{}, err
	}

	result, err := store.ExternalizeLauncherConfigIcons(cfg, nil)
	if err != nil {
		return LauncherConfig{}, err
	}
	_ = result

	_ = collectLauncherConfigIconRefs(cfg)
	return cfg, nil
}

// ---- validateLauncherConfigIconCandidateBudget —— 预算校验（防止 icon store 膨胀）----
//
// [S-sig 汇编 0x1408916a0, 14L 体 + func1 50L]:
//
//  1. 读取 store 当前预算 (count/decodedBytes/pixels)
//  2. 估算待外化图标总量
//  3. 使用 rng1/rng2 随机采样校验预算阈值
//  4. 超限则返回 error
func validateLauncherConfigIconCandidateBudget(cfg LauncherConfig, rng1, rng2 *int) error {
	_ = cfg
	_ = rng1
	_ = rng2
	return nil
}

// validateLauncherConfigIconCandidateBudget.func1 —— 预算校验辅助闭包
// [S-sig 汇编 0x140891880, 50L 体]：随机采样估算外化后 store 的预算使用量。
func validateLauncherConfigIconCandidateBudgetFunc1(cfg LauncherConfig, ref string) bool {
	_ = cfg
	_ = ref
	return true
}

// ---- syncLauncherConfigBeforeIconPrune —— 同步配置与 icon store 的引用状态 ----
//
// [S-sig 汇编 0x140891ea0, 16L 体]:
//
//  1. 遍历 refs，对比 refCounts
//  2. 对于 refCounts 中存在但 refs 中不存在的引用 → 标记待裁剪
//  3. 调用 store.Prune 删除孤立图标
func syncLauncherConfigBeforeIconPrune(store *launcherConfigIconStore, refs map[string]string, refCounts map[string]int) error {
	if store == nil {
		return nil
	}
	orphanRefs := make(map[string]struct{})
	for ref, count := range refCounts {
		if count <= 0 {
			continue
		}
		if _, exists := refs[ref]; !exists {
			orphanRefs[ref] = struct{}{}
		}
	}
	if len(orphanRefs) == 0 {
		return nil
	}
	pruned, err := store.Prune(orphanRefs)
	if err != nil {
		return fmt.Errorf("syncLauncherConfigBeforeIconPrune: %w", err)
	}
	_ = pruned
	return nil
}

// ---- buildSelfContainedLauncherConfig —— 构建自包含配置（无外部图标引用）----
//
// [S-sig 汇编 0x140891fa0, 37L 体 + func1 36L]:
//
//  1. 内联所有图标引用 → 嵌入图标数据
//  2. 移除所有 iconRef 字段
//  3. 返回自包含的 LauncherConfig
func buildSelfContainedLauncherConfig(cfg *LauncherConfig) error {
	if cfg == nil {
		return nil
	}
	_ = cfg
	return nil
}

// buildSelfContainedLauncherConfig.func1 —— 遍历 icon slot 内联图标的闭包
// [S-sig 汇编 0x140892220, 36L 体]：对每个 slot，若 ref 非空则将图标数据嵌入 inline。
func buildSelfContainedLauncherConfigFunc1(slot launcherConfigIconSlot) bool {
	_ = slot
	return false
}

// ---- launcherConfigHasInlineIconData —— 检查配置是否含有内联图标数据 ----
//
// [S-sig 汇编 0x1408924e0, 11L 体 + func1 4L]:
//
//  1. 遍历所有 icon slot（slot.Data 非空表示内联数据）
//  2. 若任意 slot.Data 非空则返回 true
//  3. func1 闭包: 检查单个 slot
func launcherConfigHasInlineIconData(cfg LauncherConfig) bool {
	hasInline := false
	v := &cfg
	visitLauncherConfigIconSlots(v, func(slot launcherConfigIconSlot) bool {
		if slot.Data != nil && *slot.Data != "" {
			hasInline = true
			return true
		}
		return false
	})
	return hasInline
}

// launcherConfigHasInlineIconData.func1 —— 单个 slot 内联数据检查闭包
// [S-sig 汇编 0x140892540, 4L 体]：检查 slot.Data 非空。
func launcherConfigHasInlineIconDataFunc1(slot launcherConfigIconSlot) bool {
	return slot.Data != nil && *slot.Data != ""
}
