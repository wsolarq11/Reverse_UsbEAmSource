# 批次 302 · 预览窗口显示 + 插件窗口对接 + 平台启动 +4（FUNCS 2993）

## 目标

落地 4 个短函数：`screenshotPreviewWindowService.Show`（预览显示透传）、
`pluginWindowService.AttachApp`（插件窗口对接）、`inputMonitorService.startPlatformOwned`
（平台启动拥有）、`BootstrapService.applyLauncherWindowSizingForShow`（显示前尺寸）。

## 基线 / 收口

| 指标 | 基线（batch 301 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2989 | 2993 |
| MARKED | 2989 | 2993 |
| S | 1451 | 1453 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1459 | 1461 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1488 | 1490 |
| USABLE | 1489 | 1491 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（`bash`/`awk` 于当前会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1490  FUNCS=2993  MARKED=2993  P=41  S-eq=1  S-inline=37  S-sig=1461  S=1453  USABLE=1491
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1453 + 37 + 1 + 1461 + 41 = 2993 = FUNCS`。
S 1451→1453（+2）、S-sig 1459→1461（+2）、FUNCS 2989→2993（+4）、FAITHFUL 1488→1490（+2）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 screenshotPreviewWindowService.Show [S 0x140996520, 160B]

`(result ScreenshotCaptureResult) error`。duffcopy 结果结构体 → ShowForOwner("preview", result)
透传。

### 3.2 pluginWindowService.AttachApp [S 0x14092d4c0, 160B]

`(app *application.App)`。lock(+0x00).Lock + 写屏障写 app(+0x08) + Unlock。

### 3.3 inputMonitorService.startPlatformOwned [S-sig 0x140861a00, 160B]

`(a, b bool)`。advancePlatformGeneration → 读 +0x100 闭包，非 nil 间接调用 (a,b)，否则
startPlatform(a,b)。体待 startPlatform 专项。

### 3.4 BootstrapService.applyLauncherWindowSizingForShow [S-sig 0x1407968c0, 160B]

`(window application.Window)`。window nil 返回；consumeLauncherDefaultSizeReset → bool、
loadLauncherUIScalePercent → int，交 applyLauncherWindowSizing(window,percent,bool)。
体待两个依赖函数专项。

## G4 独立复核

- `backend/screenshot_services.go`：+Show [S]。
- `backend/bootstrapservice_state_deps.go`：+AttachApp [S]（补 application import）。
- `backend/inputmonitor.go`：+startPlatformOwned [S-sig]。
- `backend/bootstrapservice_window.go`：+applyLauncherWindowSizingForShow [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（2 [S] + 2 [S-sig]）。FUNCS 2993/4754 = 62.96%。下一批：filesearch
SearchWithPaths 薄包装 + searchWithPathsContextMetrics 签名 / remoteicons
validateRemoteIconCachePath 路径校验。
