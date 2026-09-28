# 批次 261 · SetWindowLongPtr +1 [S]（LockOSThread + 双 proc 择优）

## 目标

落地 windowmanagement_windows.go 的 SetWindowLongPtr（544B）。该函数是 windowmanagement 域
最复杂的一个短函数：涉及线程锁定、错误码预清、SetWindowLongPtrW/SetWindowLongW 运行时择优、
以及一处**与同域其他函数语义相反**的 Errno 判据。

## 基线 / 收口

| 指标 | 基线（batch 260 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2851 | 2852 |
| MARKED | 2851 | 2852 |
| S | 1319 | 1320 |
| S-inline | 36 | 36 |
| S-sig | 1456 | 1456 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1355 | 1356 |
| 真函数（S+S-inline+S-sig） | 2811（59.13%） | 2812（59.19%） |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 本批落地（+1 [S]）

`windowManagementSetWindowLongPtr(hwnd uintptr, index int32, newValue uintptr) error` `[S 0x1409ee3e0]`：

```
runtime.LockOSThread()
defer runtime.UnlockOSThread()
procSetLastError.Call(0)                       // 预清错误码
proc := procSetWindowLongPtrW
if procSetWindowLongPtrW.Find() != nil {       // 32 位回退
    proc = procSetWindowLongW
}
r1, _, lastErr := proc.Call(hwnd, uintptr(index), newValue)
if r1 != 0 { return nil }
if errno, ok := lastErr.(syscall.Errno); ok && errno == 0 { return nil }   // ← 注意：此处 Errno(0) 视为成功
if lastErr == nil { lastErr = errors.New("未知错误") }
return fmt.Errorf("更新窗口样式失败: %w", lastErr)
```

## 关键发现：同一 Errno(0) 判据在本域内语义相反

`Errno(0)` 在两个函数里指向完全相反的结论，这是本批最重要的取证成果：

| 函数 | `lastErr` 为 `syscall.Errno(0)` 时 | 理由 |
|---|---|---|
| GetCursorPoint / GetWindowRect | **失败** → 替换为 "未知错误" → 返回 error | API 返回 0 即失败，Errno(0) 说明无错误详情，用兜底文案 |
| SetWindowLongPtr | **成功** → 直接返回 nil | SetWindowLong 的 r1 是「旧值」，0 是合法旧值；改由 lastErr 判定，Errno(0) 表示无错误发生 |

两处 asm 分支结构确实不同：GetCursorPoint 在 errno==0 时走 `errors.New("未知错误")` 构造路径；
SetWindowLongPtr 在 errno==0 时置 `rcx=0, rdi=0` 后经 `test rcx,rcx; je →0x1409ee59e`（成功返回 nil）。
**不可因"同域同判据"就复用写法**——本批已按各自 asm 分别落地。

## 全部事实的内存实证（不猜）

| 事项 | 实证方式 | 结果 |
|---|---|---|
| defer 目标 | dump funcval @0x141096E10（= 0x1409ee42c+0x6a89e4） | funcval[0]=0x1400525E0 = **runtime.UnlockOSThread** |
| SetLastError 归属 | 读 LazyProc 槽 0x141BC1DA8 | **SetLastError** |
| 择优两 proc | 读槽 0x141BC1DF0 / 0x141BC1E00 | **SetWindowLongPtrW** / **SetWindowLongW** |
| Errno itab | `0x1409ee4f5+7+0x7e47eb` | **0x1411D2CE0**（与批次 259 GetCursorPoint 算出的完全一致，交叉验证通过） |
| 格式串 | dump 0x140C6C1E0 | `更新窗口样式失败: %w`（28B） |
| SetLastError 是否来自 x/sys/windows | 全量扫 `golang.org/x/sys@v0.46.0/windows/*.go` | **不存在**（该包只有 GetLastError），故须自建 kernel32 proc |

因 x/sys/windows 无 SetLastError，本批新增：
`kernel32DLL = windows.NewLazySystemDLL("kernel32.dll")` + `procSetLastError = kernel32DLL.NewProc("SetLastError")`。

## 下一批

DisplayRects 384B；MaybeWrapCursor 288B（依赖 WrappedCursorPoint 992B，需先落地后者）；
MonitorRectForWindow 512B；targetFromWindowProcessPick 416B。P=40。FUNCS 2852/4754 = 60.00%。
