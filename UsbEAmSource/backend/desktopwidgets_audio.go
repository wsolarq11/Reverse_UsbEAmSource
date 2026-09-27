package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// normalizeDesktopReminderAudio 归一化提醒音配置。
// [S 汇编 0x1407a6660]：name=ToLower(TrimSpace(name)) 空→"system"@0x140c375e4,6B；
// path=TrimSpace(path)；vol==0→1；name=="system"→path=""；返回 (name,path,vol)。
func normalizeDesktopReminderAudio(name, path string, vol int) (string, string, int) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		name = "system"
	}
	path = strings.TrimSpace(path)
	if vol == 0 {
		vol = 1
	}
	if name == "system" {
		path = ""
	}
	return name, path, vol
}

// validateDesktopReminderAudio 校验提醒音配置，非法返回 "desktopWidgets.invalidReminderAudio" 错误。
// [S 汇编 0x1407a68c0]：name=="system"→vol∉[1,10] 非法否则 nil；name=="custom"→
// vol∉[1,10]、path 空/len>4096、!IsAbs(path)、扩展名 != ".wav"（EqualFold@0x140c347ca）任一非法；
// 其他 name 非法。
func validateDesktopReminderAudio(name, path string, vol int) error {
	switch name {
	case "system":
		if vol < 1 || vol > 10 {
			return errors.New("desktopWidgets.invalidReminderAudio")
		}
		return nil
	case "custom":
		if vol < 1 || vol > 10 {
			return errors.New("desktopWidgets.invalidReminderAudio")
		}
		if path == "" || len(path) > 4096 || !filepath.IsAbs(path) {
			return errors.New("desktopWidgets.invalidReminderAudio")
		}
		dot := strings.LastIndexByte(path, '.')
		if dot < 0 || !strings.EqualFold(path[dot:], ".wav") {
			return errors.New("desktopWidgets.invalidReminderAudio")
		}
		return nil
	default:
		return errors.New("desktopWidgets.invalidReminderAudio")
	}
}

// DesktopReminderAudioDefinition 提醒音频定义。
// 字段偏移从 asm 0x1407a6780 确证：Name@0x88、Path@0x98、Volume@0xa8；
// 前置 0x88 字节字段布局本轮未触及，用填充占位。
type DesktopReminderAudioDefinition struct {
	_      [0x88]byte
	Name   string // +0x88
	Path   string // +0x98
	Volume int    // +0xa8
}

// desktopReminderAudioFromDefinition 从音频定义构建归一化音频配置。
// [S 汇编 0x1407a6780]：def==nil→normalizeDesktopReminderAudio("", "", 0)；
// 否则 normalizeDesktopReminderAudio(def.Name, def.Path, def.Volume)。
func desktopReminderAudioFromDefinition(def *DesktopReminderAudioDefinition) (string, string, int) {
	if def == nil {
		return normalizeDesktopReminderAudio("", "", 0)
	}
	return normalizeDesktopReminderAudio(def.Name, def.Path, def.Volume)
}

// Play 播放一条提醒音（校验后异步播放，播放器互斥）。
// [S 汇编 0x1407a6ae0, 544B] 实证：
//
//	normalizeDesktopReminderAudio(Mode,Path,RepeatCount) → validate 失败即返回 err；
//	name=="custom"（len==6 + "cust"@0x74737563 + "om"@0x6d6f）时 os.Stat(path) 失败
//	或 info.Mode()&os.ModeType!=0（test eax,0x8f280000）→ "desktopWidgets.reminderAudioUnavailable"（39B）；
//	否则 go func1（p.playback.Lock + defer Unlock + playDesktopReminderAudioSequence，
//	失败 log.Printf "播放首页组件提醒音频失败: %v"），立即返回 nil。
func (p *platformDesktopWidgetAudioPlayer) Play(audio desktopReminderAudio) error {
	name, path, vol := normalizeDesktopReminderAudio(audio.Mode, audio.Path, audio.RepeatCount)
	if err := validateDesktopReminderAudio(name, path, vol); err != nil {
		return err
	}
	if name == "custom" {
		info, err := os.Stat(path)
		if err != nil || info.Mode()&os.ModeType != 0 {
			return errors.New("desktopWidgets.reminderAudioUnavailable")
		}
	}
	go func() {
		p.playback.Lock()
		defer p.playback.Unlock()
		if err := playDesktopReminderAudioSequence(name, path, vol); err != nil {
			log.Printf("播放首页组件提醒音频失败: %v", err)
		}
	}()
	return nil
}

// deliverNotification 下发一条首页组件通知并顺带播放其提醒音。
// [S 汇编 0x1407a6f20, 736B] 实证：
//
//	s==nil → "desktopWidgets.notificationFailed"（33B）；
//	s.notifier!=nil → Notify(title,message,audio!=nil)，否则 notifErr="通知服务不可用"（21B）；
//	audio!=nil 时 s.audio!=nil → Play(*audio)，否则 playErr="音频服务不可用"（21B）；
//	notifErr 非空 → log.Printf("发送首页组件通知失败: %v" /*34B*/) + 返回 notificationFailed（33B）；
//	playErr 非空 → log.Printf("播放首页组件提醒音频失败: %v" /*40B*/) + 返回 reminderAudioFailed（34B）；
//	否则返回 nil。
func (s *desktopWidgetService) deliverNotification(title, message string, audio *desktopReminderAudio) error {
	if s == nil {
		return errors.New("desktopWidgets.notificationFailed")
	}
	var notifErr error
	if s.notifier != nil {
		notifErr = s.notifier.Notify(title, message, audio != nil)
	} else {
		notifErr = errors.New("通知服务不可用")
	}
	var playErr error
	if audio != nil {
		if s.audio != nil {
			playErr = s.audio.Play(*audio)
		} else {
			playErr = errors.New("音频服务不可用")
		}
	}
	if notifErr != nil {
		log.Printf("发送首页组件通知失败: %v", notifErr)
		return errors.New("desktopWidgets.notificationFailed")
	}
	if playErr != nil {
		log.Printf("播放首页组件提醒音频失败: %v", playErr)
		return errors.New("desktopWidgets.reminderAudioFailed")
	}
	return nil
}
