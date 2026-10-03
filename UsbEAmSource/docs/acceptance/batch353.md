# 批次 353 · 启动调试日志/覆盖层窗口句柄/自动激活 profile/焦点重试调度 +4（FUNCS 3168）

## 目标

落地 4 个函数：`launcherStartupDebugLog`、`visibleOverlayNativeWindowHandlesLocked`、
`autoActivatedProfileIDForTriggeredProfilesLocked`、`scheduleFocusRetryLocked`。

## 基线 / 收口

| 指标 | 基线（batch 352 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3164 | 3168 |
| MARKED | 3164 | 3168 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1581 | 1585 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1541 | 1541 |
| USABLE | 1542 | 1542 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1541  FUNCS=3168  MARKED=3168  P=41  S-eq=1  S-inline=37  S-sig=1585  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1585 + 41 = 3168 = FUNCS`。
S-sig 1581→1585（+4）、FUNCS 3164→3168（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 launcherStartupDebugLog [S-sig 0x1408d2f40, 352B]

debugEnabled/logger 空→返回；Sprintf → time.Now().Format → Fprintf → Sync。

### 3.2 visibleOverlayNativeWindowHandlesLocked [S-sig 0x140904580, 352B]

make(map) → 遍历 overlay(+0xf8) → 非空 NativeWindowHandle → result[handle]=true。

### 3.3 autoActivatedProfileIDForTriggeredProfilesLocked [S-sig 0x140906ec0, 352B]

autoID=TrimSpace(+0xb8/+0xc0) → 遍历 triggered 匹配。

### 3.4 scheduleFocusRetryLocked [S-sig 0x14090dfa0, 352B]

stopFocusRetryLocked → ensureLifecycleLocked → AfterFunc(24000ms) → 存 timer(+0x150)。

## G4 独立复核

- `backend/hotkey_dispatch_stubs.go`：+launcherStartupDebugLog [S-sig]。
- `backend/oledblackout.go`：+visibleOverlayNativeWindowHandlesLocked [S-sig]
  +autoActivatedProfileIDForTriggeredProfilesLocked [S-sig] +scheduleFocusRetryLocked [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3168/4754 = 66.64%。下一批：plugin 域
normalizePluginPackageFile / resolvePluginCatalogBaseURL。
