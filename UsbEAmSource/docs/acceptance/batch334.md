# 批次 334 · 三字签名/天气语言归一/提醒分钟解析 +3（FUNCS 3092，65.04%）

## 目标

落地 3 个纯函数：`nameTrigramSignature`（三字签名）、
`normalizeDesktopWeatherLanguage`（天气语言归一化）、
`parseDesktopReminderMinute`（提醒分钟解析）。

## 基线 / 收口

| 指标 | 基线（batch 333 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3089 | 3092 |
| MARKED | 3089 | 3092 |
| S | 1492 | 1495 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1518 | 1518 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1529 | 1532 |
| USABLE | 1530 | 1533 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1532  FUNCS=3092  MARKED=3092  P=41  S-eq=1  S-inline=37  S-sig=1518  S=1495  USABLE=1533
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1495 + 37 + 1 + 1518 + 41 = 3092 = FUNCS`。
S 1492→1495（+3）、FUNCS 3089→3092（+3）、FAITHFUL 1529→1532（+3）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 nameTrigramSignature [S 0x1407e2b80, 288B]

len<3→0；循环 trigram 小写（A-Z→|0x20）→去重相邻重复→(h>>11)^h*0x45d9f3b→
(h>>16)^h→bts 位图（与 nameBigramSignature 同族）。

### 3.2 normalizeDesktopWeatherLanguage [S 0x1407c7d00, 288B]

TrimSpace→ToLower→IndexAny(s,"-_")；分隔符不在位置 2 或前 2 字符非 a-z→"en"；否则 s[:2]。

### 3.3 parseDesktopReminderMinute [S 0x1407b0e80, 288B]

TrimSpace→time.Parse("15:04")；err!=nil→(0,false)；否则 Hour*60+Minute→(n,true)。

## G4 独立复核

- `backend/filesearch_windows.go`：+nameTrigramSignature [S]。
- `backend/desktopwidgets_weather.go`：+normalizeDesktopWeatherLanguage [S]。
- `backend/desktopwidgets_service.go`：+parseDesktopReminderMinute [S]（+time import）。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（3 [S]）。FUNCS 3092/4754 = 65.04%（跨 65%）。下一批：bookmarks 域
sanitizeBookmarkIconFilename / resolveFirefoxBookmarkFolderTitle / filesearch 域。
