# 批次 375 · bookmark 源类型/路径段 + weather 时间派生 +7（FUNCS 3252）

## 目标

落地 6 个函数（4 新增 `[S]` + 2 个 `[S-sig]`→`[S]` 升级），并修正 1 处 `[S]` 边界 bug：

- 新增 `[S]`：`desktopWeatherTimes`、`cleanFolderSegments`、`resolveBookmarkBrowserKind`、`describeBookmarkSourceDescriptor`。
- 升级 `[S-sig]`→`[S]`：`buildChromiumRootAncestry`、`resolveBookmarkSourceKind`（原 `interface{}` 占位签名替换为真实签名）。
- 修正 `[S]`：`desktopWeatherFloat` 负值守卫 `f < 0` → `f < -math.MaxFloat64`（原实现误把所有负数归零）。

## 基线 / 收口

| 指标 | 基线（batch 374 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3248 | 3252 |
| MARKED | 3248 | 3252 |
| S | 1508 | 1514 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1661 | 1659 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1545 | 1551 |
| USABLE | 1546 | 1552 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

```
FUNCS=3252  MARKED=3252  UNMARKED=0  S=1514  S-inline=37  S-eq=1  S-sig=1659  P=41  FAITHFUL=1551  USABLE=1552  TRUE=3210
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1514 + 37 + 1 + 1659 + 41 = 3252 = FUNCS`。
FUNCS 3248→3252（+4）、S 1508→1514（+6）、S-sig 1661→1659（−2）、FAITHFUL 1545→1551（+6）、TRUE 3206→3210（+4）、UNMARKED=0/P=41 保持。
`list_missing.js` 未落地数 379→375（−4），4 个新增函数均退出缺失清单。

## G3 行为（asm 逐地址实证）

### 3.1 resolveBookmarkSourceKind [S 0x14076ad20, 448B]

`ToLower(TrimSpace(path))` 匹配 kind 名：edge/brave/chrome/vivaldi/chromium（`0x65676465`/`0x76617262`/`0x6f726863`/
`0x61766976`/`0x6d75696d6f726863`）→ 统一返回 `"chromium"`（8B @0x140c3c0fc）；`"firefox"`（7B）→ `"firefox"` @0x140c39300。
否则 `ToLower(TrimSpace(Base(TrimSpace(source))))` 匹配 `"bookmarks"`（9B `0x6b72616d6b6f6f62`）→ `"chromium"`、
`"places.sqlite"`（13B `0x732e736563616c70`）→ `"firefox"`；否则空（`xor eax/ebx`）。

### 3.2 resolveBookmarkBrowserKind [S 0x14076aee0, 512B]

`ToLower(TrimSpace(path))` 非空 → 直接返回；`kind=="firefox"`（7B 精确比较）→ `"firefox"`；
`ToLower(TrimSpace(name))` 依次 `Index` `"edge"`(4)/`"brave"`(5)/`"vivaldi"`(7)/`"chromium"`(8)/`"chrome"`(6)
（`internal/stringslite.Index`）命中即返；否则 `TrimSpace(kind)`。

### 3.3 describeBookmarkSourceDescriptor [S 0x14076a9a0, 0x340B]

`kind=resolveBookmarkSourceKind(source,path)`；firefox/chromium 分支分别走
`describeFirefoxBookmarkPath`/`describeChromiumBookmarkPath`(source) → (name,resolvedPath)，
再 `resolveBookmarkBrowserKind(kind,name,path)` → browserKind；default（kind 空）分支
`browserKind=TrimSpace(path)`、`name=inferName(Dir(TrimSpace(source)))`、resolvedPath 空。返回 4 字符串。

### 3.4 buildChromiumRootAncestry [S 0x140767fc0, 320B]

`resolveChromiumRootNameSegment(name)` 非空 → 段追加；`TrimSpace(path)` 非空 → 段追加；
`cleanFolderSegments` 规整返回。段数组 cap=2（`mov esi,2` 栈数组优化，无 growslice）。

### 3.5 cleanFolderSegments [S 0x140768dc0, 512B]

空入参 → 非 nil 空切片（`lea rax,[rip+zerobase]`）；`makeslice([]string,0,len)`；
逐段 `TrimSpace` 空跳过、与前一追加段 `==`（先比长度 `cmp rbx,rcx` 再 `memequal`）跳过，否则 append。

### 3.6 desktopWeatherTimes [S 0x1407c93a0, 0x1f7B]

三次 `time.Time.Format`（布局 35B @0x140c7708d = `time.RFC3339Nano` "2006-01-02T15:04:05.999999999Z07:00"）；
`Add` 时长常量 `0x1a3185c5000`=1800s(30min)、`0x4e94914f0000`=86400s(24h)。返回 (now, now+30min, now+24h)。

### 3.7 desktopWeatherFloat 修正 [S 0x1407c95c0]

`ucomisd` NaN→0（`jne`/`jp`）、`> +MaxFloat64`（@0x1411cd838）→0、`< -MaxFloat64`（@0x1411cd858）→0，否则返回原值。
原实现 `f < 0` 误将负数归零，已按汇编边界常量修正为 `f < -math.MaxFloat64`。

## G4 独立复核

- `backend/bookmarks_firefox.go`（新）：`resolveBookmarkSourceKind`/`resolveBookmarkBrowserKind`/`describeBookmarkSourceDescriptor` 均 `[S]`，import `path/filepath`+`strings`。
- `backend/bookmarks.go`：`buildChromiumRootAncestry`（`[S-sig]`→`[S]` 真实签名 `(string,string) []string`）、新增 `cleanFolderSegments [S]`、删除 `resolveBookmarkSourceKind` 旧 interface 存根（迁至 firefox 域）。
- `backend/desktopwidgets_weather.go`：`desktopWeatherFloat` 边界修正、新增 `desktopWeatherTimes [S]`。
- 依赖 `describeFirefoxBookmarkPath`/`describeChromiumBookmarkPath`/`inferName`/`resolveChromiumRootNameSegment`（既有签名）保持。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

6 函数落地（4 新增 [S] + 2 升级）+ 1 bug 修正。FUNCS 3252/4754 = 68.41%，FAITHFUL 1551/4754 = 32.62%，TRUE 3210/4754 = 67.52%。
下一批：bookmarks_firefox.go 剩余（`detectBookmarkSources`/`detectFirefoxBookmarkSources`/`resolveFirefoxProfileSources` 域），
或 filesearch/oled 域 480B 候选（list_missing 头部），或 P=41 → [S] 转换。
