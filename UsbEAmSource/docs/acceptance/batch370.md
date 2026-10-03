# 批次 370 · 二维码解码事件/提供器错误快照/桌面天气凭证/鼠标点覆盖层定位 +4（FUNCS 3236）

## 目标

落地 4 个函数：`BootstrapService.emitQRCodeDecoded`、
`desktopWidgetWeatherService.providerErrorSnapshot`、`desktopWeatherCredential`、
`oledBlackoutService.visibleOverlayForMousePointLocked`。

## 基线 / 收口

| 指标 | 基线（batch 369 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3232 | 3236 |
| MARKED | 3232 | 3236 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1649 | 1653 |
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
FAITHFUL=1541  FUNCS=3236  MARKED=3236  P=41  S-eq=1  S-inline=37  S-sig=1653  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1653 + 41 = 3236 = FUNCS`。
S-sig 1649→1653（+4）、FUNCS 3232→3236（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 emitQRCodeDecoded [S-sig 0x140799d80, 448B]

TrimSpace(text) 空→return；ensureLauncherWindowForShow→err→return；
revealLauncherWindowForQRCodeDecode → emitOrQueueQRCodeDecoded。

### 3.2 providerErrorSnapshot [S-sig 0x1407c72c0, 448B]

makemap_small → 遍历 field(+0x40) → mapassign_faststr 复制 → 返回。

### 3.3 desktopWeatherCredential [S-sig 0x1407c8860, 448B]

provider=="OpenMeteo"→nil；map 命中→TrimSpace 空→nil；unprotectDesktopWidgetSecret→err→error。

### 3.4 visibleOverlayForMousePointLocked [S-sig 0x14090bc80, 448B]

collectScreensLocked → 遍历 screens → TrimSpace+map 命中 → 矩形包含→返回；
否则 visibleOverlayForKeyboardInputLocked。

## G4 独立复核

- `backend/bootstrapservice_window.go`：+emitQRCodeDecoded [S-sig]。
- `backend/desktopwidgets_weather.go`：+providerErrorSnapshot [S-sig] +desktopWeatherCredential [S-sig]。
- `backend/oledblackout.go`：+visibleOverlayForMousePointLocked [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3236/4754 = 68.07%。下一批：filesearch/oled 域
VolumeIndex.activeEntryCountLocked / VolumeIndex.overlayStatsLocked。
