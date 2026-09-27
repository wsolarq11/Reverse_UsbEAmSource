// AUTO-RECONSTRUCTED — DOMAIN: gesture / hot corner normalization
// 研究用途 · UsbEAm Launcher 1.0.3 后端方法体还原
//
// 本文件承载热角配置规范化链：
//
//	normalizeHotCornerConfig  [S]   汇编 0x1408dbaa0 全量实证
//	normalizeHotCornerRules   [S]   汇编 0x1408dbf00 全量实证
//	hasEnabledHotCornerRule   [S]   汇编 0x1407a0060 全量实证
//	normalizeHotCorner        [S]   汇编 0x1408dc520 全量实证
//	normalizeMouseGestureHotkey [S] 汇编 0x1408dc800 全量实证
package main

import "strings"

// normalizeHotCornerConfig 规范化热角配置。
// [S 汇编 0x1408dbaa0, 0x460] 实证流程：
//
//	TriggerSizePx：0→3（默认），<0→1，>64→64
//	TriggerDelayMs：0→0，<0→0，>10000→10000
//	CooldownMs：0→300（默认），<0→0，>60000→60000
//	AutoDisableFullscreen：incoming.AutoDisableFullscreen || isZero(incoming)
//	Corners：normalizeHotCornerRules(incoming.Corners)
//	Enabled：incoming.Enabled 原样透传
//
// isZero 判定六字段全零（Enabled/TriggerSizePx/TriggerDelayMs/CooldownMs/
// AutoDisableFullscreen 均 0 且 Corners 为空）。
func normalizeHotCornerConfig(cfg HotCornerConfig) HotCornerConfig {
	isZero := !cfg.Enabled &&
		cfg.TriggerSizePx == 0 &&
		cfg.TriggerDelayMs == 0 &&
		cfg.CooldownMs == 0 &&
		!cfg.AutoDisableFullscreen &&
		len(cfg.Corners) == 0

	triggerSizePx := cfg.TriggerSizePx
	switch {
	case triggerSizePx == 0:
		triggerSizePx = 3
	case triggerSizePx < 0:
		triggerSizePx = 1
	case triggerSizePx > 64:
		triggerSizePx = 64
	}

	triggerDelayMs := cfg.TriggerDelayMs
	switch {
	case triggerDelayMs < 0:
		triggerDelayMs = 0
	case triggerDelayMs > 10000:
		triggerDelayMs = 10000
	}

	cooldownMs := cfg.CooldownMs
	switch {
	case cooldownMs == 0:
		cooldownMs = 300
	case cooldownMs < 0:
		cooldownMs = 0
	case cooldownMs > 60000:
		cooldownMs = 60000
	}

	return HotCornerConfig{
		Enabled:               cfg.Enabled,
		TriggerSizePx:         triggerSizePx,
		TriggerDelayMs:        triggerDelayMs,
		CooldownMs:            cooldownMs,
		AutoDisableFullscreen: cfg.AutoDisableFullscreen || isZero,
		Corners:               normalizeHotCornerRules(cfg.Corners),
	}
}

// normalizeHotCornerRules 规范化热角规则列表。
// [S 汇编 0x1408dbf00, 0x660] 实证流程：
//
//	默认 4 角基底（topLeft/topRight/bottomLeft/bottomRight，bottom 两角默认启用
//	Win / Win+D）写入 map[string]HotCornerRule。
//	入参规则逐项 normalizeHotCorner（空则跳过）+ normalizeMouseGestureHotkey 后合并覆盖。
//	最后按默认 4 角顺序重建切片（len 恒为 4）。
func normalizeHotCornerRules(rules []HotCornerRule) []HotCornerRule {
	defaultRules := []HotCornerRule{
		{Corner: "topLeft", Enabled: false, Hotkey: ""},
		{Corner: "topRight", Enabled: false, Hotkey: ""},
		{Corner: "bottomLeft", Enabled: true, Hotkey: "Win"},
		{Corner: "bottomRight", Enabled: true, Hotkey: "Win+D"},
	}

	merged := make(map[string]HotCornerRule, len(defaultRules))
	for _, r := range defaultRules {
		merged[r.Corner] = r
	}
	for _, r := range rules {
		corner := normalizeHotCorner(r.Corner)
		if corner == "" {
			continue
		}
		merged[corner] = HotCornerRule{
			Corner:  corner,
			Enabled: r.Enabled,
			Hotkey:  normalizeMouseGestureHotkey(r.Hotkey),
		}
	}

	result := make([]HotCornerRule, 0, len(defaultRules))
	for _, r := range defaultRules {
		result = append(result, merged[r.Corner])
	}
	return result
}

// hasEnabledHotCornerRule 是否存在启用的热角规则（Hotkey TrimSpace 非空）。
// [S 汇编 0x1407a0060, 0x200]：normalizeHotCornerConfig 后遍历 Corners，
// 任一 rule.Enabled && TrimSpace(rule.Hotkey) != "" 即返回 true。
func hasEnabledHotCornerRule(cfg HotCornerConfig) bool {
	normalized := normalizeHotCornerConfig(cfg)
	for _, rule := range normalized.Corners {
		if rule.Enabled && strings.TrimSpace(rule.Hotkey) != "" {
			return true
		}
	}
	return false
}

// normalizeHotCorner 规范化热角名。
// [S 汇编 0x1408dc520, 0x2e0]：TrimSpace → ToLower → 按长度分派字节比较。
// 接受四种角名的缩写 / 连字符 / 词序颠倒变体，映射到规范 camelCase；未命中返回 ""。
func normalizeHotCorner(corner string) string {
	lower := strings.ToLower(strings.TrimSpace(corner))
	switch lower {
	case "lt", "topleft", "lefttop", "top-left", "left-top":
		return "topLeft"
	case "rt", "topright", "righttop", "top-right", "right-top":
		return "topRight"
	case "lb", "bottomleft", "leftbottom", "bottom-left", "left-bottom":
		return "bottomLeft"
	case "rb", "bottomright", "rightbottom", "bottom-right", "right-bottom":
		return "bottomRight"
	default:
		return ""
	}
}

// normalizeMouseGestureHotkey 规范化鼠标手势热键。
// [S 汇编 0x1408dc800, 0xe0]：TrimSpace → 空→""；
// EqualFold(trimmed,"win") 或 EqualFold(trimmed,"windows") → "Win"；
// 否则 normalizeShortcutBindingWithOptions(trimmed, pred, true, true, true)。
func normalizeMouseGestureHotkey(hotkey string) string {
	trimmed := strings.TrimSpace(hotkey)
	if trimmed == "" {
		return ""
	}
	if strings.EqualFold(trimmed, "win") || strings.EqualFold(trimmed, "windows") {
		return "Win"
	}
	return normalizeShortcutBindingWithOptions(trimmed, launcherHotkeyBindingPredicate, true, true, true)
}
