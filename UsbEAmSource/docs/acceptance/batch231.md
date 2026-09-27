# 批次 231 — updateCornerRadiusDrag 返回确证 void

## 基线 / 收口

| 指标 | 基线（批次 230 收口） | 收口（批次 231） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1430 | **1431** |
| P | 66 | **65** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2722 | **2723**（57.27%） |

SHA256 `9094680F5D4C48BBF3CBCD1C877207AEB6B124E31629DCBE4A3056CCE856C820`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/qrcode_windows.go`：`updateCornerRadiusDrag` 返回类型收窄为 void。

### updateCornerRadiusDrag（0x14093bcc0）

`func (s *qrCodeScreenSelectionSession) updateCornerRadiusDrag(hwnd uintptr, p image.Point)`（void，签名不变）。

证据链：morestack 保存 rax(recv)/rbx/rcx/rdi；rbx=hwnd、rcx/rdi=point(X/Y)；switch 跳转表按状态
选 WM 消息 ID（0x7f82..0x7f89）送 LazyProc.Call 发送窗口消息；早期返回（0x14093bf30）与正常返回
均 xor eax 零值、无 mov eax,1/setcc，确证 void（switch 里 eax=0x7f82 等是消息 ID，非 bool 返回）。

## 关键知悉

- 复用批次 230 的 Void-vs-bool 判据：全函数返回路径都零值且无 mov eax,1/setcc = void。
- switch 跳转表里 `mov eax, 0x7f82` 这类高位立即数要辨明是「WM 消息 ID 送 SendMessage」而非
  「bool 真值返回」；0x7f82 恰在 WM_APP 区间，属自定义消息。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1431 / P=65 / UNMARKED=0`。
- **G3 行为**：纯签名订正（体空→体空），无行为变更；既有测试全量回归 PASS。
- **G4 review**：`qrcode_windows.go`（1 函数签名订正）。

## 遗留（下一批）

- P 已降至 65。剩余分布：oledblackout_windows 25、screenshot_windows 8、launcherupdate_runtime 9、
  screenshot_uia_windows 8、screenshot_pin 6、filelocator_runtime 2、qrcode_windows 1、
  oledblackout_hotkey_windows 1。
- 可快速突破口：qrcode 仅剩 applyAnnotationToolbarAction（action 结构体待解）；
  captureScreenshotWithLauncherVisibility（cb 返回 6 字类型待逆推）；screenshot_uia 三函数 8 寄存器返回。
