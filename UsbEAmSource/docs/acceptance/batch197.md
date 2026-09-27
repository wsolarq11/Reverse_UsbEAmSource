# 批次 197 — screenshot UIA 控件矩形规范化/可用性判断 2 函数 [S]

## 基线 / 收口

| 指标 | 基线（批次 196 收口） | 收口（批次 197） |
|---|---|---|
| FUNCS | 2762 | **2762** |
| S | 1215 | **1217** |
| S-inline | 36 | 36 |
| S-sig | 1397 | **1395** |
| P | 114 | 114 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2648 | 2648（55.7%） |

SHA256 `04005C3C5544C7E298CC4071DB1B129C3CA9DD080229E5E3FCF4112D20A4C525`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,279,744 B，`bash build.sh` 重建）。

## 本批内容

`backend/screenshot_uia_windows.go` 2 个函数 [S-sig]→[S]（现场 dump asm 逐条对齐）：

### `normalizeScreenshotControlRect`（0x1409af400，224B）— [S-sig]→[S]

签名 `(rect, bounds image.Rectangle) (image.Rectangle, bool)`。语义：rect 空或 bounds 空 →
`(image.Rectangle{}, false)`；否则 `r := rect.Intersect(bounds)`，`r.Dx()<2 || r.Dy()<2` →
`(image.Rectangle{}, false)`；否则 `(r, true)`。

### `screenshotUIAControlRectUsable`（0x1409af4e0，288B）— [S-sig]→[S]

签名 `(rect image.Rectangle, controlType int, bounds image.Rectangle) bool`。语义：

- `r, ok := normalizeScreenshotControlRect(rect, bounds)`；`!ok` → false。
- `controlType` 不在 `[50008, 50033]`（`sub rdx, 0xc358; cmp rdx, 0x19; ja`）→ true。
- 容器类 role（switch 跳转表 0x1411e1ea0 26 项逐项解码到 CASE_MATCH 分支）：
  `50008 List / 50009 Menu / 50010 MenuBar / 50017 StatusBar / 50018 Tab / 50021 ToolBar /
  50023 Tree / 50026 Group / 50028 DataGrid / 50030 Document / 50032 Window / 50033 Pane`
  → 返回 `!screenshotRectMatchesBounds(r, bounds)`（交集占满窗口边界则不采用）。
- 其余 role（MenuItem/ProgressBar/RadioButton/ScrollBar/Slider/Spinner/TabItem/Text/ToolTip/
  TreeItem/Custom/Thumb/DataItem/SplitButton）→ true。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；
  `go test ./backend` `ok changeme/backend`（cached）。
- **G2 count_funcs**：`FUNCS=2762 / S=1217 / S-inline=36 / S-sig=1395 / P=114 / UNMARKED=0`。
- **G3 行为**：2 函数为纯几何/数值分派，无 IO/COM 依赖；switch 跳转表经 va_read 原始字节
  26 项逐项解码（CASE_MATCH=0x1409af545 / RET_TRUE=0x1409af56c），未臆造分派。
- **G4 review**：仅编辑 screenshot_uia_windows.go 一处文件；复用已落地
  screenshotRectMatchesBounds + normalizeScreenshotControlRect；无跨文件写重叠。

## 遗留（下一批）

- §10 差集仍 57（screenshot_uia_windows.go 已存在，差集不变）。
- 同文件剩余 [S-sig]/[P] 多为 COM/IUIAutomation vtable 调用（createScreenshotUIAutomationOnCOMThread、
  readScreenshotUIAElementHitInfo、screenshotUIASelectableAncestorHitInfoAtPoint 等），需还原
  COM 包装后成批落地。
- 其余遗留：`matchNodeNameTerms`（0x1407e3ac0）、`pluginupdate.go`、`pluginwindow.go`、
  `TestReminderNotification` 0x1407a7320、`desktopCalendarStringList` 0x1407aa820。
