// gpu_test.go — GPU 偏好域纯逻辑函数行为测试（批次 153）
// 锁定 [S] 反汇编实证函数的黄金用例：normalizeGPUPreferenceMode 跳表、
// 键/值/路径规范化、注册表串解析/序列化、去重与前移语义。
package main

import "testing"

func TestNormalizeGPUPreferenceMode(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", "systemDefault"},
		{"0", "systemDefault"},
		{"1", "powerSaving"},
		{"2", "highPerformance"},
		{"system", "systemDefault"},
		{"SYSTEM", "systemDefault"},
		{"default", "systemDefault"},
		{"unspecified", "systemDefault"},
		{"systemdefault", "systemDefault"},
		{"powersaving", "powerSaving"},
		{"PowerSaving", "powerSaving"},
		{"minimumpower", "powerSaving"},
		{"power-saving", "powerSaving"},
		{"minimum-power", "powerSaving"},
		{"highperformance", "highPerformance"},
		{"high-performance", "highPerformance"},
		{"letwindowsdecide", "systemDefault"},
		{"let-windows-decide", "systemDefault"},
		{"  Let-Windows-Decide  ", "systemDefault"},
		{"let windows decide", "unknown"}, // 空格分隔形式未在跳表中，非有效键
		{"bogus", "unknown"},
		{"3", "unknown"},
	}
	for _, c := range cases {
		if got := normalizeGPUPreferenceMode(c.in); got != c.want {
			t.Fatalf("normalizeGPUPreferenceMode(%q)=%q，期望 %q", c.in, got, c.want)
		}
	}
}

func TestGPUPreferenceCanonicalSettingKey(t *testing.T) {
	cases := []struct{ in, want string }{
		{"GpuPreference", "gpupreference"},
		{"gpupreference", "gpupreference"},
		{"GPUPREFERENCE", "gpupreference"},
		{"AutoHDREnable", "autohdrenable"},
		{"SwapEffectUpgradeEnable", "swapeffectupgradeenable"},
		{"DXGIEffects", "dxgieffects"},
		{"CustomKey", "CustomKey"}, // 未知键 TrimSpace 原样（非小写）
		{"  x  ", "x"},
	}
	for _, c := range cases {
		if got := gpuPreferenceCanonicalSettingKey(c.in); got != c.want {
			t.Fatalf("gpuPreferenceCanonicalSettingKey(%q)=%q，期望 %q", c.in, got, c.want)
		}
	}
}

func TestGPUPreferenceNormalizeSettingValue(t *testing.T) {
	cases := []struct{ key, val, want string }{
		{"GpuPreference", "0", "0"},
		{"GpuPreference", "1", "1"},
		{"GpuPreference", "2", "2"},
		{"GpuPreference", "systemDefault", "0"},
		{"GpuPreference", "powerSaving", "1"},
		{"GpuPreference", "highPerformance", "2"},
		{"GpuPreference", "unknown-value", "0"}, // Atoi 失败
		{"GpuPreference", "-1", "0"},            // 负数
		{"GpuPreference", "5", "5"},             // 未知模式数字透传 Itoa
		{"GpuPreference", "", ""},
		{"AutoHDREnable", "0", "0"},
		{"AutoHDREnable", "1", "1"},
		{"AutoHDREnable", "true", "1"},
		{"DXGIEffects", "abc", "abc"}, // 非特殊键原样
		{"GpuPreference", "2;", "2"},  // 分号被剥离
	}
	for _, c := range cases {
		if got := gpuPreferenceNormalizeSettingValue(c.key, c.val); got != c.want {
			t.Fatalf("gpuPreferenceNormalizeSettingValue(%q,%q)=%q，期望 %q", c.key, c.val, got, c.want)
		}
	}
}

func TestGPUPreferenceModeFromSettings(t *testing.T) {
	cases := []struct {
		settings []GPUPreferenceSetting
		wantMode string
		wantN    int
	}{
		{[]GPUPreferenceSetting{{Key: "GpuPreference", Value: "0"}}, "systemDefault", 0},
		{[]GPUPreferenceSetting{{Key: "GpuPreference", Value: "1"}}, "powerSaving", 1},
		{[]GPUPreferenceSetting{{Key: "GpuPreference", Value: "2"}}, "highPerformance", 2},
		{[]GPUPreferenceSetting{{Key: "GpuPreference", Value: "7"}}, "unknown", 7},
		{[]GPUPreferenceSetting{{Key: "GpuPreference", Value: "x"}}, "unknown", 0},
		{nil, "systemDefault", 0},
		{[]GPUPreferenceSetting{{Key: "AutoHDREnable", Value: "1"}}, "systemDefault", 0},
	}
	for _, c := range cases {
		m, n := gpuPreferenceModeFromSettings(c.settings)
		if m != c.wantMode || n != c.wantN {
			t.Fatalf("gpuPreferenceModeFromSettings(%v)=(%q,%d)，期望 (%q,%d)", c.settings, m, n, c.wantMode, c.wantN)
		}
	}
}

