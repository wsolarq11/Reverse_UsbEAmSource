// AUTO-RECONSTRUCTED — DOMAIN: gpu (gpu.go / gpu_pick_debug.go)
// 研究用途
//
// 本文件落地 gpu 域纯逻辑函数（无外部 IO / 全局态），全部为 [S] 反汇编实证还原。
// 依赖类型：types_gpu.go（GPUPreferenceSetting / GPUPreferenceEntry / GPUPreferenceState）。
//
// 反汇编依据（docs/goresym/pipeline/tmp/*.asm.txt）：
// 寄存器规约：string = (ptr AX, len BX)；[]T = (ptr, len, cap)；mapaccess2_faststr 命中在 BL；
// runtime.memequal 判 string 相等；strings.EqualFold/ToLower/TrimSpace/Replace/Cut/IndexAny/Split。
//
// 蓝图行号引用自 docs/goresym/source_funcs.txt。
package main

import (
	"path/filepath"
	"strconv"
	"strings"
)

// normalizeGPUPreferenceMode 把任意用户/注册表 GPU 偏好表述规范化为 4 个标准值之一。
// [S 汇编 0x14085a0e0, 141L]（gpu.go 蓝图 84-121）：TrimSpace → ToLower → 按长度跳表
// （jmp [rcx+rbx*8]，表 @0x1411e0940，19 项）匹配已知名，命中即返回标准串，否则 "unknown"。
// 长度映射与魔数（little-endian）：
//
//	""/ "0"(0x30) / "system"(0x74737973+0x6d65) / "default"(0x61666564+0x6c75+0x74) /
//	"unspecified"(0x6669636570736e75+0x6569+0x64) / "systemdefault"(0x65646d6574737973+0x6c756166+0x74) /
//	"letwindowsdecide"(0x6f646e697774656c+0x6564696365647377) /
//	"let-windows-decide"(18B memequal)                          → "systemDefault"
//	"1"(0x31) / "powersaving"(0x7661737265776f70+0x6e69+0x67) /
//	"minimumpower"(0x706d756d696e696d+0x7265776f) / "power-saving"(0x61732d7265776f70+0x676e6976) /
//	"minimum-power"(0x2d6d756d696e696d+0x65776f70+0x72)          → "powerSaving"
//	"2"(0x32) / "highperformance"(0x6672657068676968+0x616d726f+0x636e+0x65) /
//	"high-performance"(0x7265702d68676968+0x65636e616d726f66)    → "highPerformance"
func normalizeGPUPreferenceMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "0", "system", "default", "unspecified", "systemdefault", "letwindowsdecide", "let-windows-decide":
		return "systemDefault"
	case "1", "powersaving", "minimumpower", "power-saving", "minimum-power":
		return "powerSaving"
	case "2", "highperformance", "high-performance":
		return "highPerformance"
	default:
		return "unknown"
	}
}

// gpuPreferenceCanonicalSettingKey 规范化设置键名：已知键 → 全小写规范形；其余键 → TrimSpace 原样。
// [S 汇编 0x14085a360]（gpu.go 蓝图 121-136）：ToLower(TrimSpace(key)) 与 4 个已知键
// （GpuPreference/AutoHDREnable/SwapEffectUpgradeEnable/DXGIEffects，均 EqualFold 判等）逐一比对；
// 命中返回 ToLower 结果，未命中返回 TrimSpace(原始 key)。asm 中命中的返回指针与比较常量地址相同，
// 证得「返回小写化已知键」而非「返回常量本身」——两者值等价，Go 源即 `return t`。
func gpuPreferenceCanonicalSettingKey(key string) string {
	t := strings.ToLower(strings.TrimSpace(key))
	switch t {
	case "gpupreference", "autohdrenable", "swapeffectupgradeenable", "dxgieffects":
		return t
	default:
		return strings.TrimSpace(key)
	}
}

