# 批次 178 — 链接浏览器/工作区路径/配置校验等纯逻辑落地（13 函数新增）

## 基线 / 收口

| 指标 | 基线（批次 177 收口） | 收口（批次 178） |
|---|---|---|
| FUNCS | 2667 | **2680** |
| S | 1111 | **1120** |
| S-inline | 35 | 35 |
| S-sig | 1406 | **1410** |
| P | 115 | 115 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2552（53.7%） | **2565（53.9%）** |
| 未落地文件（差集） | 73 | **66** |

SHA256 `BF09B2818A3F254A4A4D4748BD60799EEF2B500F1735A6062B65AD891F44F18A`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批落地（13 函数新增，7 文件新建）

| 函数 | VA | 档位 | 文件 | 关键证据 |
|---|---|---|---|---|
| `desktopTimezoneOffset` | 0x1407ab120 | [S] | desktopwidgets_clock.go | offset<0 取 "-"；hours=abs/3600（magic 除法 0x48d159e26af37c05）、minutes=(abs-hours*3600)/60；`fmt.Sprintf("UTC%s%02d:%02d"@0x140c50712)` |
| `desktopWidgetStoreError` | 0x1407c4f20 | [S] | desktopwidgets_store.go | TrimSpace(op/path)→`fmt.Errorf("%s: %s: %s"@0x140c440c7, "DESKTOP_WIDGET_STORE_INVALID"@0x140c6b920,28B)` |
| `(*launcherUpdateUserError).Error` | 0x140a05c40 | [S] | launcherupdate.go | 类型为 string 别名；nil→panicwrap，否则直接返回 string(*e)（rax=[e]/rbx=[e+8]） |
| `resolveRegisteredLinkBrowserPath` | 0x1408d1740 | [S] | linkbrowser_windows.go | key=`App Paths\`前缀@0x140c899b2,52B+TrimSpace(name)；试 CURRENT_USER/LOCAL_MACHINE 两 hive |
| `readRegistryString` | 0x1408d1820 | [S] | linkbrowser_windows.go | OpenKey(QUERY_VALUE=1) 失败→""；defer Close（deferwrap1→RegCloseKey）；TrimSpace(Trim(v,"\"")) |
| `resolveExistingExecutablePath` | 0x1408d19e0 | [S] | linkbrowser_windows.go | 遍历 paths；Trim+TrimSpace；os.Stat 失败或 info.IsDir()（itab[0x18]）跳过 |
| `buildEnvLinkBrowserPath` | 0x1408d1ac0 | [S] | linkbrowser_windows.go | v=TrimSpace(os.Getenv(name)) 空→""；filepath.Join(append(paths,v)...) |
| `setupLauncherTray` | 0x1408b2fa0 | [S-sig] | launchertray.go | 2 字参数（rax=含 app 字段结构体、rbx=第二指针）；体依赖 Wails SystemTray 菜单构建 |
| `workspacePathsEqual` | 0x1409f4560 | [S] | workspacemigration_identity.go | Clean(a)==Clean(b)→true；否则 strings.EqualFold(Clean(a),Clean(b)) |
| `validateLauncherConfigJSONStructure` | 0x140898120 | [S-sig] | launcherconfiginput.go | 序言 []byte（rax/rbx/rcx），栈 0x170，56 行深层 schema 校验待还原 |
| `validateLauncherConfigInMemoryBudget` | 0x1408987e0 | [S-sig] | launcherconfiginput.go | validateTwoFactorStoredConfig(cfg 字段) 失败→返回；字段提取待确认 |
| `validateLauncherConfigInMemoryStructure` | 0x140898880 | [S] | launcherconfiginput.go | json.Marshal 失败→err；len>64<<20→`fmt.Errorf("%s: 配置序列化后超过 %d 字节", "CONFIG_INPUT_TOO_LARGE", 67108864)`；否则 JSONStructure(data) |
| `validateLauncherConfigForPersistence` | 0x140898960 | [S-sig] | launcherconfiginput.go | InMemoryBudget 失败→返回；否则 validateLauncherConfigIconData（签名存疑） |

## 本批订正性发现（1 处，未改）

- `validateLauncherConfigIconData`（launcherconfigicon.go:266）现落地签名为 `(cfg *LauncherConfig)`（无返回），但 `validateLauncherConfigForPersistence` 汇编里按值传 2352B（0x126×8）大结构体并透传返回值，签名疑似应为 `(cfg LauncherConfig) error`。本轮未订正（需单独核验其 asm），已在 launcherconfiginput.go 注释标注。

## 差集口径再澄清

- 差集（73→66）是**文件名对齐**指标：本轮新建 7 文件，其中 `launcherupdate.go` 是 78 函数的大文件，但本批只落地其中 `(*launcherUpdateUserError).Error` 1 个函数。故差集归零≠函数全落地。
- 真实函数落地进度以 FUNCS 为准：2667→2680（+13），真函数 2552→2565（53.9%）。
- 本轮新增发现：subagent 2888f722（desktopwidgets_audio 域）上下文耗尽未产出，该域（normalizeDesktopReminderAudio 等 6 顶层函数）仍待后续落地。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；`go test ./backend` `ok changeme/backend 0.417s`。
- **G2 count_funcs**：`FUNCS=2680 / S=1120 / S-inline=35 / S-sig=1410 / P=115 / UNMARKED=0`。
- **G3 行为**：所有 [S] 函数均以汇编字面量 + .rodata 常量解码确证（格式串/前缀/阈值/cutset 均字节级）；[S-sig] 均注明签名证据与阻断。
- **G4 review**：`readRegistryString` 的 cutset 为单字节 `"`（0x22），非空白串；`resolveExistingExecutablePath` 的 IsDir 经 itab[0x18] 方法调用确证；`validateLauncherConfigInMemoryStructure` 阈值 64<<20 经 int64 常量 67108864 确证。
