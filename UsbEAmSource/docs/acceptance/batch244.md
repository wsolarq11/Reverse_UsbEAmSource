# 批次 244 — launcherupdate 下载链三函数签名订正（2 升档 + 1 修正）

## 基线 / 收口

| 指标 | 基线（批次 243 收口） | 收口（批次 244） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1449 | **1451** |
| P | 47 | **45** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2741（57.65%） | **2743**（57.69%） |

SHA256 `b8f2a593da6f09623d4e0dab7415795db624aee4f9ef0def1cb08db0a888187c`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/launcherupdate_runtime.go`：下载链三函数签名订正。

- `downloadLauncherUpdatePackageWithWorkspace`：`[P]` → `[S-sig]`，签名
  `(pkg LauncherUpdatePackageConfig) error` →
  `(workspace WorkspaceLayout, ctx, taskID int64, path string, concurrent bool, url, sha256 string, totalSize int64, root string) (int64, error)`。
- `downloadLauncherUpdatePackageConcurrently`：`[P]` → `[S-sig]`，签名
  `(ctx, pkg LauncherUpdatePackageConfig, dir string) (int64, error)` →
  `(ctx, taskID int64, access LauncherNetworkAccess, total int64, path string, concurrent bool, url, sha256 string, totalSize int64, root string) (int64, error)`。
- `downloadLauncherUpdatePackageSequentially`：订正批 242 签名（栈参数实为 8 槽），
  `(ctx, taskID, access, path, url, sha256, expectedSize)` →
  `(ctx, taskID, access, path, concurrent bool, url, sha256, totalSize int64, root string)`。

## 签名证据（调用链三方互证）

调用链：`prepareLauncherUpdatePackageWithWorkspace`（0x1408b8160）→
`downloadLauncherUpdatePackageWithWorkspace`（0x1408b8f20）→
`downloadLauncherUpdatePackageConcurrently`（0x1408ba9a0）/
`downloadLauncherUpdatePackageSequentially`（0x1408b9a60）。

| 函数 | morestack 槽 | 参数拆分 |
|---|---|---|
| downloadLauncherUpdatePackageWithWorkspace | 6（rax..r8） | workspace 值传（栈首参 160B，ConfigFile@+0x10 传 newLauncherNetworkAccess）+ rbx/rcx=ctx + rdi=taskID + rsi/r8=path（prepare 0x1408b8572/0x1408b8564 传 filepath.join 结果） |
| concurrently | 9（rax..r11） | rbx/rcx=ctx（WithCancel）+ rdi=taskID + rsi/r8=access + r9=total（Truncate/buildRanges 首参）+ r10/r11=path（os.OpenFile） |
| sequentially | 8（rax..r10） | rbx/rcx=ctx + rdi=taskID + rsi/r8=access + r9/r10=path（os.OpenFile） |

download 的栈参数（相对其 rsp，frame=0xd8）8 槽：`concurrent bool(@0x188) + url(@0x190/0x198) +
sha256(@0x1a0/0x1a8) + totalSize(@0x1b0) + root(@0x1b8/0x1c0)`，经 4 个 xmmword
（0x1408b90d6/0x1408b90ef/0x1408b9104 段）透传 concurrently/sequentially 的栈参区
（concurrently frame=0x118 → 栈首参 @0x128；sequentially frame=0x1d0 → 栈首参 @0x1e0）。

### 关键纠正

- **workspace 是 `WorkspaceLayout` 值传**（非 `LauncherUpdatePackageConfig` 打包体）；旧桩把
  `pkg LauncherUpdatePackageConfig` 整体当唯一参，实为 workspace + ctx/taskID/path 寄存器 +
  concurrent/url/sha256/totalSize/root 栈参。
- **批 242 的 sequentially 漏 2 参**：栈参数实为 8 槽（concurrent bool + url + sha256 + totalSize + root），
  批 242 只录 url/sha256/expectedSize（5 槽），漏 `concurrent bool` 与 `root string`。
- **download 返回 `(int64, error)`**（非 `error`）：尾迹分叉走 concurrently/sequentially，两者均返回 `(int64, error)`。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`vet ./backend` EXIT=0；
  `test ./backend` `ok changeme/backend`。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1451 / P=45 / UNMARKED=0`。
- **G3 行为**：纯签名订正（体空→体空零骨架 `return 0, nil`），无调用方，无行为变更；既有测试全量 PASS。
- **G4 review**：`launcherupdate_runtime.go`（2 升档 + 1 修正）。

## 遗留（下一批）

- P 已降至 45。launcherupdate_runtime.go 剩余 4 个 [P]：runLauncherUpdateTask（7 参）、
  fetchLauncherRemoteConfigWithWorkspace（workspace 栈结构体，本次已证 workspace=WorkspaceLayout，
  可顺藤解码）、beginLauncherUpdateTask（6 寄存器返回）、prepareLauncherUpdatePackageWithWorkspace
  （workspace+ctx+taskID+path + concurrent/url/sha256/totalSize/root 栈参，本次已从调用点反推）。
- 截图域、oledblackout_windows（25）、nativedrag createTemporaryDirectoryShortcut。
- §10 差集 54 文件；FUNCS 缺口核心是 697 个未落地顶层函数 + 闭包/方法。
