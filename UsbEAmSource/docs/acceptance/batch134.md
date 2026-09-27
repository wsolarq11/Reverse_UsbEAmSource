# 批次 134 验收（配置存储域：ReplacePrepared prepare 签名校正）

日期：2026-09-23
子批次：config-store-prepare-signature

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.406s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1238 / MARKED=1238 / S=928 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。

本批为领域模型签名校正（无新增函数），计数与批次 133 持平。

## G3 逻辑等价（签名校正）

`(*launcherConfigStore).ReplacePrepared` prepare 首参校正：
- `prepare func(bool, LauncherConfig) error`（原 `func(LauncherConfig) error`）；
- 汇编实证：ReplacePrepared @0x140899c14 `call prepare` 前 al=loadUnlocked 返回的初始化标志
  （@0x140899a98 `[rsp+0x129e]=al`）；SaveConfig.func1 入口 @0x140777565 `test al`，
  al==0 → `errors.New("配置尚未初始化，请先完成初始化向导" 51B @0x140c88d4f)`；
  CAS 闭包 @0x14089b2a0 `test al`，al==0 → `errors.New("配置尚未初始化" 21B @0x140c5f5b7)`。
- 体内以 `cfg.Initialized` 承载 initialized（语义等价近似，asm al=loadUnlocked 独立返回；
  待 loadUnlocked 签名专项校正后对齐）。

**ReplacePrepared 完整流程取证**（@0x140899980，frame 0x3008）：
- 参数保存：rbx=prepare、rcx=commit、rdi=options、rsi=withIcons、r8=saveWidgets、栈=widgets
  （与现有 6 参签名一致，仅 prepare 首参类型校正）；
- store==nil → 错误路径；lock store.mu(+0x3c) → loadUnlocked → prepare(initialized,cfg)
  （err→unlock 返错）→ validateTwoFactorStoredConfig（@0x140899cd3）→ commit →
  savePreparedUnlocked。

## G4 复验

`bootstrapservice.go` commit 主函数 prepare 闭包同步改为
`func(initialized bool, cfg LauncherConfig) error`；build/vet/test 全绿。

## 判定

四路 PASS，批次 134 闭环。

## 下一步（批次 135 方向）

- CompareAndSwapPrepared 签名升级：6 参（revision+prepare+options+withIcons+saveWidgets+cfg）
  + 体 TrimSpace→CAS 闭包→ReplacePrepared；需先厘清 commit 参数真实类型
  （func1 值传 cfg vs ReplacePrepared commit 指针，asm @0x14089b14c rcx=原 prepare）。
- loadUnlocked 真实签名专项：asm 显示 3 返回值（cfg+bool+error），现简化为 2 返回值。
