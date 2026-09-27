# 批次 126 验收（开机自启域：XML 定义主函数 + 字符串混淆重大修正）

日期：2026-09-23
子批次：startup-task-xml-main

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.177s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1224 / MARKED=1224 / S=913 / S-inline=35 / S-sig=275 / P=1 / UNMARKED=0`。

相对批次 125（1221/912/34/274/1）：FUNCS +3，S +1，S-inline +1，S-sig +1。

新增：
- `buildLauncherStartupTaskDefinitionXML` 0x1408b2080 [S-sig]
- `convertXMLFileToUTF16LE` 0x1408b2740 [S]
- `intPtr` [S-inline]（对应 newobject + mov [rax],7 的 Settings.Priority=7 填充）

## 重大修正：本 EXE 无字符串混淆

此前多个批次将部分字符串标注为 `[P] 混淆待取证`，**结论错误**。根因：内联
`pefile.get_offset_from_rva(va-0x140000000)` 对 `VirtualSize < SizeOfRawData` 的节映射
错误，把 VA 错读到相邻的 runtime 字符串片段。改用 `read_gostring.py`（遍历节，用
`max(vsize, rawsz)` 判断）后，全部常量复核为**明文**：

| VA | len | 明文 |
|---|---|---|
| 0x140c784e3 | 36 | 开机启动程序路径不能为空 |
| 0x140c80e94 | 42 | 开机启动任务用户标识不能为空 |
| 0x140c35d1c | 5 | PT%dS |
| 0x140c7f142 | 40 | `<?xml version="1.0" encoding="UTF-16"?>\n` |
| 0x140c441cb | 10 | .utf16.xml |
| 0x140c6f135 | 30 | 无法获取当前用户标识 |
| 0x140c551fa | 16 | InteractiveToken |
| 0x140c33d0d | 3 | S4U |
| 0x140c3366b | 2 | `  `（两个空格） |

批次 125 的 `currentLauncherStartupTaskUserID` 30B 已由混淆字节改为明文
"无法获取当前用户标识"。

### 批次 123/124 遗留混淆字符串同步复核（本批修正）

| 位置 | 原错误标注 | 正确明文 |
|---|---|---|
| `startApplicationWithShellParent` 52B @0x140c89916 | 乱码 [P] | 无法通过安全的 Explorer Shell 启动该目标 |
| `isTaskNotFoundMessage` 9B @0x140c3f7cc | 乱码 [P] | not found |
| `isTaskNotFoundMessage` 30B @0x140c6f153 | 乱码 [P] | cannot find the file specified |
| `isTaskNotFoundMessage` 9B @0x140c3f7d5 | 乱码 [P] | 找不到 |
| `isTaskNotFoundMessage` 9B @0x140c3f7de | 乱码 [P] | 不存在 |
| `isTaskNotFoundMessage` 30B @0x140c6f171 | 乱码 [P] | 系统找不到指定的文件 |

根因一致：早前用 `pefile.get_offset_from_rva` 手算 lea 目标，`VirtualSize <
SizeOfRawData` 的节导致 RVA→offset 错位，读到相邻 runtime 字符串片段。两个文件已改回明文。

## G3 逻辑等价

`buildLauncherStartupTaskDefinitionXML`（[S-sig]）：
- TrimSpace(exe) 空 → "开机启动程序路径不能为空"；TrimSpace(userID) 空 →
  "开机启动任务用户标识不能为空"；
- quoteWindowsTaskActionCommand(exe) / launcherStartupTaskLogonDelay /
  launcherStartupTaskWorkingDirectory 依赖链已证；
- LogonType：enabled→"S4U"，否则 "InteractiveToken"（cmovne @0x1408b2215/34）；
- xml.MarshalIndent(def, "", "  ") + 前缀 39B 声明头。
- [P] 待取证：XML 结构体默认模板（duffcopy 源 @0x1411e4ea8）的 Version/Xmlns/
  Principal.ID/RunLevel/Actions.Context/Settings 各字段按 Task Scheduler 标准语义还原，
  未逐字段解引用。

`convertXMLFileToUTF16LE`（[S]）：ReadFile → bytes.Replace(UTF-8 声明→UTF-16 声明, 1)
→ path+".utf16.xml" → encodeUTF16LEWithBOM → WriteFile(0600) → 返回 (临时路径,
func(){os.Remove}, nil)。

## G4 复验

新增测试：空 exe/userID 校验错误文案、XML 构建（声明头/Task 根/PT5S/S4U/Command）、
禁用态 InteractiveToken、UTF16LE 转换（BOM + UTF-16 声明替换 + cleanup 删除）。

## 判定

四路 PASS，批次 126 闭环。**字符串混淆假设撤销**，后续所有字符串一律用
`read_gostring.py` 复核。遗留：批次 123/124 已标注 `[P]` 的混淆字符串需按正确
va_to_off 回读修正。
