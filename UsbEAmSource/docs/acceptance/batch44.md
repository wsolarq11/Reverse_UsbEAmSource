# 批次 44 验收（screenshot_capture_windows GDI 捕获核心）

日期：2026-09-20
子批次：capture_windows（Screenshot 运行时族第三子批次）
目标：GDI 屏幕矩形捕获链核心 asm 直译——`qrCodeBitBlt` /
`qrCodeVirtualScreenBounds` / `normalizeScreenshotScrollingRect` /
`captureScreenshotScreenRectGDI` / `getSystemMetrics` /
`screenshotGDIScreenCaptureBackend.name`，新建 `backend/screenshot_capture_windows.go`。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build ./backend` | EXIT=0 |
| vet | `go vet ./backend` | EXIT=0 |
| gofmt | `gofmt -l backend/screenshot_capture_windows.go` | 空（无差异） |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=991 / MARKED=802 / S=622 / S-inline=5 / S-sig=103 / P=72 / UNMARKED=189`。

相对批次 43（985/796/616）：+6 函数全为 `[S]`，`P`/`UNMARKED` 无新增。

## G3 逻辑等价

关键实证（asm VA → Go）：

- `qrCodeBitBlt`（0x14095ebe0）：`dstW<=0 || dstH<=0` → nil；`procBitBlt.Call(9 args)`；`r1!=0` → nil；`lastErr==nil || errors.Is(lastErr, syscall.Errno(0))` → `"BitBlt 失败"`；否则透传 `lastErr`。sentinel itab kind=12（Uintptr，size=8），data 值 0（.rdata 0x1411cd520）。
- `qrCodeVirtualScreenBounds`（0x14095a060）：同一 GetSystemMetrics proc（0x141bc2928）六次调用；SM_X/Y/CX/CYVIRTUALSCREEN（0x4c/0x4d/0x4e/0x4f）；宽或高 `<=0` 回退到 SM_XSCREEN/SM_YSCREEN 且宽高置 0；最终对 `(x,y,x+w,y+h)` 做 min/max 规范化（cmovg 两轴）。
- `normalizeScreenshotScrollingRect`（0x1409a2c60）：先按轴 cmovg 排序 Min/Max；`Min.X>=Max.X || Max.Y<=Min.Y` → 零矩形；否则 `rect.Intersect(qrCodeVirtualScreenBounds())`。
- `captureScreenshotScreenRectGDI`（0x1409722e0）：`normalize` → 空矩形报 `"截图区域为空"`(18B) → `GetDC(0)` 失败报 `"获取屏幕设备上下文失败"`(33B，无 %w) → `defer ReleaseDC(0,hdc)` → `createAppCompatibleDC` 失败报 `"创建截图设备上下文失败: %w"`(37B) → `defer deleteAppDC` → `createAppIconCanvas(width,height)` 失败报 `"创建长截图位图失败: %w"`(31B) → `defer deleteAppObject` → `selectAppObject` 失败报 `"选择长截图位图失败: %w"`(31B) → `defer restoreAppObject` → `qrCodeBitBlt(memDC,0,0,w,h,hdc,Min.X,Min.Y,0x40CC0020)` 失败报 `"复制长截图像素失败: %w"`(31B) → `buildRGBAFromDIBBits`。
- 错误字符串 6 个全 rodata 解码：`0x140c5a032`(18B) / `0x140c740e9`(33B) / `0x140c7a005`(37B) / `0x140c711d8`(31B) / `0x140c711f7`(31B) / `0x140c71216`(31B)，`0x140c4e145`(13B "BitBlt 失败")。
- `screenshotGDIScreenCaptureBackend.name`：`selectScreenshotScreenCaptureBackendName`（0x14096e0a0）"disabled" 分支返回 rodata `"gdi"`（0x140c33d31，3B）。

## G4 测试

本批次为纯 GDI 包装与几何规范化，无独立可观测纯逻辑单测（调用真实 Win32 API 需交互式桌面）；编译 + vet + gofmt 三门禁即验收。`qrCodeVirtualScreenBounds`/`normalizeScreenshotScrollingRect` 的纯几何分支已通过代码审查对齐 asm。

## 判定

四路 PASS，批次 44 闭环。
