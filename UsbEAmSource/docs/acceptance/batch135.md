# 批次 135 验收（二因素域：validateTwoFactorStoredConfig 签名校正 + 体落地）

日期：2026-09-23
子批次：twofactor-stored-validation

## G1 门禁

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go build -tags production -trimpath ./backend` | EXIT=0 |
| vet | `go vet -tags production ./backend` | EXIT=0 |
| 全量 test | `go test -tags production -count=1 -p=1 ./backend` | ok (0.178s) |

## G2 函数覆盖

`bash tools/count_funcs.sh`：`FUNCS=1239 / MARKED=1239 / S=929 / S-inline=35 / S-sig=274 / P=1 / UNMARKED=0`。

相对批次 134（1238/928/35/274/1）：FUNCS +1，S +1。

变更：
- `validateTwoFactorStoredConfig` [S-sig]→[S]（签名校正 + 完整体）；
- `validateTwoFactorStoredEntry` [S-sig]（新增，签名实证）；
- `errTwoFactorDataCorrupted` sentinel（`TWO_FACTOR_DATA_CORRUPTED`）。

## G3 逻辑等价

`validateTwoFactorStoredConfig`（[S 汇编 0x1409c6740, 0x594B]）：
- **签名校正**：参数由 `LauncherConfig` 改为 `(TwoFactorPasswordConfig, []TwoFactorEntryConfig)`。
  汇编入口拷贝 56B 栈参（KDF@+0 Salt@+0x10 Verifier@+0x20 Blank@+0x30 = TwoFactorPasswordConfig）
  + entries slice（ptr/len，stride 0xd0 = TwoFactorEntryConfig 208B）。
- flag = Blank || TrimSpace(KDF)!="" || TrimSpace(Salt)!="" || TrimSpace(Verifier)!=""。
- flag 真：KDF 须空或 `"organi2d-v1"`（11B 字面量比较 @0x1409c6849-0x1409c686a）；
  Salt base64 解码须 16B；Verifier base64 解码须 32B。
- flag 假：Salt/Verifier 非空报 "password 配置不完整"。
- `len(entries)>0 && !flag` → "存在条目但密码配置缺失"。
- 逐条 `validateTwoFactorStoredEntry`，失败包裹 `entries[%d]: %w`。

**错误串解码**（全部 %w 包裹同一 sentinel "TWO_FACTOR_DATA_CORRUPTED" 25B @0x140c66993）：
- `%w: password.kdf 无效` 23B @0x140c630f7；
- `%w: password.salt 无效` 24B @0x140c65688；
- `%w: password.verifier 无效` 28B @0x140c6c100；
- `%w: password 配置不完整` 28B @0x140c6c11c；
- `%w: 存在条目但密码配置缺失` 37B @0x140c7a19c；
- `entries[%d]: %w` 15B @0x140c52e63。

## G4 复验

新增 `twofactor_validate_test.go`（6 用例）：空配置/合法 KDF-Salt-Verifier/非法 KDF/盐长度/验证器长度/
有条目缺密码，均校验 sentinel 包裹（`errors.Is`）。

## 判定

四路 PASS，批次 135 闭环。

## 下一步（批次 136 方向）

- `validateTwoFactorStoredEntry` 体落地（[S-sig]→[S]，0x1409c6ce0, 0x560B）。
- `normalizeTwoFactorPasswordConfig`（0x1409c7240）与 `normalizeTwoFactorEntryConfigs`
  （0x1409c7420）落地，补齐二因素保存规范化链。
