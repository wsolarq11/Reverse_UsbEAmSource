# 批次 252 · filesearch 域内存映射关闭链 + 前缀桶计数 +3 [S]

## 目标

落地 `filesearch_index_windows.go` 的缺失函数：名字前缀桶计数 + 内存映射文件关闭链
（volumeIndexMappedFile.Close / volumeNameTrigramIndex.close）。

## 基线 / 收口

| 指标 | 基线（batch 251 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2814 | 2817 |
| MARKED | 2814 | 2817 |
| S | 1283 | 1286 |
| S-inline | 36 | 36 |
| S-sig | 1455 | 1455 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2774（58.35%） | 2777（58.42%） |

构建：`bash build.sh` 成功，产物 `artifacts/UsbEAm_Launcher_rebuilt.exe`（21,322,240 B），
SHA256 `6de4d49f118647832e8738a9ee9e04f30860a2bf1dbadb55e73a7187f7a6cbda`。
`go build/vet/test ./backend` 全 EXIT=0。

## 本批落地（+3 [S]，filesearch_index_windows.go 新增 import windows）

1. `addNamePrefixBucketCount(c *namePrefixBucketCounts, name string, count int32)` `[S 0x1407e2620]`
   = nil/空名返回 → 首字节大写折叠小写 → firstByte[c0]+=count → len>=2 时 twoByte[(c0<<8)|c1]+=count。

2. `(*volumeIndexMappedFile).Close() error` `[S 0x1407fd880]`
   = nil→nil；倒序遍历 views 逐个 UnmapViewOfFile（err!=nil 且 firstErr==nil 才累积）→
   views=nil → mapping!=0 则 CloseHandle（同累积）→ mapping=0 → file!=nil 则 file.Close()
   （同累积）→ file=nil → 返回 firstErr。

3. `(*volumeNameTrigramIndex).close()` `[S 0x1407e1920]`
   = nil||closed(+0x40)||mappedFile(+0x38)==nil → 返回；否则 closed=true → 取 mappedFile →
   mappedFile=nil、signatureBytes=nil → mappedFile.Close()。

## 关键知悉

- `volumeIndexMappedFile`（types_filesearch.go:680）= `{file *os.File(+0x00), mapping uintptr(+0x08),
  views []uintptr(+0x10)}`；Close 的「仅保留首个错误」惯用法由 asm 的 `je` 方向反推锁定
  （err!=nil 且 firstErr==nil 才累积）。
- `volumeNameTrigramIndex`（types_filesearch.go:739）= `{version, signatures, signatureBytes,
  mappedFile(+0x38), closed(+0x40)}`；close 只清 signatureBytes（slice 三字）+ mappedFile，
  保留 signatures。
- `addNamePrefixBucketCount` 参数 count 落在 esi（第 5 寄存器），首字节桶 firstByte 内联在
  结构体 +0x00（256×int32=0x400），twoByte.slice 在 +0x400。

## 下一批

P=40 不变。继续按 `gap_aggregate.txt` 长度升序落地已存在文件短函数；FUNCS 2817/4754 = 59.26%。
