# 批次 312 · 运行时版本快照/垂直最大化清理/侧栏固定 +3（FUNCS 3024）

## 目标

落地 3 个短函数：`VolumeIndex.runtimeVersionSnapshot`（版本快照）、
`BootstrapService.clearLauncherVerticalMaximizeSnapshot`（快照清理）、
`BootstrapService.SetLauncherSidebarPinnedOpen`（侧栏固定打开）。

## 基线 / 收口

| 指标 | 基线（batch 311 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3021 | 3024 |
| MARKED | 3021 | 3024 |
| S | 1467 | 1469 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1475 | 1476 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1504 | 1506 |
| USABLE | 1505 | 1507 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1506  FUNCS=3024  MARKED=3024  P=41  S-eq=1  S-inline=37  S-sig=1476  S=1469  USABLE=1507
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1469 + 37 + 1 + 1476 + 41 = 3024 = FUNCS`。
S 1467→1469（+2）、S-sig 1475→1476（+1）、FUNCS 3021→3024（+3）、FAITHFUL 1504→1506（+2）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 VolumeIndex.runtimeVersionSnapshot [S 0x1407e92a0, 224B]

nil 返回 0；mu.RLock + defer RUnlock → 读 runtimeVersion(+0x590)。

### 3.2 BootstrapService.clearLauncherVerticalMaximizeSnapshot [S 0x14079b8c0, 224B]

mu(+0x540).Lock + defer Unlock → verticalMaximizeSnapshot(+0x500) 置 nil。

### 3.3 BootstrapService.SetLauncherSidebarPinnedOpen [S-sig 0x14079aac0, 224B]

resolveLauncherWindow → 非空则 typeAssert 后 applyLauncherSidebarPinnedWindowSizing(window, pinned)。

## G4 独立复核

- `backend/filesearch_index_windows.go`：+runtimeVersionSnapshot [S]。
- `backend/bootstrapservice_window.go`：+clearLauncherVerticalMaximizeSnapshot [S]
  +SetLauncherSidebarPinnedOpen [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（2 [S] + 1 [S-sig]）。FUNCS 3024/4754 = 63.61%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
