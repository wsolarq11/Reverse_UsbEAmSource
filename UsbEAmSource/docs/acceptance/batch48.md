# 批次 48 验收（screenshot DXGI/HDR 模式解析与入口探测）

日期：2026-09-20
子批次：dxgi_hdr_parse（Screenshot 运行时族第七子批次）
目标：HDR 捕获模式解析 + 调试开关 + DXGI/D3D11 入口可用性 asm 直译——`parseScreenshotBool` /
`parseScreenshotHDRCaptureMode` / `resolveScreenshotHDRCaptureMode` / `isScreenshotHDRCaptureDebugEnabled` /
`screenshotHDRCaptureDebugLog` / `screenshotDXGIHDRCaptureAvailable`，新建
`backend/screenshot_hdr_windows.go`。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| test | `go test -count=1 -p=1 -tags production ./backend` | EXIT=0（ok changeme/backend） |
| gofmt | `gofmt -l` 两个新文件 | 空（无差异） |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1018 / MARKED=829 / S=649 / S-inline=5 / S-sig=103 / P=72 / UNMARKED=189`。

相对批次 47（1012/823/643）：+6 函数全为 `[S]`，`P`/`UNMARKED` 无新增。

## G3 逻辑等价

关键实证（asm VA → Go）：

- `parseScreenshotBool`（0x140971620）：`TrimSpace`+`ToLower`；`"1"/"on"/"yes"/"true"` → (true,true)；`"0"/"no"/"off"/"false"` → (false,true)；其余 (false,false)。
- `parseScreenshotHDRCaptureMode`（0x140971280）：按 len 1-8 跳表；`"0"/"no"/"off"/"false"/"disable"/"disabled"` → "disabled"；`"1"/"on"/"hdr"/"yes"/"auto"/"true"` → "auto"；`"dxgi"/"force"/"always"` → "force"；默认 trim 空 → "auto"，否则 "disabled"。
- `resolveScreenshotHDRCaptureMode`（0x140971240）：`os.Getenv("USBEAM_SCREENSHOT_HDR_CAPTURE")`（29B，`0x140c6d789`）→ parse。
- `isScreenshotHDRCaptureDebugEnabled`（0x1409715c0）：`sync.Once` 闭包（func1 0x1409f5d00）`Getenv("USBEAM_SCREENSHOT_HDR_DEBUG")`（27B，`0x140c6a710`）；trim 空 → 保持 false；`parseScreenshotBool` ok → 赋值 value；否则 true。
- `screenshotHDRCaptureDebugLog`（0x140971440）：开关关闭 → 返回；否则 `log` 输出 `"[screenshot-hdr] "`（17B，`0x140c574a6`）+ format。
- `screenshotDXGIHDRCaptureAvailable`（0x140971160）：`CreateDXGIFactory1.Find` 失败 → "DXGI 工厂不可用: %v"（24B，`0x140c65520`）+ false；`D3D11CreateDevice.Find` 失败 → "D3D11 设备创建入口不可用: %v"（37B，`0x140c79fe0`）+ false；否则 true。
- 新增全局：`dxgiDLL`/`d3d11DLL`（`windows.NewLazySystemDLL`）+ `procCreateDXGIFactory1`/`procD3D11CreateDevice` + `screenshotHDRCaptureDebugOnce`/`screenshotHDRCaptureDebug`。
- 字符串与函数名二进制实证：`CreateDXGIFactory1`@0xC582B6、`D3D11CreateDevice`@0xC55A0F、`dxgi.dll`@0xC3A834、`d3d11.dll`@0xC3DE13。

## G4 测试

新增 `backend/screenshot_hdr_windows_test.go`：

- `TestParseScreenshotBool`：truthy/falsy/invalid 三组。
- `TestParseScreenshotHDRCaptureMode`：19 用例覆盖 disabled/auto/force/默认分支。
- `TestResolveScreenshotHDRCaptureMode`：`t.Setenv` 验证 env 读取。

全 PASS。

## 判定

四路 PASS，批次 48 闭环。
