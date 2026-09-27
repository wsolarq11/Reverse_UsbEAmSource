# 批次 110 验收（路径定位目录解析 + 资源管理器打开目录）

日期：2026-09-23
子批次：pathclip-directory
目标：落地 `resolveOpenLocationDirectory` [S] + `openPathDirectory` [S-sig]→[S]。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 定向 test | `go test -tags production -count=1 -p=1 -run 'ResolveOpenLocationDirectory|OpenPathDirectoryEmpty' ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1184 / MARKED=1184 / S=876 / S-inline=34 / S-sig=273 / P=1 / UNMARKED=0`。

相对批次 109（1183/874/34/274/1/0）：1 个新函数 `resolveOpenLocationDirectory` `[S]`；
`openPathDirectory` [S-sig]→[S]（S +2，S-sig -1）。`UNMARKED=0` 保持。

## G3 逻辑等价

`bootstrap_pathclip.go`（2 函数）：

- **resolveOpenLocationDirectory**（0x1408a87c0, 224B）：TrimSpace 空 → ""；`os.Stat` 成功 →
  IsDir 真 → 返回 path，非目录 → `filepath.Dir`；Stat 失败 → `looksLikeFilesystemPath` 假 → ""，
  真 → `filepath.Dir`。三态实证：IsDir 方法经 itab 偏移 0x18 调用（@0x1408a882b），
  Dir 走 internal/filepathlite.Dir（@0x1408a8816/8852）。
- **openPathDirectory**（0x1408a88a0, 544B）：TrimSpace 空 → newPathError；`os.Stat` 成功 →
  IsDir 真 → `withShellApartment(shellExecuteProgram("", path, "", ""))`（func1 @0x1408a8a60），
  非目录 → `revealPathInExplorer`；Stat 失败 → resolveOpenLocationDirectory 空 → newPathError，
  非空 → `withShellApartment(shellExecuteProgram("", dir, "", ""))`（func2 @0x1408a8a00）。
  两个错误分支均内联 `errors.New("路径不能为空")`（18B 消息 @0x140c59e5e，类型 @0x140b54380）。

## G4 测试

`backend/bootstrap_pathclip_test.go` 2 用例：

- `TestResolveOpenLocationDirectory`：空/目录/文件/不存在含分隔符/不存在无分隔符 5 态
  （真实 t.TempDir + os.WriteFile）。
- `TestOpenPathDirectoryEmpty`：空路径 → newPathError（"路径不能为空"）。

全 PASS。

## 判定

四路 PASS，批次 110 闭环。`openPathDirectory` 完整落地（依赖链 6/6 齐：
os.Stat/filepath.Dir/looksLikeFilesystemPath/revealPathInExplorer/shellExecuteProgram/withShellApartment），
`resolveOpenLocationDirectory` 全 `[S]`。路径定位三态语义（目录返身/文件返父/不存在走 looksLikeFilesystemPath）
均经 asm 实证。
