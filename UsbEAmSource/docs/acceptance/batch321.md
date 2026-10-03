# 批次 321 · 天气时间取早/链接文本危险判定/预览尺寸拟合 +3（FUNCS 3049）

## 目标

落地 3 个短函数：`earliestDesktopWeatherTime`（天气时间取早）、
`containsUnsafeOpenLinkText`（链接文本危险判定）、
`fitScreenshotPreviewSize`（预览尺寸拟合）。

## 基线 / 收口

| 指标 | 基线（batch 320 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3046 | 3049 |
| MARKED | 3046 | 3049 |
| S | 1477 | 1478 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1490 | 1492 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1514 | 1515 |
| USABLE | 1515 | 1516 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1515  FUNCS=3049  MARKED=3049  P=41  S-eq=1  S-inline=37  S-sig=1492  S=1478  USABLE=1516
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1478 + 37 + 1 + 1492 + 41 = 3049 = FUNCS`。
S 1477→1478（+1）、S-sig 1490→1492（+2）、FUNCS 3046→3049（+3）、FAITHFUL 1514→1515（+1）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 earliestDesktopWeatherTime [S 0x1407c6f40, 256B]

time.Parse(format(35 字符), a/b)；err1→b；err2→a；ta.Before(tb)→a 否则 b。

### 3.2 containsUnsafeOpenLinkText [S-sig 0x140769980, 224B]

Index(危险子串) 命中→true；否则遍历 rune 查控制字符表。

### 3.3 fitScreenshotPreviewSize [S-sig 0x14099b780, 224B]

负值钳 0；零则 (120,120)；等比缩放 + 最小宽度钳位。

## G4 独立复核

- `backend/desktopwidgets_weather.go`：+earliestDesktopWeatherTime [S]（+time import）。
- `backend/bookmarks.go`：+containsUnsafeOpenLinkText [S-sig]。
- `backend/screenshot_services.go`：+fitScreenshotPreviewSize [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（1 [S] + 2 [S-sig]）。FUNCS 3049/4754 = 64.14%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
