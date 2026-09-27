# UsbEAm Launcher 1.0.3 — 后端服务方法体 "自底向上" 执行图

> 目的：完整逆向还原（函数/方法体）是本工程的唯一大奖励。工具链（redress、all_types
> 类型层、source_funcs 行号蓝图、前端 binding 契约）已就位；缺失的是"逐服务还原的执行顺序、
> 契约来源、提取技巧、验证节点"的可执行手册。本文即为此而生，供具备 shell / Ghidra 的会话直接开干。
>
> 研究用途，版权（c）2026 DOGFIGHT360 合规。目标口径：可编译 + 功能一致（非字节级同哈希）。
> 边界：还原仅限个人学习研究，不发布、不商用、不破解授权。

---

## 0. 顶层验收（HANDOFF 7 的收口）

- `backend/main.go` 的 `main()` 最终调用 `app.Run()`，注册全部服务。
- `go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` 成功。
- 方法签名与前端 binding 对齐；依赖数从 14 走向 25。
- 不强求字节同哈希。

工具链回调（必须，勿用默认 go 1.27）：

```bash
export HTTPS_PROXY=http://127.0.0.1:7890 HTTP_PROXY=http://127.0.0.1:7890 ALL_PROXY=http://127.0.0.1:7890
GOBIN="$(go env GOPATH)/bin"
"$GOBIN/go1.25.12.exe" version   # go1.25.12 windows/amd64
```

反汇编还原签名源（`redress types -s -v` 含 std/vendor）；函数体还原走 Ghidra 配 Go 插件
（DWARF/pclntab），按 source_funcs.txt 的行号区间定位。

## 2. 还原顺序基准（自底向上）

优先级两条铁律：

1. **先清依赖闭包，再碰聚合层。** 每个服务被还原前，其字段依赖的全部类型/方法须先在 0 期接通，
   否则还原引到不存在符号会导致整个 backend 编译破碎。
2. **每还完一个叶子服务立即 `go build`**，把不确定性扼杀在单服务粒度，而不是等一个大炸组件崩盘。

层级划分：

```
L0 纯数据层（已完成）                 backend/ 23 个 types_<domain>.go 全部可编译
L1 无外部依赖的叶子服务                launcherAssetService、launcherConfigStore(部分)、pluginWindowService、
                                       memoryReleaseService、mouseGestureService 的内部纯逻辑
L2 依赖 L1 + w32/外部库 的服务         fileLocatorService、FileIndexService、desktopWidgetService、
                                       oledBlackoutService、windowManagementService、inputMonitorService、
                                       twoFactorService、screenshot*、launcherAssetService 的读写落地
L3 聚合编排层（BootstrapService）      initializeWithCheckpoint / attachApp / attachFileIndexService / attachWindow
                                       —— 依赖上面全部子服务
L4 入口（已编译可跑）                   main.go 的 main() + NewBootstrapService
```

## 3. 依赖闭包图谱（BootstrapService 为根）

依据：`backend/types_launcher.go` L365-452（已核实）+ `all_types.txt` L22485（一致）。BootstrapService
字段里持有的子服务指针，就是还原 L3 时必需的下层契约：

| 字段 | 类型 | 蓝图层级 | 核心依赖线索 |
|---|---|---|---|
| configStore | *launcherConfigStore | L1 | path/writeConfig/runtimeReady/iconFingerprint |
| twoFactor | *twoFactorService | L2 | 两步验证码、fetchInternetTime |
| memoryRelease | *memoryReleaseService | L1 | 内存回收释放 |
| oledBlackout | *oledBlackoutService | L2 | 防烧屏、候选窗口、desktop cursor 控制 |
| windowManagement | *windowManagementService | L2 | 窗口全屏/定位、opacity 快照 |
| mouseGestures | *mouseGestureService | L2 | 手势画布、动作执行（launcherAction/windowMove） |
| assets | *launcherAssetService | L1/L2 | Register*/lookup/HTTP Serve（见 4.1） |
| screenshotPin / screenshotPreview / screenshotSelectionToolbar | *screenshot*WindowService | L2 | 截图窗链、标注、选区工具栏 |
| pluginWindows | *pluginWindowService | L1/L2 | Open/OpenOwned/Close/PluginURL |
| inputMonitor | *inputMonitorService | L2 | 全局输入钩子 |
| fileIndex | fileSearchRuntimeWarmer | L2 | 文件索引驻留接口（types_filesearch.go L553） |
| fileLocator | *fileLocatorService | L2 | ValidateFilter、文件定位搜索 |
| desktopWidgets | *desktopWidgetService | L2 | 天气/日历/时钟/通知调度 |

