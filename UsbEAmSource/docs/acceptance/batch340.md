# 批次 340 · 预览 HTML/接缝修复/延迟显示/二实例显示 +4（FUNCS 3115）

## 目标

落地 4 个函数：`buildScreenshotPreviewWindowHTMLWithStateURL`、
`repairScreenshotScrollingGlobalSeamArtifacts`、`screenshotPreviewWindowService.waitBeforeReveal`、
`BootstrapService.showLauncherFromSecondInstance`。

## 基线 / 收口

| 指标 | 基线（batch 339 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3111 | 3115 |
| MARKED | 3111 | 3115 |
| S | 1500 | 1500 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1532 | 1536 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1537 | 1537 |
| USABLE | 1538 | 1538 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1537  FUNCS=3115  MARKED=3115  P=41  S-eq=1  S-inline=37  S-sig=1536  S=1500  USABLE=1538
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1500 + 37 + 1 + 1536 + 41 = 3115 = FUNCS`。
S-sig 1532→1536（+4）、FUNCS 3111→3115（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 buildScreenshotPreviewWindowHTMLWithStateURL [S-sig 0x14099b860, 256B]

TrimSpace → json.Marshal → 模板 Replace(json) → Replace("50")。

### 3.2 repairScreenshotScrollingGlobalSeamArtifacts [S-sig 0x1409a3960, 256B]

遍历行 → screenshotScrollingRowLooksDarkSeamArtifact → blendScreenshotScrollingRow；
统计修复数 → debug log。

### 3.3 waitBeforeReveal [S-sig 0x140999c60, 288B]

delay>0 则 NewTimer(delay) 阻塞后 shouldDisplay。

### 3.4 showLauncherFromSecondInstance [S-sig 0x140795da0, 288B]

ensureLauncherWindowForShow 非空→showLauncherWindow；否则 lock(+0x540) → field(+0x44a)=1 → unlock。

## G4 独立复核

- `backend/screenshot_preview_windows.go`：+buildScreenshotPreviewWindowHTMLWithStateURL [S-sig]
  +waitBeforeReveal [S-sig]。
- `backend/screenshot_scroll_windows.go`：+repairScreenshotScrollingGlobalSeamArtifacts [S-sig]。
- `backend/hotkey_dispatch_stubs.go`：+showLauncherFromSecondInstance [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3115/4754 = 65.53%。下一批：windows 域
windowsLauncherGlobalHotkeyManager.updateKeyboardHookHotkeys / launcherStartupDebugFatal。
