# 批次 238 — screenshotNativePinDrawOverlay 签名订正（1 升档）

## 基线 / 收口

| 指标 | 基线（批次 237 收口） | 收口（批次 238） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1442 | **1443** |
| P | 54 | **53** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2734（57.50%） | **2735**（57.53%） |

SHA256 `14a69e1f4f7cbe9e83853963064996b220d9e4f9606766672b84360c0c89a05c`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/screenshot_pin_native_windows.go`：`screenshotNativePinDrawOverlay` 升档 `[S-sig]`，
签名 `(img *image.RGBA, bounds application.Rect)` →
`(img *image.RGBA, bounds application.Rect, showCloseButton, showOpacityPanel, showHighlight bool, tick int, opacity float64)`。

### 签名证据（0x140992de0）

morestack 序言（第 213-234 行）保存 10 槽，逐寄存器实证：

| 项 | 类型 | 证据 |
|---|---|---|
| rax | img *image.RGBA | 体内 `mov rbx,[rax+0x20]` … `[rax+0x38]` 读 image.RGBA.Rect（第 47-50 行） |
| rbx/rcx/rdi/rsi | bounds application.Rect（4 int） | 透传 `CloseButtonRect`(0x140992960)/`OpacityPanelRect`(0x1409929c0) 四 int |
| r8b | showCloseButton bool | 第 38 行 `test r8b,r8b; jne` 门 close 按钮；非颜色分量 |
| r9b | showOpacityPanel bool | 第 40-41 行 `test r9b,r9b; je` 早退；第 95-98 行门 opacity 面板绘制 |
| r10b | showHighlight bool | 第 62-64 行 `movzx edx,[rsp+0xd2]; test dl,dl; je` 门高亮二次 DrawRect |
| r11 | tick int | 第 15-18 行 `test r11,r11; jle` + `bt r11d,0; jb` 作「正且偶」判定，动画帧计数 |
| xmm0 | opacity float64 | 第 8 行保存，第 121 行透传 `OpacityThumbRect`(0x140992c80) |

返回：三条 `ret` 路径（第 206/209/212 行）均无返回寄存器写入 = void。

### 关键纠正

旧注释「r11 疑为 tick/边框厚度、语义未定」已解：r11 是**动画 tick 计数**——`test r11,r11;jle`
（tick≤0 跳过）+ `bt r11d,0;jb`（奇数跳过）即「正且偶才绘制高亮边框」的闪烁逻辑，与色值无关。
三个 byte 不是 color.RGBA 分量，而是三路布尔门控（close 按钮 / opacity 面板 / 高亮），
体内硬编码色值（0xf/0x17/0x2a、0x1e/0x29/0x3b 等）独立于入参。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`vet ./backend` EXIT=0；
  `test ./backend` `ok changeme/backend`。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1443 / P=53 / UNMARKED=0`。
- **G3 行为**：纯签名订正（体空→体空），该函数无调用方，无行为变更；既有测试全量 PASS。
- **G4 review**：`screenshot_pin_native_windows.go`（1 升档）。

## 遗留（下一批）

- P 已降至 53。截图域剩余复杂 P：captureScreenshotAreaPNG（7 参数，参数4/6/7 需
  qrCodeNativeSelectionOptions 字段映射）、captureScreenshotWindowSelectionWithOptions（5 参数 +
  WindowProcessPickResult 大结构返回）、newScreenshotWindowSelectionSession（快照 6 字 + 4 栈参数）、
  resolveControlHoverForWindow、screenshotAccessibleHitTest（VARIANT+error 组合）。
- launcherupdate_runtime.go 9 个 [P]（beginLauncherUpdateTask 6 寄存器返回结构、多参 workspace 直传）。
- filelocator walkRoot/processFile（值传大结构体 prepared + 未还原聚合结构体）。
- oledblackout_windows 25（WinRT/COM 匿名上下文）。
