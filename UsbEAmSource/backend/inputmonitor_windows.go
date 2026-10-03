package main

import (
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
