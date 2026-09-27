// AUTO-RECONSTRUCTED — twoFactor helpers & icon attachment
// 研究用途
package main

// ---- BootstrapService 辅助方法 ----

// attachTwoFactorEntryStateIconURL 为单个二因素条目装配图标 URL。
// [S asm 0x1408a3780, 77L]：读 entry.IconData/IconRef →
// buildLauncherConfigIconResource("icon/two-factor", IconRef, IconData) →
// 结果写回 IconRef/IconURL，清 IconData。
func (bs *BootstrapService) attachTwoFactorEntryStateIconURL(entry *TwoFactorEntryState) {
	res := bs.buildLauncherConfigIconResource("icon/two-factor", entry.IconRef, entry.IconData)
	entry.IconRef = res.IconRef
	entry.IconURL = res.IconURL
	entry.IconData = ""
}

// attachTwoFactorConfigIconURLs 为配置装配图标 URL。
// [S asm 0x1408a3920, 136L]：遍历 TwoFactorConfig.Entries（stride 0xd0），
// 每项调 buildLauncherConfigIconResource("icon/two-factor", IconRef, IconData) →
// 结果写回 IconRef/IconURL，清 IconData（gcWriteBarrier 三字段同写）。
func (bs *BootstrapService) attachTwoFactorConfigIconURLs(cfg *TwoFactorConfig) {
	if bs == nil || cfg == nil {
		return
	}
	for i := range cfg.Entries {
		e := &cfg.Entries[i]
		// asm 寄存器：IconData @+0x60, IconRef @+0x70
		res := bs.buildLauncherConfigIconResource("icon/two-factor", e.IconRef, e.IconData)
		// asm 写回：IconRef @+0x70 (ptr+len), IconURL @+0x80 (ptr+len)
		e.IconRef = res.IconRef
		e.IconURL = res.IconURL
		// asm：xmm15 → IconData @+0x60 (zeroed)
		e.IconData = ""
	}
}

// attachTwoFactorCommandResultIconURLs 为命令结果装配图标 URL。
// [S 汇编 0x1408a3b80] 180L asm
func (bs *BootstrapService) attachTwoFactorCommandResultIconURLs(result *TwoFactorCommandResult) {
	bs.attachTwoFactorConfigIconURLs(&result.Config)
	for i := range result.State.Entries {
		bs.attachTwoFactorEntryStateIconURL(&result.State.Entries[i])
	}
}

// attachTwoFactorImportResultIconURLs 为导入结果装配图标 URL。
// [S 汇编 0x1408a3f20] 163L asm
func (bs *BootstrapService) attachTwoFactorImportResultIconURLs(result *TwoFactorImportResult) {
	bs.attachTwoFactorConfigIconURLs(&result.Config)
	for i := range result.State.Entries {
		bs.attachTwoFactorEntryStateIconURL(&result.State.Entries[i])
	}
}

// ---- 包级辅助 ----

// parseTwoFactorProvisioningDraft 解析二因素配置文本为草稿。
// [S 汇编 0x1409d6dc0]
func parseTwoFactorProvisioningDraft(text string) (TwoFactorEntryDraft, error) {
	return TwoFactorEntryDraft{}, nil
}
