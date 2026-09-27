// AUTO-RECONSTRUCTED — DOMAIN: gpu_pick_debug
// 研究用途。文件对应蓝图 gpu_pick_debug.go（source_funcs.txt 1779-1785）。
//
// 三函数均 [S] 汇编逐寄存器实证（go1.25.12 可编译 + 功能一致）：
//
//	gpuPickDebugLog                   0x14085b700  main.gpuPickDebugLog
//	ensureGPUPickDebugLogger          0x14085b9e0  main.ensureGPUPickDebugLogger
//	ensureGPUPickDebugLogger.func1    0x1409f7380  闭包（sync.Once 执行体）
//	resolveGPUPickDebugConfiguration  0x14085ba20  main.resolveGPUPickDebugConfiguration
//	parseGPUPickDebugBool             0x14085bd40  已在 gpu.go（批次 153）
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// gpuPickDebugLogger 调试日志全局态（单例结构体）。
// [S 汇编实证 0x141c13400]：一次成型 64B 连续块，字段偏移由各读写点 RIP 相对位移反查：
//
//	once     sync.Once   @0x00（12B，done 位于 @0x00）
//	enabled  bool        @0x0c（cmp byte [rip+…] 于 gpuPickDebugLog +0x56）
//	path     string      @0x10（ptr @0x10 / len @0x18，闭包三出口均写）
//	file     *os.File    @0x20（cmp qword [rip+…] 于 gpuPickDebugLog +0x66）
//	err      error       @0x28（type @0x28 / data @0x30，MkdirAll/OpenFile 失败时写）
//	mu       sync.Mutex  @0x38（lea [rip+…] 于 gpuPickDebugLog 加锁点）
var gpuPickDebugLogger = struct {
	once    sync.Once
	enabled bool
	path    string
	file    *os.File
	err     error
	mu      sync.Mutex
}{}

// resolveGPUPickDebugConfiguration 解析 GPU 拖放调试开关与日志文件路径。
// [S 汇编 0x14085ba20, 0x134]：返回 (enabled bool, filePath string) —— 寄存器 ABI
// 证得 bool 在 AX、string 在 BX(ptr)/CX(len)。
//
// 语义（逐指令）：
//
//	filePath = TrimSpace(Getenv("USBEAM_GPU_PICK_DEBUG_FILE"))      @0x140c682b9(26B)
//	enabled  = filePath != ""
//	debugVal = TrimSpace(Getenv("USBEAM_GPU_PICK_DEBUG"))           @0x140c5f6dd(21B)
//	if debugVal != "" { enabled = parseGPUPickDebugBool(debugVal) 或 未识别默认 true }
//	遍历 os.Args[1:]（跳过 argv[0]，rdi=ptr+0x10、rsi=len-1）：
//	  "--gpu-pick-debug"（16B，0x69702d7570672d2d/0x67756265642d6b63 双 qword 比较）→ enabled=true
//	  HasPrefix "--gpu-pick-debug="（17B @0x140c572fd）→ enabled = parse(arg[17:]) 或 默认 true
//	  HasPrefix "--gpu-pick-debug-file="（22B @0x140c6123c）→ filePath = TrimSpace(arg[22:])，非空则 enabled=true
func resolveGPUPickDebugConfiguration() (bool, string) {
	filePath := strings.TrimSpace(os.Getenv("USBEAM_GPU_PICK_DEBUG_FILE"))
	enabled := filePath != ""
	if v := strings.TrimSpace(os.Getenv("USBEAM_GPU_PICK_DEBUG")); v != "" {
		b, ok := parseGPUPickDebugBool(v)
		if !ok {
			b = true
		}
		enabled = b
	}
	for _, arg := range os.Args[1:] {
		arg = strings.TrimSpace(arg)
		switch {
		case arg == "--gpu-pick-debug":
			enabled = true
		case strings.HasPrefix(arg, "--gpu-pick-debug="):
			b, ok := parseGPUPickDebugBool(strings.TrimSpace(arg[len("--gpu-pick-debug="):]))
			if !ok {
				b = true
			}
			enabled = b
		case strings.HasPrefix(arg, "--gpu-pick-debug-file="):
			p := strings.TrimSpace(arg[len("--gpu-pick-debug-file="):])
			filePath = p
			if p != "" {
				enabled = true
			}
		}
	}
	return enabled, filePath
}

// ensureGPUPickDebugLogger 惰性初始化调试日志器（sync.Once 幂等）。
// [S 汇编 0x14085b9e0, 0x33]：内联 Once.Do 快路径（once.done @0x141c13400
// 判 0 → lea &once + lea &func1 → once.doSlow）。
func ensureGPUPickDebugLogger() {
	gpuPickDebugLogger.once.Do(func() {
		// ensureGPUPickDebugLogger.func1 [S 汇编 0x1409f7380, 0x486]。
		enabled, filePath := resolveGPUPickDebugConfiguration()
		if !enabled {
			return
		}
		ws := resolveWorkspaceLayout()
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			filePath = filepath.Join(ws.Root, "gpu-pick-debug.log") // @0x140c5a0c2(18B)
		} else if !filepath.IsAbs(filePath) {
			filePath = filepath.Join(ws.Root, filePath)
		}
		filePath = filepath.Clean(filePath)
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil { // 0x1ed
			gpuPickDebugLogger.enabled = true
			gpuPickDebugLogger.path = filePath
			gpuPickDebugLogger.err = err
			return
		}
		f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644) // 0x441/0x1a4
		if err != nil {
			gpuPickDebugLogger.enabled = true
			gpuPickDebugLogger.path = filePath
			gpuPickDebugLogger.err = err
			return
		}
		gpuPickDebugLogger.enabled = true
		gpuPickDebugLogger.path = filePath
		gpuPickDebugLogger.file = f
		now := time.Now().Format("2006-01-02 15:04:05.000") // @0x140c62ecf(23B)
		line := fmt.Sprintf("%s [%d] ===== GPU 拖放调试会话开始 =====\n", now, os.Getpid())
		_, _ = f.Write([]byte(line))
	})
}

// gpuPickDebugLog 写入一条带时间戳与 PID 的调试日志。
// [S 汇编 0x14085b700, 0x220]：先 ensureGPUPickDebugLogger；enabled=false 或
// file==nil 则直接返回；否则 fmt.Sprintf(format,args...) → time.Now().Format(23B)
// → GetCurrentProcessId → fmt.Sprintf("%s [%d] %s\n", ts, pid, msg) → 加锁 → defer 解锁
// → file.Write。deferwrap1 即解锁闭包（0x14085b980）。
func gpuPickDebugLog(format string, args ...any) {
	ensureGPUPickDebugLogger()
	if !gpuPickDebugLogger.enabled || gpuPickDebugLogger.file == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	ts := time.Now().Format("2006-01-02 15:04:05.000")
	line := fmt.Sprintf("%s [%d] %s\n", ts, os.Getpid(), msg) // @0x140c47502(11B)
	gpuPickDebugLogger.mu.Lock()
	defer gpuPickDebugLogger.mu.Unlock()
	_, _ = gpuPickDebugLogger.file.Write([]byte(line))
}
