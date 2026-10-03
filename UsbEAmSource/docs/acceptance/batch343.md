# 批次 343 · USN 元数据归一/Chromium 祖先/文件定位/最大化快照 +4（FUNCS 3127）

## 目标

落地 4 个函数：`normalizeFileSearchUSNFollowerMeta`、`buildChromiumRootAncestry`、
`BootstrapService.StartFileLocatorSearch`、`BootstrapService.storeLauncherVerticalMaximizeSnapshot`。

## 基线 / 收口

| 指标 | 基线（batch 342 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3123 | 3127 |
| MARKED | 3123 | 3127 |
| S | 1501 | 1501 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1543 | 1547 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1538 | 1538 |
| USABLE | 1539 | 1539 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1538  FUNCS=3127  MARKED=3127  P=41  S-eq=1  S-inline=37  S-sig=1547  S=1501  USABLE=1539
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1501 + 37 + 1 + 1547 + 41 = 3127 = FUNCS`。
S-sig 1543→1547（+4）、FUNCS 3123→3127（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 normalizeFileSearchUSNFollowerMeta [S-sig 0x140818580, 320B]

enabled==0→1；normalizeVolumeRoot；limit<字段→字段；timestamp==0→time.Now()。

### 3.2 buildChromiumRootAncestry [S-sig 0x140767fc0, 320B]

resolveChromiumRootNameSegment → TrimSpace(path) → 追加到段 → cleanFolderSegments。

### 3.3 StartFileLocatorSearch [S-sig 0x140786700, 320B]

读 fileLocator(+0x438) → StartSearch → duffcopy 返回结果。

### 3.4 storeLauncherVerticalMaximizeSnapshot [S-sig 0x14079b720, 320B]

newobject(4 字段) → lock(+0x540) → 写 field(+0x500) → unlock。

## G4 独立复核

- `backend/filesearch_index_windows.go`：+normalizeFileSearchUSNFollowerMeta [S-sig]。
- `backend/bookmarks.go`：+buildChromiumRootAncestry [S-sig]。
- `backend/hotkey_dispatch_stubs.go`：+StartFileLocatorSearch [S-sig]
  +storeLauncherVerticalMaximizeSnapshot [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3127/4754 = 65.78%。下一批：desktopwidget 域
desktopWidgetNotificationRecordAfter / desktopWeatherConfigChanged。
