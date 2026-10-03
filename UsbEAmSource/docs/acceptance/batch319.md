# 批次 319 · 路径缓存重置/Firefox 配置名/热键反注册 +3（FUNCS 3043）

## 目标

落地 3 个短函数：`nodePathCache.Reset`（路径缓存重置）、
`normalizeFirefoxProfileName`（Firefox 配置名规整）、
`unregisterLauncherHotkey`（热键反注册）。

## 基线 / 收口

| 指标 | 基线（batch 318 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3040 | 3043 |
| MARKED | 3040 | 3043 |
| S | 1473 | 1475 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1488 | 1489 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1510 | 1512 |
| USABLE | 1511 | 1513 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1512  FUNCS=3043  MARKED=3043  P=41  S-eq=1  S-inline=37  S-sig=1489  S=1475  USABLE=1513
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1475 + 37 + 1 + 1489 + 41 = 3043 = FUNCS`。
S 1473→1475（+2）、S-sig 1488→1489（+1）、FUNCS 3040→3043（+3）、FAITHFUL 1510→1512（+2）、
UNMARKED=0/P=41 保持。**突破 64%**。

## G3 行为（asm 逐地址实证）

### 3.1 nodePathCache.Reset [S 0x1407e1620, 256B]

nil 返回；清空 sliceCache；n>cap 则 make([]string,n) 否则截断长度；清空 deltaCache。

### 3.2 normalizeFirefoxProfileName [S 0x14076cfa0, 256B]

TrimSpace(name) 非空返回；否则 Base(path) 去首段点号扩展名。

### 3.3 unregisterLauncherHotkey [S-sig 0x1408a15a0, 256B]

newobject 写 id → LazyProc.Call(id) → 失败则 fmt.Errorf。

## G4 独立复核

- `backend/filesearch_windows.go`：+nodePathCache.Reset [S]。
- `backend/bookmarks.go`：+normalizeFirefoxProfileName [S]（+filepath import）。
- `backend/launcherglobalhotkey.go`：+unregisterLauncherHotkey [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（2 [S] + 1 [S-sig]）。FUNCS 3043/4754 = 64.01%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
