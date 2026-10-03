# 批次 275 · desktopwidgets_notification_windows.go 整文件落地（+2 [S]，+2 FUNCS）

## 目标

1. **整文件落地 `desktopwidgets_notification_windows.go`**（42 个未落地文件差集之一）：
   桌面小部件提醒通知 2 函数全 `[S]`。
2. 订正依赖：`showLauncherNotificationWithAudio` 签名 2→3 参（实证缺 `sound` 第三参）。

## 基线 / 收口

| 指标 | 基线（batch 274 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2883 | 2885 |
| MARKED | 2883 | 2885 |
| S | 1363 | 1365 |
| S-inline | 36 | 36 |
| S-sig | 1443 | 1443 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1399 | 1401 |
| USABLE | 1399 | 1401 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build ./backend` EXIT=0；`go1.25.12 vet ./backend` EXIT=0；
`go1.25.12 test ./backend -run DesktopWidgetNotification` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1401  FUNCS=2885  MARKED=2885  P=41  S-eq=0  S-inline=36  S-sig=1443  S=1365  USABLE=1401
```

分项自洽：`S + S-inline + S-sig + P = 1365 + 36 + 1443 + 41 = 2885 = FUNCS`。
S 1363→1365（+2）、FUNCS 2883→2885（+2）、FAITHFUL 1399→1401（+2）、P=41/S-sig=1443/UNMARKED=0 保持。

账目：新建 `desktopwidgets_notification_windows.go` 2 函数 `[S]`（+2 S +2 FUNCS +2 FAITHFUL）。
1 个闭包 `normalizeDesktopWidgetNotificationText.func1`（映射折叠控制字符）内联于 `[S]` 函数内，
不计独立 FUNCS。

## G3 行为（asm 逐地址实证）

### 3.1 normalizeDesktopWidgetNotificationText [S 0x1407aca60, 288B]

`(text string, maxLen int, fallback string) string`（rax/rbx=text、rcx=maxLen、rdi/rsi=fallback）。

- TrimSpace（0x1407aca9a）→ strings.Map（映射闭包 @0x1409f7ba0）→ strings.Fields →
  strings.Join(sep=" " @0x1411CA5C8) → []rune → 截断 → 空串回退 fallback。
- 映射闭包（0x1409f7ba0，func1）：`r=='\r'/'\n'/'\t'/r<' '/r==0x7f` → ' '，否则原样（ASCII 控制字符折叠空格）。
- 截断：`len(runes) > maxLen` → `runes[:maxLen]`（0x1407acae0 cmp len/maxLen，越界 panicSliceAcap 保护）；
  `len==0` → 返回 fallback；否则 `string(runes)`（slicerunetostring）。

### 3.2 showDesktopWidgetNotification [S 0x1407ac960, 256B]

`(title, body string, silent bool)`，无返回值（morestack 保护 rax/rbx/rcx/rdi/sil）。

- title → `normalize(title, 120, "UsbEAm Launcher")`（fallback 15B @0x140C52CBF）；
- body → `normalize(body, 240, "提醒时间已到")`（fallback 18B UTF-8 @0x140C59C9C）；
- sound 选择：`silent=true` → "silent"(6B)，否则 "ms-winsoundevent:Notification.Default"(37B)
  （cmovne 双路，常量 @0x140C3760E / @0x140C79BAF）；
- `showLauncherNotificationWithAudio(t, m, sound)`（0x1408cfd20，返回值丢弃）。

### 3.3 依赖订正

`showLauncherNotificationWithAudio`（0x1408cfd20）序言保存 rax/rbx/rcx/rdi/rsi/r8 = 3×string，
实证签名 `(title, message, sound string) error`；原 `launcherupdatetoast_windows.go` 二参存根为
签名错误，本批订正为三参（仍 [S-sig] 骨架占位，体 Toast COM 未离线实现）。

## G4 独立复核

`desktopwidgets_notification_windows.go`：新建，2 函数 `[S]`。
`desktopwidgets_notification_windows_test.go`：新建，normalize 表驱动 8 例（控制字符折叠/空白折叠/
空串回退/ASCII 截断/CJK 截断/DEL 折叠/不截断）+ show 编排 smoke。
`launcherupdatetoast_windows.go`：订正 `showLauncherNotificationWithAudio` 三参签名。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

桌面小部件提醒通知域整文件闭环：2 函数全 `[S]`，控制字符折叠闭包、fallback/声音常量、120/240
rune 上限全部 asm/rodata 实证。依赖函数 `showLauncherNotificationWithAudio` 签名订正为三参
（Toast COM 体仍是 [S-sig] 骨架）。P=41 持平。FUNCS 2885/4754 = 60.69%。未落地文件差集 42→41。
