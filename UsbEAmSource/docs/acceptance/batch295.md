# 批次 295 · launcherupdate 进程身份采集 +1 [S]（FUNCS 2951）

## 目标

继续 launcherupdate_helper_windows.go 蓝图落地：进程身份采集（queryLauncherUpdateProcessIdentity）。

## 基线 / 收口

| 指标 | 基线（batch 294 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2950 | 2951 |
| MARKED | 2950 | 2951 |
| S | 1427 | 1428 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1444 | 1444 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1464 | 1465 |
| USABLE | 1465 | 1466 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1465  FUNCS=2951  MARKED=2951  P=41  S-eq=1  S-inline=37  S-sig=1444  S=1428  USABLE=1466
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1428 + 37 + 1 + 1444 + 41 = 2951 = FUNCS`。
S 1427→1428（+1）、FUNCS 2950→2951（+1）、FAITHFUL 1464→1465（+1）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 queryLauncherUpdateProcessIdentity [S 0x1408c3d60, 2112B]

`(pid uint32) (launcherUpdateProcessIdentity, error)`。OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION|
SYNCHRONIZE=0x101000) 失败透传，defer CloseHandle；GetProcessTimes 取创建时间（Filetime
组合 High<<32|Low 后 `*100+0x5e6684f4b3960000` = Nanoseconds，时间戳常量 hex 实证为
2^64-11644473600000000000）；queryLauncherUpdateParentPID；ProcessIdToSessionId；
OpenProcessToken(TOKEN_QUERY=8) 后 defer CloseHandle；GetTokenUser 取 SID.String()；
launcherUpdateTokenIntegrityRID；launcherUpdateProcessImagePath；openLauncherUpdateReadOnlyHandle
+ GetFileInformationByHandle（VolumeSerialNumber/FileIndexHigh/Low）；hashLauncherUpdateFile。
返回 launcherUpdateProcessIdentity 96 字节（字段布局与 types_launcher.go:332 完全对齐，
偏移 pid@0/ppid@4/createdAt@8/sessionId@16/userSid@24/integrityRid@40/imagePath@48/
imageSha256@64/volumeSerial@80/fileIndexHigh@84/fileIndexLow@88 逐字节实证）。

## G4 独立复核

- `launcherupdate_helper_windows.go`：+`queryLauncherUpdateProcessIdentity`。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

1 个函数落地 [S]。FUNCS 2951/4754 = 62.07%。下一批：截图滚动几何 /
launcherupdate 剩余短函数。
