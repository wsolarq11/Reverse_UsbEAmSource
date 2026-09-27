package main

// (*launcherHotkeyRegistrationError).Error 返回底层错误的文本。
// [S 汇编 0x14089d6e0]：e==nil 或 e.Err==nil → ""；否则经 itab[0x18] 调 e.Err.Error()。
func (e *launcherHotkeyRegistrationError) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

// (*launcherHotkeyRegistrationError).Unwrap 返回底层错误。
// [S 汇编 0x14089d740]：e==nil → nil；否则返回 e.Err（interface 两字直接透传）。
func (e *launcherHotkeyRegistrationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}
