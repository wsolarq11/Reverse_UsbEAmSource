# 批次 372 · 键盘钩子/插件版本解析/映射读提供器替换关闭/拼音 sections 校验 +4（FUNCS 3244）

## 目标

落地 4 个函数：`windowsLauncherGlobalHotkeyManager.ensureKeyboardHook`、
`parsePluginVersionParts`、`closeMappedReadProviderForReplaceLocked`、
`validateVolumePinyinSections`。

## 基线 / 收口

| 指标 | 基线（batch 371 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3240 | 3244 |
| MARKED | 3240 | 3244 |
| S | 1504 | 1504 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1657 | 1661 |
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
FAITHFUL=1541  FUNCS=3244  MARKED=3244  P=41  S-eq=1  S-inline=37  S-sig=1661  S=1504  USABLE=1542
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1504 + 37 + 1 + 1661 + 41 = 3244 = FUNCS`。
S-sig 1657→1661（+4）、FUNCS 3240→3244（+4）、FAITHFUL/USABLE 保持、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 ensureKeyboardHook [S-sig 0x1408a0c00, 448B]

field(+0x20) 非空→nil；compileCallback → SetWindowsHookEx(0xd) → 失败→清理+error；
成功→field(+0x20)=hook。

### 3.2 parsePluginVersionParts [S-sig 0x14092c800, 480B]

TrimSpace 空→nil；去 'v'/'V' 前缀 → genSplit(".") → Atoi 每段→失败→nil。

### 3.3 closeMappedReadProviderForReplaceLocked [S-sig 0x1407f5be0, 480B]

TrimSpace(path) 空→return；IsSourcePath→匹配→HeapSnapshot → clearNameTrigramIndexLocked →
closeVolumeIndexReadProvider。

### 3.4 validateVolumePinyinSections [S-sig 0x14080c220, 480B]

遍历 sections 校验边界 + 段字符 ASCII → 非法→error。

## G4 独立复核

- `backend/hotkeymanager.go`：+ensureKeyboardHook [S-sig]。
- `backend/plugin_discovery.go`：+parsePluginVersionParts [S-sig]。
- `backend/filesearch_index_windows.go`：+closeMappedReadProviderForReplaceLocked [S-sig]
  +validateVolumePinyinSections [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（4 [S-sig]）。FUNCS 3244/4754 = 68.24%。下一批：filesearch/oled 域
VolumeIndex.activeEntryCountLocked / VolumeIndex.overlayStatsLocked。
