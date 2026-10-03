# 批次 289 · launcherupdate 过期 helper 清理 +1 [S]（FUNCS 2942）

## 目标

继续 launcherupdate_helper_windows.go 蓝图落地：认证临时目录下过期 helper 目录清理
（cleanupOldLauncherUpdateHelpers）。

## 基线 / 收口

| 指标 | 基线（batch 288 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2941 | 2942 |
| MARKED | 2941 | 2942 |
| S | 1418 | 1419 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1444 | 1444 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1455 | 1456 |
| USABLE | 1456 | 1457 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1456  FUNCS=2942  MARKED=2942  P=41  S-eq=1  S-inline=37  S-sig=1444  S=1419  USABLE=1457
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1419 + 37 + 1 + 1444 + 41 = 2942 = FUNCS`。
S 1418→1419（+1）、FUNCS 2941→2942（+1）、FAITHFUL 1455→1456（+1）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 cleanupOldLauncherUpdateHelpers [S 0x1408c51a0, 480B]

无返回值。os.ReadDir(os.TempDir()) 失败即返回；遍历目录项，仅当 IsDir 且
ToLower(Name)=="usbeam-launcher-updater-" 且 Info().ModTime().Before(now-24h) 时
os.RemoveAll(filepath.Join(tempDir,name))。时间阈值 -24h（立即数 0xffffb16b6eb10000 实证）。
目录名前缀常量复用 "usbeam-launcher-updater-"（@0x140C65250，与 validateHelperExec 同址）。

## G4 独立复核

- `launcherupdate_helper_windows.go`：+`cleanupOldLauncherUpdateHelpers`（import +time）。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

1 个函数落地 [S]。FUNCS 2942/4754 = 61.89%。下一批：launcherupdate 短函数
（scheduleHelperCleanup/namedPipe/frameWithTimeout 288-768B）+ 截图滚动几何。
