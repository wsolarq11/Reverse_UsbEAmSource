# 批次 262 · MonitorRectForWindow +1 [S] + VirtualScreenBounds 返回类型订正

## 目标

清理 HANDOFF 挂账的遗留订正项 `windowManagementVirtualScreenBounds` 返回类型错误
（image.Rectangle → 4×int32），并落地 windowmanagement 域的 `MonitorRectForWindow`（512B）。

## 基线 / 收口

| 指标 | 基线（batch 261 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2852 | 2853 |
| MARKED | 2852 | 2853 |
| S | 1320 | 1321 |
| S-inline | 36 | 36 |
| S-sig | 1456 | 1456 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1356 | 1357 |
| 真函数（S+S-inline+S-sig） | 2812（59.19%） | 2813（59.24%） |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 一、遗留订正：windowManagementVirtualScreenBounds 返回类型

**原实现（错）**：返回 `image.Rectangle`（4×int64，32 字节），`image.Rect(x,y,x+w,y+h)`。

**实证依据（两条独立证据）**：
1. 自身尾部 asm：`add edx, ecx` / `lea edi, [rax + rbx]` —— 全是 **32 位**加减，四个返回值
   经 rax/rbx/rcx/rdi 传回，宽度 4×4=16 字节。
2. 调用方 `getLauncherBackgroundMetrics`（0x140874520）：`sub ecx, eax` / `movsxd rcx, ecx` /
   `sub edi, ebx` / `movsxd rsi, edi` —— 先做 **32 位**减法再 movsxd 符号扩展。若返回 int64，
   就直接用 64 位 sub，不会出现 movsxd。

**订正**：返回类型改为 `windowManagementRECT`（4×int32），字段 Left/Top/Right/Bottom。

**影响面（已全量核查，仅 2 处）**：`bootstrapservice_callees.go` 的 `getLauncherBackgroundMetrics`
同步改用 `bounds.Right-bounds.Left` 等；`windowmanagement_windows.go` 的 `image` 导入随之移除
（否则 unused import 编译失败）。

## 二、本批落地（+1 [S]）

`windowManagementMonitorRectForWindow(hwnd uintptr) (int32, int32, int32, int32, error)` `[S 0x1409ee840]`：

```
hMonitor, _, _ := procMonitorFromWindow.Call(hwnd, 2)   // MONITOR_DEFAULTTONEAREST
if hMonitor == 0 { return 0,0,0,0, errors.New("无法定位目标窗口所在显示器") }
mi := &windowManagementMonitorInfo{Size: 40}            // 40 = sizeof(MONITORINFO)
r1, _, lastErr := procGetMonitorInfoW.Call(hMonitor, uintptr(unsafe.Pointer(mi)))
if r1 == 0 {
    if lastErr == nil || errors.Is(lastErr, syscall.Errno(0)) { lastErr = errors.New("未知错误") }
    return 0,0,0,0, fmt.Errorf("读取显示器边界失败: %w", lastErr)
}
return mi.Monitor.Left, mi.Monitor.Top, mi.Monitor.Right, mi.Monitor.Bottom, nil
```

新增 proc：procMonitorFromWindow / procGetMonitorInfoW（均 user32DLL）。

## 三、proc 身份内存实证（不猜）

| 全局槽位 | 实测 Name |
|---|---|
| 0x141BC1DB0 | GetSystemMetrics |
| 0x141BC1E60 | EnumDisplayMonitors |
| 0x141BC1E50 | MonitorFromWindow |
| 0x141BC1E58 | GetMonitorInfoW |

字符串实证：`读取显示器边界失败: %w`（31B @0x140C71045）、
`无法定位目标窗口所在显示器`（39B @0x140C7DF85）、`未知错误`（12B @0x140C4BA98）。

**Errno itab 第三次交叉验证**：本批 MonitorRectForWindow 的 `0x1409ee910+7+0x7e43d0 = 0x1411D2CE0`
与批次 259（GetCursorPoint）、批次 261（SetWindowLongPtr）算出的完全相同。

## 四、工具纠错：disp32 必须按字节解码

本批解析 callback 全局槽时，前次用「人工读 disp 十六进制」得出 0x141C5AB60，dump 出的却是
代码字节（0xA048A038A030A018 ...）。改为读 `.bin` 原始字节 → `[BitConverter]::ToInt32` 解 disp32
→ `next + disp`，结果一致但过程可复现。

另修正一处指令边界误判：`mov qword [rax], rcx` 前的 `lea` 位于 0x1409ee93c（非 0x1409ee946），
按正确边界重算得 0x140C4BA98，与其余函数的 "未知错误" 地址一致（自洽校验通过）。

新增 PE 段表解析用于定性：0x141C5AB60 属 .data（未初始化区，init 期写入），符合包级变量形态。

## 五、未落地（留下一批，需专项取证）

- **DisplayRects 384B**：涉及 GetSystemMetrics(SM_CMONITORS=0x50) → makeslice → EnumDisplayMonitors
  回调配对。回调经**全局 funcval 槽 0x141C5AB60**（.data，init 写入）传递，其构造方式（是否
  `syscall.NewCallback`、NewCallback 参数类型约束）尚未实证完毕，不做冒进落地。
- **targetFromWindowProcessPick 416B**：duffzero 清栈 + 大结构栈展开（返回体经栈传递），
  含 4 处 TrimSpace 与 filepathlite.Base 分支，结构待逐字段核对。
- MaybeWrapCursor 288B（依赖 WrappedCursorPoint 992B）。

## 下一批

DisplayRects（先实证 callback funcval 构造）；targetFromWindowProcessPick；MonitorRectForWindow
调用方联动核查。P=40。FUNCS 2853/4754 = 60.02%。
