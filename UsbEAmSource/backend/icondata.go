// AUTO-RECONSTRUCTED SERVICE METHODS — DOMAIN: icon data 遍历 (config 指纹链 + 槽位)
// 研究用途
//
// 契约来源：
//   - visitLauncherConfigIconSlots(0x14088d940, 3456B) 完整汇编实证 (643L asm)
//   - visitLauncherConfigIconData(0x14088d860, 96B) 汇编
//   - 回调闭包 launcherConfigIconDataFingerprint.func1(0x14088d460)
//   - 格式串常量 12 处经 PE .rodata 字节解码确证
//
// 回调签名 func(launcherConfigIconSlot) bool（4 字段展开为寄存器参数：
// rax=Path.ptr, rbx=Path.len, rcx=Data.ptr, rdi=Ref.ptr, rsi=URL.ptr；
// rdx=closure env 指针）。返回 true 中止遍历。[S]
//
// 档位：全部 [S]（完整汇编实证 + 字段偏移 + 格式串常量解码三重确证）。
package main

import "fmt"

// visitLauncherConfigIconData 遍历配置图标数据。[S 汇编实证 0x14088d860]。
//
// 闭包实证 launcherConfigIconDataFingerprint.func1(0x14088d460)：
//
//	对每个 slot 先写 Path 字符串（槽路径标识），再写 Data 字符串（非空时）。
//	写入序固定：Path → 0x00 → Data（非 nil 则经 TrimSpace）。Ref/URL 不入 hash。
//
// [S]
func visitLauncherConfigIconData(cfg *LauncherConfig, visit func(string)) {
	if cfg == nil || visit == nil {
		return
	}
	visitLauncherConfigIconSlots(cfg, func(slot launcherConfigIconSlot) bool {
		visit(slot.Path)
		if slot.Data != nil && *slot.Data != "" {
			visit(*slot.Data)
		}
		return false
	})
}

