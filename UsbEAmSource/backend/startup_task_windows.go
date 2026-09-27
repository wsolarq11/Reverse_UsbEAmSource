// startup_task_windows.go — 开机自启任务域：命令输出解码与任务缺失判定（逆向还原）
// 研究用途。反汇编目标：UsbEAm_Launcher.exe (go1.25.12/PE32+/wails v3)
// 基线：docs/goresym/pipeline/tmp/{decodeWindowsCommandOutput,isTaskNotFoundMessage}.asm.txt
//
// 本域函数（批次 124 落地）：
//   decodeWindowsBytes (0x1408b2c40)          ANSI(codepage)→UTF-16→UTF-8
//   decodeWindowsCommandOutput (0x1408b2b20)  UTF-8 优先，回退 ACP/GBK
//   isTaskNotFoundMessage (0x1408b2e60)       schtasks "任务不存在" 判定

package main

import (
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"golang.org/x/sys/windows"
)

// decodeWindowsBytes 用指定代码页把 ANSI 字节解码为 UTF-8 字符串。
// [S 汇编 0x1408b2c40, 464B(0x1d0)] len==0 或 codePage==0 → ("",false)（@0x1408b2c65/6e）；
// MultiByteToWideChar(codepage,0,data,nil,0) 取长度（@0x1408b2ca3）失败(<=0)→("",false)
// （@0x1408b2cb1）；makeslice 分配 UTF-16 缓冲；二次 MultiByteToWideChar 填充
// （@0x1408b2d59）失败→("",false)；utf16.decode（@0x1408b2dce）→slicerunetostring
// （@0x1408b2de0）→(string,true)。
func decodeWindowsBytes(data []byte, codePage uint32) (string, bool) {
	if len(data) == 0 || codePage == 0 {
		return "", false
	}
	n, err := windows.MultiByteToWideChar(codePage, 0, &data[0], int32(len(data)), nil, 0)
	if err != nil || n <= 0 {
		return "", false
	}
	buf := make([]uint16, n)
	m, err := windows.MultiByteToWideChar(codePage, 0, &data[0], int32(len(data)), &buf[0], int32(n))
	if err != nil || m <= 0 {
		return "", false
	}
	return string(utf16.Decode(buf[:m])), true
}

// decodeWindowsCommandOutput 解码 Windows 命令行输出为可读字符串。
// [S 汇编 0x1408b2b20, 288B(0x120)] 空 → ""（@0x1408b2b40/bf7）；utf8.Valid → 原样
// TrimSpace（@0x1408b2b58/bdb）；否则 decodeWindowsBytes(output,GetACP)（@0x1408b2b7a）
// 成功→TrimSpace；失败→decodeWindowsBytes(output,936=GBK)（@0x1408b2b98）成功→TrimSpace；
// 再失败→原字节转 string 后 TrimSpace（@0x1408b2baf）。
func decodeWindowsCommandOutput(output []byte) string {
	if len(output) == 0 {
		return ""
	}
	if utf8.Valid(output) {
		return strings.TrimSpace(string(output))
	}
	if s, ok := decodeWindowsBytes(output, windows.GetACP()); ok {
		return strings.TrimSpace(s)
	}
	if s, ok := decodeWindowsBytes(output, 936); ok { // GBK
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(string(output))
}

// isTaskNotFoundMessage 判断 schtasks 输出是否表示"任务不存在"。
// [S 汇编 0x1408b2e60, 288B(0x120)] TrimSpace+ToLower（@0x1408b2e77/80）后依序检查 6 个
// 子串（stringslite.Index >= 0），全部明文：
//
//	11B "cannot find"（@0x1408b2e8f→@0x140c47591）
//	9B  "not found"（@0x1408b2ebb→@0x140c3f7cc）
//	30B "cannot find the file specified"（@0x1408b2ee7→@0x140c6f153）
//	9B  "找不到"（@0x1408b2f0e→@0x140c3f7d5）
//	9B  "不存在"（@0x1408b2f3a→@0x140c3f7de）
//	30B "系统找不到指定的文件"（@0x1408b2f62→@0x140c6f171）
func isTaskNotFoundMessage(msg string) bool {
	m := strings.ToLower(strings.TrimSpace(msg))
	if strings.Index(m, "cannot find") >= 0 {
		return true
	}
	if strings.Index(m, "not found") >= 0 {
		return true
	}
	if strings.Index(m, "cannot find the file specified") >= 0 {
		return true
	}
	if strings.Index(m, "找不到") >= 0 {
		return true
	}
	if strings.Index(m, "不存在") >= 0 {
		return true
	}
	if strings.Index(m, "系统找不到指定的文件") >= 0 {
		return true
	}
	return false
}
