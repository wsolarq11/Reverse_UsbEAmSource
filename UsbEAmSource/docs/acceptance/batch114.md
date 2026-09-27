# 批次 114 验收（进程创建链：explorer 作父进程的 CreateProcess 全链闭环）

日期：2026-09-23
子批次：process-launch-createprocess
目标：落地 `createProcessWithShellParentWithVisibility` [S] + `openUnelevatedCommandLine`
[S-sig]→[S]，完成降权启动命令行全链。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1190 / MARKED=1190 / S=884 / S-inline=34 / S-sig=271 / P=1 / UNMARKED=0`。

相对批次 113（1189/882/34/272/1/0）：1 个新函数 `createProcessWithShellParentWithVisibility`
`[S]`；`openUnelevatedCommandLine` [S-sig]→[S]（S +2，S-sig -1）。`UNMARKED=0` 保持。

## G3 逻辑等价

`backend/process_launch_windows.go`（1 函数）+ `bootstrap_pathclip.go`（升级 1 函数）：

- **createProcessWithShellParentWithVisibility**（0x1408ac140, 1828B）：openShellProcessForUnelevatedLaunch
  err → 直接返回；defer CloseHandle(parent)（deferwrap2 @0x1408ac920）；NewProcThreadAttributeList(1)
  err → wrapWinError("初始化进程属性列表失败")（33B @0x140c73ba0），defer attrs.Delete()（deferwrap1
  @0x1408ac8c0）；attrs.Update(PROC_THREAD_ATTRIBUTE_PARENT_PROCESS=0x20000, &parent, 8) err →
  wrapWinError("设置父进程属性失败")（27B @0x140c69f78）；utf16.Encode(cmd/cmdLine)+NUL、
  utf16PtrOrNil(dir)、UTF16FromString("winsta0\\default")；StartupInfoEx{cb=0x70, Desktop,
  ProcThreadAttributeList=attrs.List()}；hidden 真 → STARTF_USESHOWWINDOW|SW_HIDE +
  CREATE_NO_WINDOW|EXTENDED（0x8080000），假 → EXTENDED|CREATE_NEW_CONSOLE（0x80010）；
  CreateProcess err → wrapWinError("启动进程失败")（18B @0x140c59e94）；成功 →
  CloseHandle(Thread/Process)。
- **openUnelevatedCommandLine**（0x1408aa000, 224B）：resolveCmdExePath → "/k pushd "+
  quoteCmdArgument(dir) → buildCreateProcessCommandLine(cmd,args) →
  createProcessWithShellParentWithVisibility(cmd,cmdLine,dir,false)。

## G4 测试

进程创建链均为真实 Win32 副作用（启动 cmd.exe/explorer），无单元断言；全量测试 PASS
（`TestOpenShellProcessForUnelevatedLaunch` 冒烟覆盖外壳进程打开路径）。

## 判定

四路 PASS，批次 114 闭环。降权启动命令行全链（openPathCommandLine → openUnelevatedCommandLine
→ buildCreateProcessCommandLine + createProcessWithShellParentWithVisibility →
openShellProcessForUnelevatedLaunch + isProcessHandleElevated）闭环，`P=1` 不变。
PROC_THREAD_ATTRIBUTE_PARENT_PROCESS 经 x/sys v0.46.0 ProcThreadAttributeListContainer 还原。
