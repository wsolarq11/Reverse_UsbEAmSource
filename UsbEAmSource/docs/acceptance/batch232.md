# 批次 232 — applyOLEDBlackoutHotkeyBindings 首参订正为 *windowsOLEDBlackoutHotkeyManager

## 基线 / 收口

| 指标 | 基线（批次 231 收口） | 收口（批次 232） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1431 | **1432** |
| P | 65 | **64** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2723 | **2724**（57.29%） |

SHA256 `9270E7DA8BDC5C3EE3238658BFF0C031BBC5D97B09C03FB19A204B137A5A1F2A`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/oledblackout_hotkey_windows.go`：`applyOLEDBlackoutHotkeyBindings` 首参从 interface{} 订正为
`*windowsOLEDBlackoutHotkeyManager`（该类型此前已在 types_misc.go 落地）。该文件 [P] 至此清零。

### applyOLEDBlackoutHotkeyBindings（0x14090fbc0）

`func applyOLEDBlackoutHotkeyBindings(manager interface{}, bindings []oledBlackoutHotkeyBinding)`
→ `func applyOLEDBlackoutHotkeyBindings(manager *windowsOLEDBlackoutHotkeyManager, bindings []oledBlackoutHotkeyBinding)`。

证据链：morestack 保存 rax..rdi 四寄存器 = manager 指针 + bindings slice(ptr/len/cap)；
`[rax]`=callback、`[rax+8]`=commands 解引用字段（与 types_misc.go:257 的
windowsOLEDBlackoutHotkeyManager 前两字段对齐）；返回 void。

## 关键知悉

- 首参「解引用为 2 字结构」其实是 manager 结构体头两字段 callback/commands 被读取，manager 本身
  是 1 字指针，非 interface{}（2 字 itab+data）——旧阻断「类型未落地」解除后即可订正。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1432 / P=64 / UNMARKED=0`。
- **G3 行为**：纯签名订正（体空→体空），无行为变更；既有测试全量回归 PASS。
- **G4 review**：`oledblackout_hotkey_windows.go`（1 函数签名订正）。

## 遗留（下一批）

- P 已降至 64。剩余分布：oledblackout_windows 25、screenshot_windows 8、launcherupdate_runtime 9、
  screenshot_uia_windows 8、screenshot_pin 6、filelocator_runtime 2、qrcode_windows 1。
- 可快速突破口：qrcode 仅剩 applyAnnotationToolbarAction（action 结构体字段展开已基本对齐，
  rbx 语义待解）；captureScreenshotWithLauncherVisibility（cb 返回 6 字类型待逆推）；
  screenshot_uia 三函数 8 寄存器返回。
