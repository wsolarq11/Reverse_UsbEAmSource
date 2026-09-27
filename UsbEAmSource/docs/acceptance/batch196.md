# 批次 196 — screenshot UIA 纯几何辅助 4 函数 [S]

## 基线 / 收口

| 指标 | 基线（批次 195 收口） | 收口（批次 196） |
|---|---|---|
| FUNCS | 2762 | **2762** |
| S | 1211 | **1215** |
| S-inline | 36 | 36 |
| S-sig | 1400 | **1397** |
| P | 115 | **114** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2647 | 2648（55.7%） |

SHA256 `9DC98F73905D4AD5F8D512D4DC79C8979B5E786BA64D6F0CED11350CFE354DF6`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,279,744 B，`bash build.sh` 重建）。

## 本批内容

`backend/screenshot_uia_windows.go` 4 个纯几何函数 [S]（现场 dump asm 逐条对齐）：

### `screenshotRectMatchesBounds`（0x1409af320，224B）— [S-sig]→[S]

签名 `(rect, bounds image.Rectangle) bool`。语义：rect 空或 bounds 空 → false；rect==bounds →
true；rect 包含 bounds（`Min.X<=` 且 `Max.X>=`，Y 同理）→ true；否则四边 |差值|<=2 → true，
任一边 >2 → false。

### `screenshotUIAElementHitContainsPoint`（0x1409b2c60，128B）— [P]→[S]

**签名修正为 8 槽** `(rect image.Rectangle, controlType uint32, ok bool, x, y int) bool`。旧 [P]
存根 `(rect, x, y int)` 少 2 槽；本批经调用点 0x1409b2aaa/0x1409b2ab1（`mov esi,[rsp+0x80]` /
`movzx r8d,[rsp+0x84]`）与 `readScreenshotUIAElementHitInfo` 返回 `(rect,uint32,bool,error)`
交叉确证 esi=controlType(uint32)、r8b=ok(bool)。

体（asm 逐分支）：`rect.Dx()<2 || rect.Dy()<2` → false；否则
`rect.Min.X<=x && x<rect.Max.X && rect.Min.Y<=y && y<rect.Max.Y`（controlType/ok 参数函数体未用）。

### `screenshotPreferSmallerControlRect`（0x1409aedc0，96B）— [S-sig]→[S]

签名 `(a, b image.Rectangle) bool`。语义：a 空或 b 空 → false；面积（Dx*Dy）<=0 → false；
返回 `bArea < aArea`（偏好面积更小的 b）。asm 中 imul 后 `test; jle` 为面积溢出防护。

### `screenshotPointNearWindowEdge`（0x1409af2a0，128B）— [S-sig]→[S]

签名 `(x, y int, rect image.Rectangle, threshold int) bool`。语义：rect 空或点不在 rect 内 →
false；threshold<0 钳 0（`test r9,r9; cmovl`）；返回点到 rect 四边最近距离 <= threshold
（水平 min(x-Min.X, Max.X-x) 与垂直 min(y-Min.Y, Max.Y-y) 取较小，再与 threshold 比较）。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；
  `go test ./backend` `ok changeme/backend`（cached）。
- **G2 count_funcs**：`FUNCS=2762 / S=1215 / S-inline=36 / S-sig=1397 / P=114 / UNMARKED=0`。
- **G3 行为**：4 函数均为纯几何叶函数，无 IO/COM 依赖，行为由 asm 直译；全量 `go test ./backend` PASS。
- **G4 review**：仅编辑 screenshot_uia_windows.go 一处文件；签名（含 8 槽修正）由调用点实参
  寄存器交叉实证，未臆造。

## 遗留（下一批）

- §10 差集仍 57（screenshot_uia_windows.go 已存在，差集不变）。
- 同文件相邻纯几何 `normalizeScreenshotControlRect`（0x1409af400）、
  `screenshotUIAControlRectUsable`（0x1409af4e0）留待后续。
- 其余遗留：`matchNodeNameTerms`（0x1407e3ac0）、`pluginupdate.go`、`pluginwindow.go`、
  `TestReminderNotification` 0x1407a7320、`desktopCalendarStringList` 0x1407aa820。
