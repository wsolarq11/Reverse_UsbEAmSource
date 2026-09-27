# 批次 222 — oledblackout 域两 [P] 升档 [S-sig]

## 基线 / 收口

| 指标 | 基线（批次 221 收口） | 收口（批次 222） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1417 | **1419** |
| P | 79 | **77** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2709 | **2711**（57.02%） |

SHA256 `744749C5F9C6B9914F1AECCF35B83F5B3836D50B136793CE31FC48929A41C761`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/oledblackout_windows.go` 两个 [P] 升档 [S-sig]，均为「签名待实证」类，逐寄存器对齐后
参数分组唯一。

### oledBlackoutEnumTopLevelWindows（0x14091b260）

`(candidates *[]oledBlackoutWindowProcessCandidate)` 原签名正确（单指针参数 + void）。序言仅
test rax 判 nil + morestack 单寄存器保存；EnumWindows 经回调 func1 写入 candidates。

### oledBlackoutVisibleWindowRectForScreenPause（0x14091b560）

`(hwnd uintptr, screen *application.Screen) oledBlackoutRect` →
`(hwnd uintptr, cache map[uintptr]oledBlackoutRect) (oledBlackoutRect, bool)`。

序言 morestack 保存 rax/rbx 两寄存器：rax=hwnd（透传 mapaccess2_fast64 的 key）、rbx=map
（透传 hmap）。返回 5 寄存器（4×int32 rect + esi 找到标志），订正旧存根误标 screen 类型且漏 bool。

## 关键知悉

- oledBlackoutVisibleWindowRectForScreenPause 的 map 是 hwnd→rect 缓存（key=uintptr，value=
  oledBlackoutRect 值类型 16 字节），命中路径走 oledBlackoutDebugLog，未命中路径 GetWindowRect
  后返回 (rect, found)。
- oledBlackoutBrowserMediaContinuityMatchesCandidates 实参为 slice(3)+string(2)（rsi/rdi），
  而 oledBlackoutBrowserAudioMatchesVisiblePlayingWindows 实参含 slice+string+3 额外 qword，
  后者参数分组未唯一确定，保持 [P]。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1419 / P=77 / UNMARKED=0`。
- **G3 行为**：纯签名修正（体零值→体零值），无行为变更；既有测试全量回归 PASS。
- **G4 review**：`oledblackout_windows.go`（2 函数签名修正）；vet/test/build 复验通过。

## 遗留（下一批）

- oledblackout 域剩余 [P]（25 个）：大量 `ctx interface{}` 匿名上下文闭包（约 17 个），
  GSMTC session 结构体（AudioSessionMatchesForeground/Candidate，匿名类型未落地），
  CollectScreen/VisibleMedia/TargetScreenWindowCandidates（匿名上下文字段），
  BrowserAudio/MediaContinuityMatchesCandidates（slice+string+额外 qword）。
- P 已降至 77。下一批优先继续 screenshot_windows（resolveCaptureRectForWindowAtPoint /
  resolveControlHoverForWindow 依赖 screenshotControlBoundsAtPointThroughOverlaySessionWithTimeout
  完整签名，可先攻坚该签名贯通三函数）或转 launcherupdate_runtime（10 个）。
