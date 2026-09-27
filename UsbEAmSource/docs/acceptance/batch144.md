# 批次 144 验收（二因素域：provision secret 解码链）

日期：2026-09-24
子批次：twofactor-secret-decode-chain

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `bash build.sh`（`go build -tags production -trimpath -buildmode=exe ./backend`） | EXIT=0，产物 `artifacts/UsbEAm_Launcher_rebuilt.exe` |
| vet | `go vet ./backend` | EXIT=0 |
| 全量 test | `go test ./backend` | ok (3.509s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1283 / MARKED=1283 / S=974 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

相对批次 143（1271/962/35/273/1）：FUNCS +12，S +12，其余不变（P=1 仍为
`newOLEDLifecycleContext`）。

新增 12 个（均 `backend/twofactor_provisioning.go`，全部 [S]）：9 个 secret-decode 主链 +
3 个闭链必需依赖（parse/query 链）。

主链 9 个：
- `decodeTwoFactorHexSecret` 0x1409d58a0
- `looksLikeTwoFactorBase32Secret` 0x1409d5020
- `looksLikeTwoFactorHexSecret` 0x1409d50e0
- `provisioningUsesSteamSharedSecret` 0x1409d4ea0
- `shouldPreferSteamBase64Secret` 0x1409d4e00
- `resolveSteamProvisioningSecretText` 0x1409d55a0
- `extractTwoFactorProvisioningSecretText` 0x1409d5200
- `resolveTwoFactorSecretDecoders` 0x1409d4c20
- `decodeTwoFactorSecret` 0x1409d4440

依赖 3 个：
- `parseTwoFactorProvisioningURL` 0x1409d6ba0
- `parseTwoFactorRawQuery` 0x1409d7480
- `firstNonEmptyQueryValue` 0x1409d83e0

## G3 逻辑等价

- `decodeTwoFactorHexSecret(s)`：`hex.DecodeString(stripTwoFactorSecretGrouping(s))`。
- `looksLikeTwoFactorBase32Secret(s)`：`ToUpper(strip(s))` 空 → false；逐 rune 须 ∈
  [A-Z]∪[2-7]，任一越界 false（越界判定 asm 用 `lea edi,[rdx-0x41]; cmp edi,0x19` 与
  `add edx,-0x32; cmp edx,5`）。
- `looksLikeTwoFactorHexSecret(s)`：`strip(s)` 空或奇数长度（`bt ebx,0; jb`）→ false；
  逐 rune 须 ∈ [0-9a-fA-F]。
- `parseTwoFactorProvisioningURL(text)`：TrimSpace→ToLower；若 HasPrefix "steam://"(8B
  @0x140c3c30c)/"steamguard://"(13B @0x140c4e1c7)，去前缀后 rest 截断到首个 "/?#"
  (@0x140c33d3a)；rest 含 "%"（@0x1411ca458）→ 拼 `trimmed[:len(prefix)]+"/"+rest`
  （把裸密钥 host 转 path 以保留 URL 编码），否则原样；`url.Parse`。asm 用
  `concatstring3` 三参拼接，经 capstone 复核 RIP 目标（此前手算 0x6bfb 目标误为
  0x140c3630c，实测 0x140c3c30c）。
- `parseTwoFactorRawQuery(rawQuery)`：TrimSpace 空 → `(make(url.Values),nil)`；否则
  `Replace(q,"+"(0x1411caca0),"%2B"(0x140c33d3d),-1)` 后 `url.ParseQuery`（先转 %2B 再
  解析，令字面 '+' 不被当空格）。
- `firstNonEmptyQueryValue(values,names)`：遍历 names，`values[name]` 空切片跳过，否则
  `TrimSpace(values[name][0])` 非空即返回（asm `mapaccess1_faststr` + `[rax+8]` 判切片
  len + `[rax]`/`[rax+8]` 取首元素）；全空 → ""。
- `resolveSteamProvisioningSecretText(u,query)`：先 firstNonEmptyQueryValue(["secret",
  "shared_secret","sharedSecret","key"])；空则 `TrimSpace(u.Host)` 非空返回；再
  `TrimSpace(u.Path)` 若以 '/' 开头去首个 '/'（asm `dec rbx`+`neg/sar/and` 无分支实现
  `path[1:]`）；返回 path。
- `provisioningUsesSteamSharedSecret(text)`：TrimSpace 空 → false；ToLower 后非
  otpauth:///steam:///steamguard:// 前缀 → false；parse URL + RawQuery 任一失败 → false；
  否则 firstNonEmptyQueryValue(["shared_secret","sharedSecret"]) != ""。
- `shouldPreferSteamBase64Secret(provisioningText,secret)`：
  `provisioningUsesSteamSharedSecret(provisioningText)` 真 → true；否则
  `IndexAny(secret,"+/=_-"(0x140c35d99))>=0`。
- `resolveTwoFactorSecretDecoders(provisioningText,secret,kind)`：返回
  `[]func(string)([]byte,error)`（3 元素）；`sanitizeTwoFactorKind(kind)!="steam"` →
  [base32,base64,hex]；steam 且 shouldPreferSteamBase64Secret → [base64,base32,hex]；
  steam 且 looksLikeHex 且 !looksLikeBase32 → [hex,base32,base64]；否则 [base32,base64,hex]。
  解码器函数值 @0x141096a08/0x141096a10/0x141096a18，其首 word 分别指向
  `decodeTwoFactorBase32Secret`(0x1409d56c0)/`decodeTwoFactorBase64Secret`(0x1409d5720)/
  `decodeTwoFactorHexSecret`(0x1409d58a0)。
- `extractTwoFactorProvisioningSecretText(text,kind) (string,bool)`：TrimSpace 空 → ("",false)；
  ToLower 后非 otpauth:///steam:///steamguard:// 前缀 → ("",false)；parse URL+RawQuery 失败
  → ("",false)；scheme=ToLower(TrimSpace(u.Scheme)) 按 len 分支（5B "steam" / 7B "otpauth" /
  10B "steamguard"）：steam/steamguard → resolveSteam(真)；otpauth → host=ToLower(TrimSpace
  (u.Host)) 须 "totp"/"steam" 否则 ("",false)，host=="steam" 或 sanitizeTwoFactorKind(kind)
  =="steam" → resolveSteam(真)，否则 firstNonEmptyQueryValue(4 名)(真)。
- `decodeTwoFactorSecret(text,kind) ([]byte,error)`：TrimSpace 得 trimmed；extract 成功用提取
  文本否则原始 text（asm `cmovne`）；normalizeTwoFactorSecretText 后空 →
  `errors.New("二步验证密钥不能为空")`(@0x140c6fa3b)；resolveTwoFactorSecretDecoders
  (trimmed,secret,kind) 依次解码，首个 err==nil 且 len>0 返回；全失败 →
  `errors.New("无法识别二步验证密钥格式")`(@0x140c78aa7)。

## G4 复验

`twofactor_provisioning_test.go` 新增 11 个测试函数：

- `TestDecodeTwoFactorHexSecret`："48656C6C6F" → "Hello"；分组 "48 65 6c 6c 6f" → "Hello"。
- `TestLooksLikeTwoFactorBase32Secret`：7 组真/假向量（含小写容错、'=' 越界、'1' 越界）。
- `TestLooksLikeTwoFactorHexSecret`：7 组真/假向量（含奇数长度 "48656"、非 hex "48zz"）。
- `TestParseTwoFactorRawQuery`："a=1&b=2" 双键；"s=x+y" 保留字面 '+'；空 query 得空 map。
- `TestFirstNonEmptyQueryValue`：跳过空白值取 "x"；缺失键得 ""。
- `TestResolveSteamProvisioningSecretText`：query secret / host 兜底 / path 去前导 '/'。
- `TestProvisioningUsesSteamSharedSecret`：otpauth shared_secret 真、steam shared_secret
  query 真、steam 裸 host 假、非法文本假、空假。
- `TestShouldPreferSteamBase64Secret`：含 '+' 真、纯 base32 假、shared_secret provision 真。
- `TestResolveTwoFactorSecretDecoders`：非 steam/totp 首解码器 base32 → "Hello"；steam hex
  （"AB12CD"）首解码器 hex 3 字节；steam base32 首解码器 base32 → "Hello"。
- `TestExtractTwoFactorProvisioningSecretText`：otpauth URI 提 "JBSWY3DP"；steam:// 提
  "JBSWY3DP"；空文本 false。
- `TestDecodeTwoFactorSecret`：base32 "JBSWY3DP"→"Hello"、base64 "SGVsbG8="→"Hello"、
  otpauth URI→"Hello"、steam:// URI→"Hello"、空文本 err、乱码 "!!!!" err。

全量 `go test ./backend` PASS。

## 判定

四路 PASS，批次 144 闭环。secret-decode 链 9 函数 + parse/query 3 依赖共 12 个全部 [S]，
补齐了二因素域「provision URI → 密钥文本 → 解码」最后一环，与批次 140–143 的归一化链、
加解密链、验证码生成链构成完整二因素运行时依赖闭包。

## 下一步

- 批次 145：`twoFactorService` 大方法（SetupPassword 730L / ChangePassword 625L /
  ExportEntries 1315L / 其余 import/export/provisioning），需先 dump 0x1409c7ca0 之后的
  service 方法 asm 再按依赖链推进。
- 亦可先收口二因素域剩余小纯函数池（parseTwoFactorProvisioningToken 0x1409d67e0 /
  parseTwoFactorProvisioningDraft 0x1409d6dc0 / buildTwoFactorDraftFromConfig 0x1409d7060 /
  buildSteamEntryFromProvisioning 0x1409d7540 / buildOTPAuthEntryFromProvisioning 0x1409d79e0 /
  decodeProvisioningLabel 0x1409d8260 / splitProvisioningLabel 0x1409d8300 /
  parseTwoFactorOptionalInt 0x1409d84c0 等）。
