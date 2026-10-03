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

// shouldUseLegacyExplicitLinkBrowser 判定是否使用 legacy 显式链接浏览器。
// [S-sig 0x140769b20, 320B]：url TrimSpace 空→false；cleanStringList(list) 非空→true；
// IndexAny(url,":/")>=0→true；否则末尾找扩展名截断→TrimSpace 非空→true。
// 体待链接浏览器域专项还原。
func shouldUseLegacyExplicitLinkBrowser(a, b, c interface{}) bool {
	_, _, _ = a, b, c
	return false
}

// validateExplicitLinkBrowserResolvedTarget 校验显式链接浏览器解析目标。
// [S-sig 0x1408aab60, 320B]：TrimSpace 空→error；isWindowsAbsoluteFilesystemPath →
// classifyAutomaticWindowsPath + isDisallowedExplicitLinkBrowserExecutable。体待路径分类域专项还原。
func validateExplicitLinkBrowserResolvedTarget(a string) error {
	_ = a
	return nil
}
