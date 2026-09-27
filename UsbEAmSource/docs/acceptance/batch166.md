# 批次 166 验收（launcherconfig_runtime.go：normalizeConsoleItem [P]→[S] + 两个 helper [S-sig]→[S]）

日期：2026-09-25
子批次：控制台条目归一化（3648B 汇编全量还原）

## 目标

`normalizeConsoleItem`（0x14087f560，3648 字节）从 `[P]` 升 `[S]`；顺带落地其直接依赖的
`normalizeConsoleItemKind`（0x1408803a0）与 `trimStringLimit`（0x140880920），两者 `[S-sig]`→`[S]`，
其中 `trimStringLimit` 签名由 `(s string) string` 修正为 `(s string, limit int) string`。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go test ./backend` | ok (1.7s) |
| 黄金用例 | `TestNormalizeConsoleItem` | PASS |

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2659 / S=1095 / S-inline=35 / S-sig=1345 / P=184 / UNMARKED=0`。

**真函数 = 1095 + 35 + 1345 = 2475 / 4754 = 52.1%**（相对批次 165 的 2474 增 +1，P 185→184）。

重建产物 SHA256：`B49502CAB63963E4E531AD9632A5B33C2DA17D0BBF3DBED122AAF3BF0A60A16C`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细（3 函数，`backend/launcherconfig_runtime.go`）

| 函数 | VA | 语义 |
|---|---|---|
| normalizeConsoleItem | 0x14087f560 | Kind 白名单归一化、window/mouse 特殊条目 canonical 重映射、内容清空+Icon 覆写、布局夹取、ID 重建 |
| normalizeConsoleItemKind | 0x1408803a0 | TrimSpace 后按长度跳表做大小写敏感白名单匹配（12 项），命中返 TrimSpace 原串，否则 "" |
| trimStringLimit | 0x140880920 | TrimSpace；空或 limit<=0 → ""；rune 数<=limit → 原串；否则截前 limit 个 rune |

## G4 关键实证结论

1. **normalizeConsoleItemKind 白名单**：长度跳表（len 7..20，len 12/15 直落返 ""）12 项
   `section/appGroup/audioCard/qrCodeTool/catalogItem/desktopWidget/screensTool/mouseGestureTool/
   memoryReleaseMode/gpuPreferenceEntry/oledBlackoutProfile/windowManagementTool`；大小写敏感
   （逐字节 cmp/memequal 实证）；命中走 0x140880592 返 TrimSpace 原串，未命中 0x1408805ab 返 ""。
2. **trimStringLimit 截断单位**：`runtime.countrunes` 比较 + `runtime.stringtoslicerune`/`slicerunetostring`
   成对，实证按 **rune**（非字节）截断，`<=0` 返 ""。
3. **有效 target 回落**：targetID 经 trimStringLimit 后为空时，分类用有效 target = sectionID
   （汇编 0x14087f60a `rcx=[rsp+0x160]` 回落）；但最终 targetID 仍取原始值（0x14087f747 `rax` 未回落），
   只有命中重映射才被 canonical 覆盖。
4. **canonical 重映射四分支**：①`section`+有效 target `windowManagement` → windowManagementTool；
   ②`windowManagementTool`+section `windowManagement`+target≠`windowTools` → 同 canonical；
   ③`section`+有效 target `mouseGestures` → mouseGestureTool；
   ④`mouseGestureTool`+section `mouseGestures`+target≠`gestures` → 同 canonical。
5. **内容清空**：canonical windowManagementTool 条目 → Icon="window"（0x140C376AA 6 字节）并清空
   SourceID/FolderPath/Title/Subtitle/IconData/IconRef/IconURL；canonical mouseGestureTool →
   Icon="pointer"（0x140C4942D 7 字节）并清空同组字段；其余条目内容保留。
6. **布局夹取**：LayoutX/W 上限 0x40、LayoutY/H 上限 0x200，`<=0`（有符号）→ 0（duffzero 预清零），
   即 `min(max(v,0),cap)`。
7. **早期退出**：kind/section/targetID 任一空 → 返零值 ConsoleItem（duffzero+0x11d 清输出区 0xe0 字节）。
8. **ID 重建**：结果按值 duffcopy 到栈，调 `buildConsoleItemID`，返回值写入 ID 字段。

## 残留 / 未落地（不触及）

- `normalizeConsoleIconData`（0x1408808a0）、`normalizeBookmarkFolderPathValue`（0x140883d00）
  仍为 [S-sig] 返 ""，本函数直接调用，行为等价依赖其后续落地（当前 IconData/FolderPath 归一化结果恒为 ""）。
- `normalizeConsoleItems` 仍为 [P] 存根（本批次未触碰）。
