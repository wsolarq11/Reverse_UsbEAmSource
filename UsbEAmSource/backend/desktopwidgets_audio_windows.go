package main

import (
	"errors"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// playSoundProc winmm.dll 的 PlaySoundW 延迟加载入口（LazyProc 名 "PlaySoundW"@0x140c43f7d,10B）。
var playSoundProc = windows.NewLazySystemDLL("winmm.dll").NewProc("PlaySoundW")

// playDesktopReminderSound 通过 Win32 PlaySoundW 播放提醒音。
// [S 汇编 0x1407a7b20]：UTF16PtrFromString(name) 失败→err；否则 playSoundProc.Call(ptr,0,flags)；
// r1!=0→nil；err==nil 或 errors.Is(err, ERROR_SUCCESS@data=0x0) → errors.New
// "Windows 无法播放提醒音频"@0x140c72466,32B；否则返回 err。
func playDesktopReminderSound(name string, flags uintptr) error {
	ptr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return err
	}
	r1, _, err := playSoundProc.Call(uintptr(unsafe.Pointer(ptr)), 0, flags)
	if r1 != 0 {
		return nil
	}
	if err != nil && !errors.Is(err, windows.ERROR_SUCCESS) {
		return err
	}
	return errors.New("Windows 无法播放提醒音频")
}

// messageBeepProc user32.dll 的 MessageBeep 延迟加载入口（LazyProc 名 "MessageBeep"@0x140c47376,11B）。
var messageBeepProc = windows.NewLazySystemDLL("user32.dll").NewProc("MessageBeep")

// defaultReminderSystemSounds 系统提醒音播放顺序（全局 slice@0x141bc67a0,len=3）。
var defaultReminderSystemSounds = []string{
	"Notification.Default", // @0x140c5d877,20B
	"SystemNotification",   // @0x140c59aa4,18B
	"SystemAsterisk",       // @0x140c505b4,14B
}

// playDesktopReminderSystemSound 依次尝试系统提醒音，全部失败回退 MessageBeep(1)。
// [S 汇编 0x1407a79e0]：逐条 playDesktopReminderSound(s, 0x210002) 首个 nil 即返回；
// 全失败→messageBeepProc.Call(1)；r1!=0→nil；err==nil 或 errors.Is(err, ERROR_SUCCESS) →
// errors.New "Windows 系统提醒音频不可用"@0x140c770f6,35B；否则返回 err。
func playDesktopReminderSystemSound() error {
	for _, s := range defaultReminderSystemSounds {
		if err := playDesktopReminderSound(s, 0x210002); err == nil {
			return nil
		}
	}
	_, _, err := messageBeepProc.Call(1)
	if err != nil && !errors.Is(err, windows.ERROR_SUCCESS) {
		return err
	}
	return errors.New("Windows 系统提醒音频不可用")
}

// playDesktopReminderAudioSequence 按次数循环播放提醒音（custom 播放文件、其他播放系统音）。
// [S 汇编 0x1407a78e0]：循环 count 次；name=="custom"（len==6 + "cust"@0x74737563 + "om"@0x6d6f）→
// playDesktopReminderSound(path, 0x220002)；否则 playDesktopReminderSystemSound()；err 非 nil 即返回；
// 相邻两次间隔 time.Sleep(0xee6b280=250ms)。
func playDesktopReminderAudioSequence(name, path string, count int) error {
	for i := 0; i < count; i++ {
		var err error
		if name == "custom" {
			err = playDesktopReminderSound(path, 0x220002)
		} else {
			err = playDesktopReminderSystemSound()
		}
		if err != nil {
			return err
		}
		if i+1 < count {
			time.Sleep(250 * time.Millisecond)
		}
	}
	return nil
}
