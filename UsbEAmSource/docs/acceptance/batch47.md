# 批次 47 验收（screenshot GDI 虚拟屏捕获链）

日期：2026-09-20
子批次：capture_virtual（Screenshot 运行时族第六子批次）
目标：GDI 虚拟屏捕获 + 暗化预览 + 后端选择名 asm 直译——`buildDarkenedQRCodeSelectionPreview` /
`captureQRCodeVirtualScreenSnapshotGDIWithCursorSnapshot` / `captureQRCodeVirtualScreenSnapshotGDI` /
`selectScreenshotScreenCaptureBackendName` / `screenshotGDIScreenCaptureBackend.captureVirtualScreen`，
新建 `backend/screenshot_virtual_windows.go`；修正接口签名。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| test | `go test -count=1 -p=1 -tags production ./backend` | EXIT=0（ok changeme/backend） |
| gofmt | `gofmt -l` 三个文件 | 空（无差异） |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1012 / MARKED=823 / S=643 / S-inline=5 / S-sig=103 / P=72 / UNMARKED=189`。

相对批次 46（1007/818/638）：+5 函数全为 `[S]`，`P`/`UNMARKED` 无新增。

## G3 逻辑等价

关键实证（asm VA → Go）：

- `buildDarkenedQRCodeSelectionPreview`（0x14095a200）：nil → `NewRGBA(零矩形)`；否则 `NewRGBA(img.Rect)` + `copy(Pix)`；每像素 RGB 三通道 `*0x2a` 取低 16 位 `*0x147af >> 0x17`（≈42% 亮度）。
- `captureQRCodeVirtualScreenSnapshotGDIWithCursorSnapshot`（0x140971820）：bounds 空 → "未检测到可用的屏幕区域"；`GetDC(0)` 失败 → "获取屏幕设备上下文失败"；CreateCompatibleDC/CreateDIBSection/SelectObject/BitBlt 四段 defer + fmt.Errorf 包装；ok 时 `drawScreenshotCursorSnapshotOnDC`；`buildRGBAFromDIBBits` → 暗化预览；返回 `qrCodeScreenSnapshot` 三字段。
- `captureQRCodeVirtualScreenSnapshotGDI`（0x140971700）：`captureCursor` → 现场捕获光标；否则零值 info + ok=false；返回快照（丢弃 error）。
- `selectScreenshotScreenCaptureBackendName`（0x14096e0a0）：`!preferHDR || "disabled"` → "gdi"；"force" 首个附着且边界有效显示器 → "dxgi-hdr"；"auto" 额外要求 HDR；其余 "gdi"。
- `captureVirtualScreen`（0x14096e660）：`captureCursor && cursorSnapshot.ok` → 快照变体（ok 传 true）；否则 `captureQRCodeVirtualScreenSnapshotGDI(captureCursor)`。
- 错误字符串 6 个全 rodata 解码：`0x140c740a7`(33B "未检测到可用的屏幕区域") / `0x140c740e9`(33B "获取屏幕设备上下文失败") / `0x140c7a005`(37B "创建截图设备上下文失败: %w") / `0x140c6bf08`(28B "创建截图位图失败: %w") / `0x140c6bf24`(28B "选择截图位图失败: %w") / `0x140c6bf40`(28B "复制屏幕像素失败: %w")。
- `qrCodeBitBlt` rop=0x40CC0020（SRCCOPY|CAPTUREBLT），与批次 44 一致。
- 暗化定点公式实证：`darken(255) = uint8((uint16(255*42) * 83887) >> 23) = 107`。

## G4 测试

新增 `backend/screenshot_virtual_windows_test.go`：

- `TestBuildDarkenedQRCodeSelectionPreview`：nil 空矩形；RGB 三通道暗化 + alpha 保留；`darken(255)=107`。
- `TestSelectScreenshotScreenCaptureBackendName`：7 用例（no prefer/disabled/force 附着/force 未附着/auto HDR/auto SDR/未知模式）。

全 PASS。

## 接口修正与格式债务

- `types_screenshot.go` 接口 `captureVirtualScreen` 去掉 `error`（asm 0x14096e660 两分支均只返回 6 字 snapshot，error 寄存器不参与；调用方 0x14096e280 也只取 6 字）。captureScreenRect 仍保留 error。
- `types_screenshot.go` 历史格式债务修复：CRLF→LF + `package main` 后空行 + struct 字段对齐（66 行 gofmt diff，纯格式无逻辑变更）。

## 判定

四路 PASS，批次 47 闭环。
