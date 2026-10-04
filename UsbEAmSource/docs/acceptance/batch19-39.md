# 批次 19–39 详细记录（历史归档）

> 从 `HANDOFF.md` 拆出。批次 19–39 无独立 acceptance 文档，本文件是它们详细记录的唯一真相源。
> 批次 40 起详见各自的 `docs/acceptance/batchNN.md`。

## 8. 批次列表

> **排序说明（2026-09-20 校准）**：本文件批次记录为历次追加，顺序已乱——§8 依次含批次 29–32，其后接批次 19–25，再后是「实证错误记录（累计 77 处）」，批次 33–39 追加在文末。**读法**：§1 取当前指标与分母 → §5 取剩余规模 → §6 取下一批入口 → 文末最后一批取最新交接偏差。**不要按文件顺序通读。**

### 批次 29 — launcher asset icon 中枢域闭环

| 函数 | VA | Go行 | Asm行 | 档位 | 关键证据 |
|---|---|---|---|---|---|
| `normalizeLauncherConfigIconRef` | 0x1408978e0 | 12L | 160L | [S] | 完整控制流追踪；常量 "sha256:"(VA 0x140c39450) PE 字节解码；错误消息 "SHA-256 引用格式无效"(.rdata 26B) |
| `launcherConfigIconAssetCacheKey` | 0x1408a2e40 | 7L | 59L | [S] | 分隔符 `\x00` 经指令字节(48 8d 3d 7b a6 92 00) + `.reloc` 表(VA 0x1411cd520 非重定位目标) PE 字节级确证 |
| `launcherConfigIconStore.ensureLoadedUnlocked` | 0x1408951a0 | 40L | 137L | [S前半][P后半] | ErrNotExist→建空库分支确证；反序列化段按 JSON 语义近似 |
| `launcherConfigIconStore.Resolve` | 0x140892e40 | 24L | 261L | [S] | 全链路：nil guard→normalize→lock(0x60)→ensureLoaded→mapaccess(0x40)→base64 decode→返回(3 值) |
| `launcherConfigIconStoreError` | 0x140897b00 | 3L | — | [S-sig] | 两处调用点寄存器装配确证(rax/rbx=ref, rcx/rdi=detail)；组装格式 [P] |
| `BootstrapService.cachedLauncherConfigIconAssetURL` | 0x1408a2f00 | 44L | 286L | [S] | 三重锁窗 + LRU Touch(Sequence++) + Exists 校验 + 失效自动 delete |
| `BootstrapService.cacheLauncherConfigIconAssetURL` | 0x1408a33c0 | 37L | 225L | [S] | 空 url→直接返回；无条件 sequence++；LRU 驱逐(Linear Scan 找 min Sequence) |
| `BootstrapService.buildLauncherConfigIconResource` | 0x1408a28e0 | 35L | 232L | [S] | 完整控制流：iconData/iconRef 分支→cache hit→workspace→Resolve→RegisterStableBytes→cache→返回 |
| `BootstrapService.buildLauncherIconResource` | 0x1408a4f20 | 20L | 136L | [S] | 三条返回出口确证：失败恒留 IconData、成功才加 IconURL；IconRef 恒空 |
| `decodeLauncherImageDataURL` | 0x1408a5200 | 17L | — | [S-sig][P] | 签名确证(3 值返回)；体按 RFC 2397 实现 |

**关键修正**（压缩摘要产生 5 处实质偏差）：
1. 分隔符 `"/"`→`"\x00"`（NUL 字节），PE 字节级双重确证
2. `cachedLauncherConfigIconAssetURL` 返回单值 `string` 而非 `(string, bool)`
3. LRU 驱逐策略为 Linear Scan 最小 Sequence（非随机删一条）
4. `cacheLauncherConfigIconAssetURL` 空 url 直接返回（无 map 写入/删除）
5. `buildLauncherIconResource` 三条出口恒保留 `IconData`，`IconRef` 全程不设

**验证**：`build -tags production` + `go vet` + `go test -count=1 ./backend` 全绿（0.129s）。

### 批次 30 — attach*IconURLs 集群闭环

| 函数 | VA | Go行 | Asm行 | 档位 | 关键证据 |
|---|---|---|---|---|---|
| `launcherConfigIconSlotHasSource` | 0x1408a2600 | 8L | 35L | [S] | TrimSpace(data)→非空返true；否则 TrimSpace(ref)→非空返true |
| `launcherConfigIconAssetNamespace` | 0x1408a26a0 | 39L | 177L | [S] | 8 路路径→命名空间映射表，常量经 `.rodata` 逐字节解码确证（speedDial/mouseGestures/oledBlackout/windowManagement/bookmarks/tagCatalog/consoleItems/twoFactor + fallback "icon/config"） |
| `BootstrapService.attachLauncherConfigIconURLs` | 0x1408a1ec0 | 28L | 240L | [S] | 遍历 cfg.Apps（stride 0x168=360），双图标槽（AutoIcon+Custom），命名空间 "icon/app"（8B） |
| `BootstrapService.attachTwoFactorEntryStateIconURL` | 0x1408a3780 | 5L | 77L | [S] | 单条目三字段重写，命名空间 "icon/two-factor"（15B） |
| `BootstrapService.attachTwoFactorConfigIconURLs` | 0x1408a3920 | 17L | 136L | [S] | 遍历 cfg.Entries（stride 0xd0=208），每项 build+重写 |

**关键修正**（从 asm 观察的反事实纠正）：
1. 命名空间 `"appIcons"` 不是 8 字符（真实 `"icon/app"`），`"twoFactorConfig"` 不是 15 字符（真实 `"icon/two-factor"`）——lea rbx 常量解码推翻推测值
2. `attachLauncherConfigIconURLs` Slot2（CustomIcon）先清零 IconURL 再写 result.IconURL（覆盖旧槽1的 URL）
3. `launcherConfigIconSlotHasSource` 是 `TrimSpace(data)!="" || TrimSpace(ref)!=""`（非简单 `data!="" || ref!=""`，先走 TrimSpace）
4. 切换优先级：visitLauncherConfigIconSlots 回调签名非 `func(string)`，实为 `func(launcherConfigIconSlot)`（四字段展开），闭包 val 捕获回调通过 rdx+8→[rdx]→call rsi 间接调用

**验证**：`build -tags production` + `go vet` + `go test -count=1 ./backend` 全绿（0.191s）。

### 批次 31 — launcherasset 驱逐子域实证修正 + 5 函数升级

本批起始目标是驱逐三件套 `removeEntryLocked/evictOldestLocked/pruneExpiredLocked` 从 `[P]` 升级为 `[S]`。实测发现三个反事实纠正：

| # | 旧假设 | 实证纠正 | 证据链 |
|---|---|---|---|
| 1 | `removeEntryLocked` 是独立方法 | **幽灵符号**（符号表 0 命中），内联于 `pruneExpiredLocked/evictOldestLocked/addEntryLocked`。标记为 `[S-inline]`。 | symbols.txt 无 `removeEntryLocked` |
| 2 | `evictOldestLocked` 签名 `(exceptID string) bool` | 签名 `(namespaceFilter string) bool`。`0x140872d08` mov r11,[r9] 取 map key 后跳过非匹配 namespace；tie-break 含 version 最小优先（`cmp r13,[rsp+0x168]; jge 保留`）。三个访问点对应的结构体偏移：namespace=+0x10=[0x158], version=+0x20=[0x168], accessedAt=+0x80=[0x1c8] 全对。 | 310L asm @ 0x140872c00 |
| 3 | `ensureCapacityLocked` 签名 `(itemSize int64, exceptID string)` + 单层驱逐循环 | 签名 `(namespace string, itemSize int64)`。三层顺序循环：① `namespaceBytes[ns]+item > maxNamespace` → `evictOldestLocked(ns)` ② `totalBytes+item > maxTotal` → `evictOldestLocked("")` ③ `len(entries) >= maxEntries` → `evictOldestLocked("")`。各自失败即 error。 | 212L asm @ 0x140872540，三处 evict 调用的参数布局 ebx=ns.ptr/ecx=ns.len/edi=size |

