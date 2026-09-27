# 批次 51 验收（screenshot DXGI 枚举链）

日期：2026-09-20
子批次：dxgi_enum（Screenshot 运行时族第十子批次）
目标：DXGI 适配器/输出枚举链 asm 直译——`screenshotDXGIAdapterName` /
`queryScreenshotDXGIOutput6` / `screenshotDisplayCaptureInfoFromDesc` /
`screenshotDXGIOutputDisplayInfo` / `enumerateScreenshotDXGIAdapterDisplays` /
`enumerateScreenshotDisplayCaptureInfos`，新建 `backend/screenshot_dxgi_enum_windows.go`。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| test | `go test -count=1 -p=1 -tags production ./backend` | EXIT=0（ok changeme/backend） |
| gofmt | `gofmt -l` 两个新文件 | 空（无差异） |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1035 / MARKED=846 / S=662 / S-inline=9 / S-sig=103 / P=72 / UNMARKED=189`。

相对批次 50（1026/837/656/6）：+9 标记（6 [S] + 3 [S-inline]），`P`/`UNMARKED` 无新增。

## G3 逻辑等价

关键实证（asm VA → Go）：

- `screenshotDXGIAdapterName`（0x14097eba0）：adapter 零 → `("",nil)`；`vtable[8]` GetDesc（0x40）`SyscallN(GetDesc, adapter, &desc)`；HRESULT 非零 → `fmt.Errorf("%s failed: HRESULT 0x%08X", "IDXGIAdapter.GetDesc", uint32(hresult))`；成功 → `TrimSpace(UTF16ToString(desc.Description[:]))`。
- `queryScreenshotDXGIOutput6`（0x14097f180）：output 零 → `errors.New("DXGI 输出为空")`；`vtable[0]` QueryInterface(`SyscallN(qi, output, &IID_IDXGIOutput6, &output6)`)；HRESULT 非零 → `fmt.Errorf(..., "IDXGIOutput6 unavailable", ...)`；output6 零 → `errors.New("IDXGIOutput6 查询结果为空")`。
- `screenshotDisplayCaptureInfoFromDesc`（0x14097f320）：AdapterName/DeviceName `TrimSpace`；Bounds 做 left/right 与 top/bottom min-max 归一化；`HDR = ColorSpace∈{12,13,14,16,18,19}`；`AdvancedColor = ColorSpace!=0 || BitsPerColor>8`。
- `screenshotDXGIOutputDisplayInfo`（0x14097ed40）：`queryScreenshotDXGIOutput6` → defer `releaseDXGIUnknown(output6)` → `vtable[27]` GetDesc1（0xd8）→ `screenshotDisplayCaptureInfoFromDesc`。
- `enumerateScreenshotDXGIAdapterDisplays`（0x14097e7c0）：`make([]screenshotDisplayCaptureInfo,0,2)`；循环 `vtable[7]` EnumOutputs（0x38）；NOT_FOUND → `(infos,nil)`；其他 HRESULT → 错误；output 零 → continue；逐输出 `screenshotDXGIOutputDisplayInfo` 后 append。
- `enumerateScreenshotDisplayCaptureInfos`（0x14097e1c0）：`createDXGIFactory1` → `queryDXGIFactory6` → `make(...,0,4)`；循环 `vtable[7]` EnumAdapters（0x38）；逐适配器 `enumerateScreenshotDXGIAdapterDisplays` 后 append。
- `dxgiAdapterVtable`/`dxgiFactoryVtable`/`dxgiOutput6Vtable` 对应 asm `mov rdx,[rax]; mov rax,[rdx]` 内联模式（[S-inline]）。
- vtable 偏移全链确认：`QueryInterface@0x00`、`EnumOutputs@0x38`、`GetDesc@0x40`、`EnumAdapters@0x38`、`GetDesc1@0xd8`。关键修正：`IDXGIOutput` 实为 **12** 方法（含 `SetDisplaySurface`，winapi 0.3.9 手册计数曾漏此项），故 `GetDesc1` 在槽 27（0xd8），与 asm `[rdx+0xd8]` 精确一致。
- GUID 实证：`IID_IDXGIOutput6 = 068346E8-AAEC-4B84-ADD7-137F513F77A1`@0x141962ad0。
- 字符串实证：`"IDXGIFactory.EnumAdapters"`@0x140c66d30；`"IDXGIAdapter.EnumOutputs"`@0x140c65550；`"IDXGIAdapter.GetDesc"`@0x140c5d953；`"IDXGIOutput6.GetDesc1"`@0x140c5f9a7；`"IDXGIOutput6 unavailable"`@0x140c65580；`"DXGI 输出为空"`@0x140c574b7；`"IDXGIOutput6 查询结果为空"`@0x140c71273；`"%s failed: HRESULT 0x%08X"`@0x140c66b0a。

## G4 测试

新增 `backend/screenshot_dxgi_enum_windows_test.go`（8 用例）：

- `TestIIDIDXGIOutput6`：IID 全字段字节序断言。
- `TestDXGIAdapterDescLayout`：`Sizeof==0x130` + `Description/VendorID/DedicatedVideoMemory/AdapterLuidLow` 偏移。
- `TestDXGIOutputDesc1Layout`：`Sizeof==0x98` + 10 个关键字段偏移。
- `TestScreenshotDisplayCaptureInfoFromDesc`：trim / 归一化 / HDR / AdvancedColor / 亮度全断言。
- `TestScreenshotDisplayCaptureInfoFromDescAdvancedColor`：10bpc/8bpc SDR 两分支。
- `TestScreenshotDXGIAdapterNameNil`：nil adapter → `("",nil)`。
- `TestQueryScreenshotDXGIOutput6Nil`：nil output → 含 `DXGI 输出为空` 的错误。

全 PASS。

## 判定

四路 PASS，批次 51 闭环。
