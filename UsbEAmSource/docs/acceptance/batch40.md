# 批次 40 验收记录 — appicon_windows.go 图标 GDI 层

> 记录方式：每批次四路质检（§6.1）。本文件为 batch 40 的验收存证，全套可回溯。
> 研究用途；版权（c）2026 DOGFIGHT360 合规。

## 0. 范围

重写 `backend/appicon_windows.go`（46 函数 → 45 个 + 1 辅助 `absInt`，删除 ghost 符号
`getBitmapSize`），全部升级为 `[S]` 档；重写 `backend/appicon_windows_test.go`（22 测试）；
修正 7 处外部调用点（`bootstrapservice.go` 1 处 + `startmenu_icon_windows.go` 6 处）。

反汇编资产：`work/disasm/appicon_asm/*.asm.txt`（52 个）逐函数直译。

## G1 编译/静态 — PASS

```bash
go1.25.12 build -tags production -trimpath ./backend   # EXIT=0
go1.25.12 vet   -tags production ./backend             # EXIT=0
go1.25.12 test  -count=1 -p=1 -tags production ./backend  # ok changeme/backend 0.131s
bash build.sh                                          # EXIT=0，重建 artifact
```

artifact 哈希：`1D18BFEE3B5CEEE4A36A4AAAC1107ADC10A82E65B61E293D30F1EB7F45A66AAE`
（SHA256，19,726,848 B，2026-09-20 18:36:39）。

## G2 语义契约 — 签名重写清单

签名依据：52 个 asm 资产 + `docs/goresym/symbols.txt` + `resolve_lea_strings.py` 字符串解码。
逐条核对产物：

| 符号 | 旧（错误） | 新（asm 实证） |
|---|---|---|
| getBitmapSize | 独立函数 | 删除（symbols.txt 无此符号，GetObjectW 内联于 appIconBitmapSize） |
| AppIconOptions | 24B | 64B（IconIndex+pad+Namespace+Size+ImageList+CandsPtr+CandsLen+8B 保留） |
| 全局缓存 | 缺失 | `appIconDataCache sync.Map`（resolve/resolveSystem Load/Swap 同址 0x141C12740） |
| selectAppObject | error `SelectObject failed (%d)` | `SelectObject failed`（无 %d） |
| loadAppIconInfo | `(iconInfo *appIconInfo)` | `(hicon uintptr) (*appIconInfo,int,int,error)` |
| appIconBitmapSize | `(int,int)` | `(int,int,error)`，HbmColor 优先/HbmMask 回退、掩码高度减半 |
| createAppIconCanvas | `(hdc,w,h)(uintptr,error)` | `(w,h)(uintptr,unsafe.Pointer,error)`，hdc 槽恒 0 |
| buildRGBAFromDIBBits | `([]byte,w,h)(*RGBA,error)` | `(unsafe.Pointer,w,h)*RGBA` 单返回 |
| applyAppIconMaskAlpha | 掩码零=透明 | 掩码 RGB 全零=绘制区域（语义反转） |
| appIconHasVisibleAlpha | 半透明判定 | `A != 0` |
| normalizeAppIconPNGWithSize | `([]byte,error)` | `([]byte,bool)`，DrawMask 裁剪链 |
| visibleIconBounds | 单 `image.Rectangle` | `(image.Rectangle,bool)`，nil panic 对齐 asm |
| fitIconToSquare | `(image.Image,int)` 居中平铺 | `(*image.RGBA,int)` 最近邻缩放+居中 |
| namespaceAppIconCacheKey | 双重 trim 近似 | 空键/空 ns 返回原始 key，拼接 TrimSpace(ns)+"/"+原始 key |
| buildPathAppIconCacheKey | `%s@%d` 变体 | modtime!=0 `path:%s#%d`，==0 `path:`+key |
| buildResourceAppIconCacheKey | `(url,ns,modtime)` | `(url string,a,b int64)`，`resource:%s`→`%s#%d`→`%s@%d` |
| getSystemIconIndexWithAttributes | 2 参 | `(path,flags uint32,useFileAttributes bool)(int,error)`，0x4010/0x4000 回退重试 |
| loadPrivateExtractedAppIcon | `size int` | `(path,index uint32,size uint32)`，8 arg 数组 |
| loadShellDefinedAppIcon | `size int` | `(path,index uint32,flags uint16)`，6 arg 数组 |
| loadImageListIcon | `(path string,...)` | `(imagelist uintptr,iconIndex int)`，ImageList_GetIcon 3 arg |
| loadAssociatedAppIcon | `sizeFlags` | `(path,flags uint32)`，dwAttr=(flags&0x10?0x80:0) |
| renderAppIconHandle | `(hicon,size)` | `(hicon uintptr)(*RGBA,*RGBA,error)`，flags 3/1 |
| encodeAppIconHandleToPNG | `(hicon,size)` | `(hicon uintptr)([]byte,error)` |
| saveAppIconHandleAsDataURLWithSize | `(string,error)` | 单 `string`，错误吞掉 |
| 入口链 10 函数 | `(string,error)` | 全部单 `string` 返回 |
| resolveFileSearchIconDataWithHints | `(path,hints uint32)` | `(path,hints string,isFolder bool)` |
| resolveFileSearchTypeIconData | 1 参 | `(ext string,isFolder bool)`，placeholder folder/file/placeholder.+ext |

