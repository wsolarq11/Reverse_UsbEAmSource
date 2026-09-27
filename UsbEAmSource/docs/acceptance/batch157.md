# 批次 157 验收（gpu_pick_windows.go 窗口拾取叶函数：7 函数 [S] + 2 方法包装器对齐）

日期：2026-09-25
子批次：gpu_pick_windows.go 窗口/进程解析叶链（屏幕点拾取 + 进程路径/名称解析 + 窗口进程解析）

## 目标

落地 gpu_pick_windows.go 蓝图（source_funcs.txt 1786-1797）中**无跨域依赖的 7 个叶函数**，全量
`[S]`；将 `pickWindowProcessForService` 存根签名修正为正确形态 `[S-sig]`；对齐 2 个
BootstrapService 方法包装器签名。

7 个叶函数：`getCursorScreenPoint`、`windowFromScreenPoint`、`getForegroundWindow`、
`getWindowAncestor`、`resolveProcessPathByPID`、`resolveProcessNameByPID`、
`resolveWindowProcessPath`。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go1.25.12 test ./backend` | ok (0.396s) |
| 黄金用例 | `go1.25.12 test -run 'TestResolveProcessNameByPIDZero\|TestResolveWindowProcessPathZeroHWND\|TestGetWindowAncestorZeroHWND' -v` | 全 PASS |

黄金用例说明：三条零值早退分支（pid==0、hwnd==0）在任何 Win32 I/O 之前返回，可安全实测；
分别验证错误串 `进程 ID 无效` / `窗口句柄无效` 逐字一致，`getWindowAncestor(0,·)` 返回 0。

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2657 / S=1077 / S-inline=35 / S-sig=1348 / P=197 / UNMARKED=0`。

**真函数 = 1077 + 35 + 1348 = 2460 / 4754 = 51.7%**（相对批次 156 的 2453 增 +7 [S]）。

重建产物 SHA256：`B30CDCAACF42D18F473406EB7E25A89E77174FA3379BE2486BAB9646105D2281`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细

| 函数 | 蓝图行 | VA | 档位 | 说明 |
|---|---|---|---|---|
| resolveWindowProcessPath | 243-411 | 0x14085c540 | [S] | hwnd==0→报错；GetWindowThreadProcessId→resolveProcessPathByPID→成功日志；返 (path,pid,err) |
| resolveProcessPathByPID | 265-286 | 0x14085c740 | [S] | OpenProcess(0x1000)→defer CloseHandle→QueryFullProcessImageName→UTF16ToString |
| resolveProcessNameByPID | 289-321 | 0x14085cae0 | [S] | pid==0→报错；CreateToolhelp32Snapshot→Process32First/Next 遍历→TrimSpace 名称 |
| getCursorScreenPoint | 324-337 | 0x14085cf00 | [S] | GetCursorPos LazyProc(1 arg)→(pt.X,pt.Y,err) |
| windowFromScreenPoint | 337-384 | 0x14085d000 | [S] | WindowFromPoint LazyProc(1 arg)→(hwnd,err) |
| getForegroundWindow | 349-361 | 0x14085d100 | [S] | GetForegroundWindow LazyProc(0 arg)→(hwnd,err) |
| getWindowAncestor | 361-371 | 0x14085d1a0 | [S] | hwnd==0→0；GetAncestor LazyProc(2 args)→result |
| pickWindowProcessForService | 51-83 | 0x14085be20 | [S-sig] | 存根签名修正为 `(bs *BootstrapService)(WindowProcessPickResult,error)` |
| PickGPUPreferenceTargetByDrop | 2901-2905 | 0x140790020 | [S] | 方法包装器：pick 后取 `result.Path`，err 透传 |
| PickWindowProcess | 2905-2909 | 0x140790100 | [S] | 方法包装器：`return pickWindowProcessForService(bs)` 透传 |

## G4 关键实证结论

1. **LazyProc.Call 错误惯式复刻**：三个窗口叶函数沿用批次 152 确立的 `qrCodeBitBlt` 惯式
   ——`r1!=0 → 成功`；`lastErr==nil || errors.Is(lastErr, syscall.Errno(0)) → 通用/良性`；否则 `return lastErr`。
   `errors.Is(lastErr, syscall.Errno(0))` 的 ifaceeq 分支（test rcx/type 判定）逐寄存器对照。
2. **resolveWindowProcessPath 返回序**：`(path string, pid uint32, err error)` = 返回寄存器
   (AX,BX)=path、(CX)=pid、(DI,SI)=err；asm 错误/成功两条路径寄存器重排一致。
3. **resolveProcessPathByPID 栈缓冲 64KiB**：`var buf [0x8000]uint16`（rep stosq 0x2000 qword 清零），
   `QueryFullProcessImageName(handle, 0, &buf[0], &size)`，`size==0 → "无法读取目标进程路径"`、
   转换后空 → `"目标进程路径为空"`。
4. **resolveProcessNameByPID 遍历语义**：`Process32First` 失败即返 err；命中 `ProcessID==pid` →
   `TrimSpace(UTF16ToString(ExeFile[:]))`，空 → `"目标进程名为空"`；`Process32Next` 耗尽 →
   `"未找到目标进程"`。`ProcessEntry32.Size = 0x238`（unsafe.Sizeof，duffcopy 模板字节级对照）。
5. **pickWindowProcessForService 依赖阻断**：主体调用未落地的 `beginLauncherScreenshotCapture` /
   `endLauncherScreenshotCapture` / `resolveLauncherWindow` / `showLauncherWindow` 与 `[P]`
   `pickWindowProcessFromScreenshotSelection`，故本批仅落正确签名 `[S-sig]`（旧存根签名
   `() (interface{}, error)` 错误，已修正）。
6. **两个方法包装器签名修正**：`PickGPUPreferenceTargetByDrop` 由 `result.(string)` 伪断言改为
   `result.Path` 直取；`PickWindowProcess` 由 `(interface{}, error)` 改为
   `(WindowProcessPickResult, error)`（asm 0x140790100 纯透传，duffcopy 结构体经栈）。

## 残留 [P] / 未落地（不触及）

`pickWindowProcessForService` 主体 + `.func1` 闭包（恢复启动器窗口回调，调
`endLauncherScreenshotCapture`/`showLauncherWindow`）留待启动器窗口管理专项批次；ShortcutInfo
布局对齐（bool@0x78 + target@0x08）待 resolveShortcutInfoWithIconResolver 专项批次。

## 已知偏差（诚实记录）

无。本批 7 叶函数均为 asm 直译，Win32 调用、错误分支与返回序逐条对照；两处存根签名修正均以
asm 寄存器/栈布局为据。
