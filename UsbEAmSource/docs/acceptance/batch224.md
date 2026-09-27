# 批次 224 — oledblackout_hotkey 两 [P] 升档 [S-sig]

## 基线 / 收口

| 指标 | 基线（批次 223 收口） | 收口（批次 224） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1421 | **1423** |
| P | 75 | **73** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2713 | **2715**（57.11%） |

SHA256 `2827DB44A4277896C094BE49EB7AB62292344EB3624BBBF94069DFB02051B6C3`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/oledblackout_hotkey_windows.go` 两个 [P] 升档 [S-sig]，均为「参数/返回形状待实证」类。

### buildOLEDBlackoutHotkeyRegistrations（0x140910140）

`(bindings []oledBlackoutHotkeyBinding) ([]oledBlackoutHotkeyRegistration, map[string]string)`
原签名正确。序言 rax/rbx=bindings(ptr/len，cap 序言后即覆盖未用)；makemap_small 构造 errors，
makeslice 构造 registrations；返回 slice(3)+map(1)=4 字。

### failedOLEDBlackoutHotkeyUpdateResult（0x140910980）

`(registrations []oledBlackoutHotkeyRegistration, err error) oledBlackoutHotkeyUpdateResult`
原签名正确。序言 rax/rbx/rcx=registrations(slice 3 字) + rdi/rsi=err(error 2 字，test rdi 判 nil，
[rdi+0x18] 调 Error())；返回 4 字 = RegisteredProfileIDs(3)+Errors(map 1 字)。

## 关键知悉

- 两函数原签名已正确，阻断仅为「未逐寄存器实证」——序言逐寄存器对齐后参数分组唯一，直接升档。
- applyOLEDBlackoutHotkeyBindings 仍 [P]：首参真实类型 windowsOLEDBlackoutHotkeyManager 未在
  types_oled.go 落地，interface{} 占位为 2 字与 asm 首参 1 字（指针）不匹配，保持阻断。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1423 / P=73 / UNMARKED=0`。
- **G3 行为**：纯签名实证（体零值→体零值），无行为变更；既有测试全量回归 PASS。
- **G4 review**：`oledblackout_hotkey_windows.go`（2 函数升档）。

## 遗留（下一批）

- P 已降至 73。剩余分布：oledblackout_windows 25、screenshot_windows 15、launcherupdate_runtime 9、
  screenshot_uia_windows 8、screenshot_pin 6、screenshot_uia_worker_windows 4、
  filelocator_runtime 2、qrcode_windows 3、oledblackout_hotkey_windows 1。
- 快速突破口：screenshot_windows 的 invalidateWindowChange/invalidateInfoRectChange 已升档，
  同文件 resolveCaptureRectForWindowAtPoint 依赖 screenshotControlBoundsAtPointThroughOverlaySessionWithTimeout
  完整签名；screenshot_uia_worker_windows 的 beginShutdown 返回 3 寄存器（int/error 组合）可攻。
