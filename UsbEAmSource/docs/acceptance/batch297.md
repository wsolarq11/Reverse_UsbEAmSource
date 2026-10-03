# 批次 297 · AppUserModelID 身份 + 截图预览窗口类 once 包装 +3（FUNCS 2970）

## 目标

落地三个 once 单次包装入口：`ensureLauncherAppIdentity`（AppUserModelID 配置缓存）、
`configureLauncherAppIdentity`（签名订正）、`ensureScreenshotNativePreviewWindowClass`
（预览窗口类注册缓存）。新建 `launcherappidentity_windows.go`。

## 基线 / 收口

| 指标 | 基线（batch 296 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2967 | 2970 |
| MARKED | 2967 | 2970 |
| S | 1444 | 1445 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1444 | 1446 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1481 | 1482 |
| USABLE | 1482 | 1483 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1482  FUNCS=2970  MARKED=2970  P=41  S-eq=1  S-inline=37  S-sig=1446  S=1445  USABLE=1483
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1445 + 37 + 1 + 1446 + 41 = 2970 = FUNCS`。
S 1444→1445（+1）、S-sig 1444→1446（+2）、FUNCS 2967→2970（+3）、FAITHFUL 1481→1482（+1）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 ensureLauncherAppIdentity [S 0x14086b100, 96B]

`() string`。读全局 once.done（[rip+0x13efe7b]）；未完成则 `sync.Once.doSlow`
携带闭包 f（code=func1 包装，context=nil）；闭包 0x1409f71a0 调
`configureLauncherAppIdentity()` 并把返回的 2 寄存器写入全局
[0x140c10870]=ptr / [0x140c10878]=len；尾声读该全局两字返回 string。

### 3.2 configureLauncherAppIdentity [S-sig 0x14086b160, 2650B(0xa5a)]

`() string`。签名经 ensureLauncherAppIdentity.func1 调用点实证——无参数、返回 2 寄存器
= string。体调用链 ensureLauncherNotificationIconFile → ensureLauncherToastRegistry →
setCurrentProcessExplicitAppUserModelID → ensureLauncherStartMenuShortcutWithAppID，
体待 AppUserModelID 域专项还原（当前占位返回空串）。

### 3.3 ensureScreenshotNativePreviewWindowClass [S-sig 0x14099bd80, 96B]

`() error`。签名实证——once 包装 + 返回全局 error 接口（[0x140c108e0]/[0x140c108e8]）。
func1（0x1409f5560, 675B）：GetModuleHandle(nil) → UTF16PtrFromString(0x25=37B 类名) →
newobject(WNDCLASSEX cbSize=0x50) → syscall.compileCallback 窗口过程 →
RegisterClassEx LazyProc.Call → GetLastError → fmt.Errorf；体待窗口类注册链专项还原。

## G4 独立复核

- `backend/launcherappidentity_windows.go`：新建，+2 函数（[S] + [S-sig]）+ 2 全局。
- `backend/screenshot_preview_native_windows.go`：+1 函数 [S-sig] + 2 全局 + sync 导入。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（1 [S] + 2 [S-sig]）。FUNCS 2970/4754 = 62.47%。下一批：filesearch
SearchWithPaths 薄包装 + searchWithPathsContextMetrics 签名 / remoteicons
validateRemoteIconCachePath 路径校验。