> 上表字段名均为 `backend/types_launcher.go` L365-452 的实际 struct 字段，非臆造。

## 4. 逐服务执行卡

每条"可用契约"都已在本 repo 里 `read` 验证过文件与行号，不是假设。

### 4.1 launcherAssetService（L1，无外部服务依赖 —— 首选起点）

- 蓝图：`docs/goresym/source_funcs.txt` L1912-1941 起（File: launcherasset.go）
  - 方法集合（从蓝图行直接获得）：RegisterBytes(87-674)、RegisterStableBytes(117-674)、
    RegisterFile(149-674)、register(200-684)、registerStable(240-722)、ServeAssetRequest(288-344)、
    ReadBytes(347-383)、Exists(383-406)、lookup(406-538)、Clear(427-438)、currentTime(438-445)、
    pruneExpiredLocked(445-538)、validateItemSize(453-481)、ensureCapacityLocked(481-511)、
    addEntryLocked(511-543)、evictOldestLocked(543-567)、openFileBounded(567-593)、
    readFileBounded(593-606)、parseLauncherAssetRequest(609-722)、buildLauncherAssetURL(644-653)、
    normalizeLauncherAssetNamespace(653-687)、stableLauncherAssetID(687-698)、
    pathEscapeLauncherAssetSegment(698-709)、newLauncherAssetID(709-714)
- 结构体字段（已核 `all_types` L23183+backend types_launcher L474）：lock/entries/namespaceBytes/
  totalBytes/next/now/limits
- 可实现签名（红线：只做签名+来源标注，函数体留 TODO 指向蓝图行号；参数类型须从
  `redress sign -s -v` 或前端调用点再确认，勿臆造）。

  还原起点建议：先做不动文件 IO 的纯内存方法 `register/registerStable/lookup/Clear/currentTime/
  pruneExpiredLocked`（字段完全覆盖），再落地 `RegisterBytes/RegisterFile/ReadBytes`（涉及文件 IO，
  需结合方式）；ServeAssetRequest 是 HTTP handler，最后。
- 验证节点：`RegisterBytes` 幂等（同名 entry 覆盖）+ `Exists/ReadBytes` 回读一致 + `totalBytes` 预算校验。
- 依赖：无外部服务（本域纯内部），是"第一个叶子"最稳妥选择。

### 4.2 launcherConfigStore / launcherConfig（L1，main.go 已关联）

- 蓝图：`docs/goresym/source_funcs.txt` 的 `File: launcherconfig.go` 段（约占 L1944-1995）——
  加载/默认/归一化等函数；`launcherConfigStoreForPath` 在 L2140（行 44-55）。
- `backend/main.go` 当前 `launcherConfigStoreForPath` 只 `return &launcherConfigStore{}`，未写入 path；
  真体应赋 `path: path`（`launcherConfigStore.path` 字段在 `all_types.txt` L23220 已核实存在）。
- 还原顺序建议：`loadOrCreateLauncherConfig`(579-1047) → `defaultLauncherConfigWithOptions`
  (474-3083 很大) → `normalizeLauncherConfigForSave`。此服务逻辑密集且行号长，放第二批。

### 4.3 FileIndexService（L2，文件索引；字段在 backend/types_filesearch.go L357）

- 蓝图：`docs/goresym/source_funcs.txt` 的 `filesearch_*` 段：
  `filesearch_index_windows.go`（约 L1131 起）、`filesearch_windows.go`（约 L1445 起）等。
- 依赖：VolumeIndex、fileSearchCandidateHeap 等类型字段在 `backend/types_filesearch.go` 已全；
  先恢复 VolumeIndex 的 checkpoint / overlay merge 读写，涉及 `modernc.org/sqlite`（依赖数从 14 提升的
  主要来源之一）。

