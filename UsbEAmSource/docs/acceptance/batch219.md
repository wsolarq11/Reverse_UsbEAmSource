# 批次 219 — qrcode 框选会话八 [P] 升档 [S-sig]（session 指针类型贯通）

## 基线 / 收口

| 指标 | 基线（批次 218 收口） | 收口（批次 219） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1403 | **1411** |
| P | 93 | **85** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2695 | **2703**（56.86%） |

SHA256 `5415D6667011C7F2A45B5EFDE1478B6518B1974463F82E946E40DFA21F29D74A`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

突破点：qrcode 域多个 [P] 被「多出的 rbx 参数类型未确证」卡住。溯源
`screenshotControlBoundsAtPointThroughOverlaySessionWithTimeout`（`screenshot_uia_windows.go`）的签名
`sess *screenshotWindowSelectionSession`，再沿透传链 rbx→该 session 参数贯通全链，类型唯一确定。
`backend/qrcode_windows.go` 八个 [P] 升档 [S-sig]。

### 升档

- `resolveControlHoverForWindow`（0x14093dca0）：`(hwnd,p)` → `(sess, hwnd, p)`。
- `resolveControlSelectionAtPoint`（0x14093e320）：`(p)` → `(sess, p)`。
- `resolveControlSelectionForWindowAtPoint`（0x14093e3c0）：`(hwnd,p)` → `(sess, hwnd, p, force bool)`。
- `resolveControlClickSelectionAtPoint`（0x14093e6e0）：`(hwnd,p)` → `(sess, p, controlRect)`。
- `prepareSelectionResult`（0x14093f700）：`(hint uintptr)` → `(sess)`；返回 void（无返回寄存器）。
- `resolveClickLikeControlSelection`（0x14093f880）：`(hint uintptr)` → `(sess)`。
- `drawAnnotationOverlay`（0x140948880）：`(hdc,rect)` → `(hdc, rect, bits unsafe.Pointer)`；
  rbx 实为 hdc（透传 drawQRCodeAnnotationStroke），r9 实为 bits（透传 drawAnnotationStrokesOnPaintBuffer）。
- `applyRoundedSelectionPreview`（0x140949200）：`(img *image.RGBA, rect)` →
  `(pBits unsafe.Pointer, selRect, previewRect image.Rectangle)`；旧存根误标 img 类型且漏一 rect。

## 关键知悉

- `sess *screenshotWindowSelectionSession` 是 qrcode 框选会话与 screenshot 域共用的会话句柄，
  沿 `resolveControl* → screenshotControlBoundsAtPointThroughOverlaySessionWithTimeout` 链透传。
- 大结构参数（image.Rectangle）当剩余寄存器不足 4 字时整体走栈（`applyRoundedSelectionPreview` 的
  previewRect 走 `[rsp+0x110..0x128]`），与 Go ABI「struct 不能完整入剩余寄存器则整体走栈」一致。
- void 函数尾声或 `xor eax`（finishCornerRadiusDrag）、或裸 `ret`（prepareSelectionResult），
  不能据此判别 bool/void——本批对 void 判定一律依据「全部返回路径无返回寄存器设置」。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1411 / P=85 / UNMARKED=0`。
- **G3 行为**：纯签名修正（体零值→体零值），无行为变更；既有测试全量回归 PASS。
- **G4 review**：`qrcode_windows.go`（8 函数签名修正）；vet/test/build 复验通过。

## 遗留（下一批）

- qrcode 域剩余 [P]：updateCornerRadiusDrag（返回单寄存器零值，void/bool 无法判别）、
  confirmAnnotationEditing（同上）、applyAnnotationToolbarAction（action 结构体首字段 + rbx 语义未确证）。
- 其余域 P 分布：oledblackout_windows 27、screenshot_windows 15、launcherupdate_runtime 10、
  screenshot_uia_windows 10、screenshot_pin 7、screenshot_uia_worker_windows 4、
  oledblackout_hotkey_windows 3、filelocator_runtime 2、+ 4 文件各 1。
- §10 差集 54 文件。P 已降至 85。下一批优先 screenshot_uia_windows 的 session 透传链
  （与 qrcode 同构，可复用本批「session 指针类型贯通」手法）。
