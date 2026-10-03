# 批次 337 · 租约释放/静态 mmap 重载/检查点头部/USN follower 停止 +4（FUNCS 3103）

## 目标

落地 4 个 filesearch 域函数：`volumeIndexMappedReadProvider.releaseLease`、
`VolumeIndex.CanReloadStaticMmapCheckpoint`、
`buildVolumeIndexCheckpointHeaderLocked`、`FileIndexService.stopUSNFollowerLocked`。

## 基线 / 收口

| 指标 | 基线（batch 336 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3099 | 3103 |
| MARKED | 3099 | 3103 |
| S | 1498 | 1499 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1522 | 1525 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1535 | 1536 |
| USABLE | 1536 | 1537 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1536  FUNCS=3103  MARKED=3103  P=41  S-eq=1  S-inline=37  S-sig=1525  S=1499  USABLE=1537
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1499 + 37 + 1 + 1525 + 41 = 3103 = FUNCS`。
S 1498→1499（+1）、S-sig 1522→1525（+3）、FUNCS 3099→3103（+4）、FAITHFUL 1535→1536（+1）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 releaseLease [S 0x1407e4440, 288B]

nil 早退；Mutex lock；active>0→active--；active==0 且 closing&&!closed→取 release 置 nil
+closed=true；unlock；release 非空则调用。字段名与 types_filesearch.go 一致。

### 3.2 CanReloadStaticMmapCheckpoint [S-sig 0x1407e8980, 288B]

RLock → defer RUnlock；读字段（+0xe0）；overlayStatsLocked 结果判空。

### 3.3 buildVolumeIndexCheckpointHeaderLocked [S-sig 0x1407f4e80, 288B]

64B 头部（magic 0x58444955="UIDX"、版本 0x20003、字段拷贝）。

### 3.4 stopUSNFollowerLocked [S-sig 0x14081b4e0, 288B]

normalizeVolumeRoot → map 查找 → 删除 → closechan → CancelIoEx。

## G4 独立复核

- `backend/filesearch_index_windows.go`：+releaseLease [S] +CanReloadStaticMmapCheckpoint
  [S-sig] +buildVolumeIndexCheckpointHeaderLocked [S-sig] +stopUSNFollowerLocked [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（1 [S] + 3 [S-sig]）。FUNCS 3103/4754 = 65.27%。下一批：desktopwidget
desktopWidgetScheduler.Stop / oledBlackout 域 forget/stopIdleTimerLocked。
