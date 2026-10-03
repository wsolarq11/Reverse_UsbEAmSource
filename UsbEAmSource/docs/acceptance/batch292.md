# 批次 292 · launcherupdate 认证命名管道 +2 [S]（FUNCS 2947）

## 目标

继续 launcherupdate_helper_windows.go 蓝图落地：认证命名管道创建（createLauncherUpdateNamedPipe）
与客户端连接（openLauncherUpdateNamedPipeClient）。

## 基线 / 收口

| 指标 | 基线（batch 291 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2945 | 2947 |
| MARKED | 2945 | 2947 |
| S | 1422 | 1424 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1444 | 1444 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1459 | 1461 |
| USABLE | 1460 | 1462 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1461  FUNCS=2947  MARKED=2947  P=41  S-eq=1  S-inline=37  S-sig=1444  S=1424  USABLE=1462
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1424 + 37 + 1 + 1444 + 41 = 2947 = FUNCS`。
S 1422→1424（+2）、FUNCS 2945→2947（+2）、FAITHFUL 1459→1461（+2）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 createLauncherUpdateNamedPipe [S 0x1408c2940, 352B]

`(pipeName string) (windows.Handle, error)`。getInfo(TokenUser=1,50) 失败或 User.Sid 空 →
"无法读取当前用户 SID"（hex 解码实证）；SDDL "D:P(A;;GA;;;SY)(A;;GA;;;BA)(A;;GA;;;"+sid+")"
（前缀 36B+后缀 1B 按地址解码）；SecurityDescriptorFromString/UTF16PtrFromString 失败透传；
CreateNamedPipe(FILE_FLAG_OVERLAPPED|FIRST_PIPE_INSTANCE|PIPE_ACCESS_DUPLEX=0x40080003,
PIPE_REJECT_REMOTE_CLIENTS=8, 1, 65536, 65536, 30000, sa)，sa.Length=0x18=24。

### 3.2 openLauncherUpdateNamedPipeClient [S 0x1408c3040, 416B]

`(pipeName string, timeout time.Duration) (windows.Handle, error)`。UTF16PtrFromString 失败透传；
循环 CreateFile(GENERIC_READ|GENERIC_WRITE, 0, OPEN_EXISTING,
FILE_FLAG_OVERLAPPED|FILE_ATTRIBUTE_NORMAL=0x40000080)；成功返回；PIPE_BUSY/FILE_NOT_FOUND
则 Sleep(20ms=0x1312d00) 重试，time.Now().After(deadline) 超时 →
"连接更新认证 named pipe 超时"（hex 解码实证）；其他错误透传。

## G4 独立复核

- `launcherupdate_helper_windows.go`：+`createLauncherUpdateNamedPipe`
  +`openLauncherUpdateNamedPipeClient`。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

2 个函数落地 [S]。FUNCS 2947/4754 = 61.99%。下一批：launcherupdate 短函数
（frameWithTimeout/queryProcessIdentity 640-1632B）+ 截图滚动几何。
