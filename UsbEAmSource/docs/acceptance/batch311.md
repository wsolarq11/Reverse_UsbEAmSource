# 批次 311 · 文件定位器转发/映射 provider 判定 +5（FUNCS 3021）

## 目标

落地 5 个短函数：`BootstrapService.GetFileLocatorState` / `PauseFileLocatorSearch` /
`ResumeFileLocatorSearch` / `StopFileLocatorSearch`（fileLocator 转发 ×4）、
`VolumeIndex.usesMappedReadProvider`（映射 provider 判定）。

## 基线 / 收口

| 指标 | 基线（batch 310 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3016 | 3021 |
| MARKED | 3016 | 3021 |
| S | 1462 | 1467 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1475 | 1475 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1499 | 1504 |
| USABLE | 1500 | 1505 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1504  FUNCS=3021  MARKED=3021  P=41  S-eq=1  S-inline=37  S-sig=1475  S=1467  USABLE=1505
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1467 + 37 + 1 + 1475 + 41 = 3021 = FUNCS`。
S 1462→1467（+5）、FUNCS 3016→3021（+5）、FAITHFUL 1499→1504（+5）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 BootstrapService.GetFileLocatorState [S 0x140786620, 224B]

读 fileLocator(+0x438) → GetState，duffcopy 透传大结构。

### 3.2 BootstrapService.PauseFileLocatorSearch [S 0x140786840, 224B]

读 fileLocator(+0x438) → PauseSearch，duffcopy 透传。

### 3.3 BootstrapService.ResumeFileLocatorSearch [S 0x140786920, 224B]

读 fileLocator(+0x438) → ResumeSearch，duffcopy 透传。

### 3.4 BootstrapService.StopFileLocatorSearch [S 0x140786a00, 224B]

读 fileLocator(+0x438) → StopSearch，duffcopy 透传。

### 3.5 VolumeIndex.usesMappedReadProvider [S 0x1407e7660, 224B]

nil 返回 false；mu.RLock + defer RUnlock → 比较 readProvider(+0x520) 具体类型 itab。

## G4 独立复核

- `backend/filelocator.go`：+GetFileLocatorState +Pause/Resume/StopFileLocatorSearch [S]。
- `backend/filesearch_index_windows.go`：+usesMappedReadProvider [S]（+reflect import）。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

5 个函数落地（5 [S]）。FUNCS 3021/4754 = 63.55%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
