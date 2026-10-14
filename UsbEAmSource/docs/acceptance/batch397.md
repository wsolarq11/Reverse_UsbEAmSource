# 批次 397 验收 — screenshot UIA 命中优先级 + COM 释放串，并首次接入函数级对拍

**批次**：397
**日期**：2026-12-12
**类型**：函数升级（S-sig → S）+ 新增行为对拍 harness
**域**：screenshot UIA（`backend/screenshot_uia_*.go`）

## 0. 本批的重大修正（必读）

本批在对拍接入过程中，**用原 exe 跳转表的表项数据证伪了本批早期落地中的两处实质错误**。
两处错误均已修复，且修复不是靠重新阅读 asm 得到的，而是靠机器可读的表数据。

| 项 | 早期落地（错） | 表项实证（对） |
|---|---|---|
| 0xc364 的具名含义 | 曾标为 `UIA_DocumentControlTypeId`（50030） | 0xc364 = **50020** = `UIA_TextControlTypeId`；50030 是 0xc35e |
| rank=1 的集合大小 | 只有 50008 一个 | **12 个**：50008/09/10/17/18/21/23/26/28/30/32/33 |

错误一源于十六进制换算失误；错误二源于把 `jmp [rax+rcx*8]` 的间接跳转判为"静态不可确定"后
对 26 个表项作了推断。**表数据一直在 exe 里可读**，不需要推断。

这两处错误在无对拍时不会被发现——它们各自自洽。这正是本批接入对拍的直接动因。

## 1. 目标函数

| 函数 | VA | 大小 | 档位变化 |
|---|---|---|---|
| `screenshotUIAElementRank` | 0x1409b2ce0 内联 | — | [S]（rank 表由表项确证后重写） |
| `screenshotUIAElementArea` | 0x1409b2f69-0x1409b2fc9 | — | [S] |
| `screenshotPreferUIAElementHitInfo` | 0x1409b2ce0 | 928B | [S-sig] → [S] |
| `releaseScreenshotUIAutomation` | 0x1409b31c0 | 128B | [S-sig] → [S] |
| `releaseScreenshotUIAElement` | 0x1409b3240 | 128B | [S-sig] → [S] |
| `releaseScreenshotAccessible` | 0x1409b32c0 | 128B | [S-sig] → [S] |
| `releaseScreenshotUIATreeWalker` | 0x1409b3340 | 128B | [S-sig] → [S] |
| `releaseScreenshotVtblRelease`（新增 helper） | — | — | [S] |

## 2. 跳转表表项实证（本批的核心证据）

抽取器 `tools/parity/emu/extract_func.py` 从原 exe 抽出 4 张跳转表：

| 表 | VA | 表项数 | 命中落点 | 未命中落点 |
|---|---|---|---|---|
| T1 | 0x1411e1f80 | 26 | 0x1409b2eb1（eax=1） | 0x1409b2eb8（eax=2） |
| T2 | 0x1411e2060 | 26 | 0x1409b2f50 | 0x1409b2f57 |
| T3 | 0x1411e2140 | 26 | 0x1409b2ffa | 0x1409b3026 |
| T4 | 0x1411e2220 | 26 | 0x1409b3018（eax=1） | 0x1409b301f（eax=0） |

四表**索引集合完全相同**：`{0,1,2,9,10,13,15,18,20,22,24,25}`。
索引 26/27 表项为 0，被 `cmp rcx,0x19` / `ja` 挡在表外，不可达。

由此得 rank=1 的控件类型集合（基址 0xc358）：
`{50008,50009,50010,50017,50018,50021,50023,50026,50028,50030,50032,50033}`。

**内层分支的最终结论**：T3 与 T4 索引集合相同，故 `rank(aCT)` 与 `rank(bCT)` 在窗口内
恒同值，0x1409b3021 的 `xor eax,1` 恒得 false。该分支不可达，落 `return false`。
（早期版本称此为"推证"，现已升级为**表数据实证**。）

## 3. 门禁（G1 编译 / G2 契约 / G3 行为）

```
go build -tags production -trimpath -buildmode=exe -o artifacts/UsbEAm_Launcher_rebuilt.exe ./backend  → EXIT=0
go vet ./backend                                                                                    → EXIT=0
go test -count=1 -p=1 -tags production ./backend                                                    → EXIT=0
go fmt ./backend                                                                                    → 无差异
go test -count=1 -tags production -run TestParity ./backend                                         → EXIT=0（1557 例通过）
node tools/count_funcs.js                                                                           → UNMARKED=0
```

契约数字（本批实测，非复制）：

```
FUNCS=3302  MARKED=3302  UNMARKED=0
S=1579  S-inline=37  S-eq=2  S-sig=1643  P=41
FAITHFUL=1616  USABLE=1618  TRUE=3259
TRUE/4754 = 68.56%
ΔFUNCS: 3298→3302（+4）  ΔS: 1570→1579（+9）  ΔS-sig: 1648→1643（−5）  ΔP: 41→41（0）
```