// gpuPreferenceNormalizeSettingValue 按键语义规范化设置值（单 string 返回）。
// [S 汇编 0x14085a4e0, 153L]（gpu.go 蓝图 136-166）：
//
//	ReplaceAll(value, ";", "") → TrimSpace；空 → ""
//	EqualFold(key,"GpuPreference")：normalizeGPUPreferenceMode(v)
//	  "unknown" → Atoi(v)，err 或 n<0 → "0"，否则 Itoa(n)
//	  "powerSaving"→"1"，"highPerformance"→"2"，其余→"0"
//	EqualFold(key,"AutoHDREnable")：v=="0" → "0"，否则 → "1"
//	其余 → v（原样）
//
// 魔数：';'=0x1411cac10、'0'=0x1411ca3c8、'1'=0x1411cac98；GpuPreference 13B、AutoHDREnable 13B。
func gpuPreferenceNormalizeSettingValue(key, value string) string {
	v := strings.TrimSpace(strings.ReplaceAll(value, ";", ""))
	if v == "" {
		return ""
	}
	if strings.EqualFold(key, "GpuPreference") {
		m := normalizeGPUPreferenceMode(v)
		if m == "unknown" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 {
				return "0"
			}
			return strconv.Itoa(n)
		}
		switch m {
		case "powerSaving":
			return "1"
		case "highPerformance":
			return "2"
		default:
			return "0"
		}
	}
	if strings.EqualFold(key, "AutoHDREnable") {
		if v == "0" {
			return "0"
		}
		return "1"
	}
	return v
}

// normalizeGPUPreferenceRegistrySettings 规范化设置切片：去重（键小写为索引）、
// 过滤非法键/空值，并将 GpuPreference 项前移。
// [S 汇编 0x14085a700, 273L]（gpu.go 蓝图 166-210）：
//
//	空输入 → nil；result=make(...,0,len)；seen=map[string]int
//	遍历：key=gpuPreferenceCanonicalSettingKey(s.Key)
//	  key=="" 或 IndexAny(key,"=;")>=0 → 跳过（0x140c336a1="=;"）
//	  value=gpuPreferenceNormalizeSettingValue(key, s.Value)；"" → 跳过
//	  lkey=ToLower(key)；seen[lkey] 命中 → result[idx] 覆写；否则 seen[lkey]=len 并 append
//	收尾：seen["gpupreference"] 存在且 idx>0 → 把该项移到 result[0]
//	  （copy(result[1:idx+1], result[:idx]) + 首项回填，见 typedslicecopy）
func normalizeGPUPreferenceRegistrySettings(settings []GPUPreferenceSetting) []GPUPreferenceSetting {
	if len(settings) == 0 {
		return nil
	}
	result := make([]GPUPreferenceSetting, 0, len(settings))
	seen := make(map[string]int)
	for _, s := range settings {
		key := gpuPreferenceCanonicalSettingKey(s.Key)
		if key == "" || strings.IndexAny(key, "=;") >= 0 {
			continue
		}
		value := gpuPreferenceNormalizeSettingValue(key, s.Value)
		if value == "" {
			continue
		}
		lkey := strings.ToLower(key)
		if idx, ok := seen[lkey]; ok {
			result[idx] = GPUPreferenceSetting{Key: key, Value: value}
		} else {
			seen[lkey] = len(result)
			result = append(result, GPUPreferenceSetting{Key: key, Value: value})
		}
	}
	if idx, ok := seen["gpupreference"]; ok && idx > 0 {
		entry := result[idx]
		copy(result[1:idx+1], result[:idx])
		result[0] = entry
	}
	return result
}

// ensureGPUPreferenceRegistrySettings 确保设置中存在 GpuPreference 项，缺省前插 {GpuPreference, "0"}。
// [S 汇编 0x14085ac40, 117L]（gpu.go 蓝图 210-221）：
//
//	normalizeGPUPreferenceRegistrySettings → 遍历 EqualFold(Key,"GpuPreference") 命中即原样返回；
//	否则前插（PREPEND，非尾插）：newobject 造出字面量 {Key:"GpuPreference", Value:"0"}
//	（[rax]=0x140c4e04e, [rax+8]=0xd, [rax+0x10]=0x1411ca3c8, [rax+0x18]=1）→
//	growslice 以该字面量为 1 元素旧 slice 扩展 → typedslicecopy 把 s 整体右移一位到 [1..]
//	（dst=新指针+0x20，证得前插而非追加）。
func ensureGPUPreferenceRegistrySettings(settings []GPUPreferenceSetting) []GPUPreferenceSetting {
	s := normalizeGPUPreferenceRegistrySettings(settings)
	for _, e := range s {
		if strings.EqualFold(e.Key, "GpuPreference") {
			return s
		}
	}
	return append([]GPUPreferenceSetting{{Key: "GpuPreference", Value: "0"}}, s...)
}

