# 批次 326 · 二维码显示/日志变更详情/滚动顶部稳定 +3（FUNCS 3065）

## 目标

落地 3 个短函数：`BootstrapService.revealLauncherWindowForQRCodeDecode`（二维码显示）、
`VolumeIndex.ApplyJournalChangesDetailed`（日志变更详情）、
`screenshotScrollingTopSkipLooksStable`（滚动顶部稳定判定）。

## 基线 / 收口

| 指标 | 基线（batch 325 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3062 | 3065 |
| MARKED | 3062 | 3065 |
| S | 1482 | 1482 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1501 | 1504 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1519 | 1519 |
| USABLE | 1520 | 1520 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1519  FUNCS=3065  MARKED=3065  P=41  S-eq=1  S-inline=37  S-sig=1504  S=1482  USABLE=1520
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1482 + 37 + 1 + 1504 + 41 = 3065 = FUNCS`。
S-sig 1501→1504（+3）、FUNCS 3062→3065（+3）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 BootstrapService.revealLauncherWindowForQRCodeDecode [S-sig 0x14079a160, 192B]

接口两方法判定（+0x48/+0x38）→ 满足则 showLauncherWindow(false,true)。

### 3.2 VolumeIndex.ApplyJournalChangesDetailed [S-sig 0x140804c20, 192B]

多参数转发 ApplyJournalChangesDetailedResult。

### 3.3 screenshotScrollingTopSkipLooksStable [S-sig 0x1409a5d60, 224B]

nil/n<=0 守卫 → min(lenA,lenB,n,240) → 帧差采样 <=0xa。

## G4 独立复核

- `backend/bootstrapservice_window.go`：+revealLauncherWindowForQRCodeDecode [S-sig]。
- `backend/filesearch_index_windows.go`：+ApplyJournalChangesDetailed [S-sig]。
- `backend/screenshot_scroll_windows.go`：+screenshotScrollingTopSkipLooksStable [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（3 [S-sig]）。FUNCS 3065/4754 = 64.47%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
