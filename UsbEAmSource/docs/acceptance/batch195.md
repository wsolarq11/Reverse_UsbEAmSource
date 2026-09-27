# 批次 195 — 文件搜索通配符折叠匹配（bytesMatchWildcardFold [S]）

## 基线 / 收口

| 指标 | 基线（批次 194 收口） | 收口（批次 195） |
|---|---|---|
| FUNCS | 2761 | **2762** |
| S | 1210 | **1211** |
| S-inline | 36 | 36 |
| S-sig | 1400 | 1400 |
| P | 115 | 115 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2646 | 2647（55.7%） |

SHA256 `D9A9879424FB2E100AC6B782A9B6E5168EBD3EB9CBF46AD8B034498FC5927CD0`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,279,744 B，`bash build.sh` 重建）。

## 本批内容

### `bytesMatchWildcardFold`（0x1407e3e60，480B）— [S] 完整落地

`filesearch_index_windows.go` 追加（蓝图 `filesearch_index_windows.go` Lines 936–1112，176 行源码，
现场 dump `bytesMatchWildcardFold.asm.txt` 150 行逐条对齐）。签名经调用点 `matchNodeNameTerms`
（0x1407e3ac0）三处 call 交叉确证为 `(b []byte, patterns [][]byte, caseSensitive bool) bool`。

语义（asm 逐分支）：

- `len(patterns)==0` → true（`test rsi,rsi; je` 0x1407e3e83）。
- `len(patterns)==1` → 退化为 `bytesContainsFold(b, patterns[0], caseSensitive)`
  （0x1407e3e8f 展开单一 24B slice header 后 `call bytesContainsFold` 尾调）。
- 否则顺序遍历 patterns：空片段跳过（`test r10,r10; je` 0x1407e3ecc）；每片段在 b 中找首个
  匹配起点 k（内层 `k <= len(b)-len(p)`，逐字节比较折叠仅作用 b 侧），找到则
  `b = b[k+len(p):]` 继续下一片段（0x1407e3f09–0x1407e3f1c 的 cap/len/ptr 切片三步），
  找不到 → false（0x1407e3f6b `xor eax,eax`）；全部匹配 → true（0x1407e3f22 `mov eax,1`）。

关键细节：折叠 `'A'<=c<='Z' → c|=0x20` 与 bytesContainsFold/bytesStemEqualFold 一致，仅作用
b（haystack）侧；`caseSensitive` 由 R9B 透传（`test r9b,r9b; jne` 跳过折叠）。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；
  `go test ./backend` `ok changeme/backend 0.385s`。
- **G2 count_funcs**：`FUNCS=2762 / S=1211 / S-inline=36 / S-sig=1400 / P=115 / UNMARKED=0`。
- **G3 行为**：新增 `filesearch_index_windows_test.go` 6 用例（空 patterns、单 pattern 折叠/敏感、
  顺序拼接、中间片段缺失、空片段跳过、首匹配消耗语义）全 PASS；全量 `go test ./backend` PASS。
- **G4 review**：单函数追加到 filesearch_index_windows.go + 新建测试文件；复用已落地
  bytesContainsFold；无跨文件写重叠；签名由调用点三处 call 交叉实证，未臆造。

## 遗留（下一批）

- §10 差集仍 57（filesearch_index_windows.go 已存在，差集不变）。
- 同文件 `matchNodeNameTerms`（0x1407e3ac0，928B，本批调用者）与
  `matchNodeNameTermsWithPinyin`（0x14080fe40）留待后续。
- 其余遗留：`pluginupdate.go`、`pluginwindow.go`、`TestReminderNotification` 0x1407a7320、
  `validateLauncherConfigIconData`、`desktopCalendarStringList` 0x1407aa820。
