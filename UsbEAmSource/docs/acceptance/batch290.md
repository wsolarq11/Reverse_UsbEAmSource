# 批次 290 · launcherupdate helper 重启删除登记 +1 [S]（FUNCS 2943）

## 目标

继续 launcherupdate_helper_windows.go 蓝图落地：helper 重启后删除登记
（scheduleLauncherUpdateHelperCleanup）。

## 基线 / 收口

| 指标 | 基线（batch 289 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2942 | 2943 |
| MARKED | 2942 | 2943 |
| S | 1419 | 1420 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1444 | 1444 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1456 | 1457 |
| USABLE | 1457 | 1458 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1457  FUNCS=2943  MARKED=2943  P=41  S-eq=1  S-inline=37  S-sig=1444  S=1420  USABLE=1458
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1420 + 37 + 1 + 1444 + 41 = 2943 = FUNCS`。
S 1419→1420（+1）、FUNCS 2942→2943（+1）、FAITHFUL 1456→1457（+1）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 scheduleLauncherUpdateHelperCleanup [S 0x1408c5380, 608B]

无返回值。os.Executable 失败即返回；对 exe 与 filepath.Dir(exe) 分别
UTF16PtrFromString 成功后 MoveFileEx(p,nil,MOVEFILE_DELAY_UNTIL_REBOOT=0x4)。

## G4 独立复核

- `launcherupdate_helper_windows.go`：+`scheduleLauncherUpdateHelperCleanup`。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

1 个函数落地 [S]。FUNCS 2943/4754 = 61.91%。下一批：launcherupdate 短函数
（namedPipe/frameWithTimeout 288-768B）+ 截图滚动几何。
