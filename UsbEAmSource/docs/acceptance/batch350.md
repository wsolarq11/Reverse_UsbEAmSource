# 批次 350 · Firefox 图标查询/托盘释放关闭/Widget 存储缓存/读提供者关闭 +4（FUNCS 3156）

## 目标

落地 4 个函数：`queryFirefoxBookmarkIconData`、`BootstrapService.closeLauncherWindowForTrayMemoryRelease`、
`launcherWidgetStoreForPath`、`VolumeIndex.CloseReadProvider`。

## 基线 / 收口

| 指标 | 基线（batch 349 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3152 | 3156 |
| MARKED | 3152 | 3156 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1569 | 1573 |
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
FAITHFUL=1541  FUNCS=3156  MARKED=3156  P=41  S-eq=1  S-inline=37  S-sig=1573  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1573 + 41 = 3156 = FUNCS`。
S-sig 1569→1573（+4）、FUNCS 3152→3156（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 queryFirefoxBookmarkIconData [S-sig 0x140760760, 352B]

nil→(nil,nil)；bookmarkIconCandidates → 遍历 queryFirefoxBookmarkIconBitmap/Root。

### 3.2 closeLauncherWindowForTrayMemoryRelease [S-sig 0x14079a880, 352B]

lock(+0x540) → 置 flag(+0x44b)=true → 回调 cb(+0x18)() → 恢复 flag → unlock。

### 3.3 launcherWidgetStoreForPath [S-sig 0x1407bfd60, 352B]

launcherConfigStorePathKey → HashTrieMap.Load 命中返回；否则 LoadOrStore。

### 3.4 CloseReadProvider [S-sig 0x1407e6d60, 352B]

RLock → closeVolumeIndexReadProvider → clearNameTrigramIndexLocked →
pinyinIndex(+0x538).Close + 置零 → 标志(+0x530)=false。

## G4 独立复核

- `backend/bookmarks.go`：+queryFirefoxBookmarkIconData [S-sig]。
- `backend/hotkey_dispatch_stubs.go`：+closeLauncherWindowForTrayMemoryRelease [S-sig]。
- `backend/desktopwidgets_service.go`：+launcherWidgetStoreForPath [S-sig]。
- `backend/filesearch_index_windows.go`：+CloseReadProvider [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3156/4754 = 66.39%。下一批：filesearch/launcher 域
writeAllContext / buildVolumeIndexPersistenceMetaLocked。
