# 批次 245 — launcherupdate 域 [P] 全清零（4 升档 + 1 双参订正）

## 基线 / 收口

| 指标 | 基线（批次 244 收口） | 收口（批次 245） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1451 | **1455** |
| P | 45 | **41** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2743（57.69%） | **2747**（57.78%） |

SHA256（`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B）
`5f8132cd7b71df5dd500cbdd69294521ee435b3ab638ecff737f388512cf751d`。

## 本批内容

`backend/launcherupdate_runtime.go`：launcherupdate 域最后一个 [P] 清零，全部升档 [S-sig]。

- `fetchLauncherRemoteConfigWithWorkspace`：`[P]` → `[S-sig]`，签名
  `()` → `(workspace WorkspaceLayout, ctx context.Context) (LauncherConfig, error)`。
- `prepareLauncherUpdatePackageWithWorkspace`：`[P]` → `[S-sig]`，签名
  `(pkg LauncherUpdatePackageConfig) (string, error)` →
  `(workspace WorkspaceLayout, ctx, taskID int64, executableName string, concurrent bool, url, sha256 string, totalSize int64, root string) (string, int64, error)`。
- `runLauncherUpdateTask`：`[P]` → `[S-sig]`，签名
  `(ctx, version)` → `(ctx, taskID int64, done func(), version string)`。
- `beginLauncherUpdateTask`：`[P]` → `[S-sig]`，签名
  `(version string) (int64, error)` → `(ctx context.Context) (context.Context, int64, func(), error)`。
- `isLauncherVersionUpdateAvailableText`：订正双参 `(text string)` → `(version, remoteVersion string)`。

## 签名证据（三方互证）

| 函数 | morestack 槽 | 参数拆分 |
|---|---|---|
| fetch | 3（rax/rbx/rcx） | rbx/rcx=ctx（NewRequestWithContext）+ workspace 值传（栈首参 @0x1c8，ConfigFile@+0x10 传 newLauncherNetworkAccess） |
| prepare | 6（rax..r8） | workspace 值传（@0x280）+ rbx/rcx=ctx + rdi=taskID + rsi/r8=executableName（locateLauncherUpdateContentRoot 第三参）+ 栈 8 槽 concurrent/url/sha256/totalSize/root |
| runLauncherUpdateTask | 7（rax..r9） | rbx/rcx=ctx + rdi=taskID + rsi=done func()（尾声 call rax 无参）+ r8/r9=version |
| beginLauncherUpdateTask | 3（rax/rbx/rcx） | rbx/rcx=ctx；返回 6 寄存器 =(ctx,taskID,done func(),error) |

### 关键纠正

- **beginLauncherUpdateTask 第二参是 ctx 不是 version**：InstallLauncherUpdate 调用点 0x1408b35a0
  传 `context.Background()` 全局 itab/data（`lea rbx,[rip+0x9250ca]` / `lea rcx,[rip+0x13a7423]`），
  version 是 InstallLauncherUpdate 经闭包捕获（0x1408b366d 段 [rax+0x30]/[rax+0x38]）传给
  runLauncherUpdateTask goroutine 的，非 beginLauncherUpdateTask 参数。
- **beginLauncherUpdateTask 返回 6 寄存器 = (context.Context, int64, func(), error)**：成功尾迹
  AX/BX=WithCancel 结果 ctx、CX=taskID（[0x4d0] 自增）、DI=done 闭包（捕获 channel/receiver/taskID）、
  SI/R8=error；错误尾迹（0x1408b78e9）填 version/0/常量/error，与 InstallLauncherUpdate 消费点
  （0x1408b35a5 test rsi 判 error，0x1408b35ae 存 taskID/done）对齐。
- **isLauncherVersionUpdateAvailableText 实为双参**：调用点 0x1408b38e9 设 AX/BX=version + CX/DI=远程
  版本（fetch 返回结构体字段 0），旧桩漏第二参。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`vet ./backend` EXIT=0；
  `test ./backend` `ok changeme/backend`。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1455 / P=41 / UNMARKED=0`。
- **G3 行为**：纯签名订正（体空→体空零骨架），无调用方变更，既有测试全量 PASS。
- **G4 review**：`launcherupdate_runtime.go`（launcherupdate 域 [P] 已清零）。

## 遗留（下一批）

- P=41。剩余 [P] 集中在：oledblackout_windows.go（~25）、screenshot_windows.go（~7）、
  screenshot_uia_windows.go（~5-6）、mousegestures.go（1）、nativedrag_windows.go（1）、
  oledblackout.go（1 ghost newOLEDLifecycleContext）。
- 缺失函数 `normalizeLauncherUpdatePackageRoot`（源码 602-613 行，launcherupdate 域）待补。
- §10 差集 54 文件；FUNCS 缺口核心是 697 个未落地顶层函数 + 闭包/方法。
