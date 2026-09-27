# 批次 143 验收（二因素域：TOTP/HOTP/Steam 验证码生成链）

日期：2026-09-24
子批次：twofactor-code-gen-chain

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath -buildmode=exe ./backend` | EXIT=0 |
| vet | `go vet ./backend` | EXIT=0 |
| 全量 test | `go test ./backend` | ok (0.369s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1271 / MARKED=1271 / S=962 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

相对批次 142（1266/957/35/273/1）：FUNCS +5，S +5，其余不变（P=1 仍为 `newOLEDLifecycleContext`）。

新增 5 个（均 `backend/twofactor_code.go`）：
- `newTwoFactorHMAC` [S] 0x1409d3400
- `buildHOTPValue` [S] 0x1409d3280
- `buildStandardTwoFactorCode` [S] 0x1409d2f80
- `buildSteamTwoFactorCode` [S] 0x1409d30c0
- `buildTwoFactorCode` [S] 0x1409d2c80

## G3 逻辑等价

- `newTwoFactorHMAC(secret, algorithm)`：`sanitizeTwoFactorAlgorithm(algorithm,"totp")`
  （kind 常量 4B `totp` @0x140c34986）后 switch：`"SHA256"`（dword `0x32414853`="SHA2" +
  word `0x3635`="56"）→ `hmac.New(sha256.New, secret)`；`"SHA512"`（`0x35414853`="SHA5" +
  `0x3231`="12"）→ `hmac.New(sha512.New, secret)`；默认 → `hmac.New(sha1.New, secret)`。
  三个 hash 工厂地址连续（0x141096908/0x141096910/0x141096918 = sha1/sha256/sha512.New）。
- `buildHOTPValue(secret, counter, algorithm) (uint32, error)`：`make([]byte,8)` 后 bswap
  写入 big-endian counter → `mac.Write` → `mac.Sum(nil)` → `offset=sum[len-1]&0xf`（
  `movzx esi,[rbx+rax-1]` + `and esi,0xf`）→ `Uint32(sum[offset:offset+4])&0x7fffffff`
  （`bswap eax` + `and eax,0x7fffffff`）。Write 错误透传（`test rbx; jne` → `(0,err)`）。
- `buildStandardTwoFactorCode(secret, counter, digits, algorithm) (string, error)`：
  `mod=10^digits`（循环 `rbx=rbx*10`，`lea rsi,[rbx+rbx]`+`lea rbx,[rsi+rsi*4]`）→
  `code=hotp%uint32(mod)`（`xor edx,edx`+`div ecx` 32 位无符号）→
  `fmt.Sprintf("%0*d", digits, code)`（格式串 4B `%0*d` @0x140c34996，参数 [digits int64,
  code uint32]）。
- `buildSteamTwoFactorCode(secret, counter) (string, error)`：`buildHOTPValue(secret,counter,
  "SHA1")`（算法常量 4B `SHA1` @0x140c3498e）→ `v=int64(hotp)` → 5 轮 `steamAlphabet[v%26]`
  （`idiv r8` 有符号 64 位）+ `v/=26`。字母表 26B @0x141bc6980 =
  `23456789BCDFGHJKMNPQRTVWXY`。
- `buildTwoFactorCode(entry TwoFactorEntryConfig, secret []byte, now time.Time)
  (string, int64, error)`：`sanitizeTwoFactorKind(entry.Kind)` 后 period clamp（`"steam"`
  常量比较 `0x61657473`+"m" → 30；否则 `period-5` 无符号 ≤295 即 [5,300] 保持，越界 → 30）→
  `unix=now.Unix()`（wallToInternal `0xdd7b17f80` / internalToUnix `0xfffffff1886e0900`
  内联 `time.Time.Unix()`）→ `counter=unix/period`（idiv）→ `remaining=period-(unix%period)`
  （≤0 置 period）→ steam 走 `buildSteamTwoFactorCode(secret, uint64(counter))`，否则
  digits clamp（`digits-4` 无符号 ≤6 即 [4,10] 保持，越界 → 6）+
  `sanitizeTwoFactorAlgorithm(entry.Algorithm, kind)` 后
  `buildStandardTwoFactorCode(secret, uint64(counter), digits, algorithm)`。

## G4 复验

`twofactor_code_test.go` 新增 6 个测试函数：

- `TestBuildHOTPValueRFC4226`：RFC 4226 附录 D 全部 10 组向量（counter 0–9），
  decimal 31-bit 与 6-digit 双列断言。
- `TestBuildStandardTwoFactorCode`：6-digit `287082` + 8-digit `94287082`（RFC 6238 T=1
  官方值，验证 mod 10^digits 而非 6 位补零）。
- `TestBuildSteamTwoFactorCode`：counter=0 → `GG5F5` + 5 位长度/字母表字符集。
- `TestBuildTwoFactorCodeTotp`：`time.Unix(59,0)` → code `287082`、remaining `1`。
- `TestBuildTwoFactorCodeSteam`：`SteamGuard` kind 强制 period 30，remaining `30`。
- `TestBuildTwoFactorCodeClamp`：period 999→30、digits 12→6，仍得 `287082`/remaining `1`。

全量 `go test ./backend` PASS。

## 判定

四路 PASS，批次 143 闭环。TOTP/HOTP/Steam 验证码生成链全部 [S]，RFC 4226/RFC 6238 权威
向量双通过，与批次 141/142 的密钥链、加解密链形成完整二因素运行时依赖闭包。

## 下一步

- 批次 144：secret-decode 链剩余 9 个纯函数（decodeTwoFactorSecret 0x1409d4440 /
  resolveTwoFactorSecretDecoders 0x1409d4c20 / shouldPreferSteamBase64Secret 0x1409d4e00 /
  provisioningUsesSteamSharedSecret 0x1409d4ea0 / looksLikeTwoFactorBase32Secret 0x1409d5020 /
  looksLikeTwoFactorHexSecret 0x1409d50e0 / decodeTwoFactorHexSecret 0x1409d58a0 /
  extractTwoFactorProvisioningSecretText 0x1409d5200 / resolveSteamProvisioningSecretText
  0x1409d55a0）。
