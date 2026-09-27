# 批次 41 验收记录 — bootstrapservice_callees.go 34 未标落档

> 记录方式：每批次四路质检（§6.1）。本文件为 batch 41 的验收存证，全套可回溯。
> 研究用途；版权（c）2026 DOGFIGHT360 合规。

## 0. 范围

`backend/bootstrapservice_callees.go` 34 个 UNMARKED 函数全部落档：

| 档位 | 数量 | 函数 |
|---|---|---|
| [S] | 4 | `languageDirectoriesForWorkspace`、`isSupportedAppPath`、`launcherConfigsEqual`、`resolveScreenshotPNGFromRef`（恢复批次 24 丢失标记） |
| [S-sig] | 4 | `startDetachedCommand`、`screenshotPinWindowService.PersistSnapshotsForShutdown`、`BootstrapService.loadLinkPreferences`、`resolveBookmarkPageTitleWithNetwork` |
| [P] | 23 | GPU×6、QR×4、语言×2、书签/链接/桌面/更新/配置×11 |
| 幽灵删除 | 3 | 包级 `languageDirectories`、包级 `setHotkeyCaptureLease`、`parseLauncherImageDataURLHeader` |

反汇编资产：pipeline/tmp 已有 12 个 + 本批 `va_dump.py` 现场抽取 21 个（VA 全部来自
`symbols.txt`，长度来自相邻符号差，未手抄任何文档 VA）。

## G1 编译/静态 — PASS

```bash
go build -tags production -trimpath ./backend            # EXIT=0
go vet   -tags production ./backend                      # EXIT=0
go test  -count=1 -p=1 -tags production ./backend        # ok changeme/backend
bash build.sh                                            # EXIT=0，重建 artifact
```

artifact 哈希：`B468005DD5B03F135D00FFC218BE85D8792BACD7DEEDA57DCED67CAA39A26C63`
（SHA256，19,726,848 B）。

## G2 语义契约 — 签名/符号清单

| 符号 | 判定 | 依据 |
|---|---|---|
| `main.languageDirectories`（包级） | 幽灵删除 | symbols.txt 仅有 `main.BootstrapService.languageDirectories@0x1407a2720` |
| `BootstrapService.languageDirectories` | [S-sig]→[S] | 128B asm：`workspaceSnapshot()` → duffcopy → `languageDirectoriesForWorkspace` |
| `languageDirectoriesForWorkspace` | [S] | 192B asm：TrimSpace(AppLanguageDir@+0x30)→TrimSpace(LanguageDir@+0x20)→nil/单元素 |
| `main.setHotkeyCaptureLease`（包级） | 幽灵删除 | symbols.txt 仅有 `main.BootstrapService.setHotkeyCaptureLease@0x1407906e0`（方法已实现于 bootstrapservice.go:2479） |
| `parseLauncherImageDataURLHeader` | 幽灵删除 | symbols.txt 无此符号，头解析内联于 `decodeLauncherImageDataURL@0x1408a5200` |
| `loadLinkPreferences` | 包级→方法 | symbols.txt 为 `main.BootstrapService.loadLinkPreferences@0x140769c60` |
| `startDetachedCommand` | [S-sig] `([]string,bool)` | 256B asm `mov rdx,[rax];mov r8,[rax+8]` 读 slice 首元素字符串头，非 string 内容 |
| `resolveBookmarkPageTitleWithNetwork` | [S-sig] 首参 string | 入口仅 4 寄存器（两个 string），非 WorkspaceLayout 整结构 |
| `launcherConfigsEqual` | [S] | 192B asm：json.Marshal 双侧 + 长度相等 + memequal |
| `isSupportedAppPath` | [S] | 224B asm：末段扩展名提取 + 4 个 little-endian 魔数 |
| `PersistSnapshotsForShutdown` | [S-sig] | 224B asm：lock→restoreSuppressed@+0x52 分支→shutting@+0x50/shutdownPersisted@+0x53→snapshotMirror/saveSnapshots |

## G3 行为自测 — 全部通过

新增 `backend/bootstrapservice_callees_test.go`：

| 用例 | 覆盖点 |
|---|---|
| TestLanguageDirectoriesForWorkspace | AppLanguageDir 优先、LanguageDir 回退、双空返 nil |
| TestIsSupportedAppPath | .exe/.LNK(大小写)/.url/.appref-ms 命中；无扩展名/目录路径/复合点号拒绝 |
| TestLauncherConfigsEqual | 零值相等、Apps 差异不等 |

`go test -count=1 -p=1` 全绿。

## G4 独立复核 — 结论

独立重读 asm 与实现，核对：

1. `languageDirectoriesForWorkspace` 偏移复核：struct 偏移 0x30=AppLanguageDir、0x20=LanguageDir，
   与 asm `[rsp+0x60]/[rsp+0x50]`（=rbp+0x40/rbp+0x30，减去 struct 起始 rbp+0x10）一致。
2. `isSupportedAppPath` 反扫循环：遇 `\`/`/` 早于 `.` 时 ext 为空（return false），与 asm
   `cmp dl,0x5c/0x2f → je 空扩展名` 一致；4 个魔数 little-endian 解码无误。
3. `launcherConfigsEqual`：任一侧 Marshal 出错即 false，长度不等即 false，长度相等走 memequal，
   与 asm 三条出口（0x88bd54/0x88bd4c/0x88bd46）一致。
4. `startDetachedCommand` 变长切片边界（`args[:len(args)-1]`）存疑，未确证前保持 [S-sig]，
   不制造伪 [S]；调用点 `RestartApplication` 传 `[]string{exe}` 的修正留给还原 body 时一并处理。
5. `PersistSnapshotsForShutdown` 只落锁/标志语义，`snapshotMirror`/`saveSnapshots`
   （screenshot_pin_windows.go 未落地）未接入，保持 [S-sig] 并写明阻断原因。

## 结论

批次 41 四门：G1 ✅ / G2 ✅ / G3 ✅ / G4 ✅。
`bootstrapservice_callees.go` 34 个 UNMARKED 全部落档（4 [S] + 4 [S-sig] + 23 [P] + 3 幽灵删除），
文件内 UNMARKED 归零。23 个 [P] 均属 §10 未落地域（gpu/qrcode/bookmarks/link/desktopwidgets/
launcherconfiginput），已注明真实 VA 与阻断原因，留待各专项批次。