新增：`procGetIconInfo`、`procSHGetImageList`、`procImageListGetIcon`、`comctl32DLL`、
`iidImageList` GUID、`defaultAppIconSizes [5]uintptr{0x100,0x80,0x40,0x30,0x20}`。

## G3 行为自测 — 全部通过

22 个测试函数重写以匹配新语义，`go test -count=1 -p=1` 全绿：

| 用例 | 覆盖点 |
|---|---|
| TestBuildRGBAFromDIBBits_* | unsafe.Pointer 单返回、nil/非法尺寸空图 |
| TestAppIconHasVisibleAlpha_* | A!=0 判定（含 A=255 视为 visible） |
| TestVisibleIconBounds_* | (Rect,bool) 双返回、nil panic |
| TestFitIconToSquare_Scale | 4x6→8 缩放：newW=5,newH=8,offset(1,0) |
| TestApplyAppIconMaskAlpha_* | 掩码反转：全零=绘制、非零=置 0、无 alpha 时置 0xFF |
| TestNormalizeAppIconPNGWithSize_* | ([]byte,bool)、size 0 默认 256 |
| TestReleaseComObject/DeleteApp/InvalidArgs | nil/零安全、非法参数报错 |

## G4 独立复核 — 结论

独立视角重读实现（不复用写作缓存），发现并处理：

1. `getSystemIconIndexWithAttributes` 初稿残留占位垃圾代码 `index := int(...)` / `_ = index` —— 已删除。
2. `go vet` 报 2 处 `possible misuse of unsafe.Pointer`（uintptr→Pointer 切片）—— 已把
   `createAppIconCanvas` 返回的 `pBits` 与 `AppIconOptions.CandsPtr` 改为 `unsafe.Pointer`
   类型，vet 归零。
3. `draw.DrawMask` 第 6 参 `mp` 误传 `nil`（image.Point 不可 nil）—— 已改 `image.Point{}`。
4. `resolveFileSearchTypeIconData` 初稿存在未使用的 `kind` 变量且占位符拼接与 asm 不符 ——
   已按 asm 0x140749fe0 重写（folder→"folder"；normalize 后空→"file"；否则 "placeholder."+normalized）。
5. 候选列表元素解释确证：shell 循环取元素低 16 位为 nIconSize，private 循环取低 32 位为
   cx/cy；全局列表 `[0x100,0x80,0x40,0x30,0x20]` 从 rodata 0x141be9520 解码。
6. 缓存命中分支类型断言：`HashTrieMap.Load` 命中后比对 value itab == string itab，非 string
   视为未命中；Go 侧以 `v.(string)` 双值断言等价。
7. `resolveFileSearchIconDataWithHints` 初稿误将 `hints` 忽略并用 `filepath.Ext(trimmed)` 派生
   ext —— 重读 asm 0x140749dc0 后修正：`hints` 本身经 `normalizeFileSearchIconExtension` 后即 ext，
   hints 空且 trimmed 非空才尾扫描 trimmed；特殊扩展名匹配无点小写；isFolder 分支传 `("",true)`。
   修正后重新 `bash build.sh` 重建 artifact（哈希更新见 G1）。

## 结论

批次 40 四门：G1 ✅ / G2 ✅ / G3 ✅ / G4 ✅（含 5 处自纠）。
`appicon_windows.go` 全部 45 函数 + `absInt` 为 `[S]` 档，零 UNMARKED。
