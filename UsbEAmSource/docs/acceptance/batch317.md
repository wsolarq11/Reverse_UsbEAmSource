# 批次 317 · journal 变更/多路径搜索上下文/mmap 静态判定 +3（FUNCS 3037）

## 目标

落地 3 个短函数：`VolumeIndex.ApplyJournalChanges`（journal 变更薄包装）、
`VolumeIndex.SearchWithPathsContext`（多路径搜索上下文）、
`shouldUseVolumeIndexMmapStaticProvider`（mmap 静态 provider 判定）。

## 基线 / 收口

| 指标 | 基线（batch 316 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3034 | 3037 |
| MARKED | 3034 | 3037 |
| S | 1471 | 1472 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1484 | 1486 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1508 | 1509 |
| USABLE | 1509 | 1510 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1509  FUNCS=3037  MARKED=3037  P=41  S-eq=1  S-inline=37  S-sig=1486  S=1472  USABLE=1510
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1472 + 37 + 1 + 1486 + 41 = 3037 = FUNCS`。
S 1471→1472（+1）、S-sig 1484→1486（+2）、FUNCS 3034→3037（+3）、FAITHFUL 1508→1509（+1）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 VolumeIndex.ApplyJournalChanges [S-sig 0x140804b80, 160B]

默认参数转发 ApplyJournalChangesDetailedResult。

### 3.2 VolumeIndex.SearchWithPathsContext [S-sig 0x140808160, 160B]

参数重排后 searchWithPathsContextMetrics（8 参数 + 9 字返回）。

### 3.3 shouldUseVolumeIndexMmapStaticProvider [S 0x1407e77a0, 256B]

Getenv(39 字符名) → TrimSpace → ToLower → "0"/"no"/"off"/"false"/"disable"/"disabled" → false；
空 → 全局默认 !=0；其他 → true。

## G4 独立复核

- `backend/filesearch_index_windows.go`：+ApplyJournalChanges +SearchWithPathsContext [S-sig]
  +shouldUseVolumeIndexMmapStaticProvider [S]（+strings import +volumeIndexMmapStaticDefault 全局）。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（1 [S] + 2 [S-sig]）。FUNCS 3037/4754 = 63.88%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
