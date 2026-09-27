# 批次 142 验收（二因素域：密钥解密 / 存储态归一化 / 遗留修复链）

日期：2026-09-24
子批次：twofactor-secret-decode-chain

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath -buildmode=exe ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.353s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1266 / MARKED=1266 / S=957 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

相对批次 141（1261/952/35/273/1）：FUNCS +5，S +5，其余不变（P=1 仍为 `newOLEDLifecycleContext`）。

新增 5 个：
- `decryptTwoFactorSecret` [S] 0x1409d2780（`backend/twofactor_crypto.go`，AES-GCM 解密）
- `normalizeStoredTwoFactorSecretBytes` [S] 0x1409d47c0（`backend/twofactor_secret.go`）
- `repairLegacySteamSecretBytes` [S] 0x1409d48a0（`backend/twofactor_secret.go`）
- `normalizeTwoFactorSecretText` [S] 0x1409d4600（`backend/twofactor_secret.go`）
- `looksLikeStrictTwoFactorBase32Secret` [S] 0x1409d4b60（`backend/twofactor_secret.go`）

## G3 逻辑等价

- `looksLikeStrictTwoFactorBase32Secret`：`ToUpper(TrimSpace(s))` 空 → false；逐 rune 须落在
  `[A-Z]` 或 `[2-7]`（`lea edi,[rdx-0x41]` 判 ≤0x19 + `add edx,-0x32` 判 ≤5），越界 → false。
- `normalizeTwoFactorSecretText`：`strings.NewReplacer` 删除 4 种单字符空白
  `' '`(0x1411ca5c8)/`'\t'`(0x1411ca430)/`'\r'`(0x1411ca988)/`'\n'`(0x1411ca400)，作用于
  TrimSpace 后文本（每次内联构造 Replacer：make([]string,8)+newobject+Replace）。
- `repairLegacySteamSecretBytes`：len==0 → (nil,false)；对 secret 的 `RawStdEncoding`(0x141c0f850)
  /`RawURLEncoding`(0x141c0f858) 两种 base64 编码，`TrimRight(cand,"=")`（cutset 1B @0x1411cac18）
  后 normalizeTwoFactorSecretText；map[string]struct{} 去重后仅当
  looksLikeStrictTwoFactorBase32Secret 且 decodeTwoFactorBase32Secret 成功且
  `len(decoded)<len(secret)` 时返回 (decoded,true)。（legacy 格式：secret 字节为 base32 文本被
  RawStd base64 解码所得，base64 编码还原文本后 base32 解码修复。）
- `normalizeStoredTwoFactorSecretBytes(kind, secret)`：`sanitizeTwoFactorKind(kind)=="steam"`
  （`0x61657473`+"m" 常量比较）时 repairLegacySteamSecretBytes，修复成功返回修复结果，否则原样。
- `decryptTwoFactorSecret(entry, key)`：validateTwoFactorStoredEntry → aes.NewCipher →
  cipher.NewGCM；nonce base64 解码须 == NonceSize；ciphertext base64 解码须
  Overhead<=len<=65536；gcm.Open(nil, nonce, ciphertext, nil)；plaintext 空报错；
  返回 normalizeStoredTwoFactorSecretBytes(entry.Kind, plaintext)。四个错误均
  `fmt.Errorf("%w: …", errTwoFactorDataCorrupted)`（%w 底层 0x141bc3df0/0x141bc3df8 同
  errTwoFactorDataCorrupted）：`二步验证条目的随机向量损坏`(43B)/`二步验证条目的密文损坏`(37B)/
  `二步验证条目认证失败`(34B)/`二步验证条目密钥为空`(34B)。

## G4 复验

`twofactor_secret_test.go` 新增 7 个测试函数：looksLikeStrict（8 组）、normalizeText（5 组）、
repairLegacy（"JBSWY3DP"→"Hello" 修复 + nil/随机字节不修复）、normalizeStored（steam 修复 /
totp 原样 / 非 legacy 原样）、decryptRoundTrip（encrypt→decrypt）、decryptErrors（非法条目 /
篡改密文）、base32 约定自检。全量 `-count=1` PASS。

## 判定

四路 PASS，批次 142 闭环。密钥解密 + 存储态归一化 + 遗留 Steam 修复依赖闭包全部 [S]，与
批次 141 的 encrypt 形成完整 AES-GCM round-trip。

## 下一步

- 批次 143：secret-decode 链剩余（decodeTwoFactorSecret 0x1409d4440 / resolveTwoFactorSecretDecoders
  0x1409d4c20 / shouldPreferSteamBase64Secret 0x1409d4e00 / provisioningUsesSteamSharedSecret
  0x1409d4ea0 / looksLikeTwoFactorBase32Secret 0x1409d5020 / looksLikeTwoFactorHexSecret
  0x1409d50e0 / decodeTwoFactorHexSecret 0x1409d58a0 / extractTwoFactorProvisioningSecretText
  0x1409d5200 / resolveSteamProvisioningSecretText 0x1409d55a0）。
- 批次 143+：code-gen 链（buildTwoFactorCode 0x1409d2c80 / buildStandardTwoFactorCode 0x1409d2f80 /
  buildSteamTwoFactorCode 0x1409d30c0 / buildHOTPValue 0x1409d3280 / newTwoFactorHMAC 0x1409d3400）。
