# 批次 179 — 文件搜索类型过滤 + 忽略规则域全量落地（23 函数新增 + 1 升档）

## 基线 / 收口

| 指标 | 基线（批次 178 收口） | 收口（批次 179） |
|---|---|---|
| FUNCS | 2680 | **2703** |
| S | 1120 | **1144** |
| S-inline | 35 | 35 |
| S-sig | 1410 | **1409** |
| P | 115 | 115 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2565（53.9%） | **2588（54.4%）** |
| 未落地文件（差集） | 66 | **63** |

SHA256 `F9007863FE244E607F76C24A83F03A4B1816CE41F8F3939DE7E3D0597749DC74`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批落地（23 函数新增 + 1 升档，3 文件新建）

### 忽略域 · filesearch_ignore_windows.go（8 [S]，subagent 7a883441）

| 函数 | VA | 关键证据 |
|---|---|---|
| `normalizeFileSearchIgnoredDirectoryRule` | 0x1407e0760 | Trim 引号→空判→绝对路径走 path 归一化→TrimRight `/\\`→含分隔符裸名丢弃→`.` 丢弃 |
| `normalizeFileSearchIgnoredDirectoryPath` | 0x1407e0880 | Clean→`.`丢弃→normalizeVolumeRoot EqualFold 卷根特判→TrimRight `/\\` |
| `normalizeFileSearchIgnoredDirectoryMatchPath` | 0x1407e09a0 | path 归一化→`/` 换 `\\`→ToLower |
| `isFileSearchIgnoredDirectoryPathRule` | 0x1407e0a20 | Trim 引号 + TrimSpace 后 IsAbs |
| `fileSearchPathWithinIgnoredDirectory` | 0x1407e0a80 | 空判→相等判→dir 补尾 `\\`→前缀判 |
| `newFileSearchIgnoreMatcher` | 0x1407e0be0 | 归一化→空返空 nameRules→绝对入 pathRules、裸名 ToLower 入 nameRules |
| `(fileSearchIgnoreMatcher) HasRules` | 0x1407e0e40 | nameRules 或 pathRules 任一非空 |
| `(fileSearchIgnoreMatcher) IsIgnoredDirectory` | 0x1407e0e80 | **2 参数 (name,path)**：name 命中 nameRules→true；path 归一化后前缀命中 pathRules→true |

### 类型过滤域 · filesearch_type_filter.go（13 [S]，subagent 7a883441）

| 函数 | VA | 关键证据 |
|---|---|---|
| `normalizeFileSearchTypeFilterID` | 0x140815840 | ToLower(TrimSpace)→逐 rune 限 a-z/0-9/-/_ |
| `normalizeFileSearchTypeRules` | 0x140815900 | nil→(nil,true)；map 去重；上限 128 条 |
| `normalizeFileSearchTypeRuleToken` | 0x140815be0 | rune>64/含分隔符/控制字符→false；`.*` 前缀与 glob 分支 |
| `isValidNormalizedExtensionToken` | 0x140815ec0 | a-z/0-9/.-_+；首尾非 `.`；禁 `..` |
| `isValidFileSearchTypeGlob` | 0x140815fe0 | 逐字符校验 glob 合法字符与 `-` 区间 |
| `isValidASCIICharClass` | 0x140816180 | `[...]` 字符类结构校验（`-` 区间 alnum 且升序） |
| `classifyNormalizedFileSearchTypeRule` | 0x1408162c0 | `?*.[rz][0-9][0-9]`→3；含 `*?[`→2；含 `.`→1；否则 0 |
| `compileFileSearchRuleMatcher` | 0x1408163c0 | simple FNV-1a hash 表 / compound 末字节表 / glob 正则并集 `(?i)^(?:(?:re)\|...)$` / 分卷快路径 |
| `fileSearchRuleMatcherSignature` | 0x140816fe0 | 排序规则 + 统计字段 json → sha256 前 8 字节 hex |
| `(*fileSearchRuleMatcher) Allow` | 0x140817280 | 空匹配器→true；glob/simple/compound/分卷逐层 |
| `(*fileSearchRuleMatcher) matchSimpleExtension` | 0x140817420 | 扩展名 stem 截取 + FNV-1a hash 查表 |
| `(*fileSearchRuleMatcher) matchCompoundExtension` | 0x1408175e0 | 末字节查表 + 后缀比对 |
| `fileSearchGlobToAnchoredRegexp` | 0x1408177e0 | glob→锚定正则（`*`/`?`/`[...]` 转义） |

