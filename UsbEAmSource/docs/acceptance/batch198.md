# 批次 198 — qrexternal.go 二维码外部 URL 域 4 函数 [S] + 独立成文件

## 基线 / 收口

| 指标 | 基线（批次 197 收口） | 收口（批次 198） |
|---|---|---|
| FUNCS | 2762 | **2762** |
| S | 1217 | **1221** |
| S-inline | 36 | 36 |
| S-sig | 1395 | **1391** |
| P | 114 | 114 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2648 | 2648（55.7%） |
| §10 差集 | 57 | **56** |

SHA256 `18C1130FB400012F8C2FE93B8B7B4AD19A4680669E1F534E5589E0D291BA9538`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,279,744 B，`bash build.sh` 重建）。

## 本批内容

qrexternal.go 域 4 函数 [S-sig]→[S]，并从 qrcode_windows.go 拆分为独立 `backend/qrexternal.go`
（差集 -1，源文件组织对齐 source_funcs.txt 的 `File: qrexternal.go`）：

### `OpenQRCodeExternalURL`（0x14095eda0，128B）— [S-sig]→[S]

`(b *BootstrapService) OpenQRCodeExternalURL(url string) error`：normalize 后 err 非 nil 原样
返回，否则 `openWithDefaultHandler(normalized)`（复用 bootstrap_pathclip.go）。

### `normalizeQRCodeExternalURL`（0x14095ee20，1792B）— [S-sig]→[S]

`(raw string) (string, error)`，9 类错误路径逐一经 va_read 原始字节还原中文消息：

1. TrimSpace 空或 != raw → 「链接不能为空或包含首尾空白」。
2. 含不安全文本 → 「链接包含不安全字符」。
3. 双重 `url.PathUnescape`（mode=encodePathSegment，对应 asm `net/url.unescape` mode=2）
   至多 2 次：含不安全转义 → 「链接包含不安全的转义字符」。
4. 转义非法 → 「链接转义无效」。
5. `url.Parse` 失败或 Opaque 非空 → 「链接格式无效」。
6. Scheme=lower(trim(Scheme)) 非 http/https → 「只允许 HTTP 或 HTTPS 链接」。
7. User 非 nil → 「链接不能包含用户名或密码」。
8. host=trim(Hostname()) 空 → 「链接必须包含有效主机名」。
9. `net.ParseIP` 命中 → `ip.String()`；否则 `idna.Lookup.ToASCII`（err 或 trim 后空 →
   「链接主机名无效」）。

host 重建：含 `:`（IPv6）→ `[host]` 或 `[host]:port`；否则 `ToLower(host)` + 可选
`:port`；返回 `u.String()`。

关键实证：idna Profile 全局变量指针经 `mov [rip+0x1264444]` 精算为 0x141bc3560 →
0x141be4440，其 options 布尔位 `00 01 01 01 00 00`（transitional=false, useSTD3Rules/
checkHyphens/checkJoiners=true）与 mapping=validateAndMap(0x14071dda0)、
fromPuny=validateFromPunycode(0x14071ea60)、bidirule=ValidString(0x14071cde0) 精确对应
`idna.Lookup`。

### `containsUnsafeQRCodeURLText`（0x14095f520，224B）— [S-sig]→[S]

`(s string) bool`：`strings.Contains(s, "\\")` 命中 → true；否则按 rune 判定
`r <= 0x1f || (r >= 0x7f && r <= 0x9f)`（C0 控制 / DEL / C1 控制）。256B 位图
@0x14196c560 经逐项解码 65 个 `&1` 命中 = 0x00-0x1f ∪ 0x7f-0x9f；rune>0xff 视为安全。

### `qrExternalURLInvalidError`（0x14095f600，192B）— [S-sig]→[S]

`(url string) error`：TrimSpace 空 → 「二维码链接无效」；`errors.New("QR_EXTERNAL_URL_INVALID: "+s)`
（前缀 25B ASCII @0x140c66cfe）。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；
  `go test ./backend` `ok changeme/backend`（0.353s 真实跑）。
- **G2 count_funcs**：`FUNCS=2762 / S=1221 / S-inline=36 / S-sig=1391 / P=114 / UNMARKED=0`。
- **G3 行为**：纯标准库逻辑（net/url、net、strings、idna），无 Win32/COM 依赖；9 类错误
  消息与 idna.Lookup Profile 均经原始字节实证，未臆造。
- **G4 review**：拆分改动仅 qrcode_windows.go（删 4 存根 + 收窄 import）与新建
  qrexternal.go；无跨文件写重叠；vet/test 复验通过。

## 遗留（下一批）

- §10 差集 56 文件。qrcode.go 域仍有 2 个 [S-sig] 存根（readQRCodeTextFromClipboard /
  writeQRCodeTextToClipboard，Win32 剪贴板链）。
- 其余候选：`matchNodeNameTerms`（0x1407e3ac0）、`pluginupdate.go`、`pluginwindow.go`、
  `desktopwidgets_calendar.go`、`twofactor.go`（TOTP/HOTP/加密，较大）、`transport.go`。
