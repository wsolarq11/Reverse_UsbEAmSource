# 批次 341 · 键盘钩子绑定更新/PrintScreen 消息/钩子卸载/启动致命日志 +4（FUNCS 3119）

## 目标

落地 4 个函数：`windowsLauncherGlobalHotkeyManager.updateKeyboardHookHotkeys`、
`handlePrintScreenHookMessage`、`uninstallKeyboardHook`、`launcherStartupDebugFatal`。

## 基线 / 收口

| 指标 | 基线（batch 340 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3115 | 3119 |
| MARKED | 3115 | 3119 |
| S | 1500 | 1500 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1536 | 1540 |
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
FAITHFUL=1537  FUNCS=3119  MARKED=3119  P=41  S-eq=1  S-inline=37  S-sig=1540  S=1500  USABLE=1538
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1500 + 37 + 1 + 1540 + 41 = 3119 = FUNCS`。
S-sig 1536→1540（+4）、FUNCS 3115→3119（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 updateKeyboardHookHotkeys [S-sig 0x1408a0ae0, 288B]

写 keyboardHookBindings(+0x30/+0x38/+0x40) → 置 printScreenKeyDown(+0x48)=false；
len==0→uninstallKeyboardHook；否则 ensureKeyboardHook。

### 3.2 handlePrintScreenHookMessage [S-sig 0x1408a1000, 288B]

keydown=(0x100/0x104)、keyup=(0x101/0x105)；matchPrintScreenHookRegistration 空→
置(+0x48)=0 返回 false；keydown 已置→true；否则置位 + dispatchKeyboardHookAction。

### 3.3 uninstallKeyboardHook [S-sig 0x1408a0ec0, 320B]

keyboardHook(+0x20) nil→置(+0x28)=0 返 (nil,nil)；否则 LazyProc.Call(UnhookWindowsHookEx)；
失败→fmt.Errorf。

### 3.4 launcherStartupDebugFatal [S-sig 0x1408d30a0, 320B]

launcherStartupDebugLog("fatal: %v", err) → log.Fatal(err)。

## G4 独立复核

- `backend/hotkeymanager.go`：+updateKeyboardHookHotkeys [S-sig] +handlePrintScreenHookMessage
  [S-sig] +uninstallKeyboardHook [S-sig]。
- `backend/hotkey_dispatch_stubs.go`：+launcherStartupDebugFatal [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3119/4754 = 65.61%。下一批：filesearch 域
VolumeIndex.OverlayStats / overlayStatsLocked / lockUSNFollowerSidecar。