func TestParseAndSerializeGPUPreferenceRegistrySettings(t *testing.T) {
	// 解析：分号分段、等号切键值、空段跳越、去重（小写键）、GpuPreference 前移。
	parsed := parseGPUPreferenceRegistryValue("GpuPreference=1;AutoHDREnable=0; ;dxgieffects=abc")
	if len(parsed) != 3 {
		t.Fatalf("解析后条目数=%d，期望 3：%v", len(parsed), parsed)
	}
	// GpuPreference 项前移到首位。
	if parsed[0].Key != "gpupreference" || parsed[0].Value != "1" {
		t.Fatalf("首项=%v，期望 {gpupreference 1}", parsed[0])
	}

	// 序列化：Key=Value; 拼接，空键/空值被跳过。
	serialized := serializeGPUPreferenceRegistrySettings(parsed)
	if serialized != "gpupreference=1;autohdrenable=0;dxgieffects=abc;" {
		t.Fatalf("序列化=%q", serialized)
	}

	// 往返稳定。
	reparsed := parseGPUPreferenceRegistryValue(serialized)
	if len(reparsed) != 3 || reparsed[0].Key != "gpupreference" {
		t.Fatalf("往返失败：%v", reparsed)
	}
}

func TestEnsureGPUPreferenceRegistrySettings(t *testing.T) {
	// 已有 GpuPreference 项则原样返回。
	in := []GPUPreferenceSetting{{Key: "GpuPreference", Value: "2"}}
	out := ensureGPUPreferenceRegistrySettings(in)
	if len(out) != 1 || out[0].Value != "2" {
		t.Fatalf("已有项返回=%v", out)
	}
	// 无 GpuPreference 项则前插 {GpuPreference, "0"}（字面量为原始大小写，未过规范化）。
	out2 := ensureGPUPreferenceRegistrySettings([]GPUPreferenceSetting{{Key: "AutoHDREnable", Value: "1"}})
	if len(out2) != 2 || out2[0].Key != "GpuPreference" || out2[0].Value != "0" {
		t.Fatalf("补缺返回=%v", out2)
	}
}

func TestNormalizeGPUPreferencePath(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"   ", ""},
		{"C:\\Games\\App.exe", `C:\Games\App.exe`},
		{"C:/Games/App.exe", `C:\Games\App.exe`},
		{"C:\\Games\\..\\App.exe", `C:\App.exe`},
		{".", ""},
		{"C:\\", `C:\`},
	}
	for _, c := range cases {
		if got := normalizeGPUPreferencePath(c.in); got != c.want {
			t.Fatalf("normalizeGPUPreferencePath(%q)=%q，期望 %q", c.in, got, c.want)
		}
	}
}

func TestGPUPreferencePathKey(t *testing.T) {
	if got := gpuPreferencePathKey("C:\\Games\\App.EXE"); got != `c:\games\app.exe` {
		t.Fatalf("gpuPreferencePathKey=%q", got)
	}
	if got := gpuPreferencePathKey("  "); got != "" {
		t.Fatalf("空白路径键=%q，期望空", got)
	}
}

func TestParseGPUPickDebugBool(t *testing.T) {
	cases := []struct {
		in      string
		wantVal bool
		wantOK  bool
	}{
		{"1", true, true},
		{"on", true, true},
		{"ON", true, true},
		{"yes", true, true},
		{"true", true, true},
		{"0", false, true},
		{"no", false, true},
		{"off", false, true},
		{"false", false, true},
		{"maybe", false, false},
		{"", false, false},
	}
	for _, c := range cases {
		val, ok := parseGPUPickDebugBool(c.in)
		if val != c.wantVal || ok != c.wantOK {
			t.Fatalf("parseGPUPickDebugBool(%q)=(%v,%v)，期望 (%v,%v)", c.in, val, ok, c.wantVal, c.wantOK)
		}
	}
}
