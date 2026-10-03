# 批次 299 · nodePathCache.Set + 热键管理器/日志状态释放 +4（FUNCS 2982）

## 目标

落地 4 个短函数：`nodePathCache.Set`（filesearch 缓存写入）、
`windowsOLEDBlackoutHotkeyManager.Close`/`windowsLauncherGlobalHotkeyManager.Close`
（once 幂等关闭入口）、`volumeIndexJournalPendingState.Release`（日志 pending 释放）。

## 基线 / 收口

| 指标 | 基线（batch 298 后） | 收口（本批后） |
|---|---|---|
| FUNCS | 2978 | 2982 |
| MARKED | 2978 | 2982 |
| S | 1449 | 1450 |
| S-inline | 37 | 37 |
| S-eq | 1 | 1 |
| S-sig | 1450 | 1453 |
| P | 41 | 41 |
| UNMARKED | 0 | 0 |
| FAITHFUL（S+S-inline） | 1486 | 1487 |
| USABLE | 1487 | 1488 |

`go1.25.12 build/vet/test ./backend` 全 EXIT=0（test `ok changeme/backend`）。

## G1 编译

`go1.25.12 build -tags production -trimpath -buildmode=exe ./backend` EXIT=0；
`go1.25.12 vet ./backend` EXIT=0；`go1.25.12 test -tags production ./backend` `ok changeme/backend`。

## G2 契约

`bash tools/count_funcs.sh` 活体实测：

```
FAITHFUL=1487  FUNCS=2982  MARKED=2982  P=41  S-eq=1  S-inline=37  S-sig=1453  S=1450  USABLE=1488
```

分项自洽：`S + S-inline + S-eq + S-sig + P = 1450 + 37 + 1 + 1453 + 41 = 2982 = FUNCS`。
S 1449→1450（+1）、S-sig 1450→1453（+3）、FUNCS 2978→2982（+4）、FAITHFUL 1486→1487（+1）、
UNMARKED=0/P=41 保持。

## G3 行为（asm 逐地址实证）

### 3.1 nodePathCache.Set [S 0x1407e17e0, 259B]

`(key int32, value string)`。key >= 0x40000000 走 deltaCache（nil 则 makemap_small 建表 +
mapassign_fast32）；< 走 sliceCache 索引（movsxd 符号扩展，负键/越界忽略返回）；两路带写屏障。

### 3.2 windowsOLEDBlackoutHotkeyManager.Close [S-sig 0x14090f540, 128B]

`() error`。closeOnce(+0x24).Do(func1)，闭包写局部 error 后返回；func1（0x14090f5c0, 253B）
关闭 commands/done 链。体待热键关闭链专项。

### 3.3 windowsLauncherGlobalHotkeyManager.Close [S-sig 0x14089da00, 128B]

`() error`。closeOnce(+0x4c).Do(func1)，闭包写局部 error 后返回；func1 关闭键盘钩子/commands
链。体待 hotkey 域专项。

### 3.4 volumeIndexJournalPendingState.Release [S-sig 0x140806820, 160B]

`()`。released(+0x128) 未置位则置位 + callback(+0x120) 非空调用；随后清零 view 尾部
(+0x110..0x130) 与 overlay(+0x130)。体待精确字段布局专项。

## G4 独立复核

- `backend/filesearch_windows.go`：+nodePathCache.Set [S]。
- `backend/oledblackout_hotkey_windows.go`：+windowsOLEDBlackoutHotkeyManager.Close [S-sig]。
- `backend/hotkeymanager.go`：+windowsLauncherGlobalHotkeyManager.Close [S-sig]。
- `backend/filesearch_index_windows.go`：+volumeIndexJournalPendingState.Release [S-sig]。

`go1.25.12 build/vet/test ./backend` 全 EXIT=0。

## 移交（本轮收尾）

4 个函数落地（1 [S] + 3 [S-sig]）。FUNCS 2982/4754 = 62.73%。下一批：filesearch
SearchWithPaths 薄包装 + searchWithPathsContextMetrics 签名 / remoteicons
validateRemoteIconCachePath 路径校验。
