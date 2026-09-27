# 批次 136 验收（二因素域：validateTwoFactorStoredEntry 体落地）

日期：2026-09-23
子批次：twofactor-entry-validation

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.183s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1239 / MARKED=1239 / S=930 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

相对批次 135（1239/929/35/274/1）：S +1，S-sig -1（validateTwoFactorStoredEntry [S-sig]→[S]）。

## G3 逻辑等价

`validateTwoFactorStoredEntry`（[S 汇编 0x1409c6ce0, 0x53bB]）：
- Kind ToLower 后须 `totp`/`steam`（字面量比较 @0x1409c6d27-0x1409c6d47）；
- Algorithm ToUpper 后须空/`SHA1`/`SHA256`/`SHA512`（@0x1409c6dc0-0x1409c6e14）；
- steam：Algorithm 空/`SHA1`、Digits 0/5、Period 0/30；
- totp：Digits 0/[4,10]（`Digits-4` 无符号 ≤6 @0x1409c6f51-0x1409c6f59）、
  Period 0/[5,300]（`Period-5` 无符号 ≤0x127 @0x1409c6f68-0x1409c6f73）；
- SecretNonce/SecretCiphertext 至少一项非空；nonce base64 解码须 12B；ciphertext 须 [16,65536]B。

**错误串解码**（均 %w 包裹 errTwoFactorDataCorrupted）：
- `%w: kind 无效` 15B @0x140c52e72；`%w: algorithm 无效` 20B @0x140c5db5b；
- `%w: Steam 参数无效` 22B @0x140c61554；`%w: TOTP 参数无效` 21B @0x140c5fa25；
- `%w: 缺少加密数据` 22B @0x140c6156a；`%w: secretNonce 无效` 22B @0x140c61580；
- `%w: secretCiphertext 无效` 27B @0x140c6a602。

## G4 复验

`twofactor_validate_test.go` 增补 7 用例（合法/非法 Kind/Algorithm/Steam Digits/空加密数据/Nonce 长度）。

## 判定

四路 PASS，批次 136 闭环。

## 下一步（批次 137 方向）

- `normalizeTwoFactorPasswordConfig`（0x1409c7240, 0x1e0B）落地。
- `normalizeTwoFactorEntryConfigs`（0x1409c7420, 0x380B）落地。
- `sanitizeTwoFactorKind`（0x1409c7ca0）/ `sanitizeTwoFactorEntryIcon`（0x1409c7d40）落地。
