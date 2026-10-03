# 批次 365 · 空闲定时器调度/覆盖层关闭处理/图像转不透明 RGBA/滚动画布构造 +4（FUNCS 3216）

## 目标

落地 4 个函数：`oledBlackoutService.scheduleIdleTimerAfter`、
`oledBlackoutService.handleOverlayDismiss`、`screenshotImageToOpaqueRGBA`、
`newScreenshotScrollingChunkedCanvas`。

## 基线 / 收口

| 指标 | 基线（batch 364 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3212 | 3216 |
| MARKED | 3212 | 3216 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1629 | 1633 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1541 | 1541 |
| USABLE | 1542 | 1542 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1541  FUNCS=3216  MARKED=3216  P=41  S-eq=1  S-inline=37  S-sig=1633  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1633 + 41 = 3216 = FUNCS`。
S-sig 1629→1633（+4）、FUNCS 3212→3216（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 scheduleIdleTimerAfter [S-sig 0x140902180, 416B]

newobject(closure) → time.AfterFunc → 持锁 idleRunCurrentLocked → true→存 timer(+0x128)；
false→stopTimer。

### 3.2 handleOverlayDismiss [S-sig 0x14090ad20, 416B]

lock → field(+0x88) 非空→return；dismissOverlayScreenLocked → 成功→清 field(+0xe8/+0xe0)。

### 3.3 screenshotImageToOpaqueRGBA [S-sig 0x14096bf60, 416B]

img nil/越界→nil；bounds→子矩形→image.NewRGBA → DrawMask×2。

### 3.4 newScreenshotScrollingChunkedCanvas [S-sig 0x14099f120, 416B]

rect 空/越界→error；面积>0x7270e00→error；newScreenshotScrollingCanvasChunk → 组装。

## G4 独立复核

- `backend/oledblackout.go`：+scheduleIdleTimerAfter [S-sig] +handleOverlayDismiss [S-sig]。
- `backend/screenshot_read.go`：+screenshotImageToOpaqueRGBA [S-sig]。
- `backend/screenshot_scroll_canvas_windows.go`：+newScreenshotScrollingChunkedCanvas [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3216/4754 = 67.65%。下一批：oled/filesearch 域
oledBlackoutService.clearAutoActivatedProfileIfNotVisibleLocked / VolumeIndex.baseEntryCountLocked。
