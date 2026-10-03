# 批次 364 · 目录资产 URL 解析/观众窗口关闭/选择边界计算/预览就绪等待 +4（FUNCS 3212）

## 目标

落地 4 个函数：`resolvePluginCatalogAssetURL`、`pluginWindowService.CloseAudienceOwned`、
`screenshotSelectionToolbarBoundsForSelection`、`screenshotPreviewWindowService.waitUntilReady`。

## 基线 / 收口

| 指标 | 基线（batch 363 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3208 | 3212 |
| MARKED | 3208 | 3212 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1625 | 1629 |
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
FAITHFUL=1541  FUNCS=3212  MARKED=3212  P=41  S-eq=1  S-inline=37  S-sig=1629  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1629 + 41 = 3212 = FUNCS`。
S-sig 1625→1629（+4）、FUNCS 3208→3212（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 resolvePluginCatalogAssetURL [S-sig 0x14092c380, 416B]

TrimSpace → 空→error；url.Parse scheme http/https→String；否则 base → Parse → JoinPath。

### 3.2 CloseAudienceOwned [S-sig 0x14092fa20, 416B]

TrimSpace×2 → lock(+0x0) → windows(+0x10)[owner] → window(+0x88) nil→nil；
EqualFold → close 回调。

### 3.3 screenshotSelectionToolbarBoundsForSelection [S-sig 0x1409ab5a0, 416B]

越界→默认矩形(0x334,0x1b2)；居中/钳位算术 → bounds。

### 3.4 waitUntilReady [S-sig 0x140999a60, 416B]

timeout nil→false；time.NewTimer → selectgo → 超时→false；ready→shouldDisplay。

## G4 独立复核

- `backend/plugin_discovery.go`：+resolvePluginCatalogAssetURL [S-sig]。
- `backend/bootstrapservice_state_deps.go`：+pluginWindowService.CloseAudienceOwned [S-sig]。
- `backend/screenshot_selection_toolbar_windows.go`：+screenshotSelectionToolbarBoundsForSelection [S-sig]。
- `backend/screenshot_preview_windows.go`：+screenshotPreviewWindowService.waitUntilReady [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3212/4754 = 67.56%。下一批：screenshot/oled 域
screenshotImageToOpaqueRGBA / oledBlackoutService.scheduleIdleTimerAfter。
