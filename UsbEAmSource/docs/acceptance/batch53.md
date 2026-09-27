# 批次 53 验收（screenshot DXGI 捕获域）

日期：2026-09-20
子批次：dxgi_capture（Screenshot 运行时族第十二子批次）
目标：DXGI 捕获域顶层合成/兜底链 asm 直译，把 `captureVirtualScreenDXGI` /
`captureScreenRectDXGI` 两个 `[P]` 存根转 `[S]`；深层 D3D11/桌面复制留待批次 54+。
新建 `backend/screenshot_dxgi_capture_windows.go`。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet ./backend` | EXIT=0 |
| test | `go test -count=1 -p=1 -run 'TestScreenshotDXGI|TestDefaultScreenshotScreenCaptureBackendDisabled|TestBuildScreenshotDXGI|TestCaptureScreenshotDXGI' ./backend` | EXIT=0（ok changeme/backend 0.399s） |
| gofmt | 手工 edit/write 落笔，未用 `gofmt -w` | 通过 vet 校验 |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1056 / MARKED=867 / S=680 / S-inline=9 / S-sig=106 / P=72 / UNMARKED=189`。

相对批次 52（1043/854/668/9/103/74/189）：+13 标记（10 `[S]` + 3 `[S-sig]`），
`P` −2（`captureVirtualScreenDXGI`/`captureScreenRectDXGI` 两个 `[P]→[S]`），
`UNMARKED` 无新增。深层 3 个 `[S-sig]` 桩（`captureScreenshotDXGIOutputFrameWithOptions` /
`captureScreenshotDXGIOutputFrameRegionWithOptions` / `prewarmScreenshotDXGIOutputFrameCache`）
阻断原因已在文件注释写明。

## G3 逻辑等价

关键实证（asm VA → Go，全部字符串经 rip 算术 `lea_addr+7+disp` 修正后解码）：

- `captureVirtualScreenDXGI`（0x14096ed00）：`qrCodeVirtualScreenBounds()` 空 →
  `未检测到可用的屏幕区域`@0x140c740a7(33B)；`buildScreenshotDXGIVirtualScreenImage(
  displays, bounds, captureScreenshotDXGIDisplayFrameWithFallback)` err 透传；
  `drawScreenshotCaptureCursorOnRGBA(img, bounds.Min.X/Y/Max.X/Max.Y, captureCursor, cursorSnapshot)`；
  `buildDarkenedQRCodeSelectionPreview(img)`；返回 `qrCodeScreenSnapshot{bounds, original, shaded}`（6 字）。
- `captureScreenRectDXGI`（0x14096ef60）：`normalizeScreenshotScrollingRect(rect)` 空 →
  `截图区域为空`@0x140c5a032(18B)；`captureScreenshotDXGIRegionImageWithOptions(
  displays, norm, requireFreshDXGIFrame)` err 透传；叠光标后返回 `(img, nil)`。
- `buildScreenshotDXGIVirtualScreenImage`（0x14096f1a0）：bounds 空 →
  `DXGI 虚拟屏幕区域为空`@0x140c6d732(29B)；否则全参尾调用 `buildScreenshotDXGIDesktopBoundsImage`。
- `buildScreenshotDXGIDesktopBoundsImage`（0x14096f8e0）：bounds 空 →
  `DXGI 输出区域为空`@0x140c62fcc(23B)；callback nil →
  `DXGI 显示器捕获函数未初始化`@0x140c7b4cc(38B)；`image.NewRGBA(Rect(0,0,w,h))`；
  跳过 `!AttachedToDesktop`；`inter=display.Bounds∩bounds` 非空则 callback(display)；
  失败 `DXGI 捕获显示器 %s 失败: %w`@0x140c75e7b(34B)，空图
  `DXGI 捕获显示器 %s 返回空图像`@0x140c7e1f5(39B)；源偏移
  `(inter.Min-display.Min)`、目标偏移 `(inter.Min-bounds.Min)`、
  `w=min(inter.Dx, img2.Dx-offsetX)`、`h=min(inter.Dy, img2.Dy-offsetY)`；
  `draw.DrawMask(..., draw.Src)`；零命中 →
  `DXGI 未捕获到任何已连接显示器`@0x140c80166(41B)。
- `captureScreenshotDXGIRegionImageWithOptions`（0x14096f260）：rect 空 →
  `DXGI 截图区域为空`@0x140c62fb5(23B)；`NewRGBA(Rect(0,0,w,h))`；
  每个显示器 `captureScreenshotDXGIDisplayRegionFrameWithFallback(display, inter, requireFresh)`；
  失败 `DXGI 捕获显示器 %s 区域 %v 失败: %w`@0x140c83209(44B)，空图
  `DXGI 捕获显示器 %s 区域 %v 返回空图像`@0x140c87a2a(49B)；
  目标偏移 `(inter.Min-rect.Min)`、`w=min(inter.Dx, img2.Dx)`、
  `h=min(inter.Dy, img2.Dy)`；`draw.DrawMask(..., draw.Src)`；零命中同上 41B 错误。
