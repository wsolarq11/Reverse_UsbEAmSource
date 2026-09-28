# 批次 258 · 窗口管理 user32 薄封装 +4 [S]（256–320B）

## 目标

落地 windowmanagement_windows.go 剩余 5 个短函数中的 4 个：GetClassName（256B）、
GetWindowText（320B）、GetCursorPoint（320B）、EnumDisplayMonitorProc（320B）。
MaybeWrapCursor（288B）依赖未落地的 WrappedCursorPoint（992B），延后。

## 基线 / 收口

| 指标 | 基线（batch 257 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2842 | 2846 |
| MARKED | 2842 | 2846 |
| S | 1310 | 1314 |
| S-inline | 36 | 36 |
| S-sig | 1456 | 1456 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1346 | 1350 |
| 真函数（S+S-inline+S-sig） | 2802（58.94%） | 2806（59.02%） |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 本批落地（+4 [S]，全部追加到 windowmanagement_windows.go）

1. `windowManagementGetClassName(hwnd uintptr) string` `[S 0x1409ee220]` =
   hwnd==0→""；make([]uint16,256)→GetClassNameW.Call(hwnd,&buf[0],256)；n==0→""；
   否则 TrimSpace(syscall.UTF16ToString(buf[:n]))。新增 procGetClassNameW。

2. `windowManagementGetWindowText(hwnd uintptr) string` `[S 0x1409ee0e0]` =
   hwnd==0→""；GetWindowTextW.Call(hwnd,0,0) 两阶段取长度 n；n==0→""；
   make([]uint16,n+1)→Call(hwnd,&buf[0],n+1)；n2==0→""；否则 TrimSpace(UTF16ToString(buf[:n2]))。
   新增 procGetWindowTextW。

3. `windowManagementGetCursorPoint() (int32, int32, error)` `[S 0x1409edfa0]` =
   GetCursorPos.Call(&pt)；成功→(pt.X,pt.Y,nil)；失败→lastErr==nil 时用 errors.New("未知错误")，
   再 fmt.Errorf("读取鼠标位置失败: %w",lastErr)，返回 (0,0,err)。复用 gpu_pick_windows.go 的
   procGetCursorPos（本批首版重复声明被 build 拦下，已删除本文件重复项）。

4. `windowManagementEnumDisplayMonitorProc(hMonitor, hdcMonitor uintptr, lprcMonitor *windowManagementRECT, dwData unsafe.Pointer) uintptr` `[S 0x1409ede60]` =
   dwData==nil || lprcMonitor==nil → 1；矩形无效（Right<=Left || Bottom<=Top）→ 1；
   否则 `*rects = append(*rects, *lprcMonitor)` 后返回 1（EnumDisplayMonitors 回调，
   dwData 为 &[]windowManagementRECT）。

## 关键知悉

- **proc 冲突**：procGetCursorPos 已在 gpu_pick_windows.go:20 定义（GetCursorPos 同款薄封装）。
  本批首版重复声明被 `go build` 拦下（redeclared），删除本文件重复项后复用已有。
- **vet unsafeptr 纠错**：EnumDisplayMonitorProc 首版用 `dwData uintptr` + `unsafe.Pointer(dwData)`
  触发 `go vet` "possible misuse of unsafe.Pointer"（uintptr 参数直转指针）。改签名
  `dwData unsafe.Pointer` + `(*[]windowManagementRECT)(dwData)` 后 vet 通过——asm 层面
  test rdi,rdi 判 nil 与 uintptr 版完全一致，register ABI 无差。
- **GetCursorPoint error 链**：失败分支实测两个字符串常量（RIP-relative 校准后）——
  "未知错误"（12B，len=0xc）与 "读取鼠标位置失败: %w"（28B，len=0x1c），
  逻辑为 lastErr==nil → errors.New("未知错误")，再 fmt.Errorf 包装。

## 下一批

P=40。windowmanagement_windows.go 剩余短函数：MaybeWrapCursor 288B（依赖 WrappedCursorPoint
992B，需同批落地或先落地 WrappedCursorPoint）、GetWindowRect 384B、MonitorRectForWindow 512B。
继续按 gap_aggregate.txt 长度升序推进。FUNCS 2846/4754 = 59.87%。
