# 批次 369 · 输入监视代码标签/通知图标文件/堆读视图获取/画布尾部图像 +4（FUNCS 3232）

## 目标

落地 4 个函数：`inputMonitorCodeLabel`、`ensureLauncherNotificationIconFile`、
`volumeIndexHeapReadProvider.Acquire`、`screenshotScrollingChunkedCanvas.tailImage`。

## 基线 / 收口

| 指标 | 基线（batch 368 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3228 | 3232 |
| MARKED | 3228 | 3232 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1645 | 1649 |
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
FAITHFUL=1541  FUNCS=3232  MARKED=3232  P=41  S-eq=1  S-inline=37  S-sig=1649  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1649 + 41 = 3232 = FUNCS`。
S-sig 1645→1649（+4）、FUNCS 3228→3232（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 inputMonitorCodeLabel [S-sig 0x140869940, 448B]

code==0xd+0x1c+bit0→回车标签；全局 map 命中→value；0x41-0x5a→"Key"+字母；
0x30-0x39→"Digit"+数字；0x70-0x87→"F"+数字；否则 inputMonitorKeyLabel→Replace/Sprintf。

### 3.2 ensureLauncherNotificationIconFile [S-sig 0x14086bbe0, 448B]

全局 icon 数据 nil→error；resolveLauncherNotificationIconPath→err→nil；Dir→MkdirAll→err→fmt.Errorf；
ReadFile 内容 memequal 相同→返回；否则 WriteFile。

### 3.3 volumeIndexHeapReadProvider.Acquire [S-sig 0x140a05100, 448B]

nil→panic；字段(+0x20/+0x28/+0x38/+0x40/+0x48/+0x68/+0x70/+0x78/+0x50/+0x58/+0x60) 拷贝。

### 3.4 tailImage [S-sig 0x14099ffc0, 448B]

nil/非正→nil；image.NewRGBA → 遍历 Row → memmove 复制。

## G4 独立复核

- `backend/inputmonitor_windows.go`：+inputMonitorCodeLabel [S-sig]。
- `backend/launcherappidentity_windows.go`：+ensureLauncherNotificationIconFile [S-sig]。
- `backend/filesearch_index_windows.go`：+volumeIndexHeapReadProvider.Acquire [S-sig]。
- `backend/screenshot_scroll_canvas_windows.go`：+screenshotScrollingChunkedCanvas.tailImage [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3232/4754 = 67.98%。下一批：filesearch/oled 域
VolumeIndex.activeEntryCountLocked / VolumeIndex.overlayStatsLocked。
