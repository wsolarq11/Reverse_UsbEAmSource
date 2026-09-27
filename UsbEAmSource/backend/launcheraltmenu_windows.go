package main

// suppressLauncherWindowsAltMenuMessage 抑制启动器窗口的 Alt 系统菜单消息。
// [S] ASM 0x14086b0c0（叶子，无 morestack）：cmp ebx,0x112（WM_SYSCOMMAND）+
// and ecx,0xfff0 + cmp rcx,0xf100（SC_KEYMENU）双匹配才返回 (0,true)，否则 (0,false)。
// 参数 ebx=msg、ecx=wParam；rax（首参，推断为 hwnd）未参与判定，返回 (uintptr,bool)。
func suppressLauncherWindowsAltMenuMessage(hwnd uintptr, msg uint32, wParam uintptr) (uintptr, bool) {
	if msg == 0x112 && wParam&0xfff0 == 0xf100 {
		return 0, true
	}
	return 0, false
}
