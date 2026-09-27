# 批次 203 — qrcode 剪贴板读写链 [S-sig]→[S]（readQRCodeTextFromClipboard / writeQRCodeTextToClipboard）

## 基线 / 收口

| 指标 | 基线（批次 202 收口） | 收口（批次 203） |
|---|---|---|
| FUNCS | 2769 | 2769 |
| S | 1230 | **1232** |
| S-inline | 36 | 36 |
| S-sig | 1389 | **1387** |
| P | 114 | 114 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2655 | 2655（55.8%） |
| §10 差集 | 54 | 54 |

SHA256 `C664417959FF917CBCFAF2CC72577EC561840F1A4C9CEF7DEA296310C9FA952E`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B，`bash build.sh` 重建）。

## 本批内容

`backend/qrcode_windows.go` 的 "qrcode.go domain" 段两个剪贴板存根升档（体完整翻译，
此前为 [S-sig] 空骨架）。FUNCS 不变，S +2 / S-sig -2。

### `readQRCodeTextFromClipboard`（0x140931960，96B）— [S-sig]→[S]

`func readQRCodeTextFromClipboard() (string, error)`：

1. `initQRCodeClipboard()` 失败 → `("", err)`（asm 0x140931978 分支 rcx/rdi 装配 error type/data）。
2. 否则 `string(clipboard.Read(clipboard.FmtText))`，返回 nil error。
   asm：`eax=0`（FmtText=iota 0）→ `clipboard.Read` 返回 []byte 3 寄存器 →
   `runtime.slicebytetostring`（cap 丢弃）→ string 2 word + nil error。

### `writeQRCodeTextToClipboard`（0x1409319c0，128B）— [S-sig]→[S]

`func writeQRCodeTextToClipboard(text string) error`：

1. `initQRCodeClipboard()` 失败 → 透传 err。
2. 否则 `clipboard.Write(clipboard.FmtText, []byte(text))`（丢弃返回 channel），return nil。
   asm：`runtime.stringtoslicebyte(0, ptr, len)` → `clipboard.Write(FmtText=0, []byte)` →
   xor eax/ebx（nil error）。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`（0.980s）；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2769 / S=1232 / S-inline=36 / S-sig=1387 / P=114 / UNMARKED=0`。
- **G3 行为**：两函数控制流逐寄存器追踪（qrcodeClipboard.asm 96 行），FmtText=iota 0
  （eax=0 与 `clipboard.go:128` 的 `FmtText Format = iota` 一致）、Read 返回 []byte 3 word、
  Write 返回 <-chan struct{} 丢弃，均与 golang.design/x/clipboard v0.8.0 实际 API 对齐
  （`func Read(t Format) []byte` / `func Write(t Format, buf []byte) <-chan struct{}`）。
- **G4 review**：仅改 `backend/qrcode_windows.go`（新增 import clipboard + 两函数体）；
  无跨文件写重叠；未动其余 qrcode_windows.go 存根；vet/test/build 复验通过。

## 遗留（下一批）

- §10 差集 54 文件。desktopwidgets（calendar/clock）、filesearch 索引计划域、qrcode 剪贴板
  读写链均已闭环。候选：`matchNodeNameTermsWithPinyin`（0x14080fe40，filesearch_pinyin_windows.go
  域）、`pluginupdate.go`、`pluginwindow.go`、`twofactor.go`、`transport.go`、
  `TestReminderNotification`（0x1407a7320，依赖 launcherWidgetStore.Read +
  normalize/validateDesktopReminderAudio 链）。
