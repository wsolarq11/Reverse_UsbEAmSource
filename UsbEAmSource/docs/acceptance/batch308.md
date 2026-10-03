# 批次 308 · 预览窗口对接/显示源尺寸 +3（FUNCS 3011）

## 目标

落地 3 个短函数：`screenshotPreviewWindowService.AttachApp`（对接应用）、
`screenshotPreviewWindowService.AttachAssets`（对接资源）、
`screenshotPreviewDisplaySourceSize`（显示源尺寸计算）。

## 基线 / 收口

| 指标 | 基线（batch 307 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3008 | 3011 |
| MARKED | 3008 | 3011 |
| S | 1456 | 1459 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1473 | 1473 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1493 | 1496 |
| USABLE | 1494 | 1497 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1496  FUNCS=3011  MARKED=3011  P=41  S-eq=1  S-inline=37  S-sig=1473  S=1459  USABLE=1497
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1459 + 37 + 1 + 1473 + 41 = 3011 = FUNCS`。
S 1456→1459（+3）、FUNCS 3008→3011（+3）、FAITHFUL 1493→1496（+3）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 screenshotPreviewWindowService.AttachApp [S 0x140996240, 192B]

lock(+0x08).Lock → 写屏障写 app(+0x10) → Unlock。

### 3.2 screenshotPreviewWindowService.AttachAssets [S 0x140996300, 192B]

lock(+0x08).Lock → 写屏障写 assets(+0xe0) → Unlock。

### 3.3 screenshotPreviewDisplaySourceSize [S 0x14099b5a0, 192B]

w/h 负值钳 0；w<=0||h<=0 原值返回；TrimSpace(source) 非空且 w*3<h → (360,220)，否则原值。

## G4 独立复核

- `backend/screenshot_services.go`：+AttachApp +AttachAssets +screenshotPreviewDisplaySourceSize [S]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（3 [S]）。FUNCS 3011/4754 = 63.34%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
