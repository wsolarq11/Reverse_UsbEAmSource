package main

import (
	"fmt"
	"strconv"
	"time"
)

// inputMonitorXButtonLabel 返回鼠标扩展按钮（XButton）的标签。
// [S] ASM 0x1408695e0: HIWORD(xbutton)==1→"x1"，==2→"x2"，否则 "x"+FormatInt(HIWORD,10)。
func inputMonitorXButtonLabel(xbutton uint32) string {
	switch xbutton >> 16 {
	case 1:
		return "x1"
	case 2:
		return "x2"
	}
	return "x" + strconv.FormatInt(int64(xbutton>>16), 10)
}

// timeFromWindowsTick 返回当前时间的 unix 毫秒时间戳（输入事件时间戳源）。
// [S-eq 汇编 0x1408697a0, 128B] 实证：time.Now() → 有 monotonic 时 sec=(wall<<1>>31)
// +wallToInternal(0xdd7b17f80)，无 monotonic 时 sec=ext；nsec=wall&0x3fffffff；
// rax=(sec-unixToInternal)*1000 + nsec/1e6（magic 0x431bde82d7b634db>>82）＝unix 毫秒。
// time.Time.wall 未导出，公开 API UnixMilli() 与上述 sec()/nsec() 内联展开精确等价。
func timeFromWindowsTick() int64 {
	return time.Now().UnixMilli()
}

// inputMonitorKeyLabelMap 虚拟键码到自定义标签的映射（asm 全局 map）。
var inputMonitorKeyLabelMap = map[int]string{}

// inputMonitorKeyLabel 返回虚拟键码的可读标签。
// [S 汇编 0x140869820, 288B]：'A'-'Z'/'0'-'9'→单字符；0x70-0x87（F1-F24）→"F"+FormatInt(vk-0x6f)；
// 否则查全局 map，命中返回标签；未命中 fmt.Sprintf("VK%d", vk)。
func inputMonitorKeyLabel(vk int) string {
	if vk >= 'A' && vk <= 'Z' {
		return string(rune(vk))
	}
	if vk >= '0' && vk <= '9' {
		return string(rune(vk))
	}
	if vk >= 0x70 && vk <= 0x87 {
		return "F" + strconv.FormatInt(int64(vk-0x6f), 10)
	}
	if s, ok := inputMonitorKeyLabelMap[vk]; ok {
		return s
	}
	return fmt.Sprintf("VK%d", vk)
}

// inputMonitorPostThreadMessageValue 向输入监听线程 PostThreadMessageW 投递值。
// [S-sig 0x140869660, 320B]：newobject(2 字段) → LazyProc.Call(PostThreadMessageW, 4) →
// 失败 fmt.Errorf。体待线程消息结构专项还原。
func inputMonitorPostThreadMessageValue(a, b uint32) error {
	_, _ = a, b
	return nil
}

// Stop 停止输入监听服务（StopOwner(8) 转发）。
// [S-sig 0x1408641e0, 352B]：StopOwner(8) → duffcopy × 3 → 返回 owner。体待 owner 结构专项还原。
func (s *inputMonitorService) Stop() interface{} {
	_ = s
	return nil
}
