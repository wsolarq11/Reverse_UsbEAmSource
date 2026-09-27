# 批次 233 — screenshotAccessibleLocation 补 VARIANT 参数

## 基线 / 收口

| 指标 | 基线（批次 232 收口） | 收口（批次 233） |
|---|---|---|
| FUNCS | 2788 | 2788 |
| S | 1256 | 1256 |
| S-inline | 36 | 36 |
| S-sig | 1432 | **1433** |
| P | 64 | **63** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2724 | **2725**（57.30%） |

SHA256 `75DF02BB96204D47C0A8FA4BFEC812B5381D6EE608532B0F3F9A72315F22486A`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建，21,321,728 B）。

## 本批内容

`backend/screenshot_uia_windows.go`：`screenshotAccessibleLocation` 补第二个形参
`screenshotOleVariant`（值传）。

### screenshotAccessibleLocation（0x1409b0ca0）

`func screenshotAccessibleLocation(acc *screenshotAccessible) (image.Rectangle, error)`
→ `func screenshotAccessibleLocation(acc *screenshotAccessible, v screenshotOleVariant) (image.Rectangle, error)`。

证据链：序言 bx/cx/di/si 各存 16 位（word）、r8 存 qword，与 screenshotOleVariant
（VT/R1/R2/R3 uint16 + Val int64）五寄存器字段展开一一对齐；rax=acc；返回 6 寄存器
(image.Rectangle 4 int, error 2 word)，错误路径 rect 清零 + itab/data，正常路径同样 6 寄存器。

## 关键知悉

- VARIANT 值传 = 4 个 uint16 各占独立寄存器低 16 位 + 1 个 int64 占第 5 寄存器，
  不是打包成 2 word；识别关键在序言 `mov word ptr [rsp+x], bx/cx/di/si` 的 16 位保存。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2788 / S=1256 / S-inline=36 / S-sig=1433 / P=63 / UNMARKED=0`。
- **G3 行为**：纯签名订正（体空→体空），无行为变更；既有测试全量回归 PASS。
- **G4 review**：`screenshot_uia_windows.go`（1 函数签名订正）。

## 遗留（下一批）

- P 已降至 63。剩余分布：oledblackout_windows 25、screenshot_windows 8、launcherupdate_runtime 9、
  screenshot_uia_windows 7、screenshot_pin 6、filelocator_runtime 2、qrcode_windows 1。
- 可快速突破口：filelocator walkRoot/processFile（值传大结构体 prepared + 额外 generation/ctx）；
  screenshotAccessibleHitTest（3 参数但 9 字返回）；storeSnapshotBounds（r9b 死参数，bool 未读）。
