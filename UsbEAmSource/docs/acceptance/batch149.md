# 批次 149 验收（qrcode 域：原生屏幕选区/标注 — 下集）

日期：2026-09-25
子批次：qrcode-native-b

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `bash build.sh`（`go build -tags production -trimpath -buildmode=exe ./backend`） | EXIT=0 |
| vet | `go vet ./backend` | EXIT=0 |
| 全量 test | `go test ./backend` | ok (0.405s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1432 / MARKED=1432 / S=1021 / S-inline=35 / S-sig=375 / P=1 / UNMARKED=0`。

相对批次 148（1373/990/35/347/1）：FUNCS +59，S +31，S-sig +28，其余不变（P=1 仍为
`newOLEDLifecycleContext`，UNMARKED 保持 0）。

新增 59 个（均 `backend/qrcode_native_b.go`）：31 个 [S]（体完整翻译）+ 28 个 [S-sig]
（签名证实、体零值骨架）。

**⚠️ 覆盖突破**：本批将函数覆盖推至 `1432 / 4,754 = 30.12%`，越过 30% 目标线。

## G3 逻辑等价

- 代表性 [S]（体完整翻译，签名精确）：`qrCodeAnnotationLinePaintBounds`
  （p0/p1 包围盒外扩 max(8, lineWidth+10)）、`centerRect`、`qrCodeAnnotationArrowHeadPoints`
  （箭头三角顶点）、`qrCodeCapsulePixelCoverage`/`qrCodeEllipseStrokeDistance`/
  `qrCodeEllipseStrokePixelCoverage`（胶囊/椭圆 SDF 覆盖度）、`qrCodeCirclePixelCoverage`、
  `qrCodeRoundedCornerMaskCoverage`（0=rect 外、-1=实心/掩膜无效、0-255=圆角 alpha）、
  `resizeQRCodeSelectionRect`（8 手柄缩放）、`buildQRCodeSelectionDirtyRectsWithSizeLabels`、
  `blendQRCodeRGBAPixel`、`blendQRCodeRGBAByCoverage`、`fillRGBA`、`subtractQRCodeSelectionRect`、
  `qrCodeSelectionSizeLabelText` 等。
- `captureQRCodeVirtualScreenSnapshotWithCursor` 由 [S-sig] 提升为 [S]（体仅构建
  `screenshotScreenCaptureOptions{captureCursor}` 并转发既有后端函数）。
- [S-sig] 签名存疑点（已在注释标注）：`drawQRCodeAnnotationOutlinedText` 后两 int 语义、
  `drawNumberOnRGBA`/`drawTextBlockOnRGBAWithOutline` 栈上文字/布局参数、
  `drawQRCodeAnnotationStroke/Mosaic/BlurOnImageWithOffset` 后 7 int 语义、
  `qrCodeSelectionCornerRadiusLayout` 160 字节返回（[5]image.Rectangle vs 具名结构）、
  `createQRCodePaintBuffer` 40 字节字段语义。

## G4 复验

`go test ./backend` 全量 PASS（ok 0.405s）。修复一处 UNMARKED：子代理引入的 helper
`mulByte`（a*b/255 整数缩放）原无 tier 标记致 `UNMARKED=1`，已内联至
`drawEllipseOnRGBA`/`drawSmoothCircleOnRGBA` 两处调用点并删除该函数，恢复
`UNMARKED=0`（改动仅限本文件，不影响既有测试）。

## 判定

四路 PASS，批次 149 闭环。qrcode 下集 59 函数落地，`FUNCS=1432`（30.12%）达成 30% 目标，
`P=1` 未新增、`UNMARKED=0`，三项门禁 EXIT=0。

## 下一步

- 批次 150：二因素 normalize/parse/time 包级函数（`backend/twofactor_provision_misc.go`，
  12 个，纯 [S-sig]，子代理进行中）。
