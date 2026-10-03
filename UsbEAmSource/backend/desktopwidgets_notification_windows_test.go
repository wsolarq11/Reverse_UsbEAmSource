package main

import (
	"testing"
)

func TestNormalizeDesktopWidgetNotificationText(t *testing.T) {
	cases := []struct {
		name     string
		text     string
		maxLen   int
		fallback string
		want     string
	}{
		{"control-chars-fold", "a\r\nb\tc", 120, "fb", "a b c"},
		{"whitespace-collapse", "  a   b  ", 120, "fb", "a b"},
		{"empty-fallback", "", 120, "fb", "fb"},
		{"blank-fallback", "   ", 120, "fb", "fb"},
		{"truncate-ascii", "abcde", 3, "fb", "abc"},
		{"truncate-cjk", "提醒时间已到", 2, "fb", "提醒"},
		{"del-fold", "a\x7fb", 120, "fb", "a b"},
		{"no-truncate", "hello world", 120, "fb", "hello world"},
	}
	for _, c := range cases {
		got := normalizeDesktopWidgetNotificationText(c.text, c.maxLen, c.fallback)
		if got != c.want {
			t.Errorf("%s: normalize(%q, %d, %q) = %q, want %q", c.name, c.text, c.maxLen, c.fallback, got, c.want)
		}
	}
}

func TestShowDesktopWidgetNotificationSmoke(t *testing.T) {
	// 纯编排：不触发 panic 即可（下游 showLauncherNotificationWithAudio 为骨架）。
	showDesktopWidgetNotification("title", "body", false)
	showDesktopWidgetNotification("title", "body", true)
}
