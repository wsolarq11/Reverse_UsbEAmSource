# 批次 265 · windowmanagement 光标环绕链 +3 [S]（WrappedCursorPoint 992B / MaybeWrapCursor 288B / targetFromWindowProcessPick 416B）

## 目标

落地 HANDOFF batch264 遗留的 windowmanagement 域三个函数：光标环绕计算核心
`WrappedCursorPoint`（992B）、其封装 `MaybeWrapCursor`（288B）、窗口拾取结果组装
`targetFromWindowProcessPick`（416B）。三者是本域「PickTarget → cursorWrapLoop」链路的收口
与「窗口拾取 → 目标结构」链路的组装端，均无未落地依赖，可同批落地。

## 基线 / 收口

| 指标 | 基线（batch 264 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2857 | 2860 |
| MARKED | 2857 | 2860 |
| S | 1325 | 1328 |
| S-inline | 36 | 36 |
| S-sig | 1456 | 1456 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1361 | 1364 |
| USABLE | 1361 | 1364 |
| 真函数（S+S-inline+S-sig） | 2817（59.25%） | 2820（59.32%） |

`go1.25.12 build/vet/test -tags production ./backend` 全 EXIT=0。

## G1 编译

`go1.25.12 build -tags production -trimpath ./backend` EXIT=0；
`go1.25.12 vet -tags production ./backend` EXIT=0；
`go1.25.12 test -count=1 -p=1 -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1364  FUNCS=2860  MARKED=2860  P=40  S-eq=0  S-inline=36  S-sig=1456  S=1328  USABLE=1364
```

三函数均升档 `[S]`（S 1325→1328，FUNCS 2857→2860），P/S-sig/UNMARKED 均持平。

## G3 行为（对拍测试 + asm 逐寄存器实证）

### 3.1 windowManagementWrappedCursorPoint [S 0x1409ed5c0]（992B）

签名（morestack 段 spill 8 寄存器实证）：`(x, y int32, monitors []windowManagementRECT,
wrapX, wrapY bool, guardPx int) (int32, int32, bool)`，返回 AX/BX/CX。

控制流逐地址实证：

1. 线性扫描 monitors 找首个 `Left<=x<Right && Top<=y<Bottom`；未命中 → `(x,y,false)`
   （AX/BX 未改，0x1409ed62f→0x1409ed646→0x1409ed6fd）。
2. `Right<=Left || Top>=Bottom` → `(x,y,false)`（0x1409ed64c/0x1409ed655）。
3. `windowManagementPointInCornerGuard(x,y,left,top,right,bottom,guardPx)` 命中 →
   `(x,y,false)`（0x1409ed6ed，AX/BX 恢复原值）。
4. 水平 wrap：`wrapX && x<=left` 探测 `(x-1,y)` 是否落在某 monitor；不在 →
   `WrapTargetX(monitors,y,true)`（最右，`best-3`）。`wrapX && x>=right-1` 探测 `(x+1,y)`；
   不在 → `WrapTargetX(monitors,y,false)`（最左，`best+3`）。
5. 垂直 wrap：`wrapY && y<=top` 探测 `(x,y-1)` → `WrapTargetY(monitors,x,true)`（最下）；
   `wrapY && y>=bottom-1` 探测 `(x,y+1)` → `WrapTargetY(monitors,x,false)`（最上）。
6. **关键语义**：水平与垂直均基于**原始 (x,y)**（0x1409ed736/0x1409ed76a 读 `[rsp+0x38]`
   原始 x，非已 wrap 的 newX）；返回 `ok = hOK|vOK`（0x1409ed7c1 `or ecx,ebx`）；
   相邻 monitor 命中时跳过 wrap 保持原值（0x1409ed8f7/0x1409ed899/0x1409ed808）。

黄金对拍用例（`windowmanagement_windows_test.go`，期望值来自 asm 显式路径 + 已落地
WrapTargetX/Y 的 `best∓3` 逻辑，非凭空造值）：interior / corner-guard / left-edge-wrap(→1916) /
right-edge-wrap(→3) / top-edge-wrap(→1076) / bottom-edge-wrap(→3) / dual-adjacent-no-wrap /
dual-left-edge-wrap(→3836) / both-disabled / empty-monitors 共 10 例全 PASS。

