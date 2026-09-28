package main

import "strconv"

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
