# 批次 381 · OLED 媒体连续性 remember + 截图 PNG 编码 +2（FUNCS 3266）

## 目标

1. 新增 `oledBlackoutBrowserMediaContinuity.remember` `[S]`（0x1408fc600，544B）。
2. 新增 `encodeScreenshotImagePNG` `[S]`（0x14096c780，480B）。

## 基线 / 收口

| 指标 | 基线（batch 380 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3264 | 3266 |
| MARKED | 3264 | 3266 |
| S | 1528 | 1530 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1657 | 1657 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1565 | 1567 |
| USABLE | 1566 | 1568 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

```
FUNCS=3266  MARKED=3266  UNMARKED=0  S=1530  S-inline=37  S-eq=1  S-sig=1657  P=41  FAITHFUL=1567  USABLE=1568  TRUE=3224
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1530 + 37 + 1 + 1657 + 41 = 3266 = FUNCS`。
FUNCS 3264→3266（+2）、S 1528→1530（+2）、FAITHFUL 1565→1567（+2）、TRUE 3222→3224（+2）、UNMARKED=0/P=41 保持。
`list_missing.js` 未落地数 363→361（−2），两函数退出缺失清单。

## G3 行为（asm 逐地址实证）

### 3.1 oledBlackoutBrowserMediaContinuity.remember [S 0x1408fc600]

签名 `(windowHandle uintptr, processID uint32, processPath, mediaTitle string)`（6 寄存器入参：
rax=receiver、rbx=windowHandle、ecx=processID、rdi+rsi=processPath、r8+r9=mediaTitle，由 morestack
保存序 @0x1408fc7be 实证）。

- nil receiver（@0x1408fc622）→ 直接返回。
- processPath（rdi+rsi）TrimSpace+ToLower → lowerPath（@0x1408fc649/64f/654）；mediaTitle（r8+r9）同 → lowerTitle。
- guard（@0x1408fc67d-69c）：windowHandle==0 || processID==0 || lowerPath=="" || lowerTitle=="" 任一 → 返回。
- lock.Lock（@0x1408fc6c0 cmpxchg → lockSlow）→ entries（+8）nil 则 makemap_small（@0x1408fc6fa）→
  mapassign（key=[rsp+0x38]={handle,+0x08 pid,+0x10 lowerPath ptr,+0x18 len}，@0x1408fc763）→
  value 槽写 lowerTitle（@0x1408fc76d/796）→ unlock（@0x1408fc7a5 xadd → unlockSlow）。
- key 存**原始** windowHandle/processID，path 存 lowercase；value 存 lowercase mediaTitle。

### 3.2 encodeScreenshotImagePNG [S 0x14096c780]

签名 `(img *image.RGBA, rect image.Rectangle) ([]byte, error)`（rax=img、rbx/rcx/rdi/rsi=rect 四角，
@0x14096c7a4-7b1 保存序实证）。

- `image.Rectangle.Intersect`（@0x14096c800）：`rect.Intersect(img.Rect)`（接收者 rect、参数 img.Rect，
  img.Rect 四角从 [rax+0x20]/[+0x28]/[+0x30]/[+0x38] 载入，对应 RGBA.Rect 字段偏移）。
- `reserveScreenshotImageBounds(inter.Min.X, inter.Min.Y, inter.Max.X, inter.Max.Y, 1, 8)`
  （@0x14096c810；stride=1 由 `mov esi,1`、channels=8 由 `mov r8d,8` 实证）。
- err 非空（@0x14096c815 test rbx）→ 返回 `(nil, err)`（@0x14096c81a-860 置 rax/rbx/rcx=0、rdi/rsi=err itab/data）。
- 否则 `defer release()` 后 `encodeQRCodeSelectionPNG(img, rect)`（@0x14096c88a；**用原始 rect**，
  非 intersect 结果，@0x14096c876-885 从 [rsp+0x48]/[+0x50]/[+0x58]/[+0x60] 还原原始 rect 四角）。
- 返回前调 release（@0x14096c8b3-8c0 经 closure +0 word 调用），flag [rsp+0x47] 为 defer 标记。

## G4 独立复核

- `backend/oledblackout.go`：+`remember`（[S]）；签名按 asm 修正为 5 参独立入参，非 key 结构体打包。
- `backend/screenshot_funcs.go`：+`encodeScreenshotImagePNG`（[S]）；复用已落地
  `reserveScreenshotImageBounds` 与 `encodeQRCodeSelectionPNG`，无新依赖。
- `go build/vet/test` 全仓通过；`gofmt -l ./backend` 无差异。

## 移交（本轮收尾）

2 函数落地（均 [S]）。FUNCS 3266/4754 = 68.70%，FAITHFUL 1567/4754 = 32.96%，TRUE 3224/4754 = 67.82%。
下一批：plugin update catalog 域剩余（`mergeLocalAndRemotePluginManifest`/`remoteCatalogEntryToManifest`/
`validateAndNormalizeRemoteCatalogEntries`）；`buildScreenshotThumbnailPNGFromPath`(0x14096b200) 体；
`directLauncherNetworkAccess.newHTTPClient`/`configBackedLauncherNetworkAccess.newHTTPClient` 体；
filesearch 域 480B 候选；`P=41 → [S]` 转换。
