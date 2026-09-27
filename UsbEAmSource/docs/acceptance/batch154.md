# 批次 154 验收（gpu_pick_debug.go 全量落地：调试日志器 + 配置解析链）

日期：2026-09-25
子批次：gpu_pick_debug.go（sync.Once 惰性日志器 / 环境变量 + os.Args 配置解析 / 带时间戳 PID 日志写入）

## 目标

落地 gpu_pick_debug.go 全部 6 个蓝图条目（source_funcs.txt 1779-1785），使该文件 100% 落地：
`gpuPickDebugLog`、`gpuPickDebugLog.deferwrap1`（编译器自动生成）、`ensureGPUPickDebugLogger`、
`ensureGPUPickDebugLogger.func1`（闭包）、`resolveGPUPickDebugConfiguration`、`parseGPUPickDebugBool`
（批次 153 已落地于 gpu.go）。本批补齐前 5 项中的 3 个具名函数（+2 编译器生成体）。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go1.25.12 test -count=1 ./backend` | ok (0.402s) |
| 配置解析 test | `go1.25.12 test -run TestResolveGPUPickDebugConfiguration -v` | 8/8 PASS |

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2640 / S=1055 / S-inline=35 / S-sig=1353 / P=197 / UNMARKED=0`。

**真函数 = 1055 + 35 + 1353 = 2443 / 4754 = 51.4%**（相对批次 153 的 2440 增 +3）。

重建产物 SHA256：`C341F124FE5CCE88FF2067BE04A51CC0E101CE884D077CB7B9B52254455D4F04`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细（3 具名函数 [S]，`backend/gpu_pick_debug.go`）

| 函数 | 蓝图行 | VA | asm 行 |
|---|---|---|---|
| resolveGPUPickDebugConfiguration | 96-134 | 0x14085ba20 | 185 |
| ensureGPUPickDebugLogger | 55-96 | 0x14085b9e0 | 33 |
| gpuPickDebugLog | 41-53 | 0x14085b700 | 134 |
| ensureGPUPickDebugLogger.func1（闭包） | 56-93 | 0x1409f7380 | 267 |
| gpuPickDebugLog.deferwrap1（解锁） | 51-55 | 0x14085b980 | 自动生成 |

## G4 关键实证结论

1. **resolveGPUPickDebugConfiguration 返回 (bool, string)**：寄存器 ABI 证得 bool 在 AX、
   string 在 BX(ptr)/CX(len)。`al` 判开关、`rbx/rcx` 承载文件路径贯穿 os.Args 循环。
2. **环境变量**：`USBEAM_GPU_PICK_DEBUG_FILE`（26B @0x140c682b9）→ TrimSpace 为初始路径；
   `USBEAM_GPU_PICK_DEBUG`（21B @0x140c5f6dd）→ 非空则 parseGPUPickDebugBool（未识别默认 true）
   覆盖 enabled。`enabled` 初值 = `filePath != ""`。
3. **os.Args 循环跳过 argv[0]**：`rdi = ptr(os.Args) + 0x10`、`dec rsi`（len-1），即遍历
   `os.Args[1:]`。三分支：`--gpu-pick-debug`（16B，双 qword 立即数 0x69702d7570672d2d /
   0x67756265642d6b63）→ enabled=true；`--gpu-pick-debug=`（17B @0x140c572fd）→ parse(arg[17:])；
   `--gpu-pick-debug-file=`（22B @0x140c6123c）→ filePath=TrimSpace(arg[22:])，非空则 enabled=true。
4. **闭包路径解析**：`resolveWorkspaceLayout().Root` 为基座；filePath 空 → Join(Root,
   "gpu-pick-debug.log" @0x140c5a0c2)；相对 → Join(Root, filePath)；绝对 → 原样。随后 Clean →
   Dir → MkdirAll(0o755) → OpenFile(O_WRONLY|O_CREATE|O_APPEND=0x441, 0o644)。
5. **全局态单例结构体 @0x141c13400（64B）**：once@0x00(12B)、enabled@0x0c(1B)、path@0x10(16B)、
   file@0x20(8B)、err@0x28(16B，MkdirAll/OpenFile 失败时写 error 接口，本域只写不读)、
   mu@0x38(8B)。sync.Once 12B 布局（done uint32 + Mutex，无 padding）证得 enabled 落在 0x0c。
6. **gpuPickDebugLog 写链**：ensure → enabled/file 双判 → Sprintf(format,args...) →
   time.Now().Format("2006-01-02 15:04:05.000" @0x140c62ecf) → GetCurrentProcessId →
   Sprintf("%s [%d] %s\n" @0x140c47502, ts, pid, msg) → mu.Lock（CAS 快路径 + lockSlow）→
   defer mu.Unlock → file.Write。初始会话日志 `%s [%d] ===== GPU 拖放调试会话开始 =====\n`
   （49B @0x140c87b1f）。

## 残留 [P]（不触及）

gpu_windows.go（getGPUPreferenceState 0x14085dda0 等 7 入口，Win32 注册表/DXGI/进程枚举）与
gpu_pick_windows.go（pickWindowProcessForService 等，窗口/进程解析）。这些入口最终会调用本批
gpuPickDebugLog 与批次 153 的 gpu 纯函数。
