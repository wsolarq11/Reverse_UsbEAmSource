# 批次 147 验收（qrcode 域：原生屏幕选区/标注 — 上集）

日期：2026-09-25
子批次：qrcode-native-a

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `bash build.sh`（`go build -tags production -trimpath -buildmode=exe ./backend`） | EXIT=0 |
| vet | `go vet ./backend` | EXIT=0 |
| 全量 test | `go test ./backend` | ok (0.430s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1361 / MARKED=1361 / S=990 / S-inline=35 / S-sig=335 / P=1 / UNMARKED=0`。

相对批次 146（1302/977/35/289/1）：FUNCS +59，S +13，S-sig +46，其余不变（P=1 仍为
`newOLEDLifecycleContext`）。

新增 59 个（均 `backend/qrcode_native_a.go`）：13 个 [S]（纯几何/数学）+ 46 个 [S-sig]
（Win32/GDI/LazyProc/包级全局依赖，体为恒返零值骨架）。

## G3 逻辑等价

- [S] 纯函数已逐条译自 asm，代表性：`shouldRefreshQRCodeControlHover`（阈值 4px/90ms）、
  `shouldReuseQRCodeControlHover`（严格内含 + 面积≤240000）、`qrCodeControlSelectionFromCachedHover`
  （空判 + point.In + Intersect）、`offsetQRCodeAnnotationStroke`（start/end/points 平移）、
  `unionQRCodeAnnotationRects`、`clampQRCodeAnnotationRect`、`qrCodeAnnotationPointsBounds`
  （max+1 增长）、`qrCodeAnnotationTextEstimatedBounds`
  （width=max(textWidth, max(80, runeCount*max(12,fontSize))), height=max(18, base+8)）。
- [S-sig] 签名均经 asm 序言实证（寄存器 ABI 入参/返回值、内存类结构栈透传），体为零值骨架。
- 子代理已透明标注 5 处签名存疑点：`parseQRCodeAnnotationHexColor` /
  `averageQRCodeAnnotationColor`（asm 返回整宽寄存器，整数宽度按 color.RGBA 语义假定 uint8）、
  `createQRCodeAnnotationPen`（颜色 4 独立字节寄存器，b/g 字节序与 color.RGBA 字段序不一致，
  按 4 个 uint8 形参建模）、`withQRCodeAnnotationTextMeasureHDC`（回调形参最可辩护推断）、
  `normalizeQRCodeAnnotationClipboardText`（4 组 NewReplacer 字面量在 .rodata 未解码）、
  `drawQRCodeAnnotationSmallText`（颜色字节寄存器顺序未完全证明）。

## G4 复验

`go test ./backend` 全量 PASS（ok 0.430s）。本批新增多为 [S-sig] 骨架，无新增行为路径；
既有测试全通过确认无回归。已修正一处子代理 VA 誊写错误（`newQRCodeScreenSelectionSession`
0x140934ec0 → 0x140934f20）。

## 判定

四路 PASS，批次 147 闭环。59 个 qrcode 选区/标注函数（上集）落地，`[P]` 保持 1、
`UNMARKED=0`，三项门禁 EXIT=0。

## 下一步

- 批次 148：qrcode 下集（`backend/qrcode_native_b.go`，59 函数：RGBA 绘制/圆角蒙版/
  选区边框/缩放句柄/PNG 编码）。
- 批次 149：二因素包级函数（URI 链 buildTwoFactorProvisioningURI/buildTOTPProvisioningURI/
  buildSteamProvisioningURI、parse 链 parseTwoFactorProvisioningToken/parseTwoFactorImportText、
  normalize/duplicate/allocate/filter/time 辅助）。
