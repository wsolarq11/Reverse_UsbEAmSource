# 批次 104 验收（CF_HDROP 编排侧）

日期：2026-09-23
子批次：pathclip（CF_HDROP 编排：OpenClipboard 重试 + 格式注册 + 提交链）
目标：落地 `openClipboardForFileOperation` / `registerPreferredDropEffectClipboardFormat`，
补体 `setWindowsFileClipboard`（由 `[S-sig]` 升级 `[S]`）。至此 `setFileClipboard` 文件剪贴板
全链（normalize→resolve→encode→build→open→submit）全部 `[S]`。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| test | `go test -tags production -count=1 -p=1 -run 'TestRegisterPreferredDropEffectClipboardFormat|TestSetWindowsFileClipboard' ./backend` | EXIT=0（2 用例全 PASS） |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1160 / MARKED=1160 / S=850 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。

相对批次 103（1158/1158/847/35/275/1/0）：+2 新函数（`openClipboardForFileOperation` /
`registerPreferredDropEffectClipboardFormat`），`setWindowsFileClipboard` 由 `[S-sig]` 补体转
`[S]`。`UNMARKED=0` 保持。

## G3 逻辑等价

关键实证（asm VA → Go，字符串经 resolve_lea_strings / read_gostring 解码）：

- `openClipboardForFileOperation`（0x1408ae240, 224B）：循环 `cmp rax,0x64`（100 次）、
  `i>0` 时 `time.Sleep(0x989680=10ms)`；proc 全局变量 `[rip+0x13137cc]` 与循环后
  `[rip+0x1313796]` 计算后均为 `0x140c01a58`（同一 OpenClipboard LazyProc），`Call(0)` 即
  `OpenClipboard(0)`；100 次失败后最后再试一次，仍失败 →
  `wrapWinError("打开剪贴板失败")`（@0x140c5f803, 21B）。
- `registerPreferredDropEffectClipboardFormat`（0x1408aeb60, 352B）：
  `UTF16FromString("Preferred DropEffect")`（@0x140c5d9b7, 20B）→ `RegisterClipboardFormatW`；
  返回 0 → `wrapWinError("注册剪贴板文件操作格式失败")`（@0x140c7de26, 39B）。
- `setWindowsFileClipboard`（0x1408adcc0, 640B）编排：
  `LockOSThread` + defer UnlockOSThread（闭包 @0x141096e10）→ `openClipboardForFileOperation`
  失败透传 → defer CloseClipboard（func1 0x1408ae200）→ `EmptyClipboard()`（`[rip+0x1313cef]`
  =0x140c01a68）失败 `wrapWinError("清空剪贴板失败")`（@0x140c5f7d9, 21B）→
  `buildHDropClipboardData` / `buildDropEffectClipboardData` 失败透传 →
  `SetClipboardData(0xf=CF_HDROP, hDrop)`（`[rip+0x1313c11]`=0x140c01a70）失败
  `wrapWinError("写入文件剪贴板失败")`（@0x140c69fc9, 27B）→ 成功 `hDrop=0`（defer GlobalFree(0)
  无害）→ `registerPreferredDropEffectClipboardFormat` → `SetClipboardData(format, hEffect)` 失败
  `wrapWinError("写入剪贴板文件操作失败")`（@0x140c73c03, 33B）→ 成功 `hEffect=0` →
  显式 `CloseClipboard()`（`[rip+0x1313b76]`=0x140c01a60）失败 `wrapWinError("关闭剪贴板失败")`
  （@0x140c5f7ee, 21B）。

## G4 测试

新建 `backend/pathclip_submit_windows_test.go`（2 用例）：

- `TestRegisterPreferredDropEffectClipboardFormat`：注册格式 ≥0xC000 且非 0。
- `TestSetWindowsFileClipboard`：临时文件提交 CF_HDROP；剪贴板被占用时 `t.Skip`
  （openClipboard 重试失败返回 `打开剪贴板失败`），否则验证成功。

全 PASS（本机剪贴板空闲，提交成功路径实测通过）。

## 判定

四路 PASS，批次 104 闭环。`setFileClipboard` 文件剪贴板全链 `[S]`，无遗留 stub。
