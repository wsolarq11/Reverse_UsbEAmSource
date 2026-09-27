# 批次 247 — oledblackout ghost [P] 清零（ensureLifecycleLocked [S]）

## 基线 / 收口

| 指标 | 基线（批次 246 收口） | 收口（批次 247） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | **1257** |
| S-inline | 36 | 36 |
| S-sig | 1455 | 1455 |
| P | 41 | **40** |
| UNMARKED | 0 | 0 |

SHA256（`bash build.sh` 重建 `artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B）
`457410aa5600be56a7ecc0c127bbce0bec9e86d5f7c53ff5133ba1284acbbc85`。

## 本批内容

`backend/oledblackout.go`：

- 删除 ghost 脚手架 `newOLEDLifecycleContext`（`[P]`，无独立符号，实为方法）。
- 新增 `(*oledBlackoutService).ensureLifecycleLocked`（`[S] 0x1408ff220`，单 receiver 无参无返回）。
- `newOLEDBlackoutService` 调用点由 `s.lifecycleContext, s.lifecycleCancel = newOLEDLifecycleContext()`
  改为 `s.ensureLifecycleLocked()`（与汇编 0x1408fef60 流程对齐）。

## asm 证据（0x1408ff220，完整翻译）

- morestack 序言仅存 1 槽（rax）＝单 receiver 方法，无参无返回。
- `cmp [rax+0xb0],0`：shutdownDone==nil → `makechan`（0x1408ff24b）赋 `[0xb0]`。
- `cmp [rax+0xa8],0`：lifecycleGeneration==0 → 置 1。
- `cmp [rax+0x90],0`：lifecycleContext==nil → `context.WithCancel(Background)`（0x1408ff2b1），
  ctx 存 `[0x90]/[0x98]`、cancel 存 `[0xa0]`。
- `cmp byte [rdx+0x88],0`：shuttingDown!=0 → `call [rcx]`（lifecycleCancel()）。

字段偏移与 `types_oled.go` `oledBlackoutService` 对齐：[0x88]=shuttingDown、
[0x90/0x98]=lifecycleContext、[0xa0]=lifecycleCancel、[0xa8]=lifecycleGeneration、
[0xb0]=shutdownDone。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build/vet ./backend` EXIT=0；`test ./backend` `ok changeme/backend`。
- **G2 count_funcs**：`FUNCS=2788 / S=1257 / S-inline=36 / S-sig=1455 / P=40 / UNMARKED=0`。
- **G3 行为**：纯脚手架→方法订正，无调用方（仅 newOLEDBlackoutService）变更，既有测试全量 PASS。
- **G4 review**：`oledblackout.go`（newOLEDBlackoutService 流程与注释 0x1408fef60 对齐）。

## 遗留（下一批）

- P=40（oledblackout_windows 25 / screenshot_windows 7 / screenshot_uia_windows 6 / mousegestures 1 /
  nativedrag_windows 1）。
- oledblackout_windows 的 `*Context` 系 [P] 为闭包展开（捕获结构体 + 方法字段），批量解码需先定
  捕获结构体布局；`createTemporaryDirectoryShortcut`（nativedrag）为 0 寄存器参数 + 大结构体返回。
- §10 差集 57 文件；FUNCS 缺口 697 个未落地顶层函数 + 闭包/方法。
