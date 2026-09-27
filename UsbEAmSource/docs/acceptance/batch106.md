# 批次 106 验收（二维码解码链收尾 + 返回类型修正）

日期：2026-09-23
子批次：qrcode-decode-closeout（坐标公式补体 + buildQRCodeDecodeResultFromPNG 入口 + 签名修正）
目标：`buildQRCodeDecodedEntryFromPoints` / `buildQRCodeDecodeResultFromPNG` 由 `[S-sig]`
升级 `[S]`，并修正 QR 解码相关 `BootstrapService` 方法的返回类型（`[]string` →
`QRCodeDecodeResult`）。QR 解码链自此全链落地。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| test | `go test -tags production -count=1 -p=1 -run 'BuildQRCodeDecodeResultFromPNG|ConvertQR|EntryKey' ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1165 / MARKED=1165 / S=855 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。

相对批次 105（1165/853/35/276/1/0）：函数总数不变，`buildQRCodeDecodedEntryFromPoints` /
`buildQRCodeDecodeResultFromPNG` 两个 `[S-sig]` 补体转 `[S]`（S +2，S-sig -2）。`UNMARKED=0` 保持。

## G3 逻辑等价

关键实证（asm VA → Go，字符串经 resolve_lea_strings / read_gostring 解码）：

- `buildQRCodeDecodedEntryFromPoints`（0x140932da0, 2432B）补体：
  `cmp rbx,3`（points<3）/ `sub r8,rcx`+`sub r9,rsi`（bounds 宽/高<=0）→ 空 entry；
  marker 外推 `p0 + (p2 - p1)`（`subsd`/`addsd` 链）；包围盒 = min/max(三点+marker)
  （`minsd` 链 + `pxor` 符号翻转 = math.Max 的 NaN 语义）；归一化 `(v - Min) / Max`
  （`subsd [rsp+0x60]=Min.X` → `divsd [rsp+0x40]=Max.X`）；`[rsp+0x100]`（Selectable）
  仅 duffzero 清零后 3 处 `movzx` 读取、无任何写入 → 恒 false。
- `buildQRCodeDecodeResultFromPNG`（0x140931fe0, 1056B）补体：空 png → 零值；
  `decodeScreenshotImageBytesWithBudget(png,"png",12)`（`lea [rip+0x301a3f]`="png"，
  `r8d=0xc`）失败 → `fmt.Errorf("解析截图内容失败: %w")`（@0x140c6be0c, 28B）；
  `defer release()`（`[rsp+0x130]` 闭包 @0x140932346）；`decodeQRCodesFromImage`；
  `ImageData = "data:image/png;base64,"`（@0x140c61160, 22B）+ `base64.StdEncoding.EncodeToString`
  （`concatstring2` @0x140932180）；`ImageWidth/Height = bounds.Dx()/Dy()`
  （`sub rcx,rdx`/`sub rdi,rbx` @0x1409321d8）；`len(entries)==1` 时
  `SelectedEntryKey = buildQRCodeDecodedEntryKey(...)`（`cmp rdx,1` @0x1409321e6）。
- **返回类型修正**：`DecodeQRCodesFromImageData`（0x140788dc0）asm 实证返回大结构
  （`buildQRCodeDecodeResultFromPNG` 返回经 3 次 `duffcopy+0x33a` 栈搬运），非 `[]string`。
  故 4 个方法签名由 `([]string, error)` 修正为 `(QRCodeDecodeResult, error)`：
  `DecodeQRCodesFromScreenSelection` / `DecodeQRCodesFromImageData` /
  `DecodeQRCodesFromScreenshotRef` / `captureQRCodesFromScreenSelectionForService`。
- **reader 单例修正**：`tryDecodeMultipleQRCodes`（0x140932640）实证 receiver 来自全局
  `[rip+0x12de08b]=0x140c140e0` 经多重指针链取；gozxing `QRCodeReader`/`QRCodeMultiReader`
  内嵌 decoder 字段须经 `New*` 初始化，否则 `GetDecoder()` nil panic。改为包级单例
  `qrMultiReaderSingleton`/`qrReaderSingleton`。

## G4 测试

`backend/screenshot_qrcode_decode_test.go` 增 3 用例（共 7）：

- `TestBuildQRCodeDecodeResultFromPNGEmpty`：空 PNG → 零值结果。
- `TestBuildQRCodeDecodeResultFromPNGInvalid`：无效 PNG → 错误含 `解析截图内容失败`。
- `TestBuildQRCodeDecodeResultFromPNGRoundtrip`：`generateQRCodeDataURL("hello-qr",300)` →
  `decodeDataURLPNG` → `buildQRCodeDecodeResultFromPNG`；断言 ImageData 前缀、Entries[0].Text
  = `hello-qr`、ImageWidth/Height>0。

全 PASS（roundtrip 端到端解码通过）。

## 判定

四路 PASS，批次 106 闭环。QR 解码链（生成/解码/结果转换/入口）全部 `[S]` 落地，无遗留
`[S-sig]` 于该链；仅 `captureQRCodesFromScreenSelectionForService`（0x140933c40，跨
screenshot 域）保留 `[S-sig]`，转批次 107。