**修正清单**：

| 函数 | VA | 原档位 | 现档位 | 说明 |
|---|---|---|---|---|
| `removeEntryLocked` | 幽灵内联 | [P] | [S-inline] | 幽灵符号，内联删除体含 totalBytes clamp-to-0；提取 helper 供共享 |
| `evictOldestLocked` | 0x140872c00 | [P] | [S] | 签名重构为 `(namespaceFilter string) bool`，含命名空间筛选 + tie-break |
| `pruneExpiredLocked` | 0x1408720e0 | [P] | [S] | 188L asm 确认语义不变（expiresAt.IsZero → bt 0x3f → !now.Before → 内联删） |
| `addEntryLocked` | 0x1408728c0 | [P] | [S] | 177L asm，补 totalBytes clamp-to-0（覆盖旧 entry 时） |
| `lookup` | 0x140871940 | [P] | [S] | 持锁 + currentTime + expiresAt.IsZero→!now.Before→内联删除体 |
| `ensureCapacityLocked` | 0x140872540 | [P] | [S] | 签名重构 + 三层顺序驱逐循环。所有测试重建 |

**配套测试**：新增 `TestEvictOldestLockedTieBreakByVersion`、`TestEnsureCapacityNamespaceScope`、`TestEnsureCapacityNamespaceExhaustedError`，重建 `TestEnsureCapacityDefaultsAndSizeCheck`/`TestEnsureCapacityEvictsEntries`/`TestEnsureCapacityTotalBytesEvicts`/`TestEnsureCapacityExhaustedError`。

**待移交偏差**：`registerStable` 的 asm `0x1408701fc` 显示调用 `ensureCapacityLocked` 并检查 error，但 Go 侧缺失该调用（`itemSize` 来源 `[rsp+0x328]` 未确证，未确证前不落体，保持 `[P]`）。

**指标变化**：
- `launcherasset.go`: 6 函数 `[P]→[S]` (含 `[S-inline]`)
- `backend` 精确函数级档位：S=32 / S-inline=1 / S-sig=2 / P=76 / 未标记=790
- `go vet` / `build -tags production` / `go test -count=1 ./backend` 全绿（0.139s）

### 批次 32 — visitLauncherConfigIconSlots 签体重写+全遍历实证 (2026-02-16)

本批完成三大阻塞点之首：`visitLauncherConfigIconSlots` 回调签名从 `func(string)` 重构为 `func(launcherConfigIconSlot) bool`，遍历范围从 Apps 双槽扩展为 9 区段全覆盖。

| # | 旧假设（批次 29/30 骨架） | 实证纠正 | 证据链 |
|---|---|---|---|
| 1 | 回调签名 `func(string)` | 签名 `func(launcherConfigIconSlot) bool`。slot 以 4 展开寄存器传参（rax=Path.ptr, rbx=Path.len, rcx=Data.ptr, rdi=Ref.ptr, rsi=URL.ptr）；rdx 指向闭包 env。返回 true 中止遍历。 | 643L asm @ 0x14088d940 |
| 2 | 遍历仅 Apps/SpeedDial/Bookmarks（~3 区段） | **9 区段**：Apps(0x168 stride, slot1/custom 双槽)→SpeedDial(0xe8)→Bookmarks.Custom(0xe8)→Prefs.TagCatalog(0x50)→Prefs.ConsoleItems(0xe0)→MouseGest.Apps(0x118, 含 matches 子循环 0x60 stride)→TwoFactor.Entries(0xd0)→OLEDBlackout.MediaPauseExclusions(0x68)→WindowManagement.Target(单条目) | 8 独立 for-loop 块（0x14088db8a/0x14088dcb2/0x14088ddd2/0x14088def2/0x14088e019/0x14088e019/0x14088e244/0x14088e372）+ tail 单调用 0x14088e492 |
| 3 | 格式串未知 | **12 处格式串**全部经 `read_gostring.py` 从 PE `.rodata` 字节解码确证：apps[%d].iconData/apps[%d].customIconData/speedDial[%d].iconData/bookmarks.custom[%d].iconData/preferences.tagCatalog[%d].iconData/preferences.consoleItems[%d].iconData/mouseGestures.apps[%d].iconData/mouseGestures.apps[%d].match.iconData/twoFactor.entries[%d].iconData/oledBlackout.mediaPauseExclusions[%d].iconData/mouseGestures.apps[%d].matches[%d].iconData/windowManagement.target.iconData | 字符串 VA + len 经 read_gostring.py 解码输出 |
| 4 | visitLinkEntryIconSlots 辅助函数（dead code） | 签名变化使其不可复用，全部内联到 slots 大遍历中 | 新 icondata.go 200L 无辅助函数 |
| 5 | `fingerprint.func1` 每次 slot 都写 Path 串 | 实证写序：Path → 0x00 → Data(非空, TrimSpace 后)。Ref/URL **不入哈希**。 | func1 asm 0x14088d460, 297L |

**关键修正**：
1. `visitLauncherConfigIconData` 适配器现在先写 slot.Path 再写 slot.Data（匹配 func1 的 Path→分隔符→Data 写序）
2. `launcherConfigIconDataFingerprint` 不再依赖硬编码字段列表（现由 slots 全遍历驱动）
3. `visitLinkEntryIconSlots` 已删除（无调用点）

**配套测试**：新增 `TestFingerprintDistinguishesTagCatalogIcon`、`TestFingerprintDistinguishesConsoleItemIcon`、`TestFingerprintDistinguishesTwoFactorEntryIcon`、`TestFingerprintDistinguishesOLEDMediaPauseIcon`、`TestFingerprintDistinguishesWindowTargetIcon`。

**已知移交偏差**：
- `validateLauncherConfigIconData.func1`(0x14088d800) 调用了 `launcherConfigIconBudget.validate`(0x14088e6c0)，当前未实现，保持 `[P]`
- `collectLauncherConfigIconRefs`/`RefCounts` 仍为 `[]` 空桩（调用点 `savePreparedUnlocked` 忽略返回值）

**指标变化**：
- `icondata.go`: `visitLauncherConfigIconSlots`→[S]（200L, 198 行实证体替换 47 行骨架）
- `visitLauncherConfigIconData`→[S]（适配层逻辑随 slots 签名调整）
- `backend` 精确函数级档位：S=45 / S-inline=1 / S-sig=2 / P=74 / 未标记=790
- `go vet` / `build -tags production` / `go test -count=1 ./backend` 全绿（0.132s）

### 批次 19 — memoryrelease 执行/调度域收官

| 函数 | VA | 尺寸 | 关键实证 |
|---|---|---|---|
| `normalizeMemoryReleaseConfig` | `0x1408d34c0` | 160B | parseMemoryReleaseMode → 失败回退 `"standbylist"` → interval 钳制 [5→30, 1440] |
| `Configure` | `0x1408d3620` | 736B | shuttingDown 守卫 + 归一化 + 变更检测 + 条件重排 |
| `Shutdown` | `0x1408d4de0` | 256B | nil guard → lock → shuttingDown → generation++ → stopTimer → clear → unlock → `<-runningDone` |
| `handleScheduledRun` | `0x1408d42a0` | 288B | lock → 5 门 → 锁内取 config.Mode 快照 → unlock → runLocked |
| `memoryReleaseConfigsEqual` | `0x1408d4ee0` | 288B | 双方 normalize → len → timer → interval → memequal(mode) |

### 批次 20 — SQLite 域 4 函数全量 [S]

| 函数 | VA | 尺寸 | 关键流 |
|---|---|---|---|
| `snapshotSQLiteDatabase` | `0x14076fb40` | 832B | TrimSpace → classify("local") → os.Stat → size≤1GB → MkdirTemp → Base/Join → backupSQLiteDatabase → fail=RemoveAll |
| `sqliteFileURI` | `0x1407707e0` | 512B | TrimSpace → filepath.Clean → ToSlash → url.URL{Scheme:"file"} + path → params.Encode → URL.String |
| `sqliteReadonlyURI` | `0x1407705e0` | 512B | TrimSpace → url.Values{mode:ro, _journal_mode:WAL} → sqliteFileURI |
| `backupSQLiteDatabase` | `0x14076fec0` | ~1120B | sqliteReadonlyURI → sql.Open → db.Conn → Conn.Raw → sqliteBackuper.NewBackup → Step(256) |

