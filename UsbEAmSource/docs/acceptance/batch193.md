# 批次 193 — pluginhost 全部落地（10 函数 [S] + 3 依赖，插件 HTTP 宿主闭环）

## 基线 / 收口

| 指标 | 基线（批次 192 收口） | 收口（批次 193） |
|---|---|---|
| FUNCS | 2740 | **2753** |
| S | 1189 | **1201** |
| S-inline | 36 | 36 |
| S-sig | 1400 | 1400 |
| P | 115 | **116** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2625 | 2637（55.5%） |

SHA256 `23823CC4C32AEF4452BD35DEB8360F55CA8D1DCEBB9D44C59F6AC2E8D3D03810`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批内容

### 1. `pluginhost.go`（10 funcs）— 完整落地

`PluginHostService.ServeHTTP`（0x140921b80，1824B）为入口：全局 RWMutex RLock + defer RUnlock；非 GET/HEAD
设 `Allow: GET, HEAD` + 405；`parsePluginRequestPath` 失败 400；`resolvePluginManifestUnlocked` 失败 404；
`resolvePluginAssetPath` 失败 400；HTML 资源走 `servePluginHTML`（失败 500），其余设
`Cache-Control: no-store` / `X-Content-Type-Options: nosniff` 后 `http.ServeFile`。

其余 9 函数逐条 asm 对齐：

- `resolvePluginManifestUnlocked`（0x140922300，608B）：TrimSpace→validatePluginID 早返→
  `bootstrap.pluginDirectories()`→`discoverPluginsUnlocked`→`EqualFold(ID,id)` 匹配即返回，
  未匹配 `fmt.Errorf("未找到插件: %s", id)`。
- `parsePluginRequestPath`（0x1409225c0，528B）：TrimSpace→去一个前导 `/` 再去一个前导 `\`→
  TrimPrefix `plugin-host/`→空则 `errors.New("插件路径不能为空")`→Split `/`→`TrimSpace(parts[0])`→
  validatePluginID；单段返回 `(id,"",nil)`，多段 `assetPath=Join(parts[1:],"/")`。
- `resolvePluginAssetPath`（0x140922800，480B）：TrimSpace→空或 `/` 用 `manifest.Entry` 兜底→
  反斜杠转正斜杠→TrimPrefix `/`→`resolvePluginPathSecurely(SourceDir,s,true)`→Stat 非目录直接返回；
  目录则 `Rel(SourceDir,resolved)` + `Join(rel,"index.html")` 再 `resolvePluginPathSecurely(...,false)`。
- `isHTMLPluginAsset`（0x1409229e0，160B）：TrimSpace→反扫最后一个 `.`（遇 `\`/`/` 停）→ToLower→
  `ext=='.htm'||ext=='.html'`。
- `servePluginHTML`（0x140922aa0，1184B）：`readPluginFileBounded(path,4<<20)`→`injectPluginRuntime`→
  设 4 响应头（Content-Type/Cache-Control/Referrer-Policy/X-Content-Type-Options）→`w.Write`。
- `injectPluginRuntime`（0x140922f40，2288B）：APIVersion 空则 `"1.0"`；深拷贝 Permissions；
  `resolvePluginLanguageMessages(manifest,"en-US")`→pluginRuntimeMeta JSON→`resolvePluginBaseHref`→
  `buildPluginRuntimeSnippet`→小写化找 `<head`/`<body` 的 `>`，命中 bytes.Join 三段（前缀+snippet+后缀），
  否则 bytes.Join 两段（snippet+html）。
- `resolvePluginBaseHref`（0x140923840，448B）：`Clean(TrimSpace(SourceDir))`→`Rel(srcDir,path)`→`Dir(rel)`；
  目录为 `.` 或 `\` 时 `/plugin-host/{id}/`，否则 `/plugin-host/{id}/{dir 反斜杠转正斜杠}/`。
- `buildPluginRuntimeSnippet`（0x140923a00，224B）：`fmt.Sprintf(template, Replacer.Replace(baseHref), pluginJSON)`；
  模板为 40370 字节 rodata（2 个 `%s` + 26 个 `%%`），以 `//go:embed plugin_runtime_snippet.html` 内嵌，
  与原始 rodata 字节一致；Replacer = HTML 5 对转义（`& < > " '`）。
- `urlPathEscape`（0x140923ae0，48B）：`url.PathEscape(strings.TrimSpace(s))`（mode=2=encodePathSegment）。

`ServeHTTPdeferwrap1`（0x1409222a0）为 `defer mu.RUnlock()` 的编译器自动生成包装，不手写。

### 2. `bootstrapservice_callees.go`（3 依赖）— 补齐 pluginhost 依赖链

- `pluginDirectories`（0x1407a2860，288B）**[S]**：`TrimSpace(workspaceSnapshot().PluginDir@+0x40)`→空则 nil，
  否则 `[]string{dir}`（与 GetSnapshot 0x1407742f8 内联逻辑一致）。
- `resolvePluginLanguageMessages`（0x1407a4f80，448B）**[S]**：TrimSpace 空则 `"en-US"`→新建 map→
  先 merge `I18N["en-US"].Messages`，再 lang!=`"en-US"` 时 merge `I18N[lang].Messages`（复用已有 `mergeLanguageMessages`）。
- `discoverPluginsUnlocked`（0x1407a54a0，2432B）**[P]**：插件发现域专项（os.ReadDir+排序+逐清单解析
  readPluginManifestBounded/normalizePluginIconFile/resolvePluginLocalIconURL/loadPluginLocalizations），
  当前返回空切片，体待 bootstrapservice 域续作还原。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；`go test ./backend` `ok changeme/backend 0.911s`。
- **G2 count_funcs**：`FUNCS=2753 / S=1201 / S-inline=36 / S-sig=1400 / P=116 / UNMARKED=0`。
- **G3 行为**：全部错误字符串/阈值/分支逐条 asm 对齐（405 消息 "method not allowed"、`插件路径不能为空`、
  `未找到插件: %s`、`4<<20`、`en-US`/`1.0` 兜底、`.htm`/`.html` 魔数 0x6d74682e）；模板 40370 字节与 rodata 一致。
- **G4 review**：新建 pluginhost.go + backend/plugin_runtime_snippet.html；bootstrapservice_callees.go 追加 3 依赖；
  复用已有 PluginManifest/pluginRuntimeMeta/mergeLanguageMessages 类型；无跨文件写重叠。

## 遗留（下一批）

- §10 差集 58→**57**（pluginhost.go 已落地）。
- 继续 `pluginupdate.go`、`pluginwindow.go` 等插件域未落地文件。
- `discoverPluginsUnlocked`（0x1407a54a0，2432B）升级 [S]（os.ReadDir+排序+清单解析链）。
- 其余遗留：`bytesMatchWildcardFold` 0x1407e3e60、`TestReminderNotification` 0x1407a7320、
  `validateLauncherConfigIconData`、`desktopCalendarStringList` 0x1407aa820。
