# 批次 293 · launcherupdate 带超时帧读写 +2 [S]（FUNCS 2949）

## 目标

继续 launcherupdate_helper_windows.go 蓝图落地：带超时的帧写入（writeLauncherUpdateFrameWithTimeout）
与帧读取（readLauncherUpdateFrameWithTimeout）。

## 基线 / 收口

| 指标 | 基线（batch 292 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2947 | 2949 |
| MARKED | 2947 | 2949 |
| S | 1424 | 1426 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1444 | 1444 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1461 | 1463 |
| USABLE | 1462 | 1464 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1463  FUNCS=2949  MARKED=2949  P=41  S-eq=1  S-inline=37  S-sig=1444  S=1426  USABLE=1464
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1426 + 37 + 1 + 1444 + 41 = 2949 = FUNCS`。
S 1424→1426（+2）、FUNCS 2947→2949（+2）、FAITHFUL 1461→1463（+2）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 writeLauncherUpdateFrameWithTimeout [S 0x1408c3820, 640B]

`(conn *os.File, payload []byte, timeout time.Duration) error`。make(chan error,1) +
go func1 写帧（conn.itab 编译期常量，闭包捕获 conn.data+payload 3 字段）；NewTimer 后
defer Stop；select timer.C → conn.Close() + "写入更新认证 named pipe 超时"（hex 解码实证）；
select ch → 返回写帧 err。

### 3.2 readLauncherUpdateFrameWithTimeout [S 0x1408c3440, 768B]

`(conn *os.File, maxSize int, timeout time.Duration) ([]byte, error)`。make(chan
struct{data,err},1) + go func1 读帧；select timer.C → conn.Close() +
"读取更新认证 named pipe 超时"（hex 解码实证）；select ch → 返回 (data,err)。

## G4 独立复核

- `launcherupdate_helper_windows.go`：+`writeLauncherUpdateFrameWithTimeout`
  +`readLauncherUpdateFrameWithTimeout`。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

2 个函数落地 [S]。FUNCS 2949/4754 = 62.03%。下一批：launcherupdate
（queryLauncherUpdateProcessIdentity/connectLauncherUpdateNamedPipe 1.3-1.6KB）+
截图滚动几何。
