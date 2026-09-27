// AUTO-RECONSTRUCTED — DOMAIN: oledblackout 调试日志（[S]/[S-sig] 汇编实证）
// 研究用途
//
// 契约来源：
//   - ensureOLEDBlackoutDebugLogger(0x14090ee60, 50B)  sync.Once 门卫
//   - resolveOLEDBlackoutDebugConfiguration(0x14090eea0, 992B)  环境变量 + 启动参数扫描
//   - parseOLEDBlackoutDebugBool(0x14090f160, 191B)  布尔字面量解析
//   - oledBlackoutDebugLog(0x14090eb80, 640B)  日志行落盘
//
// 常量经 .rodata 解码确凿：
//   - "USBEAM_OLED_BLACKOUT_DEBUG_FILE"（31B，调试日志文件路径环境变量）
//   - "USBEAM_OLED_BLACKOUT_DEBUG"（26B，调试使能环境变量）
//   - "2006-01-02 15:04:05.000"（23B，时间戳格式，time.Time.Format 入参）
//   - "%s [%d] %s\n"（11B，日志行格式）
//   - "--oled-blackout-debug" / "--oled-blackout-debug=" / "--oled-blackout-debug-file="（CLI 覆盖标志）
package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// OLED 熄屏调试日志全局态（.data 全局：Once / *os.File / bool 使能 / Mutex）。
var (
	oledBlackoutDebugOnce    sync.Once
	oledBlackoutDebugFile    *os.File
	oledBlackoutDebugEnabled bool
	oledBlackoutDebugMu      sync.Mutex
)

// ensureOLEDBlackoutDebugLogger 一次性初始化调试日志器。
// [S 汇编 0x14090ee60, 50B(0x32)]：sync.Once.Do(resolveOLEDBlackoutDebugConfiguration)。
func ensureOLEDBlackoutDebugLogger() {
	oledBlackoutDebugOnce.Do(resolveOLEDBlackoutDebugConfiguration)
}

// resolveOLEDBlackoutDebugConfiguration 解析环境变量/启动参数，装配调试日志全局态。
// [S-sig 汇编 0x14090eea0, 992B(0x3e0)]：环境变量部分逐条对位；启动参数切片扫描（--oled-blackout-debug[=]/
// --oled-blackout-debug-file=）为 [S-sig] 近似，待参数切片来源（疑似 os.Args）专项实证。
func resolveOLEDBlackoutDebugConfiguration() {
	filePath := strings.TrimSpace(os.Getenv("USBEAM_OLED_BLACKOUT_DEBUG_FILE"))
	debugFlag := strings.TrimSpace(os.Getenv("USBEAM_OLED_BLACKOUT_DEBUG"))

	// 汇编 0x14090eef2 setne dl：enabled 初值 = filePath != ""。
	enabled := filePath != ""
	if debugFlag != "" {
		// 汇编 0x14090eefa parseOLEDBlackoutDebugBool；!ok 时默认 true（mov eax,1 @0xef04）。
		if v, ok := parseOLEDBlackoutDebugBool(debugFlag); ok {
			enabled = v
		} else {
			enabled = true
		}
	}

	// [S-sig] CLI 覆盖扫描：逐个参数 TrimSpace 后匹配
	//   "--oled-blackout-debug"（21B）→ enabled=true
	//   "--oled-blackout-debug="（22B 前缀）→ 后缀 parseOLEDBlackoutDebugBool 覆盖 enabled
	//   "--oled-blackout-debug-file="（27B 前缀）→ 后缀作为 filePath
	for _, arg := range os.Args[1:] {
		a := strings.TrimSpace(arg)
		switch {
		case a == "--oled-blackout-debug":
			enabled = true
		case strings.HasPrefix(a, "--oled-blackout-debug="):
			if v, ok := parseOLEDBlackoutDebugBool(strings.TrimPrefix(a, "--oled-blackout-debug=")); ok {
				enabled = v
			} else {
				enabled = true
			}
		case strings.HasPrefix(a, "--oled-blackout-debug-file="):
			filePath = strings.TrimPrefix(a, "--oled-blackout-debug-file=")
		}
	}

	oledBlackoutDebugEnabled = enabled
	if filePath != "" {
		if f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644); err == nil {
			oledBlackoutDebugFile = f
		}
	}
}

// parseOLEDBlackoutDebugBool 解析调试布尔字面量。
// [S 汇编 0x14090f160, 191B(0xbf)]：TrimSpace+ToLower 后按
//
//	"0"/"no"/"off"/"false" → (false, true)
//	"1"/"on"/"yes"/"true"  → (true, true)
//	其余 → (false, false)
//
// （magic 逐字节实证：0x30/0x31、0x6f6e"no"/0x6e6f"on"、0x666f"of"+f、0x6579"ye"+s、0x65757274"true"、0x736c6166"fals"+e）
func parseOLEDBlackoutDebugBool(s string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "0", "no", "off", "false":
		return false, true
	case "1", "on", "yes", "true":
		return true, true
	default:
		return false, false
	}
}

// oledBlackoutDebugLog 写调试日志行。
// [S 汇编 0x14090eb80, 640B(0x280)]：ensure → enabled/文件非空则 fmt.Sprintf(format, args...) →
// time.Now().Format("2006-01-02 15:04:05.000") → getCurrentProcessId →
// fmt.Sprintf("%s [%d] %s\n", ts, pid, msg) → 锁 → os.File.Write。
func oledBlackoutDebugLog(format string, args ...interface{}) {
	ensureOLEDBlackoutDebugLogger()
	if !oledBlackoutDebugEnabled || oledBlackoutDebugFile == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	ts := time.Now().Format("2006-01-02 15:04:05.000")
	line := fmt.Sprintf("%s [%d] %s\n", ts, os.Getpid(), msg)
	oledBlackoutDebugMu.Lock()
	defer oledBlackoutDebugMu.Unlock()
	_, _ = oledBlackoutDebugFile.Write([]byte(line))
}
