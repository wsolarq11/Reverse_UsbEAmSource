package main

import (
	"container/list"
	"errors"
	"fmt"
	"time"

	"github.com/6tail/lunar-go/calendar"
)

// desktopCalendarStringList 把 *list.List 中所有非空 string 元素收集成 []string。
// [S 汇编 0x1407aa820, 320B] l==nil 或 l.len==0 → nil；make([]string,0,l.len)；
// 遍历 l.Front()..Next()：e.Value.(string) 断言成功且非空 → append；返回切片。
func desktopCalendarStringList(l *list.List) []string {
	if l == nil || l.Len() == 0 {
		return nil
	}
	out := make([]string, 0, l.Len())
	for e := l.Front(); e != nil; e = e.Next() {
		if s, ok := e.Value.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}

// buildDesktopCalendarMonth 构建某月（year/month）的桌面日历，now 为当前时间（零值回退 time.Now）。
// [S 汇编 0x1407a9f60, 2240B] year∈[1901,2100]、month∈[1,12]，否则
// errors.New("desktopWidgets.calendarRange")（28B @0x140c6b990）；dayCount =
// time.Date(year,month+1,0,12,0,0,0,UTC).Day()；serverNow = now.Format(RFC3339Nano)
// （"2006-01-02T15:04:05.999999999Z07:00" 35B）；make([]DesktopCalendarDay,0,dayCount)
// （元素 144B=0x90，makeslice et@0x140bebe00）；逐日 NewSolar(year,month,day,0,0,0)
// → NewLunarFromSolar；Date=fmt.Sprintf("%04d-%02d-%02d")、Day=day、
// Weekday=time.Date(...,12,...).Weekday()（absSec/86400+3 对 7 取余）、
// LunarYear=GetYearInGanZhi()（GAN[yearGanIndex+1]+ZHI[yearZhiIndex+1]，字段 +0x30/+0x38）、
// LunarMonth=GetMonthInChinese()（month<0→"闰"+MONTH[-month]，字段 +0x08）、
// LunarDay=GetDayInChinese()（DAY[day]，字段 +0x10）、SolarTerm=GetJieQi()、
// SolarFestivals=desktopCalendarStringList(GetFestivals())、LunarFestivals 同理；
// 返回 (DesktopCalendarMonth,error)。
func buildDesktopCalendarMonth(year int, month int, now time.Time) (DesktopCalendarMonth, error) {
	if year < 1901 || year > 2100 || month < 1 || month > 12 {
		return DesktopCalendarMonth{}, errors.New("desktopWidgets.calendarRange")
	}
	if now.IsZero() {
		now = time.Now()
	}
	dayCount := time.Date(year, time.Month(month+1), 0, 12, 0, 0, 0, time.UTC).Day()
	serverNow := now.Format(time.RFC3339Nano)
	days := make([]DesktopCalendarDay, 0, dayCount)
	for day := 1; day <= dayCount; day++ {
		solar := calendar.NewSolar(year, month, day, 0, 0, 0)
		lunar := calendar.NewLunarFromSolar(solar)
		days = append(days, DesktopCalendarDay{
			Date:           fmt.Sprintf("%04d-%02d-%02d", year, month, day),
			Day:            day,
			Weekday:        int(time.Date(year, time.Month(month), day, 12, 0, 0, 0, time.UTC).Weekday()),
			LunarYear:      lunar.GetYearInGanZhi(),
			LunarMonth:     lunar.GetMonthInChinese(),
			LunarDay:       lunar.GetDayInChinese(),
			SolarTerm:      lunar.GetJieQi(),
			SolarFestivals: desktopCalendarStringList(solar.GetFestivals()),
			LunarFestivals: desktopCalendarStringList(lunar.GetFestivals()),
		})
	}
	return DesktopCalendarMonth{
		Year:      year,
		Month:     month,
		ServerNow: serverNow,
		Days:      days,
	}, nil
}
