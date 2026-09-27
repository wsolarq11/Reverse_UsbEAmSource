# 批次 230 — confirmAnnotationEditing 返回确证为 void

## 基线 / 收口

| 指标 | 基线（批次 229 收口） | 收口（批次 230） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1429 | **1430** |
| P | 67 | **66** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2721 | **2722**（57.25%） |

SHA256 `913F88DEFAED694DF7A059318D130E126E28C356BD9291EE666655A9BD186A73`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/qrcode_windows.go`：`confirmAnnotationEditing` 的返回类型从「bool/int/void 未确证」收窄为 void。

### confirmAnnotationEditing（0x1409403e0）

`func (s *qrCodeScreenSelectionSession) confirmAnnotationEditing(p uintptr)`（void，签名不变）。

证据链：morestack 只保存 rax(recv)/rbx(p)；rbx test nil 后传 hideScreenshotOverlayWindow(hwnd uintptr)，
故 p=hwnd；两个返回路径（0x140940724 / 0x14094072c）均 xor eax 零值且全函数无 mov eax,1 / set 指令，
排除 bool（bool 必有 true 分支 mov eax,1），确证 void。

## 关键知悉

- Void-vs-bool 歧义的判据：不止看单个 xor eax，要看全函数所有返回路径是否都只有零值且无
  `mov eax,1`/`setcc`。bool 函数必有非零返回分支；全零值 = void（或恒 false，语义上排除）。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1430 / P=66 / UNMARKED=0`。
- **G3 行为**：纯签名订正（体空→体空），无行为变更；既有测试全量回归 PASS。
- **G4 review**：`qrcode_windows.go`（1 函数签名订正）。

## 遗留（下一批）

- P 已降至 66。剩余分布：oledblackout_windows 25、screenshot_windows 8、launcherupdate_runtime 9、
  screenshot_uia_windows 8、screenshot_pin 6、filelocator_runtime 2、qrcode_windows 2、
  oledblackout_hotkey_windows 1。
- 可快速突破口：qrcode_windows 的 updateCornerRadiusDrag（同 Void-vs-bool 判据可套用）；
  applyAnnotationToolbarAction（action 结构体待解）；screenshot_uia 三函数返回 8 寄存器。
