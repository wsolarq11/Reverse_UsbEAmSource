# 批次 278 · desktopwidgets_notification.go 整文件落地 + Notify 方法（+2 [S]，+2 FUNCS）

## 目标

整文件落地 `desktopwidgets_notification.go`（38 个未落地文件差集之一）：平台分派空函数
`platformDesktopWidgetNotifierNotify` 全 `[S]`；并补落地 `platformDesktopWidgetNotifier.Notify`
接口方法（符号 0x1407ac8e0，desktopWidgetNotifier 接口的实现，直连 showDesktopWidgetNotification）。

## 基线 / 收口

| 指标 | 基线（batch 277 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2891 | 2893 |
| MARKED | 2891 | 2893 |
| S | 1371 | 1373 |
| S-inline | 36 | 36 |
| S-sig | 1443 | 1443 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1407 | 1409 |
| USABLE | 1407 | 1409 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build ./backend` EXIT=0；`go1.25.12 vet ./backend` EXIT=0；
`go1.25.12 test ./backend -run Notifier/Notification/Notify` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1409  FUNCS=2893  MARKED=2893  P=41  S-eq=0  S-inline=36  S-sig=1443  S=1373  USABLE=1409
```

分项自洽：`S + S-inline + S-sig + P = 1373 + 36 + 1443 + 41 = 2893 = FUNCS`。
S 1371→1373（+2）、FUNCS 2891→2893（+2）、FAITHFUL 1407→1409（+2）、P=41/S-sig=1443/UNMARKED=0 保持。

账目：新建 `desktopwidgets_notification.go`（platformDesktopWidgetNotifierNotify [S]）+1、
`desktopwidgets_notification_windows.go` 追加 Notify 方法 [S] +1。

## G3 行为（asm 逐地址实证）

### 3.1 platformDesktopWidgetNotifierNotify [S VA=0（DCE）]

`(title, body string, silent bool)`。Windows 构建中被编译期死代码消除（VA=0）：Notify 方法
直连 showDesktopWidgetNotification，不经由本函数。语义为无副作用空实现（no-op）。

### 3.2 (platformDesktopWidgetNotifier) Notify [S 0x1407ac8e0, 34B]

`(title, message string, silent bool) error`。asm 实证：

- rax/rbx=title、rcx/rdi=message、rsi=silent 透传；
- `call showDesktopWidgetNotification`（0x1407ac960）；
- 返回 (nil error)（rax/rbx 清零，`add rsp,0x28; pop rbp; ret`）。

实现 desktopWidgetNotifier 接口（types_desktopwidget.go:522），使
`desktopWidgetService.deliverNotification` 的 `s.notifier.Notify(...)` 调用链闭合。

### 3.3 依赖实证

- `showDesktopWidgetNotification` [S 0x1407ac960]（batch275 落地）；
- `desktopWidgetNotifier` 接口 `Notify(string,string,bool) error`（types_desktopwidget.go:522）；
- `platformDesktopWidgetNotifier` 空结构体（types_misc.go:118）。

## G4 独立复核

`desktopwidgets_notification.go`：新建，1 函数 `[S]`（VA=0 空实现）。
`desktopwidgets_notification_windows.go`：追加 Notify 方法 `[S]`（转发 showDesktopWidgetNotification）。
`desktopwidgets_notification_windows_test.go`：追加 Notify 接口方法与 no-op 平台分派两组行为测试。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

桌面小部件通知域整文件闭环：通知器接口方法 Notify 与平台分派空函数全部落地，
`deliverNotification → notifier.Notify → showDesktopWidgetNotification` 调用链闭合。
P=41 持平。FUNCS 2893/4754 = 60.85%。未落地文件差集 39→38。
