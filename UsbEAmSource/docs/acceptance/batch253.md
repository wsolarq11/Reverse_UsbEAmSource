# 批次 253 · filesearch WAL 截断 + inputmonitor 平台拥有关闭链 +5

## 目标

落地 `filesearch_index_windows.go` 的 `truncateOpenWALAtValidOffset` 与 `inputmonitor.go` 的
平台拥有关闭链（advancePlatformGeneration / stopPlatformThread / stopPlatformOwned / closePlatformOwned）。

## 基线 / 收口

| 指标 | 基线（batch 252 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2817 | 2822 |
| MARKED | 2817 | 2822 |
| S | 1286 | 1290 |
| S-inline | 36 | 36 |
| S-sig | 1455 | 1456 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2777（58.42%） | 2782（58.52%） |

构建：`bash build.sh` 成功，产物 `artifacts/UsbEAm_Launcher_rebuilt.exe`（21,322,240 B），
SHA256 `3e62c33c96d25bb7a2bb0799d3cf53adaa3b835a8d06e6b1b554229dfc342cce`。
`go build/vet/test ./backend` 全 EXIT=0。

## 本批落地（+5）

### backend/filesearch_index_windows.go（+1 [S]，新增 import errors/os）

1. `truncateOpenWALAtValidOffset(f *os.File, offset int64) error` `[S 0x1408008e0]`
   = f==nil||offset<0 → errInvalidWALOffset；f.Truncate(offset) err!=nil → 返回；否则 f.Sync()。
   哨兵错误 `errInvalidWALOffset` 精确消息待取证（.data 静态 error 值 @0x14193c650 为 typeOff
   编码，未解码，暂用 "invalid WAL offset"）。

### backend/inputmonitor.go（+4：+3 [S] +1 [S-sig]）

2. `(*inputMonitorService).advancePlatformGeneration()` `[S 0x140861b60]`
   = nil 返回 → lock(+0x08) → platformGeneration(+0xd0).Add(1) → unlock。
3. `(*inputMonitorService).stopPlatformThread()` `[S-sig 0x140866120]`
   = 签名实证（单 receiver 无参无返回），体为平台线程域大函数（1785B）留待专项。
4. `(*inputMonitorService).stopPlatformOwned()` `[S 0x140861aa0]`
   = advancePlatformGeneration → stopOverride 非 nil 则调用，否则 stopPlatformThread。
5. `(*inputMonitorService).closePlatformOwned()` `[S 0x140861b00]`
   = advancePlatformGeneration → closeOverride 非 nil 则调用，否则 stopPlatformThread。

## 关键知悉

- `truncateOpenWALAtValidOffset` 的哨兵错误静态值（.data @0x14193c650=0x30、@0x14193c658=
  0x00292cff00292ce9）是 Go 1.25 静态 error 的 typeOff 编码，多次 va_read 定位字符串均落空
  （.text/.rdata/.data 偏移都不匹配），按「不伪造」纪律留占位。
- `stopPlatformOwned`/`closePlatformOwned` 的 asm 偏移（+0x108/+0x110）与
  types_inputmonitor.go 字段声明（startOverride@0x100/stopOverride@0x110/closeOverride@0x120）
  存在 8 字节错位，疑似 closeOnce 尺寸/对齐差异；落地以字段名让编译器分配偏移（功能等价）。
- `inputMonitorService` 偏移实证锚点：lock(+0x08)、done(+0xa0)、platformGeneration(+0xd0)。

## 下一批

P=40 不变。继续按 `gap_aggregate.txt`（已重新生成，total missing=2053/top-level=887）长度升序
落地已存在文件短函数；FUNCS 2822/4754 = 59.36%。
