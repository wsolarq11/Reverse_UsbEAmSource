# 批次 259 · 窗口矩形 + 快照归属校验 +2 [S]（含批次 258 忠实度订正）

## 目标

落地 windowmanagement_windows.go 两个函数：GetWindowRect（384B）、ValidateSnapshotOwner（352B）。
同时订正批次 258 遗留的 GetCursorPoint 失败路径判据缺陷。

## 基线 / 收口

| 指标 | 基线（batch 258 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2846 | 2848 |
| MARKED | 2846 | 2848 |
| S | 1314 | 1316 |
| S-inline | 36 | 36 |
| S-sig | 1456 | 1456 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1350 | 1352 |
| 真函数（S+S-inline+S-sig） | 2806（59.02%） | 2808（59.07%） |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 本批落地（+2 [S]，均追加到 windowmanagement_windows.go）

1. `windowManagementGetWindowRect(hwnd uintptr) (int32, int32, int32, int32, error)` `[S 0x1409ee600]` =
   GetWindowRect.Call(hwnd,&rect)；成功→(Left,Top,Right,Bottom,nil)；失败→lastErr==nil 或
   errors.Is(lastErr,syscall.Errno(0)) 时 errors.New("未知错误")，再
   fmt.Errorf("读取窗口位置失败: %w",lastErr)，返回 (0,0,0,0,err)。

2. `windowManagementValidateSnapshotOwner(hwnd uintptr, pid uint32) error` `[S 0x1409ea780]` =
   hwnd==0 或 IsWindow(hwnd)==0 → errors.New("目标窗口已失效")；pid==0 → nil；
   GetWindowThreadProcessId(hwnd,&pid2)；pid2!=pid → errors.New("目标窗口句柄已被其他进程复用")；否则 nil。

新增 proc：procIsWindow / procGetWindowRect / procGetWindowThreadProcessId（均 user32DLL）。

## 批次 258 忠实度订正（重要）

`windowManagementGetCursorPoint` 失败路径原写 `lastErr == nil` 单判据，**漏掉了 `syscall.Errno(0)` 分支**。

订正依据：对照同库同源已固化函数 `getCursorScreenPoint`（0x14085cf00，同为 GetCursorPos 薄封装），
其 asm 明确为 `test rcx,rcx; je 短路` + `cmp rcx,ErrnoItab; je → runtime.ifaceeq(Errno(0))`，
即判据是 `lastErr == nil || lastErr == syscall.Errno(0)`。本批已同步订正 GetCursorPoint（本批
GetWindowRect 直接按订正后判据落地）。

## 关键知悉：proc 身份必须内存实证，不得语义猜测

本批三个 proc 名全部由 `LazyProc.Name` 字段内存读出实证，而非按函数语义推断：

| 全局槽位 | *LazyProc | 实测 Name |
|---|---|---|
| 0x141BC1DC8 | 0x141BD1B40 | IsWindow |
| 0x141BC1DD8 | 0x141BD1BC0 | GetWindowThreadProcessId |
| 0x141BC1E38 | 0x141BD1EC0 | GetWindowRect |

顺带实证 SetWindowLongPtr 的三个 proc（留待下批）：SetLastError（0x141BD1A40）、
SetWindowLongW（0x141BD1D00）、SetWindowLongPtrW（0x141BD1C80），格式串
"更新窗口样式失败: %w"（28B）。

## 关键纠错：RIP-relative 高位进位（第三次同型踩坑）

`0x1409EA7BF + 0x011D7609` 应为 `0x141BC1DC8`，我前次误算为 `0x1411D1DC8`（把 0x011D7609
的低位 0x1D7609 直接补到 0x140.. 上，忽略 `0x409EA7BF + 0x011D7609` 产生的进位）。
首次读出的 Name.ptr=0x20/len=5381816026 即为错址垃圾，凭「len 不合理」察觉并重算。
纪律固化：disp32 加到**下一条指令地址**，逐位进位，再用读出的 len 合理性自检。

## 未落地（留后续）

- FlagNames 544B / ExStyleNames 384B / StyleNames 128B（名字枚举链，依赖 8+ 个 flag 名字符串）。
- SetWindowLongPtr 544B（LockOSThread + defer UnlockOSThread + SetLastError/SetWindowLongW/
  SetWindowLongPtrW 三 proc 择优，proc 名已实证，逻辑待逐条核对）。
- MaybeWrapCursor 288B（依赖 WrappedCursorPoint 992B）。
- DisplayRects 384B、MonitorRectForWindow 512B、targetFromWindowProcessPick 416B、
  RestoreOpacity 544B、CaptureOpacitySnapshot 1024B、ApplyOpacity 1280B。

## 下一批

继续 gap_aggregate.txt 长度升序。P=40。FUNCS 2848/4754 = 59.90%。
