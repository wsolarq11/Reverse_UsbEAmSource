# 批次 352 · USN 空闲持久化判定/输入监听停止/AppUserModelID 设置/Shell 链接创建 +4（FUNCS 3164）

## 目标

落地 4 个函数：`shouldPersistUSNFollowerIdleMeta`、`inputMonitorService.Stop`、
`setCurrentProcessExplicitAppUserModelID`、`createLauncherShellLink`。

## 基线 / 收口

| 指标 | 基线（batch 351 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3160 | 3164 |
| MARKED | 3160 | 3164 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1577 | 1581 |
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
FAITHFUL=1541  FUNCS=3164  MARKED=3164  P=41  S-eq=1  S-inline=37  S-sig=1581  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1581 + 41 = 3164 = FUNCS`。
S-sig 1577→1581（+4）、FUNCS 3160→3164（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 shouldPersistUSNFollowerIdleMeta [S-sig 0x14081d200, 352B]

idleThreshold<=0→true；否则 lastPersist.Add(threshold+30s) 与 now 比较。

### 3.2 inputMonitorService.Stop [S-sig 0x1408641e0, 352B]

StopOwner(8) → duffcopy × 3 → 返回 owner。

### 3.3 setCurrentProcessExplicitAppUserModelID [S-sig 0x14086c620, 352B]

TrimSpace 空→error；UTF16PtrFromString → LazyProc.Call(SetCurrentProcessExplicitAppUserModelID, 1)。

### 3.4 createLauncherShellLink [S-sig 0x14086d780, 352B]

CoCreateInstance(5) → 失败 fmt.Errorf。

## G4 独立复核

- `backend/filesearch_index_windows.go`：+shouldPersistUSNFollowerIdleMeta [S-sig]。
- `backend/inputmonitor_windows.go`：+inputMonitorService.Stop [S-sig]。
- `backend/launcherappidentity_windows.go`：+setCurrentProcessExplicitAppUserModelID [S-sig]
  +createLauncherShellLink [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3164/4754 = 66.55%。下一批：launcher 域
launcherStartupDebugLog / 其余 352B 候选。
