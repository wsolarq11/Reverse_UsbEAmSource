# 批次 112 验收（进程创建链：命令行组装 + 提权错误判定）

日期：2026-09-23
子批次：process-launch
目标：落地 `buildCreateProcessCommandLine` [S] + `isElevationRequiredError` [S]
（openUnelevatedCommandLine 依赖链前两环）。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 定向 test | `go test -tags production -count=1 -p=1 -run 'BuildCreateProcessCommandLine|IsElevationRequiredError' ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1187 / MARKED=1187 / S=880 / S-inline=34 / S-sig=272 / P=1 / UNMARKED=0`。

相对批次 111（1185/878/34/272/1/0）：2 个新函数均 `[S]`。`UNMARKED=0` 保持。

## G3 逻辑等价

新文件 `backend/process_launch_windows.go`（2 函数）：

- **buildCreateProcessCommandLine**（0x1408acf00, 288B）：TrimSpace(exe)/TrimSpace(args)；
  exe 空 → 返回 args；否则 `Replace(exe,`"`,`\"`,-1)`（old=`"` 1B @0x1411cac88，new=`\"`
  2B @0x140c336b5）→ `concatstring3(nil,`"`,escaped,`"`)`；args 空 → 引号包裹 exe，否则
  `concatstring3(nil,quoted,` `,args)`（空格 1B @0x1411ca5c8）。
- **isElevationRequiredError**（0x1408ace20, 224B）：err nil → false；`errors.Is(err,
  windows.ERROR_ELEVATION_REQUIRED)`（target data 0x2e4=740 @0x1411cd5c0）真 → true；
  否则 `errors.As(err,&errno)` 且 errno==740 → true，否则 false。

## G4 测试

`backend/process_launch_windows_test.go` 2 用例：

- `TestBuildCreateProcessCommandLine`：5 例（纯 exe/带 args/exe 空返回 args/Trim/
  内部引号转义）。
- `TestIsElevationRequiredError`：5 例（nil/ERROR_ELEVATION_REQUIRED/Errno(740)/
  Errno(2)/普通 error）。

全 PASS。

## 判定

四路 PASS，批次 112 闭环。`buildCreateProcessCommandLine`（纯字符串，CreateProcess 命令行
组装）与 `isElevationRequiredError`（740 提权错误判定）全 `[S]`，为 openUnelevatedCommandLine
依赖链（下一环 createProcessWithShellParentWithVisibility）铺路。
