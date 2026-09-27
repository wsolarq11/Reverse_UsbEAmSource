# 批次 102 验收（文件剪贴板链：路径规范化 + dropEffect 解析 + 编排）

日期：2026-09-23
子批次：pathclip（CF_HDROP 文件剪贴板链前置纯逻辑）
目标：落地文件剪贴板链的前置纯逻辑 `normalizePathList` / `resolveClipboardDropEffect` /
`setFileClipboard`，校正 `SetFileClipboard` 方法签名证伪（缺 `dropEffect string` 参数），
引入 `setWindowsFileClipboard` 骨架（`[S-sig]`，下一批补体）。新建
`backend/pathclip_clipboard_windows.go`。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| test | `go test -tags production -count=1 -p=1 -run 'TestNormalizePathList|TestResolveClipboardDropEffect|TestSetFileClipboard' ./backend` | EXIT=0（6 用例全 PASS） |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1155 / MARKED=1155 / S=844 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。

相对批次 101（1152/1152/841/35/275/1/0）：+3 新函数（`normalizePathList` /
`resolveClipboardDropEffect` / `setWindowsFileClipboard`），`setFileClipboard` 由
`[S-sig]` 补体转 `[S]`。`UNMARKED=0` 保持，`P=1` 不变。

## G3 逻辑等价

关键实证（asm VA → Go，字符串经 `resolve_lea_strings.py` / `read_gostring.py` 解码）：

- `normalizePathList`（0x14088c420, 704B）：入口 3 字 `[]string`；`test rbx,rbx` 空输入
  提前返回 nil（`lea [rip+0x13ce48d]` = empty slice）；`makeslice(cap=len(paths))` +
  `makemap(map[string]struct{}, len(paths))`（`test byte ptr [rax],al` = 零大小 value 写入实证）；
  循环内 `strings.TrimSpace`（空→continue）→ `internal/filepathlite.Clean` →
  `strings.ToLower`（去重键）→ `mapaccess2_faststr`（存在→continue）→ `mapassign_faststr`
  → `growslice`+append。返回 `[]string`（3 字，无 error 返回值实证）。
- `resolveClipboardDropEffect`（0x1408adbe0, 224B）：TrimSpace→ToLower 后按长度+常量比较：
  len=3 且 `"cut"`（`0x7563`+`0x74`）→ eax=2；len=4 且 `"copy"`（`0x79706f63`）→ eax=1；
  len=4 且 `"move"`（`0x65766f6d`）→ eax=2；否则 `errors.New("不支持的剪贴板文件操作")`
  （@0x140c73be2, 33B）。
- `setFileClipboard`（0x1408a9f00, 256B）：`normalizePathList` 空结果 →
  `errors.New("文件路径不能为空")`（@0x140c64ea8, 24B）；`resolveClipboardDropEffect`
  失败透传 error（rax←rbx/rbx←rcx）；否则 `setWindowsFileClipboard(normalized, effect)`
  返回值直接透传（call 后 add/pop/ret 无改写）。
- 签名证伪纠正：`BootstrapService.SetFileClipboard` 方法转发 5 字
  （`mov rax,rbx; rbx,rcx; rcx,rdi; rdi,rsi; rsi,r8`）= `(paths []string, dropEffect string)`，
  非旧 stub 的 `(paths []string)` 单参。

## G4 测试

新建 `backend/pathclip_clipboard_windows_test.go`（6 用例，纯逻辑、不触 Windows API）：

- `TestNormalizePathListEmpty`：nil / 空切片 → nil。
- `TestNormalizePathList`：TrimSpace + Clean + ToLower 去重 + 空跳过。
- `TestResolveClipboardDropEffect`：copy/COPY/空格 copy/cut/move/Move → 正确 effect；
  link/空/unknown → `不支持的剪贴板文件操作`。
- `TestSetFileClipboardEmptyPaths`：nil / 全空路径 → `文件路径不能为空`。
- `TestSetFileClipboardBadEffect`：link → `不支持的剪贴板文件操作`。
- `TestSetFileClipboardOK`：合法路径+copy → 编排成功（setWindowsFileClipboard 为 stub）。

全 PASS。

## 判定

四路 PASS，批次 102 闭环。遗留：`setWindowsFileClipboard` 为 `[S-sig]` 骨架
（`openClipboardForFileOperation` / `buildHDropClipboardData` / `buildDropEffectClipboardData`
链下一批补体后升级 `[S]`）。
