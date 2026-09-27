package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// desktopTimezoneOffset 把 UTC 偏移秒数格式化为 "UTC±HH:MM"。
// [S 汇编 0x1407ab120]：offset<0 取 "-" 否则 "+"；abs=|offset|；hours=abs/3600（0x48d159e26af37c05 magic 除法）、
// minutes=(abs-hours*3600)/60；fmt.Sprintf("UTC%s%02d:%02d"@0x140c50712, sign, hours, minutes)。
func desktopTimezoneOffset(offset int) string {
	sign := "+"
	abs := offset
	if offset < 0 {
		sign = "-"
		abs = -offset
	}
	hours := abs / 3600
	minutes := (abs - hours*3600) / 60
	return fmt.Sprintf("UTC%s%02d:%02d", sign, hours, minutes)
}

// buildDesktopWorldClockSnapshot 构建世界时钟快照，timezones 为时区列表，now 为当前时间（零值回退 time.Now）。
// [S 汇编 0x1407aa960, 1984B] len(timezones)>24 → errors.New("desktopWidgets.tooManyTimezones")
// （31B @0x140c70d7c）；serverNow=now.Format(RFC3339Nano)（35B）；make([]DesktopWorldClockTime,0,len)；
// make(map[string]struct{},len) 去重；逐时区：id=TrimSpace(zone.TimezoneID)，空或已见 → 跳过；
// time.LoadLocation(id) 失败 → fmt.Errorf("desktopWidgets.invalidTimezone: %s"@0x140c75595, id)；
// localNow=now.In(loc)；_,offset=localNow.Zone()；LocalTime=localNow.Format(RFC3339
// "2006-01-02T15:04:05Z07:00" 25B)；Offset=desktopTimezoneOffset(offset)；OffsetSeconds=offset；
// DST=localNow.IsDST()；Label=TrimSpace(zone.Label)；返回 (DesktopWorldClockSnapshot,error)。
func buildDesktopWorldClockSnapshot(timezones []DesktopWorldClockZone, now time.Time) (DesktopWorldClockSnapshot, error) {
	if len(timezones) > 24 {
		return DesktopWorldClockSnapshot{}, errors.New("desktopWidgets.tooManyTimezones")
	}
	if now.IsZero() {
		now = time.Now()
	}
	serverNow := now.Format(time.RFC3339Nano)
	clocks := make([]DesktopWorldClockTime, 0, len(timezones))
	seen := make(map[string]struct{}, len(timezones))
	for _, zone := range timezones {
		id := strings.TrimSpace(zone.TimezoneID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		loc, err := time.LoadLocation(id)
		if err != nil {
			return DesktopWorldClockSnapshot{}, fmt.Errorf("desktopWidgets.invalidTimezone: %s", id)
		}
		seen[id] = struct{}{}
		localNow := now.In(loc)
		_, offset := localNow.Zone()
		clocks = append(clocks, DesktopWorldClockTime{
			TimezoneID:    id,
			Label:         strings.TrimSpace(zone.Label),
			LocalTime:     localNow.Format(time.RFC3339),
			Offset:        desktopTimezoneOffset(offset),
			OffsetSeconds: offset,
			DST:           localNow.IsDST(),
		})
	}
	return DesktopWorldClockSnapshot{
		ServerNow: serverNow,
		Clocks:    clocks,
	}, nil
}
