# 批次 184 — 标题归一化 + 文件搜索字节折叠匹配（3 函数）

## 基线 / 收口

| 指标 | 基线（批次 183 收口） | 收口（批次 184） |
|---|---|---|
| FUNCS | 2714 | **2717** |
| S | 1160 | **1163** |
| S-inline | 36 | 36 |
| S-sig | 1403 | 1403 |
| P | 115 | 115 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2599 | **2602** |

SHA256 `848B5876B06BAC3C185320C4FB5332BCABD70270E8DC193C7EAF297B9FBD6AEC`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批落地（3 个 truly-missing 函数 → [S]）

### `normalizeDesktopWidgetTitle` 0x1407bef80（576B，desktopwidgets_service.go 新建）

`func(title, kind string) string`。`[]rune(TrimSpace(title))`，len>160 截断；len>0 → `string(runes)`；空标题按 kind 逐字节 cmp 回退：note→Note、clock→Clock、timer→Timer、weather→Weather、calendar→Calendar、reminder→Reminder、stopwatch→Stopwatch、worldClock→World Clock、其余→Widget。9 个默认串全部 `read_gostring.py` 字节级解码（首字母大写，含 "World Clock" 11B）。

### `bytesContainsFold` 0x1407e1260（224B，filesearch_index_windows.go 新建）

`func(b, sub []byte, caseSensitive bool) bool`。len(sub)==0→true；len(b)<len(sub)→false；双循环朴素匹配，外 i∈[0,len(b)-len(sub)]，内 j∈[1,len(sub))。折叠分支（`test r9b,r9b; jne` 跳过）：'A'<=c<='Z' 时 `c|=0x20`；sub 由调用方保证已小写（asm 仅折叠 b 侧）。

### `bytesStemEqualFold` 0x1407e1340（224B，同文件）

`func(b, sub []byte, caseSensitive bool) bool`。从 len(b)-1 往前找 '.'；dot<=0→false；len(sub)!=dot→false；逐字节比较 b[0:dot] 与 sub（折叠同 bytesContainsFold），全等→true。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；`go test ./backend` `ok changeme/backend 0.370s`。
- **G2 count_funcs**：`FUNCS=2717 / S=1163 / S-inline=36 / S-sig=1403 / P=115 / UNMARKED=0`。
- **G3 行为**：normalizeDesktopWidgetTitle 的 9 条默认串 + 截断上限 160（cmp rbx,0xa0）逐字节实证；两个 bytes* 函数的折叠标志经调用点 matchNodeNameTerms（0x1407e3ac0）确证为 bool 透传（r9b 保存/恢复），false=折叠 true=敏感。
- **G4 review**：两个新文件各追加独立函数；无签名变更、无跨文件写重叠。

## 遗留（下一批）

- `bytesMatchWildcardFold` 0x1407e3e60（176 行，同文件，通配符匹配）。
- `launcherWidgetStore.Read`/`loadUnlocked`/`readDesktopWidgetStoreBytes`/`normalizeDesktopWidgetDocument` 存储层依赖链（desktopwidgets_store.go）。
- `TestReminderNotification` 0x1407a7320（依赖 store.Read + normalizeDesktopWidgetTitle，后者已落地）。
