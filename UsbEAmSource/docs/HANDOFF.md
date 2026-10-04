# UsbEAm Launcher 1.0.3 — 逆向还原工程 · 技术交接文档

> 目的：把已完成的取证/还原进展、方法、环境与工具固化，让**新会话**可以无缝续接"后端服务方法体逐函数重建"这一独立大工程。
> 目标口径（用户已确认，2026-09-20 升级为**总进度 100%**）：**4,754 蓝图函数 100% 还原、110 个未落地原始文件 100% 落地、`UNMARKED=0` 且 `P=0`**；全程保持**可编译 + 功能一致**（非字节级同哈希）。
> 研究用途，版权（©2026 DOGFIGHT360）合规。

---

## 致新会话

你接手的是 UsbEAm Launcher 1.0.3 的逆向还原工程，一段已经走通 68.24% 的活。别被"逆向"两个字唬住：前端已字节级还原，后端 507 个结构体全量还原且能编译，3244 个函数已落地、其中 3202 个是真函数。你拿到的是一副已经搭好骨架、只剩填肉的躯壳，不是荒地。

三条铁律，开工前刻进脑子：

1. **先跑 count，再谈数字**。`node tools/count_funcs.js` 是唯一真相源，任何批次数字落笔前必须重跑。这工程历史上三次数字失真，根因全是"抄上一批改一改"。别当第四个。
2. **asm 实录压过一切推断**。签名、字段偏移、锁内锁外，只有反汇编指令说了算。签名没定型就老老实实留 `[S-sig]`，绝不造伪 `[S]`——伪忠实比存根更毒，它会骗过下一棒。
3. **两样东西别碰**：combridge IUnknown（128B 那些 AddRef/Release/QueryInterface，是自动提升不是活）和 `linearFloatToSRGBByte`（名字错配，落一次白落一次）。

入口只有一个：本文件 §0。环境坑在 `ENV.md`，纪律和 77 处实证错误教训在 `DISCIPLINE.md`，下一步在 §2，缺函数清单 `node tools/list_missing.js` 现查现用。其余都别翻，翻旧批次记录只会浪费你的上下文。

慢就是快。每批只做一件事：落体 → 门禁 → count → 验收 → 提交，五步闭环再开下一批。别一口气吞一个域，那会死在半路。

最后一段（68.24% → 100%）交给你了。把它做完。

---

本文件是**精简导航版**。详细内容已拆到各就各位：环境/资产 → `ENV.md`，纪律/实证错误 → `DISCIPLINE.md`，未落地清单 → `UNLANDED.md`，批次索引 → `BATCH_INDEX.md`，每批验收 → `acceptance/batchNN.md`。

---

## 0. 新会话启动卡（从这里直接开工）

**工作目录（Windows PowerShell，非 bash）**：`D:\AI\projects\Reverse_penetration\Reverse_UsbEAmSource\UsbEAmSource`

**三条门禁（精确工具链，`go` 用绝对路径）**：

```powershell
$go = "C:\Users\Administrator\go\bin\go.exe"
& $go build -tags production -trimpath -buildmode=exe -o artifacts/UsbEAm_Launcher_rebuilt.exe ./backend
& $go vet ./backend
& $go test -count=1 -p=1 -tags production ./backend
```

**活体 count（Node 等价 awk，本环境 bash/awk 不可用）**：

```powershell
node tools/count_funcs.js
```

**缺失清单（权威，自动含方法提升，勿手动落 combridge IUnknown）**：

```powershell
node tools/list_missing.js
```

**当前口径（批次 374）**：`FUNCS=3248 MARKED=3248 S=1508 S-inline=37 S-eq=1 S-sig=1661 P=41 UNMARKED=0`，真函数 3206/4754 = **68.32%**。

**单批收尾五步（SSOT 单出口，禁止抄上一批数字）**：落地函数 → 三门禁 EXIT=0 → `node tools/count_funcs.js` 取活体数字 → 写 `docs/acceptance/batchNNN.md` + 更新本文件 §1 标题/§2/进度表 + `node tools/gen_batch_index.js` 重生成索引 → dulwich commit+push（`$env:GITHUB_TOKEN = gh auth token`）→ `gh api` 轮询 CI 到 `completed success`。

**环境陷阱（换会话必读，详见 `DISCIPLINE.md` 与 `ENV.md`）**：`git.exe`/`bash.exe`/`awk`/`scoop` 不在 PATH（commit/push 用 dulwich `C:\Users\Administrator\.dsh\tmp_git_sync.py` 或纯 push 内联；count 用 `tools/count_funcs.js`）；dulwich push 网络抖动时用**纯 push 重试**（勿重复 commit，会产出重复提交）；`linearFloatToSRGBByte`(0x14097dd60) 已落地但 list_missing 仍列，**勿重落**；`nativeFileDragDataObject.*`/`nativeFileDragFormatEnumerator.*`（128B）为 combridge IUnknown 自动提升，**勿手动落**。

---

## 1. 一句话现状（批次 374 · GSMTC 来源/标题读取 +2 · FUNCS 3248，68.32%）

