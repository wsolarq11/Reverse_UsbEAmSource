# 批次 285 · launcherupdate 辅助域进程映像/终止/回滚重启 +3（全 [S]，FUNCS 2936）

## 目标

继续 launcherupdate_helper_windows.go 蓝图落地：进程完整映像路径查询（launcherUpdateProcessImagePath）、
更新进程终止（terminateLauncherUpdateProcess）、回滚后重启（restartRolledBackLauncher）。

## 基线 / 收口

| 指标 | 基线（batch 284 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2933 | 2936 |
| MARKED | 2933 | 2936 |
| S | 1410 | 1413 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1444 | 1444 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1447 | 1450 |
| USABLE | 1448 | 1451 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1450  FUNCS=2936  MARKED=2936  P=41  S-eq=1  S-inline=37  S-sig=1444  S=1413  USABLE=1451
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1413 + 37 + 1 + 1444 + 41 = 2936 = FUNCS`。
S 1410→1413（+3）、FUNCS 2933→2936（+3）、FAITHFUL 1447→1450（+3）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 launcherUpdateProcessImagePath [S 0x1408c4960, 224B]

`(handle windows.Handle) (string, error)`。栈上 `[0x8000]uint16` 清零，size=0x8000；
QueryFullProcessImageName(handle,0,&buf[0],&size) 失败 → ("",err)；
否则 `filepath.Clean(UTF16ToString(buf[:size]))`。

### 3.2 terminateLauncherUpdateProcess [S 0x1408c4e40, 192B]

`(pid uint32) (windows.Handle, error)`。OpenProcess(PROCESS_TERMINATE|SYNCHRONIZE,false,pid)
失败 → (handle,err)；成功 defer CloseHandle → TerminateProcess(handle,1) →
WaitForSingleObject(handle,10000) → (handle,nil)。调用方忽略返回值。

### 3.3 restartRolledBackLauncher [S 0x1408c4f60, 192B]

`(dir,name string) error`。filepath.Join(dir,name) → exec.Command(path)（无 args）→
cmd.Dir=dir → cmd.Start()。asm 额外携带 dead 中间参数（调用方上下文结构），核心逻辑一致。

## G4 独立复核

- `launcherupdate_helper_windows.go`：+`launcherUpdateProcessImagePath` +`terminateLauncherUpdateProcess`
  +`restartRolledBackLauncher`（import +os/exec）。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（全 [S]）。FUNCS 2936/4754 = 61.77%。下一批：launcherupdate 短函数
（sameIdentity/tokenIntegrityRID/handshakeProof/parentPID 等 320-576B）+ 截图滚动几何。