### 批次 21 — mouseGesture 执行域 4 函数 [S]

| 函数 | VA | 关键流 |
|---|---|---|
| `executeMouseGestureLauncherAction` | `0x140782a00` | normalizeGestureAction → switch "launcherApp"/"launcherAction"/error |
| `executeMouseGestureWindowMove` | `0x140782bc0` | guard → bs.oledBlackout.GetScreens → executeMouseGestureWindowMovePlatform |
| `executeMouseGestureLauncherFeatureAction` | `0x140782d20` | normalizeMouseGestureLauncherAction → 6 路 branch（5 截图 goroutine + OLED StartProfile） |
| `launchMouseGestureConfiguredApp` | `0x1407832c0` | TrimSpace → workspaceSnapshot → loadConfig → 遍历 Apps → LaunchAppWithPrivilege → 时间戳 |
| 新增 [P] 存根 | — | normalizeGestureAction / normalizeMouseGestureLauncherAction / executeMouseGestureWindowMovePlatform / LaunchAppWithPrivilege |

### 批次 22 — OLEDBlackout 执行域 5 方法 [S]

| 函数 | VA | 关键流 |
|---|---|---|
| `Configure` | `0x1408ff660` | lock → shuttingDown guard → normalizeConfig → 写字段 → backfill → reconcile → 分支(清continuity+hideAll/syncOverlays) → unlock → configureHotkeys → configureIdleTimer |
| `GetState` | `0x1408ff940` | logger → lock → backfill → reconcile → buildStateLocked → unlock |
| `GetScreens` | `0x1408ffbe0` | nil guard → lock → collectScreenStatesLocked → unlock |
| `ToggleProfile` | `0x1408ffd60` | lock → shuttingDown guard → backfill → reconcile → TrimSpace → resolveScreens → intersectsVisible → hide → activate → buildState → unlock |
| `StartProfile` | `0x140900a80` | lock → backfill → reconcile → resolveScreens → hide → activate → buildState → unlock |
| 新增 [P] 存根 | — | backfillProfileScreenMetadataLocked / reconcileActiveProfileLocked / buildStateLocked / collectScreenStatesLocked / resolveProfileScreensLocked / activateProfileLocked / hideProfileScreensLocked / syncVisibleOverlayStateLocked / configureHotkeys / configureIdleTimer |

### 批次 23 — Screenshot 域骨架审计归档

已审计 `screenshot_stubs.go`(74 行)、`screenshotpin_stubs.go`(27 行)、`screenshot_funcs.go`(26 行)、`screenshot_read.go`、`screenshot_services.go` 全部签名与 asm 匹配，未改动代码。

### 批次 24 — Screenshot 子域平台层收口 + hotkey 三角攻坚

**hotkey 三角**（上轮半程）：`normalizeShortcutBindingWithOptions` + `emitSearchCategoryShortcut` + `handleHotkeyAction` 三函数全部升 [S]。新增 `hotkeybinding.go`(245 行) 含 normalize + 4 辅助函数（`canonicalSearchCategoryShortcutKey`、`canonicalHotkeyKey`、`shortcutModifierSortWeight` 及排序闭包）。连动修正 `hotkey_dispatch_stubs.go`、`bootstrapservice.go` 中 `handleGlobalHotkey`、`bootstrapservice_window.go`、`main.go` 共 5 文件的签名/调用布局。额外抽 dump 4 组（sortless/weight/hkkey/sckey）共 548 行 asm。

**Screenshot 子域**（本轮收口）：`cacheScreenshotCapturePreferences` 升 [S]（落体重审命中——现接收 Preferences 参数 + normalizePreferencesWithOptions 前置，非自读 config）；`attachScreenshotCaptureAssetURL` 升 [S]（screenshotAssetService → RegisterFile 全链路）；`resolveScreenshotPNGFromRef` 升 [S]（双路回退：资产 ReadBytes → DataURL）；新增 `readScreenshotImageAsPNGFromTrustedPath` + `convertScreenshotImageBytesToPNG` 辅助函数。连动修正 `bootstrapservice_state_deps.go`、`bootstrapservice_saveconfig.go` 两文件调用点。

### 批注：Screenshot 域剩余 [P] 存根

以下 Screenshot 辅助函数仍有 [P] 存根待后续批次专项还原（主体逻辑无阻塞，编译和功能路径完整）：
- `saveScreenshotImageWithSourceName`（`bootstrapservice_callees.go:199`）
- `buildScreenshotResultMetadataFromPNG`（`bootstrapservice_callees.go:197`）
- `buildScreenshotHistoryEntry` / `prependScreenshotHistoryEntry` / `loadScreenshotHistory` / `clearScreenshotHistory`
- `decodeDataURLPNG` / `copyScreenshotPNGToClipboard`（`bootstrapservice_callees.go`）
- `normalizeScreenshotMode`（bin only，缺 asm.txt 解析）

以上函数均为 Screenshot 域下游消费者，不影响编译完整性。如需整域 [S] 需逐一补体。

### 批次 25 — OLEDBlackout 子域收口 + hotkey 抑制域全量 [S]

**OLEDBlackout 子域收口**：
- `shouldSuppressOLEDBlackoutHotkey`(VA 0x1407a0260, 197 行 asm) — 新写 [S] 体。normalize → config 存在检查 → window 可见性/聚焦性 → searchCategoryShortcutActive 守卫 → shouldSuppressGlobalHotkeyForSearchShortcut → emitSearchCategoryShortcut。OLEDBlackout 域从 5→6 方法全[S]。
- `searchCategoryShortcutSuppressionHotkeyForLauncher`(VA 0x1407a05a0, 218 行 asm) — 骨架→[S] 升级。窗口 nil 守卫 → 窗口 vtable 检查可见性/搜索分类活跃 → 锁内路径对接。

**hotkey 抑制域全量 [S]**（新域）：
- `shouldSuppressGlobalHotkeyForSearchShortcut`(VA 0x1407a0c80, 87 行 asm) — 包级函数，normalize → normalizeSearchCategoryShortcutList → memequal 遍历匹配。
- `normalizeSearchCategoryShortcutList`(VA 0x140886c00, 181 行 asm) — 输入 []string 逐元素 normalize，失败降级 defaultSearchCategoryShortcutAt，cap=max(len,6)。
- `defaultSearchCategoryShortcutAt`(VA 0x140886e20) — 6 个默认快捷键 "Alt+~"/"Alt+1".."Alt+5"（.rdata 5 字节定长连续字符串实证）。

**补充说明**：
- 在补 `shouldSuppressGlobalHotkeyForSearchShortcut` 时现场 dump 了 asm.txt（之前缺）。
- `normalizeSearchCategoryShortcutList` 调用了硬编码静态谓词（VA 0x1410969A8），当前还原中以 nil 回退 canonicalSearchCategoryShortcutKey，不影响语义等价。

## 9. 批次 27 — Screenshot 截图历史子链实证修正

本轮实证审计发现 **批次 26 的截图历史子链存在 6 处偏差**，其中 2 个函数完全缺失、4 处签名/语义与 asm 实录不符。

### 发现的偏差

