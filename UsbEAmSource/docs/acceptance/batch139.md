# 批次 139 验收（二因素域：sanitizeTwoFactorEntryIcon）

日期：2026-09-23
子批次：twofactor-entry-icon

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.177s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1243 / MARKED=1243 / S=934 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

相对批次 138（1242/933/35/273/1）：FUNCS +1，S +1。

新增：
- `sanitizeTwoFactorEntryIcon` [S]（`backend/twofactor_sanitize.go`）。

## G3 逻辑等价

`sanitizeTwoFactorEntryIcon`（[S 汇编 0x1409c7d40, 0x77B]）：
- `sanitizeTwoFactorKind(kind)=="steam"` → 返回 "steam"（5B @0x140c35d8f）；
- 否则 icon TrimSpace 后非空原样返回，空则返回 "lock"（4B @0x140c3498a）。

## G4 复验

`twofactor_validate_test.go` 增补 `TestSanitizeTwoFactorEntryIcon`（5 例：steam 空/steamguard/
totp 空→lock/自定义/去空格）。

## 判定

四路 PASS，批次 139 闭环。

## 下一步

- 批次 120–139 目标闭环（FUNCS=1243 / S=934 / S-sig=273 / P=1 / UNMARKED=0）。
- 后续批次：`normalizeTwoFactorPasswordConfig`（0x1409c7240）、
  `normalizeTwoFactorEntryConfigs`（0x1409c7420）、`sanitizeTwoFactorEntryIconData`（0x1409c7e00）。
- 挂起：CompareAndSwapPrepared 签名升级、loadUnlocked 真实签名专项（见批次 134 后清单）。
