# 批次 263 · DisplayRects +1 [S]（EnumDisplayMonitors 回调范式首落地）

## 目标

落地 windowManagementDisplayRects（384B）。这是本工程**第一个真正落地 `syscall.NewCallback`
回调管道**的函数——此前 oledBlackoutEnumTopLevelWindows 虽同范式，但仍停在 [S-sig] 存根。

## 基线 / 收口

| 指标 | 基线（batch 262 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2853 | 2854 |
| MARKED | 2853 | 2854 |
| S | 1321 | 1322 |
| S-inline | 36 | 36 |
| S-sig | 1456 | 1456 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1357 | 1358 |
| 真函数（S+S-inline+S-sig） | 2813（59.24%） | 2814（59.29%） |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 本批落地（+1 [S]）

`windowManagementDisplayRects() []windowManagementRECT` `[S 0x1409edce0]`：

```
count := getSystemMetrics(0x50)          // SM_CMONITORS
if count <= 0 { count = 1 }              // asm: test rax,rax; mov ecx,1; cmovle rax,rcx
rects := make([]windowManagementRECT, 0, count)
r1, _, _ := procEnumDisplayMonitors.Call(0, 0,
    windowManagementEnumDisplayMonitorCallback,     // BSS 全局槽
    uintptr(unsafe.Pointer(&rects)))
if r1 != 0 && len(rects) != 0 { return rects }
bounds := windowManagementVirtualScreenBounds()
if bounds.Right <= bounds.Left || bounds.Bottom <= bounds.Top { return nil }
return []windowManagementRECT{bounds}
```

新增 proc：procEnumDisplayMonitors（user32DLL）。

## 关键：callback 全局槽的定性（上一批挂起项，本批已证）

批次 262 曾因"回调构造方式未实证"而拒落。本批用两条证据闭环：

1. **槽位归属**：PE 段表显示 `.data` 的已初始化区止于 `0x141C0F800`（VA 0x141961000 +
   rawSz 0x2AE800），而槽位 `0x141C5AB60` 在其后 → 属 **BSS**（文件无内容，初值 0，
   init 期写入）。这正是包级 `var x = syscall.NewCallback(...)` 的形态。
2. **范式参照**：二进制内存在 `syscall.compileCallback` 符号（0x14007c660），且在
   `main_init`（0x140747bf0）与 `oledBlackoutEnumTopLevelWindows`（0x14091b379）中被调用。
   读 `win_enumTopLevelWindows.asm.txt` 确认该工程范式为：
   closure funcval（newobject 填 FUNC 指针 + 捕获变量）→ `syscall.compileCallback(fn, true)`
   → 包装进 args 数组 → LazyProc.Call。

据此落地为包级变量 `windowManagementEnumDisplayMonitorCallback = syscall.NewCallback(
windowManagementEnumDisplayMonitorProc)`，由 asm `mov rcx,[rip+disp]` 直接读取该槽，形态吻合。

## 验证：NewCallback 是运行时校验，必须靠 test 门禁兜住

`syscall.NewCallback` 的参数类型约束**不是编译期检查**，而是在包 init 阶段由
`syscall.compileCallback` 反射校验（要求：返回值为 uintptr 宽度、各实参不大于 uintptr 宽度）。
若签名非法，会在 init 期 panic，使**全部**测试失败——单跑 build/vet 发现不了。

本批回调签名 `func(hMonitor, hdcMonitor uintptr, lprcMonitor *windowManagementRECT,
dwData unsafe.Pointer) uintptr` 全部实参均为 uintptr 宽度、返回单 uintptr，`go test` 通过，
即证明 init 期未 panic、签名合法。这也是把 test 门禁列为三绿必需项的实证理由。

## 数据流闭环

回调侧（批次 258 落地）`*rects = append(*rects, r)` 与调用侧 `uintptr(unsafe.Pointer(&rects))`
配对；`len(rects)` 经 slice 头回读，构成完整「传入-填充-回读」链路。两批共同构成
EnumDisplayMonitors 管道的完整还原。

## 下一批

targetFromWindowProcessPick 416B（大结构栈展开，待逐字段核对）；MaybeWrapCursor 288B
（依赖 WrappedCursorPoint 992B）。P=40。FUNCS 2854/4754 = 60.04%。
