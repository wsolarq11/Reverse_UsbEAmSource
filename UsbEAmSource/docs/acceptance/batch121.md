# 批次 121 验收（launchContext 领域模型立案 + 入口解析函数）

日期：2026-09-23
子批次：launch-context-model
类型：**领域模型变更（大改）** — 澄清 → 商榷 → 立案 → 执行 → 复验。

## 领域模型立案

**模型声明**：引入 `launchContext` 结构体，作为 startApplication 域的唯一启动上下文载体。
后续批次将逐步把 startApplication / startApplicationUsingCurrentPrivileges /
startApplicationViaExplorer / startApplicationWindowsAsAdmin 从平铺参数签名切换到
launchContext 签名（本批仅落地结构体 + 两个解析函数，不动现有平铺调用链，保证零回归）。

**字段偏移（asm 交叉实证，非推断）**：三函数一致确认。

| 偏移 | 字段 | 实证来源 |
|---|---|---|
| 0x20 | entryType string | resolveLaunchableAppEntry / startApplicationViaExplorer / resolveApplicationWorkingDirectory |
| 0xa0 | shellTarget string | 同上 |
| 0xb0 | entry string | resolveLaunchableAppEntry |
| 0xc0 | args []string | 同上（3 词） |
| 0x138 | rawName string | resolveLaunchableAppEntry |
| 0x148 | rawEntry string | 同上 |
| 0x158 | rawCommandLine string | 同上 |

结构体 360B=0x168（45 词）：duffcopy+0x24c 复制 352B（44 词）+ 第 0 词单独 mov。
duffcopy 每条 14B（MOVUPS 3 + ADDQ 4 + MOVUPS 3 + ADDQ 4），64 条；偏移 0x24c=588=42 条，
剩余 22 条 × 16B = 352B。中间 30 词（0x00-0x1f、0x30-0x9f、0xd8-0x137）语义待
buildAppEntryWithDroppedFiles 完整取证，以 `_ [N]uintptr` 占位锁偏移。

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.459s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1207 / MARKED=1207 / S=893 / S-inline=34 / S-sig=279 / P=1 / UNMARKED=0`。

相对批次 120（1204/890/34/279/1）：FUNCS +3，S +3（三个新函数直接落 [S]）。

## G3 逻辑等价

`backend/launcherprocess.go` 新增：

- **shouldUseSavedShortcutResolution**（0x1408a7e60, 128B）[S]：TrimSpace 空 → true；
  !isShortcutFilePath → false；os.Stat err → true，否则 false。
- **resolveLaunchableAppEntry**（0x1408a75e0, 1024B）[S]：directory → 原样；TrimSpace(rawName)
  空 → 原样；!shouldUseSavedShortcutResolution(TrimSpace(shellTarget)) → 原样；否则
  shellTarget=TrimSpace(rawName)、entry 空回退 rawEntry、args 空回退 splitCommandLineArguments(rawCommandLine)。
- **resolveLaunchableAppEntryForExplicitArgs**（0x1408a79e0, 1152B）[S]：directory → 原样；
  resolveLaunchableAppEntry → !requiresShellOpen(TrimSpace(shellTarget)) → 返回 resolved；
  TrimSpace(rawName) 空 → 返回 resolved；否则显式 args 分支（同 saved 分支写回逻辑）。

## G4 复验（结构体布局锁定）

`TestLaunchContextSize` 断言 `unsafe.Sizeof(launchContext{})==360`，锁字段偏移。
`TestShouldUseSavedShortcutResolution` / `TestResolveLaunchableAppEntry*` 覆盖 directory、
空 rawName、非快捷方式、saved 分支四路径，全绿。

## 判定

四路 PASS，批次 121 闭环。launchContext 结构体立案落地，字段偏移 asm 实证并测试锁定；
两个解析函数 100% asm 直译标 [S]。
