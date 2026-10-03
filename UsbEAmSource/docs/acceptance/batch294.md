# 批次 294 · launcherupdate 命名管道连接 +1 [S]（FUNCS 2950）

## 目标

继续 launcherupdate_helper_windows.go 蓝图落地：重叠模式命名管道连接（connectLauncherUpdateNamedPipe）。

## 基线 / 收口

| 指标 | 基线（batch 293 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2949 | 2950 |
| MARKED | 2949 | 2950 |
| S | 1426 | 1427 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1444 | 1444 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1463 | 1464 |
| USABLE | 1464 | 1465 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1464  FUNCS=2950  MARKED=2950  P=41  S-eq=1  S-inline=37  S-sig=1444  S=1427  USABLE=1465
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1427 + 37 + 1 + 1444 + 41 = 2950 = FUNCS`。
S 1426→1427（+1）、FUNCS 2949→2950（+1）、FAITHFUL 1463→1464（+1）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 connectLauncherUpdateNamedPipe [S 0x1408c2aa0, 1344B]

`(pipe windows.Handle, timeout time.Duration) error`。CreateEvent(nil,1,0,nil) 失败透传，
defer CloseHandle；ConnectNamedPipe(pipe,&overlapped{HEvent})；err==nil 或
ERROR_IO_PENDING → 返回 nil；ERROR_PIPE_CONNECTED → WaitForSingleObject(event, 纳秒→毫秒
向上取整，>0xffffffff 取 0xfffffffe)；WAIT_OBJECT_0 → GetOverlappedResult(wait=false)
失败透传后返回 nil；WAIT_TIMEOUT → CancelIoEx 非 ERROR_NOT_FOUND 失败 →
"fmt 取消更新认证 named pipe 连接失败: %w"，否则 GetOverlappedResult(wait=true) +
"等待更新认证 named pipe 连接超时"；其他 → "fmt 等待更新认证 named pipe 连接返回未知状态: %d"
（三条字符串均 hex 解码实证）。

## G4 独立复核

- `launcherupdate_helper_windows.go`：+`connectLauncherUpdateNamedPipe`，import +`fmt`。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

1 个函数落地 [S]。FUNCS 2950/4754 = 62.05%。下一批：launcherupdate
（queryLauncherUpdateProcessIdentity 0x1408c3d60, 3.2KB）+ 截图滚动几何。
