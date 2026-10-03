# 批次 305 · 键盘钩子分发 + IPropertyStore.Commit +2（FUNCS 3001）

## 目标

落地 2 个短函数：`windowsLauncherGlobalHotkeyManager.dispatchKeyboardHookAction`（钩子分发）、
`launcherAppIdentityIPropertyStore.Commit`（WinRT 属性提交）。新建 `launcherAppIdentityIPropertyStore` 类型。

## 基线 / 收口

| 指标 | 基线（batch 304 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2999 | 3001 |
| MARKED | 2999 | 3001 |
| S | 1455 | 1456 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1465 | 1466 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1492 | 1493 |
| USABLE | 1493 | 1494 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1493  FUNCS=3001  MARKED=3001  P=41  S-eq=1  S-inline=37  S-sig=1466  S=1456  USABLE=1494
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1456 + 37 + 1 + 1466 + 41 = 3001 = FUNCS`。
S 1455→1456（+1）、S-sig 1465→1466（+1）、FUNCS 2999→3001（+2）、FAITHFUL 1492→1493（+1）、
UNMARKED=0/P=41 保持。**FUNCS 首次突破 3000。**

## G3 行为（asm 逐地址实证）

### 3.1 windowsLauncherGlobalHotkeyManager.dispatchKeyboardHookAction [S-sig 0x1408a1300, 192B]

`(a, b uintptr)`。读 receiver[+0x00] 回调，nil 直接返回；否则 newobject 捕获
(回调, a, b) → newproc 起 goroutine；gowrap1 体待 hotkey 域专项。

### 3.2 launcherAppIdentityIPropertyStore.Commit [S 0x14086e160, 192B]

`() error`。SyscallN(vtbl[7]@0x38 Commit, this) → HRESULT 有符号 <0 → fmt.Errorf 包装；
否则 nil。新建类型 `launcherAppIdentityIPropertyStore{vtbl *uintptr}`。

## G4 独立复核

- `backend/hotkeymanager.go`：+dispatchKeyboardHookAction [S-sig]。
- `backend/launcherappidentity_windows.go`：+Commit [S] + 类型 + fmt/syscall/unsafe import。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

2 个函数落地（1 [S] + 1 [S-sig]）。FUNCS 3001/4754 = 63.13%。下一批：filesearch
SearchWithPaths 薄包装 + searchWithPathsContextMetrics 签名 / remoteicons
validateRemoteIconCachePath 路径校验。
