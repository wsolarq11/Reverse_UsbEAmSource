# 批次 301 · 插件版本比较 + 启动器窗口隐藏 +4（FUNCS 2989）

## 目标

落地 4 个短函数：`isPluginVersionUpdateAvailable`/`comparePluginVersionText`
（插件版本比较）、`BootstrapService.HideLauncherWindow`/`HideScreenshotPreview`
（窗口隐藏入口）。新建 pluginupdate.go。

## 基线 / 收口

| 指标 | 基线（batch 300 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2985 | 2989 |
| MARKED | 2985 | 2989 |
| S | 1450 | 1451 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1456 | 1459 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1487 | 1488 |
| USABLE | 1488 | 1489 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1488  FUNCS=2989  MARKED=2989  P=41  S-eq=1  S-inline=37  S-sig=1459  S=1451  USABLE=1489
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1451 + 37 + 1 + 1459 + 41 = 2989 = FUNCS`。
S 1450→1451（+1）、S-sig 1456→1459（+3）、FUNCS 2985→2989（+4）、FAITHFUL 1487→1488（+1）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 isPluginVersionUpdateAvailable [S 0x14092c600, 192B]

`(current, latest string) bool`。各 TrimSpace；latest 空 → false；current 空 → true；
否则 comparePluginVersionText(current,latest) > 0。

### 3.2 comparePluginVersionText [S-sig 0x14092c6c0, 263B]

`(current, latest string) int`。签名实证；体待版本分片解析专项。

### 3.3 BootstrapService.HideLauncherWindow [S-sig 0x14079a220, 192B]

`()`。resolveLauncherWindow 取窗口（nil 则返回）；itab 查找 hideLauncherWindowToTray 调用。
体待窗口隐藏链专项。

### 3.4 BootstrapService.HideScreenshotPreview [S-sig 0x14079be80, 192B]

`()`。mu(+0x540) 保护读 screenshotPreview(+0x408)，非空则 screenshotPreviewWindowService.Hide。
体待字段名精确对位专项。

## G4 独立复核

- `backend/pluginupdate.go`（新建）：+isPluginVersionUpdateAvailable [S]、
  comparePluginVersionText [S-sig]。
- `backend/bootstrapservice_window.go`：+HideLauncherWindow/HideScreenshotPreview [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（1 [S] + 3 [S-sig]）。FUNCS 2989/4754 = 62.87%。下一批：filesearch
SearchWithPaths 薄包装 + searchWithPathsContextMetrics 签名 / remoteicons
validateRemoteIconCachePath 路径校验。
