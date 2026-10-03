# 批次 314 · 关闭判定/原生预览释放/journal 删除 +3（FUNCS 3029）

## 目标

落地 3 个短函数：`BootstrapService.shouldCloseLauncherWindow`（关闭判定）、
`screenshotPreviewWindowService.releaseNativePreviewWindow`（预览释放）、
`volumeIndexJournalPendingState.ApplyDelete`（journal 删除应用）。

## 基线 / 收口

| 指标 | 基线（batch 313 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3026 | 3029 |
| MARKED | 3026 | 3029 |
| S | 1469 | 1471 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1478 | 1479 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1506 | 1508 |
| USABLE | 1507 | 1509 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1508  FUNCS=3029  MARKED=3029  P=41  S-eq=1  S-inline=37  S-sig=1479  S=1471  USABLE=1509
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1471 + 37 + 1 + 1479 + 41 = 3029 = FUNCS`。
S 1469→1471（+2）、S-sig 1478→1479（+1）、FUNCS 3026→3029（+3）、FAITHFUL 1506→1508（+2）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 BootstrapService.shouldCloseLauncherWindow [S 0x14079a2e0, 224B]

mu(+0x540).Lock + defer Unlock → 读 allowLauncherWindowClose(+0x44b)。

### 3.2 screenshotPreviewWindowService.releaseNativePreviewWindow [S 0x140999800, 224B]

nil 返回；lock(+0x08) + defer Unlock → nativeWindow(+0x28) 匹配则置 nil。

### 3.3 volumeIndexJournalPendingState.ApplyDelete [S-sig 0x140806b60, 224B]

entry 解析 + mapassign_fast64 写 overlay(+0x210)。

## G4 独立复核

- `backend/bootstrapservice_window.go`：+shouldCloseLauncherWindow [S]。
- `backend/screenshot_services.go`：+releaseNativePreviewWindow [S]。
- `backend/filesearch_index_windows.go`：+ApplyDelete [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（2 [S] + 1 [S-sig]）。FUNCS 3029/4754 = 63.72%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