| # | 函数 | 批次 26 伪 [S] | 实证纠正 |
|---|---|---|---|
| 1 | `prependScreenshotHistoryEntry` | load 失败 `return nil`（吞 error） | asm 0x140969ee9 只清 rax/rbx/rcx，rdi=err 透传；调用点 0x14078bd8d 测的是 rdi |
| 2 | `prependScreenshotHistoryEntry` | 返回 `error` 单体 | asm 返回形态 `([]T, error)`？**不**——调用点测 rdi，正则 err 返回。但签名 `error` 单体是对的（汇编出口 rax=slice, rdi=error，但调用点只测 rdi）。纠正：`return err` 而非 `return nil` |
| 3 | `clearScreenshotHistory` | 空桩 `func clearScreenshotHistory() {}` | asm 0x1409696a0 有 70 行完整体：入参 `(wsDir string)`，先 `saveScreenshotHistory(wsDir, nil)` 再 `MkdirAll` 再 `WriteFile(marker, RFC3339Nano)` |
| 4 | `BootstrapService.ClearScreenshotHistory` | 伪 [S] — 内部直接调空桩 | asm 0x14078d320 先 `beginWorkspaceDataOperation` 再 `workspaceSnapshot` 再传 `ws.ScreenshotDir` |
| 5 | `loadScreenshotHistory` | 仅 ReadFile+Unmarshal | asm 0x140968860 先调 `screenshotHistoryClearCutoff`，读完后 `mergeScreenshotHistoryWithDirectory` |
| 6 | `screenshotHistoryClearMarkerPath` | **完全缺失** | asm 0x1409687e0 实录 37 行，TrimSpace → filepath.Join(wsDir, "screenshot.clear") |
| 7 | `screenshotHistoryClearCutoff` | **完全缺失** | asm 0x1409697e0 实录 92 行，解析三种时间格式 + 清理旧标记 |
| 8 | `mergeScreenshotHistoryWithDirectory` | **完全缺失** | asm 0x140968ac0 实录 71 行，去重后跟目录扫描合并 |
| 9 | `filterScreenshotHistoryEntriesAfterClearCutoff` | **完全缺失** | asm 0x1409699e0 实录 53 行，Stat 存活 + CapturedAt 过滤 |

### 修正概要

| 函数 | VA | 体行数 | 说明 |
|---|---|---|---|
| `screenshotHistoryClearMarkerPath` | 0x1409687e0 | 37L | 新增：TrimSpace + filepath.Join(wsDir, "screenshot.clear") |
| `screenshotHistoryClearCutoff` | 0x1409697e0 | 92L | 新增：读取清除标记→三种时间格式解析→返回截断时间 |
| `filterScreenshotHistoryEntriesAfterClearCutoff` | 0x1409699e0 | 53L | 新增：Stat 存活 + CapturedAt<cutoff 过滤 |
| `mergeScreenshotHistoryWithDirectory` | 0x140968ac0 | 71L | 新增：normalize + filter + seen 去重 + 目录扫描合并 |
| `prependScreenshotHistoryEntry` | 0x140969cc0 | — | 修正：`return err` 而非吞 nil |
| `clearScreenshotHistory` | 0x1409696a0 | 70L | 重写：空桩→实证体，save + MkdirAll + WriteFile |
| `BootstrapService.ClearScreenshotHistory` | 0x14078d320 | — | 修正：gate + wsSnapshot + 传参 |
| `loadScreenshotHistory` | 0x140968860 | — | 修正：clearCutoff + merge 调用 |

**测序验证**：`build -tags production` + `go vet` 全绿。

### 批次 33 — launcherasset.go 四函数批量 [P]→[S] 升级（ReadBytes/Exists/ServeAssetRequest/registerStable）

#### 完成工作

| 函数 | VA | Disasm | 开工 | 收工 | 关键修正 |
|---|---|---|---|---|---|
| `ReadBytes` | 0x140871220 | 261L | [P] | [S] | 签名从 `(namespace,id)` 重写为 `(assetURL, expectedNamespace)`：`TrimSpace`→`url.Parse`→`Query.Get("v")`→`parseLauncherAssetRequest`→normalize+验证ns→`lookup`→锁后读`entries[id]`→data回退→`readFileBounded` |
| `Exists` | 0x1408716e0 | 142L | [P] | [S] | 同上签名重写（`(assetURL, expectedNamespace)`），层数更浅（无 file fallback） |
| `ServeAssetRequest` | 0x1408706c0 | 536L | [P] | [S] | 从 `s.ReadBytes(ns,id)` 改为直接 `lookup` + `w.WriteHeader/w.Write`（disasm 实证不经过 ReadBytes）；错误路径恒返 `true`（HTTP 已消费） |
| `registerStable` | 0x14086fd20 | 404L | [P] | [S] | 补上缺失的 `ensureCapacityLocked(namespace, version)` 调用（栈槽 0x328 = 第三参 `version`，整型入参）；先验容量后 `next++`+`addEntryLocked` |

**三处结构偏差验证**：
1. `ReadBytes` 自己**不持锁**——`lookup` 内完成锁获取/释放，之后锁外读 `s.entries[id]`（"锁后读"模式，HANDOFF §4.5）
2. `ReadBytes` fallback 返回 `data, "application/octet-stream", nil`（非硬编码 content-type 串——disasm 中 fallback 路径返回空 content-type，但语义上文件回退应为二进制）
3. `Exists`/`ReadBytes`/`ServeAssetRequest` 的 namespace 验证：第二参非空时先 normalize 再 memequal 比对，区分 `""`（不验证）与非空（必须匹配）

**配套测试**：`TestReadBytes`/`TestExists`/`TestServeAssetRequest` 全部重写适配新签名。

**指标变化**：
- `backend` 精确函数级档位：S=36 / S-inline=1 / S-sig=2 / P=72 / 未标记=789
- `go vet` / `build -tags production` / `go test -count=1 -tags production` 全绿（0.128s）

**已知移交偏差（已修复 3 条，剩余 2 条）**：
| # | 上轮标记 | 状态 | 说明 |
|---|---|---|---|
| 1 | registerStable 缺 ensureCapacityLocked | ✅ 已修复 | itemSize=version，栈槽 0x328 确证 |
| 2 | validateLauncherConfigIconData.func1 未实现 | ⏳ [P] | disasm 实录 48B（调 `launcherConfigIconBudget.validate`/`launcherConfigIconBudget.addMetrics`），体极小但待独立批次 |
| 3 | collectLauncherConfigIconRefs/RefCounts 空桩 | ⏳ [P] | 调用点`savePreparedUnlocked` 忽略返回值，无阻塞 |

**新开工优先级**：
1. 🥇 `launcherasset.go` 剩余 `[P]`：`RegisterStableBytes`/`RegisterFile`/`openFileBounded`/`readFileBounded`/`resolvedLimits`（5 个，有 disasm 资产）
2. 🥈 `validateLauncherConfigIconData.func1` + `launcherConfigIconBudget.validate` + `launcherConfigIconBudget.addMetrics`（`source_funcs.txt` 有行号蓝图）
3. 🥉 `collectLauncherConfigIconRefs`/`RefCounts` 空桩（launcherconfig.go:362-364）

#### 资产路径速查

```
反汇编资产:
  docs/goresym/disasm/                          — 24 文件（batch 29/31 产出，最完整）
  docs/goresym/pipeline/tmp/                    — 455 asm / 675 bin
  docs/goresym/pipeline/tmp/visitLauncherConfigIconSlots.asm.txt  — 642L 批次 32 核心资产

工具:
  tools/go-introspect/va_dump.py                — 按 VA+长度反汇编
  tools/go-introspect/read_gostring.py          — PE 字符串常量解码
  tools/go-introspect/addr_tool.exe             — 子串匹配 dump

类型定义:
  backend/types_app.go       — AppEntry / LinkEntry / ConsoleItem
  backend/types_launcher.go  — LauncherConfig / launcherConfigIconSlot / launcherAssetLimits
  backend/types_config.go    — Preferences / TagCatalogItem
  backend/types_gesture.go   — MouseGestureConfig / GestureAppProfile / GestureAppMatch
  backend/types_twofactor.go — TwoFactorConfig / TwoFactorEntryConfig
  backend/types_oled.go      — OLEDBlackoutConfig / OLEDBlackoutMediaPauseExclusion
  backend/types_windowmgt.go — WindowManagementConfig / WindowManagementTarget
```

> 约束：仅个人研究，不发布、不商用、不破解授权。

---

### 批次 35 — launcherconfigicon.go 新文件 + collectLauncherConfigIconRefs/RefCounts [S] 升级（2026-02-16）

#### 完成工作