// gpuPreferenceModeFromSettings 从设置切片解析偏好模式与数值。
// [S 汇编 0x14085ade0]（gpu.go 蓝图 221-238）：找首个 EqualFold(Key,"GpuPreference") 项，
// Atoi(TrimSpace(Value))：0→("systemDefault",0)、1→("powerSaving",1)、2→("highPerformance",2)、
// 其他→("unknown",n)；Atoi 错→("unknown",0)；未找到→("systemDefault",0)。
func gpuPreferenceModeFromSettings(settings []GPUPreferenceSetting) (string, int) {
	for _, s := range settings {
		if strings.EqualFold(s.Key, "GpuPreference") {
			n, err := strconv.Atoi(strings.TrimSpace(s.Value))
			if err != nil {
				return "unknown", 0
			}
			switch n {
			case 0:
				return "systemDefault", 0
			case 1:
				return "powerSaving", 1
			case 2:
				return "highPerformance", 2
			default:
				return "unknown", n
			}
		}
	}
	return "systemDefault", 0
}

// parseGPUPreferenceRegistryValue 解析注册表原始串为设置切片。
// [S 汇编 0x14085af40, 180L]（gpu.go 蓝图 238-265）：
//
//	TrimSpace(raw)=="" → nil；parts=strings.Split(raw,";")
//	result（len>1 时 make(...,0,len)）→ 遍历 parts：
//	  TrimSpace 跳空 → Cut(part,"=") 不命中则跳过 → append({before, after})
//	→ normalizeGPUPreferenceRegistrySettings(result)
func parseGPUPreferenceRegistryValue(raw string) []GPUPreferenceSetting {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ";")
	var result []GPUPreferenceSetting
	if len(parts) > 1 {
		result = make([]GPUPreferenceSetting, 0, len(parts))
	}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		result = append(result, GPUPreferenceSetting{Key: key, Value: value})
	}
	return normalizeGPUPreferenceRegistrySettings(result)
}

// serializeGPUPreferenceRegistrySettings 把设置切片序列化为注册表串（Key=Value; 拼接）。
// [S 汇编 0x14085b240, 236L]（gpu.go 蓝图 265-286）：
//
//	normalizeGPUPreferenceRegistrySettings；空 → ""
//	遍历：Key=="" 或 Value=="" 跳过 → append(Key) + '='(0x3d) + append(Value) + ';'(0x3b)
//	→ string(buf)（尾部 string 转换安全校验 panicunsafestringlen/nilptr）
func serializeGPUPreferenceRegistrySettings(settings []GPUPreferenceSetting) string {
	s := normalizeGPUPreferenceRegistrySettings(settings)
	if len(s) == 0 {
		return ""
	}
	var b []byte
	for _, e := range s {
		if e.Key == "" || e.Value == "" {
			continue
		}
		b = append(b, e.Key...)
		b = append(b, '=')
		b = append(b, e.Value...)
		b = append(b, ';')
	}
	return string(b)
}

// normalizeGPUPreferencePath 规范化 GPU 偏好路径。
// [S 汇编 0x14085b600]（gpu.go 蓝图 286-300）：TrimSpace → 空返 "" →
// ReplaceAll("/", "\\") → filepath.Clean → 结果 "." 返 ""，否则返原串。
func normalizeGPUPreferencePath(path string) string {
	p := strings.TrimSpace(path)
	if p == "" {
		return ""
	}
	p = strings.ReplaceAll(p, "/", "\\")
	p = filepath.Clean(p)
	if p == "." {
		return ""
	}
	return p
}

// gpuPreferencePathKey 生成路径归一化小写键。
// [S 汇编 0x14085b6a0]（gpu.go 蓝图 300-306）：normalizeGPUPreferencePath → 空返 "" → ToLower。
func gpuPreferencePathKey(path string) string {
	p := normalizeGPUPreferencePath(path)
	if p == "" {
		return ""
	}
	return strings.ToLower(p)
}

// parseGPUPickDebugBool 解析调试布尔开关。
// [S 汇编 0x14085bd40, 72L]（gpu_pick_debug.go 蓝图 134-141）：ToLower(TrimSpace) 后按长度匹配：
//
//	"1"/"on"(0x6e6f)/"yes"(0x6579+0x73)/"true"(0x65757274)          → (true, true)
//	"0"/"no"(0x6f6e)/"off"(0x666f+0x66)/"false"(0x736c6166+0x65)   → (false, true)
//	其余 → (false, false)
//
// 返回值 (bool, bool) 由三个出口的寄存器对证得：(1,1)/(0,1)/(0,0)，第二值区分「已识别为假」与「未识别」。
func parseGPUPickDebugBool(s string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "on", "yes", "true":
		return true, true
	case "0", "no", "off", "false":
		return false, true
	default:
		return false, false
	}
}
