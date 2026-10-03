# 批次 283 · screenshot 滚动取消状态机 + 预览绘制/光标恢复 +5（全 [S]，FUNCS 2930）

## 目标

闭合截图滚动的取消判定状态机（shouldCancelScreenshotScrolling / waitScreenshotScrollingCancelable，
新增 screenshotScrollingCancelState 类型），还原原生预览 WM_PAINT 绘制（paint）与覆盖层关闭后的
光标恢复链（setScreenshotOverlayCursor / restoreScreenshotCursorAfterOverlay，新增 3 个 user32 LazyProc）。

## 基线 / 收口

| 指标 | 基线（batch 282 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2925 | 2930 |
| MARKED | 2925 | 2930 |
| S | 1402 | 1407 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1444 | 1444 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1439 | 1444 |
| USABLE | 1440 | 1445 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1444  FUNCS=2930  MARKED=2930  P=41  S-eq=1  S-inline=37  S-sig=1444  S=1407  USABLE=1445
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1407 + 37 + 1 + 1444 + 41 = 2930 = FUNCS`。
S 1402→1407（+5）、FUNCS 2925→2930（+5）、FAITHFUL 1439→1444（+5）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 (*screenshotNativePreviewWindow)paint [S 0x14099da60, 224B]

`(hwnd uintptr)`。BeginPaint(hwnd,&ps) 返回非 0 → redraw()；EndPaint(hwnd,&ps)。
LazyProc 名 "BeginPaint"/"EndPaint"（0x141BC29F0/0x141BC29F8）。

### 3.2 shouldCancelScreenshotScrolling [S 0x1409a32a0, 128B]

`(state *screenshotScrollingCancelState) bool`。nil→false；read 得 (pressed,clicked)；
armed 分支：state.pressed=pressed，pressed→false，否则 armed=false 且 false；
clicked→state.pressed=pressed 且 true；pressed→!state.pressed（上升沿）且 state.pressed=pressed；
否则 false。state 为 2 字节 {pressed bool@+0, armed bool@+1}。

### 3.3 waitScreenshotScrollingCancelable [S 0x1409a3380, 256B]

`(timeout time.Duration, state *...) bool`。timeout<=0→shouldCancel；先查 shouldCancel→true；
deadline=now.Add(timeout)；循环 Until<=0→shouldCancel；sleep=min(remaining,16ms)；
每次 sleep 后 shouldCancel→true。

### 3.4 setScreenshotOverlayCursor [S 0x140973ae0, 192B]

`(cursor uint16) bool`。ensureScreenshotCursorVisible()；LoadCursorW(0,cursor) 返回 0→false；
否则 SetCursor(hCursor)→true。LazyProc 名 "LoadCursorW"/"SetCursor"（0x141BC27F8/0x141BC2A18）。

### 3.5 restoreScreenshotCursorAfterOverlay [S 0x140973be0, 96B]

无参无返回。ReleaseCapture()（0 参）；setScreenshotOverlayCursor(0x7f00)；
!screenshotCursorCurrentlyShowing()→adjustScreenshotCursorVisibility(true)。
LazyProc 名 "ReleaseCapture"（0x141BC2910）。

## G4 独立复核

- `screenshot_preview_native_windows.go`：+`paint`。
- `screenshot_scroll_windows.go`：+`screenshotScrollingCancelState` 类型 +`shouldCancelScreenshotScrolling`
  +`waitScreenshotScrollingCancelable`（import time）。
- `screenshot_cursor_windows.go`：+`procLoadCursorW`/`procSetCursor`/`procReleaseCapture` LazyProc
  +`setScreenshotOverlayCursor`+`restoreScreenshotCursorAfterOverlay`。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

5 个函数落地（全 [S]），新增 1 类型 + 3 LazyProc；零外部依赖。
FUNCS 2930/4754 = 61.63%。下一批：截图滚动几何（setScreenshotCursorPosition/waitFrameReady/
analyzeFrameProgress/rankAppendCandidate 等 128-256B 短函数）+ screenshot preview 剩余。
