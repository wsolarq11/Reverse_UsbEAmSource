# 批次 107 验收（Shell 执行域：ShellExecuteExW / SHObjectProperties / COM 公寓 / PowerShell）

日期：2026-09-23
子批次：shell-execute-domain
目标：一次落地 Shell 执行域 8 个函数（`shellExecuteProgram` / `showShellProperties` /
`withShellApartment` / `resolveCmdExePath` / `utf16PtrOrNil` / `quoteCmdArgument` /
`startPowerShellEncodedCommand` / `encodePowerShellCommand`），全部 `[S]`。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| test | `go test -tags production -count=1 -p=1 -run 'UTF16PtrOrNil|QuoteCmdArgument|ResolveCmdExePath|EncodePowerShellCommand|WithShellApartment' ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1173 / MARKED=1173 / S=863 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。

相对批次 106（1165/855/35/274/1/0）：8 个新函数全部 `[S]` 落地（S +8，FUNCS +8）。`UNMARKED=0` 保持。

## G3 逻辑等价

关键实证（asm VA → Go；字符串经 read_gostring/resolve_lea_strings 解码；LazyProc 名经
`rip-relative` 定位到 LazyProc 对象 `name` 字段）：

- **shellExecuteProgram**（0x1408aec40, 608B）：`UTF16FromString(program)`→`lpFile`（
  `&program16[0]`，空串 `jbe panicIndex` @0x1408aec95）；`utf16PtrOrNil(verb/args/workingDir)`
  →`lpVerb/lpParameters/lpDirectory`；`SHELLEXECUTEINFOW` cbSize=0x70/fMask=0x50c
  （`movabs 0x50c00000070` @0x1408aed25 = SEE_MASK_INVOKEIDLIST|NOASYNC|FLAG_NO_UI）；
  nShow=1（@0x1408aed89）；ShellExecuteExW 返 0 → `wrapWinError("调用系统操作失败")`。
  参数序 `(verb, program, args, workingDir)`（verb 首参，UTF16FromString 的是第 2 参）。
- **showShellProperties**（0x1408aeea0, 384B）：LazyProc=`SHObjectProperties`（name 18B）；
  `VolumeName` → 卷根判定 `TrimRight(path,"\/")+EqualFold` → objType∈{SHOP_FILEPATH(2),
  SHOP_VOLUMEGUID(4)}；`Call(0,objType,&utf16[0],0)`；返 0 → `wrapWinError("打开属性窗口失败")`。
- **withShellApartment**（0x1408af020, 320B）：LazyProc=`CoInitializeEx`/`CoUninitialize`；
  `LockOSThread` → `CoInitializeEx(0,6)`（6=COINIT_APARTMENTTHREADED|DISABLE_OLE1DDE）；
  hr∈{S_OK(0),S_FALSE(1)} → `defer CoUninitialize`；`defer UnlockOSThread`；fn() 结果透传。
- **resolveCmdExePath**（0x1408af1c0, 224B）：`os.Getenv("ComSpec")+TrimSpace` 非空即返；
  否则 `os.Getenv("SystemRoot")+TrimSpace` 非空 → `filepath.Join(...,"System32","cmd.exe")`；
  兜底 `"cmd.exe"`。
- **utf16PtrOrNil**（0x1408af2a0, 128B）：`TrimSpace` 空 → `(nil,nil)`；否则
  `UTF16FromString` 返 `&u[0]`。
- **quoteCmdArgument**（0x1408af320, 128B）：`TrimSpace` → `Replace(s,`"`,`""`,-1)`（CMD
  引号转义）→ `concatstring3(`"`,s,`"`)`。
- **startPowerShellEncodedCommand**（0x1408af500, 224B）：`encodePowerShellCommand` →
  `startDetachedCommand(["powershell.exe","-NoProfile","-Sta","-EncodedCommand",encoded])`。
- **encodePowerShellCommand**（0x1408af5e0, 320B）：`utf16.Encode([]rune(cmd))`（**不含** NUL
  终止，实证 `unicode/utf16.Encode` @0x1408af612）→ `make([]byte,2*len)` 小端写入 →
  `base64.StdEncoding.EncodeToString`。

## G4 测试

`backend/shell_execute_windows_test.go` 5 用例：

- `TestUTF16PtrOrNil`：空/全空白 → nil，非空 → UTF16 内容 `abc`。
- `TestQuoteCmdArgument`：`abc`→`"abc"`、`a"b`→`"a""b"`、空白 Trim、空 → `""`。
- `TestResolveCmdExePath`：ComSpec 命中 / SystemRoot 拼接 / cmd.exe 兜底三态。
- `TestEncodePowerShellCommand`：base64 解码回 UTF-16LE 字节 `[97 0 98 0 99 0]`（不含 NUL）。
- `TestWithShellApartmentPassthrough`：nil 与 error 透传（真实 LockOSThread+CoInitializeEx）。

全 PASS。

## 判定

四路 PASS，批次 107 闭环。Shell 执行域 8 函数全 `[S]`，无 `[S-sig]` 新增。
`resolveCmdExePath` 环境变量名实测为 `"ComSpec"`（非全大写），`encodePowerShellCommand`
实测经 `unicode/utf16.Encode` 不含 NUL 终止（与 PowerShell 编码约定一致），均已在代码注释标注。
