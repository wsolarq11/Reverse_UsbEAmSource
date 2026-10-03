# 批次 332 · 滚动帧捕获/缩略图 PNG/滚动分块编码/原生显示等待 +4（FUNCS 3085）

## 目标

落地 4 个短函数：`captureScreenshotScrollingFrame`（滚动帧捕获）、
`buildScreenshotThumbnailPNGFromPath`（缩略图 PNG）、
`encodeScreenshotScrollingChunkedCanvas`（滚动分块编码）、
`screenshotPreviewWindowService.waitBeforeRevealNative`（原生显示等待）。

## 基线 / 收口

| 指标 | 基线（batch 331 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3081 | 3085 |
| MARKED | 3081 | 3085 |
| S | 1492 | 1492 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1510 | 1514 |
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
FAITHFUL=1529  FUNCS=3085  MARKED=3085  P=41  S-eq=1  S-inline=37  S-sig=1514  S=1492  USABLE=1530
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1492 + 37 + 1 + 1514 + 41 = 3085 = FUNCS`。
S-sig 1510→1514（+4）、FUNCS 3081→3085（+4）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 captureScreenshotScrollingFrame [S-sig 0x1409a2d40, 224B]

backend nil → defaultScreenshotScreenCaptureBackend；调用 backend[+0x18] 转发参数。

### 3.2 buildScreenshotThumbnailPNGFromPath [S-sig 0x14096b200, 256B]

读图 → 缩放 → PNG 编码。

### 3.3 encodeScreenshotScrollingChunkedCanvas [S-sig 0x1409a0640, 256B]

分块画布 → 编码输出。

### 3.4 screenshotPreviewWindowService.waitBeforeRevealNative [S-sig 0x140999de0, 256B]

全局计数 <=0 → shouldDisplayNative；否则 time.NewTimer + chanrecv1 → shouldDisplayNative。

## G4 独立复核

- `backend/screenshot_scroll_windows.go`：+captureScreenshotScrollingFrame
  +buildScreenshotThumbnailPNGFromPath +encodeScreenshotScrollingChunkedCanvas [S-sig]。
- `backend/screenshot_preview_windows.go`：+waitBeforeRevealNative [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3085/4754 = 64.89%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
