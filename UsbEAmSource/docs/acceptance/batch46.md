# 批次 46 验收（screenshot_cursor_draw_windows 光标绘制收口链）

日期：2026-09-20
子批次：cursor_draw（Screenshot 运行时族第五子批次）
目标：光标绘制收口链 + GDI 矩形捕获方法 asm 直译——`screenshotSystemCursorSize` /
`screenshotCursorBaseSizeFromRegistry` / `screenshotCursorDrawSize` /
`screenshotCursorIconMetrics` / `forceScreenshotDIBAlphaOpaqueInRect` /
`drawScreenshotCursorInfoOnDC` / `drawScreenshotCursorSnapshotOnDC` /
`drawScreenshotCursorSnapshotOnRGBA` / `drawScreenshotCursorOnRGBA` /
`drawScreenshotCaptureCursorOnRGBA` / `screenshotGDIScreenCaptureBackend.captureScreenRect`，
新建 `backend/screenshot_cursor_draw_windows.go`。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| test | `go test -count=1 -p=1 -tags production ./backend` | EXIT=0（ok changeme/backend） |
| gofmt | `gofmt -l` 两个新文件 | 空（无差异） |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1007 / MARKED=818 / S=638 / S-inline=5 / S-sig=103 / P=72 / UNMARKED=189`。

相对批次 45（996/807/627）：+11 函数全为 `[S]`，`P`/`UNMARKED` 无新增。

## G3 逻辑等价

关键实证（asm VA → Go）：

- `screenshotSystemCursorSize`（0x140974e00）：`GetSystemMetrics(0xd)/GetSystemMetrics(0xe)`。
- `screenshotCursorBaseSizeFromRegistry`（0x140974ca0）：`OpenKey(HKCU, "Control Panel\Cursors", QUERY_VALUE)` → `GetIntegerValue("CursorBaseSize")`；错误/0/大于 0x200 返回 0。
- `screenshotCursorDrawSize`（0x140974b80）：max(基准,w,h,系统宽,系统高)；`w<=0||h<=0||w==h` → `(max,max)`；等比缩放并钳制 ≥1。
- `screenshotCursorIconMetrics`（0x140974ac0）：`loadAppIconInfo` → `screenshotCursorDrawSize`；失败释放图标并返回零值 + false。
- `forceScreenshotDIBAlphaOpaqueInRect`（0x140974ea0）：`Intersect((x0,y0,x1,y1),(0,0,width,height))` 空 → false；逐像素 alpha 置 0xFF。
- `drawScreenshotCursorInfoOnDC`（0x140974080）：前置 `ok && Cursor!=0 && Flags&1` → 热点等比缩放 → 目标坐标 = 光标屏幕位 - 起点 - 缩放热点 → `DrawIconEx(hdc,目标X,目标Y,hIcon,drawW,drawH,0,0,3)` → `forceScreenshotDIBAlphaOpaqueInRect`。
- `drawScreenshotCursorSnapshotOnDC`（0x140973fc0）：hdc/bits 零或空 rect → false；否则透传。
- `drawScreenshotCursorSnapshotOnRGBA`（0x140974560）：尺寸校验 → `createQRCodeRGBACompatibleBitmap` → 三处 defer（DeleteObject/DeleteDC/restore）→ `drawScreenshotCursorSnapshotOnDC` → `GdiFlush` → `copyQRCodeDIBBitsToRGBA`。
- `drawScreenshotCursorOnRGBA`（0x1409744a0）：现场捕获 → snapshot 变体。
- `drawScreenshotCaptureCursorOnRGBA`（0x14096f100）：`!captureCursor` → false；`!cursor.ok` → 现场捕获；否则快照。
- `captureScreenRect`（0x14096e840）：`captureScreenshotScreenRectGDI` → 错误透传 → `normalizeScreenshotScrollingRect` → 光标叠加（返回值忽略）→ `(img,nil)`。
- 注册表字符串 2 个全 rodata 解码：`0x140c5f97d`(21B "Control Panel\Cursors")/`0x140c507f2`(14B "CursorBaseSize")。
- `DrawIconEx` diFlags=3（DI_MASK|DI_IMAGE），`GdiFlush` 新增 `procGdiFlush`。

## G4 测试

新增 `backend/screenshot_cursor_draw_windows_test.go`：

- `TestForceScreenshotDIBAlphaOpaqueInRect`：交集内 alpha 置 0xFF，其余保持 0。
- `TestForceScreenshotDIBAlphaOpaqueInRectRejects`：nil/非正尺寸/位图外矩形拒绝。
- `TestDrawScreenshotCaptureCursorDisabled`：`captureCursor=false` 返回 false。
- `TestDrawScreenshotCursorSnapshotSizeMismatch`：尺寸不匹配与 nil 图返回 false。

四项全 PASS。

## 判定

四路 PASS，批次 46 闭环。
