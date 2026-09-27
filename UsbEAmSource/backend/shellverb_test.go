package main

import (
	"testing"
)

func TestLauncherStartedForStartupTray(t *testing.T) {
	cases := []struct {
		args []string
		want bool
	}{
		{[]string{"--usbeam-startup-tray"}, true},
		{[]string{"--USBEAM-STARTUP-TRAY"}, true},
		{[]string{"  --usbeam-startup-tray  "}, true},
		{[]string{"--normal", "--usbeam-startup-tray", "x"}, true},
		{[]string{"--normal", "x"}, false},
		{nil, false},
	}
	for _, c := range cases {
		if got := launcherStartedForStartupTray(c.args); got != c.want {
			t.Errorf("launcherStartedForStartupTray(%v) = %v, want %v", c.args, got, c.want)
		}
	}
}

func TestInvokeShellVerbEmpty(t *testing.T) {
	if err := invokeShellVerb("", "x"); err == nil || err.Error() != "路径或动作不能为空" {
		t.Fatalf("空 verb 应返回 路径或动作不能为空, got %v", err)
	}
	if err := invokeShellVerb("open", "  "); err == nil || err.Error() != "路径或动作不能为空" {
		t.Fatalf("空 path 应返回 路径或动作不能为空, got %v", err)
	}
}
