package main

import (
	"math"
	"strconv"
	"strings"
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
