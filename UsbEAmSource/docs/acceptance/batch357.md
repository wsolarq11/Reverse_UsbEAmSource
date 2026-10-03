# 批次 357 · qweather host 校验/搜索双字键/搜索三字键/配置替换 +4（FUNCS 3184）

## 目标

落地 4 个函数：`isValidQWeatherAPIHost`、`selectSearchBigramKeys`、
`selectSearchTrigramKeys`、`launcherConfigStore.Replace`。

## 基线 / 收口

| 指标 | 基线（batch 356 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3180 | 3184 |
| MARKED | 3180 | 3184 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1597 | 1601 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1541 | 1541 |
| USABLE | 1542 | 1542 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1541  FUNCS=3184  MARKED=3184  P=41  S-eq=1  S-inline=37  S-sig=1601  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1601 + 41 = 3184 = FUNCS`。
S-sig 1597→1601（+4）、FUNCS 3180→3184（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 isValidQWeatherAPIHost [S-sig 0x1407c9940, 384B]

len>253→false；后缀非 .qweather.com/.net→false；按 "." split → label 空/超 63/首尾 "-"/
非 [a-z0-9-]→false。

### 3.2 selectSearchBigramKeys [S-sig 0x1407e2a00, 384B]

遍历 terms(+0x38 结构) → collectSearchTermBigramKeys 累计。

### 3.3 selectSearchTrigramKeys [S-sig 0x1407e2f80, 384B]

遍历 terms(+0x38 结构) → collectSearchTermTrigramKeys 累计。

### 3.4 launcherConfigStore.Replace [S-sig 0x140899800, 384B]

ReplacePrepared → 大结构(0x126*8) 拷贝返回。

## G4 独立复核

- `backend/desktopwidgets_weather.go`：+isValidQWeatherAPIHost [S-sig]。
- `backend/filesearch_windows.go`：+selectSearchBigramKeys [S-sig] +selectSearchTrigramKeys [S-sig]。
- `backend/launcherconfig.go`：+launcherConfigStore.Replace [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3184/4754 = 66.98%。下一批：filesearch 域
closeVolumeIndexReadProvider / clearNameTrigramIndexLocked。
