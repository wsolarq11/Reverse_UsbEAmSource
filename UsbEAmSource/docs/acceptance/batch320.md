# 批次 320 · 天气 provider 归一/名称双字签名/组件运行时判定 +3（FUNCS 3046）

## 目标

落地 3 个短函数：`normalizeDesktopWeatherProviderID`（天气 provider 归一）、
`nameBigramSignature`（名称双字签名）、
`desktopWidgetService.isRuntimeEnabled`（组件运行时判定）。

## 基线 / 收口

| 指标 | 基线（batch 319 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3043 | 3046 |
| MARKED | 3043 | 3046 |
| S | 1475 | 1477 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1489 | 1490 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1512 | 1514 |
| USABLE | 1513 | 1515 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1514  FUNCS=3046  MARKED=3046  P=41  S-eq=1  S-inline=37  S-sig=1490  S=1477  USABLE=1515
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1477 + 37 + 1 + 1490 + 41 = 3046 = FUNCS`。
S 1475→1477（+2）、S-sig 1489→1490（+1）、FUNCS 3043→3046（+3）、FAITHFUL 1512→1514（+2）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 normalizeDesktopWeatherProviderID [S 0x1407c7c00, 256B]

TrimSpace → ToLower → "qweather"→"qweather"；"openmeteo"/"open-meteo"→"openmeteo"；
"weather-api"→"weather-api"；其他→""。

### 3.2 nameBigramSignature [S 0x1407e26c0, 256B]

len<2→0；bigram 小写→去重→(h>>8)^h*0x45d9f3b→(h>>16)^h→bts 位图。

### 3.3 desktopWidgetService.isRuntimeEnabled [S-sig 0x1407bee20, 256B]

RLock(+0x48) → 读多字段(+0x58/+0x80/+0x68/+0x60=="ready")。

## G4 独立复核

- `backend/desktopwidgets_weather.go`：+normalizeDesktopWeatherProviderID [S]。
- `backend/filesearch_windows.go`：+nameBigramSignature [S]。
- `backend/desktopwidgets_service.go`：+isRuntimeEnabled [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（2 [S] + 1 [S-sig]）。FUNCS 3046/4754 = 64.07%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
