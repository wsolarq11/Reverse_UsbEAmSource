# 批次 379 · 插件目录查找 +2 / OLED 连续性克隆 +1（FUNCS 3260）

## 目标

1. 新增 `findRemoteCatalogEntry` `[S]`（0x14092b900，544B）。
2. 新增 `findPluginInUpdateState` `[S]`（0x14092b6e0，544B）。
3. 新增 `oledBlackoutBrowserMediaContinuity.clone` `[S]`（0x1408fcce0，544B）。

## 基线 / 收口

| 指标 | 基线（batch 378 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 3257 | 3260 |
| MARKED | 3257 | 3260 |
| S | 1521 | 1524 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1657 | 1657 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1558 | 1561 |
| USABLE | 1559 | 1562 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

```
FUNCS=3260  MARKED=3260  UNMARKED=0  S=1524  S-inline=37  S-eq=1  S-sig=1657  P=41  FAITHFUL=1561  USABLE=1562  TRUE=3218
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1524 + 37 + 1 + 1657 + 41 = 3260 = FUNCS`。
FUNCS 3257→3260（+3）、S 1521→1524（+3）、FAITHFUL 1558→1561（+3）、TRUE 3215→3218（+3）、UNMARKED=0/P=41 保持。
`list_missing.js` 未落地数 370→367（−3），三函数退出缺失清单。

## G3 行为（asm 逐地址实证）

### 3.1 findRemoteCatalogEntry [S 0x14092b900]

- 签名：`(name string, entries []PluginRemoteCatalogEntry) (PluginRemoteCatalogEntry, bool)`。
  name 经 rax/rbx（ptr/len），entries 经栈（ptr/len 读 [rsp+0x1b8]/[rsp+0x1c0]）。
- 遍历（步长 0x138 = sizeof(PluginRemoteCatalogEntry)）：
  `EqualFold(TrimSpace(e.ID), TrimSpace(name))`（0x14092b9e4/0x14092ba06/0x14092ba20）。
- 命中：拷贝整个 entry（首 word 手动 + duffcopy 剩余）→ `eax=1`；未命中：清零 + `eax=0`。

### 3.2 findPluginInUpdateState [S 0x14092b6e0]

- 与 3.1 同构，遍历 `[]PluginManifest`（步长 0x178 = sizeof(PluginManifest)），
  `EqualFold(TrimSpace(e.ID), TrimSpace(name))`，返回 `(PluginManifest, bool)`。

### 3.3 oledBlackoutBrowserMediaContinuity.clone [S 0x1408fcce0]

结构（types_oled.go）：`lock sync.Mutex`@+0x0、`entries map[...]string`@+0x8。

- `newobject` 新建 result；`c==nil` → 返回空 result（0x1408fcdee）。
- `lock.Lock`（cmpxchg → lockSlow）→ `len(entries)>0` 则 `make(map, n)`（0x1408fcd79 makemap）
  → mapIterStart/mapIterNext 遍历 + mapassign 复制 → `lock.Unlock` → 返回 result。

## G4 独立复核

- `backend/pluginupdate.go`：+`findRemoteCatalogEntry`/`findPluginInUpdateState`（[S]，依赖 strings + 已落地类型）。
- `backend/oledblackout.go`：+`clone`（[S]，紧邻 `clear`，依赖已落地 map key 类型）。
- 无新增 import（strings 两文件均已有）；`go build/vet/test` 全仓通过；`go fmt` 无差异。

## 移交（本轮收尾）

3 函数落地（均 [S]）。FUNCS 3260/4754 = 68.57%，FAITHFUL 1561/4754 = 32.83%，TRUE 3218/4754 = 67.69%。
下一批：plugin update catalog 域剩余（`mergeLocalAndRemotePluginManifest`/`remoteCatalogEntryToManifest`/
`validateAndNormalizeRemoteCatalogEntries`/`buildPluginUpdateCatalogStateForWorkspace`）；filesearch 域 480B 候选；
`TestHotCorner`/`PickAppTarget` delegate 体；`P=41 → [S]` 转换。
