# 批次 210 — 节点访问与墓碑链（nodeAt/nodeAtIndex/nodeName/IsTombstonedDescendant）

## 基线 / 收口

| 指标 | 基线（批次 209 收口） | 收口（批次 210） |
|---|---|---|
| FUNCS | 2782 | **2786** |
| S | 1248 | **1252** |
| S-inline | 36 | 36 |
| S-sig | 1385 | 1385 |
| P | 113 | 113 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2669 | **2673**（56.23%） |
| §10 差集 | 54 | 54 |

SHA256 `5AB200AFBBCE88F146CEEE787D252E401E265C18F138AB9B1E4470B0345C0CB2`
（`artifacts/UsbEAm_Launcher_rebuilt.exe`，21,321,728 B，`bash build.sh` 重建）。

## 本批内容

`backend/filesearch_index_windows.go` 扩展 4 函数 [S] + `backend/types_filesearch.go`
新增 `volumeIndexTombstoneLookup` 结构。FUNCS +4 / S +4。

### 新建 [S]（体完整翻译）

- `volumeIndexReadView.nodeAt`（0x1407f2060，480B）：nodeBytes 非空时从序列化字节
  （每节点 24B 小端）解析 IndexNode；否则直接取 `v.nodes[i]`。
- `volumeIndexReadView.nodeAtIndex`（0x1407f2240，544B）：nodeIndex>=0x40000000 走
  deltaNodes 覆盖；越界/负值返回零值；节点数 = nodeBytes.len/24 或 nodes.len。
- `volumeIndexReadView.nodeName`（0x1407f1e40，544B）：NameOffset==0xFFFFFFFF 表示
  delta 节点（先查 deltaByFRN 再线性扫 deltaNodes，Flags&2 跳过）；否则从
  namePool[NameOffset : NameOffset+NameLen] 切片。
- `volumeIndexTombstoneLookup.IsTombstonedDescendant`（0x1407f1a60，992B）：沿父链
  判定墓碑后代，迭代 + memo 记忆化；delta 节点递归到 ParentIdx。

### 结构

- `volumeIndexTombstoneLookup`（0xe8 字节）：嵌入 `volumeIndexReadView`（0x00..0xe0）
  + `memo map[int32]bool` @ +0xe0。

## 关键知悉

- `nodeAt` 的 1638 行源码为虚高（含大量 bounds check 展开），实际 asm 仅 480B：两条路径
  各一次 24B 读取。nodeBytes 路径字段小端：FRN@+0、NameOffset@+8、ParentIdx@+0xc、
  ModTime@+0x10、NameLen@+0x14、Flags@+0x16，与 IndexNode 结构逐一吻合。
- `nodeAtIndex` 的节点数用 magic 除法（0xaaaaaaaaaaaaaaab，除以 24）对 nodeBytes.len
  取整；nodeBytes 为空时回退 `len(nodes)`。delta 覆盖仅在 `deltaIdx < len(deltaNodes)`
  时生效。
- `nodeName` 的 delta 判定：`NameOffset == 0xFFFFFFFF` 时走 delta 路径；deltaByFRN 值
  需 `>= 0x40000000`（减 0x40000000 得 deltaNodes 索引），且 `node.Flags&2 != 0` 的
  记录跳过；fallback 从 deltaNodes 末尾向前线性扫 FRN 匹配。
- `IsTombstonedDescendant` 的入口护栏：`t==nil || len(t.tombstones)==0` 或负索引 → false；
  delta 节点递归到 `deltaNodes[deltaIdx].node.ParentIdx`。普通路径迭代沿 ParentIdx 上溯，
  每步先查 memo（命中即 break），再把当前节点压入 path（cap 8 栈优化），查 tombstones
  命中则 result=true；回填阶段把 path 上所有节点写入 memo。

## 门禁

- **G1 build/vet/test**：`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test ./backend`
  `ok changeme/backend`；`bash build.sh` EXIT=0。
- **G2 count_funcs**：`FUNCS=2786 / S=1252 / S-inline=36 / S-sig=1385 / P=113 / UNMARKED=0`。
- **G3 行为**：新增 `TestNodeAtAndNodeAtIndex` / `TestNodeName` /
  `TestIsTombstonedDescendant` 黄金用例覆盖双路径取节点、delta 覆盖、越界护栏、namePool
  切片、deltaByFRN/线性扫描/Flags 跳过、墓碑链命中/根回退/memo 回填/delta 递归，均 PASS。
- **G4 review**：扩展 `backend/filesearch_index_windows.go`（4 函数）+ `types_filesearch.go`
  （1 结构）+ 测试；vet/test/build 复验通过。

## 遗留（下一批）

- §10 差集 54 文件。`matchSearchCandidateNodeWithPinyin`（0x140810440）的全部依赖现已就绪：
  matchNodeNameTermsWithPinyin / classifyVolumePinyinMatch / IsTombstonedDescendant /
  nodeAtIndex / nodeName 均已落地，可本批收口该主函数。候选：
  `searchPinyinContextWithTombstones`（0x140810e60）、`pluginupdate.go`、
  `pluginwindow.go`、`twofactor_provision_misc.go` 各存根、`transport.go`、
  `TestReminderNotification`（0x1407a7320）。
