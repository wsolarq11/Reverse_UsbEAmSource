# 对拍 harness（行为等价验证）

> 验证 L1 `[S-eq]` 行为等价的唯一手段。策略见 `docs/STRATEGY.md`。
> 原则：先函数级，不可行降程序级；对拍不过的 L1 落体一律回炉，不算完成。

## 两级对拍

| 级 | 适用 | 手段 |
|---|---|---|
| 函数级 | 纯逻辑函数（配置解析 / 路径规范化 / 编码解码 / 哈希 / 匹配） | 抽取原函数字节码与其只读表，跑原始机器码取输出，与重建代码对拍 |
| 程序级 | GUI 行为 / 有 goroutine / 系统调用 / 逃逸的函数 | 原二进制 vs 重建二进制，比输出文件 / 配置 / 二维码 / 截图等黄金用例 |

## 目录骨架

```
tools/parity/
├── README.md          # 本文
├── emu/               # 函数级对拍工具链
│   ├── extract_func.py    # 从原 exe 抽取函数字节码 + 只读跳转表（带门禁）
│   ├── ref_<fn>.py        # 该函数的参考实现（期望值生成器）
│   └── gen_cases.py       # 对拍用例生成器
├── bin/               # 抽取产物（.code.bin / .tables.bin / .meta.json）
└── cases/<domain>/    # 对拍用例（输入 + 期望输出，黄金用例）
```

## 规则（写死）

1. L1 落体必须带对拍用例，无用例不算 L1，只能算 `[S-sig]`。
2. 黄金用例的来源：优先原二进制真实行为，其次 asm 显式路径，禁止凭空造期望值。
3. L2 替换时，该函数 L1 的用例必须仍通过（编译 + 对拍双保险）。
4. 期望值必须由独立参考实现生成，禁止从被测实现反推。
5. 参考实现中的**表数据**（跳转表 / 常量表）必须由抽取产物机械推导，禁止手写。

## 函数级对拍工作流

以 `screenshotPreferUIAElementHitInfo`（0x1409b2ce0，928B）为首个落地样本。

```
# 1) 抽取原函数字节码与只读表（门禁：0 call、表项无零、所有 rip lea 目标已声明）
python tools/parity/emu/extract_func.py 0x1409b2ce0 0x3a0 screenshotPreferUIAElementHitInfo

# 2) 生成用例（角坐标直给，避免 image.Rect 规范化语义分歧）
python tools/parity/emu/gen_cases.py > tools/parity/cases/screenshot_uia/puiahi.cases.txt

# 3) 生成期望值（rank 表由 step 1 的 .meta.json 表项机械推导）
python tools/parity/emu/ref_puiahi.py \
    tools/parity/bin/screenshotPreferUIAElementHitInfo.meta.json \
    < tools/parity/cases/screenshot_uia/puiahi.cases.txt \
    > tools/parity/cases/screenshot_uia/puiahi.expected.txt

# 4) 对拍
go test -count=1 -tags production -run TestParity ./backend
```

### 用例格式（四角坐标直给）

```
aminx aminy amaxx amaxy aCT aOK bminx bminy bmaxx bmaxy bCT bOK
```

不用「左上角 + 宽高」：Go 的 `image.Rect` 会交换 Min/Max 使矩形合法，而原函数入参
已由调用方构造完毕。直给四角可消除该规整语义在两套实现间的分歧，并使 `Min > Max`
的退化形态可被显式表达。

## 已知局限（诚实边界）

抽取器只支持**无 call 的叶子函数**。含 call 的函数需要更完整的加载器（符号解析、
外部调用桩），当前未实现，此类函数门禁直接失败而非静默降级。

参考实现的分支结构由人工按 asm 转写，与重建实现共用同一份 asm 阅读结论。因此对拍
验证的是「两次转写是否一致」，对「阅读结论本身是否有误」的覆盖依赖：抽取产物的表
数据（机械推导，独立来源）与变异测试。二者已纳入流程（见下）。

## 变异测试（对拍有效性自检）

对拍通过不等于对拍有效。抽样注入变异，确认能抓到：

| 变异 | 结果 |
|---|---|
| 从 rank=1 集合删去 50030 | 抓到（4 例不符） |
| 去掉 0xc364 提前出局分支 | 抓到（多例不符） |
| `areaB < areaA` 改 `<=` | 未抓到 —— 前有 `areaB != areaA` 守卫，该改写不可达，未抓到属正确 |

最后一行记为**已确证的等价改写**，不是漏检。
