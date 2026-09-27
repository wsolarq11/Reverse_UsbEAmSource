# 批次 140 验收（二因素域：规范化链收尾）

日期：2026-09-24
子批次：twofactor-normalize-chain

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath -buildmode=exe ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.205s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1254 / MARKED=1254 / S=945 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

相对批次 139（1243/934/35/273/1）：FUNCS +11，S +11，其余不变（P=1 仍为 `newOLEDLifecycleContext`）。

新增 11 个（`backend/twofactor_normalize.go`）：
- `normalizeTwoFactorConfig` [S] 0x1409c65a0
- `normalizeTwoFactorPasswordConfig` [S] 0x1409c7240
- `normalizeTwoFactorEntryConfigs` [S] 0x1409c7420
- `normalizeTwoFactorEntryConfig` [S] 0x1409c77a0
- `sanitizeTwoFactorEntryIconData` [S] 0x1409c7e00
- `normalizeTwoFactorOpaqueIconPayload` [S] 0x1409c7f60
- `normalizeTwoFactorOpaqueIconPayloadWithDecoder` [S] 0x1409c8120
- `resolveTwoFactorDisplayName` [S] 0x1409c84e0
- `stripTwoFactorSecretGrouping` [S] 0x1409d51a0
- `decodeTwoFactorBase32Secret` [S] 0x1409d56c0
- `decodeTwoFactorBase64Secret` [S] 0x1409d5720

## G3 逻辑等价

- `normalizeTwoFactorPasswordConfig`：Salt/Verifier TrimSpace，任一空返回零值；KDF 空默认
  `"argon2id-v1"`（11B @0x140c47683）；Blank 透传。
- `normalizeTwoFactorEntryConfig`：kind 归一化；name 空则
  `resolveTwoFactorDisplayName(kind, issuer, accountName)`（优先 accountName→issuer→kind 默认
  steam→"steam"/其余→"lock"）；steam 清空 iconRef/iconURL、digits=5、period=30；非 steam
  digits 限 4..10（默认 6）、period 限 5..300（默认 30）。
- `sanitizeTwoFactorEntryIconData`：steam 空；lower 前缀 `data:image/`（@0x140c47539）原样保留；
  `base64:`（@0x140c39568）去前缀交 normalizeTwoFactorOpaqueIconPayload；其余空。
- `normalizeTwoFactorOpaqueIconPayloadWithDecoder`：解码失败/空 → ("",false)；解码文本已是
  `data:image/` → sanitize("totp", s)；否则 StdEncoding 重编码 + `http.DetectContentType` 判
  `image/`（@0x140c375f0）→ 包装 `data:<mime>;base64,<b64>`（5B @0x140c35bd7 + 8B @0x140c3c0cc），
  非图像 → `base64:<b64>`。
- `decodeTwoFactorBase64Secret`：TrimSpace 空 → errors.New("empty secret")（@0x140c4bb4c）；
  依次 RawStdEncoding/RawURLEncoding 解码，失败补 `=` 到 4 倍数后 URLEncoding/StdEncoding 兜底。
  （编码器身份由 GOROOT go1.25.12 base64.go 声明顺序 Std/URL/RawStd/RawURL 与 .data 地址
  0x141c0f840..858 对照锁定。）

## G4 复验

`twofactor_normalize_test.go` 新增 12 个测试函数：密码配置空组件/默认 KDF、steam 特判、
digits/period 钳制、name 解析、条目去重、图标数据 sanitize、显示名解析、base32/base64 解码、
空密钥错误。全量 `-count=1` PASS。

## 判定

四路 PASS，批次 140 闭环。twofactor.go 蓝图 140 函数中规范化链（normalize×4 + sanitize×2 +
resolve×1 + decode×2 + strip×1 + opaque×2）全部 [S]。

## 下一步

- 批次 141+：`twoFactorService` 12 个大方法（SetupPassword 730L / ChangePassword 625L /
  ExportEntries 1315L 等）无 asm 资产，需现场 dump 后按依赖链推进；或转其他已就绪大域
  （filesearch / screenshot 剩余 / qrcode）。
- 挂起：CompareAndSwapPrepared 签名升级、loadUnlocked 真实签名专项。