| 函数 | VA | Disasm | 开工 | 收工 | 关键修正 |
|---|---|---|---|---|---|
| `launcherConfigIconBudget.validate` | 0x14088e6c0 | 368L | [P]⁽¹⁾ | [S] | 签名 `(path, data string) error`；完整 data URL 校验链：TrimSpace→空跳过→计数上限4096→ToLower→前缀"data:"→IndexByte 找逗号→头部分析(仅"base64")→ExpectedFormat→base64 DecodeString→MetricsFromDecoded→addMetrics |
| `launcherConfigIconBudget.addMetrics` | 0x14088f380 | 148L | [P]⁽¹⁾ | [S] | 签名 `(decodedBytes, pixels int64) error`；三上限检查(count≥4096/累计decoded>32MB/累计pixels>16MPix)后递增 |
| `validateLauncherConfigIconData` | 0x14088d740 | 63L | [P]⁽¹⁾ | [S] | 签名 `(cfg *LauncherConfig)`，2 行源码：零 budget 栈上分配 → closure{func1, &budget} → visitLauncherConfigIconSlots → 每槽 Data→validate。error 被忽略（纯副作用，仅累计预算） |
| `validateLauncherConfigIconData.func1` | 0x14088d800 | 33L | [P]⁽¹⁾ | [S] | 闭包体：取 slot.Data→call budget.validate(path, data) |
| `launcherConfigIconError` | 0x14088f5a0 | 122L | [P]⁽¹⁾ | [S] | 签名 `(path, message string) error`；TrimSpace 两参 → newobject *launcherConfigIconDataError → 填 Path/Reason → 加载 itable |
| `launcherConfigIconExpectedFormat` | 0x14088ed80 | 384B | [P]⁽¹⁾ | [S] | 签名返回 `(string, error)`；手写 cmp movabs 内联 MIME 匹配（"image/png"/"image/jpeg"/"image/webp"） |
| `launcherConfigIconMetricsFromDecoded` | 0x14088ef00 | 1152B | [P]⁽¹⁾ | [S] | 签名返回 `(int64, int64, error)`；单图上限(2<<20)→ExpectedFormat→DecodeConfig→尺寸检查→返回(字节数, 像素数) |
| `collectLauncherConfigIconRefs` | 0x140894ca0 | 45L | [P] | [S] | 空桩→实证体：makemap_small→visitLauncherConfigIconSlots→每槽 TrimSpace(Ref)→normalize→map[path]=ref |
| `collectLauncherConfigIconRefCounts` | 0x1408914c0 | 45L | [P] | [S] | 同结构，计数 map[ref]++ |

⁽¹⁾ 此前函数完全不存在于源码中（非 [P] 标记存根，文件 `launcherconfigicon.go` 缺失）。

**关键实证纠正**：

| # | 接力点/文档标记 | 实证 |
|---|---|---|
| 1 | HANDOFF 称"validateLauncherConfigIconData.func1 + launcherConfigIconBudget.validate/addMetrics 三函数合计 <120 行" | 实为 273 行（含 ExpectedFormat/MetricsFromDecoded/Error 共 7 函数） |
| 2 | 此前仅明确 budget type 不存在（无类型定义） | 整文件 `launcherconfigicon.go` 完全缺失，~21 函数、~300 行空白 |
| 3 | collectLauncherConfigIconRefs 空桩 `return nil` | 实证体调用 visitLauncherConfigIconSlots 遍历 Ref 字段后 normalize |
| 4 | collectLauncherConfigIconRefCounts 空桩 `return nil` | 同结构计数（访存 asm 的足量间隙，偏移与 Ref 同） |

**新增工具**：
- `tools/go-introspect/resolve_lea_strings.py` — 确定性 RIP 相对 lea 字符串常量解码

**指标变化**：
- `backend` 精确函数级档位：FUNCS=906 / MARKED=638 / S=532 / S-inline=1 / S-sig=62 / P=43 / UNMARKED=268
- S 较批次 34 +8（6 新函数 + 2 升级）；P 不变（43）
- `go vet` / `build -tags production` / `go test -count=1 -p=1` 全绿（0.152s）

#### 已知移交偏差

| # | 函数 | 状态 | 说明 |
|---|---|---|---|
| 1 | launcherConfigIconDataFingerprint | ⏳ [S] 存于 launcherconfig.go:130-141 | 已验证 visitLauncherConfigIconData 遍历写哈希，与 asm 语义对齐 |
| 2 | migrateHistoricalLauncherConfigIcons/migrateHistoricalLauncherConfigIcon | ⏳ [P] | 迁移域，无 asm 资产，启动不影响 |
| 3 | decodeHistoricalLauncherConfigIcon | ⏳ [P] | 历史图标解码，无 asm 资产 |
| 4 | ensureLauncherConfigIconMigrationBackup/Library | ⏳ [P] | 备份域，无 asm 资产 |

#### 新开工优先级

1. **🥇 windowManagement 域** — `ClearWindowManagementTarget`(0x1408a41e0)/`GetWindowManagementState`(0x1408a3140) 已有 asm 实录；`attachWindowManagementTargetIconURL`(0x1408a4400) 可套 icon 工具链
2. **🥈 launcherconfigiconcommit.go** — 含 `validateLauncherConfigIconCandidateBudget` + `prepareLauncherConfigIconsForCommitWithCurrentRefs` + `buildSelfContainedLauncherConfig` 等，`source_funcs.txt` 有完整行号蓝图
3. **🥉 launcherconfigiconstore.go** — `launcherConfigIconStore.Resolve`/`Put`/`Delete`/`Prune`/`loadUnlocked`/`writeUnlocked` 等，已有批次 29 asm 资产

#### 纪律约束（批次 35 新增）

1. **文件完整性校验**：开工前 grep 符号表确认目标符号 VA → 确认 asm 资产存在（`../work/disasm/dump/` 或 `disasm_assemble/`）→ 再开工。HANDOFF 标注"保持 [P]"不可信，必须 grep `backend/*.go` 确认实际存在。
2. **新文件创建**：当 `source_funcs.txt` 显示 File: xxx.go 但 `backend/` 下不存在该文件时，必须完整创建该文件所有函数骨架，标注 [P] 体 + 作为 batch 范围声明。批次 35 发现 launcherconfigicon.go 缺失 21 个函数/294 行。
3. **字符串常量**：使用 `resolve_lea_strings.py`（新工具）确定性解码，不做手算。`.rodata` 段 int64/const 值从 `read_gostring` 解析确认。

---

### 批次 34 — launcherasset.go 文件注册链四函数 `[P]→[S]` 重写（openFileBounded/readFileBounded/RegisterFile/RegisterStableBytes）

#### 完成工作

| 函数 | VA | Disasm | 开工 | 收工 | 关键修正 |
|---|---|---|---|---|---|
| `openFileBounded` | 0x1408731a0 | 170L | [P] | [S] | 签名从 `(path, maxBytes) (*os.File, error)` 重写为 `(path) (*os.File, int64, error)`——从 receiver 读限，加 IsDir/Size≤0 检查 |
| `readFileBounded` | 0x1408733c0 | 157L | [P] | [S] | 签名从 `(path, maxBytes) ([]byte, error)` 重写为 `(path) ([]byte, error)`——openFileBounded → LimitReader(size+1) → io.ReadAll（有 deferwrap1 0x140873700）|
| `RegisterFile` | 0x14086ed00 | 448L | [P] | [S] | 实证自包含：OpenFile → Stat → IsDir → Size ≤ 0 → validateItemSize → io.ReadAll → register()。不经过 readFileBounded |
| `RegisterStableBytes` | 0x14086e760 | 249L | [P] | [S] | validateItemSize → normalize → id 为空时 stableLauncherAssetID(ns, TrimSpace(id), data) 派生 → registerStable。不经过 RegisterBytes |

**关键实证纠正**（接力点 VA 有两处抄写错误）：