### 音频域 · desktopwidgets_audio.go（2 [S]，Lead）

| 函数 | VA | 关键证据 |
|---|---|---|
| `normalizeDesktopReminderAudio` | 0x1407a6660 | name 空→"system"@0x140c375e4；vol==0→1；system→path="" |
| `validateDesktopReminderAudio` | 0x1407a68c0 | system 验 vol∈[1,10]；custom 验 path 非空/≤4096/IsAbs/扩展名 EqualFold ".wav"@0x140c347ca |

### 升档（1 处）

| 函数 | VA | 订正 |
|---|---|---|
| `normalizeVolumeRoot` | 0x14088c9a0 | [S-sig] 空骨架（返回 ""）→ [S] 完整体：TrimSpace 空→""；VolumeName 非 "X:"（len≠2 或 [1]≠':'）→""；否则 ToUpper(vol)+"\\"@0x1411cac40。此修复使忽略域 normalizeFileSearchIgnoredDirectoryPath 的卷根特判分支恢复生效 |

## 关键订正与知悉

- **IsIgnoredDirectory 签名**：Lead 初判为单参数 `(path string)`，subagent 经 asm 逐寄存器实证订正为 **双参数 `(name, path string) bool`**（name=SI/R8 命中 nameRules，path=R9/R10 命中 pathRules）。Lead 初版已由 subagent 覆盖为正确版。
- **normalizeFileSearchIgnoredDirectoryPath 卷根特判**：Lead 初版漏了 `normalizeVolumeRoot` 后的 EqualFold 卷根特判 + TrimRight `/\\` 尾段，subagent 版补齐。
- **normalizeVolumeRoot 原本是 [S-sig] 空骨架**，导致卷根特判恒不生效，本批一并升档 [S]。
- **gofmt 顺带格式化** 5 个无关文件（audio_windows.go / bootstrapservice_test.go / launcherconfig_test.go / screenshot_uia_windows.go / screenshot_windows.go），纯格式规范化，语义不变。
- **normalizeFileSearchIgnoredDirectoryRules / normalizeFileSearchTypeFilters** 仍在 filesearch_normalize.go 为 identity stub，normalizeVolumeRoots 未动；默认规则表（defaultFileSearchTypeFilters ≈11KB 静态表）待后续批次还原。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；`go test ./backend` `ok changeme/backend 0.358s`。
- **G2 count_funcs**：`FUNCS=2703 / S=1144 / S-inline=35 / S-sig=1409 / P=115 / UNMARKED=0`。
- **G3 行为**：normalizeDesktopReminderAudio 的 "system" 常量、validateDesktopReminderAudio 的 ".wav" EqualFold、classifyNormalizedFileSearchTypeRule 的 "?*.[rz][0-9][0-9]" 分卷常量均 .rodata 字节级确证；FNV-1a 常量（0xcbf29ce484222325/0x100000001b3）为标准 64 位 FNV-1a。
- **G4 review**：本批为 subagent 与 Lead 并行落地同域，发生写范围重叠；Lead 初版（IsIgnoredDirectory 单参数、normalize...Path 缺卷根特判）被 subagent 正确版覆盖，最终以 asm 实证为准。并行落地时写范围重叠需在后续避免（明确分工文件清单）。
