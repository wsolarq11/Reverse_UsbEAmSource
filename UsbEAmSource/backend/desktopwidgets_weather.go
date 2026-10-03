package main

import (
	"math"
	"strconv"
	"strings"
	"time"
)

// desktopWeatherInt 将天气字符串解析为整数（TrimSpace 后 Atoi）。
// [S] ASM 0x1407c9640: strings.TrimSpace → strconv.Atoi，直接返回 (int, error)。
func desktopWeatherInt(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}

// Shutdown 取消天气服务的后台任务。
// [S] ASM 0x1407c5180: s==nil → return；s.cancel==nil → return；否则 s.cancel()。
func (s *desktopWidgetWeatherService) Shutdown() {
	if s == nil {
		return
	}
	if s.cancel != nil {
		s.cancel()
	}
}

// desktopWeatherFloat 将天气字符串解析为浮点数（TrimSpace 后 ParseFloat）。
// [S] ASM 0x1407c95c0: ParseFloat(s,64) 后做 NaN/上界/负值守卫，越界返回 0。
func desktopWeatherFloat(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if f != f || f > math.MaxFloat64 || f < 0 {
		return 0
	}
	return f
}

// normalizeDesktopWeatherProviderID 规整天气 provider ID：TrimSpace → ToLower 后按别名归一。
// [S 汇编 0x1407c7c00, 256B]："qweather"→"qweather"；"openmeteo"/"open-meteo"→"openmeteo"；
// "weather-api"→"weather-api"；其他→""。
func normalizeDesktopWeatherProviderID(id string) string {
	switch strings.ToLower(strings.TrimSpace(id)) {
	case "qweather":
		return "qweather"
	case "openmeteo", "open-meteo":
		return "openmeteo"
	case "weather-api":
		return "weather-api"
	default:
		return ""
	}
}

// earliestDesktopWeatherTime 解析两个时间串并返回较早者；任一侧解析失败返回另一侧。
// [S 汇编 0x1407c6f40, 256B]：time.Parse(format(35 字符), a/b)；err1!=nil→b；
// err2!=nil→a；ta.Before(tb)→a 否则 b。
func earliestDesktopWeatherTime(a, b string) string {
	ta, err1 := time.Parse(time.RFC3339, a)
	tb, err2 := time.Parse(time.RFC3339, b)
	if err1 != nil {
		return b
	}
	if err2 != nil {
		return a
	}
	if ta.Before(tb) {
		return a
	}
	return b
}

// normalizeDesktopWeatherLanguage 归一化天气语言代码（提取 ISO 639-1 前 2 字符）。
// [S 汇编 0x1407c7d00, 288B]：TrimSpace → ToLower → IndexAny(s,"-_")；分隔符不在位置 2
// 或前 2 字符非 a-z → 返回默认 "en"；否则返回 s[:2]。
func normalizeDesktopWeatherLanguage(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	i := strings.IndexAny(s, "-_")
	if i < 0 {
		i = len(s)
	}
	if i != 2 {
		return "en"
	}
	for j := 0; j < 2; j++ {
		if c := s[j]; c < 'a' || c > 'z' {
			return "en"
		}
	}
	return s[:2]
}
