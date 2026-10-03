# 批次 282 · screenshot 原生预览显示/透明度/几何 +4（+3 [S] +1 [S-sig]，FUNCS 2925）

## 目标

继续闭合 screenshotNativePreviewWindow 的显示链（showOnThread）、透明度设置（SetOpacity）与
图像等比居中缩放几何（screenshotNativePreviewImageRect）；redraw 以 [S-sig] 存根占位（1152B
渲染链待专项）。

## 基线 / 收口

| 指标 | 基线（batch 281 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2921 | 2925 |
| MARKED | 2921 | 2925 |
| S | 1399 | 1402 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1443 | 1444 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1436 | 1439 |
| USABLE | 1437 | 1440 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1439  FUNCS=2925  MARKED=2925  P=41  S-eq=1  S-inline=37  S-sig=1444  S=1402  USABLE=1440
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1402 + 37 + 1 + 1444 + 41 = 2925 = FUNCS`。
S 1399→1402（+3）、S-sig 1443→1444（+1）、FUNCS 2921→2925（+4）、FAITHFUL 1436→1439（+3）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 (*screenshotNativePreviewWindow)SetOpacity [S 0x14099d100, 288B]

`(opacity float64) error`。handle()==0 → `errors.New("原生截图预览窗口不可用")`；
否则 clamp：<0→0、>1→1；`PostMessageW(hwnd, 0x86a5, 0, &opacity)`（lParam 指向栈上 clamped
值）→ nil。LazyProc 名 "PostMessageW"（复用 0x141BC2858）。

### 3.2 (*screenshotNativePreviewWindow)showOnThread [S 0x14099d220, 256B]

handle()==0→return；`SetWindowPos(hwnd, TOPMOST, bounds, 0x53=SWP_SHOWWINDOW|SWP_NOACTIVATE|
SWP_NOMOVE|SWP_NOSIZE)` → `ShowWindow(hwnd, 4=SW_SHOWNOACTIVATE)` → `visible(+0xd8)=true` →
`redraw()` → `UpdateWindow(hwnd)`。LazyProc 名 "ShowWindow"/"UpdateWindow"（0x141BC2800/0x141BC2810）。

### 3.3 screenshotNativePreviewImageRect [S 0x14099e580, 288B]

`(img *image.RGBA, maxW, maxH int) image.Rectangle`。img==nil 或 maxW/maxH<=0 或 img 尺寸<=0
→ 零矩形；`scale=min(maxW/imgW, maxH/imgH)`，<=0 → 零矩形；`w/h=max(1,int(imgW/H*scale))`；
`image.Rect((maxW-w)/2, (maxH-h)/2, +w, +h)`（image.Rect 的 Min/Max 交换内联，asd 实证 cmovg 交换）。

### 3.4 (*screenshotNativePreviewWindow)redraw [S-sig 0x14099d3a0, 1152B]

渲染链（依赖 image 绘制 / WM_PAINT 处理）体存根占位，签名 `func()` 已订正。

## G4 独立复核

- `screenshot_preview_native_windows.go`：+`SetOpacity`、`showOnThread`、
  `screenshotNativePreviewImageRect`、`redraw`（存根）；import 扩展 image/math/unsafe；
  常量 +`screenshotNativePreviewOpacityMessage=0x86a5`。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（3 [S] + 1 [S-sig]），零外部依赖（redraw 为渲染链存根）。
FUNCS 2925/4754 = 61.55%。下一批：screenshot preview 剩余（paint 224B、createWindow 672B、
run 576B、SetPayload 608B、decodeImage 依赖 1312B 解码器、handleMessage 1984B 消息循环、
WindowProc 448B、render 512B、redraw 1152B）。
