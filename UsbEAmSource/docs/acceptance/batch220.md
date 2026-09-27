# 批次 220 — screenshot_uia 域两 [P] 升档 [S-sig]

## 基线 / 收口

| 指标 | 基线（批次 219 收口） | 收口（批次 220） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1411 | **1413** |
| P | 85 | **83** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2703 | **2705**（56.90%） |

SHA256 `5056F736359D969BB1293F133B16286C1E6480390FCB1820C2CB9E1EB8AF551A`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/screenshot_uia_windows.go` 两个 [P] 升档 [S-sig]，均为「序言寄存器/栈参数远超当前
形参」类阻断，逐寄存器对齐后参数分组唯一。

### shouldPreferWindowOverControlAtPoint（0x1409af160）

`(x,y,hwnd)` → `(ctrlRect, winBounds image.Rectangle, x, y int, ctrlFound, winFound bool)`。

序言：rax/rbx/rcx/rdi = ctrlRect（4 int）、rsi/r8/r9/r10 = winBounds（4 int）、
栈 [rsp+0x70]/[0x78] = x/y、栈 [rsp+0x80]/[0x81] = ctrlFound/winFound。逻辑：winFound→true、
ctrlFound→false、否则 screenshotRectMatchesBounds(ctrlRect,winBounds) 或
screenshotPointNearWindowEdge(x,y,winBounds,10)。

### screenshotPreferUIAElementHitInfo（0x1409b2ce0）

`(a,b image.Rectangle)` → `(a image.Rectangle, aControlType uint32, aOK bool, b image.Rectangle, bControlType uint32, bOK bool)`。

序言：rax/rbx/rcx/rdi = a、esi = aControlType、r8b = aOK；b 整体走栈（b 需 6 word 而剩余
寄存器不足，Go ABI「struct 不能完整入剩余寄存器则整体走栈」）。返回 bool（多条路径
mov eax,1 / xor eax / setl / setg）。函数体含 controlType 分类跳转表（0xc358 基址）与
rect 比较，未逐条翻译。

## 关键知悉

- readScreenshotUIAElementHitInfo 返回 (rect, uint32, bool, error) 的 8 寄存器布局，被
  screenshotPreferUIAElementHitInfo 的参数分组（aControlType/aOK、bControlType/bOK）复用印证。
- screenshotUIADeepestElementHitInfoAtPoint 尾迹返回 6 寄存器，确证返回 (unsafe.Pointer,
  image.Rectangle, bool)；但其参数含 depth(int:rsi, cmp 8) + state(指针:r8) + 栈大结构
  （匿名 hit 累加器），类型未唯一确定，保持 [P]。
- initializeScreenshotCOMThreadMode 参数确证为 bool(al)（tolerateModeChange，透传
  RPC_E_CHANGED_MODE 判定）；返回类型为 (装箱指针, error) 而装箱具体类型未唯一确定，保持 [P]。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1413 / P=83 / UNMARKED=0`。
- **G3 行为**：纯签名修正（体零值→体零值），无行为变更；既有测试全量回归 PASS。
- **G4 review**：`screenshot_uia_windows.go`（2 函数签名修正）；vet/test/build 复验通过。

## 遗留（下一批）

- screenshot_uia 域剩余 [P]：screenshotControlBoundsAtPointThroughOverlaySessionWithTimeout、
  screenshotControlBoundsAtPointWithWindowFallbackWithTimeoutAndPriority、
  screenshotFallbackControlBoundsAtPointWithTimeoutAndPriority（三函数同构：中间 rcx..r11
  qword + 栈槽语义未定，含 priority/mode/回调接口）、screenshotAccessibleHitTest（返回
  VARIANT 结构 + error）、initializeScreenshotCOMThreadMode（装箱指针具体类型）、
  screenshotAccessibleLocation（bx/cx/di/si hit 结构）、screenshotUIADeepestElementHitInfoAtPoint、
  screenshotUIASelectableAncestorHitInfoAtPoint（depth/state/栈 hit 累加器匿名类型）。
- 其余域 P：oledblackout_windows 27、screenshot_windows 15、launcherupdate_runtime 10、
  screenshot_pin 7、screenshot_uia_worker_windows 4、oledblackout_hotkey_windows 3、
  filelocator_runtime 2、qrcode_windows 3、+ 4 文件各 1。
- P 已降至 83。下一批优先 oledblackout_windows（27 个，体量大，多为主题绘制类，参数分组
  较规则）或 screenshot_windows（15 个）。
