# 批次 125 验收（开机自启域：XML 定义辅助函数）

日期：2026-09-23
子批次：startup-task-xml-helpers

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.377s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1221 / MARKED=1221 / S=912 / S-inline=34 / S-sig=274 / P=1 / UNMARKED=0`。

相对批次 124（1216/907/34/274/1）：FUNCS +5，S +5，S-sig 持平。
（5 个新 [S]：launcherStartupTaskLogonDelay、launcherStartupTaskWorkingDirectory、
quoteWindowsTaskActionCommand、encodeUTF16LEWithBOM、currentLauncherStartupTaskUserID。）

## G3 逻辑等价

新文件 `backend/startup_task_xml_windows.go`，为 buildLauncherStartupTaskDefinitionXML 的依赖链：

- **launcherStartupTaskLogonDelay**（0x1408b24c0, 128B）[S]：clamp（≤0→10，>100→100），
  fmt.Sprintf("PT%dS")，格式串 "PT%dS" 明文 @0x140c35d1c。
- **launcherStartupTaskWorkingDirectory**（0x1408b2540, 192B）[S]：filepath.Dir(exe)；TrimSpace
  空或 dir=="."→filepath.Dir(os.Args[0])（全局 []string[0] 推断为 os.Args）；否则 dir。
- **quoteWindowsTaskActionCommand**（0x1408b2960, 256B）[S]：TrimSpace 空→""；首尾引号→原样；
  否则 `"`+Replace(`"`→`\"`,−1)+`"`。
- **encodeUTF16LEWithBOM**（0x1408b2600, 320B）[S]：utf16.Encode([]rune)，前导 0xff 0xfe BOM，
  逐 uint16 小端复制（make([]byte,2,len*2+2) + growslice 扩容语义）。
- **currentLauncherStartupTaskUserID**（0x1408b2a60, 192B）[S]：os/user.Current()→Uid 非空→Uid；
  否则 Username 非空→Username；都空→errors.New(30B 混淆 @0x140c2d335 [P] 待取证)。

## G4 复验

新增 `startup_task_xml_windows_test.go`：logon delay 边界（0/−5/5/50/100/150）、工作目录
（绝对/相对回退）、命令引号（空/已引/加引/转义）、UTF16LE BOM（单字符/空串）、用户 ID 非空
断言，全绿。

## 判定

四路 PASS，批次 125 闭环。5 个辅助函数 100% [S]（仅 currentLauncherStartupTaskUserID 一处
30B 混淆错误文案 [P]）。遗留：buildLauncherStartupTaskDefinitionXML + convertXMLFileToUTF16LE
主函数（依赖本批 + 混淆字符串 + XML 结构体填充）下批落地。
