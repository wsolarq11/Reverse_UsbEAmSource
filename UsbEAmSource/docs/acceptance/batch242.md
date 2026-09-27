# 批次 242 — downloadLauncherUpdatePackageSequentially 签名订正（1 升档）

## 基线 / 收口

| 指标 | 基线（批次 241 收口） | 收口（批次 242） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1446 | **1447** |
| P | 50 | **49** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2738（57.59%） | **2739**（57.61%） |

SHA256 `485eae6a500c19ac30cd7d3796e6894bb61eb3c1f8bb29c3ec3e13f5b5f9500c`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/launcherupdate_runtime.go`：`downloadLauncherUpdatePackageSequentially` 升档 `[S-sig]`，签名
`(ctx context.Context, pkg LauncherUpdatePackageConfig, dir string) (int64, error)` →
`(ctx context.Context, taskID int64, access LauncherNetworkAccess, path string, url string, sha256 string, expectedSize int64) (int64, error)`。
无新增 import。

### 签名证据（0x1408b9a60）

morestack 序言（第 677-695 行）保存 8 槽（rax..r10，无 r11），逐寄存器实证：

| 项 | 类型 | 证据 |
|---|---|---|
| rax | receiver | 全程透传 |
| rbx/rcx | ctx context.Context | NewRequestWithContext 首参透传（0x1408b9aee `mov rax,rbx` / `mov rbx,rcx`） |
| rdi | taskID int64 | 0x1408ba776 作 setLauncherUpdateProgressForTask 第二参透传 |
| rsi/r8 | access LauncherNetworkAccess | 0x1408b9bd1 读 `[rsi+0x18]`=Do 方法槽，data=r8 |
| r9/r10 | path string | 0x1408b9cbe 作 `os.OpenFile` 首参 name（flags=0x242, perm=0x1b6） |

stack 5 槽：`url`（0x1e8/0x1f0）→ NewRequestWithContext 第三参；`sha256`（0x1f8/0x200）→
0x1408ba6d2 `cmp [rsp+0x200],rbx` 校验；`expectedSize int64`（0x208）→ 0x1408ba3f0 `cmp rax,rdx` 校验下载字节。

返回 rax=int64（下载字节）+ rbx/rcx=error。故返回 `(int64, error)`。

### 关键纠正

旧注释「7 寄存器无法唯一拆分 (ctx,pkg,dir,…) 组合」已解：完整形参为
`(ctx, taskID, access, path, url, sha256, expectedSize)`。旧签名把 `pkg LauncherUpdatePackageConfig`
整体压成一个结构体参、把 `dir` 当作唯一目标——实为 5 个独立标量/字符串参：taskID（进度上报）、
access（网络）、path（落盘文件）、url（下载源）、sha256+expectedSize（完整性校验）。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`vet ./backend` EXIT=0；
  `test ./backend` `ok changeme/backend`。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1447 / P=49 / UNMARKED=0`。
- **G3 行为**：纯签名订正（体空→体空零骨架 `return 0, nil`），无调用方，无行为变更；
  既有测试全量 PASS。
- **G4 review**：`launcherupdate_runtime.go`（1 升档）。

## 遗留（下一批）

- P 已降至 49。launcherupdate_runtime.go 剩余 6 个 [P]：runLauncherUpdateTask（7 参）、
  fetchLauncherRemoteConfigWithWorkspace、beginLauncherUpdateTask（6 寄存器返回）、
  prepareLauncherUpdatePackageWithWorkspace、downloadLauncherUpdatePackageWithWorkspace、
  downloadLauncherUpdatePackageConcurrently（8+ 参）。
- 截图域：captureScreenshotAreaPNG、captureScreenshotWindowSelectionWithOptions、
  newScreenshotWindowSelectionSession、resolveControlHoverForWindow、screenshotAccessibleHitTest。
- filelocator walkRoot/processFile；oledblackout_windows 25；nativedrag createTemporaryDirectoryShortcut。
