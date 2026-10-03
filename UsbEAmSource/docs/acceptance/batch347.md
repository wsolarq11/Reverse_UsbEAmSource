# 批次 347 · 通知图标路径/AppID PROPVARIANT/线程消息/插件窗口存储 +4（FUNCS 3144）

## 目标

落地 4 个函数：`resolveLauncherNotificationIconPath`、`launcherAppIdentityLPWSTRPropVariant`、
`postLauncherThreadMessage`、`pluginWindowService.store`。

## 基线 / 收口

| 指标 | 基线（batch 346 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3140 | 3144 |
| MARKED | 3140 | 3144 |
| S | 1503 | 1503 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1558 | 1562 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1540 | 1540 |
| USABLE | 1541 | 1541 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1540  FUNCS=3144  MARKED=3144  P=41  S-eq=1  S-inline=37  S-sig=1562  S=1503  USABLE=1541
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1503 + 37 + 1 + 1562 + 41 = 3144 = FUNCS`。
S-sig 1558→1562（+4）、FUNCS 3140→3144（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 resolveLauncherNotificationIconPath [S-sig 0x14086bda0, 320B]

Getenv → UserCacheDir → filepath.Join(3 段)。

### 3.2 launcherAppIdentityLPWSTRPropVariant [S-sig 0x14086db00, 320B]

UTF16FromString → PROPVARIANT{vt:0x1f, pwszVal}。

### 3.3 postLauncherThreadMessage [S-sig 0x1408a16a0, 320B]

newobject(2 字段) → LazyProc.Call(PostThreadMessageW, 4) → 失败 fmt.Errorf。

### 3.4 pluginWindowService.store [S-sig 0x140930300, 320B]

lock → shutting(+0x28)==0 且 windows(+0x10)[key]==win → mapassign → unlock。

## G4 独立复核

- `backend/launcherappidentity_windows.go`：+resolveLauncherNotificationIconPath [S-sig]
  +launcherAppIdentityLPWSTRPropVariant [S-sig]。
- `backend/hotkeymanager.go`：+postLauncherThreadMessage [S-sig]。
- `backend/bootstrapservice_state_deps.go`：+pluginWindowService.store [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3144/4754 = 66.14%。下一批：screenshot 域
screenshotScrollingChunkedCanvas.trimBottom / repairScreenshotScrollingSeamArtifacts。
