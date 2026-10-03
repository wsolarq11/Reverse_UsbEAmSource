# 批次 274 · screenshot_pin_windows.go 整文件落地（+2 [S]，+2 FUNCS）

## 目标

1. **整文件落地 `screenshot_pin_windows.go`**（43 个未落地文件差集之一）：截图钉窗口原生
   样式/透明度层 2 函数全 `[S]`。
2. 新增 Windows API 代理：`dwmapiDLL` + `procDwmSetWindowAttribute` + `procSetLayeredWindowAttributes`。

## 基线 / 收口

| 指标 | 基线（batch 273 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2881 | 2883 |
| MARKED | 2881 | 2883 |
| S | 1361 | 1363 |
| S-inline | 36 | 36 |
| S-sig | 1443 | 1443 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1397 | 1399 |
| USABLE | 1397 | 1399 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build ./backend` EXIT=0；`go1.25.12 vet ./backend` EXIT=0；
`go1.25.12 test ./backend -run ScreenshotPin` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1399  FUNCS=2883  MARKED=2883  P=41  S-eq=0  S-inline=36  S-sig=1443  S=1363  USABLE=1399
```

分项自洽：`S + S-inline + S-sig + P = 1363 + 36 + 1443 + 41 = 2883 = FUNCS`。
S 1361→1363（+2）、FUNCS 2881→2883（+2）、FAITHFUL 1397→1399（+2）、P=41/S-sig=1443/UNMARKED=0 保持。

账目：新建 `screenshot_pin_windows.go` 2 函数 `[S]`（+2 S +2 FUNCS +2 FAITHFUL）。
4 个闭包 `func1`（screenshotPinApplyWindowChrome6/7、screenshotPinApplyWindowState1）属父函数
`createWebviewPinnedWindow`/`setClickThrough`/`setOpacity`（screenshot_pin.go 仍是 [S-sig] 存根），
随父函数升档时一并处理，不计入本批。

## G3 行为（asm 逐地址实证）

### 3.1 screenshotPinSetLayeredWindowOpacity [S 0x1409944a0, 256B]

`(hwnd uintptr, opacity float64)`，无返回值（morestack 仅保护 rax + xmm0）。

- `hwnd == 0` 直接返回（test rax,rax → 0x140994575）。
- clamp 到 `[0.2, 1.0]`：下界 0.2（@0x1411CD658）、上界 1.0（@0x1411CD6D0），
  ucomisd 双分支（else-if 结构，0x1409944d1 jmp 跳过第二 if）。
- alpha 换算：`opacity < 1.0`（1.0 阈值 @0x1411CD6D0）→ `int(opacity*255)`（×255.0 @0x1411CD7E8，
  cvttsd2si 截断向零），`<=0` 时置 1（cmovle）；否则 `ecx=0xffffffff` 低字节 = 255。
- `SetLayeredWindowAttributes(hwnd, 0, alpha, 2)`：LazyProc @0x141BC1CB0（Name 实证
  "SetLayeredWindowAttributes"），参数 (hwnd, 0, uintptr(alpha), LWA_ALPHA=0x2)。

### 3.2 screenshotPinSuppressWindowBorder [S 0x1409945a0, 160B]

`(hwnd uintptr)`，无返回值（morestack 仅保护 rax）。

- `hwnd == 0` 直接返回（test rax,rax → 0x1409945b7）。
- 否则 `attr = 0xfffffffe`（mov dword，DWMWA_COLOR_NONE = -2），
  `DwmSetWindowAttribute(hwnd, 0x22, &attr, 4)`：LazyProc @0x141BC2658（Name 实证
  "DwmSetWindowAttribute"），参数 (hwnd, DWMWA_BORDER_COLOR=0x22, uintptr(&attr), 4)。

## G4 独立复核

`screenshot_pin_windows.go`：新建，2 函数 `[S]` + `dwmapiDLL`/`procDwmSetWindowAttribute`/
`procSetLayeredWindowAttributes` 全局 + 3 常量（dwmwaBorderColor/dwmwaColorNone/lwaAlpha）。
`screenshot_pin_windows_test.go`：新建，hwnd==0 分支 smoke + clamp/alpha 换算等价复现锚定。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

截图钉窗口原生样式/透明度层整文件闭环：2 函数全 `[S]`，DwmSetWindowAttribute/
SetLayeredWindowAttributes 两个 proc 与 clamp 常量全部 asm/rodata 实证。依赖此底层的
`screenshot_pin.go` 存根（setClickThrough/setOpacity/createWebviewPinnedWindow 及其闭包）仍
[S-sig]，属后续升档专项。P=41 持平。FUNCS 2883/4754 = 60.64%。未落地文件差集 43→42。
