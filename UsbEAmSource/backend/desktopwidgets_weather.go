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
// [S] ASM 0x1407c95c0: ParseFloat(s,64) 后做 NaN/±MaxFloat64 越界守卫，越界返回 0。
// 边界常量实证：0x1411cd838=+MaxFloat64、0x1411cd858=-MaxFloat64（.rdata 双精度）。
func desktopWeatherFloat(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if f != f || f > math.MaxFloat64 || f < -math.MaxFloat64 {
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

// isValidQWeatherAPIHost 校验 qweather API host 是否合法（域后缀 + label 规则）。
// [S-sig 0x1407c9940, 384B]：len>253→false；后缀非 .qweather.com/.net→false；
// 按 "." split → label 空/超 63/首尾 "-"/非 [a-z0-9-]→false。体待后缀常量专项还原。
func isValidQWeatherAPIHost(a string) bool {
	_ = a
	return false
}

// providerErrorSnapshot 返回提供器错误快照（锁内遍历 providerErrors map 复制）。
// [S 汇编 0x1407c72c0, 448B]：makemap_small 新建 result；s==nil→返回空 map；
// mu.Lock → defer mu.Unlock → 遍历 s.providerErrors(+0x40) mapassign 复制 → 返回 result。
func (s *desktopWidgetWeatherService) providerErrorSnapshot() map[string]string {
	result := make(map[string]string)
	if s == nil {
		return result
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.providerErrors {
		result[k] = v
	}
	return result
}

// setProviderError 设置/清除某天气提供器的错误信息（锁内写 providerErrors map）。
// [S 汇编 0x1407c7040, 544B(0x220)]：s==nil→return；normalizeDesktopWeatherProviderID(providerID)
// 空→return；mu.Lock→defer mu.Unlock；providerErrors nil→make；TrimSpace(errMsg) 非空→写入、
// 空→mapdelete。字段 +0x30=mu、+0x40=providerErrors。
func (s *desktopWidgetWeatherService) setProviderError(providerID, errMsg string) {
	if s == nil {
		return
	}
	id := normalizeDesktopWeatherProviderID(providerID)
	if id == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.providerErrors == nil {
		s.providerErrors = make(map[string]string)
	}
	if msg := strings.TrimSpace(errMsg); msg != "" {
		s.providerErrors[id] = msg
	} else {
		delete(s.providerErrors, id)
	}
}

// desktopWeatherCredential 解析桌面天气凭证（内置 OpenMeteo/自定义 provider + 解密）。
// [S-sig 0x1407c8860, 448B]：provider=="OpenMeteo"→nil；map 命中→TrimSpace 空→nil；
// unprotectDesktopWidgetSecret→err→error。体待凭证域专项还原。
func desktopWeatherCredential(provider string) interface{} {
	_ = provider
	return nil
}

// desktopWeatherTimes 派生三个 RFC3339Nano 时间串：now、now+30min、now+24h。
// [S 汇编 0x1407c93a0, 0x1f7B]：三次 time.Time.Format（35 字符布局 0x140c7708d =
// time.RFC3339Nano "2006-01-02T15:04:05.999999999Z07:00"）；Add 时长常量
// 0x1a3185c5000=1800s(30min)、0x4e94914f0000=86400s(24h)。返回 (now, now+30min, now+24h)。
func desktopWeatherTimes(t time.Time) (string, string, string) {
	a := t.Format(time.RFC3339Nano)
	b := t.Add(30 * time.Minute).Format(time.RFC3339Nano)
	c := t.Add(24 * time.Hour).Format(time.RFC3339Nano)
	return a, b, c
}
