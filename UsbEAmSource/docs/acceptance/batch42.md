# 批次 42 验收记录 — screenshot_png_fast.go 自定义 PNG 编码器

> 记录方式：每批次四路质检（§6.1）。本文件为 batch 42 的验收存证，全套可回溯。
> 研究用途；版权（c）2026 DOGFIGHT360 合规。

## 0. 范围

`backend/screenshot_png_fast.go` 蓝图 14 函数 + 4 内联 helper 全量 asm 直译落档：

| 档位 | 数量 | 函数 |
|---|---|---|
| [S] | 14 | `screenshotRGBAImageRowSource.Bounds/.Row`、`predictScreenshotPNGPaeth`、`filterScreenshotPNGPaeth`、`convertScreenshotRGBARowToPNG`、`selectScreenshotPNGFilter`、`writeScreenshotPNGChunk`、`writeScreenshotPNGIHDR`、`screenshotPNGIDATChunkWriter.Write/.Close/.flush`、`writeScreenshotPNGFilteredRows`、`encodeScreenshotPNGRowSource`、`encodeScreenshotRGBAWithKlauspostPNG` |
| [S-inline] | 4 | `unpremulScreenshotPNG`、`absScreenshotPNGInt`、`absScreenshotPNGByte`、`sumAbsScreenshotPNG` |
| 幽灵删除 | 1 | `encodeScreenshotPNGWithKlauspost`（symbols.txt 无此符号，VA 0x140994640 实为 `encodeScreenshotRGBAWithKlauspostPNG`） |

反汇编资产：pipeline/tmp 已有 14 个（`encodeScreenshotPNGRowSource`、`writeScreenshotPNGIHDR`、
`writeScreenshotPNGFilteredRows`、`selectScreenshotPNGFilter`、`filterScreenshotPNGPaeth`、
`predictScreenshotPNGPaeth`、`convertScreenshotRGBARowToPNG`、`writeScreenshotPNGChunk`、
`screenshotPNGIDATChunkWriter.Write/.Close/.flush`、`screenshotRGBAImageRowSource.Bounds/.Row`、
`encodeScreenshotRGBAWithKlauspostPNG`）。全部 VA 来自 `symbols.txt`，长度来自相邻符号差。

## G1 编译/静态 — PASS

```bash
go build -tags production -trimpath ./backend            # EXIT=0
go vet   -tags production ./backend                      # EXIT=0
go test  -tags production ./backend                      # ok changeme/backend
bash build.sh                                            # EXIT=0，重建 artifact
```

依赖新增 `github.com/klauspost/compress v1.18.4`（zlib，模块缓存已有，GOPROXY=off 校验通过）。

## G2 语义契约 — 签名/符号/类型清单

| 项 | 判定 | 依据 |
|---|---|---|
| `screenshotRGBAImageRowSource` | 布局修正 | `{source *image.RGBA}`(8B) → `image.RGBA` 底层(40B)：Pix@0 / Stride@0x18 / Rect@0x20，与 `Bounds`/`Row` asm 偏移 `[rax+0x20]…[rax+0x38]` 完全一致 |
| `encodeScreenshotRGBAWithKlauspostPNG` | [S] `(*screenshotRGBAImageRowSource) ([]byte,error)` | asm 0x140994640 读 receiver `+0x20/0x30`（Min.X/Max.X）与 `+0x28/0x38`（Min.Y/Max.Y），非 `*image.RGBA` 接口包装 |
| `screenshotPNGRowSource` | 保留 | `Bounds() image.Rectangle`（itab+0x18）+ `Row(int)([]byte,error)`（itab+0x20），与调用侧一致 |
| `screenshotPNGIDATChunkWriter` | 保留现有字段 | `writer io.Writer@0 / buffer []uint8@0x10 / limit int@0x28 / closed bool@0x30` 与 asm 偏移一致，仅字段名差异 |
| `writeScreenshotPNGChunk` | [S] `(io.Writer,[]byte,[4]byte) error` | len>0xffffffff → "PNG 数据块过大: %d"(23B)；8B 头 + 数据 + CRC32 IEEE |
| `writeScreenshotPNGIHDR` | [S] `(io.Writer,int,int,int,int) error` | 尺寸≤0 → 22B；>2^32 → 28B；13B IHDR，bitdepth=8，colortype=6 |
| `selectScreenshotPNGFilter` | [S] `(src,prev []byte,*screenshotPNGFilterBuffers,count int) (int,[]byte)` | filter0 全零直接返回；Sub/Up/Average/Paeth 严格更小才换 |
| `writeScreenshotPNGFilteredRows` | [S] `(io.Writer,screenshotPNGRowSource) error` | 行长≠期望 → 50B 格式串；alpha 全 255 跳过反预乘；prev=上一行拷贝 |
| 8 个错误字符串 | 全部 rodata 解码 | 见 §0 附注，UTF-8 中文原文已落码 |

