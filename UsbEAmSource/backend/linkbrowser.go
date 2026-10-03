package main

// ScanLinkBrowsers 扫描系统已安装的浏览器。
// [S 汇编 0x1408d0b40]：直接转发 detectLinkBrowsers()，返回 []StartMenuApp（三字 slice）；
// 尾段 xor edi/esi 为返回值寄存器清零（仅回传 rax/rbx/rcx）。
func (bs *BootstrapService) ScanLinkBrowsers() []StartMenuApp {
	return detectLinkBrowsers()
}

// startLinkWithExplicitBrowser 用显式浏览器启动链接（buildLinkCommand → startExplicitLinkBrowser）。
// [S-sig 0x140769a60, 192B]：buildLinkCommand → len<1 则 panicSliceB；命令[1:] → startExplicitLinkBrowser。
func startLinkWithExplicitBrowser(a, b, c interface{}) {
	_, _, _ = a, b, c
}
