# 批次 300 · 热键管理器 Update + 文件拖放注册 +3（FUNCS 2985）

## 目标

落地 3 个短函数：`windowsOLEDBlackoutHotkeyManager.Update`、
`windowsLauncherGlobalHotkeyManager.Update`（热键绑定更新入口）、
`registerLauncherFileDropHandler`（启动器文件拖放注册）。

## 基线 / 收口

| 指标 | 基线（batch 299 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2982 | 2985 |
| MARKED | 2982 | 2985 |
| S | 1450 | 1450 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1453 | 1456 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1487 | 1487 |
| USABLE | 1488 | 1488 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1487  FUNCS=2985  MARKED=2985  P=41  S-eq=1  S-inline=37  S-sig=1456  S=1450  USABLE=1488
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1450 + 37 + 1 + 1456 + 41 = 2985 = FUNCS`。
S-sig 1453→1456（+3）、FUNCS 2982→2985（+3）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 windowsOLEDBlackoutHotkeyManager.Update [S-sig 0x14090f3e0, 298B]

`(bindings []oledBlackoutHotkeyBinding) oledBlackoutHotkeyUpdateResult`。makechan(1) 打包
命令 chansend 到 commands(+0x08)，postLauncherThreadMessage(threadID(+0x20),0x80f2) 唤醒。
体待热键更新链专项。

### 3.2 windowsLauncherGlobalHotkeyManager.Update [S-sig 0x14089d8e0, 252B]

`(bindings launcherHotkeyBindings) error`。makechan(1) 打包命令 chansend 到 commands(+0x08)，
postLauncherThreadMessage(threadID(+0x18),0x80f1) 唤醒。体待 hotkey 域专项。

### 3.3 registerLauncherFileDropHandler [S-sig 0x1408d0760, 128B]

`(window application.Window)`。nil 接口返回；经 itab fun[0x138] 调窗口文件拖放注册方法
（携带全局目标 + 回调闭包）。体待 wails 窗口拖放接口方法名专项。

## G4 独立复核

- `backend/oledblackout_hotkey_windows.go`：+windowsOLEDBlackoutHotkeyManager.Update [S-sig]。
- `backend/hotkeymanager.go`：+windowsLauncherGlobalHotkeyManager.Update [S-sig]。
- `backend/hotkey_dispatch_stubs.go`：+registerLauncherFileDropHandler [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（3 [S-sig]）。FUNCS 2985/4754 = 62.79%。下一批：filesearch
SearchWithPaths 薄包装 + searchWithPathsContextMetrics 签名 / remoteicons
validateRemoteIconCachePath 路径校验。
