# 批次 380 · 截图缩略图 PNG 链路 +4（FUNCS 3264）

## 目标

1. 新增 `normalizeScreenshotSaveFormatForPath` `[S]`（0x140967ba0，544B）。
2. 新增 `buildScreenshotThumbnailPNG` `[S]`（0x14096baa0，864B）。
3. 新增 `buildScreenshotThumbnailPNGFromData` `[S]`（0x14096b8c0，480B）。
4. 新增 `buildScreenshotThumbnailPNGFromTrustedPath` `[S]`（0x14096b300，544B）。

## 基线 / 收口

| 指标 | 基线（batch 379 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3260 | 3264 |
| MARKED | 3260 | 3264 |
| S | 1524 | 1528 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1657 | 1657 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1561 | 1565 |
| USABLE | 1562 | 1566 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

```
FUNCS=3264  MARKED=3264  UNMARKED=0  S=1528  S-inline=37  S-eq=1  S-sig=1657  P=41  FAITHFUL=1565  USABLE=1566  TRUE=3222
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1528 + 37 + 1 + 1657 + 41 = 3264 = FUNCS`。
FUNCS 3260→3264（+4）、S 1524→1528（+4）、FAITHFUL 1561→1565（+4）、TRUE 3218→3222（+4）、UNMARKED=0/P=41 保持。
`list_missing.js` 未落地数 367→363（−4），四函数退出缺失清单。

## G3 行为（asm 逐地址实证）

### 3.1 normalizeScreenshotSaveFormatForPath [S 0x140967ba0]

签名 `(path, fallbackFormat string) (format, path string, err error)`（2 string 入参，3 值出参）。
- TrimSpace(path) 空 → `("png", "", "截图保存路径不能为空")`（30B @0x140c6f76b）。
- 末尾回扫：遇 `/`/`\` 断（无扩展名）；取最后 `.` 后缀 ToLower。
- `.png`→`("png",path)`；`.jpg`/`.jpeg`→`("jpg",path)`。
- 其他非空扩展名 → `("png", "", "仅支持保存 PNG 或 JPG 截图")`（34B @0x140c75e37）。
- 无扩展名 → format=normalizeScreenshotSaveFormat(fallback)，path 拼接 ".png"/".jpg"（4B @0x140c347ae/0x140c347aa）。

### 3.2 buildScreenshotThumbnailPNG [S 0x14096baa0]

- nil → "截图缩略图来源为空"（27B @0x140c6a35f）；宽/高 ≤0 → "截图缩略图尺寸无效"（27B @0x140c6a37a）。
- 竖长图 `width*3 < height`：cover 顶部裁剪，目标 360×220，裁剪高 `max(1,min(h,w*220/360))`
  （魔数 `0x5b05b05b05b05b06` 经验证 = 2^71/360，除以 360）。
- 否则 contain：宽>360 → tw=360, th=max(1,h*360/w)；th>220 → th=220, tw=max(1,tw*220/th)。
- `image.NewRGBA(0,0,tw,th)` → `draw.CatmullRom.Scale(dst,dst.Bounds(),img,sr,draw.Src,nil)`
  （kernel 0x141bc55f0 Support=2.0 = CatmullRom；op=1=draw.Src）→
  `encodeScreenshotRGBAWithKlauspostPNG((*screenshotRGBAImageRowSource)(dst))`。

### 3.3 buildScreenshotThumbnailPNGFromData [S 0x14096b8c0]

- len==0 → "截图数据为空"（18B @0x140c5a00e）。
- `decodeScreenshotImageBytesWithBudget(data, "png", 12)` 失败 → `fmt.Errorf("解析截图缩略图失败: %w")`（31B @0x140c7113d）。
- defer release → buildScreenshotThumbnailPNG(img)。

### 3.4 buildScreenshotThumbnailPNGFromTrustedPath [S 0x14096b300]

- TrimSpace 空 → "截图路径不能为空"（24B @0x140c654d8）。
- `readScreenshotImageFileLimited` 失败透传；`decodeScreenshotImageBytesWithBudget(raw, "", 12)` 失败 → 同上 fmt.Errorf。
- defer release → buildScreenshotThumbnailPNG(img)。

## G4 独立复核

- `backend/screenshot_funcs.go`：+`normalizeScreenshotSaveFormatForPath`（[S]，依赖 errors/strings 已有）。
- `backend/screenshot_scroll_windows.go`：+3 函数（[S]）；import 增 `errors`/`fmt`/`strings`/`golang.org/x/image/draw`。
- 复用已落地 `decodeScreenshotImageBytesWithBudget`/`readScreenshotImageFileLimited`/
  `encodeScreenshotRGBAWithKlauspostPNG`/`screenshotRGBAImageRowSource`，无新依赖。
- `go build/vet/test` 全仓通过；`go fmt` 无差异。

## 移交（本轮收尾）

4 函数落地（均 [S]）。FUNCS 3264/4754 = 68.66%，FAITHFUL 1565/4754 = 32.92%，TRUE 3222/4754 = 67.77%。
下一批：plugin update catalog 域剩余（`mergeLocalAndRemotePluginManifest`/`remoteCatalogEntryToManifest`/
`validateAndNormalizeRemoteCatalogEntries`）；`buildScreenshotThumbnailPNGFromPath`(0x14096b200) 体；
filesearch 域 480B 候选；`P=41 → [S]` 转换。
