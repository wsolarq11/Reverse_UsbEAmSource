# 批次 52 验收（screenshot 后端收口链）

日期：2026-09-20
子批次：capture_backend（Screenshot 运行时族第十一子批次）
目标：默认截图后端选择 + 按 options 捕获入口 + DXGI 后端接口方法 asm 直译——
`defaultScreenshotScreenCaptureBackend` / `captureScreenshotVirtualScreenSnapshotWithOptions` /
`captureScreenshotScreenRectWithOptions` / `screenshotDXGIHDRScreenCaptureBackend.name` /
`captureVirtualScreen` / `captureScreenRect`，新建
`backend/screenshot_capture_backend_windows.go`；`captureVirtualScreenDXGI` /
`captureScreenRectDXGI` 为 `[P]` 存根（阻断：依赖 DXGI 捕获域未落地）。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| test | `go test -count=1 -p=1 -tags production ./backend` | EXIT=0（ok changeme/backend） |
| gofmt | `gofmt -l` 两个新文件 | 空（无差异） |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1043 / MARKED=854 / S=668 / S-inline=9 / S-sig=103 / P=74 / UNMARKED=189`。

相对批次 51（1035/846/662/9/103/72/189）：+8 标记（6 [S] + 2 [P]），`UNMARKED` 无新增；
`P` +2 为 DXGI 捕获核心入口存根（`captureVirtualScreenDXGI`/`captureScreenRectDXGI`，
阻断原因已注释，留待 DXGI 捕获域批次转 [S]）。

## G3 逻辑等价

关键实证（asm VA → Go）：

- `defaultScreenshotScreenCaptureBackend`（0x14096de60）：`resolveScreenshotHDRCaptureMode()=="disabled"`
  → GDI；`!screenshotDXGIHDRCaptureAvailable()` → GDI；
  `enumerateScreenshotDisplayCaptureInfos()` err != nil → debug log
  （`枚举 DXGI 显示器失败，使用 GDI 后端: %v`）+ GDI；
  `selectScreenshotScreenCaptureBackendName(displays, mode, true)`；
  `logScreenshotHDRCaptureBackendDecision(mode, name, displays)`；
  `name != "dxgi-hdr"` → GDI；否则 debug log
  （`启用 DXGI HDR 截图后端: mode=%s displays=%d`）+ `&screenshotDXGIHDRScreenCaptureBackend{displays: displays}`。
- `captureScreenshotVirtualScreenSnapshotWithOptions`（0x14096e280）：
  `defaultScreenshotScreenCaptureBackend().captureVirtualScreen(options)`（itab.fun[1]@0x20）。
- `captureScreenshotScreenRectWithOptions`（0x14096e420）：
  `defaultScreenshotScreenCaptureBackend().captureScreenRect(rect, options)`（itab.fun[0]@0x18）。
- `screenshotDXGIHDRScreenCaptureBackend.name`：返回 `"dxgi-hdr"`（8 字节，
  与 `selectScreenshotScreenCaptureBackendName` 0x14096e0a0 命中分支 rodata 一致）。
- `captureVirtualScreen`（0x14096e920）：`captureVirtualScreenDXGI` err != nil →
  debug log（`DXGI 虚拟屏幕捕获失败，回退 GDI: %v`）+ `screenshotGDIScreenCaptureBackend{}.captureVirtualScreen(options)`；
  否则返回 DXGI 快照（返回 6 字，无 error）。
- `captureScreenRect`（0x14096eb40）：`captureScreenRectDXGI` err != nil →
  debug log（`DXGI 区域捕获失败，回退 GDI: rect=%v err=%v`，args 为 `rect`/`err` 两个 any）+ GDI 回退；
  否则返回 `(img, nil)`。
- 值接收者确认：`screenshotDXGIHDRScreenCaptureBackend` 为 24 字节 slice header，
  morestack 保存 3 寄存器（rax/rbx/rcx = displays ptr/len/cap），与值接收者一致；
  接口 itab 的 pointer wrapper（`(*screenshotDXGIHDRScreenCaptureBackend)`）由 Go 自动生成。
- 字符串实证：`枚举 DXGI 显示器失败，使用 GDI 后端: %v`@0x140c88f4d(51B)；
  `启用 DXGI HDR 截图后端: mode=%s displays=%d`@0x140c879f9(49B)；
  `DXGI 虚拟屏幕捕获失败，回退 GDI: %v`@0x140c8512a(46B)；
  `DXGI 区域捕获失败，回退 GDI: rect=%v err=%v`@0x140c899e6(52B)；
  `"disabled"` 8 字节常量 `0x64656c6261736964`；`"dxgi-hdr"` 8 字节常量 `0x7264682d69677864`。

## G4 测试

新增 `backend/screenshot_capture_backend_windows_test.go`（3 用例）：

- `TestScreenshotDXGIBackendName`：`name()=="dxgi-hdr"`。
- `TestDefaultScreenshotScreenCaptureBackendDisabled`：`USBEAM_SCREENSHOT_HDR_CAPTURE=disabled`
  → 后端 `name()=="gdi"`（disabled 分支在 DXGI 探测前短路，确定性）。
- `TestScreenshotDXGIBackendCaptureCorePending`：两个 `[P]` 核心入口返回非 nil error。

全 PASS。

## 判定

四路 PASS，批次 52 闭环。
