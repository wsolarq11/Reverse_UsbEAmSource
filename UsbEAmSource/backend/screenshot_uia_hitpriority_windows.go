// AUTO-RECONSTRUCTED — DOMAIN: screenshot UIA control-hit priority
// Source: UsbEAm_Launcher 1.0.3 (Go 1.25.12, PE64), main.screenshotPreferUIAElementHitInfo
// (0x1409b2ce0) 及其内联 rank / area 判定。
//
// 本文件只承载「哪个命中候选优先采用」的纯几何/集合判定，与 COM 生命周期分离。
//
// 研究用途
package main

import "image"

// UIA 控件类型标识（UIAutomationClient.h Control Type Identifiers）。
// 早期出局窗口以 0xc358(50008) 为基址、宽 0x1a；0xc356(50006) 与 0xc364(50020)
// 是窗口边界上的具名项，在 rank 判定中提前出局。
const (
	screenshotUIAImageControlTypeID uint32 = 50006 // 0xc356 UIA_ImageControlTypeId
	screenshotUIAListControlTypeID  uint32 = 50008 // 0xc358 UIA_ListControlTypeId（跳转表基址）
	screenshotUIATextControlTypeID  uint32 = 50020 // 0xc364 UIA_TextControlTypeId
)

// screenshotUIAPreferRankOne 是 rank=1 的控件类型集合。
// 数据来源：原 exe 跳转表 0x1411e1f80（T1）逐表项抽取，索引域 {0,1,2,9,10,13,15,18,20,22,24,25}
// 加基址 0xc358 换算；T2(0x1411e2060)/T3(0x1411e2140)/T4(0x1411e2220) 三表索引集合与之完全相同。
// 表项值域仅 {1,2}，索引 26/27 表项为 0 且被 `cmp rcx,0x19` 挡在表外，不可达。
// 本表由 tools/parity/emu/ref_puiahi.py 从抽取产物机械推导，二者必须一致。
var screenshotUIAPreferRankOne = [...]uint32{
	50008, // UIA_ListControlTypeId
	50009, // UIA_MenuControlTypeId
	50010, // UIA_MenuBarControlTypeId
	50017, // UIA_StatusBarControlTypeId
	50018, // UIA_TabControlTypeId
	50021, // UIA_ToolBarControlTypeId
	50023, // UIA_TreeControlTypeId
	50026, // UIA_GroupControlTypeId
	50028, // UIA_DataGridControlTypeId
	50030, // UIA_DocumentControlTypeId
	50032, // UIA_WindowControlTypeId
	50033, // UIA_PaneControlTypeId
}

// screenshotUIAElementRank 计算 UIA 命中候选的采用优先级。数值越大越优先采用。
// [S 汇编 0x1409b2ce0；跳转表 T1 0x1411e1f80 逐表项实证]
//
//	ok==false                                    → -1（0x1409b2db6 / 0x1409b2ec1 前置 eax=-1）
//	ok==true 且 rect 退化（Dx<2 或 Dy<2）         → -1（0x1409b2dc0 → 0x1409b3053）
//	controlType 0xc356(50006) 或 0xc364(50020)     → -1（同路径，命中早期即出局）
//	窗口内其余值                                   → 2（0x1409b2eb8）
//	screenshotUIAPreferRankOne 集合内            → 1（0x1409b2eb1）
//
// 跳转表索引域与 rank=1 集合的对应关系由 T1 表项抽取确证，非推断。
func screenshotUIAElementRank(rect image.Rectangle, controlType uint32, ok bool) int {
	if !ok {
		return -1
	}
	if controlType == screenshotUIAImageControlTypeID {
		return -1
	}
	if controlType == screenshotUIATextControlTypeID {
		return -1
	}
	if rect.Dx() < 2 || rect.Dy() < 2 {
		return -1
	}
	for _, t := range screenshotUIAPreferRankOne {
		if controlType == t {
			return 1
		}
	}
	return 2
}

// screenshotUIAElementArea 计算 UIA 元素矩形面积，非正面积归零。
// [S 汇编 0x1409b2f69-0x1409b2fc9]：Max.X<=Min.X → 0；Max.Y<=Min.Y → 0；
// 否则 (Max.X-Min.X)*(Max.Y-Min.Y)。
func screenshotUIAElementArea(rect image.Rectangle) int64 {
	if rect.Max.X <= rect.Min.X || rect.Max.Y <= rect.Min.Y {
		return 0
	}
	return int64(rect.Max.X-rect.Min.X) * int64(rect.Max.Y-rect.Min.Y)
}

// screenshotPreferUIAElementHitInfo 判定 a 是否应优先于 b 采用。
//
// [S 汇编 0x1409b2ce0, 928B]。寄存器实证：a 走 rax/rbx/rcx/rdi + aControlType=esi + aOK=r8b；
// b 走 [rsp+0x60] 起 0x20 字节（rect+controlType）+ [rsp+0x80] controlType + ok 在其 +0xc。
// 函数体对同一对元素重复求 rank 四次（0x1409b2d35/0x1409b2dc0 对 b，0x1409b2e36 对 a 赋值后
// 于 0x1409b2e5c 读取）——内联无副作用可观察行为一致，此处求值一次缓存。
//
//	rankA != rankB            → rankA > rankB（0x1409b2f63 jne → 0x1409b304a setg）
//	areaB <= 0                → false（0x1409b2fcd test rdi,rdi → jle 0x1409b3042）
//	areaA <= 0                → true （0x1409b2fd2 test rsi,rsi → jle 0x1409b3037）
//	areaB != areaA            → areaB < areaA（0x1409b2fd7 → 0x1409b302e setl）
//	同 rank 同面积             → false（见下）
//
// 内层分类（0x1409b2fdc 基址 0x82f149→T3 / 0x1409b300e 基址 0x82f20b→T4）经表项抽取确证：
// T3 索引集合 {0,1,2,9,10,13,15,18,20,22,24,25} 与 T4 完全相同，亦与 T1 完全相同；
// 两表同索引集合意味着 rank(aCT) 与 rank(bCT) 在窗口内恒同值，内层判定退化为恒等，
// 0x1409b3021 的 `xor eax,1` 恒得 false。该分支不可达，落 false。
//
// 行为对拍：tools/parity/cases/screenshot_uia/（1557 例，见 backend/screenshot_uia_parity_test.go）。
func screenshotPreferUIAElementHitInfo(a image.Rectangle, aControlType uint32, aOK bool, b image.Rectangle, bControlType uint32, bOK bool) bool {
	rankA := screenshotUIAElementRank(a, aControlType, aOK)
	rankB := screenshotUIAElementRank(b, bControlType, bOK)
	if rankA != rankB {
		return rankA > rankB
	}
	areaA := screenshotUIAElementArea(a)
	areaB := screenshotUIAElementArea(b)
	if areaB <= 0 {
		return false
	}
	if areaA <= 0 {
		return true
	}
	if areaB != areaA {
		return areaB < areaA
	}
	return false
}
