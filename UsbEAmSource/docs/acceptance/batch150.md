# 批次 150 验收（二因素域：normalize/parse/time 包级函数签名落地）

日期：2026-09-25
子批次：twofactor-provision-misc

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `bash build.sh`（`go build -tags production -trimpath -buildmode=exe ./backend`） | EXIT=0 |
| vet | `go vet ./backend` | EXIT=0 |
| 全量 test | `go test ./backend` | ok (0.345s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1444 / MARKED=1444 / S=1021 / S-inline=35 / S-sig=387 / P=1 / UNMARKED=0`。

相对批次 149（1432/1021/35/375/1）：FUNCS +12，S-sig +12，其余不变（P=1 仍为
`newOLEDLifecycleContext`，UNMARKED=0）。

新增 12 个（均 `backend/twofactor_provision_misc.go`，全部 [S-sig]，签名经 asm 序言实证）：

| # | 函数 | 签名 | VA |
|---|---|---|---|
| 1 | buildTwoFactorDuplicateKey | `(entry TwoFactorEntryConfig, key []byte) (string, error)` | 0x1409d4080 |
| 2 | allocateTwoFactorEntryID | `(entries []TwoFactorEntryConfig, names []string) string` | 0x1409d58e0 |
| 3 | normalizeTwoFactorEntryDraft | `(d TwoFactorEntryDraft) (TwoFactorEntryConfig, []byte, error)` | 0x1409d3640 |
| 4 | normalizeTwoFactorEntryMetadata | `(d TwoFactorEntryDraft) TwoFactorEntryConfig` | 0x1409d37c0 |
| 5 | filterTwoFactorEntriesByIDs | `(entries []TwoFactorEntryConfig, ids []string) []TwoFactorEntryConfig` | 0x1409d9480 |
| 6 | parseTwoFactorProvisioningToken | `(text string) (TwoFactorEntryConfig, []byte, error)` | 0x1409d67e0 |
| 7 | parseTwoFactorImportText | `(text string) ([]twoFactorParsedImportEntry, error)` | 0x1409d5c40 |
| 8 | extractTwoFactorImportTokens | `(text string) ([]string, error)` | 0x1409d5f80 |
| 9 | doTwoFactorTimeRequest | `(access LauncherNetworkAccess, req *http.Request, u *url.URL) (*http.Response, error)` | 0x1409dd4e0 |
| 10 | fetchTwoFactorTimeSample | `(access LauncherNetworkAccess, name string, url string) (twoFactorTimeSample, error)` | 0x1409dc600 |
| 11 | isTrustedTwoFactorTimeURL | `(u *url.URL, reference *url.URL) bool` | 0x1409dd720 |
| 12 | selectTwoFactorTimeSamples | `(samples []twoFactorTimeSample, now time.Time) (twoFactorTimeCache, error)` | 0x1409dbbe0 |

## G3 逻辑等价（签名实证）

- `doTwoFactorTimeRequest`/`fetchTwoFactorTimeSample` 的 requester 类型实证为
  `LauncherNetworkAccess`（`types_network.go` 的 `Do`/`Get` 接口）；`doTwoFactorTimeRequest`
  对其 type switch 解析 Do 后委托调用。
- `selectTwoFactorTimeSamples` 经 3 条 ret（2 条 error + 1 条 nil-error）确认返回
  `(twoFactorTimeCache, error)` 而非仅 cache。
- `isTrustedTwoFactorTimeURL` 两入参均 `*url.URL`（读 Scheme@0/User@0x20/Host@0x28），
  校验双方 scheme=https、无 userinfo、host 非空且 EqualFold 相等。
- 保留不确定性（已在注释标注，均属命名/形态层非类型层）：`allocateTwoFactorEntryID` 第二参
  命名、`fetchTwoFactorTimeSample` 的 name/url 展开 vs `twoFactorTimeTarget` 传值（ABI 等价）、
  `doTwoFactorTimeRequest` 第三参仅被闭包捕获。

## G4 复验

`go test ./backend` 全量 PASS（ok 0.345s）。本批为 [S-sig] 骨架，无新增行为路径；
既有测试全通过确认无回归。

## 判定

四路 PASS，批次 150 闭环。12 个二因素 normalize/parse/time 包级函数签名落地，`FUNCS=1444`
（30.38%），`P=1` 未新增、`UNMARKED=0`，三项门禁 EXIT=0。

## 下一步

二因素域包级函数已基本收口（provision label/URI/entry/parse/import/time 链全部落地为
[S]/[S-sig]）。剩余工作转向：二因素 12 个公开方法的正文（TOTP/密码加解密/会话密钥/存储域，
当前 [S-sig] 骨架）、配置存储域挂起项，或大域备选 filesearch/screenshot 剩余。