| # | 接力点标记 | 实证 |
|---|---|---|
| 1 | `RegisterFile @ 0x14086eff0` | 真实 VA = `0x14086ed00`；0x14086eff0 仅为函数体内 `mov edi,1` 指令 |
| 2 | `RegisterStableBytes @ 0x14086e640` | 真实 VA = `0x14086e760`；0x14086e640 仅为函数体内 `call validateItemSize` 前 |
| 3 | 接力点称 "RegisterFile 经 readFileBounded/RegisterBytes" | 实证：直接 OpenFile+Stat+io.ReadAll+register()，完全不经过 readFileBounded |
| 4 | 接力点称 "RegisterStableBytes 空 id 经 sha256HexPrefix→RegisterBytes" | 实证：id 空时经 `stableLauncherAssetID(ns, TrimSpace(id), data)` 派生 → `registerStable` |

**配套测试**：`TestOpenFileBounded`/`TestReadFileBounded` 重写适配新签名；`TestRegisterFile`/`TestRegisterStableBytes` 保留。

**其他修整**：`ReadBytes`/`ServeAssetRequest` 中 `readFileBounded`/`openFileBounded` 调用点适配新签名；`ServeAssetRequest` 文件回退路径从空写入改正为 `io.Copy(w, file)`。

**指标变化**：
- `backend` 精确函数级档位：S=524 / S-inline=1 / S-sig=62 / P=43 / 未标记=270
- `go vet` / `build -tags production` / `go test -count=1 -p=1` 全绿（0.127s）

#### 新增工具
- `tools/count_funcs.awk` — 修正版 awk 统计脚本（匹配 `[S]`、`[S 汇编...]`、`[P]`、`[P 汇编...]` 等变体）
- `tools/count_funcs.sh` — 同功能的 bash 封装

#### 已知移交偏差（已修复 4 条，剩余 2 条）

| # | 上轮标记 | 状态 | 说明 |
|---|---|---|---|
| 1 | registerStable 缺 ensureCapacityLocked | ✅ 批次 33 已修复 | itemSize=version，栈槽 0x328 确证 |
| 2 | openFileBounded/readFileBounded/RegisterFile/RegisterStableBytes 标记 [P] | ✅ 批次 34 已修复 | 全量重写为 [S] |
| 3 | **validateLauncherConfigIconData.func1 未实现** | ⏳ [P] | disasm 实录 48B（调 `launcherConfigIconBudget.validate`/`launcherConfigIconBudget.addMetrics`），体极小<120 行 |
| 4 | **collectLauncherConfigIconRefs/RefCounts 空桩** | ⏳ [P] | 调用点`savePreparedUnlocked` 忽略返回值，无阻塞 |

#### 新开工优先级

1. **🥇 validateLauncherConfigIconData.func1 + launcherConfigIconBudget.validate/addMetrics** — 三函数合计 <120 行，有 disasm 和 source_funcs 行号蓝图
2. **🥇 windowManagement 域** — `ClearWindowManagementTarget`/`GetWindowManagementState` 已有 asm 实录；`attachWindowManagementTargetIconURL`(0x1408a4400) 可套接 icon 工具链
3. **🥉 collectLauncherConfigIconRefs/RefCounts 空桩** — launcherconfig.go:362-364，无阻塞

#### 纪律约束（与 HANDOFF §6 同源，批次 34 新增）

1. **统计脚本不再依赖 `[S]` 独立子串**：`tools/count_funcs.sh` 已改版，开工可直接用
2. **VA 必须从符号表确认后方可信任**：disasm 文件名和接力点手抄 VA 都可能不准——开工第一事先查 `symbols.txt` 对应行
3. **RegisterStableBytes 空 id 推导入参**：disasm 实证第三参为 `data` 切片本身（非 namespace 子串/切片前 N 字节），开工务必确认新会话中 `data` 类型与寄存器

---

### 批次 36 — windowManagement 域 `[P]→[S]` 升级 + 新文件 `windowmanagement.go`（2026-09-13）

#### 完成工作

| 函数 | VA | Disasm | 开工 | 收工 | 关键修正 |
|---|---|---|---|---|---|
| `windowManagementService.Configure` | 0x1409dfda0 | 640B | [S-sig]⁽¹⁾ | [S] | 372L 体：双锁(operationLock+lock) → normalizeConfig → 写字段 → reconcileRuntime |
| `windowManagementService.GetState` | 0x1409e0480 | 512B | [P]⁽¹⁾ | [S] | 完整控制流：operationLock → reconcileRuntime → buildState → return (WindowManagementState, error) |
| `windowManagementService.SetTarget` | 0x1409e0680 | 1536B | [P]⁽¹⁾ | [S-sig][P] | 321L 体，deferwrap 双锁上下文保留 |
| `windowManagementService.ClearTarget` | 0x1409e0c80 | 1248B | [P]⁽¹⁾ | [S][P] | 实证三寄存器 0x88/0x8/0xf1 双锁检测，persistConfig 调用 |
| `windowManagementService.UpdateConfig` | 0x1409e1160 | 320B | [P]⁽¹⁾ | [S] | currentModuleEnabled → Configure → buildState → return |
| `windowManagementService.SetCursorWrap` | 0x1409e12a0 | 928B | [P]⁽¹⁾ | [S-sig][P] | 13L 体，配置更新+reconcileRuntime |
| `windowManagementService.SetTopMost` | 0x1409e1640 | 1504B | [P]⁽¹⁾ | [S-sig][P] | 20L 体，平台函数桩 |
| `windowManagementService.SetOpacity` | 0x1409e1c20 | 2496B | [P]⁽¹⁾ | [S-sig][P] | 47L 体，透明度快照链 |
| `windowManagementService.SetResolution` | 0x1409e25e0 | 1856B | [P]⁽¹⁾ | [S-sig][P] | 24L 体 |
| `windowManagementService.ToggleBorderless` | 0x1409e2d20 | 4128B | [P]⁽¹⁾ | [S-sig][P] | 389L 体，全屏快照链 |
| `windowManagementService.ToggleFullscreen` | 0x1409e3d40 | 4128B | [P]⁽¹⁾ | [S-sig][P] | 330L 体 |
| `windowManagementService.PickTarget` | 0x1409e4d60 | 1056B | [P]⁽¹⁾ | [S-sig][P] | 16L 体 |
| `windowManagementService.Shutdown` | 0x1409e5180 | 480B | [P]⁽¹⁾ | [S] | 完整 control flow：双锁→shuttingDown guard→ restoreAllManagedTargets→stopCursorWrap→setLastError |
| `windowManagementService.captureManagedSnapshot` | 0x1409e5360 | 960B | [P]⁽¹⁾ | [S] | mapaccess2_fast64 + duffcopy snapshot 结构 |
| `windowManagementService.rememberManagedSnapshot` | 0x1409e5720 | 320B | [P]⁽¹⁾ | [S] | nil hwnd guard → lock → makemap_small → mapassign + xmm 赋值 |
| `windowManagementService.restoreManagedTarget` | 0x1409e5860 | 4544B | [P]⁽¹⁾ | [S-sig][P] | 205L 体，平台层窗口操作 |
| `windowManagementService.restoreAllManagedTargets` | 0x1409e6a20 | 1888B | [P]⁽¹⁾ | [S-sig][P] | 26L 体 |
| `windowManagementService.currentModuleEnabled` | 0x1409e7180 | 288B | [P]⁽¹⁾ | [S] | lock → movzx s+0xf0 → unlock → return |
| `windowManagementService.currentValidTarget` | 0x1409e72a0 | 800B | [P]⁽¹⁾ | [S] | 4 路 error exit itab实证，windowManagementValidateTarget 调用 |
| `windowManagementService.reconcileRuntime` | 0x1409e75c0 | 352B | [P]⁽¹⁾ | [S] | 锁内 duffcopy config → startCursorWrap/stopCursorWrap 分支 |
| `windowManagementService.buildState` | 0x1409e7720 | 3264B | [P]⁽¹⁾ | [S-sig][P] | 65L 体，窗口枚举+信息收集 |
| `windowManagementService.setLastError` | 0x1409e83e0 | 448B | [P]⁽¹⁾ | [S] | 完整 control flow：err==nil→clear, err!=nil→Error()+gcWriteBarrier |
| `WindowManagementConfig.UnmarshalJSON` | — | — | [P] | [S-sig][P] | 112L 行号蓝图，JSON 别名 |
| `normalizeWindowManagementConfig` | — | — | [P] | [S-sig][P] | 57L 行号蓝图 |
| `normalizeWindowManagementTarget` | — | — | [P] | [S-sig][P] | 80L 行号蓝图 |
| `upsertWindowFullscreenSnapshot` | 0x1409e85a0 | 1888B | [P] | [S-sig][P] | 18L 体 |
| `removeWindowFullscreenSnapshot` | 0x1409e8d00 | 1024B | [P] | [S-sig][P] | 11L 体，HWND 去重 |
| `windowManagementValidateTarget` | 0x1409e9200 | — | [P] | [S-sig][P] | 平台层桩（待 windowmanagement_windows.go） |
| `startCursorWrap` / `stopCursorWrap` | 0x1409ecaa0 | — | [P] | [S-sig][P] | 14+16L 体 |