// visitLauncherConfigIconSlots 遍历全部配置图标槽位，对每个槽位调用 visit。
// [S 汇编实证 0x14088d940, 643L/3456B]
//
// 访问顺序及实证字段偏移（struct base calc → relative + off → &field）：
//
//  1. Apps[i]        stride 0x168 | slot1: IconData(+0x48) IconRef(+0x58) IconURL(+0x68)
//     | slot2: CustomIconData(+0x78) CustomIconRef(+0x88) IconURL(+0x68)
//  2. SpeedDial[i]   stride 0xe8  | IconData(+0x50) IconRef(+0x60) IconURL(+0x70)
//  3. Bookmarks.Custom[i] stride 0xe8 | 同 LinkEntry 偏移
//  4. Preferences.TagCatalog[i] stride 0x50 | IconData(+0x20) IconRef(+0x30) IconURL(+0x40)
//  5. Preferences.ConsoleItems[i] stride 0xe0 | IconData(+0x90) IconRef(+0xa0) IconURL(+0xb0)
//  6. MouseGestures.Apps[i] stride 0x118
//     profile: IconData(+0xd0) IconRef(+0xe0) IconURL(+0xf0)
//     match sub-loop (matches[j] count @ +0xa0, entry stride 0x60):
//     IconData(+0x68) IconRef(+0x78) IconURL(+0x88) of GestureAppMatch
//  7. TwoFactor.Entries[i] stride 0xd0 | IconData(+0x60) IconRef(+0x70) IconURL(+0x80)
//  8. OLEDBlackout.MediaPauseExclusions[i] stride 0x68 | IconData(+0x38) IconRef(+0x48) IconURL(+0x58)
//  9. WindowManagement.Target 单条目 | IconData(cfg+0x7a0) IconRef(cfg+0x7b0) IconURL(cfg+0x7c0)
//
// 格式串常量 12 处全部经 PE .rodata 字节解码确证（见 tools/go-introspect/read_gostring.py）。
// [S]
func visitLauncherConfigIconSlots(cfg *LauncherConfig, visit func(launcherConfigIconSlot) bool) {
	if cfg == nil || visit == nil {
		return
	}

	// ---- 1. Apps 循环 ----
	for i := range cfg.Apps {
		a := &cfg.Apps[i]
		slot := launcherConfigIconSlot{
			Path: fmt.Sprintf("apps[%d].iconData", i),
			Data: &a.IconData,
			Ref:  &a.IconRef,
			URL:  &a.IconURL,
		}
		if visit(slot) {
			return
		}
		slot = launcherConfigIconSlot{
			Path: fmt.Sprintf("apps[%d].customIconData", i),
			Data: &a.CustomIconData,
			Ref:  &a.CustomIconRef,
			URL:  &a.IconURL, // 两槽共享 IconURL 字段
		}
		if visit(slot) {
			return
		}
	}

	// ---- 2. SpeedDial 循环 ----
	for i := range cfg.SpeedDial {
		l := &cfg.SpeedDial[i]
		slot := launcherConfigIconSlot{
			Path: fmt.Sprintf("speedDial[%d].iconData", i),
			Data: &l.IconData,
			Ref:  &l.IconRef,
			URL:  &l.IconURL,
		}
		if visit(slot) {
			return
		}
	}

	// ---- 3. Bookmarks.Custom 循环 ----
	for i := range cfg.Bookmarks.Custom {
		l := &cfg.Bookmarks.Custom[i]
		slot := launcherConfigIconSlot{
			Path: fmt.Sprintf("bookmarks.custom[%d].iconData", i),
			Data: &l.IconData,
			Ref:  &l.IconRef,
			URL:  &l.IconURL,
		}
		if visit(slot) {
			return
		}
	}

	// ---- 4. Preferences.TagCatalog 循环 ----
	for i := range cfg.Preferences.TagCatalog {
		t := &cfg.Preferences.TagCatalog[i]
		slot := launcherConfigIconSlot{
			Path: fmt.Sprintf("preferences.tagCatalog[%d].iconData", i),
			Data: &t.IconData,
			Ref:  &t.IconRef,
			URL:  &t.IconURL,
		}
		if visit(slot) {
			return
		}
	}

	// ---- 5. Preferences.ConsoleItems 循环 ----
	for i := range cfg.Preferences.ConsoleItems {
		c := &cfg.Preferences.ConsoleItems[i]
		slot := launcherConfigIconSlot{
			Path: fmt.Sprintf("preferences.consoleItems[%d].iconData", i),
			Data: &c.IconData,
			Ref:  &c.IconRef,
			URL:  &c.IconURL,
		}
		if visit(slot) {
			return
		}
	}

	// ---- 6. MouseGestures.Apps 循环（含子循环 Matches） ----
	for i := range cfg.MouseGestures.Apps {
		p := &cfg.MouseGestures.Apps[i]
		slot := launcherConfigIconSlot{
			Path: fmt.Sprintf("mouseGestures.apps[%d].iconData", i),
			Data: &p.IconData,
			Ref:  &p.IconRef,
			URL:  &p.IconURL,
		}
		if visit(slot) {
			return
		}
		// 子循环：GesturesAppProfile.Matches（匹配条目特有的图标槽）
		for j := range p.Matches {
			m := &p.Matches[j]
			slot := launcherConfigIconSlot{
				Path: fmt.Sprintf("mouseGestures.apps[%d].matches[%d].iconData", i, j),
				Data: &m.IconData,
				Ref:  &m.IconRef,
				URL:  &m.IconURL,
			}
			if visit(slot) {
				return
			}
		}
	}

	// ---- 7. TwoFactor.Entries 循环 ----
	for i := range cfg.TwoFactor.Entries {
		e := &cfg.TwoFactor.Entries[i]
		slot := launcherConfigIconSlot{
			Path: fmt.Sprintf("twoFactor.entries[%d].iconData", i),
			Data: &e.IconData,
			Ref:  &e.IconRef,
			URL:  &e.IconURL,
		}
		if visit(slot) {
			return
		}
	}

	// ---- 8. OLEDBlackout.MediaPauseExclusions 循环 ----
	for i := range cfg.OLEDBlackout.MediaPauseExclusions {
		x := &cfg.OLEDBlackout.MediaPauseExclusions[i]
		slot := launcherConfigIconSlot{
			Path: fmt.Sprintf("oledBlackout.mediaPauseExclusions[%d].iconData", i),
			Data: &x.IconData,
			Ref:  &x.IconRef,
			URL:  &x.IconURL,
		}
		if visit(slot) {
			return
		}
	}

	// ---- 9. WindowManagement.Target 单条目 ----
	slot := launcherConfigIconSlot{
		Path: "windowManagement.target.iconData",
		Data: &cfg.WindowManagement.Target.IconData,
		Ref:  &cfg.WindowManagement.Target.IconRef,
		URL:  &cfg.WindowManagement.Target.IconURL,
	}
	visit(slot) // 单条目，asm 会调用但不检查返回值
}
