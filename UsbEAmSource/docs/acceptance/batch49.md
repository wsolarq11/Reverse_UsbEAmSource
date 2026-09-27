# 批次 49 验收（screenshot DXGI 调试格式化链）

日期：2026-09-20
子批次：dxgi_debug（Screenshot 运行时族第八子批次）
目标：DXGI 显示器调试格式化 + HDR 后端决策日志 asm 直译——`screenshotDXGIColorSpaceDebugName` /
`screenshotDisplayCaptureLabel` / `formatScreenshotDisplayCaptureInfoForDebug` /
`logScreenshotHDRCaptureBackendDecision`，新建 `backend/screenshot_dxgi_debug_windows.go`。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| test | `go test -count=1 -p=1 -tags production ./backend` | EXIT=0（ok changeme/backend） |
| gofmt | `gofmt -l` 两个新文件 | 空（无差异） |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1022 / MARKED=833 / S=653 / S-inline=5 / S-sig=103 / P=72 / UNMARKED=189`。

相对批次 48（1018/829/649）：+4 函数全为 `[S]`，`P`/`UNMARKED` 无新增。

## G3 逻辑等价

关键实证（asm VA → Go）：

- `screenshotDXGIColorSpaceDebugName`（0x140970ec0）：`uint32` 参数；0/12/13/14/16/18/19 七个命名分支，其余 `UNKNOWN(%d)`。八个格式字符串全 rodata 解码。
- `screenshotDisplayCaptureLabel`（0x1409706c0）：`info`（112B 结构栈传）；`DeviceName` trim 非空 → 返回；否则 `AdapterName` trim 非空 → `"%s#%d"`(AdapterName, OutputIndex)；否则 `"adapter=%d output=%d"`(AdapterIndex, OutputIndex)。
- `formatScreenshotDisplayCaptureInfoForDebug`（0x140970b60）：15 参数 `fmt.Sprintf`，label 走 `screenshotDisplayCaptureLabel`，颜色空间走 `screenshotDXGIColorSpaceDebugName`；格式串 173B。
- `logScreenshotHDRCaptureBackendDecision`（0x140970820）：开关关闭 → 返回；首循环找附着+边界有效+HDR 的显示器记 `needHDR`；`"DXGI HDR 后端选择: mode=%s backend=%s displays=%d needHDR=%t"`；次循环逐显示器 `"DXGI 显示器: %s"`。
- 字符串实证：格式 173B `label=%q ... luminance[min=%.2f max=%.2f fullFrame=%.2f]`@0x140c95f33；`%s#%d`@0x140c35bbe；`adapter=%d output=%d`@0x140c5dae3；`DXGI HDR 后端选择...`@0x140c8ecd9；`DXGI 显示器: %s`@0x140c5a044；8 个颜色空间名（RGB_FULL_G22_NONE_P709 等）。
- 结构字段偏移对齐：`screenshotDisplayCaptureInfo` 0x64/0x68/0x6c 为 MinLuminance/MaxLuminance/MaxFullFrameLuminance（float32），与 types_screenshot.go 一致。

## G4 测试

新增 `backend/screenshot_dxgi_debug_windows_test.go`：

- `TestScreenshotDXGIColorSpaceDebugName`：8 用例（7 命名 + UNKNOWN）。
- `TestScreenshotDisplayCaptureLabel`：device/adapter/fallback 三分支。
- `TestFormatScreenshotDisplayCaptureInfoForDebug`：label/indices/flags/luminance 子串断言。

全 PASS。

## 判定

四路 PASS，批次 49 闭环。
