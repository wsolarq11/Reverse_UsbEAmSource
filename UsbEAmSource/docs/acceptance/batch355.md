# 批次 355 · 插件窗口屏幕解析/缩略图读取/预览显示判定/纯黑行判定 +4（FUNCS 3176）

## 目标

落地 4 个函数：`resolvePluginWindowScreen`、`screenshotPreviewWindowService.readThumbnailPNG`、
`screenshotPreviewWindowService.shouldDisplay`、`screenshotScrollingRowLooksPureBlack`。

## 基线 / 收口

| 指标 | 基线（batch 354 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3172 | 3176 |
| MARKED | 3172 | 3176 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1589 | 1593 |
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
FAITHFUL=1541  FUNCS=3176  MARKED=3176  P=41  S-eq=1  S-inline=37  S-sig=1593  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1593 + 41 = 3176 = FUNCS`。
S-sig 1589→1593（+4）、FUNCS 3172→3176（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 resolvePluginWindowScreen [S-sig 0x1409316c0, 352B]

screenManager(+0x310).GetAll → 名字空→GetPrimary；EqualFold(TrimSpace) 匹配 → 返回 screen。

### 3.2 readThumbnailPNG [S-sig 0x1409963c0, 352B]

lock(+0x8) → asset(+0xe0) 空→error → launcherAssetService.ReadBytes。

### 3.3 shouldDisplay [S-sig 0x140999640, 352B]

lock(+0x8) → 字段(+0xa8/+0xb0/+0x18/+0x20) 匹配 → 返回 flag(+0xb1)==0。

### 3.4 screenshotScrollingRowLooksPureBlack [S-sig 0x1409a3fe0, 352B]

越界→false；step=max(1,行宽/100)；采样计数 → 纯黑占比判定。

## G4 独立复核

- `backend/bootstrapservice_state_deps.go`：+resolvePluginWindowScreen [S-sig]。
- `backend/screenshot_preview_windows.go`：+readThumbnailPNG [S-sig] +shouldDisplay [S-sig]。
- `backend/screenshot_scroll_windows.go`：+screenshotScrollingRowLooksPureBlack [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3176/4754 = 66.81%。下一批：screenshot 域
screenshotSelectionToolbarWindowService.SetMessages / screenshotScrollingRowLooksDarkSeamArtifact。
