# 批次 137 验收（二因素域：sanitizeTwoFactorKind / sanitizeTwoFactorAlgorithm）

日期：2026-09-23
子批次：twofactor-sanitize

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.184s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1241 / MARKED=1241 / S=932 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

相对批次 136（1239/930/35/273/1）：FUNCS +2，S +2。

新增：
- `sanitizeTwoFactorKind` [S]（`backend/twofactor_sanitize.go`）；
- `sanitizeTwoFactorAlgorithm` [S]（同上）。

## G3 逻辑等价

`sanitizeTwoFactorKind`（[S 汇编 0x1409c7ca0, 0x77B]）：
- TrimSpace→ToLower；=="steam"(5B)/"steamguard"(10B，0x6175676d61657473+"rd") 返回 "steam"，
  其余一律 "totp"。返回字面量 "steam" 5B @0x140c35d8f、"totp" 4B @0x140c34986。

`sanitizeTwoFactorAlgorithm`（[S 汇编 0x1409c83e0, 0xc8B]）：
- `sanitizeTwoFactorKind(kind)=="steam"` → 强制 "SHA1"；
- 否则 algorithm TrimSpace→ToUpper 后 "SHA256"/"SHA512" 原样返回（@0x1409c8450-0x1409c8482），
  其余 "SHA1"。返回字面量 "SHA1" 4B @0x140c3498e、"SHA256" 6B @0x140c37734、
  "SHA512" 6B @0x140c3773a。

## G4 复验

`twofactor_validate_test.go` 增补 `TestSanitizeTwoFactorKind`（7 例）与
`TestSanitizeTwoFactorAlgorithm`（7 例）。

## 判定

四路 PASS，批次 137 闭环。

## 下一步（批次 138 方向）

- `normalizeTwoFactorPasswordConfig`（0x1409c7240, 0x1e0B）落地。
- `sanitizeTwoFactorEntryIcon`（0x1409c7d40）/ `sanitizeTwoFactorEntryIconData`（0x1409c7e00）落地。
- `looksLikeTwoFactorBase32Payload`（0x1409c8320）落地。