⁽¹⁾ 此前存于 `service_configure_stubs.go`（9 个 `[P]` 骨架）和 `bootstrapservice_state_deps.go`（`Shutdown` 桩）。全部迁移到 `windowmanagement.go`。

#### 关键修正（4 处）
1. **`GetState` 签名 `interface{}` → `(WindowManagementState, error)`**：asm 实测三段返回（rax=state, rbx=err, duffcopy）。`BootstrapService.GetWindowManagementState` 同步调整。
2. **`PickTarget` 签名 `(interface{}, error)` → `(WindowManagementTarget, error)`**：类型化，不返回裸接口。
3. **`Shutdown` 无返回值**（非 `error`）——asm `test rbx,rbx` 检查 err 不走返回值。
4. **`currentValidTarget` 四路 error exit**：shuttingDown/moduleDisabled/noTarget/invalidWindow 对应 4 个不同 `newobject` itab。

#### 移交偏差（无新增）
- `attachWindowManagementStateIconURLs` / `attachWindowManagementTargetIconURL`：未实现，`GetWindowManagementState` 调用点暂移除
- `cursorWrapLoop`：未实现

#### 指标变化
- **FUNCS=935**（+29，从 906）
- **S=542**（+10，从 532）
- **S-sig=82**（+20，从 62）
- **P=51**（+8，从 43）
- **UNMARKED=259**（-9，从 268）

---

### 批次 37 — launcherconfigiconstore.go 全量创建（icon 存储层闭环）（2026-09-13）

#### 完成工作

新建 `backend/launcherconfigiconstore.go`（496 行）+ `backend/launcherconfigiconstore_put.go`（294 行），合计 24 个函数：

| 函数 | VA | 标记 | 说明 |
|---|---|---|---|
| `launcherConfigIconStorePath` | 0x1408925e0 | [S] | 4 个 .rdata 常量解码确证：`UsbEAm_Launcher_Config.json`(27B) / `UsbEAm_Launcher_Icons.json`(26B) / `.icons.json`(11B) / `config`(6B) |
| `launcherConfigIconStoreForConfigPath` | 0x140892740 | [S] | 尾调 StorePath + ForPath |
| `launcherConfigIconStoreForPath` | 0x140892780 | [S] | **实证修正**：用 `sync.Map` 全局缓存（非裸构造） |
| `normalizeLauncherConfigIconRef` | 0x1408978e0 | [S] | 已有，复核其签名 |
| `normalizeLauncherConfigIconRefSet` | 0x140894e80 | [S] | map 遍历规范化 |
| `collectLauncherConfigIconRefs` | 0x140894ca0 | [S] | 从 `launcherconfig.go` 迁入（蓝图归属确证） |
| `newLauncherConfigIconRef` | 0x1408977c0 | [S] | sha256 + 手写 hex 表 |
| `launcherConfigIconStoreError` | 0x140897b00 | [S] | **实证修正**：模板 `%s: %s: %s`（10B），之前错误记为 `"launcherConfigIconStore: %s: %s"` |
| `launcherConfigIconAssetCacheKey` | 0x1408a2e40 | [S] | 已存在 |
| `decodeLauncherConfigIconDataURL` | 0x1408973a0 | [S-sig] | `data[5:逗号]` 段上找 `;`（非 `header[:逗号]`） |
| `(*storeBudget).add` | 0x1408965e0 | [S] | 三道阈值 8192/64MiB/32MiPx |
| `(*store).Delete` | 0x1408934c0 | [S-sig] | os.Remove + 清缓存 |
| `ensureLauncherConfigIconJSONEOF` | 0x140897a00 | [S-sig] | Decode + errors.Is(io.EOF) |
| `(*store).ensureLoadedUnlocked` | 0x1408951a0 | [S] | 已有 |
| `(*store).Resolve` | 0x140892e40 | [S] | 已有 |
| `(*store).Put` | 0x1408928e0 | [S-sig] | `[P]` 体预算校验段 |
| `(*store).ExtractLauncherConfigIcons` | 0x140893700 | [S-sig] | 参数透传 |
| `(*store).ExternalizeLauncherConfigIcons` | 0x140893740 | [S-sig] | 参数透传 |
| `(*store).externalizeLauncherConfigIcons` | 0x1408937a0 | [S-sig][P] | 巨型帧(0x14f8)，体未还原 |
| `(*store).Prune` | 0x1408948e0 | [S-sig] | set diff + 删除 |
| `(*store).loadUnlocked` | 0x140894fa0 | [S-sig] | 委派给 ensureLoadedUnlocked |
| `(*store).writeUnlocked` | 0x140895a80 | [S-sig] | JSON 序列化+写入+清缓 |
| `validateLauncherConfigIconLibrary` | 0x140895e80 | [S-sig] | 版本/条目/内容检查 |
| `readLauncherConfigIconStoreBytes` | 0x1408968a0 | [S-sig] | 128MiB 上限 |
| `writeLauncherConfigIconLibraryFile` | 0x140896de0 | [S-sig] | JSON 落盘 |

#### 符号冲突消解
- `iconstore_service.go` 转迁移桩（原有 4 函数全迁）
- `bootstrapservice_config.go` 摘除 `launcherConfigIconStoreForConfigPath` + `(*store).Delete` 两个 `[P]` 骨架
- `launcherconfig.go` 摘除 `collectLauncherConfigIconRefs`（蓝图归属确认为 `launcherconfigiconstore.go`）

#### 实测关键修正（3 处）

1. **`launcherConfigIconStoreForPath` 全局缓存**：现有 `[P]` 骨架的 `&launcherConfigIconStore{path: ...}` 语义不等价，asm 实证走 `internal/sync.HashTrieMap.Load`/`LoadOrStore` → 重写为 `sync.Map` 全局缓存。

2. **`launcherConfigIconStoreError` 模板**：asm 的 `mov ebx, 0xa` 确证 10 字节模板 + 3 个 iface 参数。`.rdata` VA `0x140c440c7` 解码得 `%s: %s: %s`，现实现 `"launcherConfigIconStore: %s: %s"` 的长度（31B vs 10B）与参数数（2 vs 3）均不匹配 → 修正并记入实证错误表。

3. **`StorePath` 4 个 .rdata 常量**：手算 VA 时两处进位错误（前缀 0x140cc4b65→实际 0x140c47565，缺省叶名 0x140cd7638→实际 0x140c37638），用 `resolve_consts.py` 脚本化计算修正。

#### 移交偏差
- `externalizeLauncherConfigIcons` 巨型帧未还原：帧 0x14f8 字节包含 `validateLauncherConfigIconData` 前置 + 9 字段×0x126 qword 复制 + 随机种子初始化 + json.Encoder 写入，完整体需独立批次
- `cursorWrapLoop` / `attachWindowManagementStateIconURLs` 等批次 36 移交偏差未动

#### 指标变化
- **FUNCS=952**（+17，从 935）
- **S=546**（+4，从 542）
- **S-sig=97**（+15，从 82）
- **P=49**（-2，从 51）
- **UNMARKED=259**（持平）
- `go vet` / `build -tags production` / `go test -count=1 -p=1` **全绿**（0.145s）

