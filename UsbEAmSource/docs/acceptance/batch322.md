# 批次 322 · 输入轮询停止/三元组头解析/拼音头编码 +3（FUNCS 3052）

## 目标

落地 3 个短函数：`oledBlackoutService.stopInputPollLocked`（输入轮询停止）、
`parseVolumeNameTrigramHeader`（名称三元组头解析）、
`encodeVolumePinyinHeader`（拼音头编码）。

## 基线 / 收口

| 指标 | 基线（batch 321 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3049 | 3052 |
| MARKED | 3049 | 3052 |
| S | 1478 | 1478 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1492 | 1495 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1515 | 1515 |
| USABLE | 1516 | 1516 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

Node.js 等价 count（bash/awk 于本会话不可用，以同逻辑 Node 实现复现 awk）：

```
FAITHFUL=1515  FUNCS=3052  MARKED=3052  P=41  S-eq=1  S-inline=37  S-sig=1495  S=1478  USABLE=1516
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1478 + 37 + 1 + 1495 + 41 = 3052 = FUNCS`。
S-sig 1492→1495（+3）、FUNCS 3049→3052（+3）、UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 oledBlackoutService.stopInputPollLocked [S-sig 0x14090c100, 256B]

inputPollTimer(+0x138) 活跃则 stopTimer 置 nil；计数器(+0x140) 递增；状态字段清零。

### 3.2 parseVolumeNameTrigramHeader [S-sig 0x1407f9ba0, 256B]

len<0x40→error；魔数 "UITG"/版本校验 → 读字段(+8..+0x30)。

### 3.3 encodeVolumePinyinHeader [S-sig 0x14080bca0, 256B]

makeslice(0x80) → 写魔数 "UIYP"/字段 → 字典指纹(+0x18) → 写参数。

## G4 独立复核

- `backend/oledblackout.go`：+stopInputPollLocked [S-sig]。
- `backend/filesearch_pinyin_windows.go`：+parseVolumeNameTrigramHeader
  +encodeVolumePinyinHeader [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

3 个函数落地（3 [S-sig]）。FUNCS 3052/4754 = 64.20%。下一批：remoteicons
validateRemoteIconCachePath 路径校验 / filesearch searchWithPathsContextMetrics 签名。
