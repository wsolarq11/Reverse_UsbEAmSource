# 批次 344 · 通知记录比较/天气配置变化/UI 缩放/线程消息 +4（FUNCS 3131）

## 目标

落地 4 个函数：`desktopWidgetNotificationRecordAfter`、`desktopWeatherConfigChanged`、
`BootstrapService.loadLauncherUIScalePercent`、`inputMonitorPostThreadMessageValue`。

## 基线 / 收口

| 指标 | 基线（batch 343 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3127 | 3131 |
| MARKED | 3127 | 3131 |
| S | 1501 | 1501 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1547 | 1551 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1538 | 1538 |
| USABLE | 1539 | 1539 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1538  FUNCS=3131  MARKED=3131  P=41  S-eq=1  S-inline=37  S-sig=1551  S=1501  USABLE=1539
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1501 + 37 + 1 + 1551 + 41 = 3131 = FUNCS`。
S-sig 1547→1551（+4）、FUNCS 3127→3131（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 desktopWidgetNotificationRecordAfter [S-sig 0x1407b5000, 320B]

time.Time.Equal 相等→cmpstring id>；否则 time.Time.After。

### 3.2 desktopWeatherConfigChanged [S-sig 0x1407b6bc0, 320B]

location 不等→true；normalizeDesktopWeatherProviderOrder 后逐元素比较。

### 3.3 loadLauncherUIScalePercent [S-sig 0x14079c660, 320B]

workspaceSnapshot → loadLauncherConfigIfExists → 读 config(+0x9c8) → clamp [100,300]。

### 3.4 inputMonitorPostThreadMessageValue [S-sig 0x140869660, 320B]

newobject(2 字段) → LazyProc.Call(PostThreadMessageW, 4) → 失败 fmt.Errorf。

## G4 独立复核

- `backend/desktopwidgets_service.go`：+desktopWidgetNotificationRecordAfter [S-sig]
  +desktopWeatherConfigChanged [S-sig]。
- `backend/hotkey_dispatch_stubs.go`：+loadLauncherUIScalePercent [S-sig]。
- `backend/inputmonitor_windows.go`：+inputMonitorPostThreadMessageValue [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3131/4754 = 65.86%。下一批：oledBlackout 域
collectScreensLocked / oledBlackoutScreenIndex。
