# 批次 103 验收（CF_HDROP 数据构建链）

日期：2026-09-23
子批次：pathclip（CF_HDROP 数据构建：UTF-16 路径编码 + DROPFILES/DropEffect 内存块）
目标：落地 `encodeHDropPaths` / `buildHDropClipboardData` / `buildDropEffectClipboardData`
三个数据构建函数（依赖批 100 的 `globalAllocMoveable` / `globalLock` / `wrapWinError`），
完成 CF_HDROP 链的数据侧。`setWindowsFileClipboard`（编排 + LazyProc）仍为 `[S-sig]`，下一批补体。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| test | `go test -tags production -count=1 -p=1 -run 'TestEncodeHDropPaths|TestBuildHDropClipboardData|TestBuildDropEffectClipboardData' ./backend` | EXIT=0（4 用例全 PASS） |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1158 / MARKED=1158 / S=847 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。

相对批次 102（1155/1155/844/35/275/1/0）：+3 新函数（`encodeHDropPaths` /
`buildHDropClipboardData` / `buildDropEffectClipboardData`），全部 `[S]`。`UNMARKED=0` 保持。

## G3 逻辑等价

关键实证（asm VA → Go）：

- `encodeHDropPaths`（0x1408ae7c0, 544B）：入口 3 字 `[]string`；`makeslice([]uint16, 0, 0x100)`
  （cap=256）；逐路径 `strings.TrimSpace`（空→continue）→ `syscall.UTF16FromString`（含 null，
  `memmove` 拷贝 u16.len*2 字节即含 null 终止）→ `growslice`+append；UTF16FromString 错误透传
  （rax/rbx/rcx 清零、rdi/rsi 保留 error）；循环后空 buf → `errors.New("文件路径不能为空")`
  （@0x140c64ea8 复用, 24B）；否则 `mov word [rax+rdx*2-2],0` 补双 null 终止。返回 `([]uint16, error)`。
- `buildHDropClipboardData`（0x1408ae320, 608B）：`encodeHDropPaths` 失败透传；
  `globalAllocMoveable(len(u16)*2 + 0x14)`（`lea rax,[rbx+rbx]; lea rax,[rax+0x14]` 实证）；
  `globalLock` 失败 → `LazyProc.Call(GlobalFree)` 释放后透传（`[rip+0x13135e1]` = GlobalFree proc）；
  `defer GlobalUnlock`（deferprocStack + func1）；写 DROPFILES 头：`pFiles=0x14`（`mov dword
  [rsp+0x24],0x14`）、`pt=(0,0)`、`fNC=0`、`fWide=1`（`mov dword [rsp+0x34],1`），
  `movups` 16 字节头写入 ptr、`memmove(ptr+0x14, u16, len*2)` 写路径块。返回 `(w32.HGLOBAL, error)`。
- `buildDropEffectClipboardData`（0x1408ae5e0, 384B）：入口 `eax=dropEffect`（1 字）；
  `globalAllocMoveable(4)` → `globalLock`（失败 `GlobalFree` 后透传）→ `defer GlobalUnlock`；
  `mov dword [rax],ebx` 写 dropEffect。返回 `(w32.HGLOBAL, error)`。

## G4 测试

新建 `backend/pathclip_hdrop_windows_test.go`（4 用例，GlobalAlloc/GlobalLock 在 Windows 测试环境可用）：

- `TestEncodeHDropPaths`：多路径（含空格包裹、空路径跳过）→ UTF-16 编码含两路径、双 null 终止。
- `TestEncodeHDropPathsEmpty`：nil / 全空白 → `文件路径不能为空`。
- `TestBuildHDropClipboardData`：GlobalLock 读回验证 `pFiles=20`、`fWide=1`、路径数据在偏移 20。
- `TestBuildDropEffectClipboardData`：GlobalLock 读回验证 4 字节 = dropEffectCopy。

全 PASS（首次 `TestEncodeHDropPaths` 因测试断言用 `UTF16ToString` 遇首个 `\0` 截断而误报，修正断言
为 `\0`→`|` 占位后通过；实现本身无变更）。

## 判定

四路 PASS，批次 103 闭环。遗留：`setWindowsFileClipboard` / `openClipboardForFileOperation` /
`registerPreferredDropEffectClipboardFormat`（编排 + LazyProc 提交）为下一批补体，之后
`setFileClipboard` 全链 `[S]`。
