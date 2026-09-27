# 批次 105 验收（二维码解码链）

日期：2026-09-23
子批次：qrcode-decode（gozxing 多/单码解码 + 结果转换 + 去重 key）
目标：落地 `tryDecodeMultipleQRCodes` / `decodeQRCodesFromImage` / `convertQRCodeResults`
三函数 `[S]`，`buildQRCodeDecodedEntryKey` / `buildQRCodeDecodedEntryFromPoints`
落 `[S-sig]`（后者坐标公式下一批补体）。与批次 101 的二维码生成侧对称。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| test | `go test -tags production -count=1 -p=1 -run 'QRCode|EntryKey|ConvertQR' ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1165 / MARKED=1165 / S=853 / S-inline=35 / S-sig=276 / P=1 / UNMARKED=0`。

相对批次 104（1160/850/35/274/1/0）：+5 新函数（3 个 `[S]` + 2 个 `[S-sig]`），
`UNMARKED=0` 保持。

## G3 逻辑等价

关键实证（asm VA → Go，字符串经 resolve_lea_strings / read_gostring 解码）：

- `tryDecodeMultipleQRCodes`（0x140932640, 160B）：`DecodeMultiple`（`QRCodeMultiReader`）返回
  后 `test rdi`（err.itab）非空 → 返回 nil；`test rbx`（results.len）空 → 返回 nil；否则透传。
- `decodeQRCodesFromImage`（0x140932400, 576B）：`gozxing.NewBinaryBitmapFromImage`（0x1406e1a00）
  → hints 构建 `mapassign_fast64` key=3（`DecodeHintType_TRY_HARDER`，值 `true`）→
  `tryDecodeMultipleQRCodes` → `img.Bounds()`（`mov rcx,[rcx+0x20]` 解 itab.fun[1]）→
  `convertQRCodeResults`；`test rbx`（entries.len）非空即返回；空则单码回退
  `qrcode.QRCodeReader.Decode`（0x140712ae0）→ 单结果 `convertQRCodeResults`。
- `convertQRCodeResults`（0x1409326e0, 1728B）：nil/空 results → nil；`makeslice(cap=len)`；
  逐 result：nil 跳过 → `strings.TrimSpace(result.GetText())` 空跳过 → `buildQRCodeDecodedEntryFromPoints`
  算坐标 → 填充 Text/Format。Format 映射为 switch 表（`jmp [rcx+rdx*8]`，18 分支），
  字符串逐项实证：`AZTEC`/`CODABAR`/`CODE_39`/`CODE_93`/`CODE_128`/`DATA_MATRIX`/`EAN_8`/
  `EAN_13`/`ITF`/`MAXICODE`/`PDF_417`/`QR_CODE`/`RSS_14`/`RSS_EXPANDED`/`UPC_A`/`UPC_E`/
  `UPC_EAN_EXTENSION`/`unknown format`（与 `gozxing.BarcodeFormat.String()` 完全一致）。
  → `buildQRCodeDecodedEntryKey` → `mapaccess2_faststr`（`map[string]struct{}` 去重）→
  `mapassign_faststr` + `growslice`（`imul rdx,rbx,0x58`，元素 88B = QRCodeDecodedEntry）。
- `buildQRCodeDecodedEntryKey`（0x140933720, [S-sig]）：`!selectable` 分支实证
  `strings.Join([]string{format,text}, ":")`（0x140933b99 `call strings.Join`，sep
  @0x140c6c000:1=`:`）；selectable 分支 `fmt.Sprintf` 坐标格式串为推断（去重语义等价）。
- `buildQRCodeDecodedEntryFromPoints`（0x140932da0, 2432B, [S-sig]）：`cmp rbx,3`（points<3）
  或 `sub r8,rcx`/`sub r9,rsi`（bounds 宽/高<=0）→ 空 entry；坐标精确公式待取证。

## G4 测试

新建 `backend/screenshot_qrcode_decode_test.go`（4 用例）：

- `TestBuildQRCodeDecodedEntryKey`：非 selectable → `"QR_CODE:hello"`；selectable 含坐标。
- `TestConvertQRCodeResultsEmpty`：nil / 空 results → nil。
- `TestConvertQRCodeResults`：`TrimSpace` 生效（`"  hello  "`→`hello`），空文本被剔除，
  Format 映射 `QR_CODE`。
- `TestConvertQRCodeResultsDedup`：两条相同 result 去重后剩 1。

全 PASS。

## 判定

四路 PASS，批次 105 闭环。遗留：`buildQRCodeDecodedEntryFromPoints` 坐标公式、
`buildQRCodeDecodedEntryKey` selectable 格式串、`buildQRCodeDecodeResultFromPNG`
（0x140931fe0 入口，返回类型为 `QRCodeDecodeResult` 大结构，需连带修正
`DecodeQRCodesFromImageData` 的签名），转入批次 106。
