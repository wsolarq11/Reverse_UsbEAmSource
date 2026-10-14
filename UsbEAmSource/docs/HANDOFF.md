# UsbEAm Launcher 1.0.3 — 逆向还原工程 · 技术交接文档

> 目的：把已完成的取证/还原进展、方法、环境与工具固化，让**新会话**可以无缝续接"后端服务方法体逐函数重建"这一独立大工程。
> 目标口径（用户已确认，2026-09-20 升级为**总进度 100%**）：**4,754 蓝图函数 100% 还原、110 个未落地原始文件 100% 落地、`UNMARKED=0` 且 `P=0`**；全程保持**可编译 + 功能一致**（非字节级同哈希）。
> 研究用途，版权（©2026 DOGFIGHT360）合规。

---

## 致新会话

你接手的是 UsbEAm Launcher 1.0.3 的逆向还原工程，一段已经走通 68.56% 的活。别被"逆向"两个字唬住：前端已字节级还原，后端 507 个结构体全量还原且能编译，3302 个函数已落地、其中 3259 个是真函数。你拿到的是一副已经搭好骨架、只剩填肉的躯壳，不是荒地。

三条铁律，开工前刻进脑子：

1. **先跑 count，再谈数字**。`bash tools/count_funcs.sh` 是唯一真相源，任何批次数字落笔前必须重跑。这工程历史上三次数字失真，根因全是"抄上一批改一改"。别当第四个。
2. **asm 实录压过一切推断**。签名、字段偏移、锁内锁外，只有反汇编指令说了算。签名没定型就老老实实留 `[S-sig]`，绝不造伪 `[S]`——伪忠实比存根更毒，它会骗过下一棒。
3. **两样东西别碰**：combridge IUnknown（128B 那些 AddRef/Release/QueryInterface，是自动提升不是活）和 `linearFloatToSRGBByte`（名字错配，落一次白落一次）。

入口只有一个：本文件 §0。环境坑在 `ENV.md`，纪律和实证错误教训在 `DISCIPLINE.md`，下一步在 §2，缺函数清单 `node tools/list_missing.js` 现查现用。其余都别翻，翻旧批次记录只会浪费你的上下文。

慢就是快。每批只做一件事：落体 → 门禁 → count → 验收 → 提交，五步闭环再开下一批。别一口气吞一个域，那会死在半路。

最后一段（68.56% → 100%）交给你了。把它做完。

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

**活体 count（唯一真相源，`bash`/`awk` 本机可用，位于 Git for Windows）**：

```powershell
bash tools/count_funcs.sh
```

**缺失清单（权威，自动含方法提升，勿手动落 combridge IUnknown）**：

```powershell
node tools/list_missing.js
```

**P 存根清单**：

```powershell
bash tools/list_p.sh
```

**单批收尾五步（SSOT 单出口，禁止抄上一批数字）**：落地函数 → 三门禁 EXIT=0 → `bash tools/count_funcs.sh` 取活体数字 → 写 `docs/acceptance/batchNNN.md` + 更新本文件 §1 标题/§2/进度表 + `node tools/gen_batch_index.js` 重生成索引 → `git commit` + `git push` → `gh run list` / `gh api` 轮询 CI 到 `completed success`。

**环境事实（2026-12-12 实测核准，详见 `DISCIPLINE.md` 与 `ENV.md`）**：`git`（`C:\Program Files\Git\cmd`）、`bash` / `awk`（`C:\Program Files\Git\usr\bin`）、`gh`（`C:\Program Files\GitHub CLI`，已认证 `wsolarq11`，含 `workflow` scope）**均在 PATH**；**仅 `scoop` 不在 PATH**，且 `scoop/apps/temurin-lts-jdk/current` 为断裂链接（JDK 已被卸载，故 Ghidra 不可用）。`dulwich` **未安装**，旧的 `tmp_git_sync.py` 已不存在——commit/push 一律用原生 `git`。`linearFloatToSRGBByte`(0x14097dd60) 已落地但 list_missing 仍列，**勿重落**；`nativeFileDragDataObject.*`/`nativeFileDragFormatEnumerator.*`（128B）为 combridge IUnknown 自动提升，**勿手动落**。

---

## 1. 一句话现状（批次 397 · screenshot UIA 命中优先级 + COM 释放串 [S-sig→S] · 首次接入函数级对拍 · FUNCS 3302，68.56%）

