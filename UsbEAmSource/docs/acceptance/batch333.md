# 批次 333 · 窗口布局动作/远程图标搜索/启动日志/窗口选项 +4（FUNCS 3089）

## 目标

落地 4 个短函数：`BootstrapService.ApplyLauncherWindowLayoutAction`（窗口布局动作）、
`BootstrapService.SearchRemoteIcons`（远程图标搜索）、
`openLauncherStartupDebugLog`（启动调试日志）、
`buildLauncherWindowOptions`（窗口选项构建）。

## 基线 / 收口

| 指标 | 基线（batch 332 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3085 | 3089 |
| MARKED | 3085 | 3089 |
| S | 1492 | 1492 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1514 | 1518 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1529 | 1529 |
| USABLE | 1530 | 1530 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1529  FUNCS=3089  MARKED=3089  P=41  S-eq=1  S-inline=37  S-sig=1518  S=1492  USABLE=1530
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1492 + 37 + 1 + 1518 + 41 = 3089 = FUNCS`。
S-sig 1514→1518（+4）、FUNCS 3085→3089（+4）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 BootstrapService.ApplyLauncherWindowLayoutAction [S-sig 0x14079a9e0, 224B]

resolveLauncherWindow → 类型断言查找 action 处理器 → applyLauncherWindowLayoutAction。

### 3.2 BootstrapService.SearchRemoteIcons [S-sig 0x14095f6c0, 256B]

搜索远程图标候选并返回结果集。

### 3.3 openLauncherStartupDebugLog [S-sig 0x1408d2da0, 256B]

全局 flag → os.OpenFile(全局路径, O_CREATE|O_APPEND|O_WRONLY, 0666) → log.Logger.output。

### 3.4 buildLauncherWindowOptions [S-sig 0x1408d0280, 256B]

duffcopy 选项模板 → 宽高 imul 缩放计算 → 返回 options 结构。

## G4 独立复核

- `backend/bootstrapservice_window.go`：+ApplyLauncherWindowLayoutAction [S-sig]。
- `backend/remoteicons.go`：+SearchRemoteIcons [S-sig]。
- `backend/hotkey_dispatch_stubs.go`：+openLauncherStartupDebugLog +buildLauncherWindowOptions [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3089/4754 = 64.98%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
