// launchprivilege_test.go — 启动权限模式归一化测试（批次 122）
package main

import "testing"

func TestNormalizeAppLaunchPrivilegeMode(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"admin", "admin"},
		{"runas", "admin"},
		{"Administrator", "admin"}, // ToLower 生效
		{"standard", "standard"},
		{"normal", "standard"},
		{"unelevated", "standard"},
		{"follow", "followLauncher"},
		{"launcher", "followLauncher"},
		{"followlauncher", "followLauncher"},
		{"follow-launcher", "followLauncher"},
		{"default", "default"},
		{"", ""},
		{"  ", ""},
		{"unknown", ""},
	}
	for _, c := range cases {
		if got := normalizeAppLaunchPrivilegeMode(c.in); got != c.want {
			t.Fatalf("normalizeAppLaunchPrivilegeMode(%q)=%q，期望 %q", c.in, got, c.want)
		}
	}
}

func TestResolveEffectiveAppLaunchPrivilege(t *testing.T) {
	cases := []struct {
		configured, fallback string
		want                 string
	}{
		{"admin", "standard", "admin"},
		{"standard", "admin", "standard"},
		{"followlauncher", "x", "followLauncher"},
		{"unknown", "admin", "admin"}, // configured 无效 → fallback
		{"unknown", "follow", "followLauncher"},
		{"unknown", "unknown", "standard"}, // 都无效 → 默认
		{"default", "admin", "admin"},      // default 非有效特权 → fallback
	}
	for _, c := range cases {
		if got := resolveEffectiveAppLaunchPrivilege(c.configured, c.fallback); got != c.want {
			t.Fatalf("resolveEffectiveAppLaunchPrivilege(%q,%q)=%q，期望 %q",
				c.configured, c.fallback, got, c.want)
		}
	}
}
