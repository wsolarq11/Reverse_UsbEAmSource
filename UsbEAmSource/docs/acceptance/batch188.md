# 批次 188 — 组件存储读取链闭环（3 函数：2 [S] + 1 [S-sig]）

## 基线 / 收口

| 指标 | 基线（批次 187 收口） | 收口（批次 188） |
|---|---|---|
| FUNCS | 2724 | **2727** |
| S | 1172 | **1174** |
| S-inline | 36 | 36 |
| S-sig | 1401 | **1402** |
| P | 115 | 115 |
| UNMARKED | 0 | 0 |
| 真函数（S+S-inline+S-sig） | 2609 | **2612** |

SHA256 `10F7943CA8774DBF67B8EDEE18DE704F745AC39977C9F60FAD931EA55190122B`（`artifacts/UsbEAm_Launcher_rebuilt.exe`，`bash build.sh` 重建）。

## 本批内容（desktopwidgets_store.go，存储读取链完整闭环）

### 1. `validateDesktopWidgetDocument`（0x1407c3940，2368B）— [S-sig]

`func(doc DesktopWidgetDocument) error`。签名已还原；已证校验项（Version!=1 / Revision<0 /
len(Widgets)>200 / Widgets 条目 ID 匹配+类型白名单+Config 序列化上限 / Notes 条目+累计上限 /
Reminders 条目+累计上限 / 终序列化上限 + validateLauncherConfigJSONStructure），体待逐项字节级还原。

### 2. `loadUnlocked`（0x1407c2440，3586B）— [S]

`func(s *launcherWidgetStore) (bool, DesktopWidgetDocument, error)`。读文件→结构校验→JSON 解析→
末尾校验→归一化→文档校验→缓存 全链还原；返回 (bool,doc,error)，错误分支返回 defaultDoc 副本；
loaded/cachedExists/cachedLoadErr 三态缓存字段写入点逐一对齐 asm（仅成功与 ErrNotExist 置 loaded=true，
各错误分支仅写 cachedLoadErr）。

### 3. `Read`（0x1407c0420，608B）— [S]

`func(s *launcherWidgetStore) (bool, error)`。s==nil→(false,"首页组件存储不可用"@0x140c69cd5,27B)；
mu.Lock + defer Unlock；exists,_,err=loadUnlocked() 后返回 (exists,err)。

## 门禁

- **G1 build/vet/test**：`go1.25.12 build ./backend` EXIT=0；`go vet ./backend` EXIT=0；`go test ./backend` `ok changeme/backend 0.944s`。
- **G2 count_funcs**：`FUNCS=2727 / S=1174 / S-inline=36 / S-sig=1402 / P=115 / UNMARKED=0`。
- **G3 行为**：错误字符串（"首页组件存储路径不能为空"@0x140c782eb / "首页组件存储不可用"@0x140c69cd5 /
  "解析 JSON 失败: %v"@0x140c611e4）与 os.ErrNotExist/errors.Is/bytes.NewReader/json.NewDecoder/Decode/
  ensureDesktopWidgetJSONEOF 调用链逐条 asm 对齐；launcherWidgetStore 字段偏移（mu@0x10/loaded@0x18/
  cachedExists@0x19/cached@0x20/cachedLoadErr@0x88）与 types_launcher.go 一致。
- **G4 review**：仅改 desktopwidgets_store.go（追加 3 函数 + import bytes）；无跨文件写重叠。

## 遗留（下一批）

- `validateDesktopWidgetDocument` 升档 [S-sig]→[S]（各错误字符串+字段名逐条字节级对齐）。
- `launcherWidgetStore.Ensure`（0x1407c06e0）/ `writeDesktopWidgetDocumentFile`（0x1407c47c0）——写入侧。
- `bytesMatchWildcardFold` 0x1407e3e60（176 行，通配符匹配）。
- 继续落地 truly-missing 顶层函数（678 个）。
