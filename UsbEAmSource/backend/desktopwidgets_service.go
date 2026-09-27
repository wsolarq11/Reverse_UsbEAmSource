package main

import "strings"

// normalizeDesktopWidgetTitle 归一化首页组件标题：TrimSpace 后按 rune 截断至 160，
// 空标题回退到按组件种类（kind）确定的英文默认标题。
// [S 汇编 0x1407bef80, 576B] 实证：
//
//	[]rune(TrimSpace(title))，len>160 截断到 160；len>0 → string(runes)；
//	否则按 kind 逐字节 cmp 返回默认标题：
//	  "note"→"Note"、"clock"→"Clock"、"timer"→"Timer"、"weather"→"Weather"、
//	  "calendar"→"Calendar"、"reminder"→"Reminder"、"stopwatch"→"Stopwatch"、
//	  "worldClock"→"World Clock"、其余→"Widget"。
func normalizeDesktopWidgetTitle(title, kind string) string {
	runes := []rune(strings.TrimSpace(title))
	if len(runes) > 160 {
		runes = runes[:160]
	}
	if len(runes) > 0 {
		return string(runes)
	}
	switch kind {
	case "note":
		return "Note"
	case "clock":
		return "Clock"
	case "timer":
		return "Timer"
	case "weather":
		return "Weather"
	case "calendar":
		return "Calendar"
	case "reminder":
		return "Reminder"
	case "stopwatch":
		return "Stopwatch"
	case "worldClock":
		return "World Clock"
	default:
		return "Widget"
	}
}

// normalizeDesktopReminderScheduleKind 把喝水/久坐两类健康提醒的调度种类统一归一到
// "interval"，其余按 TrimSpace 原样返回。
// [S 汇编 0x1407bf740, 160B] 实证：
//
//	TrimSpace 后 == "water"（len5：dword 0x65746177="wate"+'r'）或
//	=="sedentary"（len9：qword 0x7261746e65646573="sedentar"+'y'）→ "interval"（8B @0x140c3c164）；
//	否则返回 TrimSpace(kind)。
func normalizeDesktopReminderScheduleKind(kind string) string {
	t := strings.TrimSpace(kind)
	if t == "water" || t == "sedentary" {
		return "interval"
	}
	return t
}

// normalizeDesktopReminderTimezone 归一化首页组件配置里的提醒定义：仅 kind=="reminder" 且
// cfg.Reminder 非空时拷贝一份并修正 ScheduleKind / Preset / TimezoneID / 音频，否则原样返回 cfg。
// [S 汇编 0x1407bf220, 896B] 实证：
//
//	kind=="reminder"（len8 + "reminder"@0x7265646e696d6572）且 cfg.Reminder 非空才归一化；
//	def=*cfg.Reminder 拷贝（newobject type size=0xb0 @0x140bfe260 = DesktopReminderDefinition）；
//	ScheduleKind=normalizeDesktopReminderScheduleKind(...)；
//	Preset=="water"/"sedentary"→"custom"（6B @0x140c37632）；TimezoneID="Local"（5B @0x140c35c3b）；
//	音频=normalizeDesktopReminderAudio(AudioMode,AudioPath,AudioRepeatCount)；cfg.Reminder=&def。
func normalizeDesktopReminderTimezone(kind string, cfg DesktopWidgetConfig) DesktopWidgetConfig {
	if kind != "reminder" || cfg.Reminder == nil {
		return cfg
	}
	def := *cfg.Reminder
	def.ScheduleKind = normalizeDesktopReminderScheduleKind(cfg.Reminder.ScheduleKind)
	if def.Preset == "water" || def.Preset == "sedentary" {
		def.Preset = "custom"
	}
	def.TimezoneID = "Local"
	def.AudioMode, def.AudioPath, def.AudioRepeatCount = normalizeDesktopReminderAudio(
		cfg.Reminder.AudioMode, cfg.Reminder.AudioPath, cfg.Reminder.AudioRepeatCount)
	cfg.Reminder = &def
	return cfg
}

// normalizeDesktopWidgetConfig 归一化首页组件配置：先走提醒归一化，再把 clock 的 12/24 小时制
// 归一到 "12"/"24"。
// [S 汇编 0x1407bf5a0, 416B] 实证：
//
//	cfg=normalizeDesktopReminderTimezone(kind,cfg)；
//	kind=="clock"（len5 + "cloc"@0x636f6c63+'k'）时 TrimSpace(cfg.HourFormat)=="12"（0x3231）→"12"，
//	否则→"24"；其余 kind 不改 HourFormat。
func normalizeDesktopWidgetConfig(kind string, cfg DesktopWidgetConfig) DesktopWidgetConfig {
	cfg = normalizeDesktopReminderTimezone(kind, cfg)
	if kind == "clock" {
		if strings.TrimSpace(cfg.HourFormat) == "12" {
			cfg.HourFormat = "12"
		} else {
			cfg.HourFormat = "24"
		}
	}
	return cfg
}
