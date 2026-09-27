# 批次 160 验收（launcherconfig_runtime.go 配置规范化：4 函数 [P]→[S]）

日期：2026-09-25
子批次：书签收藏路径 / 搜索范围循环默认 / 默认浏览器 ID / 按 ID 查浏览器 配置规范化链

## 目标

把 launcherconfig_runtime.go 中 4 个 `[P]` 存根（签名近似、体为零值）升级为 `[S]`
（asm 逐条对位），消除其签名不确定性：

`normalizeBookmarkFavoritePath`、`normalizeSearchScopeCycleDefault`、
`normalizeDefaultLinkBrowserID`、`findConfiguredLinkBrowserByID`。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go1.25.12 test -count=1 -p=1 ./backend` | ok (0.400s) |
| 黄金用例 | `go1.25.12 test -run 'TestNormalizeBookmarkFavoritePath\|TestNormalizeSearchScopeCycleDefault\|TestNormalizeDefaultLinkBrowserID\|TestFindConfiguredLinkBrowserByID' -v` | 5 组全 PASS |

黄金用例说明：4 函数均为纯字符串逻辑（TrimSpace / EqualFold / 字面量判等），无 Win32/DB/runtime
I/O，可全路径安全实测。

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2659 / S=1083 / S-inline=35 / S-sig=1348 / P=193 / UNMARKED=0`。

**真函数 = 1083 + 35 + 1348 = 2466 / 4754 = 51.9%**（相对批次 159 的 2462 增 +4 [S]，P 197→193）。

重建产物 SHA256：`BCAE10E24F51222C0B2B186F159D1299CEF97975A48F0CCC4BA86A202A597610`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细（4 函数 [S]，`backend/launcherconfig_runtime.go`）

| 函数 | VA | 语义 |
|---|---|---|
| normalizeBookmarkFavoritePath | 0x140882ee0 | TrimSpace(SourcePath)→空返零态→FolderPath 归一化→返 {source, folder} |
| normalizeSearchScopeCycleDefault | 0x1408886a0 | TrimSpace→normalizeSearchScope→逐元素判等命中返、未命中返空 |
| normalizeDefaultLinkBrowserID | 0x140883920 | TrimSpace→空返空→`__system_default__` 原样→EqualFold 命中返 TrimSpace(ID) |
| findConfiguredLinkBrowserByID | 0x140883aa0 | TrimSpace→空返 (zero,false)→EqualFold 命中返 (b,true)→未命中 (zero,false) |

## G4 关键实证结论

1. **normalizeBookmarkFavoritePath 签名确证**：4 寄存器 (rax,rbx,rcx,rdi)=BookmarkFavoritePath
   两 string 字段（SourcePath/FolderPath），返 4 字（trimmed SourcePath + normalizeBookmarkFolderPathValue(FolderPath)）；
   空 SourcePath 返全零结构（asm `test rbx / je` 后 xor 四寄存器）。
2. **normalizeSearchScopeCycleDefault 循环语义**：normalizeSearchScope(trimmed) 得 normalized，
   逐 16B 字符串元素 `cmp len + memequal` 判 `element == normalized`，命中返 normalized，否则返空
   （asm 0x140888720 cmp / 0x14088873a memequal / 0x14088876a 空返）。
3. **normalizeDefaultLinkBrowserID 18B 字面量实证**：`cmp rbx,0x12` + memequal(0x140c59bfa, 18)
   解码为 `__system_default__`（read_gostring 实证），命中返原字面量——系统默认浏览器哨兵 ID。
4. **findConfiguredLinkBrowserByID 返回签名修正**：由 `*LinkBrowser` 修正为
   `(LinkBrowser, bool)`——asm 返 LinkBrowser 0x58=88B 栈结构 + bool eax（命中 eax=1+duffcopy，
   未命中/空 eax=0+duffzero）；调用点 openLinkWithPreferences 0x1407690d7 栈拷贝实证。
5. **LinkBrowser 元素布局对齐**：0x58=88B = ID/Name/Kind/Path 四 string（各 16B）+ Args []string
   （24B），与 types_app.go:75 定义一致；循环取 [rcx]=ID.ptr 后 duffcopy 余体。

## 残留 [P] / 未落地（不触及）

launcherconfig_runtime.go 剩余 [P]：populateAppIcons、startMenuLaunchLogKeys、
loadOrCreateLauncherConfig、buildConsoleItemID、loadLauncherConfigOrDefaultIfMissing、
mergeStartMenuLaunchLog、normalizeStartMenuLaunchLog、normalizeBackgroundPreference、
normalizeTagCatalogWithDefault、normalizeFileLocatorConfig、readLauncherConfigFile、
normalizeConsoleItem（12 个，多为加载/合并链，留待后续批次）。

## 已知偏差（诚实记录）

normalizeBookmarkFavoritePath 依赖 normalizeBookmarkFolderPathValue（[S-sig] 返空）；
normalizeSearchScopeCycleDefault 依赖 normalizeSearchScope / normalizeSearchScopeCycle
（[S-sig] 返空/返 nil）——两函数的 [S] 体 asm 直译正确，但全路径行为受依赖存根约束，
黄金用例仅验证不依赖存根语义的分支。
