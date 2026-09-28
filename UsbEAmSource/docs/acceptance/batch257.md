# 批次 257 · 输入/远程图标/工作区迁移/天气浮点短函数 +4（128B）

## 目标

落地 4 个 128B top-level 函数：天气浮点解析（desktopWeatherFloat）、鼠标扩展按钮标签
（inputMonitorXButtonLabel）、远程图标内容类型校验（remoteIconContentTypeAllowed）、工作区
迁移 reparse 判定（workspaceMigrationPathIsReparse）。均为 gap_aggregate 长度升序 128B 段。

## 基线 / 收口

| 指标 | 基线（batch 256 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2838 | 2842 |
| MARKED | 2838 | 2842 |
| S | 1306 | 1310 |
| S-inline | 36 | 36 |
| S-sig | 1456 | 1456 |
| P | 40 | 40 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1342 | 1346 |
| 真函数（S+S-inline+S-sig） | 2798（58.86%） | 2802（58.94%） |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 本批落地（+4 [S]，1 追加 + 3 新文件）

### backend/desktopwidgets_weather.go（追加，+1 [S]）

1. `desktopWeatherFloat(s string) float64` `[S 0x1407c95c0]` = `strconv.ParseFloat(strings.TrimSpace(s), 64)`
   后做三层守卫：`f != f`（NaN）→ 0；`f > math.MaxFloat64`（+Inf）→ 0；`f < 0`（负数）→ 0；否则返回 f。
   asm 证据：`ucomisd f,f; jne/jp`（NaN）、`ucomisd f,MaxFloat64; ja`（上界，常量 0x1411CD838 =
   1.79769313486232E+308）、`ucomisd -0,f; jbe`（下界，常量 0x1411CD858 = 负零，即 `f >= 0` 才返回）。

### backend/inputmonitor_windows.go（新增，+1 [S]）

2. `inputMonitorXButtonLabel(xbutton uint32) string` `[S 0x1408695e0]` = `shr eax,0x10` 取 HIWORD 作
   XButton 编号：==1→"x1"、==2→"x2"，否则 "x"+`FormatInt(HIWORD,10)`。字符串常量（RIP-relative 计算
   校正后）"x1"@0x140C336A9、"x2"@0x140C336AB、前缀 "x"@0x140C33631。

### backend/remoteicons.go（新增，+1 [S]）

3. `remoteIconContentTypeAllowed(contentType, allowedType string) bool` `[S 0x140961fe0]` =
   `mime.ParseMediaType(strings.TrimSpace(contentType))`；`err != nil` → false；否则
   `strings.EqualFold(mediatype, allowedType)`。

### backend/workspacemigration_identity_windows.go（新增，+1 [S]）

4. `workspaceMigrationPathIsReparse(path string) bool` `[S 0x1409f4660]` =
   `windows.UTF16PtrFromString(path)` err → false；`windows.GetFileAttributes(ptr)` err → false；
   否则 `attrs & windows.FILE_ATTRIBUTE_REPARSE_POINT (0x400, bit10) != 0`。

## 关键知悉

- **RIP-relative 进位纠错**：本批字符串/浮点常量的目标地址初算多进 0x100000（把
  `0x869605+0x3ca0a6=0xC336AB` 误算成 `0xD096AB`），读到垃圾后按「disp32 加到下一条指令地址、
  勿忘低位进位」重算修正为 0x140C336xx / 0x1411CD8xx。这是批次 255 收口记录里点名的同型纪律。
- **desktopWeatherFloat 下界为负零（-0）**：`ucomisd -0,f; jbe` 等价 `f >= 0`，语义 = 排除负值；
  用 `f < 0 → 0` 直译即可（-0 与 +0 在 IEEE754 相等）。
- **workspaceMigrationPathIsReparse 与 pluginpath_reparse_windows.go 的 pluginPathIsReparse、
  launcherupdate_security_windows.go 的 launcherUpdatePathHasReparsePoint 同族**：本函数 error 时
  返 false（asm 返单字节 bool），后两者分别保守返 true / 返 (bool,error)，勿混淆。
- **timeFromWindowsTick 留待专项**：0x1408697a0 涉及 time.Time 内部 wall/monotonic 布局
  （`bt rax,0x3f` + 魔数除法），128B 内完成 UnixMilli 级转换，风险高于本批 4 函数，单独成批处理。

## 下一批

P=40 不变。继续按 gap_aggregate.txt 长度升序落地已存在文件短函数；下一批优先
windowmanagement_windows.go 剩余短函数（GetClassName 256B / MaybeWrapCursor 288B /
EnumDisplayMonitorProc 320B / GetCursorPoint 320B / GetWindowText 320B）。FUNCS 2842/4754 = 59.78%。
