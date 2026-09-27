# 批次 164 验收（launcherconfig_runtime.go：normalizeFileLocatorConfig [P]→[S] + normalizeFileLocatorHistoryEntries 签名修正 [S-sig]→[S]）

日期：2026-09-25
子批次：文件定位配置归一化 + 历史条目归一化（含历史条目签名修正）

## 目标

1. `normalizeFileLocatorConfig(FileLocatorConfig) FileLocatorConfig`：`[P]`→`[S]`。
2. `normalizeFileLocatorHistoryEntries`：签名由 `([]string) []string` 修正为 `([]string, bool) []string`（汇编 dil 入参），`[S-sig]`→`[S]`。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go1.25.12 test -count=1 -p=1 ./backend` | ok (0.419s) |
| 黄金用例 | `TestNormalizeFileLocatorHistoryEntries` / `TestNormalizeFileLocatorConfig` | PASS |

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2659 / S=1091 / S-inline=35 / S-sig=1347 / P=186 / UNMARKED=0`。

**真函数 = 1091 + 35 + 1347 = 2473 / 4754 = 52.0%**（相对批次 163 的 2472 增 +1，P 187→186）。

重建产物 SHA256：`1E68006985A04516311312505EAF1C7FD27A17E7194F406FF4333C75BEF0074D`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细（2 函数 [S]，`backend/launcherconfig_runtime.go`）

| 函数 | VA | 语义 |
|---|---|---|
| normalizeFileLocatorConfig | 0x14087a7a0 | 过滤器补默认、历史条目归一化、ActiveFilterID 匹配、尺寸钳制 |
| normalizeFileLocatorHistoryEntries | 0x14087b7a0 | 去空白/去空/去重/上限 12，normalizePath 时路径归一化去重键 |

## G4 关键实证结论

1. **normalizeFileLocatorConfig（285 行 asm）**：默认过滤器三字符串常量实解——ID `file-locator-default-non-text`
   （0x140C6D50B, len 29）、Remark `默认非文本排除`（0x140C5F746, len 21，UTF-8）、Type `glob`（0x140C347FE, len 4）。
2. **历史条目 flag 实证**：FileName/ContainsText 历史传 dil=0、SearchRoot 历史传 dil=1（路径归一化：
   `strings.ToLower(strings.Replace(s,"\\","/",-1))` 作去重键，输出保留原去空白串）。
3. **ActiveFilterID 匹配**：TrimSpace 后非空则遍历过滤器，`EqualFold(TrimSpace(f.ID), activeFilterID)` 命中后
   取其原 ID；未命中置空。
4. **MaxSearchFileSizeMB 钳制**：`<=0 → 32`、`>4096 → 4096`，否则保留。
5. **IncludeSubfolders**：`isZeroFileLocatorConfig` 为真时回退默认 `true`（defaultFileLocatorConfig 实测为 true）。
6. **normalizeFileLocatorHistoryEntries（179 行 asm）**：`make([]string,0,min(len,12))`、`map[string]struct{}` 去重、
   空串跳过、结果达 12 提前 break；签名 `(entries []string, normalizePath bool)` 由 dil 字节入参 + `movzx/test dl` 实证。

## 残留 / 未落地（不触及）

- `normalizeFileLocatorFilters`、`normalizeFileLocatorSearchMode`、`normalizeFileLocatorBooleanScope`、
  `isZeroFileLocatorConfig`、`defaultFileLocatorFilterValue` 仍为 `[S-sig]`（零值），normalizeFileLocatorConfig
  的调用已正确接线，其行为等价性依赖这些被调方的后续落地。
- 过滤器匹配循环的 slice 顺序 [ID,Remark,Value,Type] 依赖 types_filelocator.go `FileLocatorFilter` 布局为真。