前端层已**字节级完整还原**并验证；Go 后端**类型层（507 个结构体）已全量还原且能编译**；函数/方法体还原推进到**批次 397**，后端非测试代码 **236 个 .go 文件**，测试 **94 文件**。

**门禁活体实测（批次 397）**：`go1.25.12 build ./backend` / `go vet ./backend` / `go test ./backend` / `go fmt ./backend` 四项 **EXIT=0**（`ok changeme/backend`），对拍 `go test -run TestParity ./backend` **1557 例通过**。

**⚠️ 对拍 harness 已落地**：`tools/parity/` 此前只有 README（`cases/emu/run.sh` 均不存在），本批首次接入。工作流与纪律见 `docs/STRATEGY.md` §6 与 `tools/parity/README.md`。首个样本 `screenshotPreferUIAElementHitInfo`。**历史批次（<397）的 `[S]` 与 `[S-eq]` 落体一律未经对拍**，按 STRATEGY §5 规则 3 严格说应为 `[S-sig]`；此项列为待清偿技术债，优先在触及该域时补齐。

**当前进度分母（批次 387 实测，此后一律以此为准）**：

| 指标 | 实测值 | 目标（总进度 100%） | 取证方式 |
|---|---|---|---|
| 蓝图函数项 | 4,754 | 4,754 | `docs/goresym/source_funcs.txt` 中 `Lines: a to b (n)` 条目计数 |
| 蓝图源文件数 | 145 | 145 | 同文件 `^File: ` 条目计数 |
| 已重建函数 | 3302 | 4,754 | `bash tools/count_funcs.sh` 实测（批次 397 后） |
| 真函数（S+S-inline+S-sig） | 3259 | 4,754 | 同上，**批次 397 达 68.56%** |
| 文件覆盖 | 114/144 | **100%（144/144）** | backend 非测试文件名与蓝图 `File:` 清单逐个对名 |
| 未落地原始文件 | 30 | **0** | 差集（清单见 `UNLANDED.md`，批次 375 后实测） |
| UNMARKED | 0 | **0** | `bash tools/count_funcs.sh` 实测 |
| [P] 存根 | 41 | **0** | 同上（批次 397 后） |

**关键里程碑（已闭环域，详见各批验收）**：memoryrelease、SQLite、mouseGesture、OLEDBlackout 执行域；Screenshot 全链（PNG 编码器 / 内存预算 / GDI 捕获 / 光标 / DXGI 枚举与捕获 / HDR / 调试 / UIA 命中优先级与 COM 释放）；hotkey 三角与抑制域；launcherasset（icon 中枢 / 驱逐 / 文件注册）；launcherconfigicon（预算 / 存储 / commit 管线 / 迁移）；windowManagement；appicon GDI 渲染；plugin（id / security / host / discovery）；twofactor 全链；qrcode（框选 / 剪贴板 / 解码）；filesearch（名称搜索计划 / 排序 / 拼音 / 节点访问）；开机自启（Task Scheduler）；配置存储（CompareAndSwap/Replace 链）；filelocator 运行时；gpu 纯逻辑；desktopwidgets（日历 / 时钟 / 组件存储 CRUD）。

**⚠️ 口径纪律**：任何批次记录落笔前必须先跑 `bash tools/count_funcs.sh` 取活体数字，禁止抄上一批数字改一改。标记书写规范与档位定义见 `DISCIPLINE.md`。

---

## 2. 交接状态（批次 397 收尾后）

**下一批推荐**：先把对拍 harness 扩到同域已有 `[S]` 落体——`screenshotUIAElementHitContainsPoint`（0x1409b2c60，128B，同为无 call 叶子函数）是现成目标，`extract_func.py` 可直接抽。这既清偿 §1 的技术债，又验证工具链的可复用性（换函数只需写 `ref_<fn>.py`）。之后再推进 screenshot UIA 的 `[P]` 签名专项（6 个存根，序言已标注"参数语义未定"）。

**⚠️ `unsafe` 形态纪律（批次 397）**：`go vet` 拒绝 `uintptr(unsafe.Pointer(x))` 与 `unsafe.Add(unsafe.Pointer(uintptrVal), off)`。合规范式见 `screenshot_dxgi_deep_windows.go:100-109` `screenshotCOMVtableSlot`。

