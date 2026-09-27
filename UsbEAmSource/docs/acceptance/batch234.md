# 批次 234 — screenshot_pin 域 bounds 签名链订正（6 升档 + 1 订正）

## 基线 / 收口

| 指标 | 基线（批次 233 收口） | 收口（批次 234） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1433 | **1439** |
| P | 63 | **57** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2725 | **2731**（57.45%） |

SHA256 `E777DD37779D676AB74BFBAD3D469C9328DBEDBB1E325B192F3772246663ACC9`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/screenshot_pin.go` + `backend/screenshot_pin_native_windows.go`：screenshot pin 域
一批 `[P]` 存根的签名从汇编实证并升档 `[S-sig]`，核心是 **bounds 参数链** 的打通。

### 关键证据链（bounds 透传，非 name/generation/image）

`createPinnedWindow` 序言保存 7 寄存器（rax=s、rbx=w、rcx/rdi/rsi/r8、r9b=bool），
其 4 个 qword 参数透传 `createScreenshotNativePinWindow`。后者序言把 rcx/rdi 存对象
0x58/0x60（不钳制）、rsi/r8 钳 ≥1 存 0x68/0x70 —— 这是 `application.Rect{X,Y,Width,Height}`
的布局，而非旧注释臆测的 `(pinName,pinGeneration,image)`。据此订正整条链。

### 本批升档 / 订正清单

| 函数 | 旧签名 | 新签名 | VA |
|---|---|---|---|
| storeSnapshotBounds | `(w, application.Rect, persist)` | 同左（本批确认正确） | 0x140985cc0 |
| storeSnapshotLocked | `(w *screenshotPinnedWindow)` | `(view screenshotPinnedWindowView, bounds application.Rect, persist bool)` | 0x1409856e0 |
| createPinnedWindow | `(name string) *screenshotPinnedWindow` | `(w *screenshotPinnedWindow, bounds application.Rect, persist bool) error` | 0x1409876c0 |
| registerPinnedWindow | `(name string, w *screenshotPinnedWindow)` | `(w *screenshotPinnedWindow, bounds application.Rect, persist bool) bool` | 0x140987940 |
| createWebviewPinnedWindow | `(name string) *screenshotPinnedWindow` | `(w *screenshotPinnedWindow, bounds application.Rect, persist bool) error` | 0x140987d00 |
| buildScreenshotPinWindowHTML | `(name, imageURL string) string` | `(name, imageURL string, w, h int, scale float64) string` | 0x14098d680 |
| createScreenshotNativePinWindow | `(service, pin, pinName string, pinGeneration uint64, image *image.RGBA)` | `(service, pin, bounds application.Rect)` | 0x14098d980 |

### storeSnapshotLocked（0x1409856e0）

`func (s *screenshotPinWindowService) storeSnapshotLocked(view screenshotPinnedWindowView, bounds application.Rect, persist bool)`。

证据：rax=s、bl=persist、rcx/rdi/rsi/r8=bounds（4 int，存快照 X/Y/Width/Height）；
view（大结构体）走栈传递——调用者 duffcopy 至 [rsp]，被调者经 [rsp+0x268..0x310] 读
closed/name/imageData/sourcePath/contentKey/scale/opacity/clickThrough/autoShow/order 字段，
与 screenshotPinnedWindowView 字段偏移逐一对齐。persist（bl）序言后未被读取（死参数）。

### buildScreenshotPinWindowHTML（0x14098d680）

`func buildScreenshotPinWindowHTML(name, imageURL string, w, h int, scale float64) string`。

证据：morestack 保护 7 值（rax..r8 + xmm0）；rax/rbx=name(ptr,len)、rcx/rdi=imageURL(ptr,len)、
rsi/r8=w/h(int)、xmm0=scale(float64)；scale 钳 [min,max] 后乘常数转 int 两次入参，
fmt.Sprintf(10 参 HTML 模板, len 0x278d) 返回 string。

## 关键知悉

- **大结构体栈传递**：`screenshotPinnedWindowView`（约 0x2d8 字节）作为参数走调用者栈，
  被调者以 `[rsp+0x268]` 等偏移读字段；偏移映射 = 调用者 [rsp+0x10] 起 = view 基址。
- **application.Rect 布局**：4 int 依次占 rcx/rdi/rsi/r8；X/Y 不钳制、Width/Height 钳 ≥1，
  是 createScreenshotNativePinWindow 对象字段 0x58/0x60/0x68/0x70 的判别依据。
- **bool 死参数**：persist 在 storeSnapshotLocked 序言后被 rbx 覆盖（未读），属透传死参。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`vet ./backend` EXIT=0；
  `test ./backend` `ok changeme/backend`。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1439 / P=57 / UNMARKED=0`。
- **G3 行为**：纯签名订正/升档（体空→体空，体未还原），无行为变更；既有测试全量 PASS。
- **G4 review**：`screenshot_pin.go`、`screenshot_pin_native_windows.go`（6 升档 + 1 订正）。

## 遗留（下一批）

- P 已降至 57。剩余分布：oledblackout_windows 25、launcherupdate_runtime 9、screenshot_windows 8、
  screenshot_uia_windows 7、filelocator_runtime 2、qrcode_windows 1、其余散落约 5。
- 可快速突破口：filelocator walkRoot/processFile（值传大结构体 prepared + 额外 generation/ctx）；
  screenshot_uia_windows 的 accessibleHitTest（3 参数 9 字返回）；qrcode applyAnnotationToolbarAction。
