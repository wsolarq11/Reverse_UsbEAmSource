# 批次 205 — twofactor 供给 URI 标签构造 [S-sig]→[S]（buildTwoFactorLabel）

## 基线 / 收口

| 指标 | 基线（批次 204 收口） | 收口（批次 205） |
|---|---|---|
| FUNCS | 2769 | 2769 |
| S | 1234 | **1235** |
| S-inline | 36 | 36 |
| S-sig | 1385 | **1384** |
| P | 114 | 114 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2655 | 2655（55.8%） |
| §10 差集 | 54 | 54 |

SHA256 `8279C159F1E606A63D88AE86C7BAD3C4699BBD7D5E94F52375EFAA79C30CAA2F`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B，`bash build.sh` 重建）。

## 本批内容

`backend/twofactor_provision_uri.go` 的 `buildTwoFactorLabel` 存根升档（体完整翻译，
此前为 [S-sig] 空骨架）。FUNCS 不变，S +1 / S-sig -1。

### `buildTwoFactorLabel`（0x1409d9360，288B）— [S-sig]→[S]

`func buildTwoFactorLabel(steam bool, cfg TwoFactorEntryConfig) string`：

1. `issuer = TrimSpace(cfg.Issuer)`、`accountName = TrimSpace(cfg.AccountName)`、
   `name = TrimSpace(cfg.Name)`（栈偏移 0x30/0x40/0x20 对应 Config 字段，实证）。
2. `steam` 为真 → `issuer = "Steam"`（**大写**）。
3. `issuer != "" && accountName != ""` → `issuer + ":" + accountName`。
4. `accountName != ""` → `accountName`。
5. `issuer != ""` → `issuer`。
6. 否则 → `name`。

## 关键知悉（重要修正）

- `"Steam"` 5B @0x140c35d94（raw_hex `537465616d`，**大写 S**），与 sanitize 域的
  `"steam"` 小写 @0x140c35d8f 是**两个不同字面量**。原注释「steam 为真时 issuer 固定为
  'steam'」有误，已随本批修正为 `"Steam"`（大写）。
- 分隔符 `":"` 1B @0x1411cac58（raw_hex `3a`）。
- 控制流经 asm 四分支逐寄存器追踪：`test rcx; je`（issuer 空）→ `test r9; jne`（accountName
  非空）→ concatstring3 / 直接返回 issuer / accountName / name，四态全覆盖。
- 入参 `steam bool` 在 AL（`[rsp+0x138] = al`），cfg 值传栈 0xd0，返回 string（rax/rbx）。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`（0.359s）；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2769 / S=1235 / S-inline=36 / S-sig=1384 / P=114 / UNMARKED=0`。
- **G3 行为**：两处字符串字面量 va_read 逐字节实证；"Steam" 大写经 raw_hex 二次确认。
- **G4 review**：仅改 `backend/twofactor_provision_uri.go`（一函数体 + 注释修正）；无跨文件
  写重叠；未动同文件其余 9 个存根；vet/test/build 复验通过。

## 遗留（下一批）

- §10 差集 54 文件。desktopwidgets（calendar/clock）、filesearch 索引计划域、qrcode 剪贴板
  读写链、twofactor 供给 URI（图标编解码 + 标签构造）均已闭环。候选：
  `matchNodeNameTermsWithPinyin`（0x14080fe40，filesearch_pinyin_windows.go 域）、
  `pluginupdate.go`、`pluginwindow.go`、`twofactor_provision_misc.go` 各存根、
  `transport.go`、`TestReminderNotification`（0x1407a7320）。