- `captureScreenshotDXGIDisplayRegionFrameWithFallback`（0x14096ff00）：
  `inter=rect∩d.Bounds` 空 → `DXGI 显示器区域为空`@0x140c6848d(26B)；SDR →
  debug `DXGI 跳过 SDR 显示器，使用 GDI 单屏区域捕获: display=%s rect=%v`@0x140c91ec2(77B)
  + `captureScreenshotScreenRectGDI(inter)`；HDR 走 `captureScreenshotDXGIOutputFrameRegionWithOptions(
  d, inter.Sub(d.Bounds.Min), requireFresh)`，失败 debug
  `DXGI 捕获显示器 %s 区域 %v 失败，改用 GDI 区域兜底: %v`@0x140c9083f(70B) + GDI；
  GDI 再失败 `DXGI 区域捕获失败: %v；GDI 区域兜底失败: %w`@0x140c8bda8(56B)。
- `captureScreenshotDXGIDisplayFrameWithFallback`（0x14096fea0）：
  `captureScreenshotDXGIDisplayFrameWithFallbackUsing(d, captureScreenshotDXGIOutputFrame, captureScreenshotDXGIDisplayGDI)`。
- `captureScreenshotDXGIDisplayGDI`（目标闭包 `captureScreenshotDXGIDisplayFrameWithFallback.func1`
  0x1409f5d60）：读 `d.Bounds` 四分量后 `captureScreenshotScreenRectGDI`。
- `captureScreenshotDXGIDisplayFrameWithFallbackUsing`（0x140970320）：SDR 且 gdiFallback nil →
  `GDI 单屏兜底函数未初始化`@0x140c75e9d(34B)；SDR debug
  `DXGI 跳过 SDR 显示器，使用 GDI 单屏捕获: display=%s`@0x140c8e7e6(63B) + gdiFallback(d)；
  HDR 且 dxgiCapture nil → `DXGI 显示器捕获函数未初始化`(38B)；dxgiCapture 失败且 gdiFallback nil →
  `DXGI 捕获失败且 GDI 兜底未初始化: %w`@0x140c8601e(47B)；否则 debug
  `DXGI 捕获显示器 %s 失败，改用 GDI 单屏兜底: %v`@0x140c8d7ec(60B) + gdiFallback(d)；
  GDI 再失败 `DXGI 捕获失败: %v；GDI 兜底失败: %w`@0x140c83235(44B)。
- `captureScreenshotDXGIOutputFrame`（0x140975fc0）：尾调用
  `captureScreenshotDXGIOutputFrameWithOptions(d, false)`（`xor eax,eax`）。
- `prewarmScreenshotHDRCaptureAsync`（0x14096e4e0）：`mode=="disabled"` 或
  `!screenshotDXGIHDRCaptureAvailable()` 返回；否则 `go` 预热 worker(mode)。
- `prewarmScreenshotHDRCaptureWorker`（目标闭包 `prewarmScreenshotHDRCaptureAsync.func1`
  0x14096e580）：枚举失败 debug `DXGI HDR 预热枚举显示器失败: %v`@0x140c7f2fa(40B)；
  `selectScreenshotScreenCaptureBackendName(displays, mode, true) != "dxgi-hdr"` 返回；
  否则 `prewarmScreenshotDXGIOutputFrameCache(displays)`。
- 深层 3 个 `[S-sig]` 签名实证：`captureScreenshotDXGIOutputFrameWithOptions(
  screenshotDisplayCaptureInfo, bool) (*image.RGBA, error)`（调用点 0x140975ff5 false）、
  `captureScreenshotDXGIOutputFrameRegionWithOptions(screenshotDisplayCaptureInfo,
  image.Rectangle, bool) (*image.RGBA, error)`（调用点 0x140970020）、
  `prewarmScreenshotDXGIOutputFrameCache([]screenshotDisplayCaptureInfo)`（调用点 0x14096e63a）。

## G4 测试

改写 `backend/screenshot_capture_backend_windows_test.go`（6 用例，全确定性、不触 Windows API）：

- `TestScreenshotDXGIBackendName`：`name()=="dxgi-hdr"`。
- `TestDefaultScreenshotScreenCaptureBackendDisabled`：disabled → GDI。
- `TestBuildScreenshotDXGIVirtualScreenImageEmptyBounds`：空 bounds → error。
- `TestBuildScreenshotDXGIDesktopBoundsNoDisplays`：空 displays → error。
- `TestBuildScreenshotDXGIDesktopBoundsCompose`：display(0,0,10,10) 与
  bounds(2,2,8,8)，callback 返回 10×10 红色图 → 画布 `(0,0)-(6,6)`，
  像素(0,0)=不透明红（验证 `inter.Sub(display.Min)` 源偏移 + `inter.Sub(bounds.Min)` 目标偏移）。
- `TestCaptureScreenshotDXGIRegionImageEmpty`：空 rect / 空 displays 两路 error。
- `TestCaptureScreenshotDXGIDisplayFrameWithFallbackUsingSDRNilFallback`：SDR + nil gdiFallback → error。

全 PASS。

## 判定

四路 PASS，批次 53 闭环。
