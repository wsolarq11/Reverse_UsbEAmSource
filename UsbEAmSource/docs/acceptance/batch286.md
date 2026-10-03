# 批次 286 · launcherupdate 进程身份判定 + 令牌完整性 RID +2（全 [S]，FUNCS 2938）

## 目标

继续 launcherupdate_helper_windows.go 蓝图落地：进程身份同一判定（launcherUpdateSameIdentity，
逐字段比对 launcherUpdateProcessIdentity）与令牌完整性级别 RID 提取（launcherUpdateTokenIntegrityRID）。

## 基线 / 收口

| 指标 | 基线（batch 285 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2936 | 2938 |
| MARKED | 2936 | 2938 |
| S | 1413 | 1415 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1444 | 1444 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1450 | 1452 |
| USABLE | 1451 | 1453 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1452  FUNCS=2938  MARKED=2938  P=41  S-eq=1  S-inline=37  S-sig=1444  S=1415  USABLE=1453
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1415 + 37 + 1 + 1444 + 41 = 2938 = FUNCS`。
S 1413→1415（+2）、FUNCS 2936→2938（+2）、FAITHFUL 1450→1452（+2）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 launcherUpdateSameIdentity [S 0x1408c4be0, 320B]

`(a,b launcherUpdateProcessIdentity) bool`。逐字段短路比较：PID/ParentPID/CreatedAt/
SessionID 严格相等；UserSID 用 strings.EqualFold；IntegrityRID 严格；ImagePath 用
samePathFold；ImageSHA256 用 strings.EqualFold；VolumeSerial/FileIndexHigh 严格；
FileIndexLow 严格相等为终值。结构偏移与 launcherUpdateProcessIdentity 完全对齐。

### 3.2 launcherUpdateTokenIntegrityRID [S 0x1408c4a40, 416B]

`(token windows.Token) (uint32, error)`。GetTokenInformation(TokenIntegrityLevel)
先探测 size，size==0 → "无法读取令牌完整性级别"；make([]byte,size) 再取，err 透传；
buffer 首 8 字节为 TOKEN_MANDATORY_LABEL.Label.Sid 指针，SID 空或 SubAuthorityCount==0 →
"令牌完整性 SID 无效"；否则 SubAuthority(count-1)。

## G4 独立复核

- `launcherupdate_helper_windows.go`：+`launcherUpdateSameIdentity` +`launcherUpdateTokenIntegrityRID`
  （import +errors +unsafe）。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

2 个函数落地（全 [S]）。FUNCS 2938/4754 = 61.80%。下一批：launcherupdate 短函数
（handshakeProof/parentPID/namedPipe 等 320-768B）+ 截图滚动几何。
