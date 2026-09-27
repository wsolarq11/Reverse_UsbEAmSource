# 批次 45 验收（screenshot_cursor_windows 光标快照数据搬运层）

日期：2026-09-20
子批次：cursor 数据搬运（Screenshot 运行时族第四子批次）
目标：光标快照采集与 DIB/RGBA 双向像素搬运 asm 直译——`currentScreenshotCursorInfo` /
`captureScreenshotCursorSnapshot` / `copyRGBAToQRCodeDIBBits` /
`copyQRCodeDIBBitsToRGBA` / `createQRCodeRGBACompatibleBitmap`，
新建 `backend/screenshot_cursor_windows.go`。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| test | `go test -count=1 -p=1 -tags production ./backend` | EXIT=0（ok changeme/backend） |
| gofmt | `gofmt -l` 两个新文件 | 空（无差异） |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=996 / MARKED=807 / S=627 / S-inline=5 / S-sig=103 / P=72 / UNMARKED=189`。

相对批次 44（991/802/622）：+5 函数全为 `[S]`，`P`/`UNMARKED` 无新增。

## G3 逻辑等价

关键实证（asm VA → Go）：

- `currentScreenshotCursorInfo`（0x140974a20）：CURSORINFO 24B（cbSize=0x18），`GetCursorInfo(&info)`，返回 `(Size,Flags,Cursor,ScreenPos, r!=0)`。
- `captureScreenshotCursorSnapshot`（0x140973f80）：`ok && Cursor!=0 && Flags&1(CURSOR_SHOWING)` 才返回 info，否则零值 + false。
- `copyRGBAToQRCodeDIBBits`（0x140957760）：RGBA→BGRA，目标紧凑 stride=width*4，源偏移 `Stride*(y-Min.Y)+(x-Min.X)*4`，nil 参数直接返回。
- `copyQRCodeDIBBitsToRGBA`（0x140957980）：BGRA→RGBA，`SetRGBA(Min.X+x, Min.Y+y, {R:src[+2],G:src[+1],B:src[+0],A:src[+3]})`，nil 参数直接返回。
- `createQRCodeRGBACompatibleBitmap`（0x140957520）：img nil 或空 → `"创建截图标注文字画布失败: 图像为空"`(50B)；`CreateDIBSection` 失败 → `"创建截图标注文字画布失败"`(36B)；成功后 `copyRGBAToQRCodeDIBBits`。
- 错误字符串 2 个全 rodata 解码：`0x140c8857e`(50B)/`0x140c788f7`(36B)。

## G4 测试

新增 `backend/screenshot_cursor_windows_test.go`：

- `TestCopyQRCodeDIBBitsRoundTrip`：RGBA→BGRA→RGBA 往返像素一致。
- `TestCreateQRCodeRGBACompatibleBitmapNil`：nil 图与空图均报错。

两项全 PASS。

## 判定

四路 PASS，批次 45 闭环。
