# 批次 354 · 插件包文件规范化/目录基准 URL/包文件替换/窗口释放 +4（FUNCS 3172）

## 目标

落地 4 个函数：`normalizePluginPackageFile`、`resolvePluginCatalogBaseURL`、
`replacePluginPackageFile`、`pluginWindowService.release`。

## 基线 / 收口

| 指标 | 基线（batch 353 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3168 | 3172 |
| MARKED | 3168 | 3172 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1585 | 1589 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1541 | 1541 |
| USABLE | 1542 | 1542 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1541  FUNCS=3172  MARKED=3172  P=41  S-eq=1  S-inline=37  S-sig=1589  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1589 + 41 = 3172 = FUNCS`。
S-sig 1585→1589（+4）、FUNCS 3168→3172（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 normalizePluginPackageFile [S-sig 0x14092bca0, 352B]

TrimSpace 空→""；`\`→`/`；Index ":" <0 → path.Clean；basename 扩展名 EqualFold ".zip"。

### 3.2 resolvePluginCatalogBaseURL [S-sig 0x14092c220, 352B]

TrimSpace 空→"."；url.Parse；scheme 空/非 http/https→"."；path 不以 "/" 结尾→补 "/" → URL.String。

### 3.3 replacePluginPackageFile [S-sig 0x14092d360, 352B]

Lstat 失败 ErrNotExist→rename；已存在→remove src + error。

### 3.4 pluginWindowService.release [S-sig 0x1409301a0, 352B]

lock → windows(+0x10)[key] 匹配 gen(+0x88)+identity(+0x90) → mapdelete → unlock。

## G4 独立复核

- `backend/plugin_discovery.go`：+normalizePluginPackageFile [S-sig]
  +resolvePluginCatalogBaseURL [S-sig] +replacePluginPackageFile [S-sig]。
- `backend/bootstrapservice_state_deps.go`：+pluginWindowService.release [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3172/4754 = 66.72%。下一批：plugin/screenshot 域
resolvePluginWindowScreen / screenshotPreviewWindowService.readThumbnailPNG。