**⚠️ 抽取器能力边界**：`tools/parity/emu/extract_func.py` 只支持**无 call 的叶子函数**，含 call 直接门禁失败。直接执行原始字节码的强化路线未打通（见批次 397 验收 §4.2）。

**未落地文件差集 30**（清单见 `UNLANDED.md`，批次 375 后实测）；**P=41**；**UNMARKED=0 保持**。

**2026-12-12 整洁基线**：追踪文件 811 → 792。四项根因修复（每项详见 commit `dcfec49` 与 `STRUCTURE.md` §9）：
① `.gitignore` 根锚定漏洞导致 `UsbEAmSource/work/` 9 个产物漏入库；② `.gitattributes` + `core.autocrlf=true` 会在 checkout 时把 `.go` 改写成 CRLF；
③ 标记识别逻辑 4 份各自实现（2 份漏识别 `[S-eq]`），已统一到 `count_funcs.sh`；④ `HANDOFF.md` 声称 git/bash/awk 不在 PATH、commit 走 dulwich——实测均在 PATH，dulwich 未装。
**工具收敛**：`tools/` 32 → 21 个文件，删除重复/临时/一次性工具 16 个。旧文档若引用已删工具，视为历史记录，勿照抄命令。

**整洁基线（批次 374 后一次性）**：全仓 `gofmt` 已归位（`git show a7e52e6`，纯格式，count 不变），`go vet` 干净；后端纯 Go（`CGO_ENABLED=0`、无 `CgoFiles`），`.github/workflows/ci.yml` 已对齐本地门禁（go1.25.12 + `-tags production` build/vet/test）。新会话开工前若 `gofmt -l ./backend` 仍列文件，说明有新增手写码未格式化。

**归档状态**：批次 19–39 详细记录 → `acceptance/batch19-39.md`；批次 40–374 验收 → `acceptance/batchNN.md`（索引见 `BATCH_INDEX.md`）。

**未经对拍的历史落体（待清偿技术债）**：`STRATEGY.md` §5 规则 3 要求"L1 落体必须带对拍测试"，
但对拍 harness 直到批次 397 才首次落地。因此**批次 <397 的全部 `[S]` 与 `[S-eq]` 落体未经对拍**，
严格说只达 `[S-sig]`。处置口径：**不回溯改标记**（那会一次性调低全部历史进度且无实测依据），
改为"触及即补"——每次动到某域的既有 `[S]` 时，顺手补该域的对拍。清明口径由你定。

---

## 3. 文档地图（各就各位，单一真相源）

| 文件 | 职责 | 谁更新 |
|---|---|---|
| `HANDOFF.md`（本文件） | 精简导航 + 活体口径 + 交接状态 | 每批收尾时 |
| `ENV.md` | 目标物身份（含 SHA256）、构建参数、环境事实、反汇编资产清单、复现前置 | 环境/资产变化时 |
| `DISCIPLINE.md` | 逐寄存器还原纪律 + 累计实证错误纠正（每批开工前必读） | 新错误实证时 |
| `UNLANDED.md` | 未落地原始文件差集 + 重跑命令 | 差集变化时 |
| `BATCH_INDEX.md` | 批次 → 验收文档索引（脚本生成，勿手改） | `node tools/gen_batch_index.js` |
| `acceptance/batchNN.md` | 每批四路验收（G1 编译 / G2 契约 / G3 行为 / G4 独立复核） | 每批收尾时 |
| `acceptance/batch19-39.md` | 批次 19–39 详细记录归档（无独立验收文档） | 冻结 |
| `STRATEGY.md` | 还原策略（L0→L1→L2 分层口径）+ 标记口径 SSOT + 对拍纪律 | 策略变更时 |
| `STRUCTURE.md` | 目录职责、路径约定、**版本控制边界与整洁履历** | 结构变更时 |
| `tools/parity/README.md` | 对拍工作流、用例格式、已知局限 | 对拍扩展时 |
| `goresym/source_funcs.txt` | 4,754 蓝图函数权威清单 | 冻结 |
| `.gitattributes` | 行尾策略（`.go` 钉死 LF，Windows 脚本 CRLF，二进制标记） | 行尾策略变更时 |

**单批收尾落笔顺序**：函数落地 → 门禁 → count → `acceptance/batchNN.md` → 更新本文件 §1/§2 → `gen_batch_index.js` 重生成索引 → commit/push → CI 绿。
