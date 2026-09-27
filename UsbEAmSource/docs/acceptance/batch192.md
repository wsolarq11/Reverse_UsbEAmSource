# 批次 192 — pluginid + pluginsecurity 全部落地（8 函数 [S]，插件 ID/资源路径安全链闭环）

## 基线 / 收口

| 指标 | 基线（批次 191 收口） | 收口（批次 192） |
|---|---|---|
| FUNCS | 2732 | **2740** |
| S | 1181 | **1189** |
| S-inline | 36 | 36 |
| S-sig | 1400 | 1400 |
| P | 115 | 115 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2617 | 2625（55.2%） |

SHA256 `252F195D356BF353D3C50CBCEF724920653D0BC307FE49F7130C053C225DCB22`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批内容（pluginid 包并入 main + pluginsecurity 6 函数全部 [S]）

### 1. `pluginid.go`（原 `changeme/internal/pluginid`，2 funcs）— 完整落地

`ValidationError{Reason uint8; ReservedName string}`（Reason@0x00, ReservedName@0x08, size 0x18）+
`Error()`（0x1406dad80，384B）+ `Validate(id string) error`（0x1406daf00，896B）。校验顺序逐条 asm 对齐：
空→Reason1；>64B→Reason2；控制字符(<0x20/0x7f)→Reason4；非 ASCII(≥0x80)→Reason3；尾 '.'/' '→Reason6；
"."/".."/IndexAny("/\\:")→Reason5；首尾非字母数字→Reason7；其余字符非[字母数字._-]→Reason8；
首个 '.' 前段大写命中 AUX/CON/NUL/PRN/COM1-9/LPT1-9→Reason9。已用 22 组临时测试全通过（后删临时文件）。

### 2. `pluginsecurity.go`（6 funcs）— 完整落地

- `validatePluginID`（0x140923be0，672B）：Validate 通过→nil；errors.As 取 *ValidationError，Reason 映射中文消息。
- `normalizeValidatedPluginID`（0x140923e80，128B）：TrimSpace→validatePluginID，失败 (""，err)，成功 (trimmed，nil)。
- `readPluginFileBounded`（0x140923f00，1312B）：maxBytes≤0 拒；os.OpenFile(O_RDONLY,0)；Stat 后要求普通文件；
  Size<0 或 >maxBytes 拒；io.ReadAll(LimitReader(f,maxBytes+1)) 复核长度。
- `readPluginManifestBounded`（0x140924480，928B）：readPluginFileBounded(path,0x40000)；json.Decoder 解码
  PluginManifest；第二次 Decode 必须 io.EOF 否则拒多值/尾随；validatePluginID(m.ID)。
- `isPathWithinPluginRoot`（0x140924820，192B）：filepath.Rel 后拒 err/".."/IsAbs/"..\\" 前缀。
- `resolvePluginPathSecurely`（0x1409248e0，1792B）：path/root 去空白；root 反斜杠→正斜杠做绝对性检查→转回反斜杠
  Clean；拒绝对 root、"."/".."、上级逃逸；Join 后逐段 Lstat 拒 symlink/reparse point；EvalSymlinks 复核在 root 内；
  拒目录（allowDir=false）与非规则文件。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；`go test ./backend` `ok changeme/backend 0.901s`。
- **G2 count_funcs**：`FUNCS=2740 / S=1189 / S-inline=36 / S-sig=1400 / P=115 / UNMARKED=0`。
- **G3 行为**：pluginid 22 组校验用例全通过；pluginsecurity 6 函数全部 [S]；全部错误字符串/阈值（64B、0x40000、ModeType 0x8f280000、
  symlink bit 0x1b、"..\\" 前缀）逐条 asm 对齐。
- **G4 review**：新建 pluginid.go、pluginsecurity.go；pluginid 包并入 main（backend 唯一 package main，满足 count_funcs 口径）；
  复用已有 PluginManifest 类型；无跨文件写重叠。

## 遗留（下一批）

- §10 差集 60→**58**（pluginid.go、pluginsecurity.go 已落地）。
- 继续 `pluginhost.go`（12 funcs）、`pluginupdate.go`、`pluginwindow.go` 等插件域未落地文件。
- 其余遗留：`bytesMatchWildcardFold` 0x1407e3e60、`TestReminderNotification` 0x1407a7320、
  `validateLauncherConfigIconData`、`desktopCalendarStringList` 0x1407aa820。
