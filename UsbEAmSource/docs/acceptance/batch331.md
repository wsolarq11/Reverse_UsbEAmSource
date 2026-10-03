# 批次 331 · 覆盖层聚焦/分层透明度/滚动帧就绪/底部裁剪 +4（FUNCS 3081）

## 目标

落地 4 个短函数：`oledBlackoutService.focusVisibleOverlayLocked`（覆盖层聚焦）、
`screenshotPreviewSetLayeredWindowOpacity`（分层窗口透明度）、
`waitScreenshotScrollingFrameReady`（滚动帧就绪等待）、
`screenshotScrollingChunkedCanvas.resolveBottomTrim`（底部裁剪）。

## 基线 / 收口

| 指标 | 基线（batch 330 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3077 | 3081 |
| MARKED | 3077 | 3081 |
| S | 1492 | 1492 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1506 | 1510 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1529 | 1529 |
| USABLE | 1530 | 1530 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1529  FUNCS=3081  MARKED=3081  P=41  S-eq=1  S-inline=37  S-sig=1510  S=1492  USABLE=1530
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1492 + 37 + 1 + 1510 + 41 = 3081 = FUNCS`。
S-sig 1506→1510（+4）、FUNCS 3077→3081（+4）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 oledBlackoutService.focusVisibleOverlayLocked [S-sig 0x14090dea0, 256B]

visibleOverlayOrderLocked 迭代 → overlay map(+0xf0) 非空则 Focus + scheduleFocusRetryLocked。

### 3.2 screenshotPreviewSetLayeredWindowOpacity [S-sig 0x14099e6a0, 256B]

hwnd nil 守卫 → clamp opacity [0,1] → *255 截断 → SetLayeredWindowAttributes。

### 3.3 waitScreenshotScrollingFrameReady [S-sig 0x1409a3480, 256B]

shouldCancelScreenshotScrolling → deadline 循环 → time.Sleep(min(1s))。

### 3.4 screenshotScrollingChunkedCanvas.resolveBottomTrim [S-sig 0x14099fcc0, 256B]

自底向上扫描行 → screenshotScrollingRowLooksPureBlack 判黑。

## G4 独立复核

- `backend/oledblackout.go`：+focusVisibleOverlayLocked [S-sig]。
- `backend/screenshot_scroll_windows.go`：+screenshotPreviewSetLayeredWindowOpacity
  +waitScreenshotScrollingFrameReady +resolveBottomTrim [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3081/4754 = 64.81%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