### 4.4 其余 L2 服务（windowManagement / desktopWidget / oled / inputMonitor / twoFactor / screenshot 链）

每个文件蓝图区段（源码中 `File:` 行号）已在 §2 汇总。用同一个卡模式：字段 → 纯函数先行 → IO/系统后行 → 每层 build。

## 5. 契约来源矩阵（按置信度排序，顶层可信）

| 来源 | 形态 | 置信度 | 用途 |
|---|---|---|---|
| `goresym/all_types.txt` | 完整 struct 字段 | 极高（已核实类型层一致） | 方法骨架字段引用 |
| `source_funcs.txt` | 文件名+方法名+行号 | 高（已核实） | 定位反汇编区间 |
| `frontend/dist/assets/*.js` | Wails binding 调用方 | 高（已核实存在 13 Manager） | 对外方法名/参数契约 |
| `redress types -s -v`（运行时） | 含 std/vendor 签名 | 高（需 shell 执行） | 精确签名 |
| Ghidra 反汇编 | asm→C→Go | 验体 | 函数体填充 |

注：签名不能臆造。当前 `source_funcs.txt` 只给"函数名 + 行号"，不含参数；签名须以
`redress types -s -v` 或前端调用点为准。

## 6. 每层编译门禁（不整碎工程）

完成每一小批后（比如 launcherasset 的纯方法集），立即：

```bash
GOBIN="$(go env GOPATH)/bin"
"$GOBIN/go1.25.12.exe" build -tags production -trimpath -buildmode=exe ./backend
"$GOBIN/go1.25.12.exe" vet ./backend
```

- 重点：优先保证整体可编译，再逐层扩展（本仓库非 git，扩展前为手头的文件做一份 diff 备份以便回滚）。

## 6.1 多方质检验收（每批次强制出口 Gate）

按用户要求"按批次完工后多方质检验收"，每批次（A/B/C/D）完工**且不进入下一批**之前，必须通过以下
四路独立质检，任一 FAIL 则修复后重跑，不允许带病推进：

| 视角 | 手段 | 通过判据 |
|---|---|---|
| G1 编译/静态 | `go1.25.12 build` + `go vet ./backend` | 全绿；记录依赖数，与目标 25 逐批比对 |
| G2 语义契约 | 本批每个方法签名对齐 `frontend/dist` binding / `all_types.txt` 字段 / `source_funcs.txt` 行号 | 签名无一臆造；字段/类型交互与字段层一致 |
| G3 行为自测 | 单测/表驱动覆盖：极值边界、非法状态迁移、并发冲突（乐观锁/锁）、幂等、超时/时钟 | backend 包内 `go test` 跑绿，覆盖点逐一命中 |
| G4 独立复核 | 换一个干净复核视角（不复用本会话的 edit 缓存与先前结论），重读关键方法的字段交互与交付语义 | 复核结论可回读、记录于 `docs/acceptance/<batch>.md`，发现偏差即回修 |

验收记录规范：每批次在 `docs/acceptance/<batch>.md` 写四路结果（全链路可回溯），
未写该记录 = 该批次未完成。

## 7. 建议批次切分（每批之间都过 §6+§6.1 双门禁）

- 批次 A（叶子）：launcherAssetService（4.1）——依赖最小、字段齐全、纯函数度高。
- 批次 B（配置服务）：launcherConfigStore / launcherConfigStoreForPath —— 承接 main.go。
- 批次 C（L2）：pluginWindowService、twoFactorService、memoryReleaseService 等按"外部依赖最少"续推。
- 批次 D（编排）：initializeWithCheckpoint / attachApp / attachWindow 真体（依赖占位全部就位后）。

---

## 附：红线（来自 AGENTS.md，不可协商）

- 禁止命令行改文件（sed -i/awk/perl -i 等）；移动/删除前确认未提交改动。
- 函数行数通用 4~20 行（提示值）；核心规则可放宽，须表驱动 + 全分支单测。
- 命名具体独特（全局 grep 结果 <5 条），不用 data/handler/Manager 模糊词。
- 幂等 / 重试 / 补偿 / 审计（WAL）：还原各服务时按每服务语义落实，核心状态变更先写审计日志。