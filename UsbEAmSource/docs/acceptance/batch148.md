# 批次 148 验收（二因素域：包级 URI/entry 构建链签名落地）

日期：2026-09-25
子批次：twofactor-provision-uri

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `bash build.sh`（`go build -tags production -trimpath -buildmode=exe ./backend`） | EXIT=0 |
| vet | `go vet ./backend` | EXIT=0 |
| 全量 test | `go test ./backend` | ok (0.389s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1373 / MARKED=1373 / S=990 / S-inline=35 / S-sig=347 / P=1 / UNMARKED=0`。

相对批次 147（1361/990/35/335/1）：FUNCS +12，S-sig +12，其余不变（P=1 仍为
`newOLEDLifecycleContext`）。

新增 12 个（均 `backend/twofactor_provision_uri.go`，全部 [S-sig]，签名经 asm 序言实证）：

| # | 函数 | 签名 | VA |
|---|---|---|---|
| 1 | buildTwoFactorProvisioningURI | `(secret []byte, cfg TwoFactorEntryConfig) string` | 0x1409d8520 |
| 2 | buildSteamProvisioningURI | `(secret []byte, cfg TwoFactorEntryConfig) string` | 0x1409d8d80 |
| 3 | buildTOTPProvisioningURI | `(secret []byte, cfg TwoFactorEntryConfig) string` | 0x1409d8660 |
| 4 | buildTwoFactorLabel | `(steam bool, cfg TwoFactorEntryConfig) string` | 0x1409d9360 |
| 5 | encodeTwoFactorProvisioningIcon | `(iconData string) string` | 0x1409d9100 |
| 6 | decodeTwoFactorProvisioningIcon | `(iconData string) string` | 0x1409d92a0 |
| 7 | buildTwoFactorDraftFromConfig | `(secret string, cfg TwoFactorEntryConfig) TwoFactorEntryDraft` | 0x1409d7060 |
| 8 | buildOTPAuthEntryFromProvisioning | `(u *url.URL, query url.Values) (TwoFactorEntryConfig, []byte, error)` | 0x1409d79e0 |
| 9 | buildSteamEntryFromProvisioning | `(u *url.URL, query url.Values) (TwoFactorEntryConfig, []byte, error)` | 0x1409d7540 |
| 10 | buildTwoFactorDraftPreviewEntry | `(now int64, offset int64, sampleCount int, draft TwoFactorEntryDraft) (TwoFactorEntryState, error)` | 0x1409d3c60 |
| 11 | buildTwoFactorEntryStateWithCode | `(code string, secondsRemaining int, cfg TwoFactorEntryConfig) TwoFactorEntryState` | 0x1409d1a40 |
| 12 | buildTwoFactorEntryMetadataStates | `(entries []TwoFactorEntryConfig) []TwoFactorEntryState` | 0x1409d17c0 |

## G3 逻辑等价（签名实证）

- #3 `buildTOTPProvisioningURI` 读 Algorithm@0x90/Digits@0xa0/Period@0xa8 证明入参为
  `TwoFactorEntryConfig` 而非 Draft（两型在部分偏移同形）。
- #10 `buildTwoFactorDraftPreviewEntry` 三寄存器字此前存疑，本次经
  `normalizeTwoFactorEntryDraft`（返回 normalizedDraft[栈]+secret[]byte[ax/bx/cx]+error[di/si]）
  与 `buildStandardTwoFactorCode` 反查敲定 arg4..6 = (now, offset, sampleCount)，非 secret。
- 保留不确定性（已在注释标注）：#4 bool 与结构相对顺序（Config/Draft 在 asm 同形）、
  #5/#6 图标 kind 映射 RIP 字符串字面量未解析。

## G4 复验

`go test ./backend` 全量 PASS（ok 0.389s）。本批为 [S-sig] 骨架，无新增行为路径；
既有测试全通过确认无回归。

## 判定

四路 PASS，批次 148 闭环。12 个包级 URI/entry 构建链函数签名落地，`[P]` 保持 1、
`UNMARKED=0`，三项门禁 EXIT=0。

## 下一步

- 批次 149：qrcode 下集（`backend/qrcode_native_b.go`，59 函数）。
- 批次 150：二因素 normalize/parse/time 包级函数（`backend/twofactor_provision_misc.go`，12）。