## 4. 行为对拍（G3 实证）

本批**首次接入函数级对拍**，落在 `screenshotPreferUIAElementHitInfo`。

工具链（`tools/parity/`）：

- `emu/extract_func.py` — 从原 exe 抽函数字节码 + 只读跳转表，带门禁（0 call、表项无零、rip lea 目标全声明）
- `emu/ref_puiahi.py` — 参考实现；**rank 表由抽取产物的表项机械推导，不手写**
- `emu/gen_cases.py` — 用例生成器
- `cases/screenshot_uia/puiahi.{cases,expected}.txt` — 1557 例黄金用例
- `backend/screenshot_uia_parity_test.go` — Go 侧对拍测试

用例格式为**四角坐标直给**，不用「左上角 + 宽高」：Go 的 `image.Rect` 会交换 Min/Max
使矩形合法，而原函数入参已由调用方构造完毕。直给四角消除了该规整语义在两套实现间的
分歧，并使 `Min > Max` 的退化形态可显式表达。

### 4.1 对拍有效性的自检（变异测试）

对拍通过不等于对拍有效。抽样注入变异确认判别力：

| 变异 | 结果 |
|---|---|
| 从 rank=1 集合删去 50030 | **抓到**（4 例不符） |
| 去掉 0xc364 提前出局分支 | **抓到**（多例不符） |
| `areaB < areaA` 改 `<=` | 未抓到 —— 前有 `areaB != areaA` 守卫，该改写不可达，未抓到属正确 |

第三行记为**已确证的等价改写**，不是漏检。

### 4.2 已知局限（诚实边界）

抽取器只支持**无 call 的叶子函数**。含 call 的函数需更完整的加载器（符号解析、外部
调用桩），当前未实现；此类函数门禁直接失败而非静默降级。

参考实现的分支结构由人工按 asm 转写，与重建实现共用同一份 asm 阅读结论。因此对拍验证
的是"两次转写是否一致"；对"阅读结论本身是否有误"的覆盖依赖**表数据（机械推导）与变异
测试**这两条独立来源。二者均已纳入流程。

**尝试过但未落地**：直接执行抽取的原始字节码（最硬的对拍形态）。`extract_func.py` 已完成
字节码与表的抽取，驱动侧卡在 Go↔裸机器码的 ABI 桥：Go 对函数指针的间接调用会插入自身
的栈增长检查，而目标函数是从原 exe 抽出的裸机器码、无 Go 栈帧元信息。已尝试纯 asm 桩
（`goabiCallPtr`）绕开，崩于桩内参数数组读取，根因未定位。该项记为**未完成的强化路线**，
当前 harness 不依赖它。

## 5. 文件结构调整

`screenshot_uia_windows.go` 本批前为 474 行，超出仓库 400 行阻断线。已将"命中优先级
判定"单元（常量表、rank、area、prefer）迁至新文件：

| 文件 | 行数 | 职责 |
|---|---|---|
| `backend/screenshot_uia_hitpriority_windows.go` | 119 | 命中优先级纯几何/集合判定 |
| `backend/screenshot_uia_windows.go` | 368 | UIA/MSAA/IAccessible 其余函数与 COM 释放串 |

拆分后主体回落至 400 行以下。

## 6. 新增纪律

**`unsafe` 形态**：`go vet` 拒绝 `uintptr(unsafe.Pointer(x))` 与 `unsafe.Add(unsafe.Pointer(uintptrVal), off)`
两种形态（报 `possible misuse of unsafe.Pointer`）。本仓已过 vet 的合规范式见
`screenshot_dxgi_deep_windows.go:100-109` `screenshotCOMVtableSlot`：入口收 `uintptr`，
用 `*(**uintptr)(unsafe.Pointer(&p))` 取 vtable，再用 `uintptr(unsafe.Pointer(vtbl)) + off` 取槽。
COM 包装类还原一律照此形态。

**对拍入库纪律**：用例与期望值入库；抽取产物（`.code.bin`/`.tables.bin`/`.meta.json`）
不入库，由 `extract_func.py` 从目标 exe 重新生成（见根 `.gitignore`）。

## 7. 交付清单

- `backend/screenshot_uia_hitpriority_windows.go`（新，119 行）
- `backend/screenshot_uia_windows.go`（474 → 368 行）
- `backend/screenshot_uia_parity_test.go`（新，对拍测试）
- `tools/parity/README.md`（重写，含工作流与局限）
- `tools/parity/emu/{extract_func.py, ref_puiahi.py, gen_cases.py}`（新）
- `tools/parity/cases/screenshot_uia/puiahi.{cases,expected}.txt`（新，1557 例）
- `.gitignore`（新增 `/UsbEAmSource/tools/parity/bin/`）
