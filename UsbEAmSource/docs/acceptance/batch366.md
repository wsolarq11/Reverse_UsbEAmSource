# 批次 366 · 自动激活配置清除/基础条目计数/书签源类型解析/工具栏虚拟屏钳位 +4（FUNCS 3220）

## 目标

落地 4 个函数：`oledBlackoutService.clearAutoActivatedProfileIfNotVisibleLocked`、
`VolumeIndex.baseEntryCountLocked`、`resolveBookmarkSourceKind`、
`screenshotSelectionToolbarConstrainNativeWindowToVirtualScreen`。

## 基线 / 收口

| 指标 | 基线（batch 365 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3216 | 3220 |
| MARKED | 3216 | 3220 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1633 | 1637 |
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
FAITHFUL=1541  FUNCS=3220  MARKED=3220  P=41  S-eq=1  S-inline=37  S-sig=1637  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1637 + 41 = 3220 = FUNCS`。
S-sig 1633→1637（+4）、FUNCS 3216→3220（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 clearAutoActivatedProfileIfNotVisibleLocked [S-sig 0x14090d6a0, 416B]

profile(+0x198/+0x1a0) TrimSpace 空→return；findProfileLocked → collectScreensLocked →
不相交→清 field(+0x1a0/+0x198)。

### 3.2 baseEntryCountLocked [S-sig 0x1407e8120, 448B]

field(+0x40) 非空→返回；读视图(+0x520/+0x528) 空→0；
volumeIndexMappedReadProvider.Acquire → 计数。

### 3.3 resolveBookmarkSourceKind [S-sig 0x14076ad20, 448B]

TrimSpace+ToLower(source) 匹配 edge/brave/chrome/firefox/vivaldi/chromium → 返回 kind；
否则 Base(path) ToLower 匹配 bookmarks/places.sqlite → 返回；否则 nil。

### 3.4 ConstrainNativeWindowToVirtualScreen [S-sig 0x1409ad4c0, 416B]

hwnd nil→return；LazyProc.Call(GetWindowRect) → rect 空→return；
qrCodeVirtualScreenBounds → clamp → SetWindowPos。

## G4 独立复核

- `backend/oledblackout.go`：+clearAutoActivatedProfileIfNotVisibleLocked [S-sig]。
- `backend/filesearch_index_windows.go`：+VolumeIndex.baseEntryCountLocked [S-sig]。
- `backend/bookmarks.go`：+resolveBookmarkSourceKind [S-sig]。
- `backend/screenshot_selection_toolbar_windows.go`：+ConstrainNativeWindowToVirtualScreen [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3220/4754 = 67.73%。下一批：filesearch/bookmark 域
VolumeIndex.acquireReadView / resolveBookmarkSourceKind 关联节点。
