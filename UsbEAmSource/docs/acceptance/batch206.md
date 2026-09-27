# 批次 206 — filesearch 排序比较器域闭环（compareText/compareCandidate/Resolve/Release）

## 基线 / 收口

| 指标 | 基线（批次 205 收口） | 收口（批次 206） |
|---|---|---|
| FUNCS | 2769 | **2773** |
| S | 1235 | **1239** |
| S-inline | 36 | 36 |
| S-sig | 1384 | **1385** |
| P | 114 | **113** |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2655 | **2660**（55.95%） |
| §10 差集 | 54 | 54 |

SHA256 `AF2600E9B8EC4126EEAEF39EF711A6766F8F6991E97AF661764631021108E588`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B，`bash build.sh` 重建）。

## 本批内容

`backend/filesearch_windows.go` 排序比较器子域闭环：4 函数升档 [S]，1 个 [P] 转正，
4 个依赖存根新增。FUNCS +4 / S +4 / S-sig +1 / P -1。

### 升档（体完整翻译）

- `fileSearchCandidatePathResolver.Resolve`（0x14081ff00，416B）[S-sig]→[S]，**签名修正**
  void→`string`：nil → ""；pathResolved → Path；Index==nil → 置 pathResolved 返 Path；
  查 pathCaches[Index]，缺则建缓存；计时 `ResolvePathByNodeIndexAndFRN(NodeIndex,FRN,cache)`
  → 填 Path/NodeIndex；pathResolveDuration += time.Since；pathResolved=true；返 Path。
- `fileSearchCandidatePathResolver.Release`（0x1408200a0，224B）[S-sig]→[S]：nil → 返回；
  遍历 pathCaches 逐项 `returnNodePathCache(cache)` + `delete(pathCaches, idx)`。
- `fileSearchSortComparer.compareText`（0x140820180，320B）[S-sig]→[S]：TrimSpace 后空值
  处理（双空=0/a 空=1/b 空=-1），否则 collator 非空走 `CompareString`、空走
  `strings.Compare`；`spec.Direction=="desc"` 取反。
- `fileSearchSortComparer.compareCandidate`（0x1408202c0，864B）**[P]→[S]，签名修正**
  `bool`+2 参 → `int`+3 参（加 `resolver *fileSearchCandidatePathResolver`）：nil 三态
  （0/1/-1）；resolvePath 闭包（resolver 非空走 Resolve，否则 Path，均 TrimSpace）；
  `spec.Key` 分派：`"path"`→path/name/modifiedAt，`"modifiedAt"`→modifiedAt/name/path，
  默认→name/path/modifiedAt；字符串比较走 compareText（含 desc），ModifiedAtUnix int64
  比较独立做 desc 反转。

### 新增 [S-sig] 存根（签名经 asm/调用方实证，体未翻译）

- `newNodePathCache(count int) *nodePathCache`（0x1407e13e0）。
- `VolumeIndex.NodeCount() int`（0x140809280）。
- `VolumeIndex.ResolvePathByNodeIndexAndFRN(nodeIndex int32, frn uint64, cache *nodePathCache) (string, int32)`（0x140808f20）。
- `returnNodePathCache(cache *nodePathCache)`（0x1407e1560）。

## 关键知悉

- `scoredFileSearchCandidate` 字段偏移 asm 实证：Name@0x20、FRN@0x30、ModifiedAtUnix@0x40、
  Path@0x50（ptr）/0x58（len）、pathResolved@0x60；`fileSearchSortComparer` spec.Key@0x00、
  spec.Direction@0x10、collator@0x20；`fileSearchCandidatePathResolver` pathCaches@0x00、
  pathResolveDuration@0x08。
- Key 字面量 asm 内联小端常量："path"=0x68746170、"desc"=0x63736564、"modifiedAt"=
  0x6465696669646f6d + 0x7441；均无需 .rdata 反解。
- 三态比较（-1/0/1）语义：compareCandidate/compareText 均返回 int（原 [P] 存根误标 bool 已订正）。
- Resolve/Release/compareText/compareCandidate 体均逐寄存器追踪；依赖链
  newNodePathCache/NodeCount/ResolvePathByNodeIndexAndFRN/returnNodePathCache 标注 [S-sig]，
  未臆造签名。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`（1.008s）；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2773 / S=1239 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`。
- **G3 行为**：字段偏移经四处 asm 交叉验证；字面量经内联立即数实证。
- **G4 review**：仅改 `backend/filesearch_windows.go`（+import strings、4 函数升档、
  4 存根新增）；无跨文件写重叠；vet/test/build 复验通过。

## 遗留（下一批）

- §10 差集 54 文件。filesearch 排序比较器子域闭环，但路径解析链
  （newNodePathCache/ResolvePathByNodeIndexAndFRN/returnNodePathCache/NodeCount）仍为
  [S-sig] 待专项。候选：`matchNodeNameTermsWithPinyin`（0x14080fe40）、`pluginupdate.go`、
  `pluginwindow.go`、`twofactor_provision_misc.go` 各存根、`transport.go`、
  `TestReminderNotification`（0x1407a7320）。
