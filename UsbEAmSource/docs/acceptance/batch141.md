# 批次 141 验收（二因素域：密码 / 加解密 / 密钥派生链）

日期：2026-09-24
子批次：twofactor-crypto-chain

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath -buildmode=exe ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.368s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1261 / MARKED=1261 / S=952 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

相对批次 140（1254/945/35/273/1）：FUNCS +7，S +7，其余不变（P=1 仍为 `newOLEDLifecycleContext`）。

新增 7 个（`backend/twofactor_crypto.go`）：
- `randomBytes` [S] 0x1409d9760（crypto/rand.Read 包装）
- `zeroTwoFactorBytes` [S] 0x1409d97e0（clear 清零）
- `isTwoFactorPasswordConfigured` [S] 0x1409d1da0
- `deriveTwoFactorKey` [S] 0x1409d2320（argon2.IDKey 3/32768/4/32）
- `buildTwoFactorPasswordVerifier` [S] 0x1409d2400（sha256 + base64）
- `verifyTwoFactorPassword` [S] 0x1409d1e60
- `encryptTwoFactorSecret` [S] 0x1409d2560（AES-GCM）

## G3 逻辑等价

- `deriveTwoFactorKey`：`argon2.IDKey([]byte(TrimSpace(password)), salt, time=3, memory=32768(0x8000),
  threads=4, keyLen=32(0x20))`。mode=2=argon2id 由 x/crypto argon2.go `argon2d/argon2i/argon2id` iota
  常量（0/1/2）锁定；寄存器实证 time=3(r10)、memory=0x8000(r11)、threads=4(栈)、keyLen=0x20(栈)。
- `buildTwoFactorPasswordVerifier`：盐常量 `"usbeam-two-factor-password:"`（27B，movabs
  `0x742d6d6165627375/0x6f746361662d6f77/0x702d726f74636166/0x3a64726f77737361`）与 key 拼接后
  `sha256.Sum256`，`base64.StdEncoding` 编码前 32B。
- `verifyTwoFactorPassword`：normalize → isConfigured → KDF 须 `"argon2id-v1"`（movabs
  `0x6469326e6f677261` + `-v` + `1`）→ Salt base64 解码须 16B、Verifier base64 解码须 32B →
  `deriveTwoFactorKey` → `buildTwoFactorPasswordVerifier(key) != p.Verifier` 常量比较 → 返回 key。
  错误消息（UTF-8 实测）：`"请先设置二步验证密码"`(30B)、`fmt.Errorf("%w: 二步验证密码算法不受支持",
  errTwoFactorDataCorrupted)`(40B 格式串)、`"二步验证密码配置损坏"`(30B，salt/verifier 共用)、
  `"二步验证密码不正确"`(27B)。
- `encryptTwoFactorSecret(plaintext, key)`：aes.NewCipher → cipher.NewGCM →
  `nonce=randomBytes(NonceSize)` → `ciphertext=gcm.Seal(nil, nonce, plaintext, nil)` →
  返回 `(base64(nonce), base64(ciphertext), nil)`；任一步失败 `("", "", err)`。
- `randomBytes`：`make([]byte,n)` + `crypto/rand.Read`；`zeroTwoFactorBytes`：`clear(b)` 返回 b。

## G3-bis 既有缺陷修正

- `launcherconfig.go` `validateTwoFactorStoredConfig` KDF 判定 `"organi2d-v1"` → `"argon2id-v1"`
  （活体实证：validateTwoFactorStoredConfig 0x1409c684f movabs `0x6469326e6f677261` = "argon2id"，
  与 verifyTwoFactorPassword 同源；"organi2d-v1" 系此前抄写笔误）。连带 [S] 头长度
  `0x594B→0x5a0`、`0x53bB→0x560` 校正。
- `twofactor_validate_test.go` 3 处 KDF 固定值同步 `"argon2id-v1"`。

## G4 复验

`twofactor_crypto_test.go` 新增 7 个测试函数：randomBytes（长度/非全零/空）、zeroTwoFactorBytes
（清零/同底层数组/nil）、isTwoFactorPasswordConfigured（5 组）、deriveTwoFactorKey（32B/等价
argon2.IDKey/TrimSpace/确定性）、buildTwoFactorPasswordVerifier（独立 sha256 对照）、
verifyTwoFactorPassword（round-trip + 错误密码 + KDF/salt 错误分支）、encryptTwoFactorSecret
（独立 GCM 解密 round-trip）。全量 `-count=1` PASS。

## 判定

四路 PASS，批次 141 闭环。密码验证链（isConfigured → derive → verifier → verify）与 AES-GCM
加密链（encrypt + randomBytes + zero）全部 [S]，附带修正 1 处跨文件 KDF 抄写缺陷。

## 下一步

- 批次 142：`decryptTwoFactorSecret`（0x1409d2780，0x500）依赖已就绪
  `validateTwoFactorStoredEntry` + 未落地 `normalizeStoredTwoFactorSecretBytes`(0x1409d47c0，0xe0)，
  需先 dump 后者后一并落地。
- 后续小纯函数池：secret-decode 链（decodeTwoFactorSecret 0x1409d4440 / normalizeTwoFactorSecretText
  0x1409d4600 / repairLegacySteamSecretBytes 0x1409d48a0 / looksLikeStrictTwoFactorBase32Secret
  0x1409d4b60 / resolveTwoFactorSecretDecoders 0x1409d4c20 / decodeTwoFactorHexSecret 0x1409d58a0）、
  code-gen 链（buildTwoFactorCode 0x1409d2c80 / buildStandardTwoFactorCode 0x1409d2f80 /
  buildSteamTwoFactorCode 0x1409d30c0 / buildHOTPValue 0x1409d3280 / newTwoFactorHMAC 0x1409d3400）。
