# 批次 161 验收（launcherconfig_runtime.go 启动日志规范化链：3 函数 [P]→[S]）

日期：2026-09-25
子批次：开始菜单启动日志（StartMenuLaunchLog）键提取 / 单条归一化 / 合并

## 目标

把 launcherconfig_runtime.go 中启动日志域的 3 个 `[P]` 存根升级为 `[S]`（asm 逐条对位），
并修正 `mergeStartMenuLaunchLog` 的错误签名（原 `([]StartMenuLaunchLog, StartMenuLaunchLog) []StartMenuLaunchLog`
→ 实测 `(StartMenuLaunchLog, StartMenuLaunchLog) StartMenuLaunchLog`）：

`startMenuLaunchLogKeys`、`normalizeStartMenuLaunchLog`、`mergeStartMenuLaunchLog`。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go1.25.12 test -count=1 -p=1 ./backend` | ok (0.433s) |
| 黄金用例 | `go1.25.12 test -run 'TestStartMenuLaunchLogKeys\|TestNormalizeStartMenuLaunchLog\|TestMergeStartMenuLaunchLog' -v` | 3 组全 PASS |

黄金用例说明：三函数均为纯字符串/结构体逻辑（ToLower/TrimSpace/TrimPrefix/合并），无 Win32/DB/runtime
I/O，可全路径安全实测（含 swap 分支与 LaunchCount max 的 asm 实证边界）。

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2659 / S=1086 / S-inline=35 / S-sig=1348 / P=190 / UNMARKED=0`。

**真函数 = 1086 + 35 + 1348 = 2469 / 4754 = 51.9%**（相对批次 160 的 2466 增 +3 [S]，P 193→190）。

重建产物 SHA256：`EB12D08562A636D4ED75E095FB96567234EF91B447311779AF89DF851BC18044`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细（3 函数 [S]，`backend/launcherconfig_runtime.go`）

| 函数 | VA | 语义 |
|---|---|---|
| startMenuLaunchLogKeys | 0x140881700 | 非空字段按序 `id:`/`shortcut:`/`target:` 前缀 + ToLower 拼接，cap=3 预分配 |
| normalizeStartMenuLaunchLog | 0x140881400 | 全字段 TrimSpace + ID 去 `start-menu:` 前缀 + LaunchCount 夹取非负 + 双重空检返零值 |
| mergeStartMenuLaunchLog | 0x140881940 | LaunchCount max → 较晚者为基底 → 空 ID/Name/ShortcutPath/TargetPath 回填 |

## G4 关键实证结论

1. **startMenuLaunchLogKeys 三前缀实证**：`makeslice(0,3)` 预分配，三字段判空后 `concatstring2(前缀, ToLower(field))`；
   前缀 read_gostring 解码 `id:`(3B)/`shortcut:`(9B)/`target:`(7B)；返 (ptr,len,cap) 3 字 slice（原 [P] 误记 4 寄存器）。
2. **normalizeStartMenuLaunchLog 11B 前缀**：`cmp rbx,0xb` + memequal(0x140c4754f, 11) 解码 `start-menu:`，
   TrimPrefix 语义（`rbx-0xb` + 符号扩展夹取 ptr 偏移）；LaunchCount `test/cmovl` 夹取负值为 0；
   双重空检——首检 (ID/ShortcutPath/TargetPath 皆空) 与次检 (LaunchCount<=0 且 LastLaunchedAt 空)，
   皆 duffzero 返零态（丢弃"无身份"或"未启动"条目）。
3. **mergeStartMenuLaunchLog 签名修正**：两 88B StartMenuLaunchLog 栈传（S1@[0xe0]、S2@[0x138]），
   返单 88B 结构（[0x190]），非切片；调用点 normalizeStartMenuLaunchLogs 0x140880d93 双结构栈拷贝实证。
4. **merge 的 LaunchCount max 读原第二参**：cmpstring(B.LastLaunchedAt,A.LastLaunchedAt)>0 触发整结构 swap；
   LaunchCount max 读 [0x178]（原第二参计数），故"incoming 较新"分支保留 incoming 计数（黄金用例断言 =2）。
5. **LastLaunchedAt 尾 max 冗余**：swap 已保证较晚者为基底，尾 cmpstring 对 LastLaunchedAt 的 max 恒为空操作，
   不落地（功能等价）。

## 残留 [P] / 未落地（不触及）

launcherconfig_runtime.go 剩余 [P]：populateAppIcons、loadOrCreateLauncherConfig、buildConsoleItemID、
loadLauncherConfigOrDefaultIfMissing、normalizeBackgroundPreference、normalizeTagCatalogWithDefault、
normalizeFileLocatorConfig、readLauncherConfigFile、normalizeConsoleItem（9 个，多为加载/合并链）。
另 `normalizeStartMenuLaunchLogs`（0x140880a40，0x7c0）为 [S-sig]，是本批三函数的调用方，可下一批落地。

## 已知偏差（诚实记录）

mergeStartMenuLaunchLog 的 LaunchCount 在"incoming 较新"分支按 asm 保留 incoming 计数（读原第二参 [0x178]），
而非语义上的 max(existing,incoming)——此为原程序实际行为，忠实保留并黄金用例锁定。