前端层已**字节级完整还原**并验证；Go 后端**类型层（507 个结构体）已全量还原且能编译**；函数/方法体还原推进到**批次 374**，后端非测试代码 **125 个 .go 文件**，测试 **54 文件**。

**门禁活体实测（批次 374）**：`go1.25.12 build ./backend` / `go vet ./backend` / `go test ./backend` 三项 **EXIT=0**（`ok changeme/backend`）。

**当前进度分母（批次 374 实测，此后一律以此为准）**：

| 指标 | 实测值 | 目标（总进度 100%） | 取证方式 |
|---|---|---|---|
| 蓝图函数项 | 4,754 | 4,754 | `docs/goresym/source_funcs.txt` 中 `Lines: a to b (n)` 条目计数 |
| 蓝图源文件数 | 145 | 145 | 同文件 `^File: ` 条目计数 |
| 已重建函数 | 3248 | 4,754 | `node tools/count_funcs.js` 实测（批次 374 后） |
| 真函数（S+S-inline+S-sig） | 3206 | 4,754 | 同上，**批次 374 达 68.32%** |
| 文件覆盖 | 113/144 | **100%（144/144）** | backend 非测试文件名与蓝图 `File:` 清单逐个对名 |
| 未落地原始文件 | 31 | **0** | 差集（清单见 `UNLANDED.md`，批次 374 后实测） |
| UNMARKED | 0 | **0** | `node tools/count_funcs.js` 实测 |
| [P] 存根 | 41 | **0** | 同上（批次 278 持平） |

**关键里程碑（已闭环域，详见各批验收）**：memoryrelease、SQLite、mouseGesture、OLEDBlackout 执行域；Screenshot 全链（PNG 编码器 / 内存预算 / GDI 捕获 / 光标 / DXGI 枚举与捕获 / HDR / 调试）；hotkey 三角与抑制域；launcherasset（icon 中枢 / 驱逐 / 文件注册）；launcherconfigicon（预算 / 存储 / commit 管线 / 迁移）；windowManagement；appicon GDI 渲染；plugin（id / security / host / discovery）；twofactor 全链；qrcode（框选 / 剪贴板 / 解码）；filesearch（名称搜索计划 / 排序 / 拼音 / 节点访问）；开机自启（Task Scheduler）；配置存储（CompareAndSwap/Replace 链）；filelocator 运行时；gpu 纯逻辑；desktopwidgets（日历 / 时钟 / 组件存储 CRUD）。

**⚠️ 口径纪律**：任何批次记录落笔前必须先跑 `node tools/count_funcs.js` 取活体数字，禁止抄上一批数字改一改。标记书写规范与档位定义见 `DISCIPLINE.md`。

---

## 2. 交接状态（批次 374 收尾后）

**下一批推荐**：`node tools/list_missing.js` 头部 448B/480B 候选；filesearch 域 `VolumeIndex.activeEntryCountLocked`(0x1407e8340)/`VolumeIndex.overlayStatsLocked`(0x1407e7fe0) 优先；`P=41 → [S]` 转换（`tools/list_p.awk` 定位：mousegestures/nativedrag/oledblackout_windows/screenshot_uia_windows/screenshot_windows/webview2_process_windows）。

**未落地文件差集 31**（清单见 `UNLANDED.md`，批次 374 后实测）；**P=41 持平**；**UNMARKED=0 保持**。

**归档状态**：批次 19–39 详细记录 → `acceptance/batch19-39.md`；批次 40–374 验收 → `acceptance/batchNN.md`（索引见 `BATCH_INDEX.md`）。

---

## 3. 文档地图（各就各位，单一真相源）

| 文件 | 职责 | 谁更新 |
|---|---|---|
| `HANDOFF.md`（本文件） | 精简导航 + 活体口径 + 交接状态 | 每批收尾时 |
| `ENV.md` | 目标物身份、锁定构建参数、关键环境事实、反汇编资产清单 | 环境/资产变化时 |
| `DISCIPLINE.md` | 逐寄存器还原纪律 + 累计实证错误纠正（每批开工前必读） | 新错误实证时 |
| `UNLANDED.md` | 未落地原始文件差集 + 重跑命令 | 差集变化时 |
| `BATCH_INDEX.md` | 批次 → 验收文档索引（脚本生成，勿手改） | `node tools/gen_batch_index.js` |
| `acceptance/batchNN.md` | 每批四路验收（G1 编译 / G2 契约 / G3 行为 / G4 独立复核） | 每批收尾时 |
| `acceptance/batch19-39.md` | 批次 19–39 详细记录归档（无独立验收文档） | 冻结 |
| `STRATEGY.md` | 还原策略（L0→L1→L2 分层口径） | 策略变更时 |
| `STRUCTURE.md` | 目录职责与路径约定 | 结构变更时 |
| `goresym/source_funcs.txt` | 4,754 蓝图函数权威清单 | 冻结 |

**单批收尾落笔顺序**：函数落地 → 门禁 → count → `acceptance/batchNN.md` → 更新本文件 §1/§2 → `gen_batch_index.js` 重生成索引 → commit/push → CI 绿。
