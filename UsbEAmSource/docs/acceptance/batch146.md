# 批次 146 验收（二因素域：twoFactorService 私有辅助方法签名落地）

日期：2026-09-25
子批次：twofactor-service-private-helpers

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath -buildmode=exe ./backend` | EXIT=0 |
| vet | `go vet ./backend` | EXIT=0 |
| 全量 test | `go test ./backend` | ok (1.451s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1302 / MARKED=1302 / S=977 / S-inline=35 / S-sig=289 / P=1 / UNMARKED=0`。

相对批次 145（1286/977/35/273/1）：FUNCS +16，S-sig +16，其余不变（P=1 仍为
`newOLEDLifecycleContext`，未新增任何 `[P]`）。

新增 16 个（均 `backend/twofactor_service_helpers.go`，全部 [S-sig]，receiver 均
`*twoFactorService`）：

| # | 方法 | 签名 | VA |
|---|---|---|---|
| 1 | buildState | `(cfg TwoFactorConfig, includeCodes bool) (TwoFactorState, error)` | 0x1409d0f40 |
| 2 | loadUnlockedConfig | `() (LauncherConfig, []byte, uint64, error)` | 0x1409d0c80 |
| 3 | buildLockedState | `(cfg TwoFactorConfig) TwoFactorState` | 0x1409d15e0 |
| 4 | ParseProvisioningText | `(text string) (TwoFactorProvisioningParseResult, error)` | 0x1409cd160 |
| 5 | resolveTwoFactorExportEntryIcon | `(entry TwoFactorEntryConfig) (TwoFactorEntryConfig, error)` | 0x1409cfc00 |
| 6 | setSessionKeyForOperation | `(gen uint64, key string) bool` | 0x1409d9900 |
| 7 | beginSessionOperation | `()` | 0x1409d9860 |
| 8 | sessionKeySnapshot | `() ([]byte, uint64)` | 0x1409d9bc0 |
| 9 | clearSessionKey | `()` | 0x1409d9e00 |
| 10 | clearSessionKeyIfGeneration | `(gen uint64) bool` | 0x1409d9f00 |
| 11 | sessionOperationMatches | `(op uint64) bool` | 0x1409da0e0 |
| 12 | sessionGenerationMatches | `(gen uint64) bool` | 0x1409da220 |
| 13 | previewTimeState | `() TwoFactorTimeState` | 0x1409da360 |
| 14 | resolvePreviewCurrentTime | `() (TwoFactorTimeState, time.Time)` | 0x1409da5a0 |
| 15 | resolveCurrentTime | `(force bool) (TwoFactorTimeState, time.Time, error)` | 0x1409da820 |
| 16 | fetchInternetTime | `() (twoFactorTimeCache, error)` | 0x1409db420 |

## G3 逻辑等价（签名实证）

本批为签名落地（[S-sig]），体为恒返零值骨架（TOTP/密码加解密、会话密钥、存储域待整域
落地），无臆造逻辑。签名均经 asm 序言（寄存器 ABI 入参/返回值、内存类参数与结果溢出区）
+ 符号表实证：

- `loadUnlockedConfig` 返回内存类大结构 `LauncherConfig`（0x126 qword 透传）+ 会话密钥
  `[]byte`(AX/BX/CX) + 代际 `uint64`(DI) + `error`(SI/R8)，四返回值。
- `setSessionKeyForOperation` 参数序经寄存器实证为 `gen`（→RBX）在前、`key string`
  （→RCX/RDI）在后。
- `resolveCurrentTime` / `resolvePreviewCurrentTime` 额外返回 `time.Time`（AX/BX/CX 三寄存器）。
- 会话代际三方法（`clearSessionKeyIfGeneration` / `sessionOperationMatches` /
  `sessionGenerationMatches`）为纯状态比较，单 `uint64` 入参、`bool` 返回。

## G4 复验

`go test ./backend` 全量 PASS（ok 1.451s）。本批为 [S-sig] 骨架，无新增行为路径，
不新增测试；既有测试全部通过确认无回归。

## 判定

四路 PASS，批次 146 闭环。16 个 `twoFactorService` 私有辅助方法签名全部经 asm 实证，
`[P]` 计数保持 1、`UNMARKED=0`，三项门禁 EXIT=0。

## 下一步

- 批次 147：二因素域包级 parse 链 + URI 链（parseTwoFactorProvisioningToken 0x1409d67e0 /
  parseTwoFactorImportText 0x1409d5c40 / buildTwoFactorDraftFromConfig 0x1409d7060 /
  buildOTPAuthEntryFromProvisioning 0x1409d79e0 / buildSteamEntryFromProvisioning 0x1409d7540 /
  build URI 链 buildTwoFactorProvisioningURI 0x1409d8520 / buildTOTPProvisioningURI 0x1409d8660 /
  buildSteamProvisioningURI 0x1409d8d80 等），另含 normalize/duplicate/allocate/filter/time
  辅助函数。
- 批次 148+：qrcode 屏幕选区/标注域（118 缺失函数，Win32 GDI + RGBA 绘制）。
