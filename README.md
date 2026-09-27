# Reverse_UsbEAmSource

UsbEAm Launcher 1.0.3 的逆向还原工作区。口径：可编译 + 功能一致重建，非字节级同哈希。

## 目录

```
UsbEAmSource/   主工程（Go 后端 + 还原的 Vue 前端）
work/           反汇编 dump、Ghidra 工程与脚本、一次性分析脚本
external/       Ghidra 12.1.3 安装与原始压缩包（第三方，不参与编译）
archive/        历史脚本、依赖备份、拆分前源码备份
```

各目录的职责、路径约定与本次整理的旧新映射见 `UsbEAmSource/docs/STRUCTURE.md`。

## 还原策略

总进度 100% = 4,754 蓝图函数全部 `[S]` 忠实还原。推进采用 L0→L1→L2 分层（存根 `[P]`/`[S-sig]` → 行为等价 `[S-eq]` → 忠实还原 `[S]`），口径与验证范式以 `UsbEAmSource/docs/STRATEGY.md` 为唯一真相源。新会话接手前必读。

## 构建

```bash
cd UsbEAmSource
bash build.sh
```

产物在 `UsbEAmSource/artifacts/UsbEAm_Launcher_rebuilt.exe`。构建依赖本地代理 `127.0.0.1:7890` 与精确工具链 `go1.25.12`，详见 `UsbEAmSource/docs/HANDOFF.md`。

## 边界

仅限个人学习研究，不用于再分发、二次发布、授权绕过或商业化。目标程序版权归 DOGFIGHT360 所有。
