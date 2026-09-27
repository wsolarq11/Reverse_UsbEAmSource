# 批次 256 · 截图/桌面小组件短函数 +7（ShowWindow 常量 / 天气解析 / 画布 Bounds）

## 目标

落地 4 个文件内的 7 个短函数：截图预览与选区工具栏的 ShowWindow flags/command
（32B 返回常量）、桌面天气整型解析（64B）、天气服务 Shutdown（64B）、截图滚动画布
Bounds（64B）。均为 gap_aggregate 长度升序中的最短 top-level 函数。

## 基线 / 收口

| 指标 | 基线（batch 255 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2831 | 2838 |
| MARKED | 2831 | 2838 |
| S | 1299 | 1306 |
| S-inline | 36 | 36 |
| S-sig | 1456 | 1456 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1335 | 1342 |
| 真函数（S+S-inline+S-sig） | 2791（58.71%） | 2798（58.86%） |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 本批落地（4 文件 +7，全部 [S]）

### backend/screenshot_preview_windows.go（新增，+2 [S]）

1. `screenshotPreviewShowWindowFlags() int` `[S 0x14099e7a0]` = `mov eax,0x53; ret`（返回常量 0x53=83）。
2. `screenshotPreviewShowWindowCommand() int` `[S 0x14099e7c0]` = `mov eax,4; ret`（SW_SHOWNOACTIVATE）。

### backend/screenshot_selection_toolbar_windows.go（新增，+2 [S]）

3. `screenshotSelectionToolbarShowWindowCommand() int` `[S 0x1409ad200]` = `mov eax,8; ret`（SW_SHOWNA）。
4. `screenshotSelectionToolbarShowWindowFlags() int` `[S 0x1409ad220]` = `mov eax,0x53; ret`（0x53=83）。

### backend/desktopwidgets_weather.go（新增，+2 [S]）

5. `desktopWeatherInt(s string) (int, error)` `[S 0x1407c9640]` = `strings.TrimSpace` → `strconv.Atoi`，直接透传 (int, error)。
6. `(*desktopWidgetWeatherService)Shutdown()` `[S 0x1407c5180]` = s==nil → return；s.cancel==nil → return；否则 s.cancel()。

### backend/screenshot_scroll_canvas_windows.go（新增，+1 [S]）

7. `(*screenshotScrollingChunkedCanvas)Bounds() image.Rectangle` `[S 0x14099f520]` =
   c==nil || width<=0 || height<=0 → 零矩形；否则 `image.Rect(0,0,width,height)`。

## 关键知悉

- **ShowWindowFlags 常量 0x53=83**：两处（预览/选区工具栏）均返回 0x53，非标准 WS_*/WS_EX_*
  组合（0x53=0x40|0x10|0x02|0x01，0x02 无对应 WS_EX_* 常量），按 asm 直译还原为字面常量，
  语义名称留待调用者落地时再关联。
- **desktopWidgetWeatherService.Shutdown 偏移**：`cancel` 字段 @+0x28（types_desktopwidget.go
  store@0x00/network@0x08/ctx@0x18/cancel@0x28 布局吻合），asm 经 `[rax+0x28]` 读 funcval、
  `[rdx]` 取 fn 后 `call`，等价 `if s!=nil && s.cancel!=nil { s.cancel() }`。
- **Bounds 返回布局**：asm 正常分支 `xor eax; mov rbx,rax` 后 ret，rax=Min.X=0、rbx=Min.Y=0、
  rcx=width（Max.X）、rdi=height（Max.Y），即 `image.Rect(0,0,width,height)`；异常分支四寄存器
  清零 → 零矩形。screenshotScrollingChunkedCanvas 布局 width@0x00/height@0x08 吻合。
- 本批重新运行 `aggregate_gap.py`：total missing=2040、top-level=874（批次 255 落地后未重生成
  的旧文件已刷新，含 5 个 windowmanagement 函数已从缺口移除）。

## 下一批

P=40 不变。继续按 `gap_aggregate.txt` 长度升序落地已存在文件短函数；下一批优先
windowmanagement_windows.go 剩余短函数（GetClassName 256B / MaybeWrapCursor 288B /
EnumDisplayMonitorProc 320B / GetCursorPoint 320B 等）及 desktopwidgets_weather.go 的
desktopWeatherFloat（128B）。FUNCS 2838/4754 = 59.70%。
