# 批次 362 · 书签路径描述/托盘隐藏销毁判定/日志待定路径解析/AppUserModelID 设置 +4（FUNCS 3204）

## 目标

落地 4 个函数：`describeChromiumBookmarkPath`、
`BootstrapService.shouldDestroyLauncherWindowOnTrayHide`、
`volumeIndexJournalPendingState.resolvePath`、`setShortcutAppUserModelID`。

## 基线 / 收口

| 指标 | 基线（batch 361 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3200 | 3204 |
| MARKED | 3200 | 3204 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1617 | 1621 |
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
FAITHFUL=1541  FUNCS=3204  MARKED=3204  P=41  S-eq=1  S-inline=37  S-sig=1621  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1621 + 41 = 3204 = FUNCS`。
S-sig 1617→1621（+4）、FUNCS 3200→3204（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 describeChromiumBookmarkPath [S-sig 0x140766c20, 416B]

TrimSpace→Clean→空/"."→nil；resolveChromiumBookmarkRoots → filepath.Rel 匹配 → genSplit 首段。

### 3.2 shouldDestroyLauncherWindowOnTrayHide [S-sig 0x14079a6e0, 416B]

nil→false；workspaceSnapshot → loadLauncherConfigIfExists→err→false；
normalizePreferencesWithOptions → 首选项字节非零。

### 3.3 volumeIndexJournalPendingState.resolvePath [S-sig 0x140807140, 416B]

缓存 map 命中→返回；entry → 递归 parent → filepath.join 缓存。

### 3.4 setShortcutAppUserModelID [S-sig 0x14086cca0, 416B]

TrimSpace×3 → path/id 空→error → withLauncherAppIdentityCOM。

## G4 独立复核

- `backend/bookmarks.go`：+describeChromiumBookmarkPath [S-sig]。
- `backend/bootstrapservice_window.go`：+shouldDestroyLauncherWindowOnTrayHide [S-sig]。
- `backend/filesearch_index_windows.go`：+volumeIndexJournalPendingState.resolvePath [S-sig]。
- `backend/launcherappidentity_windows.go`：+setShortcutAppUserModelID [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3204/4754 = 67.40%。下一批：screenshot/launcher 域
screenshotSelectionToolbarState 规范化 / launcherConfigStore.CompareAndSwap。
