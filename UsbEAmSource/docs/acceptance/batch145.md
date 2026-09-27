# 批次 145 验收（二因素域：provision label/int 辅助纯函数）

日期：2026-09-24
子批次：twofactor-provisioning-helpers

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `bash build.sh`（`go build -tags production -trimpath -buildmode=exe ./backend`） | EXIT=0，产物 `artifacts/UsbEAm_Launcher_rebuilt.exe` |
| vet | `go vet ./backend` | EXIT=0 |
| 全量 test | `go test ./backend` | ok (0.360s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1286 / MARKED=1286 / S=977 / S-inline=35 / S-sig=273 / P=1 / UNMARKED=0`。

相对批次 144（1283/974/35/273/1）：FUNCS +3，S +3，其余不变（P=1 仍为
`newOLEDLifecycleContext`）。

新增 3 个（均 `backend/twofactor_provisioning.go`，全部 [S]）：
- `decodeProvisioningLabel` 0x1409d8260（0xa0）
- `splitProvisioningLabel` 0x1409d8300（0xe0）
- `parseTwoFactorOptionalInt` 0x1409d84c0（0x60）

## G3 逻辑等价

- `decodeProvisioningLabel(s string) string`：`s=TrimSpace(s)`；若非空且 `s[0]=='/'`
  （asm `cmp byte [rax],0x2f`）→ 去前导 '/'（asm `dec rbx` + `neg/sar/and` 无分支实现
  `s[1:]`）；空 → ""；否则 `url.PathUnescape(s)`（`net/url.unescape` mode=2，即
  encodePathSegment）成功 → `TrimSpace(结果)`，失败 → 返回去 '/' 后的原始 s。
- `splitProvisioningLabel(s string) (string,string)`：`s=TrimSpace(s)`；空 → ("","")；
  `strings.Cut(s, ":"(1B @0x1411cac58))`（asm `internal/stringslite.Cut`）；未找到 → ("", s)
  （asm `xor eax,ebx` 后 after=trimmed）；找到 → (TrimSpace(before), TrimSpace(after))。
- `parseTwoFactorOptionalInt(s string) int`：`strconv.Atoi(TrimSpace(s))`；`err != nil`
  （asm `test rbx; je`）→ 0，否则返回解析值。

## G4 复验

`twofactor_provisioning_test.go` 新增 3 个测试函数：

- `TestDecodeProvisioningLabel`：7 组向量（去前导 '/'、空串、"/"、PathUnescape
  `%20`、非法 `%ZZ` 回退、两侧空白）。
- `TestSplitProvisioningLabel`：5 组向量（":" 拆分、两侧空白、无冒号回退、空串、
  多冒号 "A:B:C" → 首冒号后全给 account）。
- `TestParseTwoFactorOptionalInt`：5 组向量（数字、空白包裹、非数字→0、空→0、负数）。

`go test ./backend` 全量 PASS（两次独立运行：ok 0.360s / ok 0.289s）。

## 判定

四路 PASS，批次 145 闭环。3 个 provision 辅助纯函数全部 [S]，是批次 146+ parse 链
（parseTwoFactorProvisioningToken → buildOTPAuthEntryFromProvisioning → …）的直接依赖，
签名、分隔符 ":"、unescape mode=2 均经汇编 + capstone 常量复核。

## 已知既有 flaky（本批修复）

全量 `go test ./backend` 曾偶发 `TestConfigureChangeTriggersReschedule` 失败
（memoryrelease_batch19_test.go:329，断言 `s.timer == nil`）。根因与二因素域无关：
`newConfigurableSvc()` 注入 `now=2026-07-08`（过去时刻），而 `rescheduleLocked` 用
`time.Until(nextAt)`（真实时钟）算延迟 → `d<0` → `d=0` → `time.AfterFunc(0,f)` 立即调度
回调 → 回调经 `handleScheduledRun→runLocked→stopTimerLocked`（timer 置 nil）与测试在
`Configure` 返回后立即读 `s.timer` 产生竞态。修复：`newConfigurableSvc()` 注入时刻改为
远未来 `2030-01-01`（被测代码 `time.Until` 为 asm 实证，不可改；仅改测试侧注入值），
令 `d>0`、回调不立即触发。修复后全量 `go test ./backend -count=1` 连续 3 次 PASS
（1.866s / 0.443s / 0.447s）。

## 下一步

- 批次 146：parse 链大函数（parseTwoFactorProvisioningToken 0x1409d67e0 /
  buildOTPAuthEntryFromProvisioning 0x1409d79e0 / buildSteamEntryFromProvisioning
  0x1409d7540 / buildTwoFactorDraftFromConfig 0x1409d7060 / parseTwoFactorProvisioningDraft
  0x1409d6dc0），需先厘清 entry/draft 结构的栈返回布局与 UUID/时间依赖，再逐函数推进。
- 亦可先收口二因素域 build URI 链（buildTwoFactorProvisioningURI 0x1409d8520 /
  buildTOTPProvisioningURI 0x1409d8660 / buildSteamProvisioningURI 0x1409d8d80）。
