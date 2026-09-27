package main

// launcherSecondInstanceShouldReveal 判定第二实例是否应唤起主窗口。
// [S 汇编 0x1408a55c0]：len(args) <= 1（cmp rbx,1 / jg）→ 返回 true（mov eax,1）；
// 否则切片头前进 0x10（跳过首项）+ len/cap 各减 1，尾调 launcherStartedForStartupTray(args[1:])
// 后 xor eax,1 取反。参数 rax/rbx/rcx = []string 三字，返回 bool。
func launcherSecondInstanceShouldReveal(args []string) bool {
	if len(args) <= 1 {
		return true
	}
	return !launcherStartedForStartupTray(args[1:])
}