### 3.2 windowManagementMaybeWrapCursor [S 0x1409ed4a0]（288B）

签名（prologue spill 6 寄存器 + 调用方 cursorWrapLoop 0x1409ed1d3 实证）：
`(wrapX, wrapY bool, monitors []windowManagementRECT, guardPx int) bool`。

`GetCursorPoint()` 失败（rcx=err!=nil）→ false；`WrappedCursorPoint(...)` ok=false → false；
否则 `newobject([2]uintptr)` 打包 movsxd 符号扩展后的 (newX,newY)，`procSetCursorPos.Call`
后无条件返回 true（SetCursorPos 返回值被丢弃，0x1409ed550 `mov eax,1`）。

### 3.3 targetFromWindowProcessPick [S 0x1409eea40]（416B）

签名（12 参数字 = 9 寄存器 + 3 栈，实参槽 `[rsp+0x40]/[rsp+0x50]/[rsp+0x58..0x80]/
[rsp+0x88]/[rsp+0x90..0x98]`）：

```go
func targetFromWindowProcessPick(processPath string, processID uint32,
    processName, displayName, title string, hwnd uintptr, iconData string) WindowManagementTarget
```

- `duffzero+0x142` 清零 128 字节 = `WindowManagementTarget`（9 字段 0x80，types_windowmgt.go
  L68-78），IconRef/IconURL 保持空串。
- `Path = TrimSpace(processPath)`；`ProcessName = TrimSpace(processName)`，空时且 Path 非空时
  回退 `filepath.Base(Path)`（0x1409eeab8，编译符号 `internal/filepathlite.Base`）。
- `DisplayName = TrimSpace(displayName)`，空时回退 `TrimSpace(title)`，再空回退 `ProcessName`
  （0x1409eeb7f `test rbx,rbx; cmove`）。
- `Title =` **原始** title（0x1409eeb62 直读 `[rsp+0x78]/[rsp+0x80]`，不 Trim）。
- 调用方 PickTarget（0x1409e4e33 全库唯一 `E8 rel32` 命中）以 duffcopy+0x310 把 128B 复制到
  `[rsp+0x70]`，随后校验 `[rsp+0x340]`(HWND)!=0、`[rsp+0x348]`(ProcessID)!=0、
  `[rsp+0x360..0x368]`(offset 0x20 = Path) TrimSpace!=0，完全印证布局。

黄金对拍用例（all-populated / name-empty-fallback-base / display-fallback-to-name /
all-empty 共 4 例）全 PASS。

## proc 身份内存实证

| 全局槽位 | 实测 LazyProc.Name |
|---|---|
| 0x141BC1DC0 | SetCursorPos（本批新增 procSetCursorPos） |
| 0x141BC1DB0 | GetSystemMetrics（交叉验证） |
| 0x141BC1E60 | EnumDisplayMonitors（交叉验证） |

`0x1409ed536` 的 `mov rcx,[rip+0x11d4883]` 目标经 Python 显式计算
`hex(0x1409ed53d + 0x11d4883) = 0x141BC1DC0`，与既有 proc 槽地址族 `0x141BC1xxx` 一致。

## G4 独立复核

仅改 `backend/windowmanagement_windows.go`（新增 `path/filepath` import + `procSetCursorPos`
proc + 三函数体），新建 `backend/windowmanagement_windows_test.go`（黄金对拍，不计入
count_funcs 因脚本排除 `_test`）。无既有函数签名/行为变更；全量测试回归 PASS。

## 下一批

继续 windowmanagement 域缺口：cursorWrapLoop / cursorWrapLoop.func1 / deferwrap1（
0x1409ecde0 / 0x1409ed360 / 0x1409ed440，含 time.NewTicker + selectgo 两 case 与 mutex
临界区闭包）；以及 gap_aggregate 剩余短函数。P=40。FUNCS 2860/4754 = 60.16%。
