# 批次 194 — 插件发现域全落地（8 函数 [S] + discoverPluginsUnlocked 升 [S]）

## 基线 / 收口

| 指标 | 基线（批次 193 收口） | 收口（批次 194） |
|---|---|---|
| FUNCS | 2753 | **2761** |
| S | 1201 | **1210** |
| S-inline | 36 | 36 |
| S-sig | 1400 | 1400 |
| P | 116 | **115** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2637 | 2646（55.7%） |

SHA256 `EA917E6F8DCDE7536A97B89ED47F76343A9E88C4764E1F5A5E8F9008AE04C82B`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,279,744 B，`bash build.sh` 重建）。

## 本批内容

### 1. `plugin_discovery.go`（8 funcs）— 完整落地

新建文件，落地插件发现域的语言本地化解析链 + 图标归一化链（蓝图归属 `bootstrapservice.go`
+ `pluginupdate.go`，全部 [S] 逐条 asm 对齐）：

- `normalizePluginIconFile`（0x14092bb20，384B）：TrimSpace 空则 "" → `\`→`/` → 含 `/` 则 "" →
  `path.Clean(icon)!=icon` 则 "" → 扩展名 ToLower 仅 `.svg`/`.png` 通过。
- `resolvePluginLocalIconURL`（0x14092c040，480B）：normalize → SourceDir/Icon 空则 "" →
  IconURL 空则 "" → `os.Stat(Join(srcDir,icon))` 出错或目录则 "" →
  `fmt.Sprintf("/plugin-host/%s/%s", url.PathEscape(iconURL), url.PathEscape(icon))`。
- `mergeManifestLocalizations`（0x1407a6380，384B）：遍历 src → TrimSpace(key) 空则 "en-US" →
  Messages nil 则 make → dst[key]=loc。
- `loadPluginLocalizations`（0x1407a5ec0，1728B）：merge I18N → `resolvePluginPathSecurely(SourceDir,
  "language", true)` 失败返回 → ReadDir 失败返回 → 逐项：目录跳过、非 `.ini`（EqualFold 扩展名）跳过 →
  resolve 路径 + `readPluginFileBounded(0x80000)` → `parseLanguageContent` → lang 空跳过 →
  以 `nestedLanguageMessageString` 覆盖 Name/Description（`plugin.name`/`plugin.description`）→
  `mergeLanguageMessages` → 写回 result[lang]。
- `parseLanguageContent`（0x1407a3bc0，576B）：`bufio.NewScanner(bytes.NewReader(data))` 薄封装 →
  `parseLanguageScanner` 透传。
- `parseLanguageScanner`（0x1407a3e00，2334B）：lang 默认 `TrimSuffix(Base(path),Ext)` → Scan 循环：
  去 BOM/TrimSpace → 空行与 `;`/`#` 注释跳过 → `[x]` 段落行更新 section → `Cut(line,"=")` 无 `=` 跳过 →
  key/value TrimSpace → `decodeLanguageValue(value)` → `EqualFold(section,"meta")` 则按 ToLower(key)
  分派 code/name/native_name → section 非空则 `setNestedLanguageMessage(messages, section+"."+key, value)`。
- `decodeLanguageValue`（0x1407a4760，480B）：`Index(value,"\\")<0` 原样返回 → Builder 逐 rune：
  `\n`/`\r`/`\t`/`\\` 解码，未知转义补 `\` 保留原字符，尾随孤立 `\` 保留。
- `setNestedLanguageMessage`（0x1407a4940，576B）：`Split(TrimSpace(key),".")` → 逐层：part 空则返回 →
  末层 `cur[part]=value`；中间层非 map 则建新 map 下钻。

### 2. `bootstrapservice_callees.go`（discoverPluginsUnlocked [P]→[S]）

- `discoverPluginsUnlocked`（0x1407a54a0，2432B）**[P]→[S]**：`pluginDirectories()` → 逐目录 ReadDir →
  跳过非目录/symlink → 跳过 `.` 前缀 → resolve 目录 → resolve `plugin.json` →
  `readPluginManifestBounded` → Entry 空则 `"index.html"` → resolve Entry → 填 SourceDir/Installed →
  `normalizePluginIconFile` + `readPluginFileBounded(1<<20)` 验证 → `resolvePluginLocalIconURL` →
  `loadPluginLocalizations` → `ToLower(TrimSpace(ID))` 去重 map → `sort.Slice` 按 ID 排序。
  新增依赖 import `os`/`sort`。

### 3. `nestedLanguageMessageString` TrimSpace 修正

既有 [S] 函数（bootstrapservice_callees.go:150）补 `strings.TrimSpace(key)` 与
`part=strings.TrimSpace(part)` + 空 part 返回 ""，对齐 asm 0x1407a6500（TrimSpace(key)+
TrimSpace(每个 part)+空 part→""）。此前缺 TrimSpace 与 asm 不符。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；
  `go test ./backend` `ok changeme/backend (cached)`。
- **G2 count_funcs**：`FUNCS=2761 / S=1210 / S-inline=36 / S-sig=1400 / P=115 / UNMARKED=0`。
- **G3 行为**：全部常量经 resolve_lea_targets.py 确定性解码（locale 目录 `language`@0x140c3c134、
  扩展名 `.ini`@0x140c347c2、段落匹配 `meta`@0x140c347c6、key 字面量 code/name/native_name、
  合并兜底 `en-US`@0x140c35be6、格式 `/plugin-host/%s/%s`@0x140c59fb4、BOM `\xef\xbb\xbf`@0x140c33b5d）；
  `EqualFold(section,"meta")` 同时门控 meta 键值对写入 messages map（asm jmp 0x1407a450f）。
- **G4 review**：新建 plugin_discovery.go + bootstrapservice_callees.go 追加 os/sort import 与
  discoverPluginsUnlocked 体；复用已有 readPluginFileBounded/readPluginManifestBounded/
  resolvePluginPathSecurely/mergeLanguageMessages/PluginManifest/PluginLocalization；
  无跨文件写重叠。字段偏移全部来自 asm，未臆造签名。

## 遗留（下一批）

- §10 差集仍 **57**（plugin_discovery.go 非蓝图文件名，pluginupdate.go 仅部分落地，文件名级差集不变）。
- 继续 `pluginupdate.go`、`pluginwindow.go` 插件域未落地文件（normalizePluginIconFile/
  resolvePluginLocalIconURL 已从 pluginupdate.go 蓝图迁出落地）。
- 其余遗留：`bytesMatchWildcardFold` 0x1407e3e60、`TestReminderNotification` 0x1407a7320、
  `validateLauncherConfigIconData`、`desktopCalendarStringList` 0x1407aa820。
