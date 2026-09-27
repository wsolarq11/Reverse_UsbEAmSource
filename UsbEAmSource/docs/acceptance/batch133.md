# 批次 133 验收（配置存储域：冲突错误 Error + CompareAndSwapPrepared 签名取证）

日期：2026-09-23
子批次：config-store-cas-error

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.196s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1238 / MARKED=1238 / S=928 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。

相对批次 132（1237/927/35/274/1）：FUNCS +1，S +1。

新增：
- `launcherConfigConflictError.Error` [S]（`backend/types_launcher.go`）

## G3 逻辑等价

`launcherConfigConflictError.Error()`（[S 汇编 0x140898a00, 192B]）：
- `fmt.Sprintf("%s: expected=%s actual=%s", "CONFIG_REVISION_CONFLICT", e.Expected, e.Actual)`；
- 格式串 25B @0x140c66beb = `%s: expected=%s actual=%s`；
- prefix 24B @0x140c65118 = `CONFIG_REVISION_CONFLICT`（本轮 lea 精确重算：@0x140898a3c+0x936094
  = 0x1411CEAD0 → string 头 {0x140c65118, 24}，修正前轮 0x1411EEAD0 心算错误）。

**CompareAndSwapPrepared 真实签名取证（待 ReplacePrepared 域专项落地）**：
- 签名：`(expectedRevision string, prepare funcval, options, withIcons bool, saveWidgets bool,
  cfg LauncherConfig)`——rax=store、rbx/rcx=revision、rdi=prepare、rsi=options、r8=withIcons、
  r9=saveWidgets、栈=cfg（0x126 qword）。
- 体：`strings.TrimSpace(expectedRevision)`（@0x14089b0d9）→ 构造 CAS 闭包（fn=0x14089b280，
  捕获 {TrimSpace.ptr, TrimSpace.len}）→ 尾调 ReplacePrepared（@0x14089b16c）。
- CAS 闭包（0x14089b280）语义：al==0 → `errors.New("配置尚未初始化" 21B @0x140c5f5b7)`；
  `launcherConfigRevision(cfg)` 与期望 revision 比对（@0x14089b305/320 memequal）；
  期望为空或不等 → `launcherConfigConflictError{Expected: 期望, Actual: 实际}`；相等 → nil。
- 参数映射：ReplacePrepared(rbx=CAS闭包, rcx=原prepare, rdi=options, rsi=withIcons,
  r8=saveWidgets, 栈=透传)，与现有 6 参签名近似一致。

## G4 复验

`bootstrapservice_test.go` 新增 `TestLauncherConfigConflictErrorError`
（期望 `CONFIG_REVISION_CONFLICT: expected=rev-a actual=rev-b`）。

## 判定

四路 PASS，批次 133 闭环。

## 下一步（批次 134 方向）

- ReplacePrepared 域专项：按 0x140899980 反汇编还原 prepare/commit 闭包真实签名
  （asm 实证 prepare 接收 initialized bool 首参），随后 CompareAndSwapPrepared 签名升级。
- SaveConfig / ResetConfig 完整重建（接入 func1 prepare 回调 + CompareAndSwapPrepared 6 参调用）。