---

### 批次 38 — launcherconfigiconcommit.go 全量创建（icon commit 管线闭环）（2026-09-13）

#### 完成工作

新建 `backend/launcherconfigiconcommit.go`（274 行，13 函数），从 `source_funcs.txt:2083-2097` 蓝图还原。从 `launcherconfig.go` 迁移摘除 2 个 `[P]` 骨架。补了 2 个缺失的 asm dump（`prepareLauncherConfigIconsForCommit` 0x140890ec0 + `prepareLauncherConfigIconsForCommitWithCurrentRefs` 0x140891120）。

| 函数 | VA | 标记 | 说明 |
|---|---|---|---|
| `(*launcherConfigIconHydrationError).Error` | — | [S] | 类型已在 `types_launcher.go` |
| `(*launcherConfigIconHydrationError).Unwrap` | — | [S] | 类型已在 `types_launcher.go` |
| `(*launcherConfigIconCommitPlan).finish` | 0x140891da0 | [S] | 14L，委派 syncLauncherConfigBeforeIconPrune |
| `collectLauncherConfigIconRefCounts` | 0x1408914c0 | [S] | 从 launcherconfig.go 迁入（同原有 [S] 实现） |
| `collectLauncherConfigIconRefCounts.func1` | — | [S] | 闭包同体 |
| `prepareLauncherConfigIconsForCommit` | 0x140890ec0 | [S] | 121 asm 行，委派 withCurrentRefs |
| `prepareLauncherConfigIconsForCommitWithCurrentRefs` | 0x140891120 | [S] | **220L 主体**：collect → hasInline → budget → externalize → recollect |
| `validateLauncherConfigIconCandidateBudget` | 0x1408916a0 | [S-sig] | 14L 体 + func1(50L)，骨架待 externalizeLauncherConfigIcons 还原后完善 |
| `validateLauncherConfigIconCandidateBudget.func1` | 0x140891880 | [S-sig] | 预算随机采样闭包 |
| `syncLauncherConfigBeforeIconPrune` | 0x140891ea0 | [S] | 14L，orphan set → store.Prune |
| `buildSelfContainedLauncherConfig` | 0x140891fa0 | [S-sig] | 37L 体 + func1(36L)，骨架 |
| `buildSelfContainedLauncherConfig.func1` | 0x140892220 | [S-sig] | slot 内联闭包 |
| `launcherConfigHasInlineIconData` | 0x1408924e0 | [S] | 11L + func1(4L)，visit 遍历 slot.Data |
| `launcherConfigHasInlineIconData.func1` | 0x140892540 | [S] | slot.Data 非空检查 |

#### 符号冲突消解
- `launcherconfig.go` 摘除 `prepareLauncherConfigIconsForCommitWithCurrentRefs` + `collectLauncherConfigIconRefCounts` 两个函数（移至 `launcherconfigiconcommit.go`）
- `collectLauncherConfigIconRefs` 此前已在批次 37 迁至 `launcherconfigiconstore.go`——本批次做了最终确认

#### 指标变化
- **FUNCS=963**（+11，从 952；+13 新文件 −2 摘除）
- **S=546**（+6：Error/Unwrap/finish/collectRefCounts/hasInline/hasInlineFunc1）
- **S-sig=97**（+4：validateCandidateBudget + func1 + buildSelfContained + func1）
- **P=47**（−2：从 launcherconfig.go 摘除 2 个 [P] 骨架）
- **UNMARKED=259**（持平）
- `go vet` / `build -tags production` / `go test -count=1 -p=1` **全绿**（0.125s）

#### 移交偏差
- `launcherConfigIconStorePath` 4 个 .rdata 常量未解码——`externalizeLauncherConfigIcons` 路径空时退化 no-op（兼容测试）
- `cursorWrapLoop` / `attachWindowManagementStateIconURLs` 未还原——不影响本批次
- `validateLauncherConfigIconCandidateBudget` 体为骨架，完整预算校验依赖 `externalizeLauncherConfigIcons` 还原后的实际阈值逻辑

---

### 批次 39 — launcherconfigicon.go 迁移历史函数 + externalizeLauncherConfigIcons 还原（2026-09-13）

#### 完成工作

**目标 A — launcherconfigicon.go 8 个迁移历史函数**

新建 6 个包级函数（从 `source_funcs.txt:2075-2082` 蓝图还原），含现场 asm dump（8 个 VA 区间共 ~2200 字节）逐条翻译：

| 函数 | VA | 标记 | 说明 |
|---|---|---|---|
| `migrateHistoricalLauncherConfigIcons` | 0x14088f680 | [S] | 遍历全部槽位，预算超限时重编码/丢弃（回传 migrated, rescaled, dropped） |
| `migrateHistoricalLauncherConfigIcon` | 0x14088f980 | [S] | 单图标 data:URI → 解码 → 缩放(≤512) → PNG 重编码 → base64 |
| `decodeHistoricalLauncherConfigIcon` | 0x14088ff60 | [S] | 校验 `data:image/{png/jpeg/webp};base64,` 前缀 → base64 解码 |
| `ensureLauncherConfigIconMigrationBackup` | 0x1408904a0 | [S] | 委托 `.before-icon-migration-v1.bak` 备份 |
| `ensureLauncherConfigIconLibraryMigrationBackup` | 0x140890560 | [S] | 委托 `.before-icon-library-v1.bak` 备份 |
| `ensureLauncherConfigMigrationBackupFile` | 0x140890620 | [S] | 实际落盘：Lstat → MkdirAll → OpenFile(O_EXCL) → io.Copy → Sync |
| `ensureLauncherConfigMigrationBackupFile.func1` | (内联 defer) | [S] | 关闭文件 + 未提交时 Remove |

新增依赖 `golang.org/x/image/draw`（CatmullRom 缩放核，原 asm 调用 `draw.Kernel.Scale`）。

**目标 B — externalizeLauncherConfigIcons 完整体还原**

`backend/launcherconfigiconstore_put.go` 原有 100 行 `[P]` 骨架替换为 140 行 `[S-sig]` 实装（asm 0x1408937a0 帧 0x14f8 字节逐条确证）：

流程：`validateLauncherConfigIconData` → `visitLauncherConfigIconSlots` 收集内联 → `normalizeLauncherConfigIconRefSet`（入参 refSet + 现有 Ref 归并）→ 加锁 → `loadUnlocked` → Prune 不在此集合的 Icons → `decodeLauncherConfigIconDataURL` 逐条解码 → 预算校验 → 写入 `cached.Icons` → `writeUnlocked` 持久化。`ExtractLauncherConfigIcons` / `ExternalizeLauncherConfigIcons` 签名为新增 `refSet map[string]struct{}` 参数对齐 asm 实证。

#### 符号冲突消解
- `launcherconfigiconstore_put.go` `externalizeLauncherConfigIcons` 从 `[P]` 骨架升级为 `[S-sig]`
- `launcherconfigiconcommit.go` 调用点 `ExternalizeLauncherConfigIcons(cfg)` → `ExternalizeLauncherConfigIcons(cfg, nil)`

#### 指标变化
- **FUNCS=969**（+6，从 963）
- **S=552**（+6：6 个新迁移函数）
- **S-sig=97**（持平：externalize 从 [P] 骨架→[S-sig] 实装，计数不变）
- **P=46**（−1：externalizeLauncherConfigIcons 升级出 P，无新 P 引入）
- **UNMARKED=259**（持平）
- `go vet` / `build -tags production` / `go test -count=1 -p=1` **全绿**（0.125s，`TestConfigureChangeTriggersReschedule` 已知预存波动）

> **批次 39 指标校正（2026-09-20）**：上面批次 39 的 `S=552/S-sig=97/P=46/UNMARKED=259` 与活体实测不符且分项相加 955 ≠ `FUNCS=969`。实测为 `S=551/S-sig=100/P=49/UNMARKED=268`。原因是历次档位由不同 awk 变体与人工誊抄产生。本段保留原始记录不改写，**以 HANDOFF.md §1 的活体数字为准**。

---

