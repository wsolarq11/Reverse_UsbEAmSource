# 批次 163 验收（launcherconfig_runtime.go：buildConsoleItemID + normalizeBackgroundPreference [P]→[S]）

日期：2026-09-25
子批次：控制台条目 ID 构造 + 背景偏好归一化（均为纯函数，确定性）

## 目标

把两个 `[P]` 存根升级为 `[S]` 真函数：
1. `buildConsoleItemID(ConsoleItem) string` —— 构造控制台条目复合 ID。
2. `normalizeBackgroundPreference(BackgroundPreference) BackgroundPreference` —— 归一化背景偏好。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go1.25.12 test -count=1 -p=1 ./backend` | ok (0.427s) |
| 黄金用例 | `TestBuildConsoleItemID` / `TestNormalizeBackgroundPreference` | PASS |

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2659 / S=1089 / S-inline=35 / S-sig=1348 / P=187 / UNMARKED=0`。

**真函数 = 1089 + 35 + 1348 = 2472 / 4754 = 52.0%**（相对批次 162 的 2470 增 +2 [S]，P 189→187）。

重建产物 SHA256：`B101CE67B2692CB4F4027CA045CA873841D9F5F2586604DF29BE61F320141446`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细（2 函数 [S]，`backend/launcherconfig_runtime.go`）

| 函数 | VA | 语义 |
|---|---|---|
| buildConsoleItemID | 0x1408805e0 | Kind/SectionID/TargetID 任一空→""；[Kind,SectionID,SourceID,TargetID,FolderPath] 各段 PathEscape 后 Join ":" |
| normalizeBackgroundPreference | 0x14087eae0 | 字符串 TrimSpace、浮点 clamp、nil 回退默认、ShellOpacity 迁移语义 |

## G4 关键实证结论

1. **buildConsoleItemID（154 行 asm）**：栈传 `ConsoleItem`（0xe0），读 5 字符串（偏移 0x10/0x20/0x40/0x30/0x50
   = Kind/SectionID/SourceID/TargetID/FolderPath）；零值门控 `[rsp+0x120/0x130/0x140]`（Kind/SectionID/TargetID 长度）；
   `make([]string,0,5)` → 逐段 `net/url.escape(s, mode=2)`（encodePathSegment=url.PathEscape）→
   `strings.Join(":")`。分隔符 ":" 经 va_dump 实解（0x1411CAC58 len=1 → `:`）。
2. **normalizeBackgroundPreference（202 行 asm）**：输入 `BackgroundPreference`（0x70）栈传；输出 0x68 字段显式落、
   ShellOpacity(0x68) 不落（duffzero 归零 → nil）。float64 常量实解：1.0（通用上界+ImageOpacity 默认）、
   0.65（ReadabilityOverlayOpacity 上界）、0.18（其默认）；ImageBlur 上界 0x18=24。
3. **ShellOpacity 迁移语义**：`ShellOpacity` clamp[0,1] 后的值仅作为 `SidebarTransparency`/`ContentTransparency`
   的 nil 回退默认，其自身字段不写入输出（置 nil）——与「shellOpacity 拆分迁移为 sidebar/content」设计吻合。
4. **NaN 语义一致**：asm 用 `ucomisd`+`jbe`，NaN 不落入任何钳制分支（透传）；Go `if v<0 {…} else if v>hi {…}` 对 NaN 同样透传，语义等价。

## 残留 / 未落地（不触及）

`buildConsoleItemID` 的 slice 字面量顺序 [Kind,SectionID,SourceID,TargetID,FolderPath]（SourceID 先于 TargetID）
系按 asm slot 顺序（slot2=0x40=SourceID、slot3=0x30=TargetID）实证；该顺序依赖 types_app.go 中
`ConsoleItem` 的 TargetID(0x30)/SourceID(0x40) 布局为真。若未来复核 ConsoleItem 布局有变，需同步此顺序。
