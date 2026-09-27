# 批次 168 验收（launcherconfig_runtime.go：三个 fileLocator 归一化 helper [S-sig]→[S]）

日期：2026-09-25
子批次：文件定位器 normalize 族（搜索模式 / 布尔作用域 / 过滤器类型）

## 目标

`normalizeFileLocatorSearchMode`（0x14087ae80）、`normalizeFileLocatorBooleanScope`（0x14087b060）、
`normalizeFileLocatorFilterType`（0x14087b100）三个零体 [S-sig] helper 升 [S]。

## G1 门禁（活体实测，go1.25.12）

| 项 | 命令 | 结果 |
|---|---|---|
| build | `go1.25.12 build ./backend` | EXIT=0 |
| vet | `go1.25.12 vet ./backend` | EXIT=0 |
| 全量 test | `go test ./backend` | ok (0.5s) |
| 黄金用例 | `TestNormalizeFileLocatorSearchMode` / `...BooleanScope` / `...FilterType` | PASS |

## G2 覆盖

`bash tools/count_funcs.sh`：
`FUNCS=2659 / S=1100 / S-inline=35 / S-sig=1340 / P=184 / UNMARKED=0`。

**真函数 = 1100 + 35 + 1340 = 2475 / 4754 = 52.1%**（相对批次 167 持平；S 1097→1100、S-sig 1343→1340，
三函数由零体签名级升为真体级，真函数口径不因此增减）。

重建产物 SHA256：`3CB32FEBC08AF8DC71E793618F16A5D2D2C414EBFC30DFF748FEC149617DA373`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## G3 落地明细（3 函数 [S]，`backend/launcherconfig_runtime.go`）

| 函数 | VA | 语义 |
|---|---|---|
| normalizeFileLocatorSearchMode | 0x14087ae80 | TrimSpace→ToLower 后按长度跳表白名单，命中返 canonical，否则 "plain" |
| normalizeFileLocatorBooleanScope | 0x14087b060 | TrimSpace→ToLower；"line"/"lines"→"line"，否则 "file" |
| normalizeFileLocatorFilterType | 0x14087b100 | TrimSpace→ToLower；bool 族→"boolean"，regex 族→"regex"，glob 族→"glob"，否则 "plain" |

## G4 关键实证结论

1. **normalizeFileLocatorSearchMode 跳表**：0x14087aeb3 `lea rdx,[rip+0x9617e6]` → 0x1411DC6A0（10 项
   qword），索引 `len-4`（len 4..13），逐项指向 case 地址；len 8/11 直接指向 default（返 "plain"）。
   比较常量逐字节比对：`0x6c6f6f62`="bool"、`0x64726f77`="word"、`0x65676572`="rege"、
   `0x626f6c67`（无，此为 FilterType）、`0x6765722d6c6f6f62`="bool-reg"、`0x6f772d656c6f6877`="whole-wo"、
   `0x726e61656c6f6f62`="booleanr"、`0x2d6e61656c6f6f62`="boolean-"。
2. **canonical 常量**（rip 相对逐个用 Python 计算）：`boolean`(7B@0x140C7A3A1)、
   `wholeWord`(9B)、`regex`(5B@0x140C75C54)、`booleanRegex`(12B@0x140C4B2D8)、`plain`(5B@0x140C75C4F)。
3. **normalizeFileLocatorBooleanScope**：0x14087b080 len 4 比对 `0x656e696c`="line"；len 5 比对 "line"+"s"；
   命中返 "line"(4B@0x140C4A900)，否则返 "file"(4B@0x140C46C76)。
4. **normalizeFileLocatorFilterType**：len 4 比对 "bool"/"glob"/"mask"；len 5 "regex"；len 6 "regexp"；
   len 7 "boolean"；len 8 `0x64726163646c6977`="wildcard"；canonical："boolean"(7B)、"regex"(5B)、
   "glob"(4B@0x140C4C1F7)、默认 "plain"(5B)。

## 残留 / 未落地（不触及）

- `normalizeFileLocatorFilters`（0x14087b240）、`isZeroFileLocatorConfig`（0x14087bac0）、
  `defaultTagCatalogItems`（0x14087bc20）等 [S-sig] 仍为零体，按批次落地。
- 三个 [P]（loadOrCreateLauncherConfig / loadLauncherConfigOrDefaultIfMissing / readLauncherConfigFile）
  多寄存器签名已初步取证（第二个入参非 string 而是选项切片/结构），待逐寄存器坐实后升 [S-sig]。
