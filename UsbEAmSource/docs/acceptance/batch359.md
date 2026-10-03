# 批次 359 · 子树活动计数/路径解析/输入监听启动/热键注册 +4（FUNCS 3192）

## 目标

落地 4 个函数：`VolumeIndex.baseSubtreeActiveCountsLocked`、
`VolumeIndex.ResolvePathByNodeIndex`、`inputMonitorService.Start`、`registerLauncherHotkey`。

## 基线 / 收口

| 指标 | 基线（batch 358 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3188 | 3192 |
| MARKED | 3188 | 3192 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1605 | 1609 |
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
FAITHFUL=1541  FUNCS=3192  MARKED=3192  P=41  S-eq=1  S-inline=37  S-sig=1609  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1609 + 41 = 3192 = FUNCS`。
S-sig 1605→1609（+4）、FUNCS 3188→3192（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 baseSubtreeActiveCountsLocked [S-sig 0x140803fc0, 384B]

缓存 field(+0x558) 命中→返回 field(+0x550/+0x560)；否则 buildVolumeIndexBaseSubtreeActiveCounts。

### 3.2 ResolvePathByNodeIndex [S-sig 0x140808d40, 384B]

acquireReadView → volumeIndexReadView.resolvePath(nodeIndex)。

### 3.3 inputMonitorService.Start [S-sig 0x140861d20, 384B]

StartOwner(8) → duffcopy × 3 → 返回 owner。

### 3.4 registerLauncherHotkey [S-sig 0x1408a1420, 384B]

newobject(热键结构) → LazyProc.Call(RegisterHotKey, 4) → 失败 fmt.Errorf。

## G4 独立复核

- `backend/filesearch_index_windows.go`：+baseSubtreeActiveCountsLocked [S-sig]
  +ResolvePathByNodeIndex [S-sig]。
- `backend/inputmonitor_windows.go`：+inputMonitorService.Start [S-sig]。
- `backend/hotkeymanager.go`：+registerLauncherHotkey [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3192/4754 = 67.14%。下一批：launcher 域
validateLauncherUpdateRollbackBackup / oledBlackoutService.handleOverlayDismissOnAnyKey。
