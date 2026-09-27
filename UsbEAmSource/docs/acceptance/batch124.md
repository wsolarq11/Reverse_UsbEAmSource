# 批次 124 验收（开机自启域：命令输出解码 + 任务缺失判定）

日期：2026-09-23
子批次：startup-task-decode-domain

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.466s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1216 / MARKED=1216 / S=907 / S-inline=34 / S-sig=274 / P=1 / UNMARKED=0`。

相对批次 123（1213/904/34/274/1）：FUNCS +3，S +3，S-sig 持平。
（3 个新 [S]：decodeWindowsBytes、decodeWindowsCommandOutput、isTaskNotFoundMessage。）

## G3 逻辑等价

新文件 `backend/startup_task_windows.go`，为 deleteLauncherStartupTask 的 schtasks 回退链底层：

- **decodeWindowsBytes**（0x1408b2c40, 464B）[S]：len==0/codepage==0→("",false)；
  MultiByteToWideChar 两段式（取长度→分配→填充）失败→("",false)；utf16.decode→slicerunetostring
  →(string,true)。签名 `(data []byte, codePage uint32) (string, bool)`。
- **decodeWindowsCommandOutput**（0x1408b2b20, 288B）[S]：空→""；utf8.Valid→TrimSpace 原样；
  否则 decodeWindowsBytes(GetACP)→TrimSpace；再失败 decodeWindowsBytes(936=GBK)→TrimSpace；
  再失败原文转 string TrimSpace。三级回退与 asm 一致。
- **isTaskNotFoundMessage**（0x1408b2e60, 288B）[S]：TrimSpace+ToLower 后依序查 6 个子串
  （stringslite.Index >= 0）。第 1 个 "cannot find"（11B @0x140c47591）明文 [S]；后 5 个
  为混淆字节（9B/30B/9B/9B/30B @0x140c3d7cc/75153/3d7d5/3d7de/75171），按原始字节还原并
  标注 [P] 待取证明文。

## G4 复验

新增 `startup_task_windows_test.go`：空输入、UTF-8 TrimSpace、GBK 解码（"找不到"）、
cannot find 命中（含大小写归一化）、负例、ToLower 路径，全绿。

## 判定

四路 PASS，批次 124 闭环。解码链 2 函数 100% [S]，任务缺失判定逻辑结构 [S]（5 混淆子串
[P] 待取证）。遗留：开机自启域 schtasks 混淆字符串（任务名/命令名/格式串/判定子串）待
统一破解混淆机制。
