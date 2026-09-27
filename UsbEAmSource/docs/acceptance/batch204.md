# 批次 204 — twofactor 供给 URI 图标编解码 [S-sig]→[S]（encode/decodeTwoFactorProvisioningIcon）

## 基线 / 收口

| 指标 | 基线（批次 203 收口） | 收口（批次 204） |
|---|---|---|
| FUNCS | 2769 | 2769 |
| S | 1232 | **1234** |
| S-inline | 36 | 36 |
| S-sig | 1387 | **1385** |
| P | 114 | 114 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2655 | 2655（55.8%） |
| §10 差集 | 54 | 54 |

SHA256 `79D49F3A62BEFA1721AD14F65700AFE33DD69E589592AEC554BB191341F70697`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B，`bash build.sh` 重建）。

## 本批内容

`backend/twofactor_provision_uri.go` 两个供给 URI 图标编解码存根升档（体完整翻译，
此前为 [S-sig] 空骨架，内部 RIP 相对字符串字面量未解析）。FUNCS 不变，S +2 / S-sig -2。

字符串字面量经 lea 目标重算 + va_read 实证：`"totp"` 4B @0x140c34986、`"base64:"` 7B
@0x140c39568、`"data:image/"` 11B @0x140c47539、`","` 1B @0x1411cac20、`";base64"` 7B
@0x140c39457。

### `encodeTwoFactorProvisioningIcon`（0x1409d9100，416B）— [S-sig]→[S]

`func encodeTwoFactorProvisioningIcon(iconData string) string`：

1. `sanitizeTwoFactorEntryIconData("totp", iconData)` 空 → `""`。
2. `lower=ToLower(s)`；`HasPrefix(lower,"base64:")` → `"base64:" + TrimSpace(s[7:])`。
3. `HasPrefix(lower,"data:image/")` → `Cut(s,",")` 且 found 且 `Contains(ToLower(before),";base64")`
   → `"base64:" + TrimSpace(after)`。
4. 其余 → `""`。

### `decodeTwoFactorProvisioningIcon`（0x1409d92a0，192B）— [S-sig]→[S]

`func decodeTwoFactorProvisioningIcon(iconData string) string`：

1. `TrimSpace` 空 → `""`。
2. 否则 `sanitizeTwoFactorEntryIconData("totp", s)`。

asm 中 `lower=ToLower(s)` 与 `HasPrefix(lower,"base64:")` 两分支均返回
`sanitizeTwoFactorEntryIconData("totp", s)`（历史遗留死代码，lea 目标同 0x140c34986），
故合并为单一路径，行为等价。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2769 / S=1234 / S-inline=36 / S-sig=1385 / P=114 / UNMARKED=0`。
- **G3 行为**：五处字符串字面量 va_read 逐字节实证（见上）；控制流逐寄存器追踪
  （sanitize 返回 → ToLower → HasPrefix(base64:/data:image/) → Cut/Index/TrimSpace →
  concatstring2），与 `strings` 标准库语义对齐。
- **G4 review**：仅改 `backend/twofactor_provision_uri.go`（新增 import strings +
  两函数体）；无跨文件写重叠；未动同文件其余 10 个存根；vet/test/build 复验通过。

## 遗留（下一批）

- §10 差集 54 文件。desktopwidgets（calendar/clock）、filesearch 索引计划域、qrcode 剪贴板
  读写链、twofactor 供给 URI 图标编解码均已闭环。候选：`buildTwoFactorLabel`
  （0x1409d9360）、`matchNodeNameTermsWithPinyin`（0x14080fe40，filesearch_pinyin_windows.go
  域）、`pluginupdate.go`、`pluginwindow.go`、`twofactor_provision_misc.go` 各存根、
  `transport.go`、`TestReminderNotification`（0x1407a7320）。
