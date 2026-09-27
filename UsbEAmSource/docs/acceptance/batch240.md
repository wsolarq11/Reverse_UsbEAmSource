# 批次 240 — probeLauncherUpdatePackageRange 签名订正（1 升档）

## 基线 / 收口

| 指标 | 基线（批次 239 收口） | 收口（批次 240） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1444 | **1445** |
| P | 52 | **51** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2736（57.55%） | **2737**（57.57%） |

SHA256 `df43216244c069d13fb1cc49cca4452861ab2ab40ce1e163370a32cf453e18a9`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/launcherupdate_runtime.go`：`probeLauncherUpdatePackageRange` 升档 `[S-sig]`，签名
`(ctx context.Context, url string) (int64, error)` →
`(ctx context.Context, access LauncherNetworkAccess, url string) (bool, int64, error)`。

### 签名证据（0x1408b9300）

morestack 序言（第 284-298 行）保存 6 槽，逐寄存器实证：

| 项 | 类型 | 证据 |
|---|---|---|
| rax/rbx | ctx context.Context | NewRequestWithContext 首参透传（第 21 行 call） |
| rcx/rdi | access LauncherNetworkAccess（itab/data） | 0x1408b94ff 读 `[itab+0x18]`=Do 方法槽（fun[0]），0x1408b951a `call Do(data, req, maxBytes)` |
| rsi/r8 | url string（ptr/len） | NewRequestWithContext 第三参透传 |

返回四寄存器（多处 ret 路径 + deferreturn 0x1408b9879 一致）：
- AL = bool：仅 206 且 Content-Range total>0 时为 1（`[rsp+0x48]` 在 0x1408b973a 置 1），
  200/416/错误路径均为 0；
- BX = int64：total（206 取 parseLauncherUpdateContentRangeTotal 结果；200 取 resp.ContentLength）；
- CX/DI = error。

故返回 `(bool, int64, error)`，非旧签名的 `(int64, error)`。

### 关键纠正

旧注释「rcx/rdi 与 r8 的归属无法唯一确定」已解：rcx/rdi 是 `LauncherNetworkAccess` 接口
（itab+data，Do 方法经 itab fun[0] 调用），rsi/r8 是 url（ptr/len），r8 为 url 长度而非标量。
返回比旧签名多一个 bool（Content-Range 确定态：206 响应能给出权威总大小，供调用方决定走
Range 并发下载还是 200 顺序回退）。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`vet ./backend` EXIT=0；
  `test ./backend` `ok changeme/backend`。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1445 / P=51 / UNMARKED=0`。
- **G3 行为**：纯签名订正（体空→体空零骨架 `return false,0,nil`），无调用方，无行为变更；
  既有测试全量 PASS。
- **G4 review**：`launcherupdate_runtime.go`（1 升档）。

## 遗留（下一批）

- P 已降至 51。截图域：captureScreenshotAreaPNG（7 参数）、
  captureScreenshotWindowSelectionWithOptions、newScreenshotWindowSelectionSession、
  resolveControlHoverForWindow、screenshotAccessibleHitTest。
- launcherupdate_runtime.go 剩余 8 个 [P]：runLauncherUpdateTask（7 参）、
  fetchLauncherRemoteConfigWithWorkspace、beginLauncherUpdateTask（6 寄存器返回）、
  prepareLauncherUpdatePackageWithWorkspace、downloadLauncherUpdatePackageWithWorkspace、
  downloadLauncherUpdatePackageSequentially/Concurrently、downloadLauncherUpdatePackageRange。
- filelocator walkRoot/processFile；oledblackout_windows 25；nativedrag createTemporaryDirectoryShortcut。
