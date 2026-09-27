# 批次 255 · windowmanagement_windows 短函数 +5（角落守卫 / wrap 目标 / 窗口 LongPtr / 客户区矩形）

## 目标

落地 `backend/windowmanagement_windows.go` 的 5 个短函数：角落守卫带判定、光标 wrap 目标
（X/Y）、窗口 LongPtr 读取、客户区矩形读取。均为 gap_aggregate 长度升序中已存在文件内的
128–192B 短函数；`windowManagementStyleNames` 因依赖未落地常量本批跳过。

## 基线 / 收口

| 指标 | 基线（batch 254 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2826 | 2831 |
| MARKED | 2826 | 2831 |
| S | 1294 | 1299 |
| S-inline | 36 | 36 |
| S-sig | 1456 | 1456 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1330 | 1335 |
| 真函数（S+S-inline+S-sig） | 2786（58.60%） | 2791（58.71%） |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 本批落地（+5，全部 [S]）

### backend/windowmanagement_windows.go（+5 [S]，新增 import unsafe + user32 LazyProc）

1. `windowManagementPointInCornerGuard(x,y,left,top,right,bottom,radius int) bool` `[S 0x1409ed9a0]`
   = radius<=0 || right<=left || top>=bottom → false；margin=clamp(radius+4,0,96) 再截到
   width/height；margin<=0 → false；返回 `(左带||右带) && (上带||下带)`。
2. `windowManagementWrapTargetX(monitors []windowManagementRECT, y int32, direction bool) (int32, bool)`
   `[S 0x1409eda40]` = 遍历 y∈[Top,Bottom) 且 Right>Left；direction=true 时
   best=max(best,Right-1)，否则 best=min(best,Left)；found 时返回 direction? best-3 : best+3。
3. `windowManagementWrapTargetY(monitors []windowManagementRECT, x int32, direction bool) (int32, bool)`
   `[S 0x1409edb00]` = 遍历 x∈[Left,Right) 且 Bottom>Top；direction=true 时
   best=max(best,Bottom-1)，否则 best=min(best,Top)；found 时返回 direction? best-3 : best+3。
4. `windowManagementGetWindowLongPtr(hwnd uintptr, index int32) uintptr` `[S 0x1409ee320]`
   = 显式 `procGetWindowLongPtrW.Find()`（失败→nil→Call panic），打包 `[2]uintptr{hwnd,index}`，
   LazyProc.Call(2 args)，直接返回 r1。
5. `windowManagementGetClientRectValue(hwnd uintptr) (int,int,int,int)` `[S 0x1409ee780]`
   = `procGetClientRect.Call(hwnd,&rect)`；r1==0 → (0,0,0,0)；否则
   (Left, Top, Right-Left, Bottom-Top)（int32 符号扩展为 int）。

## 关键知悉

- **命名交叉实证**：WrapTargetX 名带「X」但输入是**垂直坐标 y**（找 y 所在显示器），返回
  wrap 后的**水平目标 x**（left/right±3）；WrapTargetY 反之。由 WrappedCursorPoint 调用点锁定：
  `WrapTargetX(monitors, y, dir)`（0x1409ed8a9, edi=r9d=y）与
  `WrapTargetY(monitors, x, dir)`（0x1409ed7b4, edi=[rsp+0x38]=x）。
- **windowManagementRECT 布局**：Left@+0/Top@+4/Right@+8/Bottom@+0xc（标准 RECT），与
  types_windowmgt.go 声明一致；WrapTargetX 读 [rax+4]=Top/[rax+0xc]=Bottom 判定垂直归属，
  WrapTargetY 读 [rax]=Left/[rax+8]=Right 判定水平归属。
- **GetWindowLongPtr 的 Find+cmovne**：全局值2（[0x141BC6DE8]=0x140B3A7A0）为
  `procGetWindowLongPtrW`（*LazyProc），全局值1（[0x141BC6DF8]=0）为 nil 回退；asm 语义是
  `proc := procGetWindowLongPtrW; if proc.Find()!=nil { proc=nil }; proc.Call(...)`，Find 失败即
  nil-recv Call 触发 panic（无 A 版本 fallback，两全局值证实非 A/W 切换）。

## 遗留订正（后续批次）

- `windowManagementVirtualScreenBounds` 当前落地返回 `image.Rectangle`（4×int64），但 asm
  0x1409edbc0 尾段 `add edx,ecx`/`lea edi,[rax+rbx]` 为 32 位运算，实返 4×int32
  （等价 `windowManagementRECT`）。其调用方 `getLauncherBackgroundMetrics`
  （bootstrapservice_callees.go @312）按 image.Rectangle 语义使用，联动订正需同一专项批次，
  本批不动（不影响 5 函数落地正确性）。
- `windowManagementStyleNames`（128B 0x1409eebe0）依赖未落地的 `windowManagementFlagNames`
  （544B 0x1409eede0）+ `windowManagementExStyleNames`（384B 0x1409eec60），留待后续。

## 下一批

P=40 不变。继续按 `gap_aggregate.txt` 长度升序落地已存在文件短函数；下一批优先
windowmanagement_windows.go 剩余短函数（GetClassName 256B / MaybeWrapCursor 288B /
EnumDisplayMonitorProc 320B 等）。FUNCS 2831/4754 = 59.55%。
