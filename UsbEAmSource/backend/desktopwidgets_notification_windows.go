// desktopwidgets_notification_windows.go — 桌面小部件提醒通知（逆向还原）
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)

package main

import (
	"strings"
)

// normalizeDesktopWidgetNotificationText 规范化提醒文本。
// [S] 反汇编实证 0x1407aca60, 288B：
//
//	TrimSpace → strings.Map（控制字符折叠空格）→ strings.Fields → strings.Join(" ")；
//	转 []rune 后截断到 maxLen（越界则 slice[:maxLen]）；空串返回 fallback。
//	映射闭包（0x1409f7ba0）：r=='\r'/'\n'/'\t'/r<' '/r==0x7f → ' '。
func normalizeDesktopWidgetNotificationText(text string, maxLen int, fallback string) string {
	trimmed := strings.TrimSpace(text)
	mapped := strings.Map(func(r rune) rune {
		if r == '\r' || r == '\n' || r == '\t' || r < ' ' || r == 0x7f {
			return ' '
		}
		return r
	}, trimmed)
	joined := strings.Join(strings.Fields(mapped), " ")
	runes := []rune(joined)
	if len(runes) > maxLen {
		runes = runes[:maxLen]
	}
	if len(runes) == 0 {
		return fallback
	}
	return string(runes)
}

// showDesktopWidgetNotification 显示桌面小部件提醒通知。
// [S] 反汇编实证 0x1407ac960, 256B：
//
//	title 经 normalize(title, 120, "UsbEAm Launcher")、body 经 normalize(body, 240, "提醒时间已到")；
//	silent=true → sound="silent"，否则 "ms-winsoundevent:Notification.Default"；
//	随后 showLauncherNotificationWithAudio(title, body, sound)（返回值丢弃）。
func showDesktopWidgetNotification(title, body string, silent bool) {
	t := normalizeDesktopWidgetNotificationText(title, 120, "UsbEAm Launcher")
	m := normalizeDesktopWidgetNotificationText(body, 240, "提醒时间已到")
	sound := "ms-winsoundevent:Notification.Default"
	if silent {
		sound = "silent"
	}
	showLauncherNotificationWithAudio(t, m, sound)
}
