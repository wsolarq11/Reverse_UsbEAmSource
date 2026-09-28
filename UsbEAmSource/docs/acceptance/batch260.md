# 批次 260 · 窗口样式名枚举链 +3 [S]（FlagNames / ExStyleNames / StyleNames）

## 目标

落地 windowmanagement_windows.go 名字枚举三件套：FlagNames（544B，通用位掩码→名字）、
StyleNames（128B，WS_* 17 条）、ExStyleNames（384B，WS_EX_* 8 条）。三者构成依赖链，
一次落地（ExStyleNames/StyleNames 均调用 FlagNames，单独落地会缺符号）。

## 基线 / 收口

| 指标 | 基线（batch 259 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2848 | 2851 |
| MARKED | 2848 | 2851 |
| S | 1316 | 1319 |
| S-inline | 36 | 36 |
| S-sig | 1456 | 1456 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1352 | 1355 |
| 真函数（S+S-inline+S-sig） | 2808（59.07%） | 2811（59.13%） |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 本批落地（+3 [S]，均追加到 windowmanagement_windows.go）

1. `windowManagementFlagNames(value uint32, names []windowManagementFlagName) string` `[S 0x1409eede0]` =
   value==0 → ""；遍历表，`e.Flag==0 || value&e.Flag != e.Flag` 跳过，否则收集 e.Name；
   末尾 `strings.Join(collected, " | ")`。表中先用 2 元素栈缓冲，超过则 growslice。

2. `windowManagementStyleNames(value uint32) string` `[S 0x1409eebe0]` =
   FlagNames(value, 17 条 WS_* 切片字面量)。

3. `windowManagementExStyleNames(value uint32) string` `[S 0x1409eec60]` =
   FlagNames(value, 8 条 WS_EX_* 切片字面量)。

新增类型 `windowManagementFlagName{Flag uint32; Name string}`（表中 stride=24 字节，实证一致）。

## 表内容（逐条内存实证，flag 值与 Win32 常量完全吻合）

StyleNames（17 条，顺序即二进制顺序）：WS_POPUP(0x08000000) / WS_CHILD(0x04000000) /
WS_MINIMIZE(0x02000000) / WS_VISIBLE(0x01000000) / WS_DISABLED(0x00800000) /
WS_CLIPSIBLINGS(0x00400000) / WS_CLIPCHILDREN(0x00200000) / WS_MAXIMIZE(0x00100000) /
WS_CAPTION(0x000C0000) / WS_BORDER(0x00080000) / WS_DLGFRAME(0x00040000) /
WS_VSCROLL(0x00020000) / WS_HSCROLL(0x00010000) / WS_SYSMENU(0x00008000) /
WS_THICKFRAME(0x00004000) / WS_MINIMIZEBOX(0x00002000) / WS_MAXIMIZEBOX(0x00001000)

ExStyleNames（8 条）：WS_EX_TOPMOST(0x8) / WS_EX_TRANSPARENT(0x20) / WS_EX_TOOLWINDOW(0x80) /
WS_EX_WINDOWEDGE(0x100) / WS_EX_CLIENTEDGE(0x200) / WS_EX_APPWINDOW(0x40000) /
WS_EX_LAYERED(0x80000) / WS_EX_NOACTIVATE(0x8000000)

分隔符：`" | "`（3B @0x140C33CEC，实证）。

## 结构忠实度订正（本批自查发现）

首版把两张表写成**包级 var 数组**再传 `table[:]`。但 asm 显示两表是**每次调用在栈上构建**：
StyleNames 用 `duffcopy`（runtime.duffcopy+0x222）从只读区复制 408 字节到栈；
ExStyleNames 用 `duffzero`（duffzero+0x12b）清零后逐条 mov 填入 192 字节。
故改为**切片字面量**直传（编译器同样生成栈上构建），结构对齐二进制。

## 关键纠错：RIP-relative 进位（第四次，已改用工具算术）

本批解析 StyleNames 表地址时，`0x1409EEC00 + 0x7F7610` 应为 `0x1411E6210`，我手算成
`0x141E6210`（漏进位），导致 dump 文件未生成/读空。

**纪律升级（治本）**：不再人工心算 RIP 目标，一律用 PowerShell `[long]` 显式加法并回显
`computed:` 地址，再校验读出内容的合理性（名字必须是可打印 ASCII、len 必须在 1..96）。
本批 ExStyle 表 4 条错址（str3/4/5/8）也是靠「读出 'eBracket;LeftUpD' 这类非 WS_EX_ 前缀
垃圾」发现并纠正的。

## 下一批

SetWindowLongPtr 544B（三 proc 已实证：SetLastError/SetWindowLongW/SetWindowLongPtrW，
格式串 "更新窗口样式失败: %w"）；DisplayRects 384B；MaybeWrapCursor 288B（依赖
WrappedCursorPoint 992B）。P=40。FUNCS 2851/4754 = 59.97%。
