# 批次 361 · 平均行亮度/滚动步长更新/三字头匹配/开始菜单路径解析 +4（FUNCS 3200）

## 目标

落地 4 个函数：`screenshotScrollingAverageRowLuma`、`screenshotScrollingScrollController.update`、
`volumeNameTrigramHeaderMatchesIndex`、`resolveLauncherStartMenuShortcutPath`。

## 基线 / 收口

| 指标 | 基线（batch 360 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3196 | 3200 |
| MARKED | 3196 | 3200 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1613 | 1617 |
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
FAITHFUL=1541  FUNCS=3200  MARKED=3200  P=41  S-eq=1  S-inline=37  S-sig=1617  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1617 + 41 = 3200 = FUNCS`。
S-sig 1613→1617（+4）、FUNCS 3196→3200（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 screenshotScrollingAverageRowLuma [S-sig 0x1409a3be0, 384B]

越界→负常量；step=max(1,宽/100)；采样累加 r*0x12b+g*0x24b+b*0x72 → sum/count。

### 3.2 screenshotScrollingScrollController.update [S-sig 0x1409a5ec0, 384B]

nil/非正→return；按滚动速度分段算术 → 写 step(+0x8)。

### 3.3 volumeNameTrigramHeaderMatchesIndex [S-sig 0x1407f9ca0, 416B]

nil/magic(0x20003)/baseEntryCountLocked 校验 → 头字段比较 → timeToUnixNano 比较。

### 3.4 resolveLauncherStartMenuShortcutPath [S-sig 0x14086cb00, 416B]

Getenv(7)→TrimSpace→非空返回；UserConfigDir→err→fmt.Errorf；filepath.join(… Start Menu …)。

## G4 独立复核

- `backend/screenshot_scroll_windows.go`：+screenshotScrollingAverageRowLuma [S-sig]
  +screenshotScrollingScrollController.update [S-sig]。
- `backend/types_screenshot.go`：+screenshotScrollingScrollController 类型。
- `backend/filesearch_index_windows.go`：+volumeNameTrigramHeaderMatchesIndex [S-sig]。
- `backend/launcherappidentity_windows.go`：+resolveLauncherStartMenuShortcutPath [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3200/4754 = 67.31%。下一批：filesearch 域
searchWithPathsContextMetrics / searchCandidatesMetrics。