错误字符串实证地址：`0x140C614FC`(22B)/`0x140C61512`(22B)/`0x140C6C020`(28B)/
`0x140C6D837`(29B)/`0x140C6A452`(27B)/`0x140C6C03C`(28B)/`0x140C885E2`(50B)/`0x140C63056`(23B)。

## G3 行为自测 — 全部通过

新增 `backend/screenshot_png_fast_test.go`：

| 用例 | 覆盖点 |
|---|---|
| TestEncodeScreenshotPNGOpaque | 5x3 全 alpha=255 像素往返逐字节一致 |
| TestEncodeScreenshotPNGAlpha | premultiplied(a=0/100/255) → 解码 straight 与 unpremul 公式一致 |
| TestWriteScreenshotPNGChunk | chunk 长度/类型/数据头 + 总长 17B（8 头 + 5 数据 + 4 CRC）|

`go test -tags production -run 'TestEncodeScreenshotPNG|TestWriteScreenshotPNGChunk' -v` 全绿。
标准库 `image/png` 解码为 `*image.NRGBA`（straight alpha），与编码器 unpremul 输出一致。

## G4 独立复核 — 结论

1. `unpremulScreenshotPNG` 公式 `(c*0xff00 + a*257 - 1)/(a*257)` 与 asm 0x1409956a2 的
   `shl edx,5 / lea edx,[r10+rdx*8] / imul 0xff00 / div` 逐指令吻合，溢出 clamp `>0xff→0xff`
   对应 asm `cmp eax,0xff / jbe / mov r9d,0xffffffff`。
2. `selectScreenshotPNGFilter` Average 均值 `(left+up)>>1` 与 asm `sar 1`（算术右移，floor）
   一致；Paeth 委托 `filterScreenshotPNGPaeth`，调用点 `rax=filter4.ptr/rbx=len/rcx=cap` 吻合。
3. `writeScreenshotPNGFilteredRows` 中 `current`（asm `[rsp+0xd0]`）在 alpha 全 255 时直接
   引用 `row`，非全 255 时指向 `convertScreenshotRGBARowToPNG` 的独立缓冲 `cur`
   （asm `[rsp+0xc8]`）；prev 每行 memmove 更新为上一行，与 asm 0x1409950dd 一致。
4. itab 断言分支（asm 0x140995265 具体类型哈希查找，Fun[0] 返回值被丢弃）为编译器接口优化，
   不影响 prev 行语义，实现以 memmove 等价覆盖，已在函数注释注明。
5. 幽灵 `encodeScreenshotPNGWithKlauspost` 删除：其旧实现走标准库 `image/png`，
   symbols.txt 无此符号，VA 0x140994640 归属 `encodeScreenshotRGBAWithKlauspostPNG`。

## 结论

批次 42 四门：G1 ✅ / G2 ✅ / G3 ✅ / G4 ✅。
`screenshot_png_fast.go` 14 蓝图函数 + 4 内联 helper 全部落档（14 [S] + 4 [S-inline]），
幽灵符号 `encodeScreenshotPNGWithKlauspost` 删除，类型布局与 2 处调用点同步修正，
自定义 PNG 编码器产出经标准库解码验证正确。
