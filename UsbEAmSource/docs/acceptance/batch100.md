# 批次 100 验收（screenshot clipboard 剪贴板写入链）

日期：2026-09-23
子批次：clipboard（Screenshot 运行时族第二十三子批次）
目标：还原 `writeScreenshotPNGToClipboard` 的完整双写链（PNG 自定义格式 + CF_DIBV5），
把 4 个未落地的子函数 + 3 个共享 Win32 helper 从 asm 直译落地；
`writeScreenshotPNGToClipboard` 由旧简化体（w32 直写 CF_DIBV5）重构为完整 asm 链。
新建 `backend/screenshot_clipboard_windows.go`。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| test | `go test -tags production -count=1 -p=1 -run 'TestWrapWinError|TestGlobalAllocMoveableEmpty|TestBuildScreenshotClipboardDIBV5' ./backend` | EXIT=0（3 用例全 PASS） |
| gofmt | 手工 edit/write 落笔，未用 `gofmt -w` | 通过 vet 校验 |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1149 / MARKED=1149 / S=836 / S-inline=35 / S-sig=277 / P=1 / UNMARKED=0`。

相对批次 99（1142/1142/829/35/277/1/0）：+7 `[S]`（`wrapWinError` / `globalAllocMoveable` /
`globalLock` / `registerScreenshotPNGClipboardFormat` / `openScreenshotClipboard` /
`setScreenshotClipboardBytes` / `buildScreenshotClipboardDIBV5`），`P` 不变，
`UNMARKED=0` 保持。`writeScreenshotPNGToClipboard` 体重构（`[S]` 标记不变）。

## G3 逻辑等价

关键实证（asm VA → Go，全部字符串经 `resolve_lea_strings.py` 确定性 RIP 算术解码）：

- `wrapWinError`（0x1408af3a0, 352B）：签名 `(msg string, err error) error`；
  err==nil 或 err==`syscall.Errno(0)`（ERROR_SUCCESS，`windows.Errno` 即其别名）→
  仅 `errors.New(msg)`；否则 `fmt.Errorf("%s: %w", msg, err)`（格式串 `%s: %w`@0x140c37608,6B）。
  共享 helper，qrcode/filesearch 等域复用。
- `globalAllocMoveable`（0x1408ae9e0, 224B）：`(size int) (w32.HGLOBAL, error)`；
  size==0 → `剪贴板数据不能为空`@0x140c69fe4(27B)；`GlobalAlloc(GMEM_MOVEABLE=0x2, size)`
  返回 0 → `wrapWinError("分配剪贴板内存失败"@0x140c69fff,27B)`。
- `globalLock`（0x1408aeac0, 160B）：`(hMem) (unsafe.Pointer, error)`；`GlobalLock` 返回 0 →
  `wrapWinError("锁定剪贴板内存失败"@0x140c6a01a,27B)`。
- `registerScreenshotPNGClipboardFormat`（0x140972ee0, 224B）：`() (uint32, error)`；
  `UTF16FromString("PNG")` → `RegisterClipboardFormatW`；返回 0 →
  `wrapWinError("注册 PNG 剪贴板格式失败"@0x140c72746,32B)`。
- `openScreenshotClipboard`（0x140972fc0, 224B）：40 次循环（每次失败 `Sleep(5ms)`，立即数
  `0x4c4b40`=5e6ns）后额外尝试一次，仍失败 → `wrapWinError("打开截图剪贴板失败"@0x140c6a3cb,27B)`。
- `setScreenshotClipboardBytes`（0x1409730a0, 640B）：`(format uint32, data []byte) error`；
  空数据 → `截图剪贴板数据不能为空`@0x140c7410a(33B)；`globalAllocMoveable(len)` →
  `globalLock` → `copy` → `GlobalUnlock`（不校验返回值）→ `SetClipboardData(format, hMem)`
  返回 0 → `wrapWinError("设置截图剪贴板格式失败"@0x140c7412b,33B)`。
- `buildScreenshotClipboardDIBV5`（0x140973380, 1888B）：`([]byte, error)`；
  `decodeScreenshotImageBytesWithBudget(png,"png",12)` 失败 →
  `解析截图 PNG 失败: %w`@0x140c6a3e6(27B)；`defer release()`；
  w<=0 或 h<=0 → `截图尺寸不能为空`@0x140c65538(24B)；
  w>0x7fffffff 或 h>0x7fffffff → `截图尺寸超出 DIBV5 范围`@0x140c71235(31B)；
  像素字节 `w*h*4 > 0xffffffff` → `截图像素数据超出 DIBV5 范围`@0x140c7a02a(37B)；
  构造 124B BITMAPV5HEADER 模板（bV5Size=0x7c、bV5BitCount=32、bV5Compression=0、
  RedMask=0x00ff0000、GreenMask=0x0000ff00、BlueMask=0x000000ff、AlphaMask=0xff000000、
  bV5CSType=`'Win '`=0x57696e20、bV5Intent=4，均从模板 0x1411df5e4 dump 实证）；
  BGRA 像素：`img.At(b.Min.X+x, b.Max.Y-1-y).RGBA()` 右移 8 位，
  写入顺序 B,G,R,A（自底向上读 + 自顶向下写 + bV5Height 正值三者配合得正确 DIB）。
- `writeScreenshotPNGToClipboard`（0x140972ac0, 928B）：`(png []byte) (err error)`；
  `registerScreenshotPNGClipboardFormat` 失败透传；`buildScreenshotClipboardDIBV5`；
  `LockOSThread` + `defer UnlockOSThread` + `defer` 闭包（`CloseClipboard` 失败且 err 未置 →
  `关闭截图剪贴板失败`@0x140c6a3b0,27B，func1 0x140972e60 实证）；
  `openScreenshotClipboard` 失败透传；`EmptyClipboard` 失败 →
  `清空截图剪贴板失败`@0x140c6a395(27B)；`err1=set(format,png)`，
  `dibErr==nil` 时 `err2=set(CF_DIBV5=0x11,dib)` 否则 `err2=dibErr`；
  err1==nil → nil；err2==nil → `写入截图剪贴板 PNG 格式失败: %w`@0x140c810b6(42B)；
  双失败 → `写入截图剪贴板失败: PNG: %v; CF_DIBV5: %w`@0x140c885b0(50B)。

## G4 测试

新建 `backend/screenshot_clipboard_windows_test.go`（3 用例，确定性、不触 Windows API）：

- `TestWrapWinError`：nil / `syscall.Errno(0)` 两路返回纯消息，非零 errno 走 `%s: %w` 包装。
- `TestGlobalAllocMoveableEmpty`：size=0 报错且不触 GlobalAlloc。
- `TestBuildScreenshotClipboardDIBV5`：2×2 全不透明 PNG → 校验 bV5Size/Width/Height/Planes/
  BitCount/CSType + 4 像素 BGRA 顺序 + 自底向上翻转（row0=图像底行）。

全 PASS。

## 判定

四路 PASS，批次 100 闭环。
