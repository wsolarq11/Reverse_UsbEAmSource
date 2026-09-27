# 批次 43 验收（screenshot_image_budget 内存预算/受限读取/解码链）

日期：2026-09-20
子批次：image_budget（Screenshot 运行时族第二子批次）
目标：`screenshotImageMemoryBudget.Reserve` / `readScreenshotImageFileLimited` /
`decodeScreenshotImageBytesWithBudget` / `validateScreenshotImageConfig` /
`inspectScreenshotImageWork` / `prepareScreenshotImageWork` /
`reserveScreenshotImageBounds` 全量 asm 直译，修正调用点签名。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| test | `go test -count=1 -p=1 -tags production ./backend` | EXIT=0（ok changeme/backend） |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=985 / MARKED=796 / S=616 / S-inline=5 / S-sig=103 / P=72 / UNMARKED=189`。

相对批次 42（982/793/613）：+3 函数全为 `[S]`（`validateScreenshotImageConfig` / `prepareScreenshotImageWork` / `reserveScreenshotImageBounds`），另 5 个既有函数升级签名/实现（Reserve、inspectScreenshotImageWork、buildScreenshotResultMetadataFromPNG、readScreenshotImageFileLimited、decodeScreenshotImageBytesWithBudget）。

## G3 逻辑等价

关键实证（asm VA → Go）：

- `Reserve`（0x14097f5a0）签名修正为 `(size int64) (func(), error)`：返回 `sync.Once` 包裹的一次性释放闭包（`used -= size`，钳制到 0）。`nil` receiver → "图片内存预算器不可用"；`size <= 0` → 空函数；超限错误 3 参（size/used/limit）。
- 全局预算 `screenshotImageMemoryBudgetGlobal` 的 `limit` 静态初始化为 `0x20000000`（512 MiB，`main.init` 0x140747328）。
- `validateScreenshotImageConfig`（0x14097f8e0）：`colorModel` 参数 unused；五个校验分支（stride/maxBytes → 尺寸/maxWidth/maxHeight → maxPixels/height < width → maxPixels < width*height || channels<0 → (maxWorkSet-stride)/channels < width*height）后估算 `estimated = stride + channels*width*height`。
- `inspectScreenshotImageWork`（0x14097fbc0）：`DecodeConfig` → `TrimSpace`+`ToLower` 后 `EqualFold(expectedFormat)`（空格式跳过）→ `validate`；错误字符串 46B/28B/39B 全部 rodata 解码。
- `readScreenshotImageFileLimited`（0x140980260）：`TrimSpace` → `OpenFile(O_RDONLY,0)` → `defer Close` → `Stat` → `IsDir || size<=0 || size>64MiB` 报错 → `io.LimitedReader{N:64MiB+1}` 读全 → size 复核。
- `buildScreenshotResultMetadataFromPNG`（0x140966a80）：删除多余 `budget` 参数；`scrolling` 模式选专用配置（512MiB/32768/50000/1.2e8/512MiB）并跳过 Reserve；非 scrolling 走全局 Reserve 后立即 release；`Width/Height` 钳制 ≥0；`len(png)==0` 返回零值 + nil error。
- 调用点修正：`readExternalScreenshotImageFromPath` 改 `decodeScreenshotImageBytesWithBudget(raw, "", 12)` + `defer release()`；`bootstrapservice.go` 改 `buildScreenshotResultMetadataFromPNG(png, "")`。

## G4 测试

新增 `backend/screenshot_image_budget_test.go`：

- `TestReserveScreenshotImageBudget`：占用/释放/一次性/超限/nil receiver。
- `TestValidateScreenshotImageConfig`：正常估算 + stride/尺寸/像素/工作集边界。
- `TestReadScreenshotImageFileLimited`：正常读取 + 目录拒绝 + 空文件拒绝。
- `TestBuildScreenshotResultMetadataFromPNG`：有效 PNG 尺寸 + 空 PNG 零值 + 非 PNG 报错前缀。

四项全 PASS。

## 判定

四路 PASS，批次 43 闭环。
