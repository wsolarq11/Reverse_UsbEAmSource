# 批次 101 验收（QR 码生成与剪贴板复制链）

日期：2026-09-23
子批次：qrcode（QR 码生成 + clipboard 复制）
目标：还原 QR 码生成链（gozxing 编码 → 透明 NRGBA 渲染 → PNG → data URL）与
剪贴板复制链（clipboard 库初始化 → 图片写入），引入两个原生依赖
`github.com/makiuchi-d/gozxing` 与 `golang.design/x/clipboard`；
校正 `GenerateQRCodeDataURL` / `CopyQRCodeImageToClipboard` 两方法的签名证伪
（均缺 `size int` 参数）。新建 `backend/screenshot_qrcode_windows.go`。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| test | `go test -tags production -count=1 -p=1 -run 'TestClampQRCodeDimension|TestGenerateQRCodePNG|TestRenderTransparentQRCodeImage|TestGenerateQRCodeDataURL' ./backend` | EXIT=0（5 用例全 PASS） |
| gofmt | 手工 edit/write 落笔，未用 `gofmt -w` | 通过 vet 校验 |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1152 / MARKED=1152 / S=841 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。

相对批次 100（1149/1149/836/35/277/1/0）：+3 新函数（`clampQRCodeDimension` /
`generateQRCodePNG` / `renderTransparentQRCodeImage`），`copyQRCodeImageToClipboard` /
`generateQRCodeDataURL` 由 `[S-sig]` 补体转 `[S]`（S-sig −2），`initQRCodeClipboard`
空壳补体（原已 `[S]` 标记）。`UNMARKED=0` 保持，`P=1` 不变。

## G3 逻辑等价

关键实证（asm VA → Go，字符串经 `resolve_lea_strings.py` 确定性 RIP 算术解码）：

- `initQRCodeClipboard`（0x140931900, 44L）：`sync.Once.Do(doSlow)` 守卫，
  doSlow 闭包（0x1409f6100）= `clipboard.Init()`；首次调用后返回全局缓存的 error
  （`[rip+0x12def7d]`/`[rip+0x12def7e]` 两个 qword = error{itab,data}）。
- `generateQRCodePNG`（0x140931a40, 608B）：`strings.TrimSpace(data)` 仅用于空内容判定
  （`二维码内容不能为空`@0x140c6a2bd,27B；原始 data 参与编码，TrimSpace 结果丢弃）；
  尺寸 clamp 内联：size<0xc0(192)→0x1a4(420)、size>0x200(512)→0x200、否则原值（width/height 各独立）；
  `(&qrcode.QRCodeWriter{}).Encode(data, BarcodeFormat_QR_CODE=0xb, w, h, nil)`（receiver 空结构
  栈地址 `lea rax,[rsp+0x3f]` 实证），失败 `生成二维码失败: %w`@0x140c66ce5(25B)；
  `renderTransparentQRCodeImage` → `png.Encode`（内联 `(&png.Encoder{}).Encode(&buf, img)`），
  失败 `编码二维码图片失败: %w`@0x140c710e0(31B)；返回 `buf.Bytes()`（`[rdx+0x18]=off` 切片偏移实证）。
- `renderTransparentQRCodeImage`（0x140931ca0, 416B）：`NewNRGBA(Rect(0,0,w,h))`（w/h 负数
  cmovl 防御 = `Rect(min,0,max,0)` 规范化，正常输入等价标准 Rect）；
  位读取展开自 `gozxing.BitMatrix.Get`：`bits[rowSize*y + x/32]>>(x&31)&1`（字段偏移
  width=0/height=8/rowSize=0x10/bits.ptr=0x18/bits.len=0x20 实证）；
  位=1（黑模块）→ `NRGBA{0,0,0,0xff}`，位=0 → `NRGBA{0xff,0xff,0xff,0}`。
- `generateQRCodeDataURL`（0x140931e40, 192B）：`generateQRCodePNG` 失败透传；空 PNG 返回
  `("",nil)`；否则 `"data:image/png;base64,"`@0x140c61160(22B) + `base64.StdEncoding`。
- `copyQRCodeImageToClipboard`（0x140931f00, 224B）：`initQRCodeClipboard` 失败 →
  `初始化剪贴板失败: %w`@0x140c6bdf0(28B)；`generateQRCodePNG` 失败透传；否则
  `clipboard.Write(FmtImage=1, png)`（返回 channel 丢弃，xor eax/ebx 后返回 nil 实证）。
- 签名证伪纠正：`GenerateQRCodeDataURL` / `CopyQRCodeImageToClipboard` 方法转发 3 字
  （`mov rax,rbx; mov rbx,rcx; mov rcx,rdi`）= `(data string, size int)`，非旧 stub 的
  `(data string)` 单参；`size` 即尺寸 clamp 输入。

## G4 测试

新建 `backend/screenshot_qrcode_windows_test.go`（5 用例，确定性、不触 Windows API/clipboard）：

- `TestClampQRCodeDimension`：8 组边界（<192 / 192-512 / >512）。
- `TestGenerateQRCodePNGEmpty`：空串与全空白 → `二维码内容不能为空`。
- `TestGenerateQRCodePNG`：非空 → PNG 可解码、`*image.NRGBA`、尺寸为正。
- `TestRenderTransparentQRCodeImage`：4×4 BitMatrix 左上 2×2 置位 → 黑模块不透明黑、
  白模块 alpha=0。
- `TestGenerateQRCodeDataURL`：前缀 `data:image/png;base64,` 校验。

全 PASS。

## 判定

四路 PASS，批次 101 闭环。
