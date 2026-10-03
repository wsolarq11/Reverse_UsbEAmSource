# 批次 360 · 更新回滚校验/快捷方式解析/任意键关闭/可见配置匹配 +4（FUNCS 3196）

## 目标

落地 4 个函数：`validateLauncherUpdateRollbackBackup`、`BootstrapService.ResolveShortcut`、
`oledBlackoutService.handleOverlayDismissOnAnyKey`、`oledBlackoutService.findMatchingVisibleProfileIDLocked`。

## 基线 / 收口

| 指标 | 基线（batch 359 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3192 | 3196 |
| MARKED | 3192 | 3196 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1609 | 1613 |
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
FAITHFUL=1541  FUNCS=3196  MARKED=3196  P=41  S-eq=1  S-inline=37  S-sig=1613  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1613 + 41 = 3196 = FUNCS`。
S-sig 1609→1613（+4）、FUNCS 3192→3196（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 validateLauncherUpdateRollbackBackup [S-sig 0x1408ce6a0, 384B]

Lstat → ErrNotExist→nil；err→err；mode bit 0x1b→error；launcherUpdatePathHasReparsePoint→error。

### 3.2 BootstrapService.ResolveShortcut [S-sig 0x14077d7e0, 384B]

resolveShortcutInfoWithIconResolver(path) → duffcopy × 3。

### 3.3 handleOverlayDismissOnAnyKey [S-sig 0x14090b0c0, 384B]

lock(+0x0) → field(+0x88)/field(+0x3a) 判定 → inputDismissGuardActiveLocked →
dismissOverlayForKeyboardInputLocked。

### 3.4 findMatchingVisibleProfileIDLocked [S-sig 0x14090d840, 384B]

profiles(+0xf8) 空→""；collectScreensLocked → oledBlackoutProfileMatchesVisibleScreens →
TrimSpace(profile.id)。

## G4 独立复核

- `backend/launcherupdate_security_windows.go`：+validateLauncherUpdateRollbackBackup [S-sig]。
- `backend/bootstrapservice_window.go`：+BootstrapService.ResolveShortcut [S-sig]。
- `backend/oledblackout_windows.go`：+handleOverlayDismissOnAnyKey [S-sig]
  +findMatchingVisibleProfileIDLocked [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3196/4754 = 67.23%。下一批：filesearch 域
searchWithPathsContextMetrics / searchCandidatesMetrics。
