# 批次 181 — 文件搜索归一化链收口（3 处 [S-sig]→[S] 升档）

## 基线 / 收口

| 指标 | 基线（批次 180 收口） | 收口（批次 181） |
|---|---|---|
| FUNCS | 2707 | 2707 |
| S | 1151 | **1154** |
| S-inline | 35 | 35 |
| S-sig | 1406 | **1403** |
| P | 115 | 115 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2592 | 2592（升档不改总数，只升体还原率） |

SHA256 `E136965D8192D9111CD35EBC692E8C88B68FD3E29619496FC8C6739FDB37574F`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批升档（3 处，filesearch_normalize.go）

### `defaultFileSearchTypeFilters` 0x140812580（≈11KB 静态表完整重建）

由 `return nil` 空骨架升为完整 6 过滤器表。还原方法：

- **duffcopy 静态模板**（`0x1411e5d70`）字节级解析得 6 个 ID：`image`/`document`/`video`/`audio`/`archive`/`executable`，Label 全空、Enabled 全 true。
- **6 次 growslice**（num=64/66/33/39/48/50，元素类型 `string` 经 `0x140ae9e00` 确证 kind=24）逐区域提取 `lea rdx,[rip+…]` 字符串地址 + `mov [rax+off],imm` 长度，批量解码全部 300 条扩展名规则（见 `tools/parse_deffilters.py`，保留为逆向证据工具）。
- archive 过滤器含 5 条 glob/分卷规则：`*.7z.*`/`*.zip.*`/`*.rar.*`/`*.tar.*`/`?*.[rz][0-9][0-9]`，与 `classifyNormalizedFileSearchTypeRule` 的分卷特判（case 3）精确对应。

### `normalizeVolumeRoots` 0x14088c6e0

由恒等 stub 升为完整归一化：`len==0→nil`；`make([]string,0,len)+make(map[string]struct{})`；逐条 `normalizeVolumeRoot` 空跳过；去重键 `strings.ToLower(norm)`；append 原 norm。汇编确证去重键为 ToLower 而非原始值。

### `normalizeFileSearchRecentItems` 0x14087a140

由恒等 stub 升为完整归一化：`len==0→nil`；`cap=min(len,128)`；逐条 `TrimSpace(Path)` 空跳过；`RuneCountInString>0x7fff` 跳过；去重键 `ToLower(Replace(path,"/","\\",-1))`；`Name=trimStringLimit(Name,512)`；append 后 `len>127` 终止。汇编确证路径分隔符归一化方向为 `/`→`\`（Windows 风格）。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；`go test ./backend` `ok changeme/backend 0.373s`。
- **G2 count_funcs**：`FUNCS=2707 / S=1154 / S-inline=35 / S-sig=1403 / P=115 / UNMARKED=0`。
- **G3 行为**：6 个默认过滤器 ID/Label/Enabled 经 duffcopy 静态模板字节级确证；300 条扩展名规则逐条 `read_gostring` 解码；archive 分卷 glob 与 `classifyNormalizedFileSearchTypeRule` case 3 特判字符串完全一致。
- **G4 review**：本批仅动 filesearch_normalize.go 单文件，无并行写重叠；升档均为 `[S-sig]→[S]`，无签名变更。
