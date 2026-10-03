# 批次 371 · 远程图标目标/预览会话脚本/原生预览窗口消息/启动权限默认值 +4（FUNCS 3240）

## 目标

落地 4 个函数：`getRemoteIconTarget`、`screenshotPreviewUpdateSessionScript`、
`screenshotNativePreviewWindowProc`、`BootstrapService.loadAppLaunchPrivilegeDefault`。

## 基线 / 收口

| 指标 | 基线（batch 370 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3236 | 3240 |
| MARKED | 3236 | 3240 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1653 | 1657 |
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
FAITHFUL=1541  FUNCS=3240  MARKED=3240  P=41  S-eq=1  S-inline=37  S-sig=1657  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1657 + 41 = 3240 = FUNCS`。
S-sig 1653→1657（+4）、FUNCS 3236→3240（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 getRemoteIconTarget [S-sig 0x140962060, 448B]

TrimSpace(path) → NewRequestWithContext → 命中→回调。

### 3.2 screenshotPreviewUpdateSessionScript [S-sig 0x14099b200, 448B]

Marshal(参数) → TrimSpace → Marshal → Sprintf。

### 3.3 screenshotNativePreviewWindowProc [S-sig 0x14099c2c0, 448B]

HashTrieMap.Load 命中→handleMessage；否则 LazyProc.Call(DefWindowProc)。

### 3.4 loadAppLaunchPrivilegeDefault [S-sig 0x1408a65a0, 448B]

workspaceSnapshot → loadLauncherConfigOrDefaultIfMissing → normalizeAppLaunchPrivilegeMode →
匹配 admin/standard/followLaunch → 否则 standard。

## G4 独立复核

- `backend/remoteicons.go`：+getRemoteIconTarget [S-sig]。
- `backend/screenshot_preview_windows.go`：+screenshotPreviewUpdateSessionScript [S-sig]
  +screenshotNativePreviewWindowProc [S-sig]。
- `backend/bootstrapservice_window.go`：+loadAppLaunchPrivilegeDefault [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3240/4754 = 68.15%。下一批：filesearch/oled 域
VolumeIndex.activeEntryCountLocked / VolumeIndex.overlayStatsLocked。
