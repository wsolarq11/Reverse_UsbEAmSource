# 批次 288 · launcherupdate helper 可执行路径校验 +1 [S] + validate 签名修正（FUNCS 2941）

## 目标

继续 launcherupdate_helper_windows.go 蓝图落地：helper 可执行文件路径校验
（validateLauncherUpdateHelperExecutablePath）。连带修正 validateLauncherUpdateAbsolutePath
的返回签名（asm 实证为 `(string,error)`，原实现误写为 `error`）。

## 基线 / 收口

| 指标 | 基线（batch 287 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2940 | 2941 |
| MARKED | 2940 | 2941 |
| S | 1417 | 1418 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1444 | 1444 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1454 | 1455 |
| USABLE | 1455 | 1456 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1455  FUNCS=2941  MARKED=2941  P=41  S-eq=1  S-inline=37  S-sig=1444  S=1418  USABLE=1456
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1418 + 37 + 1 + 1444 + 41 = 2941 = FUNCS`。
S 1417→1418（+1）、FUNCS 2940→2941（+1）、FAITHFUL 1454→1455（+1）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 validateLauncherUpdateHelperExecutablePath [S 0x1408c5020, 384B]

`(path string) error`。TrimSpace+Clean 后过 validateLauncherUpdateAbsolutePath，失败 →
"更新 helper 自身路径无效"；Base 须 EqualFold "UsbEAm_Launcher_Updater.exe"；
ToLower(Base(Dir)) 须等于 "usbeam-launcher-updater-"；samePathFold(Dir(Dir),os.TempDir())。
任一步不满足 → "更新 helper 不在认证临时目录"。常量/错误字符串均按地址解码实证。

### 3.2 validateLauncherUpdateAbsolutePath 签名修正

asm 0x1408ca720 成功路径返回 `(clean,nil)`、错误路径返回 `("",err)`（4 寄存器），
实证真实签名为 `(string,error)`。原实现误写 `error`。本批改为 `(string,error)` 并
更新 3 处调用者（launcherupdate_plan_test.go:67,71、launcherupdate_transaction.go:34,141）
为 `_, err :=`。

## G4 独立复核

- `launcherupdate_helper_windows.go`：+`validateLauncherUpdateHelperExecutablePath`（import +os）。
- `launcherupdate_plan.go`：`validateLauncherUpdateAbsolutePath` 签名 `error` → `(string,error)`。
- `launcherupdate_plan_test.go` / `launcherupdate_transaction.go`：调用点改 `_, err :=`。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

1 个函数落地 [S] + 1 个签名修正。FUNCS 2941/4754 = 61.86%。下一批：launcherupdate
短函数（cleanupOldLauncherUpdateHelpers/scheduleHelperCleanup/namedPipe 288-768B）+
截图滚动几何。
