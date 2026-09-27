# 批次 235 — qrcode 标注动作链 hwnd 参数订正（1 升档 + 3 订正）

## 基线 / 收口

| 指标 | 基线（批次 234 收口） | 收口（批次 235） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1439 | **1440** |
| P | 57 | **56** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2731 | **2732**（57.46%） |

SHA256 `0706E3CA489F83C4DB140D3C59D9C58BB7DC3EBB68065076D743FD5780B03115`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批内容

`backend/qrcode_windows.go`：`applyAnnotationToolbarAction` 升档 `[S-sig]` 并补 `hwnd` 参数，
连带 `handleAnnotationToolbarActions`/`undoAnnotationAndInvalidate`/`invalidateAnnotationDirtyRect`
三个 `[S-sig]` 存根同步补 `hwnd` 参数。

### hwnd 透传链（证据链）

`qrCodeSelectionWindowProc`（0x140937b00）序言 rax=hwnd、ebx=uMsg、rcx=wParam、rdi=lParam；
经 `activeQRCodeSelectionSession` 取会话后调 `handleMessage(rax=session, rbx=hwnd, ecx=uMsg,
rdi=wParam, rsi=lParam)`。`handleMessage` 的 0x8ea 分支调 `handleAnnotationToolbarActions`
时 rax/rbx 即 session/hwnd 透传；后者序言把 rbx 存 `[rsp+0xb8]` 后原样透传
`applyAnnotationToolbarAction`。故 rbx 参数 = hwnd(uintptr)，非旧注释的"语义未确证"。

### 本批订正清单

| 函数 | 旧签名 | 新签名 |
|---|---|---|
| applyAnnotationToolbarAction | `(action screenshotSelectionToolbarAction)` | `(hwnd uintptr, action screenshotSelectionToolbarAction)` |
| handleAnnotationToolbarActions | `()` | `(hwnd uintptr)` |
| undoAnnotationAndInvalidate | `()` | `(hwnd uintptr)` |
| invalidateAnnotationDirtyRect | `(rect image.Rectangle)` | `(hwnd uintptr, rect image.Rectangle)` |

### applyAnnotationToolbarAction（0x140941180）

`func (s *qrCodeScreenSelectionSession) applyAnnotationToolbarAction(hwnd uintptr, action screenshotSelectionToolbarAction)`。

证据：morestack 保护 2 寄存器（rax=recv、rbx=hwnd）；action 值传走栈——`[rsp+0x148]`=SessionID
与 `recv[0x438]` 比较（不等则返回）、`[rsp+0x150/0x158]`=Action string TrimSpace 后分发
confirm/cancel/undo/tool/color/lineWidth 等分支；hwnd（rbx）test nil 后透传
undoAnnotationAndInvalidate/invalidateAnnotationDirtyRect。无返回值。

### invalidateAnnotationDirtyRect（0x14094bc80）

`func (s *qrCodeScreenSelectionSession) invalidateAnnotationDirtyRect(hwnd uintptr, rect image.Rectangle)`。

证据：序言 `test rbx,rbx`（hwnd nil 判）+ `cmp rcx,rsi`/`cmp rdi,r8`（rect Min/Max 判空）；
rcx/rdi/rsi/r8 = rect(Min.X/Min.Y/Max.X/Max.Y)，rbx=hwnd 存 `[rsp+0x78]` 后透传
invalidateAnnotationTextAreaWith 链。旧注释"rbx/rcx/rdi/rsi = rect"误把 hwnd 计入 rect。

## 关键知悉

- **hwnd 透传**：Windows 窗口过程 hwnd 一路透传进 Go 会话方法，是 qrcode 标注动作链的
  第二个参数；识别关键在 qrCodeSelectionWindowProc 序言 rax=hwnd → handleMessage rbx 透传。
- **action 值传走栈**：screenshotSelectionToolbarAction（0x48B）作为值传结构体整体走调用者
  栈，`[rsp+0x148]` 起即字段基址；与 recv[0x438] 的 SessionID 比较是防陈旧动作的护栏。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`vet ./backend` EXIT=0；
  `test ./backend` `ok changeme/backend`。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1440 / P=56 / UNMARKED=0`。
- **G3 行为**：纯签名订正/升档（体空→体空），无行为变更；既有测试全量 PASS。
- **G4 review**：`qrcode_windows.go`（1 升档 + 3 订正）。

## 遗留（下一批）

- P 已降至 56。剩余分布：oledblackout_windows 25、launcherupdate_runtime 9、screenshot_windows 8、
  screenshot_uia_windows 7、filelocator_runtime 2、其余散落约 5。
- 可快速突破口：filelocator walkRoot/processFile（值传大结构体 prepared + ctx/generation/root/
  results/startedAt，runSearch 调用点已读，待逐参定类型）；screenshot_uia_windows accessibleHitTest
  （3 参数 9 字返回）。
