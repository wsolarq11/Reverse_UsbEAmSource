# 批次 138 验收（二因素域：looksLikeTwoFactorBase32Payload）

日期：2026-09-23
子批次：twofactor-base32-detect

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.187s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1242 / MARKED=1242 / S=933 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

相对批次 137（1241/932/35/273/1）：FUNCS +1，S +1。

新增：
- `looksLikeTwoFactorBase32Payload` [S]（`backend/twofactor_sanitize.go`）。

## G3 逻辑等价

`looksLikeTwoFactorBase32Payload`（[S 汇编 0x1409c8320, 0x9dB]）：
- TrimSpace 后空串返回 false（@0x1409c8340-0x1409c8353）；
- 逐 rune 校验（ASCII 快径 @0x1409c8365 / 多字节 decoderune @0x1409c8374）：
  接受 A-Z（`r-0x41`≤0x19）、a-z（`r-0x61`≤0x19）、2-7（`r-0x32`≤5）、`=`（0x3d），
  其余字符返回 false（@0x1409c83aa）；全部合法返回 true（@0x1409c83b2）。

## G4 复验

`twofactor_validate_test.go` 增补 `TestLooksLikeTwoFactorBase32Payload`（8 例：合法/小写/填充/
空/空格/非法字符/数字越界）。

## 判定

四路 PASS，批次 138 闭环。

## 下一步（批次 139 方向）

- `normalizeTwoFactorPasswordConfig`（0x1409c7240, 0x1e0B）落地。
- `sanitizeTwoFactorEntryIcon`（0x1409c7d40）/ `sanitizeTwoFactorEntryIconData`（0x1409c7e00）落地。
