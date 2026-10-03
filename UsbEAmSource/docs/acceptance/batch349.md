# 批次 349 · Chromium 根名段/Firefox 路径/提醒时间解析/Widget 存储路径 +4（FUNCS 3152）

## 目标

落地 4 个函数：`resolveChromiumRootNameSegment`、`describeFirefoxBookmarkPath`、
`parseDesktopReminderTimeOfDay`、`launcherWidgetStorePath`。

## 基线 / 收口

| 指标 | 基线（batch 348 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3148 | 3152 |
| MARKED | 3148 | 3152 |
| S | 1503 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1566 | 1569 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1540 | 1541 |
| USABLE | 1541 | 1542 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1541  FUNCS=3152  MARKED=3152  P=41  S-eq=1  S-inline=37  S-sig=1569  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1569 + 41 = 3152 = FUNCS`。
S 1503→1504（+1）、S-sig 1566→1569（+3）、FUNCS 3148→3152（+4）、FAITHFUL 1540→1541（+1）、
USABLE 1541→1542（+1）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 resolveChromiumRootNameSegment [S 0x140768100, 352B]

ToLower(TrimSpace) → switch 命中系统根名（other/mobile/synced/account/managed/
bookmark_bar/reading_list）返回 ""；否则 TrimSpace 原样返回。

### 3.2 describeFirefoxBookmarkPath [S-sig 0x14076d0a0, 352B]

Clean(TrimSpace) → 空/"."→nil；Dir → 遍历 resolveFirefoxProfileSources 匹配 →
返回 (name,path)；否则 normalizeFirefoxProfileName。

### 3.3 parseDesktopReminderTimeOfDay [S-sig 0x1407b0800, 352B]

TrimSpace → time.Parse("15:04") → 失败 (0,0,false)；否则 absSec 计算 hour/minute。

### 3.4 launcherWidgetStorePath [S-sig 0x1407bfbc0, 352B]

TrimSpace 空→""；Base+EqualFold("widgets.json") → filepath.Join(Dir, sub)。

## G4 独立复核

- `backend/bookmarks.go`：+resolveChromiumRootNameSegment [S] +describeFirefoxBookmarkPath [S-sig]。
- `backend/desktopwidgets_service.go`：+parseDesktopReminderTimeOfDay [S-sig]
  +launcherWidgetStorePath [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（1 [S] + 3 [S-sig]）。FUNCS 3152/4754 = 66.30%。下一批：bookmark/desktop 域
queryFirefoxBookmarkIconData / launcherWidgetStoreForPath。
